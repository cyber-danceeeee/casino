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
  } from "@element-plus/icons-vue";
import { fetchGames, type GameCenterRecord } from "@/api/adminGames";
import {
  copyH5Navigation,
  createH5Navigation,
  deleteH5Navigation,
  listAssets,
  listH5Navigation,
  sortH5Navigation,
  updateH5Navigation,
  type AssetLibraryRecord,
  type H5NavigationAction,
  type H5NavigationBadge,
  type H5NavigationConfiguration,
  type H5NavigationPayload,
  type H5NavigationPosition,
  type H5NavigationRecord,
  type H5NavigationVisibility,
} from "@/api/adminOperations";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import { formatBackendTime } from "@/utils/timezone";

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const sorting = ref(false);
const rows = ref<H5NavigationRecord[]>([]);
const imageAssets = ref<AssetLibraryRecord[]>([]);
const lottieAssets = ref<AssetLibraryRecord[]>([]);
const soundAssets = ref<AssetLibraryRecord[]>([]);
const games = ref<GameCenterRecord[]>([]);
const overview = ref<BrandOverview | null>(null);
const editVisible = ref(false);
const previewVisible = ref(false);
const editingID = ref(0);
const draggedID = ref(0);
const activePosition = ref<H5NavigationPosition>("HOME_QUICK");
const filters = reactive({ keyword: "", });

const positions: H5NavigationPosition[] = ["HOME_QUICK", "BOTTOM_NAV", "PROFILE_MENU"];
const actionTypes: H5NavigationAction[] = [
  "OPEN_INTERNAL_PAGE",
  "OPEN_EXTERNAL_URL",
  "OPEN_DEPOSIT",
  "OPEN_WITHDRAW",
  "OPEN_SUPPORT",
  "OPEN_PROMOTION",
  "OPEN_GAME",
  "OPEN_DOWNLOAD",
];
const badges: H5NavigationBadge[] = ["NONE", "NEW", "HOT"];
const visibilityOptions: H5NavigationVisibility[] = ["ALL", "GUEST", "LOGGED_IN"];
const internalTargets = ["/", "/offers", "/share", "/profile", "/bonus-center", "/settings", "/about"];

const emptyForm = () => ({
  menuCode: "",
  labelZh: "",
  labelEn: "",
  iconAssetID: null as number | null,
  lottieAssetID: null as number | null,
  soundAssetID: null as number | null,
  displayPosition: activePosition.value,
  sortOrder: 10,
  badgeType: "NONE" as H5NavigationBadge,
  visibilityScope: "ALL" as H5NavigationVisibility,
  startAt: null as Date | null,
  endAt: null as Date | null,
  enabled: true,
  actionType: "OPEN_INTERNAL_PAGE" as H5NavigationAction,
  actionTarget: "/",
  providerCode: "",
  gameID: null as number | null,
  reason: "",
});
const form = reactive(emptyForm());

const currentRows = computed(() =>
  rows.value
    .filter((row) => configuration(row).displayPosition === activePosition.value && !configuration(row).deleted)
    .sort((a, b) => configuration(a).sortOrder - configuration(b).sortOrder || a.id - b.id),
);
const previewRows = computed(() => currentRows.value.filter((row) => configuration(row).enabled));
const targetMode = computed<"none" | "internal" | "external" | "promotion" | "game">(() => {
  if (form.actionType === "OPEN_INTERNAL_PAGE") return "internal";
  if (["OPEN_EXTERNAL_URL", "OPEN_DOWNLOAD"].includes(form.actionType)) return "external";
  if (form.actionType === "OPEN_PROMOTION") return "promotion";
  if (form.actionType === "OPEN_GAME") return "game";
  return "none";
});

function configuration(row: H5NavigationRecord): H5NavigationConfiguration {
  return row.configuration?.menuCode
    ? row.configuration
    : row.configuration;
}

function assetURL(asset?: AssetLibraryRecord) {
  return asset?.publicUrl || asset?.objectKey || "";
}

function iconFor(row: H5NavigationRecord) {
  const config = configuration(row);
  return assetURL(imageAssets.value.find((asset) => asset.id === config.iconAssetID)) || assetURL(row.iconAsset);
}

function scheduleText(value: string | null, emptyKey: string) {
  return value ? formatBackendTime(value) : t(emptyKey);
}

function gameLabel(game: GameCenterRecord) {
  const provider = String(game.providerCode || "").trim();
  const name = String(game.name || game.gameName || game.gameCode || game.providerGameCode || "-").trim();
  return provider ? `${provider} - ${name}` : name;
}

function gameLabelForConfig(config: H5NavigationConfiguration) {
  const game = games.value.find((item) => Number(item.id) === Number(config.gameID || 0));
  if (game) return gameLabel(game);
  const provider = String(config.providerCode || "").trim();
  const name = String(config.gameName || config.providerGameCode || "").trim();
  if (provider && name) return `${provider} - ${name}`;
  return name || "-";
}

async function load() {
  loading.value = true;
  try {
    const [entryResult, images, lotties, sounds, gameResult, brand] = await Promise.all([
      listH5Navigation({ page: 1, pageSize: 500, keyword: filters.keyword, }),
      listAssets({ page: 1, pageSize: 300, assetType: "IMAGE", usageType: "BUTTON_ICON" }),
      listAssets({ page: 1, pageSize: 300, assetType: "LOTTIE", usageType: "BUTTON_ICON" }),
      listAssets({ page: 1, pageSize: 300, assetType: "AUDIO", usageType: "SOUND" }),
      fetchGames({ page: 1, pageSize: 500, enabled: "true" }),
      systemAPI.brand.overview(),
    ]);
    rows.value = entryResult.list;
    imageAssets.value = images.list.filter((asset) => asset.status === "ACTIVE");
    lottieAssets.value = lotties.list.filter((asset) => asset.status === "ACTIVE");
    soundAssets.value = sounds.list.filter((asset) => asset.status === "ACTIVE");
    games.value = gameResult.list;
    overview.value = brand;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("h5EntryLibrary.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingID.value = 0;
  Object.assign(form, emptyForm(), {
    displayPosition: activePosition.value,
    sortOrder: (currentRows.value.length + 1) * 10,
  });
  editVisible.value = true;
}

function openEdit(row: H5NavigationRecord) {
	if (row.isFallback) return;
  const config = configuration(row);
  editingID.value = row.id;
  Object.assign(form, emptyForm(), config, {
    labelZh: config.labels?.["zh-CN"] || "",
    labelEn: config.labels?.["en-US"] || "",
    iconAssetID: config.iconAssetID || null,
    lottieAssetID: config.lottieAssetID || null,
    soundAssetID: config.soundAssetID || null,
    gameID: config.gameID || null,
    startAt: config.startAt ? new Date(config.startAt) : null,
    endAt: config.endAt ? new Date(config.endAt) : null,
    reason: "",
  });
  editVisible.value = true;
}

function changeAction() {
  form.actionTarget = form.actionType === "OPEN_INTERNAL_PAGE" ? "/" : "";
  form.providerCode = "";
  form.gameID = null;
}

function selectGame(gameID: number) {
  const game = games.value.find((item) => Number(item.id) === gameID);
  form.providerCode = String(game?.providerCode || "");
}

function payload(config = form): H5NavigationPayload {
	const actionNeedsTarget = ["OPEN_INTERNAL_PAGE", "OPEN_EXTERNAL_URL", "OPEN_PROMOTION", "OPEN_DOWNLOAD"].includes(config.actionType);
  return {
    menuCode: config.menuCode || undefined,
    labelZh: config.labelZh.trim(),
    labelEn: config.labelEn.trim(),
    labels: { "zh-CN": config.labelZh.trim(), "en-US": config.labelEn.trim() },
    iconAssetID: Number(config.iconAssetID || 0),
    lottieAssetID: Number(config.lottieAssetID || 0),
    soundAssetID: Number(config.soundAssetID || 0),
    displayPosition: config.displayPosition,
    sortOrder: Number(config.sortOrder || 0),
    badgeType: config.badgeType,
    visibilityScope: config.visibilityScope,
    startAt: config.startAt ? config.startAt.toISOString() : null,
    endAt: config.endAt ? config.endAt.toISOString() : null,
    enabled: config.enabled,
    actionType: config.actionType,
    actionTarget: actionNeedsTarget ? config.actionTarget.trim() : "",
    providerCode: config.actionType === "OPEN_GAME" ? config.providerCode : "",
    gameID: config.actionType === "OPEN_GAME" ? Number(config.gameID || 0) : 0,
    reason: config.reason.trim() || t("h5EntryLibrary.defaults.saveReason"),
  };
}

async function save() {
  if (!form.labelZh.trim() || !form.labelEn.trim() || (!form.iconAssetID && !form.lottieAssetID)) {
    ElMessage.warning(t("h5EntryLibrary.validation.required"));
    return;
  }
  if (["internal", "external", "promotion"].includes(targetMode.value) && !form.actionTarget.trim()) {
    ElMessage.warning(t("h5EntryLibrary.validation.targetRequired"));
    return;
  }
  if (targetMode.value === "external" && !/^https:\/\/[^\s]+$/i.test(form.actionTarget.trim())) {
    ElMessage.warning(t("h5EntryLibrary.validation.httpsRequired"));
    return;
  }
  if (targetMode.value === "game" && !form.gameID) {
    ElMessage.warning(t("h5EntryLibrary.validation.gameRequired"));
    return;
  }
  saving.value = true;
  try {
    if (editingID.value) await updateH5Navigation(editingID.value, payload());
    else await createH5Navigation(payload());
    editVisible.value = false;
    ElMessage.success(t("liveConfig.saved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("h5EntryLibrary.messages.saveFailed"));
  } finally {
    saving.value = false;
  }
}

async function toggle(row: H5NavigationRecord, enabled: boolean) {
  const config = configuration(row);
  try {
    await updateH5Navigation(row.id, {
      ...payload({
        ...emptyForm(),
        ...config,
        labelZh: config.labels?.["zh-CN"] || "",
        labelEn: config.labels?.["en-US"] || "",
        startAt: config.startAt ? new Date(config.startAt) : null,
        endAt: config.endAt ? new Date(config.endAt) : null,
        enabled,
        reason: t("h5EntryLibrary.defaults.statusReason"),
      }),
    });
    ElMessage.success(t("liveConfig.saved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("h5EntryLibrary.messages.saveFailed"));
  }
}

async function duplicate(row: H5NavigationRecord) {
	if (row.isFallback) return;
  try {
    await copyH5Navigation(row.id);
    ElMessage.success(t("h5EntryLibrary.messages.copied"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("h5EntryLibrary.messages.copyFailed"));
  }
}

async function remove(row: H5NavigationRecord) {
	if (row.isFallback) return;
  try {
    await ElMessageBox.confirm(t("h5EntryLibrary.dialogs.deleteMessage"), t("h5EntryLibrary.dialogs.deleteTitle"), {
      confirmButtonText: t("common.delete"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    });
    await deleteH5Navigation(row.id);
    ElMessage.success(t("liveConfig.saved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(error instanceof Error ? error.message : t("h5EntryLibrary.messages.deleteFailed"));
  }
}

function dragStart(row: H5NavigationRecord) {
	if (row.isFallback) return;
	draggedID.value = row.id;
}

async function dropOn(targetID: number) {
	const sortableRows = currentRows.value.filter((row) => !row.isFallback);
	const sourceIndex = sortableRows.findIndex((row) => row.id === draggedID.value);
	const targetIndex = sortableRows.findIndex((row) => row.id === targetID);
	draggedID.value = 0;
	if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return;
	const reordered = [...sortableRows];
  const [moved] = reordered.splice(sourceIndex, 1);
  reordered.splice(targetIndex, 0, moved);
  sorting.value = true;
  try {
    await sortH5Navigation(activePosition.value, reordered.map((row) => row.id), t("h5EntryLibrary.defaults.sortReason"));
    ElMessage.success(t("liveConfig.saved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("h5EntryLibrary.messages.sortFailed"));
  } finally {
    sorting.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div v-loading="loading" class="entry-page">
    <header class="page-header">
      <div>
        <p class="eyebrow">{{ t("h5EntryLibrary.eyebrow") }}</p>
        <h1>{{ t("h5EntryLibrary.title") }}</h1>
        <p class="subtitle">{{ t("h5EntryLibrary.subtitle") }}</p>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="load">{{ t("common.refresh") }}</el-button>
        <el-button :icon="Monitor" @click="previewVisible = true">{{ t("h5EntryLibrary.actions.preview") }}</el-button>

        <el-button type="success" :icon="Plus" @click="openCreate">{{ t("h5EntryLibrary.actions.create") }}</el-button>
      </div>
    </header>



    <section class="toolbar">
      <el-input v-model="filters.keyword" clearable :placeholder="t('h5EntryLibrary.placeholders.keyword')" @keyup.enter="load" />

      <el-button type="primary" @click="load">{{ t("common.search") }}</el-button>
    </section>

    <el-tabs v-model="activePosition" class="position-tabs">
      <el-tab-pane v-for="position in positions" :key="position" :name="position" :label="t(`h5EntryLibrary.positions.${position}`)" />
    </el-tabs>

    <section class="table-shell">
      <el-table :data="currentRows" row-key="id" v-loading="sorting">
		<el-table-column width="48"><template #default="{ row }"><el-icon v-if="!row.isFallback" class="drag-handle" draggable="true" @dragstart="dragStart(row)" @dragover.prevent @drop="dropOn(row.id)"><Rank /></el-icon></template></el-table-column>
        <el-table-column :label="t('h5EntryLibrary.fields.icon')" width="86">
          <template #default="{ row }"><div class="entry-icon"><img v-if="iconFor(row)" :src="iconFor(row)" alt="" /><span v-else>{{ configuration(row).labels?.['en-US']?.slice(0, 1) }}</span><i v-if="configuration(row).badgeType !== 'NONE'">{{ configuration(row).badgeType }}</i></div></template>
        </el-table-column>
        <el-table-column :label="t('h5EntryLibrary.fields.name')" min-width="170"><template #default="{ row }"><strong>{{ configuration(row).labels?.['zh-CN'] }}</strong><small>{{ configuration(row).labels?.['en-US'] }} · {{ configuration(row).menuCode }}</small></template></el-table-column>
        <el-table-column :label="t('h5EntryLibrary.fields.actionType')" min-width="170"><template #default="{ row }"><el-tag effect="plain">{{ t(`h5EntryLibrary.actionTypes.${configuration(row).actionType}`) }}</el-tag><small>{{ configuration(row).actionType === 'OPEN_GAME' ? gameLabelForConfig(configuration(row)) : configuration(row).actionTarget || '-' }}</small></template></el-table-column>
        <el-table-column :label="t('h5EntryLibrary.fields.visibility')" width="130"><template #default="{ row }">{{ t(`h5EntryLibrary.visibility.${configuration(row).visibilityScope}`) }}</template></el-table-column>
        <el-table-column :label="t('h5EntryLibrary.fields.schedule')" width="185"><template #default="{ row }"><small>{{ scheduleText(configuration(row).startAt, 'h5EntryLibrary.labels.immediate') }}</small><small>{{ scheduleText(configuration(row).endAt, 'h5EntryLibrary.labels.noExpiry') }}</small></template></el-table-column>

		<el-table-column :label="t('operations.fields.source')" width="130"><template #default="{ row }"><el-tag v-if="row.isFallback" type="info" effect="plain">{{ t('operations.values.platformDefault') }}</el-tag><span v-else>{{ t('operations.values.siteCustom') }}</span></template></el-table-column>
        <el-table-column :label="t('h5EntryLibrary.fields.enabled')" width="90"><template #default="{ row }"><el-switch :model-value="configuration(row).enabled" @change="toggle(row, Boolean($event))" /></template></el-table-column>
		<el-table-column :label="t('common.actions')" width="150" fixed="right"><template #default="{ row }"><template v-if="!row.isFallback"><el-button link type="primary" :icon="Edit" @click="openEdit(row)" /><el-button link :icon="CopyDocument" @click="duplicate(row)" /><el-button link type="danger" :icon="Delete" @click="remove(row)" /></template><span v-else>-</span></template></el-table-column>
        <template #empty>{{ t("h5EntryLibrary.messages.empty") }}</template>
      </el-table>
    </section>

    <el-dialog v-model="editVisible" :title="editingID ? t('h5EntryLibrary.dialogs.editTitle') : t('h5EntryLibrary.dialogs.createTitle')" width="780px" destroy-on-close>
      <el-form label-position="top">
        <el-row :gutter="16">
          <el-col :span="12"><el-form-item :label="t('h5EntryLibrary.fields.nameZh')" required><el-input v-model="form.labelZh" maxlength="64" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('h5EntryLibrary.fields.nameEn')" required><el-input v-model="form.labelEn" maxlength="64" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.icon')"><el-select v-model="form.iconAssetID" clearable filterable :placeholder="t('h5EntryLibrary.placeholders.unconfigured')"><el-option v-for="asset in imageAssets" :key="asset.id" :label="asset.displayName || asset.originalName" :value="asset.id" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.lottie')"><el-select v-model="form.lottieAssetID" clearable filterable :placeholder="t('h5EntryLibrary.placeholders.unconfigured')"><el-option v-for="asset in lottieAssets" :key="asset.id" :label="asset.displayName || asset.originalName" :value="asset.id" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.sound')"><el-select v-model="form.soundAssetID" clearable filterable :placeholder="t('h5EntryLibrary.placeholders.unconfigured')"><el-option v-for="asset in soundAssets" :key="asset.id" :label="asset.displayName || asset.originalName" :value="asset.id" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.position')" required><el-select v-model="form.displayPosition"><el-option v-for="position in positions" :key="position" :label="t(`h5EntryLibrary.positions.${position}`)" :value="position" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.badge')"><el-select v-model="form.badgeType"><el-option v-for="badge in badges" :key="badge" :label="t(`h5EntryLibrary.badges.${badge}`)" :value="badge" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.visibility')"><el-select v-model="form.visibilityScope"><el-option v-for="value in visibilityOptions" :key="value" :label="t(`h5EntryLibrary.visibility.${value}`)" :value="value" /></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('h5EntryLibrary.fields.actionType')" required><el-select v-model="form.actionType" @change="changeAction"><el-option v-for="action in actionTypes" :key="action" :label="t(`h5EntryLibrary.actionTypes.${action}`)" :value="action" /></el-select></el-form-item></el-col>
          <el-col v-if="targetMode === 'internal'" :span="12"><el-form-item :label="t('h5EntryLibrary.fields.actionTarget')" required><el-select v-model="form.actionTarget" filterable><el-option v-for="target in internalTargets" :key="target" :label="target" :value="target" /></el-select></el-form-item></el-col>
          <el-col v-else-if="targetMode === 'external' || targetMode === 'promotion'" :span="12"><el-form-item :label="t('h5EntryLibrary.fields.actionTarget')" required><el-input v-model="form.actionTarget" :placeholder="t(`h5EntryLibrary.targetPlaceholder.${form.actionType}`)" /></el-form-item></el-col>
          <el-col v-else-if="targetMode === 'game'" :span="12"><el-form-item :label="t('h5EntryLibrary.fields.game')" required><el-select v-model="form.gameID" filterable :placeholder="t('h5EntryLibrary.placeholders.selectGame')" @change="selectGame"><el-option v-for="game in games" :key="Number(game.id)" :label="gameLabel(game)" :value="Number(game.id)" /></el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.sortOrder')"><el-input-number v-model="form.sortOrder" :min="0" :max="9999" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.startAt')"><el-date-picker v-model="form.startAt" type="datetime" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item :label="t('h5EntryLibrary.fields.endAt')"><el-date-picker v-model="form.endAt" type="datetime" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('h5EntryLibrary.fields.enabled')"><el-switch v-model="form.enabled" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="t('h5EntryLibrary.fields.reason')"><el-input v-model="form.reason" maxlength="500" /></el-form-item></el-col>
        </el-row>
      </el-form>
      <template #footer><el-button @click="editVisible = false">{{ t("common.cancel") }}</el-button><el-button type="primary" :loading="saving" @click="save">{{ t("common.save") }}</el-button></template>
    </el-dialog>

    <el-dialog v-model="previewVisible" :title="t('h5EntryLibrary.dialogs.previewTitle')" width="500px">
      <div class="entry-preview" :class="`is-${activePosition.toLowerCase()}`">
        <button v-for="row in previewRows" :key="row.id" type="button">
          <span class="entry-icon"><img v-if="iconFor(row)" :src="iconFor(row)" alt="" /><strong v-else>{{ configuration(row).labels?.['en-US']?.slice(0, 1) }}</strong><i v-if="configuration(row).badgeType !== 'NONE'">{{ configuration(row).badgeType }}</i></span>
          <small>{{ configuration(row).labels?.['zh-CN'] }}</small>
        </button>
        <el-empty v-if="!previewRows.length" :description="t('h5EntryLibrary.messages.emptyPreview')" />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.entry-page { display: grid; gap: 16px; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; }
.page-header h1 { margin: 2px 0 6px; font-size: 24px; letter-spacing: 0; }
.eyebrow { margin: 0; color: #2563eb; font-size: 12px; font-weight: 700; text-transform: uppercase; }
.subtitle { margin: 0; color: #64748b; }
.header-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
.toolbar { display: grid; grid-template-columns: minmax(220px, 1fr) 190px auto; gap: 10px; max-width: 760px; }
.position-tabs { margin-bottom: -16px; }
.table-shell { overflow: hidden; border: 1px solid #e2e8f0; border-radius: 6px; background: #fff; }
.drag-handle { cursor: grab; color: #94a3b8; }
.entry-icon { position: relative; display: grid; place-items: center; width: 44px; height: 44px; overflow: visible; border: 1px solid #dbe3ee; border-radius: 8px; background: #f8fafc; color: #334155; font-weight: 700; }
.entry-icon img { width: 100%; height: 100%; object-fit: contain; border-radius: 7px; }
.entry-icon i { position: absolute; top: -6px; right: -12px; padding: 1px 4px; border-radius: 4px; background: #ef4444; color: #fff; font-size: 9px; font-style: normal; line-height: 14px; }
.el-table strong, .el-table small { display: block; }
.el-table small { margin-top: 3px; color: #64748b; }
.entry-preview { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; min-height: 180px; padding: 20px; border: 1px solid #dbe3ee; border-radius: 8px; background: #f8fafc; }
.entry-preview.is-bottom_nav { grid-template-columns: repeat(5, minmax(0, 1fr)); align-items: end; }
.entry-preview button { display: grid; justify-items: center; align-content: start; gap: 8px; min-width: 0; border: 0; background: transparent; color: #334155; }
.entry-preview small { max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
@media (max-width: 900px) { .page-header { flex-direction: column; } .header-actions { justify-content: flex-start; } .toolbar { grid-template-columns: 1fr; } }
@media (max-width: 700px) { :deep(.el-dialog) { width: calc(100vw - 24px) !important; } :deep(.el-dialog .el-col) { max-width: 100%; flex: 0 0 100%; } }
</style>
