<template>
  <article
    class="alert-card-premium app-card app-card-hover cursor-pointer p-4 md:p-5 transition-all duration-300"
    :class="[
      borderClass,
      highlight ? 'ring-1 ring-amber-400/50 scale-[1.01] shadow-[0_0_32px_-8px_rgba(251,191,36,0.35)]' : '',
    ]"
    role="button"
    tabindex="0"
    @click="$emit('open', alert)"
    @keydown.enter="$emit('open', alert)"
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
              v-if="alert.alert_type"
              class="inline-flex items-center gap-1 rounded-md px-2 py-0.5 text-xs font-medium border"
              :class="meta.badgeClass"
            >
              <svg
                class="h-3.5 w-3.5"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.6"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
              >
                <path v-for="(d, i) in meta.iconPaths" :key="i" :d="d" />
              </svg>
              {{ meta.label }}
            </span>
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
            <span class="text-gray-500">Object: </span> {{ objectLabel }}
            <span class="text-gray-600 mx-1.5">·</span>
            <template v-if="isSceneLevel">
              <span class="text-gray-500">Scope: </span>
              <span class="text-gray-200">Scene-wide</span>
            </template>
            <template v-else>
              <span class="text-gray-500">Tracker: </span>
              <span class="font-mono text-gray-200">{{ alert.tracker_id }}</span>
            </template>
          </p>
          <p v-if="alert.severity" class="text-xs text-gray-500">
            Severity: <span class="uppercase tracking-wide" :class="severityTextClass">{{ alert.severity }}</span>
          </p>
          <p v-if="bboxText" class="text-xs text-gray-500 font-mono truncate" :title="bboxText">
            Box:  {{ bboxText }}
          </p>
          <div v-if="hasBox" class="pt-1 max-w-[220px]">
            <BoundingBoxPreview
              :bound-box="alert.bound_box"
              :color="meta.boxColor"
              :label="boxLabel"
              :src="cameraImageUrl(alert.camera)"
            />
          </div>
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-2 shrink-0 sm:pl-2">
        <button
          v-if="alert.status === 'new' && canUpdateStatus"
          type="button"
          class="btn-ghost-sm border-amber-700/50 bg-amber-950/25 text-amber-200 hover:bg-amber-950/40"
          :disabled="busy"
          @click.stop="$emit('acknowledge', alert.id)"
        >
          {{ busy ? '…' : 'Acknowledge' }}
        </button>
        <button
          v-if="alert.status === 'acknowledged' && canUpdateStatus"
          type="button"
          class="btn-ghost-sm border-emerald-800/50 bg-emerald-950/20 text-emerald-200 hover:bg-emerald-950/35"
          :disabled="busy"
          @click.stop="$emit('resolve', alert.id)"
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
import { getAlertMeta, normalizeAlertType } from '~/composables/useAlertMeta'
import BoundingBoxPreview from '~/components/security/BoundingBoxPreview.vue'
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
  open: [alert: SecurityAlert]
}>()

const busy = computed(() => props.busyId === props.alert.id)
const canUpdateStatus = computed(() => props.alert.id > 0)

const meta = computed(() => getAlertMeta(props.alert.alert_type))

/** Scene-level alerts (e.g. fighting) carry tracker_id 0 — show scope, not a tracker. */
const isSceneLevel = computed(() => meta.value.sceneLevel === true || props.alert.tracker_id === 0)

const objectLabel = computed(() => {
  if (props.alert.label === 'person_group') return 'Group of people'
  return props.alert.label
})

const hasBox = computed(() => {
  const b = props.alert.bound_box
  return !!b && b.length >= 4 && (b[2] - b[0]) > 0 && (b[3] - b[1]) > 0
})

const boxLabel = computed(() => {
  const key = normalizeAlertType(props.alert.alert_type)
  if (key === 'fighting') return 'FIGHT'
  if (key === 'falling') return 'FALL'
  return (props.alert.label || meta.value.label).toUpperCase().slice(0, 12)
})

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
  if (props.alert.status === 'new') {
    if (normalizeAlertType(props.alert.alert_type) === 'fighting') return 'border-orange-800/60 bg-orange-950/15'
    return 'border-red-900/50 bg-red-950/10'
  }
  if (props.alert.status === 'acknowledged') return 'border-amber-900/40 bg-amber-950/5'
  return 'opacity-70'
})

/** Active (new/ack) alerts colour the stripe by type; resolved ones stay neutral. */
const statusStripeClass = computed(() => {
  if (props.alert.status === 'resolved') return 'bg-gray-500'
  if (props.alert.alert_type) return meta.value.stripeClass
  if (props.alert.status === 'new') return 'bg-red-500'
  return 'bg-amber-400'
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
