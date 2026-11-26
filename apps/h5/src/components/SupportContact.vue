<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import avatar from '../assets/images/reference/support-avatar.webp'

const route = useRoute()
const router = useRouter()
const visible = computed(() => ['home', 'offers', 'vip', 'share'].includes(String(route.name))
  && !(route.name === 'share' && ['mine', 'agent'].includes(String(route.query.tab))))
</script>

<template>
  <button v-if="visible" class="support-contact" type="button" aria-label="Contact support" @click="router.push('/support')">
    <img :src="avatar" alt="" />
    <span>Contact Me</span>
  </button>
</template>

<style scoped>
.support-contact { position: fixed; z-index: 45; right: max(10px,calc((100vw - 430px)/2 + 10px)); bottom: calc(224px + env(safe-area-inset-bottom)); display: grid; justify-items: center; width: 72px; padding: 0; }
.support-contact img { width: 72px; height: 55px; object-fit: contain; }
.support-contact span { margin-top: -4px; padding: 2px 6px; border: 1px solid #00a1d5; border-radius: 20px; background: linear-gradient(#bcefff,#39cfff); color: #003278; font-size: 10px; font-weight: 600; line-height: 1.1; white-space: nowrap; }
</style>
