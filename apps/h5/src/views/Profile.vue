<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import gsap from 'gsap'
import ActivityGuideFlow from '../components/ActivityGuideFlow.vue'
import OperationEntries from '../components/OperationEntries.vue'
import { formatPakistanPhone, isLoggedIn, mockUser, refreshMockUser } from '../mock/session'
import {
  accountMenuItems,
  profileAssets,
  supportMenuItems,
  type ProfileMenuItem,
} from '../mock/profile'
import { currencySymbol } from '../services/runtimeConfig'
import { isAndroidAppRuntime } from '../utils/runtimeEnvironment'

const router = useRouter()
const androidAppRuntime = isAndroidAppRuntime()
const menuGroups = computed<ProfileMenuItem[][]>(() => [
  androidAppRuntime ? accountMenuItems.filter((item) => item.action !== 'download') : accountMenuItems,
  supportMenuItems,
])
const vipJumpEl = ref<HTMLImageElement | null>(null)
const guideVisible = ref(false)
const balanceRefreshing = ref(false)

let vipJumpTween: gsap.core.Tween | null = null
const displayPhone = computed(() => (isLoggedIn.value ? formatPakistanPhone() : '--'))
const displayUserId = computed(() => (isLoggedIn.value ? mockUser.value.userId : '--'))
const cashBalance = computed(() => (isLoggedIn.value ? mockUser.value.cash : '0.00'))
const bonusBalance = computed(() => (isLoggedIn.value ? mockUser.value.bonus : '0.00'))

onMounted(() => {
  void refreshProfileBalance()

  if (vipJumpEl.value) {
    gsap.set(vipJumpEl.value, { x: 0 })
    vipJumpTween = gsap.to(vipJumpEl.value, {
      x: 4,
      duration: 0.38,
      ease: 'sine.inOut',
      repeat: -1,
      yoyo: true,
    })
  }
})

onBeforeUnmount(() => {
  vipJumpTween?.kill()
})

const openVipOffers = () => {
  void router.push({ name: 'vip' })
}

const openRoute = (name: string) => {
  void router.push({ name })
}

const refreshProfileBalance = async () => {
  if (!isLoggedIn.value || balanceRefreshing.value) return
  balanceRefreshing.value = true
  try {
    await refreshMockUser()
  } catch (error) {
    console.warn('[api] failed to refresh profile wallet.', error)
  } finally {
    balanceRefreshing.value = false
  }
}

const openMenuItem = (item: ProfileMenuItem) => {
  if (item.action === 'download') {
    if (androidAppRuntime) return
    guideVisible.value = true
    return
  }

  if (item.externalUrl) {
    window.location.href = item.externalUrl
    return
  }

  if (item.routeName) openRoute(item.routeName)
}
</script>

<template>
  <div class="profile-shell">
    <img class="profile-shell__bg" :src="profileAssets.background" alt="" aria-hidden="true" />

    <main class="profile-content" aria-label="Profile">
      <section class="profile-user">
        <div class="profile-user__avatar">
          <div>
            <img :src="profileAssets.avatar" alt="" />
          </div>
        </div>

        <div class="profile-user__info">
          <div class="profile-user__name">
            <strong>{{ displayPhone }}</strong>
            <button type="button" aria-label="Verify phone number" @click="openRoute('security')">
              <img :src="profileAssets.gift" alt="" aria-hidden="true" />
            </button>
          </div>
          <span class="profile-user__id">
            User ID: {{ displayUserId }}
            <svg viewBox="0 0 16 16" aria-hidden="true">
              <path d="M5 2.5A2.5 2.5 0 0 1 7.5 0h4A2.5 2.5 0 0 1 14 2.5v4A2.5 2.5 0 0 1 11.5 9h-4A2.5 2.5 0 0 1 5 6.5v-4Zm2.5-1A1 1 0 0 0 6.5 2.5v4a1 1 0 0 0 1 1h4a1 1 0 0 0 1-1v-4a1 1 0 0 0-1-1h-4ZM2 7a1 1 0 0 1 1 1v4.5A1.5 1.5 0 0 0 4.5 14H9a1 1 0 1 1 0 2H4.5A3.5 3.5 0 0 1 1 12.5V8a1 1 0 0 1 1-1Z" />
            </svg>
          </span>
          <button type="button" class="profile-vip" aria-label="VIP 0" @click="openVipOffers">
            <img class="profile-vip__badge" :src="profileAssets.vip" alt="VIP 0" />
            <img
              ref="vipJumpEl"
              class="profile-vip__jump"
              :src="profileAssets.jump"
              alt=""
              aria-hidden="true"
            />
          </button>
        </div>
      </section>

      <section class="wallet-card" aria-label="Cash Wallet">
        <img class="wallet-card__bg" :src="profileAssets.walletBg" alt="" aria-hidden="true" />
        <button type="button" class="wallet-card__records" @click="router.push({ name: 'wallet-records', query: { tab: 'Statement' } })">
          <svg viewBox="0 0 14 14" aria-hidden="true">
            <path
              d="M3.24 0h7.52c.46 0 .9.19 1.23.54.33.35.51.82.51 1.31v9.84c0 .49-.18.96-.51 1.31-.33.35-.77.54-1.23.54H3.24c-.46 0-.9-.19-1.23-.54a1.88 1.88 0 0 1-.51-1.31V1.85c0-.49.18-.96.51-1.31C2.34.19 2.78 0 3.24 0Z"
            />
            <path
              d="M10.37 9.27a.44.44 0 1 1 0 .88H3.63a.44.44 0 0 1 0-.88h6.74Zm0-2.94a.44.44 0 1 1 0 .88H3.63a.44.44 0 0 1 0-.88h6.74ZM7.44 3.38a.44.44 0 1 1 0 .89H3.63a.44.44 0 0 1 0-.89h3.81Z"
            />
          </svg>
          <span>Records</span>
          <svg class="wallet-card__chevron" viewBox="0 0 16 16" aria-hidden="true">
            <path d="m6 4 4 4-4 4" />
          </svg>
        </button>

        <div class="wallet-card__content">
          <img class="wallet-card__icon" :src="profileAssets.cashWallet" alt="" aria-hidden="true" />
          <div class="wallet-card__main">
            <span>Cash Wallet</span>
            <div class="wallet-card__amount">
              <span>{{ currencySymbol() }}</span>
              <strong>{{ cashBalance }}</strong>
              <button type="button" aria-label="Refresh cash balance" @click="refreshProfileBalance">
                <svg viewBox="0 0 18 18" aria-hidden="true">
                  <path
                    d="M14.75 3a1 1 0 1 1 2 0v6a1 1 0 1 1-2 0 5.75 5.75 0 0 0-9.89-3.99 1 1 0 0 1-1.44-1.39A7.74 7.74 0 0 1 9 1.25c4.28 0 7.75 3.47 7.75 7.75a1 1 0 1 1-2 0V3ZM3.25 15a1 1 0 1 1-2 0V9a1 1 0 1 1 2 0 5.75 5.75 0 0 0 9.74 4.14 1 1 0 0 1 1.39 1.44A7.72 7.72 0 0 1 9 16.75 7.75 7.75 0 0 1 1.25 9a1 1 0 1 1 2 0v6Z"
                  />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <div class="wallet-card__actions">
          <button type="button" @click="openRoute('wallet-deposit')">Deposit</button>
          <button type="button" @click="openRoute('cash-center')">Withdraw</button>
        </div>
      </section>

      <section class="bonus-wallet" aria-label="Bonus Wallet">
        <header>
          <div>
            <img :src="profileAssets.bonusWallet" alt="" aria-hidden="true" />
            <span>Bonus Wallet</span>
          </div>
          <button type="button" @click="openRoute('bonus-center')">
            Bonus Center
            <svg viewBox="0 0 16 16" aria-hidden="true">
              <path d="m6 4 4 4-4 4" />
            </svg>
          </button>
        </header>
        <div class="bonus-wallet__amount">
          <div>
            <span>{{ currencySymbol() }}</span>
            <strong>{{ bonusBalance }}</strong>
          </div>
          <button type="button" aria-label="Refresh bonus balance" @click="refreshProfileBalance">
            <svg viewBox="0 0 18 18" aria-hidden="true">
              <path
                d="M14.75 3a1 1 0 1 1 2 0v6a1 1 0 1 1-2 0 5.75 5.75 0 0 0-9.89-3.99 1 1 0 0 1-1.44-1.39A7.74 7.74 0 0 1 9 1.25c4.28 0 7.75 3.47 7.75 7.75a1 1 0 1 1-2 0V3ZM3.25 15a1 1 0 1 1-2 0V9a1 1 0 1 1 2 0 5.75 5.75 0 0 0 9.74 4.14 1 1 0 0 1 1.39 1.44A7.72 7.72 0 0 1 9 16.75 7.75 7.75 0 0 1 1.25 9a1 1 0 1 1 2 0v6Z"
              />
            </svg>
          </button>
        </div>
      </section>

      <OperationEntries position="PROFILE_MENU" />

      <section v-for="(group, groupIndex) in menuGroups" :key="groupIndex" class="profile-menu">
        <button
          v-for="item in group"
          :key="item.label"
          type="button"
          class="profile-menu__item"
          @click="openMenuItem(item)"
        >
          <img :src="item.icon" alt="" aria-hidden="true" />
          <span>{{ item.label }}</span>
          <i v-if="item.badge" class="profile-menu__badge" aria-hidden="true" />
          <svg viewBox="0 0 16 16" aria-hidden="true">
            <path d="m6 4 4 4-4 4" />
          </svg>
        </button>
      </section>
    </main>

    <ActivityGuideFlow
      v-if="guideVisible && !androidAppRuntime"
      :visible="guideVisible"
      initial="app-download"
      @complete="guideVisible = false"
    />
  </div>
</template>

<style scoped lang="scss">
.profile-shell {
  --bottom-nav-height: 76px;

  position: relative;
  height: 100vh;
  height: 100dvh;
  max-width: 430px;
  margin: 0 auto;
  overflow: hidden;
  color: #ffffff;
  background: #090f1e;
}

.profile-shell__bg {
  pointer-events: none;
  position: absolute;
  inset: 0 0 auto;
  width: 100%;
  aspect-ratio: 1500 / 1352;
  object-fit: contain;
  opacity: 0.8;
}

.profile-content {
  position: relative;
  z-index: 1;
  height: 100%;
  overflow-x: hidden;
  overflow-y: auto;
  overscroll-behavior: none;
  padding: max(env(safe-area-inset-top, 0px), 20px) 14px
    calc(var(--bottom-nav-height) + env(safe-area-inset-bottom) + 20px);
}

.profile-user {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding-bottom: 20px;
}

.profile-user__avatar {
  flex: 0 0 auto;
  overflow: hidden;
  border: 2px solid #3b82f6;
  border-radius: 999px;
  padding: 2px;
  background: rgba(255, 255, 255, 0.05);
  box-shadow: 0 0 6px #4995ff;

  div {
    width: 64px;
    height: 64px;
    overflow: hidden;
    border-radius: inherit;
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.profile-user__info {
  display: grid;
  gap: 6px;
  min-width: 0;
  padding-top: 2px;
  line-height: 1;

  > span {
    color: rgba(255, 255, 255, 0.7);
    font-size: 14px;
  }
}

.profile-user__name {
  display: flex;
  align-items: center;
  gap: 5px;
  min-width: 0;

  strong {
    max-width: 150px;
    overflow: hidden;
    font-size: 16px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  button {
    display: grid;
    width: 25px;
    height: 25px;
    place-items: center;
    padding: 0;
    transition: transform 160ms ease, opacity 160ms ease;

    &:active {
      opacity: 0.82;
      transform: scale(0.94);
    }
  }

  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
}

.profile-user__id {
  display: flex;
  align-items: center;
  gap: 4px;

  svg {
    width: 14px;
    height: 14px;
    flex: 0 0 auto;
    fill: rgba(255, 255, 255, 0.68);
  }
}

.profile-vip {
  display: flex;
  align-items: center;
  width: max-content;
  padding: 0;
}

.profile-vip__badge {
  width: 81px;
  height: 32px;
  object-fit: contain;
}

.profile-vip__jump {
  width: 33px;
  height: 29px;
  object-fit: contain;
  will-change: transform;
}

.wallet-card {
  position: relative;
  overflow: hidden;
  border-radius: 12px;
  background: #041426;
  box-shadow: inset 0 0 5px rgba(255, 255, 255, 0.25);
}

.wallet-card__bg {
  pointer-events: none;
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: fill;
}

.wallet-card__records {
  position: absolute;
  top: 0;
  right: 0;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 2px;
  border-radius: 0 0 0 12px;
  padding: 10px 6px 10px 10px;
  color: #ffffff;
  background: linear-gradient(90deg, #3578df 0%, #59b16a 100%);
  box-shadow: inset -3px 0 3px rgba(0, 0, 0, 0.2);
  font-size: 12px;
  font-weight: 500;
  line-height: 1;

  svg {
    width: 14px;
    height: 14px;
    fill: #f9f5f5;
  }

  svg path + path {
    fill: #97a5ca;
  }
}

.wallet-card__chevron {
  width: 12px !important;
  height: 12px !important;
  fill: none !important;

  path {
    fill: none !important;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.wallet-card__content {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px;
}

.wallet-card__icon {
  width: 44px;
  height: 44px;
  object-fit: contain;
}

.wallet-card__main {
  display: grid;
  gap: 4px;
  min-width: 0;

  > span {
    font-size: 14px;
    font-weight: 500;
  }
}

.wallet-card__amount,
.bonus-wallet__amount > div {
  display: flex;
  align-items: flex-end;
  gap: 8px;

  span {
    font-size: 16px;
    font-weight: 600;
    line-height: 1.25;
  }

  strong {
    font-size: 20px;
    font-weight: 700;
    line-height: 1.1;
  }
}

.wallet-card__amount button,
.bonus-wallet__amount button {
  display: grid;
  width: 20px;
  height: 20px;
  place-items: center;
  color: #4d81e7;

  svg {
    width: 18px;
    height: 18px;
    fill: currentColor;
  }

  &:active svg {
    transform: rotate(55deg);
  }
}

.wallet-card__actions {
  position: relative;
  z-index: 1;
  display: flex;
  gap: 8px;
  padding: 0 16px 16px;

  button {
    flex: 1 1 0;
    height: 32px;
    border-radius: 10px;
    color: #ffffff;
    font-size: 14px;
    font-weight: 500;

    &:first-child {
      background: linear-gradient(96deg, #3d8aef 18%, #2f69d1 71%);
      box-shadow: 0 1px 5px rgba(61, 137, 238, 0.4);
    }

    &:last-child {
      background: linear-gradient(93deg, #56b26a 16%, #3a9a4d 82%);
      box-shadow: 0 1px 5px rgba(86, 175, 104, 0.4);
    }
  }
}

.bonus-wallet,
.profile-menu {
  overflow: hidden;
  margin-top: 14px;
  border-radius: 12px;
  background: #1e2438;
}

.bonus-wallet header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #2a3352;
  box-shadow: 0 2px 2px rgba(0, 0, 0, 0.1);

  div,
  button {
    display: flex;
    align-items: center;
  }

  div {
    gap: 6px;
    min-width: 0;
  }

  img {
    width: 20px;
    height: 20px;
    object-fit: contain;
  }

  span {
    font-size: 14px;
    font-weight: 500;
  }

  button {
    gap: 2px;
    color: #4d81e7;
    font-size: 14px;
    font-weight: 700;
    text-decoration: underline;

    svg {
      width: 14px;
      height: 14px;
      fill: none;
    }

    path {
      stroke: currentColor;
      stroke-width: 2;
      stroke-linecap: round;
      stroke-linejoin: round;
    }
  }
}

.bonus-wallet__amount {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;

  > div {
    gap: 4px;

    span,
    strong {
      font-size: 14px;
    }
  }
}

.profile-menu__item {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
  min-height: 44px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding: 12px 16px;
  text-align: left;

  &:last-child {
    border-bottom: 0;
  }

  > img {
    width: 20px;
    height: 20px;
    flex: 0 0 auto;
    object-fit: contain;
  }

  span {
    flex: 1 1 auto;
    min-width: 0;
    margin-left: 8px;
    color: #ffffff;
    font-size: 14px;
    font-weight: 500;
  }

  svg {
    width: 16px;
    height: 16px;
    flex: 0 0 auto;
    fill: none;
    color: #9ca3af;
  }

  path {
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.profile-menu__badge {
  position: absolute;
  top: 50%;
  right: 12px;
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #ff344d;
  box-shadow: 0 0 8px rgba(255, 52, 77, 0.75);
  transform: translateY(-13px);
}

@media (max-width: 374px) {
  .profile-content {
    padding-right: 12px;
    padding-left: 12px;
  }

  .wallet-card__records {
    padding-left: 8px;
  }
}
</style>
