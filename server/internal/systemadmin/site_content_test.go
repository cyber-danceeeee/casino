package systemadmin

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"casino/internal/content"
)

type contentTestAsset struct {
	ID uint64 `gorm:"column:id;primaryKey"`

	SiteID     uint64     `gorm:"column:site_id"`
	AssetScope string     `gorm:"column:asset_scope"`
	AssetType  string     `gorm:"column:asset_type"`
	Status     string     `gorm:"column:status"`
	IsPublic   bool       `gorm:"column:is_public"`
	ObjectKey  string     `gorm:"column:object_key"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

func (contentTestAsset) TableName() string { return "game_asset_library" }

func TestAnnouncementDraftPublishIsolationAndXSS(t *testing.T) {
	service, db := newSiteContentTestService(t)
	actorA := Actor{ID: 7, Username: "h5-a", SiteID: 101}
	input := AnnouncementInput{AnnouncementDraft: AnnouncementDraft{
		Title: "Security", Summary: "Safe content", Content: `<p onclick="alert(1)">Hello <strong>world</strong><script>alert(1)</script><a href="javascript:alert(1)">bad</a></p>`,
		AnnouncementType: "SYSTEM", Priority: 5, IsPinned: true, LanguageCode: "EN", CountryCode: "MX",
		CurrencyCode: "MXN", VisibilityScope: "GUEST", Enabled: true,
	}, Reason: "create security announcement"}
	id, err := service.SaveSiteAnnouncement(101, 0, input, actorA)
	require.NoError(t, err)

	rows, err := service.PublicAnnouncements(content.Scope{SiteID: 101}, "es-MX", "MX", "MXN", false, time.Now())
	require.NoError(t, err)
	require.Empty(t, rows, "draft announcement must not be public")

	count, err := publishAnnouncementDrafts(db, 101, actorA.ID, 2)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = publishAnnouncementDrafts(db, 101, actorA.ID, 2)
	require.NoError(t, err)
	require.Zero(t, count, "repeat publish must be idempotent")
	rows, err = service.PublicAnnouncements(content.Scope{SiteID: 101}, "es-MX", "MX", "MXN", false, time.Now())
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.NotContains(t, strings.ToLower(rows[0].Content), "script")
	require.NotContains(t, strings.ToLower(rows[0].Content), "onclick")
	require.NotContains(t, strings.ToLower(rows[0].Content), "javascript:")
	require.Contains(t, rows[0].Content, "<strong>world</strong>")
	currencyMismatch, err := service.PublicAnnouncements(content.Scope{SiteID: 101}, "en", "MX", "USD", false, time.Now())
	require.NoError(t, err)
	require.Empty(t, currencyMismatch)

	_, err = service.GetSiteAnnouncement(101, id, Actor{ID: 8, SiteID: 202})
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "SITE_NOT_FOUND", domain.Code)

	require.NoError(t, service.DeleteSiteAnnouncement(101, id, actorA))
	rows, err = service.PublicAnnouncements(content.Scope{SiteID: 101}, "es", "MX", "MXN", false, time.Now())
	require.NoError(t, err)
	require.Len(t, rows, 1, "published announcement remains live until delete draft is published")
	count, err = publishAnnouncementDrafts(db, 101, actorA.ID, 3)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	rows, err = service.PublicAnnouncements(content.Scope{SiteID: 101}, "es", "MX", "MXN", false, time.Now())
	require.NoError(t, err)
	require.Empty(t, rows)
}

func TestPopupCampaignPublishFrequencyTrackingAndIsolation(t *testing.T) {
	service, db := newSiteContentTestService(t)
	component := PopupComponentRecord{ID: 1, ComponentCode: "WELCOME_BONUS_POPUP", NameZh: "欢迎奖励弹窗", NameEn: "Welcome Bonus Popup",
		DefaultActionType: "OPEN_PROMOTION", AllowedActionTypesJSON: `["OPEN_PROMOTION","CLOSE_ONLY"]`, Status: 1}
	require.NoError(t, db.Create(&component).Error)
	actorA := Actor{ID: 7, Username: "h5-a", SiteID: 101}
	input := PopupInput{PopupDraft: PopupDraft{PopupComponentID: 1, CampaignName: "Daily Offer",
		ActionType: "OPEN_PROMOTION", ActionTarget: "daily-offer", TriggerType: "HOME_ENTER", FrequencyType: "ONCE_PER_DAY",
		VisibilityScope: "ALL", Enabled: true}, Reason: "create daily popup campaign"}
	id, err := service.SaveSitePopup(101, 0, input, actorA)
	require.NoError(t, err)
	viewer := PopupViewer{SessionID: "session-a", DeviceID: "device-a"}
	rows, err := service.PublicPopups(content.Scope{SiteID: 101}, "es-MX", "MX", "MXN", false, viewer, time.Now())
	require.NoError(t, err)
	require.Empty(t, rows, "draft popup must not be public")

	count, err := publishPopupDrafts(db, 101, actorA.ID, 2)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = publishPopupDrafts(db, 101, actorA.ID, 2)
	require.NoError(t, err)
	require.Zero(t, count, "repeat popup publish must be idempotent")
	rows, err = service.PublicPopups(content.Scope{SiteID: 101}, "es-MX", "MX", "MXN", false, viewer, time.Now())
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "WELCOME_BONUS_POPUP", rows[0].ComponentCode)
	require.Equal(t, "OPEN_PROMOTION", rows[0].Action.Type)
	registeredAfterPopup := time.Now().Add(time.Minute)
	rows, err = service.PublicPopups(content.Scope{SiteID: 101}, "es-MX", "MX", "MXN", true, PopupViewer{PlayerID: 77, SessionID: "session-b", DeviceID: "device-b", RegisteredAt: &registeredAfterPopup}, time.Now())
	require.NoError(t, err)
	require.Empty(t, rows, "logged-in users must not receive popup campaigns created before registration")
	rows, err = service.PublicPopups(content.Scope{SiteID: 202}, "es-MX", "MX", "MXN", false, PopupViewer{SessionID: "session-c", DeviceID: "device-c"}, time.Now())
	require.NoError(t, err)
	require.Empty(t, rows, "h5 B must not receive h5 A popup campaigns")

	recorded, err := service.TrackPopupEvent(content.Scope{SiteID: 101}, id, "IMPRESSION", PopupEventInput{EventID: "event-1"}, viewer, "127.0.0.1", "test", time.Now())
	require.NoError(t, err)
	require.True(t, recorded)
	recorded, err = service.TrackPopupEvent(content.Scope{SiteID: 101}, id, "IMPRESSION", PopupEventInput{EventID: "event-1"}, viewer, "127.0.0.1", "test", time.Now())
	require.NoError(t, err)
	require.False(t, recorded, "duplicate popup event must be idempotent")
	rows, err = service.PublicPopups(content.Scope{SiteID: 101}, "es", "MX", "MXN", false, viewer, time.Now())
	require.NoError(t, err)
	require.Empty(t, rows, "once-per-day popup must be suppressed after an impression")

	_, err = service.GetSitePopup(202, id, Actor{ID: 8, SiteID: 202})
	require.Error(t, err, "h5 B must not read h5 A popup campaign")

	require.NoError(t, db.Model(&PopupComponentRecord{}).Where("id=?", 1).Update("status", 0).Error)
	_, err = service.SaveSitePopup(101, id, input, actorA)
	require.Error(t, err, "disabled built-in component must not be assigned")
}

func TestPopupComponentRegistryValidationAndUsageProtection(t *testing.T) {
	service, db := newSiteContentTestService(t)
	actor := Actor{ID: 1, Username: "platform-admin", Platform: true}
	input := PopupComponentInput{ComponentCode: "CUSTOM_EVENT_POPUP", NameZh: "活动弹窗", NameEn: "Event Popup", Category: "PROMOTION",
		PreviewURL: "https://casino.example.com/popup-preview/CUSTOM_EVENT_POPUP", H5MinVersion: "1.2.0",
		DefaultActionType: "OPEN_PROMOTION", AllowedActionTypes: []string{"OPEN_PROMOTION", "CLOSE_ONLY"}, Enabled: true, Reason: "register component"}
	id, err := service.SavePopupComponent(0, input, actor)
	require.NoError(t, err)
	require.NotZero(t, id)

	available, err := service.AvailablePopupComponents()
	require.NoError(t, err)
	require.Len(t, available, 1)
	require.Equal(t, []string{"CLOSE_ONLY", "OPEN_PROMOTION"}, available[0].AllowedActionTypes)

	invalid := input
	invalid.ComponentCode = "UNSAFE_PREVIEW_POPUP"
	invalid.PreviewURL = "javascript:alert(1)"
	_, err = service.SavePopupComponent(0, invalid, actor)
	require.Error(t, err)

	require.NoError(t, db.Create(&SitePopupRecord{SiteID: 101, PopupComponentID: id, PopupComponentIDDraft: id,
		CampaignName: "in use", DraftPayload: `{}`, PublishedPayload: `{}`, PublishStatus: "PUBLISHED"}).Error)
	err = service.DeletePopupComponent(id, "remove", actor)
	var domain *Error
	require.True(t, errors.As(err, &domain))
	require.Equal(t, "POPUP_COMPONENT_IN_USE", domain.Code)
}

func TestPopupFrequencyRules(t *testing.T) {
	service, db := newSiteContentTestService(t)
	scope := content.Scope{SiteID: 101}
	nowValue := time.Now().UTC().Truncate(time.Second)
	viewer := PopupViewer{SessionID: "session-a", DeviceID: "device-a"}
	businessDate := service.popupBusinessDate(scope.SiteID, nowValue)
	for _, popupID := range []uint64{10, 11, 12, 13} {
		require.NoError(t, db.Create(&popupCampaignEvent{SiteID: 101, CampaignID: popupID, SessionID: viewer.SessionID,
			DeviceID: viewer.DeviceID, EventID: "frequency-" + stringValue(popupID), EventType: "IMPRESSION", BusinessDate: businessDate, CreatedAt: nowValue}).Error)
	}
	eligible, err := service.popupFrequencyEligible(scope, 10, "EVERY_VISIT", 0, 0, viewer, nowValue)
	require.NoError(t, err)
	require.True(t, eligible)
	eligible, err = service.popupFrequencyEligible(scope, 11, "ONCE_PER_SESSION", 0, 0, viewer, nowValue)
	require.NoError(t, err)
	require.False(t, eligible)
	eligible, err = service.popupFrequencyEligible(scope, 11, "ONCE_PER_SESSION", 0, 0, PopupViewer{SessionID: "session-b", DeviceID: viewer.DeviceID}, nowValue)
	require.NoError(t, err)
	require.True(t, eligible)
	eligible, err = service.popupFrequencyEligible(scope, 12, "ONCE_PER_DAY", 0, 0, viewer, nowValue)
	require.NoError(t, err)
	require.False(t, eligible)
	eligible, err = service.popupFrequencyEligible(scope, 12, "ONCE_PER_DAY", 0, 0, viewer, nowValue.Add(24*time.Hour))
	require.NoError(t, err)
	require.True(t, eligible)
	eligible, err = service.popupFrequencyEligible(scope, 13, "EVERY_VISIT", 1, 0, viewer, nowValue)
	require.NoError(t, err)
	require.False(t, eligible, "display limit must suppress further impressions")
	eligible, err = service.popupFrequencyEligible(scope, 10, "EVERY_VISIT", 0, 60, viewer, nowValue.Add(30*time.Minute))
	require.NoError(t, err)
	require.False(t, eligible, "cooldown must suppress recent impressions")

	player := PopupViewer{PlayerID: 42, SessionID: "player-session", DeviceID: "player-device"}
	require.NoError(t, db.Create(&popupCampaignEvent{SiteID: 101, CampaignID: 14, PlayerID: player.PlayerID,
		SessionID: player.SessionID, DeviceID: player.DeviceID, EventID: "frequency-player", EventType: "IMPRESSION", BusinessDate: businessDate, CreatedAt: nowValue}).Error)
	eligible, err = service.popupFrequencyEligible(scope, 14, "ONCE_PER_USER", 0, 0, player, nowValue.Add(48*time.Hour))
	require.NoError(t, err)
	require.False(t, eligible)
}

func TestPopupActionTargetsRejectScriptAndMalformedRoutes(t *testing.T) {
	service, db := newSiteContentTestService(t)
	require.NoError(t, db.Create(&PopupComponentRecord{ID: 1, ComponentCode: "TEST_POPUP", NameZh: "测试", NameEn: "Test",
		DefaultActionType: "CLOSE_ONLY", AllowedActionTypesJSON: `["OPEN_INTERNAL_PAGE","OPEN_EXTERNAL_URL","OPEN_GAME","OPEN_PROMOTION","CLOSE_ONLY"]`, Status: 1}).Error)
	for _, input := range []PopupDraft{
		{PopupComponentID: 1, CampaignName: "script", TriggerType: "HOME_ENTER", ActionType: "OPEN_INTERNAL_PAGE", ActionTarget: "javascript:alert(1)", FrequencyType: "EVERY_VISIT", VisibilityScope: "ALL"},
		{PopupComponentID: 1, CampaignName: "external", TriggerType: "HOME_ENTER", ActionType: "OPEN_EXTERNAL_URL", ActionTarget: "javascript:alert(1)", FrequencyType: "EVERY_VISIT", VisibilityScope: "ALL"},
		{PopupComponentID: 1, CampaignName: "game", TriggerType: "HOME_ENTER", ActionType: "OPEN_GAME", ActionTarget: "../../admin", FrequencyType: "EVERY_VISIT", VisibilityScope: "ALL"},
		{PopupComponentID: 1, CampaignName: "promotion", TriggerType: "HOME_ENTER", ActionType: "OPEN_PROMOTION", ActionTarget: "data:text/html,bad", FrequencyType: "EVERY_VISIT", VisibilityScope: "ALL"},
	} {
		_, err := service.normalizePopupDraft(input)
		require.Error(t, err)
	}
	valid, err := service.normalizePopupDraft(PopupDraft{PopupComponentID: 1, CampaignName: "valid", TriggerType: "HOME_ENTER",
		ActionType: "OPEN_INTERNAL_PAGE", ActionTarget: "/offers?tab=events", FrequencyType: "EVERY_VISIT", VisibilityScope: "ALL"})
	require.NoError(t, err)
	require.Equal(t, "/offers?tab=events", valid.ActionTarget)
}

func TestAnnouncementScheduleLanguageAndLoginFilters(t *testing.T) {
	service, db := newSiteContentTestService(t)
	nowValue := time.Now().UTC().Truncate(time.Second)
	future := nowValue.Add(time.Hour)
	past := nowValue.Add(-time.Hour)
	for _, row := range []SiteAnnouncementRecord{
		{SiteID: 101, Title: "guest", Content: "guest", AnnouncementType: "GENERAL", LanguageCode: "EN", CountryCode: "MX", VisibilityScope: "GUEST", Enabled: true, ConfigVersion: 1},
		{SiteID: 101, Title: "member", Content: "member", AnnouncementType: "GENERAL", LanguageCode: "EN", CountryCode: "MX", VisibilityScope: "LOGGED_IN", Enabled: true, ConfigVersion: 1},
		{SiteID: 101, Title: "future", Content: "future", AnnouncementType: "GENERAL", LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", StartAt: &future, Enabled: true, ConfigVersion: 1},
		{SiteID: 101, Title: "expired", Content: "expired", AnnouncementType: "GENERAL", LanguageCode: "ALL", CountryCode: "ALL", VisibilityScope: "ALL", EndAt: &past, Enabled: true, ConfigVersion: 1},
	} {
		require.NoError(t, db.Create(&row).Error)
	}
	guest, err := service.PublicAnnouncements(content.Scope{SiteID: 101}, "en-US", "MX", "MXN", false, nowValue)
	require.NoError(t, err)
	require.Len(t, guest, 1)
	require.Equal(t, "guest", guest[0].Title)
	member, err := service.PublicAnnouncements(content.Scope{SiteID: 101}, "en-US", "MX", "MXN", true, nowValue)
	require.NoError(t, err)
	require.Len(t, member, 1)
	require.Equal(t, "member", member[0].Title)
}

func newSiteContentTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	service, db := newSystemAdminScopeTestService(t)
	require.NoError(t, db.AutoMigrate(&SiteAnnouncementRecord{}, &PopupComponentRecord{}, &SitePopupRecord{}, &popupCampaignEvent{}, &contentTestAsset{}))
	if !db.Migrator().HasColumn("game_site_config", "updated_by") {
		require.NoError(t, db.Exec("ALTER TABLE game_site_config ADD COLUMN updated_by INTEGER NOT NULL DEFAULT 0").Error)
	}
	require.NoError(t, db.Exec("CREATE TABLE game_admin_operation_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME, site_id INTEGER NOT NULL DEFAULT 0, operator_id INTEGER NOT NULL DEFAULT 0, operator_username TEXT NOT NULL DEFAULT '', permission TEXT NOT NULL DEFAULT '', module TEXT NOT NULL DEFAULT '', action TEXT NOT NULL DEFAULT '', result TEXT NOT NULL DEFAULT '', target_type TEXT NOT NULL DEFAULT '', target_id INTEGER NOT NULL DEFAULT 0, target_user_id INTEGER NOT NULL DEFAULT 0, target_device_id INTEGER NOT NULL DEFAULT 0, before_state TEXT, after_state TEXT, amount TEXT, currency TEXT NOT NULL DEFAULT '', reason TEXT NOT NULL DEFAULT '', ip_address TEXT NOT NULL DEFAULT '', user_agent TEXT NOT NULL DEFAULT '', correlation_id TEXT NOT NULL DEFAULT '')").Error)
	return service, db
}
