<template>
  <div class="app-card app-card-hover p-5 md:p-6">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="app-section-title">Class distribution</h3>
      <span v-if="hasData" class="app-chip text-gray-400">
        {{ totalObjects }} objects
      </span>
    </div>
    <ClientOnly>
      <VueApexCharts
        v-if="hasData"
        type="donut"
        height="260"
        :options="chartOptions"
        :series="series"
      />
      <EmptyState
        v-else
        icon="chart"
        title="No class data yet"
        message="Object class breakdown will appear once detections are reported for the selected camera."
        class="py-6"
      />
      <template #fallback>
        <div class="flex h-[260px] items-center justify-center">
          <Skeleton width="8rem" height="0.85rem" />
        </div>
      </template>
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import EmptyState from '~/components/ui/EmptyState.vue'
import Skeleton from '~/components/ui/Skeleton.vue'
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

const totalObjects = computed(() => series.value.reduce((a, b) => a + b, 0))

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
          name: { color: '#9ca3af', fontSize: '12px' },
          value: { color: '#e5e7eb', fontSize: '18px', fontWeight: 600 },
          total: {
            show: true,
            label: 'Objects',
            color: '#6b7280',
            fontSize: '11px',
            formatter: () => String(totalObjects.value),
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
