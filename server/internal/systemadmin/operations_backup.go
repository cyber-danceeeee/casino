package systemadmin

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"casino/internal/content"
	"casino/internal/eventbus"
	"casino/internal/storage"

	"gorm.io/gorm"
)

const (
	backupScopeDatabaseConfig        = "DATABASE_CONFIGURATION"
	backupScopeH5Config              = "H5_CONFIGURATION"
	backupScopePublishedConfig       = "PUBLISHED_CONFIGURATION"
	backupScopeAssetMetadata         = "ASSET_METADATA"
	backupScopeAssetFiles            = "ASSET_FILES"
	backupMaxArtifactBytes     int64 = 768 << 20
)

type BackupTaskRow struct {
	ID        uint64 `json:"id"`
	BackupNo  string `json:"backupNo" gorm:"column:backup_no"`
	ScopeType string `json:"scopeType" gorm:"column:scope_type"`

	SiteID          uint64     `json:"siteID" gorm:"column:site_id"`
	Status          string     `json:"status"`
	StorageLocation string     `json:"storageLocation" gorm:"column:storage_location"`
	ArtifactKey     string     `json:"-" gorm:"column:artifact_key"`
	SizeBytes       uint64     `json:"sizeBytes" gorm:"column:size_bytes"`
	ChecksumSHA256  string     `json:"checksumSha256" gorm:"column:checksum_sha256"`
	Encrypted       bool       `json:"encrypted"`
	CreatedBy       uint64     `json:"createdBy" gorm:"column:created_by"`
	StartedAt       *time.Time `json:"startedAt" gorm:"column:started_at"`
	CompletedAt     *time.Time `json:"completedAt" gorm:"column:completed_at"`
	ExpiresAt       *time.Time `json:"expiresAt" gorm:"column:expires_at"`
	FailureReason   string     `json:"failureReason" gorm:"column:failure_reason"`
	CreatedAt       time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt       time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

type BackupCreateInput struct {
	ScopeType string `json:"scopeType"`

	SiteID         uint64 `json:"siteID"`
	RetentionDays  int    `json:"retentionDays"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type BackupRestoreInput struct {
	Reason       string `json:"reason"`
	Confirmation string `json:"confirmation"`
	DryRun       bool   `json:"dryRun"`
}

type backupManifest struct {
	FormatVersion int    `json:"formatVersion"`
	BackupNo      string `json:"backupNo"`
	ScopeType     string `json:"scopeType"`

	SiteID    uint64          `json:"siteID"`
	CreatedAt time.Time       `json:"createdAt"`
	Data      json.RawMessage `json:"data"`
}

type h5ConfigurationBackup struct {
	Sites map[string]releaseSnapshot `json:"sites"`
}

type backupAssetRow struct {
	ID uint64 `json:"id"`

	SiteID         uint64 `json:"siteID" gorm:"column:site_id"`
	AssetScope     string `json:"assetScope" gorm:"column:asset_scope"`
	AssetType      string `json:"assetType" gorm:"column:asset_type"`
	UsageType      string `json:"usageType" gorm:"column:usage_type"`
	DisplayName    string `json:"displayName" gorm:"column:display_name"`
	OriginalName   string `json:"originalName" gorm:"column:original_name"`
	ObjectKey      string `json:"objectKey" gorm:"column:object_key"`
	MimeType       string `json:"mimeType" gorm:"column:mime_type"`
	FileSize       int64  `json:"fileSize" gorm:"column:file_size"`
	ChecksumSHA256 string `json:"checksumSha256" gorm:"column:checksum_sha256"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	DurationMS     int64  `json:"durationMs" gorm:"column:duration_ms"`
	Status         string `json:"status"`
}

func (s *Service) ListBackups(q Query) (Result, error) {
	rows := []BackupTaskRow{}
	db := s.db.Table("game_backup_tasks").Where("deleted_at IS NULL")
	db = exact(db, "scope_type", strings.ToUpper(value(q, "scopeType")))
	db = exact(db, "status", strings.ToUpper(value(q, "status")))
	if siteID := parseID(value(q, "siteID")); siteID > 0 {
		db = db.Where("site_id=?", siteID)
	}
	result, err := paginate(db.Order("created_at DESC,id DESC"), q, &rows)
	if err != nil {
		return result, err
	}
	for i := range rows {
		rows[i].StorageLocation = maskBackupLocation(rows[i].StorageLocation)
	}
	result.List = rows
	return result, nil
}

func (s *Service) BackupDetail(id uint64) (BackupTaskRow, error) {
	var row BackupTaskRow
	if err := s.db.Table("game_backup_tasks").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return row, rowNotFound("backup")
	}
	row.StorageLocation = maskBackupLocation(row.StorageLocation)
	return row, nil
}

func (s *Service) CreateBackup(input BackupCreateInput, actor Actor) (BackupTaskRow, error) {
	if !actor.Platform || actor.SiteID != 0 {
		return BackupTaskRow{}, NewError("PERMISSION_DENIED", "permission denied")
	}
	scope := strings.ToUpper(strings.TrimSpace(input.ScopeType))
	if !validBackupScope(scope) {
		return BackupTaskRow{}, NewError("INVALID_BACKUP_SCOPE", "invalid backup scope")
	}
	if scopeRequiresSite(scope) && input.SiteID == 0 {
		return BackupTaskRow{}, NewError("SITE_REQUIRED", "H5 configuration is required for this backup scope")
	}
	if input.SiteID > 0 {
		var count int64
		if err := s.db.Table("game_site_config").Where("id = ? AND deleted_at IS NULL", input.SiteID).Count(&count).Error; err != nil || count == 0 {
			return BackupTaskRow{}, rowNotFound("site")
		}
	}
	retention := input.RetentionDays
	if retention <= 0 {
		retention = 30
	}
	if retention > 365 {
		return BackupTaskRow{}, NewError("INVALID_RETENTION", "retention must not exceed 365 days")
	}
	key := backupIdempotencyKey(actor.ID, input.IdempotencyKey)
	var existing BackupTaskRow
	err := s.db.Table("game_backup_tasks").Where("created_by=? AND idempotency_key=? AND deleted_at IS NULL", actor.ID, key).Take(&existing).Error
	if err == nil {
		return existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return BackupTaskRow{}, err
	}
	backupNo, err := newBackupNo()
	if err != nil {
		return BackupTaskRow{}, err
	}
	expiresAt := now().AddDate(0, 0, retention)
	row := map[string]any{"backup_no": backupNo, "scope_type": scope, "site_id": input.SiteID, "status": "PENDING", "idempotency_key": key, "created_by": actor.ID, "expires_at": expiresAt, "created_at": now(), "updated_at": now()}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_backup_tasks").Create(row).Error; err != nil {
			return duplicateError(err)
		}
		var taskID uint64
		if err := tx.Table("game_backup_tasks").Select("id").Where("backup_no=?", backupNo).Scan(&taskID).Error; err != nil {
			return err
		}
		if err := audit(tx, targetAuditActor(actor, input.SiteID), "backup:create", "backup.task.create", "backup_task", taskID, nil, map[string]any{"backupNo": backupNo, "scopeType": scope, "retentionDays": retention}, "create verified backup"); err != nil {
			return err
		}
		_, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope: content.Scope{Platform: true}, SiteID: input.SiteID,
			EventType: eventbus.EventBackupRequested, AggregateType: "backup_task", AggregateID: fmt.Sprint(taskID),
			Source: "systemadmin", Producer: "systemadmin", IdempotencyKey: "backup-task:" + backupNo,
			RequestID: actor.CorrelationID, TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID,
			TargetConsumers: []string{eventbus.ConsumerBackup}, Payload: map[string]any{"taskID": taskID},
		})
		return err
	}); err != nil {
		return BackupTaskRow{}, err
	}
	var created BackupTaskRow
	if err := s.db.Table("game_backup_tasks").Where("backup_no=?", backupNo).Take(&created).Error; err != nil {
		return BackupTaskRow{}, err
	}
	eventbus.SignalDispatch()
	return created, nil
}

func (s *Service) executeBackupTask(id uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	if err := s.runBackupTask(ctx, id); err != nil {
		var status string
		_ = s.db.Table("game_backup_tasks").Select("status").Where("id=?", id).Scan(&status).Error
		if status == "SUCCESS" {
			return nil
		}
		message := sanitizeFailure(err.Error())
		completed := now()
		_ = s.db.Table("game_backup_tasks").Where("id=? AND status IN ('PENDING','RUNNING')", id).Updates(map[string]any{"status": "FAILED", "failure_reason": message, "completed_at": completed, "updated_at": completed}).Error
		var row BackupTaskRow
		if s.db.Table("game_backup_tasks").Where("id=?", id).Take(&row).Error == nil {
			_ = s.raiseAlert(alertFingerprint("BACKUP_FAILED", row.BackupNo, row.SiteID), row.SiteID, "WARNING", "BACKUP_FAILED", "backup-center", "Backup task failed", map[string]any{"backupNo": row.BackupNo, "error": message})
		}
		return err
	}
	return nil
}

func (s *Service) HandleBackupEvent(ctx eventbus.Context) error {
	id := uint64FromJSON(ctx.Event.Payload["taskID"])
	if id == 0 {
		return NewError("BACKUP_TASK_REQUIRED", "backup task is required")
	}
	var row BackupTaskRow
	if err := s.db.Table("game_backup_tasks").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return err
	}
	if row.SiteID != ctx.Event.SiteID {
		return NewError("BACKUP_SCOPE_MISMATCH", "backup task is outside event scope")
	}
	return s.executeBackupTask(id)
}

func (s *Service) runBackupTask(ctx context.Context, id uint64) error {
	started := now()
	claimed := s.db.Table("game_backup_tasks").Where("id=? AND (status IN ('PENDING','FAILED') OR (status='RUNNING' AND updated_at<?)) AND deleted_at IS NULL", id, started.Add(-25*time.Minute)).Updates(map[string]any{"status": "RUNNING", "started_at": started, "failure_reason": "", "updated_at": started})
	if claimed.Error != nil {
		return claimed.Error
	}
	if claimed.RowsAffected != 1 {
		return NewError("BACKUP_NOT_PENDING", "backup task is not pending")
	}
	var row BackupTaskRow
	if err := s.db.Table("game_backup_tasks").Where("id=?", id).Take(&row).Error; err != nil {
		return err
	}
	artifact, err := s.buildBackupArtifact(ctx, row)
	if err != nil {
		return err
	}
	if int64(len(artifact)) > backupMaxArtifactBytes {
		return NewError("BACKUP_TOO_LARGE", "backup artifact exceeds configured limit")
	}
	key := path.Join("backups", started.Format("2006/01"), row.BackupNo+".tar.gz")
	asset, err := storage.PutBytes(ctx, key, artifact, "application/gzip")
	if err != nil {
		return fmt.Errorf("write backup artifact: %w", err)
	}
	verified, err := storage.ReadBytes(ctx, asset.Key, backupMaxArtifactBytes)
	if err != nil {
		return fmt.Errorf("read backup artifact for verification: %w", err)
	}
	expected := sha256.Sum256(artifact)
	actual := sha256.Sum256(verified)
	if expected != actual {
		_ = storage.Delete(ctx, asset.Key)
		return NewError("BACKUP_CHECKSUM_MISMATCH", "backup artifact checksum verification failed")
	}
	completed := now()
	return s.db.Table("game_backup_tasks").Where("id=? AND status='RUNNING'", id).Updates(map[string]any{"status": "SUCCESS", "storage_location": "configured-object-storage", "artifact_key": asset.Key, "size_bytes": len(artifact), "checksum_sha256": hex.EncodeToString(actual[:]), "completed_at": completed, "failure_reason": "", "updated_at": completed}).Error
}

func (s *Service) buildBackupArtifact(ctx context.Context, row BackupTaskRow) ([]byte, error) {
	data, assets, err := s.backupData(row)
	if err != nil {
		return nil, err
	}
	manifest := backupManifest{FormatVersion: 1, BackupNo: row.BackupNo, ScopeType: row.ScopeType, SiteID: row.SiteID, CreatedAt: now(), Data: data}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := writeTarFile(tarWriter, "manifest.json", manifestBytes); err != nil {
		return nil, err
	}
	if row.ScopeType == backupScopeAssetFiles {
		var total int64
		for _, item := range assets {
			if strings.TrimSpace(item.ObjectKey) == "" {
				continue
			}
			contents, err := storage.ReadBytes(ctx, item.ObjectKey, 128<<20)
			if err != nil {
				return nil, fmt.Errorf("read asset %d: %w", item.ID, err)
			}
			total += int64(len(contents))
			if total > backupMaxArtifactBytes {
				return nil, NewError("BACKUP_TOO_LARGE", "asset backup exceeds configured limit")
			}
			name := path.Join("assets", fmt.Sprint(item.ID), path.Base(item.ObjectKey))
			if err := writeTarFile(tarWriter, name, contents); err != nil {
				return nil, err
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (s *Service) backupData(row BackupTaskRow) (json.RawMessage, []backupAssetRow, error) {
	switch row.ScopeType {
	case backupScopeH5Config:
		data := h5ConfigurationBackup{Sites: map[string]releaseSnapshot{}}
		sites := []uint64{}
		db := s.db.Table("game_site_config").Where("deleted_at IS NULL")
		if row.SiteID > 0 {
			db = db.Where("id=?", row.SiteID)
		}
		if err := db.Pluck("id", &sites).Error; err != nil {
			return nil, nil, err
		}
		if len(sites) == 0 {
			return nil, nil, rowNotFound("site")
		}
		for _, siteID := range sites {
			snapshot, err := captureReleaseSnapshot(s.db, siteID)
			if err != nil {
				return nil, nil, err
			}
			data.Sites[fmt.Sprint(siteID)] = snapshot
		}
		return marshalRaw(data), nil, nil
	case backupScopePublishedConfig:
		rows := []map[string]any{}
		db := s.db.Table("game_config_releases").Select("id, site_id, environment, version, status, modules, snapshot, rollback_source_id, created_by, published_by, created_at, published_at").Where("deleted_at IS NULL AND status IN ('PUBLISHED','ROLLED_BACK')")
		if row.SiteID > 0 {
			db = db
		}
		if row.SiteID > 0 {
			db = db.Where("site_id=?", row.SiteID)
		}
		if err := db.Order("site_id, version").Find(&rows).Error; err != nil {
			return nil, nil, err
		}
		return marshalRaw(map[string]any{"releases": rows}), nil, nil
	case backupScopeAssetMetadata, backupScopeAssetFiles:
		assets := []backupAssetRow{}
		db := s.db.Table("game_asset_library").Where("deleted_at IS NULL")
		if row.SiteID > 0 {
			db = db
		}
		if row.SiteID > 0 {
			db = db.Where("site_id IN (0,?)", row.SiteID)
		}
		if err := db.Order("id").Find(&assets).Error; err != nil {
			return nil, nil, err
		}
		return marshalRaw(map[string]any{"assets": assets}), assets, nil
	case backupScopeDatabaseConfig:
		return s.databaseConfigurationBackup()
	default:
		return nil, nil, NewError("INVALID_BACKUP_SCOPE", "invalid backup scope")
	}
}

func (s *Service) databaseConfigurationBackup() (json.RawMessage, []backupAssetRow, error) {
	result := map[string]any{}
	tables := []string{"game_system_configs", "game_country_configs", "game_site_config", "game_providers"}
	for _, table := range tables {
		if !s.db.Migrator().HasTable(table) {
			continue
		}
		rows := []map[string]any{}
		query := s.db.Table(table)
		if s.db.Migrator().HasColumn(table, "deleted_at") {
			query = query.Where("deleted_at IS NULL")
		}
		if err := query.Find(&rows).Error; err != nil {
			return nil, nil, err
		}
		result[table] = redact(rows)
	}
	return marshalRaw(result), nil, nil
}

func (s *Service) ValidateBackupRestore(id uint64) (map[string]any, error) {
	row, manifest, err := s.verifiedBackupManifest(id)
	if err != nil {
		return nil, err
	}
	restorable := row.ScopeType == backupScopeH5Config && row.SiteID > 0
	return map[string]any{"valid": true, "restorable": restorable, "backupNo": row.BackupNo, "scopeType": row.ScopeType, "siteID": row.SiteID, "formatVersion": manifest.FormatVersion, "checksum": row.ChecksumSHA256, "confirmationPhrase": "RESTORE " + row.BackupNo}, nil
}

func (s *Service) RestoreBackup(id uint64, input BackupRestoreInput, actor Actor) (map[string]any, error) {
	if !actor.Platform || actor.SiteID != 0 {
		return nil, NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := requireReason(input.Reason); err != nil {
		return nil, err
	}
	row, manifest, err := s.verifiedBackupManifest(id)
	if err != nil {
		return nil, err
	}
	if row.ScopeType != backupScopeH5Config {
		return nil, NewError("BACKUP_NOT_RESTORABLE", "this backup scope cannot be restored by the application")
	}
	if input.DryRun {
		return s.validateH5BackupPayload(row, manifest)
	}
	if strings.TrimSpace(input.Confirmation) != "RESTORE "+row.BackupNo {
		return nil, NewError("CONFIRMATION_REQUIRED", "restore confirmation phrase does not match")
	}
	validation, err := s.validateH5BackupPayload(row, manifest)
	if err != nil {
		return nil, err
	}
	preInput := BackupCreateInput{ScopeType: backupScopeH5Config, SiteID: row.SiteID, RetentionDays: 30, IdempotencyKey: fmt.Sprintf("pre-restore:%d:%s", row.ID, actor.CorrelationID)}
	pre, err := s.createAndRunBackupSync(preInput, actor)
	if err != nil {
		return nil, fmt.Errorf("pre-restore backup failed: %w", err)
	}
	var payload h5ConfigurationBackup
	if err := json.Unmarshal(manifest.Data, &payload); err != nil {
		return nil, NewError("INVALID_BACKUP_ARTIFACT", "H5 configuration payload is invalid")
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for siteKey, snapshot := range payload.Sites {
			siteID := parseID(siteKey)
			var site SiteRow
			if err := tx.Table("game_site_config").Where("id = ? AND deleted_at IS NULL", siteID).Take(&site).Error; err != nil {
				return rowNotFound("site")
			}
			version, err := nextReleaseVersion(tx, siteID, site.ConfigVersion)
			if err != nil {
				return err
			}
			if err := restoreReleaseSnapshot(tx, siteID, version, actor.ID, snapshot); err != nil {
				return err
			}
			fresh, err := captureReleaseSnapshot(tx, siteID)
			if err != nil {
				return err
			}
			key := releaseIdempotencyKey(siteID, fmt.Sprintf("backup-restore:%d:%d", row.ID, version))
			releaseID, err := createPublishingRelease(tx, siteID, version, []string{"BACKUP_RESTORE"}, key, 0, actor.ID)
			if err != nil {
				return err
			}
			publishedAt := now()
			if err := tx.Table("game_config_releases").Where("id=?", releaseID).Updates(map[string]any{"status": "PUBLISHED", "snapshot": marshal(fresh), "validation_result": marshal(map[string]any{"valid": true, "backupID": row.ID}), "published_by": actor.ID, "published_at": publishedAt, "updated_at": publishedAt}).Error; err != nil {
				return err
			}
			if err := clearWhiteLabelConfigCache(tx, siteID); err != nil {
				return err
			}
			if _, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{Scope: content.Scope{SiteID: siteID, Platform: true}, SiteID: siteID, EventType: eventbus.EventWhiteLabelConfigPublished, AggregateType: "white_label_config", AggregateID: fmt.Sprint(siteID), Source: "systemadmin", IdempotencyKey: fmt.Sprintf("backup.restore:%d:%d", row.ID, siteID), TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID, Payload: map[string]any{"backupID": row.ID, "backupNo": row.BackupNo, "configVersion": version}}); err != nil {
				return err
			}
		}
		return audit(tx, targetAuditActor(actor, row.SiteID), "backup:restore", "backup.restore", "backup_task", row.ID, map[string]any{"preRestoreBackupID": pre.ID}, validation, input.Reason)
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"restored": true, "backupID": row.ID, "backupNo": row.BackupNo, "preRestoreBackupID": pre.ID, "validation": validation}, nil
}

func (s *Service) DeleteBackup(id uint64, reason string, actor Actor) error {
	if !actor.Platform || actor.SiteID != 0 {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if err := requireReason(reason); err != nil {
		return err
	}
	var row BackupTaskRow
	if err := s.db.Table("game_backup_tasks").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("backup")
	}
	if row.Status == "RUNNING" || row.Status == "PENDING" {
		return NewError("BACKUP_RUNNING", "running backup cannot be deleted")
	}
	if row.ArtifactKey != "" {
		if err := storage.Delete(context.Background(), row.ArtifactKey); err != nil {
			return fmt.Errorf("delete backup artifact: %w", err)
		}
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_backup_tasks").Where("id=?", id).Updates(map[string]any{"status": "DELETED", "deleted_at": now(), "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, targetAuditActor(actor, row.SiteID), "backup:delete", "backup.task.delete", "backup_task", id, row, map[string]any{"status": "DELETED"}, reason)
	})
}

func (s *Service) createAndRunBackupSync(input BackupCreateInput, actor Actor) (BackupTaskRow, error) {
	created, err := s.CreateBackup(input, actor)
	if err != nil {
		return created, err
	}
	for deadline := time.Now().Add(20 * time.Minute); time.Now().Before(deadline); {
		var row BackupTaskRow
		if err := s.db.Table("game_backup_tasks").Where("id=?", created.ID).Take(&row).Error; err != nil {
			return row, err
		}
		if row.Status == "SUCCESS" {
			return row, nil
		}
		if row.Status == "FAILED" {
			return row, NewError("BACKUP_FAILED", row.FailureReason)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return created, NewError("BACKUP_TIMEOUT", "pre-restore backup timed out")
}

func (s *Service) verifiedBackupManifest(id uint64) (BackupTaskRow, backupManifest, error) {
	var row BackupTaskRow
	if err := s.db.Table("game_backup_tasks").Where("id=? AND status='SUCCESS' AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return row, backupManifest{}, rowNotFound("verified backup")
	}
	artifact, err := storage.ReadBytes(context.Background(), row.ArtifactKey, backupMaxArtifactBytes)
	if err != nil {
		return row, backupManifest{}, err
	}
	manifest, err := verifyBackupArtifact(row, artifact)
	return row, manifest, err
}

func verifyBackupArtifact(row BackupTaskRow, artifact []byte) (backupManifest, error) {
	sum := sha256.Sum256(artifact)
	if !strings.EqualFold(row.ChecksumSHA256, hex.EncodeToString(sum[:])) {
		return backupManifest{}, NewError("BACKUP_CHECKSUM_MISMATCH", "backup artifact checksum verification failed")
	}
	gzipReader, err := gzip.NewReader(bytes.NewReader(artifact))
	if err != nil {
		return backupManifest{}, NewError("INVALID_BACKUP_ARTIFACT", "backup compression is invalid")
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return backupManifest{}, NewError("INVALID_BACKUP_ARTIFACT", "backup archive cannot be read")
		}
		if header.Name != "manifest.json" || header.Size > 128<<20 {
			continue
		}
		contents := make([]byte, header.Size)
		if _, err := io.ReadFull(tarReader, contents); err != nil {
			return backupManifest{}, NewError("INVALID_BACKUP_ARTIFACT", "backup manifest cannot be read")
		}
		var manifest backupManifest
		if err := json.Unmarshal(contents, &manifest); err != nil || manifest.FormatVersion != 1 || manifest.BackupNo != row.BackupNo || manifest.SiteID != row.SiteID || manifest.ScopeType != row.ScopeType {
			return backupManifest{}, NewError("INVALID_BACKUP_ARTIFACT", "backup manifest metadata does not match")
		}
		return manifest, nil
	}
	return backupManifest{}, NewError("INVALID_BACKUP_ARTIFACT", "backup manifest is missing")
}

func (s *Service) validateH5BackupPayload(row BackupTaskRow, manifest backupManifest) (map[string]any, error) {
	var payload h5ConfigurationBackup
	if err := json.Unmarshal(manifest.Data, &payload); err != nil || len(payload.Sites) == 0 {
		return nil, NewError("INVALID_BACKUP_ARTIFACT", "H5 configuration payload is invalid")
	}
	validated := []uint64{}
	for siteKey, snapshot := range payload.Sites {
		siteID := parseID(siteKey)
		if siteID == 0 || len(snapshot.Site) == 0 {
			return nil, NewError("INVALID_BACKUP_ARTIFACT", "site snapshot is invalid")
		}
		var count int64
		if err := s.db.Table("game_site_config").Where("id = ? AND deleted_at IS NULL", siteID).Count(&count).Error; err != nil || count == 0 {
			return nil, NewError("RESTORE_SCOPE_MISMATCH", "backup contains a site outside the H5 configuration")
		}
		validated = append(validated, siteID)
	}
	return map[string]any{"valid": true, "siteIDs": validated, "siteCount": len(validated)}, nil
}

func validBackupScope(scope string) bool {
	switch scope {
	case backupScopeDatabaseConfig, backupScopeH5Config, backupScopePublishedConfig, backupScopeAssetMetadata, backupScopeAssetFiles:
		return true
	default:
		return false
	}
}

func scopeRequiresSite(scope string) bool {
	return scope == backupScopeH5Config
}

func backupIdempotencyKey(actorID uint64, provided string) string {
	provided = strings.TrimSpace(provided)
	if provided == "" {
		provided = fmt.Sprintf("auto:%d", time.Now().UnixNano())
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("backup|%d|%s", actorID, provided)))
	return "backup:" + hex.EncodeToString(sum[:])
}

func newBackupNo() (string, error) {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return "BKP" + time.Now().UTC().Format("20060102150405") + strings.ToUpper(hex.EncodeToString(random)), nil
}

func writeTarFile(writer *tar.Writer, name string, contents []byte) error {
	clean := path.Clean("/" + name)
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || clean == "." || strings.Contains(clean, "..") {
		return NewError("INVALID_ARTIFACT_PATH", "backup artifact path is invalid")
	}
	if err := writer.WriteHeader(&tar.Header{Name: clean, Mode: 0o600, Size: int64(len(contents)), ModTime: now()}); err != nil {
		return err
	}
	_, err := writer.Write(contents)
	return err
}

func marshalRaw(value any) json.RawMessage {
	data, _ := json.Marshal(value)
	return data
}

func maskBackupLocation(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "managed-object-storage"
}
