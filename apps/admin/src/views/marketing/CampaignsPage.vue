<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { Search } from "@element-plus/icons-vue";
import { useMarketingReadModel } from "@/composables/useMarketingReadModel";
import { formatBackendTime } from "@/utils/timezone";
import { statusKey, statusType } from "@/utils/presentation";
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { loading, items, errors, load } = useMarketingReadModel();
const page = ref(1);
const pageSize = ref(20);
const filters = reactive({ keyword: "", type: "", status: "" });
const pageMode = computed(() => String(route.meta.moduleKey || "campaigns"));
const filtered = computed(() =>
  items.value.filter(
    (item) =>
      (!filters.keyword ||
        `${item.name} ${item.code}`
          .toLowerCase()
          .includes(filters.keyword.toLowerCase())) &&
      (!filters.type || item.type === filters.type) &&
      (!filters.status || statusKey(item.status) === filters.status),
  ),
);
const paged = computed(() =>
  filtered.value.slice(
    (page.value - 1) * pageSize.value,
    page.value * pageSize.value,
  ),
);
function reset() {
  Object.assign(filters, { keyword: "", type: "", status: "" });
  page.value = 1;
}
onMounted(load);
</script>
<template>
  <div v-loading="loading" class="campaigns-page">
    <header class="page-header">
      <div>
        <p>{{ t("tenantExperience.marketing.eyebrow") }}</p>
        <h1>{{ t(`tenantExperience.marketing.pages.${pageMode}`) }}</h1>
        <span>{{
          t(`tenantExperience.marketing.pageDescriptions.${pageMode}`)
        }}</span>
      </div>
    </header>
    <el-alert
      v-if="errors.length"
      type="warning"
      :title="t('tenantExperience.dashboard.partialData')"
      :description="errors[0]"
      :closable="false"
    />
    <section class="filters">
      <el-input
        v-model="filters.keyword"
        clearable
        :prefix-icon="Search"
        :placeholder="t('tenantExperience.marketing.searchPlaceholder')"
      /><el-select
        v-model="filters.type"
        clearable
        :placeholder="t('tenantExperience.fields.type')"
        ><el-option
          v-for="type in [
            'PROMOTION',
            'COUPON',
            'VIP',
            'PUSH',
            'BANNER',
            'POPUP',
            'ANNOUNCEMENT',
          ]"
          :key="type"
          :label="t(`tenantExperience.campaignTypes.${type}`)"
          :value="type" /></el-select
      ><el-select
        v-model="filters.status"
        clearable
        :placeholder="t('tenantExperience.fields.status')"
        ><el-option
          v-for="status in [
            'ACTIVE',
            'DRAFT',
            'SCHEDULED',
            'PUBLISHED',
            'EXPIRED',
            'DISABLED',
          ]"
          :key="status"
          :label="t(`tenantExperience.status.${status}`)"
          :value="status" /></el-select
      ><el-button @click="reset">{{ t("common.reset") }}</el-button>
    </section>
    <section class="table-shell">
      <el-table :data="paged" row-key="rowKey"
        ><el-table-column
          prop="code"
          :label="t('tenantExperience.fields.campaignCode')"
          min-width="155"
          ><template #default="{ row }">{{
            row.code || "-"
          }}</template></el-table-column
        ><el-table-column
          prop="name"
          :label="t('tenantExperience.fields.campaignName')"
          min-width="210" /><el-table-column
          :label="t('tenantExperience.fields.type')"
          width="135"
          ><template #default="{ row }">{{
            t(`tenantExperience.campaignTypes.${row.type}`, row.type)
          }}</template></el-table-column
        ><el-table-column
          :label="t('tenantExperience.fields.status')"
          width="130"
          ><template #default="{ row }"
            ><el-tag :type="statusType(row.status)" effect="plain">{{
              t(`tenantExperience.status.${statusKey(row.status)}`, row.status)
            }}</el-tag></template
          ></el-table-column
        ><el-table-column
          :label="t('tenantExperience.fields.schedule')"
          min-width="190"
          ><template #default="{ row }"
            ><span>{{ formatBackendTime(row.startAt) }}</span
            ><small>{{ formatBackendTime(row.endAt) }}</small></template
          ></el-table-column
        ><el-table-column
          prop="audience"
          :label="t('tenantExperience.fields.audience')"
          width="130" /><el-table-column
          prop="placement"
          :label="t('tenantExperience.fields.placement')"
          width="140" /><el-table-column
          :label="t('tenantExperience.fields.publishVersion')"
          width="120"
          ><template #default="{ row }"
            >v{{ row.publishedVersion }}</template
          ></el-table-column
        ><el-table-column
          prop="updatedBy"
          :label="t('tenantExperience.fields.updatedBy')"
          width="130" /><el-table-column
          :label="t('tenantExperience.fields.updatedTime')"
          width="175"
          ><template #default="{ row }">{{
            formatBackendTime(row.updatedAt)
          }}</template></el-table-column
        ><el-table-column fixed="right" width="90"
          ><template #default="{ row }"
            ><el-button link type="primary" @click="router.push(row.path)">{{
              t("common.view")
            }}</el-button></template
          ></el-table-column
        ><template #empty
          ><el-empty
            :description="
              t(
                `tenantExperience.empty.${pageMode === 'missions' ? 'missions' : pageMode === 'referral' ? 'referral' : 'campaigns'}`,
              )
            " /></template></el-table
      ><el-pagination
        v-if="filtered.length"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="filtered.length"
        layout="total, sizes, prev, pager, next"
      />
    </section>
  </div>
</template>
<style scoped>
.campaigns-page {
  display: grid;
  gap: 16px;
}
.page-header,
.filters,
.table-shell {
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.page-header {
  padding: 20px;
}
.page-header p {
  margin: 0;
  color: var(--el-color-primary);
  font-size: 12px;
  font-weight: 700;
}
.page-header h1 {
  margin: 5px 0;
  font-size: 26px;
  letter-spacing: 0;
}
.page-header span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.filters {
  display: grid;
  grid-template-columns: minmax(240px, 2fr) repeat(2, minmax(150px, 1fr)) auto;
  gap: 10px;
  padding: 14px;
}
.table-shell {
  overflow: hidden;
}
.table-shell :deep(.el-pagination) {
  justify-content: flex-end;
  padding: 14px;
}
small {
  display: block;
  margin-top: 4px;
  color: var(--el-text-color-secondary);
}
@media (max-width: 760px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
