<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChartRow } from '@/api/adminDashboard'
import EChart from './EChart.vue'
const props=defineProps<{rows?:ChartRow[];loading?:boolean;valueKey?:string}>();const{locale}=useI18n();const option=computed(()=>({tooltip:{trigger:'axis',axisPointer:{type:'shadow'}},grid:{left:18,right:56,top:12,bottom:12,containLabel:true},xAxis:{type:'value',splitLine:{lineStyle:{color:'#ebeef5'}}},yAxis:{type:'category',inverse:true,axisTick:{show:false},data:(props.rows||[]).map(r=>String(r.name||r.id||'-'))},series:[{type:'bar',barMaxWidth:19,data:(props.rows||[]).map((r,index)=>({value:Number(r[props.valueKey||'value']||0),itemStyle:{color:index<3?['#e6a23c','#409eff','#67c23a'][index]:'#79bbff',borderRadius:[0,3,3,0]}})),label:{show:true,position:'right',formatter:(p:{value:number})=>p.value.toLocaleString(locale.value,{maximumFractionDigits:2})}}]}))
</script><template><EChart :option="option" :loading="loading" :empty="!rows?.length"/></template>
