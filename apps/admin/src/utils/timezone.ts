import { computed, ref } from 'vue'

const BACKEND_OFFSET = 8
const storedOffset = Number(window.localStorage.getItem('casino-admin-timezone') || BACKEND_OFFSET)
export const selectedTimezoneOffset = ref(Number.isFinite(storedOffset) ? storedOffset : BACKEND_OFFSET)

const standardOffsets = Array.from({ length: 53 }, (_, index) => index / 2 - 12)
export const timezoneOptions = [...new Set([...standardOffsets, 5.75, 8.75, 12.75])]
  .sort((left, right) => left - right)
  .map((value) => ({ value, label: `UTC${value >= 0 ? '+' : ''}${value}` }))

export const selectedTimezoneLabel = computed(() => `UTC${selectedTimezoneOffset.value >= 0 ? '+' : ''}${selectedTimezoneOffset.value}`)

export function setTimezoneOffset(value: number) {
  selectedTimezoneOffset.value = value
  window.localStorage.setItem('casino-admin-timezone', String(value))
}

export function timezoneOffsetForIANA(timezone: string, at = new Date()) {
  if (!timezone) return null
  try {
    const parts = new Intl.DateTimeFormat('en-US', {
      timeZone: timezone,
      year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23',
    }).formatToParts(at)
    const values = Object.fromEntries(parts.map((part) => [part.type, part.value]))
    const localized = Date.UTC(
      Number(values.year), Number(values.month) - 1, Number(values.day),
      Number(values.hour), Number(values.minute), Number(values.second),
    )
    return Math.round(((localized - at.getTime()) / 3_600_000) * 4) / 4
  } catch {
    return null
  }
}

export function applyProfileTimezone(timezone: string) {
  const offset = timezoneOffsetForIANA(timezone)
  if (offset !== null) setTimezoneOffset(offset)
}

function parseParts(value: string) {
  const match = value.match(/^(\d{4})-(\d{2})-(\d{2})(?:[ T](\d{2}):(\d{2})(?::(\d{2}))?)?$/)
  if (!match) return null
  return match.slice(1).map((part) => Number(part || 0))
}

function formatUTC(milliseconds: number) {
  const date = new Date(milliseconds)
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getUTCFullYear()}-${pad(date.getUTCMonth() + 1)}-${pad(date.getUTCDate())} ${pad(date.getUTCHours())}:${pad(date.getUTCMinutes())}:${pad(date.getUTCSeconds())}`
}

export function formatBackendTime(value: string) {
  if (!value) return '-'

  if (/T.*(?:Z|[+-]\d{2}:?\d{2})$/.test(value)) {
    const instant = new Date(value)
    if (!Number.isNaN(instant.getTime())) {
      return formatUTC(instant.getTime() + selectedTimezoneOffset.value * 60 * 60 * 1000)
    }
  }

  const parts = parseParts(value)
  if (!parts) return value
  const [year, month, day, hour, minute, second] = parts
  const instant = Date.UTC(year, month - 1, day, hour - BACKEND_OFFSET, minute, second)
  return formatUTC(instant + selectedTimezoneOffset.value * 60 * 60 * 1000)
}

export function selectedTimeToBackend(value: string, endOfDay = false) {
  if (!value) return ''
  const parts = parseParts(value)
  if (!parts) return value
  const [year, month, day] = parts
  const hour = endOfDay ? 23 : 0
  const minute = endOfDay ? 59 : 0
  const second = endOfDay ? 59 : 0
  const instant = Date.UTC(year, month - 1, day, hour - selectedTimezoneOffset.value, minute, second)
  return formatUTC(instant + BACKEND_OFFSET * 60 * 60 * 1000)
}
