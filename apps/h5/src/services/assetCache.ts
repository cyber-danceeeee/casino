const workerPath = '/asset-cache-sw.js'
const messageTimeoutMs = 2500
const registrationTimeoutMs = 2000

let registrationPromise: Promise<ServiceWorkerRegistration | null> | null = null

const waitForController = () => {
  if (navigator.serviceWorker.controller) return Promise.resolve()
  return new Promise<void>((resolve) => {
    const timer = window.setTimeout(done, 1500)
    function done() {
      window.clearTimeout(timer)
      navigator.serviceWorker.removeEventListener('controllerchange', done)
      resolve()
    }
    navigator.serviceWorker.addEventListener('controllerchange', done, { once: true })
  })
}

export const registerAssetCache = () => {
  if (registrationPromise) return registrationPromise
  if (!('serviceWorker' in navigator) || !window.isSecureContext) {
    registrationPromise = Promise.resolve(null)
    return registrationPromise
  }
  const hadController = Boolean(navigator.serviceWorker.controller)
  let refreshing = false
  navigator.serviceWorker.addEventListener('controllerchange', () => {
    if (!hadController || refreshing) return
    refreshing = true
    window.location.reload()
  })

  registrationPromise = navigator.serviceWorker
    .register(workerPath, { scope: '/', updateViaCache: 'none' })
    .then(async (registration) => {
      void registration.update().catch(() => undefined)
      await Promise.race([
        navigator.serviceWorker.ready,
        new Promise<void>((resolve) => window.setTimeout(resolve, registrationTimeoutMs)),
      ])
      await waitForController()
      return registration
    })
    .catch((error) => {
      console.warn('[asset-cache] service worker registration failed.', error)
      return null
    })
  return registrationPromise
}

const sendMessage = async (payload: Record<string, unknown>) => {
  const registration = await registerAssetCache()
  const worker = navigator.serviceWorker.controller || registration?.active
  if (!worker) return false
  return new Promise<boolean>((resolve) => {
    const channel = new MessageChannel()
    const timer = window.setTimeout(() => resolve(false), messageTimeoutMs)
    channel.port1.onmessage = (event) => {
      window.clearTimeout(timer)
      resolve(Boolean(event.data?.ok))
    }
    worker.postMessage(payload, [channel.port2])
  })
}

export const configureAssetCache = (version: string) =>
  sendMessage({ type: 'SET_ASSET_VERSION', version })

export const prefetchAssetURLs = (urls: string[]) => {
  const unique = Array.from(new Set(urls.filter(Boolean)))
  if (!unique.length) return Promise.resolve(false)
  return sendMessage({ type: 'PREFETCH_ASSETS', urls: unique })
}

export const prefetchMusicURLs = (urls: string[]) => {
  const unique = Array.from(new Set(urls.filter(Boolean)))
  if (!unique.length) return Promise.resolve(false)
  return sendMessage({ type: 'PREFETCH_MUSIC', urls: unique })
}

export const cacheApplicationShell = () => {
  const urls = [
    `${window.location.origin}/`,
    `${window.location.origin}/manifest.webmanifest`,
    ...Array.from(document.querySelectorAll<HTMLScriptElement>('script[src]'), (element) => element.src),
    ...Array.from(
      document.querySelectorAll<HTMLLinkElement>('link[rel="stylesheet"][href]'),
      (element) => element.href,
    ),
  ]
  return sendMessage({ type: 'CACHE_APP_SHELL', urls: Array.from(new Set(urls)) })
}
