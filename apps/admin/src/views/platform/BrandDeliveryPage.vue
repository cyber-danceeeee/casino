<script setup lang="ts">
import {onMounted, reactive, ref} from "vue";
import {useRouter} from "vue-router";
import {useI18n} from "vue-i18n";
import {ElMessage} from "element-plus";
import {CopyDocument, Link, View} from "@element-plus/icons-vue";
import {type BrandDeliveryRecord, systemAPI} from "@/api/adminSystem";
import {formatBackendTime} from "@/utils/timezone";

const {t} = useI18n();
const router = useRouter();
const loading = ref(false), rows = ref<BrandDeliveryRecord[]>([]), total = ref(0), page = ref(1), pageSize = ref(20);
const query = reactive({tenantName: "", tenantCode: "", domain: "", domainStatus: ""});

async function load() {
  loading.value = true;
  try {
    const result = await systemAPI.tenantDeliveries({...query, page: page.value, pageSize: pageSize.value});
    rows.value = result.list || [];
    total.value = result.total || 0
  } finally {
    loading.value = false
  }
}

async function copy(row: BrandDeliveryRecord) {
  await navigator.clipboard.writeText([`${t('tenantCenter.fields.tenantName')}: ${row.tenantName}`, `Tenant ID: ${row.tenantID}`, `Site ID: ${row.siteID}`, `${t('tenantCenter.fields.admin')}: ${row.adminUsername}`, `${t('createBrand.fields.adminUrl')}: ${row.adminUrl}`, `${t('createBrand.fields.h5Url')}: ${row.h5Url}`].join('\n'));
  ElMessage.success(t('createBrand.messages.copied'))
}

function open(url: string) {
  if (url) window.open(url, '_blank', 'noopener,noreferrer')
}

onMounted(load);
</script>
<template>
  <main class="platform-page">
    <header class="page-heading">
      <div><p>{{ t('portal.platform.tenants') }}</p>
        <h1>{{ t('tenantCenter.deliveries.title') }}</h1><span>{{ t('tenantCenter.deliveries.description') }}</span>
      </div>
    </header>
    <section class="filter-band">
      <el-form inline>
        <el-form-item :label="t('tenantCenter.fields.tenantName')">
          <el-input v-model="query.tenantName" clearable/>
        </el-form-item>
        <el-form-item :label="t('tenantCenter.fields.tenantCode')">
          <el-input v-model="query.tenantCode" clearable/>
        </el-form-item>
        <el-form-item :label="t('tenantCenter.fields.domain')">
          <el-input v-model="query.domain" clearable/>
        </el-form-item>
        <el-form-item :label="t('tenantCenter.fields.domainStatus')">
          <el-select v-model="query.domainStatus" clearable>
            <el-option v-for="value in ['PENDING','ACTIVE','DISABLED','FAILED']" :key="value"
                       :label="t(`tenantCenter.domainStatus.${value}`)" :value="value"/>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="page=1;load()">{{ t('common.search') }}</el-button>
        </el-form-item>
      </el-form>
    </section>
    <section class="table-band">
      <el-table :data="rows" v-loading="loading" row-key="id">
        <el-table-column prop="tenantName" :label="t('tenantCenter.fields.tenantName')" min-width="170">
          <template #default="{row}"><strong>{{ row.tenantName }}</strong><small>{{ row.tenantCode }}</small></template>
        </el-table-column>
        <el-table-column prop="siteName" :label="t('tenantCenter.fields.site')" min-width="140"/>
        <el-table-column prop="adminUsername" :label="t('tenantCenter.fields.admin')" min-width="130"/>
        <el-table-column prop="primaryDomain" :label="t('tenantCenter.fields.domain')" min-width="190"/>
        <el-table-column prop="domainStatus" :label="t('tenantCenter.fields.domainStatus')" width="125">
          <template #default="{row}">
            <el-tag :type="row.domainStatus==='ACTIVE'?'success':row.domainStatus==='FAILED'?'danger':'warning'">
              {{ t(`tenantCenter.domainStatus.${row.domainStatus}`) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="initialConfigVersion" :label="t('createBrand.fields.initialVersion')" width="130"/>
        <el-table-column prop="deliveredByName" :label="t('tenantCenter.fields.deliveredBy')" width="130"/>
        <el-table-column prop="deliveredAt" :label="t('tenantCenter.fields.deliveredAt')" width="175">
          <template #default="{row}">{{ formatBackendTime(row.deliveredAt) }}</template>
        </el-table-column>
        <el-table-column fixed="right" :label="t('common.actions')" width="170">
          <template #default="{row}">
            <div class="icon-actions">
              <el-tooltip :content="t('common.view')">
                <el-button circle :icon="View" @click="router.push(`/system/tenants/${row.tenantID}`)"/>
              </el-tooltip>
              <el-tooltip :content="t('createBrand.actions.copyDelivery')">
                <el-button circle :icon="CopyDocument" @click="copy(row)"/>
              </el-tooltip>
              <el-tooltip :content="t('createBrand.actions.openAdmin')">
                <el-button circle :icon="Link" @click="open(row.adminUrl)"/>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" background
                     layout="total, sizes, prev, pager, next" :total="total" @change="load"/>
    </section>
  </main>
</template>
<style scoped>.platform-page {
  display: flex;
  flex-direction: column;
  gap: 16px
}

.page-heading p {
  margin: 0 0 4px;
  color: #337ecc;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase
}

.page-heading h1 {
  margin: 0;
  font-size: 24px;
  letter-spacing: 0
}

.page-heading span {
  display: block;
  margin-top: 7px;
  color: #909399;
  font-size: 13px
}

.filter-band, .table-band {
  padding: 18px 20px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px
}

.filter-band :deep(.el-form-item) {
  margin-bottom: 0
}

.filter-band :deep(.el-input), .filter-band :deep(.el-select) {
  width: 180px
}

.table-band small {
  display: block;
  margin-top: 3px;
  color: #909399
}

.table-band .el-pagination {
  justify-content: flex-end;
  margin-top: 16px
}

.icon-actions {
  display: flex;
  gap: 7px
}</style>
