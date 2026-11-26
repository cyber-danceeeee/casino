<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import UserDetailDialog from "@/components/users/UserDetailDialog.vue";
import ReasonConfirmDialog from "@/components/users/ReasonConfirmDialog.vue";
import {
  fetchAdminUsers,
  updateAdminUserRestrictions,
  updateAdminUserRisk,
  updateAdminUserStatus,
  type AdminUser,
  type RiskLevel,
  type RiskStatus,
  type UserStatus,
} from "@/api/adminUsers";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { useAdminCenterText } from "@/composables/useAdminCenterText";
import {
  fetchAdminOptions,
  optionList,
  type AdminOptionSource,
  type AdminOptions,
} from "@/api/adminOptions";

const router = useRouter();
const { can, loadPermissions } = useAdminPermissions();
const tr = useAdminCenterText("userCenter");
const loading = ref(false);
const actionLoading = ref(false);
const rows = ref<AdminUser[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const adminOptions = ref<AdminOptions>({});
const query = reactive({
  userId: "",
  username: "",
  email: "",
  mobile: "",
  country: "",
  currency: "",
  status: "",
  kycStatus: "",
  vipLevel: "",
  channel: "",
  registrationIP: "",
  lastLoginIP: "",
  riskStatus: "",
  registeredRange: [] as string[],
  lastLoginRange: [] as string[],
});
const confirmVisible = ref(false);
const detailVisible = ref(false);
const detailUserID = ref(0);
const pending = ref<{
  row: AdminUser;
  kind: "status" | "restriction" | "risk";
  value: string;
} | null>(null);

function params() {
  const result: Record<string, string | number> = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const [key, value] of Object.entries(query))
    if (typeof value === "string" && value.trim()) result[key] = value.trim();
  if (query.registeredRange[0])
    result.registeredFrom = selectedTimeToBackend(query.registeredRange[0]);
  if (query.registeredRange[1])
    result.registeredTo = selectedTimeToBackend(query.registeredRange[1], true);
  if (query.lastLoginRange[0])
    result.lastLoginFrom = selectedTimeToBackend(query.lastLoginRange[0]);
  if (query.lastLoginRange[1])
    result.lastLoginTo = selectedTimeToBackend(query.lastLoginRange[1], true);
  return result;
}

async function load() {
  loading.value = true;
  try {
    const data = await fetchAdminUsers(params());
    rows.value = data.list;
    total.value = data.total;
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
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
    email: "",
    mobile: "",
    country: "",
    currency: "",
    status: "",
    kycStatus: "",
    vipLevel: "",
    channel: "",
    registrationIP: "",
    lastLoginIP: "",
    riskStatus: "",
    registeredRange: [],
    lastLoginRange: [],
  });
  search();
}
function indexMethod(index: number) {
  return (page.value - 1) * pageSize.value + index + 1;
}
function userPhoneText(row: Pick<AdminUser, "phoneCountryCode" | "phone" | "username">) {
  return (
    [row.phoneCountryCode, row.phone].filter(Boolean).join(" ") ||
    row.username ||
    "-"
  );
}
function options(source: AdminOptionSource) {
  return optionList(adminOptions.value, source);
}
function openUserDetail(row: AdminUser) {
  detailUserID.value = row.id;
  detailVisible.value = true;
}
function ask(
  row: AdminUser,
  kind: "status" | "restriction" | "risk",
  value: string,
) {
  pending.value = { row, kind, value };
  confirmVisible.value = true;
}
function actionTitle() {
  return pending.value?.kind === "risk"
    ? tr("Update risk status")
    : pending.value?.kind === "restriction"
      ? tr("Update account restriction")
      : tr("Update account status");
}
function actionMessage() {
  return pending.value
    ? `${actionTitle()} · ${pending.value.row.userNo || pending.value.row.username} · ${tr("This action is audited.")}`
    : "";
}

async function confirmAction(reason: string) {
  if (!pending.value) return;
  const { row, kind, value } = pending.value;
  actionLoading.value = true;
  try {
    if (kind === "status")
      await updateAdminUserStatus(row.id, value as UserStatus, reason);
    if (kind === "restriction")
      await updateAdminUserRestrictions(row.id, {
        [value]: row[value as keyof AdminUser] !== 1,
        reason,
      });
    if (kind === "risk") {
      const riskStatus = value as RiskStatus;
      const riskLevel: RiskLevel =
        riskStatus === "CLEARED"
          ? "LOW"
          : row.riskLevel === "LOW"
            ? "HIGH"
            : row.riskLevel;
      await updateAdminUserRisk(row.id, {
        riskLevel,
        riskStatus,
        assignedReviewer: "",
        reason,
      });
    }
    confirmVisible.value = false;
    ElMessage.success(tr("Updated"));
    await load();
  } catch (error) {
    ElMessage.error(
      error instanceof Error ? error.message : tr("Request failed"),
    );
  } finally {
    actionLoading.value = false;
  }
}

onMounted(async () => {
  await Promise.all([
    loadPermissions(),
    fetchAdminOptions()
      .then((options) => {
        adminOptions.value = options;
      })
      .catch(() => {
        adminOptions.value = {};
      }),
  ]);
  await load();
});
</script>

<template>
  <div class="page-stack page-stack--search-table">
    <UserCenterCard :title="tr('User Search')">
      <el-form
        :model="query"
        inline
        label-position="left"
        class="user-search-form"
        @submit.prevent="search"
      >
        <el-form-item :label="tr('Player Number')"
          ><el-input v-model="query.userId" clearable
        /></el-form-item>
        <el-form-item :label="tr('Username')"
          ><el-input v-model="query.username" clearable @keyup.enter="search"
        /></el-form-item>
        <el-form-item :label="tr('Email')"
          ><el-input v-model="query.email" clearable @keyup.enter="search"
        /></el-form-item>
        <el-form-item :label="tr('Mobile')"
          ><el-input v-model="query.mobile" clearable @keyup.enter="search"
        /></el-form-item>
        <el-form-item :label="tr('Country')"
          ><el-select v-model="query.country" clearable filterable
            ><el-option
              v-for="item in options('countries')"
              :key="item.value"
              :label="item.label"
              :value="item.value" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('Currency')"
          ><el-select v-model="query.currency" clearable filterable
            ><el-option
              v-for="item in options('currencies')"
              :key="item.value"
              :label="item.label"
              :value="item.value" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('Account Status')"
          ><el-select v-model="query.status" clearable
            ><el-option
              v-for="value in [
                'ACTIVE',
                'DISABLED',
                'FROZEN',
                'LOCKED',
                'CLOSED',
              ]"
              :key="value"
              :label="tr(value)"
              :value="value" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('KYC Status')"
          ><el-select v-model="query.kycStatus" clearable
            ><el-option
              :label="tr('NOT_SUBMITTED')"
              value="NOT_SUBMITTED" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('VIP Level')"
          ><el-select v-model="query.vipLevel" clearable filterable
            ><el-option
              v-for="item in options('vipLevels')"
              :key="item.value"
              :label="item.label"
              :value="item.value" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('Registration Channel')"
          ><el-select v-model="query.channel" clearable filterable
            ><el-option
              v-for="item in options('registrationChannels')"
              :key="item.value"
              :label="item.label"
              :value="item.value" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('Registration IP')"
          ><el-input v-model="query.registrationIP" clearable
        /></el-form-item>
        <el-form-item :label="tr('Last Login IP')"
          ><el-input v-model="query.lastLoginIP" clearable
        /></el-form-item>
        <el-form-item :label="tr('Risk Status')"
          ><el-select v-model="query.riskStatus" clearable
            ><el-option
              v-for="value in [
                'OBSERVED',
                'UNDER_REVIEW',
                'RESTRICTED',
                'CLEARED',
                'BLOCKED',
              ]"
              :key="value"
              :label="tr(value)"
              :value="value" /></el-select
        ></el-form-item>
        <el-form-item :label="tr('Registration Time')"
          ><el-date-picker
            v-model="query.registeredRange"
            type="daterange"
            value-format="YYYY-MM-DD"
        /></el-form-item>
        <el-form-item :label="tr('Last Login Time')"
          ><el-date-picker
            v-model="query.lastLoginRange"
            type="daterange"
            value-format="YYYY-MM-DD"
        /></el-form-item>
        <el-form-item
          ><el-button type="primary" :loading="loading" @click="search">{{
            tr("Search")
          }}</el-button
          ><el-button @click="reset">{{ tr("Reset") }}</el-button></el-form-item
        >
      </el-form>
    </UserCenterCard>

    <UserCenterCard :title="tr('User List')">
      <el-table
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
        />
        <el-table-column
          prop="userNo"
          :label="tr('Player Number')"
          width="150"
          show-overflow-tooltip
        />
        <el-table-column
          prop="username"
          :label="tr('Mobile Number')"
          width="140"
          show-overflow-tooltip
          ><template #default="{ row }">{{
            userPhoneText(row)
          }}</template></el-table-column
        >
        <el-table-column
          :label="tr('Email')"
          width="200"
          show-overflow-tooltip
          ><template #default="{ row }"
            ><span>{{ row.email || "-" }}</span></template
          ></el-table-column
        >
        <el-table-column prop="countryCode" :label="tr('Country')" width="90" />
        <el-table-column prop="currency" :label="tr('Currency')" width="90" />
        <el-table-column
          prop="cashBalance"
          :label="tr('Available Balance')"
          width="150"
          align="right"
          ><template #default="{ row }">
            <span class="balance-value balance-value--available">{{
              row.cashBalance || "0.00"
            }}</span>
          </template></el-table-column
        >
        <el-table-column
          prop="frozenBalance"
          :label="tr('Frozen Balance')"
          width="130"
          align="right"
          ><template #default="{ row }">
            <span class="balance-value balance-value--frozen">{{
              row.frozenBalance || "0.00"
            }}</span>
          </template></el-table-column
        >
        <el-table-column
          prop="vipLevel"
          :label="tr('VIP Level')"
          width="95"
          align="center"
        />
        <el-table-column prop="kycStatus" :label="tr('KYC Status')" width="140"
          ><template #default="{ row }"
            ><UserStatusTag :value="row.kycStatus" :label="tr(row.kycStatus)" /></template
        ></el-table-column>
        <el-table-column
          prop="statusText"
          :label="tr('Account Status')"
          width="125"
          ><template #default="{ row }"
            ><UserStatusTag :value="row.statusText" :label="tr(row.statusText)" /></template
        ></el-table-column>
        <el-table-column
          prop="riskStatus"
          :label="tr('Risk Status')"
          width="130"
          ><template #default="{ row }"
            ><UserStatusTag :value="row.riskStatus" :label="tr(row.riskStatus)" /></template
        ></el-table-column>
        <el-table-column
          prop="registrationChannel"
          :label="tr('Registration Channel')"
          width="155"
        />
        <el-table-column
          prop="registeredIP"
          :label="tr('Registration IP')"
          width="140"
        />
        <el-table-column
          prop="lastLoginIP"
          :label="tr('Last Login IP')"
          width="140"
        />
        <el-table-column :label="tr('Registration Time')" width="170"
          ><template #default="{ row }">{{
            formatBackendTime(row.registeredAt)
          }}</template></el-table-column
        >
        <el-table-column :label="tr('Last Login Time')" width="170"
          ><template #default="{ row }">{{
            formatBackendTime(row.lastLoginAt)
          }}</template></el-table-column
        >
        <el-table-column
          :label="tr('Actions')"
          fixed="right"
          width="250"
          align="center"
        >
          <template #default="{ row }"
            ><div class="table-actions">
              <el-button
                link
                type="primary"
                @click="openUserDetail(row)"
                >{{ tr("View Details") }}</el-button
              >
              <el-dropdown trigger="click">
                <el-button link type="primary">{{ tr("More") }}</el-button>
                <template #dropdown
                  ><el-dropdown-menu>
                    <el-dropdown-item
                      @click="
                        router.push({
                          path: '/users/login-records',
                          query: { userId: row.id },
                        })
                      "
                      >{{ tr("View Login Records") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      @click="
                        router.push({
                          path: '/users/devices',
                          query: { userId: row.id },
                        })
                      "
                      >{{ tr("View Devices") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="can('user:status:update')"
                      divided
                      @click="
                        ask(
                          row,
                          'status',
                          row.statusText === 'ACTIVE' ? 'DISABLED' : 'ACTIVE',
                        )
                      "
                      >{{
                        tr(row.statusText === "ACTIVE" ? "Disable" : "Enable")
                      }}
                      {{ tr("Account") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="can('user:status:update')"
                      @click="
                        ask(
                          row,
                          'status',
                          row.statusText === 'FROZEN' ? 'ACTIVE' : 'FROZEN',
                        )
                      "
                      >{{
                        tr(row.statusText === "FROZEN" ? "Unfreeze" : "Freeze")
                      }}
                      {{ tr("Account") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="can('user:restriction:update')"
                      @click="ask(row, 'restriction', 'loginEnabled')"
                      >{{ tr(row.loginEnabled === 1 ? "Disable" : "Allow") }}
                      {{ tr("Login") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="can('user:restriction:update')"
                      @click="ask(row, 'restriction', 'gameAccessEnabled')"
                      >{{
                        tr(row.gameAccessEnabled === 1 ? "Disable" : "Allow")
                      }}
                      {{ tr("Game Access") }}</el-dropdown-item
                    >
                    <el-dropdown-item
                      v-if="can('user:risk:update')"
                      @click="
                        ask(
                          row,
                          'risk',
                          row.riskStatus === 'CLEARED'
                            ? 'UNDER_REVIEW'
                            : 'CLEARED',
                        )
                      "
                      >{{
                        row.riskStatus === "CLEARED"
                          ? tr("Add Risk Flag")
                          : tr("Clear Risk")
                      }}</el-dropdown-item
                    >
                  </el-dropdown-menu></template
                >
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-row">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          background
          layout="prev, pager, next, sizes, total"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </UserCenterCard>
    <ReasonConfirmDialog
      v-model="confirmVisible"
      :title="actionTitle()"
      :message="actionMessage()"
      :loading="actionLoading"
      @confirm="confirmAction"
    />
    <UserDetailDialog
      v-model="detailVisible"
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
.secondary-value {
  margin-top: 3px;
  color: var(--admin-text-muted);
}
.balance-value {
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.balance-value--available {
  color: #1f9d55;
}
.balance-value--frozen {
  color: #f56c6c;
}
.pagination-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
