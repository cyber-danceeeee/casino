package handler

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"casino/internal/database"
	"casino/internal/systemadmin"
)

var maintenanceQueryKeys = []string{"tenantID", "siteID", "scopeType", "providerCode", "enabled", "page", "pageSize"}
var supportQueryKeys = []string{"tenantID", "siteID", "query", "searchType", "playerID", "playerNo", "status", "category", "priority", "from", "to", "eventTypes", "providerCode", "minAmount", "maxAmount", "page", "pageSize"}
var supportExportQueryKeys = []string{"tenantID", "siteID", "status", "exportType", "page", "pageSize"}
var monitorQueryKeys = []string{"tenantID", "siteID", "country", "currency", "providerCode", "channelCode", "methodCode", "status", "from", "to", "page", "pageSize"}
var alertQueryKeys = []string{"severity", "status", "alertType", "source", "page", "pageSize"}
var releaseQueryKeys = []string{"tenantID", "siteID", "environment", "status", "page", "pageSize"}
var backupQueryKeys = []string{"tenantID", "siteID", "scopeType", "status", "page", "pageSize"}

type adminRateBucket struct {
	count   int
	expires time.Time
}

var expensiveAdminRate = struct {
	sync.Mutex
	buckets map[string]adminRateBucket
}{buckets: make(map[string]adminRateBucket)}

func allowExpensiveAdminRequest(c *gin.Context, operation string, limit int, window time.Duration) bool {
	actor := systemActor(c)
	identity := strconv.FormatUint(actor.ID, 10)
	if actor.ID == 0 {
		identity = c.ClientIP()
	}
	key := operation + ":" + identity
	now := time.Now()
	expensiveAdminRate.Lock()
	bucket := expensiveAdminRate.buckets[key]
	if bucket.expires.Before(now) {
		bucket = adminRateBucket{expires: now.Add(window)}
	}
	bucket.count++
	expensiveAdminRate.buckets[key] = bucket
	if len(expensiveAdminRate.buckets) > 4096 {
		for candidate, value := range expensiveAdminRate.buckets {
			if value.expires.Before(now) {
				delete(expensiveAdminRate.buckets, candidate)
			}
		}
	}
	expensiveAdminRate.Unlock()
	if bucket.count > limit {
		systemError(c, systemadmin.NewError("RATE_LIMITED", "too many expensive administrative requests; retry later"))
		return false
	}
	return true
}

func OperationsLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP", "startedAt": systemadmin.ApplicationStartedAtForHealth(), "checkedAt": time.Now()})
}

func OperationsReadiness(c *gin.Context) {
	ready, checks := systemadmin.New(database.GetDB()).Readiness(c.Request.Context())
	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{"status": map[bool]string{true: "READY", false: "NOT_READY"}[ready], "checks": checks})
}

func (h *AdminSystemHandler) ListTenantMaintenance(c *gin.Context) {
	data, err := h.service.ListTenantMaintenance(systemQuery(c, maintenanceQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SaveTenantMaintenance(c *gin.Context) {
	var input systemadmin.TenantMaintenanceInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SaveTenantMaintenance(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}

func (h *AdminSystemHandler) ListProviderMaintenance(c *gin.Context) {
	data, err := h.service.ListProviderMaintenance(systemQuery(c, maintenanceQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SaveProviderMaintenance(c *gin.Context) {
	var input systemadmin.ProviderMaintenanceInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SaveProviderMaintenance(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}

func (h *AdminSystemHandler) ListGameMaintenance(c *gin.Context) {
	data, err := h.service.ListGameMaintenance(systemQuery(c, maintenanceQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SaveGameMaintenance(c *gin.Context) {
	var input systemadmin.GameMaintenanceInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SaveGameMaintenance(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}

func (h *AdminSystemHandler) SearchSupportPlayers(c *gin.Context) {
	if !allowExpensiveAdminRequest(c, "support-player-search", 60, time.Minute) {
		return
	}
	data, err := h.service.SearchPlayers(systemQuery(c, supportQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SupportPlayerDetail(c *gin.Context) {
	if !allowExpensiveAdminRequest(c, "support-player-detail", 120, time.Minute) {
		return
	}
	id, ok := systemNamedID(c, "playerID")
	if !ok {
		return
	}
	data, err := h.service.Player360(id, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SupportPlayerDetailByNumber(c *gin.Context) {
	if !allowExpensiveAdminRequest(c, "support-player-detail", 120, time.Minute) {
		return
	}
	data, err := h.service.Player360ByNumber(c.Param("playerNo"), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SupportPlayerTimeline(c *gin.Context) {
	if !allowExpensiveAdminRequest(c, "support-player-timeline", 60, time.Minute) {
		return
	}
	id, ok := systemNamedID(c, "playerID")
	if !ok {
		return
	}
	data, err := h.service.PlayerTimeline(id, systemQuery(c, supportQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SupportPlayerTimelineByNumber(c *gin.Context) {
	if !allowExpensiveAdminRequest(c, "support-player-timeline", 60, time.Minute) {
		return
	}
	data, err := h.service.PlayerTimelineByNumber(c.Param("playerNo"), systemQuery(c, supportQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SearchSupportOrders(c *gin.Context) {
	if !allowExpensiveAdminRequest(c, "support-order-search", 60, time.Minute) {
		return
	}
	data, err := h.service.SearchOrders(systemQuery(c, supportQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SupportOrderDetail(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.SupportOrderDetail(c.Param("kind"), id, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ListSupportCases(c *gin.Context) {
	data, err := h.service.ListSupportCases(systemQuery(c, supportQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SupportCaseDetail(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.SupportCaseDetail(id, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) CreateSupportCase(c *gin.Context) {
	var input systemadmin.SupportCaseInput
	if !systemBind(c, &input) {
		return
	}
	input.IdempotencyKey = systemIdempotencyKey(c, input.IdempotencyKey)
	id, err := h.service.CreateSupportCase(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}

func (h *AdminSystemHandler) UpdateSupportCase(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.SupportCaseInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateSupportCase(id, input, systemActor(c)))
}

func (h *AdminSystemHandler) AddSupportCaseNote(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.SupportCaseNoteInput
	if !systemBind(c, &input) {
		return
	}
	noteID, err := h.service.AddSupportCaseNote(id, input, systemActor(c))
	systemWrite(c, map[string]any{"id": noteID}, err)
}

func (h *AdminSystemHandler) ListSupportExports(c *gin.Context) {
	data, err := h.service.ListSupportExports(systemQuery(c, supportExportQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) CreateSupportExport(c *gin.Context) {
	if !allowExpensiveAdminRequest(c, "support-export-create", 10, time.Hour) {
		return
	}
	var input systemadmin.SupportExportInput
	if !systemBind(c, &input) {
		return
	}
	input.IdempotencyKey = systemIdempotencyKey(c, input.IdempotencyKey)
	data, err := h.service.CreateSupportExport(input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) DownloadSupportExport(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, filename, err := h.service.ReadSupportExport(id, systemActor(c))
	if err != nil {
		systemError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func (h *AdminSystemHandler) PaymentMonitor(c *gin.Context) {
	data, err := h.service.PaymentMonitor(systemQuery(c, monitorQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) PaymentDiagnostics(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.PaymentDiagnostics(c.Param("kind"), id, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) PaymentDiagnosticsByNumber(c *gin.Context) {
	data, err := h.service.PaymentDiagnosticsByNumber(c.Param("kind"), c.Param("orderNo"), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) MarkPaymentReview(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.MarkPaymentReview(c.Param("kind"), id, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ProviderMonitor(c *gin.Context) {
	data, err := h.service.ProviderMonitor(systemQuery(c, monitorQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ProviderDiagnostics(c *gin.Context) {
	data, err := h.service.ProviderDiagnostics(c.Param("code"), systemQuery(c, monitorQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) AddOperationalNote(c *gin.Context) {
	var input systemadmin.OperationalNoteInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.AddOperationalNote(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}

func (h *AdminSystemHandler) DetailedHealth(c *gin.Context) {
	data, err := h.service.DetailedHealth(c.Request.Context())
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) JobsOverview(c *gin.Context) {
	data, err := h.service.JobsOverview(systemQuery(c, taskSystemKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) RetryTaskExecution(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.RetryTaskExecution(id, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) RetryDeadLetter(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.RetryDeadLetter(id, input.Reason, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) RetryRabbitDeadLetter(c *gin.Context) {
	var input struct {
		Queue  string `json:"queue" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.RetryRabbitDeadLetter(input.Queue, c.Param("eventID"), input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) PurgeRabbitRetry(c *gin.Context) {
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	count, err := h.service.PurgeRabbitRetry(c.Param("queue"), input.Reason, systemActor(c))
	systemWrite(c, map[string]any{"purged": count}, err)
}

func (h *AdminSystemHandler) ListAlerts(c *gin.Context) {
	data, err := h.service.ListAlerts(systemQuery(c, alertQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) AcknowledgeAlert(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.AlertActionInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.AcknowledgeAlert(id, input, systemActor(c)))
}

func (h *AdminSystemHandler) ResolveAlert(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.AlertActionInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.ResolveAlert(id, input, systemActor(c)))
}

func (h *AdminSystemHandler) ListReleases(c *gin.Context) {
	data, err := h.service.ListReleases(systemQuery(c, releaseQueryKeys), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ReleaseDetail(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.ReleaseDetail(id, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) CompareReleases(c *gin.Context) {
	fromID, err := strconv.ParseUint(c.Query("fromID"), 10, 64)
	if err != nil || fromID == 0 {
		systemError(c, systemadmin.NewError("INVALID_REQUEST", "from release is required"))
		return
	}
	toID, err := strconv.ParseUint(c.Query("toID"), 10, 64)
	if err != nil || toID == 0 {
		systemError(c, systemadmin.NewError("INVALID_REQUEST", "to release is required"))
		return
	}
	data, serviceErr := h.service.CompareReleases(fromID, toID, systemActor(c))
	systemWrite(c, data, serviceErr)
}

func (h *AdminSystemHandler) ValidateRelease(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteID")
	if !ok {
		return
	}
	data, err := h.service.ValidateRelease(siteID, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) PublishRelease(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteID")
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.PublishSiteConfigWithKey(siteID, input.Reason, systemIdempotencyKey(c, ""), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) RollbackRelease(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReleaseRollbackInput
	if !systemBind(c, &input) {
		return
	}
	input.IdempotencyKey = systemIdempotencyKey(c, input.IdempotencyKey)
	data, err := h.service.RollbackRelease(id, input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ListBackups(c *gin.Context) {
	data, err := h.service.ListBackups(systemQuery(c, backupQueryKeys))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) BackupDetail(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.BackupDetail(id)
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) CreateBackup(c *gin.Context) {
	var input systemadmin.BackupCreateInput
	if !systemBind(c, &input) {
		return
	}
	input.IdempotencyKey = systemIdempotencyKey(c, input.IdempotencyKey)
	data, err := h.service.CreateBackup(input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ValidateBackupRestore(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.ValidateBackupRestore(id)
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) RestoreBackup(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.BackupRestoreInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.RestoreBackup(id, input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) DeleteBackup(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeleteBackup(id, input.Reason, systemActor(c)))
}
