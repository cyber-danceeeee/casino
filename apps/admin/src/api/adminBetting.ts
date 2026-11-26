import { apiGet, apiPost } from './http'
import type { PageResult } from './adminUsers'

export type BettingQuery = Record<string, string | number | undefined>
export type BettingRecord = Record<string, unknown> & { id: number }
export type BettingDetail = Record<string, unknown>

export const fetchBets = (params: BettingQuery) => apiGet<PageResult<BettingRecord>>('/admin/betting/bets', params)
export const fetchBet = (id: number) => apiGet<BettingDetail>(`/admin/betting/bets/${id}`)
export const markBetForReview = (id: number, reason: string) => apiPost<void>(`/admin/betting/bets/${id}/review`, { reason })

export const fetchRounds = (params: BettingQuery) => apiGet<PageResult<BettingRecord>>('/admin/betting/rounds', params)
export const fetchRound = (id: number) => apiGet<BettingDetail>(`/admin/betting/rounds/${id}`)

export const fetchRollbacks = (params: BettingQuery) => apiGet<PageResult<BettingRecord>>('/admin/betting/rollbacks', params)
export const fetchRollback = (id: number) => apiGet<BettingDetail>(`/admin/betting/rollbacks/${id}`)
export const investigateRollback = (id: number, reason: string) => apiPost<void>(`/admin/betting/rollbacks/${id}/investigate`, { reason })

export const fetchAbnormalBets = (params: BettingQuery) => apiGet<PageResult<BettingRecord>>('/admin/betting/abnormal-bets', params)
export const fetchAbnormalBet = (id: number) => apiGet<BettingDetail>(`/admin/betting/abnormal-bets/${id}`)
export const assignAbnormalBet = (id: number, payload: { reviewerID: number; reviewer: string; reason: string }) => apiPost<void>(`/admin/betting/abnormal-bets/${id}/assign`, payload)
export const reviewAbnormalBet = (id: number, status: string, reason: string) => apiPost<void>(`/admin/betting/abnormal-bets/${id}/review`, { status, reason })
export const clearAbnormalBet = (id: number, reason: string) => apiPost<void>(`/admin/betting/abnormal-bets/${id}/clear`, { reason })
export const restrictAbnormalBet = (id: number, reason: string) => apiPost<void>(`/admin/betting/abnormal-bets/${id}/restrict`, { reason })
export const addAbnormalNote = (id: number, content: string) => apiPost<void>(`/admin/betting/abnormal-bets/${id}/notes`, { content })

export const fetchCallbackLogs = (params: BettingQuery) => apiGet<PageResult<BettingRecord>>('/admin/betting/provider-callback-logs', params)
export const fetchCallbackLog = (id: number) => apiGet<BettingDetail>(`/admin/betting/provider-callback-logs/${id}`)
export const investigateCallback = (id: number, reason: string) => apiPost<void>(`/admin/betting/provider-callback-logs/${id}/investigate`, { reason })
