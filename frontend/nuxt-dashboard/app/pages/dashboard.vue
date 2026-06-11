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

        <!-- Quick Activity Feed -->
        <div class="app-card overflow-hidden">
          <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3">
            <h3 class="app-section-title">Recent Activity</h3>
            <NuxtLink to="/alerts" class="text-xs text-teal-300 transition-colors hover:text-teal-200 hover:underline">
              View all alerts
            </NuxtLink>
          </div>
          <div v-if="recentActivity.length" class="divide-y divide-white/[0.06]">
            <div
              v-for="(event, i) in recentActivity"
              :key="i"
              class="flex items-center gap-3 px-5 py-3 transition-colors hover:bg-white/[0.02]"
            >
              <span
                class="h-2 w-2 shrink-0 rounded-full"
                :class="event.dotClass"
              />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm text-gray-300">{{ event.text }}</p>
                <p class="text-xs text-gray-500">{{ event.time }}</p>
              </div>
              <span class="shrink-0 text-xs font-mono tabular-nums text-gray-500">{{ event.camera }}</span>
            </div>
          </div>
          <div v-else class="px-5 py-8 text-center text-sm text-gray-500">
            No recent activity. Detection events will appear here as they stream in.
          </div>
        </div>
      </div>
      <div class="space-y-6 xl:sticky xl:top-24 xl:self-start">
        <CameraStatusList
          :rows="statusRows"
          :selected-camera="selectedCamera"
          :loading="statusLoading"
          :error="statusError"
          @select="selectedCamera = $event"
        />

        <!-- Quick Links Card -->
        <div class="app-card p-4">
          <h3 class="app-section-title mb-3">Quick Actions</h3>
          <div class="grid grid-cols-2 gap-2">
            <NuxtLink
              to="/cameras"
              class="flex items-center gap-2 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-2.5 text-sm text-gray-300 transition-all duration-200 hover:border-white/10 hover:bg-white/[0.04] hover:text-gray-100"
            >
              <svg class="h-4 w-4 shrink-0 text-teal-400" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
              Cameras
            </NuxtLink>
            <NuxtLink
              to="/zones"
              class="flex items-center gap-2 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-2.5 text-sm text-gray-300 transition-all duration-200 hover:border-white/10 hover:bg-white/[0.04] hover:text-gray-100"
            >
              <svg class="h-4 w-4 shrink-0 text-cyan-400" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 12l3.5-5h7L19 12l-3.5 5h-7L5 12z" /></svg>
              Zones
            </NuxtLink>
            <NuxtLink
              to="/analytics"
              class="flex items-center gap-2 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-2.5 text-sm text-gray-300 transition-all duration-200 hover:border-white/10 hover:bg-white/[0.04] hover:text-gray-100"
            >
              <svg class="h-4 w-4 shrink-0 text-indigo-400" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75z" /></svg>
              Analytics
            </NuxtLink>
            <NuxtLink
              to="/health"
              class="flex items-center gap-2 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-2.5 text-sm text-gray-300 transition-all duration-200 hover:border-white/10 hover:bg-white/[0.04] hover:text-gray-100"
            >
              <svg class="h-4 w-4 shrink-0 text-emerald-400" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z" /></svg>
              Health
            </NuxtLink>
          </div>
        </div>
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

interface ActivityEvent {
  text: string
  time: string
  camera: string
  dotClass: string
}

const recentActivity = computed<ActivityEvent[]>(() => {
  const events: ActivityEvent[] = []
  for (const row of statusRows.value) {
    if (row.recordedAtIso) {
      const age = Date.now() - new Date(row.recordedAtIso).getTime()
      const isOnline = age < 120_000
      events.push({
        text: isOnline
          ? `${row.totalDetections ?? 0} detections at ${row.latencyMs?.toFixed(0) ?? '?'}ms latency`
          : `Camera went stale ${Math.floor(age / 60_000)}m ago`,
        time: row.recordedAt ?? '',
        camera: row.camera,
        dotClass: isOnline ? 'bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.5)]' : 'bg-amber-400',
      })
    } else {
      events.push({
        text: 'No detection data available',
        time: 'Never',
        camera: row.camera,
        dotClass: 'bg-gray-600',
      })
    }
  }
  return events.slice(0, 8)
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
