<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import gsap from 'gsap'
import type { BuiltinPopupEmits, BuiltinPopupProps } from './types'
import { runtimeConfig } from '../../services/runtimeConfig'
import glow from '../../assets/images/dialogs/activity__guide-flow-dialogs__bg_glow.webp'
import stars from '../../assets/images/dialogs/activity__guide-flow-dialogs__first_deposit_bg_stars.webp'
import bowl from '../../assets/images/dialogs/activity__guide-flow-dialogs__first_deposit_bowl.webp'
import treasure from '../../assets/images/dialogs/activity__guide-flow-dialogs__first_deposit_guide_treasure.webp'
import hand from '../../assets/images/dialogs/activity__common__tap_hand.webp'

const props = defineProps<BuiltinPopupProps>()
const emit = defineEmits<BuiltinPopupEmits>()
const root = ref<HTMLElement | null>(null)
let context: gsap.Context | null = null
const zh = computed(() => (runtimeConfig.value.language.current || runtimeConfig.value.language.default).toLowerCase().startsWith('zh'))
const copy = computed(() => zh.value ? { eyebrow: '首次充值礼遇', title: '首充奖励已开启', body: '完成首次充值，立即解锁专属奖励。', action: '立即充值', close: '关闭' } : { eyebrow: 'FIRST DEPOSIT', title: 'Your first reward awaits', body: 'Make your first deposit and unlock the welcome reward.', action: 'Deposit now', close: 'Close' })
const actionLabel = computed(() => ['NONE', 'CLOSE_ONLY'].includes(props.campaign.action.type) ? copy.value.close : copy.value.action)

onMounted(() => {
  if (!root.value) return
  context = gsap.context(() => {
    gsap.timeline().from('.deposit-stage', { y: 50, scale: .76, opacity: 0, duration: .52, ease: 'back.out(1.7)' }).from('.deposit-title, .deposit-copy, .builtin-popup-cta', { y: 16, opacity: 0, duration: .28, stagger: .08 }, '-=.15')
    gsap.to('.deposit-glow', { rotate: 360, duration: 18, ease: 'none', repeat: -1 })
    gsap.to('.deposit-treasure', { y: -9, duration: 1.6, ease: 'sine.inOut', yoyo: true, repeat: -1 })
    gsap.to('.deposit-hand', { x: -4, y: -5, scale: .92, duration: .7, ease: 'sine.inOut', yoyo: true, repeat: -1 })
  }, root.value)
})
onBeforeUnmount(() => context?.revert())
</script>

<template>
  <div ref="root" class="builtin-popup-layer deposit-popup" @click.self="emit('close')">
    <section class="deposit-stage" role="dialog" aria-modal="true" :aria-label="copy.title">
      <button class="builtin-popup-close" type="button" :aria-label="copy.close" @click="emit('close')">&times;</button>
      <img class="deposit-glow" :src="glow" alt="" aria-hidden="true" />
      <img class="deposit-stars" :src="stars" alt="" aria-hidden="true" />
      <div class="deposit-visual"><img class="deposit-bowl" :src="bowl" alt="" /><img class="deposit-treasure" :src="treasure" alt="" /><img class="deposit-hand" :src="hand" alt="" /></div>
      <div class="deposit-content"><p>{{ copy.eyebrow }}</p><h2 class="deposit-title">{{ copy.title }}</h2><span class="deposit-copy">{{ copy.body }}</span><button class="builtin-popup-cta" type="button" @click="emit('action')">{{ actionLabel }}</button></div>
    </section>
  </div>
</template>

<style scoped lang="scss">
@use './popup-shared';
.deposit-stage{position:relative;width:min(390px,100%);min-height:530px;overflow:hidden;border:2px solid #ffda73;border-radius:22px;background:#581916;box-shadow:0 24px 70px rgb(0 0 0 / 60%),inset 0 0 50px rgb(255 213 86 / 24%)}
.deposit-glow{position:absolute;top:-45px;left:calc(50% - 250px);width:500px;opacity:.84}.deposit-stars{position:absolute;inset:0;width:100%;height:100%;object-fit:cover;opacity:.68}.deposit-visual{position:relative;height:308px}.deposit-bowl{position:absolute;top:135px;left:13%;width:74%}.deposit-treasure{position:absolute;top:34px;left:9%;width:82%;filter:drop-shadow(0 18px 22px rgb(31 4 2 / 55%))}.deposit-hand{position:absolute;right:16px;bottom:9px;width:78px}.deposit-content{position:relative;z-index:2;padding:0 28px 30px;text-align:center;color:#fff}.deposit-content p{margin:0 0 6px;color:#ffd76a;font-size:12px;font-weight:900;letter-spacing:0}.deposit-title{margin:0;font-size:28px;line-height:1.08;letter-spacing:0}.deposit-copy{display:block;min-height:42px;margin:10px 0 18px;color:#ffe8d5;font-size:14px;line-height:1.45}@media(max-height:620px){.deposit-stage{min-height:500px;scale:.9}}
</style>
