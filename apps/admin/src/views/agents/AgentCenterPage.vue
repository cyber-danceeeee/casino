<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { useI18n } from "vue-i18n";
import { agentConfigs } from "./config";
import {
  fetchAdminOptions,
  optionList,
  stringOptions,
  type AdminOption,
  type AdminOptionSource,
  type AdminOptions,
} from "@/api/adminOptions";
import {
  adjustAgentWallet,
  assignAgentPlan,
  changeAgentAttribution,
  changeAgentParent,
  createAgent,
  createChannel,
  createPlan,
  deletePlan,
  fetchSettlementStatement,
  fetchAgents,
  fetchChannels,
  fetchPlans,
  generateSettlement,
  setAgentStatus,
  setChannelStatus,
  setPlanStatus,
  setWalletFrozen,
  settlementAction,
  updateAgent,
  updateChannel,
  updatePlan,
  type AgentDetail,
  type AgentRecord,
} from "@/api/adminAgents";
import { operationsAPI } from "@/api/adminStability";
import { businessColumns, businessEntries, sanitizeBusinessValue } from "@/utils/presentation";

const route = useRoute(),
  { t } = useI18n(),
  { can, loadPermissions } = useAdminPermissions();
const moduleKey = computed(() => String(route.meta.moduleKey || "agent-list")),
  config = computed(() => agentConfigs[moduleKey.value]);
const query = reactive<Record<string, string | string[]>>({}),
  rows = ref<AgentRecord[]>([]),
  total = ref(0),
  page = ref(1),
  pageSize = ref(20),
  adminOptions = ref<AdminOptions>({}),
  agentChoices = ref<AgentRecord[]>([]),
  channelChoices = ref<AgentRecord[]>([]),
  planChoices = ref<AgentRecord[]>([]),
  playerChoices = ref<Array<{ value: number; label: string }>>([]),
  playerSearching = ref(false),
  loading = ref(false);
const detailVisible = ref(false),
  detailLoading = ref(false),
  detail = ref<AgentDetail>({}),
  activeTab = ref(""),
  selected = ref<AgentRecord | null>(null);
const editVisible = ref(false),
  editLoading = ref(false),
  editID = ref(0),
  form = reactive<Record<string, any>>({});
const operationVisible = ref(false),
  operationLoading = ref(false),
  operation = reactive({
    type: "",
    reason: "",
    value: "",
    amount: "",
    adjustmentType: "MANUAL_CREDIT",
    agentId: "",
    channelId: "",
    source: "MANUAL",
    periodRange: [] as string[],
    currency: "PKR",
  });

const canCreate = computed(() =>
  Boolean(config.value?.createPermission && can(config.value.createPermission)),
);
const tableRows = computed(() => {
  if (moduleKey.value !== "agent-hierarchy") return rows.value;
  const nodes = new Map<number, AgentRecord & { children?: AgentRecord[] }>();
  rows.value.forEach((row) => nodes.set(row.id, { ...row, children: [] }));
  const roots: (AgentRecord & { children?: AgentRecord[] })[] = [];
  nodes.forEach((node) => {
    const parent = nodes.get(Number(node.parentAgentId || 0));
    if (parent) parent.children?.push(node);
    else roots.push(node);
  });
  nodes.forEach((node) => {
    if (!node.children?.length) delete node.children;
  });
  return roots;
});
function label(key: string) {
  return t(`agentCenter.fields.${key}`);
}
function translated(v: unknown) {
  const key = String(v ?? "").toLowerCase();
  const out = t(`agentCenter.values.${key}`);
  return out === `agentCenter.values.${key}` ? String(v ?? "-") : out;
}
const dynamicFilterSources: Record<string, AdminOptionSource> = {
  agentLevel: "agentLevels",
  country: "countries",
  currency: "currencies",
  commissionPlan: "commissionPlans",
  registrationChannel: "registrationChannels",
};
function filterSource(field: { key: string }) {
  return dynamicFilterSources[field.key];
}
function fieldOptions(field: { key: string; options?: string[] }) {
  const dynamic = optionList(adminOptions.value, filterSource(field));
  return dynamic.length ? dynamic : stringOptions(field.options);
}
function optionLabel(option: AdminOption) {
  return option.label && option.label !== option.value ? option.label : translated(option.value);
}
function isSelectFilter(field: { key: string; type?: string }) {
  return field.type === "select" || Boolean(filterSource(field));
}
function display(key: string, v: unknown) {
  if (v === null || v === undefined || v === "") return "-";
  if (key.toLowerCase().endsWith("at") || key.toLowerCase().includes("time"))
    return formatBackendTime(String(v));
  if (typeof v === "boolean") return v ? t("common.yes") : t("common.no");
  if (typeof v === "object") return JSON.stringify(sanitizeBusinessValue(v));
  return String(v);
}
function isObject(v: unknown): v is Record<string, unknown> {
  return Boolean(v && typeof v === "object" && !Array.isArray(v));
}
function entries(v: unknown) {
  return businessEntries(v);
}
function tabValue(key: string) {
  if (key === "overview" || key === "summary") return detail.value.record;
  return detail.value[key];
}
function arrayColumns(v: unknown) {
  return businessColumns(v, 10);
}
function agentChoiceLabel(row: AgentRecord) {
  return `${String(row.name || "-")} · ${String(row.agentCode || "-")}`;
}
function planChoiceLabel(row: AgentRecord) {
  return `${String(row.nameZh || row.nameEn || row.name || "-")} · ${String(row.code || "-")}`;
}
function channelChoiceLabel(row: AgentRecord) {
  return `${String(row.nameZh || row.nameEn || row.name || "-")} · ${String(row.code || "-")}`;
}
function rememberPlayer(record: Record<string, unknown>) {
  const value = Number(record.userId || 0);
  const number = String(record.userNo || "");
  if (!value || !number || playerChoices.value.some(item => item.value === value)) return;
  playerChoices.value.push({ value, label: `${number}${record.username ? ` · ${String(record.username)}` : ""}` });
}
async function searchPlayers(keyword: string) {
  const query = keyword.trim();
  if (!query) return;
  playerSearching.value = true;
  try {
    const result = await operationsAPI.support.players({ searchType: "AUTO", query, page: 1, pageSize: 20 });
    playerChoices.value = (result.list || []).map(row => ({
      value: Number(row.id || 0),
      label: `${String(row.playerNo || row.userNo || "-")} · ${String(row.mobile || row.username || "")}`,
    })).filter(item => item.value);
  } finally {
    playerSearching.value = false;
  }
}
async function loadBusinessChoices() {
  const [agents, channels, plans] = await Promise.all([
    fetchAgents({ page: 1, pageSize: 500 }).catch(() => ({ list: [], total: 0, page: 1, pageSize: 500 })),
    fetchChannels({ page: 1, pageSize: 500 }).catch(() => ({ list: [], total: 0, page: 1, pageSize: 500 })),
    fetchPlans({ page: 1, pageSize: 500 }).catch(() => ({ list: [], total: 0, page: 1, pageSize: 500 })),
  ]);
  agentChoices.value = agents.list || [];
  channelChoices.value = channels.list || [];
  planChoices.value = plans.list || [];
}
function normalizedQuery() {
  const out: Record<string, string | number> = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const [key, value] of Object.entries(query)) {
    if (Array.isArray(value) && value.length === 2) {
      out[key.replace("Range", "From")] = selectedTimeToBackend(
        value[0],
        false,
      );
      out[key.replace("Range", "To")] = selectedTimeToBackend(value[1], true);
    } else if (String(value ?? "").trim()) out[key] = String(value).trim();
  }
  return out;
}
async function load() {
  if (!config.value || !can(config.value.permission)) return;
  loading.value = true;
  try {
    const r = await config.value.loader(normalizedQuery());
    rows.value = r.list;
    total.value = r.total;
  } catch {
    ElMessage.error(t("agentCenter.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}
function search() {
  page.value = 1;
  load();
}
function reset() {
  Object.keys(query).forEach((k) => delete query[k]);
  page.value = 1;
  load();
}
async function view(row: AgentRecord) {
  selected.value = row;
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detail.value = await config.value.detail(row.id);
    activeTab.value = config.value.tabs[0] || "overview";
  } catch {
    ElMessage.error(t("agentCenter.messages.detailFailed"));
  } finally {
    detailLoading.value = false;
  }
}
function resetForm() {
  Object.keys(form).forEach((k) => delete form[k]);
  Object.assign(
    form,
    moduleKey.value === "agent-list"
      ? {
          agentCode: "",
          name: "",
          userId: 0,
          parentAgentId: 0,
          country: "PK",
          currency: "PKR",
          commissionPlanId: 0,
          reason: "",
        }
      : moduleKey.value === "channel-management"
        ? {
            code: "",
            nameZh: "",
            nameEn: "",
            agentId: 0,
            country: "PK",
            platform: "H5",
            registrationUrl: "",
            trackingParameters: {},
            enabled: true,
            reason: "",
          }
        : {
            code: "",
            nameZh: "",
            nameEn: "",
            model: "CPA",
            currency: "PKR",
            cpaAmount: "0.00",
            revenueShareRate: "0.0000",
            negativeCarryover: false,
            minimumSettlementAmount: "0.00",
            settlementCycle: "MONTHLY",
            deductionRules: {},
            tierRules: [],
            eligibleProducts: [],
            enabled: true,
            reason: "",
          },
  );
}
function openCreate() {
  editID.value = 0;
  resetForm();
  editVisible.value = true;
}
async function openCopy(row: AgentRecord) {
  editID.value = 0;
  resetForm();
  editVisible.value = true;
  editLoading.value = true;
  try {
    const d = await config.value.detail(row.id);
    const record = isObject(d.record) ? d.record : {};
    rememberPlayer(record);
    Object.assign(form, record, {
      code: `${String(record.code || "")}-COPY`,
      reason: "",
    });
    form.deductionRules = d.deductionRules || {};
    form.tierRules = d.tierRules || [];
    form.eligibleProducts = d.eligibleProducts || [];
  } finally {
    editLoading.value = false;
  }
}
async function openEdit(row: AgentRecord) {
  editID.value = row.id;
  resetForm();
  editVisible.value = true;
  editLoading.value = true;
  try {
    const d = await config.value.detail(row.id);
    const record = isObject(d.record) ? d.record : {};
    rememberPlayer(record);
    Object.assign(form, record);
    if (moduleKey.value === "commission-plans") {
      form.deductionRules = d.deductionRules || {};
      form.tierRules = d.tierRules || [];
      form.eligibleProducts = d.eligibleProducts || [];
    }
    form.reason = "";
  } finally {
    editLoading.value = false;
  }
}
async function save() {
  if (!String(form.reason || "").trim()) {
    ElMessage.warning(t("agentCenter.validation.reasonRequired"));
    return;
  }
  editLoading.value = true;
  try {
    if (moduleKey.value === "agent-list")
      editID.value
        ? await updateAgent(editID.value, form)
        : await createAgent(form);
    else if (moduleKey.value === "channel-management")
      editID.value
        ? await updateChannel(editID.value, form)
        : await createChannel(form);
    else
      editID.value
        ? await updatePlan(editID.value, form)
        : await createPlan(form);
    editVisible.value = false;
    ElMessage.success(t("agentCenter.messages.saved"));
    load();
  } catch {
    ElMessage.error(t("agentCenter.messages.saveFailed"));
  } finally {
    editLoading.value = false;
  }
}
function ask(type: string, row: AgentRecord) {
  selected.value = row;
  operation.type = type;
  operation.reason = "";
  operation.value = "";
  operation.amount = "";
  operation.adjustmentType = "MANUAL_CREDIT";
  operation.agentId = "";
  operation.channelId = "";
  operation.source = "MANUAL";
  operation.periodRange = [];
  operation.currency = "PKR";
  operationVisible.value = true;
}
async function runOperation() {
  if (!selected.value || !operation.reason.trim()) {
    ElMessage.warning(t("agentCenter.validation.reasonRequired"));
    return;
  }
  operationLoading.value = true;
  try {
    const id = selected.value.id;
    switch (operation.type) {
      case "agent-status":
        await setAgentStatus(
          id,
          String(selected.value.status) !== "ACTIVE",
          operation.reason,
        );
        break;
      case "change-parent":
        await changeAgentParent(
          id,
          Number(operation.value || 0),
          operation.reason,
        );
        break;
      case "assign-plan":
        await assignAgentPlan(id, Number(operation.value), operation.reason);
        break;
      case "change-attribution":
        await changeAgentAttribution({
          userId: Number(operation.value),
          agentId: id,
          channelId: Number(operation.channelId || 0),
          source: operation.source,
          reason: operation.reason,
        });
        break;
      case "channel-status":
        await setChannelStatus(
          id,
          !Boolean(selected.value.enabled),
          operation.reason,
        );
        break;
      case "plan-status":
        await setPlanStatus(
          id,
          !Boolean(selected.value.enabled),
          operation.reason,
        );
        break;
      case "delete-plan":
        await deletePlan(id, operation.reason);
        break;
      case "wallet-freeze":
        await setWalletFrozen(
          id,
          String(selected.value.status) !== "FROZEN",
          operation.reason,
        );
        break;
      case "wallet-adjust":
        await adjustAgentWallet(id, {
          type: operation.adjustmentType,
          amount: operation.amount,
          idempotencyKey: `admin-${id}-${Date.now()}`,
          reason: operation.reason,
        });
        break;
      default:
        await settlementAction(id, operation.type, operation.reason);
    }
    operationVisible.value = false;
    ElMessage.success(t("agentCenter.messages.operationSucceeded"));
    load();
  } catch {
    ElMessage.error(t("agentCenter.messages.operationFailed"));
  } finally {
    operationLoading.value = false;
  }
}
async function openGenerate() {
  selected.value = null;
  operation.type = "generate";
  operation.reason = "";
  operation.agentId = "";
  operation.periodRange = [];
  operation.currency = "PKR";
  operationVisible.value = true;
}
function backendISO(value: string, end = false) {
  const converted = selectedTimeToBackend(value, end);
  return converted ? `${converted.replace(" ", "T")}+08:00` : "";
}
async function generate() {
  if (
    !operation.agentId ||
    operation.periodRange.length !== 2 ||
    !operation.currency ||
    !operation.reason.trim()
  ) {
    ElMessage.warning(t("agentCenter.validation.generateFormat"));
    return;
  }
  operationLoading.value = true;
  try {
    const [start, end] = operation.periodRange;
    await generateSettlement({
      agentId: Number(operation.agentId),
      periodStart: backendISO(start),
      periodEnd: backendISO(end, true),
      currency: operation.currency,
      idempotencyKey: `settlement-${operation.agentId}-${start}-${end}-${operation.currency}`,
      reason: operation.reason,
    });
    operationVisible.value = false;
    ElMessage.success(t("agentCenter.messages.operationSucceeded"));
    load();
  } catch {
    ElMessage.error(t("agentCenter.messages.operationFailed"));
  } finally {
    operationLoading.value = false;
  }
}
async function copyRegistrationURL(row: AgentRecord) {
  try {
    await navigator.clipboard.writeText(String(row.registrationUrl || ""));
    ElMessage.success(t("agentCenter.messages.copied"));
  } catch {
    ElMessage.error(t("agentCenter.messages.copyFailed"));
  }
}
async function exportStatement(row: AgentRecord) {
  try {
    const statement = await fetchSettlementStatement(row.id);
    const blob = new Blob([JSON.stringify(statement, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${String(row.settlement_no || "settlement")}.json`;
    link.click();
    URL.revokeObjectURL(url);
  } catch {
    ElMessage.error(t("agentCenter.messages.exportFailed"));
  }
}
watch(moduleKey, async () => {
  Object.keys(query).forEach((k) => delete query[k]);
  page.value = 1;
  await nextTick();
  load();
});
onMounted(async () => {
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
  await loadBusinessChoices();
  load();
});
</script>

<template>
  <div class="agent-page" v-if="config">
    <UserCenterCard :title="t(`agentCenter.pages.${config.search}`)"
      ><el-form
        class="search-form"
        label-position="left"
        label-width="132px"
        @submit.prevent="search"
        ><el-row :gutter="16"
          ><el-col
            v-for="f in config.filters"
            :key="f.key"
            :xs="24"
            :sm="12"
            :lg="8"
            :xl="6"
            ><el-form-item :label="label(f.label)"
              ><el-select
                v-if="isSelectFilter(f)"
                v-model="query[f.key]"
                clearable
                filterable
                ><el-option
                  v-for="o in fieldOptions(f)"
                  :key="o.value"
                  :label="optionLabel(o)"
                  :value="o.value" /></el-select
              ><el-date-picker
                v-else-if="f.type === 'dateRange'"
                v-model="query[f.key]"
                type="datetimerange"
                value-format="YYYY-MM-DD HH:mm:ss"
                :start-placeholder="t('agentCenter.placeholders.start')"
                :end-placeholder="t('agentCenter.placeholders.end')" /><el-input
                v-else
                v-model="query[f.key]"
                clearable
                :placeholder="t('agentCenter.placeholders.enter')"
                @keyup.enter="search" /></el-form-item></el-col
        ></el-row>
        <div class="actions">
          <el-button @click="reset">{{ t("common.reset") }}</el-button
          ><el-button type="primary" native-type="submit">{{
            t("common.search")
          }}</el-button>
        </div></el-form
      ></UserCenterCard
    >
    <UserCenterCard :title="t(`agentCenter.pages.${config.title}`)"
      ><template #extra
        ><el-button
          v-if="
            moduleKey === 'agent-settlements' &&
            can('agent:settlement:generate')
          "
          type="primary"
          @click="openGenerate"
          >{{ t("agentCenter.actions.generate") }}</el-button
        ><el-button v-if="canCreate" type="primary" @click="openCreate">{{
          t("common.create")
        }}</el-button></template
      ><el-table
        v-loading="loading"
        :data="tableRows"
        row-key="id"
        :tree-props="{ children: 'children' }"
        :default-expand-all="moduleKey === 'agent-hierarchy'"
        border
        stripe
        ><el-table-column
          fixed="left"
          type="index"
          :index="(i: number) => (page - 1) * pageSize + i + 1"
          :label="t('common.index')"
          width="76"
        /><el-table-column
          v-for="col in config.columns"
          :key="col.key"
          :prop="col.key"
          :label="label(col.label)"
          :width="col.width"
          show-overflow-tooltip
          ><template #default="{ row }"
            ><UserStatusTag
              v-if="col.status"
              :status="String(row[col.key] || 'UNKNOWN')"
            /><el-switch
              v-else-if="col.boolean"
              :model-value="Boolean(row[col.key])"
              disabled
            /><span v-else-if="col.time">{{
              formatBackendTime(String(row[col.key] || ""))
            }}</span
            ><span v-else>{{ row[col.key] ?? "-" }}</span></template
          ></el-table-column
        ><el-table-column fixed="right" :label="t('common.actions')" width="260"
          ><template #default="{ row }"
            ><el-button link type="primary" @click="view(row)">{{
              t("common.view")
            }}</el-button
            ><el-button
              v-if="config.updatePermission && can(config.updatePermission)"
              link
              @click="openEdit(row)"
              >{{ t("common.edit") }}</el-button
            ><el-dropdown trigger="click"
              ><el-button link>{{ t("agentCenter.actions.more") }}</el-button
              ><template #dropdown
                ><el-dropdown-menu
                  ><template v-if="moduleKey === 'agent-list'"
                    ><el-dropdown-item
                      v-if="can('agent:status')"
                      @click="ask('agent-status', row)"
                      >{{
                        translated(
                          row.status === "ACTIVE" ? "DISABLE" : "ENABLE",
                        )
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:hierarchy:update')"
                      @click="ask('change-parent', row)"
                      >{{
                        t("agentCenter.actions.changeParent")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:plan:assign')"
                      @click="ask('assign-plan', row)"
                      >{{
                        t("agentCenter.actions.assignPlan")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:attribution:update')"
                      @click="ask('change-attribution', row)"
                      >{{
                        t("agentCenter.actions.changeAttribution")
                      }}</el-dropdown-item
                    ></template
                  ><template v-else-if="moduleKey === 'agent-hierarchy'"
                    ><el-dropdown-item
                      v-if="can('agent:hierarchy:update')"
                      @click="ask('change-parent', row)"
                      >{{
                        t("agentCenter.actions.changeParent")
                      }}</el-dropdown-item
                    ></template
                  ><template v-else-if="moduleKey === 'channel-management'"
                    ><el-dropdown-item @click="copyRegistrationURL(row)">{{
                      t("agentCenter.actions.copyRegistrationUrl")
                    }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:channel:status')"
                      @click="ask('channel-status', row)"
                      >{{
                        translated(row.enabled ? "DISABLE" : "ENABLE")
                      }}</el-dropdown-item
                    ></template
                  ><template v-else-if="moduleKey === 'commission-plans'"
                    ><el-dropdown-item
                      v-if="can('agent:commission-plan:create')"
                      @click="openCopy(row)"
                      >{{ t("agentCenter.actions.copy") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:commission-plan:update')"
                      @click="ask('plan-status', row)"
                      >{{
                        translated(row.enabled ? "DISABLE" : "ENABLE")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:commission-plan:delete')"
                      divided
                      @click="ask('delete-plan', row)"
                      >{{ t("common.delete") }}</el-dropdown-item
                    ></template
                  ><template v-else-if="moduleKey === 'agent-settlements'"
                    ><el-dropdown-item
                      v-if="
                        row.status === 'DRAFT' && can('agent:settlement:submit')
                      "
                      @click="ask('submit', row)"
                      >{{ t("agentCenter.actions.submit") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="
                        row.status === 'PENDING_REVIEW' &&
                        can('agent:settlement:approve')
                      "
                      @click="ask('approve', row)"
                      >{{ t("agentCenter.actions.approve") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="
                        row.status === 'PENDING_REVIEW' &&
                        can('agent:settlement:reject')
                      "
                      @click="ask('reject', row)"
                      >{{ t("agentCenter.actions.reject") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="
                        row.status === 'APPROVED' && can('agent:settlement:pay')
                      "
                      @click="ask('mark-paid', row)"
                      >{{ t("agentCenter.actions.markPaid") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="
                        ['DRAFT', 'REJECTED'].includes(String(row.status)) &&
                        can('agent:settlement:cancel')
                      "
                      @click="ask('cancel', row)"
                      >{{ t("common.cancel") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:settlement:export')"
                      @click="exportStatement(row)"
                      >{{
                        t("agentCenter.actions.exportStatement")
                      }}</el-dropdown-item
                    ></template
                  ><template v-else-if="moduleKey === 'agent-wallets'"
                    ><el-dropdown-item
                      v-if="can('agent:wallet:freeze')"
                      @click="ask('wallet-freeze', row)"
                      >{{
                        translated(
                          row.status === "FROZEN" ? "UNFREEZE" : "FREEZE",
                        )
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('agent:wallet:adjust')"
                      @click="ask('wallet-adjust', row)"
                      >{{
                        t("agentCenter.actions.manualAdjustment")
                      }}</el-dropdown-item
                    ></template
                  ></el-dropdown-menu
                ></template
              ></el-dropdown
            ></template
          ></el-table-column
        ></el-table
      >
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
          @current-change="load"
          @size-change="
            () => {
              page = 1;
              load();
            }
          "
        /></div
    ></UserCenterCard>

    <el-dialog
      v-model="detailVisible"
      :title="t('agentCenter.dialogs.detail')"
      width="min(1080px,94vw)"
      destroy-on-close
      ><div v-loading="detailLoading">
        <el-tabs v-model="activeTab" type="card"
          ><el-tab-pane
            v-for="tab in config.tabs"
            :key="tab"
            :name="tab"
            :label="t(`agentCenter.tabs.${tab}`)"
            ><template v-if="Array.isArray(tabValue(tab))"
              ><el-table :data="tabValue(tab)" max-height="430"
                ><el-table-column
                  v-for="k in arrayColumns(tabValue(tab))"
                  :key="k"
                  :prop="k"
                  :label="label(k)"
                  min-width="140"
                  show-overflow-tooltip
                  ><template #default="{ row }">{{
                    display(k, row[k])
                  }}</template></el-table-column
                ></el-table
              ></template
            ><el-descriptions
              v-else-if="isObject(tabValue(tab))"
              :column="2"
              border
              ><el-descriptions-item
                v-for="[k, v] in entries(tabValue(tab))"
                :key="k"
                :label="label(k)"
                >{{ display(k, v) }}</el-descriptions-item
              ></el-descriptions
            ><el-empty
              v-else
              :description="t('agentCenter.messages.empty')" /></el-tab-pane
        ></el-tabs></div
    ></el-dialog>

    <el-dialog
      v-model="editVisible"
      :title="t(editID ? 'common.edit' : 'common.create')"
      width="min(880px,94vw)"
      destroy-on-close
      ><el-form
        v-loading="editLoading"
        label-position="left"
        label-width="170px"
        ><template v-if="moduleKey === 'agent-list'"
          ><el-form-item :label="label('agentCode')"><el-input v-model="form.agentCode" :disabled="Boolean(editID)" /></el-form-item>
          <el-form-item :label="label('name')"><el-input v-model="form.name" /></el-form-item>
          <el-form-item :label="label('playerNo')"><el-select v-model="form.userId" filterable remote clearable :remote-method="searchPlayers" :loading="playerSearching" :disabled="Boolean(editID)"><el-option v-for="item in playerChoices" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
          <el-form-item :label="label('parentAgent')"><el-select v-model="form.parentAgentId" filterable clearable :disabled="Boolean(editID)"><el-option v-for="item in agentChoices.filter(agent => agent.id !== editID)" :key="item.id" :label="agentChoiceLabel(item)" :value="item.id" /></el-select></el-form-item>
          <el-form-item :label="label('country')"><el-select v-model="form.country" filterable><el-option v-for="item in optionList(adminOptions, 'countries')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
          <el-form-item :label="label('currency')"><el-select v-model="form.currency" filterable><el-option v-for="item in optionList(adminOptions, 'currencies')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
          <el-form-item :label="label('commissionPlan')"><el-select v-model="form.commissionPlanId" filterable clearable :disabled="Boolean(editID)"><el-option v-for="item in planChoices" :key="item.id" :label="planChoiceLabel(item)" :value="item.id" /></el-select></el-form-item></template
        ><template v-else-if="moduleKey === 'channel-management'"
          ><el-form-item :label="label('code')"><el-input v-model="form.code" :disabled="Boolean(editID)" /></el-form-item>
          <el-form-item :label="label('nameZh')"><el-input v-model="form.nameZh" /></el-form-item>
          <el-form-item :label="label('nameEn')"><el-input v-model="form.nameEn" /></el-form-item>
          <el-form-item :label="label('agent')"><el-select v-model="form.agentId" filterable><el-option v-for="item in agentChoices" :key="item.id" :label="agentChoiceLabel(item)" :value="item.id" /></el-select></el-form-item>
          <el-form-item :label="label('country')"><el-select v-model="form.country" filterable><el-option v-for="item in optionList(adminOptions, 'countries')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
          <el-form-item :label="label('platform')"><el-select v-model="form.platform"><el-option v-for="item in ['H5','WEB','APP']" :key="item" :label="item" :value="item" /></el-select></el-form-item>
          <el-form-item :label="label('registrationUrl')"><el-input v-model="form.registrationUrl" /></el-form-item>
          <el-form-item :label="label('enabled')"
            ><el-switch v-model="form.enabled" /></el-form-item></template
        ><template v-else
          ><el-form-item
            v-for="k in [
              'code',
              'nameZh',
              'nameEn',
              'currency',
              'cpaAmount',
              'revenueShareRate',
              'minimumSettlementAmount',
            ]"
            :key="k"
            :label="label(k)"
            ><el-input
              v-model="form[k]"
              :disabled="Boolean(editID && k === 'code')" /></el-form-item
          ><el-form-item :label="label('model')"
            ><el-select v-model="form.model"
              ><el-option
                v-for="o in [
                  'CPA',
                  'REVENUE_SHARE',
                  'HYBRID',
                  'TIERED_REVENUE_SHARE',
                ]"
                :key="o"
                :label="translated(o)"
                :value="o" /></el-select></el-form-item
          ><el-form-item :label="label('settlementCycle')"
            ><el-select v-model="form.settlementCycle"
              ><el-option
                v-for="o in ['DAILY', 'WEEKLY', 'MONTHLY']"
                :key="o"
                :label="translated(o)"
                :value="o" /></el-select></el-form-item
          ><el-form-item :label="label('negativeCarryover')"
            ><el-switch v-model="form.negativeCarryover" /></el-form-item
          ><el-form-item :label="label('enabled')"
            ><el-switch v-model="form.enabled" /></el-form-item></template
        ><el-form-item :label="label('reason')"
          ><el-input
            v-model="form.reason"
            type="textarea" /></el-form-item></el-form
      ><template #footer
        ><el-button @click="editVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button type="primary" :loading="editLoading" @click="save">{{
          t("common.save")
        }}</el-button></template
      ></el-dialog
    >

    <el-dialog
      v-model="operationVisible"
      :title="
        t(
          `agentCenter.dialogs.${operation.type === 'generate' ? 'generate' : 'operation'}`,
        )
      "
      width="520px"
      ><el-form label-position="left" label-width="170px"
        ><template v-if="operation.type === 'generate'"
          ><el-form-item :label="label('agent')"
            ><el-select v-model="operation.agentId" filterable><el-option v-for="item in agentChoices" :key="item.id" :label="agentChoiceLabel(item)" :value="String(item.id)" /></el-select></el-form-item
          ><el-form-item :label="label('settlementPeriod')"
            ><el-date-picker
              v-model="operation.periodRange"
              type="daterange"
              value-format="YYYY-MM-DD"
              :start-placeholder="t('agentCenter.placeholders.start')"
              :end-placeholder="
                t('agentCenter.placeholders.end')
              " /></el-form-item
          ><el-form-item :label="label('currency')"
            ><el-select v-model="operation.currency" filterable><el-option v-for="item in optionList(adminOptions, 'currencies')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item></template
        ><el-form-item
          v-else-if="['change-parent', 'assign-plan'].includes(operation.type)"
          :label="
            label(
              operation.type === 'change-parent'
                ? 'parentAgentId'
                : 'commissionPlanId',
            )
          "
          ><el-select v-if="operation.type === 'change-parent'" v-model="operation.value" clearable filterable><el-option v-for="item in agentChoices.filter(agent => agent.id !== selected?.id)" :key="item.id" :label="agentChoiceLabel(item)" :value="String(item.id)" /></el-select><el-select v-else v-model="operation.value" filterable><el-option v-for="item in planChoices" :key="item.id" :label="planChoiceLabel(item)" :value="String(item.id)" /></el-select></el-form-item
        ><template v-else-if="operation.type === 'change-attribution'"
          ><el-form-item :label="label('playerNo')"
            ><el-select v-model="operation.value" filterable remote :remote-method="searchPlayers" :loading="playerSearching"><el-option v-for="item in playerChoices" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item
          ><el-form-item :label="label('channel')"
            ><el-select v-model="operation.channelId" clearable filterable><el-option v-for="item in channelChoices" :key="item.id" :label="channelChoiceLabel(item)" :value="String(item.id)" /></el-select></el-form-item
          ><el-form-item :label="label('attributionSource')"
            ><el-select v-model="operation.source"><el-option v-for="item in ['MANUAL','REGISTRATION','IMPORT']" :key="item" :label="translated(item)" :value="item" /></el-select></el-form-item></template
        ><template v-else-if="operation.type === 'wallet-adjust'"
          ><el-form-item :label="label('transactionType')"
            ><el-select v-model="operation.adjustmentType"
              ><el-option
                v-for="o in ['MANUAL_CREDIT', 'MANUAL_DEBIT']"
                :key="o"
                :label="translated(o)"
                :value="o" /></el-select></el-form-item
          ><el-form-item :label="label('amount')"
            ><el-input v-model="operation.amount" /></el-form-item></template
        ><el-form-item :label="label('reason')"
          ><el-input
            v-model="operation.reason"
            type="textarea" /></el-form-item></el-form
      ><template #footer
        ><el-button @click="operationVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="operationLoading"
          @click="operation.type === 'generate' ? generate() : runOperation()"
          >{{ t("common.confirm") }}</el-button
        ></template
      ></el-dialog
    >
  </div>
</template>

<style scoped>
.agent-page {
  display: grid;
  gap: 16px;
}
.search-form :deep(.el-select),
.search-form :deep(.el-date-editor) {
  width: 100%;
}
.el-dialog :deep(.el-date-editor),
.el-dialog :deep(.el-select) {
  width: 100%;
}
.actions,
.pagination {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.pagination {
  padding-top: 18px;
}
:deep(.el-descriptions__label) {
  width: 180px;
}
:deep(.el-table .cell) {
  white-space: nowrap;
}
</style>
