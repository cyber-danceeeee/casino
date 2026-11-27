import type { PageResult } from "@/api/adminUsers";
import {
  fetchCategories,
  fetchCategory,
  fetchCollections,
  fetchCollection,
  fetchGames,
  fetchGame,
  fetchLaunchLogs,
  fetchLaunchLog,
  fetchProviders,
  fetchProvider,
  fetchSyncLogs,
  fetchSyncLog,
  fetchTags,
  fetchTag,
  type GameCenterDetail,
  type GameCenterQuery,
  type GameCenterRecord,
} from "@/api/adminGames";
import type { AdminOptionSource } from "@/api/adminOptions";

export type GameFilter = {
  key: string;
  labelKey: string;
  type?: "text" | "select" | "dateRange";
  options?: string[];
  source?: AdminOptionSource;
};
export type GameColumn = {
  key: string;
  labelKey: string;
  width?: number;
  status?: boolean;
  time?: boolean;
  image?: boolean;
  boolean?: boolean;
  tag?: boolean;
};
export type GameTab = { key: string; titleKey: string };
export type GamePageConfig = {
  searchTitleKey: string;
  tableTitleKey: string;
  permission: string;
  filters: GameFilter[];
  columns: GameColumn[];
  detailTabs: GameTab[];
  loader: (query: GameCenterQuery) => Promise<PageResult<GameCenterRecord>>;
  detailLoader: (id: number | string) => Promise<GameCenterDetail>;
};
const options = (...values: string[]) => values;
const provider = (): GameFilter => ({
  key: "providerCode",
  labelKey: "gameCenter.fields.provider",
  type: "select",
  source: "providers",
  options: options(
    "softaggregator",
    "ag",
    "hub88",
    "qtech",
    "pg-soft",
    "jili",
    "jdb",
    "cq9",
    "saba",
  ),
});
const enabled = (): GameFilter => ({
  key: "enabled",
  labelKey: "gameCenter.fields.enabled",
  type: "select",
  options: options("true", "false"),
});
const detail = (...keys: string[]): GameTab[] =>
  keys.map((key) => ({ key, titleKey: `gameCenter.tabs.${key}` }));

export const gameConfigs: Record<string, GamePageConfig> = {
  "provider-management": {
    searchTitleKey: "gameCenter.pages.providers.search",
    tableTitleKey: "gameCenter.pages.providers.title",
    permission: "game:provider:list",
    filters: [
      {
        key: "providerCode",
        labelKey: "gameCenter.fields.providerCode",
        type: "select",
        source: "providers",
      },
      { key: "providerName", labelKey: "gameCenter.fields.providerName" },
      {
        key: "status",
        labelKey: "gameCenter.fields.status",
        type: "select",
        source: "providerStatuses",
        options: options("ENABLED", "DISABLED"),
      },
      {
        key: "environment",
        labelKey: "gameCenter.fields.environment",
        type: "select",
        source: "providerEnvironments",
        options: options("production", "sandbox"),
      },
      {
        key: "lastSyncRange",
        labelKey: "gameCenter.fields.lastSyncTime",
        type: "dateRange",
      },
    ],
    columns: [
      {
        key: "providerCode",
        labelKey: "gameCenter.fields.providerCode",
        width: 150,
      },
      {
        key: "providerName",
        labelKey: "gameCenter.fields.providerName",
        width: 160,
      },
      {
        key: "environment",
        labelKey: "gameCenter.fields.environment",
        tag: true,
        width: 115,
      },
      {
        key: "status",
        labelKey: "gameCenter.fields.enabled",
        status: true,
        width: 105,
      },
      {
        key: "apiStatus",
        labelKey: "gameCenter.fields.apiStatus",
        status: true,
        width: 120,
      },
      { key: "gameCount", labelKey: "gameCenter.fields.gameCount", width: 105 },
      {
        key: "lastSyncAt",
        labelKey: "gameCenter.fields.lastSyncTime",
        time: true,
        width: 175,
      },
      {
        key: "updatedAt",
        labelKey: "gameCenter.fields.updatedTime",
        time: true,
        width: 175,
      },
    ],
    detailTabs: detail(
      "overview",
      "apiConfiguration",
      "callbackConfiguration",
      "statistics",
      "operationLogs",
    ),
    loader: fetchProviders,
    detailLoader: (id) => fetchProvider(String(id)),
  },
  "game-list": {
    searchTitleKey: "gameCenter.pages.games.search",
    tableTitleKey: "gameCenter.pages.games.title",
    permission: "game:list",
    filters: [
      { key: "gameID", labelKey: "gameCenter.fields.gameCode" },
      { key: "gameName", labelKey: "gameCenter.fields.gameName" },
      provider(),
      { key: "providerGameID", labelKey: "gameCenter.fields.providerGameID" },
      {
        key: "providerCategory",
        labelKey: "gameCenter.fields.providerCategory",
        type: "select",
        source: "providerCategories",
      },
      {
        key: "platformCategory",
        labelKey: "gameCenter.fields.platformCategory",
        type: "select",
        source: "gameCategories",
      },
      {
        key: "lastSyncedRange",
        labelKey: "gameCenter.fields.lastSyncedTime",
        type: "dateRange",
      },
      {
        key: "gameType",
        labelKey: "gameCenter.fields.gameType",
        type: "select",
        source: "gameTypes",
      },
      enabled(),
      {
        key: "supportsDemo",
        labelKey: "gameCenter.fields.supportsDemo",
        type: "select",
        options: options("true", "false"),
      },
      {
        key: "supportsFreeSpins",
        labelKey: "gameCenter.fields.supportsFreeSpins",
        type: "select",
        options: options("true", "false"),
      },
      {
        key: "isHot",
        labelKey: "gameCenter.fields.hot",
        type: "select",
        options: options("true", "false"),
      },
      {
        key: "isFeatured",
        labelKey: "gameCenter.fields.featured",
        type: "select",
        options: options("true", "false"),
      },
      {
        key: "isNew",
        labelKey: "gameCenter.fields.new",
        type: "select",
        options: options("true", "false"),
      },
      {
        key: "country",
        labelKey: "gameCenter.fields.country",
        type: "select",
        source: "countries",
      },
      {
        key: "currency",
        labelKey: "gameCenter.fields.currency",
        type: "select",
        source: "currencies",
      },
    ],
    columns: [
      {
        key: "thumbnailUrl",
        labelKey: "gameCenter.fields.thumbnail",
        image: true,
        width: 90,
      },
      { key: "gameCode", labelKey: "gameCenter.fields.gameCode", width: 165 },
      { key: "name", labelKey: "gameCenter.fields.gameName", width: 180 },
      {
        key: "providerCode",
        labelKey: "gameCenter.fields.provider",
        tag: true,
        width: 140,
      },
      {
        key: "gameType",
        labelKey: "gameCenter.fields.gameType",
        tag: true,
        width: 120,
      },
      {
        key: "providerCategoryCode",
        labelKey: "gameCenter.fields.providerCategory",
        tag: true,
        width: 160,
      },
      {
        key: "platformCategories",
        labelKey: "gameCenter.fields.platformCategories",
        width: 240,
      },
      {
        key: "providerEnabled",
        labelKey: "gameCenter.fields.providerEnabled",
        boolean: true,
        width: 125,
      },
      {
        key: "platformEnabled",
        labelKey: "gameCenter.fields.platformEnabled",
        boolean: true,
        width: 135,
      },
      {
        key: "isHot",
        labelKey: "gameCenter.fields.hot",
        boolean: true,
        width: 95,
      },
      {
        key: "isFeatured",
        labelKey: "gameCenter.fields.featured",
        boolean: true,
        width: 110,
      },
      {
        key: "isNew",
        labelKey: "gameCenter.fields.new",
        boolean: true,
        width: 95,
      },
      { key: "sortOrder", labelKey: "gameCenter.fields.sortOrder", width: 105 },
      {
        key: "configVersion",
        labelKey: "gameCenter.fields.configVersion",
        width: 120,
      },
      {
        key: "lastSyncedAt",
        labelKey: "gameCenter.fields.lastSyncedTime",
        time: true,
        width: 175,
      },
    ],
    detailTabs: detail(
      "overview",
      "providerData",
      "platformConfiguration",
      "categoriesAndTags",
      "countryAndCurrencyRules",
      "launchConfiguration",
      "operationLogs",
    ),
    loader: fetchGames,
    detailLoader: (id) => fetchGame(Number(id)),
  },
  "game-categories": {
    searchTitleKey: "gameCenter.pages.categories.search",
    tableTitleKey: "gameCenter.pages.categories.title",
    permission: "game:category:list",
    filters: [
      { key: "code", labelKey: "gameCenter.fields.code" },
      { key: "nameZh", labelKey: "gameCenter.fields.nameZh" },
      { key: "nameEn", labelKey: "gameCenter.fields.nameEn" },
      enabled(),
    ],
    columns: [
      {
        key: "icon",
        labelKey: "gameCenter.fields.icon",
        image: true,
        width: 80,
      },
      { key: "code", labelKey: "gameCenter.fields.code", width: 150 },
      { key: "nameZh", labelKey: "gameCenter.fields.nameZh", width: 150 },
      { key: "nameEn", labelKey: "gameCenter.fields.nameEn", width: 150 },
      { key: "gameCount", labelKey: "gameCenter.fields.gameCount", width: 110 },
      { key: "sortOrder", labelKey: "gameCenter.fields.sortOrder", width: 100 },
      {
        key: "enabled",
        labelKey: "gameCenter.fields.enabled",
        boolean: true,
        width: 100,
      },
      {
        key: "updatedAt",
        labelKey: "gameCenter.fields.updatedTime",
        time: true,
        width: 175,
      },
    ],
    detailTabs: detail("overview", "games", "operationLogs"),
    loader: fetchCategories,
    detailLoader: (id) => fetchCategory(Number(id)),
  },
  "game-tags": {
    searchTitleKey: "gameCenter.pages.tags.search",
    tableTitleKey: "gameCenter.pages.tags.title",
    permission: "game:tag:list",
    filters: [
      { key: "code", labelKey: "gameCenter.fields.code" },
      { key: "nameZh", labelKey: "gameCenter.fields.nameZh" },
      { key: "nameEn", labelKey: "gameCenter.fields.nameEn" },
      enabled(),
    ],
    columns: [
      {
        key: "icon",
        labelKey: "gameCenter.fields.icon",
        image: true,
        width: 80,
      },
      { key: "code", labelKey: "gameCenter.fields.code", width: 150 },
      { key: "nameZh", labelKey: "gameCenter.fields.nameZh", width: 150 },
      { key: "nameEn", labelKey: "gameCenter.fields.nameEn", width: 150 },
      { key: "gameCount", labelKey: "gameCenter.fields.gameCount", width: 110 },
      { key: "sortOrder", labelKey: "gameCenter.fields.sortOrder", width: 100 },
      {
        key: "enabled",
        labelKey: "gameCenter.fields.enabled",
        boolean: true,
        width: 100,
      },
      {
        key: "updatedAt",
        labelKey: "gameCenter.fields.updatedTime",
        time: true,
        width: 175,
      },
    ],
    detailTabs: detail("overview", "games", "operationLogs"),
    loader: fetchTags,
    detailLoader: (id) => fetchTag(Number(id)),
  },
  "game-collections": {
    searchTitleKey: "gameCenter.pages.collections.search",
    tableTitleKey: "gameCenter.pages.collections.title",
    permission: "game:collection:list",
    filters: [
      { key: "code", labelKey: "gameCenter.fields.code" },
      { key: "nameZh", labelKey: "gameCenter.fields.nameZh" },
      { key: "nameEn", labelKey: "gameCenter.fields.nameEn" },
      enabled(),
      {
        key: "startRange",
        labelKey: "gameCenter.fields.displayTime",
        type: "dateRange",
      },
    ],
    columns: [
      {
        key: "bannerUrl",
        labelKey: "gameCenter.fields.bannerUrl",
        image: true,
        width: 100,
      },
      { key: "code", labelKey: "gameCenter.fields.code", width: 150 },
      { key: "nameZh", labelKey: "gameCenter.fields.nameZh", width: 150 },
      { key: "nameEn", labelKey: "gameCenter.fields.nameEn", width: 150 },
      { key: "gameCount", labelKey: "gameCenter.fields.gameCount", width: 110 },
      { key: "sortOrder", labelKey: "gameCenter.fields.sortOrder", width: 100 },
      {
        key: "enabled",
        labelKey: "gameCenter.fields.enabled",
        boolean: true,
        width: 100,
      },
      {
        key: "startTime",
        labelKey: "gameCenter.fields.startTime",
        time: true,
        width: 175,
      },
      {
        key: "endTime",
        labelKey: "gameCenter.fields.endTime",
        time: true,
        width: 175,
      },
      {
        key: "updatedAt",
        labelKey: "gameCenter.fields.updatedTime",
        time: true,
        width: 175,
      },
    ],
    detailTabs: detail("overview", "games", "displayRules", "operationLogs"),
    loader: fetchCollections,
    detailLoader: (id) => fetchCollection(Number(id)),
  },
  "game-sync": {
    searchTitleKey: "gameCenter.pages.sync.search",
    tableTitleKey: "gameCenter.pages.sync.title",
    permission: "game:provider:sync",
    filters: [
      provider(),
      {
        key: "status",
        labelKey: "gameCenter.fields.status",
        type: "select",
        options: options("ENABLED", "DISABLED"),
      },
    ],
    columns: [
      {
        key: "providerCode",
        labelKey: "gameCenter.fields.providerCode",
        width: 160,
      },
      {
        key: "providerName",
        labelKey: "gameCenter.fields.providerName",
        width: 180,
      },
      {
        key: "status",
        labelKey: "gameCenter.fields.enabled",
        status: true,
        width: 110,
      },
      {
        key: "apiStatus",
        labelKey: "gameCenter.fields.apiStatus",
        status: true,
        width: 130,
      },
      { key: "gameCount", labelKey: "gameCenter.fields.gameCount", width: 110 },
      {
        key: "lastSyncAt",
        labelKey: "gameCenter.fields.lastSyncTime",
        time: true,
        width: 180,
      },
    ],
    detailTabs: detail("overview", "statistics", "syncLogs"),
    loader: fetchProviders,
    detailLoader: (id) => fetchProvider(String(id)),
  },
  "sync-logs": {
    searchTitleKey: "gameCenter.pages.syncLogs.search",
    tableTitleKey: "gameCenter.pages.syncLogs.title",
    permission: "game:sync-log:view",
    filters: [
      provider(),
      {
        key: "status",
        labelKey: "gameCenter.fields.status",
        type: "select",
        options: options("RUNNING", "SUCCESS", "FAILED"),
      },
      {
        key: "startedRange",
        labelKey: "gameCenter.fields.startedTime",
        type: "dateRange",
      },
      {
        key: "completedRange",
        labelKey: "gameCenter.fields.completedTime",
        type: "dateRange",
      },
      { key: "triggeredBy", labelKey: "gameCenter.fields.triggeredBy" },
    ],
    columns: [
      { key: "syncID", labelKey: "gameCenter.fields.syncID", width: 190 },
      {
        key: "providerCode",
        labelKey: "gameCenter.fields.provider",
        width: 145,
      },
      {
        key: "status",
        labelKey: "gameCenter.fields.status",
        status: true,
        width: 110,
      },
      { key: "newCount", labelKey: "gameCenter.fields.newCount", width: 95 },
      {
        key: "updatedCount",
        labelKey: "gameCenter.fields.updatedCount",
        width: 105,
      },
      {
        key: "disabledCount",
        labelKey: "gameCenter.fields.disabledCount",
        width: 110,
      },
      {
        key: "failedCount",
        labelKey: "gameCenter.fields.failedCount",
        width: 100,
      },
      { key: "durationMs", labelKey: "gameCenter.fields.duration", width: 105 },
      {
        key: "triggeredBy",
        labelKey: "gameCenter.fields.triggeredBy",
        width: 140,
      },
      {
        key: "startedAt",
        labelKey: "gameCenter.fields.startedTime",
        time: true,
        width: 175,
      },
      {
        key: "completedAt",
        labelKey: "gameCenter.fields.completedTime",
        time: true,
        width: 175,
      },
    ],
    detailTabs: detail(
      "summary",
      "failedRecords",
      "changedGames",
      "rawProviderResponse",
      "operationLogs",
    ),
    loader: fetchSyncLogs,
    detailLoader: (id) => fetchSyncLog(Number(id)),
  },
  "game-launch-records": {
    searchTitleKey: "gameCenter.pages.launchLogs.search",
    tableTitleKey: "gameCenter.pages.launchLogs.title",
    permission: "game:launch-log:view",
    filters: [
      { key: "userID", labelKey: "gameCenter.fields.userID" },
      { key: "username", labelKey: "gameCenter.fields.username" },
      provider(),
      { key: "gameID", labelKey: "gameCenter.fields.gameCode" },
      { key: "providerGameID", labelKey: "gameCenter.fields.providerGameID" },
      {
        key: "launchType",
        labelKey: "gameCenter.fields.launchType",
        type: "select",
        options: options("REAL", "DEMO"),
      },
      {
        key: "status",
        labelKey: "gameCenter.fields.status",
        type: "select",
        options: options(
          "LAUNCH_REQUESTED",
          "LAUNCH_SUCCESS",
          "LAUNCH_FAILED",
          "PLAYING",
          "CLOSED",
        ),
      },
      { key: "currency", labelKey: "gameCenter.fields.currency" },
      { key: "language", labelKey: "gameCenter.fields.language" },
      { key: "country", labelKey: "gameCenter.fields.country" },
      { key: "device", labelKey: "gameCenter.fields.device" },
      {
        key: "createdRange",
        labelKey: "gameCenter.fields.launchTime",
        type: "dateRange",
      },
    ],
    columns: [
      { key: "launchID", labelKey: "gameCenter.fields.launchID", width: 190 },
      { key: "userNo", labelKey: "gameCenter.fields.userID", width: 150 },
      { key: "username", labelKey: "gameCenter.fields.username", width: 130 },
      { key: "gameName", labelKey: "gameCenter.fields.gameName", width: 170 },
      { key: "gameCode", labelKey: "gameCenter.fields.gameCode", width: 165 },
      {
        key: "providerCode",
        labelKey: "gameCenter.fields.provider",
        tag: true,
        width: 140,
      },
      {
        key: "country",
        labelKey: "gameCenter.fields.country",
        tag: true,
        width: 95,
      },
      {
        key: "currency",
        labelKey: "gameCenter.fields.currency",
        tag: true,
        width: 95,
      },
      {
        key: "language",
        labelKey: "gameCenter.fields.language",
        tag: true,
        width: 95,
      },
      {
        key: "status",
        labelKey: "gameCenter.fields.status",
        status: true,
        width: 155,
      },
      {
        key: "launchTime",
        labelKey: "gameCenter.fields.launchTime",
        time: true,
        width: 175,
      },
      {
        key: "exitTime",
        labelKey: "gameCenter.fields.exitTime",
        time: true,
        width: 175,
      },
      {
        key: "durationSeconds",
        labelKey: "gameCenter.fields.durationSeconds",
        width: 125,
      },
      {
        key: "exitReason",
        labelKey: "gameCenter.fields.exitReason",
        tag: true,
        width: 145,
      },
    ],
    detailTabs: detail(
      "overview",
      "sessionTimeline",
      "providerRequest",
      "providerResponse",
      "countryConfiguration",
      "operationLogs",
    ),
    loader: fetchLaunchLogs,
    detailLoader: (id) => fetchLaunchLog(Number(id)),
  },
};
