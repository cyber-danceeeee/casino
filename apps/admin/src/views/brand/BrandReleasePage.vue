<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Monitor, UploadFilled } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import H5PreviewFrame from "@/components/brand/H5PreviewFrame.vue";
import { formatBackendTime } from "@/utils/timezone";
import { businessEntries, humanizeBusinessKey, sanitizeBusinessValue } from "@/utils/presentation";

const { t } = useI18n();
const loading = ref(false);
const publishing = ref(false);
const overview = ref<BrandOverview>();
const hasChanges = computed(() =>
  Boolean(overview.value?.hasUnpublished && overview.value.unpublishedChanges),
);
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
async function publish() {
  if (!overview.value?.siteID || !hasChanges.value) return;
  try {
    const result = await ElMessageBox.prompt(
      t("brandPortal.prompts.publishReason"),
      t("brandPortal.dialogs.publishTitle"),
      {
        confirmButtonText: t("brandPortal.actions.publish"),
        cancelButtonText: t("brandPortal.actions.cancel"),
        inputValidator: (value) =>
          Boolean(value.trim()) || t("brandPortal.validation.reasonRequired"),
      },
    );
    publishing.value = true;
    await systemAPI.sites.publish(overview.value.siteID, result.value.trim());
    ElMessage.success(t("brandPortal.messages.publishedToH5"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("brandPortal.messages.publishFailed"),
      );
  } finally {
    publishing.value = false;
  }
}
onMounted(load);
</script>
<template>
  <div v-loading="loading" class="release-page">
    <header class="page-header">
      <div>
        <p>{{ t("tenantExperience.release.eyebrow") }}</p>
        <h1>{{ t("tenantExperience.release.title") }}</h1>
        <span>{{ t("tenantExperience.release.description") }}</span>
      </div>
      <el-button
        type="primary"
        :icon="UploadFilled"
        :disabled="!hasChanges"
        :loading="publishing"
        @click="publish"
        >{{
          hasChanges
            ? t("brandPortal.actions.publish")
            : t("tenantExperience.release.noChanges")
        }}</el-button
      >
    </header>
    <section class="versions">
      <article>
        <span>{{ t("tenantExperience.release.published") }}</span
        ><strong>v{{ overview?.publishedVersion || 1 }}</strong
        ><small>{{ formatBackendTime(overview?.publishedAt || "") }}</small>
      </article>
      <article>
        <span>{{ t("tenantExperience.release.draft") }}</span
        ><strong>v{{ overview?.draftVersion || 1 }}</strong
        ><small>{{ formatBackendTime(overview?.draftUpdatedAt || "") }}</small>
      </article>
      <article>
        <span>{{ t("tenantExperience.release.changes") }}</span
        ><strong>{{ overview?.unpublishedChanges || 0 }}</strong
        ><small>{{ overview?.lastUpdatedBy || "-" }}</small>
      </article>
    </section>
    <section class="content-grid">
      <article class="preview">
        <header>
          <Monitor />
          <h2>{{ t("tenantExperience.release.h5Preview") }}</h2>
        </header>
        <H5PreviewFrame
          :h5-domain="overview?.h5Domain"
          :site-i-d="overview?.siteID"
        />
      </article>
      <article class="changes">
        <h2>{{ t("tenantExperience.release.changeList") }}</h2>
        <div
          v-for="change in overview?.pendingChanges || []"
          :key="change.module"
          class="change"
        >
          <header>
            <strong>{{
              t(`tenantExperience.modules.${change.module}`, change.module)
            }}</strong
            ><span
              >{{ change.updatedBy || "-" }} ·
              {{ formatBackendTime(change.updatedAt || "") }}</span
            >
          </header>
          <p>
            {{
              t(
                `tenantExperience.changeSummary.${change.module}`,
                change.summary,
              )
            }}
          </p>
          <div v-if="change.after" class="diff">
            <div
              v-for="[key, value] in businessEntries(change.after)"
              :key="key"
            >
              <span>{{
                t(
                  `tenantExperience.configKeys.${key}`,
                  humanizeBusinessKey(key),
                )
              }}</span
              ><strong>{{
                typeof value === "object" ? JSON.stringify(sanitizeBusinessValue(value)) : value
              }}</strong>
            </div>
          </div>
        </div>
        <el-empty
          v-if="!overview?.pendingChanges?.length"
          :description="t('tenantExperience.release.noChanges')"
        />
      </article>
    </section>
  </div>
</template>
<style scoped>
.release-page {
  display: grid;
  gap: 16px;
}
.page-header,
.versions,
.preview,
.changes {
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  padding: 20px;
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
.versions span,
.versions small,
.change header span,
.change p,
.diff span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.versions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
}
.versions article {
  display: grid;
  gap: 5px;
  padding: 16px;
  border-right: 1px solid var(--el-border-color-lighter);
}
.versions article:last-child {
  border: 0;
}
.versions strong {
  font-size: 24px;
}
.content-grid {
  display: grid;
  grid-template-columns: minmax(380px, 0.8fr) minmax(0, 1.2fr);
  gap: 16px;
}
.preview,
.changes {
  padding: 17px;
}
.preview > header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}
.preview h2,
.changes h2 {
  margin: 0;
  font-size: 17px;
}
.change {
  padding: 13px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.change header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}
.change p {
  margin: 6px 0;
}
.diff {
  display: grid;
  gap: 6px;
  margin-top: 9px;
}
.diff div {
  display: grid;
  grid-template-columns: 150px minmax(0, 1fr);
  gap: 10px;
}
.diff strong {
  font-size: 13px;
  overflow-wrap: anywhere;
}
@media (max-width: 900px) {
  .page-header {
    flex-direction: column;
  }
  .content-grid,
  .versions {
    grid-template-columns: 1fr;
  }
  .versions article {
    border-right: 0;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }
}
</style>
