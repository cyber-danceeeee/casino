<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Calendar, Connection, Refresh } from '@element-plus/icons-vue'
import { systemAPI, type TenantDomainSummary } from '@/api/adminSystem'
import { formatBackendTime } from '@/utils/timezone'
import { useAdminPermissions } from '@/composables/useAdminPermissions'

const { t } = useI18n(), { can } = useAdminPermissions()
const loading = ref(false), rows = ref<TenantDomainSummary[]>([]), total = ref(0), page = ref(1), pageSize = ref(20)
const verifyingID = ref(0), sslVisible = ref(false), sslSaving = ref(false), sslRow = ref<TenantDomainSummary>()
const sslExpiry = ref<Date | null>(null), sslReason = ref('')
const query = reactive({ tenantID: '', domain: '', domainType: '', lifecycleStatus: '' })

async function load() {
  loading.value = true
  try { const result = await systemAPI.domains.list({ ...query, page: page.value, pageSize: pageSize.value }); rows.value = result.list || []; total.value = result.total || 0 }
  catch (error) { ElMessage.error(error instanceof Error ? error.message : t('tenantCenter.messages.loadFailed')) }
  finally { loading.value = false }
}
async function change(row: TenantDomainSummary) {
  const next = row.enabled
  try { const { value } = await ElMessageBox.prompt(next ? t('tenantCenter.confirm.enableDomain') : t('tenantCenter.confirm.disableDomain'), t('tenantCenter.fields.reason'), { inputPattern: /\S+/, inputErrorMessage: t('tenantCenter.messages.reasonRequired') }); await systemAPI.domains.status(row.id, next, value); ElMessage.success(t('tenantCenter.messages.statusUpdated')); await load() }
  catch { row.enabled = !next }
}
async function verify(row: TenantDomainSummary) {
  verifyingID.value = row.id
  try { await systemAPI.domains.verify(row.id); ElMessage.success(t('tenantCenter.domains.verified')); await load() }
  catch (error) { ElMessage.error(error instanceof Error ? error.message : t('tenantCenter.domains.verifyFailed')); await load() }
  finally { verifyingID.value = 0 }
}
function editSSL(row: TenantDomainSummary) {
  sslRow.value = row
  sslExpiry.value = row.sslExpiresAt ? new Date(row.sslExpiresAt) : null
  sslReason.value = ''
  sslVisible.value = true
}
async function saveSSL() {
  if (!sslRow.value || !sslReason.value.trim()) return ElMessage.warning(t('tenantCenter.messages.reasonRequired'))
  sslSaving.value = true
  try { await systemAPI.domains.updateSSL(sslRow.value.id, sslExpiry.value?.toISOString() || null, sslReason.value.trim()); sslVisible.value = false; ElMessage.success(t('tenantCenter.domains.sslSaved')); await load() }
  catch (error) { ElMessage.error(error instanceof Error ? error.message : t('tenantCenter.messages.saveFailed')) }
  finally { sslSaving.value = false }
}
onMounted(load)
</script>

<template>
  <main class="platform-page">
    <header class="page-heading"><div><p>{{ t('portal.platform.tenants') }}</p><h1>{{ t('tenantCenter.domains.title') }}</h1><span>{{ t('tenantCenter.domains.description') }}</span></div><el-button :icon="Refresh" @click="load">{{ t('common.refresh') }}</el-button></header>
    <el-alert type="info" :closable="false" show-icon :title="t('tenantCenter.domains.verifyHelp')"/>
    <section class="filter-band"><el-form inline><el-form-item :label="t('tenantCenter.fields.tenantID')"><el-input v-model="query.tenantID" clearable/></el-form-item><el-form-item :label="t('tenantCenter.fields.domain')"><el-input v-model="query.domain" clearable/></el-form-item><el-form-item :label="t('tenantCenter.fields.domainType')"><el-select v-model="query.domainType" clearable><el-option v-for="value in ['PRIMARY','H5','ADMIN','API']" :key="value" :label="value" :value="value"/></el-select></el-form-item><el-form-item :label="t('common.status')"><el-select v-model="query.lifecycleStatus" clearable><el-option v-for="value in ['PENDING','ACTIVE','DISABLED','FAILED']" :key="value" :label="t(`tenantCenter.domainStatus.${value}`)" :value="value"/></el-select></el-form-item><el-form-item><el-button type="primary" @click="page=1;load()">{{ t('common.search') }}</el-button></el-form-item></el-form></section>
    <section class="table-band"><el-table :data="rows" v-loading="loading" row-key="id">
      <el-table-column prop="tenant" :label="t('tenantCenter.fields.tenantName')" min-width="150"/><el-table-column prop="site" :label="t('tenantCenter.fields.site')" min-width="140"/>
      <el-table-column prop="domain" :label="t('tenantCenter.fields.domain')" min-width="220"><template #default="{row}"><a :href="`https://${row.domain}`" target="_blank" rel="noreferrer">{{ row.domain }}</a><small v-if="row.verificationError">{{ row.verificationError }}</small></template></el-table-column>
      <el-table-column prop="domainType" :label="t('tenantCenter.fields.domainType')" width="105"><template #default="{row}"><el-tag effect="plain">{{ row.domainType }}</el-tag></template></el-table-column>
      <el-table-column :label="t('common.status')" width="100"><template #default="{row}"><el-switch v-model="row.enabled" :disabled="!can('system:domain:status')" @change="change(row)"/></template></el-table-column>
      <el-table-column prop="verificationStatus" :label="t('tenantCenter.fields.verification')" width="135"><template #default="{row}"><el-tag :type="row.verificationStatus==='VERIFIED'?'success':row.verificationStatus==='FAILED'?'danger':'warning'">{{ t(`tenantCenter.verificationStatus.${row.verificationStatus||'PENDING'}`) }}</el-tag></template></el-table-column>
      <el-table-column prop="sslStatus" label="SSL" width="115"><template #default="{row}"><el-tag :type="row.sslStatus==='ACTIVE'?'success':['FAILED','EXPIRED'].includes(row.sslStatus)?'danger':'warning'">{{ t(`tenantCenter.sslStatus.${row.sslStatus||'PENDING'}`) }}</el-tag></template></el-table-column>
      <el-table-column prop="sslExpiresAt" :label="t('tenantCenter.fields.sslExpiry')" width="175"><template #default="{row}">{{ row.sslExpiresAt?formatBackendTime(row.sslExpiresAt):'-' }}</template></el-table-column>
      <el-table-column prop="lastVerifiedAt" :label="t('tenantCenter.fields.lastVerified')" width="175"><template #default="{row}">{{ row.lastVerifiedAt?formatBackendTime(row.lastVerifiedAt):'-' }}</template></el-table-column>
      <el-table-column :label="t('common.actions')" width="120" fixed="right"><template #default="{row}"><el-button circle type="primary" plain :icon="Connection" :loading="verifyingID===row.id" :title="t('tenantCenter.domains.verify')" @click="verify(row)"/><el-button circle :icon="Calendar" :title="t('tenantCenter.domains.editSSL')" @click="editSSL(row)"/></template></el-table-column>
    </el-table><el-pagination v-model:current-page="page" v-model:page-size="pageSize" background layout="total, sizes, prev, pager, next" :total="total" @change="load"/></section>

    <el-dialog v-model="sslVisible" :title="t('tenantCenter.domains.editSSL')" width="480px"><el-form label-position="top"><el-form-item :label="t('tenantCenter.fields.sslExpiry')"><el-date-picker v-model="sslExpiry" type="datetime" style="width:100%" clearable/></el-form-item><el-form-item :label="t('tenantCenter.fields.reason')" required><el-input v-model="sslReason" type="textarea" :rows="2"/></el-form-item></el-form><template #footer><el-button @click="sslVisible=false">{{ t('common.cancel') }}</el-button><el-button type="primary" :loading="sslSaving" @click="saveSSL">{{ t('common.confirm') }}</el-button></template></el-dialog>
  </main>
</template>

<style scoped>.platform-page{display:flex;flex-direction:column;gap:16px}.page-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:16px}.page-heading p{margin:0 0 4px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.page-heading h1{margin:0;font-size:24px;letter-spacing:0}.page-heading span{display:block;margin-top:7px;color:#909399;font-size:13px}.filter-band,.table-band{padding:18px 20px;background:#fff;border:1px solid #e4e7ed;border-radius:4px}.filter-band :deep(.el-form-item){margin-bottom:0}.filter-band :deep(.el-input),.filter-band :deep(.el-select){width:170px}.table-band a{color:#337ecc;text-decoration:none}.table-band small{display:block;margin-top:4px;color:#c45656;line-height:1.35}.table-band .el-pagination{justify-content:flex-end;margin-top:16px}</style>
