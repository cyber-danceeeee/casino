import {
  createRouter,
  createWebHistory,
  type RouteRecordRaw,
} from "vue-router";
import AdminLayout from "@/layouts/AdminLayout.vue";
import { allMenuItems, dashboardItem, userCenterKeys } from "@/data/adminMenu";
import { moduleConfigs } from "@/data/moduleCatalog";
import i18n from "@/i18n";
import { useAdminPermissions } from "@/composables/useAdminPermissions";

const modulePageComponent = () => import("@/views/ModulePage.vue");

const userCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "user-list": () => import("@/views/users/UserListPage.vue"),
  "login-records": () => import("@/views/users/LoginRecordsPage.vue"),
  "device-management": () => import("@/views/users/DeviceManagementPage.vue"),
  "risk-users": () => import("@/views/users/RiskUsersPage.vue"),
	"player-tags": () => import("@/views/users/UserListPage.vue"),
};

const fundCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "wallet-accounts": () => import("@/views/funds/FundsPage.vue"),
  "wallet-ledgers": () => import("@/views/funds/FundsPage.vue"),
  "manual-adjustments": () => import("@/views/funds/FundsPage.vue"),
  "deposit-orders": () => import("@/views/funds/FundsPage.vue"),
  "withdrawal-orders": () => import("@/views/funds/FundsPage.vue"),
  "financial-reconciliation": () => import("@/views/funds/FundsPage.vue"),
  "deposit-methods": () => import("@/views/funds/DepositConfigPage.vue"),
  "deposit-channels": () => import("@/views/funds/DepositConfigPage.vue"),
  "deposit-amount-options": () => import("@/views/funds/DepositConfigPage.vue"),
};

const betCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "bet-records": () => import("@/views/betting/BettingPage.vue"),
  "rollback-records": () => import("@/views/betting/BettingPage.vue"),
  "abnormal-bets": () => import("@/views/betting/BettingPage.vue"),
  "provider-callback-logs": () => import("@/views/betting/BettingPage.vue"),
  "provider-reconciliation": () => import("@/views/betting/BettingPage.vue"),
};

const gameCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "provider-connections": () => import("@/views/games/ProviderConnectionsPage.vue"),
  "provider-management": () => import("@/views/games/GameCenterPage.vue"),
  "game-list": () => import("@/views/games/GameCenterPage.vue"),
  "game-categories": () => import("@/views/games/GameCenterPage.vue"),
  "game-tags": () => import("@/views/games/GameCenterPage.vue"),
  "game-collections": () => import("@/views/games/GameCenterPage.vue"),
  "game-sync": () => import("@/views/games/GameCenterPage.vue"),
  "game-launch-records": () => import("@/views/games/GameCenterPage.vue"),
	"game-operations": () => import("@/views/games/GameOperationsPage.vue"),
};

const operationCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "asset-library": () => import("@/views/operations/AssetLibraryPage.vue"),
  banners: () => import("@/views/operations/BannerManagementPage.vue"),
  announcements: () => import("@/views/operations/AnnouncementManagementPage.vue"),
  activities: () => import("@/views/operations/OperationsPage.vue"),
  vip: () => import("@/views/operations/OperationsPage.vue"),
  "push-messages": () => import("@/views/operations/OperationsPage.vue"),
  popups: () => import("@/views/operations/PopupManagementPage.vue"),
  "social-links": () => import("@/views/operations/OperationsPage.vue"),
  "h5-navigation": () => import("@/views/operations/H5NavigationPage.vue"),
  vouchers: () => import("@/views/operations/OperationsPage.vue"),
	"marketing-dashboard": () => import("@/views/marketing/MarketingDashboardPage.vue"),
	campaigns: () => import("@/views/marketing/CampaignsPage.vue"),
	missions: () => import("@/views/marketing/CampaignsPage.vue"),
	referral: () => import("@/views/marketing/CampaignsPage.vue"),
};

const agentCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "agent-list": () => import("@/views/agents/AgentCenterPage.vue"),
  "agent-hierarchy": () => import("@/views/agents/AgentCenterPage.vue"),
  "channel-management": () => import("@/views/agents/AgentCenterPage.vue"),
  "commission-plans": () => import("@/views/agents/AgentCenterPage.vue"),
  "commission-records": () => import("@/views/agents/AgentCenterPage.vue"),
  "agent-settlements": () => import("@/views/agents/AgentCenterPage.vue"),
  "agent-wallets": () => import("@/views/agents/AgentCenterPage.vue"),
  "agent-operation-logs": () => import("@/views/agents/AgentCenterPage.vue"),
};

const reportCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "financial-report": () => import("@/views/reports/ReportCenterPage.vue"),
  "game-report": () => import("@/views/reports/ReportCenterPage.vue"),
  "provider-report": () => import("@/views/reports/ReportCenterPage.vue"),
  "user-report": () => import("@/views/reports/ReportCenterPage.vue"),
  "channel-report": () => import("@/views/reports/ReportCenterPage.vue"),
  "agent-report": () => import("@/views/reports/ReportCenterPage.vue"),
  "promotion-report": () => import("@/views/reports/ReportCenterPage.vue"),
  "export-tasks": () => import("@/views/reports/ReportCenterPage.vue"),
};

const systemCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "brand-config": () => import("@/views/brand/BrandSettingsPage.vue"),
	"brand-overview": () => import("@/views/brand/BrandOverviewPage.vue"),
	"brand-theme": () => import("@/views/brand/BrandSettingsPage.vue"),
	"brand-seo": () => import("@/views/brand/BrandSettingsPage.vue"),
	"brand-domain": () => import("@/views/brand/BrandSettingsPage.vue"),
	"brand-region": () => import("@/views/brand/BrandSettingsPage.vue"),
	"brand-languages": () => import("@/views/brand/BrandSettingsPage.vue"),
  sites: () => import("@/views/system/SystemManagementPage.vue"),
  admins: () => import("@/views/system/SystemManagementPage.vue"),
  roles: () => import("@/views/system/SystemManagementPage.vue"),
  permissions: () => import("@/views/system/SystemManagementPage.vue"),
  departments: () => import("@/views/system/SystemManagementPage.vue"),
  menus: () => import("@/views/system/SystemManagementPage.vue"),
  "system-config": () => import("@/views/system/SystemManagementPage.vue"),
  "country-config": () => import("@/views/system/CountryConfigPage.vue"),
  "audit-logs": () => import("@/views/system/SystemManagementPage.vue"),
  "admin-login-logs": () => import("@/views/system/SystemManagementPage.vue"),
  "scheduled-tasks": () => import("@/views/system/SystemManagementPage.vue"),
  "system-monitor": () => import("@/views/system/SystemManagementPage.vue"),
};

const supportCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "support-player-search": () => import("@/views/operations/SupportCenterPage.vue"),
  "support-player-timeline": () => import("@/views/operations/SupportCenterPage.vue"),
  "support-order-search": () => import("@/views/operations/SupportCenterPage.vue"),
  "support-cases": () => import("@/views/operations/SupportCenterPage.vue"),
  "support-exports": () => import("@/views/operations/SupportCenterPage.vue"),
  "support-payment-diagnostics": () => import("@/views/operations/OperationsMonitorPage.vue"),
  "support-game-diagnostics": () => import("@/views/operations/OperationsMonitorPage.vue"),
};

const stabilityCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "payment-monitor": () => import("@/views/operations/OperationsMonitorPage.vue"),
  "provider-monitor": () => import("@/views/operations/OperationsMonitorPage.vue"),
  "maintenance-center": () => import("@/views/operations/MaintenanceCenterPage.vue"),
  "system-health": () => import("@/views/operations/PlatformStabilityPage.vue"),
  "jobs-retries": () => import("@/views/operations/PlatformStabilityPage.vue"),
  "alert-center": () => import("@/views/operations/PlatformStabilityPage.vue"),
  "backup-center": () => import("@/views/operations/PlatformStabilityPage.vue"),
};

const riskCenterComponents: Record<string, RouteRecordRaw["component"]> = {
  "risk-dashboard": () => import("@/views/risk/RiskDashboard.vue"),
  "risk-rules": () => import("@/views/risk/RiskPage.vue"),
  "risk-events": () => import("@/views/risk/RiskPage.vue"),
  "review-queue": () => import("@/views/risk/RiskPage.vue"),
  "user-associations": () => import("@/views/risk/RiskPage.vue"),
  blacklist: () => import("@/views/risk/RiskPage.vue"),
  "withdrawal-risk": () => import("@/views/risk/RiskPage.vue"),
  "risk-operation-logs": () => import("@/views/risk/RiskPage.vue"),
};

const moduleRoutes: RouteRecordRaw[] = allMenuItems
  .filter((item) => item.key !== dashboardItem.key)
  .map((item) => ({
    path: item.path,
    name: item.key,
    component:
      userCenterComponents[item.key] ||
      fundCenterComponents[item.key] ||
      gameCenterComponents[item.key] ||
      betCenterComponents[item.key] ||
      operationCenterComponents[item.key] ||
      riskCenterComponents[item.key] ||
      agentCenterComponents[item.key] ||
      reportCenterComponents[item.key] ||
      systemCenterComponents[item.key] ||
	  supportCenterComponents[item.key] ||
	  stabilityCenterComponents[item.key] ||
      modulePageComponent,
    meta: {
      title: item.label,
      i18nKey: item.i18nKey,
      moduleKey: item.key,
	  navigationPath: item.navigationPath,
    },
  }));

const registeredModuleKeys = new Set(allMenuItems.map((item) => item.key));

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/login", name: "admin-login", component: () => import("@/views/auth/AdminLogin.vue"),
      meta: { title: "Administrator Login", i18nKey: "auth.loginTitle", public: true } },
    { path: "/platform/login", redirect: "/login" },
    { path: "/system/sites", redirect: "/system/brand" },
    {
      path: "/",
      component: AdminLayout,
      children: [
        {
          path: "",
          name: dashboardItem.key,
          component: () => import("@/views/Dashboard.vue"),
          meta: { title: dashboardItem.label, i18nKey: dashboardItem.i18nKey },
        },
        {
          path: "account/profile",
          name: "account-profile",
          component: () => import("@/views/account/AccountCenterPage.vue"),
          meta: { title: "Profile", i18nKey: "accountCenter.profile.title" },
        },
        {
          path: "account/security",
          name: "account-security",
          component: () => import("@/views/account/AccountCenterPage.vue"),
          meta: { title: "Security", i18nKey: "accountCenter.security.title" },
        },
        {
          path: "account/sessions",
          name: "account-sessions",
          component: () => import("@/views/account/AccountCenterPage.vue"),
          meta: { title: "Sessions", i18nKey: "accountCenter.sessions.title" },
        },
        {
          path: "account/login-records",
          name: "account-login-records",
          component: () => import("@/views/account/AccountCenterPage.vue"),
          meta: { title: "Login Records", i18nKey: "accountCenter.loginRecords.title" },
        },

		{
		  path: "players/:playerNo",
		  name: "player-business-detail",
		  component: () => import("@/views/operations/SupportCenterPage.vue"),
		  meta: { title: "Player Detail", i18nKey: "siteExperience.support.playerDetail", moduleKey: "support-player-search", navigationPath: "/support/players" },
		},
        ...moduleRoutes,
      ],
    },
    {
      path: "/:pathMatch(.*)*",
      redirect: "/",
    },
  ],
});

router.beforeEach(async (to) => {
  const token = window.localStorage.getItem("casino-admin-token");
  const isPublic = Boolean(to.meta.public);
  const preferredLogin = "/login";
  if (!isPublic && !token) {
    return { path: preferredLogin, query: { redirect: to.fullPath } };
  }
  if (isPublic && token) {
    return "/";
  }

  const titleKey = String(to.meta.i18nKey || "");
  const title = titleKey
    ? i18n.global.t(titleKey)
    : String(to.meta.title || "Casino Admin");
  document.title = `${title} - Casino Admin`;

  const key = String(to.meta.moduleKey || "");
  if (key && !registeredModuleKeys.has(key) && !moduleConfigs[key] && !userCenterKeys.has(key)) {
    return "/";
  }

  if (key) {
    try {
      const { hasNavigation, loadPermissions } = useAdminPermissions();
      const actor = await loadPermissions();
      if (actor.mustChangePassword && to.path !== "/account/security") {
        return { path: "/account/security", query: { forcePassword: "1" } };
      }
      const navigationPath = String(to.meta.navigationPath || to.path);
      if (!hasNavigation(navigationPath)) return "/";
    } catch {
      return { path: preferredLogin, query: { redirect: to.fullPath } };
    }
  }
  if (!isPublic && !key) {
    try {
      const { loadPermissions } = useAdminPermissions();
      const actor = await loadPermissions();
      if (actor.mustChangePassword && to.path !== "/account/security") {
        return { path: "/account/security", query: { forcePassword: "1" } };
      }
    } catch {
      return { path: preferredLogin, query: { redirect: to.fullPath } };
    }
  }
  if (to.path === "/") {
    try {
      const { firstNavigationPath, hasNavigation, loadPermissions } =
        useAdminPermissions();
      await loadPermissions();
      if (!hasNavigation("/")) return firstNavigationPath() || preferredLogin;
    } catch {
      return { path: preferredLogin, query: { redirect: to.fullPath } };
    }
  }

  return true;
});

export default router;
