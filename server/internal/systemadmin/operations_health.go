package systemadmin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"casino/internal/config"
	"casino/internal/database"
	"casino/internal/eventbus"
	"casino/internal/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HealthComponent struct {
	ComponentCode       string         `json:"componentCode"`
	ComponentType       string         `json:"componentType"`
	Status              string         `json:"status"`
	LatencyMS           int64          `json:"latencyMS"`
	ErrorSummary        string         `json:"errorSummary"`
	ConsecutiveFailures uint           `json:"consecutiveFailures"`
	LastCheckedAt       time.Time      `json:"lastCheckedAt"`
	LastRecoveredAt     *time.Time     `json:"lastRecoveredAt"`
	Details             map[string]any `json:"details,omitempty"`
}

type AlertActionInput struct {
	Note   string `json:"note"`
	Reason string `json:"reason"`
}

func (s *Service) Readiness(ctx context.Context) (bool, map[string]any) {
	checks := map[string]any{}
	dbOK := false
	if sqlDB, err := s.db.DB(); err == nil {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		start := time.Now()
		err = sqlDB.PingContext(checkCtx)
		cancel()
		dbOK = err == nil
		checks["database"] = map[string]any{"status": healthStatus(err), "latencyMS": time.Since(start).Milliseconds()}
	} else {
		checks["database"] = map[string]any{"status": "UNHEALTHY"}
	}
	redisCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	start := time.Now()
	redisErr := safeRedisPing(redisCtx)
	cancel()
	checks["redis"] = map[string]any{"status": healthStatus(redisErr), "latencyMS": time.Since(start).Milliseconds()}
	rabbitOK := true
	if rabbitConfig := config.RabbitMQConfig(); rabbitConfig.Enabled {
		manager := eventbus.QueueManagerInstance()
		if manager == nil {
			rabbitOK = !rabbitConfig.Required
			checks["rabbitmq"] = map[string]any{"status": "UNHEALTHY", "message": "RabbitMQ adapter is not initialized"}
		} else {
			healthCtx, healthCancel := context.WithTimeout(ctx, 3*time.Second)
			health := manager.Health(healthCtx)
			healthCancel()
			rabbitOK = health.Status == "UP" || !rabbitConfig.Required
			checks["rabbitmq"] = health
		}
	}
	return dbOK && redisErr == nil && rabbitOK, checks
}

func (s *Service) DetailedHealth(ctx context.Context) (map[string]any, error) {
	components := []HealthComponent{{ComponentCode: "main-api", ComponentType: "API", Status: "HEALTHY", LastCheckedAt: now(), Details: map[string]any{"uptimeSeconds": int64(now().Sub(applicationStartedAt).Seconds())}}}
	components = append(components, s.checkHealth(ctx, "database", "DATABASE", func(checkCtx context.Context) (map[string]any, error) {
		sqlDB, err := s.db.DB()
		if err != nil {
			return nil, err
		}
		err = sqlDB.PingContext(checkCtx)
		stats := sqlDB.Stats()
		return map[string]any{"openConnections": stats.OpenConnections, "inUseConnections": stats.InUse}, err
	}))
	components = append(components, s.checkHealth(ctx, "redis", "CACHE", func(checkCtx context.Context) (map[string]any, error) { return nil, safeRedisPing(checkCtx) }))
	components = append(components, s.checkHealth(ctx, "object-storage", "STORAGE", func(checkCtx context.Context) (map[string]any, error) { return nil, storage.Check(checkCtx) }))
	components = append(components, s.databaseQueueHealth("event-outbox-relay", "EVENT_OUTBOX", "SELECT COUNT(*) FROM event_outbox WHERE status IN ('FAILED','DEAD') AND deleted_at IS NULL", 0, 10))
	components = append(components, s.rabbitMQHealth(ctx))
	components = append(components, s.schedulerHealth())
	components = append(components, s.workerHealth())
	components = append(components, s.externalAggregateHealth(ctx, "payment-providers", "PAYMENT_PROVIDER", func(checkCtx context.Context) ([]string, error) {
		data, err := New(s.db.WithContext(checkCtx)).PaymentMonitor(Query{Values: map[string]string{"from": now().Add(-time.Hour).Format(time.RFC3339), "to": now().Format(time.RFC3339)}}, Actor{Platform: true})
		if err != nil {
			return nil, err
		}
		rows, _ := data["health"].([]PaymentHealthRow)
		statuses := make([]string, 0, len(rows))
		for _, row := range rows {
			statuses = append(statuses, row.OperationalStatus)
		}
		return statuses, nil
	}))
	components = append(components, s.externalAggregateHealth(ctx, "game-providers", "GAME_PROVIDER", func(checkCtx context.Context) ([]string, error) {
		data, err := New(s.db.WithContext(checkCtx)).ProviderMonitor(Query{Values: map[string]string{"from": now().Add(-time.Hour).Format(time.RFC3339), "to": now().Format(time.RFC3339)}}, Actor{Platform: true})
		if err != nil {
			return nil, err
		}
		rows, _ := data["list"].([]map[string]any)
		statuses := make([]string, 0, len(rows))
		for _, row := range rows {
			statuses = append(statuses, fmt.Sprint(row["healthStatus"]))
		}
		return statuses, nil
	}))
	for index := range components {
		stored, err := s.storeHealthComponent(components[index])
		if err == nil {
			components[index] = stored
		}
		s.syncHealthAlert(components[index])
	}
	overall := "HEALTHY"
	for _, component := range components {
		if component.Status == "UNHEALTHY" {
			overall = "UNHEALTHY"
			break
		}
		if component.Status == "DEGRADED" && overall == "HEALTHY" {
			overall = "DEGRADED"
		}
	}
	return map[string]any{"status": overall, "checkedAt": now(), "components": components}, nil
}

func (s *Service) rabbitMQHealth(parent context.Context) HealthComponent {
	start := time.Now()
	manager := eventbus.QueueManagerInstance()
	if manager == nil {
		status := "UNKNOWN"
		if config.RabbitMQConfig().Enabled {
			status = "UNHEALTHY"
		}
		return HealthComponent{
			ComponentCode: "rabbitmq", ComponentType: "MESSAGE_QUEUE", Status: status,
			LatencyMS: time.Since(start).Milliseconds(), ErrorSummary: "RabbitMQ adapter is not initialized", LastCheckedAt: now(),
		}
	}
	ctx, cancel := context.WithTimeout(parent, 4*time.Second)
	defer cancel()
	health := manager.Health(ctx)
	details := map[string]any{
		"connection": health.Connection, "publisherAvailable": health.PublisherAvailable, "consumerRunning": health.ConsumerRunning,
		"missingExchanges": health.MissingExchanges, "missingQueues": health.MissingQueues,
	}
	if metrics, err := manager.Metrics(ctx); err == nil {
		details["metrics"] = metrics
	}
	status := map[string]string{"UP": "HEALTHY", "DEGRADED": "DEGRADED", "DOWN": "UNHEALTHY"}[health.Status]
	if status == "" {
		status = "UNKNOWN"
	}
	errorSummary := ""
	if status != "HEALTHY" {
		errorSummary = sanitizeFailure(health.Message)
	}
	return HealthComponent{
		ComponentCode: "rabbitmq", ComponentType: "MESSAGE_QUEUE", Status: status,
		LatencyMS: time.Since(start).Milliseconds(), ErrorSummary: errorSummary, LastCheckedAt: now(), Details: details,
	}
}

func (s *Service) checkHealth(parent context.Context, code, kind string, check func(context.Context) (map[string]any, error)) HealthComponent {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	start := time.Now()
	details, err := check(ctx)
	summary := ""
	if err != nil {
		summary = sanitizeFailure(err.Error())
	}
	return HealthComponent{ComponentCode: code, ComponentType: kind, Status: healthStatus(err), LatencyMS: time.Since(start).Milliseconds(), ErrorSummary: summary, LastCheckedAt: now(), Details: details}
}

func (s *Service) databaseQueueHealth(code, kind, query string, degraded, unhealthy int64) HealthComponent {
	var count int64
	start := time.Now()
	err := s.db.Raw(query).Scan(&count).Error
	status := healthStatus(err)
	if err == nil && count > unhealthy {
		status = "UNHEALTHY"
	} else if err == nil && count > degraded {
		status = "DEGRADED"
	}
	return HealthComponent{ComponentCode: code, ComponentType: kind, Status: status, LatencyMS: time.Since(start).Milliseconds(), ErrorSummary: errorSummary(err), LastCheckedAt: now(), Details: map[string]any{"count": count}}
}

func (s *Service) schedulerHealth() HealthComponent {
	var failed, stale int64
	since := now().Add(-24 * time.Hour)
	_ = s.db.Table("game_system_task_executions").Where("status='FAILED' AND started_at>=? AND deleted_at IS NULL", since).Count(&failed).Error
	_ = s.db.Table("game_system_tasks").Where("running=1 AND updated_at<? AND deleted_at IS NULL", now().Add(-2*time.Hour)).Count(&stale).Error
	status := "HEALTHY"
	if stale > 0 {
		status = "UNHEALTHY"
	} else if failed > 0 {
		status = "DEGRADED"
	}
	return HealthComponent{ComponentCode: "scheduler", ComponentType: "SCHEDULER", Status: status, LastCheckedAt: now(), Details: map[string]any{"failures24h": failed, "staleRunning": stale}}
}

func (s *Service) workerHealth() HealthComponent {
	var running int64
	err := s.db.Table("game_system_tasks").Where("running=1 AND deleted_at IS NULL").Count(&running).Error
	return HealthComponent{ComponentCode: "background-workers", ComponentType: "WORKER", Status: healthStatus(err), ErrorSummary: errorSummary(err), LastCheckedAt: now(), Details: map[string]any{"running": running}}
}

func (s *Service) externalAggregateHealth(parent context.Context, code, kind string, load func(context.Context) ([]string, error)) HealthComponent {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	start := time.Now()
	type result struct {
		statuses []string
		err      error
	}
	completed := make(chan result, 1)
	go func() {
		statuses, err := load(ctx)
		completed <- result{statuses: statuses, err: err}
	}()
	var statuses []string
	var err error
	select {
	case value := <-completed:
		statuses, err = value.statuses, value.err
	case <-ctx.Done():
		err = NewError("HEALTH_CHECK_TIMEOUT", "external health aggregation timed out")
	}
	status := healthStatus(err)
	if err == nil {
		status = aggregateOperationalStatus(statuses)
	}
	return HealthComponent{ComponentCode: code, ComponentType: kind, Status: status, LatencyMS: time.Since(start).Milliseconds(), ErrorSummary: errorSummary(err), LastCheckedAt: now(), Details: map[string]any{"statuses": statuses}}
}

func aggregateOperationalStatus(statuses []string) string {
	if len(statuses) == 0 {
		return "UNKNOWN"
	}
	result := "HEALTHY"
	for _, status := range statuses {
		switch status {
		case "UNAVAILABLE", "UNHEALTHY":
			return "UNHEALTHY"
		case "DEGRADED", "MAINTENANCE":
			result = "DEGRADED"
		case "UNKNOWN":
			if result == "HEALTHY" {
				result = "UNKNOWN"
			}
		}
	}
	return result
}
func healthStatus(err error) string {
	if err != nil {
		return "UNHEALTHY"
	}
	return "HEALTHY"
}
func safeRedisPing(ctx context.Context) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("redis unavailable")
		}
	}()
	client := database.GetRedisCli()
	if client == nil {
		return fmt.Errorf("redis unavailable")
	}
	return client.Ping(ctx).Err()
}
func errorSummary(err error) string {
	if err == nil {
		return ""
	}
	return sanitizeFailure(err.Error())
}
func sanitizeFailure(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 500 {
		value = value[:500]
	}
	return fmt.Sprint(redact(map[string]any{"error": value}).(map[string]any)["error"])
}

func (s *Service) storeHealthComponent(component HealthComponent) (HealthComponent, error) {
	var before struct {
		ConsecutiveFailures uint
		Status              string
		LastRecoveredAt     *time.Time
	}
	_ = s.db.Table("game_system_health_checks").Select("consecutive_failures,status,last_recovered_at").Where("component_code=?", component.ComponentCode).Take(&before).Error
	component.ConsecutiveFailures = 0
	component.LastRecoveredAt = before.LastRecoveredAt
	if component.Status == "UNHEALTHY" || component.Status == "DEGRADED" {
		component.ConsecutiveFailures = before.ConsecutiveFailures + 1
	} else if before.Status == "UNHEALTHY" || before.Status == "DEGRADED" {
		recovered := now()
		component.LastRecoveredAt = &recovered
	}
	row := map[string]any{"component_code": component.ComponentCode, "component_type": component.ComponentType, "status": component.Status, "latency_ms": component.LatencyMS, "error_summary": component.ErrorSummary, "consecutive_failures": component.ConsecutiveFailures, "last_checked_at": component.LastCheckedAt, "last_recovered_at": component.LastRecoveredAt, "details": marshal(component.Details), "updated_at": now()}
	err := s.db.Table("game_system_health_checks").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "component_code"}}, DoUpdates: clause.AssignmentColumns([]string{"component_type", "status", "latency_ms", "error_summary", "consecutive_failures", "last_checked_at", "last_recovered_at", "details", "updated_at"})}).Create(row).Error
	return component, err
}

func (s *Service) syncHealthAlert(component HealthComponent) {
	fingerprint := alertFingerprint("SYSTEM_HEALTH", component.ComponentCode, 0, 0)
	if component.Status == "UNHEALTHY" || component.Status == "DEGRADED" {
		severity := "WARNING"
		if component.Status == "UNHEALTHY" {
			severity = "CRITICAL"
		}
		_ = s.raiseAlert(fingerprint, 0, 0, severity, "SYSTEM_HEALTH", component.ComponentCode, component.Status+" "+component.ComponentCode, map[string]any{"error": component.ErrorSummary, "latencyMS": component.LatencyMS})
	} else {
		_ = s.db.Table("game_system_alerts").Where("fingerprint=? AND status IN ('OPEN','ACKNOWLEDGED')", fingerprint).Updates(map[string]any{"status": "RESOLVED", "resolved_at": now(), "updated_at": now()}).Error
	}
}

func (s *Service) raiseAlert(fingerprint string, tenantID, siteID uint64, severity, alertType, source, summary string, details map[string]any) error {
	row := map[string]any{"fingerprint": fingerprint, "tenant_id": tenantID, "site_id": siteID, "severity": severity, "alert_type": alertType, "source": source, "status": "OPEN", "occurrence_count": 1, "summary": summary, "details": marshal(redact(details)), "first_triggered_at": now(), "last_triggered_at": now(), "created_at": now(), "updated_at": now()}
	return s.db.Table("game_system_alerts").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "fingerprint"}}, DoUpdates: clause.Assignments(map[string]any{"severity": severity, "status": "OPEN", "occurrence_count": gorm.Expr("occurrence_count+1"), "summary": summary, "details": marshal(redact(details)), "last_triggered_at": now(), "acknowledged_by": 0, "acknowledged_at": nil, "resolved_by": 0, "resolved_at": nil, "updated_at": now()})}).Create(row).Error
}

func alertFingerprint(alertType, source string, tenantID, siteID uint64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d|%d", alertType, source, tenantID, siteID)))
	return hex.EncodeToString(sum[:])
}

func (s *Service) ListAlerts(q Query, actor Actor) (Result, error) {
	rows := []map[string]any{}
	db := s.db.Table("game_system_alerts").Where("deleted_at IS NULL")
	var err error
	db, err = operationalScope(db, actor, value(q, "tenantID"), value(q, "siteID"))
	if err != nil {
		return Result{}, err
	}
	db = exact(db, "severity", strings.ToUpper(value(q, "severity")))
	db = exact(db, "status", strings.ToUpper(value(q, "status")))
	db = exact(db, "alert_type", strings.ToUpper(value(q, "alertType")))
	db = like(db, "source", value(q, "source"))
	return paginate(db.Order("CASE severity WHEN 'CRITICAL' THEN 1 WHEN 'WARNING' THEN 2 ELSE 3 END,last_triggered_at DESC"), q, &rows)
}

func (s *Service) AcknowledgeAlert(id uint64, input AlertActionInput, actor Actor) error {
	return s.alertAction(id, "ACKNOWLEDGED", input, actor)
}
func (s *Service) ResolveAlert(id uint64, input AlertActionInput, actor Actor) error {
	return s.alertAction(id, "RESOLVED", input, actor)
}
func (s *Service) alertAction(id uint64, status string, input AlertActionInput, actor Actor) error {
	if err := requireReason(input.Reason); err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var before map[string]any
		db, err := operationalScope(tx.Table("game_system_alerts"), actor, "", "")
		if err != nil {
			return err
		}
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
			return rowNotFound("alert")
		}
		updates := map[string]any{"status": status, "internal_note": strings.TrimSpace(input.Note), "updated_at": now()}
		if status == "ACKNOWLEDGED" {
			updates["acknowledged_by"] = actor.ID
			updates["acknowledged_at"] = now()
		} else {
			updates["resolved_by"] = actor.ID
			updates["resolved_at"] = now()
		}
		if err := tx.Table("game_system_alerts").Where("id=?", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, map[string]string{"ACKNOWLEDGED": "alerts:acknowledge", "RESOLVED": "alerts:resolve"}[status], "alert."+strings.ToLower(status), "system_alert", id, before, updates, input.Reason)
	})
}

func (s *Service) JobsOverview(q Query, actor Actor) (map[string]any, error) {
	tasks := Result{List: []map[string]any{}, Page: 1, PageSize: 20}
	var err error
	if actor.Platform {
		tasks, err = s.ListTasks(q)
		if err != nil {
			return nil, err
		}
	}
	executions := []map[string]any{}
	if actor.Platform {
		_ = s.db.Table("game_system_task_executions").Select("id executionID,execution_no executionNo,task_code taskCode,triggered_by triggeredBy,status,started_at startedAt,completed_at completedAt,duration_ms durationMS,error_message errorMessage").Where("status IN ('FAILED','RUNNING') AND deleted_at IS NULL").Order("id DESC").Limit(200).Find(&executions).Error
	}
	retries := []map[string]any{}
	retryDB := s.db.Table("game_system_retry_records").Where("deleted_at IS NULL")
	retryDB, err = operationalScope(retryDB, actor, "", "")
	if err != nil {
		return nil, err
	}
	_ = retryDB.Select("source_type,tenant_id,site_id,status,attempt_count,max_attempts,next_attempt_at,last_error,created_at,updated_at").Order("id DESC").Limit(200).Find(&retries).Error
	dead := []map[string]any{}
	deadDB := s.db.Table("event_outbox").Where("status='DEAD' AND deleted_at IS NULL")
	deadDB, err = operationalScope(deadDB, actor, "", "")
	if err != nil {
		return nil, err
	}
	_ = deadDB.Select("id,event_type eventType,aggregate_type aggregateType,status,attempts,last_error lastError,occurred_at occurredAt").Order("id DESC").Limit(200).Find(&dead).Error
	rabbitDead := []eventbus.DeadMessage{}
	queueMetrics := eventbus.QueueMetrics{}
	if manager := eventbus.QueueManagerInstance(); actor.Platform && manager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, queue := range []string{eventbus.QueueLedger, eventbus.QueuePayment, eventbus.QueueProvider, eventbus.QueuePromotion, eventbus.QueueNotification, eventbus.QueueAudit, eventbus.QueueDashboard, eventbus.QueueSystemJobs} {
			items, readErr := manager.DeadMessages(ctx, queue, 25)
			if readErr == nil {
				rabbitDead = append(rabbitDead, items...)
			}
		}
		queueMetrics, _ = manager.Metrics(ctx)
	}
	return map[string]any{"tasks": tasks, "failedExecutions": executions, "retryRecords": retries, "deadLetters": dead, "rabbitDeadLetters": rabbitDead, "queueMetrics": queueMetrics}, nil
}

func (s *Service) RetryTaskExecution(executionID uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row TaskExecutionRow
	retryKey := fmt.Sprintf("task-execution:%d", executionID)
	var existingRetry struct{ Status string }
	if err := s.db.Table("game_system_retry_records").Select("status").Where("idempotency_key=? AND deleted_at IS NULL", retryKey).Take(&existingRetry).Error; err == nil {
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_system_task_executions").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=? AND status='FAILED' AND deleted_at IS NULL", executionID).Take(&row).Error; err != nil {
			return rowNotFound("failed_task_execution")
		}
		if err := tx.Table("game_system_task_executions").Where("id=? AND status='FAILED'", executionID).Update("status", "RETRYING").Error; err != nil {
			return err
		}
		retry := map[string]any{"source_type": "TASK_EXECUTION", "source_id": executionID, "idempotency_key": retryKey, "status": "PROCESSING", "attempt_count": 1, "max_attempts": 1, "created_by": actor.ID, "created_at": now(), "updated_at": now()}
		return tx.Table("game_system_retry_records").Create(retry).Error
	}); err != nil {
		return duplicateError(err)
	}
	if err := s.runTask(row.TaskID, "RETRY", actor); err != nil {
		_ = s.db.Table("game_system_task_executions").Where("id=? AND status='RETRYING'", executionID).Update("status", "FAILED").Error
		_ = s.db.Table("game_system_retry_records").Where("idempotency_key=?", retryKey).Updates(map[string]any{"status": "FAILED", "last_error": sanitizeFailure(err.Error()), "updated_at": now()}).Error
		return err
	}
	_ = s.db.Table("game_system_task_executions").Where("id=? AND status='RETRYING'", executionID).Update("status", "RETRIED").Error
	_ = s.db.Table("game_system_retry_records").Where("idempotency_key=?", retryKey).Updates(map[string]any{"status": "SUCCESS", "updated_at": now()}).Error
	return audit(s.db, actor, "jobs:retry", "jobs.task.retry", "task_execution", executionID, row, map[string]any{"retried": true}, reason)
}

func (s *Service) RetryDeadLetter(eventID uint64, reason string, actor Actor) (any, error) {
	if err := requireReason(reason); err != nil {
		return nil, err
	}
	return s.ReplayEvent(eventID, EventReplayInput{Reason: reason}, actor)
}

func (s *Service) RetryRabbitDeadLetter(queue, eventID, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	if !actor.Platform {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	manager := eventbus.QueueManagerInstance()
	if manager == nil {
		return NewError("RABBITMQ_UNAVAILABLE", "RabbitMQ adapter is not initialized")
	}
	if err := manager.ReplayDead(context.Background(), strings.TrimSpace(queue), strings.TrimSpace(eventID)); err != nil {
		return err
	}
	return audit(s.db, actor, "jobs:retry", "jobs.rabbitmq.dead_letter.retry", "rabbitmq_message", 0, nil, map[string]any{"queue": queue, "eventID": eventID}, reason)
}

func (s *Service) PurgeRabbitRetry(queue, reason string, actor Actor) (int, error) {
	if err := requireReason(reason); err != nil {
		return 0, err
	}
	if !actor.Platform {
		return 0, NewError("PERMISSION_DENIED", "permission denied")
	}
	manager := eventbus.QueueManagerInstance()
	if manager == nil {
		return 0, NewError("RABBITMQ_UNAVAILABLE", "RabbitMQ adapter is not initialized")
	}
	count, err := manager.PurgeRetry(context.Background(), strings.TrimSpace(queue))
	if err != nil {
		return 0, err
	}
	if err := audit(s.db, actor, "jobs:retry", "jobs.rabbitmq.retry.purge", "rabbitmq_queue", 0, nil, map[string]any{"queue": queue, "purged": count}, reason); err != nil {
		return 0, err
	}
	return count, nil
}
