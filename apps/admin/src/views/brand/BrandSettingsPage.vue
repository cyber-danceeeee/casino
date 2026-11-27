<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { useI18n } from "vue-i18n";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import H5PreviewFrame from "@/components/brand/H5PreviewFrame.vue";
const { t } = useI18n();
const loading = ref(false),
  saving = ref(false),
  previewVisible = ref(false);
const overview = ref<BrandOverview>();
const form = reactive({
  siteName: "",
  h5Domain: "",
  appName: "",
  customerService: "",
  telegram: "",
  email: "",
  seoTitle: "",
  seoDescription: "",
  maintenanceEnabled: false,
  maintenanceMessage: "",
  reason: "",
});
function object(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}
async function load() {
  loading.value = true;
  try {
    const row = await systemAPI.brand.overview();
    overview.value = row;
    const config = object(row.configuration),
      brand = object(config.brandConfig),
      seo = object(config.seoConfig),
      maintenance = object(config.maintenanceConfig);
    Object.assign(form, {
      siteName: row.brandName,
      h5Domain: row.h5Domain,
      appName: String(brand.appName || row.brandName || ""),
      customerService: String(brand.customerService || ""),
      telegram: String(brand.telegram || ""),
      email: String(brand.email || ""),
      seoTitle: String(seo.title || ""),
      seoDescription: String(seo.description || ""),
      maintenanceEnabled: Boolean(maintenance.enabled),
      maintenanceMessage: String(maintenance.message || ""),
      reason: "",
    });
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : t("common.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}
async function save() {
  if (!overview.value || saving.value) return;
  saving.value = true;
  try {
    const config = object(overview.value.configuration);
    await systemAPI.brand.save({
      ...config,
      siteName: form.siteName.trim(),
      h5Domain: form.h5Domain.trim(),
      brandConfig: {
        ...object(config.brandConfig),
        appName: form.appName.trim(),
        customerService: form.customerService.trim(),
        telegram: form.telegram.trim(),
        email: form.email.trim(),
      },
      seoConfig: {
        ...object(config.seoConfig),
        title: form.seoTitle.trim(),
        description: form.seoDescription.trim(),
      },
      maintenanceConfig: {
        ...object(config.maintenanceConfig),
        enabled: form.maintenanceEnabled,
        message: form.maintenanceMessage.trim(),
      },
      reason: form.reason.trim() || "update H5 business settings",
    });
    ElMessage.success(t("liveConfig.saved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : t("common.saveFailed"),
    );
  } finally {
    saving.value = false;
  }
}
onMounted(load);
</script>
<template>
  <div v-loading="loading" class="settings">
    <header>
      <div>
        <h1>{{ t("singleSite.settings") }}</h1>
        <p>{{ t("liveConfig.immediate") }}</p>
      </div>
      <el-button @click="previewVisible = true">{{
        t("singleSite.viewH5")
      }}</el-button>
    </header>
    <el-alert
      :title="t('liveConfig.fixedMedia')"
      type="info"
      :closable="false"
      show-icon
    />
    <el-form label-position="top" @submit.prevent="save">
      <el-form-item :label="t('brandPortal.settings.fields.brandName')"
        ><el-input v-model="form.siteName"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.appName')"
        ><el-input v-model="form.appName"
      /></el-form-item>
      <el-form-item :label="t('singleSite.h5Address')"
        ><el-input v-model="form.h5Domain"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.sections.locale')"
        ><span
          >{{ overview?.countryCode }} / {{ overview?.currencyCode }} ·
          {{ overview?.defaultLanguage }} · {{ overview?.timezone }}</span
        ><router-link to="/system/countries">{{
          t("liveConfig.changeCountry")
        }}</router-link></el-form-item
      >
      <el-form-item :label="t('brandPortal.settings.fields.customerService')"
        ><el-input v-model="form.customerService"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.telegram')"
        ><el-input v-model="form.telegram"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.email')"
        ><el-input v-model="form.email"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.seoTitle')"
        ><el-input v-model="form.seoTitle"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.seoDescription')"
        ><el-input v-model="form.seoDescription" type="textarea"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.maintenance')"
        ><el-switch v-model="form.maintenanceEnabled"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.message')"
        ><el-input v-model="form.maintenanceMessage" type="textarea"
      /></el-form-item>
      <el-form-item :label="t('brandPortal.settings.fields.reason')"
        ><el-input v-model="form.reason"
      /></el-form-item>
      <div class="actions">
        <el-button @click="load">{{ t("common.reset") }}</el-button
        ><el-button type="primary" :loading="saving" @click="save">{{
          t("common.save")
        }}</el-button>
      </div>
    </el-form>
    <el-dialog
      v-model="previewVisible"
      :title="t('singleSite.viewH5')"
      width="min(920px,94vw)"
      destroy-on-close
      ><H5PreviewFrame
        :h5-domain="overview?.h5Domain"
        :site-i-d="overview?.siteID"
    /></el-dialog>
  </div>
</template>
<style scoped>
.settings {
  display: grid;
  gap: 18px;
  max-width: 1100px;
}
header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
h1 {
  margin: 0;
}
p {
  color: var(--el-text-color-secondary);
}
.el-form {
  padding: 24px;
  border: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
}
.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
a {
  margin-left: 16px;
}
</style>
