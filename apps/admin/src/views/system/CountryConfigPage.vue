<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Edit, Delete } from '@element-plus/icons-vue'
import { systemAPI, type SystemRecord } from '@/api/adminSystem'
import { useAdminPermissions } from '@/composables/useAdminPermissions'
import { countryCodes } from '@/data/countryCodes'
import { rowID } from './config'

const { locale } = useI18n()
const { can, loadPermissions } = useAdminPermissions()
const tr = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en
const loading = ref(false)
const saving = ref(false)
const visible = ref(false)
const advanced = ref<string[]>([])
const rows = ref<SystemRecord[]>([])
const presets = ref<Record<string, any>[]>([])
const search = ref('')
const editing = ref(0)
const form = reactive<Record<string, any>>({})
const displayNames = {
  zh: new Intl.DisplayNames(['zh-CN'], { type: 'region' }),
  en: new Intl.DisplayNames(['en'], { type: 'region' }),
}
const options = countryCodes.map(code => ({ code, label: `${displayNames.zh.of(code)} / ${displayNames.en.of(code)} · ${code}` }))
const fields = ['countryCode', 'countryNameZh', 'countryNameEn', 'currencyCode', 'currencyNameZh', 'currencyNameEn', 'currencySymbol', 'currencyDecimalPlaces', 'timezone', 'defaultLanguage', 'supportedLanguages', 'phoneCountryCode', 'phoneLength', 'enabled', 'reason']
const filtered = computed(() => {
  const value = search.value.trim().toLowerCase()
  return rows.value.filter(row => [row.countryCode, row.countryNameZh, row.countryNameEn, row.currencyCode].some(item => String(item || '').toLowerCase().includes(value)))
})

async function load() {
  loading.value = true
  try {
    const [page, defaults] = await Promise.all([systemAPI.countries.list({ page: 1, pageSize: 200 }), systemAPI.countries.presets()])
    const all = [...page.list]
    for (let next = 2; all.length < page.total; next++) {
      const more = await systemAPI.countries.list({ page: next, pageSize: 200 })
      if (!more.list.length) break
      all.push(...more.list)
    }
    rows.value = all
    presets.value = defaults
  } catch { ElMessage.error(tr('国家配置加载失败', 'Could not load countries')) }
  finally { loading.value = false }
}
function reset() {
  Object.keys(form).forEach(key => delete form[key])
  Object.assign(form, { countryCode: '', enabled: true, currencyDecimalPlaces: 2, timezone: 'UTC', defaultLanguage: 'en', supportedLanguages: 'en', phoneLength: 0, reason: '' })
  advanced.value = []
}
function create() { editing.value = 0; reset(); visible.value = true }
async function edit(row: SystemRecord) {
  reset()
  editing.value = rowID(row)
  visible.value = true
  saving.value = true
  try {
    const detail = await systemAPI.countries.detail(rowID(row))
    Object.assign(form, detail.overview || row)
    form.reason = ''
  } catch { visible.value = false; ElMessage.error(tr('国家配置加载失败', 'Could not load country')) }
  finally { saving.value = false }
}
async function chooseCountry(code: string) {
  const existing = rows.value.find(row => row.countryCode === code)
  if (existing) {
    ElMessage.info(tr('该国家已存在，已打开已有配置', 'This country exists; opening its configuration'))
    await edit(existing)
    return
  }
  const preset = presets.value.find(row => row.countryCode === code)
  reset()
  Object.assign(form, { countryCode: code, countryNameZh: displayNames.zh.of(code), countryNameEn: displayNames.en.of(code) }, preset || {})
  form.enabled = true
  if (!preset) advanced.value = ['details']
}
function changeCurrency() {
  const code = String(form.currencyCode || '').trim().toUpperCase()
  form.currencyCode = code
  if (!/^[A-Z]{3}$/.test(code)) return
  try {
    form.currencyNameZh = new Intl.DisplayNames(['zh-CN'], { type: 'currency' }).of(code)
    form.currencyNameEn = new Intl.DisplayNames(['en'], { type: 'currency' }).of(code)
    form.currencySymbol = new Intl.NumberFormat('en', { style: 'currency', currency: code, currencyDisplay: 'narrowSymbol' }).formatToParts(0).find(part => part.type === 'currency')?.value || code
  } catch { /* The server validates the supplied currency code. */ }
}
async function save() {
  if (!form.countryCode || !form.currencyCode || !form.currencySymbol || !form.phoneCountryCode) {
    ElMessage.warning(tr('请选择国家，并补齐币种和电话区号', 'Choose a country and complete its currency and dialing code'))
    return
  }
  const body = Object.fromEntries(fields.map(key => [key, form[key]]))
  body.isoCountryCode = form.countryCode
  body.currencyCode = String(form.currencyCode).trim().toUpperCase()
  body.supportedLanguages = [...new Set([form.defaultLanguage, ...String(form.supportedLanguages || '').split(',')].map(value => String(value).trim()).filter(Boolean))].join(',')
  saving.value = true
  try {
    if (editing.value) await systemAPI.countries.update(editing.value, body)
    else await systemAPI.countries.create(body)
    ElMessage.success(tr('国家配置已保存', 'Country saved'))
    visible.value = false
    await load()
  } catch (error) { ElMessage.error(error instanceof Error ? error.message : tr('保存失败', 'Save failed')) }
  finally { saving.value = false }
}
async function status(row: SystemRecord) {
  try {
    const reason = await ElMessageBox.prompt(tr('请输入变更原因', 'Enter a change reason'), tr('启用／停用国家', 'Change Country Status'), { inputValidator: value => Boolean(value?.trim()) || tr('请填写原因', 'Reason required') })
    await systemAPI.countries.status(rowID(row), !Boolean(row.enabled), reason.value)
    await load()
  } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error(tr('状态修改失败', 'Could not change status')) }
}
async function remove(row: SystemRecord) {
  try {
    const reason = await ElMessageBox.prompt(tr('请输入删除原因；正在使用的国家无法删除', 'Enter a reason; countries in use cannot be deleted'), tr('删除国家', 'Delete Country'), { inputValidator: value => Boolean(value?.trim()) || tr('请填写原因', 'Reason required') })
    await systemAPI.countries.remove(rowID(row), reason.value)
    await load()
  } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error(tr('删除失败，该国家可能正在使用', 'Could not delete; this country may be in use')) }
}
onMounted(async () => { await loadPermissions(); if (can('system:country:list')) await load() })
</script>

<template>
  <div v-loading="loading" class="country-page">
    <header class="country-heading"><div><h1>{{ tr('国家配置', 'Countries') }}</h1><p>{{ tr('每次部署只启用一个国家。启用后自动停用其他国家，并立即应用到 H5 的国家、币种、语言、时区和注册电话区号。', 'One country is active per installation. Enabling it immediately applies its country, currency, language, timezone and registration dialing code to H5.') }} <RouterLink v-if="can('system:site:view')" to="/system/brand">{{ tr('打开 H5 配置', 'Open H5 Settings') }}</RouterLink></p></div><div><el-button :icon="Refresh" @click="load">{{ tr('刷新', 'Refresh') }}</el-button><el-button v-if="can('system:country:create')" type="primary" :icon="Plus" @click="create">{{ tr('新增国家', 'Add Country') }}</el-button></div></header>
    <el-input v-model="search" class="country-search" clearable :placeholder="tr('搜索国家名称、代码或币种，例如：印度尼西亚 / ID / IDR', 'Search name, country code or currency: Indonesia / ID / IDR')" />
    <el-table :data="filtered" row-key="id">
      <el-table-column :label="tr('国家', 'Country')" min-width="225"><template #default="{ row }"><strong>{{ locale.startsWith('zh') ? row.countryNameZh || row.countryNameEn : row.countryNameEn }}</strong><div class="muted">{{ row.countryCode }} · {{ row.countryNameEn }}</div></template></el-table-column>
      <el-table-column :label="tr('币种', 'Currency')" min-width="130"><template #default="{ row }">{{ row.currencyCode }} · {{ row.currencySymbol }}</template></el-table-column>
      <el-table-column prop="phoneCountryCode" :label="tr('电话区号', 'Dialing Code')" width="125" />
      <el-table-column prop="timezone" :label="tr('时区', 'Timezone')" min-width="180" />
      <el-table-column :label="tr('状态', 'Status')" width="110"><template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? tr('已启用', 'Enabled') : tr('已停用', 'Disabled') }}</el-tag></template></el-table-column>
      <el-table-column :label="tr('操作', 'Actions')" min-width="230"><template #default="{ row }"><el-button v-if="can('system:country:update')" link type="primary" :icon="Edit" @click="edit(row)">{{ tr('编辑', 'Edit') }}</el-button><el-button v-if="can('system:country:status')" link :disabled="Boolean(row.enabled)" @click="status(row)">{{ row.enabled ? tr('当前使用', 'Current') : tr('启用并应用', 'Activate') }}</el-button><el-button v-if="can('system:country:delete')" link type="danger" :icon="Delete" @click="remove(row)">{{ tr('删除', 'Delete') }}</el-button></template></el-table-column>
    </el-table>
    <el-dialog v-model="visible" :title="editing ? tr('编辑国家', 'Edit Country') : tr('新增国家', 'Add Country')" width="min(660px,94vw)" destroy-on-close>
      <el-form v-loading="saving" label-position="top" @submit.prevent="save">
        <el-form-item :label="tr('国家／地区', 'Country / Region')" required><el-select v-model="form.countryCode" filterable :disabled="Boolean(editing)" :placeholder="tr('搜索中文、英文名称或代码', 'Search country name or code')" @change="chooseCountry"><el-option v-for="option in options" :key="option.code" :label="option.label" :value="option.code" /></el-select></el-form-item>
        <div v-if="form.countryCode" class="country-summary"><strong>{{ form.currencyCode || '—' }} · {{ form.currencySymbol || '—' }}</strong><span>{{ form.timezone }}</span><span>{{ form.phoneCountryCode || '—' }}</span></div>
        <el-form-item :label="tr('应用到当前 H5', 'Apply to H5')"><el-switch v-model="form.enabled" /></el-form-item>
        <el-collapse v-model="advanced"><el-collapse-item name="details" :title="tr('高级设置（通常无需修改）', 'Advanced settings (usually unchanged)')"><div class="country-grid">
          <el-form-item :label="tr('币种代码', 'Currency Code')"><el-input v-model="form.currencyCode" maxlength="3" @change="changeCurrency" /></el-form-item>
          <el-form-item :label="tr('币种符号', 'Currency Symbol')"><el-input v-model="form.currencySymbol" maxlength="16" /></el-form-item>
          <el-form-item :label="tr('显示小数位', 'Display Decimal Places')"><el-input-number v-model="form.currencyDecimalPlaces" :min="0" :max="8" /></el-form-item>
          <el-form-item :label="tr('时区', 'Timezone')"><el-input v-model="form.timezone" placeholder="Asia/Jakarta" /></el-form-item>
          <el-form-item :label="tr('默认语言', 'Default Language')"><el-input v-model="form.defaultLanguage" placeholder="id" /></el-form-item>
          <el-form-item :label="tr('支持语言（逗号分隔）', 'Languages (comma separated)')"><el-input v-model="form.supportedLanguages" placeholder="id,en" /></el-form-item>
          <el-form-item :label="tr('电话区号', 'Dialing Code')"><el-input v-model="form.phoneCountryCode" placeholder="+62" /></el-form-item>
          <el-form-item :label="tr('手机号长度（0 为可变长度）', 'Phone Length (0 for variable)')"><el-input-number v-model="form.phoneLength" :min="0" :max="15" /></el-form-item>
        </div></el-collapse-item></el-collapse>
        <el-form-item class="country-reason" :label="tr('备注／变更原因', 'Change Reason')"><el-input v-model="form.reason" maxlength="500" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="visible = false">{{ tr('取消', 'Cancel') }}</el-button><el-button type="primary" :loading="saving" @click="save">{{ tr('保存', 'Save') }}</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.country-page { display: grid; gap: 20px; }
.country-heading { display: flex; justify-content: space-between; align-items: center; gap: 20px; }
.country-heading h1 { margin: 0 0 8px; font-size: 24px; }
.country-heading p,.muted { color: var(--el-text-color-secondary); }
.country-heading p { margin: 0; line-height: 1.6; }
.muted { margin-top: 5px; font-size: 12px; }
.country-search { max-width: 580px; }
.country-summary { display: flex; flex-wrap: wrap; gap: 12px; margin: 0 0 20px; padding: 14px; border-radius: 8px; background: var(--el-fill-color-light); }
.country-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0 20px; }
.country-page :deep(.el-select) { width: 100%; }
.country-reason { margin-top: 20px; }
@media(max-width:700px) { .country-heading { align-items: flex-start; flex-direction: column; } .country-grid { grid-template-columns: 1fr; } }
</style>
