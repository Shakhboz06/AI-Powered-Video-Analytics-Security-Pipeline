<template>
  <div class="space-y-2">
    <div
      ref="wrapRef"
      class="relative w-full max-w-[640px] rounded-lg border border-gray-700 bg-gray-900/50 overflow-hidden select-none touch-none"
      :class="isDrawing ? 'ring-2 ring-teal-500/40' : ''"
    >
      <canvas
        ref="canvasRef"
        :width="frameW"
        :height="frameH"
        class="block w-full h-auto cursor-crosshair"
        :class="cursorClass"
        @click="onCanvasClick"
        @dblclick.prevent="onDblClick"
        @mousemove="onMove"
        @mouseleave="onLeave"
      />
      <div
        v-if="isDrawing && cursorTip.visible"
        class="pointer-events-none absolute z-20 rounded border border-gray-600 bg-gray-950/95 px-2 py-1 text-[11px] font-mono text-teal-200 shadow-lg"
        :style="{ left: `${cursorTip.x}px`, top: `${cursorTip.y}px`, transform: 'translate(12px, 12px)' }"
      >
        x: {{ cursorTip.px }}, y: {{ cursorTip.py }}
      </div>
    </div>
    <p class="text-xs text-gray-500">
      Click to add points. Minimum 3 points required. Double-click or press <span class="text-gray-300">Finish</span> to close the polygon.
      <span v-if="!isDrawing" class="block mt-1 text-gray-500">Hover a zone to highlight it; click to select. Use <span class="text-gray-300">Add zone</span> to draw a new area.</span>
    </p>
  </div>
</template>

<script setup lang="ts">
import type { Point } from '~/types/security'

const props = defineProps<{
  frameW: number
  frameH: number
  /** Try in order until one loads (e.g. cam-specific .jpg, then fallback .jpg). */
  backgroundCandidates: string[]
  savedZones: Array<{
    id: number
    name: string
    polygon: Point[]
    fillColor: string
    strokeColor: string
  }>
  selectedId: number | null
  draft: Point[]
  isDraftComplete: boolean
  isDrawing: boolean
}>()

const emit = defineEmits<{
  'add-point': [point: Point]
  finish: []
  'select-zone': [id: number]
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)
const wrapRef = ref<HTMLElement | null>(null)

const bgImage = shallowRef<HTMLImageElement | null>(null)

const hoveredZoneId = ref<number | null>(null)

const cursorTip = reactive({
  visible: false,
  x: 0,
  y: 0,
  px: 0,
  py: 0,
})

const cursorClass = computed(() => {
  if (props.isDrawing) return 'cursor-crosshair'
  if (hoveredZoneId.value != null) return 'cursor-pointer'
  return 'cursor-default'
})

function getCanvasPoint(e: MouseEvent): Point | null {
  const c = canvasRef.value
  if (!c) return null
  const r = c.getBoundingClientRect()
  const x = ((e.clientX - r.left) / r.width) * props.frameW
  const y = ((e.clientY - r.top) / r.height) * props.frameH
  return { x, y }
}

function pointInPolygon(x: number, y: number, poly: Point[]): boolean {
  if (poly.length < 3) return false
  let inside = false
  for (let i = 0, j = poly.length - 1; i < poly.length; j = i++) {
    const pi = poly[i]
    const pj = poly[j]
    if (!pi || !pj) continue
    const xi = pi.x
    const yi = pi.y
    const xj = pj.x
    const yj = pj.y
    if (Math.abs(yj - yi) < 1e-9) continue
    const intersect = (yi > y) !== (yj > y) && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi
    if (intersect) inside = !inside
  }
  return inside
}

function zoneHitTest(p: Point): number | null {
  const zones = props.savedZones
  for (let i = zones.length - 1; i >= 0; i--) {
    const z = zones[i]
    if (!z || z.polygon.length < 3) continue
    if (pointInPolygon(p.x, p.y, z.polygon)) return z.id
  }
  return null
}

function loadBackground() {
  if (import.meta.server) {
    nextTick(() => draw())
    return
  }
  bgImage.value = null
  const list = props.backgroundCandidates
  if (!list.length) {
    nextTick(() => draw())
    return
  }
  const tryLoad = (idx: number) => {
    if (idx >= list.length) {
      nextTick(() => draw())
      return
    }
    const url = list[idx]
    if (!url) {
      tryLoad(idx + 1)
      return
    }
    const im = new Image()
    im.onload = () => {
      bgImage.value = im
      nextTick(() => draw())
    }
    im.onerror = () => {
      tryLoad(idx + 1)
    }
    im.src = url
  }
  tryLoad(0)
}

watch(
  () => [...props.backgroundCandidates],
  () => {
    loadBackground()
  },
  { immediate: true },
)

watch(
  () => props.isDrawing,
  (drawing) => {
    if (drawing) hoveredZoneId.value = null
  },
)

function onMove(e: MouseEvent) {
  const p = getCanvasPoint(e)
  if (!p || !wrapRef.value) return

  if (props.isDrawing) {
    hoveredZoneId.value = null
    cursorTip.visible = true
    cursorTip.px = Math.round(Math.max(0, Math.min(props.frameW, p.x)))
    cursorTip.py = Math.round(Math.max(0, Math.min(props.frameH, p.y)))
    const wrap = wrapRef.value.getBoundingClientRect()
    cursorTip.x = e.clientX - wrap.left
    cursorTip.y = e.clientY - wrap.top
  }
  else {
    cursorTip.visible = false
    hoveredZoneId.value = zoneHitTest(p)
  }
}

function onLeave() {
  cursorTip.visible = false
  hoveredZoneId.value = null
}

function onCanvasClick(e: MouseEvent) {
  if (props.isDrawing) {
    if (e.detail > 1) return
    const p = getCanvasPoint(e)
    if (!p) return
    emit('add-point', p)
    return
  }
  const p = getCanvasPoint(e)
  if (!p) return
  const id = zoneHitTest(p)
  if (id != null) emit('select-zone', id)
}

function onDblClick() {
  if (!props.isDrawing) return
  if (props.draft.length < 3) return
  emit('finish')
}

function drawBackground(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const im = bgImage.value
  if (im && im.complete && im.naturalWidth > 0) {
    const iw = im.naturalWidth
    const ih = im.naturalHeight
    const scale = Math.max(w / iw, h / ih)
    const dw = iw * scale
    const dh = ih * scale
    const ox = (w - dw) / 2
    const oy = (h - dh) / 2
    ctx.drawImage(im, ox, oy, dw, dh)
    return
  }
  const g = ctx.createLinearGradient(0, 0, w, h)
  g.addColorStop(0, '#1e293b')
  g.addColorStop(0.5, '#0f172a')
  g.addColorStop(1, '#020617')
  ctx.fillStyle = g
  ctx.fillRect(0, 0, w, h)
  ctx.strokeStyle = 'rgba(71, 85, 105, 0.35)'
  ctx.lineWidth = 1
  for (let x = 0; x < w; x += 32) {
    ctx.beginPath()
    ctx.moveTo(x, 0)
    ctx.lineTo(x, h)
    ctx.stroke()
  }
  for (let y = 0; y < h; y += 32) {
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(w, y)
    ctx.stroke()
  }
}

function draw() {
  const c = canvasRef.value
  if (!c) return
  const ctx = c.getContext('2d')
  if (!ctx) return
  const { frameW: w, frameH: h } = props
  ctx.clearRect(0, 0, w, h)
  drawBackground(ctx, w, h)

  props.savedZones.forEach((z) => {
    if (z.polygon.length < 3) return
    const selected = z.id === props.selectedId
    const hovered = z.id === hoveredZoneId.value
    const fill = z.polygon
    const p0 = fill[0]
    if (!p0) return
    ctx.beginPath()
    ctx.moveTo(p0.x, p0.y)
    for (let j = 1; j < fill.length; j++) {
      const pj = fill[j]
      if (pj) ctx.lineTo(pj.x, pj.y)
    }
    ctx.closePath()
    ctx.fillStyle = z.fillColor
    ctx.fill()
    ctx.strokeStyle = z.strokeColor
    ctx.lineWidth = selected ? 3 : hovered ? 2.5 : 2
    if (hovered && !selected) {
      ctx.shadowColor = 'rgba(255,255,255,0.25)'
      ctx.shadowBlur = 8
    }
    ctx.stroke()
    ctx.shadowBlur = 0
    if (z.name) {
      ctx.fillStyle = selected || hovered ? 'rgba(255,255,255,0.95)' : 'rgba(230,230,230,0.85)'
      ctx.font = 'bold 12px sans-serif'
      ctx.strokeStyle = 'rgba(0,0,0,0.55)'
      ctx.lineWidth = 3
      ctx.strokeText(z.name, p0.x + 4, p0.y + 14)
      ctx.fillText(z.name, p0.x + 4, p0.y + 14)
    }
  })

  if (props.draft.length) {
    const d = props.draft
    const a = d[0]
    if (a) {
      ctx.beginPath()
      ctx.moveTo(a.x, a.y)
      for (let i = 1; i < d.length; i++) {
        const p = d[i]
        if (p) ctx.lineTo(p.x, p.y)
      }
      if (props.isDraftComplete && d.length >= 3) {
        ctx.closePath()
        ctx.fillStyle = 'rgba(45, 212, 191, 0.2)'
        ctx.fill()
      }
      ctx.strokeStyle = 'rgb(45, 212, 191)'
      ctx.lineWidth = 2
      ctx.setLineDash(props.isDraftComplete ? [] : [6, 4])
      ctx.stroke()
      ctx.setLineDash([])

      d.forEach((p) => {
        ctx.beginPath()
        ctx.arc(p.x, p.y, 4, 0, Math.PI * 2)
        ctx.fillStyle = 'rgb(45, 212, 191)'
        ctx.fill()
        ctx.strokeStyle = 'rgba(0,0,0,0.5)'
        ctx.lineWidth = 1
        ctx.stroke()
      })
    }
  }
}

watch(
  () => ({
    saved: props.savedZones,
    selectedId: props.selectedId,
    draft: props.draft,
    isDraftComplete: props.isDraftComplete,
    isDrawing: props.isDrawing,
    hover: hoveredZoneId.value,
    bg: bgImage.value,
  }),
  () => nextTick(() => draw()),
  { deep: true },
)

onMounted(() => {
  nextTick(() => draw())
})

defineExpose({ redraw: draw })
</script>
