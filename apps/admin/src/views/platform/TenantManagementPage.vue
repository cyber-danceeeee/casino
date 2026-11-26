<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import { Delete, DocumentChecked, Edit, Key, Link, Plus, View } from "@element-plus/icons-vue";
import { fetchAdminOptions, optionList, type AdminOptions } from "@/api/adminOptions";
import { systemAPI, type TenantPackage, type TenantSummary } from "@/api/adminSystem";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime } from "@/utils/timezone";

const { t } = useI18n();
const router = useRouter();
const { can } = useAdminPermissions();
const loading = ref(false);
const rows = ref<TenantSummary[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const options = ref<AdminOptions>({});
const tenantPackages = ref<TenantPackage[]>([]);
const query = reactive({ tenantName: "", tenantCode: "", countryCode: "", defaultCurrency: "", lifecycleStatus: "" });
const editVisible = ref(false);
const editing = ref<TenantSummary>();
const editForm = reactive<Record<string, unknown>>({});

async function load() {
  loading.value = true;
  try {
    const result = await systemAPI.tenants.list({ ...query, page: page.value, pageSize: pageSize.value });
    rows.value = result.list || [];
    total.value = result.total || 0;
  } finally {
    loading.value = false;
  }
}

function reset() {
  Object.assign(query, { tenantName: "", tenantCode: "", countryCode: "", defaultCurrency: "", lifecycleStatus: "" });
  page.value = 1;
  void load();
}

function openEdit(row: TenantSummary) {
  editing.value = row;
  Object.assign(editForm, {
    tenantName: row.tenantName, companyName: row.companyName, contactName: row.contactName,
    contactEmail: row.contactEmail, contactPhone: row.contactPhone, planCode: row.planCode,
    packageCode: row.packageCode, countryCode: row.countryCode, defaultCurrency: row.defaultCurrency,
    defaultLanguage: row.defaultLanguage, timezone: row.timezone, expiresAt: row.expiresAt || null,
    primaryDomain: row.primaryDomain, h5Domain: row.h5Domain, adminDomain: row.adminDomain,
    enabled: row.enabled, status: row.status, maintenanceEnabled: row.maintenanceEnabled,
    maintenanceMessage: row.maintenanceMessage, remark: row.remark, reason: "",
  });
  editVisible.value = true;
}

function packageLabel(item: TenantPackage) {
  return `${item.code} · ${item.nameZh || item.nameEn}`;
}

async function saveEdit() {
  if (!editing.value || !String(editForm.reason || "").trim()) {
    ElMessage.warning(t("tenantCenter.messages.reasonRequired"));
    return;
  }
  await systemAPI.tenants.update(editing.value.id, { ...editForm });
  ElMessage.success(t("tenantCenter.messages.saved"));
  editVisible.value = false;
  await load();
}

async function removeTenant(row: TenantSummary) {
  try {
    const { value } = await ElMessageBox.prompt(t("tenantCenter.confirm.deleteTenant"), t("common.delete"), {
      inputPattern: /\S+/, inputErrorMessage: t("tenantCenter.messages.reasonRequired"),
      confirmButtonText: t("common.delete"),
      confirmButtonClass: "el-button--danger",
    });
    await systemAPI.tenants.remove(row.id, value);
    ElMessage.success(t("tenantCenter.messages.deleted"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(error instanceof Error ? error.message : t("tenantCenter.messages.deleteFailed"));
    }
  }
}

async function changeStatus(row: TenantSummary) {
  const next = row.enabled;
  try {
    const { value } = await ElMessageBox.prompt(
      next ? t("tenantCenter.confirm.enable") : t("tenantCenter.confirm.disable"),
      t("tenantCenter.fields.reason"),
      { inputPattern: /\S+/, inputErrorMessage: t("tenantCenter.messages.reasonRequired") },
    );
    await systemAPI.tenants.status(row.id, next, value);
    ElMessage.success(t("tenantCenter.messages.statusUpdated"));
  } catch {
    row.enabled = !next;
  }
}

async function resetPassword(row: TenantSummary) {
  if (!row.ownerAdminID) return;
  try {
    const { value } = await ElMessageBox.prompt(t("tenantCenter.confirm.resetPassword"), t("tenantCenter.fields.reason"), {
      inputPattern: /\S+/, inputErrorMessage: t("tenantCenter.messages.reasonRequired"),
    });
    const result = await systemAPI.admins.resetPassword(row.ownerAdminID, value);
    await ElMessageBox.alert(result.temporaryPassword, t("tenantCenter.fields.temporaryPassword"), {
      confirmButtonText: t("common.confirm"), customClass: "tenant-secret-dialog",
    });
  } catch {
    // Cancellation leaves the administrator unchanged.
  }
}

function openAdmin(row: TenantSummary) {
  if (row.adminDomain) window.open(`https://${row.adminDomain}/tenant/login`, "_blank", "noopener,noreferrer");
}

onMounted(async () => {
  const [loadedOptions, packages] = await Promise.all([
    fetchAdminOptions().catch(() => ({})),
    systemAPI.tenantPackages.list().catch(() => ({ list: [] as TenantPackage[] })),
  ]);
  options.value = loadedOptions;
  tenantPackages.value = packages.list || [];
  await load();
});
</script>

<template>
  <main class="tenant-list-page">
    <header class="page-heading">
      <div><p>{{ t("portal.platform.tenants") }}</p><h1>{{ t("tenantCenter.list.title") }}</h1><span>{{ t("tenantCenter.list.description") }}</span></div>
      <el-button type="primary" size="large" :icon="Plus" @click="router.push('/platform/tenants/create')">{{ t("portal.createBrand") }}</el-button>
    </header>

    <section class="filter-band">
      <el-form inline>
        <el-form-item :label="t('tenantCenter.fields.tenantName')"><el-input v-model="query.tenantName" clearable /></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.tenantCode')"><el-input v-model="query.tenantCode" clearable /></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.country')"><el-select v-model="query.countryCode" clearable filterable><el-option v-for="item in optionList(options,'countries')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.currency')"><el-select v-model="query.defaultCurrency" clearable filterable><el-option v-for="item in optionList(options,'currencies')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
        <el-form-item :label="t('common.status')"><el-select v-model="query.lifecycleStatus" clearable><el-option :label="t('tenantCenter.status.ACTIVE')" value="ACTIVE"/><el-option :label="t('tenantCenter.status.DISABLED')" value="DISABLED"/></el-select></el-form-item>
        <el-form-item><el-button type="primary" @click="page=1;load()">{{ t("common.search") }}</el-button><el-button @click="reset">{{ t("common.reset") }}</el-button></el-form-item>
      </el-form>
    </section>

    <section class="table-band">
      <el-table :data="rows" v-loading="loading" row-key="id" height="calc(100vh - 300px)">
        <el-table-column prop="tenantName" :label="t('tenantCenter.fields.tenantName')" min-width="160" fixed="left"><template #default="{row}"><button class="tenant-link" @click="router.push(`/system/tenants/${row.id}`)">{{ row.tenantName }}</button><small>{{ row.tenantCode }}</small></template></el-table-column>
        <el-table-column prop="countryCode" :label="t('tenantCenter.fields.country')" width="92" />
        <el-table-column prop="defaultCurrency" :label="t('tenantCenter.fields.currency')" width="92" />
        <el-table-column :label="t('common.status')" width="105"><template #default="{row}"><el-switch v-model="row.enabled" :disabled="!can('system:tenant:status')" @change="changeStatus(row)" /></template></el-table-column>
        <el-table-column prop="packageCode" :label="t('tenantCenter.fields.plan')" width="135"><template #default="{row}"><el-tag effect="plain">{{ row.packageCode || row.planCode }}</el-tag></template></el-table-column>
        <el-table-column prop="primaryDomain" :label="t('tenantCenter.fields.domain')" min-width="190" show-overflow-tooltip><template #default="{row}"><span>{{ row.primaryDomain || '-' }}</span><el-tag class="domain-state" size="small" :type="row.domainStatus==='ACTIVE'?'success':'warning'">{{ t(`tenantCenter.domainStatus.${row.domainStatus || 'PENDING'}`) }}</el-tag></template></el-table-column>
        <el-table-column prop="ownerUsername" :label="t('tenantCenter.fields.admin')" min-width="130" />
        <el-table-column prop="gameCount" :label="t('tenantCenter.fields.games')" width="90" align="right" />
        <el-table-column prop="playerCount" :label="t('tenantCenter.fields.players')" width="100" align="right" />
        <el-table-column prop="createdAt" :label="t('common.createdAt')" width="175"><template #default="{row}">{{ formatBackendTime(row.createdAt) }}</template></el-table-column>
        <el-table-column prop="lastLoginTime" :label="t('tenantCenter.fields.lastLogin')" width="175"><template #default="{row}">{{ row.lastLoginTime ? formatBackendTime(row.lastLoginTime) : '-' }}</template></el-table-column>
        <el-table-column fixed="right" :label="t('common.actions')" width="270">
          <template #default="{row}"><div class="icon-actions">
            <el-tooltip :content="t('common.view')"><el-button circle :icon="View" @click="router.push(`/system/tenants/${row.id}`)" /></el-tooltip>
            <el-tooltip :content="t('common.edit')"><el-button v-if="can('system:tenant:update')" circle :icon="Edit" @click="openEdit(row)" /></el-tooltip>
            <el-tooltip :content="t('tenantCenter.actions.resetPassword')"><el-button v-if="can('system:admin:reset-password')" circle :icon="Key" :disabled="!row.ownerAdminID" @click="resetPassword(row)" /></el-tooltip>
            <el-tooltip :content="t('tenantCenter.actions.openAdmin')"><el-button circle :icon="Link" :disabled="!row.adminDomain" @click="openAdmin(row)" /></el-tooltip>
            <el-tooltip :content="t('common.audit')"><el-button circle :icon="DocumentChecked" @click="router.push({path:'/platform/tenants/audit',query:{tenantID:row.id}})" /></el-tooltip>
            <el-tooltip :content="t('common.delete')"><el-button v-if="can('system:tenant:delete')" circle type="danger" plain :icon="Delete" @click="removeTenant(row)" /></el-tooltip>
          </div></template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" background layout="total, sizes, prev, pager, next" :total="total" @change="load" />
    </section>

    <el-dialog v-model="editVisible" :title="t('tenantCenter.edit.title')" width="760px">
      <el-form :model="editForm" label-position="top" class="edit-grid">
        <el-form-item :label="t('tenantCenter.fields.tenantName')" required><el-input v-model="editForm.tenantName" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.companyName')"><el-input v-model="editForm.companyName" /></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.plan')"><el-select v-model="editForm.packageCode" filterable @change="editForm.planCode=String($event)"><el-option v-for="item in tenantPackages" :key="item.id" :label="packageLabel(item)" :value="item.code" :disabled="item.status==='DISABLED' && item.code!==editForm.packageCode" /></el-select></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.primaryDomain')"><el-input v-model="editForm.primaryDomain" placeholder="example.com" /></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.h5Domain')"><el-input v-model="editForm.h5Domain" placeholder="casino.example.com" /></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.adminDomain')"><el-input v-model="editForm.adminDomain" placeholder="admin.example.com" /></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.country')"><el-select v-model="editForm.countryCode" filterable><el-option v-for="item in optionList(options,'countries')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.currency')"><el-select v-model="editForm.defaultCurrency" filterable><el-option v-for="item in optionList(options,'currencies')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
        <el-form-item :label="t('createBrand.fields.language')"><el-select v-model="editForm.defaultLanguage" filterable><el-option v-for="item in optionList(options,'languages')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
        <el-form-item :label="t('createBrand.fields.timezone')"><el-select v-model="editForm.timezone" filterable><el-option v-for="item in optionList(options,'timezones')" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.contactEmail')"><el-input v-model="editForm.contactEmail" /></el-form-item>
        <el-form-item class="wide" :label="t('tenantCenter.fields.remark')"><el-input v-model="editForm.remark" type="textarea" /></el-form-item>
        <el-form-item class="wide" :label="t('tenantCenter.fields.reason')" required><el-input v-model="editForm.reason" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="editVisible=false">{{ t("common.cancel") }}</el-button><el-button type="primary" @click="saveEdit">{{ t("common.save") }}</el-button></template>
    </el-dialog>
  </main>
</template>

<style scoped>
.tenant-list-page{display:flex;flex-direction:column;gap:16px;min-width:0}.page-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:18px}.page-heading p{margin:0 0 4px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.page-heading h1{margin:0;font-size:24px;letter-spacing:0}.page-heading span{display:block;margin-top:7px;color:#909399;font-size:13px}.filter-band,.table-band{padding:18px 20px;background:#fff;border:1px solid #e4e7ed;border-radius:4px}.filter-band :deep(.el-form-item){margin-bottom:0}.filter-band :deep(.el-input),.filter-band :deep(.el-select){width:168px}.table-band{min-width:0}.table-band .el-pagination{justify-content:flex-end;margin-top:16px}.tenant-link{display:block;padding:0;border:0;background:transparent;color:#337ecc;font-weight:700;cursor:pointer}.tenant-link+small{display:block;margin-top:3px;color:#909399}.domain-state{margin-left:8px}.icon-actions{display:flex;gap:7px}.icon-actions :deep(.el-button+.el-button){margin-left:0}.edit-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0 18px}.edit-grid .wide{grid-column:1/-1}.edit-grid .el-select{width:100%}@media(max-width:850px){.page-heading{flex-direction:column}.edit-grid{grid-template-columns:1fr}.edit-grid .wide{grid-column:auto}}
</style>
