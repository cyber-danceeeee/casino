<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import depositTipAnimation from '../assets/lottie/lottie-deposit-tip.json'
import ContactMeBadge from '../components/ContactMeBadge.vue'
import LottiePlayer from '../components/LottiePlayer.vue'
import { logoutMockUser, mockUser, refreshMockUser } from '../mock/session'
import {
  h5Api,
  type H5BonusCenterRecordPayload,
  type H5CashAuditPayload,
  type H5DepositAmountOptionPayload,
  type H5DepositChannelPayload,
  type H5DepositConfigPayload,
  type H5DepositMethodPayload,
  type H5DepositVoucherPayload,
  type H5DepositVoucherRewardPayload,
  type H5SupportPayload,
  type H5WalletRecordPayload,
} from '../services/h5Api'
import { currencySymbol, formatCurrency, normalizePhoneInput, phoneCountryCode, phoneLength } from '../services/runtimeConfig'
import { localDepositImage, localProfileImage } from '../utils/localAssets'
import { h5SourceImage } from '../utils/staticAssets'
import { usePreferences, type AppLanguage } from '../utils/preferences'

const route = useRoute()
const router = useRouter()

const walletRecordTabs = ['Deposit', 'Withdraw', 'Statement'] as const
const betRecordTabs = ['Records', 'Report'] as const
const supportTabs = ['FAQ', 'Notice', 'Message'] as const
const fallbackSupportCategories = ['Deposit', 'Other', 'Withdrawal']
const walletTypes = ['JazzCash', 'EasyPaisa'] as const
const walletWithdrawStatusOptions = ['All Status', 'Pending', 'Succeeded', 'Failed'] as const
const walletStatementTimeOptions = ['All Time', 'Today', 'Yesterday', 'Last 7 Days', 'Last 15 Days', 'Last 30 Days'] as const
const walletStatementTypeOptions = [
  'All Categories',
  'Deposit',
  'Withdrawal',
  'Withdrawal Fee',
  'Withdrawal Reversal',
  'Bet',
  'Payout',
  'Bet Refund',
  'Payout Clawback',
  'Credit Adjustment',
  'Debit Adjustment',
  'Bonus Transfer',
  'Cash Reward',
  'Commission',
] as const
const walletStatementDirectionOptions = ['All Directions', 'Credit', 'Debit'] as const
const betTimeOptions = ['Today', 'Yesterday', 'Last 7 Days', 'Last 15 Days', 'Last 30 Days'] as const
const betStatusOptions = ['All Status', 'Settled', 'Not Settled', 'Cancelled'] as const
const betTypeOptions = ['All', 'Bingo', 'Chance', 'Fish', 'Live', 'Scratcher', 'Slot', 'Sports', 'Table'] as const
const bonusCategoryOptions = [
  'All Categories',
  'Activity Bonus',
  'VIP Bonus',
  'Compensation Bonus',
  'Incentive Rewards',
] as const
const customerImage = localProfileImage('customer.webp')
const depositPaymentMethodIcon = localDepositImage('payment-method.webp')
const depositVoucherImage = localDepositImage('voucher_ticket.webp')
const depositFirstGiftImage = localDepositImage('first_deposit_gift.webp')
const securityGiftImage = localProfileImage('security-gift.webp')
const betHelpImage = localProfileImage('profile-help.webp')
const cashWalletBgImage = localProfileImage('profile-wallet-bg.webp')
const cashWalletIcon = localProfileImage('profile-cash-wallet.webp')
const withdrawableIcon = localProfileImage('wallet-withdrawable.webp')
const cashAuditBundleImage = h5SourceImage('bonus__cash_bundle.webp')
const settingsLanguageIcon = localProfileImage('language.webp')
const settingsMusicIcon = localProfileImage('music.webp')
const supportAvatars = [
  localProfileImage('default-avatar-007.webp'),
  localProfileImage('default-avatar-008.webp'),
  localProfileImage('default-avatar-011.webp'),
  localProfileImage('default-avatar-013.webp'),
  localProfileImage('default-avatar-015.webp'),
]
const aboutImages = [
  localProfileImage('about-us-1.webp'),
  localProfileImage('about-us-2.webp'),
  localProfileImage('about-us-3.webp'),
  localProfileImage('about-us-4.webp'),
  localProfileImage('about-us-5.webp'),
]
const aboutItems = [
  {
    title: 'Unlimited Withdrawals',
    text: 'We place no withdrawal limits on regular players. Whatever you win in the game, you can freely withdraw to your bank account.',
  },
  {
    title: 'Fast and Reliable Transactions',
    text: 'Our payment system is stable and reliable, ensuring that both deposits and withdrawals are processed instantly.',
  },
  {
    title: 'Global Game Partnerships',
    text: 'We partner with world-renowned game developers, guaranteeing fair gameplay with no manipulation.',
  },
  {
    title: 'Wide Game Variety',
    text: 'With a vast selection of games and diverse gameplay options, we have something for every gaming interest.',
  },
  {
    title: '24/7 Professional Support',
    text: 'Our dedicated support team is available around the clock to resolve any issues you encounter.',
  },
].map((item, index) => ({ ...item, image: aboutImages[index] }))

interface SupportQuestionItem {
  id: number
  category: string
  question: string
  answer: string
}

interface SupportNoticeItem {
  id: string
  title: string
  contentHTML: string
  createdAt?: string
  read: boolean
}

interface BonusCenterItem {
  id: number
  amount: string
  source: string
  category: string
  issuedTime: string
  orderId: string
  requirement: string
  status: string
}

interface WalletRecordItem {
  id: number
  title: string
  source: string
  amount: string
  balance: string
  time: string
  orderNumber: string
  status: string
  direction: string
  sign: string
  gameName: string
  providerCode: string
}

interface CashAuditItem {
  id: number
  amount: string
  createdAt: string
  type: string
  auditRequirement: string
  progressAmount: string
  requiredAmount: string
  progressPercent: number
}

interface DepositMethodItem extends H5DepositMethodPayload {
  displayIcon: string
}

interface DepositChannelItem extends H5DepositChannelPayload {
  displayIcon: string
}

interface DepositAmountOptionItem extends H5DepositAmountOptionPayload {
  amountValue: number
  bonusRewardValue: number
  cashRewardValue: number
  rewardAmountValue: number
}

interface DepositVoucherRewardItem extends H5DepositVoucherRewardPayload {
  amountValue: number
  bonusRewardValue: number
  cashRewardValue: number
  rewardAmountValue: number
}

interface DepositVoucherItem extends H5DepositVoucherPayload {
  displayIcon: string
  minAmountValue: number
  maxAmountValue: number
  maxRewardValue: number
  bonusRewardValue: number
  cashRewardValue: number
  rewardAmountValue: number
  rewards: DepositVoucherRewardItem[]
}

const fallbackSupportQuestionGroups: Record<string, string[]> = {
  Deposit: [
    'How do I make a deposit?',
    'How long does it take for my deposit to be credited?',
    'Why did my deposit fail?',
    'My deposit has not been credited after 24 hours.',
    'I still can’t make a deposit after several attempts.',
  ],
  Other: [
    'How can I become an agent?',
    'How can I change my bank account?',
    'The game is not working properly.',
  ],
  Withdrawal: [
    'How do I withdraw my funds?',
    'What are the withdrawal requirements?',
    'Why did my withdrawal fail and the funds were returned?',
    'My withdrawal has been rejected multiple times.',
    'Why is my withdrawal still processing?',
    'Why can’t I withdraw my funds?',
    'What does “Withdrawal Returned” mean?',
    'My withdrawal has not been received.',
    'Why is my withdrawal under review?',
  ],
}

const supportFaqAnswers: Record<string, string> = {
  'How do I make a deposit?':
    `<p>Go to <strong>Profile &gt; Deposit</strong>, add your bank account, then choose the amount and payment method to complete your deposit.</p><p><strong>Minimum Deposit:</strong> ${formatCurrency(100, { includeCode: true })}</p><p><strong>Maximum Deposit (per transaction):</strong> ${formatCurrency(50000, { includeCode: true })}</p>`,
  'How can I become an agent?':
    '<p>Tap the <strong>Share</strong> button at the bottom of the game homepage to get your <strong>exclusive referral link</strong> and <strong>QR code</strong>.</p><p>When your friends register through your link and make a deposit, you will earn referral commissions.</p><p>For more details, please visit the <strong>Share</strong> page.</p>',
  'How do I withdraw my funds?':
    `<p>Go to <strong>Profile &gt; Withdrawal</strong>, add your bank account, enter the withdrawal amount, and submit your request.</p><ul><li><p><strong>Minimum Withdrawal:</strong> ${formatCurrency(100, { includeCode: true })}</p></li><li><p><strong>Maximum Withdrawal (per transaction):</strong> ${formatCurrency(50000, { includeCode: true })}</p></li></ul>`,
  'How long does it take for my deposit to be credited?':
    '<p>Most deposits are credited to your game account within <strong>1-2 minutes</strong>.</p><p>If your bank is processing the transaction slowly, it may take up to <strong>30 minutes</strong>.</p><p>In most cases, deposits are completed within <strong>12 hours</strong>. Some transactions may take up to <strong>24 hours</strong>.</p><p>If your deposit has not been credited after 24 hours, please contact our <strong>Online Customer Support</strong> and provide your payment receipt or screenshot for verification.</p>',
  'How can I change my bank account?':
    '<p>For your account security, bank account information cannot be changed directly.</p><p>If you need to update your bank account, please contact <strong>Online Customer Support</strong>, and we will assist you.</p>',
  'What are the withdrawal requirements?':
    '<p>Normally, you only need to complete the required <strong>valid betting turnover</strong> before making a withdrawal.</p><p>If you have claimed a deposit bonus or any promotional rewards, you must first complete the required <strong>turnover requirement</strong> before your withdrawal can be processed.</p>',
  'Why did my deposit fail?':
    '<p>Due to high transaction volume, the payment channel may be temporarily busy.</p><p>Please wait a few minutes and try again.</p><p>If the issue continues, please contact <strong>Online Customer Support</strong> for assistance.</p>',
  'Why did my withdrawal fail and the funds were returned?':
    '<p>Your withdrawal was returned by the bank, and the funds have been credited back to your game account.</p><p>Please make sure your bank account information is correct, then submit a new withdrawal request.</p>',
  'The game is not working properly.':
    '<p>If you experience any of the following issues:</p><ul><li><p>Unable to log in</p></li><li><p>Game crashes</p></li><li><p>Unable to enter the game</p></li><li><p>Loading errors</p></li></ul><p>Please contact <strong>Online Customer Support</strong> and send us a screenshot of the issue. We will help you resolve it as soon as possible.</p>',
  'My withdrawal has been rejected multiple times.':
    '<p>If your withdrawal keeps being returned, please check that:</p><ul><li><p>Your account name is correct.</p></li><li><p>Your bank account number is correct.</p></li><li><p>Your bank details are complete and accurate.</p></li></ul><p>If everything is correct but the problem continues, please contact <strong>Online Customer Support</strong>.</p>',
  'My deposit has not been credited after 24 hours.':
    '<p>If your deposit has not been received after <strong>24 hours</strong>, please contact our <strong>Online Customer Support</strong> immediately.</p><p>Please provide:</p><ul><li><p>Payment receipt or screenshot</p></li><li><p>Bank transaction record (video or screenshot)</p></li></ul><p>Our team will verify your transaction as soon as possible. <strong>Don\'t worry-your funds are safe.</strong></p>',
  'Why is my withdrawal still processing?':
    '<p>Your withdrawal request has already been submitted to the bank.</p><p>Some banks may take longer to process transactions, so please be patient.</p><p>If the processing time is unusually long, please contact <strong>Online Customer Support</strong> to check your withdrawal status.</p>',
  'I still can’t make a deposit after several attempts.':
    '<p>Please try the following:</p><ul><li><p>Try a different deposit amount.</p></li><li><p>Use another available payment method (if applicable).</p></li><li><p>If the problem continues, take a screenshot of the error page and send it to <strong>Online Customer Support</strong> for assistance.</p></li></ul>',
  'Why can’t I withdraw my funds?':
    '<p>This may happen for one of the following reasons:</p><p>· Your withdrawable balance is insufficient.</p><p>· You have not completed the required betting turnover.</p><p>· You have reached the withdrawal limit.</p><p>Please check these conditions before submitting another withdrawal request.</p>',
  'What does “Withdrawal Returned” mean?':
    '<p>If the bank cannot complete the transfer, your withdrawal will be returned to your game account automatically.</p><p>Please verify your bank account information and submit a new withdrawal request.</p>',
  'My withdrawal has not been received.':
    '<p>Please check the status of your withdrawal order.</p><ul><li><p>If the status is <strong>Processing</strong>, please wait patiently.</p></li><li><p>If the status is <strong>Returned</strong>, the funds have already been credited back to your game account.</p></li><li><p>If the status has not been updated for a long time, please contact <strong>Online Customer Support</strong>.</p></li></ul>',
  'Why is my withdrawal under review?':
    '<p>Due to a high number of withdrawal requests, the verification process may take a little longer than usual.</p><p>Our team is processing all requests as quickly as possible. Thank you for your patience.</p>',
}

const fallbackFaqQuestions = [
  fallbackSupportQuestionGroups.Deposit[0],
  fallbackSupportQuestionGroups.Other[0],
  fallbackSupportQuestionGroups.Withdrawal[0],
  fallbackSupportQuestionGroups.Deposit[1],
  fallbackSupportQuestionGroups.Withdrawal[1],
  fallbackSupportQuestionGroups.Other[1],
  fallbackSupportQuestionGroups.Deposit[2],
  fallbackSupportQuestionGroups.Withdrawal[2],
  fallbackSupportQuestionGroups.Other[2],
  fallbackSupportQuestionGroups.Withdrawal[3],
  fallbackSupportQuestionGroups.Deposit[3],
  fallbackSupportQuestionGroups.Withdrawal[4],
  fallbackSupportQuestionGroups.Deposit[4],
  fallbackSupportQuestionGroups.Withdrawal[5],
  fallbackSupportQuestionGroups.Withdrawal[6],
  fallbackSupportQuestionGroups.Withdrawal[7],
  fallbackSupportQuestionGroups.Withdrawal[8],
]

const fallbackSupportQuestions: SupportQuestionItem[] = fallbackFaqQuestions.map((question, index) => {
  const category =
    Object.entries(fallbackSupportQuestionGroups).find(([, questions]) => questions.includes(question))?.[0] ||
    'Other'

  return {
    id: index + 1,
    category,
    question,
    answer: supportFaqAnswers[question] || '',
  }
})

const agreementItems = [
  {
    kind: 'lead',
    text: 'Effective Date: May 30, 2026',
  },
  {
    kind: 'paragraph',
    text: 'Welcome to 56mvp.com! This User Agreement (the "Agreement") is a legal contract between you ("User," "you," or "your") and 56mvp.com (the legal operating license holder for the 56mvp.com card and board game application, hereinafter "we," "us," or "our") regarding your use of our applications, products/services, and websites (collectively, the "Service").',
  },
  {
    kind: 'strong',
    text: 'By accessing, using, or purchasing our Service, you agree to be bound by the terms and conditions of this Agreement.',
  },
  {
    kind: 'paragraph',
    text: 'Please read it carefully. If you do not agree to these terms, do not access or use our Service.',
  },
  {
    kind: 'strong',
    text: 'Important Note: 56mvp.com is a card and board game application. This Service is exclusively intended for and available to users located within Pakistan. If you are not in Pakistan, please discontinue use of this Service immediately.',
  },
  { kind: 'heading', text: 'Acceptance of Terms' },
  {
    kind: 'paragraph',
    text: 'By using our Service, you confirm that you are at least 21 years old. You agree to comply with all applicable laws and regulations in Pakistan regarding your use of the Service.',
  },
  { kind: 'heading', text: 'Changes to This Agreement' },
  {
    kind: 'paragraph',
    text: 'We may update this Agreement from time to time to reflect changes in our Service, business practices, or legal requirements. We will notify you of any significant changes by posting the updated Agreement on our website or within the Service, and updating the "Effective Date" at the top. Your continued use of the Service after the effective date of the revised Agreement constitutes your acceptance of the changes.',
  },
  { kind: 'heading', text: 'Your Account' },
  { kind: 'subheading', text: '3.1 Account Creation' },
  {
    kind: 'paragraph',
    text: 'To play 56mvp.com card and board games and access certain features of our Service, you may need to create an account. You agree to provide accurate, current, and complete information during the registration process and to keep your account information updated.',
  },
  { kind: 'subheading', text: '3.2 Account Security' },
  {
    kind: 'paragraph',
    text: 'You are responsible for maintaining the confidentiality of your account password and for all activities that occur under your account. You agree to notify us immediately of any unauthorized use of your account or any other breach of security.',
  },
  { kind: 'heading', text: 'Use of the Service' },
  { kind: 'subheading', text: '4.1 Permitted Use' },
  {
    kind: 'paragraph',
    text: 'You are granted a limited, non-exclusive, non-transferable, and revocable license to use the 56mvp.com card and board game service for your personal, non-commercial use, subject to the terms of this Agreement.',
  },
  { kind: 'subheading', text: '4.2 Prohibited Conduct' },
  {
    kind: 'paragraph',
    text: 'You agree not to use the Service for unlawful purposes, disrupt the fairness of games, attempt unauthorized access, exploit game content without permission, transmit destructive code, or gain unfair advantage through cheating software, bots, or bugs.',
  },
  { kind: 'heading', text: 'Intellectual Property' },
  {
    kind: 'paragraph',
    text: 'All content, features, and functionality within the Service are the exclusive property of 56mvp.com or its licensors and are protected by copyright, trademark, patent, trade secret, and other intellectual property laws.',
  },
  { kind: 'heading', text: 'Disclaimers' },
  {
    kind: 'strong',
    text: 'THE SERVICE IS PROVIDED ON AN "AS IS" AND "AS AVAILABLE" BASIS, WITHOUT ANY WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED.',
  },
  {
    kind: 'paragraph',
    text: 'Given that this Service is a card and board game, you understand and accept that game outcomes are uncertain and may be influenced by factors such as network connectivity and device performance.',
  },
  { kind: 'heading', text: 'Limitation of Liability' },
  {
    kind: 'paragraph',
    text: 'To the fullest extent permitted by applicable law, 56mvp.com and its affiliates shall not be liable for indirect, incidental, special, consequential, or punitive damages resulting from your access to or use of the Service.',
  },
  { kind: 'heading', text: 'Contact Us' },
  {
    kind: 'paragraph',
    text: 'If you have any questions about this Agreement, please contact us.',
  },
]

const activeWalletRecordTab = ref<(typeof walletRecordTabs)[number]>('Deposit')
const subpageContentRef = ref<HTMLElement | null>(null)
const walletRecordItems = ref<WalletRecordItem[]>([])
const walletRecordsLoading = ref(false)
const walletRecordsLoadingMore = ref(false)
const walletRecordPage = ref(1)
const walletRecordTotal = ref(0)
const activeBetRecordTab = ref<(typeof betRecordTabs)[number]>('Records')
const activeSupportTab = ref<(typeof supportTabs)[number]>('FAQ')
const supportCategories = ref<string[]>([...fallbackSupportCategories])
const supportQuestionItems = ref<SupportQuestionItem[]>(fallbackSupportQuestions)
const supportNoticeItems = ref<SupportNoticeItem[]>([])
const activeSupportCategory = ref<string | null>(null)
const activeSupportQuestion = ref<string | null>(null)
const activeSupportNotice = ref<SupportNoticeItem | null>(null)
const supportPanelRef = ref<HTMLElement | null>(null)
const supportScrollTop = ref(0)
const activeWithdrawForm = ref<'bank' | 'wallet' | null>(null)
const selectedWalletType = ref<'JazzCash' | 'EasyPaisa'>('JazzCash')
const selectedWalletWithdrawStatus = ref<(typeof walletWithdrawStatusOptions)[number]>('All Status')
const selectedWalletStatementTime = ref<(typeof walletStatementTimeOptions)[number]>('All Time')
const selectedWalletStatementType = ref<(typeof walletStatementTypeOptions)[number]>('All Categories')
const selectedWalletStatementDirection = ref<(typeof walletStatementDirectionOptions)[number]>('All Directions')
const selectedBetTime = ref<(typeof betTimeOptions)[number]>('Today')
const selectedBetStatus = ref<(typeof betStatusOptions)[number]>('All Status')
const selectedBetType = ref<(typeof betTypeOptions)[number]>('All')
const selectedReportDate = ref<(typeof betTimeOptions)[number]>('Today')
const selectedReportGameType = ref<(typeof betTypeOptions)[number]>('All')

let walletRecordLoadSequence = 0
let walletRecordLoadCheckFrame = 0

const walletRecordsHasMore = computed(() => walletRecordItems.value.length < walletRecordTotal.value)
const securityPhoneMode = ref<'reset' | 'verify'>('reset')
const securityNewPhone = ref('')
const securityPhoneCode = ref('')
const securityPhonePassword = ref('')
const securityPhonePasswordVisible = ref(false)
const loginPasswordNew = ref('')
const loginPasswordConfirm = ref('')
const loginPasswordCode = ref('')
const loginPasswordVisible = ref(false)
const loginPasswordConfirmVisible = ref(false)
const showOrderDescription = ref(false)
const activeCashTab = ref<'audit' | 'records'>('audit')
const cashAuditItems = ref<CashAuditItem[]>([])
const cashAuditLoading = ref(false)
const selectedBonusCategory = ref<(typeof bonusCategoryOptions)[number]>('All Categories')
const bonusCenterRecords = ref<BonusCenterItem[]>([])
const bonusCenterLoading = ref(false)
const showDepositVoucherSheet = ref(false)
const depositPanelRef = ref<HTMLElement | null>(null)
const depositVoucherStripRef = ref<HTMLElement | null>(null)
const depositMethods = ref<DepositMethodItem[]>([])
const depositChannels = ref<DepositChannelItem[]>([])
const amountOptions = ref<DepositAmountOptionItem[]>([])
const depositVoucherOptions = ref<DepositVoucherItem[]>([])
const selectedDepositMethodCode = ref('')
const selectedDepositChannelCode = ref('')
const depositAmount = ref('')
const selectedDepositVoucherIndex = ref(0)
const depositReturnExpanded = ref(false)
const depositSubmitting = ref(false)
const showDepositProgressDialog = ref(false)
const depositProgressOrderNo = ref('')
const activePicker = ref<
  | 'walletWithdrawStatus'
  | 'walletStatementTime'
  | 'walletStatementType'
  | 'walletStatementDirection'
  | 'betTime'
  | 'betStatus'
  | 'betType'
  | 'reportDate'
  | 'reportGameType'
  | 'bonusCategory'
  | 'language'
  | null
>(null)
const { currentLanguage, languageOptions, musicEnabled, setLanguage, toggleMusicEnabled } = usePreferences()

const pageName = computed(() => String(route.name ?? ''))

const titleMap: Record<string, string> = {
  'wallet-records': 'Records',
  'wallet-deposit': 'Deposit',
  'cash-center': 'Cash Wallet Details',
  'bonus-center': 'Bonus Center',
  'bet-records': 'Bet Records',
  security: 'Security Center',
  'bank-account': 'Withdrawal Account',
  'security-phone': 'Link Account',
  'security-password': 'Login Password',
  support: 'Support',
  about: 'About Us',
  'user-agreement': 'User Agreement',
  settings: 'Settings',
}

const pageTitle = computed(() => titleMap[pageName.value] ?? 'Profile')
const showContact = computed(() => !['support', 'user-agreement'].includes(pageName.value))
const showHeaderBell = computed(() => pageName.value === 'settings')
const showBetHelp = computed(() => pageName.value === 'bet-records')
const runtimePhoneCode = computed(() => phoneCountryCode())
const runtimePhoneLength = computed(() => phoneLength())
const securityPhoneNumber = computed(() => normalizePhoneInput(mockUser.value.phone) || '3012345677')
const cashWalletBalance = computed(() => mockUser.value.cash || '0.00')
const withdrawableWalletBalance = computed(() => mockUser.value.withdrawable || '0.00')
const securityPhoneTarget = computed(() =>
  securityPhoneMode.value === 'reset' ? normalizePhoneInput(securityNewPhone.value) : securityPhoneNumber.value,
)
const securityPhoneSendDisabled = computed(() => securityPhoneTarget.value.length !== runtimePhoneLength.value)
const loginPasswordSendDisabled = computed(() => securityPhoneNumber.value.length !== runtimePhoneLength.value)
watch(securityNewPhone, (value) => {
  const next = normalizePhoneInput(value)
  if (next !== value) securityNewPhone.value = next
})
const visibleFaqQuestions = computed(() => {
  const questions = activeSupportCategory.value
    ? supportQuestionItems.value.filter((item) => item.category === activeSupportCategory.value)
    : supportQuestionItems.value

  return questions.map((item) => item.question)
})
const activeSupportAnswer = computed(() => {
  if (!activeSupportQuestion.value) return ''
  if (activeSupportNotice.value) return activeSupportNotice.value.contentHTML
  return (
    supportQuestionItems.value.find((item) => item.question === activeSupportQuestion.value)?.answer ||
    supportFaqAnswers[activeSupportQuestion.value] ||
    ''
  )
})
const isSupportHeaderCollapsed = computed(() => pageName.value === 'support' && supportScrollTop.value > 42)
const parseDepositMoney = (value: unknown) => {
  const parsed = Number(String(value ?? '').replace(/,/g, ''))
  return Number.isFinite(parsed) ? parsed : 0
}
const formatDepositNumber = (value: number) =>
  value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
const formatDepositInputAmount = (value: number) => {
  if (!Number.isFinite(value) || value <= 0) return ''
  if (Number.isInteger(value)) return String(value)
  return String(value)
}
const selectedDepositVoucher = computed(() => depositVoucherOptions.value[selectedDepositVoucherIndex.value])
const depositAmountValue = computed(() => {
  const value = Number(String(depositAmount.value).replace(/,/g, ''))
  return Number.isFinite(value) ? value : 0
})
const isDepositVoucherAvailableForAmount = (voucher: DepositVoucherItem | undefined, amount: number) => {
  if (!voucher || amount <= 0) return false
  const min = voucher.minAmountValue || 0
  const max = voucher.maxAmountValue || Number.POSITIVE_INFINITY
  return amount >= min && amount <= max
}
const getDepositVoucherRewardAmount = (voucher: DepositVoucherItem | undefined, amount: number) => {
  if (!isDepositVoucherAvailableForAmount(voucher, amount)) return 0
  return getDepositVoucherRewardDetail(voucher, amount).rewardAmountValue
}
const getDepositVoucherRewardDetail = (voucher: DepositVoucherItem | undefined, amount: number) => {
  const empty = { bonusRewardValue: 0, cashRewardValue: 0, rewardAmountValue: 0 }
  if (!isDepositVoucherAvailableForAmount(voucher, amount)) return empty

  if (voucher?.rewards?.length) {
    const exactReward = voucher.rewards.find((reward) => Math.abs(reward.amountValue - amount) < 0.01)
    return exactReward
      ? {
          bonusRewardValue: exactReward.bonusRewardValue,
          cashRewardValue: exactReward.cashRewardValue,
          rewardAmountValue: exactReward.rewardAmountValue,
        }
      : empty
  }

  return {
    bonusRewardValue: voucher?.bonusRewardValue ?? 0,
    cashRewardValue: voucher?.cashRewardValue ?? 0,
    rewardAmountValue: voucher?.rewardAmountValue ?? 0,
  }
}
const getDepositAmountOptionReward = (option: DepositAmountOptionItem) =>
  getDepositVoucherRewardAmount(selectedDepositVoucher.value, option.amountValue)
const selectedDepositReward = computed(() =>
  getDepositVoucherRewardDetail(selectedDepositVoucher.value, depositAmountValue.value),
)
const depositBonusReward = computed(() => selectedDepositReward.value.bonusRewardValue)
const depositCashReward = computed(() => selectedDepositReward.value.cashRewardValue)
const depositTotalGet = computed(() => depositAmountValue.value + depositBonusReward.value + depositCashReward.value)
const depositReturnDetails = computed(() => [
  { label: 'Deposit Amount:', value: depositAmountValue.value, dot: 'is-white' },
  ...(depositBonusReward.value > 0
    ? [{ label: 'Bonus Rewards:', value: depositBonusReward.value, dot: 'is-violet' }]
    : []),
  ...(depositCashReward.value > 0
    ? [{ label: 'Cash Rewards:', value: depositCashReward.value, dot: 'is-amber' }]
    : []),
])

const formatDepositMoney = (value: number) => formatCurrency(value)

const bonusCategoryCodes: Record<(typeof bonusCategoryOptions)[number], string> = {
  'All Categories': 'all',
  'Activity Bonus': 'activity',
  'VIP Bonus': 'vip',
  'Compensation Bonus': 'compensation',
  'Incentive Rewards': 'incentive',
}

const isDepositAmountActive = (option: DepositAmountOptionItem) =>
  Math.abs(option.amountValue - depositAmountValue.value) < 0.01

const selectDepositMethod = (methodCode: string) => {
  selectedDepositMethodCode.value = methodCode
  const matchedChannel = depositChannels.value.find((channel) => channel.methodCode === methodCode)
  if (matchedChannel) selectedDepositChannelCode.value = matchedChannel.channelCode
}

const selectDepositChannel = (channelCode: string) => {
  selectedDepositChannelCode.value = channelCode
}

const selectDepositAmount = (option: DepositAmountOptionItem) => {
  depositAmount.value = formatDepositInputAmount(option.amountValue)
}

const selectDepositVoucher = (index: number) => {
  if (!depositVoucherOptions.value[index]?.available) return
  selectedDepositVoucherIndex.value = index
  void nextTick().then(() => scrollSelectedDepositVoucherIntoView())
}

const openDepositVoucherSheet = () => {
  showDepositVoucherSheet.value = true
}

const closeDepositVoucherSheet = () => {
  showDepositVoucherSheet.value = false
}

const selectDepositVoucherFromSheet = (index: number) => {
  selectDepositVoucher(index)
  if (depositVoucherOptions.value[index]?.available) closeDepositVoucherSheet()
}

const scrollSelectedDepositVoucherIntoView = (behavior: ScrollBehavior = 'smooth') => {
  const strip = depositVoucherStripRef.value
  if (!strip) return
  const target = strip.querySelectorAll<HTMLElement>('.voucher-ticket')[selectedDepositVoucherIndex.value]
  if (!target) return
  strip.scrollTo({
    left: target.offsetLeft,
    behavior,
  })
}

const scrollDepositPanelToBottom = () => {
  const panel = depositPanelRef.value
  if (!panel) return
  panel.scrollTo({ top: panel.scrollHeight, behavior: 'smooth' })
}

const closeDepositProgressDialog = () => {
  showDepositProgressDialog.value = false
}

const viewDepositOrder = () => {
  closeDepositProgressDialog()
  void router.push({ name: 'wallet-records', query: { tab: 'Deposit' } })
}

const submitDeposit = async () => {
  if (depositSubmitting.value || depositAmountValue.value <= 0) return

  depositSubmitting.value = true
  try {
    const payload = await h5Api.createDeposit({
      amount: String(depositAmountValue.value),
      methodCode: selectedDepositMethodCode.value,
      channelCode: selectedDepositChannelCode.value,
      voucherCode: selectedDepositVoucher.value?.voucherCode,
    })
    depositProgressOrderNo.value = payload.order?.orderNo || ''
    showDepositProgressDialog.value = true
  } catch (error) {
    console.warn('[api] failed to create deposit order.', error)
  } finally {
    depositSubmitting.value = false
  }
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }

  void router.push({ name: 'profile' })
}

const openContact = () => {
  void router.push({ name: 'support' })
}

const selectWalletRecordTab = (tab: (typeof walletRecordTabs)[number]) => {
  activeWalletRecordTab.value = tab
  activePicker.value = null
  void loadWalletRecords()
}

const applyWalletRecordTabFromRoute = () => {
  if (pageName.value !== 'wallet-records') return
  const queryTab = String(route.query.tab || '')
  const matched = walletRecordTabs.find((tab) => tab.toLowerCase() === queryTab.toLowerCase())
  if (matched) activeWalletRecordTab.value = matched
}

const selectBetRecordTab = (tab: (typeof betRecordTabs)[number]) => {
  activeBetRecordTab.value = tab
  activePicker.value = null
}

const resetSupportPanelScroll = () => {
  supportScrollTop.value = 0
  supportPanelRef.value?.scrollTo({ top: 0 })
}

const handleSupportPanelScroll = (event: Event) => {
  supportScrollTop.value = (event.currentTarget as HTMLElement).scrollTop
}

const selectSupportTab = (tab: (typeof supportTabs)[number]) => {
  activeSupportTab.value = tab
  activeSupportQuestion.value = null
  activeSupportNotice.value = null
  resetSupportPanelScroll()
  if (tab === 'Notice') void loadSupport()
}

const selectSupportCategory = (category: string) => {
  activeSupportCategory.value = activeSupportCategory.value === category ? null : category
  activeSupportQuestion.value = null
  activeSupportNotice.value = null
  resetSupportPanelScroll()
}

const openSupportAnswer = (question: string) => {
  activeSupportNotice.value = null
  activeSupportQuestion.value = question
}

const openSupportNotice = (notice: SupportNoticeItem) => {
  activeSupportNotice.value = notice
  activeSupportQuestion.value = notice.title
}

const closeSupportAnswer = () => {
  activeSupportQuestion.value = null
  activeSupportNotice.value = null
}

const openBankAccount = () => {
  void router.push({ name: 'bank-account' })
}

const openSecurityPhone = () => {
  void router.push({ name: 'security-phone' })
}

const openSecurityPassword = () => {
  void router.push({ name: 'security-password' })
}

const openWithdrawForm = (type: 'bank' | 'wallet') => {
  activeWithdrawForm.value = type
}

const activePickerTitle = computed(() => {
  return activePicker.value ? 'Select' : ''
})

const activePickerOptions = computed<readonly string[]>(() => {
  if (activePicker.value === 'walletWithdrawStatus') return walletWithdrawStatusOptions
  if (activePicker.value === 'walletStatementTime') return walletStatementTimeOptions
  if (activePicker.value === 'walletStatementType') return walletStatementTypeOptions
  if (activePicker.value === 'walletStatementDirection') return walletStatementDirectionOptions
  if (activePicker.value === 'betTime') return betTimeOptions
  if (activePicker.value === 'betStatus') return betStatusOptions
  if (activePicker.value === 'betType') return betTypeOptions
  if (activePicker.value === 'reportDate') return betTimeOptions
  if (activePicker.value === 'reportGameType') return betTypeOptions
  if (activePicker.value === 'bonusCategory') return bonusCategoryOptions
  if (activePicker.value === 'language') return languageOptions
  return []
})

const activePickerValue = computed(() => {
  if (activePicker.value === 'walletWithdrawStatus') return selectedWalletWithdrawStatus.value
  if (activePicker.value === 'walletStatementTime') return selectedWalletStatementTime.value
  if (activePicker.value === 'walletStatementType') return selectedWalletStatementType.value
  if (activePicker.value === 'walletStatementDirection') return selectedWalletStatementDirection.value
  if (activePicker.value === 'betTime') return selectedBetTime.value
  if (activePicker.value === 'betStatus') return selectedBetStatus.value
  if (activePicker.value === 'betType') return selectedBetType.value
  if (activePicker.value === 'reportDate') return selectedReportDate.value
  if (activePicker.value === 'reportGameType') return selectedReportGameType.value
  if (activePicker.value === 'bonusCategory') return selectedBonusCategory.value
  if (activePicker.value === 'language') return currentLanguage.value
  return ''
})

const openPicker = (picker: NonNullable<typeof activePicker.value>) => {
  activePicker.value = picker
}

const closePicker = () => {
  activePicker.value = null
}

const openOrderDescription = () => {
  showOrderDescription.value = true
}

const closeOrderDescription = () => {
  showOrderDescription.value = false
}

const selectPickerOption = (option: string) => {
  if (activePicker.value === 'walletWithdrawStatus') selectedWalletWithdrawStatus.value = option as (typeof walletWithdrawStatusOptions)[number]
  if (activePicker.value === 'walletStatementTime') selectedWalletStatementTime.value = option as (typeof walletStatementTimeOptions)[number]
  if (activePicker.value === 'walletStatementType') selectedWalletStatementType.value = option as (typeof walletStatementTypeOptions)[number]
  if (activePicker.value === 'walletStatementDirection') selectedWalletStatementDirection.value = option as (typeof walletStatementDirectionOptions)[number]
  if (activePicker.value === 'betTime') selectedBetTime.value = option as (typeof betTimeOptions)[number]
  if (activePicker.value === 'betStatus') selectedBetStatus.value = option as (typeof betStatusOptions)[number]
  if (activePicker.value === 'betType') selectedBetType.value = option as (typeof betTypeOptions)[number]
  if (activePicker.value === 'reportDate') selectedReportDate.value = option as (typeof betTimeOptions)[number]
  if (activePicker.value === 'reportGameType') selectedReportGameType.value = option as (typeof betTypeOptions)[number]
  if (activePicker.value === 'bonusCategory') selectedBonusCategory.value = option as (typeof bonusCategoryOptions)[number]
  if (activePicker.value === 'language') setLanguage(option as AppLanguage)
  closePicker()
}

const selectWalletType = (type: (typeof walletTypes)[number]) => {
  selectedWalletType.value = type
}

const closeWithdrawForm = () => {
  activeWithdrawForm.value = null
}

const logOut = () => {
  logoutMockUser()
  void router.push({ name: 'home' })
}

const applySupportPayload = (payload: H5SupportPayload) => {
  const categoryById = new Map<number, string>()
  const categories = (payload.categories || [])
    .filter((category) => category.name)
    .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0))
    .map((category) => {
      if (category.id) categoryById.set(Number(category.id), category.name)
      return category.name
    })

  const questions = (payload.questions || [])
    .filter((question) => question.question)
    .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0))
    .map((question) => ({
      id: Number(question.id),
      category: categoryById.get(Number(question.categoryId)) || 'Other',
      question: question.question,
      answer: question.answerHTML || question.answer || '',
    }))

  if (categories.length) supportCategories.value = categories
  if (questions.length) supportQuestionItems.value = questions
  supportNoticeItems.value = (payload.notices || [])
    .filter((notice) => notice.title || notice.contentHTML)
    .map((notice) => ({
      id: notice.id,
      title: notice.title || 'Notice',
      contentHTML: notice.contentHTML || '',
      createdAt: notice.createdAt,
      read: notice.read !== false,
    }))
}

const loadSupport = async () => {
  try {
    applySupportPayload(await h5Api.getSupport('all'))
    if (activeSupportTab.value === 'Notice') void markSupportNoticesRead()
  } catch (error) {
    console.warn('[api] falling back to local support FAQ data.', error)
  }
}

const markSupportNoticesRead = async () => {
  const ids = supportNoticeItems.value
    .filter((notice) => notice.id.startsWith('notification:') && notice.read === false)
    .map((notice) => notice.id)
  if (!ids.length) return

  try {
    await h5Api.markSupportNoticesRead(ids)
    supportNoticeItems.value = supportNoticeItems.value.map((notice) =>
      ids.includes(notice.id) ? { ...notice, read: true } : notice,
    )
    window.dispatchEvent(new Event('h5-notifications-updated'))
  } catch (error) {
    console.warn('[api] failed to mark support notices read.', error)
  }
}

const formatSupportNoticeTime = (value?: string) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString()
}

const applySupportTabFromRoute = () => {
  if (pageName.value !== 'support') return
  const tab = String(route.query.tab || '').toLowerCase()
  if (tab === 'notice') activeSupportTab.value = 'Notice'
  if (tab === 'message') activeSupportTab.value = 'Message'
}

const applyDepositConfig = (payload: H5DepositConfigPayload) => {
  depositMethods.value = (payload.methods || [])
    .filter((method) => method.name)
    .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0))
    .map((method) => ({
      ...method,
      displayIcon: method.iconURL || depositPaymentMethodIcon,
    }))

  depositChannels.value = (payload.channels || [])
    .filter((channel) => channel.name)
    .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0))
    .map((channel) => ({
      ...channel,
      displayIcon: depositMethods.value.find((method) => method.methodCode === channel.methodCode)?.displayIcon || depositPaymentMethodIcon,
    }))

  amountOptions.value = (payload.amountOptions || [])
    .filter((option) => option.amount)
    .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0))
    .map((option) => {
      const bonusRewardValue = parseDepositMoney(option.bonusReward)
      const cashRewardValue = parseDepositMoney(option.cashReward)
      return {
        ...option,
        amountValue: parseDepositMoney(option.amount),
        bonusRewardValue,
        cashRewardValue,
        rewardAmountValue: parseDepositMoney(option.rewardAmount) || bonusRewardValue + cashRewardValue,
      }
    })

  depositVoucherOptions.value = (payload.vouchers || [])
    .filter((voucher) => voucher.title)
    .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0))
    .map((voucher) => {
      const bonusRewardValue = parseDepositMoney(voucher.bonusReward)
      const cashRewardValue = parseDepositMoney(voucher.cashReward)
      const rewards = (voucher.rewards || [])
        .map((reward) => {
          const rewardBonusValue = parseDepositMoney(reward.bonusReward)
          const rewardCashValue = parseDepositMoney(reward.cashReward)
          return {
            ...reward,
            amountValue: parseDepositMoney(reward.amount),
            bonusRewardValue: rewardBonusValue,
            cashRewardValue: rewardCashValue,
            rewardAmountValue: parseDepositMoney(reward.rewardAmount) || rewardBonusValue + rewardCashValue,
          }
        })
        .sort((a, b) => Number(a.sort || 0) - Number(b.sort || 0))
      return {
        ...voucher,
        displayIcon: voucher.iconURL || depositFirstGiftImage,
        minAmountValue: parseDepositMoney(voucher.minAmount),
        maxAmountValue: parseDepositMoney(voucher.maxAmount),
        maxRewardValue: parseDepositMoney(voucher.maxReward),
        bonusRewardValue,
        cashRewardValue,
        rewardAmountValue: parseDepositMoney(voucher.rewardAmount) || bonusRewardValue + cashRewardValue,
        rewards,
      }
    })

  if (!selectedDepositMethodCode.value && depositMethods.value[0]) {
    selectedDepositMethodCode.value = depositMethods.value[0].methodCode
  }
  if (!selectedDepositChannelCode.value) {
    const firstChannel =
      depositChannels.value.find((channel) => channel.methodCode === selectedDepositMethodCode.value) ||
      depositChannels.value[0]
    selectedDepositChannelCode.value = firstChannel?.channelCode || ''
  }
  if (!depositAmount.value && amountOptions.value[0]) {
    depositAmount.value = formatDepositInputAmount(amountOptions.value[0].amountValue)
  }
  const firstAvailableVoucherIndex = depositVoucherOptions.value.findIndex((voucher) => voucher.available)
  selectedDepositVoucherIndex.value = firstAvailableVoucherIndex >= 0 ? firstAvailableVoucherIndex : 0
}

const loadDepositConfig = async () => {
  try {
    applyDepositConfig(await h5Api.getDepositConfig())
  } catch (error) {
    console.warn('[api] failed to load deposit config.', error)
    depositMethods.value = []
    depositChannels.value = []
    amountOptions.value = []
    depositVoucherOptions.value = []
  }
}

const toBonusCenterItem = (record: H5BonusCenterRecordPayload): BonusCenterItem => ({
  id: record.id,
  amount: formatCurrency(record.amount),
  source: record.source || record.bonusType || record.category,
  category: record.category,
  issuedTime: record.issuedTime || '--',
  orderId: record.orderId,
  requirement: record.requirement,
  status: record.status,
})

const walletTabCode = (tab: (typeof walletRecordTabs)[number]) => tab.toLowerCase()

const walletStatementTypeCode = (value: (typeof walletStatementTypeOptions)[number]) => {
  if (value === 'All Categories') return undefined
  if (value === 'Cash Reward') return 'REGISTER_REWARD'
  return value.toUpperCase().replace(/\s+/g, '_')
}

const walletStatementDirectionCode = (value: (typeof walletStatementDirectionOptions)[number]) => {
  if (value === 'All Directions') return undefined
  return value === 'Credit' ? 'IN' : 'OUT'
}

const walletStatementTimeRange = (value: (typeof walletStatementTimeOptions)[number]) => {
  if (value === 'All Time') return {}

  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const end = new Date(start)
  end.setDate(end.getDate() + 1)
  end.setMilliseconds(end.getMilliseconds() - 1)

  if (value === 'Yesterday') {
    start.setDate(start.getDate() - 1)
    end.setDate(end.getDate() - 1)
  } else if (value.startsWith('Last ')) {
    const days = Number.parseInt(value.replace(/\D/g, ''), 10)
    if (Number.isFinite(days) && days > 1) start.setDate(start.getDate() - (days - 1))
  }

  return { startTime: start.toISOString(), endTime: end.toISOString() }
}

const toWalletRecordItem = (record: H5WalletRecordPayload): WalletRecordItem => ({
  id: record.id,
  title: record.title,
  source: record.source || '-',
  amount: `${record.sign || '+'} ${formatCurrency(record.amount)}`,
  balance: record.balance ? formatCurrency(record.balance) : '--',
  time: record.time || '--',
  orderNumber: record.orderNumber || '--',
  status: record.status,
  direction: record.direction,
  sign: record.sign || '+',
  gameName: record.gameName || '',
  providerCode: record.providerCode || '',
})

const toCashAuditItem = (record: H5CashAuditPayload): CashAuditItem => ({
  id: record.id,
  amount: formatCurrency(record.amount),
  createdAt: record.createdAt || '--',
  type: record.type || 'Cash Reward',
  auditRequirement: record.auditRequirement || '1.0x',
  progressAmount: formatCurrency(record.progressAmount),
  requiredAmount: formatCurrency(record.requiredAmount),
  progressPercent: Math.max(0, Math.min(100, Number(record.progressPercent || 0))),
})

const loadWalletSummary = async () => {
  try {
    await refreshMockUser()
  } catch (error) {
    console.warn('[api] failed to refresh wallet summary.', error)
  }
}

const loadWalletRecords = async (options: { append?: boolean } = {}) => {
  const append = options.append === true
  if (append && (walletRecordsLoading.value || walletRecordsLoadingMore.value || !walletRecordsHasMore.value)) return

  const targetPage = append ? walletRecordPage.value + 1 : 1
  const sequence = ++walletRecordLoadSequence
  if (append) {
    walletRecordsLoadingMore.value = true
  } else {
    walletRecordsLoading.value = true
    walletRecordsLoadingMore.value = false
  }

  try {
    const payload = await h5Api.getWalletRecords({
      tab: walletTabCode(activeWalletRecordTab.value),
      type:
        activeWalletRecordTab.value === 'Statement'
          ? walletStatementTypeCode(selectedWalletStatementType.value)
          : undefined,
      direction:
        activeWalletRecordTab.value === 'Statement'
          ? walletStatementDirectionCode(selectedWalletStatementDirection.value)
          : undefined,
      ...(activeWalletRecordTab.value === 'Statement'
        ? walletStatementTimeRange(selectedWalletStatementTime.value)
        : {}),
      page: targetPage,
      pageSize: 10,
    })
    if (sequence !== walletRecordLoadSequence) return

    const nextItems = (payload.items || []).map(toWalletRecordItem)
    if (append) {
      const recordsByID = new Map(walletRecordItems.value.map((record) => [record.id, record]))
      for (const record of nextItems) recordsByID.set(record.id, record)
      walletRecordItems.value = [...recordsByID.values()]
    } else {
      walletRecordItems.value = nextItems
    }
    walletRecordPage.value = payload.page || targetPage
    walletRecordTotal.value = payload.total || 0
  } catch (error) {
    if (sequence !== walletRecordLoadSequence) return
    console.warn('[api] failed to load wallet records.', error)
    if (!append) {
      walletRecordItems.value = []
      walletRecordPage.value = 1
      walletRecordTotal.value = 0
    }
  } finally {
    if (sequence === walletRecordLoadSequence) {
      walletRecordsLoading.value = false
      walletRecordsLoadingMore.value = false
      queueWalletRecordLoadCheck()
    }
  }
}

const loadMoreWalletRecords = () => {
  if (pageName.value !== 'wallet-records') return
  void loadWalletRecords({ append: true })
}

const checkWalletRecordScrollPosition = () => {
  if (pageName.value !== 'wallet-records') return
  const container = subpageContentRef.value
  if (!container) return
  const remaining = container.scrollHeight - (container.scrollTop + container.clientHeight)
  if (remaining < 320) loadMoreWalletRecords()
}

function queueWalletRecordLoadCheck() {
  window.cancelAnimationFrame(walletRecordLoadCheckFrame)
  void nextTick().then(() => {
    walletRecordLoadCheckFrame = window.requestAnimationFrame(checkWalletRecordScrollPosition)
  })
}

const handleWalletRecordScroll = () => {
  checkWalletRecordScrollPosition()
}

const loadCashAudits = async () => {
  if (cashAuditLoading.value) return
  cashAuditLoading.value = true
  try {
    const payload = await h5Api.getCashAudits()
    cashAuditItems.value = (payload.items || []).map(toCashAuditItem)
  } catch (error) {
    console.warn('[api] failed to load cash audit details.', error)
    cashAuditItems.value = []
  } finally {
    cashAuditLoading.value = false
  }
}

const loadBonusCenter = async () => {
  if (bonusCenterLoading.value) return
  bonusCenterLoading.value = true
  try {
    const payload = await h5Api.getBonusCenterRecords({
      category: bonusCategoryCodes[selectedBonusCategory.value],
      page: 1,
      pageSize: 50,
    })
    bonusCenterRecords.value = (payload.list || []).map(toBonusCenterItem)
  } catch (error) {
    console.warn('[api] failed to load bonus center records.', error)
    bonusCenterRecords.value = []
  } finally {
    bonusCenterLoading.value = false
  }
}

onMounted(() => {
  applySupportTabFromRoute()
  applyWalletRecordTabFromRoute()
  if (pageName.value === 'support') void loadSupport()
  if (pageName.value === 'wallet-deposit') void loadDepositConfig()
  if (pageName.value === 'wallet-records') void loadWalletRecords()
  if (pageName.value === 'cash-center') {
    void loadWalletSummary()
    void loadCashAudits()
  }
  if (pageName.value === 'bonus-center') {
    void loadWalletSummary()
    void loadBonusCenter()
  }
})

onBeforeUnmount(() => {
  walletRecordLoadSequence += 1
  window.cancelAnimationFrame(walletRecordLoadCheckFrame)
})

watch(pageName, (name) => {
  if (name === 'support') {
    applySupportTabFromRoute()
    void loadSupport()
  }
  if (name === 'wallet-records') {
    applyWalletRecordTabFromRoute()
    void loadWalletRecords()
  }
  if (name === 'wallet-deposit') void loadDepositConfig()
  if (name === 'cash-center') {
    void loadWalletSummary()
    void loadCashAudits()
  }
  if (name === 'bonus-center') {
    void loadWalletSummary()
    void loadBonusCenter()
  }
})

watch(
  () => route.query.tab,
  () => {
    applySupportTabFromRoute()
    applyWalletRecordTabFromRoute()
    if (pageName.value === 'support' && activeSupportTab.value === 'Notice') void loadSupport()
    if (pageName.value === 'wallet-records') void loadWalletRecords()
  },
)

watch(selectedBonusCategory, () => {
  if (pageName.value === 'bonus-center') void loadBonusCenter()
})

watch([selectedWalletStatementType, selectedWalletStatementDirection, selectedWalletStatementTime], () => {
  if (pageName.value === 'wallet-records' && activeWalletRecordTab.value === 'Statement') void loadWalletRecords()
})

watch(depositReturnExpanded, async (expanded) => {
  if (!expanded || pageName.value !== 'wallet-deposit') return
  await nextTick()
  scrollDepositPanelToBottom()
  window.setTimeout(scrollDepositPanelToBottom, 280)
})
</script>

<template>
  <div class="profile-subpage">
    <header class="subpage-header">
      <button type="button" aria-label="Back" @click="goBack">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M4.29 11.29a1 1 0 0 0 0 1.42l4 4a1 1 0 0 0 1.42-1.42L7.41 13H19a1 1 0 1 0 0-2H7.41l2.3-2.29a1 1 0 0 0-1.42-1.42l-4 4Z" />
        </svg>
      </button>
      <h1>{{ pageTitle }}</h1>
      <button
        v-if="showHeaderBell"
        type="button"
        class="subpage-header__notification"
        aria-label="Notifications"
        @click="openContact"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M19.4 17.21 18.2 14.43v-3.96c0-2.62-1.64-4.99-4-5.93C14.12 3.4 13.16 2.5 12 2.5s-2.08.9-2.2 2.04c-2.4.94-4 3.31-4 5.93v3.96L4.6 17.21c-.16.37-.12.78.08 1.1.2.33.56.53.96.53h12.72c.4 0 .76-.2.96-.53.2-.33.24-.78.08-1.1ZM12 21.5c1.12 0 2-.9 2-2.04h-4c0 1.14.88 2.04 2 2.04Z" />
        </svg>
      </button>
      <button
        v-else-if="showBetHelp"
        type="button"
        class="subpage-header__help"
        aria-label="Order Description"
        @click="openOrderDescription"
      >
        <img :src="betHelpImage" alt="help" loading="eager" decoding="sync" />
      </button>
      <span v-else class="subpage-header__spacer" />
    </header>

    <main
      ref="subpageContentRef"
      class="subpage-content"
      :class="[`subpage-content--${pageName}`, { 'is-support-collapsed': isSupportHeaderCollapsed }]"
      @scroll.passive="handleWalletRecordScroll"
    >
      <template v-if="pageName === 'wallet-records'">
        <nav class="segmented-tabs" aria-label="Wallet records">
          <button
            v-for="tab in walletRecordTabs"
            :key="tab"
            type="button"
            :class="{ 'is-active': activeWalletRecordTab === tab }"
            @click="selectWalletRecordTab(tab)"
          >
            {{ tab }}
          </button>
        </nav>
        <div v-if="activeWalletRecordTab === 'Withdraw'" class="wallet-filter-row wallet-filter-row--single">
          <button type="button" class="wallet-filter-button" @click="openPicker('walletWithdrawStatus')">
            <span>{{ selectedWalletWithdrawStatus }}</span>
            <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
              <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
        </div>
        <div v-else-if="activeWalletRecordTab === 'Statement'" class="wallet-filter-row wallet-filter-row--statement">
          <button type="button" class="wallet-filter-button" @click="openPicker('walletStatementTime')">
            <span>{{ selectedWalletStatementTime }}</span>
            <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
              <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
          <button type="button" class="wallet-filter-button" @click="openPicker('walletStatementType')">
            <span>{{ selectedWalletStatementType }}</span>
            <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
              <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
          <button type="button" class="wallet-filter-button" @click="openPicker('walletStatementDirection')">
            <span>{{ selectedWalletStatementDirection }}</span>
            <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
              <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
            </svg>
          </button>
        </div>
        <div v-if="walletRecordsLoading && !walletRecordItems.length" class="wallet-record-load-state">Loading...</div>
        <section v-else-if="walletRecordItems.length" class="wallet-source-record-list" aria-label="Wallet record list">
          <article v-for="record in walletRecordItems" :key="record.id" class="wallet-source-record-card">
            <div class="wallet-source-record-card__summary">
              <span>{{ record.title }}</span>
              <strong :class="{ 'is-debit': record.sign === '-' }">{{ record.amount }}</strong>
            </div>
            <div class="wallet-source-record-card__meta">
              <span>{{ record.time }}</span>
              <span>Balance: {{ record.balance }}</span>
            </div>
            <div v-if="record.gameName" class="wallet-source-record-card__source">
              <span>Game</span>
              <strong>{{ record.gameName }}</strong>
            </div>
            <div v-if="record.providerCode" class="wallet-source-record-card__source">
              <span>Provider</span>
              <strong>{{ record.providerCode }}</strong>
            </div>
            <div v-if="!record.gameName" class="wallet-source-record-card__source">
              <span>Source</span>
              <strong>{{ record.source }}</strong>
            </div>
          </article>
          <div v-if="walletRecordsLoadingMore" class="wallet-record-load-state">Loading...</div>
        </section>
        <section v-else class="source-empty-state" aria-live="polite">
          <div class="source-empty-state__top" aria-hidden="true" />
          <div class="source-empty-state__body">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" width="96" height="96" viewBox="0 0 96 96" aria-hidden="true">
              <g>
                <g style="opacity: 0.1;">
                  <ellipse cx="48" cy="93.12" rx="48" ry="2.88" fill="currentColor" />
                </g>
                <path
                  fill-rule="evenodd"
                  clip-rule="evenodd"
                  d="M74.55 15.93a2.53 2.53 0 0 1 5.06.08l-.08 5.2a2.53 2.53 0 1 1-5.06-.08l.08-5.2Zm2.54-.57a.61.61 0 0 0-.62.6l-.08 5.2a.61.61 0 0 0 1.22.02l.08-5.2a.61.61 0 0 0-.6-.62Zm2.8 11.08a2.53 2.53 0 0 1 2.57-2.49l5.2.08a2.53 2.53 0 1 1-.08 5.06l-5.2-.08a2.53 2.53 0 0 1-2.49-2.57Zm2.54-.57a.61.61 0 0 0-.62.6.61.61 0 0 0 .6.62l5.2.08a.61.61 0 0 0 .02-1.22l-5.2-.08ZM21.31 30.49a6.72 6.72 0 0 1 6.51-5.05h43.24c3.07 0 5.75 2.08 6.51 5.05l4.67 18.2c.22.88.27 1.79.13 2.69l-2.71 17.62a6.72 6.72 0 0 1-6.64 5.7H32.6a13.44 13.44 0 1 1-15.36-18.22l-.75-5.13a6.72 6.72 0 0 1 .14-2.64l4.68-18.22Zm-2.16 25.71a13.68 13.68 0 0 1 1.01-.04A13.44 13.44 0 0 1 33.6 69.6c0 1.09-.13 2.15-.38 3.17h39.8a4.8 4.8 0 0 0 4.74-4.07l2.72-17.67H66.17l-4.06 5.68a6.72 6.72 0 0 1-5.47 2.81H42.85a6.72 6.72 0 0 1-5.47-2.81l-4.06-5.68H18.39l.76 5.17Zm56.56-25.23 4.65 18.14h-6.3v-12.8a4.8 4.8 0 0 0-4.8-4.8H29.61a4.8 4.8 0 0 0-4.8 4.8v12.8h-6.29l4.65-18.14a4.8 4.8 0 0 1 4.65-3.61h43.24a4.8 4.8 0 0 1 4.65 3.61ZM72.14 38.54a4.78 4.78 0 0 0-2.88-.96H29.61c-1.08 0-2.08.36-2.88.96v-2.23a2.88 2.88 0 0 1 2.88-2.88h39.65a2.88 2.88 0 0 1 2.88 2.88v2.23Zm-45.4 10.57h7.08c.31 0 .6.15.78.4l4.34 6.08a4.8 4.8 0 0 0 3.91 2.01h13.79a4.8 4.8 0 0 0 3.9-2.01l4.35-6.08a.96.96 0 0 1 .78-.4h6.47v-.55l-.19-.98a2.87 2.87 0 0 0-2.69-2.01H29.61a2.88 2.88 0 0 0-2.82 2.4l-.05.58v.56ZM8.64 69.6a11.52 11.52 0 1 1 23.04 0 11.52 11.52 0 0 1-23.04 0Zm9.12-6.24a2.4 2.4 0 0 1 4.8 0v5.76a2.4 2.4 0 1 1-4.8 0v-5.76Zm2.4-.48a.48.48 0 0 0-.48.48v5.76a.48.48 0 1 0 .96 0v-5.76a.48.48 0 0 0-.48-.48Zm0 15.36a2.4 2.4 0 1 1 0-4.8 2.4 2.4 0 0 1 0 4.8Zm-.48-2.4a.48.48 0 1 0 .96 0 .48.48 0 0 0-.96 0Z"
                  fill="currentColor"
                />
              </g>
            </svg>
          </div>
          <p>Oops! Nothing here for now.</p>
        </section>
        <div
          v-if="!walletRecordsLoading && !walletRecordsLoadingMore && !walletRecordsHasMore"
          class="py-4 pb-safe text-center text-xs text-gray-400"
        >
          No more results
        </div>
      </template>

      <template v-else-if="pageName === 'wallet-deposit'">
        <section ref="depositPanelRef" class="deposit-panel" aria-label="Deposit form">
          <div class="deposit-title-row">
            <img class="deposit-title-row__icon" :src="depositPaymentMethodIcon" alt="" aria-hidden="true" />
            <h2>Payment Method</h2>
            <button type="button" class="deposit-refresh" aria-label="Refresh payment methods">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M20.24 12.24a1 1 0 0 0-1.22.72 7 7 0 0 1-11.4 3.45h1.64a1 1 0 1 0 0-2H5.1a1 1 0 0 0-1 1v4.16a1 1 0 1 0 2 0v-1.7a9 9 0 0 0 14.86-4.41 1 1 0 0 0-.72-1.22ZM4.98 11.04a7 7 0 0 1 11.4-3.45h-1.64a1 1 0 1 0 0 2h4.16a1 1 0 0 0 1-1V4.43a1 1 0 1 0-2 0v1.7A9 9 0 0 0 3.04 10.54a1 1 0 1 0 1.94.5Z" />
              </svg>
            </button>
          </div>

          <div class="deposit-method-grid">
            <button
              v-for="method in depositMethods"
              :key="method.methodCode"
              type="button"
              :class="{ 'is-active': selectedDepositMethodCode === method.methodCode }"
              @click="selectDepositMethod(method.methodCode)"
            >
              <img class="deposit-logo" :src="method.displayIcon" alt="" loading="lazy" aria-hidden="true" />
              <span>{{ method.name }}</span>
            </button>
          </div>

          <div class="channel-list">
            <button
              v-for="channel in depositChannels"
              :key="channel.channelCode"
              type="button"
              :class="{ 'is-active': selectedDepositChannelCode === channel.channelCode }"
              @click="selectDepositChannel(channel.channelCode)"
            >
              <img class="deposit-logo" :src="channel.displayIcon" alt="" loading="lazy" aria-hidden="true" />
              <span>{{ channel.name }}</span>
            </button>
          </div>

          <div class="amount-grid">
            <button
              v-for="option in amountOptions"
              :key="option.optionCode"
              type="button"
              :class="{ 'is-active': isDepositAmountActive(option) }"
              @click="selectDepositAmount(option)"
            >
              <span>{{ formatDepositMoney(option.amountValue) }}</span>
              <em v-if="getDepositAmountOptionReward(option) > 0">{{ formatDepositNumber(getDepositAmountOptionReward(option)) }}</em>
            </button>
          </div>

          <label class="deposit-amount-input">
            <span>{{ currencySymbol() }}</span>
            <input v-model="depositAmount" inputmode="decimal" aria-label="Amount" />
            <button type="button" aria-label="Clear amount" @click="depositAmount = ''">×</button>
          </label>

          <div class="deposit-label-row">
            <span><img class="deposit-label-row__icon" :src="depositVoucherImage" alt="" aria-hidden="true" /> Deposit Voucher</span>
            <button type="button" @click="openDepositVoucherSheet">More <i aria-hidden="true" /></button>
          </div>

          <div ref="depositVoucherStripRef" class="voucher-strip" aria-label="Deposit vouchers">
            <button
              v-for="(voucher, index) in depositVoucherOptions"
              :key="voucher.id"
              type="button"
              class="voucher-ticket"
              :class="{
                'is-active': selectedDepositVoucherIndex === index,
                'is-muted': selectedDepositVoucherIndex !== index,
                'is-disabled': !voucher.available,
              }"
              @click="selectDepositVoucher(index)"
            >
              <svg class="voucher-ticket__shape" viewBox="0 0 340 72" preserveAspectRatio="none" fill="none" aria-hidden="true">
                <defs>
                  <linearGradient :id="`voucher-fill-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                    <stop stop-color="#F4E3B5" />
                    <stop offset="1" stop-color="#F2C364" />
                  </linearGradient>
                  <linearGradient :id="`voucher-muted-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                    <stop stop-color="#5B4F38" />
                    <stop offset="1" stop-color="#7A5A22" />
                  </linearGradient>
                </defs>
                <path
                  :fill="selectedDepositVoucherIndex === index ? `url(#voucher-fill-${voucher.id})` : `url(#voucher-muted-${voucher.id})`"
                  d="M67.6748 0C68.7984 2.3647 71.2078 4 74 4C76.7922 4 79.2016 2.3647 80.3252 0H331C332.742 0 334.352 0.558687 335.666 1.50391C334.043 2.78584 333 4.77086 333 7C333 10.5264 335.608 13.4414 339 13.9268V19.0723C335.608 19.5576 333 22.4736 333 26C333 29.5264 335.608 32.4414 339 32.9268V38.0723C335.608 38.5576 333 41.4736 333 45C333 48.5264 335.608 51.4414 339 51.9268V57.0723C335.608 57.5576 333 60.4736 333 64C333 66.5208 334.333 68.7294 336.332 69.9619C334.917 71.2286 333.049 72 331 72H80.3252C79.2016 69.6353 76.7922 68 74 68C71.2078 68 68.7984 69.6353 67.6748 72H8C3.58172 72 0 68.4183 0 64V8C0 3.58172 3.58172 0 8 0H67.6748Z"
                />
                <path class="voucher-ticket__dash" d="M74 6L74 66" />
              </svg>
              <span class="voucher-ticket__check" aria-hidden="true">
                <svg viewBox="0 0 22 22" fill="none">
                  <path d="M10.8 18.8C13.009 18.8 15.009 17.905 16.457 16.458C17.904 15.01 18.8 13.01 18.8 10.8C18.8 8.592 17.904 6.592 16.457 5.144C15.009 3.696 13.009 2.8 10.8 2.8C8.591 2.8 6.591 3.696 5.143 5.144C3.695 6.592 2.8 8.592 2.8 10.8C2.8 13.01 3.695 15.01 5.143 16.458C6.591 17.905 8.591 18.8 10.8 18.8Z" />
                  <path v-if="voucher.available" d="M7.55 10.8L9.8 13.05L14.3 8.55" />
                </svg>
              </span>
              <img :src="voucher.displayIcon" alt="" loading="lazy" />
              <strong>{{ voucher.title }}</strong>
              <span>{{ formatDepositNumber(voucher.minAmountValue) }}–{{ formatDepositNumber(voucher.maxAmountValue) }}</span>
              <small>Available Amount Range</small>
              <em><span>Max</span> {{ formatDepositNumber(voucher.maxRewardValue) }}</em>
            </button>
          </div>
        </section>

        <div class="deposit-sticky-action">
          <div class="deposit-return-wrap" :class="{ 'is-expanded': depositReturnExpanded }">
            <button type="button" class="deposit-return-row" @click="depositReturnExpanded = !depositReturnExpanded">
              <span>
                You Will Get
                <svg width="14" height="14" viewBox="0 0 18 18" fill="none" :class="{ 'is-expanded': depositReturnExpanded }" aria-hidden="true">
                  <path d="M9 13.5L3.375 6.75H14.625L9 13.5Z" fill="#FFFFFF" stroke="#FFFFFF" stroke-linejoin="round" />
                </svg>
              </span>
              <strong>{{ formatDepositMoney(depositTotalGet) }}</strong>
            </button>
            <div class="deposit-return-details">
              <div class="deposit-return-details__inner">
                <div v-for="item in depositReturnDetails" :key="item.label" class="deposit-return-detail">
                  <span class="deposit-return-detail__dot" :class="item.dot" aria-hidden="true" />
                  <span>{{ item.label }}</span>
                  <strong>{{ formatDepositMoney(item.value) }}</strong>
                </div>
              </div>
            </div>
          </div>
          <button
            type="button"
            class="primary-action"
            :disabled="depositSubmitting || depositAmountValue <= 0"
            @click="submitDeposit"
          >
            Deposit
          </button>
        </div>

        <Transition name="deposit-progress">
          <div
            v-if="showDepositProgressDialog"
            class="deposit-progress-overlay"
            role="dialog"
            aria-modal="true"
            aria-label="Deposit in progress"
          >
            <article class="deposit-progress-dialog">
              <div class="deposit-progress-dialog__scroll">
                <div class="deposit-progress-dialog__content">
                  <div class="deposit-progress-animation" aria-hidden="true">
                    <LottiePlayer :animation-data="depositTipAnimation" :loop="true" :autoplay="showDepositProgressDialog" />
                  </div>
                  <div class="deposit-progress-title">Deposit in progress</div>
                  <div class="deposit-progress-order">Order No.: {{ depositProgressOrderNo || '-' }}</div>
                  <button type="button" class="deposit-progress-primary" @click="viewDepositOrder">View Order</button>
                  <button type="button" class="deposit-progress-secondary" @click="closeDepositProgressDialog">Deposit Again</button>
                </div>
              </div>
            </article>
          </div>
        </Transition>

        <div
          v-if="showDepositVoucherSheet"
          class="deposit-voucher-overlay"
          role="presentation"
          @click.self="closeDepositVoucherSheet"
        >
          <div class="deposit-voucher-sheet" role="dialog" aria-modal="true" aria-label="Deposit Voucher">
            <div class="deposit-voucher-sheet__handle" aria-hidden="true"><span /></div>
            <div class="deposit-voucher-sheet__header">
              <span>Deposit Voucher</span>
            </div>
            <div class="deposit-voucher-sheet__list">
              <button
                v-for="(voucher, index) in depositVoucherOptions"
                :key="`sheet-${voucher.id}`"
                type="button"
                class="voucher-ticket voucher-ticket--sheet"
                :class="{
                  'is-active': selectedDepositVoucherIndex === index,
                  'is-muted': selectedDepositVoucherIndex !== index,
                  'is-disabled': !voucher.available,
                }"
                @click="selectDepositVoucherFromSheet(index)"
              >
                <svg class="voucher-ticket__shape" viewBox="0 0 340 72" preserveAspectRatio="none" fill="none" aria-hidden="true">
                  <defs>
                    <linearGradient :id="`sheet-voucher-fill-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                      <stop stop-color="#F4E3B5" />
                      <stop offset="1" stop-color="#F2C364" />
                    </linearGradient>
                    <linearGradient :id="`sheet-voucher-muted-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                      <stop stop-color="#5B4F38" />
                      <stop offset="1" stop-color="#7A5A22" />
                    </linearGradient>
                  </defs>
                  <path
                    :fill="selectedDepositVoucherIndex === index ? `url(#sheet-voucher-fill-${voucher.id})` : `url(#sheet-voucher-muted-${voucher.id})`"
                    d="M67.6748 0C68.7984 2.3647 71.2078 4 74 4C76.7922 4 79.2016 2.3647 80.3252 0H331C332.742 0 334.352 0.558687 335.666 1.50391C334.043 2.78584 333 4.77086 333 7C333 10.5264 335.608 13.4414 339 13.9268V19.0723C335.608 19.5576 333 22.4736 333 26C333 29.5264 335.608 32.4414 339 32.9268V38.0723C335.608 38.5576 333 41.4736 333 45C333 48.5264 335.608 51.4414 339 51.9268V57.0723C335.608 57.5576 333 60.4736 333 64C333 66.5208 334.333 68.7294 336.332 69.9619C334.917 71.2286 333.049 72 331 72H80.3252C79.2016 69.6353 76.7922 68 74 68C71.2078 68 68.7984 69.6353 67.6748 72H8C3.58172 72 0 68.4183 0 64V8C0 3.58172 3.58172 0 8 0H67.6748Z"
                  />
                  <path class="voucher-ticket__dash" d="M74 6L74 66" />
                </svg>
                <span class="voucher-ticket__check" aria-hidden="true">
                  <svg viewBox="0 0 22 22" fill="none">
                    <path d="M10.8 18.8C13.009 18.8 15.009 17.905 16.457 16.458C17.904 15.01 18.8 13.01 18.8 10.8C18.8 8.592 17.904 6.592 16.457 5.144C15.009 3.696 13.009 2.8 10.8 2.8C8.591 2.8 6.591 3.696 5.143 5.144C3.695 6.592 2.8 8.592 2.8 10.8C2.8 13.01 3.695 15.01 5.143 16.458C6.591 17.905 8.591 18.8 10.8 18.8Z" />
                    <path v-if="voucher.available" d="M7.55 10.8L9.8 13.05L14.3 8.55" />
                  </svg>
                </span>
                <img :src="voucher.displayIcon" alt="" loading="lazy" />
                <strong>{{ voucher.title }}</strong>
                <span>{{ formatDepositNumber(voucher.minAmountValue) }}–{{ formatDepositNumber(voucher.maxAmountValue) }}</span>
                <small>Available Amount Range</small>
                <em><span>Max</span> {{ formatDepositNumber(voucher.maxRewardValue) }}</em>
              </button>
            </div>
          </div>
        </div>
      </template>

      <template v-else-if="pageName === 'cash-center'">
        <section class="cash-wallet-panel">
          <div class="cash-wallet-card">
            <img class="cash-wallet-card__bg" :src="cashWalletBgImage" alt="" loading="eager" decoding="async" />
            <div class="cash-wallet-card__content">
              <div class="cash-wallet-balances">
                <div class="cash-wallet-balance">
                  <div class="cash-wallet-balance__label">
                    <img :src="cashWalletIcon" alt="" />
                    <span>Cash Wallet</span>
                  </div>
                  <strong>
                  {{ currencySymbol() }}
                    {{ cashWalletBalance }}
                  </strong>
                </div>
                <div class="cash-wallet-divider" aria-hidden="true" />
                <div class="cash-wallet-balance">
                  <div class="cash-wallet-balance__label">
                    <img class="is-tilted" :src="withdrawableIcon" alt="" />
                    <span>Withdrawable</span>
                  </div>
                  <strong class="is-success">
                    {{ currencySymbol() }}
                    {{ withdrawableWalletBalance }}
                  </strong>
                </div>
              </div>
              <button type="button" class="cash-wallet-withdraw" disabled>Withdraw</button>
            </div>
          </div>
          <div class="cash-wallet-note">
            <span v-if="cashAuditItems.length>0">No bonus available to transfer.</span>
            <span v-else>Complete all Cash Audit requirements to withdraw.</span>
          </div>
        </section>
        <nav class="cash-wallet-tabs" aria-label="Cash wallet details">
          <div class="cash-wallet-tabs__inner">
            <button
              type="button"
              :class="{ 'is-active': activeCashTab === 'audit' }"
              @click="activeCashTab = 'audit'"
            >
              <span>Audit Details</span>
            </button>
            <button
              type="button"
              :class="{ 'is-active': activeCashTab === 'records' }"
              @click="activeCashTab = 'records'"
            >
              <span>Withdraw Records</span>
            </button>
            <span
              class="cash-wallet-tabs__indicator"
              :style="{ left: activeCashTab === 'audit' ? '25%' : '75%' }"
              aria-hidden="true"
            />
          </div>
        </nav>
        <section v-if="activeCashTab === 'audit' && cashAuditItems.length" class="cash-audit-list" aria-label="Audit details">
          <article v-for="item in cashAuditItems" :key="item.id" class="cash-audit-card">
            <header class="cash-audit-card__header">
              <strong>{{ item.amount }}</strong>
              <button type="button">
                <span>{{ item.createdAt }}</span>
                <svg viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                  <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </button>
            </header>
            <div class="cash-audit-card__body">
              <div>
                <span>Type</span>
                <strong>{{ item.type }}</strong>
              </div>
              <div>
                <span>Audit Requirement</span>
                <strong>{{ item.auditRequirement }}</strong>
              </div>
              <div>
                <span>Audit Progress</span>
                <strong class="cash-audit-card__progress-text">
                  <span>{{ item.progressAmount }}</span>
                  <em>/</em>
                  <span>{{ item.requiredAmount }}</span>
                </strong>
              </div>
            </div>
            <div class="cash-audit-card__bar-wrap">
              <div class="cash-audit-card__bar">
                <span class="cash-audit-card__bar-fill" :style="{ width: `${item.progressPercent}%` }" />
                <span class="cash-audit-card__bar-shine" :style="{ width: `${item.progressPercent}%` }" />
              </div>
              <img :src="cashAuditBundleImage" alt="" loading="eager" decoding="sync" class="mb-1" />
            </div>
          </article>
        </section>
        <section v-else class="source-empty-state cash-wallet-empty" aria-live="polite">
          <div class="source-empty-state__top" aria-hidden="true" />
          <div class="source-empty-state__body">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" width="96" height="96" viewBox="0 0 96 96" aria-hidden="true">
              <g>
                <g style="opacity: 0.1;">
                  <ellipse cx="48" cy="93.12" rx="48" ry="2.88" fill="currentColor" />
                </g>
                <path
                  fill-rule="evenodd"
                  clip-rule="evenodd"
                  d="M74.55 15.93a2.53 2.53 0 0 1 5.06.08l-.08 5.2a2.53 2.53 0 1 1-5.06-.08l.08-5.2Zm2.54-.57a.61.61 0 0 0-.62.6l-.08 5.2a.61.61 0 0 0 1.22.02l.08-5.2a.61.61 0 0 0-.6-.62Zm2.8 11.08a2.53 2.53 0 0 1 2.57-2.49l5.2.08a2.53 2.53 0 1 1-.08 5.06l-5.2-.08a2.53 2.53 0 0 1-2.49-2.57Zm2.54-.57a.61.61 0 0 0-.62.6.61.61 0 0 0 .6.62l5.2.08a.61.61 0 0 0 .02-1.22l-5.2-.08ZM21.31 30.49a6.72 6.72 0 0 1 6.51-5.05h43.24c3.07 0 5.75 2.08 6.51 5.05l4.67 18.2c.22.88.27 1.79.13 2.69l-2.71 17.62a6.72 6.72 0 0 1-6.64 5.7H32.6a13.44 13.44 0 1 1-15.36-18.22l-.75-5.13a6.72 6.72 0 0 1 .14-2.64l4.68-18.22Zm-2.16 25.71a13.68 13.68 0 0 1 1.01-.04A13.44 13.44 0 0 1 33.6 69.6c0 1.09-.13 2.15-.38 3.17h39.8a4.8 4.8 0 0 0 4.74-4.07l2.72-17.67H66.17l-4.06 5.68a6.72 6.72 0 0 1-5.47 2.81H42.85a6.72 6.72 0 0 1-5.47-2.81l-4.06-5.68H18.39l.76 5.17Zm56.56-25.23 4.65 18.14h-6.3v-12.8a4.8 4.8 0 0 0-4.8-4.8H29.61a4.8 4.8 0 0 0-4.8 4.8v12.8h-6.29l4.65-18.14a4.8 4.8 0 0 1 4.65-3.61h43.24a4.8 4.8 0 0 1 4.65 3.61ZM72.14 38.54a4.78 4.78 0 0 0-2.88-.96H29.61c-1.08 0-2.08.36-2.88.96v-2.23a2.88 2.88 0 0 1 2.88-2.88h39.65a2.88 2.88 0 0 1 2.88 2.88v2.23Zm-45.4 10.57h7.08c.31 0 .6.15.78.4l4.34 6.08a4.8 4.8 0 0 0 3.91 2.01h13.79a4.8 4.8 0 0 0 3.9-2.01l4.35-6.08a.96.96 0 0 1 .78-.4h6.47v-.55l-.19-.98a2.87 2.87 0 0 0-2.69-2.01H29.61a2.88 2.88 0 0 0-2.82 2.4l-.05.58v.56ZM8.64 69.6a11.52 11.52 0 1 1 23.04 0 11.52 11.52 0 0 1-23.04 0Zm9.12-6.24a2.4 2.4 0 0 1 4.8 0v5.76a2.4 2.4 0 1 1-4.8 0v-5.76Zm2.4-.48a.48.48 0 0 0-.48.48v5.76a.48.48 0 1 0 .96 0v-5.76a.48.48 0 0 0-.48-.48Zm0 15.36a2.4 2.4 0 1 1 0-4.8 2.4 2.4 0 0 1 0 4.8Zm-.48-2.4a.48.48 0 1 0 .96 0 .48.48 0 0 0-.96 0Z"
                  fill="currentColor"
                />
              </g>
            </svg>
          </div>
          <p>Oops! Nothing here for now.</p>
        </section>
      </template>

      <template v-else-if="pageName === 'bonus-center'">
        <button type="button" class="filter-button bonus-filter-button" @click="openPicker('bonusCategory')">
          <span>{{ selectedBonusCategory }}</span>
          <svg class="bonus-filter-button__icon" viewBox="0 0 12 12" fill="none" aria-hidden="true">
            <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
          </svg>
        </button>
        <section v-if="bonusCenterRecords.length" class="record-list" aria-label="Bonus list">
          <article
            v-for="record in bonusCenterRecords"
            :key="record.id"
            class="record-card"
            :data-amount="record.amount"
          >
            <header>
              <strong>{{ record.amount }}</strong>
              <span>{{ record.source }}</span>
            </header>
            <dl>
              <div>
                <dt>Source</dt>
                <dd>{{ record.source }}</dd>
              </div>
              <div>
                <dt>Issued Time</dt>
                <dd>{{ record.issuedTime }}</dd>
              </div>
              <div>
                <dt>Order ID</dt>
                <dd>{{ record.orderId }}</dd>
              </div>
              <div>
                <dt>Audit Requirement</dt>
                <dd>{{ record.requirement }}</dd>
              </div>
            </dl>
          </article>
        </section>
        <section v-else class="source-empty-state" aria-live="polite">
          <div class="source-empty-state__top" aria-hidden="true" />
          <div class="source-empty-state__body">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" width="96" height="96" viewBox="0 0 96 96" aria-hidden="true">
              <g>
                <g style="opacity: 0.1;">
                  <ellipse cx="48" cy="93.12" rx="48" ry="2.88" fill="currentColor" />
                </g>
                <path
                  fill-rule="evenodd"
                  clip-rule="evenodd"
                  d="M74.55 15.93a2.53 2.53 0 0 1 5.06.08l-.08 5.2a2.53 2.53 0 1 1-5.06-.08l.08-5.2Zm2.54-.57a.61.61 0 0 0-.62.6l-.08 5.2a.61.61 0 0 0 1.22.02l.08-5.2a.61.61 0 0 0-.6-.62Zm2.8 11.08a2.53 2.53 0 0 1 2.57-2.49l5.2.08a2.53 2.53 0 1 1-.08 5.06l-5.2-.08a2.53 2.53 0 0 1-2.49-2.57Zm2.54-.57a.61.61 0 0 0-.62.6.61.61 0 0 0 .6.62l5.2.08a.61.61 0 0 0 .02-1.22l-5.2-.08ZM21.31 30.49a6.72 6.72 0 0 1 6.51-5.05h43.24c3.07 0 5.75 2.08 6.51 5.05l4.67 18.2c.22.88.27 1.79.13 2.69l-2.71 17.62a6.72 6.72 0 0 1-6.64 5.7H32.6a13.44 13.44 0 1 1-15.36-18.22l-.75-5.13a6.72 6.72 0 0 1 .14-2.64l4.68-18.22Zm-2.16 25.71a13.68 13.68 0 0 1 1.01-.04A13.44 13.44 0 0 1 33.6 69.6c0 1.09-.13 2.15-.38 3.17h39.8a4.8 4.8 0 0 0 4.74-4.07l2.72-17.67H66.17l-4.06 5.68a6.72 6.72 0 0 1-5.47 2.81H42.85a6.72 6.72 0 0 1-5.47-2.81l-4.06-5.68H18.39l.76 5.17Zm56.56-25.23 4.65 18.14h-6.3v-12.8a4.8 4.8 0 0 0-4.8-4.8H29.61a4.8 4.8 0 0 0-4.8 4.8v12.8h-6.29l4.65-18.14a4.8 4.8 0 0 1 4.65-3.61h43.24a4.8 4.8 0 0 1 4.65 3.61Z"
                  fill="currentColor"
                />
              </g>
            </svg>
          </div>
          <p>Oops! Nothing here for now.</p>
        </section>
      </template>

      <template v-else-if="pageName === 'bet-records'">
        <nav class="segmented-tabs" aria-label="Bet records">
          <button
            v-for="tab in betRecordTabs"
            :key="tab"
            type="button"
            :class="{ 'is-active': activeBetRecordTab === tab }"
            @click="selectBetRecordTab(tab)"
          >
            {{ tab }}
          </button>
        </nav>
        <div class="filter-row" :class="{ 'filter-row--report': activeBetRecordTab === 'Report' }">
          <template v-if="activeBetRecordTab === 'Records'">
            <button type="button" class="bet-filter-button" @click="openPicker('betTime')">
              {{ selectedBetTime }}
              <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
                <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
            <button type="button" class="bet-filter-button" @click="openPicker('betStatus')">
              {{ selectedBetStatus }}
              <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
                <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
            <button type="button" class="bet-filter-button" @click="openPicker('betType')">
              {{ selectedBetType }}
              <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
                <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
          </template>
          <template v-else>
            <button type="button" class="bet-filter-button" @click="openPicker('reportDate')">
              {{ selectedReportDate }}
              <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
                <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
            <button type="button" class="bet-filter-button bet-filter-button--wide" @click="openPicker('reportGameType')">
              {{ selectedReportGameType === 'All' ? 'All Games' : selectedReportGameType }}
              <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
                <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
          </template>
        </div>
        <section class="source-empty-state" aria-live="polite">
          <div class="source-empty-state__top" aria-hidden="true" />
          <div class="source-empty-state__body">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" width="96" height="96" viewBox="0 0 96 96" aria-hidden="true">
              <g>
                <g style="opacity: 0.1;">
                  <ellipse cx="48" cy="93.12" rx="48" ry="2.88" fill="currentColor" />
                </g>
                <path
                  fill-rule="evenodd"
                  clip-rule="evenodd"
                  d="M74.55 15.93a2.53 2.53 0 0 1 5.06.08l-.08 5.2a2.53 2.53 0 1 1-5.06-.08l.08-5.2Zm2.54-.57a.61.61 0 0 0-.62.6l-.08 5.2a.61.61 0 0 0 1.22.02l.08-5.2a.61.61 0 0 0-.6-.62Zm2.8 11.08a2.53 2.53 0 0 1 2.57-2.49l5.2.08a2.53 2.53 0 1 1-.08 5.06l-5.2-.08a2.53 2.53 0 0 1-2.49-2.57Zm2.54-.57a.61.61 0 0 0-.62.6.61.61 0 0 0 .6.62l5.2.08a.61.61 0 0 0 .02-1.22l-5.2-.08ZM21.31 30.49a6.72 6.72 0 0 1 6.51-5.05h43.24c3.07 0 5.75 2.08 6.51 5.05l4.67 18.2c.22.88.27 1.79.13 2.69l-2.71 17.62a6.72 6.72 0 0 1-6.64 5.7H32.6a13.44 13.44 0 1 1-15.36-18.22l-.75-5.13a6.72 6.72 0 0 1 .14-2.64l4.68-18.22Zm-2.16 25.71a13.68 13.68 0 0 1 1.01-.04A13.44 13.44 0 0 1 33.6 69.6c0 1.09-.13 2.15-.38 3.17h39.8a4.8 4.8 0 0 0 4.74-4.07l2.72-17.67H66.17l-4.06 5.68a6.72 6.72 0 0 1-5.47 2.81H42.85a6.72 6.72 0 0 1-5.47-2.81l-4.06-5.68H18.39l.76 5.17Zm56.56-25.23 4.65 18.14h-6.3v-12.8a4.8 4.8 0 0 0-4.8-4.8H29.61a4.8 4.8 0 0 0-4.8 4.8v12.8h-6.29l4.65-18.14a4.8 4.8 0 0 1 4.65-3.61h43.24a4.8 4.8 0 0 1 4.65 3.61ZM72.14 38.54a4.78 4.78 0 0 0-2.88-.96H29.61c-1.08 0-2.08.36-2.88.96v-2.23a2.88 2.88 0 0 1 2.88-2.88h39.65a2.88 2.88 0 0 1 2.88 2.88v2.23Zm-45.4 10.57h7.08c.31 0 .6.15.78.4l4.34 6.08a4.8 4.8 0 0 0 3.91 2.01h13.79a4.8 4.8 0 0 0 3.9-2.01l4.35-6.08a.96.96 0 0 1 .78-.4h6.47v-.55l-.19-.98a2.87 2.87 0 0 0-2.69-2.01H29.61a2.88 2.88 0 0 0-2.82 2.4l-.05.58v.56ZM8.64 69.6a11.52 11.52 0 1 1 23.04 0 11.52 11.52 0 0 1-23.04 0Zm9.12-6.24a2.4 2.4 0 0 1 4.8 0v5.76a2.4 2.4 0 1 1-4.8 0v-5.76Zm2.4-.48a.48.48 0 0 0-.48.48v5.76a.48.48 0 1 0 .96 0v-5.76a.48.48 0 0 0-.48-.48Zm0 15.36a2.4 2.4 0 1 1 0-4.8 2.4 2.4 0 0 1 0 4.8Zm-.48-2.4a.48.48 0 1 0 .96 0 .48.48 0 0 0-.96 0Z"
                  fill="currentColor"
                />
              </g>
            </svg>
          </div>
          <p>Oops! Nothing here for now.</p>
        </section>
      </template>

      <template v-else-if="pageName === 'security'">
        <section class="security-source-section">
          <div class="security-source-title">Account Information</div>
          <div class="security-source-card">
            <div class="security-source-row">
              <svg class="security-source-row__icon" viewBox="0 0 18 18" fill="none" aria-hidden="true">
                <rect x="2.25" y="3.25" width="13.5" height="11.5" rx="1.7" stroke="currentColor" stroke-width="1.35" />
                <path d="M4.25 7h3.8M4.25 10h6.4M12.75 11.25v-1.1a1.6 1.6 0 0 0-3.2 0v1.1M9.05 12h4.2" stroke="currentColor" stroke-width="1.35" stroke-linecap="round" />
              </svg>
              <span>Register Method</span>
              <strong>Phone Number</strong>
            </div>
            <div class="security-source-row">
              <svg class="security-source-row__icon" viewBox="0 0 18 18" fill="none" aria-hidden="true">
                <path d="M7.15 9.15a3.05 3.05 0 1 0 0-6.1 3.05 3.05 0 0 0 0 6.1ZM2.25 15.25c.36-2.75 2.38-4.7 4.9-4.7s4.55 1.95 4.9 4.7M12.8 5.05h3M12.1 8h3.9M12.8 10.95h3" stroke="currentColor" stroke-width="1.35" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
              <span>Register Account</span>
              <strong>{{ runtimePhoneCode }} {{ securityPhoneNumber }}</strong>
            </div>
          </div>
        </section>
        <section class="security-source-section">
          <div class="security-source-title">Link Account</div>
          <div class="security-source-card">
            <button type="button" class="security-source-row" @click="openSecurityPhone">
              <svg class="security-source-row__icon" viewBox="0 0 18 18" fill="none" aria-hidden="true">
                <path d="M6.38 2.15H4.46c-1.2 0-2.08 1.14-1.7 2.27 1.55 4.6 3.84 6.89 8.43 8.44 1.13.38 2.27-.5 2.27-1.7V9.24l-2.66-1.48-1.67.83a5.44 5.44 0 0 1-2.1-2.1l.83-1.67-1.48-2.67Z" stroke="currentColor" stroke-width="1.35" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
              <span>Phone</span>
              <strong>{{ runtimePhoneCode }} {{ securityPhoneNumber }}</strong>
              <i class="security-source-row__dot" aria-hidden="true" />
              <svg class="security-source-row__chevron" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </button>
            <button type="button" class="security-source-row">
              <svg class="security-source-row__icon" viewBox="0 0 18 18" fill="none" aria-hidden="true">
                <circle cx="9" cy="9" r="7.2" stroke="currentColor" stroke-width="1.35" />
                <path d="M12.75 7.6H9.1v2.35h2.25c-.34 1-1.16 1.55-2.35 1.55A2.52 2.52 0 0 1 6.5 9 2.52 2.52 0 0 1 9 6.5c.68 0 1.22.22 1.67.62l1.55-1.55A4.73 4.73 0 0 0 9 4.35 4.66 4.66 0 1 0 9 13.65c2.4 0 4-1.65 4-4 0-.65-.1-1.3-.25-2.05Z" fill="currentColor" />
              </svg>
              <span>Google</span>
              <strong>Unlinked</strong>
            </button>
          </div>
          <p>All linked accounts can be used for login</p>
        </section>
        <section class="security-source-section">
          <div class="security-source-title">Password Manage</div>
          <div class="security-source-card">
            <button type="button" class="security-source-row" @click="openSecurityPassword">
              <svg class="security-source-row__icon" viewBox="0 0 18 18" fill="none" aria-hidden="true">
                <path d="M6.85 8.85a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM2.2 15.15c.32-2.62 2.18-4.44 4.65-4.44 1.28 0 2.4.5 3.22 1.35" stroke="currentColor" stroke-width="1.35" stroke-linecap="round" />
                <rect x="10.6" y="10.15" width="5.25" height="5" rx="1.15" stroke="currentColor" stroke-width="1.25" />
                <path d="M11.7 10.15v-1a1.55 1.55 0 1 1 3.1 0v1M13.22 12.45v.75" stroke="currentColor" stroke-width="1.25" stroke-linecap="round" />
              </svg>
              <span>Login Password</span>
              <strong class="is-primary">Change</strong>
              <svg class="security-source-row__chevron" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </button>
            <button type="button" class="security-source-row">
              <svg class="security-source-row__icon" viewBox="0 0 18 18" fill="none" aria-hidden="true">
                <rect x="2" y="3.35" width="12.5" height="10.2" rx="1.7" stroke="currentColor" stroke-width="1.35" />
                <path d="M2.1 6.55h8.7" stroke="currentColor" stroke-width="1.35" stroke-linecap="round" />
                <rect x="10.6" y="8.5" width="5.2" height="5.05" rx="1.1" stroke="currentColor" stroke-width="1.25" />
                <path d="M11.7 8.5v-.9a1.5 1.5 0 1 1 3 0v.9M13.2 10.85v.75" stroke="currentColor" stroke-width="1.25" stroke-linecap="round" />
              </svg>
              <span>Withdraw Password</span>
              <strong>Unset</strong>
              <svg class="security-source-row__chevron" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </button>
          </div>
        </section>
        <section class="security-source-section">
          <div class="security-source-title">Withdrawal Account</div>
          <div class="security-source-card">
            <button type="button" class="security-source-row" @click="openBankAccount">
              <svg class="security-source-row__icon" viewBox="0 0 18 18" fill="none" aria-hidden="true">
                <rect x="2" y="3.2" width="12.2" height="10.4" rx="1.7" stroke="currentColor" stroke-width="1.35" />
                <path d="M2.1 6.45h7.95M11.75 8.1l.35.78.82.1-.6.58.15.82-.72-.4-.73.4.15-.82-.6-.58.83-.1.35-.78ZM14.25 10.45a3.05 3.05 0 0 1-.33 1.39l.58.6-.85.85-.61-.58a3.05 3.05 0 0 1-1.38.33l-.22.8h-1.2l-.22-.8a3.05 3.05 0 0 1-1.38-.33l-.61.58-.85-.85.58-.6a3.05 3.05 0 0 1-.33-1.39" stroke="currentColor" stroke-width="1.15" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
              <span>Account Management</span>
              <strong>0 / 5</strong>
              <svg class="security-source-row__chevron" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                <path d="M6 4l4 4-4 4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </button>
          </div>
        </section>
      </template>

      <template v-else-if="pageName === 'security-phone'">
        <nav class="security-link-tabs" aria-label="Phone account modes">
          <div class="security-link-tabs__inner">
            <button
              type="button"
              :class="{ 'is-active': securityPhoneMode === 'reset' }"
              @click="securityPhoneMode = 'reset'"
            >
              Reset
            </button>
            <button
              type="button"
              :class="{ 'is-active': securityPhoneMode === 'verify' }"
              @click="securityPhoneMode = 'verify'"
            >
              <span class="security-link-tabs__label">
                Verify
                <span class="security-link-tabs__badge" aria-hidden="true" />
              </span>
            </button>
            <span
              class="security-link-tabs__indicator"
              :style="{ left: securityPhoneMode === 'reset' ? '25%' : '75%' }"
              aria-hidden="true"
            />
          </div>
        </nav>

        <form class="security-auth-form" @submit.prevent>
          <label v-if="securityPhoneMode === 'reset'" class="security-auth-field">
            <span><i>*</i>Current Linked Phone Number</span>
            <div class="security-phone-control is-muted">
              <em>{{ runtimePhoneCode }}</em>
              <input :value="securityPhoneNumber" disabled inputmode="numeric" type="tel" />
            </div>
          </label>

          <label class="security-auth-field">
            <span><i>*</i>Phone Number</span>
            <div class="security-phone-control">
              <em>{{ runtimePhoneCode }}</em>
              <input
                v-if="securityPhoneMode === 'reset'"
                v-model="securityNewPhone"
                inputmode="numeric"
                :maxlength="runtimePhoneLength"
                placeholder="3008567530"
                type="tel"
              />
              <input v-else :value="securityPhoneNumber" inputmode="numeric" readonly type="tel" />
            </div>
          </label>

          <label class="security-auth-field">
            <span><i>*</i>Verification Code</span>
            <div class="security-code-control">
              <input v-model="securityPhoneCode" inputmode="numeric" placeholder="Enter verification code" type="tel" />
              <button type="button" :disabled="securityPhoneSendDisabled">Send</button>
            </div>
          </label>

          <label v-if="securityPhoneMode === 'reset'" class="security-auth-field">
            <span><i>*</i>Verify Login Password</span>
            <div class="security-password-control">
              <input
                v-model="securityPhonePassword"
                :type="securityPhonePasswordVisible ? 'text' : 'password'"
                placeholder="Enter"
              />
              <button
                type="button"
                aria-label="Toggle password visibility"
                @click="securityPhonePasswordVisible = !securityPhonePasswordVisible"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M2 12s3.4-6.5 10-6.5S22 12 22 12s-3.4 6.5-10 6.5S2 12 2 12Zm10 3a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z" />
                  <path v-if="!securityPhonePasswordVisible" d="M3 3l18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
                </svg>
              </button>
            </div>
          </label>

          <button class="security-complete-button" type="submit" :class="{ 'has-gift': securityPhoneMode === 'verify' }">
            Complete
            <img v-if="securityPhoneMode === 'verify'" :src="securityGiftImage" alt="" aria-hidden="true" />
          </button>
        </form>
      </template>

      <template v-else-if="pageName === 'security-password'">
        <form class="security-auth-form security-auth-form--password" @submit.prevent>
          <label class="security-auth-field">
            <span><i>*</i>New Login Password</span>
            <div class="security-password-control">
              <input
                v-model="loginPasswordNew"
                :type="loginPasswordVisible ? 'text' : 'password'"
                placeholder="Enter"
              />
              <button
                type="button"
                aria-label="Toggle new password visibility"
                @click="loginPasswordVisible = !loginPasswordVisible"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M2 12s3.4-6.5 10-6.5S22 12 22 12s-3.4 6.5-10 6.5S2 12 2 12Zm10 3a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z" />
                  <path v-if="!loginPasswordVisible" d="M3 3l18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
                </svg>
              </button>
            </div>
          </label>

          <label class="security-auth-field">
            <span><i>*</i>Confirm Login Password</span>
            <div class="security-password-control">
              <input
                v-model="loginPasswordConfirm"
                :type="loginPasswordConfirmVisible ? 'text' : 'password'"
                placeholder="Enter"
              />
              <button
                type="button"
                aria-label="Toggle confirm password visibility"
                @click="loginPasswordConfirmVisible = !loginPasswordConfirmVisible"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M2 12s3.4-6.5 10-6.5S22 12 22 12s-3.4 6.5-10 6.5S2 12 2 12Zm10 3a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z" />
                  <path v-if="!loginPasswordConfirmVisible" d="M3 3l18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
                </svg>
              </button>
            </div>
          </label>

          <label class="security-auth-field">
            <span><i>*</i>Phone Number</span>
            <div class="security-phone-control is-muted">
              <em>{{ runtimePhoneCode }}</em>
              <input :value="securityPhoneNumber" readonly inputmode="numeric" type="tel" />
            </div>
          </label>

          <div class="security-rebind-row">
            <span>Phone number unavailable?</span>
            <button type="button" @click="openSecurityPhone">
              Rebind Phone
              <svg viewBox="0 0 12 12" aria-hidden="true">
                <path d="M4.5 2.5 8 6l-3.5 3.5" />
              </svg>
            </button>
          </div>

          <label class="security-auth-field">
            <span><i>*</i>Verification Code</span>
            <div class="security-code-control">
              <input v-model="loginPasswordCode" inputmode="numeric" placeholder="Enter verification code" type="tel" />
              <button type="button" :disabled="loginPasswordSendDisabled">Send</button>
            </div>
          </label>

          <button class="security-complete-button" type="submit">Complete</button>
        </form>
      </template>

      <template v-else-if="pageName === 'bank-account'">
        <section class="withdraw-count-card">
          <span>Number of bank cards</span>
          <strong>0 / 5</strong>
          <i aria-hidden="true">?</i>
        </section>
        <section class="withdraw-method">
          <div class="withdraw-method__title">
            <span>Bank Account</span>
          </div>
          <div class="withdraw-add-list">
            <button type="button" class="withdraw-add-button" @click="openWithdrawForm('bank')">
              <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
                <path d="M10 4v12M4 10h12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
              <span>Add Bank Account Account</span>
            </button>
          </div>
        </section>
        <section class="withdraw-method">
          <div class="withdraw-method__title">
            <span>Digital Wallet</span>
          </div>
          <div class="withdraw-add-list">
            <button type="button" class="withdraw-add-button" @click="openWithdrawForm('wallet')">
              <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
                <path d="M10 4v12M4 10h12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
              <span>Add Digital Wallet Account</span>
            </button>
          </div>
        </section>
      </template>

      <template v-else-if="pageName === 'support'">
        <section class="support-source-hero">
          <div class="support-source-hero__avatars" aria-hidden="true">
            <img v-for="avatar in supportAvatars" :key="avatar" :src="avatar" alt="" loading="lazy" />
          </div>
          <button type="button">Contact now</button>
          <strong>24/7 Customer Service</strong>
          <p>Chat online to solve your problem</p>
        </section>
        <nav class="support-source-tabs" aria-label="Support tabs">
          <button
            v-for="tab in supportTabs"
            :key="tab"
            type="button"
            :class="{ 'is-active': activeSupportTab === tab }"
            @click="selectSupportTab(tab)"
          >
            {{ tab }}
          </button>
        </nav>
        <div
          ref="supportPanelRef"
          class="support-panel-scroll"
          @scroll.passive="handleSupportPanelScroll"
        >
          <div v-if="activeSupportTab === 'FAQ'" class="support-categories">
            <button
              v-for="category in supportCategories"
              :key="category"
              type="button"
              :class="{ 'is-active': activeSupportCategory === category }"
              @click="selectSupportCategory(category)"
            >
              {{ category }}
            </button>
          </div>
          <section v-if="activeSupportTab === 'FAQ'" class="faq-list" aria-label="FAQ questions">
            <button
              v-for="question in visibleFaqQuestions"
              :key="question"
              type="button"
              @click="openSupportAnswer(question)"
            >
              {{ question }}
            </button>
          </section>
          <section
            v-else-if="activeSupportTab === 'Notice' && supportNoticeItems.length"
            class="support-notice-list"
            aria-label="Notice"
          >
            <button
              v-for="notice in supportNoticeItems"
              :key="notice.id"
              type="button"
              :class="{ 'is-unread': !notice.read }"
              @click="openSupportNotice(notice)"
            >
              <span class="support-notice-list__title">
                <i v-if="!notice.read" aria-hidden="true" />
                {{ notice.title }}
              </span>
              <time v-if="formatSupportNoticeTime(notice.createdAt)">
                {{ formatSupportNoticeTime(notice.createdAt) }}
              </time>
            </button>
          </section>
          <section v-else class="support-empty-panel" aria-live="polite">
            <svg viewBox="0 0 96 96" aria-hidden="true">
              <path d="M22 31.5c.7-3 3.4-5.1 6.5-5.1h39c3.1 0 5.8 2.1 6.5 5.1l5.1 20.2a6.8 6.8 0 0 1 .1 2.7l-2.9 17.5a6.8 6.8 0 0 1-6.7 5.8H32.2" />
              <path d="M18.8 52h18.1l5.2 7.2a6.1 6.1 0 0 0 5 2.6h13.7a6.1 6.1 0 0 0 5-2.6L71 52h8.2" />
              <path d="M26.8 52V37.5a4.1 4.1 0 0 1 4.1-4.1h34.2a4.1 4.1 0 0 1 4.1 4.1V52" />
              <path d="M27.6 43.6c.9-2.2 3-3.7 5.5-3.7h29.8c2.5 0 4.7 1.5 5.5 3.7" />
              <circle cx="20.3" cy="70.2" r="11.6" />
              <path d="M20.3 63.2v6.1M20.3 75.9h.1M76.6 15.8l-.1 5.4M82.4 26.5l5.4.1" />
            </svg>
            <p>Oops! Nothing here for now.</p>
          </section>
        </div>
        <Transition name="support-answer-fade">
          <div
            v-if="activeSupportQuestion"
            class="support-answer-backdrop"
            aria-hidden="true"
            @click="closeSupportAnswer"
          />
        </Transition>
        <Transition name="support-answer-sheet">
          <aside
            v-if="activeSupportQuestion"
            class="support-answer-sheet"
            role="dialog"
            aria-modal="true"
            :aria-label="activeSupportQuestion"
          >
            <header class="support-answer-sheet__header">
              <span>{{ activeSupportQuestion }}</span>
              <button type="button" aria-label="Collapse" @click="closeSupportAnswer">
                <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
                  <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                </svg>
              </button>
            </header>
            <div class="support-answer-sheet__body" v-html="activeSupportAnswer" />
          </aside>
        </Transition>
      </template>

      <template v-else-if="pageName === 'about'">
        <section class="about-source-list">
          <article v-for="item in aboutItems" :key="item.title" class="about-source-card">
            <img :src="item.image" alt="" loading="lazy" />
            <div>
              <h2>{{ item.title }}</h2>
              <p>{{ item.text }}</p>
            </div>
          </article>
        </section>
      </template>

      <template v-else-if="pageName === 'user-agreement'">
        <article class="source-legal-page">
          <template v-for="item in agreementItems" :key="`${item.kind}-${item.text}`">
            <h2 v-if="item.kind === 'heading'">{{ item.text }}</h2>
            <h3 v-else-if="item.kind === 'subheading'">{{ item.text }}</h3>
            <p v-else :class="`is-${item.kind}`">{{ item.text }}</p>
          </template>
        </article>
      </template>

      <template v-else-if="pageName === 'settings'">
        <section class="settings-source-card">
          <button type="button" class="settings-source-row" @click="openPicker('language')">
            <span>
              <img class="settings-source-row__icon" :src="settingsLanguageIcon" alt="" aria-hidden="true" />
              Language
            </span>
            <strong>{{ currentLanguage }}</strong>
            <em aria-hidden="true">⇄</em>
          </button>
          <button type="button" class="settings-source-row" @click="toggleMusicEnabled">
            <span>
              <img class="settings-source-row__icon" :src="settingsMusicIcon" alt="" aria-hidden="true" />
              Music
            </span>
            <b class="settings-switch" :class="{ 'is-on': musicEnabled }" role="switch" :aria-checked="musicEnabled">
              <i />
            </b>
          </button>
        </section>
        <button type="button" class="settings-logout" @click="logOut">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M13.424 15.1972C13.3493 15.1972 13.291 15.1972 13.233 15.1972C11.2368 15.1972 9.2408 15.1975 7.24463 15.1972C6.94869 15.1971 6.75584 15.0615 6.70828 14.8188C6.69895 14.7704 6.69981 14.7195 6.69981 14.6699C6.69945 12.896 6.69941 11.1222 6.69957 9.3484C6.69962 9.01195 6.88183 8.83296 7.2235 8.83296C9.22866 8.83287 11.2338 8.83287 13.2391 8.83287C13.2931 8.83287 13.3474 8.83287 13.424 8.83287C13.424 8.77531 13.424 8.72345 13.424 8.67141C13.424 7.0927 13.4238 5.5139 13.4242 3.93505C13.4242 3.71781 13.5278 3.54452 13.7032 3.47816C13.9051 3.40173 14.0841 3.43693 14.2363 3.59767C14.3236 3.68987 14.4185 3.77467 14.5082 3.86432C15.6385 4.99223 16.7687 6.12067 17.8987 7.24857C19.3502 8.6969 20.8018 10.1449 22.2533 11.5932C22.495 11.834 22.4963 12.0881 22.2575 12.3264C20.1803 14.4001 18.1033 16.4739 16.0254 18.5468C15.4345 19.1363 14.8404 19.7223 14.2499 20.3123C14.098 20.4643 13.9285 20.531 13.721 20.4522C13.5193 20.3756 13.4243 20.2168 13.4243 19.9671C13.424 18.4384 13.424 16.9094 13.424 15.3807C13.424 15.3267 13.424 15.2729 13.424 15.1972ZM9.69265 3.52843C9.69265 4.12877 9.69265 4.71656 9.69265 5.32088C9.63382 5.32088 9.58086 5.32088 9.52795 5.32088C8.0082 5.32088 6.48842 5.32066 4.96869 5.32098C4.00565 5.32112 3.29645 6.03178 3.2964 6.99672C3.29626 10.3357 3.30237 13.6746 3.29237 17.0133C3.28961 17.8968 3.96169 18.5332 4.58303 18.6494C4.72034 18.6748 4.86218 18.6864 5.00193 18.6867C6.50809 18.6893 8.01427 18.6883 9.52047 18.6883C9.57398 18.6883 9.62756 18.6883 9.68992 18.6883C9.68992 19.2868 9.68992 19.8744 9.68992 20.4705C9.6685 20.4748 9.64348 20.4844 9.61842 20.4844C8.017 20.4844 6.41539 20.5007 4.81436 20.4775C3.24959 20.455 1.8639 19.2675 1.56528 17.7305C1.52314 17.514 1.50251 17.2897 1.5023 17.0691C1.49841 13.6938 1.48951 10.3186 1.50438 6.94345C1.51062 5.53366 2.21104 4.5216 3.43428 3.86389C3.86585 3.63173 4.34252 3.5269 4.83388 3.52565C6.42165 3.52136 8.00945 3.52402 9.59724 3.52411C9.62403 3.52419 9.65082 3.52654 9.69265 3.52843Z" fill="currentColor"></path>
          </svg>
          LOG OUT
        </button>
      </template>

      <ContactMeBadge
        v-if="showContact"
        class="floating-contact"
        :class="{ 'floating-contact--deposit': pageName === 'wallet-deposit' }"
        :image="customerImage"
        draggable
        @click="openContact"
      />
    </main>

    <Transition name="order-help">
      <section
        v-if="showOrderDescription"
        class="order-help-overlay"
        role="dialog"
        aria-modal="true"
        aria-label="Order Description"
      >
        <div class="order-help-backdrop" @click="closeOrderDescription" />
        <article class="order-help-dialog">
          <header>
            <span>Order Description</span>
            <button type="button" aria-label="Close" @click="closeOrderDescription">
              <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
                <path d="M5 5l10 10M15 5 5 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
          </header>
          <div class="order-help-body">
            <div>
              <div class="order-help-section-title">· Order Status Description</div>
              <p>The order status update mechanism varies by game.</p>
              <p>As a result, the calculation of {Valid bets} and {Win/Loss} Amount may be delayed.</p>
              <p>Please refer to the settlement status within the corresponding game for the actual result.</p>
              <p class="is-strong">Don't worry. Once the game ends, your order will be settled correctly.</p>
            </div>
            <div class="order-help-divider" />
            <div>
              <div class="order-help-section-title">· Valid bets Description</div>
              <p>Different game types have different Valid bets calculation methods.</p>
              <div class="order-help-block">
                <p>1、SLOT, FISH</p>
                <p class="is-strong">→ Valid bets = Bet Amount</p>
              </div>
              <div class="order-help-block">
                <p>2、CHANCE, BINGO, TABLE, LIVE</p>
                <div class="order-help-nested">
                  <div>
                    <p>2.1) If Payout &gt; Bet amount(Win), calculation depends on {Win amount}</p>
                    <div class="order-help-nested">
                      <p>2.1.1) If Win amount ≥ Bet amount</p>
                      <p class="is-strong">→ Valid bets = Bet Amount</p>
                      <p>2.1.2) If Win amount &lt; Bet amount</p>
                      <p class="is-strong">→ Valid bets = Win Amount</p>
                    </div>
                  </div>
                  <div>
                    <p>2.2) If Payout = Bet amount(Draw)</p>
                    <p class="is-strong">→ Valid bets = 0.00</p>
                  </div>
                  <div>
                    <p>2.3) If Payout &lt; Bet amount(Loss)</p>
                    <p class="is-strong">→ Valid bets = Loss Amount</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <footer>
            <button type="button" @click="closeOrderDescription">Okay</button>
          </footer>
        </article>
      </section>
    </Transition>

    <Transition name="picker-sheet">
      <section
        v-if="activePicker"
        class="picker-sheet-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="activePickerTitle"
      >
        <div class="picker-sheet-backdrop" @click="closePicker" />
        <article class="picker-sheet">
          <header>
            <h2>{{ activePickerTitle }}</h2>
            <button type="button" aria-label="Close" @click="closePicker">
              <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
                <path d="M5 5l10 10M15 5 5 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
          </header>
          <div class="picker-sheet__options">
            <button
              v-for="option in activePickerOptions"
              :key="option"
              type="button"
              class="picker-sheet__option"
              :class="{ 'is-active': option === activePickerValue }"
              @click="selectPickerOption(option)"
            >
              <span>{{ option }}</span>
              <svg v-if="option === activePickerValue" viewBox="0 0 16 16" fill="none" aria-hidden="true">
                <path d="M3 8.1 6.2 11.2 13 4.8" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </button>
          </div>
        </article>
      </section>
    </Transition>

    <Transition name="withdraw-form">
      <section
        v-if="activeWithdrawForm"
        class="withdraw-form-overlay"
        role="dialog"
        aria-modal="true"
        :aria-label="activeWithdrawForm === 'bank' ? 'Bank card details' : 'E-wallet details'"
      >
        <div class="withdraw-form-backdrop" @click="closeWithdrawForm" />
        <article class="withdraw-form-sheet">
          <header>
            <h2>{{ activeWithdrawForm === 'bank' ? 'Bank card details' : 'E-wallet details' }}</h2>
            <button type="button" aria-label="Close" @click="closeWithdrawForm">
              <svg viewBox="0 0 20 20" fill="none" aria-hidden="true">
                <path d="M5 5l10 10M15 5 5 15" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
              </svg>
            </button>
          </header>

          <div class="withdraw-form-body">
            <label v-if="activeWithdrawForm === 'bank'" class="withdraw-form-field">
              <span><i>*</i>Choose Method</span>
              <button type="button" class="withdraw-select-button">
                Select Bank
                <svg viewBox="0 0 12 12" fill="none" aria-hidden="true">
                  <path d="M3 4.5 6 7.5l3-3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
                </svg>
              </button>
            </label>

            <div v-else class="withdraw-form-field">
              <span><i>*</i>Wallet type</span>
              <div class="wallet-type-grid">
                <button
                  v-for="wallet in walletTypes"
                  :key="wallet"
                  type="button"
                  :class="{ 'is-active': selectedWalletType === wallet }"
                  @click="selectWalletType(wallet)"
                >
                  <b :class="wallet === 'JazzCash' ? 'is-jazzcash' : 'is-easypaisa'" aria-hidden="true" />
                  {{ wallet }}
                </button>
              </div>
            </div>

            <label class="withdraw-form-field">
              <span><i>*</i>Real name</span>
              <input placeholder="Enter" autocomplete="off" />
            </label>

            <label class="withdraw-form-field">
              <span><i>*</i>{{ activeWithdrawForm === 'bank' ? 'Bank Account Number' : 'Account' }}</span>
              <div class="withdraw-input-with-action">
                <input inputmode="numeric" type="tel" placeholder="Enter account number" />
                <button type="button">Paste</button>
              </div>
            </label>

            <label class="withdraw-form-field">
              <span><i>*</i>Phone Number</span>
              <div class="withdraw-phone-input">
                <em>{{ runtimePhoneCode }}</em>
                <input inputmode="numeric" type="tel" :value="mockUser.phone" readonly />
              </div>
            </label>

            <p class="withdraw-form-reminder">
              <strong>Reminder：</strong>
              Withdrawal account cannot be edited. Please confirm before submitting!
            </p>

            <button type="button" class="withdraw-form-confirm" @click="closeWithdrawForm">
              Confirm
            </button>
          </div>
        </article>
      </section>
    </Transition>
  </div>
</template>

<style scoped lang="scss">
.profile-subpage {
  min-height: 100vh;
  min-height: 100dvh;
  max-width: 430px;
  margin: 0 auto;
  overflow: hidden;
  color: #ffffff;
  background: #090f1f;
}

.subpage-header {
  position: sticky;
  top: 0;
  z-index: 5;
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr) 96px;
  align-items: center;
  min-height: calc(env(safe-area-inset-top) + 54px);
  padding: env(safe-area-inset-top) 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(10, 16, 31, 0.96);
  backdrop-filter: blur(14px);

  > button:first-child {
    display: grid;
    width: 38px;
    height: 38px;
    place-items: center;
    color: #ffffff;
  }

  svg {
    width: 24px;
    height: 24px;
    fill: currentColor;
  }

  h1 {
    overflow: hidden;
    margin: 0;
    font-size: 0.875rem;
    font-weight: 700;
    line-height: 1.2;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.contact-mini {
  justify-self: end;
  height: 29px;
  padding: 0 10px;
  border-radius: 999px;
  color: #003278;
  font-size: 11px;
  font-weight: 700;
  background: linear-gradient(180deg, #baefff 0%, #38cfff 100%);
}

.subpage-header__spacer {
  width: 1px;
}

.subpage-content {
  height: calc(100vh - env(safe-area-inset-top) - 54px);
  height: calc(100dvh - env(safe-area-inset-top) - 54px);
  overflow-x: hidden;
  overflow-y: auto;
  padding: 14px 14px calc(env(safe-area-inset-bottom) + 22px);
}

.segmented-tabs,
.filter-row,
.chip-grid,
.support-categories {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.segmented-tabs {
  padding: 4px;
  border-radius: 10px;
  background: #151c31;

  button {
    flex: 1 0 auto;
    min-width: 88px;
    height: 36px;
    border-radius: 8px;
    color: rgba(255, 255, 255, 0.58);
    font-size: 13px;
    font-weight: 700;
    transition: color 180ms ease, background 180ms ease;

    &.is-active {
      color: #ffffff;
      background: linear-gradient(180deg, #438cf0 0%, #2f73de 100%);
    }
  }
}

.empty-panel,
.form-section,
.bonus-choice,
.deposit-summary,
.balance-card,
.info-card,
.account-add-card,
.support-hero,
.article-list article,
.legal-page {
  margin-top: 14px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  background: #1a2135;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.03);
}

.empty-panel {
  display: grid;
  min-height: 220px;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: rgba(255, 255, 255, 0.55);
  text-align: center;

  p {
    margin: 0;
    font-size: 14px;
  }
}

.empty-panel__icon {
  display: grid;
  width: 54px;
  height: 54px;
  place-items: center;
  border-radius: 50%;
  color: #7f93c7;
  font-size: 28px;
  font-weight: 800;
  background: #11182a;
}

.form-section,
.bonus-choice,
.info-card,
.account-add-card,
.article-list article,
.legal-page {
  padding: 14px;

  h2 {
    margin: 0 0 12px;
    color: #ffffff;
    font-size: 14px;
    font-weight: 700;
  }
}

.chip-grid,
.channel-list,
.amount-grid {
  display: grid;
  gap: 8px;
}

.chip-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.chip-grid button,
.channel-list button,
.filter-row button,
.support-categories button,
.filter-button {
  min-height: 36px;
  padding: 0 12px;
  border-radius: 8px;
  color: #dce7ff;
  font-size: 12px;
  font-weight: 700;
  background: #11182a;
}

.channel-list {
  grid-template-columns: 1fr;
}

.amount-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));

  button {
    display: grid;
    gap: 3px;
    min-height: 52px;
    padding: 8px;
    border: 1px solid rgba(77, 129, 231, 0.34);
    border-radius: 8px;
    color: #ffffff;
    background: #121b31;

    span,
    em {
      font-style: normal;
      line-height: 1.1;
    }

    em {
      color: #f9cf5a;
      font-size: 11px;
    }
  }
}

.amount-input {
  display: flex;
  align-items: center;
  height: 42px;
  margin-top: 10px;
  border: 1px solid #364169;
  border-radius: 8px;
  background: #131828;

  span {
    padding: 0 12px;
    color: rgba(255, 255, 255, 0.72);
    font-size: 13px;
  }

  input {
    min-width: 0;
    flex: 1;
    border: 0;
    outline: 0;
    color: #ffffff;
    background: transparent;
  }
}

.bonus-choice {
  display: grid;
  gap: 10px;

  article {
    display: grid;
    gap: 5px;
    padding: 10px;
    border-radius: 8px;
    background: #11182a;
  }

  strong {
    font-size: 13px;
  }

  span {
    color: rgba(255, 255, 255, 0.58);
    font-size: 12px;
  }

  em {
    color: #f8ce54;
    font-size: 12px;
    font-style: normal;
    font-weight: 800;
  }
}

.deposit-summary {
  padding: 14px;

  h2 {
    margin: 0 0 10px;
    color: rgba(255, 255, 255, 0.72);
    font-size: 13px;
  }

  strong {
    display: block;
    color: #ffffff;
    font-size: 24px;
    line-height: 1;
  }

  span {
    display: block;
    margin-top: 8px;
    color: rgba(255, 255, 255, 0.58);
    font-size: 12px;
  }
}

.primary-action,
.danger-action,
.contact-action {
  width: 100%;
  height: 45px;
  margin-top: 14px;
  border-radius: 8px;
  color: #ffffff;
  font-size: 15px;
  font-weight: 800;
}

.primary-action {
  background: linear-gradient(180deg, #438cf0 0%, #2f73de 100%);
}

.danger-action {
  background: linear-gradient(180deg, #ff5d6c 0%, #e3364a 100%);
}

.contact-action {
  background: linear-gradient(180deg, #baefff 0%, #38cfff 100%);
  color: #003278;
}

.balance-card {
  display: grid;
  gap: 10px;
  padding: 16px;
  background:
    radial-gradient(circle at 100% 0%, rgba(81, 190, 240, 0.32), transparent 44%),
    #1a2135;

  span {
    color: rgba(255, 255, 255, 0.7);
    font-size: 13px;
  }

  strong {
    font-size: 34px;
    line-height: 1;

    small {
      margin-right: 5px;
      font-size: 16px;
    }
  }

  button {
    height: 38px;
    border-radius: 8px;
    color: #ffffff;
    font-size: 14px;
    font-weight: 800;
    background: linear-gradient(180deg, #67be68 0%, #4b9f4c 100%);
  }
}

.muted-note {
  margin: 12px 0 0;
  color: rgba(255, 255, 255, 0.58);
  font-size: 13px;
}

.quick-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;

  button {
    height: 42px;
    border-radius: 8px;
    color: #dce7ff;
    font-size: 13px;
    font-weight: 700;
    background: #1a2135;
  }
}

.filter-button {
  width: max-content;
}

.record-list {
  display: grid;
  gap: 10px;
  margin-top: 14px;
}

.record-card {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  background: #1a2135;

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 12px 14px;
    background: #202842;

    strong,
    span {
      min-width: 0;
      line-height: 1.2;
    }

    strong {
      color: #ffffff;
      font-size: 15px;
      font-weight: 800;
    }

    span {
      color: #7fff9c;
      font-size: 14px;
      font-weight: 800;
      text-align: right;
    }
  }

  dl {
    display: grid;
    gap: 8px;
    margin: 0;
    padding: 12px 14px;
  }

  div {
    display: flex;
    justify-content: space-between;
    gap: 14px;
  }

  dt,
  dd {
    margin: 0;
    font-size: 12px;
    line-height: 1.35;
  }

  dt {
    flex: 0 0 auto;
    color: rgba(255, 255, 255, 0.52);
  }

  dd {
    min-width: 0;
    color: rgba(255, 255, 255, 0.82);
    text-align: right;
    overflow-wrap: anywhere;
  }

  .is-failed {
    color: #ff6873;
  }

  .is-pending {
    color: #ffd05f;
  }

  .is-succeeded {
    color: #70f08a;
  }
}

.filter-row {
  margin-top: 12px;
}

.data-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-height: 42px;
  padding: 0;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  color: #ffffff;
  text-align: left;

  span {
    color: rgba(255, 255, 255, 0.72);
    font-size: 13px;
  }

  strong {
    color: #ffffff;
    font-size: 13px;
    font-weight: 700;
  }

  &.is-clickable:active {
    opacity: 0.75;
  }
}

.info-card p {
  margin: 8px 0 0;
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
}

.large-number {
  display: block;
  font-size: 28px;
}

.account-add-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;

  h2 {
    margin: 0;
  }

  button {
    height: 36px;
    flex: 0 0 auto;
    padding: 0 12px;
    border-radius: 8px;
    color: #ffffff;
    font-size: 12px;
    font-weight: 800;
    background: linear-gradient(180deg, #438cf0 0%, #2f73de 100%);
  }
}

.support-hero {
  padding: 16px;
  background:
    radial-gradient(circle at 95% 20%, rgba(77, 129, 231, 0.45), transparent 42%),
    #1a2135;

  span,
  p {
    color: rgba(255, 255, 255, 0.68);
    font-size: 13px;
  }

  strong {
    display: block;
    margin-top: 5px;
    font-size: 21px;
  }

  p {
    margin: 8px 0 14px;
  }

  div {
    display: flex;
    gap: 10px;
  }

  button {
    height: 34px;
    padding: 0 16px;
    border-radius: 999px;
    color: #ffffff;
    font-size: 13px;
    font-weight: 800;
    background: linear-gradient(180deg, #438cf0 0%, #2f73de 100%);
  }
}

.support-categories {
  flex-wrap: wrap;
  margin-top: 12px;
}

.article-list {
  display: grid;
  gap: 12px;

  article {
    margin: 0;
  }

  p {
    margin: 0;
    color: rgba(255, 255, 255, 0.68);
    font-size: 13px;
    line-height: 1.55;
  }
}

.legal-page {
  color: rgba(255, 255, 255, 0.72);
  font-size: 13px;
  line-height: 1.62;

  p {
    margin: 0 0 14px;
  }
}

.profile-subpage {
  position: relative;
  font-family: Montserrat, Arial, Helvetica, sans-serif;
  background: #090f1f;
}

.subpage-header {
  grid-template-columns: 42px minmax(0, 1fr) 42px;
  min-height: calc(env(safe-area-inset-top) + 50px);
  padding: env(safe-area-inset-top) 14px 0;
  border-bottom-color: rgba(230, 241, 255, 0.1);
  background: #031326;
  backdrop-filter: none;

  > button:first-child {
    justify-self: start;
    width: 24px;
    height: 24px;
  }

  h1 {
    font-size: 0.875rem;
    font-weight: 500;
  }
}

.subpage-header__spacer {
  width: 42px;
}

.subpage-content {
  height: calc(100vh - env(safe-area-inset-top) - 50px);
  height: calc(100dvh - env(safe-area-inset-top) - 50px);
  padding: 14px 14px calc(env(safe-area-inset-bottom) + 22px);
  background: #090f1f;
}

.subpage-content--wallet-deposit {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 14px 14px 0;
}

.deposit-panel {
  position: relative;
  min-height: 0;
  flex: 1 1 auto;
  overflow-x: hidden;
  overflow-y: auto;
  padding-bottom: 14px;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.deposit-title-row {
  display: flex;
  align-items: center;
  min-height: 24px;
  margin-bottom: 10px;
  color: #ffffff;

  h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 500;
    line-height: 1.2;
  }
}

.deposit-title-row__icon {
  display: block;
  width: 22px;
  height: 22px;
  flex: 0 0 22px;
  margin-right: 8px;
  object-fit: contain;
}

.deposit-refresh {
  display: grid;
  width: 28px;
  height: 28px;
  margin-left: auto;
  place-items: center;
  color: #397fe5;

  svg {
    width: 20px;
    height: 20px;
    fill: currentColor;
  }
}

.deposit-method-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 7px;
  margin-bottom: 10px;
}

.deposit-method-grid button,
.channel-list button {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  height: 37px;
  min-width: 0;
  overflow: hidden;
  gap: 6px;
  border: 1px solid #1e2438;
  color: #ffffff;
  font-size: 13px;
  font-weight: 500;
  line-height: 1;
  text-align: left;
  background: #1e2438;

  &.is-active {
    border-color: #387ee5;
    background: linear-gradient(277deg, #2e3754 3%, #46537f 96%);
    box-shadow: 0 0 8px rgba(56, 126, 229, 0.5);
  }

  > span:last-child {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.deposit-method-grid button {
  padding: 0 10px;
  border-radius: 10px;
}

.channel-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  margin-bottom: 17px;

  button {
    width: 100%;
    padding: 0 10px;
    border-radius: 999px;
  }
}

.deposit-logo {
  display: block;
  width: 21px;
  height: 21px;
  flex: 0 0 21px;
  overflow: hidden;
  border-radius: 4px;
  object-fit: contain;
}

.amount-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 7px;
  margin-bottom: 15px;

  button {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 50px;
    padding: 0 6px;
    overflow: hidden;
    border: 1px solid transparent;
    border-radius: 8px;
    color: #ffffff;
    background: #1e2438;

    &.is-active {
      border-color: #387ee5;
      background: linear-gradient(277deg, #2e3754 3%, #46537f 96%);
      box-shadow: 0 0 8px rgba(56, 126, 229, 0.5);
    }

    span {
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 1;
      white-space: nowrap;
    }

    em {
      position: absolute;
      top: 0;
      right: 0;
      display: inline-flex;
      min-width: 44px;
      height: 15px;
      align-items: center;
      justify-content: center;
      padding: 0 4px;
      border-radius: 0 0 0 6px;
      color: #ffffff;
      font-size: 10px;
      font-style: italic;
      font-weight: 600;
      line-height: 1;
      background: #387ee5;
    }
  }
}

.deposit-amount-input {
  display: flex;
  height: 37px;
  align-items: center;
  overflow: hidden;
  border: 1px solid #2d3652;
  border-radius: 8px;
  background: transparent;

  > span {
    display: flex;
    width: 53px;
    height: 100%;
    align-items: center;
    padding-left: 18px;
    color: #ffffff;
    font-size: 16px;
    background: #1e2438;
  }

  input {
    min-width: 0;
    height: 100%;
    flex: 1;
    border: 0;
    outline: 0;
    padding: 0 12px;
    color: #ffffff;
    font-size: 16px;
    background: transparent;
  }

  button {
    display: grid;
    width: 30px;
    height: 100%;
    place-items: center;
    color: #8c98bb;
    font-size: 15px;
  }
}

.deposit-label-row {
  display: flex;
  height: 38px;
  align-items: center;
  justify-content: space-between;
  margin-top: 10px;
  color: #ffffff;

  > span {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #ffffff;
    font-size: 14px;
    font-weight: 500;
  }

  button {
    display: flex;
    align-items: center;
    gap: 5px;
    color: #397fe5;
    font-size: 14px;
    font-weight: 500;
  }

  i {
    width: 0;
    height: 0;
    border-top: 5px solid #397fe5;
    border-right: 5px solid transparent;
    border-left: 5px solid transparent;
  }
}

.deposit-label-row__icon {
  display: block;
  width: 20px;
  height: 20px;
  object-fit: contain;
}

.voucher-strip {
  display: flex;
  gap: 8px;
  width: 100%;
  overflow-x: auto;
  padding: 0 0 6px;
  overscroll-behavior-x: contain;
  scroll-behavior: smooth;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
  touch-action: pan-x;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar {
    display: none;
  }

  &::after {
    display: block;
    flex: 0 0 56px;
    content: "";
  }
}

.voucher-ticket {
  position: relative;
  display: block;
  width: 88%;
  height: 72px;
  flex: 0 0 88%;
  padding: 0;
  overflow: hidden;
  border-radius: 7px;
  color: #2d2410;
  text-align: left;
  scroll-snap-align: start;
  transition: opacity 180ms ease, transform 180ms ease;

  &.is-active {
    transform: translateY(-1px);
  }

  &.is-muted {
    opacity: 0.78;
  }

  &.is-disabled {
    cursor: not-allowed;
    opacity: 0.52;
  }

  > *:not(.voucher-ticket__shape) {
    z-index: 1;
  }

  > img {
    position: absolute;
    top: 50%;
    left: calc((21.3% - 50px) / 2);
    z-index: 1;
    width: 50px;
    height: 50px;
    border-radius: 6px;
    object-fit: contain;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    transform: translateY(-50%);
  }

  > strong {
    position: absolute;
    top: 10px;
    right: 44px;
    left: 78px;
    z-index: 1;
    overflow: visible;
    font-size: 14px;
    font-weight: 800;
    line-height: 16.8px;
    text-overflow: clip;
    white-space: nowrap;
  }

  > span:not(.voucher-ticket__check) {
    position: absolute;
    top: 31px;
    right: 44px;
    left: 78px;
    z-index: 1;
    overflow: visible;
    color: #2d2410;
    font-size: 14px;
    font-weight: 600;
    line-height: 16.8px;
    text-overflow: clip;
    white-space: nowrap;
  }

  > small {
    position: absolute;
    top: 49px;
    right: 44px;
    left: 78px;
    z-index: 1;
    overflow: visible;
    color: rgba(45, 36, 16, 0.72);
    font-size: 10px;
    font-weight: 500;
    line-height: 12px;
    text-overflow: clip;
    white-space: nowrap;
  }

  > em {
    position: absolute;
    right: 17px;
    bottom: 2px;
    z-index: 1;
    min-width: 86px;
    padding: 3px 7px;
    border-radius: 8px 8px 0 0;
    color: #ffffff;
    font-size: 11px;
    font-style: italic;
    font-weight: 700;
    line-height: 1;
    text-align: center;
    background: #c88200;
    box-shadow: inset 1px 1px 1px rgba(255, 255, 255, 0.25);

    span {
      color: rgba(255, 255, 255, 0.9);
      font-weight: 400;
    }
  }
}

.voucher-ticket__shape {
  position: absolute;
  inset: 2px;
  width: auto;
  height: auto;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.22));
}

.voucher-ticket__dash {
  stroke: #090f1e;
  stroke-dasharray: 2 2;
}

.voucher-ticket__check {
  position: absolute;
  top: 11px;
  right: 18px;
  display: grid;
  z-index: 2;
  width: 22px;
  height: 22px;
  place-items: center;

  svg {
    width: 22px;
    height: 22px;
  }

  path:first-child {
    fill: #1e2438;
    stroke: #505470;
    stroke-linejoin: round;
  }

  path:last-child {
    stroke: #edd496;
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.voucher-ticket.is-active .voucher-ticket__check path:first-child {
  stroke: #edd496;
}

.voucher-ticket.is-disabled .voucher-ticket__check path:first-child {
  fill: #8b6311;
  stroke: none;
}

.deposit-sticky-action {
  position: static;
  z-index: 2;
  flex: 0 0 auto;
  margin: 0 -14px;
  padding: 13px 14px calc(env(safe-area-inset-bottom) + 24px);
  background: #090f1f;
}

.deposit-return-wrap {
  margin-bottom: 24px;
}

.deposit-return-row {
  display: flex;
  width: 100%;
  min-height: 22px;
  align-items: center;
  justify-content: space-between;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;

  span {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  svg {
    transition: transform 280ms cubic-bezier(0.32, 0.72, 0, 1);

    &.is-expanded {
      transform: rotate(180deg);
    }
  }

  strong {
    color: #00d982;
    font-size: 16px;
    font-weight: 600;
  }
}

.deposit-return-details {
  max-height: 0;
  overflow: hidden;
  transition: max-height 280ms cubic-bezier(0.32, 0.72, 0, 1);
}

.deposit-return-wrap.is-expanded .deposit-return-details {
  max-height: 78px;
}

.deposit-return-details__inner {
  padding-top: 9px;
  transform: translateZ(0);
  backface-visibility: hidden;
}

.deposit-return-detail {
  display: flex;
  align-items: center;
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  font-weight: 400;
  line-height: 1.25;

  + .deposit-return-detail {
    margin-top: 8px;
  }

  > span:nth-child(2) {
    flex: 1;
  }

  strong {
    color: #ffffff;
    font-size: 14px;
    font-weight: 400;
  }
}

.deposit-return-detail__dot {
  width: 6px;
  height: 6px;
  margin-right: 6px;
  border-radius: 50%;

  &.is-white {
    background: #ffffff;
  }

  &.is-violet {
    background: #a78bfa;
  }

  &.is-amber {
    background: #fbbf24;
  }
}

.primary-action {
  height: 48px;
  margin-top: 0;
  border-radius: 6px;
  font-size: 17px;
  font-weight: 500;
  background: linear-gradient(118deg, #3d8aef, #2f69d1);
  box-shadow: 0 5px 12px rgba(47, 105, 209, 0.35);

  &:disabled {
    cursor: not-allowed;
    opacity: 0.58;
  }
}

.deposit-progress-overlay {
  position: fixed;
  inset: 0;
  z-index: 120;
  display: flex;
  width: 100%;
  max-width: 430px;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  padding: 0 20px;
  background: rgba(0, 0, 0, 0.62);
}

.deposit-progress-dialog {
  position: relative;
  width: 100%;
  overflow: visible;
  border-radius: 16px;
  color: #ffffff;
  background: #1e2438;
  transform-origin: center center;
}

.deposit-progress-dialog__scroll {
  overflow: visible;
}

.deposit-progress-dialog__content {
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow: visible;
  padding: 0 16px 16px;
}

.deposit-progress-animation {
  width: 160px;
  height: 160px;
  margin-top: -80px;
}

.deposit-progress-title {
  margin-top: 12px;
  color: #ffffff;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.2;
}

.deposit-progress-order {
  margin-top: 12px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 14px;
  font-weight: 500;
  line-height: 1.3;
}

.deposit-progress-primary,
.deposit-progress-secondary {
  width: 100%;
  height: 46px;
  border: 0;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 500;
}

.deposit-progress-primary {
  margin-top: 24px;
  color: #ffffff;
  background: linear-gradient(118deg, #3d8aef, #2f69d1);
  box-shadow: 0 5px 12px rgba(47, 105, 209, 0.35);
}

.deposit-progress-secondary {
  margin-top: 10px;
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}

.deposit-progress-enter-active,
.deposit-progress-leave-active {
  transition: opacity 220ms ease;

  .deposit-progress-dialog {
    transition: transform 220ms cubic-bezier(0.32, 0.72, 0, 1), opacity 220ms ease;
  }
}

.deposit-progress-enter-from,
.deposit-progress-leave-to {
  opacity: 0;

  .deposit-progress-dialog {
    opacity: 0;
    transform: scale(0.92);
  }
}

.deposit-voucher-overlay {
  position: fixed;
  inset: 0;
  z-index: 75;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
}

.deposit-voucher-sheet {
  width: min(430px, 100%);
  max-height: 78vh;
  overflow: hidden;
  border-radius: 16px 16px 0 0;
  background: #090f1f;
  box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.4);
}

.deposit-voucher-sheet__handle {
  display: flex;
  justify-content: center;
  padding: 12px 0 8px;

  span {
    width: 40px;
    height: 4px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.3);
  }
}

.deposit-voucher-sheet__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px 12px;
  color: #ffffff;

  > span {
    font-size: 16px;
    font-weight: 500;
  }

  button {
    display: grid;
    width: 30px;
    height: 30px;
    place-items: center;
    color: rgba(255, 255, 255, 0.55);
  }

  svg {
    width: 18px;
    height: 18px;
  }
}

.deposit-voucher-sheet__list {
  display: grid;
  gap: 8px;
  max-height: calc(78vh - 68px);
  overflow-y: auto;
  padding: 0 16px calc(env(safe-area-inset-bottom) + 24px);
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.voucher-ticket--sheet {
  width: 100%;
  flex: 0 0 auto;
  scroll-snap-align: none;

  > strong,
  > span:not(.voucher-ticket__check),
  > small {
    left: calc(21.8% + 10px);
  }
}

.floating-contact {
  position: fixed;
  right: max(calc((100vw - 430px) / 2 + 12px), 12px);
  top: min(476px, calc(100dvh - 148px));
  z-index: 40;
  width: 60px;
  color: #003278;

  img {
    width: 55px;
    height: 55px;
    object-fit: contain;
  }

  span {
    display: block;
    max-width: 68px;
    margin-top: -6px;
    padding: 3px 7px;
    border: 1px solid #00a1d5;
    border-radius: 999px;
    color: #003278;
    font-size: 8px;
    font-weight: 700;
    line-height: 1;
    white-space: nowrap;
    background: linear-gradient(#baefff 0%, #38cfff 100%);
  }
}

.floating-contact--deposit {
  top: min(522px, calc(100dvh - 118px));
}

.record-card,
.info-card,
.account-add-card,
.support-hero,
.article-list article,
.legal-page,
.empty-panel,
.balance-card {
  border-color: rgba(230, 241, 255, 0.08);
  border-radius: 8px;
  background: #1e2438;
  box-shadow: none;
}

.record-card header {
  background: #252d45;
}

.record-card dl,
.info-card,
.article-list article,
.legal-page {
  padding: 12px;
}

.segmented-tabs {
  padding: 0;
  overflow: hidden;
  border-radius: 9px;
  background: #202842;

  button {
    height: 42px;
    border-radius: 0;
    color: rgba(255, 255, 255, 0.68);
    font-size: 14px;
    font-weight: 500;

    &.is-active {
      border-radius: 9px;
      background: linear-gradient(180deg, #4f8ff0 0%, #2f69d1 100%);
    }
  }
}

.subpage-content--wallet-records,
.subpage-content--bet-records {
  padding-top: 0;

  > .segmented-tabs:first-child {
    height: 46px;
    margin: 0 -14px 16px;
    border-radius: 0;
    background: #090f1f;

    button {
      position: relative;
      height: 46px;
      border-radius: 0;
      color: rgba(255, 255, 255, 0.68);
      font-size: 14px;
      font-weight: 400;

      &.is-active {
        color: #397fe5;
        font-weight: 500;
        background: transparent;

        &::after {
          position: absolute;
          right: 28px;
          bottom: 0;
          left: 28px;
          height: 2px;
          border-radius: 999px 999px 0 0;
          background: #397fe5;
          content: "";
        }
      }
    }
  }

  .record-list {
    margin-top: 0;
  }
}

.subpage-content--wallet-records {
  .wallet-filter-row {
    display: grid;
    gap: 8px;
    margin-top: -3px;
    margin-bottom: 14px;
    overflow: visible;
  }

  .wallet-filter-row--single {
    grid-template-columns: minmax(0, 1fr);
  }

  .wallet-filter-row--statement {
    grid-template-columns: 0.9fr 1fr 1fr;
  }

  .wallet-filter-button {
    display: flex;
    min-width: 0;
    height: 33px;
    min-height: 33px;
    align-items: center;
    justify-content: space-between;
    gap: 7px;
    padding: 0 12px;
    border-radius: 999px;
    color: #ffffff;
    font-size: 12.5px;
    font-weight: 500;
    line-height: 1;
    background: #2d3652;

    span {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    svg {
      width: 12px;
      height: 12px;
      flex: 0 0 12px;
      color: rgba(255, 255, 255, 0.62);
    }
  }
}

.subpage-content--bet-records {
  .filter-row {
    display: grid;
    grid-template-columns: 0.9fr 1fr 1fr;
    gap: 8px;
    margin-top: -3px;
    overflow: visible;
  }

  .filter-row--report {
    grid-template-columns: 0.78fr minmax(0, 1fr);
  }

  .bet-filter-button {
    display: flex;
    min-width: 0;
    height: 33px;
    min-height: 33px;
    align-items: center;
    justify-content: space-between;
    gap: 7px;
    padding: 0 12px;
    border-radius: 999px;
    color: #ffffff;
    font-size: 12.5px;
    font-weight: 500;
    line-height: 1;
    background: #2d3652;

    &.bet-filter-button--wide {
      justify-content: space-between;
    }

    svg {
      width: 12px;
      height: 12px;
      flex: 0 0 12px;
      color: rgba(255, 255, 255, 0.62);
    }
  }
}

.source-empty-state {
  position: relative;
  display: flex;
  min-height: 330px;
  flex-direction: column;
  align-items: center;
  margin-top: 14px;
  color: #97a5ca;
  text-align: center;
}

.source-empty-state__top {
  position: absolute;
  top: -40px;
  right: 0;
  left: 0;
  display: flex;
  height: 40px;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

.source-empty-state__body {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 56px;

  svg {
    width: 96px;
    height: 96px;
    color: #97a5ca;
  }
}

.source-empty-state p {
  margin: 24px 0 0;
  color: #97a5ca;
  font-size: 16px;
  font-weight: 600;
  line-height: 20px;
}

.wallet-source-record-list {
  padding: 4px 0 18px;
}

.wallet-source-record-card {
  margin-bottom: 10px;
  border-radius: 8px;
  background: #1e2438;
  padding: 14px;
}

.wallet-source-record-card__summary,
.wallet-source-record-card__meta,
.wallet-source-record-card__source {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.wallet-record-load-state {
  padding: 1rem 0 calc(1rem + env(safe-area-inset-bottom));
  color: rgba(255, 255, 255, 0.5);
  font-size: 0.75rem;
  text-align: center;
}

.wallet-source-record-card__summary {
  span {
    min-width: 0;
    overflow: hidden;
    color: #ffffff;
    font-size: 1rem;
    font-weight: 500;
    line-height: 1;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    flex-shrink: 0;
    color: #74be82;
    font-size: 1.125rem;
    font-weight: 500;
    line-height: 1;

    &.is-debit {
      color: #ff7777;
    }
  }
}

.wallet-source-record-card__meta {
  margin-top: 10px;

  span {
    min-width: 0;
    overflow: hidden;
    color: rgba(255, 255, 255, 0.7);
    font-size: 0.75rem;
    font-weight: 400;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.wallet-source-record-card__source {
  margin-top: 10px;

  span {
    flex-shrink: 0;
    color: rgba(255, 255, 255, 0.7);
    font-size: 0.75rem;
    font-weight: 400;
  }

  strong {
    min-width: 0;
    overflow: hidden;
    color: #ffffff;
    font-size: 0.875rem;
    font-weight: 400;
    text-align: right;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.subpage-content--wallet-records .record-card {
  min-height: 135px;
  border: 0;
  border-radius: 8px;

  header {
    min-height: 43px;
    padding: 0 14px;
    background: #2a3352;
  }

  dl {
    gap: 12px;
    padding: 14px;
  }
}

.record-order-value {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;

  span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  button {
    display: grid;
    width: 18px;
    height: 18px;
    flex: 0 0 18px;
    place-items: center;
    color: #397fe5;

    svg {
      width: 14px;
      height: 14px;
    }
  }
}

.subpage-content--bonus-center {
  .bonus-filter-button {
    display: flex;
    width: 100%;
    height: 42px;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: -2px;
    border-radius: 6px;
    background: #2d3652;

    span {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .bonus-filter-button__icon {
    width: 12px;
    height: 12px;
    flex: 0 0 12px;
    margin-top: 1px;
  }

  .record-card {
    position: relative;
    min-height: 146px;
    overflow: visible;
    border: 0;
    border-radius: 10px;
    background: #1e2438;

    &::before {
      position: absolute;
      top: -4px;
      right: 23px;
      z-index: 1;
      min-width: 58px;
      padding: 3px 8px;
      border-radius: 0 0 6px 6px;
      color: #ffffff;
      font-size: 10px;
      font-weight: 600;
      line-height: 1.2;
      text-align: center;
      background: #7460ff;
      content: attr(data-amount);
    }

    header {
      min-height: 0;
      padding: 0;
      background: transparent;

      strong {
        display: none;
      }

      span {
        display: none;
      }
    }

    dl {
      gap: 10px;
      padding: 20px 14px 16px;
    }
  }
}

.subpage-content--cash-center {
  padding-top: 0;
  padding-right: 0;
  padding-left: 0;
  background: #090f1f;
}

.cash-wallet-panel {
  padding: 14px;
  border-radius: 12px;
  background: #131828;
}

.cash-wallet-card {
  position: relative;
  z-index: 1;
  overflow: hidden;
  border-radius: 12px;
  background: #041426;
  box-shadow: inset 0 0 5px rgba(255, 255, 255, 0.25);
}

.cash-wallet-card__bg {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  object-fit: fill;
}

.cash-wallet-card__content {
  position: relative;
  padding: 16px 16px 14px;
}

.cash-wallet-balances {
  display: flex;
  align-items: center;
}

.cash-wallet-balance {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.cash-wallet-balance__label {
  display: flex;
  align-items: center;
  gap: 4px;

  img {
    width: 16px;
    height: 16px;
    object-fit: contain;

    &.is-tilted {
      transform: rotate(8.9deg);
    }
  }
}

.cash-wallet-balance__label span {
  color: #f2f2f2;
  font-size: 0.75rem;
  font-weight: 500;
  line-height: 1.2;
}

.cash-wallet-balance strong {
  display: flex;
  align-items: flex-end;
  gap: 2px;
  color: #ffffff;
  font-family: Montserrat, sans-serif;
  font-size: 1.25rem;
  font-weight: 600;
  line-height: 1;

  small {
    font-size: 0.875rem;
    font-weight: 600;
    line-height: 1;
  }

  &.is-success,
  &.is-success small {
    color: #01f185;
  }
}

.cash-wallet-divider {
  width: 1px;
  height: 32px;
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.15);
}

.cash-wallet-withdraw {
  width: 100%;
  height: 40px;
  margin-top: 16px;
  border-radius: 10px;
  color: #ffffff;
  background: linear-gradient(180deg, #51d86d 0%, #25a85a 100%);
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  box-shadow: 0 5px 14px rgba(37, 168, 90, 0.22);

  &:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }
}

.cash-wallet-note {
  display: flex;
  //min-height: 76px;
  align-items: flex-end;
  justify-content: center;
  margin-top: -12px;
  border-radius: 0 0 12px 12px;
  padding: 24px 14px 13px;
  background: rgba(255, 255, 255, 0.05);
  text-align: center;

  span {
    color: #ffffff;
    font-size: 0.75rem;
    font-weight: 500;
    line-height: 1.2;
  }
}

.cash-wallet-tabs {
  position: relative;
  height: 44px;
  flex-shrink: 0;
  background: hsl(var(--background, 224 55% 8%));
}

.cash-wallet-tabs__inner {
  position: relative;
  display: flex;
  height: 100%;
}

.cash-wallet-tabs button {
  position: relative;
  display: flex;
  min-width: 0;
  flex: 1;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 0;
  background: transparent;
  color: #ffffff;
  font-size: 0.875rem;
  font-weight: 400;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: color 300ms ease;

  &.is-active {
    color: #3d8aef;
    font-weight: 500;
  }
}

.cash-wallet-tabs__indicator {
  position: absolute;
  bottom: 0;
  width: 20%;
  height: 3px;
  border-radius: 2px;
  background: linear-gradient(96deg, #3d8aef 18%, #2f69d1 71%);
  box-shadow: #87bbff 0 2px 8px 0;
  transform: translateX(-50%);
  transition: left 300ms ease;
}

.cash-audit-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 14px 24px;
}

.cash-audit-card {
  overflow: hidden;
  border-radius: 10px;
  background: #1e2438;
}

.cash-audit-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px 10px 0 0;
  background: #364169;

  strong {
    color: #ffffff;
    font-family: Montserrat, sans-serif;
    font-size: 1.125rem;
    font-weight: 600;
    line-height: 1;
  }

  button {
    display: flex;
    min-width: 0;
    flex-shrink: 0;
    align-items: center;
    gap: 2px;
    color: #ffffff;

    span {
      font-size: 0.75rem;
      font-weight: 400;
    }

    svg {
      width: 14px;
      height: 14px;
      flex-shrink: 0;
      color: rgba(255, 255, 255, 0.5);
    }
  }
}

.cash-audit-card__body {
  padding: 10px 14px 4px;

  div {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 4px 0;
  }

  span {
    color: rgba(255, 255, 255, 0.7);
    font-size: 0.875rem;
    font-weight: 500;
  }

  strong {
    color: #ffffff;
    font-size: 0.875rem;
    font-weight: 600;
  }
}

.cash-audit-card__progress-text {
  display: flex;
  align-items: center;
  gap: 2px;

  span:last-child {
    color: #74be82;
  }

  em {
    color: #ffffff;
    font-style: normal;
  }
}

.cash-audit-card__bar-wrap {
  position: relative;
  padding: 10px 14px 18px;

  img {
    position: absolute;
    right: 6px;
    bottom: 4px;
    width: 32px;
    height: 31px;
    object-fit: contain;
    pointer-events: none;
  }
}

.cash-audit-card__bar {
  position: relative;
  border-radius: 999px;
  padding: 2px;
  background: linear-gradient(180deg, #0a1020 0%, #27314e 100%);
  box-shadow: 0 0.5px 0.5px rgba(127, 144, 192, 0.5);

  &::before {
    display: block;
    height: 12px;
    border-radius: 999px;
    background: rgba(9, 15, 30, 0.4);
    content: "";
  }
}

.cash-audit-card__bar-fill,
.cash-audit-card__bar-shine {
  position: absolute;
  left: 2px;
  border-radius: 999px;
  pointer-events: none;
}

.cash-audit-card__bar-fill {
  top: 2px;
  bottom: 2px;
  background: linear-gradient(90deg, #047857 0%, #10b981 100%);
  box-shadow: inset 0 -1px 1px rgba(34, 242, 148, 0.5);
}

.cash-audit-card__bar-shine {
  top: 3px;
  height: 5px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.2), rgba(255, 255, 255, 0.05));
}

.cash-wallet-empty {
  min-height: calc(100dvh - env(safe-area-inset-top) - 325px);
  margin-top: 0;

  .source-empty-state__body {
    padding-top: 70px;
  }
}

.subpage-content--security {
  padding-top: 5px;
}

.subpage-content--security,
.subpage-content--security *,
.subpage-content--security *::before,
.subpage-content--security *::after {
  font-size: 0.875rem !important;
}

.subpage-content--security-phone,
.subpage-content--security-password {
  font-size: 0.875rem !important;
}

.security-source-section {
  p {
    margin: 8px 0 0;
    color: #4d81e7;
    font-size: 0.75rem !important;
    font-weight: 400;
    line-height: 1.25;
  }
}

.security-source-title {
  padding: 20px 14px 8px;
  color: #ffffff;
  font-size: 0.875rem;
  font-weight: 500;
  line-height: 1.2;
}

.security-source-card {
  overflow: hidden;
  border-radius: 8px;
  background: #1e2438;
}

.security-source-row {
  position: relative;
  display: flex;
  width: 100%;
  min-height: 45px;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  color: #ffffff;
  text-align: left;
  background: transparent;

  + .security-source-row {
    border-top: 1px solid rgba(255, 255, 255, 0.06);
  }

  span {
    min-width: 0;
    flex: 0 1 auto;
    color: #ffffff;
    font-size: 0.875rem;
    font-weight: 500;
    line-height: 1.2;
  }

  strong {
    min-width: 0;
    max-width: 180px;
    margin-left: auto;
    overflow: hidden;
    color: rgba(255, 255, 255, 0.5);
    font-size: 0.875rem;
    font-weight: 500;
    line-height: 1.2;
    text-align: right;
    text-overflow: ellipsis;
    white-space: nowrap;

    &.is-primary {
      color: #397fe5;
    }
  }
}

.security-source-row__icon {
  width: 18.72px;
  height: 18.72px;
  flex: 0 0 18.72px;
  color: #ffffff;
}

.security-source-row__chevron {
  width: 16.64px;
  height: 16.64px;
  flex: 0 0 16.64px;
  color: rgba(255, 255, 255, 0.3);
}

.security-source-row__dot {
  position: absolute;
  top: 9px;
  right: 8px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ff343b;
  box-shadow: 0 0 0 3px rgba(255, 52, 59, 0.16);
}

.subpage-content--security-phone,
.subpage-content--security-password {
  min-height: calc(100vh - 55px);
  min-height: calc(100dvh - 55px);
  padding: 0 0 34px;
  background: #090f1f;
}

.security-link-tabs {
  position: relative;
  height: 44px;
  flex-shrink: 0;
  background: hsl(var(--background, 224 55% 8%));
}

.security-link-tabs__inner {
  position: relative;
  display: flex;
  height: 100%;

  button {
    position: relative;
    display: flex;
    flex: 1;
    min-width: 0;
    cursor: pointer;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border: 0;
    color: #ffffff;
    background: transparent;
    font-size: 0.875rem;
    font-weight: 400;
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: color 300ms ease;

    img {
      mix-blend-mode: luminosity;
    }

    &.is-active {
      color: #ffffff;
    }
  }
}

.security-link-tabs__indicator {
  position: absolute;
  bottom: 0;
  width: 20%;
  height: 3px;
  border-radius: 2px;
  background: linear-gradient(96deg, #3d8aef 18%, #2f69d1 71%);
  box-shadow: #87bbff 0 2px 8px 0;
  transform: translateX(-50%);
  transition: left 300ms ease;
}

.security-link-tabs__label {
  position: relative;
  display: inline-flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
}

.security-link-tabs__badge {
  position: absolute;
  top: -4px;
  right: -10px;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #ed252b;
  animation: badge-pulse 1.45s ease-in-out infinite;
}

@keyframes badge-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(237, 37, 43, 0.4);
    opacity: 1;
    transform: scale(1);
  }

  50% {
    box-shadow: 0 0 0 5px rgba(237, 37, 43, 0);
    opacity: 0.72;
    transform: scale(0.82);
  }
}

.security-auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px 14px 0;
}

.security-auth-form--password {
  padding-top: 20px;
}

.security-auth-field {
  display: block;

  > span {
    display: block;
    margin: 0 0 10px;
    color: #ffffff;
    font-size: 0.75rem;
    font-weight: 500;
    line-height: 1.25;

    i {
      margin-right: 2px;
      color: #ff545b;
      font-style: normal;
    }
  }
}

.security-phone-control,
.security-password-control,
.security-code-control input {
  border: 1px solid #2d385d;
  border-radius: 11px;
  background: #121827;
  color: #ffffff;
}

.security-phone-control,
.security-password-control {
  display: flex;
  height: 46px;
  align-items: center;
  overflow: hidden;

  input {
    min-width: 0;
    height: 100%;
    flex: 1;
    border: 0;
    outline: 0;
    color: #ffffff;
    background: transparent;
    font-size: 0.875rem;
    font-weight: 500;
  }

  input::placeholder {
    color: #536094;
  }
}

.security-phone-control {
  em {
    display: grid;
    width: 53px;
    height: 100%;
    place-items: center;
    border-right: 1px solid #2d385d;
    color: #ffffff;
    font-size: 0.875rem;
    font-style: normal;
    font-weight: 500;
  }

  input {
    padding: 0 14px;
  }

  &.is-muted {
    em,
    input {
      color: rgba(255, 255, 255, 0.55);
      -webkit-text-fill-color: rgba(255, 255, 255, 0.55);
      opacity: 1;
    }
  }
}

.security-password-control {
  padding-left: 14px;

  button {
    display: grid;
    width: 46px;
    height: 46px;
    place-items: center;
    color: #97a5ca;
  }

  svg {
    width: 22px;
    height: 22px;
    overflow: visible;

    path:first-child {
      fill: currentColor;
    }
  }
}

.security-code-control {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 96px;
  gap: 12px;

  input {
    min-width: 0;
    height: 46px;
    padding: 0 14px;
    outline: 0;
    font-size: 0.875rem;
    font-weight: 500;

    &::placeholder {
      color: #536094;
    }
  }

  button {
    width: 96px;
    flex-shrink: 0;
    height: 46px;
    border-radius: 10px;
    border: 0;
    color: #ffffff;
    background: #3d83e6;
    box-shadow: inset 0 -1px 0 rgba(0, 0, 0, 0.08);
    font-size: 0.875rem;
    font-weight: 500;

    &:disabled {
      cursor: not-allowed;
      opacity: 0.5;
    }
  }
}

.security-complete-button {
  position: relative;
  display: grid;
  height: 46px;
  place-items: center;
  margin-top: 16px;
  border-radius: 8px;
  color: #ffffff;
  background: #3d83e6;
  box-shadow: 0 10px 22px rgba(61, 131, 230, 0.24);
  font-size: 0.875rem;
  font-weight: 500;

  img {
    position: absolute;
    top: -8px;
    right: -3px;
    width: 34px;
    height: 34px;
    object-fit: contain;
    animation: security-gift-bounce 1.45s ease-in-out infinite;
  }
}

.security-rebind-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: -2px;
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 500;
  line-height: 1.2;

  button {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    border-bottom: 1px solid currentColor;
    padding-bottom: 3px;
    color: #3984ff;
    font-size: 0.75rem;
    font-weight: 500;
  }

  svg {
    width: 12px;
    height: 12px;
    fill: none;

    path {
      stroke: currentColor;
      stroke-linecap: round;
      stroke-linejoin: round;
      stroke-width: 1.7;
    }
  }
}

@keyframes security-gift-bounce {
  0%,
  100% {
    transform: translateY(0) rotate(0deg);
  }

  50% {
    transform: translateY(-3px) rotate(4deg);
  }
}

.subpage-header__notification {
  display: grid;
  width: 38px;
  height: 38px;
  place-items: center;
  justify-self: end;
  border-radius: 14px;
  color: #ffffff;
  background: #2c3658;

  svg {
    width: 24px;
    height: 24px;
    fill: currentColor;
  }
}

.subpage-header__help {
  display: grid;
  width: 25px;
  height: 25px;
  place-items: center;
  justify-self: end;
  padding: 0;
  cursor: pointer;

  img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }
}

.subpage-content--bank-account {
  padding-top: 17px;
}

.withdraw-count-card {
  display: flex;
  height: 38px;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  border-radius: 5px;
  color: rgba(255, 255, 255, 0.82);
  background: #46537f;

  span {
    min-width: 0;
    flex: 1;
    font-size: 15px;
    font-weight: 400;
    line-height: 1;
  }

  strong {
    color: #ffffff;
    font-size: 15px;
    font-weight: 400;
    line-height: 1;
  }

  i {
    display: grid;
    width: 17px;
    height: 17px;
    place-items: center;
    border-radius: 50%;
    color: #46537f;
    font-size: 12px;
    font-style: normal;
    font-weight: 800;
    line-height: 1;
    background: rgba(255, 255, 255, 0.62);
  }
}

.withdraw-method {
  margin-top: 16px;
}

.withdraw-method__title {
  display: flex;
  align-items: center;

  span {
    color: #ffffff;
    font-size: 1rem;
    font-weight: 400;
    line-height: 1.2;
  }
}

.withdraw-add-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 12px;
}

.withdraw-add-button {
  display: flex;
  width: 100%;
  height: 50px;
  align-items: center;
  border-radius: 6px;
  padding: 0 14px;
  color: rgba(255, 255, 255, 0.6);
  background: hsl(var(--f2, 226 31% 17%));
  cursor: pointer;
  text-align: left;

  span {
    font-size: 0.875rem;
    font-weight: 400;
  }

  svg {
    width: 20px;
    height: 20px;
    flex: 0 0 20px;
    margin-right: 8px;
    color: currentColor;
  }
}

.order-help-overlay {
  position: fixed;
  inset: 0;
  z-index: 110;
  display: flex;
  width: 100%;
  max-width: 430px;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  padding: 0 20px;
}

.order-help-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.62);
}

.order-help-dialog {
  position: relative;
  z-index: 1;
  width: 100%;
  overflow: hidden;
  border-radius: 12px;
  background: #1e2438;
  color: #ffffff;
  transform-origin: center center;

  header {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    padding: 14px 16px;

    span {
      color: #ffffff;
      font-size: 0.875rem;
      font-weight: 600;
      line-height: 1.2;
    }

    button {
      position: absolute;
      right: 12px;
      display: flex;
      cursor: pointer;
      align-items: center;
      justify-content: center;
      padding: 4px;
      color: rgba(255, 255, 255, 0.5);
    }

    svg {
      width: 16px;
      height: 16px;
    }
  }
}

.order-help-body {
  max-height: 55vh;
  overflow-y: auto;
  padding: 16px;
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.75rem;
  font-weight: 400;

  p {
    margin: 0;
    line-height: 1.5;
  }

  .is-strong {
    color: #ffffff;
    font-weight: 500;
  }

  > div + div {
    margin-top: 12px;
  }
}

.order-help-section-title {
  margin-bottom: 8px;
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.2;
}

.order-help-divider {
  border-bottom: 1px dashed rgba(255, 255, 255, 0.15);
}

.order-help-block {
  margin-top: 8px;
}

.order-help-nested {
  margin-top: 4px;
  margin-left: 12px;

  > div + div,
  p + p {
    margin-top: 4px;
  }
}

.order-help-dialog footer {
  padding: 12px 16px 20px;

  button {
    width: 100%;
    height: 40px;
    border-radius: 8px;
    color: #ffffff;
    background: linear-gradient(118deg, #3d8aef, #2f69d1);
    cursor: pointer;
    font-size: 0.875rem;
    font-weight: 500;
  }
}

.order-help-enter-active,
.order-help-leave-active {
  transition: opacity 180ms ease;

  .order-help-dialog {
    transition:
      opacity 180ms ease,
      transform 180ms ease;
  }
}

.order-help-enter-from,
.order-help-leave-to {
  opacity: 0;

  .order-help-dialog {
    opacity: 0;
    transform: scale(0.96);
  }
}

.picker-sheet-overlay {
  position: fixed;
  inset: 0;
  z-index: 90;
  width: 100%;
  max-width: 430px;
  margin: 0 auto;
  color: #ffffff;
}

.picker-sheet-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
}

.picker-sheet {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  overflow: hidden;
  border-radius: 10px 10px 0 0;
  background: #1e2438;
  box-shadow: 0 -12px 28px rgba(0, 0, 0, 0.32);

  header {
    position: relative;
    display: flex;
    height: 56px;
    align-items: center;
    justify-content: center;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);

    h2 {
      margin: 0;
      color: #ffffff;
      font-size: 15px;
      font-weight: 500;
      line-height: 1;
    }

    button {
      position: absolute;
      top: 50%;
      right: 12px;
      display: grid;
      width: 32px;
      height: 32px;
      place-items: center;
      color: rgba(255, 255, 255, 0.5);
      transform: translateY(-50%);

      svg {
        width: 20px;
        height: 20px;
      }
    }
  }
}

.picker-sheet__options {
  max-height: min(64vh, 560px);
  overflow-x: hidden;
  overflow-y: auto;
  padding: 8px 0 10px;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.picker-sheet__option {
  display: flex;
  width: 100%;
  min-height: 45px;
  align-items: center;
  justify-content: space-between;
  padding: 0 15px;
  color: #b5babe;
  font-size: 15px;
  font-weight: 400;
  text-align: left;

  span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &.is-active {
    color: #397fe5;
  }

  svg {
    width: 16px;
    height: 16px;
    flex: 0 0 16px;
  }
}

.picker-sheet-enter-active,
.picker-sheet-leave-active {
  transition: opacity 180ms ease;

  .picker-sheet {
    transition: transform 220ms ease;
  }
}

.picker-sheet-enter-from,
.picker-sheet-leave-to {
  opacity: 0;

  .picker-sheet {
    transform: translateY(28px);
  }
}

.withdraw-form-overlay {
  position: fixed;
  inset: 0;
  z-index: 80;
  width: 100%;
  max-width: 430px;
  margin: 0 auto;
  color: #ffffff;
}

.withdraw-form-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.78);
}

.withdraw-form-sheet {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  overflow: hidden;
  border-radius: 12px 12px 0 0;
  background: #1e2438;
  box-shadow: 0 -12px 28px rgba(0, 0, 0, 0.3);

  header {
    position: relative;
    display: flex;
    height: 56px;
    align-items: center;
    justify-content: center;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);

    h2 {
      margin: 0;
      color: #ffffff;
      font-size: 15px;
      font-weight: 500;
      line-height: 1;
    }

    button {
      position: absolute;
      top: 50%;
      right: 12px;
      display: grid;
      width: 32px;
      height: 32px;
      place-items: center;
      color: rgba(255, 255, 255, 0.5);
      transform: translateY(-50%);

      svg {
        width: 20px;
        height: 20px;
      }
    }
  }
}

.withdraw-form-sheet button,
.withdraw-form-sheet input {
  &:focus,
  &:focus-visible {
    outline: none;
    box-shadow: none;
  }
}

.withdraw-form-body {
  display: grid;
  gap: 14px;
  padding: 16px 16px calc(env(safe-area-inset-bottom) + 24px);
}

.withdraw-form-field {
  display: grid;
  gap: 7px;
  min-width: 0;

  > span {
    color: rgba(255, 255, 255, 0.88);
    font-size: 13px;
    font-weight: 400;
    line-height: 1;
  }

  i {
    margin-right: 1px;
    color: #f65a64;
    font-style: normal;
  }

  input {
    width: 100%;
    height: 45px;
    min-width: 0;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    padding: 0 14px;
    color: #ffffff;
    font-size: 15px;
    background: #1e2438;

    &::placeholder {
      color: rgba(255, 255, 255, 0.36);
    }

    &:focus,
    &:focus-visible {
      border-color: rgba(255, 255, 255, 0.15);
      outline: none;
      box-shadow: none;
    }
  }
}

.withdraw-select-button,
.withdraw-input-with-action,
.withdraw-phone-input {
  display: flex;
  height: 45px;
  align-items: center;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  color: rgba(255, 255, 255, 0.38);
  background: #1e2438;

  &:focus,
  &:focus-within,
  &:focus-visible {
    border-color: rgba(255, 255, 255, 0.15);
    outline: none;
    box-shadow: none;
  }
}

.withdraw-select-button {
  justify-content: space-between;
  padding: 0 14px;
  font-size: 15px;
  text-align: left;

  svg {
    width: 12px;
    height: 12px;
    color: rgba(255, 255, 255, 0.4);
  }
}

.wallet-type-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;

  button {
    display: flex;
    height: 45px;
    align-items: center;
    justify-content: center;
    gap: 7px;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 8px;
    color: rgba(255, 255, 255, 0.72);
    font-size: 15px;
    background: #1e2438;

    &.is-active {
      border-color: #397fe5;
      background: #2d3652;
      box-shadow: 0 0 8px rgba(57, 127, 229, 0.36);
    }
  }

  b {
    position: relative;
    display: grid;
    width: 22px;
    height: 22px;
    place-items: center;
    border-radius: 3px;
    font-size: 13px;
    line-height: 1;
    background: #ffffff;

    &::before {
      font-weight: 900;
    }

    &.is-jazzcash::before {
      color: #d82d2d;
      content: "J";
    }

    &.is-easypaisa::before {
      color: #118d46;
      content: "e";
    }
  }
}

.withdraw-input-with-action {
  overflow: hidden;

  input {
    height: 100%;
    flex: 1;
    border: 0;
    border-radius: 0;
    background: transparent;
  }

  button {
    height: 100%;
    padding: 0 14px;
    color: #397fe5;
    font-size: 14px;
  }
}

.withdraw-phone-input {
  overflow: hidden;
  background: #0c1222;

  em {
    display: flex;
    width: 54px;
    height: 100%;
    align-items: center;
    justify-content: center;
    border-right: 1px solid rgba(255, 255, 255, 0.15);
    color: #ffffff;
    font-size: 15px;
    font-style: normal;
  }

  input {
    height: 100%;
    flex: 1;
    border: 0;
    border-radius: 0;
    background: transparent;
  }
}

.withdraw-form-reminder {
  margin: 2px 0 0;
  color: #f9c56c;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.42;

  strong {
    color: #f9c56c;
  }
}

.withdraw-form-confirm {
  width: 100%;
  height: 48px;
  margin-top: 4px;
  border-radius: 7px;
  color: #ffffff;
  font-size: 17px;
  font-weight: 500;
  background: linear-gradient(118deg, #3d8aef, #2f69d1);
  box-shadow: 0 5px 12px rgba(47, 105, 209, 0.35);
}

.withdraw-form-enter-active,
.withdraw-form-leave-active {
  transition: opacity 180ms ease;

  .withdraw-form-sheet {
    transition: transform 220ms ease;
  }
}

.withdraw-form-enter-from,
.withdraw-form-leave-to {
  opacity: 0;

  .withdraw-form-sheet {
    transform: translateY(28px);
  }
}

.subpage-content--about {
  padding-top: 6px;
  padding-bottom: calc(env(safe-area-inset-bottom) + 48px);
}

.about-source-list {
  display: grid;
  gap: 26px;
}

.about-source-card {
  position: relative;
  display: grid;
  min-height: 217px;
  align-items: end;
  padding-top: 88px;

  img {
    position: absolute;
    top: 0;
    left: 50%;
    z-index: 2;
    width: 164px;
    height: 164px;
    object-fit: contain;
    transform: translateX(-50%);
  }

  div {
    position: relative;
    z-index: 1;
    min-height: 153px;
    padding: 43px 24px 21px;
    border-radius: 3px;
    text-align: center;
    background: linear-gradient(180deg, rgba(39, 48, 78, 0) 0%, #46537f 100%);
    box-shadow: 0 8px 22px rgba(30, 45, 91, 0.2);
  }

  h2 {
    margin: 0;
    color: #ffffff;
    font-size: 18px;
    font-weight: 600;
    line-height: 1.25;
  }

  p {
    margin: 18px 0 0;
    color: rgba(255, 255, 255, 0.68);
    font-size: 13px;
    font-weight: 400;
    line-height: 1.42;
  }
}

.subpage-content--support {
  display: flex;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  padding-top: 17px;
  padding-bottom: 0;
}

.support-source-hero {
  position: relative;
  min-height: 121px;
  margin: 0;
  overflow: hidden;
  padding: 18px 16px;
  border: 0;
  border-radius: 8px;
  background:
    radial-gradient(circle at 76% 42%, rgba(49, 117, 219, 0.28), transparent 42%),
    #1e2438;
  transition:
    min-height 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
    padding 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
    margin 220ms cubic-bezier(0.215, 0.61, 0.355, 1);

  button {
    position: absolute;
    top: 16px;
    right: 16px;
    z-index: 1;
    min-width: 112px;
    height: 34px;
    padding: 0 14px;
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.92);
    font-size: 13px;
    font-weight: 400;
    background: linear-gradient(118deg, #3d8aef, #2f69d1);
    box-shadow: 0 5px 12px rgba(47, 105, 209, 0.26);
    transition:
      top 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
      min-width 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
      height 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
      font-size 220ms cubic-bezier(0.215, 0.61, 0.355, 1);
  }

  strong {
    position: relative;
    z-index: 1;
    display: block;
    margin-top: 15px;
    color: #ffffff;
    font-size: 17px;
    font-weight: 500;
    line-height: 1.2;
    transition:
      margin 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
      font-size 220ms cubic-bezier(0.215, 0.61, 0.355, 1);
  }

  p {
    position: relative;
    z-index: 1;
    margin: 8px 0 0;
    color: rgba(255, 255, 255, 0.76);
    font-size: 15px;
    line-height: 1.2;
    transition:
      opacity 180ms ease,
      max-height 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
      margin 220ms cubic-bezier(0.215, 0.61, 0.355, 1);
  }
}

.support-source-hero__avatars {
  position: relative;
  z-index: 1;
  display: flex;
  padding-left: 4px;
  transition:
    top 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
    left 220ms cubic-bezier(0.215, 0.61, 0.355, 1);

  img {
    width: 21px;
    height: 21px;
    margin-left: -3px;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.25);
    object-fit: cover;
    background: #2d3652;
    transition:
      width 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
      height 220ms cubic-bezier(0.215, 0.61, 0.355, 1),
      margin 220ms cubic-bezier(0.215, 0.61, 0.355, 1);
  }
}

.support-source-tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin: 20px 0 12px;
  flex-shrink: 0;
  transition: margin 220ms cubic-bezier(0.215, 0.61, 0.355, 1);

  button {
    position: relative;
    height: 39px;
    color: rgba(255, 255, 255, 0.84);
    font-size: 15px;
    font-weight: 400;

    &.is-active {
      color: #5a9bff;

      &::after {
        position: absolute;
        right: 30px;
        bottom: 1px;
        left: 30px;
        height: 2px;
        border-radius: 999px;
        background: #397fe5;
        box-shadow: 0 0 10px #397fe5;
        content: "";
      }
    }
  }
}

.is-support-collapsed {
  .support-source-hero {
    min-height: 56px;
    padding: 12px;
  }

  .support-source-hero__avatars {
    position: absolute;
    top: 17px;
    left: 16px;
    padding-left: 0;

    img {
      width: 22px;
      height: 22px;
      margin-left: -8px;

      &:first-child {
        margin-left: 0;
      }
    }
  }

  .support-source-hero button {
    top: 15px;
    min-width: 94px;
    height: 27px;
    padding: 0 11px;
    font-size: 12px;
  }

  .support-source-hero strong {
    display: flex;
    min-height: 32px;
    align-items: center;
    overflow: hidden;
    margin: 0 105px 0 84px;
    font-size: 16px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .support-source-hero p {
    max-height: 0;
    margin-top: 0;
    opacity: 0;
  }

  .support-source-tabs {
    margin: 8px 0 10px;
  }
}

.support-panel-scroll {
  min-height: 0;
  flex: 1;
  overflow-x: hidden;
  overflow-y: auto;
  margin: 0 -14px;
  padding: 0 14px calc(env(safe-area-inset-bottom) + 26px);
  overscroll-behavior: contain;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.subpage-content--support .support-categories {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  margin-top: 0;
  overflow: visible;

  button {
    height: 35px;
    border-radius: 18px;
    border: 1px solid transparent;
    color: #ffffff;
    font-size: 12.48px;
    font-weight: 500;
    background: #2d3652;
    text-transform: capitalize;

    &.is-active {
      border-color: #397fe5;
      background: #46537f;
      box-shadow:
        inset 0 0 8px rgba(255, 255, 255, 0.16),
        0 0 8px rgba(57, 127, 229, 0.55);
    }
  }
}

.faq-list {
  display: grid;
  gap: 8px;
  margin-top: 12px;

  button {
    position: relative;
    min-height: 51px;
    overflow: hidden;
    padding: 10px 17px;
    border-radius: 7px;
    color: rgba(255, 255, 255, 0.96);
    font-size: 15px;
    font-weight: 400;
    line-height: 1.18;
    text-align: left;
    background:
      linear-gradient(rgba(10, 20, 42, 0.58), rgba(10, 20, 42, 0.58)),
      linear-gradient(180deg, #182842 0%, #15233d 100%);

    &::after {
      position: absolute;
      inset: 0;
      border-radius: inherit;
      background: linear-gradient(180deg, rgba(255, 255, 255, 0.06), transparent 45%);
      content: "";
      pointer-events: none;
    }
  }
}

.support-answer-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.5);
}

.support-answer-sheet {
  position: fixed;
  bottom: 0;
  left: 50%;
  z-index: 1001;
  display: flex;
  width: 100%;
  max-width: 430px;
  height: 66vh;
  flex-direction: column;
  overflow: hidden;
  border-radius: 16px 16px 0 0;
  background: #1e2438;
  box-shadow: 0 -18px 32px rgba(0, 0, 0, 0.28);
  transform: translateX(-50%);
}

.support-answer-sheet__header {
  display: flex;
  height: 42px;
  flex-shrink: 0;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  background:
    linear-gradient(135deg, rgba(88, 111, 176, 0.88), rgba(52, 64, 103, 0.92) 48%, rgba(37, 45, 74, 0.95)),
    radial-gradient(circle at 70% 15%, rgba(255, 255, 255, 0.18), transparent 32%);

  span {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    color: #ffffff;
    font-size: 14.56px;
    font-weight: 400;
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  button {
    display: flex;
    width: 22px;
    height: 22px;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    color: #ffffff;

    svg {
      width: 16px;
      height: 16px;
    }
  }
}

.support-answer-sheet__body {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding: 17px 16px 26px;
  color: rgba(255, 255, 255, 0.88);
  font-size: 14.56px;
  font-weight: 400;
  line-height: 1.625;
  overscroll-behavior: contain;

  p {
    margin: 0 0 10px;

    &:empty {
      display: none;
    }
  }

  strong {
    color: #ffffff;
    font-weight: 700;
  }

  ul {
    margin: 6px 0 0;
    padding-left: 20px;
  }

  li {
    margin: 0 0 4px;
    list-style: disc;

    p {
      display: inline;
      margin: 0;
    }
  }
}

.support-answer-fade-enter-active,
.support-answer-fade-leave-active {
  transition: opacity 0.2s ease;
}

.support-answer-fade-enter-from,
.support-answer-fade-leave-to {
  opacity: 0;
}

.support-answer-sheet-enter-active,
.support-answer-sheet-leave-active {
  transition: transform 0.3s ease;
}

.support-answer-sheet-enter-from,
.support-answer-sheet-leave-to {
  transform: translate(-50%, 100%);
}

.support-notice-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 12px 24px;

  button {
    display: flex;
    min-height: 74px;
    flex-direction: column;
    align-items: stretch;
    justify-content: center;
    gap: 8px;
    border-radius: 14px;
    padding: 12px 14px;
    text-align: left;
    color: #ffffff;
    background: rgba(123, 66, 221, 0.72);
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);

    &.is-unread {
      background: linear-gradient(135deg, rgba(126, 74, 240, 0.96), rgba(86, 37, 190, 0.92));
    }
  }

  time {
    color: rgba(255, 255, 255, 0.58);
    font-size: 12px;
    font-weight: 500;
  }
}

.support-notice-list__title {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
  color: #ffffff;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.25;

  i {
    width: 7px;
    height: 7px;
    flex: 0 0 7px;
    border-radius: 999px;
    background: #ed2c35;
    box-shadow: 0 0 10px rgba(237, 44, 53, 0.9);
  }
}

.support-empty-panel {
  display: flex;
  min-height: 390px;
  align-items: center;
  flex-direction: column;
  padding-top: 55px;
  color: #97a5ca;
  text-align: center;

  svg {
    width: 100px;
    height: 100px;
    color: #97a5ca;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.1;
    stroke-linecap: round;
    stroke-linejoin: round;

    circle {
      fill: #090f1f;
    }
  }

  p {
    margin: 17px 0 0;
    color: #97a5ca;
    font-size: 17px;
    font-weight: 600;
    line-height: 1.2;
  }
}

.subpage-content--user-agreement {
  padding: 25px 20px calc(env(safe-area-inset-bottom) + 28px);
}

.source-legal-page {
  margin: 0;
  border: 0;
  border-radius: 0;
  color: rgba(255, 255, 255, 0.82);
  background: transparent;
  box-shadow: none;

  p,
  h2,
  h3 {
    margin: 0;
    font-size: 15px;
    font-weight: 400;
    line-height: 1.62;
  }

  p + p {
    margin-top: 3px;
  }

  .is-lead,
  .is-strong {
    color: rgba(255, 255, 255, 0.9);
    font-weight: 700;
  }

  h2 {
    margin-top: 9px;
    padding-top: 5px;
    border-top: 1px solid rgba(255, 255, 255, 0.72);
    color: rgba(255, 255, 255, 0.82);
  }

  h3 {
    margin-top: 5px;
    color: rgba(255, 255, 255, 0.9);
    font-weight: 500;
  }
}

.subpage-content--settings {
  padding-top: 17px;
}

.settings-source-card {
  overflow: hidden;
  border-radius: 8px;
  background: #1e2438;
}

.settings-source-row {
  display: flex;
  width: 100%;
  height: 48px;
  align-items: center;
  gap: 10px;
  padding: 0 15px;
  color: #ffffff;
  text-align: left;

  + .settings-source-row {
    border-top: 1px solid rgba(255, 255, 255, 0.06);
  }

  > span {
    display: flex;
    min-width: 0;
    flex: 1;
    align-items: center;
    gap: 10px;
    font-size: 14.56px;
    font-weight: 500;
  }

  strong {
    color: #397fe5;
    font-size: 12.48px;
    font-weight: 500;
  }

  em {
    color: rgba(255, 255, 255, 0.72);
    font-size: 17px;
    font-style: normal;
    line-height: 1;
  }
}

.settings-source-row__icon {
  width: 18px;
  height: 18px;
  flex: 0 0 18px;
  object-fit: contain;
}

.settings-switch {
  position: relative;
  display: block;
  width: 33.28px;
  height: 20.8px;
  flex: 0 0 33.28px;
  border-radius: 999px;
  background: #2d3652;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);
  transition: background 180ms ease;

  i {
    position: absolute;
    top: 3.12px;
    left: 3.12px;
    display: block;
    width: 14.56px;
    height: 14.56px;
    border-radius: 50%;
    background: #ffffff;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.24);
    transition: transform 180ms ease;
  }

  &.is-on {
    background: #397fe5;

    i {
      transform: translateX(12.48px);
    }
  }
}

.settings-logout {
  display: flex;
  width: 100%;
  height: 40px;
  align-items: center;
  justify-content: center;
  gap: 5px;
  margin-top: 17px;
  border-radius: 7px;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
  background: linear-gradient(118deg, #3d8aef, #2f69d1);
  box-shadow: 0 5px 12px rgba(47, 105, 209, 0.3);

  svg {
    width: 18px;
    height: 18px;
  }
}
</style>
