<template>
  <Transition name="overlay">
    <div
      v-if="visible"
      class="app-shell-bg fixed inset-0 z-[200] flex flex-col items-center justify-center overflow-hidden"
      aria-label="Loading"
      role="status"
    >
      <div class="relative z-10 flex flex-col items-center">
        <div class="relative">
          <div class="orbit-ring rounded-3xl" />
          <div class="orbit-ring orbit-ring-2 rounded-3xl" />
          <span
            class="relative grid h-[5.5rem] w-[5.5rem] place-items-center rounded-3xl text-[#04201c]"
            style="background-image: linear-gradient(135deg, #2dd4bf, #22d3ee 50%, #818cf8); box-shadow: 0 0 80px -8px rgba(45,212,191,0.7)"
          >
            <svg class="h-11 w-11" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path d="M12 3L4 7v10l8 4 8-4V7l-8-4z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
              <path d="M12 12l8-4M12 12v10M12 12L4 8" stroke="currentColor" stroke-width="1.5" />
            </svg>
          </span>
        </div>

        <h1 class="app-gradient-text mt-8 text-3xl font-bold tracking-tight">Security Ops</h1>
        <p class="mt-2 text-xs font-semibold uppercase tracking-[0.32em] text-gray-500">AI Video Analytics</p>

        <p class="mt-5 min-h-[1.25rem] text-sm text-gray-400">{{ statusText }}</p>

        <div class="mt-6 h-1.5 w-56 overflow-hidden rounded-full border border-white/[0.08] bg-white/[0.03] p-px">
          <div
            class="h-full rounded-full transition-all duration-300 ease-out"
            :style="{
              width: `${progress}%`,
              backgroundImage: 'linear-gradient(to right, #2dd4bf, #22d3ee, #818cf8)',
            }"
          />
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
const SPLASH_KEY = 'security-ops-splash-seen'

const visible = ref(false)
const progress = ref(0)

const statusText = computed(() => {
  if (progress.value < 40) return 'Initializing…'
  if (progress.value < 75) return 'Loading workspace…'
  if (progress.value < 98) return 'Almost ready…'
  return 'Welcome'
})

let progressTimer: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  if (sessionStorage.getItem(SPLASH_KEY)) {
    return
  }

  visible.value = true
  progressTimer = setInterval(() => {
    if (progress.value < 95) {
      progress.value += 8
    }
    else {
      progress.value = 100
    }
  }, 80)

  setTimeout(() => {
    progress.value = 100
    sessionStorage.setItem(SPLASH_KEY, '1')
    setTimeout(() => { visible.value = false }, 300)
    if (progressTimer) clearInterval(progressTimer)
  }, 1200)
})

onUnmounted(() => {
  if (progressTimer) clearInterval(progressTimer)
})
</script>
