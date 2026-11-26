<script setup lang="ts">
import { DataAnalysis } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>
<template><div class="chart-empty"><el-icon :size="38"><DataAnalysis /></el-icon><span>{{ t('dashboardBI.common.noData') }}</span></div></template>
<style scoped>.chart-empty{height:100%;min-height:220px;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:10px;color:#a8abb2;font-size:13px}.chart-empty .el-icon{color:#c0c4cc}</style>
