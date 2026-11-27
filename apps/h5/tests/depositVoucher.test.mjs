import test from 'node:test'
import assert from 'node:assert/strict'
import { depositVoucherHasReward, depositVoucherMatchesMethod } from '../src/services/depositVoucher.ts'

const voucher = { available: true, minAmount: '10000', maxAmount: '50000', methodCodes: 'qris; bank', rewards: [{ amount: '10000' }, { amount: '50000' }] }
test('voucher preview and order selection use the same method and exact reward amount', () => {
  assert.equal(depositVoucherMatchesMethod(voucher, 'QRIS'), true)
  assert.equal(depositVoucherHasReward(voucher, 'qris', 10000), true)
  assert.equal(depositVoucherHasReward(voucher, 'qris', 20000), false)
  assert.equal(depositVoucherHasReward(voucher, 'other', 10000), false)
  assert.equal(depositVoucherHasReward({...voucher, available: false}, 'qris', 10000), false)
  assert.equal(depositVoucherHasReward({...voucher, rewards: []}, 'qris', 10000), false)
  assert.equal(depositVoucherHasReward({...voucher, maxAmount: '0'}, 'bank', 50000), true)
})
