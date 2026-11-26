<script setup lang="ts">
import { useRouter } from 'vue-router'
import spinBg from '../assets/images/reference/quick-spin-bg.webp'
import spinLeft from '../assets/images/reference/quick-spin-left.webp'
import spinRight from '../assets/images/reference/quick-spin-right.webp'
import spinWheel from '../assets/images/reference/quick-spin-wheel.webp'
import rewardsBg from '../assets/images/reference/quick-rewards-bg.webp'
import rewardsHero from '../assets/images/reference/quick-rewards-hero.webp'
import checkinBg from '../assets/images/reference/quick-checkin-bg.webp'
import checkinHero from '../assets/images/reference/quick-checkin-hero.webp'

const router = useRouter()
const openActivity = (slug: string) => router.push({ name: 'activity-detail', params: { slug } })
</script>

<template>
  <section class="activity-cards" aria-label="Event shortcuts">
    <button class="activity-card activity-card--spin" type="button" @click="openActivity('invitation-wheel')">
      <img class="activity-bg" :src="spinBg" alt="" />
      <img class="spin-left" :src="spinLeft" alt="" />
      <img class="spin-right" :src="spinRight" alt="" />
      <img class="spin-wheel" :src="spinWheel" alt="" />
      <span>Invite Spin</span>
    </button>
    <button class="activity-card" type="button" @click="openActivity('welfare-card')">
      <img class="activity-bg" :src="rewardsBg" alt="" />
      <img class="rewards-hero" :src="rewardsHero" alt="" />
      <span>Rewards Card</span>
    </button>
    <button class="activity-card" type="button" @click="openActivity('daily-sign-in')">
      <img class="activity-bg" :src="checkinBg" alt="" />
      <img class="checkin-hero" :src="checkinHero" alt="" />
      <span>Daily Check-In</span>
    </button>
  </section>
</template>

<style scoped lang="scss">
.activity-cards { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: 9px; margin: 12px 0; }
.activity-card { position: relative; height: 71px; overflow: hidden; border-radius: 8px; padding: 0; isolation: isolate; }
.activity-card img { position: absolute; pointer-events: none; object-fit: contain; }
.activity-bg { inset: 0; width: 100%; height: 100%; object-fit: cover !important; }
.spin-left { width: 64px; height: 60px; left: -14px; top: 3px; }
.spin-right { width: 92px; height: 85px; right: -15px; top: -13px; }
.spin-wheel { width: 92px; height: 94px; left: 10px; top: -11px; animation: spin-sway 4s ease-in-out infinite alternate; }
.rewards-hero { width: 112px; height: 74px; left: calc(50% - 56px); top: -4px; }
.checkin-hero { width: 100px; height: 72px; left: calc(50% - 50px); top: -4px; }
.activity-card span { position: absolute; z-index: 1; bottom: 0; left: 0; width: 100%; height: 15px; color: white; background: linear-gradient(90deg,#386edb,#47acff,#386edb); font-size: 11px; font-weight: 700; font-style: italic; line-height: 15px; white-space: nowrap; }
@keyframes spin-sway { from { transform: rotate(-6deg); } to { transform: rotate(6deg); } }
@media(prefers-reduced-motion:reduce) { .spin-wheel { animation: none; } }
</style>
