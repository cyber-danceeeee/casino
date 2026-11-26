<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { ArrowLeft, Link } from "@element-plus/icons-vue";
import {
  systemAPI,
  type PlatformBranding,
  type TenantDetail,
} from "@/api/adminSystem";
import { formatBackendTime } from "@/utils/timezone";
const { t } = useI18n(),
  route = useRoute(),
  router = useRouter();
const loading = ref(false),
  detail = ref<TenantDetail>(),
  activeTab = ref(String(route.query.tab || "sites"));
const platformBranding = ref<PlatformBranding>();
const overview = computed(() => detail.value?.overview);
const value = (row: Record<string, unknown>, key: string) => row[key];
async function load() {
  loading.value = true;
  try {
    detail.value = await systemAPI.tenants.detail(Number(route.params.id));
  } finally {
    loading.value = false;
  }
}
function open(domain?: string) {
  if (domain) window.open(`https://${domain}`, "_blank", "noopener,noreferrer");
}
onMounted(() =>
  Promise.all([
    load(),
    systemAPI
      .platformBranding()
      .then((value) => (platformBranding.value = value))
      .catch(() => undefined),
  ]),
);
</script>
<template>
  <main class="tenant-detail" v-loading="loading">
    <header class="detail-heading">
      <div class="heading-main">
        <el-button
          circle
          :icon="ArrowLeft"
          @click="router.push('/system/tenants')"
        />
        <div>
          <p>{{ overview?.tenantCode || "-" }}</p>
          <h1>{{ overview?.tenantName || t("tenantCenter.detail.title") }}</h1>
          <span>Tenant ID: {{ overview?.id || "-" }}</span>
        </div>
      </div>
      <div class="heading-actions">
        <el-switch :model-value="overview?.enabled" disabled /><el-tag
          :type="overview?.enabled ? 'success' : 'danger'"
          >{{
            overview?.enabled
              ? t("tenantCenter.status.ACTIVE")
              : t("tenantCenter.status.DISABLED")
          }}</el-tag
        ><el-button
          :icon="Link"
          :disabled="!overview?.adminDomain"
          @click="open(overview?.adminDomain)"
          >{{ t("tenantCenter.actions.openAdmin") }}</el-button
        >
      </div>
    </header>
    <section v-if="overview" class="overview-band">
      <dl>
        <div>
          <dt>{{ t("tenantCenter.fields.brand") }}</dt>
          <dd>{{ overview.tenantName }}</dd>
        </div>
        <div>
          <dt>{{ t("tenantCenter.fields.plan") }}</dt>
          <dd>{{ overview.packageCode || overview.planCode }}</dd>
        </div>
        <div>
          <dt>{{ t("tenantCenter.fields.country") }}</dt>
          <dd>{{ overview.countryCode }}</dd>
        </div>
        <div>
          <dt>{{ t("tenantCenter.fields.currency") }}</dt>
          <dd>{{ overview.defaultCurrency }}</dd>
        </div>
        <div>
          <dt>{{ t("createBrand.fields.language") }}</dt>
          <dd>{{ overview.defaultLanguage }}</dd>
        </div>
        <div>
          <dt>{{ t("createBrand.fields.timezone") }}</dt>
          <dd>{{ overview.timezone }}</dd>
        </div>
        <div>
          <dt>{{ t("tenantCenter.fields.domain") }}</dt>
          <dd>{{ overview.primaryDomain || "-" }}</dd>
        </div>
        <div>
          <dt>{{ t("tenantCenter.fields.admin") }}</dt>
          <dd>{{ overview.ownerUsername || "-" }}</dd>
        </div>
      </dl>
    </section>
    <section class="detail-band">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('tenantCenter.tabs.sites')" name="sites"
          ><el-table :data="detail?.sites || []" row-key="id"
            ><el-table-column
              prop="siteName"
              :label="t('tenantCenter.fields.site')"
              min-width="150"
              ><template #default="{ row }"
                ><strong>{{ row.siteName }}</strong
                ><small>{{ row.siteCode }} · #{{ row.id }}</small></template
              ></el-table-column
            ><el-table-column
              prop="countryCode"
              :label="t('tenantCenter.fields.country')"
              width="95"
            /><el-table-column
              prop="currencyCode"
              :label="t('tenantCenter.fields.currency')"
              width="95"
            /><el-table-column
              prop="primaryDomain"
              :label="t('tenantCenter.fields.domain')"
              min-width="180"
            /><el-table-column :label="t('common.status')" width="95"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="row.enabled"
                  disabled /></template></el-table-column
            ><el-table-column
              prop="configStatus"
              :label="t('tenantCenter.fields.publishStatus')"
              width="130"
              ><template #default="{ row }"
                ><el-tag
                  :type="
                    row.configStatus === 'PUBLISHED' ? 'success' : 'warning'
                  "
                  >{{ row.configStatus }}</el-tag
                ></template
              ></el-table-column
            ><el-table-column
              prop="configVersion"
              :label="t('tenantCenter.fields.publishedVersion')"
              width="135"
            /><el-table-column
              :label="t('tenantCenter.fields.draftVersion')"
              width="120"
              ><template #default="{ row }">{{
                row.configStatus === "DRAFT"
                  ? row.configVersion + 1
                  : row.configVersion
              }}</template></el-table-column
            ><el-table-column
              prop="configPublishedAt"
              :label="t('tenantCenter.fields.lastPublished')"
              width="175"
              ><template #default="{ row }">{{
                row.configPublishedAt
                  ? formatBackendTime(row.configPublishedAt)
                  : "-"
              }}</template></el-table-column
            ></el-table
          ></el-tab-pane
        >
        <el-tab-pane :label="t('tenantCenter.tabs.admins')" name="admins"
          ><el-table :data="detail?.administrators || []" row-key="id"
            ><el-table-column
              prop="username"
              :label="t('tenantCenter.fields.admin')"
              min-width="150"
              ><template #default="{ row }"
                ><strong>{{ row.username }}</strong
                ><small>{{ row.displayName }}</small></template
              ></el-table-column
            ><el-table-column
              prop="email"
              :label="t('tenantCenter.fields.contactEmail')"
              min-width="180"
            /><el-table-column
              prop="mobile"
              :label="t('accountCenter.profile.mobile')"
              min-width="130"
            /><el-table-column :label="t('common.status')" width="95"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="row.status"
                  disabled /></template></el-table-column
            ><el-table-column label="MFA" width="90"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="row.mfaEnabled"
                  disabled /></template></el-table-column
            ><el-table-column
              prop="lastLoginIP"
              label="IP"
              width="140"
            /><el-table-column
              prop="lastLoginTime"
              :label="t('tenantCenter.fields.lastLogin')"
              width="175"
              ><template #default="{ row }">{{
                row.lastLoginTime ? formatBackendTime(row.lastLoginTime) : "-"
              }}</template></el-table-column
            ></el-table
          ></el-tab-pane
        >
        <el-tab-pane :label="t('tenantCenter.tabs.domains')" name="domains"
          ><el-table :data="detail?.domains || []" row-key="id"
            ><el-table-column
              prop="domain"
              :label="t('tenantCenter.fields.domain')"
              min-width="220"
            /><el-table-column
              prop="domainType"
              :label="t('tenantCenter.fields.domainType')"
              width="110"
              ><template #default="{ row }"
                ><el-tag effect="plain">{{ row.domainType }}</el-tag></template
              ></el-table-column
            ><el-table-column :label="t('common.status')" width="95"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="row.enabled"
                  disabled /></template></el-table-column
            ><el-table-column
              prop="verificationStatus"
              :label="t('tenantCenter.fields.verification')"
              width="140"
            /><el-table-column
              prop="sslStatus"
              label="SSL"
              width="110"
            /><el-table-column
              prop="sslExpiresAt"
              :label="t('tenantCenter.fields.sslExpiry')"
              width="175"
              ><template #default="{ row }">{{
                row.sslExpiresAt ? formatBackendTime(row.sslExpiresAt) : "-"
              }}</template></el-table-column
            ></el-table
          ></el-tab-pane
        >
        <el-tab-pane :label="t('tenantCenter.tabs.providers')" name="providers"
          ><el-table :data="detail?.providerAuthorizations || []" row-key="id"
            ><el-table-column
              prop="providerCode"
              label="Provider"
              min-width="150"
              ><template #default="{ row }"
                ><el-tag>{{ value(row, "providerCode") }}</el-tag
                ><span class="inline-name">{{
                  value(row, "name")
                }}</span></template
              ></el-table-column
            ><el-table-column
              prop="siteID"
              label="Site ID"
              width="100" /><el-table-column
              prop="apiStatus"
              :label="t('tenantCenter.fields.health')"
              width="130"
              ><template #default="{ row }"
                ><el-tag
                  :type="
                    ['UP', 'ACTIVE', 'HEALTHY'].includes(
                      String(value(row, 'apiStatus')),
                    )
                      ? 'success'
                      : 'warning'
                  "
                  >{{ value(row, "apiStatus") }}</el-tag
                ></template
              ></el-table-column
            ><el-table-column :label="t('common.status')" width="95"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="Number(value(row, 'status')) === 1"
                  disabled /></template></el-table-column
            ><el-table-column
              prop="maintenanceMessage"
              :label="t('tenantCenter.fields.maintenance')"
              min-width="200"
              show-overflow-tooltip /></el-table
        ></el-tab-pane>
        <el-tab-pane :label="t('tenantCenter.tabs.games')" name="games"
          ><el-table
            :data="detail?.gameAuthorizations || []"
            row-key="id"
            max-height="560"
            ><el-table-column
              prop="name"
              :label="t('tenantCenter.fields.game')"
              min-width="180"
              ><template #default="{ row }"
                ><div class="game-cell">
                  <img
                    v-if="value(row, 'coverURL')"
                    :src="String(value(row, 'coverURL'))"
                    alt=""
                  /><span
                    ><strong>{{ value(row, "name") }}</strong
                    ><small>{{ value(row, "providerGameCode") }}</small></span
                  >
                </div></template
              ></el-table-column
            ><el-table-column
              prop="gameType"
              :label="t('tenantCenter.fields.gameType')"
              width="120" /><el-table-column
              prop="siteID"
              label="Site ID"
              width="95" /><el-table-column
              :label="t('common.status')"
              width="95"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="Boolean(value(row, 'platformEnabled'))"
                  disabled /></template></el-table-column
            ><el-table-column
              prop="configStatus"
              :label="t('tenantCenter.fields.publishStatus')"
              width="130" /><el-table-column
              prop="configVersion"
              :label="t('tenantCenter.fields.publishedVersion')"
              width="130" /></el-table
        ></el-tab-pane>
        <el-tab-pane :label="t('tenantCenter.tabs.assets')" name="assets"
          ><el-table :data="detail?.assets || []" row-key="id"
            ><el-table-column
              prop="displayName"
              :label="t('tenantCenter.fields.asset')"
              min-width="190"
              ><template #default="{ row }"
                ><strong>{{ value(row, "displayName") }}</strong
                ><small
                  >{{ value(row, "assetType") }} /
                  {{ value(row, "usageType") }}</small
                ></template
              ></el-table-column
            ><el-table-column
              prop="siteID"
              label="Site ID"
              width="100" /><el-table-column
              prop="fileSize"
              :label="t('tenantCenter.fields.fileSize')"
              width="120"
              ><template #default="{ row }"
                >{{
                  Math.round(Number(value(row, "fileSize") || 0) / 1024)
                }}
                KB</template
              ></el-table-column
            ><el-table-column
              :label="t('tenantCenter.fields.dimensions')"
              width="120"
              ><template #default="{ row }"
                >{{ value(row, "width") || "-" }} ×
                {{ value(row, "height") || "-" }}</template
              ></el-table-column
            ><el-table-column
              prop="referenceCount"
              :label="t('tenantCenter.fields.references')"
              width="120" /><el-table-column
              :label="t('common.status')"
              width="95"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="value(row, 'status') === 'ACTIVE'"
                  disabled /></template></el-table-column></el-table
        ></el-tab-pane>
        <el-tab-pane :label="t('tenantCenter.tabs.ledger')" name="ledger"
          ><el-alert
            v-if="platformBranding?.platformUSDTAddress"
            class="usdt-alert"
            type="info"
            :closable="false"
            :title="`${t('tenantCenter.fields.platformUSDT')}: ${platformBranding.platformUSDTAddress}`"
          /><el-table :data="detail?.ledgerAccounts || []" row-key="id"
            ><el-table-column
              prop="accountNo"
              :label="t('tenantCenter.fields.account')"
              min-width="180" /><el-table-column
              prop="siteID"
              label="Site ID"
              width="100" /><el-table-column
              prop="accountType"
              :label="t('tenantCenter.fields.accountType')"
              min-width="170"
              ><template #default="{ row }"
                ><el-tag effect="plain">{{
                  value(row, "accountType")
                }}</el-tag></template
              ></el-table-column
            ><el-table-column
              prop="currency"
              :label="t('tenantCenter.fields.currency')"
              width="95" /><el-table-column
              prop="balance"
              :label="t('tenantCenter.fields.balance')"
              min-width="140"
              align="right" /><el-table-column
              prop="availableBalance"
              :label="t('tenantCenter.fields.available')"
              min-width="140"
              align="right" /><el-table-column
              prop="frozenBalance"
              :label="t('tenantCenter.fields.frozen')"
              min-width="140"
              align="right"
              ><template #default="{ row }"
                ><strong class="frozen">{{
                  value(row, "frozenBalance")
                }}</strong></template
              ></el-table-column
            ><el-table-column :label="t('common.status')" width="95"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="value(row, 'status') === 'ACTIVE'"
                  disabled /></template></el-table-column></el-table
        ></el-tab-pane>
        <el-tab-pane :label="t('tenantCenter.tabs.audit')" name="audit"
          ><el-table :data="detail?.audit?.list || []" row-key="logID"
            ><el-table-column
              prop="operator"
              :label="t('tenantCenter.fields.operator')"
              width="140"
            /><el-table-column
              prop="action"
              :label="t('tenantCenter.fields.action')"
              min-width="230"
            /><el-table-column
              prop="targetType"
              :label="t('tenantCenter.fields.target')"
              width="150"
              ><template #default="{ row }"
                >{{ row.targetType }} #{{ row.targetID }}</template
              ></el-table-column
            ><el-table-column
              prop="result"
              :label="t('tenantCenter.fields.result')"
              width="105"
              ><template #default="{ row }"
                ><el-tag
                  :type="row.result === 'SUCCESS' ? 'success' : 'danger'"
                  >{{ row.result }}</el-tag
                ></template
              ></el-table-column
            ><el-table-column
              prop="reason"
              :label="t('tenantCenter.fields.reason')"
              min-width="180"
            /><el-table-column
              prop="createdAt"
              :label="t('common.createdAt')"
              width="175"
              ><template #default="{ row }">{{
                formatBackendTime(row.createdAt)
              }}</template></el-table-column
            ></el-table
          ></el-tab-pane
        >
      </el-tabs>
    </section>
  </main>
</template>
<style scoped>
.tenant-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 360px;
}
.detail-heading,
.heading-main,
.heading-actions {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}
.detail-heading {
  justify-content: space-between;
}
.heading-main p {
  margin: 0 0 4px;
  color: #337ecc;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}
.heading-main h1 {
  margin: 0;
  font-size: 24px;
  letter-spacing: 0;
}
.heading-main span {
  display: block;
  margin-top: 5px;
  color: #909399;
  font-size: 13px;
}
.heading-actions {
  align-items: center;
}
.overview-band,
.detail-band {
  padding: 18px 20px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
}
.overview-band dl {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin: 0;
}
.overview-band dl div {
  min-width: 0;
  padding: 8px 16px;
  border-right: 1px solid #ebeef5;
}
.overview-band dl div:nth-child(4n) {
  border-right: 0;
}
.overview-band dl div:nth-child(-n + 4) {
  border-bottom: 1px solid #ebeef5;
}
.overview-band dt {
  color: #909399;
  font-size: 12px;
}
.overview-band dd {
  margin: 5px 0 0;
  overflow-wrap: anywhere;
  font-weight: 600;
}
.detail-band {
  min-width: 0;
}
.detail-band small {
  display: block;
  margin-top: 3px;
  color: #909399;
}
.inline-name {
  margin-left: 8px;
}
.game-cell {
  display: flex;
  align-items: center;
  gap: 9px;
}
.game-cell img {
  width: 36px;
  height: 36px;
  flex: 0 0 auto;
  object-fit: cover;
  border-radius: 4px;
}
.frozen {
  color: #c45656;
}
.usdt-alert {
  margin-bottom: 14px;
}
@media (max-width: 900px) {
  .detail-heading {
    flex-direction: column;
  }
  .overview-band dl {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .overview-band dl div:nth-child(odd) {
    border-right: 1px solid #ebeef5;
  }
  .overview-band dl div:nth-child(even) {
    border-right: 0;
  }
  .overview-band dl div:nth-child(-n + 6) {
    border-bottom: 1px solid #ebeef5;
  }
}
</style>
