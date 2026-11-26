<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { loginMockUser } from '../mock/session'
import { normalizePhoneInput, phoneCountryCode, phoneLength } from '../services/runtimeConfig'
import { getSafeRedirect } from '../utils/authRedirect'
import customer from '../assets/images/profile/customer.webp'
import headerBg from '../assets/images/auth-header-bg.webp'

type AuthMode = 'login' | 'register'
type LoginMethod = 'password' | 'code'

const route = useRoute()
const router = useRouter()

const phone = ref('')
const password = ref('')
const inviteCode = ref('')
const verificationCode = ref('')
const passwordVisible = ref(true)
const agreementAccepted = ref(true)
const loginMethod = ref<LoginMethod>('password')
const isSubmitting = ref(false)
const authError = ref('')

const mode = computed<AuthMode>(() => (route.name === 'auth-register' ? 'register' : 'login'))
const isRegister = computed(() => mode.value === 'register')
const title = computed(() => (isRegister.value ? 'Register' : 'Login'))
const subtitle = computed(() =>
  isRegister.value ? 'Already have an account?' : "Don't have an account？",
)
const switchLabel = computed(() => (isRegister.value ? 'Login' : 'Register'))
const actionLabel = computed(() => (isRegister.value ? 'Register' : 'Login'))
const passwordInputType = computed(() => (passwordVisible.value ? 'text' : 'password'))
const runtimePhoneCode = computed(() => phoneCountryCode())
const runtimePhoneLength = computed(() => phoneLength())
const runtimePhonePattern = computed(() => `[0-9]{${runtimePhoneLength.value}}`)

const switchAuthMode = () => {
  const redirect = getSafeRedirect(route.query.redirect)
  void router.push({
    name: isRegister.value ? 'auth-login' : 'auth-register',
    query: redirect ? { redirect } : undefined,
  })
}

const goHome = () => {
  void router.push({ name: 'home' })
}

const toggleLoginMethod = () => {
  loginMethod.value = loginMethod.value === 'password' ? 'code' : 'password'
}

const limitPhoneInput = (event: Event) => {
  const input = event.target as HTMLInputElement
  const value = normalizePhoneInput(input.value)
  phone.value = value
  input.value = value
}

const submit = async () => {
  if (isSubmitting.value) return

  authError.value = ''
  isSubmitting.value = true

  try {
    await loginMockUser(
      phone.value,
      loginMethod.value === 'password' ? password.value : verificationCode.value,
      inviteCode.value,
      isRegister.value ? 'register' : 'login',
    )
    const redirect = isRegister.value ? null : getSafeRedirect(route.query.redirect)
    await router.push(redirect || { name: 'home' })
  } catch (error) {
    authError.value = error instanceof Error ? error.message : 'Login failed'
  } finally {
    isSubmitting.value = false
  }
}

watch(
  mode,
  () => {
    passwordVisible.value = true
    loginMethod.value = 'password'
    authError.value = ''
  },
)
</script>

<template>
  <section class="auth-page" :aria-label="`${title} page`">
    <header class="auth-hero" :style="{ backgroundImage: `url(${headerBg})` }">
      <button class="auth-back" type="button" aria-label="Back" @click="goHome">
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M4.3 11.3a1 1 0 0 0 0 1.4l4 4a1 1 0 0 0 1.4-1.4L6.4 12l3.3-3.3a1 1 0 0 0-1.4-1.4l-4 4ZM5 13h14a1 1 0 1 0 0-2H5v2Z" />
        </svg>
      </button>

      <button class="auth-support" type="button">
        <img :src="customer" alt="" />
        <span>Contact Me</span>
      </button>

      <div class="auth-title-block">
        <h1>{{ title }}</h1>
        <p>
          {{ subtitle }}
          <button type="button" @click="switchAuthMode">{{ switchLabel }}</button>
        </p>
      </div>
    </header>

    <main class="auth-card">
      <form class="auth-form" @submit.prevent="submit">
        <label class="auth-field">
          <span><i>*</i>Phone Number</span>
          <div class="auth-input auth-input--phone">
            <strong>{{ runtimePhoneCode }}</strong>
            <em />
            <input
              v-model="phone"
              inputmode="numeric"
              :pattern="runtimePhonePattern"
              :maxlength="runtimePhoneLength"
              autocomplete="tel"
              placeholder="3012345678"
              type="tel"
              @input="limitPhoneInput"
            />
          </div>
        </label>

        <div class="auth-field auth-field--password">
          <div class="auth-label-row">
            <span><i>*</i>Password</span>
            <button v-if="!isRegister" class="auth-code-switch" type="button" @click="toggleLoginMethod">
              <svg viewBox="0 0 16 16" aria-hidden="true">
                <path d="M0 11.4V15.2A.8.8 0 0 0 .8 16h3.8v-1.2H1.9a.68.68 0 0 1-.7-.68V11.4H0Zm1.2-9.52c0-.38.3-.68.68-.68H4.6V0H.8A.8.8 0 0 0 0 .8v3.8h1.2V1.88ZM16 .8a.8.8 0 0 0-.8-.8h-3.8v1.2h2.72c.38 0 .68.3.68.68V4.6H16V.8Zm-1.2 13.32c0 .38-.3.68-.68.68H11.4V16h3.8a.8.8 0 0 0 .8-.8v-3.8h-1.2v2.72ZM2 6.9v-.8l1.4-.3h.7v4.7h-.9V6.6L2 6.9Zm6.5.3c0-.2-.1-.4-.2-.5-.2-.1-.4-.2-.7-.2-.4 0-.7.1-1.1.2v-.8c.4-.1.8-.2 1.2-.2.5 0 .9.1 1.2.3.3.3.5.6.4 1 0 .3 0 .5-.1.8-.1.3-.3.5-.5.7-.3.3-.7.7-1.1 1l-.3.2v.1h2.1v.8H6.3v-.9c.6-.5 1-.9 1.4-1.1.2-.2.4-.5.6-.8.1-.2.2-.4.2-.6Zm3.1 3.3v-.8c.3.1.7.1 1 .1.3 0 .6 0 .8-.1.2-.1.3-.3.3-.6 0-.2-.1-.4-.3-.5-.3-.1-.6-.2-.9-.2h-.4v-.7h.4c.3 0 .6-.1.8-.2.2-.1.3-.3.3-.5s-.1-.4-.2-.5c-.2-.1-.4-.2-.7-.2-.3 0-.6.1-.9.2V5.9c.3-.1.7-.2 1.1-.2.4 0 .9.1 1.2.3.3.2.5.6.4 1 0 .5-.3.9-.8 1.1.3 0 .5.2.7.4.2.2.3.5.3.8 0 .9-.7 1.4-2 1.4-.4-.1-.7-.1-1.1-.2Z" />
              </svg>
              {{ loginMethod === 'password' ? 'Verification Code' : 'Password' }}
            </button>
          </div>

          <div v-if="loginMethod === 'password'" class="auth-input auth-input--password">
            <input
              v-model="password"
              :type="passwordInputType"
              autocomplete="off"
              placeholder="Enter"
            />
            <button class="auth-eye" type="button" aria-label="Toggle password visibility" @click="passwordVisible = !passwordVisible">
              <svg v-if="passwordVisible" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M2 12s3.4-6.5 10-6.5S22 12 22 12s-3.4 6.5-10 6.5S2 12 2 12Zm10 3a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z" />
              </svg>
              <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                <path d="m3 3 18 18-1.4 1.4-3.1-3.1A11.8 11.8 0 0 1 12 20C5.4 20 2 12 2 12a17.4 17.4 0 0 1 4.4-5.3L1.6 4.4 3 3Zm6.1 6.1A3 3 0 0 0 12 15c.8 0 1.5-.3 2-.8L9.1 9.1ZM12 4c6.6 0 10 8 10 8a17.5 17.5 0 0 1-3.2 4.2l-3-3A3 3 0 0 0 10.8 8l-2.2-2.2A11.1 11.1 0 0 1 12 4Z" />
              </svg>
            </button>
          </div>

          <div v-else class="auth-code-row">
            <div class="auth-input">
              <input
                v-model="verificationCode"
                inputmode="numeric"
                maxlength="6"
                autocomplete="one-time-code"
                placeholder="Enter the 6-digit code"
                type="tel"
              />
            </div>
            <button class="auth-send" type="button" disabled>Send</button>
          </div>
        </div>

        <label v-if="isRegister" class="auth-field">
          <span>Invitation Code (Optional)</span>
          <div class="auth-input">
            <input v-model="inviteCode" placeholder="Enter" type="text" />
            <button class="auth-paste" type="button">Paste</button>
          </div>
        </label>

        <label class="auth-agreement">
          <input v-model="agreementAccepted" type="checkbox" />
          <span class="auth-check" aria-hidden="true">
            <svg viewBox="0 0 12 12">
              <path d="M2.5 6 5 8.5 9.5 3.5" />
            </svg>
          </span>
          <span>I am over 18 and agree to the <a href="#">User Agreement</a></span>
        </label>

        <p v-if="authError" class="auth-error">{{ authError }}</p>

        <button class="auth-submit" type="submit" :disabled="isSubmitting">
          {{ isSubmitting ? 'Please wait...' : actionLabel }}
        </button>

        <div v-if="!isRegister" class="auth-login-tip">
          <svg viewBox="0 0 16 16" fill="none" aria-hidden="true">
            <circle cx="8" cy="8" r="7" stroke="currentColor" stroke-width="1.2" />
            <path d="M8 7v4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
            <circle cx="8" cy="5" r="0.8" fill="currentColor" />
          </svg>
          <span>Automatically register on first login</span>
        </div>
      </form>
    </main>
  </section>
</template>

<style scoped lang="scss">
.auth-page {
  position: fixed;
  inset: 0;
  left: 50%;
  z-index: 850;
  display: flex;
  width: 100%;
  max-width: 430px;
  min-height: 100dvh;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
  background: #090f1f;
  color: #ffffff;
  transform: translateX(-50%);
  -webkit-overflow-scrolling: touch;
}

.auth-hero {
  position: relative;
  height: 227px;
  flex: 0 0 227px;
  padding: calc(env(safe-area-inset-top) + 24px) 24px 24px;
  overflow: hidden;
  background-position: center;
  background-repeat: no-repeat;
  background-size: cover;
}

.auth-back {
  position: relative;
  z-index: 2;
  display: grid;
  width: 24px;
  height: 24px;
  place-items: center;
  padding: 0;
  color: #ffffff;

  svg {
    width: 24px;
    height: 24px;
    fill: currentColor;
  }
}

.auth-support {
  position: absolute;
  top: calc(env(safe-area-inset-top) + 24px);
  right: 24px;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0;

  img {
    width: 44px;
    height: 55px;
    object-fit: contain;
  }

  span {
    margin-top: -6px;
    border: 1px solid #00a1d5;
    border-radius: 999px;
    padding: 2px 10px 3px;
    background: linear-gradient(180deg, #baefff 0%, #38cfff 100%);
    color: #003278;
    font-size: 10px;
    font-weight: 700;
    line-height: 1;
    white-space: nowrap;
  }
}

.auth-title-block {
  position: relative;
  z-index: 2;
  margin-top: 24px;

  h1 {
    margin: 0;
    color: #ffffff;
    font-size: 36px;
    font-weight: 650;
    line-height: 1.3;
    letter-spacing: 0;
  }

  p {
    margin: 12px 0 0;
    color: #ffffff;
    font-size: 12px;
    font-weight: 500;
    line-height: 1.4;
  }

  button {
    padding: 0;
    color: #4d81e7;
    font-size: inherit;
    font-weight: 700;
    text-decoration: underline;
  }
}

.auth-card {
  position: relative;
  z-index: 3;
  flex: 1 0 auto;
  margin-top: -48px;
  border-radius: 16px 16px 0 0;
  padding: 24px 24px calc(env(safe-area-inset-bottom) + 24px);
  background: #1e2438;
}

.auth-form {
  display: block;
}

.auth-field {
  display: grid;
  gap: 8px;

  > span,
  .auth-label-row span {
    color: #ffffff;
    font-size: 12px;
    line-height: 1;
  }

  i {
    margin-right: 4px;
    color: #ef4444;
    font-style: normal;
  }

  + .auth-field {
    margin-top: 16px;
  }
}

.auth-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 16px;
}

.auth-input {
  position: relative;
  display: flex;
  height: 46px;
  align-items: center;
  overflow: hidden;
  border: 1px solid #364169;
  border-radius: 10px;
  background: #131828;
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease;

  &:focus-within {
    border-color: #4d81e7;
    box-shadow: 0 0 0 2px rgba(77, 129, 231, 0.12);
  }

  strong {
    display: grid;
    height: 100%;
    place-items: center;
    padding: 0 12px;
    color: #ffffff;
    font-size: 14px;
    font-weight: 500;
  }

  em {
    width: 1px;
    height: 100%;
    background: #364169;
  }

  input {
    min-width: 0;
    height: 100%;
    flex: 1;
    border: 0;
    outline: 0;
    padding: 0 12px;
    background: transparent;
    color: #ffffff;
    font-size: 14px;

    &::placeholder {
      color: #46537f;
    }
  }
}

.auth-input--phone input {
  padding-left: 12px;
}

.auth-code-switch {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 0;
  color: #4d81e7;
  font-size: 12px;
  font-weight: 700;

  svg {
    width: 16px;
    height: 16px;
    fill: currentColor;
  }
}

.auth-eye {
  position: absolute;
  top: 50%;
  right: 12px;
  display: grid;
  width: 24px;
  height: 24px;
  place-items: center;
  padding: 0;
  color: #97a5ca;
  transform: translateY(-50%);

  svg {
    width: 22px;
    height: 22px;
    fill: currentColor;
  }
}

.auth-code-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 96px;
  gap: 12px;
}

.auth-send {
  height: 46px;
  border-radius: 8px;
  background: #4d81e7;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;

  &:disabled {
    opacity: 0.5;
  }
}

.auth-paste {
  height: 100%;
  flex: 0 0 auto;
  padding: 0 12px;
  color: #4d81e7;
  font-size: 12px;
  font-weight: 700;
}

.auth-agreement {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 32px;
  color: #97a5ca;
  font-size: 12px;
  line-height: 1.4;

  input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  a {
    color: #58b86b;
    text-decoration: underline;
  }
}

.auth-check {
  display: grid;
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid #4d81e7;
  border-radius: 4px;
  background: #4d81e7;

  svg {
    width: 12px;
    height: 12px;
  }

  path {
    fill: none;
    stroke: #ffffff;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-width: 1.5;
  }
}

.auth-agreement input:not(:checked) + .auth-check {
  background: transparent;

  svg {
    opacity: 0;
  }
}

.auth-submit {
  display: grid;
  width: 100%;
  height: 46px;
  place-items: center;
  margin-top: 12px;
  border-radius: 8px;
  background: linear-gradient(180deg, #438cf0 0%, #2f73de 100%);
  box-shadow: 0 10px 26px rgba(57, 127, 229, 0.26);
  color: #ffffff;
  font-size: 16px;
  font-weight: 700;
  line-height: 1;

  &:disabled {
    cursor: not-allowed;
    opacity: 0.72;
  }
}

.auth-error {
  margin: 12px 0 0;
  color: #ff7682;
  font-size: 12px;
  line-height: 1.35;
}

.auth-login-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
  line-height: 1;

  svg {
    width: 14px;
    height: 14px;
    flex: 0 0 auto;
  }
}

@media (max-width: 374px) {
  .auth-hero {
    padding-right: 20px;
    padding-left: 20px;
  }

  .auth-card {
    padding-right: 20px;
    padding-left: 20px;
  }
}
</style>
