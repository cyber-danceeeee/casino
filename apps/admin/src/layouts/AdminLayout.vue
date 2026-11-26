<script setup lang="ts">
import { computed, onMounted, ref, watch, type Component } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ArrowDown } from "@element-plus/icons-vue";
import * as ElementPlusIcons from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import { dashboardItem, menuGroupsForAdmin } from "@/data/adminMenu";
import { localeOptions, type SupportedLocale } from "@/i18n/messages";
import logoIcon from "@/assets/logo_icon.svg";
import { useI18n } from "vue-i18n";
import {
  selectedTimezoneOffset,
  applyProfileTimezone,
  setTimezoneOffset,
  timezoneOptions,
} from "@/utils/timezone";
import { useAdminAuth } from "@/composables/useAdminAuth";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import BrandDraftStatusBar from "@/components/brand/BrandDraftStatusBar.vue";
import { accountAPI, type AccountProfile } from "@/api/adminAccount";
import { systemAPI, type PlatformBranding } from "@/api/adminSystem";

const route = useRoute();
const router = useRouter();
const { t, locale } = useI18n();
const { profile, logout } = useAdminAuth();
const { actor, navigationReady, hasNavigation, navigationIcon, loadPermissions } = useAdminPermissions();
const iconComponents = ElementPlusIcons as unknown as Record<string, Component>;
const accountPreferences = ref<AccountProfile>();
const platformBranding = ref<PlatformBranding>({ platformName: "Casino Admin", platformLogoUrl: "", platformContact: "" });
const platformName = computed(() => platformBranding.value.platformName || t("app.brand"));
const platformLogo = computed(() => platformBranding.value.platformLogoUrl || logoIcon);
const platformFooter = computed(() => `© ${new Date().getFullYear()} ${platformName.value}${platformBranding.value.platformContact ? ` · ${platformBranding.value.platformContact}` : ""}`);

function applyLocale(value: SupportedLocale) {
  locale.value = value;
  window.localStorage.setItem("casino-admin-locale", value);
}

async function loadAccountPreferences() {
  const result = await accountAPI.profile();
  accountPreferences.value = result.profile;
  applyProfileTimezone(result.profile.timezone);
  if (result.profile.preferredLanguage === "zh-CN" || result.profile.preferredLanguage === "en-US") {
    applyLocale(result.profile.preferredLanguage);
  }
}

async function savePreferredLanguage(value: SupportedLocale) {
  try {
    const profile = accountPreferences.value || (await accountAPI.profile()).profile;
    await accountAPI.updateProfile({
      displayName: profile.displayName,
      email: profile.email,
      mobile: profile.mobile,
      preferredLanguage: value,
      timezone: profile.timezone,
    });
    accountPreferences.value = { ...profile, preferredLanguage: value };
  } catch {
    ElMessage.error(t("accountCenter.messages.saveFailed"));
  }
}

function configuredIcon(path: string, code: string, fallback: Component) {
  return iconComponents[navigationIcon(path, code)] || fallback;
}

const visibleMenuGroups = computed(() => {
  if (!navigationReady.value) return [];
  return menuGroupsForAdmin(actor.value?.platform !== false)
    .map((group) => ({
      ...group,
      children: group.children.filter((item) => hasNavigation(item.navigationPath || item.path)),
    }))
    .filter((group) => group.children.length > 0);
});

const administratorName = computed(
  () => profile.value?.displayName || actor.value?.username || t("app.admin"),
);
const administratorInitial = computed(() =>
  administratorName.value.slice(0, 1).toUpperCase(),
);

const activeMenu = computed(() => {
  if (route.path === "/") return dashboardItem.path;
  return route.path;
});

const defaultOpeneds = computed(() =>
  visibleMenuGroups.value
    .filter((group) =>
      group.children.some((item) =>
        route.path.startsWith(item.path.split("/").slice(0, 2).join("/")),
      ),
    )
    .map((group) => group.label),
);

const currentLocale = computed({
  get: () => locale.value as SupportedLocale,
  set: (value: SupportedLocale) => {
    applyLocale(value);
    void savePreferredLanguage(value);
  },
});

const currentTimezone = computed({
  get: () => selectedTimezoneOffset.value,
  set: (value: number) => setTimezoneOffset(value),
});

watch(
  [() => route.meta.i18nKey, () => locale.value, platformName],
  () => {
    const titleKey = String(route.meta.i18nKey || "");
    document.title = `${titleKey ? t(titleKey) : platformName.value} - ${platformName.value}`;
  },
  { immediate: true },
);

onMounted(async () => {
  await Promise.all([
    loadPermissions(),
    loadAccountPreferences().catch(() => undefined),
    systemAPI.platformBranding().then((value) => { platformBranding.value = value; }).catch(() => undefined),
  ]);
});

async function handleUserCommand(command: string) {
  if (command === "profile") return router.push("/account/profile");
  if (command === "security") return router.push("/account/security");
  if (command === "sessions") return router.push("/account/sessions");
  if (command === "login-records") return router.push("/account/login-records");
  if (command !== "logout") return;
  const loginPath = "/login";
  await logout();
  ElMessage.success(t("auth.logoutSuccess"));
  await router.replace(loginPath);
}
</script>

<template>
  <el-container class="admin-layout" direction="vertical">
    <el-header class="admin-header" height="56px">
      <div class="admin-brand">
        <img :src="platformLogo" :alt="platformName" class="admin-header-logo" />
        <span class="admin-brand-title">{{ platformName }}</span>
        <el-tag
          class="admin-brand-tag"
          type="success"
          effect="dark"
          size="small"
        >
          {{ t("singleSite.console") }}
        </el-tag>
      </div>

      <div class="admin-header-right">
        <el-select
          v-model="currentTimezone"
          class="admin-timezone-select"
          :aria-label="t('app.timezone')"
          size="default"
        >
          <el-option
            v-for="option in timezoneOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>
        <el-select
          v-model="currentLocale"
          class="admin-language-select"
          :aria-label="t('app.language')"
          size="default"
        >
          <el-option
            v-for="option in localeOptions"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>

        <el-dropdown trigger="click" @command="handleUserCommand">
          <div class="admin-user-block">
            <el-avatar :size="32" class="admin-avatar">{{
              administratorInitial
            }}</el-avatar>
            <span class="admin-user-name">{{ administratorName }}</span>
            <el-icon><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item disabled>{{
                profile?.username || t("app.superAdmin")
              }}</el-dropdown-item>
              <el-dropdown-item command="profile">{{ t("accountCenter.profile.title") }}</el-dropdown-item>
              <el-dropdown-item command="security">{{ t("accountCenter.security.title") }}</el-dropdown-item>
              <el-dropdown-item command="sessions">{{ t("accountCenter.sessions.title") }}</el-dropdown-item>
              <el-dropdown-item command="login-records">{{ t("accountCenter.loginRecords.title") }}</el-dropdown-item>
              <el-dropdown-item command="logout" divided>{{
                t("app.logout")
              }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <el-container class="admin-body">
      <el-aside class="admin-aside" width="220px">
        <el-scrollbar class="admin-menu-wrap">
          <el-menu
            class="admin-menu"
            :default-active="activeMenu"
            :default-openeds="defaultOpeneds"
            router
          >
            <el-menu-item
              v-if="navigationReady && hasNavigation(dashboardItem.path)"
              :index="dashboardItem.path"
            >
              <el-icon><component :is="configuredIcon(dashboardItem.path, dashboardItem.key, dashboardItem.icon)" /></el-icon>
              <span>{{ t(dashboardItem.i18nKey) }}</span>
            </el-menu-item>

            <el-sub-menu
              v-for="group in visibleMenuGroups"
              :key="group.key"
              :index="group.label"
            >
              <template #title>
                <el-icon><component :is="configuredIcon('', group.key, group.icon)" /></el-icon>
                <span>{{ t(group.i18nKey) }}</span>
              </template>
              <el-menu-item
                v-for="item in group.children"
                :key="item.key"
                :index="item.path"
              >
                <el-icon><component :is="configuredIcon(item.path, item.key, item.icon)" /></el-icon>
                <span>{{ t(item.i18nKey) }}</span>
              </el-menu-item>
            </el-sub-menu>
          </el-menu>
        </el-scrollbar>
      </el-aside>

      <el-main class="admin-main">
        <div class="admin-main-body">
          <BrandDraftStatusBar />
          <router-view />
        </div>
        <footer class="admin-footer">{{ platformFooter }}</footer>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped src="@/styles/layout.css"></style>

<style scoped>
.admin-avatar {
  background: #c0c4cc;
  color: #fff;
  font-size: 14px;
}

.admin-language-select {
  width: 128px;
}

.admin-timezone-select {
  width: 104px;
}
</style>
