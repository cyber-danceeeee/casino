<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ElMessage, ElMessageBox } from "element-plus";
import { UploadFilled } from "@element-plus/icons-vue";
import type { UploadRequestOptions } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { businessColumns, businessEntries, isInternalIdentifierKey } from "@/utils/presentation";
import {
  addCategoryGames,
  applyProviderCategoryMapping,
  batchUpdateGames,
  createCategory,
  createCollection,
  createLobbySection,
  createTag,
  cancelProviderSync,
  deleteCategory,
  deleteCollection,
  deleteLobbySection,
  deleteProvider,
  deleteTag,
  fetchCategories,
  fetchCategoryGameCandidates,
  fetchCategoryGames,
  fetchCategoryOperationLogs,
  fetchCollections,
  fetchImportableGames,
  fetchImportableProviderGameTypes,
  fetchProviderOptionGameTypes,
  fetchProviderGameTypes,
  fetchProviders,
  fetchSyncLog,
  fetchSyncLogItems,
  fetchSyncLogs,
  fetchSyncLogOperationLogs,
  fetchSyncLogRawResponse,
  fetchTags,
  importGames,
  launchGameDemo,
  removeCategoryGame,
  setGameCategories,
  setGameCollections,
  setGameTags,
  setProviderStatus,
  syncAllProviders,
  syncProvider,
  testProvider,
  updateCollection,
  updateCategory,
  updateCategoryGameSort,
  updateTag,
  updateLobbySection,
  updateGame,
  uploadGameAsset,
  type CollectionInput,
  type GameCenterDetail,
  type GameCenterQuery,
  type GameCenterRecord,
  type GameInput,
  type LobbySectionInput,
  type ProviderGameTypeOption,
  type TaxonomyInput,
} from "@/api/adminGames";
import {
  fetchAdminOptions,
  optionList,
  stringOptions,
  type AdminOption,
  type AdminOptions,
} from "@/api/adminOptions";
import { gameConfigs } from "./config";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";

type RecordValue = Record<string, unknown>;
type RelationKind = "categories" | "tags" | "collections";
type DetailPageState = { page: number; pageSize: number; total: number; loading: boolean };
const route = useRoute();
const router = useRouter();
const { t, te, locale } = useI18n();
const { actor, can, loadPermissions } = useAdminPermissions();
const moduleKey = computed(() => String(route.meta.moduleKey || "game-list"));
const config = computed(() => gameConfigs[moduleKey.value]);
const isPlatform = computed(() => actor.value?.platform === true);
const isGameListPage = computed(() => moduleKey.value === "game-list");
const isGameCategoriesPage = computed(() => moduleKey.value === "game-categories");
const isGameSyncPage = computed(() => moduleKey.value === "game-sync");
const isProviderModule = computed(() => moduleKey.value === "provider-management" || moduleKey.value === "game-sync");
const usesDraftWorkflow = computed(
  () =>
    !isPlatform.value &&
    ["game-list", "game-categories", "lobby-sections"].includes(moduleKey.value),
);
const visibleColumns = computed(() => {
  const columns = config.value?.columns.filter((item) => !isInternalIdentifierKey(item.key)) || [];
  if (moduleKey.value === "game-categories") {
    const categoryColumns = columns
      .filter((column) => !["code", "nameZh"].includes(column.key))
      .map((column) =>
        column.key === "nameEn"
          ? { ...column, labelKey: "gameCenter.fields.display", width: column.width || 180 }
          : column,
      );
    const insertIndex = Math.min(2, categoryColumns.length);
    return [
      ...categoryColumns.slice(0, insertIndex),
      { key: "ownership", labelKey: "gameCenter.fields.ownership", width: 120 },
      ...categoryColumns.slice(insertIndex),
    ];
  }
  if (!isPlatform.value || moduleKey.value !== "game-list") return columns;
  const hidden = new Set(["publishedState", "draftState", "hasChanges", "draftUpdatedAt"]);
  const directLabels: Record<string, string> = {
    platformEnabled: "gameCenter.fields.platformEnabled",
    isHot: "gameCenter.fields.hot",
    isFeatured: "gameCenter.fields.featured",
    isNew: "gameCenter.fields.new",
    sortOrder: "gameCenter.fields.sortOrder",
  };
  return columns
    .filter((column) => !hidden.has(column.key))
    .map((column) => ({ ...column, labelKey: directLabels[column.key] || column.labelKey }));
});
const visibleFilters = computed(() => {
  const filters = config.value?.filters.filter((item) => !isInternalIdentifierKey(item.key)) || [];
  if (moduleKey.value !== "game-categories") return filters;
  return filters
    .filter((filter) => !["code", "nameZh"].includes(filter.key))
    .map((filter) =>
      filter.key === "nameEn"
        ? { ...filter, labelKey: "gameCenter.fields.display" }
        : filter,
    );
});
const query = reactive<Record<string, string | string[]>>({});
const rows = ref<GameCenterRecord[]>([]);
const selectedRows = ref<GameCenterRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<GameCenterDetail>({});
const activeDetailTab = ref("");
const detailTargetID = ref<number | string>("");
const detailPageStates = reactive<Record<string, DetailPageState>>({});
const syncLogDetailVisible = ref(false);
const syncLogDetailLoading = ref(false);
const syncLogDetail = ref<GameCenterDetail>({});
const syncLogDetailID = ref(0);
const syncLogActiveTab = ref("summary");
const syncLogPageStates = reactive<Record<string, DetailPageState>>({});
const editVisible = ref(false);
const editLoading = ref(false);
const uploadLoading = ref(false);
const uploadTarget = ref<"gameThumbnail" | "taxonomyIcon" | "">("");
const editing = ref<GameCenterRecord | null>(null);
const reasonVisible = ref(false);
const reasonLoading = ref(false);
const reason = ref("");
const pendingAction = ref<null | (() => Promise<unknown>)>(null);
const reasonTitle = ref("");
const reasonSuccessMessage = ref("gameCenter.messages.actionSucceeded");
const relationVisible = ref(false);
const relationLoading = ref(false);
const relationKind = ref<RelationKind>("categories");
const relationIDs = ref<number[]>([]);
const relationGame = ref<GameCenterRecord | null>(null);
const categories = ref<GameCenterRecord[]>([]);
const tags = ref<GameCenterRecord[]>([]);
const collections = ref<GameCenterRecord[]>([]);
const adminOptions = ref<AdminOptions>({});
const filterProviderTypeOptions = ref<ProviderGameTypeOption[]>([]);
const mappingVisible = ref(false);
const mappingLoading = ref(false);
const mappingTypeLoading = ref(false);
const providerOptions = ref<GameCenterRecord[]>([]);
const providerTypeOptions = ref<ProviderGameTypeOption[]>([]);
const mappingTypeKey = ref("");
const mappingForm = reactive({
  providerCode: "",
  categoryIDs: [] as number[],
});
const addCategoryGamesVisible = ref(false);
const addCategoryGamesLoading = ref(false);
const candidateProvidersLoading = ref(false);
const candidateTypesLoading = ref(false);
const candidateGamesLoading = ref(false);
const candidateProviderOptions = ref<GameCenterRecord[]>([]);
const candidateTypeOptions = ref<ProviderGameTypeOption[]>([]);
const candidateGames = ref<GameCenterRecord[]>([]);
const selectedCandidateGames = ref<GameCenterRecord[]>([]);
const candidateKeyword = ref("");
const candidatePage = ref(1);
const candidatePageSize = ref(20);
const candidateTotal = ref(0);
const candidateForm = reactive({
  providerCode: "",
  typeKey: "",
});
const importGamesVisible = ref(false);
const importGamesLoading = ref(false);
const importCatalogLoading = ref(false);
const importCatalogProviderLoading = ref(false);
const importCatalogTypeLoading = ref(false);
const importCatalogRows = ref<GameCenterRecord[]>([]);
const importCatalogProviders = ref<GameCenterRecord[]>([]);
const importCatalogTypes = ref<ProviderGameTypeOption[]>([]);
const selectedImportGames = ref<GameCenterRecord[]>([]);
const importCatalogKeyword = ref("");
const importCatalogPage = ref(1);
const importCatalogPageSize = ref(20);
const importCatalogTotal = ref(0);
const importReason = ref("");
const importCatalogForm = reactive({
  providerCode: "",
  typeKey: "",
});
const batchCategoryVisible = ref(false);
const batchCategoryLoading = ref(false);
const batchCategoryForm = reactive({ ids: [] as number[], reason: "" });
type CategoryGamesState = {
  loading: boolean;
  saving: Record<number, boolean>;
  list: GameCenterRecord[];
  total: number;
  page: number;
  pageSize: number;
};
const expandedCategoryGames = reactive<Record<string, CategoryGamesState>>({});
const gameForm = reactive<GameInput>({
  name: "",
  localThumbnailUrl: "",
  platformEnabled: false,
  isHot: false,
  isFeatured: false,
  isNew: false,
  sortOrder: 0,
  supportedCountries: "",
  supportedCurrencies: "",
  supportedDevices: "",
  reason: "",
});
const taxonomyForm = reactive<TaxonomyInput>({
  code: "",
  nameZh: "",
  nameEn: "",
  icon: "",
  sortOrder: 0,
  enabled: true,
  reason: "",
});
const lobbyForm = reactive<LobbySectionInput>({
  sectionCode: "",
  sectionType: "game-list",
  titleZh: "",
  titleEn: "",
  displayStyle: "horizontal",
  sourceType: "flag",
  sourceRef: "",
  maxItems: 12,
  sortOrder: 0,
  enabled: true,
  supportedCountries: "",
  supportedCurrencies: "",
  config: {},
  reason: "",
});
const lobbyConfigText = ref("{}");
const publishingLobby = ref(false);
const brandOverview = ref<BrandOverview | null>(null);
const detailTabs = computed(() => {
  let tabs = config.value?.detailTabs || [];
  if (isPlatform.value && moduleKey.value === "game-list")
    tabs = tabs.filter((tab) => tab.key !== "draftConfiguration");
  if (
    moduleKey.value === "game-sync" ||
    moduleKey.value === "game-categories" ||
    (moduleKey.value === "sync-logs" && detail.value.summary)
  )
    return tabs;
  return tabs.filter((tab) => tab.key in detail.value);
});
const syncLogTabs = [
  "summary",
  "failedRecords",
  "changedGames",
  "rawProviderResponse",
  "operationLogs",
];
const createModule = computed(() =>
  ["game-categories", "game-tags", "game-collections", "lobby-sections"].includes(moduleKey.value),
);
const selectedGameIDs = computed(() =>
  selectedRows.value.map((item) => Number(item.id)).filter((id) => id > 0),
);
const hasSelectedGames = computed(
  () => moduleKey.value === "game-list" && selectedGameIDs.value.length > 0,
);
const managedPermissions = computed(
  () =>
    ({
      "game-categories": {
        create: "game:category:create",
        update: "game:category:update",
        delete: "game:category:delete",
      },
      "game-tags": {
        create: "game:tag:create",
        update: "game:tag:update",
        delete: "game:tag:delete",
      },
      "game-collections": {
        create: "game:collection:create",
        update: "game:collection:update",
        delete: "game:collection:delete",
      },
      "lobby-sections": {
        create: "lobby:update",
        update: "lobby:update",
        delete: "lobby:update",
      },
    })[moduleKey.value],
);
function fieldOptions(field: { key?: string; options?: string[]; source?: Parameters<typeof optionList>[1] }) {
  if (moduleKey.value === "game-list" && query.providerCode) {
    if (field.key === "providerCategory")
      return filterProviderTypeOptions.value
        .filter((item) => item.source === "providerCategory")
        .map((item) => ({ label: `${item.code} (${item.gameCount})`, value: item.code }));
    if (field.key === "gameType")
      return filterProviderTypeOptions.value
        .filter((item) => item.source === "gameType")
        .map((item) => ({ label: `${item.code} (${item.gameCount})`, value: item.code }));
  }
  const dynamic = optionList(adminOptions.value, field.source);
  return dynamic.length ? dynamic : stringOptions(field.options);
}
async function changeFilterProvider(providerCode: string) {
  delete query.providerCategory;
  delete query.gameType;
  filterProviderTypeOptions.value = [];
  if (!providerCode || moduleKey.value !== "game-list") return;
  try {
    filterProviderTypeOptions.value = await fetchProviderOptionGameTypes(providerCode);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.loadFailed"));
  }
}
function optionItemLabel(item: AdminOption) {
  return item.label && item.label !== item.value ? item.label : optionLabel(item.value);
}
function optionLabel(value: string) {
  if (value === "1" || value === "true") return t("common.enable");
  if (value === "0" || value === "2" || value === "3" || value === "false")
    return t("common.disable");
  const key = `gameCenter.values.${value.toLowerCase().replaceAll("-", "_")}`;
  return te(key) ? t(key) : value;
}
function hasExplicitValue(value: unknown) {
  return value !== null && value !== undefined && value !== "";
}
function gameDraftColumn(key: string) {
  return (
    {
      platformEnabled: "platformEnabledDraft",
      isHot: "isHotDraft",
      isFeatured: "isFeaturedDraft",
      isNew: "isNewDraft",
      sortOrder: "sortDraft",
      status: "statusDraft",
    } as Record<string, string>
  )[key];
}
function gameDraftValue(row: GameCenterRecord, key: string) {
  if (isPlatform.value) return row[key];
  const draftKey = gameDraftColumn(key);
  if (draftKey && hasExplicitValue(row[draftKey])) return row[draftKey];
  return row[key];
}
function gameDraftBoolean(row: GameCenterRecord, key: string) {
  return Boolean(gameDraftValue(row, key));
}
function gameDraftNumber(row: GameCenterRecord, key: string) {
  return Number(gameDraftValue(row, key) || 0);
}
function gameBooleanCellValue(row: GameCenterRecord, key: string) {
  if (
    moduleKey.value === "game-list" &&
    ["platformEnabled", "isHot", "isFeatured", "isNew"].includes(key)
  )
    return gameDraftBoolean(row, key);
  return Boolean(row[key]);
}
function parseDraftCategoryIDs(row: GameCenterRecord) {
  const raw = row.categoryIDsDraft;
  if (!hasExplicitValue(raw)) return null;
  if (Array.isArray(raw)) {
    return raw
      .map((item) => {
        if (isObject(item)) return Number(item.categoryId || item.categoryID || item.id);
        return Number(item);
      })
      .filter((id) => id > 0);
  }
  const text = String(raw).trim();
  if (!text) return null;
  try {
    const parsed = JSON.parse(text);
    if (Array.isArray(parsed))
      return parsed
        .map((item) => {
          if (isObject(item)) return Number(item.categoryId || item.categoryID || item.id);
          return Number(item);
        })
        .filter((id) => id > 0);
  } catch {
    return text
      .split(",")
      .map((item) => Number(item.trim()))
      .filter((id) => id > 0);
  }
  return [];
}
function valueChanged(published: unknown, draft: unknown) {
  return JSON.stringify(published ?? null) !== JSON.stringify(draft ?? null);
}
function gameDraftChangeCount(row: GameCenterRecord) {
  let count = 0;
  for (const key of ["platformEnabled", "isHot", "isFeatured", "isNew", "sortOrder"]) {
    const draftKey = gameDraftColumn(key);
    if (draftKey && hasExplicitValue(row[draftKey]) && valueChanged(row[key], row[draftKey]))
      count += 1;
  }
  const draftCategoryIDs = parseDraftCategoryIDs(row);
  if (
    String(row.configStatus || "").toUpperCase() === "DRAFT" &&
    draftCategoryIDs &&
    valueChanged(publishedCategoryIDs(row), draftCategoryIDs)
  )
    count += 1;
  return count;
}
function gameHasUnpublishedChanges(row: GameCenterRecord) {
  return (
    String(row.configStatus || "").toUpperCase() === "DRAFT" ||
    gameDraftChangeCount(row) > 0
  );
}
function gamePublishedState(row: GameCenterRecord) {
  return {
    enabled: Boolean(row.platformEnabled),
    sortOrder: Number(row.sortOrder || 0),
  };
}
function gameDraftState(row: GameCenterRecord) {
  return {
    enabled: gameDraftBoolean(row, "platformEnabled"),
    sortOrder: gameDraftNumber(row, "sortOrder"),
  };
}
function gameEnabledLabel(enabled: boolean) {
  return enabled ? t("common.enable") : t("common.disable");
}
function gameStateTagType(enabled: boolean) {
  return enabled ? "success" : "danger";
}
function gameConfigStatusType(row: GameCenterRecord) {
  return gameHasUnpublishedChanges(row) ? "warning" : "success";
}
function gameConfigStatusLabel(row: GameCenterRecord) {
  return gameHasUnpublishedChanges(row)
    ? t("gameCenter.values.unpublished_changes")
    : t("gameCenter.values.published");
}
const enumKeys = new Set([
  "apiStatus",
  "environment",
  "providerCode",
  "gameType",
  "providerCategoryCode",
  "launchType",
  "currency",
  "language",
  "country",
  "device",
  "changeType",
]);
function isEnumKey(key: string) {
  return key === "status" || enumKeys.has(key);
}
function categoryLabel(item: GameCenterRecord) {
  const preferred = locale.value === "zh-CN" ? item.nameZh : item.nameEn;
  return String(preferred || item.nameEn || item.nameZh || item.code || "-");
}
function relationLabel(item: GameCenterRecord) {
  const preferred = locale.value === "zh-CN" ? item.nameZh : item.nameEn;
  return String(preferred || item.nameEn || item.nameZh || item.code || "-");
}
function publishedCategoryIDs(row: GameCenterRecord) {
  const selected = new Set(
    String(row.platformCategories || "")
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean),
  );
  return categories.value
    .filter((item) => selected.has(String(item.code || "")))
    .map((item) => Number(item.id));
}
function categoryIDs(row: GameCenterRecord) {
  if (isPlatform.value) return publishedCategoryIDs(row);
  const draftIDs = parseDraftCategoryIDs(row);
  if (draftIDs) return draftIDs;
  return publishedCategoryIDs(row);
}
function imageSource(row: GameCenterRecord, key: string) {
  if (moduleKey.value === "game-list")
    return String(row.localThumbnailUrl || row.thumbnailUrl || "");
  return String(row[key] || "");
}
function categoryGameThumbnail(row: GameCenterRecord) {
  return String(row.localThumbnailUrl || row.thumbnailUrl || row.thumbnailPreview || "");
}
function detailLabel(key: string) {
  const normalized = key.replace(/_([a-z])/g, (_, letter: string) =>
    letter.toUpperCase(),
  );
  const path = `gameCenter.fields.${normalized}`;
  return te(path) ? t(path) : normalized;
}
function isObject(value: unknown): value is RecordValue {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function isResourceKey(key: string) {
  return (
    key === "icon" ||
    key === "thumbnailPreview" ||
    /(image|icon|logo|banner|asset|thumbnail|background|animation).*url$/i.test(
      key,
    )
  );
}
function isImageResourceKey(key: string) {
  return (
    key === "icon" ||
    key === "thumbnailPreview" ||
    /(image|icon|logo|banner|thumbnail|background).*url$/i.test(key)
  );
}
function hasDetailValue(value: unknown) {
  return value !== null && value !== undefined && value !== "";
}
function shouldShowDetailField(key: string, value: unknown) {
  if (moduleKey.value !== "game-list") return true;
  if (["providerID", "gameType", "supportsDemo"].includes(key)) return false;
  if (key === "providerCategoryCode" && !hasDetailValue(value)) return false;
  return true;
}
function detailEntries(value: unknown) {
  if (!isObject(value)) return [];
  return businessEntries(value)
    .filter(([key, item]) => shouldShowDetailField(key, item))
    .map(([key, item]) => ({ key, value: item }));
}
function detailCellValue(row: RecordValue, key: string) {
  if (key === "thumbnailPreview")
    return row.localThumbnailUrl || row.thumbnailUrl || "";
  return row[key];
}
function displayValue(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (isResourceKey(key)) return value ? t("gameCenter.messages.assetConfigured") : "-";
  if (key === "status" || key === "enabled") return optionLabel(String(value));
  if (typeof value === "boolean")
    return value ? t("common.yes") : t("common.no");
  if (/(At|Time|_at)$/.test(key) && typeof value === "string")
    return formatBackendTime(value);
  if (isObject(value) || Array.isArray(value))
    return JSON.stringify(value, null, 2);
  return String(value);
}
function tableCellValue(row: RecordValue, key: string) {
  if (moduleKey.value === "game-categories" && key === "nameEn")
    return row.nameEn || row.displayName || row.name || row.nameZh || "-";
  return row[key] ?? "-";
}
function isPlatformDefaultCategoryRecord(row: RecordValue) {
  if (row.isPlatformDefault === true || row.platformDefault === true) return true;
  return Number(row.tenantID || 0) === 0 && Number(row.siteID || 0) === 0;
}
function isPlatformDefaultCategory(row: RecordValue) {
  return moduleKey.value === "game-categories" && isPlatformDefaultCategoryRecord(row);
}
function categoryOwnershipLabel(row: RecordValue) {
  return t(isPlatformDefaultCategory(row) ? "gameCenter.values.platform_category" : "gameCenter.values.tenant_category");
}
function categoryOwnershipTagType(row: RecordValue) {
  return isPlatformDefaultCategory(row) ? "info" : "success";
}
const tableRows = computed(() => {
  return rows.value;
});
const tenantCategoryOptions = computed(() =>
  categories.value.filter((row) => !isPlatformDefaultCategoryRecord(row)),
);
const assignableCategoryOptions = computed(() =>
  isPlatform.value ? categories.value : tenantCategoryOptions.value,
);
function gameDetailConfigObject(key: string) {
  const value = detail.value[key];
  return isObject(value) ? value : {};
}
function gameDetailDraftDiffs() {
  const published = gameDetailConfigObject("platformConfiguration");
  const draft = gameDetailConfigObject("draftConfiguration");
  return [
    "platformEnabled",
    "isHot",
    "isFeatured",
    "isNew",
    "sortOrder",
    "localThumbnailUrl",
    "statusValue",
    "categoryDraft",
  ].map((key) => {
    const publishedValue =
      key === "statusValue"
        ? gameDetailConfigObject("overview").status
        : key === "categoryDraft"
          ? gameDetailConfigObject("categoriesAndTags").categories
          : published[key];
    const draftValue = draft[key] ?? publishedValue;
    return {
      key,
      published: publishedValue,
      draft: draftValue,
      changed: valueChanged(publishedValue, draftValue),
    };
  });
}
function gameDetailDraftMeta() {
  const draft = gameDetailConfigObject("draftConfiguration");
  return {
    status: String(draft.status || "PUBLISHED"),
    version: Number(draft.configVersion || 0),
    publishedAt: String(draft.publishedAt || ""),
    draftUpdatedAt: String(draft.draftUpdatedAt || ""),
  };
}
function isTimelineTab(key: string) {
  return key === "sessionTimeline";
}
function isTraceTab(key: string) {
  return key === "providerRequest" || key === "providerResponse" || key === "rawProviderResponse";
}
function traceDisplay(value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (typeof value === "string") {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}
function timelineItems(value: unknown) {
  return Array.isArray(value) ? value.filter(isObject) : [];
}
function timelineItemType(item: RecordValue) {
  const status = String(item.status || "").toUpperCase();
  if (status.includes("FAILED") || status.includes("ERROR")) return "danger";
  if (status === "LAUNCH_SUCCESS" || status === "PLAYING") return "success";
  if (status === "EXITED") return "warning";
  return "primary";
}
function timelineTime(item: RecordValue) {
  return formatBackendTime(String(item.time || ""));
}
function arrayColumns(value: unknown, tabKey = "") {
  if (!Array.isArray(value)) return [];
  const first = value.find(isObject);
  if (!first) return [];
  if (moduleKey.value === "game-categories" && tabKey === "games") {
    return [
      "thumbnailPreview",
      "gameCode",
      "name",
      "slug",
      "providerCode",
      "platformEnabled",
    ];
  }
  return businessColumns([first], 50)
        .filter(
          (key) =>
            !["deletedAt", "rawMetadata", "beforeState", "afterState"].includes(
              key,
            ) &&
            shouldShowDetailField(key, first[key]) &&
            !(
              key === "providerCategoryCode" &&
              !value.some(
                (row) => isObject(row) && hasDetailValue(row[key]),
              )
            ),
        )
        .slice(0, 12);
}
function buildParams() {
  const params: GameCenterQuery = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const field of config.value.filters) {
    const value = query[field.key];
    if (
      field.type === "dateRange" &&
      Array.isArray(value) &&
      value.length === 2
    ) {
      const prefix = field.key.replace(/Range$/, "");
      params[`${prefix}From`] = selectedTimeToBackend(value[0] || "");
      params[`${prefix}To`] = selectedTimeToBackend(value[1] || "", true);
    } else if (typeof value === "string" && value.trim())
      params[field.key] = value.trim();
  }
  return params;
}
async function load() {
  if (!config.value) return;
  loading.value = true;
  try {
    const result = await config.value.loader(buildParams());
    rows.value = result.list || [];
    total.value = result.total || 0;
    selectedRows.value = [];
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}
function search() {
  page.value = 1;
  void load();
}
function reset() {
  Object.keys(query).forEach((key) => delete query[key]);
  filterProviderTypeOptions.value = [];
  page.value = 1;
  void load();
}
function indexMethod(index: number) {
  return (page.value - 1) * pageSize.value + index + 1;
}
function categoryGamesKey(row: GameCenterRecord) {
  return String(row.id);
}
function categoryGamesState(row: GameCenterRecord): CategoryGamesState {
  const key = categoryGamesKey(row);
  if (!expandedCategoryGames[key]) {
    expandedCategoryGames[key] = {
      loading: false,
      saving: {},
      list: [],
      total: 0,
      page: 1,
      pageSize: 20,
    };
  }
  return expandedCategoryGames[key];
}
function categoryGameIndex(index: number) {
  const state = detailPageState("games");
  return (state.page - 1) * state.pageSize + index + 1;
}
function candidateGameIndex(index: number) {
  return (candidatePage.value - 1) * candidatePageSize.value + index + 1;
}
function importCatalogIndex(index: number) {
  return (importCatalogPage.value - 1) * importCatalogPageSize.value + index + 1;
}
function handleSelectionChange(selection: GameCenterRecord[]) {
  selectedRows.value = selection;
}
async function loadExpandedCategoryGames(row: GameCenterRecord) {
  if (isPlatformDefaultCategory(row)) return;
  const state = categoryGamesState(row);
  state.loading = true;
  try {
    const result = await fetchCategoryGames(row.id, {
      page: state.page,
      pageSize: state.pageSize,
    });
    state.list = result.list || [];
    state.total = result.total || 0;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.detailFailed"));
  } finally {
    state.loading = false;
  }
}
function handleCategoryExpand(row: GameCenterRecord, expandedRows: GameCenterRecord[]) {
  if (moduleKey.value !== "game-categories") return;
  if (isPlatformDefaultCategory(row)) return;
  const expanded = expandedRows.some((item) => Number(item.id) === Number(row.id));
  if (expanded && categoryGamesState(row).list.length === 0) void loadExpandedCategoryGames(row);
}
async function changeExpandedCategoryGamePage(row: GameCenterRecord, pageValue: number) {
  const state = categoryGamesState(row);
  state.page = pageValue;
  await loadExpandedCategoryGames(row);
}
async function changeExpandedCategoryGamePageSize(row: GameCenterRecord, pageSizeValue: number) {
  const state = categoryGamesState(row);
  state.page = 1;
  state.pageSize = pageSizeValue;
  await loadExpandedCategoryGames(row);
}
function editableCategoryGamePayload(row: GameCenterRecord): GameInput {
  return {
    ...gamePayload(row, "save_category_game_inline_draft"),
    name: String(row.name || "").trim(),
    platformEnabled: Boolean(row.platformEnabled),
    sortOrder: Number(row.sortOrder || 0),
  };
}
async function saveExpandedCategoryGame(category: GameCenterRecord, row: GameCenterRecord) {
  const state = categoryGamesState(category);
  state.saving[row.id] = true;
  try {
    await updateGame(row.id, editableCategoryGamePayload(row));
    await updateCategoryGameSort(
      category.id,
      row.id,
      Math.max(0, Number(row.categorySortOrder || 0)),
    );
    ElMessage.success(t(isPlatform.value ? "gameCenter.messages.savedImmediate" : "gameCenter.messages.draftSaved"));
    notifyBrandDraftChanged();
    await loadExpandedCategoryGames(category);
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.saveFailed"));
  } finally {
    state.saving[row.id] = false;
  }
}
async function removeExpandedCategoryGame(category: GameCenterRecord, row: GameCenterRecord) {
  const state = categoryGamesState(category);
  state.saving[row.id] = true;
  try {
    await ElMessageBox.confirm(
      t("gameCenter.dialogs.removeCategoryGameMessage", {
        game: String(row.name || row.gameCode || ""),
      }),
      t("gameCenter.dialogs.removeCategoryGameTitle"),
      { type: "warning" },
    );
    await removeCategoryGame(category.id, row.id);
    if (state.list.length === 1 && state.page > 1) state.page -= 1;
    ElMessage.success(t(isPlatform.value ? "gameCenter.messages.savedImmediate" : "gameCenter.messages.categoryDraftSaved"));
    notifyBrandDraftChanged();
    await loadExpandedCategoryGames(category);
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.actionFailed"));
  } finally {
    state.saving[row.id] = false;
  }
}
function detailID(row: GameCenterRecord) {
  return moduleKey.value === "provider-management" ||
    moduleKey.value === "game-sync"
    ? String(row.providerCode || "")
    : row.id;
}
async function openDetail(row: GameCenterRecord, initialTab = "") {
  if (isPlatformDefaultCategory(row)) {
    ElMessage.warning(t("gameCenter.messages.platformDefaultCategoryOperationDisabled"));
    return;
  }
  detailVisible.value = true;
  detailLoading.value = true;
  detail.value = {};
  detailTargetID.value = detailID(row);
  resetDetailPageStates();
  activeDetailTab.value = initialTab || config.value.detailTabs[0]?.key || "";
  try {
    detail.value = await config.value.detailLoader(detailTargetID.value);
    await ensureDetailTab(activeDetailTab.value);
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.detailFailed"),
    );
  } finally {
    detailLoading.value = false;
  }
}

function resetDetailPageStates() {
  Object.keys(detailPageStates).forEach((key) => delete detailPageStates[key]);
}
function detailPageState(tabKey: string) {
  if (!detailPageStates[tabKey]) {
    detailPageStates[tabKey] = {
      page: 1,
      pageSize: 20,
      total: 0,
      loading: false,
    };
  }
  return detailPageStates[tabKey];
}
function isLazyDetailTab(tabKey: string) {
  return (
    (moduleKey.value === "game-categories" &&
      ["games", "operationLogs"].includes(tabKey)) ||
    (moduleKey.value === "game-sync" && tabKey === "syncLogs") ||
    (moduleKey.value === "sync-logs" &&
      ["failedRecords", "changedGames", "rawProviderResponse", "operationLogs"].includes(tabKey))
  );
}
function isPagedDetailTab(tabKey: string) {
  return (
    (moduleKey.value === "game-categories" &&
      ["games", "operationLogs"].includes(tabKey)) ||
    (moduleKey.value === "game-sync" && tabKey === "syncLogs") ||
    (moduleKey.value === "sync-logs" &&
      ["failedRecords", "changedGames", "operationLogs"].includes(tabKey))
  );
}
async function ensureDetailTab(tabKey: string) {
  if (!isLazyDetailTab(tabKey) || !detailTargetID.value) return;
  if (detail.value[tabKey] !== undefined && !isPagedDetailTab(tabKey)) return;
  if (isPagedDetailTab(tabKey) && detail.value[tabKey] !== undefined) return;
  await loadDetailTab(tabKey);
}
async function loadDetailTab(tabKey: string) {
  if (!isLazyDetailTab(tabKey) || !detailTargetID.value) return;
  const id = Number(detailTargetID.value);
  const state = detailPageState(tabKey);
  state.loading = true;
  try {
    if (moduleKey.value === "game-categories" && tabKey === "games") {
      const result = await fetchCategoryGames(id, {
        page: state.page,
        pageSize: state.pageSize,
      });
      detail.value = { ...detail.value, games: result.list || [] };
      state.total = result.total || 0;
    } else if (
      moduleKey.value === "game-categories" &&
      tabKey === "operationLogs"
    ) {
      const result = await fetchCategoryOperationLogs(id, {
        page: state.page,
        pageSize: state.pageSize,
      });
      detail.value = { ...detail.value, operationLogs: result.list || [] };
      state.total = result.total || 0;
    } else if (moduleKey.value === "game-sync" && tabKey === "syncLogs") {
      const result = await fetchSyncLogs({
        providerCode: String(detailTargetID.value),
        page: state.page,
        pageSize: state.pageSize,
      });
      detail.value = { ...detail.value, syncLogs: result.list || [] };
      state.total = result.total || 0;
    } else if (tabKey === "failedRecords" || tabKey === "changedGames") {
      const result = await fetchSyncLogItems(
        id,
        tabKey === "failedRecords" ? "failed" : "changed",
        { page: state.page, pageSize: state.pageSize },
      );
      detail.value = { ...detail.value, [tabKey]: result.list || [] };
      state.total = result.total || 0;
    } else if (tabKey === "operationLogs") {
      const result = await fetchSyncLogOperationLogs(id, {
        page: state.page,
        pageSize: state.pageSize,
      });
      detail.value = { ...detail.value, [tabKey]: result.list || [] };
      state.total = result.total || 0;
    } else if (tabKey === "rawProviderResponse") {
      const result = await fetchSyncLogRawResponse(id);
      detail.value = { ...detail.value, [tabKey]: result };
    }
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.detailFailed"),
    );
  } finally {
    state.loading = false;
  }
}

function syncLogPageState(tabKey: string) {
  if (!syncLogPageStates[tabKey]) {
    syncLogPageStates[tabKey] = {
      page: 1,
      pageSize: 20,
      total: 0,
      loading: false,
    };
  }
  return syncLogPageStates[tabKey];
}
function resetSyncLogPageStates() {
  Object.keys(syncLogPageStates).forEach((key) => delete syncLogPageStates[key]);
}
function isLazySyncLogDetailTab(tabKey: string) {
  return ["failedRecords", "changedGames", "rawProviderResponse", "operationLogs"].includes(tabKey);
}
function isPagedSyncLogDetailTab(tabKey: string) {
  return ["failedRecords", "changedGames", "operationLogs"].includes(tabKey);
}
async function openSyncLogDetail(row: RecordValue) {
  const id = Number(row.id || 0);
  if (!id) return;
  syncLogDetailVisible.value = true;
  syncLogDetailLoading.value = true;
  syncLogDetail.value = {};
  syncLogDetailID.value = id;
  syncLogActiveTab.value = "summary";
  resetSyncLogPageStates();
  try {
    syncLogDetail.value = await fetchSyncLog(id);
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.detailFailed"),
    );
  } finally {
    syncLogDetailLoading.value = false;
  }
}
async function loadSyncLogDetailTab(tabKey: string) {
  if (!isLazySyncLogDetailTab(tabKey) || !syncLogDetailID.value) return;
  const state = syncLogPageState(tabKey);
  state.loading = true;
  try {
    if (tabKey === "failedRecords" || tabKey === "changedGames") {
      const result = await fetchSyncLogItems(
        syncLogDetailID.value,
        tabKey === "failedRecords" ? "failed" : "changed",
        { page: state.page, pageSize: state.pageSize },
      );
      syncLogDetail.value = {
        ...syncLogDetail.value,
        [tabKey]: result.list || [],
      };
      state.total = result.total || 0;
    } else if (tabKey === "operationLogs") {
      const result = await fetchSyncLogOperationLogs(syncLogDetailID.value, {
        page: state.page,
        pageSize: state.pageSize,
      });
      syncLogDetail.value = {
        ...syncLogDetail.value,
        operationLogs: result.list || [],
      };
      state.total = result.total || 0;
    } else {
      syncLogDetail.value = {
        ...syncLogDetail.value,
        rawProviderResponse: await fetchSyncLogRawResponse(syncLogDetailID.value),
      };
    }
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.detailFailed"),
    );
  } finally {
    state.loading = false;
  }
}
async function ensureSyncLogDetailTab(tabKey: string) {
  if (!isLazySyncLogDetailTab(tabKey)) return;
  if (syncLogDetail.value[tabKey] !== undefined) return;
  await loadSyncLogDetailTab(tabKey);
}
async function changeSyncLogTabPage(tabKey: string, value: number) {
  syncLogPageState(tabKey).page = value;
  await loadSyncLogDetailTab(tabKey);
}
async function changeSyncLogTabPageSize(tabKey: string, value: number) {
  const state = syncLogPageState(tabKey);
  state.pageSize = value;
  state.page = 1;
  await loadSyncLogDetailTab(tabKey);
}
async function changeDetailTabPage(tabKey: string, value: number) {
  const state = detailPageState(tabKey);
  state.page = value;
  await loadDetailTab(tabKey);
}
async function changeDetailTabPageSize(tabKey: string, value: number) {
  const state = detailPageState(tabKey);
  state.pageSize = value;
  state.page = 1;
  await loadDetailTab(tabKey);
}

function ask(
  titleKey: string,
  action: (reason: string) => Promise<unknown>,
  successMessage = "gameCenter.messages.actionSucceeded",
) {
  reason.value = "";
  reasonTitle.value = titleKey;
  reasonSuccessMessage.value = successMessage;
  pendingAction.value = () => action(reason.value.trim());
  reasonVisible.value = true;
}
function notifyBrandDraftChanged() {
  if (isPlatform.value) return;
  window.dispatchEvent(new Event("brand-config-refresh"));
}
async function confirmReason() {
  if (!pendingAction.value || !reason.value.trim()) return;
  reasonLoading.value = true;
  try {
    await pendingAction.value();
    reasonVisible.value = false;
    ElMessage.success(t(reasonSuccessMessage.value));
    notifyBrandDraftChanged();
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.actionFailed"),
    );
  } finally {
    reasonLoading.value = false;
  }
}
function toggleProvider(row: GameCenterRecord) {
  ask("gameCenter.dialogs.statusTitle", (value) =>
    setProviderStatus(
      String(row.providerCode),
      Number(row.status) !== 1,
      value,
    ),
  );
}
function toggleGame(row: GameCenterRecord) {
  updateGame(row.id, {
    ...gamePayload(row, isPlatform.value ? "toggle_display_status" : "toggle_display_status_draft"),
    platformEnabled: !gameDraftBoolean(row, "platformEnabled"),
  })
    .then(async () => {
      ElMessage.success(t(isPlatform.value ? "gameCenter.messages.savedImmediate" : "gameCenter.messages.draftSaved"));
      notifyBrandDraftChanged();
      await load();
    })
    .catch((error) => {
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("gameCenter.messages.saveFailed"),
      );
    });
}
function gamePayload(row: GameCenterRecord, reasonValue: string): GameInput {
  return {
    name: String(row.name || ""),
    localThumbnailUrl: String(row.localThumbnailUrl || ""),
    platformEnabled: gameDraftBoolean(row, "platformEnabled"),
    isHot: gameDraftBoolean(row, "isHot"),
    isFeatured: gameDraftBoolean(row, "isFeatured"),
    isNew: gameDraftBoolean(row, "isNew"),
    sortOrder: gameDraftNumber(row, "sortOrder"),
    supportedCountries: String(row.supportedCountries || ""),
    supportedCurrencies: String(row.supportedCurrencies || ""),
    supportedDevices: String(row.supportedDevices || ""),
    reason: reasonValue,
  };
}
function toggleGameFlag(
  row: GameCenterRecord,
  key: "isHot" | "isFeatured" | "isNew",
) {
  ask(
    "gameCenter.dialogs.statusTitle",
    async (value) => {
      const payload = gamePayload(row, value);
      payload[key] = !payload[key];
      await updateGame(row.id, payload);
    },
    isPlatform.value ? "gameCenter.messages.savedImmediate" : "gameCenter.messages.draftSaved",
  );
}
function changeBoolean(row: GameCenterRecord, key: string) {
  if (isPlatformDefaultCategory(row)) {
    ElMessage.warning(t("gameCenter.messages.platformDefaultCategoryOperationDisabled"));
    return;
  }
  if (moduleKey.value === "provider-management" || moduleKey.value === "game-sync") {
    if (key === "status" && can("game:provider:update")) toggleProvider(row);
    return;
  }
  if (moduleKey.value === "game-list") {
    if (key === "platformEnabled" && can("game:status:update")) toggleGame(row);
    else if (
      (key === "isHot" || key === "isFeatured" || key === "isNew") &&
      can("game:update")
    )
      toggleGameFlag(row, key);
    return;
  }
  if (key === "enabled" && managedPermissions.value?.update && can(managedPermissions.value.update))
    toggleManaged(row);
}
async function runBatchGameUpdate(
  field: "platformEnabled" | "isHot" | "isFeatured" | "isNew",
  value: boolean,
) {
  if (!selectedGameIDs.value.length) {
    ElMessage.warning(t("gameCenter.messages.selectGamesFirst"));
    return;
  }
  if (
    (field === "platformEnabled" && !can("game:status:update")) ||
    (field !== "platformEnabled" && !can("game:update"))
  )
    return;
  try {
    await ElMessageBox.confirm(
      t("gameCenter.dialogs.batchMessage", {
        count: selectedGameIDs.value.length,
        action: t(`gameCenter.actions.batch.${field}.${value ? "on" : "off"}`),
      }),
      t("gameCenter.dialogs.batchTitle"),
      {
        confirmButtonText: t("common.confirm"),
        cancelButtonText: t("common.cancel"),
        type: "warning",
      },
    );
    const result = await batchUpdateGames({
      ids: selectedGameIDs.value,
      [field]: value,
      reason: `batch_${field}_${value ? "on" : "off"}`,
    });
    ElMessage.success(
      t(isPlatform.value ? "gameCenter.messages.batchUpdatedImmediate" : "gameCenter.messages.batchDraftUpdated", {
        count: Number(result.updatedCount || selectedGameIDs.value.length),
      }),
    );
    notifyBrandDraftChanged();
    selectedRows.value = [];
    await load();
  } catch (error) {
    if (error === "cancel" || error === "close") return;
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.saveFailed"),
    );
  }
}
function openBatchCategory() {
  if (!selectedGameIDs.value.length || !can("game:category:assign")) return;
  batchCategoryForm.ids = [];
  batchCategoryForm.reason = "";
  batchCategoryVisible.value = true;
}
async function saveBatchCategory() {
  if (!selectedGameIDs.value.length || !batchCategoryForm.ids.length || !batchCategoryForm.reason.trim()) return;
  batchCategoryLoading.value = true;
  try {
    const result = await batchUpdateGames({
      ids: selectedGameIDs.value,
      addCategoryIds: batchCategoryForm.ids,
      reason: batchCategoryForm.reason.trim(),
    });
    ElMessage.success(
      t(isPlatform.value ? "gameCenter.messages.batchCategoryImmediate" : "gameCenter.messages.batchCategoryDraft", {
        count: Number(result.updatedCount || selectedGameIDs.value.length),
      }),
    );
    batchCategoryVisible.value = false;
    notifyBrandDraftChanged();
    selectedRows.value = [];
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.saveFailed"));
  } finally {
    batchCategoryLoading.value = false;
  }
}
function updateInlineCategories(row: GameCenterRecord, value: unknown) {
  if (!can("game:category:assign") || !Array.isArray(value)) return;
  const ids = value.map(Number).filter((id) => id > 0);
  setGameCategories(row.id, ids)
    .then(async () => {
      ElMessage.success(t(isPlatform.value ? "gameCenter.messages.savedImmediate" : "gameCenter.messages.draftSaved"));
      notifyBrandDraftChanged();
      await load();
    })
    .catch((error) => {
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("gameCenter.messages.saveFailed"),
      );
    });
}
function providerOptionLabel(item: GameCenterRecord) {
  return `${String(item.providerCode || "")} · ${String(item.providerName || item.name || "")}`;
}
function providerTypeKey(item: ProviderGameTypeOption) {
  return `${item.source}::${item.code}`;
}
function providerTypeLabel(item: ProviderGameTypeOption) {
  const sourceKey = `gameCenter.values.${item.source}`;
  const sourceLabel = te(sourceKey) ? t(sourceKey) : item.source;
  return `${item.code} · ${sourceLabel} · ${item.gameCount}`;
}
function selectedProviderType() {
  const [source, ...codeParts] = mappingTypeKey.value.split("::");
  const code = codeParts.join("::");
  if ((source !== "providerCategory" && source !== "gameType") || !code)
    return null;
  return { source, code } as const;
}
async function loadMappingProviders() {
  const result = await fetchProviders({ page: 1, pageSize: 100 });
  providerOptions.value = result.list || [];
}
async function loadMappingProviderTypes(providerCode: string) {
  mappingTypeKey.value = "";
  providerTypeOptions.value = [];
  if (!providerCode) return;
  mappingTypeLoading.value = true;
  try {
    providerTypeOptions.value = await fetchProviderGameTypes(providerCode);
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.loadFailed"),
    );
  } finally {
    mappingTypeLoading.value = false;
  }
}
async function openCategoryMapping() {
  mappingVisible.value = true;
  mappingLoading.value = true;
  mappingForm.providerCode = "";
  mappingTypeKey.value = "";
  mappingForm.categoryIDs = [];
  providerTypeOptions.value = [];
  try {
    await Promise.all([loadMappingProviders(), loadRelationOptions()]);
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.loadFailed"),
    );
  } finally {
    mappingLoading.value = false;
  }
}
async function saveCategoryMapping() {
  const selected = selectedProviderType();
  if (!validateRequiredValue(mappingForm.providerCode, "gameCenter.fields.provider"))
    return;
  if (!selected) {
    warnRequired("gameCenter.fields.providerGameType");
    return;
  }
  if (!mappingForm.categoryIDs.length) {
    warnRequired("gameCenter.fields.platformCategories");
    return;
  }
  mappingLoading.value = true;
  try {
    const result = await applyProviderCategoryMapping({
      providerCode: mappingForm.providerCode,
      source: selected.source,
      code: selected.code,
      categoryIds: mappingForm.categoryIDs,
      reason: "provider_type_category_mapping",
    });
    mappingVisible.value = false;
    ElMessage.success(
      t(isPlatform.value ? "gameCenter.messages.categoryMappingImmediate" : "gameCenter.messages.categoryMappingDraftSaved", {
        count: Number(result.updatedCount || 0),
      }),
    );
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error && error.message === "INVALID_LOBBY_CONFIG"
        ? t("gameCenter.validation.invalidJson")
        : error instanceof Error
        ? error.message
        : t("gameCenter.messages.saveFailed"),
    );
  } finally {
    mappingLoading.value = false;
  }
}
async function runTest(row: GameCenterRecord) {
  try {
    await testProvider(String(row.providerCode));
    ElMessage.success(t("gameCenter.messages.testSucceeded"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.testFailed"),
    );
  }
}
function runSync(row?: GameCenterRecord) {
  ask(
    row ? "gameCenter.dialogs.syncTitle" : "gameCenter.dialogs.syncAllTitle",
    (value) =>
      row
        ? syncProvider(String(row.providerCode), value)
        : syncAllProviders(value),
    "gameCenter.messages.syncQueued",
  );
}
function runCancelSync(row: GameCenterRecord) {
  ask(
    "gameCenter.dialogs.cancelSyncTitle",
    (value) => cancelProviderSync(String(row.providerCode), value),
    "gameCenter.messages.syncCancelled",
  );
}
async function runDeleteProvider(row: GameCenterRecord) {
  try {
    await ElMessageBox.confirm(
      t("gameCenter.dialogs.deleteProviderMessage", {
        provider: String(row.providerName || row.providerCode || ""),
      }),
      t("gameCenter.dialogs.deleteProviderTitle"),
      { type: "warning" },
    );
  } catch {
    return;
  }
  ask(
    "gameCenter.dialogs.deleteProviderReasonTitle",
    (value) => deleteProvider(String(row.providerCode), value),
    "gameCenter.messages.providerDeleted",
  );
}
async function runDemo(row: GameCenterRecord) {
  try {
    const result = await launchGameDemo(row.id);
    window.open(result.launchUrl, "_blank", "noopener,noreferrer");
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.demoFailed"),
    );
  }
}

async function openEdit(row?: GameCenterRecord) {
  if (row && isPlatformDefaultCategory(row)) {
    ElMessage.warning(t("gameCenter.messages.platformDefaultCategoryOperationDisabled"));
    return;
  }
  editing.value = row || null;
  editVisible.value = true;
  if (moduleKey.value === "game-list" && row) {
    editLoading.value = true;
    try {
      const data = await config.value.detailLoader(row.id);
      const draft = isObject(data.draftConfiguration)
        ? data.draftConfiguration
        : {};
      const published = isObject(data.platformConfiguration)
        ? data.platformConfiguration
        : {};
      const rules = isObject(data.countryAndCurrencyRules)
        ? data.countryAndCurrencyRules
        : {};
      const editable = isPlatform.value ? published : draft;
      Object.assign(gameForm, {
        name: String(editable.name || row.name || ""),
        localThumbnailUrl: String(
          editable.localThumbnailUrl ||
            published.localThumbnailUrl ||
            row.localThumbnailUrl ||
            "",
        ),
        platformEnabled: Boolean(
          hasExplicitValue(editable.platformEnabled)
            ? editable.platformEnabled
            : gameDraftValue(row, "platformEnabled"),
        ),
        isHot: Boolean(
          hasExplicitValue(editable.isHot) ? editable.isHot : gameDraftValue(row, "isHot"),
        ),
        isFeatured: Boolean(
          hasExplicitValue(editable.isFeatured)
            ? editable.isFeatured
            : gameDraftValue(row, "isFeatured"),
        ),
        isNew: Boolean(
          hasExplicitValue(editable.isNew) ? editable.isNew : gameDraftValue(row, "isNew"),
        ),
        sortOrder: Number(
          hasExplicitValue(editable.sortOrder)
            ? editable.sortOrder
            : gameDraftValue(row, "sortOrder") || 0,
        ),
        supportedCountries: String(rules.countries || row.supportedCountries || ""),
        supportedCurrencies: String(rules.currencies || row.supportedCurrencies || ""),
        supportedDevices: String(rules.devices || row.supportedDevices || ""),
        reason: "",
      });
    } catch (error) {
      editVisible.value = false;
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("gameCenter.messages.detailFailed"),
      );
    } finally {
      editLoading.value = false;
    }
    return;
  }
  if (moduleKey.value === "lobby-sections") {
    const source: RecordValue = row || {};
    Object.assign(lobbyForm, {
      sectionCode: String(source.sectionCode || ""),
      sectionType: String(source.sectionType || "game-list"),
      titleZh: String(source.titleZh || ""),
      titleEn: String(source.titleEn || ""),
      displayStyle: String(source.displayStyle || "horizontal"),
      sourceType: String(source.sourceType || "flag"),
      sourceRef: String(source.sourceRef || ""),
      maxItems: Number(source.maxItems || 12),
      sortOrder: Number(source.sortOrder || 0),
      enabled: row ? Boolean(source.enabled) : true,
      supportedCountries: String(source.supportedCountries || ""),
      supportedCurrencies: String(source.supportedCurrencies || ""),
      config: parseConfigValue(source.configJSON || source.configJson || "{}"),
      reason: "",
    });
    lobbyConfigText.value = JSON.stringify(lobbyForm.config || {}, null, 2);
    return;
  }
  const source: RecordValue = row || {};
  const displayName = String(source.nameEn || source.displayName || source.name || source.nameZh || "");
  Object.assign(taxonomyForm, {
    code: String(source.code || ""),
    nameZh: String(source.nameZh || displayName),
    nameEn: displayName,
    icon: String(source.icon || source.bannerUrl || ""),
    sortOrder: Number(source.sortOrder || 0),
    enabled: row ? Boolean(source.enabled) : true,
    reason: "",
  });
}
function slugifyDisplay(value: string) {
  const slug = value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return slug || `category-${Date.now()}`;
}
function taxonomyPayload(
  row: RecordValue,
  reasonValue: string,
  enabled: boolean,
): TaxonomyInput {
  if (moduleKey.value === "game-categories") {
    const display = String(row.nameEn || row.displayName || row.name || row.nameZh || "").trim();
    return {
      code: String(row.code || "").trim() || slugifyDisplay(display),
      nameZh: String(row.nameZh || "").trim() || display,
      nameEn: display,
      icon: String(row.icon || ""),
      sortOrder: Number(row.sortOrder || 0),
      enabled,
      reason: reasonValue,
    };
  }
  return {
    code: String(row.code || ""),
    nameZh: String(row.nameZh || ""),
    nameEn: String(row.nameEn || ""),
    icon: String(row.icon || ""),
    sortOrder: Number(row.sortOrder || 0),
    enabled,
    reason: reasonValue,
  };
}
function collectionPayload(
  row: RecordValue,
  reasonValue: string,
  enabled: boolean,
): CollectionInput {
  return {
    code: String(row.code || ""),
    nameZh: String(row.nameZh || ""),
    nameEn: String(row.nameEn || ""),
    descriptionZh: String(row.descriptionZh || ""),
    descriptionEn: String(row.descriptionEn || ""),
    bannerUrl: String(row.icon || row.bannerUrl || ""),
    sortOrder: Number(row.sortOrder || 0),
    enabled,
    reason: reasonValue,
  };
}
function parseConfigValue(value: unknown): Record<string, unknown> {
  if (isObject(value)) return value;
  try {
    const parsed = JSON.parse(String(value || "{}"));
    return isObject(parsed) ? parsed : {};
  } catch {
    return {};
  }
}
function lobbyPayload(reasonValue: string): LobbySectionInput {
  let configValue: Record<string, unknown> = {};
  try {
    const parsed = JSON.parse(lobbyConfigText.value || "{}");
    configValue = isObject(parsed) ? parsed : {};
  } catch {
    throw new Error("INVALID_LOBBY_CONFIG");
  }
  return {
    ...lobbyForm,
    sectionCode: String(lobbyForm.sectionCode || "").trim(),
    sectionType: String(lobbyForm.sectionType || "").trim(),
    titleZh: String(lobbyForm.titleZh || "").trim(),
    titleEn: String(lobbyForm.titleEn || "").trim(),
    displayStyle: String(lobbyForm.displayStyle || "").trim(),
    sourceType: String(lobbyForm.sourceType || "").trim(),
    sourceRef: String(lobbyForm.sourceRef || "").trim(),
    maxItems: Number(lobbyForm.maxItems || 12),
    sortOrder: Number(lobbyForm.sortOrder || 0),
    supportedCountries: String(lobbyForm.supportedCountries || "").trim(),
    supportedCurrencies: String(lobbyForm.supportedCurrencies || "").trim(),
    config: configValue,
    reason: reasonValue,
  };
}

async function publishLobbyConfiguration() {
  try {
    publishingLobby.value = true;
    brandOverview.value = await systemAPI.brand.overview();
    if (!brandOverview.value?.siteID) throw new Error(t("brandPortal.messages.previewFailed"));
    const result = await ElMessageBox.prompt(
      t("brandPortal.prompts.publishReason"),
      t("brandPortal.dialogs.publishTitle"),
      {
        confirmButtonText: t("brandPortal.actions.publish"),
        cancelButtonText: t("brandPortal.actions.cancel"),
        inputValue: t("brandPortal.defaults.publishBrandConfig"),
        inputValidator: (value) => Boolean(value.trim()) || t("brandPortal.validation.reasonRequired"),
      },
    );
    await systemAPI.sites.publish(brandOverview.value.siteID, result.value.trim());
    ElMessage.success(t("brandPortal.messages.publishedToH5"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close") {
      ElMessage.error(error instanceof Error ? error.message : t("brandPortal.messages.publishFailed"));
    }
  } finally {
    publishingLobby.value = false;
  }
}
function warnRequired(labelKey: string) {
  ElMessage.warning(
    t("gameCenter.validation.required", { field: t(labelKey) }),
  );
}
function validateRequiredValue(value: unknown, labelKey: string) {
  if (String(value ?? "").trim() !== "") return true;
  warnRequired(labelKey);
  return false;
}
function validateEditForm() {
  if (moduleKey.value === "game-list") {
    return (
      validateRequiredValue(gameForm.name, "gameCenter.fields.gameName") &&
      validateRequiredValue(gameForm.reason, "gameCenter.fields.reason")
    );
  }
  if (moduleKey.value === "lobby-sections") {
    return (
      validateRequiredValue(lobbyForm.sectionCode, "gameCenter.fields.sectionCode") &&
      validateRequiredValue(lobbyForm.sectionType, "gameCenter.fields.sectionType")
    );
  }
  if (moduleKey.value === "game-categories")
    return validateRequiredValue(taxonomyForm.nameEn, "gameCenter.fields.display");
  return (
    validateRequiredValue(taxonomyForm.code, "gameCenter.fields.code") &&
    validateRequiredValue(taxonomyForm.nameZh, "gameCenter.fields.nameZh") &&
    validateRequiredValue(taxonomyForm.nameEn, "gameCenter.fields.nameEn")
  );
}
function toggleManaged(row: GameCenterRecord) {
  if (isPlatformDefaultCategory(row)) {
    ElMessage.warning(t("gameCenter.messages.platformDefaultCategoryOperationDisabled"));
    return;
  }
  const enabled = !Boolean(row.enabled);
  const action =
    moduleKey.value === "lobby-sections"
      ? updateLobbySection(row.id, {
          sectionCode: String(row.sectionCode || ""),
          sectionType: String(row.sectionType || "game-list"),
          titleZh: String(row.titleZh || ""),
          titleEn: String(row.titleEn || ""),
          displayStyle: String(row.displayStyle || "horizontal"),
          sourceType: String(row.sourceType || "flag"),
          sourceRef: String(row.sourceRef || ""),
          maxItems: Number(row.maxItems || 12),
          sortOrder: Number(row.sortOrder || 0),
          enabled,
          supportedCountries: String(row.supportedCountries || ""),
          supportedCurrencies: String(row.supportedCurrencies || ""),
          config: parseConfigValue(row.configJSON || row.configJson || "{}"),
          reason: "toggle_lobby_section_draft",
        })
      : moduleKey.value === "game-tags"
      ? updateTag(row.id, taxonomyPayload(row, "toggle_tag_status", enabled))
      : moduleKey.value === "game-collections"
        ? updateCollection(row.id, collectionPayload(row, "toggle_collection_status", enabled))
        : updateCategory(row.id, taxonomyPayload(row, "toggle_category_status", enabled));
  action
    .then(async () => {
      ElMessage.success(
        t(
          moduleKey.value === "game-categories"
            ? isPlatform.value
              ? "gameCenter.messages.savedImmediate"
              : "gameCenter.messages.categoryDraftSaved"
            : moduleKey.value === "lobby-sections"
              ? "gameCenter.messages.lobbyDraftSaved"
            : "gameCenter.messages.saved",
        ),
      );
      notifyBrandDraftChanged();
      await load();
    })
    .catch((error) => {
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("gameCenter.messages.saveFailed"),
      );
    });
}
async function saveEdit() {
  if (!validateEditForm()) return;
  editLoading.value = true;
  try {
    if (moduleKey.value === "game-list" && editing.value)
      await updateGame(editing.value.id, { ...gameForm });
    else if (moduleKey.value === "game-categories" || moduleKey.value === "game-tags") {
      const reason =
        taxonomyForm.reason || (moduleKey.value === "game-tags" ? "save_game_tag" : "save_game_category");
      const payload =
        moduleKey.value === "game-categories"
          ? taxonomyPayload(taxonomyForm as unknown as RecordValue, reason, taxonomyForm.enabled)
          : { ...taxonomyForm, reason };
      if (moduleKey.value === "game-tags") {
        if (editing.value) await updateTag(editing.value.id, payload);
        else await createTag(payload);
      } else if (editing.value) {
        await updateCategory(editing.value.id, payload);
      } else {
        await createCategory(payload);
      }
    } else if (moduleKey.value === "game-collections") {
      const payload = collectionPayload(
        taxonomyForm as unknown as RecordValue,
        taxonomyForm.reason || "save_game_collection",
        taxonomyForm.enabled,
      );
      if (editing.value) await updateCollection(editing.value.id, payload);
      else await createCollection(payload);
    } else if (moduleKey.value === "lobby-sections") {
      const payload = lobbyPayload(lobbyForm.reason || "save_lobby_section_draft");
      if (editing.value) await updateLobbySection(editing.value.id, payload);
      else await createLobbySection(payload);
    }
    editVisible.value = false;
    ElMessage.success(
      t(
          moduleKey.value === "game-list"
          ? isPlatform.value
            ? "gameCenter.messages.savedImmediate"
            : "gameCenter.messages.draftSaved"
          : moduleKey.value === "game-categories"
            ? isPlatform.value
              ? "gameCenter.messages.savedImmediate"
              : "gameCenter.messages.categoryDraftSaved"
            : moduleKey.value === "lobby-sections"
              ? "gameCenter.messages.lobbyDraftSaved"
            : "gameCenter.messages.saved",
      ),
    );
    notifyBrandDraftChanged();
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.saveFailed"),
    );
  } finally {
    editLoading.value = false;
  }
}
function setAssetValue(
  target: "gameThumbnail" | "taxonomyIcon",
  value: string,
) {
  if (target === "gameThumbnail") gameForm.localThumbnailUrl = value;
  else taxonomyForm.icon = value;
}
async function uploadGameAssetTo(
  target: "gameThumbnail" | "taxonomyIcon",
  options: UploadRequestOptions,
) {
  uploadTarget.value = target;
  uploadLoading.value = true;
  try {
    const asset = await uploadGameAsset(options.file);
    setAssetValue(target, asset.url);
    options.onSuccess(asset);
  } catch (error) {
    const uploadError =
      error instanceof Error
        ? error
        : new Error(t("gameCenter.messages.uploadFailed"));
    options.onError(uploadError as Parameters<UploadRequestOptions["onError"]>[0]);
    ElMessage.error(uploadError.message || t("gameCenter.messages.uploadFailed"));
  } finally {
    uploadLoading.value = false;
    uploadTarget.value = "";
  }
}
const uploadGameThumbnail = (options: UploadRequestOptions) =>
  uploadGameAssetTo("gameThumbnail", options);
const uploadTaxonomyIcon = (options: UploadRequestOptions) =>
  uploadGameAssetTo("taxonomyIcon", options);
async function removeRow(row: GameCenterRecord) {
  if (isPlatformDefaultCategory(row)) {
    ElMessage.warning(t("gameCenter.messages.platformDefaultCategoryDeleteDisabled"));
    return;
  }
  try {
    const result = await ElMessageBox.prompt(
      t("gameCenter.dialogs.deleteMessage"),
      t("gameCenter.dialogs.deleteTitle"),
      {
        confirmButtonText: t("common.confirm"),
        cancelButtonText: t("common.cancel"),
        inputPlaceholder: t("gameCenter.placeholders.reason"),
        inputValidator: (value) =>
          Boolean(String(value).trim()) ||
          t("gameCenter.validation.reasonRequired"),
      },
    );
    if (moduleKey.value === "game-tags") await deleteTag(row.id, result.value);
    else if (moduleKey.value === "game-collections") await deleteCollection(row.id, result.value);
    else if (moduleKey.value === "lobby-sections") await deleteLobbySection(row.id, result.value);
    else await deleteCategory(row.id, result.value);
    ElMessage.success(t("gameCenter.messages.deleted"));
    notifyBrandDraftChanged();
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("gameCenter.messages.actionFailed"),
      );
  }
}

function candidateTypeKey(option: ProviderGameTypeOption) {
  return `${option.source}:${option.code}`;
}
function selectedCandidateType() {
  return candidateTypeOptions.value.find(
    (option) => candidateTypeKey(option) === candidateForm.typeKey,
  );
}
function mergeCandidateTypeOptions(
  tenantOptions: ProviderGameTypeOption[],
  catalogOptions: ProviderGameTypeOption[],
) {
  const merged = new Map<string, ProviderGameTypeOption>();
  for (const option of catalogOptions) {
    merged.set(candidateTypeKey(option), option);
  }
  for (const option of tenantOptions) {
    // Prefer the tenant count because the candidate list only contains
    // games already imported into the current tenant.
    merged.set(candidateTypeKey(option), option);
  }
  return Array.from(merged.values()).sort((left, right) => {
    if (left.source !== right.source) return left.source.localeCompare(right.source);
    return left.code.localeCompare(right.code);
  });
}
async function openAddCategoryGames() {
  if (!detailTargetID.value) return;
  addCategoryGamesVisible.value = true;
  candidateForm.providerCode = "";
  candidateForm.typeKey = "";
  candidateKeyword.value = "";
  candidateGames.value = [];
  selectedCandidateGames.value = [];
  candidateTypeOptions.value = [];
  candidatePage.value = 1;
  candidateTotal.value = 0;
  candidateProvidersLoading.value = true;
  try {
    let options = optionList(adminOptions.value, "providers");
    if (!options.length) {
      adminOptions.value = await fetchAdminOptions();
      options = optionList(adminOptions.value, "providers");
    }
    candidateProviderOptions.value = options.map((item, index) => ({
      id: index + 1,
      providerCode: item.value,
      providerName: item.label,
    }));
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.loadFailed"),
    );
  } finally {
    candidateProvidersLoading.value = false;
  }
}
async function changeCandidateProvider() {
  candidateForm.typeKey = "";
  candidateTypeOptions.value = [];
  candidateGames.value = [];
  selectedCandidateGames.value = [];
  candidateTotal.value = 0;
  if (!candidateForm.providerCode) return;
  candidateTypesLoading.value = true;
  try {
    const results = await Promise.allSettled([
      fetchProviderOptionGameTypes(candidateForm.providerCode),
      fetchImportableProviderGameTypes(candidateForm.providerCode),
    ]);
    const tenantOptions =
      results[0].status === "fulfilled" ? results[0].value : [];
    const catalogOptions =
      results[1].status === "fulfilled" ? results[1].value : [];
    const tenantResult = results[0];
    const catalogResult = results[1];
    if (
      tenantResult.status === "rejected" &&
      catalogResult.status === "rejected"
    ) {
      throw tenantResult.reason;
    }
    candidateTypeOptions.value = mergeCandidateTypeOptions(
      tenantOptions,
      catalogOptions,
    );
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.loadFailed"),
    );
  } finally {
    candidateTypesLoading.value = false;
  }
}
async function loadCandidateGames() {
  const type = selectedCandidateType();
  if (!detailTargetID.value || !candidateForm.providerCode || !type) {
    candidateGames.value = [];
    candidateTotal.value = 0;
    return;
  }
  candidateGamesLoading.value = true;
  try {
    const result = await fetchCategoryGameCandidates(
      Number(detailTargetID.value),
      {
        page: candidatePage.value,
        pageSize: candidatePageSize.value,
        providerCode: candidateForm.providerCode,
        source: type.source,
        categoryCode: type.code,
        keyword: candidateKeyword.value.trim() || undefined,
      },
    );
    candidateGames.value = result.list || [];
    candidateTotal.value = result.total || 0;
    selectedCandidateGames.value = [];
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.loadFailed"),
    );
  } finally {
    candidateGamesLoading.value = false;
  }
}
function changeCandidateType() {
  candidatePage.value = 1;
  candidateKeyword.value = "";
  void loadCandidateGames();
}
function searchCandidateGames() {
  candidatePage.value = 1;
  void loadCandidateGames();
}
function changeCandidateSelection(selection: GameCenterRecord[]) {
  selectedCandidateGames.value = selection;
}
async function saveAddedCategoryGames() {
  if (!detailTargetID.value || selectedCandidateGames.value.length === 0) {
    ElMessage.warning(t("gameCenter.validation.selectGames"));
    return;
  }
  addCategoryGamesLoading.value = true;
  try {
    await addCategoryGames(
      Number(detailTargetID.value),
      selectedCandidateGames.value.map((game) => Number(game.id)),
    );
    addCategoryGamesVisible.value = false;
    ElMessage.success(t(isPlatform.value ? "gameCenter.messages.savedImmediate" : "gameCenter.messages.categoryDraftSaved"));
    notifyBrandDraftChanged();
    await loadDetailTab("games");
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.saveFailed"),
    );
  } finally {
    addCategoryGamesLoading.value = false;
  }
}
function importCatalogTypeKey(option: ProviderGameTypeOption) {
  return `${option.source}:${option.code}`;
}
function selectedImportCatalogType() {
  return importCatalogTypes.value.find(
    (option) => importCatalogTypeKey(option) === importCatalogForm.typeKey,
  );
}
async function openImportGames() {
  importGamesVisible.value = true;
  importCatalogForm.providerCode = "";
  importCatalogForm.typeKey = "";
  importCatalogKeyword.value = "";
  importCatalogRows.value = [];
  importCatalogTypes.value = [];
  selectedImportGames.value = [];
  importCatalogPage.value = 1;
  importCatalogTotal.value = 0;
  importReason.value = t("gameCenter.defaults.importGamesReason");
  importCatalogProviderLoading.value = true;
  try {
    let options = optionList(adminOptions.value, "providers");
    if (!options.length) {
      adminOptions.value = await fetchAdminOptions();
      options = optionList(adminOptions.value, "providers");
    }
    importCatalogProviders.value = options.map((item, index) => ({
      id: index + 1,
      providerCode: item.value,
      providerName: item.label,
    }));
    await loadImportCatalogGames();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.loadFailed"));
  } finally {
    importCatalogProviderLoading.value = false;
  }
}
async function changeImportCatalogProvider() {
  importCatalogForm.typeKey = "";
  importCatalogTypes.value = [];
  importCatalogRows.value = [];
  selectedImportGames.value = [];
  importCatalogTotal.value = 0;
  importCatalogPage.value = 1;
  if (importCatalogForm.providerCode) {
    importCatalogTypeLoading.value = true;
    try {
      importCatalogTypes.value = await fetchImportableProviderGameTypes(
        importCatalogForm.providerCode,
      );
    } catch (error) {
      ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.loadFailed"));
    } finally {
      importCatalogTypeLoading.value = false;
    }
  }
  await loadImportCatalogGames();
}
function changeImportCatalogType() {
  importCatalogPage.value = 1;
  void loadImportCatalogGames();
}
function searchImportCatalogGames() {
  importCatalogPage.value = 1;
  void loadImportCatalogGames();
}
async function loadImportCatalogGames() {
  importCatalogLoading.value = true;
  const selectedType = selectedImportCatalogType();
  try {
    const result = await fetchImportableGames({
      page: importCatalogPage.value,
      pageSize: importCatalogPageSize.value,
      providerCode: importCatalogForm.providerCode || undefined,
      providerCategory:
        selectedType?.source === "providerCategory" ? selectedType.code : undefined,
      gameType: selectedType?.source === "gameType" ? selectedType.code : undefined,
      gameName: importCatalogKeyword.value.trim() || undefined,
    });
    importCatalogRows.value = result.list || [];
    importCatalogTotal.value = result.total || 0;
    selectedImportGames.value = [];
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.loadFailed"));
  } finally {
    importCatalogLoading.value = false;
  }
}
function changeImportSelection(selection: GameCenterRecord[]) {
  selectedImportGames.value = selection;
}
async function saveImportedGames() {
  if (!selectedImportGames.value.length) {
    ElMessage.warning(t("gameCenter.validation.selectGames"));
    return;
  }
  importGamesLoading.value = true;
  try {
    const result = await importGames({
      ids: selectedImportGames.value.map((game) => Number(game.id)),
      reason: importReason.value.trim() || t("gameCenter.defaults.importGamesReason"),
    });
    ElMessage.success(
      t("gameCenter.messages.importDraftSaved", {
        count: Number(result.importedCount || 0),
        skipped: Number(result.skippedCount || 0),
      }),
    );
    importGamesVisible.value = false;
    notifyBrandDraftChanged();
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("gameCenter.messages.saveFailed"));
  } finally {
    importGamesLoading.value = false;
  }
}
async function removeGameFromCategory(row: RecordValue) {
  if (!detailTargetID.value) return;
  try {
    await ElMessageBox.confirm(
      t("gameCenter.dialogs.removeCategoryGameMessage", {
        game: String(row.name || row.gameCode || ""),
      }),
      t("gameCenter.dialogs.removeCategoryGameTitle"),
      { type: "warning" },
    );
    await removeCategoryGame(Number(detailTargetID.value), Number(row.id));
    const state = detailPageState("games");
    if (
      Array.isArray(detail.value.games) &&
      detail.value.games.length === 1 &&
      state.page > 1
    )
      state.page -= 1;
    await loadDetailTab("games");
    await load();
    ElMessage.success(t(isPlatform.value ? "gameCenter.messages.savedImmediate" : "gameCenter.messages.categoryDraftSaved"));
    notifyBrandDraftChanged();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("gameCenter.messages.actionFailed"),
      );
  }
}

async function loadRelationOptions(kind: RelationKind = "categories") {
  if (kind === "tags") {
    const result = await fetchTags({ page: 1, pageSize: 100 });
    tags.value = result.list || [];
    return;
  }
  if (kind === "collections") {
    const result = await fetchCollections({ page: 1, pageSize: 100 });
    collections.value = result.list || [];
    return;
  }
  const c = await fetchCategories({ page: 1, pageSize: 100 });
  categories.value = c.list || [];
}
async function openRelations(row: GameCenterRecord, kind: RelationKind) {
  relationGame.value = row;
  relationKind.value = kind;
  relationIDs.value = [];
  relationVisible.value = true;
  relationLoading.value = true;
  try {
    await loadRelationOptions(kind);
    const game = await config.value.detailLoader(row.id);
    const data = game.categoriesAndTags as RecordValue | undefined;
    const values = data?.[kind];
    if (Array.isArray(values))
      relationIDs.value = values
        .filter(isObject)
        .map((item) => Number(item.id))
        .filter(Boolean);
  } finally {
    relationLoading.value = false;
  }
}
const relationOptions = computed(() =>
  relationKind.value === "tags"
    ? tags.value
    : relationKind.value === "collections"
      ? collections.value
      : assignableCategoryOptions.value,
);
const relationDialogTitle = computed(() => {
  const key = `gameCenter.dialogs.${relationKind.value}Title`;
  return te(key) ? t(key) : t("gameCenter.dialogs.relationTitle");
});
async function saveRelations() {
  if (!relationGame.value) return;
  relationLoading.value = true;
  try {
    if (relationKind.value === "tags")
      await setGameTags(relationGame.value.id, relationIDs.value);
    else if (relationKind.value === "collections")
      await setGameCollections(relationGame.value.id, relationIDs.value);
    else await setGameCategories(relationGame.value.id, relationIDs.value);
    relationVisible.value = false;
    ElMessage.success(
      t(
        relationKind.value === "categories"
          ? isPlatform.value
            ? "gameCenter.messages.savedImmediate"
            : "gameCenter.messages.draftSaved"
          : "gameCenter.messages.saved",
      ),
    );
    notifyBrandDraftChanged();
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("gameCenter.messages.saveFailed"),
    );
  } finally {
    relationLoading.value = false;
  }
}
function viewLogs(row: GameCenterRecord) {
  if (moduleKey.value === "game-sync") {
    void openDetail(row, "syncLogs");
    return;
  }
  void router.push({
    path:
      moduleKey.value === "provider-management"
        ? "/games/sync-logs"
        : "/games/launch-records",
    query:
      moduleKey.value === "provider-management"
        ? { providerCode: String(row.providerCode) }
        : { gameID: String(row.id) },
  });
}

watch(activeDetailTab, (tabKey) => {
  if (detailVisible.value && tabKey) void ensureDetailTab(tabKey);
});
watch(syncLogActiveTab, (tabKey) => {
  if (syncLogDetailVisible.value && tabKey)
    void ensureSyncLogDetailTab(tabKey);
});
watch(moduleKey, async () => {
  Object.keys(query).forEach((key) => delete query[key]);
  filterProviderTypeOptions.value = [];
  page.value = 1;
  detailVisible.value = false;
  await load();
  if (moduleKey.value === "game-list") await loadRelationOptions();
});
onMounted(async () => {
  await Promise.all([
    loadPermissions(),
    fetchAdminOptions()
      .then((options) => {
        adminOptions.value = options;
      })
      .catch(() => {
        adminOptions.value = {};
      }),
  ]);
  for (const [key, value] of Object.entries(route.query)) {
    if (typeof value === "string") query[key] = value;
  }
  if (moduleKey.value === "game-list" && query.providerCode) {
    const providerCode = String(query.providerCode);
    filterProviderTypeOptions.value = await fetchProviderOptionGameTypes(providerCode).catch(() => []);
  }
  await load();
  if (moduleKey.value === "game-list") await loadRelationOptions();
});
</script>

<template>
  <div v-if="config" class="game-center-page">
    <UserCenterCard :title="t(config.searchTitleKey)"
      ><el-form
        class="search-form"
        :model="query"
        label-position="left"
        label-width="132px"
        @keyup.enter="search"
        ><el-row :gutter="16"
          ><el-col
            v-for="field in visibleFilters"
            :key="field.key"
            :xs="24"
            :sm="12"
            :lg="8"
            :xl="6"
            ><el-form-item :label="t(field.labelKey)"
              ><el-date-picker
                v-if="field.type === 'dateRange'"
                v-model="query[field.key]"
                type="daterange"
                value-format="YYYY-MM-DD"
                :start-placeholder="t('gameCenter.placeholders.startDate')"
                :end-placeholder="
                  t('gameCenter.placeholders.endDate')
                " /><el-select
                v-else-if="field.type === 'select'"
                v-model="query[field.key]"
                clearable
                filterable
                :disabled="
                  isGameListPage &&
                  (field.key === 'providerCategory' || field.key === 'gameType') &&
                  !query.providerCode
                "
                :placeholder="t('gameCenter.placeholders.select')"
                @change="
                  field.key === 'providerCode' &&
                  changeFilterProvider(String($event || ''))
                "
                ><el-option
                  v-for="item in fieldOptions(field)"
                  :key="item.value"
                  :value="item.value"
                  :label="optionItemLabel(item)" /></el-select
              ><el-input
                v-else
                v-model="query[field.key]"
                clearable
                :placeholder="
                  t('gameCenter.placeholders.enter')
                " /></el-form-item></el-col
        ></el-row>
        <div class="search-actions">
          <el-button type="primary" @click="search">{{
            t("common.search")
          }}</el-button
          ><el-button @click="reset">{{ t("common.reset") }}</el-button>
        </div></el-form
      ></UserCenterCard
    >
    <UserCenterCard :title="t(config.tableTitleKey)"
      ><template #header
        ><div class="header-actions">
          <el-button
            v-if="
              createModule &&
              managedPermissions &&
              can(managedPermissions.create)
            "
            type="primary"
            @click="openEdit()"
            >{{ t("common.create") }}</el-button
          ><el-button
            v-if="isGameSyncPage && can('game:provider:sync')"
            type="primary"
            @click="runSync()"
            >{{ t("gameCenter.actions.syncAll") }}</el-button
          ><el-button
            v-if="isPlatform && isGameListPage && can('game:category:assign')"
            type="primary"
            @click="openCategoryMapping"
            >{{ t("gameCenter.actions.typeMapping") }}</el-button
          ><el-button
            v-if="!isPlatform && isGameListPage && can('game:update')"
            type="primary"
            @click="openImportGames"
            >{{ t("gameCenter.actions.importGames") }}</el-button
          >
          <el-button
            v-if="
              !isPlatform &&
              ['game-list', 'game-categories', 'lobby-sections'].includes(moduleKey) &&
              can('system:site:config:publish')
            "
            type="success"
            :icon="UploadFilled"
            :loading="publishingLobby"
            @click="publishLobbyConfiguration"
          >{{ t("brandPortal.actions.publish") }}</el-button>
          <template v-if="isGameListPage">
            <el-divider direction="vertical" />
            <span class="batch-count">{{
              t("gameCenter.messages.selectedCount", {
                count: selectedGameIDs.length,
              })
            }}</span>
            <el-dropdown
              trigger="click"
              :disabled="!hasSelectedGames"
            >
              <el-button :disabled="!hasSelectedGames">
                {{ t("gameCenter.actions.batch.title") }}
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item
                    :disabled="!can('game:status:update')"
                    @click="runBatchGameUpdate('platformEnabled', true)"
                    >{{ t("gameCenter.actions.batch.platformEnabled.on") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    :disabled="!can('game:status:update')"
                    @click="runBatchGameUpdate('platformEnabled', false)"
                    >{{ t("gameCenter.actions.batch.platformEnabled.off") }}</el-dropdown-item
                  >
                  <el-dropdown-item divided
                    :disabled="!can('game:update')"
                    @click="runBatchGameUpdate('isHot', true)"
                    >{{ t("gameCenter.actions.batch.isHot.on") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    :disabled="!can('game:update')"
                    @click="runBatchGameUpdate('isHot', false)"
                    >{{ t("gameCenter.actions.batch.isHot.off") }}</el-dropdown-item
                  >
                  <el-dropdown-item divided
                    :disabled="!can('game:update')"
                    @click="runBatchGameUpdate('isFeatured', true)"
                    >{{ t("gameCenter.actions.batch.isFeatured.on") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    :disabled="!can('game:update')"
                    @click="runBatchGameUpdate('isFeatured', false)"
                    >{{ t("gameCenter.actions.batch.isFeatured.off") }}</el-dropdown-item
                  >
                  <el-dropdown-item divided
                    :disabled="!can('game:update')"
                    @click="runBatchGameUpdate('isNew', true)"
                    >{{ t("gameCenter.actions.batch.isNew.on") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    :disabled="!can('game:update')"
                    @click="runBatchGameUpdate('isNew', false)"
                    >{{ t("gameCenter.actions.batch.isNew.off") }}</el-dropdown-item
                  >
                  <el-dropdown-item
                    divided
                    :disabled="!can('game:category:assign')"
                    @click="openBatchCategory"
                    >{{ t("gameCenter.actions.batch.addCategories") }}</el-dropdown-item
                  >
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </div></template
      >
      <el-alert
        v-if="usesDraftWorkflow"
        class="draft-page-alert"
        type="warning"
        :title="t('gameCenter.messages.draftModeHint')"
        :closable="false"
        show-icon
      />
      <el-alert
        v-else-if="isPlatform && (isGameListPage || isGameCategoriesPage)"
        class="draft-page-alert"
        type="info"
        :title="t('gameCenter.messages.platformImmediateHint')"
        :closable="false"
        show-icon
      />
      <el-table
        v-loading="loading"
        :data="tableRows"
        row-key="id"
        border
        :empty-text="t('gameCenter.messages.empty')"
        @selection-change="handleSelectionChange"
        @expand-change="handleCategoryExpand"
        ><el-table-column
          v-if="isGameListPage"
          type="selection"
          fixed="left"
          width="48"
        /><el-table-column
          v-if="isGameCategoriesPage"
          type="expand"
          width="48"
          fixed="left"
        >
          <template #default="{ row }">
            <div v-if="isPlatformDefaultCategory(row)" class="category-games-panel">
              <el-empty :description="t('gameCenter.messages.platformDefaultCategoryOperationDisabled')" />
            </div>
            <div v-else class="category-games-panel" v-loading="categoryGamesState(row).loading">
              <el-table
                :data="categoryGamesState(row).list"
                border
                size="small"
                :empty-text="t('gameCenter.messages.empty')"
              >
                <el-table-column :label="t('gameCenter.fields.thumbnail')" width="82">
                  <template #default="{ row: game }">
                    <el-image
                      v-if="categoryGameThumbnail(game)"
                      class="table-image"
                      :src="categoryGameThumbnail(game)"
                      fit="contain"
                      lazy
                    />
                    <span v-else>-</span>
                  </template>
                </el-table-column>
                <el-table-column :label="t('gameCenter.fields.gameName')" min-width="220">
                  <template #default="{ row: game }">
                    <el-input
                      v-model="game.name"
                      size="small"
                      :disabled="categoryGamesState(row).saving[game.id]"
                      @change="saveExpandedCategoryGame(row, game)"
                    />
                  </template>
                </el-table-column>
                <el-table-column :label="t('gameCenter.fields.provider')" width="130" show-overflow-tooltip>
                  <template #default="{ row: game }">
                    <el-tag size="small" type="info">{{ game.providerCode || "-" }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column :label="t('gameCenter.fields.providerCategory')" min-width="150" show-overflow-tooltip>
                  <template #default="{ row: game }">
                    <el-tag v-if="game.providerCategoryCode" size="small">{{ game.providerCategoryCode }}</el-tag>
                    <span v-else>-</span>
                  </template>
                </el-table-column>
                <el-table-column :label="t('gameCenter.fields.platformEnabled')" width="110" align="center">
                  <template #default="{ row: game }">
                    <el-switch
                      v-model="game.platformEnabled"
                      :disabled="categoryGamesState(row).saving[game.id]"
                      @change="saveExpandedCategoryGame(row, game)"
                    />
                  </template>
                </el-table-column>
                <el-table-column :label="t('gameCenter.fields.sortOrder')" width="120">
                  <template #default="{ row: game }">
                    <el-input-number
                      v-model="game.categorySortOrder"
                      size="small"
                      :min="0"
                      controls-position="right"
                      :disabled="categoryGamesState(row).saving[game.id]"
                      @change="saveExpandedCategoryGame(row, game)"
                    />
                  </template>
                </el-table-column>
                <el-table-column :label="t('common.action')" width="110" align="center">
                  <template #default="{ row: game }">
                    <el-button
                      link
                      type="danger"
                      :loading="categoryGamesState(row).saving[game.id]"
                      @click="removeExpandedCategoryGame(row, game)"
                    >{{ t("common.delete") }}</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <div class="category-games-pagination">
                <el-pagination
                  v-model:current-page="categoryGamesState(row).page"
                  v-model:page-size="categoryGamesState(row).pageSize"
                  background
                  small
                  layout="prev, pager, next, sizes, total"
                  :page-sizes="[10, 20, 50]"
                  :total="categoryGamesState(row).total"
                  @current-change="changeExpandedCategoryGamePage(row, $event)"
                  @size-change="changeExpandedCategoryGamePageSize(row, $event)"
                />
              </div>
            </div>
          </template>
        </el-table-column
        ><el-table-column
          type="index"
          :index="indexMethod"
          :label="t('common.index')"
          fixed="left"
          width="76"
        /><el-table-column
          v-for="column in visibleColumns"
          :key="column.key"
          :prop="column.key"
          :label="t(column.labelKey)"
          :width="column.width"
          show-overflow-tooltip
          ><template #default="{ row }"
            ><el-image
              v-if="column.image && imageSource(row, column.key)"
              class="table-image"
              :src="imageSource(row, column.key)"
              fit="contain"
              lazy
            /><span v-else-if="column.image">-</span
            ><el-tag
              v-else-if="isGameCategoriesPage && column.key === 'ownership'"
              :type="categoryOwnershipTagType(row)"
              size="small"
            >{{ categoryOwnershipLabel(row) }}</el-tag
            ><el-select
              v-else-if="
                isGameListPage &&
                column.key === 'platformCategories'
              "
              :model-value="categoryIDs(row)"
              multiple
              collapse-tags
              collapse-tags-tooltip
              :disabled="!can('game:category:assign')"
              :placeholder="t('gameCenter.fields.platformCategories')"
              @change="updateInlineCategories(row, $event)"
            >
              <el-option
                v-for="item in assignableCategoryOptions"
                :key="item.id"
                :label="categoryLabel(item)"
                :value="item.id"
              />
            </el-select
            ><el-switch
              v-else-if="
                column.key === 'status' &&
                isProviderModule
              "
              :model-value="Number(row.status) === 1"
              :disabled="!can('game:provider:update')"
              @change="changeBoolean(row, 'status')"
            /><div
              v-else-if="
                isGameListPage &&
                (column.key === 'publishedState' || column.key === 'draftState')
              "
              class="game-state-cell"
            >
              <div class="game-state-cell__row">
                <span>{{ t("gameCenter.fields.platformEnabled") }}</span>
                <el-tag
                  :type="
                    gameStateTagType(
                      column.key === 'publishedState'
                        ? gamePublishedState(row).enabled
                        : gameDraftState(row).enabled,
                    )
                  "
                  size="small"
                >
                  {{
                    gameEnabledLabel(
                      column.key === 'publishedState'
                        ? gamePublishedState(row).enabled
                        : gameDraftState(row).enabled,
                    )
                  }}
                </el-tag>
              </div>
              <div class="game-state-cell__row">
                <span>{{ t("gameCenter.fields.sortOrder") }}</span>
                <strong>{{
                  column.key === 'publishedState'
                    ? gamePublishedState(row).sortOrder
                    : gameDraftState(row).sortOrder
                }}</strong>
              </div>
            </div
            ><div
              v-else-if="isGameListPage && column.key === 'hasChanges'"
              class="game-change-cell"
            >
              <el-tag :type="gameConfigStatusType(row)" size="small">
                {{ gameConfigStatusLabel(row) }}
              </el-tag>
              <span v-if="gameDraftChangeCount(row)" class="game-change-cell__count">
                {{
                  t("gameCenter.messages.changeCount", {
                    count: gameDraftChangeCount(row),
                  })
                }}
              </span>
            </div
            ><UserStatusTag
              v-else-if="column.status || column.tag || isEnumKey(column.key)"
              :value="String(row[column.key] ?? '')"
              :label="optionLabel(String(row[column.key] ?? ''))"
            /><el-switch
              v-else-if="column.boolean"
              :model-value="gameBooleanCellValue(row, column.key)"
              :disabled="
                (isGameListPage &&
                  (column.key === 'providerEnabled' ||
                    (column.key === 'platformEnabled' &&
                      !can('game:status:update')) ||
                    (['isHot', 'isFeatured', 'isNew'].includes(column.key) &&
                      !can('game:update')))) ||
                (createModule &&
                  (!managedPermissions ||
                    !can(managedPermissions.update) ||
                    isPlatformDefaultCategory(row)))
              "
              @change="changeBoolean(row, column.key)"
            /><span
              v-else-if="isGameListPage && column.key === 'sortOrder'"
            >{{ gameDraftNumber(row, "sortOrder") }}</span
            ><span v-else-if="column.time">{{
              formatBackendTime(String(row[column.key] || ""))
            }}</span
            ><span v-else>{{ tableCellValue(row, column.key) }}</span></template
          ></el-table-column
        >
        <el-table-column
          :label="t('common.action')"
          fixed="right"
          :width="isGameSyncPage ? 430 : 350"
          ><template #default="{ row }"
            ><el-button
              link
              type="primary"
              :disabled="isPlatformDefaultCategory(row)"
              @click="openDetail(row)"
            >{{
              t("common.detail")
            }}</el-button
            ><template
              v-if="
                isProviderModule
              "
              ><el-button
                v-if="can('game:provider:test')"
                link
                @click="runTest(row)"
                >{{ t("gameCenter.actions.test") }}</el-button
              ><el-button
                v-if="can('game:provider:sync')"
                link
                type="primary"
                @click="runSync(row)"
                >{{ t("gameCenter.actions.sync") }}</el-button
              ><el-button
                v-if="can('game:provider:sync') && ['PENDING','RUNNING'].includes(String(row.latestSyncStatus || ''))"
                link
                type="warning"
                @click="runCancelSync(row)"
                >{{ t("gameCenter.actions.cancelSync") }}</el-button
              ><el-button
                v-if="can('game:sync-log:view')"
                link
                @click="viewLogs(row)"
                >{{ t("gameCenter.actions.viewLogs") }}</el-button
              ><el-button
                v-if="
                  isGameSyncPage &&
                  can('game:provider:update')
                "
                link
                type="danger"
                @click="runDeleteProvider(row)"
                >{{ t("gameCenter.actions.deleteProvider") }}</el-button
              ></template
            ><template v-else-if="isGameListPage"
              ><el-button
                v-if="can('game:update')"
                link
                @click="openEdit(row)"
                >{{ t("common.edit") }}</el-button
              ><el-dropdown
                ><el-button link type="primary">{{
                  t("common.more")
                }}</el-button
                ><template #dropdown
                  ><el-dropdown-menu
                    ><el-dropdown-item
                      v-if="can('game:category:assign')"
                      @click="openRelations(row, 'categories')"
                      >{{
                        t("gameCenter.actions.setCategories")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('game:tag:assign')"
                      @click="openRelations(row, 'tags')"
                      >{{
                        t("gameCenter.actions.setTags")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('game:collection:assign')"
                      @click="openRelations(row, 'collections')"
                      >{{
                        t("gameCenter.actions.setCollections")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="row.supportsDemo && can('game:demo:launch')"
                      divided
                      @click="runDemo(row)"
                      >{{
                        t("gameCenter.actions.launchDemo")
                      }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('game:launch-log:view')"
                      @click="viewLogs(row)"
                      >{{
                        t("gameCenter.actions.viewLaunchLogs")
                      }}</el-dropdown-item
                    ></el-dropdown-menu
                  ></template
                ></el-dropdown
              ></template
            ><template v-else-if="createModule && managedPermissions"
              ><el-button
                v-if="can(managedPermissions.update)"
                link
                :disabled="isPlatformDefaultCategory(row)"
                @click="openEdit(row)"
                >{{ t("common.edit") }}</el-button
              ><el-button
                v-if="can(managedPermissions.delete)"
                link
                type="danger"
                :disabled="isPlatformDefaultCategory(row)"
                @click="removeRow(row)"
                >{{ t("common.delete") }}</el-button
              ></template
            ></template
          ></el-table-column
        ></el-table
      >
      <div class="pagination-row">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          background
          layout="prev, pager, next, sizes, total"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </UserCenterCard>

    <el-dialog
      v-model="detailVisible"
      :title="t('gameCenter.dialogs.detailTitle')"
      width="min(1100px,94vw)"
      destroy-on-close
      ><div v-loading="detailLoading" class="detail-body">
        <el-tabs v-if="detailTabs.length" v-model="activeDetailTab" type="card"
          ><el-tab-pane
            v-for="tab in detailTabs"
            :key="tab.key"
            :name="tab.key"
            :label="t(tab.titleKey)"
            ><div
              v-loading="isLazyDetailTab(tab.key) ? detailPageState(tab.key).loading : false"
              class="detail-tab-panel"
            ><div
              v-if="
                isGameCategoriesPage &&
                tab.key === 'games' &&
                can('game:category:assign')
              "
              class="detail-tab-actions"
            >
              <el-button type="primary" @click="openAddCategoryGames">
                {{ t(isPlatform ? "gameCenter.actions.addGame" : "gameCenter.actions.addCategoryGamesDraft") }}
              </el-button>
            </div
            ><el-timeline
              v-if="isTimelineTab(tab.key)"
              class="session-timeline"
            >
              <el-timeline-item
                v-for="(item, index) in timelineItems(detail[tab.key])"
                :key="`${String(item.status || index)}-${String(item.time || index)}`"
                :timestamp="timelineTime(item)"
                :type="timelineItemType(item)"
                placement="top"
              >
                <div class="session-timeline__card">
                  <UserStatusTag
                    :value="String(item.status || '')"
                    :label="optionLabel(String(item.status || ''))"
                  />
                  <div
                    v-if="item.exitReason"
                    class="session-timeline__meta"
                  >
                    <span>{{ detailLabel("exitReason") }}</span>
                    <UserStatusTag
                      :value="String(item.exitReason || '')"
                      :label="optionLabel(String(item.exitReason || ''))"
                    />
                  </div>
                  <div
                    v-if="item.durationSeconds !== undefined && item.durationSeconds !== null && item.durationSeconds !== ''"
                    class="session-timeline__meta"
                  >
                    <span>{{ detailLabel("durationSeconds") }}</span>
                    <strong>{{ displayValue("durationSeconds", item.durationSeconds) }}</strong>
                  </div>
                </div>
              </el-timeline-item>
            </el-timeline
            ><el-card v-else-if="isTraceTab(tab.key)" shadow="never">
              <pre class="json-value trace-json">{{ traceDisplay(detail[tab.key]) }}</pre>
            </el-card
            ><div
              v-else-if="isGameListPage && tab.key === 'draftConfiguration'"
              class="draft-compare-panel"
            >
              <el-alert
                type="warning"
                :title="t('gameCenter.messages.editingDraftHint')"
                :closable="false"
                show-icon
              />
              <div class="draft-meta-row">
                <el-tag
                  :type="
                    String(gameDetailDraftMeta().status).toUpperCase() === 'DRAFT'
                      ? 'warning'
                      : 'success'
                  "
                  effect="plain"
                >
                  {{ optionLabel(gameDetailDraftMeta().status) }}
                </el-tag>
                <span>{{
                  t("gameCenter.messages.configVersion", {
                    version: gameDetailDraftMeta().version || "-",
                  })
                }}</span>
                <span>{{
                  t("gameCenter.messages.publishedAt", {
                    time: formatBackendTime(gameDetailDraftMeta().publishedAt),
                  })
                }}</span>
                <span>{{
                  t("gameCenter.messages.draftUpdatedAt", {
                    time: formatBackendTime(gameDetailDraftMeta().draftUpdatedAt),
                  })
                }}</span>
              </div>
              <el-table
                :data="gameDetailDraftDiffs()"
                border
                :empty-text="t('gameCenter.messages.empty')"
              >
                <el-table-column
                  prop="key"
                  :label="t('gameCenter.fields.configField')"
                  min-width="150"
                >
                  <template #default="{ row }">{{ detailLabel(row.key) }}</template>
                </el-table-column>
                <el-table-column
                  prop="published"
                  :label="t('gameCenter.fields.publishedValue')"
                  min-width="220"
                  show-overflow-tooltip
                >
                  <template #default="{ row }">
                    <pre
                      v-if="isObject(row.published) || Array.isArray(row.published)"
                      class="json-value"
                    >{{ displayValue(row.key, row.published) }}</pre>
                    <span v-else>{{ displayValue(row.key, row.published) }}</span>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="draft"
                  :label="t('gameCenter.fields.draftValue')"
                  min-width="220"
                  show-overflow-tooltip
                >
                  <template #default="{ row }">
                    <pre
                      v-if="isObject(row.draft) || Array.isArray(row.draft)"
                      class="json-value"
                    >{{ displayValue(row.key, row.draft) }}</pre>
                    <span v-else>{{ displayValue(row.key, row.draft) }}</span>
                  </template>
                </el-table-column>
                <el-table-column
                  prop="changed"
                  :label="t('gameCenter.fields.hasChanges')"
                  width="140"
                >
                  <template #default="{ row }">
                    <el-tag :type="row.changed ? 'warning' : 'success'" size="small">
                      {{
                        row.changed
                          ? t("gameCenter.values.unpublished_changes")
                          : t("gameCenter.values.no_changes")
                      }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </div
            ><template v-else-if="Array.isArray(detail[tab.key])">
              <el-table
                :data="detail[tab.key] as RecordValue[]"
                border
                max-height="480"
                :empty-text="t('gameCenter.messages.empty')"
                ><el-table-column
                  v-if="
                    isGameCategoriesPage &&
                    tab.key === 'games'
                  "
                  type="index"
                  :index="categoryGameIndex"
                  :label="t('common.index')"
                  width="72"
                  fixed="left"
                /><el-table-column
                  v-for="key in arrayColumns(detail[tab.key], tab.key)"
                  :key="key"
                  :label="detailLabel(key)"
                  min-width="145"
                  show-overflow-tooltip
                  ><template #default="{ row }"
                    ><UserStatusTag
                      v-if="isEnumKey(key)"
                      :value="String(detailCellValue(row, key) ?? '')"
                      :label="optionLabel(String(detailCellValue(row, key) ?? ''))"
                    /><el-image
                      v-else-if="isImageResourceKey(key) && detailCellValue(row, key)"
                      class="detail-table-image"
                      :src="String(detailCellValue(row, key))"
                      fit="cover"
                      :preview-src-list="[String(detailCellValue(row, key))]"
                      preview-teleported
                    /><span v-else>{{
                      displayValue(key, detailCellValue(row, key))
                    }}</span></template></el-table-column
                ><el-table-column
                  v-if="
                    isGameCategoriesPage &&
                    tab.key === 'games' &&
                    can('game:category:assign')
                  "
                  :label="t('common.action')"
                  fixed="right"
                  width="120"
                >
                  <template #default="{ row }">
                    <el-button link type="danger" @click="removeGameFromCategory(row)">
                      {{ t("gameCenter.actions.removeFromCategory") }}
                    </el-button>
                  </template>
                </el-table-column
                ><el-table-column
                  v-if="tab.key === 'syncLogs'"
                  :label="t('common.action')"
                  fixed="right"
                  width="100"
                >
                  <template #default="{ row }">
                    <el-button link type="primary" @click="openSyncLogDetail(row)">
                      {{ t("common.detail") }}
                    </el-button>
                  </template>
                </el-table-column
                ></el-table
              >
              <div v-if="isPagedDetailTab(tab.key)" class="detail-pagination-row">
                <el-pagination
                  v-model:current-page="detailPageState(tab.key).page"
                  v-model:page-size="detailPageState(tab.key).pageSize"
                  background
                  layout="prev, pager, next, sizes, total"
                  :page-sizes="[10, 20, 50, 100]"
                  :total="detailPageState(tab.key).total"
                  @current-change="changeDetailTabPage(tab.key, $event)"
                  @size-change="changeDetailTabPageSize(tab.key, $event)"
                />
              </div>
            </template
            ><el-card v-else-if="isObject(detail[tab.key])" shadow="never"
              ><el-descriptions :column="3" border
                ><el-descriptions-item
                  v-for="entry in detailEntries(detail[tab.key])"
                  :key="entry.key"
                  :label="detailLabel(entry.key)"
                >
                  <pre
                    v-if="isObject(entry.value) || Array.isArray(entry.value)"
                    class="json-value"
                    >{{ displayValue(entry.key, entry.value) }}</pre>
                  <el-image
                    v-else-if="isImageResourceKey(entry.key) && entry.value"
                    class="detail-resource-image"
                    :src="String(entry.value)"
                    fit="contain"
                    :preview-src-list="[String(entry.value)]"
                    preview-teleported
                  />
                  <el-tag v-else-if="isResourceKey(entry.key) && entry.value" type="success">
                    {{ t("gameCenter.messages.assetConfigured") }}
                  </el-tag>
                  <UserStatusTag
                    v-else-if="isEnumKey(entry.key)"
                    :value="String(entry.value ?? '')"
                    :label="optionLabel(String(entry.value ?? ''))"
                  /><span v-else>{{
                    displayValue(entry.key, entry.value)
                  }}</span></el-descriptions-item
                ></el-descriptions
              ></el-card
            ><el-card v-else-if="detail[tab.key]" shadow="never">
              <pre class="json-value trace-json">{{ traceDisplay(detail[tab.key]) }}</pre>
            </el-card
            ><el-empty
              v-else
              :description="t('gameCenter.messages.empty')" /></div></el-tab-pane
        ></el-tabs>
      </div>
      <template #footer
        ><el-button @click="detailVisible = false">{{
          t("common.close")
        }}</el-button></template
      ></el-dialog
    >

    <el-dialog
      v-model="syncLogDetailVisible"
      :title="t('gameCenter.dialogs.syncLogDetailTitle')"
      width="min(1100px,94vw)"
      append-to-body
      destroy-on-close
    >
      <div v-loading="syncLogDetailLoading" class="detail-body">
        <el-tabs v-model="syncLogActiveTab" type="card">
          <el-tab-pane
            v-for="tabKey in syncLogTabs"
            :key="tabKey"
            :name="tabKey"
            :label="t(`gameCenter.tabs.${tabKey}`)"
          >
            <div
              v-loading="
                isLazySyncLogDetailTab(tabKey)
                  ? syncLogPageState(tabKey).loading
                  : false
              "
              class="detail-tab-panel"
            >
              <el-card v-if="isTraceTab(tabKey)" shadow="never">
                <pre class="json-value trace-json">{{
                  traceDisplay(syncLogDetail[tabKey])
                }}</pre>
              </el-card>
              <template v-else-if="Array.isArray(syncLogDetail[tabKey])">
                <el-table
                  :data="syncLogDetail[tabKey] as RecordValue[]"
                  border
                  max-height="480"
                  :empty-text="t('gameCenter.messages.empty')"
                >
                  <el-table-column
                    v-for="key in arrayColumns(syncLogDetail[tabKey])"
                    :key="key"
                    :label="detailLabel(key)"
                    min-width="145"
                    show-overflow-tooltip
                  >
                    <template #default="{ row }">
                      <UserStatusTag
                        v-if="isEnumKey(key)"
                        :value="String(row[key] ?? '')"
                        :label="optionLabel(String(row[key] ?? ''))"
                      />
                      <span v-else>{{ displayValue(key, row[key]) }}</span>
                    </template>
                  </el-table-column>
                </el-table>
                <div
                  v-if="isPagedSyncLogDetailTab(tabKey)"
                  class="detail-pagination-row"
                >
                  <el-pagination
                    v-model:current-page="syncLogPageState(tabKey).page"
                    v-model:page-size="syncLogPageState(tabKey).pageSize"
                    background
                    layout="prev, pager, next, sizes, total"
                    :page-sizes="[10, 20, 50, 100]"
                    :total="syncLogPageState(tabKey).total"
                    @current-change="changeSyncLogTabPage(tabKey, $event)"
                    @size-change="changeSyncLogTabPageSize(tabKey, $event)"
                  />
                </div>
              </template>
              <el-card
                v-else-if="isObject(syncLogDetail[tabKey])"
                shadow="never"
              >
                <el-descriptions :column="3" border>
                  <el-descriptions-item
                    v-for="(value, key) in syncLogDetail[tabKey] as RecordValue"
                    :key="key"
                    :label="detailLabel(String(key))"
                  >
                    <UserStatusTag
                      v-if="isEnumKey(String(key))"
                      :value="String(value ?? '')"
                      :label="optionLabel(String(value ?? ''))"
                    />
                    <span v-else>{{ displayValue(String(key), value) }}</span>
                  </el-descriptions-item>
                </el-descriptions>
              </el-card>
              <el-empty
                v-else
                :description="t('gameCenter.messages.empty')"
              />
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="syncLogDetailVisible = false">
          {{ t("common.close") }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="editVisible"
      :title="
        t(
          editing
            ? 'gameCenter.dialogs.editTitle'
            : 'gameCenter.dialogs.createTitle',
        )
      "
      width="min(760px,94vw)"
      destroy-on-close
      ><el-alert
        v-if="usesDraftWorkflow"
        class="draft-dialog-alert"
        type="warning"
        :title="t('gameCenter.messages.draftModeHint')"
        :closable="false"
        show-icon
      />
      <el-form
        v-if="isGameListPage"
        v-loading="editLoading"
        label-position="left"
        label-width="132px"
        ><el-row :gutter="16"
          ><el-col :span="12"
            ><el-form-item :label="t('gameCenter.fields.gameName')"
              ><el-input v-model="gameForm.name" /></el-form-item></el-col
          ><el-col :span="24"
            ><el-form-item :label="t('gameCenter.fields.localThumbnailUrl')"
              ><div class="asset-field">
                <el-image
                  v-if="gameForm.localThumbnailUrl"
                  class="asset-field__preview"
                  :src="gameForm.localThumbnailUrl"
                  fit="contain"
                />
                <div v-else class="asset-field__empty">
                  {{ t("gameCenter.messages.noAsset") }}
                </div>
                <div class="asset-field__actions">
                  <el-upload
                    :show-file-list="false"
                    accept="image/*,.avif"
                    :http-request="uploadGameThumbnail"
                  >
                    <el-button
                      :loading="
                        uploadLoading && uploadTarget === 'gameThumbnail'
                      "
                      >{{
                        gameForm.localThumbnailUrl
                          ? t("gameCenter.actions.replaceAsset")
                          : t("gameCenter.actions.uploadAsset")
                      }}</el-button
                    >
                  </el-upload>
                  <el-button
                    v-if="gameForm.localThumbnailUrl"
                    type="danger"
                    link
                    @click="gameForm.localThumbnailUrl = ''"
                    >{{ t("gameCenter.actions.removeAsset") }}</el-button
                  >
                </div>
              </div></el-form-item></el-col
          ><el-col :span="8"
            ><el-form-item :label="t('gameCenter.fields.sortOrder')"
              ><el-input-number
                v-model="gameForm.sortOrder" /></el-form-item></el-col
          ><el-col :span="16"
            ><el-form-item :label="t('gameCenter.fields.flags')"
              ><el-checkbox v-model="gameForm.platformEnabled">{{
                t("gameCenter.fields.platformEnabled")
              }}</el-checkbox
              ><el-checkbox v-model="gameForm.isHot">{{
                t("gameCenter.fields.hot")
              }}</el-checkbox
              ><el-checkbox v-model="gameForm.isFeatured">{{
                t("gameCenter.fields.featured")
              }}</el-checkbox
              ><el-checkbox v-model="gameForm.isNew">{{
                t("gameCenter.fields.new")
              }}</el-checkbox></el-form-item
            ></el-col
          ><el-col :span="24"
            ><el-form-item :label="t('gameCenter.fields.supportedCountries')"
              ><el-input v-model="gameForm.supportedCountries" /></el-form-item
            ><el-form-item :label="t('gameCenter.fields.supportedCurrencies')"
              ><el-input v-model="gameForm.supportedCurrencies" /></el-form-item
            ><el-form-item :label="t('gameCenter.fields.supportedDevices')"
              ><el-input v-model="gameForm.supportedDevices" /></el-form-item
            ><el-form-item :label="t('gameCenter.fields.reason')" required
              ><el-input
                v-model="gameForm.reason"
                type="textarea" /></el-form-item></el-col></el-row></el-form
      ><el-form
        v-else-if="moduleKey === 'lobby-sections'"
        v-loading="editLoading"
        label-position="left"
        label-width="140px"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.sectionCode')" required>
              <el-input v-model="lobbyForm.sectionCode" :disabled="Boolean(editing)" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.sectionType')" required>
              <el-select v-model="lobbyForm.sectionType" filterable>
                <el-option
                  v-for="item in ['banner','game-list','category-list','provider-list','jackpot']"
                  :key="item"
                  :label="optionLabel(item)"
                  :value="item"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.titleZh')">
              <el-input v-model="lobbyForm.titleZh" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.titleEn')">
              <el-input v-model="lobbyForm.titleEn" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.displayStyle')">
              <el-select v-model="lobbyForm.displayStyle" filterable>
                <el-option
                  v-for="item in ['carousel','horizontal','grid','ticker']"
                  :key="item"
                  :label="optionLabel(item)"
                  :value="item"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.sourceType')">
              <el-select v-model="lobbyForm.sourceType" filterable>
                <el-option
                  v-for="item in ['banner','flag','category','provider','user','jackpot']"
                  :key="item"
                  :label="optionLabel(item)"
                  :value="item"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.sourceRef')">
              <el-input v-model="lobbyForm.sourceRef" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="t('gameCenter.fields.maxItems')">
              <el-input-number v-model="lobbyForm.maxItems" :min="1" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item :label="t('gameCenter.fields.sortOrder')">
              <el-input-number v-model="lobbyForm.sortOrder" :min="0" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('gameCenter.fields.enabled')">
              <el-switch v-model="lobbyForm.enabled" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item :label="t('gameCenter.fields.configJson')">
              <el-input v-model="lobbyConfigText" type="textarea" :rows="6" />
            </el-form-item>
            <el-form-item :label="t('gameCenter.fields.reason')">
              <el-input v-model="lobbyForm.reason" type="textarea" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form
      ><el-form v-else v-loading="editLoading" label-position="left" label-width="132px"
        ><el-row :gutter="16"
          ><el-col v-if="moduleKey !== 'game-categories'" :span="12"
            ><el-form-item :label="t('gameCenter.fields.code')" required
              ><el-input v-model="taxonomyForm.code" /></el-form-item></el-col
          ><el-col :span="12"
            ><el-form-item :label="t('gameCenter.fields.sortOrder')"
              ><el-input-number
                v-model="taxonomyForm.sortOrder" /></el-form-item></el-col
          ><el-col v-if="moduleKey !== 'game-categories'" :span="12"
            ><el-form-item :label="t('gameCenter.fields.nameZh')" required
              ><el-input v-model="taxonomyForm.nameZh" /></el-form-item></el-col
          ><el-col :span="12"
            ><el-form-item
              :label="t(isGameCategoriesPage ? 'gameCenter.fields.display' : 'gameCenter.fields.nameEn')"
              required
            >
              <el-input v-model="taxonomyForm.nameEn" /></el-form-item></el-col
          ><el-col :span="24"
            ><el-form-item :label="t('gameCenter.fields.icon')"
              ><div class="asset-field">
                <el-image
                  v-if="taxonomyForm.icon"
                  class="asset-field__preview"
                  :src="taxonomyForm.icon"
                  fit="contain"
                />
                <div v-else class="asset-field__empty">
                  {{ t("gameCenter.messages.noAsset") }}
                </div>
                <div class="asset-field__actions">
                <el-upload
                  :show-file-list="false"
                  accept="image/*,.avif"
                  :http-request="uploadTaxonomyIcon"
                >
                  <el-button
                    :loading="
                      uploadLoading && uploadTarget === 'taxonomyIcon'
                    "
                    >{{
                      taxonomyForm.icon
                        ? t("gameCenter.actions.replaceAsset")
                        : t("gameCenter.actions.uploadIcon")
                    }}</el-button
                  >
                </el-upload>
                  <el-button
                    v-if="taxonomyForm.icon"
                    type="danger"
                    link
                    @click="taxonomyForm.icon = ''"
                    >{{ t("gameCenter.actions.removeAsset") }}</el-button
                  >
                </div>
              </div></el-form-item
            ><el-form-item :label="t('gameCenter.fields.enabled')"
              ><el-switch v-model="taxonomyForm.enabled" /></el-form-item
            ></el-col></el-row></el-form
      ><template #footer
        ><el-button @click="editVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button type="primary" :loading="editLoading" @click="saveEdit">{{
          t(
            isGameListPage
              ? isPlatform
                ? 'common.save'
                : 'gameCenter.actions.saveDraft'
              : isGameCategoriesPage
                ? isPlatform
                  ? 'common.save'
                  : 'gameCenter.actions.saveCategoryDraft'
                : moduleKey === 'lobby-sections'
                  ? 'gameCenter.actions.saveLobbyDraft'
                : 'common.save',
          )
        }}</el-button></template
      ></el-dialog
    >

    <el-dialog v-model="reasonVisible" :title="t(reasonTitle)" width="480px"
      ><el-form label-position="left" label-width="112px"
        ><el-form-item :label="t('gameCenter.fields.reason')" required
          ><el-input
            v-model="reason"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit /></el-form-item></el-form
      ><template #footer
        ><el-button @click="reasonVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :disabled="!reason.trim()"
          :loading="reasonLoading"
          @click="confirmReason"
          >{{ t("common.confirm") }}</el-button
        ></template
      ></el-dialog
    >
    <el-dialog
      v-model="mappingVisible"
      :title="t('gameCenter.dialogs.categoryMappingTitle')"
      width="680px"
      destroy-on-close
      ><div v-loading="mappingLoading">
        <el-form label-position="left" label-width="140px">
          <el-form-item :label="t('gameCenter.fields.provider')" required>
            <el-select
              v-model="mappingForm.providerCode"
              filterable
              clearable
              :placeholder="t('gameCenter.placeholders.select')"
              class="relation-select"
              @change="loadMappingProviderTypes(String($event || ''))"
            >
              <el-option
                v-for="item in providerOptions"
                :key="String(item.providerCode)"
                :label="providerOptionLabel(item)"
                :value="String(item.providerCode)"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('gameCenter.fields.providerGameType')" required>
            <el-select
              v-model="mappingTypeKey"
              filterable
              clearable
              :loading="mappingTypeLoading"
              :disabled="!mappingForm.providerCode"
              :placeholder="t('gameCenter.placeholders.select')"
              class="relation-select"
            >
              <el-option
                v-for="item in providerTypeOptions"
                :key="providerTypeKey(item)"
                :label="providerTypeLabel(item)"
                :value="providerTypeKey(item)"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('gameCenter.fields.platformCategories')" required>
            <el-select
              v-model="mappingForm.categoryIDs"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              :placeholder="t('gameCenter.placeholders.select')"
              class="relation-select"
            >
              <el-option
                v-for="item in assignableCategoryOptions"
                :key="item.id"
                :label="categoryLabel(item)"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
      <template #footer
        ><el-button @click="mappingVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="mappingLoading"
          @click="saveCategoryMapping"
          >{{ t(isPlatform ? "common.save" : "gameCenter.actions.saveCategoryDraft") }}</el-button
        ></template
      ></el-dialog
    >
    <el-dialog
      v-model="relationVisible"
      :title="relationDialogTitle"
      width="620px"
      ><div v-loading="relationLoading">
        <el-select
          v-model="relationIDs"
          multiple
          filterable
          collapse-tags
          :placeholder="t('gameCenter.placeholders.select')"
          class="relation-select"
          ><el-option
            v-for="item in relationOptions"
            :key="item.id"
            :value="item.id"
            :label="`${item.code} · ${relationLabel(item)}`" /></el-select
        >
      </div>
      <template #footer
        ><el-button @click="relationVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="relationLoading"
          @click="saveRelations"
          >{{
            t(
              relationKind === 'categories'
                ? isPlatform
                  ? 'common.save'
                  : 'gameCenter.actions.saveCategoryDraft'
                : 'common.save',
            )
          }}</el-button
        ></template
      ></el-dialog
    >
    <el-dialog
      v-model="addCategoryGamesVisible"
      :title="t('gameCenter.dialogs.addCategoryGamesTitle')"
      width="min(960px,94vw)"
      append-to-body
      destroy-on-close
    >
      <el-form label-position="left" label-width="132px">
        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-form-item :label="t('gameCenter.fields.provider')" required>
              <el-select
                v-model="candidateForm.providerCode"
                filterable
                :loading="candidateProvidersLoading"
                :placeholder="t('gameCenter.placeholders.selectProvider')"
                @change="changeCandidateProvider"
              >
                <el-option
                  v-for="provider in candidateProviderOptions"
                  :key="String(provider.providerCode)"
                  :value="String(provider.providerCode)"
                  :label="String(provider.providerName || provider.providerCode)"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item :label="t('gameCenter.fields.providerCategory')" required>
              <el-select
                v-model="candidateForm.typeKey"
                filterable
                :loading="candidateTypesLoading"
                :disabled="!candidateForm.providerCode"
                :placeholder="t('gameCenter.placeholders.selectProviderCategory')"
                @change="changeCandidateType"
              >
                <el-option
                  v-for="option in candidateTypeOptions"
                  :key="candidateTypeKey(option)"
                  :value="candidateTypeKey(option)"
                  :label="`${option.code} (${option.gameCount})`"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item :label="t('gameCenter.fields.gameName')">
          <div class="candidate-search">
            <el-input
              v-model="candidateKeyword"
              clearable
              :disabled="!candidateForm.typeKey"
              :placeholder="t('gameCenter.placeholders.searchGames')"
              @keyup.enter="searchCandidateGames"
            />
            <el-button
              type="primary"
              :disabled="!candidateForm.typeKey"
              @click="searchCandidateGames"
            >
              {{ t("common.search") }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="candidateGamesLoading"
        :data="candidateGames"
        row-key="id"
        border
        max-height="420"
        :empty-text="t('gameCenter.messages.empty')"
        @selection-change="changeCandidateSelection"
      >
        <el-table-column type="selection" width="48" fixed="left" />
        <el-table-column
          :label="t('common.index')"
          type="index"
          :index="candidateGameIndex"
          width="72"
        />
        <el-table-column :label="t('gameCenter.fields.thumbnail')" width="84">
          <template #default="{ row }">
            <el-image
              v-if="row.localThumbnailUrl || row.thumbnailUrl"
              class="detail-table-image"
              :src="String(row.localThumbnailUrl || row.thumbnailUrl)"
              fit="cover"
            />
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="t('gameCenter.fields.gameName')" min-width="180" show-overflow-tooltip />
        <el-table-column prop="gameCode" :label="t('gameCenter.fields.gameCode')" min-width="150" show-overflow-tooltip />
        <el-table-column prop="providerGameId" :label="t('gameCenter.fields.providerGameId')" min-width="160" show-overflow-tooltip />
      </el-table>
      <div class="candidate-pagination">
        <span>{{ t("gameCenter.messages.selectedCount", { count: selectedCandidateGames.length }) }}</span>
        <el-pagination
          v-model:current-page="candidatePage"
          v-model:page-size="candidatePageSize"
          background
          layout="prev, pager, next, sizes, total"
          :page-sizes="[10, 20, 50]"
          :total="candidateTotal"
          @current-change="loadCandidateGames"
          @size-change="searchCandidateGames"
        />
      </div>
      <template #footer>
        <el-button @click="addCategoryGamesVisible = false">
          {{ t("common.cancel") }}
        </el-button>
        <el-button
          type="primary"
          :loading="addCategoryGamesLoading"
          :disabled="selectedCandidateGames.length === 0"
          @click="saveAddedCategoryGames"
        >
          {{ t(isPlatform ? "gameCenter.actions.addSelectedGames" : "gameCenter.actions.addSelectedGamesDraft") }}
        </el-button>
      </template>
    </el-dialog>
    <el-dialog
      v-model="importGamesVisible"
      :title="t('gameCenter.dialogs.importGamesTitle')"
      width="min(980px,94vw)"
      append-to-body
      destroy-on-close
    >
      <el-alert
        class="draft-dialog-alert"
        type="warning"
        :closable="false"
        show-icon
        :title="t('gameCenter.messages.importGamesHint')"
      />
      <el-form label-position="left" label-width="132px">
        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-form-item :label="t('gameCenter.fields.provider')">
              <el-select
                v-model="importCatalogForm.providerCode"
                clearable
                filterable
                :loading="importCatalogProviderLoading"
                :placeholder="t('gameCenter.placeholders.select')"
                @change="changeImportCatalogProvider"
              >
                <el-option
                  v-for="provider in importCatalogProviders"
                  :key="String(provider.providerCode)"
                  :value="String(provider.providerCode)"
                  :label="String(provider.providerName || provider.name || provider.providerCode)"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-form-item :label="t('gameCenter.fields.providerCategory')">
              <el-select
                v-model="importCatalogForm.typeKey"
                clearable
                filterable
                :loading="importCatalogTypeLoading"
                :disabled="!importCatalogForm.providerCode"
                :placeholder="t('gameCenter.placeholders.selectProviderCategory')"
                @change="changeImportCatalogType"
              >
                <el-option
                  v-for="option in importCatalogTypes"
                  :key="importCatalogTypeKey(option)"
                  :value="importCatalogTypeKey(option)"
                  :label="providerTypeLabel(option)"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item :label="t('gameCenter.fields.gameName')">
          <div class="candidate-search">
            <el-input
              v-model="importCatalogKeyword"
              clearable
              :placeholder="t('gameCenter.placeholders.searchGames')"
              @keyup.enter="searchImportCatalogGames"
            />
            <el-button type="primary" @click="searchImportCatalogGames">
              {{ t("common.search") }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <el-table
        v-loading="importCatalogLoading"
        :data="importCatalogRows"
        row-key="id"
        border
        max-height="420"
        :empty-text="t('gameCenter.messages.empty')"
        @selection-change="changeImportSelection"
      >
        <el-table-column type="selection" width="48" fixed="left" />
        <el-table-column
          :label="t('common.index')"
          type="index"
          :index="importCatalogIndex"
          width="72"
        />
        <el-table-column :label="t('gameCenter.fields.thumbnail')" width="84">
          <template #default="{ row }">
            <el-image
              v-if="row.localThumbnailUrl || row.thumbnailUrl"
              class="detail-table-image"
              :src="String(row.localThumbnailUrl || row.thumbnailUrl)"
              fit="cover"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" :label="t('gameCenter.fields.gameName')" min-width="190" show-overflow-tooltip />
        <el-table-column :label="t('gameCenter.fields.provider')" width="130" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.providerCode || "-" }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('gameCenter.fields.providerCategory')" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.providerCategoryCode" size="small">{{ row.providerCategoryCode }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="providerGameId" :label="t('gameCenter.fields.providerGameId')" min-width="160" show-overflow-tooltip />
      </el-table>
      <div class="candidate-pagination">
        <span>{{ t("gameCenter.messages.selectedCount", { count: selectedImportGames.length }) }}</span>
        <el-pagination
          v-model:current-page="importCatalogPage"
          v-model:page-size="importCatalogPageSize"
          background
          layout="prev, pager, next, sizes, total"
          :page-sizes="[10, 20, 50]"
          :total="importCatalogTotal"
          @current-change="loadImportCatalogGames"
          @size-change="searchImportCatalogGames"
        />
      </div>
      <el-form label-position="top" class="import-reason-form">
        <el-form-item :label="t('gameCenter.fields.reason')">
          <el-input
            v-model="importReason"
            type="textarea"
            :rows="2"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importGamesVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button
          type="primary"
          :loading="importGamesLoading"
          :disabled="selectedImportGames.length === 0"
          @click="saveImportedGames"
        >{{ t("gameCenter.actions.importSelectedGames") }}</el-button>
      </template>
    </el-dialog>
    <el-dialog
      v-model="batchCategoryVisible"
      :title="t('gameCenter.dialogs.batchCategoryTitle')"
      width="min(560px, 94vw)"
    >
      <el-form label-position="top">
        <el-form-item :label="t('gameCenter.fields.platformCategories')" required>
          <el-select
            v-model="batchCategoryForm.ids"
            multiple
            filterable
            clearable
            class="relation-select"
            :placeholder="t('gameCenter.placeholders.select')"
          >
            <el-option
              v-for="category in assignableCategoryOptions"
              :key="category.id"
              :value="category.id"
              :label="categoryLabel(category)"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('gameCenter.fields.reason')" required>
          <el-input
            v-model="batchCategoryForm.reason"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchCategoryVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button
          type="primary"
          :loading="batchCategoryLoading"
          :disabled="!batchCategoryForm.ids.length || !batchCategoryForm.reason.trim()"
          @click="saveBatchCategory"
        >{{ t(isPlatform ? "common.confirm" : "gameCenter.actions.saveCategoryDraft") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.game-center-page {
  display: grid;
  gap: 16px;
}
.search-form :deep(.el-select),
.search-form :deep(.el-date-editor) {
  width: 100%;
}
.search-actions,
.header-actions,
.pagination-row {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.header-actions {
  flex-wrap: wrap;
  align-items: center;
}
.batch-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.draft-page-alert,
.draft-dialog-alert {
  margin-bottom: 12px;
}
.game-state-cell,
.game-change-cell,
.draft-compare-panel {
  display: grid;
  gap: 8px;
}
.game-state-cell__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.game-state-cell__row span,
.game-change-cell__count {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.draft-meta-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 16px;
  color: var(--el-text-color-secondary);
}
.pagination-row {
  padding-top: 18px;
}
.table-image {
  width: 54px;
  height: 42px;
}
.detail-table-image {
  width: 52px;
  height: 52px;
  border: 1px solid var(--admin-border);
  border-radius: 4px;
  background: #f7f8fa;
}
.detail-body {
  min-height: 260px;
}
.detail-tab-panel {
  min-height: 180px;
}
.detail-pagination-row {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
}
.detail-tab-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}
.category-sort-input {
  width: 118px;
}
.category-games-panel {
  padding: 10px 12px;
  background: #f8fafc;
}
.category-games-panel :deep(.el-input-number) {
  width: 96px;
}
.category-games-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 10px;
}
.candidate-search {
  display: flex;
  width: 100%;
  gap: 10px;
}
.candidate-search :deep(.el-input) {
  flex: 1;
}
.candidate-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 14px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.import-reason-form {
  padding-top: 12px;
}
.json-value {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font: inherit;
}
.trace-json {
  max-height: 520px;
  overflow: auto;
  border-radius: 6px;
  background: #0f172a;
  color: #dbeafe;
  padding: 14px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono",
    "Courier New", monospace;
  font-size: 12px;
  line-height: 1.55;
}
.session-timeline {
  padding: 8px 12px 0;
}
.session-timeline__card {
  display: grid;
  gap: 8px;
  align-items: start;
}
.session-timeline__meta {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.detail-resource-image {
  width: 160px;
  height: 88px;
  border: 1px solid var(--admin-border);
  border-radius: 4px;
  background: #f7f8fa;
}
.relation-select {
  width: 100%;
}
.collection-games {
  display: grid;
  width: 100%;
  gap: 10px;
}
.collection-game-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 130px auto;
  align-items: center;
  gap: 10px;
}
.asset-field {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: center;
  gap: 12px;
  width: 100%;
}
.asset-field__preview,
.asset-field__empty {
  width: 72px;
  height: 72px;
  border: 1px solid var(--admin-border);
  border-radius: 4px;
  background: #f7f8fa;
}
.asset-field__preview--banner,
.asset-field__empty--banner {
  width: 160px;
  height: 76px;
}
.asset-field__empty {
  display: grid;
  place-items: center;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.asset-field__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
:deep(.el-descriptions__label) {
  width: 160px;
}
:deep(.el-table .cell) {
  white-space: nowrap;
}
</style>
