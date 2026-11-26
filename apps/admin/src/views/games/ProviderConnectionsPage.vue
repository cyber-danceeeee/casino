<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Setting } from '@element-plus/icons-vue'
import { fetchProviderConnections, saveProviderConnection, type ConnectionInput, type ConnectionList, type ProviderConnection } from '@/api/providerConnections'
import { syncProvider, testProvider } from '@/api/adminGames'
import { useAdminPermissions } from '@/composables/useAdminPermissions'
const { locale } = useI18n()
const tr=(zh:string,en:string)=>locale.value.startsWith('zh')?zh:en
const { can, loadPermissions }=useAdminPermissions()
const data=ref<ConnectionList>({list:[],market:{countryCode:'',currencyCode:''},callbackBaseUrl:'',adapters:[],countries:[]})
const loading=ref(false),saving=ref(false),visible=ref(false),editing=ref(false),busy=ref('')
const form=reactive<ConnectionInput>({providerCode:'',adapterType:'softaggregator',name:'',apiBaseUrl:'',environment:'sandbox',enabled:false,supportedCountries:'',supportedCurrencies:'',callbackIPAllowlist:[],version:0,credentials:{},clearCredentials:false,reason:''})
const selectedCountries=ref<string[]>([]),selectedCurrencies=ref<string[]>([]),ipRules=ref('')
const configured=ref<Record<string,boolean>>({})
const adapter=computed(()=>data.value.adapters.find(a=>a.type===form.adapterType))
const currencyOptions=computed(()=>[...new Set(data.value.countries.map(c=>c.currencyCode))].sort())
const labels:Record<string,[string,string]>={merchantCode:['商户号','Merchant Code'],secretKey:['商户密钥','Secret Key'],subMerchantCode:['子商户号','Submerchant Code'],apiLogin:['API 账号','API Login'],apiPassword:['API 密码','API Password'],saltKey:['回调签名密钥','Callback Signing Key']}
async function load(){loading.value=true;try{data.value=await fetchProviderConnections()}catch{ElMessage.error(tr('供应商配置加载失败','Could not load supplier configuration'))}finally{loading.value=false}}
function edit(row?:ProviderConnection){
 editing.value=Boolean(row);configured.value=row?.credentialFields||{}
 Object.assign(form,{providerCode:row?.providerCode||'',adapterType:row?.adapterType||'softaggregator',name:row?.name||'',apiBaseUrl:row?.apiBaseUrl||'',environment:row?.environment||'sandbox',enabled:row?.enabled||false,supportedCountries:row?.supportedCountries||'',supportedCurrencies:row?.supportedCurrencies||'',callbackIPAllowlist:row?.callbackIPAllowlist||[],version:row?.version||0,credentials:{},clearCredentials:false,reason:''})
 selectedCountries.value=(form.supportedCountries||data.value.market.countryCode).split(',').filter(Boolean)
 selectedCurrencies.value=(form.supportedCurrencies||data.value.market.currencyCode).split(',').filter(Boolean)
 ipRules.value=form.callbackIPAllowlist.join('\n');visible.value=true
}
async function save(){
 if(!form.providerCode||!form.name||!form.apiBaseUrl||!form.reason.trim()||!selectedCountries.value.length||!selectedCurrencies.value.length){ElMessage.warning(tr('请填写供应商、接口地址、地区规则和变更原因','Complete supplier details, API URL, market rules and change reason'));return}
 saving.value=true
 try{await saveProviderConnection({...form,supportedCountries:selectedCountries.value.join(','),supportedCurrencies:selectedCurrencies.value.join(','),callbackIPAllowlist:ipRules.value.split(/[\n,]+/).map(v=>v.trim()).filter(Boolean)})
  form.credentials={};visible.value=false;ElMessage.success(tr('配置已保存并生效','Configuration saved and active'));await load()
 }catch{ElMessage.error(tr('保存失败，请检查配置；如配置已被修改，请刷新后重试','Save failed; check configuration or reload if another change was made'))}finally{saving.value=false}
}
function callbackURL(row:ProviderConnection){try{return new URL(row.callbackPath,data.value.callbackBaseUrl||window.location.origin).href}catch{return row.callbackPath}}
async function test(row:ProviderConnection){busy.value=row.providerCode;try{await testProvider(row.providerCode);ElMessage.success(tr('连接测试通过','Connection test passed'))}catch{ElMessage.error(tr('连接测试失败，请检查接口地址和账号配置','Connection test failed; check API URL and credentials'))}finally{busy.value=''}}
async function sync(row:ProviderConnection){try{const answer=await ElMessageBox.prompt(tr('请输入同步原因；同步后请到游戏列表选择并发布游戏','Enter a sync reason; select and publish games from the game list afterwards'),tr('同步游戏目录','Synchronize Games'),{inputValidator:v=>Boolean(v?.trim())||tr('需要同步原因','Reason required')});busy.value=row.providerCode;await syncProvider(row.providerCode,answer.value);ElMessage.success(tr('同步任务已提交','Synchronization queued'))}catch(e){if(e!=='cancel'&&e!=='close')ElMessage.error(tr('同步未提交，请检查供应商状态','Could not queue synchronization; check supplier status'))}finally{busy.value=''}}
onMounted(async()=>{await loadPermissions();await load()})
</script>
<template>
 <div class="connections-page" v-loading="loading">
  <div class="page-heading"><div><h1>{{tr('供应商配置','Supplier Connections')}}</h1><p>{{tr('当前部署地区','Deployment market')}}：<strong>{{data.market.countryCode}} / {{data.market.currencyCode}}</strong></p></div><div><el-button :icon="Refresh" @click="load">{{tr('刷新','Refresh')}}</el-button><el-button v-if="can('game:provider:update')" type="primary" :icon="Plus" @click="edit()">{{tr('新增供应商','Add Supplier')}}</el-button></div></div>
  <el-alert type="info" :closable="false" show-icon :title="tr('大厅和启动按已发布的国家、币种统一校验。修改站点地区后请发布 H5 配置；供应商满足地区规则且游戏已发布，才会在 H5 展示。','Lobby and launch use the published country and currency. Publish H5 settings after changing the market. Suppliers must match the market and their games must be published.')" />
  <el-table :data="data.list" row-key="providerCode" class="supplier-table">
   <el-table-column :label="tr('供应商','Supplier')" min-width="180"><template #default="{row}"><strong>{{row.name}}</strong><div class="muted">{{row.providerCode}}</div></template></el-table-column>
   <el-table-column prop="adapterType" :label="tr('接口类型','API Protocol')" min-width="140" />
   <el-table-column :label="tr('允许的地区 / 币种','Allowed Markets')" min-width="190"><template #default="{row}">{{row.supportedCountries||'—'}} / {{row.supportedCurrencies||'—'}}</template></el-table-column>
   <el-table-column :label="tr('连接状态','Connection')" width="150"><template #default="{row}"><el-tag :type="row.enabled?'success':'info'">{{row.enabled?tr('已启用','Enabled'):tr('已停用','Disabled')}}</el-tag><div class="muted">{{row.credentialsConfigured?tr('密钥已配置','Credentials configured'):tr('待配置密钥','Credentials required')}}</div></template></el-table-column>
   <el-table-column :label="tr('地区匹配','Market Match')" width="140"><template #default="{row}"><el-tag :type="row.marketAllowed?'success':'warning'">{{row.marketAllowed?tr('符合当前地区','Matches market'):tr('不符合当前地区','Outside market')}}</el-tag></template></el-table-column>
   <el-table-column :label="tr('操作','Actions')" min-width="270"><template #default="{row}"><el-button v-if="can('game:provider:update')" :icon="Setting" link type="primary" @click="edit(row)">{{tr('配置','Configure')}}</el-button><el-button v-if="can('game:provider:test')" link :disabled="!row.enabled||!row.credentialsConfigured" :loading="busy===row.providerCode" @click="test(row)">{{tr('测试连接','Test')}}</el-button><el-button v-if="can('game:provider:sync')" link :disabled="!row.enabled||!row.credentialsConfigured||!row.marketAllowed" :loading="busy===row.providerCode" @click="sync(row)">{{tr('同步游戏','Sync Games')}}</el-button></template></el-table-column>
  </el-table>
  <el-dialog v-model="visible" :title="tr('供应商连接配置','Supplier Connection')" width="min(800px,94vw)" destroy-on-close @closed="form.credentials={}">
   <el-form label-position="top" @submit.prevent="save">
    <div class="form-grid"><el-form-item :label="tr('连接编码（创建后固定）','Connection Code (immutable)')" required><el-input v-model="form.providerCode" :disabled="editing" placeholder="supplier-pk" maxlength="64" /></el-form-item><el-form-item :label="tr('供应商品牌名称','Supplier Brand Name')" required><el-input v-model="form.name" maxlength="128" /></el-form-item>
    <el-form-item :label="tr('接口类型','API Protocol')" required><el-select v-model="form.adapterType" :disabled="editing" @change="form.credentials={}"><el-option v-for="item in data.adapters" :key="item.type" :value="item.type" :label="item.type" /></el-select></el-form-item><el-form-item :label="tr('环境','Environment')"><el-select v-model="form.environment"><el-option value="sandbox" :label="tr('测试环境','Sandbox')"/><el-option value="production" :label="tr('正式环境','Production')"/></el-select></el-form-item></div>
    <el-form-item :label="tr('API 接口地址','API Base URL')" required><el-input v-model="form.apiBaseUrl" placeholder="https://api.example.com" maxlength="500" /></el-form-item>
    <div class="form-grid"><el-form-item :label="tr('允许的国家','Allowed Countries')" required><el-select v-model="selectedCountries" multiple filterable allow-create default-first-option><el-option v-for="item in data.countries" :key="item.countryCode+'-'+item.currencyCode" :value="item.countryCode" :label="item.countryCode+' · '+item.countryName" /></el-select></el-form-item><el-form-item :label="tr('允许的币种','Allowed Currencies')" required><el-select v-model="selectedCurrencies" multiple filterable allow-create default-first-option><el-option v-for="item in currencyOptions" :key="item" :value="item" :label="item" /></el-select></el-form-item></div>
    <el-alert type="info" :closable="false" :title="tr('密钥留空保留原值，保存后不回显。停用仅阻止新展示和启动，已有游戏交易回调仍继续验证处理。','Leave credentials blank to keep saved values. Secrets are never returned. Disabling blocks new visibility and launches; existing transaction callbacks remain verifiable.')" />
    <div class="form-grid credentials"><el-form-item v-for="key in adapter?.credentialFields||[]" :key="key" :label="tr(...(labels[key]||[key,key]))"><el-input v-model="form.credentials[key]" type="password" show-password autocomplete="new-password" :placeholder="configured[key]?tr('已保存，留空保留','Saved; leave blank to keep'):tr('未配置','Not configured')" /></el-form-item></div>
    <el-form-item :label="tr('回调地址（在供应商平台登记）','Callback URL (register upstream)')"><el-input :model-value="callbackURL({callbackPath:'/api/v1/game/providers/'+form.providerCode+'/callback'} as ProviderConnection)" readonly /></el-form-item>
    <el-form-item :label="tr('回调 IP 白名单（每行一个 IP 或 CIDR）','Callback IP Allowlist (one IP or CIDR per line)')"><el-input v-model="ipRules" type="textarea" :rows="2" /></el-form-item>
    <el-form-item :label="tr('启用此连接','Enable Connection')"><el-switch v-model="form.enabled" /></el-form-item>
    <el-form-item :label="tr('变更原因','Change Reason')" required><el-input v-model="form.reason" type="textarea" maxlength="500" /></el-form-item>
   </el-form><template #footer><el-button @click="visible=false">{{tr('取消','Cancel')}}</el-button><el-button type="primary" :loading="saving" @click="save">{{tr('保存配置','Save Configuration')}}</el-button></template>
  </el-dialog>
 </div>
</template>
<style scoped>
.connections-page{display:grid;gap:20px}.page-heading{display:flex;align-items:center;justify-content:space-between;gap:16px}.page-heading h1{margin:0 0 10px;font-size:24px}.page-heading p,.muted{color:var(--el-text-color-secondary)}.muted{font-size:12px;margin-top:6px}.supplier-table{padding:12px;border-radius:8px}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:0 20px}.form-grid .el-select{width:100%}.credentials{margin-top:20px}@media(max-width:700px){.page-heading{align-items:flex-start;flex-direction:column}.form-grid{grid-template-columns:1fr}}
</style>
