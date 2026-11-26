<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  CopyDocument,
  Delete,
  Edit,
  Monitor,
  Plus,
  Rank,
  Refresh,
  UploadFilled,
} from "@element-plus/icons-vue";
import H5PreviewFrame from "@/components/brand/H5PreviewFrame.vue";
import { fetchGames, type GameCenterRecord } from "@/api/adminGames";
import {
  copySiteBanner,
  createSiteBanner,
  deleteSiteBanner,
  listAssets,
  listOperations,
  listSiteBanners,
  sortSiteBanners,
	setOperationDefaultVisibility,
  updateSiteBanner,
  type AssetLibraryRecord,
  type BannerActionType,
  type BannerVisibility,
  type SiteBannerPayload,
  type SiteBannerRecord,
} from "@/api/adminOperations";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import { formatBackendTime } from "@/utils/timezone";

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const publishing = ref(false);
const sorting = ref(false);
const rows = ref<SiteBannerRecord[]>([]);
const assets = ref<AssetLibraryRecord[]>([]);
const games = ref<GameCenterRecord[]>([]);
const promotions = ref<Array<{ label: string; value: string }>>([]);
const overview = ref<BrandOverview | null>(null);
const editVisible = ref(false);
const previewVisible = ref(false);
const editingID = ref(0);
const draggedID = ref(0);
const filters = reactive({ keyword: "", publishStatus: "" });

const emptyForm = () => ({
  name: "",
  title: "",
  subtitle: "",
  desktopAssetID: 0,
  mobileAssetID: 0,
  desktopImageUrl: "",
  mobileImageUrl: "",
  actionType: "NONE" as BannerActionType,
  actionTarget: "",
  sortOrder: 0,
  languageCode: "ALL",
  countryCode: "ALL",
  currencyCode: "ALL",
  visibilityScope: "ALL" as BannerVisibility,
  startAt: null as Date | null,
  endAt: null as Date | null,
  enabled: true,
  reason: "",
});
const form = reactive(emptyForm());

const actionTypes: BannerActionType[] = [
  "NONE",
  "OPEN_INTERNAL_PAGE",
  "OPEN_EXTERNAL_URL",
  "OPEN_GAME",
  "OPEN_PROMOTION",
  "OPEN_DEPOSIT",
  "OPEN_SUPPORT",
];
const visibilityOptions: BannerVisibility[] = ["ALL", "LOGGED_OUT", "LOGGED_IN"];
const targetRequired = computed(() =>
  !["NONE", "OPEN_DEPOSIT", "OPEN_SUPPORT"].includes(form.actionType),
);
const internalTargets = ["/", "/offers", "/bonus-center", "/profile", "/support"];
const selectedAssetMissing = computed(
  () => Boolean(form.desktopAssetID) && !assetByID(form.desktopAssetID),
);
const conflictNames = computed(() => {
  if (!form.enabled) return [];
  const start = form.startAt?.getTime() ?? Number.NEGATIVE_INFINITY;
  const end = form.endAt?.getTime() ?? Number.POSITIVE_INFINITY;
  return rows.value
    .filter((row) => !row.isFallback && row.id !== editingID.value)
    .filter((row) => {
      const config = draft(row);
      if (!config.enabled || config.deleted || config.visibilityScope !== form.visibilityScope) return false;
      const otherStart = config.startAt ? new Date(config.startAt).getTime() : Number.NEGATIVE_INFINITY;
      const otherEnd = config.endAt ? new Date(config.endAt).getTime() : Number.POSITIVE_INFINITY;
      return start <= otherEnd && otherStart <= end;
    })
    .map((row) => draft(row).name)
    .filter(Boolean);
});
const actionTargetOptions = computed(() => {
  if (form.actionType === "OPEN_INTERNAL_PAGE")
    return internalTargets.map((value) => ({ label: value, value }));
  if (form.actionType === "OPEN_GAME")
    return games.value
      .map((game) => ({
        label: `${String(game.name || game.gameName || game.gameCode || "-")} · ${String(game.providerCode || "")}`,
        value: String(game.gameCode || game.providerGameCode || ""),
      }))
      .filter((item) => item.value);
  if (form.actionType === "OPEN_PROMOTION") return promotions.value;
  return [];
});

function draft(row: SiteBannerRecord) {
  return row.draftConfiguration || row.publishedConfiguration;
}

function assetURL(asset?: AssetLibraryRecord) {
  return asset?.publicUrl || asset?.objectKey || "";
}

function assetByID(id: number) {
  return assets.value.find((asset) => asset.id === id);
}

function draftImage(row: SiteBannerRecord, mobile = false) {
	if (row.isFallback) return row.fallbackImageURL || "";
  const config = draft(row);
  const selected = assetByID(mobile ? config.mobileAssetID : config.desktopAssetID);
  if (selected) return assetURL(selected);
  return mobile
    ? assetURL(row.mobileAsset) || config.mobileImageUrl || assetURL(row.desktopAsset) || config.desktopImageUrl || ""
    : assetURL(row.desktopAsset) || config.desktopImageUrl || "";
}

function selectedBannerImage(mobile = false) {
  const asset = assetByID(mobile ? Number(form.mobileAssetID || 0) : Number(form.desktopAssetID || 0));
  if (asset) return assetURL(asset);
  return mobile ? String(form.mobileImageUrl || form.desktopImageUrl || "") : String(form.desktopImageUrl || "");
}

function actionTargetLabel(row: SiteBannerRecord) {
  const config = draft(row);
  if (!config.actionTarget) return "-";
  if (config.actionType === "OPEN_GAME") {
    const game = games.value.find((item) =>
      [item.gameCode, item.providerGameCode].some((value) => String(value || "") === config.actionTarget),
    );
    return game ? `${String(game.name || game.gameName || config.actionTarget)} · ${String(game.providerCode || "")}` : config.actionTarget;
  }
  if (config.actionType === "OPEN_PROMOTION")
    return promotions.value.find((item) => item.value === config.actionTarget)?.label || config.actionTarget;
  return config.actionTarget;
}

async function load() {
  loading.value = true;
  try {
    const [bannerResult, assetResult, brand, gameResult, promotionResult] = await Promise.all([
      listSiteBanners({ page: 1, pageSize: 200, ...filters }),
      listAssets({ page: 1, pageSize: 200, assetType: "IMAGE", usageType: "BANNER" }),
      systemAPI.brand.overview(),
      fetchGames({ page: 1, pageSize: 500, enabled: "true" }).catch(() => ({ list: [], total: 0, page: 1, pageSize: 500 })),
      listOperations("promotions", { page: 1, pageSize: 500 }).catch(() => ({ list: [], total: 0, page: 1, pageSize: 500 })),
    ]);
    rows.value = bannerResult.list;
    assets.value = assetResult.list.filter((asset) => asset.status === "ACTIVE");
    overview.value = brand;
    games.value = gameResult.list || [];
    promotions.value = (promotionResult.list || [])
      .map((item) => {
        const record = item as Record<string, unknown>;
        const value = String(record.promotionCode || record.campaignCode || record.code || record.slug || "");
        const label = String(record.title || record.name || record.promotionName || value);
        return { value, label: value && value !== label ? `${label} · ${value}` : label };
      })
      .filter((item) => item.value);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingID.value = 0;
  Object.assign(form, emptyForm(), { sortOrder: rows.value.length });
  editVisible.value = true;
}

function openEdit(row: SiteBannerRecord) {
	if (row.isFallback) return;
  const config = draft(row);
  editingID.value = row.id;
  Object.assign(form, emptyForm(), config, {
    startAt: config.startAt ? new Date(config.startAt) : null,
    endAt: config.endAt ? new Date(config.endAt) : null,
    reason: "",
  });
  editVisible.value = true;
}

function payload(): SiteBannerPayload {
  return {
    name: form.name.trim(),
    title: form.title.trim(),
    subtitle: form.subtitle.trim(),
    desktopAssetID: form.desktopAssetID,
    mobileAssetID: form.mobileAssetID,
    desktopImageUrl: String(form.desktopImageUrl || ""),
    mobileImageUrl: String(form.mobileImageUrl || ""),
    actionType: form.actionType,
    actionTarget: targetRequired.value ? form.actionTarget.trim() : "",
    sortOrder: form.sortOrder,
    languageCode: form.languageCode.trim() || "ALL",
    countryCode: form.countryCode.trim().toUpperCase() || "ALL",
    currencyCode: form.currencyCode.trim().toUpperCase() || "ALL",
    visibilityScope: form.visibilityScope,
    startAt: form.startAt ? form.startAt.toISOString() : null,
    endAt: form.endAt ? form.endAt.toISOString() : null,
    enabled: form.enabled,
    reason: form.reason.trim() || t("bannerLibrary.defaults.saveReason"),
  };
}

async function save() {
  if (!form.name.trim() || !form.title.trim() || (!form.desktopAssetID && !form.desktopImageUrl)) {
    ElMessage.warning(t("bannerLibrary.validation.required"));
    return;
  }
  if (targetRequired.value && !form.actionTarget.trim()) {
    ElMessage.warning(t("bannerLibrary.validation.targetRequired"));
    return;
  }
  saving.value = true;
  try {
    if (editingID.value) await updateSiteBanner(editingID.value, payload());
    else await createSiteBanner(payload());
    editVisible.value = false;
    ElMessage.success(t("bannerLibrary.messages.draftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.saveFailed"));
  } finally {
    saving.value = false;
  }
}

async function toggle(row: SiteBannerRecord, enabled: boolean) {
	if (row.isFallback) {
		try {
			await setOperationDefaultVisibility("banners", row.fallbackSourceID || row.id, enabled);
			row.visible = enabled;
			row.draftConfiguration.enabled = enabled;
			ElMessage.success(t("operations.messages.defaultVisibilitySaved"));
		} catch (error) {
			ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.saveFailed"));
		}
		return;
	}
  const config = draft(row);
  try {
    await updateSiteBanner(row.id, { ...config, enabled, reason: t("bannerLibrary.defaults.statusReason") });
    ElMessage.success(t("bannerLibrary.messages.statusDraftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.saveFailed"));
  }
}

async function duplicate(row: SiteBannerRecord) {
	if (row.isFallback) return;
  try {
    await copySiteBanner(row.id);
    ElMessage.success(t("bannerLibrary.messages.copied"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.copyFailed"));
  }
}

async function remove(row: SiteBannerRecord) {
	if (row.isFallback) return;
  try {
    await ElMessageBox.confirm(t("bannerLibrary.dialogs.deleteMessage"), t("bannerLibrary.dialogs.deleteTitle"), {
      confirmButtonText: t("bannerLibrary.actions.deleteDraft"),
      cancelButtonText: t("bannerLibrary.actions.cancel"),
      type: "warning",
    });
    await deleteSiteBanner(row.id);
    ElMessage.success(t("bannerLibrary.messages.deleteDraftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.deleteFailed"));
  }
}

function dragStart(row: SiteBannerRecord) {
	if (row.isFallback) return;
	draggedID.value = row.id;
}

async function dropOn(targetID: number) {
	const sortableRows = rows.value.filter((row) => !row.isFallback);
	const sourceIndex = sortableRows.findIndex((row) => row.id === draggedID.value);
	const targetIndex = sortableRows.findIndex((row) => row.id === targetID);
  draggedID.value = 0;
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return;
	const reordered = [...sortableRows];
  const [moved] = reordered.splice(sourceIndex, 1);
  reordered.splice(targetIndex, 0, moved);
  sorting.value = true;
  try {
    await sortSiteBanners(reordered.map((row) => row.id), t("bannerLibrary.defaults.sortReason"));
    ElMessage.success(t("bannerLibrary.messages.sortDraftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.sortFailed"));
    await load();
  } finally {
    sorting.value = false;
  }
}

async function publish() {
  if (!overview.value?.siteID) return;
  try {
    const result = await ElMessageBox.prompt(t("bannerLibrary.dialogs.publishMessage"), t("bannerLibrary.dialogs.publishTitle"), {
      confirmButtonText: t("bannerLibrary.actions.publish"),
      cancelButtonText: t("bannerLibrary.actions.cancel"),
      inputValue: t("bannerLibrary.defaults.publishReason"),
      inputValidator: (value) => Boolean(value.trim()) || t("bannerLibrary.validation.reasonRequired"),
    });
    publishing.value = true;
    const response = await systemAPI.sites.publish(overview.value.siteID, result.value.trim());
    ElMessage.success(t("bannerLibrary.messages.published", { count: Number(response.bannerDraftsPublished || 0) }));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(error instanceof Error ? error.message : t("bannerLibrary.messages.publishFailed"));
  } finally {
    publishing.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div v-loading="loading" class="banner-page">
    <header class="page-header">
      <div>
        <p class="eyebrow">{{ t("bannerLibrary.eyebrow") }}</p>
        <h1>{{ t("bannerLibrary.title") }}</h1>
        <p class="subtitle">{{ t("bannerLibrary.subtitle") }}</p>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="load">{{ t("bannerLibrary.actions.refresh") }}</el-button>
        <el-button :icon="Monitor" @click="previewVisible = true">{{ t("bannerLibrary.actions.preview") }}</el-button>
        <el-button type="primary" :icon="UploadFilled" :loading="publishing" @click="publish">
          {{ t("bannerLibrary.actions.publish") }}
        </el-button>
        <el-button type="success" :icon="Plus" @click="openCreate">{{ t("bannerLibrary.actions.create") }}</el-button>
      </div>
    </header>

    <el-alert type="warning" :title="t('bannerLibrary.draftNotice')" :closable="false" show-icon />

    <section class="toolbar">
      <el-input v-model="filters.keyword" clearable :placeholder="t('bannerLibrary.placeholders.keyword')" @keyup.enter="load" />
      <el-select v-model="filters.publishStatus" clearable :placeholder="t('bannerLibrary.fields.publishStatus')">
        <el-option :label="t('bannerLibrary.status.DRAFT')" value="DRAFT" />
        <el-option :label="t('bannerLibrary.status.PUBLISHED')" value="PUBLISHED" />
      </el-select>
      <el-button type="primary" @click="load">{{ t("bannerLibrary.actions.search") }}</el-button>
    </section>

    <section class="table-shell">
	  <el-table :data="rows" row-key="rowKey" v-loading="sorting">
        <el-table-column width="48">
          <template #default="{ row }">
			<el-icon v-if="!row.isFallback" class="drag-handle" draggable="true" @dragstart="dragStart(row)" @dragover.prevent @drop="dropOn(row.id)"><Rank /></el-icon>
          </template>
        </el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.preview')" width="150">
          <template #default="{ row }"><el-image class="banner-thumb" :src="draftImage(row)" fit="cover" /></template>
        </el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.name')" min-width="170">
          <template #default="{ row }"><strong>{{ draft(row).name }}</strong><small>{{ draft(row).title }}</small></template>
        </el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.target')" min-width="165">
          <template #default="{ row }"><el-tag effect="plain">{{ t(`bannerLibrary.actionsType.${draft(row).actionType}`) }}</el-tag><small v-if="draft(row).actionTarget">{{ actionTargetLabel(row) }}</small></template>
        </el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.visibility')" width="130">
          <template #default="{ row }">{{ t(`bannerLibrary.visibility.${draft(row).visibilityScope}`) }}</template>
        </el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.schedule')" min-width="190">
          <template #default="{ row }"><span>{{ formatBackendTime(draft(row).startAt || "") }}</span><small>{{ formatBackendTime(draft(row).endAt || "") }}</small></template>
        </el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.publishStatus')" width="150">
          <template #default="{ row }"><el-tag :type="row.publishStatus === 'DRAFT' ? 'warning' : 'success'">{{ t(`bannerLibrary.status.${row.publishStatus}`) }}</el-tag></template>
        </el-table-column>
		<el-table-column :label="t('operations.fields.source')" width="130"><template #default="{ row }"><el-tag v-if="row.isFallback" type="info" effect="plain">{{ t('operations.values.platformDefault') }}</el-tag><span v-else>{{ t('operations.values.siteCustom') }}</span></template></el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.enabled')" width="100">
          <template #default="{ row }"><el-switch :model-value="draft(row).enabled" @change="toggle(row, Boolean($event))" /></template>
        </el-table-column>
        <el-table-column :label="t('bannerLibrary.fields.actions')" width="150" fixed="right">
          <template #default="{ row }">
			<template v-if="!row.isFallback"><el-tooltip :content="t('bannerLibrary.actions.edit')"><el-button circle :icon="Edit" @click="openEdit(row)" /></el-tooltip>
			<el-tooltip :content="t('bannerLibrary.actions.copy')"><el-button circle :icon="CopyDocument" @click="duplicate(row)" /></el-tooltip>
			<el-tooltip :content="t('bannerLibrary.actions.deleteDraft')"><el-button circle type="danger" plain :icon="Delete" @click="remove(row)" /></el-tooltip></template><span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>
    </section>

    <el-dialog v-model="editVisible" :title="editingID ? t('bannerLibrary.dialogs.editTitle') : t('bannerLibrary.dialogs.createTitle')" width="760px" destroy-on-close>
      <el-form label-position="top" class="banner-form">
        <el-alert v-if="selectedAssetMissing" type="error" show-icon :closable="false" :title="t('bannerLibrary.messages.missingAsset')" />
        <el-alert v-if="conflictNames.length" type="warning" show-icon :closable="false" :title="t('bannerLibrary.messages.scheduleConflict', { names: conflictNames.join(', ') })" />
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.name')" required><el-input v-model="form.name" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.title')" required><el-input v-model="form.title" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item :label="t('bannerLibrary.fields.subtitle')"><el-input v-model="form.subtitle" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.desktopAsset')" required><el-select v-model="form.desktopAssetID" clearable filterable><el-option v-for="asset in assets" :key="asset.id" :label="asset.displayName || asset.originalName" :value="asset.id"><span>{{ asset.displayName || asset.originalName }}</span><small class="asset-size">{{ asset.width }} × {{ asset.height }}</small></el-option></el-select><el-image v-if="selectedBannerImage()" class="banner-form-preview" :src="selectedBannerImage()" fit="contain" :preview-src-list="[selectedBannerImage()]" preview-teleported /><small v-if="!form.desktopAssetID && form.desktopImageUrl" class="asset-hint">{{ t('bannerLibrary.messages.legacyImage') }}</small></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.mobileAsset')"><el-select v-model="form.mobileAssetID" clearable filterable><el-option v-for="asset in assets" :key="asset.id" :label="asset.displayName || asset.originalName" :value="asset.id" /></el-select><el-image v-if="selectedBannerImage(true)" class="banner-form-preview" :src="selectedBannerImage(true)" fit="contain" :preview-src-list="[selectedBannerImage(true)]" preview-teleported /><small v-if="!form.mobileAssetID && form.mobileImageUrl" class="asset-hint">{{ t('bannerLibrary.messages.legacyImage') }}</small></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.actionType')"><el-select v-model="form.actionType"><el-option v-for="item in actionTypes" :key="item" :label="t(`bannerLibrary.actionsType.${item}`)" :value="item" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item v-if="targetRequired" :label="t('bannerLibrary.fields.actionTarget')" required><el-input v-if="form.actionType==='OPEN_EXTERNAL_URL'" v-model="form.actionTarget" :placeholder="t(`bannerLibrary.targetPlaceholder.${form.actionType}`)" /><el-select v-else v-model="form.actionTarget" clearable filterable><el-option v-for="item in actionTargetOptions" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.visibility')"><el-select v-model="form.visibilityScope"><el-option v-for="item in visibilityOptions" :key="item" :label="t(`bannerLibrary.visibility.${item}`)" :value="item" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.sortOrder')"><el-input-number v-model="form.sortOrder" :min="0" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.startAt')"><el-date-picker v-model="form.startAt" type="datetime" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.endAt')"><el-date-picker v-model="form.endAt" type="datetime" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('bannerLibrary.fields.enabled')"><el-switch v-model="form.enabled" /></el-form-item></el-col>
          <el-col :span="24"><el-form-item :label="t('bannerLibrary.fields.reason')"><el-input v-model="form.reason" type="textarea" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer><el-button @click="editVisible = false">{{ t("bannerLibrary.actions.cancel") }}</el-button><el-button type="primary" :loading="saving" @click="save">{{ t("bannerLibrary.actions.saveDraft") }}</el-button></template>
    </el-dialog>

    <el-dialog v-model="previewVisible" :title="t('bannerLibrary.dialogs.previewTitle')" width="min(560px, 96vw)">
      <el-alert class="preview-note" type="info" show-icon :closable="false" :title="t('bannerLibrary.messages.realH5Preview')" />
      <H5PreviewFrame :h5-domain="overview?.h5Domain" :site-i-d="overview?.siteID" />
    </el-dialog>
  </div>
</template>

<style scoped>
.banner-page { padding: 24px; color: #18212f; }
.page-header { display: flex; justify-content: space-between; gap: 24px; align-items: flex-start; margin-bottom: 18px; }
.eyebrow { margin: 0 0 5px; color: #b42318; font-size: 12px; font-weight: 700; text-transform: uppercase; }
h1 { margin: 0; font-size: 28px; letter-spacing: 0; }
.subtitle { margin: 7px 0 0; color: #667085; }
.header-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
.toolbar { display: grid; grid-template-columns: minmax(220px, 1fr) 190px auto; gap: 10px; margin: 16px 0; }
.table-shell { border: 1px solid #e4e7ec; border-radius: 6px; overflow: hidden; background: #fff; }
.drag-handle { cursor: grab; color: #667085; font-size: 18px; }
.banner-thumb { width: 118px; height: 54px; border-radius: 4px; background: #f2f4f7; }
.banner-form-preview { display: block; width: 100%; height: 96px; margin-top: 8px; border: 1px solid var(--admin-border); border-radius: 4px; background: #f8fafc; }
.asset-hint { display: block; margin-top: 6px; color: #667085; }
strong, small { display: block; }
small { margin-top: 4px; color: #667085; }
.banner-form :deep(.el-select), .banner-form :deep(.el-date-editor), .banner-form :deep(.el-input-number) { width: 100%; }
.asset-size { float: right; margin-left: 18px; }
.banner-form>.el-alert { margin-bottom: 14px; }
.preview-note { margin-bottom: 14px; }
@media (max-width: 760px) {
  .banner-page { padding: 16px; }
  .page-header { flex-direction: column; }
  .header-actions { justify-content: flex-start; }
  .toolbar { grid-template-columns: 1fr; }
}
</style>
