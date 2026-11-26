import { apiGet, apiPost } from "@/api/http";
import type { AdminActor } from "@/api/adminUsers";

export type AdminProfile = {
  id: number;
  username: string;
  displayName: string;
  tenantID?: number;
  siteID?: number;
  defaultSiteID?: number;
  adminType: "PLATFORM" | "TENANT" | "SITE";
  mustChangePassword?: boolean;
};

export type AdminAuthStatus =
  "AUTHENTICATED" | "MFA_REQUIRED" | "MFA_SETUP_REQUIRED";

export type AdminAuthResponse = {
  status: AdminAuthStatus;
  token?: string;
  expiresAt: string;
  challengeToken?: string;
  secret?: string;
  otpauthUri?: string;
  administrator: AdminProfile;
  tenant?: { id: number; code: string; name: string };
  site?: { id: number; code: string; name: string };
  recoveryCodes?: string[];
};

export const loginAdmin = (payload: { username: string; password: string; portalType?: "PLATFORM" | "TENANT" | "SITE" }) =>
  apiPost<AdminAuthResponse>("/admin/auth/login", payload);

export const confirmAdminMFASetup = (payload: {
  challengeToken: string;
  code: string;
}) => apiPost<AdminAuthResponse>("/admin/auth/mfa/setup/confirm", payload);

export const verifyAdminMFA = (payload: {
  challengeToken: string;
  code: string;
}) => apiPost<AdminAuthResponse>("/admin/auth/mfa/verify", payload);

export const fetchCurrentAdmin = () => apiGet<AdminActor>("/admin/auth/me");
export const logoutAdmin = () => apiPost<void>("/admin/auth/logout");
