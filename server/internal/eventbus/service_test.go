package eventbus

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"casino/internal/tenancy"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestPublishIsIdempotent(t *testing.T) {
	service, db := newEventBusTestService(t)

	first, err := service.Publish(context.Background(), testEvent("same-key", 7))
	require.NoError(t, err)
	second, err := service.Publish(context.Background(), testEvent("same-key", 7))
	require.NoError(t, err)

	require.Equal(t, first.ID, second.ID)
	require.True(t, second.Idempotent)
	var count int64
	require.NoError(t, db.Model(&eventOutbox{}).Where("idempotency_key=?", "same-key").Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.Equal(t, DefaultEventVersion, first.EventVersion)
	require.Equal(t, first.EventID, first.Envelope.EventID)
	require.Equal(t, first.EventType, first.Envelope.EventType)
	require.Equal(t, uint64(7), first.Envelope.TenantID)
	require.Equal(t, uint64(3), first.Envelope.SiteID)
	require.Equal(t, "trace-same-key", first.TraceID)
	require.Equal(t, "corr-same-key", first.CorrelationID)
}

func TestPublishTxRollsBackWithBusinessTransaction(t *testing.T) {
	service, db := newEventBusTestService(t)
	err := db.Transaction(func(tx *gorm.DB) error {
		_, publishErr := service.PublishTx(tx, testEvent("rollback-key", 7))
		require.NoError(t, publishErr)
		return errors.New("business failed")
	})
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&eventOutbox{}).Where("idempotency_key=?", "rollback-key").Count(&count).Error)
	require.Equal(t, int64(0), count)
}

func TestRegisterRequiresStableVersionedConsumerName(t *testing.T) {
	service, _ := newEventBusTestService(t)
	require.ErrorIs(t, service.Register("dashboard", func(ctx Context) error { return nil }), ErrConsumerNameInvalid)
	require.NoError(t, service.Register(ConsumerDashboardDailyStats, func(ctx Context) error { return nil }))
}

func TestDispatchRetriesFailedConsumerWithoutRepeatingSuccess(t *testing.T) {
	service, db := newEventBusTestService(t)
	current := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return current }
	var successCalls, flakyCalls int
	require.NoError(t, service.Register(ConsumerDashboardDailyStats, func(ctx Context) error {
		successCalls++
		return nil
	}))
	require.NoError(t, service.Register(ConsumerRiskRules, func(ctx Context) error {
		flakyCalls++
		return errors.New("risk engine down")
	}))
	require.Len(t, service.snapshotHandlers(), 2)
	_, err := service.Publish(context.Background(), testEvent("retry-key", 7))
	require.NoError(t, err)

	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Failed: 1}, summary)
	require.Equal(t, 1, successCalls)
	require.Equal(t, 1, flakyCalls)
	require.Equal(t, StatusFailed, outboxByKey(t, db, "retry-key").Status)
	var deliveries int64
	require.NoError(t, db.Model(&eventDelivery{}).Where("event_id=?", outboxByKey(t, db, "retry-key").EventID).Count(&deliveries).Error)
	require.Equal(t, int64(2), deliveries)

	require.NoError(t, service.Register(ConsumerRiskRules, func(ctx Context) error {
		flakyCalls++
		return nil
	}))
	current = current.Add(time.Minute)
	summary, err = service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Published: 1}, summary)
	require.Equal(t, 1, successCalls)
	require.Equal(t, 2, flakyCalls)
	require.Equal(t, StatusPublished, outboxByKey(t, db, "retry-key").Status)
}

func TestConcurrentDispatchersClaimEventOnce(t *testing.T) {
	serviceA, db := newEventBusTestService(t)
	serviceB := NewWithOptions(db, Options{WorkerID: "worker-b"})
	var calls atomic.Int64
	handler := func(ctx Context) error {
		calls.Add(1)
		time.Sleep(5 * time.Millisecond)
		return nil
	}
	require.NoError(t, serviceA.Register(ConsumerDashboardDailyStats, handler))
	require.NoError(t, serviceB.Register(ConsumerDashboardDailyStats, handler))
	_, err := serviceA.Publish(context.Background(), testEvent("claim-once", 7))
	require.NoError(t, err)

	var wg sync.WaitGroup
	summaries := make([]DispatchSummary, 2)
	errs := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		summaries[0], errs[0] = serviceA.DispatchPending(context.Background(), 10)
	}()
	go func() {
		defer wg.Done()
		summaries[1], errs[1] = serviceB.DispatchPending(context.Background(), 10)
	}()
	wg.Wait()

	require.NoError(t, errs[0])
	require.NoError(t, errs[1])
	require.Equal(t, int64(1), calls.Load())
	require.Equal(t, StatusPublished, outboxByKey(t, db, "claim-once").Status)
	require.Equal(t, 1, summaries[0].Published+summaries[1].Published)
}

func TestProcessingTimeoutIsRecoveredAndDispatched(t *testing.T) {
	service, db := newEventBusTestServiceWithOptions(t, Options{WorkerID: "worker-a", LeaseTimeout: time.Minute})
	current := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return current }
	_, err := service.Publish(context.Background(), testEvent("stuck", 7))
	require.NoError(t, err)
	oldLock := current.Add(-2 * time.Minute)
	require.NoError(t, db.Model(&eventOutbox{}).Where("idempotency_key=?", "stuck").Updates(map[string]any{
		"status": StatusProcessing, "locked_by": "dead-worker", "locked_at": oldLock, "attempts": 1,
	}).Error)
	var calls int
	require.NoError(t, service.Register(ConsumerDashboardDailyStats, func(ctx Context) error {
		calls++
		return nil
	}))

	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Published: 1}, summary)
	require.Equal(t, 1, calls)
	row := outboxByKey(t, db, "stuck")
	require.Equal(t, StatusPublished, row.Status)
	require.Empty(t, row.LockedBy)
}

func TestMaxAttemptsMovesEventToDead(t *testing.T) {
	service, db := newEventBusTestServiceWithOptions(t, Options{MaxAttempts: 2})
	current := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return current }
	require.NoError(t, service.Register(ConsumerRiskRules, func(ctx Context) error {
		return errors.New("still down")
	}))
	_, err := service.Publish(context.Background(), testEvent("dead-key", 7))
	require.NoError(t, err)

	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Failed: 1}, summary)
	current = current.Add(time.Minute)
	summary, err = service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Dead: 1}, summary)
	row := outboxByKey(t, db, "dead-key")
	require.Equal(t, StatusDead, row.Status)
	require.Equal(t, uint(2), row.Attempts)
}

func TestConsumerBusinessIdempotencySurvivesDeliveryCrash(t *testing.T) {
	service, db := newEventBusTestService(t)
	require.NoError(t, db.AutoMigrate(&testConsumerProjection{}))
	current := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return current }
	var calls int
	require.NoError(t, service.Register(ConsumerDashboardDailyStats, func(ctx Context) error {
		calls++
		row := testConsumerProjection{SourceEventID: ctx.Event.EventID, TenantID: ctx.Event.TenantID, SiteID: ctx.Event.SiteID, Value: ctx.Event.Payload["value"].(string)}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return err
		}
		if calls == 1 {
			return errors.New("crash after business write")
		}
		return nil
	}))
	_, err := service.Publish(context.Background(), testEvent("idempotent-consumer", 7))
	require.NoError(t, err)

	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Failed: 1}, summary)
	current = current.Add(time.Minute)
	summary, err = service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Published: 1}, summary)
	require.Equal(t, 2, calls)
	var count int64
	require.NoError(t, db.Model(&testConsumerProjection{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestInvalidPayloadMovesToDeadWhenAttemptsExceeded(t *testing.T) {
	service, db := newEventBusTestServiceWithOptions(t, Options{MaxAttempts: 1})
	require.NoError(t, service.Register(ConsumerDashboardDailyStats, func(ctx Context) error {
		t.Fatal("handler should not be called for invalid payload")
		return nil
	}))
	_, err := service.Publish(context.Background(), testEvent("bad-payload", 7))
	require.NoError(t, err)
	require.NoError(t, db.Model(&eventOutbox{}).Where("idempotency_key=?", "bad-payload").Update("payload", "{").Error)

	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Dead: 1}, summary)
	row := outboxByKey(t, db, "bad-payload")
	require.Equal(t, StatusDead, row.Status)
	require.Contains(t, row.LastError, ErrEventPayloadInvalid.Error())
}

func TestReplayDoesNotRepeatSuccessfulDeliveryUnlessForced(t *testing.T) {
	service, db := newEventBusTestService(t)
	current := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return current }
	var dashboardCalls, riskCalls int
	require.NoError(t, service.Register(ConsumerDashboardDailyStats, func(ctx Context) error {
		dashboardCalls++
		return nil
	}))
	require.NoError(t, service.Register(ConsumerRiskRules, func(ctx Context) error {
		riskCalls++
		return errors.New("risk down")
	}))
	record, err := service.Publish(context.Background(), testEvent("replay-key", 7))
	require.NoError(t, err)
	_, err = service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, dashboardCalls)
	require.Equal(t, 1, riskCalls)

	require.NoError(t, service.Register(ConsumerRiskRules, func(ctx Context) error {
		riskCalls++
		return nil
	}))
	_, err = service.ReplayEvent(context.Background(), record.ID, tenancy.Scope{Platform: true}, false)
	require.NoError(t, err)
	current = current.Add(time.Minute)
	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Published: 1}, summary)
	require.Equal(t, 1, dashboardCalls)
	require.Equal(t, 2, riskCalls)

	_, err = service.ReplayEvent(context.Background(), record.ID, tenancy.Scope{Platform: true}, true)
	require.NoError(t, err)
	current = current.Add(time.Minute)
	summary, err = service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Published: 1}, summary)
	require.Equal(t, 2, dashboardCalls)
	require.Equal(t, 3, riskCalls)
	require.Equal(t, StatusPublished, outboxByKey(t, db, "replay-key").Status)
}

func TestReplayRangeAndConsumer(t *testing.T) {
	service, _ := newEventBusTestService(t)
	var calls int
	require.NoError(t, service.Register(ConsumerDashboardDailyStats, func(ctx Context) error {
		calls++
		return nil
	}))
	first, err := service.Publish(context.Background(), testEvent("range-a", 7))
	require.NoError(t, err)
	_, err = service.Publish(context.Background(), testEvent("range-b", 7))
	require.NoError(t, err)
	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 2, Published: 2}, summary)
	require.Equal(t, 2, calls)

	result, err := service.ReplayRange(context.Background(), tenancy.Scope{Platform: true}, ReplayRangeOptions{FromID: first.ID, ToID: first.ID})
	require.NoError(t, err)
	require.Equal(t, ReplayResult{Replayed: 1}, result)
	summary, err = service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Published: 1}, summary)
	require.Equal(t, 2, calls)

	result, err = service.ReplayConsumer(context.Background(), tenancy.Scope{Platform: true}, ReplayConsumerOptions{Consumer: ConsumerDashboardDailyStats, Force: true})
	require.NoError(t, err)
	require.Equal(t, ReplayResult{Replayed: 2}, result)
	summary, err = service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 2, Published: 2}, summary)
	require.Equal(t, 4, calls)
}

func TestDispatchWithoutConsumersDoesNotPublishOrIncrementAttempts(t *testing.T) {
	service, db := newEventBusTestService(t)
	_, err := service.Publish(context.Background(), testEvent("no-consumer", 7))
	require.NoError(t, err)

	summary, err := service.DispatchPending(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, DispatchSummary{Scanned: 1, Skipped: 1}, summary)
	row := outboxByKey(t, db, "no-consumer")
	require.Equal(t, StatusPending, row.Status)
	require.Equal(t, uint(0), row.Attempts)
}

func TestListIsTenantScoped(t *testing.T) {
	service, _ := newEventBusTestService(t)
	_, err := service.Publish(context.Background(), testEvent("tenant-7", 7))
	require.NoError(t, err)
	_, err = service.Publish(context.Background(), testEvent("tenant-8", 8))
	require.NoError(t, err)

	rows, total, err := service.List(context.Background(), tenancy.Scope{TenantID: 7}, nil, 1, 20)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	require.Equal(t, uint64(7), rows[0].TenantID)
	require.Equal(t, "tenant-7", rows[0].IdempotencyKey)
}

func TestMigrationDefinesOutboxAndDeliveryConstraints(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "migrations", "20260719_event_bus.sql"))
	require.NoError(t, err)
	migration := string(data)
	hardeningData, err := os.ReadFile(filepath.Join("..", "..", "migrations", "20260719_event_bus_hardening.sql"))
	require.NoError(t, err)
	hardening := string(hardeningData)

	require.Contains(t, migration, "CREATE TABLE IF NOT EXISTS event_outbox")
	require.Contains(t, migration, "CREATE TABLE IF NOT EXISTS event_deliveries")
	require.Contains(t, migration, "uk_event_outbox_idempotency (idempotency_key)")
	require.Contains(t, migration, "idx_event_outbox_scope_time (tenant_id, site_id, created_at, id)")
	require.Contains(t, hardening, "consumer_name")
	require.Contains(t, hardening, "uk_event_delivery_consumer','(event_id, consumer_name)'")
	require.Contains(t, hardening, "status IN (''PENDING'',''PROCESSING'',''PUBLISHED'',''FAILED'',''DEAD'')")
	require.Contains(t, hardening, "system:event:replay")
}

func newEventBusTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	return newEventBusTestServiceWithOptions(t, Options{})
}

func newEventBusTestServiceWithOptions(t *testing.T, options Options) (*Service, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, Migrate(db))
	return NewWithOptions(db, options), db
}

func testEvent(key string, tenantID uint64) Event {
	return Event{
		Scope: tenancy.Scope{Platform: true}, TenantID: tenantID, SiteID: 3,
		EventType: EventWalletBalanceChanged, AggregateType: "wallet", AggregateID: "88",
		Source: "test", IdempotencyKey: key, TraceID: "trace-" + key, CorrelationID: "corr-" + key, Payload: map[string]any{"value": key},
	}
}

func outboxByKey(t *testing.T, db *gorm.DB, key string) eventOutbox {
	t.Helper()
	var row eventOutbox
	require.NoError(t, db.Where("idempotency_key=?", key).Take(&row).Error)
	return row
}

type testConsumerProjection struct {
	ID            uint64 `gorm:"column:id;primaryKey"`
	SourceEventID string `gorm:"column:source_event_id;uniqueIndex"`
	TenantID      uint64 `gorm:"column:tenant_id"`
	SiteID        uint64 `gorm:"column:site_id"`
	Value         string `gorm:"column:value"`
}

func (testConsumerProjection) TableName() string { return "test_consumer_projections" }
