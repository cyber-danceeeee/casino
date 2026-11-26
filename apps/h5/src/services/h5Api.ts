import type {
  BannerItem,
  CategoryKey,
  GameCategory,
  GameItem,
  GameSection,
  SocialLink,
} from '../mock/games'
import { socialLinks as localSocialLinks, categories as defaultCategories } from '../mock/games'
import { h5SessionFlagKey, h5SessionTokenKey, h5SessionUserKey } from '../utils/authStorage'
import { phoneCountryCode, runtimeConfig } from './runtimeConfig'

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v1').replace(/\/$/, '')

interface ApiEnvelope<T> {
  code: number
  msg: string
  data: T
}

interface RemoteBanner {
  id: number
  bannerCode: string
  title: string
  imageURL: string
}

interface RemoteCategory {
  id?: number
  categoryKey: string
  label: string
  shortLabel: string
  iconURL: string
}

interface RemoteGame {
  id: number
  gameCode: string
  providerID: number
  providerCode: string
  providerGameId: string
  providerGameCode: string
  categoryCode: string
  platformCategoryID: number
  platformCategoryCode: string
  providerCategoryCode: string
  name: string
  gameType: string
  gameTypeCode: string
  thumbnailUrl: string
  coverURL: string
  badge?: string
  playerCount: number
  isHot?: boolean
  isFeatured?: boolean
  isNew?: boolean
  isEnabled?: boolean
  effectiveEnabled?: boolean
  sortOrder?: number
  status: number
  favorite?: boolean
}

export interface H5NavigationItem {
  id: number
  menuCode: string
  label: string
  labelZh: string
  labelEn: string
  labels: Record<string, string>
  iconURL: string
  lottieURL: string
  soundURL: string
  displayPosition: 'BOTTOM_NAV' | 'HOME_QUICK' | 'PROFILE_MENU'
  badgeType: 'NONE' | 'NEW' | 'HOT'
  actionType:
    | 'OPEN_INTERNAL_PAGE'
    | 'OPEN_EXTERNAL_URL'
    | 'OPEN_DEPOSIT'
    | 'OPEN_WITHDRAW'
    | 'OPEN_SUPPORT'
    | 'OPEN_PROMOTION'
    | 'OPEN_GAME'
    | 'OPEN_DOWNLOAD'
  actionTarget: string
  providerCode: string
  gameID: number
  providerGameCode: string
  gameName: string
  sortOrder: number
  configVersion: number
  enabled: boolean
}

interface RemoteProvider {
  id: number
  providerCode: string
  name?: string
  logoURL: string
  status?: number
}

interface RemoteSocialLink {
  id?: string
  platformCode?: string
  label: string
  icon?: string
  image?: string
  iconURL?: string
  url?: string
  shareURL?: string
  shareTitle?: string
  shareText?: string
  shareDescription?: string
}

export interface H5Wallet {
  currency: string
  cashBalance: string
  bonusBalance: string
  withdrawableBalance: string
  frozenBalance: string
}

export interface H5UserPayload {
  id: number
  userNo: string
  phoneCountryCode: string
  phone: string
  nickname: string
  avatarUrl: string
  inviteCode: string
  vipLevel: number
  status: number
  wallet: H5Wallet
}

export interface H5WalletSummaryPayload {
  user: H5UserPayload
  wallet: H5Wallet
}

export interface H5WalletRecordPayload {
  id: number
  title: string
  source: string
  amount: string
  currency: string
  direction: string
  sign: string
  balance: string
  time: string
  orderNumber: string
  status: string
  category: string
  gameId: number
  gameName: string
  providerCode: string
  providerGameId: string
}

export interface H5WalletRecordsPayload {
  tab: string
  items: H5WalletRecordPayload[]
  total: number
  page: number
  pageSize: number
}

export interface H5CashAuditPayload {
  id: number
  walletTransactionID: number
  type: string
  source: string
  amount: string
  currency: string
  createdAt: string
  auditMultiplier: string
  auditRequirement: string
  requiredAmount: string
  progressAmount: string
  progressPercent: string
  status: string
}

export interface H5CashAuditsPayload {
  items: H5CashAuditPayload[]
  total: number
}

export interface H5RegisterRewardPayload {
  id?: number
  amount: string
  currency: string
  walletType: 'cash' | 'bonus' | string
  cashBefore: string
  cashAfter: string
  transactionID?: number
  status?: number
}

export interface H5AuthPayload {
  token: string
  user: H5UserPayload
  registerReward?: H5RegisterRewardPayload
}

export interface H5PushDeviceRequest {
  token: string
  platform: 'ANDROID'
  deviceId?: string
  fingerprintHash?: string
  manufacturer?: string
  brand?: string
  model?: string
  deviceName?: string
  osVersion?: string
  sdkVersion?: number
  securityPatch?: string
  networkType?: string
  networkMetered?: boolean
  networkRoaming?: boolean
  carrierName?: string
  screenWidth?: number
  screenHeight?: number
  screenDensity?: number
  timezone?: string
  locale?: string
  batteryLevel?: number
  charging?: boolean
  notificationPermission?: string
  appVersion?: string
  appVersionCode?: number
  language?: string
}

export interface H5PendingRewardsPayload {
  registerReward?: H5RegisterRewardPayload | null
  items: H5PendingRewardItem[]
  count: number
  totalAmount: string
  currency: string
}

export interface H5PendingRewardItem {
  id: number
  activityID: number
  activityCode: string
  slug: string
  title: string
  iconURL: string
  rewardType: 'cash' | 'bonus' | 'mixed' | string
  cashAmount: string
  bonusAmount: string
  amount: string
  currency: string
  status: number
  createdAt: string
}

export interface H5ClaimRegisterRewardPayload {
  registerReward: H5RegisterRewardPayload
  user: H5UserPayload
}

export interface H5ClaimPendingRewardPayload {
  reward: H5PendingRewardItem
  user: H5UserPayload
}

export interface H5AuthRequest {
  phone: string
  password: string
  inviteCode?: string
  phoneCountryCode?: string
}

export interface HomeProvider { providerCode: string; name: string; logoUrl?: string }

export interface HomePayload {
  providers: HomeProvider[]
  banners: BannerItem[]
  categories: GameCategory[]
  socialLinks: SocialLink[]
  raw: Record<string, unknown>
}

interface RemoteHomePayload {
  providers?: HomeProvider[]
  banners: RemoteBanner[]
  categories: RemoteCategory[]
  socialLinks?: RemoteSocialLink[]
}

interface RemoteGameListPayload {
  list: RemoteGame[]
  total?: number
  page?: number
  pageSize?: number
}

export interface H5GameListParams {
  category?: CategoryKey
  page?: number
  pageSize?: number
  keyword?: string
  provider?: string
  sort?: string
  favorite?: boolean
}

export interface H5GameListResult {
  games: GameItem[]
  total: number
  page: number
  pageSize: number
}

export interface H5MusicTrack {
  id: number
  code: string
  nameZh: string
  nameEn: string
  audioUrl: string
  loop: boolean
  volumePercent: number
}

export interface H5MusicConfig {
  mode: 'RANDOM' | 'SPECIFIED'
  specifiedMusicID: number
  tracks: H5MusicTrack[]
}

export interface H5OfferEventPayload {
  id: number
  activityCode: string
  slug: string
  title: string
  cardImageURL: string
  updatedAt: string
}

interface H5VipLevelPayload {
  id: number
  levelNo: number
  name: string
  iconURL: string
  levelUpReward: string
  weeklyReward: string
  monthlyReward: string
  dailyWithdrawCount: number
}

interface H5VipRequirementPayload {
  id: number
  requirementType: 'level_up' | 'retain'
  targetLevelNo: number
  periodDays: number
  depositAmount: string
  validBetAmount: string
}

export interface H5VipPayload {
  currentLevel: number
  depositAmount: string
  levels: H5VipLevelPayload[]
  requirements: H5VipRequirementPayload[]
  benefits: Array<Record<string, string>>
}

export interface H5OffersPayload {
  events: H5OfferEventPayload[]
  vip: {
    levels: H5VipLevelPayload[]
    benefits: Array<Record<string, string>>
  }
}

export interface H5SupportCategoryPayload {
  id?: number
  categoryCode: string
  name: string
  sort?: number
  status?: number
}

export interface H5SupportQuestionPayload {
  id: number
  categoryId: number
  question: string
  answer: string
  answerHTML: string
  sort: number
  status: number
}

export interface H5SupportNoticePayload {
  id: string
  source: string
  title: string
  contentHTML: string
  noticeType?: string
  createdAt?: string
  read?: boolean
}

export interface H5SupportPayload {
  categories: H5SupportCategoryPayload[]
  questions: H5SupportQuestionPayload[]
  notices: H5SupportNoticePayload[]
}

export interface H5NotificationStatusPayload {
  unreadCount: number
}

export interface H5AnnouncementPayload {
  id: number
  title: string
  summary: string
  content: string
  announcementType: 'SYSTEM' | 'MAINTENANCE' | 'PROMOTION' | 'PAYMENT' | 'GAME' | 'GENERAL'
  priority: number
  isPinned: boolean
  publishedAt?: string
  configVersion: number
}

export interface H5PopupPayload {
  campaignId: number
  componentCode: string
  priority: number
  triggerType: 'HOME_ENTER' | 'PAGE_ENTER'
  triggerValue: string
  frequency: {
    type: 'EVERY_VISIT' | 'ONCE_PER_SESSION' | 'ONCE_PER_DAY' | 'ONCE_PER_USER'
    displayLimit: number
    cooldownSeconds: number
  }
  action: {
    type: 'NONE' | 'CLOSE_ONLY' | 'OPEN_INTERNAL_PAGE' | 'OPEN_EXTERNAL_URL' | 'OPEN_GAME' | 'OPEN_PROMOTION' | 'OPEN_DEPOSIT' | 'OPEN_WITHDRAW' | 'OPEN_SUPPORT' | 'OPEN_COUPON' | 'OPEN_VIP' | 'OPEN_DOWNLOAD'
    target: string
  }
  configVersion: number
}

export interface H5GameDetailPayload {
  id: number
  name: string
  providerCode: string
  providerGameCode: string
}

export interface H5DepositMethodPayload {
  id: number
  methodCode: string
  name: string
  methodType: string
  iconURL: string
  minAmount: string
  maxAmount: string
  sort: number
  status: number
}

export interface H5DepositChannelPayload {
  id: number
  methodID: number
  methodCode: string
  channelCode: string
  name: string
  countryCode: string
  currencyCode: string
  feeRate: string
  minAmount: string
  maxAmount: string
  sort: number
  status: number
}

export interface H5DepositAmountOptionPayload {
  id: number
  optionCode: string
  amount: string
  bonusReward: string
  cashReward: string
  rewardAmount: string
  sort: number
  status: number
}

export interface H5DepositVoucherRewardPayload {
  id: number
  amount: string
  bonusReward: string
  cashReward: string
  rewardAmount: string
  sort: number
  status: number
}

export interface H5DepositVoucherPayload {
  id: number
  voucherCode: string
  title: string
  iconURL: string
  minAmount: string
  maxAmount: string
  amountRange: string
  maxReward: string
  bonusReward: string
  cashReward: string
  rewardAmount: string
  rewards?: H5DepositVoucherRewardPayload[]
  available: boolean
  sort: number
  status: number
}

export interface H5DepositConfigPayload {
  methods: H5DepositMethodPayload[]
  channels: H5DepositChannelPayload[]
  amountOptions: H5DepositAmountOptionPayload[]
  vouchers: H5DepositVoucherPayload[]
}

export interface H5CreateDepositRequest {
  amount: string
  methodCode: string
  channelCode: string
  voucherCode?: string
}

export interface H5DepositOrderPayload {
  id: number
  orderNo: string
  userID: number
  amount: string
  paidAmount: string
  currency: string
  methodCode: string
  channelCode: string
  voucherID: number
  status: number
  expiredAt?: string
}

export interface H5CreateDepositPayload {
  order: H5DepositOrderPayload
  cashierUrl: string
}

export interface H5LaunchGameRequest {
  currency?: string
  language?: string
  country?: string
  device?: string
  return_url?: string
  sessionId?: string
}

export interface H5LaunchGamePayload {
  launchUrl: string
  sessionId?: string
  providerCode: string
  providerPlayerId: string
}

export type H5GameSessionExitReason =
  | 'USER_EXIT'
  | 'TIMEOUT'
  | 'PROVIDER_EXIT'
  | 'NETWORK_ERROR'
  | 'SYSTEM_CLOSE'
  | 'UNKNOWN'

export interface H5GameFavoritePayload {
  favorite: boolean
}

export interface H5BonusCenterCategoryPayload {
  code: string
  label: string
}

export interface H5BonusCenterRecordPayload {
  id: number
  orderId: string
  amount: string
  source: string
  category: string
  bonusType: string
  issuedTime: string
  status: string
  requirement: string
  requiredValidBet: string
  completedValidBet: string
  currentWagering: string
  remainingWagering: string
  currency: string
}

export interface H5BonusCenterPayload {
  list: H5BonusCenterRecordPayload[]
  total: number
  page: number
  pageSize: number
  categories: H5BonusCenterCategoryPayload[]
}

const readStoredUser = () => {
  if (typeof window === 'undefined') return null

  try {
    return JSON.parse(window.localStorage.getItem(h5SessionUserKey) || 'null') as H5UserPayload | null
  } catch {
    return null
  }
}

const isValidSessionToken = (token: string) => token.startsWith('user_')

const clearStoredAuth = () => {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(h5SessionFlagKey)
  window.localStorage.removeItem(h5SessionTokenKey)
  window.localStorage.removeItem(h5SessionUserKey)
  window.dispatchEvent(new CustomEvent('casino-auth-changed'))
}

const authHeaders = () => {
  if (typeof window === 'undefined') return {}

  const token = window.localStorage.getItem(h5SessionTokenKey) || ''
  const user = readStoredUser()

  return {
    ...(isValidSessionToken(token) ? { Authorization: `Bearer ${token}` } : {}),
    ...(user?.id ? { 'X-User-ID': String(user.id) } : {}),
    ...(user?.phone ? { 'X-User-Phone': user.phone } : {}),
  }
}

const contentIdentity = (storage: Storage, key: string, prefix: string) => {
  const existing = storage.getItem(key)
  if (existing) return existing
  const random = typeof window.crypto?.randomUUID === 'function'
    ? window.crypto.randomUUID()
    : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
  const value = `${prefix}-${random}`
  storage.setItem(key, value)
  return value
}

const siteContentHeaders = (): Record<string, string> => {
  if (typeof window === 'undefined') return {}
  try {
    return {
      'X-Device-ID': contentIdentity(window.localStorage, 'casino-content-device-id', 'device'),
      'X-Session-ID': contentIdentity(window.sessionStorage, 'casino-content-session-id', 'session'),
    }
  } catch {
    return {}
  }
}

const siteContentQuery = () => {
  const query = new URLSearchParams()
  const language = runtimeConfig.value.language.current || runtimeConfig.value.language.default
  const country = runtimeConfig.value.country.code
  const currency = runtimeConfig.value.currency.code
  if (language) query.set('language', language)
  if (country) query.set('country', country)
  if (currency) query.set('currency', currency)
  const suffix = query.toString()
  return suffix ? `?${suffix}` : ''
}

const popupEventID = (popupID: number, eventType: string) => {
  const random = typeof window.crypto?.randomUUID === 'function'
    ? window.crypto.randomUUID()
    : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
  return `popup-${popupID}-${eventType.toLowerCase()}-${random}`
}

const request = async <T>(path: string, init: RequestInit = {}): Promise<T> => {
  const hasBody = init.body !== undefined
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(hasBody ? { 'Content-Type': 'application/json' } : {}),
      ...authHeaders(),
      ...init.headers,
    },
  })
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`)
  }
  const body = (await response.json()) as ApiEnvelope<T>
  if (body.code !== 0) {
    if (body.code === 401) clearStoredAuth()
    throw new Error(body.msg || 'API request failed')
  }
  return body.data
}

const formatPlayers = (value: number) => {
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1).replace(/\.0$/, '')}k`
  }
  return String(value)
}

const toAssetUrl = (value: string) => {
  return value || ''
}

const supportedProviderList = (import.meta.env.VITE_SUPPORTED_GAME_PROVIDERS || '')
  .split(',')
  .map((value: string) => value.trim().toLowerCase().replace(/[^a-z0-9]/g, ''))
  .filter(Boolean)
const supportedProviderCodes = new Set(supportedProviderList)
const shouldRestrictProviders = supportedProviderCodes.size > 0

const normalizeProviderCode = (value = '') => value.trim().toLowerCase().replace(/[^a-z0-9]/g, '')

const providerIdOfGame = (game: RemoteGame) => game.providerID

const toCategoryKey = (value: string): CategoryKey => {
  const normalized = value.toLowerCase().replace(/[_\s]+/g, '-')
  if (normalized === 'slot') return 'slots'
  if (normalized === 'live' || normalized === 'casino') return 'live-casino'
  if (normalized === 'table' || normalized === 'table-game') return 'table-games'
  if (normalized === 'fish') return 'fishing'
  return normalized || 'hot'
}

const gameCategories = (game: RemoteGame): CategoryKey[] => {
  const keys = new Set<CategoryKey>()
  if (game.isHot || game.badge?.toLowerCase() === 'hot') keys.add('hot')
  keys.add(toCategoryKey(game.platformCategoryCode))
  return [...keys]
}

const toGameItem = (
  game: RemoteGame,
  category: CategoryKey,
  providerById: Map<number, RemoteProvider>,
): GameItem => {
  const provider = providerById.get(providerIdOfGame(game))

  return {
    id: game.gameCode,
    gameDbId: game.id,
    name: game.name,
    provider: provider?.name || game.providerCode,
    providerId: game.providerCode,
    providerCode: game.providerCode,
    providerGameCode: game.providerGameCode,
    cover: toAssetUrl(game.thumbnailUrl),
    categories: gameCategories(game).includes(category) ? gameCategories(game) : [category, ...gameCategories(game)],
    players: formatPlayers(game.playerCount),
    badge: game.badge || undefined,
    status: game.effectiveEnabled ? 'online' : 'maintenance',
    sortOrder: game.sortOrder,
    favorite: game.favorite === true,
  }
}

const toSocialLink = (link: RemoteSocialLink): SocialLink => {
  const id = (link.id || link.platformCode || link.label.toLowerCase().replace(/\s+/g, '-')).toLowerCase()
  const local = localSocialLinks.find((item) => item.id === id || item.label.toLowerCase() === link.label.toLowerCase())
  const image = local?.image || ''

  return {
    id,
    label: link.label || local?.label || id,
    image,
    icon: image,
    url: link.url || link.shareURL || '',
    shareURL: link.shareURL || link.url || '',
    shareTitle: link.shareTitle || '',
    shareText: link.shareText || '',
    shareDescription: link.shareDescription || '',
  }
}

const normalizeHome = (payload: RemoteHomePayload): HomePayload => {
  const categoryItems = (payload.categories || [])
    .map((category) => ({
      key: toCategoryKey(category.categoryKey),
      label: category.label,
      short: category.shortLabel,
      icon: toAssetUrl(category.iconURL) || undefined,
    }))
    .sort((left, right) => {
      if (left.key === 'hot') return -1
      if (right.key === 'hot') return 1
      return 0
    })

  return {
    providers: payload.providers || [],
    banners: (payload.banners || []).map((banner) => ({
      id: banner.bannerCode || String(banner.id),
      title: banner.title,
      image: toAssetUrl(banner.imageURL),
    })),
    categories: (payload.categories || []).some((category) => Number(category.id) > 0) ? categoryItems : defaultCategories,
    socialLinks: (payload.socialLinks || [])
      .filter((link) => link.label)
      .map(toSocialLink)
      .filter((link) => Boolean(link.image)),
    raw: payload as unknown as Record<string, unknown>,
  }
}

const normalizeGameSection = (payload: RemoteGameListPayload, category: GameCategory): GameSection => {
  const games = payload.list
  const providerById = new Map<number, RemoteProvider>()
  const enabledGames = games.filter(
    (game) =>
      game.effectiveEnabled !== false &&
      game.isEnabled !== false &&
      (!shouldRestrictProviders || supportedProviderCodes.has(normalizeProviderCode(game.providerCode))),
  )

  return {
    key: category.key,
    title: category.label,
    icon: category.icon,
    games: enabledGames.map((game) => toGameItem(game, category.key, providerById)),
    total: payload.total ?? enabledGames.length,
    page: payload.page ?? 1,
    pageSize: payload.pageSize ?? enabledGames.length,
  }
}

const buildGameListQuery = (params: H5GameListParams = {}) => {
  const query = new URLSearchParams()
  if (params.category) query.set('category', params.category)
  if (params.page) query.set('page', String(params.page))
  if (params.pageSize) query.set('pageSize', String(params.pageSize))
  if (params.keyword?.trim()) query.set('keyword', params.keyword.trim())
  if (params.provider?.trim()) query.set('provider', params.provider.trim())
  if (params.sort?.trim()) query.set('sort', params.sort.trim())
  if (params.favorite) query.set('favorite', '1')
  const suffix = query.toString()
  return `/games${suffix ? `?${suffix}` : ''}`
}

const normalizeGameList = (payload: RemoteGameListPayload, category: CategoryKey): H5GameListResult => {
  const providerById = new Map<number, RemoteProvider>()
  const enabledGames = (payload.list || []).filter(
    (game) =>
      game.effectiveEnabled !== false &&
      game.isEnabled !== false &&
      (!shouldRestrictProviders || supportedProviderCodes.has(normalizeProviderCode(game.providerCode))),
  )

  return {
    games: enabledGames.map((game) => toGameItem(game, category, providerById)),
    total: payload.total ?? enabledGames.length,
    page: payload.page ?? 1,
    pageSize: payload.pageSize ?? enabledGames.length,
  }
}

export const h5Api = {
  login: (payload: H5AuthRequest) =>
    request<H5AuthPayload>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        phoneCountryCode: phoneCountryCode(),
        ...payload,
      }),
    }),
  register: (payload: H5AuthRequest) =>
    request<H5AuthPayload>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        phoneCountryCode: phoneCountryCode(),
        ...payload,
      }),
    }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  registerPushDevice: (payload: H5PushDeviceRequest) =>
    request<{ registered: boolean }>('/push/devices', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  unregisterPushDevice: (token: string) =>
    request<{ registered: boolean }>('/push/devices', {
      method: 'DELETE',
      body: JSON.stringify({ token }),
    }),
  me: () => request<H5UserPayload>('/me'),
  getWalletSummary: () => request<H5WalletSummaryPayload>('/wallet/summary'),
  getWalletRecords: (params: {
    tab?: string
    type?: string
    direction?: string
    startTime?: string
    endTime?: string
    page?: number
    pageSize?: number
  } = {}) => {
    const query = new URLSearchParams()
    if (params.tab) query.set('tab', params.tab)
    if (params.type) query.set('type', params.type)
    if (params.direction) query.set('direction', params.direction)
    if (params.startTime) query.set('startTime', params.startTime)
    if (params.endTime) query.set('endTime', params.endTime)
    if (params.page) query.set('page', String(params.page))
    if (params.pageSize) query.set('pageSize', String(params.pageSize))
    const suffix = query.toString()
    return request<H5WalletRecordsPayload>(`/wallet/records${suffix ? `?${suffix}` : ''}`)
  },
  getCashAudits: () => request<H5CashAuditsPayload>('/wallet/cash-audits'),
  getPendingRewards: () => request<H5PendingRewardsPayload>('/rewards/pending'),
  claimRegisterReward: () =>
    request<H5ClaimRegisterRewardPayload>('/rewards/register/claim', {
      method: 'POST',
      body: JSON.stringify({}),
    }),
  claimPendingReward: (rewardID: number) =>
    request<H5ClaimPendingRewardPayload>(`/rewards/${encodeURIComponent(String(rewardID))}/claim`, {
      method: 'POST',
      body: JSON.stringify({}),
    }),
  getBonusCenterRecords: (params: { category?: string; page?: number; pageSize?: number } = {}) => {
    const query = new URLSearchParams()
    if (params.category) query.set('category', params.category)
    if (params.page) query.set('page', String(params.page))
    if (params.pageSize) query.set('pageSize', String(params.pageSize))
    const suffix = query.toString()
    return request<H5BonusCenterPayload>(`/bonus-center/records${suffix ? `?${suffix}` : ''}`)
  },
  async getHome() {
    return normalizeHome(await request<RemoteHomePayload>('/home'))
  },
  getNavigation: (position: H5NavigationItem['displayPosition']) => {
    const query = new URLSearchParams({ position })
    const language = runtimeConfig.value.language.current || runtimeConfig.value.language.default
    const country = runtimeConfig.value.country.code
    if (language) query.set('language', language)
    if (country) query.set('country', country)
    return request<H5NavigationItem[]>(`/navigation?${query.toString()}`)
  },
  async getGames(params: H5GameListParams = {}) {
    const category = params.category || 'hot'
    return normalizeGameList(await request<RemoteGameListPayload>(buildGameListQuery(params)), category)
  },
  async getGameSection(category: GameCategory, page = 1, pageSize = 9) {
    const payload = await request<RemoteGameListPayload>(
      buildGameListQuery({ category: category.key, page, pageSize }),
    )
    return normalizeGameSection(payload, category)
  },
  async getGameSections(categories: GameCategory[]) {
    const orderedCategories = [...categories].sort((left, right) => {
      if (left.key === 'hot') return -1
      if (right.key === 'hot') return 1
      return 0
    })
    const sections = await Promise.all(
      orderedCategories.map(async (category) => {
        if (category.key === 'hot') {
          const section = await h5Api.getGameSection(category, 1, 12)
          const pages: Record<number, GameItem[]> = { 1: section.games }
          return {
            ...section,
            pages,
          }
        }

        const initialBatch = await h5Api.getGameSection(category, 1, 18)
        const pages: Record<number, GameItem[]> = {
          1: initialBatch.games.slice(0, 9),
        }
        if (initialBatch.games.length > 9) {
          pages[2] = initialBatch.games.slice(9, 18)
        }

        return {
          ...initialBatch,
          games: pages[1],
          page: 1,
          pageSize: 9,
          pages,
        }
      }),
    )
    return sections
  },
  getOffers: () => request<H5OffersPayload>('/offers'),
  getActivity: (slug: string) => request<Record<string, unknown>>(`/activities/${slug}`),
  getVip: () => request<H5VipPayload>('/vip'),
  getShare: (tab = 'social') => request<Record<string, unknown>>(`/share?tab=${encodeURIComponent(tab)}`),
  getSupport: (category = 'all') => request<H5SupportPayload>(`/support?category=${encodeURIComponent(category)}`),
  getNotificationStatus: () => request<H5NotificationStatusPayload>('/support/notifications/status'),
  getAnnouncements: () => request<{ list: H5AnnouncementPayload[] }>(`/announcements${siteContentQuery()}`, { headers: siteContentHeaders() }),
  getAnnouncement: (id: number) => request<H5AnnouncementPayload>(`/announcements/${encodeURIComponent(String(id))}${siteContentQuery()}`, { headers: siteContentHeaders() }),
  getPopups: () => request<{ items: H5PopupPayload[] }>(`/popups${siteContentQuery()}`, { headers: siteContentHeaders() }),
  trackPopupEvent: (popupID: number, eventType: 'impression' | 'click' | 'dismiss', eventID = popupEventID(popupID, eventType)) =>
    request<{ recorded: boolean }>(`/popups/${encodeURIComponent(String(popupID))}/${eventType}`, {
      method: 'POST',
      headers: { ...siteContentHeaders(), 'Idempotency-Key': eventID },
      body: JSON.stringify({ eventId: eventID }),
    }),
  getGameDetail: (id: number | string) => request<H5GameDetailPayload>(`/games/${encodeURIComponent(String(id))}`),
  getMusicConfig: () => request<H5MusicConfig>('/music/config'),
  markSupportNoticesRead: (ids: string[]) =>
    request<H5NotificationStatusPayload>('/support/notices/read', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    }),
  getDepositConfig: () => request<H5DepositConfigPayload>('/wallet/deposit/config'),
  createDeposit: (payload: H5CreateDepositRequest) =>
    request<H5CreateDepositPayload>('/wallet/deposit', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  launchGameByCode: (providerCode: string | undefined, providerGameCode: string | undefined, payload: H5LaunchGameRequest = {}) => {
    const routeProviderCode = providerCode
    const routeGameCode = providerGameCode
    if (!routeProviderCode || !routeGameCode) {
      return Promise.reject(new Error('missing provider launch metadata'))
    }

    return request<H5LaunchGamePayload>(
      `/games/${encodeURIComponent(routeProviderCode)}/${encodeURIComponent(routeGameCode)}/launch`,
      {
        method: 'POST',
        body: JSON.stringify(payload),
      },
    )
  },
  launchGame: (game: GameItem, payload: H5LaunchGameRequest = {}) =>
    h5Api.launchGameByCode(game.providerCode, game.providerGameCode, payload),
  heartbeatGameSession: (sessionId: string) =>
    request<void>(`/game/session/${encodeURIComponent(sessionId)}/heartbeat`, {
      method: 'POST',
      body: JSON.stringify({}),
    }),
  exitGameSession: (sessionId: string, reason: H5GameSessionExitReason = 'USER_EXIT') =>
    request<void>(`/game/session/${encodeURIComponent(sessionId)}/exit`, {
      method: 'POST',
      body: JSON.stringify({ reason }),
    }),
  getGameFavorite: (gameId: number | string) =>
    request<H5GameFavoritePayload>(`/games/${encodeURIComponent(String(gameId))}/favorite`),
  toggleGameFavorite: (gameId: number | string) =>
    request<H5GameFavoritePayload>(`/games/${encodeURIComponent(String(gameId))}/favorite`, {
      method: 'POST',
    }),
}
