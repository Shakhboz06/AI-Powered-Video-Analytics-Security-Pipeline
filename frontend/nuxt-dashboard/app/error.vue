<template>
  <div class="app-shell-bg app-grid relative flex min-h-screen items-center justify-center p-6 font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh :particles="false" />
    <div class="relative z-10 app-card max-w-md p-10 text-center">
      <p class="app-gradient-text text-6xl font-bold tabular-nums">{{ statusCode }}</p>
      <h1 class="mt-3 text-xl font-semibold text-gray-100">{{ title }}</h1>
      <p class="mt-2 text-sm text-gray-500">{{ message }}</p>
      <div class="mt-8 flex flex-wrap justify-center gap-3">
        <button type="button" class="btn-primary" @click="goHome">Try again</button>
        <NuxtLink to="/" class="btn-ghost">Back to home</NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'
import AmbientMesh from '~/components/ui/AmbientMesh.vue'

const props = defineProps<{
  error: NuxtError
}>()

const statusCode = computed(() => props.error?.statusCode || 500)

const title = computed(() => {
  const code = statusCode.value
  if (code === 404) return 'Page not found'
  if (code === 403) return 'Access denied'
  return 'Something went wrong'
})

const message = computed(() => {
  if (statusCode.value === 404) {
    return 'The page you are looking for does not exist or has been moved.'
  }
  return props.error?.message || 'An unexpected error occurred. Please try again.'
})

function goHome() {
  clearError({ redirect: '/' })
}
</script>
