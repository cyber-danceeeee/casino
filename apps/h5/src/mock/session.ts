import { computed, ref } from 'vue'
import {
  h5Api,
  type H5AuthPayload,
  type H5PendingRewardItem,
  type H5RegisterRewardPayload,
  type H5UserPayload,
} from '../services/h5Api'
import { normalizePhoneInput, phoneCountryCode, phoneLength } from '../services/runtimeConfig'
import { h5SessionFlagKey, h5SessionTokenKey, h5SessionUserKey } from '../utils/authStorage'

export interface MockUser {
  id: number
  phone: string
  userId: string
  cash: string
  bonus: string
  withdrawable: string
  vipLevel: number
}

const defaultUser: MockUser = {
  id: 0,
  phone: '3088888889',
  userId: 'LFG21809',
  cash: '1.80',
  bonus: '0.00',
  withdrawable: '0.00',
  vipLevel: 0,
}

const registerRewardStorageKey = 'game-casino-register-reward'

const toMockUser = (user: H5UserPayload): MockUser => ({
  id: user.id || 0,
  phone: user.phone || defaultUser.phone,
  userId: user.userNo || defaultUser.userId,
  cash: user.wallet?.cashBalance || '0.00',
  bonus: user.wallet?.bonusBalance || '0.00',
  withdrawable: user.wallet?.withdrawableBalance || '0.00',
  vipLevel: user.vipLevel || 0,
})

const readInitialState = () => {
  if (typeof window === 'undefined') return false
  const token = window.localStorage.getItem(h5SessionTokenKey) || ''
  return window.localStorage.getItem(h5SessionFlagKey) === 'logged-in' && token.startsWith('user_')
}

const readStoredUser = () => {
  if (typeof window === 'undefined') return { ...defaultUser }

  try {
    const stored = JSON.parse(window.localStorage.getItem(h5SessionUserKey) || 'null') as H5UserPayload | null
    return stored ? toMockUser(stored) : { ...defaultUser }
  } catch {
    return { ...defaultUser }
  }
}

export const mockUser = ref<MockUser>(readStoredUser())
export const mockLoggedIn = ref(readInitialState())
export const showLoginInviteChest = ref(false)
export const pendingRegisterReward = ref<H5RegisterRewardPayload | null>(readStoredRegisterReward())
export const pendingActivityRewards = ref<H5PendingRewardItem[]>([])

export const isLoggedIn = computed(() => mockLoggedIn.value)

export const formatPakistanPhone = (phone = mockUser.value.phone) => {
  const clean = normalizePhoneInput(phone)
  if (clean.length !== phoneLength()) return `${phoneCountryCode()} ${clean}`
  return `${phoneCountryCode()} ${clean.slice(0, 3)} ${clean.slice(3)}`
}

function readStoredRegisterReward() {
  if (typeof window === 'undefined') return null

  try {
    return JSON.parse(window.sessionStorage.getItem(registerRewardStorageKey) || 'null') as H5RegisterRewardPayload | null
  } catch {
    return null
  }
}

const applyAuthPayload = (payload: H5AuthPayload) => {
  mockUser.value = toMockUser(payload.user)
  mockLoggedIn.value = true
  pendingRegisterReward.value = null
  showLoginInviteChest.value = false

  if (typeof window !== 'undefined') {
    window.localStorage.setItem(h5SessionFlagKey, 'logged-in')
    window.localStorage.setItem(h5SessionTokenKey, payload.token)
    window.localStorage.setItem(h5SessionUserKey, JSON.stringify(payload.user))
    window.sessionStorage.removeItem(registerRewardStorageKey)
    window.dispatchEvent(new CustomEvent('casino-auth-changed'))
  }
}

export const loginMockUser = async (
  phone?: string,
  password = '',
  inviteCode = '',
  mode: 'login' | 'register' = 'login',
) => {
  const cleanPhone = phone ? normalizePhoneInput(phone) : ''
  const payload = await h5Api[mode]({
    phone: cleanPhone,
    password,
    inviteCode,
  })

  applyAuthPayload(payload)
}

export const clearRegisterReward = () => {
  pendingRegisterReward.value = null
  if (typeof window !== 'undefined') {
    window.sessionStorage.removeItem(registerRewardStorageKey)
  }
}

export const loadPendingRegisterReward = async () => {
  if (!mockLoggedIn.value) {
    clearRegisterReward()
    pendingActivityRewards.value = []
    return null
  }

  const payload = await h5Api.getPendingRewards()
  pendingActivityRewards.value = payload.items || []
  const reward = payload.registerReward && Number(payload.registerReward.amount) > 0 ? payload.registerReward : null
  pendingRegisterReward.value = reward

  if (typeof window !== 'undefined') {
    if (reward) {
      window.sessionStorage.setItem(registerRewardStorageKey, JSON.stringify(reward))
    } else {
      window.sessionStorage.removeItem(registerRewardStorageKey)
    }
  }

  return reward
}

export const claimPendingActivityReward = async (rewardID: number) => {
  const payload = await h5Api.claimPendingReward(rewardID)
  pendingActivityRewards.value = pendingActivityRewards.value.filter((reward) => reward.id !== rewardID)
  if (payload.user) {
    mockUser.value = toMockUser(payload.user)
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(h5SessionUserKey, JSON.stringify(payload.user))
    }
  }
  return payload.reward
}

export const claimPendingRegisterReward = async () => {
  const payload = await h5Api.claimRegisterReward()
  if (payload.user) {
    mockUser.value = toMockUser(payload.user)
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(h5SessionUserKey, JSON.stringify(payload.user))
    }
  }
  return payload.registerReward
}

export const refreshMockUser = async () => {
  if (!mockLoggedIn.value) return

  const { user } = await h5Api.getWalletSummary()
  mockUser.value = toMockUser(user)

  if (typeof window !== 'undefined') {
    window.localStorage.setItem(h5SessionUserKey, JSON.stringify(user))
  }
}

export const logoutMockUser = () => {
  const pushToken = typeof window === 'undefined' ? '' : window.__CASINO_ANDROID_PUSH_TOKEN__ || ''
  if (pushToken) void h5Api.unregisterPushDevice(pushToken).catch(() => undefined)
  void h5Api.logout().catch(() => undefined)

  mockLoggedIn.value = false
  showLoginInviteChest.value = false
  pendingRegisterReward.value = null
  pendingActivityRewards.value = []

  if (typeof window !== 'undefined') {
    window.localStorage.removeItem(h5SessionFlagKey)
    window.localStorage.removeItem(h5SessionTokenKey)
    window.localStorage.removeItem(h5SessionUserKey)
    window.sessionStorage.removeItem(registerRewardStorageKey)
    window.dispatchEvent(new CustomEvent('casino-auth-changed'))
  }
}
