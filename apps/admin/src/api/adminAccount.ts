import { apiDelete, apiGet, apiPost, apiPut } from "@/api/http";

export type AccountProfile = {
  id: number;
  username: string;
  displayName: string;
  email: string;
  mobile: string;
  tenantID: number;
  tenantName: string;
  defaultSiteID: number;
  defaultSiteName: string;
  preferredLanguage: "zh-CN" | "en-US";
  timezone: string;
  mfaEnabled: boolean;
  mustChangePassword: boolean;
  lastLoginAt?: string;
  lastPasswordChangedAt?: string;
};

export type AccountProfileResponse = {
  profile: AccountProfile;
  adminType: "PLATFORM" | "TENANT" | "SITE";
  roles: string[];
};

export type AccountSession = {
  id: number;
  device: string;
  operatingSystem: string;
  browser: string;
  ipAddress: string;
  country: string;
  userAgent: string;
  createdAt: string;
  lastActiveAt?: string;
  expiredAt: string;
  status: "ACTIVE" | "REVOKED" | "EXPIRED";
  current: boolean;
};

export type AccountLoginRecord = {
  id: number;
  success: boolean;
  failureReason: string;
  ipAddress: string;
  country: string;
  device: string;
  operatingSystem: string;
  browser: string;
  userAgent: string;
  createdAt: string;
};

export const accountAPI = {
  profile: () => apiGet<AccountProfileResponse>("/admin/account/profile"),
  updateProfile: (payload: Pick<AccountProfile, "displayName" | "email" | "mobile" | "preferredLanguage" | "timezone">) =>
    apiPut<void>("/admin/account/profile", payload),
  changePassword: (payload: { currentPassword: string; newPassword: string; confirmPassword: string; revokeOtherSessions: boolean }) =>
    apiPost<void>("/admin/account/password", payload),
  mfaStatus: () => apiGet<{ enabled: boolean; required: boolean; recoveryCodesRemaining: number }>("/admin/account/mfa"),
  startMFA: (currentPassword: string) => apiPost<{ challengeToken: string; expiresAt: string; secret: string; otpauthUri: string }>("/admin/account/mfa/setup", { currentPassword }),
  confirmMFA: (challengeToken: string, code: string) => apiPost<{ recoveryCodes: string[] }>("/admin/account/mfa/setup/confirm", { challengeToken, code }),
  disableMFA: (currentPassword: string, code: string) => apiPost<void>("/admin/account/mfa/disable", { currentPassword, code }),
  regenerateRecoveryCodes: (currentPassword: string, code: string) => apiPost<{ recoveryCodes: string[] }>("/admin/account/mfa/recovery-codes/regenerate", { currentPassword, code }),
  sessions: () => apiGet<AccountSession[]>("/admin/account/sessions"),
  revokeSession: (id: number) => apiDelete<void>(`/admin/account/sessions/${id}`),
  revokeOtherSessions: () => apiDelete<void>("/admin/account/sessions/others"),
  loginRecords: () => apiGet<AccountLoginRecord[]>("/admin/account/login-records"),
};
