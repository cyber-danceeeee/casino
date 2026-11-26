package eventconsumers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"casino/internal/eventbus"
	"casino/internal/onboardingemail"
	"casino/internal/tenancy"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type recordingEmailSender struct {
	messages []onboardingemail.Message
}

func (s *recordingEmailSender) Send(_ context.Context, message onboardingemail.Message) error {
	s.messages = append(s.messages, message)
	return nil
}

func TestTenantOnboardingEmailConsumerIsIdempotent(t *testing.T) {
	t.Setenv("CASINO_NOTIFICATION_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32)))
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE game_tenant_onboarding_email_jobs (
		id INTEGER PRIMARY KEY,event_id TEXT,tenant_id INTEGER,site_id INTEGER,recipient_email TEXT,template_code TEXT,
		encrypted_payload TEXT,status TEXT,attempt_count INTEGER,last_error TEXT,idempotency_key TEXT,sent_at DATETIME,
		created_at DATETIME,updated_at DATETIME,deleted_at DATETIME
	)`).Error)
	payload, err := onboardingemail.EncryptPayload(onboardingemail.ProvisionedPayload{
		BrandName: "Demo", TenantCode: "demo", AdminUsername: "owner", TemporaryPassword: "StrongPassword9!", Language: "en",
	})
	require.NoError(t, err)
	eventID := "EVT-ONBOARDING"
	job := onboardingemail.Job{
		ID: 1, EventID: &eventID, TenantID: 7, SiteID: 8, RecipientEmail: "owner@example.com",
		TemplateCode: onboardingemail.TemplateTenantProvisioned, EncryptedPayload: payload,
		Status: "PENDING", IdempotencyKey: "tenant-provision-email:7:1", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&job).Error)
	sender := &recordingEmailSender{}
	consumer := NewWithEmailSender(db, sender)
	event := eventbus.Context{Event: eventbus.Record{
		EventID: eventID, TenantID: 7, SiteID: 8, EventType: eventbus.EventTenantProvisioned,
		Payload: map[string]any{"emailJobID": job.ID, onboardingemail.EventEncryptedPayloadKey: payload},
	}}
	require.NoError(t, consumer.HandleTenantOnboardingEmail(event))
	require.NoError(t, consumer.HandleTenantOnboardingEmail(event))
	require.Len(t, sender.messages, 1)
	require.Contains(t, sender.messages[0].Body, "StrongPassword9!")
	require.NoError(t, db.First(&job, job.ID).Error)
	require.Equal(t, "SENT", job.Status)
}

func TestTenantOnboardingEmailConsumerRejectsMismatchedOutboxCredentials(t *testing.T) {
	t.Setenv("CASINO_NOTIFICATION_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{8}, 32)))
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE game_tenant_onboarding_email_jobs (
		id INTEGER PRIMARY KEY,event_id TEXT,tenant_id INTEGER,site_id INTEGER,recipient_email TEXT,template_code TEXT,
		encrypted_payload TEXT,status TEXT,attempt_count INTEGER,last_error TEXT,idempotency_key TEXT,sent_at DATETIME,
		created_at DATETIME,updated_at DATETIME,deleted_at DATETIME
	)`).Error)
	jobPayload, err := onboardingemail.EncryptPayload(onboardingemail.ProvisionedPayload{
		BrandName: "Demo", TenantCode: "demo", AdminUsername: "owner", TemporaryPassword: "OriginalPassword9!",
	})
	require.NoError(t, err)
	eventPayload, err := onboardingemail.EncryptPayload(onboardingemail.ProvisionedPayload{
		BrandName: "Demo", TenantCode: "demo", AdminUsername: "owner", TemporaryPassword: "DifferentPassword9!",
	})
	require.NoError(t, err)
	eventID := "EVT-MISMATCH"
	job := onboardingemail.Job{
		ID: 2, EventID: &eventID, TenantID: 7, SiteID: 8, RecipientEmail: "owner@example.com",
		TemplateCode: onboardingemail.TemplateTenantProvisioned, EncryptedPayload: jobPayload,
		Status: "PENDING", IdempotencyKey: "tenant-provision-email:7:2", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&job).Error)
	sender := &recordingEmailSender{}
	err = NewWithEmailSender(db, sender).HandleTenantOnboardingEmail(eventbus.Context{Event: eventbus.Record{
		EventID: eventID, TenantID: 7, SiteID: 8, EventType: eventbus.EventTenantProvisioned,
		Payload: map[string]any{"emailJobID": job.ID, onboardingemail.EventEncryptedPayloadKey: eventPayload},
	}})
	require.EqualError(t, err, "tenant onboarding email payload does not match job")
	require.Empty(t, sender.messages)
}

func TestLedgerAuditConsumerIsIdempotentAcrossReplay(t *testing.T) {
	bus, db := newConsumerTestBus(t)
	event, err := eventbus.New(db).Publish(context.Background(), ledgerPostedEvent("audit-once"))
	require.NoError(t, err)

	summary, err := bus.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, eventbus.DispatchSummary{Scanned: 1, Published: 1}, summary)
	assertAuditCount(t, db, 1)

	_, err = bus.ReplayEvent(context.Background(), event.ID, tenancy.Scope{Platform: true}, true)
	require.NoError(t, err)
	summary, err = bus.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, eventbus.DispatchSummary{Scanned: 1, Published: 1}, summary)
	assertAuditCount(t, db, 1)
}

func TestDashboardFactConsumerWritesIncrementFact(t *testing.T) {
	bus, db := newConsumerTestBus(t)
	_, err := eventbus.New(db).Publish(context.Background(), tenantLedgerBusinessEvent("dashboard-fact", eventbus.EventTenantDepositCredited))
	require.NoError(t, err)

	summary, err := bus.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, eventbus.DispatchSummary{Scanned: 1, Published: 1}, summary)
	var fact DashboardEventFact
	require.NoError(t, db.Where("event_id<>''").Take(&fact).Error)
	require.Equal(t, eventbus.ConsumerDashboardDailyStats, fact.ConsumerName)
	require.Equal(t, "TENANT_DEPOSIT_CREDITED", fact.MetricType)
	require.Equal(t, "PKR", fact.Currency)
	require.True(t, decimal.RequireFromString(fact.Amount).Equal(decimal.NewFromInt(100)))
	assertAuditCount(t, db, 0)
}

func TestTargetConsumersPreventUnexpectedHistoricalConsumer(t *testing.T) {
	bus, db := newConsumerTestBus(t)
	_, err := eventbus.New(db).Publish(context.Background(), ledgerPostedEvent("target-audit-only"))
	require.NoError(t, err)

	summary, err := bus.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, eventbus.DispatchSummary{Scanned: 1, Published: 1}, summary)
	var deliveryCount int64
	require.NoError(t, db.Table("event_deliveries").Where("consumer_name=?", eventbus.ConsumerDashboardDailyStats).Count(&deliveryCount).Error)
	require.Equal(t, int64(0), deliveryCount)
	assertAuditCount(t, db, 1)
}

func TestConsumerBusinessIdempotencyHandlesDeliveryRetry(t *testing.T) {
	db := newConsumerTestDB(t)
	consumers := New(db)
	var calls int
	bus := eventbus.NewWithOptions(db, eventbus.Options{})
	require.NoError(t, bus.RegisterFor(eventbus.ConsumerLedgerAudit, []string{eventbus.EventLedgerTransactionPosted}, func(ctx eventbus.Context) error {
		calls++
		if err := consumers.HandleLedgerAudit(ctx); err != nil {
			return err
		}
		if calls == 1 {
			return errors.New("crash after audit write")
		}
		return nil
	}))
	_, err := eventbus.New(db).Publish(context.Background(), ledgerPostedEvent("delivery-crash"))
	require.NoError(t, err)

	summary, err := bus.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, eventbus.DispatchSummary{Scanned: 1, Failed: 1}, summary)
	time.Sleep(time.Millisecond)
	summary, err = bus.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, eventbus.DispatchSummary{Scanned: 0}, summary)
	_, err = bus.ReplayConsumer(context.Background(), tenancy.Scope{Platform: true}, eventbus.ReplayConsumerOptions{Consumer: eventbus.ConsumerLedgerAudit})
	require.NoError(t, err)
	summary, err = bus.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, eventbus.DispatchSummary{Scanned: 1, Published: 1}, summary)
	assertAuditCount(t, db, 1)
}

func TestWorkerProcessOnceDispatchesRegisteredConsumers(t *testing.T) {
	db := newConsumerTestDB(t)
	_, err := eventbus.New(db).Publish(context.Background(), ledgerPostedEvent("worker"))
	require.NoError(t, err)

	worker := NewWorker(db, WorkerOptions{WorkerID: "test-worker", BatchSize: 10})
	worker.ProcessOnce(context.Background())
	assertAuditCount(t, db, 1)
}

func TestMigrationDefinesConsumerTables(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "migrations", "20260719_event_consumers.sql"))
	require.NoError(t, err)
	migration := string(data)
	require.Contains(t, migration, "CREATE TABLE IF NOT EXISTS audit_event_records")
	require.Contains(t, migration, "CREATE TABLE IF NOT EXISTS dashboard_event_facts")
	require.Contains(t, migration, "uk_audit_event_consumer (event_id, consumer_name)")
	require.Contains(t, migration, "uk_dashboard_fact_event_metric (event_id, consumer_name, metric_type)")
	require.Contains(t, migration, "target_consumers")
}

func newConsumerTestBus(t *testing.T) (*eventbus.Service, *gorm.DB) {
	t.Helper()
	db := newConsumerTestDB(t)
	bus, err := NewService(db, eventbus.Options{})
	require.NoError(t, err)
	return bus, db
}

func newConsumerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, eventbus.Migrate(db))
	require.NoError(t, EnsureSchema(db))
	return db
}

func ledgerPostedEvent(key string) eventbus.Event {
	return eventbus.Event{
		Scope: tenancy.Scope{Platform: true}, TenantID: 7, SiteID: 3,
		EventType: eventbus.EventLedgerTransactionPosted, AggregateType: "ledger_transaction", AggregateID: "99",
		Source: "test", IdempotencyKey: "ledger:" + key, TraceID: "trace-" + key, CorrelationID: "corr-" + key,
		TargetConsumers: []string{eventbus.ConsumerLedgerAudit},
		Payload: map[string]any{
			"transactionID": uint64(99), "businessType": "TENANT_DEPOSIT", "currency": "PKR", "debitAmount": "100",
		},
		Metadata: map[string]any{"operatorID": uint64(11)},
	}
}

func tenantLedgerBusinessEvent(key, eventType string) eventbus.Event {
	event := ledgerPostedEvent(key)
	event.EventType = eventType
	event.IdempotencyKey = "ledger:" + key + ":" + eventType
	event.TargetConsumers = []string{eventbus.ConsumerDashboardDailyStats}
	return event
}

func assertAuditCount(t *testing.T, db *gorm.DB, expected int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&AuditEventRecord{}).Count(&count).Error)
	require.Equal(t, expected, count)
}
