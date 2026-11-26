export type BusinessRecord = Record<string, unknown>;

const meaningfulIdentifierKeys = new Set([
  "accountNo",
  "adminUsername",
  "agentCode",
  "assetName",
  "backupNo",
  "betNo",
  "businessOrderNo",
  "callbackNo",
  "campaignCode",
  "caseNo",
  "channelCode",
  "componentCode",
  "couponCode",
  "externalOrderNo",
  "gameCode",
  "launchNo",
  "merchantOrderNo",
  "orderNo",
  "playerNo",
  "providerCode",
  "providerGameID",
  "providerRoundID",
  "providerTransactionID",
  "releaseVersion",
  "siteCode",
  "taskNo",
  "tenantCode",
  "transactionNo",
  "txnNo",
  "upstreamOrderNo",
  "userNo",
  "externalTransactionID",
  "providerGameId",
  "providerRoundID",
  "providerTransactionID",
]);

export function isInternalIdentifierKey(key: string) {
  if (meaningfulIdentifierKeys.has(key)) return false;
  const compact = key.replace(/[_\s-]/g, "").toLowerCase();
  if (compact.includes("uuid")) return true;
  return (
    compact === "id" ||
    compact.endsWith("id") ||
    [
      "tenantid",
      "siteid",
      "userid",
      "playerid",
      "walletid",
      "accountid",
      "recordid",
      "targetid",
      "roleid",
      "permissionid",
      "createdby",
      "updatedby",
    ].includes(compact)
  );
}

export function businessEntries(value: unknown): Array<[string, unknown]> {
  if (!value || typeof value !== "object" || Array.isArray(value)) return [];
  return Object.entries(value as BusinessRecord).filter(
    ([key]) =>
      !isInternalIdentifierKey(key) &&
      !["deletedAt", "rawMetadata", "rawPayload"].includes(key),
  );
}

export function businessColumns(value: unknown, limit = 12) {
  const first = Array.isArray(value)
    ? value.find((item) => item && typeof item === "object")
    : undefined;
  return businessEntries(first)
    .map(([key]) => key)
    .slice(0, limit);
}

export function sanitizeBusinessValue(value: unknown): unknown {
  if (Array.isArray(value)) return value.map(sanitizeBusinessValue);
  if (value && typeof value === "object") {
    return Object.fromEntries(
      businessEntries(value).map(([key, item]) => [
        key,
        sanitizeBusinessValue(item),
      ]),
    );
  }
  return value;
}

export function statusKey(value: unknown) {
  return String(value ?? "UNKNOWN")
    .trim()
    .toUpperCase()
    .replace(/[\s-]+/g, "_");
}

export function statusType(
  value: unknown,
): "success" | "warning" | "danger" | "info" | "primary" {
  const key = statusKey(value);
  if (
    [
      "ACTIVE",
      "ENABLED",
      "HEALTHY",
      "ONLINE",
      "PUBLISHED",
      "SUCCESS",
      "SUCCESSFUL",
      "COMPLETED",
      "APPROVED",
      "RESOLVED",
    ].includes(key)
  )
    return "success";
  if (
    [
      "FAILED",
      "ERROR",
      "BLOCKED",
      "REJECTED",
      "UNAVAILABLE",
      "UNHEALTHY",
      "EXPIRED",
      "DEAD",
    ].includes(key)
  )
    return "danger";
  if (
    [
      "PENDING",
      "PROCESSING",
      "REVIEW",
      "UNDER_REVIEW",
      "DRAFT",
      "SCHEDULED",
      "DEGRADED",
      "MAINTENANCE",
      "WARNING",
    ].includes(key)
  )
    return "warning";
  return "info";
}

export function humanizeBusinessKey(key: string) {
  return key
    .replace(/_/g, " ")
    .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}
