<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import UserDetailDialog from "@/components/users/UserDetailDialog.vue";
import { fetchAdminLoginRecords, type LoginRecord } from "@/api/adminUsers";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { useAdminCenterText } from "@/composables/useAdminCenterText";

const route = useRoute();
const router = useRouter();
const tr = useAdminCenterText("userCenter");
const loading = ref(false);
const rows = ref<LoginRecord[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const detail = ref<LoginRecord | null>(null);
const detailUserVisible = ref(false);
const detailUserID = ref(0);
const query = reactive({
  userId: String(route.query.userId || ""),
  username: "",
  loginResult: "",
  clientIP: "",
  country: "",
  deviceID: "",
  platform: "",
  browser: "",
  os: "",
  loginRange: [] as string[],
});
function buildParams() {
  const result: Record<string, string | number> = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const [k, v] of Object.entries(query))
    if (typeof v === "string" && v.trim()) result[k] = v.trim();
  if (query.loginRange[0])
    result.loginFrom = selectedTimeToBackend(query.loginRange[0]);
  if (query.loginRange[1])
    result.loginTo = selectedTimeToBackend(query.loginRange[1], true);
  return result;
}
async function load() {
  loading.value = true;
  try {
    const data = await fetchAdminLoginRecords(buildParams());
    rows.value = data.list;
    total.value = data.total;
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : tr("Request failed"));
  } finally {
    loading.value = false;
  }
}
function search() {
  page.value = 1;
  load();
}
function reset() {
  Object.assign(query, {
    userId: "",
    username: "",
    loginResult: "",
    clientIP: "",
    country: "",
    deviceID: "",
    platform: "",
    browser: "",
    os: "",
    loginRange: [],
  });
  search();
}
function indexMethod(index: number) {
  return (page.value - 1) * pageSize.value + index + 1;
}
function userPhoneText(row: Pick<LoginRecord, "phoneCountryCode" | "phone" | "username">) {
  return (
    [row.phoneCountryCode, row.phone].filter(Boolean).join(" ") ||
    row.username ||
    "-"
  );
}
function openUserDetail(userID: number) {
  detailUserID.value = userID;
  detailUserVisible.value = true;
}
function detailLabel(key: string) {
  return tr(
    key
      .replace(/([a-z])([A-Z])/g, "$1 $2")
      .replace(/^./, (part) => part.toUpperCase()),
  );
}
onMounted(load);
</script>

<template>
  <div class="page-stack page-stack--search-table">
    <UserCenterCard :title="tr('Login Record Search')"
      ><el-form :model="query" inline class="user-search-form">
        <el-form-item :label="tr('Player Number')"
          ><el-input v-model="query.userId" clearable /></el-form-item
        ><el-form-item :label="tr('Mobile Number')"
          ><el-input v-model="query.username" clearable @keyup.enter="search"
        /></el-form-item>
        <el-form-item :label="tr('Login Result')"
          ><el-select v-model="query.loginResult" clearable
            ><el-option :label="tr('SUCCESS')" value="SUCCESS" /><el-option
              :label="tr('FAILED')"
              value="FAILED" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('IP Address')"
          ><el-input v-model="query.clientIP" clearable /></el-form-item
        ><el-form-item :label="tr('Country')"
          ><el-input v-model="query.country" clearable
        /></el-form-item>
        <el-form-item :label="tr('Platform')"
          ><el-input v-model="query.platform" clearable
        /></el-form-item>
        <el-form-item :label="tr('Browser')"
          ><el-input v-model="query.browser" clearable /></el-form-item
        ><el-form-item :label="tr('Operating System')"
          ><el-input v-model="query.os" clearable
        /></el-form-item>
        <el-form-item :label="tr('Login Time')"
          ><el-date-picker
            v-model="query.loginRange"
            type="daterange"
            value-format="YYYY-MM-DD"
        /></el-form-item>
        <el-form-item
          ><el-button type="primary" :loading="loading" @click="search">{{
            tr("Search")
          }}</el-button
          ><el-button @click="reset">{{ tr("Reset") }}</el-button></el-form-item
        >
      </el-form></UserCenterCard
    >
    <UserCenterCard :title="tr('Login Records')"
      ><el-table
        v-loading="loading"
        :data="rows"
        row-key="id"
        border
        :empty-text="tr('No data')"
      >
        <el-table-column
          type="index"
          :index="indexMethod"
          :label="tr('Index')"
          fixed="left"
          width="72"
          align="center"
        /><el-table-column
          prop="userNo"
          :label="tr('Player Number')"
          width="150"
        /><el-table-column
          prop="username"
          :label="tr('Mobile Number')"
          width="130"
          ><template #default="{ row }">{{
            userPhoneText(row)
          }}</template></el-table-column
        >
        <el-table-column
          prop="loginResult"
          :label="tr('Login Result')"
          width="120"
          ><template #default="{ row }"
            ><UserStatusTag
              :value="row.loginResult"
              :label="tr(row.loginResult)" /></template></el-table-column
        >
        <el-table-column
          prop="clientIP"
          :label="tr('IP Address')"
          width="140"
        /><el-table-column :label="tr('IP Country / Region')" width="170"
          ><template #default="{ row }">{{
            [row.country, row.region].filter(Boolean).join(" / ") || "-"
          }}</template></el-table-column
        >
        <el-table-column
          prop="deviceType"
          :label="tr('Device Type')"
          width="120"
        /><el-table-column
          prop="platform"
          :label="tr('Platform')"
          width="110"
        /><el-table-column
          prop="os"
          :label="tr('Operating System')"
          width="150"
        /><el-table-column :label="tr('Browser / App Version')" width="190"
          ><template #default="{ row }"
            >{{ row.browser || "-" }} / {{ row.appVersion || "-" }}</template
          ></el-table-column
        >
        <el-table-column :label="tr('Login Time')" width="170"
          ><template #default="{ row }">{{
            formatBackendTime(row.loginAt)
          }}</template></el-table-column
        >
        <el-table-column
          :label="tr('Actions')"
          fixed="right"
          width="220"
          align="center"
          ><template #default="{ row }"
            ><div class="table-actions">
              <el-button link type="primary" @click="detail = row">{{
                tr("View Details")
              }}</el-button
              ><el-button
                link
                type="primary"
                @click="openUserDetail(row.userID)"
                >{{ tr("View User") }}</el-button
              ><el-button
                v-if="row.deviceID"
                link
                type="primary"
                @click="
                  router.push({
                    path: '/users/devices',
                    query: { deviceKeyword: row.deviceID },
                  })
                "
                >{{ tr("View Device") }}</el-button
              >
            </div></template
          ></el-table-column
        >
      </el-table>
      <div class="pagination-row">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          background
          layout="prev, pager, next, sizes, total"
          :total="total"
          @current-change="load"
          @size-change="search"
        /></div
    ></UserCenterCard>
    <el-dialog
      v-model="detail"
      :title="tr('Login Record Details')"
      width="620px"
      ><el-tabs v-if="detail" type="card" model-value="overview"
        ><el-tab-pane :label="tr('Overview')" name="overview"
          ><el-card shadow="never"
            ><el-descriptions :column="2" border
              ><el-descriptions-item
                v-for="(value, key) in detail"
                :key="key"
                :label="detailLabel(String(key))"
                ><UserStatusTag
                  v-if="key === 'loginResult'"
                  :value="String(value)"
                  :label="tr(String(value))"
                /><template v-else>{{
                  key === "loginAt"
                    ? formatBackendTime(String(value))
                    : value || "-"
                }}</template></el-descriptions-item
              ></el-descriptions
            ></el-card></el-tab-pane></el-tabs
      ></el-dialog
    >
    <UserDetailDialog
      v-model="detailUserVisible"
      :user-id="detailUserID"
    />
  </div>
</template>
<style scoped>
.user-search-form :deep(.el-form-item) {
  margin: 0 18px 14px 0;
}
.user-search-form :deep(.el-input),
.user-search-form :deep(.el-select) {
  width: 190px;
}
.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
