<script setup lang="ts">
import { computed, reactive } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { moduleConfigs } from "@/data/moduleCatalog";

const route = useRoute();
const { t } = useI18n();
const filterModel = reactive<Record<string, string>>({});

const moduleKey = computed(() => String(route.meta.moduleKey || ""));
const config = computed(() => moduleConfigs[moduleKey.value]);
const dataColumns = computed(() =>
  (config.value?.columns || []).filter(
    (column) => column.key !== "actions" && column.key !== "id",
  ),
);
const queryFields = computed(() =>
  dataColumns.value.filter(
    (column) => column.key !== "updatedAt" && column.key !== "createdAt",
  ),
);
const rows = computed(() => []);

function resetFilters() {
  Object.keys(filterModel).forEach((key) => {
    filterModel[key as keyof typeof filterModel] = "";
  });
}
</script>

<template>
  <div v-if="config" class="page-stack page-stack--search-table">
    <el-card class="section-card" shadow="never">
      <template #header>
        <h3 class="section-title">{{ config.title }}</h3>
      </template>
      <el-form
        class="module-query-form"
        :model="filterModel"
        inline
        label-position="left"
        label-width="auto"
      >
        <el-form-item
          v-for="field in queryFields"
          :key="field.key"
          :label="field.label"
        >
          <el-input
            v-model="filterModel[field.key]"
            clearable
            :placeholder="field.label"
            style="width: 180px"
          />
        </el-form-item>
        <el-form-item class="module-query-form__actions">
          <el-button type="primary">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="section-card" shadow="never">
      <template #header>
        <div class="section-header">
          <h3 class="section-title">{{ config.title }}</h3>
        </div>
      </template>
      <el-table :data="rows" border>
        <el-table-column
          type="index"
          :label="t('common.index')"
          width="70"
          align="center"
        />
        <el-table-column
          v-for="column in dataColumns"
          :key="column.key"
          :prop="column.key"
          :label="column.label"
          :align="column.align || 'left'"
          min-width="130"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <el-switch
              v-if="column.key.toLowerCase().includes('status')"
              :model-value="
                row[column.key] === true ||
                row[column.key] === 'enabled' ||
                row[column.key] === 'active'
              "
              inline-prompt
              active-text="启"
              inactive-text="停"
            />
            <span v-else>{{ row[column.key] || "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="150" align="center">
          <template #default>
            <div class="table-actions">
              <el-button link type="primary">查看</el-button>
              <el-button link type="primary">编辑</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-row">
        <el-pagination
          background
          layout="prev, pager, next, sizes, total"
          :total="0"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.module-query-form {
  display: flex;
  flex-wrap: wrap;
  column-gap: 16px;
  row-gap: 6px;
  align-items: flex-end;
}

.module-query-form :deep(.el-form-item) {
  margin-bottom: 10px;
  margin-right: 0;
}

.module-query-form :deep(.el-form-item__label) {
  padding-right: 8px;
}

.module-query-form__actions {
  min-width: 160px;
}

.table-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
