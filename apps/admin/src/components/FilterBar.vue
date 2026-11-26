<script setup lang="ts">
defineProps<{
  filters: string[]
}>()
</script>

<template>
  <section class="filter-bar" aria-label="Filters">
    <label v-for="filter in filters" :key="filter">
      <span>{{ filter }}</span>
      <input v-if="filter === '关键词'" placeholder="输入关键词" />
      <select v-else-if="filter === '状态'">
        <option>全部状态</option>
        <option>Enabled</option>
        <option>Pending</option>
        <option>Risk Review</option>
        <option>Failed</option>
      </select>
      <select v-else-if="filter === 'Provider'">
        <option>全部 Provider</option>
        <option>SoftAggregator</option>
        <option>Hub88</option>
        <option>QTech</option>
      </select>
      <input v-else type="date" />
    </label>

    <div class="filter-bar__actions">
      <button class="primary-button" type="button">查询</button>
      <button class="ghost-button" type="button">重置</button>
    </div>
  </section>
</template>

