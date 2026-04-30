<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-xl font-semibold text-gray-100">Alerts</h1>
      <p class="mt-1 text-sm text-gray-500">Security incidents and zone breaches streamed live from the backend.</p>
    </header>

    <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div class="flex flex-wrap items-end gap-3">
        <div class="flex flex-col gap-1">
          <label class="text-xs text-gray-500 uppercase tracking-wide">Camera</label>
          <select
            v-model="filterCamera"
            class="min-w-[200px] rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
            :disabled="camerasLoading && !cameras.length"
          >
            <option value="">All cameras</option>
            <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-xs text-gray-500 uppercase tracking-wide">Status</label>
          <select
            v-model="filterStatus"
            class="min-w-[180px] rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
          >
            <option v-for="o in STATUS_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-3">
        <span
          class="inline-flex items-center gap-1.5 rounded border px-2.5 py-1 text-xs"
          :class="streamStatusClass"
        >
          <span class="h-1.5 w-1.5 rounded-full" :class="streamDotClass" />
          {{ streamStatusLabel }}
        </span>
        <label class="inline-flex items-center gap-2 text-sm text-gray-400 cursor-pointer">
          <input
            v-model="soundOn"
            type="checkbox"
            class="rounded border-gray-600"
            @change="onSoundChange"
          >
          Sound for new
        </label>
        <div class="flex flex-wrap gap-2 text-xs">
          <span class="rounded border border-red-800/50 bg-red-950/20 px-2.5 py-1 text-red-200">{{ counts.new }} new</span>
          <span class="rounded border border-amber-800/50 bg-amber-950/15 px-2.5 py-1 text-amber-200/90">{{ counts.acknowledged }} ack</span>
          <span class="rounded border border-gray-700 bg-gray-900/50 px-2.5 py-1 text-gray-400">{{ counts.resolved }} resolved</span>
        </div>
      </div>
    </div>

    <div
      v-if="notifiedFlash"
      class="rounded-lg border border-teal-800/50 bg-teal-950/20 px-3 py-2 text-sm text-teal-200/90 animate-pulse"
      role="status"
    >
      New alert received
    </div>

    <div v-if="pageError" class="rounded-lg border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-300">
      {{ pageError }}
    </div>

    <div v-if="loading && !alerts.length" class="text-sm text-gray-500">Loading alerts and connecting to stream…</div>
    <div
      v-else-if="!filteredAlerts.length"
      class="rounded-xl border border-dashed border-gray-800 p-8 text-center text-sm text-gray-500"
    >
      No alerts for this view.
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
        />
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import AlertCard from '~/components/AlertCard.vue'
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

const api = useApi()

const cameras = ref<string[]>([])
const camerasLoading = ref(true)
const filterCamera = ref('')
const filterStatus = ref<string>('all')

const alerts = ref<SecurityAlert[]>([])
const loading = ref(true)
const pageError = ref<string | null>(null)
const busyId = ref<number | null>(null)

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

const counts = computed(() => {
  return cameraFilteredAlerts.value.reduce(
    (acc, a) => {
      if (a.status === 'new') acc.new++
      else if (a.status === 'acknowledged') acc.acknowledged++
      else acc.resolved++
      return acc
    },
    { new: 0, acknowledged: 0, resolved: 0 },
  )
})

const cameraFilteredAlerts = computed(() => {
  if (!filterCamera.value) return alerts.value
  return alerts.value.filter((a) => a.camera === filterCamera.value)
})

const filteredAlerts = computed(() => {
  if (filterStatus.value === 'all') return cameraFilteredAlerts.value
  return cameraFilteredAlerts.value.filter((a) => a.status === filterStatus.value)
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

  const stream = new EventSource(api.alertStreamUrl())
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
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Update failed'
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
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Update failed'
    const j = alerts.value.findIndex((a) => a.id === id)
    if (j >= 0) alerts.value[j] = before
  }
  finally {
    busyId.value = null
  }
}
</script>
