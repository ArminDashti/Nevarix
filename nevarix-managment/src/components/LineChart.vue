<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const props = defineProps<{
  labels: string[]
  datasets: { label: string; data: number[]; color?: string }[]
}>()

const chartData = computed(() => ({
  labels: props.labels,
  datasets: props.datasets.map((d) => ({
    label: d.label,
    data: d.data,
    borderColor: d.color ?? '#c8c8c8',
    backgroundColor: (d.color ?? '#c8c8c8') + '33',
    fill: true,
    tension: 0.35,
    pointRadius: 0,
    borderWidth: 2,
  })),
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { labels: { color: '#d4d4d4' } },
  },
  scales: {
    x: { ticks: { color: '#a8a8a8', maxTicksLimit: 8 }, grid: { color: '#4a4a4a' } },
    y: { ticks: { color: '#a8a8a8' }, grid: { color: '#4a4a4a' } },
  },
}
</script>

<template>
  <div class="h-64 w-full">
    <Line :data="chartData" :options="options" />
  </div>
</template>
