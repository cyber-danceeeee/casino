// Package routers is a package dedicated to registering routes, and supports both
// manual route registration and automatic route registration.
package routers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/go-dev-frame/sponge/pkg/errcode"
	"github.com/go-dev-frame/sponge/pkg/gin/handlerfunc"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware"
	"github.com/go-dev-frame/sponge/pkg/gin/middleware/metrics"
	"github.com/go-dev-frame/sponge/pkg/gin/prof"
	"github.com/go-dev-frame/sponge/pkg/logger"

	"casino/docs"
	"casino/internal/config"
	"casino/internal/handler"
	"casino/internal/singlesite"
	"casino/internal/storage"
)

var (
	apiV1RouterFns []func(r *gin.RouterGroup) // group router functions
	// if you have other group routes you can define them here
	// example:
	//     apiV2RouterFns []func(r *gin.RouterGroup)
)

// NewRouter create a new router
func NewRouter() *gin.Engine {
	r := gin.New()
	r.ForwardedByClientIP = true
	r.RemoteIPHeaders = []string{"X-Forwarded-For", "X-Real-IP"}
	if err := r.SetTrustedProxies([]string{
		"127.0.0.1",
		"::1",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}); err != nil {
		logger.Warn("set trusted proxies error", logger.Err(err))
	}

	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		if singlesite.Active() && (singlesite.SaaSPath(c.Request.URL.Path) || (c.Request.Method == "POST" && strings.HasSuffix(c.Request.URL.Path, "/admin/system/sites")) || (c.Request.Method == "DELETE" && strings.Contains(c.Request.URL.Path, "/admin/system/sites/"))) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	})
	r.Use(middleware.Cors())

	if config.Get().HTTP.Timeout > 0 {
		// if you need more fine-grained control over your routes, set the timeout in your routes, unsetting the timeout globally here.
		r.Use(middleware.Timeout(time.Second * time.Duration(config.Get().HTTP.Timeout)))
	}

	// request id middleware
	r.Use(middleware.RequestID())

	// logger middleware, to print simple messages, replace middleware.Logging with middleware.SimpleLog
	r.Use(middleware.Logging(
		middleware.WithLog(logger.Get()),
		middleware.WithRequestIDFromContext(),
		middleware.WithIgnoreRoutes(
			"/metrics",
			"/api/v1/admin/auth/login",
			"/api/v1/admin/auth/mfa/setup/confirm",
			"/api/v1/admin/auth/mfa/verify",
		),
	))

	// metrics middleware
	if config.Get().App.EnableMetrics {
		r.Use(metrics.Metrics(r,
			//metrics.WithMetricsPath("/metrics"),                // default is /metrics
			metrics.WithIgnoreStatusCodes(http.StatusNotFound), // ignore 404 status codes
		))
	}

	// limit middleware
	if config.Get().App.EnableLimit {
		r.Use(middleware.RateLimit(
		//middleware.WithWindow(time.Second*5), // default 10s
		//middleware.WithBucket(1000), // default 100
		//middleware.WithCPUThreshold(750), // default 800
		))
	}

	// circuit breaker middleware
	if config.Get().App.EnableCircuitBreaker {
		r.Use(middleware.CircuitBreaker(
			//middleware.WithBreakerOption(
			//circuitbreaker.WithSuccess(75),           // default 60
			//circuitbreaker.WithRequest(100),          // default 100
			//circuitbreaker.WithBucket(20),            // default 10
			//circuitbreaker.WithWindow(time.Second*3), // default 3s
			//),
			//middleware.WithDegradeHandler(handler),              // Add degradation processing
			middleware.WithValidCode( // Add error codes to trigger circuit breaking
				errcode.InternalServerError.Code(),
				errcode.ServiceUnavailable.Code(),
			),
		))
	}

	// trace middleware
	if config.Get().App.EnableTrace {
		r.Use(middleware.Tracing(config.Get().App.Name))
	}

	// profile performance analysis
	if config.Get().App.EnableHTTPProfile {
		prof.Register(r, prof.WithIOWaitTime())
	}

	r.GET("/health", handlerfunc.CheckHealth)
	r.GET("/health/live", handler.OperationsLiveness)
	r.GET("/health/ready", handler.OperationsReadiness)
	r.GET("/ping", handlerfunc.Ping)
	r.GET("/codes", handlerfunc.ListCodes)
	registerStaticAssetRoutes(r)

	if config.Get().App.Env != "prod" {
		r.GET("/config", gin.WrapF(errcode.ShowConfig([]byte(config.Show()))))
		// register swagger routes, generate code via swag init
		docs.SwaggerInfo.BasePath = ""
		// access path /swagger/index.html
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// register routers, middleware support
	r.Use(generatedCRUDGuard("/api/v1"))
	registerRouters(r, "/api/v1", apiV1RouterFns)
	registerWhiteLabelStaticRoutes(r)
	// if you have other group routes you can add them here
	// example:
	//    registerRouters(r, "/api/v2", apiV2RouteFns, middleware.Auth())

	return r
}

func registerStaticAssetRoutes(r *gin.Engine) {
	handler := func(c *gin.Context) {
		name := strings.TrimPrefix(c.Param("filepath"), "/")
		if name == "" || strings.Contains(name, "\x00") {
			c.Status(http.StatusNotFound)
			return
		}
		staticRoot := config.StaticDir()
		localPath := filepath.Clean(filepath.Join(staticRoot, filepath.FromSlash(name)))
		if relative, err := filepath.Rel(staticRoot, localPath); err != nil || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." {
			c.Status(http.StatusNotFound)
			return
		}
		if fileInfo, err := os.Stat(localPath); err == nil && !fileInfo.IsDir() {
			c.File(localPath)
			return
		}
		proxyStaticAsset(c, "/static/"+name)
	}
	r.GET("/static/*filepath", handler)
	r.HEAD("/static/*filepath", handler)
}

func registerWhiteLabelStaticRoutes(r *gin.Engine) {
	r.GET("/wl-test/h5", serveWhiteLabelH5)
	r.HEAD("/wl-test/h5", serveWhiteLabelH5)
	r.GET("/wl-test/h5/*filepath", serveWhiteLabelH5)
	r.HEAD("/wl-test/h5/*filepath", serveWhiteLabelH5)
	r.GET("/wl-test/admin", serveWhiteLabelAdmin)
	r.HEAD("/wl-test/admin", serveWhiteLabelAdmin)
	r.GET("/wl-test/admin/*filepath", serveWhiteLabelAdmin)
	r.HEAD("/wl-test/admin/*filepath", serveWhiteLabelAdmin)
	if config.Get().App.Env != "prod" {
		r.GET("/", serveWhiteLabelH5)
		r.HEAD("/", serveWhiteLabelH5)
		r.GET("/admin", serveWhiteLabelAdmin)
		r.HEAD("/admin", serveWhiteLabelAdmin)
		r.GET("/admin/*filepath", serveWhiteLabelAdmin)
		r.HEAD("/admin/*filepath", serveWhiteLabelAdmin)
	}
	r.NoRoute(func(c *gin.Context) {
		if shouldServeWhiteLabelShell(c) {
			if strings.HasPrefix(c.Request.URL.Path, "/wl-test/admin") || (config.Get().App.Env != "prod" && strings.HasPrefix(c.Request.URL.Path, "/admin")) {
				serveWhiteLabelAdmin(c)
				return
			}
			serveWhiteLabelH5(c)
			return
		}
		c.Status(http.StatusNotFound)
	})
}

func shouldServeWhiteLabelShell(c *gin.Context) bool {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		return false
	}
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/static/") ||
		path == "/health" || strings.HasPrefix(path, "/health/") || path == "/ping" || path == "/codes" || path == "/metrics" ||
		strings.HasPrefix(path, "/swagger/") {
		return false
	}
	if strings.HasPrefix(path, "/wl-test/h5") || strings.HasPrefix(path, "/wl-test/admin") {
		return true
	}
	if config.Get().App.Env == "prod" {
		return false
	}
	accept := c.GetHeader("Accept")
	return accept == "" || strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")
}

func serveWhiteLabelH5(c *gin.Context) {
	c.File(filepath.Join(config.StaticDir(), "wl", "h5", "index.html"))
}

func serveWhiteLabelAdmin(c *gin.Context) {
	c.File(filepath.Join(config.StaticDir(), "wl", "admin", "index.html"))
}

func proxyStaticAsset(c *gin.Context, assetRef string) {
	refs := []string{assetRef}
	if strings.HasPrefix(assetRef, "/static/") {
		// Older uploads were stored both with and without the configured OSS
		// object prefix. Try the canonical path first, then the legacy raw key.
		refs = append(refs, strings.TrimPrefix(assetRef, "/static/"))
	}
	var response *http.Response
	var lastStatus = http.StatusNotFound
	for _, ref := range refs {
		signedURL, err := storage.PublicURL(c.Request.Context(), ref)
		if err != nil || signedURL == "" || strings.HasPrefix(signedURL, "/static/") {
			continue
		}
		request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, signedURL, nil)
		if err != nil {
			continue
		}
		if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
			request.Header.Set("Range", rangeHeader)
		}
		candidate, err := http.DefaultClient.Do(request)
		if err != nil {
			continue
		}
		lastStatus = candidate.StatusCode
		if candidate.StatusCode >= http.StatusOK && candidate.StatusCode < http.StatusMultipleChoices {
			response = candidate
			break
		}
		candidate.Body.Close()
	}
	if response == nil {
		c.Status(lastStatus)
		return
	}
	defer response.Body.Close()
	for _, key := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Last-Modified", "ETag"} {
		if value := response.Header.Get(key); value != "" {
			c.Header(key, value)
		}
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.Status(response.StatusCode)
	if c.Request.Method != http.MethodHead {
		_, _ = io.Copy(c.Writer, response.Body)
	}
}

func registerRouters(r *gin.Engine, groupPath string, routerFns []func(*gin.RouterGroup), handlers ...gin.HandlerFunc) {
	rg := r.Group(groupPath, handlers...)
	for _, fn := range routerFns {
		fn(rg)
	}
}
