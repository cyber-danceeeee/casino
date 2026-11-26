<script setup lang="ts">
const digits = '0000500343'.split('')
</script>

<template>
  <section class="rebate-jackpot" aria-label="Rebate Jackpot">
    <div class="rebate-jackpot__inner">
      <div class="rebate-jackpot__title" aria-hidden="true">
        <svg viewBox="0 0 1600 380" preserveAspectRatio="xMinYMid meet">
          <text
            fill="#FFFFFF"
            stroke="#2B41C0"
            stroke-width="56"
            x="28"
            y="190"
            dominant-baseline="middle"
            paint-order="stroke fill"
            stroke-linecap="round"
            stroke-linejoin="round"
            text-anchor="start"
          >
            Rebate jackpot
          </text>
        </svg>
      </div>

      <svg class="rebate-jackpot__bg" viewBox="0 0 347 72" fill="none" aria-hidden="true">
        <path
          d="M0 21C0 16.58 3.58 13 8 13H144H180H272.7C275.91 13 278.81 11.08 280.07 8.12L281.43 4.88C282.69 1.92 285.59 0 288.8 0H339C343.42 0 347 3.58 347 8V64C347 68.42 343.42 72 339 72H8C3.58 72 0 68.42 0 64V21Z"
          fill="url(#jackpot-bg)"
        />
        <defs>
          <linearGradient id="jackpot-bg" x1="173.5" y1="72" x2="173.5" y2="13" gradientUnits="userSpaceOnUse">
            <stop stop-color="#4573D8" />
            <stop offset="0.895" stop-color="#2B41C0" />
          </linearGradient>
        </defs>
      </svg>

      <div class="rebate-jackpot__period">
        <div>
          <span>Monthly</span>
          <span>Weekly</span>
          <span>Daily</span>
        </div>
      </div>

      <div class="rebate-jackpot__digits">
        <span v-for="(digit, index) in digits" :key="`${digit}-${index}`" class="digit">
          <svg viewBox="0 0 100 100" aria-hidden="true">
            <text
              fill="#FFFFFF"
              stroke="#3C4BC8"
              dominant-baseline="middle"
              paint-order="stroke fill"
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="7"
              text-anchor="middle"
              x="50"
              y="56"
            >
              {{ digit }}
            </text>
          </svg>
        </span>
        <span class="digit-dot" />
      </div>
    </div>
  </section>
</template>

<style scoped lang="scss">
.rebate-jackpot {
  margin: 12px 0 10px;
}

.rebate-jackpot__inner {
  position: relative;
  width: 100%;
  overflow: hidden;
  aspect-ratio: 347 / 72;
  cursor: pointer;
}

.rebate-jackpot__title {
  pointer-events: none;
  position: absolute;
  top: -8px;
  left: 4px;
  z-index: 2;
  width: 78%;
  height: 40px;

  svg {
    width: 100%;
    height: 100%;
  }

  text {
    font-family: Montserrat, sans-serif;
    font-size: 170px;
    font-style: italic;
    font-weight: 900;
  }
}

.rebate-jackpot__bg {
  pointer-events: none;
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.rebate-jackpot__period {
  pointer-events: none;
  position: absolute;
  top: 5.5%;
  left: 82.4%;
  z-index: 2;
  width: 15.9%;
  height: 25%;
  overflow: hidden;

  div {
    display: flex;
    height: 300%;
    flex-direction: column;
    animation: period-slide 6s ease-in-out infinite;
  }

  span {
    display: flex;
    height: 33.333%;
    align-items: center;
    justify-content: center;
    color: #ffffff;
    font-size: 12px;
    font-weight: 900;
    line-height: 1;
  }
}

.rebate-jackpot__digits {
  pointer-events: none;
  position: absolute;
  top: 30.6%;
  left: 0;
  z-index: 2;
  display: flex;
  width: 100%;
  height: 69.4%;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.digit {
  position: relative;
  display: grid;
  width: 31.5px;
  height: 40px;
  place-items: center;
  overflow: hidden;
  border-radius: 5px;
  background: linear-gradient(180deg, #647cfd 0%, #68c0ff 100%);
  animation: digit-breathe 2.4s ease-in-out infinite;

  @for $i from 1 through 10 {
    &:nth-child(#{$i}) {
      animation-delay: #{$i * 70}ms;
    }
  }

  svg {
    width: 100%;
    height: 100%;
    filter: drop-shadow(0 0 2px rgba(60, 75, 200, 0.8))
      drop-shadow(0 2px 0 rgba(0, 0, 0, 0.5));
  }

  text {
    font-family: Montserrat, sans-serif;
    font-size: 78px;
    font-weight: 700;
  }
}

.digit-dot {
  position: absolute;
  bottom: -1px;
  left: 80%;
  width: 10px;
  height: 10px;
  border: 2px solid #445fd7;
  border-radius: 50%;
  background: #ffffff;
  transform: translateX(-50%);
}

@keyframes period-slide {
  0%,
  28% {
    transform: translateY(0);
  }

  34%,
  62% {
    transform: translateY(-33.333%);
  }

  68%,
  96% {
    transform: translateY(-66.666%);
  }

  100% {
    transform: translateY(0);
  }
}

@keyframes digit-breathe {
  0%,
  90%,
  100% {
    transform: translateY(0);
  }

  94% {
    transform: translateY(-3px);
  }
}

@media (max-width: 374px) {
  .rebate-jackpot__digits {
    gap: 2px;
  }

  .digit {
    width: 29px;
    height: 37px;
  }

  .rebate-jackpot__period span {
    font-size: 10px;
  }
}
</style>
