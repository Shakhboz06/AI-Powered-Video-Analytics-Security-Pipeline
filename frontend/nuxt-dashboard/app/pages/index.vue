<template>
  <div class="space-y-6">
    <header class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-xl font-semibold text-gray-100">Live monitoring</h1>
        <p class="mt-1 text-sm text-gray-500">
          Camera list, snapshot, and status use the refresh interval below. Changing camera refreshes the snapshot immediately.
        </p>
      </div>
      <div class="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end sm:justify-end">
        <div class="flex items-center gap-3">
          <label class="text-xs text-gray-500 uppercase tracking-wide shrink-0">Camera</label>
          <select
            v-model="selectedCamera"
            class="min-w-[200px] rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-teal-500/40 focus:border-teal-600"
            :disabled="camerasLoading && !cameras.length"
          >
            <option v-if="!cameras.length" value="">No cameras</option>
            <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
        <div class="flex items-center gap-3">
          <label class="text-xs text-gray-500 uppercase tracking-wide shrink-0">Refresh</label>
          <select
            v-model.number="pollIntervalMs"
            class="min-w-[200px] rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200 focus:outline-none focus:ring-2 focus:ring-teal-500/40 focus:border-teal-600"
            @change="onPollIntervalChange"
          >
            <option v-for="o in POLL_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </div>
      </div>
    </header>

    <div class="grid gap-6 xl:grid-cols-3">
      <div class="xl:col-span-2 space-y-6">
        <LiveSnapshotCard
          :camera="selectedCamera"
          :snapshot="liveSnapshot"
          :loading="liveLoading"
          :error="liveError"
          :last-updated="lastUpdatedFormatted"
        />
        <ClassDonutChart :total-objects="liveSnapshot?.total_objects ?? null" />
      </div>
      <div>
        <CameraStatusList
          :rows="statusRows"
          :selected-camera="selectedCamera"
          :loading="statusLoading"
          :error="statusError"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import LiveSnapshotCard from '~/components/security/LiveSnapshotCard.vue'
import ClassDonutChart from '~/components/security/ClassDonutChart.vue'
import CameraStatusList from '~/components/security/CameraStatusList.vue'
import type { CameraStatusRow } from '~/types/security'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const api = useApi()

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

let pollTimer: ReturnType<typeof setInterval> | undefined

function isValidPollMs(ms: number) {
  return POLL_OPTIONS.some(o => o.value === ms)
}

async function runPollTick() {
  await loadCameras()
  await loadLive()
  await loadStatusForCameras()
}

function startPollTimer() {
  if (pollTimer) clearInterval(pollTimer)
  pollTimer = setInterval(runPollTick, pollIntervalMs.value)
}

function onPollIntervalChange() {
  if (import.meta.client) {
    localStorage.setItem(POLL_MS_STORAGE_KEY, String(pollIntervalMs.value))
  }
  startPollTimer()
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
            } satisfies CameraStatusRow
          }
          return {
            camera,
            totalDetections: d.total_detections,
            latencyMs: d.latency_ms,
            recordedAt: new Date(d.recorded_at).toLocaleString(),
          } satisfies CameraStatusRow
        }
        catch {
          return {
            camera,
            totalDetections: null,
            latencyMs: null,
            recordedAt: null,
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
  await loadCameras()
  await loadLive()
  await loadStatusForCameras()
  startPollTimer()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

watch(selectedCamera, () => {
  loadLive()
})

// Surface camera list errors in console; optional banner could use camerasError
watchEffect(() => {
  if (camerasError.value) {
    console.warn(camerasError.value)
  }
})
</script>
