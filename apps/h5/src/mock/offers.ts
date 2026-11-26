import redPacketCard from '../assets/images/offers-red-packet.png'
import inviteSpinCard from '../assets/images/offers-invite-spin.png'
import inviteChestCard from '../assets/images/offers-invite-chest.png'
import redPacketMain from '../assets/images/offers-red-packet-rain-1.webp'
import redPacketBackground from '../assets/images/offers-red-packet-rain-bg.webp'
import redPacketConfetti from '../assets/images/offers-red-packet-rain-confetti.webp'
import redPacketFalling from '../assets/images/offers-red-packet-rain-falling.webp'
import redPacketPackets from '../assets/images/offers-red-packet-rain-packets.webp'
import inviteSpinMain from '../assets/images/offers-invite-spin-detail-1.webp'
import inviteSpinWheel from '../assets/images/offers-invite-spin-detail-2.webp'
import inviteSpinCoins from '../assets/images/offers-invite-spin-detail-3.webp'
import inviteSpinPanel from '../assets/images/offers-invite-spin-detail-4.webp'
import inviteChestMain from '../assets/images/dialogs/offers-invite-chest-detail-1.webp'
import inviteChestSide from '../assets/images/dialogs/offers-invite-chest-detail-2.webp'
import inviteChestBox from '../assets/images/dialogs/offers-invite-chest-detail-3.webp'
import inviteChestGold1 from '../assets/images/dialogs/offers-invite-chest-gold-1.webp'
import inviteChestGold2 from '../assets/images/dialogs/offers-invite-chest-gold-2.webp'
import inviteChestGold3 from '../assets/images/dialogs/offers-invite-chest-gold-3.webp'
import { localOfferImage, localVipImage } from '../utils/localAssets'

const image = (name: string) => `/static/h5/images/${name}`

export type OfferTab = 'events' | 'vip'
export type VipRequirementTab = 'level-up' | 'retain'

export interface OfferEvent {
  id: string
  title: string
  image: string
  detailSlug?: string
  targetPath?: string
  unread?: boolean
  activityCode?: string
  updatedAt?: string
}

export type OfferActivityKind = 'deposit' | 'spin' | 'chest' | 'red-packet' | 'checkin' | 'card'

export interface OfferActivityDetail {
  slug: string
  eventId: string
  kind: OfferActivityKind
  title: string
  subtitle: string
  reward: string
  maxReward: string
  note: string
  cta: string
  hero: string
  accent: string
  glow: string
  banner?: string
  bundle?: string
  rulesTitle: string
  rewardRows: Array<{
    label: string
    value: string
    bonus: string
  }>
  rules: string[]
}

export interface VipBenefit {
  label: string
  value: string
  delta: string
  type: 'reward' | 'privilege'
}

export interface VipRequirement {
  level: number
  image: string
  deposit: string
  validBets: string
}

export const offerTabAssets = {
  events: {
    active: localOfferImage('tab-event-active.webp'),
    inactive: localOfferImage('tab-event-inactive.webp'),
  },
  vip: {
    active: localOfferImage('tab-vip-active.webp'),
    inactive: localOfferImage('tab-vip-inactive.webp'),
  },
}

export const vipAssets = {
  current: localVipImage('vip0.webp'),
  deposit: localVipImage('deposit.webp'),
  emptyBonus: localVipImage('bonus-empty.webp'),
  info: localVipImage('info.webp'),
}

export const offerEvents: OfferEvent[] = [
  {
    id: 'first-deposit',
    title: 'First Deposit Rewards',
    image: image('offers-first-deposit.png'),
    detailSlug: 'first-deposit-bonus',
  },
  {
    id: 'recharge',
    title: 'Recharge Rewards',
    image: image('offers-recharge.png'),
    detailSlug: 'deposit-bonus',
  },
  {
    id: 'daily-rebate',
    title: 'Daily Deposit Rebate',
    image: image('offers-daily-rebate.png'),
    detailSlug: 'daily-deposit-cashback',
  },
  {
    id: 'invite-spin',
    title: 'Spin & Invite',
    image: inviteSpinCard,
    detailSlug: 'invitation-wheel',
  },
  {
    id: 'benefit-card',
    title: 'Benefit Card',
    image: image('offers-benefit-card.png'),
    detailSlug: 'welfare-card',
  },
  {
    id: 'invite-chest',
    title: 'Invite Chest',
    image: inviteChestCard,
    detailSlug: 'invitation-chest',
  },
  {
    id: 'invite-cashback',
    title: 'Invite Deposit Cashback',
    image: image('offers-invite-cashback.png'),
    detailSlug: 'sub-deposit-reward',
  },
  {
    id: 'weekly-deposit',
    title: 'Weekly Deposit Bonus',
    image: image('offers-weekly-deposit.png'),
    detailSlug: 'periodic-deposit-reward/weekly',
  },
  {
    id: 'daily-checkin',
    title: 'Daily Check-in Rewards',
    image: image('offers-daily-checkin.png'),
    detailSlug: 'daily-sign-in',
  },
  {
    id: 'red-packet',
    title: 'Red Packet Rain',
    image: redPacketCard,
    detailSlug: 'red-packet-rain',
  },
  {
    id: 'daily-rescue',
    title: 'Daily Rescue Cash',
    image: image('offers-daily-rescue.png'),
    detailSlug: 'daily-relief',
  },
  {
    id: 'cashback-leaderboard',
    title: 'Cashback Leaderboard',
    image: image('offers-cashback-leaderboard.png'),
    detailSlug: 'rebate-ranking',
  },
  {
    id: 'phone-verification',
    title: 'Phone Verification Bonus',
    image: image('offers-phone-verification.png'),
    targetPath: '/auth/register?redirect=%2Foffers',
  },
  {
    id: 'app-download',
    title: 'App Download Bonus',
    image: image('offers-app-download.png'),
  },
]

const depositRows = [
  { label: '100.00-299.99', value: '+30.00', bonus: '+20.00' },
  { label: '300.00-499.99', value: '+80.00', bonus: '+60.00' },
  { label: '500.00-999.99', value: '+150.00', bonus: '+100.00' },
  { label: '1,000.00+', value: '+300.00', bonus: '+200.00' },
]

export const offerActivityAssets = {
  firstDepositBundle: image('offers-first-deposit-bundle.webp'),
  dailyDepositBundle: image('offers-daily-deposit-bundle.webp'),
  rulesBanner: image('offers-rules-banner.webp'),
  rulesBanner1: image('offers-rules-banner-1.webp'),
  rulesBanner2: image('offers-rules-banner-2.webp'),
  rulesBundle: image('offers-rules-bundle.webp'),
  rulesBundle1: image('offers-rules-bundle-1.webp'),
  redPacketMain,
  redPacketBg: redPacketBackground,
  redPacketConfetti,
  redPacketFalling,
  redPacketPackets,
  inviteSpinMain,
  inviteSpinWheel,
  inviteSpinCoins,
  inviteSpinPanel,
  inviteChestMain,
  inviteChestSide,
  inviteChestBox,
  inviteChestGold1,
  inviteChestGold2,
  inviteChestGold3,
}

export const offerActivityDetails: OfferActivityDetail[] = [
  {
    slug: 'first-deposit-bonus',
    eventId: 'first-deposit',
    kind: 'deposit',
    title: 'First Deposit Super Rewards',
    subtitle: 'Get it on your first deposit',
    reward: '+100% Rewards',
    maxReward: 'Maximum Reward: +200',
    note: 'First deposit only',
    cta: 'Deposit Now',
    hero: offerActivityAssets.firstDepositBundle,
    accent: '#e16137',
    glow: '#7a46ff',
    banner: offerActivityAssets.rulesBanner,
    bundle: offerActivityAssets.rulesBundle1,
    rulesTitle: 'Reward Details',
    rewardRows: depositRows,
    rules: [
      'Only the first successful deposit can participate in this event.',
      'The bonus is issued according to the deposit range and must complete the audit requirement before withdrawal.',
      'This activity cannot be combined with other deposit rewards unless the platform allows it.',
    ],
  },
  {
    slug: 'recharge-reward',
    eventId: 'recharge',
    kind: 'deposit',
    title: 'Recharge Rewards',
    subtitle: 'Recharge every day and get more',
    reward: '+30% Bonus',
    maxReward: 'Maximum Reward: +500',
    note: 'Daily recharge bonus',
    cta: 'Deposit Now',
    hero: offerActivityAssets.dailyDepositBundle,
    accent: '#29a8ff',
    glow: '#00c8ff',
    banner: offerActivityAssets.rulesBanner1,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Recharge Details',
    rewardRows: [
      { label: '100.00-499.99', value: '+18.00', bonus: '+8%' },
      { label: '500.00-999.99', value: '+58.00', bonus: '+10%' },
      { label: '1,000.00-4,999.99', value: '+128.00', bonus: '+12%' },
      { label: '5,000.00+', value: '+500.00', bonus: '+15%' },
    ],
    rules: [
      'Rewards are calculated from the successful recharge amount of the day.',
      'Rewards must be claimed on the activity page before the event expires.',
      'Abnormal accounts or duplicate accounts may be excluded from this event.',
    ],
  },
  {
    slug: 'daily-deposit',
    eventId: 'daily-rebate',
    kind: 'deposit',
    title: 'Daily Deposit Rebate',
    subtitle: 'Deposit daily, boost your rebate',
    reward: '+20% Cashback',
    maxReward: 'Maximum Reward: +300',
    note: 'Valid for today',
    cta: 'Join Now',
    hero: offerActivityAssets.dailyDepositBundle,
    accent: '#39d47c',
    glow: '#21efb4',
    banner: offerActivityAssets.rulesBanner2,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Cashback Details',
    rewardRows: [
      { label: '100.00-299.99', value: '+12.00', bonus: '1x audit' },
      { label: '300.00-799.99', value: '+36.00', bonus: '2x audit' },
      { label: '800.00-1,999.99', value: '+88.00', bonus: '3x audit' },
      { label: '2,000.00+', value: '+300.00', bonus: '5x audit' },
    ],
    rules: [
      'Cashback is based on the total daily deposit amount.',
      'The activity resets every day according to platform time.',
      'The platform reserves the right to review suspicious activity before issuing rewards.',
    ],
  },
  {
    slug: 'invitation-wheel',
    eventId: 'invite-spin',
    kind: 'spin',
    title: 'Spin & Invite',
    subtitle: 'Invite friends and spin for prizes',
    reward: 'Lucky Spin',
    maxReward: 'Win cash, bonus and chests',
    note: 'Invite to unlock spins',
    cta: 'Invite Now',
    hero: offerActivityAssets.inviteSpinMain,
    accent: '#ff43d1',
    glow: '#8248ff',
    banner: offerActivityAssets.rulesBanner1,
    bundle: offerActivityAssets.inviteSpinPanel,
    rulesTitle: 'Spin Rules',
    rewardRows: [
      { label: '1 Friend', value: '1 Spin', bonus: 'Random' },
      { label: '3 Friends', value: '3 Spins', bonus: '+Chest' },
      { label: '5 Friends', value: '6 Spins', bonus: '+Bonus' },
      { label: '10 Friends', value: '12 Spins', bonus: '+Cash' },
    ],
    rules: [
      'Invite valid users to receive spin chances.',
      'Each valid invitation must meet the platform activity requirement.',
      'Spin rewards are mock data here and are ready to connect to the API later.',
    ],
  },
  {
    slug: 'welfare-card',
    eventId: 'benefit-card',
    kind: 'card',
    title: 'Benefit Card',
    subtitle: 'Unlock exclusive daily benefits',
    reward: 'Daily Welfare',
    maxReward: 'Multiple bonuses available',
    note: 'Limited welfare card',
    cta: 'Activate Now',
    hero: offerActivityAssets.rulesBundle1,
    accent: '#ffc342',
    glow: '#ff67ad',
    banner: offerActivityAssets.rulesBanner,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Benefit Rules',
    rewardRows: [
      { label: 'Daily Login', value: '+5.00', bonus: 'Daily' },
      { label: 'Deposit Boost', value: '+20%', bonus: 'Active' },
      { label: 'Loss Rescue', value: '+10%', bonus: 'Daily' },
      { label: 'VIP Extra', value: '+1', bonus: 'Chance' },
    ],
    rules: [
      'Benefits are refreshed after the card is activated.',
      'Daily benefits must be claimed before the daily reset.',
      'Card rewards follow the same audit requirements as bonus wallet rewards.',
    ],
  },
  {
    slug: 'invitation-chest',
    eventId: 'invite-chest',
    kind: 'chest',
    title: 'Invite Chest',
    subtitle: 'Open chests with your friends',
    reward: 'Treasure Bonus',
    maxReward: 'Big cash surprise inside',
    note: 'Invite to open chests',
    cta: 'Invite Now',
    hero: offerActivityAssets.inviteChestMain,
    accent: '#ffc13a',
    glow: '#8c57ff',
    banner: offerActivityAssets.rulesBanner2,
    bundle: offerActivityAssets.inviteChestSide,
    rulesTitle: 'Chest Rules',
    rewardRows: [
      { label: 'Bronze Chest', value: '+10.00', bonus: '1 invite' },
      { label: 'Silver Chest', value: '+50.00', bonus: '3 invites' },
      { label: 'Gold Chest', value: '+120.00', bonus: '5 invites' },
      { label: 'Mega Chest', value: '+300.00', bonus: '10 invites' },
    ],
    rules: [
      'Chest progress grows after each valid invited user.',
      'Rewards are issued after the chest is opened.',
      'Duplicate or invalid invitations do not count toward chest progress.',
    ],
  },
  {
    slug: 'subordinate-recharge',
    eventId: 'invite-cashback',
    kind: 'deposit',
    title: 'Invite Deposit Cashback',
    subtitle: 'Earn when your friends recharge',
    reward: '+3% Cashback',
    maxReward: 'Unlimited commission style bonus',
    note: 'Invite deposit reward',
    cta: 'Share Now',
    hero: offerActivityAssets.dailyDepositBundle,
    accent: '#42d4ff',
    glow: '#3f78ff',
    banner: offerActivityAssets.rulesBanner1,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Cashback Rules',
    rewardRows: [
      { label: 'Level 1 Friend', value: '+3%', bonus: 'Deposit' },
      { label: 'Level 2 Friend', value: '+1%', bonus: 'Deposit' },
      { label: 'Daily Cap', value: 'No Cap', bonus: 'Active' },
      { label: 'Settlement', value: 'Daily', bonus: 'Auto' },
    ],
    rules: [
      'Cashback is calculated from invited users successful deposits.',
      'Settlement data is displayed after platform review.',
      'Invite rewards are not valid for duplicate or risk-controlled accounts.',
    ],
  },
  {
    slug: 'weekly-deposit',
    eventId: 'weekly-deposit',
    kind: 'deposit',
    title: 'Weekly Deposit Bonus',
    subtitle: 'Deposit every week for richer rewards',
    reward: '+25% Weekly',
    maxReward: 'Maximum Reward: +700',
    note: 'Weekly reward cycle',
    cta: 'Deposit Now',
    hero: offerActivityAssets.firstDepositBundle,
    accent: '#6c8cff',
    glow: '#48d5ff',
    banner: offerActivityAssets.rulesBanner2,
    bundle: offerActivityAssets.rulesBundle1,
    rulesTitle: 'Weekly Details',
    rewardRows: [
      { label: '500.00-999.99', value: '+70.00', bonus: 'Weekly' },
      { label: '1,000.00-4,999.99', value: '+160.00', bonus: 'Weekly' },
      { label: '5,000.00-9,999.99', value: '+360.00', bonus: 'Weekly' },
      { label: '10,000.00+', value: '+700.00', bonus: 'Weekly' },
    ],
    rules: [
      'The weekly cycle follows platform settlement time.',
      'Only successful deposits during the cycle are counted.',
      'Rewards must complete the required valid bets before withdrawal.',
    ],
  },
  {
    slug: 'sign-in',
    eventId: 'daily-checkin',
    kind: 'checkin',
    title: 'Daily Check-in Rewards',
    subtitle: 'Log in daily, boost your rewards',
    reward: '7 Day Streak',
    maxReward: 'Rewards increase every day',
    note: 'Check in once per day',
    cta: 'Check In',
    hero: image('offers-daily-checkin.png'),
    accent: '#31b8ff',
    glow: '#57ffd4',
    banner: offerActivityAssets.rulesBanner,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Check-in Rules',
    rewardRows: [
      { label: 'Day 1', value: '+1.00', bonus: 'Bonus' },
      { label: 'Day 3', value: '+5.00', bonus: 'Bonus' },
      { label: 'Day 5', value: '+12.00', bonus: 'Bonus' },
      { label: 'Day 7', value: '+30.00', bonus: 'Cash' },
    ],
    rules: [
      'Each account can check in once per day.',
      'Missing a day may reset the consecutive check-in progress.',
      'Check-in rewards are issued to the matching wallet after claim.',
    ],
  },
  {
    slug: 'red-packet-rain',
    eventId: 'red-packet',
    kind: 'red-packet',
    title: 'Red Packet Rain',
    subtitle: 'Catch lucky packets while they fall',
    reward: 'Lucky Cash Rain',
    maxReward: 'Random packets every round',
    note: 'Open during event time',
    cta: 'Start Catching',
    hero: offerActivityAssets.redPacketMain,
    accent: '#ff4545',
    glow: '#ffcc56',
    banner: offerActivityAssets.rulesBanner1,
    bundle: offerActivityAssets.redPacketBg,
    rulesTitle: 'Packet Rules',
    rewardRows: [
      { label: 'Small Packet', value: '+1.00', bonus: 'Random' },
      { label: 'Lucky Packet', value: '+10.00', bonus: 'Random' },
      { label: 'Mega Packet', value: '+77.00', bonus: 'Random' },
      { label: 'Rain Round', value: 'Limited', bonus: 'Timer' },
    ],
    rules: [
      'Tap falling red packets during the active event period.',
      'Each round has a limited duration and a random reward result.',
      'Rewards shown in this demo are static and ready for backend integration.',
    ],
  },
  {
    slug: 'daily-relief',
    eventId: 'daily-rescue',
    kind: 'deposit',
    title: 'Daily Rescue Cash',
    subtitle: 'Get rescued after tough rounds',
    reward: '+10% Rescue',
    maxReward: 'Maximum Rescue: +200',
    note: 'Daily loss relief',
    cta: 'Claim Rescue',
    hero: offerActivityAssets.dailyDepositBundle,
    accent: '#47d487',
    glow: '#32ffe6',
    banner: offerActivityAssets.rulesBanner,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Rescue Rules',
    rewardRows: [
      { label: 'Loss 100+', value: '+5.00', bonus: 'Relief' },
      { label: 'Loss 500+', value: '+30.00', bonus: 'Relief' },
      { label: 'Loss 1,000+', value: '+80.00', bonus: 'Relief' },
      { label: 'Loss 3,000+', value: '+200.00', bonus: 'Relief' },
    ],
    rules: [
      'Daily rescue is based on qualifying valid losses.',
      'The reward can be claimed after daily settlement.',
      'Accounts with abnormal game behavior may be excluded.',
    ],
  },
  {
    slug: 'rebate-ranking',
    eventId: 'cashback-leaderboard',
    kind: 'card',
    title: 'Cashback Leaderboard',
    subtitle: 'Rank up for bigger rebates',
    reward: 'Top Ranking Rewards',
    maxReward: 'Daily leaderboard settlement',
    note: 'Rank up for rebate',
    cta: 'View Ranking',
    hero: image('offers-cashback-leaderboard.png'),
    accent: '#ffd153',
    glow: '#627dff',
    banner: offerActivityAssets.rulesBanner2,
    bundle: offerActivityAssets.rulesBundle1,
    rulesTitle: 'Leaderboard Rules',
    rewardRows: [
      { label: 'Top 1', value: '+888.00', bonus: 'Cash' },
      { label: 'Top 2', value: '+588.00', bonus: 'Cash' },
      { label: 'Top 3', value: '+388.00', bonus: 'Cash' },
      { label: 'Top 10', value: '+88.00', bonus: 'Cash' },
    ],
    rules: [
      'Ranking is calculated from valid rebate activity during the cycle.',
      'Rewards are settled after the leaderboard closes.',
      'Final ranking is subject to platform review.',
    ],
  },
  {
    slug: 'phone-verification',
    eventId: 'phone-verification',
    kind: 'card',
    title: 'Phone Verification Bonus',
    subtitle: 'Secure your account and claim bonus',
    reward: '+123.00 Bonus',
    maxReward: 'Verify once, claim once',
    note: 'Bind phone to claim',
    cta: 'Verify Now',
    hero: image('offers-phone-verification.png'),
    accent: '#38cfff',
    glow: '#4d81e7',
    banner: offerActivityAssets.rulesBanner,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Verification Rules',
    rewardRows: [
      { label: 'Phone Bind', value: '+123.00', bonus: 'Bonus' },
      { label: 'Security', value: 'Improved', bonus: 'Account' },
      { label: 'Login', value: 'Safer', bonus: 'Protect' },
      { label: 'Claim', value: 'Once', bonus: 'Reward' },
    ],
    rules: [
      'Each account can receive the phone verification reward once.',
      'The phone number must pass platform verification.',
      'The reward is issued after the binding is successful.',
    ],
  },
  {
    slug: 'app-download',
    eventId: 'app-download',
    kind: 'card',
    title: 'Download App Bonus',
    subtitle: 'For a better experience and enhanced security',
    reward: '+77.00 Bonus',
    maxReward: 'Download and claim bonus',
    note: 'Install app to claim',
    cta: 'Download App',
    hero: image('offers-app-download.png'),
    accent: '#20c5ff',
    glow: '#76f1fa',
    banner: offerActivityAssets.rulesBanner1,
    bundle: offerActivityAssets.rulesBundle,
    rulesTitle: 'Download Rules',
    rewardRows: [
      { label: 'App Install', value: '+44.43', bonus: 'Bonus' },
      { label: 'Secure Login', value: '+10.00', bonus: 'Extra' },
      { label: 'Push Enabled', value: '+12.00', bonus: 'Reward' },
      { label: 'First Claim', value: '+77.00', bonus: 'Max' },
    ],
    rules: [
      'Download the official app and log in with the same account.',
      'The app reward can be claimed once per eligible account.',
      'If the device or account is flagged as abnormal, the reward may require review.',
    ],
  },
]

export const getOfferActivityDetail = (slug: string) =>
  offerActivityDetails.find((activity) => activity.slug === slug)

export const vipBenefits: VipBenefit[] = [
  {
    label: 'Level Up Reward',
    value: 'Rs 5.00',
    delta: '+5.00',
    type: 'reward',
  },
  {
    label: 'Weekly Reward',
    value: 'Rs 1.00',
    delta: '+1.00',
    type: 'reward',
  },
  {
    label: 'Monthly Reward',
    value: 'Rs 3.00',
    delta: '+3.00',
    type: 'reward',
  },
  {
    label: 'Additional Daily Withdrawal Count',
    value: '1',
    delta: '+1',
    type: 'privilege',
  },
]

export const vipRequirements: VipRequirement[] = [
  { level: 1, image: localVipImage('vip1.webp'), deposit: '100.00', validBets: '300.00' },
  { level: 2, image: localVipImage('vip2.webp'), deposit: '500.00', validBets: '1,500.00' },
  { level: 3, image: localVipImage('vip3.webp'), deposit: '1,000.00', validBets: '3,000.00' },
  { level: 4, image: localVipImage('vip4.webp'), deposit: '3,000.00', validBets: '10,000.00' },
  { level: 5, image: localVipImage('vip5.webp'), deposit: '5,000.00', validBets: '18,000.00' },
  { level: 6, image: localVipImage('vip6.webp'), deposit: '10,000.00', validBets: '35,000.00' },
  { level: 7, image: localVipImage('vip7.webp'), deposit: '20,000.00', validBets: '80,000.00' },
  { level: 8, image: localVipImage('vip8.webp'), deposit: '50,000.00', validBets: '200,000.00' },
  { level: 9, image: localVipImage('vip9.webp'), deposit: '100,000.00', validBets: '400,000.00' },
  { level: 10, image: localVipImage('vip10.webp'), deposit: '300,000.00', validBets: '1,200,000.00' },
  { level: 11, image: localVipImage('vip11.webp'), deposit: '500,000.00', validBets: '2,000,000.00' },
  { level: 12, image: localVipImage('vip12.webp'), deposit: '1,000,000.00', validBets: '4,000,000.00' },
  { level: 13, image: localVipImage('vip13.webp'), deposit: '2,000,000.00', validBets: '8,000,000.00' },
  { level: 14, image: localVipImage('vip14.webp'), deposit: '5,000,000.00', validBets: '20,000,000.00' },
  { level: 15, image: localVipImage('vip15.webp'), deposit: '10,000,000.00', validBets: '40,000,000.00' },
]
