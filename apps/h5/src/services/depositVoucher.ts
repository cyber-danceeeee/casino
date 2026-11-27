import type { H5DepositVoucherPayload } from './h5Api'

export function depositVoucherMatchesMethod(voucher: H5DepositVoucherPayload, methodCode: string) {
  const methods = (voucher.methodCodes || '').split(/[,;\n]/).map(value => value.trim().toUpperCase()).filter(Boolean)
  return voucher.available && (!methods.length || methods.includes('ALL') || methods.includes(methodCode.trim().toUpperCase()))
}

export function depositVoucherHasReward(voucher: H5DepositVoucherPayload | undefined, methodCode: string, amount: number) {
  if (!voucher || !depositVoucherMatchesMethod(voucher, methodCode) || amount <= 0) return false
  if (amount < Number(voucher.minAmount) || (Number(voucher.maxAmount) > 0 && amount > Number(voucher.maxAmount))) return false
  return Boolean(voucher.rewards?.some(reward => Number(reward.amount) === amount))
}
