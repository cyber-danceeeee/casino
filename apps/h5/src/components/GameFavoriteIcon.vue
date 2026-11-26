<script setup lang="ts">
import heart from '../assets/images/heart-gray.webp'

defineProps<{
  active: boolean
}>()
</script>

<template>
  <svg
    v-if="active"
    viewBox="0 0 14.4 12.45"
    aria-hidden="true"
    class="game-favorite-icon game-favorite-icon--active"
  >
    <path d="M0.787 5.919 4.894 11.286c1.182 1.545 3.43 1.545 4.612 0l4.107-5.367c1.049-1.371 1.049-3.32 0-4.691C12.077-.779 8.526-.183 7.2 1.982 5.874-.183 2.323-.779.787 1.228-.262 2.599-.262 4.548.787 5.919Z" />
  </svg>
  <img v-else :src="heart" alt="" aria-hidden="true" class="game-favorite-icon" />
</template>

<style scoped>
.game-favorite-icon {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.game-favorite-icon--active {
  overflow: visible;
  fill: #ef4444;
  filter: drop-shadow(0 0 0.2rem rgba(239, 68, 68, 0.65));
}
</style>
