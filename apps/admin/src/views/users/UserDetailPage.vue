<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import UserCenterCard from "@/components/users/UserCenterCard.vue";
import UserStatusTag from "@/components/users/UserStatusTag.vue";
import ReasonConfirmDialog from "@/components/users/ReasonConfirmDialog.vue";
import {
  addAdminUserNote,
  fetchAdminDevices,
  fetchAdminLoginRecords,
  fetchAdminRiskUser,
  fetchAdminUser,
  fetchAdminUserNotes,
  fetchAdminUserOperationLogs,
  updateAdminUserRestrictions,
  updateAdminUserTags,
  type LoginRecord,
  type OperationLog,
  type RiskUserDetail,
  type UserDetail,
  type UserDevice,
  type UserNote,
} from "@/api/adminUsers";
import { formatBackendTime } from "@/utils/timezone";
import { useAdminPermissions } from "@/composables/useAdminPermissions";
import { useAdminCenterText } from "@/composables/useAdminCenterText";
const route = useRoute(),
  router = useRouter();
const props = withDefaults(
  defineProps<{ userId?: number; embedded?: boolean }>(),
  { userId: 0, embedded: false },
);
const { can, loadPermissions } = useAdminPermissions();
const tr = useAdminCenterText("userCenter");
const userID = computed(() =>
  Number(props.userId || route.params.userId || route.query.userId || 0),
);
const loading = ref(false),
  actionLoading = ref(false),
  activeTab = ref("overview"),
  user = ref<UserDetail | null>(null),
  notes = ref<UserNote[]>([]),
  loginRecords = ref<LoginRecord[]>([]),
  devices = ref<UserDevice[]>([]),
  risk = ref<RiskUserDetail | null>(null),
  logs = ref<OperationLog[]>([]),
  noteContent = ref(""),
  tagInput = ref<string[]>([]),
  confirmVisible = ref(false),
  pendingRestriction = ref<{ key: string; value: boolean } | null>(null);
async function load() {
  if (!userID.value) return;
  loading.value = true;
  try {
    const [detail, noteData, loginData, deviceData, logData] =
      await Promise.all([
        fetchAdminUser(userID.value),
        fetchAdminUserNotes(userID.value),
        fetchAdminLoginRecords({ userId: userID.value, page: 1, pageSize: 20 }),
        fetchAdminDevices({ userId: userID.value, page: 1, pageSize: 20 }),
        fetchAdminUserOperationLogs(userID.value, { page: 1, pageSize: 20 }),
      ]);
    user.value = detail;
    tagInput.value = detail.internalTags
      ? detail.internalTags
          .split(",")
          .map((v) => v.trim())
          .filter(Boolean)
      : [];
    notes.value = noteData;
    loginRecords.value = loginData.list;
    devices.value = deviceData.list;
    logs.value = logData.list;
    try {
      risk.value = await fetchAdminRiskUser(userID.value);
    } catch {
      risk.value = null;
    }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : tr("Request failed"));
    user.value = null;
  } finally {
    loading.value = false;
  }
}
async function addNote() {
  const content = noteContent.value.trim();
  if (!content) return;
  actionLoading.value = true;
  try {
    await addAdminUserNote(userID.value, content);
    noteContent.value = "";
    notes.value = await fetchAdminUserNotes(userID.value);
    ElMessage.success(tr("Note added"));
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : tr("Request failed"));
  } finally {
    actionLoading.value = false;
  }
}
async function saveTags() {
  actionLoading.value = true;
  try {
    await updateAdminUserTags(
      userID.value,
      tagInput.value,
      "Update internal user tags",
    );
    await load();
    ElMessage.success(tr("Tags updated"));
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : tr("Request failed"));
  } finally {
    actionLoading.value = false;
  }
}
function askRestriction(key: string, value: boolean) {
  pendingRestriction.value = { key, value };
  confirmVisible.value = true;
}
async function confirmRestriction(reason: string) {
  if (!pendingRestriction.value) return;
  actionLoading.value = true;
  try {
    await updateAdminUserRestrictions(userID.value, {
      [pendingRestriction.value.key]: pendingRestriction.value.value,
      reason,
    });
    confirmVisible.value = false;
    await load();
    ElMessage.success(tr("Restriction updated"));
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : tr("Request failed"));
  } finally {
    actionLoading.value = false;
  }
}
const basicItems = computed(() =>
  user.value
    ? [
        ["Player Number", user.value.userNo],
        ["Username", user.value.username],
        ["Email", user.value.email],
        ["Mobile", `${user.value.phoneCountryCode} ${user.value.phone}`],
        ["Country", user.value.countryCode],
        ["Language", user.value.language],
        ["Currency", user.value.currency],
        ["Registration Channel", user.value.registrationChannel],
        ["Referral Code", user.value.referralCode],
        ["Registration Time", formatBackendTime(user.value.registeredAt)],
        ["Account Status", user.value.statusText],
      ]
    : [],
);
const summaryItems = computed(() =>
  user.value
    ? [
        ["Available Balance", user.value.availableBalance],
        ["Frozen Balance", user.value.frozenBalance],
        ["Bonus Balance", user.value.bonusBalance],
        ["VIP Level", user.value.vipLevel],
        ["KYC Status", user.value.kycStatus],
        ["Risk Status", user.value.riskStatus],
        ["Total Deposit", user.value.totalDeposit],
        ["Total Withdrawal", user.value.totalWithdrawal],
        ["Total Bet", user.value.totalBet],
        ["Total Win", user.value.totalWin],
        ["Last Active Time", formatBackendTime(user.value.lastActiveAt)],
      ]
    : [],
);
function isStatusItem(label: unknown) {
  return ["Account Status", "KYC Status", "Risk Status"].includes(
    String(label),
  );
}
watch(userID, load, { immediate: true });
onMounted(loadPermissions);
</script>
<template>
  <div :class="embedded ? 'user-detail-embedded' : 'page-stack'">
    <UserCenterCard
      :title="
        user ? `${tr('User Detail')} · ${user.userNo}` : tr('User Detail')
      "
      ><template #header
        ><el-button v-if="!embedded" @click="router.push('/users/list')">{{
          tr("Back to User List")
        }}</el-button></template
      ><el-skeleton v-if="loading" :rows="8" animated /><template
        v-else-if="user"
        ><div class="detail-heading">
          <div>
            <strong>{{ user.username || user.nickname || user.userNo }}</strong
            ><span>{{ user.phoneCountryCode }} {{ user.phone }}</span>
          </div>
          <div>
            <UserStatusTag
              :value="user.statusText"
              :label="tr(user.statusText)"
            />
            <UserStatusTag
              :value="user.riskStatus"
              :label="tr(user.riskStatus)"
            />
          </div>
        </div>
        <el-tabs v-model="activeTab" type="card">
          <el-tab-pane :label="tr('Overview')" name="overview"
            ><section class="detail-section">
              <h3>{{ tr("Basic Information") }}</h3>
              <el-descriptions :column="3" border
                ><el-descriptions-item
                  v-for="item in basicItems"
                  :key="String(item[0])"
                  :label="tr(String(item[0]))"
                  ><UserStatusTag
                    v-if="isStatusItem(item[0])"
                    :value="String(item[1] || '')"
                    :label="tr(String(item[1] || ''))"
                  /><template v-else>{{
                    item[1] || "-"
                  }}</template></el-descriptions-item
                ></el-descriptions
              >
            </section>
            <section class="detail-section">
              <h3>{{ tr("Account Summary") }}</h3>
              <el-descriptions :column="3" border
                ><el-descriptions-item
                  v-for="item in summaryItems"
                  :key="String(item[0])"
                  :label="tr(String(item[0]))"
                  ><UserStatusTag
                    v-if="isStatusItem(item[0])"
                    :value="String(item[1] || '')"
                    :label="tr(String(item[1] || ''))"
                  /><template v-else>{{
                    item[1] ?? "-"
                  }}</template></el-descriptions-item
                ></el-descriptions
              >
              <div class="future-links">
                <span>{{ tr("Related records:") }}</span
                ><el-button
                  link
                  type="primary"
                  @click="router.push('/funds/ledgers')"
                  >{{ tr("Wallet Transactions") }}</el-button
                ><el-button
                  link
                  type="primary"
                  @click="router.push('/bets/records')"
                  >{{ tr("Betting Records") }}</el-button
                ><el-button
                  link
                  type="primary"
                  @click="router.push('/funds/deposits')"
                  >{{ tr("Deposits") }}</el-button
                ><el-button
                  link
                  type="primary"
                  @click="router.push('/funds/withdrawals')"
                  >{{ tr("Withdrawals") }}</el-button
                >
              </div>
            </section>
            <section class="detail-section">
              <h3>{{ tr("Registration and Login") }}</h3>
              <el-descriptions :column="3" border
                ><el-descriptions-item :label="tr('Registration IP')">{{
                  user.registeredIP || "-"
                }}</el-descriptions-item
                ><el-descriptions-item :label="tr('Registration Device')">{{
                  user.registrationDevice || "-"
                }}</el-descriptions-item
                ><el-descriptions-item :label="tr('Last Login IP')">{{
                  user.lastLoginIP || "-"
                }}</el-descriptions-item
                ><el-descriptions-item :label="tr('Last Login Device')">{{
                  user.lastLoginDevice || "-"
                }}</el-descriptions-item
                ><el-descriptions-item :label="tr('Last Login Time')">{{
                  formatBackendTime(user.lastLoginAt)
                }}</el-descriptions-item
                ><el-descriptions-item :label="tr('Failed Login Count')">{{
                  user.failedLoginCount
                }}</el-descriptions-item
                ><el-descriptions-item :label="tr('Last Password Change')">{{
                  formatBackendTime(user.lastPasswordChangedAt)
                }}</el-descriptions-item></el-descriptions
              >
            </section>
            <section class="detail-section">
              <h3>{{ tr("Account Restrictions") }}</h3>
              <div class="restriction-grid">
                <div
                  v-for="item in [
                    { key: 'loginEnabled', label: 'Login enabled' },
                    {
                      key: 'gameAccessEnabled',
                      label: 'Game access enabled',
                    },
                    { key: 'depositEnabled', label: 'Deposit enabled' },
                    { key: 'withdrawalEnabled', label: 'Withdrawal enabled' },
                    {
                      key: 'riskReviewRequired',
                      label: 'Risk review required',
                    },
                  ]"
                  :key="item.key"
                >
                  <span>{{ tr(item.label) }}</span
                  ><el-switch
                    :model-value="user[item.key as keyof UserDetail] === 1"
                    :disabled="!can('user:restriction:update')"
                    @change="askRestriction(item.key, Boolean($event))"
                  />
                </div>
              </div>
            </section>
            <section class="detail-section">
              <h3>{{ tr("Internal Tags") }}</h3>
              <div class="tag-editor">
                <el-select
                  v-model="tagInput"
                  multiple
                  filterable
                  allow-create
                  default-first-option
                  :placeholder="tr('Add internal tags')"
                /><el-button
                  v-if="can('user:note:create')"
                  type="primary"
                  :loading="actionLoading"
                  @click="saveTags"
                  >{{ tr("Save Tags") }}</el-button
                >
              </div>
            </section>
            <section class="detail-section">
              <h3>{{ tr("Internal Notes") }}</h3>
              <div v-if="can('user:note:create')" class="note-editor">
                <el-input
                  v-model="noteContent"
                  type="textarea"
                  :rows="3"
                  maxlength="1000"
                  show-word-limit
                /><el-button
                  type="primary"
                  :loading="actionLoading"
                  @click="addNote"
                  >{{ tr("Add Note") }}</el-button
                >
              </div>
              <el-timeline
                ><el-timeline-item
                  v-for="note in notes"
                  :key="note.id"
                  :timestamp="formatBackendTime(note.createdAt)"
                  ><strong>{{ note.operatorUsername }}</strong>
                  <p>{{ note.content }}</p></el-timeline-item
                ></el-timeline
              >
            </section></el-tab-pane
          >
          <el-tab-pane :label="tr('Login Records')" name="login"
            ><el-table :data="loginRecords" border
              ><el-table-column prop="loginResult" :label="tr('Result')"
                ><template #default="{ row }"
                  ><UserStatusTag
                    :value="row.loginResult"
                    :label="tr(row.loginResult)" /></template></el-table-column
              ><el-table-column
                prop="clientIP"
                :label="tr('IP')"
              /><el-table-column :label="tr('Time')"
                ><template #default="{ row }">{{
                  formatBackendTime(row.loginAt)
                }}</template></el-table-column
              ></el-table
            ></el-tab-pane
          >
          <el-tab-pane :label="tr('Devices')" name="devices"
            ><el-table :data="devices" border
              ><el-table-column prop="platform"
                :label="tr('Platform')"
              /><el-table-column prop="statusText" :label="tr('Status')"
                ><template #default="{ row }"
                  ><UserStatusTag
                    :value="row.statusText.toUpperCase()"
                    :label="
                      tr(row.statusText.toUpperCase())
                    " /></template></el-table-column
              ><el-table-column :label="tr('Last Seen')"
                ><template #default="{ row }">{{
                  formatBackendTime(row.lastSeenAt)
                }}</template></el-table-column
              ></el-table
            ></el-tab-pane
          >
          <el-tab-pane :label="tr('Risk Information')" name="risk"
            ><el-empty v-if="!risk" :description="tr('No risk profile')" />
            <template v-else>
              <el-descriptions :column="2" border>
                <el-descriptions-item :label="tr('Risk Level')">
                  <UserStatusTag
                    :value="risk.profile.riskLevel"
                    :label="tr(risk.profile.riskLevel)"
                  />
                </el-descriptions-item>
                <el-descriptions-item :label="tr('Risk Status')">
                  <UserStatusTag
                    :value="risk.profile.riskStatus"
                    :label="tr(risk.profile.riskStatus)"
                  />
                </el-descriptions-item>
                <el-descriptions-item :label="tr('Risk Tags')">{{
                  risk.profile.riskTags || "-"
                }}</el-descriptions-item>
                <el-descriptions-item :label="tr('Latest Reason')">{{
                  risk.profile.latestReason || "-"
                }}</el-descriptions-item>
              </el-descriptions>
              <el-table
                class="risk-event-table"
                :data="risk.events"
                border
                :empty-text="tr('No risk events')"
              >
                <el-table-column prop="eventType" :label="tr('Event')" />
                <el-table-column prop="riskLevel" :label="tr('Risk Level')">
                  <template #default="{ row }">
                    <UserStatusTag
                      :value="row.riskLevel"
                      :label="tr(row.riskLevel)"
                    />
                  </template>
                </el-table-column>
                <el-table-column
                  prop="description"
                  :label="tr('Description')"
                />
                <el-table-column :label="tr('Time')" width="180">
                  <template #default="{ row }">{{
                    formatBackendTime(row.createdAt)
                  }}</template>
                </el-table-column>
              </el-table>
            </template>
          </el-tab-pane>
          <el-tab-pane :label="tr('Operation Logs')" name="logs"
            ><el-table :data="logs" border
              ><el-table-column
                prop="operatorUsername"
                :label="tr('Operator')"
              /><el-table-column
                prop="action"
                :label="tr('Action')"
              /><el-table-column
                prop="reason"
                :label="tr('Reason')"
              /><el-table-column :label="tr('Time')"
                ><template #default="{ row }">{{
                  formatBackendTime(row.createdAt)
                }}</template></el-table-column
              ></el-table
            ></el-tab-pane
          >
        </el-tabs></template
      ><el-empty v-else :description="tr('User not found')" /></UserCenterCard
    ><ReasonConfirmDialog
      v-model="confirmVisible"
      :title="tr('Update Account Restriction')"
      :message="
        tr(
          'This restriction is enforced by the backend and the change will be audited.',
        )
      "
      :loading="actionLoading"
      @confirm="confirmRestriction"
    />
  </div>
</template>
<style scoped>
.detail-heading {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.detail-heading strong {
  display: block;
  font-size: 18px;
}
.risk-event-table {
  margin-top: 16px;
}
.detail-heading span {
  display: block;
  margin-top: 4px;
  color: var(--admin-text-muted);
}
.detail-section {
  margin: 22px 0;
}
.detail-section h3 {
  margin: 0 0 12px;
  font-size: 15px;
}
.restriction-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.restriction-grid > div {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--admin-border);
  border-radius: 4px;
}
.note-editor,
.tag-editor {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 20px;
}
.note-editor .el-button,
.tag-editor .el-button {
  flex: none;
}
.tag-editor .el-select {
  width: 520px;
}
.future-links {
  margin-top: 12px;
  color: var(--admin-text-secondary);
}
.detail-json {
  white-space: pre-wrap;
  color: var(--admin-text-secondary);
}
</style>
