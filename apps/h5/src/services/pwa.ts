import { computed, ref } from 'vue'

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed'; platform: string }>
}

const installPrompt = ref<BeforeInstallPromptEvent | null>(null)
const installed = ref(false)
let initialized = false

const isStandalone = () =>
  window.matchMedia('(display-mode: standalone)').matches ||
  Boolean((navigator as Navigator & { standalone?: boolean }).standalone)

export const pwaInstallAvailable = computed(
  () => !installed.value && Boolean(installPrompt.value),
)

export const initializePWAInstall = () => {
  if (initialized) return
  initialized = true
  installed.value = isStandalone()

  window.addEventListener('beforeinstallprompt', (event) => {
    event.preventDefault()
    installPrompt.value = event as BeforeInstallPromptEvent
  })

  window.addEventListener('appinstalled', () => {
    installed.value = true
    installPrompt.value = null
  })
}

export const installPWA = async () => {
  if (isStandalone()) {
    installed.value = true
    return true
  }

  const prompt = installPrompt.value
  if (!prompt) return false

  await prompt.prompt()
  const choice = await prompt.userChoice
  installPrompt.value = null
  return choice.outcome === 'accepted'
}
