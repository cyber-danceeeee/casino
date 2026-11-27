<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument, Delete, Edit, Monitor, Plus, Rank, Refresh, } from '@element-plus/icons-vue'
import { fetchAdminOptions, optionList, type AdminOption, type AdminOptionSource, type AdminOptions } from '@/api/adminOptions'
import SafeRichTextEditor from '@/components/operations/SafeRichTextEditor.vue'
import { systemAPI, type AnnouncementConfiguration, type AnnouncementType, type BrandOverview, type ContentVisibility, type SiteAnnouncementRecord } from '@/api/adminSystem'
import { formatBackendTime } from '@/utils/timezone'

const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const sorting = ref(false)
const rows = ref<SiteAnnouncementRecord[]>([])
const adminOptions = ref<AdminOptions>({})
const overview = ref<BrandOverview | null>(null)
const editVisible = ref(false)
const previewVisible = ref(false)
const previewRecord = ref<SiteAnnouncementRecord | null>(null)
const editingID = ref(0)
const draggedID = ref(0)
const filters = reactive({ keyword: '', languageCode: '' })
const types: AnnouncementType[] = ['SYSTEM', 'MAINTENANCE', 'PROMOTION', 'PAYMENT', 'GAME', 'GENERAL']
const visibility: ContentVisibility[] = ['ALL', 'GUEST', 'LOGGED_IN']

const emptyForm = () => ({
  title: '', summary: '', content: '', announcementType: 'GENERAL' as AnnouncementType,
  priority: 0, sortOrder: 0, isPinned: false, languageCode: 'ALL', countryCode: 'ALL',
  currencyCode: 'ALL', visibilityScope: 'ALL' as ContentVisibility, startAt: null as Date | null,
  endAt: null as Date | null, enabled: true, reason: '',
})
const form = reactive(emptyForm())

const filteredRows = computed(() => rows.value.filter((row) => !filters.languageCode || configuration(row).languageCode === filters.languageCode))
function configuration(row: SiteAnnouncementRecord) { return row.configuration }
function dimensionOptions(source: AdminOptionSource): AdminOption[] {
  const brandValue = source === 'countries' ? overview.value?.countryCode : source === 'currencies' ? overview.value?.currencyCode : overview.value?.defaultLanguage
  const values = [{ label: 'ALL', value: 'ALL' }, ...optionList(adminOptions.value, source)]
  if (brandValue) values.push({ label: brandValue, value: brandValue })
  if (source === 'languages') values.push({ label: 'ZH', value: 'ZH' }, { label: 'EN', value: 'EN' })
  const normalized = values.map((item) => ({ ...item, value: item.value.toUpperCase() }))
  return [...new Map(normalized.map((item) => [item.value, item])).values()]
}

async function load() {
  loading.value = true
  try {
    const brand = await systemAPI.brand.overview()
    overview.value = brand
    const [result, options] = await Promise.all([
      systemAPI.announcements.list(brand.siteID, { page: 1, pageSize: 200, keyword: filters.keyword, }),
      fetchAdminOptions().catch(() => ({})),
    ])
    rows.value = result.list
    adminOptions.value = options
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('announcementLibrary.messages.loadFailed'))
  } finally { loading.value = false }
}

function openCreate() {
  editingID.value = 0
  Object.assign(form, emptyForm(), { sortOrder: rows.value.length })
  editVisible.value = true
}
function openEdit(row: SiteAnnouncementRecord) {
  const value = configuration(row)
  editingID.value = row.id
  Object.assign(form, emptyForm(), value, { startAt: value.startAt ? new Date(value.startAt) : null, endAt: value.endAt ? new Date(value.endAt) : null, reason: '' })
  editVisible.value = true
}
function payload(overrides: Partial<AnnouncementConfiguration> = {}): AnnouncementConfiguration & { reason: string } {
  return {
    title: form.title.trim(), summary: form.summary.trim(), content: form.content,
    announcementType: form.announcementType, priority: form.priority, sortOrder: form.sortOrder,
    isPinned: form.isPinned, languageCode: form.languageCode.trim() || 'ALL', countryCode: form.countryCode.trim().toUpperCase() || 'ALL',
    currencyCode: form.currencyCode.trim().toUpperCase() || 'ALL', visibilityScope: form.visibilityScope,
    startAt: form.startAt ? form.startAt.toISOString() : null, endAt: form.endAt ? form.endAt.toISOString() : null,
    enabled: form.enabled, ...overrides, reason: form.reason.trim() || t('announcementLibrary.defaults.saveReason'),
  }
}
async function save() {
  if (!overview.value?.siteID || !form.title.trim() || !form.content.trim()) { ElMessage.warning(t('announcementLibrary.validation.required')); return }
  saving.value = true
  try {
    if (editingID.value) await systemAPI.announcements.update(overview.value.siteID, editingID.value, payload())
    else await systemAPI.announcements.create(overview.value.siteID, payload())
    editVisible.value = false
    ElMessage.success(t('liveConfig.saved'))
    window.dispatchEvent(new Event('brand-config-refresh'))
    await load()
  } catch (error) { ElMessage.error(error instanceof Error ? error.message : t('announcementLibrary.messages.saveFailed')) }
  finally { saving.value = false }
}
async function updateConfiguration(row: SiteAnnouncementRecord, changes: Partial<AnnouncementConfiguration>, reason: string) {
  if (!overview.value?.siteID) return
  try {
    await systemAPI.announcements.update(overview.value.siteID, row.id, { ...configuration(row), ...changes, reason })
    ElMessage.success(t('liveConfig.saved'))
    window.dispatchEvent(new Event('brand-config-refresh'))
    await load()
  } catch (error) { ElMessage.error(error instanceof Error ? error.message : t('announcementLibrary.messages.saveFailed')); await load() }
}
async function duplicate(row: SiteAnnouncementRecord) {
  if (!overview.value?.siteID) return
  try { await systemAPI.announcements.copy(overview.value.siteID, row.id); ElMessage.success(t('announcementLibrary.messages.copied')); window.dispatchEvent(new Event('brand-config-refresh')); await load() }
  catch (error) { ElMessage.error(error instanceof Error ? error.message : t('announcementLibrary.messages.copyFailed')) }
}
async function remove(row: SiteAnnouncementRecord) {
  if (!overview.value?.siteID) return
  try {
    await ElMessageBox.confirm(t('announcementLibrary.dialogs.deleteMessage'), t('announcementLibrary.dialogs.deleteTitle'), { type: 'warning', confirmButtonText: t('common.delete'), cancelButtonText: t('announcementLibrary.actions.cancel') })
    await systemAPI.announcements.remove(overview.value.siteID, row.id)
    ElMessage.success(t('liveConfig.saved')); window.dispatchEvent(new Event('brand-config-refresh')); await load()
  } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error(error instanceof Error ? error.message : t('announcementLibrary.messages.deleteFailed')) }
}
function openPreview(row?: SiteAnnouncementRecord) { previewRecord.value = row || filteredRows.value.find((item) => configuration(item).enabled && !configuration(item).deleted) || null; previewVisible.value = true }
function previewDocument(content: string) {
  return `<!doctype html><html><head><meta charset="utf-8"><style>body{font:15px/1.65 system-ui;color:#18212f;padding:22px}a{color:#175cd3}p{margin:0 0 12px}</style></head><body>${content}</body></html>`
}
function dragStart(id: number) { draggedID.value = id }
async function dropOn(targetID: number) {
  if (!overview.value?.siteID) return
  const sourceIndex = rows.value.findIndex((row) => row.id === draggedID.value); const targetIndex = rows.value.findIndex((row) => row.id === targetID); draggedID.value = 0
  if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex) return
  const ordered = [...rows.value]; const [moved] = ordered.splice(sourceIndex, 1); ordered.splice(targetIndex, 0, moved); rows.value = ordered; sorting.value = true
  try { await systemAPI.announcements.reorder(overview.value.siteID, ordered.map((row) => row.id), t('announcementLibrary.defaults.sortReason')); ElMessage.success(t('announcementLibrary.messages.sortSaved')); window.dispatchEvent(new Event('brand-config-refresh')); await load() }
  catch (error) { ElMessage.error(error instanceof Error ? error.message : t('announcementLibrary.messages.sortFailed')); await load() }
  finally { sorting.value = false }
}
onMounted(load)
</script>

<template>
  <div v-loading="loading" class="content-page">
    <header class="page-header">
      <div><p class="eyebrow">{{ t('announcementLibrary.eyebrow') }}</p><h1>{{ t('announcementLibrary.title') }}</h1><p>{{ t('announcementLibrary.subtitle') }}</p></div>
      <div class="header-actions"><el-button :icon="Refresh" @click="load">{{ t('announcementLibrary.actions.refresh') }}</el-button><el-button :icon="Monitor" @click="openPreview()">{{ t('announcementLibrary.actions.preview') }}</el-button><el-button type="success" :icon="Plus" @click="openCreate">{{ t('announcementLibrary.actions.create') }}</el-button></div>
    </header>

    <section class="toolbar"><el-input v-model="filters.keyword" clearable :placeholder="t('announcementLibrary.placeholders.keyword')" @keyup.enter="load" /><el-select v-model="filters.languageCode" clearable filterable :placeholder="t('announcementLibrary.fields.language')"><el-option v-for="item in dimensionOptions('languages')" :key="item.value" :label="item.label" :value="item.value" /></el-select><el-button type="primary" @click="load">{{ t('announcementLibrary.actions.search') }}</el-button></section>
    <section class="table-shell"><el-table :data="filteredRows" row-key="id" v-loading="sorting">
      <el-table-column width="48"><template #default="{ row }"><el-icon class="drag-handle" draggable="true" @dragstart="dragStart(row.id)" @dragover.prevent @drop="dropOn(row.id)"><Rank /></el-icon></template></el-table-column>
      <el-table-column :label="t('announcementLibrary.fields.title')" min-width="230"><template #default="{ row }"><strong>{{ configuration(row).title }}</strong><small>{{ configuration(row).summary || '-' }}</small></template></el-table-column>
      <el-table-column :label="t('announcementLibrary.fields.type')" width="135"><template #default="{ row }"><el-tag effect="plain">{{ t(`announcementLibrary.types.${configuration(row).announcementType}`) }}</el-tag></template></el-table-column>
      <el-table-column :label="t('announcementLibrary.fields.pinned')" width="90"><template #default="{ row }"><el-switch :model-value="configuration(row).isPinned" @change="updateConfiguration(row, { isPinned: Boolean($event) }, t('announcementLibrary.defaults.pinReason'))" /></template></el-table-column>
      <el-table-column :label="t('announcementLibrary.fields.visibility')" width="120"><template #default="{ row }">{{ t(`announcementLibrary.visibility.${configuration(row).visibilityScope}`) }}</template></el-table-column>
      <el-table-column :label="t('announcementLibrary.fields.schedule')" min-width="180"><template #default="{ row }"><span>{{ formatBackendTime(configuration(row).startAt || '') }}</span><small>{{ formatBackendTime(configuration(row).endAt || '') }}</small></template></el-table-column>

      <el-table-column :label="t('announcementLibrary.fields.enabled')" width="90"><template #default="{ row }"><el-switch :model-value="configuration(row).enabled" @change="updateConfiguration(row, { enabled: Boolean($event) }, t('announcementLibrary.defaults.statusReason'))" /></template></el-table-column>
      <el-table-column :label="t('announcementLibrary.fields.actions')" width="190" fixed="right"><template #default="{ row }"><el-tooltip :content="t('announcementLibrary.actions.preview')"><el-button circle :icon="Monitor" @click="openPreview(row)" /></el-tooltip><el-tooltip :content="t('announcementLibrary.actions.edit')"><el-button circle :icon="Edit" @click="openEdit(row)" /></el-tooltip><el-tooltip :content="t('announcementLibrary.actions.copy')"><el-button circle :icon="CopyDocument" @click="duplicate(row)" /></el-tooltip><el-tooltip :content="t('common.delete')"><el-button circle type="danger" plain :icon="Delete" @click="remove(row)" /></el-tooltip></template></el-table-column>
    </el-table></section>

    <el-dialog v-model="editVisible" :title="editingID ? t('announcementLibrary.dialogs.editTitle') : t('announcementLibrary.dialogs.createTitle')" width="800px" destroy-on-close><el-form label-position="top"><el-row :gutter="16">
      <el-col :span="16"><el-form-item :label="t('announcementLibrary.fields.title')" required><el-input v-model="form.title" /></el-form-item></el-col><el-col :span="8"><el-form-item :label="t('announcementLibrary.fields.type')"><el-select v-model="form.announcementType"><el-option v-for="item in types" :key="item" :label="t(`announcementLibrary.types.${item}`)" :value="item" /></el-select></el-form-item></el-col>
      <el-col :span="24"><el-form-item :label="t('announcementLibrary.fields.summary')"><el-input v-model="form.summary" /></el-form-item></el-col><el-col :span="24"><el-form-item :label="t('announcementLibrary.fields.content')" required><SafeRichTextEditor v-model="form.content" /></el-form-item></el-col>
      <el-col :span="8"><el-form-item :label="t('announcementLibrary.fields.visibility')"><el-select v-model="form.visibilityScope"><el-option v-for="item in visibility" :key="item" :label="t(`announcementLibrary.visibility.${item}`)" :value="item" /></el-select></el-form-item></el-col><el-col :span="8"><el-form-item :label="t('announcementLibrary.fields.priority')"><el-input-number v-model="form.priority" /></el-form-item></el-col><el-col :span="8"><el-form-item :label="t('announcementLibrary.fields.sortOrder')"><el-input-number v-model="form.sortOrder" :min="0" /></el-form-item></el-col>
      <el-col :span="12"><el-form-item :label="t('announcementLibrary.fields.startAt')"><el-date-picker v-model="form.startAt" type="datetime" /></el-form-item></el-col><el-col :span="12"><el-form-item :label="t('announcementLibrary.fields.endAt')"><el-date-picker v-model="form.endAt" type="datetime" /></el-form-item></el-col>
      <el-col :span="8"><el-form-item :label="t('announcementLibrary.fields.pinned')"><el-switch v-model="form.isPinned" /></el-form-item></el-col><el-col :span="8"><el-form-item :label="t('announcementLibrary.fields.enabled')"><el-switch v-model="form.enabled" /></el-form-item></el-col><el-col :span="24"><el-form-item :label="t('announcementLibrary.fields.reason')"><el-input v-model="form.reason" type="textarea" /></el-form-item></el-col>
    </el-row></el-form><template #footer><el-button @click="editVisible = false">{{ t('announcementLibrary.actions.cancel') }}</el-button><el-button type="primary" :loading="saving" @click="save">{{ t('common.save') }}</el-button></template></el-dialog>
    <el-dialog v-model="previewVisible" :title="t('announcementLibrary.dialogs.previewTitle')" width="760px"><article v-if="previewRecord" class="announcement-preview"><el-tag effect="plain">{{ t(`announcementLibrary.types.${configuration(previewRecord).announcementType}`) }}</el-tag><h2>{{ configuration(previewRecord).title }}</h2><p>{{ configuration(previewRecord).summary }}</p><iframe sandbox="" :srcdoc="previewDocument(configuration(previewRecord).content)" /></article><el-empty v-else :description="t('announcementLibrary.messages.emptyPreview')" /></el-dialog>
  </div>
</template>

<style scoped>
.content-page{padding:24px;color:#18212f}.page-header{display:flex;justify-content:space-between;gap:24px;align-items:flex-start;margin-bottom:18px}.page-header h1{margin:0;font-size:28px;letter-spacing:0}.page-header p{margin:7px 0 0;color:#667085}.eyebrow{color:#b42318!important;font-size:12px;font-weight:700;text-transform:uppercase}.header-actions{display:flex;flex-wrap:wrap;justify-content:flex-end;gap:8px}.toolbar{display:grid;grid-template-columns:minmax(220px,1fr) 180px 150px auto;gap:10px;margin:16px 0}.table-shell{border:1px solid #e4e7ec;border-radius:6px;overflow:hidden;background:#fff}.drag-handle{cursor:grab;color:#667085;font-size:18px}strong,small{display:block}small{margin-top:4px;color:#667085}.el-form :deep(.el-select),.el-form :deep(.el-date-editor),.el-form :deep(.el-input-number){width:100%}.announcement-preview{padding:10px}.announcement-preview h2{font-size:24px;margin:14px 0 6px}.announcement-preview>p{color:#667085}.announcement-preview iframe{width:100%;height:320px;border:1px solid #e4e7ec;border-radius:5px;background:#fff}@media(max-width:760px){.content-page{padding:16px}.page-header{flex-direction:column}.header-actions{justify-content:flex-start}.toolbar{grid-template-columns:1fr}}
</style>
