<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { isLoggedIn, mockUser } from '../mock/session'
import { h5Api } from '../services/h5Api'
import { localDialogImage } from '../utils/localAssets'
import { formatCurrency, runtimeConfig, toRuntimeAssetUrl } from '../services/runtimeConfig'

const gift = localDialogImage('mini-gift.webp')
const router = useRouter()
const logo = computed(() =>
  toRuntimeAssetUrl(runtimeConfig.value.branding.headerLogoUrl || runtimeConfig.value.branding.logoUrl),
)
const appName = computed(() => runtimeConfig.value.branding.appName || runtimeConfig.value.branding.siteName || 'Casino')
const cashLabel = computed(() => formatCurrency(String(mockUser.value.cash || '0').replace(/[^\d.]/g, '') || 0))
const unreadNotifications = ref(0)
const hasUnreadNotifications = computed(() => isLoggedIn.value && unreadNotifications.value > 0)

const emit = defineEmits<{
  login: []
  register: []
}>()

const openDeposit = () => {
  void router.push({ name: 'wallet-deposit' })
}

const openCash = () => {
  void router.push({ name: 'cash-center' })
}

const refreshNotificationStatus = async () => {
  if (!isLoggedIn.value) {
    unreadNotifications.value = 0
    return
  }
  try {
    const status = await h5Api.getNotificationStatus()
    unreadNotifications.value = Number(status.unreadCount || 0)
  } catch {
    unreadNotifications.value = 0
  }
}

const openSupportNotice = () => {
  void router.push({ name: 'support', query: { tab: 'notice' } })
}

onMounted(() => {
  void refreshNotificationStatus()
  window.addEventListener('h5-notifications-updated', refreshNotificationStatus)
})

onBeforeUnmount(() => {
  window.removeEventListener('h5-notifications-updated', refreshNotificationStatus)
})

watch(isLoggedIn, () => {
  void refreshNotificationStatus()
})
</script>

<template>
  <header class="app-header">
    <div class="brand">
      <img v-if="logo" :src="logo" :alt="appName" /><strong v-else class="brand-name">{{ appName }}</strong>
    </div>

    <template v-if="isLoggedIn">
      <button class="cash-pill" type="button" @click="openCash">
        <span>Cash</span>
        <div class="ml-auto font-bold text-3.5">{{ cashLabel }}</div>
      </button>
      <button class="deposit-button" type="button" @click="openDeposit">
        <span>Deposit</span>
        <img :src="gift" alt="" aria-hidden="true" />
        <i class="gift-point gift-point--one" aria-hidden="true" />
        <i class="gift-point gift-point--two" aria-hidden="true" />
        <i class="gift-point gift-point--three" aria-hidden="true" />
      </button>
    </template>

    <template v-else>
      <button class="auth-button auth-button--login" type="button" @click="emit('login')">Login</button>
      <button class="auth-button auth-button--register" type="button" @click="emit('register')">
        <img :src="gift" alt="" aria-hidden="true" />
        <i class="gift-point gift-point--one" aria-hidden="true" />
        <i class="gift-point gift-point--two" aria-hidden="true" />
        <i class="gift-point gift-point--three" aria-hidden="true" />
        <span>Register</span>
      </button>
    </template>

    <button
      class="notify-button"
      :class="{ 'has-unread': hasUnreadNotifications }"
      type="button"
      aria-label="Notifications"
      @click="openSupportNotice"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path
          d="M12 22a2.7 2.7 0 0 0 2.56-1.85H9.44A2.7 2.7 0 0 0 12 22Zm7.2-5.6-1.55-2.1V9.75c0-3.18-1.72-5.84-4.73-6.55a1 1 0 0 0-1.84 0c-3.01.71-4.73 3.37-4.73 6.55v4.55L4.8 16.4a1.45 1.45 0 0 0 1.17 2.32h12.06a1.45 1.45 0 0 0 1.17-2.32Z"
        />
      </svg>
      <span v-if="hasUnreadNotifications" class="notify-button__badge" aria-hidden="true" />
    </button>
  </header>
</template>

<style scoped lang="scss">
.brand-name { display:block; max-width:130px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; font-size:18px; font-weight:800; color:var(--runtime-text-color,#fff); }
.app-header {
  position: sticky;
  top: 0;
  z-index: 40;
  display: flex;
  height: 68px;
  flex-shrink: 0;
  align-items: center;
  width: 100%;
  padding: 0 15px;
  border-bottom: 1px solid rgba(213, 230, 255, 0.12);
  color: #ffffff;
  background: hsl(var(--background, 224 55% 8%));
}

.cash-pill {
  display: flex;
  flex: 1 1 0;
  height: 38px;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  border-radius: 6px;
  padding: 0 4px;
  color: #ffffff;
  font-size: 16px;
  line-height: 1rem;
  background: rgba(57, 127, 229, 0.3);

  span {
    display: grid;
    width: 37px;
    height: 30px;
    place-items: center;
    border-radius: 6px;
    background: #397fe5;
    font-size: 10px;
    font-weight: 700;
  }

  strong {
    min-width: 0;
    overflow: hidden;
    font-size: 16px;
    font-weight: 800;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.deposit-button {
  position: relative;
  display: grid;
  width: 76px;
  height: 38px;
  flex: 0 0 76px;
  place-items: center;
  overflow: visible;
  margin-left: 6px;
  border-radius: 6px;
  color: #ffffff;
  font-size: 14px;
  font-weight: 500;
  background: linear-gradient(96deg, var(--runtime-primary-color, #3d8aef) 18%, var(--runtime-button-color, #2f69d1) 71%);
  box-shadow: 0 4px 14px rgba(61, 138, 239, 0.34);
  cursor: pointer;

  img {
    position: absolute;
    top: -12px;
    right: -12px;
    width: 31px;
    height: 32px;
    object-fit: contain;
    animation: header-gift 2.6s ease-in-out infinite;
  }
}

.brand {
  width: 89px;
  flex: 0 0 89px;
  min-width: 0;
  margin-right: 0;

  img {
    width: 89px;
    margin-left: 3px;
    height: auto;
    object-fit: contain;
    filter: drop-shadow(0 4px 3px rgba(0, 0, 0, 0.3));
  }
}

.auth-button {
  position: relative;
  display: grid;
  height: 40px;
  flex: 1 1 0;
  place-items: center;
  overflow: visible;
  margin-left: 8px;
  border-radius: 8px;
  color: #ffffff;
  font-size: 14.5px;
  font-weight: 500;
  line-height: 1;
  transition:
    transform 180ms ease,
    filter 180ms ease;

  &:active {
    transform: translateY(1px) scale(0.985);
  }

  &--login {
    background: linear-gradient(180deg, var(--runtime-primary-color, #438cf0), var(--runtime-button-color, #2f73de));
    box-shadow: 0 10px 26px rgba(57, 127, 229, 0.26);
  }

  &--register {
    background: linear-gradient(180deg, #5db66e, #479c57);
    box-shadow: 0 10px 26px rgba(82, 177, 101, 0.22);

    img {
      position: absolute;
      top: -12px;
      right: -7px;
      width: 28px;
      height: 30px;
      object-fit: contain;
      animation: header-gift 2.6s ease-in-out infinite;
    }
  }
}

.gift-point {
  position: absolute;
  z-index: 2;
  clip-path: polygon(99.75% 55.02%, 72.38% 69.98%, 60.6% 98.86%, 37.91% 77.46%, 6.8% 75.18%, 20.15% 46.99%, 12.7% 16.7%, 43.64% 20.68%, 70.15% 4.24%, 75.92% 34.89%);
  animation: point-burst 2.5s linear infinite;

  &--one {
    top: -10px;
    right: 11px;
    width: 4px;
    height: 4px;
    background: #27a9b7;
  }

  &--two {
    top: -6px;
    right: 2px;
    width: 8px;
    height: 8px;
    background: #ea826a;
    animation-delay: 160ms;
  }

  &--three {
    top: 10px;
    right: -1px;
    width: 6px;
    height: 6px;
    background: #ffc107;
    animation-delay: 320ms;
  }
}

.notify-button {
  position: relative;
  display: grid;
  width: 38px;
  height: 38px;
  flex: 0 0 38px;
  place-items: center;
  margin-left: 6px;
  border-radius: 14px;
  background: #29334f;

  svg {
    width: 24px;
    height: 24px;
    fill: #ffffff;
    filter: drop-shadow(0 2px 0 rgba(0, 0, 0, 0.18));
  }
}

.notify-button__badge {
  position: absolute;
  top: 7px;
  right: 7px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #ed2c35;
  box-shadow: 0 0 0 2px rgba(41, 51, 79, 0.9);

  &::before {
    position: absolute;
    inset: 0;
    border-radius: inherit;
    background: inherit;
    animation: notification-badge-pulse 1.35s ease-out infinite;
    content: '';
  }
}

@keyframes header-gift {
  0%,
  100% {
    transform: rotate(0deg) translateY(0);
  }

  50% {
    transform: rotate(-14deg) translateY(-2px);
  }
}

@keyframes point-burst {
  0%,
  100% {
    opacity: 1;
    transform: translate(0, 0) scale(1);
  }

  50% {
    opacity: 0.72;
    transform: translate(-6px, 6px) scale(0.72) rotate(20deg);
  }
}

@keyframes notification-badge-pulse {
  0% {
    opacity: 0.8;
    transform: scale(1);
  }

  100% {
    opacity: 0;
    transform: scale(2.5);
  }
}

@media (max-width: 374px) {
  .app-header {
    padding-right: 12px;
    padding-left: 12px;
  }

  .brand {
    width: 78px;
    flex-basis: 78px;
    margin-right: 6px;
  }

  .brand img {
    width: 76px;
  }

  .auth-button {
    margin-left: 6px;
    font-size: 15px;
  }

  .cash-pill span {
    width: 33px;
    height: 27px;
  }

  .cash-pill,
  .cash-pill strong {
    font-size: 14px;
  }

  .deposit-button {
    width: 72px;
    flex-basis: 72px;
    margin-left: 5px;
    font-size: 13px;
  }

  .notify-button {
    margin-left: 5px;
  }
}
</style>
