<script setup lang="ts">
import { computed, ref } from 'vue'
import type { GameItem } from '../mock/games'

const props = defineProps<{
  game: GameItem
}>()

const emit = defineEmits<{
  play: []
}>()

const imageLoaded = ref(false)
const isMaintenance = computed(() => props.game.status === 'maintenance')
</script>

<template>
  <button
    class="game-card"
    type="button"
    :disabled="isMaintenance"
    :aria-label="game.name"
    @click="emit('play')"
  >
    <span class="game-card__cover">
      <span v-if="!imageLoaded" class="game-card__skeleton" aria-hidden="true" />
      <img
        :src="game.cover"
        :alt="game.name"
        loading="lazy"
        decoding="async"
        :class="{ 'is-loaded': imageLoaded }"
        @load="imageLoaded = true"
      />
      <span v-if="game.badge" class="game-card__badge">{{ game.badge }}</span>
      <span v-if="isMaintenance" class="game-card__maintenance">Maintenance</span>
    </span>

    <span class="game-card__body">
      <span class="game-card__name">{{ game.name }}</span>
      <span class="game-card__meta">
        <span>{{ game.provider }}</span>
        <span>{{ game.players }}</span>
      </span>
    </span>
  </button>
</template>

<style scoped lang="scss">
.game-card {
  display: block;
  width: 100%;
  padding: 0;
  overflow: hidden;
  border: 1px solid rgba(205, 225, 255, 0.1);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(31, 41, 67, 0.98), rgba(14, 21, 37, 0.98));
  text-align: left;
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.2);
  transition:
    transform 180ms ease,
    border-color 180ms ease,
    filter 180ms ease;

  &:active:not(:disabled) {
    transform: translateY(1px) scale(0.985);
  }

  &:disabled {
    filter: grayscale(0.35);
  }
}

.game-card__cover {
  position: relative;
  display: block;
  aspect-ratio: 1;
  overflow: hidden;
  background: #111b31;
  contain: paint;

  img {
    position: absolute;
    inset: 0;
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center;
    opacity: 0;
    transform: scale(1.06);
    transform-origin: center;
    transition:
      opacity 260ms ease,
      transform 380ms ease;

    &.is-loaded {
      opacity: 1;
      transform: scale(1.06);
    }
  }
}

.game-card__skeleton {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.07), transparent),
    linear-gradient(135deg, #162238, #223350);
  background-size: 220% 100%, 100% 100%;
  animation: shimmer 1.2s linear infinite;
}

.game-card__badge {
  position: absolute;
  top: 7px;
  left: 7px;
  max-width: calc(100% - 14px);
  padding: 3px 6px;
  overflow: hidden;
  border-radius: 6px;
  background: linear-gradient(180deg, #ffe46b, #d8871e);
  color: #351600;
  font-size: 10px;
  font-weight: 900;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.game-card__maintenance {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  padding: 12px;
  background: rgba(5, 9, 18, 0.7);
  color: rgba(255, 255, 255, 0.9);
  font-size: 12px;
  font-weight: 800;
  text-align: center;
}

.game-card__body {
  display: block;
  padding: 9px 9px 10px;
}

.game-card__name {
  display: block;
  min-width: 0;
  overflow: hidden;
  color: #ffffff;
  font-size: 13px;
  font-weight: 800;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.game-card__meta {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  margin-top: 6px;
  color: rgba(211, 222, 248, 0.58);
  font-size: 11px;
  line-height: 1;

  span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@keyframes shimmer {
  0% {
    background-position: 220% 0, 0 0;
  }

  100% {
    background-position: -220% 0, 0 0;
  }
}
</style>
