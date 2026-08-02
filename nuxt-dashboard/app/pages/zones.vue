<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Geofencing"
      title="Zones"
      subtitle="Draw restricted areas on a representative frame. Schedules are evaluated in UTC."
    >
      <template #actions>
        <span v-if="liveSummary" class="app-chip font-mono text-xs text-gray-400">
          Latest: {{ liveSummary }}
        </span>
        <span v-else-if="selectedCamera && liveLoading" class="app-chip text-xs text-gray-500">
          Loading live metadata…
        </span>
        <button
          type="button"
          class="btn-primary"
          :disabled="!selectedCamera || saving || isDrawing"
          @click="startAddZone"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
          Add zone
        </button>
      </template>
    </PageHeader>

    <FilterToolbar>
      <div class="flex flex-col gap-1">
        <label class="app-label" for="zones-camera">Camera</label>
        <select
          id="zones-camera"
          v-model="selectedCamera"
          class="app-input min-w-[200px]"
          :disabled="camerasLoading && !cameras.length"
        >
          <option v-if="!cameras.length" value="">No cameras</option>
          <option v-for="c in cameras" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <template #actions>
        <button
          type="button"
          class="btn-primary"
          :disabled="!selectedCamera || saving || isDrawing"
          @click="startAddZone"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
          Add zone
        </button>
        <button
          v-if="isDrawing && !isDraftComplete"
          type="button"
          class="btn-primary"
          :disabled="draft.length < 3"
          @click="finishPolygon"
        >
          Finish
        </button>
        <button
          v-if="isDrawing"
          type="button"
          class="btn-ghost text-red-300 hover:text-red-200"
          @click="cancelDraw"
        >
          Cancel
        </button>
      </template>
    </FilterToolbar>

    <div v-if="pageError" class="app-banner-error">{{ pageError }}</div>

    <!-- Drawing mode instruction banner -->
    <div
      v-if="isDrawing"
      class="rounded-xl border border-teal-500/40 bg-teal-950/30 px-5 py-4"
    >
      <div class="flex items-start gap-3">
        <svg class="mt-0.5 h-5 w-5 shrink-0 text-teal-400" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9.53 16.122a3 3 0 00-5.78 1.128 2.25 2.25 0 01-2.4 2.245 4.5 4.5 0 008.4-2.245c0-.399-.078-.78-.22-1.128zm0 0a15.998 15.998 0 003.388-1.62m-5.043-.025a15.994 15.994 0 011.622-3.395m3.42 3.42a15.995 15.995 0 004.764-4.648l3.876-5.814a1.151 1.151 0 00-1.597-1.597L14.146 6.32a15.996 15.996 0 00-4.649 4.763m3.42 3.42a6.776 6.776 0 00-3.42-3.42" />
        </svg>
        <div>
          <p class="text-sm font-medium text-teal-200">Drawing mode active</p>
          <p class="mt-1 text-sm text-teal-300/80">
            <strong>Click on the canvas</strong> to place polygon points (minimum 3).
            <strong>Double-click</strong> or press the <strong>Finish</strong> button to close the shape.
          </p>
          <p class="mt-1.5 text-xs text-teal-100">
            This zone will be assigned to camera:
            <span class="rounded bg-teal-500/20 px-1.5 py-0.5 font-mono font-semibold text-teal-200">{{ selectedCamera || '—' }}</span>
          </p>
          <p v-if="draft.length" class="mt-1.5 text-xs font-mono text-teal-400">
            {{ draft.length }} point{{ draft.length === 1 ? '' : 's' }} placed{{ draft.length < 3 ? ` — need ${3 - draft.length} more` : ' — ready to finish' }}
          </p>
        </div>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-[1fr,320px]">
      <div>
        <ZoneDrawer
          :frame-w="FRAME_W"
          :frame-h="FRAME_H"
          :camera="selectedCamera"
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

      <aside class="app-card p-4 h-fit max-h-[min(80vh,720px)] overflow-y-auto">
        <h2 class="app-section-title mb-3">Zones for this camera</h2>
        <EmptyState
          v-if="!selectedCamera"
          icon="zone"
          title="Select a camera"
          message="Choose a camera to view and manage its restricted zones."
          class="py-6"
        />
        <div v-else-if="zonesLoading" class="space-y-2 py-2">
          <Skeleton v-for="i in 3" :key="i" width="100%" height="4rem" rounded="rounded-xl" />
        </div>
        <EmptyState
          v-else-if="!zonesList.length"
          icon="zone"
          title="No zones yet"
          message="Draw your first restricted area on the frame to start generating zone alerts."
          class="py-6"
        >
          <template #action>
            <button type="button" class="btn-primary" :disabled="isDrawing" @click="startAddZone">
              Add zone
            </button>
          </template>
        </EmptyState>
        <ul v-else class="space-y-2">
          <li
            v-for="z in zonesList"
            :key="z.id"
            @click="selectedZoneId = z.id"
            class="rounded-xl border p-3 cursor-pointer transition-all duration-200"
            :class="z.id === selectedZoneId ? 'border-teal-500/50 bg-teal-500/[0.08] ring-1 ring-teal-500/20' : 'border-white/[0.06] hover:border-white/10 bg-white/[0.02] hover:bg-white/[0.04]'"
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

    <AppModal
      :open="showModal"
      :title="modalMode === 'create' ? 'New zone' : 'Edit zone'"
      :subtitle="`For camera “${selectedCamera}” · configure name, schedule, and alert thresholds.`"
      @close="closeModal"
    >
      <div class="space-y-4">
            <div class="rounded-lg border border-teal-500/30 bg-teal-950/20 px-3 py-2 text-sm text-teal-200">
              <span class="text-teal-400/80">Assigned camera:</span>
              <span class="font-mono font-semibold">{{ selectedCamera || '—' }}</span>
            </div>
            <div>
              <label class="app-label mb-1.5 block">Name <span class="text-red-400">*</span></label>
              <input
                v-model="form.name"
                type="text"
                class="app-input w-full"
                placeholder="e.g. Loading dock"
              >
            </div>
            <label class="inline-flex items-center gap-2 text-sm text-gray-300">
              <input id="z-active" v-model="form.is_active" type="checkbox" class="app-checkbox">
              Active
            </label>
            <div>
              <p class="app-label mb-2">Schedule</p>
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
                      class="app-input mt-0.5 block px-2 py-1.5"
                    >
                  </div>
                  <div>
                    <span class="text-xs text-gray-500">Until</span>
                    <input
                      v-model="form.activeUntil"
                      type="time"
                      class="app-input mt-0.5 block px-2 py-1.5"
                    >
                  </div>
                </div>
              </div>
            </div>
            <div>
              <label class="app-label mb-1.5 block">Loiter threshold (seconds)</label>
              <input
                v-model.number="form.loiterThresholdSeconds"
                type="number"
                min="1"
                step="1"
                class="app-input w-full"
                placeholder="Optional (e.g. 30)"
              >
            </div>
            <div>
              <label class="app-label mb-1.5 block">Default severity</label>
              <select
                v-model="form.defaultSeverity"
                class="app-input w-full"
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
      <template #footer>
        <button type="button" class="btn-ghost" :disabled="saving" @click="closeModal">Cancel</button>
        <button
          type="button"
          class="btn-primary disabled:opacity-50"
          :disabled="saving || !form.name.trim() || !polygonForSave.length"
          @click="submitForm"
        >
          {{ saving ? 'Saving…' : 'Save' }}
        </button>
      </template>
    </AppModal>

  </div>
</template>

<script setup lang="ts">
import PageHeader from '~/components/ui/PageHeader.vue'
import AppModal from '~/components/ui/AppModal.vue'
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
const toast = useToast()
const confirm = useConfirm()

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
    cameras.value = res.cameras ?? []
    if (cameras.value.length) selectedCamera.value = cameras.value[0]!
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
    const res = await api.listZones(camera)
    const data = res.data ?? []
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
    const created = modalMode.value === 'create'
    await loadZones(selectedCamera.value)
    toast.success(created ? 'Zone created' : 'Zone updated', { description: body.name })
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Save failed'
    toast.error('Failed to save zone')
  }
  finally {
    saving.value = false
  }
}

async function onDeleteZone(z: SecurityZone) {
  const ok = await confirm.ask({
    title: 'Delete zone?',
    message: `"${z.name}" and its detection rules will be permanently removed.`,
    confirmLabel: 'Delete',
    danger: true,
  })
  if (!ok) return
  pageError.value = null
  try {
    await api.deleteZone(z.id)
    if (selectedZoneId.value === z.id) selectedZoneId.value = null
    await loadZones(selectedCamera.value)
    toast.success('Zone deleted', { description: z.name })
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Delete failed'
    toast.error('Failed to delete zone')
  }
}
</script>
