<template>
  <div class="radar-chart">
    <Radar :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Radar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  RadialLinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  Legend,
} from 'chart.js'
import type { VersionKPIs } from '@/services/api'
import { toQualityScores, RADAR_LABELS } from '@/utils/kpiScore'

ChartJS.register(RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend)

const props = defineProps<{
  versionA: VersionKPIs
  versionB: VersionKPIs
}>()

const chartData = computed(() => ({
  labels: RADAR_LABELS,
  datasets: [
    {
      label: props.versionA.version,
      data: toQualityScores(props.versionA),
      borderColor: 'rgba(25, 118, 210, 0.9)',
      backgroundColor: 'rgba(25, 118, 210, 0.15)',
      pointBackgroundColor: 'rgba(25, 118, 210, 0.9)',
    },
    {
      label: props.versionB.version,
      data: toQualityScores(props.versionB),
      borderColor: 'rgba(211, 47, 47, 0.9)',
      backgroundColor: 'rgba(211, 47, 47, 0.15)',
      pointBackgroundColor: 'rgba(211, 47, 47, 0.9)',
    },
  ],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: true,
  scales: {
    r: {
      min: 0,
      max: 100,
      ticks: { display: false },
      pointLabels: { font: { size: 12 } },
    },
  },
  plugins: {
    legend: { position: 'bottom' as const },
    tooltip: {
      callbacks: {
        label: (ctx: { dataset: { label?: string }; raw: unknown }) =>
          `${ctx.dataset.label ?? ''}: ${Number(ctx.raw).toFixed(1)}`,
      },
    },
  },
}
</script>

<style lang="scss" scoped>
.radar-chart {
  max-width: 480px;
  margin: 1.5rem auto;
}
</style>

