<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { resolvePopupComponent } from '../components/popups/popupRegistry'
import type { H5PopupPayload } from '../services/h5Api'

const route = useRoute()
const router = useRouter()
const componentCode = computed(() => String(route.params.componentCode || '').toUpperCase())
const popupComponent = computed(() => resolvePopupComponent(componentCode.value))
const actionType = computed(() => String(route.query.actionType || 'CLOSE_ONLY') as H5PopupPayload['action']['type'])
const campaign = computed<H5PopupPayload>(() => ({
  campaignId: 0,
  componentCode: componentCode.value,
  priority: Number(route.query.priority || 0),
  triggerType: String(route.query.triggerType || 'HOME_ENTER') as H5PopupPayload['triggerType'],
  triggerValue: String(route.query.triggerValue || ''),
  frequency: {
    type: String(route.query.frequencyType || 'EVERY_VISIT') as H5PopupPayload['frequency']['type'],
    displayLimit: Number(route.query.displayLimit || 0),
    cooldownSeconds: Number(route.query.cooldownMinutes || 0) * 60,
  },
  action: { type: actionType.value, target: String(route.query.actionTarget || '') },
  configVersion: 0,
}))
const close = () => router.push({ name: 'home' })
</script>

<template>
  <component :is="popupComponent" v-if="popupComponent" :campaign="campaign" @close="close" @action="close" />
  <main v-else class="preview-missing"><h1>Popup component unavailable</h1><code>{{ componentCode }}</code></main>
</template>

<style scoped>
.preview-missing{display:grid;min-height:100vh;place-content:center;gap:12px;padding:24px;color:#f5f6f8;text-align:center;background:#090b10}.preview-missing h1{margin:0;font-size:22px;letter-spacing:0}.preview-missing code{color:#ffc54a}
</style>
