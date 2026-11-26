const CACHE_NAME = 'casino-h5-assets-v2'
const MUSIC_CACHE_NAME = 'casino-h5-music-v1'
const CACHE_META_URL = new URL('/__casino_asset_cache_version__', self.location.origin).href
const MAX_CACHE_ENTRIES = 500
const MAX_MUSIC_CACHE_ENTRIES = 40
const SHELL_CACHE_PREFIX = 'casino-h5-shell-'
const SHELL_CACHE_NAME = `${SHELL_CACHE_PREFIX}v1`
const SHELL_FALLBACK_URL = new URL('/', self.location.origin).href
const SHELL_MAX_CACHE_ENTRIES = 80
const INSTALL_SHELL_URLS = [
  '/',
  '/manifest.webmanifest',
  '/pwa/icon-192.png',
  '/pwa/icon-512.png',
  '/pwa/icon-maskable-512.png',
  '/pwa/apple-touch-icon.png',
]

const assetExtensions = /\.(?:avif|gif|jpe?g|json|mp3|mp4|ogg|png|svg|wav|webm|webp|woff2?|ttf)(?:$|\?)/i
const audioExtensions = /\.(?:aac|flac|m4a|mp3|oga|ogg|opus|wav)(?:$|\?)/i
const signedQueryKeys = /^(?:authorization|expires|ossaccesskeyid|signature|x-amz-|x-oss-)/i

const cacheKey = (input) => {
  const url = new URL(typeof input === 'string' ? input : input.url)
  const hasSignature = Array.from(url.searchParams.keys()).some((key) => signedQueryKeys.test(key))
  if (hasSignature) {
    url.search = ''
  } else if (url.origin === self.location.origin && url.pathname.startsWith('/static/')) {
    url.searchParams.delete('_v')
    url.searchParams.delete('v')
  }
  url.hash = ''
  return url.href
}

const shouldCache = (request) => {
  if (request.method !== 'GET') return false
  if (request.url === CACHE_META_URL || request.url.includes('/asset-cache-sw.js')) return false
  if (request.destination && ['audio', 'font', 'image', 'video'].includes(request.destination)) return true
  return assetExtensions.test(new URL(request.url).pathname)
}

const isMusicRequest = (request) => {
  if (request.destination === 'audio') return true
  return audioExtensions.test(new URL(request.url).pathname)
}

const trimCache = async (cache, maxEntries = MAX_CACHE_ENTRIES) => {
  const keys = (await cache.keys()).filter((request) => request.url !== CACHE_META_URL)
  if (keys.length <= maxEntries) return
  await Promise.all(keys.slice(0, keys.length - maxEntries).map((request) => cache.delete(request)))
}

const trimShellCache = async (cache) => {
  const keys = await cache.keys()
  if (keys.length <= SHELL_MAX_CACHE_ENTRIES) return
  await Promise.all(keys.slice(0, keys.length - SHELL_MAX_CACHE_ENTRIES).map((request) => cache.delete(request)))
}

const putResponse = async (cache, key, response, maxEntries = MAX_CACHE_ENTRIES) => {
  if (!(response.ok || response.type === 'opaque')) return
  try {
    await cache.put(key, response.clone())
    await trimCache(cache, maxEntries)
  } catch {
    // Quota pressure must never prevent the asset from rendering.
  }
}

const fetchAndCache = async (request, cache, key, maxEntries) => {
  const response = await fetch(request)
  const contentType = response.headers.get('content-type') || ''
  if (request.destination === 'image' && !contentType.toLowerCase().startsWith('image/')) return response
  await putResponse(cache, key, response, maxEntries)
  return response
}

const cacheFirst = async (request) => {
  const musicRequest = isMusicRequest(request)
  const cache = await caches.open(musicRequest ? MUSIC_CACHE_NAME : CACHE_NAME)
  const maxEntries = musicRequest ? MAX_MUSIC_CACHE_ENTRIES : MAX_CACHE_ENTRIES
  const key = cacheKey(request)
  const cached = await cache.match(key, { ignoreVary: true })
  if (cached) return cached
  return fetchAndCache(request, cache, key, maxEntries)
}

const putShellResponse = async (cache, key, response) => {
  if (!response.ok || response.type === 'opaque') return
  try {
    await cache.put(key, response.clone())
    await trimShellCache(cache)
  } catch {
    // The network response remains usable when storage quota is exhausted.
  }
}

const cacheShellURLs = async (urls) => {
  const cache = await caches.open(SHELL_CACHE_NAME)
  await Promise.allSettled(urls.map(async (value) => {
    const url = new URL(value, self.location.origin)
    if (url.origin !== self.location.origin) return
    const response = await fetch(url.href, { cache: 'no-cache', credentials: 'same-origin' })
    const key = url.pathname === '/' ? SHELL_FALLBACK_URL : url.href
    await putShellResponse(cache, key, response)
  }))
}

const networkFirstNavigation = async (request) => {
  const cache = await caches.open(SHELL_CACHE_NAME)
  try {
    const response = await fetch(request)
    if (response.ok) await putShellResponse(cache, SHELL_FALLBACK_URL, response)
    return response
  } catch (error) {
    const fallback = await cache.match(SHELL_FALLBACK_URL, { ignoreSearch: true })
    if (fallback) return fallback
    throw error
  }
}

const staleWhileRevalidateShell = async (request) => {
  const cache = await caches.open(SHELL_CACHE_NAME)
  const cached = await cache.match(request, { ignoreVary: true })
  const network = fetch(request).then(async (response) => {
    await putShellResponse(cache, request, response)
    return response
  })
  if (cached) {
    void network.catch(() => undefined)
    return cached
  }
  return network
}

const isShellRequest = (request) => {
  const url = new URL(request.url)
  if (url.origin !== self.location.origin) return false
  if (url.pathname.startsWith('/src/') || url.pathname.startsWith('/@vite/') || url.pathname.startsWith('/@id/')) {
    return false
  }
  if (['script', 'style', 'worker'].includes(request.destination)) return true
  return /\.(?:css|js|mjs|webmanifest)(?:$|\?)/i.test(url.pathname)
}

const applyVersion = async (version) => {
  if (!version) return
  let cache = await caches.open(CACHE_NAME)
  const stored = await cache.match(CACHE_META_URL)
  const current = stored ? await stored.text() : ''
  if (current === version) return
  await caches.delete(CACHE_NAME)
  await caches.delete(MUSIC_CACHE_NAME)
  cache = await caches.open(CACHE_NAME)
  await cache.put(CACHE_META_URL, new Response(version, {
    headers: { 'Content-Type': 'text/plain', 'Cache-Control': 'no-store' },
  }))
}

const prefetch = async (urls, forceMusic = false) => {
  for (const value of urls) {
    if (!value) continue
    const url = new URL(value, self.location.origin)
    const musicRequest = forceMusic || audioExtensions.test(url.pathname)
    const cache = await caches.open(musicRequest ? MUSIC_CACHE_NAME : CACHE_NAME)
    const maxEntries = musicRequest ? MAX_MUSIC_CACHE_ENTRIES : MAX_CACHE_ENTRIES
    const key = cacheKey(url.href)
    if (await cache.match(key, { ignoreVary: true })) continue
    try {
      const response = await fetch(url.href, {
        credentials: url.origin === self.location.origin ? 'same-origin' : 'omit',
        mode: url.origin === self.location.origin ? 'same-origin' : 'no-cors',
      })
      await putResponse(cache, key, response, maxEntries)
    } catch {
      // A later DOM request can retry without blocking application startup.
    }
  }
}

self.addEventListener('install', (event) => {
  event.waitUntil(cacheShellURLs(INSTALL_SHELL_URLS).then(() => self.skipWaiting()))
})

self.addEventListener('activate', (event) => {
  event.waitUntil((async () => {
    const names = await caches.keys()
    await Promise.all(
      names
        .filter((name) => name.startsWith(SHELL_CACHE_PREFIX) && name !== SHELL_CACHE_NAME)
        .map((name) => caches.delete(name)),
    )
    await self.clients.claim()
  })())
})

self.addEventListener('fetch', (event) => {
  if (event.request.method !== 'GET') return
  if (event.request.mode === 'navigate') {
    event.respondWith(networkFirstNavigation(event.request))
    return
  }
  if (isShellRequest(event.request)) {
    event.respondWith(staleWhileRevalidateShell(event.request).catch(() => fetch(event.request)))
    return
  }
  if (!shouldCache(event.request)) return
  event.respondWith(cacheFirst(event.request).catch(() => fetch(event.request)))
})

self.addEventListener('message', (event) => {
  const payload = event.data || {}
  const respond = (data) => event.ports[0]?.postMessage(data)
  if (payload.type === 'SET_ASSET_VERSION') {
    event.waitUntil(applyVersion(String(payload.version || '')).then(() => respond({ ok: true })))
    return
  }
  if (payload.type === 'PREFETCH_ASSETS') {
    const urls = Array.isArray(payload.urls) ? payload.urls.map(String) : []
    event.waitUntil(prefetch(urls).then(() => respond({ ok: true })))
    return
  }
  if (payload.type === 'PREFETCH_MUSIC') {
    const urls = Array.isArray(payload.urls) ? payload.urls.map(String) : []
    event.waitUntil(prefetch(urls, true).then(() => respond({ ok: true })))
    return
  }
  if (payload.type === 'CACHE_APP_SHELL') {
    const urls = Array.isArray(payload.urls) ? payload.urls.map(String) : []
    event.waitUntil(cacheShellURLs(urls).then(() => respond({ ok: true })))
  }
})
