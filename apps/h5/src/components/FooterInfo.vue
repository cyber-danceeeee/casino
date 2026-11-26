<script setup lang="ts">
import type { HomeProvider } from '../services/h5Api'
import { toRuntimeAssetUrl } from '../services/runtimeConfig'
import type { SocialLink } from '../mock/games'
import { buildShareUrl, resolveShareText } from '../utils/shareLinks'
import { isLoggedIn } from '../mock/session'

defineProps<{
  socials: SocialLink[]
  providers: HomeProvider[]
}>()

const quickRegister = new URL('../assets/images/quick-register.webp', import.meta.url).href

const openSocial = async (social: SocialLink) => {
  if (typeof window === 'undefined') return

  const targetUrl = `${window.location.origin}/`
  const shareUrl = buildShareUrl(social, targetUrl)
  if (!shareUrl) return

  if (social.id === 'instagram' || social.id === 'tiktok') {
    await navigator.clipboard?.writeText(resolveShareText(social, targetUrl)).catch(() => undefined)
  }
  window.open(shareUrl, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <footer class="footer-info">
    <button class="quick-register" type="button" aria-label="Quick Registration" v-if="!isLoggedIn">
      <img :src="quickRegister" alt="Quick Registration" loading="lazy" />
    </button>

    <div v-if="providers.length" class="provider-marquee" aria-label="Game providers">
      <div class="provider-marquee__track">
        <div
          v-for="(provider, index) in [...providers, ...providers]"
          :key="`${provider.providerCode}-${index}`"
          class="provider-logo"
        >
          <img v-if="provider.logoUrl" :src="toRuntimeAssetUrl(provider.logoUrl)" :alt="provider.name" loading="lazy" /><span v-else>{{ provider.name }}</span>
        </div>
      </div>
    </div>

    <div class="social-links" aria-label="Social links">
      <button v-for="social in socials" :key="social.id" type="button" :aria-label="social.label" @click="openSocial(social)">
        <img :src="social.image" :alt="social.label" loading="lazy" />
      </button>
    </div>

    <nav class="legal-links" aria-label="Legal">
      <button type="button">About Us</button>
      <button type="button">User Agreement</button>
      <button type="button">FAQ</button>
    </nav>

    <p>
      This website offers gaming with risk experience. To be a user of our site, you must be over 21
      years of age. We are not liable for any violations of your local iGaming laws. Play responsibly
      and have fun.
    </p>
    <p>Copyright © 2019-2030 11N. All rights reserved.</p>
  </footer>
</template>

<style scoped lang="scss">
.footer-info {
  padding: 4px 0 108px;
  border-top: 1px solid rgba(73, 86, 126, 0.62);
}

.quick-register {
  width: 100%;
  padding: 0;
  border-radius: 14px;

  img {
    width: 100%;
    aspect-ratio: 347 / 142;
    object-fit: contain;
  }
}

.provider-marquee {
  margin: 16px -12px 0;
  overflow: hidden;
}

.provider-marquee__track {
  display: flex;
  width: max-content;
  animation: provider-scroll 24s linear infinite;
}

.provider-logo {
  flex: 0 0 112px;
  padding-left: 8px;

  img {
    width: 112px;
    height: 44px;
    border-radius: 4px;
    object-fit: contain;
    background: rgba(255, 255, 255, 0.04);
    opacity: 0.82;
  }
}

.social-links {
  display: flex;
  justify-content: center;
  gap: 13px;
  margin-top: 22px;

  button {
    width: 36px;
    height: 36px;
    padding: 0;
    overflow: hidden;
    border-radius: 50%;
  }

  img {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: contain;
  }
}

.legal-links {
  display: flex;
  justify-content: space-around;
  gap: 16px;
  margin: 22px 2px 0;

  button {
    color: #ffffff;
    font-size: 13px;
    font-weight: 500;
    line-height: 1;
    text-decoration: underline;
  }
}

p {
  margin: 22px 8px 0;
  color: rgba(255, 255, 255, 0.5);
  font-size: 13px;
  font-weight: 500;
  line-height: 1.6;
  text-align: center;
}

@keyframes provider-scroll {
  from {
    transform: translate3d(0, 0, 0);
  }

  to {
    transform: translate3d(-50%, 0, 0);
  }
}
</style>
