import { ref, shallowRef } from 'vue'
import { phoneLimits } from '../utils/phoneNumber'

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/$/, '')
const API_ORIGIN = API_BASE_URL.startsWith('http') ? new URL(API_BASE_URL).origin : ''
const runtimeConfigStoragePrefix = 'casino-h5-runtime-config-single-site-v2'

export interface RuntimeConfig {
  resourceVersion: string
  country: {
    code: string
    isoCountryCode?: string
    name: string
    nameZh?: string
    nameEn?: string
    timezone: string
    phoneCountryCode?: string
    phoneLength?: number
  }
  currency: {
    code: string
    symbol: string
    decimalPlaces: number
    nameZh?: string
    nameEn?: string
  }
  language: {
    default: string
    current: string
    supported: string[]
  }
  seo?: { title?: string; description?: string }
  branding: {
    appName?: string
    siteName?: string
    logoUrl?: string
    faviconUrl?: string
    loginLogoUrl?: string
    headerLogoUrl?: string
    loadingLogoUrl?: string
    appIconUrl?: string
    loadingAnimationUrl?: string
    loadingBackgroundUrl?: string
    theme?: Record<string, string>
  }
  maintenance: {
    enabled: boolean
    title?: string
    message?: string
    estimatedRecoveryAt?: string
    supportURL?: string
    supportEmail?: string
  }
}

interface ApiEnvelope<T> {
  code: number
  msg: string
  data: T
}

export const defaultRuntimeConfig: RuntimeConfig = {
  resourceVersion: 'runtime-default-v1',
  country: { code: '', isoCountryCode: '', name: '', nameEn: '', timezone: 'UTC', phoneCountryCode: '+92', phoneLength: 10 },
  currency: { code: '', symbol: '', decimalPlaces: 2, nameEn: '' },
  language: { default: '', current: '', supported: [] },
  branding: {
    theme: {
      primaryColor: '#3d8aef',
      buttonColor: '#2f69d1',
      backgroundColor: '#050b18',
      textColor: '#ffffff',
    },
  },
  maintenance: { enabled: false, message: '' },
}

export const runtimeConfig = shallowRef<RuntimeConfig>(defaultRuntimeConfig)
export const runtimeConfigReady = ref(false)

export function toRuntimeAssetUrl(url?: string) {
  if (!url) return ''
  if (/^https?:\/\//i.test(url) || url.startsWith('data:')) return url
  if (url.startsWith('/')) return API_ORIGIN ? `${API_ORIGIN}${url}` : url
  return url
}

function applyTheme(theme?: Record<string, string>) {
  if (!theme) return
  const root = document.documentElement
  const map: Record<string, string> = {
    primaryColor: '--runtime-primary-color',
    secondaryColor: '--runtime-secondary-color',
    buttonColor: '--runtime-button-color',
    backgroundColor: '--runtime-background-color',
    textColor: '--runtime-text-color',
  }
  for (const [key, variable] of Object.entries(map)) {
    const value = theme[key]
    if (value) root.style.setProperty(variable, value)
  }
}

function setLink(rel: string, href: string, sizes?: string) {
  const selector = sizes ? `link[rel="${rel}"][sizes="${sizes}"]` : `link[rel="${rel}"]`
  let link = document.head.querySelector<HTMLLinkElement>(selector)
  if (!link) {
    link = document.createElement('link')
    link.rel = rel
    if (sizes) link.sizes = sizes
    document.head.append(link)
  }
  link.href = href
}

function applyDocumentBranding(config: RuntimeConfig) {
  const title = config.seo?.title || config.branding.appName || config.branding.siteName
  if (title) document.title = title
  let description = document.head.querySelector<HTMLMetaElement>('meta[name="description"]')
  if (!description) {
    description = document.createElement('meta')
    description.name = 'description'
    document.head.append(description)
  }
  description.content = config.seo?.description || ''

  const language = config.language.current || config.language.default || 'en'
  document.documentElement.lang = language

  const themeColor = config.branding.theme?.backgroundColor || '#07101f'
  const themeMeta = document.head.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
  themeMeta?.setAttribute('content', themeColor)

  const favicon = toRuntimeAssetUrl(config.branding.faviconUrl || config.branding.appIconUrl)
  if (favicon) setLink('icon', favicon)

  const appIcon = toRuntimeAssetUrl(config.branding.appIconUrl || config.branding.faviconUrl)
  if (appIcon) setLink('apple-touch-icon', appIcon, '180x180')
}

function queryDefaultCountry() {
  const params = new URLSearchParams(window.location.search)
  return params.get('country') || window.localStorage.getItem('h5-runtime-country') || ''
}

function queryDefaultLanguage() {
  const params = new URLSearchParams(window.location.search)
  return params.get('language') || window.localStorage.getItem('h5-runtime-language') || ''
}

function runtimeConfigStorageKey(country: string, language: string) {
  return `${runtimeConfigStoragePrefix}:${country || 'default'}:${language || 'default'}`
}

function readCachedRuntimeConfig(key: string) {
  try {
    const value = JSON.parse(window.localStorage.getItem(key) || '') as RuntimeConfig
    return value?.resourceVersion && value?.country && value?.currency && value?.language ? value : null
  } catch {
    return null
  }
}

function writeCachedRuntimeConfig(key: string, value: RuntimeConfig) {
  try {
    window.localStorage.setItem(key, JSON.stringify(value))
  } catch {
    // Storage may be unavailable in private browsing; network config still works.
  }
}

export async function loadRuntimeConfig() {
  const params = new URLSearchParams()
  const country = queryDefaultCountry()
  const language = queryDefaultLanguage()
  const storageKey = runtimeConfigStorageKey(country, language)
  if (country) params.set('country', country)
  if (language) params.set('language', language)

  try {
    const response = await fetch(`${API_BASE_URL}/app/config${params.toString() ? `?${params}` : ''}`, {
      headers: { Accept: 'application/json' },
      cache: 'no-store',
    })
    const envelope = (await response.json()) as ApiEnvelope<RuntimeConfig>
    if (!response.ok || envelope.code !== 0 || !envelope.data) throw new Error(envelope.msg || 'runtime config failed')
    runtimeConfig.value = {
      ...defaultRuntimeConfig,
      ...envelope.data,
      branding: { ...defaultRuntimeConfig.branding, ...(envelope.data.branding || {}) },
      currency: { ...defaultRuntimeConfig.currency, ...(envelope.data.currency || {}) },
      language: { ...defaultRuntimeConfig.language, ...(envelope.data.language || {}) },
      maintenance: { ...defaultRuntimeConfig.maintenance, ...(envelope.data.maintenance || {}) },
    }
    writeCachedRuntimeConfig(storageKey, runtimeConfig.value)
  } catch (error) {
    console.warn('[runtime-config] using fallback config.', error)
    runtimeConfig.value = readCachedRuntimeConfig(storageKey) || defaultRuntimeConfig
  } finally {
    applyTheme(runtimeConfig.value.branding.theme)
    applyDocumentBranding(runtimeConfig.value)
    runtimeConfigReady.value = true
  }
  return runtimeConfig.value
}

export function formatCurrency(value: string | number, options: { includeCode?: boolean } = {}) {
  const amount = typeof value === 'number' ? value : Number(String(value).replace(/,/g, ''))
  const config = runtimeConfig.value.currency
  const decimals = Number.isFinite(config.decimalPlaces) ? config.decimalPlaces : 2
  const formatted = Number.isFinite(amount)
    ? amount.toLocaleString('en-US', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
    : '0'.padEnd(decimals ? decimals + 2 : 1, '0')
  // const separator = config.symbol && /[A-Za-z]$/.test(config.symbol) ? ' ' : ''
  const suffix = options.includeCode ? ` ${config.code}` : ''
  // return `${config.symbol}${separator}${formatted}${suffix}`
  return `${formatted}${suffix}`
}

export function currencySymbol() {
  return runtimeConfig.value.currency.symbol
}

export function phoneCountryCode() {
  return runtimeConfig.value.country.phoneCountryCode || '+92'
}

export function phoneLength() {
  return phoneLimits(runtimeConfig.value.country.phoneLength ?? 10, phoneCountryCode()).max
}

export function phoneMinimumLength() {
  return phoneLimits(runtimeConfig.value.country.phoneLength ?? 10, phoneCountryCode()).min
}

export function normalizePhoneInput(value: string) {
  return value.replace(/\D/g, '').slice(0, phoneLength())
}
