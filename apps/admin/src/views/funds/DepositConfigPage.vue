<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox, type UploadRequestOptions } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime } from "@/utils/timezone";
import {
  createDepositConfig,
  deleteDepositConfig,
  fetchDepositConfig,
  fetchDepositConfigs,
  updateDepositConfig,
  type DepositConfigPayload,
  type DepositConfigResource,
  type DepositConfigRow,
} from "@/api/adminFunds";
import { uploadOperationAsset } from "@/api/adminOperations";
import { businessColumns, businessEntries, sanitizeBusinessValue } from "@/utils/presentation";
import {
  fetchAdminOptions,
  optionList,
  stringOptions,
  type AdminOption,
  type AdminOptionSource,
  type AdminOptions,
} from "@/api/adminOptions";

type Field = {
  key: string;
  type?: "text" | "select" | "number" | "money" | "switch" | "image";
  options?: string[];
  source?: AdminOptionSource;
  required?: boolean;
};
type Column = Field & { width?: number; time?: boolean };
type PageConfig = {
  resource: DepositConfigResource;
  permissionPrefix: string;
  filters: Field[];
  columns: Column[];
  form: Field[];
};
type ObjectValue = Record<string, unknown>;

const configs: Record<string, PageConfig> = {
  "deposit-methods": {
    resource: "deposit-methods",
    permissionPrefix: "fund:deposit-method",
    filters: [
      { key: "methodCode" }, { key: "name" },
      { key: "methodType", type: "select", options: ["deposit", "both"] },
      { key: "status", type: "select", options: ["1", "2"] },
    ],
    columns: [
      { key: "iconURL", type: "image", width: 82 }, { key: "methodCode", width: 160 },
      { key: "name", width: 170 }, { key: "methodType", width: 120 },
      { key: "minAmount", width: 130 }, { key: "maxAmount", width: 130 },
      { key: "channelCount", width: 110 }, { key: "sort", width: 90 },
      { key: "status", type: "switch", width: 92 }, { key: "updatedAt", width: 178, time: true },
    ],
    form: [
      { key: "methodCode", required: true }, { key: "name", required: true },
      { key: "methodType", type: "select", options: ["deposit", "both"], required: true },
      { key: "iconURL", type: "image" }, { key: "minAmount", type: "money", required: true },
      { key: "maxAmount", type: "money", required: true }, { key: "sort", type: "number" },
      { key: "status", type: "switch" },
    ],
  },
  "deposit-channels": {
    resource: "deposit-channels",
    permissionPrefix: "fund:deposit-channel",
    filters: [
      { key: "channelCode" }, { key: "name" }, { key: "methodCode" },
      { key: "countryCode", type: "select", source: "countries" },
      { key: "currencyCode", type: "select", source: "currencies" },
      { key: "status", type: "select", options: ["1", "2"] },
    ],
    columns: [
      { key: "channelCode", width: 160 }, { key: "name", width: 170 },
      { key: "methodCode", width: 150 }, { key: "countryCode", width: 110 },
      { key: "currencyCode", width: 110 }, { key: "feeRate", width: 110 },
      { key: "minAmount", width: 130 }, { key: "maxAmount", width: 130 },
      { key: "sort", width: 90 }, { key: "status", type: "switch", width: 92 },
      { key: "updatedAt", width: 178, time: true },
    ],
    form: [
      { key: "channelCode", required: true }, { key: "name", required: true },
      { key: "countryCode", type: "select", source: "countries", required: true },
      { key: "currencyCode", type: "select", source: "currencies", required: true },
      { key: "methodId", type: "select", required: true }, { key: "feeRate", type: "money", required: true },
      { key: "minAmount", type: "money", required: true }, { key: "maxAmount", type: "money", required: true },
      { key: "sort", type: "number" }, { key: "status", type: "switch" },
    ],
  },
  "deposit-amount-options": {
    resource: "deposit-amount-options",
    permissionPrefix: "fund:deposit-option",
    filters: [{ key: "optionCode" }, { key: "status", type: "select", options: ["1", "2"] }],
    columns: [
      { key: "optionCode", width: 170 }, { key: "amount", width: 145 },
      { key: "defaultBonusReward", width: 165 }, { key: "defaultCashReward", width: 160 },
      { key: "defaultRewardAmount", width: 165 }, { key: "sort", width: 90 },
      { key: "status", type: "switch", width: 92 }, { key: "updatedAt", width: 178, time: true },
    ],
    form: [
      { key: "optionCode", required: true }, { key: "amount", type: "money", required: true },
      { key: "defaultBonusReward", type: "money", required: true },
      { key: "defaultCashReward", type: "money", required: true },
      { key: "defaultRewardAmount", type: "money", required: true },
      { key: "sort", type: "number" }, { key: "status", type: "switch" },
    ],
  },
};

const route = useRoute();
const { t } = useI18n();
const { can, loadPermissions } = useAdminPermissions();
const moduleKey = computed(() => String(route.meta.moduleKey || "deposit-methods"));
const config = computed(() => configs[moduleKey.value]);
const query = reactive<Record<string, string>>({});
const rows = ref<DepositConfigRow[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<ObjectValue>({});
const activeTab = ref("");
const editVisible = ref(false);
const editLoading = ref(false);
const editingID = ref(0);
const form = reactive<ObjectValue>({});
const uploading = ref(false);
const methods = ref<DepositConfigRow[]>([]);
const adminOptions = ref<AdminOptions>({});

const titleKey = computed(() => `depositConfig.pages.${moduleKey.value}`);
const detailTabs = computed(() => Object.keys(detail.value));

function label(key: string) { return t(`depositConfig.fields.${key}`); }
function valueLabel(value: unknown) {
  const key = `depositConfig.values.${String(value).toLowerCase()}`;
  return t(key);
}
function fieldOptions(field: Field): AdminOption[] {
  return field.source ? optionList(adminOptions.value, field.source) : stringOptions(field.options);
}
function applyCountryCurrency(countryCode: string) {
  const country = optionList(adminOptions.value, "countries").find((item) => item.value === countryCode);
  if (country?.currencyCode) form.currencyCode = country.currencyCode;
}
function isObject(value: unknown): value is ObjectValue {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function indexMethod(index: number) { return (page.value - 1) * pageSize.value + index + 1; }
function initializeQuery() {
  Object.keys(query).forEach((key) => delete query[key]);
  config.value.filters.forEach((field) => { query[field.key] = ""; });
}
function defaultValue(field: Field) {
  if (field.type === "switch") return true;
  if (field.type === "number") return 0;
  if (field.type === "money") return "0.00";
  if (field.key === "methodType") return "deposit";
  return "";
}
function resetForm() {
  Object.keys(form).forEach((key) => delete form[key]);
  config.value.form.forEach((field) => { form[field.key] = defaultValue(field); });
  if (moduleKey.value === "deposit-channels") {
    const country = optionList(adminOptions.value, "countries")[0];
    if (country) {
      form.countryCode = country.value;
      form.currencyCode = country.currencyCode || optionList(adminOptions.value, "currencies")[0]?.value || "";
    }
  }
}
function buildQuery() {
  const result: Record<string, string | number> = { page: page.value, pageSize: pageSize.value };
  Object.entries(query).forEach(([key, value]) => { if (value.trim()) result[key] = value.trim(); });
  return result;
}
async function loadRows() {
  if (!can(`${config.value.permissionPrefix}:list`)) return;
  loading.value = true;
  try {
    const data = await fetchDepositConfigs(config.value.resource, buildQuery());
    rows.value = data.list;
    total.value = data.total;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("depositConfig.messages.loadFailed"));
  } finally { loading.value = false; }
}
async function loadMethods() {
  if (moduleKey.value !== "deposit-channels") return;
  const data = await fetchDepositConfigs("deposit-methods", { page: 1, pageSize: 100 });
  methods.value = data.list;
}
async function loadAdminOptions() {
  if (moduleKey.value !== "deposit-channels") return;
  adminOptions.value = await fetchAdminOptions().catch(() => ({}));
}
function search() { page.value = 1; void loadRows(); }
function reset() { initializeQuery(); search(); }
function overview(source: ObjectValue) {
  return isObject(source.overview) ? source.overview : {};
}
async function openDetail(row: DepositConfigRow) {
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    detail.value = await fetchDepositConfig(config.value.resource, row.id);
    activeTab.value = Object.keys(detail.value)[0] || "";
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("depositConfig.messages.loadFailed"));
  } finally { detailLoading.value = false; }
}
function assignForm(source: ObjectValue) {
  config.value.form.forEach((field) => {
    const value = source[field.key];
    form[field.key] = field.type === "switch" ? Number(value) === 1 : (value ?? defaultValue(field));
  });
}
async function openCreate() {
  editingID.value = 0;
  resetForm();
  await loadMethods();
  editVisible.value = true;
}
async function openEdit(row: DepositConfigRow) {
  editingID.value = row.id;
  resetForm();
  editVisible.value = true;
  editLoading.value = true;
  try {
    await loadMethods();
    assignForm(overview(await fetchDepositConfig(config.value.resource, row.id)));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("depositConfig.messages.loadFailed"));
  } finally { editLoading.value = false; }
}
function payload(): DepositConfigPayload {
  const result: DepositConfigPayload = {};
  config.value.form.forEach((field) => {
    const value = form[field.key] ?? defaultValue(field);
    result[field.key] = field.type === "switch" ? (value ? 1 : 2) : value;
  });
  return result;
}
function validate() {
  const missing = config.value.form.find((field) => field.required && String(form[field.key] ?? "").trim() === "");
  if (!missing) return true;
  ElMessage.warning(t("depositConfig.validation.required", { field: label(missing.key) }));
  return false;
}
async function save() {
  if (!validate()) return;
  editLoading.value = true;
  try {
    if (editingID.value) await updateDepositConfig(config.value.resource, editingID.value, payload());
    else await createDepositConfig(config.value.resource, payload());
    ElMessage.success(t("depositConfig.messages.saved"));
    editVisible.value = false;
    await loadRows();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("depositConfig.messages.saveFailed"));
  } finally { editLoading.value = false; }
}
async function toggle(row: DepositConfigRow, enabled: boolean) {
  try {
    const data = await fetchDepositConfig(config.value.resource, row.id);
    const source = overview(data);
    resetForm();
    assignForm(source);
    form.status = enabled;
    await updateDepositConfig(config.value.resource, row.id, payload());
    await loadRows();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("depositConfig.messages.saveFailed"));
  }
}
async function remove(row: DepositConfigRow) {
  try {
    await ElMessageBox.confirm(t("depositConfig.dialogs.deleteMessage"), t("depositConfig.dialogs.deleteTitle"), { type: "warning" });
    await deleteDepositConfig(config.value.resource, row.id);
    ElMessage.success(t("depositConfig.messages.deleted"));
    await loadRows();
  } catch (error) {
    if (error === "cancel" || error === "close") return;
    ElMessage.error(error instanceof Error ? error.message : t("depositConfig.messages.deleteFailed"));
  }
}
async function uploadIcon(options: UploadRequestOptions) {
  uploading.value = true;
  try {
    const asset = await uploadOperationAsset(options.file, "image");
    form.iconURL = asset.url;
    options.onSuccess(asset);
  } catch (error) {
    const uploadError = error instanceof Error ? error : new Error(t("depositConfig.messages.uploadFailed"));
    options.onError(uploadError as Parameters<UploadRequestOptions["onError"]>[0]);
  } finally { uploading.value = false; }
}
function objectEntries(value: unknown) {
  return businessEntries(value);
}
function tableColumns(value: unknown) {
	return businessColumns(value, 10);
}
function displayValue(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (isResourceKey(key)) return value ? t("depositConfig.messages.assetConfigured") : "-";
  if (/(At|Time)$/.test(key) && typeof value === "string") return formatBackendTime(value);
  if (typeof value === "boolean") return value ? t("common.yes") : t("common.no");
  if (typeof value === "object") return JSON.stringify(sanitizeBusinessValue(value));
  return String(value);
}
function isResourceKey(key: string) {
  return /(image|icon|logo|banner|asset|thumbnail|background).*url$/i.test(key);
}

watch(moduleKey, async () => { initializeQuery(); page.value = 1; await loadPermissions(); await loadAdminOptions(); await loadRows(); });
onMounted(async () => { initializeQuery(); await loadPermissions(); await loadAdminOptions(); await loadRows(); });
</script>

<template>
  <div v-if="config" class="deposit-config-page">
    <UserCenterCard :title="t(`${titleKey}.search`)">
      <el-form inline label-position="left" class="deposit-config-search" @submit.prevent="search">
        <el-form-item v-for="field in config.filters" :key="field.key" :label="label(field.key)">
          <el-select v-if="field.type === 'select'" v-model="query[field.key]" clearable>
            <el-option v-for="option in fieldOptions(field)" :key="option.value" :label="field.source ? option.label : valueLabel(option.value)" :value="option.value" />
          </el-select>
          <el-input v-else v-model="query[field.key]" clearable @keyup.enter="search" />
        </el-form-item>
        <div class="deposit-config-search__actions">
          <el-button type="primary" @click="search">{{ t("common.search") }}</el-button>
          <el-button @click="reset">{{ t("common.reset") }}</el-button>
        </div>
      </el-form>
    </UserCenterCard>

    <UserCenterCard :title="t(`${titleKey}.title`)">
      <template #extra>
        <el-button v-if="can(`${config.permissionPrefix}:create`)" type="primary" @click="openCreate">{{ t("common.create") }}</el-button>
      </template>
      <el-table v-loading="loading" :data="rows" row-key="id" border>
        <el-table-column fixed="left" type="index" :index="indexMethod" :label="t('common.index')" width="72" align="center" />
        <el-table-column v-for="column in config.columns" :key="column.key" :prop="column.key" :label="label(column.key)" :width="column.width" show-overflow-tooltip>
          <template #default="scope">
            <el-image v-if="column.type === 'image' && scope.row[column.key]" class="config-icon" :src="String(scope.row[column.key])" fit="contain" />
            <el-switch v-else-if="column.type === 'switch'" :model-value="Number(scope.row[column.key]) === 1" :disabled="!can(`${config.permissionPrefix}:update`)" @change="(value: boolean) => toggle(scope.row, value)" />
            <UserStatusTag v-else-if="column.type === 'select' || ['methodType'].includes(column.key)" :value="String(scope.row[column.key])" :label="valueLabel(scope.row[column.key])" />
            <span v-else-if="column.time">{{ formatBackendTime(String(scope.row[column.key] || '')) }}</span>
            <span v-else>{{ scope.row[column.key] ?? "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column fixed="right" :label="t('common.action')" width="210" align="center">
          <template #default="scope">
            <div class="table-actions">
              <el-button link type="primary" @click="openDetail(scope.row)">{{ t("common.detail") }}</el-button>
              <el-button v-if="can(`${config.permissionPrefix}:update`)" link type="primary" @click="openEdit(scope.row)">{{ t("common.edit") }}</el-button>
              <el-button v-if="can(`${config.permissionPrefix}:delete`)" link type="danger" @click="remove(scope.row)">{{ t("common.delete") }}</el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty><el-empty :description="t('depositConfig.messages.empty')" /></template>
      </el-table>
      <div class="deposit-config-pagination">
        <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :page-sizes="[20, 50, 100]" layout="prev, pager, next, sizes, total" :total="total" @current-change="loadRows" @size-change="search" />
      </div>
    </UserCenterCard>

    <el-dialog v-model="detailVisible" :title="t('depositConfig.dialogs.detail')" width="min(960px, 92vw)" append-to-body destroy-on-close>
      <div v-loading="detailLoading">
        <el-tabs v-model="activeTab" type="card">
          <el-tab-pane v-for="tab in detailTabs" :key="tab" :name="tab" :label="label(tab)">
            <el-table v-if="Array.isArray(detail[tab])" :data="detail[tab]" border max-height="420">
              <el-table-column v-for="key in tableColumns(detail[tab])" :key="key" :label="label(key)" min-width="140" show-overflow-tooltip>
                <template #default="scope">{{ displayValue(key, scope.row[key]) }}</template>
              </el-table-column>
            </el-table>
            <el-descriptions v-else-if="isObject(detail[tab])" :column="2" border>
              <el-descriptions-item v-for="([key, value]) in objectEntries(detail[tab])" :key="key" :label="label(key)">
                <el-image
                  v-if="isResourceKey(key) && value"
                  class="detail-resource-image"
                  :src="String(value)"
                  fit="contain"
                  :preview-src-list="[String(value)]"
                  preview-teleported
                />
                <span v-else>{{ displayValue(key, value) }}</span>
              </el-descriptions-item>
            </el-descriptions>
            <el-empty v-else :description="t('depositConfig.messages.empty')" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>

    <el-dialog v-model="editVisible" :title="editingID ? t('depositConfig.dialogs.edit') : t('depositConfig.dialogs.create')" width="min(760px, 94vw)" append-to-body destroy-on-close>
      <el-form v-loading="editLoading" label-position="left" label-width="150px" class="deposit-config-form">
        <el-form-item v-for="field in config.form" :key="field.key" :label="label(field.key)" :required="field.required">
          <el-switch v-if="field.type === 'switch'" v-model="form[field.key]" />
          <el-input-number v-else-if="field.type === 'number'" v-model="form[field.key]" :min="0" controls-position="right" />
          <el-select v-else-if="field.key === 'methodId'" v-model="form[field.key]" filterable>
            <el-option v-for="method in methods" :key="method.id" :label="`${method.methodCode} · ${method.name}`" :value="method.id" />
          </el-select>
          <el-select v-else-if="field.type === 'select'" v-model="form[field.key]" @change="field.key === 'countryCode' && applyCountryCurrency(String($event))">
            <el-option v-for="option in fieldOptions(field)" :key="option.value" :label="field.source ? option.label : valueLabel(option.value)" :value="option.value" />
          </el-select>
          <div v-else-if="field.type === 'image'" class="config-upload">
            <el-upload :show-file-list="false" accept="image/*,.avif" :http-request="uploadIcon"><el-button :loading="uploading">{{ form[field.key] ? t("depositConfig.actions.replaceIcon") : t("depositConfig.actions.uploadIcon") }}</el-button></el-upload>
            <el-button v-if="form[field.key]" type="danger" link @click="form[field.key] = ''">{{ t("depositConfig.actions.removeIcon") }}</el-button>
            <el-image v-if="form[field.key]" class="config-upload__preview" :src="String(form[field.key])" fit="contain" />
          </div>
          <el-input v-else v-model="form[field.key]" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" :loading="editLoading" @click="save">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.deposit-config-page { display: grid; gap: 16px; }
.deposit-config-search { display: flex; flex-wrap: wrap; gap: 12px 18px; align-items: center; }
.deposit-config-search :deep(.el-form-item) { width: 300px; margin: 0; }
.deposit-config-search :deep(.el-form-item__label) { width: 112px; justify-content: flex-start; }
.deposit-config-search :deep(.el-input), .deposit-config-search :deep(.el-select) { width: 188px; }
.deposit-config-search__actions { display: flex; gap: 8px; margin-left: auto; }
.config-icon { width: 48px; height: 36px; }
.table-actions { display: flex; align-items: center; justify-content: center; gap: 2px; white-space: nowrap; }
.deposit-config-pagination { display: flex; justify-content: flex-end; padding-top: 16px; }
.deposit-config-form :deep(.el-select), .deposit-config-form :deep(.el-input-number) { width: 100%; }
.config-upload { display: grid; grid-template-columns: auto auto minmax(0, 1fr); align-items: center; gap: 8px; width: 100%; }
.config-upload__preview { grid-column: 1 / -1; width: 120px; height: 72px; border: 1px solid var(--admin-border); border-radius: 4px; }
.detail-resource-image { width: 120px; height: 72px; border: 1px solid var(--admin-border); border-radius: 4px; background: #f7f8fa; }
@media (max-width: 760px) {
  .deposit-config-search :deep(.el-form-item) { width: 100%; }
  .deposit-config-search :deep(.el-input), .deposit-config-search :deep(.el-select) { width: calc(100% - 112px); }
}
</style>
