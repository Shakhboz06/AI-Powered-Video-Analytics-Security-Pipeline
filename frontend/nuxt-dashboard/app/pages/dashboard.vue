<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Operations"
      title="Live Monitoring"
      subtitle="Real-time detection feed, camera health, and class distribution across your fleet."
    />

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      <KpiCard
        v-for="kpi in kpiCards"
        :key="kpi.key"
        :label="kpi.label"
        :value="kpi.value"
        :suffix="kpi.suffix"
        :hint="kpi.hint"
        :decimals="kpi.decimals"
        :icon-path="kpi.iconPath"
        :icon-wrap-class="kpi.iconWrapClass"
        :accent-class="kpi.accentClass"
        :spark="kpi.spark"
        :spark-color="kpi.sparkColor"
      />
    </section>

    <!-- Controls -->
    <FilterToolbar>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="monitor-camera">Camera</label>
        <select
          id="monitor-camera"
          v-model="selectedCamera"
          class="app-input min-w-[200px]"
          :disabled="camerasLoading && !cameras.length"
        >
          <option v-if="!cameras.length" value="">No cameras</option>
          <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="monitor-interval">Auto-refresh</label>
        <select
          id="monitor-interval"
          v-model.number="pollIntervalMs"
          class="app-input min-w-[200px]"
          @change="onPollIntervalChange"
        >
          <option v-for="o in POLL_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
      </div>
      <template #actions>
        <span class="app-chip" :class="refreshing ? 'text-teal-300' : 'text-gray-400'">
          <span class="h-1.5 w-1.5 rounded-full" :class="refreshing ? 'live-dot bg-teal-400' : 'bg-gray-500'" />
          {{ refreshCountdownLabel }}
        </span>
        <button
          type="button"
          class="btn-primary"
          :disabled="refreshing"
          @click="manualRefresh"
        >
          <svg
            class="h-4 w-4"
            :class="refreshing ? 'animate-spin' : ''"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182" />
          </svg>
          {{ refreshing ? 'Refreshing…' : 'Refresh now' }}
        </button>
      </template>
    </FilterToolbar>

    <div v-if="camerasError" class="app-banner-warning">{{ camerasError }}</div>

    <!-- Main layout -->
    <div class="grid gap-6 xl:grid-cols-[1fr,min(380px,100%)]">
      <div class="space-y-6">
        <LiveSnapshotCard
          :camera="selectedCamera"
          :snapshot="liveSnapshot"
          :loading="liveLoading"
          :error="liveError"
          :last-updated="lastUpdatedFormatted"
        />
        <ClassDonutChart :total-objects="liveSnapshot?.total_objects ?? null" />
      </div>
      <div class="xl:sticky xl:top-24 xl:self-start">
        <CameraStatusList
          :rows="statusRows"
          :selected-camera="selectedCamera"
          :loading="statusLoading"
          :error="statusError"
          @select="selectedCamera = $event"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import LiveSnapshotCard from '~/components/security/LiveSnapshotCard.vue'
import ClassDonutChart from '~/components/security/ClassDonutChart.vue'
import CameraStatusList from '~/components/security/CameraStatusList.vue'
import PageHeader from '~/components/ui/PageHeader.vue'
import FilterToolbar from '~/components/ui/FilterToolbar.vue'
import KpiCard from '~/components/ui/KpiCard.vue'
import type { CameraStatusRow } from '~/types/security'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const api = useApi()
const toast = useToast()

const POLL_OPTIONS = [
  { label: '3 seconds', value: 3 * 1000 },
  { label: '1 minute', value: 60 * 1000 },
  { label: '5 minutes', value: 5 * 60 * 1000 },
  { label: '15 minutes', value: 15 * 60 * 1000 },
  { label: '30 minutes', value: 30 * 60 * 1000 },
  { label: '1 hour', value: 60 * 60 * 1000 },
] as const

const POLL_MS_STORAGE_KEY = 'live_dashboard_poll_ms'

const pollIntervalMs = ref<number>(5 * 60 * 1000)
const refreshCountdownSec = ref(0)
const refreshing = ref(false)
const detectionHistory = ref<number[]>([])
const latencyHistory = ref<number[]>([])

const cameras = ref<string[]>([])
const camerasLoading = ref(true)
const camerasError = ref<string | null>(null)

const selectedCamera = ref('')

const liveSnapshot = ref<{
  total_detections: number
  total_objects: Record<string, number>
  latency_ms: number
  recorded_at: string
} | null>(null)
const liveLoading = ref(false)
const liveError = ref<string | null>(null)
const lastUpdatedAt = ref<Date | null>(null)

const statusRows = ref<CameraStatusRow[]>([])
const statusLoading = ref(false)
const statusError = ref<string | null>(null)

const lastUpdatedFormatted = computed(() => {
  if (!lastUpdatedAt.value) return null
  return lastUpdatedAt.value.toLocaleString()
})

const refreshCountdownLabel = computed(() => {
  if (refreshing.value) return 'Updating…'
  const sec = refreshCountdownSec.value
  if (sec <= 0) return 'Next refresh soon'
  if (sec < 60) return `Next in ${sec}s`
  const min = Math.floor(sec / 60)
  const rem = sec % 60
  return rem ? `Next in ${min}m ${rem}s` : `Next in ${min}m`
})

const onlineCount = computed(() =>
  statusRows.value.filter(r => r.recordedAtIso && (Date.now() - new Date(r.recordedAtIso).getTime()) < 120_000).length,
)

const avgLatency = computed(() => {
  const vals = statusRows.value.map(r => r.latencyMs).filter((v): v is number => v != null)
  if (!vals.length) return null
  return vals.reduce((a, b) => a + b, 0) / vals.length
})

const kpiCards = computed(() => [
  {
    key: 'online',
    label: 'Cameras online',
    value: onlineCount.value,
    suffix: cameras.value.length ? ` / ${cameras.value.length}` : undefined,
    hint: cameras.value.length ? `${cameras.value.length - onlineCount.value} offline or stale` : 'No cameras configured',
    iconPath: 'M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.774 48.774 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316z',
    iconWrapClass: 'border-emerald-500/20 bg-emerald-950/30',
    accentClass: 'text-emerald-400',
    spark: [] as number[],
    sparkColor: '#34d399',
  },
  {
    key: 'detections',
    label: 'Active detections',
    value: liveSnapshot.value?.total_detections ?? 0,
    hint: selectedCamera.value ? `On ${selectedCamera.value}` : 'Select a camera',
    iconPath: 'M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z M15 12a3 3 0 11-6 0 3 3 0 016 0z',
    iconWrapClass: 'border-teal-500/20 bg-teal-950/30',
    accentClass: 'text-teal-400',
    spark: detectionHistory.value,
    sparkColor: '#2dd4bf',
  },
  {
    key: 'latency',
    label: 'Avg latency',
    value: avgLatency.value ?? 0,
    decimals: 1,
    suffix: ' ms',
    hint: avgLatency.value != null
      ? (avgLatency.value < 300 ? 'Fleet within normal range' : 'Elevated — check health')
      : 'Waiting for data',
    iconPath: 'M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z',
    iconWrapClass: 'border-cyan-500/20 bg-cyan-950/30',
    accentClass: 'text-cyan-400',
    spark: latencyHistory.value,
    sparkColor: '#22d3ee',
  },
  {
    key: 'classes',
    label: 'Object classes',
    value: Object.keys(liveSnapshot.value?.total_objects ?? {}).length,
    hint: 'Distinct types in latest frame',
    iconPath: 'M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z',
    iconWrapClass: 'border-indigo-500/20 bg-indigo-950/30',
    accentClass: 'text-indigo-400',
    spark: [] as number[],
    sparkColor: '#818cf8',
  },
])

let pollTimer: ReturnType<typeof setInterval> | undefined
let countdownTimer: ReturnType<typeof setInterval> | undefined

function isValidPollMs(ms: number) {
  return POLL_OPTIONS.some(o => o.value === ms)
}

function resetCountdown() {
  refreshCountdownSec.value = Math.round(pollIntervalMs.value / 1000)
}

function startCountdownTimer() {
  if (countdownTimer) clearInterval(countdownTimer)
  resetCountdown()
  countdownTimer = setInterval(() => {
    refreshCountdownSec.value = Math.max(0, refreshCountdownSec.value - 1)
  }, 1000)
}

async function runPollTick(silent = false) {
  if (!silent) refreshing.value = true
  try {
    await loadCameras()
    await loadLive()
    await loadStatusForCameras()
    resetCountdown()
  }
  finally {
    refreshing.value = false
  }
}

async function manualRefresh() {
  await runPollTick()
  toast.success('Live data refreshed')
}

function startPollTimer() {
  if (pollTimer) clearInterval(pollTimer)
  resetCountdown()
  pollTimer = setInterval(() => runPollTick(true), pollIntervalMs.value)
}

function onPollIntervalChange() {
  if (import.meta.client) {
    localStorage.setItem(POLL_MS_STORAGE_KEY, String(pollIntervalMs.value))
  }
  startPollTimer()
  startCountdownTimer()
}

async function loadCameras() {
  camerasLoading.value = true
  camerasError.value = null
  try {
    const res = await api.getCameras()
    cameras.value = res.cameras ?? []
    if (!selectedCamera.value && cameras.value.length) {
      selectedCamera.value = cameras.value[0]!
    }
    else if (selectedCamera.value && !cameras.value.includes(selectedCamera.value) && cameras.value.length) {
      selectedCamera.value = cameras.value[0]!
    }
  }
  catch (e: unknown) {
    camerasError.value = e instanceof Error ? e.message : 'Failed to load cameras'
  }
  finally {
    camerasLoading.value = false
  }
}

async function loadLive() {
  if (!selectedCamera.value) {
    liveSnapshot.value = null
    liveError.value = null
    return
  }
  liveLoading.value = true
  liveError.value = null
  try {
    const res = await api.getLive(selectedCamera.value)
    const d = res.detections?.[0]
    if (d) {
      liveSnapshot.value = {
        total_detections: d.total_detections,
        total_objects: d.total_objects ?? {},
        latency_ms: d.latency_ms,
        recorded_at: d.recorded_at,
      }
      lastUpdatedAt.value = new Date()
      detectionHistory.value = [...detectionHistory.value.slice(-19), d.total_detections]
      latencyHistory.value = [...latencyHistory.value.slice(-19), d.latency_ms]
    }
    else {
      liveSnapshot.value = null
    }
  }
  catch (e: unknown) {
    liveError.value = e instanceof Error ? e.message : 'Live feed failed'
    liveSnapshot.value = null
  }
  finally {
    liveLoading.value = false
  }
}

async function loadStatusForCameras() {
  if (!cameras.value.length) {
    statusRows.value = []
    return
  }
  statusLoading.value = true
  statusError.value = null
  try {
    const results = await Promise.all(
      cameras.value.map(async (camera) => {
        try {
          const res = await api.getLive(camera)
          const d = res.detections?.[0]
          if (!d) {
            return {
              camera,
              totalDetections: null,
              latencyMs: null,
              recordedAt: null,
              recordedAtIso: null,
            } satisfies CameraStatusRow
          }
          return {
            camera,
            totalDetections: d.total_detections,
            latencyMs: d.latency_ms,
            recordedAt: new Date(d.recorded_at).toLocaleString(),
            recordedAtIso: d.recorded_at,
          } satisfies CameraStatusRow
        }
        catch {
          return {
            camera,
            totalDetections: null,
            latencyMs: null,
            recordedAt: null,
            recordedAtIso: null,
          } satisfies CameraStatusRow
        }
      }),
    )
    statusRows.value = results
  }
  catch (e: unknown) {
    statusError.value = e instanceof Error ? e.message : 'Status refresh failed'
  }
  finally {
    statusLoading.value = false
  }
}

onMounted(async () => {
  if (import.meta.client) {
    const raw = localStorage.getItem(POLL_MS_STORAGE_KEY)
    if (raw != null) {
      const n = Number(raw)
      if (!Number.isNaN(n) && isValidPollMs(n)) pollIntervalMs.value = n
    }
  }
  await runPollTick(true)
  startPollTimer()
  startCountdownTimer()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (countdownTimer) clearInterval(countdownTimer)
})

watch(selectedCamera, () => {
  loadLive()
})
</script>
