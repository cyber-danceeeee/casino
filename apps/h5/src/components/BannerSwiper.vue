<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { BannerItem } from '../mock/games'

const props = defineProps<{
  banners: BannerItem[]
}>()

let timer: number | undefined
let touchStartX = 0
let transitionResetTimer: number | undefined
const activeIndex = ref(0)
const viewportRef = ref<HTMLElement | null>(null)
const slideStep = ref(343)
const slideWidth = ref(331)
const sidePadding = ref(22)
const realIndex = ref(1)
const isAnimating = ref(true)
const slideGap = 12
const maxSlideWidth = 374

const loopBanners = computed(() => {
  if (props.banners.length <= 1) return props.banners

  return [
    props.banners[props.banners.length - 1],
    ...props.banners,
    props.banners[0],
  ]
})

const trackStyle = computed(() => ({
  transform: `translate3d(-${realIndex.value * slideStep.value}px, 0, 0)`,
  transition: isAnimating.value ? 'transform 540ms cubic-bezier(0.16, 1, 0.3, 1)' : 'none',
  '--slide-width': `${slideWidth.value}px`,
  '--slide-padding': `${sidePadding.value}px`,
  '--slide-gap': `${slideGap}px`,
}))

const measure = () => {
  const viewport = viewportRef.value
  if (!viewport) return
  const width = viewport.clientWidth
  const minEdgeGap = width <= 374 ? 16 : 22
  slideWidth.value = Math.min(maxSlideWidth, Math.max(0, width - minEdgeGap * 2))
  sidePadding.value = Math.max(minEdgeGap, Math.round((width - slideWidth.value) / 2))
  slideStep.value = slideWidth.value + slideGap
}

const goTo = (index: number) => {
  if (!props.banners.length) return
  const next = index + 1
  realIndex.value = next
  activeIndex.value = (index + props.banners.length) % props.banners.length

  if (transitionResetTimer) window.clearTimeout(transitionResetTimer)
  transitionResetTimer = window.setTimeout(() => {
    if (realIndex.value === 0) {
      isAnimating.value = false
      realIndex.value = props.banners.length
    } else if (realIndex.value === props.banners.length + 1) {
      isAnimating.value = false
      realIndex.value = 1
    }

    window.requestAnimationFrame(() => {
      isAnimating.value = true
    })
  }, 560)
}

const start = () => {
  stop()
  timer = window.setInterval(() => goTo(activeIndex.value + 1), 3300)
}

const stop = () => {
  if (timer) window.clearInterval(timer)
  timer = undefined
}

const onTouchStart = (event: TouchEvent) => {
  touchStartX = event.touches[0]?.clientX ?? 0
  stop()
}

const onTouchEnd = (event: TouchEvent) => {
  const endX = event.changedTouches[0]?.clientX ?? touchStartX
  const delta = endX - touchStartX

  if (Math.abs(delta) > 34) {
    goTo(activeIndex.value + (delta < 0 ? 1 : -1))
  }

  start()
}

watch(
  () => props.banners.length,
  async () => {
    activeIndex.value = 0
    realIndex.value = props.banners.length > 1 ? 1 : 0
    await nextTick()
    measure()
    start()
  },
)

onMounted(() => {
  measure()
  window.addEventListener('resize', measure)
  start()
})

onBeforeUnmount(() => {
  stop()
  if (transitionResetTimer) window.clearTimeout(transitionResetTimer)
  window.removeEventListener('resize', measure)
})
</script>

<template>
  <section class="banner-swiper" aria-label="Promotions">
    <div
      ref="viewportRef"
      class="banner-swiper__viewport"
      @touchstart.passive="onTouchStart"
      @touchend="onTouchEnd"
    >
      <div class="banner-swiper__track" :style="trackStyle">
        <article
          v-for="(banner, index) in loopBanners"
          :key="`${banner.id}-${index}`"
          class="banner-slide"
        >
          <img
            :src="banner.image"
            :alt="banner.title"
            :loading="index <= 1 ? 'eager' : 'lazy'"
            decoding="async"
          />
        </article>
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.banner-swiper {
  margin: 10px 0 0;
  overflow: hidden;
}

.banner-swiper__viewport {
  width: 100%;
  overflow: hidden;
}

.banner-swiper__track {
  display: flex;
  gap: var(--slide-gap, 12px);
  padding: 0 var(--slide-padding, 22px);
  will-change: transform;
}

.banner-slide {
  flex: 0 0 var(--slide-width, min(374px, calc(100% - 44px)));
  width: var(--slide-width, min(374px, calc(100% - 44px)));
  max-width: var(--slide-width, 374px);
  aspect-ratio: 347 / 140;
  overflow: hidden;
  border-radius: 14px;
  background: #10192d;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

@media (min-width: 431px) {
  .banner-slide {
    flex-basis: var(--slide-width, 374px);
  }
}

@media (max-width: 374px) {
  .banner-slide {
    flex-basis: var(--slide-width, calc(100% - 32px));
  }
}
</style>
