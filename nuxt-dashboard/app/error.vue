<template>
  <div class="app-shell-bg app-grid relative flex min-h-screen items-center justify-center p-6 font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh :particles="false" />
    <div class="relative z-10 max-w-lg text-center">
      <div class="relative mx-auto mb-8 h-48 w-48">
        <div class="absolute inset-0 animate-pulse rounded-full bg-gradient-to-br from-teal-500/10 via-cyan-500/5 to-indigo-500/10 blur-2xl" />
        <div class="relative flex h-full items-center justify-center">
          <p class="app-gradient-text text-8xl font-bold tabular-nums tracking-tight" style="text-shadow: 0 0 80px rgba(45, 212, 191, 0.3)">{{ statusCode }}</p>
        </div>
        <div class="absolute -inset-4 rounded-full border border-teal-500/10" style="animation: app-glow-pulse 3s ease-in-out infinite" />
        <div class="absolute -inset-8 rounded-full border border-indigo-500/5" style="animation: app-glow-pulse 3s ease-in-out infinite 1.5s" />
      </div>

      <h1 class="text-2xl font-bold tracking-tight text-gray-50">{{ title }}</h1>
      <p class="mx-auto mt-3 max-w-md text-base text-gray-400">{{ message }}</p>

      <div v-if="statusCode === 404" class="mx-auto mt-6 max-w-xs">
        <div class="app-card p-4">
          <p class="text-xs font-medium uppercase tracking-wide text-gray-500">Requested path</p>
          <p class="mt-1 truncate font-mono text-sm text-gray-300">{{ requestedPath }}</p>
        </div>
      </div>

      <div class="mt-8 flex flex-wrap justify-center gap-3">
        <button type="button" class="btn-primary px-6 py-3" @click="goHome">
          <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" /></svg>
          Go home
        </button>
        <button type="button" class="btn-ghost px-6 py-3" @click="goBack">
          <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 15L3 9m0 0l6-6M3 9h12a6 6 0 010 12h-3" /></svg>
          Go back
        </button>
      </div>

      <p class="mt-10 text-xs text-gray-600">
        If this problem persists, contact your system administrator.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'
import AmbientMesh from '~/components/ui/AmbientMesh.vue'

const props = defineProps<{
  error: NuxtError
}>()

const route = useRoute()

const statusCode = computed(() => props.error?.statusCode || 500)
const requestedPath = computed(() => route.fullPath || '/')

const title = computed(() => {
  const code = statusCode.value
  if (code === 404) return 'Page not found'
  if (code === 403) return 'Access denied'
  if (code === 401) return 'Authentication required'
  if (code >= 500) return 'Server error'
  return 'Something went wrong'
})

const message = computed(() => {
  const code = statusCode.value
  if (code === 404) return 'The page you\'re looking for doesn\'t exist or has been moved to a different location.'
  if (code === 403) return 'You don\'t have permission to access this resource. Please sign in with an authorized account.'
  if (code === 401) return 'Your session may have expired. Please sign in again to continue.'
  return props.error?.message || 'An unexpected error occurred. Our team has been notified.'
})

function goHome() {
  clearError({ redirect: '/' })
}

function goBack() {
  if (window.history.length > 1) {
    window.history.back()
  } else {
    clearError({ redirect: '/' })
  }
}
</script>
