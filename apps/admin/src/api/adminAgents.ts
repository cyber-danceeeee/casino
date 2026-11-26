import { apiDelete, apiGet, apiPatch, apiPost } from "./http";
import type { PageResult } from "./adminUsers";

export type AgentQuery = Record<string, string | number | undefined>;
export type AgentRecord = Record<string, unknown> & { id: number };
export type AgentDetail = Record<string, unknown>;
export const fetchAgents = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/agents", p);
export const fetchAgent = (id: number) =>
  apiGet<AgentDetail>(`/admin/agents/${id}`);
export const createAgent = (p: Record<string, unknown>) =>
  apiPost<{ id: number }>("/admin/agents", p);
export const updateAgent = (id: number, p: Record<string, unknown>) =>
  apiPatch<void>(`/admin/agents/${id}`, p);
export const setAgentStatus = (id: number, enabled: boolean, reason: string) =>
  apiPatch<void>(`/admin/agents/${id}/status`, { enabled, reason });
export const changeAgentParent = (
  id: number,
  parentAgentId: number,
  reason: string,
) =>
  apiPost<void>(`/admin/agents/${id}/change-parent`, { parentAgentId, reason });
export const assignAgentPlan = (
  id: number,
  commissionPlanId: number,
  reason: string,
) =>
  apiPost<void>(`/admin/agents/${id}/assign-plan`, {
    commissionPlanId,
    reason,
  });
export const changeAgentAttribution = (p: Record<string, unknown>) =>
  apiPost<void>("/admin/agent-attributions/change", p);
export const fetchHierarchy = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/agent-hierarchy", p);
export const fetchChannels = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/agent-channels", p);
export const fetchChannel = (id: number) =>
  apiGet<AgentDetail>(`/admin/agent-channels/${id}`);
export const createChannel = (p: Record<string, unknown>) =>
  apiPost<{ id: number }>("/admin/agent-channels", p);
export const updateChannel = (id: number, p: Record<string, unknown>) =>
  apiPatch<void>(`/admin/agent-channels/${id}`, p);
export const setChannelStatus = (
  id: number,
  enabled: boolean,
  reason: string,
) => apiPatch<void>(`/admin/agent-channels/${id}/status`, { enabled, reason });
export const fetchPlans = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/commission-plans", p);
export const fetchPlan = (id: number) =>
  apiGet<AgentDetail>(`/admin/commission-plans/${id}`);
export const createPlan = (p: Record<string, unknown>) =>
  apiPost<{ id: number }>("/admin/commission-plans", p);
export const updatePlan = (id: number, p: Record<string, unknown>) =>
  apiPatch<void>(`/admin/commission-plans/${id}`, p);
export const setPlanStatus = (id: number, enabled: boolean, reason: string) =>
  apiPatch<void>(`/admin/commission-plans/${id}/status`, { enabled, reason });
export const deletePlan = (id: number, reason: string) =>
  apiDelete<void>(`/admin/commission-plans/${id}`, { reason });
export const fetchCommissions = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/commission-records", p);
export const fetchCommission = (id: number) =>
  apiGet<AgentDetail>(`/admin/commission-records/${id}`);
export const fetchSettlements = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/agent-settlements", p);
export const fetchSettlement = (id: number) =>
  apiGet<AgentDetail>(`/admin/agent-settlements/${id}`);
export const fetchSettlementStatement = (id: number) =>
  apiGet<AgentDetail>(`/admin/agent-settlements/${id}/export`);
export const generateSettlement = (p: Record<string, unknown>) =>
  apiPost<{ id: number }>("/admin/agent-settlements/generate", p);
export const settlementAction = (id: number, action: string, reason: string) =>
  apiPost<void>(`/admin/agent-settlements/${id}/${action}`, { reason });
export const fetchAgentWallets = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/agent-wallets", p);
export const fetchAgentWallet = (id: number) =>
  apiGet<AgentDetail>(`/admin/agent-wallets/${id}`);
export const setWalletFrozen = (id: number, frozen: boolean, reason: string) =>
  apiPost<void>(
    `/admin/agent-wallets/${id}/${frozen ? "freeze" : "unfreeze"}`,
    { reason },
  );
export const adjustAgentWallet = (id: number, p: Record<string, unknown>) =>
  apiPost<void>(`/admin/agent-wallets/${id}/adjustments`, p);
export const fetchAgentLogs = (p: AgentQuery) =>
  apiGet<PageResult<AgentRecord>>("/admin/agent-operation-logs", p);
export const fetchAgentLog = (id: number) =>
  apiGet<AgentDetail>(`/admin/agent-operation-logs/${id}`);
