<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, watch, type ComponentPublicInstance } from 'vue'
import type { GameItem, GameSection } from '../mock/games'
import GameFavoriteIcon from './GameFavoriteIcon.vue'

const props = defineProps<{
  sections: GameSection[]
}>()
const visibleSections = computed(() => props.sections.filter((section, index) => index === 0 || section.games.length > 0))

const emit = defineEmits<{
  play: [game: GameItem]
  viewAll: [section: GameSection]
  page: [section: GameSection, page: number]
}>()

const isHotSection = (section: GameSection) => String(section.key).toLowerCase() === 'hot'
const pageViewportRefs = new Map<string, HTMLElement>()
const pageScrollTimers = new Map<string, number>()
const syncedPages = new Map<string, number>()

const pageCount = (section: GameSection) => {
  const total = section.total ?? section.games.length
  const pageSize = section.pageSize || 9
  return Math.max(1, Math.min(3, Math.ceil(total / pageSize)))
}

const currentPage = (section: GameSection) => section.page || 1

const loadedPageNumbers = (section: GameSection) => {
  const pages = Object.keys(section.pages || {})
    .map(Number)
    .filter((page) => Number.isInteger(page) && page >= 1 && page <= pageCount(section))
    .sort((left, right) => left - right)

  return pages.length ? pages : [currentPage(section)]
}

const gamesForPage = (section: GameSection, page: number) => {
  return section.pages?.[page] ?? (page === currentPage(section) ? section.games : [])
}

const syncSectionScroll = (section: GameSection, behavior: ScrollBehavior = 'smooth') => {
  const viewport = pageViewportRefs.get(String(section.key))
  const page = currentPage(section)
  if (!viewport || syncedPages.get(String(section.key)) === page) return

  const index = loadedPageNumbers(section).indexOf(page)
  if (index < 0) return

  syncedPages.set(String(section.key), page)
  viewport.scrollTo({ left: index * viewport.clientWidth, behavior })
}

const setPageViewportRef = (
  section: GameSection,
  element: Element | ComponentPublicInstance | null,
) => {
  const key = String(section.key)
  if (element instanceof HTMLElement) {
    pageViewportRefs.set(key, element)
    void nextTick(() => syncSectionScroll(section, 'auto'))
    return
  }

  pageViewportRefs.delete(key)
  syncedPages.delete(key)
}

const settlePagedScroll = (section: GameSection, viewport: HTMLElement) => {
  const key = String(section.key)
  const pages = loadedPageNumbers(section)
  const index = Math.max(0, Math.min(pages.length - 1, Math.round(viewport.scrollLeft / viewport.clientWidth)))
  const page = pages[index]
  if (!page || page === currentPage(section)) return

  syncedPages.set(key, page)
  emit('page', section, page)
}

const onPagedScroll = (section: GameSection, event: Event) => {
  const viewport = event.currentTarget as HTMLElement
  const key = String(section.key)
  const activeTimer = pageScrollTimers.get(key)
  if (activeTimer) window.clearTimeout(activeTimer)

  pageScrollTimers.set(key, window.setTimeout(() => {
    pageScrollTimers.delete(key)
    settlePagedScroll(section, viewport)
  }, 120))
}

const goPrev = (section: GameSection) => {
  const page = currentPage(section)
  if (page > 1) emit('page', section, page - 1)
}

const goNext = (section: GameSection) => {
  const page = currentPage(section)
  if (page < pageCount(section)) emit('page', section, page + 1)
}

watch(
  () => props.sections.map((section) => `${section.key}:${currentPage(section)}:${loadedPageNumbers(section).join(',')}`).join('|'),
  async () => {
    await nextTick()
    for (const section of props.sections) {
      if (!isHotSection(section)) syncSectionScroll(section)
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  for (const timer of pageScrollTimers.values()) window.clearTimeout(timer)
  pageScrollTimers.clear()
  pageViewportRefs.clear()
  syncedPages.clear()
})
</script>

<template>
  <section class="game-sections" aria-label="Game sections">
    <article
      v-for="section in visibleSections"
      :id="`section-${section.key}`"
      :key="section.key"
      class="game-section"
      :class="{ 'game-section--hot': isHotSection(section) }"
    >
      <button class="game-section__head" type="button" @click="emit('viewAll', section)">
        <div v-if="section.icon" class="game-section__icon">
          <img
            width="28"
            height="28"
            class="h-full w-full object-cover transition-opacity duration-200 opacity-100"
            :src="section.icon"
            alt=""
            draggable="false"
          />
        </div>
        <span class="text-lg font-600 uppercase">{{ section.title }}</span>
      </button>

      <div v-if="isHotSection(section)" class="game-section__viewport game-section__viewport--hot">
        <div class="game-section__grid game-section__grid--hot">
          <div
            v-for="game in section.games.slice(0, 12)"
            :key="`${section.key}-${game.id}`"
            class="game-section__card"
          >
            <button
              class="game-section__cover"
              type="button"
              :disabled="game.status === 'maintenance'"
              :aria-label="game.name"
              @click="emit('play', game)"
            >
              <img class="game-section__image" :src="game.cover" :alt="game.name" loading="lazy" />
              <span v-if="game.provider" class="game-section__provider">{{ game.provider }}</span>
              <span class="game-section__heart">
                <GameFavoriteIcon :active="game.favorite === true" />
              </span>
              <span v-if="game.status === 'maintenance'" class="game-section__mask">Maintenance</span>
            </button>
          </div>
        </div>
      </div>

      <div
        v-else
        :ref="(element) => setPageViewportRef(section, element)"
        class="game-section__viewport game-section__viewport--paged"
        @scroll.passive="onPagedScroll(section, $event)"
      >
        <div class="game-section__pages">
          <div
            v-for="pageNumber in loadedPageNumbers(section)"
            :key="`${section.key}-page-${pageNumber}`"
            class="game-section__page"
          >
            <div class="game-section__grid">
              <div
                v-for="game in gamesForPage(section, pageNumber)"
                :key="`${section.key}-${pageNumber}-${game.id}`"
                class="game-section__card"
              >
                <button
                  class="game-section__cover"
                  type="button"
                  :disabled="game.status === 'maintenance'"
                  :aria-label="game.name"
                  @click="emit('play', game)"
                >
                  <img class="game-section__image" :src="game.cover" :alt="game.name" loading="lazy" />
                  <span v-if="game.provider" class="game-section__provider">{{ game.provider }}</span>
                  <span class="game-section__heart">
                    <GameFavoriteIcon :active="game.favorite === true" />
                  </span>
                  <span v-if="game.status === 'maintenance'" class="game-section__mask">Maintenance</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="!isHotSection(section) && (section.total ?? section.games.length) > 9"
        class="game-section__pager"
      >
        <button aria-label="Prev" class="game-section__page-btn game-section__page-btn--prev" type="button" :disabled="currentPage(section) <= 1" @click="goPrev(section)">
          <svg class="h-4 w-4 rotate-180" viewBox="0 0 24 24" fill="none">
            <path d="M9 6l6 6-6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
        <button class="game-section__view-all" type="button" @click="emit('viewAll', section)">
          View All {{ section.total ?? section.games.length }} Games
        </button>
        <button aria-label="Next" class="game-section__page-btn game-section__page-btn--next" type="button" :disabled="currentPage(section) >= pageCount(section)" @click="goNext(section)">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none">
            <path d="M9 6l6 6-6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
      </div>
    </article>
  </section>
</template>

<style scoped lang="scss">
.game-sections {
  display: grid;
  gap: 24px;
  width: 100%;
  min-width: 0;
  padding-bottom: 18px;
}

.game-section {
  min-width: 0;
}

.game-section__head {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 0 0.875rem;
  color: #ffffff;
  text-align: left;
}

.game-section__icon {
  position: relative;
  flex: 0 0 1.75rem;
  width: 1.75rem;
  height: 1.75rem;
  margin-right: 0.5rem;
  overflow: hidden;
  border-radius: 0.25rem;
}

.game-section__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.875rem;
  padding: 0.75rem 0.875rem 0;
}

.game-section__viewport--paged {
  width: 100%;
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.game-section__pages {
  display: flex;
  width: 100%;
  align-items: flex-start;
}

.game-section__page {
  flex: 0 0 100%;
  width: 100%;
  scroll-snap-align: start;
  scroll-snap-stop: always;
}

.game-section__viewport--hot {
  --game-card-width: calc((min(100vw, 26.875rem) - 3.5rem) / 3.5);

  padding: 0.75rem 0.875rem 0;
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  scroll-padding-inline: 0.875rem;
  scroll-snap-type: x proximity;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.game-section__grid--hot {
  grid-template-columns: repeat(6, var(--game-card-width));
  grid-template-rows: repeat(2, minmax(0, auto));
  grid-auto-flow: row;
  width: max-content;
  padding: 0;
}

.game-section__grid--hot .game-section__card {
  width: var(--game-card-width);
  scroll-snap-align: start;
}

.game-section__card {
  min-width: 0;
}

.game-section__cover {
  position: relative;
  display: block;
  width: 100%;
  aspect-ratio: 1;
  padding: 0;
  overflow: hidden;
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  contain: paint;
}

.game-section__cover:active:not(:disabled) {
  transform: scale(0.985);
}

.game-section__cover:disabled {
  filter: grayscale(0.35);
}

.game-section__image {
  position: absolute;
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
  object-fit: fill;
  object-position: center;
}

.game-section__heart {
  position: absolute;
  top: 0.375rem;
  right: 0.375rem;
  z-index: 2;
  width: 1.125rem;
  height: 1.125rem;
}

.game-section__provider {
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

.game-section__mask {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 0.5rem;
  background: rgba(5, 9, 18, 0.68);
  color: #ffffff;
  font-size: 0.6875rem;
  font-weight: 800;
  text-align: center;
}

.game-section__pager {
  display: flex;
  height: 2.125rem;
  gap: 0.5rem;
  padding: 0 0.875rem;
  margin-top: 1rem;
}

.game-section__page-btn {
  display: flex;
  width: 2.75rem;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  background: linear-gradient(118deg,#3d8aef,#2f69d1);
  color: #ffffff;
}

.game-section__page-btn:disabled {
  cursor: default;
  opacity: 0.45;
}

.game-section__page-btn--prev {
  border-radius: 999px 0 0 999px;
}

.game-section__page-btn--next {
  border-radius: 0 999px 999px 0;
}

.game-section__view-all {
  display: flex;
  flex: 1;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  border-radius: 0.125rem;
  background: hsl(131 36% 52% / 1);
  color: #ffffff;
  font-size: 0.875rem;
  line-height: 1;
}

@media (max-width: 374px) {
  .game-section__viewport--hot {
    --game-card-width: calc((100vw - 2.5rem) / 3);
  }

  .game-section__head,
  .game-section__grid,
  .game-section__pager {
    padding-right: 0.625rem;
    padding-left: 0.625rem;
  }

  .game-section__grid {
    gap: 0.625rem;
  }

  .game-section__viewport--hot {
    padding-right: 0.625rem;
    padding-left: 0.625rem;
    scroll-padding-inline: 0.625rem;
  }
}
</style>
