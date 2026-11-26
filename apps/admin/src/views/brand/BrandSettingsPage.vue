<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Monitor, UploadFilled } from "@element-plus/icons-vue";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import H5PreviewFrame from "@/components/brand/H5PreviewFrame.vue";
import { formatBackendTime } from "@/utils/timezone";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const publishing = ref(false);
const previewVisible = ref(false);
const previewLoading = ref(false);
const overview = ref<BrandOverview | null>(null);

const form = reactive({
  siteName: "",
  logo: "",
  appName: "",
  customerService: "",
  telegram: "",
  email: "",
  countryCode: "",
  currencyCode: "",
  defaultLanguage: "",
  timezone: "",
  seoTitle: "",
  seoDescription: "",
  primaryColor: "#E53935",
  secondaryColor: "#FFB300",
  buttonColor: "#E53935",
  backgroundColor: "#070A12",
  textColor: "#FFFFFF",
  maintenanceEnabled: false,
  maintenanceMessage: "",
  reason: "",
});

const statusType = computed(() =>
  String(overview.value?.publishStatus || "").toUpperCase() === "DRAFT" ? "warning" : "success",
);

function objectValue(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? value as Record<string, unknown>
    : {};
}

function assignFromOverview(row: BrandOverview) {
  const draft = objectValue(row.draftConfig);
  const brand = objectValue(draft.brandConfig);
  const theme = objectValue(draft.themeConfig);
  const seo = objectValue(draft.seoConfig);
  const maintenance = objectValue(draft.maintenanceConfig);
  Object.assign(form, {
    siteName: String(draft.siteName || row.brandName || ""),
    logo: String(brand.logo || ""),
    appName: String(brand.appName || row.brandName || ""),
    customerService: String(brand.customerService || ""),
    telegram: String(brand.telegram || ""),
    email: String(brand.email || ""),
    countryCode: String(draft.countryCode || row.countryCode || ""),
    currencyCode: String(draft.currencyCode || row.currencyCode || ""),
    defaultLanguage: String(draft.defaultLanguage || row.defaultLanguage || "en"),
    timezone: String(draft.timezone || row.timezone || "Asia/Kuala_Lumpur"),
    seoTitle: String(seo.title || ""),
    seoDescription: String(seo.description || ""),
    primaryColor: String(theme.primary || "#E53935"),
    secondaryColor: String(theme.secondary || "#FFB300"),
    buttonColor: String(theme.button || theme.primary || "#E53935"),
    backgroundColor: String(theme.background || "#070A12"),
    textColor: String(theme.text || "#FFFFFF"),
    maintenanceEnabled: Boolean(maintenance.enabled),
    maintenanceMessage: String(maintenance.message || ""),
    reason: "",
  });
}

async function load() {
  loading.value = true;
  try {
    const row = await systemAPI.brand.overview();
    overview.value = row;
    assignFromOverview(row);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.settingsLoadFailed"));
  } finally {
    loading.value = false;
  }
}

function payload() {
  return {
    siteName: form.siteName.trim(),
    countryCode: form.countryCode.trim().toUpperCase(),
    currencyCode: form.currencyCode.trim().toUpperCase(),
    defaultLanguage: form.defaultLanguage.trim(),
    timezone: form.timezone.trim(),
    brandConfig: {
      appName: form.appName.trim() || form.siteName.trim(),
      logo: form.logo.trim(),
      customerService: form.customerService.trim(),
      telegram: form.telegram.trim(),
      email: form.email.trim(),
    },
    themeConfig: {
      primary: form.primaryColor,
      secondary: form.secondaryColor,
      button: form.buttonColor,
      background: form.backgroundColor,
      text: form.textColor,
    },
    featureConfig: objectValue(overview.value?.draftConfig?.featureConfig),
    maintenanceConfig: {
      enabled: form.maintenanceEnabled,
      message: form.maintenanceMessage.trim(),
    },
    seoConfig: {
      title: form.seoTitle.trim(),
      description: form.seoDescription.trim(),
    },
    reason: form.reason.trim() || t("brandPortal.defaults.saveBrandDraft"),
  };
}

async function saveDraft() {
  if (!overview.value?.siteID) return;
  saving.value = true;
  try {
    await systemAPI.sites.update(overview.value.siteID, payload());
    ElMessage.success(t("brandPortal.messages.draftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.saveFailed"));
  } finally {
    saving.value = false;
  }
}

async function preview() {
  if (!overview.value?.siteID) return;
  previewVisible.value = true;
  previewLoading.value = true;
  try {
    await systemAPI.sites.preview(overview.value.siteID);
  } catch (error) {
    previewVisible.value = false;
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.previewFailed"));
  } finally {
    previewLoading.value = false;
  }
}

async function publish() {
  if (!overview.value?.siteID) return;
  try {
    const result = await ElMessageBox.prompt(t("brandPortal.prompts.publishReason"), t("brandPortal.dialogs.publishTitle"), {
      confirmButtonText: t("brandPortal.actions.publish"),
      cancelButtonText: t("brandPortal.actions.cancel"),
      inputValue: t("brandPortal.defaults.publishBrandConfig"),
      inputValidator: (value) => Boolean(value.trim()) || t("brandPortal.validation.reasonRequired"),
    });
    publishing.value = true;
    await systemAPI.sites.publish(overview.value.siteID, result.value.trim());
    ElMessage.success(t("brandPortal.messages.publishedToH5"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.publishFailed"));
  } finally {
    publishing.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div v-loading="loading" class="brand-settings">
    <section class="settings-header">
      <div>
        <p class="eyebrow">{{ t("brandPortal.settings.eyebrow") }}</p>
        <h1>{{ overview?.brandName || "-" }}</h1>
        <div class="status-row">
          <el-tag :type="statusType" effect="plain">{{ overview?.publishStatus || "-" }}</el-tag>
          <span>{{ t("brandPortal.statusBar.publishedVersion", { version: overview?.publishedVersion || "-" }) }}</span>
          <span>{{ t("brandPortal.statusBar.draftVersion", { version: overview?.draftVersion || "-" }) }}</span>
          <span>{{ t("brandPortal.statusBar.lastPublished", { time: formatBackendTime(String(overview?.publishedAt || "")) }) }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button :icon="Monitor" @click="preview">{{ t("brandPortal.actions.preview") }}</el-button>
        <el-button type="primary" :icon="UploadFilled" :loading="publishing" @click="publish">
          {{ t("brandPortal.actions.publish") }}
        </el-button>
      </div>
    </section>

    <el-alert
      type="warning"
      :title="t('brandPortal.settings.draftWarning')"
      :closable="false"
      show-icon
    />

    <el-form class="settings-form" label-position="left" label-width="150px">
      <section class="form-section">
        <h2>{{ t("brandPortal.settings.sections.brand") }}</h2>
        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-form-item :label="t('brandPortal.settings.fields.brandName')">
              <el-input v-model="form.siteName" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item :label="t('brandPortal.settings.fields.appName')">
              <el-input v-model="form.appName" />
            </el-form-item>
          </el-col>
          <el-col :xs="24">
            <el-form-item :label="t('brandPortal.settings.fields.logo')">
              <el-input v-model="form.logo" placeholder="https://..." />
            </el-form-item>
          </el-col>
        </el-row>
      </section>

      <section class="form-section">
        <h2>{{ t("brandPortal.settings.sections.theme") }}</h2>
        <el-row :gutter="16">
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.primary')"><div class="color-field"><el-color-picker v-model="form.primaryColor" /><el-input v-model="form.primaryColor" /></div></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.secondary')"><div class="color-field"><el-color-picker v-model="form.secondaryColor" /><el-input v-model="form.secondaryColor" /></div></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.button')"><div class="color-field"><el-color-picker v-model="form.buttonColor" /><el-input v-model="form.buttonColor" /></div></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.background')"><div class="color-field"><el-color-picker v-model="form.backgroundColor" /><el-input v-model="form.backgroundColor" /></div></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.text')"><div class="color-field"><el-color-picker v-model="form.textColor" /><el-input v-model="form.textColor" /></div></el-form-item></el-col>
        </el-row>
      </section>

      <section class="form-section">
        <h2>{{ t("brandPortal.settings.sections.locale") }}</h2>
        <el-row :gutter="16">
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.country')"><el-input v-model="form.countryCode" disabled /></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.currency')"><el-input v-model="form.currencyCode" disabled /></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.language')"><el-input v-model="form.defaultLanguage" /></el-form-item></el-col>
          <el-col :xs="24" :md="8"><el-form-item :label="t('brandPortal.settings.fields.timezone')"><el-input v-model="form.timezone" /></el-form-item></el-col>
        </el-row>
      </section>

      <section class="form-section">
        <h2>{{ t("brandPortal.settings.sections.supportSeo") }}</h2>
        <el-row :gutter="16">
          <el-col :xs="24" :md="12"><el-form-item :label="t('brandPortal.settings.fields.customerService')"><el-input v-model="form.customerService" /></el-form-item></el-col>
          <el-col :xs="24" :md="12"><el-form-item :label="t('brandPortal.settings.fields.telegram')"><el-input v-model="form.telegram" /></el-form-item></el-col>
          <el-col :xs="24" :md="12"><el-form-item :label="t('brandPortal.settings.fields.email')"><el-input v-model="form.email" /></el-form-item></el-col>
          <el-col :xs="24" :md="12"><el-form-item :label="t('brandPortal.settings.fields.seoTitle')"><el-input v-model="form.seoTitle" /></el-form-item></el-col>
          <el-col :xs="24"><el-form-item :label="t('brandPortal.settings.fields.seoDescription')"><el-input v-model="form.seoDescription" type="textarea" /></el-form-item></el-col>
        </el-row>
      </section>

      <section class="form-section">
        <h2>{{ t("brandPortal.settings.sections.maintenance") }}</h2>
        <el-form-item :label="t('brandPortal.settings.fields.maintenance')">
          <el-switch v-model="form.maintenanceEnabled" />
        </el-form-item>
        <el-form-item :label="t('brandPortal.settings.fields.message')">
          <el-input v-model="form.maintenanceMessage" type="textarea" />
        </el-form-item>
        <el-form-item :label="t('brandPortal.settings.fields.reason')">
          <el-input v-model="form.reason" type="textarea" :placeholder="t('brandPortal.defaults.saveBrandDraft')" />
        </el-form-item>
      </section>
    </el-form>

    <section class="footer-actions">
      <el-button @click="load">{{ t("brandPortal.actions.reset") }}</el-button>
      <el-button type="primary" :loading="saving" @click="saveDraft">{{ t("brandPortal.actions.saveDraft") }}</el-button>
    </section>

    <el-dialog v-model="previewVisible" :title="t('brandPortal.dialogs.previewTitle')" width="min(920px, 94vw)" destroy-on-close>
      <H5PreviewFrame v-loading="previewLoading" :h5-domain="overview?.h5Domain" :site-i-d="overview?.siteID" />
      <template #footer>
        <el-button @click="previewVisible = false">{{ t("brandPortal.actions.close") }}</el-button>
        <el-button type="primary" :loading="publishing" @click="publish">{{ t("brandPortal.actions.publish") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.brand-settings {
  display: grid;
  gap: 16px;
}
.settings-header,
.form-section,
.footer-actions {
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px;
}
.eyebrow {
  margin: 0 0 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  text-transform: uppercase;
}
h1,
h2 {
  margin: 0;
}
.status-row,
.header-actions,
.footer-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}
.status-row {
  margin-top: 12px;
  color: var(--el-text-color-secondary);
}
.settings-form {
  display: grid;
  gap: 16px;
}
.form-section {
  padding: 18px;
}
.form-section h2 {
  margin-bottom: 14px;
  font-size: 16px;
}
.color-field {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 8px;
  width: 100%;
}
.footer-actions {
  justify-content: flex-end;
  padding: 14px;
}
@media (max-width: 900px) {
  .settings-header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
