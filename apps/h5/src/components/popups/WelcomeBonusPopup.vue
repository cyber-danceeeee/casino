<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import gsap from 'gsap'
import type { BuiltinPopupEmits, BuiltinPopupProps } from './types'
import { runtimeConfig } from '../../services/runtimeConfig'
import halo from '../../assets/images/dialogs/dialog-halo.webp'
import confetti from '../../assets/images/dialogs/dialog-confetti.webp'
import gift from '../../assets/images/dialogs/dialog-gift.webp'

const props = defineProps<BuiltinPopupProps>()
const emit = defineEmits<BuiltinPopupEmits>()
const root = ref<HTMLElement | null>(null)
let context: gsap.Context | null = null
const zh = computed(() => (runtimeConfig.value.language.current || runtimeConfig.value.language.default).toLowerCase().startsWith('zh'))
const copy = computed(() => zh.value ? { badge: 'WELCOME', title: '欢迎加入', body: '你的新人专属礼遇已经准备好了。', action: '查看奖励', close: '关闭' } : { badge: 'WELCOME', title: 'A gift for you', body: 'Your new-player welcome reward is ready.', action: 'View reward', close: 'Close' })
const actionLabel = computed(() => ['NONE', 'CLOSE_ONLY'].includes(props.campaign.action.type) ? copy.value.close : copy.value.action)
onMounted(() => { if (!root.value) return; context = gsap.context(() => { gsap.timeline().from('.welcome-panel', { scale: .55, opacity: 0, rotate: -4, duration: .55, ease: 'back.out(1.8)' }).from('.welcome-gift', { y: -55, opacity: 0, rotate: 8, duration: .45, ease: 'bounce.out' }, '-=.2').from('.welcome-copy > *', { y: 14, opacity: 0, stagger: .07, duration: .25 }, '-=.2'); gsap.to('.welcome-halo', { rotate: 360, duration: 20, ease: 'none', repeat: -1 }); gsap.to('.welcome-gift', { y: -8, duration: 1.5, ease: 'sine.inOut', yoyo: true, repeat: -1 }) }, root.value) })
onBeforeUnmount(() => context?.revert())
</script>
<template><div ref="root" class="builtin-popup-layer welcome-popup" @click.self="emit('close')"><section class="welcome-panel" role="dialog" aria-modal="true" :aria-label="copy.title"><button class="builtin-popup-close" type="button" :aria-label="copy.close" @click="emit('close')">&times;</button><img class="welcome-halo" :src="halo" alt="" /><img class="welcome-confetti" :src="confetti" alt="" /><img class="welcome-gift" :src="gift" alt="" /><div class="welcome-copy"><b>{{ copy.badge }}</b><h2>{{ copy.title }}</h2><p>{{ copy.body }}</p><button class="builtin-popup-cta" type="button" @click="emit('action')">{{ actionLabel }}</button></div></section></div></template>
<style scoped lang="scss">
@use './popup-shared';
.welcome-panel{position:relative;width:min(380px,100%);min-height:510px;overflow:hidden;border:2px solid #d8afff;border-radius:28px;background:#28144f;box-shadow:0 30px 80px rgb(0 0 0 / 65%),inset 0 -80px 100px rgb(224 77 187 / 22%)}.welcome-halo{position:absolute;top:-38px;left:calc(50% - 225px);width:450px;opacity:.76}.welcome-confetti{position:absolute;inset:10px 0 auto;width:100%;opacity:.9}.welcome-gift{position:absolute;top:56px;left:calc(50% - 118px);width:236px;filter:drop-shadow(0 20px 24px rgb(10 3 29 / 56%))}.welcome-copy{position:absolute;right:22px;bottom:28px;left:22px;text-align:center;color:white}.welcome-copy b{display:inline-block;padding:5px 12px;border:1px solid #f5d4ff;border-radius:14px;color:#ffe493;background:#6e2d85;font-size:11px}.welcome-copy h2{margin:10px 0 4px;font-size:30px;letter-spacing:0}.welcome-copy p{min-height:40px;margin:0 0 18px;color:#eedffc;font-size:14px;line-height:1.45}@media(max-height:620px){.welcome-panel{scale:.9}}
</style>
