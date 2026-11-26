<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Document, Monitor, UploadFilled } from "@element-plus/icons-vue";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import H5PreviewFrame from "@/components/brand/H5PreviewFrame.vue";
import { formatBackendTime } from "@/utils/timezone";
import { useI18n } from "vue-i18n";
import {
  businessEntries,
  humanizeBusinessKey,
  statusKey,
} from "@/utils/presentation";

const route = useRoute();
const { t } = useI18n();
const loading = ref(false);
const publishing = ref(false);
const previewVisible = ref(false);
const previewLoading = ref(false);
const changesVisible = ref(false);
const overview = ref<BrandOverview | null>(null);
let refreshTimer: number | undefined;

const statusType = computed(() =>
  String(overview.value?.publishStatus || "").toUpperCase() === "DRAFT"
    ? "warning"
    : "success",
);

const unpublishedCount = computed(() =>
  Number(overview.value?.unpublishedChanges || 0),
);

async function load() {
  if (loading.value) return;
  loading.value = true;
  try {
    overview.value = await systemAPI.brand.overview();
  } catch {
    overview.value = null;
  } finally {
    loading.value = false;
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
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("brandPortal.messages.previewFailed"),
    );
  } finally {
    previewLoading.value = false;
  }
}

async function publish() {
  if (!overview.value?.siteID) return;
  try {
    const result = await ElMessageBox.prompt(
      t("brandPortal.prompts.publishReason"),
      t("brandPortal.dialogs.publishTitle"),
      {
        confirmButtonText: t("brandPortal.actions.publish"),
        cancelButtonText: t("brandPortal.actions.cancel"),
        inputValue: t("brandPortal.defaults.publishBrandConfig"),
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
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("brandPortal.messages.publishFailed"),
      );
    }
  } finally {
    publishing.value = false;
  }
}

function handleRefresh() {
  void load();
}

watch(
  () => route.fullPath,
  () => {
    void load();
  },
);

onMounted(() => {
  void load();
  window.addEventListener("brand-config-refresh", handleRefresh);
  refreshTimer = window.setInterval(handleRefresh, 15000);
});

onBeforeUnmount(() => {
  window.removeEventListener("brand-config-refresh", handleRefresh);
  if (refreshTimer) window.clearInterval(refreshTimer);
});
</script>

<template>
  <section v-if="overview" v-loading="loading" class="brand-draft-bar">
    <div class="brand-draft-bar__main">
      <strong>{{ overview.brandName || overview.currentSite || "-" }}</strong>
      <el-tag :type="statusType" effect="plain" size="small">
        {{
          t(
            `tenantExperience.status.${statusKey(overview.publishStatus)}`,
            overview.publishStatus || "-",
          )
        }}
      </el-tag>
      <span>{{
        t("brandPortal.statusBar.publishedVersion", {
          version: overview.publishedVersion || "-",
        })
      }}</span>
      <span>{{
        t("brandPortal.statusBar.draftVersion", {
          version: overview.draftVersion || "-",
        })
      }}</span>
      <span>{{
        t("brandPortal.statusBar.unpublishedChanges", {
          count: unpublishedCount,
        })
      }}</span>
      <span>{{
        t("brandPortal.statusBar.lastPublished", {
          time: formatBackendTime(String(overview.publishedAt || "")),
        })
      }}</span>
      <span>{{
        t("tenantExperience.release.lastUpdated", {
          time: formatBackendTime(String(overview.draftUpdatedAt || "")),
        })
      }}</span>
    </div>
    <div class="brand-draft-bar__actions">
      <el-button
        :icon="Document"
        size="small"
        :disabled="!unpublishedCount"
        @click="changesVisible = true"
        >{{ t("tenantExperience.release.viewChanges") }}</el-button
      >
      <el-button :icon="Monitor" size="small" @click="preview">{{
        t("brandPortal.actions.preview")
      }}</el-button>
      <el-button
        type="primary"
        :icon="UploadFilled"
        size="small"
        :loading="publishing"
        :disabled="!unpublishedCount"
        @click="publish"
      >
        {{ t("brandPortal.actions.publish") }}
      </el-button>
    </div>

    <el-dialog
      v-model="previewVisible"
      :title="t('brandPortal.dialogs.previewTitle')"
      width="min(920px, 94vw)"
      destroy-on-close
    >
      <H5PreviewFrame
        v-loading="previewLoading"
        :h5-domain="overview?.h5Domain"
        :site-i-d="overview?.siteID"
      />
      <template #footer>
        <el-button @click="previewVisible = false">{{
          t("brandPortal.actions.close")
        }}</el-button>
        <el-button type="primary" :loading="publishing" @click="publish">{{
          t("brandPortal.actions.publish")
        }}</el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="changesVisible"
      :title="t('tenantExperience.release.changeList')"
      size="min(560px, 92vw)"
    >
      <div
        v-for="change in overview.pendingChanges || []"
        :key="change.module"
        class="change-item"
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
            t(`tenantExperience.changeSummary.${change.module}`, change.summary)
          }}
        </p>
        <div v-if="change.after" class="change-diff">
          <div v-for="[key, value] in businessEntries(change.after)" :key="key">
            <span>{{
              t(`tenantExperience.configKeys.${key}`, humanizeBusinessKey(key))
            }}</span
            ><strong>{{
              typeof value === "object" ? JSON.stringify(value) : value
            }}</strong>
          </div>
        </div>
      </div>
      <el-empty
        v-if="!overview.pendingChanges?.length"
        :description="t('tenantExperience.release.noChanges')"
      />
    </el-drawer>
  </section>
</template>

<style scoped>
.brand-draft-bar {
  position: sticky;
  top: 0;
  z-index: 8;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 44px;
  margin: 0 0 14px;
  padding: 8px 12px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.06);
}
.brand-draft-bar__main,
.brand-draft-bar__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.brand-draft-bar__main {
  min-width: 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.brand-draft-bar__main strong {
  color: var(--el-text-color-primary);
}
.change-item {
  padding: 14px 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.change-item header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}
.change-item header span,
.change-item p,
.change-diff span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.change-item p {
  margin: 7px 0;
}
.change-diff {
  display: grid;
  gap: 6px;
}
.change-diff div {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr);
  gap: 10px;
}
.change-diff strong {
  font-size: 13px;
  overflow-wrap: anywhere;
}
@media (max-width: 900px) {
  .brand-draft-bar {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
