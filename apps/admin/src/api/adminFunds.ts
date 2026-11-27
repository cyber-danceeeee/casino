import { requestID } from '@/utils/requestID'
import { apiDelete, apiGet, apiPatch, apiPost, apiPut, http, type ApiResponse } from './http'
import type { PageResult } from './adminUsers'

export type FundQuery = Record<string, string | number | undefined>
export type WalletStatus = 'ACTIVE' | 'FROZEN'
export type AdjustmentStatus = 'PENDING_APPROVAL' | 'COMPLETED' | 'REJECTED' | 'CANCELLED' | 'FAILED'
export type DepositStatus = 'PENDING' | 'SUCCESS' | 'FAILED' | 'CANCELLED' | 'EXPIRED' | 'MANUAL_REVIEW'
export type WithdrawalStatus = 'PENDING' | 'RISK_REVIEW' | 'FINANCE_REVIEW' | 'PROCESSING' | 'SUCCESS' | 'FAILED' | 'REJECTED' | 'CANCELLED'

export type WalletRow = {
  id: number; userID: number; userNo: string; username: string; phoneCountryCode: string; phone: string; countryCode: string; vipLevel: number
  currency: string; availableBalance: string; frozenBalance: string; bonusBalance: string
  withdrawableBalance: string; status: WalletStatus; accountStatus: string; kycStatus: string
  riskStatus: string; lastTransactionAt: string; createdAt: string; updatedAt: string
}

export type TransactionRow = {
  id: number; transactionID: string; externalTransactionID: string; userID: number; userNo: string
  username: string; phoneCountryCode: string; phone: string; walletID: number; transactionType: string; direction: string; amount: string
  currency: string; balanceBefore: string; balanceAfter: string; frozenBalanceBefore: string
  frozenBalanceAfter: string; referenceType: string; referenceID: number; referenceNo: string
  providerCode: string; status: string; operator: string; idempotencyKey: string
  failureReason: string; createdAt: string; completedAt: string; rawMetadata?: string
}

export type AdjustmentRow = {
  id: number; adjustmentID: string; userID: number; userNo: string; username: string; phoneCountryCode: string; phone: string; walletID: number
  adjustmentType: 'CREDIT' | 'DEBIT'; amount: string; currency: string; reasonCode: string
  detailedReason: string; supportingReference: string; internalRemark: string; status: AdjustmentStatus
  requestedBalance: string; balanceBefore: string; balanceAfter: string; applicant: string
  applicantID: number; approver: string; approvalReason: string; walletTransactionID: number
  failureReason: string; createdAt: string; approvedAt: string; completedAt: string
}

export type DepositRow = {
  id: number; platformOrderID: string; channelOrderID: string; userID: number; userNo: string
  username: string; paymentChannel: string; paymentMethod: string; requestedAmount: string
  fee: string; actualReceivedAmount: string; currency: string; status: DepositStatus
  walletTransactionID: number; reviewer: string; reviewReason: string; evidenceReference: string
  channelResponseCode: string; channelResponseMessage: string; failureReason: string
  createdAt: string; expiredAt: string; paidAt: string; callbackAt: string; completedAt: string
}

export type WithdrawalRow = {
  id: number; platformOrderID: string; channelOrderID: string; userID: number; userNo: string
  username: string; requestedAmount: string; fee: string; actualPayoutAmount: string; currency: string
  paymentMethod: string; recipientAccountMasked: string; recipientNameMasked: string; bankName: string
  kycStatus: string; riskStatus: string; status: WithdrawalStatus; auditStatus: string; reviewer: string
  reviewRemark: string; freezeRecordID: number; finalDebitTransactionID: number; refundTransactionID: number
  failureReason: string; createdAt: string; submittedAt: string; reviewedAt: string; approvedAt: string; completedAt: string
}

export type ReconciliationRow = {
  id: number; reconciliationID: string; businessDate: string; businessType: string; channel: string
  currency: string; mode: string; platformOrderCount: number; platformAmount: string
  channelOrderCount: number; channelAmount: string; differenceAmount: string; differenceCount: number
  status: string; generatedBy: string; generatedAt: string; completedAt: string
}

export type FundDetail = Record<string, unknown>

export type DepositConfigResource = 'deposit-methods' | 'deposit-channels' | 'deposit-amount-options'
export type DepositConfigRow = Record<string, unknown> & { id: number; status: number }
export type DepositConfigPayload = Record<string, unknown>

export const fetchWallets = (params: FundQuery) => apiGet<PageResult<WalletRow>>('/admin/funds/wallets', params)
export const fetchWallet = (id: number) => apiGet<FundDetail>(`/admin/funds/wallets/${id}`)
export const updateWalletStatus = (id: number, frozen: boolean, reason: string) => apiPatch<void>(`/admin/funds/wallets/${id}/${frozen ? 'freeze' : 'unfreeze'}`, { reason })

export const fetchTransactions = (params: FundQuery) => apiGet<PageResult<TransactionRow>>('/admin/funds/transactions', params)
export const fetchTransaction = (id: number) => apiGet<FundDetail>(`/admin/funds/transactions/${id}`)

export const fetchAdjustments = (params: FundQuery) => apiGet<PageResult<AdjustmentRow>>('/admin/funds/adjustments', params)
export const fetchAdjustment = (id: number) => apiGet<FundDetail>(`/admin/funds/adjustments/${id}`)
export async function createAdjustment(payload: Record<string, unknown>) {
  const key = requestID()
  const { data } = await http.post<ApiResponse<{ id: number }>>('/admin/funds/adjustments', payload, { headers: { 'Idempotency-Key': key } })
  if (data.code !== 0) throw new Error(data.msg)
  return data.data
}
export const decideAdjustment = (id: number, action: 'approve' | 'reject' | 'cancel', reason: string) => apiPost<void>(`/admin/funds/adjustments/${id}/${action}`, { reason })

export const fetchDeposits = (params: FundQuery) => apiGet<PageResult<DepositRow>>('/admin/funds/deposits', params)
export const fetchDeposit = (id: number) => apiGet<FundDetail>(`/admin/funds/deposits/${id}`)
export const reviewDeposit = (id: number, reason: string) => apiPost<void>(`/admin/funds/deposits/${id}/review`, { reason })
export const confirmDeposit = (id: number, payload: { reason: string; evidenceReference: string; channelOrderID: string; receivedAmount: string }) => apiPost<void>(`/admin/funds/deposits/${id}/confirm`, payload)
export const cancelDeposit = (id: number, reason: string) => apiPost<void>(`/admin/funds/deposits/${id}/cancel`, { reason })

export const fetchWithdrawals = (params: FundQuery) => apiGet<PageResult<WithdrawalRow>>('/admin/funds/withdrawals', params)
export const fetchWithdrawal = (id: number) => apiGet<FundDetail>(`/admin/funds/withdrawals/${id}`)
export const decideWithdrawal = (id: number, action: 'risk-review' | 'finance-review' | 'approve' | 'reject' | 'retry', reason: string) => apiPost<void>(`/admin/funds/withdrawals/${id}/${action}`, { reason })

export const fetchReconciliations = (params: FundQuery) => apiGet<PageResult<ReconciliationRow>>('/admin/funds/reconciliations', params)
export const fetchReconciliation = (id: number) => apiGet<FundDetail>(`/admin/funds/reconciliations/${id}`)
export const generateReconciliation = (payload: { businessDate: string; businessType: string; channel: string; currency: string }) => apiPost<{ id: number }>('/admin/funds/reconciliations', payload)
export const resolveDifference = (id: number, differenceID: number, reason: string) => apiPost<void>(`/admin/funds/reconciliations/${id}/differences/${differenceID}/resolve`, { reason })

export const fetchDepositConfigs = (resource: DepositConfigResource, params: FundQuery) =>
  apiGet<PageResult<DepositConfigRow>>(`/admin/funds/${resource}`, params)
export const fetchDepositConfig = (resource: DepositConfigResource, id: number) =>
  apiGet<FundDetail>(`/admin/funds/${resource}/${id}`)
export const createDepositConfig = (resource: DepositConfigResource, payload: DepositConfigPayload) =>
  apiPost<{ id: number }>(`/admin/funds/${resource}`, payload)
export const updateDepositConfig = (resource: DepositConfigResource, id: number, payload: DepositConfigPayload) =>
  apiPut<{ id: number }>(`/admin/funds/${resource}/${id}`, payload)
export const deleteDepositConfig = (resource: DepositConfigResource, id: number) =>
  apiDelete<void>(`/admin/funds/${resource}/${id}`)
