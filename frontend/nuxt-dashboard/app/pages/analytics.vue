<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Insights"
      title="Analytics"
      subtitle="Historical detection metrics for the selected camera and time range."
    />

    <FilterToolbar>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="analytics-camera">Camera</label>
        <select
          id="analytics-camera"
          v-model="selectedCamera"
          class="app-input min-w-[200px]"
        >
          <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="analytics-range">Range</label>
        <select
          id="analytics-range"
          v-model="rangePreset"
          class="app-input min-w-[200px]"
        >
          <option value="5m">Last 5 minutes</option>
          <option value="15m">Last 15 minutes</option>
          <option value="1h">Last 1 hour</option>
          <option value="6h">Last 6 hours</option>
          <option value="24h">Last 24 hours</option>
          <option value="custom">Custom</option>
        </select>
      </div>
      <div v-if="rangePreset === 'custom'" class="flex flex-wrap gap-3">
        <div class="flex flex-col gap-1">
          <label class="app-label">Start</label>
          <input
            v-model="customStart"
            type="datetime-local"
            class="app-input"
          >
        </div>
        <div class="flex flex-col gap-1">
          <label class="app-label">End</label>
          <input
            v-model="customEnd"
            type="datetime-local"
            class="app-input"
          >
        </div>
      </div>
      <template #actions>
        <button
          type="button"
          class="btn-primary"
          :disabled="loading"
          @click="refresh"
        >
          {{ loading ? 'Loading…' : 'Refresh' }}
        </button>
      </template>
    </FilterToolbar>

    <div v-if="pageError" class="app-banner-error">{{ pageError }}</div>

    <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div v-for="card in summaryCards" :key="card.label" class="kpi-card app-card app-card-hover">
        <p class="app-label">{{ card.label }}</p>
        <p class="mt-2 text-2xl font-semibold tabular-nums text-gray-100">
          <AnimatedNumber v-if="card.num != null" :value="card.num" :decimals="card.decimals" />
          <span v-else>—</span>
        </p>
        <div v-if="card.spark.length > 1" class="mt-2">
          <Sparkline :data="card.spark" :color="card.color" />
        </div>
      </div>
    </section>

    <div class="grid gap-6 xl:grid-cols-2">
      <div class="app-card p-5">
        <h3 class="app-section-title mb-4">Detections over time</h3>
        <ClientOnly>
          <VueApexCharts v-if="timelineSeries.length" type="area" height="320" :options="timelineOptions" :series="timelineSeries" />
          <div v-else class="h-[320px] flex items-center justify-center text-gray-500 text-sm">No timeline data</div>
          <template #fallback><div class="h-[320px] p-1"><Skeleton width="100%" height="100%" rounded="rounded-lg" /></div></template>
        </ClientOnly>
      </div>
      <div class="app-card p-5">
        <h3 class="app-section-title mb-4">Class counts over time</h3>
        <ClientOnly>
          <VueApexCharts v-if="classSeries.length" type="line" height="320" :options="classChartOptions" :series="classSeries" />
          <div v-else class="h-[320px] flex items-center justify-center text-gray-500 text-sm">No per-class series</div>
          <template #fallback><div class="h-[320px] p-1"><Skeleton width="100%" height="100%" rounded="rounded-lg" /></div></template>
        </ClientOnly>
      </div>
    </div>

    <div class="app-card p-5">
      <h3 class="app-section-title mb-4">Total detections by class (range)</h3>
      <ClientOnly>
        <VueApexCharts v-if="barSeries.length" type="bar" height="360" :options="barOptions" :series="barSeries" />
        <div v-else class="h-[360px] flex items-center justify-center text-gray-500 text-sm">No class aggregates</div>
        <template #fallback><div class="h-[360px] p-1"><Skeleton width="100%" height="100%" rounded="rounded-lg" /></div></template>
      </ClientOnly>
    </div>
  </div>
</template>

<script setup lang="ts">
import Skeleton from '~/components/ui/Skeleton.vue'
import AnimatedNumber from '~/components/ui/AnimatedNumber.vue'
import Sparkline from '~/components/ui/Sparkline.vue'
import { formatAxisInteger, useChartTheme } from '~/composables/useChartTheme'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const api = useApi()
const { seriesColors, base } = useChartTheme()

const cameras = ref<string[]>([])
const selectedCamera = ref('')
const rangePreset = ref<'5m' | '15m' | '1h' | '6h' | '24h' | 'custom'>('1h')
const customStart = ref('')
const customEnd = ref('')

const loading = ref(false)
const pageError = ref<string | null>(null)

const metricsData = ref<{
  times: string[]
  totals: number[]
  objects: Record<string, number>[]
} | null>(null)

const summaryStats = ref<{
  avgDetections: number | null
  peakDetections: number | null
  totalDetections: number | null
  avgLatency: number | null
}>({
  avgDetections: null,
  peakDetections: null,
  totalDetections: null,
  avgLatency: null,
})

const classTotals = ref<Record<string, number>>({})

function bucketForPreset(): string {
  switch (rangePreset.value) {
    case '5m':
      return '5m'
    case '15m':
      return '15m'
    case '1h':
      return '1h'
    case '6h':
      return '6h'
    case '24h':
      return '24h'
    default:
      return '5m'
  }
}

function computeRange(): { start: Date; end: Date } | null {
  const end = new Date()
  if (rangePreset.value === 'custom') {
    if (!customStart.value || !customEnd.value) return null
    const s = new Date(customStart.value)
    const e = new Date(customEnd.value)
    if (Number.isNaN(s.getTime()) || Number.isNaN(e.getTime()) || s >= e) return null
    return { start: s, end: e }
  }
  const ms: Record<string, number> = {
    '5m': 5 * 60 * 1000,
    '15m': 15 * 60 * 1000,
    '1h': 60 * 60 * 1000,
    '6h': 6 * 60 * 60 * 1000,
    '24h': 24 * 60 * 60 * 1000,
  }
  const delta = ms[rangePreset.value] ?? 60 * 60 * 1000
  return { start: new Date(end.getTime() - delta), end }
}

const totalsSpark = computed(() => metricsData.value?.totals ?? [])

const summaryCards = computed(() => [
  { label: 'Avg detections / frame', num: summaryStats.value.avgDetections, decimals: 2, spark: totalsSpark.value, color: '#2dd4bf' },
  { label: 'Peak detections', num: summaryStats.value.peakDetections, decimals: 0, spark: totalsSpark.value, color: '#22d3ee' },
  { label: 'Total detections', num: summaryStats.value.totalDetections, decimals: 0, spark: totalsSpark.value, color: '#818cf8' },
  { label: 'Avg latency (ms)', num: summaryStats.value.avgLatency, decimals: 2, spark: [] as number[], color: '#fbbf24' },
])

const timelineSeries = computed(() => {
  const m = metricsData.value
  if (!m?.times.length) return []
  return [{ name: 'Detections', data: m.totals.map((y, i) => ({ x: m.times[i]!, y })) }]
})

const timelineOptions = computed(() => ({
  ...base,
  colors: [seriesColors[1] ?? '#3b82f6'],
  chart: { ...base.chart, type: 'area' },
  fill: {
    type: 'gradient',
    gradient: { shadeIntensity: 1, opacityFrom: 0.35, opacityTo: 0.05 },
  },
  xaxis: {
    type: 'datetime',
    labels: { datetimeUTC: true },
  },
  yaxis: {
    title: { text: 'Count', style: { color: '#6b7280' } },
    labels: { formatter: formatAxisInteger },
  },
  tooltip: {
    y: { formatter: formatAxisInteger },
  },
}))

const classSeries = computed(() => {
  const m = metricsData.value
  if (!m?.times.length || !m.objects.length) return []
  const keys = new Set<string>()
  for (const o of m.objects) {
    Object.keys(o).forEach(k => keys.add(k))
  }
  const ordered = [...keys].sort()
  return ordered.map((key, idx) => ({
    name: key,
    color: seriesColors[idx % seriesColors.length],
    data: m.times.map((t, i) => ({
      x: t,
      y: m.objects[i]?.[key] ?? 0,
    })),
  }))
})

const classChartOptions = computed(() => ({
  ...base,
  colors: classSeries.value.map((_, i) => seriesColors[i % seriesColors.length]!),
  chart: { ...base.chart, type: 'line' },
  stroke: { width: 2, curve: 'smooth' },
  xaxis: { type: 'datetime', labels: { datetimeUTC: true } },
  yaxis: { labels: { formatter: formatAxisInteger } },
  tooltip: { y: { formatter: formatAxisInteger } },
  legend: { ...base.legend, position: 'bottom' },
}))

const sortedClassEntries = computed(() =>
  Object.entries(classTotals.value).sort((a, b) => b[1] - a[1]),
)

const barSeries = computed(() => {
  if (!sortedClassEntries.value.length) return []
  return [{
    name: 'Detections',
    data: sortedClassEntries.value.map(([, v]) => v),
  }]
})

const barOptions = computed(() => ({
  ...base,
  colors: [seriesColors[0] ?? '#22d3ee'],
  chart: { ...base.chart, type: 'bar' },
  plotOptions: { bar: { horizontal: true, borderRadius: 4 } },
  xaxis: {
    categories: sortedClassEntries.value.map(([k]) => k),
    labels: { formatter: formatAxisInteger },
  },
  tooltip: { y: { formatter: formatAxisInteger } },
}))

async function loadCameras() {
  const res = await api.getCameras()
  cameras.value = res.cameras ?? []
  if (!selectedCamera.value && cameras.value[0]) selectedCamera.value = cameras.value[0]!
}

function parseMetrics(res: Awaited<ReturnType<typeof api.getMetrics>>) {
  const d = res.detections?.[0]
  if (!d) {
    metricsData.value = null
    return
  }
  const times = Array.isArray(d.recorded_at) ? d.recorded_at : []
  const totals = Array.isArray(d.total_detections)
    ? d.total_detections.map(Number)
    : [Number(d.total_detections)]
  const objects = Array.isArray(d.total_objects) ? d.total_objects : [d.total_objects as Record<string, number>]
  metricsData.value = { times, totals, objects }
}

async function refresh() {
  pageError.value = null
  const range = computeRange()
  if (!selectedCamera.value) {
    pageError.value = 'Select a camera.'
    return
  }
  if (!range) {
    pageError.value = 'Choose valid custom start and end times.'
    return
  }
  loading.value = true
  try {
    const startIso = range.start.toISOString()
    const endIso = range.end.toISOString()

    const [metricsRes, summaryRes, classesRes] = await Promise.all([
      api.getMetrics(selectedCamera.value, startIso),
      api.getSummary({
        camera: selectedCamera.value,
        bucket: bucketForPreset(),
        start: startIso,
        end: endIso,
      }),
      api.getClasses(selectedCamera.value, startIso, endIso),
    ])

    parseMetrics(metricsRes)

    const agg = summaryRes.aggregated_summary?.[0]
    if (agg) {
      const avgD = avg(agg.avg_total_detections)
      const peak = agg.max_total_detections?.length ? Math.max(...agg.max_total_detections) : null
      const total = agg.sum_total_detections?.reduce((a, b) => a + b, 0) ?? null
      const avgL = avg(agg.avg_latency_ms)
      summaryStats.value = {
        avgDetections: avgD,
        peakDetections: peak,
        totalDetections: total,
        avgLatency: avgL,
      }
    }
    else {
      summaryStats.value = { avgDetections: null, peakDetections: null, totalDetections: null, avgLatency: null }
    }

    classTotals.value = classesRes.class_summary ?? {}
  }
  catch (e: unknown) {
    pageError.value = e instanceof Error ? e.message : 'Failed to load analytics'
  }
  finally {
    loading.value = false
  }
}

function avg(arr: number[] | undefined) {
  if (!arr?.length) return null
  const s = arr.reduce((a, b) => a + b, 0)
  return s / arr.length
}

onMounted(async () => {
  try {
    await loadCameras()
    await refresh()
  }
  catch (e: unknown) {
    pageError.value = e instanceof Error ? e.message : 'Failed to initialize'
  }
})

watch([selectedCamera, rangePreset], () => {
  if (selectedCamera.value) refresh()
})
</script>
