package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	HeaderOutboxID      = "x-outbox-id"
	HeaderRetryCount    = "x-retry-count"
	HeaderOriginalQueue = "x-original-queue"
	HeaderLastError     = "x-last-error"
)

// RelayPending claims durable outbox rows and publishes them through the
// configured adapter. The row is marked published only after publisher confirm.
func (s *Service) RelayPending(ctx context.Context, publisher EventPublisher, eventTypes []string, limit int) (DispatchSummary, error) {
	if publisher == nil {
		return DispatchSummary{}, fmt.Errorf("event publisher is required")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if err := s.recoverExpiredProcessing(ctx); err != nil {
		return DispatchSummary{}, err
	}
	rows, err := s.claimDueByEventTypes(ctx, limit, eventTypes)
	if err != nil {
		return DispatchSummary{}, err
	}
	summary := DispatchSummary{Scanned: len(rows)}
	for _, row := range rows {
		record, parseErr := recordFromRowStrict(row)
		if parseErr != nil {
			status, markErr := s.markFailed(ctx, row.ID, row.Attempts, parseErr.Error())
			if markErr != nil {
				return summary, markErr
			}
			if status == StatusDead {
				summary.Dead++
			} else {
				summary.Failed++
			}
			continue
		}
		body, marshalErr := json.Marshal(record.Envelope)
		if marshalErr != nil {
			_, _ = s.markFailed(ctx, row.ID, row.Attempts, marshalErr.Error())
			summary.Failed++
			continue
		}
		message := PublishMessage{
			EventID: record.EventID, OutboxID: record.ID, RoutingKey: RoutingKey(record.EventType), Body: body,
			Headers: map[string]any{HeaderOutboxID: record.ID, HeaderRetryCount: 0}, OccurredAt: record.OccurredAt,
			DeliveryMode: 2, ContentType: "application/json", CorrelationID: record.CorrelationID, MessageType: record.EventType,
		}
		if publishErr := publisher.Publish(ctx, message); publishErr != nil {
			status, markErr := s.markFailed(ctx, row.ID, row.Attempts, publishErr.Error())
			if markErr != nil {
				return summary, markErr
			}
			if status == StatusDead {
				summary.Dead++
			} else {
				summary.Failed++
			}
			continue
		}
		if err := s.markPublished(ctx, row.ID); err != nil {
			return summary, err
		}
		summary.Published++
	}
	return summary, nil
}

func (s *Service) claimDueByEventTypes(ctx context.Context, limit int, eventTypes []string) ([]eventOutbox, error) {
	normalized := make([]string, 0, len(eventTypes))
	seen := map[string]struct{}{}
	for _, eventType := range eventTypes {
		eventType = NormalizeType(eventType)
		if eventType == "" {
			continue
		}
		if _, ok := seen[eventType]; ok {
			continue
		}
		seen[eventType] = struct{}{}
		normalized = append(normalized, eventType)
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	sort.Strings(normalized)
	now := s.now()
	claimed := []eventOutbox{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows := []eventOutbox{}
		query := lockForClaim(dueQuery(tx, now).Where("event_type IN ?", normalized))
		if err := query.Order("next_attempt_at ASC,id ASC").Limit(limit).Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if row.Attempts >= s.maxAttempts {
				if err := s.markDeadTx(tx, row.ID, row.Attempts, "event exceeded maximum publish attempts"); err != nil {
					return err
				}
				continue
			}
			result := tx.Model(&eventOutbox{}).
				Where("id=? AND status IN ? AND next_attempt_at<=? AND deleted_at IS NULL", row.ID, []string{StatusPending, StatusFailed}, now).
				Updates(map[string]any{
					"status": StatusProcessing, "attempts": gorm.Expr("attempts+1"), "locked_by": s.workerID,
					"locked_at": now, "last_error": "", "updated_at": now,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			row.Status = StatusProcessing
			row.Attempts++
			row.LockedBy = s.workerID
			row.LockedAt = &now
			claimed = append(claimed, row)
		}
		return nil
	})
	return claimed, err
}

// Dispatch invokes one stable consumer while retaining database-backed delivery
// idempotency. Consumers must still make their own business write idempotent.
func (s *Service) Dispatch(ctx context.Context, message DeliveryMessage, registration ConsumerRegistration) error {
	var envelope EventEnvelope
	if err := json.Unmarshal(message.Body, &envelope); err != nil {
		return ErrEventPayloadInvalid
	}
	if envelope.EventVersion <= 0 {
		var legacy struct {
			EventVersion int `json:"event_version"`
		}
		_ = json.Unmarshal(message.Body, &legacy)
		envelope.EventVersion = legacy.EventVersion
	}
	if envelope.EventVersion <= 0 {
		envelope.EventVersion = DefaultEventVersion
	}
	if strings.TrimSpace(envelope.EventID) == "" || strings.TrimSpace(envelope.EventType) == "" {
		return ErrEventPayloadInvalid
	}
	if envelope.Payload == nil {
		envelope.Payload = map[string]any{}
	}
	if envelope.Metadata == nil {
		envelope.Metadata = map[string]any{}
	}
	if !registrationAccepts(registration, envelope) {
		return nil
	}
	record := Record{
		ID: headerUint64(message.Headers, HeaderOutboxID), EventID: envelope.EventID, TenantID: envelope.TenantID, SiteID: envelope.SiteID,
		EventType: envelope.EventType, EventVersion: envelope.EventVersion, AggregateType: envelope.AggregateType, AggregateID: envelope.AggregateID,
		Source: envelope.Producer, TraceID: envelope.TraceID, RequestID: envelope.RequestID, CorrelationID: envelope.CorrelationID,
		CausationID: envelope.CausationID, Producer: envelope.Producer, OccurredAt: envelope.OccurredAt,
		TargetConsumers: envelope.TargetConsumers, Envelope: envelope, Payload: envelope.Payload, Metadata: envelope.Metadata, Status: StatusPublished,
	}
	if record.OccurredAt.IsZero() {
		record.OccurredAt = time.Now()
	}
	item := handlerItem{consumer: registration.Name, handler: registration.Handler}
	if err := s.ensureDeliveries(ctx, record, []handlerItem{item}); err != nil {
		return err
	}
	return s.dispatchToConsumer(ctx, record, registration.Name, registration.Handler)
}

func registrationAccepts(registration ConsumerRegistration, envelope EventEnvelope) bool {
	if len(registration.EventTypes) > 0 {
		matched := false
		for _, eventType := range registration.EventTypes {
			if NormalizeType(eventType) == envelope.EventType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(envelope.TargetConsumers) == 0 {
		return true
	}
	for _, consumer := range envelope.TargetConsumers {
		if consumer == registration.Name {
			return true
		}
	}
	return false
}

func headerUint64(headers map[string]any, key string) uint64 {
	value := headers[key]
	switch typed := value.(type) {
	case uint64:
		return typed
	case uint:
		return uint64(typed)
	case int:
		if typed > 0 {
			return uint64(typed)
		}
	case int32:
		if typed > 0 {
			return uint64(typed)
		}
	case int64:
		if typed > 0 {
			return uint64(typed)
		}
	case float64:
		if typed > 0 {
			return uint64(typed)
		}
	}
	return 0
}

// RoutingKey keeps the public event protocol independent from Go constants.
func RoutingKey(eventType string) string {
	switch NormalizeType(eventType) {
	case EventLedgerTransactionPosted:
		return "ledger.transaction.posted"
	case EventWalletBalanceChanged:
		return "wallet.balance.changed"
	case EventPlayerDepositSucceeded:
		return "payment.deposit.success"
	case EventPlayerWithdrawApproved:
		return "payment.withdraw.success"
	case EventProviderCallbackReceived:
		return "provider.callback.received"
	case EventGameBetPlaced:
		return "provider.bet"
	case EventGameBetSettled:
		return "provider.win"
	case EventTenantDepositCredited:
		return "tenant.deposit.credited"
	case EventTenantCreditChanged:
		return "tenant.credit.changed"
	case EventTenantFundsFrozen:
		return "tenant.funds.frozen"
	case EventTenantFundsUnfrozen:
		return "tenant.funds.unfrozen"
	case EventTenantLedgerReversed:
		return "ledger.transaction.reversed"
	case EventWhiteLabelConfigPublished:
		return "tenant.config.published"
	case EventTenantProvisioned:
		return "tenant.provisioned"
	case EventSupportExportRequested:
		return "task.support.export.requested"
	case EventReportExportRequested:
		return "task.report.export.requested"
	case EventProviderSyncRequested:
		return "task.provider.sync.requested"
	case EventBackupRequested:
		return "task.backup.requested"
	case EventSystemTaskRequested:
		return "task.system.execution.requested"
	case EventNotificationSendRequested:
		return "task.notification.send.requested"
	default:
		value := strings.TrimSuffix(NormalizeType(eventType), ".v1")
		return strings.ReplaceAll(value, "_", ".")
	}
}
