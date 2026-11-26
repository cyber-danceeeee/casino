<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import ReasonConfirmDialog from "@/components/users/ReasonConfirmDialog.vue";
import UserDetailDialog from "@/components/users/UserDetailDialog.vue";
import {
  fetchAdminDevices,
  fetchAdminPushDevices,
  fetchAdminDevice,
  fetchAdminDeviceUsers,
  updateAdminDevice,
  type AdminUser,
  type PushDevice,
  type UserDevice,
} from "@/api/adminUsers";
import { formatBackendTime, selectedTimeToBackend } from "@/utils/timezone";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { useAdminCenterText } from "@/composables/useAdminCenterText";
const route = useRoute();
const { can, loadPermissions } = useAdminPermissions();
const tr = useAdminCenterText("userCenter");
const loading = ref(false),
  actionLoading = ref(false),
  rows = ref<UserDevice[]>([]),
  pushRows = ref<PushDevice[]>([]),
  total = ref(0),
  page = ref(1),
  pageSize = ref(20),
  detailVisible = ref(false),
  pushDetailVisible = ref(false),
  activeDetailTab = ref("overview"),
  detailUserVisible = ref(false),
  detailUserID = ref(0),
  current = ref<UserDevice | null>(null),
  currentPush = ref<PushDevice | null>(null),
  associated = ref<AdminUser[]>([]),
  confirmVisible = ref(false);
const deviceSource = ref<"login" | "fcm">("login");
const pending = ref<{
  row: UserDevice;
  payload: {
    status?: "ACTIVE" | "BLOCKED";
    trustStatus?: "NORMAL" | "TRUSTED";
  };
} | null>(null);
const query = reactive({
  userId: String(route.query.userId || ""),
  username: "",
  deviceKeyword: String(route.query.deviceKeyword || ""),
  fingerprint: "",
  platform: "",
  status: "",
  riskStatus: "",
  pushStatus: "",
  networkType: "",
  notificationPermission: "",
  firstRange: [] as string[],
  lastRange: [] as string[],
});
function params() {
  const r: Record<string, string | number> = {
    page: page.value,
    pageSize: pageSize.value,
  };
  for (const [k, v] of Object.entries(query))
    if (typeof v === "string" && v.trim()) r[k] = v.trim();
  if (query.firstRange[0])
    r.firstSeenFrom = selectedTimeToBackend(query.firstRange[0]);
  if (query.firstRange[1])
    r.firstSeenTo = selectedTimeToBackend(query.firstRange[1], true);
  if (query.lastRange[0])
    r.lastSeenFrom = selectedTimeToBackend(query.lastRange[0]);
  if (query.lastRange[1])
    r.lastSeenTo = selectedTimeToBackend(query.lastRange[1], true);
  return r;
}
async function load() {
  loading.value = true;
  try {
    const d =
      deviceSource.value === "fcm"
        ? await fetchAdminPushDevices(params())
        : await fetchAdminDevices(params());
    if (deviceSource.value === "fcm") pushRows.value = d.list as PushDevice[];
    else rows.value = d.list as UserDevice[];
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
    deviceKeyword: "",
    fingerprint: "",
    platform: "",
    status: "",
    riskStatus: "",
    pushStatus: "",
    networkType: "",
    notificationPermission: "",
    firstRange: [],
    lastRange: [],
  });
  search();
}
function switchDeviceSource() {
  page.value = 1;
  void load();
}
function indexMethod(i: number) {
  return (page.value - 1) * pageSize.value + i + 1;
}
async function view(row: UserDevice) {
  detailVisible.value = true;
  activeDetailTab.value = "overview";
  current.value = await fetchAdminDevice(row.id);
  associated.value = await fetchAdminDeviceUsers(row.id);
}
function viewPush(row: PushDevice) {
  currentPush.value = row;
  pushDetailVisible.value = true;
}
function networkTagType(value: string) {
  if (value === "WIFI" || value === "ETHERNET") return "success";
  if (value === "CELLULAR") return "primary";
  if (value === "VPN") return "warning";
  if (value === "OFFLINE") return "danger";
  return "info";
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
function ask(
  row: UserDevice,
  payload: {
    status?: "ACTIVE" | "BLOCKED";
    trustStatus?: "NORMAL" | "TRUSTED";
  },
) {
  pending.value = { row, payload };
  confirmVisible.value = true;
}
async function confirm(reason: string) {
  if (!pending.value) return;
  actionLoading.value = true;
  try {
    await updateAdminDevice(pending.value.row.id, {
      ...pending.value.payload,
      reason,
    });
    confirmVisible.value = false;
    ElMessage.success(tr("Updated"));
    await load();
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
    <UserCenterCard :title="tr('Device Search')"
      ><el-tabs
        v-model="deviceSource"
        type="card"
        @tab-change="switchDeviceSource"
      >
        <el-tab-pane :label="tr('Login Devices')" name="login" />
        <el-tab-pane :label="tr('FCM Devices')" name="fcm" />
      </el-tabs>
      <el-form inline class="user-search-form">
        <el-form-item :label="tr('Player Number')"
          ><el-input v-model="query.userId" /></el-form-item
        ><el-form-item :label="tr('Username')"
          ><el-input v-model="query.username" /></el-form-item
        ><el-form-item
          v-if="deviceSource === 'login'"
          :label="tr('Fingerprint')"
          ><el-input v-model="query.fingerprint" /></el-form-item
        ><el-form-item :label="tr('Platform')"
          ><el-input v-model="query.platform"
        /></el-form-item>
        <el-form-item
          v-if="deviceSource === 'login'"
          :label="tr('Device Status')"
          ><el-select v-model="query.status" clearable
            ><el-option :label="tr('ACTIVE')" value="1" /><el-option
              :label="tr('BLOCKED')"
              value="2" /></el-select></el-form-item
        ><el-form-item
          v-if="deviceSource === 'login'"
          :label="tr('Risk Status')"
          ><el-select v-model="query.riskStatus" clearable
            ><el-option :label="tr('NORMAL')" value="1" /><el-option
              :label="tr('OBSERVED')"
              value="2" /><el-option
              :label="tr('RISKY')"
              value="3" /></el-select
        ></el-form-item>
        <el-form-item v-if="deviceSource === 'fcm'" :label="tr('FCM Status')"
          ><el-select v-model="query.pushStatus" clearable
            ><el-option :label="tr('ENABLED')" value="ENABLED" /><el-option
              :label="tr('INVALIDATED')"
              value="INVALIDATED" /></el-select
        ></el-form-item>
        <el-form-item v-if="deviceSource === 'fcm'" :label="tr('Network Type')"
          ><el-select v-model="query.networkType" clearable
            ><el-option
              v-for="item in [
                'WIFI',
                'CELLULAR',
                'ETHERNET',
                'VPN',
                'OFFLINE',
                'OTHER',
              ]"
              :key="item"
              :label="tr(item)"
              :value="item" /></el-select
        ></el-form-item>
        <el-form-item
          v-if="deviceSource === 'fcm'"
          :label="tr('Notification Permission')"
          ><el-select v-model="query.notificationPermission" clearable
            ><el-option :label="tr('GRANTED')" value="GRANTED" />
            <el-option :label="tr('DENIED')" value="DENIED" /></el-select
        ></el-form-item>
        <el-form-item v-if="deviceSource === 'login'" :label="tr('First Seen')"
          ><el-date-picker
            v-model="query.firstRange"
            type="daterange"
            value-format="YYYY-MM-DD" /></el-form-item
        ><el-form-item :label="tr('Last Seen')"
          ><el-date-picker
            v-model="query.lastRange"
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
    <UserCenterCard
      :title="tr(deviceSource === 'fcm' ? 'FCM Devices' : 'Device Management')"
      ><el-table
        v-if="deviceSource === 'login'"
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
          align="center"
        /><el-table-column
          prop="userNo"
          :label="tr('Player Number')"
          width="150"
        /><el-table-column
          prop="nickname"
          :label="tr('Username')"
          width="130"
        /><el-table-column
          prop="platform"
          :label="tr('Platform')"
          width="100"
        /><el-table-column
          prop="deviceModel"
          :label="tr('Device Model')"
          width="150"
        /><el-table-column
          prop="os"
          :label="tr('Operating System')"
          width="150"
        /><el-table-column :label="tr('App / Browser Version')" width="180"
          ><template #default="{ row }"
            >{{ row.appVersion || "-" }} / {{ row.browser || "-" }}</template
          ></el-table-column
        ><el-table-column
          prop="fingerprintHash"
          :label="tr('Fingerprint Summary')"
          width="170"
        /><el-table-column
          prop="firstIP"
          :label="tr('First IP')"
          width="140"
        /><el-table-column
          prop="lastIP"
          :label="tr('Last IP')"
          width="140"
        /><el-table-column :label="tr('First Seen')" width="170"
          ><template #default="{ row }">{{
            formatBackendTime(row.firstSeenAt)
          }}</template></el-table-column
        ><el-table-column :label="tr('Last Seen')" width="170"
          ><template #default="{ row }">{{
            formatBackendTime(row.lastSeenAt)
          }}</template></el-table-column
        ><el-table-column
          prop="loginCount"
          :label="tr('Login Count')"
          width="105"
        /><el-table-column
          prop="associatedAccounts"
          :label="tr('Associated Accounts')"
          width="150"
        /><el-table-column
          prop="statusText"
          :label="tr('Device Status')"
          width="120"
          ><template #default="{ row }"
            ><UserStatusTag
              :value="row.statusText.toUpperCase()"
              :label="
                tr(row.statusText.toUpperCase())
              " /></template></el-table-column
        ><el-table-column
          prop="riskStatusText"
          :label="tr('Risk Status')"
          width="120"
          ><template #default="{ row }"
            ><UserStatusTag
              :value="row.riskStatusText"
              :label="tr(row.riskStatusText)" /></template
        ></el-table-column>
        <el-table-column :label="tr('Actions')" fixed="right" width="240"
          ><template #default="{ row }"
            ><div class="table-actions">
              <el-button link type="primary" @click="view(row)">{{
                tr("View Details")
              }}</el-button
              ><el-button
                link
                type="primary"
                @click="openUserDetail(row.userID)"
                >{{ tr("View User") }}</el-button
              ><el-dropdown v-if="can('user:device:manage')"
                ><el-button link type="primary">{{ tr("Manage") }}</el-button
                ><template #dropdown
                  ><el-dropdown-menu
                    ><el-dropdown-item
                      @click="
                        ask(row, {
                          trustStatus:
                            row.trustStatus === 2 ? 'NORMAL' : 'TRUSTED',
                        })
                      "
                      >{{ tr(row.trustStatus === 2 ? "Untrust" : "Trust") }}
                      {{ tr("Device") }}</el-dropdown-item
                    ><el-dropdown-item
                      @click="
                        ask(row, {
                          status: row.status === 2 ? 'ACTIVE' : 'BLOCKED',
                        })
                      "
                      >{{ tr(row.status === 2 ? "Unblock" : "Block") }}
                      {{ tr("Device") }}</el-dropdown-item
                    ></el-dropdown-menu
                  ></template
                ></el-dropdown
              >
            </div></template
          ></el-table-column
        ></el-table
      >
      <el-table
        v-else
        v-loading="loading"
        :data="pushRows"
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
        <el-table-column :label="tr('Mobile')" min-width="160">
          <template #default="{ row }">
            {{
              [row.phoneCountryCode, row.phone].filter(Boolean).join(" ") || "-"
            }}
          </template>
        </el-table-column>
        <el-table-column
          prop="nickname"
          :label="tr('Username')"
          min-width="130"
        />
        <el-table-column
          :label="tr('Device Model')"
          min-width="175"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{
              [row.manufacturer, row.deviceModel].filter(Boolean).join(" ") ||
              "-"
            }}
          </template>
        </el-table-column>
        <el-table-column :label="tr('OS Version')" min-width="140">
          <template #default="{ row }"
            >Android {{ row.osVersion || "-" }} / SDK
            {{ row.sdkVersion || "-" }}</template
          >
        </el-table-column>
        <el-table-column :label="tr('Network Type')" width="125">
          <template #default="{ row }">
            <el-tag :type="networkTagType(row.networkType)" effect="light">{{
              tr(row.networkType || "OTHER")
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="carrierName"
          :label="tr('Carrier')"
          min-width="135"
          show-overflow-tooltip
        />
        <el-table-column :label="tr('Notification Permission')" width="155">
          <template #default="{ row }">
            <el-tag
              :type="
                row.notificationPermission === 'GRANTED' ? 'success' : 'danger'
              "
              effect="light"
            >
              {{ tr(row.notificationPermission || "DENIED") }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="tr('App Version')" width="140">
          <template #default="{ row }"
            >{{ row.appVersion || "-" }} ({{
              row.appVersionCode || 0
            }})</template
          >
        </el-table-column>
        <el-table-column prop="lastIP" :label="tr('Last IP')" min-width="135" />
        <el-table-column :label="tr('FCM Status')" width="120">
          <template #default="{ row }">
            <UserStatusTag
              :value="row.status"
              :label="tr(row.status)"
              :type-override="row.status === 'ENABLED' ? 'success' : 'danger'"
            />
          </template>
        </el-table-column>
        <el-table-column :label="tr('Last Seen')" width="175">
          <template #default="{ row }">{{
            formatBackendTime(row.lastSeenAt)
          }}</template>
        </el-table-column>
        <el-table-column :label="tr('Invalidated At')" width="175">
          <template #default="{ row }">{{
            formatBackendTime(row.invalidatedAt)
          }}</template>
        </el-table-column>
        <el-table-column
          prop="lastError"
          :label="tr('Last Error')"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column :label="tr('Actions')" fixed="right" width="190">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewPush(row)">{{
              tr("View Details")
            }}</el-button>
            <el-button
              link
              type="primary"
              @click="openUserDetail(row.userID)"
              >{{ tr("View User") }}</el-button
            >
          </template>
        </el-table-column>
      </el-table>
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
      :title="tr('Device Details')"
      width="min(900px, 92vw)"
      destroy-on-close
      append-to-body
    >
      <el-tabs v-if="current" v-model="activeDetailTab" type="card">
        <el-tab-pane :label="tr('Overview')" name="overview">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item
                v-for="key in [
                  'deviceID',
                  'platform',
                  'deviceModel',
                  'os',
                  'browser',
                  'appVersion',
                  'fingerprintHash',
                  'firstIP',
                  'lastIP',
                  'firstSeenAt',
                  'lastSeenAt',
                  'loginCount',
                  'riskStatusText',
                  'trustStatusText',
                ]"
                :key="key"
                :label="detailLabel(key)"
              >
                <UserStatusTag
                  v-if="['riskStatusText', 'trustStatusText'].includes(key)"
                  :value="String(current[key as keyof UserDevice] || '')"
                  :label="tr(String(current[key as keyof UserDevice] || ''))"
                />
                <template v-else>{{
                  key.endsWith("At")
                    ? formatBackendTime(
                        String(current[key as keyof UserDevice] || ""),
                      )
                    : current[key as keyof UserDevice] || "-"
                }}</template>
              </el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-tab-pane>
        <el-tab-pane :label="tr('Associated Users')" name="users">
          <el-card shadow="never">
            <el-table :data="associated" border :empty-text="tr('No data')">
              <el-table-column prop="userNo" :label="tr('Player Number')" />
              <el-table-column prop="username" :label="tr('Username')" />
              <el-table-column prop="statusText" :label="tr('Status')">
                <template #default="{ row }">
                  <UserStatusTag
                    :value="row.statusText"
                    :label="tr(row.statusText)"
                  />
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
    <el-dialog
      v-model="pushDetailVisible"
      :title="tr('Device Details')"
      width="min(980px, 94vw)"
      destroy-on-close
      append-to-body
    >
      <el-tabs v-if="currentPush" type="card" model-value="overview">
        <el-tab-pane :label="tr('Overview')" name="overview">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item :label="tr('Mobile')">
                {{
                  [currentPush.phoneCountryCode, currentPush.phone]
                    .filter(Boolean)
                    .join(" ") || "-"
                }}
              </el-descriptions-item>
              <el-descriptions-item :label="tr('Username')">{{
                currentPush.nickname || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Manufacturer')">{{
                currentPush.manufacturer || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Brand')">{{
                currentPush.brand || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Device Model')">{{
                currentPush.deviceModel || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Device Name')">{{
                currentPush.deviceName || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('OS Version')"
                >Android
                {{ currentPush.osVersion || "-" }}</el-descriptions-item
              >
              <el-descriptions-item :label="tr('SDK Version')">{{
                currentPush.sdkVersion || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Security Patch')">{{
                currentPush.securityPatch || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Network Type')">
                <el-tag
                  :type="networkTagType(currentPush.networkType)"
                  effect="light"
                  >{{ tr(currentPush.networkType || "OTHER") }}</el-tag
                >
              </el-descriptions-item>
              <el-descriptions-item :label="tr('Metered Network')">{{
                tr(currentPush.networkMetered ? "Yes" : "No")
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Roaming')">{{
                tr(currentPush.networkRoaming ? "Yes" : "No")
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Carrier')">{{
                currentPush.carrierName || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Screen Resolution')"
                >{{ currentPush.screenWidth || 0 }} x
                {{ currentPush.screenHeight || 0 }}</el-descriptions-item
              >
              <el-descriptions-item :label="tr('Screen Density')">{{
                currentPush.screenDensity || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Timezone')">{{
                currentPush.timezone || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Locale')">{{
                currentPush.locale || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Battery')">{{
                currentPush.batteryLevel >= 0
                  ? `${currentPush.batteryLevel}%`
                  : "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Charging')">{{
                tr(currentPush.charging ? "Yes" : "No")
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Notification Permission')">
                <el-tag
                  :type="
                    currentPush.notificationPermission === 'GRANTED'
                      ? 'success'
                      : 'danger'
                  "
                  effect="light"
                >
                  {{ tr(currentPush.notificationPermission || "DENIED") }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item :label="tr('App Version')"
                >{{ currentPush.appVersion || "-" }} ({{
                  currentPush.appVersionCode || 0
                }})</el-descriptions-item
              >
              <el-descriptions-item :label="tr('Language')">{{
                currentPush.language || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Last IP')">{{
                currentPush.lastIP || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('User Agent')" :span="2">{{
                currentPush.userAgent || "-"
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('Last Seen')">{{
                formatBackendTime(currentPush.lastSeenAt)
              }}</el-descriptions-item>
              <el-descriptions-item :label="tr('FCM Status')">
                <UserStatusTag
                  :value="currentPush.status"
                  :label="tr(currentPush.status)"
                  :type-override="
                    currentPush.status === 'ENABLED' ? 'success' : 'danger'
                  "
                />
              </el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
    <UserDetailDialog v-model="detailUserVisible" :user-id="detailUserID" />
    <ReasonConfirmDialog
      v-model="confirmVisible"
      :title="tr('Update Device')"
      :message="
        tr(
          'This device control is enforced during login and will be recorded in the audit log.',
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
</style>
