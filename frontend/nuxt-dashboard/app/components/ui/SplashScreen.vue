<template>
  <Transition name="overlay">
    <div
      v-if="visible"
      class="app-shell-bg fixed inset-0 z-[200] flex flex-col items-center justify-center overflow-hidden"
      aria-label="Loading"
      role="status"
    >
      <AmbientMesh :particles="false" />

      <div class="relative z-10 flex flex-col items-center">
        <div class="relative">
          <div class="orbit-ring rounded-3xl" />
          <div class="orbit-ring orbit-ring-2 rounded-3xl" />
          <div class="absolute inset-0 animate-ping rounded-3xl bg-teal-400/8" style="animation-duration: 2.8s" />
          <span
            class="relative grid h-[5.5rem] w-[5.5rem] place-items-center rounded-3xl text-[#04201c]"
            style="background-image: linear-gradient(135deg, #2dd4bf, #22d3ee 50%, #818cf8); box-shadow: 0 0 80px -8px rgba(45,212,191,0.7), 0 0 120px -20px rgba(129,140,248,0.4)"
          >
            <svg class="h-11 w-11" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path d="M12 3L4 7v10l8 4 8-4V7l-8-4z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
              <path d="M12 12l8-4M12 12v10M12 12L4 8" stroke="currentColor" stroke-width="1.5" />
            </svg>
          </span>
        </div>

        <h1 class="app-gradient-text text-glow mt-8 text-4xl font-bold tracking-tight">Security Ops</h1>
        <p class="mt-2 text-xs font-semibold uppercase tracking-[0.32em] text-gray-500">AI Video Analytics</p>

        <p class="mt-5 min-h-[1.25rem] text-sm text-gray-400 transition-all duration-300">
          {{ statusText }}
        </p>

        <div class="mt-6 h-1.5 w-56 overflow-hidden rounded-full border border-white/[0.08] bg-white/[0.03] p-px">
          <div
            class="h-full rounded-full transition-all duration-300 ease-out"
            :style="{
              width: `${progress}%`,
              backgroundImage: 'linear-gradient(to right, #2dd4bf, #22d3ee, #818cf8)',
              boxShadow: '0 0 16px rgba(45,212,191,0.55)',
            }"
          />
        </div>
        <p class="mt-2.5 font-mono text-[10px] tabular-nums tracking-widest text-gray-600">{{ Math.round(progress) }}%</p>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import AmbientMesh from '~/components/ui/AmbientMesh.vue'

const visible = ref(true)
const progress = ref(0)

const statusText = computed(() => {
  if (progress.value < 30) return 'Initializing vision models…'
  if (progress.value < 55) return 'Loading detection pipeline…'
  if (progress.value < 80) return 'Syncing analytics engine…'
  if (progress.value < 98) return 'Preparing your workspace…'
  return 'Welcome back'
})

let progressTimer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  progressTimer = setInterval(() => {
    if (progress.value < 90) {
      progress.value += Math.random() * 6 + 2
    }
    else if (progress.value < 100) {
      progress.value = Math.min(100, progress.value + 1.5)
    }
  }, 100)

  setTimeout(() => {
    progress.value = 100
    setTimeout(() => { visible.value = false }, 400)
  }, 1800)
})

onUnmounted(() => {
  if (progressTimer) clearInterval(progressTimer)
})
</script>
