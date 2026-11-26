<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { CopyDocument, Key, Lock, Monitor, Refresh, User } from "@element-plus/icons-vue";
import QrcodeVue from "qrcode.vue";
import { useI18n } from "vue-i18n";
import {
  accountAPI,
  type AccountLoginRecord,
  type AccountProfileResponse,
  type AccountSession,
} from "@/api/adminAccount";
import { useAdminPermissions } from "@/composables/useAdminPermissions";

const route = useRoute();
const router = useRouter();
const { t, locale } = useI18n();
const { loadPermissions } = useAdminPermissions();

const loading = ref(false);
const saving = ref(false);
const account = ref<AccountProfileResponse>();
const sessions = ref<AccountSession[]>([]);
const loginRecords = ref<AccountLoginRecord[]>([]);
const mfaStatus = ref({ enabled: false, required: false, recoveryCodesRemaining: 0 });
const recoveryCodes = ref<string[]>([]);
const recoveryVisible = ref(false);
const mfaVisible = ref(false);
const mfaMode = ref<"setup" | "disable" | "regenerate">("setup");
const mfaChallenge = ref<{ challengeToken: string; secret: string; otpauthUri: string }>();

const profileForm = reactive({
  displayName: "",
  email: "",
  mobile: "",
  preferredLanguage: "zh-CN" as "zh-CN" | "en-US",
  timezone: "Asia/Kuala_Lumpur",
});
const passwordForm = reactive({
  currentPassword: "",
  newPassword: "",
  confirmPassword: "",
  revokeOtherSessions: true,
});
const mfaForm = reactive({ currentPassword: "", code: "" });
const accountTimezoneOptions = [
  "UTC",
  "Asia/Kuala_Lumpur",
  "Asia/Shanghai",
  "Asia/Tokyo",
  "Asia/Karachi",
  "Europe/London",
  "America/Mexico_City",
  "America/New_York",
];

const activeTab = computed(() => {
  if (route.path.endsWith("/security")) return "security";
  if (route.path.endsWith("/sessions")) return "sessions";
  if (route.path.endsWith("/login-records")) return "login-records";
  return "profile";
});
const forcePassword = computed(() => route.query.forcePassword === "1" || account.value?.profile.mustChangePassword);

function formatDate(value?: string) {
  if (!value) return "-";
  return new Intl.DateTimeFormat(locale.value, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}

async function loadProfile() {
  account.value = await accountAPI.profile();
  Object.assign(profileForm, {
    displayName: account.value.profile.displayName,
    email: account.value.profile.email,
    mobile: account.value.profile.mobile,
    preferredLanguage: account.value.profile.preferredLanguage || "zh-CN",
    timezone: account.value.profile.timezone || "Asia/Kuala_Lumpur",
  });
}

async function loadSecurity() {
  mfaStatus.value = await accountAPI.mfaStatus();
}

async function loadSessions() {
  sessions.value = await accountAPI.sessions();
}

async function loadLoginRecords() {
  loginRecords.value = await accountAPI.loginRecords();
}

async function loadPage() {
  loading.value = true;
  try {
    await loadProfile();
    if (activeTab.value === "security") await loadSecurity();
    if (activeTab.value === "sessions") await loadSessions();
    if (activeTab.value === "login-records") await loadLoginRecords();
  } catch {
    ElMessage.error(t("accountCenter.messages.loadFailed"));
  } finally {
    loading.value = false;
  }
}

async function saveProfile() {
  saving.value = true;
  try {
    await accountAPI.updateProfile(profileForm);
    const cached = window.localStorage.getItem("casino-admin-profile");
    if (cached) {
      const value = JSON.parse(cached);
      value.displayName = profileForm.displayName;
      window.localStorage.setItem("casino-admin-profile", JSON.stringify(value));
    }
    locale.value = profileForm.preferredLanguage;
    window.localStorage.setItem("casino-admin-locale", profileForm.preferredLanguage);
    ElMessage.success(t("accountCenter.messages.profileSaved"));
    await loadProfile();
  } catch {
    ElMessage.error(t("accountCenter.messages.saveFailed"));
  } finally {
    saving.value = false;
  }
}

async function changePassword() {
  if (!passwordForm.currentPassword || !passwordForm.newPassword || !passwordForm.confirmPassword) {
    ElMessage.warning(t("accountCenter.messages.passwordRequired"));
    return;
  }
  saving.value = true;
  try {
    await accountAPI.changePassword(passwordForm);
    Object.assign(passwordForm, { currentPassword: "", newPassword: "", confirmPassword: "" });
    const actor = await loadPermissions(true);
    actor.mustChangePassword = false;
    ElMessage.success(t("accountCenter.messages.passwordChanged"));
    await loadProfile();
    if (route.query.forcePassword) await router.replace("/account/security");
  } catch {
    ElMessage.error(t("accountCenter.messages.passwordFailed"));
  } finally {
    saving.value = false;
  }
}

function openMFA(mode: "setup" | "disable" | "regenerate") {
  mfaMode.value = mode;
  mfaChallenge.value = undefined;
  Object.assign(mfaForm, { currentPassword: "", code: "" });
  mfaVisible.value = true;
}

async function startMFASetup() {
  if (!mfaForm.currentPassword) return ElMessage.warning(t("accountCenter.messages.currentPasswordRequired"));
  try {
    mfaChallenge.value = await accountAPI.startMFA(mfaForm.currentPassword);
  } catch {
    ElMessage.error(t("accountCenter.messages.mfaFailed"));
  }
}

async function submitMFA() {
  saving.value = true;
  try {
    if (mfaMode.value === "setup") {
      if (!mfaChallenge.value) {
        await startMFASetup();
        return;
      }
      const result = await accountAPI.confirmMFA(mfaChallenge.value.challengeToken, mfaForm.code);
      recoveryCodes.value = result.recoveryCodes;
      recoveryVisible.value = true;
    } else if (mfaMode.value === "disable") {
      await accountAPI.disableMFA(mfaForm.currentPassword, mfaForm.code);
    } else {
      const result = await accountAPI.regenerateRecoveryCodes(mfaForm.currentPassword, mfaForm.code);
      recoveryCodes.value = result.recoveryCodes;
      recoveryVisible.value = true;
    }
    mfaVisible.value = false;
    ElMessage.success(t("accountCenter.messages.mfaUpdated"));
    await loadSecurity();
  } catch {
    ElMessage.error(t("accountCenter.messages.mfaFailed"));
  } finally {
    saving.value = false;
  }
}

async function copyRecoveryCodes() {
  await navigator.clipboard.writeText(recoveryCodes.value.join("\n"));
  ElMessage.success(t("accountCenter.messages.copied"));
}

async function revokeSession(row: AccountSession) {
  await ElMessageBox.confirm(t("accountCenter.sessions.revokeConfirm"), t("common.confirm"), { type: "warning" });
  await accountAPI.revokeSession(row.id);
  ElMessage.success(t("accountCenter.messages.sessionRevoked"));
  await loadSessions();
}

async function revokeOthers() {
  await ElMessageBox.confirm(t("accountCenter.sessions.revokeOthersConfirm"), t("common.confirm"), { type: "warning" });
  await accountAPI.revokeOtherSessions();
  ElMessage.success(t("accountCenter.messages.sessionsRevoked"));
  await loadSessions();
}

function navigate(tab: string) {
  const routes: Record<string, string> = {
    profile: "/account/profile",
    security: "/account/security",
    sessions: "/account/sessions",
    "login-records": "/account/login-records",
  };
  void router.push(routes[tab]);
}

onMounted(() => {
  const storedCodes = window.sessionStorage.getItem("casino-admin-recovery-codes");
  if (storedCodes) {
    try {
      recoveryCodes.value = JSON.parse(storedCodes);
      recoveryVisible.value = recoveryCodes.value.length > 0;
    } finally {
      window.sessionStorage.removeItem("casino-admin-recovery-codes");
    }
  }
  void loadPage();
});
</script>

<template>
  <main v-loading="loading" class="account-page">
    <header class="account-heading">
      <div>
        <p>{{ t("singleSite.console") }}</p>
        <h1>{{ t("accountCenter.title") }}</h1>
        <span>{{ account?.profile.username }} · {{ account?.roles.join(", ") }}</span>
      </div>
      <el-tag :type="account?.profile.mfaEnabled ? 'success' : 'danger'" effect="plain">
        {{ account?.profile.mfaEnabled ? t("accountCenter.security.mfaEnabled") : t("accountCenter.security.mfaDisabled") }}
      </el-tag>
    </header>

    <el-alert
      v-if="forcePassword"
      class="force-alert"
      type="warning"
      :closable="false"
      show-icon
      :title="t('accountCenter.security.forcePasswordTitle')"
      :description="t('accountCenter.security.forcePasswordDescription')"
    />

    <el-tabs :model-value="activeTab" class="account-tabs" @tab-change="navigate(String($event))">
      <el-tab-pane name="profile" :label="t('accountCenter.profile.title')" />
      <el-tab-pane name="security" :label="t('accountCenter.security.title')" />
      <el-tab-pane name="sessions" :label="t('accountCenter.sessions.title')" />
      <el-tab-pane name="login-records" :label="t('accountCenter.loginRecords.title')" />
    </el-tabs>

    <section v-if="activeTab === 'profile'" class="account-panel">
      <div class="panel-title"><el-icon><User /></el-icon><div><h2>{{ t("accountCenter.profile.title") }}</h2><p>{{ t("accountCenter.profile.description") }}</p></div></div>
      <el-form :model="profileForm" label-position="top" class="form-grid">
        <el-form-item :label="t('accountCenter.profile.username')"><el-input :model-value="account?.profile.username" disabled /></el-form-item>
        <el-form-item :label="t('accountCenter.profile.displayName')"><el-input v-model="profileForm.displayName" /></el-form-item>
        <el-form-item :label="t('accountCenter.profile.email')"><el-input v-model="profileForm.email" /></el-form-item>
        <el-form-item :label="t('accountCenter.profile.mobile')"><el-input v-model="profileForm.mobile" /></el-form-item>
        <el-form-item :label="t('accountCenter.profile.defaultSite')"><el-input :model-value="account?.profile.defaultSiteName || '-'" disabled /></el-form-item>
        <el-form-item :label="t('accountCenter.profile.language')"><el-select v-model="profileForm.preferredLanguage"><el-option label="简体中文" value="zh-CN" /><el-option label="English" value="en-US" /></el-select></el-form-item>
        <el-form-item :label="t('accountCenter.profile.timezone')"><el-select v-model="profileForm.timezone" filterable allow-create><el-option v-for="option in accountTimezoneOptions" :key="option" :label="option" :value="option" /></el-select></el-form-item>
      </el-form>
      <div class="panel-actions"><el-button type="primary" :loading="saving" @click="saveProfile">{{ t("common.save") }}</el-button></div>
    </section>

    <template v-else-if="activeTab === 'security'">
      <section class="account-panel">
        <div class="panel-title"><el-icon><Lock /></el-icon><div><h2>{{ t("accountCenter.security.changePassword") }}</h2><p>{{ t("accountCenter.security.passwordRule") }}</p></div></div>
        <el-form :model="passwordForm" label-position="top" class="form-grid form-grid--password">
          <el-form-item :label="t('accountCenter.security.currentPassword')"><el-input v-model="passwordForm.currentPassword" type="password" show-password /></el-form-item>
          <el-form-item :label="t('accountCenter.security.newPassword')"><el-input v-model="passwordForm.newPassword" type="password" show-password /></el-form-item>
          <el-form-item :label="t('accountCenter.security.confirmPassword')"><el-input v-model="passwordForm.confirmPassword" type="password" show-password /></el-form-item>
        </el-form>
        <el-checkbox v-model="passwordForm.revokeOtherSessions">{{ t("accountCenter.security.revokeOtherSessions") }}</el-checkbox>
        <div class="panel-actions"><el-button type="primary" :loading="saving" @click="changePassword">{{ t("accountCenter.security.updatePassword") }}</el-button></div>
      </section>

      <section class="account-panel mfa-panel">
        <div class="panel-title"><el-icon><Key /></el-icon><div><h2>{{ t("accountCenter.security.mfaTitle") }}</h2><p>{{ t("accountCenter.security.mfaDescription") }}</p></div></div>
        <div class="mfa-summary">
          <div><strong>{{ mfaStatus.enabled ? t("accountCenter.security.mfaEnabled") : t("accountCenter.security.mfaDisabled") }}</strong><span>{{ t("accountCenter.security.recoveryRemaining", { count: mfaStatus.recoveryCodesRemaining }) }}</span></div>
          <div class="mfa-actions">
            <el-button :icon="Refresh" @click="openMFA('setup')">{{ mfaStatus.enabled ? t("accountCenter.security.rebindMFA") : t("accountCenter.security.enableMFA") }}</el-button>
            <el-button v-if="mfaStatus.enabled" @click="openMFA('regenerate')">{{ t("accountCenter.security.regenerateCodes") }}</el-button>
            <el-button v-if="mfaStatus.enabled" type="danger" plain @click="openMFA('disable')">{{ t("accountCenter.security.disableMFA") }}</el-button>
          </div>
        </div>
      </section>
    </template>

    <section v-else-if="activeTab === 'sessions'" class="account-panel">
      <div class="panel-title panel-title--actions"><div class="panel-title"><el-icon><Monitor /></el-icon><div><h2>{{ t("accountCenter.sessions.title") }}</h2><p>{{ t("accountCenter.sessions.description") }}</p></div></div><el-button type="danger" plain @click="revokeOthers">{{ t("accountCenter.sessions.revokeOthers") }}</el-button></div>
      <el-table :data="sessions" stripe>
        <el-table-column :label="t('accountCenter.sessions.device')" min-width="190"><template #default="{ row }"><strong>{{ row.device }} · {{ row.browser }}</strong><div class="muted">{{ row.operatingSystem }}</div></template></el-table-column>
        <el-table-column prop="ipAddress" :label="t('accountCenter.sessions.ip')" min-width="150" />
        <el-table-column :label="t('accountCenter.sessions.lastActive')" min-width="180"><template #default="{ row }">{{ formatDate(row.lastActiveAt || row.createdAt) }}</template></el-table-column>
        <el-table-column :label="t('common.status')" width="120"><template #default="{ row }"><el-tag :type="row.status === 'ACTIVE' ? 'success' : 'info'">{{ t(`accountCenter.sessions.status.${row.status}`) }}</el-tag></template></el-table-column>
        <el-table-column :label="t('common.action')" width="120" fixed="right"><template #default="{ row }"><el-tag v-if="row.current" type="success" effect="plain">{{ t("accountCenter.sessions.current") }}</el-tag><el-button v-else-if="row.status === 'ACTIVE'" link type="danger" @click="revokeSession(row)">{{ t("accountCenter.sessions.revoke") }}</el-button></template></el-table-column>
      </el-table>
    </section>

    <section v-else class="account-panel">
      <div class="panel-title"><el-icon><Monitor /></el-icon><div><h2>{{ t("accountCenter.loginRecords.title") }}</h2><p>{{ t("accountCenter.loginRecords.description") }}</p></div></div>
      <el-table :data="loginRecords" stripe>
        <el-table-column :label="t('accountCenter.loginRecords.result')" width="110"><template #default="{ row }"><el-tag :type="row.success ? 'success' : 'danger'">{{ row.success ? t("accountCenter.loginRecords.success") : t("accountCenter.loginRecords.failed") }}</el-tag></template></el-table-column>
        <el-table-column :label="t('accountCenter.sessions.device')" min-width="190"><template #default="{ row }">{{ row.device }} · {{ row.browser }}<div class="muted">{{ row.operatingSystem }}</div></template></el-table-column>
        <el-table-column prop="ipAddress" :label="t('accountCenter.sessions.ip')" min-width="150" />
        <el-table-column :label="t('common.date')" min-width="180"><template #default="{ row }">{{ formatDate(row.createdAt) }}</template></el-table-column>
      </el-table>
    </section>

    <el-dialog v-model="mfaVisible" :title="t(`accountCenter.security.${mfaMode}Dialog`)" width="480px">
      <el-form label-position="top">
        <el-form-item :label="t('accountCenter.security.currentPassword')"><el-input v-model="mfaForm.currentPassword" type="password" show-password :disabled="Boolean(mfaChallenge)" /></el-form-item>
        <template v-if="mfaMode === 'setup' && mfaChallenge">
          <div class="mfa-setup"><QrcodeVue :value="mfaChallenge.otpauthUri" :size="176" /><code>{{ mfaChallenge.secret }}</code></div>
        </template>
        <el-form-item v-if="mfaMode !== 'setup' || mfaChallenge" :label="t('accountCenter.security.codeOrRecovery')"><el-input v-model="mfaForm.code" maxlength="16" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="mfaVisible = false">{{ t("common.cancel") }}</el-button><el-button type="primary" :loading="saving" @click="submitMFA">{{ mfaMode === 'setup' && !mfaChallenge ? t("common.confirm") : t("common.save") }}</el-button></template>
    </el-dialog>

    <el-dialog v-model="recoveryVisible" :title="t('accountCenter.security.recoveryCodes')" width="520px" :close-on-click-modal="false">
      <el-alert type="warning" :closable="false" show-icon :title="t('accountCenter.security.recoveryWarning')" />
      <div class="recovery-grid"><code v-for="code in recoveryCodes" :key="code">{{ code }}</code></div>
      <template #footer><el-button :icon="CopyDocument" @click="copyRecoveryCodes">{{ t("accountCenter.security.copyCodes") }}</el-button><el-button type="primary" @click="recoveryVisible = false">{{ t("common.confirm") }}</el-button></template>
    </el-dialog>
  </main>
</template>

<style scoped>
.account-page{display:flex;flex-direction:column;gap:18px}.account-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:20px}.account-heading p{margin:0 0 6px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.account-heading h1{margin:0;font-size:24px;letter-spacing:0}.account-heading span{display:block;margin-top:7px;color:#909399;font-size:13px}.force-alert{margin:0}.account-tabs{background:#fff;padding:0 18px;border:1px solid #e4e7ed;border-radius:6px}.account-panel{padding:22px;background:#fff;border:1px solid #e4e7ed;border-radius:6px}.panel-title{display:flex;align-items:flex-start;gap:10px}.panel-title>.el-icon{margin-top:2px;color:#337ecc;font-size:20px}.panel-title h2{margin:0;font-size:17px;letter-spacing:0}.panel-title p{margin:5px 0 0;color:#909399;font-size:13px}.panel-title--actions{align-items:center;justify-content:space-between}.form-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:0 18px;margin-top:20px}.form-grid .el-select{width:100%}.form-grid--password{grid-template-columns:repeat(3,minmax(0,1fr))}.panel-actions{display:flex;justify-content:flex-end;margin-top:16px}.mfa-panel{margin-top:0}.mfa-summary{display:flex;align-items:center;justify-content:space-between;gap:20px;margin-top:20px;padding-top:18px;border-top:1px solid #ebeef5}.mfa-summary strong,.mfa-summary span{display:block}.mfa-summary span,.muted{margin-top:4px;color:#909399;font-size:12px}.mfa-actions{display:flex;flex-wrap:wrap;justify-content:flex-end;gap:8px}.mfa-setup{display:flex;flex-direction:column;align-items:center;gap:12px;margin-bottom:18px}.mfa-setup code{font-size:13px}.recovery-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:8px;margin-top:18px}.recovery-grid code{padding:10px;background:#f5f7fa;border:1px solid #e4e7ed;border-radius:4px;text-align:center}@media(max-width:900px){.form-grid,.form-grid--password{grid-template-columns:1fr}.mfa-summary,.panel-title--actions{align-items:flex-start;flex-direction:column}.mfa-actions{justify-content:flex-start}.account-heading{align-items:flex-start;flex-direction:column}.recovery-grid{grid-template-columns:1fr}}
</style>
