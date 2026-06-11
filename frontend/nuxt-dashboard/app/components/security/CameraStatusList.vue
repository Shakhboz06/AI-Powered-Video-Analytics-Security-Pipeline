<template>
  <div class="app-card flex flex-col p-4 md:p-5 xl:max-h-[calc(100vh-8rem)]">
    <div class="mb-3 flex items-center justify-between gap-3">
      <div>
        <h3 class="app-section-title">Camera status</h3>
        <p class="mt-0.5 text-xs text-gray-500">{{ rows.length }} configured · click to switch</p>
      </div>
      <span
        v-if="loading"
        class="inline-flex items-center gap-1.5 text-xs text-gray-500"
      >
        <svg class="h-3 w-3 animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
        Syncing
      </span>
    </div>

    <div v-if="error" class="rounded-lg border border-red-900/50 bg-red-950/30 px-3 py-2 text-sm text-red-300">
      {{ error }}
    </div>

    <div v-else-if="loading && !rows.length" class="flex-1 space-y-2">
      <div
        v-for="i in 5"
        :key="i"
        class="flex items-center justify-between gap-3 rounded-xl border border-white/[0.06] bg-white/[0.02] px-3 py-2.5"
      >
        <div class="min-w-0 flex-1 space-y-2">
          <Skeleton width="50%" height="0.85rem" />
          <Skeleton width="70%" height="0.7rem" />
        </div>
        <div class="space-y-2 text-right">
          <Skeleton width="2.5rem" height="0.85rem" />
          <Skeleton width="3rem" height="0.7rem" />
        </div>
      </div>
    </div>

    <EmptyState
      v-else-if="!rows.length"
      icon="camera"
      title="No cameras online"
      message="Configured cameras will appear here once they start reporting detections."
      class="flex-1 py-8"
    />

    <ul v-else class="custom-scrollbar -mx-1 flex-1 space-y-1.5 overflow-y-auto px-1">
      <li
        v-for="row in rows"
        :key="row.camera"
      >
        <button
          type="button"
          class="group flex w-full items-center justify-between gap-3 rounded-xl border px-3 py-2.5 text-left transition-all duration-200"
          :class="row.camera === selectedCamera
            ? 'border-teal-500/40 bg-teal-950/20 ring-1 ring-teal-500/30'
            : 'border-white/[0.06] bg-white/[0.02] hover:border-white/10 hover:bg-white/[0.04]'"
          @click="$emit('select', row.camera)"
        >
          <div class="flex min-w-0 items-center gap-2.5">
            <span
              class="h-2 w-2 shrink-0 rounded-full"
              :class="statusDotClass(row)"
              :title="statusLabel(row)"
            />
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-gray-200 group-hover:text-gray-100">
                {{ row.camera }}
              </p>
              <p v-if="row.recordedAtIso" class="truncate text-xs text-gray-500">
                {{ formatRelative(row.recordedAtIso) }}
              </p>
              <p v-else class="text-xs text-gray-600">No data</p>
            </div>
          </div>
          <div class="shrink-0 text-right">
            <p class="text-sm font-medium tabular-nums text-gray-200">
              {{ row.totalDetections ?? '—' }}
            </p>
            <p
              class="text-xs tabular-nums"
              :class="row.latencyMs != null ? latencyColorClass(row.latencyMs) : 'text-gray-600'"
            >
              {{ row.latencyMs != null ? `${row.latencyMs.toFixed(0)} ms` : 'offline' }}
            </p>
          </div>
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import Skeleton from '~/components/ui/Skeleton.vue'
import EmptyState from '~/components/ui/EmptyState.vue'
import { latencyColorClass } from '~/composables/useChartTheme'
import { formatRelativeAgo } from '~/composables/useRelativeTime'
import type { CameraStatusRow } from '~/types/security'

defineProps<{
  rows: CameraStatusRow[]
  selectedCamera: string
  loading: boolean
  error: string | null
}>()

defineEmits<{
  select: [camera: string]
}>()

const nowMs = ref(Date.now())
let tickTimer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  tickTimer = setInterval(() => { nowMs.value = Date.now() }, 15_000)
})
onUnmounted(() => {
  if (tickTimer) clearInterval(tickTimer)
})

function formatRelative(iso: string) {
  return formatRelativeAgo(iso, nowMs.value)
}

function statusFor(row: CameraStatusRow): 'online' | 'stale' | 'offline' {
  if (!row.recordedAtIso) return 'offline'
  const age = nowMs.value - new Date(row.recordedAtIso).getTime()
  if (age < 120_000) return 'online'
  if (age < 600_000) return 'stale'
  return 'offline'
}

function statusDotClass(row: CameraStatusRow) {
  const s = statusFor(row)
  if (s === 'online') return 'bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.6)]'
  if (s === 'stale') return 'bg-amber-400'
  return 'bg-gray-600'
}

function statusLabel(row: CameraStatusRow) {
  const s = statusFor(row)
  if (s === 'online') return 'Online'
  if (s === 'stale') return 'Stale data'
  return 'Offline'
}
</script>
