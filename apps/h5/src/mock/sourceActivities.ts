const sourceImage = (path: string) => {
  return `/static/h5/images/source-activity/${path.replace(/\//g, '__')}`
}

export type SourceActivityVariant =
  | 'first-deposit'
  | 'deposit-bonus'
  | 'daily-deposit'
  | 'periodic-green'
  | 'periodic-gold'
  | 'wheel'
  | 'chest'
  | 'sign-in'
  | 'red-packet'
  | 'relief'
  | 'welfare'
  | 'sub-deposit'
  | 'ranking'
  | 'security'
  | 'app'

export interface SourceActivityRow {
  label: string
  value: string
  extra: string
}

export interface SourceActivityConfig {
  slug: string
  aliases?: string[]
  variant: SourceActivityVariant
  title: string
  subtitle: string
  badge: string
  cta: string
  ctaType: 'deposit' | 'claim' | 'share' | 'download' | 'verify' | 'rules'
  ruleTitle: string
  theme: 'red' | 'blue' | 'green' | 'gold' | 'purple'
  images: Record<string, string>
  rows: SourceActivityRow[]
  rules: string[]
}

const depositRows: SourceActivityRow[] = [
  { label: '100.00-299.99', value: '+30.00', extra: '+20.00' },
  { label: '300.00-499.99', value: '+80.00', extra: '+60.00' },
  { label: '500.00-999.99', value: '+150.00', extra: '+100.00' },
  { label: '1,000.00-2,999.99', value: '+300.00', extra: '+200.00' },
  { label: '3,000.00-4,999.99', value: '+900.00', extra: '+600.00' },
  { label: '5,000.00-9,999.99', value: '+1,750.00', extra: '+750.00' },
  { label: '10,000.00-29,999.99', value: '+3,500.00', extra: '+1,500.00' },
  { label: '30,000.00-49,999.99', value: '+12,000.00', extra: '+3,000.00' },
  { label: '50,000.00+', value: '+20,000.00', extra: '+5,000.00' },
]

export const sourceActivityAssets = {
  tapHand: sourceImage('activity/shared/tap_hand.webp'),
  cash: sourceImage('bonus/icon_cash.webp'),
  cashBonus: sourceImage('bonus/icon_cash_bonus.webp'),
  bonus: sourceImage('bonus/bonus_icon.webp'),
  bonusIcon: sourceImage('bonus/icon_bonus.webp'),
  coinStack: sourceImage('bonus/coin_stack.webp'),
  rays: sourceImage('bonus/rays.webp'),
  sunburst: sourceImage('bonus/sunburst.webp'),
  lightBurst: sourceImage('bonus/light_burst.webp'),
  rewardShine: sourceImage('bonus/reward_shine.webp'),
}

export const sourceActivities: SourceActivityConfig[] = [
  {
    slug: 'first-deposit-bonus',
    aliases: ['first-deposit'],
    variant: 'first-deposit',
    title: 'First Deposit Super Rewards',
    subtitle: 'Get it on your first deposit',
    badge: 'Maximum Reward: +25,000.00',
    cta: 'Deposit Now',
    ctaType: 'deposit',
    ruleTitle: 'Reward Details',
    theme: 'red',
    images: {
      pageBg: sourceImage('activity/first-deposit-bonus/page_bgs.webp'),
      bundle: sourceImage('activity/first-deposit-bonus/first_deposit_bundle.webp'),
      heroBox: sourceImage('activity/first-deposit-bonus/hero_box.webp'),
      giftBox: sourceImage('activity/first-deposit-bonus/gift_box.webp'),
      treasure: sourceImage('activity/first-deposit-bonus/treasure.webp'),
      coin: sourceImage('activity/first-deposit-bonus/holographic_coin.webp'),
      banner: sourceImage('activity/first-deposit-bonus/treasure_banner.webp'),
      frame: sourceImage('activity/first-deposit-bonus/reward_card_frame.webp'),
      desc: sourceImage('activity/first-deposit-bonus/bonus_description.webp'),
      rulesBanner: sourceImage('activity/first-deposit-bonus/rules_banner.webp'),
      lightCash: sourceImage('activity/first-deposit-bonus/light_rays_cash.webp'),
      lightBonus: sourceImage('activity/first-deposit-bonus/light_rays_bonus.webp'),
    },
    rows: depositRows,
    rules: [
      'This promotion is available when you make a deposit using the <strong>First Deposit Rewards Coupon</strong>. The coupon can only be used once.',
      'After using the coupon, you can receive both <strong>Bonus</strong> and <strong>Cash</strong> rewards based on your deposit amount.',
      'Bonus rewards will be credited to your <strong>Bonus Wallet</strong> and can be unlocked step by step through betting activity.',
      'Cash rewards will be credited to your <strong>Cash Wallet</strong> and require <strong>3x valid bets</strong> before withdrawal.',
      'Each device can receive this reward for only 1 account.',
      'Phone number verification is required before claiming the reward.',
      'Rewards may be cancelled in cases of duplicate accounts, abnormal activity, or any violation of promotion rules.',
    ],
  },
  {
    slug: 'deposit-bonus',
    aliases: ['recharge-reward'],
    variant: 'deposit-bonus',
    title: 'Recharge Rewards',
    subtitle: 'Get rewards on every deposit',
    badge: 'Maximum Reward Single Reward',
    cta: 'Deposit Now',
    ctaType: 'deposit',
    ruleTitle: 'Recharge Details',
    theme: 'red',
    images: {
      pageBg: sourceImage('activity/deposit-bonus/content_bg.webp'),
      heroBg: sourceImage('activity/deposit-bonus/hero_bg.webp'),
      light: sourceImage('bonus/light_burst.webp'),
      pile: sourceImage('activity/deposit-bonus/coin_pile.webp'),
      chest: sourceImage('activity/deposit-bonus/treasure_chest.webp'),
      coin: sourceImage('activity/deposit-bonus/gold_coin.webp'),
      contentBg: sourceImage('activity/deposit-bonus/content_bg.webp'),
      bundle: sourceImage('activity/deposit-bonus/daily_deposit_bundle.webp'),
      rulesBanner: sourceImage('activity/deposit-bonus/rules_banner.webp'),
    },
    rows: [
      { label: 'Single Reward', value: 'Bonus Rewards', extra: '10,000.00' },
      { label: 'Claim Rewards', value: 'Cash Rewards Claimed: 0', extra: 'Bonus Rewards Claimed: 0' },
      { label: '100.00-499.99', value: '+18.00', extra: '+8%' },
      { label: '500.00-999.99', value: '+58.00', extra: '+10%' },
      { label: '1,000.00-4,999.99', value: '+128.00', extra: '+12%' },
      { label: '5,000.00+', value: '+500.00', extra: '+15%' },
    ],
    rules: [
      'This promotion is available when you make a qualifying deposit during the event.',
      'After depositing, you can receive a random Bonus reward based on your deposit tier.',
      'The actual Bonus amount will be randomly issued within the reward range shown on the promotion page.',
      'Bonus rewards will be credited to your Bonus Wallet and can be unlocked step by step through betting activity.',
      'Each device can receive this reward for only 1 account.',
      'Phone number verification is required before claiming the reward.',
      'Rewards may be cancelled in cases of duplicate accounts, abnormal activity, or any violation of promotion rules.',
    ],
  },
  {
    slug: 'daily-deposit-cashback',
    aliases: ['daily-deposit'],
    variant: 'daily-deposit',
    title: 'DEPOSIT REBATE EVERYDAY！',
    subtitle: 'The more you deposit, the more you earn — no requirements!',
    badge: 'Increase Rebate',
    cta: 'Increase Rebate Rate',
    ctaType: 'deposit',
    ruleTitle: 'Event Rules',
    theme: 'green',
    images: {
      pageBg: sourceImage('activity/daily-deposit-cashback/page_bg.webp'),
      heroBg: sourceImage('activity/daily-deposit-cashback/hero_bg.webp'),
      main: sourceImage('activity/daily-deposit-cashback/hero_main.webp'),
      button: sourceImage('activity/daily-deposit-cashback/increase_rebate_btn.webp'),
      left: sourceImage('activity/daily-deposit-cashback/card_corner_left.webp'),
      right: sourceImage('activity/daily-deposit-cashback/card_corner_right.webp'),
      bundle: sourceImage('activity/daily-deposit-cashback/hero_main.webp'),
      rulesBanner: sourceImage('activity/daily-deposit-cashback/rules_bundle.webp'),
    },
    rows: [
      { label: '3,000.00+', value: '0.00%', extra: '0.50%' },
      { label: '5,000.00+', value: '0.00%', extra: '1.00%' },
      { label: '10,000.00+', value: '0.00%', extra: '1.50%' },
      { label: '30,000.00+', value: '0.00%', extra: '2.00%' },
      { label: '50,000.00+', value: '0.00%', extra: '2.50%' },
      { label: '100,000.00+', value: '0.00%', extra: '3.00%' },
      { label: '300,000.00+', value: '0.00%', extra: '3.50%' },
      { label: '500,000.00+', value: '0.00%', extra: '4.00%' },
    ],
    rules: [
      'This promotion is available for qualifying daily deposits during the event.',
      'You can receive a Bonus rebate based on your total successful deposit amount for the day.',
      'The higher your daily deposit amount, the higher your Bonus rebate rate.',
      'Bonus rewards will be credited to your Bonus Wallet and can be unlocked step by step through betting activity.',
      'This reward is available only in the 56MVP APP. Please claim it from the APP to enjoy the benefit.',
      'Each device can receive this reward for only 1 account.',
      'Rewards may be cancelled in cases of duplicate accounts, abnormal activity, or any violation of promotion rules.',
    ],
  },
  {
    slug: 'periodic-deposit-reward/weekly',
    aliases: ['weekly-deposit'],
    variant: 'periodic-green',
    title: 'Weekly Deposit Bonus',
    subtitle: 'Complete weekly cycles and unlock gifts',
    badge: 'Weekly Reward',
    cta: 'Deposit Now',
    ctaType: 'deposit',
    ruleTitle: 'Weekly Details',
    theme: 'green',
    images: {
      pageBg: sourceImage('activity/periodic-deposit-reward/page_bg_green.webp'),
      heroBg: sourceImage('activity/periodic-deposit-reward/hero_bg_green.webp'),
      main: sourceImage('activity/periodic-deposit-reward/hero_main_green.webp'),
      claimGift: sourceImage('activity/periodic-deposit-reward/claim_gift.webp'),
    },
    rows: [
      { label: 'Rs 500.00 - Rs 999.99', value: '1st Rs5.00 / 2nd Rs5.00', extra: '3rd Rs10.00' },
      { label: 'Rs 1,000.00 - Rs 2,999.99', value: '1st Rs10.00 / 2nd Rs10.00', extra: '3rd Rs20.00' },
      { label: 'Rs 3,000.00 - Rs 4,999.99', value: '1st Rs20.00 / 2nd Rs30.00', extra: '3rd Rs50.00' },
      { label: 'Rs 5,000.00+', value: '1st Rs30.00 / 2nd Rs50.00', extra: '3rd Rs70.00' },
    ],
    rules: [
      'The weekly cycle follows the platform settlement calendar.',
      'Only successful deposits within the weekly cycle are counted.',
      'Rewards must complete valid-bet requirements before withdrawal.',
    ],
  },
  {
    slug: 'periodic-deposit-reward/monthly',
    aliases: ['monthly-deposit'],
    variant: 'periodic-gold',
    title: 'Monthly Deposit Bonus',
    subtitle: 'Higher cycle, higher reward',
    badge: 'Monthly Reward',
    cta: 'Deposit Now',
    ctaType: 'deposit',
    ruleTitle: 'Monthly Details',
    theme: 'gold',
    images: {
      pageBg: sourceImage('activity/periodic-deposit-reward/page_bg_gold.webp'),
      heroBg: sourceImage('activity/periodic-deposit-reward/hero_bg_gold.webp'),
      main: sourceImage('activity/periodic-deposit-reward/hero_main_gold.webp'),
    },
    rows: [
      { label: 'Rs 2,000.00+', value: 'Rs 188.00', extra: 'Monthly' },
      { label: 'Rs 8,000.00+', value: 'Rs 588.00', extra: 'Monthly' },
      { label: 'Rs 20,000.00+', value: 'Rs 1,288.00', extra: 'Monthly' },
      { label: 'Rs 50,000.00+', value: 'Rs 3,888.00', extra: 'Monthly' },
    ],
    rules: [
      'Monthly rewards are settled after the activity period ends.',
      'The same deposit record cannot be counted in multiple monthly tiers.',
      'The final reward is based on platform review and activity status.',
    ],
  },
  {
    slug: 'invitation-wheel',
    variant: 'wheel',
    title: 'Round 1',
    subtitle: 'Rs 1,500.00',
    badge: 'Notice',
    cta: 'Start Withdraw',
    ctaType: 'share',
    ruleTitle: 'Event Rules',
    theme: 'purple',
    images: {
      pageBg: sourceImage('activity/invitation-wheel/top_decor.webp'),
      top: sourceImage('activity/invitation-wheel/top_decor.webp'),
      wheel: sourceImage('activity/invitation-wheel/wheel_5.webp'),
      ring: sourceImage('activity/invitation-wheel/wheel_13.webp'),
      pointer: sourceImage('activity/invitation-wheel/wheel_16.webp'),
      pedestal: sourceImage('activity/invitation-wheel/spin_pedestal.svg'),
      luckyPointer: sourceImage('activity/invitation-wheel/wheel_6.webp'),
      progressCheck: sourceImage('activity/invitation-wheel/progress_check.svg'),
      progressCoin: sourceImage('activity/invitation-wheel/wheel_4.webp'),
      glowingSun: sourceImage('activity/invitation-wheel/glowing-sun.webp'),
      glowingStar: sourceImage('activity/invitation-wheel/glowing-star.webp'),
      rulesBanner: sourceImage('activity/invitation-wheel/rules_banner.webp'),
      bundle: sourceImage('activity/invitation-wheel/rules_bundle.webp'),
      prize1: sourceImage('activity/invitation-wheel/prize_1.webp'),
      prize2: sourceImage('activity/invitation-wheel/prize_2.webp'),
      prize3: sourceImage('activity/invitation-wheel/prize_3.webp'),
      prize4: sourceImage('activity/invitation-wheel/prize_4.webp'),
      prize5: sourceImage('activity/invitation-wheel/prize_5.webp'),
      prize6: sourceImage('activity/invitation-wheel/prize_6.webp'),
      prize7: sourceImage('activity/invitation-wheel/prize_7.webp'),
      prize8: sourceImage('activity/invitation-wheel/prize_8.webp'),
    },
    rows: [
      { label: '1 Valid Invite', value: '1 Spin', extra: 'Random prize' },
      { label: '3 Valid Invites', value: '3 Spins', extra: 'Bonus chance' },
      { label: '5 Valid Invites', value: '6 Spins', extra: 'Chest chance' },
      { label: '10 Valid Invites', value: '12 Spins', extra: 'Big reward' },
    ],
    rules: [
      'This promotion is available during the event period.',
      'You can get spin chances by joining the event, sharing the event, or inviting new members.',
      'Each valid invited member can help you get more spin chances and speed up your reward progress.',
      'Complete the required spin progress to unlock the Cash reward.',
      'Cash rewards will be credited to your <strong>Cash Wallet</strong> and require <strong>3x valid bets</strong> before withdrawal.',
      'You can participate again each time you meet the promotion requirements.',
      'Each device can receive this reward for only 1 account.',
      'Rewards may be reviewed before being issued.',
      'Rewards may be cancelled in cases of duplicate accounts, abnormal activity, or any violation of promotion rules.',
    ],
  },
  {
    slug: 'invitation-chest',
    variant: 'chest',
    title: 'Invite Chest',
    subtitle: 'Open treasure with your team',
    badge: 'Invite Chest',
    cta: 'Invite Now',
    ctaType: 'share',
    ruleTitle: 'Chest Rules',
    theme: 'purple',
    images: {
      stage: sourceImage('activity/invitation-chest/stage_bg.webp'),
      hostess: sourceImage('activity/invitation-chest/hostess.webp'),
      treasure: sourceImage('activity/invitation-chest/treasure_box.webp'),
      blueChest: sourceImage('activity/invitation-chest/blue_chest.webp'),
      titleFrame: sourceImage('activity/invitation-chest/title_frame.webp'),
      tunnel: sourceImage('activity/invitation-chest/tech_tunnel_bg.webp'),
      bundle: sourceImage('activity/invitation-chest/chest_bundle.webp'),
      rulesBanner: sourceImage('activity/invitation-chest/rules_banner.webp'),
    },
    rows: [
      { label: 'Share to Get Chests', value: 'Rs 100.00 Chest', extra: 'Your friends can receive Rs 10.00 rewards' },
      { label: 'How to Get?', value: 'Invite 1 registrations', extra: 'Get 0 / Opened 0 / Earned 0.00' },
      { label: 'How to Open?', value: 'Deposit >= Rs 500.00', extra: 'Valid bets >= Rs 3,000.00' },
      { label: 'Chest Progress', value: '1 2 3 4 5 6 7 8 9 10', extra: 'Keep inviting to open more chests' },
    ],
    rules: [
      'Chest progress increases after valid invited users are confirmed.',
      'Rewards can be claimed after the chest opens.',
      'Duplicate or risk-controlled invitations do not count.',
    ],
  },
  {
    slug: 'daily-sign-in',
    aliases: ['sign-in'],
    variant: 'sign-in',
    title: 'Daily Check-in Rewards',
    subtitle: 'Log in daily, boost your rewards',
    badge: '7 Day Rewards',
    cta: 'Check In',
    ctaType: 'claim',
    ruleTitle: 'Check-in Rules',
    theme: 'blue',
    images: {
      viewport: sourceImage('activity/daily-sign-in/viewport_bg.webp'),
      treasure: sourceImage('activity/daily-sign-in/treasure.webp'),
      carLeft: sourceImage('activity/daily-sign-in/car_left.webp'),
      carRight: sourceImage('activity/daily-sign-in/car_right.webp'),
      lit: sourceImage('activity/daily-sign-in/check_badge_lit.webp'),
      dim: sourceImage('activity/daily-sign-in/check_badge_dim.webp'),
      bell: sourceImage('activity/daily-sign-in/bell.webp'),
    },
    rows: [
      { label: 'Day 1', value: 'Rs 1.00', extra: 'Bonus' },
      { label: 'Day 2', value: 'Rs 2.00', extra: 'Bonus' },
      { label: 'Day 3', value: 'Rs 5.00', extra: 'Bonus' },
      { label: 'Day 7', value: 'Rs 30.00', extra: 'Cash' },
    ],
    rules: [
      'Each account can check in once per day.',
      'Continuous sign-in rewards increase with the streak.',
      'Missed days may reset the consecutive sign-in progress.',
    ],
  },
  {
    slug: 'red-packet-rain',
    variant: 'red-packet',
    title: 'Red Packet Rain',
    subtitle: 'Catch packets while they fall',
    badge: 'Lucky Cash Rain',
    cta: 'Start Catching',
    ctaType: 'claim',
    ruleTitle: 'Packet Rules',
    theme: 'red',
    images: {
      heroBg: sourceImage('activity/red-packet-rain/hero_bg.webp'),
      packet: sourceImage('activity/red-packet-rain/hero_packet.webp'),
      packets: sourceImage('activity/red-packet-rain/decor_packets.webp'),
      confetti: sourceImage('activity/red-packet-rain/decor_confetti.webp'),
      falling: sourceImage('activity/home-entries/red_packet_rain_falling.webp'),
      stage: sourceImage('activity/red-packet-rain/stage_bg.webp'),
      subtitle: sourceImage('activity/red-packet-rain/subtitle_banner.webp'),
      countdown: sourceImage('activity/red-packet-rain/countdown_ribbon.webp'),
      rewardCash: sourceImage('activity/red-packet-rain/reward_cash.webp'),
      rulesPackets: sourceImage('activity/red-packet-rain/rules_packets.webp'),
      rulesSparkle: sourceImage('activity/red-packet-rain/rules_sparkle.webp'),
      bannerGlow: sourceImage('activity/red-packet-rain/banner_glow.webp'),
      cellBonus: sourceImage('activity/red-packet-rain/cell_bonus.webp'),
      halo: sourceImage('activity/red-packet-rain/icon_halo.webp'),
      haloSmall: sourceImage('activity/red-packet-rain/icon_halo_sm.webp'),
    },
    rows: [
      { label: 'Small Packet', value: 'Random', extra: 'Cash' },
      { label: 'Lucky Packet', value: 'Random', extra: 'Bonus' },
      { label: 'Mega Packet', value: 'Up to 77.00', extra: 'Cash' },
      { label: 'Round Time', value: 'Limited', extra: 'Timer' },
    ],
    rules: [
      'Tap red packets during the active rain round.',
      'Each packet reward is random and subject to activity status.',
      'Rewards are issued after the round ends and the claim is confirmed.',
    ],
  },
  {
    slug: 'daily-relief',
    variant: 'relief',
    title: 'Daily Rescue Cash',
    subtitle: 'Get rescued after difficult rounds',
    badge: 'Relief Bonus',
    cta: 'Claim Rescue',
    ctaType: 'claim',
    ruleTitle: 'Rescue Rules',
    theme: 'green',
    images: {
      pageBg: sourceImage('activity/daily-relief/page_bg.webp'),
      top: sourceImage('activity/daily-relief/hero_top.webp'),
      main: sourceImage('activity/daily-relief/hero_main.webp'),
      coinBg: sourceImage('activity/daily-relief/coin_bg.webp'),
      ribbon: sourceImage('activity/daily-relief/ribbon_banner.webp'),
      bundle: sourceImage('activity/daily-relief/rules_bundle.webp'),
    },
    rows: [
      { label: '100.00+', value: '0.50%', extra: 'Valid bets' },
      { label: '1,000.00+', value: '1.00%', extra: 'Valid bets' },
      { label: '5,000.00+', value: '1.50%', extra: 'Valid bets' },
      { label: '10,000.00+', value: '2.00%', extra: 'Valid bets' },
      { label: '50,000.00+', value: '2.50%', extra: 'Valid bets' },
      { label: '100,000.00+', value: '3.00%', extra: 'Valid bets' },
      { label: '500,000.00+', value: '3.50%', extra: 'Valid bets' },
      { label: '1,000,000.00+', value: '4.00%', extra: 'Valid bets' },
      { label: '5,000,000.00+', value: '4.50%', extra: 'Valid bets' },
      { label: '10,000,000.00+', value: '5.00%', extra: 'Valid bets' },
    ],
    rules: [
      'Daily rescue is calculated from qualifying valid losses.',
      'The reward can be claimed after settlement.',
      'Abnormal gameplay may be excluded from the event.',
    ],
  },
  {
    slug: 'welfare-card',
    variant: 'welfare',
    title: 'Benefit Card',
    subtitle: 'Unlock exclusive daily benefits',
    badge: 'Welfare Card',
    cta: 'Activate Now',
    ctaType: 'claim',
    ruleTitle: 'Benefit Rules',
    theme: 'gold',
    images: {
      heroBg: sourceImage('activity/welfare-card/hero_bg.webp'),
      main: sourceImage('activity/welfare-card/hero_main.webp'),
      coins: sourceImage('activity/welfare-card/hero_coins.webp'),
      chest: sourceImage('activity/welfare-card/card_chest.webp'),
      weekThumb: sourceImage('activity/welfare-card/card_thumb_week.webp'),
      monthThumb: sourceImage('activity/welfare-card/card_thumb_month.webp'),
    },
    rows: [
      { label: 'Weekly Card', value: 'Rs 125.00', extra: 'Instant 20.00 / Daily 15.00 / Price 100.00' },
      { label: 'Monthly Card', value: 'Rs 420.00', extra: 'Instant 60.00 / Daily 12.00 / Price 300.00' },
      { label: 'Monthly Card', value: 'Rs 700.00', extra: 'Instant 100.00 / Daily 20.00 / Price 500.00' },
      { label: 'Monthly Card', value: 'Rs 1400.00', extra: 'Instant 200.00 / Daily 40.00 / Price 1,000.00' },
    ],
    rules: [
      'Benefits refresh after the welfare card is activated.',
      'Daily benefits must be claimed before reset.',
      'Card rewards follow the same audit requirements as bonus wallet rewards.',
    ],
  },
  {
    slug: 'sub-deposit-reward',
    aliases: ['subordinate-recharge'],
    variant: 'sub-deposit',
    title: 'Invite Deposit Cashback',
    subtitle: 'Earn when invited users deposit',
    badge: 'Invite Cashback',
    cta: 'Share Now',
    ctaType: 'share',
    ruleTitle: 'Cashback Rules',
    theme: 'blue',
    images: {
      pageBg: sourceImage('activity/sub-deposit-reward/page_bg.webp'),
      heroBg: sourceImage('activity/sub-deposit-reward/hero_bg.webp'),
      main: sourceImage('activity/sub-deposit-reward/hero_main.webp'),
      ribbon: sourceImage('activity/sub-deposit-reward/ribbon_cycle_range.webp'),
    },
    rows: [
      { label: '100.00', value: '0.01%', extra: 'Cycle Total Recharge (Direct)' },
      { label: '1,000.00', value: '0.02%', extra: 'Cycle Total Recharge (Direct)' },
      { label: '10,000.00', value: '0.02%', extra: 'Cycle Total Recharge (Direct)' },
      { label: '100,000.00', value: '0.03%', extra: 'Cycle Total Recharge (Direct)' },
      { label: '1,000,000.00', value: '0.03%', extra: 'Cycle Total Recharge (Direct)' },
      { label: '10,000,000.00', value: '0.04%', extra: 'Cycle Total Recharge (Direct)' },
    ],
    rules: [
      'Cashback is calculated from invited users successful deposits.',
      'Settlement data is displayed after platform review.',
      'Duplicate or risk-controlled accounts do not count.',
    ],
  },
  {
    slug: 'rebate-ranking',
    variant: 'ranking',
    title: 'Cashback Leaderboard',
    subtitle: 'Rank up for bigger rebates',
    badge: 'Daily Ranking',
    cta: 'View Ranking',
    ctaType: 'rules',
    ruleTitle: 'Leaderboard Rules',
    theme: 'blue',
    images: {
      top1: sourceImage('activity/rebate-ranking/ranking-top1-badge.webp'),
      top2: sourceImage('activity/rebate-ranking/ranking-top2-badge.webp'),
      top3: sourceImage('activity/rebate-ranking/ranking-top3-badge.webp'),
    },
    rows: [
      { label: 'Top 1', value: 'Rs 888.00', extra: 'Cash' },
      { label: 'Top 2', value: 'Rs 588.00', extra: 'Cash' },
      { label: 'Top 3', value: 'Rs 388.00', extra: 'Cash' },
      { label: 'Top 10', value: 'Rs 88.00', extra: 'Cash' },
    ],
    rules: [
      'Ranking is calculated from valid rebate activity during the cycle.',
      'Rewards are settled after the leaderboard closes.',
      'Final ranking is subject to platform review.',
    ],
  },
  {
    slug: 'phone-verification',
    variant: 'security',
    title: 'Phone Verification Bonus',
    subtitle: 'Secure your account and claim rewards',
    badge: 'Security Bonus',
    cta: 'Verify Now',
    ctaType: 'verify',
    ruleTitle: 'Verification Rules',
    theme: 'blue',
    images: {
      heroBg: sourceImage('activity/welfare-card/hero_bg.webp'),
      main: sourceImage('bonus/icon_cash_bonus.webp'),
    },
    rows: [
      { label: 'Phone Bind', value: 'Rs 123.00', extra: 'Bonus' },
      { label: 'Security', value: 'Improved', extra: 'Account' },
      { label: 'Login', value: 'Safer', extra: 'Protect' },
      { label: 'Claim', value: 'Once', extra: 'Reward' },
    ],
    rules: [
      'Each account can receive the verification reward once.',
      'The phone number must pass platform verification.',
      'The reward is issued after binding succeeds.',
    ],
  },
  {
    slug: 'app-download',
    variant: 'app',
    title: 'Download App Bonus',
    subtitle: 'For a better experience and enhanced security',
    badge: 'Download Bonus',
    cta: 'Download App',
    ctaType: 'download',
    ruleTitle: 'Download Rules',
    theme: 'blue',
    images: {
      heroBg: sourceImage('activity/welfare-card/hero_bg.webp'),
      main: sourceImage('activity/guide-flow-dialogs/bg_glow.webp'),
    },
    rows: [
      { label: 'App Install', value: 'Rs 44.43', extra: 'Bonus' },
      { label: 'Secure Login', value: 'Rs 10.00', extra: 'Extra' },
      { label: 'Push Enabled', value: 'Rs 12.00', extra: 'Reward' },
      { label: 'First Claim', value: 'Rs 77.00', extra: 'Max' },
    ],
    rules: [
      'Download the official app and log in with the same account.',
      'The app reward can be claimed once per eligible account.',
      'Abnormal device or account status may require review.',
    ],
  },
]

const sourceActivityMap = new Map<string, SourceActivityConfig>()

for (const activity of sourceActivities) {
  sourceActivityMap.set(activity.slug, activity)
  for (const alias of activity.aliases ?? []) {
    sourceActivityMap.set(alias, activity)
  }
}

export const normalizeActivitySlug = (slug: string) =>
  slug.replace(/^\/+|\/+$/g, '').replace(/^activity\//, '')

export const getSourceActivity = (slug: string) =>
  sourceActivityMap.get(normalizeActivitySlug(slug))
