<script setup lang="ts">
import { computed } from 'vue'
import { runtimeConfig } from '../services/runtimeConfig'

const isChinese = computed(() => (runtimeConfig.value.language.current || runtimeConfig.value.language.default).toLowerCase().startsWith('zh'))
const title = computed(() => runtimeConfig.value.maintenance.title || (isChinese.value ? '站点维护中' : 'Site maintenance'))
const message = computed(() => runtimeConfig.value.maintenance.message || (isChinese.value ? '我们正在进行系统维护，请稍后再试。' : 'We are carrying out scheduled maintenance. Please try again later.'))
</script>

<template>
  <main class="maintenance-screen">
    <section>
      <span class="maintenance-mark" aria-hidden="true">56</span>
      <p>{{ isChinese ? '服务暂不可用' : 'SERVICE TEMPORARILY UNAVAILABLE' }}</p>
      <h1>{{ title }}</h1>
      <div class="maintenance-line" />
      <span class="maintenance-message">{{ message }}</span>
      <a v-if="runtimeConfig.maintenance.supportURL" :href="runtimeConfig.maintenance.supportURL" target="_blank" rel="noopener noreferrer">
        {{ isChinese ? '联系客服' : 'Contact support' }}
      </a>
    </section>
  </main>
</template>

<style scoped>
.maintenance-screen{min-height:100dvh;display:grid;place-items:center;padding:28px;background:#07101f;color:#fff}.maintenance-screen section{width:min(520px,100%);text-align:center}.maintenance-mark{display:grid;place-items:center;width:72px;height:72px;margin:0 auto 24px;border:2px solid #f4c653;border-radius:50%;color:#f4c653;font-size:25px;font-weight:800}.maintenance-screen p{margin:0 0 12px;color:#f4c653;font-size:12px;font-weight:700}.maintenance-screen h1{margin:0;font-size:clamp(28px,7vw,48px);line-height:1.15;letter-spacing:0}.maintenance-line{width:64px;height:3px;margin:24px auto;background:#f4c653}.maintenance-message{display:block;color:#c9d3e2;font-size:16px;line-height:1.7}.maintenance-screen a{display:inline-flex;margin-top:28px;padding:11px 18px;border:1px solid #f4c653;border-radius:6px;color:#f4c653;text-decoration:none}
</style>
