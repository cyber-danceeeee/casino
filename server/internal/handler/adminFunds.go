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
	"casino/internal/depositconfig"
	"casino/internal/fundadmin"
)

type AdminFundsHandler interface {
	ListWallets(*gin.Context)
	GetWallet(*gin.Context)
	FreezeWallet(*gin.Context)
	UnfreezeWallet(*gin.Context)
	ListTransactions(*gin.Context)
	GetTransaction(*gin.Context)
	ListAdjustments(*gin.Context)
	GetAdjustment(*gin.Context)
	CreateAdjustment(*gin.Context)
	ApproveAdjustment(*gin.Context)
	RejectAdjustment(*gin.Context)
	CancelAdjustment(*gin.Context)
	ListDeposits(*gin.Context)
	GetDeposit(*gin.Context)
	ReviewDeposit(*gin.Context)
	ConfirmDeposit(*gin.Context)
	CancelDeposit(*gin.Context)
	ListWithdrawals(*gin.Context)
	GetWithdrawal(*gin.Context)
	RiskReviewWithdrawal(*gin.Context)
	FinanceReviewWithdrawal(*gin.Context)
	ApproveWithdrawal(*gin.Context)
	RejectWithdrawal(*gin.Context)
	RetryWithdrawal(*gin.Context)
	ListReconciliations(*gin.Context)
	GetReconciliation(*gin.Context)
	GenerateReconciliation(*gin.Context)
	ResolveDifference(*gin.Context)
	ListDepositMethods(*gin.Context)
	GetDepositMethod(*gin.Context)
	CreateDepositMethod(*gin.Context)
	UpdateDepositMethod(*gin.Context)
	DeleteDepositMethod(*gin.Context)
	ListDepositChannels(*gin.Context)
	GetDepositChannel(*gin.Context)
	CreateDepositChannel(*gin.Context)
	UpdateDepositChannel(*gin.Context)
	DeleteDepositChannel(*gin.Context)
	ListDepositAmountOptions(*gin.Context)
	GetDepositAmountOption(*gin.Context)
	CreateDepositAmountOption(*gin.Context)
	UpdateDepositAmountOption(*gin.Context)
	DeleteDepositAmountOption(*gin.Context)
}

type adminFundsHandler struct {
	service       *fundadmin.Service
	depositConfig *depositconfig.Service
}

func NewAdminFundsHandler() AdminFundsHandler {
	db := database.GetDB()
	return &adminFundsHandler{service: fundadmin.New(db), depositConfig: depositconfig.New(db)}
}

func (h *adminFundsHandler) ListWallets(c *gin.Context) {
	data, err := h.service.ListWallets(fundListQuery(c, walletQueryKeys))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetWallet(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.service.GetWallet(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) FreezeWallet(c *gin.Context)   { h.walletStatus(c, true) }
func (h *adminFundsHandler) UnfreezeWallet(c *gin.Context) { h.walletStatus(c, false) }
func (h *adminFundsHandler) walletStatus(c *gin.Context, frozen bool) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	var body fundadmin.ReasonCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.SetWalletStatus(id, frozen, body.Reason, fundActor(c)))
}

func (h *adminFundsHandler) ListTransactions(c *gin.Context) {
	data, err := h.service.ListTransactions(fundListQuery(c, transactionQueryKeys))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetTransaction(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.service.GetTransaction(id, adminTenancyScope(c))
	h.write(c, data, err)
}

func (h *adminFundsHandler) ListAdjustments(c *gin.Context) {
	data, err := h.service.ListAdjustments(fundListQuery(c, adjustmentQueryKeys))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetAdjustment(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.service.GetAdjustment(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) CreateAdjustment(c *gin.Context) {
	var body fundadmin.CreateAdjustmentCommand
	if !fundBind(c, &body) {
		return
	}
	body.IdempotencyKey = strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	id, err := h.service.CreateAdjustment(body, fundActor(c))
	if err != nil {
		fundError(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}
func (h *adminFundsHandler) ApproveAdjustment(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	actor, _ := adminauth.FromContext(c)
	large, err := h.service.IsLargeAdjustment(id, adminTenancyScope(c))
	if err != nil {
		fundError(c, err)
		return
	}
	if large && !actor.Can(adminauth.PermissionFundAdjustmentLarge) {
		fundError(c, &fundadmin.Error{Code: "LARGE_AMOUNT_PERMISSION_REQUIRED", Message: "large amount approval permission required"})
		return
	}
	h.adjustmentDecision(c, id, h.service.ApproveAdjustment)
}
func (h *adminFundsHandler) RejectAdjustment(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.adjustmentDecision(c, id, h.service.RejectAdjustment)
}
func (h *adminFundsHandler) CancelAdjustment(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.adjustmentDecision(c, id, h.service.CancelAdjustment)
}
func (h *adminFundsHandler) adjustmentDecision(c *gin.Context, id uint64, fn func(uint64, fundadmin.ReasonCommand, fundadmin.Actor) error) {
	var body fundadmin.ReasonCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, fn(id, body, fundActor(c)))
}

func (h *adminFundsHandler) ListDeposits(c *gin.Context) {
	data, err := h.service.ListDeposits(fundListQuery(c, depositQueryKeys))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetDeposit(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.service.GetDeposit(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) ReviewDeposit(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	var body fundadmin.ReasonCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.ReviewDeposit(id, body, fundActor(c)))
}
func (h *adminFundsHandler) ConfirmDeposit(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	var body fundadmin.DepositConfirmCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.ConfirmDeposit(id, body, fundActor(c)))
}
func (h *adminFundsHandler) CancelDeposit(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	var body fundadmin.ReasonCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.CancelDeposit(id, body, fundActor(c)))
}

func (h *adminFundsHandler) ListWithdrawals(c *gin.Context) {
	data, err := h.service.ListWithdrawals(fundListQuery(c, withdrawalQueryKeys))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetWithdrawal(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.service.GetWithdrawal(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) RiskReviewWithdrawal(c *gin.Context)    { h.withdrawReview(c, "RISK") }
func (h *adminFundsHandler) FinanceReviewWithdrawal(c *gin.Context) { h.withdrawReview(c, "FINANCE") }
func (h *adminFundsHandler) withdrawReview(c *gin.Context, stage string) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	var body fundadmin.ReasonCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.ReviewWithdrawal(id, stage, body, fundActor(c)))
}
func (h *adminFundsHandler) ApproveWithdrawal(c *gin.Context) {
	h.withdrawAction(c, h.service.ApproveWithdrawal)
}
func (h *adminFundsHandler) RejectWithdrawal(c *gin.Context) {
	h.withdrawAction(c, h.service.RejectWithdrawal)
}
func (h *adminFundsHandler) RetryWithdrawal(c *gin.Context) {
	h.withdrawAction(c, h.service.RetryWithdrawal)
}
func (h *adminFundsHandler) withdrawAction(c *gin.Context, fn func(uint64, fundadmin.ReasonCommand, fundadmin.Actor) error) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	var body fundadmin.ReasonCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, fn(id, body, fundActor(c)))
}

func (h *adminFundsHandler) ListReconciliations(c *gin.Context) {
	data, err := h.service.ListReconciliations(fundListQuery(c, reconciliationQueryKeys))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetReconciliation(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.service.GetReconciliation(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GenerateReconciliation(c *gin.Context) {
	var body fundadmin.GenerateReconciliationCommand
	if !fundBind(c, &body) {
		return
	}
	id, err := h.service.GenerateReconciliation(body, fundActor(c))
	if err != nil {
		fundError(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}
func (h *adminFundsHandler) ResolveDifference(c *gin.Context) {
	reconciliationID, ok := fundPathID(c)
	if !ok {
		return
	}
	differenceID, err := strconv.ParseUint(c.Param("differenceID"), 10, 64)
	if err != nil || differenceID == 0 {
		fundError(c, &fundadmin.Error{Code: "INVALID_REQUEST", Message: "invalid difference id"})
		return
	}
	var body fundadmin.ReasonCommand
	if !fundBind(c, &body) {
		return
	}
	h.writeAction(c, h.service.ResolveDifference(reconciliationID, differenceID, body, fundActor(c)))
}

func (h *adminFundsHandler) ListDepositMethods(c *gin.Context) {
	data, err := h.depositConfig.ListMethods(depositConfigList(c, []string{"methodCode", "name", "methodType", "status"}))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetDepositMethod(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.depositConfig.GetMethod(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) CreateDepositMethod(c *gin.Context) { h.saveDepositMethod(c, 0) }
func (h *adminFundsHandler) UpdateDepositMethod(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.saveDepositMethod(c, id)
}
func (h *adminFundsHandler) saveDepositMethod(c *gin.Context, id uint64) {
	var body depositconfig.MethodInput
	if !fundBind(c, &body) {
		return
	}
	savedID, err := h.depositConfig.SaveMethod(id, body, depositConfigActor(c))
	h.write(c, gin.H{"id": savedID}, err)
}
func (h *adminFundsHandler) DeleteDepositMethod(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.writeAction(c, h.depositConfig.DeleteMethod(id, depositConfigActor(c)))
}

func (h *adminFundsHandler) ListDepositChannels(c *gin.Context) {
	data, err := h.depositConfig.ListChannels(depositConfigList(c, []string{"channelCode", "name", "methodCode", "countryCode", "currencyCode", "status"}))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetDepositChannel(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.depositConfig.GetChannel(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) CreateDepositChannel(c *gin.Context) { h.saveDepositChannel(c, 0) }
func (h *adminFundsHandler) UpdateDepositChannel(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.saveDepositChannel(c, id)
}
func (h *adminFundsHandler) saveDepositChannel(c *gin.Context, id uint64) {
	var body depositconfig.ChannelInput
	if !fundBind(c, &body) {
		return
	}
	savedID, err := h.depositConfig.SaveChannel(id, body, depositConfigActor(c))
	h.write(c, gin.H{"id": savedID}, err)
}
func (h *adminFundsHandler) DeleteDepositChannel(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.writeAction(c, h.depositConfig.DeleteChannel(id, depositConfigActor(c)))
}

func (h *adminFundsHandler) ListDepositAmountOptions(c *gin.Context) {
	data, err := h.depositConfig.ListAmountOptions(depositConfigList(c, []string{"optionCode", "status"}))
	h.write(c, data, err)
}
func (h *adminFundsHandler) GetDepositAmountOption(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	data, err := h.depositConfig.GetAmountOption(id, adminTenancyScope(c))
	h.write(c, data, err)
}
func (h *adminFundsHandler) CreateDepositAmountOption(c *gin.Context) {
	h.saveDepositAmountOption(c, 0)
}
func (h *adminFundsHandler) UpdateDepositAmountOption(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.saveDepositAmountOption(c, id)
}
func (h *adminFundsHandler) saveDepositAmountOption(c *gin.Context, id uint64) {
	var body depositconfig.AmountOptionInput
	if !fundBind(c, &body) {
		return
	}
	savedID, err := h.depositConfig.SaveAmountOption(id, body, depositConfigActor(c))
	h.write(c, gin.H{"id": savedID}, err)
}
func (h *adminFundsHandler) DeleteDepositAmountOption(c *gin.Context) {
	id, ok := fundPathID(c)
	if !ok {
		return
	}
	h.writeAction(c, h.depositConfig.DeleteAmountOption(id, depositConfigActor(c)))
}

func (h *adminFundsHandler) write(c *gin.Context, data any, err error) {
	if err != nil {
		fundError(c, err)
		return
	}
	response.Success(c, data)
}
func (h *adminFundsHandler) writeAction(c *gin.Context, err error) {
	if err != nil {
		fundError(c, err)
		return
	}
	response.Success(c)
}

func fundActor(c *gin.Context) fundadmin.Actor {
	actor, _ := adminauth.FromContext(c)
	correlation := c.GetHeader("X-Request-ID")
	if correlation == "" {
		correlation = c.GetString("request_id")
	}
	return fundadmin.Actor{ID: actor.ID, Username: actor.Username, Permissions: actor.Permissions, Scope: adminTenancyScope(c), IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: correlation}
}
func depositConfigActor(c *gin.Context) depositconfig.Actor {
	actor, _ := adminauth.FromContext(c)
	correlation := c.GetHeader("X-Request-ID")
	if correlation == "" {
		correlation = c.GetString("request_id")
	}
	return depositconfig.Actor{ID: actor.ID, Username: actor.Username, Scope: adminTenancyScope(c), IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: correlation}
}
func fundPathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		fundError(c, &fundadmin.Error{Code: "INVALID_REQUEST", Message: "invalid id"})
		return 0, false
	}
	return id, true
}
func fundBind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		fundError(c, &fundadmin.Error{Code: "INVALID_REQUEST", Message: "invalid request body"})
		return false
	}
	return true
}
func fundListQuery(c *gin.Context, keys []string) fundadmin.ListQuery {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	values := map[string]string{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values[key] = value
		}
	}
	return fundadmin.ListQuery{Page: page, PageSize: pageSize, Values: values, Scope: adminTenancyScope(c)}
}
func depositConfigList(c *gin.Context, keys []string) depositconfig.ListQuery {
	query := fundListQuery(c, keys)
	return depositconfig.ListQuery{Page: query.Page, PageSize: query.PageSize, Values: query.Values, Scope: query.Scope}
}
func fundError(c *gin.Context, err error) {
	var configError *depositconfig.DomainError
	if errors.As(err, &configError) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": configError.Message, "data": gin.H{"errorCode": configError.Code}})
		return
	}
	var domain *fundadmin.Error
	if errors.As(err, &domain) {
		c.JSON(http.StatusOK, gin.H{"code": http.StatusBadRequest, "msg": domain.Message, "data": gin.H{"errorCode": domain.Code}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

var walletQueryKeys = []string{"walletID", "userNo", "username", "phone", "currency", "status", "country", "vipLevel", "riskStatus", "availableBalanceMin", "availableBalanceMax", "frozenBalanceMin", "frozenBalanceMax", "createdFrom", "createdTo", "updatedFrom", "updatedTo"}
var transactionQueryKeys = []string{"transactionID", "externalTransactionID", "userNo", "username", "walletID", "transactionType", "direction", "currency", "referenceType", "referenceNo", "providerCode", "status", "operator", "amountMin", "amountMax", "createdFrom", "createdTo", "completedFrom", "completedTo"}
var adjustmentQueryKeys = []string{"adjustmentID", "userNo", "username", "phone", "walletID", "adjustmentType", "currency", "reasonCode", "status", "applicant", "approver", "amountMin", "amountMax", "createdFrom", "createdTo", "completedFrom", "completedTo"}
var depositQueryKeys = []string{"platformOrderID", "channelOrderID", "userNo", "username", "paymentChannel", "paymentMethod", "currency", "status", "amountMin", "amountMax", "createdFrom", "createdTo", "paidFrom", "paidTo", "callbackFrom", "callbackTo"}
var withdrawalQueryKeys = []string{"platformOrderID", "channelOrderID", "userNo", "username", "currency", "paymentMethod", "riskStatus", "kycStatus", "reviewer", "status", "amountMin", "amountMax", "createdFrom", "createdTo", "reviewedFrom", "reviewedTo", "completedFrom", "completedTo"}
var reconciliationQueryKeys = []string{"reconciliationID", "businessDate", "businessType", "channel", "currency", "status", "differenceType", "platformOrderID", "channelOrderID", "generatedFrom", "generatedTo"}
