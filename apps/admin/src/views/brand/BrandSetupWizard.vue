<script setup lang="ts">
import { onMounted, reactive, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Monitor, UploadFilled } from "@element-plus/icons-vue";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";
import H5PreviewFrame from "@/components/brand/H5PreviewFrame.vue";
import {
  fetchGames,
  fetchLobbySections,
  updateGame,
  updateLobbySection,
  type GameCenterRecord,
  type LobbySectionInput,
} from "@/api/adminGames";
import { useI18n } from "vue-i18n";

const props = defineProps<{ overview: BrandOverview }>();
const emit = defineEmits<{ completed: [] }>();
const { t } = useI18n();

const active = ref(0);
const saving = ref(false);
const publishing = ref(false);
const previewVisible = ref(false);
const previewLoading = ref(false);
const gamesLoading = ref(false);
const lobbyLoading = ref(false);
const games = ref<GameCenterRecord[]>([]);
const lobbySections = ref<GameCenterRecord[]>([]);

const form = reactive({
  siteName: "",
  appName: "",
  logo: "",
  primaryColor: "#E53935",
  secondaryColor: "#FFB300",
  buttonColor: "#E53935",
  backgroundColor: "#070A12",
  textColor: "#FFFFFF",
  countryCode: "",
  currencyCode: "",
  defaultLanguage: "",
  timezone: "",
  customerService: "",
  email: "",
});

function objectValue(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}

function assignFromOverview(row: BrandOverview) {
  const draft = objectValue(row.draftConfig);
  const brand = objectValue(draft.brandConfig);
  const theme = objectValue(draft.themeConfig);
  Object.assign(form, {
    siteName: String(draft.siteName || row.brandName || ""),
    appName: String(brand.appName || row.brandName || ""),
    logo: String(brand.logo || ""),
    primaryColor: String(theme.primary || "#E53935"),
    secondaryColor: String(theme.secondary || "#FFB300"),
    buttonColor: String(theme.button || theme.primary || "#E53935"),
    backgroundColor: String(theme.background || "#070A12"),
    textColor: String(theme.text || "#FFFFFF"),
    countryCode: String(draft.countryCode || row.countryCode || ""),
    currencyCode: String(draft.currencyCode || row.currencyCode || ""),
    defaultLanguage: String(draft.defaultLanguage || row.defaultLanguage || "en"),
    timezone: String(draft.timezone || row.timezone || "Asia/Kuala_Lumpur"),
    customerService: String(brand.customerService || ""),
    email: String(brand.email || ""),
  });
}

function brandPayload() {
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
      email: form.email.trim(),
    },
    themeConfig: {
      primary: form.primaryColor,
      secondary: form.secondaryColor,
      button: form.buttonColor,
      background: form.backgroundColor,
      text: form.textColor,
    },
    featureConfig: objectValue(props.overview.draftConfig?.featureConfig),
    maintenanceConfig: objectValue(props.overview.draftConfig?.maintenanceConfig),
    seoConfig: objectValue(props.overview.draftConfig?.seoConfig),
    reason: "initial brand setup draft",
  };
}

async function saveBrandDraft() {
  saving.value = true;
  try {
    await systemAPI.sites.update(props.overview.siteID, brandPayload());
    ElMessage.success(t("brandPortal.messages.draftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.saveFailed"));
  } finally {
    saving.value = false;
  }
}

function hasExplicitValue(value: unknown) {
  return value !== null && value !== undefined && value !== "";
}

function gameBool(row: GameCenterRecord, key: string) {
  const draft = row[`${key}Draft`];
  return Boolean(hasExplicitValue(draft) ? draft : row[key]);
}

function gameNumber(row: GameCenterRecord, key: string) {
  const draft = row[`${key}Draft`];
  return Number(hasExplicitValue(draft) ? draft : row[key] || 0);
}

async function loadGames() {
  gamesLoading.value = true;
  try {
    const result = await fetchGames({ page: 1, pageSize: 50 });
    games.value = result.list || [];
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.gamesLoadFailed"));
  } finally {
    gamesLoading.value = false;
  }
}

async function saveGameEnabled(row: GameCenterRecord, enabled: boolean) {
  try {
    await updateGame(row.id, {
      name: String(row.name || ""),
      localThumbnailUrl: String(row.localThumbnailUrl || row.thumbnailUrl || ""),
      platformEnabled: enabled,
      isHot: gameBool(row, "isHot"),
      isFeatured: gameBool(row, "isFeatured"),
      isNew: gameBool(row, "isNew"),
      sortOrder: gameNumber(row, "sortOrder"),
      supportedCountries: String(row.supportedCountries || ""),
      supportedCurrencies: String(row.supportedCurrencies || ""),
      supportedDevices: String(row.supportedDevices || ""),
      reason: "initial_game_selection_draft",
    });
    ElMessage.success(t("brandPortal.messages.gameDraftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await loadGames();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.saveGameDraftFailed"));
  }
}

function parseConfig(value: unknown): Record<string, unknown> {
  if (objectValue(value) === value) return value as Record<string, unknown>;
  try {
    const parsed = JSON.parse(String(value || "{}"));
    return objectValue(parsed);
  } catch {
    return {};
  }
}

function lobbyPayload(row: GameCenterRecord, patch: Partial<LobbySectionInput>): LobbySectionInput {
  return {
    sectionCode: String(row.sectionCode || ""),
    sectionType: String(row.sectionType || "game-list"),
    titleZh: String(row.titleZh || ""),
    titleEn: String(row.titleEn || ""),
    displayStyle: String(row.displayStyle || "horizontal"),
    sourceType: String(row.sourceType || "flag"),
    sourceRef: String(row.sourceRef || ""),
    maxItems: Number(row.maxItems || 12),
    sortOrder: Number(row.sortOrder || 0),
    enabled: Boolean(row.enabled),
    supportedCountries: String(row.supportedCountries || ""),
    supportedCurrencies: String(row.supportedCurrencies || ""),
    config: parseConfig(row.configJSON || row.configJson || "{}"),
    reason: "initial_lobby_setup_draft",
    ...patch,
  };
}

async function loadLobby() {
  lobbyLoading.value = true;
  try {
    const result = await fetchLobbySections({ page: 1, pageSize: 50 });
    lobbySections.value = result.list || [];
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.lobbyLoadFailed"));
  } finally {
    lobbyLoading.value = false;
  }
}

async function saveLobbySection(row: GameCenterRecord, patch: Partial<LobbySectionInput>) {
  try {
    await updateLobbySection(row.id, lobbyPayload(row, patch));
    ElMessage.success(t("brandPortal.messages.lobbyDraftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await loadLobby();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.saveLobbyDraftFailed"));
  }
}

async function preview() {
  previewVisible.value = true;
  previewLoading.value = true;
  try {
    await systemAPI.sites.preview(props.overview.siteID);
  } catch (error) {
    previewVisible.value = false;
    ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.previewFailed"));
  } finally {
    previewLoading.value = false;
  }
}

async function publish() {
  try {
    const result = await ElMessageBox.prompt(t("brandPortal.prompts.publishReason"), t("brandPortal.dialogs.publishTitle"), {
      confirmButtonText: t("brandPortal.actions.publish"),
      cancelButtonText: t("brandPortal.actions.cancel"),
      inputValue: t("brandPortal.defaults.initialBrandPublish"),
      inputValidator: (value) => Boolean(value.trim()) || t("brandPortal.validation.reasonRequired"),
    });
    publishing.value = true;
    await systemAPI.sites.publish(props.overview.siteID, result.value.trim());
    ElMessage.success(t("brandPortal.messages.publishedToH5"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    emit("completed");
  } catch (error) {
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.publishFailed"));
    }
  } finally {
    publishing.value = false;
  }
}

function nextStep() {
  active.value = Math.min(active.value + 1, 5);
  if (active.value === 3 && !games.value.length) void loadGames();
  if (active.value === 4 && !lobbySections.value.length) void loadLobby();
}

watch(
  () => props.overview,
  (row) => assignFromOverview(row),
  { immediate: true },
);

onMounted(() => {
  void loadGames();
  void loadLobby();
});
</script>

<template>
  <div class="setup-wizard">
    <section class="wizard-header">
      <div>
        <p class="eyebrow">{{ t("brandPortal.wizard.eyebrow") }}</p>
        <h1>{{ form.siteName || props.overview.brandName || t("brandPortal.wizard.fallbackTitle") }}</h1>
      </div>
      <el-tag effect="plain">{{ t("brandPortal.wizard.draftVersion", { version: props.overview.draftVersion || 1 }) }}</el-tag>
    </section>

    <el-steps :active="active" finish-status="success" class="wizard-steps">
      <el-step :title="t('brandPortal.wizard.steps.brand')" />
      <el-step :title="t('brandPortal.wizard.steps.logoTheme')" />
      <el-step :title="t('brandPortal.wizard.steps.locale')" />
      <el-step :title="t('brandPortal.wizard.steps.games')" />
      <el-step :title="t('brandPortal.wizard.steps.lobby')" />
      <el-step :title="t('brandPortal.wizard.steps.publish')" />
    </el-steps>

    <section v-if="active <= 2" class="wizard-panel">
      <el-form label-position="left" label-width="140px">
        <template v-if="active === 0">
          <el-form-item :label="t('brandPortal.settings.fields.brandName')"><el-input v-model="form.siteName" /></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.appName')"><el-input v-model="form.appName" /></el-form-item>
          <el-form-item :label="t('brandPortal.wizard.fields.supportUrl')"><el-input v-model="form.customerService" /></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.email')"><el-input v-model="form.email" /></el-form-item>
        </template>
        <template v-else-if="active === 1">
          <el-form-item :label="t('brandPortal.settings.fields.logo')"><el-input v-model="form.logo" placeholder="https://..." /></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.primary')"><div class="color-field"><el-color-picker v-model="form.primaryColor" /><el-input v-model="form.primaryColor" /></div></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.secondary')"><div class="color-field"><el-color-picker v-model="form.secondaryColor" /><el-input v-model="form.secondaryColor" /></div></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.button')"><div class="color-field"><el-color-picker v-model="form.buttonColor" /><el-input v-model="form.buttonColor" /></div></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.background')"><div class="color-field"><el-color-picker v-model="form.backgroundColor" /><el-input v-model="form.backgroundColor" /></div></el-form-item>
        </template>
        <template v-else>
          <el-form-item :label="t('brandPortal.settings.fields.country')"><el-input v-model="form.countryCode" disabled /></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.currency')"><el-input v-model="form.currencyCode" disabled /></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.language')"><el-input v-model="form.defaultLanguage" /></el-form-item>
          <el-form-item :label="t('brandPortal.settings.fields.timezone')"><el-input v-model="form.timezone" /></el-form-item>
        </template>
      </el-form>
      <div class="panel-actions">
        <el-button :disabled="active === 0" @click="active--">{{ t("brandPortal.actions.back") }}</el-button>
        <el-button :loading="saving" @click="saveBrandDraft">{{ t("brandPortal.actions.saveDraft") }}</el-button>
        <el-button type="primary" @click="nextStep">{{ t("brandPortal.actions.next") }}</el-button>
      </div>
    </section>

    <section v-else-if="active === 3" class="wizard-panel">
      <el-table v-loading="gamesLoading" :data="games" height="420">
        <el-table-column prop="thumbnailUrl" :label="t('brandPortal.wizard.fields.icon')" width="76">
          <template #default="{ row }">
            <img v-if="row.thumbnailUrl || row.localThumbnailUrl" class="game-icon" :src="String(row.localThumbnailUrl || row.thumbnailUrl)" alt="" />
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="t('brandPortal.wizard.fields.game')" min-width="180" />
        <el-table-column prop="providerCode" :label="t('brandPortal.wizard.fields.provider')" width="120">
          <template #default="{ row }"><el-tag effect="plain">{{ row.providerCode || "-" }}</el-tag></template>
        </el-table-column>
        <el-table-column :label="t('brandPortal.wizard.fields.enabled')" width="120">
          <template #default="{ row }">
            <el-switch :model-value="gameBool(row, 'platformEnabled')" @change="(value: boolean | string | number) => saveGameEnabled(row, Boolean(value))" />
          </template>
        </el-table-column>
      </el-table>
      <div class="panel-actions">
        <el-button @click="active--">{{ t("brandPortal.actions.back") }}</el-button>
        <el-button @click="loadGames">{{ t("brandPortal.actions.reload") }}</el-button>
        <el-button type="primary" @click="nextStep">{{ t("brandPortal.actions.next") }}</el-button>
      </div>
    </section>

    <section v-else-if="active === 4" class="wizard-panel">
      <el-table v-loading="lobbyLoading" :data="lobbySections" height="420">
        <el-table-column prop="sectionCode" :label="t('brandPortal.wizard.fields.section')" min-width="160" />
        <el-table-column prop="titleEn" :label="t('brandPortal.wizard.fields.title')" min-width="160" />
        <el-table-column prop="sectionType" :label="t('brandPortal.wizard.fields.type')" width="120">
          <template #default="{ row }"><el-tag effect="plain">{{ row.sectionType || "-" }}</el-tag></template>
        </el-table-column>
        <el-table-column :label="t('brandPortal.wizard.fields.sort')" width="150">
          <template #default="{ row }">
            <el-input-number
              :model-value="Number(row.sortOrder || 0)"
              :min="0"
              :controls="false"
              size="small"
              @change="(value: number | undefined) => saveLobbySection(row, { sortOrder: Number(value || 0) })"
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('brandPortal.wizard.fields.enabled')" width="120">
          <template #default="{ row }">
            <el-switch :model-value="Boolean(row.enabled)" @change="(value: boolean | string | number) => saveLobbySection(row, { enabled: Boolean(value) })" />
          </template>
        </el-table-column>
      </el-table>
      <div class="panel-actions">
        <el-button @click="active--">{{ t("brandPortal.actions.back") }}</el-button>
        <el-button @click="loadLobby">{{ t("brandPortal.actions.reload") }}</el-button>
        <el-button type="primary" @click="nextStep">{{ t("brandPortal.actions.next") }}</el-button>
      </div>
    </section>

    <section v-else class="wizard-panel publish-panel">
      <dl>
        <dt>{{ t("brandPortal.wizard.publishPanel.brand") }}</dt><dd>{{ form.siteName || "-" }}</dd>
        <dt>{{ t("brandPortal.wizard.publishPanel.domain") }}</dt><dd>{{ props.overview.h5Domain || "-" }}</dd>
        <dt>{{ t("brandPortal.wizard.publishPanel.currency") }}</dt><dd>{{ form.currencyCode || "-" }}</dd>
        <dt>{{ t("brandPortal.wizard.publishPanel.draftVersion") }}</dt><dd>v{{ props.overview.draftVersion || 1 }}</dd>
      </dl>
      <div class="panel-actions">
        <el-button @click="active--">{{ t("brandPortal.actions.back") }}</el-button>
        <el-button :icon="Monitor" @click="preview">{{ t("brandPortal.actions.preview") }}</el-button>
        <el-button type="primary" :icon="UploadFilled" :loading="publishing" @click="publish">{{ t("brandPortal.actions.publish") }}</el-button>
      </div>
    </section>

    <el-dialog v-model="previewVisible" :title="t('brandPortal.dialogs.previewTitle')" width="min(920px, 94vw)" destroy-on-close>
      <H5PreviewFrame v-loading="previewLoading" :h5-domain="props.overview.h5Domain" :site-i-d="props.overview.siteID" />
      <template #footer>
        <el-button @click="previewVisible = false">{{ t("brandPortal.actions.close") }}</el-button>
        <el-button type="primary" :loading="publishing" @click="publish">{{ t("brandPortal.actions.publish") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.setup-wizard {
  display: grid;
  gap: 16px;
}
.wizard-header,
.wizard-panel {
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.wizard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 20px;
}
.wizard-steps {
  padding: 12px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.wizard-panel {
  padding: 18px;
}
.eyebrow {
  margin: 0 0 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  text-transform: uppercase;
}
h1 {
  margin: 0;
}
.color-field {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 8px;
  width: 100%;
}
.panel-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 16px;
}
.game-icon {
  width: 44px;
  height: 44px;
  object-fit: cover;
  border: 1px solid var(--el-border-color-lighter);
}
.publish-panel dl {
  display: grid;
  grid-template-columns: 140px minmax(0, 1fr);
  gap: 12px;
  margin: 0;
}
.publish-panel dt {
  color: var(--el-text-color-secondary);
}
.publish-panel dd {
  min-width: 0;
  margin: 0;
  overflow-wrap: anywhere;
}
@media (max-width: 900px) {
  .wizard-header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
