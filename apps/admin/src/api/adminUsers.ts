import { apiGet, apiPatch, apiPost } from './http'

export type PageResult<T> = { list: T[]; total: number; page: number; pageSize: number }
export type AdminUserQuery = Record<string, string | number | undefined>
export type UserStatus = 'ACTIVE' | 'DISABLED' | 'FROZEN' | 'LOCKED' | 'CLOSED'
export type RiskLevel = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
export type RiskStatus = 'OBSERVED' | 'UNDER_REVIEW' | 'RESTRICTED' | 'CLEARED' | 'BLOCKED'

export type AdminActor = {
  id: number
  sessionID?: number
  username: string
  displayName?: string
  tenantID?: number
  siteID?: number
  defaultSiteID?: number
  platform?: boolean
  adminType: 'PLATFORM' | 'TENANT'
  mustChangePassword?: boolean
  roleCodes?: string[]
  permissions: string[]
}

export type AdminUser = {
  id: number
  userNo: string
  username: string
  email: string
  phoneCountryCode: string
  phone: string
  nickname: string
  countryCode: string
  language: string
  currency: string
  vipLevel: number
  kycStatus: string
  status: number
  statusText: UserStatus
  riskLevel: RiskLevel
  riskStatus: RiskStatus
  internalTags: string
  loginEnabled: number
  gameAccessEnabled: number
  registrationChannel: string
  registeredIP: string
  lastLoginIP: string
  registeredAt: string
  lastLoginAt: string
  updatedAt: string
  cashBalance: string
  frozenBalance: string
  bonusBalance: string
  withdrawableBalance: string
}

export type UserDetail = AdminUser & {
  referralCode: string
  registrationDevice: string
  lastLoginDevice: string
  failedLoginCount: number
  loginEnabled: number
  gameAccessEnabled: number
  depositEnabled: number
  withdrawalEnabled: number
  riskReviewRequired: number
  availableBalance: string
  totalDeposit: string
  totalWithdrawal: string
  totalBet: string
  totalWin: string
  lastActiveAt: string
  lastPasswordChangedAt: string
}

export type LoginRecord = {
  id: number
  userID: number
  userNo: string
  username: string
  phoneCountryCode: string
  phone: string
  nickname: string
  success: number
  loginResult: 'SUCCESS' | 'FAILED'
  failureReason: string
  deviceID: string
  clientIP: string
  country: string
  region: string
  deviceType: string
  platform: string
  os: string
  browser: string
  appVersion: string
  userAgent: string
  loginAt: string
}

export type UserDevice = {
  id: number
  userID: number
  userNo: string
  phoneCountryCode: string
  phone: string
  nickname: string
  deviceID: string
  platform: string
  deviceModel: string
  os: string
  browser: string
  appVersion: string
  fingerprintHash: string
  firstIP: string
  lastIP: string
  clientIP: string
  userAgent: string
  status: number
  statusText: 'active' | 'blocked'
  trustStatus: number
  trustStatusText: 'NORMAL' | 'TRUSTED'
  riskStatus: number
  riskStatusText: 'NORMAL' | 'OBSERVED' | 'RISKY'
  loginCount: number
  associatedAccounts: number
  firstSeenAt: string
  lastSeenAt: string
}

export type PushDevice = {
  id: number
  userID: number
  userNo: string
  phoneCountryCode: string
  phone: string
  nickname: string
  platform: string
  deviceID: string
  fingerprintHash: string
  manufacturer: string
  brand: string
  deviceModel: string
  deviceName: string
  osVersion: string
  sdkVersion: number
  securityPatch: string
  networkType: string
  networkMetered: boolean
  networkRoaming: boolean
  carrierName: string
  screenWidth: number
  screenHeight: number
  screenDensity: number
  timezone: string
  locale: string
  batteryLevel: number
  charging: boolean
  notificationPermission: string
  appVersion: string
  appVersionCode: number
  language: string
  lastIP: string
  userAgent: string
  enabled: boolean
  status: 'ENABLED' | 'INVALIDATED'
  lastSeenAt: string
  invalidatedAt: string
  createdAt: string
  updatedAt: string
  lastError: string
}

export type RiskUser = {
  id: number
  userID: number
  userNo: string
  username: string
  phone: string
  countryCode: string
  riskLevel: RiskLevel
  riskStatus: RiskStatus
  riskTags: string
  triggerCount: number
  registeredIP: string
  lastLoginIP: string
  associatedAccounts: number
  assignedReviewer: string
  latestReason: string
  latestRiskAt: string
  updatedAt: string
}
export type RiskEvent = { id: number; eventType: string; riskLevel: RiskLevel; source: string; description: string; operatorUsername: string; createdAt: string }
export type RiskUserDetail = { profile: RiskUser; events: RiskEvent[]; devices: UserDevice[] }

export type UserNote = { id: number; operatorID: number; operatorUsername: string; content: string; createdAt: string }
export type OperationLog = { id: number; operatorUsername: string; action: string; beforeState: string; afterState: string; reason: string; createdAt: string }

export function fetchAdminActor() { return apiGet<AdminActor>('/admin/auth/me') }
export function fetchAdminUsers(params: AdminUserQuery) { return apiGet<PageResult<AdminUser>>('/admin/users', params) }
export function fetchAdminUser(id: number) { return apiGet<UserDetail>(`/admin/users/${id}`) }
export function fetchAdminUserNotes(id: number) { return apiGet<UserNote[]>(`/admin/users/${id}/notes`) }
export function addAdminUserNote(id: number, content: string) { return apiPost<void>(`/admin/users/${id}/notes`, { content }) }
export function updateAdminUserTags(id: number, tags: string[], reason: string) { return apiPatch<{ tags: string[] }>(`/admin/users/${id}/tags`, { tags, reason }) }
export function fetchAdminUserOperationLogs(id: number, params: AdminUserQuery) { return apiGet<PageResult<OperationLog>>(`/admin/users/${id}/operation-logs`, params) }
export function fetchAdminLoginRecords(params: AdminUserQuery) { return apiGet<PageResult<LoginRecord>>('/admin/users/login-records', params) }
export function fetchAdminDevices(params: AdminUserQuery) { return apiGet<PageResult<UserDevice>>('/admin/users/devices', params) }
export function fetchAdminPushDevices(params: AdminUserQuery) { return apiGet<PageResult<PushDevice>>('/admin/users/push-devices', params) }
export function fetchAdminDevice(id: number) { return apiGet<UserDevice>(`/admin/users/devices/${id}`) }
export function fetchAdminDeviceUsers(id: number) { return apiGet<AdminUser[]>(`/admin/users/devices/${id}/users`) }
export function fetchAdminRiskUsers(params: AdminUserQuery) { return apiGet<PageResult<RiskUser>>('/admin/users/risk', params) }
export function fetchAdminRiskUser(userID: number) { return apiGet<RiskUserDetail>(`/admin/users/risk/${userID}`) }

export function updateAdminUserStatus(id: number, status: UserStatus, reason: string) {
  return apiPatch<{ status: UserStatus }>(`/admin/users/${id}/status`, { status, reason })
}
export function updateAdminUserRestrictions(id: number, payload: Record<string, boolean | string>) {
  return apiPatch<Record<string, boolean>>(`/admin/users/${id}/restrictions`, payload)
}
export function updateAdminUserRisk(id: number, payload: { riskLevel: RiskLevel; riskStatus: RiskStatus; assignedReviewer: string; reason: string }) {
  return apiPatch<Record<string, string>>(`/admin/users/${id}/risk-status`, payload)
}
export function updateAdminDevice(id: number, payload: { status?: 'ACTIVE' | 'BLOCKED'; trustStatus?: 'NORMAL' | 'TRUSTED'; reason: string }) {
  return apiPatch<Record<string, string>>(`/admin/users/devices/${id}/status`, payload)
}
