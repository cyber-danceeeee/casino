import { brandLogo } from '../config/branding'
import type { PreloadAsset } from '../utils/preloadAssets'
import { localDialogImage, localHomeImage, localRegisterRewardImage } from '../utils/localAssets'
import { h5Audio } from '../utils/staticAssets'
import { runtimeConfig, toRuntimeAssetUrl } from '../services/runtimeConfig'
import { banners, categories } from './games'
import authCustomer from '../assets/images/profile/customer.webp'
import authHeaderBackground from '../assets/images/auth-header-bg.webp'
import activityRedBackground from '../assets/images/activity-red-bg.webp'
import activityRedConfetti from '../assets/images/activity-red-confetti.webp'
import activityRedFalling from '../assets/images/activity-red-falling.webp'
import activityRedHero from '../assets/images/activity-red-hero.webp'
import activityRedPackets from '../assets/images/activity-red-packets.webp'
import activitySpinBackground from '../assets/images/activity-spin-bg.webp'
import activitySpinFloatA from '../assets/images/activity-spin-float-a.webp'
import activitySpinFloatB from '../assets/images/activity-spin-float-b.webp'
import activitySpinWheel from '../assets/images/activity-spin-wheel.webp'
import activityChestBackground from '../assets/images/activity-chest-bg.webp'
import activityChestGoldA from '../assets/images/activity-chest-gold-a.webp'
import activityChestGoldB from '../assets/images/activity-chest-gold-b.webp'
import activityChestGoldC from '../assets/images/activity-chest-gold-c.webp'
import activityChestLeft from '../assets/images/activity-chest-left.webp'
import activityChestRight from '../assets/images/activity-chest-right.webp'

const image = (fileName: string) => `/static/h5/images/${fileName}`
const audio = h5Audio

const activityImages = [
  activityRedBackground,
  activityRedConfetti,
  activityRedFalling,
  activityRedHero,
  activityRedPackets,
  activitySpinBackground,
  activitySpinFloatA,
  activitySpinFloatB,
  activitySpinWheel,
  activityChestBackground,
  activityChestGoldA,
  activityChestGoldB,
  activityChestGoldC,
  activityChestLeft,
  activityChestRight,
]

const guideImages = [
  localDialogImage('dialog-halo.webp'),
  localDialogImage('dialog-confetti.webp'),
  localDialogImage('dialog-gift.webp'),
  localDialogImage('dialog-app-hero.webp'),
  localDialogImage('activity__guide-flow-dialogs__bg_glow.webp'),
  localDialogImage('activity__guide-flow-dialogs__first_deposit_bg_stars.webp'),
  localDialogImage('activity__guide-flow-dialogs__first_deposit_bowl.webp'),
  localDialogImage('activity__guide-flow-dialogs__first_deposit_guide_treasure.webp'),
  localDialogImage('activity__guide-flow-dialogs__invite_chest_banner.webp'),
  localDialogImage('activity__guide-flow-dialogs__invite_chest_chest.webp'),
  localDialogImage('activity__common__tap_hand.webp'),
  localDialogImage('bonus__rays.webp'),
  localDialogImage('offers-invite-chest-detail-1.webp'),
  localDialogImage('offers-invite-chest-detail-2.webp'),
  localDialogImage('offers-invite-chest-detail-3.webp'),
  localDialogImage('offers-invite-chest-gold-1.webp'),
  localDialogImage('offers-invite-chest-gold-2.webp'),
  localDialogImage('offers-invite-chest-gold-3.webp'),
  localDialogImage('social-facebook.webp'),
  localDialogImage('social-instagram.webp'),
  localDialogImage('social-messenger.webp'),
  localDialogImage('social-telegram.webp'),
  localDialogImage('social-tiktok.webp'),
  localDialogImage('social-whatsapp.webp'),
  localDialogImage('social-x.webp'),
]

const registerRewardImages = [
  localRegisterRewardImage('rays.webp'),
  localRegisterRewardImage('claim_success_bg.webp'),
  localRegisterRewardImage('hero.webp'),
  localRegisterRewardImage('banner.webp'),
  localRegisterRewardImage('icon_cash.webp'),
  localRegisterRewardImage('cash_wallet_icon.webp'),
  localRegisterRewardImage('bonus_wallet_icon.webp'),
]

const floatEntryImages = [
  localHomeImage('float-entry-treasure.webp'),
  localHomeImage('incentive-bottle-90.webp'),
  localHomeImage('incentive-dialog-banner.webp'),
  localHomeImage('incentive-droplet.webp'),
  localHomeImage('float-entry/first-deposit.webp'),
  localHomeImage('float-entry/deposit-bonus.webp'),
  localHomeImage('float-entry/daily-cashback.webp'),
  localHomeImage('float-entry/rewards-card.webp'),
  localHomeImage('float-entry/lucky-wheel.webp'),
  localHomeImage('float-entry/chest-pro.webp'),
  localHomeImage('float-entry/referral.webp'),
  localHomeImage('float-entry/weekly-reward.webp'),
  localHomeImage('float-entry/daily-sign-in.webp'),
  localHomeImage('float-entry/red-packet-rain.webp'),
  localHomeImage('float-entry/cash-back.webp'),
  localHomeImage('float-entry/rebate-rank.webp'),
  localHomeImage('float-entry/phone-verification.webp'),
  localHomeImage('float-entry/download-app.webp'),
]

const criticalImages = [
  brandLogo,
  authCustomer,
  authHeaderBackground,
  localDialogImage('mini-gift.webp'),
  localDialogImage('top-chests.webp'),
  image('world-cup.gif'),
  ...banners.map((banner) => banner.image),
  ...categories.flatMap((category) => (category.icon ? [category.icon] : [])),
  ...activityImages,
  ...guideImages,
  ...registerRewardImages,
  ...floatEntryImages,
]

export const criticalPreloadAssets: PreloadAsset[] = [
  ...criticalImages.map((src) => ({ src, type: 'image' as const })),
  {
    src: audio('invitation-wheel-spin.mp3'),
    type: 'audio',
    label: 'invitation wheel spin',
    timeoutMs: 5000,
  },
  {
    src: audio('invitation-wheel-end.wav'),
    type: 'audio',
    label: 'invitation wheel end',
    timeoutMs: 5000,
  },
]

export const runtimePreloadAssets = (): PreloadAsset[] => {
  const branding = runtimeConfig.value.branding
  return [
    branding.logoUrl,
    branding.faviconUrl,
    branding.loginLogoUrl,
    branding.headerLogoUrl,
    branding.loadingLogoUrl,
    branding.appIconUrl,
    branding.loadingBackgroundUrl,
  ]
    .map((url) => toRuntimeAssetUrl(url))
    .filter(Boolean)
    .map((src) => ({ src, type: 'image' as const }))
}

export const runtimeDeferredAssetURLs = () => {
  const animation = toRuntimeAssetUrl(runtimeConfig.value.branding.loadingAnimationUrl)
  return animation ? [animation] : []
}
