<template>
  <div
    class="rounded-xl border p-5 md:p-6"
    :class="borderClasses"
  >
    <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
      <div>
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wide">Live snapshot</h2>
        <p v-if="camera" class="mt-1 text-xs text-gray-500">Camera: {{ camera }}</p>
      </div>
      <div v-if="lastUpdated" class="text-xs text-gray-500">
        Last updated: <span class="text-gray-300 font-mono">{{ lastUpdated }}</span>
      </div>
    </div>

    <div v-if="loading" class="mt-8 flex items-center justify-center py-12 text-gray-500">
      <span class="inline-flex items-center gap-2">
        <span class="h-2 w-2 rounded-full bg-teal-400 animate-pulse" />
        Loading live data…
      </span>
    </div>

    <div v-else-if="error" class="mt-6 rounded-lg bg-red-950/30 border border-red-900/50 px-4 py-3 text-sm text-red-300">
      {{ error }}
    </div>

    <div v-else-if="!snapshot" class="mt-8 py-12 text-center text-gray-500">
      No recent detections for this camera.
    </div>

    <div v-else class="mt-6 grid gap-6 sm:grid-cols-2">
      <div>
        <p class="text-xs text-gray-500 uppercase tracking-wide mb-2">Total detections</p>
        <p class="text-4xl md:text-5xl font-semibold tabular-nums text-gray-100">
          {{ snapshot.total_detections }}
        </p>
      </div>
      <div>
        <p class="text-xs text-gray-500 uppercase tracking-wide mb-2">Inference latency</p>
        <p
          class="text-3xl md:text-4xl font-semibold tabular-nums"
          :class="latencyColorClass(snapshot.latency_ms)"
        >
          {{ snapshot.latency_ms.toFixed(1) }}<span class="text-lg text-gray-500 ml-1">ms</span>
        </p>
        <p class="mt-2 text-xs" :class="latencyHintClass(snapshot.latency_ms)">
          {{ latencyHint(snapshot.latency_ms) }}
        </p>
      </div>
    </div>

    <div v-if="snapshot && Object.keys(snapshot.total_objects).length" class="mt-8">
      <p class="text-xs text-gray-500 uppercase tracking-wide mb-3">Class breakdown</p>
      <div class="flex flex-wrap gap-2">
        <span
          v-for="(count, cls) in sortedClasses(snapshot.total_objects)"
          :key="cls"
          class="inline-flex items-center gap-1.5 rounded-md border border-gray-700 bg-gray-900/50 px-2.5 py-1 text-sm text-gray-200"
        >
          <span class="text-gray-400">{{ cls }}</span>
          <span class="font-mono text-teal-300">{{ count }}</span>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { latencyBgClass, latencyColorClass } from '~/composables/useChartTheme'

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
}>()

const borderClasses = computed(() => {
  if (!props.snapshot) return 'border-gray-800 bg-[#12181f]'
  return ['border', latencyBgClass(props.snapshot.latency_ms)].join(' ')
})

function sortedClasses(obj: Record<string, number>) {
  return Object.fromEntries(
    Object.entries(obj).sort((a, b) => b[1] - a[1]),
  )
}

function latencyHint(ms: number) {
  if (ms < 300) return 'Within normal range'
  if (ms <= 500) return 'Elevated — monitor'
  return 'High latency — investigate'
}

function latencyHintClass(ms: number) {
  if (ms < 300) return 'text-emerald-500/90'
  if (ms <= 500) return 'text-amber-400/90'
  return 'text-red-400/90'
}
</script>
