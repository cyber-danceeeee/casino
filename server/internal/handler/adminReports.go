package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/adminauth"
	"casino/internal/database"
	"casino/internal/reportadmin"
)

type AdminReportsHandler struct{ services *reportadmin.Services }

func NewAdminReportsHandler() *AdminReportsHandler {
	return &AdminReportsHandler{services: reportadmin.New(database.GetDB())}
}

func (h *AdminReportsHandler) Financial(c *gin.Context) {
	d, e := h.services.Financial.Report(reportQuery(c, financialReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) FinancialDetail(c *gin.Context) {
	d, e := h.services.Financial.Detail(reportQuery(c, financialReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) Games(c *gin.Context) {
	d, e := h.services.Game.Report(reportQuery(c, gameReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) GameDetail(c *gin.Context) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	d, e := h.services.Game.Detail(id, reportQuery(c, gameReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) Providers(c *gin.Context) {
	d, e := h.services.Provider.Report(reportQuery(c, providerReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) ProviderDetail(c *gin.Context) {
	code := strings.TrimSpace(c.Param("providerCode"))
	if code == "" {
		reportError(c, reportadmin.NewError("INVALID_REQUEST", "invalid provider code"))
		return
	}
	d, e := h.services.Provider.Detail(code, reportQuery(c, providerReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) Users(c *gin.Context) {
	d, e := h.services.User.Report(reportQuery(c, userReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) UserDetail(c *gin.Context) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	d, e := h.services.User.Detail(id, reportQuery(c, userReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) Channels(c *gin.Context) {
	d, e := h.services.Channel.Report(reportQuery(c, channelReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) ChannelDetail(c *gin.Context) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	d, e := h.services.Channel.Detail(id, reportQuery(c, channelReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) Agents(c *gin.Context) {
	d, e := h.services.Agent.Report(reportQuery(c, agentReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) AgentDetail(c *gin.Context) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	d, e := h.services.Agent.Detail(id, reportQuery(c, agentReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) Promotions(c *gin.Context) {
	d, e := h.services.Promotion.Report(reportQuery(c, promotionReportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) PromotionDetail(c *gin.Context) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	d, e := h.services.Promotion.Detail(id, reportQuery(c, promotionReportKeys))
	reportWrite(c, d, e)
}

func (h *AdminReportsHandler) CreateExport(c *gin.Context) {
	var in reportadmin.ExportCommand
	if !reportBind(c, &in) {
		return
	}
	adminActor, _ := adminauth.FromContext(c)
	if !adminActor.Can(exportReportPermission(in.ReportType)) {
		reportError(c, reportadmin.NewError("PERMISSION_DENIED", "permission denied"))
		return
	}
	id, e := h.services.Export.Create(in, reportActor(c))
	reportWrite(c, map[string]any{"id": id}, e)
}
func (h *AdminReportsHandler) ListExports(c *gin.Context) {
	d, e := h.services.Export.List(reportQuery(c, exportKeys))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) ExportDetail(c *gin.Context) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	actor, _ := adminauth.FromContext(c)
	d, e := h.services.Export.Detail(id, reportActor(c), actor.Can(adminauth.PermissionReportSensitiveView))
	reportWrite(c, d, e)
}
func (h *AdminReportsHandler) DownloadExport(c *gin.Context) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	actor, _ := adminauth.FromContext(c)
	path, format, e := h.services.Export.Download(id, reportActor(c), actor.Can(adminauth.PermissionReportSensitiveView))
	if e != nil {
		reportError(c, e)
		return
	}
	_ = format
	c.Redirect(http.StatusFound, path)
}
func (h *AdminReportsHandler) RetryExport(c *gin.Context)  { h.exportAction(c, true) }
func (h *AdminReportsHandler) CancelExport(c *gin.Context) { h.exportAction(c, false) }
func (h *AdminReportsHandler) exportAction(c *gin.Context, retry bool) {
	id, ok := reportID(c)
	if !ok {
		return
	}
	var in reportadmin.ExportAction
	if !reportBind(c, &in) {
		return
	}
	var e error
	if retry {
		e = h.services.Export.Retry(id, in.Reason, reportActor(c))
	} else {
		e = h.services.Export.Cancel(id, in.Reason, reportActor(c))
	}
	reportWrite(c, nil, e)
}

func reportActor(c *gin.Context) reportadmin.Actor {
	a, _ := adminauth.FromContext(c)
	return reportadmin.Actor{ID: a.ID, Username: a.Username, IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: c.GetHeader("X-Request-ID"), Scope: adminTenancyScope(c)}
}
func reportID(c *gin.Context) (uint64, bool) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil || id == 0 {
		reportError(c, reportadmin.NewError("INVALID_REQUEST", "invalid id"))
		return 0, false
	}
	return id, true
}
func reportBind(c *gin.Context, out any) bool {
	if e := c.ShouldBindJSON(out); e != nil {
		reportError(c, reportadmin.NewError("INVALID_REQUEST", "invalid request body"))
		return false
	}
	return true
}
func reportQuery(c *gin.Context, keys []string) reportadmin.Query {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	values := map[string]string{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values[key] = value
		}
	}
	return reportadmin.Query{Page: p, PageSize: ps, Values: values, Scope: adminTenancyScope(c)}
}
func reportWrite(c *gin.Context, data any, err error) {
	if err != nil {
		reportError(c, err)
		return
	}
	response.Success(c, data)
}
func reportError(c *gin.Context, err error) {
	var serviceErr *reportadmin.ServiceError
	if errors.As(err, &serviceErr) {
		status := http.StatusBadRequest
		if serviceErr.Code == "PERMISSION_DENIED" {
			status = http.StatusForbidden
		}
		c.JSON(http.StatusOK, gin.H{"code": status, "msg": serviceErr.Message, "data": gin.H{"errorCode": serviceErr.Code}})
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusNotFound, "msg": "record not found", "data": gin.H{"errorCode": "NOT_FOUND"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}
func exportReportPermission(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "FINANCIAL":
		return adminauth.PermissionReportFinancialView
	case "GAME":
		return adminauth.PermissionReportGameView
	case "PROVIDER":
		return adminauth.PermissionReportProviderView
	case "USER":
		return adminauth.PermissionReportUserView
	case "CHANNEL":
		return adminauth.PermissionReportChannelView
	case "AGENT":
		return adminauth.PermissionReportAgentView
	default:
		return adminauth.PermissionReportPromotionView
	}
}

var commonReportKeys = []string{"dateFrom", "dateTo", "granularity", "timezone", "country", "currency"}
var financialReportKeys = append(append([]string{}, commonReportKeys...), "paymentChannel", "userChannel")
var gameReportKeys = append(append([]string{}, commonReportKeys...), "provider", "gameID", "gameCode", "gameName", "providerGameID", "gameType", "platformCategory", "device")
var providerReportKeys = append(append([]string{}, commonReportKeys...), "provider", "gameType")
var userReportKeys = append(append([]string{}, commonReportKeys...), "userID", "playerNo", "username", "vipLevel", "registrationChannel", "agent", "riskStatus")
var channelReportKeys = append(append([]string{}, commonReportKeys...), "channel", "agent", "platform")
var agentReportKeys = append(append([]string{}, commonReportKeys...), "agent", "agentLevel", "parentAgent", "commissionPlan")
var promotionReportKeys = append(append([]string{}, commonReportKeys...), "promotion", "promotionType", "status")
var exportKeys = []string{"taskID", "reportType", "operator", "status", "createdFrom", "createdTo", "completedFrom", "completedTo"}
