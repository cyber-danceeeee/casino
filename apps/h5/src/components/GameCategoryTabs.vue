<script setup lang="ts">
import { nextTick, ref, watch, type ComponentPublicInstance } from 'vue'
import type { CategoryKey, GameCategory } from '../mock/games'

const props = defineProps<{
  categories: GameCategory[]
  modelValue: CategoryKey
}>()

const emit = defineEmits<{
  'update:modelValue': [value: CategoryKey]
}>()

const tabRefs = new Map<CategoryKey, HTMLButtonElement>()
const scrollRef = ref<HTMLDivElement | null>(null)
const displayedValue = ref<CategoryKey>(props.modelValue)

const setTabRef = (key: CategoryKey, element: Element | ComponentPublicInstance | null) => {
  if (element instanceof HTMLButtonElement) {
    tabRefs.set(key, element)
    return
  }

  tabRefs.delete(key)
}

const centerActiveTab = async (key = displayedValue.value) => {
  await nextTick()
  const container = scrollRef.value
  const tab = tabRefs.get(key)
  if (!container || !tab) return

  const left = tab.offsetLeft - (container.clientWidth - tab.offsetWidth) / 2
  container.scrollTo({ left: Math.max(0, left), behavior: 'smooth' })
}

const selectCategory = (key: CategoryKey) => {
  displayedValue.value = key
  void centerActiveTab(key)
  emit('update:modelValue', key)
}

watch(
  () => props.modelValue,
  (value) => {
    displayedValue.value = value
    void centerActiveTab(value)
  },
  { immediate: true },
)
</script>

<template>
  <nav class="category-tabs" aria-label="Game categories">
    <div ref="scrollRef" class="category-tabs__scroll">
      <button
        v-for="category in categories"
        :key="category.key"
        type="button"
        class="category-tab"
        :class="{ 'is-active': category.key === displayedValue }"
        :aria-pressed="category.key === displayedValue"
        :ref="(element) => setTabRef(category.key, element)"
        @click="selectCategory(category.key)"
      >
        {{ category.label }}
      </button>
    </div>

    <button class="category-search" type="button" aria-label="Search games">
      <svg viewBox="0 0 20 20" aria-hidden="true">
        <path
          d="M9.1 14.6a5.5 5.5 0 1 0 0-11 5.5 5.5 0 0 0 0 11Zm4.35.3a7.22 7.22 0 1 1 1.46-1.46l2.92 2.92a1 1 0 0 1-1.42 1.42l-2.96-2.88Z"
        />
      </svg>
    </button>
  </nav>
</template>

<style scoped lang="scss">
.category-tabs {
  position: sticky;
  top: 66px;
  z-index: 35;
  display: flex;
  align-items: center;
  margin: 0 0 20px;
  padding: 8px 14px;
  border-top: 1px solid rgba(213, 230, 255, 0.1);
  border-bottom: 1px solid rgba(213, 230, 255, 0.1);
  background: #090f1f;
}

.category-tabs__scroll {
  display: flex;
  flex: 1;
  gap: 28px;
  align-items: center;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.category-tab {
  flex: 0 0 auto;
  min-width: 78px;
  height: 24px;
  padding: 0 12px;
  border-radius: 999px;
  color: #ffffff;
  font-size: 15px;
  font-weight: 700;
  line-height: 1;
  text-align: center;
  text-transform: uppercase;
  transition:
    background 180ms ease,
    color 180ms ease;

  &.is-active {
    background: linear-gradient(180deg, #4d9bff, #2f73de);
  }
}

.category-search {
  display: grid;
  flex: 0 0 40px;
  width: 40px;
  height: 40px;
  place-items: center;
  color: rgba(255, 255, 255, 0.74);

  svg {
    width: 27px;
    height: 27px;
    fill: currentColor;
  }
}

@media (max-width: 374px) {
  .category-tabs {
    padding-right: 10px;
    padding-left: 10px;
  }

  .category-tabs__scroll {
    gap: 20px;
  }

  .category-tab {
    min-width: 70px;
    padding: 0 10px;
    font-size: 14px;
  }
}
</style>
