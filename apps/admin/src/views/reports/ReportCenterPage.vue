<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  watch,
} from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import * as echarts from "echarts/core";
import { LineChart } from "echarts/charts";
import {
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import { useI18n } from "vue-i18n";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import {
  cancelReportExport,
  createReportExport,
  downloadReportExport,
  retryReportExport,
  type ReportChart,
  type ReportDetail,
} from "@/api/adminReports";
import {
  formatBackendTime,
  selectedTimeToBackend,
  selectedTimezoneOffset,
} from "@/utils/timezone";
import { reportConfigs } from "./config";
import {
  businessColumns,
  businessEntries,
  isInternalIdentifierKey,
  sanitizeBusinessValue,
} from "@/utils/presentation";

echarts.use([
  LineChart,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  CanvasRenderer,
]);

const route = useRoute();
const { t } = useI18n();
const { can, loadPermissions } = useAdminPermissions();
const moduleKey = computed(() =>
  String(route.meta.moduleKey || "financial-report"),
);
const config = computed(() => reportConfigs[moduleKey.value]);
const visibleFilters = computed(
  () =>
    config.value?.filters.filter(
      (item) => !isInternalIdentifierKey(item.key),
    ) || [],
);
const visibleColumns = computed(
  () =>
    config.value?.columns.filter(
      (item) => !isInternalIdentifierKey(item.key),
    ) || [],
);
const query = reactive<Record<string, string | string[]>>({});
const rows = ref<Array<Record<string, unknown>>>([]);
const summaries = ref<Array<Record<string, unknown>>>([]);
const charts = ref<ReportChart[]>([]);
const metricBasis = ref<Record<string, string>>({});
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<ReportDetail>({});
const activeTab = ref("");
const chartElements = ref<HTMLElement[]>([]);
const chartInstances = new Map<HTMLElement, echarts.ECharts>();

function pad(value: number) {
  return String(value).padStart(2, "0");
}

function dateText(date: Date) {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

function resetDefaultQuery() {
  Object.keys(query).forEach((key) => delete query[key]);
  if (moduleKey.value !== "export-tasks") {
    const end = new Date();
    const start = new Date(end);
    start.setDate(start.getDate() - 29);
    query.dateRange = [dateText(start), dateText(end)];
    query.granularity = "day";
  }
}

function timezoneValue() {
  const offset = selectedTimezoneOffset.value;
  return `${offset >= 0 ? "+" : "-"}${pad(Math.abs(offset))}:00`;
}

function normalizedQuery(includePage = true) {
  const out: Record<string, string | number> = {};
  if (includePage) {
    out.page = page.value;
    out.pageSize = pageSize.value;
  }
  for (const [key, value] of Object.entries(query)) {
    if (Array.isArray(value) && value.length === 2) {
      const base = key.replace("Range", "");
      out[`${base}From`] = selectedTimeToBackend(value[0], false);
      out[`${base}To`] = selectedTimeToBackend(value[1], true);
    } else if (String(value ?? "").trim()) {
      out[key] = String(value).trim();
    }
  }
  if (moduleKey.value !== "export-tasks") out.timezone = timezoneValue();
  return out;
}

function exportFilters() {
  const filters: Record<string, string> = {};
  for (const [key, value] of Object.entries(normalizedQuery(false)))
    filters[key] = String(value);
  return filters;
}

async function load() {
  if (!config.value || !can(config.value.permission)) return;
  loading.value = true;
  try {
    const result = await config.value.loader(normalizedQuery());
    rows.value = result.list || [];
    summaries.value = result.summaryByCurrency || [];
    charts.value = result.charts || [];
    metricBasis.value = result.metricBasis || {};
    total.value = result.total || 0;
    await nextTick();
    renderCharts();
  } catch {
    ElMessage.error(t("reportCenter.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}

function search() {
  page.value = 1;
  load();
}

function reset() {
  resetDefaultQuery();
  page.value = 1;
  load();
}

function fieldLabel(key: string) {
  return t(`reportCenter.fields.${key}`);
}

function valueLabel(value: string) {
  const key = `reportCenter.values.${value.toLowerCase()}`;
  const translated = t(key);
  return translated === key ? value : translated;
}

function isObject(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function display(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (key.toLowerCase().includes("time") || key.toLowerCase().endsWith("at"))
    return formatBackendTime(String(value));
  if (
    key.toLowerCase().includes("rate") ||
    key.toLowerCase().includes("rtp") ||
    key.toLowerCase().includes("roi") ||
    key.toLowerCase().includes("conversion")
  )
    return `${value}%`;
  if (typeof value === "boolean")
    return value ? t("common.yes") : t("common.no");
  if (typeof value === "object") return JSON.stringify(sanitizeBusinessValue(value));
  return String(value);
}

function isMoneyKey(key: string) {
  return /(amount|bet$|win$|ggr$|ngr$|cost$|fee$|commission$|wagering$|cashflow$|profit$|maximumwin$|usernetresult$|totaldeposit$|totalwithdrawal$|settledamount$|pendingsettlement$)/i.test(
    key,
  );
}

function summaryDisplay(summary: Record<string, unknown>, key: string) {
  const value = summary[key];
  if (value === null || value === undefined || value === "") return "-";
  if (/rate|rtp|roi|conversion/i.test(key)) return `${value}%`;
  if (isMoneyKey(key)) return `${summary.currency || ""} ${value}`.trim();
  return String(value);
}

async function view(row: Record<string, unknown>) {
  if (!config.value.detail) return;
  const key = row[config.value.detailKey || "id"];
  if (key === undefined || key === null || key === "") return;
  detailVisible.value = true;
  detailLoading.value = true;
  activeTab.value = config.value.tabs[0] || "";
  try {
    detail.value = await config.value.detail(
      String(key),
      normalizedQuery(false),
    );
  } catch {
    detailVisible.value = false;
    ElMessage.error(t("reportCenter.messages.detailFailed"));
  } finally {
    detailLoading.value = false;
  }
}

function tabValue(tab: string) {
  return detail.value[tab];
}

function objectEntries(value: unknown) {
  return isObject(value) ? businessEntries(value) : [];
}

function arrayColumns(value: unknown) {
  return businessColumns(value);
}

function setChartElement(element: unknown, index: number) {
  if (element instanceof HTMLElement) chartElements.value[index] = element;
}

function renderCharts() {
  charts.value.forEach((chart, index) => {
    const element = chartElements.value[index];
    if (!element) return;
    const instance = chartInstances.get(element) || echarts.init(element);
    chartInstances.set(element, instance);
    instance.setOption(
      {
        color: [
          "#409eff",
          "#67c23a",
          "#e6a23c",
          "#f56c6c",
          "#7f56d9",
          "#00a6a6",
        ],
        tooltip: { trigger: "axis" },
        legend: { top: 0, textStyle: { color: "#606266" } },
        grid: { left: 20, right: 20, bottom: 20, top: 44, containLabel: true },
        xAxis: { type: "category", data: chart.categories, boundaryGap: false },
        yAxis: {
          type: "value",
          splitLine: { lineStyle: { color: "#ebeef5" } },
        },
        series: chart.series.map((series) => ({
          name: `${fieldLabel(series.name)}${series.currency ? ` (${series.currency})` : ""}`,
          type: "line",
          smooth: true,
          symbolSize: 6,
          data: series.data.map((value) => Number(value)),
        })),
      },
      true,
    );
  });
}

function handleResize() {
  chartInstances.forEach((instance) => instance.resize());
}

async function createExport(format: "CSV" | "XLSX") {
  if (!can("report:export:create")) return;
  try {
    await createReportExport({
      reportType: config.value.type,
      filters: exportFilters(),
      fileFormat: format,
      idempotencyKey: `${config.value.type}-${format}-${Date.now()}`,
    });
    ElMessage.success(t("reportCenter.messages.exportCreated"));
  } catch {
    ElMessage.error(t("reportCenter.messages.exportFailed"));
  }
}

async function askReason(
  action: "retry" | "cancel",
  row: Record<string, unknown>,
) {
  try {
    const result = await ElMessageBox.prompt(
      t("reportCenter.messages.reasonPrompt"),
      t(`reportCenter.actions.${action}`),
      {
        confirmButtonText: t("common.confirm"),
        cancelButtonText: t("common.cancel"),
        inputValidator: (value) =>
          Boolean(value.trim()) || t("reportCenter.validation.reasonRequired"),
      },
    );
    if (action === "retry")
      await retryReportExport(Number(row.id), result.value);
    else await cancelReportExport(Number(row.id), result.value);
    ElMessage.success(t("reportCenter.messages.operationSucceeded"));
    load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(t("reportCenter.messages.operationFailed"));
  }
}

async function download(row: Record<string, unknown>) {
  try {
    await downloadReportExport(Number(row.id), String(row.fileFormat || "csv"));
  } catch {
    ElMessage.error(t("reportCenter.messages.downloadFailed"));
  }
}

watch(moduleKey, async () => {
  resetDefaultQuery();
  rows.value = [];
  summaries.value = [];
  charts.value = [];
  page.value = 1;
  await nextTick();
  load();
});

watch(selectedTimezoneOffset, () => {
  if (moduleKey.value !== "export-tasks") load();
});

onMounted(async () => {
  resetDefaultQuery();
  await loadPermissions();
  await load();
  window.addEventListener("resize", handleResize);
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
  chartInstances.forEach((instance) => instance.dispose());
  chartInstances.clear();
});
</script>

<template>
  <div v-if="config" class="report-page">
    <UserCenterCard :title="t('reportCenter.sections.filters')">
      <el-form
        class="search-form"
        label-position="left"
        label-width="132px"
        @submit.prevent="search"
      >
        <el-row :gutter="16">
          <el-col
            v-for="filter in visibleFilters"
            :key="filter.key"
            :xs="24"
            :sm="12"
            :lg="8"
            :xl="6"
          >
            <el-form-item :label="fieldLabel(filter.key)">
              <el-date-picker
                v-if="filter.type === 'dateRange'"
                v-model="query[filter.key]"
                type="daterange"
                value-format="YYYY-MM-DD"
                :start-placeholder="t('reportCenter.placeholders.startDate')"
                :end-placeholder="t('reportCenter.placeholders.endDate')"
              />
              <el-select
                v-else-if="filter.type === 'select'"
                v-model="query[filter.key]"
                clearable
              >
                <el-option
                  v-for="option in filter.options"
                  :key="option"
                  :label="valueLabel(option)"
                  :value="option"
                />
              </el-select>
              <el-input
                v-else
                v-model="query[filter.key]"
                clearable
                :placeholder="t('reportCenter.placeholders.enter')"
                @keyup.enter="search"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <div class="form-actions">
          <el-button @click="reset">{{ t("common.reset") }}</el-button>
          <el-button type="primary" native-type="submit">{{
            t("common.search")
          }}</el-button>
        </div>
      </el-form>
    </UserCenterCard>

    <div v-if="summaries.length" class="summary-groups">
      <UserCenterCard
        v-for="summary in summaries"
        :key="String(summary.currency || 'summary')"
        :title="`${t('reportCenter.sections.summary')} · ${summary.currency || t('reportCenter.values.all')}`"
      >
        <div class="summary-grid">
          <div v-for="key in config.summaries" :key="key" class="summary-card">
            <span>{{ fieldLabel(key) }}</span>
            <strong>{{ summaryDisplay(summary, key) }}</strong>
          </div>
        </div>
      </UserCenterCard>
    </div>

    <UserCenterCard
      v-if="charts.length"
      :title="t('reportCenter.sections.charts')"
    >
      <div class="chart-grid">
        <section
          v-for="(chart, index) in charts"
          :key="chart.key"
          class="chart-panel"
        >
          <h3>{{ t(`reportCenter.charts.${chart.key}`) }}</h3>
          <div
            :ref="(element) => setChartElement(element, index)"
            class="chart-canvas"
          />
        </section>
      </div>
    </UserCenterCard>

    <UserCenterCard
      :title="
        t(
          `menu.reports.${moduleKey === 'financial-report' ? 'financialReport' : moduleKey === 'game-report' ? 'gameReport' : moduleKey === 'provider-report' ? 'providerReport' : moduleKey === 'user-report' ? 'userReport' : moduleKey === 'channel-report' ? 'channelReport' : moduleKey === 'agent-report' ? 'agentReport' : moduleKey === 'promotion-report' ? 'promotionReport' : 'exportTasks'}`,
        )
      "
    >
      <template #extra>
        <el-dropdown
          v-if="moduleKey !== 'export-tasks' && can('report:export:create')"
          @command="createExport"
        >
          <el-button type="primary">{{
            t("reportCenter.actions.export")
          }}</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="CSV">CSV</el-dropdown-item>
              <el-dropdown-item command="XLSX">XLSX</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
      <el-table v-loading="loading" :data="rows" row-key="id" border stripe>
        <el-table-column
          fixed="left"
          type="index"
          :index="(index: number) => (page - 1) * pageSize + index + 1"
          :label="t('common.index')"
          width="76"
        />
        <el-table-column
          v-for="column in visibleColumns"
          :key="column.key"
          :prop="column.key"
          :label="fieldLabel(column.key)"
          :width="column.width"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <UserStatusTag
              v-if="column.status"
              :status="String(row[column.key] || 'UNKNOWN')"
            />
            <span v-else-if="column.time">{{
              formatBackendTime(String(row[column.key] || ""))
            }}</span>
            <span v-else-if="column.percentage">{{
              row[column.key] === null || row[column.key] === undefined
                ? "-"
                : `${row[column.key]}%`
            }}</span>
            <span v-else-if="column.money"
              >{{ row.currency || "" }} {{ row[column.key] ?? "0.00" }}</span
            >
            <span v-else>{{ row[column.key] ?? "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column
          v-if="config.detail || moduleKey === 'export-tasks'"
          fixed="right"
          :label="t('common.actions')"
          width="210"
        >
          <template #default="{ row }">
            <el-button
              v-if="config.detail"
              link
              type="primary"
              @click="view(row)"
              >{{ t("common.view") }}</el-button
            >
            <template v-if="moduleKey === 'export-tasks'">
              <el-button
                v-if="row.status === 'SUCCESS' && can('report:export:download')"
                link
                type="success"
                @click="download(row)"
                >{{ t("reportCenter.actions.download") }}</el-button
              >
              <el-dropdown
                v-if="
                  (row.status === 'FAILED' && can('report:export:retry')) ||
                  (['PENDING', 'PROCESSING'].includes(String(row.status)) &&
                    can('report:export:cancel'))
                "
              >
                <el-button link>{{ t("reportCenter.actions.more") }}</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-if="
                        row.status === 'FAILED' && can('report:export:retry')
                      "
                      @click="askReason('retry', row)"
                      >{{ t("reportCenter.actions.retry") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="
                        ['PENDING', 'PROCESSING'].includes(
                          String(row.status),
                        ) && can('report:export:cancel')
                      "
                      @click="askReason('cancel', row)"
                      >{{ t("reportCenter.actions.cancel") }}</el-dropdown-item
                    >
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </UserCenterCard>

    <el-dialog
      v-model="detailVisible"
      :title="t('reportCenter.sections.details')"
      width="min(980px, 94vw)"
      destroy-on-close
    >
      <div v-loading="detailLoading">
        <el-tabs v-model="activeTab" type="card">
          <el-tab-pane
            v-for="tab in config.tabs"
            :key="tab"
            :name="tab"
            :label="t(`reportCenter.tabs.${tab}`)"
          >
            <el-table
              v-if="Array.isArray(tabValue(tab))"
              :data="tabValue(tab) as Array<Record<string, unknown>>"
              max-height="460"
              border
            >
              <el-table-column
                v-for="key in arrayColumns(tabValue(tab))"
                :key="key"
                :prop="key"
                :label="fieldLabel(key)"
                min-width="140"
                show-overflow-tooltip
              >
                <template #default="{ row }">{{
                  display(key, row[key])
                }}</template>
              </el-table-column>
            </el-table>
            <el-descriptions
              v-else-if="isObject(tabValue(tab))"
              :column="2"
              border
            >
              <el-descriptions-item
                v-for="[key, value] in objectEntries(tabValue(tab))"
                :key="key"
                :label="fieldLabel(key)"
                >{{ display(key, value) }}</el-descriptions-item
              >
            </el-descriptions>
            <el-empty v-else :description="t('reportCenter.messages.noData')" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.report-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.search-form :deep(.el-select),
.search-form :deep(.el-date-editor) {
  width: 100%;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
.summary-groups {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(168px, 1fr));
  gap: 12px;
}
.summary-card {
  min-height: 84px;
  padding: 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  background: var(--el-fill-color-extra-light);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
}
.summary-card span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.summary-card strong {
  color: var(--el-text-color-primary);
  font-size: 20px;
  font-weight: 600;
}
.chart-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 16px;
}
.chart-panel {
  min-width: 0;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  padding: 14px;
}
.chart-panel h3 {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}
.chart-canvas {
  width: 100%;
  height: 300px;
}
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  padding-top: 18px;
}
@media (max-width: 760px) {
  .chart-grid {
    grid-template-columns: 1fr;
  }
  .chart-canvas {
    height: 240px;
  }
}
</style>
