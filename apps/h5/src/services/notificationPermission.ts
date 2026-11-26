const promptStorageKey = 'casino-h5-notification-prompt:v1'

export type NotificationPermissionState = 'default' | 'granted' | 'denied' | 'unsupported'

interface AndroidMessageBridge {
  postMessage: (message: string) => void
}

declare global {
  interface Window {
    CasinoAndroid?: AndroidMessageBridge
  }
}

function hasNativeBridge() {
  return typeof window.CasinoAndroid?.postMessage === 'function'
}

function readPromptSeen() {
  try {
    return window.localStorage.getItem(promptStorageKey) === '1'
  } catch {
    return false
  }
}

export function markNotificationPromptSeen() {
  try {
    window.localStorage.setItem(promptStorageKey, '1')
  } catch {
    // A blocked storage API must not prevent the permission request.
  }
}

export function shouldShowNotificationPrompt() {
  if (typeof window === 'undefined' || readPromptSeen()) return false
  if (hasNativeBridge()) return true
  return window.isSecureContext && 'Notification' in window && Notification.permission === 'default'
}

export async function requestNotificationPermission(): Promise<NotificationPermissionState> {
  markNotificationPromptSeen()

  if (hasNativeBridge()) {
    window.CasinoAndroid?.postMessage(JSON.stringify({ type: 'REQUEST_NOTIFICATION_PERMISSION' }))
    return 'default'
  }

  if (!window.isSecureContext || !('Notification' in window)) return 'unsupported'
  if (Notification.permission !== 'default') return Notification.permission
  return Notification.requestPermission()
}
