<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Download, Plus, Refresh, Search, View } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import {
  operationsAPI,
  operationIdempotencyKey,
  type OpsQuery,
  type OpsRecord,
} from "@/api/adminStability";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import {
  businessColumns,
  businessEntries,
  humanizeBusinessKey,
  sanitizeBusinessValue,
  statusKey,
  statusType,
} from "@/utils/presentation";
import { formatBackendTime } from "@/utils/timezone";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const { can, loadPermissions } = useAdminPermissions();
const mode = computed(() =>
  String(route.meta.moduleKey || "support-player-search"),
);
const loading = ref(false);
const rows = ref<OpsRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<Record<string, unknown>>({});
const detailTitle = ref("");
const createVisible = ref(false);
const query = reactive<Record<string, string>>({
  searchType: "PLAYER_NUMBER",
  keyword: "",
  playerNo: "",
  status: "",
  category: "",
  priority: "",
  from: "",
  to: "",
});
const caseForm = reactive({
  playerNo: "",
  category: "ACCOUNT",
  priority: "NORMAL",
  title: "",
  description: "",
  reason: "",
});

const isPlayerSearch = computed(
  () =>
    mode.value === "support-player-search" ||
    mode.value === "player-business-detail",
);
const isTimeline = computed(() => mode.value === "support-player-timeline");
const isOrders = computed(() => mode.value === "support-order-search");
const isCases = computed(() => mode.value === "support-cases");
const searchTypes = [
  "PLAYER_NUMBER",
  "PHONE",
  "EMAIL",
  "USERNAME",
  "DEPOSIT_ORDER",
  "WITHDRAWAL_ORDER",
  "GAME_ROUND_ID",
  "PROVIDER_TRANSACTION_ID",
];
const detailTabs = computed(() =>
  Object.entries(detail.value).filter(
    ([key, value]) =>
      !["overview", "summary", "order", "case"].includes(key) &&
      (Array.isArray(value) || (value && typeof value === "object")),
  ),
);
const overviewEntries = computed(() =>
  businessEntries(
    detail.value.overview || detail.value.order || detail.value.case,
  ),
);

function persistKey() {
  return `tenant-support:${mode.value}`;
}
function restoreFilters() {
  try {
    Object.assign(
      query,
      JSON.parse(sessionStorage.getItem(persistKey()) || "{}"),
    );
  } catch {
    sessionStorage.removeItem(persistKey());
  }
}
function normalizedQuery() {
  const result: OpsQuery = {
    page: page.value,
    pageSize: pageSize.value,
  };
  if (query.keyword) result.query = query.keyword;
  if (query.searchType) result.searchType = query.searchType;
  if (query.playerNo) result.playerNo = query.playerNo;
  if (query.status) result.status = query.status;
  if (query.category) result.category = query.category;
  if (query.priority) result.priority = query.priority;
  if (query.from) result.from = query.from;
  if (query.to) result.to = query.to;
  sessionStorage.setItem(persistKey(), JSON.stringify(query));
  return result;
}

async function load() {
  loading.value = true;
  try {
    let result: { list: OpsRecord[]; total: number };
    if (isPlayerSearch.value) {
      if (!query.keyword) {
        rows.value = [];
        total.value = 0;
        return;
      }
      result = await operationsAPI.support.players(normalizedQuery());
    } else if (isTimeline.value) {
      if (!query.playerNo) {
        rows.value = [];
        total.value = 0;
        return;
      }
      result = await operationsAPI.support.timelineByNumber(
        query.playerNo,
        normalizedQuery(),
      );
    } else if (isOrders.value) {
      if (!query.keyword && !query.playerNo) {
        rows.value = [];
        total.value = 0;
        return;
      }
      result = await operationsAPI.support.orders(normalizedQuery());
    } else if (isCases.value)
      result = await operationsAPI.support.cases(normalizedQuery());
    else result = await operationsAPI.support.exports(normalizedQuery());
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    rows.value = [];
    total.value = 0;
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("tenantExperience.common.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}
function reset() {
  Object.assign(query, {
    searchType: "PLAYER_NUMBER",
    keyword: "",
    playerNo: "",
    status: "",
    category: "",
    priority: "",
    from: "",
    to: "",
  });
  page.value = 1;
  sessionStorage.removeItem(persistKey());
  void load();
}
async function openPlayer(row: OpsRecord) {
  const playerNo = String(row.userNo || row.playerNo || "");
  if (!playerNo) return;
  await router.push(`/players/${encodeURIComponent(playerNo)}`);
  detailTitle.value = `${t("tenantExperience.support.playerNumber")}: ${playerNo}`;
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detail.value = await operationsAPI.support.playerByNumber(playerNo);
  } catch (error) {
    detailVisible.value = false;
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("tenantExperience.common.loadFailed"),
    );
  } finally {
    detailLoading.value = false;
  }
}
async function openOrder(row: OpsRecord) {
  detailTitle.value = String(
    row.orderNo ||
      row.externalOrderNo ||
      t("tenantExperience.support.orderDetail"),
  );
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detail.value = await operationsAPI.support.orderDetail(
      String(row.orderType),
      Number(row.id),
    );
  } catch (error) {
    detailVisible.value = false;
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("tenantExperience.common.loadFailed"),
    );
  } finally {
    detailLoading.value = false;
  }
}
async function openCase(row: OpsRecord) {
  detailTitle.value = String(
    row.case_no || row.caseNo || t("tenantExperience.support.caseDetail"),
  );
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detail.value = await operationsAPI.support.caseDetail(Number(row.id));
  } catch (error) {
    detailVisible.value = false;
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("tenantExperience.common.loadFailed"),
    );
  } finally {
    detailLoading.value = false;
  }
}
async function createCase() {
  if (
    !caseForm.title.trim() ||
    !caseForm.description.trim() ||
    !caseForm.reason.trim()
  ) {
    ElMessage.warning(t("tenantExperience.validation.required"));
    return;
  }
  try {
    await ElMessageBox.confirm(
      t("tenantExperience.support.caseConfirm", {
        player: caseForm.playerNo || t("tenantExperience.support.noPlayer"),
      }),
      t("tenantExperience.support.createCase"),
      { type: "warning" },
    );
    await operationsAPI.support.createCase({
      ...caseForm,
      status: "OPEN",
      idempotencyKey: operationIdempotencyKey("support-case"),
    });
    createVisible.value = false;
    Object.assign(caseForm, {
      playerNo: "",
      category: "ACCOUNT",
      priority: "NORMAL",
      title: "",
      description: "",
      reason: "",
    });
    ElMessage.success(t("tenantExperience.support.caseCreated"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("tenantExperience.common.actionFailed"),
      );
  }
}
async function download(row: OpsRecord) {
  try {
    await operationsAPI.support.downloadExport(Number(row.id));
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("tenantExperience.common.actionFailed"),
    );
  }
}
function valueText(value: unknown) {
  if (value == null || value === "") return "-";
  if (typeof value === "object") return JSON.stringify(sanitizeBusinessValue(value));
  return String(value);
}
function emptyDescription() {
  if (isPlayerSearch.value) return t("tenantExperience.empty.players");
  if (isTimeline.value) return t("tenantExperience.empty.timeline");
  if (isOrders.value) return t("tenantExperience.empty.orders");
  if (isCases.value) return t("tenantExperience.empty.cases");
  return t("tenantExperience.empty.exports");
}

watch(
  () => route.fullPath,
  async () => {
    restoreFilters();
    page.value = 1;
    if (route.params.playerNo) {
      query.keyword = String(route.params.playerNo);
      query.searchType = "PLAYER_NUMBER";
      await load();
      if (rows.value[0]) await openPlayer(rows.value[0]);
    } else await load();
  },
);
onMounted(async () => {
  await loadPermissions();
  restoreFilters();
  if (route.params.playerNo) {
    query.keyword = String(route.params.playerNo);
    query.searchType = "PLAYER_NUMBER";
  }
  await load();
  if (route.params.playerNo && rows.value[0]) await openPlayer(rows.value[0]);
});
</script>

<template>
  <div class="support-page">
    <header class="page-header">
      <div>
        <p>{{ t("tenantExperience.support.eyebrow") }}</p>
        <h1>{{ t(`tenantExperience.support.pages.${mode}`) }}</h1>
        <span>{{ t(`tenantExperience.support.descriptions.${mode}`) }}</span>
      </div>
      <div>
        <el-button :icon="Refresh" @click="load">{{
          t("common.refresh")
        }}</el-button
        ><el-button
          v-if="isCases && can('support:case:write')"
          type="primary"
          :icon="Plus"
          @click="createVisible = true"
          >{{ t("tenantExperience.support.createCase") }}</el-button
        >
      </div>
    </header>
    <section class="filters">
      <el-form inline>
        <template v-if="isPlayerSearch"
          ><el-form-item :label="t('tenantExperience.support.searchBy')"
            ><el-select v-model="query.searchType" style="width: 210px"
              ><el-option
                v-for="item in searchTypes"
                :key="item"
                :label="t(`tenantExperience.support.searchTypes.${item}`)"
                :value="item" /></el-select></el-form-item
          ><el-form-item :label="t('tenantExperience.support.searchValue')"
            ><el-input
              v-model="query.keyword"
              clearable
              style="width: 280px"
              @keyup.enter="
                page = 1;
                load();
              " /></el-form-item
        ></template>
        <template v-else-if="isTimeline"
          ><el-form-item :label="t('tenantExperience.support.playerNumber')"
            ><el-input v-model="query.playerNo" clearable /></el-form-item
        ></template>
        <template v-else-if="isOrders"
          ><el-form-item :label="t('tenantExperience.support.orderNumber')"
            ><el-input v-model="query.keyword" clearable /></el-form-item
          ><el-form-item :label="t('tenantExperience.support.playerNumber')"
            ><el-input v-model="query.playerNo" clearable /></el-form-item
        ></template>
        <template v-else-if="isCases"
          ><el-form-item :label="t('tenantExperience.support.caseNumber')"
            ><el-input v-model="query.keyword" clearable /></el-form-item
          ><el-form-item :label="t('tenantExperience.support.playerNumber')"
            ><el-input v-model="query.playerNo" clearable /></el-form-item
          ><el-form-item :label="t('tenantExperience.fields.status')"
            ><el-select v-model="query.status" clearable
              ><el-option
                v-for="item in [
                  'OPEN',
                  'INVESTIGATING',
                  'WAITING_CUSTOMER',
                  'RESOLVED',
                  'CLOSED',
                ]"
                :key="item"
                :label="t(`tenantExperience.status.${item}`)"
                :value="item" /></el-select></el-form-item
        ></template>
        <template v-else
          ><el-form-item :label="t('tenantExperience.fields.status')"
            ><el-select v-model="query.status" clearable
              ><el-option
                v-for="item in ['PENDING', 'RUNNING', 'SUCCESS', 'FAILED']"
                :key="item"
                :label="t(`tenantExperience.status.${item}`)"
                :value="item" /></el-select></el-form-item
        ></template>
        <el-form-item
          ><el-button
            type="primary"
            :icon="Search"
            @click="
              page = 1;
              load();
            "
            >{{ t("common.search") }}</el-button
          ><el-button @click="reset">{{
            t("common.reset")
          }}</el-button></el-form-item
        >
      </el-form>
    </section>
    <section class="table-shell">
      <el-table v-loading="loading" :data="rows" row-key="id">
        <template v-if="isPlayerSearch"
          ><el-table-column
            prop="userNo"
            :label="t('tenantExperience.support.playerNumber')"
            min-width="160" /><el-table-column
            prop="username"
            :label="t('tenantExperience.fields.username')"
            min-width="135" /><el-table-column
            :label="t('tenantExperience.fields.mobile')"
            min-width="150"
            ><template #default="{ row }"
              >{{ row.phoneCountryCode }} {{ row.phone }}</template
            ></el-table-column
          ><el-table-column
            prop="countryCode"
            :label="t('tenantExperience.fields.country')"
            width="100" /><el-table-column
            prop="currency"
            :label="t('tenantExperience.fields.currency')"
            width="95" /><el-table-column
            :label="t('tenantExperience.fields.availableBalance')"
            min-width="135"
            align="right"
            ><template #default="{ row }"
              ><strong>{{ row.availableBalance }}</strong></template
            ></el-table-column
          ><el-table-column
            prop="bonusBalance"
            :label="t('tenantExperience.fields.bonusBalance')"
            min-width="125"
            align="right" /><el-table-column
            prop="vipLevel"
            :label="t('tenantExperience.fields.vipLevel')"
            width="95" /><el-table-column
            :label="t('tenantExperience.fields.riskStatus')"
            width="125"
            ><template #default="{ row }"
              ><el-tag :type="statusType(row.riskStatus)" effect="plain">{{
                t(
                  `tenantExperience.status.${statusKey(row.riskStatus)}`,
                  row.riskStatus,
                )
              }}</el-tag></template
            ></el-table-column
          ><el-table-column
            :label="t('tenantExperience.fields.lastLogin')"
            min-width="170"
            ><template #default="{ row }">{{
              formatBackendTime(row.lastLoginAt)
            }}</template></el-table-column
          ><el-table-column fixed="right" width="80"
            ><template #default="{ row }"
              ><el-button
                circle
                :icon="View"
                @click="openPlayer(row)" /></template></el-table-column
        ></template>
        <template v-else-if="isTimeline"
          ><el-table-column
            :label="t('tenantExperience.fields.event')"
            min-width="180"
            ><template #default="{ row }">{{
              t(
                `tenantExperience.events.${statusKey(row.eventType)}`,
                row.eventType,
              )
            }}</template></el-table-column
          ><el-table-column
            prop="businessOrderNo"
            :label="t('tenantExperience.support.businessNumber')"
            min-width="190"
          /><el-table-column
            prop="amount"
            :label="t('tenantExperience.fields.amount')"
            width="120"
          /><el-table-column
            prop="currency"
            :label="t('tenantExperience.fields.currency')"
            width="90"
          /><el-table-column
            :label="t('tenantExperience.fields.status')"
            width="130"
            ><template #default="{ row }"
              ><el-tag :type="statusType(row.status)" effect="plain">{{
                t(
                  `tenantExperience.status.${statusKey(row.status)}`,
                  row.status,
                )
              }}</el-tag></template
            ></el-table-column
          ><el-table-column
            :label="t('tenantExperience.fields.time')"
            width="175"
            ><template #default="{ row }">{{
              formatBackendTime(row.occurredAt)
            }}</template></el-table-column
          ></template
        >
        <template v-else-if="isOrders"
          ><el-table-column
            prop="orderNo"
            :label="t('tenantExperience.support.orderNumber')"
            min-width="190" /><el-table-column
            prop="externalOrderNo"
            :label="t('tenantExperience.support.upstreamNumber')"
            min-width="190" /><el-table-column
            prop="playerNo"
            :label="t('tenantExperience.support.playerNumber')"
            min-width="150" /><el-table-column
            :label="t('tenantExperience.fields.type')"
            width="135"
            ><template #default="{ row }">{{
              t(
                `tenantExperience.orderTypes.${statusKey(row.orderType)}`,
                row.orderType,
              )
            }}</template></el-table-column
          ><el-table-column
            prop="providerCode"
            :label="t('tenantExperience.fields.provider')"
            width="130" /><el-table-column
            prop="amount"
            :label="t('tenantExperience.fields.amount')"
            width="120" /><el-table-column
            prop="currency"
            :label="t('tenantExperience.fields.currency')"
            width="90" /><el-table-column
            :label="t('tenantExperience.fields.status')"
            width="125"
            ><template #default="{ row }"
              ><el-tag :type="statusType(row.status)" effect="plain">{{
                t(
                  `tenantExperience.status.${statusKey(row.status)}`,
                  row.status,
                )
              }}</el-tag></template
            ></el-table-column
          ><el-table-column
            :label="t('tenantExperience.fields.createdTime')"
            width="175"
            ><template #default="{ row }">{{
              formatBackendTime(row.createdAt)
            }}</template></el-table-column
          ><el-table-column fixed="right" width="80"
            ><template #default="{ row }"
              ><el-button
                circle
                :icon="View"
                @click="openOrder(row)" /></template></el-table-column
        ></template>
        <template v-else-if="isCases"
          ><el-table-column
            prop="case_no"
            :label="t('tenantExperience.support.caseNumber')"
            min-width="175" /><el-table-column
            prop="title"
            :label="t('tenantExperience.fields.title')"
            min-width="200" /><el-table-column
            prop="playerNo"
            :label="t('tenantExperience.support.playerNumber')"
            min-width="150" /><el-table-column
            :label="t('tenantExperience.fields.priority')"
            width="110"
            ><template #default="{ row }"
              ><el-tag :type="statusType(row.priority)" effect="plain">{{
                t(
                  `tenantExperience.priority.${statusKey(row.priority)}`,
                  row.priority,
                )
              }}</el-tag></template
            ></el-table-column
          ><el-table-column
            :label="t('tenantExperience.fields.status')"
            width="140"
            ><template #default="{ row }"
              ><el-tag :type="statusType(row.status)" effect="plain">{{
                t(
                  `tenantExperience.status.${statusKey(row.status)}`,
                  row.status,
                )
              }}</el-tag></template
            ></el-table-column
          ><el-table-column
            prop="assignedAdminUsername"
            :label="t('tenantExperience.fields.assignee')"
            width="140" /><el-table-column
            :label="t('tenantExperience.fields.updatedTime')"
            width="175"
            ><template #default="{ row }">{{
              formatBackendTime(row.updated_at || row.updatedAt)
            }}</template></el-table-column
          ><el-table-column fixed="right" width="80"
            ><template #default="{ row }"
              ><el-button
                circle
                :icon="View"
                @click="openCase(row)" /></template></el-table-column
        ></template>
        <template v-else
          ><el-table-column
            prop="taskNo"
            :label="t('tenantExperience.support.exportNumber')"
            min-width="210" /><el-table-column
            prop="exportType"
            :label="t('tenantExperience.fields.type')"
            min-width="150" /><el-table-column
            prop="rowCount"
            :label="t('tenantExperience.fields.records')"
            width="110" /><el-table-column
            :label="t('tenantExperience.fields.status')"
            width="125"
            ><template #default="{ row }"
              ><el-tag :type="statusType(row.status)" effect="plain">{{
                t(
                  `tenantExperience.status.${statusKey(row.status)}`,
                  row.status,
                )
              }}</el-tag></template
            ></el-table-column
          ><el-table-column
            :label="t('tenantExperience.fields.createdTime')"
            width="175"
            ><template #default="{ row }">{{
              formatBackendTime(row.createdAt)
            }}</template></el-table-column
          ><el-table-column fixed="right" width="80"
            ><template #default="{ row }"
              ><el-button
                v-if="row.status === 'SUCCESS'"
                circle
                :icon="Download"
                @click="download(row)" /></template></el-table-column
        ></template>
        <template #empty
          ><el-empty :description="emptyDescription()"
            ><el-button
              v-if="isPlayerSearch || isOrders"
              type="primary"
              @click="reset"
              >{{ t("tenantExperience.common.clearFilters") }}</el-button
            ></el-empty
          ></template
        > </el-table
      ><el-pagination
        v-if="total"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @change="load"
      />
    </section>

    <el-drawer
      v-model="detailVisible"
      :title="detailTitle"
      size="min(900px,96vw)"
      destroy-on-close
      ><div v-loading="detailLoading">
        <section v-if="overviewEntries.length" class="detail-grid">
          <div v-for="[key, value] in overviewEntries" :key="key">
            <span>{{
              t(`tenantExperience.detailKeys.${key}`, humanizeBusinessKey(key))
            }}</span
            ><strong>{{ valueText(value) }}</strong>
          </div>
        </section>
        <el-tabs v-if="detailTabs.length"
          ><el-tab-pane
            v-for="[key, value] in detailTabs"
            :key="key"
            :label="
              t(`tenantExperience.detailTabs.${key}`, humanizeBusinessKey(key))
            "
            ><el-table
              v-if="Array.isArray(value)"
              :data="value"
              max-height="430"
              ><el-table-column
                v-for="column in businessColumns(value)"
                :key="column"
                :prop="column"
                :label="
                  t(
                    `tenantExperience.detailKeys.${column}`,
                    humanizeBusinessKey(column),
                  )
                "
                min-width="145"
                show-overflow-tooltip
            /></el-table>
            <section v-else class="detail-grid">
              <div v-for="[field, item] in businessEntries(value)" :key="field">
                <span>{{
                  t(
                    `tenantExperience.detailKeys.${field}`,
                    humanizeBusinessKey(field),
                  )
                }}</span
                ><strong>{{ valueText(item) }}</strong>
              </div>
            </section></el-tab-pane
          ></el-tabs
        >
      </div></el-drawer
    >

    <el-dialog
      v-model="createVisible"
      :title="t('tenantExperience.support.createCase')"
      width="min(640px,94vw)"
      destroy-on-close
      ><el-form label-position="top"
        ><el-form-item :label="t('tenantExperience.support.playerNumber')"
          ><el-input v-model="caseForm.playerNo" /></el-form-item
        ><el-row :gutter="14"
          ><el-col :span="12"
            ><el-form-item :label="t('tenantExperience.fields.category')"
              ><el-select v-model="caseForm.category"
                ><el-option
                  v-for="item in [
                    'ACCOUNT',
                    'DEPOSIT',
                    'WITHDRAWAL',
                    'WALLET',
                    'GAME',
                    'PROMOTION',
                    'TECHNICAL',
                    'OTHER',
                  ]"
                  :key="item"
                  :label="t(`tenantExperience.caseCategories.${item}`)"
                  :value="item" /></el-select></el-form-item></el-col
          ><el-col :span="12"
            ><el-form-item :label="t('tenantExperience.fields.priority')"
              ><el-select v-model="caseForm.priority"
                ><el-option
                  v-for="item in ['LOW', 'NORMAL', 'HIGH', 'URGENT']"
                  :key="item"
                  :label="t(`tenantExperience.priority.${item}`)"
                  :value="item" /></el-select></el-form-item></el-col></el-row
        ><el-form-item :label="t('tenantExperience.fields.title')" required
          ><el-input v-model="caseForm.title" /></el-form-item
        ><el-form-item
          :label="t('tenantExperience.fields.description')"
          required
          ><el-input
            v-model="caseForm.description"
            type="textarea"
            :rows="4" /></el-form-item
        ><el-form-item :label="t('tenantExperience.fields.reason')" required
          ><el-input
            v-model="caseForm.reason"
            type="textarea" /></el-form-item></el-form
      ><template #footer
        ><el-button @click="createVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button type="primary" @click="createCase">{{
          t("tenantExperience.support.createCase")
        }}</el-button></template
      ></el-dialog
    >
  </div>
</template>

<style scoped>
.support-page {
  display: grid;
  gap: 16px;
}
.page-header,
.filters,
.table-shell {
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
}
.page-header h1 {
  margin: 5px 0;
  font-size: 26px;
  letter-spacing: 0;
}
.page-header span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.filters {
  padding: 14px 14px 2px;
}
.table-shell {
  overflow: hidden;
}
.table-shell :deep(.el-pagination) {
  justify-content: flex-end;
  padding: 14px;
}
.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1px;
  margin-bottom: 16px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-border-color-lighter);
}
.detail-grid > div {
  display: grid;
  gap: 5px;
  min-height: 64px;
  padding: 10px;
  background: var(--el-bg-color);
}
.detail-grid span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.detail-grid strong {
  font-size: 14px;
  overflow-wrap: anywhere;
}
.el-select {
  width: 100%;
}
@media (max-width: 760px) {
  .page-header {
    flex-direction: column;
  }
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
