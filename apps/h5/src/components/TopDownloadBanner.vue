<script setup lang="ts">
import { localDialogImage } from '../utils/localAssets'

const emit = defineEmits<{
  close: []
  download: []
}>()

const chest = localDialogImage('top-chests.webp')
</script>

<template>
  <aside class="top-download" aria-label="Download app bonus">
    <div class="top-download__left">
      <button class="top-download__close" type="button" aria-label="Close" @click="emit('close')">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M18 6 6 18M6 6l12 12" />
        </svg>
      </button>

      <img class="top-download__chest" :src="chest" alt="" aria-hidden="true" />

      <p>
        <span>Download the app to claim</span>
        <strong>Bonus</strong>
      </p>
    </div>

    <button class="top-download__button" type="button" @click="emit('download')">Download</button>
  </aside>
</template>

<style scoped lang="scss">
.top-download {
  position: sticky;
  top: 0;
  z-index: 45;
  display: flex;
  width: 100%;
  height: 40px;
  align-items: center;
  justify-content: space-between;
  padding: 0 5px;
  color: #ffffff;
  background: linear-gradient(to top, #11357e, #162f64);
}

.top-download__left {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 4px;
}

.top-download__close {
  display: grid;
  width: 24px;
  height: 24px;
  flex: 0 0 auto;
  place-items: center;
  padding: 0;
  color: #97a5ca;

  svg {
    width: 16px;
    height: 16px;
    stroke: currentColor;
    stroke-linecap: round;
    stroke-width: 2;
  }
}

.top-download__chest {
  width: 24px;
  height: 24px;
  flex: 0 0 auto;
  object-fit: contain;
  animation: top-download-tada 1.5s ease-in-out infinite;
}

.top-download p {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 4px;
  margin: 0;
  white-space: nowrap;

  span {
    min-width: 0;
    overflow: hidden;
    color: rgba(255, 255, 255, 0.9);
    font-size: 12px;
    line-height: 1;
    text-overflow: ellipsis;
  }

  strong {
    background: linear-gradient(180deg, #e3bb0c 22.73%, #f0e54e 47.96%, #c77722 70.45%);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
    font-family: Helvetica, Arial, sans-serif;
    font-size: 16px;
    font-weight: 800;
    line-height: 1;
    text-shadow: 0 4px 12px rgba(225, 182, 13, 0.2);
  }
}

.top-download__button {
  height: 28px;
  flex: 0 0 auto;
  border-radius: 6px;
  padding: 0 12px;
  background: linear-gradient(to right, #3d8aef, #2f69d1);
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  line-height: 1;
}

@keyframes top-download-tada {
  0%,
  100% {
    transform: rotate(0) scale(1);
  }

  10%,
  20% {
    transform: rotate(-8deg) scale(0.96);
  }

  30%,
  50%,
  70%,
  90% {
    transform: rotate(8deg) scale(1.08);
  }

  40%,
  60%,
  80% {
    transform: rotate(-8deg) scale(1.08);
  }
}

@media (max-width: 374px) {
  .top-download p span {
    max-width: 142px;
  }

  .top-download__button {
    padding: 0 10px;
  }
}
</style>
