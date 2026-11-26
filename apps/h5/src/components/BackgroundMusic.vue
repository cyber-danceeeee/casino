<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { H5MusicTrack } from '../services/h5Api'
import { loadMusicSession } from '../services/musicSession'
import { usePreferences } from '../utils/preferences'

const route = useRoute()
const audioRef = ref<HTMLAudioElement | null>(null)
const music = ref('')
const track = ref<H5MusicTrack | null>(null)
const playlist = ref<H5MusicTrack[]>([])
const currentTrackIndex = ref(-1)
const loading = ref(false)
const unlocked = ref(false)
const unlockListenersActive = ref(false)
let retryTimer: number | undefined
const { musicEnabled } = usePreferences()
const canPlayMusic = computed(() => route.name !== 'game-play')

const play = async (muted: boolean) => {
  const audio = audioRef.value
  if (!audio || !musicEnabled.value || !canPlayMusic.value) return false

  audio.muted = muted
  audio.volume = Math.min(1, Math.max(0, Number(track.value?.volumePercent ?? 32) / 100))
  try {
    await audio.play()
    if (!muted) {
      unlocked.value = true
      removeUnlockListeners()
    }
    return true
  } catch {
    addUnlockListeners()
    return false
  }
}

const startPlayback = async () => {
  if (!musicEnabled.value || !canPlayMusic.value) return
  if (unlocked.value) {
    await play(false)
    return
  }
  if (await play(false)) return

  // Browsers that block audible autoplay can still preload muted audio. The
  // first user gesture below unmutes the already-running track immediately.
  await play(true)
}

const scheduleRetry = () => {
  if (retryTimer !== undefined) return
  retryTimer = window.setTimeout(() => {
    retryTimer = undefined
    void loadMusic()
  }, 3000)
}

const selectTrack = async (index: number) => {
  const selected = playlist.value[index]
  if (!selected) return

  currentTrackIndex.value = index
  track.value = selected
  music.value = selected.audioUrl
  await nextTick()
  audioRef.value?.load()
}

const loadMusic = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const session = await loadMusicSession()
    const tracks = session.playlist
    if (!tracks.length) {
      scheduleRetry()
      return
    }

    playlist.value = tracks
    await selectTrack(0)
    if (musicEnabled.value && canPlayMusic.value) {
      addUnlockListeners()
      await startPlayback()
    }
  } catch (error) {
    console.warn('[background-music] failed to load music configuration.', error)
    scheduleRetry()
  } finally {
    loading.value = false
  }
}

const unlock = () => {
  removeUnlockListeners()
  unlocked.value = true
  void play(false)
}

const pause = () => {
  const audio = audioRef.value
  if (!audio) return

  audio.pause()
  audio.muted = true
  removeUnlockListeners()
}

const handleMusicPreference = (event: Event) => {
  const enabled = (event as CustomEvent<{ enabled: boolean }>).detail?.enabled
  if (!enabled || !canPlayMusic.value) {
    pause()
    return
  }

  addUnlockListeners()
  void play(false)
}

const addUnlockListeners = () => {
  if (!musicEnabled.value || !canPlayMusic.value || unlockListenersActive.value) return
  unlockListenersActive.value = true
  window.addEventListener('pointerdown', unlock, { capture: true, passive: true })
  window.addEventListener('touchend', unlock, { capture: true, passive: true })
  window.addEventListener('keydown', unlock, { capture: true })
}

const removeUnlockListeners = () => {
  unlockListenersActive.value = false
  window.removeEventListener('pointerdown', unlock, true)
  window.removeEventListener('touchend', unlock, true)
  window.removeEventListener('keydown', unlock, true)
}

const handleTrackError = async () => {
  if (playlist.value.length < 2) return

  const nextIndex = (currentTrackIndex.value + 1) % playlist.value.length
  if (nextIndex === 0) return
  await selectTrack(nextIndex)
  await startPlayback()
}

onMounted(() => {
  window.addEventListener('game-casino:music-preference', handleMusicPreference)
  addUnlockListeners()
  void loadMusic()
})

onBeforeUnmount(() => {
  window.removeEventListener('game-casino:music-preference', handleMusicPreference)
  if (retryTimer !== undefined) window.clearTimeout(retryTimer)
  removeUnlockListeners()
})

watch([musicEnabled, canPlayMusic], async ([enabled, allowed]) => {
  const audio = audioRef.value

  if (!enabled || !allowed) {
    if (audio) pause()
    return
  }

  await nextTick()
  addUnlockListeners()
  await startPlayback()
})
</script>

<template>
  <audio
    v-if="music && canPlayMusic"
    ref="audioRef"
    :src="music"
    :loop="track?.loop !== false"
    preload="auto"
    autoplay
    muted
    playsinline
    @error="handleTrackError"
  />
</template>
