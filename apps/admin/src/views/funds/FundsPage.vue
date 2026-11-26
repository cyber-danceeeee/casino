<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import ReasonConfirmDialog from "@/components/users/ReasonConfirmDialog.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { useAdminCenterText } from "@/composables/useAdminCenterText";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { businessColumns, businessEntries, isInternalIdentifierKey, sanitizeBusinessValue } from "@/utils/presentation";
import { fundPageConfigs, type FundRecord } from "./config";
import {
  fetchAdminOptions,
  optionList,
  type AdminOption,
  type AdminOptionSource,
  type AdminOptions,
} from "@/api/adminOptions";
import {
  cancelDeposit,
  confirmDeposit,
  createAdjustment,
  decideAdjustment,
  decideWithdrawal,
  fetchAdjustment,
  fetchDeposit,
  fetchReconciliation,
  fetchTransaction,
  fetchWallet,
  fetchWallets,
  fetchWithdrawal,
  generateReconciliation,
  resolveDifference,
  reviewDeposit,
  updateWalletStatus,
  type FundDetail,
  type WalletRow,
} from "@/api/adminFunds";

type DetailGroup = {
  key: string;
  title: string;
  object?: FundRecord;
  rows?: FundRecord[];
};
type ReasonAction = {
  title: string;
  message: string;
  run: (reason: string) => Promise<unknown>;
};

const route = useRoute();
const { can, loadPermissions } = useAdminPermissions();
const tr = useAdminCenterText("fundsCenter");
const moduleKey = computed(() =>
  String(route.meta.moduleKey || "wallet-accounts"),
);
const config = computed(() => fundPageConfigs[moduleKey.value]);
const visibleFilters = computed(() => config.value?.filters.filter((item) => !isInternalIdentifierKey(item.key)) || []);
const visibleColumns = computed(() => config.value?.columns.filter((item) => !isInternalIdentifierKey(item.key)) || []);
const query = reactive<Record<string, string | string[]>>({});
const rows = ref<FundRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const adminOptions = ref<AdminOptions>({});
const loading = ref(false);

const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<FundDetail>({});
const activeDetailTab = ref("");
const selectedRow = ref<FundRecord | null>(null);

const reasonVisible = ref(false);
const reasonLoading = ref(false);
const reasonAction = ref<ReasonAction | null>(null);
const adjustmentVisible = ref(false);
const adjustmentLoading = ref(false);
const walletOptions = ref<WalletRow[]>([]);
const walletOptionLoading = ref(false);
const adjustmentReasonOptions = [
  "BONUS_COMPENSATION",
  "CUSTOMER_SERVICE_COMPENSATION",
  "PROMOTION_ADJUSTMENT",
  "CORRECTION",
  "CHARGEBACK",
  "RISK_DEDUCTION",
  "OTHER",
];
const adjustmentForm = reactive({
  userID: "",
  walletID: "",
  adjustmentType: "CREDIT",
  amount: "",
  currency: "PKR",
  reasonCode: "",
  detailedReason: "",
  supportingReference: "",
  internalRemark: "",
});
const depositConfirmVisible = ref(false);
const depositConfirmLoading = ref(false);
const depositConfirmForm = reactive({
  reason: "",
  evidenceReference: "",
  channelOrderID: "",
  receivedAmount: "",
});
const reconciliationVisible = ref(false);
const reconciliationLoading = ref(false);
const reconciliationForm = reactive({
  businessDate: "",
  businessType: "DEPOSIT",
  channel: "",
  currency: "PKR",
});

const detailGroups = computed<DetailGroup[]>(() =>
  (config.value?.detailTabs || []).flatMap(({ key, title }) => {
    if (!(key in detail.value)) return [];
    const value = detail.value[key];
    if (Array.isArray(value))
      return { key, title, rows: value.filter(isRecord) };
    if (isRecord(value)) return { key, title, object: value };
    return { key, title, object: { value } };
  }),
);

function isRecord(value: unknown): value is FundRecord {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function humanize(value: string) {
  return value
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .replace(/_/g, " ")
    .replace(/^./, (part) => part.toUpperCase());
}
function valueText(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (typeof value === "boolean") return tr(value ? "Yes" : "No");
  if (key.toLowerCase().endsWith("at") || key.toLowerCase().includes("time"))
    return formatBackendTime(String(value));
  if (typeof value === "object") return JSON.stringify(sanitizeBusinessValue(value));
  return String(value);
}
function objectEntries(value?: FundRecord) {
  return businessEntries(value);
}
function arrayColumns(list?: FundRecord[]) {
  return businessColumns(list, 9);
}
function rowValue(row: FundRecord, key: string) {
  return row[key];
}
function userPhoneText(row: FundRecord) {
  return (
    [row.phoneCountryCode, row.phone]
      .map((value) => String(value || "").trim())
      .filter(Boolean)
      .join(" ") ||
    String(row.username || row.phone || "-")
  );
}
const enumKeys = new Set([
  "status",
  "accountStatus",
  "kycStatus",
  "riskStatus",
  "currency",
  "transactionType",
  "direction",
  "referenceType",
  "providerCode",
  "adjustmentType",
  "reasonCode",
  "paymentChannel",
  "paymentMethod",
  "businessType",
  "mode",
  "auditStatus",
  "differenceType",
  "resolutionStatus",
  "decision",
]);
function isEnumKey(key: string) {
  return (
    enumKeys.has(key) ||
    /(^|_)(status|type|direction|currency)$/.test(key) ||
    /(Result|Stage)$/.test(key)
  );
}
function isTaggedValue(key: string, value: unknown) {
  return isEnumKey(key) || typeof value === "boolean";
}
function taggedLabel(value: unknown) {
  if (typeof value === "boolean") return tr(value ? "Yes" : "No");
  return tr(String(value || "UNKNOWN"));
}
const dynamicFilterSources: Record<string, AdminOptionSource> = {
  currency: "currencies",
  country: "countries",
  vipLevel: "vipLevels",
  providerCode: "providers",
  paymentChannel: "paymentChannels",
  paymentMethod: "paymentMethods",
  channel: "paymentChannels",
  registrationChannel: "registrationChannels",
};
function filterSource(field: { key: string }) {
  return dynamicFilterSources[field.key];
}
function fieldOptions(field: { key: string; options?: AdminOption[] }) {
  const dynamic = optionList(adminOptions.value, filterSource(field));
  return dynamic.length ? dynamic : field.options || [];
}
function filterOptionLabel(option: AdminOption) {
  return option.label && option.label !== option.value ? option.label : tr(option.label);
}
function isSelectFilter(field: { key: string; type?: string }) {
  return field.type === "select" || Boolean(filterSource(field));
}
function indexMethod(index: number) {
  return (page.value - 1) * pageSize.value + index + 1;
}

function initializeQuery() {
  for (const key of Object.keys(query)) delete query[key];
  for (const field of config.value?.filters || [])
    query[field.key] = field.type === "dateRange" ? [] : "";
}

function buildParams() {
  const result: Record<string, string | number> = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const [key, value] of Object.entries(query)) {
    if (Array.isArray(value)) {
      if (!value[0]) continue;
      const base = key.replace(/Range$/, "");
      result[`${base}From`] = selectedTimeToBackend(value[0]);
      result[`${base}To`] = selectedTimeToBackend(value[1] || value[0], true);
    } else if (value.trim()) result[key] = value.trim();
  }
  return result;
}

async function load() {
  if (!config.value || loading.value) return;
  loading.value = true;
  try {
    const data = await config.value.loader(buildParams());
    rows.value = data.list;
    total.value = data.total;
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
  } finally {
    loading.value = false;
  }
}
function search() {
  page.value = 1;
  void load();
}
function reset() {
  initializeQuery();
  search();
}

async function view(row: FundRecord) {
  selectedRow.value = row;
  detail.value = {};
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    const id = Number(row.id);
    const loaders: Record<string, (value: number) => Promise<FundDetail>> = {
      "wallet-accounts": fetchWallet,
      "wallet-ledgers": fetchTransaction,
      "manual-adjustments": fetchAdjustment,
      "deposit-orders": fetchDeposit,
      "withdrawal-orders": fetchWithdrawal,
      "financial-reconciliation": fetchReconciliation,
    };
    detail.value = await loaders[moduleKey.value](id);
    await nextTick();
    activeDetailTab.value = detailGroups.value[0]?.key || "";
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
    detailVisible.value = false;
  } finally {
    detailLoading.value = false;
  }
}

function ask(
  title: string,
  message: string,
  run: (reason: string) => Promise<unknown>,
) {
  reasonAction.value = { title, message, run };
  reasonVisible.value = true;
}
async function confirmReason(reason: string) {
  if (!reasonAction.value || reasonLoading.value) return;
  reasonLoading.value = true;
  try {
    await reasonAction.value.run(reason);
    reasonVisible.value = false;
    ElMessage.success(tr("Operation completed"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
  } finally {
    reasonLoading.value = false;
  }
}

function walletAction(row: FundRecord) {
  const frozen = row.status !== "FROZEN";
  ask(
    tr(frozen ? "Freeze Wallet" : "Unfreeze Wallet"),
    `${tr("Player Number")} ${String(row.userNo || "-")} · ${row.currency || ""} ${row.availableBalance || "-"} · ${tr("This action is audited.")}`,
    (reason) => updateWalletStatus(Number(row.id), frozen, reason),
  );
}
function walletPhone(row: Pick<WalletRow, "phoneCountryCode" | "phone">) {
  return [row.phoneCountryCode, row.phone].filter(Boolean).join(" ");
}
function walletOptionLabel(row: WalletRow) {
  const phone = walletPhone(row);
  return `${row.userNo} · ${phone || "-"} · ${row.currency} ${row.availableBalance}`;
}
const selectedAdjustmentWallet = computed(() =>
  walletOptions.value.find((item) => Number(item.id) === Number(adjustmentForm.walletID)),
);
async function loadWalletOptions(keyword = "") {
  walletOptionLoading.value = true;
  try {
    const params: Record<string, string | number> = { page: 1, pageSize: 20 };
    const value = keyword.trim();
    if (value) params.userNo = value;
    const data = await fetchWallets(params);
    walletOptions.value = data.list;
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
  } finally {
    walletOptionLoading.value = false;
  }
}
function selectAdjustmentWallet(walletID: string | number) {
  const wallet = walletOptions.value.find(
    (item) => Number(item.id) === Number(walletID),
  );
  if (!wallet) return;
  adjustmentForm.walletID = String(wallet.id);
  adjustmentForm.userID = String(wallet.userID);
  adjustmentForm.currency = wallet.currency;
}
function openAdjustmentDialog() {
  Object.assign(adjustmentForm, {
    userID: "",
    walletID: "",
    adjustmentType: "CREDIT",
    amount: "",
    currency: "PKR",
    reasonCode: "",
    detailedReason: "",
    supportingReference: "",
    internalRemark: "",
  });
  adjustmentVisible.value = true;
  void loadWalletOptions();
}
function adjustmentAction(
  row: FundRecord,
  action: "approve" | "reject" | "cancel",
) {
  ask(
    `${tr(humanize(action))} · ${tr("Manual Adjustments")}`,
    `${tr("Adjustment Number")} ${String(row.adjustmentID || "-")} · ${tr("Player Number")} ${String(row.userNo || "-")} · ${tr("This action is audited.")}`,
    (reason) => decideAdjustment(Number(row.id), action, reason),
  );
}
function depositAction(row: FundRecord, action: "review" | "cancel") {
  ask(
    `${tr(humanize(action))} · ${tr("Deposit Orders")}`,
    `${tr("Deposit Order Number")} ${String(row.platformOrderID || "-")} · ${tr("Player Number")} ${String(row.userNo || "-")} · ${tr("This action is audited.")}`,
    (reason) =>
      action === "review"
        ? reviewDeposit(Number(row.id), reason)
        : cancelDeposit(Number(row.id), reason),
  );
}
function withdrawalAction(
  row: FundRecord,
  action: "risk-review" | "finance-review" | "approve" | "reject" | "retry",
) {
  ask(
    `${tr(humanize(action))} · ${tr("Withdrawal Orders")}`,
    `${tr("Withdrawal Order Number")} ${String(row.platformOrderID || "-")} · ${tr("Player Number")} ${String(row.userNo || "-")} · ${tr("This action is audited.")}`,
    (reason) => decideWithdrawal(Number(row.id), action, reason),
  );
}

function openDepositConfirm(row: FundRecord) {
  selectedRow.value = row;
  Object.assign(depositConfirmForm, {
    reason: "",
    evidenceReference: "",
    channelOrderID: String(row.channelOrderID || ""),
    receivedAmount: String(row.requestedAmount || ""),
  });
  depositConfirmVisible.value = true;
}
async function submitDepositConfirm() {
  if (
    !selectedRow.value ||
    !depositConfirmForm.reason.trim() ||
    depositConfirmLoading.value
  )
    return;
  depositConfirmLoading.value = true;
  try {
    await confirmDeposit(Number(selectedRow.value.id), {
      ...depositConfirmForm,
    });
    depositConfirmVisible.value = false;
    ElMessage.success(tr("Deposit credited successfully"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
  } finally {
    depositConfirmLoading.value = false;
  }
}

async function submitAdjustment() {
  if (adjustmentLoading.value) return;
  adjustmentLoading.value = true;
  try {
    await createAdjustment({
      ...adjustmentForm,
      userID: Number(adjustmentForm.userID),
      walletID: Number(adjustmentForm.walletID),
    });
    adjustmentVisible.value = false;
    ElMessage.success(tr("Adjustment submitted for approval"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
  } finally {
    adjustmentLoading.value = false;
  }
}

async function submitReconciliation() {
  if (reconciliationLoading.value) return;
  reconciliationLoading.value = true;
  try {
    await generateReconciliation({ ...reconciliationForm });
    reconciliationVisible.value = false;
    ElMessage.success(tr("Reconciliation generated"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
  } finally {
    reconciliationLoading.value = false;
  }
}

function resolveRow(row: FundRecord) {
  if (!selectedRow.value) return;
  ask(
    tr("Resolve Difference"),
    `${tr("Difference")} ${String(row.id)} · ${tr("This action is audited.")}`,
    async (reason) => {
      await resolveDifference(
        Number(selectedRow.value?.id),
        Number(row.id),
        reason,
      );
      detail.value = await fetchReconciliation(Number(selectedRow.value?.id));
    },
  );
}

watch(moduleKey, () => {
  initializeQuery();
  page.value = 1;
  void load();
});
onMounted(async () => {
  initializeQuery();
  await Promise.all([
    loadPermissions(),
    fetchAdminOptions()
      .then((options) => {
        adminOptions.value = options;
      })
      .catch(() => {
        adminOptions.value = {};
      }),
  ]);
  await load();
});
</script>

<template>
  <div v-if="config" class="page-stack page-stack--search-table funds-page">
    <UserCenterCard :title="tr(config.searchTitle)">
      <el-form
        :model="query"
        inline
        class="fund-search-form"
        @submit.prevent="search"
      >
        <el-form-item
          v-for="field in visibleFilters"
          :key="field.key"
          :label="tr(field.label)"
        >
          <el-select
            v-if="isSelectFilter(field)"
            v-model="query[field.key]"
            clearable
            filterable
          >
            <el-option
              v-for="option in fieldOptions(field)"
              :key="option.value"
              :label="filterOptionLabel(option)"
              :value="option.value"
            />
          </el-select>
          <el-date-picker
            v-else-if="field.type === 'dateRange'"
            v-model="query[field.key]"
            type="daterange"
            value-format="YYYY-MM-DD"
            range-separator="-"
          />
          <el-input
            v-else
            v-model="query[field.key]"
            clearable
            :placeholder="tr(field.label)"
            @keyup.enter="search"
          />
        </el-form-item>
        <el-form-item
          ><el-button type="primary" :loading="loading" @click="search">{{
            tr("Search")
          }}</el-button
          ><el-button @click="reset">{{ tr("Reset") }}</el-button></el-form-item
        >
      </el-form>
    </UserCenterCard>

    <UserCenterCard :title="tr(config.tableTitle)">
      <template #header>
        <el-button
          v-if="
            moduleKey === 'manual-adjustments' && can('fund:adjustment:create')
          "
          type="primary"
          @click="openAdjustmentDialog"
          >{{ tr("Create Adjustment") }}</el-button
        >
        <el-button
          v-if="
            moduleKey === 'financial-reconciliation' &&
            can('fund:reconciliation:generate')
          "
          type="primary"
          @click="reconciliationVisible = true"
          >{{ tr("Generate Reconciliation") }}</el-button
        >
      </template>
      <el-table
        v-loading="loading"
        :data="rows"
        row-key="id"
        border
        :empty-text="tr('No data')"
      >
        <el-table-column
          type="index"
          :index="indexMethod"
          :label="tr('Index')"
          fixed="left"
          width="72"
          align="center"
        />
        <el-table-column
          v-for="column in visibleColumns"
          :key="column.key"
          :prop="column.key"
          :label="tr(column.label)"
          :width="column.width"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span v-if="column.key === 'username' || column.key === 'phone'">
              {{ userPhoneText(row) }}
            </span>
            <UserStatusTag
              v-else-if="column.status"
              :value="String(rowValue(row, column.key) || 'UNKNOWN')"
              :label="tr(String(rowValue(row, column.key) || 'UNKNOWN'))"
            />
            <UserStatusTag
              v-else-if="isEnumKey(column.key)"
              :value="String(rowValue(row, column.key) || 'UNKNOWN')"
              :label="tr(String(rowValue(row, column.key) || 'UNKNOWN'))"
            />
            <span v-else-if="column.time">{{
              formatBackendTime(String(rowValue(row, column.key) || ""))
            }}</span>
            <el-tag
              v-else-if="column.dangerTag"
              type="danger"
              effect="light"
              round
            >
              {{ rowValue(row, column.key) || "0.00" }}
            </el-tag>
            <span
              v-else-if="column.money"
              :class="[
                'money-value',
                String(rowValue(row, 'direction')) === 'OUT'
                  ? 'money-value--out'
                  : '',
              ]"
              >{{ rowValue(row, column.key) || "0.00" }}</span
            >
            <span v-else>{{ rowValue(row, column.key) ?? "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="tr('Actions')"
          fixed="right"
          width="285"
          align="center"
        >
          <template #default="{ row }">
            <div class="table-actions">
              <el-button link type="primary" @click="view(row)">{{
                tr("View")
              }}</el-button>
              <el-button
                v-if="
                  moduleKey === 'wallet-accounts' &&
                  can(
                    row.status === 'FROZEN'
                      ? 'fund:wallet:unfreeze'
                      : 'fund:wallet:freeze',
                  )
                "
                link
                :type="row.status === 'FROZEN' ? 'success' : 'warning'"
                @click="walletAction(row)"
                >{{
                  tr(row.status === "FROZEN" ? "Unfreeze" : "Freeze")
                }}</el-button
              >
              <template
                v-if="
                  moduleKey === 'manual-adjustments' &&
                  row.status === 'PENDING_APPROVAL'
                "
              >
                <el-button
                  v-if="can('fund:adjustment:approve')"
                  link
                  type="success"
                  @click="adjustmentAction(row, 'approve')"
                  >{{ tr("Approve") }}</el-button
                >
                <el-button
                  v-if="can('fund:adjustment:reject')"
                  link
                  type="danger"
                  @click="adjustmentAction(row, 'reject')"
                  >{{ tr("Reject") }}</el-button
                >
                <el-button
                  v-if="can('fund:adjustment:cancel')"
                  link
                  @click="adjustmentAction(row, 'cancel')"
                  >{{ tr("Cancel") }}</el-button
                >
              </template>
              <template
                v-if="
                  moduleKey === 'deposit-orders' &&
                  ['PENDING', 'MANUAL_REVIEW'].includes(String(row.status))
                "
              >
                <el-button
                  v-if="row.status === 'PENDING' && can('fund:deposit:review')"
                  link
                  type="warning"
                  @click="depositAction(row, 'review')"
                  >{{ tr("Review") }}</el-button
                >
                <el-button
                  v-if="can('fund:deposit:confirm')"
                  link
                  type="success"
                  @click="openDepositConfirm(row)"
                  >{{ tr("Confirm") }}</el-button
                >
                <el-button
                  v-if="can('fund:deposit:cancel')"
                  link
                  type="danger"
                  @click="depositAction(row, 'cancel')"
                  >{{ tr("Cancel") }}</el-button
                >
              </template>
              <template v-if="moduleKey === 'withdrawal-orders'">
                <el-button
                  v-if="
                    row.status === 'PENDING' &&
                    can('fund:withdrawal:risk-review')
                  "
                  link
                  type="warning"
                  @click="withdrawalAction(row, 'risk-review')"
                  >{{ tr("Risk Review") }}</el-button
                >
                <el-button
                  v-if="
                    row.status === 'RISK_REVIEW' &&
                    can('fund:withdrawal:finance-review')
                  "
                  link
                  type="warning"
                  @click="withdrawalAction(row, 'finance-review')"
                  >{{ tr("Finance Review") }}</el-button
                >
                <el-button
                  v-if="
                    row.status === 'FINANCE_REVIEW' &&
                    can('fund:withdrawal:approve')
                  "
                  link
                  type="success"
                  @click="withdrawalAction(row, 'approve')"
                  >{{ tr("Approve") }}</el-button
                >
                <el-button
                  v-if="
                    ['PENDING', 'RISK_REVIEW', 'FINANCE_REVIEW'].includes(
                      String(row.status),
                    ) && can('fund:withdrawal:reject')
                  "
                  link
                  type="danger"
                  @click="withdrawalAction(row, 'reject')"
                  >{{ tr("Reject") }}</el-button
                >
                <el-button
                  v-if="row.status === 'FAILED' && can('fund:withdrawal:retry')"
                  link
                  @click="withdrawalAction(row, 'retry')"
                  >{{ tr("Retry") }}</el-button
                >
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-row">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          background
          layout="prev, pager, next, sizes, total"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </UserCenterCard>

    <el-dialog
      v-model="detailVisible"
      :title="`${tr(config.tableTitle)} · ${tr('Details')}`"
      width="min(1080px, 92vw)"
      destroy-on-close
      append-to-body
      class="fund-detail-dialog"
    >
      <div v-loading="detailLoading" class="fund-detail-body">
        <el-tabs
          v-if="detailGroups.length"
          v-model="activeDetailTab"
          type="card"
        >
          <el-tab-pane
            v-for="group in detailGroups"
            :key="group.key"
            :name="group.key"
            :label="tr(group.title)"
          >
            <el-card v-if="group.object" shadow="never" class="detail-card">
              <el-descriptions :column="2" border>
                <el-descriptions-item
                  v-for="[key, value] in objectEntries(group.object)"
                  :key="key"
                  :label="tr(humanize(key))"
                  ><UserStatusTag
                    v-if="isTaggedValue(key, value)"
                    :value="String(value || 'UNKNOWN')"
                    :label="taggedLabel(value)"
                  /><template v-else>{{ valueText(key, value) }}</template></el-descriptions-item
                >
              </el-descriptions>
            </el-card>
            <el-card v-else shadow="never" class="detail-card">
              <el-table
                :data="group.rows"
                border
                max-height="460"
                :empty-text="tr('No data')"
              >
                <el-table-column
                  v-for="key in arrayColumns(group.rows)"
                  :key="key"
                  :prop="key"
                  :label="tr(humanize(key))"
                  min-width="135"
                  show-overflow-tooltip
                  ><template #default="{ row }"><UserStatusTag
                    v-if="isTaggedValue(key, row[key])"
                    :value="String(row[key] || 'UNKNOWN')"
                    :label="taggedLabel(row[key])"
                  /><template v-else>{{ valueText(key, row[key]) }}</template></template></el-table-column
                >
                <el-table-column
                  v-if="group.key === 'differences'"
                  :label="tr('Actions')"
                  fixed="right"
                  width="100"
                  ><template #default="{ row }"
                    ><el-button
                      v-if="
                        row.resolution_status === 'OPEN' &&
                        can('fund:reconciliation:resolve')
                      "
                      link
                      type="primary"
                      @click="resolveRow(row)"
                      >{{ tr("Resolve") }}</el-button
                    ></template
                  ></el-table-column
                >
              </el-table>
            </el-card>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer
        ><el-button @click="detailVisible = false">{{
          tr("Close")
        }}</el-button></template
      >
    </el-dialog>

    <ReasonConfirmDialog
      v-model="reasonVisible"
      :title="reasonAction?.title || tr('Confirm Operation')"
      :message="reasonAction?.message || ''"
      :loading="reasonLoading"
      @confirm="confirmReason"
    />

    <el-dialog
      v-model="adjustmentVisible"
      :title="tr('Create Manual Adjustment')"
      width="680px"
      destroy-on-close
      append-to-body
    >
      <el-form :model="adjustmentForm" label-width="150px">
        <el-form-item :label="tr('Player Number')" required>
          <el-select
            v-model="adjustmentForm.walletID"
            filterable
            remote
            clearable
            :remote-method="loadWalletOptions"
            :loading="walletOptionLoading"
            :placeholder="tr('Search by Player Number')"
            @change="selectAdjustmentWallet"
          >
            <el-option
              v-for="wallet in walletOptions"
              :key="wallet.id"
              :label="walletOptionLabel(wallet)"
              :value="String(wallet.id)"
            />
          </el-select>
        </el-form-item>
        <el-alert
          v-if="selectedAdjustmentWallet"
          type="info"
          :closable="false"
          show-icon
        >
          <template #title>
            {{ selectedAdjustmentWallet.userNo }} · {{ walletPhone(selectedAdjustmentWallet) || "-" }}
          </template>
          {{ tr("Available Balance") }}: {{ selectedAdjustmentWallet.currency }} {{ selectedAdjustmentWallet.availableBalance }} ·
          {{ tr("Frozen Balance") }}: {{ selectedAdjustmentWallet.frozenBalance }}
        </el-alert>
        <el-form-item :label="tr('Adjustment Type')" required
          ><el-radio-group v-model="adjustmentForm.adjustmentType"
            ><el-radio-button value="CREDIT">{{ tr("Credit") }}</el-radio-button
            ><el-radio-button value="DEBIT">{{
              tr("Debit")
            }}</el-radio-button></el-radio-group
          ></el-form-item
        >
        <el-form-item :label="tr('Amount')" required
          ><el-input v-model="adjustmentForm.amount"
            ><template #append>{{
              adjustmentForm.currency
            }}</template></el-input
          ></el-form-item
        >
        <el-form-item :label="tr('Currency')" required
          ><el-input v-model="adjustmentForm.currency" disabled /></el-form-item
        ><el-form-item :label="tr('Reason Code')" required
          ><el-select
            v-model="adjustmentForm.reasonCode"
            :placeholder="tr('Select reason')"
          >
            <el-option
              v-for="reasonItem in adjustmentReasonOptions"
              :key="reasonItem"
              :label="tr(reasonItem)"
              :value="reasonItem"
            />
          </el-select></el-form-item>
        <el-form-item :label="tr('Detailed Reason')" required
          ><el-input
            v-model="adjustmentForm.detailedReason"
            type="textarea"
            :rows="3"
            maxlength="1000" /></el-form-item
        ><el-form-item :label="tr('Supporting Ref.')"
          ><el-input
            v-model="adjustmentForm.supportingReference" /></el-form-item
        ><el-form-item :label="tr('Internal Remark')"
          ><el-input
            v-model="adjustmentForm.internalRemark"
            type="textarea"
            :rows="2"
        /></el-form-item> </el-form
      ><template #footer
        ><el-button @click="adjustmentVisible = false">{{
          tr("Cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="adjustmentLoading"
          :disabled="
            !adjustmentForm.userID ||
            !adjustmentForm.walletID ||
            !adjustmentForm.amount ||
            !adjustmentForm.reasonCode ||
            !adjustmentForm.detailedReason
          "
          @click="submitAdjustment"
          >{{ tr("Submit for Approval") }}</el-button
        ></template
      >
    </el-dialog>

    <el-dialog
      v-model="depositConfirmVisible"
      :title="tr('Confirm Deposit Credit')"
      width="600px"
      destroy-on-close
      append-to-body
    >
      <el-alert
        :title="
          tr(
            `This action credits the player's wallet and creates an immutable ledger entry.`,
          )
        "
        type="warning"
        :closable="false"
        show-icon
      />
      <el-form
        :model="depositConfirmForm"
        label-width="170px"
        class="operation-form"
        ><el-form-item :label="tr('Received Amount')" required
          ><el-input
            v-model="depositConfirmForm.receivedAmount" /></el-form-item
        ><el-form-item :label="tr('Upstream Order Number')"
          ><el-input
            v-model="depositConfirmForm.channelOrderID" /></el-form-item
        ><el-form-item :label="tr('Evidence Reference')"
          ><el-input
            v-model="depositConfirmForm.evidenceReference" /></el-form-item
        ><el-form-item :label="tr('Operator Reason')" required
          ><el-input
            v-model="depositConfirmForm.reason"
            type="textarea"
            :rows="3" /></el-form-item
      ></el-form>
      <template #footer
        ><el-button @click="depositConfirmVisible = false">{{
          tr("Cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="depositConfirmLoading"
          :disabled="
            !depositConfirmForm.receivedAmount ||
            !depositConfirmForm.reason.trim()
          "
          @click="submitDepositConfirm"
          >{{ tr("Confirm and Credit") }}</el-button
        ></template
      >
    </el-dialog>

    <el-dialog
      v-model="reconciliationVisible"
      :title="tr('Generate Reconciliation')"
      width="580px"
      destroy-on-close
      append-to-body
    >
      <el-form :model="reconciliationForm" label-width="140px"
        ><el-form-item :label="tr('Business Date')" required
          ><el-date-picker
            v-model="reconciliationForm.businessDate"
            type="date"
            value-format="YYYY-MM-DD" /></el-form-item
        ><el-form-item :label="tr('Business Type')" required
          ><el-select v-model="reconciliationForm.businessType"
            ><el-option :label="tr('Deposit')" value="DEPOSIT" /><el-option
              :label="tr('Withdrawal')"
              value="WITHDRAWAL" /><el-option
              :label="tr('Adjustment')"
              value="ADJUSTMENT" /></el-select></el-form-item
        ><el-form-item :label="tr('Channel')"
          ><el-input v-model="reconciliationForm.channel" /></el-form-item
        ><el-form-item :label="tr('Currency')" required
          ><el-input v-model="reconciliationForm.currency" /></el-form-item
      ></el-form>
      <template #footer
        ><el-button @click="reconciliationVisible = false">{{
          tr("Cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="reconciliationLoading"
          :disabled="
            !reconciliationForm.businessDate ||
            !reconciliationForm.businessType ||
            !reconciliationForm.currency
          "
          @click="submitReconciliation"
          >{{ tr("Generate") }}</el-button
        ></template
      >
    </el-dialog>
  </div>
</template>

<style scoped>
.fund-search-form :deep(.el-form-item) {
  margin: 0 18px 14px 0;
}
.fund-search-form :deep(.el-input),
.fund-search-form :deep(.el-select) {
  width: 190px;
}
.fund-search-form :deep(.el-date-editor) {
  width: 250px;
}
.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.table-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  white-space: nowrap;
}
.money-value {
  font-variant-numeric: tabular-nums;
  color: #67c23a;
}
.money-value--out {
  color: #f56c6c;
}
:deep(.el-tag) {
  font-variant-numeric: tabular-nums;
}
.fund-detail-body {
  min-height: 180px;
  max-height: 68vh;
  overflow: auto;
}
.detail-card {
  border-color: var(--admin-border);
}
.operation-form {
  margin-top: 20px;
}
.operation-form :deep(.el-input),
.operation-form :deep(.el-select),
.operation-form :deep(.el-date-editor),
:deep(.el-dialog .el-form .el-select) {
  width: 100%;
}
</style>
