<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import ActivityGuideFlow from '../components/ActivityGuideFlow.vue'
import AppHeader from '../components/AppHeader.vue'
import type { H5OffersPayload } from '../services/h5Api'
import { isActivityUnread, loadCachedOffers, markActivityRead, onActivityReadChange, publishActivityUnreadState } from '../services/activityReadState'
import { isAndroidAppRuntime } from '../utils/runtimeEnvironment'
import { offerEvents as mockOfferEvents, type OfferEvent } from '../mock/offers'

const router = useRouter()
const androidAppRuntime = isAndroidAppRuntime()
const guideVisible = ref(false)
const offerEvents = ref<OfferEvent[]>(mockOfferEvents)
const mockEventBySlug = new Map(mockOfferEvents.flatMap((event) => (
  event.detailSlug ? [[event.detailSlug, event] as const] : []
)))
const mockEventById = new Map(mockOfferEvents.map((event) => [event.id, event] as const))

const isDownloadEvent = (event: Pick<OfferEvent, 'id' | 'detailSlug'>) =>
  event.id === 'app-download' || event.detailSlug === 'app-download'

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
    offerEvents.value = normalizeOfferEvents(await loadCachedOffers())
  } catch (error) {
    console.warn('[offers] failed to load events.', error)
    offerEvents.value = normalizeOfferEvents({ events: [], vip: { levels: [], benefits: [] } })
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
</script>

<template>
  <div class="offers-shell text-white">
    <AppHeader @login="openLogin" @register="openRegister" />
    <main class="offers-content">
      <section class="events-pane" aria-label="Events">
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
    </main>
    <ActivityGuideFlow v-if="guideVisible" :visible="guideVisible" initial="app-download" @complete="closeGuide" />
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

@media (max-width: 374px) {
  .events-pane { padding-right: 10px; padding-left: 10px; }
}
</style>
