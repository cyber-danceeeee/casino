<script setup lang="ts">
import type { Metric } from '@/data/moduleCatalog'

defineProps<{
  metric: Metric
}>()
</script>

<template>
  <article class="metric-card" :class="`is-${metric.tone}`">
    <span>{{ metric.label }}</span>
    <strong>{{ metric.value }}</strong>
    <em>{{ metric.trend }}</em>
  </article>
</template>

