export type RiskField = {
  key: string;
  type?:
    | "text"
    | "select"
    | "dateRange"
    | "number"
    | "switch"
    | "textarea"
    | "datetime";
  options?: string[];
  required?: boolean;
};
export type RiskColumn = {
  key: string;
  width?: number;
  status?: boolean;
  time?: boolean;
  masked?: boolean;
};
export type RiskConfig = {
  resource: string;
  listPermission: string;
  createPermission?: string;
  filters: RiskField[];
  columns: RiskColumn[];
  tabs: string[];
};
const f = (
  key: string,
  type: RiskField["type"] = "text",
  options?: string[],
): RiskField => ({ key, type, options });
const levels = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];
const dates = (key = "createdRange") => f(key, "dateRange");
export const riskConfigs: Record<string, RiskConfig> = {
  "risk-rules": {
    resource: "rules",
    listPermission: "risk:rule:list",
    createPermission: "risk:rule:create",
    filters: [
      f("code"),
      f("name"),
      f("ruleType", "select", [
        "REGISTRATION",
        "LOGIN",
        "DEVICE",
        "IP",
        "DEPOSIT",
        "WITHDRAWAL",
        "BETTING",
        "BONUS",
        "PROVIDER_CALLBACK",
        "MANUAL",
      ]),
      f("riskLevel", "select", levels),
      f("enabled", "select", ["true", "false"]),
      f("business"),
      dates("updatedRange"),
    ],
    columns: [
      { key: "code", width: 150 },
      { key: "nameZh", width: 160 },
      { key: "nameEn", width: 160 },
      { key: "ruleType", width: 140, status: true },
      { key: "riskLevel", width: 100, status: true },
      { key: "priority", width: 85 },
      { key: "enabled", width: 90, status: true },
      { key: "triggerCount", width: 105 },
      { key: "lastTriggeredTime", width: 170, time: true },
      { key: "updatedAt", width: 170, time: true },
    ],
    tabs: [
      "basicInformation",
      "conditions",
      "actions",
      "scope",
      "testResults",
      "triggerRecords",
      "operationLogs",
    ],
  },
  "risk-events": {
    resource: "events",
    listPermission: "risk:event:list",
    filters: [
      f("userID"),
      f("username"),
      f("rule"),
      f("riskType"),
      f("riskLevel", "select", levels),
      f("status", "select", [
        "PENDING",
        "UNDER_REVIEW",
        "CONFIRMED",
        "CLEARED",
        "RESTRICTED",
        "BLOCKED",
      ]),
      f("provider"),
      f("ip"),
      dates(),
    ],
    columns: [
      { key: "userNo", width: 140 },
      { key: "username", width: 120 },
      { key: "riskType", width: 150, status: true },
      { key: "riskLevel", width: 100, status: true },
      { key: "rule", width: 130 },
      { key: "triggerValue", width: 160 },
      { key: "ip", width: 120, masked: true },
      { key: "status", width: 120, status: true },
      { key: "assignedReviewer", width: 130 },
      { key: "createdAt", width: 170, time: true },
      { key: "updatedAt", width: 170, time: true },
    ],
    tabs: [
      "riskSummary",
      "triggerData",
      "userInformation",
      "relatedIPs",
      "relatedDevices",
      "relatedAccounts",
      "relatedDeposits",
      "relatedWithdrawals",
      "relatedBets",
      "reviewHistory",
      "internalNotes",
      "operationLogs",
    ],
  },
  "review-queue": {
    resource: "reviews",
    listPermission: "risk:review:list",
    filters: [
      f("businessType", "select", [
        "USER",
        "WITHDRAWAL",
        "DEPOSIT",
        "BET",
        "BONUS",
        "DEVICE",
        "IP",
        "PROVIDER_CALLBACK",
      ]),
      f("userID"),
      f("username"),
      f("riskLevel", "select", levels),
      f("status", "select", [
        "PENDING",
        "ASSIGNED",
        "UNDER_REVIEW",
        "APPROVED",
        "REJECTED",
        "CLEARED",
        "ESCALATED",
      ]),
      f("reviewer"),
      dates(),
    ],
    columns: [
      { key: "businessType", width: 140, status: true },
      { key: "userNo", width: 140 },
      { key: "username", width: 120 },
      { key: "riskLevel", width: 100, status: true },
      { key: "reason", width: 200 },
      { key: "assignedReviewer", width: 130 },
      { key: "status", width: 120, status: true },
      { key: "createdAt", width: 170, time: true },
      { key: "updatedAt", width: 170, time: true },
    ],
    tabs: ["overview", "reviewHistory", "internalNotes", "operationLogs"],
  },
  "user-associations": {
    resource: "associations",
    listPermission: "risk:association:list",
    filters: [
      f("userID"),
      f("username"),
      f("associationType", "select", [
        "DEVICE",
        "IP",
        "WITHDRAWAL_ACCOUNT",
        "DEPOSIT_ACCOUNT",
        "FINGERPRINT",
        "REFERRAL",
        "EMAIL",
        "MOBILE",
      ]),
      f("relatedUserID"),
      f("ip"),
      f("riskLevel", "select", levels),
    ],
    columns: [
      { key: "userNo", width: 140 },
      { key: "username", width: 120 },
      { key: "associationType", width: 170, status: true },
      { key: "associationValueMasked", width: 190, masked: true },
      { key: "relatedAccountCount", width: 130 },
      { key: "riskLevel", width: 100, status: true },
      { key: "firstSeenAt", width: 170, time: true },
      { key: "lastSeenAt", width: 170, time: true },
    ],
    tabs: [
      "overview",
      "relationshipGraph",
      "associatedUsers",
      "devices",
      "ips",
      "associations",
    ],
  },
  blacklist: {
    resource: "blacklists",
    listPermission: "risk:blacklist:list",
    createPermission: "risk:blacklist:create",
    filters: [
      f("type", "select", [
        "USER",
        "DEVICE",
        "IP",
        "EMAIL",
        "MOBILE",
        "BANK_ACCOUNT",
        "WALLET_ADDRESS",
        "PAYMENT_ACCOUNT",
      ]),
      f("value"),
      f("reason"),
      f("status", "select", ["ENABLED", "DISABLED"]),
      f("createdBy"),
      dates(),
      dates("expiresRange"),
    ],
    columns: [
      { key: "type", width: 150, status: true },
      { key: "valueMasked", width: 180, masked: true },
      { key: "reason", width: 220 },
      { key: "status", width: 105, status: true },
      { key: "relatedUserCount", width: 130 },
      { key: "createdBy", width: 120 },
      { key: "createdAt", width: 170, time: true },
      { key: "expiresAt", width: 170, time: true },
    ],
    tabs: [
      "overview",
      "relatedUsers",
      "triggerRecords",
      "statusHistory",
      "operationLogs",
    ],
  },
  "withdrawal-risk": {
    resource: "withdrawals",
    listPermission: "risk:withdrawal:list",
    filters: [
      f("withdrawalOrderID"),
      f("userID"),
      f("username"),
      f("riskLevel", "select", levels),
      f("riskStatus"),
      f("kycStatus"),
      f("country"),
      f("currency"),
      f("amountMin", "number"),
      f("amountMax", "number"),
      dates(),
    ],
    columns: [
      { key: "withdrawal_order_id", width: 180 },
      { key: "user_no", width: 140 },
      { key: "username", width: 120 },
      { key: "amount", width: 120 },
      { key: "currency", width: 85 },
      { key: "kyc_status", width: 120, status: true },
      { key: "risk_level", width: 100, status: true },
      { key: "risk_status", width: 135, status: true },
      { key: "associated_account_count", width: 140 },
      { key: "review_status", width: 130, status: true },
      { key: "created_at", width: 170, time: true },
    ],
    tabs: [
      "withdrawalSummary",
      "userRiskProfile",
      "depositAndWithdrawalHistory",
      "bettingSummary",
      "bonusUsage",
      "devicesAndIPs",
      "associatedAccounts",
      "reviewHistory",
      "operationLogs",
    ],
  },
  "risk-operation-logs": {
    resource: "operation-logs",
    listPermission: "risk:operation-log:view",
    filters: [
      f("operatorName"),
      f("action"),
      f("targetType"),
      f("riskLevel", "select", levels),
      dates(),
    ],
    columns: [
      { key: "operator_username", width: 130 },
      { key: "action", width: 190 },
      { key: "target_type", width: 130 },
      { key: "before_state", width: 220 },
      { key: "after_state", width: 220 },
      { key: "reason", width: 180 },
      { key: "ip_address", width: 130, masked: true },
      { key: "created_at", width: 170, time: true },
    ],
    tabs: ["overview"],
  },
};
