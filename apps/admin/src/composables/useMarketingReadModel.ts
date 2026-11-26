import { ref } from "vue";
import {
  listOperations,
  listSiteBanners,
  type OperationRecord,
} from "@/api/adminOperations";
import { systemAPI } from "@/api/adminSystem";

export type CampaignReadModel = {
  rowKey: string;
  code: string;
  name: string;
  type: string;
  status: string;
  startAt: string;
  endAt: string;
  audience: string;
  placement: string;
  publishedVersion: number | string;
  updatedBy: string;
  updatedAt: string;
  path: string;
};

function text(record: Record<string, unknown>, ...keys: string[]) {
  for (const key of keys)
    if (record[key] != null && String(record[key]).trim())
      return String(record[key]);
  return "";
}
function configuration(record: Record<string, unknown>) {
  const value = record.draftConfiguration || record.publishedConfiguration;
  return value && typeof value === "object"
    ? (value as Record<string, unknown>)
    : record;
}
function operationalStatus(
  record: Record<string, unknown>,
  config: Record<string, unknown>,
) {
  if (text(record, "publishStatus") === "DRAFT") return "DRAFT";
  if (config.enabled === false || record.enabled === false) return "DISABLED";
  const now = Date.now();
  const start = Date.parse(text(config, "startAt", "startTime"));
  const end = Date.parse(text(config, "endAt", "endTime"));
  if (Number.isFinite(start) && start > now) return "SCHEDULED";
  if (Number.isFinite(end) && end < now) return "EXPIRED";
  return text(record, "status", "publishStatus") || "ACTIVE";
}
function model(
  record: OperationRecord,
  type: string,
  path: string,
  index: number,
): CampaignReadModel {
  const config = configuration(record);
  const component =
    record.component && typeof record.component === "object"
      ? (record.component as Record<string, unknown>)
      : {};
  const code =
    text(
      record,
      "campaignCode",
      "promotionCode",
      "couponCode",
      "code",
      "menuCode",
    ) || text(component, "componentCode");
  return {
    rowKey: `${type}:${record.id || index}`,
    code,
    name:
      text(
        config,
        "campaignName",
        "name",
        "nameZh",
        "nameEn",
        "title",
        "titleZh",
        "titleEn",
      ) ||
      text(record, "campaignName", "name", "nameZh", "nameEn", "title") ||
      code,
    type,
    status: operationalStatus(record, config),
    startAt: text(config, "startAt", "startTime"),
    endAt: text(config, "endAt", "endTime"),
    audience:
      text(config, "visibilityScope", "audienceType", "audience") || "ALL",
    placement:
      text(config, "position", "displayPosition", "triggerType") || "-",
    publishedVersion:
      (record.configVersion as number) ||
      text(record, "publishedVersion") ||
      "-",
    updatedBy:
      text(record, "updatedByName", "operator", "createdByName") || "-",
    updatedAt: text(record, "draftUpdatedAt", "updatedAt", "createdAt"),
    path,
  };
}

export function useMarketingReadModel() {
  const loading = ref(false);
  const items = ref<CampaignReadModel[]>([]);
  const errors = ref<string[]>([]);
  async function load() {
    loading.value = true;
    errors.value = [];
    try {
      const overview = await systemAPI.brand.overview();
      const results = await Promise.allSettled([
        listOperations("promotions", { page: 1, pageSize: 200 }),
        listOperations("vouchers", { page: 1, pageSize: 200 }),
        listOperations("vip-levels", { page: 1, pageSize: 200 }),
        listOperations("notifications", { page: 1, pageSize: 200 }),
        listSiteBanners({ page: 1, pageSize: 200 }),
        systemAPI.popups.list(overview.siteID, { page: 1, pageSize: 200 }),
        systemAPI.announcements.list(overview.siteID, {
          page: 1,
          pageSize: 200,
        }),
      ]);
      const specs = [
        ["PROMOTION", "/operations/activities"],
        ["COUPON", "/operations/vouchers"],
        ["VIP", "/operations/vip"],
        ["PUSH", "/operations/push-messages"],
        ["BANNER", "/operations/banners"],
        ["POPUP", "/operations/popups"],
        ["ANNOUNCEMENT", "/operations/announcements"],
      ];
      const merged: CampaignReadModel[] = [];
      results.forEach((result, resultIndex) => {
        if (result.status === "rejected") {
          errors.value.push(
            result.reason instanceof Error
              ? result.reason.message
              : String(result.reason),
          );
          return;
        }
        const list = (result.value as { list?: OperationRecord[] }).list || [];
        list.forEach((record, index) =>
          merged.push(
            model(record, specs[resultIndex][0], specs[resultIndex][1], index),
          ),
        );
      });
      items.value = merged.sort((left, right) =>
        right.updatedAt.localeCompare(left.updatedAt),
      );
    } finally {
      loading.value = false;
    }
  }
  return { loading, items, errors, load };
}
