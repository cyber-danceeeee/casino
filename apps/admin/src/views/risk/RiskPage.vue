<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { businessColumns, businessEntries, sanitizeBusinessValue } from "@/utils/presentation";
import {
  createRisk,
  deleteRisk,
  getRisk,
  listRisk,
  riskPatch,
  riskPost,
  updateRisk,
  type RiskDetail,
  type RiskPayload,
  type RiskQuery,
  type RiskRecord,
} from "@/api/adminRisk";
import { riskConfigs } from "./config";

type Obj = Record<string, unknown>;
type RuleCondition = {
  fieldName: string;
  operator: string;
  comparisonValue: string;
  windowSeconds: number;
  sortOrder: number;
};
type RuleAction = {
  actionType: string;
  parameterValue: string;
  sortOrder: number;
};
const route = useRoute(),
  router = useRouter(),
  { t, te } = useI18n(),
  { actor, can, loadPermissions } = useAdminPermissions();
const moduleKey = computed(() => String(route.meta.moduleKey || "risk-events")),
  config = computed(() => riskConfigs[moduleKey.value]);
const query = reactive<Record<string, string | string[]>>({}),
  rows = ref<RiskRecord[]>([]),
  total = ref(0),
  page = ref(1),
  pageSize = ref(20),
  loading = ref(false);
const detailVisible = ref(false),
  detailLoading = ref(false),
  detail = ref<RiskDetail>({}),
  activeTab = ref("");
const editVisible = ref(false),
  editLoading = ref(false),
  editingID = ref(0),
  form = reactive<Obj>({}),
  conditions = ref<RuleCondition[]>([]),
  actions = ref<RuleAction[]>([]);
const actionVisible = ref(false),
  actionLoading = ref(false),
  actionForm = reactive({
    action: "",
    id: 0,
    userID: 0,
    reviewerID: "",
    reviewerName: "",
    decision: "",
    reason: "",
    internalRemark: "",
    riskTags: "",
    content: "",
    riskLevel: "HIGH",
    riskType: "MANUAL",
    expiresAt: "",
  });
const detailPermissions: Record<string, string> = {
  "risk-rules": "risk:rule:list",
  "risk-events": "risk:event:view",
  "review-queue": "risk:review:view",
  "user-associations": "risk:association:view",
  blacklist: "risk:blacklist:list",
  "withdrawal-risk": "risk:withdrawal:view",
  "risk-operation-logs": "risk:operation-log:view",
};
const withdrawalDecisions = computed(() =>
  [
    ["APPROVE", "risk:withdrawal:approve"],
    ["REJECT", "risk:withdrawal:reject"],
    ["ADDITIONAL_REVIEW", "risk:withdrawal:restrict"],
    ["RESTRICT_WITHDRAWAL", "risk:withdrawal:restrict"],
    ["FREEZE_ACCOUNT", "risk:withdrawal:restrict"],
  ]
    .filter(([, permission]) => can(permission))
    .map(([decision]) => decision),
);
const reviewDecisions = computed(() => [
  ...(can("risk:review:decision")
    ? ["UNDER_REVIEW", "APPROVE", "REJECT", "CLEAR"]
    : []),
  ...(can("risk:review:escalate") ? ["ESCALATE"] : []),
]);
const visibleTabs = computed(() =>
  config.value.tabs.filter((key) => key in detail.value),
);
const isObject = (v: unknown): v is Obj =>
  Boolean(v && typeof v === "object" && !Array.isArray(v));
const valueKey = (v: unknown) =>
  `riskCenter.values.${String(v ?? "").toLowerCase()}`;
const translated = (v: unknown) =>
  te(valueKey(v)) ? t(valueKey(v)) : String(v ?? "-");
const fieldLabel = (key: string) =>
  te(`riskCenter.fields.${key}`)
    ? t(`riskCenter.fields.${key}`)
    : key.replace(/_/g, " ");
function display(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (typeof value === "boolean")
    return value ? t("common.yes") : t("common.no");
  if (typeof value === "string" && /(At|Time|_at)$/.test(key))
    return formatBackendTime(value);
  if (isObject(value) || Array.isArray(value))
    return JSON.stringify(sanitizeBusinessValue(value), null, 2);
  return translated(value);
}
function buildQuery() {
  const out: RiskQuery = { page: page.value, pageSize: pageSize.value };
  for (const f of config.value.filters) {
    const v = query[f.key];
    if (f.type === "dateRange" && Array.isArray(v) && v.length === 2) {
      const base = f.key.replace(/Range$/, "");
      out[`${base}From`] = selectedTimeToBackend(v[0] || "");
      out[`${base}To`] = selectedTimeToBackend(v[1] || "", true);
    } else if (typeof v === "string" && v.trim()) out[f.key] = v.trim();
  }
  return out;
}
async function loadRows() {
  if (!config.value || !can(config.value.listPermission)) return;
  loading.value = true;
  try {
    const result = await listRisk(config.value.resource, buildQuery());
    rows.value = result.list || [];
    total.value = result.total;
  } catch (e) {
    ElMessage.error(
      e instanceof Error ? e.message : t("riskCenter.messages.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}
function search() {
  page.value = 1;
  loadRows();
}
function reset() {
  Object.keys(query).forEach((k) => delete query[k]);
  page.value = 1;
  loadRows();
}
async function openDetail(row: RiskRecord) {
  const id = Number(
    moduleKey.value === "user-associations" ? row.userID : row.id,
  );
  if (!id) return;
  detailVisible.value = true;
  detailLoading.value = true;
  detail.value = {};
  try {
    detail.value = await getRisk(config.value.resource, id);
    activeTab.value = visibleTabs.value[0] || "";
  } catch (e) {
    ElMessage.error(
      e instanceof Error ? e.message : t("riskCenter.messages.loadFailed"),
    );
  } finally {
    detailLoading.value = false;
  }
}
function tableColumns(v: unknown) {
  return businessColumns(v, 10).filter((key) => !["valueHash", "valueLast4"].includes(key));
}
function detailEntries(value: unknown) {
  return businessEntries(value);
}
function rowIndex(index: number) {
  return (page.value - 1) * pageSize.value + index + 1;
}
function resetForm() {
  Object.keys(form).forEach((k) => delete form[k]);
  conditions.value = [];
  actions.value = [];
  if (moduleKey.value === "risk-rules") {
    Object.assign(form, {
      code: "",
      nameZh: "",
      nameEn: "",
      descriptionZh: "",
      descriptionEn: "",
      ruleType: "MANUAL",
      riskLevel: "MEDIUM",
      priority: 0,
      enabled: false,
      countries: "",
      currencies: "",
      channels: "",
      startTime: "",
      endTime: "",
    });
    conditions.value = [
      {
        fieldName: "",
        operator: "EQUAL",
        comparisonValue: "",
        windowSeconds: 0,
        sortOrder: 0,
      },
    ];
    actions.value = [
      { actionType: "CREATE_EVENT", parameterValue: "", sortOrder: 0 },
    ];
  } else
    Object.assign(form, {
      type: "IP",
      value: "",
      reason: "",
      enabled: true,
      expiresAt: "",
    });
}
function detailSource(d: RiskDetail) {
  const value = d.basicInformation || d.overview;
  return isObject(value) ? value : {};
}
async function openCreate() {
  editingID.value = 0;
  resetForm();
  editVisible.value = true;
}
async function openEdit(row: RiskRecord) {
  if (!row.id) return;
  editingID.value = row.id;
  resetForm();
  editVisible.value = true;
  editLoading.value = true;
  try {
    const data = await getRisk(config.value.resource, row.id);
    Object.assign(form, detailSource(data));
    if (Array.isArray(data.conditions))
      conditions.value = data.conditions.map((v) => {
        const x = isObject(v) ? v : {};
        return {
          fieldName: String(x.fieldName || ""),
          operator: String(x.operator || "EQUAL"),
          comparisonValue: String(x.comparisonValue || ""),
          windowSeconds: Number(x.windowSeconds || 0),
          sortOrder: Number(x.sortOrder || 0),
        };
      });
    if (Array.isArray(data.actions))
      actions.value = data.actions.map((v) => {
        const x = isObject(v) ? v : {};
        return {
          actionType: String(x.actionType || "CREATE_EVENT"),
          parameterValue: String(x.parameterValue || ""),
          sortOrder: Number(x.sortOrder || 0),
        };
      });
  } catch (e) {
    ElMessage.error(
      e instanceof Error ? e.message : t("riskCenter.messages.loadFailed"),
    );
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}
function rulePayload(): RiskPayload {
  return {
    ...form,
    startTime: form.startTime || null,
    endTime: form.endTime || null,
    conditions: conditions.value,
    actions: actions.value,
  };
}
async function submitEdit() {
  editLoading.value = true;
  try {
    const payload =
      moduleKey.value === "risk-rules"
        ? rulePayload()
        : { ...form, expiresAt: form.expiresAt || null };
    if (editingID.value)
      await updateRisk(config.value.resource, editingID.value, payload);
    else await createRisk(config.value.resource, payload);
    ElMessage.success(t("riskCenter.messages.succeeded"));
    editVisible.value = false;
    loadRows();
  } catch (e) {
    ElMessage.error(
      e instanceof Error ? e.message : t("riskCenter.messages.failed"),
    );
  } finally {
    editLoading.value = false;
  }
}
function ask(action: string, row: RiskRecord, decision = "") {
  Object.assign(actionForm, {
    action,
    id: Number(row.id || 0),
    userID: Number(row.userID || 0),
    reviewerID: action.includes("assign") ? String(actor.value?.id || "") : "",
    reviewerName: action.includes("assign") ? String(actor.value?.displayName || actor.value?.username || "") : "",
    decision,
    reason: "",
    internalRemark: "",
    riskTags: "",
    content: "",
    riskLevel: "HIGH",
    riskType: "MANUAL",
    expiresAt: "",
  });
  actionVisible.value = true;
}
async function direct(action: string, row: RiskRecord) {
  try {
    if (action === "claim") {
      await ElMessageBox.confirm(
        t("riskCenter.dialogs.claimMessage"),
        t("riskCenter.dialogs.confirmTitle"),
        { type: "warning" },
      );
      await riskPost("reviews", Number(row.id), "claim");
    } else if (action === "toggle-rule")
      await riskPatch("rules", Number(row.id), "status", {
        enabled: !Boolean(row.enabled),
        reason: t("riskCenter.messages.statusChanged"),
      });
    else if (action === "copy-rule")
      await riskPost("rules", Number(row.id), "copy");
    ElMessage.success(t("riskCenter.messages.succeeded"));
    loadRows();
  } catch (e) {
    if (e === "cancel" || e === "close") return;
    ElMessage.error(
      e instanceof Error ? e.message : t("riskCenter.messages.failed"),
    );
  }
}
async function submitAction() {
  if (!actionForm.reason.trim() && !["note"].includes(actionForm.action)) {
    ElMessage.warning(t("riskCenter.validation.reason"));
    return;
  }
  actionLoading.value = true;
  try {
    const common = {
      decision: actionForm.decision,
      reason: actionForm.reason.trim(),
      internalRemark: actionForm.internalRemark.trim(),
      riskTags: actionForm.riskTags.trim(),
    };
    switch (actionForm.action) {
      case "assign-event":
        await riskPost("events", actionForm.id, "assign", {
          reviewerID: Number(actionForm.reviewerID),
          reviewerName: actionForm.reviewerName,
          reason: actionForm.reason,
        });
        break;
      case "assign-review":
        await riskPost("reviews", actionForm.id, "assign", {
          reviewerID: Number(actionForm.reviewerID),
          reviewerName: actionForm.reviewerName,
          reason: actionForm.reason,
        });
        break;
      case "event-review":
      case "event-confirm":
      case "event-clear":
      case "event-restrict":
      case "event-block":
        await riskPost(
          "events",
          actionForm.id,
          actionForm.action.replace("event-", ""),
          common,
        );
        break;
      case "review-decision":
        await riskPost("reviews", actionForm.id, "decision", common);
        break;
      case "note":
        await riskPost(
          moduleKey.value === "review-queue" ? "reviews" : "events",
          actionForm.id,
          "notes",
          { content: actionForm.content },
        );
        break;
      case "blacklist-update":
        await updateRisk("blacklists", actionForm.id, {
          enabled: actionForm.decision === "ENABLE",
          reason: actionForm.reason,
          expiresAt: actionForm.expiresAt || null,
        });
        break;
      case "withdrawal-decision":
        await riskPost("withdrawals", actionForm.id, "decision", common);
        break;
      case "create-event":
        await createRisk("events", {
          userID: actionForm.userID,
          riskType: actionForm.riskType,
          riskLevel: actionForm.riskLevel,
          reason: actionForm.reason,
        });
        break;
      case "add-blacklist":
        await createRisk("blacklists", {
          type: actionForm.decision || "USER",
          value: String(actionForm.userID),
          reason: actionForm.reason,
          enabled: true,
        });
        break;
      case "test-rule": {
        let facts: Obj = {};
        try {
          facts = JSON.parse(actionForm.internalRemark || "{}") as Obj;
        } catch {
          throw new Error(t("riskCenter.validation.json"));
        }
        await riskPost("rules", actionForm.id, "test", { facts });
        break;
      }
      case "delete":
        await deleteRisk(
          config.value.resource,
          actionForm.id,
          actionForm.reason,
        );
        break;
    }
    ElMessage.success(t("riskCenter.messages.succeeded"));
    actionVisible.value = false;
    loadRows();
  } catch (e) {
    ElMessage.error(
      e instanceof Error ? e.message : t("riskCenter.messages.failed"),
    );
  } finally {
    actionLoading.value = false;
  }
}
function viewUser(row: RiskRecord) {
	const playerNo = String(row.playerNo || row.userNo || "").trim();
	if (playerNo) router.push(`/players/${encodeURIComponent(playerNo)}`);
}
watch(moduleKey, async () => {
  Object.keys(query).forEach((k) => delete query[k]);
  page.value = 1;
  await loadPermissions();
  loadRows();
});
onMounted(async () => {
  await loadPermissions();
  loadRows();
});
</script>

<template>
  <div v-if="config" class="risk-page">
    <UserCenterCard :title="t(`riskCenter.pages.${moduleKey}.search`)"
      ><el-form
        class="risk-search"
        label-position="left"
        @submit.prevent="search"
        ><el-form-item
          v-for="f in config.filters"
          :key="f.key"
          :label="fieldLabel(f.key)"
          ><el-select
            v-if="f.type === 'select'"
            v-model="query[f.key]"
            clearable
            :placeholder="t('riskCenter.placeholders.select')"
            ><el-option
              v-for="o in f.options"
              :key="o"
              :label="translated(o)"
              :value="o" /></el-select
          ><el-date-picker
            v-else-if="f.type === 'dateRange'"
            v-model="query[f.key]"
            type="daterange"
            value-format="YYYY-MM-DD"
            :start-placeholder="t('riskCenter.placeholders.startDate')"
            :end-placeholder="t('riskCenter.placeholders.endDate')" /><el-input
            v-else
            v-model="query[f.key]"
            clearable
            :type="f.type === 'number' ? 'number' : 'text'"
            :placeholder="t('riskCenter.placeholders.enter')"
            @keyup.enter="search"
        /></el-form-item>
        <div class="risk-search__actions">
          <el-button type="primary" @click="search">{{
            t("common.search")
          }}</el-button
          ><el-button @click="reset">{{ t("common.reset") }}</el-button>
        </div></el-form
      ></UserCenterCard
    ><UserCenterCard :title="t(`riskCenter.pages.${moduleKey}.title`)"
      ><template #actions
        ><el-button
          v-if="config.createPermission && can(config.createPermission)"
          type="primary"
          @click="openCreate"
          >{{ t("common.create") }}</el-button
        ></template
      ><el-table v-loading="loading" :data="rows" row-key="id" stripe
        ><el-table-column
          type="index"
          fixed="left"
          :label="t('common.index')"
          width="72"
          :index="rowIndex"
        /><el-table-column
          v-for="c in config.columns"
          :key="c.key"
          :prop="c.key"
          :label="fieldLabel(c.key)"
          :width="c.width"
          show-overflow-tooltip
          ><template #default="{ row }"
            ><UserStatusTag
              v-if="c.status"
              :value="String(row[c.key] ?? '')"
              :label="translated(row[c.key])"
            /><span v-else-if="c.time">{{
              row[c.key] ? formatBackendTime(String(row[c.key])) : "-"
            }}</span
            ><span v-else>{{ row[c.key] ?? "-" }}</span></template
          ></el-table-column
        ><el-table-column fixed="right" :label="t('common.actions')" width="310"
          ><template #default="{ row }"
            ><el-button
              v-if="can(detailPermissions[moduleKey])"
              link
              type="primary"
              @click="openDetail(row)"
              >{{ t("common.view") }}</el-button
            ><template v-if="moduleKey === 'risk-rules'"
              ><el-button
                v-if="can('risk:rule:update')"
                link
                @click="openEdit(row)"
                >{{ t("common.edit") }}</el-button
              ><el-button
                v-if="can('risk:rule:status')"
                link
                @click="direct('toggle-rule', row)"
                >{{
                  row.enabled ? t("common.disable") : t("common.enable")
                }}</el-button
              ><el-button
                v-if="can('risk:rule:test')"
                link
                @click="ask('test-rule', row)"
                >{{ t("riskCenter.actions.test") }}</el-button
              ><el-button
                v-if="can('risk:rule:create')"
                link
                @click="direct('copy-rule', row)"
                >{{ t("riskCenter.actions.copy") }}</el-button
              ><el-button
                v-if="can('risk:rule:delete')"
                link
                type="danger"
                @click="ask('delete', row)"
                >{{ t("common.delete") }}</el-button
              ></template
            ><template v-else-if="moduleKey === 'risk-events'"
              ><el-dropdown trigger="click"
                ><el-button link type="primary">{{
                  t("common.more")
                }}</el-button
                ><template #dropdown
                  ><el-dropdown-menu
                    ><el-dropdown-item
                      v-if="can('risk:event:assign')"
                      @click="ask('assign-event', row)"
                      >{{ t("riskCenter.actions.assign") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('risk:event:review')"
                      @click="ask('event-review', row, 'UNDER_REVIEW')"
                      >{{ t("riskCenter.actions.review") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('risk:event:review')"
                      @click="ask('event-confirm', row, 'CONFIRMED')"
                      >{{
                        t("riskCenter.actions.confirmRisk")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('risk:event:clear')"
                      @click="ask('event-clear', row, 'CLEARED')"
                      >{{ t("riskCenter.actions.clear") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('risk:event:restrict')"
                      @click="ask('event-restrict', row, 'RESTRICTED')"
                      >{{ t("riskCenter.actions.restrict") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('risk:event:block')"
                      @click="ask('event-block', row, 'BLOCKED')"
                      >{{ t("riskCenter.actions.block") }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('risk:event:note')"
                      @click="ask('note', row)"
                      >{{ t("riskCenter.actions.note") }}</el-dropdown-item
                    ></el-dropdown-menu
                  ></template
                ></el-dropdown
              ></template
            ><template v-else-if="moduleKey === 'review-queue'"
              ><el-button
                v-if="can('risk:review:claim')"
                link
                @click="direct('claim', row)"
                >{{ t("riskCenter.actions.claim") }}</el-button
              ><el-button
                v-if="can('risk:review:assign')"
                link
                @click="ask('assign-review', row)"
                >{{ t("riskCenter.actions.assign") }}</el-button
              ><el-button
                v-if="
                  can('risk:review:decision') || can('risk:review:escalate')
                "
                link
                @click="ask('review-decision', row)"
                >{{ t("riskCenter.actions.decision") }}</el-button
              ><el-button
                v-if="can('risk:review:decision')"
                link
                @click="ask('note', row)"
                >{{ t("riskCenter.actions.note") }}</el-button
              ></template
            ><template v-else-if="moduleKey === 'user-associations'"
              ><el-button
                v-if="can('risk:event:review')"
                link
                @click="ask('create-event', row)"
                >{{ t("riskCenter.actions.createEvent") }}</el-button
              ><el-button
                v-if="can('risk:blacklist:create')"
                link
                @click="ask('add-blacklist', row, 'USER')"
                >{{ t("riskCenter.actions.addBlacklist") }}</el-button
              ></template
            ><template v-else-if="moduleKey === 'blacklist'"
              ><el-button
                v-if="can('risk:blacklist:update')"
                link
                @click="
                  ask(
                    'blacklist-update',
                    row,
                    row.status === 'ENABLED' ? 'DISABLE' : 'ENABLE',
                  )
                "
                >{{
                  row.status === "ENABLED"
                    ? t("common.disable")
                    : t("common.enable")
                }}</el-button
              ><el-button
                v-if="can('risk:blacklist:remove')"
                link
                type="danger"
                @click="ask('delete', row)"
                >{{ t("riskCenter.actions.remove") }}</el-button
              ></template
            ><template v-else-if="moduleKey === 'withdrawal-risk'"
              ><el-button
                v-if="withdrawalDecisions.length"
                link
                @click="ask('withdrawal-decision', row)"
                >{{ t("riskCenter.actions.decision") }}</el-button
              ><el-button link @click="viewUser(row)">{{
                t("riskCenter.actions.viewUser")
              }}</el-button></template
            ></template
          ></el-table-column
        ></el-table
      >
      <div class="risk-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="total"
          @current-change="loadRows"
          @size-change="
            () => {
              page = 1;
              loadRows();
            }
          "
        /></div></UserCenterCard
    ><el-dialog
      v-model="detailVisible"
      :title="t('riskCenter.dialogs.detail')"
      width="min(1080px,94vw)"
      destroy-on-close
      ><div v-loading="detailLoading">
        <el-tabs v-model="activeTab" type="card"
          ><el-tab-pane
            v-for="tab in visibleTabs"
            :key="tab"
            :name="tab"
            :label="t(`riskCenter.tabs.${tab}`)"
            ><template v-if="Array.isArray(detail[tab])"
              ><el-table :data="detail[tab]" max-height="430"
                ><el-table-column
                  v-for="c in tableColumns(detail[tab])"
                  :key="c"
                  :prop="c"
                  :label="fieldLabel(c)"
                  min-width="130"
                  show-overflow-tooltip /></el-table></template
            ><template v-else-if="isObject(detail[tab])"
              ><div class="detail-grid">
                <div v-for="[k, v] in detailEntries(detail[tab])" :key="k" class="detail-item">
                  <span>{{ fieldLabel(k) }}</span>
                  <pre>{{ display(k, v) }}</pre>
                </div>
              </div></template
            ><el-empty
              v-else
              :description="t('riskCenter.messages.empty')" /></el-tab-pane
        ></el-tabs></div></el-dialog
    ><el-dialog
      v-model="editVisible"
      :title="t(editingID ? 'common.edit' : 'common.create')"
      width="min(920px,94vw)"
      destroy-on-close
      ><el-form
        v-loading="editLoading"
        label-width="150px"
        label-position="left"
        ><template v-if="moduleKey === 'risk-rules'"
          ><div class="form-grid">
            <el-form-item
              v-for="key in [
                'code',
                'nameZh',
                'nameEn',
                'ruleType',
                'riskLevel',
                'priority',
                'countries',
                'currencies',
                'channels',
                'startTime',
                'endTime',
              ]"
              :key="key"
              :label="fieldLabel(key)"
              ><el-select v-if="key === 'ruleType'" v-model="form[key]"
                ><el-option
                  v-for="o in [
                    'REGISTRATION',
                    'LOGIN',
                    'DEVICE',
                    'IP',
                    'DEPOSIT',
                    'WITHDRAWAL',
                    'BETTING',
                    'BONUS',
                    'PROVIDER_CALLBACK',
                    'MANUAL',
                  ]"
                  :key="o"
                  :label="translated(o)"
                  :value="o" /></el-select
              ><el-select v-else-if="key === 'riskLevel'" v-model="form[key]"
                ><el-option
                  v-for="o in ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL']"
                  :key="o"
                  :label="translated(o)"
                  :value="o" /></el-select
              ><el-date-picker
                v-else-if="key === 'startTime' || key === 'endTime'"
                v-model="form[key]"
                type="datetime"
                value-format="YYYY-MM-DDTHH:mm:ssZ" /><el-input-number
                v-else-if="key === 'priority'"
                v-model="form[key]" /><el-input
                v-else
                v-model="form[key]" /></el-form-item
            ><el-form-item :label="fieldLabel('enabled')"
              ><el-switch v-model="form.enabled"
            /></el-form-item>
          </div>
          <el-divider>{{ t("riskCenter.tabs.conditions") }}</el-divider>
          <div v-for="(c, i) in conditions" :key="i" class="rule-row">
            <el-input
              v-model="c.fieldName"
              :placeholder="fieldLabel('fieldName')"
            /><el-select v-model="c.operator"
              ><el-option
                v-for="o in [
                  'EQUAL',
                  'NOT_EQUAL',
                  'GREATER_THAN',
                  'GREATER_THAN_OR_EQUAL',
                  'LESS_THAN',
                  'LESS_THAN_OR_EQUAL',
                  'IN',
                  'NOT_IN',
                  'CONTAINS',
                  'COUNT_WITHIN',
                  'SUM_WITHIN',
                ]"
                :key="o"
                :label="translated(o)"
                :value="o" /></el-select
            ><el-input
              v-model="c.comparisonValue"
              :placeholder="fieldLabel('comparisonValue')"
            /><el-input-number v-model="c.windowSeconds" /><el-button
              type="danger"
              plain
              @click="conditions.splice(i, 1)"
              >{{ t("common.remove") }}</el-button
            >
          </div>
          <el-button
            @click="
              conditions.push({
                fieldName: '',
                operator: 'EQUAL',
                comparisonValue: '',
                windowSeconds: 0,
                sortOrder: conditions.length,
              })
            "
            >{{ t("riskCenter.actions.addCondition") }}</el-button
          ><el-divider>{{ t("riskCenter.tabs.actions") }}</el-divider>
          <div v-for="(a, i) in actions" :key="i" class="rule-row">
            <el-select v-model="a.actionType"
              ><el-option
                v-for="o in [
                  'CREATE_EVENT',
                  'ADD_RISK_TAG',
                  'REQUIRE_REVIEW',
                  'DISABLE_LOGIN',
                  'DISABLE_GAME',
                  'DISABLE_WITHDRAWAL',
                  'FREEZE_ACCOUNT',
                  'BLOCK_DEVICE',
                  'BLOCK_IP',
                ]"
                :key="o"
                :label="translated(o)"
                :value="o" /></el-select
            ><el-input
              v-model="a.parameterValue"
              :placeholder="fieldLabel('parameterValue')"
            /><el-button type="danger" plain @click="actions.splice(i, 1)">{{
              t("common.remove")
            }}</el-button>
          </div>
          <el-button
            @click="
              actions.push({
                actionType: 'CREATE_EVENT',
                parameterValue: '',
                sortOrder: actions.length,
              })
            "
            >{{ t("riskCenter.actions.addAction") }}</el-button
          ></template
        ><template v-else
          ><el-form-item :label="fieldLabel('type')"
            ><el-select v-model="form.type"
              ><el-option
                v-for="o in [
                  'USER',
                  'DEVICE',
                  'IP',
                  'EMAIL',
                  'MOBILE',
                  'BANK_ACCOUNT',
                  'WALLET_ADDRESS',
                  'PAYMENT_ACCOUNT',
                ]"
                :key="o"
                :label="translated(o)"
                :value="o" /></el-select></el-form-item
          ><el-form-item :label="fieldLabel('value')"
            ><el-input v-model="form.value" /></el-form-item
          ><el-form-item :label="fieldLabel('reason')"
            ><el-input v-model="form.reason" type="textarea" /></el-form-item
          ><el-form-item :label="fieldLabel('enabled')"
            ><el-switch v-model="form.enabled" /></el-form-item
          ><el-form-item :label="fieldLabel('expiresAt')"
            ><el-date-picker
              v-model="form.expiresAt"
              type="datetime"
              value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item></template></el-form
      ><template #footer
        ><el-button @click="editVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button type="primary" :loading="editLoading" @click="submitEdit">{{
          t("common.save")
        }}</el-button></template
      ></el-dialog
    ><el-dialog
      v-model="actionVisible"
      :title="t('riskCenter.dialogs.operation')"
      width="520px"
      ><el-form label-width="140px" label-position="left"
        ><el-form-item
          v-if="actionForm.action.includes('assign')"
          :label="fieldLabel('reviewerName')"
          ><strong>{{ actionForm.reviewerName || '-' }}</strong></el-form-item
        ><el-form-item
          v-if="actionForm.action === 'review-decision'"
          :label="fieldLabel('decision')"
          ><el-select v-model="actionForm.decision"
            ><el-option
              v-for="o in reviewDecisions"
              :key="o"
              :label="translated(o)"
              :value="o" /></el-select></el-form-item
        ><el-form-item
          v-if="actionForm.action === 'withdrawal-decision'"
          :label="fieldLabel('decision')"
          ><el-select v-model="actionForm.decision"
            ><el-option
              v-for="o in withdrawalDecisions"
              :key="o"
              :label="translated(o)"
              :value="o" /></el-select></el-form-item
        ><el-form-item
          v-if="actionForm.action === 'create-event'"
          :label="fieldLabel('riskType')"
          ><el-input v-model="actionForm.riskType" /></el-form-item
        ><el-form-item
          v-if="actionForm.action === 'create-event'"
          :label="fieldLabel('riskLevel')"
          ><el-select v-model="actionForm.riskLevel"
            ><el-option
              v-for="o in ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL']"
              :key="o"
              :label="translated(o)"
              :value="o" /></el-select></el-form-item
        ><el-form-item
          v-if="actionForm.action === 'note'"
          :label="fieldLabel('content')"
          ><el-input
            v-model="actionForm.content"
            type="textarea" /></el-form-item
        ><template v-else
          ><el-form-item :label="fieldLabel('reason')"
            ><el-input
              v-model="actionForm.reason"
              type="textarea" /></el-form-item
          ><el-form-item
            v-if="actionForm.action === 'blacklist-update'"
            :label="fieldLabel('expiresAt')"
            ><el-date-picker
              v-model="actionForm.expiresAt"
              type="datetime"
              value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item
          ><el-form-item
            v-if="
              ['review-decision', 'withdrawal-decision'].includes(
                actionForm.action,
              )
            "
            :label="fieldLabel('internalRemark')"
            ><el-input
              v-model="actionForm.internalRemark"
              type="textarea" /></el-form-item
          ><el-form-item
            v-if="
              ['review-decision', 'withdrawal-decision'].includes(
                actionForm.action,
              )
            "
            :label="fieldLabel('riskTags')"
            ><el-input v-model="actionForm.riskTags" /></el-form-item
          ><el-form-item
            v-if="actionForm.action === 'test-rule'"
            :label="fieldLabel('testFacts')"
            ><el-input
              v-model="actionForm.internalRemark"
              type="textarea"
              placeholder='{"amount":100}' /></el-form-item></template></el-form
      ><template #footer
        ><el-button @click="actionVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="actionLoading"
          @click="submitAction"
          >{{ t("common.confirm") }}</el-button
        ></template
      ></el-dialog
    >
  </div>
</template>

<style scoped>
.risk-page {
  display: grid;
  gap: 16px;
}
.risk-search {
  display: grid;
  grid-template-columns: repeat(4, minmax(210px, 1fr));
  gap: 14px 18px;
}
.risk-search :deep(.el-form-item) {
  margin: 0;
}
.risk-search :deep(.el-form-item__label) {
  width: 116px;
}
.risk-search__actions {
  grid-column: 1/-1;
  display: flex;
  justify-content: flex-end;
}
.risk-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}
.detail-item {
  min-width: 0;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 10px 12px;
}
.detail-item > span {
  display: block;
  color: #909399;
  font-size: 12px;
  margin-bottom: 6px;
}
.detail-item pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font: inherit;
  color: #303133;
}
.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0 20px;
}
.rule-row {
  display: grid;
  grid-template-columns: 1fr 1.25fr 1.4fr 130px auto;
  gap: 8px;
  margin-bottom: 10px;
}
@media (max-width: 1100px) {
  .risk-search {
    grid-template-columns: repeat(2, minmax(240px, 1fr));
  }
}
@media (max-width: 720px) {
  .risk-search,
  .detail-grid,
  .form-grid {
    grid-template-columns: 1fr;
  }
  .rule-row {
    grid-template-columns: 1fr;
  }
}
</style>
