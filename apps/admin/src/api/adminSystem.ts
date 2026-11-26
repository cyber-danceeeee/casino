import { apiDelete, apiGet, apiPatch, apiPost, apiPut, http, type ApiResponse } from '@/api/http'

export type SystemQuery = Record<string, string | number | undefined>
export type SystemRecord = Record<string, unknown> & { id?: number; enabled?: boolean; status?: boolean | string | number }
export type SystemPage = { list: SystemRecord[]; total: number; page: number; pageSize: number }
export type SystemDetail = Record<string, unknown>
export type AdminNavigationItem = { id: number; parentID?: number | null; code: string; routePath: string; icon: string }
export type PlatformBranding = { platformName: string; platformLogoUrl: string; platformContact: string }
export type UploadedAsset = { url: string; key: string; name: string; size: number; mimeType: string }
export type BrandOverview = SystemDetail & {
  siteID: number
  brandName: string
  countryCode: string
  currencyCode: string
  defaultLanguage: string
  timezone: string
  currentSite: string
  siteCode: string
  primaryDomain: string
  adminDomain: string
  h5Domain: string
  maintenanceEnabled: boolean
  gamesEnabled: number
  gamesConfigured: number
  providersEnabled: number
  packageCode: string
  gameLimit: number
  providerLimit: number
  publishStatus: string
  publishedVersion: number
  draftVersion: number
  publishedAt: string
  draftUpdatedAt: string
  hasUnpublished: boolean
	unpublishedChanges: number
	unpublishedModules: string[]
	pendingChanges: Array<{ module: string; summary: string; updatedBy: string; updatedAt: string; before?: SystemDetail; after?: SystemDetail }>
	lastUpdatedBy: string
  isInitialized: boolean
	setupChecklist: Array<{ key: string; status: 'COMPLETED' | 'INCOMPLETE' | 'WARNING' | 'ERROR'; path: string }>
	setupCompleted: number
	setupTotal: number
	setupPercent: number
  publishedConfig?: SystemDetail
  draftConfig?: SystemDetail
  configMeta?: SystemDetail
}

export type ContentVisibility = 'ALL' | 'GUEST' | 'LOGGED_IN'
export type AnnouncementType = 'SYSTEM' | 'MAINTENANCE' | 'PROMOTION' | 'PAYMENT' | 'GAME' | 'GENERAL'
export type AnnouncementConfiguration = {
  title: string; summary: string; content: string; announcementType: AnnouncementType
  priority: number; sortOrder: number; isPinned: boolean; languageCode: string
  countryCode: string; currencyCode: string; visibilityScope: ContentVisibility
  startAt: string | null; endAt: string | null; enabled: boolean; deleted?: boolean
}
export type SiteAnnouncementRecord = {
  id: number; siteID: number; publishStatus: 'DRAFT' | 'PUBLISHED'
  configVersion: number; draftUpdatedAt: string; publishedAt: string
  draftConfiguration: AnnouncementConfiguration; publishedConfiguration: AnnouncementConfiguration
}
export type PopupFrequency = 'EVERY_VISIT' | 'ONCE_PER_SESSION' | 'ONCE_PER_DAY' | 'ONCE_PER_USER'
export type PopupTrigger = 'HOME_ENTER' | 'PAGE_ENTER'
export type PopupActionType = 'NONE' | 'CLOSE_ONLY' | 'OPEN_INTERNAL_PAGE' | 'OPEN_EXTERNAL_URL' | 'OPEN_GAME' | 'OPEN_PROMOTION' | 'OPEN_DEPOSIT' | 'OPEN_WITHDRAW' | 'OPEN_SUPPORT' | 'OPEN_COUPON' | 'OPEN_VIP' | 'OPEN_DOWNLOAD'
export type PopupComponentRecord = {
  id: number; componentCode: string; nameZh: string; nameEn: string; descriptionZh: string; descriptionEn: string
  category: string; previewURL: string; previewVideoURL: string; h5MinVersion: string
  defaultActionType: PopupActionType; allowedActionTypes: PopupActionType[]; status: number; enabled: boolean
  sortOrder: number; createdAt: string; updatedAt: string
}
export type PopupComponentInput = Omit<PopupComponentRecord, 'id' | 'status' | 'createdAt' | 'updatedAt'> & { reason?: string }
export type PopupConfiguration = {
  popupComponentID: number; campaignName: string; actionType: PopupActionType; actionTarget: string
  triggerType: PopupTrigger; triggerValue: string; frequencyType: PopupFrequency
  displayLimit: number; cooldownMinutes: number; visibilityScope: ContentVisibility
  startAt: string | null; endAt: string | null; priority: number; enabled: boolean; deleted?: boolean
}
export type SitePopupRecord = {
  id: number; siteID: number; publishStatus: 'DRAFT' | 'PUBLISHED'
  configVersion: number; draftUpdatedAt: string; publishedAt: string
  draftConfiguration: PopupConfiguration; publishedConfiguration: PopupConfiguration
  component?: PopupComponentRecord
}

const root = '/admin/system'

export const systemAPI = {
  navigation: () => apiGet<{ list: AdminNavigationItem[] }>(`${root}/navigation`),
  admins: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/admins`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/admins/${id}`),
    create: (payload: unknown) => apiPost<{ id: number }>(`${root}/admins`, payload),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/admins/${id}`, payload),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/admins/${id}/status`, { enabled, reason }),
    resetPassword: (id: number, reason: string) => apiPost<{ temporaryPassword: string }>(`${root}/admins/${id}/reset-password`, { reason }),
    resetMFA: (id: number, reason: string) => apiPost<void>(`${root}/admins/${id}/reset-mfa`, { reason }),
    forceLogout: (id: number, reason: string) => apiPost<void>(`${root}/admins/${id}/force-logout`, { reason }),
    roles: (id: number, ids: number[], reason: string) => apiPut<void>(`${root}/admins/${id}/roles`, { ids, reason }),
    department: (id: number, departmentID: number | null, reason: string) => apiPut<void>(`${root}/admins/${id}/department`, { departmentID, reason }),
  },
  roles: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/roles`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/roles/${id}`),
    create: (payload: unknown) => apiPost<{ id: number }>(`${root}/roles`, payload),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/roles/${id}`, payload),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/roles/${id}/status`, { enabled, reason }),
    permissions: (id: number, ids: number[], reason: string) => apiPut<void>(`${root}/roles/${id}/permissions`, { ids, reason }),
    menus: (id: number, ids: number[], reason: string) => apiPut<void>(`${root}/roles/${id}/menus`, { ids, reason }),
    copy: (id: number, payload: unknown) => apiPost<{ id: number }>(`${root}/roles/${id}/copy`, payload),
    remove: (id: number, reason: string) => apiDelete<void>(`${root}/roles/${id}`, { reason }),
  },
  permissions: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/permissions`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/permissions/${id}`),
    create: (payload: unknown) => apiPost<{ id: number }>(`${root}/permissions`, payload),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/permissions/${id}`, payload),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/permissions/${id}/status`, { enabled, reason }),
    remove: (id: number, reason: string) => apiDelete<void>(`${root}/permissions/${id}`, { reason }),
  },
  departments: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/departments`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/departments/${id}`),
    create: (payload: unknown) => apiPost<{ id: number }>(`${root}/departments`, payload),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/departments/${id}`, payload),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/departments/${id}/status`, { enabled, reason }),
    remove: (id: number, reason: string) => apiDelete<void>(`${root}/departments/${id}`, { reason }),
  },
  menus: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/menus`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/menus/${id}`),
    create: (payload: unknown) => apiPost<{ id: number }>(`${root}/menus`, payload),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/menus/${id}`, payload),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/menus/${id}/status`, { enabled, reason }),
    copy: (id: number, payload: unknown) => apiPost<{ id: number }>(`${root}/menus/${id}/copy`, payload),
    remove: (id: number, reason: string) => apiDelete<void>(`${root}/menus/${id}`, { reason }),
  },
  configs: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/configs`, params),
    update: (key: string, value: string, reason: string) => apiPut<void>(`${root}/configs/${encodeURIComponent(key)}`, { value, reason }),
  },
  platformBranding: () => apiGet<PlatformBranding>('/admin/platform-branding'),
  sites: {
    detail: (id: number) => apiGet<SystemDetail>(`${root}/sites/${id}`),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/sites/${id}`, payload),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/sites/${id}/status`, { enabled, reason }),
    preview: (id: number) => apiGet<SystemDetail>(`${root}/sites/${id}/config/preview`),
    publish: (id: number, reason: string) => apiPost<SystemDetail>(`${root}/sites/${id}/config/publish`, { reason }),
    uploadAPK: async (file: File) => uploadForm<UploadedAsset>(`${root}/sites/apk/upload`, file),
  },
  announcements: {
    list: (siteID: number, params: SystemQuery) => apiGet<{ list: SiteAnnouncementRecord[]; total: number; page: number; pageSize: number }>(`${root}/sites/${siteID}/announcements`, params),
    detail: (siteID: number, id: number) => apiGet<{ overview: SiteAnnouncementRecord }>(`${root}/sites/${siteID}/announcements/${id}`),
    create: (siteID: number, payload: AnnouncementConfiguration & { reason?: string }) => apiPost<{ id: number }>(`${root}/sites/${siteID}/announcements`, payload),
    update: (siteID: number, id: number, payload: AnnouncementConfiguration & { reason?: string }) => apiPut<{ id: number }>(`${root}/sites/${siteID}/announcements/${id}`, payload),
    copy: (siteID: number, id: number) => apiPost<{ id: number }>(`${root}/sites/${siteID}/announcements/${id}/copy`),
    remove: (siteID: number, id: number) => apiDelete<void>(`${root}/sites/${siteID}/announcements/${id}`),
    reorder: (siteID: number, ids: number[], reason?: string) => apiPost<void>(`${root}/sites/${siteID}/announcements/reorder`, { ids, reason }),
  },
  popupComponents: {
    list: (params: SystemQuery) => apiGet<{ list: PopupComponentRecord[]; total: number; page: number; pageSize: number }>(`${root}/popup-components`, params),
    available: () => apiGet<{ list: PopupComponentRecord[] }>(`${root}/popup-components/available`),
    create: (payload: PopupComponentInput) => apiPost<{ id: number }>(`${root}/popup-components`, payload),
    update: (id: number, payload: PopupComponentInput) => apiPut<{ id: number }>(`${root}/popup-components/${id}`, payload),
    remove: (id: number, reason: string) => apiDelete<void>(`${root}/popup-components/${id}`, { reason }),
  },
  popups: {
    list: (siteID: number, params: SystemQuery) => apiGet<{ list: SitePopupRecord[]; total: number; page: number; pageSize: number }>(`${root}/sites/${siteID}/popups`, params),
    detail: (siteID: number, id: number) => apiGet<{ overview: SitePopupRecord }>(`${root}/sites/${siteID}/popups/${id}`),
    create: (siteID: number, payload: PopupConfiguration & { reason?: string }) => apiPost<{ id: number }>(`${root}/sites/${siteID}/popups`, payload),
    update: (siteID: number, id: number, payload: PopupConfiguration & { reason?: string }) => apiPut<{ id: number }>(`${root}/sites/${siteID}/popups/${id}`, payload),
    copy: (siteID: number, id: number) => apiPost<{ id: number }>(`${root}/sites/${siteID}/popups/${id}/copy`),
    remove: (siteID: number, id: number) => apiDelete<void>(`${root}/sites/${siteID}/popups/${id}`),
    reorder: (siteID: number, ids: number[], reason?: string) => apiPost<void>(`${root}/sites/${siteID}/popups/reorder`, { ids, reason }),
  },
  countries: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/countries`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/countries/${id}`),
    create: (payload: unknown) => apiPost<{ id: number }>(`${root}/countries`, payload),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/countries/${id}`, payload),
    remove: (id: number, reason: string) => apiDelete<void>(`${root}/countries/${id}`, { reason }),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/countries/${id}/status`, { enabled, reason }),
    setDefault: (id: number, reason: string) => apiPost<void>(`${root}/countries/${id}/default`, { reason }),
    uploadAsset: async (file: File, kind: 'image' | 'lottie') => {
      const form = new FormData()
      form.append('file', file)
      form.append('kind', kind)
      const { data } = await http.post<ApiResponse<{ url: string; name: string; size: number; mimeType: string }>>(
        `${root}/countries/assets/upload`,
        form,
        { headers: { 'Content-Type': 'multipart/form-data' } },
      )
      if (data.code !== 0) throw new Error(data.msg || 'Upload failed')
      return data.data
    },
  },
  auditLogs: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/audit-logs`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/audit-logs/${id}`),
  },
  loginLogs: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/login-logs`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/login-logs/${id}`),
  },
  tasks: {
    list: (params: SystemQuery) => apiGet<SystemPage>(`${root}/tasks`, params),
    detail: (id: number) => apiGet<SystemDetail>(`${root}/tasks/${id}`),
    create: (payload: unknown) => apiPost<{ id: number }>(`${root}/tasks`, payload),
    update: (id: number, payload: unknown) => apiPut<void>(`${root}/tasks/${id}`, payload),
    status: (id: number, enabled: boolean, reason: string) => apiPatch<void>(`${root}/tasks/${id}/status`, { enabled, reason }),
    run: (id: number, reason: string) => apiPost<void>(`${root}/tasks/${id}/run`, { reason }),
    remove: (id: number, reason: string) => apiDelete<void>(`${root}/tasks/${id}`, { reason }),
    executions: (id: number, params: SystemQuery) => apiGet<SystemPage>(`${root}/tasks/${id}/executions`, params),
  },
  monitoring: {
    get: () => apiGet<SystemDetail>(`${root}/monitoring`),
    service: (name: string) => apiGet<SystemDetail>(`${root}/monitoring/services/${encodeURIComponent(name)}`),
  },
  brand: {
    markets: () => apiGet<SystemPage>(`${root}/h5/markets`),
    save: (payload: Record<string, unknown>) => apiPut<void>(`${root}/h5/settings`, payload),
    overview: () => apiGet<BrandOverview>(`${root}/brand/overview`),
  },
}

export async function uploadPlatformAsset(file: File) {
  return uploadForm<UploadedAsset>(`${root}/configs/assets/upload`, file)
}

async function uploadForm<T>(url: string, file: File): Promise<T> {
  const form = new FormData()
  form.append('file', file)
  const { data } = await http.post<ApiResponse<T>>(url, form, { headers: { 'Content-Type': 'multipart/form-data' } })
  if (data.code !== 0) throw new Error(data.msg || 'Upload failed')
  return data.data
}
