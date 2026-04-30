<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-xl font-semibold text-gray-100">System health</h1>
      <p class="mt-1 text-sm text-gray-500">Latency trends and high-latency incidents</p>
    </header>

    <div class="flex flex-col gap-4 lg:flex-row lg:flex-wrap lg:items-end">
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-500 uppercase tracking-wide">Camera</label>
        <select
          v-model="selectedCamera"
          class="min-w-[200px] rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
        >
          <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-500 uppercase tracking-wide">Range</label>
        <select
          v-model="rangePreset"
          class="min-w-[200px] rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
        >
          <option value="5m">Last 5 minutes</option>
          <option value="15m">Last 15 minutes</option>
          <option value="1h">Last 1 hour</option>
          <option value="6h">Last 6 hours</option>
          <option value="24h">Last 24 hours</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-500 uppercase tracking-wide">Threshold (ms)</label>
        <input
          v-model.number="threshold"
          type="number"
          min="1"
          step="1"
          class="w-32 rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
          @change="onThresholdCommitted"
        >
      </div>
      <button
        type="button"
        class="rounded-lg bg-teal-600 hover:bg-teal-500 text-white text-sm font-medium px-4 py-2 self-end"
        :disabled="loading"
        @click="refresh"
      >
        {{ loading ? 'Loading…' : 'Refresh' }}
      </button>
    </div>

    <div v-if="pageError" class="rounded-lg border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-300">
      {{ pageError }}
    </div>

    <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-xl border border-gray-800 bg-[#12181f] p-4">
        <p class="text-xs text-gray-500 uppercase tracking-wide">Avg latency</p>
        <p class="mt-2 text-2xl font-semibold tabular-nums text-gray-100">{{ fmt(stats.avg) }}</p>
      </div>
      <div class="rounded-xl border border-gray-800 bg-[#12181f] p-4">
        <p class="text-xs text-gray-500 uppercase tracking-wide">P95 latency</p>
        <p class="mt-2 text-2xl font-semibold tabular-nums text-gray-100">{{ fmt(stats.p95) }}</p>
      </div>
      <div class="rounded-xl border border-gray-800 bg-[#12181f] p-4">
        <p class="text-xs text-gray-500 uppercase tracking-wide">P99 latency</p>
        <p class="mt-2 text-2xl font-semibold tabular-nums text-gray-100">{{ fmt(stats.p99) }}</p>
      </div>
      <div class="rounded-xl border border-gray-800 bg-[#12181f] p-4">
        <p class="text-xs text-gray-500 uppercase tracking-wide">High-latency frames</p>
        <p class="mt-2 text-2xl font-semibold tabular-nums text-red-300">{{ stats.incidentCount ?? '—' }}</p>
      </div>
    </section>

    <div class="grid gap-6 xl:grid-cols-2">
      <div class="rounded-xl border border-gray-800 bg-[#12181f] p-5">
        <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wide mb-4">Latency timeline</h3>
        <ClientOnly>
          <VueApexCharts
            v-if="latencySeries[0]?.data?.length"
            type="line"
            height="320"
            :options="latencyChartOptions"
            :series="latencySeries"
          />
          <div v-else class="h-[320px] flex items-center justify-center text-gray-500 text-sm">No latency samples</div>
          <template #fallback><div class="h-[320px] flex items-center justify-center text-gray-500">Loading…</div></template>
        </ClientOnly>
      </div>
      <div class="rounded-xl border border-gray-800 bg-[#12181f] p-5">
        <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wide mb-4">Latency distribution</h3>
        <ClientOnly>
          <VueApexCharts
            v-if="latencyHistogram.counts.length"
            type="bar"
            height="320"
            :options="histOptions"
            :series="histSeries"
          />
          <div v-else class="h-[320px] flex items-center justify-center text-gray-500 text-sm">Not enough data for histogram</div>
          <template #fallback><div class="h-[320px] flex items-center justify-center text-gray-500">Loading…</div></template>
        </ClientOnly>
      </div>
    </div>

    <div class="rounded-xl border border-gray-800 bg-[#12181f] overflow-hidden">
      <div class="px-5 py-4 border-b border-gray-800">
        <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wide">High latency incidents</h3>
        <p class="text-xs text-gray-500 mt-1">Frames where latency exceeded {{ threshold }} ms</p>
      </div>
      <div class="overflow-x-auto">
        <table class="min-w-full text-left text-sm">
          <thead class="bg-gray-900/50 text-xs uppercase text-gray-500">
            <tr>
              <th class="px-4 py-3 cursor-pointer hover:text-gray-300" @click="toggleSort('time')">Time {{ sortArrow('time') }}</th>
              <th class="px-4 py-3 cursor-pointer hover:text-gray-300" @click="toggleSort('camera')">Camera {{ sortArrow('camera') }}</th>
              <th class="px-4 py-3 text-right cursor-pointer hover:text-gray-300" @click="toggleSort('latency')">Latency (ms) {{ sortArrow('latency') }}</th>
              <th class="px-4 py-3 text-right cursor-pointer hover:text-gray-300" @click="toggleSort('detections')">Detections {{ sortArrow('detections') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-800/80">
            <tr v-if="!sortedRows.length">
              <td colspan="4" class="px-4 py-8 text-center text-gray-500">No incidents in this range</td>
            </tr>
            <tr v-for="(row, i) in sortedRows" :key="i" class="hover:bg-gray-900/30">
              <td class="px-4 py-2.5 font-mono text-gray-300 whitespace-nowrap">{{ row.time }}</td>
              <td class="px-4 py-2.5 text-gray-300">{{ row.camera }}</td>
              <td class="px-4 py-2.5 text-right tabular-nums text-red-300">{{ row.latency.toFixed(2) }}</td>
              <td class="px-4 py-2.5 text-right tabular-nums text-gray-200">{{ row.detections }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatAxisInteger, useChartTheme } from '~/composables/useChartTheme'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const api = useApi()
const { seriesColors, base } = useChartTheme()

const cameras = ref<string[]>([])
const selectedCamera = ref('')
const rangePreset = ref<'5m' | '15m' | '1h' | '6h' | '24h'>('1h')
const threshold = ref(500)

const loading = ref(false)
const pageError = ref<string | null>(null)

const latencies = ref<{ t: string; ms: number }[]>([])

const incidentRows = ref<Array<{ time: string; camera: string; latency: number; detections: number }>>([])

const stats = ref({
  avg: null as number | null,
  p95: null as number | null,
  p99: null as number | null,
  incidentCount: null as number | null,
})

type SortKey = 'time' | 'camera' | 'latency' | 'detections'
const sortKey = ref<SortKey>('time')
const sortDir = ref<'asc' | 'desc'>('desc')

const latencySeries = computed(() => [{
  name: 'Latency (ms)',
  data: latencies.value.map(p => ({ x: p.t, y: p.ms })),
}])

const latencyChartOptions = computed(() => ({
  ...base,
  colors: [seriesColors[1] ?? '#3b82f6'],
  chart: { ...base.chart, type: 'line', animations: { enabled: false } },
  stroke: { width: 2, curve: 'smooth' },
  xaxis: { type: 'datetime', labels: { datetimeUTC: true } },
  yaxis: {
    title: { text: 'ms', style: { color: '#6b7280' } },
    labels: {
      formatter: formatAxisInteger,
    },
  },
  tooltip: {
    y: {
      formatter: (val: number) => `${formatAxisInteger(val)} ms`,
    },
  },
  annotations: {
    yaxis: [{
      y: threshold.value,
      borderColor: '#f87171',
      borderWidth: 2,
      strokeDashArray: 4,
      label: {
        text: `${formatAxisInteger(threshold.value)} ms`,
        style: { color: '#f87171', background: 'transparent' },
      },
    }],
  },
}))

const latencyHistogram = computed(() => {
  const values = latencies.value.map(l => l.ms)
  if (values.length < 2) return { labels: [] as string[], counts: [] as number[] }
  const sorted = [...values].sort((a, b) => a - b)
  const min = sorted[0]!
  const max = sorted[sorted.length - 1]!
  const bins = 12
  const step = Math.max((max - min) / bins, 1e-6)
  const labels: string[] = []
  const counts = new Array(bins).fill(0) as number[]
  for (let i = 0; i < bins; i++) {
    const lo = min + i * step
    const hi = i === bins - 1 ? max : min + (i + 1) * step
    labels.push(`${Math.round(lo)}–${Math.round(hi)}`)
  }
  for (const v of values) {
    let idx = Math.floor((v - min) / step)
    if (idx >= bins) idx = bins - 1
    if (idx < 0) idx = 0
    counts[idx] = (counts[idx] ?? 0) + 1
  }
  return { labels, counts }
})

const histSeries = computed(() => [{
  name: 'Frames',
  data: latencyHistogram.value.counts,
}])

const histOptions = computed(() => ({
  ...base,
  colors: [seriesColors[4] ?? '#34d399'],
  chart: { ...base.chart, type: 'bar' },
  plotOptions: { bar: { columnWidth: '85%', borderRadius: 2 } },
  xaxis: { categories: latencyHistogram.value.labels },
  yaxis: {
    labels: { formatter: formatAxisInteger },
  },
}))

const sortedRows = computed(() => {
  const rows = [...incidentRows.value]
  const dir = sortDir.value === 'asc' ? 1 : -1
  const key = sortKey.value
  rows.sort((a, b) => {
    let cmp = 0
    if (key === 'time') cmp = a.time.localeCompare(b.time)
    else if (key === 'camera') cmp = a.camera.localeCompare(b.camera)
    else if (key === 'latency') cmp = a.latency - b.latency
    else cmp = a.detections - b.detections
    return cmp * dir
  })
  return rows
})

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  }
  else {
    sortKey.value = key
    sortDir.value = key === 'time' ? 'desc' : 'asc'
  }
}

function sortArrow(key: SortKey) {
  if (sortKey.value !== key) return ''
  return sortDir.value === 'asc' ? '▲' : '▼'
}

function fmt(n: number | null) {
  if (n == null || Number.isNaN(n)) return '—'
  return n.toFixed(2)
}

function computeRange() {
  const end = new Date()
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

function percentile(sorted: number[], p: number): number | null {
  if (!sorted.length) return null
  const idx = (p / 100) * (sorted.length - 1)
  const lo = Math.floor(idx)
  const hi = Math.ceil(idx)
  if (lo === hi) return sorted[lo]!
  return sorted[lo]! + (sorted[hi]! - sorted[lo]!) * (idx - lo)
}

async function loadCameras() {
  const res = await api.getCameras()
  cameras.value = res.cameras ?? []
  if (!selectedCamera.value && cameras.value[0]) selectedCamera.value = cameras.value[0]!
}

async function refresh() {
  pageError.value = null
  if (!selectedCamera.value) {
    pageError.value = 'Select a camera.'
    return
  }
  loading.value = true
  const { start, end } = computeRange()
  const startIso = start.toISOString()
  const endIso = end.toISOString()
  try {
    const [metricsRes, healthRes] = await Promise.all([
      api.getMetrics(selectedCamera.value, startIso),
      api.getHealth({
        camera: selectedCamera.value,
        start: startIso,
        end: endIso,
        threshold: threshold.value,
      }),
    ])

    const d = metricsRes.detections?.[0]
    const times = d && Array.isArray(d.recorded_at) ? d.recorded_at as string[] : []
    const lats = d && Array.isArray(d.latency_ms) ? (d.latency_ms as number[]).map(Number) : []
    latencies.value = times.map((t, i) => ({ t, ms: lats[i] ?? 0 })).filter(p => p.t)

    const nums = [...lats].filter(n => !Number.isNaN(n)).sort((a, b) => a - b)
    const avg = nums.length ? nums.reduce((a, b) => a + b, 0) / nums.length : null
    stats.value.avg = avg
    stats.value.p95 = nums.length ? percentile(nums, 95) : null
    stats.value.p99 = nums.length ? percentile(nums, 99) : null

    const h = healthRes.health_summary
    const n = h?.recorded_at?.length ?? 0
    stats.value.incidentCount = n
    incidentRows.value = []
    for (let i = 0; i < n; i++) {
      incidentRows.value.push({
        time: new Date(h.recorded_at[i]!).toLocaleString(),
        camera: h.camera[i] ?? '',
        latency: Number(h.latency_ms[i]),
        detections: Number(h.total_detections[i]),
      })
    }
  }
  catch (e: unknown) {
    pageError.value = e instanceof Error ? e.message : 'Failed to load health data'
  }
  finally {
    loading.value = false
  }
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

/** Refetch health only when the user commits the threshold (change/blur), not on every v-model tick. */
function onThresholdCommitted() {
  if (selectedCamera.value) refresh()
}
</script>
