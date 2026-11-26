package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/handler"
)

func operationDefaultVisibility(resource string, h *handler.AdminOperationsHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("operation_default_resource", resource)
		h.DefaultVisibility(c)
	}
}

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) { adminOperationsRouter(group, handler.NewAdminOperationsHandler()) })
}

func adminOperationsRouter(group *gin.RouterGroup, h *handler.AdminOperationsHandler) {
	group.POST("/promo-codes/redeem", h.RedeemPromoCode)
	group.GET("/music/config", h.H5MusicConfig)
	group.GET("/navigation", h.H5Navigation)
	g := group.Group("/admin")
	g.Use(adminauth.Middleware())
	g.GET("/banners", adminauth.Require(adminauth.PermissionOperationBannerList), h.ListBanners)
	g.GET("/banners/:id", adminauth.Require(adminauth.PermissionOperationBannerList), h.GetBanner)
	g.POST("/banners", adminauth.Require(adminauth.PermissionOperationBannerCreate), h.CreateBanner)
	g.PUT("/banners/sort", adminauth.Require(adminauth.PermissionOperationBannerUpdate), h.SortBanners)
	g.PUT("/banners/:id", adminauth.Require(adminauth.PermissionOperationBannerUpdate), h.UpdateBanner)
	g.POST("/banners/:id/copy", adminauth.Require(adminauth.PermissionOperationBannerCreate), h.CopyBanner)
	g.DELETE("/banners/:id", adminauth.Require(adminauth.PermissionOperationBannerDelete), h.DeleteBanner)
	g.PATCH("/banners/:id/default-visibility", adminauth.Require(adminauth.PermissionOperationBannerUpdate), operationDefaultVisibility("BANNER", h))

	g.GET("/announcements", adminauth.Require(adminauth.PermissionOperationAnnouncementList), h.ListAnnouncements)
	g.GET("/announcements/:id", adminauth.Require(adminauth.PermissionOperationAnnouncementList), h.GetAnnouncement)
	g.POST("/announcements", adminauth.Require(adminauth.PermissionOperationAnnouncementCreate), h.CreateAnnouncement)
	g.PUT("/announcements/:id", adminauth.Require(adminauth.PermissionOperationAnnouncementUpdate), h.UpdateAnnouncement)
	g.DELETE("/announcements/:id", adminauth.Require(adminauth.PermissionOperationAnnouncementDelete), h.DeleteAnnouncement)

	g.GET("/promotions", adminauth.Require(adminauth.PermissionOperationPromotionList), h.ListPromotions)
	g.GET("/promotions/:id", adminauth.Require(adminauth.PermissionOperationPromotionList), h.GetPromotion)
	g.GET("/promotions/:id/claims", adminauth.Require(adminauth.PermissionOperationPromotionList), h.ListPromotionClaims)
	g.POST("/promotions", adminauth.Require(adminauth.PermissionOperationPromotionCreate), h.CreatePromotion)
	g.PUT("/promotions/:id", adminauth.Require(adminauth.PermissionOperationPromotionUpdate), h.UpdatePromotion)
	g.PATCH("/promotions/:id/status", adminauth.Require(adminauth.PermissionOperationPromotionStatus), h.PromotionStatus)
	g.DELETE("/promotions/:id", adminauth.Require(adminauth.PermissionOperationPromotionUpdate), h.DeletePromotion)

	g.GET("/bonuses", adminauth.Require(adminauth.PermissionOperationBonusList), h.ListBonuses)
	g.GET("/bonuses/:id", adminauth.Require(adminauth.PermissionOperationBonusView), h.GetBonus)
	g.POST("/bonuses/:id/activate", adminauth.Require(adminauth.PermissionOperationBonusActivate), h.ActivateBonus)
	g.POST("/bonuses/:id/cancel", adminauth.Require(adminauth.PermissionOperationBonusCancel), h.CancelBonus)
	g.POST("/bonuses/:id/forfeit", adminauth.Require(adminauth.PermissionOperationBonusForfeit), h.ForfeitBonus)

	g.GET("/vip-levels", adminauth.Require(adminauth.PermissionOperationVIPList), h.ListVIPLevels)
	g.GET("/vip-levels/:id", adminauth.Require(adminauth.PermissionOperationVIPList), h.GetVIPLevel)
	g.GET("/vip-users", adminauth.Require(adminauth.PermissionOperationVIPList), h.ListVIPUsers)
	g.POST("/vip-users/by-number/:playerNo/adjust", adminauth.Require(adminauth.PermissionOperationVIPAdjust), h.AdjustVIPByNumber)
	g.POST("/vip-levels", adminauth.Require(adminauth.PermissionOperationVIPUpdate), h.CreateVIPLevel)
	g.PUT("/vip-levels/:id", adminauth.Require(adminauth.PermissionOperationVIPUpdate), h.UpdateVIPLevel)
	g.POST("/vip-users/:userId/adjust", adminauth.Require(adminauth.PermissionOperationVIPAdjust), h.AdjustVIP)

	g.GET("/promo-codes", adminauth.Require(adminauth.PermissionOperationPromoCodeList), h.ListPromoCodes)
	g.GET("/promo-codes/:id", adminauth.Require(adminauth.PermissionOperationPromoCodeList), h.GetPromoCode)
	g.GET("/promo-codes/:id/usage-records", adminauth.Require(adminauth.PermissionOperationPromoCodeList), h.ListPromoCodeUsage)
	g.POST("/promo-codes", adminauth.Require(adminauth.PermissionOperationPromoCodeCreate), h.CreatePromoCode)
	g.PUT("/promo-codes/:id", adminauth.Require(adminauth.PermissionOperationPromoCodeUpdate), h.UpdatePromoCode)
	g.DELETE("/promo-codes/:id", adminauth.Require(adminauth.PermissionOperationPromoCodeDelete), h.DeletePromoCode)

	g.GET("/notifications", adminauth.Require(adminauth.PermissionOperationNotificationList), h.ListNotifications)
	g.GET("/notifications/audience-options", adminauth.Require(adminauth.PermissionOperationNotificationList), h.NotificationAudienceOptions)
	g.GET("/notifications/:id", adminauth.Require(adminauth.PermissionOperationNotificationList), h.GetNotification)
	g.POST("/notifications", adminauth.Require(adminauth.PermissionOperationNotificationCreate), h.CreateNotification)
	g.PUT("/notifications/:id", adminauth.Require(adminauth.PermissionOperationNotificationUpdate), h.UpdateNotification)
	g.POST("/notifications/:id/schedule", adminauth.Require(adminauth.PermissionOperationNotificationSend), h.ScheduleNotification)
	g.POST("/notifications/:id/send", adminauth.Require(adminauth.PermissionOperationNotificationSend), h.SendNotification)
	g.POST("/notifications/:id/cancel", adminauth.Require(adminauth.PermissionOperationNotificationCancel), h.CancelNotification)
	g.DELETE("/notifications/:id", adminauth.Require(adminauth.PermissionOperationNotificationUpdate), h.DeleteNotification)

	g.GET("/assets", adminauth.Require(adminauth.PermissionOperationAssetList), h.ListAssets)
	g.GET("/assets/:id", adminauth.Require(adminauth.PermissionOperationAssetList), h.GetAsset)
	g.GET("/assets/:id/references", adminauth.Require(adminauth.PermissionOperationAssetList), h.AssetReferences)
	g.POST("/assets/upload", adminauth.Require(adminauth.PermissionOperationAssetUpload), h.UploadAssetLibrary)
	g.PUT("/assets/:id", adminauth.Require(adminauth.PermissionOperationAssetUpdate), h.UpdateAsset)
	g.DELETE("/assets/:id", adminauth.Require(adminauth.PermissionOperationAssetDelete), h.DeleteAsset)
	g.POST("/operation-assets/upload", adminauth.Require(adminauth.PermissionOperationAssetUpload), h.UploadOperationAsset)

	g.GET("/music-settings", adminauth.Require(adminauth.PermissionOperationMusicList), h.GetMusicPlaybackConfig)
	g.PUT("/music-settings", adminauth.Require(adminauth.PermissionOperationMusicUpdate), h.UpdateMusicPlaybackConfig)
	g.GET("/music", adminauth.Require(adminauth.PermissionOperationMusicList), h.ListMusic)
	g.GET("/music/:id", adminauth.Require(adminauth.PermissionOperationMusicList), h.GetMusic)
	g.POST("/music", adminauth.Require(adminauth.PermissionOperationMusicCreate), h.CreateMusic)
	g.PUT("/music/:id", adminauth.Require(adminauth.PermissionOperationMusicUpdate), h.UpdateMusic)
	g.PATCH("/music/:id/status", adminauth.Require(adminauth.PermissionOperationMusicUpdate), h.MusicStatus)
	g.DELETE("/music/:id", adminauth.Require(adminauth.PermissionOperationMusicDelete), h.DeleteMusic)

	g.GET("/social-links", adminauth.Require(adminauth.PermissionOperationSocialLinkList), h.ListSocialLinks)
	g.GET("/social-links/:id", adminauth.Require(adminauth.PermissionOperationSocialLinkList), h.GetSocialLink)
	g.POST("/social-links", adminauth.Require(adminauth.PermissionOperationSocialLinkCreate), h.CreateSocialLink)
	g.PUT("/social-links/:id", adminauth.Require(adminauth.PermissionOperationSocialLinkUpdate), h.UpdateSocialLink)
	g.DELETE("/social-links/:id", adminauth.Require(adminauth.PermissionOperationSocialLinkDelete), h.DeleteSocialLink)

	g.GET("/h5-navigation", adminauth.Require(adminauth.PermissionOperationH5NavigationList), h.ListH5Navigation)
	g.PUT("/h5-navigation/sort", adminauth.Require(adminauth.PermissionOperationH5NavigationUpdate), h.SortH5Navigation)
	g.GET("/h5-navigation/:id", adminauth.Require(adminauth.PermissionOperationH5NavigationList), h.GetH5Navigation)
	g.POST("/h5-navigation", adminauth.Require(adminauth.PermissionOperationH5NavigationCreate), h.CreateH5Navigation)
	g.PUT("/h5-navigation/:id", adminauth.Require(adminauth.PermissionOperationH5NavigationUpdate), h.UpdateH5Navigation)
	g.POST("/h5-navigation/:id/copy", adminauth.Require(adminauth.PermissionOperationH5NavigationCreate), h.CopyH5Navigation)
	g.DELETE("/h5-navigation/:id", adminauth.Require(adminauth.PermissionOperationH5NavigationDelete), h.DeleteH5Navigation)

	g.GET("/vouchers", adminauth.Require(adminauth.PermissionOperationVoucherList), h.ListVouchers)
	g.GET("/vouchers/:id", adminauth.Require(adminauth.PermissionOperationVoucherList), h.GetVoucher)
	g.GET("/vouchers/:id/usage-records", adminauth.Require(adminauth.PermissionOperationVoucherUsageView), h.ListVoucherUsage)
	g.POST("/vouchers", adminauth.Require(adminauth.PermissionOperationVoucherCreate), h.CreateVoucher)
	g.PUT("/vouchers/:id", adminauth.Require(adminauth.PermissionOperationVoucherUpdate), h.UpdateVoucher)
	g.PUT("/vouchers/:id/reward-tiers", adminauth.Require(adminauth.PermissionOperationVoucherTierUpdate), h.SaveVoucherRewardTiers)
	g.DELETE("/vouchers/:id", adminauth.Require(adminauth.PermissionOperationVoucherDelete), h.DeleteVoucher)

	g.PATCH("/promotions/:id/default-visibility", adminauth.Require(adminauth.PermissionOperationPromotionUpdate), operationDefaultVisibility("PROMOTION", h))
	g.PATCH("/vip-levels/:id/default-visibility", adminauth.Require(adminauth.PermissionOperationVIPUpdate), operationDefaultVisibility("VIP_LEVEL", h))
	g.PATCH("/music/:id/default-visibility", adminauth.Require(adminauth.PermissionOperationMusicUpdate), operationDefaultVisibility("MUSIC", h))
	g.PATCH("/social-links/:id/default-visibility", adminauth.Require(adminauth.PermissionOperationSocialLinkUpdate), operationDefaultVisibility("SOCIAL_LINK", h))
	g.PATCH("/vouchers/:id/default-visibility", adminauth.Require(adminauth.PermissionOperationVoucherUpdate), operationDefaultVisibility("VOUCHER", h))

	g.GET("/register-reward", adminauth.Require(adminauth.PermissionOperationRegisterRewardView), h.GetRegisterRewardConfig)
	g.PUT("/register-reward", adminauth.Require(adminauth.PermissionOperationRegisterRewardUpdate), h.UpdateRegisterRewardConfig)
}
