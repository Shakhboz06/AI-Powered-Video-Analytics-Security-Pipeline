<template>
  <div class="app-shell-bg app-grid relative flex min-h-screen font-[family-name:var(--font-outfit)]">
    <AmbientMesh />

    <div class="relative z-10 flex min-h-screen w-full flex-col lg:flex-row">
      <!-- Visual panel -->
      <aside class="relative hidden flex-1 flex-col justify-between overflow-hidden border-r p-10 lg:flex" style="border-color: var(--app-border)">
        <AppBrand to="/" size="lg" :show-tagline="true" />
        <div class="pointer-events-none absolute inset-0 opacity-40" style="background: radial-gradient(ellipse 80% 60% at 70% 50%, rgba(45,212,191,0.08), transparent)" />
        <div class="relative max-w-md">
          <p class="app-chip text-teal-300">
            <span class="live-dot h-1.5 w-1.5 rounded-full bg-teal-400" />
            AI-powered security
          </p>
          <h2 class="mt-5 text-3xl font-bold leading-tight tracking-tight text-gray-50">
            Your command center for <span class="app-gradient-text">real-time threat detection</span>
          </h2>
          <p class="mt-4 text-sm leading-relaxed text-gray-400">
            Monitor cameras, configure zones, and respond to fall, fight, and intrusion alerts — all from one intelligent dashboard.
          </p>
          <ul class="mt-8 space-y-3 text-sm text-gray-400">
            <li v-for="item in highlights" :key="item" class="flex items-center gap-2.5">
              <svg class="h-4 w-4 shrink-0 text-teal-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
              {{ item }}
            </li>
          </ul>
        </div>
        <p class="text-xs text-gray-600">© {{ year }} Security Ops Platform</p>
      </aside>

      <!-- Form panel -->
      <main class="flex flex-1 items-center justify-center p-4 sm:p-8">
        <div class="w-full max-w-md">
          <div class="mb-8 flex flex-col items-center text-center lg:hidden">
            <AppBrand size="lg" :show-tagline="true" />
          </div>
          <div class="app-card p-1">
            <slot />
          </div>
        </div>
      </main>
    </div>
    <ToastHost />
  </div>
</template>

<script setup lang="ts">
import AppBrand from '~/components/ui/AppBrand.vue'
import AmbientMesh from '~/components/ui/AmbientMesh.vue'
import ToastHost from '~/components/ui/ToastHost.vue'

const year = new Date().getFullYear()
const highlights = [
  'Sub-300ms inference on live feeds',
  'Fighting & fall detection built-in',
  'Zone-based intrusion alerting',
  'Streaming alert workflow',
]
</script>
