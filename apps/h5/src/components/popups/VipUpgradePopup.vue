<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import gsap from 'gsap'
import type { BuiltinPopupEmits, BuiltinPopupProps } from './types'
import { runtimeConfig } from '../../services/runtimeConfig'
import halo from '../../assets/images/dialogs/dialog-halo.webp'
import confetti from '../../assets/images/dialogs/dialog-confetti.webp'
import vipBadge from '../../assets/images/vip/vip15.webp'

const props = defineProps<BuiltinPopupProps>()
const emit = defineEmits<BuiltinPopupEmits>()
const root = ref<HTMLElement | null>(null)
let context: gsap.Context | null = null
const zh = computed(() => (runtimeConfig.value.language.current || runtimeConfig.value.language.default).toLowerCase().startsWith('zh'))
const copy = computed(() => zh.value ? { kicker: 'VIP 专属', title: '等级权益升级', body: '查看新的 VIP 权益、返水与提现额度。', action: '查看 VIP', close: '关闭' } : { kicker: 'VIP EXCLUSIVE', title: 'Benefits upgraded', body: 'Explore your new VIP benefits, cashback, and limits.', action: 'View VIP', close: 'Close' })
const actionLabel = computed(() => ['NONE', 'CLOSE_ONLY'].includes(props.campaign.action.type) ? copy.value.close : copy.value.action)
onMounted(() => { if (!root.value) return; context = gsap.context(() => { gsap.timeline().from('.vip-panel', { scale: .7, opacity: 0, duration: .55, ease: 'back.out(1.7)' }).from('.vip-badge', { scale: 1.6, opacity: 0, duration: .55, ease: 'power3.out' }, '-=.25').from('.vip-copy > *', { y: 16, opacity: 0, stagger: .07, duration: .25 }, '-=.25'); gsap.to('.vip-halo', { rotate: 360, duration: 22, ease: 'none', repeat: -1 }); gsap.to('.vip-badge', { filter: 'drop-shadow(0 0 28px rgba(255,217,114,.92))', duration: 1.2, yoyo: true, repeat: -1 }) }, root.value) })
onBeforeUnmount(() => context?.revert())
</script>
<template><div ref="root" class="builtin-popup-layer vip-popup" @click.self="emit('close')"><section class="vip-panel" role="dialog" aria-modal="true" :aria-label="copy.title"><button class="builtin-popup-close" type="button" :aria-label="copy.close" @click="emit('close')">&times;</button><img class="vip-halo" :src="halo" alt="" /><img class="vip-confetti" :src="confetti" alt="" /><img class="vip-badge" :src="vipBadge" alt="" /><div class="vip-copy"><b>{{ copy.kicker }}</b><h2>{{ copy.title }}</h2><p>{{ copy.body }}</p><button class="builtin-popup-cta" type="button" @click="emit('action')">{{ actionLabel }}</button></div></section></div></template>
<style scoped lang="scss">
@use './popup-shared';
.vip-panel{position:relative;width:min(385px,100%);min-height:500px;overflow:hidden;border:2px solid #f1cf75;border-radius:18px;background:#17151c;box-shadow:0 30px 80px rgb(0 0 0 / 72%),inset 0 -90px 100px rgb(106 52 132 / 24%)}.vip-halo{position:absolute;top:-50px;left:calc(50% - 245px);width:490px;opacity:.65}.vip-confetti{position:absolute;top:0;left:0;width:100%;opacity:.68}.vip-badge{position:absolute;top:42px;left:calc(50% - 130px);width:260px;filter:drop-shadow(0 0 18px rgb(255 217 114 / 66%))}.vip-copy{position:absolute;right:24px;bottom:28px;left:24px;text-align:center;color:#fff}.vip-copy b{color:#f6d578;font-size:11px;letter-spacing:0}.vip-copy h2{margin:7px 0;font-family:Georgia,serif;font-size:29px;letter-spacing:0}.vip-copy p{min-height:40px;margin:0 0 18px;color:#e8dfed;font-size:14px;line-height:1.45}@media(max-height:620px){.vip-panel{scale:.9}}
</style>
