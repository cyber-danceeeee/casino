<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import depositTipAnimation from '../assets/lottie/lottie-deposit-tip.json'
import type {
  H5DepositAmountOptionPayload,
  H5DepositChannelPayload,
  H5DepositMethodPayload,
  H5DepositVoucherPayload,
  H5DepositVoucherRewardPayload,
} from '../services/h5Api'
import { h5Api } from '../services/h5Api'
import { depositVoucherMatchesMethod, depositVoucherHasReward } from '../services/depositVoucher'
import { localDepositImage } from '../utils/localAssets'
import LottiePlayer from './LottiePlayer.vue'
import { currencySymbol, formatCurrency } from '../services/runtimeConfig'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  viewOrder: []
}>()

interface DepositMethodItem extends H5DepositMethodPayload {
  minAmountValue: number
  maxAmountValue: number
  displayIcon: string
}

interface DepositChannelItem extends H5DepositChannelPayload {
  minAmountValue: number
  maxAmountValue: number
  displayIcon: string
}

interface DepositAmountOptionItem extends H5DepositAmountOptionPayload {
  amountValue: number
}

interface DepositVoucherRewardItem extends H5DepositVoucherRewardPayload {
  amountValue: number
  bonusRewardValue: number
  cashRewardValue: number
  rewardAmountValue: number
}

interface DepositVoucherItem extends H5DepositVoucherPayload {
  minAmountValue: number
  maxAmountValue: number
  maxRewardValue: number
  bonusRewardValue: number
  cashRewardValue: number
  rewardAmountValue: number
  displayIcon: string
  rewards: DepositVoucherRewardItem[]
}

const paymentMethodIcon = localDepositImage('payment-method.webp')
const voucherTicketIcon = localDepositImage('voucher_ticket.webp')
const defaultVoucherIcon = localDepositImage('first_deposit_gift.webp')

const methods = ref<DepositMethodItem[]>([])
const channels = ref<DepositChannelItem[]>([])
const amountOptions = ref<DepositAmountOptionItem[]>([])
const vouchers = ref<DepositVoucherItem[]>([])
const selectedMethodCode = ref('')
const selectedChannelCode = ref('')
const selectedVoucherIndex = ref(0)
const amount = ref('')
const returnExpanded = ref(true)
const submitting = ref(false)
const progressVisible = ref(false)
const progressOrderNo = ref('')
const voucherStripRef = ref<HTMLElement | null>(null)
const voucherSheetVisible = ref(false)

const close = () => {
  emit('update:modelValue', false)
}

const parseMoney = (value: unknown) => {
  const numeric = Number(String(value ?? '').replace(/,/g, ''))
  return Number.isFinite(numeric) ? numeric : 0
}

const formatNumber = (value: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)

const formatMoney = (value: number) => formatCurrency(value)

const formatInputAmount = (value: number) => {
  if (!Number.isFinite(value)) return ''
  return Number.isInteger(value) ? String(value) : String(value)
}

const amountValue = computed(() => parseMoney(amount.value))
const selectedVoucher = computed(() => vouchers.value[selectedVoucherIndex.value])

const isVoucherSelectable = (voucher: DepositVoucherItem) => depositVoucherMatchesMethod(voucher, selectedMethodCode.value)
const isVoucherAvailableForAmount = (voucher: DepositVoucherItem | undefined, targetAmount: number) =>
  depositVoucherHasReward(voucher, selectedMethodCode.value, targetAmount)
const configLoading = ref(false)
const configError = ref('')
const canSubmit = computed(() => !configLoading.value && !configError.value && amountValue.value > 0 &&
  channels.value.some(channel => channel.channelCode === selectedChannelCode.value && channel.methodCode === selectedMethodCode.value))
const getVoucherRewardDetail = (voucher: DepositVoucherItem | undefined, targetAmount: number) => {
  const empty = { bonusRewardValue: 0, cashRewardValue: 0, rewardAmountValue: 0 }
  if (!isVoucherAvailableForAmount(voucher, targetAmount)) return empty

  const matched = voucher?.rewards.find((reward) => reward.amountValue === targetAmount)
  if (matched) return matched

  return {
    bonusRewardValue: voucher?.bonusRewardValue || 0,
    cashRewardValue: voucher?.cashRewardValue || 0,
    rewardAmountValue: voucher?.rewardAmountValue || 0,
  }
}

const getOptionReward = (option: DepositAmountOptionItem) =>
  getVoucherRewardDetail(selectedVoucher.value, option.amountValue).rewardAmountValue

const selectedReward = computed(() => getVoucherRewardDetail(selectedVoucher.value, amountValue.value))
const totalGet = computed(() => amountValue.value + selectedReward.value.rewardAmountValue)
const returnDetails = computed(() => {
  const items = [
    { label: 'Deposit Amount:', value: amountValue.value, dot: 'is-white' },
    { label: 'Bonus Rewards:', value: selectedReward.value.bonusRewardValue, dot: 'is-violet' },
    { label: 'Cash Rewards:', value: selectedReward.value.cashRewardValue, dot: 'is-amber' },
  ]
  return items.filter((item) => item.value > 0 || item.label === 'Deposit Amount:')
})

const selectedChannels = computed(() =>
  channels.value.filter((channel) => !selectedMethodCode.value || channel.methodCode === selectedMethodCode.value),
)

const selectMethod = (methodCode: string) => {
  selectedMethodCode.value = methodCode
  const firstChannel = channels.value.find((channel) => channel.methodCode === methodCode)
  if (firstChannel) selectedChannelCode.value = firstChannel.channelCode
}

const selectChannel = (channelCode: string) => {
  selectedChannelCode.value = channelCode
}

const selectAmount = (option: DepositAmountOptionItem) => {
  amount.value = formatInputAmount(option.amountValue)
}

const scrollSelectedVoucherIntoView = (behavior: ScrollBehavior = 'smooth') => {
  const target = voucherStripRef.value?.querySelectorAll<HTMLElement>('.game-voucher-ticket')[selectedVoucherIndex.value]
  target?.scrollIntoView({ behavior, inline: 'center', block: 'nearest' })
}

const selectVoucher = (index: number) => {
  const voucher = vouchers.value[index]
  if (!voucher || !isVoucherSelectable(voucher)) return
  selectedVoucherIndex.value = index
  void nextTick().then(() => scrollSelectedVoucherIntoView())
}

const selectVoucherFromSheet = (index: number) => {
  selectVoucher(index)
  const voucher = vouchers.value[index]
  if (voucher && isVoucherSelectable(voucher)) voucherSheetVisible.value = false
}

const applyConfig = (payload: Awaited<ReturnType<typeof h5Api.getDepositConfig>>) => {
  methods.value = (payload.methods || []).map((method) => ({
    ...method,
    minAmountValue: parseMoney(method.minAmount),
    maxAmountValue: parseMoney(method.maxAmount),
    displayIcon: method.iconURL || paymentMethodIcon,
  }))
  channels.value = (payload.channels || []).map((channel) => ({
    ...channel,
    minAmountValue: parseMoney(channel.minAmount),
    maxAmountValue: parseMoney(channel.maxAmount),
    displayIcon: methods.value.find((method) => method.methodCode === channel.methodCode)?.displayIcon || paymentMethodIcon,
  }))
  amountOptions.value = (payload.amountOptions || []).map((option) => ({
    ...option,
    amountValue: parseMoney(option.amount),
  }))
  vouchers.value = (payload.vouchers || []).map((voucher) => {
    const bonusRewardValue = parseMoney(voucher.bonusReward)
    const cashRewardValue = parseMoney(voucher.cashReward)
    return {
      ...voucher,
      minAmountValue: parseMoney(voucher.minAmount),
      maxAmountValue: parseMoney(voucher.maxAmount),
      maxRewardValue: parseMoney(voucher.maxReward),
      bonusRewardValue,
      cashRewardValue,
      rewardAmountValue: parseMoney(voucher.rewardAmount) || bonusRewardValue + cashRewardValue,
      displayIcon: voucher.iconURL || defaultVoucherIcon,
      rewards: (voucher.rewards || []).map((reward) => {
        const rewardBonusValue = parseMoney(reward.bonusReward)
        const rewardCashValue = parseMoney(reward.cashReward)
        return {
          ...reward,
          amountValue: parseMoney(reward.amount),
          bonusRewardValue: rewardBonusValue,
          cashRewardValue: rewardCashValue,
          rewardAmountValue: parseMoney(reward.rewardAmount) || rewardBonusValue + rewardCashValue,
        }
      }),
    }
  })

  if (!selectedMethodCode.value && methods.value[0]) selectMethod(methods.value[0].methodCode)
  if (!selectedChannelCode.value && selectedChannels.value[0]) selectedChannelCode.value = selectedChannels.value[0].channelCode
  if (!amount.value && amountOptions.value[0]) amount.value = formatInputAmount(amountOptions.value[0].amountValue)
  const availableVoucher = vouchers.value.findIndex((voucher) => isVoucherSelectable(voucher))
  selectedVoucherIndex.value = availableVoucher >= 0 ? availableVoucher : 0
  void nextTick().then(() => scrollSelectedVoucherIntoView('auto'))
}

const loadConfig = async () => {
  if (configLoading.value) return
  configLoading.value = true
  configError.value = ''
  try {
    applyConfig(await h5Api.getDepositConfig())
  } catch {
    configError.value = 'Unable to load deposit options. Please refresh and try again.'
  } finally {
    configLoading.value = false
  }
}

const submitDeposit = async () => {
  if (submitting.value || !canSubmit.value) return
  submitting.value = true
  try {
    const payload = await h5Api.createDeposit({
      amount: amount.value,
      methodCode: selectedMethodCode.value,
      channelCode: selectedChannelCode.value,
      voucherCode: isVoucherAvailableForAmount(selectedVoucher.value, amountValue.value) ? selectedVoucher.value?.voucherCode : undefined,
    })
    progressOrderNo.value = payload.order?.orderNo || ''
    progressVisible.value = true
  } finally {
    submitting.value = false
  }
}

const closeProgress = () => {
  progressVisible.value = false
}

const viewOrder = () => {
  progressVisible.value = false
  close()
  emit('viewOrder')
}

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) void loadConfig()
  },
  { immediate: true },
)
</script>

<template>
  <Teleport to="body">
    <Transition name="game-deposit">
      <div v-if="modelValue" class="game-deposit-overlay" role="dialog" aria-modal="true" aria-label="Deposit">
        <div class="game-deposit-sheet">
          <div class="game-deposit-sheet__header">
            <span>Deposit</span>
            <button type="button" aria-label="Close" @click="close">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M6.29 7.71a1 1 0 0 1 1.42-1.42L12 10.59l4.29-4.3a1 1 0 1 1 1.42 1.42L13.41 12l4.3 4.29a1 1 0 0 1-1.42 1.42L12 13.41l-4.29 4.3a1 1 0 0 1-1.42-1.42l4.3-4.29-4.3-4.29Z" />
              </svg>
            </button>
          </div>

          <div class="game-deposit-sheet__body">
            <section class="game-deposit-panel" aria-label="Deposit form">
              <div class="game-deposit-title">
                <img :src="paymentMethodIcon" alt="" aria-hidden="true" />
                <h2>Payment Method</h2>
                <button type="button" aria-label="Refresh payment methods" @click="loadConfig">
                  <svg viewBox="0 0 24 24" aria-hidden="true">
                    <path d="M20.24 12.24a1 1 0 0 0-1.22.72 7 7 0 0 1-11.4 3.45h1.64a1 1 0 1 0 0-2H5.1a1 1 0 0 0-1 1v4.16a1 1 0 1 0 2 0v-1.7a9 9 0 0 0 14.86-4.41 1 1 0 0 0-.72-1.22ZM4.98 11.04a7 7 0 0 1 11.4-3.45h-1.64a1 1 0 1 0 0 2h4.16a1 1 0 0 0 1-1V4.43a1 1 0 1 0-2 0v1.7A9 9 0 0 0 3.04 10.54a1 1 0 1 0 1.94.5Z" />
                  </svg>
                </button>
              </div>

              <div class="game-deposit-methods">
                <button
                  v-for="method in methods"
                  :key="method.methodCode"
                  type="button"
                  :class="{ 'is-active': selectedMethodCode === method.methodCode }"
                  @click="selectMethod(method.methodCode)"
                >
                  <img class="game-deposit-logo" :src="method.displayIcon" alt="" loading="lazy" aria-hidden="true" />
                  <span>{{ method.name }}</span>
                </button>
              </div>

              <p v-if="configLoading" class="game-deposit-info" role="status">Loading deposit options…</p>
              <p v-else-if="configError" class="game-deposit-info" role="alert">{{ configError }}</p>
              <p v-else-if="!methods.length" class="game-deposit-info" role="status">Payment channels are currently unavailable. Please contact support or try again later.</p>
              <div class="game-deposit-channels">
                <button
                  v-for="channel in selectedChannels"
                  :key="channel.channelCode"
                  type="button"
                  :class="{ 'is-active': selectedChannelCode === channel.channelCode }"
                  @click="selectChannel(channel.channelCode)"
                >
                  <img class="game-deposit-logo" :src="channel.displayIcon" alt="" loading="lazy" aria-hidden="true" />
                  <span>{{ channel.name }}</span>
                </button>
              </div>

              <div class="game-deposit-amounts">
                <button
                  v-for="option in amountOptions"
                  :key="option.optionCode"
                  type="button"
                  :class="{ 'is-active': option.amountValue === amountValue }"
                  @click="selectAmount(option)"
                >
                  <span>{{ formatMoney(option.amountValue) }}</span>
                  <em v-if="getOptionReward(option) > 0">{{ formatNumber(getOptionReward(option)) }}</em>
                </button>
              </div>

              <label class="game-deposit-input">
                <span>{{ currencySymbol() }}</span>
                <input v-model="amount" inputmode="decimal" aria-label="Amount" />
                <button type="button" aria-label="Clear amount" @click="amount = ''">×</button>
              </label>

              <div class="game-deposit-label">
                <span><img :src="voucherTicketIcon" alt="" aria-hidden="true" /> Deposit Voucher</span>
                <button type="button" @click="voucherSheetVisible = true">More <i aria-hidden="true" /></button>
              </div>

              <p v-if="!configLoading && !configError && !vouchers.length" class="game-deposit-info">No deposit vouchers are available for your account.</p>
              <p v-if="selectedVoucher?.unavailableReason" class="game-deposit-info">{{ selectedVoucher.unavailableReason }}</p>
              <div ref="voucherStripRef" class="game-voucher-strip" aria-label="Deposit vouchers">
                <button
                  v-for="(voucher, index) in vouchers"
                  :key="voucher.id"
                  type="button"
                  class="game-voucher-ticket"
                  :class="{
                    'is-active': selectedVoucherIndex === index,
                    'is-muted': selectedVoucherIndex !== index,
                    'is-disabled': !isVoucherSelectable(voucher),
                  }"
                  @click="selectVoucher(index)"
                >
                  <svg class="game-voucher-ticket__shape" viewBox="0 0 340 72" preserveAspectRatio="none" fill="none" aria-hidden="true">
                    <defs>
                      <linearGradient :id="`game-voucher-fill-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                        <stop stop-color="#F4E3B5" />
                        <stop offset="1" stop-color="#F2C364" />
                      </linearGradient>
                      <linearGradient :id="`game-voucher-muted-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                        <stop stop-color="#5B4F38" />
                        <stop offset="1" stop-color="#7A5A22" />
                      </linearGradient>
                    </defs>
                    <path
                      :fill="selectedVoucherIndex === index ? `url(#game-voucher-fill-${voucher.id})` : `url(#game-voucher-muted-${voucher.id})`"
                      d="M67.6748 0C68.7984 2.3647 71.2078 4 74 4C76.7922 4 79.2016 2.3647 80.3252 0H331C332.742 0 334.352 0.558687 335.666 1.50391C334.043 2.78584 333 4.77086 333 7C333 10.5264 335.608 13.4414 339 13.9268V19.0723C335.608 19.5576 333 22.4736 333 26C333 29.5264 335.608 32.4414 339 32.9268V38.0723C335.608 38.5576 333 41.4736 333 45C333 48.5264 335.608 51.4414 339 51.9268V57.0723C335.608 57.5576 333 60.4736 333 64C333 66.5208 334.333 68.7294 336.332 69.9619C334.917 71.2286 333.049 72 331 72H80.3252C79.2016 69.6353 76.7922 68 74 68C71.2078 68 68.7984 69.6353 67.6748 72H8C3.58172 72 0 68.4183 0 64V8C0 3.58172 3.58172 0 8 0H67.6748Z"
                    />
                    <path d="M74 6L74 66" stroke="#090F1E" stroke-dasharray="2 2" />
                  </svg>
                  <span class="game-voucher-ticket__check" aria-hidden="true">
                    <svg viewBox="0 0 22 22" fill="none">
                      <path d="M10.8 18.8C13.009 18.8 15.009 17.905 16.457 16.458C17.904 15.01 18.8 13.01 18.8 10.8C18.8 8.592 17.904 6.592 16.457 5.144C15.009 3.696 13.009 2.8 10.8 2.8C8.591 2.8 6.591 3.696 5.143 5.144C3.695 6.592 2.8 8.592 2.8 10.8C2.8 13.01 3.695 15.01 5.143 16.458C6.591 17.905 8.591 18.8 10.8 18.8Z" />
                      <path v-if="isVoucherSelectable(voucher)" d="M7.55 10.8L9.8 13.05L14.3 8.55" />
                    </svg>
                  </span>
                  <img :src="voucher.displayIcon" alt="" loading="lazy" />
                  <strong>{{ voucher.title }}</strong>
                  <span>{{ formatNumber(voucher.minAmountValue) }}–{{ formatNumber(voucher.maxAmountValue) }}</span>
                  <small>Available Amount Range</small>
                  <em><span>Max</span> {{ formatNumber(voucher.maxRewardValue) }}</em>
                </button>
              </div>
            </section>

            <div class="game-deposit-action">
              <div class="game-deposit-return" :class="{ 'is-expanded': returnExpanded }">
                <button type="button" @click="returnExpanded = !returnExpanded">
                  <span>
                    You Will Get
                    <svg width="14" height="14" viewBox="0 0 18 18" fill="none" :class="{ 'is-expanded': returnExpanded }" aria-hidden="true">
                      <path d="M9 13.5L3.375 6.75H14.625L9 13.5Z" fill="#FFFFFF" stroke="#FFFFFF" stroke-linejoin="round" />
                    </svg>
                  </span>
                  <strong>{{ formatMoney(totalGet) }}</strong>
                </button>
                <div class="game-deposit-return__details">
                  <div>
                    <p v-for="item in returnDetails" :key="item.label">
                      <i :class="item.dot" aria-hidden="true" />
                      <span>{{ item.label }}</span>
                      <strong>{{ formatMoney(item.value) }}</strong>
                    </p>
                  </div>
                </div>
              </div>
              <button class="game-deposit-submit" type="button" :disabled="submitting || !canSubmit" @click="submitDeposit">
                Deposit
              </button>
            </div>
          </div>
        </div>

        <Transition name="game-deposit-progress">
          <div v-if="progressVisible" class="game-deposit-progress" role="dialog" aria-modal="true" aria-label="Deposit in progress">
            <article>
              <div class="game-deposit-progress__animation" aria-hidden="true">
                <LottiePlayer :animation-data="depositTipAnimation" :loop="true" :autoplay="progressVisible" />
              </div>
              <h3>Deposit in progress</h3>
              <p>Order No.: {{ progressOrderNo || '-' }}</p>
              <button type="button" class="is-primary" @click="viewOrder">View Order</button>
              <button type="button" @click="closeProgress">Deposit Again</button>
            </article>
          </div>
        </Transition>

        <Transition name="game-deposit">
          <div v-if="voucherSheetVisible" class="game-deposit-voucher-layer" @click.self="voucherSheetVisible = false">
            <div class="game-deposit-voucher-sheet">
              <div class="game-deposit-voucher-sheet__handle"><span /></div>
              <h3>Deposit Voucher</h3>
              <div class="game-deposit-voucher-sheet__list">
                <button
                  v-for="(voucher, index) in vouchers"
                  :key="`sheet-${voucher.id}`"
                  type="button"
                  class="game-voucher-ticket is-sheet"
                  :class="{
                    'is-active': selectedVoucherIndex === index,
                    'is-muted': selectedVoucherIndex !== index,
                    'is-disabled': !isVoucherSelectable(voucher),
                  }"
                  @click="selectVoucherFromSheet(index)"
                >
                  <svg class="game-voucher-ticket__shape" viewBox="0 0 340 72" preserveAspectRatio="none" fill="none" aria-hidden="true">
                    <defs>
                      <linearGradient :id="`game-sheet-voucher-fill-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                        <stop stop-color="#F4E3B5" />
                        <stop offset="1" stop-color="#F2C364" />
                      </linearGradient>
                      <linearGradient :id="`game-sheet-voucher-muted-${voucher.id}`" x1="169.5" y1="0" x2="169.5" y2="72" gradientUnits="userSpaceOnUse">
                        <stop stop-color="#5B4F38" />
                        <stop offset="1" stop-color="#7A5A22" />
                      </linearGradient>
                    </defs>
                    <path
                      :fill="selectedVoucherIndex === index ? `url(#game-sheet-voucher-fill-${voucher.id})` : `url(#game-sheet-voucher-muted-${voucher.id})`"
                      d="M67.6748 0C68.7984 2.3647 71.2078 4 74 4C76.7922 4 79.2016 2.3647 80.3252 0H331C332.742 0 334.352 0.558687 335.666 1.50391C334.043 2.78584 333 4.77086 333 7C333 10.5264 335.608 13.4414 339 13.9268V19.0723C335.608 19.5576 333 22.4736 333 26C333 29.5264 335.608 32.4414 339 32.9268V38.0723C335.608 38.5576 333 41.4736 333 45C333 48.5264 335.608 51.4414 339 51.9268V57.0723C335.608 57.5576 333 60.4736 333 64C333 66.5208 334.333 68.7294 336.332 69.9619C334.917 71.2286 333.049 72 331 72H80.3252C79.2016 69.6353 76.7922 68 74 68C71.2078 68 68.7984 69.6353 67.6748 72H8C3.58172 72 0 68.4183 0 64V8C0 3.58172 3.58172 0 8 0H67.6748Z"
                    />
                    <path d="M74 6L74 66" stroke="#090F1E" stroke-dasharray="2 2" />
                  </svg>
                  <span class="game-voucher-ticket__check" aria-hidden="true">
                    <svg viewBox="0 0 22 22" fill="none">
                      <path d="M10.8 18.8C13.009 18.8 15.009 17.905 16.457 16.458C17.904 15.01 18.8 13.01 18.8 10.8C18.8 8.592 17.904 6.592 16.457 5.144C15.009 3.696 13.009 2.8 10.8 2.8C8.591 2.8 6.591 3.696 5.143 5.144C3.695 6.592 2.8 8.592 2.8 10.8C2.8 13.01 3.695 15.01 5.143 16.458C6.591 17.905 8.591 18.8 10.8 18.8Z" />
                      <path v-if="isVoucherSelectable(voucher)" d="M7.55 10.8L9.8 13.05L14.3 8.55" />
                    </svg>
                  </span>
                  <img :src="voucher.displayIcon" alt="" loading="lazy" />
                  <strong>{{ voucher.title }}</strong>
                  <span>{{ formatNumber(voucher.minAmountValue) }}–{{ formatNumber(voucher.maxAmountValue) }}</span>
                  <small>Available Amount Range</small>
                  <em><span>Max</span> {{ formatNumber(voucher.maxRewardValue) }}</em>
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped lang="scss">
.game-deposit-info { color: #b4c7de; font-size: 12px; line-height: 1.6; margin: 10px 0; }
.game-deposit-overlay {
  position: fixed;
  inset: 0;
  z-index: 2147483646;
  isolation: isolate;
  display: flex;
  max-width: 430px;
  align-items: flex-end;
  margin: 0 auto;
  color: #ffffff;
  background: rgba(0, 0, 0, 0.5);
}

.game-deposit-sheet {
  width: 100%;
  max-height: 85dvh;
  overflow: hidden;
  border-radius: 16px 16px 0 0;
  background: #1a2035;
}

.game-deposit-sheet__header {
  position: relative;
  display: flex;
  height: 44px;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);

  span {
    font-size: 14px;
    font-weight: 600;
  }

  button {
    position: absolute;
    right: 14px;
    display: grid;
    width: 28px;
    height: 28px;
    place-items: center;
    color: rgba(255, 255, 255, 0.6);

    svg {
      width: 17px;
      height: 17px;
      fill: currentColor;
    }
  }
}

.game-deposit-sheet__body {
  display: flex;
  max-height: calc(85dvh - 44px);
  flex-direction: column;
  background: #090f1e;
}

.game-deposit-panel {
  min-height: 0;
  flex: 1 1 auto;
  overflow-x: hidden;
  overflow-y: auto;
  padding: 16px 14px 10px;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.game-deposit-title,
.game-deposit-label {
  display: flex;
  align-items: center;
  color: #ffffff;
}

.game-deposit-title {
  margin-bottom: 10px;

  img {
    width: 22px;
    height: 22px;
    margin-right: 8px;
    object-fit: contain;
  }

  h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 500;
  }

  button {
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
}

.game-deposit-methods {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 7px;
  margin-bottom: 10px;
}

.game-deposit-methods button,
.game-deposit-channels button {
  display: flex;
  height: 37px;
  min-width: 0;
  align-items: center;
  justify-content: flex-start;
  gap: 6px;
  overflow: hidden;
  border: 1px solid #1e2438;
  color: #ffffff;
  background: #1e2438;

  &.is-active {
    border-color: #387ee5;
    background: linear-gradient(277deg, #2e3754 3%, #46537f 96%);
    box-shadow: 0 0 8px rgba(56, 126, 229, 0.5);
  }

  > span:last-child {
    min-width: 0;
    overflow: hidden;
    font-size: 13px;
    font-weight: 500;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.game-deposit-methods button {
  padding: 0 10px;
  border-radius: 10px;
}

.game-deposit-channels {
  display: grid;
  gap: 8px;
  margin-bottom: 17px;

  button {
    width: 100%;
    padding: 0 10px;
    border-radius: 999px;
  }
}

.game-deposit-logo {
  display: block;
  width: 21px;
  height: 21px;
  flex: 0 0 21px;
  overflow: hidden;
  border-radius: 4px;
  object-fit: contain;
}

.game-deposit-amounts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 7px;
  margin-bottom: 15px;

  button {
    position: relative;
    display: flex;
    min-height: 50px;
    align-items: center;
    justify-content: center;
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
      font-weight: 500;
      white-space: nowrap;
    }

    em {
      position: absolute;
      top: 0;
      right: 0;
      min-width: 44px;
      height: 15px;
      padding: 0 4px;
      border-radius: 0 0 0 6px;
      color: #ffffff;
      font-size: 10px;
      font-style: italic;
      font-weight: 600;
      line-height: 15px;
      background: #387ee5;
    }
  }
}

.game-deposit-input {
  display: flex;
  height: 37px;
  align-items: center;
  overflow: hidden;
  border: 1px solid #2d3652;
  border-radius: 8px;

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
    width: 30px;
    height: 100%;
    color: #8c98bb;
    font-size: 15px;
  }
}

.game-deposit-label {
  height: 38px;
  justify-content: space-between;
  margin-top: 10px;

  > span {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;

    img {
      width: 20px;
      height: 20px;
      object-fit: contain;
    }
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

.game-voucher-strip {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding: 0 0 6px;
  scroll-behavior: smooth;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }

  &::after {
    display: block;
    flex: 0 0 56px;
    content: "";
  }
}

.game-voucher-ticket {
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

  &.is-active {
    transform: translateY(-1px);
  }

  &.is-muted {
    opacity: 0.78;
  }

  &.is-disabled {
    opacity: 0.52;
  }

  > *:not(.game-voucher-ticket__shape) {
    position: absolute;
    z-index: 1;
  }

  > img {
    top: 50%;
    left: calc((21.3% - 50px) / 2);
    width: 50px;
    height: 50px;
    border-radius: 6px;
    object-fit: contain;
    transform: translateY(-50%);
  }

  > strong {
    top: 10px;
    right: 44px;
    left: 78px;
    overflow: visible;
    font-size: 14px;
    font-weight: 800;
    line-height: 16.8px;
    white-space: nowrap;
  }

  > span:not(.game-voucher-ticket__check) {
    top: 31px;
    right: 44px;
    left: 78px;
    color: #2d2410;
    font-size: 14px;
    font-weight: 600;
    line-height: 16.8px;
    white-space: nowrap;
  }

  > small {
    top: 49px;
    right: 44px;
    left: 78px;
    color: rgba(45, 36, 16, 0.72);
    font-size: 10px;
    font-weight: 500;
    line-height: 12px;
    white-space: nowrap;
  }

  > em {
    right: 17px;
    bottom: 2px;
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

    span {
      color: rgba(255, 255, 255, 0.9);
      font-weight: 400;
    }
  }
}

.game-voucher-ticket__shape {
  position: absolute;
  inset: 2px;
  width: auto;
  height: auto;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.22));
}

.game-voucher-ticket__check {
  top: 11px;
  right: 18px;
  display: grid;
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
  }

  path:last-child {
    stroke: #edd496;
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.game-voucher-ticket.is-active .game-voucher-ticket__check path:first-child {
  stroke: #edd496;
}

.game-voucher-ticket.is-sheet {
  width: 100%;
  flex-basis: auto;
}

.game-deposit-action {
  flex: 0 0 auto;
  padding: 13px 14px calc(env(safe-area-inset-bottom) + 24px);
  background: #090f1e;
}

.game-deposit-return {
  margin-bottom: 24px;

  > button {
    display: flex;
    width: 100%;
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
      color: #56c36c;
      font-size: 16px;
      font-weight: 600;
    }
  }
}

.game-deposit-return__details {
  max-height: 0;
  overflow: hidden;
  transition: max-height 280ms cubic-bezier(0.32, 0.72, 0, 1);

  .game-deposit-return.is-expanded & {
    max-height: 78px;
  }

  > div {
    padding-top: 9px;
  }

  p {
    display: flex;
    align-items: center;
    margin: 0;
    color: rgba(255, 255, 255, 0.9);
    font-size: 14px;

    + p {
      margin-top: 8px;
    }

    i {
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

    span {
      flex: 1;
    }

    strong {
      color: #ffffff;
      font-size: 14px;
      font-weight: 400;
    }
  }
}

.game-deposit-submit {
  width: 100%;
  height: 48px;
  border-radius: 6px;
  color: #ffffff;
  font-size: 17px;
  font-weight: 500;
  background: linear-gradient(118deg, #3d8aef, #2f69d1);
  box-shadow: 0 5px 12px rgba(47, 105, 209, 0.35);

  &:disabled {
    opacity: 0.58;
  }
}

.game-deposit-progress,
.game-deposit-voucher-layer {
  position: fixed;
  inset: 0;
  z-index: 1;
  display: flex;
  max-width: 430px;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  background: rgba(0, 0, 0, 0.62);
}

.game-deposit-progress article {
  width: calc(100% - 40px);
  overflow: visible;
  border-radius: 16px;
  padding: 0 16px 16px;
  text-align: center;
  background: #1e2438;

  h3 {
    margin: 12px 0 0;
    font-size: 20px;
    font-weight: 600;
  }

  p {
    margin: 12px 0 24px;
    color: rgba(255, 255, 255, 0.6);
    font-size: 14px;
  }

  button {
    width: 100%;
    height: 46px;
    border-radius: 6px;
    color: #ffffff;
    font-size: 16px;
    background: rgba(255, 255, 255, 0.1);

    + button {
      margin-top: 10px;
    }

    &.is-primary {
      background: linear-gradient(118deg, #3d8aef, #2f69d1);
    }
  }
}

.game-deposit-progress__animation {
  width: 160px;
  height: 160px;
  margin: -80px auto 0;
}

.game-deposit-voucher-layer {
  align-items: flex-end;
}

.game-deposit-voucher-sheet {
  width: 100%;
  max-height: 56dvh;
  padding: 22px 22px 18px;
  overflow-y: auto;
  border-radius: 20px 20px 0 0;
  background: #090f1e;

  h3 {
    margin: 0 0 16px;
    font-size: 18px;
    font-weight: 500;
  }
}

.game-deposit-voucher-sheet__handle {
  position: absolute;
  left: 0;
  right: 0;
  top: 10px;
  display: flex;
  justify-content: center;

  span {
    width: 80px;
    height: 7px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.28);
  }
}

.game-deposit-voucher-sheet__list {
  display: grid;
  gap: 12px;
}

.game-deposit-enter-active,
.game-deposit-leave-active,
.game-deposit-progress-enter-active,
.game-deposit-progress-leave-active {
  transition: opacity 180ms ease;
}

.game-deposit-enter-from,
.game-deposit-leave-to,
.game-deposit-progress-enter-from,
.game-deposit-progress-leave-to {
  opacity: 0;
}

.game-deposit-enter-active .game-deposit-sheet,
.game-deposit-leave-active .game-deposit-sheet {
  transition: transform 220ms cubic-bezier(0.32, 0.72, 0, 1);
}

.game-deposit-enter-from .game-deposit-sheet,
.game-deposit-leave-to .game-deposit-sheet {
  transform: translateY(100%);
}
</style>
