<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import {
  businessColumns,
  businessEntries,
  isInternalIdentifierKey,
  sanitizeBusinessValue,
} from "@/utils/presentation";
import {
  fetchAdminOptions,
  optionList,
  stringOptions,
  type AdminOption,
  type AdminOptionSource,
  type AdminOptions,
} from "@/api/adminOptions";
import {
  addAbnormalNote,
  assignAbnormalBet,
  clearAbnormalBet,
  investigateCallback,
  investigateRollback,
  markBetForReview,
  restrictAbnormalBet,
  reviewAbnormalBet,
  type BettingDetail,
  type BettingQuery,
  type BettingRecord,
} from "@/api/adminBetting";
import { bettingConfigs } from "./config";

type RecordValue = Record<string, unknown>;
type PendingAction = {
  type: string;
  titleKey: string;
  messageKey: string;
  permission: string;
  row: BettingRecord;
};

const route = useRoute();
const router = useRouter();
const { t, te } = useI18n();
const { actor, can, loadPermissions } = useAdminPermissions();
const moduleKey = computed(() => String(route.meta.moduleKey || "bet-records"));
const config = computed(() => bettingConfigs[moduleKey.value]);
const visibleFilters = computed(
  () => config.value?.filters.filter((item) => !isInternalIdentifierKey(item.key)) || [],
);
const visibleColumns = computed(
  () => config.value?.columns.filter((item) => !isInternalIdentifierKey(item.key)) || [],
);
const query = reactive<Record<string, string | string[]>>({});
const rows = ref<BettingRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const adminOptions = ref<AdminOptions>({});
const loading = ref(false);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<BettingDetail>({});
const activeDetailTab = ref("");
const actionVisible = ref(false);
const actionLoading = ref(false);
const pendingAction = ref<PendingAction | null>(null);
const actionForm = reactive({
  reason: "",
  reviewStatus: "UNDER_REVIEW",
  note: "",
});

const detailTabs = computed(() =>
  (config.value?.detailTabs || []).filter((tab) => tab.key in detail.value),
);

function optionLabel(value: string) {
  const key = `betting.values.${value.toLowerCase()}`;
  return te(key) ? t(key) : value;
}
const dynamicFilterSources: Record<string, AdminOptionSource> = {
  providerCode: "providers",
  gameType: "gameTypes",
  currency: "currencies",
};
function filterSource(field: { key: string; source?: AdminOptionSource }) {
  return field.source || dynamicFilterSources[field.key];
}
function filterOptions(field: { key: string; options?: string[]; source?: AdminOptionSource }) {
  const dynamic = optionList(adminOptions.value, filterSource(field));
  return dynamic.length ? dynamic : stringOptions(field.options);
}
function filterOptionLabel(item: AdminOption) {
  return item.label && item.label !== item.value ? item.label : optionLabel(item.value);
}
function isSelectFilter(field: { key: string; type?: string; source?: AdminOptionSource }) {
  return field.type === "select" || Boolean(filterSource(field));
}

function detailLabel(key: string) {
  const normalized = key.replace(/_([a-z])/g, (_, letter: string) =>
    letter.toUpperCase(),
  );
  const translationKey = `betting.fields.${normalized}`;
  return te(translationKey) ? t(translationKey) : normalized;
}

function shouldRenderTag(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return false;
  return [
    "signature_result",
    "signatureResult",
    "timestamp_validation",
    "timestampValidation",
    "processing_result",
    "processingResult",
    "status",
    "review_status",
    "risk_level",
  ].includes(key);
}

function tagTypeOverride(key: string, value: unknown) {
  if (["signatureResult", "signature_result"].includes(key)) {
    return String(value || "").toUpperCase() === "VALID" ? "success" : "danger";
  }
  return "";
}
function profitTagType(value: unknown) {
  const amount = Number(value || 0);
  if (amount < 0) return "danger";
  if (amount > 0) return "success";
  return "info";
}

function isObject(value: unknown): value is RecordValue {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function displayValue(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (typeof value === "boolean")
    return value ? t("common.yes") : t("common.no");
  if (
    (/(At|Time|_at)$/.test(key) || key.toLowerCase() === "timestamp") &&
    typeof value === "string"
  )
    return formatBackendTime(value);
  if (isObject(value) || Array.isArray(value))
    return JSON.stringify(sanitizeBusinessValue(value), null, 2);
  return String(value);
}

function arrayColumns(value: unknown) {
  return businessColumns(value).filter(
    (key) => !["raw_request", "raw_response"].includes(key),
  );
}

function buildParams(): BettingQuery {
  const params: BettingQuery = { page: page.value, pageSize: pageSize.value };
  for (const field of config.value.filters) {
    const value = query[field.key];
    if (
      field.type === "dateRange" &&
      Array.isArray(value) &&
      value.length === 2
    ) {
      const prefix = field.key.replace(/Range$/, "");
      params[`${prefix}From`] = selectedTimeToBackend(value[0] || "");
      params[`${prefix}To`] = selectedTimeToBackend(value[1] || "", true);
    } else if (typeof value === "string" && value.trim()) {
      params[field.key] = value.trim();
    }
  }
  return params;
}

async function load() {
  if (!config.value) return;
  loading.value = true;
  try {
    const result = await config.value.loader(buildParams());
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : t("betting.messages.loadFailed"),
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
  Object.keys(query).forEach((key) => delete query[key]);
  page.value = 1;
  void load();
}
function indexMethod(index: number) {
  return (page.value - 1) * pageSize.value + index + 1;
}

async function openDetail(row: BettingRecord) {
  detailVisible.value = true;
  detailLoading.value = true;
  detail.value = {};
  activeDetailTab.value = config.value.detailTabs[0]?.key || "";
  try {
    detail.value = await config.value.detailLoader(row.id);
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("betting.messages.detailFailed"),
    );
  } finally {
    detailLoading.value = false;
  }
}

function ask(
  row: BettingRecord,
  type: string,
  titleKey: string,
  messageKey: string,
  permission: string,
) {
  pendingAction.value = { row, type, titleKey, messageKey, permission };
  Object.assign(actionForm, {
    reason: "",
    reviewStatus: "UNDER_REVIEW",
    note: "",
  });
  actionVisible.value = true;
}

async function confirmAction() {
  const action = pendingAction.value;
  if (!action) return;
  actionLoading.value = true;
  try {
    const id = action.row.id;
    if (action.type === "bet-review")
      await markBetForReview(id, actionForm.reason.trim());
    else if (action.type === "rollback-investigate")
      await investigateRollback(id, actionForm.reason.trim());
    else if (action.type === "callback-investigate")
      await investigateCallback(id, actionForm.reason.trim());
    else if (action.type === "abnormal-assign")
      await assignAbnormalBet(id, {
        reviewerID: Number(actor.value?.id || 0),
        reviewer: String(actor.value?.displayName || actor.value?.username || ""),
        reason: actionForm.reason.trim(),
      });
    else if (action.type === "abnormal-review")
      await reviewAbnormalBet(
        id,
        actionForm.reviewStatus,
        actionForm.reason.trim(),
      );
    else if (action.type === "abnormal-clear")
      await clearAbnormalBet(id, actionForm.reason.trim());
    else if (action.type === "abnormal-restrict")
      await restrictAbnormalBet(id, actionForm.reason.trim());
    else if (action.type === "abnormal-note")
      await addAbnormalNote(id, actionForm.note.trim());
    actionVisible.value = false;
    ElMessage.success(t("betting.messages.actionSucceeded"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("betting.messages.actionFailed"),
    );
  } finally {
    actionLoading.value = false;
  }
}

function actionDisabled() {
  if (pendingAction.value?.type === "abnormal-note")
    return !actionForm.note.trim();
  if (pendingAction.value?.type === "abnormal-assign")
    return !actor.value?.id || !actionForm.reason.trim();
  return !actionForm.reason.trim();
}

function viewUser(row: BettingRecord) {
  if (row.userNo)
    void router.push(`/players/${encodeURIComponent(String(row.userNo))}`);
}

watch(moduleKey, () => {
  Object.keys(query).forEach((key) => delete query[key]);
  page.value = 1;
  detailVisible.value = false;
  void load();
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
  await load();
});
</script>

<template>
  <div v-if="config" class="betting-page">
    <UserCenterCard :title="t(config.searchTitleKey)">
      <el-form
        class="search-form"
        :model="query"
        label-position="left"
        label-width="132px"
        @keyup.enter="search"
      >
        <el-row :gutter="16">
          <el-col
            v-for="field in visibleFilters"
            :key="field.key"
            :xs="24"
            :sm="12"
            :lg="8"
            :xl="6"
          >
            <el-form-item :label="t(field.labelKey)">
              <el-date-picker
                v-if="field.type === 'dateRange'"
                v-model="query[field.key]"
                type="daterange"
                value-format="YYYY-MM-DD"
                :start-placeholder="t('betting.placeholders.startDate')"
                :end-placeholder="t('betting.placeholders.endDate')"
              />
              <el-select
                v-else-if="isSelectFilter(field)"
                v-model="query[field.key]"
                clearable
                filterable
                :placeholder="t('betting.placeholders.select')"
              >
                <el-option
                  v-for="item in filterOptions(field)"
                  :key="item.value"
                  :label="filterOptionLabel(item)"
                  :value="item.value"
                />
              </el-select>
              <el-input
                v-else
                v-model="query[field.key]"
                clearable
                :placeholder="t('betting.placeholders.enter')"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <div class="search-actions">
          <el-button type="primary" @click="search">{{
            t("common.search")
          }}</el-button
          ><el-button @click="reset">{{ t("common.reset") }}</el-button>
        </div>
      </el-form>
    </UserCenterCard>

    <UserCenterCard :title="t(config.tableTitleKey)">
      <el-table
        v-loading="loading"
        :data="rows"
        row-key="id"
        border
        :empty-text="t('betting.messages.empty')"
      >
        <el-table-column
          type="index"
          :index="indexMethod"
          :label="t('common.index')"
          fixed="left"
          width="76"
        />
        <el-table-column
          v-for="column in visibleColumns"
          :key="column.key"
          :prop="column.key"
          :label="t(column.labelKey)"
          :width="column.width"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <UserStatusTag
              v-if="column.status"
              :value="String(row[column.key] || '')"
              :label="optionLabel(String(row[column.key] || ''))"
              :type-override="tagTypeOverride(column.key, row[column.key])"
            />
            <span v-else-if="column.imageKey" class="game-name-with-icon">
              <span>{{ row[column.key] ?? "-" }}</span>
              <el-image
                v-if="row[column.imageKey]"
                class="betting-game-icon"
                :src="String(row[column.imageKey])"
                fit="cover"
                :preview-src-list="[String(row[column.imageKey])]"
                preview-teleported
              />
            </span>
            <el-tag
              v-else-if="column.profit"
              :type="profitTagType(row[column.key])"
              effect="light"
              round
            >
              {{ row[column.key] ?? "0.00" }}
            </el-tag>
            <span v-else-if="column.time">{{
              formatBackendTime(String(row[column.key] || ""))
            }}</span>
            <span v-else-if="column.money">{{
              row[column.key] ?? "0.00"
            }}</span>
            <span v-else>{{ row[column.key] ?? "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.action')" fixed="right" width="280">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">{{
              t("common.detail")
            }}</el-button>
            <el-button
              v-if="row.userID"
              link
              type="primary"
              @click="viewUser(row)"
              >{{ t("betting.actions.viewUser") }}</el-button
            >
            <el-button
              v-if="moduleKey === 'bet-records' && can('bet:review')"
              link
              type="warning"
              @click="
                ask(
                  row,
                  'bet-review',
                  'betting.dialogs.markReviewTitle',
                  'betting.dialogs.markReviewMessage',
                  'bet:review',
                )
              "
              >{{ t("betting.actions.markReview") }}</el-button
            >
            <el-button
              v-if="
                moduleKey === 'rollback-records' && can('bet:rollback:review')
              "
              link
              type="warning"
              @click="
                ask(
                  row,
                  'rollback-investigate',
                  'betting.dialogs.investigateTitle',
                  'betting.dialogs.investigateMessage',
                  'bet:rollback:review',
                )
              "
              >{{ t("betting.actions.investigate") }}</el-button
            >
            <el-button
              v-if="
                moduleKey === 'provider-callback-logs' &&
                can('bet:callback-log:view')
              "
              link
              type="warning"
              @click="
                ask(
                  row,
                  'callback-investigate',
                  'betting.dialogs.investigateTitle',
                  'betting.dialogs.investigateMessage',
                  'bet:callback-log:view',
                )
              "
              >{{ t("betting.actions.investigate") }}</el-button
            >
            <el-dropdown v-if="moduleKey === 'abnormal-bets'">
              <el-button link type="primary">{{
                t("betting.actions.manage")
              }}</el-button>
              <template #dropdown
                ><el-dropdown-menu>
                  <el-dropdown-item
                    v-if="can('bet:abnormal:assign')"
                    @click="
                      ask(
                        row,
                        'abnormal-assign',
                        'betting.dialogs.assignTitle',
                        'betting.dialogs.assignMessage',
                        'bet:abnormal:assign',
                      )
                    "
                    >{{ t("betting.actions.assign") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    v-if="can('bet:abnormal:review')"
                    @click="
                      ask(
                        row,
                        'abnormal-review',
                        'betting.dialogs.reviewTitle',
                        'betting.dialogs.reviewMessage',
                        'bet:abnormal:review',
                      )
                    "
                    >{{ t("betting.actions.review") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    v-if="can('bet:abnormal:clear')"
                    @click="
                      ask(
                        row,
                        'abnormal-clear',
                        'betting.dialogs.clearTitle',
                        'betting.dialogs.clearMessage',
                        'bet:abnormal:clear',
                      )
                    "
                    >{{ t("betting.actions.clear") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    v-if="can('bet:abnormal:restrict')"
                    divided
                    @click="
                      ask(
                        row,
                        'abnormal-restrict',
                        'betting.dialogs.restrictTitle',
                        'betting.dialogs.restrictMessage',
                        'bet:abnormal:restrict',
                      )
                    "
                    >{{ t("betting.actions.restrict") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    v-if="can('bet:abnormal:note')"
                    @click="
                      ask(
                        row,
                        'abnormal-note',
                        'betting.dialogs.noteTitle',
                        'betting.dialogs.noteMessage',
                        'bet:abnormal:note',
                      )
                    "
                    >{{ t("betting.actions.addNote") }}</el-dropdown-item
                  >
                </el-dropdown-menu></template
              >
            </el-dropdown>
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
      :title="t('betting.dialogs.detailTitle')"
      width="min(1100px, 94vw)"
      destroy-on-close
    >
      <div v-loading="detailLoading" class="detail-body">
        <el-tabs v-if="detailTabs.length" v-model="activeDetailTab" type="card">
          <el-tab-pane
            v-for="tab in detailTabs"
            :key="tab.key"
            :name="tab.key"
            :label="t(tab.titleKey)"
          >
            <template v-if="Array.isArray(detail[tab.key])">
              <el-table
                :data="detail[tab.key] as RecordValue[]"
                border
                max-height="480"
                :empty-text="t('betting.messages.empty')"
              >
                <el-table-column
                  v-for="key in arrayColumns(detail[tab.key])"
                  :key="key"
                  :label="detailLabel(key)"
                  min-width="145"
                  show-overflow-tooltip
                  ><template #default="{ row }">{{
                    displayValue(key, row[key])
                  }}</template></el-table-column
                >
              </el-table>
            </template>
            <el-card v-else-if="isObject(detail[tab.key])" shadow="never">
              <el-descriptions :column="3" border
                ><el-descriptions-item
                  v-for="[key, value] in businessEntries(detail[tab.key])"
                  :key="key"
                  :label="detailLabel(String(key))"
                >
                  <pre
                    v-if="isObject(value) || Array.isArray(value)"
                    class="json-value"
                    >{{ displayValue(String(key), value) }}</pre>
                  <UserStatusTag
                    v-else-if="shouldRenderTag(String(key), value)"
                    :value="String(value || '')"
                    :label="optionLabel(String(value || ''))"
                    :type-override="tagTypeOverride(String(key), value)"
                  />
                  <span v-else>{{
                    displayValue(String(key), value)
                  }}</span></el-descriptions-item
                ></el-descriptions
              >
            </el-card>
            <el-empty v-else :description="t('betting.messages.empty')" />
          </el-tab-pane>
        </el-tabs>
        <el-empty
          v-else-if="!detailLoading"
          :description="t('betting.messages.empty')"
        />
      </div>
      <template #footer
        ><el-button @click="detailVisible = false">{{
          t("common.close")
        }}</el-button></template
      >
    </el-dialog>

    <el-dialog
      v-model="actionVisible"
      :title="pendingAction ? t(pendingAction.titleKey) : ''"
      width="480px"
    >
      <p class="dialog-message">
        {{ pendingAction ? t(pendingAction.messageKey) : "" }}
      </p>
      <el-form label-position="left" label-width="132px">
        <el-alert
          v-if="pendingAction?.type === 'abnormal-assign'"
          type="info"
          :closable="false"
          :title="String(actor?.displayName || actor?.username || '-')"
        />
        <el-form-item
          v-if="pendingAction?.type === 'abnormal-review'"
          :label="t('betting.fields.reviewStatus')"
          required
          ><el-select v-model="actionForm.reviewStatus"
            ><el-option
              v-for="item in ['UNDER_REVIEW', 'CONFIRMED', 'RESTRICTED']"
              :key="item"
              :value="item"
              :label="optionLabel(item)" /></el-select
        ></el-form-item>
        <el-form-item
          v-if="pendingAction?.type === 'abnormal-note'"
          :label="t('betting.fields.internalNote')"
          required
          ><el-input
            v-model="actionForm.note"
            type="textarea"
            :rows="4"
            maxlength="1000"
            show-word-limit
        /></el-form-item>
        <el-form-item
          v-else
          :label="t('betting.fields.operatorReason')"
          required
          ><el-input
            v-model="actionForm.reason"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit
        /></el-form-item>
      </el-form>
      <template #footer
        ><el-button @click="actionVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="actionLoading"
          :disabled="actionDisabled()"
          @click="confirmAction"
          >{{ t("common.confirm") }}</el-button
        ></template
      >
    </el-dialog>
  </div>
</template>

<style scoped>
.betting-page {
  display: grid;
  gap: 16px;
}
.search-form :deep(.el-select),
.search-form :deep(.el-date-editor) {
  width: 100%;
}
.search-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 2px;
}
.pagination-row {
  display: flex;
  justify-content: flex-end;
  padding-top: 18px;
}
.game-name-with-icon {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}
.betting-game-icon {
  width: 24px;
  height: 24px;
  flex: 0 0 24px;
  border-radius: 4px;
  overflow: hidden;
  background: var(--admin-bg-soft);
}
.detail-body {
  min-height: 260px;
}
.json-value {
  margin: 0;
  max-width: 100%;
  white-space: pre-wrap;
  word-break: break-word;
  font: inherit;
}
.dialog-message {
  margin: 0 0 16px;
  color: var(--admin-text-secondary);
  line-height: 1.6;
}
:deep(.el-descriptions__label) {
  width: 155px;
}
:deep(.el-table .cell) {
  white-space: nowrap;
}
</style>
