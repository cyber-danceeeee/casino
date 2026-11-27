package adminauth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"casino/internal/config"
	"casino/internal/database"
)

const actorContextKey = "admin_actor"

const superAdminRoleCode = "super-admin"

const (
	PermissionUserList                      = "user:list"
	PermissionUserView                      = "user:view"
	PermissionUserStatusUpdate              = "user:status:update"
	PermissionUserRestrictionUpdate         = "user:restriction:update"
	PermissionUserRiskUpdate                = "user:risk:update"
	PermissionUserNoteCreate                = "user:note:create"
	PermissionLoginRecordView               = "user:login-record:view"
	PermissionDeviceView                    = "user:device:view"
	PermissionDeviceManage                  = "user:device:manage"
	PermissionSensitiveView                 = "user:sensitive:view"
	PermissionFundWalletList                = "fund:wallet:list"
	PermissionFundWalletView                = "fund:wallet:view"
	PermissionFundWalletFreeze              = "fund:wallet:freeze"
	PermissionFundWalletUnfreeze            = "fund:wallet:unfreeze"
	PermissionFundTransactionList           = "fund:transaction:list"
	PermissionFundTransactionView           = "fund:transaction:view"
	PermissionFundAdjustmentList            = "fund:adjustment:list"
	PermissionFundAdjustmentView            = "fund:adjustment:view"
	PermissionFundAdjustmentCreate          = "fund:adjustment:create"
	PermissionFundAdjustmentApprove         = "fund:adjustment:approve"
	PermissionFundAdjustmentReject          = "fund:adjustment:reject"
	PermissionFundAdjustmentCancel          = "fund:adjustment:cancel"
	PermissionFundAdjustmentLarge           = "fund:adjustment:large-amount"
	PermissionFundDepositList               = "fund:deposit:list"
	PermissionFundDepositView               = "fund:deposit:view"
	PermissionFundDepositReview             = "fund:deposit:review"
	PermissionFundDepositConfirm            = "fund:deposit:confirm"
	PermissionFundDepositCancel             = "fund:deposit:cancel"
	PermissionFundWithdrawalList            = "fund:withdrawal:list"
	PermissionFundWithdrawalView            = "fund:withdrawal:view"
	PermissionFundWithdrawalRisk            = "fund:withdrawal:risk-review"
	PermissionFundWithdrawalFinance         = "fund:withdrawal:finance-review"
	PermissionFundWithdrawalApprove         = "fund:withdrawal:approve"
	PermissionFundWithdrawalReject          = "fund:withdrawal:reject"
	PermissionFundWithdrawalRetry           = "fund:withdrawal:retry"
	PermissionFundWithdrawalSensitive       = "fund:withdrawal:sensitive-view"
	PermissionFundReconciliationList        = "fund:reconciliation:list"
	PermissionFundReconciliationView        = "fund:reconciliation:view"
	PermissionFundReconciliationGenerate    = "fund:reconciliation:generate"
	PermissionFundReconciliationResolve     = "fund:reconciliation:resolve"
	PermissionFundReconciliationExport      = "fund:reconciliation:export"
	PermissionFundDepositMethodList         = "fund:deposit-method:list"
	PermissionFundDepositMethodCreate       = "fund:deposit-method:create"
	PermissionFundDepositMethodUpdate       = "fund:deposit-method:update"
	PermissionFundDepositMethodDelete       = "fund:deposit-method:delete"
	PermissionFundDepositChannelList        = "fund:deposit-channel:list"
	PermissionFundDepositChannelCreate      = "fund:deposit-channel:create"
	PermissionFundDepositChannelUpdate      = "fund:deposit-channel:update"
	PermissionFundDepositChannelDelete      = "fund:deposit-channel:delete"
	PermissionFundDepositOptionList         = "fund:deposit-option:list"
	PermissionFundDepositOptionCreate       = "fund:deposit-option:create"
	PermissionFundDepositOptionUpdate       = "fund:deposit-option:update"
	PermissionFundDepositOptionDelete       = "fund:deposit-option:delete"
	PermissionBetList                       = "bet:list"
	PermissionBetView                       = "bet:view"
	PermissionBetReview                     = "bet:review"
	PermissionBetRoundList                  = "bet:round:list"
	PermissionBetRoundView                  = "bet:round:view"
	PermissionBetRollbackList               = "bet:rollback:list"
	PermissionBetRollbackView               = "bet:rollback:view"
	PermissionBetRollbackReview             = "bet:rollback:review"
	PermissionBetAbnormalList               = "bet:abnormal:list"
	PermissionBetAbnormalView               = "bet:abnormal:view"
	PermissionBetAbnormalAssign             = "bet:abnormal:assign"
	PermissionBetAbnormalReview             = "bet:abnormal:review"
	PermissionBetAbnormalClear              = "bet:abnormal:clear"
	PermissionBetAbnormalRestrict           = "bet:abnormal:restrict"
	PermissionBetAbnormalNote               = "bet:abnormal:note"
	PermissionBetCallbackLogList            = "bet:callback-log:list"
	PermissionBetCallbackLogView            = "bet:callback-log:view"
	PermissionBetReconciliationList         = "bet:reconciliation:list"
	PermissionBetReconciliationView         = "bet:reconciliation:view"
	PermissionBetReconciliationGenerate     = "bet:reconciliation:generate"
	PermissionBetReconciliationResolve      = "bet:reconciliation:resolve"
	PermissionOperationBannerList           = "operation:banner:list"
	PermissionOperationBannerCreate         = "operation:banner:create"
	PermissionOperationBannerUpdate         = "operation:banner:update"
	PermissionOperationBannerDelete         = "operation:banner:delete"
	PermissionOperationAnnouncementList     = "operation:announcement:list"
	PermissionOperationAnnouncementCreate   = "operation:announcement:create"
	PermissionOperationAnnouncementUpdate   = "operation:announcement:update"
	PermissionOperationAnnouncementDelete   = "operation:announcement:delete"
	PermissionOperationPromotionList        = "operation:promotion:list"
	PermissionOperationPromotionCreate      = "operation:promotion:create"
	PermissionOperationPromotionUpdate      = "operation:promotion:update"
	PermissionOperationPromotionStatus      = "operation:promotion:status"
	PermissionOperationBonusList            = "operation:bonus:list"
	PermissionOperationBonusView            = "operation:bonus:view"
	PermissionOperationBonusActivate        = "operation:bonus:activate"
	PermissionOperationBonusCancel          = "operation:bonus:cancel"
	PermissionOperationBonusForfeit         = "operation:bonus:forfeit"
	PermissionOperationVIPList              = "operation:vip:list"
	PermissionOperationVIPUpdate            = "operation:vip:update"
	PermissionOperationVIPAdjust            = "operation:vip:adjust"
	PermissionOperationPromoCodeList        = "operation:promo-code:list"
	PermissionOperationPromoCodeCreate      = "operation:promo-code:create"
	PermissionOperationPromoCodeUpdate      = "operation:promo-code:update"
	PermissionOperationPromoCodeDelete      = "operation:promo-code:delete"
	PermissionOperationNotificationList     = "operation:notification:list"
	PermissionOperationNotificationCreate   = "operation:notification:create"
	PermissionOperationNotificationUpdate   = "operation:notification:update"
	PermissionOperationNotificationSend     = "operation:notification:send"
	PermissionOperationNotificationCancel   = "operation:notification:cancel"
	PermissionOperationsAnnouncementView    = "operations:announcement:view"
	PermissionOperationsAnnouncementEdit    = "operations:announcement:edit"
	PermissionOperationsAnnouncementDelete  = "operations:announcement:delete"
	PermissionOperationsPopupView           = "operations:popup:view"
	PermissionOperationsPopupEdit           = "operations:popup:edit"
	PermissionOperationsPopupDelete         = "operations:popup:delete"
	PermissionOperationAssetList            = "operation:asset:list"
	PermissionOperationAssetUpload          = "operation:asset:upload"
	PermissionOperationAssetUpdate          = "operation:asset:update"
	PermissionOperationAssetDelete          = "operation:asset:delete"
	PermissionOperationMusicList            = "operation:music:list"
	PermissionOperationMusicCreate          = "operation:music:create"
	PermissionOperationMusicUpdate          = "operation:music:update"
	PermissionOperationMusicDelete          = "operation:music:delete"
	PermissionOperationSocialLinkList       = "operation:social-link:list"
	PermissionOperationSocialLinkCreate     = "operation:social-link:create"
	PermissionOperationSocialLinkUpdate     = "operation:social-link:update"
	PermissionOperationSocialLinkDelete     = "operation:social-link:delete"
	PermissionOperationH5NavigationList     = "operation:h5-navigation:list"
	PermissionOperationH5NavigationCreate   = "operation:h5-navigation:create"
	PermissionOperationH5NavigationUpdate   = "operation:h5-navigation:update"
	PermissionOperationH5NavigationDelete   = "operation:h5-navigation:delete"
	PermissionOperationVoucherList          = "operation:voucher:list"
	PermissionOperationVoucherCreate        = "operation:voucher:create"
	PermissionOperationVoucherUpdate        = "operation:voucher:update"
	PermissionOperationVoucherDelete        = "operation:voucher:delete"
	PermissionOperationVoucherTierUpdate    = "operation:voucher:tier:update"
	PermissionOperationVoucherUsageView     = "operation:voucher:usage:view"
	PermissionOperationRegisterRewardView   = "operation:register-reward:view"
	PermissionOperationRegisterRewardUpdate = "operation:register-reward:update"
	PermissionRiskDashboardView             = "risk:dashboard:view"
	PermissionRiskRuleList                  = "risk:rule:list"
	PermissionRiskRuleCreate                = "risk:rule:create"
	PermissionRiskRuleUpdate                = "risk:rule:update"
	PermissionRiskRuleDelete                = "risk:rule:delete"
	PermissionRiskRuleTest                  = "risk:rule:test"
	PermissionRiskRuleStatus                = "risk:rule:status"
	PermissionRiskEventList                 = "risk:event:list"
	PermissionRiskEventView                 = "risk:event:view"
	PermissionRiskEventAssign               = "risk:event:assign"
	PermissionRiskEventReview               = "risk:event:review"
	PermissionRiskEventClear                = "risk:event:clear"
	PermissionRiskEventRestrict             = "risk:event:restrict"
	PermissionRiskEventBlock                = "risk:event:block"
	PermissionRiskEventNote                 = "risk:event:note"
	PermissionRiskReviewList                = "risk:review:list"
	PermissionRiskReviewView                = "risk:review:view"
	PermissionRiskReviewClaim               = "risk:review:claim"
	PermissionRiskReviewAssign              = "risk:review:assign"
	PermissionRiskReviewDecision            = "risk:review:decision"
	PermissionRiskReviewEscalate            = "risk:review:escalate"
	PermissionRiskAssociationList           = "risk:association:list"
	PermissionRiskAssociationView           = "risk:association:view"
	PermissionRiskBlacklistList             = "risk:blacklist:list"
	PermissionRiskBlacklistCreate           = "risk:blacklist:create"
	PermissionRiskBlacklistUpdate           = "risk:blacklist:update"
	PermissionRiskBlacklistRemove           = "risk:blacklist:remove"
	PermissionRiskWithdrawalList            = "risk:withdrawal:list"
	PermissionRiskWithdrawalView            = "risk:withdrawal:view"
	PermissionRiskWithdrawalApprove         = "risk:withdrawal:approve"
	PermissionRiskWithdrawalReject          = "risk:withdrawal:reject"
	PermissionRiskWithdrawalRestrict        = "risk:withdrawal:restrict"
	PermissionRiskOperationLogView          = "risk:operation-log:view"
	PermissionGameProviderList              = "game:provider:list"
	PermissionGameProviderView              = "game:provider:view"
	PermissionGameProviderUpdate            = "game:provider:update"
	PermissionGameProviderTest              = "game:provider:test"
	PermissionGameProviderSync              = "game:provider:sync"
	PermissionGameList                      = "game:list"
	PermissionGameView                      = "game:view"
	PermissionGameUpdate                    = "game:update"
	PermissionGameStatusUpdate              = "game:status:update"
	PermissionGameCategoryAssign            = "game:category:assign"
	PermissionGameTagAssign                 = "game:tag:assign"
	PermissionGameCollectionAssign          = "game:collection:assign"
	PermissionGameDemoLaunch                = "game:demo:launch"
	PermissionGameCategoryList              = "game:category:list"
	PermissionGameCategoryCreate            = "game:category:create"
	PermissionGameCategoryUpdate            = "game:category:update"
	PermissionGameCategoryDelete            = "game:category:delete"
	PermissionGameTagList                   = "game:tag:list"
	PermissionGameTagCreate                 = "game:tag:create"
	PermissionGameTagUpdate                 = "game:tag:update"
	PermissionGameTagDelete                 = "game:tag:delete"
	PermissionGameCollectionList            = "game:collection:list"
	PermissionGameCollectionCreate          = "game:collection:create"
	PermissionGameCollectionUpdate          = "game:collection:update"
	PermissionGameCollectionDelete          = "game:collection:delete"
	PermissionGameSyncLogView               = "game:sync-log:view"
	PermissionGameLaunchLogView             = "game:launch-log:view"
	PermissionGameAssetUpload               = "game:asset:upload"
	PermissionAgentList                     = "agent:list"
	PermissionAgentView                     = "agent:view"
	PermissionAgentCreate                   = "agent:create"
	PermissionAgentUpdate                   = "agent:update"
	PermissionAgentStatus                   = "agent:status"
	PermissionAgentHierarchyView            = "agent:hierarchy:view"
	PermissionAgentHierarchyUpdate          = "agent:hierarchy:update"
	PermissionAgentPlanAssign               = "agent:plan:assign"
	PermissionAgentChannelList              = "agent:channel:list"
	PermissionAgentChannelCreate            = "agent:channel:create"
	PermissionAgentChannelUpdate            = "agent:channel:update"
	PermissionAgentChannelStatus            = "agent:channel:status"
	PermissionAgentCommissionPlanList       = "agent:commission-plan:list"
	PermissionAgentCommissionPlanCreate     = "agent:commission-plan:create"
	PermissionAgentCommissionPlanUpdate     = "agent:commission-plan:update"
	PermissionAgentCommissionPlanDelete     = "agent:commission-plan:delete"
	PermissionAgentCommissionRecordList     = "agent:commission-record:list"
	PermissionAgentCommissionRecordView     = "agent:commission-record:view"
	PermissionAgentCommissionRecordAdjust   = "agent:commission-record:adjust"
	PermissionAgentSettlementList           = "agent:settlement:list"
	PermissionAgentSettlementView           = "agent:settlement:view"
	PermissionAgentSettlementGenerate       = "agent:settlement:generate"
	PermissionAgentSettlementSubmit         = "agent:settlement:submit"
	PermissionAgentSettlementApprove        = "agent:settlement:approve"
	PermissionAgentSettlementReject         = "agent:settlement:reject"
	PermissionAgentSettlementPay            = "agent:settlement:pay"
	PermissionAgentSettlementCancel         = "agent:settlement:cancel"
	PermissionAgentSettlementExport         = "agent:settlement:export"
	PermissionAgentWalletList               = "agent:wallet:list"
	PermissionAgentWalletView               = "agent:wallet:view"
	PermissionAgentWalletFreeze             = "agent:wallet:freeze"
	PermissionAgentWalletAdjust             = "agent:wallet:adjust"
	PermissionAgentOperationLogView         = "agent:operation-log:view"
	PermissionAgentAttributionUpdate        = "agent:attribution:update"
	PermissionReportFinancialView           = "report:financial:view"
	PermissionReportGameView                = "report:game:view"
	PermissionReportProviderView            = "report:provider:view"
	PermissionReportUserView                = "report:user:view"
	PermissionReportChannelView             = "report:channel:view"
	PermissionReportAgentView               = "report:agent:view"
	PermissionReportPromotionView           = "report:promotion:view"
	PermissionReportExportCreate            = "report:export:create"
	PermissionReportExportList              = "report:export:list"
	PermissionReportExportDownload          = "report:export:download"
	PermissionReportExportRetry             = "report:export:retry"
	PermissionReportExportCancel            = "report:export:cancel"
	PermissionReportSensitiveView           = "report:sensitive:view"
	PermissionDashboardView                 = "dashboard:view"
	PermissionDashboardFinancialView        = "dashboard:financial:view"
	PermissionDashboardUserView             = "dashboard:user:view"
	PermissionDashboardGameView             = "dashboard:game:view"
	PermissionDashboardProviderView         = "dashboard:provider:view"
	PermissionDashboardChannelView          = "dashboard:channel:view"
	PermissionDashboardAgentView            = "dashboard:agent:view"
	PermissionDashboardPromotionView        = "dashboard:promotion:view"
	PermissionDashboardRiskView             = "dashboard:risk:view"
	PermissionDashboardRealtimeView         = "dashboard:realtime:view"
	PermissionDashboardAlertView            = "dashboard:alert:view"
	PermissionDashboardSensitiveView        = "dashboard:sensitive:view"
	PermissionSystemAdminList               = "system:admin:list"
	PermissionSystemAdminView               = "system:admin:view"
	PermissionSystemAdminCreate             = "system:admin:create"
	PermissionSystemAdminUpdate             = "system:admin:update"
	PermissionSystemAdminStatus             = "system:admin:status"
	PermissionSystemAdminResetPassword      = "system:admin:reset-password"
	PermissionSystemAdminResetMFA           = "system:admin:reset-mfa"
	PermissionSystemAdminForceLogout        = "system:admin:force-logout"
	PermissionSystemAdminAssignRole         = "system:admin:assign-role"
	PermissionSystemRoleList                = "system:role:list"
	PermissionSystemRoleCreate              = "system:role:create"
	PermissionSystemRoleUpdate              = "system:role:update"
	PermissionSystemRoleDelete              = "system:role:delete"
	PermissionSystemRoleAssignPermission    = "system:role:assign-permission"
	PermissionSystemRoleAssignMenu          = "system:role:assign-menu"
	PermissionSystemPermissionList          = "system:permission:list"
	PermissionSystemPermissionCreate        = "system:permission:create"
	PermissionSystemPermissionUpdate        = "system:permission:update"
	PermissionSystemPermissionDelete        = "system:permission:delete"
	PermissionSystemDepartmentList          = "system:department:list"
	PermissionSystemDepartmentCreate        = "system:department:create"
	PermissionSystemDepartmentUpdate        = "system:department:update"
	PermissionSystemDepartmentDelete        = "system:department:delete"
	PermissionSystemMenuList                = "system:menu:list"
	PermissionSystemMenuCreate              = "system:menu:create"
	PermissionSystemMenuUpdate              = "system:menu:update"
	PermissionSystemMenuDelete              = "system:menu:delete"
	PermissionSystemConfigList              = "system:config:list"
	PermissionSystemConfigUpdate            = "system:config:update"
	PermissionSystemConfigSecretUpdate      = "system:config:secret-update"
	PermissionSystemPopupComponentList      = "system:popup-component:list"
	PermissionSystemPopupComponentCreate    = "system:popup-component:create"
	PermissionSystemPopupComponentUpdate    = "system:popup-component:update"
	PermissionSystemPopupComponentDelete    = "system:popup-component:delete"
	PermissionSystemCountryList             = "system:country:list"
	PermissionSystemCountryView             = "system:country:view"
	PermissionSystemCountryCreate           = "system:country:create"
	PermissionSystemCountryUpdate           = "system:country:update"
	PermissionSystemCountryDelete           = "system:country:delete"
	PermissionSystemCountryStatus           = "system:country:status"
	PermissionSystemCountryBranding         = "system:country:branding"
	PermissionSystemCountrySetDefault       = "system:country:set-default"
	PermissionSystemAuditLogView            = "system:audit-log:view"
	PermissionSystemLoginLogView            = "system:login-log:view"
	PermissionSystemTaskList                = "system:task:list"
	PermissionSystemTaskCreate              = "system:task:create"
	PermissionSystemTaskUpdate              = "system:task:update"
	PermissionSystemTaskDelete              = "system:task:delete"
	PermissionSystemTaskRun                 = "system:task:run"
	PermissionSystemMonitoringView          = "system:monitoring:view"
	PermissionSystemEventList               = "system:event:list"
	PermissionSystemEventView               = "system:event:view"
	PermissionSystemEventDispatch           = "system:event:dispatch"
	PermissionSystemEventReplay             = "system:event:replay"

	PermissionSystemSiteView           = "system:site:view"
	PermissionSystemSiteUpdate         = "system:site:update"
	PermissionSystemSiteStatus         = "system:site:status"
	PermissionSupportPlayerRead        = "support:player:read"
	PermissionSupportTimelineRead      = "support:timeline:read"
	PermissionSupportCaseRead          = "support:case:read"
	PermissionSupportCaseWrite         = "support:case:write"
	PermissionSupportExport            = "support:export"
	PermissionMonitorPaymentRead       = "monitor:payment:read"
	PermissionMonitorPaymentReview     = "monitor:payment:review"
	PermissionMonitorProviderRead      = "monitor:provider:read"
	PermissionOperationalNoteWrite     = "operations:note:write"
	PermissionMaintenanceSiteRead      = "maintenance:site:read"
	PermissionMaintenanceSiteWrite     = "maintenance:site:write"
	PermissionMaintenanceProviderRead  = "maintenance:provider:read"
	PermissionMaintenanceProviderWrite = "maintenance:provider:write"
	PermissionMaintenanceGameRead      = "maintenance:game:read"
	PermissionMaintenanceGameWrite     = "maintenance:game:write"
	PermissionHealthSystemRead         = "health:system:read"
	PermissionJobsRead                 = "jobs:read"
	PermissionJobsRetry                = "jobs:retry"
	PermissionJobsExecute              = "jobs:execute"
	PermissionAlertsRead               = "alerts:read"
	PermissionAlertsAcknowledge        = "alerts:acknowledge"
	PermissionAlertsResolve            = "alerts:resolve"
	PermissionBackupRead               = "backup:read"
	PermissionBackupCreate             = "backup:create"
	PermissionBackupRestore            = "backup:restore"
	PermissionBackupDelete             = "backup:delete"
)

var userCenterPermissions = []string{
	PermissionUserList,
	PermissionUserView,
	PermissionUserStatusUpdate,
	PermissionUserRestrictionUpdate,
	PermissionUserRiskUpdate,
	PermissionUserNoteCreate,
	PermissionLoginRecordView,
	PermissionDeviceView,
	PermissionDeviceManage,
	PermissionSensitiveView,
	PermissionGameProviderList,
	PermissionGameProviderView,
	PermissionGameProviderUpdate,
	PermissionGameProviderTest,
	PermissionGameProviderSync,
	PermissionGameList,
	PermissionGameView,
	PermissionGameUpdate,
	PermissionGameStatusUpdate,
	PermissionGameCategoryAssign,
	PermissionGameTagAssign,
	PermissionGameCollectionAssign,
	PermissionGameDemoLaunch,
	PermissionGameCategoryList,
	PermissionGameCategoryCreate,
	PermissionGameCategoryUpdate,
	PermissionGameCategoryDelete,
	PermissionGameTagList,
	PermissionGameTagCreate,
	PermissionGameTagUpdate,
	PermissionGameTagDelete,
	PermissionGameCollectionList,
	PermissionGameCollectionCreate,
	PermissionGameCollectionUpdate,
	PermissionGameCollectionDelete,
	PermissionGameSyncLogView,
	PermissionGameLaunchLogView,
	PermissionGameAssetUpload,
	PermissionAgentList,
	PermissionAgentView,
	PermissionAgentCreate,
	PermissionAgentUpdate,
	PermissionAgentStatus,
	PermissionAgentHierarchyView,
	PermissionAgentHierarchyUpdate,
	PermissionAgentPlanAssign,
	PermissionAgentChannelList,
	PermissionAgentChannelCreate,
	PermissionAgentChannelUpdate,
	PermissionAgentChannelStatus,
	PermissionAgentCommissionPlanList,
	PermissionAgentCommissionPlanCreate,
	PermissionAgentCommissionPlanUpdate,
	PermissionAgentCommissionPlanDelete,
	PermissionAgentCommissionRecordList,
	PermissionAgentCommissionRecordView,
	PermissionAgentCommissionRecordAdjust,
	PermissionAgentSettlementList,
	PermissionAgentSettlementView,
	PermissionAgentSettlementGenerate,
	PermissionAgentSettlementSubmit,
	PermissionAgentSettlementApprove,
	PermissionAgentSettlementReject,
	PermissionAgentSettlementPay,
	PermissionAgentSettlementCancel,
	PermissionAgentSettlementExport,
	PermissionAgentWalletList,
	PermissionAgentWalletView,
	PermissionAgentWalletFreeze,
	PermissionAgentWalletAdjust,
	PermissionAgentOperationLogView,
	PermissionAgentAttributionUpdate,
	PermissionReportFinancialView,
	PermissionReportGameView,
	PermissionReportProviderView,
	PermissionReportUserView,
	PermissionReportChannelView,
	PermissionReportAgentView,
	PermissionReportPromotionView,
	PermissionReportExportCreate,
	PermissionReportExportList,
	PermissionReportExportDownload,
	PermissionReportExportRetry,
	PermissionReportExportCancel,
	PermissionReportSensitiveView,
	PermissionDashboardView,
	PermissionDashboardFinancialView,
	PermissionDashboardUserView,
	PermissionDashboardGameView,
	PermissionDashboardProviderView,
	PermissionDashboardChannelView,
	PermissionDashboardAgentView,
	PermissionDashboardPromotionView,
	PermissionDashboardRiskView,
	PermissionDashboardRealtimeView,
	PermissionDashboardAlertView,
	PermissionDashboardSensitiveView,
	PermissionFundWalletList,
	PermissionFundWalletView,
	PermissionFundWalletFreeze,
	PermissionFundWalletUnfreeze,
	PermissionFundTransactionList,
	PermissionFundTransactionView,
	PermissionFundAdjustmentList,
	PermissionFundAdjustmentView,
	PermissionFundAdjustmentCreate,
	PermissionFundAdjustmentApprove,
	PermissionFundAdjustmentReject,
	PermissionFundAdjustmentCancel,
	PermissionFundAdjustmentLarge,
	PermissionFundDepositList,
	PermissionFundDepositView,
	PermissionFundDepositReview,
	PermissionFundDepositConfirm,
	PermissionFundDepositCancel,
	PermissionFundWithdrawalList,
	PermissionFundWithdrawalView,
	PermissionFundWithdrawalRisk,
	PermissionFundWithdrawalFinance,
	PermissionFundWithdrawalApprove,
	PermissionFundWithdrawalReject,
	PermissionFundWithdrawalRetry,
	PermissionFundWithdrawalSensitive,
	PermissionFundReconciliationList,
	PermissionFundReconciliationView,
	PermissionFundReconciliationGenerate,
	PermissionFundReconciliationResolve,
	PermissionFundReconciliationExport,
	PermissionFundDepositMethodList,
	PermissionFundDepositMethodCreate,
	PermissionFundDepositMethodUpdate,
	PermissionFundDepositMethodDelete,
	PermissionFundDepositChannelList,
	PermissionFundDepositChannelCreate,
	PermissionFundDepositChannelUpdate,
	PermissionFundDepositChannelDelete,
	PermissionFundDepositOptionList,
	PermissionFundDepositOptionCreate,
	PermissionFundDepositOptionUpdate,
	PermissionFundDepositOptionDelete,
	PermissionBetList,
	PermissionBetView,
	PermissionBetReview,
	PermissionBetRoundList,
	PermissionBetRoundView,
	PermissionBetRollbackList,
	PermissionBetRollbackView,
	PermissionBetRollbackReview,
	PermissionBetAbnormalList,
	PermissionBetAbnormalView,
	PermissionBetAbnormalAssign,
	PermissionBetAbnormalReview,
	PermissionBetAbnormalClear,
	PermissionBetAbnormalRestrict,
	PermissionBetAbnormalNote,
	PermissionBetCallbackLogList,
	PermissionBetCallbackLogView,
	PermissionBetReconciliationList,
	PermissionBetReconciliationView,
	PermissionBetReconciliationGenerate,
	PermissionBetReconciliationResolve,
	PermissionOperationBannerList,
	PermissionOperationBannerCreate,
	PermissionOperationBannerUpdate,
	PermissionOperationBannerDelete,
	PermissionOperationAnnouncementList,
	PermissionOperationAnnouncementCreate,
	PermissionOperationAnnouncementUpdate,
	PermissionOperationAnnouncementDelete,
	PermissionOperationPromotionList,
	PermissionOperationPromotionCreate,
	PermissionOperationPromotionUpdate,
	PermissionOperationPromotionStatus,
	PermissionOperationBonusList,
	PermissionOperationBonusView,
	PermissionOperationBonusActivate,
	PermissionOperationBonusCancel,
	PermissionOperationBonusForfeit,
	PermissionOperationVIPList,
	PermissionOperationVIPUpdate,
	PermissionOperationVIPAdjust,
	PermissionOperationPromoCodeList,
	PermissionOperationPromoCodeCreate,
	PermissionOperationPromoCodeUpdate,
	PermissionOperationPromoCodeDelete,
	PermissionOperationNotificationList,
	PermissionOperationNotificationCreate,
	PermissionOperationNotificationUpdate,
	PermissionOperationNotificationSend,
	PermissionOperationNotificationCancel,
	PermissionOperationsAnnouncementView,
	PermissionOperationsAnnouncementEdit,
	PermissionOperationsAnnouncementDelete,
	PermissionOperationsPopupView,
	PermissionOperationsPopupEdit,
	PermissionOperationsPopupDelete,
	PermissionOperationAssetList,
	PermissionOperationAssetUpload,
	PermissionOperationAssetUpdate,
	PermissionOperationAssetDelete,
	PermissionOperationMusicList,
	PermissionOperationMusicCreate,
	PermissionOperationMusicUpdate,
	PermissionOperationMusicDelete,
	PermissionOperationSocialLinkList,
	PermissionOperationSocialLinkCreate,
	PermissionOperationSocialLinkUpdate,
	PermissionOperationSocialLinkDelete,
	PermissionOperationH5NavigationList,
	PermissionOperationH5NavigationCreate,
	PermissionOperationH5NavigationUpdate,
	PermissionOperationH5NavigationDelete,
	PermissionOperationVoucherList,
	PermissionOperationVoucherCreate,
	PermissionOperationVoucherUpdate,
	PermissionOperationVoucherDelete,
	PermissionOperationVoucherTierUpdate,
	PermissionOperationVoucherUsageView,
	PermissionOperationRegisterRewardView,
	PermissionOperationRegisterRewardUpdate,
	PermissionRiskDashboardView,
	PermissionRiskRuleList,
	PermissionRiskRuleCreate,
	PermissionRiskRuleUpdate,
	PermissionRiskRuleDelete,
	PermissionRiskRuleTest,
	PermissionRiskRuleStatus,
	PermissionRiskEventList,
	PermissionRiskEventView,
	PermissionRiskEventAssign,
	PermissionRiskEventReview,
	PermissionRiskEventClear,
	PermissionRiskEventRestrict,
	PermissionRiskEventBlock,
	PermissionRiskEventNote,
	PermissionRiskReviewList,
	PermissionRiskReviewView,
	PermissionRiskReviewClaim,
	PermissionRiskReviewAssign,
	PermissionRiskReviewDecision,
	PermissionRiskReviewEscalate,
	PermissionRiskAssociationList,
	PermissionRiskAssociationView,
	PermissionRiskBlacklistList,
	PermissionRiskBlacklistCreate,
	PermissionRiskBlacklistUpdate,
	PermissionRiskBlacklistRemove,
	PermissionRiskWithdrawalList,
	PermissionRiskWithdrawalView,
	PermissionRiskWithdrawalApprove,
	PermissionRiskWithdrawalReject,
	PermissionRiskWithdrawalRestrict,
	PermissionRiskOperationLogView,
	PermissionSystemAdminList,
	PermissionSystemAdminView,
	PermissionSystemAdminCreate,
	PermissionSystemAdminUpdate,
	PermissionSystemAdminStatus,
	PermissionSystemAdminResetPassword,
	PermissionSystemAdminResetMFA,
	PermissionSystemAdminForceLogout,
	PermissionSystemAdminAssignRole,
	PermissionSystemRoleList,
	PermissionSystemRoleCreate,
	PermissionSystemRoleUpdate,
	PermissionSystemRoleDelete,
	PermissionSystemRoleAssignPermission,
	PermissionSystemRoleAssignMenu,
	PermissionSystemPermissionList,
	PermissionSystemPermissionCreate,
	PermissionSystemPermissionUpdate,
	PermissionSystemPermissionDelete,
	PermissionSystemDepartmentList,
	PermissionSystemDepartmentCreate,
	PermissionSystemDepartmentUpdate,
	PermissionSystemDepartmentDelete,
	PermissionSystemMenuList,
	PermissionSystemMenuCreate,
	PermissionSystemMenuUpdate,
	PermissionSystemMenuDelete,
	PermissionSystemConfigList,
	PermissionSystemConfigUpdate,
	PermissionSystemConfigSecretUpdate,
	PermissionSystemPopupComponentList,
	PermissionSystemPopupComponentCreate,
	PermissionSystemPopupComponentUpdate,
	PermissionSystemPopupComponentDelete,
	PermissionSystemCountryList,
	PermissionSystemCountryView,
	PermissionSystemCountryCreate,
	PermissionSystemCountryUpdate,
	PermissionSystemCountryDelete,
	PermissionSystemCountryStatus,
	PermissionSystemCountryBranding,
	PermissionSystemCountrySetDefault,
	PermissionSystemAuditLogView,
	PermissionSystemLoginLogView,
	PermissionSystemTaskList,
	PermissionSystemTaskCreate,
	PermissionSystemTaskUpdate,
	PermissionSystemTaskDelete,
	PermissionSystemTaskRun,
	PermissionSystemMonitoringView,
	PermissionSystemEventList,
	PermissionSystemEventView,
	PermissionSystemEventDispatch,
	PermissionSystemEventReplay,

	PermissionSystemSiteView,
	PermissionSystemSiteUpdate,
	PermissionSystemSiteStatus,
	PermissionSupportPlayerRead,
	PermissionSupportTimelineRead,
	PermissionSupportCaseRead,
	PermissionSupportCaseWrite,
	PermissionSupportExport,
	PermissionMonitorPaymentRead,
	PermissionMonitorPaymentReview,
	PermissionMonitorProviderRead,
	PermissionOperationalNoteWrite,
	PermissionMaintenanceSiteRead,
	PermissionMaintenanceSiteWrite,
	PermissionMaintenanceProviderRead,
	PermissionMaintenanceProviderWrite,
	PermissionMaintenanceGameRead,
	PermissionMaintenanceGameWrite,
	PermissionHealthSystemRead,
	PermissionJobsRead,
	PermissionJobsRetry,
	PermissionJobsExecute,
	PermissionAlertsRead,
	PermissionAlertsAcknowledge,
	PermissionAlertsResolve,
	PermissionBackupRead,
	PermissionBackupCreate,
	PermissionBackupRestore,
	PermissionBackupDelete,
}

type Actor struct {
	ID          uint64 `json:"id"`
	SessionID   uint64 `json:"sessionID"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`

	SiteID             uint64   `json:"siteID"`
	DefaultSiteID      uint64   `json:"defaultSiteID"`
	Platform           bool     `json:"platform"`
	AdminType          string   `json:"adminType"`
	MustChangePassword bool     `json:"mustChangePassword"`
	RoleCodes          []string `json:"roleCodes"`
	Permissions        []string `json:"permissions"`
}

func (a Actor) Can(permission string) bool {
	for _, value := range a.Permissions {
		if value == permission || value == "*" {
			return true
		}
	}
	return false
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, ok := resolveActor(c)
		if !ok {
			abort(c, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "authentication required")
			return
		}
		c.Set(actorContextKey, actor)
		if actor.MustChangePassword && !passwordChangeAllowedPath(c.Request.URL.Path) {
			abort(c, http.StatusForbidden, "PASSWORD_CHANGE_REQUIRED", "password change required")
			return
		}
		c.Next()
	}
}

func passwordChangeAllowedPath(path string) bool {
	return strings.Contains(path, "/admin/account/") ||
		strings.HasSuffix(path, "/admin/auth/me") ||
		strings.HasSuffix(path, "/admin/auth/logout") ||
		strings.HasSuffix(path, "/admin/me") ||
		strings.HasSuffix(path, "/admin/system/navigation")
}

func Require(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, ok := FromContext(c)
		if !ok || !actor.Can(permission) {
			abort(c, http.StatusForbidden, "PERMISSION_DENIED", "permission denied")
			return
		}
		c.Next()
	}
}

func RequireAny(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, ok := FromContext(c)
		if ok {
			for _, permission := range permissions {
				if actor.Can(permission) {
					c.Next()
					return
				}
			}
		}
		abort(c, http.StatusForbidden, "PERMISSION_DENIED", "permission denied")
	}
}

func FromContext(c *gin.Context) (Actor, bool) {
	value, ok := c.Get(actorContextKey)
	if !ok {
		return Actor{}, false
	}
	actor, ok := value.(Actor)
	return actor, ok
}

func ResolveActor(c *gin.Context) (Actor, bool) {
	return resolveActor(c)
}

func StoreActor(c *gin.Context, actor Actor) {
	c.Set(actorContextKey, actor)
}

func AllUserCenterPermissions() []string {
	result := make([]string, len(userCenterPermissions))
	copy(result, userCenterPermissions)
	return result
}

func resolveActor(c *gin.Context) (Actor, bool) {
	token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	if token == "" && config.Get().App.Env != "prod" && config.Get().AdminAuth.AllowDevBypass {
		return Actor{ID: 1, Username: "development-admin", DisplayName: "Development Administrator", Platform: true, AdminType: "PLATFORM", Permissions: []string{"*"}}, true
	}
	if token == "" {
		return Actor{}, false
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	row := struct {
		ID          uint64 `gorm:"column:id"`
		SessionID   uint64 `gorm:"column:session_id"`
		Username    string `gorm:"column:username"`
		DisplayName string `gorm:"column:display_name"`

		SiteID             uint64 `gorm:"column:site_id"`
		MustChangePassword bool   `gorm:"column:must_change_password"`
	}{}
	err := database.GetDB().Table("game_admin_sessions AS s").
		Select("a.id, s.id AS session_id, a.username, a.display_name, a.site_id, a.must_change_password").
		Joins("INNER JOIN game_admin_accounts AS a ON a.id = s.admin_id AND a.deleted_at IS NULL AND a.status = 1").
		Where("s.token_hash = ? AND s.revoked_at IS NULL AND s.expired_at > ? AND s.deleted_at IS NULL", tokenHash, time.Now()).
		Take(&row).Error
	if err != nil {
		return Actor{}, false
	}
	_ = database.GetDB().Table("game_admin_sessions").Where("id=? AND (last_active_at IS NULL OR last_active_at < ?)", row.SessionID, time.Now().Add(-time.Minute)).
		Updates(map[string]any{"last_active_at": time.Now(), "updated_at": time.Now()}).Error

	roleCodes := []string{}
	_ = database.GetDB().Table("game_admin_account_roles AS ar").Distinct("r.code").
		Joins("INNER JOIN game_admin_roles AS r ON r.id=ar.role_id AND r.status=1 AND r.deleted_at IS NULL").
		Where("ar.admin_id=? AND ar.deleted_at IS NULL", row.ID).Order("r.code").Pluck("r.code", &roleCodes).Error
	makeActor := func(permissions []string) Actor {
		return Actor{
			ID: row.ID, SessionID: row.SessionID, Username: row.Username, DisplayName: row.DisplayName,
			SiteID: 1, DefaultSiteID: 1,
			Platform: false, AdminType: "SITE", MustChangePassword: row.MustChangePassword,
			RoleCodes: roleCodes, Permissions: permissions,
		}
	}

	var superAdminCount int64
	_ = database.GetDB().Table("game_admin_account_roles AS ar").
		Joins("INNER JOIN game_admin_roles AS r ON r.id = ar.role_id AND r.status = 1 AND r.system_protected = 1 AND r.deleted_at IS NULL").
		Where("ar.admin_id = ? AND ar.deleted_at IS NULL AND r.code = ?", row.ID, superAdminRoleCode).
		Count(&superAdminCount).Error
	if superAdminCount > 0 {
		permissions := []string{}
		_ = database.GetDB().Table("game_admin_permissions").
			Where("enabled = 1 AND deleted_at IS NULL").
			Order("code").
			Pluck("code", &permissions).Error
		permissions = append([]string{"*"}, permissions...)
		return makeActor(permissions), true
	}

	permissions := []string{}
	_ = database.GetDB().Table("game_admin_account_roles AS ar").
		Distinct("p.code").
		Joins("INNER JOIN game_admin_roles AS r ON r.id = ar.role_id AND r.status = 1 AND r.deleted_at IS NULL").
		Joins("INNER JOIN game_admin_role_permissions AS rp ON rp.role_id = r.id AND rp.deleted_at IS NULL").
		Joins("INNER JOIN game_admin_permissions AS p ON p.id = rp.permission_id AND p.enabled = 1 AND p.deleted_at IS NULL").
		Where("ar.admin_id = ? AND ar.deleted_at IS NULL", row.ID).
		Pluck("p.code", &permissions).Error
	return makeActor(permissions), true
}

func abort(c *gin.Context, status int, code string, message string) {
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code": status,
		"msg":  message,
		"data": gin.H{"errorCode": code},
	})
}
