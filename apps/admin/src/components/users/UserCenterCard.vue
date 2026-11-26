<script setup lang="ts">defineProps<{ title: string }>()</script>
<template>
  <el-card class="section-card" shadow="never">
    <template #header><div class="section-header"><h3 class="section-title">{{ title }}</h3><slot name="header" /></div></template>
    <slot />
  </el-card>
</template>
