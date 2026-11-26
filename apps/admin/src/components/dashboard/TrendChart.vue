<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChartResponse } from '@/api/adminDashboard'
import EChart from './EChart.vue'
const props=withDefaults(defineProps<{data?:ChartResponse;loading?:boolean;barMetrics?:string[]}>(),{barMetrics:()=>[]})
const {t,locale}=useI18n()
const colors=['#409eff','#67c23a','#e6a23c','#f56c6c','#7b61ff','#00a6a6']
const option=computed(()=>({color:colors,tooltip:{trigger:'axis',valueFormatter:(v:unknown)=>Number(v).toLocaleString(locale.value,{maximumFractionDigits:2})},legend:{top:0,type:'scroll'},grid:{left:18,right:18,top:46,bottom:36,containLabel:true},dataZoom:[{type:'inside'},{type:'slider',height:14,bottom:2}],xAxis:{type:'category',data:(props.data?.rows||[]).map(row=>row.time),boundaryGap:props.barMetrics.length>0},yAxis:{type:'value',splitLine:{lineStyle:{color:'#ebeef5'}}},series:(props.data?.metrics||[]).map(metric=>({name:t(`dashboardBI.metrics.${metric}`),type:props.barMetrics.includes(metric)?'bar':'line',smooth:true,symbolSize:5,barMaxWidth:20,data:(props.data?.rows||[]).map(row=>Number(row[metric]||0))}))}))
</script>
<template><EChart :option="option" :loading="loading" :empty="!data?.rows?.length"/></template>
