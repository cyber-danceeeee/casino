import { h5Api, type H5MusicConfig, type H5MusicTrack } from './h5Api'
import { prefetchMusicURLs } from './assetCache'
import { runtimeConfig } from './runtimeConfig'

export interface H5MusicSession {
  config: H5MusicConfig
  playlist: H5MusicTrack[]
}

let sessionPromise: Promise<H5MusicSession> | null = null

const musicManifestPrefix = 'casino-h5-music-manifest'
const lastRandomTrackKey = 'casino-h5-last-random-music'

const manifestKey = () => `${musicManifestPrefix}:${runtimeConfig.value.resourceVersion || 'default'}`

const readMusicManifest = () => {
  try {
    const config = JSON.parse(window.localStorage.getItem(manifestKey()) || '') as H5MusicConfig
    return config?.mode && Array.isArray(config.tracks) ? config : null
  } catch {
    return null
  }
}

const writeMusicManifest = (config: H5MusicConfig) => {
  try {
    for (let index = window.localStorage.length - 1; index >= 0; index -= 1) {
      const key = window.localStorage.key(index)
      if (key?.startsWith(`${musicManifestPrefix}:`) && key !== manifestKey()) {
        window.localStorage.removeItem(key)
      }
    }
    window.localStorage.setItem(manifestKey(), JSON.stringify(config))
  } catch {
    // Cache Storage still handles audio when metadata persistence is unavailable.
  }
}

const readLastRandomTrackID = () => {
  try {
    return Number(window.localStorage.getItem(lastRandomTrackKey)) || 0
  } catch {
    return 0
  }
}

const writeLastRandomTrackID = (id: number) => {
  try {
    window.localStorage.setItem(lastRandomTrackKey, String(id))
  } catch {
    // A repeated track is harmless when storage is unavailable.
  }
}

const shuffleTracks = (tracks: H5MusicTrack[]) => {
  const shuffled = [...tracks]
  for (let index = shuffled.length - 1; index > 0; index -= 1) {
    const target = Math.floor(Math.random() * (index + 1))
    ;[shuffled[index], shuffled[target]] = [shuffled[target], shuffled[index]]
  }
  return shuffled
}

const createMusicSession = (config: H5MusicConfig): H5MusicSession => {
  const tracks = (config.tracks || []).filter((track) => track.audioUrl)
  if (config.mode === 'RANDOM') {
    const previousTrackID = readLastRandomTrackID()
    const randomPool = tracks.length > 1
      ? tracks.filter((track) => track.id !== previousTrackID)
      : tracks
    const firstTrack = randomPool[Math.floor(Math.random() * randomPool.length)]
    const remainingTracks = shuffleTracks(tracks.filter((track) => track.id !== firstTrack?.id))
    if (firstTrack) writeLastRandomTrackID(firstTrack.id)
    return { config, playlist: firstTrack ? [firstTrack, ...remainingTracks] : [] }
  }

  const specifiedIndex = tracks.findIndex((track) => track.id === config.specifiedMusicID)
  const startIndex = Math.max(0, specifiedIndex)
  return {
    config,
    playlist: [...tracks.slice(startIndex), ...tracks.slice(0, startIndex)],
  }
}

const refreshMusicManifest = async () => {
  const config = await h5Api.getMusicConfig()
  writeMusicManifest(config)
  void prefetchMusicURLs((config.tracks || []).map((track) => track.audioUrl))
  return config
}

export const loadMusicSession = () => {
  if (sessionPromise) return sessionPromise
  const cachedConfig = readMusicManifest()
  sessionPromise = (cachedConfig
    ? Promise.resolve(cachedConfig).then((config) => {
        void refreshMusicManifest().catch((error) => {
          console.warn('[background-music] failed to refresh cached music configuration.', error)
        })
        return config
      })
    : refreshMusicManifest())
    .then((config) => {
      const session = createMusicSession(config)
      void prefetchMusicURLs(session.playlist.map((track) => track.audioUrl))
      return session
    })
    .catch((error) => {
      sessionPromise = null
      throw error
    })
  return sessionPromise
}

export const loadMusicSessionWithin = async (timeoutMs: number) => {
  let timer = 0
  try {
    return await Promise.race([
      loadMusicSession(),
      new Promise<null>((resolve) => {
        timer = window.setTimeout(() => resolve(null), timeoutMs)
      }),
    ])
  } finally {
    if (timer) window.clearTimeout(timer)
  }
}
