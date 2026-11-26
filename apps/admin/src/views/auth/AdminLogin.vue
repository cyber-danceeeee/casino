<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { ArrowLeft, Key, Lock, User } from "@element-plus/icons-vue";
import QrcodeVue from "qrcode.vue";
import { useI18n } from "vue-i18n";
import {
  confirmAdminMFASetup,
  loginAdmin,
  verifyAdminMFA,
  type AdminAuthResponse,
} from "@/api/adminAuth";
import { useAdminAuth } from "@/composables/useAdminAuth";
import { localeOptions, type SupportedLocale } from "@/i18n/messages";
import logoIcon from "@/assets/logo_icon.svg";

type LoginStep = "credentials" | "setup" | "verify";

const route = useRoute();
const router = useRouter();
const { t, locale } = useI18n();
const { establishSession } = useAdminAuth();

const step = ref<LoginStep>("credentials");
const loading = ref(false);
const challenge = ref<AdminAuthResponse | null>(null);
const credentials = reactive({ username: "", password: "" });
const verificationCode = ref("");

const currentLocale = computed({
  get: () => locale.value as SupportedLocale,
  set: (value: SupportedLocale) => {
    locale.value = value;
    window.localStorage.setItem("casino-admin-locale", value);
  },
});

const title = computed(() => {
  if (step.value === "setup") return t("auth.setupTitle");
  if (step.value === "verify") return t("auth.verifyTitle");
  return t("auth.loginTitle");
});

const subtitle = computed(() => {
  if (step.value === "setup") return t("auth.setupSubtitle");
  if (step.value === "verify") return t("auth.verifySubtitle");
  return t("auth.loginSubtitle");
});

function resetChallenge() {
  challenge.value = null;
  verificationCode.value = "";
  step.value = "credentials";
}

async function finishLogin(response: AdminAuthResponse) {
  if (!response.token) throw new Error("missing authentication token");
  if (response.recoveryCodes?.length) {
    window.sessionStorage.setItem(
      "casino-admin-recovery-codes",
      JSON.stringify(response.recoveryCodes),
    );
  }
  await establishSession(response.token, response.administrator);
  if (response.administrator.mustChangePassword) {
    await router.replace({ path: "/account/security", query: { forcePassword: "1" } });
    return;
  }
  const redirect =
    typeof route.query.redirect === "string" &&
    route.query.redirect.startsWith("/")
      ? route.query.redirect
      : "/";
  await router.replace(redirect);
}

async function submitCredentials() {
  credentials.username = credentials.username.trim();
  if (
    !credentials.username ||
    !credentials.password
  ) {
    ElMessage.warning(t("auth.requiredFields"));
    return;
  }
  loading.value = true;
  try {
    const response = await loginAdmin({ ...credentials });
    challenge.value = response;
    verificationCode.value = "";
    if (response.status === "MFA_SETUP_REQUIRED") step.value = "setup";
    else if (response.status === "MFA_REQUIRED") step.value = "verify";
    else await finishLogin(response);
  } catch {
    ElMessage.error(t("auth.loginFailed"));
  } finally {
    loading.value = false;
  }
}

async function submitMFA() {
  const code = step.value === "setup"
    ? verificationCode.value.replace(/\D/g, "").slice(0, 6)
    : verificationCode.value.trim().toUpperCase().slice(0, 16);
  verificationCode.value = code;
  if ((step.value === "setup" && code.length !== 6) || !code || !challenge.value?.challengeToken) {
    ElMessage.warning(t("auth.codeRequired"));
    return;
  }
  loading.value = true;
  try {
    const payload = { challengeToken: challenge.value.challengeToken, code };
    const response =
      step.value === "setup"
        ? await confirmAdminMFASetup(payload)
        : await verifyAdminMFA(payload);
    await finishLogin(response);
  } catch {
    ElMessage.error(t("auth.invalidCode"));
  } finally {
    loading.value = false;
  }
}

async function copySecret() {
  if (!challenge.value?.secret) return;
  await navigator.clipboard.writeText(challenge.value.secret);
  ElMessage.success(t("auth.secretCopied"));
}
</script>

<template>
  <main class="login-page">
    <section class="login-brand-panel" aria-hidden="true">
      <div class="login-brand-mark">
        <img :src="logoIcon" alt="" />
        <div>
          <strong>{{ t("app.brand") }}</strong>
          <span>{{ t("auth.consoleLabel") }}</span>
        </div>
      </div>
      <div class="login-brand-content">
        <p>{{ t("singleSite.console") }}</p>
        <h1>{{ t("auth.brandTitle") }}</h1>
        <span>{{ t("auth.brandDescription") }}</span>
      </div>
      <div class="login-brand-footer">{{ t("auth.securityNotice") }}</div>
    </section>

    <section class="login-form-panel">
      <el-select
        v-model="currentLocale"
        class="login-locale"
        :aria-label="t('app.language')"
      >
        <el-option
          v-for="option in localeOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
      <div class="login-mobile-brand">
        <img :src="logoIcon" alt="Casino Admin" />
        <strong>{{ t("app.brand") }}</strong>
      </div>

      <div class="login-form-wrap">
        <button
          v-if="step !== 'credentials'"
          type="button"
          class="login-back"
          @click="resetChallenge"
        >
          <el-icon><ArrowLeft /></el-icon>
          {{ t("auth.backToLogin") }}
        </button>

        <header class="login-heading">
          <span class="login-step-icon"
            ><el-icon
              ><Key v-if="step !== 'credentials'" /><Lock v-else /></el-icon
          ></span>
          <h2>{{ title }}</h2>
          <p>{{ subtitle }}</p>
        </header>

        <form
          v-if="step === 'credentials'"
          class="login-form"
          @submit.prevent="submitCredentials"
        >
          <label>
            <span>{{ t("auth.username") }}</span>
            <el-input
              v-model="credentials.username"
              size="large"
              autocomplete="username"
              :placeholder="t('auth.usernamePlaceholder')"
            >
              <template #prefix
                ><el-icon><User /></el-icon
              ></template>
            </el-input>
          </label>
          <label>
            <span>{{ t("auth.password") }}</span>
            <el-input
              v-model="credentials.password"
              size="large"
              type="password"
              autocomplete="current-password"
              show-password
              :placeholder="t('auth.passwordPlaceholder')"
            >
              <template #prefix
                ><el-icon><Lock /></el-icon
              ></template>
            </el-input>
          </label>
          <el-button
            class="login-submit"
            type="primary"
            size="large"
            native-type="submit"
            :loading="loading"
          >
            {{ t("auth.signIn") }}
          </el-button>
        </form>

        <div v-else class="mfa-form">
          <template v-if="step === 'setup'">
            <div class="mfa-qr">
              <QrcodeVue
                v-if="challenge?.otpauthUri"
                :value="challenge.otpauthUri"
                :size="180"
                level="M"
              />
            </div>
            <ol class="mfa-instructions">
              <li>{{ t("auth.setupStepOne") }}</li>
              <li>{{ t("auth.setupStepTwo") }}</li>
              <li>{{ t("auth.setupStepThree") }}</li>
            </ol>
            <button type="button" class="mfa-secret" @click="copySecret">
              <span>{{ t("auth.manualKey") }}</span>
              <strong>{{ challenge?.secret }}</strong>
              <em>{{ t("auth.copy") }}</em>
            </button>
          </template>

          <label class="mfa-code-label">
            <span>{{ t("auth.verificationCode") }}</span>
            <el-input
              v-model="verificationCode"
              size="large"
              :inputmode="step === 'setup' ? 'numeric' : 'text'"
              :maxlength="step === 'setup' ? 6 : 16"
              autocomplete="one-time-code"
              :placeholder="step === 'setup' ? t('auth.codePlaceholder') : t('auth.codeOrRecoveryPlaceholder')"
              @keyup.enter="submitMFA"
            />
          </label>
          <el-button
            class="login-submit"
            type="primary"
            size="large"
            :loading="loading"
            @click="submitMFA"
          >
            {{
              step === "setup"
                ? t("auth.bindAndContinue")
                : t("auth.verifyAndLogin")
            }}
          </el-button>
          <p class="mfa-expiry">{{ t("auth.challengeExpiry") }}</p>
        </div>

      </div>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(360px, 42%) 1fr;
  background: #f5f7fa;
  color: #303133;
}
.login-brand-panel {
  position: relative;
  display: flex;
  min-height: 100vh;
  flex-direction: column;
  overflow: hidden;
  padding: 42px 52px;
  color: #fff;
  background: linear-gradient(145deg, #173b74 0%, #245ca8 48%, #16375f 100%);
}
.login-brand-panel::after {
  content: "";
  position: absolute;
  inset: 0;
  opacity: 0.18;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.12) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.12) 1px, transparent 1px);
  background-size: 36px 36px;
  mask-image: linear-gradient(to bottom, #000, transparent 82%);
}
.login-brand-mark,
.login-brand-content,
.login-brand-footer {
  position: relative;
  z-index: 1;
}
.login-brand-mark {
  display: flex;
  align-items: center;
  gap: 13px;
}
.login-brand-mark img {
  width: 42px;
  height: 42px;
}
.login-brand-mark div {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.login-brand-mark strong {
  font-size: 17px;
}
.login-brand-mark span {
  font-size: 11px;
  letter-spacing: 0.14em;
  color: rgba(255, 255, 255, 0.66);
}
.login-brand-content {
  max-width: 470px;
  margin: auto 0;
}
.login-brand-content p {
  margin: 0 0 18px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #9bc7ff;
}
.login-brand-content h1 {
  max-width: 430px;
  margin: 0 0 20px;
  font-size: clamp(30px, 3.2vw, 48px);
  line-height: 1.18;
  letter-spacing: 0;
}
.login-brand-content > span {
  display: block;
  max-width: 430px;
  font-size: 14px;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.72);
}
.login-brand-footer {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.55);
}
.login-form-panel {
  position: relative;
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
  padding: 40px;
}
.login-locale {
  position: absolute;
  top: 24px;
  right: 28px;
  width: 112px;
}
.login-form-wrap {
  width: min(420px, 100%);
}
.portal-switch {
  display: block;
  margin-top: 22px;
  color: #337ecc;
  font-size: 13px;
  text-align: center;
  text-decoration: none;
}
.login-heading {
  margin-bottom: 28px;
}
.login-step-icon {
  display: inline-flex;
  width: 42px;
  height: 42px;
  align-items: center;
  justify-content: center;
  margin-bottom: 18px;
  border-radius: 8px;
  color: #fff;
  background: #409eff;
  box-shadow: 0 8px 18px rgba(64, 158, 255, 0.25);
}
.login-step-icon .el-icon {
  font-size: 21px;
}
.login-heading h2 {
  margin: 0 0 9px;
  font-size: 25px;
  font-weight: 600;
}
.login-heading p {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: #909399;
}
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.login-form label,
.mfa-code-label {
  display: flex;
  flex-direction: column;
  gap: 7px;
  font-size: 13px;
  font-weight: 500;
  color: #606266;
}
.login-submit {
  width: 100%;
  height: 44px;
  margin-top: 6px;
}
.login-back {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  margin: 0 0 22px;
  padding: 0;
  border: 0;
  color: #606266;
  background: transparent;
  cursor: pointer;
}
.mfa-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.mfa-qr {
  align-self: center;
  padding: 13px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  line-height: 0;
}
.mfa-instructions {
  display: grid;
  gap: 7px;
  margin: 0;
  padding-left: 22px;
  color: #606266;
  font-size: 12px;
  line-height: 1.55;
}
.mfa-secret {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 5px 12px;
  width: 100%;
  padding: 10px 12px;
  text-align: left;
  border: 1px dashed #b7c7df;
  border-radius: 6px;
  color: #606266;
  background: #f4f8fd;
  cursor: pointer;
}
.mfa-secret span {
  font-size: 11px;
}
.mfa-secret strong {
  grid-column: 1;
  overflow-wrap: anywhere;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  color: #303133;
}
.mfa-secret em {
  grid-column: 2;
  grid-row: 1 / 3;
  align-self: center;
  font-size: 12px;
  font-style: normal;
  color: #409eff;
}
.mfa-expiry {
  margin: -6px 0 0;
  text-align: center;
  font-size: 11px;
  color: #a8abb2;
}
.login-mobile-brand {
  display: none;
}
@media (max-width: 820px) {
  .login-page {
    grid-template-columns: 1fr;
  }
  .login-brand-panel {
    display: none;
  }
  .login-form-panel {
    min-height: 100vh;
    flex-direction: column;
    align-items: flex-start;
    justify-content: flex-start;
    padding: 32px 24px;
    background: #fff;
  }
  .login-locale {
    top: 27px;
    right: 24px;
    width: 104px;
  }
  .login-mobile-brand {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 58px;
    font-size: 14px;
  }
  .login-mobile-brand img {
    width: 34px;
    height: 34px;
  }
}
</style>
