import { h5Api } from './h5Api'

interface ShareConfigCacheEntry {
  expiresAt: number
  payload: Record<string, unknown>
}

const cache = new Map<string, ShareConfigCacheEntry>()
const pending = new Map<string, Promise<Record<string, unknown>>>()
const cacheTtlMs = 15 * 60 * 1000

export const getCachedShareConfig = (tab: string) => {
  const key = tab.trim().toLowerCase() || 'social'
  const cached = cache.get(key)
  if (cached && cached.expiresAt > Date.now()) {
    return Promise.resolve(cached.payload)
  }

  const activeRequest = pending.get(key)
  if (activeRequest) return activeRequest

  const request = h5Api.getShare(key)
    .then((payload) => {
      cache.set(key, {
        expiresAt: Date.now() + cacheTtlMs,
        payload,
      })
      return payload
    })
    .finally(() => {
      pending.delete(key)
    })

  pending.set(key, request)
  return request
}
