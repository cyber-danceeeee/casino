<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import loadingAnimation from '../assets/lottie/lottie-loading.json'
import GameDepositSheet from '../components/GameDepositSheet.vue'
import LottiePlayer from '../components/LottiePlayer.vue'
import { refreshMockUser } from '../mock/session'
import { h5Api, type H5GameSessionExitReason } from '../services/h5Api'
import { runtimeConfig } from '../services/runtimeConfig'

const route = useRoute()
const router = useRouter()

const routeLaunchUrl = computed(() => String(route.query.url || ''))
const providerCode = computed(() => String(route.query.providerCode || ''))
const providerGameCode = computed(() => String(route.query.providerGameCode || ''))
const launchRouteKey = computed(() => [routeLaunchUrl.value, providerCode.value, providerGameCode.value].join('|'))
const launchUrl = ref('')
const gameId = computed(() => String(route.query.gameId || ''))
const routeFavorite = computed(() => String(route.query.favorite || '') === '1')
const frameLoaded = ref(false)
const sessionId = ref('')
const sessionClosed = ref(true)
const frameKey = ref(0)
const showQuickPanel = ref(false)
const isFavorite = ref(false)
const favoriteBusy = ref(false)
const depositSheetVisible = ref(false)
const launchBusy = ref(false)
const launchError = ref('')
const quickSide = ref<'left' | 'right'>('left')
const quickPosition = ref({ x: 0, y: 0 })
const quickWidth = ref(0)
const isQuickDragging = ref(false)
const quickVisible = computed(() => launchBusy.value || Boolean(launchUrl.value) || Boolean(launchError.value))
let launchSequence = 0
let launchedRouteKey = ''
let favoriteCollapseTimer = 0
let heartbeatTimer = 0
let quickPointerId = -1
let quickDragStartX = 0
let quickDragStartY = 0
let quickDragStartLeft = 0
let quickDragStartTop = 0
let quickDragMoved = false
let suppressQuickClick = false

const quickStyle = computed(() => ({
  width: `${quickWidth.value}px`,
  transform: `translate3d(${
    quickSide.value === 'right' ? quickPosition.value.x - (quickWidth.value - 42) : quickPosition.value.x
  }px, ${quickPosition.value.y}px, 0)`,
}))

const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max)

const getQuickBounds = () => {
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  const appWidth = Math.min(viewportWidth, 430)
  const appLeft = Math.max(0, (viewportWidth - appWidth) / 2)
  const appRight = appLeft + appWidth
  const width = Math.min(306, appWidth - 28)

  return {
    width,
    appLeft,
    appRight,
    leftX: appLeft + 10,
    rightX: appRight - 10 - 42,
    minX: appLeft + 10,
    maxX: appRight - 10 - 42,
    minY: 12,
    maxY: Math.max(12, viewportHeight - 68),
  }
}

const placeQuickPanel = (side = quickSide.value, y = quickPosition.value.y) => {
  const bounds = getQuickBounds()
  quickWidth.value = bounds.width
  quickSide.value = side
  quickPosition.value = {
    x: side === 'left' ? bounds.leftX : bounds.rightX,
    y: clamp(y || 36, bounds.minY, bounds.maxY),
  }
}

const handleQuickResize = () => {
  placeQuickPanel(quickSide.value, quickPosition.value.y)
}

const toggleQuickPanel = () => {
  if (suppressQuickClick) {
    suppressQuickClick = false
    return
  }

  showQuickPanel.value = !showQuickPanel.value
}

const stopHeartbeat = () => {
  if (heartbeatTimer) {
    window.clearInterval(heartbeatTimer)
    heartbeatTimer = 0
  }
}

const sendHeartbeat = async () => {
  if (!sessionId.value || sessionClosed.value) return

  try {
    await h5Api.heartbeatGameSession(sessionId.value)
  } catch (error) {
    console.warn('[game] session heartbeat failed.', error)
  }
}

const startHeartbeat = () => {
  if (!sessionId.value || sessionClosed.value || heartbeatTimer) return

  void sendHeartbeat()
  heartbeatTimer = window.setInterval(() => {
    void sendHeartbeat()
  }, 30000)
}

const closeGameSession = async (reason: H5GameSessionExitReason) => {
  if (!sessionId.value || sessionClosed.value) return

  const closingSessionId = sessionId.value
  sessionClosed.value = true
  stopHeartbeat()

  try {
    await h5Api.exitGameSession(closingSessionId, reason)
  } catch (error) {
    console.warn('[game] session exit failed.', error)
  }
}

const goBack = async () => {
  await closeGameSession('USER_EXIT')
  try {
    await refreshMockUser()
  } catch (error) {
    console.warn('[game] wallet refresh after exit failed.', error)
  }
  await router.push({ name: 'home' })
}

const openDeposit = () => {
  showQuickPanel.value = false
  depositSheetVisible.value = true
}

const favoriteStorageKey = computed(() => (gameId.value ? `game-casino-favorite:${gameId.value}` : ''))

const readStoredFavorite = () => {
  if (!favoriteStorageKey.value) return false
  return window.localStorage.getItem(favoriteStorageKey.value) === '1'
}

const writeStoredFavorite = (favorite: boolean) => {
  if (!favoriteStorageKey.value) return
  window.localStorage.setItem(favoriteStorageKey.value, favorite ? '1' : '0')
}

const scheduleQuickPanelCollapse = () => {
  if (favoriteCollapseTimer) {
    window.clearTimeout(favoriteCollapseTimer)
  }

  favoriteCollapseTimer = window.setTimeout(() => {
    showQuickPanel.value = false
    favoriteCollapseTimer = 0
  }, 650)
}

const loadFavorite = async () => {
  if (!gameId.value) {
    isFavorite.value = false
    return
  }

  isFavorite.value = route.query.favorite === undefined
    ? readStoredFavorite()
    : routeFavorite.value
  writeStoredFavorite(isFavorite.value)
}

const toggleFavorite = async () => {
  if (!gameId.value || favoriteBusy.value) return

  const previous = isFavorite.value
  const optimistic = !previous
  isFavorite.value = optimistic
  writeStoredFavorite(optimistic)
  favoriteBusy.value = true

  try {
    const result = await h5Api.toggleGameFavorite(gameId.value)
    isFavorite.value = result.favorite
    writeStoredFavorite(result.favorite)
  } catch (error) {
    isFavorite.value = previous
    writeStoredFavorite(previous)
    console.warn('[game] favorite toggle failed.', error)
  } finally {
    favoriteBusy.value = false
    scheduleQuickPanelCollapse()
  }
}

const refreshGame = () => {
  if (launchBusy.value) return
  void startLaunch(true)
}

const handleFrameLoaded = () => {
  frameLoaded.value = true
  startHeartbeat()
}

const stopQuickDrag = (event?: PointerEvent) => {
  if (quickPointerId < 0) return

  window.removeEventListener('pointermove', moveQuickDrag)
  window.removeEventListener('pointerup', stopQuickDrag)
  window.removeEventListener('pointercancel', stopQuickDrag)

  if (quickDragMoved && event) {
    const viewportCenter = window.innerWidth / 2
    const side = quickPosition.value.x + 21 < viewportCenter ? 'left' : 'right'
    placeQuickPanel(side, quickPosition.value.y)
    suppressQuickClick = true
  }

  quickPointerId = -1
  isQuickDragging.value = false
}

const moveQuickDrag = (event: PointerEvent) => {
  if (event.pointerId !== quickPointerId) return

  const deltaX = event.clientX - quickDragStartX
  const deltaY = event.clientY - quickDragStartY
  if (Math.abs(deltaX) > 4 || Math.abs(deltaY) > 4) {
    quickDragMoved = true
    showQuickPanel.value = false
  }

  const bounds = getQuickBounds()
  const nextX = clamp(quickDragStartLeft + deltaX, bounds.minX, bounds.maxX)
  quickWidth.value = bounds.width
  quickSide.value = nextX + 21 < window.innerWidth / 2 ? 'left' : 'right'
  quickPosition.value = {
    x: nextX,
    y: clamp(quickDragStartTop + deltaY, bounds.minY, bounds.maxY),
  }
  event.preventDefault()
}

const startQuickDrag = (event: PointerEvent) => {
  if (event.pointerType === 'mouse' && event.button !== 0) return

  quickPointerId = event.pointerId
  quickDragStartX = event.clientX
  quickDragStartY = event.clientY
  quickDragStartLeft = quickPosition.value.x
  quickDragStartTop = quickPosition.value.y
  quickDragMoved = false
  isQuickDragging.value = true

  window.addEventListener('pointermove', moveQuickDrag, { passive: false })
  window.addEventListener('pointerup', stopQuickDrag)
  window.addEventListener('pointercancel', stopQuickDrag)
}

const createLaunchSessionID = () => {
  const randomPart = typeof window.crypto?.randomUUID === 'function'
    ? window.crypto.randomUUID().replace(/-/g, '')
    : `${Date.now().toString(36)}${Math.random().toString(36).slice(2)}`
  return `web_${randomPart}`
}

const startLaunch = async (force = false) => {
  const routeKey = launchRouteKey.value
  if (launchBusy.value || (!force && launchedRouteKey === routeKey)) return
  launchBusy.value = true
  launchedRouteKey = routeKey
  const sequence = ++launchSequence
  launchError.value = ''

  await closeGameSession('SYSTEM_CLOSE')
  if (sequence !== launchSequence) return

  frameLoaded.value = false
  showQuickPanel.value = false
  stopHeartbeat()
  sessionId.value = ''
  sessionClosed.value = true
  frameKey.value += 1

  if (routeLaunchUrl.value && !force) {
    launchUrl.value = routeLaunchUrl.value
    launchBusy.value = false
    return
  }

  launchUrl.value = ''

  if (!providerCode.value || !providerGameCode.value) {
    launchError.value = 'Missing provider launch metadata'
    launchBusy.value = false
    return
  }

  try {
    const result = await h5Api.launchGameByCode(providerCode.value, providerGameCode.value, {
      currency: runtimeConfig.value.currency.code,
      language: runtimeConfig.value.language.current || runtimeConfig.value.language.default,
      country: runtimeConfig.value.country.code,
      device: 'mobile',
      return_url: window.location.origin,
      sessionId: createLaunchSessionID(),
    })
    if (sequence !== launchSequence) return

    if (!result.launchUrl) {
      throw new Error('empty launch url')
    }
    sessionId.value = result.sessionId || ''
    sessionClosed.value = !sessionId.value
    launchUrl.value = result.launchUrl
  } catch (error) {
    if (sequence !== launchSequence) return
    console.warn('[game] launch failed.', error)
    launchError.value = error instanceof Error ? error.message : 'Game launch failed'
  } finally {
    if (sequence === launchSequence) {
      launchBusy.value = false
    }
  }
}

watch(launchRouteKey, () => {
  void startLaunch()
}, { immediate: true })

watch(gameId, () => {
  void loadFavorite()
}, { immediate: true })

onMounted(() => {
  placeQuickPanel('left', 36)
  window.addEventListener('resize', handleQuickResize)
})

onBeforeUnmount(() => {
  if (favoriteCollapseTimer) {
    window.clearTimeout(favoriteCollapseTimer)
  }
  void closeGameSession('SYSTEM_CLOSE')
  window.removeEventListener('resize', handleQuickResize)
  stopQuickDrag()
})

const viewDepositOrder = () => {
  void router.push({ name: 'wallet-records', query: { tab: 'Deposit' } })
}
</script>

<template>
  <div class="game-play" :aria-busy="launchBusy || (Boolean(launchUrl) && !frameLoaded) ? 'true' : undefined">
    <main v-if="launchUrl" class="game-play__frame-wrap" :class="{ 'is-covered': depositSheetVisible }">
      <iframe
        :key="frameKey"
        class="game-play__frame"
        :class="{ 'is-loaded': frameLoaded }"
        :src="launchUrl"
        title="Game"
        allow="autoplay *; fullscreen *; clipboard-read *; clipboard-write *"
        allowfullscreen
        @load="handleFrameLoaded"
      />
    </main>

    <main v-else class="game-play__loading" aria-label="Loading game">
      <section v-if="launchError" class="game-play__error" role="alert">
        <strong>Game launch failed</strong>
        <p>{{ launchError }}</p>
        <button type="button" @click="startLaunch(true)">Retry</button>
      </section>
      <div v-else class="game-play__lottie">
        <LottiePlayer :animation-data="loadingAnimation" :loop="true" :autoplay="true" />
      </div>
    </main>

    <Transition name="game-play-loading">
      <div v-if="launchUrl && !frameLoaded" class="game-play__loading game-play__loading--overlay" aria-label="Loading game">
        <div class="game-play__lottie">
          <LottiePlayer :animation-data="loadingAnimation" :loop="true" :autoplay="true" />
        </div>
      </div>
    </Transition>

    <Transition name="game-play-float">
      <div
        v-if="quickVisible"
        class="game-play__quick"
        :class="{
          'is-open': showQuickPanel,
          'is-right': quickSide === 'right',
          'is-dragging': isQuickDragging,
        }"
        :style="quickStyle"
        role="toolbar"
        aria-label="Game quick actions"
      >
        <button
          class="game-play__quick-toggle"
          type="button"
          aria-label="Toggle quick actions"
          :aria-expanded="showQuickPanel"
          @pointerdown="startQuickDrag"
          @click="toggleQuickPanel"
        >
          <svg v-if="showQuickPanel" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M6.292893,7.707107Q6.152241,7.566455,6.07612,7.382683Q6,7.198912,6,7Q6,6.801087,6.076121,6.617317Q6.152241,6.433545,6.292893,6.292893Q6.433545,6.152241,6.617317,6.076121Q6.801087,6,7,6Q7.198912,6,7.382683,6.07612Q7.566455,6.152241,7.707107,6.292893L17.7071,16.29289Q17.8478,16.43355,17.9239,16.61732Q18,16.80109,18,17Q18,17.19891,17.9239,17.3827Q17.8478,17.56645,17.7071,17.7071Q17.56645,17.8478,17.3827,17.9239Q17.19891,18,17,18Q16.80109,18,16.61732,17.9239Q16.43355,17.8478,16.29289,17.7071L6.292893,7.707107Z" />
            <path d="M17.7071,7.707107Q17.8478,7.566455,17.9239,7.382683Q18,7.198912,18,7Q18,6.801087,17.9239,6.617317Q17.8478,6.433545,17.7071,6.292893Q17.56645,6.152241,17.3827,6.076121Q17.19891,6,17,6Q16.80109,6,16.61732,6.07612Q16.43355,6.152241,16.29289,6.292893L6.292893,16.29289Q6.152241,16.43355,6.076121,16.61732Q6,16.80109,6,17Q6,17.19891,6.076121,17.3827Q6.152241,17.56645,6.292893,17.7071Q6.433545,17.8478,6.617317,17.9239Q6.801087,18,7,18Q7.198912,18,7.382683,17.9239Q7.566455,17.8478,7.707107,17.7071L17.7071,7.707107Z" />
          </svg>
          <svg v-else viewBox="0 0 28 28" aria-hidden="true">
            <path d="M4.63720703125,5.97021484375L23.30390703125,5.97021484375Q23.71810703125,5.97021484375,24.01100703125,6.26310784375Q24.30390703125,6.55600084375,24.30390703125,6.97021484375Q24.30390703125,7.38442784375,24.01100703125,7.67732184375Q23.71810703125,7.97021484375,23.30390703125,7.97021484375L4.63720703125,7.97021484375Q4.22299403125,7.97021484375,3.93010003125,7.67732184375Q3.63720703125,7.38442784375,3.63720703125,6.97021484375Q3.63720703125,6.55600084375,3.93010003125,6.26310784375Q4.22299403125,5.97021484375,4.63720703125,5.97021484375Z" />
            <path d="M4.63720703125,12.97021484375L23.30390703125,12.97021484375Q23.71810703125,12.97021484375,24.01100703125,13.26310784375Q24.30390703125,13.55600084375,24.30390703125,13.97021484375Q24.30390703125,14.38442784375,24.01100703125,14.67732184375Q23.71810703125,14.97021484375,23.30390703125,14.97021484375L4.63720703125,14.97021484375Q4.22299403125,14.97021484375,3.93010003125,14.67732184375Q3.63720703125,14.38442784375,3.63720703125,13.97021484375Q3.63720703125,13.55600084375,3.93010003125,13.26310784375Q4.22299403125,12.97021484375,4.63720703125,12.97021484375Z" />
            <path d="M4.63720703125,19.97021484375L23.30390703125,19.97021484375Q23.71810703125,19.97021484375,24.01100703125,20.26310784375Q24.30390703125,20.55600084375,24.30390703125,20.97021484375Q24.30390703125,21.38442784375,24.01100703125,21.67732184375Q23.71810703125,21.97021484375,23.30390703125,21.97021484375L4.63720703125,21.97021484375Q4.22299403125,21.97021484375,3.93010003125,21.67732184375Q3.63720703125,21.38442784375,3.63720703125,20.97021484375Q3.63720703125,20.55600084375,3.93010003125,20.26310784375Q4.22299403125,19.97021484375,4.63720703125,19.97021484375Z" />
          </svg>
        </button>

        <div class="game-play__quick-panel" :aria-hidden="!showQuickPanel" :inert="!showQuickPanel">
          <button class="game-play__quick-item" type="button" @click="openDeposit">
            <svg viewBox="0 0 22 22" aria-hidden="true">
              <path d="M8.24061310875,6.48590609375L16.680819218750003,6.49507609375Q16.94934921875,6.49536609375,17.18190921875,6.36109609375Q17.354169218750002,6.26164609375,17.47525921875,6.10383609375Q17.59634921875,5.94602609375,17.64783921875,5.75389609375Q17.69931921875,5.561756093750001,17.673349218749998,5.3645460937500005Q17.64738921875,5.16733609375,17.54792921875,4.99507609375L15.433919218749999,1.33349609375Q15.38466921875,1.24820009375,15.31972921875,1.17415009375Q15.25478921875,1.10010009375,15.17664921875,1.04014309375Q15.098519218749999,0.98018509375,15.01017921875,0.93662309375Q14.921849218750001,0.89306209375,14.82670921875,0.86757009375Q14.73157921875,0.84207909375,14.63329921875,0.83563709375Q14.53501921875,0.82919609375,14.43736921875,0.84205109375Q14.33971921875,0.85490709375,14.24644921875,0.88656509375Q14.15318921875,0.91822409375,14.06788921875,0.96747009375L7.74170121875,4.61988609375Q7.62780521875,4.68563609375,7.53478521875,4.77860609375Q7.44176521875,4.87157609375,7.37594621875,4.98543609375Q7.31012621875,5.0992960937500005,7.27598421875,5.22630609375Q7.24184221875,5.35330609375,7.24170021875,5.48482609375Q7.24158921875,5.58331609375,7.26070221875,5.679936093749999Q7.27981221875,5.77655609375,7.31740421875,5.86758609375Q7.35499721875,5.95862609375,7.40962721875,6.04057609375Q7.46425621875,6.12252609375,7.53382521875,6.19224609375Q7.60339321875,6.26196609375,7.68522621875,6.31677609375Q7.76705921875,6.37158609375,7.85801221875,6.40937609375Q7.94896621875,6.44716609375,8.04554421875,6.46648609375Q8.14212181875,6.48580609375,8.24061310875,6.48590609375ZM14.94876921875,4.49319609375L11.96675921875,4.48995609375L14.20186921875,3.19951609375L14.94876921875,4.49319609375Z" fill-rule="evenodd" />
              <path d="M0.833251953125,6.416667L0.833251953125,19.25Q0.833251953125,20.0439,1.394632953125,20.6053Q1.956008953125,21.1667,2.749918953125,21.1667L19.249951953125,21.1667Q20.043851953125,21.1667,20.605251953125,20.6053Q21.166551953125,20.0439,21.166551953125,19.25L21.166551953125,6.416667Q21.166551953125,5.622748,20.605251953125,5.061373Q20.043851953125,4.5,19.249951953125,4.5L2.749918953125,4.5Q1.956009953125,4.5,1.394633953125,5.06137Q0.833251953125,5.622746,0.833251953125,6.416667ZM2.833251953125,19.1667L19.166551953125,19.1667L19.166551953125,6.5L2.833251953125,6.5L2.833251953125,19.1667Z" fill-rule="evenodd" />
              <path d="M13.7651167,10.48858510625Q12.75,11.45536690625,12.75,12.83317390625Q12.75,14.21097390625,13.7651166,15.17775390625Q14.75955,16.12483390625,16.15625,16.12483390625L20.16667,16.12483390625Q20.26516,16.12483390625,20.36176,16.10562390625Q20.45836,16.08640390625,20.54935,16.04871390625Q20.640340000000002,16.01102390625,20.72224,15.95630390625Q20.80413,15.90158390625,20.87377,15.83194390625Q20.94342,15.76230390625,20.99814,15.68040390625Q21.05285,15.59851390625,21.09055,15.50752390625Q21.128239999999998,15.41652390625,21.14745,15.31992390625Q21.16667,15.22332390625,21.16667,15.12483390625L21.16667,10.54150390625Q21.16667,10.44301250625,21.14745,10.34641390625Q21.128239999999998,10.24981490625,21.09055,10.15882090625Q21.05285,10.06782590625,20.99814,9.98593390625Q20.94342,9.90404090625,20.87377,9.83439690625Q20.80413,9.76475290625,20.72224,9.71003490625Q20.640340000000002,9.65531590625,20.54935,9.61762490625Q20.45836,9.57993390625,20.36176,9.560718906249999Q20.26516,9.54150390625,20.16667,9.54150390625L16.15625,9.54150390625Q14.75955,9.54150390625,13.7651167,10.48858510625ZM15.14443,13.72948390625Q14.75,13.35383390625,14.75,12.83317390625Q14.75,12.31250390625,15.14443,11.93685390625Q15.55955,11.54150390625,16.15625,11.54150390625L19.16667,11.54150390625L19.16667,14.12483390625L16.15625,14.12483390625Q15.55955,14.12483390625,15.14443,13.72948390625Z" fill-rule="evenodd" />
              <path d="M19.166748046875,7.5625Q19.166748046875,7.4640086,19.185963046875,7.36741Q19.205178046874998,7.270811,19.242869046875,7.179817Q19.280560046875,7.088822,19.335278046875,7.00693Q19.389997046875,6.925037,19.459641046875,6.855393Q19.529285046875,6.785749,19.611178046875,6.73103Q19.693070046875,6.676312,19.784065046875,6.638621Q19.875059046875,6.60093,19.971658046875,6.581715Q20.068256646875,6.5625,20.166748046875,6.5625Q20.265239446875,6.5625,20.361838046875,6.581715Q20.458437046875,6.60093,20.549431046875,6.638621Q20.640426046875,6.676312,20.722318046875,6.73103Q20.804211046875,6.785749,20.873855046875,6.855393Q20.943499046875,6.925037,20.998218046875,7.00693Q21.052936046875,7.088822,21.090627046875,7.179817Q21.128318046875002,7.270811,21.147533046875,7.36741Q21.166748046875,7.4640086,21.166748046875,7.5625L21.166748046875,18.5625Q21.166748046875,18.661,21.147533046875,18.7576Q21.128318046875002,18.8542,21.090627046875,18.9452Q21.052936046875,19.0362,20.998218046875,19.1181Q20.943499046875,19.2,20.873855046875,19.2696Q20.804211046875,19.339199999999998,20.722318046875,19.394Q20.640426046875,19.448700000000002,20.549431046875,19.4864Q20.458437046875,19.5241,20.361838046875,19.543300000000002Q20.265239446875,19.5625,20.166748046875,19.5625Q20.068256646875,19.5625,19.971658046875,19.543300000000002Q19.875059046875,19.5241,19.784065046875,19.4864Q19.693070046875,19.448700000000002,19.611178046875,19.394Q19.529285046875,19.339199999999998,19.459641046875,19.2696Q19.389997046875,19.2,19.335278046875,19.1181Q19.280560046875,19.0362,19.242869046875,18.9452Q19.205178046874998,18.8542,19.185963046875,18.7576Q19.166748046875,18.661,19.166748046875,18.5625L19.166748046875,7.5625Z" fill-rule="evenodd" />
            </svg>
            <span>Deposit</span>
          </button>

          <button class="game-play__quick-item" type="button" :class="{ 'is-active': isFavorite }" :disabled="favoriteBusy || !gameId" @click="toggleFavorite">
            <svg v-if="!isFavorite" viewBox="0 0 16 14" aria-hidden="true">
              <path d="M4.973308828125,0.07505603125000004Q2.388578828125,-0.27521796875000004,0.991576828125,1.55017303125L0.991575828125,1.55017403125Q0.050050828124999946,2.78041203125,0.050048828125,4.35158203125Q0.050045828125000025,5.92274203125,0.991567828125,7.15298203125L5.098838828125,12.51973203125Q6.210498828125,13.97223203125,8.000048828125,13.97223203125Q9.789598828125,13.97223203125,10.901248828125,12.51973203125L15.008548828125,7.15299203125Q15.950048828125,5.92275203125,15.950048828125,4.35158203125Q15.950048828125,2.78041203125,15.008548828125,1.55016403125Q13.611548828125,-0.27521796875000004,11.026748828125,0.07505603125000004Q9.144428828125,0.33014503125,8.000048828125,1.5311830312499999Q6.855658828125,0.33014503125,4.973308828125,0.07505603125000004ZM1.550048828125,4.35158203125Q1.5500498281250001,3.28854203125,2.1827588281250003,2.46181203125Q3.050348828125,1.32817603125,4.771868828125,1.5614700312499998Q6.532928828125,1.80012203125,7.360408828125,3.15166203125Q7.454598828125,3.30549203125,7.608428828125,3.39967203125Q7.735658828125,3.47757203125,7.883018828125,3.50085203125Q8.030378828125,3.52413203125,8.175428828125,3.48924203125Q8.320468828125,3.45436203125,8.441128828124999,3.36662203125Q8.561788828125,3.27889203125,8.639678828125,3.15165203125Q9.467158828125,1.80012203125,11.228248828125,1.5614690312500001Q12.949748828125,1.32817503125,13.817348828125,2.46181203125Q14.450048828125,3.28854203125,14.450048828125,4.35158203125Q14.450048828125,5.41462203125,13.817348828125,6.24135203125L9.710058828125,11.60813203125Q9.048708828125,12.47223203125,8.000048828125,12.47223203125Q6.951378828125,12.47223203125,6.290028828125,11.60813203125L2.182748828125,6.24135203125Q1.5500468281250002,5.41462203125,1.550048828125,4.35158203125Z" fill-rule="evenodd" />
            </svg>
            <svg v-else viewBox="0 0 14.400019645690918 12.444425582885742" aria-hidden="true" class="game-play__heart-solid">
              <path d="M0.787112,5.91933L4.89438,11.2861C6.07639,12.8305,8.3236,12.8305,9.5056,11.2861L13.6129,5.91934C14.6624,4.54802,14.6624,2.59947,13.6129,1.22816C12.0766,-0.779254,8.52568,-0.183052,7.2,1.98221C5.87431,-0.183052,2.32342,-0.779254,0.78712,1.22816C-0.262371,2.59947,-0.262373,4.54802,0.787112,5.91933Z" />
            </svg>
            <span>Favs</span>
          </button>

          <button class="game-play__quick-item" type="button" @click="refreshGame">
            <svg viewBox="0 0 18 18" aria-hidden="true">
              <path d="M14.75,3Q14.75,2.9015086,14.769214999999999,2.80491Q14.78843,2.708311,14.826121,2.617317Q14.863812,2.526322,14.91853,2.44443Q14.973249,2.362537,15.042893,2.292893Q15.112537,2.223249,15.19443,2.16853Q15.276322,2.1138120000000002,15.367317,2.076121Q15.458311,2.03843,15.55491,2.019215Q15.6515086,2,15.75,2Q15.8484914,2,15.94509,2.019215Q16.041689,2.03843,16.132683,2.076121Q16.223678,2.1138120000000002,16.30557,2.16853Q16.387463,2.223249,16.457107,2.292893Q16.526751,2.362537,16.58147,2.44443Q16.636188,2.526322,16.673879,2.617317Q16.711570000000002,2.708311,16.730785,2.80491Q16.75,2.9015086,16.75,3L16.75,9Q16.75,9.09849,16.730785,9.19509Q16.711570000000002,9.29169,16.673879,9.38268Q16.636188,9.47368,16.58147,9.55557Q16.526751,9.63746,16.457107,9.70711Q16.387463,9.77675,16.30557,9.83147Q16.223678,9.88619,16.132683,9.92388Q16.041689,9.96157,15.94509,9.98078Q15.8484914,10,15.75,10Q15.6515086,10,15.55491,9.98078Q15.458311,9.96157,15.367317,9.92388Q15.276322,9.88619,15.19443,9.83147Q15.112537,9.77675,15.042893,9.70711Q14.973249,9.63746,14.91853,9.55557Q14.863812,9.47368,14.826121,9.38268Q14.78843,9.29169,14.769214999999999,9.19509Q14.75,9.09849,14.75,9L14.75,3Z" fill-rule="evenodd" />
              <path d="M1.25,9Q1.25,8.9015086,1.269215,8.80491Q1.28843,8.708311,1.326121,8.617317Q1.363812,8.526322,1.41853,8.44443Q1.473249,8.362537,1.5428929999999998,8.292893Q1.612537,8.223249,1.69443,8.16853Q1.776322,8.113812,1.867317,8.076121Q1.9583110000000001,8.03843,2.05491,8.019214999999999Q2.1515086,8,2.25,8Q2.3484914,8,2.44509,8.019214999999999Q2.541689,8.03843,2.632683,8.076121Q2.723678,8.113812,2.80557,8.16853Q2.887463,8.223249,2.957107,8.292893Q3.026751,8.362537,3.08147,8.44443Q3.1361879999999998,8.526322,3.173879,8.617317Q3.21157,8.708311,3.230785,8.80491Q3.25,8.9015086,3.25,9L3.25,15Q3.25,15.09849,3.230785,15.19509Q3.21157,15.29169,3.173879,15.38268Q3.1361879999999998,15.47368,3.08147,15.55557Q3.026751,15.63746,2.957107,15.70711Q2.887463,15.77675,2.80557,15.83147Q2.723678,15.88619,2.632683,15.92388Q2.541689,15.96157,2.44509,15.98078Q2.3484914,16,2.25,16Q2.1515086,16,2.05491,15.98078Q1.9583110000000001,15.96157,1.867317,15.92388Q1.776322,15.88619,1.69443,15.83147Q1.612537,15.77675,1.5428929999999998,15.70711Q1.473249,15.63746,1.41853,15.55557Q1.363812,15.47368,1.326121,15.38268Q1.28843,15.29169,1.269215,15.19509Q1.25,15.09849,1.25,15L1.25,9Z" fill-rule="evenodd" />
              <path d="M9,1.25Q5.70947,1.25,3.42357,3.6179699999999997Q3.2885999999999997,3.75779,3.21582,3.93798Q3.1430379999999998,4.11817,3.1430379999999998,4.3125Q3.1430379999999998,4.41099,3.162252,4.50759Q3.181467,4.60419,3.219158,4.695180000000001Q3.25685,4.78618,3.3115699999999997,4.8680699999999995Q3.3662900000000002,4.94996,3.43593,5.01961Q3.50557,5.08925,3.5874699999999997,5.1439699999999995Q3.66936,5.19869,3.76035,5.2363800000000005Q3.85135,5.27407,3.94795,5.29329Q4.04455,5.3125,4.14304,5.3125Q4.34654,5.3125,4.53386,5.23297Q4.72118,5.15343,4.86251,5.00702L4.86262,5.0069099999999995Q6.55871,3.25,9,3.25Q11.38172,3.25,13.0659,4.93414Q14.75,6.61828,14.75,9Q14.75,9.09849,14.7692,9.19509Q14.7884,9.29169,14.8261,9.38268Q14.8638,9.47368,14.9185,9.55557Q14.9732,9.63746,15.0429,9.70711Q15.1125,9.77675,15.1944,9.83147Q15.2763,9.88619,15.3673,9.92388Q15.4583,9.96157,15.5549,9.98078Q15.6515,10,15.75,10Q15.8485,10,15.9451,9.98078Q16.0417,9.96157,16.1327,9.92388Q16.2237,9.88619,16.3056,9.83147Q16.3875,9.77675,16.4571,9.70711Q16.526699999999998,9.63746,16.5815,9.55557Q16.636200000000002,9.47368,16.6739,9.38268Q16.7116,9.29169,16.730800000000002,9.19509Q16.75,9.09849,16.75,9L16.75,8.99854Q16.749499999999998,5.78933,14.4801,3.51993Q12.21015,1.25,9,1.25ZM3.25,9Q3.25,8.90151,3.230785,8.80491Q3.21157,8.70831,3.173879,8.61732Q3.1361879999999998,8.52632,3.08147,8.44443Q3.026751,8.36254,2.957107,8.29289Q2.887463,8.22325,2.80557,8.16853Q2.723678,8.11381,2.632683,8.07612Q2.541689,8.03843,2.44509,8.019210000000001Q2.3484914,8,2.25,8Q2.1515086,8,2.05491,8.019210000000001Q1.9583110000000001,8.03843,1.867317,8.07612Q1.776322,8.11381,1.69443,8.16853Q1.612537,8.22325,1.5428929999999998,8.29289Q1.473249,8.36254,1.41853,8.44443Q1.363812,8.52632,1.326121,8.61732Q1.28843,8.70831,1.269215,8.80491Q1.25,8.90151,1.25,9L1.25,9.000969999999999Q1.2503440000000001,12.21049,3.51993,14.4801Q5.7898499999999995,16.75,9,16.75Q12.12948,16.75,14.3807,14.5777L14.3818,14.5767L14.382,14.5764Q14.5284,14.4351,14.608,14.2478Q14.6875,14.0605,14.6875,13.857Q14.6875,13.7585,14.6683,13.6619Q14.6491,13.5653,14.6114,13.4743Q14.5737,13.3833,14.519,13.3014Q14.4642,13.2195,14.3946,13.1499Q14.325,13.0802,14.2431,13.0255Q14.1612,12.9708,14.0702,12.9331Q13.9792,12.8954,13.8826,12.8762Q13.786,12.857,13.6875,12.857Q13.4932,12.857,13.313,12.9297Q13.1328,13.0025,12.993,13.1375Q11.32253,14.75,9,14.75Q6.61828,14.75,4.93414,13.0659Q3.25,11.38172,3.25,9Z" fill-rule="evenodd" />
            </svg>
            <span>Refresh</span>
          </button>

          <button class="game-play__quick-item" type="button" @click="goBack">
            <svg viewBox="0 0 12 12" aria-hidden="true">
              <path d="M1.5,11.25L6,11.25Q6.07387,11.25,6.14632,11.23559Q6.21877,11.22118,6.28701,11.19291Q6.35526,11.16464,6.41668,11.1236Q6.4781,11.08256,6.53033,11.03033Q6.58256,10.9781,6.6236,10.91668Q6.66464,10.85526,6.69291,10.78701Q6.72118,10.71877,6.73559,10.64632Q6.75,10.57387,6.75,10.5Q6.75,10.42613,6.73559,10.35368Q6.72118,10.28123,6.69291,10.21299Q6.66464,10.14474,6.6236,10.08332Q6.58256,10.0219,6.53033,9.96967Q6.4781,9.91744,6.41668,9.8764Q6.35526,9.83536,6.28701,9.80709Q6.21877,9.77882,6.14632,9.76441Q6.07387,9.75,6,9.75L2.25,9.75L2.25,2.25L5.99792,2.25Q6.07179,2.25,6.14424,2.235589Q6.21669,2.221178,6.28494,2.19291Q6.35318,2.164641,6.4146,2.123602Q6.47602,2.082563,6.52825,2.03033Q6.58049,1.978097,6.62153,1.916678Q6.66257,1.855258,6.69083,1.787013Q6.7191,1.718767,6.73351,1.646318Q6.74792,1.5738685000000001,6.74792,1.5Q6.74792,1.4261314999999999,6.73351,1.353682Q6.7191,1.281233,6.69083,1.212987Q6.66257,1.144742,6.62153,1.083322Q6.58049,1.021903,6.52825,0.96967Q6.47602,0.917437,6.4146,0.876398Q6.35318,0.835359,6.28494,0.80709Q6.21669,0.778822,6.14424,0.764411Q6.07179,0.75,5.99792,0.75L1.5,0.75Q1.4261314999999999,0.75,1.353682,0.764411Q1.281233,0.778822,1.212987,0.80709Q1.144742,0.835359,1.083322,0.876398Q1.021903,0.917437,0.96967,0.96967Q0.917437,1.021903,0.876398,1.083322Q0.835359,1.144742,0.80709,1.212987Q0.778822,1.281233,0.764411,1.353682Q0.75,1.4261314999999999,0.75,1.5L0.75,10.5Q0.75,10.57387,0.764411,10.64632Q0.778822,10.71877,0.80709,10.78701Q0.835359,10.85526,0.876398,10.91668Q0.917437,10.9781,0.96967,11.03033Q1.021903,11.08256,1.083322,11.1236Q1.144742,11.16464,1.212987,11.19291Q1.281233,11.22118,1.353682,11.23559Q1.4261314999999999,11.25,1.5,11.25Z" fill-rule="evenodd" />
              <path d="M8.780484,3.219824L11.03033,5.46967Q11.08256,5.5219000000000005,11.1236,5.5833200000000005Q11.16464,5.6447400000000005,11.19291,5.71299Q11.22118,5.78123,11.23559,5.853680000000001Q11.25,5.926130000000001,11.25,6Q11.25,6.073869999999999,11.23559,6.146319999999999Q11.22118,6.21877,11.19291,6.28701Q11.16464,6.3552599999999995,11.1236,6.4166799999999995Q11.08256,6.4780999999999995,11.03033,6.53033L8.78033,8.78033Q8.674841,8.885819999999999,8.537013,8.942910000000001Q8.399184,9,8.25,9Q8.1761315,9,8.103682,8.98559Q8.031233,8.97118,7.962987,8.942910000000001Q7.894742,8.91464,7.833322,8.8736Q7.771903,8.83256,7.71967,8.78033Q7.667437,8.728100000000001,7.626398,8.66668Q7.585359,8.605260000000001,7.55709,8.53701Q7.528822,8.46877,7.514411,8.39632Q7.5,8.32387,7.5,8.25Q7.5,8.100819999999999,7.55709,7.96299Q7.614181,7.82516,7.71967,7.71967L9.43934,6L7.719824,4.2804839999999995L7.71967,4.28033Q7.614181,4.174841,7.55709,4.037013Q7.5,3.899184,7.5,3.75Q7.5,3.6761315,7.514411,3.603682Q7.528822,3.531233,7.55709,3.462987Q7.585359,3.394742,7.626398,3.333322Q7.667437,3.271903,7.71967,3.21967Q7.771903,3.167437,7.833322,3.126398Q7.894742,3.085359,7.962987,3.05709Q8.031233,3.028822,8.103682,3.014411Q8.1761315,3,8.25,3Q8.399184,3,8.537013,3.05709Q8.674841,3.114181,8.78033,3.21967L8.780484,3.219824Z" fill-rule="evenodd" />
              <path d="M3.999955654144287,5.248135566711426L10.499955654144287,5.248135566711426Q10.573825654144287,5.248135566711426,10.646275654144286,5.262546566711426Q10.718725654144286,5.276957566711426,10.786965654144288,5.305225566711425Q10.855215654144288,5.333494566711426,10.916635654144287,5.374533566711426Q10.978055654144288,5.415572566711425,11.030285654144286,5.467805566711426Q11.082515654144288,5.520038566711426,11.123555654144287,5.581457566711426Q11.164595654144287,5.642877566711426,11.192865654144288,5.711122566711426Q11.221135654144287,5.779368566711426,11.235545654144287,5.851817566711426Q11.249955654144287,5.924267066711426,11.249955654144287,5.998135566711426Q11.249955654144287,6.0720040667114255,11.235545654144287,6.144453566711426Q11.221135654144287,6.2169025667114255,11.192865654144288,6.285148566711426Q11.164595654144287,6.353393566711426,11.123555654144287,6.414813566711426Q11.082515654144288,6.476232566711426,11.030285654144286,6.528465566711426Q10.978055654144288,6.580698566711426,10.916635654144287,6.621737566711426Q10.855215654144288,6.662776566711425,10.786965654144288,6.691045566711426Q10.718725654144286,6.719313566711426,10.646275654144286,6.733724566711426Q10.573825654144287,6.748135566711426,10.499955654144287,6.748135566711426L3.999955654144287,6.748135566711426Q3.926087154144287,6.748135566711426,3.853637654144287,6.733724566711426Q3.781188654144287,6.719313566711426,3.712942654144287,6.691045566711426Q3.644697654144287,6.662776566711425,3.583277654144287,6.621737566711426Q3.521858654144287,6.580698566711426,3.469625654144287,6.528465566711426Q3.417392654144287,6.476232566711426,3.376353654144287,6.414813566711426Q3.335314654144287,6.353393566711426,3.307045654144287,6.285148566711426Q3.278777654144287,6.2169025667114255,3.264366654144287,6.144453566711426Q3.249955654144287,6.0720040667114255,3.249955654144287,5.998135566711426Q3.249955654144287,5.924267066711426,3.264366654144287,5.851817566711426Q3.278777654144287,5.779368566711426,3.307045654144287,5.711122566711426Q3.335314654144287,5.642877566711426,3.376353654144287,5.581457566711426Q3.417392654144287,5.520038566711426,3.469625654144287,5.467805566711426Q3.521858654144287,5.415572566711425,3.583277654144287,5.374533566711426Q3.644697654144287,5.333494566711426,3.712942654144287,5.305225566711425Q3.781188654144287,5.276957566711426,3.853637654144287,5.262546566711426Q3.926087154144287,5.248135566711426,3.999955654144287,5.248135566711426Z" fill-rule="evenodd" />
            </svg>
            <span>Exit</span>
          </button>
        </div>
      </div>
    </Transition>

    <GameDepositSheet v-model="depositSheetVisible" @view-order="viewDepositOrder" />
  </div>
</template>

<style scoped lang="scss">
.game-play {
  position: relative;
  min-height: 100vh;
  min-height: 100dvh;
  max-width: 430px;
  margin: 0 auto;
  overflow: hidden;
  color: #ffffff;
  background: #003f77;
}

.game-play__frame-wrap {
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  padding: 0;
  background: #003f77;
}

.game-play__frame {
  display: block;
  width: 100%;
  height: 100vh;
  height: 100dvh;
  border: 0;
  background: #003f77;
  opacity: 0;
  transition: opacity 180ms ease;
}

.game-play__frame.is-loaded {
  opacity: 1;
}

.game-play__loading {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: grid;
  place-items: center;
  min-height: 100vh;
  min-height: 100dvh;
  background: #003f77;
}

.game-play__frame-wrap.is-covered {
  visibility: hidden;
}

.game-play__loading--overlay {
  position: fixed;
  max-width: 430px;
  margin: 0 auto;
  inset-inline: 0;
}

.game-play__lottie {
  width: min(100%, 375px);
  aspect-ratio: 375 / 158;
}

.game-play__error {
  display: flex;
  width: min(86%, 340px);
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 22px 18px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 14px;
  background: rgba(9, 15, 30, 0.42);
  box-shadow: 0 14px 32px rgba(0, 0, 0, 0.22);
  text-align: center;

  strong {
    font-size: 18px;
    font-weight: 700;
  }

  p {
    max-width: 100%;
    margin: 0;
    color: rgba(255, 255, 255, 0.78);
    font-size: 13px;
    line-height: 1.45;
    overflow-wrap: anywhere;
  }

  button {
    height: 38px;
    min-width: 116px;
    border: 0;
    border-radius: 10px;
    background: linear-gradient(90deg, #3d8aef, #2f69d1);
    color: #ffffff;
    font-size: 14px;
    font-weight: 700;
  }
}

.game-play__quick {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 50001;
  height: 56px;
  color: #ffffff;
  pointer-events: none;
  touch-action: none;
  transition: transform 260ms cubic-bezier(0.22, 1, 0.36, 1);
  will-change: transform;
}

.game-play__quick.is-dragging {
  transition: none;
}

.game-play__quick-toggle {
  position: absolute;
  left: 0;
  top: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border: 0;
  border-radius: 50%;
  background: linear-gradient(310deg, #2e3754 15%, #666e89 84%);
  color: #ffffff;
  box-shadow:
    rgba(135, 146, 164, 0.71) 0 0 4px 0,
    rgba(135, 146, 164, 0.24) 0 0 0 3px,
    rgba(135, 146, 164, 0.42) 0 0 12px 2px;
  opacity: 0.8;
  transform: scale(0.92);
  transform-origin: center;
  pointer-events: auto;
  touch-action: none;
  transition:
    background 220ms ease,
    box-shadow 220ms ease,
    opacity 220ms ease,
    transform 220ms ease;

  &:active {
    transform: scale(0.86);
  }

  svg {
    width: 28px;
    height: 28px;
    flex: 0 0 auto;
    fill: currentColor;
  }
}

.game-play__quick.is-open .game-play__quick-toggle {
  background: #3c445b;
  box-shadow:
    0 0 0 5px rgba(255, 255, 255, 0.08),
    0 0 20px rgba(255, 255, 255, 0.18),
    0 12px 22px rgba(0, 0, 0, 0.52);
  opacity: 0.96;
  transform: scale(0.92);

  svg {
    width: 22px;
    height: 22px;
  }
}

.game-play__quick.is-right .game-play__quick-toggle {
  right: 0;
  left: auto;
}

.game-play__quick-panel {
  position: absolute;
  top: 0;
  left: 66px;
  display: flex;
  align-items: center;
  gap: 16px;
  width: calc(100% - 66px);
  height: 56px;
  overflow: hidden;
  padding: 0 16px;
  border-radius: 12px;
  background: #3c445b;
  box-shadow: 0 12px 26px rgba(0, 0, 0, 0.24);
  opacity: 0;
  clip-path: inset(0 100% 0 0 round 12px);
  transform: translateX(-10px);
  transform-origin: left center;
  pointer-events: none;
  transition:
    clip-path 320ms cubic-bezier(0.22, 1, 0.36, 1),
    opacity 180ms ease,
    transform 320ms cubic-bezier(0.22, 1, 0.36, 1);
}

.game-play__quick.is-open .game-play__quick-panel {
  opacity: 1;
  clip-path: inset(0 0 0 0 round 12px);
  transform: translateX(0);
  pointer-events: auto;
}

.game-play__quick.is-right .game-play__quick-panel {
  right: 66px;
  left: auto;
  clip-path: inset(0 0 0 100% round 12px);
  transform: translateX(10px);
  transform-origin: right center;
}

.game-play__quick.is-right.is-open .game-play__quick-panel {
  clip-path: inset(0 0 0 0 round 12px);
  transform: translateX(0);
}

.game-play__quick-item {
  display: flex;
  width: 40px;
  flex: 0 0 40px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  min-width: 0;
  height: 48px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #ffffff;
  font-size: 10px;
  font-weight: 500;
  line-height: 1;
  text-align: center;
  touch-action: manipulation;
  opacity: 0;
  transform: translateX(-10px) scale(0.96);
  transition:
    opacity 180ms ease,
    transform 240ms cubic-bezier(0.22, 1, 0.36, 1),
    color 160ms ease;
  transition-delay: 0ms;

  &:disabled {
    opacity: 0.55;
  }

  svg {
    width: 22px;
    height: 22px;
    flex: 0 0 auto;
    fill: currentColor;
  }

  span {
    white-space: nowrap;
  }

  &.is-active {
    color: #ef4444;

    svg {
      filter: drop-shadow(0 0 5px rgba(239, 68, 68, 0.35));
    }
  }
}

.game-play__quick.is-open .game-play__quick-item {
  opacity: 1;
  transform: translateX(0) scale(1);
}

.game-play__quick.is-open .game-play__quick-item:nth-child(1) {
  transition-delay: 70ms;
}

.game-play__quick.is-open .game-play__quick-item:nth-child(2) {
  transition-delay: 112ms;
}

.game-play__quick.is-open .game-play__quick-item:nth-child(3) {
  transition-delay: 154ms;
}

.game-play__quick.is-open .game-play__quick-item:nth-child(4) {
  transition-delay: 196ms;
}

.game-play__heart-solid {
  width: 22px !important;
  height: 22px !important;
  fill: currentColor;
}

.game-play-loading-enter-active,
.game-play-loading-leave-active {
  transition: opacity 180ms ease;
}

.game-play-loading-enter-from,
.game-play-loading-leave-to,
.game-play-float-enter-from,
.game-play-float-leave-to {
  opacity: 0;
}

.game-play-float-enter-active,
.game-play-float-leave-active {
  transition: opacity 180ms ease;
}

.game-play-float-enter-from,
.game-play-float-leave-to {
  opacity: 0;
}

</style>
