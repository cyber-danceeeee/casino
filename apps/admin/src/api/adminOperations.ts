import {
  apiDelete,
  apiGet,
  apiPatch,
  apiPost,
  apiPut,
  http,
  type ApiResponse,
} from "./http";
import type { PageResult } from "./adminUsers";

export type OperationRecord = Record<string, unknown> & {
  id: number;
  isFallback?: boolean;
  canEdit?: boolean;
  canDelete?: boolean;
  visible?: boolean;
};
export type OperationDetail = Record<string, unknown>;
export type OperationQuery = Record<string, string | number | undefined>;
export type OperationPayload = Record<string, unknown>;

export type NotificationAudienceOption = {
  label: string;
  value: string;
  phone?: string;
  userNo?: string;
  vipLevel?: number;
  nameZh?: string;
  nameEn?: string;
  tagType?: string;
};

export type OperationResource =
  | "banners"
  | "announcements"
  | "promotions"
  | "bonuses"
  | "vip-levels"
  | "promo-codes"
  | "notifications"
  | "music"
  | "social-links"
  | "h5-navigation"
  | "vouchers"
  | "register-reward";

export type OperationAsset = {
  url: string;
  name: string;
  size: number;
  mimeType: string;
};

export type AssetType = "IMAGE" | "AUDIO" | "VIDEO" | "LOTTIE";
export type AssetScope = "PLATFORM" | "TENANT";
export type AssetUsage =
  | "GENERAL"
  | "LOGO"
  | "FAVICON"
  | "BANNER"
  | "BACKGROUND"
  | "BUTTON_ICON"
  | "GAME_PLACEHOLDER"
  | "POPUP"
  | "LOADING"
  | "SOUND";

export type AssetLibraryRecord = {
  id: number;
  tenantID: number;
  siteID: number;
  assetScope: AssetScope;
  assetType: AssetType;
  usageType: AssetUsage;
  displayName: string;
  originalName: string;
  objectKey: string;
  publicUrl: string;
  mimeType: string;
  fileExt: string;
  fileSize: number;
  checksumSha256: string;
  width: number;
  height: number;
  durationMs: number;
  referenceCount: number;
  isPublic: boolean;
  tags: string;
  altText: string;
  status: string;
  createdBy: number;
  createdByName: string;
  updatedBy: number;
  createdAt: string;
  updatedAt: string;
};

export type AssetReference = {
  table: string;
  column: string;
  recordID: number;
  tenantID: number;
  siteID: number;
  targetType: string;
};

export type AssetDetail = {
  overview: AssetLibraryRecord;
  references: AssetReference[];
  operationLogs?: OperationRecord[];
};

export type AssetUploadPayload = {
  assetType?: AssetType;
  usageType?: AssetUsage;
  assetScope?: AssetScope;
  displayName?: string;
  altText?: string;
  tags?: string;
};

export type AssetUpdatePayload = {
  usageType: AssetUsage;
  displayName: string;
  altText?: string;
  tags?: string;
  reason: string;
};

export type BannerActionType =
  | "NONE"
  | "OPEN_INTERNAL_PAGE"
  | "OPEN_EXTERNAL_URL"
  | "OPEN_GAME"
  | "OPEN_PROMOTION"
  | "OPEN_DEPOSIT"
  | "OPEN_SUPPORT";

export type BannerVisibility = "ALL" | "LOGGED_OUT" | "LOGGED_IN";

export type BannerConfiguration = {
  name: string;
  title: string;
  subtitle: string;
  desktopAssetID: number;
  mobileAssetID: number;
  desktopImageUrl?: string;
  mobileImageUrl?: string;
  actionType: BannerActionType;
  actionTarget: string;
  sortOrder: number;
  languageCode: string;
  countryCode: string;
  currencyCode: string;
  visibilityScope: BannerVisibility;
  startAt: string | null;
  endAt: string | null;
  enabled: boolean;
  deleted?: boolean;
};

export type SiteBannerRecord = {
  id: number;
  tenantID: number;
  siteID: number;
  name: string;
  publishStatus: "DRAFT" | "PUBLISHED";
  configVersion: number;
  draftUpdatedAt: string;
  publishedAt: string;
  draftConfiguration: BannerConfiguration;
  publishedConfiguration: BannerConfiguration;
  desktopAsset?: AssetLibraryRecord;
  mobileAsset?: AssetLibraryRecord;
	isFallback: boolean;
	canEdit: boolean;
	canDelete: boolean;
	visible: boolean;
	fallbackSourceID: number;
	fallbackImageURL: string;
	rowKey: string;
};

export type SiteBannerPayload = BannerConfiguration & { reason?: string };

export type H5NavigationAction =
  | "OPEN_INTERNAL_PAGE"
  | "OPEN_EXTERNAL_URL"
  | "OPEN_DEPOSIT"
  | "OPEN_WITHDRAW"
  | "OPEN_SUPPORT"
  | "OPEN_PROMOTION"
  | "OPEN_GAME"
  | "OPEN_DOWNLOAD";

export type H5NavigationPosition = "BOTTOM_NAV" | "HOME_QUICK" | "PROFILE_MENU";
export type H5NavigationBadge = "NONE" | "NEW" | "HOT";
export type H5NavigationVisibility = "ALL" | "GUEST" | "LOGGED_IN";

export type H5NavigationConfiguration = {
  menuCode: string;
  labels: Record<string, string>;
  iconAssetID: number;
  lottieAssetID: number;
  soundAssetID: number;
  displayPosition: H5NavigationPosition;
  sortOrder: number;
  badgeType: H5NavigationBadge;
  visibilityScope: H5NavigationVisibility;
  countryCodes: string[];
  languageCodes: string[];
  startAt: string | null;
  endAt: string | null;
  enabled: boolean;
  actionType: H5NavigationAction;
  actionTarget: string;
  providerCode: string;
  gameID: number;
  providerGameCode: string;
  gameName: string;
  deleted?: boolean;
};

export type H5NavigationRecord = {
  id: number;
  tenantID: number;
  siteID: number;
  menuCode: string;
  publishStatus: "DRAFT" | "PUBLISHED";
  configVersion: number;
  draftUpdatedAt: string | null;
  publishedAt: string | null;
	platformSourceID: number;
	isFallback: boolean;
	canEdit: boolean;
	canDelete: boolean;
	visible: boolean;
  draftConfiguration: H5NavigationConfiguration;
  publishedConfiguration: H5NavigationConfiguration;
  iconAsset?: AssetLibraryRecord;
  lottieAsset?: AssetLibraryRecord;
  soundAsset?: AssetLibraryRecord;
};

export type H5NavigationPayload = Omit<
  H5NavigationConfiguration,
  "menuCode" | "providerGameCode" | "gameName" | "deleted" | "countryCodes" | "languageCodes"
> & {
  menuCode?: string;
  labelZh: string;
  labelEn: string;
  reason?: string;
};

export type MusicPlaybackOption = {
  id: number;
  nameZh: string;
  nameEn: string;
  scene: string;
  enabled: boolean;
};

export type MusicPlaybackConfig = {
  mode: "RANDOM" | "SPECIFIED";
  specifiedMusicID: number;
  musicOptions: MusicPlaybackOption[];
};

export async function uploadOperationAsset(
  file: File,
  kind: "image" | "audio" | "video",
) {
  const form = new FormData();
  form.append("kind", kind);
  form.append("file", file);
  const { data } = await http.post<ApiResponse<OperationAsset>>(
    "/admin/operation-assets/upload",
    form,
  );
  if (data.code !== 0) throw new Error(data.msg);
  return data.data;
}

export async function uploadAssetLibrary(
  file: File,
  payload: AssetUploadPayload,
) {
  const form = new FormData();
  form.append("file", file);
  Object.entries(payload).forEach(([key, value]) => {
    if (value !== undefined && value !== "") form.append(key, String(value));
  });
  const { data } = await http.post<ApiResponse<AssetLibraryRecord>>(
    "/admin/assets/upload",
    form,
  );
  if (data.code !== 0) throw new Error(data.msg);
  return data.data;
}

export const listAssets = (query: OperationQuery) =>
  apiGet<PageResult<AssetLibraryRecord>>("/admin/assets", query);

export const getAsset = (id: number) =>
  apiGet<AssetDetail>(`/admin/assets/${id}`);

export const updateAsset = (id: number, payload: AssetUpdatePayload) =>
  apiPut<void>(`/admin/assets/${id}`, payload);

export const deleteAsset = (id: number, reason: string) =>
  apiDelete<void>(`/admin/assets/${id}`, { reason });

export const listAssetReferences = (id: number) =>
  apiGet<AssetReference[]>(`/admin/assets/${id}/references`);

export const listSiteBanners = (query: OperationQuery) =>
  apiGet<PageResult<SiteBannerRecord>>("/admin/banners", query);

export const getSiteBanner = (id: number) =>
  apiGet<{ overview: SiteBannerRecord }>(`/admin/banners/${id}`);

export const createSiteBanner = (payload: SiteBannerPayload) =>
  apiPost<{ id: number }>("/admin/banners", payload);

export const updateSiteBanner = (id: number, payload: SiteBannerPayload) =>
  apiPut<{ id: number }>(`/admin/banners/${id}`, payload);

export const copySiteBanner = (id: number) =>
  apiPost<{ id: number }>(`/admin/banners/${id}/copy`);

export const deleteSiteBanner = (id: number) =>
  apiDelete<void>(`/admin/banners/${id}`);

export const sortSiteBanners = (ids: number[], reason?: string) =>
  apiPut<void>("/admin/banners/sort", { ids, reason });

export const listH5Navigation = (query: OperationQuery) =>
  apiGet<PageResult<H5NavigationRecord>>("/admin/h5-navigation", query);

export const createH5Navigation = (payload: H5NavigationPayload) =>
  apiPost<{ id: number }>("/admin/h5-navigation", payload);

export const updateH5Navigation = (id: number, payload: H5NavigationPayload) =>
  apiPut<{ id: number }>(`/admin/h5-navigation/${id}`, payload);

export const copyH5Navigation = (id: number) =>
  apiPost<{ id: number }>(`/admin/h5-navigation/${id}/copy`);

export const deleteH5Navigation = (id: number) =>
  apiDelete<void>(`/admin/h5-navigation/${id}`);

export const sortH5Navigation = (
  displayPosition: H5NavigationPosition,
  ids: number[],
  reason?: string,
) => apiPut<void>("/admin/h5-navigation/sort", { displayPosition, ids, reason });

export const listOperations = (
  resource: OperationResource,
  query: OperationQuery,
) => apiGet<PageResult<OperationRecord>>(`/admin/${resource}`, query);

export const getOperation = (resource: OperationResource, id: number) =>
  apiGet<OperationDetail>(`/admin/${resource}/${id}`);

export const createOperation = (
  resource: OperationResource,
  payload: OperationPayload,
) => apiPost<{ id: number }>(`/admin/${resource}`, payload);

export const updateOperation = (
  resource: OperationResource,
  id: number,
  payload: OperationPayload,
) => apiPut<{ id: number }>(`/admin/${resource}/${id}`, payload);

export const deleteOperation = (resource: OperationResource, id: number) =>
  apiDelete<void>(`/admin/${resource}/${id}`);

export const setOperationDefaultVisibility = (
  resource: OperationResource,
  id: number,
  enabled: boolean,
) => apiPatch<void>(`/admin/${resource}/${id}/default-visibility`, { enabled });

export const getMusicPlaybackConfig = () =>
  apiGet<MusicPlaybackConfig>("/admin/music-settings");

export const updateMusicPlaybackConfig = (
  payload: Pick<MusicPlaybackConfig, "mode" | "specifiedMusicID"> & {
    reason?: string;
  },
) => apiPut<void>("/admin/music-settings", payload);

export const setMusicEnabled = (id: number, enabled: boolean) =>
  apiPatch<void>(`/admin/music/${id}/status`, { enabled });

export const copyBanner = (id: number) =>
  apiPost<{ id: number }>(`/admin/banners/${id}/copy`);
export const setPromotionStatus = (
  id: number,
  status: string,
  reason: string,
) => apiPatch<void>(`/admin/promotions/${id}/status`, { status, reason });
export const transitionBonus = (
  id: number,
  action: "activate" | "cancel" | "forfeit",
  reason: string,
) => apiPost<void>(`/admin/bonuses/${id}/${action}`, { reason });
export const adjustVIP = (userID: number, level: number, reason: string) =>
  apiPost<void>(`/admin/vip-users/${userID}/adjust`, { level, reason });
export const adjustVIPByNumber = (playerNo: string, level: number, reason: string) =>
  apiPost<void>(`/admin/vip-users/by-number/${encodeURIComponent(playerNo)}/adjust`, { level, reason });
export const queueNotification = (
  id: number,
  action: "schedule" | "send" | "cancel",
  payload: OperationPayload,
) => apiPost<void>(`/admin/notifications/${id}/${action}`, payload);

export const getNotificationAudienceOptions = (params: {
  audienceType: string;
  keyword?: string;
  selected?: string;
}) =>
  apiGet<NotificationAudienceOption[]>(
    "/admin/notifications/audience-options",
    params,
  );

export const saveVoucherRewardTiers = (
  id: number,
  rewardTiers: Array<Record<string, unknown>>,
) => apiPut<void>(`/admin/vouchers/${id}/reward-tiers`, { rewardTiers });

export const listVoucherUsage = (id: number, query: OperationQuery) =>
  apiGet<PageResult<OperationRecord>>(
    `/admin/vouchers/${id}/usage-records`,
    query,
  );

export type RegisterRewardConfig = {
  enabled: boolean;
  minAmount: string;
  maxAmount: string;
  auditMultiplier: string;
  updatedAt?: string;
};

export const getRegisterRewardConfig = () =>
  apiGet<RegisterRewardConfig>("/admin/register-reward");

export const updateRegisterRewardConfig = (
  payload: RegisterRewardConfig & { reason?: string },
) => apiPut<void>("/admin/register-reward", payload);
