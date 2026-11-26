<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { CircleCheck, Warning } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import type { BrandOverview } from "@/api/adminSystem";

const props = defineProps<{ overview: BrandOverview; compact?: boolean }>();
const router = useRouter();
const { t } = useI18n();
const checklist = computed(() => props.overview.setupChecklist || []);
const next = computed(() =>
  checklist.value.find((item) => item.status !== "COMPLETED"),
);
</script>

<template>
  <section class="setup-progress">
    <header>
      <div>
        <span>{{ t("tenantExperience.setup.eyebrow") }}</span>
        <h2>{{ t("tenantExperience.setup.title") }}</h2>
        <p>
          {{
            t("tenantExperience.setup.summary", {
              completed: overview.setupCompleted || 0,
              total: overview.setupTotal || checklist.length,
            })
          }}
        </p>
      </div>
      <div class="progress-value">
        <strong>{{ overview.setupPercent || 0 }}%</strong
        ><el-progress
          :percentage="overview.setupPercent || 0"
          :show-text="false"
        />
      </div>
    </header>
    <div v-if="!compact" class="checklist">
      <button
        v-for="item in checklist"
        :key="item.key"
        type="button"
        @click="router.push(item.path)"
      >
        <el-icon :class="item.status.toLowerCase()"
          ><CircleCheck v-if="item.status === 'COMPLETED'" /><Warning v-else
        /></el-icon>
        <span>{{ t(`tenantExperience.setup.items.${item.key}`) }}</span>
        <el-tag
          :type="
            item.status === 'COMPLETED'
              ? 'success'
              : item.status === 'ERROR'
                ? 'danger'
                : 'warning'
          "
          size="small"
          effect="plain"
        >
          {{ t(`tenantExperience.setup.status.${item.status}`) }}
        </el-tag>
      </button>
    </div>
    <el-button v-if="next" type="primary" @click="router.push(next.path)">{{
      t("tenantExperience.setup.continue")
    }}</el-button>
  </section>
</template>

<style scoped>
.setup-progress {
  padding: 18px;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 16px;
}
header span {
  font-size: 12px;
  font-weight: 700;
  color: var(--el-color-primary);
}
h2 {
  margin: 4px 0 0;
  font-size: 19px;
  letter-spacing: 0;
}
p {
  margin: 6px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.progress-value {
  width: 190px;
  text-align: right;
}
.progress-value strong {
  display: block;
  margin-bottom: 7px;
  font-size: 22px;
}
.checklist {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  margin-bottom: 16px;
  background: var(--el-border-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
}
.checklist button {
  display: grid;
  grid-template-columns: 20px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  min-height: 44px;
  padding: 8px 10px;
  border: 0;
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
  text-align: left;
  cursor: pointer;
}
.checklist button:hover {
  background: var(--el-fill-color-light);
}
.completed {
  color: var(--el-color-success);
}
.incomplete,
.warning {
  color: var(--el-color-warning);
}
.error {
  color: var(--el-color-danger);
}
@media (max-width: 900px) {
  header {
    flex-direction: column;
  }
  .progress-value {
    width: 100%;
    text-align: left;
  }
  .checklist {
    grid-template-columns: 1fr;
  }
}
</style>
