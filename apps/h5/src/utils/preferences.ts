import { ref } from 'vue'

export type AppLanguage = 'English'

const languageKey = 'game-casino-language'
const musicKey = 'game-casino-music-enabled'
const languageOptions = ['English'] as const

const readLanguage = (): AppLanguage => {
  if (typeof window === 'undefined') return 'English'
  const saved = window.localStorage.getItem(languageKey)
  return languageOptions.includes(saved as AppLanguage) ? (saved as AppLanguage) : 'English'
}

const readMusicEnabled = () => {
  if (typeof window === 'undefined') return true
  return window.localStorage.getItem(musicKey) !== 'false'
}

const currentLanguage = ref<AppLanguage>(readLanguage())
const musicEnabled = ref(readMusicEnabled())

const setLanguage = (language: AppLanguage) => {
  currentLanguage.value = language
  if (typeof document !== 'undefined') document.documentElement.lang = 'en'
  if (typeof window !== 'undefined') window.localStorage.setItem(languageKey, language)
}

const setMusicEnabled = (enabled: boolean) => {
  musicEnabled.value = enabled
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(musicKey, String(enabled))
    window.dispatchEvent(new CustomEvent('game-casino:music-preference', { detail: { enabled } }))
  }
}

const toggleMusicEnabled = () => {
  setMusicEnabled(!musicEnabled.value)
}

export const usePreferences = () => ({
  currentLanguage,
  languageOptions,
  musicEnabled,
  setLanguage,
  setMusicEnabled,
  toggleMusicEnabled,
})
