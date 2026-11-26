<script setup lang="ts">
import { computed } from 'vue'
import { Bottom, Minus, Top } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import type { MetricCardData } from '@/api/adminDashboard'
const props = defineProps<{ metric: MetricCardData }>()
const { t, locale } = useI18n()
const change = computed(() => props.metric.changePercent == null ? null : Number(props.metric.changePercent))
const formatted = computed(() => {
  const value = props.metric.value
  if (value == null) return '-'
  if (props.metric.format === 'percent') return `${Number(value).toLocaleString(locale.value, { maximumFractionDigits: 2 })}%`
  if (props.metric.format === 'money') return `${props.metric.currency || ''} ${Number(value).toLocaleString(locale.value, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`.trim()
  return Number(value).toLocaleString(locale.value)
})
</script>
<template><article class="dashboard-metric"><div class="dashboard-metric__label">{{ t(`dashboardBI.metrics.${metric.key}`) }}</div><strong>{{ formatted }}</strong><div class="dashboard-metric__compare" :class="{ up: change != null && change > 0, down: change != null && change < 0 }"><el-icon><Top v-if="change != null && change > 0"/><Bottom v-else-if="change != null && change < 0"/><Minus v-else/></el-icon><span>{{ change == null ? t('dashboardBI.common.notAvailable') : `${Math.abs(change).toFixed(2)}%` }}</span><em>{{ t('dashboardBI.common.previousPeriod') }}</em></div></article></template>
<style scoped>.dashboard-metric{position:relative;overflow:hidden;min-height:112px;padding:17px 18px;border:1px solid #ebeef5;border-radius:6px;background:#fff;box-shadow:0 1px 3px rgba(31,45,61,.04)}.dashboard-metric::before{content:"";position:absolute;inset:0 auto 0 0;width:3px;background:#409eff}.dashboard-metric__label{font-size:13px;color:#606266}.dashboard-metric strong{display:block;margin-top:10px;color:#303133;font-size:23px;line-height:1.2;font-weight:600}.dashboard-metric__compare{display:flex;align-items:center;gap:4px;margin-top:9px;color:#909399;font-size:12px}.dashboard-metric__compare em{margin-left:3px;font-style:normal}.dashboard-metric__compare.up{color:#67c23a}.dashboard-metric__compare.down{color:#f56c6c}</style>
