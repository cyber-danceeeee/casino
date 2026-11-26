<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Edit, Plus, Refresh } from '@element-plus/icons-vue'
import { systemAPI, type TenantPackage, type TenantPackageInput } from '@/api/adminSystem'

const { t, locale } = useI18n()
const loading = ref(false), saving = ref(false), dialogVisible = ref(false)
const rows = ref<TenantPackage[]>([]), editingID = ref(0)
const statuses = ['AVAILABLE', 'CONTACT', 'LEGACY', 'DISABLED']
const limitFields: Array<{ key: 'sites' | 'games' | 'providers' | 'storageGB' | 'players' | 'sortOrder'; min: number }> = [
  { key: 'sites', min: -1 }, { key: 'games', min: -1 }, { key: 'providers', min: -1 },
  { key: 'storageGB', min: -1 }, { key: 'players', min: -1 }, { key: 'sortOrder', min: 0 },
]
const form = reactive<TenantPackageInput>(emptyForm())

function emptyForm(): TenantPackageInput {
  return { code: '', nameZh: '', nameEn: '', sites: 1, games: 0, providers: 0, storageGB: 0, players: 0, status: 'AVAILABLE', sortOrder: 0, reason: '' }
}
const localizedName = (row: TenantPackage) => locale.value.startsWith('zh') ? row.nameZh : row.nameEn
const limit = (value: number) => value < 0 ? t('tenantCenter.packages.unlimited') : value.toLocaleString()
const canSave = computed(() => /^[A-Z][A-Z0-9_]{1,63}$/.test(form.code) && form.nameZh.trim() && form.nameEn.trim())

async function load() {
  loading.value = true
  try { rows.value = (await systemAPI.tenantPackages.list()).list || [] }
  catch (error) { ElMessage.error(error instanceof Error ? error.message : t('tenantCenter.messages.loadFailed')) }
  finally { loading.value = false }
}
function createPackage() {
  editingID.value = 0
  Object.assign(form, emptyForm())
  dialogVisible.value = true
}
function editPackage(row: TenantPackage) {
  editingID.value = row.id
  Object.assign(form, { ...row, reason: '' })
  dialogVisible.value = true
}
async function save() {
  if (!canSave.value) return ElMessage.warning(t('tenantCenter.packages.validation'))
  saving.value = true
  try {
    const payload = { ...form, code: form.code.trim().toUpperCase(), nameZh: form.nameZh.trim(), nameEn: form.nameEn.trim(), reason: form.reason?.trim() || t('tenantCenter.packages.defaultReason') }
    if (editingID.value) await systemAPI.tenantPackages.update(editingID.value, payload)
    else await systemAPI.tenantPackages.create(payload)
    dialogVisible.value = false
    ElMessage.success(t('tenantCenter.packages.saved'))
    await load()
  } catch (error) { ElMessage.error(error instanceof Error ? error.message : t('tenantCenter.messages.saveFailed')) }
  finally { saving.value = false }
}
async function removePackage(row: TenantPackage) {
  try {
    const { value } = await ElMessageBox.prompt(t('tenantCenter.packages.deleteConfirm'), t('common.delete'), { inputPattern: /\S+/, inputErrorMessage: t('tenantCenter.messages.reasonRequired') })
    await systemAPI.tenantPackages.remove(row.id, value)
    ElMessage.success(t('tenantCenter.packages.deleted'))
    await load()
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error instanceof Error ? error.message : t('tenantCenter.packages.deleteFailed'))
  }
}
onMounted(load)
</script>

<template>
  <main class="platform-page">
    <header class="page-heading"><div><p>{{ t('portal.platform.tenants') }}</p><h1>{{ t('tenantCenter.packages.title') }}</h1><span>{{ t('tenantCenter.packages.description') }}</span></div><div class="heading-actions"><el-button :icon="Refresh" @click="load">{{ t('common.refresh') }}</el-button><el-button type="primary" :icon="Plus" @click="createPackage">{{ t('tenantCenter.packages.create') }}</el-button></div></header>
    <section class="table-band"><el-table :data="rows" v-loading="loading" row-key="id">
      <el-table-column :label="t('tenantCenter.fields.plan')" min-width="190"><template #default="{row}"><strong>{{ localizedName(row) }}</strong><small>{{ row.code }}</small></template></el-table-column>
      <el-table-column prop="tenantCount" :label="t('tenantCenter.packages.tenants')" width="110" align="right"/>
      <el-table-column :label="t('tenantCenter.packages.sites')" width="105" align="right"><template #default="{row}">{{ limit(row.sites) }}</template></el-table-column>
      <el-table-column :label="t('tenantCenter.packages.games')" width="115" align="right"><template #default="{row}">{{ limit(row.games) }}</template></el-table-column>
      <el-table-column :label="t('tenantCenter.packages.providers')" width="125" align="right"><template #default="{row}">{{ limit(row.providers) }}</template></el-table-column>
      <el-table-column :label="t('tenantCenter.packages.storage')" width="120" align="right"><template #default="{row}">{{ row.storageGB < 0 ? limit(row.storageGB) : `${row.storageGB} GB` }}</template></el-table-column>
      <el-table-column :label="t('tenantCenter.packages.players')" width="130" align="right"><template #default="{row}">{{ limit(row.players) }}</template></el-table-column>
      <el-table-column prop="status" :label="t('tenantCenter.packages.availability')" width="145"><template #default="{row}"><el-tag :type="row.status === 'AVAILABLE' ? 'success' : row.status === 'DISABLED' ? 'danger' : 'warning'">{{ t(`tenantCenter.packageStatus.${row.status}`) }}</el-tag></template></el-table-column>
      <el-table-column :label="t('common.actions')" width="110" fixed="right"><template #default="{row}"><el-button circle :icon="Edit" @click="editPackage(row)"/><el-button circle type="danger" plain :icon="Delete" :disabled="row.tenantCount > 0" @click="removePackage(row)"/></template></el-table-column>
    </el-table></section>

    <el-dialog v-model="dialogVisible" :title="editingID ? t('tenantCenter.packages.edit') : t('tenantCenter.packages.create')" width="720px">
      <el-form label-position="top"><el-row :gutter="16">
        <el-col :span="8"><el-form-item :label="t('tenantCenter.packages.code')" required><el-input v-model="form.code" :disabled="Boolean(editingID)" @input="form.code=String($event).toUpperCase()"/></el-form-item></el-col>
        <el-col :span="8"><el-form-item :label="t('tenantCenter.packages.nameZh')" required><el-input v-model="form.nameZh"/></el-form-item></el-col>
        <el-col :span="8"><el-form-item :label="t('tenantCenter.packages.nameEn')" required><el-input v-model="form.nameEn"/></el-form-item></el-col>
        <el-col v-for="field in limitFields" :key="field.key" :span="8"><el-form-item :label="t(`tenantCenter.packages.${field.key}`)"><el-input-number v-model="form[field.key]" :min="field.min" controls-position="right"/></el-form-item></el-col>
        <el-col :span="8"><el-form-item :label="t('tenantCenter.packages.availability')"><el-select v-model="form.status"><el-option v-for="status in statuses" :key="status" :label="t(`tenantCenter.packageStatus.${status}`)" :value="status"/></el-select></el-form-item></el-col>
        <el-col :span="24"><el-form-item :label="t('tenantCenter.fields.reason')"><el-input v-model="form.reason" type="textarea" :rows="2"/></el-form-item></el-col>
      </el-row></el-form>
      <template #footer><el-button @click="dialogVisible=false">{{ t('common.cancel') }}</el-button><el-button type="primary" :loading="saving" @click="save">{{ t('common.confirm') }}</el-button></template>
    </el-dialog>
  </main>
</template>

<style scoped>.platform-page{display:flex;flex-direction:column;gap:16px}.page-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:16px}.page-heading p{margin:0 0 4px;color:#337ecc;font-size:12px;font-weight:700;text-transform:uppercase}.page-heading h1{margin:0;font-size:24px;letter-spacing:0}.page-heading span{display:block;margin-top:7px;color:#909399;font-size:13px}.heading-actions{display:flex;gap:8px}.table-band{padding:18px 20px;background:#fff;border:1px solid #e4e7ed;border-radius:4px}.table-band strong,.table-band small{display:block}.table-band small{margin-top:3px;color:#909399}.el-form :deep(.el-input-number),.el-form :deep(.el-select){width:100%}@media(max-width:720px){.page-heading{flex-direction:column}}</style>
