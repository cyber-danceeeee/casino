<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChartRow } from '@/api/adminDashboard'
import EChart from './EChart.vue'
const props=withDefaults(defineProps<{rows?:ChartRow[];loading?:boolean;type?:'pie'|'bar';nameKey?:string;valueKey?:string}>(),{type:'pie',nameKey:'name',valueKey:'value'})
const {t,locale}=useI18n()
const label=(value:unknown)=>{const key=`dashboardBI.values.${String(value)}`;const translated=t(key);return translated===key?String(value):translated}
const option=computed(()=>props.type==='bar'?{tooltip:{trigger:'axis',axisPointer:{type:'shadow'}},grid:{left:15,right:28,top:12,bottom:12,containLabel:true},xAxis:{type:'value',splitLine:{lineStyle:{color:'#ebeef5'}}},yAxis:{type:'category',inverse:true,data:(props.rows||[]).map(r=>label(r[props.nameKey]))},series:[{type:'bar',barMaxWidth:22,data:(props.rows||[]).map(r=>Number(r[props.valueKey]||0)),itemStyle:{color:'#409eff',borderRadius:[0,3,3,0]},label:{show:true,position:'right',formatter:(p:{value:number})=>p.value.toLocaleString(locale.value)}}]}:{tooltip:{trigger:'item',formatter:'{b}<br/>{c} ({d}%)'},legend:{type:'scroll',bottom:0},series:[{type:'pie',radius:['46%','70%'],center:['50%','44%'],avoidLabelOverlap:true,itemStyle:{borderColor:'#fff',borderWidth:2},data:(props.rows||[]).map(r=>({name:label(r[props.nameKey]),value:Number(r[props.valueKey]||0)}))}]})
</script>
<template><EChart :option="option" :loading="loading" :empty="!rows?.length"/></template>
