import type { Router } from 'vue-router'
import { isLoggedIn } from '../mock/session'
import type { H5NavigationItem } from './h5Api'
import { primeGameAudio } from '../utils/gameAudio'

function playEntrySound(url: string) {
  if (!url) return
  const audio = new Audio(url)
  audio.volume = 0.7
  void audio.play().catch(() => undefined)
}

function requireLogin(router: Router) {
  if (isLoggedIn.value) return true
  void router.push({ name: 'auth-login', query: { redirect: router.currentRoute.value.fullPath || '/' } })
  return false
}

function openHTTPSURL(value: string) {
  try {
    const target = new URL(value, window.location.origin)
    if (target.protocol !== 'https:') return
    window.open(target.toString(), '_blank', 'noopener,noreferrer')
  } catch {
    // The API validates targets; malformed values are ignored defensively.
  }
}

export async function openNavigationEntry(item: H5NavigationItem, router: Router) {
  playEntrySound(item.soundURL)
  switch (item.actionType) {
    case 'OPEN_INTERNAL_PAGE':
      if (item.actionTarget.startsWith('/') && !item.actionTarget.startsWith('//')) await router.push(item.actionTarget)
      return
    case 'OPEN_EXTERNAL_URL':
    case 'OPEN_DOWNLOAD':
      openHTTPSURL(item.actionTarget)
      return
    case 'OPEN_DEPOSIT':
      if (requireLogin(router)) await router.push({ name: 'wallet-deposit' })
      return
    case 'OPEN_WITHDRAW':
      if (requireLogin(router)) await router.push({ name: 'cash-center' })
      return
    case 'OPEN_SUPPORT':
      await router.push({ name: 'support' })
      return
    case 'OPEN_PROMOTION':
      if (item.actionTarget.startsWith('/')) await router.push(item.actionTarget)
      else await router.push({ name: 'activity-detail', params: { slug: item.actionTarget } })
      return
    case 'OPEN_GAME':
      if (!requireLogin(router)) return
      primeGameAudio()
      await router.push({
        name: 'game-play',
        query: {
          ...(item.gameID ? { gameId: String(item.gameID) } : {}),
          providerCode: item.providerCode,
          providerGameCode: item.providerGameCode,
          gameName: item.gameName,
        },
      })
  }
}
