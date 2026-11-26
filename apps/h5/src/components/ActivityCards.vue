<script setup lang="ts">
import redBg from '../assets/images/activity-red-bg.webp'
import redConfetti from '../assets/images/activity-red-confetti.webp'
import redPackets from '../assets/images/activity-red-packets.webp'
import redFalling from '../assets/images/activity-red-falling.webp'
import redHero from '../assets/images/activity-red-hero.webp'
import spinBg from '../assets/images/activity-spin-bg.webp'
import spinA from '../assets/images/activity-spin-float-a.webp'
import spinB from '../assets/images/activity-spin-float-b.webp'
import spinWheel from '../assets/images/activity-spin-wheel.webp'
import chestBg from '../assets/images/activity-chest-bg.webp'
import chestLeft from '../assets/images/activity-chest-left.webp'
import chestRight from '../assets/images/activity-chest-right.webp'
import chestGoldA from '../assets/images/activity-chest-gold-a.webp'
import chestGoldB from '../assets/images/activity-chest-gold-b.webp'
import chestGoldC from '../assets/images/activity-chest-gold-c.webp'

const images = {
  redBg,
  redConfetti,
  redPackets,
  redFalling,
  redHero,
  spinBg,
  spinA,
  spinB,
  spinWheel,
  chestBg,
  chestLeft,
  chestRight,
  chestGoldA,
  chestGoldB,
  chestGoldC,
}
</script>

<template>
  <section class="activity-cards" aria-label="Event shortcuts">
    <button class="activity-card activity-card--red" type="button" aria-label="Red Packet Rain">
      <img class="red-bg" :src="images.redBg" alt="" />
      <img class="red-confetti" :src="images.redConfetti" alt="" />
      <img class="red-packets" :src="images.redPackets" alt="" />
      <img class="red-falling red-falling--a" :src="images.redFalling" alt="" />
      <img class="red-falling red-falling--b" :src="images.redFalling" alt="" />
      <img class="red-falling red-falling--c" :src="images.redFalling" alt="" />
      <img class="red-hero" :src="images.redHero" alt="" />
      <span class="activity-label activity-label--timer">00 : 01 : 26</span>
    </button>

    <button class="activity-card activity-card--spin" type="button" aria-label="Invite Spin">
      <img class="activity-bg" :src="images.spinBg" alt="" />
      <img class="spin-float spin-float--a" :src="images.spinA" alt="" />
      <img class="spin-float spin-float--b" :src="images.spinB" alt="" />
      <img class="spin-wheel" :src="images.spinWheel" alt="" />
      <span class="activity-label">Invite Spin</span>
    </button>

    <button class="activity-card activity-card--chest" type="button" aria-label="Invite Chest">
      <img class="activity-bg" :src="images.chestBg" alt="" />
      <img class="chest-gold chest-gold--a" :src="images.chestGoldA" alt="" />
      <img class="chest-gold chest-gold--b" :src="images.chestGoldB" alt="" />
      <img class="chest-gold chest-gold--c" :src="images.chestGoldC" alt="" />
      <img class="chest-left" :src="images.chestLeft" alt="" />
      <img class="chest-right" :src="images.chestRight" alt="" />
      <span class="activity-label">Invite Chest</span>
    </button>
  </section>
</template>

<style scoped lang="scss">
.activity-cards {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.activity-card {
  position: relative;
  height: 68px;
  padding: 0;
  overflow: hidden;
  border-radius: 9px;
  background: #142448;
  transform: translateZ(0);

  img {
    pointer-events: none;
    position: absolute;
  }
}

.activity-card--spin {
  background: linear-gradient(
    254deg,
    rgb(249, 239, 165) 0%,
    rgb(250, 171, 200) 24.04%,
    rgb(232, 123, 249) 44.86%,
    rgb(167, 62, 242) 69.07%,
    rgb(79, 22, 222) 95.9%
  );
}

.activity-card--chest {
  background: linear-gradient(
    254deg,
    rgb(250, 185, 255) 0%,
    rgb(205, 174, 255) 22.42%,
    rgb(141, 100, 254) 46.66%,
    rgb(103, 68, 230) 74.28%,
    rgb(36, 46, 148) 100%
  );
}

.activity-bg,
.red-bg {
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.activity-card--spin .activity-bg,
.activity-card--chest .activity-bg {
  opacity: 0.94;
}

.red-confetti {
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  animation: confetti-sway 2.6s ease-in-out infinite;
}

.red-packets {
  top: -2px;
  left: 0;
  width: 100%;
  height: 44px;
  object-fit: contain;
  animation: packets-float 1.8s ease-in-out infinite;
}

.red-falling {
  top: -8px;
  width: 15px;
  height: 20px;
  object-fit: contain;

  &--a {
    left: 6px;
    animation: fall-a 1.6s ease-in infinite;
  }

  &--b {
    left: 48px;
    animation: fall-b 1.9s ease-in 0.4s infinite;
  }

  &--c {
    right: 9px;
    animation: fall-c 1.75s ease-in 0.85s infinite;
  }
}

.red-hero {
  top: 4px;
  left: 50%;
  width: 58px;
  height: 55px;
  object-fit: contain;
  transform: translateX(-50%);
  animation: hero-bob 1.6s ease-in-out infinite;
}

.spin-float {
  object-fit: contain;
  animation: float-card 2s ease-in-out infinite;
  --spin-float-rotate: -24deg;
  --spin-float-scale-x: 1;

  &--a {
    top: 13px;
    left: 12px;
    width: 36px;
    --spin-float-rotate: 18deg;
    --spin-float-scale-x: -1;
  }

  &--b {
    top: -10px;
    right: -1px;
    width: 52px;
    animation-delay: 0.25s;
    --spin-float-rotate: -18deg;
    --spin-float-scale-x: 1;
  }
}

.spin-wheel {
  top: 50%;
  left: 50%;
  width: 94px;
  max-width: 72%;
  transform: translate(-50%, -50%);
  animation: pulse-wheel 1.5s ease-in-out infinite;
}

.chest-gold {
  z-index: 2;
  width: 15px;
  object-fit: contain;
  animation: spin-gold 2.2s ease-in-out infinite;

  &--a {
    top: 32px;
    left: 4px;
  }

  &--b {
    top: 24px;
    right: 9px;
    animation-delay: 0.3s;
  }

  &--c {
    bottom: 27px;
    left: 28px;
    animation-delay: 0.6s;
  }
}

.chest-left {
  bottom: -11px;
  left: 8px;
  z-index: 2;
  width: 67px;
  animation: bounce-chest 2s ease-out infinite;
}

.chest-right {
  right: 12px;
  bottom: 8px;
  z-index: 1;
  width: 54px;
  animation: bounce-chest 2s ease-out 0.3s infinite;
}

.activity-label {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 6;
  display: grid;
  height: 22px;
  place-items: center;
  padding-top: 2px;
  background: linear-gradient(90deg, #2b41c0 0%, #68bdff 50%, #2b41c0 100%);
  color: #ffffff;
  font-size: 13px;
  font-style: italic;
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.42);
  clip-path: ellipse(70% 70% at 50% 100%);

  &--timer {
    font-size: 12px;
    letter-spacing: 0.5px;
  }
}

@keyframes confetti-sway {
  0%,
  100% {
    transform: translateY(0) rotate(-4deg);
  }

  50% {
    transform: translateY(-4px) rotate(4deg);
  }
}

@keyframes packets-float {
  0%,
  100% {
    transform: translateY(0) rotate(-2deg);
  }

  50% {
    transform: translateY(4px) rotate(2deg);
  }
}

@keyframes hero-bob {
  0%,
  100% {
    transform: translateX(-50%) translateY(0) rotate(-3deg) scale(1);
  }

  50% {
    transform: translateX(-50%) translateY(-5px) rotate(3deg) scale(1.04);
  }
}

@keyframes fall-a {
  0% {
    transform: translateY(-14px) rotate(-12deg);
    opacity: 0;
  }

  15%,
  85% {
    opacity: 1;
  }

  100% {
    transform: translateY(72px) rotate(12deg);
    opacity: 0;
  }
}

@keyframes fall-b {
  0% {
    transform: translateY(-12px) rotate(10deg);
    opacity: 0;
  }

  15%,
  85% {
    opacity: 1;
  }

  100% {
    transform: translateY(76px) rotate(-8deg);
    opacity: 0;
  }
}

@keyframes fall-c {
  0% {
    transform: translateY(-18px) rotate(-6deg);
    opacity: 0;
  }

  15%,
  85% {
    opacity: 1;
  }

  100% {
    transform: translateY(70px) rotate(16deg);
    opacity: 0;
  }
}

@keyframes float-card {
  0%,
  100% {
    transform: rotate(var(--spin-float-rotate)) scaleX(var(--spin-float-scale-x)) translateY(0);
  }

  50% {
    transform: rotate(var(--spin-float-rotate)) scaleX(var(--spin-float-scale-x)) translateY(-12px);
  }
}

@keyframes pulse-wheel {
  0%,
  100% {
    transform: translate(-50%, -50%) scale(1);
  }

  25%,
  75% {
    transform: translate(-50%, -50%) scale(0.92);
  }

  50% {
    transform: translate(-50%, -50%) scale(1.05);
  }
}

@keyframes spin-gold {
  0%,
  100% {
    transform: rotate(0deg) translateY(0);
  }

  50% {
    transform: rotate(180deg) translateY(-8px);
  }
}

@keyframes bounce-chest {
  0%,
  100% {
    transform: translateY(0);
  }

  15% {
    transform: translateY(-6px);
  }

  30% {
    transform: translateY(0);
  }
}

@media (max-width: 374px) {
  .activity-cards {
    gap: 8px;
  }

  .activity-label {
    font-size: 12px;
  }
}
</style>
