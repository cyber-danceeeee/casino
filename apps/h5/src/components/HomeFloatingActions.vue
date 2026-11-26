<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { localDialogImage, localHomeImage } from '../utils/localAssets'
import rewardCoins from '../assets/images/home/reward-float/coins.webp'
import rewardGlow from '../assets/images/home/reward-float/glow.webp'
import rewardSparkles from '../assets/images/home/reward-float/sparkles.webp'
import { isAndroidAppRuntime } from '../utils/runtimeEnvironment'

const props = withDefaults(defineProps<{
  unreadActivitySlugs?: string[]
  pendingRewardCount?: number
}>(), {
  unreadActivitySlugs: () => [],
  pendingRewardCount: 0,
})

const emit = defineEmits<{
  events: []
  incentive: []
  rewards: []
  activity: [slug: string]
}>()

const router = useRouter()
const androidAppRuntime = isAndroidAppRuntime()
const treasure = localHomeImage('float-entry-treasure.webp')
const bottle = localHomeImage('incentive-bottle-90.webp')
const incentiveBanner = localHomeImage('incentive-dialog-banner.webp')
const incentiveDroplet = localHomeImage('incentive-droplet.webp')
const incentiveGlow = localDialogImage('activity__guide-flow-dialogs__bg_glow.webp')
const incentiveRays = localDialogImage('bonus__rays.webp')
const bottleMissing = ref(false)
const expanded = ref(false)
const incentiveAnimating = ref(false)
const incentiveDialogVisible = ref(false)
const floatingY = ref(253)
const floatingDragging = ref(false)
const floatingSideHidden = ref(false)
const suppressNextClick = ref(false)
let incentiveTimer: number | undefined
let scrollIdleTimer: number | undefined
let dragState:
  | {
      pointerId: number
      startX: number
      startY: number
      originY: number
      minY: number
      maxY: number
      moved: boolean
    }
  | undefined

const activityEntries = [
  {
    label: 'First Deposit Rewards',
    image: localHomeImage('float-entry/first-deposit.webp'),
    slug: 'first-deposit-bonus',
    iconClass: 'is-first-deposit',
  },
  {
    label: 'Deposit Bonus',
    image: localHomeImage('float-entry/deposit-bonus.webp'),
    slug: 'deposit-bonus',
    iconClass: 'is-deposit-bonus',
  },
  {
    label: 'Daily Cashback',
    image: localHomeImage('float-entry/daily-cashback.webp'),
    slug: 'daily-deposit-cashback',
    iconClass: 'is-daily-cashback',
  },
  {
    label: 'Rewards Card',
    image: localHomeImage('float-entry/rewards-card.webp'),
    slug: 'welfare-card',
    iconClass: 'is-rewards-card',
  },
  {
    label: 'Lucky Wheel',
    image: localHomeImage('float-entry/lucky-wheel.webp'),
    slug: 'invitation-wheel',
    iconClass: 'is-lucky-wheel',
  },
  {
    label: 'Invite Chest',
    image: localHomeImage('float-entry/chest-pro.webp'),
    slug: 'invitation-chest',
    iconClass: 'is-invite-chest',
  },
  {
    label: 'Referral',
    image: localHomeImage('float-entry/referral.webp'),
    slug: 'sub-deposit-reward',
    iconClass: 'is-referral',
  },
  {
    label: 'Weekly Reward',
    image: localHomeImage('float-entry/weekly-reward.webp'),
    slug: 'periodic-deposit-reward/weekly',
    iconClass: 'is-weekly-reward',
  },
  {
    label: 'Daily Sign In',
    image: localHomeImage('float-entry/daily-sign-in.webp'),
    slug: 'daily-sign-in',
    iconClass: 'is-daily-signin',
  },
  {
    label: 'Red Packet Rain',
    image: localHomeImage('float-entry/red-packet-rain.webp'),
    slug: 'red-packet-rain',
    iconClass: 'is-red-packet',
  },
  {
    label: 'Cash Back',
    image: localHomeImage('float-entry/cash-back.webp'),
    slug: 'daily-relief',
    iconClass: 'is-cash-back',
  },
  {
    label: 'Rebate Leaderboard',
    image: localHomeImage('float-entry/rebate-rank.webp'),
    slug: 'rebate-ranking',
    iconClass: 'is-rebate-rank',
  },
  {
    label: 'Phone Verification',
    image: localHomeImage('float-entry/phone-verification.webp'),
    slug: 'phone-verification',
    iconClass: 'is-phone-verification',
  },
  {
    label: 'Download App',
    image: localHomeImage('float-entry/download-app.webp'),
    slug: 'app-download',
    iconClass: 'is-download-app',
  },
] as const

const unreadActivitySet = computed(() => new Set(props.unreadActivitySlugs))
const visibleActivityEntries = computed(() =>
  androidAppRuntime
    ? activityEntries.filter((entry) => entry.slug !== 'app-download')
    : activityEntries,
)

const floatingTransform = computed(() => {
  const sideOffset = floatingSideHidden.value && !floatingDragging.value ? 42 : 0
  return `translate3d(${sideOffset}px, ${floatingY.value}px, 0)`
})

const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max)

const toggleEvents = () => {
  expanded.value = !expanded.value
}

const closeEvents = () => {
  expanded.value = false
}

const openActivity = (slug: string) => {
  expanded.value = false
  emit('activity', slug)
  void router.push({ path: `/activity/${slug}` })
}

const openIncentiveDialog = () => {
  if (incentiveAnimating.value) return
  expanded.value = false
  incentiveAnimating.value = true
  window.clearTimeout(incentiveTimer)
  incentiveTimer = window.setTimeout(() => {
    incentiveAnimating.value = false
    incentiveDialogVisible.value = true
  }, 520)
}

const openRewardsDialog = () => {
  expanded.value = false
  emit('rewards')
}

const closeIncentiveDialog = () => {
  incentiveDialogVisible.value = false
}

const revealFloatingActions = () => {
  floatingSideHidden.value = false
}

const onWindowScroll = () => {
  if (floatingDragging.value || incentiveDialogVisible.value) return
  expanded.value = false
  floatingSideHidden.value = true
  window.clearTimeout(scrollIdleTimer)
  scrollIdleTimer = window.setTimeout(revealFloatingActions, 520)
}

const stopFloatingDrag = () => {
  window.removeEventListener('pointermove', moveFloatingActions)
  window.removeEventListener('pointerup', stopFloatingDrag)
  window.removeEventListener('pointercancel', stopFloatingDrag)
  if (dragState?.moved) {
    suppressNextClick.value = true
  }
  dragState = undefined
  floatingDragging.value = false
}

const moveFloatingActions = (event: PointerEvent) => {
  if (!dragState || event.pointerId !== dragState.pointerId) return
  const deltaY = event.clientY - dragState.startY
  const deltaX = event.clientX - dragState.startX

  if (!dragState.moved && Math.hypot(deltaX, deltaY) > 4) {
    dragState.moved = true
    floatingDragging.value = true
    expanded.value = false
  }

  if (!dragState.moved) return
  event.preventDefault()
  floatingY.value = clamp(dragState.originY + deltaY, dragState.minY, dragState.maxY)
}

const startFloatingDrag = (event: PointerEvent) => {
  if (event.pointerType === 'mouse' && event.button !== 0) return
  const target = event.target as HTMLElement | null
  if (target?.closest('.incentive-dialog, .home-events-panel')) return
  window.clearTimeout(scrollIdleTimer)
  floatingSideHidden.value = false
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  dragState = {
    pointerId: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    originY: floatingY.value,
    minY: floatingY.value - rect.top + 8,
    maxY: floatingY.value + window.innerHeight - rect.bottom - 8,
    moved: false,
  }
  window.addEventListener('pointermove', moveFloatingActions, { passive: false })
  window.addEventListener('pointerup', stopFloatingDrag)
  window.addEventListener('pointercancel', stopFloatingDrag)
}

const blockClickAfterDrag = (event: MouseEvent) => {
  if (!suppressNextClick.value) return
  event.preventDefault()
  event.stopPropagation()
  suppressNextClick.value = false
}

onMounted(() => {
  window.addEventListener('scroll', onWindowScroll, { passive: true })
})

onBeforeUnmount(() => {
  window.clearTimeout(incentiveTimer)
  window.clearTimeout(scrollIdleTimer)
  window.removeEventListener('scroll', onWindowScroll)
  stopFloatingDrag()
})
</script>

<template>
  <div
    class="home-floating-actions"
    :class="{ 'is-expanded': expanded, 'is-dragging': floatingDragging }"
    :style="{ transform: floatingTransform }"
    aria-label="Quick entries"
    @pointerdown="startFloatingDrag"
    @click.capture="blockClickAfterDrag"
  >
    <Transition name="home-events-panel">
      <section v-if="expanded" class="home-events-panel" aria-label="Event shortcuts">
        <button type="button" class="home-events-panel__close" aria-label="Close quick entries" @click="closeEvents">
          <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
            <path d="M5 5l10 10M15 5 5 15" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          </svg>
        </button>
        <div class="home-events-panel__grid">
          <button
            v-for="entry in visibleActivityEntries"
            :key="entry.slug"
            type="button"
            class="home-events-panel__item"
            :class="entry.iconClass"
            @click="openActivity(entry.slug)"
          >
            <img :src="entry.image" :alt="entry.label" loading="eager" decoding="async" />
            <span>{{ entry.label }}</span>
            <i v-if="unreadActivitySet.has(entry.slug)" aria-hidden="true" />
          </button>
        </div>
      </section>
    </Transition>

    <div class="home-float-entry-wrap" :class="{ 'is-hidden': expanded }">
      <button type="button" class="home-float-entry" aria-label="Quick entry" @click="toggleEvents">
        <svg
          preserveAspectRatio="none"
          width="100%"
          height="100%"
          overflow="visible"
          viewBox="0 0 60 63"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          class="home-float-entry__shell"
          aria-hidden="true"
        >
          <foreignObject x="0" y="0" width="60" height="63">
            <div class="home-float-entry__blur" xmlns="http://www.w3.org/1999/xhtml" />
          </foreignObject>
          <g filter="url(#homeFloatEntryShadow)">
            <circle cx="30" cy="30" r="30" fill="#090F1E" />
            <circle cx="30" cy="30" r="30" fill="#002A62" fill-opacity="0.8" />
            <circle cx="30" cy="30" r="29" stroke="url(#homeFloatEntryBorder)" stroke-width="2" />
          </g>
          <defs>
            <filter id="homeFloatEntryShadow" x="0" y="0" width="60" height="63" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
              <feFlood flood-opacity="0" result="BackgroundImageFix" />
              <feColorMatrix in="SourceAlpha" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha" />
              <feOffset dy="3" />
              <feComposite in2="hardAlpha" operator="out" />
              <feColorMatrix type="matrix" values="0 0 0 0 0.084 0 0 0 0 0.184 0 0 0 0 0.434 0 0 0 1 0" />
              <feBlend mode="normal" in2="BackgroundImageFix" result="drop" />
              <feBlend mode="normal" in="SourceGraphic" in2="drop" result="shape" />
              <feColorMatrix in="SourceAlpha" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha" />
              <feGaussianBlur stdDeviation="2" />
              <feComposite in2="hardAlpha" operator="arithmetic" k2="-1" k3="1" />
              <feColorMatrix type="matrix" values="0 0 0 0 0 0 0 0 0 0.119 0 0 0 0 0.322 0 0 0 1 0" />
              <feBlend mode="normal" in2="shape" />
            </filter>
            <linearGradient id="homeFloatEntryBorder" x1="30" y1="60" x2="30" y2="0" gradientUnits="userSpaceOnUse">
              <stop stop-color="#5AEBFF" />
              <stop offset="0.5" stop-color="#2996FF" />
              <stop offset="1" stop-color="#00B4F6" />
            </linearGradient>
          </defs>
        </svg>
        <svg
          preserveAspectRatio="none"
          width="100%"
          height="100%"
          overflow="visible"
          viewBox="0 0 54 30"
          fill="none"
          class="home-float-entry__shine"
          aria-hidden="true"
        >
          <path
            d="M27 0.8C41.37 0.8 53 12.45 53 26.82c0 .47-.01.94-.04 1.4C52.23 14.66 40.9 3.88 27 3.88 13.1 3.88 1.77 14.66 1.04 28.22 1.01 27.76 1 27.29 1 26.82 1 12.45 12.63.8 27 .8Z"
            fill="white"
            fill-opacity="0.16"
          />
        </svg>
        <div class="home-float-entry__icon">
          <img :src="treasure" alt="" loading="eager" decoding="sync" />
        </div>
        <div class="home-float-entry__label">
          <span>Events</span>
        </div>
        <span v-if="unreadActivitySlugs.length" class="home-float-entry__badge" aria-hidden="true">
          <span />
        </span>
      </button>
    </div>

    <button
      v-if="pendingRewardCount > 0"
      type="button"
      class="home-reward-entry"
      :class="{ 'is-hidden': expanded }"
      aria-label="Rewards"
      @click="openRewardsDialog"
    >
      <span class="home-reward-entry__shell" aria-hidden="true" />
      <img class="home-reward-entry__glow" :src="rewardGlow" alt="" aria-hidden="true" />
      <img class="home-reward-entry__sparkles" :src="rewardSparkles" alt="" aria-hidden="true" />
      <img class="home-reward-entry__coins" :src="rewardCoins" alt="" aria-hidden="true" />
      <span class="home-reward-entry__label">Rewards</span>
      <span class="home-reward-entry__count">{{ pendingRewardCount > 99 ? '99+' : pendingRewardCount }}</span>
      <span class="home-reward-entry__badge" aria-hidden="true"><span /></span>
    </button>

    <button
      type="button"
      class="home-incentive"
      :class="{ 'is-boosting': incentiveAnimating }"
      aria-label="Incentive Points"
      @click="openIncentiveDialog"
    >
      <span class="home-incentive__base" aria-hidden="true" />
      <svg viewBox="0 0 64 64" fill="none" class="home-incentive__ring" aria-hidden="true">
        <defs>
          <radialGradient id="homeIncentiveProgressGlow" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(32 32) rotate(90) scale(27)">
            <stop offset="0.83" stop-color="#75E8F0" />
            <stop offset="1" stop-color="#00C3EA" />
          </radialGradient>
          <linearGradient id="homeIncentiveProgressTrack" x1="32" y1="9" x2="32" y2="55" gradientUnits="userSpaceOnUse">
            <stop stop-color="#0A2255" />
            <stop offset="1" stop-color="#0F3D7A" />
          </linearGradient>
          <filter id="homeIncentiveTrackInset">
            <feOffset dx="1" dy="1" />
            <feGaussianBlur stdDeviation="1" />
            <feComposite in2="SourceAlpha" operator="arithmetic" k2="-1" k3="1" />
            <feColorMatrix values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.1 0" />
            <feBlend in2="SourceGraphic" />
          </filter>
        </defs>
        <circle cx="32" cy="32" r="27" stroke="url(#homeIncentiveProgressTrack)" stroke-width="5" fill="none" filter="url(#homeIncentiveTrackInset)" />
        <circle
          class="home-incentive__progress"
          cx="32"
          cy="32"
          r="27"
          stroke="url(#homeIncentiveProgressGlow)"
          stroke-width="5"
          stroke-linecap="round"
          fill="none"
          stroke-dasharray="169.646"
          stroke-dashoffset="16.9646"
          transform="rotate(-90 32 32)"
        />
      </svg>
      <span class="home-incentive__sparkles" aria-hidden="true">
        <span class="home-incentive__sparkles-inner">
          <i v-for="index in 8" :key="index" />
        </span>
      </span>
      <span class="home-incentive__icon">
        <img
          v-if="!bottleMissing"
          :src="bottle"
          alt=""
          loading="eager"
          decoding="sync"
          @error="bottleMissing = true"
        />
        <span v-else class="home-incentive__fallback" aria-hidden="true" />
      </span>
    </button>
  </div>

  <Teleport to="body">
    <Transition name="incentive-dialog">
      <section
        v-if="incentiveDialogVisible"
        class="incentive-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="incentive-dialog-title"
      >
        <img class="incentive-dialog__glow" :src="incentiveGlow" alt="" aria-hidden="true" />
        <img class="incentive-dialog__rays" :src="incentiveRays" alt="" aria-hidden="true" />
        <div class="incentive-dialog__wrap">
          <div class="incentive-dialog__meter" aria-hidden="true">
            <svg viewBox="0 0 64 64" fill="none" class="incentive-dialog__meter-ring">
              <defs>
                <radialGradient id="incentiveDialogProgressGlow" cx="0" cy="0" r="1" gradientUnits="userSpaceOnUse" gradientTransform="translate(32 32) rotate(90) scale(30.25)">
                  <stop offset="0.83" stop-color="#75E8F0" />
                  <stop offset="1" stop-color="#00C3EA" />
                </radialGradient>
                <linearGradient id="incentiveDialogProgressTrack" x1="32" y1="9" x2="32" y2="55" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#0A2255" />
                  <stop offset="1" stop-color="#0F3D7A" />
                </linearGradient>
                <filter id="incentiveDialogTrackInset">
                  <feOffset dx="1" dy="1" />
                  <feGaussianBlur stdDeviation="1" />
                  <feComposite in2="SourceAlpha" operator="arithmetic" k2="-1" k3="1" />
                  <feColorMatrix values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.1 0" />
                  <feBlend in2="SourceGraphic" />
                </filter>
              </defs>
              <circle cx="32" cy="32" r="30.25" stroke="url(#incentiveDialogProgressTrack)" stroke-width="3.5" fill="none" filter="url(#incentiveDialogTrackInset)" />
              <circle cx="32" cy="32" r="30.25" stroke="url(#incentiveDialogProgressGlow)" stroke-width="3.5" stroke-linecap="round" fill="none" stroke-dasharray="190.066" stroke-dashoffset="19.0066" transform="rotate(-90 32 32)" />
            </svg>
            <span class="incentive-dialog__meter-sparkles">
              <span>
                <i v-for="index in 8" :key="index" />
              </span>
            </span>
            <img class="incentive-dialog__bottle" :src="bottle" alt="" />
          </div>

          <article class="incentive-dialog__card">
            <div class="incentive-dialog__banner">
              <img :src="incentiveBanner" alt="" aria-hidden="true" />
              <svg id="incentive-dialog-title" viewBox="0 0 260 40" preserveAspectRatio="xMidYMid meet" fill="none" aria-label="Points Details">
                <defs>
                  <linearGradient id="incentiveDialogTitleFill" x1="130" y1="0" x2="130" y2="40" gradientUnits="userSpaceOnUse">
                    <stop offset="0" stop-color="#FFFFFF" />
                    <stop offset="0.55" stop-color="#FFF5C4" />
                    <stop offset="1" stop-color="#FFD940" />
                  </linearGradient>
                  <path id="incentiveDialogTitlePath" d="M 0 30.96 Q 130 22.96 260 30.96" />
                </defs>
                <text
                  text-anchor="middle"
                  font-family="Poetsen One, Montserrat, sans-serif"
                  font-weight="400"
                  font-style="italic"
                  font-size="28"
                  letter-spacing="0.5"
                  fill="url(#incentiveDialogTitleFill)"
                  stroke="#7A1D0C"
                  stroke-width="4"
                  stroke-linejoin="round"
                  paint-order="stroke"
                >
                  <textPath href="#incentiveDialogTitlePath" startOffset="50%">Points Details</textPath>
                </text>
              </svg>
            </div>
            <button type="button" aria-label="Close" class="incentive-dialog__close" @click="closeIncentiveDialog">
              <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path d="M6 6l12 12M18 6 6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
              </svg>
            </button>
            <div class="incentive-dialog__panel">
              <div class="incentive-dialog__panel-inner">
                <div class="incentive-dialog__dots" aria-hidden="true">
                  <svg
                    v-for="index in 73"
                    :key="index"
                    viewBox="0 0 8 8"
                    class="incentive-dialog__dot-svg"
                    fill="none"
                    :style="{ '--delay': `${(index - 1) * -0.0493151}s` }"
                  >
                    <defs>
                      <filter :id="`incentiveDotGlow${index}`" x="-100%" y="-100%" width="300%" height="300%">
                        <feGaussianBlur in="SourceAlpha" stdDeviation="0.8" />
                        <feFlood flood-color="#013874" />
                        <feComposite in2="SourceAlpha" operator="in" result="outerGlow" />
                        <feMerge>
                          <feMergeNode in="outerGlow" />
                          <feMergeNode in="SourceGraphic" />
                        </feMerge>
                      </filter>
                    </defs>
                    <circle cx="4" cy="4" r="4" fill="white" :filter="`url(#incentiveDotGlow${index})`" />
                    <circle cx="4" cy="4" r="2.6" fill="none" stroke="#59FF98" stroke-width="0.5" opacity="0.85" />
                  </svg>
                </div>
                <div class="incentive-dialog__meter-space" aria-hidden="true" />
                <p class="incentive-dialog__score">90 / 100 for Surprise</p>
                <div class="incentive-dialog__tips">
                  <div class="incentive-dialog__tips-track">
                    <div class="incentive-dialog__tip">
                      <div class="incentive-dialog__mock-top" />
                      <div class="incentive-dialog__mock-body">
                        <p>Clear the red dot <span class="incentive-dialog__red-dot" /></p>
                        <p>Collect points <img :src="incentiveDroplet" alt="" /></p>
                      </div>
                    </div>
                    <div class="incentive-dialog__tip">
                      <div class="incentive-dialog__mock-pills">
                        <span />
                        <span />
                        <span />
                      </div>
                      <div class="incentive-dialog__mock-body">
                        <p>Clear the red dot <span class="incentive-dialog__red-dot" /></p>
                        <p>Collect points <img :src="incentiveDroplet" alt="" /></p>
                      </div>
                    </div>
                    <div class="incentive-dialog__tip">
                      <div class="incentive-dialog__mock-bars">
                        <span />
                        <span />
                      </div>
                      <div class="incentive-dialog__mock-body">
                        <p>Clear the red dot <span class="incentive-dialog__red-dot" /></p>
                        <p>Collect points <img :src="incentiveDroplet" alt="" /></p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <button type="button" class="incentive-dialog__ok" @click="closeIncentiveDialog">
              <span class="incentive-dialog__ok-base" aria-hidden="true" />
              <span class="incentive-dialog__ok-top" aria-hidden="true" />
              <span class="incentive-dialog__ok-shine" aria-hidden="true" />
              <span class="incentive-dialog__ok-label">Okay</span>
            </button>
          </article>
        </div>
      </section>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.home-floating-actions {
  position: fixed;
  top: 0;
  right: max(calc((100vw - 430px) / 2 + 12px), 12px);
  z-index: 50;
  display: flex;
  width: 60px;
  touch-action: none;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  transition: transform 300ms ease;
  user-select: none;
  will-change: transform;
}

.home-floating-actions.is-dragging {
  transition: none;
}

.home-floating-actions.is-expanded {
  touch-action: pan-y;
}

.home-events-panel {
  position: absolute;
  right: 0;
  top: 0;
  width: 150px;
  height: 266px;
  max-height: min(266px, calc(100vh - 116px));
  overflow: visible;
  border: 1.5px solid #2996ff;
  border-left-color: transparent;
  border-radius: 14px;
  padding: 8px 8px 0;
  background:
    linear-gradient(rgba(0, 42, 98, 0.8), rgba(0, 42, 98, 0.8)),
    #090f1e;
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.36),
    inset 0 1px 0 rgba(255, 255, 255, 0.12);

  &::before {
    position: absolute;
    top: 4px;
    right: 4px;
    left: 4px;
    height: 68px;
    border-radius: 10px;
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.16), rgba(255, 255, 255, 0));
    content: "";
    pointer-events: none;
  }

  &::after {
    position: absolute;
    top: 21px;
    left: -1.5px;
    width: 1.5px;
    height: 37px;
    background:
      linear-gradient(rgba(0, 42, 98, 0.8), rgba(0, 42, 98, 0.8)),
      #090f1e;
    content: "";
    pointer-events: none;
  }
}

.home-events-panel__close {
  position: absolute;
  top: 21.7px;
  left: -24px;
  z-index: 2;
  display: grid;
  width: 24px;
  height: 36px;
  place-items: center;
  border: 1.5px solid #2996ff;
  border-right: 0;
  border-radius: 12px 0 0 12px;
  background:
    linear-gradient(rgba(0, 42, 98, 0.8), rgba(0, 42, 98, 0.8)),
    #090f1e;
  color: #d2f1ff;

  svg {
    width: 20px;
    height: 20px;

    path {
      stroke-width: 2.4;
    }
  }
}

.home-events-panel__grid {
  display: grid;
  grid-template-columns: repeat(2, 67px);
  gap: 0;
  height: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: none;
  touch-action: pan-y;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar {
    display: none;
  }

  > *:nth-child(n + 3) {
    margin-top: -6px;
  }
}

.home-events-panel__item {
  position: relative;
  display: block;
  min-width: 0;
  width: 67px;
  height: 67px;
  overflow: visible;
  border: 0;
  padding: 0;
  background: transparent;
  transition: transform 100ms ease-out;

  &::before {
    position: absolute;
    top: 8px;
    left: 8px;
    width: 51px;
    height: 51px;
    border: 1.5px solid #48b6ff;
    border-radius: 9px;
    background: linear-gradient(180deg, #326ed5 0%, #143f88 100%);
    box-shadow:
      inset 0 0 2px rgba(0, 0, 0, 0.28),
      inset 0 -0.5px 0 rgba(255, 255, 255, 0.18),
      0 1px 0 rgba(0, 0, 0, 0.25);
    content: "";
  }

  &:active {
    transform: scale(0.92);
  }

  img {
    position: absolute;
    top: 50%;
    left: 50%;
    z-index: 1;
    width: var(--icon-w, 54px);
    height: var(--icon-h, 54px);
    object-fit: contain;
    transform: translate(-50%, -50%) translate(var(--icon-x, 0), var(--icon-y, 0));
    filter: drop-shadow(0 4px 4px rgba(0, 0, 0, 0.32));
  }

  span {
    position: absolute;
    right: 0;
    bottom: 6px;
    left: 0;
    z-index: 1;
    opacity: 1;
    color: #ffffff;
    font-family: "Poetsen One", Montserrat, sans-serif;
    font-size: 10px;
    font-weight: 400;
    line-height: 0.9;
    text-align: center;
    text-transform: capitalize;
    -webkit-text-stroke: 3px #062f75;
    paint-order: stroke fill;
  }

  i {
    position: absolute;
    top: 8px;
    right: 6px;
    z-index: 2;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #ed252b;

    &::before {
      position: absolute;
      inset: 0;
      border-radius: inherit;
      background: inherit;
      animation: home-badge-ping 1.45s cubic-bezier(0, 0, 0.2, 1) infinite;
      content: "";
    }
  }
}

.home-events-panel__item.is-first-deposit {
  --icon-w: 50px;
  --icon-h: 62px;
}

.home-events-panel__item.is-deposit-bonus,
.home-events-panel__item.is-download-app {
  --icon-w: 52px;
  --icon-h: 52px;
}

.home-events-panel__item.is-daily-cashback {
  --icon-w: 56px;
  --icon-h: 56px;
  --icon-x: -4px;
  --icon-y: -3px;
}

.home-events-panel__item.is-rewards-card,
.home-events-panel__item.is-lucky-wheel,
.home-events-panel__item.is-daily-signin,
.home-events-panel__item.is-phone-verification {
  --icon-w: 56px;
  --icon-h: 56px;
}

.home-events-panel__item.is-invite-chest,
.home-events-panel__item.is-weekly-reward {
  --icon-w: 60px;
  --icon-h: 58px;
}

.home-events-panel__item.is-referral {
  --icon-w: 56px;
  --icon-h: 58px;
}

.home-events-panel__item.is-red-packet {
  --icon-w: 60px;
  --icon-h: 60px;
  --icon-y: -4px;
}

.home-events-panel__item.is-cash-back {
  --icon-w: 60px;
  --icon-h: 64px;
}

.home-events-panel__item.is-rebate-rank {
  --icon-w: 58px;
  --icon-h: 68px;
}

.home-float-entry-wrap {
  position: relative;
  width: 60px;
  height: 60px;
  transition: height 220ms ease-out;
}

.home-float-entry-wrap.is-hidden {
  height: 266px;
  opacity: 0;
  pointer-events: none;
}

.home-float-entry,
.home-reward-entry,
.home-incentive {
  position: relative;
  width: 60px;
  height: 60px;
  overflow: visible;
  border: 0;
  padding: 0;
  background: transparent;
  cursor: pointer;
  touch-action: none;
}

.home-reward-entry {
  transition: opacity 180ms ease, transform 180ms ease;

  &.is-hidden {
    opacity: 0;
    pointer-events: none;
  }
}

.home-reward-entry__shell {
  position: absolute;
  inset: 0;
  border: 2px solid #25aaff;
  border-radius: 50%;
  background: linear-gradient(180deg, rgba(6, 74, 155, 0.98), rgba(0, 36, 94, 0.98));
  box-shadow: 0 3px 0 #071a52, inset 0 0 9px rgba(65, 207, 255, 0.72);
}

.home-reward-entry__glow,
.home-reward-entry__sparkles,
.home-reward-entry__coins {
  pointer-events: none;
  position: absolute;
  object-fit: contain;
}

.home-reward-entry__glow {
  inset: -5px;
  width: 70px;
  height: 70px;
  animation: homeRewardGlow 2.8s linear infinite;
}

.home-reward-entry__sparkles {
  inset: 3px;
  width: 54px;
  height: 54px;
  animation: homeRewardSparkles 1.7s ease-in-out infinite;
}

.home-reward-entry__coins {
  top: 4px;
  left: 10px;
  width: 42px;
  height: 42px;
  filter: drop-shadow(0 3px 3px rgba(0, 0, 0, 0.32));
  animation: homeRewardCoins 2.2s ease-in-out infinite;
}

.home-reward-entry__label {
  position: absolute;
  right: 4px;
  bottom: 4px;
  left: 4px;
  color: #ffffff;
  font-family: 'Poetsen One', Montserrat, sans-serif;
  font-size: 10px;
  line-height: 1;
  text-align: center;
  text-shadow: -1px -1px 0 #063178, 1px -1px 0 #063178, -1px 1px 0 #063178, 1px 1px 0 #063178;
}

.home-reward-entry__count {
  position: absolute;
  top: 5px;
  left: 5px;
  display: grid;
  min-width: 16px;
  height: 16px;
  place-items: center;
  border-radius: 999px;
  padding: 0 4px;
  color: #173600;
  background: #a9f04f;
  font-size: 9px;
  font-weight: 800;
}

.home-reward-entry__badge {
  position: absolute;
  top: 6px;
  right: 7px;
  width: 8px;
  height: 8px;

  &::before,
  span {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: #58e13e;
  }

  &::before {
    animation: home-badge-ping 1.35s cubic-bezier(0, 0, 0.2, 1) infinite;
    content: '';
  }
}

@keyframes homeRewardGlow {
  to { transform: rotate(360deg); }
}

@keyframes homeRewardSparkles {
  0%, 100% { opacity: 0.55; transform: scale(0.92); }
  50% { opacity: 1; transform: scale(1.08); }
}

@keyframes homeRewardCoins {
  0%, 100% { transform: translateY(1px) rotate(-2deg); }
  50% { transform: translateY(-3px) rotate(3deg); }
}

.home-float-entry__shell {
  pointer-events: none;
  position: absolute;
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
}

.home-float-entry__blur {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  backdrop-filter: blur(1.3px);
}

.home-float-entry__shine {
  pointer-events: none;
  position: absolute;
  top: 4px;
  left: 4px;
  width: 52px;
  height: 27px;
}

.home-float-entry__icon {
  pointer-events: none;
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;

  img {
    width: 39px;
    height: 39px;
    object-fit: contain;
    transform: translateY(-1.5px);
    animation: home-float-icon 2.2s ease-in-out infinite;
  }
}

.home-float-entry__label {
  pointer-events: none;
  position: absolute;
  bottom: 7px;
  left: 12px;
  display: flex;
  width: 36px;
  height: 13px;
  align-items: center;
  justify-content: center;
  border-radius: 0 0 12px 12px;
  background: linear-gradient(rgba(40, 66, 132, 0) 0%, #022554 54.71%);

  span {
    background-image: linear-gradient(#ffffff 47.076%, #2697fe 117.73%);
    background-clip: text;
    color: transparent;
    font-family: Montserrat, sans-serif;
    font-size: 10px;
    font-style: italic;
    font-weight: 800;
    line-height: 1.2;
    text-transform: capitalize;
    -webkit-text-fill-color: transparent;
  }
}

.home-float-entry__badge {
  position: absolute;
  top: 10px;
  right: 9px;
  z-index: 1;
  display: inline-flex;
  width: 8px;
  height: 8px;

  &::before,
  span {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    background: #ed252b;
  }

  &::before {
    animation: home-badge-ping 1.45s cubic-bezier(0, 0, 0.2, 1) infinite;
    content: "";
  }
}

.home-incentive__base {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: linear-gradient(#0f56a2 0%, #0773e6 100%);
}

.home-incentive__ring {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.home-incentive__progress {
  filter: drop-shadow(0 0 4px rgba(117, 232, 240, 0.78));
}

.home-incentive__sparkles {
  pointer-events: none;
  position: absolute;
  inset: 0;
  transform: rotate(324deg);
}

.home-incentive__sparkles-inner {
  position: absolute;
  top: -1px;
  left: 50%;
  width: 28px;
  height: 28px;
  transform: translateX(-50%) translateY(-50%) rotate(-324deg);

  i {
    position: absolute;
    border-radius: 50%;
    background: #61f1ff;
    box-shadow: 0 0 6px #61f1ff;
    animation: home-incentive-sparkle 1.65s ease-in-out infinite;

    &:nth-child(1) {
      top: 5.8px;
      left: 12px;
      width: 2px;
      height: 2px;
      animation-delay: -0.1s;
    }

    &:nth-child(2) {
      top: 6.3px;
      left: 18.4px;
      width: 1px;
      height: 1px;
      animation-delay: -0.3s;
    }

    &:nth-child(3) {
      top: 10.9px;
      left: 20px;
      width: 2px;
      height: 2px;
      animation-delay: -0.5s;
    }

    &:nth-child(4) {
      top: 10.3px;
      left: 5.4px;
      width: 1px;
      height: 1px;
      animation-delay: -0.7s;
    }

    &:nth-child(5) {
      top: 10.3px;
      left: 9.4px;
      width: 1px;
      height: 1px;
      animation-delay: -0.9s;
    }

    &:nth-child(6) {
      top: 15.3px;
      left: 17.4px;
      width: 1px;
      height: 1px;
      animation-delay: -1.1s;
    }

    &:nth-child(7) {
      top: 19.4px;
      left: 15.5px;
      width: 2.7px;
      height: 2.7px;
      animation-delay: -1.3s;
    }

    &:nth-child(8) {
      top: 21.3px;
      left: 11.4px;
      width: 1px;
      height: 1px;
      animation-delay: -1.5s;
    }
  }
}

.home-incentive__icon {
  pointer-events: none;
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;

  img {
    width: 20px;
    height: 38px;
    object-fit: contain;
  }
}

.home-incentive.is-boosting .home-incentive__icon img,
.home-incentive.is-boosting .home-incentive__fallback {
  animation: home-incentive-bottle-tap 520ms cubic-bezier(0.2, 0.9, 0.2, 1) both;
}

.home-incentive__fallback {
  position: relative;
  display: block;
  width: 22px;
  height: 38px;
  border: 2px solid rgba(117, 232, 240, 0.9);
  border-radius: 8px 8px 12px 12px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88) 0 8px, transparent 8px),
    linear-gradient(180deg, #7ff7ff 0%, #2bb6ff 55%, #1678db 100%);
  box-shadow: 0 0 10px rgba(0, 195, 234, 0.55);
  animation: home-bottle-wobble 1.8s ease-in-out infinite;

  &::before {
    position: absolute;
    top: -9px;
    left: 50%;
    width: 10px;
    height: 11px;
    border-radius: 4px 4px 2px 2px;
    background: #c8fbff;
    transform: translateX(-50%);
    content: "";
  }
}

.incentive-dialog {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: rgba(1, 6, 22, 0.8);
  color: #ffffff;
}

.incentive-dialog__glow,
.incentive-dialog__rays {
  pointer-events: none;
  position: fixed;
  top: 0;
  left: 50%;
  width: min(430px, 100vw);
  object-fit: cover;
  transform: translateX(-50%);
  user-select: none;
}

.incentive-dialog__glow {
  z-index: 0;
  height: 459px;
  opacity: 0.7;
  transform: translateX(-50%) scaleX(-1);
}

.incentive-dialog__rays {
  z-index: 0;
  height: min(708px, 100vh);
  opacity: 0.6;
  animation: incentive-dialog-rays 10s linear infinite;
}

.incentive-dialog__wrap {
  position: relative;
  z-index: 1;
  width: min(300px, calc(100vw - 48px));
  margin-top: 40px;
  overflow: visible;
}

.incentive-dialog__meter {
  position: absolute;
  top: 61.5px;
  left: 50%;
  z-index: 3;
  width: 150px;
  height: 150px;
  transform: translateX(-50%);
}

.incentive-dialog__meter-ring {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  filter: drop-shadow(0 0 8px rgba(0, 195, 234, 0.5));
}

.incentive-dialog__meter-sparkles {
  pointer-events: none;
  position: absolute;
  inset: 0;
  transform: rotate(324deg);

  > span {
    position: absolute;
    top: -1px;
    left: 50%;
    width: 28px;
    height: 28px;
    transform: translateX(-50%) translateY(-50%) rotate(-324deg);
  }

  i {
    position: absolute;
    border-radius: 50%;
    background: #61f1ff;
    box-shadow: 0 0 7px #61f1ff;
    animation: home-incentive-sparkle 1.65s ease-in-out infinite;

    &:nth-child(1) {
      top: 5.8px;
      left: 12px;
      width: 2px;
      height: 2px;
      animation-delay: -0.1s;
    }

    &:nth-child(2) {
      top: 6.3px;
      left: 18.4px;
      width: 1px;
      height: 1px;
      animation-delay: -0.3s;
    }

    &:nth-child(3) {
      top: 10.9px;
      left: 20px;
      width: 2px;
      height: 2px;
      animation-delay: -0.5s;
    }

    &:nth-child(4) {
      top: 10.3px;
      left: 5.4px;
      width: 1px;
      height: 1px;
      animation-delay: -0.7s;
    }

    &:nth-child(5) {
      top: 10.3px;
      left: 9.4px;
      width: 1px;
      height: 1px;
      animation-delay: -0.9s;
    }

    &:nth-child(6) {
      top: 15.3px;
      left: 17.4px;
      width: 1px;
      height: 1px;
      animation-delay: -1.1s;
    }

    &:nth-child(7) {
      top: 19.4px;
      left: 15.5px;
      width: 2.7px;
      height: 2.7px;
      animation-delay: -1.3s;
    }

    &:nth-child(8) {
      top: 21.3px;
      left: 11.4px;
      width: 1px;
      height: 1px;
      animation-delay: -1.5s;
    }
  }
}

.incentive-dialog__bottle {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 55px;
  height: 104px;
  object-fit: contain;
  transform: translate(-50%, -50%);
  animation: incentive-dialog-bottle 2.1s ease-in-out infinite;
  user-select: none;
}

.incentive-dialog__card {
  position: relative;
  z-index: 2;
  padding-top: 0;
  animation: incentive-card-in 260ms cubic-bezier(0.18, 0.89, 0.32, 1.28) both;
}

.incentive-dialog__banner {
  pointer-events: none;
  position: absolute;
  top: -20px;
  left: 50%;
  z-index: 4;
  width: 360px;
  height: 80px;
  transform: translateX(-50%);

  img,
  svg {
    position: absolute;
    left: 50%;
    display: block;
    transform: translateX(-50%);
  }

  img {
    top: 0;
    width: 360px;
    height: 80px;
    object-fit: contain;
  }

  svg {
    top: 13px;
    width: 260px;
    height: 36px;
    transform: translateX(-50%) translateY(-10px);
  }
}

.incentive-dialog__close {
  position: absolute;
  top: -48px;
  right: -16px;
  z-index: 5;
  display: flex;
  width: 32px;
  height: 32px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.15);
  color: #ffffff;
  cursor: pointer;
  backdrop-filter: blur(6px);

  svg {
    width: 16px;
    height: 16px;
  }
}

.incentive-dialog__panel {
  position: relative;
  height: min(415px, calc(100vh - 176px));
  min-height: 350px;
  overflow: hidden;
  border: 1px solid #49d9f6;
  border-radius: 14px;
  padding: 2px;
  background: linear-gradient(180deg, #003865 0%, #09549b 100%);
}

.incentive-dialog__panel-inner {
  position: relative;
  height: 100%;
  overflow: hidden;
  border: 1px solid rgba(146, 228, 255, 0.26);
  border-radius: 12px;
  background: linear-gradient(#113c9e 4.37%, #092a77 99.23%);
  box-shadow: inset 0 -4px 16px #5990e3;
}

.incentive-dialog__dots {
  position: absolute;
  top: 9px;
  right: 28px;
  left: 28px;
  display: flex;
  height: 8px;
  align-items: center;
  justify-content: space-between;
}

.incentive-dialog__dot-svg {
  display: block;
  width: 8px;
  height: 8px;
  flex: 0 0 auto;
  overflow: visible;
  animation: incentive-dot-chase 3.6s linear infinite;
  animation-delay: var(--delay);
}

.incentive-dialog__meter-space {
  height: 188px;
}

.incentive-dialog__score {
  position: relative;
  z-index: 2;
  margin: 12px 0 0;
  background: linear-gradient(180deg, #ffffd6 0%, #ffeb40 100%);
  background-clip: text;
  color: transparent;
  font-family: Montserrat, sans-serif;
  font-size: 16px;
  font-weight: 800;
  line-height: 1.2;
  text-align: center;
  text-transform: capitalize;
  filter: drop-shadow(0 0 4px #1464c7);
  -webkit-text-fill-color: transparent;
}

.incentive-dialog__tips {
  position: absolute;
  right: 10px;
  bottom: 10px;
  left: 10px;
  z-index: 2;
  height: 126px;
  overflow: hidden;
  border-radius: 10px;
  background: rgba(5, 5, 5, 0.5);
  padding: 0 10px;
}

.incentive-dialog__tips-track {
  display: flex;
  height: 100%;
  animation: incentive-tips-slide 7.2s ease-in-out infinite;
}

.incentive-dialog__tip {
  min-width: 100%;
  height: 100%;
  padding: 0 4px;
}

.incentive-dialog__mock-top,
.incentive-dialog__mock-body,
.incentive-dialog__mock-pills span,
.incentive-dialog__mock-bars span {
  background:
    linear-gradient(0deg, #012061 0%, #012061 100%),
    linear-gradient(#2853df 0%, #007ee9 100%);
}

.incentive-dialog__mock-top {
  width: 100%;
  height: 38px;
  border-radius: 0 0 6px 6px;
}

.incentive-dialog__mock-body {
  display: flex;
  min-height: 70px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  border-radius: 6px 6px 0 0;
  margin-top: 8px;
  padding: 6px 8px;
}

.incentive-dialog__mock-pills {
  display: flex;
  gap: 9px;
  padding-top: 10px;

  span {
    height: 26px;
    flex: 1;
    border-radius: 999px;
  }
}

.incentive-dialog__mock-bars {
  display: flex;
  gap: 10px;
  border-bottom: 1px solid #124974;
  padding: 10px 24px 8px;
  background:
    linear-gradient(0deg, #012061 0%, #012061 100%),
    linear-gradient(#2853df 0%, #007ee9 100%);

  span {
    height: 18px;
    flex: 1;
    border-radius: 999px;
    background: #134774;
  }
}

.incentive-dialog__tip p {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 0;
  color: #ffffff;
  font-family: Montserrat, sans-serif;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.3;
  text-align: center;
  text-shadow:
    -1px 0 #175ebb,
    0 1px #175ebb,
    1px 0 #175ebb,
    0 -1px #175ebb;

  img {
    width: 18px;
    height: 18px;
    object-fit: contain;
  }
}

.incentive-dialog__red-dot {
  display: block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ed252b;
}

.incentive-dialog__ok {
  position: relative;
  display: flex;
  width: 170px;
  height: 40px;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 0;
  margin: 40px auto 0;
  padding: 0 16px 4px;
  background: transparent;
  cursor: pointer;

  &:active {
    opacity: 0.9;
  }
}

.incentive-dialog__ok-base,
.incentive-dialog__ok-top,
.incentive-dialog__ok-shine {
  pointer-events: none;
  position: absolute;
}

.incentive-dialog__ok-base {
  inset: 12px 0 0;
  border-radius: 28px;
  background: linear-gradient(#2373fd 0%, #264eab 100%);
  box-shadow:
    inset 0 -1px 4px #2654b7,
    inset 0 0 0 0.5px #4d82ff;
}

.incentive-dialog__ok-top {
  inset: 0 0 4px;
  overflow: hidden;
  border-radius: 40px;
  background: linear-gradient(10.39deg, #2172ff 23.36%, #3e75ee 76.64%);
  box-shadow:
    inset 0 1px 0 #c8fffd,
    inset 0 -1px 0 #c8fffd,
    inset 0 0 14px rgba(248, 200, 25, 0.24);
}

.incentive-dialog__ok-shine {
  top: 3px;
  right: 2px;
  left: 2px;
  height: 29px;
  border-radius: 40px;
  background: linear-gradient(rgba(255, 255, 255, 0.5) 0%, rgba(255, 255, 255, 0) 45%);
}

.incentive-dialog__ok-label {
  position: relative;
  z-index: 1;
  color: #fffffc;
  font-family: Montserrat, sans-serif;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.2;
  text-shadow: rgba(0, 0, 0, 0.16) 0 0 4px;
}

.incentive-dialog-enter-active,
.incentive-dialog-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.incentive-dialog-enter-from,
.incentive-dialog-leave-to {
  opacity: 0;
}

.incentive-dialog-enter-active .incentive-dialog__wrap,
.incentive-dialog-leave-active .incentive-dialog__wrap {
  transition:
    opacity 200ms ease,
    transform 200ms ease;
}

.incentive-dialog-enter-from .incentive-dialog__wrap,
.incentive-dialog-leave-to .incentive-dialog__wrap {
  opacity: 0;
  transform: scale(0.94);
}

.home-events-panel-enter-active,
.home-events-panel-leave-active {
  transition:
    opacity 180ms ease,
    transform 180ms ease;
}

.home-events-panel-enter-from,
.home-events-panel-leave-to {
  opacity: 0;
  transform: translateX(12px) scale(0.96);
}

@keyframes home-float-icon {
  0%,
  100% {
    transform: translateY(-1.5px) scale(1);
  }

  50% {
    transform: translateY(-4px) scale(1.05);
  }
}

@keyframes home-badge-ping {
  75%,
  100% {
    opacity: 0;
    transform: scale(2.1);
  }
}

@keyframes home-incentive-sparkle {
  0%,
  100% {
    opacity: 0.12;
    transform: scale(0.55);
  }

  50% {
    opacity: 1;
    transform: scale(1.12);
  }
}

@keyframes home-bottle-wobble {
  0%,
  100% {
    transform: scale(0.96) rotate(-3deg);
  }

  50% {
    transform: scale(1.03) rotate(3deg);
  }
}

@keyframes home-incentive-bottle-tap {
  0% {
    transform: scale(1) translateX(0) rotate(0deg);
  }

  22% {
    transform: scale(1.2) translateX(0) rotate(0deg);
  }

  38% {
    transform: scale(1.18) translateX(-3px) rotate(-7deg);
  }

  54% {
    transform: scale(1.18) translateX(3px) rotate(7deg);
  }

  70% {
    transform: scale(1.12) translateX(-1px) rotate(-3deg);
  }

  100% {
    transform: scale(1) translateX(0) rotate(0deg);
  }
}

@keyframes incentive-dialog-rays {
  0% {
    transform: translateX(-50%) rotate(0deg) scale(1.04);
  }

  100% {
    transform: translateX(-50%) rotate(360deg) scale(1.04);
  }
}

@keyframes incentive-dialog-bottle {
  0%,
  100% {
    transform: translate(-50%, calc(-50% - 1px)) scale(1);
  }

  50% {
    transform: translate(-50%, calc(-50% - 5px)) scale(1.025);
  }
}

@keyframes incentive-card-in {
  0% {
    opacity: 0;
    transform: scale(0.9);
  }

  100% {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes incentive-dot-chase {
  0%,
  100% {
    opacity: 0.55;
    transform: scale(0.72);
  }

  45% {
    opacity: 1;
    transform: scale(1.12);
  }
}

@keyframes incentive-tips-slide {
  0%,
  27% {
    transform: translateX(0);
  }

  34%,
  61% {
    transform: translateX(-100%);
  }

  68%,
  94% {
    transform: translateX(-200%);
  }

  100% {
    transform: translateX(0);
  }
}
</style>
