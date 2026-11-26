package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/content"
	"casino/internal/database"
	"casino/internal/model"
)

type gameActivityActionRequest struct {
	Action string `json:"action"`
}

func (h *gameActivityEventsHandler) Offers(c *gin.Context) {
	db := database.GetDB()
	scope := h5ContentScope(c)
	var events []model.GameActivityEvents
	now := time.Now()
	_ = content.ApplySite(db.Model(&model.GameActivityEvents{}), "", scope).
		Where("status = ? AND status_code = 'ACTIVE'", 1).
		Where("start_at IS NULL OR start_at <= ?", now).
		Where("end_at IS NULL OR end_at >= ?", now).
		Order("priority DESC, sort ASC, id ASC").
		Find(&events).Error
	country, currency, language := h5PromotionAudience(c, db, scope.SiteID)
	filtered := make([]model.GameActivityEvents, 0, len(events))
	for _, event := range events {
		if h5PromotionCSVContains(event.Countries, country) && h5PromotionCSVContains(event.Currencies, currency) && h5PromotionCSVContains(event.Languages, language) {
			filtered = append(filtered, event)
		}
	}

	var vipLevels []model.GameVipLevels
	_ = content.ApplySite(db.Model(&model.GameVipLevels{}), "", scope).
		Where("status = ?", 1).
		Order("level_no ASC").
		Find(&vipLevels).Error
	response.Success(c, gin.H{
		"events": h5ActivityEventsPayload(filtered),
		"vip": gin.H{
			"levels":   h5VipLevelsPayload(vipLevels),
			"benefits": h5ConfigJSONAssets(c, db, "h5.vip.benefits", []gin.H{}),
		},
	})
}

func (h *gameActivityEventsHandler) GetByIDOrSlug(c *gin.Context) {
	value := strings.TrimSpace(c.Param("id"))
	db := database.GetDB()
	scope := h5ContentScope(c)
	activity, err := h5FindActivePromotion(db, scope, value, time.Now())
	if err != nil {
		h5Error(c, 404, "promotion not found")
		return
	}
	country, currency, language := h5PromotionAudience(c, db, scope.SiteID)
	if !h5PromotionCSVContains(activity.Countries, country) || !h5PromotionCSVContains(activity.Currencies, currency) || !h5PromotionCSVContains(activity.Languages, language) {
		h5Error(c, 404, "promotion not found")
		return
	}

	var assets []model.GameActivityAssets
	_ = db.Where("activity_id = ?", activity.ID).Order("sort ASC, id ASC").Find(&assets).Error
	var rules []model.GameActivityRules
	_ = db.Where("activity_id = ? AND status = ?", activity.ID, 1).Order("sort ASC, id ASC").Find(&rules).Error
	var rewardRules []model.GameActivityRewardRules
	_ = db.Where("activity_id = ? AND status = ?", activity.ID, 1).Order("sort ASC, id ASC").Find(&rewardRules).Error
	var engineConditions []model.GamePromotionParticipationRule
	_ = db.Model(&model.GamePromotionParticipationRule{}).
		Where("site_id IN (0, ?)", activity.SiteID).
		Where("promotion_id=? AND enabled=1 AND deleted_at IS NULL", activity.ID).Order("sort_order,id").Find(&engineConditions).Error
	var engineRewards []model.GamePromotionRewardRule
	_ = db.Model(&model.GamePromotionRewardRule{}).
		Where("site_id IN (0, ?)", activity.SiteID).
		Where("promotion_id=? AND enabled=1 AND deleted_at IS NULL", activity.ID).Order("sort_order,id").Find(&engineRewards).Error

	response.Success(c, gin.H{
		"activity":         h5ActivityEventPayload(activity),
		"assets":           h5ActivityAssetsPayload(assets),
		"assetMap":         h5ActivityAssetMap(assets),
		"rules":            rules,
		"rewardRules":      rewardRules,
		"rewardRows":       h5ActivityRewardRows(rewardRules),
		"engineConditions": engineConditions,
		"engineRewards":    engineRewards,
		"state":            h5PromotionState(c, db, scope, activity),
	})
}

func h5ActivityEventsPayload(events []model.GameActivityEvents) []gin.H {
	result := make([]gin.H, 0, len(events))
	for _, event := range events {
		result = append(result, h5ActivityEventPayload(event))
	}
	return result
}

func h5ActivityEventPayload(event model.GameActivityEvents) gin.H {
	return gin.H{
		"id":            event.ID,
		"activityCode":  event.ActivityCode,
		"slug":          event.Slug,
		"kind":          event.Kind,
		"nameZh":        event.NameZh,
		"nameEn":        event.NameEn,
		"descriptionZh": event.DescriptionZh,
		"descriptionEn": event.DescriptionEn,
		"promotionType": event.PromotionType,
		"triggerType":   event.TriggerType,
		"claimMode":     event.ClaimMode,
		"title":         event.Title,
		"subtitle":      event.Subtitle,
		"rewardText":    event.RewardText,
		"maxRewardText": event.MaxRewardText,
		"note":          event.Note,
		"ctaText":       event.CtaText,
		"actionUrl":     event.ActionURL,
		"heroURL":       h5AssetURL(event.HeroURL),
		"cardImageURL":  h5AssetURL(event.CardImageURL),
		"status":        event.Status,
		"statusCode":    event.StatusCode,
		"claimCount":    event.ClaimCount,
		"sort":          event.Sort,
		"startAt":       event.StartAt,
		"endAt":         event.EndAt,
		"createdAt":     event.CreatedAt,
		"updatedAt":     event.UpdatedAt,
	}
}

func h5VipLevelsPayload(levels []model.GameVipLevels) []gin.H {
	result := make([]gin.H, 0, len(levels))
	for _, level := range levels {
		result = append(result, gin.H{
			"id":                 level.ID,
			"levelNo":            level.LevelNo,
			"name":               level.Name,
			"iconURL":            h5AssetURL(level.IconURL),
			"levelUpReward":      h5DecimalString(level.LevelUpReward),
			"weeklyReward":       h5DecimalString(level.WeeklyReward),
			"monthlyReward":      h5DecimalString(level.MonthlyReward),
			"dailyWithdrawCount": level.DailyWithdrawCount,
			"status":             level.Status,
		})
	}
	return result
}

func (h *gameActivityEventsHandler) Action(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	req := &gameActivityActionRequest{}
	_ = c.ShouldBindJSON(req)
	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "claim"
	}

	db := database.GetDB()
	scope := h5ContentScope(c)
	activity, err := h5FindActivePromotion(db, scope, slug, time.Now())
	if err != nil {
		h5Error(c, 404, "promotion not found")
		return
	}
	country, currency, language := h5PromotionAudience(c, db, scope.SiteID)
	if !h5PromotionCSVContains(activity.Countries, country) || !h5PromotionCSVContains(activity.Currencies, currency) || !h5PromotionCSVContains(activity.Languages, language) {
		h5Error(c, 404, "promotion not found")
		return
	}
	state := h5PromotionState(c, db, scope, activity)
	state["action"] = action
	if activity.ClaimMode == "AUTO" {
		state["message"] = "reward is fulfilled automatically by the configured trigger"
	} else {
		state["status"] = "TRIGGER_NOT_CONNECTED"
		state["message"] = "manual fulfillment is not connected for this promotion"
	}
	response.Success(c, gin.H{"slug": slug, "action": action, "state": state})
}

func h5FindActivePromotion(db *gorm.DB, scope content.Scope, value string, now time.Time) (model.GameActivityEvents, error) {
	var activity model.GameActivityEvents
	base := func(query *gorm.DB) *gorm.DB {
		return query.
			Where("status=1 AND status_code='ACTIVE' AND deleted_at IS NULL").
			Where("start_at IS NULL OR start_at<=?", now).
			Where("end_at IS NULL OR end_at>=?", now)
	}
	query := base(content.ApplySite(db.Model(&model.GameActivityEvents{}), "", scope))
	if id, err := strconv.ParseUint(value, 10, 64); err == nil {
		query = query.Where("id=?", id)
	} else {
		query = query.Where("slug=?", value)
	}
	err := query.First(&activity).Error
	return activity, err
}

func h5PromotionState(c *gin.Context, db *gorm.DB, scope content.Scope, activity model.GameActivityEvents) gin.H {
	state := gin.H{"claimable": false, "claimMode": activity.ClaimMode, "triggerType": activity.TriggerType, "claimed": false}
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		return state
	}
	scope = h5UserContentScope(db, user.ID, scope)
	var claims []model.GamePromotionClaim
	_ = content.ApplySite(db.Model(&model.GamePromotionClaim{}), "", scope).
		Where("promotion_id=? AND user_id=? AND status='COMPLETED' AND deleted_at IS NULL", activity.ID, user.ID).
		Order("id DESC").Limit(20).Find(&claims).Error
	state["claimed"] = len(claims) > 0
	state["claims"] = claims
	return state
}

func h5PromotionAudience(c *gin.Context, db *gorm.DB, siteID uint64) (string, string, string) {
	country, currency := "", ""
	language := strings.TrimSpace(c.Query("language"))
	if siteID > 0 && (country == "" || currency == "" || language == "") {
		site := h5BootstrapSite(db, siteID)
		if country == "" {
			country = strings.ToUpper(site.CountryCode)
		}
		if currency == "" {
			currency = strings.ToUpper(site.CurrencyCode)
		}
		if language == "" {
			language = site.DefaultLanguage
		}
	}
	return country, currency, language
}

func h5PromotionCSVContains(csv string, value string) bool {
	if strings.TrimSpace(csv) == "" || strings.TrimSpace(value) == "" {
		return true
	}
	for _, item := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(item), "ALL") || strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}

func h5DefaultActivityPayload(slug string) gin.H {
	return gin.H{
		"activity": gin.H{
			"slug":    slug,
			"title":   h5Title(slug),
			"status":  1,
			"ctaText": "Join Now",
		},
		"assets":      []gin.H{},
		"assetMap":    gin.H{},
		"rules":       []gin.H{},
		"rewardRules": []gin.H{},
		"rewardRows":  []gin.H{},
		"state":       gin.H{"claimable": false},
	}
}

func h5ActivityAssetsPayload(assets []model.GameActivityAssets) []gin.H {
	result := make([]gin.H, 0, len(assets))
	for _, asset := range assets {
		result = append(result, gin.H{
			"id":         asset.ID,
			"activityID": asset.ActivityID,
			"assetKey":   asset.AssetKey,
			"assetURL":   h5AssetURL(asset.AssetURL),
			"sort":       asset.Sort,
			"createdAt":  asset.CreatedAt,
			"updatedAt":  asset.UpdatedAt,
		})
	}
	return result
}

func h5ActivityAssetMap(assets []model.GameActivityAssets) gin.H {
	result := gin.H{}
	for _, asset := range assets {
		result[asset.AssetKey] = h5AssetURL(asset.AssetURL)
	}
	return result
}

func h5ActivityRewardRows(rules []model.GameActivityRewardRules) []gin.H {
	rows := make([]gin.H, 0, len(rules))
	for _, rule := range rules {
		rows = append(rows, gin.H{
			"label":                 rule.Label,
			"value":                 rule.RewardValue,
			"extra":                 rule.BonusText,
			"minAmount":             h5DecimalString(rule.MinAmount),
			"maxAmount":             h5DecimalString(rule.MaxAmount),
			"rewardCash":            h5DecimalString(rule.RewardCash),
			"rewardBonus":           h5DecimalString(rule.RewardBonus),
			"requirementMultiplier": h5DecimalString(rule.RequirementMultiplier),
		})
	}
	return rows
}

func h5ActivityState(slug string) gin.H {
	switch slug {
	case "invitation-wheel":
		return gin.H{
			"round":          1,
			"remainingSpins": 1,
			"amount":         "0.00",
			"targetAmount":   "1500.00",
			"endsAt":         time.Now().Add(2*time.Hour + 23*time.Minute + 59*time.Second),
		}
	case "daily-deposit-cashback":
		return gin.H{"claimable": false, "cashbackAmount": "0.00"}
	default:
		return gin.H{"claimable": false}
	}
}

func h5ActivityStateAfterAction(slug string, action string) gin.H {
	if slug == "invitation-wheel" && (action == "start" || action == "spin") {
		return gin.H{
			"status":         "spun",
			"remainingSpins": 0,
			"prize": gin.H{
				"label":  "15",
				"amount": "15.00",
				"type":   "cash",
			},
			"amount":       "15.00",
			"targetAmount": "1500.00",
		}
	}
	return gin.H{"status": "accepted"}
}
