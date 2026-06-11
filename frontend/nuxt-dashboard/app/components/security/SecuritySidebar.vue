<template>
  <!-- Mobile backdrop -->
  <Transition name="fade">
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm lg:hidden"
      @click="closeSidebar"
    />
  </Transition>

  <aside
    class="sidebar-premium fixed inset-y-0 left-0 z-50 flex w-72 flex-col border-r app-glass transition-transform duration-300 ease-out lg:translate-x-0"
    :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
    style="border-color: var(--app-border)"
  >
    <!-- Brand -->
    <div class="flex h-16 items-center justify-between px-5">
      <div class="min-w-0" @click="closeSidebar">
        <AppBrand to="/dashboard" :show-tagline="true" />
      </div>
      <button
        type="button"
        class="rounded-lg p-1.5 text-gray-400 hover:bg-white/5 hover:text-gray-200 lg:hidden"
        aria-label="Close menu"
        @click="closeSidebar"
      >
        <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" /></svg>
      </button>
    </div>

    <div class="mx-5 h-px" style="background-image: linear-gradient(to right, transparent, var(--app-border), transparent)" />

    <!-- Nav -->
    <nav class="custom-scrollbar flex-1 space-y-1 overflow-y-auto px-3 py-5">
      <p class="px-3 pb-2 text-[10px] font-semibold uppercase tracking-[0.16em] text-gray-600">Operations</p>
      <NuxtLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="nav-link group"
        :class="{ 'nav-link-active': isActive(item.to) }"
        @click="closeSidebar"
      >
        <svg
          class="h-5 w-5 shrink-0 transition-transform duration-200 group-hover:scale-110"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path v-for="(d, i) in item.icon" :key="i" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="d" />
        </svg>
        <span class="flex-1 truncate">{{ item.label }}</span>
        <span
          v-if="item.badge && newAlertCount > 0"
          class="live-dot flex h-5 min-w-[1.25rem] items-center justify-center rounded-full bg-red-500 px-1.5 text-[10px] font-bold tabular-nums text-white"
        >{{ newAlertCount > 99 ? '99+' : newAlertCount }}</span>
      </NuxtLink>
    </nav>

    <!-- Footer / user -->
    <div class="border-t p-3" style="border-color: var(--app-border)">
      <div class="flex items-center gap-3 rounded-xl border bg-white/[0.02] px-3 py-2.5" style="border-color: var(--app-border)">
        <span
          class="avatar-ring grid h-9 w-9 shrink-0 place-items-center rounded-full text-sm font-bold text-[#04201c]"
          style="background-image: linear-gradient(135deg, #5eead4, #22d3ee 50%, #818cf8)"
        >{{ initials }}</span>
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm font-medium text-gray-100">{{ auth.user?.username || 'Operator' }}</p>
          <p class="truncate text-xs text-gray-500">{{ auth.user?.email || 'Signed in' }}</p>
        </div>
        <button
          type="button"
          class="rounded-lg p-2 text-gray-400 transition-colors hover:bg-red-500/10 hover:text-red-300"
          title="Log out"
          aria-label="Log out"
          @click="onLogout"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import AppBrand from '~/components/ui/AppBrand.vue'

const route = useRoute()
const auth = useAuthStore()
const router = useRouter()
const api = useApi()
const { sidebarOpen, closeSidebar } = useSecurityShell()

const navItems = [
  {
    to: '/dashboard',
    label: 'Live Monitoring',
    icon: [
      'M15 12a3 3 0 11-6 0 3 3 0 016 0z',
      'M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z',
    ],
  },
  {
    to: '/cameras',
    label: 'Cameras',
    icon: ['M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z'],
  },
  {
    to: '/alerts',
    label: 'Alerts',
    badge: true,
    icon: ['M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9'],
  },
  {
    to: '/analytics',
    label: 'Analytics',
    icon: ['M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'],
  },
  {
    to: '/zones',
    label: 'Zones',
    icon: ['M5 12l3.5-5h7L19 12l-3.5 5h-7L5 12z'],
  },
  {
    to: '/health',
    label: 'System Health',
    icon: ['M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z'],
  },
] as const

const newAlertCount = ref(0)
let alertStream: EventSource | null = null

const initials = computed(() => {
  const name = auth.user?.username || auth.user?.email || 'OP'
  return name.trim().slice(0, 2).toUpperCase()
})

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

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
