<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Refresh, Search, Setting, } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";
import {
  batchUpdateGames,
  fetchCategories,
  fetchGames,
  fetchProviderOptionGameTypes,
  type GameCenterRecord,
  type ProviderGameTypeOption,
} from "@/api/adminGames";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { fetchAdminOptions, optionList, type AdminOption } from "@/api/adminOptions";
import { systemAPI, type BrandOverview } from "@/api/adminSystem";

const { t } = useI18n();
const router = useRouter();
const { can, loadPermissions } = useAdminPermissions();
const loading = ref(false);
const rows = ref<GameCenterRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const selected = ref<GameCenterRecord[]>([]);
const categories = ref<GameCenterRecord[]>([]);
const providerOptions = ref<AdminOption[]>([]);
const providerCategoryOptions = ref<ProviderGameTypeOption[]>([]);
const filters = reactive({ keyword: "", providerCode: "", providerCategory: "", maintenance: "" });
const categoryDialogVisible = ref(false);
const categorySaving = ref(false);
const categoryForm = reactive({ ids: [] as number[], reason: "" });
const brandOverview = ref<BrandOverview | null>(null);

const selectedLabel = computed(() =>
  selected.value
    .map(
      (row) =>
        `${row.name || "-"} (${row.gameCode || row.providerGameId || "-"})`,
    )
    .join("\n"),
);
function configurationValue(row:GameCenterRecord,key:string){return row[key];}

function categoryNames(row: GameCenterRecord) {
  return Array.isArray(row.platformCategories)
    ? row.platformCategories
        .map(
          (item: Record<string, unknown>) =>
            item.nameZh || item.nameEn || item.code,
        )
        .join(", ")
    : String(row.providerCategoryCode || "-");
}
async function load() {
  loading.value = true;
  try {
    const result = await fetchGames({
      page: page.value,
      pageSize: pageSize.value,
      gameName: filters.keyword || undefined,
      providerCode: filters.providerCode || undefined,
      providerCategory: filters.providerCategory || undefined,
      maintenance: filters.maintenance || undefined,
    });
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    rows.value = [];
    ElMessage.error(
      error instanceof Error
        ? error.message
        : t("siteExperience.common.loadFailed"),
    );
  } finally {
    loading.value = false;
  }
}
async function changeProvider(providerCode: string) {
  filters.providerCategory = "";
  providerCategoryOptions.value = [];
  if (!providerCode) return;
  try {
    providerCategoryOptions.value = (await fetchProviderOptionGameTypes(providerCode))
      .filter((item) => item.source === "providerCategory");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("siteExperience.common.loadFailed"));
  }
}
async function loadCategories() {
  try {
    categories.value =
      (await fetchCategories({ page: 1, pageSize: 200 })).list || [];
  } catch {
    categories.value = [];
  }
}
async function batch(
  field: "maintenance" | "isHot" | "isFeatured" | "isNew",
  value: boolean,
) {
  if (!selected.value.length) return;
  try {
    const prompt = await ElMessageBox.prompt(
      `${t(`siteExperience.gameOps.actions.${field}.${value ? "on" : "off"}`)}\n\n${selectedLabel.value}`,
      t("siteExperience.gameOps.confirmTitle"),
      {
        confirmButtonText: t("common.confirm"),
        cancelButtonText: t("common.cancel"),
        inputPlaceholder: t("siteExperience.fields.reason"),
        inputValidator: (reason) =>
          Boolean(reason.trim()) ||
          t("siteExperience.validation.reasonRequired"),
        type: "warning",
      },
    );
    const result = await batchUpdateGames({
      ids: selected.value.map((row) => row.id),
      [field]: value,
      reason: prompt.value.trim(),
    });
    ElMessage.success(
      t("siteExperience.gameOps.updated", {
        count: result.updatedCount || selected.value.length,
      }),
    );
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    if (error !== "cancel" && error !== "close")
      ElMessage.error(
        error instanceof Error
          ? error.message
          : t("siteExperience.common.actionFailed"),
      );
  }
}
async function changeMaintenance(row: GameCenterRecord, enabled: boolean) {
  try {
    await batchUpdateGames({ids:[row.id], maintenance:enabled, reason:"toggle_game_maintenance"});
    ElMessage.success(t("gameCenter.messages.savedImmediate"));
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch(error) {
    ElMessage.error(error instanceof Error ? error.message : t("siteExperience.common.actionFailed"));
  }
}
async function assignCategory() {
  if (!selected.value.length) return;
  categoryForm.ids = [];
  categoryForm.reason = "";
  categoryDialogVisible.value = true;
}
async function saveCategories() {
  if (!selected.value.length || !categoryForm.ids.length || !categoryForm.reason.trim()) return;
  categorySaving.value = true;
  try {
    await batchUpdateGames({
      ids: selected.value.map((row) => row.id),
      addCategoryIds: categoryForm.ids,
      reason: categoryForm.reason.trim(),
    });
    ElMessage.success(t("siteExperience.gameOps.categoryAssigned"));
    categoryDialogVisible.value = false;
    window.dispatchEvent(new Event("brand-config-refresh"));
    await load();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t("siteExperience.common.actionFailed"));
  } finally {
    categorySaving.value = false;
  }
}
function reset() {
  Object.assign(filters, { keyword: "", providerCode: "", providerCategory: "", maintenance: "" });
  providerCategoryOptions.value = [];
  page.value = 1;
  void load();
}
onMounted(async () => {
  await loadPermissions();
  const [options, overview] = await Promise.all([
    fetchAdminOptions().catch(() => ({})),
    systemAPI.brand.overview().catch(() => null),
    loadCategories(),
  ]);
  providerOptions.value = optionList(options, "providers");
  brandOverview.value = overview;
  await load();
});
</script>
<template>
  <div class="game-operations">
    <header class="page-header">
      <div>
        <p>{{ t("siteExperience.gameOps.eyebrow") }}</p>
        <h1>{{ t("siteExperience.gameOps.title") }}</h1>
        <span>{{ t("siteExperience.gameOps.description") }}</span>
      </div>
      <div class="header-actions">

        <el-button :icon="Refresh" @click="load">{{ t("common.refresh") }}</el-button>

      </div>
    </header>
    <section class="filters">
      <el-input
        v-model="filters.keyword"
        clearable
        :placeholder="t('siteExperience.gameOps.searchPlaceholder')"
        @keyup.enter="
          page = 1;
          load();
        "
      /><el-select
        v-model="filters.providerCode"
        clearable
        filterable
        :placeholder="t('siteExperience.fields.provider')"
        @change="changeProvider"
      ><el-option v-for="item in providerOptions" :key="item.value" :label="item.label" :value="item.value" /></el-select
      ><el-select
        v-model="filters.providerCategory"
        clearable
        filterable
        :disabled="!filters.providerCode"
        :placeholder="t('gameCenter.fields.providerCategory')"
      ><el-option v-for="item in providerCategoryOptions" :key="item.code" :label="`${item.code} (${item.gameCount})`" :value="item.code" /></el-select
      ><el-select
        v-model="filters.maintenance"
        clearable
        :placeholder="t('gameCenter.fields.maintenance')"
        ><el-option
          :label="t('gameCenter.fields.maintenanceOn')"
          value="true" /><el-option
          :label="t('gameCenter.fields.maintenanceOff')"
          value="false" /></el-select
      ><el-button
        type="primary"
        :icon="Search"
        @click="
          page = 1;
          load();
        "
        >{{ t("common.search") }}</el-button
      ><el-button @click="reset">{{ t("common.reset") }}</el-button>
    </section>
    <section v-if="selected.length" class="batch-bar">
      <strong>{{
        t("siteExperience.gameOps.selected", { count: selected.length })
      }}</strong
      ><el-button-group
        ><el-button
          v-if="can('game:update')"
          @click="batch('maintenance', true)"
          >{{
            t("siteExperience.gameOps.actions.maintenance.on")
          }}</el-button
        ><el-button
          v-if="can('game:update')"
          @click="batch('maintenance', false)"
          >{{
            t("siteExperience.gameOps.actions.maintenance.off")
          }}</el-button
        ><el-button v-if="can('game:update')" @click="batch('isHot', true)">{{
          t("siteExperience.gameOps.actions.isHot.on")
        }}</el-button
        ><el-button
          v-if="can('game:update')"
          @click="batch('isFeatured', true)"
          >{{ t("siteExperience.gameOps.actions.isFeatured.on") }}</el-button
        ><el-button v-if="can('game:update')" @click="batch('isNew', true)">{{
          t("siteExperience.gameOps.actions.isNew.on")
        }}</el-button
        ><el-button v-if="can('game:update')" @click="assignCategory">{{
          t("siteExperience.gameOps.assignCategory")
        }}</el-button></el-button-group
      >
    </section>
    <section class="table-shell">
      <el-table
        v-loading="loading"
        :data="rows"
        row-key="id"
        @selection-change="selected = $event"
        ><el-table-column type="selection" width="46" /><el-table-column
          :label="t('siteExperience.fields.cover')"
          width="76"
          ><template #default="{ row }"
            ><el-image
              class="cover"
              :src="
                String(
                  row.localThumbnailUrl ||
                    row.coverUrl ||
                    row.thumbnailUrl ||
                    '',
                )
              "
              fit="cover" /></template></el-table-column
        ><el-table-column
          prop="gameCode"
          :label="t('siteExperience.fields.gameCode')"
          min-width="165"
        /><el-table-column
          prop="name"
          :label="t('siteExperience.fields.gameName')"
          min-width="180"
        /><el-table-column
          prop="providerCode"
          :label="t('siteExperience.fields.provider')"
          width="130"
          ><template #default="{ row }"
            ><el-tag effect="plain">{{ row.providerCode }}</el-tag></template
          ></el-table-column
        ><el-table-column
          :label="t('siteExperience.fields.category')"
          min-width="170"
          ><template #default="{ row }">{{
            categoryNames(row)
          }}</template></el-table-column
        ><el-table-column
          :label="t('gameCenter.fields.maintenance')"
          width="95"
          ><template #default="{ row }"
            ><el-switch
              :model-value="Number(row.status) === 2"
              :disabled="!can('game:update')"
              @change="changeMaintenance(row, Boolean($event))" /></template></el-table-column
        ><el-table-column
          :label="t('siteExperience.fields.popular')"
          width="90"
          ><template #default="{ row }"
            ><el-switch
              :model-value="Boolean(configurationValue(row, 'isHot'))"
              disabled /></template></el-table-column
        ><el-table-column
          :label="t('siteExperience.fields.recommended')"
          width="115"
          ><template #default="{ row }"
            ><el-switch
              :model-value="Boolean(configurationValue(row, 'isFeatured'))"
              disabled /></template></el-table-column
        ><el-table-column :label="t('siteExperience.fields.new')" width="80"
          ><template #default="{ row }"
            ><el-switch
              :model-value="Boolean(configurationValue(row, 'isNew'))"
              disabled /></template></el-table-column
        ><el-table-column :label="t('siteExperience.fields.sort')" width="85"
          ><template #default="{ row }">{{
            configurationValue(row, "sortOrder") ?? "-"
          }}</template></el-table-column
        ><el-table-column fixed="right" width="80"
          ><template #default="{ row }"
            ><el-button
              circle
              :icon="Setting"
              @click="
                router.push({
                  path: '/games/list',
                  query: { gameCode: row.gameCode },
                })
              " /></template></el-table-column
        ><template #empty
          ><el-empty :description="t('siteExperience.empty.games')"
            ><el-button @click="reset">{{
              t("siteExperience.common.clearFilters")
            }}</el-button></el-empty
          ></template
        ></el-table
      ><el-pagination
        v-if="total"
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @change="load"
      />
    </section>
    <el-dialog v-model="categoryDialogVisible" :title="t('siteExperience.gameOps.assignCategory')" width="min(560px, 94vw)">
      <el-form label-position="top">
        <el-form-item :label="t('siteExperience.gameOps.categories')" required>
          <el-select v-model="categoryForm.ids" multiple filterable clearable class="full-width" :placeholder="t('gameCenter.placeholders.select')">
            <el-option v-for="category in categories" :key="category.id" :value="category.id" :label="String(category.nameZh || category.nameEn || category.code)" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('siteExperience.fields.reason')" required>
          <el-input v-model="categoryForm.reason" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="categorySaving" :disabled="!categoryForm.ids.length || !categoryForm.reason.trim()" @click="saveCategories">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<style scoped>
.game-operations {
  display: grid;
  gap: 16px;
}
.page-header,
.filters,
.batch-bar,
.table-shell {
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
}
.page-header {
  display: flex;
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
.page-header span {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.filters {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 150px auto auto;
  gap: 10px;
  padding: 14px;
}
.header-actions { display: flex; flex-wrap: wrap; align-items: center; justify-content: flex-end; gap: 10px; }
.full-width { width: 100%; }
.batch-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
}
.cover {
  width: 50px;
  height: 50px;
  border-radius: 4px;
  background: var(--el-fill-color);
}
.table-shell {
  overflow: hidden;
}
.table-shell :deep(.el-pagination) {
  justify-content: flex-end;
  padding: 14px;
}
@media (max-width: 900px) {
  .filters {
    grid-template-columns: 1fr;
  }
  .batch-bar {
    align-items: flex-start;
    flex-direction: column;
  }
  .batch-bar .el-button-group {
    display: flex;
    flex-wrap: wrap;
  }
}
</style>
