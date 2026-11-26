package systemadmin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"casino/internal/eventbus"
	"casino/internal/tenancy"
	"gorm.io/gorm"
)

type ConfigReleaseRow struct {
	ID               uint64     `json:"id"`
	TenantID         uint64     `json:"tenantID"`
	SiteID           uint64     `json:"siteID"`
	Environment      string     `json:"environment"`
	Version          uint64     `json:"version"`
	Status           string     `json:"status"`
	Modules          string     `json:"modules"`
	RollbackSourceID uint64     `json:"rollbackSourceID"`
	FailureReason    string     `json:"failureReason"`
	CreatedBy        uint64     `json:"createdBy"`
	PublishedBy      uint64     `json:"publishedBy"`
	CreatedAt        time.Time  `json:"createdAt"`
	PublishedAt      *time.Time `json:"publishedAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type ReleaseRollbackInput struct {
	Reason         string `json:"reason"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type failedReleaseAttempt struct {
	TenantID  uint64
	SiteID    uint64
	Version   uint64
	Key       string
	Modules   string
	CreatedBy uint64
}

type releaseSnapshot struct {
	Site          map[string]any   `json:"site"`
	LobbySections []map[string]any `json:"lobbySections"`
	Games         []map[string]any `json:"games"`
	Banners       []map[string]any `json:"banners"`
	Announcements []map[string]any `json:"announcements"`
	Popups        []map[string]any `json:"popups"`
	Navigation    []map[string]any `json:"navigation"`
}

func releaseIdempotencyKey(tenantID, siteID uint64, provided string) string {
	provided = strings.TrimSpace(provided)
	if provided == "" {
		provided = fmt.Sprintf("auto:%d", time.Now().UnixNano())
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("release|%d|%d|%s", tenantID, siteID, provided)))
	return "release:" + hex.EncodeToString(sum[:])
}

func existingReleaseByKey(tx *gorm.DB, tenantID, siteID uint64, key string) (ConfigReleaseRow, bool, error) {
	var row ConfigReleaseRow
	err := tx.Table("game_config_releases").Where("tenant_id=? AND site_id=? AND idempotency_key=? AND deleted_at IS NULL", tenantID, siteID, key).Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return row, false, nil
	}
	if err != nil {
		return row, false, err
	}
	if row.Status == "PUBLISHING" {
		return row, false, NewError("RELEASE_PUBLISHING", "release is already publishing")
	}
	if row.Status == "FAILED" {
		return row, false, NewError("RELEASE_FAILED", "the idempotent release previously failed")
	}
	return row, true, nil
}

func releaseResult(row ConfigReleaseRow, idempotent bool) map[string]any {
	return map[string]any{"releaseID": row.ID, "tenantId": row.TenantID, "siteId": row.SiteID, "configVersion": row.Version, "publishedAt": row.PublishedAt, "status": row.Status, "modules": splitCSV(row.Modules), "idempotent": idempotent}
}

func pendingReleaseModules(tx *gorm.DB, row SiteRow) ([]string, bool, error) {
	modules := []string{}
	if strings.EqualFold(row.ConfigStatus, "DRAFT") {
		modules = append(modules, "BRAND")
	}
	checks := []struct{ table, column, module string }{{"game_lobby_sections", "publish_status", "LOBBY"}, {"game_games", "config_status", "GAMES"}, {"game_site_banners", "publish_status", "BANNERS"}, {"game_site_announcements", "publish_status", "ANNOUNCEMENTS"}, {"game_tenant_popup_campaigns", "publish_status", "POPUPS"}, {"game_h5_navigation_items", "publish_status", "NAVIGATION"}}
	for _, check := range checks {
		var count int64
		if err := tx.Table(check.table).Where("tenant_id=? AND site_id=? AND "+check.column+"='DRAFT' AND deleted_at IS NULL", row.TenantID, row.ID).Count(&count).Error; err != nil {
			return nil, false, err
		}
		if count > 0 {
			modules = append(modules, check.module)
		}
	}
	return modules, len(modules) > 0, nil
}

func nextReleaseVersion(tx *gorm.DB, tenantID, siteID, current uint64) (uint64, error) {
	var maximum uint64
	if err := tx.Table("game_config_releases").Select("COALESCE(MAX(version),0)").Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", tenantID, siteID).Scan(&maximum).Error; err != nil {
		return 0, err
	}
	if current > maximum {
		maximum = current
	}
	if maximum == 0 {
		maximum = 1
	}
	return maximum + 1, nil
}

func createPublishingRelease(tx *gorm.DB, tenantID, siteID, version uint64, modules []string, key string, rollbackSourceID, actorID uint64) (uint64, error) {
	row := map[string]any{"tenant_id": tenantID, "site_id": siteID, "environment": "PRODUCTION", "version": version, "status": "PUBLISHING", "modules": strings.Join(modules, ","), "rollback_source_id": rollbackSourceID, "idempotency_key": key, "created_by": actorID, "created_at": now(), "updated_at": now()}
	if err := tx.Table("game_config_releases").Create(row).Error; err != nil {
		return 0, duplicateError(err)
	}
	var id uint64
	if err := tx.Table("game_config_releases").Select("id").Where("idempotency_key=?", key).Take(&id).Error; err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Service) recordFailedRelease(attempt failedReleaseAttempt, cause error) {
	reason := sanitizeFailure(cause.Error())
	row := map[string]any{"tenant_id": attempt.TenantID, "site_id": attempt.SiteID, "environment": "PRODUCTION", "version": attempt.Version, "status": "FAILED", "modules": attempt.Modules, "idempotency_key": attempt.Key, "failure_reason": reason, "created_by": attempt.CreatedBy, "created_at": now(), "updated_at": now()}
	_ = s.db.Table("game_config_releases").Create(row).Error
	_ = s.raiseAlert(alertFingerprint("RELEASE_FAILED", fmt.Sprintf("%d:%d", attempt.TenantID, attempt.SiteID), attempt.TenantID, attempt.SiteID), attempt.TenantID, attempt.SiteID, "WARNING", "RELEASE_FAILED", "release-center", "Configuration release failed", map[string]any{"version": attempt.Version, "error": reason})
}

func captureReleaseSnapshot(tx *gorm.DB, tenantID, siteID uint64) (releaseSnapshot, error) {
	var site SiteRow
	if err := tx.Table("game_saas_sites s").Select(siteSelect()).Joins("LEFT JOIN game_saas_tenants t ON t.id=s.tenant_id AND t.deleted_at IS NULL").Where("s.id=? AND s.tenant_id=? AND s.deleted_at IS NULL", siteID, tenantID).Take(&site).Error; err != nil {
		return releaseSnapshot{}, err
	}
	snapshot := releaseSnapshot{Site: sitePublishedConfig(site)}
	tables := []struct {
		name    string
		columns string
		target  *[]map[string]any
	}{{"game_lobby_sections", "id,section_code,section_type,title_zh,title_en,display_style,source_type,source_ref,max_items,sort_order,enabled,supported_countries,supported_currencies,config_json,publish_status", &snapshot.LobbySections}, {"game_games", "id,platform_enabled,is_hot,is_featured,is_new,sort,status,maintenance_message,platform_category_id,platform_category_code,config_status", &snapshot.Games}, {"game_site_banners", "id,name,title,subtitle,desktop_asset_id,mobile_asset_id,action_type,action_target,sort_order,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,published_payload,publish_status", &snapshot.Banners}, {"game_site_announcements", "id,title,summary,content,announcement_type,priority,sort_order,is_pinned,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,published_payload,publish_status", &snapshot.Announcements}, {"game_tenant_popup_campaigns", "id,popup_component_id,campaign_name,action_type,action_target,trigger_type,trigger_value,frequency_type,display_limit,cooldown_minutes,visibility_scope,start_at,end_at,priority,enabled,published_payload,publish_status", &snapshot.Popups}, {"game_h5_navigation_items", "id,menu_code,label_zh,label_en,labels_json,icon_asset_id,lottie_asset_id,sound_asset_id,display_position,action_type,action_target,page_route,provider_code,game_id,provider_game_code,game_name,sort_order,badge_type,visibility_scope,country_codes,language_codes,start_at,end_at,enabled,published_payload,publish_status", &snapshot.Navigation}}
	for _, table := range tables {
		if err := tx.Table(table.name).Select(table.columns).Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", tenantID, siteID).Order("id ASC").Find(table.target).Error; err != nil {
			return releaseSnapshot{}, err
		}
	}
	return snapshot, nil
}

func (s *Service) ListReleases(q Query, actor Actor) (Result, error) {
	rows := []ConfigReleaseRow{}
	if !s.db.Migrator().HasTable("game_config_releases") {
		page, pageSize := q.Page, q.PageSize
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 20
		}
		return Result{List: rows, Page: page, PageSize: pageSize}, nil
	}
	db := s.db.Table("game_config_releases").
		Select("id,tenant_id,site_id,environment,version,status,modules,rollback_source_id,failure_reason,created_by,published_by,created_at,published_at,updated_at").
		Where("deleted_at IS NULL")
	db, _ = operationalScope(db, actor, value(q, "tenantID"), value(q, "siteID"))
	db = exact(db, "status", strings.ToUpper(value(q, "status")))
	db = exact(db, "environment", strings.ToUpper(value(q, "environment")))
	return paginate(db.Order("id DESC"), q, &rows)
}

func (s *Service) ReleaseDetail(id uint64, actor Actor) (map[string]any, error) {
	var row map[string]any
	db := s.db.Table("game_config_releases").Where("id=? AND deleted_at IS NULL", id)
	db, _ = operationalScope(db, actor, "", "")
	if err := db.Take(&row).Error; err != nil {
		return nil, rowNotFound("release")
	}
	audits := []map[string]any{}
	_ = s.db.Table("game_admin_operation_logs").Select("id,action,result,operator_username operator,reason,correlation_id requestID,created_at createdAt").Where("target_type IN ('site','config_release') AND (target_id=? OR correlation_id=?) AND deleted_at IS NULL", row["site_id"], fmt.Sprint(row["idempotency_key"])).Order("id DESC").Limit(100).Find(&audits).Error
	return map[string]any{"release": row, "audit": audits}, nil
}

func (s *Service) CompareReleases(fromID, toID uint64, actor Actor) (map[string]any, error) {
	from, err := s.releaseSnapshotForActor(fromID, actor)
	if err != nil {
		return nil, err
	}
	to, err := s.releaseSnapshotForActor(toID, actor)
	if err != nil {
		return nil, err
	}
	if from.TenantID != to.TenantID || from.SiteID != to.SiteID {
		return nil, NewError("RELEASE_SCOPE_MISMATCH", "releases belong to different sites")
	}
	return map[string]any{"fromVersion": from.Version, "toVersion": to.Version, "before": json.RawMessage(from.Snapshot), "after": json.RawMessage(to.Snapshot)}, nil
}

type storedReleaseSnapshot struct {
	ConfigReleaseRow
	Snapshot string
}

func (s *Service) releaseSnapshotForActor(id uint64, actor Actor) (storedReleaseSnapshot, error) {
	var row storedReleaseSnapshot
	db := s.db.Table("game_config_releases").Select("*").Where("id=? AND status IN ('PUBLISHED','ROLLED_BACK') AND deleted_at IS NULL", id)
	db, _ = operationalScope(db, actor, "", "")
	if err := db.Take(&row).Error; err != nil {
		return row, rowNotFound("release")
	}
	return row, nil
}

func (s *Service) ValidateRelease(siteID uint64, actor Actor) (map[string]any, error) {
	row, err := s.siteRowForActor(s.db, siteID, actor, false)
	if err != nil {
		return nil, err
	}
	modules, pending, err := pendingReleaseModules(s.db, row)
	if err != nil {
		return nil, err
	}
	_, previewErr := s.PreviewSiteConfig(siteID, actor)
	errors := []string{}
	if previewErr != nil {
		errors = append(errors, previewErr.Error())
	}
	return map[string]any{"valid": len(errors) == 0 && pending, "hasPendingChanges": pending, "modules": modules, "errors": errors}, nil
}

func (s *Service) RollbackRelease(sourceID uint64, input ReleaseRollbackInput, actor Actor) (map[string]any, error) {
	if err := requireReason(input.Reason); err != nil {
		return nil, err
	}
	source, err := s.releaseSnapshotForActor(sourceID, actor)
	if err != nil {
		return nil, err
	}
	key := releaseIdempotencyKey(source.TenantID, source.SiteID, input.IdempotencyKey)
	var result map[string]any
	var failed failedReleaseAttempt
	err = s.db.Transaction(func(tx *gorm.DB) error {
		site, err := s.siteRowForActor(tx, source.SiteID, actor, true)
		if err != nil {
			return err
		}
		if existing, ok, err := existingReleaseByKey(tx, site.TenantID, site.ID, key); err != nil {
			return err
		} else if ok {
			result = releaseResult(existing, true)
			return nil
		}
		version, err := nextReleaseVersion(tx, site.TenantID, site.ID, site.ConfigVersion)
		if err != nil {
			return err
		}
		releaseID, err := createPublishingRelease(tx, site.TenantID, site.ID, version, []string{"ROLLBACK"}, key, source.ID, actor.ID)
		if err != nil {
			return err
		}
		failed = failedReleaseAttempt{TenantID: site.TenantID, SiteID: site.ID, Version: version, Key: key, Modules: "ROLLBACK", CreatedBy: actor.ID}
		var snapshot releaseSnapshot
		if err := json.Unmarshal([]byte(source.Snapshot), &snapshot); err != nil {
			return NewError("INVALID_RELEASE_SNAPSHOT", "release snapshot is invalid")
		}
		if err := restoreReleaseSnapshot(tx, site.TenantID, site.ID, version, actor.ID, snapshot); err != nil {
			return err
		}
		if err := clearWhiteLabelConfigCache(tx, site.TenantID, site.ID); err != nil {
			return err
		}
		publishedAt := now()
		if _, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{Scope: tenancy.Scope{TenantID: site.TenantID, SiteID: site.ID, Platform: actor.Platform}, TenantID: site.TenantID, SiteID: site.ID, EventType: eventbus.EventWhiteLabelConfigPublished, AggregateType: "white_label_config", AggregateID: fmt.Sprint(site.ID), Source: "systemadmin", IdempotencyKey: fmt.Sprintf("white_label.config_rollback:%d:%d", site.ID, version), TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID, Payload: map[string]any{"tenantId": site.TenantID, "siteId": site.ID, "configVersion": version, "rollbackSourceID": source.ID, "publishedAt": publishedAt}}); err != nil {
			return err
		}
		fresh, err := captureReleaseSnapshot(tx, site.TenantID, site.ID)
		if err != nil {
			return err
		}
		if err := tx.Table("game_config_releases").Where("id=?", releaseID).Updates(map[string]any{"status": "PUBLISHED", "snapshot": marshal(fresh), "validation_result": marshal(map[string]any{"valid": true, "rollbackSourceID": source.ID}), "published_by": actor.ID, "published_at": publishedAt, "updated_at": publishedAt}).Error; err != nil {
			return err
		}
		if err := tx.Table("game_config_releases").Where("id=?", source.ID).Update("status", "ROLLED_BACK").Error; err != nil {
			return err
		}
		if err := audit(tx, targetAuditActor(actor, site.TenantID, site.ID), "release:rollback", "configuration.release.rollback", "config_release", releaseID, map[string]any{"currentVersion": site.ConfigVersion}, map[string]any{"version": version, "sourceID": source.ID}, input.Reason); err != nil {
			return err
		}
		result = map[string]any{"releaseID": releaseID, "tenantId": site.TenantID, "siteId": site.ID, "configVersion": version, "rollbackSourceID": source.ID, "publishedAt": publishedAt, "idempotent": false}
		return nil
	})
	if err != nil && failed.Key != "" {
		s.recordFailedRelease(failed, err)
	}
	return result, err
}

func restoreReleaseSnapshot(tx *gorm.DB, tenantID, siteID, version, actorID uint64, snapshot releaseSnapshot) error {
	site := snapshot.Site
	updates := map[string]any{"site_name": site["siteName"], "country_code": site["countryCode"], "currency_code": site["currencyCode"], "default_language": site["defaultLanguage"], "timezone": site["timezone"], "brand_config": marshal(site["brandConfig"]), "theme_config": marshal(site["themeConfig"]), "feature_config": marshal(site["featureConfig"]), "maintenance_config": marshal(site["maintenanceConfig"]), "seo_config": marshal(site["seoConfig"]), "site_name_draft": site["siteName"], "country_code_draft": site["countryCode"], "currency_code_draft": site["currencyCode"], "default_language_draft": site["defaultLanguage"], "timezone_draft": site["timezone"], "brand_config_draft": marshal(site["brandConfig"]), "theme_config_draft": marshal(site["themeConfig"]), "feature_config_draft": marshal(site["featureConfig"]), "maintenance_config_draft": marshal(site["maintenanceConfig"]), "seo_config_draft": marshal(site["seoConfig"]), "config_status": "PUBLISHED", "config_version": version, "config_published_at": now(), "config_published_by": actorID, "updated_by": actorID, "updated_at": now()}
	if err := tx.Table("game_saas_sites").Where("id=? AND tenant_id=? AND deleted_at IS NULL", siteID, tenantID).Updates(updates).Error; err != nil {
		return err
	}
	tables := []struct {
		name    string
		rows    []map[string]any
		disable map[string]any
		allowed map[string]bool
	}{{"game_lobby_sections", snapshot.LobbySections, map[string]any{"enabled": 0}, allowedReleaseColumns("id,section_code,section_type,title_zh,title_en,display_style,source_type,source_ref,max_items,sort_order,enabled,supported_countries,supported_currencies,config_json,publish_status")}, {"game_games", snapshot.Games, map[string]any{"platform_enabled": 0, "status": 3}, allowedReleaseColumns("id,platform_enabled,is_hot,is_featured,is_new,sort,status,maintenance_message,platform_category_id,platform_category_code,config_status")}, {"game_site_banners", snapshot.Banners, map[string]any{"enabled": 0}, allowedReleaseColumns("id,name,title,subtitle,desktop_asset_id,mobile_asset_id,action_type,action_target,sort_order,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,published_payload,publish_status")}, {"game_site_announcements", snapshot.Announcements, map[string]any{"enabled": 0}, allowedReleaseColumns("id,title,summary,content,announcement_type,priority,sort_order,is_pinned,language_code,country_code,currency_code,visibility_scope,start_at,end_at,enabled,published_payload,publish_status")}, {"game_tenant_popup_campaigns", snapshot.Popups, map[string]any{"enabled": 0}, allowedReleaseColumns("id,popup_component_id,campaign_name,action_type,action_target,trigger_type,trigger_value,frequency_type,display_limit,cooldown_minutes,visibility_scope,start_at,end_at,priority,enabled,published_payload,publish_status")}, {"game_h5_navigation_items", snapshot.Navigation, map[string]any{"enabled": 0}, allowedReleaseColumns("id,menu_code,label_zh,label_en,labels_json,icon_asset_id,lottie_asset_id,sound_asset_id,display_position,action_type,action_target,page_route,provider_code,game_id,provider_game_code,game_name,sort_order,badge_type,visibility_scope,country_codes,language_codes,start_at,end_at,enabled,published_payload,publish_status")}}
	for _, table := range tables {
		if err := tx.Table(table.name).Where("tenant_id=? AND site_id=? AND deleted_at IS NULL", tenantID, siteID).Updates(table.disable).Error; err != nil {
			return err
		}
		for _, raw := range table.rows {
			id := uint64FromJSON(raw["id"])
			values := map[string]any{}
			for key, value := range raw {
				if key != "id" && table.allowed[key] {
					values[key] = value
				}
			}
			values["config_version"] = version
			values["published_at"] = now()
			values["published_by"] = actorID
			values["updated_at"] = now()
			values["deleted_at"] = nil
			if payload, ok := values["published_payload"]; ok {
				values["draft_payload"] = payload
			}
			if table.name == "game_tenant_popup_campaigns" {
				values["popup_component_id_draft"] = values["popup_component_id"]
			}
			if err := tx.Table(table.name).Where("id=? AND tenant_id=? AND site_id=?", id, tenantID, siteID).Updates(values).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func allowedReleaseColumns(value string) map[string]bool {
	result := map[string]bool{}
	for _, column := range strings.Split(value, ",") {
		result[column] = true
	}
	return result
}
func uint64FromJSON(value any) uint64 {
	switch v := value.(type) {
	case float64:
		return uint64(v)
	case uint64:
		return v
	case int64:
		return uint64(v)
	case json.Number:
		result, _ := v.Int64()
		return uint64(result)
	default:
		return parseID(fmt.Sprint(value))
	}
}
func splitCSV(value string) []string {
	result := []string{}
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
