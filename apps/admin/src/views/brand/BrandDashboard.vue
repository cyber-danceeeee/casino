<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import {
  Brush,
  Grid,
  Monitor,
  Promotion,
  Search,
  UploadFilled,
} from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import {
  dashboardApi,
  type ChartResponse,
  type DashboardFilter,
  type MetricCardData,
} from "@/api/adminDashboard";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import { operationsAPI, type OpsRecord } from "@/api/adminStability";
import DashboardMetricCard from "@/components/dashboard/DashboardMetricCard.vue";
import TrendChart from "@/components/dashboard/TrendChart.vue";
import BrandSetupProgress from "@/components/brand/BrandSetupProgress.vue";
import H5PreviewFrame from "@/components/brand/H5PreviewFrame.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { timezoneOffsetForIANA } from "@/utils/timezone";
import { statusKey, statusType } from "@/utils/presentation";

const router = useRouter();
const { t } = useI18n();
const { hasNavigation } = useAdminPermissions();
const loading = ref(false);
const overview = ref<BrandOverview | null>(null);
const metrics = ref<MetricCardData[]>([]);
const financialTrend = ref<ChartResponse>();
const playerTrend = ref<ChartResponse>();
const providers = ref<OpsRecord[]>([]);
const paymentHealth = ref<OpsRecord[]>([]);
const openCases = ref(0);
const criticalAlerts = ref(0);
const failedJobs = ref(0);
const previewVisible = ref(false);
const range = ref<"today" | "yesterday" | "last7" | "last30" | "custom">("today");
const customRange = ref<[string, string] | null>(null);
const loadErrors = reactive<string[]>([]);

const metricOrder = [
  "depositCount",
  "depositAmount",
  "withdrawalCount",
  "withdrawalAmount",
  "betCount",
  "totalBet",
  "ggr",
  "onlineUsers",
  "registeredUsers",
  "depositSuccessRate",
  "withdrawalSuccessRate",
  "pendingWithdrawalCount",
];
const topMetrics = computed(
  () =>
    metricOrder
      .map((key) => metrics.value.find((metric) => metric.key === key))
      .filter(Boolean) as MetricCardData[],
);

function offsetText(timezone: string) {
  const offset = timezoneOffsetForIANA(timezone) ?? 0;
  const sign = offset >= 0 ? "+" : "-";
  const absolute = Math.abs(offset);
  const hours = Math.floor(absolute);
  const minutes = Math.round((absolute - hours) * 60);
  return `${sign}${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}

function dateText(date: Date, end = false) {
  const pad = (value: number) => String(value).padStart(2, "0");
  return `${date.getUTCFullYear()}-${pad(date.getUTCMonth() + 1)}-${pad(date.getUTCDate())} ${end ? "23:59:59" : "00:00:00"}`;
}

function siteToday() {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: overview.value?.timezone || "UTC",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(new Date());
  const value = Object.fromEntries(parts.map((part) => [part.type, part.value]));
  return new Date(Date.UTC(Number(value.year), Number(value.month) - 1, Number(value.day)));
}

function dashboardFilter(): DashboardFilter {
  if (range.value === "custom" && customRange.value?.length === 2) {
    return {
      startTime: `${customRange.value[0]} 00:00:00`,
      endTime: `${customRange.value[1]} 23:59:59`,
      timezone: offsetText(overview.value?.timezone || "UTC"),
      granularity: "day",
      country: overview.value?.countryCode,
      currency: overview.value?.currencyCode || "PKR",
    };
  }
  const now = siteToday();
  const start = new Date(now);
  const end = new Date(now);
  if (range.value === "yesterday") {
    start.setUTCDate(start.getUTCDate() - 1);
    end.setUTCDate(end.getUTCDate() - 1);
  }
  if (range.value === "last7") start.setUTCDate(start.getUTCDate() - 6);
  if (range.value === "last30") start.setUTCDate(start.getUTCDate() - 29);
  return {
    startTime: dateText(start),
    endTime: dateText(end, true),
    timezone: offsetText(overview.value?.timezone || "UTC"),
    granularity:
      range.value === "today" || range.value === "yesterday" ? "hour" : "day",
    country: overview.value?.countryCode,
    currency: overview.value?.currencyCode || "PKR",
  };
}

function changeRange() {
  if (range.value !== "custom" || customRange.value) void loadDashboardData();
}

function rows(value: unknown) {
  return Array.isArray(value) ? (value as OpsRecord[]) : [];
}

function monitorRows(value: unknown) {
  if (!value || typeof value !== "object") return [];
  const record = value as Record<string, unknown>;
  return rows(record.list).length ? rows(record.list) : rows(record.health);
}

async function loadDashboardData() {
  const filter = dashboardFilter();
  const results = await Promise.allSettled([
    dashboardApi.overview(filter),
    dashboardApi.chart("/financial-trend", filter),
    dashboardApi.chart("/betting-trend", filter),
    dashboardApi.chart("/user-trend", filter),
    operationsAPI.monitor.payments({
      from: filter.startTime,
      to: filter.endTime,
    }),
    operationsAPI.monitor.providers({
      from: filter.startTime,
      to: filter.endTime,
    }),
    operationsAPI.support.cases({ status: "OPEN", page: 1, pageSize: 1 }),
    operationsAPI.alerts.list({
      severity: "CRITICAL",
      status: "OPEN",
      page: 1,
      pageSize: 1,
    }),
    operationsAPI.jobs.overview({}),
  ]);
  loadErrors.splice(0);
  const capture = <T,>(index: number, apply: (value: T) => void) => {
    const result = results[index];
    if (result.status === "fulfilled") apply(result.value as T);
    else
      loadErrors.push(
        result.reason instanceof Error
          ? result.reason.message
          : String(result.reason),
      );
  };
  capture<{ cards: MetricCardData[] }>(0, (value) => {
    metrics.value = value.cards || [];
  });
  let finance: ChartResponse | undefined;
  let betting: ChartResponse | undefined;
  capture<ChartResponse>(1, (value) => {
    finance = value;
  });
  capture<ChartResponse>(2, (value) => {
    betting = value;
  });
  if (finance) {
    const bettingByTime = new Map(
      (betting?.rows || []).map((row) => [row.time, row]),
    );
    financialTrend.value = {
      ...finance,
      metrics: [
        "depositAmount",
        "withdrawalAmount",
        "totalBet",
        "totalWin",
        "ggr",
      ],
      rows: finance.rows.map((row) => ({
        ...row,
        ...(bettingByTime.get(row.time) || {}),
      })),
    };
  }
  capture<ChartResponse>(3, (value) => {
    playerTrend.value = value;
  });
  capture<Record<string, unknown>>(4, (value) => {
    paymentHealth.value = monitorRows(value);
  });
  capture<Record<string, unknown>>(5, (value) => {
    providers.value = monitorRows(value);
  });
  capture<{ total: number }>(6, (value) => {
    openCases.value = Number(value.total || 0);
  });
  capture<{ total: number }>(7, (value) => {
    criticalAlerts.value = Number(value.total || 0);
  });
  capture<Record<string, unknown>>(8, (value) => {
    failedJobs.value =
      rows(value.failedExecutions).length +
      rows(value.retryRecords).filter((item) => statusKey(item.status) === "FAILED").length +
      rows(value.deadLetters).length;
  });
}

async function load() {
  loading.value = true;
  try {
    overview.value = await systemAPI.brand.overview();
    await loadDashboardData();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("siteExperience.common.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}

const operationalCards = computed(() => [
  {
    key: "paymentHealth",
    value:
      paymentHealth.value[0]?.operationalStatus ||
      paymentHealth.value[0]?.healthStatus ||
      "UNKNOWN",
    path: "/ops/payments",
  },
  {
    key: "providerHealth",
    value: providers.value.some((row) =>
      ["UNAVAILABLE", "UNHEALTHY"].includes(
        statusKey(row.operationalStatus || row.healthStatus),
      ),
    )
      ? "UNHEALTHY"
      : providers.value.length
        ? "HEALTHY"
        : "UNKNOWN",
    path: "/ops/providers",
  },
  {
    key: "maintenance",
    value: overview.value?.maintenanceEnabled ? "MAINTENANCE" : "ONLINE",
    path: "/ops/maintenance",
  },
  {
    key: "site",
    value: overview.value?.primaryDomain ? "ONLINE" : "UNAVAILABLE",
    path: "/system/brand",
  },
  {
    key: "publishedVersion",
    value: overview.value?.publishedVersion
      ? `v${overview.value.publishedVersion}`
      : "-",
    path: "/brand/publish",
  },
  {
    key: "unpublished",
    value: overview.value?.unpublishedChanges || 0,
    path: "/brand/publish",
  },
  { key: "criticalAlerts", value: criticalAlerts.value, path: "/ops/alerts" },
  { key: "failedJobs", value: failedJobs.value, path: "/ops/jobs" },
  { key: "openCases", value: openCases.value, path: "/support/cases" },
]);

const pendingWork = computed(() =>
  [
    {
      key: "withdrawals",
      count: Number(
        metrics.value.find((item) => item.key === "pendingWithdrawalCount")
          ?.value || 0,
      ),
      path: "/funds/withdrawals",
    },
    { key: "supportCases", count: openCases.value, path: "/support/cases" },
    {
      key: "configuration",
      count: overview.value?.unpublishedChanges || 0,
      path: "/brand/publish",
    },
    { key: "alerts", count: criticalAlerts.value, path: "/ops/alerts" },
    { key: "failedJobs", count: failedJobs.value, path: "/ops/jobs" },
  ].filter((item) => item.count > 0),
);

const quickActions = [
  {
    key: "searchPlayer",
    path: "/support/players",
    navigation: "/support/players",
    icon: Search,
  },
  {
    key: "searchOrder",
    path: "/support/orders",
    navigation: "/support/orders",
    icon: Search,
  },
  {
    key: "banner",
    path: "/operations/banners",
    navigation: "/operations/banners",
    icon: Promotion,
  },
  {
    key: "popup",
    path: "/operations/popups",
    navigation: "/operations/popups",
    icon: Promotion,
  },
  {
    key: "games",
    path: "/games/operations",
    navigation: "/games/list",
    icon: Grid,
  },
  {
    key: "maintenance",
    path: "/ops/maintenance",
    navigation: "/ops/maintenance",
    icon: Brush,
  },
  {
    key: "preview",
    path: "/brand/preview",
    navigation: "/system/brand",
    icon: Monitor,
  },
  {
    key: "publish",
    path: "/brand/publish",
    navigation: "/system/brand",
    icon: UploadFilled,
  },
].filter((item) => hasNavigation(item.navigation));

onMounted(load);
</script>

<template>
  <div v-loading="loading" class="site-dashboard">
    <header class="page-header">
      <div>
        <p>{{ t("siteExperience.dashboard.eyebrow") }}</p>
        <h1>
          {{ overview?.brandName || t("siteExperience.dashboard.title") }}
        </h1>
        <span>{{ overview?.timezone }} · {{ overview?.currencyCode }}</span>
      </div>
      <div class="range-control">
        <el-radio-group v-model="range" size="small" @change="changeRange">
          <el-radio-button value="today">{{
            t("siteExperience.ranges.today")
          }}</el-radio-button
          ><el-radio-button value="yesterday">{{
            t("siteExperience.ranges.yesterday")
          }}</el-radio-button
          ><el-radio-button value="last7">{{
            t("siteExperience.ranges.last7")
          }}</el-radio-button
          ><el-radio-button value="last30">{{
            t("siteExperience.ranges.last30")
          }}</el-radio-button
          ><el-radio-button value="custom">{{
            t("siteExperience.ranges.custom")
          }}</el-radio-button>
        </el-radio-group>
        <el-date-picker
          v-if="range === 'custom'"
          v-model="customRange"
          type="daterange"
          value-format="YYYY-MM-DD"
          size="small"
          @change="loadDashboardData"
        />
      </div>
    </header>

    <el-alert
      v-if="loadErrors.length"
      type="warning"
      :title="t('siteExperience.dashboard.partialData')"
      :description="loadErrors[0]"
      show-icon
      :closable="false"
    />

    <section class="metric-grid">
      <DashboardMetricCard
        v-for="metric in topMetrics"
        :key="metric.key"
        :metric="metric"
      />
    </section>

    <BrandSetupProgress v-if="overview" :overview="overview" />

    <section class="section-block">
      <header>
        <div>
          <h2>{{ t("siteExperience.dashboard.operationalOverview") }}</h2>
          <p>{{ t("siteExperience.dashboard.operationalDescription") }}</p>
        </div>
      </header>
      <div class="ops-grid">
        <button
          v-for="item in operationalCards"
          :key="item.key"
          type="button"
          @click="router.push(item.path)"
        >
          <span>{{
            t(`siteExperience.dashboard.operations.${item.key}`)
          }}</span
          ><el-tag
            v-if="
              typeof item.value === 'string' &&
              !String(item.value).startsWith('v')
            "
            :type="statusType(item.value)"
            effect="plain"
            >{{
              t(
                `siteExperience.status.${statusKey(item.value)}`,
                String(item.value),
              )
            }}</el-tag
          ><strong v-else>{{ item.value }}</strong>
        </button>
      </div>
    </section>

    <section class="chart-grid">
      <article class="section-block">
        <header>
          <div>
            <h2>{{ t("siteExperience.dashboard.financialTrend") }}</h2>
            <p>{{ overview?.currencyCode }}</p>
          </div>
        </header>
        <TrendChart :data="financialTrend" :loading="loading" />
      </article>
      <article class="section-block">
        <header>
          <div>
            <h2>{{ t("siteExperience.dashboard.playerTrend") }}</h2>
            <p>{{ t("siteExperience.dashboard.playerTrendDescription") }}</p>
          </div>
        </header>
        <TrendChart :data="playerTrend" :loading="loading" />
      </article>
    </section>

    <section class="two-column">
      <article class="section-block">
        <header>
          <div>
            <h2>{{ t("siteExperience.dashboard.providerHealth") }}</h2>
            <p>{{ t("siteExperience.dashboard.providerDescription") }}</p>
          </div>
          <el-button
            link
            type="primary"
            @click="router.push('/ops/providers')"
            >{{ t("common.view") }}</el-button
          >
        </header>
        <el-table :data="providers.slice(0, 6)" size="small"
          ><el-table-column
            :label="t('siteExperience.fields.provider')"
            min-width="160"
            ><template #default="{ row }"
              ><strong>{{
                row.providerName || row.name || row.providerCode
              }}</strong
              ><small>{{ row.providerCode }}</small></template
            ></el-table-column
          ><el-table-column
            :label="t('siteExperience.fields.enabledGames')"
            prop="enabledGames"
            width="120" /><el-table-column
            :label="t('siteExperience.fields.status')"
            width="125"
            ><template #default="{ row }"
              ><el-tag
                :type="statusType(row.operationalStatus || row.healthStatus)"
                effect="plain"
                >{{
                  t(
                    `siteExperience.status.${statusKey(row.operationalStatus || row.healthStatus)}`,
                    String(row.operationalStatus || row.healthStatus || "-"),
                  )
                }}</el-tag
              ></template
            ></el-table-column
          ><template #empty
            ><el-empty
              :description="t('siteExperience.empty.provider')" /></template
        ></el-table>
      </article>

      <article class="section-block">
        <header>
          <div>
            <h2>{{ t("siteExperience.dashboard.pendingWork") }}</h2>
            <p>{{ t("siteExperience.dashboard.pendingDescription") }}</p>
          </div>
        </header>
        <div v-if="pendingWork.length" class="pending-list">
          <button
            v-for="item in pendingWork"
            :key="item.key"
            type="button"
            @click="router.push(item.path)"
          >
            <span>{{
              t(`siteExperience.dashboard.pending.${item.key}`)
            }}</span
            ><el-tag type="warning" round>{{ item.count }}</el-tag>
          </button>
        </div>
        <el-empty v-else :description="t('siteExperience.empty.pending')" />
      </article>
    </section>

    <section class="quick-actions">
      <h2>{{ t("siteExperience.dashboard.quickActions") }}</h2>
      <el-button
        v-for="item in quickActions"
        :key="item.key"
        :icon="item.icon"
        @click="router.push(item.path)"
        >{{ t(`siteExperience.dashboard.quick.${item.key}`) }}</el-button
      >
    </section>

    <el-dialog
      v-model="previewVisible"
      :title="t('brandPortal.dialogs.previewTitle')"
      width="min(920px, 94vw)"
      destroy-on-close
      ><H5PreviewFrame
        :h5-domain="overview?.h5Domain"
        :site-i-d="overview?.siteID"
    /></el-dialog>
  </div>
</template>

<style scoped>
.site-dashboard {
  display: grid;
  gap: 16px;
}
.page-header,
.section-block,
.quick-actions {
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  padding: 20px;
}
.page-header p {
  margin: 0;
  color: var(--el-color-primary);
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}
.page-header h1 {
  margin: 5px 0;
  font-size: 26px;
  letter-spacing: 0;
}
.page-header span,
header p {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.range-control {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}
.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}
.section-block {
  min-width: 0;
  padding: 17px;
}
.section-block > header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}
.section-block h2,
.quick-actions h2 {
  margin: 0;
  font-size: 17px;
  letter-spacing: 0;
}
.section-block header p {
  margin: 5px 0 0;
}
.ops-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border: 1px solid var(--el-border-color-lighter);
}
.ops-grid button,
.pending-list button {
  border: 0;
  background: transparent;
  text-align: left;
  cursor: pointer;
}
.ops-grid button {
  display: flex;
  min-height: 68px;
  padding: 13px;
  flex-direction: column;
  align-items: flex-start;
  justify-content: space-between;
  border-right: 1px solid var(--el-border-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.ops-grid button:hover,
.pending-list button:hover {
  background: var(--el-fill-color-light);
}
.ops-grid span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.ops-grid strong {
  font-size: 18px;
}
.chart-grid,
.two-column {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
.chart-grid .section-block {
  height: 390px;
}
.pending-list {
  display: grid;
}
.pending-list button {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 44px;
  padding: 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.pending-list button:last-child {
  border-bottom: 0;
}
.quick-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 9px;
  padding: 15px;
}
.quick-actions h2 {
  margin-right: 8px;
}
.section-block small {
  display: block;
  margin-top: 3px;
  color: var(--el-text-color-secondary);
}
@media (max-width: 1100px) {
  .metric-grid,
  .ops-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 760px) {
  .page-header {
    flex-direction: column;
  }
  .metric-grid,
  .ops-grid,
  .chart-grid,
  .two-column {
    grid-template-columns: 1fr;
  }
  .chart-grid .section-block {
    height: 340px;
  }
}
</style>
