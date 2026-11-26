<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Plus, Refresh, Search } from '@element-plus/icons-vue'
import { fetchAdminOptions, optionList, type AdminOptions } from '@/api/adminOptions'
import { operationsAPI, type OpsRecord } from '@/api/adminStability'
import { useAdminPermissions } from '@/composables/useAdminPermissions'
import { formatBackendTime } from '@/utils/timezone'

const { t } = useI18n()
const { actor, can, loadPermissions } = useAdminPermissions()
const activeTab = ref<'SITE' | 'PROVIDER' | 'GAME'>('SITE')
const loading = ref(false)
const errorMessage = ref('')
const rows = ref<OpsRecord[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const options = ref<AdminOptions>({})
const dialogVisible = ref(false)
const filters = reactive({ siteID: '', providerCode: '', enabled: '' })
const form = reactive<Record<string, unknown>>({})

const isPlatform = computed(() => actor.value?.platform === true)
const siteOptions = computed(() => optionList(options.value, 'sites'))
const providerOptions = computed(() => optionList(options.value, 'providers'))
const gameOptions = computed(() => optionList(options.value, 'games'))
const writePermission = computed(() => `maintenance:${activeTab.value.toLowerCase()}:write`)

function baseForm() {
  return { siteID: Number(filters.siteID || actor.value?.siteID || 0),
    scopeType: 'SITE', providerCode: '', gameID: 0, enabled: true, startAt: null, endAt: null,
    reason: '', title: '', messageZh: '', messageEn: '', estimatedRecoveryAt: null,
    supportURL: '', supportEmail: '', supportChat: '', allowlistedIPsText: '',
    allowAdminBypass: true,
    affectedCountriesText: '', affectedCurrenciesText: '', allowExistingSessions: true, hideFromLobby: true, blockNewLaunches: true,
  }
}

function params() { return { ...filters, page: page.value, pageSize: pageSize.value } }

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = activeTab.value === 'SITE' ? await operationsAPI.maintenance.site(params()) : activeTab.value === 'PROVIDER' ? await operationsAPI.maintenance.providers(params()) : await operationsAPI.maintenance.games(params())
    rows.value = result.list || []
    total.value = result.total || 0
  } catch (error) { errorMessage.value = error instanceof Error ? error.message : t('opsV1.common.loadFailed') }
  finally { loading.value = false }
}

function field(row: OpsRecord, camel: string, snake?: string) { return row[camel] ?? row[snake || camel] }
function listText(value: unknown) { return Array.isArray(value) ? value.join(',') : String(value || '') }
function bool(value: unknown) { return value === true || value === 1 || value === '1' }
function optionLabel(items: Array<{ value: string | number; label: string }>, value: unknown) {
  return items.find(item => String(item.value) === String(value || ''))?.label || '-'
}

function openCreate() { Object.assign(form, baseForm()); dialogVisible.value = true }
function openEdit(row: OpsRecord) {
  const initial = baseForm()
  Object.assign(form, initial, { siteID: Number(field(row, 'siteID', 'site_id') || 0),
    scopeType: field(row, 'scopeType', 'scope_type') || 'SITE', providerCode: field(row, 'providerCode', 'provider_code') || '', gameID: Number(field(row, 'gameID', 'game_id') || 0),
    enabled: bool(row.enabled), startAt: field(row, 'startAt', 'start_at') || null, endAt: field(row, 'endAt', 'end_at') || null,
    reason: '', title: row.title || '', messageZh: field(row, 'messageZh', 'message_zh') || '', messageEn: field(row, 'messageEn', 'message_en') || '',
    estimatedRecoveryAt: field(row, 'estimatedRecoveryAt', 'estimated_recovery_at') || null,
    supportURL: field(row, 'supportURL', 'support_url') || '', supportEmail: field(row, 'supportEmail', 'support_email') || '', supportChat: field(row, 'supportChat', 'support_chat') || '',
    allowlistedIPsText: listText(field(row, 'allowlistedIPs', 'allowlisted_ips')),
    allowAdminBypass: bool(field(row, 'allowAdminBypass', 'allow_admin_bypass')),
    affectedCountriesText: listText(field(row, 'affectedCountries', 'affected_countries')), affectedCurrenciesText: listText(field(row, 'affectedCurrencies', 'affected_currencies')),
    allowExistingSessions: bool(field(row, 'allowExistingSessions', 'allow_existing_sessions')), hideFromLobby: bool(field(row, 'hideFromLobby', 'hide_from_lobby')), blockNewLaunches: bool(field(row, 'blockNewLaunches', 'block_new_launches')),
  })
  dialogVisible.value = true
}

function csv(value: unknown) { return String(value || '').split(',').map(item => item.trim()).filter(Boolean) }
async function save() {
  if (!String(form.reason || '').trim()) return ElMessage.warning(t('opsV1.common.reasonRequired'))
  const payload: Record<string, unknown> = {
    ...form, siteID: Number(form.siteID || 0), gameID: Number(form.gameID || 0),
    allowlistedIPs: csv(form.allowlistedIPsText), affectedCountries: csv(form.affectedCountriesText), affectedCurrencies: csv(form.affectedCurrenciesText),
  }
  delete payload.allowlistedIPsText; delete payload.affectedCountriesText; delete payload.affectedCurrenciesText
  try {
    await ElMessageBox.confirm(
      t('opsV1.maintenance.confirmSave', { target: activeTab.value, enabled: form.enabled ? t('common.enable') : t('common.disable') }),
      t('opsV1.maintenance.confirmTitle'),
      { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') },
    )
  } catch { return }
  if (activeTab.value === 'SITE') await operationsAPI.maintenance.saveSite(payload)
  else if (activeTab.value === 'PROVIDER') await operationsAPI.maintenance.saveProvider(payload)
  else await operationsAPI.maintenance.saveGame(payload)
  dialogVisible.value = false
  ElMessage.success(t('opsV1.common.saved'))
  await load()
}

function changeTab() { page.value = 1; rows.value = []; void load() }
function reset() { Object.assign(filters, { siteID: '', providerCode: '', enabled: '' }); page.value = 1; void load() }

onMounted(async () => { await loadPermissions(); options.value = await fetchAdminOptions().catch(() => ({})); await load() })
</script>

<template>
  <main class="maintenance-page">
    <header class="page-heading"><div><p>{{ t('opsV1.maintenance.eyebrow') }}</p><h1>{{ t('opsV1.maintenance.title') }}</h1><span>{{ t('opsV1.maintenance.description') }}</span></div><div><el-button circle :icon="Refresh" @click="load" /><el-button v-if="can(writePermission)" type="primary" :icon="Plus" @click="openCreate">{{ t('common.create') }}</el-button></div></header>
    <el-alert type="warning" show-icon :closable="false" :title="t('opsV1.maintenance.callbackNotice')" />
    <el-tabs v-model="activeTab" class="mode-tabs" @tab-change="changeTab"><el-tab-pane :label="t('opsV1.maintenance.tabs.site')" name="SITE" /><el-tab-pane :label="t('opsV1.maintenance.tabs.provider')" name="PROVIDER" /><el-tab-pane :label="t('opsV1.maintenance.tabs.game')" name="GAME" /></el-tabs>
    <section class="filter-band"><el-form inline>

      <el-form-item v-if="isPlatform" :label="t('opsV1.fields.siteID')"><el-select v-model="filters.siteID" clearable filterable><el-option v-for="item in siteOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item v-if="activeTab!=='SITE'" :label="t('opsV1.fields.provider')"><el-select v-model="filters.providerCode" clearable filterable><el-option v-for="item in providerOptions" :key="item.value" :label="item.label" :value="String(item.value)" /></el-select></el-form-item>
      <el-form-item :label="t('opsV1.fields.status')"><el-select v-model="filters.enabled" clearable><el-option :label="t('common.enable')" value="1" /><el-option :label="t('common.disable')" value="0" /></el-select></el-form-item>
      <el-form-item><el-button type="primary" :icon="Search" @click="page=1;load()">{{ t('common.search') }}</el-button><el-button @click="reset">{{ t('common.reset') }}</el-button></el-form-item>
    </el-form></section>
    <el-alert v-if="errorMessage" :title="errorMessage" type="error" :closable="false" show-icon />
    <section class="table-band" v-loading="loading"><el-table :data="rows" row-key="id" empty-text=" ">

      <el-table-column v-if="isPlatform" :label="t('opsV1.fields.siteID')" min-width="150"><template #default="{row}">{{ optionLabel(siteOptions, field(row,'siteID','site_id')) }}</template></el-table-column>
      <el-table-column v-if="activeTab==='SITE'" :label="t('opsV1.maintenance.scope')" width="110"><template #default="{row}"><el-tag effect="plain">{{ field(row,'scopeType','scope_type') }}</el-tag></template></el-table-column>
      <el-table-column v-if="activeTab==='PROVIDER'" :label="t('opsV1.fields.provider')" min-width="140"><template #default="{row}"><el-tag>{{ field(row,'providerCode','provider_code') }}</el-tag></template></el-table-column>
      <el-table-column v-if="activeTab==='GAME'" :label="t('opsV1.fields.game')" min-width="180"><template #default="{row}"><strong>{{ field(row,'gameName','game_name') || '-' }}</strong><small>{{ field(row,'gameCode','game_code') || '-' }} · {{ field(row,'providerCode','provider_code') }}</small></template></el-table-column>
      <el-table-column :label="t('opsV1.fields.status')" width="100"><template #default="{row}"><el-switch :model-value="bool(row.enabled)" disabled /></template></el-table-column>
      <el-table-column prop="reason" :label="t('opsV1.fields.reason')" min-width="190" show-overflow-tooltip />
      <el-table-column :label="t('opsV1.maintenance.window')" min-width="250"><template #default="{row}">{{ formatBackendTime(String(field(row,'startAt','start_at') || '')) }}<span class="window-separator">→</span>{{ formatBackendTime(String(field(row,'endAt','end_at') || '')) }}</template></el-table-column>
      <el-table-column :label="t('opsV1.maintenance.blockLaunch')" width="130"><template #default="{row}"><el-switch :model-value="bool(field(row,'blockNewLaunches','block_new_launches'))" disabled /></template></el-table-column>
      <el-table-column width="70"><template #default="{row}"><el-button v-if="can(writePermission)" circle :icon="Edit" @click="openEdit(row)" /></template></el-table-column>
    </el-table><el-empty v-if="!loading&&!rows.length" :description="t('opsV1.common.empty')" /><el-pagination v-if="total" v-model:current-page="page" v-model:page-size="pageSize" background layout="total, sizes, prev, pager, next" :total="total" @change="load" /></section>

    <el-dialog v-model="dialogVisible" :title="t('opsV1.maintenance.editTitle')" width="min(820px, 95vw)"><el-form label-position="top" :model="form"><div class="form-grid">

      <el-form-item v-if="isPlatform" :label="t('opsV1.fields.siteID')"><el-select v-model="form.siteID" clearable filterable><el-option v-for="item in siteOptions" :key="item.value" :label="item.label" :value="Number(item.value)" /></el-select></el-form-item>
      <el-form-item v-if="activeTab==='SITE'" :label="t('opsV1.maintenance.scope')"><el-segmented v-model="form.scopeType" :options="['SITE']" /></el-form-item>
      <el-form-item v-if="activeTab==='PROVIDER'" :label="t('opsV1.fields.provider')" required><el-select v-model="form.providerCode" filterable><el-option v-for="item in providerOptions" :key="item.value" :label="item.label" :value="item.value" /></el-select></el-form-item>
      <el-form-item v-if="activeTab==='GAME'" :label="t('opsV1.fields.game')" required><el-select v-model="form.gameID" filterable><el-option v-for="item in gameOptions" :key="item.value" :label="item.label" :value="Number(item.value)" /></el-select></el-form-item>
      <el-form-item :label="t('opsV1.fields.status')"><el-switch v-model="form.enabled" /></el-form-item>
      <el-form-item :label="t('opsV1.fields.from')"><el-date-picker v-model="form.startAt" type="datetime" value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item>
      <el-form-item :label="t('opsV1.fields.to')"><el-date-picker v-model="form.endAt" type="datetime" value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item>
      <el-form-item class="span-2" :label="t('opsV1.fields.reason')" required><el-input v-model="form.reason" /></el-form-item>
      <template v-if="activeTab==='SITE'"><el-form-item :label="t('opsV1.fields.title')"><el-input v-model="form.title" /></el-form-item><el-form-item :label="t('opsV1.maintenance.estimatedRecovery')"><el-date-picker v-model="form.estimatedRecoveryAt" type="datetime" value-format="YYYY-MM-DDTHH:mm:ssZ" /></el-form-item><el-form-item :label="t('opsV1.maintenance.supportURL')"><el-input v-model="form.supportURL" /></el-form-item><el-form-item :label="t('opsV1.maintenance.supportEmail')"><el-input v-model="form.supportEmail" /></el-form-item><el-form-item :label="t('opsV1.maintenance.supportChat')"><el-input v-model="form.supportChat" /></el-form-item><el-form-item :label="t('opsV1.maintenance.ipAllowlist')"><el-input v-model="form.allowlistedIPsText" /></el-form-item><el-form-item :label="t('opsV1.maintenance.adminBypass')"><el-switch v-model="form.allowAdminBypass" /></el-form-item></template>
      <template v-else><el-form-item :label="t('opsV1.maintenance.countries')"><el-input v-model="form.affectedCountriesText" /></el-form-item><el-form-item :label="t('opsV1.maintenance.currencies')"><el-input v-model="form.affectedCurrenciesText" /></el-form-item><el-form-item :label="t('opsV1.maintenance.allowExisting')"><el-switch v-model="form.allowExistingSessions" /></el-form-item><el-form-item :label="t('opsV1.maintenance.hideLobby')"><el-switch v-model="form.hideFromLobby" /></el-form-item><el-form-item :label="t('opsV1.maintenance.blockLaunch')"><el-switch v-model="form.blockNewLaunches" /></el-form-item></template>
      <el-form-item class="span-2" :label="t('opsV1.maintenance.messageZh')"><el-input v-model="form.messageZh" type="textarea" :rows="2" /></el-form-item>
      <el-form-item class="span-2" :label="t('opsV1.maintenance.messageEn')"><el-input v-model="form.messageEn" type="textarea" :rows="2" /></el-form-item>
    </div></el-form><template #footer><el-button @click="dialogVisible=false">{{ t('common.cancel') }}</el-button><el-button type="primary" @click="save">{{ t('common.save') }}</el-button></template></el-dialog>
  </main>
</template>

<style scoped>
.maintenance-page{display:flex;flex-direction:column;gap:14px}.page-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.page-heading p{margin:0 0 4px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.page-heading h1{margin:0;font-size:24px;letter-spacing:0}.page-heading span{display:block;margin-top:7px;color:#909399;font-size:13px}.mode-tabs{padding:0 18px;background:#fff;border:1px solid #e4e7ed;border-radius:4px}.filter-band,.table-band{padding:16px 18px;background:#fff;border:1px solid #e4e7ed;border-radius:4px}.filter-band :deep(.el-select),.filter-band :deep(.el-input){width:180px}.table-band small{display:block;margin-top:3px;color:#909399}.table-band>.el-empty{padding:36px 0}.table-band :deep(.el-pagination){justify-content:flex-end;margin-top:16px}.window-separator{margin:0 8px;color:#909399}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:0 16px}.span-2{grid-column:1/-1}@media(max-width:700px){.page-heading{flex-direction:column}.filter-band :deep(.el-form-item),.filter-band :deep(.el-select),.filter-band :deep(.el-input){width:100%}.form-grid{grid-template-columns:1fr}.span-2{grid-column:auto}}
</style>
