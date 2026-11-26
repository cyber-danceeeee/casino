import { apiDelete, apiGet, apiPatch, apiPost, apiPut } from './http'
import type { PageResult } from './adminUsers'

export type RiskRecord = Record<string, unknown> & { id?: number; userID?: number }
export type RiskDetail = Record<string, unknown>
export type RiskQuery = Record<string, string | number | undefined>
export type RiskPayload = Record<string, unknown>

export const getRiskDashboard = (query: RiskQuery) => apiGet<RiskDetail>('/admin/risk/dashboard', query)
export const listRisk = (resource: string, query: RiskQuery) => apiGet<PageResult<RiskRecord>>(`/admin/risk/${resource}`, query)
export const getRisk = (resource: string, id: number) => apiGet<RiskDetail>(`/admin/risk/${resource}/${id}`)
export const createRisk = (resource: string, payload: RiskPayload) => apiPost<{ id: number }>(`/admin/risk/${resource}`, payload)
export const updateRisk = (resource: string, id: number, payload: RiskPayload) => apiPut<void>(`/admin/risk/${resource}/${id}`, payload)
export const deleteRisk = (resource: string, id: number, reason: string) => apiDelete<void>(`/admin/risk/${resource}/${id}`, { reason })
export const riskPost = (resource: string, id: number, action: string, payload: RiskPayload = {}) => apiPost<void>(`/admin/risk/${resource}/${id}/${action}`, payload)
export const riskPatch = (resource: string, id: number, action: string, payload: RiskPayload) => apiPatch<void>(`/admin/risk/${resource}/${id}/${action}`, payload)
