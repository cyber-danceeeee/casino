<script lang="ts">
const homeBootState = {
  catalogDone: false,
  preloadDone: false,
  preloadProgress: 0,
  preloadPromise: null as Promise<void> | null,
  splashCompleted: false,
  initialGuideCompleted: false,
}
</script>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import ActivityGuideFlow from '../components/ActivityGuideFlow.vue'
import ActivityCards from '../components/ActivityCards.vue'
import AppHeader from '../components/AppHeader.vue'
import BannerSwiper from '../components/BannerSwiper.vue'
import FooterInfo from '../components/FooterInfo.vue'
import GameCategoryTabs from '../components/GameCategoryTabs.vue'
import GameSections from '../components/GameSections.vue'
import HomeFloatingActions from '../components/HomeFloatingActions.vue'
import HomeInviteChestDialog from '../components/HomeInviteChestDialog.vue'
import NotificationPermissionPrompt from '../components/NotificationPermissionPrompt.vue'
import OperationEntries from '../components/OperationEntries.vue'
import PendingRewardsDialog from '../components/PendingRewardsDialog.vue'
import RebateJackpot from '../components/RebateJackpot.vue'
import RegisterRewardDialog from '../components/RegisterRewardDialog.vue'
import TopDownloadBanner from '../components/TopDownloadBanner.vue'
import WelcomeSplash from '../components/WelcomeSplash.vue'
import {
  banners as mockBanners,
  type CategoryKey,
  type GameCategory,
  type GameItem,
  type GameSection,
} from '../mock/games'
import { offerEvents as mockOfferEvents } from '../mock/offers'
import { criticalPreloadAssets, runtimeDeferredAssetURLs, runtimePreloadAssets } from '../mock/preloadResources'
import {
  claimPendingRegisterReward,
  claimPendingActivityReward,
  clearRegisterReward,
  isLoggedIn,
  loadPendingRegisterReward,
  pendingRegisterReward,
  pendingActivityRewards,
  refreshMockUser,
  showLoginInviteChest,
} from '../mock/session'
import { h5Api, type H5OfferEventPayload } from '../services/h5Api'
import {
  isActivityUnread,
  loadCachedOffers,
  markActivityRead,
  onActivityReadChange,
  publishActivityUnreadState,
} from '../services/activityReadState'
import { prefetchAssetURLs } from '../services/assetCache'
import { getCachedHomeCatalog, loadHomeCatalog, type HomeCatalogSnapshot } from '../services/homeCatalogCache'
import { loadMusicSessionWithin } from '../services/musicSession'
import { shouldShowNotificationPrompt } from '../services/notificationPermission'
import { preloadAssets } from '../utils/preloadAssets'
import { primeGameAudio } from '../utils/gameAudio'
import { isAndroidAppRuntime } from '../utils/runtimeEnvironment'

const cachedHomeCatalog = getCachedHomeCatalog()
if (cachedHomeCatalog) {
  homeBootState.catalogDone = true
}

const activeCategory = ref<CategoryKey>('hot')
const router = useRouter()
const androidAppRuntime = isAndroidAppRuntime()
const homeBanners = ref(cachedHomeCatalog?.home.banners.length ? cachedHomeCatalog.home.banners : mockBanners)
const homeCategories = ref<GameCategory[]>(cachedHomeCatalog?.home.categories ?? [])
const homeGameSections = ref<GameSection[]>(cachedHomeCatalog?.gameSections ?? [])
const homeProviders = ref(cachedHomeCatalog?.home.providers ?? [])
const homeSocialLinks = ref(cachedHomeCatalog?.home.socialLinks ?? [])
const splashVisible = ref(!homeBootState.splashCompleted)
const splashReady = ref(homeBootState.preloadDone && homeBootState.catalogDone)
const splashProgress = ref(homeBootState.preloadDone ? 100 : homeBootState.preloadProgress)
const guideVisible = ref(false)
const notificationPromptVisible = ref(false)
const rewardsDialogVisible = ref(false)
const homeOfferEvents = ref<H5OfferEventPayload[]>([])
const unreadActivitySlugs = ref<string[]>([])
const guideInitial = ref<'first-deposit' | 'register-gift' | 'app-download'>('first-deposit')
const downloadBannerVisible = ref(!androidAppRuntime)
let scrollFrame = 0
let categoryScrollTimer = 0
let initialFlowTimer = 0
let requestedCategory: CategoryKey | null = null
let componentActive = true
let pendingRewardPromise: Promise<void> | null = null
let stopActivityReadListener: () => void = () => undefined
const homePageRequests = new Map<string, Promise<GameSection>>()

if (cachedHomeCatalog?.home.categories.length) {
  activeCategory.value = cachedHomeCatalog.home.categories[0].key
}

const updateSplashReady = () => {
  splashReady.value = homeBootState.preloadDone && homeBootState.catalogDone
}

const launchGame = (game: GameItem) => {
  if (game.status === 'maintenance') return

  if (!isLoggedIn.value) {
    void router.push({ name: 'auth-login', query: { redirect: '/' } })
    return
  }

  primeGameAudio()

  void router.push({
    name: 'game-play',
    query: {
      ...(game.gameDbId ? { gameId: String(game.gameDbId) } : {}),
      providerCode: game.providerCode,
      providerGameCode: game.providerGameCode,
      gameName: game.name,
      favorite: game.favorite ? '1' : '0',
    },
  })
}

const loadHomePendingReward = () => {
  if (!isLoggedIn.value) {
    clearRegisterReward()
    return Promise.resolve()
  }

  pendingRewardPromise = loadPendingRegisterReward()
    .then(() => undefined)
    .catch((error) => {
      console.warn('[api] failed to load pending register reward.', error)
      clearRegisterReward()
    })
  return pendingRewardPromise
}

const refreshHomeActivityUnread = () => {
  unreadActivitySlugs.value = homeOfferEvents.value
    .filter(isActivityUnread)
    .map((event) => event.slug)
    .filter(Boolean)
  publishActivityUnreadState(unreadActivitySlugs.value.length > 0)
}

const loadHomeActivityUnread = async () => {
  try {
    const offers = await loadCachedOffers()
    homeOfferEvents.value = offers.events || []
    refreshHomeActivityUnread()
  } catch (error) {
    console.warn('[api] failed to load activity unread state.', error)
    homeOfferEvents.value = mockOfferEvents.map((event, index) => ({
      id: index + 1,
      activityCode: event.id,
      slug: event.detailSlug || event.id,
      title: event.title,
      cardImageURL: event.image,
      updatedAt: 'bundled-v1',
    }))
    refreshHomeActivityUnread()
  }
}

const markHomeActivityRead = (slug: string) => {
  const event = homeOfferEvents.value.find((item) => item.slug === slug)
  if (event) markActivityRead(event)
}

const scheduleInitialGuide = () => {
  if (initialFlowTimer) window.clearTimeout(initialFlowTimer)
  guideInitial.value = 'first-deposit'
  initialFlowTimer = window.setTimeout(() => {
    initialFlowTimer = 0
    if (!componentActive) return
    guideVisible.value = true
  }, 160)
}

const resumeInitialHomeFlow = () => {
  if (splashVisible.value || pendingRegisterReward.value || notificationPromptVisible.value) return

  if (shouldShowNotificationPrompt()) {
    notificationPromptVisible.value = true
    return
  }

  if (homeBootState.initialGuideCompleted || showLoginInviteChest.value) return
  scheduleInitialGuide()
}

const onSplashComplete = async () => {
  homeBootState.splashCompleted = true
  splashVisible.value = false

  if (pendingRewardPromise) {
    await pendingRewardPromise
  }

  resumeInitialHomeFlow()
}

const onGuideComplete = () => {
  homeBootState.initialGuideCompleted = true
  guideVisible.value = false
}

const openDownloadGuide = () => {
  if (androidAppRuntime) return
  guideInitial.value = 'app-download'
  guideVisible.value = true
}

const getSectionMarkerY = () => {
  const tabs = document.querySelector<HTMLElement>('.category-tabs')
  const tabsRect = tabs?.getBoundingClientRect()
  const stickyTop = tabs ? Number.parseFloat(window.getComputedStyle(tabs).top) : Number.NaN

  if (tabsRect && Number.isFinite(stickyTop)) {
    return stickyTop + tabsRect.height + 14
  }

  return 132
}

const updateActiveCategoryFromScroll = () => {
  scrollFrame = 0

  const markerY = getSectionMarkerY()
  if (requestedCategory) {
    activeCategory.value = requestedCategory
    return
  }

  let nextCategory = homeGameSections.value[0]?.key ?? activeCategory.value
  let nearestCategory = nextCategory
  let nearestDistance = Number.POSITIVE_INFINITY

  for (const section of homeGameSections.value) {
    const element = document.getElementById(`section-${section.key}`)
    if (!element) continue

    const rect = element.getBoundingClientRect()
    const distance = Math.abs(rect.top - markerY)

    if (distance < nearestDistance) {
      nearestDistance = distance
      nearestCategory = section.key
    }

    if (rect.top <= markerY) {
      nextCategory = section.key
    }
  }

  const documentHeight = document.documentElement.scrollHeight
  if (window.scrollY + window.innerHeight >= documentHeight - 2) {
    nextCategory = homeGameSections.value.at(-1)?.key ?? nextCategory
  }

  if (nextCategory === homeGameSections.value[0]?.key && nearestDistance < Number.POSITIVE_INFINITY) {
    nextCategory = nearestCategory
  }

  if (nextCategory !== activeCategory.value) {
    activeCategory.value = nextCategory
  }
}

const settleRequestedCategory = () => {
  const category = requestedCategory
  if (!category) return

  const section = document.getElementById(`section-${category}`)
  activeCategory.value = category

  if (section) {
    const top = Math.max(0, window.scrollY + section.getBoundingClientRect().top - getSectionMarkerY())
    if (Math.abs(section.getBoundingClientRect().top - getSectionMarkerY()) > 6) {
      window.scrollTo({ top, behavior: 'auto' })
    }
  }

  if (categoryScrollTimer) window.clearTimeout(categoryScrollTimer)
  categoryScrollTimer = window.setTimeout(() => {
    requestedCategory = null
    categoryScrollTimer = 0
  }, 140)
}

const requestActiveCategoryUpdate = () => {
  if (requestedCategory) {
    if (categoryScrollTimer) window.clearTimeout(categoryScrollTimer)
    categoryScrollTimer = window.setTimeout(settleRequestedCategory, 240)
  }

  if (scrollFrame) return
  scrollFrame = window.requestAnimationFrame(updateActiveCategoryFromScroll)
}

const closeDownloadBanner = () => {
  downloadBannerVisible.value = false
  void nextTick(requestActiveCategoryUpdate)
}

const closeLoginInviteChest = () => {
  showLoginInviteChest.value = false
  if (androidAppRuntime) {
    homeBootState.initialGuideCompleted = true
    return
  }
  guideInitial.value = 'app-download'
  window.setTimeout(() => {
    if (!componentActive) return
    guideVisible.value = true
  }, 160)
}

const completeRegisterReward = () => {
  homeBootState.initialGuideCompleted = true
  clearRegisterReward()
  void refreshMockUser()
  window.setTimeout(resumeInitialHomeFlow, 120)
}

const completeNotificationPrompt = () => {
  notificationPromptVisible.value = false
  window.setTimeout(resumeInitialHomeFlow, 120)
}

const claimRegisterReward = () => claimPendingRegisterReward()

const claimActivityReward = (rewardID: number) => claimPendingActivityReward(rewardID)

const openPendingRewards = () => {
  if (!pendingActivityRewards.value.length) return
  rewardsDialogVisible.value = true
}

const onPendingRewardClaimed = () => {
  if (!pendingActivityRewards.value.length) {
    window.setTimeout(() => {
      rewardsDialogVisible.value = false
    }, 280)
  }
}

const openLogin = () => {
  void router.push({ name: 'auth-login' })
}

const openRegister = () => {
  void router.push({ name: 'auth-register' })
}

const openFloatingEvents = () => {
  void router.push({ name: 'offers', query: { tab: 'events' } })
}

const openGameCategory = (section: GameSection) => {
  void router.push({ name: 'game-category', params: { category: section.key } })
}

const getSectionPageCount = (section: GameSection) => {
  const pageSize = section.pageSize || 9
  return Math.max(1, Math.ceil((section.total ?? section.games.length) / pageSize))
}

const requestHomeSectionPage = (category: GameCategory, page: number) => {
  const key = `${category.key}:${page}`
  const pending = homePageRequests.get(key)
  if (pending) return pending

  const request = h5Api.getGameSection(category, page, 9).finally(() => {
    homePageRequests.delete(key)
  })
  homePageRequests.set(key, request)
  return request
}

const mergeHomeSectionPage = (sectionKey: CategoryKey, page: number, nextPage: GameSection, activate: boolean) => {
  homeGameSections.value = homeGameSections.value.map((item) => {
    if (item.key !== sectionKey) return item

    const pages = {
      ...(item.pages || { [item.page || 1]: item.games }),
      [page]: nextPage.games,
    }

    return {
      ...item,
      total: nextPage.total ?? item.total,
      pageSize: 9,
      pages,
      ...(activate ? { games: nextPage.games, page } : {}),
    }
  })
}

const preloadSectionPageCovers = (section: GameSection) => {
  void preloadAssets(
    section.games.map((game) => ({ src: game.cover, type: 'image' as const })),
    { concurrency: 4, timeoutMs: 6000 },
  )
}

const prefetchFollowingHomePage = async (sectionKey: CategoryKey, currentPage: number) => {
  const section = homeGameSections.value.find((item) => item.key === sectionKey)
  const category = homeCategories.value.find((item) => item.key === sectionKey)
  if (!section || !category || sectionKey === 'hot') return

  const nextPage = currentPage + 1
  if (nextPage > getSectionPageCount(section) || section.pages?.[nextPage]) return

  try {
    const prefetched = await requestHomeSectionPage(category, nextPage)
    mergeHomeSectionPage(sectionKey, nextPage, prefetched, false)
    preloadSectionPageCovers(prefetched)
  } catch (error) {
    console.warn('[api] failed to prefetch the next home game page.', error)
  }
}

const loadHomeSectionPage = async (section: GameSection, page: number) => {
  const category = homeCategories.value.find((item) => item.key === section.key)
  if (!category) return

  try {
    const latestSection = homeGameSections.value.find((item) => item.key === section.key) || section
    const cachedGames = latestSection.pages?.[page]

    if (cachedGames) {
      homeGameSections.value = homeGameSections.value.map((item) =>
        item.key === section.key ? { ...item, games: cachedGames, page } : item,
      )
    } else {
      const nextSection = await requestHomeSectionPage(category, page)
      mergeHomeSectionPage(section.key, page, nextSection, true)
      preloadSectionPageCovers(nextSection)
    }

    void prefetchFollowingHomePage(section.key, page)
    await nextTick(requestActiveCategoryUpdate)
  } catch (error) {
    console.warn('[api] failed to load home game section page.', error)
  }
}

const selectCategory = (category: CategoryKey) => {
  const section = document.getElementById(`section-${category}`)
  if (!section) return

  activeCategory.value = category
  requestedCategory = category
  if (categoryScrollTimer) window.clearTimeout(categoryScrollTimer)
  categoryScrollTimer = window.setTimeout(settleRequestedCategory, 2000)

  const top = Math.max(0, window.scrollY + section.getBoundingClientRect().top - getSectionMarkerY())
  window.scrollTo({ top, behavior: 'smooth' })
}

const applyHomeCatalog = (catalog: HomeCatalogSnapshot) => {
  homeProviders.value = catalog.home.providers ?? []
  homeBanners.value = catalog.home.banners.length ? catalog.home.banners : mockBanners
  homeCategories.value = catalog.home.categories
  homeSocialLinks.value = catalog.home.socialLinks
  homeGameSections.value = catalog.gameSections
  void prefetchAssetURLs(homeBanners.value.map((banner) => banner.image))

  for (const section of catalog.gameSections) {
    const initialGames = section.key === 'hot'
      ? section.games
      : Object.values(section.pages || { 1: section.games }).flat().slice(0, 18)
    preloadSectionPageCovers({ ...section, games: initialGames })
  }

  if (!homeCategories.value.some((category) => category.key === activeCategory.value)) {
    activeCategory.value = homeCategories.value[0]?.key ?? 'hot'
  }
}

const loadHomeData = async () => {
  try {
    applyHomeCatalog(await loadHomeCatalog())
    await nextTick(requestActiveCategoryUpdate)
  } catch (error) {
    console.warn('[api] failed to load home game catalog.', error)
    homeCategories.value = []
    homeGameSections.value = []
  } finally {
    homeBootState.catalogDone = true
    updateSplashReady()
  }
}

const refreshHomeAfterResume = () => {
  void loadHomeCatalog(true).then(applyHomeCatalog).catch((error) => {
    console.warn('[api] failed to refresh the resumed home catalog.', error)
  })
  void refreshMockUser().catch((error) => {
    console.warn('[api] failed to refresh the resumed wallet summary.', error)
  })
}

onMounted(async () => {
  componentActive = true
  window.addEventListener('scroll', requestActiveCategoryUpdate, { passive: true })
  window.addEventListener('resize', requestActiveCategoryUpdate, { passive: true })
  window.addEventListener('casino-app-resumed', refreshHomeAfterResume)
  void nextTick(requestActiveCategoryUpdate)
  void loadHomeData()
  stopActivityReadListener = onActivityReadChange(refreshHomeActivityUnread)
  void loadHomeActivityUnread()
  const rewardRequest = loadHomePendingReward()
  if (!splashVisible.value) {
    void rewardRequest.then(resumeInitialHomeFlow)
  }

  if (!homeBootState.preloadDone && !homeBootState.preloadPromise) {
    homeBootState.preloadPromise = loadMusicSessionWithin(4000)
      .catch(() => null)
      .then(async (musicSession) => {
        const currentMusic = musicSession?.playlist[0]?.audioUrl
        const preloadQueue = [
          ...criticalPreloadAssets,
          ...runtimePreloadAssets(),
          ...(currentMusic ? [{ src: currentMusic, type: 'audio' as const, label: 'background music', timeoutMs: 8000 }] : []),
        ]
        const result = await preloadAssets(preloadQueue, {
          concurrency: 7,
          onProgress: ({ percent }) => {
            homeBootState.preloadProgress = percent
            if (componentActive) {
              splashProgress.value = percent
            }
          },
        })
        const deferredURLs = [
          ...runtimeDeferredAssetURLs(),
          ...(musicSession?.playlist.slice(1).map((track) => track.audioUrl) || []),
        ]
        if (deferredURLs.length) {
          window.setTimeout(() => void prefetchAssetURLs(deferredURLs), 0)
        }

        if (result.failed > 0) {
          console.warn(`[preload] ${result.failed}/${result.total} critical resources did not finish before timeout.`)
        }

        homeBootState.preloadProgress = 100
        homeBootState.preloadDone = true
        updateSplashReady()
      })
  }

  await homeBootState.preloadPromise

  if (!componentActive) return

  splashProgress.value = 100
  updateSplashReady()
})

onBeforeUnmount(() => {
  componentActive = false
  stopActivityReadListener()
  window.removeEventListener('scroll', requestActiveCategoryUpdate)
  window.removeEventListener('resize', requestActiveCategoryUpdate)
  window.removeEventListener('casino-app-resumed', refreshHomeAfterResume)
  if (scrollFrame) window.cancelAnimationFrame(scrollFrame)
  if (categoryScrollTimer) window.clearTimeout(categoryScrollTimer)
  if (initialFlowTimer) window.clearTimeout(initialFlowTimer)
  homePageRequests.clear()
})
</script>

<template>
  <div class="home-shell text-white" :class="{ 'has-download-banner': downloadBannerVisible }">
    <TopDownloadBanner
      v-if="downloadBannerVisible"
      @close="closeDownloadBanner"
      @download="openDownloadGuide"
    />

      <AppHeader @login="openLogin" @register="openRegister" />

    <main class="home-content">
      <BannerSwiper :banners="homeBanners" />
      <div class="home-gutter">
        <OperationEntries position="HOME_QUICK" />
        <RebateJackpot />
        <ActivityCards />
      </div>

      <GameCategoryTabs
        v-if="homeCategories.length"
        :model-value="activeCategory"
        :categories="homeCategories"
        @update:model-value="selectCategory"
      />

      <GameSections
        :sections="homeGameSections"
        @play="launchGame"
        @view-all="openGameCategory"
        @page="loadHomeSectionPage"
      />
      <FooterInfo :socials="homeSocialLinks" :providers="homeProviders" />
    </main>

    <HomeFloatingActions
      v-if="!splashVisible"
      :unread-activity-slugs="unreadActivitySlugs"
      :pending-reward-count="pendingActivityRewards.length"
      @events="openFloatingEvents"
      @incentive="openFloatingEvents"
      @activity="markHomeActivityRead"
      @rewards="openPendingRewards"
    />

    <PendingRewardsDialog
      v-if="rewardsDialogVisible && !pendingRegisterReward && !splashVisible"
      :items="pendingActivityRewards"
      :claim-reward="claimActivityReward"
      @claimed="onPendingRewardClaimed"
      @close="rewardsDialogVisible = false"
    />

    <HomeInviteChestDialog
      v-if="showLoginInviteChest && !pendingRegisterReward && !notificationPromptVisible && !splashVisible && !guideVisible"
      @close="closeLoginInviteChest"
    />

    <RegisterRewardDialog
      v-if="pendingRegisterReward && !splashVisible"
      :reward="pendingRegisterReward"
      :claim-reward="claimRegisterReward"
      @complete="completeRegisterReward"
    />

    <NotificationPermissionPrompt
      v-if="notificationPromptVisible && !pendingRegisterReward && !splashVisible"
      @complete="completeNotificationPrompt"
    />

    <ActivityGuideFlow
      :visible="guideVisible && !pendingRegisterReward && !notificationPromptVisible"
      :initial="guideInitial"
      :chain-first-deposit="!isLoggedIn"
      @complete="onGuideComplete"
    />
    <WelcomeSplash
      v-if="splashVisible"
      :progress="splashProgress"
      :ready="splashReady"
      @complete="onSplashComplete"
    />

  </div>
</template>

<style scoped lang="scss">
.home-shell {
  position: relative;
  width: 100%;
  min-height: 100vh;
  max-width: 430px;
  margin: 0 auto;
  overflow: clip;
  background: var(--runtime-background-color, #090f1f);
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.04);
}

.home-content {
  min-height: 100vh;
  padding: 0;
  background: var(--runtime-background-color, #090f1f);
}

.home-gutter {
  padding: 0 14px;
}

.has-download-banner {
  :deep(.app-header) {
    top: 38px;
  }

  :deep(.category-tabs) {
    top: 104px;
  }
}

@media (max-width: 374px) {
  .home-gutter {
    padding-right: 10px;
    padding-left: 10px;
  }
}
</style>
