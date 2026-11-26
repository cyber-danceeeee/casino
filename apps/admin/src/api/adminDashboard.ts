import type { AxiosRequestConfig } from 'axios'
import { http, type ApiResponse } from './http'

export type DashboardFilter = {
  startTime: string
  endTime: string
  timezone: string
  granularity: 'hour' | 'day' | 'week' | 'month'
  country?: string
  currency: string
  providerCode?: string
  categoryId?: number
  channelId?: number
  agentId?: number
  metric?: string
  limit?: 10 | 20
}

export type DashboardOption = { label: string; value: string | number }
export type DashboardOptions = {
  countries: DashboardOption[]
  currencies: DashboardOption[]
  providers: DashboardOption[]
  categories: DashboardOption[]
  channels: DashboardOption[]
  agents: DashboardOption[]
}
export type ChartRow = Record<string, string | number | null>
export type ChartResponse = {
  dimensions: string[]
  metrics: string[]
  rows: ChartRow[]
  currency?: string
  moneyFormat?: string
}
export type MetricCardData = { key: string; value: string | number | null; previousValue?: string | number | null; changePercent?: string | null; format?: 'money' | 'count' | 'percent'; currency?: string }
export type OverviewResponse = { cards: MetricCardData[]; currency: string; moneyFormat: string; metricBasis: Record<string, string> }
export type RiskResponse = { cards: MetricCardData[]; byType: ChartRow[]; byLevel: ChartRow[]; trend: ChartRow[]; reviewResults: ChartRow[] }
export type RealtimeResponse = { rows: ChartRow[]; refreshedAt: string }
export type AlertResponse = { rows: ChartRow[] }

async function get<T>(path: string, params?: Record<string, unknown>, signal?: AbortSignal) {
  const config: AxiosRequestConfig = { params, signal }
  const { data } = await http.get<ApiResponse<T>>(`/admin/dashboard${path}`, config)
  if (data.code !== 0) throw new Error(data.msg)
  return data.data
}

export const dashboardApi = {
  filters: (signal?: AbortSignal) => get<DashboardOptions>('/filters', undefined, signal),
  overview: (params: DashboardFilter, signal?: AbortSignal) => get<OverviewResponse>('/overview', params, signal),
  chart: (path: string, params: DashboardFilter, signal?: AbortSignal) => get<ChartResponse>(path, params, signal),
  risk: (params: DashboardFilter, signal?: AbortSignal) => get<RiskResponse>('/risk-overview', params, signal),
  realtime: (params: DashboardFilter, signal?: AbortSignal) => get<RealtimeResponse>('/realtime', params, signal),
  alerts: (params: DashboardFilter, signal?: AbortSignal) => get<AlertResponse>('/alerts', params, signal),
}
