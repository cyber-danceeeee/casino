<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { useI18n } from "vue-i18n";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import BrandSetupProgress from "@/components/brand/BrandSetupProgress.vue";
import { formatBackendTime } from "@/utils/timezone";

const { t } = useI18n();
const loading = ref(false);
const overview = ref<BrandOverview>();
async function load() {
  loading.value = true;
  try {
    overview.value = await systemAPI.brand.overview();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("tenantExperience.common.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}
onMounted(load);
</script>
<template>
  <div v-loading="loading" class="brand-overview-page">
    <header class="page-header">
      <div>
        <p>{{ t("tenantExperience.brand.eyebrow") }}</p>
        <h1>{{ t("tenantExperience.brand.title") }}</h1>
        <span>{{ t("tenantExperience.brand.description") }}</span>
      </div>
    </header>
    <BrandSetupProgress v-if="overview" :overview="overview" />
    <section v-if="overview" class="business-details">
      <h2>{{ overview.brandName }}</h2>
      <dl>
        <dt>{{ t("tenantExperience.fields.brandCode") }}</dt>
        <dd>{{ overview.tenantCode || "-" }}</dd>
        <dt>{{ t("tenantExperience.fields.site") }}</dt>
        <dd>{{ overview.currentSite }} · {{ overview.siteCode }}</dd>
        <dt>{{ t("tenantExperience.fields.market") }}</dt>
        <dd>
          {{ overview.countryCode }} · {{ overview.currencyCode }} ·
          {{ overview.defaultLanguage }}
        </dd>
        <dt>{{ t("tenantExperience.fields.timezone") }}</dt>
        <dd>{{ overview.timezone }}</dd>
        <dt>{{ t("tenantExperience.fields.h5Domain") }}</dt>
        <dd>{{ overview.h5Domain || "-" }}</dd>
        <dt>{{ t("tenantExperience.fields.publishVersion") }}</dt>
        <dd>v{{ overview.publishedVersion || 1 }}</dd>
        <dt>{{ t("tenantExperience.fields.lastPublished") }}</dt>
        <dd>{{ formatBackendTime(overview.publishedAt || "") }}</dd>
      </dl>
    </section>
  </div>
</template>
<style scoped>
.brand-overview-page {
  display: grid;
  gap: 16px;
}
.page-header,
.business-details {
  padding: 20px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
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
.page-header span,
dt {
  color: var(--el-text-color-secondary);
}
.business-details h2 {
  margin: 0 0 16px;
  font-size: 19px;
}
.business-details dl {
  display: grid;
  grid-template-columns: 160px minmax(0, 1fr);
  gap: 12px;
  margin: 0;
}
.business-details dd {
  margin: 0;
  overflow-wrap: anywhere;
}
@media (max-width: 640px) {
  .business-details dl {
    grid-template-columns: 1fr;
    gap: 4px;
  }
  .business-details dd {
    margin-bottom: 10px;
  }
}
</style>
