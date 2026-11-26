<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import gsap from 'gsap'
import type { BuiltinPopupEmits, BuiltinPopupProps } from './types'
import { runtimeConfig } from '../../services/runtimeConfig'
import rays from '../../assets/images/dialogs/bonus__rays.webp'
import chests from '../../assets/images/dialogs/top-chests.webp'
import gift from '../../assets/images/dialogs/mini-gift.webp'
import coins from '../../assets/images/dialogs/offers-invite-chest-gold-2.webp'

const props = defineProps<BuiltinPopupProps>()
const emit = defineEmits<BuiltinPopupEmits>()
const root = ref<HTMLElement | null>(null)
let context: gsap.Context | null = null
const zh = computed(() => (runtimeConfig.value.language.current || runtimeConfig.value.language.default).toLowerCase().startsWith('zh'))
const copy = computed(() => zh.value ? { kicker: '每日回馈', title: '返水奖励已更新', body: '查看今日可领取的返水与回馈活动。', action: '查看返水', close: '关闭' } : { kicker: 'DAILY RETURN', title: 'Cashback is ready', body: 'See today’s available cashback and return offers.', action: 'View cashback', close: 'Close' })
const actionLabel = computed(() => ['NONE', 'CLOSE_ONLY'].includes(props.campaign.action.type) ? copy.value.close : copy.value.action)
onMounted(() => { if (!root.value) return; context = gsap.context(() => { gsap.timeline().from('.cashback-ticket', { x: 70, opacity: 0, rotate: 6, duration: .5, ease: 'back.out(1.5)' }).from('.cashback-chests', { y: -35, opacity: 0, duration: .45, ease: 'bounce.out' }, '-=.25').from('.cashback-copy > *', { x: -18, opacity: 0, stagger: .07, duration: .25 }, '-=.25'); gsap.to('.cashback-rays', { rotate: 360, duration: 15, ease: 'none', repeat: -1 }); gsap.to('.cashback-gift', { rotate: -7, duration: .8, ease: 'sine.inOut', yoyo: true, repeat: -1 }) }, root.value) })
onBeforeUnmount(() => context?.revert())
</script>
<template><div ref="root" class="builtin-popup-layer cashback-popup" @click.self="emit('close')"><section class="cashback-ticket" role="dialog" aria-modal="true" :aria-label="copy.title"><button class="builtin-popup-close" type="button" :aria-label="copy.close" @click="emit('close')">&times;</button><div class="cashback-visual"><img class="cashback-rays" :src="rays" alt="" /><img class="cashback-chests" :src="chests" alt="" /><img class="cashback-coins" :src="coins" alt="" /><img class="cashback-gift" :src="gift" alt="" /></div><div class="cashback-copy"><b>{{ copy.kicker }}</b><h2>{{ copy.title }}</h2><p>{{ copy.body }}</p><button class="builtin-popup-cta" type="button" @click="emit('action')">{{ actionLabel }}</button></div></section></div></template>
<style scoped lang="scss">
@use './popup-shared';
.cashback-ticket{position:relative;width:min(405px,100%);min-height:470px;overflow:hidden;border:2px solid #f8ce63;border-radius:16px;background:#123b43;box-shadow:0 28px 75px rgb(0 0 0 / 66%),inset 0 0 70px rgb(62 221 188 / 16%)}.cashback-ticket::before,.cashback-ticket::after{position:absolute;top:286px;width:30px;height:30px;border-radius:50%;background:#03050a;content:''}.cashback-ticket::before{left:-16px}.cashback-ticket::after{right:-16px}.cashback-visual{position:relative;height:292px;border-bottom:1px dashed rgb(255 229 147 / 65%)}.cashback-rays{position:absolute;top:-98px;left:calc(50% - 260px);width:520px;opacity:.45}.cashback-chests{position:absolute;top:42px;left:12%;width:76%;filter:drop-shadow(0 15px 22px rgb(0 0 0 / 55%))}.cashback-coins{position:absolute;right:10px;bottom:-8px;width:110px}.cashback-gift{position:absolute;left:22px;bottom:18px;width:70px}.cashback-copy{padding:18px 28px 28px;text-align:center;color:#fff}.cashback-copy b{color:#86f0d8;font-size:11px;letter-spacing:0}.cashback-copy h2{margin:6px 0;font-size:27px;letter-spacing:0}.cashback-copy p{margin:0 0 17px;color:#d5eee8;font-size:14px;line-height:1.4}@media(max-height:620px){.cashback-ticket{scale:.92}}
</style>
