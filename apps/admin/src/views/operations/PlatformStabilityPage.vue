<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Delete, Plus, Refresh, RefreshLeft, Search, View } from '@element-plus/icons-vue'
import { fetchAdminOptions, optionList, type AdminOptions } from '@/api/adminOptions'
import { operationIdempotencyKey, operationsAPI, type OpsRecord } from '@/api/adminStability'
import { useAdminPermissions } from '@/composables/useAdminPermissions'
import { formatBackendTime } from '@/utils/timezone'
import { businessColumns, businessEntries, humanizeBusinessKey, sanitizeBusinessValue } from '@/utils/presentation'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { actor, can, loadPermissions } = useAdminPermissions()
const loading = ref(false)
const errorMessage = ref('')
const rows = ref<OpsRecord[]>([])
const data = ref<Record<string, unknown>>({})
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const options = ref<AdminOptions>({})
const detailVisible = ref(false)
const detail = ref<Record<string, unknown>>({})
const subTab = ref('tasks')
const createVisible = ref(false)
const filters = reactive({ siteID: '', status: '', severity: '', environment: '', scopeType: '' })
const releaseAction = reactive({ siteID: '', fromID: '', toID: '' })
const backupForm = reactive({ scopeType: 'H5_CONFIGURATION', siteID: '', retentionDays: 30 })

const mode = computed(() => String(route.meta.moduleKey || 'system-health'))
const isPlatform = computed(() => actor.value?.platform === true)
const title = computed(() => t(`opsV1.stability.pages.${mode.value}`))
const siteOptions = computed(() => optionList(options.value, 'sites'))
const currentSiteName = computed(() => optionLabel(siteOptions.value, releaseAction.siteID || actor.value?.siteID))
const releaseOptions = computed(() => rows.value.map(row => ({
  value: String(row.id || ''),
  label: `v${row.version || '-'} · ${formatBackendTime(String(row.publishedAt || row.createdAt || ''))}`,
})).filter(item => item.value))
const pageMode = computed(() => ['alert-center', 'release-center', 'backup-center'].includes(mode.value))
const jobRows = computed(() => {
  if (subTab.value === 'tasks') return array(object(data.value.tasks).list)
  if (subTab.value === 'queueMetrics') return array(object(data.value.queueMetrics).queues)
  return array(data.value[subTab.value])
})
const statusOptions = computed(() => {
  if (mode.value === 'alert-center') return ['OPEN', 'ACKNOWLEDGED', 'RESOLVED']
  if (mode.value === 'release-center') return ['DRAFT', 'VALIDATING', 'READY', 'PUBLISHING', 'PUBLISHED', 'FAILED', 'ROLLED_BACK']
  return ['PENDING', 'RUNNING', 'SUCCESS', 'FAILED', 'EXPIRED', 'DELETED']
})

function params() { return { ...filters, page: page.value, pageSize: pageSize.value } }
async function load() {
  loading.value = true; errorMessage.value = ''
  try {
    if (mode.value === 'system-health') { data.value = await operationsAPI.health(); rows.value = array(data.value.components); total.value = rows.value.length }
    else if (mode.value === 'jobs-retries') { data.value = await operationsAPI.jobs.overview(params()); rows.value = []; total.value = 0 }
    else {
      const result = mode.value === 'alert-center' ? await operationsAPI.alerts.list(params()) : mode.value === 'release-center' ? await operationsAPI.releases.list(params()) : await operationsAPI.backups.list(params())
      rows.value = result.list || []; total.value = result.total || 0
    }
  } catch (error) { errorMessage.value = error instanceof Error ? error.message : t('opsV1.common.loadFailed') }
  finally { loading.value = false }
}

function array(value: unknown) { return Array.isArray(value) ? value as OpsRecord[] : [] }
function object(value: unknown) { return value && typeof value === 'object' && !Array.isArray(value) ? value as OpsRecord : {} }
function display(value: unknown) { return value === null || value === undefined || value === '' ? '-' : String(value) }
function optionLabel(items: Array<{ value: string | number; label: string }>, value: unknown) {
  return items.find(item => String(item.value) === String(value || ''))?.label || '-'
}
function businessDisplay(value: unknown) {
  if (value && typeof value === 'object') return JSON.stringify(sanitizeBusinessValue(value))
  return display(value)
}
function statusType(status: unknown) {
  const value = String(status || '').toUpperCase()
  if (['HEALTHY', 'SUCCESS', 'PUBLISHED', 'READY', 'RESOLVED'].includes(value)) return 'success'
  if (['UNHEALTHY', 'FAILED', 'DEAD', 'CRITICAL'].includes(value)) return 'danger'
  if (['DEGRADED', 'PENDING', 'RUNNING', 'PROCESSING', 'ACKNOWLEDGED', 'WARNING', 'PUBLISHING'].includes(value)) return 'warning'
  return 'info'
}

async function askReason(titleText: string) {
  try { return (await ElMessageBox.prompt(t('opsV1.common.reasonPrompt'), titleText, { inputPattern: /\S+/, inputErrorMessage: t('opsV1.common.reasonRequired') })).value.trim() }
  catch { return '' }
}

async function retryJob(row: OpsRecord, kind: 'execution' | 'outbox' | 'rabbit' = 'execution') {
  const reason = await askReason(t('opsV1.stability.retry'))
  if (!reason) return
  if (kind === 'outbox') await operationsAPI.jobs.retryDeadLetter(Number(row.id), reason)
  else if (kind === 'rabbit') await operationsAPI.jobs.retryRabbitDeadLetter(String(row.eventID || ''), String(row.queue || ''), reason)
  else await operationsAPI.jobs.retryExecution(Number(row.executionID || row.id), reason)
  ElMessage.success(t('opsV1.stability.retryQueued')); await load()
}

function viewRabbitPayload(row: OpsRecord) {
  detail.value = row
  detailVisible.value = true
}

async function purgeRetryQueue(row: OpsRecord) {
  const queue = String(row.queue || '')
  const reason = await askReason(t('opsV1.stability.purgeRetry'))
  if (!reason) return
  try { await ElMessageBox.confirm(t('opsV1.stability.purgeRetryConfirm', { queue }), t('opsV1.stability.purgeRetry'), { type: 'warning' }) }
  catch { return }
  const result = await operationsAPI.jobs.purgeRabbitRetry(queue, reason)
  ElMessage.success(t('opsV1.stability.purgedMessages', { count: result.purged })); await load()
}

async function setTaskStatus(row: OpsRecord, enabled: boolean) {
  const reason = await askReason(enabled ? t('opsV1.stability.resume') : t('opsV1.stability.pause'))
  if (!reason) return
  await operationsAPI.jobs.taskStatus(Number(row.id), enabled, reason)
  ElMessage.success(t('opsV1.common.saved')); await load()
}

async function runTask(row: OpsRecord) {
  const reason = await askReason(t('opsV1.stability.runNow'))
  if (!reason) return
  await operationsAPI.jobs.runTask(Number(row.id), reason)
  ElMessage.success(t('opsV1.stability.runQueued')); await load()
}

function openAlertDiagnostic(row: OpsRecord) {
  const type = String(row.alert_type || row.alertType || '').toUpperCase()
  if (type.includes('PAYMENT') || type.includes('WITHDRAW')) void router.push('/ops/payments')
  else if (type.includes('PROVIDER')) void router.push('/ops/providers')
  else if (type.includes('RELEASE')) void router.push('/ops/releases')
  else if (type.includes('BACKUP')) void router.push('/ops/backups')
  else void router.push('/ops/jobs')
}

async function alertAction(row: OpsRecord, action: 'acknowledge' | 'resolve') {
  const reason = await askReason(t(`opsV1.stability.${action}`))
  if (!reason) return
  if (action === 'acknowledge') await operationsAPI.alerts.acknowledge(Number(row.id), reason, '')
  else await operationsAPI.alerts.resolve(Number(row.id), reason, '')
  ElMessage.success(t('opsV1.common.saved')); await load()
}

async function openRelease(row: OpsRecord) {
  detailVisible.value = true; detail.value = {}
  try { detail.value = await operationsAPI.releases.detail(Number(row.id)) }
  catch (error) { detailVisible.value = false; ElMessage.error(error instanceof Error ? error.message : t('opsV1.common.loadFailed')) }
}

async function validateRelease() {
  if (!Number(releaseAction.siteID)) return ElMessage.warning(t('opsV1.stability.siteRequired'))
  detail.value = await operationsAPI.releases.validate(Number(releaseAction.siteID)); detailVisible.value = true
}

async function publishRelease() {
  if (!Number(releaseAction.siteID)) return ElMessage.warning(t('opsV1.stability.siteRequired'))
  const reason = await askReason(t('opsV1.stability.publish'))
  if (!reason) return
  try { await ElMessageBox.confirm(t('opsV1.stability.publishConfirm', { site: currentSiteName.value }), t('opsV1.stability.publish'), { type: 'warning' }) }
  catch { return }
  await operationsAPI.releases.publish(Number(releaseAction.siteID), reason, operationIdempotencyKey('release-publish')); ElMessage.success(t('opsV1.stability.published')); await load()
}

async function compareReleases() {
  if (!Number(releaseAction.fromID) || !Number(releaseAction.toID)) return ElMessage.warning(t('opsV1.stability.releasesRequired'))
  detail.value = await operationsAPI.releases.compare(Number(releaseAction.fromID), Number(releaseAction.toID)); detailVisible.value = true
}

async function rollback(row: OpsRecord) {
  const reason = await askReason(t('opsV1.stability.rollback'))
  if (!reason) return
  try { await ElMessageBox.confirm(t('opsV1.stability.rollbackConfirm', { version: row.version }), t('opsV1.stability.rollback'), { type: 'warning' }) }
  catch { return }
  await operationsAPI.releases.rollback(Number(row.id), reason, operationIdempotencyKey('release-rollback')); ElMessage.success(t('opsV1.stability.rollbackCreated')); detailVisible.value = false; await load()
}

async function openBackup(row: OpsRecord) {
  detailVisible.value = true; detail.value = {}
  try { detail.value = await operationsAPI.backups.detail(Number(row.id)) }
  catch (error) { detailVisible.value = false; ElMessage.error(error instanceof Error ? error.message : t('opsV1.common.loadFailed')) }
}

async function createBackup() {
  await operationsAPI.backups.create({ ...backupForm, siteID: Number(backupForm.siteID || 0), idempotencyKey: operationIdempotencyKey('backup') })
  createVisible.value = false; ElMessage.success(t('opsV1.stability.backupQueued')); await load()
}

async function restoreBackup(row: OpsRecord) {
  const id = Number(row.id)
  const validation = await operationsAPI.backups.validate(id)
  if (!validation.restorable) return ElMessage.warning(t('opsV1.stability.notRestorable'))
  const reason = await askReason(t('opsV1.stability.restore'))
  if (!reason) return
  try {
    const confirm = await ElMessageBox.prompt(t('opsV1.stability.restorePrompt', { phrase: validation.confirmationPhrase }), t('opsV1.stability.restore'), { inputPattern: new RegExp(`^${String(validation.confirmationPhrase).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}$`), inputErrorMessage: t('opsV1.stability.confirmationMismatch'), confirmButtonText: t('opsV1.stability.restore') })
    await operationsAPI.backups.restore(id, { reason, confirmation: confirm.value, dryRun: false })
    ElMessage.success(t('opsV1.stability.restored')); await load()
  } catch { /* cancelled */ }
}

async function deleteBackup(row: OpsRecord) {
  const reason = await askReason(t('opsV1.stability.deleteBackup'))
  if (!reason) return
  try { await ElMessageBox.confirm(t('opsV1.stability.deleteBackupConfirm', { backup: row.backupNo }), t('opsV1.stability.deleteBackup'), { type: 'warning' }) }
  catch { return }
  await operationsAPI.backups.remove(Number(row.id), reason); ElMessage.success(t('opsV1.common.saved')); await load()
}

function reset() { Object.assign(filters, { siteID: '', status: '', severity: '', environment: '', scopeType: '' }); page.value = 1; void load() }
watch(() => route.meta.moduleKey, () => { rows.value = []; data.value = {}; detail.value = {}; page.value = 1; void load() })
onMounted(async () => {
  await loadPermissions()
  if (!isPlatform.value) {
		subTab.value = 'retryRecords'
    const siteID = String(actor.value?.siteID || '')
    filters.siteID = siteID
    releaseAction.siteID = siteID
    backupForm.siteID = siteID
  }
  options.value = await fetchAdminOptions().catch(() => ({}))
  await load()
})
</script>

<template>
  <main class="stability-page">
    <header class="page-heading"><div><p>{{ t('opsV1.stability.eyebrow') }}</p><h1>{{ title }}</h1><span>{{ t('opsV1.stability.description') }}</span></div><div class="heading-actions"><el-button circle :icon="Refresh" :loading="loading" @click="load" /><el-button v-if="mode==='backup-center'&&can('backup:create')" type="primary" :icon="Plus" @click="createVisible=true">{{ t('opsV1.stability.createBackup') }}</el-button></div></header>

    <section v-if="pageMode" class="filter-band"><el-form inline>

      <el-form-item v-if="mode==='alert-center'" :label="t('opsV1.stability.severity')"><el-select v-model="filters.severity" clearable><el-option v-for="item in ['INFO','WARNING','CRITICAL']" :key="item" :label="t(`opsV1.status.${item}`)" :value="item" /></el-select></el-form-item>
      <el-form-item v-if="isPlatform&&mode!=='alert-center'" :label="t('opsV1.fields.siteID')"><el-select v-model="filters.siteID" clearable filterable><el-option v-for="item in siteOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item :label="t('opsV1.fields.status')"><el-select v-model="filters.status" clearable><el-option v-for="item in statusOptions" :key="item" :label="t(`opsV1.status.${item}`)" :value="item" /></el-select></el-form-item>
      <el-form-item v-if="mode==='release-center'" :label="t('opsV1.stability.environment')"><el-select v-model="filters.environment" clearable><el-option v-for="item in ['PRODUCTION','STAGING']" :key="item" :label="item" :value="item" /></el-select></el-form-item>
      <el-form-item v-if="mode==='backup-center'" :label="t('opsV1.stability.scope')"><el-select v-model="filters.scopeType" clearable><el-option v-for="item in ['DATABASE_CONFIGURATION','H5_CONFIGURATION','PUBLISHED_CONFIGURATION','ASSET_METADATA','ASSET_FILES']" :key="item" :label="item" :value="item" /></el-select></el-form-item>
      <el-form-item><el-button type="primary" :icon="Search" @click="page=1;load()">{{ t('common.search') }}</el-button><el-button @click="reset">{{ t('common.reset') }}</el-button></el-form-item>
    </el-form></section>

    <section v-if="mode==='release-center'" class="action-band"><el-select v-if="isPlatform" v-model="releaseAction.siteID" filterable :placeholder="t('opsV1.fields.siteID')"><el-option v-for="item in siteOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select><strong v-else class="site-name">{{ currentSiteName }}</strong><el-button :icon="Check" @click="validateRelease">{{ t('opsV1.stability.validate') }}</el-button><el-button v-if="can('release:publish')" type="primary" @click="publishRelease">{{ t('opsV1.stability.publish') }}</el-button><el-select v-model="releaseAction.fromID" clearable :placeholder="t('opsV1.stability.fromRelease')"><el-option v-for="item in releaseOptions" :key="`from-${item.value}`" :label="item.label" :value="item.value" /></el-select><el-select v-model="releaseAction.toID" clearable :placeholder="t('opsV1.stability.toRelease')"><el-option v-for="item in releaseOptions" :key="`to-${item.value}`" :label="item.label" :value="item.value" /></el-select><el-button @click="compareReleases">{{ t('opsV1.stability.compare') }}</el-button></section>
    <el-alert v-if="errorMessage" :title="errorMessage" type="error" :closable="false" show-icon />

    <section v-if="mode==='system-health'" class="health-summary"><el-tag size="large" :type="statusType(data.status)">{{ t(`opsV1.status.${data.status || 'UNKNOWN'}`) }}</el-tag><span>{{ t('opsV1.stability.lastChecked') }}: {{ formatBackendTime(String(data.checkedAt || '')) }}</span></section>
    <section v-if="mode==='jobs-retries'" class="job-tabs"><el-tabs v-model="subTab"><el-tab-pane v-if="isPlatform" :label="t('opsV1.stability.scheduledJobs')" name="tasks" /><el-tab-pane v-if="isPlatform" :label="t('opsV1.stability.failedExecutions')" name="failedExecutions" /><el-tab-pane :label="t('opsV1.stability.retryRecords')" name="retryRecords" /><el-tab-pane :label="t('opsV1.stability.deadLetters')" name="deadLetters" /><el-tab-pane v-if="isPlatform" :label="t('opsV1.stability.rabbitDeadLetters')" name="rabbitDeadLetters" /><el-tab-pane v-if="isPlatform" :label="t('opsV1.stability.queueMetrics')" name="queueMetrics" /></el-tabs></section>

    <section class="table-band" v-loading="loading"><el-table :data="mode==='jobs-retries'?jobRows:rows" row-key="id" empty-text=" ">
      <template v-if="mode==='system-health'"><el-table-column prop="componentCode" :label="t('opsV1.stability.component')" min-width="180"><template #default="{row}"><strong>{{ row.componentCode }}</strong><small>{{ row.componentType }}</small></template></el-table-column><el-table-column prop="status" :label="t('opsV1.fields.status')" width="130"><template #default="{row}"><el-tag :type="statusType(row.status)">{{ t(`opsV1.status.${row.status}`) }}</el-tag></template></el-table-column><el-table-column prop="latencyMS" :label="t('opsV1.monitor.avgLatency')" width="120" align="right"><template #default="{row}">{{ row.latencyMS }} ms</template></el-table-column><el-table-column prop="consecutiveFailures" :label="t('opsV1.stability.consecutiveFailures')" width="150" align="right" /><el-table-column prop="errorSummary" :label="t('opsV1.stability.errorSummary')" min-width="230" show-overflow-tooltip /><el-table-column prop="lastCheckedAt" :label="t('opsV1.stability.lastChecked')" width="175"><template #default="{row}">{{ formatBackendTime(row.lastCheckedAt) }}</template></el-table-column></template>
      <template v-else-if="mode==='jobs-retries'">
        <el-table-column :label="subTab==='queueMetrics'?t('opsV1.stability.queue'):t('opsV1.stability.task')" min-width="210"><template #default="{row}"><strong>{{ row.queue || row.taskCode || row.eventType || row.source_type }}</strong><small v-if="row.executionNo">{{ row.executionNo }}</small></template></el-table-column>
        <el-table-column :label="t('opsV1.fields.status')" width="140"><template #default="{row}"><el-switch v-if="subTab==='tasks'" :model-value="Boolean(row.enabled)" :disabled="!can('jobs:execute')||Boolean(row.running)" @change="setTaskStatus(row,Boolean($event))" /><span v-else-if="subTab==='queueMetrics'">{{ t('opsV1.stability.readyUnacked', { ready: row.ready, unacked: row.unacked }) }}</span><el-tag v-else :type="statusType(row.status || (subTab==='rabbitDeadLetters'?'DEAD':'UNKNOWN'))">{{ t(`opsV1.status.${row.status || (subTab==='rabbitDeadLetters'?'DEAD':'UNKNOWN')}`) }}</el-tag></template></el-table-column>
        <el-table-column :label="subTab==='queueMetrics'?t('opsV1.stability.depth'):t('opsV1.stability.attempts')" width="110" align="right"><template #default="{row}">{{ row.depth ?? row.retryCount ?? row.attempts ?? row.attempt_count ?? '-' }}</template></el-table-column>
        <el-table-column :label="subTab==='queueMetrics'?t('opsV1.stability.consumers'):t('opsV1.stability.errorSummary')" min-width="230" show-overflow-tooltip><template #default="{row}">{{ subTab==='queueMetrics' ? `${row.consumers ?? 0} / ${Number(row.consumeTps || 0).toFixed(2)} TPS` : (row.error || row.lastRunResult || row.errorMessage || row.lastError || row.last_error || '-') }}</template></el-table-column>
        <el-table-column :label="t('common.actions')" width="150"><template #default="{row}"><el-button v-if="can('jobs:retry')&&subTab==='failedExecutions'" circle :icon="RefreshLeft" @click="retryJob(row,'execution')" /><el-button v-if="can('jobs:retry')&&subTab==='deadLetters'" circle :icon="RefreshLeft" @click="retryJob(row,'outbox')" /><template v-if="subTab==='rabbitDeadLetters'"><el-button circle :icon="View" @click="viewRabbitPayload(row)" /><el-button v-if="can('jobs:retry')" circle :icon="RefreshLeft" @click="retryJob(row,'rabbit')" /></template><el-button v-if="can('jobs:retry')&&subTab==='queueMetrics'&&!['retry.events','dead.events'].includes(String(row.queue||''))&&!String(row.queue||'').includes('.retry.')&&!String(row.queue||'').endsWith('.dead')" circle type="danger" :icon="Delete" @click="purgeRetryQueue(row)" /><el-button v-if="can('jobs:execute')&&subTab==='tasks'" size="small" :disabled="!row.enabled||row.running" @click="runTask(row)">{{ t('opsV1.stability.runNow') }}</el-button></template></el-table-column>
      </template>
      <template v-else-if="mode==='alert-center'"><el-table-column prop="severity" :label="t('opsV1.stability.severity')" width="110"><template #default="{row}"><el-tag :type="statusType(row.severity)">{{ t(`opsV1.status.${row.severity}`) }}</el-tag></template></el-table-column><el-table-column prop="summary" :label="t('opsV1.stability.summary')" min-width="240" /><el-table-column prop="source" :label="t('opsV1.stability.source')" width="160" /><el-table-column prop="occurrence_count" :label="t('opsV1.stability.occurrences')" width="110" align="right" /><el-table-column prop="status" :label="t('opsV1.fields.status')" width="130"><template #default="{row}"><el-tag :type="statusType(row.status)">{{ t(`opsV1.status.${row.status}`) }}</el-tag></template></el-table-column><el-table-column prop="last_triggered_at" :label="t('opsV1.stability.lastTriggered')" width="175"><template #default="{row}">{{ formatBackendTime(row.last_triggered_at) }}</template></el-table-column><el-table-column :label="t('common.actions')" width="220"><template #default="{row}"><el-button circle :icon="View" @click="openAlertDiagnostic(row)" /><el-button v-if="can('alerts:acknowledge')&&row.status==='OPEN'" size="small" @click="alertAction(row,'acknowledge')">{{ t('opsV1.stability.acknowledge') }}</el-button><el-button v-if="can('alerts:resolve')&&row.status!=='RESOLVED'" size="small" type="success" @click="alertAction(row,'resolve')">{{ t('opsV1.stability.resolve') }}</el-button></template></el-table-column></template>
      <template v-else-if="mode==='release-center'"><el-table-column prop="version" :label="t('opsV1.stability.version')" width="100"><template #default="{row}"><strong>v{{ row.version }}</strong></template></el-table-column><el-table-column v-if="isPlatform" :label="t('opsV1.fields.siteID')" min-width="150"><template #default="{row}">{{ optionLabel(siteOptions,row.siteID) }}</template></el-table-column><el-table-column prop="modules" :label="t('opsV1.stability.modules')" min-width="210" show-overflow-tooltip /><el-table-column prop="status" :label="t('opsV1.fields.status')" width="130"><template #default="{row}"><el-tag :type="statusType(row.status)">{{ t(`opsV1.status.${row.status}`) }}</el-tag></template></el-table-column><el-table-column prop="publishedAt" :label="t('opsV1.stability.publishedAt')" width="175"><template #default="{row}">{{ formatBackendTime(row.publishedAt) }}</template></el-table-column><el-table-column width="70"><template #default="{row}"><el-button circle :icon="View" @click="openRelease(row)" /></template></el-table-column></template>
      <template v-else><el-table-column prop="backupNo" :label="t('opsV1.stability.backupNo')" min-width="220" /><el-table-column prop="scopeType" :label="t('opsV1.stability.scope')" min-width="210"><template #default="{row}">{{ humanizeBusinessKey(String(row.scopeType || '')) }}</template></el-table-column><el-table-column prop="status" :label="t('opsV1.fields.status')" width="120"><template #default="{row}"><el-tag :type="statusType(row.status)">{{ t(`opsV1.status.${row.status}`) }}</el-tag></template></el-table-column><el-table-column prop="sizeBytes" :label="t('opsV1.stability.size')" width="120" align="right" /><el-table-column prop="checksumSha256" :label="t('opsV1.stability.checksum')" min-width="180" show-overflow-tooltip /><el-table-column prop="encrypted" :label="t('opsV1.stability.encrypted')" width="100"><template #default="{row}"><el-switch :model-value="Boolean(row.encrypted)" disabled /></template></el-table-column><el-table-column prop="completedAt" :label="t('opsV1.stability.completedAt')" width="175"><template #default="{row}">{{ formatBackendTime(row.completedAt) }}</template></el-table-column><el-table-column :label="t('common.actions')" width="170"><template #default="{row}"><el-button circle :icon="View" @click="openBackup(row)" /><el-button v-if="can('backup:restore')&&row.status==='SUCCESS'" circle :icon="RefreshLeft" @click="restoreBackup(row)" /><el-button v-if="can('backup:delete')&&!['PENDING','RUNNING'].includes(String(row.status))" circle type="danger" :icon="Delete" @click="deleteBackup(row)" /></template></el-table-column></template>
    </el-table><el-empty v-if="!loading && !(mode==='jobs-retries'?jobRows:rows).length" :description="t('opsV1.common.empty')" /><el-pagination v-if="pageMode&&total" v-model:current-page="page" v-model:page-size="pageSize" background layout="total, sizes, prev, pager, next" :total="total" @change="load" /></section>

    <el-drawer v-model="detailVisible" :title="t('common.detail')" size="min(980px, 96vw)"><template v-if="mode==='release-center'&&detail.release"><div class="drawer-actions"><el-button v-if="can('release:rollback')&&['PUBLISHED','ROLLED_BACK'].includes(String(object(detail.release).status))" type="warning" @click="rollback(object(detail.release))">{{ t('opsV1.stability.rollback') }}</el-button></div><el-descriptions :column="2" border><el-descriptions-item v-for="[key,value] in businessEntries(object(detail.release))" :key="key" :label="humanizeBusinessKey(key)">{{ businessDisplay(value) }}</el-descriptions-item></el-descriptions><h3>{{ t('opsV1.stability.audit') }}</h3><el-table :data="array(detail.audit)"><el-table-column v-for="column in businessColumns(array(detail.audit))" :key="column" :prop="column" :label="humanizeBusinessKey(column)" min-width="140" /></el-table></template><el-descriptions v-else :column="1" border><el-descriptions-item v-for="[key,value] in businessEntries(detail)" :key="key" :label="humanizeBusinessKey(key)">{{ businessDisplay(value) }}</el-descriptions-item></el-descriptions></el-drawer>

    <el-dialog v-model="createVisible" :title="t('opsV1.stability.createBackup')" width="min(620px, 94vw)"><el-form label-position="top"><el-form-item :label="t('opsV1.stability.scope')"><el-select v-model="backupForm.scopeType"><el-option v-for="item in ['DATABASE_CONFIGURATION','H5_CONFIGURATION','PUBLISHED_CONFIGURATION','ASSET_METADATA','ASSET_FILES']" :key="item" :label="humanizeBusinessKey(item)" :value="item" /></el-select></el-form-item><el-form-item v-if="isPlatform" :label="t('opsV1.fields.siteID')"><el-select v-model="backupForm.siteID" clearable filterable><el-option v-for="item in siteOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item><el-form-item :label="t('opsV1.stability.retentionDays')"><el-input-number v-model="backupForm.retentionDays" :min="1" :max="365" /></el-form-item></el-form><template #footer><el-button @click="createVisible=false">{{ t('common.cancel') }}</el-button><el-button type="primary" @click="createBackup">{{ t('opsV1.stability.createBackup') }}</el-button></template></el-dialog>
  </main>
</template>

<style scoped>
.stability-page{display:flex;flex-direction:column;gap:14px;min-width:0}.page-heading,.heading-actions,.section-heading,.drawer-actions{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.page-heading p{margin:0 0 4px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.page-heading h1{margin:0;font-size:24px;letter-spacing:0}.page-heading span{display:block;margin-top:7px;color:#909399;font-size:13px}.filter-band,.action-band,.health-summary,.job-tabs,.table-band{padding:16px 18px;background:#fff;border:1px solid #e4e7ed;border-radius:4px}.filter-band :deep(.el-select),.filter-band :deep(.el-input){width:190px}.action-band{display:flex;flex-wrap:wrap;gap:10px}.action-band>.el-select{width:190px}.action-band .el-input{width:150px}.health-summary{display:flex;align-items:center;gap:14px}.health-summary span{color:#606266}.job-tabs{padding-bottom:0}.table-band small{display:block;margin-top:3px;color:#909399}.table-band>.el-empty{padding:36px 0}.table-band :deep(.el-pagination){justify-content:flex-end;margin-top:16px}.drawer-actions{justify-content:flex-end;margin-bottom:14px}pre{white-space:pre-wrap;word-break:break-word;background:#f5f7fa;padding:14px;border-radius:4px}@media(max-width:700px){.page-heading{flex-direction:column}.filter-band :deep(.el-form-item),.filter-band :deep(.el-select),.filter-band :deep(.el-input){width:100%}.action-band>.el-select,.action-band .el-input{width:100%}}
</style>
