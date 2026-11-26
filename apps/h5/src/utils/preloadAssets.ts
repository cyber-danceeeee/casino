export type PreloadAssetType = 'image' | 'audio'

export interface PreloadAsset {
  src: string
  type: PreloadAssetType
  label?: string
  timeoutMs?: number
}

export interface PreloadProgress {
  loaded: number
  total: number
  failed: number
  percent: number
  asset?: PreloadAsset
}

export interface PreloadResult {
  loaded: number
  total: number
  failed: number
}

interface PreloadOptions {
  concurrency?: number
  timeoutMs?: number
  onProgress?: (progress: PreloadProgress) => void
}

const DEFAULT_TIMEOUT = 8000
const DEFAULT_CONCURRENCY = 6

const uniqueAssets = (assets: PreloadAsset[]) => {
  const seen = new Map<string, PreloadAsset>()

  for (const asset of assets) {
    if (!asset.src || seen.has(asset.src)) continue
    seen.set(asset.src, asset)
  }

  return Array.from(seen.values())
}

const withTimeout = (task: Promise<void>, timeoutMs: number) =>
  new Promise<boolean>((resolve) => {
    let settled = false
    const timer = window.setTimeout(() => {
      if (settled) return
      settled = true
      resolve(false)
    }, timeoutMs)

    task
      .then(() => {
        if (settled) return
        settled = true
        window.clearTimeout(timer)
        resolve(true)
      })
      .catch(() => {
        if (settled) return
        settled = true
        window.clearTimeout(timer)
        resolve(false)
      })
  })

const preloadImage = (src: string, timeoutMs: number) =>
  withTimeout(
    new Promise<void>((resolve, reject) => {
      const img = new Image()

      img.decoding = 'async'
      img.loading = 'eager'
      img.onload = () => {
        if ('decode' in img) {
          img.decode().then(resolve).catch(resolve)
          return
        }

        resolve()
      }
      img.onerror = reject
      img.src = src
    }),
    timeoutMs,
  )

const preloadAudio = (src: string, timeoutMs: number) =>
  withTimeout(
    new Promise<void>((resolve, reject) => {
      const audio = new Audio()

      const cleanup = () => {
        audio.removeEventListener('canplaythrough', onReady)
        audio.removeEventListener('loadeddata', onReady)
        audio.removeEventListener('error', onError)
      }
      const onReady = () => {
        cleanup()
        resolve()
      }
      const onError = () => {
        cleanup()
        reject(new Error(`Failed to preload audio: ${src}`))
      }

      audio.preload = 'auto'
      audio.muted = true
      audio.addEventListener('canplaythrough', onReady, { once: true })
      audio.addEventListener('loadeddata', onReady, { once: true })
      audio.addEventListener('error', onError, { once: true })
      audio.src = src
      audio.load()
    }),
    timeoutMs,
  )

const preloadOne = (asset: PreloadAsset, timeoutMs: number) => {
  const assetTimeout = asset.timeoutMs ?? timeoutMs

  if (asset.type === 'audio') {
    return preloadAudio(asset.src, assetTimeout)
  }

  return preloadImage(asset.src, assetTimeout)
}

export const preloadAssets = async (
  assets: PreloadAsset[],
  options: PreloadOptions = {},
): Promise<PreloadResult> => {
  const queue = uniqueAssets(assets)
  const total = queue.length
  const concurrency = Math.max(1, Math.min(options.concurrency ?? DEFAULT_CONCURRENCY, total || 1))
  const timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT
  let loaded = 0
  let failed = 0
  let cursor = 0

  options.onProgress?.({ loaded, total, failed, percent: total ? 0 : 100 })

  const report = (asset?: PreloadAsset) => {
    const percent = total ? Math.round((loaded / total) * 100) : 100
    options.onProgress?.({ loaded, total, failed, percent, asset })
  }

  const worker = async () => {
    while (cursor < total) {
      const asset = queue[cursor]
      cursor += 1

      const ok = await preloadOne(asset, timeoutMs)
      loaded += 1
      if (!ok) failed += 1
      report(asset)
    }
  }

  await Promise.all(Array.from({ length: concurrency }, worker))

  return { loaded, total, failed }
}
