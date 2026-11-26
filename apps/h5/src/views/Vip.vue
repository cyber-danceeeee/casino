<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from '../components/AppHeader.vue'
import { isLoggedIn } from '../mock/session'
import { h5Api, type H5VipPayload } from '../services/h5Api'
import { currencySymbol, runtimeConfig, toRuntimeAssetUrl } from '../services/runtimeConfig'
import { localVipImage } from '../utils/localAssets'
import rewardsEntry from '../assets/images/reference/vip-rewards-entry.webp'
import vipZeroLabel from '../assets/images/reference/vip-label-zero.webp'
import cashIcon from '../assets/images/register-reward/icon_cash.webp'

const router = useRouter()
const payload = ref<H5VipPayload | null>(null)
const loading = ref(true)
const loadError = ref(false)
const selectedLevel = ref(1)
const requirementTab = ref<'level_up' | 'retain'>('level_up')
const dialog = ref<HTMLDialogElement | null>(null)
const dialogTitle = ref('')
const dialogSections = ref<Array<{ title: string; text: string }>>([])
let loadVersion = 0

const money = (value: unknown) => {
  const amount = Number(value)
  return Number.isFinite(amount)
    ? amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
    : '—'
}
const levels = computed(() => [...(payload.value?.levels || [])]
  .filter((level) => Number.isInteger(Number(level.levelNo)) && Number(level.levelNo) > 0)
  .sort((a, b) => Number(a.levelNo) - Number(b.levelNo)))
const currentLevel = computed(() => Number(payload.value?.currentLevel || 0))
const selected = computed(() => levels.value.find((level) => Number(level.levelNo) === selectedLevel.value))
const current = computed(() => levels.value.find((level) => Number(level.levelNo) === currentLevel.value))
const iconFor = (level: number, iconURL = '') => toRuntimeAssetUrl(iconURL) || localVipImage(`vip${level}.webp`)
const nextRequirement = computed(() => [...(payload.value?.requirements || [])]
  .filter((row) => row.requirementType === 'level_up'
    && Number(row.targetLevelNo) > currentLevel.value
    && levels.value.some((level) => Number(level.levelNo) === Number(row.targetLevelNo)))
  .sort((a, b) => Number(a.targetLevelNo) - Number(b.targetLevelNo))[0])
const depositAmount = computed(() => payload.value?.depositAmount || '0.00')
const atMaxLevel = computed(() => levels.value.length > 0
  && currentLevel.value >= Number(levels.value.at(-1)?.levelNo))
const depositProgress = computed(() => {
  if (!nextRequirement.value) return atMaxLevel.value ? 100 : 0
  const required = Number(nextRequirement.value.depositAmount)
  const amount = Number(depositAmount.value)
  return required > 0 && Number.isFinite(amount) ? Math.min(100, Math.max(0, amount / required * 100)) : 0
})
const rewards = computed(() => selected.value ? [
  { key: 'upgrade', label: 'Level Up Reward', amount: selected.value.levelUpReward },
  { key: 'weekly', label: 'Weekly Reward', amount: selected.value.weeklyReward },
  { key: 'monthly', label: 'Monthly Reward', amount: selected.value.monthlyReward },
] : [])
const requirementRows = computed(() => (payload.value?.requirements || [])
  .filter((row) => row.requirementType === requirementTab.value
    && levels.value.some((level) => Number(level.levelNo) === Number(row.targetLevelNo)))
  .map((row) => ({
    ...row,
    icon: iconFor(Number(row.targetLevelNo), levels.value.find((level) => Number(level.levelNo) === Number(row.targetLevelNo))?.iconURL),
  }))
  .sort((a, b) => Number(a.targetLevelNo) - Number(b.targetLevelNo)))

const adjacentLevel = (step: number) => {
  const index = levels.value.findIndex((level) => Number(level.levelNo) === selectedLevel.value)
  if (!levels.value.length) return selectedLevel.value
  return Number(levels.value[(index + step + levels.value.length) % levels.value.length]!.levelNo)
}
const selectLevel = (step: number) => { selectedLevel.value = adjacentLevel(step) }

const loadVip = async () => {
  const version = ++loadVersion
  loading.value = true
  loadError.value = false
  try {
    const result = await h5Api.getVip()
    if (version !== loadVersion) return
    payload.value = result
    selectedLevel.value = levels.value.find((level) => Number(level.levelNo) === Math.max(1, currentLevel.value))?.levelNo
      ?? Number(levels.value[0]?.levelNo || 1)
  } catch {
    if (version === loadVersion) loadError.value = true
  } finally {
    if (version === loadVersion) loading.value = false
  }
}

const openDialog = async (title: string, sections: Array<{ title: string; text: string }>) => {
  dialogTitle.value = title
  dialogSections.value = sections
  await nextTick()
  if (dialog.value && !dialog.value.open) dialog.value.showModal()
}
const openRewardInfo = (reward: { label: string; amount: string }) => {
  void openDialog(reward.label, [{
    title: `VIP${selectedLevel.value} ${reward.label}`,
    text: `Configured cash reward: ${currencySymbol()} ${money(reward.amount)}. Eligibility and claiming conditions apply. Contact support for help with an eligible reward.`,
  }])
}
const openRules = () => {
  const sections = [
    { title: 'Upgrade', text: 'The level table shows the deposit and valid-bet requirements for each VIP level. Your deposit progress includes settled deposits in the deployment currency.' },
    { title: 'VIP Retention', text: 'The Retain Requirement tab lists the configured period, deposit and valid-bet requirements. A dash means no retention period has been configured for that level.' },
    { title: 'Rewards', text: 'Use the arrows to view each level’s upgrade, weekly and monthly cash rewards. Amounts and requirements come from the current VIP configuration.' },
  ]
  if (selected.value) sections.push({ title: `VIP${selectedLevel.value} Privileges`, text: `Configured daily withdrawal count: ${selected.value.dailyWithdrawCount}. Withdrawal eligibility and account limits still apply.` })
  sections.push({ title: 'Help', text: 'Contact customer support for reward eligibility, claiming conditions or questions about your VIP status.' })
  void openDialog('VIP Rules', sections)
}
const closeOutsideDialog = (event: MouseEvent) => {
  const element = dialog.value
  if (!element || event.target !== element) return
  const rect = element.getBoundingClientRect()
  if (event.clientX < rect.left || event.clientX > rect.right || event.clientY < rect.top || event.clientY > rect.bottom) element.close()
}

onMounted(() => { void loadVip() })
watch(isLoggedIn, () => {
  payload.value = null
  dialog.value?.close()
  void loadVip()
})
onBeforeUnmount(() => { ++loadVersion; dialog.value?.close() })
</script>

<template>
  <div class="vip-shell text-white">
    <AppHeader @login="router.push({ name: 'auth-login', query: { redirect: '/vip' } })" @register="router.push({ name: 'auth-register', query: { redirect: '/vip' } })" />
    <main class="vip-pane" aria-label="VIP" :aria-busy="loading">
      <article class="vip-card" aria-label="Current VIP level">
        <img v-if="currentLevel === 0" class="vip-card__level" :src="vipZeroLabel" alt="VIP0" />
        <span v-else class="vip-card__level-label"><img v-if="iconFor(currentLevel, current?.iconURL)" :src="iconFor(currentLevel, current?.iconURL)" alt="" />VIP {{ currentLevel }}</span>
        <div class="vip-card__deposit">
          <img :src="localVipImage('deposit.webp')" alt="" />
          <span>Deposit: <strong>{{ money(depositAmount) }}</strong><template v-if="nextRequirement"> / {{ money(nextRequirement.depositAmount) }} to VIP{{ nextRequirement.targetLevelNo }}</template><template v-else-if="!loading && atMaxLevel"> · Maximum VIP level</template><template v-else-if="!loading && levels.length"> · Requirements not configured</template></span>
        </div>
        <div class="vip-card__progress" role="progressbar" aria-label="VIP deposit progress" :aria-valuenow="Math.round(depositProgress)" aria-valuemin="0" aria-valuemax="100"><span :style="{ width: `${depositProgress}%` }" /></div>
      </article>

      <button class="vip-rewards-entry" type="button" @click="router.push('/activity/welfare-card')">
        <img :src="rewardsEntry" alt="" />
        <strong>UNLOCK MORE REWARDS</strong>
        <span>Explore weekly &amp; monthly<br />reward cards</span>
        <em>View Rewards <b aria-hidden="true">›</b></em>
      </button>
      <div class="vip-tip"><img :src="localVipImage('bonus-empty.webp')" alt="" /><span>Upgrade your VIP level, get generous benefits!</span></div>

      <div v-if="loadError" class="vip-notice" role="alert">Unable to load VIP information. <button type="button" @click="loadVip">Try again</button></div>
      <article class="vip-benefits" aria-label="VIP benefits">
        <header class="vip-benefits__header">
          <button type="button" :disabled="levels.length < 2" :aria-label="`Previous level: VIP${adjacentLevel(-1)}`" @click="selectLevel(-1)"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m10 7-5 5 5 5M5 12h14" /></svg></button>
          <h1 aria-live="polite">VIP{{ selectedLevel }} Benefits</h1>
          <button type="button" :disabled="levels.length < 2" :aria-label="`Next level: VIP${adjacentLevel(1)}`" @click="selectLevel(1)"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m14 7 5 5-5 5M5 12h14" /></svg></button>
        </header>
        <div class="vip-benefits__body">
          <h2>VIP Benefits</h2>
          <div v-if="loading && !payload" class="vip-notice" role="status">Loading VIP benefits…</div>
          <div v-else-if="!rewards.length" class="vip-notice">No VIP benefits configured.</div>
          <div v-for="reward in rewards" :key="reward.key" class="benefit-row" :class="{ 'is-alt': reward.key === 'weekly' }">
            <span class="benefit-row__label">{{ reward.label }}</span>
            <div class="benefit-row__value">
              <span class="benefit-row__cash"><i><img :src="cashIcon" alt="" />CASH</i><b>{{ currencySymbol() }} {{ money(reward.amount) }}</b></span>
              <span class="benefit-row__delta"><svg viewBox="0 0 12 16" aria-hidden="true"><path d="m6 0 5 7H8c0 4-2 7-6 9l-2-3c4-1 5-3 5-6H2L6 0Z" /></svg>{{ money(reward.amount) }}</span>
              <button type="button" class="benefit-row__info" :aria-label="`About ${reward.label}`" @click="openRewardInfo(reward)"><img :src="localVipImage('info.webp')" alt="" /></button>
            </div>
          </div>
        </div>
      </article>

      <section class="vip-requirements" aria-label="VIP requirements">
        <div class="requirement-tabs" role="tablist" aria-label="VIP requirement type">
          <button id="vip-upgrade-tab" type="button" role="tab" aria-controls="vip-requirement-panel" :aria-selected="requirementTab === 'level_up'" :class="{ 'is-active': requirementTab === 'level_up' }" @click="requirementTab = 'level_up'">Level Up Requirement</button>
          <button id="vip-retain-tab" type="button" role="tab" aria-controls="vip-requirement-panel" :aria-selected="requirementTab === 'retain'" :class="{ 'is-active': requirementTab === 'retain' }" @click="requirementTab = 'retain'">Retain Requirement</button>
        </div>
        <div id="vip-requirement-panel" role="tabpanel" :aria-labelledby="requirementTab === 'level_up' ? 'vip-upgrade-tab' : 'vip-retain-tab'">
          <table class="requirement-table">
            <thead><tr><th scope="col">{{ requirementTab === 'retain' ? 'Level' : 'Next Level' }}</th><th scope="col">{{ requirementTab === 'retain' ? 'Period' : 'Base Level' }}</th><th scope="col">Deposit</th><th scope="col">Valid bets</th></tr></thead>
            <tbody><tr v-for="row in requirementRows" :key="row.id"><td><img v-if="row.icon" :src="row.icon" :alt="`VIP${row.targetLevelNo}`" loading="lazy" /><span v-else>VIP{{ row.targetLevelNo }}</span></td><td>{{ requirementTab === 'retain' ? (row.periodDays > 0 ? `${row.periodDays} Days` : '—') : `VIP ${row.targetLevelNo}` }}</td><td>{{ money(row.depositAmount) }}</td><td>{{ money(row.validBetAmount) }}</td></tr></tbody>
          </table>
          <p v-if="!loading && !requirementRows.length" class="vip-notice">No {{ requirementTab === 'retain' ? 'retention' : 'upgrade' }} requirements configured.</p>
        </div>
      </section>
      <button class="vip-info" type="button" @click="openRules"><span aria-hidden="true">ⓘ</span>VIP Rules Explanation</button>
    </main>
    <Teleport to="body">
      <dialog ref="dialog" class="vip-dialog" aria-labelledby="vip-dialog-title" @click="closeOutsideDialog">
        <header><h2 id="vip-dialog-title">{{ dialogTitle }}</h2><button type="button" aria-label="Close VIP information" @click="dialog?.close()">×</button></header>
        <div class="vip-dialog__content"><section v-for="section in dialogSections" :key="section.title"><h3>{{ section.title }}</h3><p>{{ section.text }}</p></section><p class="vip-dialog__currency">Currency: {{ runtimeConfig.currency.code }}</p></div>
        <button class="vip-dialog__confirm" type="button" @click="dialog?.close()">Okay</button>
      </dialog>
    </Teleport>
  </div>
</template>

<style scoped lang="scss">
.vip-shell {
  width: 100%;
  max-width: 430px;
  min-height: 100dvh;
  margin: 0 auto;
  background: #090f1f;
}

.vip-pane {
  display: grid;
  gap: 16px;
  padding: 15px 15px calc(90px + env(safe-area-inset-bottom));
}

.vip-card {
  height: 100px;
  overflow: hidden;
  padding: 8px 12px 14px;
  border-radius: 12px;
  background: linear-gradient(105deg,#3f429d,#778eff);
}

.vip-card__level {
  display: block;
  width: auto;
  height: 28px;
  margin-bottom: 18px;
}

.vip-card__level-label {
  display: flex;
  align-items: center;
  height: 28px;
  gap: 4px;
  margin-bottom: 18px;
  font-size: 15px;
  font-weight: 700;
  img {
    height: 28px;
    width: 29px;
    object-fit: contain;
  }
}

.vip-card__deposit {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  line-height: 16px;
  white-space: nowrap;
  img {
    width: 18px;
    height: 18px;
  }
  strong {
    color: #7bb8ff;
    font-size: 12px;
    font-weight: 500;
  }
}

.vip-card__progress {
  height: 4px;
  margin-top: 8px;
  border-radius: 8px;
  overflow: hidden;
  background: #60bdcb55;
  span {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: #8df6f2;
    transition: width 250ms ease;
  }
}

.vip-rewards-entry {
  position: relative;
  width: 100%;
  padding: 0;
  overflow: hidden;
  border-radius: 16px;
  text-align: left;
  img {
    display: block;
    width: 100%;
  }
  strong,span,em {
    position: absolute;
    left: 16px;
  }
  strong {
    top: 12px;
    color: #ffe900;
    font-size: 18px;
    line-height: 1;
    font-weight: 900;
    text-shadow: 0 1px 1px #994b00;
  }
  span {
    top: 32px;
    color: #d5e4ff;
    font-size: 12px;
    line-height: 1.2;
  }
  em {
    bottom: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    min-width: 120px;
    height: 29px;
    border-radius: 20px;
    background: linear-gradient(#ffec29,#ffb800);
    box-shadow: inset 0 -2px 0 #db8600;
    color: #17120a;
    font-size: 13px;
    font-style: normal;
    font-weight: 800;
  }
  em b {
    font-size: 20px;
    line-height: 1;
  }
}

.vip-tip {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 35px;
  border-radius: 6px;
  padding: 0 8px 0 40px;
  background: linear-gradient(90deg,#621fd7,#ba76f7);
  font-size: 12px;
  font-weight: 500;
  line-height: 1.3;
  img {
    position: absolute;
    top: -10px;
    left: -8px;
    width: 46px;
    height: 40px;
    object-fit: contain;
  }
}

.vip-benefits {
  overflow: hidden;
  border-radius: 6px;
  padding-bottom: 6px;
  background: linear-gradient(#3714a1,#211169);
}

.vip-benefits__header {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 18px;
  border-block: 1px solid #b5bcff4d;
  background: linear-gradient(270deg,#7734dc00,#7e56ff 49%,#7835dc00);
  font-style: italic;
  text-shadow: 0 1px #5311c1c4;
  h1 {
    flex: 1;
    margin: 0;
    text-align: center;
    font-size: 14px;
    font-weight: 700;
  }
  button {
    display: grid;
    width: 24px;
    height: 24px;
    padding: 0;
    place-items: center;
    color: white;
  }
  button:disabled {
    opacity: .35;
  }
  svg {
    width: 16px;
    height: 16px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.vip-benefits__body {
  padding: 0 6px;
  h2 {
    margin: 0;
    padding: 12px 0 8px;
    text-align: center;
    font-size: 12px;
    font-weight: 400;
    line-height: 15px;
  }
}

.benefit-row {
  display: grid;
  grid-template-columns: 36fr 49fr;
  gap: 1px;
  min-height: 33px;
  font-size: 12px;
  font-weight: 500;
}

.benefit-row__label {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
  background: linear-gradient(90deg,#4c33db,#795fef);
  text-align: center;
}

.benefit-row__value {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  padding: 0 6px 0 15px;
  background: linear-gradient(90deg,#795fef,#4c33db);
}

.benefit-row.is-alt {
  .benefit-row__label {
    background: linear-gradient(90deg,#795fef,#4c33db);
  }
  .benefit-row__value {
    background: linear-gradient(90deg,#4c33db,#795fef);
  }
}

.benefit-row__cash {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 3px;
  white-space: nowrap;
  i {
    display: flex;
    align-items: center;
    height: 12px;
    border-radius: 8px;
    padding: 0 3px 0 1px;
    background: linear-gradient(#c2f594,#33cf84);
    box-shadow: inset 0 0 0 .5px #e8fecd8c;
    color: #156223;
    font-size: 8px;
    font-style: normal;
    font-weight: 700;
  }
  img {
    width: 10px;
    height: 10px;
  }
  b {
    color: #66dc89;
    font-size: 11px;
    font-weight: 600;
  }
}

.benefit-row__delta {
  display: flex;
  align-items: center;
  gap: 2px;
  color: #67efff;
  font-size: 10px;
  font-weight: 700;
  white-space: nowrap;
  svg {
    width: 8px;
    height: 12px;
    fill: currentColor;
    opacity: .8;
  }
}

.benefit-row__info {
  flex-shrink: 0;
  display: grid;
  place-items: center;
  width: 14px;
  height: 24px;
  padding: 0;
  img {
    width: 12px;
    height: 12px;
  }
}

.vip-requirements {
  min-width: 0;
  overflow: hidden;
  border-radius: 12px 12px 0 0;
  background: #1a2036;
}

.requirement-tabs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  height: 37px;
  background: linear-gradient(#26314e,#303b5d);
  button {
    position: relative;
    min-width: 0;
    border-radius: 12px 18px 0 0;
    padding: 0 4px;
    color: #b4baca;
    font-size: 12px;
    font-style: italic;
    white-space: nowrap;
    z-index: 0;
  }
  button.is-active {
    background: linear-gradient(#526ab9,#3b4873);
    color: #fff;
  }
  button.is-active::after {
    position: absolute;
    inset: 0 -13px 0 auto;
    width: 23px;
    border-radius: 0 0 0 20px;
    background: inherit;
    content: '';
    transform: skewX(23deg);
    z-index: -1;
  }
  button:last-child.is-active {
    border-radius: 18px 12px 0 0;
  }
  button:last-child.is-active::after {
    inset: 0 auto 0 -13px;
    transform: skewX(-23deg);
    border-radius: 0 0 20px 0;
  }
}

.requirement-table {
  width: 100%;
  table-layout: fixed;
  border-collapse: separate;
  border-spacing: 1px;
  background: #1a2036;
  font-size: 12px;
  text-align: center;
  th,td {
    height: 41px;
    padding: 0 2px;
    line-height: 1.2;
    overflow-wrap: anywhere;
  }
  th {
    background: #397fe5;
    font-weight: 500;
  }
  th:first-child {
    width: 23%;
    background: #f59e0b;
  }
  td {
    background: #333c5c;
  }
  td img {
    display: block;
    width: 26px;
    height: 25px;
    margin: auto;
    object-fit: contain;
  }
}

.vip-info {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 5px;
  min-height: 28px;
  padding: 0;
  color: #ffffff80;
  font-size: 12px;
}

.vip-notice {
  padding: 14px 8px;
  color: #d7dced;
  text-align: center;
  font-size: 12px;
  button {
    color: #8fc5ff;
    text-decoration: underline;
  }
}

.vip-dialog {
  position: fixed;
  inset: 0;
  width: min(360px,calc(100vw - 32px));
  max-height: min(720px,85dvh);
  margin: auto;
  overflow: hidden;
  border: 1px solid #4c5c85;
  border-radius: 14px;
  padding: 0 20px 20px;
  background: #20283f;
  color: #eef3ff;
  box-shadow: 0 24px 80px #0009;
  &[open] {
    display: flex;
    flex-direction: column;
  }
  &::backdrop {
    background: #000a;
  }
  header {
    position: sticky;
    top: 0;
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: space-between;
    padding: 18px 0 12px;
    background: #20283f;
  }
  h2 {
    margin: 0;
    font-size: 18px;
  }
  header button {
    width: 30px;
    height: 30px;
    font-size: 27px;
    color: #b5c4e2;
  }
}

.vip-dialog__content {
  min-height: 0;
  overflow-y: auto;
  h3 {
    margin: 14px 0 6px;
    font-size: 14px;
  }
  p {
    margin: 0;
    font-size: 13px;
    line-height: 1.6;
    color: #c3cce1;
  }
  .vip-dialog__currency {
    margin-top: 16px;
    font-size: 12px;
  }
}

.vip-dialog__confirm {
  flex-shrink: 0;
  width: 100%;
  min-height: 42px;
  margin-top: 20px;
  border-radius: 8px;
  background: linear-gradient(#4397ff,#2d6aca);
  color: #fff;
  font-size: 15px;
}

@media (max-width: 374px) {
  .vip-pane {
    padding-inline: 10px;
  }
  .vip-rewards-entry strong {
    font-size: 16px;
  }
  .vip-tip {
    font-size: 11px;
  }
  .benefit-row__value {
    padding-left: 7px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .vip-card__progress span {
    transition: none;
  }
}
</style>
