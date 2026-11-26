<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useMarketingReadModel } from "@/composables/useMarketingReadModel";
import { formatBackendTime } from "@/utils/timezone";
import { statusKey, statusType } from "@/utils/presentation";
const { t } = useI18n();
const router = useRouter();
const { loading, items, errors, load } = useMarketingReadModel();
const metrics = computed(() => [
  {
    key: "active",
    value: items.value.filter((i) =>
      ["ACTIVE", "PUBLISHED"].includes(statusKey(i.status)),
    ).length,
  },
  {
    key: "scheduled",
    value: items.value.filter((i) => statusKey(i.status) === "SCHEDULED")
      .length,
  },
  {
    key: "draft",
    value: items.value.filter((i) => statusKey(i.status) === "DRAFT").length,
  },
  {
    key: "expired",
    value: items.value.filter((i) => statusKey(i.status) === "EXPIRED").length,
  },
  {
    key: "banners",
    value: items.value.filter(
      (i) =>
        i.type === "BANNER" &&
        ["ACTIVE", "PUBLISHED"].includes(statusKey(i.status)),
    ).length,
  },
  {
    key: "popups",
    value: items.value.filter(
      (i) =>
        i.type === "POPUP" &&
        ["ACTIVE", "PUBLISHED"].includes(statusKey(i.status)),
    ).length,
  },
  {
    key: "promotions",
    value: items.value.filter(
      (i) =>
        i.type === "PROMOTION" &&
        ["ACTIVE", "PUBLISHED"].includes(statusKey(i.status)),
    ).length,
  },
  {
    key: "coupons",
    value: items.value.filter(
      (i) =>
        i.type === "COUPON" &&
        ["ACTIVE", "PUBLISHED"].includes(statusKey(i.status)),
    ).length,
  },
]);
onMounted(load);
</script>
<template>
  <div v-loading="loading" class="marketing-dashboard">
    <header class="page-header">
      <div>
        <p>{{ t("siteExperience.marketing.eyebrow") }}</p>
        <h1>{{ t("siteExperience.marketing.title") }}</h1>
        <span>{{ t("siteExperience.marketing.description") }}</span>
      </div>
      <el-button type="primary" @click="router.push('/marketing/campaigns')">{{
        t("siteExperience.marketing.viewCampaigns")
      }}</el-button>
    </header>
    <el-alert
      v-if="errors.length"
      type="warning"
      :title="t('siteExperience.dashboard.partialData')"
      :description="errors[0]"
      :closable="false"
      show-icon
    />
    <section class="metric-grid">
      <article v-for="metric in metrics" :key="metric.key">
        <span>{{ t(`siteExperience.marketing.metrics.${metric.key}`) }}</span
        ><strong>{{ metric.value }}</strong>
      </article>
    </section>
    <section class="recent">
      <header>
        <h2>{{ t("siteExperience.marketing.recent") }}</h2>
      </header>
      <el-table :data="items.slice(0, 10)" row-key="rowKey"
        ><el-table-column
          prop="name"
          :label="t('siteExperience.fields.campaignName')"
          min-width="210"
        /><el-table-column
          :label="t('siteExperience.fields.type')"
          width="135"
          ><template #default="{ row }">{{
            t(`siteExperience.campaignTypes.${row.type}`, row.type)
          }}</template></el-table-column
        ><el-table-column
          :label="t('siteExperience.fields.status')"
          width="130"
          ><template #default="{ row }"
            ><el-tag :type="statusType(row.status)" effect="plain">{{
              t(`siteExperience.status.${statusKey(row.status)}`, row.status)
            }}</el-tag></template
          ></el-table-column
        ><el-table-column
          :label="t('siteExperience.fields.updatedTime')"
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
          ><el-empty :description="t('siteExperience.empty.campaigns')"
            ><el-button
              type="primary"
              @click="router.push('/operations/activities')"
              >{{ t("siteExperience.marketing.createPromotion") }}</el-button
            ></el-empty
          ></template
        ></el-table
      >
    </section>
  </div>
</template>
<style scoped>
.marketing-dashboard {
  display: grid;
  gap: 16px;
}
.page-header,
.metric-grid,
.recent {
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
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
.page-header span,
.metric-grid span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}
.metric-grid article {
  display: grid;
  gap: 7px;
  padding: 16px;
  border-right: 1px solid var(--el-border-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.metric-grid strong {
  font-size: 25px;
}
.recent {
  padding: 17px;
}
.recent > header {
  margin-bottom: 12px;
}
.recent h2 {
  margin: 0;
  font-size: 17px;
}
@media (max-width: 760px) {
  .page-header {
    flex-direction: column;
  }
  .metric-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
