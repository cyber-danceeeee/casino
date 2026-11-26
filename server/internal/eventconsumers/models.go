package eventconsumers

import "time"

type AuditEventRecord struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement;index:idx_audit_event_scope_time,priority:4"`
	EventID         string    `gorm:"column:event_id;type:varchar(64);not null;uniqueIndex:uk_audit_event_consumer,priority:1"`
	EventType       string    `gorm:"column:event_type;type:varchar(128);not null"`
	EventVersion    int       `gorm:"column:event_version;not null;default:1"`
	ConsumerName    string    `gorm:"column:consumer_name;type:varchar(128);not null;uniqueIndex:uk_audit_event_consumer,priority:2"`
	TenantID        uint64    `gorm:"column:tenant_id;not null;default:0;index:idx_audit_event_scope_time,priority:1"`
	SiteID          uint64    `gorm:"column:site_id;not null;default:0;index:idx_audit_event_scope_time,priority:2"`
	AggregateType   string    `gorm:"column:aggregate_type;type:varchar(64);not null;default:''"`
	AggregateID     string    `gorm:"column:aggregate_id;type:varchar(128);not null;default:''"`
	OperatorID      uint64    `gorm:"column:operator_id"`
	TransactionID   uint64    `gorm:"column:transaction_id;not null;default:0;index:idx_audit_event_transaction,priority:1"`
	BusinessType    string    `gorm:"column:business_type;type:varchar(64);not null;default:''"`
	Currency        string    `gorm:"column:currency;type:varchar(16);not null;default:'';index:idx_audit_event_transaction,priority:2"`
	Amount          string    `gorm:"column:amount;type:decimal(36,18);not null;default:0"`
	TraceID         string    `gorm:"column:trace_id;type:varchar(128);not null;default:'';index:idx_audit_event_trace,priority:1"`
	CorrelationID   string    `gorm:"column:correlation_id;type:varchar(128);not null;default:'';index:idx_audit_event_trace,priority:2"`
	CausationID     string    `gorm:"column:causation_id;type:varchar(128);not null;default:'';index:idx_audit_event_trace,priority:3"`
	PayloadSnapshot string    `gorm:"column:payload_snapshot;type:json"`
	OccurredAt      time.Time `gorm:"column:occurred_at;not null;index:idx_audit_event_scope_time,priority:3"`
	ConsumedAt      time.Time `gorm:"column:consumed_at"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (AuditEventRecord) TableName() string { return "audit_event_records" }

type DashboardEventFact struct {
	ID            uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	EventID       string    `gorm:"column:event_id;type:varchar(64);not null;uniqueIndex:uk_dashboard_fact_event_metric,priority:1"`
	ConsumerName  string    `gorm:"column:consumer_name;type:varchar(128);not null;uniqueIndex:uk_dashboard_fact_event_metric,priority:2"`
	TenantID      uint64    `gorm:"column:tenant_id;not null;default:0;index:idx_dashboard_fact_scope_date,priority:1"`
	SiteID        uint64    `gorm:"column:site_id;not null;default:0;index:idx_dashboard_fact_scope_date,priority:2"`
	BusinessDate  time.Time `gorm:"column:business_date;type:date;not null;index:idx_dashboard_fact_scope_date,priority:3"`
	MetricType    string    `gorm:"column:metric_type;type:varchar(64);not null;uniqueIndex:uk_dashboard_fact_event_metric,priority:3;index:idx_dashboard_fact_scope_date,priority:4"`
	Currency      string    `gorm:"column:currency;type:varchar(16);not null;default:'';index:idx_dashboard_fact_scope_date,priority:5"`
	Amount        string    `gorm:"column:amount;type:decimal(36,18);not null;default:0"`
	TraceID       string    `gorm:"column:trace_id;type:varchar(128);not null;default:'';index:idx_dashboard_fact_trace,priority:1"`
	CorrelationID string    `gorm:"column:correlation_id;type:varchar(128);not null;default:'';index:idx_dashboard_fact_trace,priority:2"`
	CausationID   string    `gorm:"column:causation_id;type:varchar(128);not null;default:'';index:idx_dashboard_fact_trace,priority:3"`
	OccurredAt    time.Time `gorm:"column:occurred_at"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (DashboardEventFact) TableName() string { return "dashboard_event_facts" }
