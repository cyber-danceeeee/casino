package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/adminauth"
	"casino/internal/database"
	"casino/internal/riskadmin"
)

type AdminRiskHandler struct{ service *riskadmin.Service }

func NewAdminRiskHandler() *AdminRiskHandler {
	return &AdminRiskHandler{service: riskadmin.New(database.GetDB())}
}
func (h *AdminRiskHandler) Dashboard(c *gin.Context) {
	data, err := h.service.Dashboard(c.Query("from"), c.Query("to"))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) ListRules(c *gin.Context) {
	data, err := h.service.Rules.List(riskList(c, ruleRiskKeys))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) GetRule(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		data, err := h.service.Rules.Detail(id)
		h.write(c, data, err)
	}
}
func (h *AdminRiskHandler) CreateRule(c *gin.Context) {
	var in riskadmin.RuleInput
	if riskBind(c, &in) {
		id, err := h.service.Rules.Save(0, in, riskActor(c))
		h.write(c, map[string]any{"id": id}, err)
	}
}
func (h *AdminRiskHandler) UpdateRule(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.RuleInput
	if riskBind(c, &in) {
		id, err := h.service.Rules.Save(id, in, riskActor(c))
		h.write(c, map[string]any{"id": id}, err)
	}
}
func (h *AdminRiskHandler) CopyRule(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		newID, err := h.service.Rules.Copy(id, riskActor(c))
		h.write(c, map[string]any{"id": newID}, err)
	}
}
func (h *AdminRiskHandler) RuleStatus(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.RuleStatusCommand
	if riskBind(c, &in) {
		h.action(c, h.service.Rules.Status(id, in, riskActor(c)))
	}
}
func (h *AdminRiskHandler) TestRule(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.RuleTestCommand
	if riskBind(c, &in) {
		data, err := h.service.Rules.Test(id, in, riskActor(c))
		h.write(c, data, err)
	}
}
func (h *AdminRiskHandler) DeleteRule(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in struct {
		Reason string `json:"reason" binding:"required,max=500"`
	}
	if riskBind(c, &in) {
		h.action(c, h.service.Rules.Delete(id, in.Reason, riskActor(c)))
	}
}

func (h *AdminRiskHandler) ListEvents(c *gin.Context) {
	data, err := h.service.Events.List(riskList(c, eventRiskKeys))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) GetEvent(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		data, err := h.service.Events.Detail(id)
		h.write(c, data, err)
	}
}
func (h *AdminRiskHandler) CreateEvent(c *gin.Context) {
	var in struct {
		UserID    uint64 `json:"userID" binding:"required"`
		RiskType  string `json:"riskType" binding:"required"`
		RiskLevel string `json:"riskLevel" binding:"required"`
		Reason    string `json:"reason" binding:"required,max=500"`
	}
	if riskBind(c, &in) {
		id, err := h.service.Events.Create(in.UserID, in.RiskType, in.RiskLevel, in.Reason, riskActor(c))
		h.write(c, map[string]any{"id": id}, err)
	}
}
func (h *AdminRiskHandler) AssignEvent(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.AssignmentCommand
	if riskBind(c, &in) {
		h.action(c, h.service.Events.Assign(id, in, riskActor(c)))
	}
}
func (h *AdminRiskHandler) EventAction(target string, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := riskID(c)
		if !ok {
			return
		}
		var in riskadmin.EventActionCommand
		if riskBind(c, &in) {
			h.action(c, h.service.Events.Transition(id, target, in, permission, riskActor(c)))
		}
	}
}
func (h *AdminRiskHandler) EventNote(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.NoteCommand
	if riskBind(c, &in) {
		h.action(c, h.service.Events.AddNote(id, in, riskActor(c)))
	}
}

func (h *AdminRiskHandler) ListReviews(c *gin.Context) {
	data, err := h.service.Reviews.List(riskList(c, reviewRiskKeys))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) GetReview(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		data, err := h.service.Reviews.Detail(id)
		h.write(c, data, err)
	}
}
func (h *AdminRiskHandler) ClaimReview(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		h.action(c, h.service.Reviews.Claim(id, riskActor(c)))
	}
}
func (h *AdminRiskHandler) AssignReview(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.AssignmentCommand
	if riskBind(c, &in) {
		h.action(c, h.service.Reviews.Assign(id, in, riskActor(c)))
	}
}
func (h *AdminRiskHandler) DecideReview(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.ReviewDecisionCommand
	if riskBind(c, &in) {
		actor, _ := adminauth.FromContext(c)
		permission := adminauth.PermissionRiskReviewDecision
		if strings.EqualFold(in.Decision, "ESCALATE") {
			permission = adminauth.PermissionRiskReviewEscalate
		}
		if !actor.Can(permission) {
			riskError(c, &riskadmin.Error{Code: "PERMISSION_DENIED", Message: "permission denied"})
			return
		}
		h.action(c, h.service.Reviews.Decision(id, in, riskActor(c)))
	}
}
func (h *AdminRiskHandler) ReviewNote(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.NoteCommand
	if riskBind(c, &in) {
		h.action(c, h.service.Reviews.AddNote(id, in, riskActor(c)))
	}
}

func (h *AdminRiskHandler) ListAssociations(c *gin.Context) {
	data, err := h.service.Associations.List(riskList(c, associationRiskKeys))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) GetAssociations(c *gin.Context) {
	id, ok := riskParamID(c, "userId")
	if ok {
		data, err := h.service.Associations.Detail(id, riskActor(c).Scope)
		h.write(c, data, err)
	}
}
func (h *AdminRiskHandler) ListBlacklists(c *gin.Context) {
	data, err := h.service.Blacklists.List(riskList(c, blacklistRiskKeys))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) GetBlacklist(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		data, err := h.service.Blacklists.Detail(id)
		h.write(c, data, err)
	}
}
func (h *AdminRiskHandler) CreateBlacklist(c *gin.Context) {
	var in riskadmin.BlacklistInput
	if riskBind(c, &in) {
		id, err := h.service.Blacklists.Create(in, riskActor(c))
		h.write(c, map[string]any{"id": id}, err)
	}
}
func (h *AdminRiskHandler) UpdateBlacklist(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.BlacklistUpdateCommand
	if riskBind(c, &in) {
		h.action(c, h.service.Blacklists.Update(id, in, riskActor(c)))
	}
}
func (h *AdminRiskHandler) RemoveBlacklist(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in struct {
		Reason string `json:"reason" binding:"required,max=500"`
	}
	if riskBind(c, &in) {
		h.action(c, h.service.Blacklists.Remove(id, in.Reason, riskActor(c)))
	}
}

func (h *AdminRiskHandler) ListWithdrawals(c *gin.Context) {
	data, err := h.service.Withdrawals.List(riskList(c, withdrawalRiskKeys))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) GetWithdrawal(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		data, err := h.service.Withdrawals.Detail(id, adminContentScope(c))
		h.write(c, data, err)
	}
}
func (h *AdminRiskHandler) DecideWithdrawal(c *gin.Context) {
	id, ok := riskID(c)
	if !ok {
		return
	}
	var in riskadmin.WithdrawalDecisionCommand
	if riskBind(c, &in) {
		actor, _ := adminauth.FromContext(c)
		permission := adminauth.PermissionRiskWithdrawalRestrict
		switch strings.ToUpper(in.Decision) {
		case "APPROVE":
			permission = adminauth.PermissionRiskWithdrawalApprove
		case "REJECT":
			permission = adminauth.PermissionRiskWithdrawalReject
		}
		if !actor.Can(permission) {
			riskError(c, &riskadmin.Error{Code: "PERMISSION_DENIED", Message: "permission denied"})
			return
		}
		h.action(c, h.service.Withdrawals.Decision(id, in, riskActor(c)))
	}
}
func (h *AdminRiskHandler) ListOperationLogs(c *gin.Context) {
	data, err := h.service.OperationLogs(riskList(c, operationLogRiskKeys))
	h.write(c, data, err)
}
func (h *AdminRiskHandler) GetOperationLog(c *gin.Context) {
	id, ok := riskID(c)
	if ok {
		data, err := h.service.OperationLogDetail(id)
		h.write(c, data, err)
	}
}

func (h *AdminRiskHandler) write(c *gin.Context, data any, err error) {
	if err != nil {
		riskError(c, err)
		return
	}
	response.Success(c, data)
}
func (h *AdminRiskHandler) action(c *gin.Context, err error) {
	if err != nil {
		riskError(c, err)
		return
	}
	response.Success(c)
}
func riskID(c *gin.Context) (uint64, bool) { return riskParamID(c, "id") }
func riskParamID(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		riskError(c, &riskadmin.Error{Code: "INVALID_REQUEST", Message: "invalid id"})
		return 0, false
	}
	return id, true
}
func riskBind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		riskError(c, &riskadmin.Error{Code: "INVALID_REQUEST", Message: "invalid request body"})
		return false
	}
	return true
}
func riskActor(c *gin.Context) riskadmin.Actor {
	a, _ := adminauth.FromContext(c)
	correlation := c.GetHeader("X-Request-ID")
	if correlation == "" {
		correlation = c.GetString("request_id")
	}
	return riskadmin.Actor{ID: a.ID, Username: a.Username, Scope: adminContentScope(c), IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: correlation}
}
func riskList(c *gin.Context, keys []string) riskadmin.ListQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	values := map[string]string{}
	for _, key := range keys {
		if v := strings.TrimSpace(c.Query(key)); v != "" {
			values[key] = v
		}
	}
	return riskadmin.ListQuery{Page: page, PageSize: size, Values: values, Scope: adminContentScope(c)}
}
func riskError(c *gin.Context, err error) {
	var domain *riskadmin.Error
	if errors.As(err, &domain) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": domain.Message, "data": gin.H{"errorCode": domain.Code}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

var ruleRiskKeys = []string{"code", "name", "ruleType", "riskLevel", "enabled", "business", "updatedFrom", "updatedTo"}
var eventRiskKeys = []string{"eventID", "userID", "username", "rule", "riskType", "riskLevel", "status", "provider", "gameID", "ip", "deviceID", "createdFrom", "createdTo"}
var reviewRiskKeys = []string{"reviewID", "businessType", "userID", "username", "riskLevel", "status", "reviewer", "createdFrom", "createdTo"}
var associationRiskKeys = []string{"userID", "username", "associationType", "relatedUserID", "deviceID", "ip", "riskLevel"}
var blacklistRiskKeys = []string{"type", "value", "reason", "status", "createdBy", "createdFrom", "createdTo", "expiresFrom", "expiresTo"}
var withdrawalRiskKeys = []string{"withdrawalOrderID", "userID", "username", "riskLevel", "riskStatus", "kycStatus", "country", "currency", "amountMin", "amountMax", "createdFrom", "createdTo"}
var operationLogRiskKeys = []string{"operatorName", "action", "targetType", "targetID", "riskLevel", "createdFrom", "createdTo"}
