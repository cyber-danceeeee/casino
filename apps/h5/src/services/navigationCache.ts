import { h5Api, type H5NavigationItem } from './h5Api'
import { isLoggedIn } from '../mock/session'
import { runtimeConfig } from './runtimeConfig'

const cachedNavigation = new Map<string, H5NavigationItem[]>()
const pendingNavigation = new Map<string, Promise<H5NavigationItem[]>>()

export async function loadNavigation(position: H5NavigationItem['displayPosition']) {
  const key = `${position}:${isLoggedIn.value ? 'member' : 'guest'}:${runtimeConfig.value.language.current}`
  const cached = cachedNavigation.get(key)
  if (cached) return cached
  const pending = pendingNavigation.get(key)
  if (pending) return pending
  const request = h5Api.getNavigation(position)
    .then((items) => {
      const active = items.filter((item) => item.enabled && item.displayPosition === position)
      cachedNavigation.set(key, active)
      return active
    })
    .finally(() => {
      pendingNavigation.delete(key)
    })
  pendingNavigation.set(key, request)
  return request
}

export function clearNavigationCache() {
  cachedNavigation.clear()
  pendingNavigation.clear()
}
