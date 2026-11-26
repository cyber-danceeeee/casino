package systemadmin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"casino/internal/content"
	"casino/internal/eventbus"
	"casino/internal/reportadmin"

	"gorm.io/gorm"
)

type TaskRow struct {
	ID              uint64          `json:"id"`
	TaskCode        string          `json:"taskCode" gorm:"column:task_code"`
	NameZh          string          `json:"nameZh" gorm:"column:name_zh"`
	NameEn          string          `json:"nameEn" gorm:"column:name_en"`
	TaskType        string          `json:"taskType" gorm:"column:task_type"`
	HandlerCode     string          `json:"handlerCode" gorm:"column:handler_code"`
	CronExpression  string          `json:"cronExpression" gorm:"column:cron_expression"`
	Payload         json.RawMessage `json:"payload"`
	TimeoutSeconds  uint            `json:"timeoutSeconds" gorm:"column:timeout_seconds"`
	Enabled         bool            `json:"enabled"`
	Running         bool            `json:"running"`
	SystemProtected bool            `json:"systemProtected" gorm:"column:system_protected"`
	LastRunAt       *time.Time      `json:"lastRunAt" gorm:"column:last_run_at"`
	LastRunResult   string          `json:"lastRunResult" gorm:"column:last_run_result"`
	NextRunAt       *time.Time      `json:"nextRunAt" gorm:"column:next_run_at"`
	UpdatedAt       time.Time       `json:"updatedAt" gorm:"column:updated_at"`
}
type TaskExecutionRow struct {
	ID            uint64     `json:"executionID"`
	ExecutionNo   string     `json:"executionNo" gorm:"column:execution_no"`
	TaskID        uint64     `json:"taskID" gorm:"column:task_id"`
	TaskCode      string     `json:"taskCode" gorm:"column:task_code"`
	TriggeredBy   string     `json:"triggeredBy" gorm:"column:triggered_by"`
	OperatorID    uint64     `json:"operatorID" gorm:"column:operator_id"`
	Status        string     `json:"status"`
	StartedAt     time.Time  `json:"startedAt" gorm:"column:started_at"`
	CompletedAt   *time.Time `json:"completedAt" gorm:"column:completed_at"`
	DurationMS    uint64     `json:"durationMS" gorm:"column:duration_ms"`
	ErrorMessage  string     `json:"errorMessage" gorm:"column:error_message"`
	CorrelationID string     `json:"correlationID" gorm:"column:correlation_id"`
}

func (s *Service) ListTasks(q Query) (Result, error) {
	rows := []TaskRow{}
	db := s.db.Table("game_system_tasks").Where("deleted_at IS NULL")
	db = like(db, "task_code", value(q, "taskCode"))
	name := value(q, "taskName")
	if name != "" {
		db = db.Where("(name_zh LIKE ? OR name_en LIKE ?)", "%"+name+"%", "%"+name+"%")
	}
	db = exact(db, "task_type", value(q, "taskType"))
	db = exact(db, "enabled", value(q, "status"))
	db = exact(db, "last_run_result", value(q, "lastRunResult"))
	db = timeRange(db, "next_run_at", value(q, "nextRunFrom"), value(q, "nextRunTo"))
	return paginate(db.Order("task_code"), q, &rows)
}
func registeredTaskHandlers() map[string]bool {
	return map[string]bool{"report.daily_aggregation": true, "system.session_cleanup": true, "operation.expired_bonus": true}
}
func validateTask(input TaskInput, creating bool) error {
	if creating {
		if err := validateCode(input.TaskCode); err != nil {
			return err
		}
	}
	if input.NameZh == "" || input.NameEn == "" {
		return NewError("TASK_NAME_REQUIRED", "task names are required")
	}
	if !registeredTaskHandlers()[input.HandlerCode] {
		return NewError("UNREGISTERED_TASK_HANDLER", "task handler is not registered")
	}
	if _, err := nextCronTime(input.CronExpression, now()); err != nil {
		return err
	}
	if input.TimeoutSeconds == 0 || input.TimeoutSeconds > 3600 {
		return NewError("INVALID_TASK_TIMEOUT", "timeout must be between 1 and 3600 seconds")
	}
	return nil
}
func taskMap(input TaskInput, creating bool) (map[string]any, error) {
	next, err := nextCronTime(input.CronExpression, now())
	if err != nil {
		return nil, err
	}
	row := map[string]any{"name_zh": input.NameZh, "name_en": input.NameEn, "task_type": input.TaskType, "handler_code": input.HandlerCode, "cron_expression": input.CronExpression, "payload": marshal(input.Payload), "timeout_seconds": input.TimeoutSeconds, "enabled": boolValue(boolDefault(input.Enabled, true)), "next_run_at": next, "updated_at": now()}
	if creating {
		row["task_code"] = input.TaskCode
		row["created_at"] = now()
	}
	return row, nil
}
func (s *Service) CreateTask(input TaskInput, actor Actor) (uint64, error) {
	if err := validateTask(input, true); err != nil {
		return 0, err
	}
	row, _ := taskMap(input, true)
	if err := s.db.Table("game_system_tasks").Create(row).Error; err != nil {
		return 0, duplicateError(err)
	}
	id := row["id"].(uint64)
	if err := audit(s.db, actor, "system:task:create", "system.task.create", "scheduled_task", id, nil, row, ""); err != nil {
		return 0, err
	}
	return id, nil
}
func (s *Service) UpdateTask(id uint64, input TaskInput, reason string, actor Actor) error {
	if err := validateTask(input, false); err != nil {
		return err
	}
	var before map[string]any
	if err := s.db.Table("game_system_tasks").Where("id=? AND deleted_at IS NULL", id).Take(&before).Error; err != nil {
		return rowNotFound("scheduled task")
	}
	updates, _ := taskMap(input, false)
	delete(updates, "enabled")
	updates["version"] = gorm.Expr("version+1")
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_system_tasks").Where("id=? AND running=0", id).Updates(updates).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:task:update", "system.task.update", "scheduled_task", id, before, updates, reason)
	})
}
func (s *Service) TaskStatus(id uint64, enabled bool, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row TaskRow
	if err := s.db.Table("game_system_tasks").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("scheduled task")
	}
	next := row.NextRunAt
	if enabled {
		value, err := nextCronTime(row.CronExpression, now())
		if err != nil {
			return err
		}
		next = &value
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_system_tasks").Where("id=? AND running=0", id).Updates(map[string]any{"enabled": boolValue(enabled), "next_run_at": next, "updated_at": now()}).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:task:update", "system.task.status", "scheduled_task", id, map[string]any{"enabled": row.Enabled}, map[string]any{"enabled": enabled}, reason)
	})
}
func (s *Service) DeleteTask(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	var row TaskRow
	if err := s.db.Table("game_system_tasks").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return rowNotFound("scheduled task")
	}
	if row.SystemProtected {
		return NewError("PROTECTED_TASK", "system-protected task cannot be deleted")
	}
	if row.Running {
		return NewError("TASK_RUNNING", "running task cannot be deleted")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_system_tasks").Where("id=?", id).Update("deleted_at", now()).Error; err != nil {
			return err
		}
		return audit(tx, actor, "system:task:delete", "system.task.delete", "scheduled_task", id, row, nil, reason)
	})
}
func (s *Service) TaskDetail(id uint64) (map[string]any, error) {
	var row TaskRow
	if err := s.db.Table("game_system_tasks").Where("id=? AND deleted_at IS NULL", id).Take(&row).Error; err != nil {
		return nil, rowNotFound("scheduled task")
	}
	executions := []TaskExecutionRow{}
	_ = s.db.Table("game_system_task_executions").Where("task_id=? AND deleted_at IS NULL", id).Order("started_at DESC").Limit(100).Find(&executions).Error
	return map[string]any{"overview": row, "executionLogs": executions, "registeredHandlers": sortedHandlerCodes(), "operationLogs": s.operationLogs("scheduled_task", id)}, nil
}
func sortedHandlerCodes() []string {
	out := []string{}
	for code := range registeredTaskHandlers() {
		out = append(out, code)
	}
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
func (s *Service) ListTaskExecutions(id uint64, q Query) (Result, error) {
	rows := []TaskExecutionRow{}
	db := s.db.Table("game_system_task_executions").Where("task_id=? AND deleted_at IS NULL", id)
	db = exact(db, "status", value(q, "status"))
	db = timeRange(db, "started_at", value(q, "startedFrom"), value(q, "startedTo"))
	return paginate(db.Order("started_at DESC,id DESC"), q, &rows)
}
func (s *Service) RunTaskNow(id uint64, reason string, actor Actor) error {
	if err := requireReason(reason); err != nil {
		return err
	}
	if err := s.runTask(id, "MANUAL", actor); err != nil {
		return err
	}
	return audit(s.db, actor, "system:task:run", "system.task.run", "scheduled_task", id, nil, map[string]any{"triggered": true}, reason)
}

func (s *Service) runTask(id uint64, triggeredBy string, actor Actor) error {
	result := s.db.Table("game_system_tasks").Where("id=? AND enabled=1 AND running=0 AND deleted_at IS NULL", id).Updates(map[string]any{"running": 1, "updated_at": now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return NewError("TASK_UNAVAILABLE", "task is disabled, missing, or already running")
	}
	var task TaskRow
	if err := s.db.Table("game_system_tasks").Where("id=?", id).Take(&task).Error; err != nil {
		_ = s.db.Table("game_system_tasks").Where("id=?", id).Update("running", 0).Error
		return err
	}
	executionNo := fmt.Sprintf("TASK%d", time.Now().UnixNano())
	startedAt := now()
	execution := TaskExecutionRow{
		ExecutionNo: executionNo, TaskID: task.ID, TaskCode: task.TaskCode,
		TriggeredBy: triggeredBy, OperatorID: actor.ID, Status: "QUEUED",
		StartedAt: startedAt, CorrelationID: actor.CorrelationID,
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("game_system_task_executions").Create(&execution).Error; err != nil {
			return err
		}
		_, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
			Scope: content.Scope{Platform: true}, EventType: eventbus.EventSystemTaskRequested,
			AggregateType: "system_task_execution", AggregateID: fmt.Sprint(execution.ID), Source: "systemadmin", Producer: "systemadmin",
			IdempotencyKey: "system-task-execution:" + execution.ExecutionNo, RequestID: actor.CorrelationID,
			TraceID: actor.CorrelationID, CorrelationID: actor.CorrelationID, TargetConsumers: []string{eventbus.ConsumerSystemTask},
			Payload: map[string]any{"taskID": task.ID, "executionID": execution.ID},
		})
		return err
	}); err != nil {
		_ = s.db.Table("game_system_tasks").Where("id=?", id).Update("running", 0).Error
		return err
	}
	eventbus.SignalDispatch()
	return nil
}
func (s *Service) executeTask(task TaskRow, executionID uint64) error {
	started := now()
	timeout := time.Duration(task.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	errChannel := make(chan error, 1)
	go func() { errChannel <- s.invokeTaskHandler(ctx, task) }()
	var err error
	select {
	case err = <-errChannel:
	case <-ctx.Done():
		err = NewError("TASK_TIMEOUT", "task execution timed out")
	}
	completed := now()
	status, message := "SUCCESS", ""
	if err != nil {
		status, message = "FAILED", err.Error()
		if len(message) > 1000 {
			message = message[:1000]
		}
	}
	next, nextErr := nextCronTime(task.CronExpression, completed)
	if nextErr != nil {
		next = completed.Add(time.Hour)
	}
	persistErr := s.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Table("game_system_task_executions").Where("id=?", executionID).Updates(map[string]any{"status": status, "completed_at": completed, "duration_ms": completed.Sub(started).Milliseconds(), "error_message": message, "updated_at": completed}).Error; e != nil {
			return e
		}
		return tx.Table("game_system_tasks").Where("id=?", task.ID).Updates(map[string]any{"running": 0, "last_run_at": completed, "last_run_result": status, "next_run_at": next, "updated_at": completed}).Error
	})
	if persistErr != nil {
		return persistErr
	}
	return err
}

func (s *Service) HandleSystemTaskEvent(ctx eventbus.Context) error {
	taskID := uint64FromJSON(ctx.Event.Payload["taskID"])
	executionID := uint64FromJSON(ctx.Event.Payload["executionID"])
	if taskID == 0 || executionID == 0 {
		return NewError("TASK_EXECUTION_REQUIRED", "system task execution is required")
	}
	var task TaskRow
	if err := s.db.Table("game_system_tasks").Where("id=? AND deleted_at IS NULL", taskID).Take(&task).Error; err != nil {
		return err
	}
	var execution TaskExecutionRow
	if err := s.db.Table("game_system_task_executions").Where("id=? AND task_id=? AND deleted_at IS NULL", executionID, taskID).Take(&execution).Error; err != nil {
		return err
	}
	if execution.Status == "SUCCESS" {
		return nil
	}
	restarted := now()
	staleBefore := restarted.Add(-time.Duration(task.TimeoutSeconds)*time.Second - time.Minute)
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if execution.Status == "FAILED" {
			result := tx.Table("game_system_tasks").Where("id=? AND running=0", taskID).Updates(map[string]any{"running": 1, "updated_at": restarted})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return NewError("TASK_ALREADY_RUNNING", "scheduled task is already running")
			}
		}
		result := tx.Table("game_system_task_executions").Where("id=? AND (status IN ('QUEUED','FAILED') OR (status='RUNNING' AND started_at<?))", executionID, staleBefore).Updates(map[string]any{"status": "RUNNING", "started_at": restarted, "completed_at": nil, "error_message": "", "updated_at": restarted})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return NewError("TASK_EXECUTION_ALREADY_RUNNING", "system task execution is already running")
		}
		return nil
	}); err != nil {
		return err
	}
	return s.executeTask(task, executionID)
}
func (s *Service) invokeTaskHandler(ctx context.Context, task TaskRow) error {
	payload := map[string]any{}
	_ = json.Unmarshal(task.Payload, &payload)
	db := s.db.WithContext(ctx)
	switch task.HandlerCode {
	case "report.daily_aggregation":
		timezone, _ := payload["timezone"].(string)
		if timezone == "" {
			timezone = "+08:00"
		}
		date, _ := payload["businessDate"].(string)
		if date == "" {
			date = now().AddDate(0, 0, -1).Format("2006-01-02")
		}
		return reportadmin.New(db).RefreshDaily(date, timezone)
	case "system.session_cleanup":
		return db.Table("game_admin_sessions").Where("expired_at<=? AND revoked_at IS NULL AND deleted_at IS NULL", now()).Updates(map[string]any{"revoked_at": now(), "updated_at": now()}).Error
	case "operation.expired_bonus":
		return db.Table("game_bonus_records").Where("expires_at<=? AND status_code IN ('PENDING','ACTIVE') AND deleted_at IS NULL", now()).Updates(map[string]any{"status_code": "EXPIRED", "updated_at": now()}).Error
	default:
		return NewError("UNREGISTERED_TASK_HANDLER", "task handler is not registered")
	}
}

var schedulerOnce sync.Once

func StartScheduler(service *Service) {
	schedulerOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for current := range ticker.C {
				ids := []uint64{}
				_ = service.db.Table("game_system_tasks").Where("enabled=1 AND running=0 AND next_run_at<=? AND deleted_at IS NULL", current).Limit(20).Pluck("id", &ids).Error
				for _, id := range ids {
					_ = service.runTask(id, "SCHEDULE", Actor{Username: "system-scheduler"})
				}
			}
		}()
	})
}

func nextCronTime(expression string, after time.Time) (time.Time, error) {
	fields := strings.Fields(expression)
	if len(fields) != 5 {
		return time.Time{}, NewError("INVALID_CRON", "cron expression must have five fields")
	}
	for index, spec := range fields {
		min, max := cronBounds(index)
		if !validCronField(spec, min, max) {
			return time.Time{}, NewError("INVALID_CRON", "invalid cron expression")
		}
	}
	candidate := after.Truncate(time.Minute).Add(time.Minute)
	limit := candidate.AddDate(1, 0, 0)
	for candidate.Before(limit) {
		if cronMatch(fields[0], candidate.Minute(), 0, 59) && cronMatch(fields[1], candidate.Hour(), 0, 23) &&
			cronMatch(fields[2], candidate.Day(), 1, 31) && cronMatch(fields[3], int(candidate.Month()), 1, 12) &&
			cronMatch(fields[4], int(candidate.Weekday()), 0, 6) {
			return candidate, nil
		}
		candidate = candidate.Add(time.Minute)
	}
	return time.Time{}, NewError("INVALID_CRON", "cron expression has no execution within one year")
}
func cronBounds(index int) (int, int) {
	switch index {
	case 0:
		return 0, 59
	case 1:
		return 0, 23
	case 2:
		return 1, 31
	case 3:
		return 1, 12
	default:
		return 0, 6
	}
}
func validCronField(spec string, min int, max int) bool {
	if spec == "*" {
		return true
	}
	for _, part := range strings.Split(spec, ",") {
		base, step := part, 1
		if strings.Contains(part, "/") {
			pieces := strings.Split(part, "/")
			if len(pieces) != 2 {
				return false
			}
			base = pieces[0]
			value, err := strconv.Atoi(pieces[1])
			if err != nil || value < 1 {
				return false
			}
			step = value
		}
		_ = step
		if base == "*" {
			continue
		}
		if strings.Contains(base, "-") {
			pieces := strings.Split(base, "-")
			if len(pieces) != 2 {
				return false
			}
			from, e1 := strconv.Atoi(pieces[0])
			to, e2 := strconv.Atoi(pieces[1])
			if e1 != nil || e2 != nil || from < min || to > max || from > to {
				return false
			}
			continue
		}
		value, err := strconv.Atoi(base)
		if err != nil || value < min || value > max {
			return false
		}
	}
	return true
}
func cronMatch(spec string, value int, min int, max int) bool {
	if !validCronField(spec, min, max) {
		return false
	}
	for _, part := range strings.Split(spec, ",") {
		base, step := part, 1
		if strings.Contains(part, "/") {
			pieces := strings.Split(part, "/")
			base = pieces[0]
			step, _ = strconv.Atoi(pieces[1])
		}
		from, to := min, max
		if base != "*" {
			if strings.Contains(base, "-") {
				pieces := strings.Split(base, "-")
				from, _ = strconv.Atoi(pieces[0])
				to, _ = strconv.Atoi(pieces[1])
			} else {
				from, _ = strconv.Atoi(base)
				to = from
			}
		}
		if value >= from && value <= to && (value-from)%step == 0 {
			return true
		}
	}
	return false
}
