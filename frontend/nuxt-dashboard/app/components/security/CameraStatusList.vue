<template>
  <div class="rounded-xl border border-gray-800 bg-[#12181f] p-4 md:p-5">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wide">Camera status</h3>
      <span v-if="loading" class="text-xs text-gray-500">Refreshing…</span>
    </div>
    <div v-if="error" class="text-sm text-red-400">{{ error }}</div>
    <ul v-else class="space-y-2 max-h-[320px] overflow-y-auto custom-scrollbar">
      <li v-if="!rows.length" class="text-sm text-gray-500 py-6 text-center">
        No cameras reported in the last hour.
      </li>
      <li
        v-for="row in rows"
        :key="row.camera"
        class="flex items-center justify-between gap-3 rounded-lg border border-gray-800/80 bg-gray-900/30 px-3 py-2.5"
        :class="row.camera === selectedCamera ? 'ring-1 ring-teal-500/40' : ''"
      >
        <div class="min-w-0">
          <p class="text-sm font-medium text-gray-200 truncate">{{ row.camera }}</p>
          <p v-if="row.recordedAt" class="text-xs text-gray-500 font-mono truncate">{{ row.recordedAt }}</p>
        </div>
        <div class="text-right shrink-0">
          <p class="text-sm tabular-nums text-gray-200">{{ row.totalDetections ?? '—' }}</p>
          <p
            class="text-xs tabular-nums"
            :class="row.latencyMs != null ? latencyColorClass(row.latencyMs) : 'text-gray-500'"
          >
            {{ row.latencyMs != null ? `${row.latencyMs.toFixed(0)} ms` : '—' }}
          </p>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { latencyColorClass } from '~/composables/useChartTheme'
import type { CameraStatusRow } from '~/types/security'

defineProps<{
  rows: CameraStatusRow[]
  selectedCamera: string
  loading: boolean
  error: string | null
}>()
</script>
