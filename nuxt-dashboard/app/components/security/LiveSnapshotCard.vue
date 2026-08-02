<template>
  <div
    class="app-card overflow-hidden"
    :class="borderClasses"
  >
    <!-- Camera feed hero -->
    <div class="monitor-feed feed-hud relative aspect-video w-full overflow-hidden bg-[#05080c]">
      <span class="hud-corner hud-tl" aria-hidden="true" />
      <span class="hud-corner hud-tr" aria-hidden="true" />
      <span class="hud-corner hud-bl" aria-hidden="true" />
      <span class="hud-corner hud-br" aria-hidden="true" />
      <!--
        Continuous MJPEG stream from the dashboard hub. The browser renders the
        multipart/x-mixed-replace response as live video natively, so binding the
        src directly starts playback as soon as the first frame arrives.
      -->
      <img
        v-if="streamUrl"
        :src="feedSrc"
        :alt="`Live feed from ${camera || 'camera'}`"
        class="absolute inset-0 h-full w-full object-fill"
        @error="onStreamError"
      >
      <div
        v-else
        class="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-gradient-to-br from-gray-950 via-[#0a1018] to-gray-950"
      >
        <svg class="h-10 w-10 text-gray-600" fill="none" stroke="currentColor" stroke-width="1.2" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.774 48.774 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316z" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 11-9 0 4.5 4.5 0 019 0z" />
        </svg>
        <p class="text-xs text-gray-500">Select a camera to view its live feed</p>
      </div>

      <div class="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/85 via-black/25 to-black/10" />
      <div v-if="!loading && snapshot" class="scan-line pointer-events-none absolute left-0 right-0 h-px bg-teal-400/35" />

      <!-- Restricted-zone overlay -->
      <svg
        v-if="zones && zones.length"
        class="pointer-events-none absolute inset-0 h-full w-full"
        viewBox="0 0 640 480"
        preserveAspectRatio="xMidYMid slice"
        aria-hidden="true"
      >
        <polygon
          v-for="z in zones"
          :key="z.id"
          :points="polygonPoints(z.polygon)"
          fill="rgba(45, 212, 191, 0.14)"
          stroke="rgb(45, 212, 191)"
          stroke-width="2"
          vector-effect="non-scaling-stroke"
        />
      </svg>

      <!-- Top bar -->
      <div class="absolute inset-x-0 top-0 flex items-center justify-between gap-3 p-3 sm:p-4">
        <div class="flex items-center gap-2">
          <span
            class="app-chip border-red-500/30 bg-red-950/50 text-red-200 backdrop-blur-sm"
          >
            <span class="live-dot h-1.5 w-1.5 rounded-full bg-red-400" />
            Live
          </span>
          <span v-if="camera" class="hidden app-chip backdrop-blur-sm sm:inline-flex">
            {{ camera }}
          </span>
        </div>
        <div v-if="relativeUpdated" class="app-chip backdrop-blur-sm text-gray-300">
          <svg class="h-3 w-3 text-gray-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ relativeUpdated }}
        </div>
      </div>

      <!-- Bottom overlay stats -->
      <div
        v-if="snapshot && !loading"
        class="absolute inset-x-0 bottom-0 grid grid-cols-3 gap-px border-t border-white/[0.06] bg-black/50 backdrop-blur-md"
      >
        <div class="px-4 py-3">
          <p class="app-label">Detections</p>
          <p class="mt-0.5 text-xl font-semibold tabular-nums text-white sm:text-2xl">
            <AnimatedNumber :value="snapshot.total_detections" />
          </p>
        </div>
        <div class="px-4 py-3">
          <p class="app-label">Latency</p>
          <p class="mt-0.5 text-xl font-semibold tabular-nums sm:text-2xl" :class="latencyColorClass(snapshot.latency_ms)">
            <AnimatedNumber :value="snapshot.latency_ms" :decimals="1" />
            <span class="text-sm font-normal text-gray-400"> ms</span>
          </p>
        </div>
        <div class="px-4 py-3">
          <p class="app-label">Classes</p>
          <p class="mt-0.5 text-xl font-semibold tabular-nums text-white sm:text-2xl">
            {{ classCount }}
          </p>
        </div>
      </div>
    </div>

    <!-- Body -->
    <div class="p-5 md:p-6">
      <div v-if="error" class="rounded-xl border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-300">
        {{ error }}
      </div>

      <EmptyState
        v-else-if="!loading && !snapshot && camera"
        icon="camera"
        title="No recent detections"
        message="This camera has not reported detections in the latest poll window."
      />

      <EmptyState
        v-else-if="!loading && !camera"
        icon="camera"
        title="Select a camera"
        message="Choose a camera from the toolbar or status list to begin monitoring."
      />

      <template v-else-if="snapshot">
        <div v-if="Object.keys(snapshot.total_objects).length">
          <p class="app-section-title mb-3">Class breakdown</p>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(count, cls) in sortedClasses(snapshot.total_objects)"
              :key="cls"
              class="app-chip"
            >
              <span class="text-gray-400">{{ cls }}</span>
              <span class="font-mono font-semibold text-teal-300">{{ count }}</span>
            </span>
          </div>
        </div>
        <div v-else class="text-sm text-gray-500">
          No objects detected in the current frame.
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import AnimatedNumber from '~/components/ui/AnimatedNumber.vue'
import EmptyState from '~/components/ui/EmptyState.vue'
import { latencyBgClass, latencyColorClass } from '~/composables/useChartTheme'
import { formatRelativeAgo } from '~/composables/useRelativeTime'
import type { Point, SecurityZone } from '~/types/security'

const props = defineProps<{
  camera: string
  snapshot: {
    total_detections: number
    total_objects: Record<string, number>
    latency_ms: number
    recorded_at: string
  } | null
  loading: boolean
  error: string | null
  lastUpdated: string | null
  zones?: SecurityZone[]
}>()

function polygonPoints(polygon: Point[]): string {
  return polygon.map((p) => `${p.x},${p.y}`).join(' ')
}

const config = useRuntimeConfig()
const apiBase = (config.public.apiBase as string).replace(/\/$/, '')

const nowMs = ref(Date.now())
let tickTimer: ReturnType<typeof setInterval> | undefined

// Host-agnostic URL of the continuous MJPEG stream for the selected camera.
// Camera names can contain spaces (e.g. "room 1"), so the name is URL-encoded.
const streamUrl = computed(() => {
  if (!props.camera) return ''
  return `${apiBase}/api/v1/live/${encodeURIComponent(props.camera)}/stream`
})

// The value actually bound to <img>. It mirrors streamUrl, but reconnects append
// a cache-busting param so the browser opens a fresh connection.
const feedSrc = ref('')
let reconnectTimer: ReturnType<typeof setTimeout> | undefined

function clearReconnect() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = undefined
  }
}

function openStream(bustCache = false) {
  clearReconnect()
  if (!streamUrl.value) {
    feedSrc.value = ''
    return
  }
  feedSrc.value = bustCache ? `${streamUrl.value}?t=${Date.now()}` : streamUrl.value
}

// MJPEG streams don't reconnect on their own: if the backend restarts or the
// network blips, the <img> goes blank and stays blank. Re-open it after a short
// delay so the feed self-heals.
function onStreamError() {
  if (reconnectTimer) return
  reconnectTimer = setTimeout(() => {
    reconnectTimer = undefined
    openStream(true)
  }, 2000)
}

const classCount = computed(() => Object.keys(props.snapshot?.total_objects ?? {}).length)

const relativeUpdated = computed(() => {
  if (!props.snapshot?.recorded_at) return null
  return formatRelativeAgo(props.snapshot.recorded_at, nowMs.value)
})

const borderClasses = computed(() => {
  if (!props.snapshot) return ''
  return latencyBgClass(props.snapshot.latency_ms)
})

// Bind the stream immediately on mount and whenever the selected camera changes.
watch(streamUrl, () => openStream(), { immediate: true })

onMounted(() => {
  tickTimer = setInterval(() => { nowMs.value = Date.now() }, 15_000)
})
onUnmounted(() => {
  if (tickTimer) clearInterval(tickTimer)
  clearReconnect()
})

function sortedClasses(obj: Record<string, number>) {
  return Object.fromEntries(Object.entries(obj).sort((a, b) => b[1] - a[1]))
}
</script>
