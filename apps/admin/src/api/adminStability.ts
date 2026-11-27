import { requestID } from '@/utils/requestID'
import { apiDelete, apiGet, apiPatch, apiPost, apiPut, http } from '@/api/http'

export type OpsQuery = Record<string, string | number | boolean | undefined>
export type OpsRecord = Record<string, unknown> & { id?: number; status?: string; siteID?: number }
export type OpsPage<T = OpsRecord> = { list: T[]; total: number; page: number; pageSize: number }

const admin = '/admin'
const system = '/admin/system'

export const operationsAPI = {
  support: {
    players: (params: OpsQuery) => apiGet<OpsPage>(`${admin}/support/players`, params),
    player: (id: number) => apiGet<Record<string, unknown>>(`${admin}/support/players/${id}`),
	playerByNumber: (playerNo: string) => apiGet<Record<string, unknown>>(`${admin}/support/players/by-number/${encodeURIComponent(playerNo)}`),
    timeline: (id: number, params: OpsQuery) => apiGet<OpsPage>(`${admin}/support/players/${id}/timeline`, params),
	timelineByNumber: (playerNo: string, params: OpsQuery) => apiGet<OpsPage>(`${admin}/support/players/by-number/${encodeURIComponent(playerNo)}/timeline`, params),
    orders: (params: OpsQuery) => apiGet<OpsPage>(`${admin}/support/orders`, params),
    orderDetail: (kind: string, id: number) => apiGet<Record<string, unknown>>(`${admin}/support/orders/${encodeURIComponent(kind)}/${id}`),
    cases: (params: OpsQuery) => apiGet<OpsPage>(`${admin}/support/cases`, params),
    caseDetail: (id: number) => apiGet<Record<string, unknown>>(`${admin}/support/cases/${id}`),
    createCase: (payload: unknown) => apiPost<{ id: number }>(`${admin}/support/cases`, payload),
    updateCase: (id: number, payload: unknown) => apiPut<void>(`${admin}/support/cases/${id}`, payload),
    addCaseNote: (id: number, payload: unknown) => apiPost<{ id: number }>(`${admin}/support/cases/${id}/notes`, payload),
    exports: (params: OpsQuery) => apiGet<OpsPage>(`${admin}/support/exports`, params),
    createExport: (payload: unknown) => apiPost<OpsRecord>(`${admin}/support/exports`, payload),
    downloadExport: async (id: number) => {
      const response = await http.get<Blob>(`${admin}/support/exports/${id}/download`, { responseType: 'blob' })
      const disposition = String(response.headers['content-disposition'] || '')
      const filename = disposition.match(/filename="?([^";]+)"?/)?.[1] || `support-export-${id}.csv`
      const url = URL.createObjectURL(response.data)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      link.click()
      URL.revokeObjectURL(url)
    },
  },
  monitor: {
    payments: (params: OpsQuery) => apiGet<Record<string, unknown>>(`${admin}/monitor/payments`, params),
    paymentDetail: (kind: string, id: number) => apiGet<Record<string, unknown>>(`${admin}/monitor/payments/orders/${kind}/${id}`),
    paymentDetailByNumber: (kind: string, orderNo: string) => apiGet<Record<string, unknown>>(`${admin}/monitor/payments/orders/${encodeURIComponent(kind)}/by-number/${encodeURIComponent(orderNo)}`),
    markReview: (kind: string, id: number, reason: string) => apiPost<void>(`${admin}/monitor/payments/orders/${kind}/${id}/review`, { reason }),
    providers: (params: OpsQuery) => apiGet<Record<string, unknown>>(`${admin}/monitor/providers`, params),
    providerDetail: (code: string, params: OpsQuery) => apiGet<Record<string, unknown>>(`${admin}/monitor/providers/${encodeURIComponent(code)}`, params),
    note: (payload: unknown) => apiPost<{ id: number }>(`${admin}/operations/notes`, payload),
  },
  maintenance: {
    site: (params: OpsQuery) => apiGet<OpsPage>(`${admin}/maintenance/site`, params),
    saveSite: (payload: unknown) => apiPut<{ id: number }>(`${admin}/maintenance/site`, payload),
    providers: (params: OpsQuery) => apiGet<OpsPage>(`${admin}/maintenance/providers`, params),
    saveProvider: (payload: unknown) => apiPut<{ id: number }>(`${admin}/maintenance/providers`, payload),
    games: (params: OpsQuery) => apiGet<OpsPage>(`${admin}/maintenance/games`, params),
    saveGame: (payload: unknown) => apiPut<{ id: number }>(`${admin}/maintenance/games`, payload),
  },

  health: () => apiGet<Record<string, unknown>>(`${system}/health`),
  jobs: {
    overview: (params: OpsQuery) => apiGet<Record<string, unknown>>(`${system}/jobs`, params),
    taskStatus: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${system}/jobs/tasks/${id}/status`, { enabled, reason }),
    runTask: (id: number, reason: string) => apiPost<void>(`${system}/jobs/tasks/${id}/run`, { reason }),
    retryExecution: (id: number, reason: string) => apiPost<void>(`${system}/jobs/executions/${id}/retry`, { reason }),
    retryDeadLetter: (id: number, reason: string) => apiPost<Record<string, unknown>>(`${system}/jobs/dead-letters/${id}/retry`, { reason }),
    retryRabbitDeadLetter: (eventID: string, queue: string, reason: string) => apiPost<void>(`${system}/jobs/rabbit/dead-letters/${encodeURIComponent(eventID)}/retry`, { queue, reason }),
    purgeRabbitRetry: (queue: string, reason: string) => apiPost<{ purged: number }>(`${system}/jobs/rabbit/retry/${encodeURIComponent(queue)}/purge`, { reason }),
  },
  alerts: {
    list: (params: OpsQuery) => apiGet<OpsPage>(`${system}/alerts`, params),
    acknowledge: (id: number, reason: string, note: string) => apiPost<void>(`${system}/alerts/${id}/acknowledge`, { reason, note }),
    resolve: (id: number, reason: string, note: string) => apiPost<void>(`${system}/alerts/${id}/resolve`, { reason, note }),
  },
  backups: {
    list: (params: OpsQuery) => apiGet<OpsPage>(`${system}/backups`, params),
    detail: (id: number) => apiGet<OpsRecord>(`${system}/backups/${id}`),
    create: (payload: unknown) => apiPost<OpsRecord>(`${system}/backups`, payload),
    validate: (id: number) => apiPost<Record<string, unknown>>(`${system}/backups/${id}/validate`),
    restore: (id: number, payload: unknown) => apiPost<Record<string, unknown>>(`${system}/backups/${id}/restore`, payload),
    remove: (id: number, reason: string) => apiDelete<void>(`${system}/backups/${id}`, { reason }),
  },
}

export function operationIdempotencyKey(prefix: string) {
  return `${prefix}:${Date.now()}:${requestID()}`
}
