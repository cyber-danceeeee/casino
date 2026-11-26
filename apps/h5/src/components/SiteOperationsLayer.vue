<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { isLoggedIn } from '../mock/session'
import {
  h5Api,
  type H5AnnouncementPayload,
  type H5PopupPayload,
} from '../services/h5Api'
import { runtimeConfig, runtimeConfigReady } from '../services/runtimeConfig'
import { installPWA } from '../services/pwa'
import { resolvePopupComponent } from './popups/popupRegistry'

const route = useRoute()
const router = useRouter()
const announcements = ref<H5AnnouncementPayload[]>([])
const activeAnnouncement = ref<H5AnnouncementPayload | null>(null)
const announcementsVisible = ref(false)
const popupQueue = ref<H5PopupPayload[]>([])
const activePopup = ref<H5PopupPayload | null>(null)
let loadSequence = 0
const activePopupComponent = computed(() => activePopup.value ? resolvePopupComponent(activePopup.value.componentCode) : null)

const isChinese = computed(() => {
  const language = runtimeConfig.value.language.current || runtimeConfig.value.language.default
  return language.toLowerCase().startsWith('zh')
})
const copy = computed(() => isChinese.value
  ? { announcements: '公告', close: '关闭', published: '发布时间' }
  : { announcements: 'Announcements', close: 'Close', published: 'Published' })
const announcementCount = computed(() => Math.min(announcements.value.length, 99))

function popupMatchesRoute(popup: H5PopupPayload) {
  if (popup.triggerType === 'HOME_ENTER') return route.name === 'home'
  if (popup.triggerType !== 'PAGE_ENTER' || !popup.triggerValue) return false
  try {
    const target = router.resolve(popup.triggerValue)
    return popup.triggerValue.includes('?') ? target.fullPath === route.fullPath : target.path === route.path
  } catch {
    return false
  }
}

function showNextPopup() {
  if (activePopup.value) return
  activePopup.value = popupQueue.value.find((popup) => popupMatchesRoute(popup) && Boolean(resolvePopupComponent(popup.componentCode))) || null
  if (activePopup.value) popupQueue.value = popupQueue.value.filter((popup) => popup.campaignId !== activePopup.value?.campaignId)
  if (activePopup.value) {
    void h5Api.trackPopupEvent(activePopup.value.campaignId, 'impression').catch((error) => {
      console.warn('[site-content] failed to track popup impression.', error)
    })
  }
}

async function loadSiteContent() {
  if (!runtimeConfigReady.value) return
  const sequence = ++loadSequence
  try {
    const [announcementResult, popupResult] = await Promise.all([
      h5Api.getAnnouncements(),
      h5Api.getPopups(),
    ])
    if (sequence !== loadSequence) return
    announcements.value = announcementResult.list || []
    activeAnnouncement.value = announcements.value[0] || null
    popupQueue.value = (popupResult.items || []).filter((popup) => {
      const supported = Boolean(resolvePopupComponent(popup.componentCode))
      if (!supported) console.warn(`[site-content] unsupported popup component: ${popup.componentCode}`)
      return supported
    })
    activePopup.value = null
    showNextPopup()
  } catch (error) {
    if (sequence === loadSequence) console.warn('[site-content] failed to load published content.', error)
  }
}

function openAnnouncements() {
  activeAnnouncement.value ||= announcements.value[0] || null
  announcementsVisible.value = true
}

function closeAnnouncements() {
  announcementsVisible.value = false
}

function closePopup(track = true) {
  const popup = activePopup.value
  activePopup.value = null
  if (popup && track) {
    void h5Api.trackPopupEvent(popup.campaignId, 'dismiss').catch((error) => {
      console.warn('[site-content] failed to track popup dismissal.', error)
    })
  }
  window.setTimeout(showNextPopup, 180)
}

function safeInternalPath(target: string) {
  const value = target.trim()
  return value.startsWith('/') && !value.startsWith('//') ? value : ''
}

async function openGame(target: string) {
  const gameID = Number(target)
  if (!Number.isInteger(gameID) || gameID <= 0) return
  try {
    const game = await h5Api.getGameDetail(gameID)
    if (!game.providerCode || !game.providerGameCode) return
    const destination = router.resolve({
      name: 'game-play',
      query: {
        gameId: String(game.id),
        providerCode: game.providerCode,
        providerGameCode: game.providerGameCode,
        gameName: game.name,
      },
    }).fullPath
    if (!isLoggedIn.value) {
      await router.push({ name: 'auth-login', query: { redirect: destination } })
      return
    }
    await router.push(destination)
  } catch (error) {
    console.warn('[site-content] failed to open popup game.', error)
  }
}

async function executePopupAction(popup: H5PopupPayload) {
  void h5Api.trackPopupEvent(popup.campaignId, 'click').catch((error) => {
    console.warn('[site-content] failed to track popup click.', error)
  })
  const target = popup.action.target.trim()
  switch (popup.action.type) {
    case 'OPEN_INTERNAL_PAGE': {
      const path = safeInternalPath(target)
      if (path) await router.push(path)
      break
    }
    case 'OPEN_EXTERNAL_URL': {
      try {
        const url = new URL(target)
        if (url.protocol === 'https:') window.open(url.toString(), '_blank', 'noopener,noreferrer')
      } catch {
        // The server validates targets; ignore malformed stale content defensively.
      }
      break
    }
    case 'OPEN_GAME':
      await openGame(target)
      break
    case 'OPEN_PROMOTION': {
      const path = safeInternalPath(target)
      await router.push(path || { name: 'activity-detail', params: { slug: target } })
      break
    }
    case 'OPEN_DEPOSIT':
      await router.push({ name: 'wallet-deposit' })
      break
    case 'OPEN_WITHDRAW':
      await router.push({ name: 'cash-center' })
      break
    case 'OPEN_SUPPORT':
      await router.push({ name: 'support' })
      break
    case 'OPEN_COUPON':
      await router.push({ name: 'bonus-center' })
      break
    case 'OPEN_VIP':
      await router.push({ name: 'offers', query: { tab: 'vip' } })
      break
    case 'OPEN_DOWNLOAD':
      if (!await installPWA()) await router.push({ name: 'activity-detail', params: { slug: 'app-download' } })
      break
  }
  closePopup(false)
}

function executeActivePopupAction() {
  if (activePopup.value) void executePopupAction(activePopup.value)
}

function formatPublishedAt(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(isChinese.value ? 'zh-CN' : 'en-US', {
    year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit',
  }).format(date)
}

function onAuthChanged() {
  void loadSiteContent()
}

watch(runtimeConfigReady, (ready) => {
  if (ready) void loadSiteContent()
}, { immediate: true })
watch(() => [runtimeConfig.value.language.current, runtimeConfig.value.country.code, runtimeConfig.value.currency.code], () => {
  if (runtimeConfigReady.value) void loadSiteContent()
})
watch(() => route.fullPath, () => {
  announcementsVisible.value = false
  activePopup.value = null
  if (runtimeConfigReady.value) void loadSiteContent()
})

onMounted(() => window.addEventListener('casino-auth-changed', onAuthChanged))
onBeforeUnmount(() => window.removeEventListener('casino-auth-changed', onAuthChanged))
</script>

<template>
  <button
    v-if="announcements.length"
    class="site-notice-trigger"
    type="button"
    :aria-label="copy.announcements"
    @click="openAnnouncements"
  >
    <span aria-hidden="true">i</span>
    <b>{{ announcementCount }}</b>
  </button>

  <Teleport to="body">
    <div v-if="announcementsVisible" class="site-content-backdrop" @click.self="closeAnnouncements">
      <section class="site-announcement-dialog" role="dialog" aria-modal="true" :aria-label="copy.announcements">
        <header>
          <h2>{{ copy.announcements }}</h2>
          <button type="button" :aria-label="copy.close" @click="closeAnnouncements">&times;</button>
        </header>
        <div class="site-announcement-layout">
          <nav :aria-label="copy.announcements">
            <button
              v-for="announcement in announcements"
              :key="announcement.id"
              type="button"
              :class="{ 'is-active': activeAnnouncement?.id === announcement.id }"
              @click="activeAnnouncement = announcement"
            >
              <strong>{{ announcement.title }}</strong>
              <span>{{ announcement.summary || announcement.announcementType }}</span>
            </button>
          </nav>
          <article v-if="activeAnnouncement">
            <p class="site-announcement-type">{{ activeAnnouncement.announcementType }}</p>
            <h3>{{ activeAnnouncement.title }}</h3>
            <p v-if="activeAnnouncement.summary" class="site-announcement-summary">{{ activeAnnouncement.summary }}</p>
            <div class="site-announcement-content" v-html="activeAnnouncement.content" />
            <time v-if="formatPublishedAt(activeAnnouncement.publishedAt)">
              {{ copy.published }}: {{ formatPublishedAt(activeAnnouncement.publishedAt) }}
            </time>
          </article>
        </div>
      </section>
    </div>

    <component :is="activePopupComponent" v-if="activePopup && activePopupComponent" :campaign="activePopup" @close="closePopup()" @action="executeActivePopupAction" />
  </Teleport>
</template>

<style scoped lang="scss">
.site-notice-trigger {
  position: fixed;
  right: max(14px, calc((100vw - 430px) / 2 + 14px));
  bottom: calc(82px + env(safe-area-inset-bottom));
  z-index: 45;
  display: grid;
  width: 44px;
  height: 44px;
  place-items: center;
  border: 1px solid rgb(255 197 74 / 55%);
  border-radius: 8px;
  color: #181006;
  background: #ffc54a;
  box-shadow: 0 8px 20px rgb(0 0 0 / 32%);

  span {
    font: 800 20px/1 Georgia, serif;
  }

  b {
    position: absolute;
    top: -6px;
    right: -6px;
    min-width: 19px;
    height: 19px;
    padding: 0 4px;
    border: 2px solid #080b11;
    border-radius: 10px;
    color: white;
    background: #e64141;
    font-size: 10px;
    line-height: 15px;
  }
}

.site-content-backdrop {
  position: fixed;
  inset: 0;
  z-index: 210;
  display: grid;
  place-items: center;
  padding: max(18px, env(safe-area-inset-top)) 16px max(18px, env(safe-area-inset-bottom));
  background: rgb(0 0 0 / 76%);
  backdrop-filter: blur(5px);
}

.site-announcement-dialog {
  width: min(760px, 100%);
  max-height: min(720px, 88vh);
  overflow: hidden;
  border: 1px solid #323845;
  border-radius: 8px;
  color: #f5f6f8;
  background: #111318;
  box-shadow: 0 24px 70px rgb(0 0 0 / 55%);

  > header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 58px;
    padding: 0 18px;
    border-bottom: 1px solid #292e38;

    h2 {
      margin: 0;
      font-size: 18px;
    }

    button {
      width: 36px;
      height: 36px;
      border: 0;
      color: #d9dde5;
      background: transparent;
      font-size: 28px;
      line-height: 1;
    }
  }
}

.site-announcement-layout {
  display: grid;
  grid-template-columns: minmax(180px, 240px) minmax(0, 1fr);
  min-height: 360px;
  max-height: calc(min(720px, 88vh) - 59px);

  nav {
    overflow: auto;
    border-right: 1px solid #292e38;
    background: #0c0e12;

    button {
      display: block;
      width: 100%;
      padding: 14px 16px;
      border: 0;
      border-bottom: 1px solid #222630;
      color: #dfe3ea;
      text-align: left;
      background: transparent;

      &.is-active {
        color: white;
        background: #242933;
        box-shadow: inset 3px 0 #ffc54a;
      }

      strong,
      span {
        display: block;
      }

      strong {
        overflow: hidden;
        font-size: 14px;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      span {
        margin-top: 5px;
        overflow: hidden;
        color: #9299a5;
        font-size: 11px;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }

  article {
    overflow: auto;
    padding: 24px;

    h3 {
      margin: 5px 0 8px;
      font-size: 22px;
    }

    time {
      display: block;
      margin-top: 24px;
      color: #7f8794;
      font-size: 11px;
    }
  }
}

.site-announcement-type {
  margin: 0;
  color: #ffc54a;
  font-size: 11px;
  font-weight: 800;
}

.site-announcement-summary {
  margin: 0 0 18px;
  color: #aab1bd;
  font-size: 13px;
}

.site-announcement-content {
  color: #e7e9ed;
  font-size: 14px;
  line-height: 1.7;

  :deep(a) {
    color: #66b5ff;
  }
}

@media (max-width: 620px) {
  .site-announcement-dialog {
    max-height: 90vh;
  }

  .site-announcement-layout {
    grid-template-columns: 1fr;
    grid-template-rows: auto minmax(0, 1fr);
    max-height: calc(90vh - 59px);

    nav {
      display: flex;
      max-height: 92px;
      overflow-x: auto;
      border-right: 0;
      border-bottom: 1px solid #292e38;

      button {
        flex: 0 0 180px;
        border-right: 1px solid #222630;
        border-bottom: 0;

        &.is-active {
          box-shadow: inset 0 -3px #ffc54a;
        }
      }
    }

    article {
      padding: 20px;
    }
  }
}
</style>
