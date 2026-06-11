<template>
  <svg
    :viewBox="`0 0 ${width} ${height}`"
    class="w-full"
    :style="{ height: `${height}px` }"
    preserveAspectRatio="none"
    aria-hidden="true"
  >
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.28" />
        <stop offset="100%" :stop-color="color" stop-opacity="0" />
      </linearGradient>
    </defs>
    <template v-if="points.length > 1">
      <path :d="areaPath" :fill="`url(#${gradId})`" />
      <path :d="linePath" fill="none" :stroke="color" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
    </template>
  </svg>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  data: number[]
  color?: string
  width?: number
  height?: number
}>(), {
  color: '#2dd4bf',
  width: 120,
  height: 36,
})

const gradId = `spark-${Math.random().toString(36).slice(2, 9)}`

const points = computed(() => {
  const d = props.data?.filter((n) => Number.isFinite(n)) ?? []
  if (d.length < 2) return [] as { x: number; y: number }[]
  const min = Math.min(...d)
  const max = Math.max(...d)
  const range = max - min || 1
  const pad = 2
  const w = props.width - pad * 2
  const h = props.height - pad * 2
  return d.map((v, i) => ({
    x: pad + (i / (d.length - 1)) * w,
    y: pad + (1 - (v - min) / range) * h,
  }))
})

const linePath = computed(() =>
  points.value.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(2)},${p.y.toFixed(2)}`).join(' '),
)

const areaPath = computed(() => {
  if (points.value.length < 2) return ''
  const first = points.value[0]!
  const last = points.value[points.value.length - 1]!
  return `${linePath.value} L${last.x.toFixed(2)},${props.height} L${first.x.toFixed(2)},${props.height} Z`
})
</script>
