<template>
  <div class="relative overflow-hidden rounded-lg border border-white/[0.06] bg-gray-950/60">
    <canvas
      ref="canvasRef"
      :width="PIPELINE_FRAME_WIDTH"
      :height="PIPELINE_FRAME_HEIGHT"
      class="block w-full h-auto aspect-[4/3]"
      :aria-label="`${label} location overlay`"
    />
  </div>
</template>

<script setup lang="ts">
import {
  normalizeBoundBox,
  PIPELINE_FRAME_HEIGHT,
  PIPELINE_FRAME_WIDTH,
} from '~/composables/useBoundBox'

const props = withDefaults(defineProps<{
  boundBox: [number, number, number, number] | number[] | null | undefined
  color?: string
  label?: string
  src?: string | null
}>(), {
  color: '#fb923c',
  label: 'BOX',
  src: null,
})

const canvasRef = ref<HTMLCanvasElement | null>(null)
const bgImage = shallowRef<HTMLImageElement | null>(null)
const imgError = ref(false)

/** Bboxes from the API are always in pipeline pixel space (640×480). */
const box = computed(() => normalizeBoundBox(props.boundBox))

function loadImage(url: string | null | undefined) {
  bgImage.value = null
  imgError.value = false
  if (!url || !import.meta.client) {
    nextTick(() => draw())
    return
  }
  const im = new Image()
  im.onload = () => {
    bgImage.value = im
    nextTick(() => draw())
  }
  im.onerror = () => {
    imgError.value = true
    nextTick(() => draw())
  }
  im.src = url
}

watch(() => props.src, (url) => loadImage(url), { immediate: true })
watch(
  () => [props.boundBox, props.color, props.label, bgImage.value, imgError.value],
  () => nextTick(() => draw()),
  { deep: true },
)

function drawGrid(ctx: CanvasRenderingContext2D) {
  const w = PIPELINE_FRAME_WIDTH
  const h = PIPELINE_FRAME_HEIGHT
  const g = ctx.createLinearGradient(0, 0, w, h)
  g.addColorStop(0, '#1e293b')
  g.addColorStop(0.5, '#0f172a')
  g.addColorStop(1, '#020617')
  ctx.fillStyle = g
  ctx.fillRect(0, 0, w, h)
  ctx.strokeStyle = 'rgba(71, 85, 105, 0.35)'
  ctx.lineWidth = 1
  for (let x = 0; x < w; x += w / 3) {
    ctx.beginPath()
    ctx.moveTo(x, 0)
    ctx.lineTo(x, h)
    ctx.stroke()
  }
  for (let y = 0; y < h; y += h / 3) {
    ctx.beginPath()
    ctx.moveTo(0, y)
    ctx.lineTo(w, y)
    ctx.stroke()
  }
}

function draw() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const w = PIPELINE_FRAME_WIDTH
  const h = PIPELINE_FRAME_HEIGHT
  ctx.clearRect(0, 0, w, h)

  const im = bgImage.value
  if (im && im.complete && im.naturalWidth > 0 && !imgError.value) {
    // Scale frame into pipeline coordinates — same space as bound_box from backend.
    ctx.drawImage(im, 0, 0, w, h)
    ctx.fillStyle = 'rgba(0, 0, 0, 0.08)'
    ctx.fillRect(0, 0, w, h)
  }
  else {
    drawGrid(ctx)
  }

  const b = box.value
  if (!b) {
    if (!im || imgError.value) {
      ctx.fillStyle = '#4b5563'
      ctx.font = `${Math.max(12, h * 0.06)}px monospace`
      ctx.textAlign = 'center'
      ctx.textBaseline = 'middle'
      ctx.fillText('no box', w / 2, h / 2)
    }
    return
  }

  const [x1, y1, x2, y2] = b
  const bw = x2 - x1
  const bh = y2 - y1
  const stroke = Math.max(2, w * 0.004)
  const labelH = Math.max(14, h * 0.05)
  const labelW = labelH * 0.62 * (props.label.length + 1)
  const labelY = Math.max(0, y1 - labelH)

  ctx.globalAlpha = 0.14
  ctx.fillStyle = props.color
  ctx.fillRect(x1, y1, bw, bh)
  ctx.globalAlpha = 1

  ctx.strokeStyle = props.color
  ctx.lineWidth = stroke
  ctx.strokeRect(x1, y1, bw, bh)

  ctx.fillStyle = props.color
  ctx.fillRect(x1, labelY, labelW, labelH)
  ctx.fillStyle = '#0b0f14'
  ctx.font = `700 ${labelH * 0.6}px monospace`
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  ctx.fillText(props.label, x1 + labelH * 0.35, labelY + labelH * 0.55)
}

onMounted(() => nextTick(() => draw()))
</script>
