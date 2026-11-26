import { createI18n } from 'vue-i18n'
import { messages, type SupportedLocale } from './messages'

const savedLocale = window.localStorage.getItem('casino-admin-locale') as SupportedLocale | null
const defaultLocale: SupportedLocale = savedLocale && savedLocale in messages ? savedLocale : 'zh-CN'

const i18n = createI18n({
  legacy: false,
  locale: defaultLocale,
  fallbackLocale: 'zh-CN',
  messages,
})

export default i18n
