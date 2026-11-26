import type { FundQuery, FundDetail } from "@/api/adminFunds";
import {
  fetchAdjustments,
  fetchDeposits,
  fetchReconciliations,
  fetchTransactions,
  fetchWallets,
  fetchWithdrawals,
  type AdjustmentRow,
  type DepositRow,
  type ReconciliationRow,
  type TransactionRow,
  type WalletRow,
  type WithdrawalRow,
} from "@/api/adminFunds";
import type { PageResult } from "@/api/adminUsers";

const transactionTypeValues = [
  "DEPOSIT",
  "WITHDRAWAL",
  "BET_DEBIT",
  "BET_CREDIT",
  "BET_REFUND",
  "ROLLBACK",
  "BONUS_CREDIT",
  "BONUS_UNLOCK",
  "BONUS_TRANSFER",
  "CASH_REWARD",
  "MANUAL_CREDIT",
  "MANUAL_DEBIT",
  "FREEZE",
  "UNFREEZE",
  "COMMISSION_CREDIT",
  "SETTLEMENT_CREDIT",
  "REVERSAL",
];

export type FundRecord = Record<string, unknown>;
export type FilterOption = { label: string; value: string };
export type FilterField = {
  key: string;
  label: string;
  type?: "select" | "dateRange";
  options?: FilterOption[];
};
export type TableColumn = {
  key: string;
  label: string;
  width?: number;
  money?: boolean;
  time?: boolean;
  status?: boolean;
  dangerTag?: boolean;
};
export type FundPageConfig = {
  key: string;
  searchTitle: string;
  tableTitle: string;
  filters: FilterField[];
  columns: TableColumn[];
  detailTabs: Array<{ key: string; title: string }>;
  loader: (query: FundQuery) => Promise<PageResult<FundRecord>>;
};

function records<T extends object>(
  promise: Promise<PageResult<T>>,
): Promise<PageResult<FundRecord>> {
  return promise.then((page) => ({
    ...page,
    list: page.list.map((row) => Object.fromEntries(Object.entries(row))),
  }));
}

const statuses = (...values: string[]) =>
  values.map((value) => ({ label: value, value }));
const baseUserFilters: FilterField[] = [
  { key: "phone", label: "Mobile Number" },
  { key: "userNo", label: "Player Number" },
  { key: "currency", label: "Currency" },
];
const rangeFilters: FilterField[] = [
  { key: "amountMin", label: "Min Amount" },
  { key: "amountMax", label: "Max Amount" },
  { key: "createdRange", label: "Created Time", type: "dateRange" },
];

export const fundPageConfigs: Record<string, FundPageConfig> = {
  "wallet-accounts": {
    key: "wallet-accounts",
    searchTitle: "Wallet Account Search",
    tableTitle: "Wallet Accounts",
    detailTabs: [
      { key: "wallet", title: "Overview" },
      { key: "balanceComposition", title: "Balance Composition" },
      { key: "recentTransactions", title: "Recent Transactions" },
      { key: "freezes", title: "Frozen Funds" },
      { key: "relatedUser", title: "Related User" },
      { key: "operationLogs", title: "Operation Logs" },
    ],
    loader: (query) => records<WalletRow>(fetchWallets(query)),
    filters: [
      ...baseUserFilters,
      {
        key: "status",
        label: "Wallet Status",
        type: "select",
        options: statuses("ACTIVE", "FROZEN"),
      },
      { key: "country", label: "Country" },
      { key: "vipLevel", label: "VIP Level" },
      { key: "riskStatus", label: "Risk Status" },
      { key: "availableBalanceMin", label: "Min Available" },
      { key: "availableBalanceMax", label: "Max Available" },
      { key: "frozenBalanceMin", label: "Min Frozen" },
      { key: "frozenBalanceMax", label: "Max Frozen" },
      { key: "createdRange", label: "Created Time", type: "dateRange" },
      { key: "updatedRange", label: "Updated Time", type: "dateRange" },
    ],
    columns: [
      { key: "userNo", label: "Player Number", width: 150 },
      { key: "phone", label: "Mobile Number", width: 145 },
      { key: "currency", label: "Currency", width: 90 },
      {
        key: "availableBalance",
        label: "Available Balance",
        width: 145,
        money: true,
      },
      {
        key: "frozenBalance",
        label: "Frozen Balance",
        width: 135,
        money: true,
        dangerTag: true,
      },
      { key: "bonusBalance", label: "Bonus Balance", width: 130, money: true },
      { key: "status", label: "Wallet Status", width: 120, status: true },
      { key: "updatedAt", label: "Updated Time", width: 175, time: true },
    ],
  },
  "wallet-ledgers": {
    key: "wallet-ledgers",
    searchTitle: "Wallet Transaction Search",
    tableTitle: "Wallet Transactions",
    detailTabs: [
      { key: "transaction", title: "Transaction Overview" },
      { key: "balanceChange", title: "Balance Change" },
      { key: "relatedBusiness", title: "Related Business" },
      { key: "rawMetadata", title: "Raw Metadata" },
      { key: "operationLogs", title: "Operation Logs" },
    ],
    loader: (query) => records<TransactionRow>(fetchTransactions(query)),
    filters: [
      { key: "transactionID", label: "Transaction Number" },
      { key: "externalTransactionID", label: "Upstream Transaction Number" },
      ...baseUserFilters,
      { key: "transactionType", label: "Transaction Type", type: "select", options: statuses(...transactionTypeValues) },
      {
        key: "direction",
        label: "Direction",
        type: "select",
        options: statuses("IN", "OUT"),
      },
      { key: "referenceType", label: "Reference Type" },
      { key: "referenceNo", label: "Reference No." },
      { key: "providerCode", label: "Provider" },
      { key: "status", label: "Status" },
      ...rangeFilters,
      { key: "completedRange", label: "Completed Time", type: "dateRange" },
    ],
    columns: [
      { key: "transactionID", label: "Transaction Number", width: 190 },
      { key: "userNo", label: "Player Number", width: 145 },
      { key: "transactionType", label: "Type", width: 160 },
      { key: "amount", label: "Amount", width: 120, money: true },
      { key: "currency", label: "Currency", width: 90 },
      {
        key: "balanceBefore",
        label: "Balance Before",
        width: 135,
        money: true,
      },
      { key: "balanceAfter", label: "Balance After", width: 135, money: true },
      { key: "referenceNo", label: "Business Order Number", width: 190 },
      { key: "status", label: "Status", width: 110, status: true },
      { key: "createdAt", label: "Created Time", width: 175, time: true },
    ],
  },
  "manual-adjustments": {
    key: "manual-adjustments",
    searchTitle: "Manual Adjustment Search",
    tableTitle: "Manual Adjustments",
    detailTabs: [
      { key: "adjustment", title: "Application" },
      { key: "approvals", title: "Approval History" },
      { key: "walletResult", title: "Wallet Result" },
      { key: "operationLogs", title: "Audit Logs" },
    ],
    loader: (query) => records<AdjustmentRow>(fetchAdjustments(query)),
    filters: [
      { key: "adjustmentID", label: "Adjustment Number" },
      ...baseUserFilters,
      {
        key: "adjustmentType",
        label: "Type",
        type: "select",
        options: statuses("CREDIT", "DEBIT"),
      },
      {
        key: "reasonCode",
        label: "Reason Code",
        type: "select",
        options: statuses(
          "BONUS_COMPENSATION",
          "CUSTOMER_SERVICE_COMPENSATION",
          "PROMOTION_ADJUSTMENT",
          "CORRECTION",
          "CHARGEBACK",
          "RISK_DEDUCTION",
          "OTHER",
        ),
      },
      {
        key: "status",
        label: "Status",
        type: "select",
        options: statuses(
          "PENDING_APPROVAL",
          "COMPLETED",
          "REJECTED",
          "CANCELLED",
          "FAILED",
        ),
      },
      { key: "applicant", label: "Applicant" },
      { key: "approver", label: "Approver" },
      ...rangeFilters,
      { key: "completedRange", label: "Completed Time", type: "dateRange" },
    ],
    columns: [
      { key: "adjustmentID", label: "Adjustment Number", width: 190 },
      { key: "userNo", label: "Player Number", width: 145 },
      { key: "username", label: "Mobile Number", width: 145 },
      { key: "adjustmentType", label: "Type", width: 100, status: true },
      { key: "amount", label: "Amount", width: 120, money: true },
      { key: "currency", label: "Currency", width: 90 },
      { key: "reasonCode", label: "Reason Code", width: 140 },
      { key: "status", label: "Status", width: 150, status: true },
      { key: "applicant", label: "Applicant", width: 130 },
      { key: "approver", label: "Approver", width: 130 },
      { key: "createdAt", label: "Created Time", width: 175, time: true },
      { key: "completedAt", label: "Completed Time", width: 175, time: true },
    ],
  },
  "deposit-orders": {
    key: "deposit-orders",
    searchTitle: "Deposit Order Search",
    tableTitle: "Deposit Orders",
    detailTabs: [
      { key: "order", title: "Order Overview" },
      { key: "paymentInformation", title: "Payment Information" },
      { key: "walletCredit", title: "Wallet Credit" },
      { key: "callbacks", title: "Callback Logs" },
      { key: "operationLogs", title: "Operation Logs" },
    ],
    loader: (query) => records<DepositRow>(fetchDeposits(query)),
    filters: [
      { key: "platformOrderID", label: "Deposit Order Number" },
      { key: "channelOrderID", label: "Upstream Order Number" },
      ...baseUserFilters,
      { key: "paymentChannel", label: "Payment Channel" },
      { key: "paymentMethod", label: "Payment Method" },
      {
        key: "status",
        label: "Status",
        type: "select",
        options: statuses(
          "PENDING",
          "MANUAL_REVIEW",
          "SUCCESS",
          "FAILED",
          "CANCELLED",
          "EXPIRED",
        ),
      },
      ...rangeFilters,
      { key: "paidRange", label: "Paid Time", type: "dateRange" },
      { key: "callbackRange", label: "Callback Time", type: "dateRange" },
    ],
    columns: [
      { key: "platformOrderID", label: "Deposit Order Number", width: 200 },
      { key: "channelOrderID", label: "Upstream Order Number", width: 190 },
      { key: "userNo", label: "Player Number", width: 145 },
      { key: "paymentChannel", label: "Channel", width: 130 },
      { key: "paymentMethod", label: "Method", width: 120 },
      { key: "requestedAmount", label: "Requested", width: 120, money: true },
      { key: "fee", label: "Fee", width: 100, money: true },
      {
        key: "actualReceivedAmount",
        label: "Received",
        width: 120,
        money: true,
      },
      { key: "currency", label: "Currency", width: 90 },
      { key: "status", label: "Status", width: 130, status: true },
      { key: "createdAt", label: "Created Time", width: 175, time: true },
      { key: "paidAt", label: "Paid Time", width: 175, time: true },
      { key: "reviewer", label: "Reviewer", width: 125 },
    ],
  },
  "withdrawal-orders": {
    key: "withdrawal-orders",
    searchTitle: "Withdrawal Order Search",
    tableTitle: "Withdrawal Orders",
    detailTabs: [
      { key: "order", title: "Order Overview" },
      { key: "userReview", title: "User Review" },
      { key: "fundsAndFreeze", title: "Funds and Freeze" },
      { key: "paymentInformation", title: "Payment Information" },
      { key: "approvals", title: "Approval History" },
      { key: "payouts", title: "Payout Logs" },
      { key: "operationLogs", title: "Operation Logs" },
    ],
    loader: (query) => records<WithdrawalRow>(fetchWithdrawals(query)),
    filters: [
      { key: "platformOrderID", label: "Withdrawal Order Number" },
      { key: "channelOrderID", label: "Upstream Order Number" },
      ...baseUserFilters,
      { key: "paymentMethod", label: "Payment Method" },
      {
        key: "status",
        label: "Status",
        type: "select",
        options: statuses(
          "PENDING",
          "RISK_REVIEW",
          "FINANCE_REVIEW",
          "PROCESSING",
          "SUCCESS",
          "FAILED",
          "REJECTED",
          "CANCELLED",
        ),
      },
      { key: "riskStatus", label: "Risk Status" },
      { key: "kycStatus", label: "KYC Status" },
      { key: "reviewer", label: "Reviewer" },
      ...rangeFilters,
      { key: "reviewedRange", label: "Reviewed Time", type: "dateRange" },
      { key: "completedRange", label: "Completed Time", type: "dateRange" },
    ],
    columns: [
      { key: "platformOrderID", label: "Withdrawal Order Number", width: 200 },
      { key: "channelOrderID", label: "Upstream Order Number", width: 190 },
      { key: "userNo", label: "Player Number", width: 145 },
      { key: "requestedAmount", label: "Requested", width: 120, money: true },
      { key: "fee", label: "Fee", width: 100, money: true },
      { key: "actualPayoutAmount", label: "Payout", width: 120, money: true },
      { key: "currency", label: "Currency", width: 90 },
      { key: "paymentMethod", label: "Method", width: 120 },
      { key: "recipientAccountMasked", label: "Recipient Account", width: 155 },
      { key: "riskStatus", label: "Risk Status", width: 120, status: true },
      { key: "kycStatus", label: "KYC Status", width: 120, status: true },
      { key: "status", label: "Status", width: 135, status: true },
      { key: "createdAt", label: "Created Time", width: 175, time: true },
      { key: "reviewer", label: "Reviewer", width: 125 },
    ],
  },
  "financial-reconciliation": {
    key: "financial-reconciliation",
    searchTitle: "Reconciliation Search",
    tableTitle: "Financial Reconciliation",
    detailTabs: [
      { key: "record", title: "Summary" },
      { key: "matchedRecords", title: "Matched Records" },
      { key: "differences", title: "Difference Records" },
      { key: "logs", title: "Processing Logs" },
      { key: "resolutionHistory", title: "Resolution History" },
      { key: "operationLogs", title: "Operation Logs" },
    ],
    loader: (query) => records<ReconciliationRow>(fetchReconciliations(query)),
    filters: [
      { key: "reconciliationID", label: "Reconciliation Number" },
      { key: "businessDate", label: "Business Date" },
      {
        key: "businessType",
        label: "Business Type",
        type: "select",
        options: statuses("DEPOSIT", "WITHDRAWAL", "ADJUSTMENT"),
      },
      { key: "channel", label: "Channel" },
      { key: "currency", label: "Currency" },
      {
        key: "status",
        label: "Status",
        type: "select",
        options: statuses("MATCHED", "DIFFERENCE", "RESOLVED"),
      },
      { key: "differenceType", label: "Difference Type" },
      { key: "platformOrderID", label: "Business Order Number" },
      { key: "channelOrderID", label: "Upstream Order Number" },
      { key: "generatedRange", label: "Generated Time", type: "dateRange" },
    ],
    columns: [
      { key: "reconciliationID", label: "Reconciliation Number", width: 200 },
      { key: "businessDate", label: "Business Date", width: 120 },
      { key: "businessType", label: "Business Type", width: 130 },
      { key: "channel", label: "Channel", width: 120 },
      { key: "currency", label: "Currency", width: 90 },
      { key: "platformOrderCount", label: "Platform Orders", width: 130 },
      {
        key: "platformAmount",
        label: "Platform Amount",
        width: 140,
        money: true,
      },
      { key: "channelOrderCount", label: "Ledger Entries", width: 120 },
      { key: "channelAmount", label: "Ledger Amount", width: 130, money: true },
      { key: "differenceAmount", label: "Difference", width: 120, money: true },
      { key: "differenceCount", label: "Diff Count", width: 105 },
      { key: "status", label: "Status", width: 115, status: true },
      { key: "generatedBy", label: "Generated By", width: 135 },
      { key: "generatedAt", label: "Generated Time", width: 175, time: true },
    ],
  },
};

export type DetailLoader = (id: number) => Promise<FundDetail>;
