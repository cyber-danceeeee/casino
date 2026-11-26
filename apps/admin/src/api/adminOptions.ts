import { apiGet } from "./http";

export type AdminOption = {
  label: string;
  value: string;
  labelZh?: string;
  labelEn?: string;
  currencyCode?: string;
  defaultLanguage?: string;
  timezone?: string;
};

export type AdminOptionSource =
  | "providers"
  | "sites"
  | "games"
  | "gameCategories"
  | "gameTags"
  | "gameCollections"
  | "countries"
  | "currencies"
  | "phoneCountryCodes"
  | "languages"
  | "timezones"
  | "gameTypes"
  | "providerCategories"
  | "devices"
  | "registrationChannels"
  | "vipLevels"
  | "agentLevels"
  | "commissionPlans"
  | "paymentMethods"
  | "paymentChannels"
  | "withdrawAccountTypes"
  | "booleanValues"
  | "providerEnvironments"
  | "providerStatuses"
  | "launchTypes"
  | "launchStatuses";

export type AdminOptions = Partial<Record<AdminOptionSource, AdminOption[]>>;

export const fetchAdminOptions = () => apiGet<AdminOptions>("/admin/options");

export function optionList(
  options: AdminOptions,
  source?: AdminOptionSource,
): AdminOption[] {
  return source ? options[source] || [] : [];
}

export function stringOptions(values?: string[]): AdminOption[] {
  return (values || []).map((value) => ({ label: value, value }));
}
