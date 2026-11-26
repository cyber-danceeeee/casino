import type { PageResult } from "@/api/adminUsers";
import type { AdminOptionSource } from "@/api/adminOptions";
import {
  fetchAbnormalBets,
  fetchAbnormalBet,
  fetchBets,
  fetchBet,
  fetchCallbackLogs,
  fetchCallbackLog,
  fetchRollbacks,
  fetchRollback,
  type BettingDetail,
  type BettingQuery,
  type BettingRecord,
} from "@/api/adminBetting";

export type FilterType = "text" | "select" | "dateRange";
export type FilterField = {
  key: string;
  labelKey: string;
  type?: FilterType;
  options?: string[];
  source?: AdminOptionSource;
};
export type Column = {
  key: string;
  labelKey: string;
  width?: number;
  money?: boolean;
  time?: boolean;
  status?: boolean;
  profit?: boolean;
  imageKey?: string;
};
export type DetailTab = { key: string; titleKey: string };
export type BettingPageConfig = {
  key: string;
  searchTitleKey: string;
  tableTitleKey: string;
  permission: string;
  filters: FilterField[];
  columns: Column[];
  detailTabs: DetailTab[];
  loader: (query: BettingQuery) => Promise<PageResult<BettingRecord>>;
  detailLoader: (id: number) => Promise<BettingDetail>;
};

const option = (...values: string[]) => values;
const user = (): FilterField[] => [
  { key: "userNo", labelKey: "betting.fields.userID" },
  { key: "username", labelKey: "betting.fields.userPhone" },
];
const provider = (): FilterField => ({
  key: "providerCode",
  labelKey: "betting.fields.provider",
  type: "select",
  options: option("softaggregator", "hub88", "qtech"),
});
const created = (): FilterField => ({
  key: "createdRange",
  labelKey: "betting.fields.createdTime",
  type: "dateRange",
});

export const bettingConfigs: Record<string, BettingPageConfig> = {
  "bet-records": {
    key: "bet-records",
    searchTitleKey: "betting.pages.records.search",
    tableTitleKey: "betting.pages.records.title",
    permission: "bet:list",
    filters: [
      { key: "platformBetID", labelKey: "betting.fields.platformBetID" },
      {
        key: "providerTransactionID",
        labelKey: "betting.fields.providerTransactionID",
      },
      { key: "providerRoundID", labelKey: "betting.fields.providerRoundID" },
      ...user(),
      provider(),
      { key: "gameID", labelKey: "betting.fields.gameID" },
      { key: "providerGameID", labelKey: "betting.fields.providerGameID" },
      { key: "gameType", labelKey: "betting.fields.gameType" },
      { key: "currency", labelKey: "betting.fields.currency" },
      {
        key: "status",
        labelKey: "betting.fields.betStatus",
        type: "select",
        options: option(
          "PENDING",
          "BET_ACCEPTED",
          "SETTLED",
          "REFUNDED",
          "ROLLED_BACK",
          "CANCELLED",
          "FAILED",
        ),
      },
      { key: "amountMin", labelKey: "betting.fields.betAmountMin" },
      { key: "amountMax", labelKey: "betting.fields.betAmountMax" },
      { key: "winAmountMin", labelKey: "betting.fields.winAmountMin" },
      { key: "winAmountMax", labelKey: "betting.fields.winAmountMax" },
      created(),
      {
        key: "settledRange",
        labelKey: "betting.fields.settledTime",
        type: "dateRange",
      },
      {
        key: "rollbackStatus",
        labelKey: "betting.fields.rollbackStatus",
        type: "select",
        options: option("NONE", "ROLLED_BACK"),
      },
    ],
    columns: [
      {
        key: "platformBetID",
        labelKey: "betting.fields.platformBetID",
        width: 190,
      },
      { key: "userNo", labelKey: "betting.fields.userID", width: 145 },
      { key: "username", labelKey: "betting.fields.userPhone", width: 135 },
      {
        key: "gameName",
        labelKey: "betting.fields.gameName",
        width: 190,
        imageKey: "gameIcon",
      },
      {
        key: "providerCode",
        labelKey: "betting.fields.provider",
        width: 140,
        status: true,
      },
      {
        key: "providerRoundID",
        labelKey: "betting.fields.providerRoundID",
        width: 190,
      },
      {
        key: "betAmount",
        labelKey: "betting.fields.betAmount",
        width: 125,
        money: true,
      },
      {
        key: "validBetAmount",
        labelKey: "betting.fields.validBetAmount",
        width: 135,
        money: true,
      },
      {
        key: "winAmount",
        labelKey: "betting.fields.winAmount",
        width: 120,
        money: true,
      },
      {
        key: "platformProfit",
        labelKey: "betting.fields.platformProfit",
        width: 130,
        profit: true,
      },
      { key: "currency", labelKey: "betting.fields.currency", width: 90 },
      {
        key: "status",
        labelKey: "betting.fields.betStatus",
        width: 130,
        status: true,
      },
      {
        key: "rollbackStatus",
        labelKey: "betting.fields.rollbackStatus",
        width: 135,
        status: true,
      },
      {
        key: "betAt",
        labelKey: "betting.fields.betTime",
        width: 175,
        time: true,
      },
      {
        key: "settledAt",
        labelKey: "betting.fields.settledTime",
        width: 175,
        time: true,
      },
    ],
    detailTabs: [
      { key: "overview", titleKey: "betting.tabs.overview" },
      {
        key: "transactionTimeline",
        titleKey: "betting.tabs.transactionTimeline",
      },
      {
        key: "walletTransactions",
        titleKey: "betting.tabs.walletTransactions",
      },
      { key: "providerRequests", titleKey: "betting.tabs.providerRequests" },
      { key: "roundInformation", titleKey: "betting.tabs.roundInformation" },
      { key: "operationLogs", titleKey: "betting.tabs.operationLogs" },
    ],
    loader: fetchBets,
    detailLoader: fetchBet,
  },
  "rollback-records": {
    key: "rollback-records",
    searchTitleKey: "betting.pages.rollbacks.search",
    tableTitleKey: "betting.pages.rollbacks.title",
    permission: "bet:rollback:list",
    filters: [
      { key: "rollbackID", labelKey: "betting.fields.rollbackID" },
      {
        key: "originalTransactionID",
        labelKey: "betting.fields.originalTransactionID",
      },
      {
        key: "providerTransactionID",
        labelKey: "betting.fields.providerTransactionID",
      },
      { key: "providerRoundID", labelKey: "betting.fields.providerRoundID" },
      ...user(),
      provider(),
      {
        key: "operation",
        labelKey: "betting.fields.operation",
        type: "select",
        options: option("DEBIT", "CREDIT", "REFUND"),
      },
      { key: "currency", labelKey: "betting.fields.currency" },
      {
        key: "status",
        labelKey: "betting.fields.status",
        type: "select",
        options: option("PENDING", "COMPLETED", "FAILED", "CONFLICT"),
      },
      created(),
      {
        key: "completedRange",
        labelKey: "betting.fields.completedTime",
        type: "dateRange",
      },
    ],
    columns: [
      { key: "rollbackID", labelKey: "betting.fields.rollbackID", width: 190 },
      {
        key: "originalTransactionID",
        labelKey: "betting.fields.originalTransactionID",
        width: 190,
      },
      { key: "userNo", labelKey: "betting.fields.userID", width: 145 },
      { key: "username", labelKey: "betting.fields.userPhone", width: 135 },
      {
        key: "providerCode",
        labelKey: "betting.fields.provider",
        width: 140,
        status: true,
      },
      { key: "gameName", labelKey: "betting.fields.gameName", width: 170 },
      {
        key: "originalOperation",
        labelKey: "betting.fields.originalOperation",
        width: 145,
      },
      {
        key: "originalAmount",
        labelKey: "betting.fields.originalAmount",
        width: 130,
        money: true,
      },
      {
        key: "rollbackAmount",
        labelKey: "betting.fields.rollbackAmount",
        width: 130,
        money: true,
      },
      { key: "currency", labelKey: "betting.fields.currency", width: 90 },
      {
        key: "status",
        labelKey: "betting.fields.status",
        width: 120,
        status: true,
      },
      {
        key: "createdAt",
        labelKey: "betting.fields.createdTime",
        width: 175,
        time: true,
      },
      {
        key: "completedAt",
        labelKey: "betting.fields.completedTime",
        width: 175,
        time: true,
      },
    ],
    detailTabs: [
      { key: "overview", titleKey: "betting.tabs.overview" },
      {
        key: "originalTransaction",
        titleKey: "betting.tabs.originalTransaction",
      },
      {
        key: "rollbackTransaction",
        titleKey: "betting.tabs.rollbackTransaction",
      },
      { key: "walletEffect", titleKey: "betting.tabs.walletEffect" },
      { key: "providerRequest", titleKey: "betting.tabs.providerRequest" },
      { key: "operationLogs", titleKey: "betting.tabs.operationLogs" },
    ],
    loader: fetchRollbacks,
    detailLoader: fetchRollback,
  },
  "abnormal-bets": {
    key: "abnormal-bets",
    searchTitleKey: "betting.pages.abnormal.search",
    tableTitleKey: "betting.pages.abnormal.title",
    permission: "bet:abnormal:list",
    filters: [
      ...user(),
      provider(),
      { key: "game", labelKey: "betting.fields.gameName" },
      {
        key: "riskLevel",
        labelKey: "betting.fields.riskLevel",
        type: "select",
        options: option("LOW", "MEDIUM", "HIGH", "CRITICAL"),
      },
      {
        key: "riskType",
        labelKey: "betting.fields.riskType",
        type: "select",
        options: option(
          "HIGH_FREQUENCY_BETTING",
          "LARGE_BET",
          "HIGH_WIN_RATE",
          "ABNORMAL_WIN_AMOUNT",
          "DUPLICATE_TRANSACTION",
          "TRANSACTION_CONFLICT",
          "FREQUENT_ROLLBACK",
          "CURRENCY_MISMATCH",
          "PROVIDER_CALLBACK_ANOMALY",
          "MANUAL_REVIEW",
        ),
      },
      {
        key: "reviewStatus",
        labelKey: "betting.fields.reviewStatus",
        type: "select",
        options: option(
          "PENDING",
          "UNDER_REVIEW",
          "CONFIRMED",
          "CLEARED",
          "RESTRICTED",
        ),
      },
      { key: "assignedReviewer", labelKey: "betting.fields.assignedReviewer" },
      created(),
      {
        key: "updatedRange",
        labelKey: "betting.fields.updatedTime",
        type: "dateRange",
      },
    ],
    columns: [
      { key: "userNo", labelKey: "betting.fields.userID", width: 145 },
      { key: "username", labelKey: "betting.fields.userPhone", width: 135 },
      {
        key: "providerCode",
        labelKey: "betting.fields.provider",
        width: 140,
        status: true,
      },
      { key: "gameName", labelKey: "betting.fields.gameName", width: 160 },
      {
        key: "riskType",
        labelKey: "betting.fields.riskType",
        width: 220,
        status: true,
      },
      {
        key: "riskLevel",
        labelKey: "betting.fields.riskLevel",
        width: 110,
        status: true,
      },
      {
        key: "betAmount",
        labelKey: "betting.fields.betAmount",
        width: 120,
        money: true,
      },
      {
        key: "winAmount",
        labelKey: "betting.fields.winAmount",
        width: 120,
        money: true,
      },
      {
        key: "providerRoundID",
        labelKey: "betting.fields.providerRoundID",
        width: 185,
      },
      {
        key: "reviewStatus",
        labelKey: "betting.fields.reviewStatus",
        width: 140,
        status: true,
      },
      {
        key: "assignedReviewer",
        labelKey: "betting.fields.assignedReviewer",
        width: 150,
      },
      {
        key: "createdAt",
        labelKey: "betting.fields.createdTime",
        width: 175,
        time: true,
      },
      {
        key: "updatedAt",
        labelKey: "betting.fields.updatedTime",
        width: 175,
        time: true,
      },
    ],
    detailTabs: [
      { key: "riskSummary", titleKey: "betting.tabs.riskSummary" },
      { key: "betInformation", titleKey: "betting.tabs.betInformation" },
      {
        key: "relatedTransactions",
        titleKey: "betting.tabs.relatedTransactions",
      },
      { key: "relatedRounds", titleKey: "betting.tabs.relatedRounds" },
      {
        key: "relatedDevicesAndIPs",
        titleKey: "betting.tabs.relatedDevicesAndIPs",
      },
      { key: "reviewHistory", titleKey: "betting.tabs.reviewHistory" },
      { key: "internalNotes", titleKey: "betting.tabs.internalNotes" },
      { key: "operationLogs", titleKey: "betting.tabs.operationLogs" },
    ],
    loader: fetchAbnormalBets,
    detailLoader: fetchAbnormalBet,
  },
  "provider-callback-logs": {
    key: "provider-callback-logs",
    searchTitleKey: "betting.pages.callbacks.search",
    tableTitleKey: "betting.pages.callbacks.title",
    permission: "bet:callback-log:list",
    filters: [
      provider(),
      {
        key: "operation",
        labelKey: "betting.fields.operation",
        type: "select",
        options: option("BALANCE", "DEBIT", "CREDIT", "REFUND", "ROLLBACK"),
      },
      { key: "userPhone", labelKey: "betting.fields.userPhone" },
      { key: "providerPlayerID", labelKey: "betting.fields.providerPlayerID" },
      {
        key: "providerTransactionID",
        labelKey: "betting.fields.providerTransactionID",
      },
      { key: "providerRoundID", labelKey: "betting.fields.providerRoundID" },
      { key: "gameID", labelKey: "betting.fields.gameID" },
      {
        key: "httpMethod",
        labelKey: "betting.fields.httpMethod",
        type: "select",
        options: option("GET", "POST"),
      },
      {
        key: "signatureResult",
        labelKey: "betting.fields.signatureResult",
        type: "select",
        options: option("VALID", "INVALID", "PENDING"),
      },
      {
        key: "processingResult",
        labelKey: "betting.fields.processingResult",
        type: "select",
        options: option("SUCCESS", "FAILED", "PENDING"),
      },
      { key: "sourceIP", labelKey: "betting.fields.sourceIP" },
      created(),
    ],
    columns: [
      {
        key: "providerCode",
        labelKey: "betting.fields.provider",
        width: 140,
        status: true,
      },
      { key: "operation", labelKey: "betting.fields.operation", width: 105 },
      {
        key: "providerTransactionID",
        labelKey: "betting.fields.providerTransactionID",
        width: 190,
      },
      {
        key: "providerRoundID",
        labelKey: "betting.fields.providerRoundID",
        width: 185,
      },
      { key: "userPhone", labelKey: "betting.fields.userPhone", width: 145 },
      { key: "gameName", labelKey: "betting.fields.gameName", width: 150 },
      {
        key: "providerGameID",
        labelKey: "betting.fields.providerGameID",
        width: 170,
      },
      {
        key: "amount",
        labelKey: "betting.fields.amount",
        width: 115,
        money: true,
      },
      { key: "currency", labelKey: "betting.fields.currency", width: 90 },
      {
        key: "signatureResult",
        labelKey: "betting.fields.signatureResult",
        width: 130,
        status: true,
      },
      {
        key: "processingResult",
        labelKey: "betting.fields.processingResult",
        width: 140,
        status: true,
      },
      { key: "httpStatus", labelKey: "betting.fields.httpStatus", width: 100 },
      { key: "sourceIP", labelKey: "betting.fields.sourceIP", width: 145 },
      {
        key: "createdAt",
        labelKey: "betting.fields.createdTime",
        width: 175,
        time: true,
      },
    ],
    detailTabs: [
      { key: "overview", titleKey: "betting.tabs.overview" },
      { key: "request", titleKey: "betting.tabs.request" },
      { key: "validation", titleKey: "betting.tabs.validation" },
      { key: "normalizedCommand", titleKey: "betting.tabs.normalizedCommand" },
      { key: "walletResult", titleKey: "betting.tabs.walletResult" },
      { key: "providerResponse", titleKey: "betting.tabs.providerResponse" },
      { key: "operationLogs", titleKey: "betting.tabs.operationLogs" },
    ],
    loader: fetchCallbackLogs,
    detailLoader: fetchCallbackLog,
  },
};
