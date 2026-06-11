<template>
  <div class="relative overflow-hidden rounded-lg border border-white/[0.06] bg-gray-950/60">
    <img
      v-if="src && !imgError"
      :src="src"
      alt=""
      class="absolute inset-0 h-full w-full object-cover opacity-80"
      @error="imgError = true"
    >
    <svg
      :viewBox="`0 0 ${frameWidth} ${frameHeight}`"
      class="relative w-full h-auto"
      preserveAspectRatio="xMidYMid meet"
      role="img"
      :aria-label="`${label} location overlay`"
    >
      <rect
        v-if="!showImage"
        :width="frameWidth"
        :height="frameHeight"
        fill="#0b0f14"
      />
      <rect
        v-else
        :width="frameWidth"
        :height="frameHeight"
        fill="#000"
        fill-opacity="0.15"
      />
      <g v-if="!showImage">
        <line :x1="frameWidth / 3" y1="0" :x2="frameWidth / 3" :y2="frameHeight" stroke="#1f2937" stroke-width="1" />
        <line :x1="(frameWidth / 3) * 2" y1="0" :x2="(frameWidth / 3) * 2" :y2="frameHeight" stroke="#1f2937" stroke-width="1" />
        <line x1="0" :y1="frameHeight / 3" :x2="frameWidth" :y2="frameHeight / 3" stroke="#1f2937" stroke-width="1" />
        <line x1="0" :y1="(frameHeight / 3) * 2" :x2="frameWidth" :y2="(frameHeight / 3) * 2" stroke="#1f2937" stroke-width="1" />
      </g>

      <template v-if="hasBox">
        <rect
          :x="box.x"
          :y="box.y"
          :width="box.w"
          :height="box.h"
          fill="none"
          :stroke="color"
          :stroke-width="strokeWidth"
        />
        <rect
          :x="box.x"
          :y="box.y"
          :width="box.w"
          :height="box.h"
          :fill="color"
          fill-opacity="0.12"
        />
        <g>
          <rect
            :x="box.x"
            :y="Math.max(0, box.y - labelHeight)"
            :width="labelWidth"
            :height="labelHeight"
            :fill="color"
          />
          <text
            :x="box.x + labelHeight * 0.35"
            :y="Math.max(0, box.y - labelHeight) + labelHeight * 0.72"
            :font-size="labelHeight * 0.6"
            font-family="monospace"
            font-weight="700"
            fill="#0b0f14"
          >{{ label }}</text>
        </g>
      </template>

      <text
        v-else
        :x="frameWidth / 2"
        :y="frameHeight / 2"
        text-anchor="middle"
        :font-size="frameHeight * 0.06"
        fill="#4b5563"
        font-family="monospace"
      >no box</text>
    </svg>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  boundBox: [number, number, number, number] | number[] | null | undefined
  color?: string
  label?: string
  src?: string | null
  frameWidth?: number
  frameHeight?: number
}>(), {
  color: '#fb923c',
  label: 'BOX',
  src: null,
  frameWidth: 1920,
  frameHeight: 1080,
})

const imgError = ref(false)
const showImage = computed(() => !!props.src && !imgError.value)

watch(() => props.src, () => { imgError.value = false })

const strokeWidth = computed(() => Math.max(2, props.frameWidth * 0.004))
const labelHeight = computed(() => Math.max(18, props.frameHeight * 0.05))
const labelWidth = computed(() => labelHeight.value * 0.62 * (props.label.length + 1))

const hasBox = computed(() => {
  const b = props.boundBox
  if (!b || b.length < 4) return false
  const [x1, y1, x2, y2] = b
  return Number.isFinite(x1) && Number.isFinite(y1) && Number.isFinite(x2) && Number.isFinite(y2)
    && (x2 - x1) > 0 && (y2 - y1) > 0
})

const box = computed(() => {
  const b = props.boundBox ?? [0, 0, 0, 0]
  const x1 = Math.min(b[0]!, b[2]!)
  const y1 = Math.min(b[1]!, b[3]!)
  const x2 = Math.max(b[0]!, b[2]!)
  const y2 = Math.max(b[1]!, b[3]!)
  const cx1 = Math.max(0, Math.min(x1, props.frameWidth))
  const cy1 = Math.max(0, Math.min(y1, props.frameHeight))
  const cx2 = Math.max(0, Math.min(x2, props.frameWidth))
  const cy2 = Math.max(0, Math.min(y2, props.frameHeight))
  return { x: cx1, y: cy1, w: cx2 - cx1, h: cy2 - cy1 }
})
</script>
