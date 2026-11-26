<script setup lang="ts">
import GameCard from './GameCard.vue'
import type { GameItem } from '../mock/games'

defineProps<{
  games: GameItem[]
}>()

const emit = defineEmits<{
  play: [game: GameItem]
}>()
</script>

<template>
  <section class="game-section" aria-label="Game list">
    <div class="game-section__head">
      <h1>Featured Games</h1>
      <span>{{ games.length }} live games</span>
    </div>

    <div class="game-grid">
      <GameCard v-for="game in games" :key="game.id" :game="game" @play="emit('play', game)" />
    </div>
  </section>
</template>

<style scoped lang="scss">
.game-section {
  margin-top: 6px;
}

.game-section__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;

  h1 {
    margin: 0;
    color: #ffffff;
    font-size: 17px;
    line-height: 1.2;
  }

  span {
    flex: none;
    color: rgba(224, 233, 255, 0.58);
    font-size: 11px;
  }
}

.game-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}
</style>
