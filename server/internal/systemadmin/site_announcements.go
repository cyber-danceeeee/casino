package systemadmin

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
)

var announcementTypes = map[string]bool{
	"SYSTEM": true, "MAINTENANCE": true, "PROMOTION": true,
	"PAYMENT": true, "GAME": true, "GENERAL": true,
}

type AnnouncementConfiguration struct {
	Title            string     `json:"title"`
	Summary          string     `json:"summary"`
	Content          string     `json:"content"`
	AnnouncementType string     `json:"announcementType"`
	Priority         int        `json:"priority"`
	SortOrder        int        `json:"sortOrder"`
	IsPinned         bool       `json:"isPinned"`
	LanguageCode     string     `json:"languageCode"`
	CountryCode      string     `json:"countryCode"`
	CurrencyCode     string     `json:"currencyCode"`
	VisibilityScope  string     `json:"visibilityScope"`
	StartAt          *time.Time `json:"startAt"`
	EndAt            *time.Time `json:"endAt"`
	Enabled          bool       `json:"enabled"`
	Deleted          bool       `json:"deleted,omitempty"`
}

type AnnouncementInput struct {
	AnnouncementConfiguration
	Reason string `json:"reason"`
}

type AnnouncementReorderInput struct {
	IDs    []uint64 `json:"ids"`
	Reason string   `json:"reason"`
}

type SiteAnnouncementRecord struct {
	ID uint64 `gorm:"column:id;primaryKey" json:"id"`

	SiteID               uint64                    `gorm:"column:site_id" json:"siteID"`
	Title                string                    `gorm:"column:title" json:"title"`
	Summary              string                    `gorm:"column:summary" json:"summary"`
	Content              string                    `gorm:"column:content" json:"content"`
	AnnouncementType     string                    `gorm:"column:announcement_type" json:"announcementType"`
	Priority             int                       `gorm:"column:priority" json:"priority"`
	SortOrder            int                       `gorm:"column:sort_order" json:"sortOrder"`
	IsPinned             bool                      `gorm:"column:is_pinned" json:"isPinned"`
	LanguageCode         string                    `gorm:"column:language_code" json:"languageCode"`
	CountryCode          string                    `gorm:"column:country_code" json:"countryCode"`
	CurrencyCode         string                    `gorm:"column:currency_code" json:"currencyCode"`
	VisibilityScope      string                    `gorm:"column:visibility_scope" json:"visibilityScope"`
	StartAt              *time.Time                `gorm:"column:start_at" json:"startAt"`
	EndAt                *time.Time                `gorm:"column:end_at" json:"endAt"`
	Enabled              bool                      `gorm:"column:enabled" json:"enabled"`
	ConfigurationPayload string                    `gorm:"column:configuration_json" json:"-"`
	ConfigVersion        uint64                    `gorm:"column:config_version" json:"configVersion"`
	CreatedBy            uint64                    `gorm:"column:created_by" json:"createdBy"`
	UpdatedBy            uint64                    `gorm:"column:updated_by" json:"updatedBy"`
	CreatedAt            time.Time                 `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt            time.Time                 `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt            *time.Time                `gorm:"column:deleted_at" json:"-"`
	Configuration        AnnouncementConfiguration `gorm:"-" json:"configuration"`
}

func (SiteAnnouncementRecord) TableName() string { return "game_site_announcements" }

type PublicAnnouncement struct {
	ID               uint64 `json:"id"`
	Title            string `json:"title"`
	Summary          string `json:"summary"`
	Content          string `json:"content"`
	AnnouncementType string `json:"announcementType"`
	Priority         int    `json:"priority"`
	IsPinned         bool   `json:"isPinned"`
	ConfigVersion    uint64 `json:"configVersion"`
}

func (s *Service) ListSiteAnnouncements(siteID uint64, q Query, actor Actor) (Result, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return Result{}, err
	}
	rows := []SiteAnnouncementRecord{}
	db := s.db.Model(&SiteAnnouncementRecord{}).Where("site_id = ? AND deleted_at IS NULL", scope.SiteID)
	db = like(db, "title", value(q, "keyword"))
	result, err := paginate(db.Order("is_pinned DESC,priority DESC,sort_order ASC,id ASC"), q, &rows)
	if err != nil {
		return result, err
	}
	for index := range rows {
		hydrateAnnouncement(&rows[index])
	}
	result.List = rows
	return result, nil
}

func (s *Service) GetSiteAnnouncement(siteID uint64, id uint64, actor Actor) (map[string]any, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return nil, err
	}
	row, err := s.announcementByID(s.db, scope, id, false)
	if err != nil {
		return nil, err
	}
	hydrateAnnouncement(&row)
	return map[string]any{"overview": row}, nil
}

func (s *Service) SaveSiteAnnouncement(siteID uint64, id uint64, input AnnouncementInput, actor Actor) (uint64, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return 0, err
	}
	dimensions, err := loadSiteContentDimensions(s.db, scope)
	if err != nil {
		return 0, err
	}
	input.CountryCode = dimensions.CountryCode
	input.CurrencyCode = dimensions.CurrencyCode
	input.LanguageCode = dimensions.DefaultLanguage
	configuration, err := normalizeAnnouncementConfiguration(input.AnnouncementConfiguration)
	if err != nil {
		return 0, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before SiteAnnouncementRecord
		if id > 0 {
			var err error
			before, err = s.announcementByID(tx, scope, id, true)
			if err != nil {
				return err
			}
		} else {
			row := SiteAnnouncementRecord{ConfigurationPayload: "{}", SiteID: scope.SiteID, Title: configuration.Title, CreatedBy: actor.ID, UpdatedBy: actor.ID, CreatedAt: now(), UpdatedAt: now()}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		}
		values, err := announcementValues(configuration, actor.ID)
		if err != nil {
			return err
		}
		if err := tx.Table("game_site_announcements").Where("id=? AND site_id=?", id, scope.SiteID).Updates(values).Error; err != nil {
			return err
		}
		if err := touchSiteConfiguration(tx, scope.SiteID, actor.ID); err != nil {
			return err
		}
		return audit(tx, actor, "operations:announcement:edit", "save", "game_site_announcements", id, before, configuration, input.Reason)
	})
	return id, err
}

func (s *Service) CopySiteAnnouncement(siteID uint64, id uint64, actor Actor) (uint64, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return 0, err
	}
	row, err := s.announcementByID(s.db, scope, id, false)
	if err != nil {
		return 0, err
	}
	configuration := decodeAnnouncement(row.ConfigurationPayload)
	configuration.Title += " Copy"
	configuration.Enabled = false
	return s.SaveSiteAnnouncement(siteID, 0, AnnouncementInput{AnnouncementConfiguration: configuration, Reason: "copy announcement configuration"}, actor)
}

func (s *Service) DeleteSiteAnnouncement(siteID uint64, id uint64, actor Actor) error {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		before, err := s.announcementByID(tx, scope, id, true)
		if err != nil {
			return err
		}
		if err := tx.Table("game_site_announcements").Where("id=? AND site_id=?", id, scope.SiteID).Updates(map[string]any{"deleted_at": now(), "updated_at": now(), "enabled": 0}).Error; err != nil {
			return err
		}
		if err := touchSiteConfiguration(tx, scope.SiteID, actor.ID); err != nil {
			return err
		}
		return audit(tx, actor, "operations:announcement:edit", "delete", "game_site_announcements", id, before, nil, "")
	})
}

func (s *Service) ReorderSiteAnnouncements(siteID uint64, input AnnouncementReorderInput, actor Actor) error {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range input.IDs {
			row, err := s.announcementByID(tx, scope, id, true)
			if err != nil {
				return err
			}
			configuration := decodeAnnouncement(row.ConfigurationPayload)
			configuration.SortOrder = index
			values, err := announcementValues(configuration, actor.ID)
			if err != nil {
				return err
			}
			if err := tx.Table("game_site_announcements").Where("site_id=? AND id=?", scope.SiteID, id).Updates(values).Error; err != nil {
				return err
			}
		}
		if err := touchSiteConfiguration(tx, scope.SiteID, actor.ID); err != nil {
			return err
		}
		return audit(tx, actor, "operations:announcement:edit", "sort", "game_site_announcements", 0, nil, input.IDs, input.Reason)
	})
}

func (s *Service) PublicAnnouncements(scope content.Scope, language string, country string, currency string, loggedIn bool, at time.Time) ([]PublicAnnouncement, error) {
	if scope.SiteID == 0 {
		return []PublicAnnouncement{}, nil
	}
	rows := []PublicAnnouncement{}
	err := s.db.Table("game_site_announcements").Select("id,title,summary,content,announcement_type,priority,is_pinned,updated_at,config_version").
		Where("site_id = ? AND enabled = 1 AND deleted_at IS NULL", scope.SiteID).
		Where("visibility_scope IN ?", contentVisibilityValues(loggedIn)).
		Where("(language_code='' OR language_code='ALL' OR LOWER(language_code) IN ?)", contentLanguageCandidates(language)).
		Where("(country_code='' OR country_code='ALL' OR UPPER(country_code)=?)", strings.ToUpper(strings.TrimSpace(country))).
		Where("(currency_code='' OR currency_code='ALL' OR UPPER(currency_code)=?)", strings.ToUpper(strings.TrimSpace(currency))).
		Where("(start_at IS NULL OR start_at<=?) AND (end_at IS NULL OR end_at>=?)", at, at).
		Order("is_pinned DESC,priority DESC,sort_order ASC,updated_at DESC,id DESC").Find(&rows).Error
	return rows, err
}

func (s *Service) PublicAnnouncement(scope content.Scope, id uint64, language string, country string, currency string, loggedIn bool, at time.Time) (PublicAnnouncement, error) {
	rows, err := s.PublicAnnouncements(scope, language, country, currency, loggedIn, at)
	if err != nil {
		return PublicAnnouncement{}, err
	}
	for _, row := range rows {
		if row.ID == id {
			return row, nil
		}
	}
	return PublicAnnouncement{}, NewError("ANNOUNCEMENT_NOT_FOUND", "announcement not found")
}

func normalizeAnnouncementConfiguration(input AnnouncementConfiguration) (AnnouncementConfiguration, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Content = sanitizeAnnouncementHTML(input.Content)
	input.AnnouncementType = strings.ToUpper(strings.TrimSpace(input.AnnouncementType))
	if input.Title == "" || input.Content == "" {
		return AnnouncementConfiguration{}, NewError("ANNOUNCEMENT_FIELDS_REQUIRED", "title and content are required")
	}
	if !announcementTypes[input.AnnouncementType] {
		return AnnouncementConfiguration{}, NewError("INVALID_ANNOUNCEMENT_TYPE", "invalid announcement type")
	}
	visibility, err := normalizeContentVisibility(input.VisibilityScope)
	if err != nil {
		return AnnouncementConfiguration{}, err
	}
	if err := validateContentTime(input.StartAt, input.EndAt); err != nil {
		return AnnouncementConfiguration{}, err
	}
	input.VisibilityScope = visibility
	input.LanguageCode = normalizeContentDimension(input.LanguageCode)
	input.CountryCode = normalizeContentDimension(input.CountryCode)
	input.CurrencyCode = normalizeContentDimension(input.CurrencyCode)
	return input, nil
}

func (s *Service) announcementByID(db *gorm.DB, scope content.Scope, id uint64, lock bool) (SiteAnnouncementRecord, error) {
	var row SiteAnnouncementRecord
	query := db.Model(&SiteAnnouncementRecord{}).Where("id = ? AND site_id = ? AND deleted_at IS NULL", id, scope.SiteID)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return row, NewError("ANNOUNCEMENT_NOT_FOUND", "announcement not found")
	}
	return row, nil
}

func hydrateAnnouncement(row *SiteAnnouncementRecord) {
	row.Configuration = decodeAnnouncement(row.ConfigurationPayload)
}

func decodeAnnouncement(raw string) AnnouncementConfiguration {
	var value AnnouncementConfiguration
	_ = json.Unmarshal([]byte(raw), &value)
	return value
}
