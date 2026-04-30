<template>
  <aside
    class="fixed inset-y-0 left-0 z-40 w-64 border-r border-gray-800/80 bg-[#12181f] flex flex-col"
  >
    <div class="h-16 flex items-center px-5 border-b border-gray-800/80">
      <NuxtLink to="/" class="flex items-center gap-2 text-teal-400 font-semibold tracking-tight">
        <svg class="w-8 h-8 shrink-0" viewBox="0 0 24 24" fill="none" aria-hidden="true">
          <path
            d="M12 3L4 7v10l8 4 8-4V7l-8-4z"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linejoin="round"
          />
          <path d="M12 12l8-4M12 12v10M12 12L4 8" stroke="currentColor" stroke-width="1.5" />
        </svg>
        <span class="text-gray-100">Security Ops</span>
      </NuxtLink>
    </div>

    <nav class="flex-1 px-3 py-6 space-y-1">
      <NuxtLink
        to="/"
        class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
        :class="isActive('/') ? 'bg-teal-500/15 text-teal-300' : 'text-gray-400 hover:bg-gray-800/60 hover:text-gray-200'"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        </svg>
        Live Monitoring
      </NuxtLink>
      <NuxtLink
        to="/cameras"
        class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
        :class="isActive('/cameras') ? 'bg-teal-500/15 text-teal-300' : 'text-gray-400 hover:bg-gray-800/60 hover:text-gray-200'"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        Cameras
      </NuxtLink>
      <NuxtLink
        to="/alerts"
        class="flex items-center justify-between gap-2 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
        :class="isActive('/alerts') ? 'bg-teal-500/15 text-teal-300' : 'text-gray-400 hover:bg-gray-800/60 hover:text-gray-200'"
      >
        <span class="flex items-center gap-3 min-w-0">
          <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          <span class="truncate">Alerts</span>
        </span>
        <span
          v-if="newAlertCount > 0"
          class="shrink-0 min-w-[1.25rem] h-5 px-1.5 flex items-center justify-center rounded-full bg-red-600 text-[10px] font-semibold text-white tabular-nums"
        >{{ newAlertCount > 99 ? '99+' : newAlertCount }}</span>
      </NuxtLink>
      <NuxtLink
        to="/analytics"
        class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
        :class="isActive('/analytics') ? 'bg-teal-500/15 text-teal-300' : 'text-gray-400 hover:bg-gray-800/60 hover:text-gray-200'"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
        Analytics
      </NuxtLink>
      <NuxtLink
        to="/zones"
        class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
        :class="isActive('/zones') ? 'bg-teal-500/15 text-teal-300' : 'text-gray-400 hover:bg-gray-800/60 hover:text-gray-200'"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 12l3.5-5h7L19 12l-3.5 5h-7L5 12z" />
        </svg>
        Zones
      </NuxtLink>
      <NuxtLink
        to="/health"
        class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
        :class="isActive('/health') ? 'bg-teal-500/15 text-teal-300' : 'text-gray-400 hover:bg-gray-800/60 hover:text-gray-200'"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
        </svg>
        System Health
      </NuxtLink>
    </nav>

    <div class="p-4 border-t border-gray-800/80 space-y-3">
      <div class="rounded-lg bg-gray-900/50 px-3 py-2 border border-gray-800/80">
        <p class="text-xs text-gray-500 uppercase tracking-wide">Signed in</p>
        <p class="text-sm text-gray-200 truncate">{{ auth.user?.username || '—' }}</p>
        <p class="text-xs text-gray-500 truncate">{{ auth.user?.email }}</p>
      </div>
      <button
        type="button"
        class="w-full flex items-center justify-center gap-2 rounded-lg border border-gray-700 bg-gray-900/40 px-3 py-2 text-sm text-gray-300 hover:bg-red-950/40 hover:border-red-800/60 hover:text-red-300 transition-colors"
        @click="onLogout"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
        </svg>
        Log out
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
const route = useRoute()
const auth = useAuthStore()
const router = useRouter()
const api = useApi()

const newAlertCount = ref(0)
let alertStream: EventSource | null = null

onMounted(() => {
  if (!import.meta.client) return
  alertStream = new EventSource(api.alertStreamUrl())
  alertStream.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data) as { status?: string }
      if (!payload.status || payload.status === 'new') newAlertCount.value++
    }
    catch {
      newAlertCount.value++
    }
  }
})

onUnmounted(() => {
  if (alertStream) alertStream.close()
})

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

function onLogout() {
  auth.logout()
  router.push('/auth/login')
}
</script>
