<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import {
  ElMessage,
  ElMessageBox,
  type UploadRequestOptions,
} from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import H5NavigationIconPreview from "@/components/operations/H5NavigationIconPreview.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { businessColumns, businessEntries, isInternalIdentifierKey, sanitizeBusinessValue } from "@/utils/presentation";
import {
  adjustVIPByNumber,
  copyBanner,
  createOperation,
  deleteOperation,
  getRegisterRewardConfig,
  getMusicPlaybackConfig,
  getNotificationAudienceOptions,
  getOperation,
  listOperations,
  queueNotification,
  saveVoucherRewardTiers,
  setPromotionStatus,
  setMusicEnabled,
  setOperationDefaultVisibility,
  transitionBonus,
  updateOperation,
  updateRegisterRewardConfig,
  updateMusicPlaybackConfig,
  uploadOperationAsset,
  type OperationDetail,
  type OperationPayload,
  type OperationQuery,
  type OperationRecord,
  type NotificationAudienceOption,
} from "@/api/adminOperations";
import { operationConfigs, type OperationColumn, type OperationField } from "./config";
import { fetchGames } from "@/api/adminGames";
import {
  fetchAdminOptions,
  optionList,
  type AdminOptions,
} from "@/api/adminOptions";

type ObjectValue = Record<string, unknown>;
type ParticipationRule = {
  ruleType: string;
  comparisonOperator: string;
  numericValue: string;
  textValue: string;
  enabled: boolean;
  sortOrder: number;
};
type RewardRule = {
  rewardType: string;
  calculationType: string;
  rewardAmount: string;
  percentageRate: string;
  currency: string;
  wageringMultiplier: string;
  minimumAmount: string;
  maximumReward: string;
  expirationDays: number;
  enabled: boolean;
  sortOrder: number;
};
type BenefitRule = {
  benefitCode: string;
  nameZh: string;
  nameEn: string;
  benefitValue: string;
  valueUnit: string;
  enabled: boolean;
  sortOrder: number;
};
type VoucherRewardTier = {
  id: number;
  amountOptionID: number;
  amount: string;
  bonusReward: string;
  cashReward: string;
  rewardAmount: string;
  sort: number;
  enabled: boolean;
};

const route = useRoute();
const router = useRouter();
const { t, te } = useI18n();
const { can, loadPermissions } = useAdminPermissions();
const moduleKey = computed(() => String(route.meta.moduleKey || "banners"));
const config = computed(() => operationConfigs[moduleKey.value]);
const visibleFilters = computed(() => config.value?.filters.filter((item) => !isInternalIdentifierKey(item.key)) || []);
const visibleColumns = computed(() => config.value?.columns.filter((item) => !isInternalIdentifierKey(item.key)) || []);
const isRegisterRewardPage = computed(() => moduleKey.value === "register-reward");
const isMusicPage = computed(() => moduleKey.value === "music");
const isVouchersPage = computed(() => moduleKey.value === "vouchers");
const showRegisterRewardPanel = computed(
  () => isRegisterRewardPage.value || isVouchersPage.value,
);
const query = reactive<Record<string, string | string[]>>({});
const rows = ref<OperationRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<OperationDetail>({});
const activeTab = ref("");
const editVisible = ref(false);
const editLoading = ref(false);
const editingID = ref(0);
const form = reactive<ObjectValue>({});
const participationRules = ref<ParticipationRule[]>([]);
const rewardRules = ref<RewardRule[]>([]);
const eligibleGameSelection = ref<number[]>([]);
const eligibleGameOptions = ref<Array<{ value: number; label: string }>>([]);
const benefits = ref<BenefitRule[]>([]);
const voucherRewardTiers = ref<VoucherRewardTier[]>([]);
const reasonVisible = ref(false);
const reasonLoading = ref(false);
const reasonForm = reactive({
  action: "",
  id: 0,
  reason: "",
  status: "",
  playerNo: "",
  level: "",
  scheduledAt: "",
});
const previewVisible = ref(false);
const previewURL = ref("");
const uploadingField = ref("");
const switchingColumn = reactive({ id: 0, key: "" });
const musicPlaybackLoading = ref(false);
const musicPlaybackSaving = ref(false);
const musicPlaybackForm = reactive({
  mode: "RANDOM" as "RANDOM" | "SPECIFIED",
  specifiedMusicID: 0,
  reason: "",
});
const musicPlaybackOptions = ref<
  Array<{
    id: number;
    nameZh: string;
    nameEn: string;
    scene: string;
    enabled: boolean;
  }>
>([]);
const registerRewardLoading = ref(false);
const registerRewardSaving = ref(false);
const registerRewardForm = reactive({
  enabled: true,
  minAmount: 1,
  maxAmount: 20,
  auditMultiplier: 1,
  updatedAt: "",
  reason: "",
});
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providerGameOptions = ref<Array<{ label: string; value: number }>>([]);
const providerGamesLoading = ref(false);
const notificationAudienceSelection = ref<string[]>([]);
const notificationAudienceOptions = ref<NotificationAudienceOption[]>([]);
const notificationAudienceLoading = ref(false);
const adminOptions = ref<AdminOptions>({});
let providerGameSearchSequence = 0;
let notificationAudienceSearchSequence = 0;
const notificationOptionAudienceTypes = new Set([
  "SELECTED_USERS",
  "VIP_LEVEL",
  "USER_TAG",
]);
const notificationRemoteAudienceTypes = new Set(["SELECTED_USERS", "USER_TAG"]);

const visibleTabs = computed(() =>
  config.value.tabs.filter((tab) => tab.key in detail.value),
);
const canCreate = computed(() =>
  Boolean(config.value.createPermission && can(config.value.createPermission)),
);
const hasFallbackRows = computed(() => rows.value.some((row) => row.isFallback));

function isObject(value: unknown): value is ObjectValue {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function translatedValue(value: unknown) {
  const text = String(value ?? "");
  const key = `operations.values.${text.toLowerCase()}`;
  return te(key) ? t(key) : text || "-";
}
function splitOptionValues(value: unknown) {
  return String(value || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}
function fieldOptions(field: OperationField) {
  if (field.source === "providers") return providerOptions.value;
  if (field.source === "providerGames") return providerGameOptions.value;
  if (field.source) return optionList(adminOptions.value, field.source);
  return (field.options || []).map((value) => ({
    label: translatedValue(value),
    value,
  }));
}
const promotionCurrencyOptions = computed(() => {
  const options = optionList(adminOptions.value, "currencies");
  return options.length
    ? options
    : ["PKR", "USD", "MXN", "JPY", "THB", "IDR"].map((value) => ({ label: value, value }));
});
function isFieldVisible(field: OperationField) {
  if (moduleKey.value === "push-messages" && field.key === "audienceValue")
    return form.audienceType !== "ALL_USERS";
  if (moduleKey.value !== "h5-navigation") return true;
  if (field.key === "pageRoute") return form.actionType === "PAGE";
  if (field.key === "providerCode" || field.key === "gameId")
    return form.actionType === "GAME";
  return true;
}
function usesNotificationAudienceOptions(
  audienceType = String(form.audienceType || ""),
) {
  return notificationOptionAudienceTypes.has(audienceType);
}
function usesRemoteNotificationAudience(
  audienceType = String(form.audienceType || ""),
) {
  return notificationRemoteAudienceTypes.has(audienceType);
}
function notificationAudiencePlaceholder() {
  if (form.audienceType === "SELECTED_USERS")
    return t("operations.placeholders.searchUsers");
  if (form.audienceType === "USER_TAG")
    return t("operations.placeholders.searchUserTags");
  return t("operations.placeholders.selectVIPLevels");
}
async function loadNavigationProviders() {
  if (moduleKey.value !== "h5-navigation") return;
  providerOptions.value = optionList(adminOptions.value, "providers").map((item) => ({
    label: item.label,
    value: item.value,
  })).filter((item) => item.value);
}
async function loadNavigationGames(providerCode: string, gameName = "") {
  const sequence = ++providerGameSearchSequence;
  providerGameOptions.value = [];
  if (!providerCode) return;
  providerGamesLoading.value = true;
  try {
    const result = await fetchGames({
      page: 1,
      pageSize: 100,
      providerCode,
      gameName: gameName.trim() || undefined,
    });
    if (sequence !== providerGameSearchSequence) return;
    providerGameOptions.value = result.list.map((item) => ({
      label: `${String(item.name || item.gameName || item.gameCode || item.providerGameId || "-")} · ${String(item.providerCode || "")}`,
      value: Number(item.id),
    })).filter((item) => Number.isInteger(item.value) && item.value > 0);
  } finally {
    if (sequence === providerGameSearchSequence) {
      providerGamesLoading.value = false;
    }
  }
}
async function loadEligibleGameOptions() {
  if (eligibleGameOptions.value.length) return;
  const result = await fetchGames({ page: 1, pageSize: 500 });
  eligibleGameOptions.value = result.list.map(item => ({
    value: Number(item.id),
    label: `${String(item.name || item.gameName || item.gameCode || "-")} · ${String(item.gameCode || item.providerGameId || "-")} · ${String(item.providerCode || "")}`,
  })).filter(item => Number.isInteger(item.value) && item.value > 0);
}
function searchNavigationGames(gameName: string) {
  if (moduleKey.value !== "h5-navigation") return;
  void loadNavigationGames(String(form.providerCode || ""), gameName);
}
function handleFieldChange(field: OperationField) {
  if (moduleKey.value === "push-messages" && field.key === "audienceType") {
    form.audienceValue = "";
    notificationAudienceSelection.value = [];
    notificationAudienceOptions.value = [];
    if (usesNotificationAudienceOptions())
      void loadNotificationAudienceOptions();
    return;
  }
  if (moduleKey.value === "activities") {
    if (field.key === "type") {
      const depositPromotion = ["FIRST_DEPOSIT_BONUS", "DEPOSIT_BONUS"].includes(String(form.type));
      form.triggerType = depositPromotion
        ? "DEPOSIT_SUCCEEDED"
        : form.type === "REGISTRATION_BONUS"
          ? "REGISTRATION_COMPLETED"
          : form.type === "CASHBACK"
            ? "GAME_BET_SETTLED"
            : "MANUAL_CLAIM";
      form.claimMode = depositPromotion ? "AUTO" : "MANUAL";
    }
    if (field.key === "triggerType")
      form.claimMode = form.triggerType === "DEPOSIT_SUCCEEDED" ? "AUTO" : "MANUAL";
    return;
  }
  if (moduleKey.value !== "h5-navigation") return;
  if (field.key === "actionType") {
    if (form.actionType === "PAGE") {
      form.providerCode = "";
      form.gameId = 0;
      providerGameOptions.value = [];
    } else {
      form.pageRoute = "";
    }
  }
  if (field.key === "providerCode") {
    form.gameId = 0;
    void loadNavigationGames(String(form.providerCode || ""));
  }
}

async function loadNotificationAudienceOptions(keyword = "", selected = "") {
  if (
    moduleKey.value !== "push-messages" ||
    !usesNotificationAudienceOptions()
  ) return;
  const sequence = ++notificationAudienceSearchSequence;
  notificationAudienceLoading.value = true;
  try {
    const options = await getNotificationAudienceOptions({
      audienceType: String(form.audienceType),
      keyword: keyword.trim() || undefined,
      selected: selected || undefined,
    });
    if (sequence !== notificationAudienceSearchSequence) return;
    notificationAudienceOptions.value = options;
  } finally {
    if (sequence === notificationAudienceSearchSequence)
      notificationAudienceLoading.value = false;
  }
}

function searchNotificationUsers(keyword: string) {
  void loadNotificationAudienceOptions(keyword);
}
function detailLabel(key: string) {
  const normalized = key.replace(/_([a-z])/g, (_, letter: string) =>
    letter.toUpperCase(),
  );
  const translationKey = `operations.fields.${normalized}`;
  return te(translationKey) ? t(translationKey) : normalized;
}
function detailMenuCode(tab: string) {
  const value = detail.value[tab];
  return isObject(value) ? String(value.menuCode || "") : "";
}
function detailEntries(tab: string) {
  const value = detail.value[tab];
  if (!isObject(value)) return [];
  return businessEntries(value).filter(
    ([key]) => !(moduleKey.value === "h5-navigation" && key === "menuCode"),
  );
}
function isResourceKey(key: string) {
  return /(image|icon|logo|banner|asset|thumbnail|background|animation|audio).*url$/i.test(key);
}
function isImageResourceKey(key: string) {
  return /(image|icon|logo|banner|thumbnail|background).*url$/i.test(key);
}
function isAudioResourceKey(key: string) {
  return /audio.*url$/i.test(key);
}
function displayValue(key: string, value: unknown) {
  if (value === null || value === undefined || value === "") return "-";
  if (isResourceKey(key)) return value ? t("operations.messages.assetConfigured") : "-";
  if (typeof value === "boolean")
    return value ? t("common.yes") : t("common.no");
  if (/(At|Time|_at)$/.test(key) && typeof value === "string")
    return formatBackendTime(value);
  if (typeof value === "string") return translatedValue(value);
  if (isObject(value) || Array.isArray(value))
    return JSON.stringify(sanitizeBusinessValue(value), null, 2);
  return String(value);
}
function arrayColumns(value: unknown) {
  if (!Array.isArray(value)) return [];
  const first = value.find(isObject);
  return first ? businessColumns([first], 10) : [];
}
function rowValue(row: OperationRecord, key: string) {
  return row[key];
}
function columnText(
  row: OperationRecord,
  field: { key: string; time?: boolean; status?: boolean },
) {
  const value = rowValue(row, field.key);
  if (field.time && typeof value === "string") return formatBackendTime(value);
  if (field.status) return translatedValue(value);
  return String(value ?? "-");
}

function buildQuery(): OperationQuery {
  const params: OperationQuery = { page: page.value, pageSize: pageSize.value };
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
async function loadRows() {
  if (isRegisterRewardPage.value) return;
  if (!config.value || !can(config.value.listPermission)) return;
  loading.value = true;
  try {
    const result = await listOperations(config.value.resource, buildQuery());
    rows.value = result.list;
    total.value = result.total;
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}

async function loadMusicPlaybackSettings() {
  if (!isMusicPage.value || !can("operation:music:list")) return;
  musicPlaybackLoading.value = true;
  try {
    const data = await getMusicPlaybackConfig();
    musicPlaybackForm.mode = data.mode;
    musicPlaybackForm.specifiedMusicID = data.specifiedMusicID || 0;
    musicPlaybackForm.reason = "";
    musicPlaybackOptions.value = data.musicOptions || [];
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.loadFailed"),
    );
  } finally {
    musicPlaybackLoading.value = false;
  }
}

async function saveMusicPlaybackSettings() {
  if (
    musicPlaybackForm.mode === "SPECIFIED" &&
    !musicPlaybackForm.specifiedMusicID
  ) {
    ElMessage.warning(t("operations.validation.specifiedMusicRequired"));
    return;
  }
  musicPlaybackSaving.value = true;
  try {
    await updateMusicPlaybackConfig({
      mode: musicPlaybackForm.mode,
      specifiedMusicID:
        musicPlaybackForm.mode === "SPECIFIED"
          ? musicPlaybackForm.specifiedMusicID
          : 0,
      reason: musicPlaybackForm.reason.trim(),
    });
    ElMessage.success(t("operations.messages.saved"));
    await loadMusicPlaybackSettings();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.saveFailed"),
    );
  } finally {
    musicPlaybackSaving.value = false;
  }
}

async function changeMusicEnabled(row: OperationRecord, value: unknown) {
  const enabled = Boolean(value);
  switchingColumn.id = row.id;
  switchingColumn.key = "enabled";
  try {
    await setMusicEnabled(row.id, enabled);
    row.enabled = enabled;
    ElMessage.success(t("operations.messages.operationSucceeded"));
    await loadMusicPlaybackSettings();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.operationFailed"),
    );
  } finally {
    switchingColumn.id = 0;
    switchingColumn.key = "";
  }
}

async function loadRegisterRewardConfig() {
  if (!can("operation:register-reward:view")) return;
  registerRewardLoading.value = true;
  try {
    const data = await getRegisterRewardConfig();
    registerRewardForm.enabled = data.enabled;
    registerRewardForm.minAmount = Number(data.minAmount || 0);
    registerRewardForm.maxAmount = Number(data.maxAmount || 0);
    registerRewardForm.auditMultiplier = Number(data.auditMultiplier || 0);
    registerRewardForm.updatedAt = data.updatedAt || "";
    registerRewardForm.reason = "";
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.loadFailed"),
    );
  } finally {
    registerRewardLoading.value = false;
  }
}
function validateRegisterRewardConfig() {
  const minAmount = registerRewardForm.minAmount;
  const maxAmount = registerRewardForm.maxAmount;
  const auditMultiplier = registerRewardForm.auditMultiplier;
  if (
    !Number.isFinite(minAmount) ||
    !Number.isFinite(maxAmount) ||
    !Number.isFinite(auditMultiplier) ||
    minAmount < 0 ||
    maxAmount < 0 ||
    auditMultiplier < 0
  ) {
    ElMessage.warning(t("operations.validation.nonNegativeNumber"));
    return false;
  }
  if (maxAmount < minAmount) {
    ElMessage.warning(t("operations.validation.amountRange"));
    return false;
  }
  return true;
}
async function submitRegisterRewardConfig() {
  if (!validateRegisterRewardConfig()) return;
  registerRewardSaving.value = true;
  try {
    await updateRegisterRewardConfig({
      enabled: registerRewardForm.enabled,
      minAmount: Number(registerRewardForm.minAmount).toFixed(2),
      maxAmount: Number(registerRewardForm.maxAmount).toFixed(2),
      auditMultiplier: Number(registerRewardForm.auditMultiplier).toFixed(1),
      reason: registerRewardForm.reason.trim(),
    });
    ElMessage.success(t("operations.messages.saved"));
    await loadRegisterRewardConfig();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.saveFailed"),
    );
  } finally {
    registerRewardSaving.value = false;
  }
}
function search() {
  page.value = 1;
  loadRows();
}
function reset() {
  Object.keys(query).forEach((key) => delete query[key]);
  page.value = 1;
  loadRows();
}

async function openDetail(row: OperationRecord, tab = "") {
	if (row.isFallback) return;
  detailVisible.value = true;
  detailLoading.value = true;
  detail.value = {};
  try {
    detail.value = await getOperation(config.value.resource, row.id);
    activeTab.value =
      (tab && visibleTabs.value.some((item) => item.key === tab) && tab) ||
      visibleTabs.value[0]?.key ||
      "";
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.loadFailed"),
    );
  } finally {
    detailLoading.value = false;
  }
}

function defaultValue(field: OperationField): unknown {
  if (field.multiple) return field.key === "channel" ? ["FCM"] : [];
  if (field.type === "switch")
    return field.key === "enabled" || field.key === "loop";
  if (field.key === "status" && moduleKey.value === "vouchers") return "1";
  if (field.key === "volumePercent") return 70;
  if (field.type === "number") return 0;
  const defaults: Record<string, string> = {
    channel: "FCM",
    audienceType: "ALL_USERS",
    platform: "ALL",
    platforms: "ALL",
    languages: "zh-CN,en-US",
    currencies: "PKR",
    currency: "PKR",
    targetType: "NONE",
    status: "DRAFT",
    voucherType: "DEPOSIT_BONUS",
    showFrequency: "EVERY_LOGIN",
    totalBudget: "0",
    requiredDeposit: "0",
    requiredValidBet: "0",
    withdrawalLimit: "0",
    cashbackRate: "0",
    birthdayReward: "0",
    monthlyReward: "0",
    minAmount: "0.00",
    maxAmount: "0.00",
    maxReward: "0.00",
  };
  return defaults[field.key] ?? "";
}
async function uploadAsset(
  field: OperationField,
  options: UploadRequestOptions,
) {
  uploadingField.value = field.key;
  try {
    const asset = await uploadOperationAsset(
      options.file,
      field.type === "audio" ? "audio" : "image",
    );
    form[field.key] = asset.url;
    options.onSuccess(asset);
    ElMessage.success(t("operations.messages.uploadSucceeded"));
  } catch (error) {
    const uploadError =
      error instanceof Error
        ? error
        : new Error(t("operations.messages.uploadFailed"));
    options.onError(
      uploadError as Parameters<UploadRequestOptions["onError"]>[0],
    );
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.uploadFailed"),
    );
  } finally {
    uploadingField.value = "";
  }
}
function assetUploadRequest(field: OperationField) {
  return (options: UploadRequestOptions) => uploadAsset(field, options);
}
function resetForm() {
  Object.keys(form).forEach((key) => delete form[key]);
  config.value.form.forEach((field) => {
    form[field.key] = defaultValue(field);
  });
  participationRules.value = [];
  rewardRules.value = [];
  benefits.value = [];
  voucherRewardTiers.value = [];
  eligibleGameSelection.value = [];
  notificationAudienceSelection.value = [];
  notificationAudienceOptions.value = [];
}
function detailSource(data: OperationDetail) {
  const source = data.overview || data.basicInformation;
  return isObject(source) ? source : {};
}
function assignForm(source: ObjectValue) {
  config.value.form.forEach((field) => {
    const aliases: Record<string, string> = {
      code: "promotionID",
      startTime: "startAt",
      endTime: "endAt",
    };
    const value = source[field.key] ?? source[aliases[field.key]];
    form[field.key] = field.multiple
      ? splitOptionValues(value)
      : value ?? defaultValue(field);
  });
  if (moduleKey.value === "push-messages") {
    notificationAudienceSelection.value = String(form.audienceValue || "")
      .split(",")
      .map((value) => value.trim())
      .filter(Boolean);
  }
}
async function openCreate() {
  editingID.value = 0;
  resetForm();
  editVisible.value = true;
  if (moduleKey.value === "push-messages") form.channel = ["FCM"];
  if (moduleKey.value === "activities") {
    await loadEligibleGameOptions();
    form.type = "FIRST_DEPOSIT_BONUS";
    form.triggerType = "DEPOSIT_SUCCEEDED";
    form.claimMode = "AUTO";
    form.currencies = ["PKR"];
    form.languages = ["zh-CN", "en-US"];
    participationRules.value = [normalizeParticipation({ ruleType: "FIRST_DEPOSIT" })];
    rewardRules.value = [normalizeReward({ rewardType: "BONUS", calculationType: "FIXED" })];
  }
  if (moduleKey.value === "h5-navigation") {
    form.actionType = "PAGE";
    await loadNavigationProviders();
  }
}
async function openEdit(row: OperationRecord) {
	if (row.isFallback) return;
  editingID.value = row.id;
  resetForm();
  editVisible.value = true;
  editLoading.value = true;
  try {
    const data = await getOperation(config.value.resource, row.id);
    const source = detailSource(data);
    assignForm(source);
    if (
      moduleKey.value === "push-messages" &&
      usesNotificationAudienceOptions()
    ) {
      await loadNotificationAudienceOptions(
        "",
        notificationAudienceSelection.value.join(","),
      );
    }
    if (moduleKey.value === "h5-navigation") {
      form.menuCode = String(source.menuCode || "");
      await loadNavigationProviders();
      await loadNavigationGames(String(form.providerCode || ""));
      const selectedGameID = Number(form.gameId || 0);
      const selectedGameName = String(detailSource(data).gameName || "");
      if (
        selectedGameID > 0 &&
        !providerGameOptions.value.some((item) => item.value === selectedGameID)
      ) {
        providerGameOptions.value.unshift({
          label: selectedGameName || String(selectedGameID),
          value: selectedGameID,
        });
      }
    }
    participationRules.value = Array.isArray(data.participationRules)
      ? data.participationRules.map((item) => normalizeParticipation(item))
      : [];
    rewardRules.value = Array.isArray(data.rewardRules)
      ? data.rewardRules.map((item) => normalizeReward(item))
      : [];
    benefits.value = Array.isArray(data.benefits)
      ? data.benefits.map((item) => normalizeBenefit(item))
      : [];
    voucherRewardTiers.value = Array.isArray(data.rewardTiers)
      ? data.rewardTiers.map((item) => normalizeVoucherTier(item))
      : [];
    await loadEligibleGameOptions();
    eligibleGameSelection.value = Array.isArray(data.eligibleGames)
      ? data.eligibleGames
          .map((item) => Number(isObject(item) ? item.id : 0))
          .filter((value) => Number.isInteger(value) && value > 0)
      : [];
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.loadFailed"),
    );
    editVisible.value = false;
  } finally {
    editLoading.value = false;
  }
}
function normalizeParticipation(value: unknown): ParticipationRule {
  const item = isObject(value) ? value : {};
  return {
    ruleType: String(item.ruleType || ""),
    comparisonOperator: String(item.comparisonOperator || "EQ"),
    numericValue: String(item.numericValue ?? ""),
    textValue: String(item.textValue || ""),
    enabled: item.enabled !== false,
    sortOrder: Number(item.sortOrder || 0),
  };
}
function normalizeReward(value: unknown): RewardRule {
  const item = isObject(value) ? value : {};
  return {
    rewardType: String(item.rewardType || "BONUS"),
    calculationType: String(item.calculationType || "FIXED"),
    rewardAmount: String(item.rewardAmount || "0"),
    percentageRate: String(item.percentageRate || "0"),
    currency: String(item.currency || "PKR"),
    wageringMultiplier: String(item.wageringMultiplier || "1"),
    minimumAmount: String(item.minimumAmount || "0"),
    maximumReward: String(item.maximumReward || "0"),
    expirationDays: Number(item.expirationDays || 0),
    enabled: item.enabled !== false,
    sortOrder: Number(item.sortOrder || 0),
  };
}
function normalizeBenefit(value: unknown): BenefitRule {
  const item = isObject(value) ? value : {};
  return {
    benefitCode: String(item.benefitCode || ""),
    nameZh: String(item.nameZh || ""),
    nameEn: String(item.nameEn || ""),
    benefitValue: String(item.benefitValue || "0"),
    valueUnit: String(item.valueUnit || "AMOUNT"),
    enabled: item.enabled !== false,
    sortOrder: Number(item.sortOrder || 0),
  };
}
function normalizeVoucherTier(value: unknown): VoucherRewardTier {
  const item = isObject(value) ? value : {};
  return {
    id: Number(item.id || 0),
    amountOptionID: Number(item.amountOptionID || 0),
    amount: String(item.amount || "0.00"),
    bonusReward: String(item.bonusReward || "0.00"),
    cashReward: String(item.cashReward || "0.00"),
    rewardAmount: String(item.rewardAmount || "0.00"),
    sort: Number(item.sort || 0),
    enabled: Number(item.status ?? 1) === 1,
  };
}
function addParticipationRule() {
  participationRules.value.push(normalizeParticipation({}));
}
function addRewardRule() {
  rewardRules.value.push(normalizeReward({}));
}
function addBenefit() {
  benefits.value.push(normalizeBenefit({}));
}
function addVoucherTier() {
  voucherRewardTiers.value.push(normalizeVoucherTier({}));
}
function removeAt<T>(items: T[], index: number) {
  items.splice(index, 1);
}
function formPayload(): OperationPayload {
  const payload: OperationPayload = {};
  config.value.form.forEach((field) => {
    const value = form[field.key];
    payload[field.key] = field.multiple
      ? Array.isArray(value)
        ? value.map((item) => String(item)).join(",")
        : ""
      : value ?? defaultValue(field);
  });
  if (payload.titleZh !== undefined && payload.titleEn === undefined)
    payload.titleEn = String(payload.titleZh || "").trim();
  if (payload.contentZh !== undefined && payload.contentEn === undefined)
    payload.contentEn = String(payload.contentZh || "").trim();
  if (payload.descriptionZh !== undefined && payload.descriptionEn === undefined)
    payload.descriptionEn = String(payload.descriptionZh || "").trim();
  if (moduleKey.value === "push-messages") {
    payload.titleEn = String(payload.titleEn || payload.titleZh || "").trim();
    payload.contentEn = String(payload.contentEn || payload.contentZh || "").trim();
    payload.audienceValue = form.audienceType === "ALL_USERS"
      ? ""
      : usesNotificationAudienceOptions(String(form.audienceType))
        ? notificationAudienceSelection.value.join(",")
        : String(form.audienceValue || "").trim();
  }
  if (moduleKey.value === "activities") {
    payload.participationRules = participationRules.value;
    payload.rewardRules = rewardRules.value;
    payload.eligibleGameIDs = eligibleGameSelection.value;
  }
  if (moduleKey.value === "vip") payload.benefits = benefits.value;
  if (moduleKey.value === "vouchers") {
    payload.titleEn = String(payload.titleEn || payload.titleZh || "").trim();
    payload.descriptionEn = String(payload.descriptionEn || payload.descriptionZh || "").trim();
    payload.status = Number(payload.status || 1);
  }
  return payload;
}
function validateForm() {
  for (const field of config.value.form)
    if (field.required && String(form[field.key] ?? "").trim() === "") {
      ElMessage.warning(
        t("operations.validation.required", { field: t(field.labelKey) }),
      );
      return false;
    }
  if (moduleKey.value === "h5-navigation") {
    const requiredKey = form.actionType === "GAME"
      ? (!form.providerCode ? "operations.fields.providerCode" : !form.gameId ? "operations.fields.gameName" : "")
      : (!form.pageRoute ? "operations.fields.pageRoute" : "");
    if (requiredKey) {
      ElMessage.warning(t("operations.validation.required", { field: t(requiredKey) }));
      return false;
    }
  }
  if (
    moduleKey.value === "push-messages" &&
    form.audienceType !== "ALL_USERS" &&
    (usesNotificationAudienceOptions(String(form.audienceType))
      ? notificationAudienceSelection.value.length === 0
      : !String(form.audienceValue || "").trim())
  ) {
    ElMessage.warning(t("operations.validation.audienceRequired"));
    return false;
  }
  if (moduleKey.value === "activities") {
    if (!rewardRules.value.some((rule) => rule.enabled)) {
      ElMessage.warning(t("operations.validation.rewardRequired"));
      return false;
    }
    if (
      form.status === "ACTIVE" &&
      (form.triggerType !== "DEPOSIT_SUCCEEDED" || form.claimMode !== "AUTO")
    ) {
      ElMessage.warning(t("operations.validation.triggerNotConnected"));
      return false;
    }
  }
  const start = form.startTime;
  const end = form.endTime || form.expiresAt;
  if (
    start &&
    end &&
    new Date(String(end)).getTime() < new Date(String(start)).getTime()
  ) {
    ElMessage.warning(t("operations.validation.timeRange"));
    return false;
  }
  return true;
}
async function submitForm() {
  if (!validateForm()) return;
  editLoading.value = true;
  try {
    let savedID = editingID.value;
    if (editingID.value) {
      const saved = await updateOperation(
        config.value.resource,
        editingID.value,
        formPayload(),
      );
      savedID = saved.id;
    } else {
      const saved = await createOperation(config.value.resource, formPayload());
      savedID = saved.id;
    }
    if (moduleKey.value === "vouchers" && can("operation:voucher:tier:update")) {
      await saveVoucherRewardTiers(
        savedID,
        voucherRewardTiers.value.map((tier) => ({
          id: tier.id,
          amountOptionID: tier.amountOptionID,
          amount: tier.amount,
          bonusReward: tier.bonusReward,
          cashReward: tier.cashReward,
          rewardAmount: tier.rewardAmount,
          sort: tier.sort,
          status: tier.enabled ? 1 : 2,
        })),
      );
    }
    ElMessage.success(t("operations.messages.saved"));
    editVisible.value = false;
    loadRows();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.saveFailed"),
    );
  } finally {
    editLoading.value = false;
  }
}

function ask(action: string, row: OperationRecord, status = "") {
  reasonForm.action = action;
  reasonForm.id = row.id;
  reasonForm.reason = "";
  reasonForm.status = status;
  reasonForm.playerNo = "";
  reasonForm.level = "";
  reasonForm.scheduledAt = "";
  reasonVisible.value = true;
}
async function submitReason() {
  if (!reasonForm.reason.trim()) {
    ElMessage.warning(t("operations.validation.reasonRequired"));
    return;
  }
  reasonLoading.value = true;
  try {
    if (reasonForm.action === "promotion-status")
      await setPromotionStatus(
        reasonForm.id,
        reasonForm.status,
        reasonForm.reason.trim(),
      );
    else if (["activate", "cancel", "forfeit"].includes(reasonForm.action))
      await transitionBonus(
        reasonForm.id,
        reasonForm.action as "activate" | "cancel" | "forfeit",
        reasonForm.reason.trim(),
      );
    else if (reasonForm.action === "vip-adjust")
      await adjustVIPByNumber(
        reasonForm.playerNo.trim(),
        Number(reasonForm.level),
        reasonForm.reason.trim(),
      );
    else if (
      ["schedule", "send", "notification-cancel"].includes(reasonForm.action)
    )
      await queueNotification(
        reasonForm.id,
        reasonForm.action === "notification-cancel"
          ? "cancel"
          : (reasonForm.action as "schedule" | "send"),
        {
          reason: reasonForm.reason.trim(),
          scheduledAt: reasonForm.scheduledAt || undefined,
        },
      );
    ElMessage.success(t("operations.messages.operationSucceeded"));
    reasonVisible.value = false;
    loadRows();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.operationFailed"),
    );
  } finally {
    reasonLoading.value = false;
  }
}
async function remove(row: OperationRecord) {
	if (row.isFallback) return;
  try {
    await ElMessageBox.confirm(
      t("operations.dialogs.deleteMessage"),
      t("operations.dialogs.deleteTitle"),
      {
        type: "warning",
        confirmButtonText: t("common.confirm"),
        cancelButtonText: t("common.cancel"),
      },
    );
    await deleteOperation(config.value.resource, row.id);
    ElMessage.success(t("operations.messages.deleted"));
    loadRows();
  } catch (error) {
    if (error === "cancel" || error === "close") return;
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.operationFailed"),
    );
  }
}
function switchValue(row: OperationRecord, column: OperationColumn) {
  if (moduleKey.value === "vouchers" && column.key === "status")
    return Number(row.status) === 1;
  return Boolean(rowValue(row, column.key));
}
function canSwitchColumn(column: OperationColumn, row?: OperationRecord) {
	if (!column.switch) return false;
	if (row?.isFallback)
		return Boolean(config.value.updatePermission && can(config.value.updatePermission));
  if (moduleKey.value === "vouchers" && column.key === "status")
    return can("operation:voucher:update");
  if (moduleKey.value === "music" && column.key === "enabled")
    return can("operation:music:update");
  return Boolean(
    column.key === "enabled" &&
      config.value.updatePermission &&
      can(config.value.updatePermission),
  );
}
function isSwitching(row: OperationRecord, column: OperationColumn) {
  return switchingColumn.id === row.id && switchingColumn.key === column.key;
}
async function changeSwitch(row: OperationRecord, column: OperationColumn, value: unknown) {
	if (!canSwitchColumn(column, row)) return;
	if (row.isFallback) {
		await changeFallbackVisibility(row, Boolean(value));
		return;
	}
  if (moduleKey.value === "music" && column.key === "enabled") {
    await changeMusicEnabled(row, value);
    return;
  }
  if (moduleKey.value === "vouchers" && column.key === "status") {
    await toggleVoucherStatus(row, Boolean(value));
    return;
  }
  if (column.key === "enabled") {
    await setOperationEnabled(row, Boolean(value));
  }
}
async function changeFallbackVisibility(row: OperationRecord, enabled: boolean) {
	switchingColumn.id = row.id;
	switchingColumn.key = "fallbackVisibility";
	try {
		await setOperationDefaultVisibility(config.value.resource, row.id, enabled);
		row.visible = enabled;
		if ("enabled" in row) row.enabled = enabled;
		if (moduleKey.value === "vouchers") row.status = enabled ? 1 : 2;
		ElMessage.success(t("operations.messages.defaultVisibilitySaved"));
	} catch (error) {
		ElMessage.error(error instanceof Error ? error.message : t("operations.messages.operationFailed"));
	} finally {
		switchingColumn.id = 0;
		switchingColumn.key = "";
	}
}
async function setOperationEnabled(row: OperationRecord, enabled: boolean) {
  switchingColumn.id = row.id;
  switchingColumn.key = "enabled";
  try {
    const data = await getOperation(config.value.resource, row.id);
    const source = detailSource(data);
    const payload: OperationPayload = {};
    const aliases: Record<string, string> = {
      code: "promotionID",
      startTime: "startAt",
      endTime: "endAt",
    };
    config.value.form.forEach((field) => {
      payload[field.key] =
        source[field.key] ?? source[aliases[field.key]] ?? defaultValue(field);
    });
    payload.enabled = enabled;
    if (moduleKey.value === "vip")
      payload.benefits = Array.isArray(data.benefits) ? data.benefits : [];
    await updateOperation(config.value.resource, row.id, payload);
    row.enabled = enabled;
    ElMessage.success(t("operations.messages.operationSucceeded"));
    loadRows();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.operationFailed"),
    );
  } finally {
    switchingColumn.id = 0;
    switchingColumn.key = "";
  }
}
async function toggleEnabled(row: OperationRecord) {
  await setOperationEnabled(row, !Boolean(row.enabled));
}
async function duplicate(row: OperationRecord) {
  try {
    if (moduleKey.value === "vouchers") {
      const data = await getOperation("vouchers", row.id);
      const source = detailSource(data);
      const payload: OperationPayload = {};
      config.value.form.forEach((field) => {
        payload[field.key] = source[field.key] ?? defaultValue(field);
      });
      const sourceCode = String(source.voucherCode || "VOUCHER").slice(0, 48);
      payload.voucherCode = `${sourceCode}-COPY-${Date.now().toString().slice(-6)}`;
      payload.status = 2;
      const created = await createOperation("vouchers", payload);
      if (
        can("operation:voucher:tier:update") &&
        Array.isArray(data.rewardTiers)
      ) {
        await saveVoucherRewardTiers(
          created.id,
          data.rewardTiers.map((value) => {
            const tier = normalizeVoucherTier(value);
            return {
              amountOptionID: tier.amountOptionID,
              amount: tier.amount,
              bonusReward: tier.bonusReward,
              cashReward: tier.cashReward,
              rewardAmount: tier.rewardAmount,
              sort: tier.sort,
              status: tier.enabled ? 1 : 2,
            };
          }),
        );
      }
    } else {
      await copyBanner(row.id);
    }
    ElMessage.success(t("operations.messages.copied"));
    loadRows();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.operationFailed"),
    );
  }
}
function preview(row: OperationRecord) {
  previewURL.value = String(
    row.iconURL || row.imageUrl || row.mobileImageUrl || "",
  );
  previewVisible.value = true;
}
async function toggleVoucherStatus(row: OperationRecord, enabled?: boolean) {
  switchingColumn.id = row.id;
  switchingColumn.key = "status";
  try {
    const data = await getOperation("vouchers", row.id);
    const source = detailSource(data);
    const payload: OperationPayload = {};
    config.value.form.forEach((field) => {
      payload[field.key] = source[field.key] ?? defaultValue(field);
    });
    const nextEnabled = enabled ?? Number(row.status) !== 1;
    payload.status = nextEnabled ? 1 : 2;
    await updateOperation("vouchers", row.id, payload);
    row.status = payload.status;
    ElMessage.success(t("operations.messages.operationSucceeded"));
    loadRows();
  } catch (error) {
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("operations.messages.operationFailed"),
    );
  } finally {
    switchingColumn.id = 0;
    switchingColumn.key = "";
  }
}
function viewUser(row: OperationRecord) {
  if (row.userNo) router.push(`/players/${encodeURIComponent(String(row.userNo))}`);
}
function exportPromo(row: OperationRecord) {
  const content = `${t("operations.fields.promoCode")},${t("operations.fields.usedCount")},${t("operations.fields.usageLimit")}\n${String(row.code || "")},${String(row.usedCount || 0)},${String(row.totalUsageLimit || 0)}\n`;
  const url = URL.createObjectURL(
    new Blob([content], { type: "text/csv;charset=utf-8" }),
  );
  const link = document.createElement("a");
  link.href = url;
  link.download = `promo-${String(row.code || row.id)}.csv`;
  link.click();
  URL.revokeObjectURL(url);
}

watch(moduleKey, async () => {
  Object.keys(query).forEach((key) => delete query[key]);
  page.value = 1;
  await loadPermissions();
  if (showRegisterRewardPanel.value) loadRegisterRewardConfig();
  if (!isRegisterRewardPage.value) {
    loadRows();
    if (isMusicPage.value) loadMusicPlaybackSettings();
  }
});
onMounted(async () => {
  const [, options] = await Promise.all([
    loadPermissions(),
    fetchAdminOptions().catch(() => ({})),
  ]);
  adminOptions.value = options;
  if (showRegisterRewardPanel.value) loadRegisterRewardConfig();
  if (!isRegisterRewardPage.value) {
    loadRows();
    if (isMusicPage.value) loadMusicPlaybackSettings();
  }
});
</script>

<template>
  <div v-if="showRegisterRewardPanel" class="operations-page">
    <UserCenterCard :title="t('operations.pages.registerReward.title')">
      <el-form
        v-loading="registerRewardLoading"
        class="register-reward-form"
        label-position="left"
        label-width="168px"
      >
        <el-form-item :label="t('operations.fields.enabled')">
          <el-switch v-model="registerRewardForm.enabled" />
        </el-form-item>
        <el-form-item :label="t('operations.fields.minAmount')" required>
          <el-input-number
            v-model="registerRewardForm.minAmount"
            :min="0"
            :precision="2"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item :label="t('operations.fields.maxAmount')" required>
          <el-input-number
            v-model="registerRewardForm.maxAmount"
            :min="0"
            :precision="2"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item
          :label="t('operations.fields.auditMultiplier')"
          required
        >
          <el-input-number
            v-model="registerRewardForm.auditMultiplier"
            :min="0"
            :precision="1"
            :step="0.1"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item :label="t('operations.fields.updatedTime')">
          <span class="register-reward-form__readonly">
            {{
              registerRewardForm.updatedAt
                ? formatBackendTime(registerRewardForm.updatedAt)
                : "-"
            }}
          </span>
        </el-form-item>
        <el-form-item :label="t('operations.fields.operatorReason')">
          <el-input
            v-model="registerRewardForm.reason"
            type="textarea"
            :rows="3"
            :placeholder="t('operations.placeholders.optionalReason')"
          />
        </el-form-item>
        <el-alert
          class="register-reward-form__hint"
          type="info"
          :closable="false"
          :title="t('operations.registerReward.auditHint')"
          show-icon
        />
        <div class="register-reward-form__actions">
          <el-button @click="loadRegisterRewardConfig">{{
            t("common.reset")
          }}</el-button>
          <el-button
            v-if="can('operation:register-reward:update')"
            type="primary"
            :loading="registerRewardSaving"
            @click="submitRegisterRewardConfig"
            >{{ t("common.save") }}</el-button
          >
        </div>
      </el-form>
    </UserCenterCard>
  </div>
  <div v-if="config && !isRegisterRewardPage" class="operations-page">
    <UserCenterCard
      v-if="isMusicPage"
      :title="t('operations.music.playbackSettings')"
    >
      <el-form
        v-loading="musicPlaybackLoading"
        class="music-playback-form"
        label-position="left"
        label-width="150px"
      >
        <el-form-item :label="t('operations.fields.playbackMode')" required>
          <el-radio-group v-model="musicPlaybackForm.mode">
            <el-radio-button value="RANDOM">{{
              t("operations.values.random")
            }}</el-radio-button>
            <el-radio-button value="SPECIFIED">{{
              t("operations.values.specified")
            }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          v-if="musicPlaybackForm.mode === 'SPECIFIED'"
          :label="t('operations.fields.specifiedMusic')"
          required
        >
          <el-select
            v-model="musicPlaybackForm.specifiedMusicID"
            filterable
            :placeholder="t('operations.placeholders.select')"
          >
            <el-option
              v-for="option in musicPlaybackOptions.filter((item) => item.enabled)"
              :key="option.id"
              :value="option.id"
              :label="`${option.nameZh} / ${option.nameEn}`"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('operations.fields.operatorReason')">
          <el-input
            v-model="musicPlaybackForm.reason"
            :placeholder="t('operations.placeholders.optionalReason')"
          />
        </el-form-item>
        <div class="music-playback-form__actions">
          <el-button @click="loadMusicPlaybackSettings">{{
            t("common.reset")
          }}</el-button>
          <el-button
            v-if="can('operation:music:update')"
            type="primary"
            :loading="musicPlaybackSaving"
            @click="saveMusicPlaybackSettings"
            >{{ t("common.save") }}</el-button
          >
        </div>
      </el-form>
    </UserCenterCard>
    <UserCenterCard :title="t(config.searchTitleKey)">
      <el-form
        class="operations-search"
        label-position="left"
        @submit.prevent="search"
      >
        <el-form-item
          v-for="field in visibleFilters"
          :key="field.key"
          :label="t(field.labelKey)"
        >
          <el-select
            v-if="field.type === 'select'"
            v-model="query[field.key]"
            clearable
            :placeholder="t('operations.placeholders.select')"
          >
            <el-option
              v-for="option in field.options"
              :key="option"
              :label="translatedValue(option)"
              :value="option"
            />
          </el-select>
          <el-date-picker
            v-else-if="field.type === 'dateRange'"
            v-model="query[field.key]"
            type="daterange"
            value-format="YYYY-MM-DD"
            :start-placeholder="t('operations.placeholders.startDate')"
            :end-placeholder="t('operations.placeholders.endDate')"
          />
          <el-input
            v-else
            v-model="query[field.key]"
            clearable
            :placeholder="t('operations.placeholders.enter')"
            @keyup.enter="search"
          />
        </el-form-item>
        <div class="operations-search__actions">
          <el-button type="primary" @click="search">{{
            t("common.search")
          }}</el-button
          ><el-button @click="reset">{{ t("common.reset") }}</el-button>
        </div>
      </el-form>
    </UserCenterCard>

    <UserCenterCard :title="t(config.tableTitleKey)">
      <template #header
        ><el-button v-if="canCreate" type="primary" @click="openCreate">{{
          t("operations.actions.create")
        }}</el-button></template
      >
      <el-table
        v-loading="loading"
        :data="rows"
        row-key="id"
        border
        empty-text=" "
      >
        <el-table-column
          fixed="left"
          type="index"
          :index="(index: number) => (page - 1) * pageSize + index + 1"
          :label="t('common.index')"
          width="72"
        />
        <el-table-column
          v-for="column in visibleColumns"
          :key="column.key"
          :prop="column.key"
          :label="t(column.labelKey)"
          :width="column.width"
          show-overflow-tooltip
        >
          <template #default="scope">
            <H5NavigationIconPreview
              v-if="moduleKey === 'h5-navigation' && column.image"
              :menu-code="String(scope.row.menuCode || '')"
              :icon-url="String(rowValue(scope.row, column.key) || '')"
            />
            <el-image
              v-else-if="column.image && rowValue(scope.row, column.key)"
              class="operations-thumbnail"
              :src="String(rowValue(scope.row, column.key))"
              fit="contain"
              :preview-src-list="[String(rowValue(scope.row, column.key))]"
              preview-teleported
            />
            <div
              v-else-if="
                column.status &&
                moduleKey === 'push-messages' &&
                column.key === 'channel'
              "
              class="notification-channel-tags"
            >
              <UserStatusTag
                v-for="channel in splitOptionValues(rowValue(scope.row, column.key))"
                :key="channel"
                :value="channel"
                :label="translatedValue(channel)"
              />
            </div>
            <div
              v-else-if="
                moduleKey === 'push-messages' &&
                column.key === 'audienceValue'
              "
              class="notification-channel-tags"
            >
              <template v-if="splitOptionValues(rowValue(scope.row, column.key)).length">
                <UserStatusTag
                  v-for="audience in splitOptionValues(rowValue(scope.row, column.key))"
                  :key="audience"
                  :value="audience"
                  :label="audience"
                />
              </template>
              <span v-else>-</span>
            </div>
            <UserStatusTag
              v-else-if="column.status"
              :value="String(rowValue(scope.row, column.key) || '')"
              :label="translatedValue(rowValue(scope.row, column.key))"
            />
            <el-switch
              v-else-if="column.switch && !scope.row.isFallback"
              :model-value="switchValue(scope.row, column)"
              :loading="isSwitching(scope.row, column)"
              :disabled="!canSwitchColumn(column, scope.row)"
              @change="
                (value: string | number | boolean) =>
                  changeSwitch(scope.row, column, value)
              "
            />
            <span v-else-if="column.switch">{{
              translatedValue(scope.row.visible ? "true" : "false")
            }}</span>
            <span v-else>{{ columnText(scope.row, column) }}</span>
          </template>
        </el-table-column>
		<el-table-column v-if="hasFallbackRows" :label="t('operations.fields.source')" width="180">
			<template #default="scope">
				<div v-if="scope.row.isFallback" class="fallback-source">
					<el-tag type="info" effect="plain">{{ t("operations.values.platformDefault") }}</el-tag>
					<el-switch
						:model-value="Boolean(scope.row.visible)"
						:loading="switchingColumn.id === scope.row.id && switchingColumn.key === 'fallbackVisibility'"
						:disabled="!config.updatePermission || !can(config.updatePermission)"
						@change="changeFallbackVisibility(scope.row, Boolean($event))"
					/>
				</div>
				<span v-else>{{ t("operations.values.siteCustom") }}</span>
			</template>
		</el-table-column>
        <el-table-column fixed="right" :label="t('common.action')" width="250">
          <template #default="scope">
            <div class="table-actions">
			  <el-button v-if="!scope.row.isFallback" link type="primary" @click="openDetail(scope.row)">{{
                t("common.detail")
              }}</el-button>
              <el-button
                v-if="
                  config.updatePermission &&
                  can(config.updatePermission) &&
				  config.form.length &&
				  !scope.row.isFallback
                "
                link
                type="primary"
                @click="openEdit(scope.row)"
                >{{ t("common.edit") }}</el-button
              >
			  <el-dropdown v-if="!scope.row.isFallback" trigger="click">
                <el-button link type="primary">{{
                  t("operations.actions.more")
                }}</el-button>
                <template #dropdown
                  ><el-dropdown-menu>
                    <el-dropdown-item
                      v-if="
                        moduleKey === 'banners' ||
                        moduleKey === 'popups' ||
                        moduleKey === 'vouchers'
                      "
                      @click="preview(scope.row)"
                      >{{ t("operations.actions.preview") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="
                        (moduleKey === 'banners' &&
                          can('operation:banner:create')) ||
                        (moduleKey === 'vouchers' &&
                          can('operation:voucher:create'))
                      "
                      @click="duplicate(scope.row)"
                      >{{ t("operations.actions.copy") }}</el-dropdown-item
                    >
                    <template v-if="moduleKey === 'vouchers'">
                      <el-dropdown-item
                        v-if="can('operation:voucher:update')"
                        @click="toggleVoucherStatus(scope.row)"
                        >{{
                          Number(scope.row.status) === 1
                            ? t("operations.actions.disable")
                            : t("operations.actions.enable")
                        }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('operation:voucher:tier:update')"
                        @click="openEdit(scope.row)"
                        >{{ t("operations.tabs.rewardTiers") }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('operation:voucher:usage:view')"
                        @click="openDetail(scope.row, 'usageRecords')"
                        >{{ t("operations.tabs.usageRecords") }}</el-dropdown-item
                      >
                    </template>
                    <el-dropdown-item
                      v-if="
                        config.updatePermission &&
                        can(config.updatePermission) &&
                        config.form.some((field) => field.key === 'enabled') &&
                        moduleKey !== 'music'
                      "
                      @click="toggleEnabled(scope.row)"
                      >{{
                        Boolean(scope.row.enabled)
                          ? t("operations.actions.disable")
                          : t("operations.actions.enable")
                      }}</el-dropdown-item
                    >
                    <template
                      v-if="
                        moduleKey === 'activities' &&
                        can('operation:promotion:status')
                      "
                      ><el-dropdown-item
                        v-for="status in [
                          'ACTIVE',
                          'PAUSED',
                          'ENDED',
                          'CANCELLED',
                        ]"
                        :key="status"
                        @click="ask('promotion-status', scope.row, status)"
                        >{{ translatedValue(status) }}</el-dropdown-item
                      ></template
                    >
                    <template v-if="moduleKey === 'bonus'"
                      ><el-dropdown-item
                        v-if="
                          scope.row.status === 'PENDING' &&
                          can('operation:bonus:activate')
                        "
                        @click="ask('activate', scope.row)"
                        >{{
                          t("operations.actions.activate")
                        }}</el-dropdown-item
                      ><el-dropdown-item
                        v-if="
                          ['PENDING', 'ACTIVE'].includes(
                            String(scope.row.status),
                          ) && can('operation:bonus:cancel')
                        "
                        @click="ask('cancel', scope.row)"
                        >{{ t("operations.actions.cancel") }}</el-dropdown-item
                      ><el-dropdown-item
                        v-if="
                          ['PENDING', 'ACTIVE'].includes(
                            String(scope.row.status),
                          ) && can('operation:bonus:forfeit')
                        "
                        @click="ask('forfeit', scope.row)"
                        >{{ t("operations.actions.forfeit") }}</el-dropdown-item
                      ><el-dropdown-item @click="viewUser(scope.row)">{{
                        t("operations.actions.viewUser")
                      }}</el-dropdown-item></template
                    >
                    <el-dropdown-item
                      v-if="moduleKey === 'vip' && can('operation:vip:adjust')"
                      @click="ask('vip-adjust', scope.row)"
                      >{{
                        t("operations.actions.adjustUserVIP")
                      }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="moduleKey === 'promo-codes'"
                      @click="exportPromo(scope.row)"
                      >{{ t("common.export") }}</el-dropdown-item
                    >
                    <template v-if="moduleKey === 'push-messages'"
                      ><el-dropdown-item
                        v-if="can('operation:notification:send')"
                        @click="ask('schedule', scope.row)"
                        >{{
                          t("operations.actions.schedule")
                        }}</el-dropdown-item
                      ><el-dropdown-item
                        v-if="can('operation:notification:send')"
                        @click="ask('send', scope.row)"
                        >{{ t("operations.actions.send") }}</el-dropdown-item
                      ><el-dropdown-item
                        v-if="can('operation:notification:cancel')"
                        @click="ask('notification-cancel', scope.row)"
                        >{{ t("operations.actions.cancel") }}</el-dropdown-item
                      ></template
                    >
                    <el-dropdown-item
                      v-if="
                        config.deletePermission && can(config.deletePermission)
                      "
                      divided
                      @click="remove(scope.row)"
                      >{{ t("operations.actions.delete") }}</el-dropdown-item
                    >
                  </el-dropdown-menu></template
                >
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
        <template #empty
          ><el-empty :description="t('operations.messages.empty')"
        /></template>
      </el-table>
      <div class="operations-pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100]"
          layout="prev, pager, next, sizes, total"
          :total="total"
          @current-change="loadRows"
          @size-change="search"
        />
      </div>
    </UserCenterCard>

    <el-dialog
      v-model="detailVisible"
      :title="t('operations.dialogs.detail')"
      width="min(960px, 92vw)"
      append-to-body
      destroy-on-close
    >
      <div v-loading="detailLoading">
        <el-tabs v-if="visibleTabs.length" v-model="activeTab" type="card"
          ><el-tab-pane
            v-for="tab in visibleTabs"
            :key="tab.key"
            :name="tab.key"
            :label="t(tab.labelKey)"
          >
            <template v-if="Array.isArray(detail[tab.key])"
              ><el-table :data="detail[tab.key]" border max-height="430"
                ><el-table-column
                  v-for="key in arrayColumns(detail[tab.key])"
                  :key="key"
                  :label="detailLabel(key)"
                  min-width="140"
                  show-overflow-tooltip
                  ><template #default="scope">{{
                    displayValue(key, scope.row[key])
                  }}</template></el-table-column
                ></el-table
              ></template
            >
            <el-descriptions
              v-else-if="isObject(detail[tab.key])"
              :column="2"
              border
              ><el-descriptions-item
                v-for="[key, value] in detailEntries(tab.key)"
                :key="key"
                :label="detailLabel(key)"
              >
                <pre
                  v-if="isObject(value) || Array.isArray(value)"
                  class="operations-json"
                  >{{ displayValue(String(key), value) }}</pre>
                <H5NavigationIconPreview
                  v-else-if="moduleKey === 'h5-navigation' && String(key) === 'iconUrl'"
                  :menu-code="detailMenuCode(tab.key)"
                  :icon-url="String(value || '')"
                  :size="58"
                />
                <el-image
                  v-else-if="isImageResourceKey(String(key)) && value"
                  class="detail-resource-image"
                  :src="String(value)"
                  fit="contain"
                  :preview-src-list="[String(value)]"
                  preview-teleported
                />
                <audio
                  v-else-if="isAudioResourceKey(String(key)) && value"
                  class="detail-resource-audio"
                  :src="String(value)"
                  controls
                  preload="metadata"
                />
                <el-tag v-else-if="isResourceKey(String(key)) && value" type="success">
                  {{ t("operations.messages.assetConfigured") }}
                </el-tag>
                <span v-else>{{
                  displayValue(String(key), value)
                }}</span></el-descriptions-item
              ></el-descriptions
            >
            <el-empty
              v-else
              :description="t('operations.messages.empty')"
            /> </el-tab-pane
        ></el-tabs>
      </div>
    </el-dialog>

    <el-dialog
      v-model="editVisible"
      :title="
        editingID
          ? t('operations.dialogs.edit')
          : t('operations.dialogs.create')
      "
      width="min(900px, 94vw)"
      append-to-body
      destroy-on-close
    >
      <div v-loading="editLoading">
        <el-form
          class="operations-form"
          label-position="left"
          label-width="132px"
        >
          <el-form-item
            v-for="field in config.form"
            v-show="isFieldVisible(field)"
            :key="field.key"
            :label="t(field.labelKey)"
            :required="field.required"
            :class="{ 'is-wide': field.span === 2 }"
          >
            <el-switch
              v-if="field.type === 'switch'"
              v-model="form[field.key]"
            />
            <el-input-number
              v-else-if="field.type === 'number'"
              v-model="form[field.key]"
              :min="0"
              controls-position="right"
            />
            <el-select
              v-else-if="
                moduleKey === 'push-messages' &&
                field.key === 'audienceValue' &&
                usesNotificationAudienceOptions()
              "
              v-model="notificationAudienceSelection"
              :loading="notificationAudienceLoading"
              :remote="usesRemoteNotificationAudience()"
              :remote-method="usesRemoteNotificationAudience() ? searchNotificationUsers : undefined"
              :placeholder="notificationAudiencePlaceholder()"
              filterable
              multiple
              collapse-tags
              collapse-tags-tooltip
            >
              <el-option
                v-for="option in notificationAudienceOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              >
                <div class="notification-audience-option">
                  <span>{{ option.label }}</span>
                  <small v-if="option.tagType">{{
                    translatedValue(`${String(option.tagType).toLowerCase()}_tag`)
                  }}</small>
                  <small v-else-if="option.userNo">{{ option.userNo }}</small>
                </div>
              </el-option>
            </el-select>
            <el-select
              v-else-if="field.type === 'select'"
              v-model="form[field.key]"
              :loading="field.source === 'providerGames' && providerGamesLoading"
              :remote="field.source === 'providerGames'"
              :remote-method="field.source === 'providerGames' ? searchNavigationGames : undefined"
              filterable
              :multiple="field.multiple"
              @change="handleFieldChange(field)"
              ><el-option
                v-for="option in fieldOptions(field)"
                :key="String(option.value)"
                :label="option.label"
                :value="option.value"
            /></el-select>
            <el-date-picker
              v-else-if="field.type === 'datetime'"
              v-model="form[field.key]"
              type="datetime"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
            />
            <el-input
              v-else-if="field.type === 'textarea'"
              v-model="form[field.key]"
              type="textarea"
              :rows="3"
            />
            <div
              v-else-if="field.type === 'image' || field.type === 'audio'"
              class="asset-field"
            >
              <el-upload
                :show-file-list="false"
                :accept="
                  field.type === 'audio'
                    ? 'audio/*,.mp4,.m4a,.aac'
                    : 'image/*,.avif,.svg,image/svg+xml'
                "
                :http-request="assetUploadRequest(field)"
              >
                <el-button :loading="uploadingField === field.key">{{
                  form[field.key]
                    ? t("operations.actions.replaceAsset")
                    : t(
                        field.type === "audio"
                          ? "operations.actions.uploadAudio"
                          : "operations.actions.uploadImage",
                      )
                }}</el-button>
              </el-upload>
              <el-button
                v-if="form[field.key]"
                type="danger"
                link
                @click="form[field.key] = ''"
                >{{ t("operations.actions.removeAsset") }}</el-button
              >
              <H5NavigationIconPreview
                v-if="moduleKey === 'h5-navigation' && field.type === 'image'"
                :menu-code="String(form.menuCode || '')"
                :icon-url="String(form[field.key] || '')"
                :size="64"
              />
              <el-image
                v-else-if="field.type === 'image' && form[field.key]"
                class="asset-field__image"
                :src="String(form[field.key])"
                fit="contain"
              />
              <audio
                v-else-if="field.type === 'audio' && form[field.key]"
                :src="String(form[field.key])"
                controls
                preload="metadata"
              />
            </div>
            <el-input v-else v-model="form[field.key]" />
          </el-form-item>
        </el-form>
        <section v-if="moduleKey === 'activities'" class="operations-rules">
          <div class="operations-rules__head">
            <strong>{{ t("operations.tabs.participationRules") }}</strong
            ><el-button @click="addParticipationRule">{{
              t("operations.actions.addRule")
            }}</el-button>
          </div>
          <div
            v-for="(rule, index) in participationRules"
            :key="index"
            class="operations-rule-row"
          >
            <el-select v-model="rule.ruleType" filterable
              ><el-option
                v-for="item in ['FIRST_DEPOSIT', 'DEPOSIT_SEQUENCE', 'DEPOSIT_AMOUNT', 'COUNTRY', 'CURRENCY', 'USER_VIP_LEVEL']"
                :key="item"
                :label="translatedValue(item)"
                :value="item" /></el-select
            ><el-select v-model="rule.comparisonOperator"
              ><el-option
                v-for="item in ['EQ', 'NE', 'GT', 'GTE', 'LT', 'LTE', 'IN']"
                :key="item"
                :label="item"
                :value="item" /></el-select
            ><el-input
              v-model="rule.numericValue"
              :placeholder="t('operations.fields.numericValue')"
            /><el-input
              v-model="rule.textValue"
              :placeholder="t('operations.fields.textValue')"
            /><el-tooltip :content="t('operations.fields.enabled')"
              ><el-switch
                v-model="rule.enabled"
                :aria-label="t('operations.fields.enabled')" /></el-tooltip
            ><el-button
              type="danger"
              link
              @click="removeAt(participationRules, index)"
              >{{ t("operations.actions.remove") }}</el-button
            >
          </div>
          <div class="operations-rules__head">
            <strong>{{ t("operations.tabs.rewardRules") }}</strong
            ><el-button @click="addRewardRule">{{
              t("operations.actions.addRule")
            }}</el-button>
          </div>
          <div
            v-for="(rule, index) in rewardRules"
            :key="index"
            class="operations-rule-row operations-rule-row--reward"
          >
            <el-select v-model="rule.rewardType"
              ><el-option
                v-for="item in ['CASH', 'BONUS']"
                :key="item"
                :label="translatedValue(item)"
                :value="item" /></el-select
            ><el-select v-model="rule.calculationType"
              ><el-option
                v-for="item in ['FIXED', 'PERCENTAGE']"
                :key="item"
                :label="translatedValue(item)"
                :value="item" /></el-select
            ><el-input
              v-if="rule.calculationType === 'FIXED'"
              v-model="rule.rewardAmount"
              :placeholder="t('operations.fields.rewardAmount')"
            /><el-input
              v-else
              v-model="rule.percentageRate"
              :placeholder="t('operations.fields.percentageRate')"
            /><el-select v-model="rule.currency" filterable
              ><el-option
                v-for="item in promotionCurrencyOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value" /></el-select
            ><el-input
              v-model="rule.wageringMultiplier"
              :placeholder="t('operations.fields.wageringMultiplier')"
            /><el-input
              v-model="rule.minimumAmount"
              :placeholder="t('operations.fields.minimumAmount')"
            /><el-input
              v-model="rule.maximumReward"
              :placeholder="t('operations.fields.maximumReward')"
            /><el-tooltip :content="t('operations.fields.expirationDays')"
              ><el-input-number
                v-model="rule.expirationDays"
                :aria-label="t('operations.fields.expirationDays')"
                :min="0" /></el-tooltip
            ><el-tooltip :content="t('operations.fields.enabled')"
              ><el-switch
                v-model="rule.enabled"
                :aria-label="t('operations.fields.enabled')" /></el-tooltip
            ><el-button
              type="danger"
              link
              @click="removeAt(rewardRules, index)"
              >{{ t("operations.actions.remove") }}</el-button
            >
          </div>
          <el-form-item :label="t('operations.fields.eligibleGames')"
            ><el-select v-model="eligibleGameSelection" multiple filterable collapse-tags><el-option v-for="item in eligibleGameOptions" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
        </section>
        <section v-if="moduleKey === 'vip'" class="operations-rules">
          <div class="operations-rules__head">
            <strong>{{ t("operations.tabs.benefits") }}</strong
            ><el-button @click="addBenefit">{{
              t("operations.actions.addBenefit")
            }}</el-button>
          </div>
          <div
            v-for="(benefit, index) in benefits"
            :key="index"
            class="operations-rule-row"
          >
            <el-input
              v-model="benefit.benefitCode"
              :placeholder="t('operations.fields.benefitCode')"
            /><el-input
              v-model="benefit.nameZh"
              :placeholder="t('operations.fields.nameZh')"
            /><el-input
              v-model="benefit.nameEn"
              :placeholder="t('operations.fields.nameEn')"
            /><el-input
              v-model="benefit.benefitValue"
              :placeholder="t('operations.fields.benefitValue')"
            /><el-select v-model="benefit.valueUnit"
              ><el-option
                v-for="item in ['AMOUNT', 'PERCENT', 'COUNT']"
                :key="item"
                :label="translatedValue(item)"
                :value="item" /></el-select
            ><el-button type="danger" link @click="removeAt(benefits, index)">{{
              t("operations.actions.remove")
            }}</el-button>
          </div>
        </section>
        <section v-if="moduleKey === 'vouchers'" class="operations-rules">
          <div class="operations-rules__head">
            <strong>{{ t("operations.tabs.rewardTiers") }}</strong>
            <el-button
              v-if="can('operation:voucher:tier:update')"
              @click="addVoucherTier"
              >{{ t("operations.actions.addRewardTier") }}</el-button
            >
          </div>
          <div
            v-for="(tier, index) in voucherRewardTiers"
            :key="tier.id || `new-${index}`"
            class="voucher-tier-row"
          >
            <el-input
              v-model="tier.amount"
              :placeholder="t('operations.fields.depositAmount')"
            />
            <el-input
              v-model="tier.bonusReward"
              :placeholder="t('operations.fields.bonusReward')"
            />
            <el-input
              v-model="tier.cashReward"
              :placeholder="t('operations.fields.cashReward')"
            />
            <el-input
              v-model="tier.rewardAmount"
              :placeholder="t('operations.fields.rewardAmount')"
            />
            <el-input-number
              v-model="tier.sort"
              :min="0"
              controls-position="right"
            />
            <el-switch v-model="tier.enabled" />
            <el-button
              type="danger"
              link
              @click="removeAt(voucherRewardTiers, index)"
              >{{ t("operations.actions.remove") }}</el-button
            >
          </div>
        </section>
      </div>
      <template #footer
        ><el-button @click="editVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button type="primary" :loading="editLoading" @click="submitForm">{{
          t("common.confirm")
        }}</el-button></template
      >
    </el-dialog>

    <el-dialog
      v-model="reasonVisible"
      :title="t('operations.dialogs.confirmAction')"
      width="480px"
      append-to-body
      ><el-form label-position="left" label-width="112px">
        <el-form-item
          v-if="reasonForm.action === 'vip-adjust'"
          :label="t('operations.fields.playerNo')"
          required
          ><el-input v-model="reasonForm.playerNo" /></el-form-item
        ><el-form-item
          v-if="reasonForm.action === 'vip-adjust'"
          :label="t('operations.fields.level')"
          required
          ><el-input v-model="reasonForm.level" /></el-form-item
        ><el-form-item
          v-if="reasonForm.action === 'schedule'"
          :label="t('operations.fields.scheduledTime')"
          required
          ><el-date-picker
            v-model="reasonForm.scheduledAt"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item
        ><el-form-item :label="t('operations.fields.operatorReason')" required
          ><el-input
            v-model="reasonForm.reason"
            type="textarea"
            :rows="3" /></el-form-item></el-form
      ><template #footer
        ><el-button @click="reasonVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="reasonLoading"
          @click="submitReason"
          >{{ t("common.confirm") }}</el-button
        ></template
      ></el-dialog
    >
    <el-dialog
      v-model="previewVisible"
      :title="t('operations.actions.preview')"
      width="min(720px, 90vw)"
      append-to-body
      ><el-image class="operations-preview" :src="previewURL" fit="contain"
    /></el-dialog>
  </div>
</template>

<style scoped>
.operations-page {
  display: grid;
  gap: 16px;
}
.operations-search {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  align-items: center;
}
.operations-search :deep(.el-form-item) {
  width: 300px;
  margin: 0;
}
.operations-search :deep(.el-form-item__label) {
  width: 112px;
  justify-content: flex-start;
}
.operations-search :deep(.el-form-item__content) {
  min-width: 0;
}
.operations-search :deep(.el-input),
.operations-search :deep(.el-select),
.operations-search :deep(.el-date-editor) {
  width: 188px;
}
.operations-search__actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}
.register-reward-form {
  max-width: 640px;
}
.register-reward-form :deep(.el-input-number),
.register-reward-form :deep(.el-textarea) {
  width: 320px;
}
.register-reward-form__readonly {
  color: var(--el-text-color-regular);
}
.register-reward-form__hint {
  margin: 8px 0 16px;
}
.register-reward-form__actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  width: 488px;
}
.music-playback-form {
  max-width: 720px;
}
.music-playback-form :deep(.el-select),
.music-playback-form :deep(.el-input) {
  width: 360px;
}
.music-playback-form__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  width: 510px;
}
.operations-thumbnail {
  width: 58px;
  height: 42px;
}
.operations-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
.operations-json {
  max-width: 100%;
  max-height: 220px;
  overflow: auto;
  white-space: pre-wrap;
  margin: 0;
  font: 12px/1.5 monospace;
}
.operations-form {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 18px;
}
.operations-form .is-wide {
  grid-column: 1/-1;
}
.operations-form :deep(.el-select),
.operations-form :deep(.el-date-editor),
.operations-form :deep(.el-input-number) {
  width: 100%;
}
.notification-audience-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}
.notification-audience-option small {
  color: var(--el-text-color-secondary);
}
.notification-channel-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.fallback-source {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.asset-field {
  display: grid;
  grid-template-columns: auto auto minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  width: 100%;
}
.asset-field__image {
  grid-column: 1/-1;
  width: 160px;
  height: 88px;
  border: 1px solid var(--admin-border);
  border-radius: 4px;
  background: #f7f8fa;
}
.asset-field audio {
  grid-column: 1/-1;
  width: min(420px, 100%);
  height: 36px;
}
.detail-resource-image {
  width: 160px;
  height: 88px;
  border: 1px solid var(--admin-border);
  border-radius: 4px;
  background: #f7f8fa;
}
.detail-resource-audio {
  width: min(420px, 100%);
  height: 36px;
}
.operations-rules {
  border-top: 1px solid var(--el-border-color-lighter);
  padding-top: 14px;
  margin-top: 6px;
}
.operations-rules__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 12px 0;
}
.operations-rule-row {
  display: grid;
  grid-template-columns: repeat(6, minmax(90px, 1fr)) auto;
  gap: 8px;
  margin-bottom: 8px;
}
.operations-rule-row--reward {
  grid-template-columns: repeat(4, minmax(140px, 1fr));
}
.voucher-tier-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(110px, 1fr)) 110px 52px auto;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
.operations-preview {
  width: 100%;
  min-height: 240px;
}
@media (max-width: 760px) {
  .operations-search :deep(.el-form-item) {
    width: 100%;
  }
  .operations-search :deep(.el-input),
  .operations-search :deep(.el-select),
  .operations-search :deep(.el-date-editor) {
    width: calc(100% - 112px);
  }
  .operations-form {
    grid-template-columns: 1fr;
  }
  .operations-form .is-wide {
    grid-column: auto;
  }
  .operations-rule-row {
    grid-template-columns: 1fr 1fr;
  }
  .voucher-tier-row {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
