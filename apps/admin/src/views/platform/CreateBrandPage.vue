<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import type { UploadRequestOptions } from "element-plus";
import { CopyDocument, Link, Plus, RefreshRight, SwitchButton, UploadFilled } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import { systemAPI, type TenantPackage, type TenantProvisionResult } from "@/api/adminSystem";
import { fetchAdminOptions, optionList, type AdminOptions } from "@/api/adminOptions";
import { ApiError } from "@/api/http";
import { generateTemporaryPassword } from "@/utils/temporaryPassword";

const { t, locale } = useI18n();
const router = useRouter();
const loading = ref(false);
const result = ref<TenantProvisionResult>();
const options = ref<AdminOptions>({});
const packages = ref<TenantPackage[]>([]);
const apkUploading = ref(false);
const form = reactive({
  tenantCode: "",
  tenantName: "",
  companyName: "",
  siteCode: "",
  siteName: "",
  primaryDomain: "",
  h5Domain: "",
  adminDomain: "",
  countryCode: "MX",
  currencyCode: "MXN",
  language: "es-MX",
  timezone: "America/Mexico_City",
  planCode: "PROFESSIONAL",
  packageCode: "PROFESSIONAL",
  providerCodes: [] as string[],
  androidAPKObjectKey: "",
  androidAPKName: "",
  androidAPKSize: 0,
  ownerUsername: "",
  ownerPassword: generateTemporaryPassword(),
  ownerEmail: "",
  reason: "Create and deliver tenant brand",
});

const countryOptions = computed(() => optionList(options.value, "countries"));

function countryLabel(item: (typeof countryOptions.value)[number]) {
  return locale.value.startsWith("zh")
    ? item.labelZh || item.labelEn || item.label
    : item.labelEn || item.labelZh || item.label;
}

function applyCountryDefaults(countryCode: string) {
  const country = countryOptions.value.find((item) => item.value === countryCode);
  if (!country) return;
  form.currencyCode = country.currencyCode || "";
  form.language = country.defaultLanguage || "";
  form.timezone = country.timezone || "";
}

const deliveryRows = computed(() => {
  if (!result.value) return [];
  return [
    [t("createBrand.fields.brandName"), result.value.brandName],
    // ["Tenant ID", result.value.tenantID],
    // ["Site ID", result.value.siteID],
    [t("createBrand.fields.adminUsername"), result.value.adminUsername],
    [t("createBrand.fields.temporaryPassword"), result.value.temporaryPassword],
    [t("createBrand.fields.email"), result.value.deliveryEmail],
    [t("createBrand.fields.emailNotificationStatus"), t(`createBrand.emailStatus.${result.value.emailNotificationStatus}`)],
    [t("createBrand.fields.adminUrl"), result.value.adminUrl],
    [t("createBrand.fields.h5Url"), result.value.h5Url],
    [t("createBrand.fields.domainStatus"), result.value.domainStatus],
    [t("createBrand.fields.initialVersion"), result.value.initialConfigVersion],
  ];
});

function normalizeCode() {
  const normalize = (value: string) => value.trim().toLowerCase().replace(/[^a-z0-9-]/g, "-").replace(/-+/g, "-").replace(/^-|-$/g, "");
  form.tenantCode = normalize(form.tenantCode);
  if (!form.siteCode) form.siteCode = `${form.tenantCode}-main`;
  form.siteCode = normalize(form.siteCode);
  if (!form.siteName) form.siteName = form.tenantName;
}

function regeneratePassword() {
  form.ownerPassword = generateTemporaryPassword();
}

async function uploadAPK(options: UploadRequestOptions) {
  apkUploading.value = true;
  try {
    const asset = await systemAPI.sites.uploadAPK(options.file);
    form.androidAPKObjectKey = asset.key;
    form.androidAPKName = asset.name;
    form.androidAPKSize = asset.size;
    options.onSuccess(asset);
    ElMessage.success(t("createBrand.messages.apkUploaded"));
  } catch (error) {
    options.onError(error as Parameters<UploadRequestOptions["onError"]>[0]);
    ElMessage.error(error instanceof Error ? error.message : t("createBrand.messages.apkFailed"));
  } finally { apkUploading.value = false; }
}

async function submit() {
  normalizeCode();
  if (!form.tenantCode || !form.tenantName || !form.primaryDomain || !form.ownerUsername || !form.ownerPassword || !form.ownerEmail) {
    ElMessage.warning(t("createBrand.messages.required"));
    return;
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.ownerEmail)) {
    ElMessage.warning(t("createBrand.messages.invalidEmail"));
    return;
  }
  const codePattern = /^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])$/;
  if (!codePattern.test(form.tenantCode) || !codePattern.test(form.siteCode)) {
    ElMessage.warning(t("createBrand.messages.invalidCode"));
    return;
  }
  if (!form.countryCode || !form.currencyCode || !form.language || !form.timezone) {
    ElMessage.warning(t("createBrand.messages.countryRequired"));
    return;
  }
  loading.value = true;
  try {
    result.value = await systemAPI.tenants.provision({ ...form });
    ElMessage.success(t("createBrand.messages.created"));
  } catch (error) {
    const errorKey = error instanceof ApiError ? {
      INVALID_TENANT_CODE: "invalidCode",
      INVALID_SITE_CODE: "invalidCode",
      COUNTRY_NOT_FOUND: "countryUnavailable",
      COUNTRY_CONFIG_INCOMPLETE: "countryIncomplete",
      DUPLICATE_VALUE: "duplicate",
      WEAK_PASSWORD: "weakPassword",
      OWNER_EMAIL_REQUIRED: "required",
      INVALID_OWNER_EMAIL: "invalidEmail",
    }[error.errorCode] : undefined;
    ElMessage.error(errorKey ? t(`createBrand.messages.${errorKey}`) : error instanceof Error ? error.message : t("createBrand.messages.failed"));
  } finally {
    loading.value = false;
  }
}

function createAnother() {
  result.value = undefined;
  regeneratePassword();
}

async function copyDelivery() {
  await navigator.clipboard.writeText(deliveryRows.value.map(([label, value]) => `${label}: ${value || "-"}`).join("\n"));
  ElMessage.success(t("createBrand.messages.copied"));
}

function openURL(value?: string) {
  if (value) window.open(value, "_blank", "noopener,noreferrer");
}

onMounted(async () => {
  const [adminOptions, packageResult] = await Promise.all([
    fetchAdminOptions().catch(() => ({})),
    systemAPI.tenantPackages.list().catch(() => ({ list: [] as TenantPackage[] })),
  ]);
  options.value = adminOptions;
  packages.value = packageResult.list.filter((item) => item.status !== "DISABLED");
  if (packages.value.length && !packages.value.some((item) => item.code === form.packageCode)) {
    form.packageCode = packages.value[0].code;
    form.planCode = form.packageCode;
  }
  const selected = countryOptions.value.find((item) => item.value === form.countryCode) || countryOptions.value[0];
  if (selected) {
    form.countryCode = selected.value;
    applyCountryDefaults(selected.value);
  }
});
</script>

<template>
  <main class="create-brand-page">
    <header class="page-heading">
      <div><p>{{ t("portal.platform.tenants") }}</p><h1>{{ t("portal.createBrand") }}</h1><span>{{ t("createBrand.description") }}</span></div>
      <el-tag type="danger" effect="plain">{{ t("portal.platform.badge") }}</el-tag>
    </header>

    <section v-if="!result" class="provision-panel" v-loading="loading">
      <div class="section-heading"><el-icon><Plus /></el-icon><div><h2>{{ t("createBrand.brandSection") }}</h2><p>{{ t("createBrand.brandSectionDescription") }}</p></div></div>
      <el-form :model="form" label-position="top" class="provision-grid">
        <el-form-item :label="t('createBrand.fields.tenantCode')" required><el-input v-model="form.tenantCode" @blur="normalizeCode" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.brandName')" required><el-input v-model="form.tenantName" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.companyName')"><el-input v-model="form.companyName" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.siteCode')"><el-input v-model="form.siteCode" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.siteName')"><el-input v-model="form.siteName" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.primaryDomain')" required><el-input v-model="form.primaryDomain" placeholder="brand.example.com" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.h5Domain')"><el-input v-model="form.h5Domain" placeholder="play.example.com" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.adminDomain')"><el-input v-model="form.adminDomain" placeholder="admin.example.com" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.country')" required><el-select v-model="form.countryCode" filterable @change="applyCountryDefaults"><el-option v-for="item in countryOptions" :key="item.value" :label="countryLabel(item)" :value="item.value" /></el-select></el-form-item>
        <el-form-item :label="t('createBrand.fields.currency')"><el-select v-model="form.currencyCode" disabled><el-option v-if="form.currencyCode" :label="form.currencyCode" :value="form.currencyCode" /></el-select></el-form-item>
        <el-form-item :label="t('createBrand.fields.language')"><el-select v-model="form.language" disabled><el-option v-if="form.language" :label="form.language" :value="form.language" /></el-select></el-form-item>
        <el-form-item :label="t('createBrand.fields.timezone')"><el-select v-model="form.timezone" disabled><el-option v-if="form.timezone" :label="form.timezone" :value="form.timezone" /></el-select></el-form-item>
        <el-form-item :label="t('tenantCenter.fields.plan')"><el-select v-model="form.packageCode" @change="form.planCode=String($event)"><el-option v-for="item in packages" :key="item.id" :label="locale.startsWith('zh') ? item.nameZh : item.nameEn" :value="item.code"/></el-select></el-form-item>
        <el-form-item :label="t('createBrand.fields.defaultProvider')"><el-select v-model="form.providerCodes" multiple collapse-tags clearable><el-option v-for="item in optionList(options,'providers')" :key="item.value" :label="item.label" :value="item.value"/></el-select></el-form-item>
        <el-form-item :label="t('createBrand.fields.androidAPK')">
          <div class="apk-field"><el-upload :show-file-list="false" accept=".apk,application/vnd.android.package-archive" :http-request="uploadAPK"><el-button :icon="UploadFilled" :loading="apkUploading">{{ form.androidAPKName || t('createBrand.actions.uploadAPK') }}</el-button></el-upload><el-button v-if="form.androidAPKObjectKey" link type="danger" @click="form.androidAPKObjectKey='';form.androidAPKName='';form.androidAPKSize=0">{{ t('common.delete') }}</el-button></div>
        </el-form-item>
      </el-form>

      <div class="section-heading owner-heading"><el-icon><SwitchButton /></el-icon><div><h2>{{ t("createBrand.ownerSection") }}</h2><p>{{ t("createBrand.ownerSectionDescription") }}</p></div></div>
      <el-form :model="form" label-position="top" class="provision-grid">
        <el-form-item :label="t('createBrand.fields.adminUsername')" required><el-input v-model="form.ownerUsername" autocomplete="off" /></el-form-item>
        <el-form-item :label="t('createBrand.fields.temporaryPassword')" required>
          <el-input v-model="form.ownerPassword" disabled>
            <template #append><el-tooltip :content="t('createBrand.actions.regeneratePassword')"><el-button :icon="RefreshRight" :aria-label="t('createBrand.actions.regeneratePassword')" @click="regeneratePassword" /></el-tooltip></template>
          </el-input>
        </el-form-item>
        <el-form-item :label="t('createBrand.fields.email')" required><el-input v-model="form.ownerEmail" type="email" autocomplete="email" /></el-form-item>
      </el-form>
      <div class="form-actions"><el-button type="primary" size="large" :loading="loading" @click="submit">{{ t("createBrand.actions.createAndDeliver") }}</el-button></div>
    </section>

    <section v-else class="delivery-panel">
      <div class="delivery-heading"><div><p>{{ t("createBrand.messages.created") }}</p><h2>{{ result.brandName }}</h2></div><el-tag type="warning" effect="plain">{{ result.domainStatus }}</el-tag></div>
      <el-alert type="warning" show-icon :closable="false" :title="t('createBrand.oneTimePasswordNotice')" />
      <dl class="delivery-grid"><template v-for="[label, value] in deliveryRows" :key="String(label)"><dt>{{ label }}</dt><dd :class="{ secret: label === t('createBrand.fields.temporaryPassword') }">{{ value || '-' }}</dd></template></dl>
      <div class="delivery-actions">
        <el-button :icon="CopyDocument" @click="copyDelivery">{{ t("createBrand.actions.copyDelivery") }}</el-button>
        <el-button :icon="Link" @click="openURL(result.adminUrl)">{{ t("createBrand.actions.openAdmin") }}</el-button>
        <el-button :icon="Link" @click="openURL(result.h5Url)">{{ t("createBrand.actions.openH5") }}</el-button>
        <el-button :icon="Link" @click="router.push(`/system/tenants/${result.tenantID}`)">{{ t("createBrand.actions.viewTenant") }}</el-button>
        <el-button type="primary" @click="createAnother">{{ t("createBrand.actions.createAnother") }}</el-button>
      </div>
    </section>
  </main>
</template>

<style scoped>
.create-brand-page{display:flex;flex-direction:column;gap:20px}.page-heading,.delivery-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:20px}.page-heading p,.delivery-heading p{margin:0 0 6px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.page-heading h1,.delivery-heading h2{margin:0;letter-spacing:0}.page-heading h1{font-size:25px}.page-heading span{display:block;margin-top:8px;color:#909399;font-size:13px}.provision-panel,.delivery-panel{padding:24px;background:#fff;border:1px solid #e4e7ed;border-radius:6px}.section-heading{display:flex;align-items:flex-start;gap:10px}.section-heading>.el-icon{margin-top:2px;color:#337ecc;font-size:20px}.section-heading h2{margin:0;font-size:17px;letter-spacing:0}.section-heading p{margin:5px 0 0;color:#909399;font-size:13px}.owner-heading{margin-top:28px;padding-top:24px;border-top:1px solid #ebeef5}.provision-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:0 18px;margin-top:18px}.provision-grid .el-select{width:100%}.apk-field{display:flex;align-items:center;gap:8px;min-width:0}.form-actions{display:flex;justify-content:flex-end;margin-top:18px}.delivery-panel{max-width:980px}.delivery-grid{display:grid;grid-template-columns:190px minmax(0,1fr);margin:20px 0;border:1px solid #e4e7ed;border-bottom:0}.delivery-grid dt,.delivery-grid dd{min-width:0;margin:0;padding:12px 14px;border-bottom:1px solid #e4e7ed}.delivery-grid dt{background:#f5f7fa;color:#606266;font-weight:600}.delivery-grid dd{overflow-wrap:anywhere}.delivery-grid .secret{color:#c45656;font-family:ui-monospace,SFMono-Regular,Menlo,monospace;font-weight:700}.delivery-actions{display:flex;flex-wrap:wrap;gap:10px}@media(max-width:980px){.provision-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:640px){.provision-grid{grid-template-columns:1fr}.delivery-grid{grid-template-columns:1fr}.delivery-grid dt{border-bottom:0}.page-heading{flex-direction:column}}
</style>
