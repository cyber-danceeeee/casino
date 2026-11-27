<script setup lang="ts">
import {
  computed,
  nextTick,
  onBeforeUnmount,
  onMounted,
  reactive,
  ref,
  type Component,
  watch,
} from "vue";
import { useRoute } from "vue-router";
import { ElMessage, ElMessageBox, ElTree, type UploadRequestOptions } from "element-plus";
import * as ElementPlusIcons from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";

import {
  systemAPI,
  uploadPlatformAsset,
  type SystemDetail,
  type SystemRecord,
} from "@/api/adminSystem";
import { rowID, systemConfigs } from "./config";

const route = useRoute();
const { t, te, locale } = useI18n();
const { can, loadPermissions, reloadNavigation } = useAdminPermissions();
const moduleKey = computed(() => String(route.meta.moduleKey || "admins"));
const config = computed(() => {
  const source = systemConfigs[moduleKey.value];
  if (!source) return source;
  const internal = new Set(["siteID", "site"]);
  return { ...source, filters: source.filters.filter((field) => !internal.has(field.key)), columns: source.columns.filter((field) => !internal.has(field.key)), form: source.form?.filter((field) => !internal.has(field.key)) };
});
const query = reactive<Record<string, string | string[]>>({});
const rows = ref<SystemRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const loading = ref(false);
const detailVisible = ref(false);
const detailLoading = ref(false);
const detail = ref<SystemDetail>({});
const activeTab = ref("");
const selected = ref<SystemRecord | null>(null);
const editVisible = ref(false);
const editLoading = ref(false);
const editID = ref(0);
const form = reactive<Record<string, any>>({});
const operationVisible = ref(false);
const operationLoading = ref(false);
const operation = reactive({ type: "", reason: "", ids: "", departmentID: "" });
type PermissionTreeNode = {
  id: string | number;
  label: string;
  children?: PermissionTreeNode[];
  permissionID?: number;
};
type MenuTreeNode = {
  value: number;
  label: string;
  children?: MenuTreeNode[];
  disabled?: boolean;
};
type SelectOption = { value: string; label: string };
type DetailConfigItem = { key: string; path: string; value: unknown };
type DetailConfigGroup = { key: string; title: string; items: DetailConfigItem[] };
const permissionTree = ref<PermissionTreeNode[]>([]);
const permissionTreeRef = ref<InstanceType<typeof ElTree>>();
const permissionTreeLoading = ref(false);
const menuTreeOptions = ref<MenuTreeNode[]>([]);
const selectedMenuIDs = ref<number[]>([]);
const menuOptionsLoading = ref(false);
const permissionCodeOptions = ref<SelectOption[]>([]);
const menuEditorLoading = ref(false);
const iconComponents = ElementPlusIcons as unknown as Record<string, Component>;
const menuIconOptions = Object.keys(iconComponents).sort((left, right) =>
  left.localeCompare(right),
);
const copyForm = reactive<Record<string, unknown>>({});
const monitoring = ref<SystemDetail>({});
const monitorDetailVisible = ref(false);
const monitorDetailLoading = ref(false);
const monitorDetail = ref<SystemDetail>({});
const monitorDetailService = ref("");
const autoRefresh = ref(true);
const apkUploading = ref(false);
const platformLogoPreview = ref("");
let monitorTimer: number | undefined;
const countryImageAssetKeys = new Set([
  "logoUrl",
  "faviconUrl",
  "loginLogoUrl",
  "headerLogoUrl",
  "loadingLogoUrl",
  "appIconUrl",
  "loadingBackgroundUrl",
]);
const countryColorKeys = new Set([
  "primaryColor",
  "secondaryColor",
  "buttonColor",
  "backgroundColor",
  "textColor",
]);
const hiddenDetailIDKeys = new Set([
  "id",
  "siteid",
  "adminid",
  "roleid",
  "operatorid",
  "targetid",
  "logid",
  "operationlogid",
  "administratorid",
  "owneradminid",
]);
const configGroupOrder = [
  "brandConfig",
  "themeConfig",
  "featureConfig",
  "maintenanceConfig",
  "seoConfig",
];

const isMonitoring = computed(() => moduleKey.value === "system-monitor");
const isSiteModule = computed(() => moduleKey.value === "sites");
const operationTitle = computed(() =>
  operation.type ? t(`systemCenter.actions.${operation.type}`) : "",
);
const canCreate = computed(() =>
  Boolean(config.value?.createPermission && can(config.value.createPermission)),
);
const treeRows = computed(() => {
  if (!["departments", "menus"].includes(moduleKey.value)) return rows.value;
  const nodes = new Map<number, SystemRecord & { children?: SystemRecord[] }>();
  rows.value.forEach((row) => nodes.set(rowID(row), { ...row, children: [] }));
  const roots: Array<SystemRecord & { children?: SystemRecord[] }> = [];
  nodes.forEach((node) => {
    const parent = nodes.get(Number(node.parentID || 0));
    if (parent) parent.children?.push(node);
    else roots.push(node);
  });
  nodes.forEach((node) => {
    if (!node.children?.length) delete node.children;
  });
  return roots;
});

function label(key: string) {
  return key.split(".").map((segment) => {
    if (/^\d+$/.test(segment)) return `#${Number(segment) + 1}`;
    const path = `systemCenter.fields.${segment}`;
    if (te(path)) return t(path);
    return segment.replace(/([a-z0-9])([A-Z])/g, "$1 $2").replace(/_/g, " ");
  }).join(" / ");
}
function valueText(value: unknown) {
  const raw = String(value ?? "");
  const key = `systemCenter.values.${raw.toLowerCase()}`;
  return te(key) ? t(key) : raw || "-";
}
function normalizedKey(key: string) {
  return key.replace(/[^a-zA-Z0-9]/g, "").toLowerCase();
}
function isHiddenDetailKey(key: string, tab = "") {
  const last = key.split(".").pop() || key;
  const normalized = normalizedKey(last);
  if (tab === "operationLogs")
    return hiddenDetailIDKeys.has(normalized) || normalized === "correlationid";
  if (moduleKey.value === "sites")
    return hiddenDetailIDKeys.has(normalized);
  return false;
}
function parseJSONValue(value: unknown): unknown {
  if (typeof value !== "string") return value;
  const raw = value.trim();
  if (!raw || !/^[{[]/.test(raw)) return value;
  try {
    return JSON.parse(raw);
  } catch {
    return value;
  }
}
function structuredValue(value: unknown): unknown {
  const parsed = parseJSONValue(value);
  if (Array.isArray(parsed))
    return parsed.map((item) => structuredValue(item));
  if (isObject(parsed)) {
    return Object.fromEntries(
      Object.entries(parsed).map(([key, item]) => [key, structuredValue(item)]),
    );
  }
  return parsed;
}
function optionLabel(fieldKey: string, option: string) {
  if (fieldKey === "defaultLanguage" || fieldKey === "supportedLanguages") {
    const path = `systemCenter.languages.${option}`;
    return te(path) ? t(path) : option;
  }
  return valueText(option);
}
function splitListValue(value: unknown) {
  if (Array.isArray(value)) {
    return value.map((item) => String(item).trim()).filter(Boolean);
  }
  const raw = String(value ?? "").trim();
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) return splitListValue(parsed);
  } catch {
    // Keep comma-delimited legacy values compatible with the backend.
  }
  return raw
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}
function normalizeCountryForm() {
  if (moduleKey.value !== "country-config") return;
  form.supportedLanguages = splitListValue(form.supportedLanguages);
  if (!form.supportedLanguages.length) form.supportedLanguages = ["en"];
  if (!form.defaultLanguage) form.defaultLanguage = form.supportedLanguages[0] || "en";
  if (!form.timezone) form.timezone = "Asia/Karachi";
  if (!form.phoneCountryCode) form.phoneCountryCode = "+92";
  if (!form.phoneLength) form.phoneLength = 10;
  if (form.currencyDecimalPlaces === undefined || form.currencyDecimalPlaces === null || form.currencyDecimalPlaces === "")
    form.currencyDecimalPlaces = 2;
  if (form.countryCode) form.countryCode = String(form.countryCode).toUpperCase();
  if (form.currencyCode) form.currencyCode = String(form.currencyCode).toUpperCase();
}
function isObject(value: unknown): value is Record<string, unknown> {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}
function display(key: string, value: unknown) {
  const normalizedValue = structuredValue(value);
  if (normalizedValue === null || normalizedValue === undefined || normalizedValue === "") return "-";
  if (key.toLowerCase().endsWith("at") || key.toLowerCase().includes("time"))
    return formatBackendTime(String(normalizedValue));
  if (typeof normalizedValue === "boolean")
    return normalizedValue ? t("common.yes") : t("common.no");
  if (typeof normalizedValue === "object") return JSON.stringify(normalizedValue, null, 2);
  return valueText(normalizedValue);
}
function isByteMetric(key: string) {
  return key.toLowerCase().endsWith("bytes");
}
function formatBytes(value: unknown) {
  const bytes = Number(value || 0);
  if (!Number.isFinite(bytes) || bytes < 0) return "-";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let amount = bytes;
  let unit = 0;
  while (amount >= 1024 && unit < units.length - 1) {
    amount /= 1024;
    unit += 1;
  }
  return `${amount.toFixed(unit === 0 ? 0 : 2)} ${units[unit]}`;
}
function isStatusMetric(key: string) {
  const normalized = key.toLowerCase();
  return normalized === "status" || normalized === "enabled" || normalized.endsWith("status");
}
function enabledValue(value: unknown) {
  return value === true || value === 1 || value === "1" || String(value).toUpperCase() === "ENABLED";
}
function monitorTagType(key: string, value: unknown) {
  if (key.toLowerCase() === "enabled") return enabledValue(value) ? "success" : "danger";
  const normalized = String(value ?? "").toUpperCase();
  if (["UP", "SUCCESS", "HEALTHY", "ENABLED"].includes(normalized)) return "success";
  if (["DOWN", "FAILED", "UNAVAILABLE", "DISABLED"].includes(normalized)) return "danger";
  return "warning";
}
function monitorDisplay(key: string, value: unknown) {
  if (isByteMetric(key)) return formatBytes(value);
  if (key.toLowerCase().includes("percentage") || key.toLowerCase().endsWith("rate")) {
    const number = Number(value);
    return Number.isFinite(number) ? `${number.toFixed(2)}%` : "-";
  }
  if (key === "uptimeSeconds") {
    const seconds = Math.max(0, Number(value || 0));
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${days}${t("systemCenter.units.day")} ${hours}${t("systemCenter.units.hour")} ${minutes}${t("systemCenter.units.minute")}`;
  }
  if (key.toLowerCase() === "enabled")
    return enabledValue(value) ? valueText("enabled") : valueText("disabled");
  return display(key, value);
}
function objectEntries(value: unknown) {
  const normalizedValue = structuredValue(value);
  return isObject(normalizedValue) ? Object.entries(normalizedValue) : [];
}
function flattenedObjectEntries(value: unknown, prefix = ""): Array<[string, unknown]> {
  const normalizedValue = structuredValue(value);
  if (Array.isArray(normalizedValue)) {
    if (!normalizedValue.length) return prefix ? [[prefix, []]] : [];
    return normalizedValue.flatMap((item, index) => flattenedObjectEntries(item, prefix ? `${prefix}.${index}` : String(index)));
  }
  if (isObject(normalizedValue)) {
    const entries = Object.entries(normalizedValue);
    if (!entries.length) return prefix ? [[prefix, {}]] : [];
    return entries.flatMap(([key, item]) => flattenedObjectEntries(item, prefix ? `${prefix}.${key}` : key));
  }
  return prefix ? [[prefix, normalizedValue]] : [];
}
function detailObjectEntries(tab: string) {
  const entries = moduleKey.value === "sites"
    ? flattenedObjectEntries(tabValue(tab))
    : objectEntries(tabValue(tab));
  return entries.filter(([key]) => {
    if (isHiddenDetailKey(key, tab)) return false;
    if (moduleKey.value === "admins" && tab === "overview")
      return !["id", "administratorID"].includes(key);
    if (tab === "operationLogs")
      return !["id", "logID", "operationLogID", "targetID"].includes(key);
    if (moduleKey.value === "menus" && tab === "overview")
      return key !== "componentPath";
    return true;
  });
}
function resolveMenuIcon(value: unknown) {
  return iconComponents[String(value || "")] || null;
}
function isCountryImageAssetKey(key: string) {
  return moduleKey.value === "country-config" && countryImageAssetKeys.has(key);
}
function isCountryColorKey(key: string) {
  return moduleKey.value === "country-config" && countryColorKeys.has(key);
}
function isLottieAssetKey(key: string) {
  return moduleKey.value === "country-config" && key === "loadingAnimationUrl";
}
function isSupportedLanguagesKey(key: string) {
  return moduleKey.value === "country-config" && key === "supportedLanguages";
}
function statusPermission() {
  return config.value?.statusPermission || config.value?.updatePermission || "";
}
function canChangeStatus(row: SystemRecord) {
  const permission = statusPermission();
  return Boolean(config.value?.status) && (!permission || can(permission)) && !row.systemProtected;
}
function isCountryDefaultColumn(key: string) {
  return moduleKey.value === "country-config" && key === "isDefault";
}
function isCountryEnabledColumn(key: string) {
  return moduleKey.value === "country-config" && key === "enabled";
}
function booleanSwitchDisabled(row: SystemRecord, key: string) {
  if (isCountryEnabledColumn(key)) return !canChangeStatus(row);
  if (isCountryDefaultColumn(key))
    return !can("system:country:set-default") || Boolean(row.isDefault) || !Boolean(row.enabled ?? row.status);
  if (key === "enabled" && config.value?.status) return !canChangeStatus(row);
  return true;
}
function colorValue(value: unknown) {
  return String(value || "").trim();
}
function isColorLike(value: unknown) {
  return /^#(?:[0-9a-f]{3}|[0-9a-f]{6}|[0-9a-f]{8})$/i.test(colorValue(value));
}
function isConfigBooleanValue(value: unknown) {
  return typeof value === "boolean";
}
function resultTagType(value: unknown) {
  const normalized = String(value ?? "").toUpperCase();
  if (["SUCCESS", "SUCCEEDED", "PASS", "PASSED", "OK", "1", "TRUE"].includes(normalized))
    return "success";
  if (["FAILED", "FAIL", "ERROR", "REJECTED", "0", "FALSE"].includes(normalized))
    return "danger";
  return "warning";
}
function isResultKey(key: string) {
  return normalizedKey(key) === "result" || normalizedKey(key).endsWith("result");
}
function resultText(value: unknown) {
  const normalized = String(value ?? "").toUpperCase();
  if (["SUCCESS", "SUCCEEDED", "PASS", "PASSED", "OK", "1", "TRUE"].includes(normalized))
    return valueText("success");
  if (["FAILED", "FAIL", "ERROR", "REJECTED", "0", "FALSE"].includes(normalized))
    return valueText("failed");
  return display("result", value);
}
function siteConfigGroups(value: unknown): DetailConfigGroup[] {
  const normalizedValue = structuredValue(value);
  if (!isObject(normalizedValue)) return [];

  const entries = Object.entries(normalizedValue)
    .filter(([key]) => !isHiddenDetailKey(key))
    .map(([key, item]) => [key, structuredValue(item)] as [string, unknown]);

  const ordered = entries.sort(([left], [right]) => {
    const leftIndex = configGroupOrder.indexOf(left);
    const rightIndex = configGroupOrder.indexOf(right);
    if (leftIndex >= 0 || rightIndex >= 0)
      return (leftIndex >= 0 ? leftIndex : 99) - (rightIndex >= 0 ? rightIndex : 99);
    return left.localeCompare(right);
  });

  return ordered.map(([key, item]) => {
    const itemEntries = isObject(item) || Array.isArray(item)
      ? flattenedObjectEntries(item)
      : [[key, item] as [string, unknown]];
    return {
      key,
      title: label(key),
      items: itemEntries
        .filter(([childKey]) => !isHiddenDetailKey(childKey))
        .map(([childKey, childValue]) => ({
          key: childKey,
          path: childKey,
          value: childValue,
        })),
    };
  }).filter((group) => group.items.length);
}
function isSiteConfigTab(tab: string) {
  return isSiteModule.value && ["configuration"].includes(tab);
}
function countryAssetURL(value: unknown) {
  return String(value || "").trim();
}
function clearAssetField(key: string) {
  form[key] = "";
}
function loadingBackgroundFallbackColor() {
  return colorValue(form.backgroundColor) || "#003f77";
}
function detailBackgroundFallbackColor() {
  const theme = tabValue("theme");
  return isObject(theme) ? colorValue(theme.backgroundColor) || "#003f77" : "#003f77";
}
function arrayColumns(value: unknown, tab = "") {
  return Array.isArray(value) && value.length && isObject(value[0])
    ? Object.keys(value[0]).filter((key) => {
        if (isHiddenDetailKey(key, tab)) return false;
        if (tab === "administrators") return !key.toLowerCase().endsWith("id");
        if (tab === "operationLogs") return !["id", "logID", "operationLogID", "targetID"].includes(key);
        return key !== "targetID";
      }).slice(0, 12)
    : [];
}
function tabValue(tab: string) {
  return detail.value[tab];
}
function normalizedQuery() {
  const out: Record<string, string | number> = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const [key, raw] of Object.entries(query)) {
    if (Array.isArray(raw) && raw.length === 2) {
      const base = key.replace("Range", "");
      out[`${base}From`] = selectedTimeToBackend(raw[0], false);
      out[`${base}To`] = selectedTimeToBackend(raw[1], true);
    } else if (String(raw ?? "").trim()) {
      let value = String(raw).trim();
      if (key === "status")
        value = value === "ENABLED" ? "1" : value === "DISABLED" ? "0" : value;
      if (key === "isDefault")
        value = value === "YES" ? "1" : value === "NO" ? "0" : value;
      if (key === "loginResult")
        value = value === "SUCCESS" ? "1" : value === "FAILED" ? "0" : value;
      out[key] = value;
    }
  }
  return out;
}
async function load() {
  if (isMonitoring.value) {
    await loadMonitoring();
    return;
  }
  if (!config.value || !can(config.value.permission)) return;
  loading.value = true;
  try {
    const result = await config.value.loader(normalizedQuery());
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch {
    ElMessage.error(t("systemCenter.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}
async function loadMonitoring() {
  if (!can("system:monitoring:view")) return;
  loading.value = true;
  try {
    monitoring.value = await systemAPI.monitoring.get();
  } catch {
    ElMessage.error(t("systemCenter.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}
async function viewMonitoringService(service: string) {
  monitorDetailService.value = service;
  monitorDetailVisible.value = true;
  monitorDetailLoading.value = true;
  try {
    monitorDetail.value = await systemAPI.monitoring.service(service);
  } catch {
    monitorDetailVisible.value = false;
    ElMessage.error(t("systemCenter.messages.detailFailed"));
  } finally {
    monitorDetailLoading.value = false;
  }
}
function monitoringServiceForSection(section: string) {
  if (section === "runtime") return "runtime";
  if (section === "providers") return "providers";
  return "application";
}
function monitorRowService(section: string, row: Record<string, unknown>) {
  if (section === "services") return String(row.service || "application");
  if (section === "providers") return "providers";
  return monitoringServiceForSection(section);
}
function search() {
  page.value = 1;
  load();
}
function reset() {
  Object.keys(query).forEach((key) => delete query[key]);
  page.value = 1;
  load();
}

async function view(row: SystemRecord) {
  if (!config.value?.detail) return;
  selected.value = row;
  detailVisible.value = true;
  detailLoading.value = true;
  activeTab.value = config.value.tabs[0] || "overview";
  try {
    detail.value = await config.value.detail(rowID(row));
  } catch {
    detailVisible.value = false;
    ElMessage.error(t("systemCenter.messages.detailFailed"));
  } finally {
    detailLoading.value = false;
  }
}
function resetForm() {
  Object.keys(form).forEach((key) => delete form[key]);
  for (const field of config.value?.form || []) {
    form[field.key] =
      moduleKey.value === "menus" && field.key === "parentID"
        ? null
        : field.type === "boolean"
        ? true
        : field.type === "number"
          ? 0
          : field.type === "multiSelect"
            ? []
          : field.type === "json"
            ? "{}"
            : "";
  }
  normalizeCountryForm();
}
async function openCreate() {
  editID.value = 0;
  resetForm();
  editVisible.value = true;
  if (moduleKey.value === "menus") await loadMenuEditorOptions();
}
async function openEdit(row: SystemRecord) {
  editID.value = rowID(row);
  resetForm();
  selected.value = row;
  editVisible.value = true;
  if (moduleKey.value === "menus") await loadMenuEditorOptions(editID.value);
  if (moduleKey.value === "system-config") {
    Object.assign(form, {
      value: row.value === "******" ? "" : row.value,
      reason: "",
    });
    if (String(row.key) === "basic.platform_logo_url") {
      platformLogoPreview.value = (await systemAPI.platformBranding().catch(() => undefined))?.platformLogoUrl || "";
    }
    return;
  }
  if (!config.value?.detail) {
    Object.assign(form, row);
    return;
  }
  editLoading.value = true;
  try {
    const data = await config.value.detail(editID.value);
    Object.assign(form, isObject(data.overview) ? data.overview : row);
    if (isSiteModule.value && isObject(data.configuration))
      Object.assign(form, data.configuration);
    for (const field of config.value?.form || []) {
      if (field.type === "json" && typeof form[field.key] === "object")
        form[field.key] = JSON.stringify(form[field.key], null, 2);
    }
    normalizeCountryForm();
    form.reason = "";
  } finally {
    editLoading.value = false;
  }
}
function payload() {
  const out: Record<string, unknown> = {};
  for (const field of config.value?.form || []) {
    if (editID.value && field.createOnly) continue;
    let value = form[field.key];
    if (field.type === "number")
      value = value === "" || value === null ? null : Number(value);
    if (field.type === "multiSelect")
      value = Array.isArray(value) ? value.join(",") : String(value || "");
    if (field.type === "color" || field.type === "asset" || field.type === "text")
      value = String(value || "").trim();
    if (field.type === "json") {
      try {
        value = JSON.parse(String(value || "{}"));
      } catch {
        throw new Error("INVALID_JSON");
      }
    }
    out[field.key] = value;
    if (field.type === "apk") {
      out.androidAPKObjectKey = String(form.androidAPKObjectKey || "").trim();
      out.androidAPKName = String(form.androidAPKName || "").trim();
      out.androidAPKSize = Number(form.androidAPKSize || 0);
    }
  }
  if (moduleKey.value === "country-config") {
    const languages = splitListValue(out.supportedLanguages);
    if (!languages.includes(String(out.defaultLanguage || "")))
      languages.unshift(String(out.defaultLanguage || "en"));
    out.supportedLanguages = languages.join(",");
    out.phoneCountryCode = String(out.phoneCountryCode || "+92");
    out.phoneLength = Number(out.phoneLength || 10);
    out.currencyDecimalPlaces = Number(out.currencyDecimalPlaces ?? 2);
  }
  if (
    editID.value &&
    ["departments", "menus", "scheduled-tasks"].includes(moduleKey.value)
  )
    out.reason = String(
      form.reason || t("systemCenter.messages.updatedFromAdmin"),
    );
  return out;
}
async function uploadAsset(fieldKey: string, kind: "image" | "lottie", options: UploadRequestOptions) {
  try {
    const asset = await systemAPI.countries.uploadAsset(options.file, kind);
    form[fieldKey] = asset.url;
    options.onSuccess(asset);
    ElMessage.success(t("systemCenter.messages.assetUploaded"));
  } catch (uploadError) {
    options.onError(uploadError as Parameters<UploadRequestOptions["onError"]>[0]);
    ElMessage.error(t("systemCenter.messages.assetUploadFailed"));
  }
}
function assetUploadRequest(field: { key: string; assetKind?: "image" | "lottie" }) {
  return (options: UploadRequestOptions) =>
    uploadAsset(field.key, field.assetKind || "image", options);
}
async function uploadAPK(options: UploadRequestOptions) {
  apkUploading.value = true;
  try {
    const asset = await systemAPI.sites.uploadAPK(options.file);
    form.androidAPKObjectKey = asset.key;
    form.androidAPKName = asset.name;
    form.androidAPKSize = asset.size;
    options.onSuccess(asset);
    ElMessage.success(t("systemCenter.messages.apkUploaded"));
  } catch (uploadError) {
    options.onError(uploadError as Parameters<UploadRequestOptions["onError"]>[0]);
    ElMessage.error(t("systemCenter.messages.assetUploadFailed"));
  } finally { apkUploading.value = false; }
}
function clearAPK() {
  form.androidAPKObjectKey = "";
  form.androidAPKName = "";
  form.androidAPKSize = 0;
}
async function uploadPlatformLogo(options: UploadRequestOptions) {
  try {
    const asset = await uploadPlatformAsset(options.file);
    form.value = asset.key;
    platformLogoPreview.value = asset.url;
    options.onSuccess(asset);
    ElMessage.success(t("systemCenter.messages.assetUploaded"));
  } catch (uploadError) {
    options.onError(uploadError as Parameters<UploadRequestOptions["onError"]>[0]);
    ElMessage.error(t("systemCenter.messages.assetUploadFailed"));
  }
}
async function save() {
  editLoading.value = true;
  try {
    if (moduleKey.value === "system-config") {
      await systemAPI.configs.update(
        String(selected.value?.key || ""),
        String(form.value || ""),
        String(form.reason || ""),
      );
    } else if (editID.value && config.value?.update)
      await config.value.update(editID.value, payload());
    else if (config.value?.create) await config.value.create(payload());
    editVisible.value = false;
    ElMessage.success(
      t(isSiteModule.value ? ("systemCenter.messages.siteSaved") : "systemCenter.messages.saved"),
    );
    if (moduleKey.value === "menus") await reloadNavigation();
    load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error && error.message === "INVALID_JSON"
        ? t("systemCenter.validation.invalidJson")
        : t("systemCenter.messages.operationFailed"),
    );
  } finally {
    editLoading.value = false;
  }
}

async function reasonPrompt(titleKey: string) {
  const result = await ElMessageBox.prompt(
    t("systemCenter.messages.reasonPrompt"),
    t(titleKey),
    {
      confirmButtonText: t("common.confirm"),
      cancelButtonText: t("common.cancel"),
      inputValidator: (value) =>
        Boolean(value.trim()) || t("systemCenter.validation.reasonRequired"),
    },
  );
  return result.value.trim();
}
async function toggle(row: SystemRecord, desired?: boolean) {
  if (!config.value?.status) return;
  const current = Boolean(row.enabled ?? row.status);
  const next = desired ?? !current;
  if (next === current) return;
  try {
    const reason = await reasonPrompt("systemCenter.actions.changeStatus");
    await config.value.status(rowID(row), next, reason);
    ElMessage.success(t("systemCenter.messages.operationSucceeded"));
    if (moduleKey.value === "menus") await reloadNavigation();
    load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(t("systemCenter.messages.operationFailed"));
  }
}
async function handleBooleanSwitch(row: SystemRecord, key: string, value: boolean | string | number) {
  if (isCountryEnabledColumn(key)) {
    await toggle(row, Boolean(value));
    return;
  }
  if (isCountryDefaultColumn(key) && Boolean(value) && !row.isDefault) {
    await setDefaultCountry(row);
  }
}
function handleBooleanChange(row: SystemRecord, key: string, value: unknown) {
  void handleBooleanSwitch(row, key, value as boolean | string | number);
}
async function setDefaultCountry(row: SystemRecord) {
  try {
    const reason = await reasonPrompt("systemCenter.actions.setDefaultCountry");
    await systemAPI.countries.setDefault(rowID(row), reason);
    ElMessage.success(t("systemCenter.messages.operationSucceeded"));
    load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(t("systemCenter.messages.operationFailed"));
  }
}
async function remove(row: SystemRecord) {
  if (!config.value?.remove) return;
  try {
    const reason = await reasonPrompt("common.delete");
    await config.value.remove(rowID(row), reason);
    ElMessage.success(t("systemCenter.messages.operationSucceeded"));
    if (moduleKey.value === "menus") await reloadNavigation();
    load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(t("systemCenter.messages.operationFailed"));
  }
}
function permissionName(row: SystemRecord) {
  const localized = String(
    (locale.value === "zh-CN" ? row.nameZh : row.nameEn) || "",
  );
  return localized || String(row.code || "-");
}
function menuName(row: SystemRecord) {
  const localized = String(
    (locale.value === "zh-CN" ? row.nameZh : row.nameEn) || "",
  );
  return localized || String(row.code || "-");
}
async function loadAllMenus(enabledOnly = false) {
  const all: SystemRecord[] = [];
  let current = 1;
  let totalRows = 0;
  do {
    const result = await systemAPI.menus.list({
      page: current,
      pageSize: 100,
      ...(enabledOnly ? { status: "1" } : {}),
    });
    all.push(...result.list);
    totalRows = result.total;
    current += 1;
  } while (all.length < totalRows);
  return all;
}
function buildMenuTree(rows: SystemRecord[], disabledRootID = 0): MenuTreeNode[] {
  const nodes = new Map<number, MenuTreeNode>();
  const parentIDs = new Map<number, number>();
  rows.forEach((row) => {
    const id = rowID(row);
    nodes.set(id, {
      value: id,
      label: menuName(row),
      children: [],
      disabled: id === disabledRootID,
    });
    parentIDs.set(id, Number(row.parentID || 0));
  });
  const roots: MenuTreeNode[] = [];
  nodes.forEach((node, id) => {
    const parent = nodes.get(parentIDs.get(id) || 0);
    if (parent) parent.children?.push(node);
    else roots.push(node);
  });
  function disableDescendants(node: MenuTreeNode, disabled: boolean) {
    if (disabled) node.disabled = true;
    node.children?.forEach((child) =>
      disableDescendants(child, disabled || child.value === disabledRootID),
    );
    if (!node.children?.length) delete node.children;
  }
  roots.forEach((node) => disableDescendants(node, node.value === disabledRootID));
  return roots;
}
async function loadAllPermissionOptions() {
  const all: SystemRecord[] = [];
  let current = 1;
  let totalRows = 0;
  do {
    const result = await systemAPI.permissions.list({
      page: current,
      pageSize: 100,
      status: "1",
    });
    all.push(...result.list);
    totalRows = result.total;
    current += 1;
  } while (all.length < totalRows);
  return all.map((row) => ({
    value: String(row.code || ""),
    label: `${permissionName(row)} (${String(row.code || "-")})`,
  })).filter((option) => option.value);
}
async function loadMenuEditorOptions(disabledRootID = 0) {
  menuEditorLoading.value = true;
  try {
    const [menus, permissions] = await Promise.all([
      loadAllMenus(),
      loadAllPermissionOptions(),
    ]);
    menuTreeOptions.value = buildMenuTree(menus, disabledRootID);
    permissionCodeOptions.value = permissions;
  } catch {
    ElMessage.error(t("systemCenter.messages.loadFailed"));
  } finally {
    menuEditorLoading.value = false;
  }
}
function buildPermissionTree(rows: SystemRecord[]): PermissionTreeNode[] {
  const modules = new Map<string, Map<string, PermissionTreeNode[]>>();
  rows.forEach((row) => {
    const module = String(row.module || "other");
    const type = String(row.permissionType || "ACTION");
    if (!modules.has(module)) modules.set(module, new Map());
    const types = modules.get(module)!;
    if (!types.has(type)) types.set(type, []);
    types
      .get(type)!
      .push({
        id: Number(row.id),
        permissionID: Number(row.id),
        label: permissionName(row),
      });
  });
  return Array.from(modules.entries()).map(([module, types]) => ({
    id: `module:${module}`,
    label: valueText(module),
    children: Array.from(types.entries()).map(([type, children]) => ({
      id: `module:${module}:type:${type}`,
      label: valueText(type),
      children,
    })),
  }));
}
function permissionRowsForTab(tab: string): SystemRecord[] {
  const value = tabValue(tab);
  if (moduleKey.value === "roles" && tab === "permissions" && Array.isArray(value))
    return value.filter(isObject) as SystemRecord[];
  if (moduleKey.value === "admins" && tab === "rolesAndPermissions" && isObject(value)) {
    return Array.isArray(value.permissions)
      ? value.permissions.filter(isObject) as SystemRecord[]
      : [];
  }
  return [];
}
function isPermissionTreeTab(tab: string) {
  return permissionRowsForTab(tab).length > 0 ||
    (moduleKey.value === "roles" && tab === "permissions") ||
    (moduleKey.value === "admins" && tab === "rolesAndPermissions");
}
function detailPermissionTree(tab: string) {
  return buildPermissionTree(permissionRowsForTab(tab));
}
function detailRoles(tab: string): SystemRecord[] {
  const value = tabValue(tab);
  return moduleKey.value === "admins" && tab === "rolesAndPermissions" && isObject(value) && Array.isArray(value.roles)
    ? value.roles.filter(isObject) as SystemRecord[]
    : [];
}
async function loadPermissionTree(roleID: number) {
  permissionTreeLoading.value = true;
  try {
    const all: SystemRecord[] = [];
    let current = 1;
    let totalRows = 0;
    do {
      const result = await systemAPI.permissions.list({
        page: current,
        pageSize: 100,
        status: "1",
      });
      all.push(...result.list);
      totalRows = result.total;
      current += 1;
    } while (all.length < totalRows);
    permissionTree.value = buildPermissionTree(all);
    const role = await systemAPI.roles.detail(roleID);
    const assigned = Array.isArray(role.permissions)
      ? role.permissions
          .map((item) => Number(isObject(item) ? item.id : 0))
          .filter((id) => id > 0)
      : [];
    await nextTick();
    permissionTreeRef.value?.setCheckedKeys(assigned);
  } catch {
    ElMessage.error(t("systemCenter.messages.loadFailed"));
  } finally {
    permissionTreeLoading.value = false;
  }
}
async function loadRoleMenuTree(roleID: number) {
  menuOptionsLoading.value = true;
  try {
    const [menus, role] = await Promise.all([
      loadAllMenus(true),
      systemAPI.roles.detail(roleID),
    ]);
    menuTreeOptions.value = buildMenuTree(menus);
    selectedMenuIDs.value = Array.isArray(role.menus)
      ? role.menus
          .map((item) => Number(isObject(item) ? item.id : 0))
          .filter((id) => id > 0)
      : [];
  } catch {
    ElMessage.error(t("systemCenter.messages.loadFailed"));
  } finally {
    menuOptionsLoading.value = false;
  }
}
async function openOperation(type: string, row: SystemRecord) {
  selected.value = row;
  Object.assign(operation, { type, reason: "", ids: "", departmentID: "" });
  operationVisible.value = true;
  if (type === "assign-permissions") await loadPermissionTree(rowID(row));
  if (type === "assign-menus") await loadRoleMenuTree(rowID(row));
}
async function openCopy(type: "copy-role" | "copy-menu", row: SystemRecord) {
  selected.value = row;
  operationLoading.value = true;
  try {
    const source =
      type === "copy-role"
        ? await systemAPI.roles.detail(rowID(row))
        : await systemAPI.menus.detail(rowID(row));
    const overview = isObject(source.overview) ? source.overview : row;
    Object.keys(copyForm).forEach((key) => delete copyForm[key]);
    Object.assign(copyForm, overview, {
      code: `${String(overview.code || "copy")}-copy-${Date.now().toString().slice(-6)}`,
      nameZh:
        `${String(overview.nameZh || "")} ${t("systemCenter.values.copySuffix")}`.trim(),
      nameEn: `${String(overview.nameEn || "")} Copy`.trim(),
      enabled: true,
    });
    Object.assign(operation, { type, reason: "", ids: "", departmentID: "" });
    operationVisible.value = true;
  } catch {
    ElMessage.error(t("systemCenter.messages.detailFailed"));
  } finally {
    operationLoading.value = false;
  }
}
function copyFields() {
  return ["roles", "menus"].includes(moduleKey.value)
    ? config.value?.form || []
    : [];
}
function copyPayload() {
  const out: Record<string, unknown> = {};
  for (const field of copyFields()) {
    let value = copyForm[field.key];
    if (field.type === "number")
      value = value === "" || value === null ? null : Number(value);
    out[field.key] = value;
  }
  out.reason = operation.reason.trim();
  return out;
}
async function submitOperation() {
  if (!selected.value || !operation.reason.trim()) {
    ElMessage.warning(t("systemCenter.validation.reasonRequired"));
    return;
  }
  const id = rowID(selected.value);
  const ids =
    operation.type === "assign-permissions"
      ? (permissionTreeRef.value?.getCheckedKeys(true) || [])
          .map(Number)
          .filter((value) => value > 0)
      : operation.type === "assign-menus"
        ? selectedMenuIDs.value
        : operation.ids
          .split(",")
          .map((value) => Number(value.trim()))
          .filter((value) => value > 0);
  operationLoading.value = true;
  try {
    switch (operation.type) {
      case "reset-password": {
        const data = await systemAPI.admins.resetPassword(id, operation.reason);
        await ElMessageBox.alert(
          data.temporaryPassword,
          t("systemCenter.dialogs.temporaryPassword"),
          { confirmButtonText: t("common.confirm") },
        );
        break;
      }
      case "reset-mfa":
        await systemAPI.admins.resetMFA(id, operation.reason);
        break;
      case "force-logout":
        await systemAPI.admins.forceLogout(id, operation.reason);
        break;
      case "assign-roles":
        await systemAPI.admins.roles(id, ids, operation.reason);
        break;
      case "move-department":
        await systemAPI.admins.department(
          id,
          operation.departmentID ? Number(operation.departmentID) : null,
          operation.reason,
        );
        break;
      case "assign-permissions":
        await systemAPI.roles.permissions(id, ids, operation.reason);
        break;
      case "assign-menus":
        await systemAPI.roles.menus(id, ids, operation.reason);
        break;
      case "run-task":
        await systemAPI.tasks.run(id, operation.reason);
        break;
      case "copy-role":
        await systemAPI.roles.copy(id, copyPayload());
        break;
      case "copy-menu":
        await systemAPI.menus.copy(id, copyPayload());
        break;
    }
    operationVisible.value = false;
    ElMessage.success(t("systemCenter.messages.operationSucceeded"));
    load();
  } catch {
    ElMessage.error(t("systemCenter.messages.operationFailed"));
  } finally {
    operationLoading.value = false;
  }
}
function operationNeedsIDs() {
  return operation.type === "assign-roles";
}
function operationNeedsDepartment() {
  return operation.type === "move-department";
}
function operationIsCopy() {
  return ["copy-role", "copy-menu"].includes(operation.type);
}


function setupMonitorTimer() {
  if (monitorTimer) window.clearInterval(monitorTimer);
  monitorTimer = undefined;
  if (isMonitoring.value && autoRefresh.value)
    monitorTimer = window.setInterval(loadMonitoring, 15000);
}
watch(autoRefresh, setupMonitorTimer);
watch(moduleKey, async () => {
  Object.keys(query).forEach((key) => delete query[key]);
  page.value = 1;
  pageSize.value = moduleKey.value === "menus" ? 100 : 20;
  detailVisible.value = false;
  await nextTick();
  setupMonitorTimer();
  load();
});
onMounted(async () => {
  try {
    await loadPermissions();
  } catch {
    ElMessage.error(t("systemCenter.messages.loadFailed"));
    return;
  }
  pageSize.value = moduleKey.value === "menus" ? 100 : 20;
  setupMonitorTimer();
  load();
});
onBeforeUnmount(() => {
  if (monitorTimer) window.clearInterval(monitorTimer);
});
</script>

<template>
  <div class="system-page">
    <template v-if="isMonitoring">
      <UserCenterCard :title="t('systemCenter.pages.systemMonitoring')">
        <template #header>
          <el-switch
            v-model="autoRefresh"
            :active-text="t('systemCenter.actions.autoRefresh')"
          />
          <el-button type="primary" @click="loadMonitoring">{{
            t("systemCenter.actions.refresh")
          }}</el-button>
        </template>
        <div v-loading="loading" class="monitor-grid">
          <el-card
            v-for="(section, sectionKey) in monitoring"
            :key="sectionKey"
            shadow="never"
          >
            <template #header
              ><div class="monitor-card-header">
                <span>{{ label(String(sectionKey)) }}</span
                ><el-button
                  link
                  type="primary"
                  @click="
                    viewMonitoringService(
                      monitoringServiceForSection(String(sectionKey)),
                    )
                  "
                  >{{ t("common.view") }}</el-button
                >
              </div></template
            >
            <el-descriptions v-if="isObject(section)" :column="1" border>
              <el-descriptions-item
                v-for="[key, value] in objectEntries(section)"
                :key="key"
                :label="label(key)"
              >
                <div v-if="isObject(value)" class="monitor-metric-list">
                  <div
                    v-for="[childKey, childValue] in objectEntries(value)"
                    :key="childKey"
                    class="monitor-metric-list__row"
                  >
                    <span>{{ label(childKey) }}</span>
                    <el-tag
                      v-if="isStatusMetric(childKey)"
                      :type="monitorTagType(childKey, childValue)"
                      size="small"
                    >{{ monitorDisplay(childKey, childValue) }}</el-tag>
                    <strong v-else>{{ monitorDisplay(childKey, childValue) }}</strong>
                  </div>
                </div>
                <el-tag
                  v-else-if="isStatusMetric(key)"
                  :type="monitorTagType(key, value)"
                  size="small"
                >{{ monitorDisplay(key, value) }}</el-tag>
                <span v-else>{{ monitorDisplay(key, value) }}</span>
              </el-descriptions-item>
            </el-descriptions>
            <el-table
              v-else-if="Array.isArray(section)"
              :data="section"
              border
              max-height="360"
            >
              <el-table-column
                v-for="key in arrayColumns(section)"
                :key="key"
                :prop="key"
                :label="label(key)"
                min-width="130"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <el-tag
                    v-if="isStatusMetric(key)"
                    :type="monitorTagType(key, row[key])"
                    size="small"
                  >{{ monitorDisplay(key, row[key]) }}</el-tag>
                  <span v-else>{{ monitorDisplay(key, row[key]) }}</span>
                </template>
              </el-table-column>
              <el-table-column
                fixed="right"
                :label="t('common.actions')"
                width="92"
              >
                <template #default="{ row }"
                  ><el-button
                    link
                    type="primary"
                    @click="
                      viewMonitoringService(
                        monitorRowService(String(sectionKey), row),
                      )
                    "
                    >{{ t("common.view") }}</el-button
                  ></template
                >
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </UserCenterCard>
    </template>

    <template v-else-if="config">
      <UserCenterCard :title="t('systemCenter.sections.search')">
        <el-form
          class="search-form"
          label-position="left"
          label-width="132px"
          @submit.prevent="search"
        >
          <el-row :gutter="16">
            <el-col
              v-for="field in config.filters"
              :key="field.key"
              :xs="24"
              :sm="12"
              :lg="8"
              :xl="6"
            >
              <el-form-item :label="label(field.key)">
                <el-select
                  v-if="field.type === 'select'"
                  v-model="query[field.key]"
                  clearable
                >
                  <el-option
                    v-for="option in field.options"
                    :key="option"
                    :label="valueText(option)"
                    :value="option"
                  />
                </el-select>
                <el-date-picker
                  v-else-if="field.type === 'dateRange'"
                  v-model="query[field.key]"
                  type="datetimerange"
                  value-format="YYYY-MM-DD HH:mm:ss"
                  :start-placeholder="t('systemCenter.placeholders.start')"
                  :end-placeholder="t('systemCenter.placeholders.end')"
                />
                <el-input
                  v-else
                  v-model="query[field.key]"
                  clearable
                  :placeholder="t('systemCenter.placeholders.enter')"
                  @keyup.enter="search"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <div class="form-actions">
            <el-button @click="reset">{{ t("common.reset") }}</el-button
            ><el-button type="primary" native-type="submit">{{
              t("common.search")
            }}</el-button>
          </div>
        </el-form>
      </UserCenterCard>

      <UserCenterCard
        :title="
          t(
            `menu.system.${moduleKey === 'system-config' ? 'systemConfig' : moduleKey === 'country-config' ? 'countryConfig' : moduleKey === 'admin-login-logs' ? 'loginLogs' : moduleKey === 'scheduled-tasks' ? 'scheduledTasks' : moduleKey === 'audit-logs' ? 'auditLogs' : moduleKey}`,
          )
        "
      >
        <template #header
          ><el-button
            v-if="canCreate"
            type="primary"
            @click="openCreate()"
          >{{
            t("common.create")
          }}</el-button></template
        >
        <el-table
          v-loading="loading"
          :data="treeRows"
          :row-key="(row: SystemRecord) => rowID(row)"
          :tree-props="{ children: 'children' }"
          :default-expand-all="['departments', 'menus'].includes(moduleKey)"
          border
          stripe
        >
          <el-table-column
            fixed="left"
            type="index"
            :index="(index: number) => (page - 1) * pageSize + index + 1"
            :label="t('common.index')"
            width="76"
          />
          <el-table-column
            v-for="column in config.columns"
            :key="column.key"
            :prop="column.key"
            :label="label(column.key)"
            :width="column.width"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <el-switch
                v-if="column.kind === 'boolean'"
                :model-value="Boolean(row[column.key])"
                :disabled="booleanSwitchDisabled(row, column.key)"
                @change="handleBooleanChange(row, column.key, $event)"
              />
              <UserStatusTag
                v-else-if="column.kind === 'status'"
                :status="String(row[column.key] || 'UNKNOWN')"
                :label="valueText(row[column.key])"
              />
              <span v-else-if="column.kind === 'time'">{{
                formatBackendTime(String(row[column.key] || ""))
              }}</span>
              <div
                v-else-if="moduleKey === 'menus' && column.key === 'icon'"
                class="menu-icon-cell"
              >
                <el-icon v-if="resolveMenuIcon(row[column.key])" :size="20">
                  <component :is="resolveMenuIcon(row[column.key])" />
                </el-icon>
                <span v-else>-</span>
              </div>
              <span v-else>{{ display(column.key, row[column.key]) }}</span>
            </template>
          </el-table-column>
          <el-table-column
            fixed="right"
            :label="t('common.actions')"
            :width="isSiteModule ? 390 : 290"
          >
            <template #default="{ row }">
              <el-button
                v-if="config.detail"
                link
                type="primary"
                @click="view(row)"
                >{{ t("common.view") }}</el-button
              >
              <el-button
                v-if="
                  config.updatePermission &&
                  can(config.updatePermission) &&
                  !config.readOnly &&
                  !row.systemProtected
                "
                link
                @click="openEdit(row)"
                >{{ t("common.edit") }}</el-button
              >


              <el-dropdown v-if="!config.readOnly" trigger="click">
                <el-button link>{{ t("systemCenter.actions.more") }}</el-button>
                <template #dropdown
                  ><el-dropdown-menu>
                    <el-dropdown-item
                      v-if="canChangeStatus(row)"
                      @click="toggle(row)"
                      >{{
                        valueText(
                          Boolean(row.enabled ?? row.status)
                            ? "DISABLE"
                            : "ENABLE",
                        )
                      }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="
                        moduleKey === 'country-config' &&
                        can('system:country:set-default') &&
                        !row.isDefault &&
                        Boolean(row.enabled ?? row.status)
                      "
                      @click="setDefaultCountry(row)"
                      >{{ t("systemCenter.actions.setDefaultCountry") }}</el-dropdown-item
                    >
                    <template v-if="moduleKey === 'admins'">
                      <el-dropdown-item
                        v-if="can('system:admin:assign-role')"
                        @click="openOperation('assign-roles', row)"
                        >{{
                          t("systemCenter.actions.assignRoles")
                        }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('system:admin:update')"
                        @click="openOperation('move-department', row)"
                        >{{
                          t("systemCenter.actions.moveDepartment")
                        }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('system:admin:reset-password')"
                        @click="openOperation('reset-password', row)"
                        >{{
                          t("systemCenter.actions.resetPassword")
                        }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('system:admin:reset-mfa')"
                        @click="openOperation('reset-mfa', row)"
                        >{{
                          t("systemCenter.actions.resetMFA")
                        }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('system:admin:force-logout')"
                        @click="openOperation('force-logout', row)"
                        >{{
                          t("systemCenter.actions.forceLogout")
                        }}</el-dropdown-item
                      >
                    </template>
                    <template v-if="moduleKey === 'roles'">
                      <el-dropdown-item
                        v-if="can('system:role:assign-permission')"
                        @click="openOperation('assign-permissions', row)"
                        >{{
                          t("systemCenter.actions.assignPermissions")
                        }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('system:role:assign-menu')"
                        @click="openOperation('assign-menus', row)"
                        >{{
                          t("systemCenter.actions.assignMenus")
                        }}</el-dropdown-item
                      >
                      <el-dropdown-item
                        v-if="can('system:role:create')"
                        @click="openCopy('copy-role', row)"
                        >{{
                          t("systemCenter.actions.copyRole")
                        }}</el-dropdown-item
                      >
                    </template>
                    <el-dropdown-item
                      v-if="moduleKey === 'menus' && can('system:menu:create')"
                      @click="openCopy('copy-menu', row)"
                      >{{
                        t("systemCenter.actions.copyMenu")
                      }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="
                        moduleKey === 'scheduled-tasks' &&
                        can('system:task:run')
                      "
                      @click="openOperation('run-task', row)"
                      >{{ t("systemCenter.actions.runNow") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="
                        config.remove &&
                        config.deletePermission &&
                        can(config.deletePermission) &&
                        !row.systemProtected
                      "
                      divided
                      @click="remove(row)"
                      >{{ t("common.delete") }}</el-dropdown-item
                    >
                  </el-dropdown-menu></template
                >
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>
        <div class="pagination">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            :total="total"
            @current-change="load"
            @size-change="
              () => {
                page = 1;
                load();
              }
            "
          />
        </div>
      </UserCenterCard>
    </template>

    <el-dialog
      v-model="detailVisible"
      :title="t('systemCenter.dialogs.detail')"
      width="min(1080px,94vw)"
      destroy-on-close
    >
      <div v-loading="detailLoading">

        <el-tabs v-model="activeTab" type="card">
          <el-tab-pane
            v-for="tab in config?.tabs || []"
            :key="tab"
            :name="tab"
            :label="t(`systemCenter.tabs.${tab}`)"
          >
            <div v-if="isPermissionTreeTab(tab)" class="permission-detail">
              <div v-if="detailRoles(tab).length" class="permission-detail__roles">
                <span class="permission-detail__caption">{{ label("roles") }}</span>
                <el-tag
                  v-for="role in detailRoles(tab)"
                  :key="String(role.id || role.code)"
                  type="primary"
                  effect="plain"
                >
                  {{ locale === "zh-CN" ? role.nameZh : role.nameEn }}
                </el-tag>
              </div>
              <el-tree
                v-if="detailPermissionTree(tab).length"
                class="permission-detail__tree"
                :data="detailPermissionTree(tab)"
                node-key="id"
                default-expand-all
                :expand-on-click-node="false"
              />
              <el-empty v-else :description="t('systemCenter.messages.empty')" />
            </div>
            <div v-else-if="isSiteConfigTab(tab)" class="site-config-groups">
              <section
                v-for="group in siteConfigGroups(tabValue(tab))"
                :key="group.key"
                class="site-config-group"
              >
                <h3>{{ group.title }}</h3>
                <dl class="site-config-group__items">
                  <template v-for="item in group.items" :key="item.path">
                    <dt>{{ label(item.key) }}</dt>
                    <dd>
                      <div v-if="isColorLike(item.value)" class="config-color-value">
                        <span
                          class="config-color-value__swatch"
                          :style="{ background: colorValue(item.value) }"
                        />
                        <span>{{ display(item.key, item.value) }}</span>
                      </div>
                      <el-switch
                        v-else-if="isConfigBooleanValue(item.value)"
                        :model-value="Boolean(item.value)"
                        disabled
                        inline-prompt
                        :active-text="valueText('enable')"
                        :inactive-text="valueText('disable')"
                      />
                      <pre v-else-if="isObject(structuredValue(item.value)) || Array.isArray(structuredValue(item.value))">{{ display(item.key, item.value) }}</pre>
                      <span v-else>{{ display(item.key, item.value) }}</span>
                    </dd>
                  </template>
                </dl>
              </section>
              <el-empty
                v-if="!siteConfigGroups(tabValue(tab)).length"
                :description="t('systemCenter.messages.empty')"
              />
            </div>
            <el-table
              v-else-if="Array.isArray(tabValue(tab))"
              :data="tabValue(tab)"
              max-height="440"
              border
            >
              <el-table-column
                v-for="key in arrayColumns(tabValue(tab), tab)"
                :key="key"
                :prop="key"
                :label="label(key)"
                min-width="140"
                show-overflow-tooltip
                ><template #default="{ row }">
                  <el-switch
                    v-if="tab === 'administrators' && ['status', 'enabled'].includes(key)"
                    :model-value="enabledValue(row[key])"
                    disabled
                  />
                  <el-tag
                    v-else-if="isResultKey(key)"
                    :type="resultTagType(row[key])"
                    size="small"
                  >{{ resultText(row[key]) }}</el-tag>
                  <span v-else>{{ display(key, row[key]) }}</span>
                </template></el-table-column
              >
            </el-table>
            <el-descriptions
              v-else-if="isObject(structuredValue(tabValue(tab)))"
              :column="2"
              border
              ><el-descriptions-item
                v-for="[key, value] in detailObjectEntries(tab)"
                :key="key"
                :label="label(key)"
              >
                <div v-if="isCountryImageAssetKey(key)" class="country-detail-asset">
                  <el-image
                    v-if="countryAssetURL(value)"
                    class="country-detail-asset__image"
                    :src="countryAssetURL(value)"
                    fit="contain"
                    :preview-src-list="[countryAssetURL(value)]"
                    preview-teleported
                  />
                  <div
                    v-else-if="key === 'loadingBackgroundUrl'"
                    class="country-detail-asset__background"
                    :style="{ background: detailBackgroundFallbackColor() }"
                  />
                  <span v-else>-</span>
                </div>
                <div v-else-if="isLottieAssetKey(key)" class="country-detail-asset">
                  <el-tag v-if="countryAssetURL(value)" type="success">
                    {{ t("systemCenter.messages.assetConfigured") }}
                  </el-tag>
                  <span v-else>-</span>
                </div>
                <div v-else-if="isCountryColorKey(key)" class="country-color-value">
                  <span
                    class="country-color-value__swatch"
                    :style="{ background: colorValue(value) || 'transparent' }"
                  />
                  <span>{{ display(key, value) }}</span>
                </div>
                <div v-else-if="isSupportedLanguagesKey(key)" class="country-language-tags">
                  <el-tag
                    v-for="item in splitListValue(value)"
                    :key="item"
                    size="small"
                  >
                    {{ optionLabel(key, item) }}
                  </el-tag>
                </div>
                <div
                  v-else-if="moduleKey === 'menus' && key === 'icon'"
                  class="menu-icon-detail"
                >
                  <el-icon v-if="resolveMenuIcon(value)" :size="24">
                    <component :is="resolveMenuIcon(value)" />
                  </el-icon>
                  <span>{{ display(key, value) }}</span>
                </div>
                <div v-else-if="isColorLike(value)" class="config-color-value">
                  <span
                    class="config-color-value__swatch"
                    :style="{ background: colorValue(value) }"
                  />
                  <span>{{ display(key, value) }}</span>
                </div>
                <el-switch
                  v-else-if="isConfigBooleanValue(value)"
                  :model-value="Boolean(value)"
                  disabled
                  inline-prompt
                  :active-text="valueText('enable')"
                  :inactive-text="valueText('disable')"
                />
                <el-tag
                  v-else-if="isResultKey(key)"
                  :type="resultTagType(value)"
                  size="small"
                >{{ resultText(value) }}</el-tag>
                <span v-else>{{ display(key, value) }}</span>
              </el-descriptions-item
              ></el-descriptions
            >
            <el-card
              v-else-if="tabValue(tab) !== null && tabValue(tab) !== undefined"
              shadow="never"
            >
              <pre>{{ display(tab, tabValue(tab)) }}</pre>
            </el-card>
            <el-empty v-else :description="t('systemCenter.messages.empty')" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>



    <el-dialog
      v-model="editVisible"
      :title="t(editID ? 'common.edit' : 'common.create')"
      width="min(780px,94vw)"
      destroy-on-close
    >

      <el-form
        v-loading="editLoading || menuEditorLoading"
        label-position="left"
        label-width="170px"
      >
        <el-form-item
          v-for="field in config?.form.filter(
            (item) => !(editID && item.createOnly),
          ) || []"
          :key="field.key"
          :label="label(field.key)"
        >
          <div v-if="moduleKey === 'system-config' && selected?.key === 'basic.platform_logo_url' && field.key === 'value'" class="asset-field">
            <el-image v-if="platformLogoPreview" class="asset-field__image" :src="platformLogoPreview" fit="contain" :preview-src-list="[platformLogoPreview]" preview-teleported/>
            <el-upload :show-file-list="false" :http-request="uploadPlatformLogo" accept="image/*,.svg,.avif"><el-button>{{ platformLogoPreview ? t('systemCenter.actions.replaceAsset') : t('systemCenter.actions.uploadAsset') }}</el-button></el-upload>
          </div>
          <el-switch
            v-else-if="field.type === 'boolean'"
            v-model="form[field.key]"
          />
          <el-tree-select
            v-else-if="moduleKey === 'menus' && field.key === 'parentID'"
            v-model="form[field.key]"
            :data="menuTreeOptions"
            check-strictly
            clearable
            filterable
            default-expand-all
            :placeholder="t('systemCenter.placeholders.selectParentMenu')"
          />
          <el-select
            v-else-if="moduleKey === 'menus' && field.key === 'icon'"
            v-model="form[field.key]"
            clearable
            filterable
            :placeholder="t('systemCenter.placeholders.selectIcon')"
          >
            <el-option
              v-for="iconName in menuIconOptions"
              :key="iconName"
              :label="iconName"
              :value="iconName"
            >
              <div class="menu-icon-option">
                <el-icon :size="18"><component :is="resolveMenuIcon(iconName)" /></el-icon>
                <span>{{ iconName }}</span>
              </div>
            </el-option>
          </el-select>
          <el-select
            v-else-if="moduleKey === 'menus' && field.key === 'permissionCode'"
            v-model="form[field.key]"
            clearable
            filterable
            :placeholder="t('systemCenter.placeholders.selectPermission')"
          >
            <el-option
              v-for="option in permissionCodeOptions"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </el-select>
          <el-select
            v-else-if="field.type === 'select'"
            v-model="form[field.key]"
            filterable
            ><el-option
              v-for="option in field.options"
              :key="option"
              :label="optionLabel(field.key, option)"
              :value="option"
          /></el-select>
          <el-select
            v-else-if="field.type === 'multiSelect'"
            v-model="form[field.key]"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
          >
            <el-option
              v-for="option in field.options"
              :key="option"
              :label="optionLabel(field.key, option)"
              :value="option"
            />
          </el-select>
          <el-input-number
            v-else-if="field.type === 'number'"
            v-model="form[field.key]"
            :min="0"
            controls-position="right"
          />
          <div v-else-if="field.type === 'color'" class="color-field">
            <el-color-picker v-model="form[field.key]" />
            <el-input v-model="form[field.key]" clearable />
          </div>
          <el-input
            v-else-if="field.type === 'json'"
            v-model="form[field.key]"
            type="textarea"
            :rows="6"
          />
          <div v-else-if="field.type === 'asset'" class="asset-field">
            <div class="asset-field__preview">
              <el-image
                v-if="isCountryImageAssetKey(field.key) && countryAssetURL(form[field.key])"
                class="asset-field__image"
                :src="countryAssetURL(form[field.key])"
                fit="contain"
                :preview-src-list="[countryAssetURL(form[field.key])]"
                preview-teleported
              />
              <div
                v-else-if="field.key === 'loadingBackgroundUrl'"
                class="asset-field__background"
                :style="{ background: loadingBackgroundFallbackColor() }"
              >
                {{ t("systemCenter.messages.loadingBackgroundFallback") }}
              </div>
              <div v-else-if="isLottieAssetKey(field.key) && countryAssetURL(form[field.key])" class="asset-field__lottie">
                <el-tag type="success">Lottie</el-tag>
                <span>{{ t("systemCenter.messages.assetConfigured") }}</span>
              </div>
              <el-empty
                v-else
                class="asset-field__empty"
                :description="t('systemCenter.messages.noAsset')"
              />
            </div>
            <el-upload
              :show-file-list="false"
              :http-request="assetUploadRequest(field)"
              :accept="field.assetKind === 'lottie' ? '.json,application/json' : 'image/*,.svg,.ico,.avif'"
            >
              <el-button>{{
                form[field.key]
                  ? t("systemCenter.actions.replaceAsset")
                  : t("systemCenter.actions.uploadAsset")
              }}</el-button>
            </el-upload>
            <el-button
              v-if="form[field.key]"
              @click="clearAssetField(field.key)"
            >
              {{ t("common.delete") }}
            </el-button>
          </div>
          <div v-else-if="field.type === 'apk'" class="asset-field">
            <el-tag v-if="form.androidAPKObjectKey" type="success" effect="plain">{{ form.androidAPKName || t('systemCenter.messages.apkConfigured') }}</el-tag>
            <el-upload :show-file-list="false" :http-request="uploadAPK" accept=".apk,application/vnd.android.package-archive">
              <el-button :loading="apkUploading">{{ form.androidAPKObjectKey ? t('systemCenter.actions.replaceAPK') : t('systemCenter.actions.uploadAPK') }}</el-button>
            </el-upload>
            <el-button v-if="form.androidAPKObjectKey" @click="clearAPK">{{ t('common.delete') }}</el-button>
          </div>
          <el-input
            v-else
            v-model="form[field.key]"
            :type="field.type === 'password' ? 'password' : 'text'"
            :show-password="field.type === 'password'"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer
        ><el-button @click="editVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button type="primary" :loading="editLoading" @click="save">{{
          t("common.save")
        }}</el-button></template
      >
    </el-dialog>

    <el-dialog
      v-model="operationVisible"
      :title="operationTitle"
      width="min(560px,92vw)"
      destroy-on-close
    >
      <el-form label-position="left" label-width="150px">
        <template v-if="operationIsCopy()">
          <el-form-item
            v-for="field in copyFields()"
            :key="field.key"
            :label="label(field.key)"
          >
            <el-switch
              v-if="field.type === 'boolean'"
              v-model="copyForm[field.key]"
            />
            <el-select
              v-else-if="field.type === 'select'"
              v-model="copyForm[field.key]"
              ><el-option
                v-for="option in field.options"
                :key="option"
                :label="valueText(option)"
                :value="option"
            /></el-select>
            <el-input-number
              v-else-if="field.type === 'number'"
              v-model="copyForm[field.key]"
              :min="0"
            />
            <el-input v-else v-model="copyForm[field.key]" />
          </el-form-item>
        </template>
        <el-form-item v-if="operationNeedsIDs()" :label="label('selectedIDs')"
          ><el-input
            v-model="operation.ids"
            :placeholder="t('systemCenter.placeholders.commaSeparatedIDs')"
        /></el-form-item>
        <el-form-item
          v-if="operation.type === 'assign-permissions'"
          :label="label('selectedPermissions')"
        >
          <div v-loading="permissionTreeLoading" class="permission-tree">
            <el-tree
              ref="permissionTreeRef"
              :data="permissionTree"
              node-key="id"
              show-checkbox
              default-expand-all
              :props="{ label: 'label', children: 'children' }"
            />
          </div>
        </el-form-item>
        <el-form-item
          v-if="operation.type === 'assign-menus'"
          :label="label('selectedMenus')"
        >
          <el-cascader
            v-model="selectedMenuIDs"
            v-loading="menuOptionsLoading"
            class="menu-cascader"
            :options="menuTreeOptions"
            :props="{
              multiple: true,
              emitPath: false,
              checkStrictly: true,
              value: 'value',
              label: 'label',
              children: 'children',
            }"
            clearable
            filterable
            collapse-tags
            collapse-tags-tooltip
            :placeholder="t('systemCenter.placeholders.selectMenus')"
          />
        </el-form-item>
        <el-form-item
          v-if="operationNeedsDepartment()"
          :label="label('departmentID')"
          ><el-input v-model="operation.departmentID"
        /></el-form-item>
        <el-form-item :label="label('reason')"
          ><el-input v-model="operation.reason" type="textarea" :rows="3"
        /></el-form-item>
      </el-form>
      <template #footer
        ><el-button @click="operationVisible = false">{{
          t("common.cancel")
        }}</el-button
        ><el-button
          type="primary"
          :loading="operationLoading"
          @click="submitOperation"
          >{{ t("common.confirm") }}</el-button
        ></template
      >
    </el-dialog>

    <el-dialog
      v-model="monitorDetailVisible"
      :title="`${t('systemCenter.dialogs.monitoringDetail')} - ${label(monitorDetailService)}`"
      width="min(960px,94vw)"
      destroy-on-close
    >
      <div v-loading="monitorDetailLoading" class="monitor-detail">
        <el-card
          v-for="[key, value] in objectEntries(monitorDetail)"
          :key="key"
          shadow="never"
        >
          <template #header>{{ label(key) }}</template>
          <el-descriptions v-if="isObject(value)" :column="1" border>
            <el-descriptions-item
              v-for="[childKey, childValue] in objectEntries(value)"
              :key="childKey"
              :label="label(childKey)"
            >
              <div v-if="isObject(childValue)" class="monitor-metric-list">
                <div
                  v-for="[metricKey, metricValue] in objectEntries(childValue)"
                  :key="metricKey"
                  class="monitor-metric-list__row"
                >
                  <span>{{ label(metricKey) }}</span>
                  <el-tag
                    v-if="isStatusMetric(metricKey)"
                    :type="monitorTagType(metricKey, metricValue)"
                    size="small"
                  >{{ monitorDisplay(metricKey, metricValue) }}</el-tag>
                  <strong v-else>{{ monitorDisplay(metricKey, metricValue) }}</strong>
                </div>
              </div>
              <el-tag
                v-else-if="isStatusMetric(childKey)"
                :type="monitorTagType(childKey, childValue)"
                size="small"
              >{{ monitorDisplay(childKey, childValue) }}</el-tag>
              <span v-else>{{ monitorDisplay(childKey, childValue) }}</span>
            </el-descriptions-item>
          </el-descriptions>
          <el-table v-else-if="Array.isArray(value)" :data="value" border>
            <el-table-column
              v-for="columnKey in arrayColumns(value)"
              :key="columnKey"
              :prop="columnKey"
              :label="label(columnKey)"
              min-width="130"
            >
              <template #default="{ row }">
                <el-tag
                  v-if="isStatusMetric(columnKey)"
                  :type="monitorTagType(columnKey, row[columnKey])"
                  size="small"
                >{{ monitorDisplay(columnKey, row[columnKey]) }}</el-tag>
                <span v-else>{{ monitorDisplay(columnKey, row[columnKey]) }}</span>
              </template>
            </el-table-column>
          </el-table>
          <span v-else>{{ monitorDisplay(key, value) }}</span>
        </el-card>
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.system-page {
  display: grid;
  gap: 16px;
}
.search-form :deep(.el-select),
.search-form :deep(.el-date-editor),
.search-form :deep(.el-input) {
  width: 100%;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.permission-detail {
  min-height: 220px;
}
.permission-detail__roles {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 14px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.permission-detail__caption {
  margin-right: 4px;
  color: var(--el-text-color-secondary);
}
.permission-detail__tree {
  padding: 8px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
}
.monitor-metric-list {
  display: grid;
  gap: 8px;
}
.monitor-metric-list__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.monitor-metric-list__row > span {
  color: var(--el-text-color-secondary);
}
.pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
.tab-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}
.site-config-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  margin-bottom: 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  background: var(--el-fill-color-extra-light);
}
.site-config-bar__meta,
.site-config-bar__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 14px;
}
.site-config-bar__meta {
  color: var(--el-text-color-secondary);
}
.site-config-groups {
  display: grid;
  gap: 14px;
}
.site-config-group {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  background: var(--el-fill-color-extra-light);
}
.site-config-group h3 {
  margin: 0;
  padding: 11px 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 700;
}
.site-config-group__items {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}
.site-config-group__items dt,
.site-config-group__items dd {
  min-width: 0;
  margin: 0;
  padding: 10px 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.site-config-group__items dt {
  color: var(--el-text-color-secondary);
}
.site-config-group__items dd {
  overflow-wrap: anywhere;
}
.site-config-group__items dt:nth-last-child(2),
.site-config-group__items dd:last-child {
  border-bottom: 0;
}
.config-color-value {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.config-color-value__swatch {
  width: 20px;
  height: 20px;
  flex: 0 0 auto;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  box-shadow: inset 0 0 0 1px rgb(255 255 255 / 45%);
}
.provision-result {
  margin-top: 16px;
}
.delivery-card {
  margin-top: 18px;
  padding: 16px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-extra-light);
}
.delivery-card__header,
.delivery-card__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.delivery-card__header p {
  margin: 0 0 6px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.delivery-card__header h3 {
  margin: 0;
}
.delivery-card__actions {
  justify-content: flex-end;
  margin-top: 14px;
}
.delivery-password {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-weight: 700;
  color: var(--el-color-danger);
}
.monitor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
.monitor-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.monitor-detail {
  display: grid;
  gap: 12px;
}
.permission-tree {
  width: 100%;
  max-height: 420px;
  overflow: auto;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  padding: 8px 12px;
}
.menu-cascader,
.menu-icon-cell,
.menu-icon-detail {
  width: 100%;
}
.menu-icon-cell,
.menu-icon-detail,
.menu-icon-option {
  display: flex;
  align-items: center;
  gap: 8px;
}
.menu-icon-option {
  min-width: 0;
}
.menu-icon-option span {
  overflow: hidden;
  text-overflow: ellipsis;
}
.asset-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  gap: 8px;
  width: 100%;
}
.asset-field__preview {
  grid-column: 1 / -1;
  min-height: 86px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  background: var(--el-fill-color-extra-light);
}
.asset-field__image {
  width: 144px;
  height: 72px;
  border-radius: 4px;
  background: repeating-conic-gradient(#f4f6fa 0 25%, #e8edf5 0 50%) 50% / 16px 16px;
}
.asset-field__background {
  width: 100%;
  min-height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  color: #fff;
  text-shadow: 0 1px 2px rgb(0 0 0 / 42%);
}
.asset-field__lottie {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
  overflow-wrap: anywhere;
}
.asset-field__empty {
  width: 100%;
  padding: 0;
}
.asset-field__empty :deep(.el-empty__image) {
  display: none;
}
.asset-field :deep(.el-input) {
  min-width: 0;
}
.color-field {
  display: grid;
  grid-template-columns: auto minmax(0, 220px);
  align-items: center;
  gap: 10px;
  width: 100%;
}
.country-detail-asset {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}
.country-detail-asset__image {
  width: 92px;
  height: 48px;
  flex: 0 0 auto;
  border-radius: 4px;
  border: 1px solid var(--el-border-color-lighter);
  background: repeating-conic-gradient(#f4f6fa 0 25%, #e8edf5 0 50%) 50% / 14px 14px;
}
.country-detail-asset__background {
  width: 120px;
  height: 48px;
  flex: 0 0 auto;
  border-radius: 4px;
  border: 1px solid var(--el-border-color-lighter);
}
.country-color-value,
.country-language-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.country-color-value__swatch {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid var(--el-border-color);
}
pre {
  margin: 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font: inherit;
}
@media (max-width: 900px) {
  .monitor-grid {
    grid-template-columns: 1fr;
  }
  .site-config-group__items {
    grid-template-columns: 1fr;
  }
  .site-config-group__items dt {
    padding-bottom: 2px;
    border-bottom: 0;
  }
  .site-config-group__items dd {
    padding-top: 2px;
  }
}
</style>
