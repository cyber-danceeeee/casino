import { localProfileImage } from '../utils/localAssets'

export interface ProfileMenuItem {
  label: string
  icon: string
  routeName?: string
  externalUrl?: string
  action?: 'download'
  badge?: boolean
}

export interface WalletRecord {
  type: string
  amount: string
  time: string
  orderNumber: string
  status: 'Failed' | 'Pending' | 'Succeeded'
}

export interface CashAuditRecord {
  amount: string
  time: string
  type: string
  requirement: string
  progress: string
}

export interface BonusCenterRecord {
  amount: string
  source: string
  issuedTime: string
  orderId: string
  requirement: string
}

export const profileAssets = {
  background: localProfileImage('profile-vector.webp'),
  avatar: localProfileImage('1(1).webp'),
  gift: localProfileImage('profile-gift.webp'),
  vip: localProfileImage('profile-vip0.webp'),
  jump: localProfileImage('profile-jump.webp'),
  walletBg: localProfileImage('profile-wallet-bg.webp'),
  cashWallet: localProfileImage('profile-cash-wallet.webp'),
  bonusWallet: localProfileImage('profile-bonus-wallet.webp'),
}

export const profileBalances = {
  cash: '1.80',
  bonus: '0.00',
  withdrawable: '0.00',
}

export const accountMenuItems: ProfileMenuItem[] = [
  { label: 'Bet Records', icon: localProfileImage('profile-statement.webp'), routeName: 'bet-records' },
  { label: 'Contact Support', icon: localProfileImage('profile-customer-service.webp'), externalUrl: 'https://direct.lc.chat/19801054/1' },
  { label: 'Security Center', icon: localProfileImage('profile-security-center.webp'), routeName: 'security', badge: true },
  { label: 'Withdrawal Account', icon: localProfileImage('profile-withdraw-account.webp'), routeName: 'bank-account' },
  { label: 'Download App', icon: localProfileImage('profile-download.webp'), action: 'download' },
]

export const supportMenuItems: ProfileMenuItem[] = [
  { label: 'About Us', icon: localProfileImage('profile-about-us.webp'), routeName: 'about' },
  { label: 'User Agreement', icon: localProfileImage('profile-user-agreement.webp'), routeName: 'user-agreement' },
  { label: 'FAQ', icon: localProfileImage('profile-faq.webp'), routeName: 'support' },
  { label: 'Settings', icon: localProfileImage('profile-settings.webp'), routeName: 'settings' },
]

export const walletRecords: WalletRecord[] = [
  {
    type: 'Deposit',
    amount: '+ Rs 100.00',
    time: '2026-07-01 18:02:37',
    orderNumber: '58ref4gf2qeywt',
    status: 'Failed',
  },
  {
    type: 'Deposit',
    amount: '+ Rs 100.00',
    time: '2026-07-01 18:02:03',
    orderNumber: 'dzpsyks85rm3v5',
    status: 'Pending',
  },
  {
    type: 'Deposit',
    amount: '+ Rs 500.00',
    time: '2026-07-01 18:00:28',
    orderNumber: 'h7zn9smh8kgr3g',
    status: 'Succeeded',
  },
  {
    type: 'Deposit',
    amount: '+ Rs 100.00',
    time: '2026-07-01 17:59:06',
    orderNumber: 'bsa3qvprwhhmph',
    status: 'Pending',
  },
  {
    type: 'Deposit',
    amount: '+ Rs 150.00',
    time: '2026-06-21 16:52:24',
    orderNumber: 'k2mzh7nh8vqted',
    status: 'Succeeded',
  },
]

export const cashAuditRecords: CashAuditRecord[] = [
  {
    amount: 'Rs 29.77',
    time: '2026-07-02 22:19:30',
    type: 'Cash Reward',
    requirement: '148.85(5x)',
    progress: '0.00',
  },
  {
    amount: 'Rs 2,389.51',
    time: '2026-07-02 22:19:25',
    type: 'Cash Reward',
    requirement: '11947.55(5x)',
    progress: '0.00',
  },
  {
    amount: 'Rs 100.00',
    time: '2026-07-02 22:19:19',
    type: 'Cash Reward',
    requirement: '100.00(1x)',
    progress: '0.00',
  },
  {
    amount: 'Rs 10.00',
    time: '2026-07-02 22:19:14',
    type: 'Cash Reward',
    requirement: '10.00(1x)',
    progress: '0.00',
  },
]

export const bonusCenterRecords: BonusCenterRecord[] = [
  {
    amount: 'Rs 123.00',
    source: 'Phone Bind',
    issuedTime: '2026-06-13 18:26:04',
    orderId: 'h79rfmcasrnx8s',
    requirement: 'Valid Bet 1x',
  },
  {
    amount: 'Rs 1.00',
    source: 'Level Up Reward',
    issuedTime: '2026-06-21 16:53:27',
    orderId: 'njhyg92yq4rkqh',
    requirement: 'Valid Bet 1x',
  },
  {
    amount: 'Rs 10.00',
    source: 'Deposit Bonus',
    issuedTime: '2026-06-30 04:33:35',
    orderId: '58rhv4xevg26a3',
    requirement: 'Valid Bet 2x',
  },
  {
    amount: 'Rs 44.43',
    source: 'Download App',
    issuedTime: '2026-07-01 00:24:39',
    orderId: 'ft49b55wfh7h5v',
    requirement: 'Valid Bet 1x',
  },
]
