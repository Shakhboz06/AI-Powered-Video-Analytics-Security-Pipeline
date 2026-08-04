<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Operations"
      title="Live Monitoring"
      subtitle="Real-time detection feed, camera health, and class distribution across your fleet."
    />

    <ServiceHoursNotice />

    <section class="grid gap-3 sm:grid-cols-2">
      <KpiCard
        label="Cameras online"
        :value="onlineCount"
        :suffix="cameras.length ? ` / ${cameras.length}` : undefined"
        :hint="cameras.length ? `${cameras.length - onlineCount} offline or stale` : 'No cameras configured'"
        icon-path="M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.774 48.774 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316z"
        icon-wrap-class="border-emerald-500/20 bg-emerald-950/30"
        accent-class="text-emerald-400"
        spark-color="#34d399"
      />
      <KpiCard
        label="Fleet avg latency"
        :value="avgLatency ?? 0"
        :decimals="1"
        suffix=" ms"
        :hint="avgLatency != null ? (avgLatency < 300 ? 'Fleet within normal range' : 'Elevated — check health') : 'Waiting for data'"
        icon-path="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z"
        icon-wrap-class="border-cyan-500/20 bg-cyan-950/30"
        accent-class="text-cyan-400"
        :spark="latencyHistory"
        spark-color="#22d3ee"
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
          :zones="liveZones"
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
import ServiceHoursNotice from '~/components/ServiceHoursNotice.vue'
import FilterToolbar from '~/components/ui/FilterToolbar.vue'
import KpiCard from '~/components/ui/KpiCard.vue'
import type { CameraStatusRow, SecurityZone } from '~/types/security'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const api = useApi()

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
const liveZones = ref<SecurityZone[]>([])

const statusRows = ref<CameraStatusRow[]>([])
const statusLoading = ref(false)
const statusError = ref<string | null>(null)

const lastUpdatedFormatted = computed(() => {
  if (!lastUpdatedAt.value) return null
  return lastUpdatedAt.value.toLocaleString()
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

// The live video is now a continuous MJPEG stream (see LiveSnapshotCard), so the
// page no longer polls to refresh the feed. We load the supporting detection
// metadata (KPIs, camera status, class breakdown) once on mount; the camera
// watcher below refreshes the selected camera's snapshot when it changes.
onMounted(async () => {
  await loadCameras()
  await loadLive()
  await loadZones()
  await loadStatusForCameras()
})

async function loadZones() {
  if (!selectedCamera.value) {
    liveZones.value = []
    return
  }
  try {
    const { data } = await api.listZones(selectedCamera.value)
    liveZones.value = (data ?? []).filter((z) => z.is_active && z.polygon?.length >= 3)
  }
  catch {
    liveZones.value = []
  }
}

watch(selectedCamera, () => {
  loadLive()
  loadZones()
})
</script>
