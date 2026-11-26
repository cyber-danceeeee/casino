export function isAndroidAppRuntime() {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') return false

  return (
    /(?:^|\s)CasinoAndroid\/[\w.-]+(?:\s|$)/i.test(navigator.userAgent) ||
    typeof window.CasinoAndroid?.postMessage === 'function'
  )
}
