<template>
  <!-- Availability notice — subtle, reuses the app's teal/cyan tokens.
       Turns green while live analysis is running, amber when it's offline. -->
  <div
    class="flex items-start gap-2.5 rounded-xl border px-4 py-2.5 text-sm transition-colors"
    :class="state.box"
  >
    <span class="mt-1 h-2 w-2 shrink-0 rounded-full" :class="state.dot" aria-hidden="true" />
    <div class="min-w-0">
      <p class="font-medium" :class="state.label">{{ state.title }}</p>
      <p class="mt-0.5 leading-relaxed text-gray-400">
        Live analysis runs Mon–Fri, 10:00–16:00 CET. Outside these hours, uploads return no results.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
// `null` until mounted so SSR and the first client render match (no hydration
// mismatch); the real live/offline state is computed on the client.
const live = ref<boolean | null>(null)
let timer: ReturnType<typeof setInterval> | null = null

// True when "now" falls within Mon–Fri 10:00–16:00 in Central European Time.
// Using the IANA zone means DST (CET/CEST) is handled automatically.
function isWithinServiceHours(): boolean {
  const parts = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Europe/Berlin',
    weekday: 'short',
    hour: 'numeric',
    hour12: false,
  }).formatToParts(new Date())

  const weekday = parts.find(p => p.type === 'weekday')?.value ?? ''
  let hour = Number(parts.find(p => p.type === 'hour')?.value ?? '0')
  if (hour === 24) hour = 0 // some engines report midnight as 24

  const isWeekday = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'].includes(weekday)
  return isWeekday && hour >= 10 && hour < 16
}

const state = computed(() => {
  if (live.value === null) {
    // Pre-mount / SSR: neutral teal, no live/offline claim yet.
    return {
      box: 'border-teal-500/20 bg-teal-950/20',
      dot: 'bg-teal-400/70',
      label: 'text-teal-100/80',
      title: 'Service hours',
    }
  }
  if (live.value) {
    return {
      box: 'border-emerald-500/25 bg-emerald-950/25',
      dot: 'live-dot bg-emerald-400',
      label: 'text-emerald-200',
      title: 'Analysis live now',
    }
  }
  return {
    box: 'border-amber-500/25 bg-amber-950/25',
    dot: 'bg-amber-400',
    label: 'text-amber-200',
    title: 'Currently offline',
  }
})

onMounted(() => {
  const update = () => { live.value = isWithinServiceHours() }
  update()
  timer = setInterval(update, 60_000) // re-check each minute so it flips at the boundary
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
