import {
  fetchAgentReport,
  fetchAgentReportDetail,
  fetchChannelReport,
  fetchChannelReportDetail,
  fetchFinancialReport,
  fetchFinancialReportDetail,
  fetchGameReport,
  fetchGameReportDetail,
  fetchPromotionReport,
  fetchPromotionReportDetail,
  fetchProviderReport,
  fetchProviderReportDetail,
  fetchReportExport,
  fetchReportExports,
  fetchUserReport,
  fetchUserReportDetail,
  type ReportDetail,
  type ReportQuery,
  type ReportResult,
} from '@/api/adminReports'

export type ReportFilter = {
  key: string
  type?: 'text' | 'select' | 'dateRange'
  options?: string[]
}

export type ReportColumn = {
  key: string
  width?: number
  money?: boolean
  percentage?: boolean
  time?: boolean
  status?: boolean
}

export type ReportConfig = {
  type: string
  permission: string
  filters: ReportFilter[]
  summaries: string[]
  columns: ReportColumn[]
  tabs: string[]
  loader: (query: ReportQuery) => Promise<ReportResult>
  detail?: (key: string | number, query: ReportQuery) => Promise<ReportDetail>
  detailKey?: string
  hasChart?: boolean
}

const f = (key: string, type: ReportFilter['type'] = 'text', options?: string[]): ReportFilter => ({ key, type, options })
const c = (key: string, width = 140, kind?: 'money' | 'percentage' | 'time' | 'status'): ReportColumn => ({
  key,
  width,
  ...(kind ? { [kind]: true } : {}),
})
const common = [
  f('dateRange', 'dateRange'),
  f('granularity', 'select', ['day', 'week', 'month']),
  f('country'),
  f('currency'),
]

export const reportConfigs: Record<string, ReportConfig> = {
  'financial-report': {
    type: 'FINANCIAL', permission: 'report:financial:view', loader: fetchFinancialReport,
    detail: (_, query) => fetchFinancialReportDetail(query), detailKey: 'period', hasChart: true,
    filters: [...common, f('paymentChannel'), f('userChannel')],
    summaries: ['depositAmount', 'withdrawalAmount', 'netCashFlow', 'totalBet', 'totalWin', 'ggr', 'bonusCost', 'providerCost', 'paymentFee', 'agentCommission', 'ngr'],
    columns: [c('period', 130), c('currency', 90), c('depositAmount', 145, 'money'), c('withdrawalAmount', 155, 'money'), c('totalBet', 135, 'money'), c('totalWin', 135, 'money'), c('ggr', 125, 'money'), c('bonusCost', 130, 'money'), c('providerCost', 135, 'money'), c('paymentFee', 125, 'money'), c('agentCommission', 150, 'money'), c('ngr', 125, 'money'), c('activeUsers', 115)],
    tabs: ['summary', 'deposits', 'withdrawals', 'betting', 'costs', 'adjustments'],
  },
  'game-report': {
    type: 'GAME', permission: 'report:game:view', loader: fetchGameReport,
    detail: (id, query) => fetchGameReportDetail(Number(id), query), detailKey: 'gameID', hasChart: true,
    filters: [...common, f('provider'), f('gameCode'), f('gameName'), f('providerGameID'), f('gameType'), f('platformCategory'), f('device')],
    summaries: ['gameCount', 'uniquePlayers', 'totalRounds', 'totalBet', 'totalWin', 'ggr', 'actualRTP', 'averageBet', 'maximumWin'],
    columns: [c('gameCode', 170), c('game', 180), c('provider', 145), c('gameType', 125), c('currency', 90), c('uniquePlayers', 125), c('rounds', 100), c('totalBet', 130, 'money'), c('validBet', 130, 'money'), c('totalWin', 130, 'money'), c('ggr', 120, 'money'), c('actualRTP', 115, 'percentage'), c('averageBet', 125, 'money'), c('maximumWin', 130, 'money')],
    tabs: ['overview', 'trend', 'users', 'bettingRecords', 'roundRecords'],
  },
  'provider-report': {
    type: 'PROVIDER', permission: 'report:provider:view', loader: fetchProviderReport,
    detail: (code, query) => fetchProviderReportDetail(String(code), query), detailKey: 'provider', hasChart: true,
    filters: [...common, f('provider'), f('gameType')],
    summaries: ['providerCount', 'totalGames', 'uniquePlayers', 'totalBet', 'totalWin', 'ggr', 'actualRTP', 'callbackSuccessRate', 'launchSuccessRate'],
    columns: [c('provider', 155), c('currency', 90), c('gameCount', 110), c('uniquePlayers', 125), c('totalBet', 130, 'money'), c('totalWin', 130, 'money'), c('ggr', 120, 'money'), c('actualRTP', 115, 'percentage'), c('launchCount', 115), c('launchSuccessRate', 145, 'percentage'), c('callbackCount', 125), c('callbackSuccessRate', 155, 'percentage'), c('averageCallbackLatency', 175)],
    tabs: ['overview', 'games', 'bettingTrend', 'callbackMetrics', 'launchMetrics', 'errors'],
  },
  'user-report': {
    type: 'USER', permission: 'report:user:view', loader: fetchUserReport,
    detail: (id, query) => fetchUserReportDetail(Number(id), query), detailKey: 'userID',
    filters: [...common, f('playerNo'), f('username'), f('vipLevel'), f('registrationChannel'), f('agent'), f('riskStatus')],
    summaries: ['registeredUsers', 'activeUsers', 'depositingUsers', 'firstDepositUsers', 'bettingUsers', 'returningUsers', 'churnedUsers'],
    columns: [c('playerNo', 165), c('username', 145), c('country', 95), c('currency', 90), c('registrationChannel', 160), c('agent', 140), c('totalDeposit', 140, 'money'), c('totalWithdrawal', 150, 'money'), c('totalBet', 130, 'money'), c('totalWin', 130, 'money'), c('userNetResult', 145, 'money'), c('bonusAmount', 135, 'money'), c('lastActiveAt', 175, 'time')],
    tabs: ['overview', 'financialSummary', 'bettingSummary', 'bonusSummary', 'activityTrend'],
  },
  'channel-report': {
    type: 'CHANNEL', permission: 'report:channel:view', loader: fetchChannelReport,
    detail: (id, query) => fetchChannelReportDetail(Number(id), query), detailKey: 'channelID',
    filters: [...common, f('channel'), f('agent'), f('platform')],
    summaries: ['registrations', 'firstDepositUsers', 'depositingUsers', 'depositAmount', 'bettingUsers', 'totalBet', 'ggr', 'conversionRate'],
    columns: [c('channel', 170), c('agent', 145), c('currency', 90), c('registrations', 120), c('firstDepositUsers', 155), c('firstDepositConversion', 175, 'percentage'), c('depositingUsers', 145), c('depositAmount', 140, 'money'), c('bettingUsers', 135), c('totalBet', 130, 'money'), c('ggr', 120, 'money'), c('ngr', 120, 'money')],
    tabs: ['overview', 'registrationFunnel', 'depositFunnel', 'bettingFunnel', 'users', 'trend'],
  },
  'agent-report': {
    type: 'AGENT', permission: 'report:agent:view', loader: fetchAgentReport,
    detail: (id, query) => fetchAgentReportDetail(Number(id), query), detailKey: 'agentID',
    filters: [...common, f('agent'), f('agentLevel'), f('parentAgent'), f('commissionPlan')],
    summaries: ['agentCount', 'attributedUsers', 'firstDepositUsers', 'depositAmount', 'totalBet', 'ggr', 'commissionAmount', 'settlementAmount'],
    columns: [c('agent', 165), c('level', 90), c('parentAgent', 150), c('currency', 90), c('attributedUsers', 145), c('firstDepositUsers', 155), c('depositAmount', 140, 'money'), c('totalBet', 130, 'money'), c('ggr', 120, 'money'), c('commissionAmount', 155, 'money'), c('settledAmount', 145, 'money'), c('pendingSettlement', 165, 'money')],
    tabs: ['overview', 'users', 'revenue', 'commissionRecords', 'settlements', 'trend'],
  },
  'promotion-report': {
    type: 'PROMOTION', permission: 'report:promotion:view', loader: fetchPromotionReport,
    detail: (id, query) => fetchPromotionReportDetail(Number(id), query), detailKey: 'promotionID',
    filters: [...common, f('promotion'), f('promotionType'), f('status')],
    summaries: ['promotionCount', 'participatingUsers', 'claimCount', 'rewardAmount', 'wageringAmount', 'completedUsers', 'expiredUsers', 'promotionGGR', 'promotionROI'],
    columns: [c('promotion', 185), c('type', 145, 'status'), c('currency', 90), c('participatingUsers', 155), c('claimCount', 115), c('rewardAmount', 140, 'money'), c('requiredWagering', 160, 'money'), c('completedWagering', 170, 'money'), c('completionRate', 145, 'percentage'), c('promotionGGR', 145, 'money'), c('promotionROI', 135, 'percentage')],
    tabs: ['overview', 'participants', 'bonusRecords', 'wageringProgress', 'relatedBets', 'trend'],
  },
  'export-tasks': {
    type: 'EXPORT', permission: 'report:export:list', loader: fetchReportExports,
    detail: (id) => fetchReportExport(Number(id)), detailKey: 'id',
    filters: [f('taskID'), f('reportType', 'select', ['FINANCIAL', 'GAME', 'PROVIDER', 'USER', 'CHANNEL', 'AGENT', 'PROMOTION']), f('operator'), f('status', 'select', ['PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'EXPIRED', 'CANCELLED']), f('createdRange', 'dateRange'), f('completedRange', 'dateRange')],
    summaries: [],
    columns: [c('taskID', 190), c('reportType', 135, 'status'), c('filtersSummary', 260), c('fileFormat', 100), c('rowCount', 105), c('status', 120, 'status'), c('operator', 130), c('createdAt', 175, 'time'), c('completedAt', 175, 'time'), c('expiresAt', 175, 'time')],
    tabs: ['task', 'filters'],
  },
}
