<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChartRow } from '@/api/adminDashboard'
import EChart from './EChart.vue'
const props=defineProps<{rows?:ChartRow[];loading?:boolean}>();const{t}=useI18n();const option=computed(()=>({tooltip:{trigger:'item',formatter:'{b}: {c}'},series:[{type:'funnel',left:'8%',right:'8%',top:10,bottom:10,minSize:'18%',maxSize:'100%',sort:'descending',gap:3,label:{show:true,position:'inside',formatter:'{b}\n{c}'},itemStyle:{borderColor:'#fff',borderWidth:1},data:(props.rows||[]).map(r=>({name:t(`dashboardBI.metrics.${r.stage}`),value:Number(r.value||0)}))}]}))
</script><template><EChart :option="option" :loading="loading" :empty="!rows?.length"/></template>
