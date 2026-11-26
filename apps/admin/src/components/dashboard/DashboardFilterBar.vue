<script setup lang="ts">
import { computed, reactive, watch } from "vue";
import { useI18n } from "vue-i18n";
import type { DashboardFilter, DashboardOptions } from "@/api/adminDashboard";
const props = defineProps<{
  modelValue: DashboardFilter;
  options?: DashboardOptions;
  loading?: boolean;
}>();
const emit = defineEmits<{
  (e: "update:modelValue", v: DashboardFilter): void;
  (e: "apply"): void;
}>();
const { t } = useI18n();
const model = reactive({ ...props.modelValue });
watch(
  () => props.modelValue,
  (v) => Object.assign(model, v),
  { deep: true },
);
const presets = computed(() =>
  [
    "today",
    "yesterday",
    "last7Days",
    "last30Days",
    "thisMonth",
    "lastMonth",
    "custom",
  ].map((value) => ({ value, label: t(`dashboardBI.ranges.${value}`) })),
);
const preset = reactive({ value: "last7Days" });
function pad(v: number) {
  return String(v).padStart(2, "0");
}
function text(d: Date, end = false) {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${end ? "23:59:59" : "00:00:00"}`;
}
function applyPreset(value: string) {
  preset.value = value;
  const now = new Date();
  let start = new Date(now),
    end = new Date(now);
  if (value === "today") {
  } else if (value === "yesterday") {
    start.setDate(start.getDate() - 1);
    end = new Date(start);
  } else if (value === "last7Days") {
    start.setDate(start.getDate() - 6);
  } else if (value === "last30Days") {
    start.setDate(start.getDate() - 29);
  } else if (value === "thisMonth") {
    start = new Date(now.getFullYear(), now.getMonth(), 1);
  } else if (value === "lastMonth") {
    start = new Date(now.getFullYear(), now.getMonth() - 1, 1);
    end = new Date(now.getFullYear(), now.getMonth(), 0);
  } else {
    return;
  }
  model.startTime = text(start);
  model.endTime = text(end, true);
  const days = (end.getTime() - start.getTime()) / 86400000;
  model.granularity =
    days <= 1 ? "hour" : days > 92 ? "month" : days > 45 ? "week" : "day";
}
function submit() {
  emit("update:modelValue", { ...model });
  emit("apply");
}
</script>
<template>
  <el-card class="section-card dashboard-filters" shadow="never"
    ><el-form :model="model" inline label-position="left"
      ><el-form-item :label="t('dashboardBI.filters.timeRange')"
        ><el-select
          :model-value="preset.value"
          style="width: 132px"
          @change="applyPreset"
          ><el-option
            v-for="item in presets"
            :key="item.value"
            :label="item.label"
            :value="item.value" /></el-select></el-form-item
      ><el-form-item v-if="preset.value === 'custom'"
        ><el-date-picker
          v-model="model.startTime"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          :placeholder="t('dashboardBI.filters.startTime')" /><span
          class="range-separator"
          >-</span
        ><el-date-picker
          v-model="model.endTime"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          :placeholder="t('dashboardBI.filters.endTime')" /></el-form-item
      ><el-form-item :label="t('dashboardBI.filters.currency')"
        ><el-select v-model="model.currency" style="width: 100px"
          ><el-option
            v-for="item in options?.currencies"
            :key="item.value"
            :label="item.label"
            :value="String(item.value)" /></el-select></el-form-item
      ><el-form-item :label="t('dashboardBI.filters.country')"
        ><el-select
          v-model="model.country"
          clearable
          filterable
          style="width: 110px"
          ><el-option
            v-for="item in options?.countries"
            :key="item.value"
            :label="item.label"
            :value="String(item.value)" /></el-select></el-form-item
      ><el-form-item :label="t('dashboardBI.filters.provider')"
        ><el-select
          v-model="model.providerCode"
          clearable
          filterable
          style="width: 150px"
          ><el-option
            v-for="item in options?.providers"
            :key="item.value"
            :label="item.label"
            :value="String(item.value)" /></el-select></el-form-item
      ><el-form-item :label="t('dashboardBI.filters.category')"
        ><el-select
          v-model="model.categoryId"
          clearable
          filterable
          style="width: 145px"
          ><el-option
            v-for="item in options?.categories"
            :key="item.value"
            :label="item.label"
            :value="Number(item.value)" /></el-select></el-form-item
      ><el-form-item :label="t('dashboardBI.filters.channel')"
        ><el-select
          v-model="model.channelId"
          clearable
          filterable
          style="width: 145px"
          ><el-option
            v-for="item in options?.channels"
            :key="item.value"
            :label="item.label"
            :value="Number(item.value)" /></el-select></el-form-item
      ><el-form-item :label="t('dashboardBI.filters.agent')"
        ><el-select
          v-model="model.agentId"
          clearable
          filterable
          style="width: 145px"
          ><el-option
            v-for="item in options?.agents"
            :key="item.value"
            :label="item.label"
            :value="Number(item.value)" /></el-select></el-form-item
      ><el-form-item
        ><el-button type="primary" :loading="loading" @click="submit">{{
          t("common.search")
        }}</el-button></el-form-item
      ></el-form
    ></el-card
  >
</template>
<style scoped>
.dashboard-filters :deep(.el-card__body) {
  padding-bottom: 8px;
}
.dashboard-filters :deep(.el-form-item) {
  margin-bottom: 12px;
  margin-right: 14px;
}
.range-separator {
  padding: 0 6px;
  color: #909399;
}
</style>
