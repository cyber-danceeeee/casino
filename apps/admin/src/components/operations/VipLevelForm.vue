<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
const props = defineProps<{ form: Record<string, unknown>; editing: boolean }>()
const { t } = useI18n()
const retentionEnabled = computed(() => Number(props.form.retainPeriodDays || 0) > 0)
const rewards = ['levelUpReward', 'weeklyReward', 'monthlyReward']
const upgrade = ['requiredDeposit', 'requiredValidBet']
const retention = ['retainDeposit', 'retainValidBet']
function toggleRetention(enabled: boolean) {
  props.form.retainPeriodDays = enabled ? 30 : 0
  if (!enabled) props.form.retainDeposit = props.form.retainValidBet = '0.00'
}
</script>
<template>
  <div class="vip-level-form">
    <el-alert :title="t('vipEditor.linked')" :description="t('vipEditor.help')" type="info" :closable="false" show-icon />
    <el-form label-position="top">
      <section class="vip-level-form__section">
        <h3>{{ t('vipEditor.rewards') }}</h3>
        <div class="vip-level-form__grid">
          <el-form-item :label="t('operations.fields.level')" required><el-input-number :model-value="Number(form.level || 1)" :min="1" :disabled="editing" @update:model-value="form.level = $event || 1" /></el-form-item>
          <el-form-item v-for="key in rewards" :key="key" :label="t(`operations.fields.${key}`)" required><el-input :model-value="String(form[key] ?? '0.00')" inputmode="decimal" @update:model-value="form[key] = $event" /></el-form-item>
          <el-form-item :label="t('operations.fields.dailyWithdrawCount')"><el-input-number :model-value="Number(form.dailyWithdrawCount || 0)" :min="0" @update:model-value="form.dailyWithdrawCount = $event || 0" /></el-form-item>
          <el-form-item :label="t('operations.fields.enabled')"><el-switch :model-value="Boolean(form.enabled)" @change="form.enabled = Boolean($event)" /></el-form-item>
        </div>
      </section>
      <section class="vip-level-form__section">
        <h3>{{ t('vipEditor.upgrade') }}</h3>
        <div class="vip-level-form__grid"><el-form-item v-for="key in upgrade" :key="key" :label="t(`operations.fields.${key}`)" required><el-input :model-value="String(form[key] ?? '0.00')" inputmode="decimal" @update:model-value="form[key] = $event" /></el-form-item></div>
      </section>
      <section class="vip-level-form__section">
        <h3>{{ t('vipEditor.retention') }}</h3>
        <el-form-item :label="t('vipEditor.retainEnabled')"><el-switch :model-value="retentionEnabled" @change="toggleRetention(Boolean($event))" /></el-form-item>
        <div v-if="retentionEnabled" class="vip-level-form__grid">
          <el-form-item :label="t('operations.fields.retainPeriodDays')" required><el-input-number :model-value="Number(form.retainPeriodDays)" :min="1" @update:model-value="form.retainPeriodDays = $event || 1" /></el-form-item>
          <el-form-item v-for="key in retention" :key="key" :label="t(`operations.fields.${key}`)" required><el-input :model-value="String(form[key] ?? '0.00')" inputmode="decimal" @update:model-value="form[key] = $event" /></el-form-item>
        </div>
        <p v-else>{{ t('vipEditor.retainDisabled') }}</p>
      </section>
    </el-form>
  </div>
</template>
<style scoped>
.vip-level-form {display:grid;gap:16px;}
.vip-level-form__section {border:1px solid var(--el-border-color-lighter);border-radius:10px;padding:16px;margin-bottom:14px;}
.vip-level-form__section h3 {margin:0 0 14px;font-size:15px;}
.vip-level-form__section p {color:var(--el-text-color-secondary);margin:0;font-size:13px;}
.vip-level-form__grid {display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0 20px;}
.vip-level-form :deep(.el-input-number) {width:100%;}
@media(max-width:600px){.vip-level-form__grid{grid-template-columns:1fr;}}
</style>
