import test from 'node:test'
import assert from 'node:assert/strict'
import { phoneLimits } from '../src/utils/phoneNumber.ts'

test('Indonesian variable-length numbers fit within the E.164 limit', () => {
  const limits = phoneLimits(0, '+62')
  assert.deepEqual(limits, { min: 6, max: 13 })
  for (const phone of ['81234567890', '821234567890']) {
    assert.ok(new RegExp(`^[0-9]{${limits.min},${limits.max}}$`).test(phone))
    assert.equal(phone.slice(0, limits.max), phone)
  }
})

test('Existing fixed phone lengths retain their behavior', () => {
  assert.deepEqual(phoneLimits(10, '+92'), { min: 10, max: 10 })
  assert.deepEqual(phoneLimits(10, '+91'), { min: 10, max: 10 })
})
