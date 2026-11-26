package eventconsumers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"casino/internal/config"
	"casino/internal/eventbus"
	"casino/internal/gameadmin"
	"casino/internal/notificationpush"
	"casino/internal/onboardingemail"
	"casino/internal/promotionengine"
	"casino/internal/reportadmin"
	"casino/internal/systemadmin"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var dashboardMetricByEvent = map[string]string{
	eventbus.EventTenantDepositCredited: "TENANT_DEPOSIT_CREDITED",
	eventbus.EventTenantCreditChanged:   "TENANT_CREDIT_CHANGED",
	eventbus.EventTenantFundsFrozen:     "TENANT_FUNDS_FROZEN",
	eventbus.EventTenantFundsUnfrozen:   "TENANT_FUNDS_UNFROZEN",
}

func EnsureSchema(db *gorm.DB) error {
	return db.AutoMigrate(&AuditEventRecord{}, &DashboardEventFact{}, &onboardingemail.Job{})
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
	emailConfig := config.EmailConfig()
	consumers := NewWithEmailSender(db, onboardingemail.NewSender(emailConfig))
	registrations := []eventbus.ConsumerRegistration{
		{
			Name: eventbus.ConsumerLedgerAudit, Queue: eventbus.QueueAudit,
			EventTypes: []string{eventbus.EventLedgerTransactionPosted}, RoutingKeys: []string{"ledger.transaction.posted"},
			Handler: consumers.HandleLedgerAudit,
		},
		{
			Name: eventbus.ConsumerEventJournal, Queue: eventbus.QueueAudit,
			EventTypes: []string{
				eventbus.EventTenantLedgerReversed, eventbus.EventWhiteLabelConfigPublished,
				eventbus.EventWalletBalanceChanged, eventbus.EventPlayerWithdrawApproved,
				eventbus.EventGameBetPlaced, eventbus.EventGameBetSettled, eventbus.EventSettlementCompleted,
				eventbus.EventProviderCallbackReceived,
			},
			RoutingKeys: []string{
				"ledger.transaction.reversed", "tenant.config.published", "wallet.balance.changed", "payment.withdraw.success",
				"provider.bet", "provider.win", "settlement.completed", "provider.callback.received",
			},
			Handler: consumers.HandleEventJournal,
		},
		{
			Name: eventbus.ConsumerDashboardDailyStats, Queue: eventbus.QueueDashboard,
			EventTypes:  []string{eventbus.EventTenantDepositCredited, eventbus.EventTenantCreditChanged, eventbus.EventTenantFundsFrozen, eventbus.EventTenantFundsUnfrozen},
			RoutingKeys: []string{"tenant.deposit.credited", "tenant.credit.changed", "tenant.funds.frozen", "tenant.funds.unfrozen"},
			Handler:     consumers.HandleDashboardFact,
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
	if onboardingemail.IsConfigured(emailConfig) {
		registrations = append(registrations, eventbus.ConsumerRegistration{
			Name: eventbus.ConsumerTenantOnboardingEmail, Queue: eventbus.QueueNotification,
			EventTypes: []string{eventbus.EventTenantProvisioned}, RoutingKeys: []string{"tenant.provisioned"},
			Handler: consumers.HandleTenantOnboardingEmail,
		})
	}
	return registrations
}

type Consumers struct {
	db          *gorm.DB
	emailSender onboardingemail.Sender
}

func New(db *gorm.DB) *Consumers {
	return NewWithEmailSender(db, onboardingemail.NewSender(config.EmailConfig()))
}

func NewWithEmailSender(db *gorm.DB, sender onboardingemail.Sender) *Consumers {
	return &Consumers{db: db, emailSender: sender}
}

func (c *Consumers) HandleTenantOnboardingEmail(ctx eventbus.Context) error {
	event := ctx.Event
	if event.EventType != eventbus.EventTenantProvisioned {
		return nil
	}
	jobID := uintPayload(event.Payload, "emailJobID")
	if jobID == 0 {
		return fmt.Errorf("tenant onboarding email job is required")
	}
	var job onboardingemail.Job
	if err := c.db.Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", jobID, event.TenantID, event.SiteID).Take(&job).Error; err != nil {
		return err
	}
	if job.Status == "SENT" {
		return nil
	}
	if job.EventID != nil && *job.EventID != event.EventID {
		return fmt.Errorf("tenant onboarding email event does not match job")
	}
	encryptedPayload := stringPayload(event.Payload, onboardingemail.EventEncryptedPayloadKey)
	if encryptedPayload == "" {
		// Events created before credentials were embedded in the outbox remain deliverable.
		encryptedPayload = job.EncryptedPayload
	} else if encryptedPayload != job.EncryptedPayload {
		return fmt.Errorf("tenant onboarding email payload does not match job")
	}
	payload, err := onboardingemail.DecryptPayload(encryptedPayload)
	if err != nil {
		return err
	}
	if c.emailSender == nil {
		return fmt.Errorf("tenant onboarding email sender is not configured")
	}
	now := time.Now()
	claimed := c.db.Model(&onboardingemail.Job{}).Where("id=? AND (status IN ('PENDING','FAILED','DEAD') OR (status='PROCESSING' AND updated_at<?))", job.ID, now.Add(-10*time.Minute)).Updates(map[string]any{
		"status": "PROCESSING", "attempt_count": gorm.Expr("attempt_count+1"), "last_error": "", "updated_at": now,
	})
	if claimed.Error != nil {
		return claimed.Error
	}
	if claimed.RowsAffected == 0 {
		return fmt.Errorf("tenant onboarding email is already processing")
	}
	if err := c.emailSender.Send(context.Background(), onboardingemail.ProvisionedMessage(job.RecipientEmail, payload)); err != nil {
		message := err.Error()
		if len(message) > 500 {
			message = message[:500]
		}
		_ = c.db.Model(&onboardingemail.Job{}).Where("id=?", job.ID).Updates(map[string]any{
			"status": "FAILED", "last_error": message, "updated_at": time.Now(),
		}).Error
		return err
	}
	sentAt := time.Now()
	return c.db.Model(&onboardingemail.Job{}).Where("id=?", job.ID).Updates(map[string]any{
		"status": "SENT", "last_error": "", "sent_at": sentAt, "updated_at": sentAt,
	}).Error
}

func (c *Consumers) HandleLedgerAudit(ctx eventbus.Context) error {
	event := ctx.Event
	if event.EventType != eventbus.EventLedgerTransactionPosted {
		return nil
	}
	amount, err := eventAmount(event.Payload)
	if err != nil {
		return err
	}
	now := time.Now()
	record := AuditEventRecord{
		EventID: event.EventID, EventType: event.EventType, EventVersion: event.EventVersion, ConsumerName: eventbus.ConsumerLedgerAudit,
		TenantID: event.TenantID, SiteID: event.SiteID, AggregateType: event.AggregateType, AggregateID: event.AggregateID,
		OperatorID: uintPayload(event.Metadata, "operatorID"), TransactionID: uintPayload(event.Payload, "transactionID"),
		BusinessType: stringPayload(event.Payload, "businessType"), Currency: stringPayload(event.Payload, "currency"), Amount: amount.StringFixed(18),
		TraceID: event.TraceID, CorrelationID: event.CorrelationID, CausationID: event.CausationID,
		PayloadSnapshot: marshal(event.Envelope), OccurredAt: event.OccurredAt, ConsumedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	return c.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error
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
		TenantID: event.TenantID, SiteID: event.SiteID, AggregateType: event.AggregateType, AggregateID: event.AggregateID,
		OperatorID: uintPayload(event.Metadata, "operatorID"), TransactionID: uintPayload(event.Payload, "transactionID"),
		BusinessType: businessType, Currency: stringPayload(event.Payload, "currency"), Amount: amount.StringFixed(18),
		TraceID: event.TraceID, CorrelationID: event.CorrelationID, CausationID: event.CausationID,
		PayloadSnapshot: marshal(event.Envelope), OccurredAt: event.OccurredAt, ConsumedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	return c.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error
}

func (c *Consumers) HandleDashboardFact(ctx eventbus.Context) error {
	event := ctx.Event
	metricType := dashboardMetricByEvent[event.EventType]
	if metricType == "" {
		return nil
	}
	amount, err := eventAmount(event.Payload)
	if err != nil {
		return err
	}
	occurred := event.OccurredAt
	if occurred.IsZero() {
		occurred = time.Now()
	}
	businessDate := time.Date(occurred.Year(), occurred.Month(), occurred.Day(), 0, 0, 0, 0, time.UTC)
	now := time.Now()
	fact := DashboardEventFact{
		EventID: event.EventID, ConsumerName: eventbus.ConsumerDashboardDailyStats, TenantID: event.TenantID, SiteID: event.SiteID,
		BusinessDate: businessDate, MetricType: metricType, Currency: stringPayload(event.Payload, "currency"), Amount: amount.StringFixed(18),
		TraceID: event.TraceID, CorrelationID: event.CorrelationID, CausationID: event.CausationID,
		OccurredAt: occurred, CreatedAt: now, UpdatedAt: now,
	}
	return c.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&fact).Error
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
