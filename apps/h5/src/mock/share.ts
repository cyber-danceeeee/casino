import shareBackground from '../assets/images/share_bg.webp'
import taskBadge from '../assets/images/share-badge.svg'
import treasure from '../assets/images/treasure_box.webp'
import smallChestClosed from '../assets/images/small_chest_closed.webp'
import smallChestOpen from '../assets/images/small_chest_open.webp'
import largeChestClosed from '../assets/images/large_chest_closed.webp'
import ribbon from '../assets/images/ribbon_banner.webp'
import rankingStage from '../assets/images/ranking-stage.svg'
import agentRule from '../assets/images/agent-rule.webp'
import agentCardTitle from '../assets/images/agent/card-title-bg.webp'
import agentLevelChart from '../assets/images/agent/level-chart.webp'
import agentStar from '../assets/images/agent/star.webp'
import customer from '../assets/images/profile/customer.webp'
import agentAvatar from '../assets/images/profile/1(1).webp'
import inviteSpin from '../assets/images/share-invite-spin.webp'
import inviteChest from '../assets/images/share-invite-chest.webp'
import referralsRewards from '../assets/images/share-referrals-rewards.webp'
import facebook from '../assets/images/dialogs/social-facebook.webp'
import messenger from '../assets/images/dialogs/social-messenger.webp'
import whatsapp from '../assets/images/dialogs/social-whatsapp.webp'
import telegram from '../assets/images/dialogs/social-telegram.webp'
import instagram from '../assets/images/dialogs/social-instagram.webp'
import tiktok from '../assets/images/dialogs/social-tiktok.webp'
import xSocial from '../assets/images/dialogs/social-x.webp'
import avatar007 from '../assets/images/profile/default-avatar-007.webp'
import avatar008 from '../assets/images/profile/default-avatar-008.webp'
import avatar011 from '../assets/images/profile/default-avatar-011.webp'
import avatar013 from '../assets/images/profile/default-avatar-013.webp'
import avatar015 from '../assets/images/profile/default-avatar-015.webp'
import avatar016 from '../assets/images/profile/default-avatar-016.webp'
import avatar018 from '../assets/images/profile/default-avatar-018.webp'
import avatar019 from '../assets/images/profile/default-avatar-019.webp'
import avatar020 from '../assets/images/profile/default-avatar-020.webp'
import rankingTop1Badge from '../assets/images/ranking-top1-badge.webp'
import rankingTop2Badge from '../assets/images/ranking-top2-badge.webp'
import rankingTop3Badge from '../assets/images/ranking-top3-badge.webp'

export type ShareTab = 'share' | 'ranking' | 'agent'
export type AgentTab = 'overview' | 'referrals' | 'commission'

export interface ShareAction {
  label: string
  tone: 'violet' | 'blue' | 'rose'
}

export interface SocialLink {
  id?: string
  label: string
  icon: string
  image?: string
  url?: string
  shareURL?: string
  shareTitle?: string
  shareText?: string
  shareDescription?: string
}

export interface ShareHomeEntry {
  id: string
  title: string
  image: string
  background: string
}

export interface ChestStat {
  label: string
  value: string
  navigable?: boolean
}

export interface RankingUser {
  rank: number
  name: string
  total: string
  avatar: string
  badge?: string
}

export interface AgentCommissionRule {
  referrals: string
  validBets: string
  commission: string
}

export const shareTabs: Array<{ key: ShareTab; label: string }> = [
  { key: 'share', label: 'Share' },
  { key: 'ranking', label: 'Ranking' },
  { key: 'agent', label: 'Agent' },
]

export const agentTabs: Array<{ key: AgentTab; label: string }> = [
  { key: 'overview', label: 'Overview' },
  { key: 'referrals', label: 'Referrals' },
  { key: 'commission', label: 'Comm.' },
]

export const shareAssets = {
  background: shareBackground,
  badge: taskBadge,
  treasure,
  smallChestClosed,
  smallChestOpen,
  largeChestClosed,
  ribbon,
  rankingStage,
  agentRule,
  agentCardTitle,
  agentLevelChart,
  agentStar,
  customer,
  agentAvatar,
}

export const shareActions: ShareAction[] = [
  { label: 'Generate Poster', tone: 'violet' },
  { label: 'Copy Link', tone: 'blue' },
  { label: 'Share Out', tone: 'rose' },
]

export const socialLinks: SocialLink[] = [
  { label: 'Facebook', icon: facebook },
  { label: 'Messenger', icon: messenger },
  { label: 'WhatsApp', icon: whatsapp },
  { label: 'Telegram', icon: telegram },
  { label: 'Instagram', icon: instagram },
  { label: 'TikTok', icon: tiktok },
  { label: 'X', icon: xSocial },
]

export const shareHomeEntries: ShareHomeEntry[] = [
  {
    id: 'invite-spin',
    title: 'Invite Spin',
    image: inviteSpin,
    background:
      'linear-gradient(232.408deg, rgb(249, 239, 165) 0%, rgb(250, 171, 200) 24.04%, rgb(232, 123, 249) 44.86%, rgb(167, 62, 242) 69.07%, rgb(79, 22, 222) 95.9%)',
  },
  {
    id: 'invite-chest',
    title: 'Invite Chest',
    image: inviteChest,
    background:
      'linear-gradient(232.408deg, rgb(250, 185, 255) 0%, rgb(205, 174, 255) 22.42%, rgb(141, 100, 254) 46.66%, rgb(103, 68, 230) 74.28%, rgb(36, 46, 148) 100%)',
  },
  {
    id: 'referrals-rewards',
    title: 'Referrals Rewards',
    image: referralsRewards,
    background:
      'linear-gradient(232.408deg, rgb(246, 167, 141) 0%, rgb(249, 179, 113) 11.88%, rgb(250, 169, 38) 47.09%, rgb(240, 132, 45) 68.69%, rgb(237, 66, 45) 100%)',
  },
]

export const chestStats: ChestStat[] = [
  { label: 'Get', value: '0', navigable: true },
  { label: 'Opened', value: '0' },
  { label: 'Earned', value: '0.00', navigable: true },
]

export const calendarDays = Array.from({ length: 30 }, (_, index) => index + 1)

export const chestRewards = [
  'Rs 1,500.00',
  'Rs 3,000.00',
  'Rs 9,000.00',
  'Rs 15,000.00',
  'Rs 30,000.00',
  'Rs 90,000.00',
  'Rs 150,000.00',
  'Rs 300,000.00',
]

export const rankingDate = '2026-07-05 (Yesterday)'

export const rankingTopUsers: RankingUser[] = [
  {
    rank: 2,
    name: 'Q******9',
    total: '11,773.49',
    avatar: avatar011,
    badge: rankingTop2Badge,
  },
  {
    rank: 1,
    name: 'R******8',
    total: '11,944.89',
    avatar: avatar016,
    badge: rankingTop1Badge,
  },
  {
    rank: 3,
    name: 'U******7',
    total: '11,248.34',
    avatar: avatar013,
    badge: rankingTop3Badge,
  },
]

export const rankingRows: RankingUser[] = [
  { rank: 4, name: 'U******4', total: '11,208.25', avatar: avatar018 },
  { rank: 5, name: 'H******4', total: '10,400.35', avatar: avatar015 },
  { rank: 6, name: 'E******5', total: '10,043.52', avatar: avatar013 },
  { rank: 7, name: 'Y******4', total: '10,009.47', avatar: avatar020 },
  { rank: 8, name: 'M******0', total: '9,553.38', avatar: avatar019 },
  { rank: 9, name: 'C******4', total: '8,723.95', avatar: avatar013 },
  { rank: 10, name: 'Y******0', total: '8,368.38', avatar: avatar007 },
  { rank: 11, name: 'E******3', total: '8,310.58', avatar: avatar011 },
  { rank: 12, name: 'O******5', total: '8,125.94', avatar: avatar016 },
  { rank: 13, name: 'F******3', total: '8,098.43', avatar: avatar020 },
  { rank: 14, name: 'L******6', total: '7,941.55', avatar: avatar008 },
]

export const selfRanking = {
  rank: '---',
  earnings: '--',
}

export const agentCommissionRules: AgentCommissionRule[] = [
  { referrals: '>=1', validBets: '>=0.00', commission: '0.70%+15.00%' },
  { referrals: '>=3', validBets: '>=10,000.00', commission: '0.71%+15.00%' },
  { referrals: '>=8', validBets: '>=100,000.00', commission: '0.73%+15.00%' },
  { referrals: '>=15', validBets: '>=3,000,000.00', commission: '0.75%+15.00%' },
  { referrals: '>=21', validBets: '>=5,000,000.00', commission: '0.77%+15.00%' },
  { referrals: '>=55', validBets: '>=10,000,000.00', commission: '0.80%+15.00%' },
  { referrals: '>=89', validBets: '>=30,000,000.00', commission: '0.90%+15.00%' },
  { referrals: '>=110', validBets: '>=50,000,000.00', commission: '1.00%+15.00%' },
  { referrals: '>=144', validBets: '>=70,000,000.00', commission: '1.15%+15.00%' },
  { referrals: '>=377', validBets: '>=150,000,000.00', commission: '1.50%+15.00%' },
  { referrals: '>=777', validBets: '>=300,000,000.00', commission: '2.00%+15.00%' },
]
