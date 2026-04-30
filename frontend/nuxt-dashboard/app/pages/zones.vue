<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-xl font-semibold text-gray-100">Zones</h1>
      <p class="mt-1 text-sm text-gray-500">Draw restricted areas on a {{ FRAME_W }}×{{ FRAME_H }} representative frame (placeholder imagery per camera until live frames are available). Schedules are evaluated in UTC.</p>
      <p v-if="liveSummary" class="mt-2 text-xs text-gray-400 font-mono border-l-2 border-teal-600/50 pl-2">
        Latest live: {{ liveSummary }}
      </p>
      <p v-else-if="selectedCamera && liveLoading" class="mt-2 text-xs text-gray-500">Loading live metadata…</p>
    </header>

    <div class="flex flex-wrap items-end gap-4">
      <div class="flex flex-col gap-1">
        <label class="text-xs text-gray-500 uppercase tracking-wide">Camera</label>
        <select
          v-model="selectedCamera"
          class="min-w-[200px] rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
          :disabled="camerasLoading && !cameras.length"
        >
          <option v-if="!cameras.length" value="">No cameras</option>
          <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <button
        v-if="selectedCamera"
        type="button"
        class="rounded-lg border border-gray-600 bg-gray-800/60 px-4 py-2 text-sm text-gray-200 hover:bg-gray-800"
        :disabled="!selectedCamera || saving || isDrawing"
        @click="startAddZone"
      >
        Add zone
      </button>
      <button
        v-if="isDrawing && !isDraftComplete"
        type="button"
        class="rounded-lg bg-teal-600 hover:bg-teal-500 text-white text-sm font-medium px-4 py-2"
        :disabled="draft.length < 3"
        @click="finishPolygon"
      >
        Finish
      </button>
      <button
        v-if="isDrawing"
        type="button"
        class="rounded-lg border border-red-800/50 px-4 py-2 text-sm text-red-300"
        @click="cancelDraw"
      >
        Cancel
      </button>
    </div>

    <div v-if="pageError" class="rounded-lg border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-300">
      {{ pageError }}
    </div>

    <div class="grid gap-6 lg:grid-cols-[1fr,320px]">
      <div>
        <ZoneDrawer
          :frame-w="FRAME_W"
          :frame-h="FRAME_H"
          :background-candidates="backgroundCandidates"
          :saved-zones="decoratedSavedZones"
          :selected-id="selectedZoneId"
          :draft="isDrawing ? draft : []"
          :is-draft-complete="isDraftComplete"
          :is-drawing="isDrawing"
          @add-point="onAddPoint"
          @finish="finishPolygon"
          @select-zone="selectedZoneId = $event"
        />
      </div>

      <aside class="rounded-xl border border-gray-800 bg-[#12181f] p-4 h-fit max-h-[min(80vh,720px)] overflow-y-auto">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wide mb-3">Zones for this camera</h2>
        <p v-if="!selectedCamera" class="text-sm text-gray-500">Select a camera</p>
        <p v-else-if="zonesLoading" class="text-sm text-gray-500">Loading…</p>
        <p v-else-if="!zonesList.length" class="text-sm text-gray-500">No zones yet</p>
        <ul v-else class="space-y-2">
          <li
            v-for="z in zonesList"
            :key="z.id"
            @click="selectedZoneId = z.id"
            class="rounded-lg border p-3 cursor-pointer transition-colors"
            :class="z.id === selectedZoneId ? 'border-teal-500/50 bg-teal-950/20' : 'border-gray-800 hover:border-gray-700 bg-gray-900/30'"
          >
            <div class="flex items-start justify-between gap-2">
              <p class="text-sm font-medium text-gray-200 truncate">{{ z.name }}</p>
              <span
                class="shrink-0 text-[10px] uppercase tracking-wide rounded px-1.5 py-0.5"
                :class="z.is_active ? 'bg-emerald-950/50 text-emerald-300 border border-emerald-800/40' : 'bg-gray-800 text-gray-500 border border-gray-700'"
              >{{ z.is_active ? 'on' : 'off' }}</span>
            </div>
            <p class="text-xs text-gray-500 mt-1">{{ scheduleSummary(z) }}</p>
            <p v-if="z.loiter_threshold_seconds != null" class="text-xs text-gray-500 mt-0.5">
              Loiter threshold: <span class="text-gray-300 font-mono">{{ z.loiter_threshold_seconds }}s</span>
            </p>
            <p v-if="z.default_severity" class="text-xs text-gray-500 mt-0.5">
              Default severity: <span class="uppercase tracking-wide text-gray-300">{{ z.default_severity }}</span>
            </p>
            <div class="mt-2 flex flex-wrap gap-1.5">
              <button
                type="button"
                class="text-xs text-teal-300 hover:underline"
                @click.stop="openEditModal(z)"
              >Edit</button>
              <button
                type="button"
                class="text-xs text-red-300/80 hover:underline"
                @click.stop="onDeleteZone(z)"
              >Delete</button>
            </div>
          </li>
        </ul>
      </aside>
    </div>

    <Teleport to="body">
      <div
        v-if="showModal"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60"
        @click.self="closeModal"
      >
        <div
          class="w-full max-w-md rounded-xl border border-gray-700 bg-[#12181f] p-5 shadow-xl"
          @keydown.esc="closeModal"
        >
          <h3 class="text-lg font-semibold text-gray-100">
            {{ modalMode === 'create' ? 'New zone' : 'Edit zone' }}
          </h3>
          <div class="mt-4 space-y-4">
            <div>
              <label class="text-xs text-gray-500 uppercase">Name <span class="text-red-400">*</span></label>
              <input
                v-model="form.name"
                type="text"
                class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
                placeholder="e.g. Loading dock"
              >
            </div>
            <div class="flex items-center gap-2">
              <input id="z-active" v-model="form.is_active" type="checkbox" class="rounded border-gray-600">
              <label for="z-active" class="text-sm text-gray-300">Active</label>
            </div>
            <div>
              <p class="text-xs text-gray-500 uppercase mb-2">Schedule</p>
              <div class="flex flex-col gap-2">
                <label class="flex items-center gap-2 text-sm text-gray-300">
                  <input v-model="form.schedule" type="radio" value="always" class="border-gray-600"> Always
                </label>
                <label class="flex items-center gap-2 text-sm text-gray-300">
                  <input v-model="form.schedule" type="radio" value="time" class="border-gray-600"> Time window (UTC)
                </label>
                <div v-if="form.schedule === 'time'" class="pl-5 flex flex-wrap gap-3">
                  <div>
                    <span class="text-xs text-gray-500">From</span>
                    <input
                      v-model="form.activeFrom"
                      type="time"
                      class="mt-0.5 block rounded-lg border border-gray-700 bg-gray-900/80 px-2 py-1.5 text-sm text-gray-200"
                    >
                  </div>
                  <div>
                    <span class="text-xs text-gray-500">Until</span>
                    <input
                      v-model="form.activeUntil"
                      type="time"
                      class="mt-0.5 block rounded-lg border border-gray-700 bg-gray-900/80 px-2 py-1.5 text-sm text-gray-200"
                    >
                  </div>
                </div>
              </div>
            </div>
            <div>
              <label class="text-xs text-gray-500 uppercase">Loiter threshold (seconds)</label>
              <input
                v-model.number="form.loiterThresholdSeconds"
                type="number"
                min="1"
                step="1"
                class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
                placeholder="Optional (e.g. 30)"
              >
            </div>
            <div>
              <label class="text-xs text-gray-500 uppercase">Default severity</label>
              <select
                v-model="form.defaultSeverity"
                class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            <div
              v-if="modalMode === 'edit' && form.zoneId"
              class="pt-1"
            >
              <button
                type="button"
                class="text-sm text-amber-300/90 hover:underline"
                @click="startChangeShape"
              >Change shape</button>
            </div>
          </div>
          <div class="mt-6 flex justify-end gap-2">
            <button
              type="button"
              class="rounded-lg border border-gray-600 px-4 py-2 text-sm text-gray-300"
              :disabled="saving"
              @click="closeModal"
            >Cancel</button>
            <button
              type="button"
              class="rounded-lg bg-teal-600 hover:bg-teal-500 text-white text-sm font-medium px-4 py-2"
              :disabled="saving || !form.name.trim() || !polygonForSave.length"
              @click="submitForm"
            >
              {{ saving ? 'Saving…' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="successToast"
        class="fixed bottom-6 left-1/2 z-[60] -translate-x-1/2 rounded-lg border border-emerald-800/60 bg-emerald-950/90 px-4 py-2.5 text-sm text-emerald-100 shadow-lg"
        role="status"
      >
        {{ successToast }}
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import ZoneDrawer from '~/components/ZoneDrawer.vue'
import type { Point, SecurityZone } from '~/types/security'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const FRAME_W = 640
const FRAME_H = 480

/** Semi-transparent fill (~0.3) + solid stroke: red, blue, green, purple, orange */
const ZONE_PALETTE = [
  { fill: 'rgba(239, 68, 68, 0.3)', stroke: 'rgb(220, 38, 38)' },
  { fill: 'rgba(59, 130, 246, 0.3)', stroke: 'rgb(37, 99, 235)' },
  { fill: 'rgba(34, 197, 94, 0.3)', stroke: 'rgb(22, 163, 74)' },
  { fill: 'rgba(168, 85, 247, 0.3)', stroke: 'rgb(147, 51, 234)' },
  { fill: 'rgba(249, 115, 22, 0.3)', stroke: 'rgb(234, 88, 12)' },
] as const

const api = useApi()

const cameras = ref<string[]>([])
const camerasLoading = ref(true)
const selectedCamera = ref('')

const zonesList = ref<SecurityZone[]>([])
const zonesLoading = ref(false)
const pageError = ref<string | null>(null)
const saving = ref(false)

const selectedZoneId = ref<number | null>(null)

const isDrawing = ref(false)
const isDraftComplete = ref(false)
const draft = ref<Point[]>([])

const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const polygonForSave = ref<Point[]>([])
/** When true, finishing a polygon reopens the edit modal with a new shape (no form reset). */
const redrawingForEdit = ref(false)

const form = reactive({
  name: '',
  is_active: true,
  schedule: 'always' as 'always' | 'time',
  activeFrom: '18:00',
  activeUntil: '06:00',
  loiterThresholdSeconds: null as number | null,
  defaultSeverity: 'medium',
  zoneId: null as number | null,
})

const liveSnapshot = ref<{
  total_detections: number
  latency_ms: number
  recorded_at: string
} | null>(null)
const liveLoading = ref(false)
let livePoll: ReturnType<typeof setInterval> | null = null

const successToast = ref<string | null>(null)
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(msg: string) {
  successToast.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    successToast.value = null
    toastTimer = null
  }, 3200)
}

function sanitizeCameraFilePart(name: string) {
  const s = name.replace(/[^a-zA-Z0-9_-]/g, '_').slice(0, 64)
  return s || 'camera'
}

function cameraBackgroundCandidates(camera: string | undefined): string[] {
  if (!camera) return ['/camera-backgrounds/cam1.jpg']
  const safe = sanitizeCameraFilePart(camera)
  let h = 0
  for (let i = 0; i < camera.length; i++)
    h = (h + camera.charCodeAt(i) * (i + 1)) % 100000
  const variant = (h % 2) + 1
  return [
    `/camera-backgrounds/${safe}.jpg`,
    `/camera-backgrounds/cam${variant}.jpg`,
    '/camera-backgrounds/cam1.jpg',
  ]
}

const backgroundCandidates = computed(() =>
  cameraBackgroundCandidates(selectedCamera.value || undefined))

const liveSummary = computed(() => {
  const s = liveSnapshot.value
  if (!s) return ''
  const t = new Date(s.recorded_at)
  const time = Number.isNaN(t.getTime()) ? s.recorded_at : t.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  return `${s.total_detections} detections · ${s.latency_ms.toFixed(0)} ms latency · ${time}`
})

async function refreshLive() {
  if (!selectedCamera.value) {
    liveSnapshot.value = null
    return
  }
  liveLoading.value = true
  try {
    const res = await api.getLive(selectedCamera.value)
    const row = res.detections?.[0]
    liveSnapshot.value = row
      ? {
          total_detections: row.total_detections,
          latency_ms: row.latency_ms,
          recorded_at: row.recorded_at,
        }
      : null
  }
  catch {
    liveSnapshot.value = null
  }
  finally {
    liveLoading.value = false
  }
}

onMounted(async () => {
  camerasLoading.value = true
  try {
    const res = await api.getCameras()
    cameras.value = res.cameras
    if (res.cameras.length) selectedCamera.value = res.cameras[0]!
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Could not load cameras'
  }
  finally {
    camerasLoading.value = false
  }
  void refreshLive()
  livePoll = setInterval(() => { void refreshLive() }, 15_000)
})

onUnmounted(() => {
  if (livePoll) clearInterval(livePoll)
  if (toastTimer) clearTimeout(toastTimer)
})

watch(selectedCamera, async (cam) => {
  if (!cam) {
    zonesList.value = []
    liveSnapshot.value = null
    return
  }
  await loadZones(cam)
  void refreshLive()
})

function zoneColors(i: number, selected: boolean) {
  const p = ZONE_PALETTE[i % ZONE_PALETTE.length]!
  if (selected) {
    return {
      fillColor: p.fill.replace('0.3', '0.45'),
      strokeColor: p.stroke,
    }
  }
  return { fillColor: p.fill, strokeColor: p.stroke }
}

const decoratedSavedZones = computed(() =>
  zonesList.value.map((z, i) => {
    const { fillColor, strokeColor } = zoneColors(i, z.id === selectedZoneId.value)
    return { id: z.id, name: z.name, polygon: z.polygon, fillColor, strokeColor }
  }))

function isoToTimeInput(s: string | null | undefined): string {
  if (!s) return '00:00'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return '00:00'
  return `${String(d.getUTCHours()).padStart(2, '0')}:${String(d.getUTCMinutes()).padStart(2, '0')}`
}

function timeInputToReferenceIso(t: string): string {
  const m = t.match(/^(\d{1,2}):(\d{2})$/)
  if (!m) return '2000-01-01T00:00:00.000Z'
  const h = Math.min(23, Math.max(0, parseInt(m[1]!, 10)))
  const min = Math.min(59, Math.max(0, parseInt(m[2]!, 10)))
  return `2000-01-01T${String(h).padStart(2, '0')}:${String(min).padStart(2, '0')}:00.000Z`
}

function scheduleSummary(z: SecurityZone) {
  if (z.active_from == null && z.active_until == null) return 'Always on'
  return `${isoToTimeInput(z.active_from)} – ${isoToTimeInput(z.active_until)} UTC`
}

function resetForm() {
  form.name = ''
  form.is_active = true
  form.schedule = 'always'
  form.activeFrom = '18:00'
  form.activeUntil = '06:00'
  form.loiterThresholdSeconds = null
  form.defaultSeverity = 'medium'
  form.zoneId = null
}

async function loadZones(camera: string) {
  zonesLoading.value = true
  pageError.value = null
  try {
    const { data } = await api.listZones(camera)
    zonesList.value = data
    selectedZoneId.value = data[0]?.id ?? null
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to load zones'
    zonesList.value = []
  }
  finally {
    zonesLoading.value = false
  }
}

function startAddZone() {
  redrawingForEdit.value = false
  cancelDraw()
  isDrawing.value = true
  isDraftComplete.value = false
  draft.value = []
  selectedZoneId.value = null
}

function onAddPoint(p: Point) {
  if (isDraftComplete.value) return
  const x = Math.max(0, Math.min(FRAME_W, p.x))
  const y = Math.max(0, Math.min(FRAME_H, p.y))
  draft.value = [...draft.value, { x, y }]
}

function finishPolygon() {
  if (draft.value.length < 3) return
  isDraftComplete.value = true
  isDrawing.value = false
  polygonForSave.value = draft.value.map((p) => ({ ...p }))
  draft.value = []
  isDraftComplete.value = false
  if (redrawingForEdit.value) {
    redrawingForEdit.value = false
    modalMode.value = 'edit'
    showModal.value = true
  }
  else {
    modalMode.value = 'create'
    resetForm()
    showModal.value = true
  }
}

function cancelDraw() {
  isDrawing.value = false
  isDraftComplete.value = false
  draft.value = []
}

function openEditModal(z: SecurityZone) {
  redrawingForEdit.value = false
  cancelDraw()
  selectedZoneId.value = z.id
  modalMode.value = 'edit'
  polygonForSave.value = z.polygon.map((p) => ({ ...p }))
  form.name = z.name
  form.is_active = z.is_active
  if (z.active_from == null && z.active_until == null) {
    form.schedule = 'always'
  }
  else {
    form.schedule = 'time'
    form.activeFrom = isoToTimeInput(z.active_from)
    form.activeUntil = isoToTimeInput(z.active_until)
  }
  form.loiterThresholdSeconds = z.loiter_threshold_seconds ?? null
  form.defaultSeverity = z.default_severity ?? 'medium'
  form.zoneId = z.id
  showModal.value = true
}

function startChangeShape() {
  if (!form.zoneId) return
  showModal.value = false
  redrawingForEdit.value = true
  isDrawing.value = true
  isDraftComplete.value = false
  draft.value = []
}

function closeModal() {
  if (saving.value) return
  showModal.value = false
  polygonForSave.value = []
  isDraftComplete.value = false
}

function buildZoneBody(): {
  name: string
  camera: string
  polygon: Point[]
  is_active: boolean
  active_from: string | null
  active_until: string | null
  loiter_threshold_seconds: number | null
  default_severity: string
} {
  const fromTime = timeInputToReferenceIso(form.activeFrom)
  const untilTime = timeInputToReferenceIso(form.activeUntil)
  const rawLoiter = form.loiterThresholdSeconds
  const loiter =
    rawLoiter == null || Number.isNaN(rawLoiter)
      ? null
      : Math.max(1, Math.floor(rawLoiter))
  return {
    name: form.name.trim(),
    camera: selectedCamera.value,
    polygon: polygonForSave.value,
    is_active: form.is_active,
    active_from: form.schedule === 'time' ? fromTime : null,
    active_until: form.schedule === 'time' ? untilTime : null,
    loiter_threshold_seconds: loiter,
    default_severity: form.defaultSeverity,
  }
}

async function submitForm() {
  if (!selectedCamera.value || !polygonForSave.value.length) return
  saving.value = true
  pageError.value = null
  try {
    const body = buildZoneBody()
    if (modalMode.value === 'create') {
      await api.createZone(body)
    }
    else if (form.zoneId != null) {
      await api.updateZone(form.zoneId, body)
    }
    showModal.value = false
    polygonForSave.value = []
    await loadZones(selectedCamera.value)
    showToast(modalMode.value === 'create' ? 'Zone created' : 'Zone updated')
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Save failed'
  }
  finally {
    saving.value = false
  }
}

async function onDeleteZone(z: SecurityZone) {
  if (!confirm(`Delete zone "${z.name}"?`)) return
  pageError.value = null
  try {
    await api.deleteZone(z.id)
    if (selectedZoneId.value === z.id) selectedZoneId.value = null
    await loadZones(selectedCamera.value)
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Delete failed'
  }
}
</script>
