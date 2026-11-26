<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { installPWA } from '../services/pwa'
import { localDialogImage } from '../utils/localAssets'
import { isAndroidAppRuntime } from '../utils/runtimeEnvironment'

type GuideDialog = 'first-deposit' | 'invite-chest' | 'register-gift' | 'app-download'

const props = defineProps<{
  visible: boolean
  initial?: GuideDialog
  chainFirstDeposit?: boolean
}>()

const emit = defineEmits<{
  complete: []
}>()

const router = useRouter()
const androidAppRuntime = isAndroidAppRuntime()
const currentDialog = ref<GuideDialog | null>(null)
let previousOverflow = ''

const halo = localDialogImage('dialog-halo.webp')
const confetti = localDialogImage('dialog-confetti.webp')
const gift = localDialogImage('dialog-gift.webp')
const appHero = localDialogImage('dialog-app-hero.webp')
const firstDepositGlow = localDialogImage('activity__guide-flow-dialogs__bg_glow.webp')
const firstDepositStars = localDialogImage('activity__guide-flow-dialogs__first_deposit_bg_stars.webp')
const firstDepositBowl = localDialogImage('activity__guide-flow-dialogs__first_deposit_bowl.webp')
const firstDepositTreasure = localDialogImage('activity__guide-flow-dialogs__first_deposit_guide_treasure.webp')
const inviteChest = localDialogImage('activity__guide-flow-dialogs__invite_chest_chest.webp')
const inviteChestBanner = localDialogImage('activity__guide-flow-dialogs__invite_chest_banner.webp')
const guideRays = localDialogImage('bonus__rays.webp')
const tapHand = localDialogImage('activity__common__tap_hand.webp')
const inviteSocials = [
  { label: 'Facebook', image: localDialogImage('social-facebook.webp') },
  { label: 'Messenger', image: localDialogImage('social-messenger.webp') },
  { label: 'WhatsApp', image: localDialogImage('social-whatsapp.webp') },
  { label: 'Telegram', image: localDialogImage('social-telegram.webp') },
  { label: 'Instagram', image: localDialogImage('social-instagram.webp') },
  { label: 'TikTok', image: localDialogImage('social-tiktok.webp') },
  { label: 'X', image: localDialogImage('social-x.webp') },
]

const lockScroll = () => {
  previousOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
}

const unlockScroll = () => {
  document.body.style.overflow = previousOverflow
}

const closeCurrent = () => {
  if (currentDialog.value === 'first-deposit') {
    currentDialog.value = 'invite-chest'
    return
  }

  if (currentDialog.value === 'invite-chest' && props.chainFirstDeposit !== false) {
    currentDialog.value = 'register-gift'
    return
  }

  if (currentDialog.value === 'register-gift') {
    if (androidAppRuntime) {
      currentDialog.value = null
      emit('complete')
      return
    }
    currentDialog.value = 'app-download'
    return
  }

  currentDialog.value = null
  emit('complete')
}

const installApplication = async () => {
  await installPWA()
  currentDialog.value = null
  emit('complete')
}

const openFirstDeposit = () => {
  currentDialog.value = null
  emit('complete')
  void router.push({ name: 'wallet-deposit' })
}

const shareInvite = async () => {
  const link = `${window.location.origin}/share?tab=social`
  try {
    await navigator.clipboard?.writeText(link)
  } catch {
    // Clipboard may be unavailable on non-secure local origins; the dialog text still mirrors source behavior.
  }
}

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      const initialDialog = props.initial ?? 'register-gift'
      if (androidAppRuntime && initialDialog === 'app-download') {
        currentDialog.value = null
        emit('complete')
        return
      }

      currentDialog.value = initialDialog
      lockScroll()
      return
    }

    currentDialog.value = null
    unlockScroll()
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  unlockScroll()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="guide-overlay">
      <section v-if="visible && currentDialog" class="guide-flow" aria-label="Promotions">
        <Transition name="guide-dialog" mode="out-in">
          <article
            v-if="currentDialog === 'first-deposit'"
            key="first-deposit"
            class="guide-card guide-card--first"
          >
            <img class="guide-first__glow" :src="firstDepositGlow" alt="" aria-hidden="true" />
            <div class="guide-first__stars" aria-hidden="true">
              <img class="guide-first__stars-img" :src="firstDepositStars" alt="" />
              <svg class="guide-first__title" viewBox="0 0 286 73" preserveAspectRatio="xMidYMid meet" aria-label="First Deposit Super Rewards">
                <defs>
                  <linearGradient id="guide-first-stroke" x1="0" y1="0" x2="0" y2="73" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stop-color="#d334b8" />
                    <stop offset="100%" stop-color="#9d3ddd" />
                  </linearGradient>
                  <linearGradient id="guide-first-fill" x1="0" y1="0" x2="0" y2="73" gradientUnits="userSpaceOnUse">
                    <stop offset="11.427%" stop-color="#ffffff" />
                    <stop offset="42.931%" stop-color="#fff3c4" />
                    <stop offset="61.675%" stop-color="#ffffff" />
                    <stop offset="96.348%" stop-color="#fff3c4" />
                  </linearGradient>
                  <filter id="guide-first-shadow" x="-10%" y="-10%" width="120%" height="130%">
                    <feDropShadow dx="-1" dy="3" stdDeviation="0" flood-color="#132971" />
                  </filter>
                </defs>
                <g
                  filter="url(#guide-first-shadow)"
                  font-family="Montserrat, Helvetica Neue, sans-serif"
                  font-weight="900"
                  font-style="italic"
                  font-size="32"
                  text-anchor="middle"
                  stroke="url(#guide-first-stroke)"
                  stroke-width="3"
                  stroke-linejoin="round"
                  fill="url(#guide-first-fill)"
                  paint-order="stroke"
                >
                  <text x="50%" y="32">First Deposit</text>
                  <text x="50%" y="68">Super Rewards</text>
                </g>
              </svg>
            </div>

            <div class="guide-first__bowl">
              <img class="guide-first__bowl-img" :src="firstDepositBowl" alt="" aria-hidden="true" />
              <p>Complete your first deposit to claim your Rewards !</p>
            </div>

            <div class="guide-first__treasure-wrap">
              <img class="guide-first__treasure" :src="firstDepositTreasure" alt="" aria-hidden="true" />
              <button class="guide-first__cta" type="button" @click="openFirstDeposit">
                <span aria-hidden="true" />
                <strong>Deposit Now</strong>
              </button>
              <img class="guide-first__hand" :src="tapHand" alt="" aria-hidden="true" />
            </div>

            <button class="guide-close guide-close--first" type="button" aria-label="Close" @click="closeCurrent">
              <span />
            </button>
          </article>

          <article
            v-else-if="currentDialog === 'invite-chest'"
            key="invite-chest"
            class="guide-card guide-card--invite"
          >
            <img class="guide-invite__glow" :src="firstDepositGlow" alt="" aria-hidden="true" />
            <img class="guide-invite__rays" :src="guideRays" alt="" aria-hidden="true" />

            <div class="guide-invite__chest-wrap">
              <img class="guide-invite__chest" :src="inviteChest" alt="" aria-hidden="true" />
            </div>
            <button class="guide-close guide-close--invite" type="button" aria-label="Close" @click="closeCurrent">
              <span />
            </button>

            <div class="guide-invite__banner">
              <img :src="inviteChestBanner" alt="" aria-hidden="true" />
              <svg viewBox="0 0 220 40" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                <defs>
                  <path id="guide-invite-arc" d="M 10 28 Q 110 14 210 28" />
                  <filter id="guide-invite-shadow" x="-10%" y="-10%" width="120%" height="160%">
                    <feDropShadow dx="0" dy="2" stdDeviation="1.5" flood-color="#1a3a8e" flood-opacity="0.6" />
                  </filter>
                </defs>
                <text
                  font-family="'Arial Rounded MT Bold', 'Arial Black', sans-serif"
                  font-size="20"
                  font-weight="900"
                  fill="none"
                  stroke="#2848ae"
                  stroke-width="4"
                  stroke-linejoin="round"
                  letter-spacing="1.2"
                >
                  <textPath href="#guide-invite-arc" startOffset="50%" text-anchor="middle">Invite Chest</textPath>
                </text>
                <text
                  font-family="'Arial Rounded MT Bold', 'Arial Black', sans-serif"
                  font-size="20"
                  font-weight="900"
                  fill="#ffffff"
                  filter="url(#guide-invite-shadow)"
                  letter-spacing="1.2"
                >
                  <textPath href="#guide-invite-arc" startOffset="50%" text-anchor="middle">Invite Chest</textPath>
                </text>
              </svg>
            </div>

            <div class="guide-invite__panel">
              <p class="guide-invite__headline">Invite friends to earn commission. Earn unlimited rewards!</p>
              <div class="guide-invite__socials">
                <button
                  v-for="social in inviteSocials"
                  :key="social.label"
                  type="button"
                  :aria-label="social.label"
                  @click="shareInvite"
                >
                  <img :src="social.image" :alt="social.label" />
                </button>
                <button class="guide-invite__share" type="button" aria-label="Share" @click="shareInvite">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                    <path d="M13 2L22 11L13 19.5V14C6 14 3 21.5 3 21.5C3 13 5.5 7.5 13 7.5V2Z" fill="white" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </button>
              </div>
              <p class="guide-invite__hint">The link has been copied.<br />Click to share.</p>
            </div>
          </article>

          <article
            v-else-if="currentDialog === 'register-gift'"
            key="register-gift"
            class="guide-card guide-card--gift"
          >
            <svg class="guide-title guide-title--gift" viewBox="0 0 285 98" aria-hidden="true">
              <defs>
                <path id="guide-gift-arc" d="M 15 72 Q 142 35 270 72" />
                <linearGradient id="guide-gift-title" x1="0" y1="0" x2="0" y2="98">
                  <stop offset="0%" stop-color="#fff8e7" />
                  <stop offset="100%" stop-color="#ffdd88" />
                </linearGradient>
              </defs>
              <text class="guide-title__stroke">
                <textPath href="#guide-gift-arc" startOffset="50%" text-anchor="middle">REGISTER GIFT</textPath>
              </text>
              <text class="guide-title__fill">
                <textPath href="#guide-gift-arc" startOffset="50%" text-anchor="middle">REGISTER GIFT</textPath>
              </text>
            </svg>

            <img class="guide-halo guide-halo--gift" :src="halo" alt="" aria-hidden="true" />
            <img class="guide-confetti guide-confetti--gift" :src="confetti" alt="" aria-hidden="true" />

            <div class="gift-panel">
              <img class="gift-panel__hero" :src="gift" alt="" aria-hidden="true" />
              <div class="gift-panel__glass" aria-hidden="true" />
              <div class="gift-panel__inner" aria-hidden="true" />
              <div class="gift-panel__content">
                <div class="guide-amount">
                  <strong>1.00 to 77.00</strong>
                  <span>CASH</span>
                </div>
                <div class="gift-panel__subtitle">
                  <i />
                  <span>Nice to Meet You!</span>
                  <i />
                </div>
              </div>

              <div class="gift-panel__curve" aria-hidden="true" />
              <button class="guide-cta guide-cta--gift" type="button" @click="closeCurrent">
                Claim Now
              </button>
            </div>

            <button class="guide-close" type="button" aria-label="Close" @click="closeCurrent">
              <span />
            </button>
          </article>

          <article v-else key="app-download" class="guide-card guide-card--download">
            <svg class="download-title" viewBox="0 0 325 116" aria-hidden="true">
              <defs>
                <path id="guide-download-arc" d="M 54 74 Q 162 36 271 74" />
                <linearGradient id="guide-download-fill" x1="0" y1="34" x2="0" y2="76">
                  <stop offset="0%" stop-color="#ffffff" />
                  <stop offset="46%" stop-color="#d3fcff" />
                  <stop offset="100%" stop-color="#16a2ff" />
                </linearGradient>
                <filter id="guide-download-glow" x="0" y="0" width="325" height="116">
                  <feGaussianBlur stdDeviation="12.95" />
                </filter>
              </defs>
              <ellipse
                class="download-title__glow"
                cx="162.5"
                cy="68"
                rx="116"
                ry="20"
                filter="url(#guide-download-glow)"
              />
              <text class="download-title__outer">
                <textPath href="#guide-download-arc" startOffset="50%" text-anchor="middle">DOWNLOAD APP</textPath>
              </text>
              <text class="download-title__stroke">
                <textPath href="#guide-download-arc" startOffset="50%" text-anchor="middle">DOWNLOAD APP</textPath>
              </text>
              <text class="download-title__fill">
                <textPath href="#guide-download-arc" startOffset="50%" text-anchor="middle">DOWNLOAD APP</textPath>
              </text>
            </svg>
            <img class="guide-halo guide-halo--download" :src="halo" alt="" aria-hidden="true" />
            <img class="guide-confetti guide-confetti--download" :src="confetti" alt="" aria-hidden="true" />
            <img class="download-hero" :src="appHero" alt="" aria-hidden="true" />

            <div class="download-panel">
              <div class="download-panel__shine" aria-hidden="true" />
              <div class="download-panel__content">
                <div class="guide-amount guide-amount--download">
                  <strong>1.00 to 77.00</strong>
                  <span>Bonus</span>
                </div>
                <p>For a Better Experience, For Enhanced Security.</p>
              </div>

              <div class="download-panel__curve" aria-hidden="true" />
              <button class="guide-cta guide-cta--download" type="button" @click="installApplication">
                Download App
              </button>
            </div>

            <button class="guide-close" type="button" aria-label="Close" @click="closeCurrent">
              <span />
            </button>
          </article>
        </Transition>
      </section>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.guide-flow {
  position: fixed;
  inset: 0;
  z-index: 900;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100dvh;
  overflow: hidden;
  background: rgba(2, 6, 17, 0.62);
  backdrop-filter: blur(2px);
}

.guide-card {
  position: relative;
  width: min(calc(100vw - 58px), 316px);
  transform-origin: center center;
}

.guide-card--gift {
  top: 26px;
}

.guide-card--download {
  top: 51px;
  width: min(calc(100vw - 78px), 298px);
}

.guide-card--first {
  align-self: flex-start;
  display: flex;
  width: min(100vw, 430px);
  min-height: 0;
  margin-top: clamp(-112px, calc(100dvh - 704px), 0px);
  flex-direction: column;
  align-items: center;
  overflow: visible;
}

.guide-card--invite {
  top: -46px;
  display: flex;
  width: min(calc(100vw - 64px), 308px);
  flex-direction: column;
  align-items: center;
}

.guide-first__glow,
.guide-first__stars,
.guide-first__stars-img,
.guide-first__title,
.guide-first__bowl-img,
.guide-first__treasure {
  position: absolute;
  pointer-events: none;
}

.guide-first__glow {
  top: 200px;
  left: 50%;
  z-index: -1;
  width: min(360px, 96vw);
  opacity: 0.8;
  transform: translateX(-50%);
  animation: guide-first-glow 2.8s ease-in-out infinite;
}

.guide-first__stars {
  position: relative;
  z-index: 1;
  width: min(323px, 86vw);
  aspect-ratio: 1043 / 1313;
  overflow: visible;
  animation: guide-first-top-float 2.45s ease-in-out infinite;
}

.guide-first__stars-img {
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  user-select: none;
  filter: drop-shadow(0 13px 16px rgba(25, 4, 72, 0.32));
}

.guide-first__title {
  top: 28%;
  left: 50%;
  z-index: 2;
  width: min(286px, 74vw);
  overflow: visible;
  transform: translateX(-50%);
  animation: guide-first-title 2.3s ease-in-out infinite;
}

.guide-first__bowl {
  position: relative;
  z-index: 3;
  width: min(284px, 76vw);
  margin-top: -132px;
  aspect-ratio: 1136 / 756;
  filter: drop-shadow(0 8px 9px rgba(21, 8, 83, 0.24));
}

.guide-first__bowl-img {
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  user-select: none;
}

.guide-first__bowl p {
  position: absolute;
  top: calc(50% - 20px);
  left: 50%;
  z-index: 1;
  width: 65%;
  margin: 0;
  color: #fff901;
  font-family: Montserrat, sans-serif;
  font-size: clamp(16px, 4.6vw, 18px);
  font-style: italic;
  font-weight: 800;
  line-height: 1.2;
  text-align: center;
  text-shadow: 0 1px 0 #a3b5fe;
  transform: translate(-50%, -50%);
}

.guide-first__treasure-wrap {
  position: relative;
  z-index: 4;
  width: min(356px, 95vw);
  margin-top: -220px;
  aspect-ratio: 1426 / 1412;
  animation: guide-first-treasure-float 2.45s ease-in-out infinite;
}

.guide-first__treasure {
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter:
    drop-shadow(0 12px 16px rgba(27, 7, 74, 0.32))
    saturate(1.08);
}

.guide-first__cta {
  position: absolute;
  left: 50%;
  bottom: 20px;
  z-index: 6;
  min-width: min(180px, 52vw);
  height: 44px;
  padding: 0;
  border: 0;
  background: transparent;
  filter: drop-shadow(0 4px 2px rgba(49, 16, 6, 0.24));
  cursor: pointer;
  translate: -50% 0;
  transition: transform 160ms ease;
}

.guide-first__cta:active {
  transform: translateY(1px) scale(0.985);
}

.guide-first__cta > span {
  position: absolute;
  inset: 0;
  overflow: hidden;
  border-top: 1.5px solid #c8fffd;
  border-bottom: 1.5px solid #c8fffd;
  border-radius: 999px;
  background: linear-gradient(-61.4049deg, #78ffeb 15.063%, #2172ff 50.822%, #78ffeb 113.28%);
  box-shadow: inset 0 0 14px rgba(248, 200, 25, 0.24);
}

.guide-first__cta > span::before {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 88%;
  height: 36px;
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.5), rgba(255, 255, 255, 0) 45%);
  content: "";
  transform: translate(-50%, -50%);
}

.guide-first__cta strong {
  position: relative;
  z-index: 1;
  color: #fffffc;
  font-family: Montserrat, sans-serif;
  font-size: clamp(16px, 4.2vw, 18px);
  font-weight: 800;
  line-height: 1;
  text-shadow: 0 0 4px rgba(0, 0, 0, 0.16);
  text-transform: uppercase;
  white-space: nowrap;
}

.guide-first__hand {
  position: absolute;
  left: calc(50% + 72px);
  bottom: 2px;
  z-index: 7;
  width: 40px;
  pointer-events: none;
  transform-origin: 70% 78%;
  animation: guide-first-hand 420ms ease-in-out infinite;
}

.guide-invite__glow,
.guide-invite__rays {
  position: absolute;
  left: 50%;
  z-index: 0;
  width: 300px;
  height: 300px;
  pointer-events: none;
  transform: translateX(-50%);
}

.guide-invite__glow {
  top: 48px;
  opacity: 0.6;
  animation: guide-first-glow 2.8s ease-in-out infinite;
}

.guide-invite__rays {
  top: 0;
  opacity: 0.6;
  transform: translateX(-50%) rotate(67deg);
  animation: guide-invite-rays 8s linear infinite;
}

.guide-invite__chest-wrap {
  position: relative;
  z-index: 1;
  width: 320px;
  max-width: calc(100vw - 42px);
  margin-bottom: -58px;
  animation: guide-invite-chest-float 2.4s ease-in-out infinite;
}

.guide-invite__chest {
  display: block;
  width: 60%;
  height: auto;
  margin: 0 auto;
  pointer-events: none;
  user-select: none;
  filter: drop-shadow(0 10px 12px rgba(10, 22, 88, 0.24));
}

.guide-invite__banner {
  position: relative;
  z-index: 4;
  width: 308px;
  max-width: 96vw;
  height: 81px;
  margin-bottom: -40px;
  pointer-events: none;
}

.guide-invite__banner img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  user-select: none;
}

.guide-invite__banner svg {
  position: absolute;
  top: 40%;
  left: 50%;
  width: 232px;
  transform: translate(-50%, -50%);
}

.guide-invite__panel {
  position: relative;
  z-index: 3;
  width: 300px;
  max-width: calc(100vw - 72px);
  border: 4px solid rgba(255, 255, 255, 0.4);
  border-radius: 28px;
  padding: 40px 20px 24px;
  background: #ecf5ff;
  background-clip: padding-box;
  box-shadow: 0 0 43px rgba(253, 244, 224, 0.36);
}

.guide-invite__headline,
.guide-invite__hint {
  margin: 0;
  color: #3655cf;
  font-family: Montserrat, sans-serif;
  text-align: center;
  text-transform: capitalize;
}

.guide-invite__headline {
  font-size: 16px;
  font-weight: 800;
  line-height: 1.2;
}

.guide-invite__socials {
  display: grid;
  grid-template-columns: repeat(4, 50px);
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
}

.guide-invite__socials button {
  display: grid;
  width: 50px;
  height: 50px;
  place-items: center;
  overflow: hidden;
  border: 0;
  border-radius: 12px;
  padding: 0;
  background: transparent;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition:
    transform 150ms ease,
    filter 150ms ease;

  &:active {
    transform: scale(0.95);
  }
}

.guide-invite__socials img {
  display: block;
  width: 50px;
  height: 50px;
  object-fit: contain;
  user-select: none;
}

.guide-invite__socials .guide-invite__share {
  background: linear-gradient(180deg, #1cc2cc 0%, #12929a 100%);
}

.guide-invite__socials .guide-invite__share svg {
  width: 30px;
  height: 30px;
}

.guide-invite__hint {
  margin-top: 20px;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.2;
}

.guide-title {
  position: absolute;
  left: 50%;
  z-index: 5;
  width: 100%;
  transform: translateX(-50%);
  overflow: visible;
}

.guide-title--gift {
  top: -158px;
  width: 100%;
}

.guide-title__stroke,
.guide-title__fill {
  font-family: "Poetsen One", Montserrat, sans-serif;
  font-size: 21px;
  font-weight: 900;
  letter-spacing: 1px;
}

.guide-title__stroke {
  fill: none;
  stroke: #d42b1a;
  stroke-linejoin: round;
  stroke-width: 3px;
}

.guide-title__fill {
  fill: url("#guide-gift-title");
  filter: drop-shadow(0 4px 2px #8b1a00);
}

.guide-halo,
.guide-confetti {
  position: absolute;
  left: 50%;
  z-index: 1;
  pointer-events: none;
  object-fit: contain;
  transform: translateX(-50%);
}

.guide-halo {
  width: 100%;
  aspect-ratio: 994 / 708;
  animation: guide-halo-pulse 2.4s ease-in-out infinite;
}

.guide-confetti {
  width: 86%;
  aspect-ratio: 837 / 720;
  animation: guide-confetti-sway 2.8s ease-in-out infinite;
}

.guide-halo--gift {
  top: -130px;
}

.guide-confetti--gift {
  top: -104px;
}

.gift-panel {
  position: relative;
  z-index: 2;
  height: 286px;
  overflow: visible;
  border-radius: 58px;
  background: transparent;
}

.gift-panel::before {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 1;
  height: 204px;
  border-radius: 58px;
  background: linear-gradient(90deg, #fffbd2 0%, #ebfef2 100%);
  box-shadow: 0 0 12px rgba(251, 252, 226, 0.75);
  content: "";
  pointer-events: none;
}

.gift-panel__glass {
  position: absolute;
  bottom: 20px;
  left: 50%;
  z-index: 2;
  width: 286px;
  height: 270px;
  border-radius: 56px;
  background: rgba(247, 247, 234, 0.3);
  box-shadow: 0 0 8px rgba(255, 255, 239, 0.15);
  pointer-events: none;
  transform: translateX(-50%);
}

.gift-panel__hero {
  position: absolute;
  bottom: calc(100% - 58px);
  left: 50%;
  z-index: 7;
  width: 152px;
  transform: translateX(-50%);
  animation: guide-gift-drop 700ms cubic-bezier(0.16, 1, 0.3, 1) both;
}

.gift-panel__inner {
  position: absolute;
  top: -1px;
  right: 17px;
  left: 17px;
  z-index: 3;
  height: 252px;
  overflow: hidden;
  border-radius: 58px;
  background:
    linear-gradient(130deg, rgba(255, 235, 218, 0.65) 0 22%, transparent 23%),
    linear-gradient(140deg, #f73548 12%, #f53b76 66%, #e99a5e 100%);
  box-shadow:
    0 10px 18px rgba(255, 132, 102, 0.18),
    inset 0 0 18px rgba(255, 255, 255, 0.62);
}

.gift-panel__content {
  position: absolute;
  right: 22px;
  bottom: 145px;
  left: 22px;
  z-index: 5;
  display: grid;
  justify-items: center;
  gap: 23px;
}

.guide-amount {
  position: relative;
  display: inline-flex;
  align-items: flex-start;
  max-width: 100%;

  strong {
    overflow: hidden;
    color: #ffffff;
    font-size: 28px;
    font-weight: 900;
    line-height: 1;
    text-overflow: ellipsis;
    text-shadow: 0 2px 4px rgba(90, 0, 20, 0.34);
    white-space: nowrap;
  }

  span {
    position: absolute;
    top: -24px;
    right: auto;
    left: calc(100% - 32px);
    display: grid;
    min-width: 60px;
    height: 28px;
    place-items: center;
    padding: 0 9px 3px;
    border: 1px solid rgba(255, 255, 255, 0.5);
    border-radius: 999px;
    background: linear-gradient(180deg, #ffe069 0%, #ffb12c 100%);
    color: #7a3d00;
    font-size: 13px;
    font-weight: 800;
    line-height: 1;
    box-shadow: 0 2px 2px rgba(0, 0, 0, 0.15);

    &::after {
      position: absolute;
      right: 36px;
      bottom: -5px;
      width: 10px;
      height: 10px;
      background: #ffbd35;
      clip-path: polygon(50% 100%, 0 0, 100% 0);
      content: "";
    }
  }
}

.gift-panel__subtitle {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.84);
  font-size: 16px;
  line-height: 1;
  font-weight: 500;

  i {
    width: 28px;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.75));

    &:last-child {
      background: linear-gradient(90deg, rgba(255, 255, 255, 0.75), transparent);
    }
  }
}

.gift-panel__curve {
  position: absolute;
  bottom: 0;
  left: 0;
  z-index: 4;
  width: 100%;
  height: 145px;
  background: linear-gradient(
    90deg,
    rgba(255, 245, 208, 0.65) -0.3%,
    rgba(253, 193, 168, 0.65) 12.65%,
    rgba(255, 229, 193, 0.65) 49.6%,
    rgba(252, 231, 179, 0.65) 76.47%,
    rgba(254, 198, 194, 0.65) 99.5%
  );
  mask-image: url("data:image/svg+xml;utf8,%3Csvg%20xmlns%3D'http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg'%20viewBox%3D'0%200%20314%20145'%20preserveAspectRatio%3D'none'%3E%3Cpath%20d%3D'M313.039%2090.2871C313.039%20120.11%20288.862%20144.287%20259.039%20144.287H54C24.1767%20144.287%200%20120.11%200%2090.2871V13.3887C43.3631%2035.9763%2096.9854%2049.332%20155.031%2049.332C214.412%2049.332%20269.165%2035.3561%20313.039%2011.8164V90.2871Z'%20fill%3D'white'%2F%3E%3C%2Fsvg%3E");
  mask-repeat: no-repeat;
  mask-size: 100% 100%;
  -webkit-mask-image: url("data:image/svg+xml;utf8,%3Csvg%20xmlns%3D'http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg'%20viewBox%3D'0%200%20314%20145'%20preserveAspectRatio%3D'none'%3E%3Cpath%20d%3D'M313.039%2090.2871C313.039%20120.11%20288.862%20144.287%20259.039%20144.287H54C24.1767%20144.287%200%20120.11%200%2090.2871V13.3887C43.3631%2035.9763%2096.9854%2049.332%20155.031%2049.332C214.412%2049.332%20269.165%2035.3561%20313.039%2011.8164V90.2871Z'%20fill%3D'white'%2F%3E%3C%2Fsvg%3E");
  -webkit-mask-repeat: no-repeat;
  -webkit-mask-size: 100% 100%;
  pointer-events: none;
}

.gift-panel__curve::after {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 51px;
  background: linear-gradient(
    90deg,
    #fff5d0 0%,
    #fdc1a8 12.98%,
    #ffe5c1 50%,
    #fce7b3 76.92%,
    #f8e7d6 100%
  );
  content: "";
  mask-image: url("data:image/svg+xml;utf8,%3Csvg%20xmlns%3D'http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg'%20viewBox%3D'0%200%20313%2051'%20preserveAspectRatio%3D'none'%3E%3Cpath%20d%3D'M0%200C0%200%2061.6756%2037.5383%20156.432%2037.5383C251.188%2037.5383%20312.863%200%20312.863%200V23.2632C312.863%2023.2632%20251.749%2050.2273%20156.432%2050.2273C61.1149%2050.2273%200%2023.2632%200%2023.2632V0Z'%20fill%3D'white'%2F%3E%3C%2Fsvg%3E");
  mask-repeat: no-repeat;
  mask-size: 100% 100%;
  -webkit-mask-image: url("data:image/svg+xml;utf8,%3Csvg%20xmlns%3D'http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg'%20viewBox%3D'0%200%20313%2051'%20preserveAspectRatio%3D'none'%3E%3Cpath%20d%3D'M0%200C0%200%2061.6756%2037.5383%20156.432%2037.5383C251.188%2037.5383%20312.863%200%20312.863%200V23.2632C312.863%2023.2632%20251.749%2050.2273%20156.432%2050.2273C61.1149%2050.2273%200%2023.2632%200%2023.2632V0Z'%20fill%3D'white'%2F%3E%3C%2Fsvg%3E");
  -webkit-mask-repeat: no-repeat;
  -webkit-mask-size: 100% 100%;
}

.guide-cta {
  position: absolute;
  left: 50%;
  z-index: 6;
  display: grid;
  place-items: center;
  border-radius: 999px;
  color: #ffffff;
  font-weight: 700;
  line-height: 1;
  transform: translateX(-50%);
  transition:
    transform 180ms ease,
    filter 180ms ease;

  &:active {
    transform: translateX(-50%) translateY(1px) scale(0.985);
  }
}

.guide-cta--gift {
  bottom: 15px;
  width: 205px;
  height: 58px;
  border: 4px solid #ffdfc7;
  background: linear-gradient(180deg, #f73554 0%, #f35c34 100%);
  box-shadow:
    0 4px 0 rgba(0, 0, 0, 0.12),
    inset 0 2px 3px rgba(255, 255, 255, 0.16);
  font-size: 26px;
}

.download-title {
  position: absolute;
  top: -123px;
  left: 50%;
  z-index: 5;
  width: 325px;
  overflow: visible;
  transform: translateX(-50%);
}

.download-title__glow {
  fill: #f2f56c;
  opacity: 0.36;
}

.download-title__outer,
.download-title__stroke,
.download-title__fill {
  font-family: "Poetsen One", Montserrat, sans-serif;
  font-size: 24px;
  font-weight: 900;
  letter-spacing: 0.5px;
}

.download-title__outer {
  fill: none;
  stroke: #0c62b6;
  stroke-linejoin: round;
  stroke-width: 6px;
}

.download-title__stroke {
  fill: none;
  stroke: #b9f6ff;
  stroke-linejoin: round;
  stroke-width: 3px;
}

.download-title__fill {
  fill: url("#guide-download-fill");
  filter:
    drop-shadow(0 3px 0 rgba(6, 72, 134, 0.75))
    drop-shadow(0 8px 14px rgba(0, 0, 0, 0.3));
}

.guide-halo--download {
  top: -124px;
  filter: hue-rotate(-150deg);
}

.guide-confetti--download {
  top: -115px;
  filter: hue-rotate(-150deg);
}

.download-hero {
  position: absolute;
  top: -51px;
  left: 50%;
  z-index: 6;
  width: 151px;
  transform: translateX(-50%);
  animation: guide-hero-float 2s ease-in-out infinite;
}

.download-panel {
  position: relative;
  z-index: 2;
  height: 315px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.36);
  border-radius: 64px;
  background: rgba(255, 255, 255, 0.05);
  box-shadow: inset 0 0 18px rgba(255, 255, 255, 0.62);
}

.download-panel__shine {
  position: absolute;
  inset: 9px 9px 12px;
  z-index: 1;
  overflow: hidden;
  border-radius: 56px;
  background:
    linear-gradient(130deg, rgba(218, 255, 251, 0.72) 0 18%, transparent 19%),
    linear-gradient(135deg, #3f9ef7 12%, #3fbbf9 66%, #76f1fa 100%);
  box-shadow: inset 0 0 18px rgba(255, 255, 255, 0.62);

  &::before,
  &::after {
    position: absolute;
    content: "";
    filter: blur(34px);
    transform: rotate(42deg);
  }

  &::before {
    top: -105px;
    left: -6px;
    width: 132px;
    height: 244px;
    background: rgba(116, 252, 209, 0.48);
  }

  &::after {
    right: -45px;
    bottom: -64px;
    width: 44px;
    height: 244px;
    background: rgba(121, 255, 204, 0.36);
  }
}

.download-panel__content {
  position: absolute;
  top: 86px;
  right: 18px;
  left: 18px;
  z-index: 4;
  display: grid;
  justify-items: center;
  gap: 15px;

  p {
    width: 250px;
    margin: 0;
    color: rgba(255, 255, 255, 0.86);
    font-size: 16px;
    font-weight: 600;
    line-height: 1.25;
    text-align: center;
    text-shadow: 0 2px 4px rgba(0, 40, 80, 0.36);
  }
}

.guide-amount--download {
  strong {
    text-shadow: 0 2px 4px rgba(0, 40, 80, 0.4);
  }

  span {
    min-width: 64px;
    text-transform: uppercase;
  }
}

.download-panel__curve {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 3;
  height: 118px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.9) 0%, rgba(255, 255, 255, 0.7) 100%);
  backdrop-filter: blur(4px);
  mask-image: url("data:image/svg+xml;utf8,%3Csvg%20xmlns%3D'http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg'%20viewBox%3D'0%200%20299%20166'%20preserveAspectRatio%3D'none'%3E%3Cpath%20d%3D'M299%20101.4C299%20136.747%20270.346%20165.4%20235%20165.4H64C28.6538%20165.4%202.57738e-07%20136.747%200%20101.4V0.124023C41.6377%2022.1247%2093.3527%2035.1582%20149.384%2035.1582C205.52%2035.1581%20257.322%2022.0754%20299%200V101.4Z'%20fill%3D'white'%2F%3E%3C%2Fsvg%3E");
  mask-repeat: no-repeat;
  mask-size: 100% 100%;
  transform: translateZ(0);
  will-change: backdrop-filter;
  -webkit-mask-image: url("data:image/svg+xml;utf8,%3Csvg%20xmlns%3D'http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg'%20viewBox%3D'0%200%20299%20166'%20preserveAspectRatio%3D'none'%3E%3Cpath%20d%3D'M299%20101.4C299%20136.747%20270.346%20165.4%20235%20165.4H64C28.6538%20165.4%202.57738e-07%20136.747%200%20101.4V0.124023C41.6377%2022.1247%2093.3527%2035.1582%20149.384%2035.1582C205.52%2035.1581%20257.322%2022.0754%20299%200V101.4Z'%20fill%3D'white'%2F%3E%3C%2Fsvg%3E");
  -webkit-mask-repeat: no-repeat;
  -webkit-mask-size: 100% 100%;
  pointer-events: none;
}

.guide-cta--download {
  bottom: 32px;
  width: 228px;
  height: 40px;
  background: linear-gradient(183.55deg, #20c5ff 14%, #007df1 88%);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.18);
  font-size: 16px;
}

.guide-close {
  position: absolute;
  top: -43px;
  right: -14px;
  z-index: 8;
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  padding: 0;

  &::before {
    position: absolute;
    inset: 5px;
    border: 2px solid #ffffff;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.1);
    content: "";
  }

  span {
    position: relative;
    width: 20px;
    height: 20px;

    &::before,
    &::after {
      position: absolute;
      top: 9px;
      left: -1px;
      width: 22px;
      height: 3px;
      border-radius: 999px;
      background: #ffffff;
      content: "";
    }

    &::before {
      transform: rotate(45deg);
    }

    &::after {
      transform: rotate(-45deg);
    }
  }
}

.guide-close.guide-close--first {
  position: relative;
  top: auto;
  right: auto;
  bottom: auto;
  width: 40px;
  height: 40px;
  margin-top: 24px;
  transform: none;
}

.guide-close.guide-close--first::before {
  inset: 4px;
  border-width: 1px;
}

.guide-close.guide-close--invite {
  top: 52px;
  right: -8px;
  z-index: 8;
}

.guide-overlay-enter-active,
.guide-overlay-leave-active {
  transition: opacity 220ms ease;
}

.guide-overlay-enter-from,
.guide-overlay-leave-to {
  opacity: 0;
}

.guide-dialog-enter-active {
  transition:
    opacity 260ms ease,
    transform 360ms cubic-bezier(0.16, 1, 0.3, 1);
}

.guide-dialog-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.guide-dialog-enter-from {
  opacity: 0;
  transform: translateY(28px) scale(0.94);
}

.guide-dialog-leave-to {
  opacity: 0;
  transform: translateY(-12px) scale(0.98);
}

@keyframes guide-halo-pulse {
  0% {
    opacity: 0.42;
    transform: translateX(-50%) scale(0.99);
  }

  18% {
    opacity: 1;
    transform: translateX(-50%) scale(1.04);
  }

  36% {
    opacity: 0.58;
    transform: translateX(-50%) scale(1.01);
  }

  54% {
    opacity: 0.96;
    transform: translateX(-50%) scale(1.035);
  }

  78%,
  100% {
    opacity: 0.68;
    transform: translateX(-50%) scale(1);
  }
}

@keyframes guide-confetti-sway {
  0%,
  100% {
    transform: translateX(-50%) rotate(-3deg);
  }

  50% {
    transform: translateX(-50%) translateY(-5px) rotate(3deg);
  }
}

@keyframes guide-gift-drop {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(18px) scale(0.78) rotate(-12deg);
  }

  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0) scale(1) rotate(0);
  }
}

@keyframes guide-hero-float {
  0%,
  100% {
    transform: translateX(-50%) translateY(0);
  }

  50% {
    transform: translateX(-50%) translateY(-7px);
  }
}

@keyframes guide-first-glow {
  0%,
  100% {
    opacity: 0.62;
    transform: translateX(-50%) scale(0.98);
  }

  50% {
    opacity: 0.9;
    transform: translateX(-50%) scale(1.04);
  }
}

@keyframes guide-first-top-float {
  0%,
  100% {
    transform: translateY(0);
  }

  50% {
    transform: translateY(2px);
  }
}

@keyframes guide-first-title {
  0%,
  100% {
    transform: translateX(-50%) translateY(0);
  }

  50% {
    transform: translateX(-50%) translateY(-3px);
  }
}

@keyframes guide-first-treasure-float {
  0%,
  100% {
    transform: translateY(0);
  }

  50% {
    transform: translateY(-5px);
  }
}

@keyframes guide-first-hand {
  0%,
  100% {
    transform: translateY(0) scale(0.96);
  }

  50% {
    transform: translateY(3px) scale(0.84);
  }
}

@keyframes guide-invite-rays {
  from {
    transform: translateX(-50%) rotate(67deg);
  }

  to {
    transform: translateX(-50%) rotate(427deg);
  }
}

@keyframes guide-invite-chest-float {
  0%,
  100% {
    transform: translateY(0);
  }

  50% {
    transform: translateY(-5px);
  }
}

@media (max-width: 374px) {
  .guide-card {
    width: min(calc(100vw - 42px), 316px);
  }

  .guide-card--download {
    width: min(calc(100vw - 62px), 298px);
  }

  .guide-card--first {
    width: 100vw;
  }

  .guide-first__treasure-wrap {
    width: min(340px, 96vw);
    margin-top: -212px;
  }

  .guide-first__cta {
    bottom: 19px;
  }

  .guide-first__hand {
    left: calc(50% + 65px);
    bottom: 2px;
  }

  .guide-card--invite {
    width: min(calc(100vw - 48px), 308px);
  }

  .guide-invite__socials {
    grid-template-columns: repeat(4, 46px);
    gap: 14px;
  }

  .guide-invite__socials button,
  .guide-invite__socials img {
    width: 46px;
    height: 46px;
  }
}
</style>
