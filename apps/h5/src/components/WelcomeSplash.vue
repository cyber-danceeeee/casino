<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import LottiePlayer from './LottiePlayer.vue'
import { runtimeConfig, toRuntimeAssetUrl } from '../services/runtimeConfig'

const props = withDefaults(
  defineProps<{
    progress: number
    ready: boolean
    minDuration?: number
  }>(),
  {
    minDuration: 1500,
  },
)

const emit = defineEmits<{
  complete: []
}>()

let frameId = 0
let startTime = 0
let completeTimer: number | undefined
let hasCompleted = false
const visibleProgress = ref(0)
const isLeaving = ref(false)

const progressStyle = computed(() => ({
  width: `${visibleProgress.value}%`,
}))
const logo = computed(() =>
  toRuntimeAssetUrl(runtimeConfig.value.branding.loadingLogoUrl || runtimeConfig.value.branding.logoUrl),
)
const loadingAnimation = computed(() => toRuntimeAssetUrl(runtimeConfig.value.branding.loadingAnimationUrl))
const splashStyle = computed(() => ({
  backgroundColor: runtimeConfig.value.branding.theme?.backgroundColor || '#003f77',
  backgroundImage: runtimeConfig.value.branding.loadingBackgroundUrl
    ? `url("${toRuntimeAssetUrl(runtimeConfig.value.branding.loadingBackgroundUrl)}")`
    : '',
}))

const ariaProgress = computed(() => Math.round(visibleProgress.value))

const clampProgress = (value: number) => Math.max(0, Math.min(100, value))

const complete = () => {
  if (hasCompleted) return

  hasCompleted = true
  visibleProgress.value = 100
  completeTimer = window.setTimeout(() => {
    isLeaving.value = true
    window.setTimeout(() => emit('complete'), 280)
  }, 420)
}

const tick = (time: number) => {
  if (!startTime) startTime = time

  const elapsed = time - startTime
  const idleProgress = Math.min(92, 8 + elapsed / 86)
  const realProgress = clampProgress(props.progress)
  const targetProgress = props.ready ? 100 : Math.min(96, Math.max(realProgress, idleProgress))
  const delta = targetProgress - visibleProgress.value

  if (delta > 0) {
    visibleProgress.value = Math.min(targetProgress, visibleProgress.value + Math.max(0.45, delta * 0.14))
  }

  if (props.ready && elapsed >= props.minDuration && visibleProgress.value >= 99.5) {
    complete()
    return
  }

  frameId = window.requestAnimationFrame(tick)
}

onMounted(() => {
  frameId = window.requestAnimationFrame(tick)
})

onBeforeUnmount(() => {
  if (frameId) window.cancelAnimationFrame(frameId)
  if (completeTimer) window.clearTimeout(completeTimer)
})
</script>

<template>
  <section class="welcome-splash" :class="{ 'is-leaving': isLeaving }" :style="splashStyle" aria-label="Loading">
    <div v-if="loadingAnimation" class="welcome-splash__lottie">
      <LottiePlayer :src="loadingAnimation" :loop="true" :autoplay="true" />
    </div>
    <img v-else class="welcome-splash__logo" :src="logo" alt="Brand" />

    <div
      class="welcome-splash__progress"
      role="progressbar"
      :aria-valuenow="ariaProgress"
      aria-valuemin="0"
      aria-valuemax="100"
    >
      <span class="welcome-splash__bar" :style="progressStyle" />
    </div>

    <p class="welcome-splash__text">
      loading<span class="welcome-splash__dot">.</span><span class="welcome-splash__dot">.</span><span class="welcome-splash__dot">.</span>
    </p>
  </section>
</template>

<style scoped lang="scss">
.welcome-splash {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: grid;
  justify-items: center;
  min-height: 100dvh;
  overflow: hidden;
  background: #003f77;
  opacity: 1;
  transition: opacity 280ms ease;

  &.is-leaving {
    opacity: 0;
  }
}

.welcome-splash__logo {
  position: absolute;
  top: 39.2%;
  left: 50%;
  width: min(240px, 62vw);
  height: auto;
  object-fit: contain;
  transform: translate(-50%, -50%);
  filter: drop-shadow(0 5px 0 rgba(0, 38, 78, 0.75)) drop-shadow(0 10px 5px rgba(0, 0, 0, 0.32));
  animation: splash-logo-in 620ms cubic-bezier(0.16, 1, 0.3, 1) both;
}

.welcome-splash__lottie {
  position: absolute;
  top: 32%;
  left: 50%;
  width: min(240px, 62vw);
  height: min(180px, 48vw);
  transform: translate(-50%, -50%);
}

.welcome-splash__progress {
  position: absolute;
  top: 76.3%;
  left: 50%;
  width: min(318px, calc(100vw - 64px));
  height: 25px;
  padding: 0;
  overflow: hidden;
  border: 4px solid #ffffff;
  border-radius: 999px;
  background:
    repeating-linear-gradient(
      120deg,
      rgba(255, 255, 255, 0.2) 0,
      rgba(255, 255, 255, 0.2) 13px,
      rgba(120, 130, 142, 0.3) 13px,
      rgba(120, 130, 142, 0.3) 26px
    ),
    linear-gradient(180deg, #d7e0e2, #b9c3c5);
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.42) inset,
    0 4px 9px rgba(0, 18, 42, 0.28);
  transform: translate(-50%, -50%);
}

.welcome-splash__bar {
  display: block;
  height: 100%;
  min-width: 0;
  border-radius: inherit;
  background:
    linear-gradient(180deg, rgba(141, 211, 255, 0.95) 0%, rgba(16, 142, 255, 0.95) 45%, #0787ef 100%),
    #0888ef;
  box-shadow:
    0 3px 5px rgba(0, 35, 95, 0.3),
    0 1px 0 rgba(255, 255, 255, 0.45) inset;
  transition: width 90ms linear;
}

.welcome-splash__text {
  position: absolute;
  top: 82.1%;
  left: 50%;
  margin: 0;
  color: #ffffff;
  font-size: clamp(26px, 7.8vw, 38px);
  font-weight: 400;
  line-height: 1;
  transform: translateX(-50%);
}

.welcome-splash__dot {
  display: inline-block;
  animation: splash-dot 1.2s ease-in-out infinite both;

  &:nth-child(2) {
    animation-delay: 160ms;
  }

  &:nth-child(3) {
    animation-delay: 320ms;
  }
}

@keyframes splash-logo-in {
  from {
    opacity: 0;
    transform: translate(-50%, -45%) skewX(-7deg) scale(0.92);
  }

  to {
    opacity: 1;
    transform: translate(-50%, -50%) skewX(-7deg) scale(1);
  }
}

@keyframes splash-dot {
  0%,
  100% {
    opacity: 0.35;
    transform: translateY(0);
  }

  50% {
    opacity: 1;
    transform: translateY(-3px);
  }
}

@media (max-width: 374px) {
  .welcome-splash__progress {
    width: calc(100vw - 52px);
    height: 23px;
  }
}
</style>
