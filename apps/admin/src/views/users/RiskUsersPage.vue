<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import ReasonConfirmDialog from "@/components/users/ReasonConfirmDialog.vue";
import UserDetailDialog from "@/components/users/UserDetailDialog.vue";
import {
  fetchAdminRiskUser,
  fetchAdminRiskUsers,
  updateAdminUserRisk,
  updateAdminUserStatus,
  type RiskStatus,
  type RiskUser,
  type RiskUserDetail,
  type UserStatus,
} from "@/api/adminUsers";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { useAdminCenterText } from "@/composables/useAdminCenterText";
const { can, loadPermissions } = useAdminPermissions();
const tr = useAdminCenterText("userCenter");
const loading = ref(false),
  actionLoading = ref(false),
  rows = ref<RiskUser[]>([]),
  total = ref(0),
  page = ref(1),
  pageSize = ref(20),
  detail = ref<RiskUserDetail | null>(null),
  detailVisible = ref(false),
  activeDetailTab = ref("profile"),
  detailUserVisible = ref(false),
  detailUserID = ref(0),
  confirmVisible = ref(false);
const pending = ref<{
  row: RiskUser;
  action: "risk" | "status";
  value: string;
} | null>(null);
const query = reactive({
  userId: "",
  username: "",
  riskLevel: "",
  riskStatus: "",
  riskTag: "",
  country: "",
  registrationIP: "",
  lastLoginIP: "",
  assignedReviewer: "",
  createdRange: [] as string[],
  updatedRange: [] as string[],
});
function params() {
  const r: Record<string, string | number> = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const [k, v] of Object.entries(query))
    if (typeof v === "string" && v.trim()) r[k] = v.trim();
  if (query.createdRange[0])
    r.createdFrom = selectedTimeToBackend(query.createdRange[0]);
  if (query.createdRange[1])
    r.createdTo = selectedTimeToBackend(query.createdRange[1], true);
  if (query.updatedRange[0])
    r.updatedFrom = selectedTimeToBackend(query.updatedRange[0]);
  if (query.updatedRange[1])
    r.updatedTo = selectedTimeToBackend(query.updatedRange[1], true);
  return r;
}
async function load() {
  loading.value = true;
  try {
    const d = await fetchAdminRiskUsers(params());
    rows.value = d.list;
    total.value = d.total;
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
    riskLevel: "",
    riskStatus: "",
    riskTag: "",
    country: "",
    registrationIP: "",
    lastLoginIP: "",
    assignedReviewer: "",
    createdRange: [],
    updatedRange: [],
  });
  search();
}
function indexMethod(i: number) {
  return (page.value - 1) * pageSize.value + i + 1;
}
async function view(row: RiskUser) {
  try {
    detail.value = await fetchAdminRiskUser(row.userID);
    activeDetailTab.value = "profile";
    detailVisible.value = true;
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : tr("Request failed"));
  }
}
function openUserDetail(userID: number) {
  detailUserID.value = userID;
  detailUserVisible.value = true;
}
function ask(row: RiskUser, action: "risk" | "status", value: string) {
  pending.value = { row, action, value };
  confirmVisible.value = true;
}
async function confirm(reason: string) {
  if (!pending.value) return;
  const { row, action, value } = pending.value;
  actionLoading.value = true;
  try {
    if (action === "status")
      await updateAdminUserStatus(row.userID, value as UserStatus, reason);
    else
      await updateAdminUserRisk(row.userID, {
        riskLevel: value === "CLEARED" ? "LOW" : row.riskLevel,
        riskStatus: value as RiskStatus,
        assignedReviewer: row.assignedReviewer,
        reason,
      });
    confirmVisible.value = false;
    await load();
    ElMessage.success(tr("Updated"));
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : tr("Request failed"));
  } finally {
    actionLoading.value = false;
  }
}
onMounted(async () => {
  await loadPermissions();
  await load();
});
</script>
<template>
  <div class="page-stack page-stack--search-table">
    <UserCenterCard :title="tr('Risk User Search')"
      ><el-form inline class="user-search-form"
        ><el-form-item :label="tr('Player Number')"
          ><el-input v-model="query.userId" /></el-form-item
        ><el-form-item :label="tr('Username')"
          ><el-input v-model="query.username" /></el-form-item
        ><el-form-item :label="tr('Risk Level')"
          ><el-select v-model="query.riskLevel" clearable
            ><el-option
              v-for="v in ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL']"
              :key="v"
              :label="tr(v)"
              :value="v" /></el-select></el-form-item
        ><el-form-item :label="tr('Risk Status')"
          ><el-select v-model="query.riskStatus" clearable
            ><el-option
              v-for="v in [
                'OBSERVED',
                'UNDER_REVIEW',
                'RESTRICTED',
                'CLEARED',
                'BLOCKED',
              ]"
              :key="v"
              :label="tr(v)"
              :value="v" /></el-select></el-form-item
        ><el-form-item :label="tr('Risk Tag')"
          ><el-input v-model="query.riskTag" /></el-form-item
        ><el-form-item :label="tr('Country')"
          ><el-input v-model="query.country" /></el-form-item
        ><el-form-item :label="tr('Registration IP')"
          ><el-input v-model="query.registrationIP" /></el-form-item
        ><el-form-item :label="tr('Last Login IP')"
          ><el-input v-model="query.lastLoginIP" /></el-form-item
        ><el-form-item :label="tr('Assigned Reviewer')"
          ><el-input v-model="query.assignedReviewer" /></el-form-item
        ><el-form-item :label="tr('Created Time')"
          ><el-date-picker
            v-model="query.createdRange"
            type="daterange"
            value-format="YYYY-MM-DD" /></el-form-item
        ><el-form-item :label="tr('Updated Time')"
          ><el-date-picker
            v-model="query.updatedRange"
            type="daterange"
            value-format="YYYY-MM-DD" /></el-form-item
        ><el-form-item
          ><el-button type="primary" @click="search">{{
            tr("Search")
          }}</el-button
          ><el-button @click="reset">{{ tr("Reset") }}</el-button></el-form-item
        ></el-form
      ></UserCenterCard
    >
    <UserCenterCard :title="tr('Risk Users')"
      ><el-table
        v-loading="loading"
        :data="rows"
        row-key="id"
        border
        :empty-text="tr('No data')"
        ><el-table-column
          type="index"
          :index="indexMethod"
          :label="tr('Index')"
          fixed="left"
          width="72"
        /><el-table-column
          prop="userNo"
          :label="tr('Player Number')"
          width="150"
        /><el-table-column
          prop="username"
          :label="tr('Username')"
          width="140"
        /><el-table-column
          prop="riskLevel"
          :label="tr('Risk Level')"
          width="110"
          ><template #default="{ row }"
            ><UserStatusTag
              :value="row.riskLevel"
              :label="tr(row.riskLevel)" /></template></el-table-column
        ><el-table-column
          prop="riskStatus"
          :label="tr('Risk Status')"
          width="140"
          ><template #default="{ row }"
            ><UserStatusTag
              :value="row.riskStatus"
              :label="tr(row.riskStatus)" /></template></el-table-column
        ><el-table-column
          prop="riskTags"
          :label="tr('Risk Tags')"
          width="180"
        /><el-table-column
          prop="triggerCount"
          :label="tr('Trigger Count')"
          width="110"
        /><el-table-column
          prop="registeredIP"
          :label="tr('Registration IP')"
          width="140"
        /><el-table-column
          prop="lastLoginIP"
          :label="tr('Last Login IP')"
          width="140"
        /><el-table-column
          prop="associatedAccounts"
          :label="tr('Associated Accounts')"
          width="150"
        /><el-table-column
          prop="assignedReviewer"
          :label="tr('Assigned Reviewer')"
          width="150"
        /><el-table-column :label="tr('Latest Risk Time')" width="170"
          ><template #default="{ row }">{{
            formatBackendTime(row.latestRiskAt)
          }}</template></el-table-column
        ><el-table-column :label="tr('Updated Time')" width="170"
          ><template #default="{ row }">{{
            formatBackendTime(row.updatedAt)
          }}</template></el-table-column
        ><el-table-column :label="tr('Actions')" fixed="right" width="270"
          ><template #default="{ row }"
            ><div class="table-actions">
              <el-button link type="primary" @click="view(row)">{{
                tr("Risk Detail")
              }}</el-button
              ><el-button
                link
                type="primary"
                @click="openUserDetail(row.userID)"
                >{{ tr("View User") }}</el-button
              ><el-dropdown v-if="can('user:risk:update')"
                ><el-button link type="primary">{{ tr("Manage") }}</el-button
                ><template #dropdown
                  ><el-dropdown-menu
                    ><el-dropdown-item
                      @click="ask(row, 'risk', 'UNDER_REVIEW')"
                      >{{ tr("Mark Under Review") }}</el-dropdown-item
                    ><el-dropdown-item
                      @click="ask(row, 'risk', 'RESTRICTED')"
                      >{{ tr("Restrict Account") }}</el-dropdown-item
                    ><el-dropdown-item @click="ask(row, 'risk', 'CLEARED')">{{
                      tr("Clear Risk")
                    }}</el-dropdown-item
                    ><el-dropdown-item @click="ask(row, 'risk', 'BLOCKED')">{{
                      tr("Block Risk")
                    }}</el-dropdown-item
                    ><el-dropdown-item
                      v-if="can('user:status:update')"
                      divided
                      @click="ask(row, 'status', 'LOCKED')"
                      >{{ tr("Lock Account") }}</el-dropdown-item
                    ></el-dropdown-menu
                  ></template
                ></el-dropdown
              >
            </div></template
          ></el-table-column
        ></el-table
      >
      <div class="pagination-row">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          background
          layout="prev,pager,next,sizes,total"
          :total="total"
          @current-change="load"
          @size-change="search"
        /></div
    ></UserCenterCard>
    <el-dialog
      v-model="detailVisible"
      :title="tr('Risk Detail')"
      width="min(980px, 92vw)"
      destroy-on-close
      append-to-body
    >
      <el-tabs v-if="detail" v-model="activeDetailTab" type="card">
        <el-tab-pane :label="tr('Current Risk Profile')" name="profile">
          <el-card shadow="never">
            <el-descriptions :column="2" border
              ><el-descriptions-item :label="tr('Risk Level')"
                ><UserStatusTag
                  :value="detail.profile.riskLevel"
                  :label="tr(detail.profile.riskLevel)" /></el-descriptions-item
              ><el-descriptions-item :label="tr('Risk Status')"
                ><UserStatusTag
                  :value="detail.profile.riskStatus"
                  :label="
                    tr(detail.profile.riskStatus)
                  " /></el-descriptions-item
              ><el-descriptions-item :label="tr('Risk Tags')">{{
                detail.profile.riskTags || "-"
              }}</el-descriptions-item
              ><el-descriptions-item :label="tr('Assigned Reviewer')">{{
                detail.profile.assignedReviewer || "-"
              }}</el-descriptions-item
              ><el-descriptions-item :label="tr('Related IPs')" :span="2"
                >{{ detail.profile.registeredIP || "-" }} /
                {{ detail.profile.lastLoginIP || "-" }}</el-descriptions-item
              ><el-descriptions-item :label="tr('Latest Reason')" :span="2">{{
                detail.profile.latestReason || "-"
              }}</el-descriptions-item></el-descriptions
            >
          </el-card>
        </el-tab-pane>
        <el-tab-pane :label="tr('Trigger and Status History')" name="history">
          <el-card shadow="never">
            <el-table
              :data="detail.events"
              border
              :empty-text="tr('No risk events')"
              ><el-table-column
                prop="eventType"
                :label="tr('Event')"
              /><el-table-column prop="riskLevel" :label="tr('Level')"
                ><template #default="{ row }"
                  ><UserStatusTag
                    :value="row.riskLevel"
                    :label="tr(row.riskLevel)" /></template></el-table-column
              ><el-table-column
                prop="description"
                :label="tr('Description')"
              /><el-table-column
                prop="operatorUsername"
                :label="tr('Operator')"
              /><el-table-column :label="tr('Time')"
                ><template #default="{ row }">{{
                  formatBackendTime(row.createdAt)
                }}</template></el-table-column
              ></el-table
            >
          </el-card>
        </el-tab-pane>
        <el-tab-pane :label="tr('Related Devices')" name="devices">
          <el-card shadow="never">
            <el-table
              :data="detail.devices"
              border
              :empty-text="tr('No related devices')"
              ><el-table-column prop="platform"
                :label="tr('Platform')" /><el-table-column
                prop="lastIP"
                :label="tr('Last IP')" /><el-table-column
                prop="riskStatusText"
                :label="tr('Risk Status')"
                ><template #default="{ row }"
                  ><UserStatusTag
                    :value="row.riskStatusText"
                    :label="
                      tr(row.riskStatusText)
                    " /></template></el-table-column
            ></el-table>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
    <UserDetailDialog v-model="detailUserVisible" :user-id="detailUserID" />
    ><ReasonConfirmDialog
      v-model="confirmVisible"
      :title="tr('Risk Account Action')"
      :message="
        tr(
          'High-risk actions require a reason and are recorded in the immutable audit log.',
        )
      "
      :loading="actionLoading"
      @confirm="confirm"
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
.risk-json {
  white-space: pre-wrap;
  color: var(--admin-text-secondary);
}
</style>
