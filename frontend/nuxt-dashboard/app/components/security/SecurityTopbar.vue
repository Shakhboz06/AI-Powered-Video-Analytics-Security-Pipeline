<template>
  <header
    class="sticky top-0 z-30 flex h-16 items-center gap-3 border-b app-glass px-4 md:px-6"
    style="border-color: var(--app-border)"
  >
    <button
      type="button"
      class="rounded-lg p-2 text-gray-400 hover:bg-white/5 hover:text-gray-200 lg:hidden"
      aria-label="Open menu"
      @click="toggleSidebar"
    >
      <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h16" /></svg>
    </button>

    <div class="min-w-0">
      <div class="flex items-center gap-2 text-[11px] text-gray-500">
        <span>Security Ops</span>
        <span class="text-gray-700">/</span>
        <span class="text-gray-400">{{ pageTitle }}</span>
      </div>
      <h1 class="truncate text-base font-semibold text-gray-100">{{ pageTitle }}</h1>
    </div>

    <div class="ml-auto flex items-center gap-2 sm:gap-3">
      <button
        type="button"
        class="hidden items-center gap-2 rounded-xl border px-3 py-1.5 text-xs text-gray-400 transition-colors hover:text-gray-200 md:inline-flex"
        style="border-color: var(--app-border); background-color: rgba(255, 255, 255, 0.02)"
        @click="palette.show()"
      >
        <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" /></svg>
        <span>Search</span>
        <kbd class="rounded border border-white/10 bg-white/5 px-1.5 py-0.5 text-[10px] font-medium text-gray-500">{{ metaKeyLabel }}K</kbd>
      </button>

      <button
        type="button"
        class="rounded-xl border p-2 text-gray-400 transition-colors hover:text-gray-200"
        style="border-color: var(--app-border); background-color: rgba(255, 255, 255, 0.02)"
        :data-tooltip="theme === 'dark' ? 'Light mode' : 'Dark mode'"
        aria-label="Toggle theme"
        @click="toggleTheme"
      >
        <svg v-if="theme === 'dark'" class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" /></svg>
        <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.7" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z" /></svg>
      </button>

      <span
        class="hidden items-center gap-2 rounded-full border px-3 py-1.5 text-xs font-medium text-emerald-300 sm:inline-flex"
        style="border-color: rgba(16, 185, 129, 0.35); background-color: rgba(16, 185, 129, 0.08)"
      >
        <span class="live-dot h-1.5 w-1.5 rounded-full bg-emerald-400" />
        System Online
      </span>

      <div
        class="hidden items-center gap-2 rounded-full border px-3 py-1.5 text-xs md:inline-flex"
        style="border-color: var(--app-border); background-color: rgba(255, 255, 255, 0.02)"
      >
        <svg class="h-3.5 w-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span class="font-mono tabular-nums text-gray-300">{{ clock }}</span>
      </div>

      <span
        class="avatar-ring grid h-9 w-9 place-items-center rounded-full text-sm font-bold text-[#04201c]"
        style="background-image: linear-gradient(135deg, #5eead4, #22d3ee 50%, #818cf8)"
        :title="auth.user?.username || 'Operator'"
      >{{ initials }}</span>
    </div>
  </header>
</template>

<script setup lang="ts">
const route = useRoute()
const auth = useAuthStore()
const { toggleSidebar } = useSecurityShell()
const palette = useCommandPalette()
const { theme, toggle: toggleTheme } = useTheme()

const metaKeyLabel = ref('Ctrl ')
onMounted(() => {
  if (navigator.platform.toLowerCase().includes('mac')) metaKeyLabel.value = '\u2318'
})

const TITLES: Record<string, string> = {
  '/dashboard': 'Live Monitoring',
  '/cameras': 'Cameras',
  '/alerts': 'Alerts',
  '/analytics': 'Analytics',
  '/zones': 'Zones',
  '/health': 'System Health',
}

const pageTitle = computed(() => {
  const path = route.path
  if (TITLES[path]) return TITLES[path]
  const match = Object.keys(TITLES).find((p) => path.startsWith(p))
  return match ? TITLES[match] : 'Dashboard'
})

const initials = computed(() => {
  const name = auth.user?.username || auth.user?.email || 'OP'
  return name.trim().slice(0, 2).toUpperCase()
})

const clock = ref('')
let timer: ReturnType<typeof setInterval> | null = null

function tick() {
  clock.value = new Date().toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

onMounted(() => {
  tick()
  timer = setInterval(tick, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
