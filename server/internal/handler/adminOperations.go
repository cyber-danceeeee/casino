package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/adminauth"
	"casino/internal/database"
	"casino/internal/depositconfig"
	"casino/internal/operationadmin"
	"casino/internal/storage"
)

type AdminOperationsHandler struct {
	service       *operationadmin.Service
	depositConfig *depositconfig.Service
}

func NewAdminOperationsHandler() *AdminOperationsHandler {
	db := database.GetDB()
	return &AdminOperationsHandler{service: operationadmin.New(db), depositConfig: depositconfig.New(db)}
}

func (h *AdminOperationsHandler) H5Navigation(c *gin.Context) {
	db := database.GetDB()
	scope := h5ContentScope(c)
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
	data, err := h.service.H5Navigation(scope, language, country, loggedIn, c.Query("position"), time.Now())
	if err != nil {
		h5Error(c, http.StatusInternalServerError, "failed to load navigation entries")
		return
	}
	response.Success(c, data)
}

func (h *AdminOperationsHandler) ListH5Navigation(c *gin.Context) {
	data, err := h.service.ListH5Navigation(operationList(c, h5NavigationOperationKeys))
	h.write(c, data, err)
}

func (h *AdminOperationsHandler) GetH5Navigation(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetH5Navigation(id, adminContentScope(c))
	})
}

func (h *AdminOperationsHandler) CreateH5Navigation(c *gin.Context) {
	var in operationadmin.H5NavigationInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveH5Navigation(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}

func (h *AdminOperationsHandler) UpdateH5Navigation(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.H5NavigationInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveH5Navigation(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}

func (h *AdminOperationsHandler) DeleteH5Navigation(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.service.DeleteH5Navigation(id, operationActor(c)) })
}

func (h *AdminOperationsHandler) CopyH5Navigation(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	newID, err := h.service.CopyH5Navigation(id, operationActor(c))
	h.write(c, gin.H{"id": newID}, err)
}

func (h *AdminOperationsHandler) SortH5Navigation(c *gin.Context) {
	var in operationadmin.H5NavigationSortInput
	if !operationBind(c, &in) {
		return
	}
	h.write(c, nil, h.service.SortH5Navigation(in, operationActor(c)))
}

func (h *AdminOperationsHandler) ListBanners(c *gin.Context) {
	data, err := h.service.ListSiteBanners(operationList(c, []string{"keyword", "publishStatus"}))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetBanner(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetSiteBanner(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) CreateBanner(c *gin.Context) {
	var in operationadmin.SiteBannerInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveSiteBanner(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdateBanner(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.SiteBannerInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveSiteBanner(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) CopyBanner(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	newID, err := h.service.CopySiteBanner(id, operationActor(c))
	h.write(c, gin.H{"id": newID}, err)
}
func (h *AdminOperationsHandler) DeleteBanner(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.service.DeleteSiteBanner(id, operationActor(c)) })
}

func (h *AdminOperationsHandler) SortBanners(c *gin.Context) {
	var in operationadmin.SiteBannerSortInput
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.SortSiteBanners(in, operationActor(c)))
}

func (h *AdminOperationsHandler) ListAnnouncements(c *gin.Context) {
	data, err := h.service.ListAnnouncements(operationList(c, announcementOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetAnnouncement(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetAnnouncement(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) CreateAnnouncement(c *gin.Context) {
	var in operationadmin.AnnouncementInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveAnnouncement(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdateAnnouncement(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.AnnouncementInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveAnnouncement(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) DeleteAnnouncement(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.service.DeleteAnnouncement(id, operationActor(c)) })
}

func (h *AdminOperationsHandler) ListPromotions(c *gin.Context) {
	data, err := h.service.ListPromotions(operationList(c, promotionOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetPromotion(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetPromotion(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) ListPromotionClaims(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	data, err := h.service.ListPromotionClaims(id, operationList(c, []string{"status", "currency", "phone"}))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) CreatePromotion(c *gin.Context) {
	var in operationadmin.PromotionInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SavePromotion(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdatePromotion(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.PromotionInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SavePromotion(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) DeletePromotion(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	h.writeAction(c, h.service.DeletePromotion(id, operationActor(c)))
}
func (h *AdminOperationsHandler) PromotionStatus(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason" binding:"required,max=500"`
	}
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.SetPromotionStatus(id, in.Status, in.Reason, operationActor(c)))
}

func (h *AdminOperationsHandler) ListBonuses(c *gin.Context) {
	data, err := h.service.ListBonuses(operationList(c, bonusOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetBonus(c *gin.Context) { h.detail(c, h.service.GetBonus) }
func (h *AdminOperationsHandler) ActivateBonus(c *gin.Context) {
	h.bonusTransition(c, "ACTIVE", adminauth.PermissionOperationBonusActivate)
}
func (h *AdminOperationsHandler) CancelBonus(c *gin.Context) {
	h.bonusTransition(c, "CANCELLED", adminauth.PermissionOperationBonusCancel)
}
func (h *AdminOperationsHandler) ForfeitBonus(c *gin.Context) {
	h.bonusTransition(c, "FORFEITED", adminauth.PermissionOperationBonusForfeit)
}
func (h *AdminOperationsHandler) bonusTransition(c *gin.Context, target string, permission string) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.ReasonCommand
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.TransitionBonus(id, target, permission, in.Reason, operationActor(c)))
}

func (h *AdminOperationsHandler) ListVIPLevels(c *gin.Context) {
	data, err := h.service.ListVIPLevels(operationList(c, vipOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) ListVIPUsers(c *gin.Context) {
	data, err := h.service.ListVIPUsers(operationList(c, []string{"userID", "username", "level"}))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetVIPLevel(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetVIPLevel(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) CreateVIPLevel(c *gin.Context) {
	var in operationadmin.VIPLevelInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveVIPLevel(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdateVIPLevel(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.VIPLevelInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveVIPLevel(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) AdjustVIP(c *gin.Context) {
	userID, ok := operationParamID(c, "userId")
	if !ok {
		return
	}
	var in operationadmin.VIPAdjustCommand
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.AdjustVIP(userID, in, operationActor(c)))
}

func (h *AdminOperationsHandler) AdjustVIPByNumber(c *gin.Context) {
	var in operationadmin.VIPAdjustCommand
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.AdjustVIPByNumber(c.Param("playerNo"), in, operationActor(c)))
}

func (h *AdminOperationsHandler) ListPromoCodes(c *gin.Context) {
	data, err := h.service.ListPromoCodes(operationList(c, promoCodeOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetPromoCode(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetPromoCode(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) ListPromoCodeUsage(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	data, err := h.service.ListPromoCodeUsage(id, operationList(c, nil))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) CreatePromoCode(c *gin.Context) {
	var in operationadmin.PromoCodeInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SavePromoCode(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdatePromoCode(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.PromoCodeInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SavePromoCode(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) DeletePromoCode(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.service.DeletePromoCode(id, operationActor(c)) })
}
func (h *AdminOperationsHandler) RedeemPromoCode(c *gin.Context) {
	var in operationadmin.RedeemPromoCodeCommand
	if !operationBind(c, &in) {
		return
	}
	in.IdempotencyKey = strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	in.UserID = h5HeaderUint64(c, "X-User-ID")
	if in.UserID == 0 {
		operationError(c, &operationadmin.Error{Code: "AUTHENTICATION_REQUIRED", Message: "authentication required"})
		return
	}
	data, err := h.service.RedeemPromoCode(in)
	h.write(c, data, err)
}

func (h *AdminOperationsHandler) ListNotifications(c *gin.Context) {
	data, err := h.service.ListNotifications(operationList(c, notificationOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) NotificationAudienceOptions(c *gin.Context) {
	data, err := h.service.NotificationAudienceOptions(c.Query("audienceType"), c.Query("keyword"), c.Query("selected"), adminContentScope(c))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetNotification(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetNotification(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) CreateNotification(c *gin.Context) {
	var in operationadmin.NotificationInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveNotification(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdateNotification(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.NotificationInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveNotification(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) ScheduleNotification(c *gin.Context) { h.queueNotification(c, true) }
func (h *AdminOperationsHandler) SendNotification(c *gin.Context)     { h.queueNotification(c, false) }
func (h *AdminOperationsHandler) queueNotification(c *gin.Context, scheduled bool) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in struct {
		ScheduledAt *time.Time `json:"scheduledAt"`
		Reason      string     `json:"reason" binding:"required,max=500"`
	}
	if !operationBind(c, &in) {
		return
	}
	if scheduled && in.ScheduledAt == nil {
		operationError(c, &operationadmin.Error{Code: "SCHEDULE_TIME_REQUIRED", Message: "scheduled time is required"})
		return
	}
	if !scheduled {
		in.ScheduledAt = nil
	}
	h.writeAction(c, h.service.QueueNotification(id, in.ScheduledAt, in.Reason, operationActor(c)))
}
func (h *AdminOperationsHandler) CancelNotification(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.ReasonCommand
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.CancelNotification(id, in.Reason, operationActor(c)))
}
func (h *AdminOperationsHandler) DeleteNotification(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.service.DeleteNotification(id, operationActor(c)) })
}

func (h *AdminOperationsHandler) ListMusic(c *gin.Context) {
	data, err := h.service.ListMusic(operationList(c, musicOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) H5MusicConfig(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	data, err := h.service.H5MusicConfig(adminContentScope(c))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetMusicPlaybackConfig(c *gin.Context) {
	data, err := h.service.GetMusicPlaybackConfig(adminContentScope(c))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) UpdateMusicPlaybackConfig(c *gin.Context) {
	var in operationadmin.MusicPlaybackConfigInput
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.SaveMusicPlaybackConfig(in, operationActor(c)))
}
func (h *AdminOperationsHandler) GetMusic(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetMusic(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) CreateMusic(c *gin.Context) {
	var in operationadmin.MusicInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveMusic(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdateMusic(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.MusicInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveMusic(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) DeleteMusic(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.service.DeleteMusic(id, operationActor(c)) })
}
func (h *AdminOperationsHandler) MusicStatus(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in struct {
		Enabled *bool `json:"enabled" binding:"required"`
	}
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.SetMusicEnabled(id, *in.Enabled, operationActor(c)))
}

func (h *AdminOperationsHandler) DefaultVisibility(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.DefaultVisibilityInput
	if !operationBind(c, &in) {
		return
	}
	resource := c.GetString("operation_default_resource")
	h.writeAction(c, h.service.SetDefaultVisibility(resource, id, in.Enabled, in.Reason, operationActor(c)))
}

func (h *AdminOperationsHandler) ListSocialLinks(c *gin.Context) {
	data, err := h.service.ListSocialLinks(operationList(c, socialLinkOperationKeys))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetSocialLink(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetSocialLink(id, adminContentScope(c))
	})
}
func (h *AdminOperationsHandler) CreateSocialLink(c *gin.Context) {
	var in operationadmin.SocialLinkInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveSocialLink(0, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) UpdateSocialLink(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.SocialLinkInput
	if !operationBind(c, &in) {
		return
	}
	id, err := h.service.SaveSocialLink(id, in, operationActor(c))
	h.write(c, gin.H{"id": id}, err)
}
func (h *AdminOperationsHandler) DeleteSocialLink(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.service.DeleteSocialLink(id, operationActor(c)) })
}

func (h *AdminOperationsHandler) ListVouchers(c *gin.Context) {
	query := operationList(c, voucherOperationKeys)
	data, err := h.depositConfig.ListVouchers(depositconfig.ListQuery{Page: query.Page, PageSize: query.PageSize, Values: query.Values, Scope: query.Scope})
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) GetVoucher(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	data, err := h.depositConfig.GetVoucher(id, adminContentScope(c))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) ListVoucherUsage(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	data, err := h.depositConfig.ListVoucherUsage(id, depositConfigList(c, []string{"userID", "username", "status"}))
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) CreateVoucher(c *gin.Context) { h.saveVoucher(c, 0) }
func (h *AdminOperationsHandler) UpdateVoucher(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	h.saveVoucher(c, id)
}
func (h *AdminOperationsHandler) saveVoucher(c *gin.Context, id uint64) {
	var input depositconfig.VoucherInput
	if !operationBind(c, &input) {
		return
	}
	savedID, err := h.depositConfig.SaveVoucher(id, input, depositConfigActor(c))
	h.write(c, gin.H{"id": savedID}, err)
}
func (h *AdminOperationsHandler) SaveVoucherRewardTiers(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var input struct {
		RewardTiers []depositconfig.RewardTierInput `json:"rewardTiers" binding:"required"`
	}
	if !operationBind(c, &input) {
		return
	}
	h.writeAction(c, h.depositConfig.SaveVoucherTiers(id, input.RewardTiers, depositConfigActor(c)))
}
func (h *AdminOperationsHandler) DeleteVoucher(c *gin.Context) {
	h.action(c, func(id uint64) error { return h.depositConfig.DeleteVoucher(id, depositConfigActor(c)) })
}

func (h *AdminOperationsHandler) GetRegisterRewardConfig(c *gin.Context) {
	data, err := h.service.GetRegisterRewardConfig(adminContentScope(c))
	h.write(c, data, err)
}

func (h *AdminOperationsHandler) UpdateRegisterRewardConfig(c *gin.Context) {
	var in operationadmin.RegisterRewardConfigInput
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.SaveRegisterRewardConfig(in, operationActor(c)))
}

func (h *AdminOperationsHandler) ListAssets(c *gin.Context) {
	data, err := h.service.ListAssets(operationList(c, assetOperationKeys))
	h.write(c, data, err)
}

func (h *AdminOperationsHandler) GetAsset(c *gin.Context) {
	h.detail(c, func(id uint64) (map[string]any, error) {
		return h.service.GetAsset(id, adminContentScope(c))
	})
}

func (h *AdminOperationsHandler) UploadAssetLibrary(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		operationError(c, &operationadmin.Error{Code: "FILE_REQUIRED", Message: "file is required"})
		return
	}
	input := operationadmin.AssetUploadInput{
		AssetType:   c.PostForm("assetType"),
		UsageType:   c.PostForm("usageType"),
		AssetScope:  c.PostForm("assetScope"),
		DisplayName: c.PostForm("displayName"),
		AltText:     c.PostForm("altText"),
		Tags:        c.PostForm("tags"),
	}
	data, err := h.service.UploadAsset(c.Request.Context(), file, input, operationActor(c))
	h.write(c, data, err)
}

func (h *AdminOperationsHandler) UpdateAsset(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.AssetUpdateInput
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.UpdateAsset(id, in, operationActor(c)))
}

func (h *AdminOperationsHandler) DeleteAsset(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	var in operationadmin.ReasonCommand
	if !operationBind(c, &in) {
		return
	}
	h.writeAction(c, h.service.DeleteAsset(id, in.Reason, operationActor(c)))
}

func (h *AdminOperationsHandler) AssetReferences(c *gin.Context) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	data, err := h.service.GetAssetReferences(id, adminContentScope(c))
	h.write(c, data, err)
}

func (h *AdminOperationsHandler) UploadOperationAsset(c *gin.Context) {
	kind := strings.ToLower(strings.TrimSpace(c.PostForm("kind")))
	file, err := c.FormFile("file")
	if err != nil {
		operationError(c, &operationadmin.Error{Code: "FILE_REQUIRED", Message: "file is required"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]map[string]bool{
		"image": {".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".avif": true, ".gif": true, ".svg": true},
		"audio": {".mp3": true, ".wav": true, ".ogg": true, ".m4a": true, ".aac": true, ".mp4": true},
		"video": {".mp4": true, ".webm": true, ".mov": true, ".m4v": true},
	}
	limits := map[string]int64{"image": 10 << 20, "audio": 30 << 20, "video": 100 << 20}
	if _, ok := allowed[kind]; !ok || !allowed[kind][ext] || file.Size <= 0 || file.Size > limits[kind] {
		operationError(c, &operationadmin.Error{Code: "INVALID_ASSET", Message: "unsupported file type or size"})
		return
	}
	name, err := storage.RandomName(ext)
	if err != nil {
		operationError(c, err)
		return
	}
	asset, err := storage.UploadFileHeader(c.Request.Context(), "admin/uploads/"+kind+"/"+name, file)
	if err != nil {
		operationError(c, err)
		return
	}
	if err = h.service.RecordAssetUpload(kind, asset.Key, operationActor(c)); err != nil {
		operationError(c, err)
		return
	}
	response.Success(c, asset)
}

func (h *AdminOperationsHandler) detail(c *gin.Context, fn func(uint64) (map[string]any, error)) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	data, err := fn(id)
	h.write(c, data, err)
}
func (h *AdminOperationsHandler) action(c *gin.Context, fn func(uint64) error) {
	id, ok := operationID(c)
	if !ok {
		return
	}
	h.writeAction(c, fn(id))
}
func (h *AdminOperationsHandler) write(c *gin.Context, data any, err error) {
	if err != nil {
		operationError(c, err)
		return
	}
	response.Success(c, data)
}
func (h *AdminOperationsHandler) writeAction(c *gin.Context, err error) {
	if err != nil {
		operationError(c, err)
		return
	}
	response.Success(c)
}

func operationID(c *gin.Context) (uint64, bool) { return operationParamID(c, "id") }
func operationParamID(c *gin.Context, param string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil || id == 0 {
		operationError(c, &operationadmin.Error{Code: "INVALID_REQUEST", Message: "invalid id"})
		return 0, false
	}
	return id, true
}
func operationBind(c *gin.Context, target any) bool {
	body, err := c.GetRawData()
	if err != nil {
		operationError(c, &operationadmin.Error{Code: "INVALID_REQUEST", Message: "invalid request body"})
		return false
	}
	body = normalizeOperationJSONBody(body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err := c.ShouldBindJSON(target); err != nil {
		operationError(c, &operationadmin.Error{Code: "INVALID_REQUEST", Message: "invalid request body"})
		return false
	}
	return true
}

func normalizeOperationJSONBody(body []byte) []byte {
	var payload any
	if len(body) == 0 || json.Unmarshal(body, &payload) != nil {
		return body
	}
	changed := normalizeOperationTimeFields(payload)
	if !changed {
		return body
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return body
	}
	return normalized
}

func normalizeOperationTimeFields(value any) bool {
	changed := false
	switch item := value.(type) {
	case map[string]any:
		for key, field := range item {
			if text, ok := field.(string); ok && strings.TrimSpace(text) == "" && isOperationNullableTimeKey(key) {
				item[key] = nil
				changed = true
				continue
			}
			if normalizeOperationTimeFields(field) {
				changed = true
			}
		}
	case []any:
		for _, child := range item {
			if normalizeOperationTimeFields(child) {
				changed = true
			}
		}
	}
	return changed
}

func isOperationNullableTimeKey(key string) bool {
	switch key {
	case "startTime", "endTime", "expiresAt", "expirationTime", "scheduledAt", "sentAt", "publishedAt":
		return true
	default:
		return false
	}
}
func operationActor(c *gin.Context) operationadmin.Actor {
	actor, _ := adminauth.FromContext(c)
	correlation := c.GetHeader("X-Request-ID")
	if correlation == "" {
		correlation = c.GetString("request_id")
	}
	return operationadmin.Actor{ID: actor.ID, Username: actor.Username, Scope: adminContentScope(c), IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: correlation}
}
func operationList(c *gin.Context, keys []string) operationadmin.ListQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	values := map[string]string{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values[key] = value
		}
	}
	return operationadmin.ListQuery{Page: page, PageSize: pageSize, Values: values, Scope: adminContentScope(c)}
}
func operationError(c *gin.Context, err error) {
	var configError *depositconfig.DomainError
	if errors.As(err, &configError) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": configError.Message, "data": gin.H{"errorCode": configError.Code}})
		return
	}
	var domain *operationadmin.Error
	if errors.As(err, &domain) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": domain.Message, "data": gin.H{"errorCode": domain.Code}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

var bannerOperationKeys = []string{"title", "position", "language", "platform", "country", "enabled", "displayFrom", "displayTo"}
var announcementOperationKeys = []string{"title", "type", "language", "enabled", "publishedFrom", "publishedTo"}
var promotionOperationKeys = []string{"promotionID", "name", "type", "triggerType", "status", "country", "currency", "startFrom", "startTo", "endFrom", "endTo"}
var bonusOperationKeys = []string{"bonusID", "userID", "username", "promotionID", "bonusType", "currency", "status", "createdFrom", "createdTo", "expirationFrom", "expirationTo"}
var vipOperationKeys = []string{"level", "code", "nameZh", "nameEn", "enabled"}
var promoCodeOperationKeys = []string{"code", "promotionID", "rewardType", "enabled", "usageMin", "usageMax", "createdFrom", "createdTo", "expirationFrom", "expirationTo"}
var notificationOperationKeys = []string{"notificationID", "title", "channel", "audienceType", "status", "scheduledFrom", "scheduledTo", "sentFrom", "sentTo"}
var musicOperationKeys = []string{"code", "name", "scene", "enabled"}
var socialLinkOperationKeys = []string{"platformCode", "label", "enabled"}
var voucherOperationKeys = []string{"voucherCode", "title", "voucherType", "status", "availability", "amountMin", "amountMax", "startFrom", "endTo"}
var h5NavigationOperationKeys = []string{"keyword", "displayPosition", "actionType", "publishStatus"}
var assetOperationKeys = []string{"keyword", "assetType", "usageType", "assetScope", "status"}
