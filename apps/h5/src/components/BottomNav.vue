<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ActivityUnreadBadge from './ActivityUnreadBadge.vue'
import LottiePlayer from './LottiePlayer.vue'
import type { H5NavigationItem } from '../services/h5Api'
import { loadNavigation } from '../services/navigationCache'
import { openNavigationEntry } from '../services/navigationActions'
import {
  loadActivityUnreadState,
  onActivityReadChange,
  onActivityUnreadStateChange,
} from '../services/activityReadState'
import { runtimeConfig } from '../services/runtimeConfig'
import { h5Image } from '../utils/staticAssets'

const worldCup = h5Image('world-cup-source.gif')
const worldCupProviderGameCode = 'xa403d3a763119db'
const fallbackItems: H5NavigationItem[] = [
  { id: -1, menuCode: 'home', label: 'Home', labelZh: '首页', labelEn: 'Home', labels: { 'zh-CN': '首页', 'en-US': 'Home' }, iconURL: '', lottieURL: '', soundURL: '', displayPosition: 'BOTTOM_NAV', badgeType: 'NONE', actionType: 'OPEN_INTERNAL_PAGE', actionTarget: '/', providerCode: '', gameID: 0, providerGameCode: '', gameName: '', sortOrder: 10, configVersion: 0, enabled: true },
  { id: -2, menuCode: 'offers', label: 'Offers', labelZh: '优惠', labelEn: 'Offers', labels: { 'zh-CN': '优惠', 'en-US': 'Offers' }, iconURL: '', lottieURL: '', soundURL: '', displayPosition: 'BOTTOM_NAV', badgeType: 'NONE', actionType: 'OPEN_INTERNAL_PAGE', actionTarget: '/offers', providerCode: '', gameID: 0, providerGameCode: '', gameName: '', sortOrder: 20, configVersion: 0, enabled: true },
  { id: -3, menuCode: 'world', label: 'World Cup', labelZh: '世界杯', labelEn: 'World Cup', labels: { 'zh-CN': '世界杯', 'en-US': 'World Cup' }, iconURL: '', lottieURL: '', soundURL: '', displayPosition: 'BOTTOM_NAV', badgeType: 'NONE', actionType: 'OPEN_GAME', actionTarget: worldCupProviderGameCode, providerCode: 'softaggregator', gameID: 0, providerGameCode: worldCupProviderGameCode, gameName: 'World Cup', sortOrder: 30, configVersion: 0, enabled: true },
  { id: -4, menuCode: 'share', label: 'Share', labelZh: '分享', labelEn: 'Share', labels: { 'zh-CN': '分享', 'en-US': 'Share' }, iconURL: '', lottieURL: '', soundURL: '', displayPosition: 'BOTTOM_NAV', badgeType: 'NONE', actionType: 'OPEN_INTERNAL_PAGE', actionTarget: '/share', providerCode: '', gameID: 0, providerGameCode: '', gameName: '', sortOrder: 40, configVersion: 0, enabled: true },
  { id: -5, menuCode: 'profile', label: 'Profile', labelZh: '我的', labelEn: 'Profile', labels: { 'zh-CN': '我的', 'en-US': 'Profile' }, iconURL: '', lottieURL: '', soundURL: '', displayPosition: 'BOTTOM_NAV', badgeType: 'NONE', actionType: 'OPEN_INTERNAL_PAGE', actionTarget: '/profile', providerCode: '', gameID: 0, providerGameCode: '', gameName: '', sortOrder: 50, configVersion: 0, enabled: true },
]

const route = useRoute()
const router = useRouter()
const configuredItems = ref<H5NavigationItem[]>([])
const navigationLoaded = ref(false)
const offersUnread = ref(false)
const navItems = computed(() => {
  const items = navigationLoaded.value ? configuredItems.value : fallbackItems
  return items.map((item) => {
    if (item.menuCode === 'world' || item.menuCode === 'world-cup') {
      return { ...item, badgeType: 'NONE' as const }
    }
    return item
  })
})
const isChinese = computed(() => runtimeConfig.value.language.current.toLowerCase().startsWith('zh'))
const itemLabel = (item: H5NavigationItem) => item.label || (isChinese.value ? item.labelZh : item.labelEn)

const refreshNavigation = async () => {
  try {
    configuredItems.value = await loadNavigation('BOTTOM_NAV')
    navigationLoaded.value = true
  } catch (error) {
    console.warn('[navigation] using fallback menu.', error)
  }
}

const refreshOffersUnread = async () => {
  try {
    offersUnread.value = await loadActivityUnreadState()
  } catch (error) {
    console.warn('[offers] failed to resolve unread state.', error)
  }
}

let stopActivityReadListener: () => void = () => undefined
let stopActivityUnreadListener: () => void = () => undefined

onMounted(() => {
  void refreshNavigation()
  stopActivityReadListener = onActivityReadChange(() => void refreshOffersUnread())
  stopActivityUnreadListener = onActivityUnreadStateChange((unread) => {
    offersUnread.value = unread
  })
  void refreshOffersUnread()
})

onBeforeUnmount(() => {
  stopActivityReadListener()
  stopActivityUnreadListener()
})

const activeIndex = computed(() => {
  const routeName = String(route.name ?? '')
  if (routeName === 'game-play') {
    const providerCode = String(route.query.providerCode || '')
    const providerGameCode = String(route.query.providerGameCode || '')
    const gameId = Number(route.query.gameId || 0)
    const exact = navItems.value.findIndex((item) => item.actionType === 'OPEN_GAME'
      && ((item.gameID > 0 && item.gameID === gameId)
        || (item.providerCode === providerCode && item.providerGameCode === providerGameCode)))
    if (exact >= 0) return exact
  }
  const direct = navItems.value.findIndex((item) => item.actionType === 'OPEN_INTERNAL_PAGE'
    && (item.actionTarget === '/'
      ? route.path === '/'
      : route.path === item.actionTarget || route.path.startsWith(`${item.actionTarget}/`)))
  if (direct >= 0) return direct
  const groupedCode = routeName === 'activity-detail'
    ? 'offers'
    : ['profile', 'wallet-records', 'wallet-deposit', 'cash-center', 'bonus-center', 'bet-records', 'security', 'bank-account', 'support', 'about', 'user-agreement', 'settings'].includes(routeName)
      ? 'profile'
      : 'home'
  const index = navItems.value.findIndex((item) => item.menuCode === groupedCode)
  return index >= 0 ? index : 0
})

const visualActiveIndex = ref(0)
watch(activeIndex, (index) => {
  visualActiveIndex.value = index
}, { immediate: true })

const indicatorStyle = computed(() => ({
  left: navItems.value.length
    ? `calc(8px + (100% - 16px) * ${(visualActiveIndex.value + 0.5) / navItems.value.length})`
    : '50%',
}))

const openNavItem = (item: H5NavigationItem, index: number) => {
  visualActiveIndex.value = index
  if (item.actionType === 'OPEN_INTERNAL_PAGE' && item.actionTarget === route.path) return
  void openNavigationEntry(item, router).catch(() => {
    visualActiveIndex.value = activeIndex.value
  })
}
</script>

<template>
  <nav v-if="navItems.length" class="bottom-nav" :style="{ '--nav-count': navItems.length }" aria-label="Primary navigation">
    <button
      v-for="(item, index) in navItems"
      :key="item.id || item.menuCode"
      type="button"
      class="bottom-nav__item"
      :class="{ 'is-active': index === visualActiveIndex, 'is-featured': item.actionType === 'OPEN_GAME' }"
      :aria-current="index === activeIndex ? 'page' : undefined"
      @click="openNavItem(item, index)"
    >
      <span class="bottom-nav__icon" aria-hidden="true">
        <ActivityUnreadBadge v-if="item.menuCode === 'offers' && offersUnread" class="bottom-nav__activity-badge" />
        <i v-if="item.menuCode === 'profile'" class="bottom-nav__badge" />
        <LottiePlayer v-if="item.lottieURL" :src="item.lottieURL" />
        <img v-else-if="item.iconURL" :src="item.iconURL" alt="" />
        <svg v-else-if="item.menuCode === 'home'" viewBox="0 0 24 24">
          <path d="M4 4h6v6H4V4Zm10 0h6v6h-6V4ZM4 14h6v6H4v-6Zm10 0h6v6h-6v-6Z" />
        </svg>
        <svg v-else-if="item.menuCode === 'offers'" viewBox="0 0 24 24">
            <path d="M5 10h14v10H5V10Zm2-4a3 3 0 0 1 5-2.24A3 3 0 0 1 17 6v2H7V6Zm6 14V8h-2v12h2Z" />
        </svg>
        <span v-else-if="item.actionType === 'OPEN_GAME'" class="bottom-nav__placeholder" />
        <svg v-else-if="item.menuCode === 'share'" viewBox="0 0 24 24">
          <path
            d="M18 16.1c-.76 0-1.44.3-1.96.77L8.9 12.72a3.22 3.22 0 0 0 0-1.44l7.06-4.11A3 3 0 1 0 15 5c0 .23.03.45.08.66L8.02 9.77a3 3 0 1 0 0 4.46l7.12 4.16c-.04.19-.06.4-.06.61a2.92 2.92 0 1 0 2.92-2.9Z"
          />
        </svg>
        <svg v-else viewBox="0 0 24 24">
          <path
            d="M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-7 9a7 7 0 0 1 14 0H5Z"
          />
        </svg>
        <em v-if="item.badgeType !== 'NONE'" class="bottom-nav__entry-badge">{{ item.badgeType }}</em>
      </span>
      <span class="bottom-nav__label">{{ itemLabel(item) }}</span>
      <div v-if="item.menuCode === 'world' && !item.iconURL && !item.lottieURL" class="bottom-nav__featured-art" aria-hidden="true">
        <img :src="worldCup" alt="" />
      </div>
    </button>
    <span class="bottom-nav__indicator" :style="indicatorStyle" aria-hidden="true" />
  </nav>
</template>

<style scoped lang="scss">
.bottom-nav {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 50%;
  z-index: 30;
  display: grid;
  grid-template-columns: repeat(var(--nav-count, 5), minmax(0, 1fr));
  width: 100%;
  max-width: 430px;
  min-height: calc(var(--bottom-nav-height, 76px) + env(safe-area-inset-bottom));
  padding: 12px 8px calc(env(safe-area-inset-bottom) + 11px);
  border-top: 1px solid rgba(216, 234, 255, 0.09);
  background: rgba(8, 15, 30, 0.96);
  backdrop-filter: blur(16px);
  transform: translateX(-50%);
}

.bottom-nav__item {
  position: relative;
  display: grid;
  justify-items: center;
  gap: 5px;
  min-width: 0;
  padding: 0;
  color: rgba(207, 220, 248, 0.62);

  &.is-active {
    color: #ffffff;

    .bottom-nav__icon {
      transform: translateY(-2px) scale(1.08);
    }
  }

  &.is-featured {
    overflow: visible;
  }
}

.bottom-nav__icon {
  position: relative;
  display: grid;
  width: 27px;
  height: 27px;
  place-items: center;
  transition: transform 240ms ease;

  svg,
  img,
  .bottom-nav__placeholder {
    width: 27px;
    height: 27px;
  }

  img {
    object-fit: contain;
  }

  :deep(canvas),
  :deep(svg) {
    width: 100%;
    height: 100%;
  }

  svg {
    fill: currentColor;
  }
}

.bottom-nav__badge {
  position: absolute;
  top: -1px;
  right: -3px;
  z-index: 2;
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #ff344d;
  box-shadow: 0 0 7px rgba(255, 52, 77, 0.78);
}

.bottom-nav__activity-badge {
  top: -4px;
  right: -6px;
}

.bottom-nav__entry-badge {
  position: absolute;
  top: -7px;
  right: -15px;
  z-index: 3;
  min-width: 24px;
  padding: 2px 4px;
  border-radius: 6px;
  color: #ffffff;
  background: #ed3358;
  font-size: 8px;
  font-style: normal;
  font-weight: 700;
  line-height: 1;
  text-align: center;
}

.bottom-nav__placeholder {
  display: block;
}

.bottom-nav__featured-art {
  pointer-events: none;
  position: absolute;
  top: -31px;
  left: 50%;
  width: 62px;
  height: 62px;
  transform: translateX(-50%);

  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
}

.bottom-nav__label {
  display: block;
  max-width: 100%;
  overflow: hidden;
  font-size: 13px;
  font-weight: 300;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bottom-nav__indicator {
  position: absolute;
  bottom: calc(env(safe-area-inset-bottom) + 1px);
  left: 10%;
  width: 42px;
  height: 3px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(57, 127, 229, 0), #397fe5 48%, rgba(57, 127, 229, 0));
  box-shadow: 0 -1px 10px #397fe5;
  transform: translateX(-50%);
  transition: left 420ms cubic-bezier(0.22, 1, 0.36, 1);
  will-change: left;
}

</style>
