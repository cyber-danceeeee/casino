package handler

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/database"
	"casino/internal/systemadmin"
	"casino/internal/tenancy"
)

func PublicAnnouncements(c *gin.Context) {
	scope, language, country, currency, loggedIn, _ := h5SiteContentContext(c)
	rows, err := systemadmin.New(database.GetDB()).PublicAnnouncements(scope, language, country, currency, loggedIn, time.Now())
	if err != nil {
		h5Error(c, 500, "failed to load announcements")
		return
	}
	response.Success(c, gin.H{"list": rows})
}

func PublicAnnouncement(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		h5Error(c, 400, "invalid announcement id")
		return
	}
	scope, language, country, currency, loggedIn, _ := h5SiteContentContext(c)
	row, err := systemadmin.New(database.GetDB()).PublicAnnouncement(scope, id, language, country, currency, loggedIn, time.Now())
	if err != nil {
		h5Error(c, 404, "announcement not found")
		return
	}
	response.Success(c, row)
}

func PublicPopups(c *gin.Context) {
	scope, language, country, currency, loggedIn, viewer := h5SiteContentContext(c)
	rows, err := systemadmin.New(database.GetDB()).PublicPopups(scope, language, country, currency, loggedIn, viewer, time.Now())
	if err != nil {
		h5Error(c, 500, "failed to load popups")
		return
	}
	response.Success(c, gin.H{"items": rows})
}

func TrackPopupImpression(c *gin.Context) { trackPopupEvent(c, "IMPRESSION") }
func TrackPopupClick(c *gin.Context)      { trackPopupEvent(c, "CLICK") }
func TrackPopupDismiss(c *gin.Context)    { trackPopupEvent(c, "DISMISS") }

func trackPopupEvent(c *gin.Context, eventType string) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		h5Error(c, 400, "invalid popup id")
		return
	}
	input := systemadmin.PopupEventInput{}
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		h5Error(c, 400, "invalid request body")
		return
	}
	if input.EventID == "" {
		input.EventID = strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	}
	scope, _, _, _, _, viewer := h5SiteContentContext(c)
	recorded, err := systemadmin.New(database.GetDB()).TrackPopupEvent(scope, id, eventType, input, viewer, c.ClientIP(), c.GetHeader("User-Agent"), time.Now())
	if err != nil {
		h5Error(c, 400, err.Error())
		return
	}
	response.Success(c, gin.H{"recorded": recorded})
}

func h5SiteContentContext(c *gin.Context) (tenancy.Scope, string, string, string, bool, systemadmin.PopupViewer) {
	db := database.GetDB()
	scope := h5TenancyScope(c)
	user, _, loggedIn := h5AuthenticatedUser(db, c)
	language := strings.TrimSpace(c.Query("language"))
	if language == "" {
		language = strings.TrimSpace(c.GetHeader("Accept-Language"))
		if index := strings.Index(language, ","); index >= 0 {
			language = language[:index]
		}
	}
	country := strings.TrimSpace(c.Query("country"))
	currency := strings.TrimSpace(c.Query("currency"))
	if scope.SiteID > 0 && (language == "" || country == "" || currency == "") {
		site := h5BootstrapSite(db, scope.SiteID)
		if language == "" {
			language = site.DefaultLanguage
		}
		if country == "" {
			country = site.CountryCode
		}
		if currency == "" {
			currency = site.CurrencyCode
		}
	}
	deviceID := strings.TrimSpace(h5FirstNonEmpty(c.GetHeader("X-Device-ID"), c.GetHeader("X-Device-Id")))
	if deviceID == "" {
		deviceID = h5UserSessionHash(c.ClientIP() + "|" + c.GetHeader("User-Agent"))[:32]
	}
	sessionID := strings.TrimSpace(c.GetHeader("X-Session-ID"))
	if sessionID == "" && h5BearerToken(c) != "" {
		sessionID = h5UserSessionHash(h5BearerToken(c))[:32]
	}
	if sessionID == "" {
		sessionID = deviceID
	}
	viewer := systemadmin.PopupViewer{SessionID: h5Truncate(sessionID, 128), DeviceID: h5Truncate(deviceID, 128)}
	if loggedIn {
		viewer.PlayerID = user.ID
		viewer.RegisteredAt = &user.CreatedAt
	}
	return scope, language, country, currency, loggedIn, viewer
}
