package routers

import (
	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
	"casino/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		adminFundsRouter(group, handler.NewAdminFundsHandler())
	})
}

func adminFundsRouter(group *gin.RouterGroup, h handler.AdminFundsHandler) {
	g := group.Group("/admin/funds")
	g.Use(adminauth.Middleware())

	g.GET("/wallets", adminauth.Require(adminauth.PermissionFundWalletList), h.ListWallets)
	g.GET("/wallets/:id", adminauth.Require(adminauth.PermissionFundWalletView), h.GetWallet)
	g.PATCH("/wallets/:id/freeze", adminauth.Require(adminauth.PermissionFundWalletFreeze), h.FreezeWallet)
	g.PATCH("/wallets/:id/unfreeze", adminauth.Require(adminauth.PermissionFundWalletUnfreeze), h.UnfreezeWallet)

	g.GET("/transactions", adminauth.Require(adminauth.PermissionFundTransactionList), h.ListTransactions)
	g.GET("/transactions/:id", adminauth.Require(adminauth.PermissionFundTransactionView), h.GetTransaction)

	g.GET("/adjustments", adminauth.Require(adminauth.PermissionFundAdjustmentList), h.ListAdjustments)
	g.GET("/adjustments/:id", adminauth.Require(adminauth.PermissionFundAdjustmentView), h.GetAdjustment)
	g.POST("/adjustments", adminauth.Require(adminauth.PermissionFundAdjustmentCreate), h.CreateAdjustment)
	g.POST("/adjustments/:id/approve", adminauth.Require(adminauth.PermissionFundAdjustmentApprove), h.ApproveAdjustment)
	g.POST("/adjustments/:id/reject", adminauth.Require(adminauth.PermissionFundAdjustmentReject), h.RejectAdjustment)
	g.POST("/adjustments/:id/cancel", adminauth.Require(adminauth.PermissionFundAdjustmentCancel), h.CancelAdjustment)

	g.GET("/deposits", adminauth.Require(adminauth.PermissionFundDepositList), h.ListDeposits)
	g.GET("/deposits/:id", adminauth.Require(adminauth.PermissionFundDepositView), h.GetDeposit)
	g.POST("/deposits/:id/review", adminauth.Require(adminauth.PermissionFundDepositReview), h.ReviewDeposit)
	g.POST("/deposits/:id/confirm", adminauth.Require(adminauth.PermissionFundDepositConfirm), h.ConfirmDeposit)
	g.POST("/deposits/:id/cancel", adminauth.Require(adminauth.PermissionFundDepositCancel), h.CancelDeposit)

	g.GET("/withdrawals", adminauth.Require(adminauth.PermissionFundWithdrawalList), h.ListWithdrawals)
	g.GET("/withdrawals/:id", adminauth.Require(adminauth.PermissionFundWithdrawalView), h.GetWithdrawal)
	g.POST("/withdrawals/:id/risk-review", adminauth.Require(adminauth.PermissionFundWithdrawalRisk), h.RiskReviewWithdrawal)
	g.POST("/withdrawals/:id/finance-review", adminauth.Require(adminauth.PermissionFundWithdrawalFinance), h.FinanceReviewWithdrawal)
	g.POST("/withdrawals/:id/approve", adminauth.Require(adminauth.PermissionFundWithdrawalApprove), h.ApproveWithdrawal)
	g.POST("/withdrawals/:id/reject", adminauth.Require(adminauth.PermissionFundWithdrawalReject), h.RejectWithdrawal)
	g.POST("/withdrawals/:id/retry", adminauth.Require(adminauth.PermissionFundWithdrawalRetry), h.RetryWithdrawal)

	g.GET("/reconciliations", adminauth.Require(adminauth.PermissionFundReconciliationList), h.ListReconciliations)
	g.GET("/reconciliations/:id", adminauth.Require(adminauth.PermissionFundReconciliationView), h.GetReconciliation)
	g.POST("/reconciliations", adminauth.Require(adminauth.PermissionFundReconciliationGenerate), h.GenerateReconciliation)
	g.POST("/reconciliations/:id/differences/:differenceID/resolve", adminauth.Require(adminauth.PermissionFundReconciliationResolve), h.ResolveDifference)

	g.GET("/deposit-methods", adminauth.Require(adminauth.PermissionFundDepositMethodList), h.ListDepositMethods)
	g.GET("/deposit-methods/:id", adminauth.Require(adminauth.PermissionFundDepositMethodList), h.GetDepositMethod)
	g.POST("/deposit-methods", adminauth.Require(adminauth.PermissionFundDepositMethodCreate), h.CreateDepositMethod)
	g.PUT("/deposit-methods/:id", adminauth.Require(adminauth.PermissionFundDepositMethodUpdate), h.UpdateDepositMethod)
	g.DELETE("/deposit-methods/:id", adminauth.Require(adminauth.PermissionFundDepositMethodDelete), h.DeleteDepositMethod)

	g.GET("/deposit-channels", adminauth.Require(adminauth.PermissionFundDepositChannelList), h.ListDepositChannels)
	g.GET("/deposit-channels/:id", adminauth.Require(adminauth.PermissionFundDepositChannelList), h.GetDepositChannel)
	g.POST("/deposit-channels", adminauth.Require(adminauth.PermissionFundDepositChannelCreate), h.CreateDepositChannel)
	g.PUT("/deposit-channels/:id", adminauth.Require(adminauth.PermissionFundDepositChannelUpdate), h.UpdateDepositChannel)
	g.DELETE("/deposit-channels/:id", adminauth.Require(adminauth.PermissionFundDepositChannelDelete), h.DeleteDepositChannel)

	g.GET("/deposit-amount-options", adminauth.Require(adminauth.PermissionFundDepositOptionList), h.ListDepositAmountOptions)
	g.GET("/deposit-amount-options/:id", adminauth.Require(adminauth.PermissionFundDepositOptionList), h.GetDepositAmountOption)
	g.POST("/deposit-amount-options", adminauth.Require(adminauth.PermissionFundDepositOptionCreate), h.CreateDepositAmountOption)
	g.PUT("/deposit-amount-options/:id", adminauth.Require(adminauth.PermissionFundDepositOptionUpdate), h.UpdateDepositAmountOption)
	g.DELETE("/deposit-amount-options/:id", adminauth.Require(adminauth.PermissionFundDepositOptionDelete), h.DeleteDepositAmountOption)
}
