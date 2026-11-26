<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import loadingAnimation from '../assets/lottie/lottie-loading.json'
import LottiePlayer from './LottiePlayer.vue'
import type { GameItem } from '../mock/games'

const props = defineProps<{
  visible: boolean
  game: GameItem
}>()

const emit = defineEmits<{
  close: []
  complete: []
}>()

let frame = 0
let startTime = 0
const progress = ref(0)

const easeOutCubic = (value: number) => 1 - Math.pow(1 - value, 3)

const cancel = () => {
  if (frame) window.cancelAnimationFrame(frame)
  frame = 0
}

const run = (time: number) => {
  if (!startTime) startTime = time

  const elapsed = time - startTime
  const ratio = Math.min(elapsed / 2400, 1)
  progress.value = Math.min(100, Math.round(easeOutCubic(ratio) * 100))

  if (ratio < 1) {
    frame = window.requestAnimationFrame(run)
    return
  }

  frame = 0
  emit('complete')
}

const start = () => {
  cancel()
  progress.value = 0
  startTime = 0
  frame = window.requestAnimationFrame(run)
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) start()
    else cancel()
  },
  { immediate: true },
)

onBeforeUnmount(cancel)
</script>

<template>
  <Transition name="loading-fade">
    <div v-if="visible" class="game-loading" role="dialog" aria-modal="true">
      <div class="game-loading__stage">
        <LottiePlayer :animation-data="loadingAnimation" :loop="true" :autoplay="visible" />
      </div>
    </div>
  </Transition>
</template>

<style scoped lang="scss">
.game-loading {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  width: 100%;
  max-width: 430px;
  min-height: 100vh;
  margin: 0 auto;
  padding: 0;
  overflow: hidden;
  background: #003f77;
  inset-inline: 0;
}

.game-loading__stage {
  width: min(100%, 375px);
  aspect-ratio: 375 / 158;
  max-width: 100%;
}

.loading-fade-enter-active,
.loading-fade-leave-active {
  transition: opacity 220ms ease;
}

.loading-fade-enter-from,
.loading-fade-leave-to {
  opacity: 0;
}

</style>
