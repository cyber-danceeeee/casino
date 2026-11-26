<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import {
  ElMessage,
  ElMessageBox,
  type UploadRequestOptions,
} from "element-plus";
import {
  CopyDocument,
  Delete,
  Edit,
  Link as LinkIcon,
  Refresh,
  Search,
  Upload,
  View,
} from "@element-plus/icons-vue";
import {
  deleteAsset,
  getAsset,
  listAssets,
  updateAsset,
  uploadAssetLibrary,
  type AssetDetail,
  type AssetLibraryRecord,
  type AssetScope,
  type AssetType,
  type AssetUsage,
} from "@/api/adminOperations";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime } from "@/utils/timezone";
import { useI18n } from "vue-i18n";

const { actor, loadPermissions } = useAdminPermissions();
const { t } = useI18n();

const assetTypes = computed<Array<{ label: string; value: AssetType }>>(() => [
  { label: t("assetLibrary.types.IMAGE"), value: "IMAGE" },
  { label: t("assetLibrary.types.AUDIO"), value: "AUDIO" },
  { label: t("assetLibrary.types.VIDEO"), value: "VIDEO" },
  { label: t("assetLibrary.types.LOTTIE"), value: "LOTTIE" },
]);

const usageOptions = computed<Array<{ label: string; value: AssetUsage }>>(() => [
  { label: t("assetLibrary.usages.GENERAL"), value: "GENERAL" },
  { label: t("assetLibrary.usages.LOGO"), value: "LOGO" },
  { label: t("assetLibrary.usages.FAVICON"), value: "FAVICON" },
  { label: t("assetLibrary.usages.BANNER"), value: "BANNER" },
  { label: t("assetLibrary.usages.BACKGROUND"), value: "BACKGROUND" },
  { label: t("assetLibrary.usages.BUTTON_ICON"), value: "BUTTON_ICON" },
  { label: t("assetLibrary.usages.GAME_PLACEHOLDER"), value: "GAME_PLACEHOLDER" },
  { label: t("assetLibrary.usages.POPUP"), value: "POPUP" },
  { label: t("assetLibrary.usages.LOADING"), value: "LOADING" },
  { label: t("assetLibrary.usages.SOUND"), value: "SOUND" },
]);

const filters = reactive({
  keyword: "",
  assetType: "",
  usageType: "",
  assetScope: "",
});
const pagination = reactive({ page: 1, pageSize: 20 });
const rows = ref<AssetLibraryRecord[]>([]);
const total = ref(0);
const loading = ref(false);

const uploadVisible = ref(false);
const uploading = ref(false);
const uploadForm = reactive<{
  assetType: AssetType | "";
  usageType: AssetUsage;
  assetScope: AssetScope;
  displayName: string;
  altText: string;
  tags: string;
}>({
  assetType: "",
  usageType: "GENERAL",
  assetScope: "SITE",
  displayName: "",
  altText: "",
  tags: "",
});

const editVisible = ref(false);
const saving = ref(false);
const editing = ref<AssetLibraryRecord | null>(null);
const editForm = reactive({
  usageType: "GENERAL" as AssetUsage,
  displayName: "",
  altText: "",
  tags: "",
  reason: "",
});

const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<AssetDetail | null>(null);

const canUploadPlatformAssets = computed(() => Boolean(actor.value?.platform));
const scopeOptions = computed<Array<{ label: string; value: AssetScope }>>(() => [
  { label: t("assetLibrary.scopes.PLATFORM"), value: "PLATFORM" },
  { label: t("assetLibrary.scopes.SITE"), value: "SITE" },
]);
const uploadScopeOptions = computed(() =>
  canUploadPlatformAssets.value
    ? scopeOptions.value
    : scopeOptions.value.filter((item) => item.value === "SITE"),
);

async function loadRows() {
  loading.value = true;
  try {
    const result = await listAssets({
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: filters.keyword || undefined,
      assetType: filters.assetType || undefined,
      usageType: filters.usageType || undefined,
      assetScope: filters.assetScope || undefined,
    });
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("assetLibrary.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}

function resetFilters() {
  filters.keyword = "";
  filters.assetType = "";
  filters.usageType = "";
  filters.assetScope = "";
  pagination.page = 1;
  void loadRows();
}

function resetUploadForm() {
  uploadForm.assetType = "";
  uploadForm.usageType = "GENERAL";
  uploadForm.assetScope = canUploadPlatformAssets.value ? "PLATFORM" : "SITE";
  uploadForm.displayName = "";
  uploadForm.altText = "";
  uploadForm.tags = "";
}

async function handleUpload(options: UploadRequestOptions) {
  uploading.value = true;
  try {
    const asset = await uploadAssetLibrary(options.file, {
      assetType: uploadForm.assetType || undefined,
      usageType: uploadForm.usageType,
      assetScope: uploadForm.assetScope,
      displayName: uploadForm.displayName,
      altText: uploadForm.altText,
      tags: uploadForm.tags,
    });
    options.onSuccess(asset);
    ElMessage.success(t("assetLibrary.messages.uploaded"));
    uploadVisible.value = false;
    resetUploadForm();
    pagination.page = 1;
    await loadRows();
  } catch (error) {
    const uploadError =
      error instanceof Error ? error : new Error(t("assetLibrary.messages.uploadFailed"));
    options.onError(
      uploadError as Parameters<UploadRequestOptions["onError"]>[0],
    );
    ElMessage.error(uploadError.message);
  } finally {
    uploading.value = false;
  }
}

function openEdit(row: AssetLibraryRecord) {
  editing.value = row;
  editForm.usageType = row.usageType || "GENERAL";
  editForm.displayName = row.displayName || row.originalName || "";
  editForm.altText = row.altText || "";
  editForm.tags = row.tags || "";
  editForm.reason = "";
  editVisible.value = true;
}

async function saveEdit() {
  if (!editing.value) return;
  if (!editForm.reason.trim()) {
    ElMessage.warning(t("assetLibrary.validation.reasonRequired"));
    return;
  }
  saving.value = true;
  try {
    await updateAsset(editing.value.id, {
      usageType: editForm.usageType,
      displayName: editForm.displayName,
      altText: editForm.altText,
      tags: editForm.tags,
      reason: editForm.reason,
    });
    ElMessage.success(t("assetLibrary.messages.updated"));
    editVisible.value = false;
    await loadRows();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("assetLibrary.messages.updateFailed"));
  } finally {
    saving.value = false;
  }
}

async function openDetail(row: AssetLibraryRecord) {
  detailVisible.value = true;
  detailLoading.value = true;
  detail.value = null;
  try {
    detail.value = await getAsset(row.id);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("assetLibrary.messages.detailFailed"));
  } finally {
    detailLoading.value = false;
  }
}

async function confirmDelete(row: AssetLibraryRecord) {
  try {
    const result = await ElMessageBox.prompt(
      t("assetLibrary.dialogs.deletePrompt"),
      t("assetLibrary.dialogs.deleteTitle"),
      {
      confirmButtonText: t("assetLibrary.actions.delete"),
      cancelButtonText: t("assetLibrary.actions.cancel"),
      inputType: "textarea",
      inputValidator: (value) =>
        Boolean(value && value.trim()) || t("assetLibrary.validation.deleteReasonRequired"),
    },
    );
    await deleteAsset(row.id, String(result.value || ""));
    ElMessage.success(t("assetLibrary.messages.deleted"));
    await loadRows();
  } catch (error) {
    if (error === "cancel" || error === "close") return;
    ElMessage.error(error instanceof Error ? error.message : t("assetLibrary.messages.deleteFailed"));
  }
}

async function copyAssetLink(row: AssetLibraryRecord) {
  const link = row.publicUrl || row.objectKey;
  if (!link) return;
  try {
    await navigator.clipboard.writeText(link);
    ElMessage.success(t("assetLibrary.messages.linkCopied"));
  } catch {
    ElMessage.error(t("assetLibrary.messages.copyFailed"));
  }
}

function assetUrl(row: AssetLibraryRecord) {
  return row.publicUrl || row.objectKey || "";
}

function isPreviewImage(row: AssetLibraryRecord) {
  return row.assetType === "IMAGE" && Boolean(assetUrl(row));
}

function typeTag(type: AssetType) {
  if (type === "IMAGE") return "success";
  if (type === "AUDIO") return "warning";
  if (type === "VIDEO") return "primary";
  return "info";
}

function scopeTag(scope: AssetScope) {
  return scope === "PLATFORM" ? "warning" : "success";
}

function assetTypeLabel(value: string) {
  return t(`assetLibrary.types.${value}`);
}

function assetUsageLabel(value: string) {
  return t(`assetLibrary.usages.${value}`);
}

function assetScopeLabel(value: string) {
  return t(`assetLibrary.scopes.${value}`);
}

function formatSize(size: number) {
  if (!size) return "-";
  if (size >= 1024 * 1024) return `${(size / 1024 / 1024).toFixed(2)} MB`;
  if (size >= 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${size} B`;
}

function formatDuration(durationMs: number) {
  if (!durationMs) return "-";
  const seconds = Math.round(durationMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const remain = seconds % 60;
  return `${minutes}:${remain.toString().padStart(2, "0")}`;
}

function assetSpec(row: AssetLibraryRecord) {
  if (row.width && row.height) return `${row.width} x ${row.height}`;
  if (row.durationMs) return formatDuration(row.durationMs);
  return row.fileExt || "-";
}

function shortChecksum(value: string) {
  return value ? `${value.slice(0, 10)}...${value.slice(-6)}` : "-";
}

function handlePageChange(page: number) {
  pagination.page = page;
  void loadRows();
}

function handleSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.page = 1;
  void loadRows();
}

onMounted(async () => {
  await loadPermissions();
  resetUploadForm();
  await loadRows();
});
</script>

<template>
  <div class="asset-library-page">
    <header class="page-header">
      <div>
        <p class="eyebrow">{{ t("assetLibrary.eyebrow") }}</p>
        <h1>{{ t("assetLibrary.title") }}</h1>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="loadRows">{{ t("assetLibrary.actions.refresh") }}</el-button>
        <el-button type="primary" :icon="Upload" @click="uploadVisible = true">
          {{ t("assetLibrary.actions.upload") }}
        </el-button>
      </div>
    </header>

    <section class="toolbar">
      <el-form :inline="true" class="filter-form" @submit.prevent>
        <el-form-item :label="t('assetLibrary.fields.keyword')">
          <el-input
            v-model="filters.keyword"
            clearable
            :placeholder="t('assetLibrary.placeholders.keyword')"
            class="filter-input"
            @keyup.enter="pagination.page = 1; loadRows()"
          />
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.type')">
          <el-select v-model="filters.assetType" clearable :placeholder="t('assetLibrary.placeholders.all')" class="filter-select">
            <el-option v-for="item in assetTypes" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.usage')">
          <el-select v-model="filters.usageType" clearable :placeholder="t('assetLibrary.placeholders.all')" class="filter-select">
            <el-option v-for="item in usageOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="canUploadPlatformAssets" :label="t('assetLibrary.fields.scope')">
          <el-select v-model="filters.assetScope" clearable :placeholder="t('assetLibrary.placeholders.all')" class="filter-select">
            <el-option v-for="item in scopeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="pagination.page = 1; loadRows()">{{ t("assetLibrary.actions.search") }}</el-button>
          <el-button @click="resetFilters">{{ t("assetLibrary.actions.reset") }}</el-button>
        </el-form-item>
      </el-form>
    </section>

    <section class="table-panel">
      <div class="table-meta">
        <strong>{{ total }}</strong>
        <span>{{ t("assetLibrary.labels.assetUnit") }}</span>
      </div>
      <el-table :data="rows" row-key="id" v-loading="loading" class="asset-table">
        <el-table-column :label="t('assetLibrary.fields.preview')" width="112">
          <template #default="{ row }">
            <div class="preview-box">
              <el-image
                v-if="isPreviewImage(row)"
                :src="assetUrl(row)"
                fit="cover"
                :preview-src-list="[assetUrl(row)]"
                preview-teleported
              />
              <audio v-else-if="row.assetType === 'AUDIO' && assetUrl(row)" :src="assetUrl(row)" controls />
              <video v-else-if="row.assetType === 'VIDEO' && assetUrl(row)" :src="assetUrl(row)" muted />
              <el-tag v-else>{{ assetTypeLabel(row.assetType) }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.asset')" min-width="220">
          <template #default="{ row }">
            <div class="name-cell">
              <span class="asset-name">{{ row.displayName || row.originalName }}</span>
              <span>{{ row.originalName || row.objectKey }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.type')" width="96">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.assetType)">{{ assetTypeLabel(row.assetType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.usage')" width="150">
          <template #default="{ row }">
            <el-tag effect="plain">{{ assetUsageLabel(row.usageType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.scope')" width="110">
          <template #default="{ row }">
            <el-tag :type="scopeTag(row.assetScope)">{{ assetScopeLabel(row.assetScope) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.spec')" width="118">
          <template #default="{ row }">{{ assetSpec(row) }}</template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.size')" width="110">
          <template #default="{ row }">{{ formatSize(row.fileSize) }}</template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.references')" width="90">
          <template #default="{ row }">
            <el-tag :type="row.referenceCount > 0 ? 'warning' : 'info'">{{ row.referenceCount }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.checksum')" width="164">
          <template #default="{ row }">
            <span class="checksum">{{ shortChecksum(row.checksumSha256) }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('assetLibrary.fields.createdAt')" width="180">
          <template #default="{ row }">{{ formatBackendTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column fixed="right" :label="t('assetLibrary.fields.actions')" width="180">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-tooltip :content="t('assetLibrary.actions.detail')">
                <el-button link type="primary" :icon="View" @click="openDetail(row)" />
              </el-tooltip>
              <el-tooltip :content="t('assetLibrary.actions.copyLink')">
                <el-button link type="primary" :icon="CopyDocument" @click="copyAssetLink(row)" />
              </el-tooltip>
              <el-tooltip :content="t('assetLibrary.actions.edit')">
                <el-button link type="primary" :icon="Edit" @click="openEdit(row)" />
              </el-tooltip>
              <el-tooltip :content="t('assetLibrary.actions.delete')">
                <el-button link type="danger" :icon="Delete" @click="confirmDelete(row)" />
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-row">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next"
          :current-page="pagination.page"
          :page-size="pagination.pageSize"
          :page-sizes="[20, 50, 100]"
          :total="total"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </section>

    <el-dialog v-model="uploadVisible" :title="t('assetLibrary.dialogs.uploadTitle')" width="560px" @open="resetUploadForm">
      <el-form label-position="top" class="asset-form">
        <el-form-item :label="t('assetLibrary.fields.type')">
          <el-select v-model="uploadForm.assetType" clearable :placeholder="t('assetLibrary.placeholders.autoDetect')">
            <el-option v-for="item in assetTypes" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.usage')">
          <el-select v-model="uploadForm.usageType">
            <el-option v-for="item in usageOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="canUploadPlatformAssets" :label="t('assetLibrary.fields.scope')">
          <el-segmented
            v-model="uploadForm.assetScope"
            :options="uploadScopeOptions"
          />
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.displayName')">
          <el-input v-model="uploadForm.displayName" maxlength="160" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.altText')">
          <el-input v-model="uploadForm.altText" maxlength="255" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.tags')">
          <el-input v-model="uploadForm.tags" maxlength="500" :placeholder="t('assetLibrary.placeholders.tags')" />
        </el-form-item>
        <el-upload
          drag
          :http-request="handleUpload"
          :show-file-list="false"
          :disabled="uploading"
          accept=".jpg,.jpeg,.png,.webp,.avif,.gif,.svg,.ico,.mp3,.wav,.ogg,.m4a,.aac,.mp4,.webm,.mov,.m4v,.json,.lottie"
        >
          <el-icon class="upload-icon"><Upload /></el-icon>
          <div class="el-upload__text">{{ t("assetLibrary.upload.dropText") }}</div>
        </el-upload>
      </el-form>
    </el-dialog>

    <el-dialog v-model="editVisible" :title="t('assetLibrary.dialogs.editTitle')" width="520px">
      <el-form label-position="top" class="asset-form">
        <el-form-item :label="t('assetLibrary.fields.displayName')">
          <el-input v-model="editForm.displayName" maxlength="160" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.usage')">
          <el-select v-model="editForm.usageType">
            <el-option v-for="item in usageOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.altText')">
          <el-input v-model="editForm.altText" maxlength="255" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.tags')">
          <el-input v-model="editForm.tags" maxlength="500" />
        </el-form-item>
        <el-form-item :label="t('assetLibrary.fields.reason')">
          <el-input v-model="editForm.reason" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">{{ t("assetLibrary.actions.cancel") }}</el-button>
        <el-button type="primary" :loading="saving" @click="saveEdit">{{ t("assetLibrary.actions.save") }}</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" :title="t('assetLibrary.dialogs.detailTitle')" size="540px">
      <el-skeleton v-if="detailLoading" animated :rows="8" />
      <template v-else-if="detail">
        <section class="drawer-section">
          <div class="drawer-preview">
            <el-image
              v-if="isPreviewImage(detail.overview)"
              :src="assetUrl(detail.overview)"
              fit="contain"
              :preview-src-list="[assetUrl(detail.overview)]"
              preview-teleported
            />
            <audio v-else-if="detail.overview.assetType === 'AUDIO' && assetUrl(detail.overview)" :src="assetUrl(detail.overview)" controls />
            <video v-else-if="detail.overview.assetType === 'VIDEO' && assetUrl(detail.overview)" :src="assetUrl(detail.overview)" controls />
            <el-tag v-else size="large">{{ assetTypeLabel(detail.overview.assetType) }}</el-tag>
          </div>
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="t('assetLibrary.fields.name')">{{ detail.overview.displayName }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.scope')">{{ assetScopeLabel(detail.overview.assetScope) }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.usage')">{{ assetUsageLabel(detail.overview.usageType) }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.file')">{{ detail.overview.originalName }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.spec')">{{ assetSpec(detail.overview) }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.size')">{{ formatSize(detail.overview.fileSize) }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.checksum')">{{ detail.overview.checksumSha256 }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.createdBy')">{{ detail.overview.createdByName || "-" }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.createdAt')">{{ formatBackendTime(detail.overview.createdAt) }}</el-descriptions-item>
            <el-descriptions-item :label="t('assetLibrary.fields.link')">
              <el-link :href="assetUrl(detail.overview)" target="_blank" :icon="LinkIcon">{{ t("assetLibrary.actions.open") }}</el-link>
            </el-descriptions-item>
          </el-descriptions>
        </section>
        <section class="drawer-section">
          <h3>{{ t("assetLibrary.sections.referenceCheck") }}</h3>
          <el-table :data="detail.references || []" size="small" :empty-text="t('assetLibrary.labels.noReferences')">
            <el-table-column prop="targetType" :label="t('assetLibrary.fields.target')" min-width="180" />
          </el-table>
        </section>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.asset-library-page {
  min-height: 100%;
  padding: 24px;
  background: #f6f7fb;
  color: #1f2937;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.page-header h1 {
  margin: 2px 0 0;
  font-size: 24px;
  font-weight: 700;
}

.eyebrow {
  margin: 0;
  color: #64748b;
  font-size: 12px;
  text-transform: uppercase;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.toolbar,
.table-panel {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
}

.toolbar {
  padding: 16px 16px 2px;
  margin-bottom: 16px;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
}

.filter-input {
  width: 220px;
}

.filter-select {
  width: 168px;
}

.table-panel {
  padding: 14px 16px 16px;
}

.table-meta {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 10px;
  color: #64748b;
}

.table-meta strong {
  color: #111827;
  font-size: 20px;
}

.preview-box {
  display: grid;
  place-items: center;
  width: 80px;
  height: 56px;
  overflow: hidden;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: #f8fafc;
}

.preview-box :deep(.el-image),
.preview-box video {
  width: 100%;
  height: 100%;
}

.preview-box audio {
  width: 76px;
}

.name-cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
  color: #64748b;
  font-size: 12px;
}

.asset-name {
  overflow: hidden;
  color: #111827;
  font-size: 14px;
  font-weight: 650;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.checksum {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  color: #475569;
  font-size: 12px;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 14px;
}

.asset-form :deep(.el-select),
.asset-form :deep(.el-input) {
  width: 100%;
}

.upload-icon {
  margin-top: 8px;
  color: #409eff;
  font-size: 28px;
}

.drawer-section {
  margin-bottom: 18px;
}

.drawer-section h3 {
  margin: 0 0 10px;
  font-size: 15px;
}

.drawer-preview {
  display: grid;
  place-items: center;
  min-height: 180px;
  margin-bottom: 16px;
  overflow: hidden;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #f8fafc;
}

.drawer-preview :deep(.el-image),
.drawer-preview video {
  width: 100%;
  max-height: 240px;
}

.drawer-preview audio {
  width: 92%;
}

@media (max-width: 760px) {
  .asset-library-page {
    padding: 16px;
  }

  .page-header {
    flex-direction: column;
  }

  .header-actions,
  .filter-input,
  .filter-select {
    width: 100%;
  }
}
</style>
