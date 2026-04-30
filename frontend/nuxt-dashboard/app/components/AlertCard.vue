<template>
  <article
    class="rounded-xl border p-4 transition-all duration-300"
    :class="[
      borderClass,
      highlight ? 'ring-1 ring-amber-400/40 scale-[1.01]' : '',
    ]"
  >
    <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div class="flex gap-3 min-w-0">
        <div
          class="shrink-0 w-2 self-stretch rounded-full"
          :class="statusStripeClass"
          aria-hidden="true"
        ></div>
        <div class="min-w-0 space-y-1.5">
          <div class="flex flex-wrap items-center gap-2 text-sm">
            <time
              :title="exactTime"
              class="text-gray-200 font-mono"
            >{{ relativeTime }}</time>
            <span
              class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium border"
              :class="statusBadgeClass"
            >{{ alert.status }}</span>
          </div>
          <p class="text-sm text-gray-300">
            <span class="text-gray-500">Camera: </span> {{ alert.camera }}
            <span class="text-gray-600 mx-1.5">·</span>
            <span class="text-gray-500">Zone: </span> {{ alert.zone_name || '—' }}
          </p>
          <p class="text-sm text-gray-400">
            <span class="text-gray-500">Object: </span> {{ alert.label }}
            <span class="text-gray-600 mx-1.5">·</span>
            <span class="text-gray-500">Tracker: </span>
            <span class="font-mono text-gray-200">{{ alert.tracker_id }}</span>
          </p>
          <p v-if="alert.alert_type" class="text-xs text-gray-500">
            Type: <span class="uppercase tracking-wide text-gray-300">{{ alert.alert_type }}</span>
          </p>
          <p v-if="alert.severity" class="text-xs text-gray-500">
            Severity: <span class="uppercase tracking-wide" :class="severityTextClass">{{ alert.severity }}</span>
          </p>
          <p v-if="bboxText" class="text-xs text-gray-500 font-mono truncate" :title="bboxText">
            Box:  {{ bboxText }}
          </p>
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-2 shrink-0 sm:pl-2">
        <button
          v-if="alert.status === 'new' && canUpdateStatus"
          type="button"
          class="rounded-lg border border-amber-700/60 bg-amber-950/30 px-3 py-1.5 text-sm font-medium text-amber-200 hover:bg-amber-950/50 transition-colors"
          :disabled="busy"
          @click="$emit('acknowledge', alert.id)"
        >
          {{ busy ? '…' : 'Acknowledge' }}
        </button>
        <button
          v-if="alert.status === 'acknowledged' && canUpdateStatus"
          type="button"
          class="rounded-lg border border-emerald-800/60 bg-emerald-950/20 px-3 py-1.5 text-sm font-medium text-emerald-200/90 hover:bg-emerald-950/40 transition-colors"
          :disabled="busy"
          @click="$emit('resolve', alert.id)"
        >
          {{ busy ? '…' : 'Resolve' }}
        </button>
        <span
          v-if="alert.status === 'resolved'"
          class="text-xs text-gray-500"
        >Closed</span>
        <span
          v-else-if="!canUpdateStatus"
          class="rounded border border-gray-700 bg-gray-900/60 px-2 py-1 text-xs text-gray-400"
        >Live event</span>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { formatRelativeAgo } from '~/composables/useRelativeTime'
import type { SecurityAlert } from '~/types/security'

const props = defineProps<{
  alert: SecurityAlert
  busyId: number | null
  /** Brief highlight when a new row appears */
  highlight?: boolean
  /** Tick to refresh relative time labels */
  nowTick?: number
}>()

defineEmits<{
  acknowledge: [id: number]
  resolve: [id: number]
}>()

const busy = computed(() => props.busyId === props.alert.id)
const canUpdateStatus = computed(() => props.alert.id > 0)

const relativeTime = computed(() => {
  props.nowTick
  return formatRelativeAgo(props.alert.recorded_at)
})

const exactTime = computed(() => {
  const d = new Date(props.alert.recorded_at)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString(undefined, { dateStyle: 'full', timeStyle: 'long' })
})

const bboxText = computed(() => {
  const b = props.alert.bound_box
  if (!b || b.length < 4) return ''
  return b.map((n) => n.toFixed(0)).join(', ')
})

const borderClass = computed(() => {
  if (props.alert.status === 'new') return 'border-red-900/50 bg-red-950/10'
  if (props.alert.status === 'acknowledged') return 'border-amber-900/40 bg-amber-950/5'
  return 'border-gray-800 bg-[#12181f]/80'
})

const statusStripeClass = computed(() => {
  if (props.alert.status === 'new') return 'bg-red-500'
  if (props.alert.status === 'acknowledged') return 'bg-amber-400'
  return 'bg-gray-500'
})

const statusBadgeClass = computed(() => {
  if (props.alert.status === 'new') return 'border-red-800/50 bg-red-950/30 text-red-200'
  if (props.alert.status === 'acknowledged') return 'border-amber-800/50 bg-amber-950/30 text-amber-200'
  return 'border-gray-600 bg-gray-900/60 text-gray-400'
})

const severityTextClass = computed(() => {
  const severity = props.alert.severity?.toLowerCase()
  if (severity === 'critical') return 'text-red-300'
  if (severity === 'high') return 'text-orange-300'
  if (severity === 'medium') return 'text-amber-300'
  return 'text-gray-300'
})
</script>
