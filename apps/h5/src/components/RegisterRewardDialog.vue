<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import gsap from 'gsap'
import type * as Pixi from 'pixi.js'
import type { H5RegisterRewardPayload } from '../services/h5Api'
import { runtimeConfig } from '../services/runtimeConfig'
import { localProfileImage, localRegisterRewardImage } from '../utils/localAssets'

const props = defineProps<{
  reward: H5RegisterRewardPayload
  claimReward?: () => Promise<H5RegisterRewardPayload | null>
}>()

const emit = defineEmits<{
  complete: []
}>()

const rays = localRegisterRewardImage('rays.webp')
const glow = localRegisterRewardImage('claim_success_bg.webp')
const hero = localRegisterRewardImage('hero.webp')
const banner = localRegisterRewardImage('banner.webp')
const cashIcon = localRegisterRewardImage('icon_cash.webp')
const cashWalletIcon = localRegisterRewardImage('cash_wallet_icon.webp')
const bonusWalletIcon = localRegisterRewardImage('bonus_wallet_icon.webp')
const withdrawableIcon = localProfileImage('wallet-withdrawable.webp')
const claimAudio = new URL('../assets/audio/coin-claim.mp3', import.meta.url).href
const coinFlyAudio = new URL('../assets/audio/coin-fly.wav', import.meta.url).href

const claiming = ref(false)
const closing = ref(false)
const walletsVisible = ref(false)
const cashWalletRef = ref<HTMLElement | null>(null)
const pixiHostRef = ref<HTMLElement | null>(null)
const amountRef = ref<HTMLElement | null>(null)
const amountIconRef = ref<HTMLElement | null>(null)
const displayCash = ref(Number(props.reward.cashBefore || 0))
const displayReward = ref(Number(props.reward.amount || 0))
const finalCash = ref(Number(props.reward.cashAfter || 0) || displayCash.value + displayReward.value)
let closeTimer = 0
let pixiApp: Pixi.Application | null = null
let flightTimeline: gsap.core.Timeline | null = null
let walletTimeline: gsap.core.Timeline | null = null
let numberTween: gsap.core.Tween | null = null

const currencyLabel = computed(() => {
  const rewardCurrency = String(props.reward.currency || '').trim().toUpperCase()
  const runtimeCurrency = runtimeConfig.value.currency
  if (rewardCurrency && rewardCurrency === String(runtimeCurrency.code || '').trim().toUpperCase()) {
    return runtimeCurrency.symbol || runtimeCurrency.code || rewardCurrency
  }
  return rewardCurrency || runtimeCurrency.symbol || runtimeCurrency.code || ''
})

const formatMoney = (value: number) =>
  value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })

const playSound = (source = claimAudio, volume = 0.9) => {
  const audio = new Audio(source)
  audio.volume = volume
  void audio.play().catch(() => undefined)
}

const cubicBezier = (
  start: { x: number; y: number },
  controlA: { x: number; y: number },
  controlB: { x: number; y: number },
  end: { x: number; y: number },
  progress: number,
) => {
  const inverse = 1 - progress
  return {
    x:
      inverse * inverse * inverse * start.x +
      3 * inverse * inverse * progress * controlA.x +
      3 * inverse * progress * progress * controlB.x +
      progress * progress * progress * end.x,
    y:
      inverse * inverse * inverse * start.y +
      3 * inverse * inverse * progress * controlA.y +
      3 * inverse * progress * progress * controlB.y +
      progress * progress * progress * end.y,
  }
}

const getFlightPoints = () => {
  const source = amountIconRef.value?.getBoundingClientRect() || amountRef.value?.getBoundingClientRect()
  const target = cashWalletRef.value?.getBoundingClientRect()
  const fallbackStart = { x: window.innerWidth / 2 - 54, y: window.innerHeight * 0.72 }
  const fallbackEnd = { x: window.innerWidth / 2, y: 58 }
  const start = source
    ? { x: source.left + source.width / 2, y: source.top + source.height / 2 }
    : fallbackStart
  const end = target
    ? { x: target.left + target.width * 0.55, y: target.top + target.height * 0.48 }
    : fallbackEnd
  const controlA = {
    x: start.x - 56,
    y: Math.max(6, start.y - 230),
  }
  const controlB = {
    x: end.x + 28,
    y: Math.max(4, end.y - 150),
  }

  return { start, controlA, controlB, end }
}

const clearPixi = () => {
  flightTimeline?.kill()
  flightTimeline = null
  if (pixiApp) {
    pixiApp.destroy()
    pixiApp = null
  }
  pixiHostRef.value?.replaceChildren()
}

const createCashSprite = (SpriteCtor: typeof Pixi.Sprite, texture: Pixi.Texture, size: number) => {
  const sprite = new SpriteCtor(texture)
  sprite.anchor.set(0.5)
  const textureSize = Math.max(texture.width || 1, texture.height || 1)
  const baseScale = size / textureSize
  sprite.scale.set(baseScale)
  return { sprite, baseScale }
}

const playWalletBounce = () => {
  const wallet = cashWalletRef.value
  if (!wallet) return
  walletTimeline?.kill()
  walletTimeline = gsap.timeline()
  walletTimeline
    .to(wallet, {
      scale: 1.08,
      filter: 'brightness(1.24)',
      duration: 0.16,
      ease: 'back.out(2.2)',
    })
    .to(wallet, {
      scale: 1,
      filter: 'brightness(1)',
      duration: 0.26,
      ease: 'power2.out',
    })
}

const animateNumber = () =>
  new Promise<void>((resolve) => {
    numberTween?.kill()
    const state = {
      cash: displayCash.value,
      reward: displayReward.value,
    }

    numberTween = gsap.to(state, {
      cash: finalCash.value,
      reward: 0,
      duration: 0.9,
      ease: 'power3.out',
      onUpdate: () => {
        displayCash.value = state.cash
        displayReward.value = Math.max(0, state.reward)
      },
      onComplete: () => {
        displayCash.value = finalCash.value
        displayReward.value = 0
        resolve()
      },
    })
  })

const applyRewardPayload = (reward: H5RegisterRewardPayload) => {
  const before = Number(reward.cashBefore || displayCash.value || 0)
  const amount = Number(reward.amount || displayReward.value || 0)
  displayCash.value = before
  displayReward.value = amount
  finalCash.value = Number(reward.cashAfter || 0) || before + amount
}

const playCoinFly = async (onCollecting?: () => void) => {
  const host = pixiHostRef.value
  if (!host) return

  clearPixi()
  const { Application, Assets, Container, Sprite } = await import('pixi.js')
  const app = new Application()
  pixiApp = app
  await app.init({
    width: window.innerWidth,
    height: window.innerHeight,
    backgroundAlpha: 0,
    antialias: true,
    autoDensity: true,
    resolution: Math.min(window.devicePixelRatio || 1, 2),
  })
  host.appendChild(app.canvas)

  const texture = await Assets.load<Pixi.Texture>(cashIcon)
  const layer = new Container()
  app.stage.addChild(layer)

  const { start, controlA, controlB, end } = getFlightPoints()
  const { sprite: mainIcon, baseScale: mainScale } = createCashSprite(Sprite, texture, 42)
  mainIcon.position.set(start.x, start.y)
  mainIcon.alpha = 0
  layer.addChild(mainIcon)

  const particles = Array.from({ length: 30 }, (_, index) => {
    const size = 13 + (index % 5) * 2.8
    const { sprite, baseScale } = createCashSprite(Sprite, texture, size)
    const drift = (index - 14.5) * 4.3
    const arcLift = 18 + (index % 6) * 8
    const startPoint = { x: start.x + drift * 0.55, y: start.y + (index % 2 ? 11 : -8) }
    const endPoint = { x: end.x + drift * 0.18, y: end.y + (index % 4 - 1.5) * 4 }
    const controlPointA = {
      x: controlA.x + drift * 1.1,
      y: Math.max(4, controlA.y - arcLift),
    }
    const controlPointB = {
      x: controlB.x + drift * 0.55,
      y: Math.max(4, controlB.y - (index % 2 ? 18 : -6)),
    }
    sprite.position.set(startPoint.x, startPoint.y)
    sprite.alpha = 0
    layer.addChild(sprite)
    return { sprite, baseScale, startPoint, controlPointA, controlPointB, endPoint }
  })

  await new Promise<void>((resolve) => {
    const mainState = { progress: 0 }
    flightTimeline = gsap.timeline({
      onComplete: () => {
        playWalletBounce()
        window.setTimeout(() => {
          clearPixi()
          resolve()
        }, 120)
      },
    })

    flightTimeline.to(mainState, {
      progress: 1,
      duration: 1.08,
      ease: 'power3.inOut',
      onUpdate: () => {
        const progress = mainState.progress
        const point = cubicBezier(start, controlA, controlB, end, progress)
        const scale = progress < 0.16 ? 0.68 + progress * 2.9 : Math.max(0.18, 1.14 - progress * 0.98)
        mainIcon.position.set(point.x, point.y)
        mainIcon.scale.set(mainScale * scale)
        mainIcon.rotation = -0.16 + progress * 0.62
        mainIcon.alpha = progress < 0.1 ? progress / 0.1 : progress > 0.86 ? Math.max(0, (1 - progress) / 0.14) : 1
      },
    }, 0)

    flightTimeline.call(() => {
      onCollecting?.()
    }, undefined, 0.46)

    particles.forEach((particle, index) => {
      const state = { progress: 0 }
      const delay = 0.03 + index * 0.024
      flightTimeline?.to(state, {
        progress: 1,
        duration: 0.76 + (index % 4) * 0.06,
        ease: 'power3.inOut',
        onUpdate: () => {
          const progress = state.progress
          const point = cubicBezier(
            particle.startPoint,
            particle.controlPointA,
            particle.controlPointB,
            particle.endPoint,
            progress,
          )
          const scale = progress < 0.25 ? 0.55 + progress * 2 : Math.max(0.14, 0.98 - progress * 0.82)
          particle.sprite.position.set(point.x, point.y)
          particle.sprite.scale.set(particle.baseScale * scale)
          particle.sprite.rotation = (index % 2 ? -1 : 1) * (0.35 + progress * 1.2)
          particle.sprite.alpha = progress < 0.16 ? progress / 0.16 : progress > 0.78 ? Math.max(0, (1 - progress) / 0.22) : 0.95
        },
      }, delay)
    })
  })
}

const claim = async () => {
  if (claiming.value) return
  claiming.value = true
  try {
    const claimedReward = props.claimReward ? await props.claimReward() : props.reward
    if (claimedReward) {
      applyRewardPayload(claimedReward)
    }
  } catch (error) {
    claiming.value = false
    return
  }
  walletsVisible.value = true
  await nextTick()
  playSound(claimAudio, 0.95)
  playSound(coinFlyAudio, 0.72)
  let numberPromise: Promise<void> | null = null
  const coinPromise = playCoinFly(() => {
    numberPromise ||= animateNumber()
  })

  await coinPromise
  numberPromise ||= animateNumber()
  await numberPromise
  closeTimer = window.setTimeout(() => {
    closing.value = true
    closeTimer = window.setTimeout(() => emit('complete'), 320)
  }, 560)
}

onMounted(() => {
  playSound()
})

onBeforeUnmount(() => {
  if (closeTimer) window.clearTimeout(closeTimer)
  numberTween?.kill()
  walletTimeline?.kill()
  clearPixi()
})
</script>

<template>
  <Teleport to="body">
    <section class="register-reward" :class="{ 'is-claiming': claiming, 'is-closing': closing }" aria-label="Register gift">
      <div class="register-reward__shade" />
      <div ref="pixiHostRef" class="register-reward__pixi" aria-hidden="true" />

      <div class="register-reward__wallets" :class="{ 'is-visible': walletsVisible }">
        <button class="reward-wallet reward-wallet--bonus" type="button">
          <span class="reward-wallet__content">
            <span class="reward-wallet__label">Bonus Wallet</span>
            <span class="reward-wallet__amount">
              <em>{{ currencyLabel }}</em>
              <strong>0.00</strong>
              <svg viewBox="0 0 5 8" aria-hidden="true">
                <path d="M1 1L4 4L1 7" />
              </svg>
            </span>
          </span>
          <img :src="bonusWalletIcon" alt="" aria-hidden="true" />
        </button>
        <button
          ref="cashWalletRef"
          class="reward-wallet reward-wallet--cash"
          type="button"
        >
          <span class="reward-wallet__content">
            <span class="reward-wallet__label">Cash Wallet</span>
            <span class="reward-wallet__amount">
              <em>{{ currencyLabel }}</em>
              <strong>{{ formatMoney(displayCash) }}</strong>
              <svg viewBox="0 0 5 8" aria-hidden="true">
                <path d="M1 1L4 4L1 7" />
              </svg>
            </span>
          </span>
          <img :src="cashWalletIcon" alt="" aria-hidden="true" />
        </button>
        <button class="reward-wallet reward-wallet--withdraw" type="button">
          <span class="reward-wallet__content">
            <span class="reward-wallet__label">Withdrawable</span>
            <span class="reward-wallet__amount">
              <em>{{ currencyLabel }}</em>
              <strong>0.00</strong>
              <svg viewBox="0 0 5 8" aria-hidden="true">
                <path d="M1 1L4 4L1 7" />
              </svg>
            </span>
          </span>
          <img :src="withdrawableIcon" alt="" aria-hidden="true" />
        </button>
      </div>

      <article class="register-reward__dialog">
        <div class="register-reward__stage">
          <img class="register-reward__rays" :src="rays" alt="" aria-hidden="true" />
          <img class="register-reward__glow" :src="glow" alt="" aria-hidden="true" />
          <img class="register-reward__hero" :src="hero" alt="" aria-hidden="true" />

          <div class="register-reward__banner">
            <img :src="banner" alt="" aria-hidden="true" />
            <svg viewBox="0 0 182 28" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <defs>
                <path id="register-gift-title-arc" d="M 4 22 Q 91 14 178 22" />
                <linearGradient id="register-gift-title-grad" x1="0" y1="0" x2="0" y2="28" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stop-color="#FFD400" />
                  <stop offset="36.59%" stop-color="#FFD400" />
                  <stop offset="100%" stop-color="#FCFB6B" />
                </linearGradient>
              </defs>
              <text
                font-family="Montserrat, sans-serif"
                font-size="13"
                font-weight="800"
                letter-spacing="1"
                fill="url(#register-gift-title-grad)"
                stroke="#9F1617"
                stroke-width="1"
                paint-order="stroke"
              >
                <textPath href="#register-gift-title-arc" startOffset="50%" text-anchor="middle">Register Gift</textPath>
              </text>
            </svg>
          </div>
        </div>

        <div ref="amountRef" class="register-reward__amount">
          <img ref="amountIconRef" :src="cashIcon" alt="" aria-hidden="true" />
          <strong>{{ formatMoney(displayReward) }}</strong>
        </div>

        <button class="register-reward__ok" type="button" :disabled="claiming" @click="claim">
          <span class="register-reward__ok-bottom" />
          <span class="register-reward__ok-top" />
          <span class="register-reward__ok-shine" />
          <span class="register-reward__ok-text">Okay</span>
        </button>
      </article>
    </section>
  </Teleport>
</template>

<style scoped lang="scss">
.register-reward {
  position: fixed;
  inset: 0;
  z-index: 6000;
  display: flex;
  justify-content: center;
  min-height: 100dvh;
  overflow: hidden;
  color: #ffffff;
  pointer-events: auto;
  transition: opacity 0.3s ease, transform 0.3s ease;
  animation: registerRewardFadeIn 0.36s ease-out both;

  &.is-closing {
    opacity: 0;
    transform: scale(0.985);
    pointer-events: none;
  }
}

@keyframes registerRewardFadeIn {
  from {
    opacity: 0;
    transform: scale(0.985);
  }

  to {
    opacity: 1;
    transform: scale(1);
  }
}

.register-reward__shade {
  position: absolute;
  inset: 0;
  background: rgba(2, 8, 22, 0.78);
  backdrop-filter: blur(1px);
}

.register-reward__pixi {
  position: fixed;
  inset: 0;
  z-index: 5;
  pointer-events: none;

  :deep(canvas) {
    display: block;
    width: 100%;
    height: 100%;
  }
}

.register-reward__wallets {
  position: fixed;
  top: 0;
  left: 50%;
  z-index: 2;
  display: grid;
  width: min(100%, 430px);
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  padding: calc(env(safe-area-inset-top) + 12px) 8px 12px;
  border-radius: 0 0 12px 12px;
  background: #171d2d;
  opacity: 0;
  transform: translateX(-50%) translateY(-16px) scale(0.96);
  pointer-events: none;
  transition: opacity 0.26s ease, transform 0.26s ease;

  &.is-visible {
    opacity: 1;
    transform: translateX(-50%);
    pointer-events: auto;
  }
}

.reward-wallet {
  position: relative;
  height: 47px;
  min-width: 0;
  overflow: visible;
  border: 0.5px solid transparent;
  border-radius: 12px;
  color: #ffffff;
  text-align: left;
  text-shadow: 0 2px 0 rgba(27, 39, 76, 0.55);

  img {
    position: absolute;
    right: -8px;
    top: -8px;
    z-index: 0;
    width: 39px;
    height: 39px;
    object-fit: contain;
    filter: drop-shadow(0 4px 4px rgba(0, 0, 0, 0.22));
  }

  &::before,
  &::after {
    position: absolute;
    pointer-events: none;
    content: '';
  }

  &::before {
    inset: 0 0 50%;
    border-radius: 12px 12px 0 0;
    background: linear-gradient(rgba(255, 255, 255, 0.42), rgba(255, 255, 255, 0));
  }

  &::after {
    top: -20%;
    bottom: -20%;
    left: 6%;
    width: 45%;
    transform: skewX(-18deg);
    opacity: 0.7;
    background: linear-gradient(rgba(255, 255, 255, 0), rgba(255, 255, 255, 0.1));
  }
}

.reward-wallet--bonus {
  background:
    linear-gradient(0deg, #b26df6, #814eee) padding-box,
    linear-gradient(90deg, #ffe53e, rgba(224, 152, 58, 0.4), #ffe53e) border-box;
  box-shadow: #e83bff 0 0 2px 1px inset;
}

.reward-wallet--cash {
  background:
    linear-gradient(0deg, #6499f6, #3d7cf1) padding-box,
    linear-gradient(90deg, #ffe53e, rgba(224, 152, 58, 0.4), #ffe53e) border-box;
  box-shadow: #02e6ff 0 0 2px 1px inset;
}

.reward-wallet--withdraw {
  background:
    linear-gradient(0deg, #3fc27a, #15b84b) padding-box,
    linear-gradient(90deg, #ffe53e, rgba(224, 152, 58, 0.4), #ffe53e) border-box;
  box-shadow: #6dffbd 0 0 2px 1px inset;
}

.reward-wallet__content {
  position: relative;
  z-index: 1;
  display: flex;
  height: 100%;
  flex-direction: column;
  justify-content: space-between;
  padding: 5px 28px 5px 8px;
}

.reward-wallet__label {
  overflow: hidden;
  font-family: 'Poetsen One', Montserrat, sans-serif;
  font-size: 13px;
  font-weight: 400;
  line-height: 1.25;
  white-space: nowrap;
  text-overflow: ellipsis;
  -webkit-text-stroke: 1px rgba(0, 68, 160, 0.85);
  paint-order: stroke;
}

.reward-wallet__amount {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: 2px;
  color: #ffffff;

  em {
    font-size: 12px;
    font-style: normal;
    font-weight: 800;
  }

  strong {
    overflow: hidden;
    font-family: Montserrat, sans-serif;
    font-size: 14px;
    font-weight: 900;
    line-height: 1;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  svg {
    width: 5px;
    height: 8px;
    flex: 0 0 auto;
    transform: translateY(1px);

    path {
      fill: none;
      stroke: currentColor;
      stroke-width: 2;
      stroke-linecap: round;
      stroke-linejoin: round;
    }
  }
}

.register-reward__dialog {
  position: relative;
  z-index: 1;
  width: min(100%, 430px);
  min-height: 100dvh;
  padding-top: calc(env(safe-area-inset-top) + 92px);
  display: flex;
  flex-direction: column;
  align-items: center;
  pointer-events: none;
}

.register-reward__stage {
  position: relative;
  width: 100%;
  height: 354px;
  overflow: visible;
}

.register-reward__rays,
.register-reward__glow,
.register-reward__hero {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  pointer-events: none;
  user-select: none;
}

.register-reward__rays {
  animation: registerRewardSpin 7.6s linear infinite;
  opacity: 0.98;
}

.register-reward__glow {
  transform: scale(0.96);
  opacity: 0.95;
  mix-blend-mode: screen;
}

.register-reward__hero {
  transform: translateX(-3%) translateY(8px) scale(1.08);
  //animation: registerRewardHeroFloat 2.6s ease-in-out infinite;
}

.register-reward__banner {
  position: absolute;
  left: 50%;
  bottom: 32px;
  width: 280px;
  height: 72px;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: fill;
  }

  svg {
    position: relative;
    z-index: 1;
    width: calc(100% - 34px);
    margin-top: -20px;
    filter: drop-shadow(0 1.5px 0 #ad4202);
  }
}

.register-reward__amount {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: -16px;
  min-height: 50px;

  img {
    width: 35px;
    height: 36px;
    object-fit: contain;
  }

  strong {
    background: linear-gradient(180deg, #b6fa73 0%, #68b41e 100%);
    background-clip: text;
    -webkit-text-fill-color: transparent;
    font-size: 32px;
    font-weight: 900;
    line-height: 1;
    filter: drop-shadow(0 2px 0 rgba(43, 83, 12, 0.36));
  }
}

.register-reward__ok {
  position: relative;
  width: 170px;
  height: 38px;
  margin-top: 20px;
  border: 0;
  cursor: pointer;
  pointer-events: auto;
  background: transparent;
  filter: drop-shadow(0 4px 16px rgba(0, 0, 0, 0.14));

  &:disabled {
    cursor: default;
  }
}

.register-reward__ok-bottom,
.register-reward__ok-top,
.register-reward__ok-shine {
  position: absolute;
  pointer-events: none;
}

.register-reward__ok-bottom {
  inset: 8px 0 0;
  border: 0.5px solid #ffa43a;
  border-radius: 28px;
  background: linear-gradient(180deg, #ffdc34 0%, #fe762d 100%);
  box-shadow: inset 0 -1px 4px #ff971d;
}

.register-reward__ok-top {
  inset: 0 0 8px;
  border-radius: 18px;
  background: linear-gradient(180deg, #fff0b0 0%, #ffbc18 100%);
  box-shadow: inset 0 0 4px #fff2c6, inset 0 -4px 4px rgba(255, 255, 255, 0.25);
}

.register-reward__ok-shine {
  left: 7px;
  right: 7px;
  top: 3px;
  height: 20px;
  border-radius: 999px;
  background: linear-gradient(180deg, #fffdba 0%, rgba(255, 253, 186, 0.04) 100%);
}

.register-reward__ok-text {
  position: relative;
  z-index: 1;
  color: #973305;
  font-size: 16px;
  font-weight: 800;
}

.register-reward.is-claiming {
  .register-reward__amount {
    animation: registerRewardAmountPop 0.62s ease-in-out infinite alternate;
  }

  .register-reward__ok {
    opacity: 0.64;
    transform: scale(0.98);
  }
}

@keyframes registerRewardSpin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes registerRewardHeroFloat {
  0%, 100% {
    transform: translateX(-3%) translateY(8px) scale(1.08);
  }
  50% {
    transform: translateX(-3%) translateY(-2px) scale(1.1);
  }
}

@keyframes registerRewardAmountPop {
  from {
    transform: scale(1);
  }
  to {
    transform: scale(1.06);
  }
}

@media (max-width: 370px) {
  .reward-wallet {
    .reward-wallet__content {
      padding-right: 22px;
    }

    .reward-wallet__label {
      font-size: 11px;
    }

    .reward-wallet__amount strong {
      font-size: 12px;
    }
  }

  .register-reward__dialog {
    padding-top: calc(env(safe-area-inset-top) + 88px);
  }
}
</style>
