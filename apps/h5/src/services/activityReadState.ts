import { h5Api, type H5OfferEventPayload, type H5OffersPayload } from './h5Api'

const storageKey = 'casino-h5-activity-read-v1'
const changeEvent = 'casino:activity-read-change'
const unreadChangeEvent = 'casino:activity-unread-change'
let cachedOffers: H5OffersPayload | null = null
let pendingRequest: Promise<H5OffersPayload> | null = null
let cachedUnreadState: boolean | null = null

type ReadMarkers = Record<string, string>

const eventKey = (event: Pick<H5OfferEventPayload, 'activityCode' | 'slug' | 'id'>) =>
  event.activityCode || event.slug || String(event.id)

const readMarkers = (): ReadMarkers => {
  if (typeof window === 'undefined') return {}
  try {
    return JSON.parse(window.localStorage.getItem(storageKey) || '{}') as ReadMarkers
  } catch {
    return {}
  }
}

export const isActivityUnread = (event: H5OfferEventPayload) => {
  const marker = readMarkers()[eventKey(event)]
  if (!marker) return true
  return marker !== (event.updatedAt || 'read')
}

export const markActivityRead = (event: H5OfferEventPayload) => {
  if (typeof window === 'undefined') return
  const markers = readMarkers()
  markers[eventKey(event)] = event.updatedAt || 'read'
  window.localStorage.setItem(storageKey, JSON.stringify(markers))
  window.dispatchEvent(new CustomEvent(changeEvent))
}

export const onActivityReadChange = (listener: () => void) => {
  if (typeof window === 'undefined') return () => undefined
  window.addEventListener(changeEvent, listener)
  return () => window.removeEventListener(changeEvent, listener)
}

export const hasUnreadActivities = (payload: H5OffersPayload) =>
  (payload.events || []).some((event) => isActivityUnread(event))

export const publishActivityUnreadState = (unread: boolean) => {
  cachedUnreadState = unread
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent<boolean>(unreadChangeEvent, { detail: unread }))
}

export const onActivityUnreadStateChange = (listener: (unread: boolean) => void) => {
  if (typeof window === 'undefined') return () => undefined
  const handler = (event: Event) => listener((event as CustomEvent<boolean>).detail)
  window.addEventListener(unreadChangeEvent, handler)
  return () => window.removeEventListener(unreadChangeEvent, handler)
}

export const loadCachedOffers = async (force = false) => {
  if (!force && cachedOffers) return cachedOffers
  if (!force && pendingRequest) return pendingRequest
  pendingRequest = h5Api.getOffers()
  try {
    cachedOffers = await pendingRequest
    return cachedOffers
  } finally {
    pendingRequest = null
  }
}

export const loadActivityUnreadState = async () => {
  try {
    const unread = hasUnreadActivities(await loadCachedOffers())
    publishActivityUnreadState(unread)
    return unread
  } catch (error) {
    if (cachedUnreadState !== null) return cachedUnreadState
    throw error
  }
}
