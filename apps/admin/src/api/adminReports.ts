import { apiGet, apiPost, http } from '@/api/http'

export type ReportQuery = Record<string, string | number | undefined>

export type ReportChartSeries = {
  name: string
  currency?: string
  data: string[]
}

export type ReportChart = {
  key: string
  categories: string[]
  series: ReportChartSeries[]
}

export type ReportResult = {
  summaryByCurrency: Array<Record<string, unknown>>
  list: Array<Record<string, unknown>>
  total: number
  page: number
  pageSize: number
  charts: ReportChart[]
  metricBasis: Record<string, string>
}

export type ReportDetail = Record<string, unknown>

export type ReportExportTask = Record<string, unknown> & {
  id: number
  taskID: string
  reportType: string
  fileFormat: 'CSV' | 'XLSX'
  status: string
}

export type CreateReportExportPayload = {
  reportType: string
  filters: Record<string, string>
  fileFormat: 'CSV' | 'XLSX'
  idempotencyKey: string
}

export const fetchFinancialReport = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/reports/financial', params)
export const fetchFinancialReportDetail = (params: ReportQuery) =>
  apiGet<ReportDetail>('/admin/reports/financial/detail', params)
export const fetchGameReport = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/reports/games', params)
export const fetchGameReportDetail = (id: number, params: ReportQuery) =>
  apiGet<ReportDetail>(`/admin/reports/games/${id}`, params)
export const fetchProviderReport = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/reports/providers', params)
export const fetchProviderReportDetail = (code: string, params: ReportQuery) =>
  apiGet<ReportDetail>(`/admin/reports/providers/${encodeURIComponent(code)}`, params)
export const fetchUserReport = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/reports/users', params)
export const fetchUserReportDetail = (id: number, params: ReportQuery) =>
  apiGet<ReportDetail>(`/admin/reports/users/${id}`, params)
export const fetchChannelReport = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/reports/channels', params)
export const fetchChannelReportDetail = (id: number, params: ReportQuery) =>
  apiGet<ReportDetail>(`/admin/reports/channels/${id}`, params)
export const fetchAgentReport = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/reports/agents', params)
export const fetchAgentReportDetail = (id: number, params: ReportQuery) =>
  apiGet<ReportDetail>(`/admin/reports/agents/${id}`, params)
export const fetchPromotionReport = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/reports/promotions', params)
export const fetchPromotionReportDetail = (id: number, params: ReportQuery) =>
  apiGet<ReportDetail>(`/admin/reports/promotions/${id}`, params)

export const createReportExport = (payload: CreateReportExportPayload) =>
  apiPost<{ id: number }>('/admin/report-exports', payload)
export const fetchReportExports = (params: ReportQuery) =>
  apiGet<ReportResult>('/admin/report-exports', params)
export const fetchReportExport = (id: number) =>
  apiGet<ReportDetail>(`/admin/report-exports/${id}`)
export const retryReportExport = (id: number, reason: string) =>
  apiPost<void>(`/admin/report-exports/${id}/retry`, { reason })
export const cancelReportExport = (id: number, reason: string) =>
  apiPost<void>(`/admin/report-exports/${id}/cancel`, { reason })

export async function downloadReportExport(id: number, format: string) {
  const response = await http.get(`/admin/report-exports/${id}/download`, {
    responseType: 'blob',
  })
  const url = URL.createObjectURL(response.data)
  const link = document.createElement('a')
  link.href = url
  link.download = `report-${id}.${format.toLowerCase()}`
  link.click()
  URL.revokeObjectURL(url)
}
