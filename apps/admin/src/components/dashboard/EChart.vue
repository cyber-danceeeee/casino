<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts/core'
import { BarChart, FunnelChart, LineChart, PieChart } from 'echarts/charts'
import { DataZoomComponent, GridComponent, LegendComponent, TitleComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import ChartEmptyState from './ChartEmptyState.vue'

echarts.use([BarChart, FunnelChart, LineChart, PieChart, DataZoomComponent, GridComponent, LegendComponent, TitleComponent, TooltipComponent, CanvasRenderer])

const props = withDefaults(defineProps<{ option: echarts.EChartsCoreOption; loading?: boolean; empty?: boolean; height?: number }>(), { loading: false, empty: false, height: 320 })
const host = ref<HTMLElement>()
let chart: echarts.ECharts | undefined
let observer: ResizeObserver | undefined

function render() {
  if (!host.value || props.empty) return
  chart ||= echarts.getInstanceByDom(host.value) || echarts.init(host.value)
  chart.setOption(props.option, true)
  props.loading ? chart.showLoading('default', { color: '#409eff', maskColor: 'rgba(255,255,255,.72)' }) : chart.hideLoading()
}
function resize() { chart?.resize() }

watch(() => [props.option, props.loading, props.empty], async () => { await nextTick(); render() }, { deep: true })
onMounted(() => { render(); if (host.value) { observer = new ResizeObserver(resize); observer.observe(host.value) }; window.addEventListener('resize', resize) })
onBeforeUnmount(() => { observer?.disconnect(); window.removeEventListener('resize', resize); chart?.dispose(); chart = undefined })
</script>
<template><div class="echart-shell" :style="{ height: `${height}px` }"><ChartEmptyState v-if="empty && !loading"/><div v-show="!empty || loading" ref="host" class="echart-host" /></div></template>
<style scoped>.echart-shell,.echart-host{width:100%;min-width:0}.echart-host{height:100%}</style>
