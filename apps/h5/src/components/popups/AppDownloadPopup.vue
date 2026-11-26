<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import gsap from 'gsap'
import type { BuiltinPopupEmits, BuiltinPopupProps } from './types'
import { runtimeConfig } from '../../services/runtimeConfig'
import halo from '../../assets/images/dialogs/dialog-halo.webp'
import confetti from '../../assets/images/dialogs/dialog-confetti.webp'
import hero from '../../assets/images/dialogs/dialog-app-hero.webp'

const props = defineProps<BuiltinPopupProps>()
const emit = defineEmits<BuiltinPopupEmits>()
const root = ref<HTMLElement | null>(null)
let context: gsap.Context | null = null
const zh = computed(() => (runtimeConfig.value.language.current || runtimeConfig.value.language.default).toLowerCase().startsWith('zh'))
const copy = computed(() => zh.value ? { kicker: '官方应用', title: '安装 App，体验更流畅', body: '更快启动、更稳定游戏，并及时接收奖励通知。', action: '立即安装', close: '关闭' } : { kicker: 'OFFICIAL APP', title: 'A faster way to play', body: 'Launch faster, play smoothly, and receive reward alerts.', action: 'Install app', close: 'Close' })
const actionLabel = computed(() => ['NONE', 'CLOSE_ONLY'].includes(props.campaign.action.type) ? copy.value.close : copy.value.action)
onMounted(() => { if (!root.value) return; context = gsap.context(() => { gsap.timeline().from('.app-panel', { y: 70, opacity: 0, duration: .5, ease: 'power3.out' }).from('.app-hero', { x: -60, rotate: -8, opacity: 0, duration: .55, ease: 'back.out(1.5)' }, '-=.25').from('.app-copy > *', { x: 18, opacity: 0, stagger: .07, duration: .25 }, '-=.3'); gsap.to('.app-halo', { rotate: 360, duration: 20, ease: 'none', repeat: -1 }); gsap.to('.app-hero', { y: -7, duration: 1.4, ease: 'sine.inOut', yoyo: true, repeat: -1 }) }, root.value) })
onBeforeUnmount(() => context?.revert())
</script>
<template><div ref="root" class="builtin-popup-layer app-popup" @click.self="emit('close')"><section class="app-panel" role="dialog" aria-modal="true" :aria-label="copy.title"><button class="builtin-popup-close" type="button" :aria-label="copy.close" @click="emit('close')">&times;</button><img class="app-halo" :src="halo" alt="" /><img class="app-confetti" :src="confetti" alt="" /><img class="app-hero" :src="hero" alt="" /><div class="app-copy"><b>{{ copy.kicker }}</b><h2>{{ copy.title }}</h2><p>{{ copy.body }}</p><button class="builtin-popup-cta" type="button" @click="emit('action')">{{ actionLabel }}</button></div></section></div></template>
<style scoped lang="scss">
@use './popup-shared';
.app-panel{position:relative;width:min(410px,100%);min-height:455px;overflow:hidden;border:2px solid #84d8ff;border-radius:20px;background:#102744;box-shadow:0 30px 80px rgb(0 0 0 / 68%),inset 0 0 80px rgb(63 197 255 / 18%)}.app-halo{position:absolute;top:-95px;left:calc(50% - 205px);width:410px;opacity:.55}.app-confetti{position:absolute;inset:0;width:100%;opacity:.6}.app-hero{position:absolute;top:42px;left:calc(50% - 103px);width:206px;filter:drop-shadow(0 18px 24px rgb(0 0 0 / 48%))}.app-copy{position:absolute;top:244px;right:28px;left:28px;text-align:center;color:#fff}.app-copy b{color:#80d9ff;font-size:11px;letter-spacing:0}.app-copy h2{margin:7px 0;font-size:27px;line-height:1.08;letter-spacing:0}.app-copy p{margin:0 auto 18px;max-width:300px;min-height:40px;color:#d8edfb;font-size:13px;line-height:1.45}.app-copy :deep(.builtin-popup-cta){min-width:190px}@media(max-height:620px){.app-panel{scale:.92}}
</style>
