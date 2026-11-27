<script setup lang="ts">
import { computed } from 'vue'
import BrandMarquee from './BrandMarquee.vue'
import { runtimeConfig } from '../services/runtimeConfig'
import { useRouter } from 'vue-router'
import type { SocialLink } from '../mock/games'
import { buildShareUrl, resolveShareText } from '../utils/shareLinks'
import { isLoggedIn } from '../mock/session'

const props = defineProps<{
  socials: SocialLink[]
}>()
const router = useRouter()
const brandName = computed(() => runtimeConfig.value.branding.appName || runtimeConfig.value.branding.siteName || '56MVP')

// Cached home data from older versions can contain both defaults and H5 rows.
const visibleSocials = computed(() => {
  const seen = new Set<string>()
  return props.socials.filter((social) => {
    const key = social.id.trim().toLowerCase()
    if (!key || seen.has(key)) return false
    seen.add(key)
    return true
  })
})

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
    <button class="quick-register" type="button" aria-label="Quick Registration" v-if="!isLoggedIn" @click="router.push('/auth/register')">
      <img :src="quickRegister" alt="Quick Registration" loading="lazy" />
    </button>

    <BrandMarquee />

    <div v-if="visibleSocials.length" class="social-links" aria-label="Social links">
      <button v-for="social in visibleSocials" :key="social.id" type="button" :aria-label="social.label" @click="openSocial(social)">
        <img :src="social.image" :alt="social.label" loading="lazy" />
      </button>
    </div>

    <nav class="legal-links" aria-label="Legal">
      <button type="button" @click="router.push('/about')">About Us</button>
      <button type="button" @click="router.push('/auth/agreement')">User Agreement</button>
      <button type="button" @click="router.push('/support')">FAQ</button>
    </nav>

    <p>
      This website offers gaming with risk experience. To be a user of our site, you must be over 21
      years of age. We are not liable for any violations of your local iGaming laws. Play responsibly
      and have fun.
    </p>
    <p>Copyright © 2019-2030 {{ brandName }}. All rights reserved.</p>
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

.social-links {
  display: flex;
  flex-wrap: wrap;
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

</style>
