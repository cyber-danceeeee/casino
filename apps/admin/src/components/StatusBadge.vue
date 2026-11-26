<script setup lang="ts">
import type { Tone } from '@/data/moduleCatalog'

withDefaults(defineProps<{
  tone?: Tone
  label: string
}>(), {
  tone: 'neutral',
})
</script>

<template>
  <span class="status-badge" :class="`is-${tone}`">{{ label }}</span>
</template>

