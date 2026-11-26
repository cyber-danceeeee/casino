<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import StatusBadge from '@/components/StatusBadge.vue'
import type { TableColumn, TableRow } from '@/data/moduleCatalog'

const props = defineProps<{
  columns: TableColumn[]
  rows: TableRow[]
}>()
const { t } = useI18n()

const tableColumns = computed(() => {
  const businessColumns = props.columns.filter((column) => column.key !== 'actions')
  return [...businessColumns, { key: 'actions', label: '操作', align: 'center' as const }]
})
</script>

<template>
  <div class="data-table">
    <table>
      <thead>
        <tr>
          <th class="is-center">{{ t('common.index') }}</th>
          <th v-for="column in tableColumns" :key="column.key" :class="column.align ? `is-${column.align}` : undefined">
            {{ column.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, index) in rows" :key="row.id">
          <td class="is-center">{{ index + 1 }}</td>
          <td v-for="column in tableColumns" :key="column.key" :class="column.align ? `is-${column.align}` : undefined">
            <template v-if="column.key === 'actions'">
              <button class="table-action" type="button">详情</button>
              <button class="table-action" type="button">审计</button>
            </template>
            <template v-else-if="column.key === 'status' || column.key.endsWith('Status')">
              <StatusBadge :label="String(row[column.key] || '-')" :tone="row.statusTone" />
            </template>
            <template v-else>
              {{ row[column.key] || '-' }}
            </template>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
