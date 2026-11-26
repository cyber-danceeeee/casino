package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/adminauth"
	"casino/internal/agentadmin"
	"casino/internal/database"
)

type AdminAgentsHandler struct{ service *agentadmin.Service }

func NewAdminAgentsHandler() *AdminAgentsHandler {
	return &AdminAgentsHandler{service: agentadmin.New(database.GetDB())}
}

func (h *AdminAgentsHandler) ListAgents(c *gin.Context) {
	d, e := h.service.ListAgents(agentQuery(c, agentKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GetAgent(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.AgentDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) CreateAgent(c *gin.Context) {
	var in agentadmin.AgentInput
	if !agentBind(c, &in) {
		return
	}
	id, e := h.service.SaveAgent(0, in, agentActor(c))
	agentWrite(c, map[string]any{"id": id}, e)
}
func (h *AdminAgentsHandler) UpdateAgent(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.AgentInput
	if !agentBind(c, &in) {
		return
	}
	_, e := h.service.SaveAgent(id, in, agentActor(c))
	agentWrite(c, nil, e)
}
func (h *AdminAgentsHandler) AgentStatus(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.StatusCommand
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.SetAgentStatus(id, in, agentActor(c)))
}
func (h *AdminAgentsHandler) AgentUsers(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.AgentUsers(id, agentQuery(c, []string{"username", "channel"}))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) AgentChildren(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	q := agentQuery(c, agentKeys)
	q.Values["parentAgentID"] = strconv.FormatUint(id, 10)
	d, e := h.service.ListAgents(q)
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) Hierarchy(c *gin.Context) {
	d, e := h.service.Hierarchy(agentQuery(c, agentKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) ChangeParent(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.ChangeParentInput
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.ChangeParent(id, in, agentActor(c)))
}
func (h *AdminAgentsHandler) AssignPlan(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.AssignPlanInput
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.AssignPlan(id, in, agentActor(c)))
}

func (h *AdminAgentsHandler) ListChannels(c *gin.Context) {
	d, e := h.service.ListChannels(agentQuery(c, channelKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GetChannel(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.ChannelDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) CreateChannel(c *gin.Context) {
	var in agentadmin.ChannelInput
	if !agentBind(c, &in) {
		return
	}
	id, e := h.service.SaveChannel(0, in, agentActor(c))
	agentWrite(c, map[string]any{"id": id}, e)
}
func (h *AdminAgentsHandler) UpdateChannel(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.ChannelInput
	if !agentBind(c, &in) {
		return
	}
	_, e := h.service.SaveChannel(id, in, agentActor(c))
	agentWrite(c, nil, e)
}
func (h *AdminAgentsHandler) ChannelStatus(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.StatusCommand
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.SetChannelStatus(id, in, agentActor(c)))
}

func (h *AdminAgentsHandler) ListPlans(c *gin.Context) {
	d, e := h.service.ListPlans(agentQuery(c, planKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GetPlan(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.PlanDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) CreatePlan(c *gin.Context) {
	var in agentadmin.CommissionPlanInput
	if !agentBind(c, &in) {
		return
	}
	id, e := h.service.SavePlan(0, in, agentActor(c))
	agentWrite(c, map[string]any{"id": id}, e)
}
func (h *AdminAgentsHandler) UpdatePlan(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.CommissionPlanInput
	if !agentBind(c, &in) {
		return
	}
	_, e := h.service.SavePlan(id, in, agentActor(c))
	agentWrite(c, nil, e)
}
func (h *AdminAgentsHandler) PlanStatus(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.StatusCommand
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.SetPlanStatus(id, in, agentActor(c)))
}
func (h *AdminAgentsHandler) DeletePlan(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.ReasonCommand
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.DeletePlan(id, in.Reason, agentActor(c)))
}

func (h *AdminAgentsHandler) ListCommissions(c *gin.Context) {
	d, e := h.service.ListCommissionRecords(agentQuery(c, commissionKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GetCommission(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.CommissionDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) ListSettlements(c *gin.Context) {
	d, e := h.service.ListSettlements(agentQuery(c, settlementKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GetSettlement(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.SettlementDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) ExportSettlement(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.SettlementDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GenerateSettlement(c *gin.Context) {
	var in agentadmin.GenerateSettlementInput
	if !agentBind(c, &in) {
		return
	}
	id, e := h.service.GenerateSettlement(in, agentActor(c))
	agentWrite(c, map[string]any{"id": id}, e)
}
func (h *AdminAgentsHandler) SettlementAction(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := agentID(c)
		if !ok {
			return
		}
		var in agentadmin.ReasonCommand
		if !agentBind(c, &in) {
			return
		}
		agentWrite(c, nil, h.service.SettlementAction(id, action, in.Reason, agentActor(c)))
	}
}

func (h *AdminAgentsHandler) ListWallets(c *gin.Context) {
	d, e := h.service.ListWallets(agentQuery(c, walletKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GetWallet(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.WalletDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) ListWalletTransactions(c *gin.Context) {
	d, e := h.service.ListWalletTransactions(agentQuery(c, []string{"walletID", "agentID", "transactionType", "createdFrom", "createdTo"}))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) WalletFrozen(frozen bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := agentID(c)
		if !ok {
			return
		}
		var in agentadmin.ReasonCommand
		if !agentBind(c, &in) {
			return
		}
		agentWrite(c, nil, h.service.SetWalletFrozen(id, frozen, in.Reason, agentActor(c)))
	}
}
func (h *AdminAgentsHandler) WalletAdjustment(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	var in agentadmin.WalletAdjustmentInput
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.AdjustWallet(id, in, agentActor(c)))
}
func (h *AdminAgentsHandler) ChangeAttribution(c *gin.Context) {
	var in agentadmin.AttributionInput
	if !agentBind(c, &in) {
		return
	}
	agentWrite(c, nil, h.service.ChangeAttribution(in, agentActor(c)))
}
func (h *AdminAgentsHandler) ListLogs(c *gin.Context) {
	d, e := h.service.ListLogs(agentQuery(c, logKeys))
	agentWrite(c, d, e)
}
func (h *AdminAgentsHandler) GetLog(c *gin.Context) {
	id, ok := agentID(c)
	if !ok {
		return
	}
	d, e := h.service.LogDetail(id, adminTenancyScope(c))
	agentWrite(c, d, e)
}

func agentActor(c *gin.Context) agentadmin.Actor {
	a, _ := adminauth.FromContext(c)
	return agentadmin.Actor{ID: a.ID, Username: a.Username, Scope: adminTenancyScope(c), IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: c.GetHeader("X-Request-ID")}
}
func agentID(c *gin.Context) (uint64, bool) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil || id == 0 {
		agentError(c, agentadmin.NewError("INVALID_REQUEST", "invalid id"))
		return 0, false
	}
	return id, true
}
func agentBind(c *gin.Context, out any) bool {
	if e := c.ShouldBindJSON(out); e != nil {
		agentError(c, agentadmin.NewError("INVALID_REQUEST", "invalid request body"))
		return false
	}
	return true
}
func agentQuery(c *gin.Context, keys []string) agentadmin.ListQuery {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	v := map[string]string{}
	for _, k := range keys {
		if x := strings.TrimSpace(c.Query(k)); x != "" {
			v[k] = x
		}
	}
	return agentadmin.ListQuery{Page: p, PageSize: ps, Values: v, Scope: adminTenancyScope(c)}
}
func agentWrite(c *gin.Context, data any, e error) {
	if e != nil {
		agentError(c, e)
		return
	}
	response.Success(c, data)
}
func agentError(c *gin.Context, e error) {
	var de *agentadmin.Error
	if errors.As(e, &de) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": de.Message, "data": gin.H{"errorCode": de.Code}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

var agentKeys = []string{"agentID", "agentCode", "agentName", "agentLevel", "parentAgent", "parentAgentID", "status", "country", "currency", "commissionPlan", "registrationChannel", "createdFrom", "createdTo"}
var channelKeys = []string{"channelCode", "channelName", "agent", "country", "platform", "status", "createdFrom", "createdTo"}
var planKeys = []string{"planCode", "planName", "commissionModel", "currency", "status", "createdFrom", "createdTo"}
var commissionKeys = []string{"recordID", "recordNo", "agentID", "agentCode", "agentName", "userID", "playerNo", "commissionType", "businessDate", "currency", "status", "createdFrom", "createdTo"}
var settlementKeys = []string{"settlementID", "settlementNo", "agentID", "agentCode", "agent", "settlementCycle", "currency", "status", "amountMin", "amountMax", "createdFrom", "createdTo", "completedFrom", "completedTo"}
var walletKeys = []string{"agentID", "agentCode", "agentName", "walletID", "walletNo", "currency", "walletStatus", "balanceMin", "balanceMax", "updatedFrom", "updatedTo"}
var logKeys = []string{"operator", "action", "agentID", "targetType", "targetID", "createdFrom", "createdTo"}
