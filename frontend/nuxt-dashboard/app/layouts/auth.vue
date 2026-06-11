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

          <!-- Testimonial card -->
          <div class="mt-10 rounded-2xl border bg-white/[0.02] p-5" style="border-color: var(--app-border)">
            <div class="flex gap-1 text-amber-400">
              <svg v-for="i in 5" :key="i" class="h-4 w-4" fill="currentColor" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" /></svg>
            </div>
            <p class="mt-3 text-sm leading-relaxed text-gray-300">
              "Deployed Security Ops across 40 cameras in our warehouse. The fall detection alone has prevented three serious incidents."
            </p>
            <div class="mt-3 flex items-center gap-3">
              <span class="grid h-8 w-8 place-items-center rounded-full text-xs font-bold text-[#04201c]" style="background-image: linear-gradient(135deg, #5eead4, #22d3ee)">JM</span>
              <div>
                <p class="text-sm font-medium text-gray-200">James Mitchell</p>
                <p class="text-xs text-gray-500">Head of Security, LogiCorp</p>
              </div>
            </div>
          </div>
        </div>
        <p class="text-xs text-gray-600">&copy; {{ year }} Security Ops Platform</p>
      </aside>

      <!-- Form panel -->
      <main class="flex flex-1 items-center justify-center p-4 sm:p-8">
        <div class="w-full max-w-md">
          <div class="mb-8 flex flex-col items-center text-center lg:hidden">
            <AppBrand size="lg" :show-tagline="true" />
          </div>
          <div class="app-card overflow-hidden p-1">
            <slot />
          </div>
          <p class="mt-6 text-center text-xs text-gray-600">
            By signing in, you agree to our Terms and Privacy Policy.
          </p>
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
  'Multi-camera fleet management',
]
</script>
