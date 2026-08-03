<template>
  <span class="tabular-nums">{{ display }}</span>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  value: number
  decimals?: number
  duration?: number
}>(), {
  decimals: 0,
  duration: 700,
})

const current = ref(props.value)
let raf: number | null = null

const display = computed(() => {
  const n = current.value
  return n.toLocaleString(undefined, {
    minimumFractionDigits: props.decimals,
    maximumFractionDigits: props.decimals,
  })
})

function animateTo(target: number) {
  if (!import.meta.client) {
    current.value = target
    return
  }
  if (raf) cancelAnimationFrame(raf)
  const start = current.value
  const delta = target - start
  if (delta === 0) return
  const startTime = performance.now()
  const step = (now: number) => {
    const t = Math.min(1, (now - startTime) / props.duration)
    const eased = 1 - Math.pow(1 - t, 3)
    current.value = start + delta * eased
    if (t < 1) raf = requestAnimationFrame(step)
    else current.value = target
  }
  raf = requestAnimationFrame(step)
}

onMounted(() => animateTo(props.value))
watch(() => props.value, (v) => animateTo(v))
onUnmounted(() => {
  if (raf) cancelAnimationFrame(raf)
})
</script>
