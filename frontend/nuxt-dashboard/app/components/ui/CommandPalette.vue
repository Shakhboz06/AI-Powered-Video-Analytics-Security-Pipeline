<template>
  <Teleport to="body">
    <Transition name="overlay">
      <div
        v-if="open"
        class="fixed inset-0 z-[95] flex items-start justify-center bg-black/60 p-4 pt-[12vh] backdrop-blur-sm"
        @click.self="hide"
      >
        <Transition name="pop" appear>
          <div
            v-if="open"
            class="w-full max-w-lg overflow-hidden rounded-2xl border border-white/10 bg-[#0c1117] shadow-2xl"
            role="dialog"
            aria-modal="true"
            aria-label="Command palette"
          >
            <div class="flex items-center gap-3 border-b border-white/[0.06] px-4">
              <svg class="h-4 w-4 shrink-0 text-gray-500" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" /></svg>
              <input
                ref="inputEl"
                v-model="query"
                type="text"
                placeholder="Search pages and actions…"
                class="w-full bg-transparent py-3.5 text-sm text-gray-100 outline-none placeholder:text-gray-600"
                @keydown.down.prevent="move(1)"
                @keydown.up.prevent="move(-1)"
                @keydown.enter.prevent="runActive()"
                @keydown.esc="hide"
              >
              <kbd class="hidden rounded border border-white/10 bg-white/5 px-1.5 py-0.5 text-[10px] text-gray-500 sm:inline">ESC</kbd>
            </div>

            <div class="max-h-[50vh] overflow-y-auto p-2 custom-scrollbar">
              <p v-if="!results.length" class="px-3 py-8 text-center text-sm text-gray-500">No results for "{{ query }}"</p>
              <template v-for="(group, gi) in grouped" :key="group.label">
                <p v-if="group.items.length" class="px-3 pb-1 pt-3 text-[10px] font-semibold uppercase tracking-[0.16em] text-gray-600">{{ group.label }}</p>
                <button
                  v-for="item in group.items"
                  :key="item.id"
                  type="button"
                  class="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left text-sm transition-colors"
                  :class="item.index === activeIndex ? 'bg-teal-500/15 text-teal-200' : 'text-gray-300 hover:bg-white/5'"
                  @click="run(item)"
                  @mousemove="activeIndex = item.index"
                >
                  <svg class="h-4 w-4 shrink-0" :class="item.index === activeIndex ? 'text-teal-300' : 'text-gray-500'" fill="none" stroke="currentColor" stroke-width="1.6" viewBox="0 0 24 24"><path v-for="(d, i) in item.icon" :key="i" stroke-linecap="round" stroke-linejoin="round" :d="d" /></svg>
                  <span class="flex-1 truncate">{{ item.label }}</span>
                  <span v-if="item.hint" class="text-xs text-gray-600">{{ item.hint }}</span>
                </button>
                <div v-if="group.items.length && gi < grouped.length - 1" class="my-1" />
              </template>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
interface Command {
  id: string
  label: string
  group: 'Navigation' | 'Actions'
  icon: string[]
  hint?: string
  run: () => void
}

const { open, hide, toggle } = useCommandPalette()
const router = useRouter()
const { toggle: toggleTheme } = useTheme()

const query = ref('')
const activeIndex = ref(0)
const inputEl = ref<HTMLInputElement | null>(null)

const NAV_ICON: Record<string, string[]> = {
  '/': ['M15 12a3 3 0 11-6 0 3 3 0 016 0z', 'M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z'],
  '/cameras': ['M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z'],
  '/alerts': ['M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5'],
  '/analytics': ['M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z'],
  '/zones': ['M5 12l3.5-5h7L19 12l-3.5 5h-7L5 12z'],
  '/health': ['M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z'],
}

const commands = computed<Command[]>(() => [
  { id: 'nav-home', label: 'Live Monitoring', group: 'Navigation', icon: NAV_ICON['/']!, run: () => router.push('/dashboard') },
  { id: 'nav-cameras', label: 'Cameras', group: 'Navigation', icon: NAV_ICON['/cameras']!, run: () => router.push('/cameras') },
  { id: 'nav-alerts', label: 'Alerts', group: 'Navigation', icon: NAV_ICON['/alerts']!, run: () => router.push('/alerts') },
  { id: 'nav-analytics', label: 'Analytics', group: 'Navigation', icon: NAV_ICON['/analytics']!, run: () => router.push('/analytics') },
  { id: 'nav-zones', label: 'Zones', group: 'Navigation', icon: NAV_ICON['/zones']!, run: () => router.push('/zones') },
  { id: 'nav-health', label: 'System Health', group: 'Navigation', icon: NAV_ICON['/health']!, run: () => router.push('/health') },
  {
    id: 'action-theme',
    label: 'Toggle light / dark theme',
    group: 'Actions',
    hint: 'Theme',
    icon: ['M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z'],
    run: () => toggleTheme(),
  },
])

function fuzzy(text: string, q: string) {
  const t = text.toLowerCase()
  const s = q.toLowerCase().trim()
  if (!s) return true
  let i = 0
  for (const ch of t) {
    if (ch === s[i]) i++
    if (i === s.length) return true
  }
  return t.includes(s)
}

const results = computed(() => commands.value.filter((c) => fuzzy(c.label, query.value)))

const grouped = computed(() => {
  let idx = 0
  const groups: { label: string; items: (Command & { index: number })[] }[] = [
    { label: 'Navigation', items: [] },
    { label: 'Actions', items: [] },
  ]
  for (const c of results.value) {
    const target = groups.find((g) => g.label === c.group)!
    target.items.push({ ...c, index: idx++ })
  }
  return groups
})

function move(delta: number) {
  const n = results.value.length
  if (!n) return
  activeIndex.value = (activeIndex.value + delta + n) % n
}

function run(item: Command) {
  item.run()
  hide()
}

function runActive() {
  const flat = grouped.value.flatMap((g) => g.items)
  const item = flat.find((i) => i.index === activeIndex.value)
  if (item) run(item)
}

watch(open, async (v) => {
  if (v) {
    query.value = ''
    activeIndex.value = 0
    await nextTick()
    inputEl.value?.focus()
  }
})

watch(query, () => { activeIndex.value = 0 })

function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    toggle()
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>
