<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, View } from '@element-plus/icons-vue'
import { fetchAdminOptions, optionList, type AdminOptions } from '@/api/adminOptions'
import { operationsAPI, type OpsRecord } from '@/api/adminStability'
import { useAdminPermissions } from '@/composables/useAdminPermissions'
import { formatBackendTime } from '@/utils/timezone'
import { businessColumns, businessEntries, humanizeBusinessKey, sanitizeBusinessValue } from '@/utils/presentation'

const route = useRoute()
const { t, te } = useI18n()
const { can, loadPermissions } = useAdminPermissions()
const loading = ref(false)
const errorMessage = ref('')
const data = ref<Record<string, unknown>>({})
const detail = ref<Record<string, unknown>>({})
const detailVisible = ref(false)
const options = ref<AdminOptions>({})
const filters = reactive({ siteID: '', providerCode: '', channelCode: '', methodCode: '', country: '', currency: '', status: '', from: '', to: '' })
const diagnostic = reactive({ kind: 'DEPOSIT', orderNo: '', providerCode: '' })

const mode = computed(() => String(route.meta.moduleKey || 'payment-monitor'))
const paymentMode = computed(() => ['payment-monitor', 'support-payment-diagnostics'].includes(mode.value))
const diagnosticMode = computed(() => mode.value.startsWith('support-'))
const providerOptions = computed(() => optionList(options.value, 'providers'))
const channelOptions = computed(() => optionList(options.value, 'paymentChannels'))
const methodOptions = computed(() => optionList(options.value, 'paymentMethods'))
const countryOptions = computed(() => optionList(options.value, 'countries'))
const currencyOptions = computed(() => optionList(options.value, 'currencies'))
const statusOptions = computed(() => paymentMode.value
  ? [['1', 'PENDING'], ['2', 'SUCCESS'], ['3', 'FAILED'], ['4', 'CANCELLED'], ['5', 'EXPIRED'], ['6', 'PROCESSING']]
  : [['HEALTHY', 'HEALTHY'], ['DEGRADED', 'DEGRADED'], ['UNAVAILABLE', 'UNAVAILABLE'], ['MAINTENANCE', 'MAINTENANCE'], ['UNKNOWN', 'UNKNOWN']])
const title = computed(() => t(`opsV1.monitor.pages.${mode.value}`))
const healthRows = computed(() => paymentMode.value ? array(data.value.health) : array(data.value.list))
const metrics = computed(() => {
  if (!paymentMode.value) return []
  const deposit = object(data.value.deposit), withdrawal = object(data.value.withdrawal)
  return [
    ['depositCount', deposit.count], ['depositAmount', deposit.amount], ['depositRate', `${Number(deposit.successRate || 0).toFixed(1)}%`], ['depositPending', deposit.pendingCount], ['depositTimeout', deposit.timeoutCount],
    ['withdrawCount', withdrawal.count], ['withdrawAmount', withdrawal.amount], ['withdrawRate', `${Number(withdrawal.successRate || 0).toFixed(1)}%`], ['withdrawPending', withdrawal.pendingCount], ['withdrawTimeout', withdrawal.timeoutCount],
  ]
})

function params() {
  return { ...filters, page: 1, pageSize: 100 }
}

async function load() {
  loading.value = true
  errorMessage.value = ''
  try { data.value = paymentMode.value ? await operationsAPI.monitor.payments(params()) : await operationsAPI.monitor.providers(params()) }
  catch (error) { errorMessage.value = error instanceof Error ? error.message : t('opsV1.common.loadFailed') }
  finally { loading.value = false }
}

async function openPaymentDiagnostic() {
  if (!diagnostic.orderNo.trim()) return ElMessage.warning(t('opsV1.monitor.orderRequired'))
  detailVisible.value = true
  detail.value = {}
  try { detail.value = await operationsAPI.monitor.paymentDetailByNumber(diagnostic.kind, diagnostic.orderNo.trim()) }
  catch (error) { detailVisible.value = false; ElMessage.error(error instanceof Error ? error.message : t('opsV1.common.loadFailed')) }
}

async function openProvider(row?: OpsRecord) {
  const code = String(row?.providerCode || diagnostic.providerCode || filters.providerCode).trim()
  if (!code) return ElMessage.warning(t('opsV1.monitor.providerRequired'))
  detailVisible.value = true
  detail.value = {}
  try { detail.value = await operationsAPI.monitor.providerDetail(code, params()) }
  catch (error) { detailVisible.value = false; ElMessage.error(error instanceof Error ? error.message : t('opsV1.common.loadFailed')) }
}

async function markReview() {
  const order = object(detail.value.order)
  const id = Number(order.id)
  if (!id) return
  const reason = await askReason(t('opsV1.monitor.markReview'))
  if (!reason) return
  await operationsAPI.monitor.markReview(String(detail.value.orderType || diagnostic.kind), id, reason)
  ElMessage.success(t('opsV1.common.saved'))
}

async function addNote() {
  const order = object(detail.value.order)
  const id = Number(order.id)
  if (!id) return
  try {
    const note = await ElMessageBox.prompt(t('opsV1.monitor.notePrompt'), t('opsV1.monitor.addNote'), { inputPattern: /\S+/, inputErrorMessage: t('opsV1.common.required') })
    const reason = await askReason(t('opsV1.monitor.addNote'))
    if (!reason) return
    await operationsAPI.monitor.note({ siteID: Number(order.site_id || 0), targetType: `PAYMENT_${String(detail.value.orderType || diagnostic.kind)}`, targetID: String(id), note: note.value, reason })
    ElMessage.success(t('opsV1.common.saved'))
  } catch { /* cancelled */ }
}

async function askReason(titleText: string) {
  try { return (await ElMessageBox.prompt(t('opsV1.common.reasonPrompt'), titleText, { inputPattern: /\S+/, inputErrorMessage: t('opsV1.common.reasonRequired') })).value.trim() }
  catch { return '' }
}

function statusType(status: unknown) {
  const value = String(status || '').toUpperCase()
  if (['HEALTHY', 'SUCCESS', 'ACTIVE', 'VALID'].includes(value)) return 'success'
  if (['UNAVAILABLE', 'UNHEALTHY', 'FAILED', 'INVALID'].includes(value)) return 'danger'
  if (['DEGRADED', 'MAINTENANCE', 'PENDING'].includes(value)) return 'warning'
  return 'info'
}
function object(value: unknown) { return value && typeof value === 'object' && !Array.isArray(value) ? value as OpsRecord : {} }
function array(value: unknown) { return Array.isArray(value) ? value as OpsRecord[] : [] }
function display(value: unknown) {
  if (value && typeof value === 'object') return JSON.stringify(sanitizeBusinessValue(value))
  return value === null || value === undefined || value === '' ? '-' : String(value)
}
function statusLabel(value: unknown) {
  const status = String(value || 'UNKNOWN').toUpperCase()
  const key = `opsV1.status.${status}`
  return te(key) ? t(key) : status
}
function columns(value: unknown) { return businessColumns(value, 9) }
function reset() { Object.assign(filters, { siteID: '', providerCode: '', channelCode: '', methodCode: '', country: '', currency: '', status: '', from: '', to: '' }); void load() }

watch(() => route.meta.moduleKey, () => { data.value = {}; detail.value = {}; void load() })
onMounted(async () => { await loadPermissions(); options.value = await fetchAdminOptions().catch(() => ({})); await load() })
</script>

<template>
  <main class="monitor-page">
    <header class="page-heading"><div><p>{{ t('opsV1.monitor.eyebrow') }}</p><h1>{{ title }}</h1><span>{{ t('opsV1.monitor.description') }}</span></div><el-button circle :icon="Refresh" :loading="loading" @click="load" /></header>
    <section class="filter-band"><el-form inline @submit.prevent>

      <el-form-item v-if="!paymentMode" :label="t('opsV1.fields.provider')"><el-select v-model="filters.providerCode" clearable filterable><el-option v-for="item in providerOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item v-if="paymentMode" :label="t('opsV1.monitor.channel')"><el-select v-model="filters.channelCode" clearable filterable><el-option v-for="item in channelOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item v-if="paymentMode" :label="t('opsV1.monitor.method')"><el-select v-model="filters.methodCode" clearable filterable><el-option v-for="item in methodOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item :label="t('opsV1.fields.country')"><el-select v-model="filters.country" clearable filterable><el-option v-for="item in countryOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item :label="t('opsV1.fields.currency')"><el-select v-model="filters.currency" clearable filterable><el-option v-for="item in currencyOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item :label="t('opsV1.fields.status')"><el-select v-model="filters.status" clearable><el-option v-for="item in statusOptions" :key="item[0]" :label="t(`opsV1.status.${item[1]}`)" :value="item[0]" /></el-select></el-form-item>
      <el-form-item><el-date-picker v-model="filters.from" type="date" value-format="YYYY-MM-DD" :placeholder="t('opsV1.fields.from')" /></el-form-item>
      <el-form-item><el-date-picker v-model="filters.to" type="date" value-format="YYYY-MM-DD" :placeholder="t('opsV1.fields.to')" /></el-form-item>
      <el-form-item><el-button type="primary" :icon="Search" @click="load">{{ t('common.search') }}</el-button><el-button @click="reset">{{ t('common.reset') }}</el-button></el-form-item>
    </el-form></section>
    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

    <section v-if="diagnosticMode" class="diagnostic-band">
      <template v-if="paymentMode"><el-select v-model="diagnostic.kind"><el-option :label="t('opsV1.monitor.deposit')" value="DEPOSIT" /><el-option :label="t('opsV1.monitor.withdrawal')" value="WITHDRAWAL" /></el-select><el-input v-model="diagnostic.orderNo" :placeholder="t('opsV1.monitor.orderID')" @keyup.enter="openPaymentDiagnostic" /><el-button type="primary" :icon="Search" @click="openPaymentDiagnostic">{{ t('opsV1.monitor.diagnose') }}</el-button></template>
      <template v-else><el-input v-model="diagnostic.providerCode" :placeholder="t('opsV1.monitor.providerCode')" @keyup.enter="openProvider()" /><el-button type="primary" :icon="Search" @click="openProvider()">{{ t('opsV1.monitor.diagnose') }}</el-button></template>
    </section>

    <section v-if="paymentMode" class="metric-grid" v-loading="loading"><article v-for="item in metrics" :key="String(item[0])"><span>{{ t(`opsV1.monitor.metrics.${item[0]}`) }}</span><strong>{{ display(item[1]) }}</strong></article></section>

    <section class="table-band" v-loading="loading"><div class="section-heading"><h2>{{ paymentMode ? t('opsV1.monitor.paymentHealth') : t('opsV1.monitor.providerHealth') }}</h2><span>{{ data.from ? `${formatBackendTime(String(data.from))} - ${formatBackendTime(String(data.to || ''))}` : '' }}</span></div>
      <el-table :data="healthRows" row-key="providerCode" empty-text=" ">
        <el-table-column prop="providerCode" label="Provider" min-width="140"><template #default="{row}"><el-tag effect="plain">{{ row.providerCode || '-' }}</el-tag><small>{{ row.name || row.methodCode }}</small></template></el-table-column>
        <el-table-column v-if="paymentMode" prop="kind" :label="t('opsV1.monitor.kind')" width="110" />
        <el-table-column :label="t('opsV1.fields.status')" width="130"><template #default="{row}"><el-tag :type="statusType(row.operationalStatus || row.healthStatus)">{{ statusLabel(row.operationalStatus || row.healthStatus) }}</el-tag></template></el-table-column>
        <el-table-column v-if="paymentMode" prop="totalRequests" :label="t('opsV1.monitor.totalRequests')" width="120" align="right" />
        <el-table-column v-if="paymentMode" prop="successCount" :label="t('opsV1.monitor.successCount')" width="110" align="right" />
        <el-table-column v-if="paymentMode" prop="failedCount" :label="t('opsV1.monitor.failedCount')" width="110" align="right" />
        <el-table-column v-if="paymentMode" prop="pendingCount" :label="t('opsV1.monitor.pendingCount')" width="110" align="right" />
        <el-table-column v-if="paymentMode" prop="timeoutCount" :label="t('opsV1.monitor.timeoutCount')" width="110" align="right" />
        <el-table-column v-if="!paymentMode" prop="enabledGameCount" :label="t('opsV1.monitor.enabledGames')" width="120" align="right" />
        <el-table-column prop="successRate" :label="paymentMode?t('opsV1.monitor.successRate'):t('opsV1.monitor.launchRate')" width="120" align="right"><template #default="{row}">{{ Number(row.successRate ?? row.launchSuccessRate ?? 0).toFixed(1) }}%</template></el-table-column>
        <el-table-column v-if="!paymentMode" prop="callbackSuccessRate" :label="t('opsV1.monitor.callbackRate')" width="130" align="right"><template #default="{row}">{{ row.callbackSuccessRate===undefined?'-':`${Number(row.callbackSuccessRate).toFixed(1)}%` }}</template></el-table-column>
        <el-table-column prop="averageLatencyMS" :label="t('opsV1.monitor.avgLatency')" width="120" align="right"><template #default="{row}">{{ display(row.averageLatencyMS) }} ms</template></el-table-column>
        <el-table-column v-if="paymentMode" prop="p95LatencyMS" :label="t('opsV1.monitor.p95Latency')" width="120" align="right"><template #default="{row}">{{ display(row.p95LatencyMS) }} ms</template></el-table-column>
        <el-table-column prop="failureCount" :label="t('opsV1.monitor.failures')" width="100" align="right"><template #default="{row}">{{ row.failureCount ?? row.failedCount ?? 0 }}</template></el-table-column>
        <el-table-column v-if="!paymentMode" prop="retryQueueCount" :label="t('opsV1.monitor.retryQueue')" width="110" align="right" />
        <el-table-column :label="t('opsV1.monitor.lastSuccess')" width="175"><template #default="{row}">{{ formatBackendTime(String(row.lastSuccessfulTransaction || row.lastSuccessfulRequest || '')) }}</template></el-table-column>
        <el-table-column :label="t('opsV1.monitor.lastFailure')" width="175"><template #default="{row}">{{ formatBackendTime(String(row.lastFailedTransaction || row.lastFailedRequest || '')) }}</template></el-table-column>
        <el-table-column v-if="!paymentMode" width="70"><template #default="{row}"><el-button circle :icon="View" @click="openProvider(row)" /></template></el-table-column>
      </el-table>
      <el-empty v-if="!loading && !healthRows.length" :description="t('opsV1.common.empty')" />
    </section>

    <el-drawer v-model="detailVisible" :title="t('opsV1.monitor.diagnosticDetail')" size="min(980px, 96vw)">
      <template v-if="paymentMode">
        <div v-if="can('monitor:payment:review') || can('operations:note:write')" class="drawer-actions"><el-button v-if="can('monitor:payment:review')" type="warning" @click="markReview">{{ t('opsV1.monitor.markReview') }}</el-button><el-button v-if="can('operations:note:write')" @click="addNote">{{ t('opsV1.monitor.addNote') }}</el-button></div>
        <el-descriptions v-if="detail.order" :column="2" border><el-descriptions-item v-for="[key,value] in businessEntries(object(detail.order))" :key="key" :label="humanizeBusinessKey(key)">{{ display(value) }}</el-descriptions-item></el-descriptions>
        <el-tabs><el-tab-pane v-for="key in ['walletTransactions','callbacks','events','audit','notes']" :key="key" :label="t(`opsV1.monitor.tabs.${key}`)"><el-table :data="array(detail[key])"><el-table-column v-for="column in columns(detail[key])" :key="column" :prop="column" :label="humanizeBusinessKey(column)" min-width="140" show-overflow-tooltip /></el-table><el-empty v-if="!array(detail[key]).length" :description="t('opsV1.common.empty')" /></el-tab-pane></el-tabs>
      </template>
      <template v-else>
        <el-descriptions v-if="detail.health" :column="2" border><el-descriptions-item v-for="[key,value] in businessEntries(object(detail.health))" :key="key" :label="humanizeBusinessKey(key)">{{ display(value) }}</el-descriptions-item></el-descriptions>
        <el-tabs><el-tab-pane v-for="key in Object.keys(detail).filter(key=>Array.isArray(detail[key]))" :key="key" :label="t(`opsV1.monitor.tabs.${key}`)"><el-table :data="array(detail[key])"><el-table-column v-for="column in columns(detail[key])" :key="column" :prop="column" :label="humanizeBusinessKey(column)" min-width="140" show-overflow-tooltip /></el-table><el-empty v-if="!array(detail[key]).length" :description="t('opsV1.common.empty')" /></el-tab-pane></el-tabs>
      </template>
    </el-drawer>
  </main>
</template>

<style scoped>
.monitor-page{display:flex;flex-direction:column;gap:16px;min-width:0}.page-heading,.section-heading,.drawer-actions{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.page-heading p{margin:0 0 4px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.page-heading h1{margin:0;font-size:24px;letter-spacing:0}.page-heading span,.section-heading span{display:block;margin-top:7px;color:#909399;font-size:13px}.filter-band,.diagnostic-band,.table-band{padding:16px 18px;background:#fff;border:1px solid #e4e7ed;border-radius:4px}.filter-band :deep(.el-select),.filter-band :deep(.el-input){width:165px}.diagnostic-band{display:flex;gap:10px}.diagnostic-band .el-select{width:150px}.diagnostic-band .el-input{max-width:320px}.metric-grid{display:grid;grid-template-columns:repeat(5,minmax(0,1fr));border:1px solid #e4e7ed;background:#fff;border-radius:4px}.metric-grid article{padding:16px;border-right:1px solid #ebeef5;border-bottom:1px solid #ebeef5}.metric-grid span{display:block;color:#606266;font-size:12px}.metric-grid strong{display:block;margin-top:8px;font-size:21px;letter-spacing:0}.section-heading{align-items:center;margin-bottom:12px}.section-heading h2{margin:0;font-size:16px;letter-spacing:0}.table-band small{display:block;margin-top:3px;color:#909399}.table-band>.el-empty{padding:36px 0}.drawer-actions{justify-content:flex-end;margin-bottom:14px}@media(max-width:900px){.metric-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:700px){.page-heading{flex-direction:column}.filter-band :deep(.el-form-item),.filter-band :deep(.el-select),.filter-band :deep(.el-input){width:100%}.diagnostic-band{flex-wrap:wrap}.diagnostic-band .el-input{max-width:none}}
</style>
