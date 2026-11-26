<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import LottiePlayer from './LottiePlayer.vue'
import type { H5NavigationItem } from '../services/h5Api'
import { openNavigationEntry } from '../services/navigationActions'
import { loadNavigation } from '../services/navigationCache'

const props = defineProps<{
  position: 'HOME_QUICK' | 'PROFILE_MENU'
}>()

const router = useRouter()
const entries = ref<H5NavigationItem[]>([])

onMounted(async () => {
  try {
    entries.value = await loadNavigation(props.position)
  } catch (error) {
    console.warn(`[navigation] failed to load ${props.position}.`, error)
  }
})
</script>

<template>
  <section
    v-if="entries.length"
    class="operation-entries"
    :class="`operation-entries--${position.toLowerCase().replace('_', '-')}`"
  >
    <button
      v-for="entry in entries"
      :key="entry.id"
      type="button"
      class="operation-entry"
      @click="openNavigationEntry(entry, router)"
    >
      <span class="operation-entry__asset" aria-hidden="true">
        <LottiePlayer v-if="entry.lottieURL" :src="entry.lottieURL" />
        <img v-else-if="entry.iconURL" :src="entry.iconURL" alt="" />
        <span v-else class="operation-entry__placeholder">{{ entry.label.slice(0, 1) }}</span>
        <em v-if="entry.badgeType !== 'NONE'">{{ entry.badgeType }}</em>
      </span>
      <span class="operation-entry__label">{{ entry.label }}</span>
      <svg v-if="position === 'PROFILE_MENU'" viewBox="0 0 16 16" aria-hidden="true">
        <path d="m6 4 4 4-4 4" />
      </svg>
    </button>
  </section>
</template>

<style scoped lang="scss">
.operation-entries--home-quick {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px 6px;
  margin: 14px 0 16px;

  .operation-entry {
    display: grid;
    min-width: 0;
    min-height: 72px;
    justify-items: center;
    align-content: start;
    gap: 7px;
    padding: 2px 0;
    color: #e8f0ff;
  }

  .operation-entry__asset {
    width: 44px;
    height: 44px;
  }

  .operation-entry__label {
    width: 100%;
    overflow: hidden;
    font-size: 12px;
    line-height: 1.25;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.operation-entries--profile-menu {
  margin: 12px 0;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: 8px;
  background: rgba(10, 18, 34, 0.76);

  .operation-entry {
    display: grid;
    grid-template-columns: 32px minmax(0, 1fr) 16px;
    align-items: center;
    gap: 11px;
    width: 100%;
    min-height: 54px;
    padding: 8px 14px;
    color: #f2f6ff;
    text-align: left;

    & + .operation-entry {
      border-top: 1px solid rgba(255, 255, 255, 0.06);
    }

    > svg {
      width: 16px;
      height: 16px;
      fill: none;
      stroke: rgba(224, 233, 251, 0.52);
      stroke-linecap: round;
      stroke-linejoin: round;
      stroke-width: 1.7;
    }
  }

  .operation-entry__asset {
    width: 32px;
    height: 32px;
  }

  .operation-entry__label {
    min-width: 0;
    overflow-wrap: anywhere;
    font-size: 14px;
    line-height: 1.3;
  }
}

.operation-entry__asset {
  position: relative;
  display: grid;
  flex: 0 0 auto;
  place-items: center;

  :deep(canvas),
  :deep(svg),
  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  em {
    position: absolute;
    top: -3px;
    right: -9px;
    min-width: 23px;
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
}

.operation-entry__placeholder {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  border-radius: 8px;
  color: #ffffff;
  background: #2475d9;
  font-size: 18px;
  font-weight: 700;
}
</style>
