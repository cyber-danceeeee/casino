package eventbus

import (
	"context"
	"errors"
	"strings"
	"time"

	"casino/internal/tenancy"
)

const (
	DefaultEventVersion = 1

	EventPlayerDepositSucceeded    = "player.deposit_succeeded.v1"
	EventPlayerWithdrawApproved    = "player.withdraw_approved.v1"
	EventWalletBalanceChanged      = "wallet.balance_changed.v1"
	EventGameBetPlaced             = "game.bet_placed.v1"
	EventGameBetSettled            = "game.bet_settled.v1"
	EventSettlementCompleted       = "settlement.completed.v1"
	EventTenantCreditChanged       = "tenant.credit_changed.v1"
	EventTenantDepositCredited     = "tenant.deposit_credited.v1"
	EventTenantFundsFrozen         = "tenant.funds_frozen.v1"
	EventTenantFundsUnfrozen       = "tenant.funds_unfrozen.v1"
	EventLedgerTransactionPosted   = "ledger.transaction_posted.v1"
	EventTenantLedgerPosted        = EventLedgerTransactionPosted
	EventTenantLedgerReversed      = "ledger.transaction_reversed.v1"
	EventProviderCallbackReceived  = "provider.callback_received.v1"
	EventWhiteLabelConfigPublished = "white_label.config_published.v1"
	EventTenantProvisioned         = "tenant.provisioned.v1"
	EventSupportExportRequested    = "task.support_export.requested.v1"
	EventReportExportRequested     = "task.report_export.requested.v1"
	EventProviderSyncRequested     = "task.provider_sync.requested.v1"
	EventBackupRequested           = "task.backup.requested.v1"
	EventSystemTaskRequested       = "task.system_execution.requested.v1"
	EventNotificationSendRequested = "task.notification_send.requested.v1"

	ConsumerDashboardDailyStats   = "dashboard.daily_stats.v1"
	ConsumerLedgerAudit           = "audit.ledger_transaction.v1"
	ConsumerEventJournal          = "audit.event_journal.v1"
	ConsumerRiskRules             = "risk.rules.v1"
	ConsumerPromotionDeposit      = "promotion.deposit_trigger.v1"
	ConsumerTenantOnboardingEmail = "notification.tenant_onboarding_email.v1"
	ConsumerSupportExport         = "system.support_export.v1"
	ConsumerReportExport          = "system.report_export.v1"
	ConsumerProviderSync          = "system.provider_sync.v1"
	ConsumerBackup                = "system.backup.v1"
	ConsumerSystemTask            = "system.scheduled_task.v1"
	ConsumerNotificationPush      = "notification.push.v1"

	QueueLedger       = "ledger.events"
	QueuePayment      = "payment.events"
	QueueProvider     = "provider.events"
	QueuePromotion    = "promotion.events"
	QueueNotification = "notification.events"
	QueueAudit        = "audit.events"
	QueueDashboard    = "dashboard.events"
	QueueSystemJobs   = "system.jobs"
	QueueRetry        = "retry.events"
	QueueDead         = "dead.events"

	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusPublished  = "PUBLISHED"
	StatusFailed     = "FAILED"
	StatusDead       = "DEAD"
	StatusSuccess    = "SUCCESS"
)

var (
	ErrInvalidScope        = NewError("INVALID_SCOPE", "invalid event tenant scope")
	ErrEventTypeRequired   = NewError("EVENT_TYPE_REQUIRED", "event type is required")
	ErrIdempotencyRequired = NewError("IDEMPOTENCY_KEY_REQUIRED", "event idempotency key is required")
	ErrConsumerRequired    = NewError("CONSUMER_REQUIRED", "event consumer is required")
	ErrConsumerNameInvalid = NewError("CONSUMER_NAME_INVALID", "event consumer name must be stable and versioned")
	ErrHandlerRequired     = NewError("HANDLER_REQUIRED", "event handler is required")
	ErrEventNotFound       = NewError("EVENT_NOT_FOUND", "event not found")
	ErrEventPayloadInvalid = NewError("EVENT_PAYLOAD_INVALID", "event payload is invalid")
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func NewError(code, message string) error {
	return &Error{Code: code, Message: message}
}

func ErrorCode(err error) string {
	var target *Error
	if errors.As(err, &target) {
		return target.Code
	}
	return "INTERNAL_ERROR"
}

type Event struct {
	Scope           tenancy.Scope
	TenantID        uint64
	SiteID          uint64
	EventType       string
	EventVersion    int
	AggregateType   string
	AggregateID     string
	Source          string
	IdempotencyKey  string
	TraceID         string
	RequestID       string
	CorrelationID   string
	CausationID     string
	Producer        string
	OccurredAt      time.Time
	TargetConsumers []string
	Payload         any
	Metadata        any
}

type EventEnvelope struct {
	EventID         string         `json:"event_id"`
	EventType       string         `json:"event_type"`
	EventVersion    int            `json:"version"`
	TenantID        uint64         `json:"tenant_id"`
	SiteID          uint64         `json:"site_id"`
	RequestID       string         `json:"request_id"`
	AggregateType   string         `json:"aggregate_type"`
	AggregateID     string         `json:"aggregate_id"`
	OccurredAt      time.Time      `json:"occurred_at"`
	Producer        string         `json:"producer"`
	TraceID         string         `json:"trace_id"`
	CorrelationID   string         `json:"correlation_id"`
	CausationID     string         `json:"causation_id"`
	TargetConsumers []string       `json:"target_consumers"`
	Payload         map[string]any `json:"payload"`
	Metadata        map[string]any `json:"metadata"`
}

type Record struct {
	ID              uint64         `json:"id"`
	EventID         string         `json:"eventID"`
	TenantID        uint64         `json:"tenantID"`
	SiteID          uint64         `json:"siteID"`
	EventType       string         `json:"eventType"`
	EventVersion    int            `json:"eventVersion"`
	AggregateType   string         `json:"aggregateType"`
	AggregateID     string         `json:"aggregateID"`
	Source          string         `json:"source"`
	IdempotencyKey  string         `json:"idempotencyKey"`
	TraceID         string         `json:"traceID"`
	RequestID       string         `json:"requestID"`
	CorrelationID   string         `json:"correlationID"`
	CausationID     string         `json:"causationID"`
	Producer        string         `json:"producer"`
	OccurredAt      time.Time      `json:"occurredAt"`
	TargetConsumers []string       `json:"targetConsumers"`
	Envelope        EventEnvelope  `json:"envelope"`
	Payload         map[string]any `json:"payload"`
	Metadata        map[string]any `json:"metadata"`
	Status          string         `json:"status"`
	Attempts        uint           `json:"attempts"`
	NextAttemptAt   time.Time      `json:"nextAttemptAt"`
	PublishedAt     *time.Time     `json:"publishedAt"`
	LockedBy        string         `json:"lockedBy"`
	LockedAt        *time.Time     `json:"lockedAt"`
	LastError       string         `json:"lastError"`
	Idempotent      bool           `json:"idempotent"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

type Delivery struct {
	ID         uint64     `json:"id"`
	EventID    string     `json:"eventID"`
	OutboxID   uint64     `json:"outboxID"`
	TenantID   uint64     `json:"tenantID"`
	SiteID     uint64     `json:"siteID"`
	EventType  string     `json:"eventType"`
	Consumer   string     `json:"consumer"`
	Status     string     `json:"status"`
	Attempts   uint       `json:"attempts"`
	LastError  string     `json:"lastError"`
	ConsumedAt *time.Time `json:"consumedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type DispatchSummary struct {
	Scanned   int `json:"scanned"`
	Published int `json:"published"`
	Failed    int `json:"failed"`
	Dead      int `json:"dead"`
	Skipped   int `json:"skipped"`
}

type ReplayRangeOptions struct {
	FromID    uint64
	ToID      uint64
	EventType string
	Status    string
	Limit     int
	Force     bool
}

type ReplayConsumerOptions struct {
	Consumer string
	Limit    int
	Force    bool
}

type ReplayResult struct {
	Replayed int `json:"replayed"`
}

type Handler func(ctx Context) error

type Context struct {
	Event Record
}

type ConsumerRegistration struct {
	Name        string
	Queue       string
	EventTypes  []string
	RoutingKeys []string
	Concurrency int
	Handler     Handler
}

type PublishMessage struct {
	EventID       string
	OutboxID      uint64
	RoutingKey    string
	Body          []byte
	Headers       map[string]any
	OccurredAt    time.Time
	DeliveryMode  uint8
	ContentType   string
	CorrelationID string
	MessageType   string
}

type DeliveryMessage struct {
	Queue      string
	RoutingKey string
	Body       []byte
	Headers    map[string]any
	ReceivedAt time.Time
}

type EventPublisher interface {
	Publish(ctx context.Context, message PublishMessage) error
	Available(ctx context.Context) error
}

type EventConsumer interface {
	Start(ctx context.Context, registrations []ConsumerRegistration, dispatcher EventDispatcher) error
	Close() error
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, message DeliveryMessage, registration ConsumerRegistration) error
}

type QueueManager interface {
	EnsureTopology(ctx context.Context) error
	Health(ctx context.Context) QueueHealth
	Metrics(ctx context.Context) (QueueMetrics, error)
	DeadMessages(ctx context.Context, queue string, limit int) ([]DeadMessage, error)
	ReplayDead(ctx context.Context, queue, eventID string) error
	PurgeRetry(ctx context.Context, queue string) (int, error)
	Close() error
}

type QueueHealth struct {
	Status             string   `json:"status"`
	Connection         bool     `json:"connection"`
	PublisherAvailable bool     `json:"publisherAvailable"`
	ConsumerRunning    bool     `json:"consumerRunning"`
	MissingExchanges   []string `json:"missingExchanges"`
	MissingQueues      []string `json:"missingQueues"`
	Message            string   `json:"message"`
}

type QueueMetric struct {
	Queue      string  `json:"queue"`
	Depth      int     `json:"depth"`
	Ready      int     `json:"ready"`
	Unacked    int     `json:"unacked"`
	Consumers  int     `json:"consumers"`
	PublishTPS float64 `json:"publishTps"`
	ConsumeTPS float64 `json:"consumeTps"`
	AvgLatency float64 `json:"averageLatencyMs"`
}

type QueueMetrics struct {
	Queues        []QueueMetric `json:"queues"`
	DeadMessages  int           `json:"deadMessages"`
	RetryMessages int           `json:"retryMessages"`
	CapturedAt    time.Time     `json:"capturedAt"`
}

type DeadMessage struct {
	Queue      string         `json:"queue"`
	EventID    string         `json:"eventID"`
	EventType  string         `json:"eventType"`
	RoutingKey string         `json:"routingKey"`
	RetryCount int            `json:"retryCount"`
	Error      string         `json:"error"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"createdAt"`
}

func NormalizeType(value string) string {
	return strings.TrimSpace(value)
}
