<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Incident Feed"
      title="Alerts"
      subtitle="Security incidents and zone breaches streamed live from the backend."
    />

    <FilterToolbar>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="alert-camera">Camera</label>
        <select
          id="alert-camera"
          v-model="filterCamera"
          class="app-input min-w-[200px]"
          :disabled="camerasLoading && !cameras.length"
        >
          <option value="">All cameras</option>
          <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="alert-status">Status</label>
        <select
          id="alert-status"
          v-model="filterStatus"
          class="app-input min-w-[200px]"
        >
          <option v-for="o in STATUS_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="alert-type">Type</label>
        <select
          id="alert-type"
          v-model="filterType"
          class="app-input min-w-[200px]"
        >
          <option v-for="o in ALERT_TYPE_FILTER_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
      </div>
      <template #actions>
        <span
          class="app-chip"
          :class="streamStatusClass"
        >
          <span class="h-1.5 w-1.5 rounded-full" :class="streamDotClass" />
          {{ streamStatusLabel }}
        </span>
        <label class="inline-flex cursor-pointer items-center gap-2 text-sm text-gray-400">
          <input
            v-model="soundOn"
            type="checkbox"
            class="rounded border-gray-600"
            @change="onSoundChange"
          >
          Sound for new
        </label>
        <div class="flex flex-wrap gap-2 text-xs">
          <span class="app-chip border-red-800/50 bg-red-950/20 text-red-200">{{ counts.new }} new</span>
          <span class="app-chip border-amber-800/50 bg-amber-950/15 text-amber-200/90">{{ counts.acknowledged }} ack</span>
          <span class="app-chip text-gray-400">{{ counts.resolved }} resolved</span>
        </div>
      </template>
    </FilterToolbar>

    <div class="grid gap-3 grid-cols-2 sm:grid-cols-4 xl:grid-cols-8">
      <div
        v-for="card in statCards"
        :key="card.key"
        class="kpi-card app-card app-card-hover"
        :class="card.cardClass"
      >
        <div class="flex items-center gap-2">
          <svg
            class="h-4 w-4 shrink-0"
            :class="card.accentText"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.6"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path v-for="(d, i) in card.iconPaths" :key="i" :d="d" />
          </svg>
          <span class="text-xs uppercase tracking-wide text-gray-400">{{ card.label }}</span>
        </div>
        <p class="mt-2 text-2xl font-semibold tabular-nums text-gray-100">
          <AnimatedNumber :value="card.count" />
        </p>
      </div>
    </div>

    <div
      v-if="notifiedFlash"
      class="rounded-lg border border-teal-800/50 bg-teal-950/20 px-3 py-2 text-sm text-teal-200/90 animate-pulse"
      role="status"
    >
      New alert received
    </div>

    <div v-if="pageError" class="app-banner-error">{{ pageError }}</div>

    <ul v-if="loading && !alerts.length" class="space-y-3">
      <li v-for="i in 4" :key="i" class="app-card p-4">
        <div class="flex items-start gap-3">
          <Skeleton width="0.5rem" height="3.5rem" rounded="rounded-full" />
          <div class="flex-1 space-y-2.5">
            <div class="flex gap-2">
              <Skeleton width="5rem" height="1.25rem" />
              <Skeleton width="6rem" height="1.25rem" />
            </div>
            <Skeleton width="60%" height="0.85rem" />
            <Skeleton width="45%" height="0.85rem" />
          </div>
        </div>
      </li>
    </ul>
    <div v-else-if="!filteredAlerts.length" class="app-card">
      <EmptyState
        icon="shield"
        title="No alerts for this view"
        message="You're all clear. New incidents will stream in here live as they are detected."
      />
    </div>
    <ul v-else class="space-y-3">
      <li v-for="a in filteredAlerts" :key="a.id">
        <AlertCard
          :alert="a"
          :busy-id="busyId"
          :highlight="!!flashIds[a.id]"
          :now-tick="nowTick"
          @acknowledge="onAck"
          @resolve="onResolve"
          @open="openDrawer"
        />
      </li>
    </ul>

    <AlertDetailDrawer
      :alert="selectedAlert"
      :busy-id="busyId"
      @close="selectedAlertId = null"
      @acknowledge="onAck"
      @resolve="onResolve"
    />
  </div>
</template>

<script setup lang="ts">
import AlertCard from '~/components/AlertCard.vue'
import AlertDetailDrawer from '~/components/security/AlertDetailDrawer.vue'
import AnimatedNumber from '~/components/ui/AnimatedNumber.vue'
import EmptyState from '~/components/ui/EmptyState.vue'
import Skeleton from '~/components/ui/Skeleton.vue'
import {
  ALERT_TYPE_FILTER_OPTIONS,
  getAlertMeta,
  normalizeAlertType,
} from '~/composables/useAlertMeta'
import type { SecurityAlert } from '~/types/security'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const STATUS_OPTIONS = [
  { value: 'all', label: 'All' },
  { value: 'new', label: 'New' },
  { value: 'acknowledged', label: 'Acknowledged' },
  { value: 'resolved', label: 'Resolved' },
] as const

const SOUND_KEY = 'alerts_sound_on'

/** Collapse rule-based + ML alerts of the same kind for one event within this window. */
const DEDUPE_WINDOW_MS = 5_000

const api = useApi()
const toast = useToast()

const cameras = ref<string[]>([])
const camerasLoading = ref(true)
const filterCamera = ref('')
const filterStatus = ref<string>('all')
const filterType = ref<string>('all')

const alerts = ref<SecurityAlert[]>([])
const loading = ref(true)
const pageError = ref<string | null>(null)
const busyId = ref<number | null>(null)

const selectedAlertId = ref<number | null>(null)
const selectedAlert = computed(() => alerts.value.find((a) => a.id === selectedAlertId.value) ?? null)

function openDrawer(a: SecurityAlert) {
  selectedAlertId.value = a.id
}

const soundOn = ref(true)
/** Ids to highlight briefly; using object + reassign for reliable reactivity */
const flashIds = ref<Record<number, true>>({})
const notifiedFlash = ref(false)

const nowTick = ref(0)
let relInterval: ReturnType<typeof setInterval> | null = null
let alertStream: EventSource | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempt = 0
let syntheticAlertId = -1
const streamStatus = ref<'connecting' | 'connected' | 'disconnected'>('connecting')

onMounted(() => {
  if (import.meta.client) {
    const v = localStorage.getItem(SOUND_KEY)
    if (v === '0') soundOn.value = false
  }
  loadCameras()
  const rel = setInterval(() => { nowTick.value++ }, 30_000)
  relInterval = rel
  void startAlertsFeed()
})

onUnmounted(() => {
  if (relInterval) clearInterval(relInterval)
  if (alertStream) alertStream.close()
  if (reconnectTimer) clearTimeout(reconnectTimer)
})

function onSoundChange() {
  if (import.meta.client) {
    localStorage.setItem(SOUND_KEY, soundOn.value ? '1' : '0')
  }
}

const cameraFilteredAlerts = computed(() => {
  if (!filterCamera.value) return alerts.value
  return alerts.value.filter((a) => a.camera === filterCamera.value)
})

/**
 * Rule-based and ML detectors can both emit a "falling" alert for the same fall.
 * Collapse near-simultaneous falling alerts that share a camera + tracker so the
 * same event is shown once (keeps the highest-severity / most recent instance).
 */
function dedupeFalling(list: SecurityAlert[]): SecurityAlert[] {
  const severityRank = (s?: string | null) => {
    const v = (s ?? '').toLowerCase()
    if (v === 'critical') return 3
    if (v === 'high') return 2
    if (v === 'medium') return 1
    return 0
  }
  const kept: SecurityAlert[] = []
  // newest-first list; track the last kept falling event per camera+tracker
  const lastByKey = new Map<string, { idx: number; time: number }>()
  for (const a of list) {
    if (normalizeAlertType(a.alert_type) !== 'falling') {
      kept.push(a)
      continue
    }
    const key = `${a.camera}|${a.tracker_id}`
    const time = new Date(a.recorded_at).getTime()
    const prev = lastByKey.get(key)
    if (prev && Math.abs(prev.time - time) <= DEDUPE_WINDOW_MS) {
      const existing = kept[prev.idx]!
      if (severityRank(a.severity) > severityRank(existing.severity)) {
        kept[prev.idx] = a
      }
      continue
    }
    kept.push(a)
    lastByKey.set(key, { idx: kept.length - 1, time })
  }
  return kept
}

const baseAlerts = computed(() => dedupeFalling(cameraFilteredAlerts.value))

const counts = computed(() => {
  return baseAlerts.value.reduce(
    (acc, a) => {
      if (a.status === 'new') acc.new++
      else if (a.status === 'acknowledged') acc.acknowledged++
      else acc.resolved++
      return acc
    },
    { new: 0, acknowledged: 0, resolved: 0 },
  )
})

const STAT_CARD_TYPES = ['fighting', 'falling', 'intrusion', 'running', 'brandishing', 'abandoned_object', 'loitering']

interface StatCard {
  key: string
  label: string
  count: number
  iconPaths: string[]
  accentText: string
  cardClass: string
}

const statCards = computed<StatCard[]>(() => {
  const typeCount = new Map<string, number>()
  for (const a of baseAlerts.value) {
    const key = normalizeAlertType(a.alert_type)
    if (key) typeCount.set(key, (typeCount.get(key) ?? 0) + 1)
  }
  const cards: StatCard[] = STAT_CARD_TYPES.map((key) => {
    const meta = getAlertMeta(key)
    return {
      key,
      label: meta.label,
      count: typeCount.get(key) ?? 0,
      iconPaths: meta.iconPaths,
      accentText: meta.accentText,
      cardClass: key === 'fighting'
        ? 'border-orange-500/40 bg-orange-500/[0.08]'
        : '',
    }
  })
  cards.push({
    key: 'total',
    label: 'Total alerts',
    count: baseAlerts.value.length,
    iconPaths: getAlertMeta(null).iconPaths,
    accentText: 'text-gray-300',
    cardClass: '',
  })
  return cards
})

const filteredAlerts = computed(() => {
  return baseAlerts.value.filter((a) => {
    if (filterStatus.value !== 'all' && a.status !== filterStatus.value) return false
    if (filterType.value !== 'all' && normalizeAlertType(a.alert_type) !== filterType.value) return false
    return true
  })
})

const streamStatusLabel = computed(() => {
  if (streamStatus.value === 'connected') return 'Live'
  if (streamStatus.value === 'connecting') return 'Connecting'
  return 'Reconnecting'
})

const streamStatusClass = computed(() => {
  if (streamStatus.value === 'connected') return 'border-emerald-800/50 bg-emerald-950/20 text-emerald-200'
  if (streamStatus.value === 'connecting') return 'border-amber-800/50 bg-amber-950/20 text-amber-200'
  return 'border-red-800/50 bg-red-950/20 text-red-200'
})

const streamDotClass = computed(() => {
  if (streamStatus.value === 'connected') return 'bg-emerald-400'
  if (streamStatus.value === 'connecting') return 'bg-amber-400 animate-pulse'
  return 'bg-red-400 animate-pulse'
})

function playBeep() {
  if (!import.meta.client || !soundOn.value) return
  const w = globalThis as unknown as { AudioContext?: typeof AudioContext; webkitAudioContext?: typeof AudioContext }
  const Ctx = w.AudioContext ?? w.webkitAudioContext
  if (!Ctx) return
  try {
    const ctx = new Ctx()
    const o = ctx.createOscillator()
    const g = ctx.createGain()
    o.type = 'sine'
    o.frequency.setValueAtTime(880, ctx.currentTime)
    o.connect(g)
    g.connect(ctx.destination)
    g.gain.setValueAtTime(0.12, ctx.currentTime)
    g.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.15)
    o.start()
    o.stop(ctx.currentTime + 0.15)
    setTimeout(() => { ctx.close().catch(() => {}) }, 200)
  }
  catch { /* empty */ }
}

function markFlashing(ids: number[]) {
  if (!ids.length) return
  const m = { ...flashIds.value }
  for (const id of ids) m[id] = true
  flashIds.value = m
  setTimeout(() => {
    const m2 = { ...flashIds.value }
    for (const id of ids) delete m2[id]
    flashIds.value = m2
  }, 2_200)
  notifiedFlash.value = true
  setTimeout(() => { notifiedFlash.value = false }, 2_500)
}

async function loadCameras() {
  camerasLoading.value = true
  try {
    const res = await api.getCameras()
    cameras.value = res.cameras
  }
  catch {
    /* optional */
  }
  finally {
    camerasLoading.value = false
  }
}

async function loadInitialAlerts() {
  loading.value = true
  try {
    const { alerts: list } = await api.listAlerts({ status: 'all' })
    alerts.value = list
    pageError.value = null
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to load initial alerts'
  }
  finally {
    loading.value = false
  }
}

type StreamAlertPayload = Partial<SecurityAlert> & {
  id?: number
  alert_id?: string
  AlertId?: string
  alertId?: string
  zone_id?: number | null
  ZoneID?: number | null
  tracker_id?: number
  TrackerID?: number
  bound_box?: [number, number, number, number]
  BoundBox?: [number, number, number, number]
  recorded_at?: string
  RecordedAt?: string
  alert_type?: string
  AlertType?: string
  severity?: string
  Severity?: string
  camera?: string
  Camera?: string
  label?: string
  Label?: string
}

function normalizeStreamAlert(raw: StreamAlertPayload): SecurityAlert {
  const id = typeof raw.id === 'number' && raw.id > 0 ? raw.id : syntheticAlertId--
  const zoneId = raw.zone_id ?? raw.ZoneID ?? null
  const trackerId = raw.tracker_id ?? raw.TrackerID ?? 0
  const boundBox = raw.bound_box ?? raw.BoundBox ?? ([0, 0, 0, 0] as [number, number, number, number])
  const recordedAt = raw.recorded_at ?? raw.RecordedAt ?? new Date().toISOString()

  return {
    id,
    alert_id: raw.alert_id ?? raw.AlertId ?? raw.alertId ?? null,
    camera: raw.camera ?? raw.Camera ?? 'unknown',
    zone_name: raw.zone_name ?? '',
    zone_id: zoneId,
    tracker_id: trackerId,
    bound_box: boundBox,
    label: raw.label ?? raw.Label ?? 'object',
    status: raw.status ?? 'new',
    alert_type: raw.alert_type ?? raw.AlertType ?? null,
    severity: raw.severity ?? raw.Severity ?? null,
    recorded_at: recordedAt,
  }
}

function upsertAlert(alert: SecurityAlert) {
  const idx = alerts.value.findIndex((a) => a.id === alert.id)
  if (idx >= 0) {
    alerts.value = [
      ...alerts.value.slice(0, idx),
      { ...alerts.value[idx]!, ...alert },
      ...alerts.value.slice(idx + 1),
    ]
    return
  }

  alerts.value = [alert, ...alerts.value].slice(0, 250)
  if (alert.status === 'new') {
    playBeep()
    markFlashing([alert.id])
    const meta = getAlertMeta(alert.alert_type)
    toast.warning(`${meta.label} detected`, {
      description: `${alert.camera}${alert.zone_name ? ` · ${alert.zone_name}` : ''}`,
    })
  }
}

async function startAlertsFeed() {
  await loadInitialAlerts()
  connectAlertStream()
}

function connectAlertStream() {
  if (!import.meta.client) return
  streamStatus.value = 'connecting'
  pageError.value = null

  if (alertStream) alertStream.close()
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }

  const stream = new EventSource(api.alertStreamUrl(), {withCredentials: true})
  alertStream = stream

  stream.onopen = () => {
    reconnectAttempt = 0
    streamStatus.value = 'connected'
    pageError.value = null
  }

  stream.onmessage = (event) => {
    loading.value = false
    pageError.value = null
    try {
      const payload = JSON.parse(event.data) as StreamAlertPayload
      upsertAlert(normalizeStreamAlert(payload))
    }
    catch {
      pageError.value = 'Received invalid alert stream payload'
    }
  }

  stream.onerror = () => {
    streamStatus.value = 'disconnected'
    pageError.value = 'Alert stream disconnected; reconnecting…'
    stream.close()
    if (alertStream === stream) alertStream = null
    scheduleReconnect()
  }
}

function scheduleReconnect() {
  if (!import.meta.client || reconnectTimer) return
  reconnectAttempt += 1
  const delay = Math.min(30_000, 1000 * 2 ** Math.min(reconnectAttempt, 5))
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    connectAlertStream()
  }, delay)
}

async function onAck(id: number) {
  const i = alerts.value.findIndex((a) => a.id === id)
  if (i < 0) return
  const before = { ...alerts.value[i]! }
  const optimistic: SecurityAlert = { ...before, status: 'acknowledged' as const }
  alerts.value = [
    ...alerts.value.slice(0, i),
    optimistic,
    ...alerts.value.slice(i + 1),
  ]
  busyId.value = id
  try {
    const { alert } = await api.patchAlertStatus(id, 'acknowledged')
    const j = alerts.value.findIndex((a) => a.id === id)
    if (j >= 0) alerts.value[j] = alert
    toast.success('Alert acknowledged')
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Update failed'
    toast.error('Could not acknowledge alert')
    const j = alerts.value.findIndex((a) => a.id === id)
    if (j >= 0) alerts.value[j] = before
  }
  finally {
    busyId.value = null
  }
}

async function onResolve(id: number) {
  const i = alerts.value.findIndex((a) => a.id === id)
  if (i < 0) return
  const before = { ...alerts.value[i]! }
  const optimistic: SecurityAlert = { ...before, status: 'resolved' as const }
  alerts.value = [
    ...alerts.value.slice(0, i),
    optimistic,
    ...alerts.value.slice(i + 1),
  ]
  busyId.value = id
  try {
    const { alert } = await api.patchAlertStatus(id, 'resolved')
    const j = alerts.value.findIndex((a) => a.id === id)
    if (j >= 0) alerts.value[j] = alert
    toast.success('Alert resolved')
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Update failed'
    toast.error('Could not resolve alert')
    const j = alerts.value.findIndex((a) => a.id === id)
    if (j >= 0) alerts.value[j] = before
  }
  finally {
    busyId.value = null
  }
}
</script>
