export const getSafeRedirect = (value: unknown) => {
  const target = Array.isArray(value) ? value[0] : value

  if (typeof target !== 'string') return ''
  if (!target.startsWith('/') || target.startsWith('//')) return ''
  if (target.startsWith('/auth/')) return ''

  return target
}
