package notificationpush

import (
	"context"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/go-dev-frame/sponge/pkg/logger"
	"google.golang.org/api/option"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/config"
	"casino/internal/eventbus"
	"casino/internal/model"
	"casino/internal/notificationaudience"
	"casino/internal/notificationchannel"
	"casino/internal/tenancy"
)

const maxFCMBatchSize = 500

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

type Worker struct {
	db      *gorm.DB
	config  config.FCM
	client  *messaging.Client
	initErr error
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type audienceUser struct {
	ID       uint64 `gorm:"column:id"`
	Language string `gorm:"column:language"`
}

type pushTarget struct {
	UserID        uint64
	Language      string
	DeviceTokenID uint64
	Token         string
}

func EnsureSchema(db *gorm.DB) error {
	return db.AutoMigrate(&model.GamePushDeviceToken{}, &model.GameNotificationDelivery{})
}

func New(db *gorm.DB, cfg config.FCM) *Worker {
	w := &Worker{db: db, config: cfg}
	if !cfg.Enabled {
		return w
	}
	if strings.TrimSpace(cfg.ProjectID) == "" {
		w.initErr = errors.New("FCM project ID is required")
		return w
	}
	if strings.TrimSpace(cfg.CredentialsFile) == "" {
		w.initErr = errors.New("FCM service account credentials file is required")
		return w
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.ProjectID}, option.WithCredentialsFile(cfg.CredentialsFile))
	if err != nil {
		w.initErr = fmt.Errorf("initialize Firebase app: %w", err)
		return w
	}
	w.client, w.initErr = app.Messaging(ctx)
	if w.initErr != nil {
		w.initErr = fmt.Errorf("initialize Firebase messaging client: %w", w.initErr)
	}
	return w
}

func (w *Worker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		interval := time.Duration(w.config.WorkerPollIntervalSeconds) * time.Second
		if interval <= 0 {
			interval = 5 * time.Second
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		w.enqueueAvailable(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.enqueueAvailable(ctx)
			}
		}
	}()
}

func (w *Worker) enqueueAvailable(ctx context.Context) {
	for processed := 0; processed < 50; processed++ {
		queued := false
		err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var job model.GameNotificationJob
			result := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
				Where("status='PENDING' AND deleted_at IS NULL AND (scheduled_at IS NULL OR scheduled_at<=?)", time.Now()).
				Order("id ASC").Limit(1).Take(&job)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			if result.Error != nil {
				return result.Error
			}
			_, err := eventbus.New(tx).PublishTx(tx, eventbus.Event{
				Scope: tenancy.Scope{TenantID: job.TenantID, SiteID: job.SiteID}, TenantID: job.TenantID, SiteID: job.SiteID,
				EventType: eventbus.EventNotificationSendRequested, AggregateType: "notification_job", AggregateID: fmt.Sprint(job.ID),
				Source: "notificationpush", Producer: "notificationpush", IdempotencyKey: fmt.Sprintf("notification-send:%d", job.ID),
				TargetConsumers: []string{eventbus.ConsumerNotificationPush}, Payload: map[string]any{"jobID": job.ID},
			})
			if err != nil {
				return err
			}
			if err := tx.Model(&model.GameNotificationJob{}).Where("id=? AND status='PENDING'", job.ID).Updates(map[string]any{"status": "QUEUED", "last_error": ""}).Error; err != nil {
				return err
			}
			queued = true
			return nil
		})
		if err != nil {
			logger.Errorf("[notification-scheduler] enqueue job: %v", err)
			return
		}
		if !queued {
			return
		}
		eventbus.SignalDispatch()
	}
}

func (w *Worker) HandleEvent(ctx eventbus.Context) error {
	jobID := eventPayloadUint64(ctx.Event.Payload["jobID"])
	if jobID == 0 {
		return errors.New("notification job is required")
	}
	var job model.GameNotificationJob
	if err := w.db.Where("id=? AND tenant_id=? AND site_id=? AND deleted_at IS NULL", jobID, ctx.Event.TenantID, ctx.Event.SiteID).Take(&job).Error; err != nil {
		return err
	}
	if job.Status == "SUCCESS" || job.Status == "PARTIAL_FAILED" || job.Status == "CANCELLED" {
		return nil
	}
	result := w.db.Model(&model.GameNotificationJob{}).
		Where("id=? AND (status IN ('QUEUED','PENDING','FAILED') OR (status='PROCESSING' AND updated_at<?))", job.ID, time.Now().Add(-10*time.Minute)).
		Updates(map[string]any{"status": "PROCESSING", "attempts": gorm.Expr("attempts+1"), "last_error": ""})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("notification job is already processing")
	}
	job.Status = "PROCESSING"
	return w.processJob(context.Background(), job)
}

func eventPayloadUint64(value any) uint64 {
	switch typed := value.(type) {
	case uint64:
		return typed
	case int:
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

// InitializationError reports why Firebase Messaging is unavailable without
// making the rest of the casino service depend on the push provider.
func (w *Worker) InitializationError() error {
	return w.initErr
}

func (w *Worker) Stop() error {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	return nil
}

func (w *Worker) processAvailable(ctx context.Context) {
	_ = w.db.WithContext(ctx).Model(&model.GameNotificationJob{}).
		Where("status='PROCESSING' AND updated_at<?", time.Now().Add(-10*time.Minute)).
		Updates(map[string]any{"status": "PENDING", "last_error": "worker lease expired"}).Error

	for processed := 0; processed < 10; processed++ {
		job, ok, err := w.claimJob(ctx)
		if err != nil {
			logger.Errorf("[notification-worker] claim job: %v", err)
			return
		}
		if !ok {
			return
		}
		if err := w.processJob(ctx, job); err != nil {
			logger.Errorf("[notification-worker] job %d failed: %v", job.ID, err)
		}
	}
}

func (w *Worker) claimJob(ctx context.Context) (model.GameNotificationJob, bool, error) {
	var job model.GameNotificationJob
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status='PENDING' AND deleted_at IS NULL AND (scheduled_at IS NULL OR scheduled_at<=?)", time.Now()).
			Order("id ASC").Limit(1).Take(&job)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		if result.Error != nil {
			return result.Error
		}
		return tx.Model(&model.GameNotificationJob{}).Where("id=? AND status='PENDING'", job.ID).
			Updates(map[string]any{"status": "PROCESSING", "attempts": gorm.Expr("attempts+1"), "last_error": ""}).Error
	})
	return job, job.ID > 0, err
}

func (w *Worker) processJob(ctx context.Context, job model.GameNotificationJob) error {
	var notification model.GameNotification
	if err := w.db.WithContext(ctx).Where("id=? AND deleted_at IS NULL", job.NotificationID).Take(&notification).Error; err != nil {
		return w.finishWithError(ctx, job, "NOTIFICATION_NOT_FOUND", err)
	}
	if notification.Status == "CANCELLED" {
		return w.db.WithContext(ctx).Model(&model.GameNotificationJob{}).Where("id=?", job.ID).Update("status", "CANCELLED").Error
	}

	users, err := w.audienceUsers(ctx, notification)
	if err != nil {
		return w.finishWithError(ctx, job, "AUDIENCE_RESOLUTION_FAILED", err)
	}
	hasInApp := notificationchannel.Contains(notification.Channel, "IN_APP")
	hasFCM := notificationchannel.Contains(notification.Channel, "FCM")
	hasWebPush := notificationchannel.Contains(notification.Channel, "WEB_PUSH")
	if hasInApp && !hasFCM && !hasWebPush {
		return w.finish(ctx, job, notification, len(users), 0, "", "SUCCESS", "SENT")
	}
	if !hasFCM && hasWebPush {
		if hasInApp {
			return w.finish(ctx, job, notification, len(users), 0, "WEB_PUSH_NOT_CONFIGURED", "PARTIAL_FAILED", "PARTIAL_FAILED")
		}
		return w.finishWithError(ctx, job, "UNSUPPORTED_PUSH_CHANNEL", fmt.Errorf("channel %s has no sender", notification.Channel))
	}
	if !hasFCM {
		return w.finishWithError(ctx, job, "UNSUPPORTED_PUSH_CHANNEL", fmt.Errorf("channel %s has no sender", notification.Channel))
	}
	if w.initErr != nil {
		return w.finishWithError(ctx, job, "FCM_NOT_CONFIGURED", w.initErr)
	}

	targets, err := w.pushTargets(ctx, notification, users)
	if err != nil {
		return w.finishWithError(ctx, job, "DEVICE_RESOLUTION_FAILED", err)
	}
	existingSuccess, err := w.successfulUserIDs(ctx, notification)
	if err != nil {
		return w.finishWithError(ctx, job, "DELIVERY_RESOLUTION_FAILED", err)
	}

	targetByUser := make(map[uint64]pushTarget, len(targets))
	for _, target := range targets {
		targetByUser[target.UserID] = target
	}
	successCount := len(existingSuccess)
	failureCount := 0
	pending := make([]pushTarget, 0, len(users))
	for _, user := range users {
		if existingSuccess[user.ID] {
			continue
		}
		target, ok := targetByUser[user.ID]
		if !ok {
			failureCount++
			_ = w.saveDelivery(ctx, notification, job.ID, user.ID, 0, "FAILED", "", "NO_ACTIVE_DEVICE", "no active FCM device", nil)
			continue
		}
		pending = append(pending, target)
	}

	groups := map[string][]pushTarget{"zh": {}, "en": {}}
	for _, target := range pending {
		language := "en"
		if strings.HasPrefix(strings.ToLower(target.Language), "zh") {
			language = "zh"
		}
		groups[language] = append(groups[language], target)
	}

	lastError := ""
	for language, group := range groups {
		for start := 0; start < len(group); start += w.batchSize() {
			end := start + w.batchSize()
			if end > len(group) {
				end = len(group)
			}
			succeeded, failed, batchError := w.sendBatch(ctx, notification, job, language, group[start:end])
			successCount += succeeded
			failureCount += failed
			if batchError != "" {
				lastError = batchError
			}
		}
	}

	jobStatus, notificationStatus := "SUCCESS", "SENT"
	if failureCount > 0 && successCount > 0 {
		jobStatus, notificationStatus = "PARTIAL_FAILED", "PARTIAL_FAILED"
	} else if failureCount > 0 {
		jobStatus, notificationStatus = "FAILED", "FAILED"
	}
	if hasWebPush && notificationStatus == "SENT" {
		lastError = "WEB_PUSH_NOT_CONFIGURED"
		jobStatus, notificationStatus = "PARTIAL_FAILED", "PARTIAL_FAILED"
	}
	return w.finish(ctx, job, notification, successCount, failureCount, lastError, jobStatus, notificationStatus)
}

func (w *Worker) audienceUsers(ctx context.Context, notification model.GameNotification) ([]audienceUser, error) {
	scope := tenancy.Scope{TenantID: notification.TenantID, SiteID: notification.SiteID}
	query, err := notificationaudience.UsersScoped(w.db.WithContext(ctx), scope, notification.AudienceType, notification.AudienceValue)
	if err != nil {
		return nil, err
	}
	var users []audienceUser
	settingsJoin := "LEFT JOIN game_user_settings AS us ON us.user_id=u.id AND us.deleted_at IS NULL"
	settingsArgs := []any{}
	if scope.TenantScoped() {
		settingsJoin += " AND us.tenant_id = ?"
		settingsArgs = append(settingsArgs, scope.TenantID)
	}
	if scope.SiteScoped() {
		settingsJoin += " AND us.site_id IN (0, ?)"
		settingsArgs = append(settingsArgs, scope.SiteID)
	}
	err = query.Select("u.id,COALESCE(NULLIF(us.language,''),NULLIF(u.language,''),'en') AS language").
		Joins(settingsJoin, settingsArgs...).
		Where("u.status=1 AND COALESCE(us.push_enabled,1)=1").
		Order("u.id ASC").Scan(&users).Error
	return users, err
}

func (w *Worker) pushTargets(ctx context.Context, notification model.GameNotification, users []audienceUser) ([]pushTarget, error) {
	if len(users) == 0 {
		return nil, nil
	}
	scope := tenancy.Scope{TenantID: notification.TenantID, SiteID: notification.SiteID}
	languageByUser := make(map[uint64]string, len(users))
	ids := make([]uint64, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
		languageByUser[user.ID] = user.Language
	}

	targets := make([]pushTarget, 0, len(users))
	for start := 0; start < len(ids); start += 1000 {
		end := start + 1000
		if end > len(ids) {
			end = len(ids)
		}
		var tokens []model.GamePushDeviceToken
		subquery := tenancy.ApplySite(w.db.WithContext(ctx).Model(&model.GamePushDeviceToken{}), "", scope).
			Select("user_id,MAX(id) AS id").
			Where("user_id IN ? AND enabled=1 AND invalidated_at IS NULL AND deleted_at IS NULL", ids[start:end]).
			Group("user_id")
		if err := tenancy.ApplySite(w.db.WithContext(ctx).Table("game_push_device_tokens AS t"), "t", scope).
			Joins("JOIN (?) AS latest ON latest.id=t.id", subquery).
			Where("t.deleted_at IS NULL").Scan(&tokens).Error; err != nil {
			return nil, err
		}
		for _, token := range tokens {
			targets = append(targets, pushTarget{UserID: token.UserID, Language: languageByUser[token.UserID], DeviceTokenID: token.ID, Token: token.Token})
		}
	}
	return targets, nil
}

func (w *Worker) successfulUserIDs(ctx context.Context, notification model.GameNotification) (map[uint64]bool, error) {
	var rows []model.GameNotificationDelivery
	scope := tenancy.Scope{TenantID: notification.TenantID, SiteID: notification.SiteID}
	err := tenancy.ApplySite(w.db.WithContext(ctx).Model(&model.GameNotificationDelivery{}), "", scope).
		Select("user_id").
		Where("notification_id=? AND status='SUCCESS' AND deleted_at IS NULL", notification.ID).
		Find(&rows).Error
	result := make(map[uint64]bool, len(rows))
	for _, row := range rows {
		result[row.UserID] = true
	}
	return result, err
}

func (w *Worker) sendBatch(ctx context.Context, notification model.GameNotification, job model.GameNotificationJob, language string, targets []pushTarget) (int, int, string) {
	if len(targets) == 0 {
		return 0, 0, ""
	}
	title, body := notification.TitleEn, notification.ContentEn
	if language == "zh" {
		title, body = notification.TitleZh, notification.ContentZh
	}
	title = truncateText(plainText(title), 120)
	body = truncateText(plainText(body), 240)
	tokens := make([]string, 0, len(targets))
	now := time.Now()
	for _, target := range targets {
		tokens = append(tokens, target.Token)
		_ = w.saveDelivery(ctx, notification, job.ID, target.UserID, target.DeviceTokenID, "PENDING", "", "", "", &now)
	}

	message := &messaging.MulticastMessage{
		Tokens:       tokens,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data: map[string]string{
			"notificationId": notification.NotificationID,
			"path":           defaultString(w.config.DefaultPath, "/support?tab=notice"),
			"title":          title,
			"body":           body,
			"type":           "ADMIN_NOTIFICATION",
		},
		Android: &messaging.AndroidConfig{
			Priority:              "high",
			RestrictedPackageName: w.config.AndroidPackageName,
			Notification: &messaging.AndroidNotification{
				ChannelID:    defaultString(w.config.AndroidChannelID, "casino_rewards"),
				DefaultSound: true,
				Priority:     messaging.PriorityHigh,
			},
		},
	}
	response, err := w.client.SendEachForMulticast(ctx, message)
	if err != nil {
		for _, target := range targets {
			_ = w.saveDelivery(ctx, notification, job.ID, target.UserID, target.DeviceTokenID, "FAILED", "", "FCM_BATCH_FAILED", err.Error(), &now)
		}
		return 0, len(targets), err.Error()
	}

	succeeded, failed := 0, 0
	lastError := ""
	for index, result := range response.Responses {
		target := targets[index]
		if result.Success {
			succeeded++
			sentAt := time.Now()
			_ = w.saveDelivery(ctx, notification, job.ID, target.UserID, target.DeviceTokenID, "SUCCESS", result.MessageID, "", "", &sentAt)
			continue
		}
		failed++
		code := fcmErrorCode(result.Error)
		lastError = result.Error.Error()
		_ = w.saveDelivery(ctx, notification, job.ID, target.UserID, target.DeviceTokenID, "FAILED", "", code, result.Error.Error(), &now)
		if messaging.IsUnregistered(result.Error) || messaging.IsInvalidArgument(result.Error) {
			invalidatedAt := time.Now()
			_ = tenancy.ApplySite(w.db.WithContext(ctx).Model(&model.GamePushDeviceToken{}), "", tenancy.Scope{TenantID: notification.TenantID, SiteID: notification.SiteID}).Where("id=?", target.DeviceTokenID).
				Updates(map[string]any{"enabled": false, "invalidated_at": &invalidatedAt, "last_error": truncateText(result.Error.Error(), 500)}).Error
		}
	}
	return succeeded, failed, lastError
}

func (w *Worker) saveDelivery(ctx context.Context, notification model.GameNotification, jobID, userID, deviceTokenID uint64, status, messageID, errorCode, errorMessage string, at *time.Time) error {
	delivery := model.GameNotificationDelivery{TenantID: notification.TenantID, SiteID: notification.SiteID, NotificationID: notification.ID, JobID: jobID, UserID: userID, DeviceTokenID: deviceTokenID, Status: status, ProviderMessageID: messageID, ErrorCode: errorCode, ErrorMessage: truncateText(errorMessage, 500), AttemptedAt: at}
	if status == "SUCCESS" {
		delivery.SentAt = at
	}
	return w.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "notification_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"job_id", "device_token_id", "status", "provider_message_id", "error_code", "error_message", "attempted_at", "sent_at", "updated_at"}),
	}).Create(&delivery).Error
}

func (w *Worker) finishWithError(ctx context.Context, job model.GameNotificationJob, code string, err error) error {
	message := code + ": " + err.Error()
	if updateErr := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if result := tx.Model(&model.GameNotificationJob{}).Where("id=?", job.ID).Updates(map[string]any{"status": "FAILED", "last_error": truncateText(message, 500)}); result.Error != nil {
			return result.Error
		}
		return tx.Model(&model.GameNotification{}).Where("id=?", job.NotificationID).Updates(map[string]any{"status": "FAILED", "failure_count": gorm.Expr("target_user_count"), "sent_at": time.Now()}).Error
	}); updateErr != nil {
		return updateErr
	}
	return err
}

func (w *Worker) finish(ctx context.Context, job model.GameNotificationJob, notification model.GameNotification, successCount, failureCount int, lastError, jobStatus, notificationStatus string) error {
	now := time.Now()
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.GameNotificationJob{}).Where("id=?", job.ID).
			Updates(map[string]any{"status": jobStatus, "last_error": truncateText(lastError, 500)}).Error; err != nil {
			return err
		}
		return tx.Model(&model.GameNotification{}).Where("id=?", notification.ID).Updates(map[string]any{
			"status":            notificationStatus,
			"target_user_count": successCount + failureCount,
			"success_count":     successCount,
			"failure_count":     failureCount,
			"sent_at":           &now,
		}).Error
	})
}

func (w *Worker) batchSize() int {
	if w.config.BatchSize <= 0 || w.config.BatchSize > maxFCMBatchSize {
		return maxFCMBatchSize
	}
	return w.config.BatchSize
}

func plainText(value string) string {
	return strings.Join(strings.Fields(html.UnescapeString(htmlTagPattern.ReplaceAllString(value, " "))), " ")
}

func truncateText(value string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxRunes])
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func fcmErrorCode(err error) string {
	switch {
	case messaging.IsUnregistered(err):
		return "FCM_UNREGISTERED"
	case messaging.IsInvalidArgument(err):
		return "FCM_INVALID_ARGUMENT"
	case messaging.IsQuotaExceeded(err):
		return "FCM_QUOTA_EXCEEDED"
	case messaging.IsUnavailable(err):
		return "FCM_UNAVAILABLE"
	default:
		return "FCM_SEND_FAILED"
	}
}
