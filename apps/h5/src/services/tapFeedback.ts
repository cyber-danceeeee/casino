const interactiveSelector = [
  'button:not(:disabled)',
  'a[href]',
  '[role="button"]:not([aria-disabled="true"])',
  '[role="tab"]:not([aria-disabled="true"])',
  '[role="menuitem"]:not([aria-disabled="true"])',
  '[role="switch"]:not([aria-disabled="true"])',
  'summary',
  '[data-tap-feedback]',
].join(',')

interface NativeMessageBridge {
  postMessage: (message: string) => void
}

interface ActivePointer {
  target: HTMLElement
  startX: number
  startY: number
}

const activePointers = new Map<number, ActivePointer>()
const releaseTimers = new WeakMap<HTMLElement, number>()
let feedbackLayer: HTMLElement | null = null
let installed = false
let lastHapticAt = 0

const getInteractiveTarget = (eventTarget: EventTarget | null) => {
  if (!(eventTarget instanceof Element)) return null
  if (eventTarget.closest('[data-no-tap-feedback], input, textarea, select, [contenteditable="true"]')) {
    return null
  }

  const target = eventTarget.closest<HTMLElement>(interactiveSelector)
  if (!target || target.closest('[inert]')) return null
  return target
}

const ensureFeedbackLayer = () => {
  if (feedbackLayer?.isConnected) return feedbackLayer
  feedbackLayer = document.createElement('div')
  feedbackLayer.className = 'tap-feedback-layer'
  feedbackLayer.setAttribute('aria-hidden', 'true')
  document.body.appendChild(feedbackLayer)
  return feedbackLayer
}

const createBurst = (x: number, y: number) => {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

  const burst = document.createElement('span')
  burst.className = 'tap-feedback-burst'
  burst.style.setProperty('--tap-x', `${x}px`)
  burst.style.setProperty('--tap-y', `${y}px`)

  for (let index = 0; index < 7; index += 1) {
    const particle = document.createElement('i')
    particle.style.setProperty('--tap-angle', `${index * (360 / 7) - 90}deg`)
    particle.style.setProperty('--tap-offset', `${-(17 + (index % 3) * 4)}px`)
    particle.style.setProperty('--tap-delay', `${index * 7}ms`)
    burst.appendChild(particle)
  }

  ensureFeedbackLayer().appendChild(burst)
  window.setTimeout(() => burst.remove(), 520)
}

const pressTarget = (target: HTMLElement) => {
  const timer = releaseTimers.get(target)
  if (timer) window.clearTimeout(timer)
  target.classList.remove('tap-feedback-release')
  target.classList.add('tap-feedback-active')
}

const releaseTarget = (target: HTMLElement) => {
  target.classList.remove('tap-feedback-active')
  target.classList.add('tap-feedback-release')
  const timer = window.setTimeout(() => {
    target.classList.remove('tap-feedback-release')
    releaseTimers.delete(target)
  }, 220)
  releaseTimers.set(target, timer)
}

const releasePointer = (pointerId: number) => {
  const active = activePointers.get(pointerId)
  if (!active) return
  releaseTarget(active.target)
  activePointers.delete(pointerId)
}

const releaseAllPointers = () => {
  for (const pointerId of activePointers.keys()) releasePointer(pointerId)
}

const sendNativeHaptic = () => {
  const now = performance.now()
  if (now - lastHapticAt < 45) return

  const bridge = (window as Window & { CasinoAndroid?: NativeMessageBridge }).CasinoAndroid
  if (typeof bridge?.postMessage !== 'function') return

  lastHapticAt = now
  bridge.postMessage(JSON.stringify({ type: 'HAPTIC_TAP' }))
}

export const installTapFeedback = () => {
  if (installed || typeof document === 'undefined') return
  installed = true

  document.addEventListener('pointerdown', (event) => {
    if (!event.isPrimary || event.button !== 0) return
    const target = getInteractiveTarget(event.target)
    if (!target) return

    activePointers.set(event.pointerId, {
      target,
      startX: event.clientX,
      startY: event.clientY,
    })
    pressTarget(target)
    createBurst(event.clientX, event.clientY)
  }, { passive: true })

  document.addEventListener('pointermove', (event) => {
    const active = activePointers.get(event.pointerId)
    if (!active) return
    if (Math.hypot(event.clientX - active.startX, event.clientY - active.startY) > 10) {
      releasePointer(event.pointerId)
    }
  }, { passive: true })

  document.addEventListener('pointerup', (event) => releasePointer(event.pointerId), { passive: true })
  document.addEventListener('pointercancel', (event) => releasePointer(event.pointerId), { passive: true })
  document.addEventListener('click', (event) => {
    if (getInteractiveTarget(event.target)) sendNativeHaptic()
  }, { passive: true })
  window.addEventListener('blur', releaseAllPointers)
  window.addEventListener('scroll', releaseAllPointers, { passive: true, capture: true })
}
