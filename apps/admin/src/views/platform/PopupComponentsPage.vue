<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Plus, Refresh, VideoPlay } from '@element-plus/icons-vue'
import { systemAPI, type PopupActionType, type PopupComponentInput, type PopupComponentRecord } from '@/api/adminSystem'

const { t, locale } = useI18n()
const loading = ref(false), saving = ref(false), dialogVisible = ref(false)
const rows = ref<PopupComponentRecord[]>([])
const editingID = ref(0)
const filters = reactive({ keyword: '', category: '', enabled: '' })
const categories = ['DEPOSIT', 'PROMOTION', 'VIP', 'DOWNLOAD', 'GENERAL']
const actionTypes: PopupActionType[] = ['CLOSE_ONLY', 'OPEN_INTERNAL_PAGE', 'OPEN_EXTERNAL_URL', 'OPEN_GAME', 'OPEN_PROMOTION', 'OPEN_DEPOSIT', 'OPEN_WITHDRAW', 'OPEN_SUPPORT', 'OPEN_COUPON', 'OPEN_VIP', 'OPEN_DOWNLOAD']

function emptyForm(): PopupComponentInput {
  return { componentCode: '', nameZh: '', nameEn: '', descriptionZh: '', descriptionEn: '', category: 'GENERAL', previewURL: '', previewVideoURL: '', h5MinVersion: '1.0.0', defaultActionType: 'CLOSE_ONLY', allowedActionTypes: ['CLOSE_ONLY'], enabled: true, sortOrder: 0, reason: '' }
}
const form = reactive<PopupComponentInput>(emptyForm())
const localizedName = (row: PopupComponentRecord) => locale.value.startsWith('zh') ? row.nameZh : row.nameEn
const localizedDescription = (row: PopupComponentRecord) => locale.value.startsWith('zh') ? row.descriptionZh : row.descriptionEn
const canSave = computed(() => /^[A-Z][A-Z0-9_]{2,63}$/.test(form.componentCode) && form.nameZh.trim() && form.nameEn.trim() && form.allowedActionTypes.length > 0 && form.allowedActionTypes.includes(form.defaultActionType))

async function load() {
  loading.value = true
  try {
    const result = await systemAPI.popupComponents.list({ page: 1, pageSize: 200, ...filters })
    rows.value = result.list || []
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('popupComponents.messages.loadFailed'))
  } finally { loading.value = false }
}
function openCreate() {
  editingID.value = 0
  Object.assign(form, emptyForm())
  dialogVisible.value = true
}
function openEdit(row: PopupComponentRecord) {
  editingID.value = row.id
  Object.assign(form, { ...row, allowedActionTypes: [...row.allowedActionTypes], reason: '' })
  dialogVisible.value = true
}
function syncDefaultAction() {
  if (!form.allowedActionTypes.includes(form.defaultActionType)) form.defaultActionType = form.allowedActionTypes[0] || 'CLOSE_ONLY'
}
async function save() {
  if (!canSave.value) { ElMessage.warning(t('popupComponents.validation.required')); return }
  saving.value = true
  try {
    const payload = { ...form, componentCode: form.componentCode.trim().toUpperCase(), nameZh: form.nameZh.trim(), nameEn: form.nameEn.trim(), reason: form.reason?.trim() || t('popupComponents.defaults.saveReason') }
    if (editingID.value) await systemAPI.popupComponents.update(editingID.value, payload)
    else await systemAPI.popupComponents.create(payload)
    dialogVisible.value = false
    ElMessage.success(t('popupComponents.messages.saved'))
    await load()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('popupComponents.messages.saveFailed'))
  } finally { saving.value = false }
}
async function toggle(row: PopupComponentRecord, enabled: boolean) {
  try {
    await systemAPI.popupComponents.update(row.id, { ...row, enabled, allowedActionTypes: [...row.allowedActionTypes], reason: t('popupComponents.defaults.statusReason') })
    ElMessage.success(t('popupComponents.messages.statusSaved'))
    await load()
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('popupComponents.messages.saveFailed'))
    await load()
  }
}
async function remove(row: PopupComponentRecord) {
  try {
    const result = await ElMessageBox.prompt(t('popupComponents.dialogs.deleteMessage'), t('popupComponents.dialogs.deleteTitle'), { type: 'warning', confirmButtonText: t('popupComponents.actions.delete'), cancelButtonText: t('popupComponents.actions.cancel'), inputValidator: (value) => Boolean(value.trim()) || t('popupComponents.validation.reasonRequired') })
    await systemAPI.popupComponents.remove(row.id, result.value.trim())
    ElMessage.success(t('popupComponents.messages.deleted'))
    await load()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error instanceof Error ? error.message : t('popupComponents.messages.deleteFailed'))
  }
}
function openPreview(row: PopupComponentRecord) {
  const target = row.previewVideoURL || row.previewURL
  if (target) window.open(target, '_blank', 'noopener,noreferrer')
  else ElMessage.info(t('popupComponents.messages.noPreview'))
}

onMounted(load)
</script>

<template>
  <div class="component-page">
    <header class="page-header">
      <div><p class="eyebrow">{{ t('popupComponents.eyebrow') }}</p><h1>{{ t('popupComponents.title') }}</h1><p>{{ t('popupComponents.subtitle') }}</p></div>
      <div class="header-actions"><el-button :icon="Refresh" @click="load">{{ t('popupComponents.actions.refresh') }}</el-button><el-button type="primary" :icon="Plus" @click="openCreate">{{ t('popupComponents.actions.create') }}</el-button></div>
    </header>
    <el-alert type="info" :title="t('popupComponents.notice')" :closable="false" show-icon />
    <section class="toolbar">
      <el-input v-model="filters.keyword" clearable :placeholder="t('popupComponents.placeholders.keyword')" @keyup.enter="load" />
      <el-select v-model="filters.category" clearable :placeholder="t('popupComponents.fields.category')"><el-option v-for="item in categories" :key="item" :label="t(`popupComponents.categories.${item}`)" :value="item" /></el-select>
      <el-select v-model="filters.enabled" clearable :placeholder="t('popupComponents.fields.status')"><el-option :label="t('popupComponents.status.enabled')" value="true" /><el-option :label="t('popupComponents.status.disabled')" value="false" /></el-select>
      <el-button type="primary" @click="load">{{ t('popupComponents.actions.search') }}</el-button>
    </section>
    <section class="table-shell"><el-table :data="rows" v-loading="loading" row-key="id">
      <el-table-column :label="t('popupComponents.fields.code')" min-width="210"><template #default="{ row }"><strong>{{ row.componentCode }}</strong><small>H5 &gt;= {{ row.h5MinVersion }}</small></template></el-table-column>
      <el-table-column :label="t('popupComponents.fields.name')" min-width="210"><template #default="{ row }"><strong>{{ localizedName(row) }}</strong><small>{{ localizedDescription(row) }}</small></template></el-table-column>
      <el-table-column :label="t('popupComponents.fields.category')" width="140"><template #default="{ row }"><el-tag effect="plain">{{ t(`popupComponents.categories.${row.category}`) }}</el-tag></template></el-table-column>
      <el-table-column :label="t('popupComponents.fields.actionsAllowed')" min-width="260"><template #default="{ row }"><el-tag v-for="action in row.allowedActionTypes" :key="action" class="action-tag" size="small">{{ t(`popupLibrary.actionType.${action}`) }}</el-tag></template></el-table-column>
      <el-table-column :label="t('popupComponents.fields.sortOrder')" width="90" prop="sortOrder" />
      <el-table-column :label="t('popupComponents.fields.status')" width="100"><template #default="{ row }"><el-switch :model-value="row.enabled" @change="toggle(row, Boolean($event))" /></template></el-table-column>
      <el-table-column :label="t('popupComponents.fields.actions')" width="150" fixed="right"><template #default="{ row }"><el-tooltip :content="t('popupComponents.actions.preview')"><el-button circle :icon="VideoPlay" @click="openPreview(row)" /></el-tooltip><el-tooltip :content="t('popupComponents.actions.edit')"><el-button circle :icon="Edit" @click="openEdit(row)" /></el-tooltip><el-tooltip :content="t('popupComponents.actions.delete')"><el-button circle type="danger" plain :icon="Delete" @click="remove(row)" /></el-tooltip></template></el-table-column>
    </el-table></section>

    <el-dialog v-model="dialogVisible" :title="editingID ? t('popupComponents.dialogs.editTitle') : t('popupComponents.dialogs.createTitle')" width="820px" destroy-on-close>
      <el-form label-position="top"><el-row :gutter="16">
        <el-col :span="12"><el-form-item :label="t('popupComponents.fields.code')" required><el-input v-model="form.componentCode" :disabled="Boolean(editingID)" @input="form.componentCode = String($event).toUpperCase()" /></el-form-item></el-col>
        <el-col :span="6"><el-form-item :label="t('popupComponents.fields.category')"><el-select v-model="form.category"><el-option v-for="item in categories" :key="item" :label="t(`popupComponents.categories.${item}`)" :value="item" /></el-select></el-form-item></el-col>
        <el-col :span="6"><el-form-item :label="t('popupComponents.fields.h5MinVersion')"><el-input v-model="form.h5MinVersion" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item :label="t('popupComponents.fields.nameZh')" required><el-input v-model="form.nameZh" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item :label="t('popupComponents.fields.nameEn')" required><el-input v-model="form.nameEn" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item :label="t('popupComponents.fields.descriptionZh')"><el-input v-model="form.descriptionZh" type="textarea" :rows="2" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item :label="t('popupComponents.fields.descriptionEn')"><el-input v-model="form.descriptionEn" type="textarea" :rows="2" /></el-form-item></el-col>
        <el-col :span="24"><el-alert type="info" :closable="false" :title="t('popupComponents.previewHelp')" /></el-col>
        <el-col :span="12"><el-form-item :label="t('popupComponents.fields.previewURL')"><el-input v-model="form.previewURL" :placeholder="t('popupComponents.placeholders.previewURL')" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item :label="t('popupComponents.fields.previewVideoURL')"><el-input v-model="form.previewVideoURL" :placeholder="t('popupComponents.placeholders.previewVideoURL')" /></el-form-item></el-col>
        <el-col v-if="form.previewVideoURL || form.previewURL" :span="24"><div class="registered-preview"><video v-if="form.previewVideoURL" :src="form.previewVideoURL" controls /><iframe v-else :src="form.previewURL" :title="t('popupComponents.actions.preview')" /></div></el-col>
        <el-col :span="16"><el-form-item :label="t('popupComponents.fields.actionsAllowed')" required><el-select v-model="form.allowedActionTypes" multiple filterable @change="syncDefaultAction"><el-option v-for="action in actionTypes" :key="action" :label="t(`popupLibrary.actionType.${action}`)" :value="action" /></el-select></el-form-item></el-col>
        <el-col :span="8"><el-form-item :label="t('popupComponents.fields.defaultAction')"><el-select v-model="form.defaultActionType"><el-option v-for="action in form.allowedActionTypes" :key="action" :label="t(`popupLibrary.actionType.${action}`)" :value="action" /></el-select></el-form-item></el-col>
        <el-col :span="8"><el-form-item :label="t('popupComponents.fields.sortOrder')"><el-input-number v-model="form.sortOrder" :min="0" /></el-form-item></el-col>
        <el-col :span="8"><el-form-item :label="t('popupComponents.fields.status')"><el-switch v-model="form.enabled" /></el-form-item></el-col>
        <el-col :span="24"><el-form-item :label="t('popupComponents.fields.reason')"><el-input v-model="form.reason" type="textarea" :rows="2" /></el-form-item></el-col>
      </el-row></el-form>
      <template #footer><el-button @click="dialogVisible = false">{{ t('popupComponents.actions.cancel') }}</el-button><el-button type="primary" :loading="saving" @click="save">{{ t('popupComponents.actions.save') }}</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.component-page{padding:24px;color:#18212f}.page-header{display:flex;justify-content:space-between;gap:24px;align-items:flex-start;margin-bottom:18px}.page-header h1{margin:0;font-size:28px;letter-spacing:0}.page-header p{margin:7px 0 0;color:#667085}.eyebrow{color:#175cd3!important;font-size:12px;font-weight:700;text-transform:uppercase}.header-actions{display:flex;gap:8px}.toolbar{display:grid;grid-template-columns:minmax(240px,1fr) 180px 150px auto;gap:10px;margin:16px 0}.table-shell{border:1px solid #e4e7ec;border-radius:6px;overflow:hidden;background:#fff}strong,small{display:block}small{margin-top:4px;color:#667085}.action-tag{margin:2px 4px 2px 0}.registered-preview{height:300px;margin:8px 0 16px;border:1px solid #dcdfe6;background:#f5f7fa}.registered-preview iframe,.registered-preview video{width:100%;height:100%;border:0;object-fit:contain}.el-form :deep(.el-select),.el-form :deep(.el-input-number){width:100%}@media(max-width:760px){.component-page{padding:16px}.page-header{flex-direction:column}.toolbar{grid-template-columns:1fr}}
</style>
