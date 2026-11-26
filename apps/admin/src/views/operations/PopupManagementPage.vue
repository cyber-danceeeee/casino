<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  CopyDocument,
  Delete,
  Edit,
  Plus,
  Rank,
  Refresh,
  UploadFilled,
  VideoPlay,
} from "@element-plus/icons-vue";
import { fetchGames, type GameCenterRecord } from "@/api/adminGames";
import {
  systemAPI,
  type BrandOverview,
  type ContentVisibility,
  type PopupActionType,
  type PopupComponentRecord,
  type PopupConfiguration,
  type PopupFrequency,
  type PopupTrigger,
  type SitePopupRecord,
} from "@/api/adminSystem";
import { formatBackendTime } from "@/utils/timezone";

const { t, locale } = useI18n();
const loading = ref(false),
  saving = ref(false),
  publishing = ref(false),
  sorting = ref(false);
const rows = ref<SitePopupRecord[]>([]),
  components = ref<PopupComponentRecord[]>([]),
  games = ref<GameCenterRecord[]>([]);
const overview = ref<BrandOverview | null>(null);
const editVisible = ref(false),
  editingID = ref(0),
  draggedID = ref(0),
  wizardStep = ref(0);
const filters = reactive({ keyword: "", publishStatus: "" });
const frequencies: PopupFrequency[] = [
  "EVERY_VISIT",
  "ONCE_PER_SESSION",
  "ONCE_PER_DAY",
  "ONCE_PER_USER",
];
const triggers: PopupTrigger[] = ["HOME_ENTER", "PAGE_ENTER"];
const visibility: ContentVisibility[] = ["ALL", "GUEST", "LOGGED_IN"];

const emptyForm = () => ({
  popupComponentID: 0,
  campaignName: "",
  actionType: "CLOSE_ONLY" as PopupActionType,
  actionTarget: "",
  triggerType: "HOME_ENTER" as PopupTrigger,
  triggerValue: "",
  frequencyType: "ONCE_PER_SESSION" as PopupFrequency,
  displayLimit: 0,
  cooldownMinutes: 0,
  visibilityScope: "ALL" as ContentVisibility,
  startAt: null as Date | null,
  endAt: null as Date | null,
  priority: 0,
  enabled: true,
  reason: "",
});
const form = reactive(emptyForm());
const selectedComponent = computed(() =>
  components.value.find((item) => item.id === form.popupComponentID),
);
const allowedActions = computed(
  () => selectedComponent.value?.allowedActionTypes || [],
);
const targetlessActions: PopupActionType[] = [
  "NONE",
  "CLOSE_ONLY",
  "OPEN_DEPOSIT",
  "OPEN_WITHDRAW",
  "OPEN_SUPPORT",
  "OPEN_COUPON",
  "OPEN_VIP",
  "OPEN_DOWNLOAD",
];
const targetRequired = computed(
  () => !targetlessActions.includes(form.actionType),
);
const triggerTargetRequired = computed(() => form.triggerType === "PAGE_ENTER");
type PopupPreviewConfiguration = Pick<
  PopupConfiguration,
  | "campaignName"
  | "actionType"
  | "actionTarget"
  | "triggerType"
  | "triggerValue"
  | "frequencyType"
  | "displayLimit"
  | "cooldownMinutes"
  | "priority"
>;
const draft = (row: SitePopupRecord) =>
  row.draftConfiguration || row.publishedConfiguration;
const componentName = (component?: PopupComponentRecord) =>
  !component
    ? "-"
    : locale.value.startsWith("zh")
      ? component.nameZh
      : component.nameEn;
const componentDescription = (component?: PopupComponentRecord) =>
  !component
    ? ""
    : locale.value.startsWith("zh")
      ? component.descriptionZh
      : component.descriptionEn;
const gameLabel = (game: GameCenterRecord) =>
  `${String(game.name || game.gameName || game.gameCode || "-")} · ${String(game.providerCode || "")}`;
function h5Origin() {
  const raw = String(overview.value?.h5Domain || "").trim();
  if (!raw) return "";
  return /^https?:\/\//i.test(raw)
    ? raw.replace(/\/$/, "")
    : `https://${raw.replace(/\/$/, "")}`;
}
function popupPreviewURL(
  componentCode: string,
  configuration: PopupPreviewConfiguration = form,
) {
  const origin = h5Origin();
  if (!origin || !componentCode) return "";
  const url = new URL(
    `/popup-preview/${encodeURIComponent(componentCode)}`,
    `${origin}/`,
  );
  url.searchParams.set("campaignName", configuration.campaignName || "");
  url.searchParams.set("actionType", configuration.actionType || "CLOSE_ONLY");
  url.searchParams.set("actionTarget", configuration.actionTarget || "");
  url.searchParams.set(
    "triggerType",
    configuration.triggerType || "HOME_ENTER",
  );
  url.searchParams.set("triggerValue", configuration.triggerValue || "");
  url.searchParams.set(
    "frequencyType",
    configuration.frequencyType || "EVERY_VISIT",
  );
  url.searchParams.set("displayLimit", String(configuration.displayLimit || 0));
  url.searchParams.set(
    "cooldownMinutes",
    String(configuration.cooldownMinutes || 0),
  );
  url.searchParams.set("priority", String(configuration.priority || 0));
  return url.toString();
}
const livePreviewURL = computed(() =>
  popupPreviewURL(selectedComponent.value?.componentCode || ""),
);

async function load() {
  loading.value = true;
  try {
    const brand = await systemAPI.brand.overview();
    overview.value = brand;
    const [campaignResult, componentResult, gameResult] = await Promise.all([
      systemAPI.popups.list(brand.siteID, {
        page: 1,
        pageSize: 200,
        keyword: filters.keyword,
        publishStatus: filters.publishStatus,
      }),
      systemAPI.popupComponents.available(),
      fetchGames({ page: 1, pageSize: 200 }).catch(() => ({
        list: [],
        total: 0,
        page: 1,
        pageSize: 200,
      })),
    ]);
    rows.value = campaignResult.list || [];
    components.value = componentResult.list || [];
    games.value = gameResult.list;
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("popupLibrary.messages.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}
function applyComponentDefaults() {
  const component = selectedComponent.value;
  if (!component) return;
  if (!component.allowedActionTypes.includes(form.actionType))
    form.actionType = component.defaultActionType;
  if (targetlessActions.includes(form.actionType)) form.actionTarget = "";
}
function openCreate() {
  editingID.value = 0;
  Object.assign(form, emptyForm(), {
    popupComponentID: components.value[0]?.id || 0,
    priority: rows.value.length + 1,
  });
  applyComponentDefaults();
  wizardStep.value = 0;
  editVisible.value = true;
}
function selectComponent(component: PopupComponentRecord) {
  form.popupComponentID = component.id;
  applyComponentDefaults();
}
function openEdit(row: SitePopupRecord) {
  const value = draft(row);
  editingID.value = row.id;
  Object.assign(form, emptyForm(), value, {
    startAt: value.startAt ? new Date(value.startAt) : null,
    endAt: value.endAt ? new Date(value.endAt) : null,
    reason: "",
  });
  wizardStep.value = 1;
  editVisible.value = true;
}
function payload(
  overrides: Partial<PopupConfiguration> = {},
): PopupConfiguration & { reason: string } {
  return {
    popupComponentID: form.popupComponentID,
    campaignName: form.campaignName.trim(),
    actionType: form.actionType,
    actionTarget: targetRequired.value ? form.actionTarget.trim() : "",
    triggerType: form.triggerType,
    triggerValue: triggerTargetRequired.value ? form.triggerValue.trim() : "",
    frequencyType: form.frequencyType,
    displayLimit: form.displayLimit,
    cooldownMinutes: form.cooldownMinutes,
    visibilityScope: form.visibilityScope,
    startAt: form.startAt ? form.startAt.toISOString() : null,
    endAt: form.endAt ? form.endAt.toISOString() : null,
    priority: form.priority,
    enabled: form.enabled,
    ...overrides,
    reason: form.reason.trim() || t("popupLibrary.defaults.saveReason"),
  };
}
async function save() {
  if (
    !overview.value?.siteID ||
    !form.popupComponentID ||
    !form.campaignName.trim()
  ) {
    ElMessage.warning(t("popupLibrary.validation.required"));
    return;
  }
  if (
    (targetRequired.value && !form.actionTarget.trim()) ||
    (triggerTargetRequired.value && !form.triggerValue.trim())
  ) {
    ElMessage.warning(t("popupLibrary.validation.targetRequired"));
    return;
  }
  if (
    form.frequencyType === "ONCE_PER_USER" &&
    form.visibilityScope === "GUEST"
  ) {
    ElMessage.warning(t("popupLibrary.validation.onceUserLogin"));
    return;
  }
  saving.value = true;
  try {
    if (editingID.value)
      await systemAPI.popups.update(
        overview.value.siteID,
        editingID.value,
        payload(),
      );
    else await systemAPI.popups.create(overview.value.siteID, payload());
    editVisible.value = false;
    ElMessage.success(t("popupLibrary.messages.draftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("popupLibrary.messages.saveFailed"),
    );
  } finally {
    saving.value = false;
  }
}
async function toggle(row: SitePopupRecord, enabled: boolean) {
  if (!overview.value?.siteID) return;
  try {
    await systemAPI.popups.update(overview.value.siteID, row.id, {
      ...draft(row),
      enabled,
      reason: t("popupLibrary.defaults.statusReason"),
    });
    ElMessage.success(t("popupLibrary.messages.draftSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("popupLibrary.messages.saveFailed"),
    );
    await load();
  }
}
async function duplicate(row: SitePopupRecord) {
  if (!overview.value?.siteID) return;
  try {
    await systemAPI.popups.copy(overview.value.siteID, row.id);
    ElMessage.success(t("popupLibrary.messages.copied"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("popupLibrary.messages.copyFailed"),
    );
  }
}
async function remove(row: SitePopupRecord) {
  if (!overview.value?.siteID) return;
  try {
    await ElMessageBox.confirm(
      t("popupLibrary.dialogs.deleteMessage"),
      t("popupLibrary.dialogs.deleteTitle"),
      {
        type: "warning",
        confirmButtonText: t("popupLibrary.actions.deleteDraft"),
        cancelButtonText: t("popupLibrary.actions.cancel"),
      },
    );
    await systemAPI.popups.remove(overview.value.siteID, row.id);
    ElMessage.success(t("popupLibrary.messages.deleteSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("popupLibrary.messages.deleteFailed"),
      );
  }
}
function preview(row: SitePopupRecord) {
  const target = popupPreviewURL(
    row.component?.componentCode || "",
    draft(row),
  );
  if (target) window.open(target, "_blank", "noopener,noreferrer");
  else ElMessage.info(t("popupLibrary.messages.noPreview"));
}
function dragStart(row: SitePopupRecord) {
  draggedID.value = row.id;
}
async function dropOn(targetID: number) {
  if (!overview.value?.siteID) return;
  const sourceIndex = rows.value.findIndex((row) => row.id === draggedID.value),
    targetIndex = rows.value.findIndex((row) => row.id === targetID);
  draggedID.value = 0;
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return;
  const ordered = [...rows.value];
  const [moved] = ordered.splice(sourceIndex, 1);
  ordered.splice(targetIndex, 0, moved);
  sorting.value = true;
  try {
    await systemAPI.popups.reorder(
      overview.value.siteID,
      ordered.map((row) => row.id),
      t("popupLibrary.defaults.sortReason"),
    );
    ElMessage.success(t("popupLibrary.messages.sortSaved"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("popupLibrary.messages.sortFailed"),
    );
    await load();
  } finally {
    sorting.value = false;
  }
}
async function publish() {
  if (!overview.value?.siteID) return;
  try {
    const prompt = await ElMessageBox.prompt(
      t("popupLibrary.dialogs.publishMessage"),
      t("popupLibrary.dialogs.publishTitle"),
      {
        confirmButtonText: t("popupLibrary.actions.publish"),
        cancelButtonText: t("popupLibrary.actions.cancel"),
        inputValue: t("popupLibrary.defaults.publishReason"),
        inputValidator: (value) =>
          Boolean(value.trim()) || t("popupLibrary.validation.reasonRequired"),
      },
    );
    publishing.value = true;
    const result = await systemAPI.sites.publish(
      overview.value.siteID,
      prompt.value.trim(),
    );
    ElMessage.success(
      t("popupLibrary.messages.published", {
        count: Number(result.popupDraftsPublished || 0),
      }),
    );
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("popupLibrary.messages.publishFailed"),
      );
  } finally {
    publishing.value = false;
  }
}
onMounted(load);
</script>

<template>
  <div v-loading="loading" class="content-page">
    <header class="page-header">
      <div>
        <p class="eyebrow">{{ t("popupLibrary.eyebrow") }}</p>
        <h1>{{ t("popupLibrary.title") }}</h1>
        <p>{{ t("popupLibrary.subtitle") }}</p>
      </div>
      <div class="header-actions">
        <el-button :icon="Refresh" @click="load">{{
          t("popupLibrary.actions.refresh")
        }}</el-button
        ><el-button
          type="primary"
          :icon="UploadFilled"
          :loading="publishing"
          @click="publish"
          >{{ t("popupLibrary.actions.publish") }}</el-button
        ><el-button type="success" :icon="Plus" @click="openCreate">{{
          t("popupLibrary.actions.create")
        }}</el-button>
      </div>
    </header>
    <el-alert
      type="warning"
      :title="t('popupLibrary.draftNotice')"
      :closable="false"
      show-icon
    />
    <section class="toolbar">
      <el-input
        v-model="filters.keyword"
        clearable
        :placeholder="t('popupLibrary.placeholders.keyword')"
        @keyup.enter="load"
      /><el-select
        v-model="filters.publishStatus"
        clearable
        :placeholder="t('popupLibrary.fields.publishStatus')"
        ><el-option
          :label="t('popupLibrary.status.DRAFT')"
          value="DRAFT" /><el-option
          :label="t('popupLibrary.status.PUBLISHED')"
          value="PUBLISHED" /></el-select
      ><el-button type="primary" @click="load">{{
        t("popupLibrary.actions.search")
      }}</el-button>
    </section>
    <section class="table-shell">
      <el-table :data="rows" row-key="id" v-loading="sorting">
        <el-table-column width="48"
          ><template #default="{ row }"
            ><el-icon
              class="drag-handle"
              draggable="true"
              @dragstart="dragStart(row)"
              @dragover.prevent
              @drop="dropOn(row.id)"
              ><Rank /></el-icon></template
        ></el-table-column>
        <el-table-column
          :label="t('popupLibrary.fields.component')"
          min-width="230"
          ><template #default="{ row }"
            ><strong>{{ componentName(row.component) }}</strong
            ><small>{{ row.component?.componentCode }}</small
            ><small>{{ componentDescription(row.component) }}</small></template
          ></el-table-column
        >
        <el-table-column
          :label="t('popupLibrary.fields.campaignName')"
          min-width="180"
          ><template #default="{ row }"
            ><strong>{{ draft(row).campaignName }}</strong
            ><small
              >{{ t(`popupLibrary.trigger.${draft(row).triggerType}`)
              }}<template v-if="draft(row).triggerValue">
                · {{ draft(row).triggerValue }}</template
              ></small
            ></template
          ></el-table-column
        >
        <el-table-column :label="t('popupLibrary.fields.frequency')" width="150"
          ><template #default="{ row }"
            ><el-tag effect="plain">{{
              t(`popupLibrary.frequency.${draft(row).frequencyType}`)
            }}</el-tag></template
          ></el-table-column
        >
        <el-table-column
          :label="t('popupLibrary.fields.actionType')"
          min-width="170"
          ><template #default="{ row }"
            ><span>{{
              t(`popupLibrary.actionType.${draft(row).actionType}`)
            }}</span
            ><small v-if="draft(row).actionTarget">{{
              draft(row).actionTarget
            }}</small></template
          ></el-table-column
        >
        <el-table-column
          :label="t('popupLibrary.fields.schedule')"
          min-width="180"
          ><template #default="{ row }"
            ><span>{{ formatBackendTime(draft(row).startAt || "") }}</span
            ><small>{{
              formatBackendTime(draft(row).endAt || "")
            }}</small></template
          ></el-table-column
        >
        <el-table-column :label="t('popupLibrary.fields.priority')" width="90"
          ><template #default="{ row }">{{
            draft(row).priority
          }}</template></el-table-column
        >
        <el-table-column
          :label="t('popupLibrary.fields.publishStatus')"
          width="145"
          ><template #default="{ row }"
            ><el-tag
              :type="row.publishStatus === 'DRAFT' ? 'warning' : 'success'"
              >{{ t(`popupLibrary.status.${row.publishStatus}`) }}</el-tag
            ></template
          ></el-table-column
        >
        <el-table-column :label="t('popupLibrary.fields.enabled')" width="90"
          ><template #default="{ row }"
            ><el-switch
              :model-value="draft(row).enabled"
              @change="toggle(row, Boolean($event))" /></template
        ></el-table-column>
        <el-table-column
          :label="t('popupLibrary.fields.actions')"
          width="190"
          fixed="right"
          ><template #default="{ row }"
            ><el-tooltip :content="t('popupLibrary.actions.preview')"
              ><el-button
                circle
                :icon="VideoPlay"
                @click="preview(row)" /></el-tooltip
            ><el-tooltip :content="t('popupLibrary.actions.edit')"
              ><el-button
                circle
                :icon="Edit"
                @click="openEdit(row)" /></el-tooltip
            ><el-tooltip :content="t('popupLibrary.actions.copy')"
              ><el-button
                circle
                :icon="CopyDocument"
                @click="duplicate(row)" /></el-tooltip
            ><el-tooltip :content="t('popupLibrary.actions.deleteDraft')"
              ><el-button
                circle
                type="danger"
                plain
                :icon="Delete"
                @click="remove(row)" /></el-tooltip></template
        ></el-table-column>
        <template #empty>
          <el-empty :description="t('popupLibrary.messages.empty')">
            <el-button type="primary" :icon="Plus" @click="openCreate">{{
              t("popupLibrary.actions.create")
            }}</el-button>
          </el-empty>
        </template>
      </el-table>
    </section>

    <el-dialog
      v-model="editVisible"
      :title="
        editingID
          ? t('popupLibrary.dialogs.editTitle')
          : t('popupLibrary.dialogs.createTitle')
      "
      width="min(1040px, 96vw)"
      destroy-on-close
    >
      <el-steps
        :active="wizardStep"
        finish-status="success"
        align-center
        class="campaign-steps"
      >
        <el-step :title="t('popupLibrary.wizard.template')" />
        <el-step :title="t('popupLibrary.wizard.delivery')" />
        <el-step :title="t('popupLibrary.wizard.action')" />
        <el-step :title="t('popupLibrary.wizard.preview')" />
      </el-steps>
      <section v-show="wizardStep === 0" class="template-grid">
        <button
          v-for="component in components"
          :key="component.componentCode"
          type="button"
          class="template-card"
          :class="{ selected: component.id === form.popupComponentID }"
          @click="selectComponent(component)"
        >
          <div class="template-media">
            <video
              v-if="component.previewVideoURL"
              :src="component.previewVideoURL"
              muted
              loop
              playsinline
            />
            <img
              v-else-if="component.previewURL"
              :src="component.previewURL"
              alt=""
            />
            <VideoPlay v-else />
          </div>
          <strong>{{ componentName(component) }}</strong>
          <span
            >{{ component.category }} · H5
            {{ component.h5MinVersion || "-" }}</span
          >
          <small>{{ componentDescription(component) }}</small>
          <el-tag
            :type="component.enabled ? 'success' : 'info'"
            effect="plain"
            >{{
              component.enabled
                ? t("popupLibrary.wizard.available")
                : t("popupLibrary.wizard.unavailable")
            }}</el-tag
          >
        </button>
      </section>
      <el-form
        v-show="wizardStep === 1 || wizardStep === 2"
        label-position="top"
        ><el-row :gutter="16">
          <el-col v-show="false" :span="12"
            ><el-form-item :label="t('popupLibrary.fields.component')" required
              ><el-select
                v-model="form.popupComponentID"
                filterable
                @change="applyComponentDefaults"
                ><el-option
                  v-for="component in components"
                  :key="component.id"
                  :label="`${componentName(component)} · ${component.componentCode}`"
                  :value="component.id" /></el-select></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="12"
            ><el-form-item
              :label="t('popupLibrary.fields.campaignName')"
              required
              ><el-input v-model="form.campaignName" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="12"
            ><el-form-item :label="t('popupLibrary.fields.trigger')"
              ><el-select v-model="form.triggerType"
                ><el-option
                  v-for="item in triggers"
                  :key="item"
                  :label="t(`popupLibrary.trigger.${item}`)"
                  :value="item" /></el-select></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="12"
            ><el-form-item
              v-if="triggerTargetRequired"
              :label="t('popupLibrary.fields.triggerTarget')"
              required
              ><el-input
                v-model="form.triggerValue"
                placeholder="/offers" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 2" :span="12"
            ><el-form-item :label="t('popupLibrary.fields.actionType')"
              ><el-select v-model="form.actionType"
                ><el-option
                  v-for="item in allowedActions"
                  :key="item"
                  :label="t(`popupLibrary.actionType.${item}`)"
                  :value="item" /></el-select></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 2" :span="12"
            ><el-form-item
              v-if="targetRequired"
              :label="t('popupLibrary.fields.actionTarget')"
              required
              ><el-select
                v-if="form.actionType === 'OPEN_GAME'"
                v-model="form.actionTarget"
                filterable
                ><el-option
                  v-for="game in games"
                  :key="String(game.gameCode)"
                  :label="gameLabel(game)"
                  :value="String(game.gameCode)" /></el-select
              ><el-input v-else v-model="form.actionTarget" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="8"
            ><el-form-item :label="t('popupLibrary.fields.frequency')"
              ><el-select v-model="form.frequencyType"
                ><el-option
                  v-for="item in frequencies"
                  :key="item"
                  :label="t(`popupLibrary.frequency.${item}`)"
                  :value="item" /></el-select></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="8"
            ><el-form-item :label="t('popupLibrary.fields.displayLimit')"
              ><el-input-number
                v-model="form.displayLimit"
                :min="0" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="8"
            ><el-form-item :label="t('popupLibrary.fields.cooldown')"
              ><el-input-number
                v-model="form.cooldownMinutes"
                :min="0" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="8"
            ><el-form-item :label="t('popupLibrary.fields.visibility')"
              ><el-select v-model="form.visibilityScope"
                ><el-option
                  v-for="item in visibility"
                  :key="item"
                  :label="t(`popupLibrary.visibility.${item}`)"
                  :value="item" /></el-select></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="8"
            ><el-form-item :label="t('popupLibrary.fields.priority')"
              ><el-input-number
                v-model="form.priority"
                :min="0" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="8"
            ><el-form-item :label="t('popupLibrary.fields.enabled')"
              ><el-switch v-model="form.enabled" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 1" :span="12"
            ><el-form-item :label="t('popupLibrary.fields.startAt')"
              ><el-date-picker
                v-model="form.startAt"
                type="datetime" /></el-form-item></el-col
          ><el-col v-show="wizardStep === 1" :span="12"
            ><el-form-item :label="t('popupLibrary.fields.endAt')"
              ><el-date-picker
                v-model="form.endAt"
                type="datetime" /></el-form-item
          ></el-col>
          <el-col v-show="wizardStep === 2" :span="24"
            ><el-form-item :label="t('popupLibrary.fields.reason')"
              ><el-input
                v-model="form.reason"
                type="textarea"
                :rows="2" /></el-form-item
          ></el-col> </el-row
      ></el-form>
      <section v-show="wizardStep === 3" class="preview-step">
        <header>
          <strong>{{
            form.campaignName || t("popupLibrary.wizard.unnamed")
          }}</strong
          ><span>{{ componentName(selectedComponent) }}</span>
        </header>
        <div class="phone-preview">
          <iframe
            v-if="livePreviewURL"
            :src="livePreviewURL"
            :title="t('popupLibrary.wizard.preview')"
          />
          <el-empty
            v-else
            :description="t('popupLibrary.messages.noPreview')"
          />
        </div>
      </section>
      <template #footer>
        <el-button @click="editVisible = false">{{
          t("popupLibrary.actions.cancel")
        }}</el-button>
        <el-button v-if="wizardStep > 0" @click="wizardStep--">{{
          t("brandPortal.actions.back")
        }}</el-button>
        <el-button
          v-if="wizardStep < 3"
          type="primary"
          :disabled="wizardStep === 0 && !form.popupComponentID"
          @click="wizardStep++"
          >{{ t("brandPortal.actions.next") }}</el-button
        >
        <el-button v-else type="primary" :loading="saving" @click="save">{{
          t("popupLibrary.actions.saveDraft")
        }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.content-page {
  padding: 24px;
  color: #18212f;
}
.page-header {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  align-items: flex-start;
  margin-bottom: 18px;
}
.page-header h1 {
  margin: 0;
  font-size: 28px;
  letter-spacing: 0;
}
.page-header p {
  margin: 7px 0 0;
  color: #667085;
}
.eyebrow {
  color: #b42318 !important;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}
.header-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}
.toolbar {
  display: grid;
  grid-template-columns: minmax(240px, 1fr) 190px auto;
  gap: 10px;
  margin: 16px 0;
}
.table-shell {
  border: 1px solid #e4e7ec;
  border-radius: 6px;
  overflow: hidden;
  background: #fff;
}
.campaign-steps {
  margin: 4px 0 24px;
}
.template-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  max-height: 520px;
  overflow: auto;
}
.template-card {
  display: grid;
  gap: 7px;
  min-width: 0;
  padding: 10px;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  color: var(--el-text-color-primary);
  text-align: left;
  background: var(--el-bg-color);
  cursor: pointer;
}
.template-card.selected {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 2px var(--el-color-primary-light-8);
}
.template-card span,
.template-card small {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.template-media {
  display: grid;
  aspect-ratio: 16 / 9;
  place-items: center;
  overflow: hidden;
  border-radius: 4px;
  color: var(--el-text-color-placeholder);
  background: var(--el-fill-color-light);
}
.template-media img,
.template-media video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.template-media svg {
  width: 30px;
}
.preview-step {
  display: grid;
  justify-items: center;
  gap: 12px;
}
.preview-step > header {
  text-align: center;
}
.preview-step > header span {
  display: block;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
}
.phone-preview {
  width: min(390px, 100%);
  aspect-ratio: 9 / 16;
  overflow: hidden;
  border: 8px solid #1f2937;
  border-radius: 28px;
  background: #090b10;
}
.phone-preview iframe {
  width: 100%;
  height: 100%;
  border: 0;
}
.drag-handle {
  cursor: grab;
  color: #667085;
  font-size: 18px;
}
strong,
small {
  display: block;
}
small {
  margin-top: 4px;
  color: #667085;
}
.el-form :deep(.el-select),
.el-form :deep(.el-date-editor),
.el-form :deep(.el-input-number) {
  width: 100%;
}
@media (max-width: 760px) {
  .content-page {
    padding: 16px;
  }
  .page-header {
    flex-direction: column;
  }
  .header-actions {
    justify-content: flex-start;
  }
  .toolbar {
    grid-template-columns: 1fr;
  }
  .template-grid {
    grid-template-columns: 1fr;
  }
}
</style>
