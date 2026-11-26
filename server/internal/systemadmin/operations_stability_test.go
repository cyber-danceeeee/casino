package systemadmin

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOperationsTenantScopeCannotBeOverridden(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	current := time.Now().UTC()
	for _, row := range []struct{ id, tenantID, siteID int }{{1, 1, 101}, {2, 2, 202}} {
		if err := db.Exec(`INSERT INTO game_tenant_maintenance
			(id,tenant_id,site_id,scope_type,enabled,reason,created_at,updated_at)
			VALUES (?,?,?,'SITE',1,'test',?,?)`, row.id, row.tenantID, row.siteID, current, current).Error; err != nil {
			t.Fatal(err)
		}
	}
	result, err := service.ListTenantMaintenance(Query{Values: map[string]string{"tenantID": "2", "siteID": "202"}}, Actor{TenantID: 1, SiteID: 101})
	if err != nil {
		t.Fatal(err)
	}
	rows, ok := result.List.(*[]tenantMaintenanceRow)
	if !ok || len(*rows) != 1 || (*rows)[0].TenantID != 1 || (*rows)[0].SiteID != 101 {
		t.Fatalf("tenant scope escaped: %#v", result.List)
	}
}

func TestSupportOrderSearchIsScopedAndFiltersStatus(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	current := time.Now().UTC()
	for _, row := range []struct {
		id, tenantID, siteID, status int
		orderNo                      string
	}{{1, 1, 101, 2, "shared-order"}, {2, 1, 101, 1, "pending-order"}, {3, 2, 202, 2, "shared-order"}} {
		if err := db.Exec(`INSERT INTO game_deposit_orders
			(id,tenant_id,site_id,user_id,order_no,channel_order_no,amount,currency,status,channel_code,created_at)
			VALUES (?,?,?,77,?,'external',100,'USD',?,'test',?)`, row.id, row.tenantID, row.siteID, row.orderNo, row.status, current).Error; err != nil {
			t.Fatal(err)
		}
	}
	result, err := service.SearchOrders(Query{Values: map[string]string{"query": "shared-order", "status": "SUCCESS", "tenantID": "2", "siteID": "202"}}, Actor{TenantID: 1, SiteID: 101})
	if err != nil {
		t.Fatal(err)
	}
	rows, ok := result.List.([]map[string]any)
	if !ok || len(rows) != 1 || fmt.Sprint(rows[0]["tenantID"]) != "1" || fmt.Sprint(rows[0]["status"]) != "SUCCESS" {
		t.Fatalf("unexpected scoped order result: %#v", result.List)
	}
}

func TestSupportCaseDetailIsTenantScoped(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	current := time.Now().UTC()
	for _, row := range []struct{ id, tenantID, siteID int }{{1, 1, 101}, {2, 2, 202}} {
		if err := db.Exec(`INSERT INTO game_support_cases
			(id,case_no,tenant_id,site_id,player_id,category,priority,status,title,created_at,updated_at)
			VALUES (?,?,?,?,77,'ACCOUNT','NORMAL','OPEN','test',?,?)`, row.id, fmt.Sprintf("SC-%d", row.id), row.tenantID, row.siteID, current, current).Error; err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.SupportCaseDetail(2, Actor{TenantID: 1, SiteID: 101}); ErrorCode(err) != "SUPPORT_CASE_NOT_FOUND" {
		t.Fatalf("tenant A accessed tenant B case: %v", err)
	}
	detail, err := service.SupportCaseDetail(1, Actor{TenantID: 1, SiteID: 101})
	if err != nil || detail["case"] == nil {
		t.Fatalf("tenant A could not read its own case: %#v %v", detail, err)
	}
}

func TestTenantMaintenanceWriteCannotEscapeActorScope(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	current := time.Now().UTC()
	if err := db.Exec(`INSERT INTO game_saas_sites (id,tenant_id,site_code,created_at,updated_at) VALUES
		(101,1,'site-a',?,?),(202,2,'site-b',?,?)`, current, current, current, current).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.SaveTenantMaintenance(TenantMaintenanceInput{TenantID: 2, SiteID: 202, ScopeType: "SITE", Enabled: true, Reason: "planned"}, Actor{ID: 9, Username: "tenant-a", TenantID: 1, SiteID: 101}); err != nil {
		t.Fatal(err)
	}
	var tenantA, tenantB int64
	_ = db.Table("game_tenant_maintenance").Where("tenant_id=1 AND site_id=101").Count(&tenantA).Error
	_ = db.Table("game_tenant_maintenance").Where("tenant_id=2 AND site_id=202").Count(&tenantB).Error
	if tenantA != 1 || tenantB != 0 {
		t.Fatalf("maintenance escaped actor scope: tenantA=%d tenantB=%d", tenantA, tenantB)
	}
}

func TestSupportMaskingAndCSVFormulaProtection(t *testing.T) {
	row := PlayerSupportRow{Phone: "60123456789", Email: "operator@example.com", RegisteredIP: "192.168.10.20", LastLoginIP: "10.1.2.3"}
	maskPlayerSupport(&row, false)
	if row.Phone != "*******6789" || row.Email != "o*******@example.com" || row.RegisteredIP != "192.168.*.*" || row.LastLoginIP != "10.1.*.*" {
		t.Fatalf("sensitive fields were not masked: %#v", row)
	}
	for _, value := range []string{"=1+1", "+cmd", "-10", "@SUM(A1)", "  =hidden"} {
		if escaped := escapeCSVFormula(value); len(escaped) == 0 || escaped[0] != '\'' {
			t.Fatalf("formula %q was not escaped: %q", value, escaped)
		}
	}
	if escaped := escapeCSVFormula("ordinary"); escaped != "ordinary" {
		t.Fatalf("ordinary CSV value changed: %q", escaped)
	}
}

func TestSystemAlertFingerprintDeduplicatesOccurrences(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	fingerprint := alertFingerprint("PROVIDER_UNHEALTHY", "provider-a", 1, 101)
	for index := 0; index < 2; index++ {
		if err := service.raiseAlert(fingerprint, 1, 101, "CRITICAL", "PROVIDER_UNHEALTHY", "provider-a", "Provider unavailable", map[string]any{"secret": "masked"}); err != nil {
			t.Fatal(err)
		}
	}
	var count, occurrences int64
	if err := db.Table("game_system_alerts").Where("fingerprint=?", fingerprint).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("game_system_alerts").Select("occurrence_count").Where("fingerprint=?", fingerprint).Scan(&occurrences).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 || occurrences != 2 {
		t.Fatalf("alert dedup count=%d occurrences=%d", count, occurrences)
	}
}

func TestAlertAcknowledgeAndResolveAreAudited(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	fingerprint := alertFingerprint("QUEUE_BACKLOG", "outbox", 0, 0)
	if err := service.raiseAlert(fingerprint, 0, 0, "WARNING", "QUEUE_BACKLOG", "outbox", "Queue backlog", nil); err != nil {
		t.Fatal(err)
	}
	var alertID uint64
	if err := db.Table("game_system_alerts").Select("id").Where("fingerprint=?", fingerprint).Scan(&alertID).Error; err != nil {
		t.Fatal(err)
	}
	actor := Actor{ID: 88, Username: "platform-operator", Platform: true, CorrelationID: "request-alert-test"}
	if err := service.AcknowledgeAlert(alertID, AlertActionInput{Reason: "investigating", Note: "assigned"}, actor); err != nil {
		t.Fatal(err)
	}
	if err := service.ResolveAlert(alertID, AlertActionInput{Reason: "recovered", Note: "verified"}, actor); err != nil {
		t.Fatal(err)
	}
	var auditCount int64
	if err := db.Table("game_admin_operation_logs").Where("target_type='system_alert' AND target_id=? AND operator_id=?", alertID, actor.ID).Count(&auditCount).Error; err != nil {
		t.Fatal(err)
	}
	if auditCount != 2 {
		t.Fatalf("expected acknowledge and resolve audits, got %d", auditCount)
	}
}

func TestTenantAlertsAndJobsAreScoped(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	current := time.Now().UTC()
	for _, row := range []struct{ id, tenantID, siteID int }{{1, 1, 101}, {2, 2, 202}} {
		if err := db.Exec(`INSERT INTO game_system_alerts
			(id,fingerprint,tenant_id,site_id,severity,alert_type,source,status,occurrence_count,summary,first_triggered_at,last_triggered_at,created_at,updated_at)
			VALUES (?,?,?,?, 'WARNING','PAYMENT','channel','OPEN',1,'test',?,?,?,?)`, row.id, fmt.Sprintf("alert-%d", row.id), row.tenantID, row.siteID, current, current, current, current).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`INSERT INTO game_system_retry_records
			(id,source_type,source_id,tenant_id,site_id,idempotency_key,status,attempt_count,max_attempts,created_at,updated_at)
			VALUES (?,'PAYMENT',?,?,?,?, 'FAILED',1,5,?,?)`, row.id, row.id, row.tenantID, row.siteID, fmt.Sprintf("retry-%d", row.id), current, current).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`INSERT INTO event_outbox
			(id,event_id,event_type,tenant_id,site_id,status,attempts,last_error,occurred_at,created_at,updated_at)
			VALUES (?,?, 'payment.failed.v1',?,?,'DEAD',5,'test',?,?,?)`, row.id, fmt.Sprintf("event-%d", row.id), row.tenantID, row.siteID, current, current, current).Error; err != nil {
			t.Fatal(err)
		}
	}
	actor := Actor{ID: 9, Username: "tenant-a", TenantID: 1, SiteID: 101}
	alerts, err := service.ListAlerts(Query{}, actor)
	if err != nil {
		t.Fatal(err)
	}
	alertRows, ok := alerts.List.(*[]map[string]any)
	if !ok || len(*alertRows) != 1 || fmt.Sprint((*alertRows)[0]["tenant_id"]) != "1" {
		t.Fatalf("tenant alerts escaped scope: %#v", alerts.List)
	}
	if err := service.AcknowledgeAlert(2, AlertActionInput{Reason: "cross tenant"}, actor); ErrorCode(err) != "ALERT_NOT_FOUND" {
		t.Fatalf("tenant acknowledged another tenant alert: %v", err)
	}
	jobs, err := service.JobsOverview(Query{}, actor)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs["retryRecords"].([]map[string]any)) != 1 || len(jobs["deadLetters"].([]map[string]any)) != 1 {
		t.Fatalf("tenant jobs escaped scope: %#v", jobs)
	}
}

func TestBackupArtifactChecksumAndManifestValidation(t *testing.T) {
	row := BackupTaskRow{BackupNo: "BKP-TEST", ScopeType: backupScopeTenantConfig, TenantID: 1, SiteID: 101}
	manifest := backupManifest{FormatVersion: 1, BackupNo: row.BackupNo, ScopeType: row.ScopeType, TenantID: row.TenantID, SiteID: row.SiteID, CreatedAt: time.Now(), Data: json.RawMessage(`{"sites":{"101":{"site":{"id":101}}}}`)}
	artifact := testBackupArtifact(t, manifest)
	sum := sha256.Sum256(artifact)
	row.ChecksumSHA256 = hex.EncodeToString(sum[:])
	decoded, err := verifyBackupArtifact(row, artifact)
	if err != nil || decoded.BackupNo != row.BackupNo {
		t.Fatalf("valid backup was rejected: %#v %v", decoded, err)
	}
	tampered := append([]byte(nil), artifact...)
	tampered[len(tampered)-1] ^= 0xff
	if _, err := verifyBackupArtifact(row, tampered); ErrorCode(err) != "BACKUP_CHECKSUM_MISMATCH" {
		t.Fatalf("tampered backup error = %v", err)
	}
	wrong := row
	wrong.BackupNo = "OTHER"
	wrong.ChecksumSHA256 = row.ChecksumSHA256
	if _, err := verifyBackupArtifact(wrong, artifact); ErrorCode(err) != "INVALID_BACKUP_ARTIFACT" {
		t.Fatalf("mismatched manifest error = %v", err)
	}
}

func TestReleaseKeysAreStableAndScopeBound(t *testing.T) {
	first := releaseIdempotencyKey(1, 101, "publish-request")
	if first != releaseIdempotencyKey(1, 101, "publish-request") {
		t.Fatal("release idempotency key is not stable")
	}
	if first == releaseIdempotencyKey(2, 202, "publish-request") {
		t.Fatal("release idempotency key is not tenant/site scoped")
	}
}

func TestTaskRetryReturnsIdempotentlyForExistingRequest(t *testing.T) {
	service, db := newOperationsStabilityTestService(t)
	if err := db.Exec(`INSERT INTO game_system_retry_records
		(id,source_type,source_id,idempotency_key,status,created_at,updated_at)
		VALUES (1,'TASK_EXECUTION',42,'task-execution:42','SUCCESS',?,?)`, time.Now(), time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	if err := service.RetryTaskExecution(42, "operator retry", Actor{ID: 9, Platform: true}); err != nil {
		t.Fatalf("duplicate retry was not idempotent: %v", err)
	}
}

func testBackupArtifact(t *testing.T, manifest backupManifest) []byte {
	t.Helper()
	contents, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := writeTarFile(tarWriter, "manifest.json", contents); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func newOperationsStabilityTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, statement := range []string{
		`CREATE TABLE game_tenant_maintenance (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, scope_type TEXT, enabled INTEGER, start_at DATETIME, end_at DATETIME, reason TEXT, title TEXT, message_zh TEXT, message_en TEXT, estimated_recovery_at DATETIME, support_url TEXT, support_email TEXT, support_chat TEXT, allowlisted_ips TEXT, allow_platform_bypass INTEGER, allow_tenant_bypass INTEGER, created_by INTEGER, updated_by INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_saas_sites (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_code TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_deposit_orders (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, user_id INTEGER, order_no TEXT, channel_order_no TEXT, amount NUMERIC, currency TEXT, status INTEGER, channel_code TEXT, created_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_support_cases (id INTEGER PRIMARY KEY, case_no TEXT, tenant_id INTEGER, site_id INTEGER, player_id INTEGER, category TEXT, priority TEXT, status TEXT, title TEXT, description TEXT, assigned_admin_id INTEGER, created_by INTEGER, resolution TEXT, related_order_ids TEXT, related_transaction_ids TEXT, idempotency_key TEXT, closed_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_support_case_notes (id INTEGER PRIMARY KEY, tenant_id INTEGER, site_id INTEGER, case_id INTEGER, author_id INTEGER, author_username TEXT, note TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_system_alerts (id INTEGER PRIMARY KEY AUTOINCREMENT, fingerprint TEXT NOT NULL UNIQUE, tenant_id INTEGER, site_id INTEGER, severity TEXT, alert_type TEXT, source TEXT, status TEXT, occurrence_count INTEGER, summary TEXT, details TEXT, first_triggered_at DATETIME, last_triggered_at DATETIME, acknowledged_by INTEGER, acknowledged_at DATETIME, resolved_by INTEGER, resolved_at DATETIME, internal_note TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_system_retry_records (id INTEGER PRIMARY KEY AUTOINCREMENT, source_type TEXT, source_id INTEGER, tenant_id INTEGER, site_id INTEGER, idempotency_key TEXT UNIQUE, status TEXT, attempt_count INTEGER, max_attempts INTEGER, next_attempt_at DATETIME, last_error TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE event_outbox (id INTEGER PRIMARY KEY AUTOINCREMENT, event_id TEXT, event_type TEXT, tenant_id INTEGER, site_id INTEGER, aggregate_type TEXT, aggregate_id TEXT, status TEXT, attempts INTEGER, last_error TEXT, occurred_at DATETIME, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE game_admin_operation_logs (id INTEGER PRIMARY KEY AUTOINCREMENT, operator_id INTEGER, tenant_id INTEGER, site_id INTEGER, operator_username TEXT, permission TEXT, module TEXT, action TEXT, result TEXT, target_type TEXT, target_id INTEGER, target_user_id INTEGER, before_state TEXT, after_state TEXT, reason TEXT, ip_address TEXT, user_agent TEXT, correlation_id TEXT, currency TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	return New(db), db
}
