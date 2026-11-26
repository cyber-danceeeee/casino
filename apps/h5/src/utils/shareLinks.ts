export interface ShareableSocialLink {
  id?: string
  platformCode?: string
  label?: string
  url?: string
  shareURL?: string
  shareTitle?: string
  shareText?: string
  shareDescription?: string
}

const encode = (value: string) => encodeURIComponent(value)

export const resolveShareText = (social: ShareableSocialLink, targetUrl: string) => {
  const title = social.shareTitle || 'Join 56MVP'
  return (social.shareText || `${title}: {url}`)
    .replace(/\{url\}/g, targetUrl)
    .replace(/\{title\}/g, title)
    .replace(/\{description\}/g, social.shareDescription || '')
}

export const buildShareUrl = (social: ShareableSocialLink, targetUrl: string) => {
  const rawUrl = social.shareURL || social.url || ''
  if (!rawUrl) return ''

  const title = social.shareTitle || 'Join 56MVP'
  const description = social.shareDescription || ''
  const text = resolveShareText(social, targetUrl)
  const replacements: Record<string, string> = {
    '{url}': encode(targetUrl),
    '{text}': encode(text),
    '{title}': encode(title),
    '{description}': encode(description),
  }

  if (Object.keys(replacements).some((token) => rawUrl.includes(token))) {
    return Object.entries(replacements).reduce(
      (nextUrl, [token, value]) => nextUrl.replaceAll(token, value),
      rawUrl,
    )
  }

  const separator = rawUrl.includes('?') ? '&' : '?'
  const platform = (social.id || social.platformCode || social.label || '').toLowerCase()
  if (platform.includes('whatsapp')) return `${rawUrl}${separator}text=${encode(text)}`
  if (platform.includes('telegram')) return `${rawUrl}${separator}url=${encode(targetUrl)}&text=${encode(text)}`
  if (platform === 'x' || platform.includes('twitter')) return `${rawUrl}${separator}url=${encode(targetUrl)}&text=${encode(text)}`
  if (platform.includes('facebook')) return `${rawUrl}${separator}u=${encode(targetUrl)}&quote=${encode(text)}`
  if (platform.includes('messenger')) return `${rawUrl}${separator}link=${encode(targetUrl)}`
  return rawUrl
}
