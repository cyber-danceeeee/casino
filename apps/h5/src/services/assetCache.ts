const workerPath = '/asset-cache-sw.js'
const messageTimeoutMs = 2500
const registrationTimeoutMs = 2000

let registrationPromise: Promise<ServiceWorkerRegistration | null> | null = null

const serviceWorkers = () => {
  if (typeof window === 'undefined' || !window.isSecureContext || typeof navigator === 'undefined') {
    return null
  }
  try {
    return navigator.serviceWorker ?? null
  } catch {
    return null
  }
}

const waitForController = (workers: ServiceWorkerContainer) => {
  if (workers.controller) return Promise.resolve()
  return new Promise<void>((resolve) => {
    const timer = window.setTimeout(done, 1500)
    function done() {
      window.clearTimeout(timer)
      workers.removeEventListener('controllerchange', done)
      resolve()
    }
    workers.addEventListener('controllerchange', done, { once: true })
  })
}

export const registerAssetCache = () => {
  if (registrationPromise) return registrationPromise
  const workers = serviceWorkers()
  if (!workers) {
    registrationPromise = Promise.resolve(null)
    return registrationPromise
  }
  const hadController = Boolean(workers.controller)
  let refreshing = false
  workers.addEventListener('controllerchange', () => {
    if (!hadController || refreshing) return
    refreshing = true
    window.location.reload()
  })

  registrationPromise = Promise.resolve()
    .then(() => workers.register(workerPath, { scope: '/', updateViaCache: 'none' }))
    .then(async (registration) => {
      void registration.update().catch(() => undefined)
      await Promise.race([
        workers.ready,
        new Promise<void>((resolve) => window.setTimeout(resolve, registrationTimeoutMs)),
      ])
      await waitForController(workers)
      return registration
    })
    .catch((error) => {
      console.warn('[asset-cache] service worker registration failed.', error)
      return null
    })
  return registrationPromise
}

const sendMessage = async (payload: Record<string, unknown>) => {
  const workers = serviceWorkers()
  if (!workers || typeof MessageChannel === 'undefined') return false
  const registration = await registerAssetCache()
  const worker = workers.controller || registration?.active
  if (!worker) return false
  return new Promise<boolean>((resolve) => {
    const channel = new MessageChannel()
    const done = (ok: boolean) => {
      window.clearTimeout(timer)
      channel.port1.close()
      channel.port2.close()
      resolve(ok)
    }
    const timer = window.setTimeout(() => done(false), messageTimeoutMs)
    channel.port1.onmessage = (event) => {
      done(Boolean(event.data?.ok))
    }
    channel.port1.onmessageerror = () => done(false)
    try {
      worker.postMessage(payload, [channel.port2])
    } catch {
      done(false)
    }
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
