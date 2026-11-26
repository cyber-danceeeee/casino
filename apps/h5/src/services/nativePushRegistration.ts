import { h5Api, type H5PushDeviceRequest } from './h5Api'
import { h5SessionFlagKey } from '../utils/authStorage'

type NativePushDeviceMetadata = Omit<H5PushDeviceRequest, 'token' | 'platform'>

interface PushTokenEvent extends Event {
  detail?: Partial<H5PushDeviceRequest>
}

declare global {
  interface Window {
    __CASINO_ANDROID_PUSH_TOKEN__?: string
    __CASINO_ANDROID_PUSH_DEVICE__?: Partial<H5PushDeviceRequest>
  }
}

let activeRegistration = ''
let registrationPromise: Promise<void> | null = null

function isAuthenticated() {
  return window.localStorage.getItem(h5SessionFlagKey) === 'logged-in'
}

function currentAppVersion() {
  const match = navigator.userAgent.match(/CasinoAndroid\/([\w.-]+)/i)
  return match?.[1] || ''
}

async function registerToken(token: string, metadata: NativePushDeviceMetadata = {}) {
  const normalized = token.trim()
  if (!normalized || !isAuthenticated()) return
  const payload: H5PushDeviceRequest = {
    ...metadata,
    token: normalized,
    platform: 'ANDROID',
    appVersion: metadata.appVersion || currentAppVersion(),
    language: metadata.language || navigator.language || 'en',
  }
  const registrationSignature = JSON.stringify(payload)
  if (activeRegistration === registrationSignature) return
  if (registrationPromise) await registrationPromise
  if (activeRegistration === registrationSignature) return

  registrationPromise = h5Api.registerPushDevice(payload).then(() => {
    activeRegistration = registrationSignature
  }).catch((error) => {
    console.warn('[push] device registration failed.', error)
  }).finally(() => {
    registrationPromise = null
  })
  await registrationPromise
}

function syncCurrentToken() {
  const token = window.__CASINO_ANDROID_PUSH_TOKEN__ || ''
  const detail = window.__CASINO_ANDROID_PUSH_DEVICE__ || {}
  const { token: _token, platform: _platform, ...metadata } = detail
  if (token) void registerToken(token, metadata)
}

export function initializeNativePushRegistration() {
  window.addEventListener('casino-push-token', (event) => {
    const detail = (event as PushTokenEvent).detail || {}
    const token = detail.token?.trim() || ''
    if (!token) return
    const { token: _token, platform: _platform, ...metadata } = detail
    window.__CASINO_ANDROID_PUSH_TOKEN__ = token
    window.__CASINO_ANDROID_PUSH_DEVICE__ = metadata
    void registerToken(token, metadata)
  })
  window.addEventListener('casino-auth-changed', syncCurrentToken)
  syncCurrentToken()
}
