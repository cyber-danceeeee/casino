import { apiDelete, apiGet, apiPatch, apiPost, apiPut, http, type ApiResponse } from './http'
import type { PageResult } from './adminUsers'

export type GameCenterQuery = Record<string, string | number | undefined>
export type GameCenterRecord = Record<string, unknown> & { id: number }
export type GameCenterDetail = Record<string, unknown>

export type GameInput = {
  backgroundUrl?: string
  launchImageUrl?: string
  maintenanceMessage?: string
  rtp?: number | null
  volatility?: string
  supportsDemo?: boolean
  supportsFreeSpins?: boolean
  name: string
  localThumbnailUrl: string
  platformEnabled: boolean
  isHot: boolean
  isFeatured: boolean
  isNew: boolean
  sortOrder: number
  supportedCountries: string
  supportedCurrencies: string
  supportedDevices: string
  reason: string
}

export type GameBatchInput = {
  ids: number[]
  platformEnabled?: boolean
  isHot?: boolean
  isFeatured?: boolean
  isNew?: boolean
  addCategoryIds?: number[]
  reason?: string
}

export type GameImportInput = {
  ids: number[]
  reason?: string
}

export type GameImportResult = {
  importedCount: number
  skippedCount: number
  sourceGameIDs: number[]
}

export type TaxonomyInput = {
  code: string
  nameZh: string
  nameEn: string
  icon: string
  sortOrder: number
  enabled: boolean
  reason?: string
}

export type CollectionInput = {
  code: string
  nameZh: string
  nameEn: string
  descriptionZh?: string
  descriptionEn?: string
  bannerUrl?: string
  sortOrder: number
  enabled: boolean
  startTime?: string
  endTime?: string
  games?: Array<{ gameId: number; sortOrder: number }>
  reason?: string
}


export type ProviderGameTypeOption = {
  providerCode: string
  source: 'providerCategory' | 'gameType'
  code: string
  gameCount: number
}

export type ProviderCategoryMappingInput = {
  providerCode: string
  source: 'providerCategory' | 'gameType'
  code: string
  categoryIds: number[]
  reason?: string
}

export type CategoryGameCandidateQuery = {
  page: number
  pageSize: number
  providerCode: string
  source: 'providerCategory' | 'gameType'
  categoryCode: string
  keyword?: string
}

export type GameAsset = {
  url: string
  name: string
  size: number
  mimeType: string
}

export async function uploadGameAsset(file: File) {
  const form = new FormData()
  form.append('file', file)
  const { data } = await http.post<ApiResponse<GameAsset>>('/admin/game-assets/upload', form)
  if (data.code !== 0) throw new Error(data.msg)
  return data.data
}

export const fetchProviders = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/game-providers', params)
export const fetchProvider = (code: string) => apiGet<GameCenterDetail>(`/admin/game-providers/${encodeURIComponent(code)}`)
export const setProviderStatus = (code: string, enabled: boolean, reason: string) => apiPatch<void>(`/admin/game-providers/${encodeURIComponent(code)}/status`, { enabled, reason })
export const testProvider = (code: string) => apiPost<Record<string, unknown>>(`/admin/game-providers/${encodeURIComponent(code)}/test`)
export const syncProvider = (code: string, reason: string) => apiPost<GameCenterRecord>(`/admin/game-providers/${encodeURIComponent(code)}/sync`, { reason })
export const cancelProviderSync = (code: string, reason: string) => apiPost<GameCenterRecord>(`/admin/game-providers/${encodeURIComponent(code)}/sync/cancel`, { reason })
export const deleteProvider = (code: string, reason: string) => apiDelete<{ providerCode: string; deletedCount: number; providerDeleted: boolean }>(`/admin/game-providers/${encodeURIComponent(code)}`, { reason })
export const fetchProviderGameTypes = (code: string) => apiGet<ProviderGameTypeOption[]>(`/admin/game-providers/${encodeURIComponent(code)}/game-types`)
export const fetchProviderOptionGameTypes = (code: string) => apiGet<ProviderGameTypeOption[]>(`/admin/game-provider-options/${encodeURIComponent(code)}/game-types`)
export const fetchImportableProviderGameTypes = (code: string) => apiGet<ProviderGameTypeOption[]>(`/admin/games/catalog/provider-options/${encodeURIComponent(code)}/game-types`)
export const syncAllProviders = (reason: string) => apiPost<GameCenterRecord[]>('/admin/game-sync', { reason })

export const fetchGames = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/games', params)
export const fetchImportableGames = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/games/catalog', params)
export const fetchGame = (id: number) => apiGet<GameCenterDetail>(`/admin/games/${id}`)
export const updateGame = (id: number, payload: GameInput) => apiPatch<void>(`/admin/games/${id}`, payload)
export const batchUpdateGames = (payload: GameBatchInput) => apiPatch<{ updatedCount: number }>('/admin/games/batch', payload)
export const importGames = (payload: GameImportInput) => apiPost<GameImportResult>('/admin/games/import', payload)
export const setGameStatus = (id: number, enabled: boolean, reason: string) => apiPatch<void>(`/admin/games/${id}/status`, { enabled, reason })
export const setGameCategories = (id: number, ids: number[], reason = 'inline_category_update') => apiPut<void>(`/admin/games/${id}/categories`, { ids, reason })
export const applyProviderCategoryMapping = (payload: ProviderCategoryMappingInput) => apiPost<Record<string, unknown>>('/admin/game-category-mapping', payload)
export const launchGameDemo = (id: number) => apiPost<{ launchUrl: string; expiresAt?: string }>(`/admin/games/${id}/demo`)

export const fetchCategories = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/game-categories', params)
export const fetchCategory = (id: number) => apiGet<GameCenterDetail>(`/admin/game-categories/${id}`)
export const fetchCategoryGames = (id: number, params: GameCenterQuery) =>
  apiGet<PageResult<GameCenterRecord>>(`/admin/game-categories/${id}/games`, params)
export const fetchCategoryGameCandidates = (id: number, params: CategoryGameCandidateQuery) =>
  apiGet<PageResult<GameCenterRecord>>(`/admin/game-categories/${id}/candidate-games`, params)
export const fetchCategoryOperationLogs = (id: number, params: GameCenterQuery) =>
  apiGet<PageResult<GameCenterRecord>>(`/admin/game-categories/${id}/operation-logs`, params)
export const createCategory = (payload: TaxonomyInput) => apiPost<{ id: number }>('/admin/game-categories', payload)
export const updateCategory = (id: number, payload: TaxonomyInput) => apiPatch<void>(`/admin/game-categories/${id}`, payload)
export const addCategoryGames = (id: number, ids: number[]) => apiPost<void>(`/admin/game-categories/${id}/games`, { ids })
export const updateCategoryGameSort = (id: number, gameId: number, sortOrder: number) =>
  apiPatch<void>(`/admin/game-categories/${id}/games/${gameId}`, { sortOrder })
export const removeCategoryGame = (id: number, gameId: number) =>
  apiDelete<void>(`/admin/game-categories/${id}/games/${gameId}`, {})
export const deleteCategory = (id: number, reason: string) => apiDelete<void>(`/admin/game-categories/${id}`, { reason })

export const fetchTags = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/game-tags', params)
export const fetchTag = (id: number) => apiGet<GameCenterDetail>(`/admin/game-tags/${id}`)
export const createTag = (payload: TaxonomyInput) => apiPost<{ id: number }>('/admin/game-tags', payload)
export const updateTag = (id: number, payload: TaxonomyInput) => apiPatch<void>(`/admin/game-tags/${id}`, payload)
export const setGameTags = (id: number, ids: number[], reason = 'inline_tag_update') => apiPut<void>(`/admin/games/${id}/tags`, { ids, reason })
export const setTagGames = (id: number, ids: number[], reason = 'tag_games_update') => apiPut<void>(`/admin/game-tags/${id}/games`, { ids, reason })
export const deleteTag = (id: number, reason: string) => apiDelete<void>(`/admin/game-tags/${id}`, { reason })

export const fetchCollections = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/game-collections', params)
export const fetchCollection = (id: number) => apiGet<GameCenterDetail>(`/admin/game-collections/${id}`)
export const createCollection = (payload: CollectionInput) => apiPost<{ id: number }>('/admin/game-collections', payload)
export const updateCollection = (id: number, payload: CollectionInput) => apiPatch<void>(`/admin/game-collections/${id}`, payload)
export const setGameCollections = (id: number, ids: number[], reason = 'inline_collection_update') => apiPut<void>(`/admin/games/${id}/collections`, { ids, reason })
export const deleteCollection = (id: number, reason: string) => apiDelete<void>(`/admin/game-collections/${id}`, { reason })

export const fetchSyncLogs = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/game-sync-logs', params)
export const fetchSyncLog = (id: number) => apiGet<GameCenterDetail>(`/admin/game-sync-logs/${id}`)
export const fetchSyncLogItems = (id: number, type: 'failed' | 'changed', params: GameCenterQuery) =>
  apiGet<PageResult<GameCenterRecord>>(`/admin/game-sync-logs/${id}/items`, { ...params, type })
export const fetchSyncLogRawResponse = (id: number) => apiGet<GameCenterDetail>(`/admin/game-sync-logs/${id}/raw-response`)
export const fetchSyncLogOperationLogs = (id: number, params: GameCenterQuery) =>
  apiGet<PageResult<GameCenterRecord>>(`/admin/game-sync-logs/${id}/operation-logs`, params)
export const fetchLaunchLogs = (params: GameCenterQuery) => apiGet<PageResult<GameCenterRecord>>('/admin/game-launch-logs', params)
export const fetchLaunchLog = (id: number) => apiGet<GameCenterDetail>(`/admin/game-launch-logs/${id}`)
