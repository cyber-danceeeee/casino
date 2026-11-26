package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"casino/internal/content"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db           *gorm.DB
	mu           sync.RWMutex
	handlers     map[string]handlerRegistration
	now          func() time.Time
	workerID     string
	maxAttempts  uint
	leaseTimeout time.Duration
}

type Options struct {
	WorkerID     string
	MaxAttempts  uint
	LeaseTimeout time.Duration
}

func New(db *gorm.DB) *Service {
	return NewWithOptions(db, Options{})
}

func NewWithOptions(db *gorm.DB, options Options) *Service {
	workerID := strings.TrimSpace(options.WorkerID)
	if workerID == "" {
		workerID = fmt.Sprintf("event-worker-%d-%d", os.Getpid(), time.Now().UnixNano())
	}
	maxAttempts := options.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 10
	}
	leaseTimeout := options.LeaseTimeout
	if leaseTimeout <= 0 {
		leaseTimeout = 5 * time.Minute
	}
	return &Service{db: db, handlers: map[string]handlerRegistration{}, now: time.Now, workerID: workerID, maxAttempts: maxAttempts, leaseTimeout: leaseTimeout}
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&eventOutbox{}, &eventDelivery{})
}

func (s *Service) Register(consumer string, handler Handler) error {
	return s.RegisterFor(consumer, nil, handler)
}

func (s *Service) RegisterFor(consumer string, eventTypes []string, handler Handler) error {
	consumer = strings.TrimSpace(consumer)
	if consumer == "" {
		return ErrConsumerRequired
	}
	if !stableConsumerPattern.MatchString(consumer) {
		return ErrConsumerNameInvalid
	}
	if handler == nil {
		return ErrHandlerRequired
	}
	types := map[string]struct{}{}
	for _, eventType := range eventTypes {
		eventType = NormalizeType(eventType)
		if eventType != "" {
			types[eventType] = struct{}{}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[consumer] = handlerRegistration{handler: handler, eventTypes: types}
	return nil
}

func (s *Service) Publish(ctx context.Context, event Event) (Record, error) {
	var result Record
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = s.PublishTx(tx, event)
		return err
	})
	return result, err
}

func (s *Service) PublishTx(tx *gorm.DB, event Event) (Record, error) {
	if err := validateEvent(event); err != nil {
		return Record{}, err
	}
	if existing, err := s.eventByIdempotencyTx(tx, event.IdempotencyKey); err == nil {
		existing.Idempotent = true
		return existing, nil
	} else if err != gorm.ErrRecordNotFound {
		return Record{}, err
	}
	now := s.now()
	eventVersion := event.EventVersion
	if eventVersion <= 0 {
		eventVersion = DefaultEventVersion
	}
	occurredAt := event.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = now
	}
	payload, err := normalizeMap(event.Payload)
	if err != nil {
		return Record{}, ErrEventPayloadInvalid
	}
	metadata, err := normalizeMap(event.Metadata)
	if err != nil {
		return Record{}, ErrEventPayloadInvalid
	}
	targetConsumers, err := normalizeConsumers(event.TargetConsumers)
	if err != nil {
		return Record{}, err
	}
	eventID := uniqueID("EVT")
	eventType := NormalizeType(event.EventType)
	envelope := EventEnvelope{
		EventID: eventID, EventType: eventType, EventVersion: eventVersion, SiteID: event.SiteID,
		RequestID: strings.TrimSpace(event.RequestID), Producer: strings.TrimSpace(event.Producer),
		AggregateType: strings.TrimSpace(event.AggregateType), AggregateID: strings.TrimSpace(event.AggregateID), OccurredAt: occurredAt,
		TraceID: strings.TrimSpace(event.TraceID), CorrelationID: strings.TrimSpace(event.CorrelationID), CausationID: strings.TrimSpace(event.CausationID),
		TargetConsumers: targetConsumers, Payload: payload, Metadata: metadata,
	}
	if envelope.RequestID == "" {
		envelope.RequestID = envelope.TraceID
	}
	if envelope.Producer == "" {
		envelope.Producer = strings.TrimSpace(event.Source)
	}
	row := eventOutbox{
		EventID: eventID,

		SiteID:          event.SiteID,
		EventType:       eventType,
		EventVersion:    eventVersion,
		AggregateType:   strings.TrimSpace(event.AggregateType),
		AggregateID:     strings.TrimSpace(event.AggregateID),
		Source:          strings.TrimSpace(event.Source),
		IdempotencyKey:  strings.TrimSpace(event.IdempotencyKey),
		TraceID:         envelope.TraceID,
		CorrelationID:   strings.TrimSpace(event.CorrelationID),
		CausationID:     envelope.CausationID,
		TargetConsumers: marshal(targetConsumers),
		Payload:         marshal(envelope),
		Metadata:        marshal(metadata),
		Status:          StatusPending,
		NextAttemptAt:   now,
		OccurredAt:      occurredAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := tx.Create(&row).Error; err != nil {
		if isDuplicateError(err) {
			existing, existingErr := s.eventByIdempotencyTx(tx, event.IdempotencyKey)
			if existingErr == nil {
				existing.Idempotent = true
			}
			return existing, existingErr
		}
		return Record{}, err
	}
	return recordFromRow(row), nil
}

func (s *Service) DispatchPending(ctx context.Context, limit int) (DispatchSummary, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	handlers := s.snapshotHandlers()
	if len(handlers) == 0 {
		count, err := s.countDue(ctx, limit)
		if err != nil {
			return DispatchSummary{}, err
		}
		return DispatchSummary{Scanned: count, Skipped: count}, nil
	}
	if err := s.recoverExpiredProcessing(ctx); err != nil {
		return DispatchSummary{}, err
	}
	rows, err := s.claimDue(ctx, limit, handlers)
	if err != nil {
		return DispatchSummary{}, err
	}
	summary := DispatchSummary{Scanned: len(rows)}
	for _, row := range rows {
		status, err := s.dispatchClaimed(ctx, row, handlers)
		if err != nil {
			summary.Failed++
			continue
		}
		switch status {
		case StatusPublished:
			summary.Published++
		case StatusDead:
			summary.Dead++
		case StatusFailed:
			summary.Failed++
		default:
			summary.Skipped++
		}
	}
	return summary, nil
}

func (s *Service) countDue(ctx context.Context, limit int) (int, error) {
	rows := []eventOutbox{}
	now := s.now()
	if err := dueQuery(s.db.WithContext(ctx), now).Order("next_attempt_at ASC,id ASC").Limit(limit).Find(&rows).Error; err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (s *Service) claimDue(ctx context.Context, limit int, handlers []handlerItem) ([]eventOutbox, error) {
	now := s.now()
	claimed := []eventOutbox{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows := []eventOutbox{}
		query := dueQuery(tx, now)
		if eventTypes, all := targetEventTypes(handlers); !all {
			query = query.Where("event_type IN ?", eventTypes)
		}
		if err := lockForClaim(query).Order("next_attempt_at ASC,id ASC").Limit(limit).Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			if row.Attempts >= s.maxAttempts {
				if err := s.markDeadTx(tx, row.ID, row.Attempts, "event exceeded maximum dispatch attempts"); err != nil {
					return err
				}
				continue
			}
			result := tx.Model(&eventOutbox{}).
				Where("id=? AND status IN ? AND next_attempt_at<=? AND deleted_at IS NULL", row.ID, []string{StatusPending, StatusFailed}, now).
				Updates(map[string]any{
					"status": StatusProcessing, "attempts": gorm.Expr("attempts+1"), "locked_by": s.workerID, "locked_at": now,
					"last_error": "", "updated_at": now,
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
			row.LastError = ""
			claimed = append(claimed, row)
		}
		return nil
	})
	return claimed, err
}

func dueQuery(db *gorm.DB, now time.Time) *gorm.DB {
	return db.Where("status IN ? AND next_attempt_at<=? AND deleted_at IS NULL", []string{StatusPending, StatusFailed}, now)
}

func (s *Service) recoverExpiredProcessing(ctx context.Context) error {
	now := s.now()
	cutoff := now.Add(-s.leaseTimeout)
	return s.db.WithContext(ctx).Model(&eventOutbox{}).
		Where("status=? AND locked_at IS NOT NULL AND locked_at<=? AND deleted_at IS NULL", StatusProcessing, cutoff).
		Updates(map[string]any{
			"status": StatusFailed, "last_error": "processing lease expired", "locked_by": "", "locked_at": nil,
			"next_attempt_at": now, "updated_at": now,
		}).Error
}

func (s *Service) dispatchClaimed(ctx context.Context, row eventOutbox, handlers []handlerItem) (string, error) {
	record, err := recordFromRowStrict(row)
	if err != nil {
		return s.markFailed(ctx, row.ID, row.Attempts, err.Error())
	}
	handlers = targetHandlers(record, handlers)
	if len(handlers) == 0 {
		if err := s.restorePending(ctx, row.ID); err != nil {
			return StatusFailed, err
		}
		return StatusPending, nil
	}
	if err := s.ensureDeliveries(ctx, record, handlers); err != nil {
		return StatusFailed, err
	}
	var failures []string
	for _, item := range handlers {
		if err := s.dispatchToConsumer(ctx, record, item.consumer, item.handler); err != nil {
			failures = append(failures, item.consumer+": "+err.Error())
		}
	}
	if len(failures) > 0 {
		status, err := s.markFailed(ctx, row.ID, row.Attempts, strings.Join(failures, "; "))
		if err != nil {
			return status, err
		}
		return status, nil
	}
	return StatusPublished, s.markPublished(ctx, row.ID)
}

func (s *Service) ensureDeliveries(ctx context.Context, record Record, handlers []handlerItem) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := s.now()
		for _, item := range handlers {
			row := eventDelivery{
				EventID: record.EventID, OutboxID: record.ID, SiteID: record.SiteID,
				EventType: record.EventType, LegacyName: item.consumer, Consumer: item.consumer, Status: StatusPending, CreatedAt: now, UpdatedAt: now,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) dispatchToConsumer(ctx context.Context, record Record, consumer string, handler Handler) error {
	skipped := false
	var delivery eventDelivery
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := s.now()
		if err := lockForUpdate(tx).Where("event_id=? AND consumer_name=? AND deleted_at IS NULL", record.EventID, consumer).Take(&delivery).Error; err != nil {
			return err
		}
		if delivery.Status == StatusSuccess {
			skipped = true
			return nil
		}
		return tx.Model(&eventDelivery{}).Where("id=?", delivery.ID).Updates(map[string]any{
			"status": StatusPending, "attempts": gorm.Expr("attempts+1"), "last_error": "", "updated_at": now,
		}).Error
	})
	if err != nil || skipped {
		return err
	}
	if err = handler(Context{Event: record}); err != nil {
		_ = s.updateDelivery(ctx, delivery.ID, StatusFailed, err.Error())
		return err
	}
	return s.updateDelivery(ctx, delivery.ID, StatusSuccess, "")
}

func (s *Service) updateDelivery(ctx context.Context, id uint64, status string, message string) error {
	now := s.now()
	updates := map[string]any{"status": status, "last_error": truncate(message, 1000), "updated_at": now}
	if status == StatusSuccess {
		updates["consumed_at"] = now
	}
	return s.db.WithContext(ctx).Model(&eventDelivery{}).Where("id=?", id).Updates(updates).Error
}

func (s *Service) markPublished(ctx context.Context, id uint64) error {
	now := s.now()
	return s.db.WithContext(ctx).Model(&eventOutbox{}).Where("id=?", id).Updates(map[string]any{
		"status": StatusPublished, "published_at": now, "last_error": "", "locked_by": "", "locked_at": nil, "updated_at": now,
	}).Error
}

func (s *Service) restorePending(ctx context.Context, id uint64) error {
	now := s.now()
	return s.db.WithContext(ctx).Model(&eventOutbox{}).Where("id=?", id).Updates(map[string]any{
		"status": StatusPending, "locked_by": "", "locked_at": nil, "updated_at": now,
	}).Error
}

func (s *Service) markFailed(ctx context.Context, id uint64, attempts uint, message string) (string, error) {
	now := s.now()
	if attempts >= s.maxAttempts {
		err := s.db.WithContext(ctx).Model(&eventOutbox{}).Where("id=?", id).Updates(map[string]any{
			"status": StatusDead, "last_error": truncate(message, 1000), "locked_by": "", "locked_at": nil,
			"next_attempt_at": now, "updated_at": now,
		}).Error
		return StatusDead, err
	}
	err := s.db.WithContext(ctx).Model(&eventOutbox{}).Where("id=?", id).Updates(map[string]any{
		"status": StatusFailed, "last_error": truncate(message, 1000), "locked_by": "", "locked_at": nil,
		"next_attempt_at": now.Add(backoffDelay(id)), "updated_at": now,
	}).Error
	return StatusFailed, err
}

func (s *Service) markDeadTx(tx *gorm.DB, id uint64, attempts uint, message string) error {
	now := s.now()
	return tx.Model(&eventOutbox{}).Where("id=?", id).Updates(map[string]any{
		"status": StatusDead, "attempts": attempts, "last_error": truncate(message, 1000), "locked_by": "", "locked_at": nil,
		"next_attempt_at": now, "updated_at": now,
	}).Error
}

func (s *Service) Get(ctx context.Context, id uint64, scope content.Scope) (Record, []Delivery, error) {
	var row eventOutbox
	db := s.db.WithContext(ctx).Where("id=? AND deleted_at IS NULL", id)
	db = applyScope(db, scope)
	if err := db.Take(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Record{}, nil, ErrEventNotFound
		}
		return Record{}, nil, err
	}
	deliveries := []eventDelivery{}
	if err := s.db.WithContext(ctx).Where("outbox_id=? AND deleted_at IS NULL", row.ID).Order("consumer_name ASC").Find(&deliveries).Error; err != nil {
		return Record{}, nil, err
	}
	out := make([]Delivery, 0, len(deliveries))
	for _, item := range deliveries {
		out = append(out, deliveryFromRow(item))
	}
	return recordFromRow(row), out, nil
}

func (s *Service) ReplayEvent(ctx context.Context, id uint64, scope content.Scope, force bool) (Record, error) {
	var record Record
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row eventOutbox
		db := applyScope(lockForUpdate(tx).Where("id=? AND deleted_at IS NULL", id), scope)
		if err := db.Take(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrEventNotFound
			}
			return err
		}
		now := s.now()
		if err := tx.Model(&eventOutbox{}).Where("id=?", row.ID).Updates(map[string]any{
			"status": StatusPending, "last_error": "", "locked_by": "", "locked_at": nil, "next_attempt_at": now, "updated_at": now,
		}).Error; err != nil {
			return err
		}
		if force {
			if err := tx.Model(&eventDelivery{}).Where("outbox_id=? AND deleted_at IS NULL", row.ID).Updates(map[string]any{
				"status": StatusPending, "last_error": "", "consumed_at": nil, "updated_at": now,
			}).Error; err != nil {
				return err
			}
		}
		row.Status = StatusPending
		row.LastError = ""
		row.LockedBy = ""
		row.LockedAt = nil
		row.NextAttemptAt = now
		row.UpdatedAt = now
		record = recordFromRow(row)
		return nil
	})
	return record, err
}

func (s *Service) ReplayRange(ctx context.Context, scope content.Scope, options ReplayRangeOptions) (ReplayResult, error) {
	limit := replayLimit(options.Limit)
	db := applyScope(s.db.WithContext(ctx).Where("deleted_at IS NULL"), scope)
	if options.FromID > 0 {
		db = db.Where("id>=?", options.FromID)
	}
	if options.ToID > 0 {
		db = db.Where("id<=?", options.ToID)
	}
	if value := strings.TrimSpace(options.EventType); value != "" {
		db = db.Where("event_type=?", value)
	}
	if value := strings.TrimSpace(options.Status); value != "" {
		db = db.Where("status=?", strings.ToUpper(value))
	}
	rows := []eventOutbox{}
	if err := db.Order("id ASC").Limit(limit).Find(&rows).Error; err != nil {
		return ReplayResult{}, err
	}
	count, err := s.replayRows(ctx, rows, options.Force)
	return ReplayResult{Replayed: count}, err
}

func (s *Service) ReplayConsumer(ctx context.Context, scope content.Scope, options ReplayConsumerOptions) (ReplayResult, error) {
	consumer := strings.TrimSpace(options.Consumer)
	if consumer == "" {
		return ReplayResult{}, ErrConsumerRequired
	}
	if !stableConsumerPattern.MatchString(consumer) {
		return ReplayResult{}, ErrConsumerNameInvalid
	}
	limit := replayLimit(options.Limit)
	db := applyScope(s.db.WithContext(ctx).Model(&eventDelivery{}).Where("deleted_at IS NULL AND consumer_name=?", consumer), scope)
	if !options.Force {
		db = db.Where("status<>?", StatusSuccess)
	}
	deliveries := []eventDelivery{}
	if err := db.Order("outbox_id ASC").Limit(limit).Find(&deliveries).Error; err != nil {
		return ReplayResult{}, err
	}
	ids := make([]uint64, 0, len(deliveries))
	seen := map[uint64]struct{}{}
	for _, delivery := range deliveries {
		if _, ok := seen[delivery.OutboxID]; ok {
			continue
		}
		seen[delivery.OutboxID] = struct{}{}
		ids = append(ids, delivery.OutboxID)
	}
	if len(ids) == 0 {
		return ReplayResult{}, nil
	}
	rows := []eventOutbox{}
	if err := applyScope(s.db.WithContext(ctx).Where("id IN ? AND deleted_at IS NULL", ids), scope).Find(&rows).Error; err != nil {
		return ReplayResult{}, err
	}
	count, err := s.replayRows(ctx, rows, options.Force)
	return ReplayResult{Replayed: count}, err
}

func (s *Service) replayRows(ctx context.Context, rows []eventOutbox, force bool) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := s.now()
		ids := make([]uint64, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		if err := tx.Model(&eventOutbox{}).Where("id IN ?", ids).Updates(map[string]any{
			"status": StatusPending, "last_error": "", "locked_by": "", "locked_at": nil, "next_attempt_at": now, "updated_at": now,
		}).Error; err != nil {
			return err
		}
		if force {
			return tx.Model(&eventDelivery{}).Where("outbox_id IN ? AND deleted_at IS NULL", ids).Updates(map[string]any{
				"status": StatusPending, "last_error": "", "consumed_at": nil, "updated_at": now,
			}).Error
		}
		return nil
	})
	return len(rows), err
}

func (s *Service) List(ctx context.Context, scope content.Scope, values map[string]string, page int, pageSize int) ([]Record, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	db := applyScope(s.db.WithContext(ctx).Model(&eventOutbox{}).Where("deleted_at IS NULL"), scope)
	if value := strings.TrimSpace(values["eventType"]); value != "" {
		db = db.Where("event_type=?", value)
	}
	if value := strings.TrimSpace(values["status"]); value != "" {
		db = db.Where("status=?", strings.ToUpper(value))
	}
	if value := strings.TrimSpace(values["aggregateType"]); value != "" {
		db = db.Where("aggregate_type=?", value)
	}
	if value := strings.TrimSpace(values["aggregateID"]); value != "" {
		db = db.Where("aggregate_id=?", value)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := []eventOutbox{}
	if err := db.Order("created_at DESC,id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]Record, 0, len(rows))
	for _, row := range rows {
		out = append(out, recordFromRow(row))
	}
	return out, total, nil
}

type handlerItem struct {
	consumer   string
	handler    Handler
	eventTypes map[string]struct{}
}

type handlerRegistration struct {
	handler    Handler
	eventTypes map[string]struct{}
}

func (s *Service) snapshotHandlers() []handlerItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]handlerItem, 0, len(s.handlers))
	for consumer, registration := range s.handlers {
		out = append(out, handlerItem{consumer: consumer, handler: registration.handler, eventTypes: registration.eventTypes})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].consumer < out[j].consumer })
	return out
}

func targetEventTypes(handlers []handlerItem) ([]string, bool) {
	values := map[string]struct{}{}
	for _, item := range handlers {
		if len(item.eventTypes) == 0 {
			return nil, true
		}
		for eventType := range item.eventTypes {
			values[eventType] = struct{}{}
		}
	}
	out := make([]string, 0, len(values))
	for eventType := range values {
		out = append(out, eventType)
	}
	sort.Strings(out)
	return out, false
}

func targetHandlers(record Record, handlers []handlerItem) []handlerItem {
	allowed := map[string]struct{}{}
	for _, consumer := range record.TargetConsumers {
		allowed[consumer] = struct{}{}
	}
	out := make([]handlerItem, 0, len(handlers))
	for _, item := range handlers {
		if len(allowed) > 0 {
			if _, ok := allowed[item.consumer]; !ok {
				continue
			}
		}
		if len(item.eventTypes) > 0 {
			if _, ok := item.eventTypes[record.EventType]; !ok {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}

func validateEvent(event Event) error {
	if err := validateScope(event.Scope, event.SiteID); err != nil {
		return err
	}
	if NormalizeType(event.EventType) == "" {
		return ErrEventTypeRequired
	}
	if strings.TrimSpace(event.IdempotencyKey) == "" {
		return ErrIdempotencyRequired
	}
	return nil
}

func validateScope(scope content.Scope, siteID uint64) error {
	if scope.Platform {
		return nil
	}
	if siteID == 0 || scope.SiteID != siteID {
		return ErrInvalidScope
	}
	if scope.SiteID > 0 && siteID > 0 && scope.SiteID != siteID {
		return ErrInvalidScope
	}
	return nil
}

func applyScope(db *gorm.DB, scope content.Scope) *gorm.DB {
	if !scope.SiteScoped() {
		return db
	}
	db = db
	if scope.SiteScoped() {
		db = db.Where("site_id IN (0, ?)", scope.SiteID)
	}
	return db
}

func (s *Service) eventByIdempotencyTx(tx *gorm.DB, key string) (Record, error) {
	var row eventOutbox
	if err := tx.Where("idempotency_key=? AND deleted_at IS NULL", strings.TrimSpace(key)).Take(&row).Error; err != nil {
		return Record{}, err
	}
	return recordFromRow(row), nil
}

func recordFromRow(row eventOutbox) Record {
	record, err := recordFromRowStrict(row)
	if err == nil {
		return record
	}
	version := row.EventVersion
	if version <= 0 {
		version = DefaultEventVersion
	}
	occurredAt := row.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = row.CreatedAt
	}
	return Record{
		ID: row.ID, EventID: row.EventID, SiteID: row.SiteID, EventType: row.EventType,
		EventVersion: version, AggregateType: row.AggregateType, AggregateID: row.AggregateID, Source: row.Source, IdempotencyKey: row.IdempotencyKey,
		TraceID: row.TraceID, RequestID: row.TraceID, CorrelationID: row.CorrelationID, CausationID: row.CausationID, Producer: row.Source, OccurredAt: occurredAt,
		TargetConsumers: unmarshalStringSlice(row.TargetConsumers), Payload: map[string]any{}, Metadata: unmarshalMap(row.Metadata), Status: row.Status, Attempts: row.Attempts,
		NextAttemptAt: row.NextAttemptAt, PublishedAt: row.PublishedAt, LockedBy: row.LockedBy, LockedAt: row.LockedAt,
		LastError: row.LastError, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func recordFromRowStrict(row eventOutbox) (Record, error) {
	envelope, err := envelopeFromRow(row)
	if err != nil {
		return Record{}, err
	}
	return Record{
		ID: row.ID, EventID: row.EventID, SiteID: row.SiteID, EventType: row.EventType,
		EventVersion:  envelope.EventVersion,
		AggregateType: row.AggregateType, AggregateID: row.AggregateID, Source: row.Source, IdempotencyKey: row.IdempotencyKey,
		TraceID: row.TraceID, RequestID: envelope.RequestID, CorrelationID: row.CorrelationID, CausationID: row.CausationID, Producer: envelope.Producer, OccurredAt: row.OccurredAt,
		TargetConsumers: envelope.TargetConsumers, Envelope: envelope, Payload: envelope.Payload, Metadata: envelope.Metadata,
		Status: row.Status, Attempts: row.Attempts, NextAttemptAt: row.NextAttemptAt, PublishedAt: row.PublishedAt, LockedBy: row.LockedBy, LockedAt: row.LockedAt,
		LastError: row.LastError, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func deliveryFromRow(row eventDelivery) Delivery {
	return Delivery{ID: row.ID, EventID: row.EventID, OutboxID: row.OutboxID, SiteID: row.SiteID, EventType: row.EventType, Consumer: row.Consumer, Status: row.Status, Attempts: row.Attempts, LastError: row.LastError, ConsumedAt: row.ConsumedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}

func envelopeFromRow(row eventOutbox) (EventEnvelope, error) {
	var envelope EventEnvelope
	if err := json.Unmarshal([]byte(row.Payload), &envelope); err != nil {
		return EventEnvelope{}, ErrEventPayloadInvalid
	}
	if strings.TrimSpace(envelope.EventID) != "" {
		if envelope.EventVersion <= 0 {
			var legacy struct {
				EventVersion int `json:"event_version"`
			}
			_ = json.Unmarshal([]byte(row.Payload), &legacy)
			envelope.EventVersion = legacy.EventVersion
		}
		if envelope.EventVersion <= 0 {
			envelope.EventVersion = row.EventVersion
		}
		if envelope.EventVersion <= 0 {
			envelope.EventVersion = DefaultEventVersion
		}
		if envelope.Payload == nil {
			envelope.Payload = map[string]any{}
		}
		if envelope.Metadata == nil {
			envelope.Metadata = unmarshalMap(row.Metadata)
		}
		if envelope.RequestID == "" {
			envelope.RequestID = row.TraceID
		}
		if envelope.Producer == "" {
			envelope.Producer = row.Source
		}
		if len(envelope.TargetConsumers) == 0 {
			envelope.TargetConsumers = unmarshalStringSlice(row.TargetConsumers)
		}
		return envelope, nil
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
		return EventEnvelope{}, ErrEventPayloadInvalid
	}
	occurredAt := row.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = row.CreatedAt
	}
	version := row.EventVersion
	if version <= 0 {
		version = DefaultEventVersion
	}
	return EventEnvelope{
		EventID: row.EventID, EventType: row.EventType, EventVersion: version, SiteID: row.SiteID,
		RequestID: row.TraceID, Producer: row.Source,
		AggregateType: row.AggregateType, AggregateID: row.AggregateID, OccurredAt: occurredAt,
		TraceID: row.TraceID, CorrelationID: row.CorrelationID, CausationID: row.CausationID, TargetConsumers: unmarshalStringSlice(row.TargetConsumers), Payload: payload, Metadata: unmarshalMap(row.Metadata),
	}, nil
}

func marshal(value any) string {
	if value == nil {
		return "{}"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func normalizeMap(value any) (map[string]any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizeConsumers(values []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if !stableConsumerPattern.MatchString(value) {
			return nil, ErrConsumerNameInvalid
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}

func unmarshalMap(value string) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func unmarshalStringSlice(value string) []string {
	out := []string{}
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func uniqueID(prefix string) string {
	return fmt.Sprintf("%s%s", prefix, time.Now().UTC().Format("20060102150405.000000000"))
}

func isDuplicateError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

var stableConsumerPattern = regexp.MustCompile(`^[a-z0-9]+([._:-][a-z0-9]+)*\.v[0-9]+$`)

func lockForUpdate(db *gorm.DB) *gorm.DB {
	if db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		return db
	}
	return db.Clauses(clause.Locking{Strength: "UPDATE"})
}

func lockForClaim(db *gorm.DB) *gorm.DB {
	if db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		return db
	}
	return db.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
}

func backoffDelay(id uint64) time.Duration {
	seconds := 30 + int(id%30)
	return time.Duration(seconds) * time.Second
}

func replayLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 500 {
		return 500
	}
	return limit
}
