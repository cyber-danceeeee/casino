<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import BackgroundMusic from './components/BackgroundMusic.vue'
import BottomNav from './components/BottomNav.vue'
import SiteOperationsLayer from './components/SiteOperationsLayer.vue'
import MaintenanceScreen from './components/MaintenanceScreen.vue'
import SupportContact from './components/SupportContact.vue'
import { runtimeConfig } from './services/runtimeConfig'

const route = useRoute()
const showBottomNav = computed(() => ['home', 'offers', 'vip', 'share', 'profile'].includes(String(route.name || '')))
const maintenanceEnabled = computed(() => runtimeConfig.value.maintenance.enabled)
</script>

<template>
  <MaintenanceScreen v-if="maintenanceEnabled" />
  <template v-else>
    <RouterView />
    <BottomNav v-if="showBottomNav" />
    <BackgroundMusic />
    <SupportContact />
    <SiteOperationsLayer />
  </template>
</template>
