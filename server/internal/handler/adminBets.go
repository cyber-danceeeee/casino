package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/adminauth"
	"casino/internal/betadmin"
)

type AdminBetsHandler interface {
	ListBets(*gin.Context)
	GetBet(*gin.Context)
	MarkBetForReview(*gin.Context)
	ListRounds(*gin.Context)
	GetRound(*gin.Context)
	ListRollbacks(*gin.Context)
	GetRollback(*gin.Context)
	MarkRollbackForInvestigation(*gin.Context)
	ListAbnormal(*gin.Context)
	GetAbnormal(*gin.Context)
	AssignAbnormal(*gin.Context)
	ReviewAbnormal(*gin.Context)
	ClearAbnormal(*gin.Context)
	RestrictAbnormal(*gin.Context)
	AddAbnormalNote(*gin.Context)
	ListCallbacks(*gin.Context)
	GetCallback(*gin.Context)
	MarkCallbackForInvestigation(*gin.Context)
	ListReconciliationRuns(*gin.Context)
	GenerateReconciliationRun(*gin.Context)
	GetReconciliationRun(*gin.Context)
	ListReconciliationItems(*gin.Context)
	UpdateReconciliationItemStatus(*gin.Context)
}

type adminBetsHandler struct{ service *betadmin.Service }

func NewAdminBetsHandler() AdminBetsHandler { return &adminBetsHandler{service: betadmin.NewService()} }

func (h *adminBetsHandler) ListBets(c *gin.Context) {
	data, err := h.service.ListBets(betListQuery(c, betQueryKeys))
	h.write(c, data, err)
}
func (h *adminBetsHandler) ListRounds(c *gin.Context) {
	data, err := h.service.ListRounds(betListQuery(c, roundQueryKeys))
	h.write(c, data, err)
}
func (h *adminBetsHandler) ListRollbacks(c *gin.Context) {
	data, err := h.service.ListRollbacks(betListQuery(c, rollbackQueryKeys))
	h.write(c, data, err)
}
func (h *adminBetsHandler) ListAbnormal(c *gin.Context) {
	data, err := h.service.ListAbnormal(betListQuery(c, abnormalQueryKeys))
	h.write(c, data, err)
}
func (h *adminBetsHandler) ListCallbacks(c *gin.Context) {
	data, err := h.service.ListCallbacks(betListQuery(c, callbackQueryKeys))
	h.write(c, data, err)
}
func (h *adminBetsHandler) ListReconciliationRuns(c *gin.Context) {
	data, err := h.service.ListReconciliationRuns(betListQuery(c, reconciliationRunQueryKeys))
	h.write(c, data, err)
}
func (h *adminBetsHandler) ListReconciliationItems(c *gin.Context) {
	data, err := h.service.ListReconciliationItems(betListQuery(c, reconciliationItemQueryKeys))
	h.write(c, data, err)
}

func (h *adminBetsHandler) GetBet(c *gin.Context) {
	if id, ok := betPathID(c); ok {
		data, err := h.service.BetDetail(id, adminContentScope(c))
		h.write(c, data, err)
	}
}
func (h *adminBetsHandler) GetRound(c *gin.Context) {
	if id, ok := betPathID(c); ok {
		data, err := h.service.RoundDetail(id, adminContentScope(c))
		h.write(c, data, err)
	}
}
func (h *adminBetsHandler) GetRollback(c *gin.Context) {
	if id, ok := betPathID(c); ok {
		data, err := h.service.RollbackDetail(id, adminContentScope(c))
		h.write(c, data, err)
	}
}
func (h *adminBetsHandler) GetAbnormal(c *gin.Context) {
	if id, ok := betPathID(c); ok {
		data, err := h.service.AbnormalDetail(id, adminContentScope(c))
		h.write(c, data, err)
	}
}
func (h *adminBetsHandler) GetCallback(c *gin.Context) {
	if id, ok := betPathID(c); ok {
		data, err := h.service.CallbackDetail(id, adminContentScope(c))
		h.write(c, data, err)
	}
}
func (h *adminBetsHandler) GetReconciliationRun(c *gin.Context) {
	if id, ok := betPathID(c); ok {
		data, err := h.service.ReconciliationRunDetail(id, adminContentScope(c))
		h.write(c, data, err)
	}
}

func (h *adminBetsHandler) MarkBetForReview(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.ReasonCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.MarkBetForReview(id, body, betActor(c)))
}
func (h *adminBetsHandler) MarkRollbackForInvestigation(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.ReasonCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.MarkRollbackForInvestigation(id, body, betActor(c)))
}
func (h *adminBetsHandler) MarkCallbackForInvestigation(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.ReasonCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.MarkCallbackForInvestigation(id, body, betActor(c)))
}
func (h *adminBetsHandler) GenerateReconciliationRun(c *gin.Context) {
	var body betadmin.GenerateReconciliationCommand
	if !betBind(c, &body) {
		return
	}
	data, err := h.service.GenerateProviderReconciliation(body, betActor(c))
	h.write(c, data, err)
}
func (h *adminBetsHandler) UpdateReconciliationItemStatus(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.ReconciliationStatusCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.UpdateReconciliationItemStatus(id, body, betActor(c)))
}
func (h *adminBetsHandler) AssignAbnormal(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.AssignCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.AssignAbnormal(id, body, betActor(c)))
}
func (h *adminBetsHandler) ReviewAbnormal(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.ReviewCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.ReviewAbnormal(id, body, betActor(c)))
}
func (h *adminBetsHandler) ClearAbnormal(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.ReasonCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.ClearAbnormal(id, body, betActor(c)))
}
func (h *adminBetsHandler) RestrictAbnormal(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.ReasonCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.RestrictAbnormal(id, body, betActor(c)))
}
func (h *adminBetsHandler) AddAbnormalNote(c *gin.Context) {
	id, ok := betPathID(c)
	if !ok {
		return
	}
	var body betadmin.NoteCommand
	if !betBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.AddAbnormalNote(id, body, betActor(c)))
}

func (h *adminBetsHandler) write(c *gin.Context, data any, err error) {
	if err != nil {
		betError(c, err)
		return
	}
	response.Success(c, data)
}
func (h *adminBetsHandler) writeAction(c *gin.Context, err error) {
	if err != nil {
		betError(c, err)
		return
	}
	response.Success(c)
}
func betPathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		betError(c, &betadmin.Error{Code: "INVALID_REQUEST", Message: "invalid id"})
		return 0, false
	}
	return id, true
}
func betBind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		betError(c, &betadmin.Error{Code: "INVALID_REQUEST", Message: "invalid request body"})
		return false
	}
	return true
}
func betActor(c *gin.Context) betadmin.Actor {
	actor, _ := adminauth.FromContext(c)
	correlation := c.GetHeader("X-Request-ID")
	if correlation == "" {
		correlation = c.GetString("request_id")
	}
	return betadmin.Actor{ID: actor.ID, Username: actor.Username, Scope: adminContentScope(c), IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: correlation}
}
func betListQuery(c *gin.Context, keys []string) betadmin.ListQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	values := map[string]string{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values[key] = value
		}
	}
	return betadmin.ListQuery{Page: page, PageSize: pageSize, Values: values, Scope: adminContentScope(c)}
}
func betError(c *gin.Context, err error) {
	var domain *betadmin.Error
	if errors.As(err, &domain) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": domain.Message, "data": gin.H{"errorCode": domain.Code}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

var betQueryKeys = []string{"platformBetID", "providerTransactionID", "providerRoundID", "userNo", "username", "providerCode", "gameID", "providerGameID", "gameType", "currency", "status", "amountMin", "amountMax", "winAmountMin", "winAmountMax", "createdFrom", "createdTo", "settledFrom", "settledTo", "rollbackStatus"}
var roundQueryKeys = []string{"providerRoundID", "userNo", "username", "providerCode", "game", "currency", "status", "betAmountMin", "betAmountMax", "winAmountMin", "winAmountMax", "startedFrom", "startedTo", "completedFrom", "completedTo"}
var rollbackQueryKeys = []string{"rollbackID", "originalTransactionID", "providerTransactionID", "providerRoundID", "userNo", "username", "providerCode", "operation", "currency", "status", "createdFrom", "createdTo", "completedFrom", "completedTo"}
var abnormalQueryKeys = []string{"userNo", "username", "providerCode", "game", "riskLevel", "riskType", "reviewStatus", "assignedReviewer", "createdFrom", "createdTo", "updatedFrom", "updatedTo"}
var callbackQueryKeys = []string{"providerCode", "operation", "userNo", "userPhone", "providerPlayerID", "providerTransactionID", "providerRoundID", "gameID", "httpMethod", "signatureResult", "processingResult", "sourceIP", "createdFrom", "createdTo"}
var reconciliationRunQueryKeys = []string{"providerCode", "currency", "status", "businessDateFrom", "businessDateTo", "createdFrom", "createdTo"}
var reconciliationItemQueryKeys = []string{"runID", "providerCode", "currency", "discrepancyType", "status", "providerTransactionID", "providerRoundID", "userNo", "businessDateFrom", "businessDateTo", "createdFrom", "createdTo"}
