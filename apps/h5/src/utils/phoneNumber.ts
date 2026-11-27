export function phoneLimits(configuredLength: unknown, dialingCode: string) {
  const fixed = Number(configuredLength)
  const e164Maximum = 15 - dialingCode.replace(/\D/g, '').length
  if (Number.isInteger(fixed) && fixed > 0) {
    const length = Math.min(fixed, e164Maximum)
    return { min: length, max: length }
  }
  return { min: 6, max: e164Maximum }
}
