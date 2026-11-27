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

export const systemConfigs: Record<string, SystemConfig> = {
  admins: {
    permission: 'system:admin:list', createPermission: 'system:admin:create', updatePermission: 'system:admin:update',
    filters: [ f('siteID'), f('site'), f('username'), f('displayName'), f('department'), f('role'), f('status', 'select', ['ENABLED', 'DISABLED']), f('mobile'), f('email'), dates('created'), dates('lastLogin')],
    columns: [ c('site', 150), c('username'), c('displayName'), c('department'), c('roles', 180), c('mobile'), c('email', 190), c('status', 95, 'boolean'), c('lastLoginIP'), c('lastLoginTime', 175, 'time'), c('createdAt', 175, 'time')],
    tabs: ['overview', 'rolesAndPermissions', 'departments', 'loginSecurity', 'loginLogs', 'operationLogs'],
    form: [fForm('username', 'text', undefined, true), fForm('displayName'), fForm('password', 'password', undefined, true), fForm('siteID', 'number'), fForm('departmentID', 'number'), fForm('mobile'), fForm('email'), fForm('status', 'boolean', undefined, true)],
    loader: systemAPI.admins.list, detail: (id) => detail(systemAPI.admins.detail(id)), create: systemAPI.admins.create, update: systemAPI.admins.update, status: systemAPI.admins.status,
  },
  roles: {
    permission: 'system:role:list', createPermission: 'system:role:create', updatePermission: 'system:role:update', deletePermission: 'system:role:delete',
    filters: [ f('code'), f('name'), f('status', 'select', ['ENABLED', 'DISABLED']), f('dataScope', 'select', ['ALL', 'DEPARTMENT', 'DEPARTMENT_AND_CHILDREN', 'SELF', 'CUSTOM']), dates('created')],
    columns: [ c('code'), c('nameZh'), c('nameEn'), c('administratorCount', 135), c('permissionCount', 125), c('dataScope', 190, 'status'), c('enabled', 95, 'boolean'), c('updatedAt', 175, 'time')],
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
