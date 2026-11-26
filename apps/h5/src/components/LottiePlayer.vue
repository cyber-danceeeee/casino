<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import lottie from 'lottie-web/build/player/lottie_light'
import type { AnimationItem } from 'lottie-web'

const props = withDefaults(
  defineProps<{
    animationData?: Record<string, unknown>
    src?: string
    loop?: boolean | number
    autoplay?: boolean
    renderer?: 'svg' | 'canvas' | 'html'
    preserveAspectRatio?: string
    speed?: number
  }>(),
  {
    loop: true,
    autoplay: true,
    renderer: 'svg',
    preserveAspectRatio: 'xMidYMid meet',
    speed: 1,
  },
)

const container = ref<HTMLDivElement | null>(null)
const remoteAnimationData = shallowRef<Record<string, unknown> | null>(null)
let animation: AnimationItem | null = null
let sourceSequence = 0

const destroy = () => {
  animation?.destroy()
  animation = null
}

const mountAnimation = () => {
  if (!container.value) return
  const data = remoteAnimationData.value || props.animationData
  if (!data) return

  destroy()
  animation = lottie.loadAnimation({
    container: container.value,
    renderer: props.renderer,
    loop: props.loop,
    autoplay: props.autoplay,
    animationData: data,
    rendererSettings: {
      preserveAspectRatio: props.preserveAspectRatio,
      progressiveLoad: true,
    },
  })
  animation.setSpeed(props.speed)
}

watch(
  () => [props.animationData, remoteAnimationData.value, props.renderer, props.preserveAspectRatio] as const,
  mountAnimation,
)

async function loadRemoteSource(src?: string) {
  const sequence = ++sourceSequence
  remoteAnimationData.value = null
  if (!src) {
    mountAnimation()
    return
  }
  try {
    const response = await fetch(src)
    const data = (await response.json()) as Record<string, unknown>
    if (sequence === sourceSequence) remoteAnimationData.value = data
  } catch (error) {
    console.warn('[lottie] failed to load animation.', error)
    if (sequence === sourceSequence) mountAnimation()
  }
}

watch(
  () => props.src,
  (src) => {
    void loadRemoteSource(src)
  },
  { immediate: true },
)

onMounted(() => {
  if (!props.src) mountAnimation()
})

watch(
  () => props.loop,
  () => {
    if (animation) animation.loop = props.loop
  },
)

watch(
  () => props.autoplay,
  (autoplay) => {
    if (!animation) return
    if (autoplay) animation.play()
    else animation.pause()
  },
)

watch(
  () => props.speed,
  (speed) => {
    animation?.setSpeed(speed)
  },
)

onBeforeUnmount(destroy)
</script>

<template>
  <div ref="container" class="lottie-player" aria-hidden="true" />
</template>

<style scoped>
.lottie-player {
  width: 100%;
  height: 100%;
}

.lottie-player :deep(svg),
.lottie-player :deep(canvas) {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
