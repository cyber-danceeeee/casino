<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import loadingAnimation from '../assets/lottie/lottie-loading.json'
import AppHeader from '../components/AppHeader.vue'
import GameFavoriteIcon from '../components/GameFavoriteIcon.vue'
import LottiePlayer from '../components/LottiePlayer.vue'
import type { CategoryKey, GameCategory, GameItem } from '../mock/games'
import { isLoggedIn } from '../mock/session'
import { h5Api } from '../services/h5Api'
import { getCachedHomeCatalog } from '../services/homeCatalogCache'
import { primeGameAudio } from '../utils/gameAudio'

type SheetKind = 'sort' | 'provider'

interface ProviderOption {
  code: string
  label: string
}

const route = useRoute()
const router = useRouter()

const pageSize = 30
const categories = ref<GameCategory[]>(getCachedHomeCatalog()?.home.categories ?? [])
const games = ref<GameItem[]>([])
const providerOptions = ref<ProviderOption[]>([])
const favoriteMap = ref<Record<string, boolean>>({})
const loading = ref(false)
const loadingMore = ref(false)
const providerLoading = ref(false)
const queryText = ref('')
const keyword = ref('')
const selectedProvider = ref('')
const selectedSort = ref('default')
const favoriteOnly = ref(false)
const page = ref(1)
const total = ref(0)
const sheet = ref<SheetKind | null>(null)
const errorText = ref('')

let loadSeq = 0

const openLogin = () => {
  void router.push({ name: 'auth-login', query: { redirect: route.fullPath } })
}

const openRegister = () => {
  void router.push({ name: 'auth-register', query: { redirect: route.fullPath } })
}

const categoryKey = computed(() => String(route.params.category || route.query.category || 'hot') as CategoryKey)
const category = computed<GameCategory>(() => {
  return categories.value.find((item) => item.key === categoryKey.value) || {
    key: categoryKey.value,
    label: String(categoryKey.value).replace(/-/g, ' '),
    short: String(categoryKey.value).slice(0, 2).toUpperCase(),
  }
})

const hasMore = computed(() => games.value.length < total.value)

const sortOptions = [
  { value: 'default', label: 'Sort by Default' },
  { value: 'name-asc', label: 'Name A-Z' },
  { value: 'name-desc', label: 'Name Z-A' },
  { value: 'provider', label: 'Provider' },
  { value: 'newest', label: 'Newest' },
]

const sortLabel = computed(() => sortOptions.find((item) => item.value === selectedSort.value)?.label || sortOptions[0].label)
const providerLabel = computed(() => {
  if (!selectedProvider.value) return 'All Provider'
  return providerOptions.value.find((item) => item.code === selectedProvider.value)?.label || selectedProvider.value
})
const sheetOptions = computed(() => {
  if (sheet.value === 'sort') {
    return sortOptions.map((item) => ({ value: item.value, label: item.label }))
  }
  return providerOptions.value.map((item) => ({ value: item.code, label: item.label }))
})

const favoriteKey = (game: GameItem) => String(game.gameDbId || game.id)
const isFavorite = (game: GameItem) => favoriteMap.value[favoriteKey(game)] === true

const loadCategories = async () => {
  if (categories.value.length) return
  try {
    const home = await h5Api.getHome()
    categories.value = home.categories
  } catch (error) {
    console.warn('[api] failed to load game categories.', error)
  }
}

const applyFavoriteStatuses = (items: GameItem[]) => {
  favoriteMap.value = {
    ...favoriteMap.value,
    ...Object.fromEntries(items.map((game) => [favoriteKey(game), game.favorite === true])),
  }
}

const loadProviderOptions = async () => {
  providerLoading.value = true
  try {
    const result = await h5Api.getGames({
      category: categoryKey.value,
      page: 1,
      pageSize: 200,
    })
    const providers = new Map<string, string>()
    for (const game of result.games) {
      if (game.providerCode) providers.set(game.providerCode, game.provider || game.providerCode)
    }
    providerOptions.value = [...providers.entries()].map(([code, label]) => ({ code, label }))
    if (selectedProvider.value && !providers.has(selectedProvider.value)) {
      selectedProvider.value = ''
    }
  } catch (error) {
    console.warn('[api] failed to load provider options.', error)
    providerOptions.value = []
  } finally {
    providerLoading.value = false
  }
}

const mergeGames = (current: GameItem[], next: GameItem[]) => {
  const seen = new Set<string>()
  const merged: GameItem[] = []
  for (const game of [...current, ...next]) {
    const key = String(game.gameDbId || `${game.providerCode || ''}:${game.providerGameCode || game.id}`)
    if (seen.has(key)) continue
    seen.add(key)
    merged.push(game)
  }
  return merged
}

const loadGames = async (options: { append?: boolean } = {}) => {
  const append = options.append === true
  if (append) {
    if (loading.value || loadingMore.value || !hasMore.value) return
    loadingMore.value = true
  } else {
    loading.value = true
    errorText.value = ''
  }

  const seq = ++loadSeq
  try {
    const result = await h5Api.getGames({
      category: categoryKey.value,
      page: page.value,
      pageSize,
      keyword: keyword.value,
      provider: selectedProvider.value,
      sort: selectedSort.value === 'default' ? undefined : selectedSort.value,
      favorite: favoriteOnly.value,
    })
    if (seq !== loadSeq) return
    games.value = append ? mergeGames(games.value, result.games) : result.games
    total.value = result.total
    page.value = result.page
    applyFavoriteStatuses(result.games)
  } catch (error) {
    if (seq !== loadSeq) return
    console.warn('[api] failed to load games.', error)
    if (!append) {
      games.value = []
      total.value = 0
      errorText.value = 'Failed to load games'
    }
  } finally {
    if (seq === loadSeq) {
      if (append) {
        loadingMore.value = false
      } else {
        loading.value = false
      }
    }
  }
}

const loadNextPage = () => {
  if (loading.value || loadingMore.value || !hasMore.value) return
  page.value += 1
  void loadGames({ append: true })
}

const onWindowScroll = () => {
  const scrollBottom = window.scrollY + window.innerHeight
  const pageBottom = document.documentElement.scrollHeight
  if (pageBottom - scrollBottom < 360) {
    loadNextPage()
  }
}

const resetForCategory = async () => {
  window.scrollTo({ top: 0, behavior: 'auto' })
  await loadCategories()
  page.value = 1
  total.value = 0
  games.value = []
  queryText.value = ''
  keyword.value = ''
  selectedProvider.value = ''
  selectedSort.value = 'default'
  favoriteOnly.value = false
  await loadProviderOptions()
  await loadGames()
  await nextTick()
  window.scrollTo({ top: 0, behavior: 'auto' })
}

const submitSearch = () => {
  keyword.value = queryText.value.trim()
  page.value = 1
  void loadGames()
}

const selectSort = (value: string) => {
  selectedSort.value = value
  sheet.value = null
  page.value = 1
  void loadGames()
}

const selectProvider = (value: string) => {
  selectedProvider.value = value
  sheet.value = null
  page.value = 1
  void loadGames()
}

const toggleFavoriteOnly = () => {
  if (!isLoggedIn.value) {
    void router.push({ name: 'auth-login', query: { redirect: route.fullPath } })
    return
  }
  favoriteOnly.value = !favoriteOnly.value
  page.value = 1
  void loadGames()
}

const toggleFavorite = async (game: GameItem, event: MouseEvent) => {
  event.stopPropagation()
  if (!isLoggedIn.value) {
    void router.push({ name: 'auth-login', query: { redirect: route.fullPath } })
    return
  }
  if (!game.gameDbId) return

  try {
    const result = await h5Api.toggleGameFavorite(game.gameDbId)
    favoriteMap.value = {
      ...favoriteMap.value,
      [favoriteKey(game)]: result.favorite,
    }
    if (favoriteOnly.value && !result.favorite) {
      void loadGames()
    }
  } catch (error) {
    console.warn('[api] failed to toggle game favorite.', error)
  }
}

const launchGame = (game: GameItem) => {
  if (game.status === 'maintenance') return

  if (!isLoggedIn.value) {
    void router.push({ name: 'auth-login', query: { redirect: route.fullPath } })
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
      favorite: isFavorite(game) ? '1' : '0',
    },
  })
}

const quickStart = () => {
  const playableGames = games.value.filter((item) => item.status !== 'maintenance')
  const game = playableGames[Math.floor(Math.random() * playableGames.length)]
  if (game) launchGame(game)
}

watch(categoryKey, () => {
  void resetForCategory()
}, { immediate: true })

onMounted(() => {
  window.scrollTo({ top: 0, behavior: 'auto' })
  window.addEventListener('scroll', onWindowScroll, { passive: true })
})

onBeforeUnmount(() => {
  loadSeq += 1
  window.removeEventListener('scroll', onWindowScroll)
})
</script>

<template>
  <main class="game-category-page text-white">
    <div class="game-category-page__sticky">
      <AppHeader @login="openLogin" @register="openRegister" />

      <div class="game-category-page__toolbar px-3.5 py-4">
        <div class="flex h-7 items-center">
          <button type="button" class="game-category-page__back" aria-label="Back" @click="router.back()">
            <svg viewBox="0 0 24 24" fill="none" width="24" height="24" class="h-6 w-6 shrink-0 cursor-pointer text-white">
              <path d="M15 18l-6-6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
          <div v-if="category.icon" class="relative overflow-hidden h-6 w-6 ml-2 shrink-0 rounded-sm">
            <img alt="" width="24" height="24" class="h-full w-full object-cover transition-opacity duration-200 opacity-100" draggable="false" :src="category.icon" />
          </div>
          <span class="ml-2 text-lg font-600 uppercase">{{ category.label }}</span>
          <button type="button" class="ml-auto shrink-0 cursor-pointer border-none bg-transparent p-0 text-sm text-[hsl(var(--primary))] disabled:opacity-50" :disabled="loading || !games.length" @click="quickStart">
            Quick Start
          </button>
        </div>

        <div class="mt-4 flex items-center">
          <form class="relative flex-1" @submit.prevent="submitSearch">
            <svg viewBox="0 0 24 24" fill="none" width="20" height="20" class="pointer-events-none absolute left-3.5 top-1/2 h-5 w-5 text-white/40 -translate-y-1/2">
              <path d="M21 21l-4.35-4.35M11 19a8 8 0 100-16 8 8 0 000 16z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
            <input
              v-model="queryText"
              aria-label="Search games"
              class="h-10 w-full rounded-md bg-white/8 pl-12 pr-3 text-sm text-white focus-visible:outline-none"
              @keyup.enter="submitSearch"
            />
          </form>
          <button class="ml-2.5 h-10 w-10 flex cursor-pointer items-center justify-center rounded-md bg-f4" type="button" :class="{ 'is-favorite-active': favoriteOnly }" @click="toggleFavoriteOnly">
            <svg v-if="favoriteOnly" xmlns="http://www.w3.org/2000/svg" fill="currentColor" width="14.400019645690918" height="12.444425582885742" viewBox="0 0 14.400019645690918 12.444425582885742" class="h-4 w-4 text-red-500">
              <path d="M0.787112,5.91933L4.89438,11.2861C6.07639,12.8305,8.3236,12.8305,9.5056,11.2861L13.6129,5.91934C14.6624,4.54802,14.6624,2.59947,13.6129,1.22816C12.0766,-0.779254,8.52568,-0.183052,7.2,1.98221C5.87431,-0.183052,2.32342,-0.779254,0.78712,1.22816C-0.262371,2.59947,-0.262373,4.54802,0.787112,5.91933Z" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" fill="currentColor" width="14.400019645690918" height="12.444425582885742" viewBox="0 0 14.400019645690918 12.444425582885742" class="h-4 w-4 text-tf3">
              <path d="M0.787112,5.91933L4.89438,11.2861C6.07639,12.8305,8.3236,12.8305,9.5056,11.2861L13.6129,5.91934C14.6624,4.54802,14.6624,2.59947,13.6129,1.22816C12.0766,-0.779254,8.52568,-0.183052,7.2,1.98221C5.87431,-0.183052,2.32342,-0.779254,0.78712,1.22816C-0.262371,2.59947,-0.262373,4.54802,0.787112,5.91933Z" />
            </svg>
          </button>
        </div>

        <div class="mt-2.5 flex items-center gap-2.5">
          <button type="button" class="game-category-page__select" @click="sheet = 'sort'">
            <span class="truncate">{{ sortLabel }}</span>
            <svg class="h-3 w-3 shrink-0 text-white/50" viewBox="0 0 12 12" fill="none">
              <path d="M3 4.5L6 7.5L9 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
          <button type="button" class="game-category-page__select" @click="sheet = 'provider'">
            <span class="truncate">{{ providerLoading ? 'Loading...' : providerLabel }}</span>
            <svg class="h-3 w-3 shrink-0 text-white/50" viewBox="0 0 12 12" fill="none">
              <path d="M3 4.5L6 7.5L9 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div class="relative">
      <div class="absolute left-0 right-0 flex items-center justify-center" style="top: -40px; height: 40px;"></div>
      <div v-if="loading" class="game-category-page__loading" aria-label="Loading">
        <div class="game-category-page__loading-stage">
          <LottiePlayer :animation-data="loadingAnimation" :loop="true" :autoplay="true" />
        </div>
      </div>
      <div v-else-if="errorText" class="game-category-page__state">{{ errorText }}</div>
      <template v-else>
        <div class="game-category-page__grid">
          <div v-for="game in games" :key="game.id" class="game-category-page__card">
            <div role="button" tabindex="0" @click="launchGame(game)" @keydown.enter.prevent="launchGame(game)">
              <div class="game-category-page__cover">
                <div class="relative overflow-hidden h-full w-full">
                  <img
                    :alt="game.name"
                    width="160"
                    loading="lazy"
                    class="game-category-page__image"
                    draggable="false"
                    :src="game.cover"
                  />
                </div>
                <span v-if="game.provider" class="game-category-page__provider">{{ game.provider }}</span>
                <button class="game-category-page__heart" type="button" @click="toggleFavorite(game, $event)">
                  <GameFavoriteIcon :active="isFavorite(game)" />
                </button>
                <div v-if="game.status === 'maintenance'" class="game-category-page__mask">Maintenance</div>
              </div>
            </div>
            <div class="mt-2 truncate text-xs text-white">{{ game.name }}</div>
          </div>
        </div>

        <div v-if="loadingMore" class="py-4 text-center text-xs text-gray-400">Loading...</div>
        <div v-else-if="!hasMore" class="py-4 pb-safe text-center text-xs text-gray-400">No more results</div>
      </template>
    </div>

    <div v-if="sheet" class="game-category-sheet" @click.self="sheet = null">
      <div class="game-category-sheet__panel">
        <div class="game-category-sheet__title">{{ sheet === 'sort' ? 'Sort' : 'Provider' }}</div>
        <button
          v-if="sheet === 'provider'"
          type="button"
          class="game-category-sheet__option"
          :class="{ 'is-active': selectedProvider === '' }"
          @click="selectProvider('')"
        >
          <span>All Provider</span>
          <span v-if="selectedProvider === ''">✓</span>
        </button>
        <button
          v-for="option in sheetOptions"
          :key="option.value"
          type="button"
          class="game-category-sheet__option"
          :class="{ 'is-active': sheet === 'sort' ? selectedSort === option.value : selectedProvider === option.value }"
          @click="sheet === 'sort' ? selectSort(option.value) : selectProvider(option.value)"
        >
          <span>{{ option.label }}</span>
          <span v-if="sheet === 'sort' ? selectedSort === option.value : selectedProvider === option.value">✓</span>
        </button>
      </div>
    </div>
  </main>
</template>

<style scoped lang="scss">
.game-category-page {
  width: 100%;
  min-height: 100vh;
  max-width: 430px;
  margin: 0 auto;
  overflow-anchor: none;
  background: #090f1f;
  color: #ffffff;
}

.game-category-page__sticky {
  position: sticky;
  top: 0;
  z-index: 60;
  background: #090f1f;
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.08);
}

.game-category-page__sticky :deep(.app-header) {
  position: relative;
  top: auto;
  z-index: auto;
}

.game-category-page__toolbar {
  background: #090f1f;
}

.game-category-page__back {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.5rem;
  height: 1.5rem;
}

.game-category-page__select {
  display: flex;
  flex: 1;
  height: 2.5rem;
  cursor: pointer;
  align-items: center;
  justify-content: space-between;
  border: none;
  border-radius: 0.375rem;
  background: rgba(255, 255, 255, 0.08);
  padding: 0 0.75rem;
  color: #ffffff;
  font-size: 0.875rem;
}

.game-category-page__mask {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgba(5, 9, 18, 0.68);
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 800;
  text-align: center;
}

.game-category-page__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.875rem;
  padding: 0 0.875rem;
}

.game-category-page__card {
  min-width: 0;
}

.game-category-page__cover {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  overflow: hidden;
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  contain: paint;
}

.game-category-page__image {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: fill;
  object-position: center;
  opacity: 1;
  transition: opacity 200ms ease;
}

.game-category-page__heart {
  position: absolute;
  top: 0.375rem;
  right: 0.375rem;
  z-index: 2;
  width: 1.125rem;
  height: 1.125rem;
  padding: 0;
  cursor: pointer;
  border: 0;
  background: transparent;
}

.game-category-page__provider {
  position: absolute;
  top: 0.375rem;
  left: 0.375rem;
  z-index: 2;
  max-width: calc(100% - 2.5rem);
  padding: 0.125rem 0.3125rem;
  overflow: hidden;
  border-radius: 0.25rem;
  background: rgba(9, 15, 30, 0.76);
  color: #ffffff;
  font-size: 0.625rem;
  font-weight: 600;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.game-category-page__state {
  display: grid;
  min-height: 13.75rem;
  place-items: center;
  color: rgba(255, 255, 255, 0.58);
  font-size: 0.875rem;
}

.game-category-page__loading {
  display: grid;
  min-height: 13.75rem;
  place-items: center;
  overflow: hidden;
  background: #090f1f;
}

@media (max-width: 374px) {
  .game-category-page__grid {
    gap: 0.625rem;
    padding-right: 0.625rem;
    padding-left: 0.625rem;
  }
}

.game-category-page__loading-stage {
  width: min(100%, 23.4375rem);
  aspect-ratio: 375 / 158;
}

.game-category-sheet {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(0, 0, 0, 0.52);
}

.game-category-sheet__panel {
  width: 100%;
  max-width: 430px;
  max-height: 70vh;
  overflow-y: auto;
  border-radius: 1rem 1rem 0 0;
  background: #1e2438;
  padding: 0.75rem 0 1rem;
  box-shadow: 0 -12px 34px rgba(0, 0, 0, 0.35);
}

.game-category-sheet__title {
  padding: 0.5rem 1rem 0.75rem;
  color: #ffffff;
  font-size: 1rem;
  font-weight: 700;
  text-align: center;
}

.game-category-sheet__option {
  display: flex;
  width: 100%;
  min-height: 3rem;
  align-items: center;
  justify-content: space-between;
  border: none;
  background: transparent;
  padding: 0 1.25rem;
  color: rgba(255, 255, 255, 0.68);
  font-size: 0.9375rem;
  text-align: left;

  &.is-active {
    color: #3d8aef;
  }
}
</style>
