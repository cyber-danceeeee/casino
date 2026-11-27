<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { footerBrandImages } from '../config/footerBrands'
const root = ref<HTMLElement | null>(null)
const visible = ref(false)
let observer: IntersectionObserver | undefined
onMounted(() => {
  if (!root.value || typeof IntersectionObserver === 'undefined') { visible.value = true; return }
  observer = new IntersectionObserver(([entry]) => { visible.value = Boolean(entry?.isIntersecting) })
  observer.observe(root.value)
})
onBeforeUnmount(() => observer?.disconnect())
</script>
<template>
  <section ref="root" class="brand-marquee" aria-label="Game brands">
    <div class="brand-marquee__track" :class="{ 'is-paused': !visible }">
      <div v-for="copy in 2" :key="copy" class="brand-marquee__group" :aria-hidden="copy === 2">
        <div v-for="(source, index) in footerBrandImages" :key="index" class="brand-marquee__card"><img :src="source" alt="" loading="lazy" width="112" height="44" /></div>
      </div>
    </div>
  </section>
</template>
<style scoped>
.brand-marquee {overflow:hidden;min-width:0;width:100%;margin:16px 0 0;}
.brand-marquee__track {display:flex;width:max-content;animation:footer-brand-scroll 41s linear infinite;}
.brand-marquee__track.is-paused {animation-play-state:paused;}
.brand-marquee__group {display:flex;flex-shrink:0;gap:8px;padding-right:8px;}
.brand-marquee__card {flex:0 0 112px;width:112px;height:44px;border-radius:3px;background:#141a2c;}
.brand-marquee__card img {display:block;width:100%;height:100%;object-fit:contain;user-select:none;}
@keyframes footer-brand-scroll {to{transform:translateX(-50%);}}
@media(hover:hover){.brand-marquee:hover .brand-marquee__track{animation-play-state:paused;}}
@media(prefers-reduced-motion:reduce){.brand-marquee{overflow-x:auto;}.brand-marquee__track{animation:none;}.brand-marquee__group[aria-hidden="true"]{display:none;}}
</style>
