package eventconsumers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"casino/internal/eventbus"
	"casino/internal/gameadmin"
	"casino/internal/notificationpush"
	"casino/internal/promotionengine"
	"casino/internal/reportadmin"
	"casino/internal/systemadmin"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func EnsureSchema(db *gorm.DB) error {
	return db.AutoMigrate(&AuditEventRecord{})
}

func NewService(db *gorm.DB, options eventbus.Options) (*eventbus.Service, error) {
	service := eventbus.NewWithOptions(db, options)
	for _, registration := range registrations(db, false) {
		if err := service.RegisterFor(registration.Name, registration.EventTypes, registration.Handler); err != nil {
			return nil, err
		}
	}
	return service, nil
}

func Registrations(db *gorm.DB, pushWorkers ...*notificationpush.Worker) []eventbus.ConsumerRegistration {
	return registrations(db, true, pushWorkers...)
}

func registrations(db *gorm.DB, includeAsync bool, pushWorkers ...*notificationpush.Worker) []eventbus.ConsumerRegistration {
	consumers := New(db)
	registrations := []eventbus.ConsumerRegistration{
		{
			Name: eventbus.ConsumerEventJournal, Queue: eventbus.QueueAudit,
			EventTypes: []string{
				eventbus.EventWalletBalanceChanged, eventbus.EventPlayerWithdrawApproved,
				eventbus.EventGameBetPlaced, eventbus.EventGameBetSettled, eventbus.EventSettlementCompleted,
				eventbus.EventProviderCallbackReceived,
			},
			RoutingKeys: []string{
				"provider.bet", "provider.win", "settlement.completed", "provider.callback.received",
			},
			Handler: consumers.HandleEventJournal,
		},
		{
			Name: eventbus.ConsumerPromotionDeposit, Queue: eventbus.QueuePromotion,
			EventTypes: []string{eventbus.EventPlayerDepositSucceeded}, RoutingKeys: []string{"payment.deposit.success"},
			Handler: promotionengine.New(db).HandleDepositEvent,
		},
	}
	if includeAsync {
		registrations = append(registrations,
			eventbus.ConsumerRegistration{
				Name: eventbus.ConsumerSupportExport, Queue: eventbus.QueueSystemJobs,
				EventTypes: []string{eventbus.EventSupportExportRequested}, RoutingKeys: []string{"task.support.export.requested"},
				Handler: systemadmin.New(db).HandleSupportExportEvent,
			},
			eventbus.ConsumerRegistration{
				Name: eventbus.ConsumerReportExport, Queue: eventbus.QueueSystemJobs,
				EventTypes: []string{eventbus.EventReportExportRequested}, RoutingKeys: []string{"task.report.export.requested"},
				Handler: reportadmin.New(db).Export.HandleEvent,
			},
			eventbus.ConsumerRegistration{
				Name: eventbus.ConsumerProviderSync, Queue: eventbus.QueueProvider,
				EventTypes: []string{eventbus.EventProviderSyncRequested}, RoutingKeys: []string{"task.provider.sync.requested"},
				Handler: gameadmin.New(db).HandleProviderSyncEvent,
			},
			eventbus.ConsumerRegistration{
				Name: eventbus.ConsumerBackup, Queue: eventbus.QueueSystemJobs,
				EventTypes: []string{eventbus.EventBackupRequested}, RoutingKeys: []string{"task.backup.requested"},
				Handler: systemadmin.New(db).HandleBackupEvent,
			},
			eventbus.ConsumerRegistration{
				Name: eventbus.ConsumerSystemTask, Queue: eventbus.QueueSystemJobs,
				EventTypes: []string{eventbus.EventSystemTaskRequested}, RoutingKeys: []string{"task.system.execution.requested"},
				Handler: systemadmin.New(db).HandleSystemTaskEvent,
			},
		)
	}
	if len(pushWorkers) > 0 && pushWorkers[0] != nil {
		registrations = append(registrations, eventbus.ConsumerRegistration{
			Name: eventbus.ConsumerNotificationPush, Queue: eventbus.QueueNotification,
			EventTypes: []string{eventbus.EventNotificationSendRequested}, RoutingKeys: []string{"task.notification.send.requested"},
			Handler: pushWorkers[0].HandleEvent,
		})
	}

	return registrations
}

type Consumers struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Consumers {
	return &Consumers{db: db}
}

func (c *Consumers) HandleEventJournal(ctx eventbus.Context) error {
	event := ctx.Event
	amount, err := eventAmount(event.Payload)
	if err != nil {
		return err
	}
	now := time.Now()
	businessType := stringPayload(event.Payload, "businessType")
	if businessType == "" {
		businessType = event.EventType
	}
	record := AuditEventRecord{
		EventID: event.EventID, EventType: event.EventType, EventVersion: event.EventVersion, ConsumerName: eventbus.ConsumerEventJournal,
		SiteID: event.SiteID, AggregateType: event.AggregateType, AggregateID: event.AggregateID,
		OperatorID: uintPayload(event.Metadata, "operatorID"), TransactionID: uintPayload(event.Payload, "transactionID"),
		BusinessType: businessType, Currency: stringPayload(event.Payload, "currency"), Amount: amount.StringFixed(18),
		TraceID: event.TraceID, CorrelationID: event.CorrelationID, CausationID: event.CausationID,
		PayloadSnapshot: marshal(event.Envelope), OccurredAt: event.OccurredAt, ConsumedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	return c.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error
}

func eventAmount(payload map[string]any) (decimal.Decimal, error) {
	for _, key := range []string{"amount", "debitAmount", "creditAmount"} {
		value := strings.TrimSpace(fmt.Sprint(payload[key]))
		if value == "" || value == "<nil>" {
			continue
		}
		amount, err := decimal.NewFromString(value)
		if err != nil {
			return decimal.Zero, err
		}
		return amount, nil
	}
	return decimal.Zero, nil
}

func stringPayload(payload map[string]any, key string) string {
	value := payload[key]
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func uintPayload(payload map[string]any, key string) uint64 {
	value := payload[key]
	switch typed := value.(type) {
	case uint64:
		return typed
	case uint:
		return uint64(typed)
	case int:
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
	case json.Number:
		parsed, _ := typed.Int64()
		if parsed > 0 {
			return uint64(parsed)
		}
	case string:
		parsed, err := decimal.NewFromString(strings.TrimSpace(typed))
		if err == nil && parsed.IsPositive() {
			return parsed.BigInt().Uint64()
		}
	}
	return 0
}

func marshal(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
