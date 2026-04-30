<template>
  <div class="rounded-xl border border-gray-800 bg-[#12181f] p-5 md:p-6">
    <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wide mb-4">Class distribution</h3>
    <ClientOnly>
      <VueApexCharts
        v-if="hasData"
        type="donut"
        height="280"
        :options="chartOptions"
        :series="series"
      />
      <div v-else class="h-[280px] flex items-center justify-center text-gray-500 text-sm">
        No class data to chart
      </div>
      <template #fallback>
        <div class="h-[280px] flex items-center justify-center text-gray-500">Loading chart…</div>
      </template>
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import { useChartTheme } from '~/composables/useChartTheme'

const props = defineProps<{
  totalObjects: Record<string, number> | null
}>()

const { seriesColors, base } = useChartTheme()

const hasData = computed(() => {
  if (!props.totalObjects) return false
  return Object.keys(props.totalObjects).length > 0
})

const series = computed(() => {
  if (!props.totalObjects) return []
  return Object.values(props.totalObjects)
})

const labels = computed(() => {
  if (!props.totalObjects) return []
  return Object.keys(props.totalObjects)
})

const chartOptions = computed(() => ({
  ...base,
  labels: labels.value,
  colors: seriesColors,
  plotOptions: {
    pie: {
      donut: {
        size: '68%',
        labels: {
          show: true,
          name: { color: '#9ca3af' },
          value: { color: '#e5e7eb' },
          total: {
            show: true,
            label: 'Objects',
            color: '#6b7280',
            formatter: () => String(series.value.reduce((a, b) => a + b, 0)),
          },
        },
      },
    },
  },
  legend: {
    ...base.legend,
    position: 'bottom' as const,
  },
  dataLabels: { enabled: false },
  stroke: { width: 0 },
}))
</script>
