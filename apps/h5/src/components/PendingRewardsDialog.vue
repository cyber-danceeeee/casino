<script setup lang="ts">
import { computed, ref } from 'vue'
import type { H5PendingRewardItem } from '../services/h5Api'
import { formatCurrency } from '../services/runtimeConfig'
import baseImage from '../assets/images/home/pending-rewards/base.webp'
import backgroundImage from '../assets/images/home/pending-rewards/bg.webp'
import borderImage from '../assets/images/home/pending-rewards/border.webp'
import ribbonImage from '../assets/images/home/pending-rewards/ribbon.webp'
import claimGiftImage from '../assets/images/home/pending-rewards/claim-gift.webp'
import rewardCoinsImage from '../assets/images/home/reward-float/coins.webp'

const props = defineProps<{
  items: H5PendingRewardItem[]
  claimReward: (rewardID: number) => Promise<unknown>
}>()

const emit = defineEmits<{
  close: []
  claimed: [rewardID: number]
}>()

const claimingID = ref(0)
const totalAmount = computed(() =>
  props.items.reduce((total, item) => total + Number(item.amount || 0), 0),
)

const claim = async (reward: H5PendingRewardItem) => {
  if (claimingID.value) return
  claimingID.value = reward.id
  try {
    await props.claimReward(reward.id)
    emit('claimed', reward.id)
  } finally {
    claimingID.value = 0
  }
}
</script>

<template>
  <Teleport to="body">
    <section class="pending-rewards" role="dialog" aria-modal="true" aria-labelledby="pending-rewards-title">
      <button class="pending-rewards__shade" type="button" aria-label="Close" @click="emit('close')" />
      <article class="pending-rewards__dialog">
        <img class="pending-rewards__background" :src="backgroundImage" alt="" aria-hidden="true" />
        <img class="pending-rewards__border" :src="borderImage" alt="" aria-hidden="true" />
        <img class="pending-rewards__base" :src="baseImage" alt="" aria-hidden="true" />

        <header class="pending-rewards__header">
          <img :src="ribbonImage" alt="" aria-hidden="true" />
          <h2 id="pending-rewards-title">Rewards Unclaimed</h2>
        </header>
        <button class="pending-rewards__close" type="button" aria-label="Close" @click="emit('close')">
          <svg viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M6 6l12 12M18 6 6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          </svg>
        </button>

        <div class="pending-rewards__summary">
          <span>Count <strong>{{ items.length }}</strong></span>
          <span>Total Amount <strong>{{ formatCurrency(totalAmount) }}</strong></span>
        </div>

        <div class="pending-rewards__list">
          <article v-for="reward in items" :key="reward.id" class="pending-reward-card">
            <div class="pending-reward-card__icon">
              <img :src="reward.iconURL || claimGiftImage" :alt="reward.title" @error="($event.currentTarget as HTMLImageElement).src = claimGiftImage" />
            </div>
            <div class="pending-reward-card__copy">
              <strong>{{ reward.title || 'Bonus' }}</strong>
              <span>{{ reward.rewardType === 'bonus' ? 'Bonus Wallet' : 'Cash Wallet' }}</span>
            </div>
            <div class="pending-reward-card__amount">
              <strong>{{ formatCurrency(reward.amount) }}</strong>
              <button type="button" :disabled="Boolean(claimingID)" @click="claim(reward)">
                {{ claimingID === reward.id ? 'Claiming...' : 'Claim' }}
              </button>
            </div>
          </article>

          <div v-if="!items.length" class="pending-rewards__empty">
            <img :src="rewardCoinsImage" alt="" aria-hidden="true" />
            <span>No rewards pending</span>
          </div>
        </div>
      </article>
    </section>
  </Teleport>
</template>

<style scoped lang="scss">
.pending-rewards {
  position: fixed;
  inset: 0;
  z-index: 5900;
  display: grid;
  place-items: center;
  padding: 20px 8px;
  animation: pendingRewardFade 220ms ease-out both;
}

.pending-rewards__shade {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(1, 6, 28, 0.78);
  backdrop-filter: blur(1px);
}

.pending-rewards__dialog {
  position: relative;
  width: min(375px, calc(100vw - 16px));
  height: min(528px, calc(100dvh - 80px));
  min-height: 430px;
  overflow: visible;
  color: #ffffff;
  filter: drop-shadow(0 18px 30px rgba(0, 0, 0, 0.42));
  animation: pendingRewardPop 320ms cubic-bezier(0.2, 0.8, 0.2, 1.08) both;
}

.pending-rewards__background,
.pending-rewards__border,
.pending-rewards__base {
  pointer-events: none;
  position: absolute;
  object-fit: fill;
}

.pending-rewards__background {
  inset: 25px 14px 34px;
  width: calc(100% - 28px);
  height: calc(100% - 59px);
}

.pending-rewards__border {
  inset: 0;
  width: 100%;
  height: 100%;
}

.pending-rewards__base {
  right: -3px;
  bottom: -4px;
  left: -3px;
  width: calc(100% + 6px);
  height: auto;
}

.pending-rewards__header {
  position: absolute;
  top: -20px;
  left: 50%;
  z-index: 3;
  width: 278px;
  transform: translateX(-50%);

  img {
    display: block;
    width: 100%;
    height: 82px;
    object-fit: fill;
  }

  h2 {
    position: absolute;
    inset: 17px 22px auto;
    margin: 0;
    color: #fff4a6;
    font-family: 'Poetsen One', Montserrat, sans-serif;
    font-size: 24px;
    font-weight: 800;
    text-align: center;
    text-shadow: 0 2px 0 #7a2204, 0 3px 8px rgba(84, 22, 0, 0.56);
  }
}

.pending-rewards__close {
  position: absolute;
  top: 24px;
  right: 22px;
  z-index: 4;
  display: grid;
  width: 30px;
  height: 30px;
  place-items: center;
  border: 1px solid rgba(119, 212, 255, 0.64);
  border-radius: 999px;
  color: #ffffff;
  background: #174e9f;

  svg {
    width: 18px;
    height: 18px;
  }
}

.pending-rewards__summary {
  position: absolute;
  top: 77px;
  right: 30px;
  left: 30px;
  z-index: 2;
  display: grid;
  grid-template-columns: 1fr 1.35fr;
  gap: 8px;

  span {
    display: flex;
    min-width: 0;
    align-items: center;
    justify-content: space-between;
    gap: 4px;
    padding: 8px 10px;
    border: 1px solid rgba(114, 194, 255, 0.45);
    border-radius: 7px;
    color: #c8ddff;
    background: rgba(7, 44, 119, 0.78);
    font-size: 11px;
  }

  strong {
    color: #ffe267;
    font-size: 13px;
    white-space: nowrap;
  }
}

.pending-rewards__list {
  position: absolute;
  top: 124px;
  right: 28px;
  bottom: 70px;
  left: 28px;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 9px;
  overflow-y: auto;
  padding: 4px 2px 18px;
  scrollbar-width: none;
}

.pending-rewards__list::-webkit-scrollbar {
  display: none;
}

.pending-reward-card {
  display: grid;
  flex: 0 0 auto;
  grid-template-columns: 52px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  min-height: 74px;
  padding: 9px 10px;
  border: 1px solid rgba(139, 208, 255, 0.74);
  border-radius: 10px;
  background: linear-gradient(180deg, rgba(51, 119, 228, 0.96), rgba(22, 71, 167, 0.98));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.26), 0 5px 10px rgba(0, 20, 77, 0.28);
}

.pending-reward-card__icon {
  display: grid;
  width: 50px;
  height: 50px;
  place-items: center;
  border-radius: 9px;
  background: linear-gradient(180deg, #ffc64f, #ff7a22);

  img {
    width: 44px;
    height: 44px;
    object-fit: contain;
  }
}

.pending-reward-card__copy {
  display: grid;
  min-width: 0;
  gap: 4px;

  strong {
    overflow: hidden;
    font-size: 13px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    color: rgba(255, 255, 255, 0.72);
    font-size: 10px;
  }
}

.pending-reward-card__amount {
  display: grid;
  justify-items: end;
  gap: 6px;

  strong {
    color: #fff171;
    font-size: 13px;
    white-space: nowrap;
  }

  button {
    min-width: 64px;
    height: 27px;
    border: 1px solid #fff1a1;
    border-radius: 999px;
    color: #712900;
    background: linear-gradient(180deg, #fff086, #ffae22);
    font-size: 11px;
    font-weight: 700;
    box-shadow: 0 3px 0 #bd560d;
  }

  button:disabled {
    opacity: 0.65;
  }
}

.pending-rewards__empty {
  display: grid;
  flex: 1;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: #abc8ef;
  font-size: 13px;

  img {
    width: 96px;
    height: 96px;
    object-fit: contain;
    opacity: 0.72;
  }
}

@keyframes pendingRewardFade {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes pendingRewardPop {
  from { opacity: 0; transform: translateY(14px) scale(0.94); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
</style>
