package eventbus

import "time"

type eventOutbox struct {
	ID      uint64 `gorm:"column:id;primaryKey"`
	EventID string `gorm:"column:event_id;uniqueIndex:uk_event_outbox_event_id"`

	SiteID          uint64     `gorm:"column:site_id"`
	EventType       string     `gorm:"column:event_type"`
	EventVersion    int        `gorm:"column:event_version"`
	AggregateType   string     `gorm:"column:aggregate_type"`
	AggregateID     string     `gorm:"column:aggregate_id"`
	Source          string     `gorm:"column:source"`
	IdempotencyKey  string     `gorm:"column:idempotency_key;uniqueIndex:uk_event_outbox_idempotency"`
	TraceID         string     `gorm:"column:trace_id"`
	CorrelationID   string     `gorm:"column:correlation_id"`
	CausationID     string     `gorm:"column:causation_id"`
	TargetConsumers string     `gorm:"column:target_consumers;type:json"`
	Payload         string     `gorm:"column:payload"`
	Metadata        string     `gorm:"column:metadata"`
	Status          string     `gorm:"column:status"`
	Attempts        uint       `gorm:"column:attempts"`
	NextAttemptAt   time.Time  `gorm:"column:next_attempt_at"`
	PublishedAt     *time.Time `gorm:"column:published_at"`
	LockedBy        string     `gorm:"column:locked_by"`
	LockedAt        *time.Time `gorm:"column:locked_at"`
	LastError       string     `gorm:"column:last_error"`
	OccurredAt      time.Time  `gorm:"column:occurred_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at"`
}

func (eventOutbox) TableName() string { return "event_outbox" }

type eventDelivery struct {
	ID       uint64 `gorm:"column:id;primaryKey"`
	EventID  string `gorm:"column:event_id;uniqueIndex:uk_event_delivery_consumer"`
	OutboxID uint64 `gorm:"column:outbox_id"`

	SiteID     uint64     `gorm:"column:site_id"`
	EventType  string     `gorm:"column:event_type"`
	LegacyName string     `gorm:"column:consumer"`
	Consumer   string     `gorm:"column:consumer_name;uniqueIndex:uk_event_delivery_consumer"`
	Status     string     `gorm:"column:status"`
	Attempts   uint       `gorm:"column:attempts"`
	LastError  string     `gorm:"column:last_error"`
	ConsumedAt *time.Time `gorm:"column:consumed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at"`
}

func (eventDelivery) TableName() string { return "event_deliveries" }
