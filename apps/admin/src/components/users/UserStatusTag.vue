<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ value?: string; status?: string; label?: string; typeOverride?: string }>(), { value: '', status: '', typeOverride: '' })
const normalizedValue = computed(() => props.value || props.status)
const normalizedKey = computed(() => String(normalizedValue.value || '').toUpperCase())
const type = computed(() => {
  if (props.typeOverride) return props.typeOverride
  if (['ACTIVE', 'SUCCESS', 'VALID', 'CREDIT', 'CLEARED', 'NORMAL', 'TRUSTED', 'PAID', 'APPROVED', 'COMPLETED', 'ENABLED', 'ONLINE'].includes(normalizedKey.value)) return 'success'
  if (['DEBIT', 'DISABLED', 'FROZEN', 'OBSERVED', 'MEDIUM', 'PENDING', 'PENDING_REVIEW', 'DRAFT', 'PROCESSING', 'REVIEW', 'BUILDING', 'MAINTENANCE', 'SUSPENDED'].includes(normalizedKey.value)) return 'warning'
  if (['ROLLBACK', 'ROLLED_BACK', 'INVALID', 'LOCKED', 'CLOSED', 'FAILED', 'BLOCKED', 'RESTRICTED', 'HIGH', 'CRITICAL', 'RISKY', 'REJECTED', 'CANCELLED', 'EXPIRED'].includes(normalizedKey.value)) return 'danger'
  if (['BALANCE', 'NOT_CHECKED'].includes(normalizedKey.value)) return 'info'
  return 'info'
})
const fixedColor = computed(() => {
  if (normalizedKey.value === 'REFUND') return '#626aef'
  return ''
})
const palette = ['#2f73de', '#626aef', '#7b61ff', '#00a870', '#d97706', '#0891b2', '#b83280', '#475569']
const color = computed(() => {
  if (fixedColor.value) return fixedColor.value
  if (['BALANCE', 'NOT_CHECKED'].includes(normalizedKey.value)) return ''
  if (type.value !== 'info' || !normalizedKey.value) return ''
  let hash = 0
  for (const char of normalizedKey.value) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  return palette[hash % palette.length]
})
</script>

<template>
  <el-tag
    :type="color ? undefined : type"
    :color="color || undefined"
    :style="color ? { color: '#fff', borderColor: color } : undefined"
    size="small"
  >{{ label || normalizedValue || '-' }}</el-tag>
</template>
