import axios from "axios";
import type { AxiosRequestConfig } from "axios";

export type ApiResponse<T> = {
  code: number;
  msg: string;
  data: T;
};

export class ApiError extends Error {
  readonly errorCode: string;

  constructor(message: string, errorCode = "") {
    super(message);
    this.name = "ApiError";
    this.errorCode = errorCode;
  }
}

export const http = axios.create({
  baseURL: "/api/v1",
  timeout: 15000,
});

http.interceptors.request.use((config) => {
  const token = window.localStorage.getItem("casino-admin-token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

function unwrap<T>(response: ApiResponse<T>) {
  if (response.code !== 0) {
    if (response.code === 401) {
      const cachedProfile = window.localStorage.getItem("casino-admin-profile");
      window.localStorage.removeItem("casino-admin-token");
      window.localStorage.removeItem("casino-admin-profile");
      if (!window.location.pathname.endsWith("/login")) {
        const redirect = `${window.location.pathname}${window.location.search}`;
        let loginPath = "/platform/login";
        try {
          const profile = JSON.parse(cachedProfile || "{}");
          if (profile.adminType === "TENANT") loginPath = "/tenant/login";
        } catch {
          // Fall back to the platform login when cached identity data is invalid.
        }
        window.location.assign(
          `${loginPath}?redirect=${encodeURIComponent(redirect)}`,
        );
      }
    }
    const errorCode =
      response.data && typeof response.data === "object" && "errorCode" in response.data
        ? String((response.data as { errorCode?: unknown }).errorCode || "")
        : "";
    throw new ApiError(response.msg || "Request failed", errorCode);
  }
  return response.data;
}

export async function apiGet<T>(url: string, params?: Record<string, unknown>) {
  const { data } = await http.get<ApiResponse<T>>(url, { params });
  return unwrap(data);
}

export async function apiPut<T>(url: string, payload?: unknown) {
  const { data } = await http.put<ApiResponse<T>>(url, payload);
  return unwrap(data);
}

export async function apiPost<T>(url: string, payload?: unknown, config?: AxiosRequestConfig) {
  const { data } = await http.post<ApiResponse<T>>(url, payload, config);
  return unwrap(data);
}

export async function apiPatch<T>(url: string, payload?: unknown) {
  const { data } = await http.patch<ApiResponse<T>>(url, payload);
  return unwrap(data);
}

export async function apiDelete<T>(url: string, payload?: unknown) {
  const { data } = await http.delete<ApiResponse<T>>(url, { data: payload });
  return unwrap(data);
}
