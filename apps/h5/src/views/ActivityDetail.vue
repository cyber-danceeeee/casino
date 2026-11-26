<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { gsap } from 'gsap'
import ActivityGuideFlow from '../components/ActivityGuideFlow.vue'
import ContactMeBadge from '../components/ContactMeBadge.vue'
import {
  getSourceActivity,
  normalizeActivitySlug,
  sourceActivityAssets,
  type SourceActivityConfig,
} from '../mock/sourceActivities'
import { localProfileImage } from '../utils/localAssets'
import { h5Audio } from '../utils/staticAssets'
import { formatCurrency } from '../services/runtimeConfig'
import { isAndroidAppRuntime } from '../utils/runtimeEnvironment'

const route = useRoute()
const router = useRouter()
const androidAppRuntime = isAndroidAppRuntime()
const rootRef = ref<HTMLElement | null>(null)
const rulesRef = ref<HTMLElement | null>(null)
const wheelDetailsRef = ref<HTMLElement | null>(null)
const guideVisible = ref(false)
const rulesDialogVisible = ref(false)
const dailyCashbackTab = ref<'level' | 'records'>('level')
const wheelRecordTab = ref<'notice' | 'history'>('notice')
const wheelStarted = ref(false)
const wheelStarting = ref(false)
const wheelCountdown = ref(['02', '23', '05', '14'])
const wheelSpinning = ref(false)
const wheelRemainingSpins = ref(1)
const wheelSpinRotation = ref(0)
const wheelSpinCount = ref(0)
const wheelSpinResultIndex = ref<number | null>(null)
const wheelFloatingAmount = ref<string | null>(null)
const wheelCurrentAmount = ref(0)
const customerImage = localProfileImage('customer.webp')
const wheelSpinAudioSrc = h5Audio('invitation-wheel-spin.mp3')
const wheelEndAudioSrc = h5Audio('invitation-wheel-end.wav')
const formatActivityMoney = (value: number | string, options?: { includeCode?: boolean }) =>
  formatCurrency(value, options)
let animationContext: gsap.Context | undefined
let wheelStartTimer: number | undefined
let wheelCountdownTimer: number | undefined
let wheelCountdownDeadline = Date.now()
let wheelSpinEndTimer: number | undefined
let wheelFloatingTimer: number | undefined
let wheelSpinAudioEl: HTMLAudioElement | undefined

const routeSlug = computed(() => normalizeActivitySlug(route.path))
const activity = computed(() => getSourceActivity(routeSlug.value))
const now = ref(new Date())
const rainPackets = Array.from({ length: 14 }, (_, index) => index)
const checkinDays = ['Day 1', 'Day 2', 'Day 3', 'Day 4', 'Day 5', 'Day 6', 'Day 7']
const chestSteps = Array.from({ length: 30 }, (_, index) => index + 1)
const countdownBlocks = [
  { value: '10', label: 'Hours' },
  { value: '54', label: 'Minutes' },
  { value: '50', label: 'Seconds' },
]
const wheelNoticeRows = [
  { name: 'X******8', time: '2026-07-10 10:22:32', tone: 'gold' },
  { name: 'Q******2', time: '2026-07-10 10:15:29', tone: 'blue' },
  { name: 'A******7', time: '2026-07-10 10:12:43', tone: 'violet' },
  { name: 'E******5', time: '2026-07-10 10:04:58', tone: 'orange' },
  { name: 'Z******7', time: '2026-07-10 09:55:29', tone: 'orange' },
  { name: 'S******2', time: '2026-07-10 09:46:43', tone: 'pink' },
  { name: 'L******2', time: '2026-07-10 09:42:44', tone: 'cyan' },
  { name: 'P******7', time: '2026-07-10 09:43:07', tone: 'gold' },
  { name: 'J******5', time: '2026-07-10 09:31:18', tone: 'green' },
  { name: 'U******8', time: '2026-07-10 09:25:02', tone: 'mint' },
]
const wheelNoticeLoopRows = [...wheelNoticeRows, ...wheelNoticeRows]
const wheelLabels = ['1.5k', '15', '0.01-75', '30', '0.1-30', '150', '11-1.4k', '10-1.5k']
const wheelRewardValues = [1500, 15, 75, 30, 30, 150, 1400, 1500]
const wheelSpinTargets = [1, 3, 5, 2]
const wheelHistory = {
  amount: formatActivityMoney(1500),
  startTime: '2026-07-10 15:21:22',
  endTime: '2026-07-13 15:21:22',
  status: 'Pending',
}
const wheelPrizes = computed(() => {
  const detail = activity.value
  if (!detail) return []
  return Array.from({ length: 8 }, (_, index) => detail.images[`prize${index + 1}`]).filter(Boolean)
})
const wheelProgressPercent = computed(() => Math.min(100, Math.max(0, (wheelCurrentAmount.value / 1500) * 100)))
const wheelSpinStyle = computed(() => ({
  transform: `translateX(-50%) rotate(${wheelSpinRotation.value}deg)`,
  transition: wheelSpinning.value ? 'transform 5000ms cubic-bezier(0.12, 0.74, 0.18, 1)' : 'none',
}))
const wheelProgressStyle = computed(() => ({
  width: `${wheelProgressPercent.value}%`,
}))
const wheelCurrentAmountLabel = computed(() => formatActivityMoney(wheelCurrentAmount.value))
const wheelRemainingAmountLabel = computed(() => formatActivityMoney(Math.max(0, 1500 - wheelCurrentAmount.value)))
const wheelTargetAmountLabel = computed(() => formatActivityMoney(1500))
const wheelMidProgressComplete = computed(() => wheelProgressPercent.value >= 50)
const wheelEndProgressComplete = computed(() => wheelProgressPercent.value >= 100)
const daySuffix = (day: number) => {
  if (day >= 11 && day <= 13) return 'th'
  const lastDigit = day % 10
  if (lastDigit === 1) return 'st'
  if (lastDigit === 2) return 'nd'
  if (lastDigit === 3) return 'rd'
  return 'th'
}
const formatDatePart = (date: Date) =>
  `${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
const formatDateTime = (date: Date, time: string) =>
  `${date.getFullYear()}-${formatDatePart(date)} ${time}`
const todayDisplay = computed(() => `${now.value.getDate()}${daySuffix(now.value.getDate())},${now.value.toLocaleString('en-US', { month: 'short' })}`)
const tomorrow = computed(() => {
  const date = new Date(now.value)
  date.setDate(date.getDate() + 1)
  return date
})
const dailyClaimRange = computed(() =>
  `${formatDateTime(now.value, '03:00:00')} - ${formatDateTime(tomorrow.value, '02:59:59')}`,
)
const weeklyPeriodRange = computed(() => {
  const start = new Date(now.value)
  const day = start.getDay() || 7
  start.setDate(start.getDate() - day + 1)
  const end = new Date(start)
  end.setDate(end.getDate() + 7)

  return `${formatDateTime(start, '03:00:00')} - ${formatDateTime(end, '02:59:59')}`
})
const firstDepositMaximum = computed(() => {
  const match = activity.value?.badge.match(/Maximum Reward:\\s*(.+)$/)
  return match?.[1] || '+25,000.00'
})

const tableActivityVariants = new Set([
  'daily-deposit',
  'periodic-green',
  'periodic-gold',
  'relief',
  'sub-deposit',
])

const pageStyle = computed(() => {
  const detail = activity.value
  const fallbackBg = detail?.images.pageBg || detail?.images.heroBg || ''
  return {
    '--source-page-bg': fallbackBg ? `url("${fallbackBg}")` : 'none',
    '--source-first-frame': detail?.images.frame ? `url("${detail.images.frame}")` : 'none',
    '--source-first-treasure-banner': detail?.images.banner ? `url("${detail.images.banner}")` : 'none',
  }
})

const cardGridClasses = computed(() => {
  const detail = activity.value
  if (!detail) return {}

  return {
    'source-card-grid--table': tableActivityVariants.has(detail.variant),
    'source-card-grid--welfare': detail.variant === 'welfare',
    'source-card-grid--chest': detail.variant === 'chest',
  }
})

const tableHead = computed(() => {
  const detail = activity.value
  if (!detail) return []

  if (detail.variant === 'daily-deposit') return ['Deposit Amount', 'Total', 'Details']
  if (detail.variant === 'periodic-green' || detail.variant === 'periodic-gold') return ['Deposit Amount', '1st / 2nd', '3rd']
  if (detail.variant === 'relief') return ['Valid Bets', 'Cashback', 'Type']
  if (detail.variant === 'sub-deposit') return ['Cycle Total Recharge', 'Reward Ratio', 'Type']
  return []
})

const welfareCardThumb = (index: number) => {
  const detail = activity.value
  if (!detail) return ''
  return index === 0 ? detail.images.weekThumb : detail.images.monthThumb
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }

  void router.push({ name: 'offers', query: { tab: 'events' } })
}

const scrollToRules = () => {
  rulesRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const openRules = () => {
  if (
    activity.value?.variant === 'first-deposit'
    || activity.value?.variant === 'deposit-bonus'
    || activity.value?.variant === 'daily-deposit'
    || activity.value?.variant === 'wheel'
  ) {
    rulesDialogVisible.value = true
    return
  }

  scrollToRules()
}

const closeRules = () => {
  rulesDialogVisible.value = false
}

const openContact = () => {
  void router.push({ name: 'support' })
}

const closeGuide = () => {
  guideVisible.value = false
}

const padCountdown = (value: number) => String(value).padStart(2, '0')

const updateWheelCountdown = () => {
  const remaining = Math.max(0, wheelCountdownDeadline - Date.now())
  const totalSeconds = Math.floor(remaining / 1000)
  const days = Math.floor(totalSeconds / 86400)
  const hours = Math.floor((totalSeconds % 86400) / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  wheelCountdown.value = [days, hours, minutes, seconds].map(padCountdown)
}

const resetWheelCountdown = () => {
  wheelCountdownDeadline = Date.now() + ((2 * 24 * 60 * 60) + (23 * 60 * 60) + (5 * 60) + 14) * 1000
  updateWheelCountdown()
}

const scrollToWheelDetails = (tab: 'notice' | 'history' = 'history') => {
  wheelRecordTab.value = tab
  void nextTick(() => {
    wheelDetailsRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  })
}

const stopWheelSpinAudio = () => {
  if (!wheelSpinAudioEl) return

  wheelSpinAudioEl.pause()
  wheelSpinAudioEl.currentTime = 0
  wheelSpinAudioEl = undefined
}

const clearWheelSpinTimers = () => {
  if (wheelSpinEndTimer) {
    window.clearTimeout(wheelSpinEndTimer)
    wheelSpinEndTimer = undefined
  }

  if (wheelFloatingTimer) {
    window.clearTimeout(wheelFloatingTimer)
    wheelFloatingTimer = undefined
  }
}

const resetWheelSpinState = () => {
  clearWheelSpinTimers()
  stopWheelSpinAudio()
  wheelSpinning.value = false
  wheelRemainingSpins.value = 1
  wheelSpinRotation.value = 0
  wheelSpinCount.value = 0
  wheelSpinResultIndex.value = null
  wheelFloatingAmount.value = null
  wheelCurrentAmount.value = 0
}

const playWheelSpinAudio = () => {
  stopWheelSpinAudio()
  wheelSpinAudioEl = new Audio(wheelSpinAudioSrc)
  wheelSpinAudioEl.loop = true
  wheelSpinAudioEl.volume = 0.46
  void wheelSpinAudioEl.play().catch(() => {
    stopWheelSpinAudio()
  })
}

const playWheelEndAudio = () => {
  const audio = new Audio(wheelEndAudioSrc)
  audio.volume = 0.62
  audio.playbackRate = 1.25
  void audio.play().catch(() => {})
}

const spinWheel = () => {
  if (!wheelStarted.value || wheelSpinning.value) return

  if (wheelRemainingSpins.value <= 0) {
    scrollToWheelDetails('notice')
    return
  }

  clearWheelSpinTimers()
  wheelFloatingAmount.value = null
  wheelSpinning.value = true
  wheelRemainingSpins.value = Math.max(0, wheelRemainingSpins.value - 1)
  const targetIndex = wheelSpinTargets[wheelSpinCount.value % wheelSpinTargets.length]
  wheelSpinCount.value += 1
  wheelSpinResultIndex.value = targetIndex
  playWheelSpinAudio()

  const currentRotation = wheelSpinRotation.value
  const normalizedRotation = ((currentRotation % 360) + 360) % 360
  const baseRotation = currentRotation - normalizedRotation
  const targetRotation = baseRotation + 6 * 360 - targetIndex * 45 + 22.5

  window.requestAnimationFrame(() => {
    wheelSpinRotation.value = targetRotation <= currentRotation ? targetRotation + 360 : targetRotation
  })

  wheelSpinEndTimer = window.setTimeout(() => {
    const reward = wheelRewardValues[targetIndex] ?? 0
    wheelSpinning.value = false
    stopWheelSpinAudio()
    playWheelEndAudio()
    wheelFloatingAmount.value = `+${reward.toLocaleString('en-US')}`
    wheelCurrentAmount.value = Math.min(1500, wheelCurrentAmount.value + Math.min(reward, 104.97))
    wheelSpinEndTimer = undefined
    wheelFloatingTimer = window.setTimeout(() => {
      wheelFloatingAmount.value = null
      wheelFloatingTimer = undefined
    }, 3000)
  }, 5200)
}

const startWheelRound = () => {
  if (wheelStarted.value || wheelStarting.value) return

  resetWheelSpinState()
  wheelStarting.value = true
  if (wheelStartTimer) window.clearTimeout(wheelStartTimer)
  wheelStartTimer = window.setTimeout(() => {
    wheelStarted.value = true
    wheelStarting.value = false
    wheelStartTimer = undefined
    void nextTick(() => playAnimations())
  }, 920)
}

const handleCta = (detail: SourceActivityConfig) => {
  if (detail.ctaType === 'download') {
    if (androidAppRuntime) return
    guideVisible.value = true
    return
  }

  if (detail.ctaType === 'deposit') {
    void router.push({ name: 'wallet-deposit' })
    return
  }

  if (detail.ctaType === 'share') {
    void router.push({ name: 'share', query: { tab: 'social' } })
    return
  }

  if (detail.ctaType === 'verify') {
    void router.push({ name: 'security' })
    return
  }

  scrollToRules()
}

const playAnimations = async () => {
  await nextTick()
  animationContext?.revert()

  if (!rootRef.value || !activity.value) return

  animationContext = gsap.context(() => {
    const animateTargets = (selector: string, animate: (targets: HTMLElement[]) => void) => {
      const targets = Array.from(rootRef.value?.querySelectorAll<HTMLElement>(selector) || [])
      if (targets.length) animate(targets)
    }

    animateTargets('.source-nav button', (targets) => gsap.from(targets, {
      y: -10,
      autoAlpha: 0,
      duration: 0.34,
      stagger: 0.07,
      ease: 'power2.out',
    }))

    animateTargets('.source-hero__copy > *, .source-deposit-heading > *, .source-badge', (targets) => gsap.from(targets, {
      y: 20,
      autoAlpha: 0,
      duration: 0.56,
      stagger: 0.08,
      ease: 'back.out(1.35)',
    }))

    animateTargets('.source-layer, .source-rank-card, .source-check-day', (targets) => gsap.from(targets, {
      y: 24,
      scale: 0.9,
      autoAlpha: 0,
      duration: 0.64,
      stagger: 0.06,
      ease: 'back.out(1.4)',
    }))

    animateTargets('.source-panel, .source-deposit-panel, .source-table-head, .source-card, .source-rule-card', (targets) => gsap.from(targets, {
      y: 22,
      autoAlpha: 0,
      duration: 0.46,
      stagger: 0.05,
      ease: 'power2.out',
    }))

    animateTargets('.source-cta', (targets) =>
      gsap.fromTo(
        targets,
        { scale: 0.78, autoAlpha: 0 },
        {
          scale: 1,
          autoAlpha: 1,
          duration: 0.42,
          delay: 0.18,
          ease: 'back.out(1.8)',
          clearProps: 'transform,opacity,visibility',
        },
      ),
    )

    animateTargets('.source-float', (targets) => gsap.to(targets, {
      y: (index) => (index % 2 === 0 ? -10 : 10),
      x: (index) => (index % 3 === 0 ? 7 : -7),
      rotate: (index) => (index % 2 === 0 ? 7 : -7),
      duration: (index) => 1.9 + index * 0.12,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
      stagger: 0.09,
    }))

    animateTargets('.source-first-percent', (targets) => gsap.to(targets, {
      scale: 1.04,
      rotate: 0,
      transformOrigin: '50% 64%',
      duration: 1.55,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
    }))

    animateTargets('.source-first__bundle', (targets) => gsap.to(targets, {
      y: -2,
      // scale: 1.025,
      transformOrigin: '50% 62%',
      duration: 2.35,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
    }))

    animateTargets('.source-first__gift-float', (targets) => gsap.to(targets, {
      x: 5,
      y: -8,
      rotate: 4,

      transformOrigin: '50% 58%',
      duration: 2.8,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
      overwrite: 'auto',
    }))

    animateTargets('.source-first__treasure-float', (targets) => gsap.to(targets, {
      x: -6,
      y: -10,
      rotate: -3,
      scale: 1.08,
      transformOrigin: '50% 58%',
      duration: 2.15,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
    }))

    animateTargets('.source-first__treasure-float', (targets) => gsap.to(targets, {
      transformOrigin: '50% 50%',
      duration: 1.7,
      repeat: -1,
      ease: 'none',
    }))

    animateTargets('.source-first__coin-float', (targets) => gsap.to(targets, {
      rotate: 360,
      transformOrigin: '50% 50%',
      duration: 1.7,
      repeat: -1,
      ease: 'none',
    }))

    animateTargets('.source-first__coin-float', (targets) => gsap.to(targets, {
      x: -5,
      y: 6,
      scale: 1.07,
      transformOrigin: '50% 50%',
      duration: 1.95,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
    }))

    animateTargets('.source-rays', (targets) => gsap.to(targets, {
      rotate: 360,
      transformOrigin: '50% 50%',
      duration: 11,
      repeat: -1,
      ease: 'none',
    }))

    animateTargets('.source-glow', (targets) => gsap.to(targets, {
      autoAlpha: 0.44,
      scale: 1.08,
      duration: 1.4,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
      stagger: 0.12,
    }))

    animateTargets('.source-wheel__wheel', (targets) => gsap.to(targets, {
      rotate: 360,
      transformOrigin: '50% 50%',
      duration: 9,
      repeat: -1,
      ease: 'none',
    }))

    animateTargets('.source-wheel__prize', (targets) => gsap.to(targets, {
      y: -7,
      scale: 1.08,
      duration: 1.2,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
      stagger: 0.08,
    }))

    animateTargets('.source-rain-packet', (targets) => gsap.to(targets, {
      y: 760,
      x: (index) => (index % 2 === 0 ? 16 : -16),
      rotate: (index) => 90 + index * 12,
      duration: (index) => 3 + (index % 5) * 0.38,
      repeat: -1,
      ease: 'none',
      stagger: {
        each: 0.2,
        repeat: -1,
      },
    }))

    animateTargets('.source-shimmer', (targets) => gsap.to(targets, {
      xPercent: 140,
      duration: 1.8,
      repeat: -1,
      repeatDelay: 1.1,
      ease: 'power1.inOut',
    }))

    animateTargets('.source-tap-hand', (targets) => gsap.to(targets, {
      x: 0,
      y: 0,
      rotate: 0,
      scale: 0.8,
      transformOrigin: '72% 78%',
      duration: 0.11,
      yoyo: true,
      repeat: -1,
      repeatDelay: 0.12,
      ease: 'power1.inOut',
    }))

    animateTargets('.source-halo-pulse', (targets) => gsap.to(targets, {
      scale: 1.14,
      autoAlpha: 0.36,
      duration: 1.05,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
      stagger: 0.14,
    }))

    animateTargets('.source-bell', (targets) => gsap.to(targets, {
      rotate: 10,
      transformOrigin: '50% 12%',
      duration: 0.52,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
    }))

    animateTargets('.source-car', (targets) => gsap.to(targets, {
      x: (index) => (index === 0 ? -8 : 8),
      duration: 1.2,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
      stagger: 0.1,
    }))
  }, rootRef.value)
}

onMounted(() => {
  now.value = new Date()
  resetWheelCountdown()
  wheelCountdownTimer = window.setInterval(updateWheelCountdown, 1000)
  void playAnimations()
})

watch(routeSlug, () => {
  rulesDialogVisible.value = false
  dailyCashbackTab.value = 'level'
  wheelRecordTab.value = 'notice'
  wheelStarted.value = false
  wheelStarting.value = false
  resetWheelSpinState()
  if (wheelStartTimer) {
    window.clearTimeout(wheelStartTimer)
    wheelStartTimer = undefined
  }
  resetWheelCountdown()
  void playAnimations()
})

onBeforeUnmount(() => {
  animationContext?.revert()
  resetWheelSpinState()
  if (wheelStartTimer) {
    window.clearTimeout(wheelStartTimer)
    wheelStartTimer = undefined
  }
  if (wheelCountdownTimer) {
    window.clearInterval(wheelCountdownTimer)
    wheelCountdownTimer = undefined
  }
})
</script>

<template>
  <section
    v-if="activity"
    ref="rootRef"
    class="source-activity"
    :class="[
      `source-activity--${activity.variant}`,
      `source-activity--${activity.theme}`,
      {
        'is-wheel-started': activity.variant === 'wheel' && wheelStarted,
        'is-wheel-starting': activity.variant === 'wheel' && wheelStarting,
      },
    ]"
    :style="pageStyle"
  >
    <nav class="source-nav" aria-label="Activity navigation">
      <button type="button" class="source-pressable" aria-label="Back" @click="goBack">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M4.29 11.29a1 1 0 0 0 0 1.42l4 4a1 1 0 0 0 1.42-1.42L7.41 13H19a1 1 0 1 0 0-2H7.41l2.3-2.29a1 1 0 0 0-1.42-1.42l-4 4Z" />
        </svg>
      </button>
      <button type="button" class="source-pressable" @click="openRules">Event Rules</button>
    </nav>

    <main class="source-scroll">
      <section class="source-hero">
        <div v-if="activity.variant === 'red-packet'" class="source-rain-layer" aria-hidden="true">
          <img
            v-for="packet in rainPackets"
            :key="packet"
            class="source-rain-packet"
            :src="activity.images.falling"
            alt=""
            :style="{ left: `${(packet * 29) % 95}%`, top: `${-280 - packet * 38}px` }"
          />
        </div>

        <template v-if="activity.variant === 'first-deposit'">
          <div class="source-first-heading">
            <h1 class="source-first-title" :aria-label="activity.title">
              <span class="source-first-title__shadow" aria-hidden="true">{{ activity.title }}</span>
              <span class="source-first-title__stroke" aria-hidden="true">{{ activity.title }}</span>
              <span class="source-first-title__fill" aria-hidden="true">{{ activity.title }}</span>
            </h1>
            <p>{{ activity.subtitle }}</p>
            <div class="source-first-percent">
              <span>+50.00%</span>
              <em>Rewards</em>
            </div>
            <div class="source-first-max">Maximum Reward: <strong>{{ firstDepositMaximum }}</strong></div>
          </div>
        </template>

        <template v-else-if="activity.variant === 'deposit-bonus'">
          <div class="source-deposit-heading">
            <h1>{{ activity.subtitle }}</h1>
          </div>
        </template>

        <template v-else-if="activity.variant === 'daily-deposit'">
          <div class="source-daily-heading">
            <h1>{{ activity.title }}</h1>
            <p>{{ activity.subtitle }}</p>
          </div>
        </template>

        <template v-else-if="activity.variant === 'wheel'">
          <div class="source-wheel-copy" aria-hidden="true" />
        </template>

        <template v-else>
          <div class="source-hero__copy">
            <p>{{ activity.subtitle }}</p>
            <h1>{{ activity.title }}</h1>
          </div>

          <div class="source-badge">
            <span>{{ activity.badge }}</span>
          </div>
        </template>

        <div class="source-stage" :class="{ 'source-stage--first': activity.variant === 'first-deposit' }">
          <template v-if="activity.variant === 'first-deposit'">
            <img class="source-layer source-first__bundle" :src="activity.images.heroBox || activity.images.bundle" alt="" />
            <img class="source-layer source-float source-first__gift-float" :src="activity.images.giftBox" alt="" />
            <img class="source-layer source-first__treasure-float" :src="activity.images.treasure" alt="" aria-hidden="true" />
            <img class="source-layer source-first__coin-float" :src="activity.images.coin" alt="" aria-hidden="true" />
          </template>

          <template v-else-if="activity.variant === 'deposit-bonus'">
            <img class="source-layer source-deposit__bg" :src="activity.images.heroBg" alt="" />
            <img class="source-layer source-deposit__light" :src="activity.images.light" alt="" aria-hidden="true" />
            <img class="source-layer source-deposit__pile" :src="activity.images.pile" alt="" aria-hidden="true" />
            <div class="source-layer source-deposit__lower-bg" aria-hidden="true" />
            <img class="source-layer source-deposit__chest" :src="activity.images.chest" alt="" />
            <img class="source-layer source-deposit__coin" :src="activity.images.coin" alt="" aria-hidden="true" />
            <img class="source-layer source-deposit__cash" :src="sourceActivityAssets.cash" alt="" aria-hidden="true" />
            <svg
              class="source-deposit__arc"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 375 19"
              fill="none"
              preserveAspectRatio="none"
              aria-hidden="true"
            >
              <defs>
                <filter id="source-deposit-arc-divider-glow" x="-1.5" y="0" width="378" height="19" filterUnits="userSpaceOnUse" color-interpolation-filters="sRGB">
                  <feFlood flood-opacity="0" result="BackgroundImageFix" />
                  <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape" />
                  <feColorMatrix in="SourceAlpha" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0" result="hardAlpha" />
                  <feOffset />
                  <feGaussianBlur stdDeviation="0.5" />
                  <feComposite in2="hardAlpha" operator="arithmetic" k2="-1" k3="1" />
                  <feColorMatrix type="matrix" values="0 0 0 0 1 0 0 0 0 1 0 0 0 0 1 0 0 0 0.85 0" />
                  <feBlend mode="normal" in2="shape" result="effect1_innerShadow" />
                </filter>
              </defs>
              <g filter="url(#source-deposit-arc-divider-glow)">
                <path class="source-deposit__arc-path" d="M0 1.5C0 1.5 76 17.5 187.5 17.5C299 17.5 375 1.5 375 1.5" stroke="#FFD5D4" stroke-width="3" stroke-linecap="round" stroke-dasharray="0.3 8" />
              </g>
            </svg>
          </template>

          <template v-else-if="activity.variant === 'daily-deposit'">
            <img class="source-layer source-daily__bg" :src="activity.images.heroBg" alt="" />
            <img class="source-layer source-daily__main" :src="activity.images.main" alt="" />
          </template>

          <template v-else-if="activity.variant === 'periodic-green' || activity.variant === 'periodic-gold'">
            <img class="source-layer source-illustrated__bg" :src="activity.images.heroBg" alt="" />
            <img class="source-layer source-float source-illustrated__main source-illustrated__main--large" :src="activity.images.main" alt="" />
            <img class="source-layer source-rays source-illustrated__rays" :src="sourceActivityAssets.rays" alt="" aria-hidden="true" />
            <img v-if="activity.images.claimGift" class="source-layer source-float source-periodic__gift" :src="activity.images.claimGift" alt="" />
          </template>

          <template v-else-if="activity.variant === 'wheel'">
            <span class="source-wheel__stage-anchor" aria-hidden="true" />
          </template>

          <template v-else-if="activity.variant === 'chest'">
            <img class="source-layer source-chest__tunnel" :src="activity.images.tunnel" alt="" />
            <img class="source-layer source-chest__stage" :src="activity.images.stage" alt="" />
            <img class="source-layer source-chest__title" :src="activity.images.titleFrame" alt="" />
            <strong class="source-layer source-chest__frame-title">Invite Chest</strong>
            <img class="source-layer source-float source-chest__hostess" :src="activity.images.hostess" alt="" />
            <img class="source-layer source-float source-chest__treasure" :src="activity.images.treasure" alt="" />
            <img class="source-layer source-float source-chest__blue" :src="activity.images.blueChest" alt="" />
          </template>

          <template v-else-if="activity.variant === 'sign-in'">
            <img class="source-layer source-sign__bg" :src="activity.images.viewport" alt="" />
            <img class="source-layer source-float source-sign__treasure" :src="activity.images.treasure" alt="" />
            <img class="source-layer source-car source-sign__car-left" :src="activity.images.carLeft" alt="" aria-hidden="true" />
            <img class="source-layer source-car source-sign__car-right" :src="activity.images.carRight" alt="" aria-hidden="true" />
            <img class="source-layer source-bell source-sign__bell" :src="activity.images.bell" alt="" aria-hidden="true" />
            <div class="source-sign__board" aria-label="Check in rewards">
              <span
                v-for="(day, index) in checkinDays"
                :key="day"
                class="source-check-day"
                :class="{ 'is-active': index === 0 }"
              >
                <img :src="index === 0 ? activity.images.lit : activity.images.dim" alt="" />
                <b>{{ day }}</b>
              </span>
            </div>
          </template>

          <template v-else-if="activity.variant === 'red-packet'">
            <img class="source-layer source-red__bg" :src="activity.images.heroBg" alt="" />
            <img class="source-layer source-red__stage" :src="activity.images.stage" alt="" aria-hidden="true" />
            <img class="source-layer source-red__subtitle" :src="activity.images.subtitle" alt="" />
            <img class="source-layer source-red__countdown" :src="activity.images.countdown" alt="" />
            <img class="source-layer source-glow source-halo-pulse source-red__halo" :src="activity.images.halo" alt="" aria-hidden="true" />
            <img class="source-layer source-glow source-red__reward" :src="activity.images.rewardCash" alt="" />
            <img class="source-layer source-red__cell" :src="activity.images.cellBonus" alt="" />
            <img class="source-layer source-float source-red__packet" :src="activity.images.packet" alt="" />
            <img class="source-layer source-float source-red__packets" :src="activity.images.packets" alt="" aria-hidden="true" />
            <img class="source-layer source-red__confetti" :src="activity.images.confetti" alt="" aria-hidden="true" />
          </template>

          <template v-else-if="activity.variant === 'relief'">
            <img class="source-layer source-illustrated__bg" :src="activity.images.top" alt="" />
            <img class="source-layer source-float source-illustrated__main source-illustrated__main--large" :src="activity.images.main" alt="" />
            <img class="source-layer source-glow source-relief__coin-bg" :src="activity.images.coinBg" alt="" aria-hidden="true" />
          </template>

          <template v-else-if="activity.variant === 'welfare'">
            <img class="source-layer source-illustrated__bg" :src="activity.images.heroBg" alt="" />
            <img class="source-layer source-float source-welfare__main" :src="activity.images.main" alt="" />
            <img class="source-layer source-float source-welfare__coins" :src="activity.images.coins" alt="" aria-hidden="true" />
            <img class="source-layer source-float source-welfare__chest" :src="activity.images.chest" alt="" />
          </template>

          <template v-else-if="activity.variant === 'sub-deposit'">
            <img class="source-layer source-illustrated__bg" :src="activity.images.heroBg" alt="" />
            <img class="source-layer source-float source-illustrated__main source-illustrated__main--large" :src="activity.images.main" alt="" />
            <img class="source-layer source-sub__ribbon" :src="activity.images.ribbon" alt="" />
          </template>

          <template v-else-if="activity.variant === 'ranking'">
            <div class="source-ranking">
              <article class="source-rank-card source-rank-card--2">
                <img :src="activity.images.top2" alt="" />
                <strong>TOP 2</strong>
                <span>{{ formatActivityMoney(588) }}</span>
              </article>
              <article class="source-rank-card source-rank-card--1">
                <img :src="activity.images.top1" alt="" />
                <strong>TOP 1</strong>
                <span>{{ formatActivityMoney(888) }}</span>
              </article>
              <article class="source-rank-card source-rank-card--3">
                <img :src="activity.images.top3" alt="" />
                <strong>TOP 3</strong>
                <span>{{ formatActivityMoney(388) }}</span>
              </article>
            </div>
          </template>

          <template v-else>
            <img class="source-layer source-rays source-simple__rays" :src="sourceActivityAssets.rays" alt="" aria-hidden="true" />
            <img class="source-layer source-float source-simple__main" :src="activity.images.main" alt="" />
            <img class="source-layer source-float source-simple__cash" :src="sourceActivityAssets.cashBonus" alt="" />
          </template>
        </div>

        <div v-if="activity.variant === 'first-deposit'" class="source-first-limit">
          <svg viewBox="0 0 16 16" aria-hidden="true">
            <path d="M8 1.6a6.4 6.4 0 1 0 0 12.8A6.4 6.4 0 0 0 8 1.6Zm.4 3.7v3.1l2 1.7-.8 1-2.5-2.1V5.3h1.3Z" />
          </svg>
          <span>First deposit only</span>
        </div>

        <button v-if="activity.variant !== 'daily-deposit' && activity.variant !== 'wheel' && !(androidAppRuntime && activity.ctaType === 'download')" type="button" class="source-cta source-pressable" @click="handleCta(activity)">
          <span class="source-shimmer" aria-hidden="true" />
          <strong>{{ activity.cta }}</strong>
          <img v-if="sourceActivityAssets.tapHand" class="source-tap-hand" :src="sourceActivityAssets.tapHand" alt="" aria-hidden="true" />
        </button>
      </section>

      <section ref="rulesRef" class="source-content" style="margin-top: 54px">
        <template v-if="activity.variant === 'deposit-bonus'">
          <article class="source-deposit-panel source-deposit-panel--max">
            <div class="source-deposit-ribbon source-deposit-ribbon--light">
              <span>{{ activity.badge }}</span>
            </div>
            <div class="source-deposit-max-row">
              <div class="source-deposit-reward-icon" aria-hidden="true">
                <img class="source-deposit-reward-icon__rays" :src="sourceActivityAssets.sunburst" alt="" />
                <img class="source-deposit-reward-icon__coins" :src="sourceActivityAssets.coinStack" alt="" />
              </div>
              <div>
                <span>Bonus Rewards</span>
                <strong>10,000.00</strong>
              </div>
            </div>
          </article>

          <article class="source-deposit-panel source-deposit-panel--claim">
            <div class="source-deposit-ribbon source-deposit-ribbon--red">
              <span>Claim Rewards</span>
            </div>
            <div class="source-deposit-claimed">
              <p>
                <img :src="sourceActivityAssets.cash" alt="" aria-hidden="true" />
                <span>Cash Rewards Claimed: 0</span>
              </p>
              <p>
                <img :src="sourceActivityAssets.coinStack" alt="" aria-hidden="true" />
                <span>Bonus Rewards Claimed: 0</span>
              </p>
            </div>
            <div class="source-deposit-empty">
              <svg viewBox="0 0 96 96" aria-hidden="true">
                <ellipse cx="48" cy="92.8" rx="45" ry="2.7" fill="currentColor" opacity="0.12" />
                <path d="M75.3 31.2H27.8c-3.2 0-6 2.2-6.8 5.3l-4.4 17.2c-.2.8-.2 1.6-.1 2.4l.6 4.4A13.5 13.5 0 0 0 20.2 87c5.7 0 10.6-3.5 12.6-8.5h40c3.4 0 6.3-2.5 6.8-5.8l2.7-17.5c.1-.9.1-1.8-.1-2.6l-4.4-16.9a7 7 0 0 0-6.8-4.5Zm-45.7 6.1h39.7c1.6 0 2.9 1.3 2.9 2.9v10.4h-6.5c-.3 0-.6.2-.8.4l-4.3 6.1a6.7 6.7 0 0 1-5.5 2.8H42.8a6.7 6.7 0 0 1-5.5-2.8L33 51a1 1 0 0 0-.8-.4h-7.5V40.2c0-1.6 1.3-2.9 2.9-2.9Zm-9.4 45.6a9.6 9.6 0 1 1 0-19.2 9.6 9.6 0 0 1 0 19.2Zm0-16.8c1 0 1.8.8 1.8 1.8v5.8a1.8 1.8 0 1 1-3.6 0v-5.8c0-1 .8-1.8 1.8-1.8Zm0 11.9a1.8 1.8 0 1 1 0 3.6 1.8 1.8 0 0 1 0-3.6Z" fill="currentColor" />
                <path d="M77 13.9a2.5 2.5 0 0 1 2.4 2.6l-.1 5.2a2.5 2.5 0 1 1-5-.1l.1-5.2A2.5 2.5 0 0 1 77 13.9ZM82.4 24.4l5.2.1a2.5 2.5 0 0 1-.1 5l-5.2-.1a2.5 2.5 0 0 1 .1-5Z" fill="currentColor" />
              </svg>
              <p>Oops! Nothing here for now.</p>
            </div>
          </article>
        </template>

        <template v-else-if="activity.variant === 'wheel'">
          <Transition name="source-wheel-state" mode="out-in">
            <div v-if="!wheelStarted" key="intro" class="source-wheel-shell source-wheel-shell--intro">
              <article class="source-wheel-round-card">
                <h1>{{ activity.title }}</h1>
                <strong>{{ wheelTargetAmountLabel }}</strong>
                <button type="button" class="source-wheel-withdraw source-pressable" :disabled="wheelStarting" @click="startWheelRound">
                  <span v-if="wheelStarting" class="source-wheel-withdraw__spinner" aria-hidden="true" />
                  {{ wheelStarting ? 'Starting' : 'Start Withdraw' }}
                </button>
              </article>

              <article class="source-wheel-record-card">
                <div class="source-wheel-record-tabs">
                  <button type="button" :class="{ 'is-active': wheelRecordTab === 'notice' }" @click="wheelRecordTab = 'notice'">Notice</button>
                  <button type="button" :class="{ 'is-active': wheelRecordTab === 'history' }" @click="wheelRecordTab = 'history'">History</button>
                </div>
                <div v-if="wheelRecordTab === 'notice'" class="source-wheel-record-list">
                  <div class="source-wheel-record-track">
                    <div
                      v-for="(row, index) in wheelNoticeLoopRows"
                      :key="`${row.name}-${row.time}-${index}`"
                      class="source-wheel-record-row"
                    >
                      <span class="source-wheel-record-avatar" :data-tone="row.tone" aria-hidden="true" />
                      <div class="source-wheel-record-meta">
                        <strong>{{ row.name }}</strong>
                      </div>
                      <div class="source-wheel-record-status">
                        <time>{{ row.time }}</time>
                        <em>Withdrawal successful</em>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else class="source-wheel-record-empty">
                  <p>Oops! Nothing here for now.</p>
                </div>
              </article>
            </div>

            <div v-else key="started" class="source-wheel-started">
              <article class="source-wheel-status-card">
                <div class="source-wheel-status-card__head">
                  <h1>Round 1</h1>
                  <div class="source-wheel-countdown" aria-label="Round countdown">
                    <template v-for="(block, index) in wheelCountdown" :key="`${block}-${index}`">
                      <span>{{ block }}</span><i v-if="index < wheelCountdown.length - 1">:</i>
                    </template>
                  </div>
                  <button type="button" class="source-wheel-details source-pressable" @click="scrollToWheelDetails('history')">Details</button>
                </div>
                <strong>{{ wheelCurrentAmountLabel }}</strong>
                <div class="source-wheel-progress" aria-hidden="true">
                  <span class="source-wheel-progress__icon source-wheel-progress__icon--start">
                    <img :src="activity.images.progressCheck" alt="" />
                  </span>
                  <i :style="wheelProgressStyle" />
                  <span class="source-wheel-progress__icon source-wheel-progress__icon--mid">
                    <img :src="wheelMidProgressComplete ? activity.images.progressCheck : activity.images.progressCoin" alt="" />
                  </span>
                  <span class="source-wheel-progress__icon source-wheel-progress__icon--end">
                    <img :src="wheelEndProgressComplete ? activity.images.progressCheck : activity.images.progressCoin" alt="" />
                  </span>
                </div>
                <p v-if="wheelCurrentAmount <= 0">Withdraw instantly at <b>{{ wheelTargetAmountLabel }}</b></p>
                <p v-else>Just {{ wheelRemainingAmountLabel }} more to withdraw {{ wheelTargetAmountLabel }}</p>
              </article>

              <section class="source-wheel-game" aria-label="Lucky wheel">
                <img class="source-wheel-game__star source-wheel-game__star--left" :src="activity.images.glowingStar" alt="" aria-hidden="true" />
                <img class="source-wheel-game__star source-wheel-game__star--right" :src="activity.images.glowingStar" alt="" aria-hidden="true" />
                <div
                  class="source-wheel-game__rotor"
                  :class="{ 'is-spinning': wheelSpinning }"
                  :style="wheelSpinStyle"
                  aria-hidden="true"
                >
                  <img class="source-wheel-game__wheel" :src="activity.images.wheel" alt="" />
                  <span
                    v-for="(label, index) in wheelLabels"
                    :key="label"
                    class="source-wheel-game__label"
                    :style="{ transform: `translate(-50%, -50%) rotate(${index * 45}deg) translateY(calc(var(--wheel-size) * -0.355))` }"
                  >
                    {{ label }}
                  </span>
                  <img
                    v-for="(prize, index) in wheelPrizes"
                    :key="`${prize}-${index}`"
                    class="source-wheel-game__prize"
                    :src="prize"
                    alt=""
                    :style="{ transform: `translate(-50%, -50%) rotate(${index * 45}deg) translateY(calc(var(--wheel-size) * -0.225))` }"
                  />
                </div>
                <button type="button" class="source-wheel-game__center source-pressable" :disabled="wheelSpinning" @click="spinWheel">
                  <img :src="activity.images.luckyPointer" alt="Lucky" />
                </button>
                <img v-if="activity.images.pedestal" class="source-wheel-game__pedestal" :src="activity.images.pedestal" alt="" aria-hidden="true" />
                <Transition name="source-wheel-floating">
                  <div v-if="wheelFloatingAmount" class="source-wheel-floating">
                    {{ wheelFloatingAmount }}
                  </div>
                </Transition>
              </section>

              <div class="source-wheel-action-row">
                <button type="button" class="source-wheel-action source-wheel-action--share source-pressable">
                  <span>Share for Spins</span>
                </button>
                <button type="button" class="source-wheel-action source-wheel-action--free source-pressable" :disabled="wheelSpinning" @click="spinWheel">
                  <span>{{ wheelSpinning ? 'Spinning' : 'Free Spin' }}</span>
                </button>
              </div>

              <div class="source-wheel-spin-meta">
                <span>Remaining:</span>
                <b>{{ wheelRemainingSpins }}</b>
                <button type="button" aria-label="Refresh" class="source-pressable">↻</button>
                <button type="button" class="source-pressable" @click="scrollToWheelDetails('history')">History</button>
                <button type="button" class="source-pressable" @click="scrollToWheelDetails('history')">Records</button>
              </div>
              <div class="source-wheel-tip">💡 Earn 1 spins by sharing daily</div>

              <article ref="wheelDetailsRef" class="source-wheel-record-card source-wheel-record-card--started">
                <div class="source-wheel-record-tabs">
                  <button type="button" :class="{ 'is-active': wheelRecordTab === 'notice' }" @click="wheelRecordTab = 'notice'">Notice</button>
                  <button type="button" :class="{ 'is-active': wheelRecordTab === 'history' }" @click="wheelRecordTab = 'history'">History</button>
                </div>
                <div v-if="wheelRecordTab === 'notice'" class="source-wheel-record-list">
                  <div class="source-wheel-record-track">
                    <div
                      v-for="(row, index) in wheelNoticeLoopRows"
                      :key="`started-${row.name}-${row.time}-${index}`"
                      class="source-wheel-record-row"
                    >
                      <span class="source-wheel-record-avatar" :data-tone="row.tone" aria-hidden="true" />
                      <div class="source-wheel-record-meta">
                        <strong>{{ row.name }}</strong>
                      </div>
                      <div class="source-wheel-record-status">
                        <time>{{ row.time }}</time>
                        <em>Withdrawal successful</em>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else class="source-wheel-history-panel">
                  <article class="source-wheel-history-card">
                    <header>
                      <h2>Round 1</h2>
                      <strong>{{ wheelHistory.amount }}</strong>
                    </header>
                    <dl>
                      <div>
                        <dt>Start Time</dt>
                        <dd>{{ wheelHistory.startTime }}</dd>
                      </div>
                      <div>
                        <dt>End Time</dt>
                        <dd>{{ wheelHistory.endTime }}</dd>
                      </div>
                      <div>
                        <dt>Status</dt>
                        <dd class="is-pending">{{ wheelHistory.status }}</dd>
                      </div>
                    </dl>
                  </article>
                </div>
              </article>
            </div>
          </Transition>
        </template>

        <template v-else>
        <div v-if="activity.variant !== 'daily-deposit'" class="source-content__banner">
          <img
            v-if="activity.images.rulesBanner || activity.images.ribbon || activity.images.desc"
            :src="activity.variant === 'first-deposit' ? activity.images.desc : (activity.images.rulesBanner || activity.images.ribbon)"
            alt=""
          />
          <svg v-if="activity.variant === 'first-deposit'" class="source-content__arc-title" viewBox="0 0 305 92" aria-hidden="true">
            <defs>
              <path id="first-reward-title-arc" d="M 44 61 Q 152.5 34 261 61" />
            </defs>
            <text>
              <textPath href="#first-reward-title-arc" startOffset="50%" text-anchor="middle">{{ activity.ruleTitle }}</textPath>
            </text>
          </svg>
          <strong v-else>{{ activity.ruleTitle }}</strong>
        </div>

        <div v-if="activity.variant === 'daily-deposit'" class="source-daily-panel">
          <div class="source-daily-panel__title">Today's cashback ({{ todayDisplay }})</div>
          <div class="source-daily-reward-row">
            <div class="source-daily-reward">
              <img class="source-daily-reward__burst" :src="sourceActivityAssets.sunburst" alt="" aria-hidden="true" />
              <img class="source-daily-reward__icon" :src="sourceActivityAssets.bonusIcon" alt="" aria-hidden="true" />
              <strong>* 0.00</strong>
            </div>
            <div class="source-daily-reward">
              <img class="source-daily-reward__burst" :src="sourceActivityAssets.sunburst" alt="" aria-hidden="true" />
              <img class="source-daily-reward__icon source-daily-reward__icon--cash" :src="sourceActivityAssets.cash" alt="" aria-hidden="true" />
              <strong>* 0.00</strong>
            </div>
          </div>
          <div class="source-daily-claim">
            <span>Claimed Time:</span>
            <strong>{{ dailyClaimRange }}</strong>
          </div>
          <p class="source-daily-progress-copy">
            Deposit 0.00/3,000.00 to get <strong>0.50%</strong> rebate (Max 4.00%)
          </p>
          <div class="source-daily-meter"><i /></div>
          <button type="button" class="source-daily-increase source-pressable" @click="handleCta(activity)">
            <img :src="activity.images.button" alt="" aria-hidden="true" />
            <span>Increase Rebate Rate</span>
          </button>
          <img class="source-daily-corner source-daily-corner--left" :src="activity.images.left" alt="" aria-hidden="true" />
          <img class="source-daily-corner source-daily-corner--right" :src="activity.images.right" alt="" aria-hidden="true" />
        </div>

        <div v-else-if="activity.variant === 'relief'" class="source-panel source-progress-panel">
          <div>
            <span>0.50% Cashback Daily</span>
            <strong>Your Cashback Today {{ formatActivityMoney(0) }}</strong>
          </div>
          <p>Valid Bet 0 / 100 to unlock 0.50%</p>
          <div class="source-meter"><i style="width: 6%" /></div>
          <em>Claimed Time: {{ dailyClaimRange }}</em>
        </div>

        <div v-else-if="activity.variant === 'periodic-green' || activity.variant === 'periodic-gold'" class="source-panel source-cycle-panel">
          <div class="source-cycle-panel__head">
            <span>Current Period</span>
            <strong>{{ weeklyPeriodRange }}</strong>
          </div>
          <div class="source-cycle-panel__stats">
            <span>Minimum Reward <b>{{ formatActivityMoney(5) }}</b></span>
            <span>Maximum Reward <b>{{ formatActivityMoney(70) }}</b></span>
          </div>
          <div class="source-countdown">
            <span v-for="block in countdownBlocks" :key="block.label">
              <strong>{{ block.value }}</strong>
              <em>{{ block.label }}</em>
            </span>
          </div>
          <div class="source-meter"><i style="width: 18%" /></div>
          <em>Progress 0/3 deposits · Resets every Monday</em>
        </div>

        <div v-else-if="activity.variant === 'sub-deposit'" class="source-panel source-cycle-panel">
          <div class="source-cycle-panel__head">
            <span>Current Reward Cycle</span>
            <strong>{{ dailyClaimRange }}</strong>
          </div>
          <div class="source-cycle-panel__stats">
            <span>Direct Subordinates' Total Top-up <b>{{ formatActivityMoney(0) }}</b></span>
            <span>Total Bonus <b>{{ formatActivityMoney(0) }}</b></span>
          </div>
          <div class="source-countdown">
            <span v-for="block in countdownBlocks" :key="block.label">
              <strong>{{ block.value }}</strong>
              <em>{{ block.label }}</em>
            </span>
          </div>
        </div>

        <div v-else-if="activity.variant === 'chest'" class="source-panel source-chest-panel">
          <strong>Your earned bonus</strong>
          <span>Total Rewards Earned {{ formatActivityMoney(0) }}</span>
          <p>Your friends can receive a {{ formatActivityMoney(10) }} rewards!</p>
          <div class="source-chest-steps">
            <i v-for="step in chestSteps" :key="step" :class="{ 'is-active': step === 1 }">{{ step }}</i>
          </div>
        </div>

        <div v-else-if="activity.variant === 'red-packet'" class="source-panel source-red-panel">
          <img :src="activity.images.rulesPackets" alt="" aria-hidden="true" />
          <div>
            <strong>Catch red packets in the active round</strong>
            <span>Round countdown · Lucky cash · Bonus packets</span>
          </div>
        </div>

        <div v-else-if="activity.variant === 'welfare'" class="source-welfare-tabs">
          <button type="button" class="is-active">Card List</button>
          <button type="button">My Cards</button>
        </div>

        <div v-if="activity.variant === 'daily-deposit'" class="source-daily-tabs">
          <button type="button" :class="{ 'is-active': dailyCashbackTab === 'level' }" @click="dailyCashbackTab = 'level'">
            Rebate Level Description
          </button>
          <button type="button" :class="{ 'is-active': dailyCashbackTab === 'records' }" @click="dailyCashbackTab = 'records'">
            Rebate Records
          </button>
        </div>

        <template v-if="activity.variant === 'daily-deposit'">
          <template v-if="dailyCashbackTab === 'level'">
            <div class="source-daily-table-title">
              <span>RETURN LEVEL DESCRIPTION</span>
            </div>
            <div class="source-daily-table">
              <div class="source-daily-table__head">
                <span>Level</span>
                <span>Deposit Amount</span>
                <span>Cash</span>
                <span>Bonus</span>
                <span>Total</span>
              </div>
              <div
                v-for="(row, index) in activity.rows"
                :key="`${row.label}-${row.extra}`"
                class="source-daily-table__row"
              >
                <span>{{ index + 1 }}</span>
                <span>{{ row.label }}</span>
                <span>{{ row.value }}</span>
                <span>{{ row.extra }}</span>
                <span>{{ row.extra }}</span>
              </div>
            </div>
          </template>
          <div v-else class="source-daily-empty">
            <svg viewBox="0 0 96 96" aria-hidden="true">
              <ellipse cx="48" cy="92.8" rx="45" ry="2.7" fill="currentColor" opacity="0.12" />
              <path d="M75.3 31.2H27.8c-3.2 0-6 2.2-6.8 5.3l-4.4 17.2c-.2.8-.2 1.6-.1 2.4l.6 4.4A13.5 13.5 0 0 0 20.2 87c5.7 0 10.6-3.5 12.6-8.5h40c3.4 0 6.3-2.5 6.8-5.8l2.7-17.5c.1-.9.1-1.8-.1-2.6l-4.4-16.9a7 7 0 0 0-6.8-4.5Zm-45.7 6.1h39.7c1.6 0 2.9 1.3 2.9 2.9v10.4h-6.5c-.3 0-.6.2-.8.4l-4.3 6.1a6.7 6.7 0 0 1-5.5 2.8H42.8a6.7 6.7 0 0 1-5.5-2.8L33 51a1 1 0 0 0-.8-.4h-7.5V40.2c0-1.6 1.3-2.9 2.9-2.9Zm-9.4 45.6a9.6 9.6 0 1 1 0-19.2 9.6 9.6 0 0 1 0 19.2Zm0-16.8c1 0 1.8.8 1.8 1.8v5.8a1.8 1.8 0 1 1-3.6 0v-5.8c0-1 .8-1.8 1.8-1.8Zm0 11.9a1.8 1.8 0 1 1 0 3.6 1.8 1.8 0 0 1 0-3.6Z" fill="currentColor" />
              <path d="M77 13.9a2.5 2.5 0 0 1 2.4 2.6l-.1 5.2a2.5 2.5 0 1 1-5-.1l.1-5.2A2.5 2.5 0 0 1 77 13.9ZM82.4 24.4l5.2.1a2.5 2.5 0 0 1-.1 5l-5.2-.1a2.5 2.5 0 0 1 .1-5Z" fill="currentColor" />
            </svg>
            <p>Oops! Nothing here for now.</p>
          </div>
        </template>

        <template v-else>
        <div v-if="tableHead.length" class="source-table-head">
          <span v-for="head in tableHead" :key="head">{{ head }}</span>
        </div>

        <div class="source-card-grid" :class="cardGridClasses">
          <article
            v-for="(row, index) in activity.rows"
            :key="`${row.label}-${row.value}`"
            class="source-card source-pressable"
            :class="{ 'source-card--welfare': activity.variant === 'welfare' }"
          >
            <img v-if="activity.variant === 'welfare'" class="source-card__thumb" :src="welfareCardThumb(index)" alt="" />
            <template v-if="activity.variant === 'first-deposit'">
              <div class="source-first-reward-pair">
                <div>
                  <img class="source-first-reward-pair__rays" :src="activity.images.lightBonus" alt="" aria-hidden="true" />
                  <img class="source-first-reward-pair__icon" :src="sourceActivityAssets.coinStack" alt="" aria-hidden="true" />
                  <strong>{{ row.value }}</strong>
                </div>
                <div>
                  <img class="source-first-reward-pair__rays" :src="activity.images.lightCash" alt="" aria-hidden="true" />
                  <img class="source-first-reward-pair__icon" :src="sourceActivityAssets.cash" alt="" aria-hidden="true" />
                  <strong>{{ row.extra }}</strong>
                </div>
              </div>
              <span class="source-first-amount">{{ row.label }}</span>
            </template>
            <div v-else class="source-card__copy">
              <span>{{ row.label }}</span>
              <strong>{{ row.value }}</strong>
              <em>{{ row.extra }}</em>
            </div>
          </article>
        </div>

        <article v-if="activity.variant !== 'first-deposit'" class="source-rule-card">
          <h2>Event Rules</h2>
          <ol>
            <li v-for="rule in activity.rules" :key="rule">{{ rule }}</li>
          </ol>
        </article>
        </template>
        </template>
      </section>
    </main>

    <ContactMeBadge class="source-contact" :image="customerImage" draggable dock-on-idle @click="openContact" />

    <img
      v-if="activity.variant === 'first-deposit' && activity.images.banner"
      class="source-first-fixed-treasure"
      :src="activity.images.banner"
      alt=""
      aria-hidden="true"
    />

    <Transition name="source-rules-dialog">
      <div
        v-if="(activity.variant === 'first-deposit' || activity.variant === 'deposit-bonus' || activity.variant === 'daily-deposit' || activity.variant === 'wheel') && rulesDialogVisible"
        class="source-rules-dialog"
        :class="`source-rules-dialog--${activity.variant}`"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="`${activity.variant}-rules-title`"
        @click.self="closeRules"
      >
        <div class="source-rules-dialog__body">
          <div class="source-rules-dialog__bundle" aria-hidden="true">
            <img :src="activity.images.bundle || activity.images.heroBox" alt="" />
          </div>
          <div class="source-rules-dialog__content">
            <div class="source-rules-dialog__banner" :id="`${activity.variant}-rules-title`">
              <img :src="activity.images.rulesBanner" alt="" aria-hidden="true" />
              <span>Event Rules</span>
            </div>
            <div class="source-rules-dialog__frame">
              <div class="source-rules-dialog__inner">
                <div class="source-rules-dialog__scroll">
                  <ol>
                    <li v-for="rule in activity.rules" :key="rule">
                      <p v-html="rule" />
                    </li>
                  </ol>
                </div>
              </div>
            </div>
          </div>
          <div class="source-rules-dialog__actions">
            <button type="button" class="source-rules-dialog__ok source-pressable" @click="closeRules">
              <span class="source-rules-dialog__ok-base" aria-hidden="true" />
              <span class="source-rules-dialog__ok-top" aria-hidden="true" />
              <span class="source-rules-dialog__ok-shine" aria-hidden="true" />
              <span class="source-rules-dialog__ok-spark" aria-hidden="true" />
              <span class="source-rules-dialog__ok-label">Okay</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <ActivityGuideFlow
      v-if="guideVisible"
      :visible="guideVisible"
      initial="app-download"
      @complete="closeGuide"
    />
  </section>

  <section v-else class="activity-not-found">
    <header>
      <button type="button" aria-label="Back" @click="goBack">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M4.29 11.29a1 1 0 0 0 0 1.42l4 4a1 1 0 0 0 1.42-1.42L7.41 13H19a1 1 0 1 0 0-2H7.41l2.3-2.29a1 1 0 0 0-1.42-1.42l-4 4Z" />
        </svg>
      </button>
      <h1>Page Not Found</h1>
    </header>
    <button type="button" @click="router.push({ name: 'offers', query: { tab: 'events' } })">
      Back to Offers
    </button>
  </section>
</template>

<style scoped lang="scss">
.source-activity,
.activity-not-found {
  position: relative;
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  max-width: 430px;
  margin: 0 auto;
  overflow: hidden;
  color: #ffffff;
  background: #070d1c;
  font-family: Montserrat, Poppins, Arial, Helvetica, sans-serif;
}

.source-scroll {
  height: 100vh;
  height: 100dvh;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: none;
  background:
    var(--source-page-bg) center top / 100% auto no-repeat,
    radial-gradient(circle at 50% 0%, rgba(69, 112, 255, 0.36), transparent 39%),
    linear-gradient(180deg, #101632 0%, #0a1022 54%, #070d1c 100%);
  background-attachment: local, local, local;
}

.source-activity--red .source-scroll {
  background:
    var(--source-page-bg) center top / 100% auto no-repeat,
    radial-gradient(circle at 50% 2%, rgba(255, 189, 77, 0.42), transparent 36%),
    linear-gradient(180deg, #5f1718 0%, #2d0821 51%, #090915 100%);
}

.source-activity--green .source-scroll {
  background:
    var(--source-page-bg) center top / 100% auto no-repeat,
    radial-gradient(circle at 50% 2%, rgba(50, 255, 198, 0.35), transparent 36%),
    linear-gradient(180deg, #123f35 0%, #17316c 48%, #070d1c 100%);
}

.source-activity--gold .source-scroll {
  background:
    var(--source-page-bg) center top / 100% auto no-repeat,
    radial-gradient(circle at 50% 2%, rgba(255, 218, 91, 0.42), transparent 35%),
    linear-gradient(180deg, #563915 0%, #331a44 48%, #070d1c 100%);
}

.source-activity--purple .source-scroll {
  background:
    var(--source-page-bg) center top / 100% auto no-repeat,
    radial-gradient(circle at 50% 3%, rgba(198, 96, 255, 0.5), transparent 38%),
    linear-gradient(180deg, #34106f 0%, #0e1845 54%, #070d1c 100%);
}

.source-activity--first-deposit .source-scroll {
  background:
    var(--source-page-bg) center top / 100% auto no-repeat,
    radial-gradient(circle at 50% 4%, rgba(174, 99, 255, 0.38), transparent 42%),
    linear-gradient(180deg, #6f32d4 0%, #4d1395 42%, #30076b 72%, #220548 100%);
  background-color: #220548;
}

.source-activity--deposit-bonus .source-scroll {
  background:
    var(--source-page-bg) center top / 100% auto repeat-y,
    linear-gradient(180deg, #631415 0%, #4a1011 48%, #270708 100%);
  background-color: #4a1011;
}

.source-scroll,
.source-activity--red .source-scroll,
.source-activity--green .source-scroll,
.source-activity--gold .source-scroll,
.source-activity--purple .source-scroll,
.source-activity--first-deposit .source-scroll {
  background-attachment: local, local, local;
}

.source-nav {
  position: fixed;
  top: max(8px, env(safe-area-inset-top));
  left: max(calc((100vw - 430px) / 2 + 14px), 14px);
  right: max(calc((100vw - 430px) / 2), 0px);
  z-index: 20;
  display: flex;
  height: 46px;
  align-items: center;
  justify-content: space-between;
  pointer-events: none;

  button {
    pointer-events: auto;
  }

  button:first-child {
    display: grid;
    width: 30px;
    height: 30px;
    place-items: center;
    border: 1px solid rgba(255, 255, 255, 0.52);
    border-radius: 999px;
    color: #ffffff;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(8px);
    filter: drop-shadow(0 2px 5px rgba(0, 0, 0, 0.28));
  }

  button:last-child {
    height: 27px;
    padding: 0 12px;
    border: 1px solid rgba(255, 255, 255, 0.55);
    border-right: 0;
    border-radius: 14px 0 0 14px;
    color: #ffffff;
    font-size: 12px;
    font-weight: 600;
    background: rgba(255, 255, 255, 0.16);
    box-shadow:
      inset -3px 0 3px rgba(0, 0, 0, 0.12),
      0 2px 7px rgba(0, 0, 0, 0.28);
    backdrop-filter: blur(8px);
  }

  svg {
    width: 19px;
    height: 19px;
    fill: currentColor;
  }
}

.source-hero {
  position: relative;
  min-height: 620px;
  overflow: hidden;
  padding: calc(env(safe-area-inset-top) + 54px) 14px 0;
  isolation: isolate;
}

.source-hero::before {
  position: absolute;
  top: -120px;
  right: -90px;
  left: -90px;
  z-index: -1;
  height: 360px;
  background:
    repeating-conic-gradient(from -16deg at 50% 8%, rgba(255, 255, 255, 0.23) 0deg 8deg, transparent 8deg 18deg),
    radial-gradient(circle at 50% 13%, rgba(255, 255, 255, 0.82), transparent 8%),
    radial-gradient(circle at 50% 18%, rgba(87, 136, 255, 0.52), transparent 48%);
  mix-blend-mode: screen;
  opacity: 0.6;
  content: "";
}

.source-activity--first-deposit .source-hero {
  z-index: 3;
  min-height: 562px;
  padding: calc(env(safe-area-inset-top) + 30px) 14px 0;
}

.source-activity--first-deposit .source-hero::before {
  display: none;
}

.source-activity--deposit-bonus .source-hero {
  min-height: 492px;
  overflow: visible;
  padding: 0;
}

.source-activity--deposit-bonus .source-hero::before {
  display: none;
}

.source-activity--daily-deposit .source-scroll {
  background:
    var(--source-page-bg) center top / 100% auto repeat-y,
    linear-gradient(180deg, #42100c 0%, #3a0a08 68%, #14487a 100%);
  background-color: #14487a;
}

.source-activity--daily-deposit .source-hero {
  min-height: 213px;
  overflow: visible;
  padding: 0;
}

.source-activity--daily-deposit .source-hero::before {
  display: none;
}

.source-activity--wheel {
  background: #93d5e0;
}

.source-activity--wheel .source-scroll {
  position: relative;
  overflow-y: auto;
  background: linear-gradient(180deg, #6753f0 5%, #c3edff 19%, #d0f4a9 52%, #9ed2e7 67%, #8ec9dc 100%);
  background-attachment: local;
}

.source-activity--wheel.is-wheel-started .source-scroll {
  overflow-y: auto;
}

.source-activity--wheel .source-scroll::before {
  position: absolute;
  top: -30px;
  right: 0;
  left: 0;
  z-index: 0;
  height: 710px;
  background: var(--source-page-bg) center top / 100% auto no-repeat;
  pointer-events: none;
  transform-origin: 50% 0;
  animation: source-wheel-flags-sway 2.9s ease-in-out infinite alternate;
  content: "";
  will-change: transform, filter;
}

.source-activity--wheel .source-hero {
  min-height: 0;
  height: 0;
  overflow: visible;
  padding: 0;
}

.source-activity--wheel .source-hero::before,
.source-activity--wheel .source-stage {
  display: none;
}

.source-wheel-copy,
.source-wheel__stage-anchor {
  display: none;
}

.source-daily-heading {
  position: absolute;
  top: calc(env(safe-area-inset-top) + 66px);
  left: 20px;
  z-index: 6;
  width: 270px;
  pointer-events: none;

  h1 {
    margin: 0;
    color: #fff7a3;
    font-size: 24px;
    font-style: italic;
    font-weight: 900;
    letter-spacing: 0;
    line-height: 1.08;
    text-shadow:
      0 2px 0 #5b1f16,
      0 4px 10px rgba(36, 0, 0, 0.36);
  }

  p {
    width: 198px;
    margin: 11px 0 0;
    color: #ffffff;
    font-size: 12px;
    font-style: italic;
    font-weight: 500;
    line-height: 1.23;
    text-shadow: 0 2px 3px rgba(33, 0, 0, 0.42);
  }
}

.source-deposit-heading {
  position: absolute;
  top: calc(env(safe-area-inset-top) + 80px);
  right: 24px;
  left: 24px;
  z-index: 7;
  text-align: center;
  pointer-events: none;

  h1 {
    margin: 0;
    color: #fff994;
    font-size: clamp(26px, 7.5vw, 31px);
    font-style: italic;
    font-weight: 900;
    letter-spacing: 0;
    line-height: 1.13;
    text-shadow:
      0 3px 0 rgba(101, 25, 26, 0.92),
      0 7px 13px rgba(31, 5, 6, 0.4);
  }
}

.source-first-heading {
  position: relative;
  z-index: 5;
  display: flex;
  min-height: 298px;
  flex-direction: column;
  align-items: center;
  text-align: center;

  p {
    margin: 0 0 11px;
    color: rgba(255, 255, 255, 0.92);
    font-size: 15px;
    font-weight: 800;
    line-height: 1.2;
    text-shadow: 0 2px 3px rgba(24, 4, 86, 0.42);
  }
}

.source-first-title {
  position: relative;
  width: min(402px, 100%);
  min-height: 77px;
  margin: 22px auto 8px;
  color: #ffffff;
  font-family: Montserrat, sans-serif;
  font-size: 32px;
  font-style: italic;
  font-weight: 900;
  letter-spacing: 0;
  line-height: 1.2;
  text-align: center;

  span {
    display: block;
    width: 100%;
  }
}

.source-first-title__shadow {
  position: absolute;
  inset: 0;
  z-index: 0;
  color: #132971;
  transform: translate(-1px, 2px);
  -webkit-text-fill-color: #132971;
  -webkit-text-stroke: 2px #132971;
}

.source-first-title__stroke {
  position: relative;
  z-index: 1;
  color: #b83dcc;
  -webkit-text-fill-color: #b83dcc;
  -webkit-text-stroke: 3px #b83dcc;
}

.source-first-title__fill {
  position: absolute;
  inset: 0;
  z-index: 2;
  color: transparent;
  background: linear-gradient(180deg, #ffffff 11.43%, #fff3c4 42.93%, #ffffff 61.68%, #fff3c4 96.35%);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  -webkit-text-stroke: 0;
}

.source-first-percent {
  position: relative;
  display: flex;
  align-items: baseline;
  gap: 6px;
  transform: none;
  filter: drop-shadow(4px 6px 0 rgba(74, 23, 147, 0.68));

  span,
  em {
    position: relative;
    z-index: 2;
    display: block;
    color: transparent;
    background: linear-gradient(180deg, #ffefa6 17%, #ffd28d 75%);
    background-clip: text;
    font-style: italic;
    line-height: 0.95;
    text-shadow: none;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  span {
    font-size: clamp(47px, 12.8vw, 56px);
    font-weight: 800;
    -webkit-text-stroke: 2px #d4a04a;
  }

  em {
    font-size: clamp(15px, 4.2vw, 19px);
    font-weight: 700;
    -webkit-text-stroke: 1px #d4a04a;
  }
}

.source-first-max {
  display: inline-flex;
  min-height: 42px;
  align-items: center;
  justify-content: center;
  margin-top: 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 999px;
  padding: 0 22px;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.14);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.22),
    0 8px 16px rgba(55, 15, 129, 0.18);
  font-size: 16px;
  font-weight: 800;
  backdrop-filter: blur(5px);

  strong {
    margin-left: 4px;
    font-size: 24px;
    font-weight: 900;
  }
}

.source-hero__copy {
  position: relative;
  z-index: 4;
  min-height: 104px;
  text-align: center;

  p {
    margin: 0 0 7px;
    color: rgba(255, 255, 255, 0.82);
    font-size: 13px;
    font-weight: 700;
    line-height: 1.25;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.25);
  }

  h1 {
    max-width: 345px;
    margin: 0 auto;
    color: #fff994;
    font-size: clamp(29px, 9vw, 38px);
    font-style: italic;
    font-weight: 900;
    letter-spacing: 0;
    line-height: 1.12;
    text-shadow:
      0 3px 0 rgba(92, 16, 112, 0.85),
      0 7px 13px rgba(32, 4, 67, 0.48);
  }
}

.source-badge {
  position: relative;
  z-index: 4;
  width: max-content;
  max-width: 92%;
  margin: 0 auto 8px;
  padding: 8px 18px;
  border: 1px solid rgba(255, 255, 255, 0.38);
  border-radius: 999px;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.16);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 8px 20px rgba(0, 0, 0, 0.18);
  backdrop-filter: blur(5px);

  span {
    display: block;
    font-size: 16px;
    font-weight: 900;
    line-height: 1;
    text-shadow: 0 2px 2px rgba(0, 0, 0, 0.28);
  }
}

.source-stage {
  position: relative;
  z-index: 3;
  height: 342px;
  margin: 0 -14px;
}

.source-stage--first {
  width: min(100vw, 430px);
  height: min(249px, 57.95vw);
  margin: -146px -14px 0;
}

.source-layer {
  position: absolute;
  object-fit: contain;
  pointer-events: none;
  user-select: none;
}

.source-rays {
  mix-blend-mode: screen;
}

.source-pressable {
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  transition:
    transform 0.16s ease,
    filter 0.16s ease;

  &:active {
    transform: translateY(2px) scale(0.97);
    filter: brightness(0.92);
  }
}

.source-cta {
  position: relative;
  z-index: 8;
  display: flex;
  width: 190px;
  height: 46px;
  align-items: center;
  justify-content: center;
  margin: -18px auto 0;
  overflow: hidden;
  border: 0;
  border-radius: 999px;
  color: #843e00;
  background: linear-gradient(180deg, #fff2a6 0%, #ffd63f 50%, #fb8a23 100%);
  box-shadow:
    inset 0 2px 4px rgba(255, 255, 255, 0.82),
    inset 0 -3px 0 rgba(143, 52, 0, 0.18),
    0 7px 18px rgba(0, 0, 0, 0.24);
  font-size: 17px;
  font-weight: 900;
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.54);

  strong {
    position: relative;
    z-index: 2;
  }
}

.source-activity--first-deposit .source-cta {
  width: 190px;
  height: 43px;
  margin-top: 10px;
  overflow: visible;
  color: #973305;
  background: linear-gradient(180deg, #fff0b0 0%, #ffbc18 62%, #fe762d 100%);
  box-shadow:
    inset 0 3px 5px rgba(255, 255, 255, 0.76),
    inset 0 -4px 4px rgba(255, 255, 255, 0.24),
    0 4px 16px rgba(0, 0, 0, 0.14);
  font-size: 17px;
}

.source-activity--first-deposit .source-tap-hand {
  right: -37px;
  bottom: -27px;
  width: 68px;
}

.source-tap-hand {
  position: absolute;
  right: -28px;
  bottom: -15px;
  z-index: 4;
  width: 58px;
  pointer-events: none;
  filter: drop-shadow(0 5px 7px rgba(0, 0, 0, 0.3));
}

.source-shimmer {
  position: absolute;
  top: -18px;
  bottom: -18px;
  left: -70px;
  width: 42px;
  background: rgba(255, 255, 255, 0.55);
  filter: blur(4px);
  transform: rotate(19deg);
}

.source-first__bundle {
  left: 50%;
  top: 36.7%;
  z-index: 4;
  width: 61.3%;
  translate: -50% 0;
  filter: drop-shadow(0 10px 14px rgba(36, 8, 81, 0.28));
}

.source-first__gift-float {
  left: -0.8%;
  top: 57.5%;
  z-index: 5;
  width: 25.9%;
  filter: drop-shadow(0 8px 10px rgba(42, 7, 82, 0.28));
}

.source-first__treasure-float {
  right: 1%;
  top: 86.3%;
  z-index: 6;
  width: 21%;
  filter: drop-shadow(0 8px 10px rgba(42, 7, 82, 0.28));
}

.source-first__coin-float {
  right: 5.9%;
  top: 55.7%;
  z-index: 5;
  width: 13.8%;
  filter: drop-shadow(0 8px 10px rgba(42, 7, 82, 0.24));
}

.source-first-limit {
  position: relative;
  z-index: 7;
  display: flex;
  min-width: 236px;
  width: max-content;
  height: 30px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 42px auto 0;
  border-radius: 999px;
  color: #ffffff;
  background:
    linear-gradient(90deg, rgba(183, 74, 46, 0), rgba(152, 94, 71, 0.72) 24%, rgba(202, 184, 141, 0.7) 76%, rgba(183, 74, 46, 0)),
    rgba(255, 255, 255, 0.05);
  font-size: 13px;
  font-weight: 700;
  text-shadow: 0 2px 3px rgba(55, 11, 83, 0.35);

  svg {
    width: 16px;
    height: 16px;
    fill: currentColor;
  }
}

.source-first__box {
  z-index: 1;
  left: 50%;
  bottom: 26px;
  width: min(390px, 108vw);
  transform: translateX(-50%);
}

.source-first__cash-rays {
  z-index: 2;
  left: 19%;
  top: 70px;
  width: 118px;
}

.source-first__bonus-rays {
  z-index: 2;
  right: 17%;
  top: 78px;
  width: 112px;
}

.source-first__gift {
  z-index: 6;
  left: 50%;
  top: 48px;
  width: 142px;
  transform: translateX(-50%);
  filter: drop-shadow(0 10px 12px rgba(107, 16, 20, 0.35));
}

.source-first__treasure {
  z-index: 6;
  left: 20px;
  bottom: 108px;
  width: 104px;
}

.source-first__coin {
  z-index: 6;
  right: 40px;
  bottom: 118px;
  width: 52px;
}

.source-first__banner {
  z-index: 7;
  left: 50%;
  bottom: 94px;
  width: min(330px, 92vw);
  transform: translateX(-50%);
}

.source-first__banner-text {
  z-index: 8;
  left: 50%;
  bottom: 108px;
  width: max-content;
  transform: translateX(-50%);
  color: #fff994;
  font-size: 17px;
  font-style: italic;
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 2px 2px rgba(87, 16, 99, 0.72);
}

.source-first__reward-frame {
  z-index: 10;
  right: 32px;
  bottom: 24px;
  width: 122px;
  height: 176px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  span,
  strong,
  em {
    position: relative;
    z-index: 1;
    display: block;
    width: 100%;
    padding: 0 9px 0 18px;
    color: #ffffff;
    line-height: 1.1;
    text-shadow: 0 2px 3px rgba(25, 5, 62, 0.48);
  }

  span {
    margin-top: 54px;
    font-size: 8px;
    font-weight: 800;
  }

  strong {
    margin-top: 6px;
    color: #fff994;
    font-size: 22px;
    font-weight: 900;
  }

  em {
    margin-top: 8px;
    color: rgba(255, 255, 255, 0.78);
    font-size: 8px;
    font-style: normal;
    font-weight: 800;
  }
}

.source-first__desc {
  z-index: 9;
  right: 42px;
  bottom: 90px;
  width: 102px;
}

.source-deposit__bg {
  inset: 0 0 auto;
  z-index: 1;
  width: 100%;
  height: 518px;
  object-fit: cover;
}

.source-deposit__light {
  left: 50%;
  top: 0;
  z-index: 2;
  width: min(390px, 100vw);
  opacity: 0.88;
  transform: translateX(-50%);
  mix-blend-mode: screen;
}

.source-deposit__pile {
  left: 50%;
  top: 282px;
  z-index: 3;
  width: 100%;
  clip-path: inset(0 0 96px 0);
  transform: translateX(-50%);
}

.source-deposit__lower-bg {
  inset: 358px 0 auto;
  z-index: 4;
  display: block;
  width: 100%;
  height: 178px;
  background:
    radial-gradient(circle at 8% 28%, rgba(255, 176, 31, 0.18) 0 18px, transparent 34px),
    radial-gradient(circle at 92% 34%, rgba(255, 222, 79, 0.15) 0 22px, transparent 42px),
    linear-gradient(180deg, #850704 0%, #650403 38%, #400303 100%);
  box-shadow: inset 0 22px 30px rgba(66, 0, 0, 0.18);
  mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 375 178' preserveAspectRatio='none'%3E%3Cpath d='M0 1.5C0 1.5 76 17.5 187.5 17.5C299 17.5 375 1.5 375 1.5V178H0V1.5Z' fill='white'/%3E%3C/svg%3E");
  mask-repeat: no-repeat;
  mask-size: 100% 100%;
  -webkit-mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 375 178' preserveAspectRatio='none'%3E%3Cpath d='M0 1.5C0 1.5 76 17.5 187.5 17.5C299 17.5 375 1.5 375 1.5V178H0V1.5Z' fill='white'/%3E%3C/svg%3E");
  -webkit-mask-repeat: no-repeat;
  -webkit-mask-size: 100% 100%;
}

.source-deposit__chest {
  left: 50%;
  top: 160px;
  z-index: 5;
  width: 216px;
  transform: translateX(-50%);
  animation: source-deposit-chest-pulse 1.65s ease-in-out infinite alternate;
  transform-origin: 50% 62%;
}

.source-deposit__coin {
  left: 118px;
  top: 292px;
  z-index: 6;
  width: 50px;
}

.source-deposit__cash {
  right: 128px;
  top: 286px;
  z-index: 6;
  width: 74px;
}

.source-deposit__arc {
  position: absolute;
  top: 359px;
  left: 0;
  z-index: 7;
  width: 100%;
  height: 19px;
  pointer-events: none;
}

.source-deposit__arc-path {
  animation: source-deposit-arc-flow 1.45s linear infinite;
}

.source-activity--deposit-bonus .source-stage {
  position: absolute;
  inset: 0;
  z-index: 2;
  height: 518px;
  margin: 0;
}

.source-activity--daily-deposit .source-stage {
  position: absolute;
  inset: 0;
  z-index: 2;
  height: 301px;
  margin: 0;
}

.source-daily__bg {
  inset: 0 0 auto;
  z-index: 1;
  width: 100%;
  height: 301px;
  object-fit: cover;
}

.source-daily__main {
  top: 54px;
  right: -3px;
  z-index: 4;
  width: 221px;
  animation: source-daily-main-float 2.4s ease-in-out infinite alternate;
  transform-origin: 50% 58%;
}

.source-activity--deposit-bonus .source-cta {
  position: absolute;
  top: 419px;
  left: 50%;
  z-index: 40;
  width: 174px;
  height: 46px;
  margin: 0;
  color: #8d4a0f;
  background: transparent;
  box-shadow: none;
  font-size: 16px;
  overflow: visible;
  transform: translateX(-50%);
}

.source-activity--deposit-bonus .source-cta::before,
.source-activity--deposit-bonus .source-cta::after {
  position: absolute;
  content: "";
  pointer-events: none;
}

.source-activity--deposit-bonus .source-cta::before {
  inset: 5px 0 0;
  border: 0.5px solid #ffa43a;
  border-radius: 28px;
  background: linear-gradient(180deg, #ffdc34 0%, #fe762d 100%);
  box-shadow:
    inset 0 -1px 4px #ff971d,
    0 8px 14px rgba(67, 12, 4, 0.22);
}

.source-activity--deposit-bonus .source-cta::after {
  inset: 0 0 5px;
  border-bottom: 1px solid rgba(255, 221, 163, 0.4);
  border-radius: 18px;
  background: linear-gradient(180deg, #fff0b0 0%, #ffbc18 100%);
  box-shadow:
    inset 0 0 4px #fff2c6,
    inset 0 -4px 4px rgba(255, 255, 255, 0.25);
}

.source-activity--deposit-bonus .source-cta strong {
  position: relative;
  z-index: 2;
}

.source-activity--deposit-bonus .source-shimmer {
  display: none;
}

.source-activity--deposit-bonus .source-tap-hand {
  right: -49px;
  bottom: -33px;
  z-index: 45;
  width: 70px;
}

.source-illustrated__bg {
  left: 50%;
  top: 10px;
  width: min(390px, 108vw);
  transform: translateX(-50%);
}

.source-illustrated__main {
  left: 50%;
  top: 64px;
  width: min(330px, 88vw);
  transform: translateX(-50%);
}

.source-illustrated__main--large {
  top: 50px;
  width: min(355px, 96vw);
}

.source-illustrated__button {
  left: 50%;
  bottom: 32px;
  width: min(230px, 64vw);
  transform: translateX(-50%);
}

.source-illustrated__left {
  left: 26px;
  bottom: 60px;
  width: 66px;
}

.source-illustrated__right {
  right: 24px;
  bottom: 62px;
  width: 68px;
}

.source-illustrated__rays {
  left: 50%;
  top: 4px;
  width: 340px;
  opacity: 0.55;
  transform: translateX(-50%);
}

.source-periodic__gift {
  z-index: 5;
  right: 34px;
  bottom: 58px;
  width: 78px;
}

.source-wheel__top {
  left: 50%;
  top: -12px;
  width: min(420px, 118vw);
  transform: translateX(-50%);
}

.source-wheel__sun {
  left: 50%;
  top: 48px;
  width: 280px;
  transform: translateX(-50%);
}

.source-wheel__wheel {
  left: 50%;
  top: 74px;
  width: 256px;
  transform: translateX(-50%);
  filter: drop-shadow(0 12px 18px rgba(18, 3, 66, 0.44));
}

.source-wheel__ring {
  left: 50%;
  top: 64px;
  width: 276px;
  transform: translateX(-50%);
}

.source-wheel__pointer {
  left: 50%;
  top: 130px;
  width: 82px;
  transform: translateX(-50%);
}

.source-wheel__prize {
  width: 38px;
}

.source-chest__tunnel {
  z-index: 1;
  inset: 0;
  width: 100%;
  height: 340px;
  object-fit: cover;
  opacity: 0.9;
}

.source-chest__stage {
  z-index: 2;
  left: 50%;
  bottom: 18px;
  width: min(390px, 108vw);
  transform: translateX(-50%);
}

.source-chest__title {
  z-index: 5;
  left: 50%;
  top: 18px;
  width: min(300px, 82vw);
  transform: translateX(-50%);
}

.source-chest__frame-title {
  z-index: 6;
  left: 50%;
  top: 31px;
  width: max-content;
  transform: translateX(-50%);
  color: #fff994;
  font-size: 16px;
  font-style: italic;
  font-weight: 900;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.42);
}

.source-chest__hostess {
  z-index: 4;
  left: 6px;
  bottom: 58px;
  width: 148px;
}

.source-chest__treasure {
  z-index: 4;
  right: 20px;
  bottom: 78px;
  width: 158px;
}

.source-chest__blue {
  z-index: 5;
  left: 50%;
  bottom: 24px;
  width: 154px;
  transform: translateX(-50%);
}

.source-sign__bg {
  inset: 0;
  width: 100%;
  height: 360px;
  object-fit: cover;
}

.source-sign__treasure {
  left: 50%;
  top: 40px;
  width: 168px;
  transform: translateX(-50%);
}

.source-sign__car-left {
  left: 0;
  top: 128px;
  width: 120px;
}

.source-sign__car-right {
  right: 0;
  top: 134px;
  width: 120px;
}

.source-sign__bell {
  right: 60px;
  top: 82px;
  width: 40px;
}

.source-sign__board {
  position: absolute;
  right: 14px;
  bottom: 24px;
  left: 14px;
  z-index: 4;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  padding: 12px;
  border-radius: 20px;
  background: rgba(16, 29, 71, 0.76);
  box-shadow:
    inset 0 0 0 1px rgba(151, 202, 255, 0.28),
    0 12px 24px rgba(0, 0, 0, 0.24);
  backdrop-filter: blur(8px);
}

.source-check-day {
  position: relative;
  display: grid;
  min-height: 55px;
  place-items: center;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.08);

  img {
    width: 30px;
  }

  b {
    color: #ffffff;
    font-size: 10px;
    font-weight: 800;
  }

  &.is-active {
    background: linear-gradient(180deg, #47dfff 0%, #2769ff 100%);
  }
}

.source-red__bg {
  left: 50%;
  top: 0;
  width: min(390px, 108vw);
  transform: translateX(-50%);
}

.source-red__stage {
  left: 50%;
  bottom: 6px;
  z-index: 1;
  width: min(390px, 108vw);
  transform: translateX(-50%);
}

.source-red__subtitle {
  left: 50%;
  top: 44px;
  z-index: 5;
  width: min(292px, 78vw);
  transform: translateX(-50%);
}

.source-red__countdown {
  left: 50%;
  bottom: 26px;
  z-index: 7;
  width: min(205px, 56vw);
  transform: translateX(-50%);
}

.source-red__halo {
  left: 50%;
  top: 94px;
  z-index: 2;
  width: 178px;
  transform: translateX(-50%);
  opacity: 0.62;
}

.source-red__reward {
  left: 26px;
  bottom: 68px;
  z-index: 6;
  width: 86px;
}

.source-red__cell {
  right: 25px;
  bottom: 74px;
  z-index: 6;
  width: 92px;
}

.source-red__packet {
  left: 50%;
  top: 34px;
  z-index: 8;
  width: 230px;
  transform: translateX(-50%);
}

.source-red__packets {
  right: 2px;
  bottom: 64px;
  z-index: 7;
  width: 168px;
}

.source-red__confetti {
  left: 50%;
  top: 10px;
  z-index: 9;
  width: min(365px, 98vw);
  transform: translateX(-50%);
}

.source-rain-layer {
  position: absolute;
  inset: 0;
  z-index: 2;
  overflow: hidden;
  pointer-events: none;
}

.source-rain-packet {
  position: absolute;
  width: 38px;
  pointer-events: none;
}

.source-relief__coin-bg {
  left: 50%;
  bottom: 54px;
  width: min(330px, 88vw);
  transform: translateX(-50%);
}

.source-welfare__main {
  left: 50%;
  top: 64px;
  width: 220px;
  transform: translateX(-50%);
}

.source-welfare__coins {
  left: 30px;
  bottom: 80px;
  width: 132px;
}

.source-welfare__chest {
  right: 34px;
  bottom: 62px;
  width: 138px;
}

.source-sub__ribbon {
  left: 50%;
  bottom: 42px;
  width: min(300px, 80vw);
  transform: translateX(-50%);
}

.source-ranking {
  position: absolute;
  inset: 40px 14px 18px;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  align-items: end;
  gap: 8px;
}

.source-rank-card {
  position: relative;
  display: flex;
  min-height: 170px;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  overflow: hidden;
  border-radius: 26px 26px 14px 14px;
  padding: 12px 7px;
  text-align: center;
  background: linear-gradient(180deg, rgba(65, 110, 222, 0.42), rgba(43, 59, 116, 0.1));
  box-shadow:
    inset 0 0 0 1px rgba(108, 150, 255, 0.45),
    0 15px 24px rgba(0, 0, 0, 0.22);

  img {
    position: absolute;
    top: 12px;
    width: 70px;
  }

  strong {
    margin-top: 78px;
    color: #fff994;
    font-size: 15px;
    font-weight: 900;
  }

  span {
    color: #ffffff;
    font-size: 12px;
    font-weight: 800;
  }
}

.source-rank-card--1 {
  min-height: 224px;

  img {
    width: 86px;
  }
}

.source-simple__rays {
  left: 50%;
  top: 10px;
  width: 340px;
  transform: translateX(-50%);
}

.source-simple__main {
  left: 50%;
  top: 90px;
  width: 150px;
  transform: translateX(-50%);
}

.source-simple__cash {
  left: 50%;
  bottom: 72px;
  width: 160px;
  transform: translateX(-50%);
}

.source-content {
  position: relative;
  margin: -12px 14px 34px;
  padding: 48px 12px 18px;
  border-radius: 26px 26px 18px 18px;
  background:
    radial-gradient(circle at 50% 0, rgba(97, 132, 255, 0.4), transparent 42%),
    linear-gradient(180deg, rgba(54, 38, 145, 0.96) 0%, rgba(19, 23, 60, 0.98) 100%);
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.12),
    inset 0 8px 20px rgba(255, 255, 255, 0.08),
    0 -4px 12px rgba(97, 132, 255, 0.18);
}

.source-activity--first-deposit .source-content::before {
  display: none;
}

.source-activity--red .source-content {
  background:
    radial-gradient(circle at 50% 0, rgba(255, 216, 116, 0.38), transparent 42%),
    linear-gradient(180deg, rgba(180, 38, 38, 0.98) 0%, rgba(75, 19, 33, 0.98) 100%);
}

.source-activity--green .source-content {
  background:
    radial-gradient(circle at 50% 0, rgba(111, 255, 212, 0.33), transparent 42%),
    linear-gradient(180deg, rgba(23, 123, 91, 0.98) 0%, rgba(21, 29, 71, 0.98) 100%);
}

.source-activity--gold .source-content {
  background:
    radial-gradient(circle at 50% 0, rgba(255, 221, 126, 0.4), transparent 42%),
    linear-gradient(180deg, rgba(134, 81, 20, 0.98) 0%, rgba(56, 30, 63, 0.98) 100%);
}

.source-activity--first-deposit .source-content {
  z-index: 4;
  min-height: 720px;
  margin: -14px 12px 106px;
  padding: 38px 26px 78px;
  border-radius: 20px;
  background: var(--source-first-frame) center top / 100% 100% no-repeat;
  box-shadow: none;
}

.source-activity--deposit-bonus .source-content {
  z-index: 5;
  margin: 0 16px 46px !important;
  padding: 0 0 22px;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.source-activity--daily-deposit .source-content {
  z-index: 5;
  margin: 0 14px 0 !important;
  padding: 0 0 24px;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.source-activity--wheel .source-content {
  z-index: 5;
  margin: calc(env(safe-area-inset-top) + 66px) 17px 0 !important;
  padding: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.source-wheel-shell {
  display: flex;
  height: calc(100dvh - env(safe-area-inset-top) - 66px - 19px);
  min-height: 650px;
  flex-direction: column;
  gap: 14px;
}

.source-wheel-shell--intro {
  transform-origin: 50% 13%;
}

.is-wheel-starting .source-wheel-shell--intro {
  animation: source-wheel-intro-launch 0.92s cubic-bezier(0.16, 0.9, 0.2, 1) forwards;
}

.source-wheel-round-card {
  position: relative;
  min-height: 184px;
  overflow: hidden;
  border-radius: 16px;
  padding: 19px 14px 13px;
  background:
    radial-gradient(circle at 50% 0, rgba(255, 255, 255, 0.95) 0 29%, rgba(255, 255, 255, 0.76) 61%, rgba(255, 255, 255, 0.55) 100%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.84) 0%, rgba(255, 255, 255, 0.62) 100%);
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.62),
    0 5px 14px rgba(49, 36, 121, 0.13);
  text-align: center;

  h1 {
    margin: 0;
    color: #0c1031;
    font-size: 19px;
    font-weight: 800;
    line-height: 1.2;
  }

  strong {
    display: block;
    margin-top: 22px;
    color: #191a2d;
    font-size: 28px;
    font-weight: 800;
    letter-spacing: 0;
    line-height: 1.08;
  }

  span {
    font-size: 16px;
    font-weight: 700;
  }
}

.source-wheel-withdraw {
  position: absolute;
  right: 14px;
  bottom: 13px;
  left: 14px;
  height: 50px;
  border: 0;
  border-radius: 10px;
  color: #ffffff;
  background: linear-gradient(180deg, #f157ff 0%, #8c26f9 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 7px 14px rgba(111, 32, 202, 0.18);
  font-size: 16px;
  font-weight: 800;
}

.source-wheel-withdraw:disabled {
  cursor: wait;
  opacity: 0.92;
}

.source-wheel-withdraw__spinner {
  display: inline-block;
  width: 15px;
  height: 15px;
  margin-right: 8px;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top-color: #ffffff;
  border-radius: 50%;
  vertical-align: -2px;
  animation: source-wheel-spin-loader 0.62s linear infinite;
}

.source-wheel-state-enter-active,
.source-wheel-state-leave-active {
  transition:
    opacity 0.34s ease,
    transform 0.42s cubic-bezier(0.2, 0.9, 0.25, 1.08);
}

.source-wheel-state-enter-from {
  opacity: 0;
  transform: translateY(18px) scale(0.96);
}

.source-wheel-state-leave-to {
  opacity: 0;
  transform: translateY(-18px) scale(1.04);
}

.source-wheel-started {
  position: relative;
  min-height: 1080px;
  margin: 0 -3px;
  padding-bottom: 28px;
}

.source-wheel-status-card {
  position: relative;
  z-index: 4;
  min-height: 148px;
  overflow: hidden;
  border-radius: 18px;
  padding: 12px 16px 11px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.99) 0%, rgba(255, 255, 255, 0.96) 36%, rgba(247, 251, 255, 0.94) 58%, rgba(242, 255, 248, 0.9) 100%);
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.78),
    0 5px 14px rgba(49, 36, 121, 0.14);
  text-align: center;
}

.source-wheel-status-card__head {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 8px;

  h1 {
    margin: 0;
    color: #121327;
    font-family: "Poetsen One", Montserrat, sans-serif;
    font-size: 15px;
    font-style: italic;
    font-weight: 900;
    line-height: 1;
    white-space: nowrap;
  }
}

.source-wheel-countdown {
  display: flex;
  min-width: 0;
  justify-content: center;
  align-items: center;
  gap: 3px;
  color: #f19610;
  font-size: 15px;
  font-weight: 800;

  span {
    display: grid;
    min-width: 25px;
    height: 25px;
    place-items: center;
    border: 1px solid rgba(240, 159, 25, 0.68);
    border-radius: 4px;
    background: rgba(255, 251, 238, 0.92);
    box-shadow: inset 0 0 0 1px rgba(255, 221, 133, 0.5);
  }

  i {
    color: #f19610;
    font-style: normal;
  }
}

.source-wheel-details {
  border: 0;
  padding: 0;
  color: #8d18ff;
  background: transparent;
  font-size: 17px;
  font-weight: 600;
  text-decoration: underline;
}

.source-wheel-status-card > strong {
  display: block;
  margin-top: 15px;
  color: #8b18f4;
  font-size: 43px;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1;

  span {
    margin-right: 6px;
    font-size: 24px;
    font-weight: 700;
  }
}

.source-wheel-progress {
  position: relative;
  height: 24px;
  margin: 9px 15px 0;

  &::before {
    position: absolute;
    top: 5px;
    right: 0;
    left: 0;
    height: 7px;
    border-radius: 999px;
    background: linear-gradient(90deg, #d9b5ed 0%, #d8a9ee 50%, #d3a1e8 100%);
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.36);
    content: "";
  }

  i {
    position: absolute;
    top: 5px;
    left: 0;
    max-width: 100%;
    height: 7px;
    border-radius: 999px;
    background: linear-gradient(90deg, #bf25f3 0%, #ca4af5 52%, #d781f5 100%);
    box-shadow: 0 0 8px rgba(196, 45, 242, 0.32);
    transition: width 0.35s ease;
  }
}

.source-wheel-progress__icon {
  position: absolute;
  z-index: 2;
  top: -11px;
  display: grid;
  width: 32px;
  height: 32px;
  place-items: center;
  filter: drop-shadow(0 2px 5px rgba(72, 23, 126, 0.22));

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
}

.source-wheel-progress__icon--start {
  left: 0;
  transform: translateX(-50%);
}

.source-wheel-progress__icon--mid {
  left: 50%;
  transform: translateX(-50%);
}

.source-wheel-progress__icon--end {
  right: 0;
  transform: translateX(50%);
}

.source-wheel-status-card p {
  margin: 7px 0 0;
  color: rgba(31, 35, 54, 0.82);
  font-size: 15px;
  font-weight: 500;
  line-height: 1.2;

  b {
    color: #6f23bc;
    font-weight: 900;
  }
}

.source-wheel-game {
  position: relative;
  z-index: 6;
  --wheel-size: min(348px, 90vw);
  height: calc(var(--wheel-size) + 151px);
  margin: -36px -20px 0;
  pointer-events: none;
}

.source-wheel-game::after {
  position: absolute;
  right: -16px;
  bottom: 0;
  left: -16px;
  z-index: 1;
  height: 92px;
  border-radius: 50%;
  background: radial-gradient(ellipse at center, rgba(255, 226, 255, 0.2), rgba(112, 23, 193, 0.04) 58%, transparent 72%);
  content: "";
  pointer-events: none;
}

.source-wheel-game__rotor,
.source-wheel-game__wheel,
.source-wheel-game__star,
.source-wheel-game__label,
.source-wheel-game__prize,
.source-wheel-game__pedestal {
  position: absolute;
  display: block;
  object-fit: contain;
  user-select: none;
  pointer-events: none;
}

.source-wheel-game__rotor {
  left: 50%;
  top: 0;
  z-index: 4;
  width: var(--wheel-size);
  aspect-ratio: 1;
  border-radius: 50%;
  filter: drop-shadow(0 10px 11px rgba(93, 30, 165, 0.28));
  transform-origin: 50% 50%;
  will-change: transform;
}

.source-wheel-game__rotor.is-spinning {
  filter: drop-shadow(0 12px 13px rgba(93, 30, 165, 0.34)) saturate(1.08);
}

.source-wheel-game__wheel {
  inset: 0;
  width: 100%;
  height: 100%;
}

.source-wheel-game__center {
  position: absolute;
  left: 50%;
  top: calc(var(--wheel-size) * 0.46);
  z-index: 9;
  display: grid;
  width: 76px;
  height: 76px;
  place-items: center;
  border: 0;
  padding: 0;
  background: transparent;
  filter: drop-shadow(0 3px 6px rgba(52, 16, 102, 0.28));
  transform: translate(-50%, -50%);
  cursor: pointer;
  pointer-events: auto;
  appearance: none;

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
    pointer-events: none;
  }
}

.source-wheel-game__center:disabled {
  cursor: wait;
  filter: brightness(1.08);
}

.source-wheel-game__label {
  left: 50%;
  top: 50%;
  z-index: 7;
  color: #a739f5;
  font-size: 14px;
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 1px 0 rgba(255, 245, 183, 0.5);
  white-space: nowrap;
}

.source-wheel-game__prize {
  left: 50%;
  top: 50%;
  z-index: 5;
  width: 42px;
  height: 42px;
  transform-origin: 50% 50%;
}

.source-wheel-game__star {
  z-index: 1;
  width: 46px;
  opacity: 0.72;
  mix-blend-mode: screen;
  animation: source-wheel-star-pulse 1.6s ease-in-out infinite alternate;
}

.source-wheel-game__star--left {
  left: 38px;
  top: 90px;
}

.source-wheel-game__star--right {
  right: 35px;
  top: 53px;
  animation-delay: -0.65s;
}

.source-wheel-game__pedestal {
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 2;
  width: 100%;
  height: auto;
}

.source-wheel-floating {
  position: absolute;
  left: 50%;
  top: 158px;
  z-index: 10;
  color: #fff36c;
  font-size: 31px;
  font-weight: 900;
  line-height: 1;
  text-shadow:
    0 3px 0 rgba(119, 32, 178, 0.75),
    0 0 10px rgba(255, 255, 255, 0.9);
  transform: translateX(-50%);
  pointer-events: none;
}

.source-wheel-floating-enter-active,
.source-wheel-floating-leave-active {
  transition:
    opacity 0.36s ease,
    transform 0.5s cubic-bezier(0.22, 0.86, 0.26, 1.18);
}

.source-wheel-floating-enter-from {
  opacity: 0;
  transform: translate(-50%, 12px) scale(0.78);
}

.source-wheel-floating-leave-to {
  opacity: 0;
  transform: translate(-50%, -35px) scale(1.1);
}

.source-wheel-action-row {
  position: relative;
  z-index: 8;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
  margin: -160px -20px 0;
  padding: 12px 22px 3px;
}

.source-wheel-action-row::before {
  position: absolute;
  top: -6px;
  right: 0;
  bottom: -92px;
  left: 0;
  z-index: -1;
  background: transparent;
  content: "";
}

.source-wheel-action {
  position: relative;
  height: 68px;
  border: 0;
  padding: 0;
  color: #ffffff;
  background: transparent;
  box-shadow: none;
  cursor: pointer;

  span {
    pointer-events: none;
  }

  &::before,
  &::after {
    position: absolute;
    content: "";
  }

  &::before {
    left: 0;
    right: 0;
    top: 17px;
    bottom: 0;
    z-index: 1;
    border-radius: 13px;
    transform: skewX(-4deg);
    transform-origin: 50% 100%;
    box-shadow: 0 8px 0 rgba(74, 20, 137, 0.72);
  }

  &::after {
    left: 6px;
    right: 6px;
    top: 0;
    z-index: 2;
    height: 56px;
    border-radius: 13px;
    transform: skewX(-4deg);
    transform-origin: 50% 100%;
    box-shadow:
      inset 0 5px 0 rgba(255, 255, 255, 0.26),
      inset 0 -5px 0 rgba(0, 0, 0, 0.13),
      0 2px 0 rgba(255, 255, 255, 0.2);
  }

  span {
    position: absolute;
    top: 9px;
    left: 15%;
    right: 15%;
    z-index: 3;
    display: grid;
    min-height: 34px;
    place-items: center;
    color: #ffffff;
    font-size: clamp(13px, 3.7vw, 15px);
    font-weight: 900;
    line-height: 1.05;
    text-align: center;
    word-break: normal;
    text-wrap: balance;
  }
}

.source-wheel-action:disabled {
  cursor: wait;
  filter: brightness(1.06) saturate(0.95);
}

.source-wheel-action--share {
  &::before {
    background: linear-gradient(180deg, #9b18ce 0%, #6d159e 100%);
  }

  &::after {
    border: 1px solid rgba(255, 180, 255, 0.42);
    background: linear-gradient(180deg, #ff72f5 0%, #ef3be8 48%, #bd18d2 100%);
  }

  span {
    text-shadow:
      -1.5px -1.5px 0 #9c1fa6,
      1.5px -1.5px 0 #9c1fa6,
      -1.5px 1.5px 0 #9c1fa6,
      1.5px 1.5px 0 #9c1fa6,
      0 1.5px 0 #9c1fa6;
  }
}

.source-wheel-action--free {
  &::before {
    background: linear-gradient(180deg, #c66c00 0%, #8d4200 100%);
  }

  &::after {
    border: 1px solid rgba(255, 246, 143, 0.55);
    background: linear-gradient(180deg, #ffea47 0%, #ffd01a 48%, #f19700 100%);
  }

  span {
    text-shadow:
      -1.5px -1.5px 0 #c65700,
      1.5px -1.5px 0 #c65700,
      -1.5px 1.5px 0 #c65700,
      1.5px 1.5px 0 #c65700,
      0 1.5px 0 #c65700;
  }
}

.source-wheel-spin-meta {
  position: relative;
  z-index: 8;
  display: flex;
  height: 34px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #fff477;
  font-size: 16px;
  font-weight: 600;

  b {
    display: grid;
    width: 27px;
    height: 25px;
    place-items: center;
    border-radius: 5px;
    color: #ffffff;
    background: rgba(102, 42, 187, 0.82);
    font-size: 15px;
    font-weight: 900;
  }

  button {
    border: 0;
    padding: 0;
    color: rgba(255, 255, 255, 0.95);
    background: transparent;
    font-size: 16px;
    font-weight: 600;
    text-decoration: underline;
  }

  button[aria-label="Refresh"] {
    width: 20px;
    color: #ffffff;
    font-size: 21px;
    line-height: 1;
    text-decoration: none;
  }
}

.source-wheel-tip {
  position: relative;
  z-index: 7;
  height: 31px;
  margin: 28px -20px 0;
  display: grid;
  place-items: center;
  color: #ffffff;
  background: rgba(124, 38, 205, 0.96);
  font-size: 13px;
  font-weight: 700;
}

.source-wheel-record-card {
  min-height: 0;
  flex: 1;
  overflow: hidden;
  border-radius: 8px;
  padding: 0 10px 10px;
  background:
    radial-gradient(circle at 50% 0, rgba(166, 113, 255, 0.64), transparent 43%),
    linear-gradient(180deg, #7f32de 0%, #5d1db9 45%, #4d169c 100%);
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.18),
    0 5px 14px rgba(62, 29, 122, 0.16);
}

.source-wheel-record-card--started {
  height: 560px;
  margin: 29px -3px 0;
  padding: 0 14px 14px;
  border-radius: 14px;
}

.source-wheel-record-tabs {
  display: grid;
  height: 58px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.22);

  button {
    position: relative;
    height: 100%;
    border: 0;
    padding: 0;
    color: rgba(255, 255, 255, 0.58);
    background: transparent;
    font-size: 16px;
    font-weight: 500;
  }

  .is-active {
    color: #fff8da;
  }

  .is-active::after {
    position: absolute;
    right: 27px;
    bottom: 7px;
    left: 27px;
    height: 4px;
    border-radius: 999px;
    background: #fff4bd;
    box-shadow: 0 0 8px rgba(255, 246, 191, 0.62);
    content: "";
  }
}

.source-wheel-record-list {
  height: calc(100% - 56px);
  overflow: hidden;
  padding: 4px 0 4px;
  mask-image: linear-gradient(180deg, transparent 0, #000 10px, #000 calc(100% - 10px), transparent 100%);
}

.source-wheel-record-track {
  animation: source-wheel-notice-scroll 9s linear infinite;
  will-change: transform;
}

.source-wheel-record-list:hover .source-wheel-record-track {
  animation-play-state: paused;
}

.source-wheel-record-empty {
  display: grid;
  height: calc(100% - 56px);
  place-items: center;
  color: rgba(255, 255, 255, 0.68);
  font-size: 13px;
  font-weight: 700;
  text-align: center;

  p {
    margin: 0;
  }
}

.source-wheel-record-row {
  display: grid;
  min-height: 62px;
  grid-template-columns: 37px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  margin-bottom: 11px;
  border-radius: 11px;
  padding: 8px 10px 8px 11px;
  background: rgba(117, 69, 218, 0.82);
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, 0.06),
    0 2px 5px rgba(42, 16, 103, 0.12);

}

.source-wheel-record-avatar {
  position: relative;
  display: block;
  width: 31px;
  height: 31px;
  overflow: hidden;
  border-radius: 50%;
  background: linear-gradient(145deg, #88e7ff 0%, #5771ff 47%, #2f1a9b 100%);
  box-shadow:
    inset 0 0 0 2px rgba(255, 255, 255, 0.2),
    0 2px 4px rgba(24, 8, 68, 0.24);
}

.source-wheel-record-avatar::before {
  position: absolute;
  top: 7px;
  left: 50%;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  content: "";
  transform: translateX(-50%);
}

.source-wheel-record-avatar::after {
  position: absolute;
  right: 6px;
  bottom: 5px;
  left: 6px;
  height: 8px;
  border-radius: 9px 9px 4px 4px;
  background: rgba(255, 255, 255, 0.9);
  content: "";
}

.source-wheel-record-avatar[data-tone="gold"] {
  background: linear-gradient(145deg, #ffe66a 0%, #f9942f 50%, #a64b18 100%);
}

.source-wheel-record-avatar[data-tone="green"],
.source-wheel-record-avatar[data-tone="mint"] {
  background: linear-gradient(145deg, #c0ff84 0%, #29bb84 50%, #0b6d74 100%);
}

.source-wheel-record-avatar[data-tone="pink"] {
  background: linear-gradient(145deg, #ffb6f5 0%, #f04ca5 50%, #801e9f 100%);
}

.source-wheel-record-avatar[data-tone="cyan"] {
  background: linear-gradient(145deg, #9ff8ff 0%, #20a9ff 50%, #1950b8 100%);
}

.source-wheel-record-avatar[data-tone="violet"] {
  background: linear-gradient(145deg, #d9a8ff 0%, #8b59ff 50%, #381596 100%);
}

.source-wheel-record-avatar[data-tone="orange"] {
  background: linear-gradient(145deg, #ffd480 0%, #ff7658 55%, #ad2b55 100%);
}

.source-wheel-record-meta {
  min-width: 0;

  strong {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: #ffffff;
    font-size: 15px;
    font-weight: 500;
    line-height: 1.2;
  }
}

.source-wheel-record-status {
  display: grid;
  justify-items: end;
  gap: 7px;

  time {
    display: block;
    color: rgba(255, 255, 255, 0.94);
    font-size: 13px;
    font-weight: 600;
    line-height: 1.1;
    white-space: nowrap;
  }

  em {
    display: inline-flex;
    min-width: 148px;
    height: 24px;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    padding: 0 11px;
    color: #98600d;
    background: linear-gradient(180deg, #fff38b 0%, #ffdf63 100%);
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.5),
      0 2px 5px rgba(82, 43, 3, 0.12);
    font-size: 12px;
    font-style: normal;
    font-weight: 600;
    white-space: nowrap;
  }
}

.source-wheel-history-panel {
  height: calc(100% - 58px);
  padding: 17px 7px;
}

.source-wheel-history-card {
  min-height: 119px;
  border-radius: 12px;
  padding: 16px 15px;
  color: #ffffff;
  background: rgba(69, 18, 141, 0.58);

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  h2,
  strong,
  dl,
  dt,
  dd {
    margin: 0;
  }

  h2 {
    font-family: "Poetsen One", Montserrat, sans-serif;
    font-size: 15px;
    font-style: italic;
    font-weight: 900;
  }

  strong {
    font-size: 17px;
    font-weight: 800;
  }

  dl {
    display: grid;
    gap: 10px;
    margin-top: 18px;
  }

  dl > div {
    display: grid;
    grid-template-columns: 1fr auto;
    align-items: center;
    gap: 12px;
  }

  dt,
  dd {
    font-size: 14px;
    font-weight: 500;
  }

  dd {
    text-align: right;
  }

  .is-pending {
    color: #2da7ff;
    font-weight: 700;
  }
}

.source-daily-panel {
  position: relative;
  min-height: 323px;
  overflow: hidden;
  border: 1px solid rgba(255, 230, 130, 0.92);
  border-radius: 10px;
  padding: 47px 14px 18px;
  background:
    linear-gradient(180deg, rgba(102, 8, 4, 0.96) 0%, rgba(105, 9, 5, 0.94) 52%, rgba(52, 7, 5, 0.96) 100%);
  box-shadow:
    inset 0 0 18px rgba(255, 230, 146, 0.16),
    0 6px 12px rgba(31, 2, 0, 0.28);
}

.source-daily-panel::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 83px;
  background:
    radial-gradient(circle at 50% 100%, rgba(255, 214, 97, 0.13), transparent 58%),
    repeating-linear-gradient(-25deg, rgba(255, 234, 169, 0.04) 0 1px, transparent 1px 8px);
  opacity: 0.9;
  content: "";
  pointer-events: none;
}

.source-daily-panel__title {
  position: absolute;
  top: 0;
  left: 50%;
  z-index: 3;
  display: grid;
  width: 258px;
  height: 31px;
  place-items: center;
  color: #281900;
  background: linear-gradient(90deg, #fff6bc 0%, #d4ab52 52%, #fff6bc 100%);
  clip-path: polygon(0 0, 100% 0, 92% 100%, 8% 100%);
  font-size: 14px;
  font-weight: 800;
  line-height: 1;
  text-align: center;
  transform: translateX(-50%);
}

.source-daily-reward-row {
  position: relative;
  z-index: 2;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 13px;
}

.source-daily-reward {
  position: relative;
  display: flex;
  height: 55px;
  align-items: center;
  justify-content: center;
  gap: 8px;
  overflow: hidden;
  border-radius: 5px;
  background: rgba(38, 0, 0, 0.54);

  strong {
    position: relative;
    z-index: 2;
    color: #fff0a2;
    font-size: 18px;
    font-weight: 800;
    line-height: 1;
  }
}

.source-daily-reward__burst {
  position: absolute;
  left: 20px;
  width: 66px;
  height: 66px;
  opacity: 0.86;
  animation: source-deposit-spin 10s linear infinite;
}

.source-daily-reward__icon {
  position: relative;
  z-index: 1;
  width: 37px;
  height: 35px;
  object-fit: contain;
}

.source-daily-reward__icon--cash {
  width: 42px;
  height: 42px;
}

.source-daily-claim {
  position: relative;
  z-index: 2;
  min-height: 57px;
  margin-top: 14px;
  border: 1px solid rgba(255, 211, 147, 0.14);
  border-radius: 6px;
  padding: 9px 10px;
  color: #ffffff;
  background: rgba(66, 0, 0, 0.18);

  span,
  strong {
    display: block;
    line-height: 1.3;
  }

  span {
    font-size: 12px;
    font-weight: 700;
  }

  strong {
    margin-top: 2px;
    font-size: 12px;
    font-weight: 500;
  }
}

.source-daily-progress-copy {
  position: relative;
  z-index: 2;
  margin: 14px 0 7px;
  color: #ffffff;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.45;

  strong {
    color: #ffe59a;
    font-size: 14px;
    font-weight: 900;
  }
}

.source-daily-meter {
  position: relative;
  z-index: 2;
  height: 12px;
  overflow: hidden;
  border: 1px solid rgba(255, 202, 168, 0.45);
  border-radius: 999px;
  background: rgba(68, 0, 0, 0.5);
  box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.38);

  i {
    display: block;
    width: 0%;
    height: 100%;
    border-radius: inherit;
    background: linear-gradient(90deg, #ffe978 0%, #ffbc3b 100%);
  }
}

.source-daily-increase {
  position: relative;
  z-index: 3;
  display: grid;
  width: 255px;
  height: 42px;
  place-items: center;
  margin: 17px auto 0;
  border: 0;
  padding: 0;
  color: #fff8c4;
  background: transparent;
  font-size: 15px;
  font-weight: 800;
  text-shadow: 0 2px 1px rgba(96, 15, 5, 0.9);

  img,
  span {
    grid-area: 1 / 1;
  }

  img {
    width: 255px;
    height: 42px;
    object-fit: contain;
  }

  span {
    position: relative;
    z-index: 2;
    margin-top: 1px;
  }
}

.source-daily-corner {
  position: absolute;
  bottom: 4px;
  z-index: 1;
  width: 86px;
  height: 81px;
  object-fit: contain;
  pointer-events: none;
}

.source-daily-corner--left {
  left: 0;
}

.source-daily-corner--right {
  right: 0;
}

.source-daily-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  margin: 10px 0 30px;

  button {
    min-height: 36px;
    border: 1px solid transparent;
    border-radius: 8px;
    color: rgba(255, 255, 255, 0.68);
    background: rgba(37, 11, 6, 0.34);
    font-size: 13px;
    font-weight: 600;
    line-height: 1;
  }

  button.is-active {
    border-color: #ffdf9f;
    color: #ffffff;
    background:
      linear-gradient(180deg, rgba(141, 55, 26, 0.98) 0%, rgba(100, 19, 11, 0.98) 100%);
    box-shadow:
      inset 0 0 0 1px rgba(255, 255, 255, 0.18),
      0 4px 8px rgba(22, 0, 0, 0.2);
    font-weight: 800;
  }
}

.source-daily-table-title {
  display: grid;
  min-height: 36px;
  place-items: center;
  margin-top: -10px;
  color: #ffffff;
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1;
  text-align: center;

  span {
    position: relative;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    text-shadow: 0 2px 2px rgba(0, 0, 0, 0.32);

    &::before,
    &::after {
      display: block;
      width: 32px;
      height: 1px;
      background: linear-gradient(90deg, transparent 0%, #ffd6a0 100%);
      content: "";
    }

    &::after {
      transform: scaleX(-1);
    }
  }
}

.source-daily-table {
  overflow: hidden;
  border: 1px solid rgba(255, 223, 159, 0.52);
  border-radius: 6px 6px 0 0;
  background: rgba(75, 0, 0, 0.55);
}

.source-daily-table__head,
.source-daily-table__row {
  display: grid;
  grid-template-columns: 0.64fr 1.45fr repeat(3, 0.82fr);
}

.source-daily-table__head span,
.source-daily-table__row span {
  display: grid;
  min-height: 31px;
  place-items: center;
  border-right: 1px solid rgba(255, 223, 159, 0.35);
  border-bottom: 1px solid rgba(255, 223, 159, 0.25);
  padding: 6px 4px;
  color: #ffffff;
  font-size: 11px;
  line-height: 1.15;
  text-align: center;
}

.source-daily-table__head span {
  background: rgba(64, 7, 4, 0.78);
  font-weight: 800;
}

.source-daily-table__row span {
  background: rgba(77, 4, 3, 0.55);
  font-weight: 500;
}

.source-daily-table__head span:last-child,
.source-daily-table__row span:last-child {
  border-right: 0;
}

.source-daily-table__row:last-child span {
  border-bottom: 0;
}

.source-daily-empty {
  display: flex;
  min-height: 264px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  text-align: center;

  svg {
    width: 104px;
    height: 104px;
    opacity: 0.96;
  }

  p {
    margin: 8px 0 0;
    color: #d2d6ff;
    font-size: 16px;
    font-weight: 600;
  }
}

.source-deposit-panel {
  position: relative;
  border: 1px solid #ffb0b1;
  border-radius: 10px;
  background: linear-gradient(180deg, #c32d29 0%, #a51f1d 100%);
  box-shadow:
    inset 0 0 5px 5px rgba(255, 255, 255, 0.12),
    0 10px 18px rgba(41, 3, 3, 0.22);
}

.source-deposit-panel--max {
  min-height: 119px;
  margin-top: 14px;
  padding: 39px 12px 13px;
}

.source-deposit-panel--claim {
  min-height: 286px;
  margin-top: 46px;
  padding: 28px 12px 24px;
}

.source-deposit-ribbon {
  position: absolute;
  top: 0;
  left: 50%;
  z-index: 4;
  display: grid;
  height: 36px;
  place-items: center;
  transform: translate(-50%, -50%);
  filter: drop-shadow(0 3px 0 rgba(108, 18, 18, 0.16));

  &::before,
  &::after {
    position: absolute;
    content: "";
    pointer-events: none;
  }

  &::before {
    inset: 0;
    border: 1px solid rgba(255, 255, 255, 0.94);
    clip-path: polygon(5% 0, 95% 0, 100% 50%, 95% 100%, 5% 100%, 0 50%);
  }

  &::after {
    inset: 5px 10px;
    border: 1px solid rgba(255, 255, 255, 0.92);
    clip-path: polygon(6% 0, 94% 0, 100% 50%, 94% 100%, 6% 100%, 0 50%);
  }

  span {
    position: relative;
    z-index: 2;
    font-size: 14px;
    font-style: italic;
    font-weight: 800;
    line-height: 1;
    white-space: nowrap;
  }
}

.source-deposit-ribbon--light {
  width: min(299px, 86vw);
  color: #e8210e;
  background:
    linear-gradient(108deg, transparent 0 15%, rgba(255, 255, 255, 0.24) 16% 29%, transparent 30%),
    linear-gradient(90deg, #ffd6bd 0%, #ffdc7b 100%);
  clip-path: polygon(5% 0, 95% 0, 100% 50%, 95% 100%, 5% 100%, 0 50%);
  text-shadow: 0 1px 0 #ffffff;
}

.source-deposit-ribbon--red {
  width: min(219px, 64vw);
  color: #ffffff;
  background:
    linear-gradient(108deg, transparent 0 20%, rgba(255, 255, 255, 0.2) 21% 36%, transparent 37%),
    linear-gradient(90deg, #b11f13 0%, #db292b 100%);
  clip-path: polygon(7% 0, 93% 0, 100% 50%, 93% 100%, 7% 100%, 0 50%);
  text-shadow: 0 1px 0 #000000;
}

.source-deposit-max-row {
  display: flex;
  height: 60px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  overflow: hidden;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.38);

  span,
  strong {
    display: block;
    color: #ffffff;
  }

  span {
    font-size: 12px;
    font-weight: 500;
    line-height: 1.2;
  }

  strong {
    margin-top: 2px;
    font-size: 15px;
    font-weight: 700;
    line-height: 1.1;
  }
}

.source-deposit-reward-icon {
  position: relative;
  width: 48px;
  height: 48px;
  flex: 0 0 auto;
}

.source-deposit-reward-icon__rays,
.source-deposit-reward-icon__coins {
  position: absolute;
  object-fit: contain;
}

.source-deposit-reward-icon__rays {
  inset: -5px;
  width: 58px;
  height: 58px;
  animation: source-deposit-spin 8s linear infinite;
}

.source-deposit-reward-icon__coins {
  inset: 3px;
  width: 42px;
  height: 42px;
}

.source-deposit-claimed {
  display: grid;
  gap: 7px;
  margin-top: 4px;

  p {
    display: flex;
    align-items: center;
    gap: 5px;
    margin: 0;
    color: #ffffff;
    font-size: 12px;
    font-weight: 500;
    line-height: 1.2;
  }

  img {
    width: 21px;
    height: 21px;
    object-fit: contain;
  }
}

.source-deposit-empty {
  display: flex;
  min-height: 178px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #ffffff;

  svg {
    width: 96px;
    height: 96px;
    opacity: 0.96;
  }

  p {
    margin: 7px 0 0;
    color: #97a5ca;
    font-size: 16px;
    font-weight: 600;
    line-height: 1.25;
    text-align: center;
  }
}

.source-content__banner {
  position: absolute;
  top: -18px;
  left: 50%;
  display: grid;
  width: min(300px, 84vw);
  min-height: 38px;
  place-items: center;
  transform: translateX(-50%);

  img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  strong {
    position: relative;
    z-index: 2;
    color: #ffffff;
    font-size: 15px;
    font-style: italic;
    font-weight: 900;
    text-shadow: 0 2px 3px rgba(42, 5, 5, 0.42);
  }
}

.source-activity--first-deposit .source-content__banner {
  top: -82px;
  width: min(332px, 92vw);
  min-height: 104px;

  strong {
    margin-top: 11px;
    color: #ffffff;
    font-size: 22px;
    font-style: italic;
    font-weight: 900;
    line-height: 1;
    text-shadow:
      0 2px 0 rgba(110, 48, 222, 0.85),
      0 4px 8px rgba(24, 8, 80, 0.34);
    transform: rotate(-4deg);
  }
}

.source-content__arc-title {
  position: relative;
  z-index: 2;
  width: 100%;
  height: 92px;
  overflow: visible;
  pointer-events: none;
  transform: translateY(-8px);

  text {
    fill: #ffffff;
    font-family: Montserrat, sans-serif;
    font-size: 20px;
    font-style: italic;
    font-weight: 900;
    filter:
      drop-shadow(0 2px 0 rgba(110, 48, 222, 0.85))
      drop-shadow(0 4px 8px rgba(24, 8, 80, 0.34));
  }
}

.source-activity--first-deposit .source-card-grid {
  position: relative;
  z-index: 1;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.source-activity--first-deposit .source-card {
  display: flex;
  height: 118px;
  min-height: 118px;
  flex-direction: column;
  justify-content: space-between;
  overflow: visible;
  border: 1px solid rgba(127, 82, 248, 0.28);
  border-radius: 11px;
  padding: 8px;
  background: #382568;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.035);
  box-sizing: border-box;
}

.source-activity--first-deposit .source-card::after {
  display: none;
}

.source-first-reward-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;

  > div {
    position: relative;
    display: flex;
    min-height: 64px;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 1px solid #8652ff;
    border-radius: 7px;
    background: rgba(255, 255, 255, 0.08);
  }

  strong {
    position: relative;
    z-index: 2;
    margin: 2px 0 0;
    color: #fbd89d;
    font-size: 10px;
    font-weight: 800;
    line-height: 1.05;
    text-align: center;
    text-shadow:
      0 1px 0 #c41214,
      0 0 4px rgba(255, 255, 255, 0.2);
  }

  > div:nth-child(2) strong {
    color: #dcefe0;
    text-shadow:
      0 1px 0 #622ccd,
      0 0 4px rgba(255, 255, 255, 0.18);
  }
}

.source-first-reward-pair__rays {
  position: absolute;
  top: 6px;
  left: 50%;
  width: 38px;
  height: 38px;
  opacity: 0.34;
  transform: translateX(-50%);
}

.source-first-reward-pair__icon {
  position: relative;
  z-index: 1;
  width: 26px;
  height: 26px;
  object-fit: contain;
}

.source-first-amount {
  display: block;
  min-height: 30px;
  margin-top: 6px;
  color: #ffffff;
  font-size: 10px;
  font-weight: 700;
  line-height: 1.25;
  text-align: center;
}

.source-panel {
  position: relative;
  margin-bottom: 10px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 18px;
  padding: 12px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.16), rgba(255, 255, 255, 0.04)),
    rgba(0, 0, 0, 0.26);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    0 8px 18px rgba(0, 0, 0, 0.18);

  span,
  strong,
  em,
  p {
    position: relative;
    z-index: 1;
  }
}

.source-progress-panel {
  div {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 10px;
  }

  span {
    color: rgba(255, 255, 255, 0.72);
    font-size: 11px;
    font-weight: 700;
  }

  strong {
    max-width: 170px;
    color: #fff994;
    font-size: 14px;
    font-weight: 900;
    line-height: 1.2;
    text-align: right;
  }

  p {
    margin: 8px 0;
    color: rgba(255, 255, 255, 0.76);
    font-size: 10px;
    font-weight: 600;
    line-height: 1.4;
  }

  em {
    display: block;
    margin-top: 7px;
    color: #ffffff;
    font-size: 11px;
    font-style: normal;
    font-weight: 800;
  }
}

.source-meter {
  position: relative;
  height: 9px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(2, 8, 25, 0.42);
  box-shadow:
    inset 0 1px 3px rgba(0, 0, 0, 0.34),
    0 1px 0 rgba(255, 255, 255, 0.15);

  i {
    position: absolute;
    inset: 0 auto 0 0;
    border-radius: inherit;
    background: linear-gradient(90deg, #22e2ff 0%, #23ffba 54%, #fff47c 100%);
    box-shadow: 0 0 12px rgba(55, 255, 206, 0.55);
  }
}

.source-cycle-panel {
  display: grid;
  gap: 10px;

  > em {
    color: rgba(255, 255, 255, 0.76);
    font-size: 11px;
    font-style: normal;
    font-weight: 700;
  }
}

.source-cycle-panel__head {
  display: grid;
  gap: 4px;

  span {
    color: rgba(255, 255, 255, 0.66);
    font-size: 11px;
    font-weight: 700;
  }

  strong {
    color: #ffffff;
    font-size: 12px;
    font-weight: 800;
    line-height: 1.35;
  }
}

.source-cycle-panel__stats {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;

  span {
    min-height: 48px;
    border-radius: 13px;
    padding: 8px;
    color: rgba(255, 255, 255, 0.72);
    background: rgba(255, 255, 255, 0.08);
    font-size: 10px;
    font-weight: 700;
    line-height: 1.25;
  }

  b {
    display: block;
    margin-top: 5px;
    color: #fff994;
    font-size: 13px;
    font-weight: 900;
  }
}

.source-countdown {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 7px;

  span {
    display: grid;
    min-height: 48px;
    place-items: center;
    border-radius: 13px;
    background: rgba(0, 0, 0, 0.28);
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);
  }

  strong {
    color: #fff994;
    font-size: 18px;
    font-weight: 900;
    line-height: 1;
  }

  em {
    color: rgba(255, 255, 255, 0.72);
    font-size: 9px;
    font-style: normal;
    font-weight: 700;
  }
}

.source-chest-panel {
  display: grid;
  gap: 7px;

  strong {
    color: #fff994;
    font-size: 15px;
    font-weight: 900;
  }

  span,
  p {
    margin: 0;
    color: rgba(255, 255, 255, 0.78);
    font-size: 12px;
    font-weight: 700;
  }
}

.source-chest-steps {
  display: grid;
  grid-template-columns: repeat(10, minmax(0, 1fr));
  gap: 5px;
  margin-top: 5px;

  i {
    display: grid;
    height: 22px;
    place-items: center;
    border-radius: 999px;
    color: rgba(255, 255, 255, 0.62);
    background: rgba(255, 255, 255, 0.08);
    font-size: 9px;
    font-style: normal;
    font-weight: 800;
  }

  .is-active {
    color: #2b1d00;
    background: linear-gradient(180deg, #fff4a1 0%, #ffb727 100%);
  }
}

.source-red-panel {
  display: grid;
  grid-template-columns: 82px 1fr;
  align-items: center;
  gap: 10px;
  background:
    radial-gradient(circle at 10% 0%, rgba(255, 244, 144, 0.45), transparent 48%),
    linear-gradient(180deg, rgba(255, 81, 63, 0.8), rgba(115, 12, 23, 0.88));

  img {
    width: 82px;
    object-fit: contain;
  }

  strong {
    display: block;
    color: #fff994;
    font-size: 14px;
    font-weight: 900;
    line-height: 1.25;
  }

  span {
    display: block;
    margin-top: 5px;
    color: rgba(255, 255, 255, 0.78);
    font-size: 11px;
    font-weight: 700;
    line-height: 1.35;
  }
}

.source-welfare-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 10px;
  border-radius: 18px;
  padding: 5px;
  background: rgba(0, 0, 0, 0.26);

  button {
    height: 38px;
    border: 0;
    border-radius: 14px;
    color: rgba(255, 255, 255, 0.76);
    background: transparent;
    font-size: 13px;
    font-weight: 900;
  }

  .is-active {
    color: #743700;
    background: linear-gradient(180deg, #fff6b0 0%, #ffbe3d 100%);
    box-shadow:
      inset 0 2px 0 rgba(255, 255, 255, 0.65),
      0 7px 14px rgba(63, 22, 0, 0.24);
  }
}

.source-table-head {
  display: grid;
  grid-template-columns: 1.2fr 0.9fr 1.15fr;
  min-height: 38px;
  overflow: hidden;
  border-radius: 14px 14px 0 0;
  background: linear-gradient(180deg, rgba(75, 127, 255, 0.96), rgba(44, 91, 207, 0.96));

  span {
    display: grid;
    place-items: center;
    border-right: 1px solid rgba(8, 15, 41, 0.55);
    color: #ffffff;
    font-size: 11px;
    font-weight: 900;
    line-height: 1.15;
    text-align: center;
  }

  span:last-child {
    border-right: 0;
  }
}

.source-card-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 9px;
}

.source-card-grid--table {
  grid-template-columns: 1fr;
  gap: 0;
  overflow: hidden;
  border-radius: 0 0 14px 14px;
}

.source-card-grid--welfare,
.source-card-grid--chest {
  grid-template-columns: 1fr;
}

.source-card {
  position: relative;
  min-height: 84px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 13px;
  padding: 11px 10px;
  background: rgba(0, 0, 0, 0.32);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.12);

  &::after {
    position: absolute;
    inset: 0;
    background: linear-gradient(120deg, transparent 0 44%, rgba(255, 255, 255, 0.12) 50%, transparent 56% 100%);
    transform: translateX(-100%);
    animation: source-card-shine 4s ease-in-out infinite;
    content: "";
  }

  span,
  strong,
  em {
    position: relative;
    z-index: 1;
    display: block;
  }

  span {
    color: rgba(255, 255, 255, 0.72);
    font-size: 11px;
    font-weight: 600;
    line-height: 1.25;
  }

  strong {
    margin-top: 7px;
    color: #fff994;
    font-size: 17px;
    font-weight: 900;
    line-height: 1.1;
    text-shadow: 0 2px 1px rgba(58, 7, 7, 0.38);
  }

  em {
    margin-top: 5px;
    color: #ffffff;
    font-size: 11px;
    font-style: normal;
    font-weight: 700;
    line-height: 1.2;
  }
}

.source-card__copy {
  position: relative;
  z-index: 1;
}

.source-card-grid--table .source-card {
  display: grid;
  min-height: 51px;
  grid-template-columns: 1.2fr 0.9fr 1.15fr;
  align-items: stretch;
  border-width: 0 1px 1px;
  border-radius: 0;
  padding: 0;
  background: rgba(9, 18, 49, 0.62);
}

.source-card-grid--table .source-card__copy {
  display: contents;
}

.source-card-grid--table span,
.source-card-grid--table strong,
.source-card-grid--table em {
  display: grid;
  min-width: 0;
  height: 100%;
  place-items: center;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  padding: 7px 6px;
  margin: 0;
  text-align: center;
  word-break: break-word;
}

.source-card-grid--table em {
  border-right: 0;
}

.source-card-grid--table span {
  color: #ffffff;
  font-size: 10px;
  font-weight: 800;
}

.source-card-grid--table strong {
  color: #fff994;
  font-size: 13px;
  font-weight: 900;
}

.source-card-grid--table em {
  color: rgba(255, 255, 255, 0.86);
  font-size: 9px;
  font-weight: 700;
  line-height: 1.25;
}

.source-card--welfare {
  display: grid;
  min-height: 118px;
  grid-template-columns: 86px 1fr;
  align-items: center;
  gap: 10px;
  border-color: rgba(255, 232, 135, 0.3);
  border-radius: 18px;
  padding: 10px;
  background:
    radial-gradient(circle at 8% 14%, rgba(255, 240, 157, 0.28), transparent 34%),
    linear-gradient(135deg, rgba(108, 69, 27, 0.72), rgba(39, 31, 70, 0.88));
}

.source-card__thumb {
  position: relative;
  z-index: 1;
  width: 82px;
  height: 82px;
  object-fit: contain;
  filter: drop-shadow(0 8px 12px rgba(40, 18, 0, 0.34));
}

.source-card--welfare strong {
  color: #fff3a0;
  font-size: 19px;
}

.source-card--welfare em {
  color: rgba(255, 255, 255, 0.82);
  font-size: 10px;
}

.source-activity--first-deposit .source-first-reward-pair strong {
  margin: 2px 0 0;
  color: #fbd89d;
  font-size: 10px;
  line-height: 1.05;
  text-shadow:
    0 1px 0 #c41214,
    0 0 4px rgba(255, 255, 255, 0.2);
}

.source-activity--first-deposit .source-first-reward-pair > div:nth-child(2) strong {
  color: #dcefe0;
  text-shadow:
    0 1px 0 #622ccd,
    0 0 4px rgba(255, 255, 255, 0.18);
}

.source-activity--first-deposit .source-first-reward-pair__icon {
  width: 26px;
  height: 26px;
}

.source-rule-card {
  margin-top: 13px;
  border-radius: 16px;
  padding: 14px;
  background: rgba(0, 0, 0, 0.27);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);

  h2 {
    margin: 0 0 10px;
    color: #ffffff;
    font-size: 16px;
    font-weight: 900;
  }

  ol {
    margin: 0;
    padding-left: 18px;
    color: rgba(255, 255, 255, 0.78);
    font-size: 12px;
    font-weight: 500;
    line-height: 1.65;
  }

  li + li {
    margin-top: 5px;
  }
}

.source-contact {
  position: fixed;
  top: min(374px, calc(100dvh - 154px));
  right: max(calc((100vw - 430px) / 2), 0px);
  bottom: auto;
  z-index: 34;
}

.source-activity--deposit-bonus .source-contact {
  top: min(548px, calc(100dvh - 190px));
  right: max(calc((100vw - 430px) / 2), 0px);
}

.source-activity--daily-deposit .source-contact {
  top: min(548px, calc(100dvh - 190px));
  right: max(calc((100vw - 430px) / 2), 0px);
}

.source-activity--wheel .source-contact {
  top: min(548px, calc(100dvh - 190px));
  right: max(calc((100vw - 430px) / 2), 0px);
}

.source-first-fixed-treasure {
  position: fixed;
  left: 50%;
  bottom: 0;
  z-index: 32;
  display: block;
  width: 100vw;
  max-width: 430px;
  height: auto;
  object-fit: contain;
  object-position: center bottom;
  pointer-events: none;
  transform: translateX(-50%);
  user-select: none;
}

.source-rules-dialog {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 50%;
  z-index: 36;
  display: flex;
  width: min(430px, 100vw);
  align-items: flex-start;
  justify-content: center;
  overflow-y: auto;
  padding: 48px 20px 18px;
  background: rgba(3, 7, 24, 0.76);
  backdrop-filter: blur(2px);
  transform: translateX(-50%);
}

.source-rules-dialog__body {
  position: relative;
  width: min(398px, 100%);
}

.source-rules-dialog__bundle {
  position: relative;
  width: 100%;
  height: 128px;
  pointer-events: none;

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
    filter: drop-shadow(0 12px 16px rgba(22, 7, 60, 0.36));
  }
}

.source-rules-dialog__content {
  position: relative;
}

.source-rules-dialog__banner {
  position: absolute;
  top: -10px;
  right: -10px;
  left: -10px;
  z-index: 4;
  pointer-events: none;
  transform: translateY(-50%);

  img {
    display: block;
    width: 100%;
    aspect-ratio: 1308 / 278;
    object-fit: contain;
  }

  span {
    position: absolute;
    top: 40%;
    right: 0;
    left: 0;
    color: #ffe591;
    font-size: 26px;
    font-style: italic;
    font-weight: 800;
    line-height: 1;
    text-align: center;
    text-shadow: 0 2px 3px rgba(0, 0, 0, 0.4);
    transform: translateY(-50%);
  }
}

.source-rules-dialog__frame {
  border-radius: 16px;
  padding: 7px;
  background: linear-gradient(180deg, #a636f3 0%, #a994ff 100%);
  box-shadow:
    0 0 14px rgba(255, 255, 255, 0.25),
    inset 0 0 4px rgba(255, 255, 255, 0.55);
}

.source-rules-dialog__inner {
  height: clamp(420px, 57dvh, 500px);
  overflow: hidden;
  border-radius: 12px;
  background: linear-gradient(180deg, #f8efff 0%, #f0ebff 100%);
  box-shadow: 0 0 4px rgba(0, 0, 0, 0.25);
}

.source-rules-dialog__scroll {
  height: 100%;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 12px 14px 14px;
  color: #774cf3;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.52;

  ol {
    margin: 0;
    padding-left: 24px;
    list-style: decimal outside;
  }

  li {
    display: list-item;
    margin: 0;
  }

  li + li {
    margin-top: 7px;
  }

  p {
    margin: 0;
  }

  strong {
    font-weight: 800;
  }
}

.source-rules-dialog__actions {
  display: flex;
  justify-content: center;
  margin-top: 18px;
}

.source-rules-dialog__ok {
  position: relative;
  z-index: 4;
  display: inline-flex;
  width: 156px;
  height: 38px;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 28px;
  padding: 0 30px 4px;
  color: #9d3704;
  background: transparent;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.14);
}

.source-rules-dialog__ok-base,
.source-rules-dialog__ok-top,
.source-rules-dialog__ok-shine,
.source-rules-dialog__ok-spark {
  position: absolute;
  pointer-events: none;
}

.source-rules-dialog__ok-base {
  inset: 5px 0 0;
  border: 0.5px solid #ffa43a;
  border-radius: 28px;
  background: linear-gradient(180deg, #ffdc34 0%, #fe762d 100%);
  box-shadow: inset 0 -1px 4px #ff971d;
}

.source-rules-dialog__ok-top {
  inset: 0 0 5px;
  border-bottom: 1px solid rgba(255, 221, 163, 0.4);
  border-radius: 18px;
  background: linear-gradient(180deg, #fff0b0 0%, #ffbc18 100%);
  box-shadow:
    inset 0 0 4px #fff2c6,
    inset 0 -4px 4px rgba(255, 255, 255, 0.25);
}

.source-rules-dialog__ok-shine {
  top: 3px;
  right: 7px;
  left: 7px;
  height: 20px;
  border-radius: 999px;
  background: linear-gradient(180deg, #fffdba 0%, rgba(255, 253, 186, 0.04) 100%);
}

.source-rules-dialog__ok-spark {
  top: 6px;
  left: 12px;
  width: 7px;
  height: 4px;
  border-radius: 999px;
  background: #ffffff;
  filter: blur(0.5px);
  transform: rotate(-12deg);
}

.source-rules-dialog__ok-label {
  position: relative;
  z-index: 2;
  color: #973305;
  font-size: 15px;
  font-weight: 600;
  line-height: 1;
}

.source-rules-dialog--deposit-bonus {
  align-items: flex-start;
  padding-top: 72px;
  background: rgba(18, 3, 3, 0.78);
}

.source-rules-dialog--deposit-bonus .source-rules-dialog__body {
  width: min(340px, 100%);
}

.source-rules-dialog--deposit-bonus .source-rules-dialog__bundle {
  z-index: 3;
  height: 168px;
  margin-bottom: -70px;
  transform: translateY(-24px);

  img {
    filter: drop-shadow(0 12px 16px rgba(44, 7, 7, 0.34));
  }
}

.source-rules-dialog--deposit-bonus .source-rules-dialog__banner {
  top: -5px;
  right: -12px;
  left: -12px;

  span {
    top: 49%;
    color: #ffe591;
    font-size: 24px;
    font-weight: 900;
    text-shadow: 0 2px 2px rgba(80, 21, 8, 0.5);
  }
}

.source-rules-dialog--deposit-bonus .source-rules-dialog__frame {
  border-radius: 15px;
  padding: 7px;
  background: linear-gradient(180deg, #ffe2a4 0%, #ffb956 100%);
  box-shadow:
    0 0 12px rgba(255, 226, 164, 0.34),
    inset 0 0 5px rgba(255, 255, 255, 0.62);
}

.source-rules-dialog--deposit-bonus .source-rules-dialog__inner {
  height: clamp(356px, 50dvh, 408px);
  border-radius: 12px;
  background: linear-gradient(180deg, #fff8ef 0%, #fff0dc 100%);
}

.source-rules-dialog--deposit-bonus .source-rules-dialog__scroll {
  padding: 48px 18px 16px;
  color: #774cf3;
  font-size: 14px;
  line-height: 1.54;
}

.source-rules-dialog--deposit-bonus .source-rules-dialog__actions {
  margin-top: 33px;
}

.source-rules-dialog--daily-deposit {
  align-items: flex-start;
  padding-top: 65px;
  background: rgba(12, 3, 3, 0.78);
}

.source-rules-dialog--daily-deposit .source-rules-dialog__body {
  width: min(340px, 100%);
}

.source-rules-dialog--daily-deposit .source-rules-dialog__bundle {
  z-index: 3;
  height: 128px;
  margin-bottom: 13px;

  img {
    width: 204px;
    margin: 0 auto;
    filter: drop-shadow(0 12px 16px rgba(44, 7, 7, 0.34));
  }
}

.source-rules-dialog--daily-deposit .source-rules-dialog__banner {
  top: -8px;
  right: 0;
  left: 0;

  img {
    height: 59px;
    aspect-ratio: auto;
  }

  span {
    top: 49%;
    color: #fff7a3;
    font-size: 23px;
    font-style: italic;
    font-weight: 900;
    text-shadow:
      0 2px 0 rgba(87, 21, 8, 0.9),
      0 4px 8px rgba(42, 0, 0, 0.35);
  }
}

.source-rules-dialog--daily-deposit .source-rules-dialog__frame {
  border-radius: 13px;
  padding: 7px;
  background: linear-gradient(180deg, #c55134 0%, #f1c48e 100%);
  box-shadow:
    0 0 12px rgba(255, 226, 164, 0.28),
    inset 0 0 5px rgba(255, 255, 255, 0.52);
}

.source-rules-dialog--daily-deposit .source-rules-dialog__inner {
  height: clamp(372px, 53dvh, 400px);
  border-radius: 10px;
  background: linear-gradient(180deg, #fffaf2 0%, #fff5ea 100%);
}

.source-rules-dialog--daily-deposit .source-rules-dialog__scroll {
  padding: 22px 18px 16px;
  color: #774cf3;
  font-size: 14px;
  line-height: 1.54;
}

.source-rules-dialog--daily-deposit .source-rules-dialog__actions {
  margin-top: 31px;
}

.source-rules-dialog--wheel {
  align-items: flex-start;
  padding-top: 20px;
  background: rgba(4, 7, 22, 0.74);
}

.source-rules-dialog--wheel .source-rules-dialog__body {
  width: min(315px, calc(100vw - 32px));
}

.source-rules-dialog--wheel .source-rules-dialog__bundle {
  z-index: 3;
  height: 118px;
  margin-bottom: 0;
  transform: translateY(0);

  img {
    width: 210px;
    margin: 0 auto;
    filter: drop-shadow(0 12px 16px rgba(30, 8, 70, 0.36));
  }
}

.source-rules-dialog--wheel .source-rules-dialog__banner {
  top: -2px;
  right: -31px;
  left: -31px;

  span {
    top: 47%;
    color: #ffe591;
    font-size: 22px;
    font-weight: 900;
    text-shadow: 0 2px 3px rgba(0, 0, 0, 0.42);
  }
}

.source-rules-dialog--wheel .source-rules-dialog__frame {
  border-radius: 16px;
  padding: 7px;
  background: linear-gradient(180deg, #8149cf 0%, #d194ff 100%);
  box-shadow:
    0 0 14px rgba(255, 255, 255, 0.25),
    inset 0 0 4px rgba(255, 255, 255, 0.55);
}

.source-rules-dialog--wheel .source-rules-dialog__inner {
  height: clamp(328px, 45dvh, 386px);
  border-radius: 12px;
  background: linear-gradient(180deg, #f5efff 0%, #f6ebff 100%);
}

.source-rules-dialog--wheel .source-rules-dialog__scroll {
  padding: 22px 18px 17px;
  color: #774cf3;
  font-size: 13px;
  line-height: 1.52;
}

.source-rules-dialog--wheel .source-rules-dialog__actions {
  margin-top: 18px;
}

.source-rules-dialog-enter-active,
.source-rules-dialog-leave-active {
  transition: opacity 0.18s ease;
}

.source-rules-dialog-enter-from,
.source-rules-dialog-leave-to {
  opacity: 0;
}

.source-rules-dialog-enter-active .source-rules-dialog__body,
.source-rules-dialog-leave-active .source-rules-dialog__body {
  transition:
    opacity 0.2s ease,
    transform 0.24s cubic-bezier(0.2, 0.9, 0.25, 1.12);
}

.source-rules-dialog-enter-from .source-rules-dialog__body,
.source-rules-dialog-leave-to .source-rules-dialog__body {
  opacity: 0;
  transform: translateY(18px) scale(0.94);
}

.activity-not-found {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 18px;
  padding: 24px;
  text-align: center;

  header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  button {
    height: 42px;
    border: 0;
    border-radius: 12px;
    color: #ffffff;
    background: #397fe5;
    font-weight: 800;
  }

  svg {
    width: 22px;
    height: 22px;
    fill: currentColor;
  }
}

@keyframes source-card-shine {
  0%,
  58% {
    transform: translateX(-105%);
  }

  100% {
    transform: translateX(105%);
  }
}

@keyframes source-deposit-spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes source-deposit-chest-pulse {
  from {
    transform: translateX(-50%) scale(0.985);
  }

  to {
    transform: translateX(-50%) scale(1.035);
  }
}

@keyframes source-deposit-arc-flow {
  to {
    stroke-dashoffset: -8.3;
  }
}

@keyframes source-daily-main-float {
  from {
    transform: translate3d(0, 0, 0) scale(0.985);
  }

  to {
    transform: translate3d(-2px, -7px, 0) scale(1.025);
  }
}

@keyframes source-wheel-intro-launch {
  0% {
    filter: brightness(1);
    opacity: 1;
    transform: translate3d(0, 0, 0) scale(1);
  }

  34% {
    filter: brightness(1.2);
    opacity: 1;
    transform: translate3d(0, -7px, 0) scale(1.025);
  }

  100% {
    filter: brightness(1.08) blur(1px);
    opacity: 0;
    transform: translate3d(0, -44px, 0) scale(1.12);
  }
}

@keyframes source-wheel-spin-loader {
  to {
    transform: rotate(360deg);
  }
}

@keyframes source-wheel-game-spin {
  from {
    transform: translateX(-50%) rotate(0deg);
  }

  to {
    transform: translateX(-50%) rotate(360deg);
  }
}

@keyframes source-wheel-star-pulse {
  from {
    opacity: 0.42;
    transform: translate3d(0, 0, 0) scale(0.92) rotate(-4deg);
  }

  to {
    opacity: 0.9;
    transform: translate3d(0, -5px, 0) scale(1.08) rotate(4deg);
  }
}

@keyframes source-wheel-flags-sway {
  0% {
    filter: saturate(1);
    transform: translate3d(0, 0, 0) rotate(-0.25deg) scale(1.002);
  }

  50% {
    filter: saturate(1.08) brightness(1.03);
    transform: translate3d(0, 4px, 0) rotate(0.35deg) scale(1.006);
  }

  100% {
    filter: saturate(1.02);
    transform: translate3d(0, 1px, 0) rotate(-0.15deg) scale(1.004);
  }
}

@keyframes source-wheel-notice-scroll {
  from {
    transform: translate3d(0, 0, 0);
  }

  to {
    transform: translate3d(0, -50%, 0);
  }
}

@media (max-width: 374px) {
  .source-hero__copy h1 {
    font-size: 28px;
  }

  .source-stage {
    height: 322px;
  }

  .source-wheel__wheel {
    width: 238px;
  }

  .source-wheel__ring {
    width: 258px;
  }
}
</style>
