import type { GameSection } from '../mock/games'
import { h5Api, type HomePayload } from './h5Api'
import { runtimeConfig } from './runtimeConfig'
import { h5SessionTokenKey } from '../utils/authStorage'

export interface HomeCatalogSnapshot {
  home: HomePayload
  gameSections: GameSection[]
}

let cachedHomeCatalog: HomeCatalogSnapshot | null = null
let pendingHomeCatalog: Promise<HomeCatalogSnapshot> | null = null
let cachedAt = 0
let pendingSession = ''
let requestVersion = 0

const storagePrefix = 'casino-h5-home-catalog-v5'
const cacheTtlMs = 5 * 60 * 1000

interface StoredHomeCatalog {
  savedAt: number
  snapshot: HomeCatalogSnapshot
}

const storageKey = () => `${storagePrefix}:${runtimeConfig.value.resourceVersion || 'default'}`

const catalogOnly = (snapshot: HomeCatalogSnapshot): HomeCatalogSnapshot => {
  const { balance: _balance, ...raw } = snapshot.home.raw || {}
  return { ...snapshot, home: { ...snapshot.home, balance: null, raw } }
}

const readStoredCatalog = () => {
  if (cachedHomeCatalog || typeof window === 'undefined') return
  try {
    const stored = JSON.parse(window.localStorage.getItem(storageKey()) || '') as StoredHomeCatalog
    if (!stored?.snapshot?.home || !Array.isArray(stored.snapshot.gameSections)) return
    if (!Number.isFinite(stored.savedAt) || Date.now() - stored.savedAt >= cacheTtlMs) {
      window.localStorage.removeItem(storageKey())
      return
    }
    cachedHomeCatalog = catalogOnly(stored.snapshot)
    cachedAt = Number(stored.savedAt) || 0
  } catch {
    // Missing or outdated cache falls through to the API.
  }
}

const storeCatalog = (snapshot: HomeCatalogSnapshot) => {
  snapshot = catalogOnly(snapshot)
  cachedHomeCatalog = snapshot
  cachedAt = Date.now()
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(storageKey(), JSON.stringify({ savedAt: cachedAt, snapshot }))
  } catch {
    // A full/private storage area must not prevent the home page from loading.
  }
}

export const getCachedHomeCatalog = () => {
  readStoredCatalog()
  return cachedHomeCatalog
}

export const loadHomeCatalog = async (force = false, onHome?: (home: HomePayload) => void) => {
  readStoredCatalog()
  if (!force && cachedHomeCatalog && Date.now() - cachedAt < cacheTtlMs) return cachedHomeCatalog
  const session = typeof window === 'undefined' ? '' : window.localStorage.getItem(h5SessionTokenKey) || ''
  if (!force && pendingHomeCatalog && pendingSession === session) return pendingHomeCatalog
  pendingSession = session
  const version = ++requestVersion

  const request: Promise<HomeCatalogSnapshot> = (async () => {
    try {
      const home = await h5Api.getHome()
      onHome?.(home)
      const gameSections = await h5Api.getGameSections(home.categories)

      const snapshot: HomeCatalogSnapshot = {
        home: { ...home, balance: null },
        gameSections,
      }
      if (version === requestVersion) storeCatalog(snapshot)

      return snapshot
    } catch (error) {
      if (cachedHomeCatalog) return cachedHomeCatalog
      throw error
    }
  })().finally(() => {
    if (pendingHomeCatalog === request) pendingHomeCatalog = null
  })
  pendingHomeCatalog = request
  return request
}
