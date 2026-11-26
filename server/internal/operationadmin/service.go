package operationadmin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/model"
	"casino/internal/operationdefaults"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type siteDimensions struct {
	CountryCode     string `gorm:"column:country_code"`
	CurrencyCode    string `gorm:"column:currency_code"`
	DefaultLanguage string `gorm:"column:default_language"`
}

func loadSiteDimensions(db *gorm.DB, scope content.Scope) (siteDimensions, error) {
	var row siteDimensions
	err := db.Table("game_site_config").Select("country_code,currency_code,default_language").
		Where("id = ? AND deleted_at IS NULL", scope.SiteID).Take(&row).Error
	if err != nil {
		return row, newError("SITE_NOT_FOUND", "H5 was not found")
	}
	return row, nil
}

type bannerRecord struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	SiteID         uint64     `json:"siteID"`
	BannerCode     string     `json:"bannerCode"`
	Position       string     `json:"position"`
	Platform       string     `json:"platform"`
	Countries      string     `json:"countries"`
	Languages      string     `json:"languages"`
	Title          string     `json:"-"`
	TitleZh        string     `json:"titleZh"`
	TitleEn        string     `json:"titleEn"`
	ImageURL       string     `json:"imageUrl"`
	MobileImageURL string     `json:"mobileImageUrl"`
	LinkType       string     `json:"-"`
	LinkURL        string     `json:"-"`
	TargetType     string     `json:"targetType"`
	TargetValue    string     `json:"targetValue"`
	Sort           int        `json:"-"`
	SortOrder      int        `json:"sortOrder"`
	Status         int        `json:"-"`
	Enabled        bool       `json:"enabled"`
	StartAt        *time.Time `json:"startTime"`
	EndAt          *time.Time `json:"endTime"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"-"`
}

func (bannerRecord) TableName() string { return "game_banners" }

type announcementRecord struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	SiteID      uint64     `json:"siteID"`
	NoticeCode  string     `json:"announcementID"`
	NoticeType  string     `json:"type"`
	Priority    int        `json:"priority"`
	Countries   string     `json:"countries"`
	Platforms   string     `json:"platforms"`
	Languages   string     `json:"languages"`
	Title       string     `json:"-"`
	TitleZh     string     `json:"titleZh"`
	TitleEn     string     `json:"titleEn"`
	ContentHTML string     `json:"-"`
	ContentZh   string     `json:"contentZh"`
	ContentEn   string     `json:"contentEn"`
	Status      int        `json:"-"`
	Enabled     bool       `json:"enabled"`
	StartAt     *time.Time `json:"startTime"`
	EndAt       *time.Time `json:"endTime"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"-"`
}

func (announcementRecord) TableName() string { return "game_notices" }

type promotionRecord struct {
	operationdefaults.Metadata
	ID uint64 `gorm:"primaryKey" json:"id"`

	SiteID          uint64           `json:"siteID"`
	ActivityCode    string           `json:"promotionID"`
	Slug            string           `json:"slug"`
	Kind            string           `json:"-"`
	PromotionType   string           `json:"type"`
	TriggerType     string           `json:"triggerType"`
	ClaimMode       string           `json:"claimMode"`
	StackingGroup   string           `json:"stackingGroup"`
	Priority        int              `json:"priority"`
	Title           string           `json:"-"`
	NameZh          string           `json:"nameZh"`
	NameEn          string           `json:"nameEn"`
	Subtitle        string           `json:"-"`
	DescriptionZh   string           `json:"descriptionZh"`
	DescriptionEn   string           `json:"descriptionEn"`
	ActionURL       string           `gorm:"column:action_url" json:"actionUrl"`
	Countries       string           `json:"countries"`
	Currencies      string           `json:"currencies"`
	Languages       string           `json:"languages"`
	Status          int              `json:"-"`
	StatusCode      string           `json:"status"`
	TotalBudget     *decimal.Decimal `json:"totalBudget"`
	PerUserLimit    uint             `json:"perUserLimit"`
	TotalClaimLimit uint             `json:"totalClaimLimit"`
	ClaimCount      uint             `json:"claimCount"`
	RewardAmount    *decimal.Decimal `json:"rewardAmount"`
	StartAt         *time.Time       `json:"startTime"`
	EndAt           *time.Time       `json:"endTime"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
	DeletedAt       *time.Time       `json:"-"`
}

func (promotionRecord) TableName() string { return "game_activity_events" }

type eligibleGameRecord struct {
	ID               uint64 `gorm:"column:id" json:"id"`
	Name             string `gorm:"column:name" json:"name"`
	ProviderID       uint64 `gorm:"column:provider_id" json:"providerID"`
	ProviderGameCode string `gorm:"column:provider_game_code" json:"providerGameCode"`
}

type promotionClaimRecord struct {
	model.GamePromotionClaim
	PromotionCode string `gorm:"column:promotion_code" json:"promotionCode"`
	PromotionName string `gorm:"column:promotion_name" json:"promotionName"`
	Phone         string `gorm:"column:phone" json:"phone"`
}

type vipRecord struct {
	operationdefaults.Metadata
	ID uint64 `gorm:"primaryKey" json:"id"`

	SiteID             uint64           `json:"siteID"`
	LevelNo            uint             `json:"level"`
	Code               string           `json:"code"`
	Name               string           `json:"-"`
	NameZh             string           `json:"nameZh"`
	NameEn             string           `json:"nameEn"`
	IconURL            string           `json:"icon"`
	RequiredExperience uint64           `json:"requiredExperience"`
	RequiredDeposit    *decimal.Decimal `json:"requiredDeposit"`
	RequiredValidBet   *decimal.Decimal `json:"requiredValidBet"`
	WithdrawalLimit    *decimal.Decimal `json:"withdrawalLimit"`
	CashbackRate       *decimal.Decimal `json:"cashbackRate"`
	BirthdayReward     *decimal.Decimal `json:"birthdayReward"`
	MonthlyReward      *decimal.Decimal `json:"monthlyReward"`
	LevelUpReward      *decimal.Decimal `json:"levelUpReward"`
	WeeklyReward       *decimal.Decimal `json:"weeklyReward"`
	DailyWithdrawCount uint             `json:"dailyWithdrawCount"`
	RetainPeriodDays   uint             `gorm:"-" json:"retainPeriodDays"`
	RetainDeposit      string           `gorm:"-" json:"retainDeposit"`
	RetainValidBet     string           `gorm:"-" json:"retainValidBet"`
	SortOrder          int              `json:"sortOrder"`
	Status             int              `json:"-"`
	Enabled            bool             `json:"enabled"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
	DeletedAt          *time.Time       `json:"-"`
	UserCount          int64            `gorm:"-" json:"userCount"`
}

func (vipRecord) TableName() string { return "game_vip_levels" }

func normalizedQuery(q ListQuery) (int, int, int) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q.Page, q.PageSize, (q.Page - 1) * q.PageSize
}

func page(db *gorm.DB, q ListQuery, out any) (PageResult, error) {
	p, ps, offset := normalizedQuery(q)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return PageResult{}, err
	}
	if err := db.Offset(offset).Limit(ps).Find(out).Error; err != nil {
		return PageResult{}, err
	}
	return PageResult{List: out, Total: total, Page: p, PageSize: ps}, nil
}

func contains(db *gorm.DB, column string, value string) *gorm.DB {
	value = strings.TrimSpace(value)
	if value == "" {
		return db
	}
	return db.Where(column+" LIKE ?", "%"+value+"%")
}
func equal(db *gorm.DB, column string, value string) *gorm.DB {
	value = strings.TrimSpace(value)
	if value == "" {
		return db
	}
	return db.Where(column+" = ?", value)
}
func boolFilter(db *gorm.DB, column string, value string) *gorm.DB {
	if value == "true" || value == "1" {
		return db.Where(column + " = 1")
	}
	if value == "false" || value == "0" {
		return db.Where(column + " = 0")
	}
	return db
}
func rangeFilter(db *gorm.DB, column string, from string, to string) *gorm.DB {
	if from != "" {
		db = db.Where(column+" >= ?", from)
	}
	if to != "" {
		db = db.Where(column+" <= ?", to)
	}
	return db
}
func validateTime(start *time.Time, end *time.Time) error {
	if start != nil && end != nil && end.Before(*start) {
		return newError("INVALID_TIME_RANGE", "end time must be after start time")
	}
	return nil
}
func parseDecimal(value string, field string) (decimal.Decimal, error) {
	v, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil || v.IsNegative() {
		return decimal.Zero, newError("INVALID_AMOUNT", field+" must be a non-negative decimal")
	}
	return v, nil
}
func generatedCode(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }

func validateExternalURL(targetType string, value string) error {
	if targetType != "EXTERNAL_URL" {
		return nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return newError("INVALID_EXTERNAL_URL", "external URL must use http or https")
	}
	return nil
}

func operationStatus(enabled bool) int {
	if enabled {
		return 1
	}
	return 2
}
func audit(tx *gorm.DB, actor Actor, permission string, action string, target string, id uint64, before any, after any, reason string) error {
	b, _ := json.Marshal(before)
	a, _ := json.Marshal(after)
	return tx.Create(&model.GameAdminOperationLogs{SiteID: actor.Scope.SiteID, OperatorID: actor.ID, OperatorUsername: actor.Username, Permission: permission, Action: action, TargetType: target, TargetID: id, BeforeState: string(b), AfterState: string(a), Reason: reason, IPAddress: actor.IP, UserAgent: actor.UserAgent, CorrelationID: actor.CorrelationID}).Error
}

func (s *Service) operationLogs(target string, id uint64, scope content.Scope) ([]model.GameAdminOperationLogs, error) {
	var rows []model.GameAdminOperationLogs
	err := content.ApplySite(s.db.Model(&model.GameAdminOperationLogs{}), "", scope).
		Where("target_type=? AND target_id=?", target, id).
		Order("id DESC").
		Limit(100).
		Find(&rows).Error
	return rows, err
}
func notFound(err error, code string) error {
	if err == gorm.ErrRecordNotFound {
		return newError(code, "record not found")
	}
	return err
}

func (s *Service) ListBanners(q ListQuery) (PageResult, error) {
	var rows []bannerRecord
	db := content.ApplySite(s.db.Model(&bannerRecord{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "title_zh", q.Values["title"])
	db = equal(db, "position", q.Values["position"])
	db = equal(db, "platform", q.Values["platform"])
	db = contains(db, "languages", q.Values["language"])
	db = contains(db, "countries", q.Values["country"])
	db = boolFilter(db, "enabled", q.Values["enabled"])
	db = rangeFilter(db, "start_at", q.Values["displayFrom"], q.Values["displayTo"])
	result, err := page(db.Order("sort_order ASC,id DESC"), q, &rows)
	for i := range rows {
		signBannerAssets(&rows[i])
	}
	result.List = rows
	return result, err
}
func (s *Service) GetBanner(id uint64, scope content.Scope) (map[string]any, error) {
	var row bannerRecord
	if err := content.ApplySite(s.db.Model(&bannerRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "BANNER_NOT_FOUND")
	}
	signBannerAssets(&row)
	logs, _ := s.operationLogs("banner", id, scope)
	return map[string]any{"overview": row, "displayRules": row, "operationLogs": logs}, nil
}
func (s *Service) SaveBanner(id uint64, in BannerInput, actor Actor) (uint64, error) {
	if err := validateTime(in.StartTime, in.EndTime); err != nil {
		return 0, err
	}
	if err := validateExternalURL(in.TargetType, in.TargetValue); err != nil {
		return 0, err
	}
	return id, s.db.Transaction(func(tx *gorm.DB) error {
		var before bannerRecord
		action := "create"
		permission := "operation:banner:create"
		if id > 0 {
			if err := content.ApplySite(tx.Model(&bannerRecord{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "BANNER_NOT_FOUND")
			}
			action = "update"
			permission = "operation:banner:update"
		} else {
			before.BannerCode = generatedCode("BNR")
		}
		row := bannerRecord{ID: id, SiteID: actor.Scope.SiteID, BannerCode: before.BannerCode, Position: in.Position, Platform: in.Platform, Countries: in.Countries, Languages: in.Languages, Title: in.TitleEn, TitleZh: in.TitleZh, TitleEn: in.TitleEn, ImageURL: operationAssetRef(in.ImageURL), MobileImageURL: operationAssetRef(in.MobileImageURL), LinkType: strings.ToLower(in.TargetType), LinkURL: in.TargetValue, TargetType: in.TargetType, TargetValue: in.TargetValue, Sort: in.SortOrder, SortOrder: in.SortOrder, Status: operationStatus(in.Enabled), Enabled: in.Enabled, StartAt: in.StartTime, EndAt: in.EndTime}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&bannerRecord{}), "", actor.Scope).Where("id=?", id).Select(
			"position", "platform", "countries", "languages", "title", "title_zh", "title_en",
			"image_url", "mobile_image_url", "link_type", "link_url", "target_type", "target_value",
			"sort", "sort_order", "status", "enabled", "start_at", "end_at",
		).Updates(row).Error; err != nil {
			return err
		}
		if err := content.ApplySite(tx.Model(&bannerRecord{}), "", actor.Scope).Where("id=?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "banner", id, before, row, "")
	})
}
func (s *Service) CopyBanner(id uint64, actor Actor) (uint64, error) {
	var row bannerRecord
	if err := content.ApplySite(s.db.Model(&bannerRecord{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return 0, notFound(err, "BANNER_NOT_FOUND")
	}
	row.ID = 0
	row.CreatedAt = time.Time{}
	row.UpdatedAt = time.Time{}
	row.BannerCode = generatedCode("BNR")
	row.Enabled = false
	row.Status = 2
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:banner:create", "copy", "banner", row.ID, nil, row, "")
	}); err != nil {
		return 0, err
	}
	return row.ID, nil
}
func (s *Service) DeleteBanner(id uint64, actor Actor) error {
	return s.softDelete("game_banners", "banner", id, "operation:banner:delete", actor)
}

func (s *Service) ListAnnouncements(q ListQuery) (PageResult, error) {
	var rows []announcementRecord
	db := content.ApplySite(s.db.Model(&announcementRecord{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "title_zh", q.Values["title"])
	db = equal(db, "notice_type", q.Values["type"])
	db = contains(db, "languages", q.Values["language"])
	db = boolFilter(db, "enabled", q.Values["enabled"])
	db = rangeFilter(db, "start_at", q.Values["publishedFrom"], q.Values["publishedTo"])
	return page(db.Order("priority DESC,id DESC"), q, &rows)
}
func (s *Service) GetAnnouncement(id uint64, scope content.Scope) (map[string]any, error) {
	var row announcementRecord
	if err := content.ApplySite(s.db.Model(&announcementRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "ANNOUNCEMENT_NOT_FOUND")
	}
	logs, _ := s.operationLogs("announcement", id, scope)
	return map[string]any{"basicInformation": row, "chineseContent": map[string]any{"title": row.TitleZh, "content": row.ContentZh}, "englishContent": map[string]any{"title": row.TitleEn, "content": row.ContentEn}, "displayRules": row, "operationLogs": logs}, nil
}
func (s *Service) SaveAnnouncement(id uint64, in AnnouncementInput, actor Actor) (uint64, error) {
	if err := validateTime(in.StartTime, in.EndTime); err != nil {
		return 0, err
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before announcementRecord
		action := "create"
		permission := "operation:announcement:create"
		if id > 0 {
			if err := content.ApplySite(tx.Model(&announcementRecord{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "ANNOUNCEMENT_NOT_FOUND")
			}
			action = "update"
			permission = "operation:announcement:update"
		} else {
			before.NoticeCode = generatedCode("ANN")
		}
		row := announcementRecord{ID: id, SiteID: actor.Scope.SiteID, NoticeCode: before.NoticeCode, NoticeType: in.Type, Priority: in.Priority, Countries: in.Countries, Platforms: in.Platforms, Languages: in.Languages, Title: in.TitleEn, TitleZh: in.TitleZh, TitleEn: in.TitleEn, ContentHTML: in.ContentEn, ContentZh: in.ContentZh, ContentEn: in.ContentEn, Status: operationStatus(in.Enabled), Enabled: in.Enabled, StartAt: in.StartTime, EndAt: in.EndTime}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&announcementRecord{}), "", actor.Scope).Where("id=?", id).Select(
			"notice_type", "priority", "countries", "platforms", "languages", "title", "title_zh", "title_en",
			"content_html", "content_zh", "content_en", "status", "enabled", "start_at", "end_at",
		).Updates(row).Error; err != nil {
			return err
		}
		if err := content.ApplySite(tx.Model(&announcementRecord{}), "", actor.Scope).Where("id=?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "announcement", id, before, row, "")
	})
	return id, err
}
func (s *Service) DeleteAnnouncement(id uint64, actor Actor) error {
	return s.softDelete("game_notices", "announcement", id, "operation:announcement:delete", actor)
}

func (s *Service) ListPromotions(q ListQuery) (PageResult, error) {
	var rows []promotionRecord
	db := content.ApplySite(s.db.Model(&promotionRecord{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "activity_code", q.Values["promotionID"])
	db = contains(db, "name_zh", q.Values["name"])
	db = equal(db, "promotion_type", q.Values["type"])
	db = equal(db, "trigger_type", q.Values["triggerType"])
	if !q.Scope.SiteScoped() {
		db = equal(db, "status_code", q.Values["status"])
	}
	db = contains(db, "countries", q.Values["country"])
	db = contains(db, "currencies", q.Values["currency"])
	db = rangeFilter(db, "start_at", q.Values["startFrom"], q.Values["startTo"])
	db = rangeFilter(db, "end_at", q.Values["endFrom"], q.Values["endTo"])
	result, err := page(db.Order("id DESC"), q, &rows)
	if err != nil {
		return result, err
	}
	for i := range rows {
		rows[i].Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: rows[i].Status == 1 && rows[i].StatusCode == "ACTIVE"}
	}
	result.List = rows
	return result, nil
}
func (s *Service) GetPromotion(id uint64, scope content.Scope) (map[string]any, error) {
	var row promotionRecord
	if err := content.ApplySite(s.db.Model(&promotionRecord{}), "", scope).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "PROMOTION_NOT_FOUND")
	}
	row.Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: row.Status == 1 && row.StatusCode == "ACTIVE"}
	var participation []model.GamePromotionParticipationRule
	var rewards []model.GamePromotionRewardRule
	var games []eligibleGameRecord
	if err := content.ApplySite(s.db.Model(&model.GamePromotionParticipationRule{}), "", scope).Where("promotion_id=? AND deleted_at IS NULL", id).Order("sort_order,id").Find(&participation).Error; err != nil {
		return nil, err
	}
	if err := content.ApplySite(s.db.Model(&model.GamePromotionRewardRule{}), "", scope).Where("promotion_id=? AND deleted_at IS NULL", id).Order("sort_order,id").Find(&rewards).Error; err != nil {
		return nil, err
	}
	if err := content.ApplySite(s.db.Table("game_promotion_eligible_games r"), "r", scope).
		Scopes(func(db *gorm.DB) *gorm.DB { return content.ApplySite(db, "g", scope) }).
		Select("g.id,g.name,g.provider_id,g.provider_game_code").
		Joins("JOIN game_games g ON g.id=r.game_id AND g.deleted_at IS NULL").
		Where("r.promotion_id=? AND r.deleted_at IS NULL", id).
		Order("r.id").
		Scan(&games).Error; err != nil {
		return nil, err
	}
	claims, _ := s.ListPromotionClaims(id, ListQuery{Page: 1, PageSize: 50, Scope: scope})
	logs, _ := s.operationLogs("promotion", id, scope)
	return map[string]any{"overview": row, "participationRules": participation, "rewardRules": rewards, "eligibleGames": games, "userRecords": claims.List, "statistics": map[string]any{"claimCount": row.ClaimCount, "rewardAmount": row.RewardAmount}, "operationLogs": logs}, nil
}

func (s *Service) ListPromotionClaims(id uint64, q ListQuery) (PageResult, error) {
	var promotionCount int64
	if err := content.ApplySite(s.db.Model(&promotionRecord{}), "", q.Scope).Where("id=? AND deleted_at IS NULL", id).Count(&promotionCount).Error; err != nil {
		return PageResult{}, err
	}
	if promotionCount == 0 {
		return PageResult{}, newError("PROMOTION_NOT_FOUND", "record not found")
	}
	var rows []promotionClaimRecord
	db := content.ApplySite(s.db.Table("game_promotion_claims c"), "c", q.Scope).
		Select(`c.*, p.activity_code AS promotion_code, COALESCE(NULLIF(p.name_en,''),NULLIF(p.name_zh,''),p.title) AS promotion_name,
			CONCAT(COALESCE(u.phone_country_code,''),COALESCE(u.phone,'')) AS phone`).
		Joins("JOIN game_activity_events AS p ON p.id = c.promotion_id AND p.site_id = c.site_id AND p.deleted_at IS NULL").
		Joins("JOIN game_users AS u ON u.id = c.user_id AND u.site_id = c.site_id AND u.deleted_at IS NULL").
		Where("c.promotion_id=? AND c.deleted_at IS NULL", id)
	db = equal(db, "c.status", q.Values["status"])
	db = equal(db, "c.currency", q.Values["currency"])
	db = contains(db, "CONCAT(COALESCE(u.phone_country_code,''),COALESCE(u.phone,''))", q.Values["phone"])
	return page(db.Order("c.id DESC"), q, &rows)
}
func (s *Service) SavePromotion(id uint64, in PromotionInput, actor Actor) (uint64, error) {
	normalizePromotionInput(&in)
	if err := validatePromotionInput(in); err != nil {
		return 0, err
	}
	if err := validateTime(in.StartTime, in.EndTime); err != nil {
		return 0, err
	}
	budget, err := parseDecimal(in.TotalBudget, "totalBudget")
	if err != nil {
		return 0, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before promotionRecord
		action := "create"
		permission := "operation:promotion:create"
		if id > 0 {
			if err := content.ApplySite(tx.Model(&promotionRecord{}), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "PROMOTION_NOT_FOUND")
			}
			action = "update"
			permission = "operation:promotion:update"
		} else {
			before.ActivityCode = in.Code
		}
		rewardAmount := decimal.Zero
		claimCount := uint(0)
		if id > 0 {
			rewardAmount = decimalValue(before.RewardAmount)
			claimCount = before.ClaimCount
		}
		row := promotionRecord{ID: id, SiteID: actor.Scope.SiteID, ActivityCode: in.Code, Slug: strings.ToLower(strings.ReplaceAll(in.Code, "_", "-")), Kind: strings.ToLower(in.Type), PromotionType: in.Type, TriggerType: in.TriggerType, ClaimMode: in.ClaimMode, StackingGroup: in.StackingGroup, Priority: in.Priority, Title: in.NameEn, NameZh: in.NameZh, NameEn: in.NameEn, DescriptionZh: in.DescriptionZh, DescriptionEn: in.DescriptionEn, ActionURL: strings.TrimSpace(in.ActionURL), Countries: in.Countries, Currencies: in.Currencies, Languages: in.Languages, Status: promotionLegacyStatus(in.Status), StatusCode: in.Status, TotalBudget: &budget, PerUserLimit: in.PerUserLimit, TotalClaimLimit: in.TotalClaimLimit, ClaimCount: claimCount, RewardAmount: &rewardAmount, StartAt: in.StartTime, EndAt: in.EndTime}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&promotionRecord{}), "", actor.Scope).Where("id=?", id).Select(
			"activity_code", "slug", "kind", "promotion_type", "trigger_type", "claim_mode", "stacking_group", "priority", "title", "name_zh", "name_en",
			"description_zh", "description_en", "action_url", "countries", "currencies", "status", "status_code",
			"languages", "total_budget", "per_user_limit", "total_claim_limit", "start_at", "end_at",
		).Updates(row).Error; err != nil {
			return err
		}
		if err := replacePromotionRules(tx, id, in, actor.Scope); err != nil {
			return err
		}
		if err := content.ApplySite(tx.Model(&promotionRecord{}), "", actor.Scope).Where("id=?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "promotion", id, before, row, "")
	})
	return id, err
}

func (s *Service) DeletePromotion(id uint64, actor Actor) error {
	return s.softDelete("game_activity_events", "promotion", id, "operation:promotion:update", actor)
}

func decimalValue(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}
func promotionLegacyStatus(v string) int {
	switch v {
	case "ACTIVE", "SCHEDULED":
		return 1
	case "ENDED", "CANCELLED":
		return 3
	default:
		return 2
	}
}
func replacePromotionRules(tx *gorm.DB, id uint64, in PromotionInput, scope content.Scope) error {
	now := time.Now()
	if err := content.ApplySite(tx.Table("game_promotion_participation_rules"), "", scope).Where("promotion_id = ?", id).Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
		return err
	}
	if err := content.ApplySite(tx.Table("game_promotion_reward_rules"), "", scope).Where("promotion_id = ?", id).Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
		return err
	}
	if err := content.ApplySite(tx.Table("game_promotion_eligible_games"), "", scope).Where("promotion_id = ?", id).Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error; err != nil {
		return err
	}
	for _, r := range in.ParticipationRules {
		var n *decimal.Decimal
		if strings.TrimSpace(r.NumericValue) != "" {
			v, e := parseDecimal(r.NumericValue, "numericValue")
			if e != nil {
				return e
			}
			n = &v
		}
		if err := tx.Create(&model.GamePromotionParticipationRule{SiteID: scope.SiteID, PromotionID: id, RuleType: r.RuleType, ComparisonOperator: r.ComparisonOperator, NumericValue: n, TextValue: r.TextValue, Enabled: r.Enabled, SortOrder: r.SortOrder}).Error; err != nil {
			return err
		}
	}
	for _, r := range in.RewardRules {
		amount, e := parseDecimal(r.RewardAmount, "rewardAmount")
		if e != nil {
			return e
		}
		w, e := parseDecimal(r.WageringMultiplier, "wageringMultiplier")
		if e != nil {
			return e
		}
		min, e := parseDecimal(r.MinimumAmount, "minimumAmount")
		if e != nil {
			return e
		}
		max, e := parseDecimal(r.MaximumReward, "maximumReward")
		if e != nil {
			return e
		}
		percentage := decimal.Zero
		if strings.TrimSpace(r.PercentageRate) != "" {
			percentage, e = parseDecimal(r.PercentageRate, "percentageRate")
			if e != nil {
				return e
			}
		}
		if err := tx.Create(&model.GamePromotionRewardRule{SiteID: scope.SiteID, PromotionID: id, RewardType: r.RewardType, CalculationType: r.CalculationType, RewardAmount: &amount, PercentageRate: &percentage, Currency: r.Currency, WageringMultiplier: &w, MinimumAmount: &min, MaximumReward: &max, ExpirationDays: r.ExpirationDays, Enabled: r.Enabled, SortOrder: r.SortOrder}).Error; err != nil {
			return err
		}
	}
	for _, gameID := range in.EligibleGameIDs {
		if err := tx.Table("game_promotion_eligible_games").Create(map[string]any{"site_id": scope.SiteID, "promotion_id": id, "game_id": gameID, "created_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) SetPromotionStatus(id uint64, status string, reason string, actor Actor) error {
	status = strings.ToUpper(strings.TrimSpace(status))
	return s.db.Transaction(func(tx *gorm.DB) error {
		var row promotionRecord
		if err := content.ApplySite(tx.Model(&promotionRecord{}), "", actor.Scope).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
			return notFound(err, "PROMOTION_NOT_FOUND")
		}
		before := row
		if status == "ACTIVE" && !promotionExecutable(row.TriggerType, row.ClaimMode) {
			return newError("PROMOTION_TRIGGER_NOT_CONNECTED", "this promotion trigger is not connected to an executable business event")
		}
		row.StatusCode = status
		row.Status = promotionLegacyStatus(status)
		if err := content.ApplySite(tx.Model(&promotionRecord{}), "", actor.Scope).Where("id=?", id).Updates(map[string]any{"status_code": status, "status": row.Status}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:promotion:status", "status", "promotion", id, before, row, reason)
	})
}

func (s *Service) softDelete(table string, target string, id uint64, permission string, actor Actor) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := content.ApplySite(tx.Table(table), "", actor.Scope).Where("id=? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return newError(strings.ToUpper(target)+"_NOT_FOUND", "record not found")
		}
		if err := content.ApplySite(tx.Table(table), "", actor.Scope).Where("id=?", id).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, "delete", target, id, map[string]any{"deleted": false}, map[string]any{"deleted": true}, "")
	})
}
