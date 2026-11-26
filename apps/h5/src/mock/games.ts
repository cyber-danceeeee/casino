import socialFacebook from '../assets/images/dialogs/social-facebook.webp'
import socialInstagram from '../assets/images/dialogs/social-instagram.webp'
import socialMessenger from '../assets/images/dialogs/social-messenger.webp'
import socialTelegram from '../assets/images/dialogs/social-telegram.webp'
import socialTiktok from '../assets/images/dialogs/social-tiktok.webp'
import socialWhatsapp from '../assets/images/dialogs/social-whatsapp.webp'
import socialX from '../assets/images/dialogs/social-x.webp'

export type CategoryKey = string

export interface GameCategory {
  key: CategoryKey
  label: string
  icon?: string
  short: string
}

export interface BannerItem {
  id: string
  image: string
  title: string
}

export interface GameItem {
  id: string
  gameDbId?: number | string
  name: string
  provider: string
  providerId?: string
  providerCode?: string
  providerGameCode?: string
  launchUrl?: string
  cover: string
  categories: CategoryKey[]
  players: string
  badge?: string
  status?: 'online' | 'maintenance'
  sortOrder?: number
  favorite?: boolean
}

export interface GameSection {
  key: CategoryKey
  title: string
  icon?: string
  games: GameItem[]
  pages?: Record<number, GameItem[]>
  total?: number
  page?: number
  pageSize?: number
}

export interface ProviderLogo {
  id: string
  image: string
}

export interface SocialLink {
  id: string
  label: string
  image: string
  icon?: string
  url?: string
  shareURL?: string
  shareTitle?: string
  shareText?: string
  shareDescription?: string
}

const image = (fileName: string) => `/static/h5/images/${fileName}`

export const categories: GameCategory[] = [
  { key: 'hot', label: 'Hot', short: 'H', icon: image('category-hot.webp') },
  { key: 'slots', label: 'Slots', short: 'S', icon: image('category-slot.webp') },
  { key: 'live-casino', label: 'Live Casino', short: 'LC', icon: image('category-casino.webp') },
  { key: 'table-games', label: 'Table Games', short: 'TG', icon: image('category-casino.webp') },
  { key: 'crash', label: 'Crash', short: 'CR', icon: image('category-hot.webp') },
  { key: 'fishing', label: 'Fishing', short: 'F', icon: image('category-casino.webp') },
  { key: 'sports', label: 'Sports', short: 'SP' },
]

export const banners: BannerItem[] = [
  { id: 'checkin', title: 'Daily Check-in Rewards', image: image('offers-daily-checkin.png') },
  { id: 'red-packet', title: 'Red Packet Rain', image: image('offers-red-packet.png') },
  { id: 'refer', title: 'Refer & Earn', image: image('offers-benefit-card.png') },
  { id: 'app-download', title: 'App Download Bonus', image: image('offers-app-download.png') },
  { id: 'recharge', title: 'Recharge Bonus', image: image('offers-recharge.png') },
]

export const games: GameItem[] = []

export const gameSections: GameSection[] = []

export const providerLogos: ProviderLogo[] = Array.from({ length: 11 }, (_, index) => ({
  id: `provider-${index + 1}`,
  image: image(`provider-${index + 1}.webp`),
}))

export const socialLinks: SocialLink[] = [
  { id: 'telegram', label: 'Telegram', image: socialTelegram },
  { id: 'x', label: 'X', image: socialX },
  { id: 'facebook', label: 'Facebook', image: socialFacebook },
  { id: 'messenger', label: 'Messenger', image: socialMessenger },
  { id: 'tiktok', label: 'TikTok', image: socialTiktok },
  { id: 'instagram', label: 'Instagram', image: socialInstagram },
  { id: 'whatsapp', label: 'WhatsApp', image: socialWhatsapp },
]
