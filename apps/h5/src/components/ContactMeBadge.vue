<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import defaultCustomerImage from '../assets/images/profile/customer.webp'

const props = withDefaults(
  defineProps<{
    image?: string
    label?: string
    draggable?: boolean
    dockOnIdle?: boolean
  }>(),
  {
    image: defaultCustomerImage,
    label: 'Contact Me',
    draggable: false,
    dockOnIdle: false,
  },
)

const emit = defineEmits<{
  click: []
}>()

const rootRef = ref<HTMLElement | null>(null)
const dragStyle = ref<Record<string, string>>({})
const isDragging = ref(false)
const isDockActive = ref(false)
const isScrollHidden = ref(false)
const dockSide = ref<'left' | 'right'>('right')

let pointerId: number | null = null
let startX = 0
let startY = 0
let startLeft = 0
let startTop = 0
let didDrag = false
let suppressNextClick = false
let scrollIdleTimer: number | undefined

const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max)

const dragBounds = (rect: DOMRect) => {
  const shellWidth = Math.min(window.innerWidth, 430)
  const shellLeft = Math.max((window.innerWidth - shellWidth) / 2, 0)
  const gap = props.dockOnIdle ? 0 : 6

  return {
    left: shellLeft + gap,
    right: shellLeft + shellWidth - rect.width - gap,
    top: gap,
    bottom: window.innerHeight - rect.height - gap,
  }
}

const updateDragPosition = (left: number, top: number, rect: DOMRect) => {
  const bounds = dragBounds(rect)
  dragStyle.value = {
    position: 'fixed',
    left: `${clamp(left, bounds.left, bounds.right)}px`,
    top: `${clamp(top, bounds.top, bounds.bottom)}px`,
    right: 'auto',
    bottom: 'auto',
  }
}

const componentStyle = computed(() => {
  const style = props.draggable ? { ...dragStyle.value } : {}
  const dockOffset = dockSide.value === 'right' ? '34px' : '-34px'
  style['--contact-dock-x'] = dockOffset
  style['--contact-scroll-x'] = dockOffset
  return style
})

const handlePointerDown = (event: PointerEvent) => {
  if (!props.draggable || event.button !== 0) return

  const target = rootRef.value
  if (!target) return

  const rect = target.getBoundingClientRect()
  pointerId = event.pointerId
  startX = event.clientX
  startY = event.clientY
  startLeft = rect.left
  startTop = rect.top
  didDrag = false
  isDragging.value = true
  isScrollHidden.value = false
  isDockActive.value = true
  window.clearTimeout(scrollIdleTimer)
  target.setPointerCapture(event.pointerId)
  window.addEventListener('pointermove', handlePointerMove, { passive: false })
  window.addEventListener('pointerup', finishDrag)
  window.addEventListener('pointercancel', finishDrag)
}

const handlePointerMove = (event: PointerEvent) => {
  if (!props.draggable || !isDragging.value || pointerId !== event.pointerId) return

  const target = rootRef.value
  if (!target) return

  const deltaX = event.clientX - startX
  const deltaY = event.clientY - startY
  if (Math.abs(deltaX) > 4 || Math.abs(deltaY) > 4) {
    didDrag = true
  }

  updateDragPosition(startLeft + deltaX, startTop + deltaY, target.getBoundingClientRect())
  event.preventDefault()
}

const finishDrag = (event: PointerEvent) => {
  if (!props.draggable || pointerId !== event.pointerId) return

  if (rootRef.value?.hasPointerCapture(event.pointerId)) {
    rootRef.value.releasePointerCapture(event.pointerId)
  }
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerup', finishDrag)
  window.removeEventListener('pointercancel', finishDrag)
  const rect = rootRef.value?.getBoundingClientRect()
  if (rect) {
    const shellWidth = Math.min(window.innerWidth, 430)
    const shellLeft = Math.max((window.innerWidth - shellWidth) / 2, 0)
    const nextDockSide = rect.left + rect.width / 2 < shellLeft + shellWidth / 2 ? 'left' : 'right'
    dockSide.value = nextDockSide

    if (props.dockOnIdle) {
      const bounds = dragBounds(rect)
      const nextLeft = nextDockSide === 'left' ? bounds.left : bounds.right
      dragStyle.value = {
        position: 'fixed',
        left: `${nextLeft}px`,
        top: `${clamp(rect.top, bounds.top, bounds.bottom)}px`,
        right: 'auto',
        bottom: 'auto',
      }
    }
  }
  pointerId = null
  isDragging.value = false
  isDockActive.value = false
  suppressNextClick = didDrag
}

const handleClick = () => {
  if (suppressNextClick) {
    suppressNextClick = false
    return
  }

  emit('click')
}

const showDocked = () => {
  if (props.dockOnIdle) {
    isDockActive.value = true
  }
}

const hideDocked = () => {
  if (props.dockOnIdle && !isDragging.value) {
    isDockActive.value = false
  }
}

const revealAfterScroll = () => {
  isScrollHidden.value = false
  if (props.dockOnIdle) {
    isDockActive.value = true
  }
}

const handleScroll = () => {
  if (!props.draggable || isDragging.value) return

  isScrollHidden.value = true
  window.clearTimeout(scrollIdleTimer)
  scrollIdleTimer = window.setTimeout(revealAfterScroll, 520)
}

onMounted(() => {
  if (props.draggable) {
    window.addEventListener('scroll', handleScroll, { passive: true, capture: true })
  }
})

onBeforeUnmount(() => {
  window.clearTimeout(scrollIdleTimer)
  window.removeEventListener('scroll', handleScroll, true)
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerup', finishDrag)
  window.removeEventListener('pointercancel', finishDrag)
})
</script>

<template>
  <button
    ref="rootRef"
    type="button"
    class="contact-me-badge"
    :class="{
      'is-draggable': draggable,
      'is-dragging': isDragging,
      'is-docked': dockOnIdle,
      'is-active': isDockActive || isDragging,
      'is-scroll-hidden': isScrollHidden && !isDragging,
    }"
    :style="componentStyle"
    :draggable="false"
    @click="handleClick"
    @mouseenter="showDocked"
    @mouseleave="hideDocked"
    @focusin="showDocked"
    @focusout="hideDocked"
    @pointerdown="handlePointerDown"
    @pointermove="handlePointerMove"
    @pointerup="finishDrag"
    @pointercancel="finishDrag"
  >
    <img :src="image" alt="" loading="lazy" draggable="false" />
    <span>{{ label }}</span>
  </button>
</template>

<style scoped lang="scss">
.contact-me-badge {
  display: flex;
  width: 60px;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  border: 0;
  padding: 0;
  color: #003278;
  background: transparent;
  cursor: pointer;
  touch-action: none;
  transition: transform 300ms ease;
  will-change: transform;
  -webkit-tap-highlight-color: transparent;

  &.is-draggable {
    cursor: grab;
  }

  &.is-docked {
    transform: translateX(var(--contact-dock-x, 34px));
  }

  &.is-docked.is-active {
    transform: translateX(0);
  }

  &.is-scroll-hidden {
    transform: translateX(var(--contact-scroll-x, 34px));
  }

  &.is-dragging {
    cursor: grabbing;
    transition: none;
  }

  img {
    width: 44px;
    height: 55px;
    object-fit: contain;
    filter: drop-shadow(0 4px 8px rgba(0, 20, 65, 0.18));
    pointer-events: none;
  }

  span {
    display: block;
    max-width: none;
    margin-top: -6px;
    border: 1px solid #00a1d5;
    border-radius: 999px;
    padding: 2px 8px;
    color: #003278;
    background: linear-gradient(#baeaff 0%, #38cfff 100%);
    box-shadow: 0 2px 5px rgba(0, 77, 128, 0.16);
    font-size: 8px;
    font-weight: 600;
    line-height: 1;
    white-space: nowrap;
  }
}
</style>
