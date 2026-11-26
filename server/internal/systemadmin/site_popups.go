package systemadmin

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/tenancy"
)

var (
	popupComponentCode = regexp.MustCompile(`^[A-Z][A-Z0-9_]{2,63}$`)
	popupFrequencies   = map[string]bool{
		"EVERY_VISIT": true, "ONCE_PER_SESSION": true, "ONCE_PER_DAY": true, "ONCE_PER_USER": true,
	}
	popupTriggers = map[string]bool{"HOME_ENTER": true, "PAGE_ENTER": true}
)

type PopupComponentRecord struct {
	ID                     uint64     `gorm:"column:id;primaryKey" json:"id"`
	ComponentCode          string     `gorm:"column:component_code" json:"componentCode"`
	NameZh                 string     `gorm:"column:name_zh" json:"nameZh"`
	NameEn                 string     `gorm:"column:name_en" json:"nameEn"`
	DescriptionZh          string     `gorm:"column:description_zh" json:"descriptionZh"`
	DescriptionEn          string     `gorm:"column:description_en" json:"descriptionEn"`
	Category               string     `gorm:"column:category" json:"category"`
	PreviewURL             string     `gorm:"column:preview_url" json:"previewURL"`
	PreviewVideoURL        string     `gorm:"column:preview_video_url" json:"previewVideoURL"`
	H5MinVersion           string     `gorm:"column:h5_min_version" json:"h5MinVersion"`
	DefaultActionType      string     `gorm:"column:default_action_type" json:"defaultActionType"`
	AllowedActionTypesJSON string     `gorm:"column:allowed_action_types" json:"-"`
	Status                 int8       `gorm:"column:status" json:"status"`
	SortOrder              int        `gorm:"column:sort_order" json:"sortOrder"`
	CreatedBy              uint64     `gorm:"column:created_by" json:"createdBy"`
	UpdatedBy              uint64     `gorm:"column:updated_by" json:"updatedBy"`
	CreatedAt              time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt              time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt              *time.Time `gorm:"column:deleted_at" json:"-"`
	AllowedActionTypes     []string   `gorm:"-" json:"allowedActionTypes"`
	Enabled                bool       `gorm:"-" json:"enabled"`
}

func (PopupComponentRecord) TableName() string { return "game_popup_components" }

type PopupComponentInput struct {
	ComponentCode      string   `json:"componentCode"`
	NameZh             string   `json:"nameZh"`
	NameEn             string   `json:"nameEn"`
	DescriptionZh      string   `json:"descriptionZh"`
	DescriptionEn      string   `json:"descriptionEn"`
	Category           string   `json:"category"`
	PreviewURL         string   `json:"previewURL"`
	PreviewVideoURL    string   `json:"previewVideoURL"`
	H5MinVersion       string   `json:"h5MinVersion"`
	DefaultActionType  string   `json:"defaultActionType"`
	AllowedActionTypes []string `json:"allowedActionTypes"`
	SortOrder          int      `json:"sortOrder"`
	Enabled            bool     `json:"enabled"`
	Reason             string   `json:"reason"`
}

type PopupDraft struct {
	PopupComponentID uint64     `json:"popupComponentID"`
	CampaignName     string     `json:"campaignName"`
	ActionType       string     `json:"actionType"`
	ActionTarget     string     `json:"actionTarget"`
	TriggerType      string     `json:"triggerType"`
	TriggerValue     string     `json:"triggerValue"`
	FrequencyType    string     `json:"frequencyType"`
	DisplayLimit     int        `json:"displayLimit"`
	CooldownMinutes  int        `json:"cooldownMinutes"`
	VisibilityScope  string     `json:"visibilityScope"`
	StartAt          *time.Time `json:"startAt"`
	EndAt            *time.Time `json:"endAt"`
	Priority         int        `json:"priority"`
	Enabled          bool       `json:"enabled"`
	Deleted          bool       `json:"deleted,omitempty"`
}

type PopupInput struct {
	PopupDraft
	Reason string `json:"reason"`
}

type PopupReorderInput struct {
	IDs    []uint64 `json:"ids"`
	Reason string   `json:"reason"`
}

type SitePopupRecord struct {
	ID                     uint64                `gorm:"column:id;primaryKey" json:"id"`
	TenantID               uint64                `gorm:"column:tenant_id" json:"tenantID"`
	SiteID                 uint64                `gorm:"column:site_id" json:"siteID"`
	PopupComponentID       uint64                `gorm:"column:popup_component_id" json:"popupComponentID"`
	PopupComponentIDDraft  uint64                `gorm:"column:popup_component_id_draft" json:"popupComponentIDDraft"`
	CampaignName           string                `gorm:"column:campaign_name" json:"campaignName"`
	ActionType             string                `gorm:"column:action_type" json:"actionType"`
	ActionTarget           string                `gorm:"column:action_target" json:"actionTarget"`
	TriggerType            string                `gorm:"column:trigger_type" json:"triggerType"`
	TriggerValue           string                `gorm:"column:trigger_value" json:"triggerValue"`
	FrequencyType          string                `gorm:"column:frequency_type" json:"frequencyType"`
	DisplayLimit           int                   `gorm:"column:display_limit" json:"displayLimit"`
	CooldownMinutes        int                   `gorm:"column:cooldown_minutes" json:"cooldownMinutes"`
	VisibilityScope        string                `gorm:"column:visibility_scope" json:"visibilityScope"`
	StartAt                *time.Time            `gorm:"column:start_at" json:"startAt"`
	EndAt                  *time.Time            `gorm:"column:end_at" json:"endAt"`
	Priority               int                   `gorm:"column:priority" json:"priority"`
	Enabled                bool                  `gorm:"column:enabled" json:"enabled"`
	DraftPayload           string                `gorm:"column:draft_payload" json:"-"`
	PublishedPayload       string                `gorm:"column:published_payload" json:"-"`
	PublishStatus          string                `gorm:"column:publish_status" json:"publishStatus"`
	ConfigVersion          uint64                `gorm:"column:config_version" json:"configVersion"`
	CreatedBy              uint64                `gorm:"column:created_by" json:"createdBy"`
	UpdatedBy              uint64                `gorm:"column:updated_by" json:"updatedBy"`
	DraftUpdatedAt         *time.Time            `gorm:"column:draft_updated_at" json:"draftUpdatedAt"`
	PublishedAt            *time.Time            `gorm:"column:published_at" json:"publishedAt"`
	PublishedBy            uint64                `gorm:"column:published_by" json:"publishedBy"`
	CreatedAt              time.Time             `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt              time.Time             `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt              *time.Time            `gorm:"column:deleted_at" json:"-"`
	DraftConfiguration     PopupDraft            `gorm:"-" json:"draftConfiguration"`
	PublishedConfiguration PopupDraft            `gorm:"-" json:"publishedConfiguration"`
	Component              *PopupComponentRecord `gorm:"-" json:"component"`
}

func (SitePopupRecord) TableName() string { return "game_tenant_popup_campaigns" }

type PopupViewer struct {
	PlayerID     uint64
	RegisteredAt *time.Time
	SessionID    string
	DeviceID     string
}

type PublicPopupFrequency struct {
	Type            string `json:"type"`
	DisplayLimit    int    `json:"displayLimit"`
	CooldownSeconds int    `json:"cooldownSeconds"`
}

type PublicPopupAction struct {
	Type   string `json:"type"`
	Target string `json:"target"`
}

type PublicPopup struct {
	CampaignID    uint64               `json:"campaignId"`
	ComponentCode string               `json:"componentCode"`
	Priority      int                  `json:"priority"`
	TriggerType   string               `json:"triggerType"`
	TriggerValue  string               `json:"triggerValue"`
	Frequency     PublicPopupFrequency `json:"frequency"`
	Action        PublicPopupAction    `json:"action"`
	ConfigVersion uint64               `json:"configVersion"`
}

type PopupEventInput struct {
	EventID string `json:"eventId"`
}

type popupCampaignEvent struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	TenantID     uint64    `gorm:"column:tenant_id;uniqueIndex:uk_popup_campaign_event"`
	SiteID       uint64    `gorm:"column:site_id;uniqueIndex:uk_popup_campaign_event"`
	CampaignID   uint64    `gorm:"column:campaign_id"`
	PlayerID     uint64    `gorm:"column:player_id"`
	SessionID    string    `gorm:"column:session_id"`
	DeviceID     string    `gorm:"column:device_id"`
	EventID      string    `gorm:"column:event_id;uniqueIndex:uk_popup_campaign_event"`
	EventType    string    `gorm:"column:event_type"`
	BusinessDate string    `gorm:"column:business_date"`
	IPAddress    string    `gorm:"column:ip_address"`
	UserAgent    string    `gorm:"column:user_agent"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (popupCampaignEvent) TableName() string { return "game_popup_campaign_events" }

func (s *Service) ListPopupComponents(q Query) (Result, error) {
	rows := []PopupComponentRecord{}
	db := s.db.Model(&PopupComponentRecord{}).Where("deleted_at IS NULL")
	if keyword := strings.TrimSpace(value(q, "keyword")); keyword != "" {
		db = db.Where("component_code LIKE ? OR name_zh LIKE ? OR name_en LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	db = exact(db, "category", strings.ToUpper(value(q, "category")))
	if enabled := strings.TrimSpace(value(q, "enabled")); enabled != "" {
		db = db.Where("status=?", boolString(enabled))
	}
	result, err := paginate(db.Order("sort_order ASC,id ASC"), q, &rows)
	for i := range rows {
		hydratePopupComponent(&rows[i])
	}
	result.List = rows
	return result, err
}

func (s *Service) AvailablePopupComponents() ([]PopupComponentRecord, error) {
	rows := []PopupComponentRecord{}
	err := s.db.Model(&PopupComponentRecord{}).Where("status=1 AND deleted_at IS NULL").Order("sort_order ASC,id ASC").Find(&rows).Error
	for i := range rows {
		hydratePopupComponent(&rows[i])
	}
	return rows, err
}

func (s *Service) SavePopupComponent(id uint64, input PopupComponentInput, actor Actor) (uint64, error) {
	input.ComponentCode = strings.ToUpper(strings.TrimSpace(input.ComponentCode))
	input.NameZh = strings.TrimSpace(input.NameZh)
	input.NameEn = strings.TrimSpace(input.NameEn)
	input.Category = strings.ToUpper(strings.TrimSpace(input.Category))
	input.H5MinVersion = strings.TrimSpace(input.H5MinVersion)
	if !popupComponentCode.MatchString(input.ComponentCode) || input.NameZh == "" || input.NameEn == "" {
		return 0, NewError("POPUP_COMPONENT_FIELDS_REQUIRED", "component code and names are required")
	}
	if input.Category == "" {
		input.Category = "GENERAL"
	}
	if input.H5MinVersion == "" {
		input.H5MinVersion = "1.0.0"
	}
	if err := validatePopupPreviewURL(input.PreviewURL); err != nil {
		return 0, err
	}
	if err := validatePopupPreviewURL(input.PreviewVideoURL); err != nil {
		return 0, err
	}
	allowed, err := normalizePopupActionList(input.AllowedActionTypes)
	if err != nil {
		return 0, err
	}
	input.DefaultActionType = strings.ToUpper(strings.TrimSpace(input.DefaultActionType))
	if !popupContainsString(allowed, input.DefaultActionType) {
		return 0, NewError("INVALID_DEFAULT_POPUP_ACTION", "default action must be allowed")
	}
	allowedJSON, _ := json.Marshal(allowed)
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before PopupComponentRecord
		permission := "system:popup-component:create"
		if id > 0 {
			permission = "system:popup-component:update"
			if err := tx.Model(&PopupComponentRecord{}).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
				return NewError("POPUP_COMPONENT_NOT_FOUND", "popup component not found")
			}
			if before.ComponentCode != input.ComponentCode {
				return NewError("POPUP_COMPONENT_CODE_IMMUTABLE", "component code cannot be changed")
			}
		}
		values := map[string]any{
			"component_code": input.ComponentCode, "name_zh": input.NameZh, "name_en": input.NameEn,
			"description_zh": strings.TrimSpace(input.DescriptionZh), "description_en": strings.TrimSpace(input.DescriptionEn),
			"category": input.Category, "preview_url": strings.TrimSpace(input.PreviewURL), "preview_video_url": strings.TrimSpace(input.PreviewVideoURL),
			"h5_min_version": input.H5MinVersion, "default_action_type": input.DefaultActionType,
			"allowed_action_types": string(allowedJSON), "status": boolValue(input.Enabled), "sort_order": input.SortOrder,
			"updated_by": actor.ID, "updated_at": now(), "deleted_at": nil,
		}
		if id == 0 {
			values["created_by"], values["created_at"] = actor.ID, now()
			result := tx.Table("game_popup_components").Create(values)
			if result.Error != nil {
				return result.Error
			}
			idValue, ok := values["id"].(uint64)
			if ok {
				id = idValue
			} else {
				var created PopupComponentRecord
				if err := tx.Model(&PopupComponentRecord{}).Where("component_code=?", input.ComponentCode).Take(&created).Error; err != nil {
					return err
				}
				id = created.ID
			}
		} else if err := tx.Model(&PopupComponentRecord{}).Where("id=?", id).Updates(values).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, "save", "popup_component", id, before, input, input.Reason)
	})
	return id, err
}

func (s *Service) DeletePopupComponent(id uint64, reason string, actor Actor) error {
	var count int64
	if err := s.db.Table("game_tenant_popup_campaigns").Where("(popup_component_id=? OR popup_component_id_draft=?) AND deleted_at IS NULL", id, id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return NewError("POPUP_COMPONENT_IN_USE", "popup component is used by tenant campaigns")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var before PopupComponentRecord
		if err := tx.Model(&PopupComponentRecord{}).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
			return NewError("POPUP_COMPONENT_NOT_FOUND", "popup component not found")
		}
		if err := tx.Model(&PopupComponentRecord{}).Where("id=?", id).Updates(map[string]any{"deleted_at": now(), "status": 0, "updated_by": actor.ID, "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:popup-component:delete", "delete", "popup_component", id, before, nil, reason)
	})
}

func (s *Service) ListSitePopups(siteID uint64, q Query, actor Actor) (Result, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return Result{}, err
	}
	rows := []SitePopupRecord{}
	db := s.db.Model(&SitePopupRecord{}).Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", scope.TenantID, scope.SiteID)
	db = like(db, "campaign_name", value(q, "keyword"))
	db = exact(db, "publish_status", strings.ToUpper(value(q, "publishStatus")))
	if err := db.Order("priority DESC,id ASC").Find(&rows).Error; err != nil {
		return Result{}, err
	}
	if err := s.hydratePopupCampaigns(rows); err != nil {
		return Result{}, err
	}
	list, total, pageNumber, pageSize := pagePopupCampaigns(rows, q.Page, q.PageSize)
	return Result{List: list, Total: total, Page: pageNumber, PageSize: pageSize}, nil
}

func (s *Service) GetSitePopup(siteID, id uint64, actor Actor) (map[string]any, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return nil, err
	}
	row, err := s.popupByID(s.db, scope, id, false)
	if err != nil {
		return nil, err
	}
	if err := s.hydratePopupCampaigns([]SitePopupRecord{row}); err != nil {
		return nil, err
	}
	hydratePopup(&row)
	component, _ := s.popupComponent(row.DraftConfiguration.PopupComponentID, false)
	row.Component = &component
	return map[string]any{"overview": row}, nil
}

func (s *Service) SaveSitePopup(siteID, id uint64, input PopupInput, actor Actor) (uint64, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return 0, err
	}
	draft, err := s.normalizePopupDraft(input.PopupDraft)
	if err != nil {
		return 0, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var before SitePopupRecord
		if id > 0 {
			before, err = s.popupByID(tx, scope, id, true)
			if err != nil {
				return err
			}
		}
		nowValue := now()
		if id == 0 {
			row := SitePopupRecord{TenantID: scope.TenantID, SiteID: scope.SiteID, PopupComponentIDDraft: draft.PopupComponentID,
				CampaignName: draft.CampaignName, DraftPayload: marshal(draft), PublishedPayload: "{}", PublishStatus: siteContentDraft,
				CreatedBy: actor.ID, UpdatedBy: actor.ID, DraftUpdatedAt: &nowValue, CreatedAt: nowValue, UpdatedAt: nowValue}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := tx.Model(&SitePopupRecord{}).Where("id=? AND tenant_id=? AND site_id=?", id, scope.TenantID, scope.SiteID).
			Updates(map[string]any{"popup_component_id_draft": draft.PopupComponentID, "campaign_name": draft.CampaignName,
				"draft_payload": marshal(draft), "publish_status": siteContentDraft, "draft_updated_at": nowValue,
				"updated_by": actor.ID, "updated_at": nowValue}).Error; err != nil {
			return err
		}
		if err := markSiteContentDraft(tx, scope, actor.ID); err != nil {
			return err
		}
		return audit(tx, actor, "operations:popup:edit", "save_draft", "popup_campaign", id, before, draft, input.Reason)
	})
	return id, err
}

func (s *Service) CopySitePopup(siteID, id uint64, actor Actor) (uint64, error) {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return 0, err
	}
	row, err := s.popupByID(s.db, scope, id, false)
	if err != nil {
		return 0, err
	}
	draft := currentPopupDraft(row)
	draft.CampaignName += " Copy"
	draft.Enabled = false
	return s.SaveSitePopup(siteID, 0, PopupInput{PopupDraft: draft, Reason: "copy popup campaign draft"}, actor)
}

func (s *Service) DeleteSitePopup(siteID, id uint64, actor Actor) error {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return err
	}
	row, err := s.popupByID(s.db, scope, id, false)
	if err != nil {
		return err
	}
	draft := currentPopupDraft(row)
	draft.Deleted = true
	return s.db.Transaction(func(tx *gorm.DB) error {
		nowValue := now()
		if err := tx.Model(&SitePopupRecord{}).Where("id=? AND tenant_id=? AND site_id=?", id, scope.TenantID, scope.SiteID).
			Updates(map[string]any{"draft_payload": marshal(draft), "publish_status": siteContentDraft, "draft_updated_at": nowValue, "updated_by": actor.ID, "updated_at": nowValue}).Error; err != nil {
			return err
		}
		if err := markSiteContentDraft(tx, scope, actor.ID); err != nil {
			return err
		}
		return audit(tx, actor, "operations:popup:delete", "delete_draft", "popup_campaign", id, row, draft, "")
	})
}

func (s *Service) ReorderSitePopups(siteID uint64, input PopupReorderInput, actor Actor) error {
	scope, err := s.contentSiteScope(s.db, siteID, actor, false)
	if err != nil {
		return err
	}
	if len(input.IDs) == 0 {
		return NewError("INVALID_REQUEST", "popup campaign IDs are required")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range input.IDs {
			row, err := s.popupByID(tx, scope, id, true)
			if err != nil {
				return err
			}
			draft := currentPopupDraft(row)
			draft.Priority = len(input.IDs) - index
			if err := tx.Model(&SitePopupRecord{}).Where("id=? AND tenant_id=? AND site_id=?", id, scope.TenantID, scope.SiteID).
				Updates(map[string]any{"draft_payload": marshal(draft), "publish_status": siteContentDraft, "draft_updated_at": now(), "updated_by": actor.ID, "updated_at": now()}).Error; err != nil {
				return err
			}
		}
		if err := markSiteContentDraft(tx, scope, actor.ID); err != nil {
			return err
		}
		return audit(tx, actor, "operations:popup:edit", "reorder_draft", "popup_campaign", 0, nil, input.IDs, input.Reason)
	})
}

func (s *Service) PublicPopups(scope tenancy.Scope, _ string, _ string, _ string, loggedIn bool, viewer PopupViewer, at time.Time) ([]PublicPopup, error) {
	if !scope.SiteScoped() {
		return []PublicPopup{}, nil
	}
	type row struct {
		ID              uint64 `gorm:"column:id"`
		ComponentCode   string `gorm:"column:component_code"`
		ActionType      string `gorm:"column:action_type"`
		ActionTarget    string `gorm:"column:action_target"`
		TriggerType     string `gorm:"column:trigger_type"`
		TriggerValue    string `gorm:"column:trigger_value"`
		FrequencyType   string `gorm:"column:frequency_type"`
		DisplayLimit    int    `gorm:"column:display_limit"`
		CooldownMinutes int    `gorm:"column:cooldown_minutes"`
		Priority        int    `gorm:"column:priority"`
		ConfigVersion   uint64 `gorm:"column:config_version"`
	}
	rows := []row{}
	query := s.db.Table("game_tenant_popup_campaigns p").
		Select("p.id,c.component_code,p.action_type,p.action_target,p.trigger_type,p.trigger_value,p.frequency_type,p.display_limit,p.cooldown_minutes,p.priority,p.config_version").
		Joins("JOIN game_popup_components c ON c.id=p.popup_component_id AND c.status=1 AND c.deleted_at IS NULL").
		Where("p.tenant_id=? AND p.site_id=? AND p.config_version>0 AND p.enabled=1 AND p.deleted_at IS NULL", scope.TenantID, scope.SiteID).
		Where("p.visibility_scope IN ?", contentVisibilityValues(loggedIn)).
		Where("(p.start_at IS NULL OR p.start_at<=?) AND (p.end_at IS NULL OR p.end_at>=?)", at, at)
	if loggedIn && viewer.RegisteredAt != nil && !viewer.RegisteredAt.IsZero() {
		query = query.Where("p.created_at >= ?", *viewer.RegisteredAt)
	}
	err := query.Order("p.priority DESC,p.id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]PublicPopup, 0, len(rows))
	for _, item := range rows {
		eligible, err := s.popupFrequencyEligible(scope, item.ID, item.FrequencyType, item.DisplayLimit, item.CooldownMinutes, viewer, at)
		if err != nil {
			return nil, err
		}
		if !eligible {
			continue
		}
		out = append(out, PublicPopup{CampaignID: item.ID, ComponentCode: item.ComponentCode, Priority: item.Priority,
			TriggerType: item.TriggerType, TriggerValue: item.TriggerValue,
			Frequency: PublicPopupFrequency{Type: item.FrequencyType, DisplayLimit: item.DisplayLimit, CooldownSeconds: item.CooldownMinutes * 60},
			Action:    PublicPopupAction{Type: item.ActionType, Target: item.ActionTarget}, ConfigVersion: item.ConfigVersion})
	}
	return out, nil
}

func (s *Service) TrackPopupEvent(scope tenancy.Scope, campaignID uint64, eventType string, input PopupEventInput, viewer PopupViewer, ip, userAgent string, at time.Time) (bool, error) {
	eventType = strings.ToUpper(strings.TrimSpace(eventType))
	if eventType != "IMPRESSION" && eventType != "CLICK" && eventType != "DISMISS" {
		return false, NewError("INVALID_POPUP_EVENT", "invalid popup event")
	}
	var count int64
	err := s.db.Table("game_tenant_popup_campaigns p").Joins("JOIN game_popup_components c ON c.id=p.popup_component_id AND c.status=1 AND c.deleted_at IS NULL").
		Where("p.id=? AND p.tenant_id=? AND p.site_id=? AND p.config_version>0 AND p.enabled=1 AND p.deleted_at IS NULL", campaignID, scope.TenantID, scope.SiteID).Count(&count).Error
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, NewError("POPUP_CAMPAIGN_NOT_FOUND", "popup campaign not found")
	}
	businessDate := s.popupBusinessDate(scope.SiteID, at)
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" {
		bucket := businessDate
		if eventType == "IMPRESSION" && viewer.SessionID != "" {
			bucket = viewer.SessionID
		}
		eventID = contentEventHash("popup-campaign", stringValue(campaignID), eventType, popupIdentity(viewer), bucket)
	}
	if len(eventID) > 128 {
		eventID = contentEventHash(eventID)
	}
	row := popupCampaignEvent{TenantID: scope.TenantID, SiteID: scope.SiteID, CampaignID: campaignID, PlayerID: viewer.PlayerID,
		SessionID: truncateContent(viewer.SessionID, 128), DeviceID: truncateContent(viewer.DeviceID, 128), EventID: eventID,
		EventType: eventType, BusinessDate: businessDate, IPAddress: truncateContent(ip, 64), UserAgent: truncateContent(userAgent, 500), CreatedAt: at}
	result := s.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tenant_id"}, {Name: "site_id"}, {Name: "event_id"}}, DoNothing: true}).Create(&row)
	return result.RowsAffected > 0, result.Error
}

func (s *Service) normalizePopupDraft(input PopupDraft) (PopupDraft, error) {
	input.CampaignName = strings.TrimSpace(input.CampaignName)
	input.TriggerType = strings.ToUpper(strings.TrimSpace(input.TriggerType))
	input.TriggerValue = strings.TrimSpace(input.TriggerValue)
	input.FrequencyType = strings.ToUpper(strings.TrimSpace(input.FrequencyType))
	if input.PopupComponentID == 0 || input.CampaignName == "" {
		return PopupDraft{}, NewError("POPUP_CAMPAIGN_FIELDS_REQUIRED", "popup component and campaign name are required")
	}
	component, err := s.popupComponent(input.PopupComponentID, true)
	if err != nil {
		return PopupDraft{}, err
	}
	if !popupTriggers[input.TriggerType] {
		return PopupDraft{}, NewError("INVALID_POPUP_TRIGGER", "invalid popup trigger")
	}
	if input.TriggerType == "PAGE_ENTER" && !isSafeInternalContentPath(input.TriggerValue) {
		return PopupDraft{}, NewError("INVALID_POPUP_TRIGGER_TARGET", "page trigger must be an absolute site path")
	}
	if input.TriggerType == "HOME_ENTER" {
		input.TriggerValue = ""
	}
	if !popupFrequencies[input.FrequencyType] {
		return PopupDraft{}, NewError("INVALID_POPUP_FREQUENCY", "invalid popup display frequency")
	}
	visibility, err := normalizeContentVisibility(input.VisibilityScope)
	if err != nil {
		return PopupDraft{}, err
	}
	if input.FrequencyType == "ONCE_PER_USER" && visibility == "GUEST" {
		return PopupDraft{}, NewError("POPUP_LOGIN_REQUIRED", "once per user campaigns require logged-in visibility")
	}
	if input.DisplayLimit < 0 || input.CooldownMinutes < 0 {
		return PopupDraft{}, NewError("INVALID_POPUP_LIMIT", "popup limits must be non-negative")
	}
	if err := validateContentTime(input.StartAt, input.EndAt); err != nil {
		return PopupDraft{}, err
	}
	action, target, err := validateSiteContentAction(input.ActionType, input.ActionTarget)
	if err != nil {
		return PopupDraft{}, err
	}
	if !popupContainsString(component.AllowedActionTypes, action) {
		return PopupDraft{}, NewError("POPUP_ACTION_NOT_ALLOWED", "action is not allowed by this popup component")
	}
	input.ActionType, input.ActionTarget, input.VisibilityScope = action, target, visibility
	return input, nil
}

func (s *Service) popupComponent(id uint64, active bool) (PopupComponentRecord, error) {
	var row PopupComponentRecord
	db := s.db.Model(&PopupComponentRecord{}).Where("id=? AND deleted_at IS NULL", id)
	if active {
		db = db.Where("status=1")
	}
	if err := db.Take(&row).Error; err != nil {
		return row, NewError("POPUP_COMPONENT_NOT_FOUND", "popup component not found or disabled")
	}
	hydratePopupComponent(&row)
	return row, nil
}

func (s *Service) popupByID(db *gorm.DB, scope tenancy.Scope, id uint64, lock bool) (SitePopupRecord, error) {
	var row SitePopupRecord
	query := db.Model(&SitePopupRecord{}).Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", id, scope.TenantID, scope.SiteID)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Take(&row).Error; err != nil {
		return row, NewError("POPUP_CAMPAIGN_NOT_FOUND", "popup campaign not found")
	}
	return row, nil
}

func (s *Service) hydratePopupCampaigns(rows []SitePopupRecord) error {
	ids := make([]uint64, 0, len(rows)*2)
	for i := range rows {
		hydratePopup(&rows[i])
		ids = append(ids, rows[i].DraftConfiguration.PopupComponentID, rows[i].PublishedConfiguration.PopupComponentID)
	}
	components := []PopupComponentRecord{}
	if len(ids) > 0 {
		if err := s.db.Model(&PopupComponentRecord{}).Where("id IN ? AND deleted_at IS NULL", ids).Find(&components).Error; err != nil {
			return err
		}
	}
	byID := map[uint64]*PopupComponentRecord{}
	for i := range components {
		hydratePopupComponent(&components[i])
		byID[components[i].ID] = &components[i]
	}
	for i := range rows {
		rows[i].Component = byID[rows[i].DraftConfiguration.PopupComponentID]
	}
	return nil
}

func hydratePopup(row *SitePopupRecord) {
	row.DraftConfiguration = decodePopup(row.DraftPayload)
	row.PublishedConfiguration = decodePopup(row.PublishedPayload)
}

func decodePopup(raw string) PopupDraft {
	var value PopupDraft
	_ = json.Unmarshal([]byte(raw), &value)
	return value
}

func currentPopupDraft(row SitePopupRecord) PopupDraft {
	draft := decodePopup(row.DraftPayload)
	if strings.TrimSpace(row.DraftPayload) == "" || draft.PopupComponentID == 0 {
		draft = decodePopup(row.PublishedPayload)
	}
	return draft
}

func hydratePopupComponent(row *PopupComponentRecord) {
	row.Enabled = row.Status == 1
	_ = json.Unmarshal([]byte(row.AllowedActionTypesJSON), &row.AllowedActionTypes)
}

func normalizePopupActionList(values []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if !siteContentActions[value] {
			return nil, NewError("INVALID_POPUP_ACTION", "unsupported popup action")
		}
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil, NewError("POPUP_ACTION_REQUIRED", "at least one popup action is required")
	}
	sort.Strings(result)
	return result, nil
}

func validatePopupPreviewURL(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return NewError("INVALID_POPUP_PREVIEW_URL", "popup preview URL must use https")
	}
	return nil
}

func popupContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func boolString(value string) int {
	if strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "enabled") {
		return 1
	}
	return 0
}

func pagePopupCampaigns(rows []SitePopupRecord, pageNumber, pageSize int) ([]SitePopupRecord, int64, int, int) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	total := int64(len(rows))
	start := (pageNumber - 1) * pageSize
	if start >= len(rows) {
		return []SitePopupRecord{}, total, pageNumber, pageSize
	}
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], total, pageNumber, pageSize
}

func (s *Service) popupFrequencyEligible(scope tenancy.Scope, campaignID uint64, frequency string, displayLimit, cooldown int, viewer PopupViewer, at time.Time) (bool, error) {
	base := s.popupViewerEvents(scope, campaignID, viewer).Where("event_type='IMPRESSION'")
	if displayLimit > 0 {
		var total int64
		if err := base.Count(&total).Error; err != nil {
			return false, err
		}
		if total >= int64(displayLimit) {
			return false, nil
		}
	}
	var count int64
	switch frequency {
	case "ONCE_PER_SESSION":
		if err := base.Where("session_id=?", viewer.SessionID).Count(&count).Error; err != nil {
			return false, err
		}
	case "ONCE_PER_DAY":
		if err := base.Where("business_date=?", s.popupBusinessDate(scope.SiteID, at)).Count(&count).Error; err != nil {
			return false, err
		}
	case "ONCE_PER_USER":
		if err := base.Count(&count).Error; err != nil {
			return false, err
		}
	}
	if count > 0 {
		return false, nil
	}
	if cooldown > 0 {
		var latest struct {
			CreatedAt time.Time `gorm:"column:created_at"`
		}
		err := base.Select("created_at").Order("created_at DESC").Take(&latest).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}
		if !latest.CreatedAt.IsZero() && latest.CreatedAt.Add(time.Duration(cooldown)*time.Minute).After(at) {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) popupViewerEvents(scope tenancy.Scope, campaignID uint64, viewer PopupViewer) *gorm.DB {
	db := s.db.Table("game_popup_campaign_events").Where("tenant_id=? AND site_id=? AND campaign_id=?", scope.TenantID, scope.SiteID, campaignID)
	if viewer.PlayerID > 0 {
		return db.Where("player_id=?", viewer.PlayerID)
	}
	return db.Where("player_id=0 AND device_id=?", viewer.DeviceID)
}

func (s *Service) popupBusinessDate(siteID uint64, at time.Time) string {
	var timezone string
	_ = s.db.Table("game_saas_sites").Select("timezone").Where("id=?", siteID).Scan(&timezone).Error
	location, err := time.LoadLocation(timezone)
	if err == nil {
		at = at.In(location)
	}
	return at.Format("2006-01-02")
}

func popupIdentity(viewer PopupViewer) string {
	if viewer.PlayerID > 0 {
		return "player:" + stringValue(viewer.PlayerID)
	}
	return "device:" + viewer.DeviceID
}

func stringValue(value uint64) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = digits[value%10]
		value /= 10
	}
	return string(buf[index:])
}

func truncateContent(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) > limit {
		return value[:limit]
	}
	return value
}
