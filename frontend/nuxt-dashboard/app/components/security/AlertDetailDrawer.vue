<template>
  <Teleport to="body">
    <Transition name="overlay">
      <div
        v-if="alert"
        class="fixed inset-0 z-[85] bg-black/60 backdrop-blur-sm"
        @click="$emit('close')"
      />
    </Transition>
    <Transition name="drawer">
      <aside
        v-if="alert"
        class="fixed inset-y-0 right-0 z-[86] flex w-full max-w-md flex-col border-l app-glass"
        style="border-color: var(--app-border)"
        role="dialog"
        aria-modal="true"
        aria-label="Alert details"
      >
        <header class="flex items-center justify-between gap-3 border-b px-5 py-4" style="border-color: var(--app-border)">
          <div class="flex items-center gap-2.5 min-w-0">
            <span class="grid h-9 w-9 shrink-0 place-items-center rounded-xl border" :class="meta.badgeClass" style="border-color: var(--app-border)">
              <svg class="h-4.5 w-4.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path v-for="(d, i) in meta.iconPaths" :key="i" :d="d" /></svg>
            </span>
            <div class="min-w-0">
              <h2 class="truncate text-base font-semibold text-gray-100">{{ meta.label }}</h2>
              <p class="text-xs text-gray-500">{{ relativeTime }}</p>
            </div>
          </div>
          <button type="button" class="rounded-lg p-2 text-gray-400 hover:bg-white/5 hover:text-gray-200" aria-label="Close" @click="$emit('close')">
            <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </header>

        <div class="flex-1 space-y-5 overflow-y-auto px-5 py-5 custom-scrollbar">
          <div class="flex flex-wrap gap-2">
            <span class="app-chip" :class="statusChipClass">{{ alert.status }}</span>
            <span v-if="alert.severity" class="app-chip" :class="severityChipClass">{{ alert.severity }} severity</span>
          </div>

          <div v-if="hasBox">
            <p class="app-section-title mb-2">Location</p>
            <BoundingBoxPreview :bound-box="alert.bound_box" :color="meta.boxColor" :label="boxLabel" :src="cameraImageUrl(alert.camera)" />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-xl border p-3" style="border-color: var(--app-border)">
              <p class="app-label">Camera</p>
              <p class="mt-1 truncate text-sm text-gray-200">{{ alert.camera }}</p>
            </div>
            <div class="rounded-xl border p-3" style="border-color: var(--app-border)">
              <p class="app-label">Zone</p>
              <p class="mt-1 truncate text-sm text-gray-200">{{ alert.zone_name || '—' }}</p>
            </div>
            <div class="rounded-xl border p-3" style="border-color: var(--app-border)">
              <p class="app-label">Object</p>
              <p class="mt-1 truncate text-sm text-gray-200">{{ objectLabel }}</p>
            </div>
            <div class="rounded-xl border p-3" style="border-color: var(--app-border)">
              <p class="app-label">{{ isSceneLevel ? 'Scope' : 'Tracker' }}</p>
              <p class="mt-1 truncate text-sm text-gray-200">{{ isSceneLevel ? 'Scene-wide' : `#${alert.tracker_id}` }}</p>
            </div>
          </div>

          <div class="rounded-xl border p-3" style="border-color: var(--app-border)">
            <p class="app-label">Recorded at</p>
            <p class="mt-1 font-mono text-sm text-gray-300">{{ exactTime }}</p>
          </div>

          <div v-if="bboxText" class="rounded-xl border p-3" style="border-color: var(--app-border)">
            <p class="app-label">Bounding box</p>
            <p class="mt-1 font-mono text-sm text-gray-300">[{{ bboxText }}]</p>
          </div>
        </div>

        <footer v-if="canUpdate" class="flex gap-2 border-t px-5 py-4" style="border-color: var(--app-border)">
          <button
            v-if="alert.status === 'new'"
            type="button"
            class="btn-ghost flex-1"
            :disabled="busy"
            @click="$emit('acknowledge', alert.id)"
          >{{ busy ? '…' : 'Acknowledge' }}</button>
          <button
            v-if="alert.status !== 'resolved'"
            type="button"
            class="btn-primary flex-1"
            :disabled="busy"
            @click="$emit('resolve', alert.id)"
          >{{ busy ? '…' : 'Resolve' }}</button>
          <span v-if="alert.status === 'resolved'" class="flex-1 text-center text-sm text-gray-500">Closed</span>
        </footer>
      </aside>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import BoundingBoxPreview from '~/components/security/BoundingBoxPreview.vue'
import { formatRelativeAgo } from '~/composables/useRelativeTime'
import { getAlertMeta, normalizeAlertType } from '~/composables/useAlertMeta'
import type { SecurityAlert } from '~/types/security'

const props = defineProps<{
  alert: SecurityAlert | null
  busyId: number | null
}>()

defineEmits<{
  close: []
  acknowledge: [id: number]
  resolve: [id: number]
}>()

const meta = computed(() => getAlertMeta(props.alert?.alert_type))
const busy = computed(() => !!props.alert && props.busyId === props.alert.id)
const canUpdate = computed(() => !!props.alert && props.alert.id > 0)
const isSceneLevel = computed(() => meta.value.sceneLevel === true || props.alert?.tracker_id === 0)
const objectLabel = computed(() => props.alert?.label === 'person_group' ? 'Group of people' : (props.alert?.label ?? '—'))

const hasBox = computed(() => {
  const b = props.alert?.bound_box
  return !!b && b.length >= 4 && (b[2] - b[0]) > 0 && (b[3] - b[1]) > 0
})

const boxLabel = computed(() => {
  const key = normalizeAlertType(props.alert?.alert_type)
  if (key === 'fighting') return 'FIGHT'
  if (key === 'falling') return 'FALL'
  return (props.alert?.label || meta.value.label).toUpperCase().slice(0, 12)
})

const bboxText = computed(() => {
  const b = props.alert?.bound_box
  if (!b || b.length < 4) return ''
  return b.map((n) => n.toFixed(0)).join(', ')
})

const relativeTime = computed(() => props.alert ? formatRelativeAgo(props.alert.recorded_at) : '')
const exactTime = computed(() => {
  if (!props.alert) return ''
  const d = new Date(props.alert.recorded_at)
  return Number.isNaN(d.getTime()) ? props.alert.recorded_at : d.toLocaleString(undefined, { dateStyle: 'full', timeStyle: 'long' })
})

const statusChipClass = computed(() => {
  if (props.alert?.status === 'new') return 'border-red-800/50 bg-red-950/30 text-red-200'
  if (props.alert?.status === 'acknowledged') return 'border-amber-800/50 bg-amber-950/30 text-amber-200'
  return 'text-gray-400'
})

const severityChipClass = computed(() => {
  const s = props.alert?.severity?.toLowerCase()
  if (s === 'critical') return 'border-red-800/50 text-red-300'
  if (s === 'high') return 'border-orange-800/50 text-orange-300'
  if (s === 'medium') return 'border-amber-800/50 text-amber-300'
  return 'text-gray-400'
})
</script>
