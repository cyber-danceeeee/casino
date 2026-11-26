<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import gsap from 'gsap'
import ContactMeBadge from '../components/ContactMeBadge.vue'
import { currencySymbol, formatCurrency } from '../services/runtimeConfig'
import { getCachedShareConfig } from '../services/shareConfigCache'
import { buildShareUrl, resolveShareText } from '../utils/shareLinks'
import {
  agentCommissionRules as mockAgentCommissionRules,
  agentTabs,
  calendarDays,
  chestRewards as mockChestRewards,
  chestStats,
  rankingDate,
  rankingRows,
  rankingTopUsers,
  shareActions as mockShareActions,
  shareAssets,
  shareHomeEntries as mockShareHomeEntries,
  shareTabs,
  selfRanking,
  socialLinks as mockSocialLinks,
  type AgentCommissionRule,
  type AgentTab,
  type ShareAction,
  type ShareHomeEntry,
  type ShareTab,
  type SocialLink,
} from '../mock/share'

const firstQueryValue = (value: unknown) => (Array.isArray(value) ? value[0] : value)

const normalizeShareTab = (value: unknown): ShareTab => {
  const raw = firstQueryValue(value)
  if (raw === 'social' || raw === 'share') return 'share'
  if (raw === 'mine') return 'agent'
  if (raw === 'ranking' || raw === 'agent') return raw
  return 'share'
}

const shareTabToQuery = (tab: ShareTab) => {
  if (tab === 'share') return 'social'
  if (tab === 'agent') return 'mine'
  return tab
}

interface AgentProfilePayload {
  displayName: string
  userNo: string
  inviteCode: string
  avatarUrl: string
}

interface AgentSummaryPayload {
  claimableCommission: string
  totalSettlement: string
  directSettlement: string
  indirectSettlement: string
  directReferrals: number
  indirectReferrals: number
  directCommission: string
  indirectCommission: string
  directRate: string
  indirectRate: string
  totalBettors: number
  validBetAmount: string
  settlementPeriod: string
}

interface AgentReferralPayload {
  id: number
  userNo: string
  displayName: string
  phone: string
  avatarUrl: string
  level: string
  registeredAt: string
}

interface AgentCommissionPayload {
  id: number
  periodStart: string
  periodEnd: string
  referralsCount: number
  validBetAmount: string
  commissionAmount: string
  bonusAmount: string
  status: number
}

const route = useRoute()
const router = useRouter()
const activeTab = ref<ShareTab>(normalizeShareTab(route.query.tab))
const activeAgentTab = ref<AgentTab>('overview')
const scrollEl = ref<HTMLElement | null>(null)
const selfRankEntriesLayer = ref<HTMLElement | null>(null)
const selfRankSocialLayer = ref<HTMLElement | null>(null)
const shareActions = ref<ShareAction[]>(mockShareActions)
const socialLinks = ref<SocialLink[]>(mockSocialLinks)
const shareHomeEntries = ref<ShareHomeEntry[]>(mockShareHomeEntries)
const chestRewards = ref<string[]>(mockChestRewards)
const agentCommissionRules = ref<AgentCommissionRule[]>(mockAgentCommissionRules)
const inviteLink = ref('')
const agentProfile = ref<AgentProfilePayload>({
  displayName: 'Guest',
  userNo: '--',
  inviteCode: '--',
  avatarUrl: '',
})
const agentSummary = ref<AgentSummaryPayload>({
  claimableCommission: '0.00',
  totalSettlement: '0.00',
  directSettlement: '0.00',
  indirectSettlement: '0.00',
  directReferrals: 0,
  indirectReferrals: 0,
  directCommission: '0.00',
  indirectCommission: '0.00',
  directRate: '0.0000',
  indirectRate: '0.0000',
  totalBettors: 0,
  validBetAmount: '0.00',
  settlementPeriod: '',
})
const agentReferrals = ref<AgentReferralPayload[]>([])
const agentCommissionRecords = ref<AgentCommissionPayload[]>([])
const referralSearch = ref('')
const showDirectReferrals = ref(false)
const showIndirectReferrals = ref(false)
const localShareEntryImages = new Map(mockShareHomeEntries.map((entry) => [entry.id, entry.image]))

let selfRankTimeline: gsap.core.Timeline | null = null

const agentHeroBackground = computed(() => `url(${shareAssets.background})`)
const filteredAgentReferrals = computed(() => {
  const keyword = referralSearch.value.trim().toLowerCase()
  return agentReferrals.value.filter((referral) => {
    if (!keyword) return true
    return [referral.userNo, referral.displayName, referral.phone]
      .some((value) => String(value || '').toLowerCase().includes(keyword))
  })
})

const tabIndicatorStyle = computed(() => {
  const index = shareTabs.findIndex((tab) => tab.key === activeTab.value)
  return {
    left: `${(index + 0.5) * (100 / shareTabs.length)}%`,
  }
})

const agentSettlementRange = computed(() => {
  const value = agentSummary.value.settlementPeriod
  const match = /^(\d{4})(\d{2})(\d{2})$/.exec(value)
  if (!match) return '--'
  const date = `${match[1]}-${match[2]}-${match[3]}`
  return `${date} 00:00:00 - ${date} 23:59:59`
})

const copyAgentInviteCode = async () => {
  const value = agentProfile.value.inviteCode
  if (!value || value === '--') return
  await navigator.clipboard?.writeText(value).catch(() => undefined)
}

const selectTab = async (tab: ShareTab) => {
  activeTab.value = tab
  const nextQueryTab = shareTabToQuery(tab)
  const currentQueryTab = firstQueryValue(route.query.tab)
  if (currentQueryTab !== nextQueryTab) {
    await router.replace({
      query: {
        ...route.query,
        tab: nextQueryTab,
      },
    })
  }
  await nextTick()
  scrollEl.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

const formatAmount = (value: unknown) => {
  const numeric = Number(String(value ?? '0').replace(/,/g, ''))
  if (!Number.isFinite(numeric)) return String(value ?? '0')

  return numeric.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

const formatPercent = (value: unknown) => {
  const numeric = Number(value)
  if (!Number.isFinite(numeric)) return String(value ?? '0%')
  return `${(numeric * 100).toFixed(2)}%`
}

const defaultInviteLink = () => {
  if (typeof window === 'undefined') return ''
  return `${window.location.origin}/auth/register`
}

const normalizeSocialLink = (link: SocialLink): SocialLink | null => {
  const id = (link.id || link.label.toLowerCase().replace(/\s+/g, '-')).toLowerCase()
  const local = mockSocialLinks.find((item) => item.id === id || item.label.toLowerCase() === link.label.toLowerCase())
  const icon = local?.icon || local?.image || ''
  if (!link.label || !icon) return null

  return {
    ...link,
    id,
    icon,
    image: icon,
    url: link.url || link.shareURL || '',
    shareURL: link.shareURL || link.url || '',
  }
}

const socialIcon = (social: SocialLink) => social.icon || social.image || ''

const writeShareText = async (social?: SocialLink) => {
  const targetUrl = inviteLink.value || defaultInviteLink()
  const text = social ? resolveShareText(social, targetUrl) : targetUrl
  await navigator.clipboard?.writeText(text).catch(() => undefined)
}

const openSocialShare = async (social: SocialLink) => {
  if (typeof window === 'undefined') return

  const targetUrl = inviteLink.value || defaultInviteLink()
  const shareUrl = buildShareUrl(social, targetUrl)
  if (!shareUrl) return

  if (social.id === 'instagram' || social.id === 'tiktok') {
    await writeShareText(social)
  }
  window.open(shareUrl, '_blank', 'noopener,noreferrer')
}

const handleShareAction = async (action: ShareAction) => {
  const label = action.label.toLowerCase()
  if (label.includes('copy')) {
    await writeShareText(socialLinks.value[0])
    return
  }
  if (label.includes('share')) {
    const targetUrl = inviteLink.value || defaultInviteLink()
    const text = resolveShareText(socialLinks.value[0] || {}, targetUrl)
    if (navigator.share) {
      await navigator.share({ title: 'Join 56MVP', text, url: targetUrl }).catch(() => undefined)
      return
    }
    await writeShareText(socialLinks.value[0])
  }
}

const loadShareConfig = async () => {
  try {
    const requestedTab = shareTabToQuery(activeTab.value)
    const payload = await getCachedShareConfig(requestedTab)
    if (requestedTab !== shareTabToQuery(activeTab.value)) return
    const remote = payload as {
      inviteLink?: string
      socialLinks?: SocialLink[]
      actions?: ShareAction[]
      homeEntries?: ShareHomeEntry[]
      chestRewards?: string[]
      agentRules?: Array<{
        minReferrals?: number
        minValidBet?: string
        commissionRate?: string
        bonusRate?: string
      }>
      agent?: {
        profile?: Partial<AgentProfilePayload>
        summary?: Partial<AgentSummaryPayload>
        referrals?: AgentReferralPayload[]
      }
      commissionRecords?: AgentCommissionPayload[]
    }

    inviteLink.value = remote.inviteLink || inviteLink.value || defaultInviteLink()

    if (remote.socialLinks?.length) {
      socialLinks.value = remote.socialLinks
        .map(normalizeSocialLink)
        .filter((link): link is SocialLink => Boolean(link))
    }

    if (remote.actions?.length) {
      shareActions.value = remote.actions
        .filter((action) => action.label)
        .map((action) => ({
          label: action.label,
          tone: ['violet', 'blue', 'rose'].includes(action.tone) ? action.tone : 'blue',
        }))
    }

    if (remote.homeEntries?.length) {
      shareHomeEntries.value = remote.homeEntries
        .filter((entry) => entry.id && entry.title)
        .map((entry) => ({
          ...entry,
          image: localShareEntryImages.get(entry.id) || entry.image,
        }))
        .filter((entry) => Boolean(entry.image))
    }

    if (remote.chestRewards?.length) {
      chestRewards.value = remote.chestRewards
    }

    if (remote.agentRules?.length) {
      agentCommissionRules.value = remote.agentRules.map((rule) => ({
        referrals: `>=${rule.minReferrals ?? 0}`,
        validBets: `>=${formatAmount(rule.minValidBet)}`,
        commission: `${formatPercent(rule.commissionRate)}+${formatPercent(rule.bonusRate)}`,
      }))
    }

    if (remote.agent?.profile) {
      agentProfile.value = {
        ...agentProfile.value,
        ...remote.agent.profile,
      }
    }
    if (remote.agent?.summary) {
      agentSummary.value = {
        ...agentSummary.value,
        ...remote.agent.summary,
      }
    }
    agentReferrals.value = remote.agent?.referrals || []
    agentCommissionRecords.value = remote.commissionRecords || []
  } catch (error) {
    console.warn('[api] falling back to local share config.', error)
  }
}

const stopSelfRankRotation = () => {
  selfRankTimeline?.kill()
  selfRankTimeline = null

  const layers = [selfRankEntriesLayer.value, selfRankSocialLayer.value].filter(Boolean)
  if (layers.length) {
    gsap.set(layers, { clearProps: 'all' })
  }
}

const startSelfRankRotation = async () => {
  await nextTick()
  stopSelfRankRotation()

  const entries = selfRankEntriesLayer.value
  const social = selfRankSocialLayer.value
  if (activeTab.value !== 'ranking' || !entries || !social) return

  gsap.set(entries, { autoAlpha: 1, x: 0, pointerEvents: 'auto' })
  gsap.set(social, { autoAlpha: 0, x: -16, pointerEvents: 'none' })

  selfRankTimeline = gsap
    .timeline({ repeat: -1, repeatDelay: 0.8 })
    .to(entries, {
      autoAlpha: 0,
      x: 16,
      duration: 0.45,
      ease: 'power2.inOut',
      pointerEvents: 'none',
    }, '+=3')
    .to(social, {
      autoAlpha: 1,
      x: 0,
      duration: 0.45,
      ease: 'power2.inOut',
      pointerEvents: 'auto',
    }, '<')
    .to(social, {
      autoAlpha: 0,
      x: -16,
      duration: 0.45,
      ease: 'power2.inOut',
      pointerEvents: 'none',
    }, '+=3')
    .to(entries, {
      autoAlpha: 1,
      x: 0,
      duration: 0.45,
      ease: 'power2.inOut',
      pointerEvents: 'auto',
    }, '<')
}

watch(
  () => route.query.tab,
  async (tab) => {
    const nextTab = normalizeShareTab(tab)
    if (activeTab.value === nextTab) return

    activeTab.value = nextTab
    await nextTick()
    scrollEl.value?.scrollTo({ top: 0, behavior: 'auto' })
  },
)

watch(
  activeTab,
  (tab) => {
    void loadShareConfig()

    if (tab === 'ranking') {
      void startSelfRankRotation()
    } else {
      stopSelfRankRotation()
    }
  },
  { flush: 'post', immediate: true },
)

onBeforeUnmount(() => {
  stopSelfRankRotation()
})
</script>

<template>
  <div class="share-shell">
    <img class="share-shell__bg" :src="shareAssets.background" alt="" aria-hidden="true" />

    <header class="share-tabs" role="tablist" aria-label="Share menu">
      <button
        v-for="tab in shareTabs"
        :key="tab.key"
        type="button"
        class="share-tabs__button"
        :class="{ 'is-active': activeTab === tab.key }"
        role="tab"
        :aria-selected="activeTab === tab.key"
        @click="selectTab(tab.key)"
      >
        {{ tab.label }}
      </button>
      <span class="share-tabs__indicator" :style="tabIndicatorStyle" aria-hidden="true" />
    </header>

    <main ref="scrollEl" class="share-content" :class="{ 'is-ranking': activeTab === 'ranking' }">
      <section v-if="activeTab === 'share'" class="share-pane" aria-label="Share">
        <div class="share-link-wrap">
          <article class="share-link-card">
            <svg viewBox="0 0 349 90" fill="none" preserveAspectRatio="none" aria-hidden="true">
              <rect x="0.75" y="0.75" width="347" height="88" rx="10" fill="url(#shareLinkFill)" fill-opacity="0.5" />
              <rect x="0.75" y="0.75" width="347" height="88" rx="10" stroke="url(#shareLinkStrokeA)" stroke-width="1.5" vector-effect="non-scaling-stroke" />
              <rect x="0.75" y="0.75" width="347" height="88" rx="10" stroke="url(#shareLinkStrokeB)" stroke-width="1.5" vector-effect="non-scaling-stroke" />
              <defs>
                <linearGradient id="shareLinkFill" x1="0.75" y1="0.75" x2="347.75" y2="0.75" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#021033" />
                  <stop offset="0.485" stop-color="#0c2fa1" />
                  <stop offset="1" stop-color="#000f32" />
                </linearGradient>
                <linearGradient id="shareLinkStrokeA" x1="174.25" y1="88.75" x2="174.25" y2="0.75" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#58ecff" stop-opacity="0" />
                  <stop offset="0.5" stop-color="#479dff" stop-opacity="0.5" />
                  <stop offset="1" stop-color="#58ecff" stop-opacity="0" />
                </linearGradient>
                <linearGradient id="shareLinkStrokeB" x1="5.81" y1="0.75" x2="155.2" y2="117.09" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#479dff" stop-opacity="0.5" />
                  <stop offset="1" stop-color="#58ecff" stop-opacity="0" />
                </linearGradient>
              </defs>
            </svg>
            <div class="share-link-card__body">
              <p>💰 Earn up to 30% commission on this amazing platform!</p>
              <p>⏰ Commissions are settled daily.</p>
              <p>🎯 Join me and start earning now:</p>
              <div class="share-link-card__login">Log in to get your exclusive share link</div>
            </div>
          </article>

          <div class="share-actions">
            <button
              v-for="action in shareActions"
              :key="action.label"
              type="button"
              class="share-actions__button"
              :class="`is-${action.tone}`"
              @click="handleShareAction(action)"
            >
              {{ action.label }}
            </button>
          </div>
        </div>

        <div class="social-links" aria-label="Share channels">
          <button
            v-for="social in socialLinks"
            :key="social.label"
            type="button"
            class="social-links__button"
            :aria-label="social.label"
            @click="openSocialShare(social)"
          >
            <img :src="socialIcon(social)" :alt="social.label" loading="lazy" />
          </button>
        </div>

        <section class="chest-section" aria-label="Invite Chest">
          <header class="chest-section__header">
            <div>
              <img :src="shareAssets.treasure" alt="" aria-hidden="true" />
              <h1>Invite Chest</h1>
            </div>
            <button type="button">Details</button>
          </header>

          <article class="task-card task-card--primary">
            <div class="task-card__gloss" aria-hidden="true" />
            <div class="task-card__badge task-card__badge--wide">
              <img class="task-card__badge-bg" :src="shareAssets.badge" alt="" aria-hidden="true" />
              <strong>{{ formatCurrency(100) }}</strong>
              <span>Chest</span>
            </div>

            <div class="task-card__content">
              <div class="task-card__block task-card__block--narrow">
                <span>How to Get?</span>
                <strong>Invite 1 registrations</strong>
              </div>
              <div class="task-card__block">
                <span>How to Open?</span>
                <p>Invited registrations Deposit ≥ {{ formatCurrency(500) }}</p>
                <p>Invited registrations Valid bets ≥ {{ formatCurrency(3000) }}</p>
              </div>

              <div class="stats-strip">
                <div v-for="stat in chestStats" :key="stat.label" class="stats-strip__item">
                  <span>{{ stat.label }}</span>
                  <strong>
                    {{ stat.value }}
                    <svg v-if="stat.navigable" viewBox="0 0 6 10" aria-hidden="true">
                      <path d="M.22 8 3.58 4.64.22 1.28A.75.75 0 0 1 .75 0c.2 0 .39.08.53.22l3.89 3.89a.75.75 0 0 1 0 1.06L1.28 9.06A.75.75 0 0 1 .75 9.28.75.75 0 0 1 .22 8Z" />
                    </svg>
                  </strong>
                </div>
              </div>

              <div class="calendar-scroll">
                <div class="calendar-scroll__chests">
                  <button v-for="day in calendarDays" :key="`chest-${day}`" type="button" class="calendar-chest">
                    <img :src="shareAssets.smallChestClosed" alt="" loading="lazy" />
                  </button>
                </div>
                <div class="calendar-scroll__line">
                  <span aria-hidden="true" />
                </div>
                <div class="calendar-scroll__days">
                  <span v-for="day in calendarDays" :key="`day-${day}`">{{ day }}</span>
                </div>
              </div>
            </div>
          </article>

          <div class="chest-connector" aria-hidden="true">
            <span />
            <span />
          </div>

          <article class="task-card task-card--pro">
            <div class="task-card__gloss" aria-hidden="true" />
            <div class="task-card__badge task-card__badge--small">
              <img class="task-card__badge-bg" :src="shareAssets.badge" alt="" aria-hidden="true" />
              <span>Chest Pro</span>
            </div>

            <div class="task-card__content">
              <div class="task-card__block task-card__block--narrow">
                <span>How to Get?</span>
                <div class="task-card__inline">
                  <img :src="shareAssets.smallChestClosed" alt="" loading="lazy" />
                  <strong>* 1</strong>
                </div>
                <strong>Get 1 Chests first</strong>
              </div>
              <div class="task-card__block">
                <span>How to Open?</span>
                <div class="task-card__inline">
                  <img :src="shareAssets.smallChestOpen" alt="" loading="lazy" />
                  <strong>* 5</strong>
                </div>
                <p>Open 5 Chests first</p>
              </div>

              <div class="stats-strip">
                <div v-for="stat in chestStats" :key="`pro-${stat.label}`" class="stats-strip__item">
                  <span>{{ stat.label }}</span>
                  <strong>
                    {{ stat.value }}
                    <svg v-if="stat.navigable && stat.label === 'Earned'" viewBox="0 0 6 10" aria-hidden="true">
                      <path d="M.22 8 3.58 4.64.22 1.28A.75.75 0 0 1 .75 0c.2 0 .39.08.53.22l3.89 3.89a.75.75 0 0 1 0 1.06L1.28 9.06A.75.75 0 0 1 .75 9.28.75.75 0 0 1 .22 8Z" />
                    </svg>
                  </strong>
                </div>
              </div>

              <div class="reward-scroll">
                <button v-for="reward in chestRewards" :key="reward" type="button" class="reward-chest">
                  <img :src="shareAssets.largeChestClosed" alt="" loading="lazy" />
                  <span>{{ reward }}</span>
                </button>
              </div>
            </div>
          </article>
        </section>
      </section>

      <section v-else-if="activeTab === 'ranking'" class="ranking-pane" aria-label="Ranking">
        <article class="ranking-date">
          <div class="ranking-date__card">
            <svg
              class="ranking-date__shape"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 347 83"
              fill="none"
              preserveAspectRatio="none"
              aria-hidden="true"
            >
              <path
                d="M165.304 0.5C238.631 0.5 313.95 20.6142 338.549 27.8105C343.323 29.2073 346.5 33.5929 346.5 38.584V71C346.5 77.3513 341.351 82.5 335 82.5H12C5.64873 82.5 0.5 77.3513 0.5 71V38.0283C0.500026 33.2999 3.39209 29.0808 7.84375 27.541C29.2523 20.1359 92.855 0.500039 165.304 0.5Z"
                fill="url(#rankingDateFill)"
                stroke="url(#rankingDateStroke)"
              />
              <defs>
                <linearGradient id="rankingDateFill" x1="173.5" y1="0" x2="173.5" y2="83" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#2B3B74" />
                  <stop offset="1" stop-color="#2B3B74" stop-opacity="0" />
                </linearGradient>
                <linearGradient id="rankingDateStroke" x1="173.5" y1="5.71428" x2="173.5" y2="57.1428" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#416EDE" />
                  <stop offset="1" stop-color="#416EDE" stop-opacity="0" />
                </linearGradient>
              </defs>
            </svg>
            <div class="ranking-date__content">
              <span>Ranking Date</span>
              <strong>{{ rankingDate }}</strong>
            </div>
          </div>
          <div class="ranking-date__ribbon">
            <img :src="shareAssets.ribbon" alt="" aria-hidden="true" />
            <svg viewBox="0 0 347 60" aria-hidden="true">
              <defs>
                <path id="rankingArc" d="M 20,50 Q 173.5,10 327,50" />
              </defs>
              <text>
                <textPath href="#rankingArc" startOffset="50%" text-anchor="middle" textLength="300" lengthAdjust="spacingAndGlyphs">
                  Period's Commission &amp; Rewards
                </textPath>
              </text>
            </svg>
          </div>
        </article>

        <article class="ranking-podium">
          <div class="ranking-podium__badges">
            <div
              v-for="user in rankingTopUsers"
              :key="user.rank"
              class="ranking-top-user"
              :class="`rank-${user.rank}`"
            >
              <div class="ranking-top-user__badge">
                <img class="ranking-top-user__avatar" :src="user.avatar" :alt="user.name" loading="lazy" />
                <img class="ranking-top-user__frame" :src="user.badge" alt="" aria-hidden="true" loading="lazy" />
              </div>
              <strong>{{ user.name }}</strong>
            </div>
          </div>

          <div class="ranking-podium__base" aria-label="Top ranking totals">
            <img class="ranking-podium__stage" :src="shareAssets.rankingStage" alt="" aria-hidden="true" />
            <div class="ranking-podium__totals">
              <div v-for="user in rankingTopUsers" :key="`total-${user.rank}`" :class="`rank-${user.rank}`">
                <strong>{{ user.total }}</strong>
                <span>Total</span>
              </div>
            </div>
          </div>
        </article>

        <div class="ranking-table-label">
          <span>Rank</span>
          <span>|</span>
          <span>User</span>
          <strong>Total</strong>
        </div>

        <article class="ranking-table">
          <div
            v-for="row in rankingRows"
            :key="row.rank"
            class="ranking-row"
            :class="{ 'is-alt': row.rank % 2 === 0 }"
          >
            <span class="ranking-row__rank">{{ row.rank }}</span>
            <div class="ranking-row__user">
              <img :src="row.avatar" :alt="row.name" loading="lazy" />
              <span>{{ row.name }}</span>
            </div>
            <strong>{{ row.total }}</strong>
          </div>
        </article>

        <article class="ranking-self-card">
          <header>
            <span>Your Rank &amp; Earnings:</span>
            <strong>{{ selfRanking.rank }}</strong>
            <span>&amp;</span>
            <strong class="ranking-self-card__money">
              <span>{{ currencySymbol() }}</span>
              <span>{{ selfRanking.earnings }}</span>
            </strong>
          </header>
          <div class="ranking-self-card__body">
            <div class="ranking-self-card__switcher">
              <div ref="selfRankSocialLayer" class="ranking-self-card__layer ranking-self-card__layer--social">
                <button
                  v-for="social in socialLinks"
                  :key="`self-${social.label}`"
                  type="button"
                  class="ranking-self-card__social"
                  :aria-label="social.label"
                  @click="openSocialShare(social)"
                >
                  <img :src="socialIcon(social)" :alt="social.label" loading="lazy" />
                </button>
              </div>
              <div ref="selfRankEntriesLayer" class="ranking-self-card__layer ranking-self-card__layer--entries">
                <div class="ranking-self-card__entries">
                  <button
                    v-for="entry in shareHomeEntries"
                    :key="entry.id"
                    type="button"
                    class="ranking-self-card__entry"
                    :style="{ background: entry.background }"
                    :aria-label="entry.title"
                  >
                    <img :src="entry.image" :alt="entry.title" loading="lazy" />
                  </button>
                </div>
              </div>
            </div>
            <p>Invite friends and earn unlimited rewards</p>
          </div>
        </article>
      </section>

      <section v-else class="agent-pane" aria-label="Agent">
        <section class="agent-hero">
          <ContactMeBadge class="agent-contact" :image="shareAssets.customer" />

          <div class="agent-profile">
            <div class="agent-profile__avatar">
              <svg viewBox="0 0 73 73" aria-hidden="true">
                <circle cx="36.5" cy="36.5" r="30" />
                <circle cx="36.5" cy="36.5" r="31" />
              </svg>
              <img :src="agentProfile.avatarUrl || shareAssets.agentAvatar" :alt="agentProfile.displayName" loading="lazy" />
            </div>
            <div>
              <h1>{{ agentProfile.displayName }}</h1>
              <p><span>ID:</span> {{ agentProfile.userNo }}</p>
              <p class="agent-profile__invite">
                <span>Invitation Code:</span> {{ agentProfile.inviteCode }}
                <button type="button" aria-label="Copy invitation code" @click="copyAgentInviteCode">
                  <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
                    <rect x="6" y="4" width="10" height="11" rx="1.5" stroke="currentColor" stroke-width="1.4" />
                    <path d="M4 7v8.5A1.5 1.5 0 0 0 5.5 17H13" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
                  </svg>
                </button>
              </p>
            </div>
          </div>

          <nav class="agent-tabs" aria-label="Agent sections">
            <button
              v-for="tab in agentTabs"
              :key="tab.key"
              type="button"
              :class="{ 'is-active': activeAgentTab === tab.key }"
              @click="activeAgentTab = tab.key"
            >
              {{ tab.label }}
            </button>
          </nav>
        </section>

        <div class="agent-body">
          <template v-if="activeAgentTab === 'overview'">
            <article class="agent-claim-card">
              <header>
                <span>Claimable Commission</span>
                <button type="button" class="agent-customer-link">
                  <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path d="M4 13v-2a8 8 0 0 1 16 0v2M4 13H2v4h3v-6M20 13h2v4h-3v-6M19 17c0 2-2 3-5 3" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                  Customer Service
                </button>
              </header>
              <strong class="agent-claim-card__amount">{{ formatCurrency(agentSummary.claimableCommission) }}</strong>
              <div class="agent-claim-card__summary">
                <div><span>Total</span><strong>{{ formatCurrency(agentSummary.totalSettlement) }}</strong></div>
                <div><span>Claimed</span><strong>{{ formatCurrency(agentSummary.totalSettlement) }}</strong></div>
              </div>
              <button class="agent-claim-card__button" type="button" :disabled="Number(agentSummary.claimableCommission) <= 0">
                Claim <em>{{ formatCurrency(agentSummary.claimableCommission) }}</em> to Cash Wallet
              </button>
              <small>Start Date of Agency: -</small>
            </article>

            <section class="agent-settlement">
              <div class="agent-section-title">
                <img :src="shareAssets.agentCardTitle" alt="" aria-hidden="true" />
                <span><img :src="shareAssets.agentStar" alt="" aria-hidden="true" /> Total Settlement <img :src="shareAssets.agentStar" alt="" aria-hidden="true" /></span>
              </div>
              <div class="agent-settlement__cards">
                <article>
                  <header><span>Direct</span><svg viewBox="0 0 44 40" fill="none" aria-hidden="true"><path d="M22 4v9M9 20v-7h26v7M8 24h8v9H8zM28 24h8v9h-8zM18 13h8v9h-8z" stroke="currentColor" stroke-width="3" stroke-linejoin="round" /></svg></header>
                  <div><strong>{{ formatAmount(agentSummary.directSettlement) }}</strong><span>Referrals Number</span></div>
                </article>
                <article>
                  <header><span>Indirect</span><svg viewBox="0 0 44 40" fill="none" aria-hidden="true"><path d="M8 10h20a8 8 0 0 1 8 8v1M36 9v10H26M36 30H16a8 8 0 0 1-8-8v-1M8 31V21h10" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" /></svg></header>
                  <div><strong>{{ formatAmount(agentSummary.indirectSettlement) }}</strong><span>Referrals Number</span></div>
                </article>
              </div>
            </section>

            <article class="agent-period-report">
              <header>
                <span>Settlement Period</span>
                <button type="button">{{ agentSummary.settlementPeriod || '--' }}<svg viewBox="0 0 12 8" fill="none" aria-hidden="true"><path d="m2 2 4 4 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg></button>
              </header>
              <div class="agent-period-report__range"><span>Date Range: {{ agentSettlementRange }}</span><em>Unsettled</em></div>
              <div class="agent-chart" aria-label="Commission trend">
                <div class="agent-chart__axis"><span>1</span><span>0.8</span><span>0.6</span><span>0.4</span><span>0.2</span><span>0</span></div>
                <div class="agent-chart__plot"><i v-for="line in 6" :key="line" /><div class="agent-chart__labels"><span>13:00</span><span>14:00</span><span>15:00</span><span>16:00</span></div></div>
              </div>
              <div class="agent-chart__legend"><span><i class="is-direct" />Direct Referrals Contribute Commission</span><span><i class="is-indirect" />Indirect Referrals Contribute Commission</span></div>
              <div class="agent-contribution">
                <div><strong>{{ formatAmount(agentSummary.validBetAmount) }} ›</strong><span>Direct Referrals Valid bets</span></div>
                <em>*{{ formatPercent(agentSummary.directRate) }}</em>
                <div><strong>{{ formatAmount(agentSummary.directCommission) }} ›</strong><span>Direct Referrals Contribute Commission</span></div>
              </div>
              <div class="agent-contribution__stats">
                <div><strong>{{ agentSummary.directReferrals }} ›</strong><span>New Direct Referrals</span></div>
                <div><strong>{{ agentSummary.directReferrals }} ›</strong><span>Active Direct Referrals</span></div>
                <div><strong>{{ agentSummary.totalBettors }} ›</strong><span>New Active Direct Referrals</span></div>
              </div>
            </article>

            <img class="agent-level-chart" :src="shareAssets.agentLevelChart" alt="Referral commission level chart" loading="lazy" />

            <article class="commission-card">
              <div class="commission-card__glow" aria-hidden="true" />
              <header>
                <span>Commission Rules</span>
                <strong>Active referral commission rates</strong>
              </header>
              <div class="commission-table">
                <div class="commission-table__head">
                  <span>Active Direct Referrals Number</span>
                  <span>Direct Referrals Valid bets</span>
                  <span>Combined Commission Rate</span>
                </div>
                <div class="commission-table__divider commission-table__divider--plus" aria-hidden="true">
                  <span>+</span>
                </div>
                <div class="commission-table__divider commission-table__divider--equal" aria-hidden="true">
                  <span>=</span>
                </div>
                <div v-for="row in agentCommissionRules" :key="`${row.referrals}-${row.validBets}`" class="commission-table__row">
                  <span>{{ row.referrals }}</span>
                  <span>{{ row.validBets }}</span>
                  <span>{{ row.commission }}</span>
                </div>
              </div>
            </article>
          </template>

          <template v-else-if="activeAgentTab === 'referrals'">
            <div class="agent-period">
              <span>Settlement Period</span>
              <button type="button">{{ agentSummary.settlementPeriod || '--' }}<svg viewBox="0 0 12 8" fill="none" aria-hidden="true"><path d="m2 2 4 4 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg></button>
            </div>
            <div class="agent-metrics agent-metrics--referrals">
              <article><span>Direct</span><strong>{{ agentSummary.directReferrals }}</strong></article>
              <article><span>Indirect</span><strong>{{ agentSummary.indirectReferrals }}</strong></article>
            </div>
            <div class="agent-referral-tools">
              <label>
                <svg viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="m21 21-4.35-4.35M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
                <input v-model="referralSearch" type="search" placeholder="Referral ID" />
              </label>
              <button type="button">Sort by Registration<svg viewBox="0 0 12 16" fill="none" aria-hidden="true"><path d="m6 1 4 6H2l4-6ZM6 15l-4-6h8l-4 6Z" fill="currentColor" /></svg></button>
            </div>
            <div class="agent-referral-filters">
              <label><input v-model="showDirectReferrals" type="checkbox" /> Only Active Play</label>
              <label><input v-model="showIndirectReferrals" type="checkbox" /> Only New Registration</label>
            </div>
            <div v-if="filteredAgentReferrals.length" class="agent-referral-list">
              <article v-for="referral in filteredAgentReferrals" :key="referral.id">
                <img :src="referral.avatarUrl || shareAssets.agentAvatar" :alt="referral.displayName" />
                <div><strong>{{ referral.displayName }}</strong><span>{{ referral.userNo }}</span></div>
                <time>{{ referral.registeredAt ? new Date(referral.registeredAt).toLocaleDateString() : '--' }}</time>
              </article>
            </div>
            <article v-else class="agent-empty">
              <svg viewBox="0 0 120 92" fill="none" aria-hidden="true">
                <path d="M27 22h52l10 52H18l9-52Z" stroke="currentColor" stroke-width="5" stroke-linejoin="round" />
                <path d="M40 22c0-12 8-18 15-18s15 6 15 18M39 48h32" stroke="currentColor" stroke-width="5" stroke-linecap="round" />
                <circle cx="94" cy="21" r="13" stroke="currentColor" stroke-width="5" /><path d="M94 15v8M94 29v1" stroke="currentColor" stroke-width="4" stroke-linecap="round" />
              </svg>
              <p>Oops! Nothing here for now.</p>
            </article>
          </template>

          <template v-else>
            <div class="agent-period">
              <span>Settlement Period</span>
              <button type="button">{{ agentSummary.settlementPeriod || '--' }}<svg viewBox="0 0 12 8" fill="none" aria-hidden="true"><path d="m2 2 4 4 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg></button>
            </div>
            <div class="agent-metrics">
              <article>
                <span>Direct</span>
                <strong>{{ formatAmount(agentSummary.directCommission) }} <em>({{ formatPercent(agentSummary.directRate) }})</em></strong>
              </article>
              <article>
                <span>Indirect</span>
                <strong>{{ formatAmount(agentSummary.indirectCommission) }} <em>({{ formatPercent(agentSummary.indirectRate) }})</em></strong>
              </article>
            </div>
            <div class="agent-commission-summary">
              <p><span>Total Bettors:</span><strong>{{ agentSummary.totalBettors }}</strong></p>
              <p><span>Total Valid bets:</span><strong>{{ formatAmount(agentSummary.validBetAmount) }}</strong></p>
            </div>
            <div v-if="agentCommissionRecords.length" class="agent-commission-list">
              <article v-for="record in agentCommissionRecords" :key="record.id">
                <div><span>Commission</span><strong>{{ formatCurrency(Number(record.commissionAmount) + Number(record.bonusAmount)) }}</strong></div>
                <div><span>Valid bets</span><strong>{{ formatCurrency(record.validBetAmount) }}</strong></div>
              </article>
            </div>
            <article v-else class="agent-empty">
              <svg viewBox="0 0 120 92" fill="none" aria-hidden="true">
                <path d="M27 22h52l10 52H18l9-52Z" stroke="currentColor" stroke-width="5" stroke-linejoin="round" />
                <path d="M40 22c0-12 8-18 15-18s15 6 15 18M39 48h32" stroke="currentColor" stroke-width="5" stroke-linecap="round" />
                <circle cx="94" cy="21" r="13" stroke="currentColor" stroke-width="5" /><path d="M94 15v8M94 29v1" stroke="currentColor" stroke-width="4" stroke-linecap="round" />
              </svg>
              <p>Oops! Nothing here for now.</p>
            </article>
          </template>
        </div>
      </section>
    </main>

  </div>
</template>

<style scoped lang="scss">
.share-shell {
  --bottom-nav-height: 76px;
  --share-top-offset: max(env(safe-area-inset-top, 0px), 12px);

  position: relative;
  display: flex;
  flex-direction: column;
  height: 100vh;
  height: 100dvh;
  min-height: 100vh;
  min-height: 100dvh;
  max-width: 430px;
  margin: 0 auto;
  overflow: hidden;
  color: #ffffff;
  background: linear-gradient(180deg, #000933 0%, #000933 42%, #000e2b 100%);
}

.share-shell__bg {
  pointer-events: none;
  position: absolute;
  inset: 0 0 auto;
  width: 100%;
  height: auto;
  min-height: 900px;
  object-fit: cover;
  object-position: top center;
}

.share-tabs {
  position: sticky;
  top: 0;
  z-index: 20;
  flex: 0 0 44px;
  height: 44px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(0, 9, 51, 0.82);
  backdrop-filter: blur(8px);
}

.share-tabs__button {
  position: relative;
  z-index: 1;
  width: 33.3333%;
  height: 100%;
  color: #ffffff;
  font-size: 14px;
  font-weight: 400;
  transition:
    color 220ms ease,
    font-weight 220ms ease,
    transform 180ms ease;
  outline: none;

  &.is-active {
    color: #4d81e7;
    font-weight: 600;
  }

  &:focus-visible {
    text-shadow: 0 0 10px rgba(77, 129, 231, 0.75);
  }

  &:active {
    transform: scale(0.96);
  }
}

.share-tabs__indicator {
  position: absolute;
  bottom: 0;
  width: 13.3333%;
  height: 3px;
  border-radius: 2px;
  background: linear-gradient(96deg, #3d8aef 18%, #2f69d1 71%);
  box-shadow: 0 2px 8px rgba(135, 187, 255, 0.9);
  transform: translateX(-50%);
  transition: left 260ms ease;
}

.share-content {
  position: relative;
  z-index: 1;
  flex: 1 1 auto;
  min-height: 0;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: none;
  padding-bottom: calc(var(--bottom-nav-height) + env(safe-area-inset-bottom) + 16px);
}

.share-content.is-ranking {
  padding-bottom: calc(var(--bottom-nav-height) + env(safe-area-inset-bottom) + 144px);
}

.share-pane {
  min-height: 100%;
  padding: 16px 14px 0;
}

.share-link-wrap {
  margin: 0 -14px;
  padding: 0 14px 12px;
}

.share-link-card {
  position: relative;
  min-height: 90px;
  overflow: hidden;
  border-radius: 10px;
  backdrop-filter: blur(4px);

  svg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
  }
}

.share-link-card__body {
  position: relative;
  padding: 14px 14px 12px;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.45;
  overflow-wrap: anywhere;

  p {
    margin: 0;
    color: #f59e0b;
    word-break: break-all;
  }
}

.share-link-card__login {
  margin-top: 4px;
  color: #ffffff;
  word-break: break-word;
}

.share-actions {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  margin-top: 8px;
}

.share-actions__button {
  display: grid;
  min-width: 0;
  height: 32px;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.24);
  border-radius: 6px;
  color: #ffffff;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  transition:
    filter 180ms ease,
    transform 180ms ease;

  &.is-violet {
    background: radial-gradient(366.67% 100% at 50% 0%, rgba(255, 255, 255, 0.3), rgba(255, 255, 255, 0)), #584ff2;
    box-shadow: 0 3px 6px 1px rgba(93, 84, 242, 0.48);
  }

  &.is-blue {
    background: radial-gradient(366.67% 100% at 50% 0%, rgba(255, 255, 255, 0.3), rgba(255, 255, 255, 0)), #2078ff;
    box-shadow: 0 2px 6px -2px rgba(68, 133, 235, 0.48);
  }

  &.is-rose {
    background: radial-gradient(370% 100% at 50% 0%, rgba(255, 255, 255, 0.3), rgba(255, 255, 255, 0)), #de429f;
    box-shadow: 0 2px 6px -2px rgba(255, 98, 65, 0.48);
  }

  &:active {
    filter: brightness(1.08);
    transform: scale(0.96);
  }
}

.social-links {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 14px;
}

.social-links__button {
  width: 40px;
  height: 40px;
  overflow: hidden;
  border-radius: 12px;
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.12);
  transition: transform 180ms ease;

  img {
    width: 100%;
    height: 100%;
  }

  &:active {
    transform: scale(0.93);
  }
}

.chest-section {
  margin: 0 -14px;
  padding: 14px 14px 0;
  border-radius: 16px 16px 0 0;
  background: rgba(210, 226, 249, 0.1);
}

.chest-section__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 24px;
  margin-bottom: 12px;

  div {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  img {
    width: 22px;
    height: 22px;
    object-fit: contain;
  }

  h1 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
    line-height: 1.15;
  }

  button {
    padding: 0;
    font-size: 12px;
    font-weight: 500;
    text-decoration: underline;
    background: linear-gradient(#3d8aef, #2f69d1);
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
}

.task-card {
  position: relative;
  border: 1px solid #e6d263;
  border-radius: 10px;
  padding: 17px 17px 14px;
  color: rgba(255, 255, 255, 0.9);
  background: rgba(0, 34, 78, 0.8);
  box-shadow:
    0 3px 0 #7b5919,
    inset 0 0 4px #001e52;
}

.task-card__gloss {
  pointer-events: none;
  position: absolute;
  inset: 2px 4px auto;
  height: 13px;
  border-radius: 9px 9px 0 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.18), rgba(255, 255, 255, 0.04));
  filter: blur(0.4px);
}

.task-card__badge {
  pointer-events: none;
  position: absolute;
  top: 18px;
  right: -10px;
  z-index: 2;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  width: 91px;
  height: 33px;
  padding: 0 12px 10px 18px;
  color: #f9c56c;
  font-style: italic;
  line-height: 1;
  white-space: nowrap;

  strong,
  span {
    position: relative;
    z-index: 1;
    font-size: 12px;
    font-weight: 700;
  }
}

.task-card__badge-bg {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  object-fit: fill;
}

.task-card__badge--wide {
  width: 138px;
  gap: 3px;
  padding-left: 24px;
  padding-right: 8px;
}

.task-card__badge--small {
  justify-content: center;
  padding-left: 18px;
  padding-right: 9px;
}

.task-card__content {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 14px;
}

.task-card__block {
  display: grid;
  gap: 7px;
  font-size: 14px;
  line-height: 1.25;

  span:first-child {
    color: #97a5ca;
    font-size: 12px;
    font-weight: 500;
  }

  strong {
    font-size: 14px;
    font-weight: 700;
  }

  p {
    margin: 0;
  }
}

.task-card__block--narrow {
  padding-right: 68px;
}

.task-card__inline {
  display: flex;
  align-items: center;
  gap: 8px;

  img {
    width: 24px;
    height: 24px;
    object-fit: contain;
  }
}

.stats-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  height: 50px;
  overflow: hidden;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.4);
}

.stats-strip__item {
  position: relative;
  display: grid;
  place-items: center;
  gap: 4px;
  font-size: 12px;
  color: #97a5ca;

  + .stats-strip__item::before {
    content: "";
    position: absolute;
    top: 10px;
    bottom: 10px;
    left: 0;
    width: 1px;
    background: linear-gradient(180deg, rgba(240, 246, 255, 0), rgba(240, 246, 255, 0.5), rgba(240, 246, 255, 0));
  }

  strong {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: #ffffff;
    font-size: 14px;
    font-weight: 600;
    line-height: 1;
  }

  svg {
    width: 10px;
    height: 10px;
    fill: currentColor;
  }
}

.calendar-scroll {
  margin: 0 -17px -4px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 14px 32px 14px;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.calendar-scroll__chests,
.calendar-scroll__days {
  display: flex;
  width: max-content;
  min-width: 100%;
}

.calendar-chest {
  display: grid;
  flex: 0 0 48px;
  height: 68px;
  place-items: end center;

  + .calendar-chest {
    margin-left: 24px;
  }

  img {
    width: 48px;
    height: 46px;
    object-fit: contain;
    filter: grayscale(1);
    opacity: 0.6;
  }
}

.calendar-scroll__line {
  position: relative;
  width: max-content;
  min-width: 100%;
  height: 12px;

  &::before {
    content: "";
    position: absolute;
    top: 50%;
    right: 0;
    left: 0;
    height: 2px;
    border-radius: 999px;
    background: #2e3754;
    transform: translateY(-50%);
  }

  span {
    position: absolute;
    top: 50%;
    left: 0;
    width: 0;
    height: 2px;
    border-radius: 999px;
    background: #f59e0b;
    transform: translateY(-50%);
  }
}

.calendar-scroll__days {
  span {
    display: grid;
    flex: 0 0 48px;
    min-width: 20px;
    height: 20px;
    place-items: center;
    border-radius: 999px;
    color: #97a5ca;
    background: #2a344b;
    font-size: 12px;
    font-weight: 500;

    + span {
      margin-left: 24px;
    }
  }
}

.chest-connector {
  position: relative;
  z-index: 2;
  display: flex;
  justify-content: space-between;
  width: calc(100% - 48px);
  height: 28px;
  margin: -4px auto 0;

  span {
    width: 8px;
    height: 32px;
    border-radius: 999px;
    background: linear-gradient(90deg, #e6d263, #8e6a26);
    box-shadow: 0 4px 4px rgba(30, 86, 215, 0.26);
  }
}

.task-card--pro {
  margin-top: -1px;
  padding: 14px;
}

.reward-scroll {
  display: flex;
  gap: 16px;
  margin: 0 -14px -6px;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 0 14px 14px;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.reward-chest {
  display: grid;
  flex: 0 0 120px;
  justify-items: center;
  gap: 12px;
  transition: transform 180ms ease;

  img {
    width: 102px;
    height: 93px;
    object-fit: contain;
    filter: grayscale(1) drop-shadow(0 8px 10px rgba(0, 0, 0, 0.32));
    opacity: 0.6;
  }

  span {
    padding: 4px 8px;
    border: 1px solid #46537f;
    border-radius: 6px;
    color: #97a5ca;
    background: #364169;
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
  }

  &:active {
    transform: scale(0.96);
  }
}

.ranking-pane {
  min-height: 100%;
  padding: 0px 14px 0;
}

.ranking-date {
  position: relative;
  padding-top: 54px;
}

.ranking-date__card {
  position: relative;
  height: 88px;
}

.ranking-date__shape {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.ranking-date__content {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 6px;

  span {
    color: #b5babe;
    font-size: 12px;
    font-weight: 500;
    line-height: 1.2;
  }

  strong {
    font-size: 14px;
    font-weight: 700;
    line-height: 1.2;
  }
}

.ranking-date__ribbon {

  position: absolute;
  top: 28px;
  left: 50%;
  width: 90%;
  transform: translateX(-50%);

  img {
    width: 100%;
    aspect-ratio: 1164 / 278;
    object-fit: contain;
  }

  svg {
    position: absolute;
    top: 33%;
    left: 50%;
    width: 75%;
    overflow: visible;
    transform: translate(-50%, -50%);
  }

  text {
    fill: #ffffff;
    font-family: Montserrat, sans-serif;
    font-size: 22px;
    font-style: italic;
    font-weight: 700;
    filter: drop-shadow(0 1px 3px rgba(0, 0, 0, 0.6));
  }
}

.ranking-podium {
  position: relative;
  margin-top: 78px;
}

.ranking-podium__badges {
  pointer-events: none;
  position: absolute;
  right: 0;
  bottom: calc(100% - 25px);
  left: 0;
  z-index: 2;
  display: grid;
  grid-template-columns: 29% 38% 31%;
  align-items: end;
  padding: 0 8px;
}

.ranking-top-user {
  display: grid;
  justify-items: center;

  &.rank-1 {
    transform: translateY(-20px);
  }

  strong {
    max-width: 100%;
    overflow: hidden;
    margin-top: 4px;
    font-size: 12px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
    filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.35));
  }

  &.rank-1 strong {
    color: #ffef74;
  }

  &.rank-2 strong {
    color: #e7d1ff;
  }

  &.rank-3 strong {
    color: #cddef8;
  }
}

.ranking-top-user__badge {
  position: relative;
  width: 72px;
  height: 72px;

  .rank-1 & {
    width: 80px;
    height: 80px;
  }
}

.ranking-top-user__avatar {
  position: absolute;
  top: 50%;
  left: 50%;
  z-index: 0;
  width: 46px;
  height: 46px;
  border-radius: 999px;
  object-fit: cover;
  transform: translate(-50%, -50%);

  .rank-1 & {
    width: 54px;
    height: 54px;
  }
}

.ranking-top-user__frame {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.ranking-podium__base {
  position: relative;
  overflow: hidden;
}

.ranking-podium__stage {
  display: block;
  width: 100%;
  aspect-ratio: 347 / 77;
  object-fit: fill;
}

.ranking-podium__totals {
  position: absolute;
  inset: 0;

  > div {
    position: absolute;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    text-align: center;

    &.rank-2 {
      top: 33.12%;
      left: 0.14%;
      width: 32.85%;
      height: 66.23%;
    }

    &.rank-1 {
      top: 8.44%;
      left: 28.96%;
      width: 42.07%;
      height: 90.91%;
    }

    &.rank-3 {
      top: 33.12%;
      left: 67%;
      width: 32.85%;
      height: 66.23%;
    }
  }
}

.ranking-podium__totals strong {
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.4);
}

.ranking-podium__totals span {
  color: rgba(255, 255, 255, 0.78);
  font-size: 11px;
  font-weight: 500;
}

.ranking-table-label {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 16px;
  color: #737373;
  font-size: 11px;
  font-weight: 500;

  strong {
    margin-left: auto;
    color: #737373;
    font-weight: 500;
  }
}

.ranking-table {
  overflow: hidden;
  margin-top: 8px;
  border: 1px solid #1e2438;
  border-radius: 10px;
}

.ranking-row {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) auto;
  align-items: center;
  gap: 14px;
  height: 52px;
  padding: 0 20px;
  background: #131828;

  &.is-alt {
    background: #1e2438;
  }
}

.ranking-row__rank {
  text-align: center;
  font-size: 16px;
  font-weight: 500;
}

.ranking-row__user {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;

  img {
    width: 32px;
    height: 32px;
    flex: 0 0 auto;
    border-radius: 999px;
    object-fit: cover;
    background: rgba(255, 255, 255, 0.2);
  }

  span {
    overflow: hidden;
    font-size: 12px;
    font-weight: 400;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.ranking-row strong {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
}

.ranking-self-card {
  position: fixed;
  bottom: calc(var(--bottom-nav-height) + env(safe-area-inset-bottom));
  left: 50%;
  z-index: 25;
  width: 100%;
  max-width: 402px;
  overflow: hidden;
  margin: 0;
  border-radius: 12px 12px 0 0;
  background: linear-gradient(90deg, #46537f 0%, #667dcc 100%);
  box-shadow: 0 4px 6px rgba(1, 9, 26, 0.24);
  transform: translateX(-50%);
}

.ranking-self-card header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: 30px;
  padding: 6px 14px;
  background: rgba(0, 0, 0, 0.12);
  font-size: 12px;
  line-height: 1;

  span {
    font-weight: 500;
  }

  strong {
    font-size: 14px;
    font-weight: 600;
  }
}

.ranking-self-card__money {
  display: inline-flex;
  align-items: center;
  gap: 2px;

  span {
    background: linear-gradient(#fbf9b9 32.76%, #e89d2c 82.35%);
    background-clip: text;
    -webkit-text-fill-color: transparent;
  }
}

.ranking-self-card__body {
  display: grid;
  gap: 8px;
  padding: 12px 14px;
}

.ranking-self-card__switcher {
  position: relative;
  height: 42px;
  overflow: hidden;
}

.ranking-self-card__layer {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
}

.ranking-self-card__layer--social {
  justify-content: space-between;
  opacity: 0;
  pointer-events: none;
  transform: translateX(-16px);
}

.ranking-self-card__layer--entries {
  opacity: 1;
  pointer-events: auto;
  transform: translateX(0);
}

.ranking-self-card__social {
  width: 42px;
  height: 42px;
  overflow: hidden;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(3, 12, 38, 0.18);
  transition: transform 180ms ease;

  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  &:active {
    transform: scale(0.94);
  }
}

.ranking-self-card__entries {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  width: 100%;
  height: 42px;
}

.ranking-self-card__entry {
  min-width: 0;
  overflow: hidden;
  border-radius: 8px;
  transition: transform 180ms ease;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  &:active {
    transform: scale(0.97);
  }
}

.ranking-self-card p {
  margin: 0;
  text-align: center;
  font-size: 14px;
  font-style: italic;
  font-weight: 700;
  line-height: 1.2;
  background: linear-gradient(90deg, #e89d2c 0%, #fbf9b9 39.9%, #fbf9b9 60.1%, #e89d2c 100%);
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.agent-pane {
  min-height: 100%;
}

.agent-hero {
  position: relative;
  padding: 12px 14px 0;
  background-image: v-bind(agentHeroBackground);
  background-repeat: no-repeat;
  background-position: center top;
  background-size: cover;
}

.agent-contact {
  position: absolute;
  top: 12px;
  right: 12px;
  display: grid;
  justify-items: center;
  padding: 0;

  img {
    width: 44px;
    height: 55px;
    object-fit: contain;
  }

  span {
    margin-top: -6px;
    padding: 2px 10px;
    border: 1px solid #00a1d5;
    border-radius: 999px;
    color: #003278;
    background: linear-gradient(#baeaff, #38cfff);
    font-size: 10px;
    font-weight: 600;
    white-space: nowrap;
  }
}

.agent-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-right: 72px;
}

.agent-profile__avatar {
  position: relative;
  width: 73px;
  height: 73px;
  flex: 0 0 auto;

  svg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    fill: none;

    circle:first-child {
      stroke: #dcfdfe;
      stroke-width: 1;
      filter: drop-shadow(0 0 2px rgba(117, 241, 255, 0.6));
    }

    circle:last-child {
      stroke: rgba(220, 253, 254, 0.3);
      stroke-width: 1;
      filter: drop-shadow(0 0 3px rgba(117, 241, 255, 0.3));
    }
  }

  img {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 54px;
    height: 54px;
    border-radius: 999px;
    object-fit: cover;
    transform: translate(-50%, -50%);
  }
}

.agent-profile h1 {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.2;
}

.agent-profile p {
  margin: 4px 0 0;
  font-size: 12px;
  font-weight: 300;
  line-height: 1.2;

  span {
    color: rgba(255, 255, 255, 0.8);
  }
}

.agent-tabs {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  height: 36px;
  margin: 20px -14px 0;
  border-radius: 10px 10px 0 0;
  background: rgba(53, 62, 96, 0.4);
}

.agent-tabs__indicator {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 33.3333%;
  border-radius: 10px 10px 0 0;
  background: linear-gradient(0deg, #3b4873, #4c61a9 45.56%, #526ab9);
  transition: left 220ms ease;
}

.agent-tabs button {
  position: relative;
  z-index: 1;
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  font-style: italic;
  font-weight: 400;
  outline: none;

  &.is-active {
    color: #ffffff;
    font-weight: 700;
    filter: drop-shadow(0 0 5px rgba(255, 255, 255, 0.3));
  }

  &:focus-visible {
    text-shadow: 0 0 10px rgba(255, 255, 255, 0.45);
  }
}

.agent-body {
  display: grid;
  gap: 14px;
  padding: 14px 14px 0;
  background: linear-gradient(180deg, #3b4873 0%, rgba(59, 73, 116, 0) 64px);
}

.agent-claim-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 4px 12px;
  min-height: 82px;
  border: 1px solid rgba(113, 174, 255, 0.5);
  border-radius: 12px;
  padding: 13px 14px;
  background: linear-gradient(135deg, #163ba1, #14246f 58%, #071a55);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.16), 0 8px 18px rgba(0, 0, 0, 0.2);

  span {
    color: rgba(255, 255, 255, 0.72);
    font-size: 12px;
  }

  strong {
    grid-column: 1;
    color: #ffe778;
    font-size: 23px;
    line-height: 1.15;
  }

  button {
    grid-row: 1 / 3;
    grid-column: 2;
    min-width: 78px;
    height: 34px;
    border: 1px solid #fff0a3;
    border-radius: 999px;
    color: #712a00;
    background: linear-gradient(180deg, #ffef87, #ffac20);
    font-size: 13px;
    font-weight: 700;
    box-shadow: 0 3px 0 #b24d0b;
  }

  button:disabled {
    filter: grayscale(0.45);
    opacity: 0.55;
  }
}

.agent-settlement {
  display: grid;
  gap: 9px;

  > header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 2px;
    font-size: 13px;

    strong {
      color: #ffe778;
      font-size: 15px;
    }
  }
}

.agent-settlement__cards,
.agent-metrics {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.agent-settlement__cards article,
.agent-metrics article {
  display: grid;
  min-width: 0;
  min-height: 72px;
  place-items: center;
  align-content: center;
  gap: 7px;
  border: 1px solid rgba(108, 150, 255, 0.36);
  border-radius: 10px;
  padding: 10px;
  background: linear-gradient(180deg, rgba(44, 68, 144, 0.96), rgba(24, 37, 92, 0.98));

  span {
    color: rgba(255, 255, 255, 0.66);
    font-size: 11px;
    text-align: center;
  }

  strong {
    font-size: 16px;
  }

  em {
    color: #71e798;
    font-size: 10px;
    font-style: normal;
  }
}

.agent-period {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 42px;
  border-radius: 8px;
  padding: 0 12px;
  background: rgba(21, 31, 66, 0.88);

  span {
    color: rgba(255, 255, 255, 0.64);
    font-size: 12px;
  }

  strong {
    color: #ffffff;
    font-size: 13px;
  }
}

.agent-referral-tools {
  display: flex;
  gap: 8px;

  label {
    position: relative;
    flex: 1;
  }

  svg {
    position: absolute;
    top: 50%;
    left: 11px;
    width: 16px;
    height: 16px;
    color: rgba(255, 255, 255, 0.45);
    transform: translateY(-50%);
  }

  input,
  button {
    height: 38px;
    border: 0;
    border-radius: 8px;
    color: #ffffff;
    background: #1e2947;
    font-size: 12px;
  }

  input {
    width: 100%;
    padding: 0 10px 0 36px;
    outline: none;
  }

  button {
    width: 68px;
  }
}

.agent-referral-filters {
  display: flex;
  gap: 18px;
  color: rgba(255, 255, 255, 0.76);
  font-size: 12px;

  label {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  input {
    width: 15px;
    height: 15px;
    accent-color: #3d8aef;
  }
}

.agent-referral-list,
.agent-commission-list {
  display: grid;
  gap: 8px;
}

.agent-referral-list article {
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  min-height: 58px;
  border-radius: 9px;
  padding: 8px 10px;
  background: #1e2947;

  img {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    object-fit: cover;
  }

  div {
    display: grid;
    min-width: 0;
    gap: 3px;
  }

  strong {
    overflow: hidden;
    font-size: 12px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span,
  time {
    color: rgba(255, 255, 255, 0.55);
    font-size: 10px;
  }
}

.agent-metrics--secondary article {
  min-height: 62px;
}

.agent-commission-list article {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  border-radius: 9px;
  padding: 11px;
  background: #1e2947;

  div {
    display: grid;
    gap: 5px;
  }

  span {
    color: rgba(255, 255, 255, 0.58);
    font-size: 10px;
  }

  strong {
    font-size: 13px;
  }
}

.agent-rule-card,
.commission-card,
.agent-empty {
  position: relative;
  overflow: hidden;
  border-radius: 12px;
  background: linear-gradient(#01005a 0%, #01005a 40%, #163a8a 100%);
}

.agent-rule-card__glow,
.commission-card__glow {
  position: absolute;
  top: -10px;
  left: 50%;
  width: 286px;
  height: 64px;
  background: #1144c2;
  clip-path: polygon(0 0, 12% 100%, 88% 100%, 100% 0);
  filter: blur(70px);
  transform: translateX(-50%);
}

.agent-rule-card img {
  position: relative;
  z-index: 1;
  width: 100%;
  object-fit: contain;
}

.commission-card {
  --f1: 225 34% 12%;
  --f3: 226 29% 39%;

  padding: 16px 10px;
}

.commission-card header {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 4px;
  margin-bottom: 10px;

  span {
    color: rgba(255, 255, 255, 0.7);
    font-size: 14px;
  }

  strong {
    font-size: 12px;
    font-weight: 700;
  }
}

.commission-table {
  position: relative;
  z-index: 1;
  display: grid;
  border: 1px solid hsl(var(--f3));
  border-bottom: 0;
}

.commission-table__head,
.commission-table__row {
  display: grid;
  grid-template-columns: 1.08fr 1fr 1fr;

  span {
    display: grid;
    min-height: 40px;
    place-items: center;
    border-right: 1px solid hsl(var(--f3));
    border-bottom: 1px solid hsl(var(--f3));
    padding: 0 6px;
    text-align: center;

    &:last-child {
      border-right: 0;
    }
  }
}

.commission-table__head {
  background: hsl(var(--f3));

  span {
    color: rgba(255, 255, 255, 0.8);
    font-size: 10px;
    font-weight: 400;
    line-height: 1.35;
  }
}

.commission-table__divider {
  position: absolute;
  top: 12px;
  z-index: 2;
  display: grid;
  width: 14px;
  height: 14px;
  place-items: center;
  border: 1px solid hsl(var(--f3));
  border-radius: 999px;
  background: hsl(var(--f1));
  transform: translateX(-50%);

  span {
    color: rgba(255, 255, 255, 0.8);
    font-size: 8px;
    font-weight: 500;
    line-height: 1;
  }
}

.commission-table__divider--plus {
  left: 33.33%;
}

.commission-table__divider--equal {
  left: 66.66%;
}

.commission-table__row span {
  color: #ffffff;
  font-size: 12px;
  font-weight: 500;
}

.agent-empty {
  display: grid;
  min-height: 180px;
  place-items: center;
  align-content: center;
  gap: 8px;
  padding: 24px;
  text-align: center;

  svg {
    width: 108px;
    height: 84px;
    color: #8496c3;
  }

  h2,
  p {
    position: relative;
    z-index: 1;
    margin: 0;
  }

  h2 {
    font-size: 18px;
    font-weight: 700;
  }

  p {
    color: #97a5ca;
    font-size: 15px;
    font-weight: 600;
  }
}

/* The source client keeps the Agent surface intentionally dense and game-like. */
.agent-hero {
  padding: 18px 14px 0;
  background-size: 100% auto;
}

.agent-profile {
  min-height: 94px;
  gap: 14px;
  padding-right: 82px;
}

.agent-profile h1 {
  margin-bottom: 12px;
  font-size: 18px;
  font-weight: 700;
}

.agent-profile p {
  margin-top: 5px;
  font-size: 13px;
}

.agent-profile__invite {
  display: flex;
  align-items: center;
  white-space: nowrap;
}

.agent-profile__invite button {
  display: inline-flex;
  width: 22px;
  height: 22px;
  align-items: center;
  justify-content: center;
  margin-left: 4px;
  border: 0;
  color: rgba(255, 255, 255, 0.82);
  background: transparent;
}

.agent-profile__invite svg {
  width: 18px;
  height: 18px;
}

.agent-tabs {
  height: 42px;
  margin-top: 18px;
  overflow: visible;
  border-radius: 12px 12px 0 0;
  background: rgba(63, 78, 130, 0.52);
}

.agent-tabs button {
  z-index: 1;
  border-radius: 12px 12px 0 0;
  font-size: 16px;
  transition:
    color 220ms ease,
    background 220ms ease,
    transform 180ms ease;
}

.agent-tabs button.is-active {
  background: linear-gradient(180deg, #5d78ce 0%, #475a9d 82%, #43558d 100%);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.12);
}

.agent-tabs button:active {
  transform: scale(0.98);
}

.agent-body {
  gap: 16px;
  min-height: 650px;
  padding: 16px 14px 22px;
  background: linear-gradient(180deg, rgba(66, 82, 137, 0.92) 0, rgba(0, 11, 53, 0.18) 120px, rgba(0, 9, 51, 0) 360px);
}

.agent-claim-card {
  display: block;
  min-height: 0;
  border: 1px solid #dbe7ff;
  border-radius: 14px;
  padding: 18px 16px 16px;
  background: linear-gradient(180deg, #05006d 0%, #162780 48%, #1e4399 100%);
  box-shadow: inset 0 0 18px rgba(53, 105, 255, 0.22), 0 6px 14px rgba(0, 0, 0, 0.22);
}

.agent-claim-card > header {
  display: flex;
  min-height: 24px;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.agent-claim-card > header > span {
  color: #59d496;
  font-size: 15px;
}

.agent-claim-card .agent-customer-link {
  display: flex;
  align-items: center;
  gap: 5px;
  border: 0;
  color: #97a8db;
  background: transparent;
  font-size: 12px;
  font-weight: 400;
  text-decoration: underline;
  box-shadow: none;
}

.agent-customer-link svg {
  width: 18px;
  height: 18px;
}

.agent-claim-card__amount {
  display: block;
  margin-top: 4px;
  color: #50d08d;
  font-size: 25px;
  line-height: 1.25;
}

.agent-claim-card__summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 9px;
  overflow: hidden;
  border-radius: 11px;
  background: rgba(4, 10, 83, 0.58);
}

.agent-claim-card__summary div {
  position: relative;
  display: grid;
  min-height: 76px;
  place-items: center;
  align-content: center;
  gap: 8px;
}

.agent-claim-card__summary div + div::before {
  position: absolute;
  top: 14px;
  bottom: 14px;
  left: 0;
  width: 1px;
  background: rgba(151, 168, 219, 0.36);
  content: '';
}

.agent-claim-card__summary span {
  color: rgba(255, 255, 255, 0.9);
  font-size: 15px;
}

.agent-claim-card__summary strong {
  color: #ffffff;
  font-size: 17px;
}

.agent-claim-card__button {
  width: 100%;
  height: 36px;
  margin-top: 10px;
  border: 0;
  border-radius: 999px;
  color: #ffffff;
  background: #4a9aa8;
  font-size: 13px;
}

.agent-claim-card__button em {
  font-weight: 700;
}

.agent-claim-card__button:disabled {
  color: #aeb5cf;
  opacity: 0.72;
}

.agent-claim-card small {
  display: block;
  margin-top: 10px;
  color: #9aa9d2;
  font-size: 11px;
  text-align: center;
}

.agent-settlement {
  position: relative;
  overflow: hidden;
  border-radius: 14px;
  padding: 48px 14px 14px;
  background: linear-gradient(180deg, #060070 0%, #11298d 52%, #1d459c 100%);
}

.agent-section-title {
  position: absolute;
  top: 0;
  left: 50%;
  width: 72%;
  height: 44px;
  transform: translateX(-50%);
}

.agent-section-title > img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: fill;
}

.agent-section-title > span {
  position: relative;
  z-index: 1;
  display: flex;
  height: 100%;
  align-items: center;
  justify-content: center;
  gap: 9px;
  color: #fff420;
  font-size: 17px;
  font-weight: 700;
}

.agent-section-title > span img {
  width: 18px;
  height: 18px;
}

.agent-settlement__cards {
  gap: 10px;
}

.agent-settlement__cards article {
  display: block;
  min-height: 128px;
  overflow: hidden;
  border: 0;
  border-radius: 11px;
  padding: 0;
  background: linear-gradient(180deg, #33479f 0%, #273374 100%);
}

.agent-settlement__cards header {
  display: flex;
  height: 54px;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  color: #a1b2e9;
}

.agent-settlement__cards header span {
  font-size: 15px;
}

.agent-settlement__cards header svg {
  width: 38px;
  height: 38px;
}

.agent-settlement__cards article > div {
  display: grid;
  min-height: 74px;
  place-items: center;
  align-content: center;
  gap: 5px;
  background: rgba(255, 255, 255, 0.06);
}

.agent-settlement__cards article > div strong {
  color: #ffffff;
  font-size: 20px;
}

.agent-settlement__cards article > div span {
  color: #aeb8dd;
  font-size: 12px;
}

.agent-period-report {
  border-radius: 14px;
  padding: 16px 14px;
  background: linear-gradient(180deg, #06006e 0%, #172f91 100%);
}

.agent-period-report > header,
.agent-period {
  display: flex;
  min-height: 42px;
  align-items: center;
  justify-content: space-between;
}

.agent-period-report > header > span {
  font-size: 18px;
  font-weight: 700;
}

.agent-period-report button,
.agent-period button {
  display: flex;
  min-width: 132px;
  height: 34px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border: 1px solid rgba(116, 141, 220, 0.45);
  border-radius: 999px;
  padding: 0 12px;
  color: #ffffff;
  background: #29365f;
  font-size: 13px;
}

.agent-period-report button svg,
.agent-period button svg {
  width: 12px;
  height: 8px;
  color: #9ca8c9;
}

.agent-period-report__range {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 4px;
  color: #95a5d4;
  font-size: 11px;
}

.agent-period-report__range em {
  flex: 0 0 auto;
  border-radius: 5px;
  padding: 4px 6px;
  color: #ffffff;
  background: #ff4b55;
  font-size: 10px;
  font-style: normal;
}

.agent-chart {
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr);
  height: 244px;
  margin-top: 12px;
}

.agent-chart__axis {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: space-between;
  padding: 0 7px 19px 0;
  color: #a8b7e9;
  font-size: 9px;
}

.agent-chart__plot {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding-bottom: 19px;
}

.agent-chart__plot > i {
  width: 100%;
  border-top: 1px dashed rgba(120, 139, 217, 0.4);
}

.agent-chart__labels {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  display: flex;
  justify-content: space-around;
  color: #9cabda;
  font-size: 9px;
}

.agent-chart__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 14px;
  border-bottom: 1px solid rgba(129, 151, 218, 0.25);
  padding: 10px 0 14px 62px;
  color: #d5dcf5;
  font-size: 10px;
}

.agent-chart__legend span {
  display: flex;
  align-items: center;
  gap: 7px;
}

.agent-chart__legend i {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.agent-chart__legend .is-direct {
  background: #5c9cff;
}

.agent-chart__legend .is-indirect {
  background: #84ecf4;
}

.agent-contribution {
  position: relative;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
  margin-top: 12px;
}

.agent-contribution > div,
.agent-contribution__stats > div {
  display: grid;
  min-height: 72px;
  align-content: center;
  gap: 8px;
  border-radius: 8px;
  padding: 10px;
  background: linear-gradient(180deg, #354aa2, #2e3d84);
}

.agent-contribution strong,
.agent-contribution__stats strong {
  color: #70b7ff;
  font-size: 17px;
  text-shadow: 0 1px #0a174e;
}

.agent-contribution span,
.agent-contribution__stats span {
  font-size: 11px;
  line-height: 1.35;
}

.agent-contribution > em {
  position: absolute;
  top: 29px;
  left: 50%;
  z-index: 2;
  width: 68px;
  padding: 5px 8px;
  color: #ffffff;
  background: #0cadd0;
  clip-path: polygon(10% 0, 90% 0, 100% 50%, 90% 100%, 10% 100%, 0 50%);
  font-size: 11px;
  font-style: normal;
  font-weight: 700;
  text-align: center;
  transform: translateX(-50%);
}

.agent-contribution__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  margin-top: 6px;
}

.agent-contribution__stats > div {
  min-height: 74px;
  text-align: center;
}

.agent-level-chart {
  display: block;
  width: 100%;
  border-radius: 14px;
}

.agent-period {
  min-height: 48px;
  border-radius: 8px;
  padding: 0 14px;
  background: #303b60;
}

.agent-period > span {
  font-size: 15px;
}

.agent-period button {
  border: 0;
  border-left: 1px solid rgba(159, 176, 226, 0.35);
  border-radius: 0;
  background: transparent;
}

.agent-metrics {
  gap: 0;
  overflow: hidden;
  border-radius: 8px;
  background: #27345e;
}

.agent-metrics article {
  min-height: 50px;
  border: 0;
  border-radius: 0;
  padding: 6px 10px;
  background: transparent;
}

.agent-metrics article:first-child {
  background: linear-gradient(180deg, #5d73c1, #4a5aa0);
}

.agent-metrics span {
  color: #bdc6e1;
  font-size: 13px;
}

.agent-metrics strong {
  color: #dbe1f5;
  font-size: 14px;
  text-shadow: 0 1px #132152;
}

.agent-metrics strong em {
  color: inherit;
  font-size: 10px;
}

.agent-referral-tools {
  gap: 12px;
}

.agent-referral-tools label,
.agent-referral-tools > button {
  flex: 1;
}

.agent-referral-tools input,
.agent-referral-tools button {
  height: 38px;
  border-radius: 7px;
  background: #323c5f;
}

.agent-referral-tools > button {
  display: flex;
  width: auto;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.agent-referral-tools > button svg {
  position: static;
  width: 11px;
  height: 15px;
  flex: 0 0 auto;
  transform: none;
}

.agent-referral-filters {
  justify-content: space-between;
  gap: 10px;
  color: #a9b5d7;
  font-size: 12px;
}

.agent-referral-filters input {
  appearance: none;
  width: 16px;
  height: 16px;
  border: 1px solid #7f90c0;
  border-radius: 1px;
  background: transparent;
}

.agent-referral-filters input:checked {
  border-color: #4d8ef2;
  background: #4d8ef2;
  box-shadow: inset 0 0 0 3px #15214c;
}

.agent-commission-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.agent-commission-summary p {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  color: #aab6d7;
  font-size: 12px;
}

.agent-commission-summary strong {
  color: #ffffff;
  font-weight: 400;
}

.agent-empty {
  min-height: 240px;
  border-radius: 0;
  padding-top: 18px;
  background: transparent;
}

.agent-empty svg {
  width: 100px;
  height: 78px;
  color: #94a5d3;
}

.agent-empty p {
  color: #a1add1;
  font-size: 17px;
}

.commission-card {
  border-radius: 14px;
}

@media (max-width: 374px) {
  .share-pane,
  .ranking-pane {
    padding-right: 12px;
    padding-left: 12px;
  }

  .social-links__button {
    width: 36px;
    height: 36px;
    border-radius: 10px;
  }

  .ranking-row {
    gap: 10px;
    padding: 0 14px;
  }

  .ranking-self-card {
    width: calc(100% - 24px);
  }
}
</style>
