<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ActivityGuideFlow from '../components/ActivityGuideFlow.vue'
import AppHeader from '../components/AppHeader.vue'
import { h5Api, type H5OffersPayload, type H5VipPayload } from '../services/h5Api'
import {
  isActivityUnread,
  loadCachedOffers,
  markActivityRead,
  onActivityReadChange,
  publishActivityUnreadState,
} from '../services/activityReadState'
import { formatCurrency } from '../services/runtimeConfig'
import { isAndroidAppRuntime } from '../utils/runtimeEnvironment'
import rewardsEntry from '../assets/images/reference/vip-rewards-entry.webp'
import vipZeroLabel from '../assets/images/reference/vip-label-zero.webp'
import {
  offerEvents as mockOfferEvents,
  vipAssets,
  vipRequirements as mockVipRequirements,
  type VipBenefit,
  type OfferEvent,
  type OfferTab,
  type VipRequirementTab,
} from '../mock/offers'

const firstQueryValue = (value: unknown) => (Array.isArray(value) ? value[0] : value)

const normalizeOfferTab = (value: unknown): OfferTab => {
  const raw = firstQueryValue(value)
  if (raw === 'vip') return 'vip'
  return 'events'
}

const route = useRoute()
const router = useRouter()
const androidAppRuntime = isAndroidAppRuntime()
const activeTab = ref<OfferTab>(route.name === 'vip' ? 'vip' : normalizeOfferTab(route.query.tab))
const requirementTab = ref<VipRequirementTab>('level-up')
const selectedVipLevel = ref(1)
const guideVisible = ref(false)
const offerEvents = ref<OfferEvent[]>(mockOfferEvents)
const vipLevels = ref<H5VipPayload['levels']>([])
const vipRequirementSource = ref<H5VipPayload['requirements']>([])
const currentVipLevel = ref(0)
const mockEventBySlug = new Map(mockOfferEvents.flatMap((event) => (
  event.detailSlug ? [[event.detailSlug, event] as const] : []
)))
const mockEventById = new Map(mockOfferEvents.map((event) => [event.id, event] as const))

const isDownloadEvent = (event: Pick<OfferEvent, 'id' | 'detailSlug'>) =>
  event.id === 'app-download' || event.detailSlug === 'app-download'

const formatMoney = (value: unknown) => {
  const numeric = Number(String(value ?? '0').replace(/,/g, ''))
  if (!Number.isFinite(numeric)) return String(value ?? '0.00')

  return numeric.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

const normalizeOfferEvents = (payload: H5OffersPayload) => {
  const events = (payload.events || [])
    .filter((event) => event.cardImageURL)
    .map((event) => {
      const localEvent = mockEventBySlug.get(event.slug) || mockEventById.get(event.activityCode)

      return {
        id: event.activityCode || localEvent?.id || String(event.id),
        title: event.title || localEvent?.title || '',
        image: localEvent?.image || event.cardImageURL,
        detailSlug: event.slug || localEvent?.detailSlug,
        targetPath: localEvent?.targetPath,
        activityCode: event.activityCode,
        updatedAt: event.updatedAt,
        unread: isActivityUnread(event),
      }
    })
    .filter((event) => !androidAppRuntime || !isDownloadEvent(event))

  if (events.length) return events
  return mockOfferEvents.map((event, index) => {
    const readState = {
      id: index + 1,
      activityCode: event.id,
      slug: event.detailSlug || event.id,
      title: event.title,
      cardImageURL: event.image,
      updatedAt: 'bundled-v1',
    }
    return {
      ...event,
      activityCode: event.id,
      updatedAt: readState.updatedAt,
      unread: isActivityUnread(readState),
    }
  }).filter((event) => !androidAppRuntime || !isDownloadEvent(event))
}

const fallbackVipRows = computed(() =>
  mockVipRequirements.map((row) => ({
    ...row,
    period: row.level <= 5 ? '7 Days' : '30 Days',
  })),
)

const levelUpRequirementRows = computed(() => {
  const rows = vipRequirementSource.value
    .filter((row) => row.requirementType === 'level_up')
    .map((row) => ({
      level: Number(row.targetLevelNo),
      image:
        vipLevels.value.find((level) => Number(level.levelNo) === Number(row.targetLevelNo))?.iconURL ||
        mockVipRequirements.find((item) => item.level === Number(row.targetLevelNo))?.image ||
        vipAssets.current,
      deposit: formatMoney(row.depositAmount),
      validBets: formatMoney(row.validBetAmount),
      period: row.periodDays ? `${row.periodDays} Days` : '7 Days',
    }))
    .sort((a, b) => a.level - b.level)

  return rows.length ? rows : fallbackVipRows.value
})

const retainRequirementRows = computed(() => {
  const rows = vipRequirementSource.value
    .filter((row) => row.requirementType === 'retain')
    .map((row) => ({
      level: Number(row.targetLevelNo),
      image:
        vipLevels.value.find((level) => Number(level.levelNo) === Number(row.targetLevelNo))?.iconURL ||
        mockVipRequirements.find((item) => item.level === Number(row.targetLevelNo))?.image ||
        vipAssets.current,
      deposit: formatMoney(row.depositAmount),
      validBets: formatMoney(row.validBetAmount),
      period: row.periodDays ? `${row.periodDays} Days` : '30 Days',
    }))
    .sort((a, b) => a.level - b.level)

  return rows.length ? rows : fallbackVipRows.value
})

const maxVipLevel = computed(() => {
  const fromApi = vipLevels.value.at(-1)?.levelNo
  return fromApi ? Number(fromApi) : mockVipRequirements.at(-1)?.level ?? 15
})

const selectedVipInfo = computed(() =>
  vipLevels.value.find((level) => Number(level.levelNo) === selectedVipLevel.value),
)

const currentVipIcon = computed(() => {
  if (currentVipLevel.value === 0) return vipZeroLabel
  return vipLevels.value.find((level) => Number(level.levelNo) === currentVipLevel.value)?.iconURL || vipAssets.current
})

const nextVipRequirement = computed(
  () =>
    levelUpRequirementRows.value.find((row) => row.level === Math.max(currentVipLevel.value + 1, 1)) ||
    levelUpRequirementRows.value[0],
)

const selectedVipBenefits = computed<VipBenefit[]>(() => {
  const level = selectedVipLevel.value
  const info = selectedVipInfo.value

  if (info) {
    return [
      {
        label: 'Level Up Reward',
        value: formatCurrency(info.levelUpReward),
        delta: `+${formatMoney(info.levelUpReward)}`,
        type: 'reward',
      },
      {
        label: 'Weekly Reward',
        value: formatCurrency(info.weeklyReward),
        delta: `+${formatMoney(info.weeklyReward)}`,
        type: 'reward',
      },
      {
        label: 'Monthly Reward',
        value: formatCurrency(info.monthlyReward),
        delta: `+${formatMoney(info.monthlyReward)}`,
        type: 'reward',
      },
      {
        label: 'Additional Daily Withdrawal Count',
        value: String(info.dailyWithdrawCount || level),
        delta: `+${info.dailyWithdrawCount || level}`,
        type: 'privilege',
      },
    ]
  }

  const amount = (value: number) => formatMoney(value * level)

  return [
    {
      label: 'Level Up Reward',
      value: formatCurrency(5 * level),
      delta: `+${amount(5)}`,
      type: 'reward',
    },
    {
      label: 'Weekly Reward',
      value: formatCurrency(1 * level),
      delta: `+${amount(1)}`,
      type: 'reward',
    },
    {
      label: 'Monthly Reward',
      value: formatCurrency(3 * level),
      delta: `+${amount(3)}`,
      type: 'reward',
    },
    {
      label: 'Additional Daily Withdrawal Count',
      value: String(level),
      delta: `+${level}`,
      type: 'privilege',
    },
  ]
})

const rewardBenefits = computed(() =>
  selectedVipBenefits.value.filter((benefit) => benefit.type === 'reward'),
)
const privilegeBenefits = computed(() =>
  selectedVipBenefits.value.filter((benefit) => benefit.type === 'privilege'),
)

const requirementRows = computed(() =>
  requirementTab.value === 'retain' ? retainRequirementRows.value : levelUpRequirementRows.value,
)

const requirementHeaders = computed(() =>
  requirementTab.value === 'retain'
    ? ['Level', 'Period', 'Deposit', 'Valid bets']
    : ['Level', 'Base Level', 'Deposit', 'Valid bets'],
)



const openLogin = () => {
  void router.push({ name: 'auth-login' })
}

const openRegister = () => {
  void router.push({ name: 'auth-register' })
}


const openDownloadGuide = () => {
  if (androidAppRuntime) return
  guideVisible.value = true
}

const closeGuide = () => {
  guideVisible.value = false
}

const showPreviousVip = () => {
  selectedVipLevel.value =
    selectedVipLevel.value === 1 ? maxVipLevel.value : selectedVipLevel.value - 1
}

const showNextVip = () => {
  selectedVipLevel.value =
    selectedVipLevel.value === maxVipLevel.value ? 1 : selectedVipLevel.value + 1
}

const openOffer = (event: OfferEvent) => {
  if (event.activityCode) {
    markActivityRead({
      id: Number(event.id) || 0,
      activityCode: event.activityCode,
      slug: event.detailSlug || '',
      title: event.title,
      cardImageURL: event.image,
      updatedAt: event.updatedAt || '',
    })
    event.unread = false
  }
  if (event.targetPath) {
    void router.push(event.targetPath)
    return
  }

  if (event.detailSlug) {
    void router.push({ path: `/activity/${event.detailSlug}` })
    return
  }

  if (event.id === 'app-download') {
    openDownloadGuide()
  }
}

const loadOffers = async () => {
  try {
    const [offers, vip] = await Promise.all([loadCachedOffers(), h5Api.getVip()])
    offerEvents.value = normalizeOfferEvents(offers)
    vipLevels.value = vip.levels || []
    vipRequirementSource.value = vip.requirements || []
    currentVipLevel.value = Number(vip.currentLevel || 0)
  } catch (error) {
    console.warn('[api] falling back to local offers/vip mock data.', error)
    offerEvents.value = normalizeOfferEvents({
      events: [],
      vip: { levels: [], benefits: [] },
    })
  } finally {
    publishActivityUnreadState(offerEvents.value.some((event) => event.unread))
  }
}

const refreshUnreadState = () => {
  offerEvents.value = offerEvents.value.map((event) => {
    if (!event.activityCode) return event
    return {
      ...event,
      unread: isActivityUnread({
        id: Number(event.id) || 0,
        activityCode: event.activityCode,
        slug: event.detailSlug || '',
        title: event.title,
        cardImageURL: event.image,
        updatedAt: event.updatedAt || '',
      }),
    }
  })
  publishActivityUnreadState(offerEvents.value.some((event) => event.unread))
}

let stopActivityReadListener: () => void = () => undefined

onMounted(() => {
  stopActivityReadListener = onActivityReadChange(refreshUnreadState)
  void loadOffers()
})

onBeforeUnmount(() => stopActivityReadListener())

watch(
  () => [route.name, route.query.tab] as const,
  async ([name, tab]) => {
    const nextTab = name === 'vip' ? 'vip' : normalizeOfferTab(tab)
    if (activeTab.value === nextTab) return

    activeTab.value = nextTab
    await nextTick()
    window.scrollTo({ top: 0, behavior: 'auto' })
  },
)
</script>

<template>
  <div class="offers-shell text-white">
    <AppHeader @login="openLogin" @register="openRegister" />

    <main class="offers-content">
      <section v-if="activeTab === 'events'" class="events-pane" aria-label="Events">
        <button
          v-for="event in offerEvents"
          :key="event.id"
          type="button"
          class="event-card"
          :aria-label="event.title"
          @click="openOffer(event)"
        >
          <img :src="event.image" :alt="event.title" loading="lazy" decoding="async" />
        </button>
      </section>

      <section v-else class="vip-pane" aria-label="VIP">
        <article class="vip-card">
          <div class="vip-card__shine" aria-hidden="true" />
          <img class="vip-card__level" :src="currentVipIcon" :alt="`VIP${currentVipLevel}`" loading="eager" />

          <div class="vip-card__deposit">
            <img :src="vipAssets.deposit" alt="" aria-hidden="true" />
            <span>Deposit:</span>
            <strong>0.00</strong>
            <span>/ {{ nextVipRequirement?.deposit ?? '100.00' }}</span>
            <em>to VIP{{ nextVipRequirement?.level ?? 1 }}</em>
          </div>

          <div
            class="vip-card__progress"
            role="progressbar"
            aria-valuenow="0"
            aria-valuemin="0"
            aria-valuemax="100"
          >
            <span />
          </div>
        </article>

        <button class="vip-rewards-entry" type="button" @click="router.push('/activity/welfare-card')"><img :src="rewardsEntry" alt="" /><strong>UNLOCK MORE REWARDS</strong><span>Explore weekly &amp; monthly reward cards</span><em>View Rewards</em></button>
        <div class="vip-tip">
          <img :src="vipAssets.emptyBonus" alt="" aria-hidden="true" loading="lazy" />
          <span>Upgrade your VIP level, get generous benefits!</span>
        </div>

        <article class="vip-benefits">
          <header class="vip-benefits__header">
            <button
              type="button"
              :aria-label="`Previous VIP${selectedVipLevel === 1 ? maxVipLevel : selectedVipLevel - 1}`"
              @click="showPreviousVip"
            >
              <svg viewBox="0 0 14 14" aria-hidden="true">
                <path
                  d="M6.55 10.77a.8.8 0 0 1-1.13 0l-3.19-3.2a.8.8 0 0 1 0-1.14l3.19-3.2a.8.8 0 0 1 1.13 1.14L4.73 6.2h6.47a.8.8 0 0 1 0 1.6H4.73l1.82 1.83a.8.8 0 0 1 0 1.14Z"
                />
              </svg>
            </button>
            <h2>VIP{{ selectedVipLevel }} Benefits</h2>
            <button
              type="button"
              :aria-label="`Next VIP${selectedVipLevel === maxVipLevel ? 1 : selectedVipLevel + 1}`"
              @click="showNextVip"
            >
              <svg viewBox="0 0 14 14" aria-hidden="true">
                <path
                  d="M7.45 3.23a.8.8 0 0 1 1.13 0l3.19 3.2a.8.8 0 0 1 0 1.14l-3.19 3.2a.8.8 0 0 1-1.13-1.14L9.27 7.8H2.8a.8.8 0 0 1 0-1.6h6.47L7.45 4.37a.8.8 0 0 1 0-1.14Z"
                />
              </svg>
            </button>
          </header>

          <div class="vip-benefits__body">
            <h3>VIP Benefits</h3>
            <div
              v-for="benefit in rewardBenefits"
              :key="benefit.label"
              class="benefit-row"
              :class="{ 'is-alt': benefit.label === 'Weekly Reward' }"
            >
              <span class="benefit-row__label">{{ benefit.label }}</span>
              <span class="benefit-row__value">
                <span>{{ benefit.value }}</span>
                <span class="benefit-row__delta">{{ benefit.delta }}</span>
                <img :src="vipAssets.info" alt="" aria-hidden="true" loading="lazy" />
              </span>
            </div>

            <h3>VIP Privileges</h3>
            <div v-for="benefit in privilegeBenefits" :key="benefit.label" class="benefit-row benefit-row--privilege">
              <span class="benefit-row__label">{{ benefit.label }}</span>
              <span class="benefit-row__value">
                <span>{{ benefit.value }}</span>
                <span class="benefit-row__delta">{{ benefit.delta }}</span>
              </span>
            </div>
          </div>
        </article>

        <article class="vip-requirements">
          <div class="requirement-tabs">
            <svg
              class="requirement-tabs__art"
              viewBox="0 0 347 36"
              preserveAspectRatio="none"
              aria-hidden="true"
            >
              <defs>
                <linearGradient id="vipRequirementTabActive" x1="198.5" y1="36" x2="198.5" y2="0">
                  <stop stop-color="#3B4873" />
                  <stop offset="0.455621" stop-color="#4C61A9" />
                  <stop offset="1" stop-color="#526AB9" />
                </linearGradient>
                <linearGradient id="vipRequirementTabInactive" x1="156" y1="36" x2="347" y2="36">
                  <stop stop-color="#353E60" />
                  <stop offset="1" stop-color="#42538A" />
                </linearGradient>
              </defs>
              <path
                class="requirement-tabs__inactive-path"
                d="M166.1 10.34C167.61 6.51 171.3 4 175.41 4H337C342.52 4 347 8.48 347 14V36H156L166.1 10.34Z"
                :transform="requirementTab === 'retain' ? 'translate(347 0) scale(-1 1)' : undefined"
              />
              <path
                class="requirement-tabs__active-path"
                d="M198.5 36H0V10C0 4.48 4.48 0 10 0H165.82C170.11 0 173.92 2.73 175.3 6.8L180.67 22.7C183.28 30.41 190.37 35.7 198.5 36Z"
                :transform="requirementTab === 'retain' ? 'translate(347 0) scale(-1 1)' : undefined"
              />
            </svg>
            <button
              type="button"
              :class="{ 'is-active': requirementTab === 'level-up' }"
              @click="requirementTab = 'level-up'"
            >
              Level Up Requirement
            </button>
            <button
              type="button"
              :class="{ 'is-active': requirementTab === 'retain' }"
              @click="requirementTab = 'retain'"
            >
              Retain Requirement
            </button>
          </div>

          <div class="requirement-table">
            <div class="requirement-table__head requirement-table__row">
              <span v-for="header in requirementHeaders" :key="header">{{ header }}</span>
            </div>
            <div
              v-for="row in requirementRows"
              :key="row.level"
              class="requirement-table__row requirement-table__body-row"
            >
              <span><img :src="row.image" :alt="`VIP${row.level}`" loading="lazy" /></span>
              <span>{{ requirementTab === 'retain' ? row.period : `VIP ${row.level}` }}</span>
              <span>{{ row.deposit }}</span>
              <span>{{ row.validBets }}</span>
            </div>
          </div>
        </article>

        <button class="vip-info" type="button">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path
              d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20Zm1 15h-2v-6h2v6Zm0-8h-2V7h2v2Z"
            />
          </svg>
          <span>VIP Info</span>
        </button>
      </section>
    </main>


    <ActivityGuideFlow
      v-if="guideVisible"
      :visible="guideVisible"
      initial="app-download"
      @complete="closeGuide"
    />
  </div>
</template>

<style scoped lang="scss">
.offers-shell {
  position: relative;
  width: 100%;
  min-height: 100vh;
  max-width: 430px;
  margin: 0 auto;
  overflow: clip;
  background: #090f1f;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.04);
}

.offers-content {
  min-height: 100vh;
  padding-bottom: calc(env(safe-area-inset-bottom) + 86px);
  background: #090f1f;
}

.events-pane {
  display: grid;
  gap: 12px;
  padding: 14px 14px 8px;
}

.event-card {
  position: relative;
  width: 100%;
  overflow: visible;
  aspect-ratio: 347 / 140;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 10px;
  background: #151c30;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.24);
  transform: translateZ(0);
  transition:
    transform 180ms ease,
    filter 180ms ease;

  img {
    width: 100%;
    height: 100%;
    object-fit: fill;
    border-radius: inherit;
  }

  &:active {
    filter: brightness(1.08);
    transform: scale(0.985);
  }
}

.vip-pane {
  display: grid;
  gap: 12px;
  padding: 14px 14px 8px;
}

.vip-card {
  position: relative;
  overflow: hidden;
  min-height: 106px;
  border-radius: 12px;
  padding: 8px 12px 14px;
  background: linear-gradient(248deg, #7b96ff 0%, #2e389e 100%);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.28),
    0 12px 26px rgba(21, 30, 102, 0.36);
}

.vip-card__shine {
  position: absolute;
  inset: -36px -42px auto auto;
  width: 150px;
  height: 150px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(139, 248, 253, 0.45), rgba(139, 248, 253, 0));
  animation: vip-glow 2.8s ease-in-out infinite;
}

.vip-card__level {
  position: relative;
  width: auto;
  height: 28px;
  margin: 0 0 20px;
  object-fit: contain;
}

.vip-card__deposit {
  position: relative;
  display: flex;
  align-items: center;
  gap: 5px;
  color: rgba(255, 255, 255, 0.86);
  font-size: 14px;
  font-weight: 500;
  line-height: 1;

  img {
    width: 18px;
    height: 18px;
  }

  strong {
    color: #7bb8ff;
    font-size: 12px;
    font-weight: 700;
  }

  em {
    color: #ffffff;
    font-style: normal;
  }
}

.vip-card__progress {
  position: relative;
  height: 8px;
  margin-top: 4px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(98, 202, 214, 0.2);

  span {
    display: block;
    width: 0;
    height: 100%;
    border-radius: inherit;
    background: linear-gradient(270deg, #87f8fd, #4797dd);
  }
}

.vip-rewards-entry { position: relative; padding: 0; border-radius: 10px; overflow: hidden; text-align: left; }
.vip-rewards-entry strong,.vip-rewards-entry span,.vip-rewards-entry em { position: absolute; left: 14px; max-width: 66%; }
.vip-rewards-entry strong { top: 20px; font-size: 13px; }
.vip-rewards-entry span { top: 43px; color: #b2cefa; font-size: 10px; }
.vip-rewards-entry em { bottom: 15px; padding: 5px 12px; border-radius: 16px; background: #66d9f8; color: #063e7e; font-size: 10px; font-style: normal; font-weight: 700; }
.vip-rewards-entry img { width: 100%; height: auto; display: block; }

.vip-tip {
  position: relative;
  display: flex;
  min-height: 42px;
  align-items: center;
  overflow: visible;
  border-radius: 6px;
  padding: 0 12px 0 40px;
  background: transparent;
  color: #8e9aad;
  font-size: 12px;
  font-weight: 500;

  img {
    position: absolute;
    top: -10px;
    left: -8px;
    width: 46px;
    height: 40px;
    object-fit: contain;
  }
}

.vip-benefits {
  overflow: hidden;
  border: 1px solid #34425f;
  border-radius: 12px;
  padding-bottom: 6px;
  background: #10192e;
}

.vip-requirements {
  overflow: hidden;
  border-radius: 12px 12px 0 0;
  background: #1a2036;
  box-shadow: inset 0 0 0 1px rgba(57, 127, 229, 0.35);
}

.vip-benefits__header {
  display: flex;
  height: 36px;
  align-items: center;
  gap: 8px;
  padding: 0 18px;
  background: transparent;
  border-bottom: 1px solid #34425f;
  color: #eef4ff;

  h2 {
    flex: 1;
    margin: 0;
    text-align: center;
    font-size: 14px;
    font-weight: 800;
    line-height: 1;
  }

  button {
    display: grid;
    width: 24px;
    height: 24px;
    place-items: center;
    color: #ffffff;
  }

  svg {
    width: 14px;
    height: 14px;
    fill: currentColor;
  }
}

.vip-benefits__body {
  padding: 0 6px 8px;

  h3 {
    margin: 0;
    padding: 11px 0 8px 6px;
    color: rgba(255, 255, 255, 0.9);
    text-align: left;
    font-size: 13px;
    font-weight: 600;
    line-height: 1;
  }
}

.benefit-row {
  display: grid;
  grid-template-columns: 36fr 49fr;
  height: 33px;
  gap: 1px;
  border-bottom: 1px dashed #263451;
  color: #ffffff;
  font-size: 12px;
  font-weight: 500;
}

.benefit-row__label,
.benefit-row__value {
  display: flex;
  min-width: 0;
  align-items: center;
  background: transparent;
}

.benefit-row__label {
  justify-content: flex-start;
  padding: 0 5px;
  color: rgba(255, 255, 255, 0.9);
  text-align: center;
}

.benefit-row__value {
  justify-content: space-between;
  padding: 0 8px 0 16px;
  background: transparent;

  img {
    width: 12px;
    height: 12px;
  }
}

.benefit-row__delta {
  color: #8fbeff;
  font-size: 12px;
  font-weight: 700;
}

.benefit-row--privilege {
  grid-template-columns: minmax(0, 1fr) 104px;

  .benefit-row__label {
    justify-content: flex-end;
    padding-right: 8px;
    font-size: 11px;
    white-space: nowrap;
  }

  .benefit-row__value {
    justify-content: flex-start;
    gap: 10px;
    padding-left: 12px;
  }
}

.requirement-tabs {
  position: relative;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  height: 36px;
  overflow: hidden;
  background: transparent;

  .requirement-tabs__art {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }

  .requirement-tabs__active-path {
    fill: url("#vipRequirementTabActive");
  }

  .requirement-tabs__inactive-path {
    fill: url("#vipRequirementTabInactive");
    opacity: 0.4;
  }

  button {
    position: relative;
    z-index: 1;
    color: rgba(255, 255, 255, 0.7);
    font-size: 12px;
    font-style: italic;
    font-weight: 700;
    line-height: 1;
    transition:
      color 180ms ease,
      text-shadow 180ms ease;

    &.is-active {
      color: #ffffff;
      text-shadow: 0 0 5px rgba(255, 255, 255, 0.3);
    }
  }
}

.requirement-table {
  display: grid;
  gap: 1px;
  padding: 1px;
  background:
    linear-gradient(rgba(81, 105, 183, 0) 0%, #397fe5 100%) border-box,
    #1a2036;
}

.requirement-table__row {
  display: grid;
  grid-template-columns: 80fr 84fr 93fr 90fr;
  gap: 1px;

  span {
    display: flex;
    min-width: 0;
    height: 40px;
    align-items: center;
    justify-content: center;
    padding: 0 2px;
    text-align: center;
    font-size: 12px;
    line-height: 1;
  }

  img {
    width: 26px;
    height: 25px;
    object-fit: contain;
  }
}

.requirement-table__head {
  color: #ffffff;
  font-weight: 700;

  span {
    background: #3d8aef;

    &:first-child {
      background: #f59e0b;
    }
  }
}

.requirement-table__body-row {
  color: rgba(255, 255, 255, 0.92);

  span {
    background: #333c5c;
  }
}

.vip-info {
  display: flex;
  height: 28px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
  font-weight: 400;

  svg {
    width: 14px;
    height: 14px;
    fill: currentColor;
  }
}

.has-download-banner {
  :deep(.app-header) {
    top: 38px;
  }

  .offers-tabs {
    top: 104px;
  }
}

@keyframes vip-glow {
  0%,
  100% {
    opacity: 0.55;
    transform: scale(1);
  }

  50% {
    opacity: 0.95;
    transform: scale(1.14);
  }
}

@media (max-width: 374px) {
  .events-pane,
  .vip-pane {
    padding-right: 10px;
    padding-left: 10px;
  }

  .benefit-row--privilege {
    grid-template-columns: minmax(0, 1fr) 92px;

    .benefit-row__label {
      font-size: 10px;
    }
  }
}
</style>
