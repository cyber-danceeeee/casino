package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/database"
	"casino/internal/operationadmin"
)

func PublicSiteBanners(c *gin.Context) {
	db := database.GetDB()
	scope := h5TenancyScope(c)
	_, _, loggedIn := h5AuthenticatedUser(db, c)
	language := strings.TrimSpace(c.Query("language"))
	if language == "" {
		language = strings.TrimSpace(c.GetHeader("Accept-Language"))
		if index := strings.Index(language, ","); index >= 0 {
			language = language[:index]
		}
	}
	country := strings.TrimSpace(c.Query("country"))
	if scope.SiteID > 0 && (language == "" || country == "") {
		site := h5BootstrapSite(db, scope.SiteID)
		if language == "" {
			language = site.DefaultLanguage
		}
		if country == "" {
			country = site.CountryCode
		}
	}
	rows, err := operationadmin.New(db).PublicSiteBanners(scope, language, country, loggedIn, time.Now())
	if err != nil {
		h5Error(c, 500, "failed to load banners")
		return
	}
	response.Success(c, gin.H{"list": rows})
}
