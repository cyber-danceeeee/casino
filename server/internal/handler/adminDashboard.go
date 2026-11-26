package handler

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"casino/internal/content"
	"casino/internal/dashboardadmin"
	"casino/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
)

type dashboardCacheEntry struct {
	data      any
	expiresAt time.Time
}
type AdminDashboardHandler struct {
	service *dashboardadmin.Service
	cacheMu sync.RWMutex
	cache   map[string]dashboardCacheEntry
}

func NewAdminDashboardHandler() *AdminDashboardHandler {
	return &AdminDashboardHandler{service: dashboardadmin.New(database.GetDB()), cache: map[string]dashboardCacheEntry{}}
}

func (h *AdminDashboardHandler) Filters(c *gin.Context) {
	data, err := h.service.Filters(adminContentScope(c))
	dashboardWrite(c, data, err)
}
func (h *AdminDashboardHandler) Overview(c *gin.Context)       { h.call(c, h.service.Overview) }
func (h *AdminDashboardHandler) FinancialTrend(c *gin.Context) { h.call(c, h.service.FinancialTrend) }
func (h *AdminDashboardHandler) BettingTrend(c *gin.Context)   { h.call(c, h.service.BettingTrend) }
func (h *AdminDashboardHandler) UserTrend(c *gin.Context)      { h.call(c, h.service.UserTrend) }
func (h *AdminDashboardHandler) UserFunnel(c *gin.Context)     { h.call(c, h.service.UserFunnel) }
func (h *AdminDashboardHandler) OrderStatus(c *gin.Context)    { h.call(c, h.service.OrderStatus) }
func (h *AdminDashboardHandler) GameCategories(c *gin.Context) { h.call(c, h.service.GameCategories) }
func (h *AdminDashboardHandler) TopGames(c *gin.Context)       { h.call(c, h.service.TopGames) }
func (h *AdminDashboardHandler) ProviderPerformance(c *gin.Context) {
	h.call(c, h.service.ProviderPerformance)
}
func (h *AdminDashboardHandler) CountryDistribution(c *gin.Context) {
	h.call(c, h.service.CountryDistribution)
}
func (h *AdminDashboardHandler) ChannelPerformance(c *gin.Context) {
	h.call(c, h.service.ChannelPerformance)
}
func (h *AdminDashboardHandler) AgentPerformance(c *gin.Context) {
	h.call(c, h.service.AgentPerformance)
}
func (h *AdminDashboardHandler) PromotionPerformance(c *gin.Context) {
	h.call(c, h.service.PromotionPerformance)
}
func (h *AdminDashboardHandler) RiskOverview(c *gin.Context) { h.call(c, h.service.RiskOverview) }
func (h *AdminDashboardHandler) Realtime(c *gin.Context)     { h.call(c, h.service.Realtime) }
func (h *AdminDashboardHandler) Alerts(c *gin.Context)       { h.call(c, h.service.Alerts) }

func (h *AdminDashboardHandler) call(c *gin.Context, fn func(dashboardadmin.Filter) (map[string]any, error)) {
	scope := adminContentScope(c)
	cacheKey := dashboardCacheKey(c, scope)
	h.cacheMu.RLock()
	cached, exists := h.cache[cacheKey]
	h.cacheMu.RUnlock()
	if exists && time.Now().Before(cached.expiresAt) {
		response.Success(c, cached.data)
		return
	}
	values := map[string]string{}
	for _, key := range []string{"startTime", "endTime", "timezone", "granularity", "country", "currency", "providerCode", "categoryId", "channelId", "agentId", "metric", "limit"} {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values[key] = value
		}
	}
	filter, err := dashboardadmin.ParseFilter(values)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": "invalid dashboard filters", "data": nil})
		return
	}
	filter.Scope = scope
	data, err := fn(filter)
	if err == nil {
		ttl := 45 * time.Second
		if strings.HasSuffix(c.Request.URL.Path, "/realtime") {
			ttl = 8 * time.Second
		}
		h.cacheMu.Lock()
		h.cache[cacheKey] = dashboardCacheEntry{data: data, expiresAt: time.Now().Add(ttl)}
		h.cacheMu.Unlock()
	}
	dashboardWrite(c, data, err)
}

func dashboardCacheKey(c *gin.Context, scope content.Scope) string {
	return c.Request.URL.Path + "?" + c.Request.URL.RawQuery + "|site=" + strconv.FormatUint(scope.SiteID, 10) + "|platform=" + strconv.FormatBool(scope.Platform)
}

func dashboardWrite(c *gin.Context, data any, err error) {
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "dashboard query failed", "data": nil})
		return
	}
	response.Success(c, data)
}
