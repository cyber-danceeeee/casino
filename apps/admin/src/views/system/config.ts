import { systemAPI, type SystemPage, type SystemQuery, type SystemRecord } from '@/api/adminSystem'

export type Filter = { key: string; type?: 'text' | 'select' | 'dateRange'; options?: string[] }
export type Column = { key: string; width?: number; kind?: 'time' | 'boolean' | 'status' | 'json' }
export type FormField = { key: string; type?: 'text' | 'select' | 'multiSelect' | 'number' | 'boolean' | 'password' | 'json' | 'asset' | 'apk' | 'color'; options?: string[]; createOnly?: boolean; assetKind?: 'image' | 'lottie' }
export type SystemConfig = {
  permission: string
  createPermission?: string
  updatePermission?: string
  statusPermission?: string
  deletePermission?: string
  filters: Filter[]
  columns: Column[]
  tabs: string[]
  form: FormField[]
  loader: (query: SystemQuery) => Promise<SystemPage>
  detail?: (id: number) => Promise<Record<string, unknown>>
  create?: (payload: Record<string, unknown>) => Promise<unknown>
  update?: (id: number, payload: Record<string, unknown>) => Promise<unknown>
  remove?: (id: number, reason: string) => Promise<unknown>
  status?: (id: number, enabled: boolean, reason: string) => Promise<unknown>
  readOnly?: boolean
}

const f = (key: string, type: Filter['type'] = 'text', options?: string[]): Filter => ({ key, type, options })
const c = (key: string, width = 150, kind?: Column['kind']): Column => ({ key, width, kind })
const dates = (key: string) => f(`${key}Range`, 'dateRange')
const detail = (value: unknown) => value as Promise<Record<string, unknown>>
const timezoneOptions = [
  'UTC',
  'Asia/Karachi',
  'Asia/Shanghai',
  'Asia/Kuala_Lumpur',
  'Asia/Tokyo',
  'Asia/Bangkok',
  'Asia/Jakarta',
  'Asia/Manila',
  'Asia/Ho_Chi_Minh',
  'Asia/Kolkata',
  'Asia/Dhaka',
  'Asia/Singapore',
  'Asia/Dubai',
  'Africa/Lagos',
  'Africa/Johannesburg',
  'America/New_York',
  'America/Los_Angeles',
  'America/Mexico_City',
  'America/Toronto',
  'America/Sao_Paulo',
  'America/Argentina/Buenos_Aires',
  'America/Bogota',
  'America/Lima',
  'Europe/London',
  'Europe/Istanbul',
  'Australia/Sydney',
]
const languageOptions = [
  'en', 'en-US', 'zh-CN', 'ur', 'es', 'es-MX', 'es-AR', 'es-CO', 'es-PE',
  'pt-BR', 'ja', 'th', 'id', 'fil', 'vi', 'hi', 'bn', 'ms', 'ta', 'fr', 'tr', 'ar',
]
const countryCodeOptions = [
  'PK', 'MX', 'BR', 'MY', 'TH', 'ID', 'JP', 'PH', 'VN', 'IN', 'BD', 'SG',
  'NG', 'ZA', 'US', 'CA', 'GB', 'AU', 'TR', 'AR', 'CO', 'PE', 'AE',
]
const phoneCountryCodeOptions = [
  '+1', '+27', '+44', '+51', '+52', '+54', '+55', '+57', '+60', '+61', '+62',
  '+63', '+65', '+66', '+81', '+84', '+90', '+91', '+92', '+234', '+880', '+971',
]

export const systemConfigs: Record<string, SystemConfig> = {
  tenants: {
    permission: 'system:tenant:list', createPermission: 'system:tenant:create', updatePermission: 'system:tenant:update', statusPermission: 'system:tenant:status', deletePermission: 'system:tenant:delete',
    filters: [f('tenantCode'), f('tenantName'), f('companyName'), f('planCode', 'select', ['STANDARD', 'PREMIUM', 'ENTERPRISE']), f('defaultCurrency'), f('status', 'select', ['ENABLED', 'DISABLED']), f('lifecycleStatus', 'select', ['ACTIVE', 'SUSPENDED', 'DISABLED']), dates('created')],
    columns: [c('tenantCode', 140), c('tenantName', 180), c('companyName', 180), c('planCode', 125, 'status'), c('defaultCurrency', 120), c('defaultLanguage', 130), c('timezone', 170), c('siteCount', 105), c('adminCount', 115), c('status', 110, 'status'), c('enabled', 95, 'boolean'), c('createdAt', 175, 'time')],
    tabs: ['overview', 'sites', 'administrators', 'operationLogs'],
    form: [
      fForm('tenantCode', 'text', undefined, true),
      fForm('tenantName'),
      fForm('companyName'),
      fForm('contactName'),
      fForm('contactEmail'),
      fForm('contactPhone'),
      fForm('planCode', 'select', ['STANDARD', 'PREMIUM', 'ENTERPRISE']),
      fForm('defaultCurrency'),
      fForm('defaultLanguage', 'select', languageOptions),
      fForm('timezone', 'select', timezoneOptions),
      fForm('enabled', 'boolean', undefined, true),
      fForm('status', 'select', ['ACTIVE', 'SUSPENDED', 'DISABLED']),
      fForm('remark'),
      fForm('ownerUsername', 'text', undefined, true),
      fForm('ownerPassword', 'password', undefined, true),
      fForm('ownerEmail', 'text', undefined, true),
      fForm('reason'),
    ],
    loader: systemAPI.tenants.list, detail: (id) => detail(systemAPI.tenants.detail(id)), create: systemAPI.tenants.create, update: systemAPI.tenants.update, remove: systemAPI.tenants.remove, status: systemAPI.tenants.status,
  },
  sites: {
    permission: 'system:site:list', createPermission: 'system:site:create', updatePermission: 'system:site:update', statusPermission: 'system:site:status', deletePermission: 'system:site:delete',
    filters: [f('tenantID'), f('tenant'), f('siteCode'), f('siteName'), f('primaryDomain'), f('adminDomain'), f('h5Domain'), f('countryCode'), f('currencyCode'), f('status', 'select', ['ENABLED', 'DISABLED']), f('lifecycleStatus', 'select', ['BUILDING', 'ONLINE', 'MAINTENANCE', 'DISABLED']), dates('created')],
    columns: [c('tenant', 160), c('siteCode', 130), c('siteName', 170), c('primaryDomain', 190), c('adminDomain', 190), c('h5Domain', 190), c('countryCode', 110), c('currencyCode', 115), c('configStatus', 135, 'status'), c('configVersion', 115), c('configPublishedAt', 175, 'time'), c('configDraftUpdatedAt', 175, 'time'), c('status', 120, 'status'), c('enabled', 95, 'boolean'), c('updatedAt', 175, 'time')],
    tabs: ['overview', 'publishedConfig', 'draftConfig', 'configMeta', 'brandConfig', 'administrators', 'operationLogs'],
    form: [
      fForm('tenantID', 'number'),
      fForm('siteCode', 'text', undefined, true),
      fForm('siteName'),
      fForm('primaryDomain'),
      fForm('adminDomain'),
      fForm('h5Domain'),
      fForm('androidPackageName'),
      fForm('androidAPKObjectKey', 'apk'),
      fForm('iosBundleID'),
      fForm('countryCode'),
      fForm('currencyCode'),
      fForm('defaultLanguage', 'select', languageOptions),
      fForm('timezone', 'select', timezoneOptions),
      fForm('brandConfig', 'json'),
      fForm('enabled', 'boolean', undefined, true),
      fForm('status', 'select', ['BUILDING', 'ONLINE', 'MAINTENANCE', 'DISABLED']),
      fForm('remark'),
      fForm('reason'),
    ],
    loader: systemAPI.sites.list, detail: (id) => detail(systemAPI.sites.detail(id)), create: systemAPI.sites.create, update: systemAPI.sites.update, remove: systemAPI.sites.remove, status: systemAPI.sites.status,
  },
  admins: {
    permission: 'system:admin:list', createPermission: 'system:admin:create', updatePermission: 'system:admin:update',
    filters: [f('tenantID'), f('tenant'), f('siteID'), f('site'), f('username'), f('displayName'), f('department'), f('role'), f('status', 'select', ['ENABLED', 'DISABLED']), f('mobile'), f('email'), dates('created'), dates('lastLogin')],
    columns: [c('tenant', 150), c('site', 150), c('username'), c('displayName'), c('department'), c('roles', 180), c('mobile'), c('email', 190), c('status', 95, 'boolean'), c('lastLoginIP'), c('lastLoginTime', 175, 'time'), c('createdAt', 175, 'time')],
    tabs: ['overview', 'rolesAndPermissions', 'departments', 'loginSecurity', 'loginLogs', 'operationLogs'],
    form: [fForm('username', 'text', undefined, true), fForm('displayName'), fForm('password', 'password', undefined, true), fForm('tenantID', 'number'), fForm('siteID', 'number'), fForm('departmentID', 'number'), fForm('mobile'), fForm('email'), fForm('status', 'boolean', undefined, true)],
    loader: systemAPI.admins.list, detail: (id) => detail(systemAPI.admins.detail(id)), create: systemAPI.admins.create, update: systemAPI.admins.update, status: systemAPI.admins.status,
  },
  roles: {
    permission: 'system:role:list', createPermission: 'system:role:create', updatePermission: 'system:role:update', deletePermission: 'system:role:delete',
    filters: [f('tenantID'), f('tenant'), f('code'), f('name'), f('status', 'select', ['ENABLED', 'DISABLED']), f('dataScope', 'select', ['ALL', 'DEPARTMENT', 'DEPARTMENT_AND_CHILDREN', 'SELF', 'CUSTOM']), dates('created')],
    columns: [c('tenant', 150), c('code'), c('nameZh'), c('nameEn'), c('administratorCount', 135), c('permissionCount', 125), c('dataScope', 190, 'status'), c('enabled', 95, 'boolean'), c('updatedAt', 175, 'time')],
    tabs: ['overview', 'permissions', 'menus', 'administrators', 'operationLogs'],
    form: [fForm('code', 'text', undefined, true), fForm('nameZh'), fForm('nameEn'), fForm('descriptionZh'), fForm('descriptionEn'), fForm('dataScope', 'select', ['ALL', 'DEPARTMENT', 'DEPARTMENT_AND_CHILDREN', 'SELF', 'CUSTOM']), fForm('enabled', 'boolean', undefined, true)],
    loader: systemAPI.roles.list, detail: (id) => detail(systemAPI.roles.detail(id)), create: systemAPI.roles.create, update: systemAPI.roles.update, remove: systemAPI.roles.remove, status: systemAPI.roles.status,
  },
  permissions: {
    permission: 'system:permission:list', createPermission: 'system:permission:create', updatePermission: 'system:permission:update', deletePermission: 'system:permission:delete',
    filters: [f('code'), f('name'), f('module'), f('permissionType', 'select', ['MENU', 'PAGE', 'ACTION', 'API', 'DATA']), f('status', 'select', ['ENABLED', 'DISABLED'])],
    columns: [c('code', 220), c('nameZh'), c('nameEn'), c('module'), c('permissionType', 110, 'status'), c('apiPath', 230), c('assignedRoleCount', 130), c('enabled', 95, 'boolean')],
    tabs: ['overview', 'assignedRoles', 'operationLogs'],
    form: [fForm('code', 'text', undefined, true), fForm('nameZh'), fForm('nameEn'), fForm('module'), fForm('permissionType', 'select', ['MENU', 'PAGE', 'ACTION', 'API', 'DATA']), fForm('apiPath'), fForm('enabled', 'boolean', undefined, true)],
    loader: systemAPI.permissions.list, detail: (id) => detail(systemAPI.permissions.detail(id)), create: systemAPI.permissions.create, update: systemAPI.permissions.update, remove: systemAPI.permissions.remove, status: systemAPI.permissions.status,
  },
  departments: {
    permission: 'system:department:list', createPermission: 'system:department:create', updatePermission: 'system:department:update', deletePermission: 'system:department:delete',
    filters: [f('code'), f('name'), f('parentID'), f('status', 'select', ['ENABLED', 'DISABLED'])],
    columns: [c('code'), c('nameZh'), c('nameEn'), c('parent'), c('leader'), c('administratorCount', 140), c('childCount', 110), c('sortOrder', 100), c('enabled', 95, 'boolean'), c('updatedAt', 175, 'time')],
    tabs: ['overview', 'administrators', 'operationLogs'],
    form: [fForm('code', 'text', undefined, true), fForm('parentID', 'number'), fForm('nameZh'), fForm('nameEn'), fForm('leaderID', 'number'), fForm('sortOrder', 'number'), fForm('enabled', 'boolean', undefined, true)],
    loader: systemAPI.departments.list, detail: (id) => detail(systemAPI.departments.detail(id)), create: systemAPI.departments.create, update: systemAPI.departments.update, remove: systemAPI.departments.remove, status: systemAPI.departments.status,
  },
  menus: {
    permission: 'system:menu:list', createPermission: 'system:menu:create', updatePermission: 'system:menu:update', deletePermission: 'system:menu:delete',
    filters: [f('code'), f('name'), f('menuType', 'select', ['DIRECTORY', 'MENU', 'BUTTON', 'EXTERNAL_LINK']), f('permissionCode'), f('status', 'select', ['ENABLED', 'DISABLED'])],
    columns: [c('code'), c('nameZh'), c('nameEn'), c('menuType', 125, 'status'), c('parent'), c('routePath', 200), c('icon'), c('permissionCode', 210), c('sortOrder', 100), c('visible', 90, 'boolean'), c('enabled', 90, 'boolean')],
    tabs: ['overview', 'children', 'assignedRoles', 'operationLogs'],
    form: [fForm('code', 'text', undefined, true), fForm('parentID', 'number'), fForm('menuType', 'select', ['DIRECTORY', 'MENU', 'BUTTON', 'EXTERNAL_LINK']), fForm('nameZh'), fForm('nameEn'), fForm('routePath'), fForm('icon'), fForm('permissionCode'), fForm('sortOrder', 'number'), fForm('visible', 'boolean'), fForm('enabled', 'boolean', undefined, true), fForm('cacheable', 'boolean'), fForm('externalLink')],
    loader: systemAPI.menus.list, detail: (id) => detail(systemAPI.menus.detail(id)), create: systemAPI.menus.create, update: systemAPI.menus.update, remove: systemAPI.menus.remove, status: systemAPI.menus.status,
  },
  'system-config': {
    permission: 'system:config:list', updatePermission: 'system:config:update',
    filters: [f('key'), f('name'), f('category', 'select', ['PLATFORM', 'BASIC', 'SECURITY', 'WALLET', 'PAYMENT', 'RISK', 'NOTIFICATION', 'FILE_STORAGE']), f('valueType', 'select', ['STRING', 'NUMBER', 'BOOLEAN', 'JSON', 'ENUM', 'SECRET'])],
    columns: [c('key', 240), c('nameZh'), c('nameEn'), c('category', 140, 'status'), c('value', 220), c('valueType', 110, 'status'), c('editable', 95, 'boolean'), c('sensitive', 95, 'boolean'), c('restartRequired', 125, 'boolean'), c('updatedAt', 175, 'time')],
    tabs: ['overview'], form: [fForm('value'), fForm('reason')], loader: systemAPI.configs.list,
  },
  'country-config': {
    permission: 'system:country:list', createPermission: 'system:country:create', updatePermission: 'system:country:update', statusPermission: 'system:country:status', deletePermission: 'system:country:delete',
    filters: [f('countryCode'), f('countryName'), f('currencyCode'), f('status', 'select', ['ENABLED', 'DISABLED']), f('isDefault', 'select', ['YES', 'NO']), dates('created')],
    columns: [c('countryCode', 110), c('isoCountryCode', 125), c('countryName', 170), c('currencyCode', 115), c('currencySymbol', 115), c('phoneCountryCode', 115), c('phoneLength', 110), c('defaultLanguage', 130), c('timezone', 180), c('isDefault', 100, 'boolean'), c('enabled', 95, 'boolean'), c('updatedAt', 175, 'time')],
    tabs: ['overview', 'logoAssets', 'theme', 'loadingAssets', 'operationLogs'],
    form: [
      fForm('countryCode'),
      fForm('isoCountryCode', 'select', countryCodeOptions),
      fForm('countryNameZh'),
      fForm('countryNameEn'),
      fForm('currencyCode'),
      fForm('currencyNameZh'),
      fForm('currencyNameEn'),
      fForm('currencySymbol'),
      fForm('currencyDecimalPlaces', 'number'),
      fForm('timezone', 'select', timezoneOptions),
      fForm('defaultLanguage', 'select', languageOptions),
      fForm('supportedLanguages', 'multiSelect', languageOptions),
      fForm('phoneCountryCode', 'select', phoneCountryCodeOptions),
      fForm('phoneLength', 'number'),
      fForm('enabled', 'boolean', undefined, true),
      fForm('logoUrl', 'asset', undefined, false, 'image'),
      fForm('faviconUrl', 'asset', undefined, false, 'image'),
      fForm('loginLogoUrl', 'asset', undefined, false, 'image'),
      fForm('headerLogoUrl', 'asset', undefined, false, 'image'),
      fForm('loadingLogoUrl', 'asset', undefined, false, 'image'),
      fForm('appIconUrl', 'asset', undefined, false, 'image'),
      fForm('loadingAnimationUrl', 'asset', undefined, false, 'lottie'),
      fForm('loadingBackgroundUrl', 'asset', undefined, false, 'image'),
      fForm('primaryColor', 'color'),
      fForm('secondaryColor', 'color'),
      fForm('buttonColor', 'color'),
      fForm('backgroundColor', 'color'),
      fForm('textColor', 'color'),
      fForm('reason'),
    ],
    loader: systemAPI.countries.list, detail: (id) => detail(systemAPI.countries.detail(id)), create: systemAPI.countries.create, update: systemAPI.countries.update, remove: systemAPI.countries.remove, status: systemAPI.countries.status,
  },
  'audit-logs': {
    permission: 'system:audit-log:view', readOnly: true,
    filters: [f('operator'), f('module'), f('action'), f('targetType'), f('result', 'select', ['SUCCESS', 'FAILED']), f('ip'), f('correlationID'), dates('created')],
    columns: [c('operator', 150), c('module'), c('action', 210), c('targetType'), c('result', 105, 'status'), c('ip', 135), c('createdAt', 175, 'time')],
    tabs: ['overview', 'beforeState', 'afterState', 'requestInformation', 'errorInformation'], form: [], loader: systemAPI.auditLogs.list, detail: (id) => detail(systemAPI.auditLogs.detail(id)),
  },
  'admin-login-logs': {
    permission: 'system:login-log:view', readOnly: true,
    filters: [f('administrator'), f('loginResult', 'select', ['SUCCESS', 'FAILED']), f('ip'), f('country'), f('device'), f('browser'), dates('created')],
    columns: [c('logID', 100), c('administrator'), c('loginResult', 110, 'status'), c('failureReason', 180), c('ip'), c('country'), c('device'), c('operatingSystem'), c('browser'), c('createdAt', 175, 'time')],
    tabs: ['overview', 'requestInformation'], form: [], loader: systemAPI.loginLogs.list, detail: (id) => detail(systemAPI.loginLogs.detail(id)),
  },
  'scheduled-tasks': {
    permission: 'system:task:list', createPermission: 'system:task:create', updatePermission: 'system:task:update', deletePermission: 'system:task:delete',
    filters: [f('taskCode'), f('taskName'), f('taskType'), f('status', 'select', ['ENABLED', 'DISABLED']), f('lastRunResult', 'select', ['SUCCESS', 'FAILED']), dates('nextRun')],
    columns: [c('taskCode', 190), c('nameZh'), c('nameEn'), c('taskType'), c('cronExpression', 130), c('enabled', 95, 'boolean'), c('lastRunAt', 175, 'time'), c('lastRunResult', 120, 'status'), c('nextRunAt', 175, 'time'), c('updatedAt', 175, 'time')],
    tabs: ['overview', 'executionLogs', 'registeredHandlers', 'operationLogs'],
    form: [fForm('taskCode', 'text', undefined, true), fForm('nameZh'), fForm('nameEn'), fForm('taskType'), fForm('handlerCode', 'select', ['report.daily_aggregation', 'system.session_cleanup', 'operation.expired_bonus']), fForm('cronExpression'), fForm('payload', 'json'), fForm('timeoutSeconds', 'number'), fForm('enabled', 'boolean', undefined, true)],
    loader: systemAPI.tasks.list, detail: (id) => detail(systemAPI.tasks.detail(id)), create: systemAPI.tasks.create, update: systemAPI.tasks.update, remove: systemAPI.tasks.remove, status: systemAPI.tasks.status,
  },
}

function fForm(key: string, type: FormField['type'] = 'text', options?: string[], createOnly = false, assetKind?: FormField['assetKind']): FormField {
  return { key, type, options, createOnly, assetKind }
}

export function rowID(row: SystemRecord) { return Number(row.id || row.administratorID || row.logID || 0) }
