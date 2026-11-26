import type { GameSection } from '../mock/games'
import { h5Api, type HomePayload } from './h5Api'
import { runtimeConfig } from './runtimeConfig'

export interface HomeCatalogSnapshot {
  home: HomePayload
  gameSections: GameSection[]
}

let cachedHomeCatalog: HomeCatalogSnapshot | null = null
let pendingHomeCatalog: Promise<HomeCatalogSnapshot> | null = null
let cachedAt = 0

const storagePrefix = 'casino-h5-home-catalog-v4'
const cacheTtlMs = 5 * 60 * 1000

interface StoredHomeCatalog {
  savedAt: number
  snapshot: HomeCatalogSnapshot
}

const storageKey = () => `${storagePrefix}:${runtimeConfig.value.resourceVersion || 'default'}`

const readStoredCatalog = () => {
  if (cachedHomeCatalog || typeof window === 'undefined') return
  try {
    const stored = JSON.parse(window.localStorage.getItem(storageKey()) || '') as StoredHomeCatalog
    if (!stored?.snapshot?.home || !Array.isArray(stored.snapshot.gameSections)) return
    if (!Number.isFinite(stored.savedAt) || Date.now() - stored.savedAt >= cacheTtlMs) {
      window.localStorage.removeItem(storageKey())
      return
    }
    cachedHomeCatalog = stored.snapshot
    cachedAt = Number(stored.savedAt) || 0
  } catch {
    // Missing or outdated cache falls through to the API.
  }
}

const storeCatalog = (snapshot: HomeCatalogSnapshot) => {
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

export const loadHomeCatalog = async (force = false) => {
  readStoredCatalog()
  if (!force && cachedHomeCatalog && Date.now() - cachedAt < cacheTtlMs) return cachedHomeCatalog
  if (pendingHomeCatalog) return pendingHomeCatalog

  pendingHomeCatalog = (async () => {
    try {
      const home = await h5Api.getHome()
      const gameSections = await h5Api.getGameSections(home.categories)

      const snapshot: HomeCatalogSnapshot = {
        home,
        gameSections,
      }
      storeCatalog(snapshot)

      return snapshot
    } catch (error) {
      if (cachedHomeCatalog) return cachedHomeCatalog
      throw error
    }
  })().finally(() => {
    pendingHomeCatalog = null
  })

  return pendingHomeCatalog
}
