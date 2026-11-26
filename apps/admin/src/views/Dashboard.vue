<script setup lang="ts">
import { onMounted } from "vue";
import BrandDashboard from "@/views/brand/BrandDashboard.vue";
import PlatformDashboard from "@/views/platform/PlatformDashboard.vue";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
const { actor, loadPermissions } = useAdminPermissions();
onMounted(loadPermissions);
</script>
<template><BrandDashboard v-if="actor?.platform === false"/><PlatformDashboard v-else-if="actor"/><div v-else v-loading="true" class="dashboard-loading"/></template>
<style scoped>.dashboard-loading{min-height:360px}</style>
