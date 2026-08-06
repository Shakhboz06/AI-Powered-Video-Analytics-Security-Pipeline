<template>
  <div class="app-shell-bg app-grid relative min-h-screen overflow-x-hidden font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh />

    <header class="sticky top-0 z-30 nav-glass">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 md:px-8">
        <AppBrand to="/" :show-tagline="false" />
        <div class="hidden items-center gap-8 text-sm text-gray-400 md:flex">
          <a href="#pipeline" class="transition-colors hover:text-gray-100" @click.prevent="smoothScroll('#pipeline')">How it works</a>
          <a href="#capabilities" class="transition-colors hover:text-gray-100" @click.prevent="smoothScroll('#capabilities')">Detection</a>
          <a href="#stats" class="transition-colors hover:text-gray-100" @click.prevent="smoothScroll('#stats')">Platform</a>
          <NuxtLink to="/upload" class="transition-colors hover:text-gray-100">Analyze a video</NuxtLink>
        </div>
        <div class="flex items-center gap-2.5">
          <template v-if="isAuthenticated">
            <NuxtLink to="/dashboard" class="btn-primary">Open Dashboard</NuxtLink>
          </template>
          <template v-else>
            <NuxtLink to="/auth/login" class="btn-ghost hidden sm:inline-flex">Sign in</NuxtLink>
            <NuxtLink to="/auth/register" class="btn-primary">Get started</NuxtLink>
          </template>
          <!-- Mobile hamburger -->
          <button class="ml-2 grid h-9 w-9 place-items-center rounded-lg border border-white/10 text-gray-400 transition-colors hover:text-gray-100 md:hidden" @click="toggleMobileMenu" aria-label="Toggle menu">
            <svg v-if="!mobileMenuOpen" class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" /></svg>
            <svg v-else class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
      </nav>
    </header>

    <!-- Mobile menu drawer overlay -->
    <Transition name="mobile-menu">
      <div v-if="mobileMenuOpen" class="fixed inset-0 z-40 md:hidden" @click.self="mobileMenuOpen = false">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" />
        <div class="absolute right-0 top-0 flex h-full w-72 flex-col gap-2 border-l border-white/10 bg-[#0c1117]/95 p-6 pt-20 shadow-2xl backdrop-blur-xl">
          <button class="absolute right-4 top-4 grid h-9 w-9 place-items-center rounded-lg text-gray-400 transition-colors hover:text-gray-100" @click="mobileMenuOpen = false" aria-label="Close menu">
            <svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
          <a href="#pipeline" class="rounded-lg px-4 py-3 text-sm text-gray-300 transition-colors hover:bg-white/5 hover:text-gray-100" @click.prevent="smoothScroll('#pipeline'); mobileMenuOpen = false">How it works</a>
          <a href="#capabilities" class="rounded-lg px-4 py-3 text-sm text-gray-300 transition-colors hover:bg-white/5 hover:text-gray-100" @click.prevent="smoothScroll('#capabilities'); mobileMenuOpen = false">Detection</a>
          <a href="#stats" class="rounded-lg px-4 py-3 text-sm text-gray-300 transition-colors hover:bg-white/5 hover:text-gray-100" @click.prevent="smoothScroll('#stats'); mobileMenuOpen = false">Platform</a>
          <NuxtLink to="/upload" class="rounded-lg px-4 py-3 text-sm text-gray-300 transition-colors hover:bg-white/5 hover:text-gray-100" @click="mobileMenuOpen = false">Analyze a video</NuxtLink>
          <hr class="my-3 border-white/10" />
          <template v-if="isAuthenticated">
            <NuxtLink to="/dashboard" class="btn-primary w-full text-center" @click="mobileMenuOpen = false">Open Dashboard</NuxtLink>
          </template>
          <template v-else>
            <NuxtLink to="/auth/login" class="btn-ghost w-full text-center" @click="mobileMenuOpen = false">Sign in</NuxtLink>
            <NuxtLink to="/auth/register" class="btn-primary w-full text-center" @click="mobileMenuOpen = false">Get started</NuxtLink>
          </template>
        </div>
      </div>
    </Transition>

    <section class="relative z-10 mx-auto max-w-7xl px-5 pb-16 pt-10 md:px-8 md:pt-20">
      <div class="grid items-center gap-12 lg:grid-cols-2">
        <div class="animate-fade-up">
          <span class="app-chip border-teal-500/30 bg-teal-950/30 text-teal-300">
            <span class="live-dot h-1.5 w-1.5 rounded-full bg-teal-400" />
            AI video analytics platform
          </span>
          <h1 class="hero-display mt-5 font-bold text-gray-50">
            See every threat<br>
            <span class="app-gradient-text text-glow">before it escalates.</span>
          </h1>
          <p class="mt-5 max-w-xl text-base leading-relaxed text-gray-400 sm:text-lg">
            Security Ops turns any camera into an intelligent sensor — detecting intrusions, falls, fights,
            weapons and abandoned objects in real time, then routing alerts to your team instantly.
          </p>
          <div class="mt-8 flex flex-wrap items-center gap-3">
            <NuxtLink :to="isAuthenticated ? '/dashboard' : '/auth/register'" class="btn-primary px-5 py-3 text-base">
              {{ isAuthenticated ? 'Open Dashboard' : 'Start monitoring' }}
              <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" /></svg>
            </NuxtLink>
            <NuxtLink to="/upload" class="btn-ghost px-5 py-3 text-base">
              Try it — analyze a video
              <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" /></svg>
            </NuxtLink>
          </div>
          <div class="mt-8 flex flex-wrap items-center gap-x-6 gap-y-2 text-xs text-gray-500">
            <span v-for="t in trustBadges" :key="t" class="flex items-center gap-1.5">
              <svg class="h-4 w-4 text-emerald-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
              {{ t }}
            </span>
          </div>
        </div>

        <div class="float-slow relative">
          <div class="pointer-events-none absolute -left-6 top-8 z-10 hidden animate-fade-up rounded-xl border border-orange-500/30 bg-orange-950/40 px-3 py-2 text-xs text-orange-200 shadow-lg backdrop-blur-md lg:block" style="animation-delay: 0.3s">
            <span class="live-dot mr-1.5 inline-block h-1.5 w-1.5 rounded-full bg-orange-400" /> Fight detected
          </div>
          <div class="pointer-events-none absolute -right-4 bottom-16 z-10 hidden animate-fade-up rounded-xl border border-amber-500/30 bg-amber-950/40 px-3 py-2 text-xs text-amber-200 shadow-lg backdrop-blur-md lg:block" style="animation-delay: 0.5s">
            Fall alert · CAM-02
          </div>
          <div class="app-card overflow-hidden p-3">
            <div class="mb-2.5 flex items-center justify-between px-1">
              <span class="flex items-center gap-2 text-xs font-medium text-gray-300">
                <span class="live-dot h-2 w-2 rounded-full bg-red-500" /> LIVE · CAM-01
              </span>
              <!-- REC indicator -->
              <span class="flex items-center gap-1.5">
                <span class="flex items-center gap-1 rounded bg-red-600/80 px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-white rec-pulse">
                  <span class="rec-dot inline-block h-1.5 w-1.5 rounded-full bg-white" />
                  REC
                </span>
                <span class="font-mono text-[11px] text-gray-500">{{ clock }}</span>
              </span>
            </div>
            <!-- Gradient line under LIVE · CAM-01 -->
            <div class="mb-2 h-px w-full" style="background: linear-gradient(to right, transparent, rgba(45,212,191,0.5), rgba(251,146,60,0.3), transparent)" />
            <div class="monitor-feed relative aspect-video overflow-hidden rounded-xl">
              <div class="absolute inset-0" style="background: radial-gradient(120% 120% at 30% 10%, #14202b, #0a0e13 70%)" />
              <div class="absolute inset-0 opacity-40" style="background-image: linear-gradient(to right, rgba(255,255,255,0.05) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.05) 1px, transparent 1px); background-size: 28px 28px" />
              <div class="scan-line absolute left-0 right-0 h-px" style="background: linear-gradient(to right, transparent, #2dd4bf, transparent); box-shadow: 0 0 12px #2dd4bf" />
              <div class="absolute left-[14%] top-[30%] h-[46%] w-[22%] rounded border-2 border-teal-400" style="box-shadow: 0 0 16px rgba(45,212,191,0.4)">
                <span class="absolute -top-5 left-0 rounded bg-teal-400 px-1.5 py-0.5 text-[10px] font-bold text-[#04201c]">PERSON 98%</span>
              </div>
              <div class="absolute right-[16%] top-[40%] h-[40%] w-[26%] rounded border-2 border-orange-400" style="box-shadow: 0 0 16px rgba(251,146,60,0.45)">
                <span class="absolute -top-5 left-0 rounded bg-orange-400 px-1.5 py-0.5 text-[10px] font-bold text-[#04201c]">FIGHT</span>
              </div>
              <div class="absolute inset-x-0 bottom-0 grid grid-cols-3 gap-px border-t border-white/[0.06] bg-black/50 backdrop-blur-md">
                <div class="px-3 py-2 text-center"><p class="text-[9px] uppercase text-gray-500">Objects</p><p class="text-sm font-semibold text-white">12</p></div>
                <div class="px-3 py-2 text-center"><p class="text-[9px] uppercase text-gray-500">Latency</p><p class="text-sm font-semibold text-emerald-300">148ms</p></div>
                <div class="px-3 py-2 text-center"><p class="text-[9px] uppercase text-orange-300/80">Alerts</p><p class="text-sm font-semibold text-orange-300">2</p></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Marquee -->
    <section class="relative z-10 border-y py-4" style="border-color: var(--app-border)">
      <div class="overflow-hidden">
        <div class="marquee-track gap-8 px-4">
          <template v-for="repeat in 2" :key="repeat">
            <span v-for="tag in marqueeTags" :key="`${repeat}-${tag}`" class="stat-pill shrink-0 text-gray-400">{{ tag }}</span>
          </template>
        </div>
      </div>
    </section>

    <section id="pipeline" class="relative z-10 mx-auto max-w-7xl px-5 py-12 md:px-8">
      <div class="text-center">
        <p class="page-eyebrow">Architecture</p>
        <h2 class="mt-2 text-3xl font-bold tracking-tight text-gray-50">From frame to alert in milliseconds</h2>
        <p class="mx-auto mt-3 max-w-2xl text-gray-400">A streaming pipeline ingests video, runs AI inference, and surfaces actionable intelligence in your dashboard.</p>
      </div>
      <div class="stagger-in mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div v-for="(step, i) in pipeline" :key="step.title" class="app-card app-card-hover pipeline-card">
          <span class="app-gradient-text text-2xl font-bold tabular-nums">{{ String(i + 1).padStart(2, '0') }}</span>
          <h3 class="mt-3 text-sm font-semibold text-gray-100">{{ step.title }}</h3>
          <p class="mt-1.5 text-sm text-gray-500">{{ step.desc }}</p>
        </div>
      </div>
    </section>

    <section id="capabilities" class="relative z-10 mx-auto max-w-7xl px-5 py-12 md:px-8">
      <div class="text-center">
        <p class="page-eyebrow">Detection engine</p>
        <h2 class="mt-2 text-3xl font-bold tracking-tight text-gray-50 md:text-4xl">Seven threat types, one platform</h2>
        <p class="mx-auto mt-3 max-w-2xl text-gray-400">Computer-vision models and rule engines work together to flag what matters and ignore the noise.</p>
      </div>
      <div class="stagger-in mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="f in features" :key="f.title" class="app-card app-card-hover group p-5 md:p-6">
          <span class="grid h-11 w-11 place-items-center rounded-xl border transition-transform duration-300 group-hover:scale-110" :class="f.badge" style="border-color: var(--app-border)">
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path v-for="(d, i) in f.icon" :key="i" :d="d" /></svg>
          </span>
          <h3 class="mt-4 text-base font-semibold text-gray-100">{{ f.title }}</h3>
          <p class="mt-1.5 text-sm leading-relaxed text-gray-400">{{ f.desc }}</p>
        </div>
      </div>
    </section>

    <section id="stats" class="relative z-10 mx-auto max-w-7xl px-5 py-12 md:px-8">
      <div class="app-card grid gap-8 p-8 sm:grid-cols-2 lg:grid-cols-4">
        <div v-for="s in stats" :key="s.label" class="text-center">
          <p class="app-gradient-text text-4xl font-bold tracking-tight">{{ s.displayValue }}</p>
          <p class="mt-1 text-sm text-gray-400">{{ s.label }}</p>
        </div>
      </div>
    </section>

    <!-- Detection evidence gallery (replaces the logo wall) -->
    <DetectionGallery />

    <section class="relative z-10 mx-auto max-w-7xl px-5 py-16 md:px-8">
      <div class="app-card relative overflow-hidden p-10 text-center md:p-14">
        <div class="pointer-events-none absolute inset-0 opacity-60" style="background: radial-gradient(60% 120% at 50% 0%, rgba(45,212,191,0.14), transparent 70%)" />
        <div class="pointer-events-none absolute -right-20 -top-20 h-64 w-64 rounded-full bg-teal-500/10 blur-3xl" />
        <h2 class="relative text-3xl font-bold tracking-tight text-gray-50 md:text-4xl">Ready to secure your space?</h2>
        <p class="relative mx-auto mt-3 max-w-xl text-gray-400">Connect your first camera in minutes and start receiving intelligent alerts today.</p>
        <div class="relative mt-8 flex flex-wrap justify-center gap-3">
          <NuxtLink :to="isAuthenticated ? '/dashboard' : '/auth/register'" class="btn-primary px-6 py-3 text-base">
            {{ isAuthenticated ? 'Go to Dashboard' : 'Create free account' }}
          </NuxtLink>
          <NuxtLink v-if="!isAuthenticated" to="/auth/login" class="btn-ghost px-6 py-3 text-base">Sign in</NuxtLink>
        </div>
      </div>
    </section>

    <footer class="relative z-10 border-t" style="border-color: var(--app-border)">
      <div class="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 px-5 py-8 text-sm text-gray-500 md:flex-row md:px-8">
        <AppBrand size="sm" :show-tagline="false" />
        <div class="flex flex-wrap items-center justify-center gap-x-4 gap-y-1">
          <p>&copy; {{ year }} Security Ops &middot; Video Analytics Platform</p>
          <span class="inline-flex items-center gap-1 rounded-full border border-teal-500/20 bg-teal-950/30 px-2.5 py-0.5 text-[11px] font-medium text-teal-300">
            <svg class="h-3 w-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09z" /></svg>
            Built with Security
          </span>
        </div>
        <div class="flex items-center gap-5">
          <a href="#" class="transition-colors hover:text-gray-300">Privacy</a>
          <a href="#" class="transition-colors hover:text-gray-300">Terms</a>
          <a href="#" class="transition-colors hover:text-gray-300">API Docs</a>
          <NuxtLink to="/dashboard" class="transition-colors hover:text-gray-300">Dashboard</NuxtLink>
          <NuxtLink to="/auth/login" class="transition-colors hover:text-gray-300">Sign in</NuxtLink>
      </div>
    </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import AppBrand from '~/components/ui/AppBrand.vue'
import AmbientMesh from '~/components/ui/AmbientMesh.vue'
import DetectionGallery from '~/components/DetectionGallery.vue'

const marqueeTags = [
  'YOLO Detection', 'Pose Estimation', 'Fall ML', 'Fight Classification',
  'Zone Intrusion', 'Weapon Detection', 'Abandoned Object', 'Real-time SSE',
  'TimescaleDB', 'Kafka Pipeline', 'Sub-300ms Latency',
]

definePageMeta({ layout: false })

const auth = useAuthStore()
const isAuthenticated = computed(() => auth.isAuthenticated)
const year = new Date().getFullYear()

const trustBadges = ['Sub-second detection', 'Multi-camera fleet', '24/7 alerting', 'ML + rules hybrid']


const clock = ref('--:--:--')
let timer: ReturnType<typeof setInterval> | null = null

// Mobile menu state
const mobileMenuOpen = ref(false)
function toggleMobileMenu() {
  mobileMenuOpen.value = !mobileMenuOpen.value
}

// Smooth scroll helper
function smoothScroll(selector: string) {
  document.querySelector(selector)?.scrollIntoView({ behavior: 'smooth' })
}

// Animated stats
interface StatItem {
  numericValue: number
  prefix: string
  suffix: string
  label: string
  displayValue: string
}

const stats = reactive<StatItem[]>([
  { numericValue: 7, prefix: '', suffix: '+', label: 'Threat types detected', displayValue: '0+' },
  { numericValue: 300, prefix: '<', suffix: 'ms', label: 'Inference latency', displayValue: '<0ms' },
  { numericValue: 24, prefix: '', suffix: '/7', label: 'Continuous monitoring', displayValue: '0/7' },
  { numericValue: 0, prefix: '', suffix: '∞', label: 'Cameras supported', displayValue: '∞' },
])

let statsAnimated = false

function animateStats() {
  if (statsAnimated) return
  statsAnimated = true

  const duration = 1500
  const startTime = performance.now()

  function update(currentTime: number) {
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)
    // Ease out cubic
    const eased = 1 - Math.pow(1 - progress, 3)

    for (const s of stats) {
      if (s.numericValue === 0) {
        // Special case for infinity symbol - always show it
        s.displayValue = s.suffix
      } else {
        const current = Math.round(eased * s.numericValue)
        s.displayValue = `${s.prefix}${current}${s.suffix}`
      }
    }

    if (progress < 1) {
      requestAnimationFrame(update)
    }
  }

  requestAnimationFrame(update)
}

onMounted(() => {
  // Clock
  const tick = () => { clock.value = new Date().toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' }) }
  tick()
  timer = setInterval(tick, 1000)

  // Stats intersection observer
  const statsEl = document.querySelector('#stats')
  if (statsEl) {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            animateStats()
            observer.disconnect()
          }
        }
      },
      { threshold: 0.3 },
    )
    observer.observe(statsEl)
  }
})
onUnmounted(() => { if (timer) clearInterval(timer) })

const pipeline = [
  { title: 'Ingest', desc: 'Frames pulled from RTSP, files, or USB cameras into Kafka.' },
  { title: 'Detect', desc: 'YOLO + pose models classify objects, falls, and fights.' },
  { title: 'Aggregate', desc: 'Zone rules, cooldowns, and deduplication produce clean alerts.' },
  { title: 'Respond', desc: 'Stream alerts to dashboard, email, and operator workflows.' },
]

const features = [
  { title: 'Object & people tracking', desc: 'YOLO-powered detection with persistent tracker IDs across frames and cameras.', badge: 'bg-sky-500/15 text-sky-300', icon: ['M15 12a3 3 0 11-6 0 3 3 0 016 0z', 'M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z'] },
  { title: 'Zone intrusion', desc: 'Draw restricted areas with schedules; get alerted the instant someone steps in.', badge: 'bg-red-500/15 text-red-300', icon: ['M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z'] },
  { title: 'Fall detection', desc: 'Pose-based ML spots people who collapse — critical for care and safety.', badge: 'bg-amber-500/15 text-amber-300', icon: ['M2.25 6 9 12.75l4.306-4.307a11.95 11.95 0 015.814 5.519l.93 1.95'] },
  { title: 'Fight detection', desc: 'Video-classification models recognise physical altercations between people.', badge: 'bg-orange-500/15 text-orange-300', icon: ['M15.362 5.214A8.252 8.252 0 0112 21 8.25 8.25 0 016.038 7.047 8.287 8.287 0 009 9.601a8.983 8.983 0 013.361-6.867 8.21 8.21 0 003 2.48z'] },
  { title: 'Weapon & abandoned objects', desc: 'Flag brandished weapons and unattended bags left stationary in a scene.', badge: 'bg-violet-500/15 text-violet-300', icon: ['M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z'] },
  { title: 'Live analytics & alerts', desc: 'Streaming dashboards, latency health, and instant acknowledge/resolve workflows.', badge: 'bg-teal-500/15 text-teal-300', icon: ['M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z'] },
]
</script>

<style scoped>
/* REC indicator pulsing animation */
.rec-pulse {
  animation: rec-pulse-anim 1.5s ease-in-out infinite;
}
.rec-dot {
  animation: rec-dot-blink 1.5s ease-in-out infinite;
}
@keyframes rec-pulse-anim {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
@keyframes rec-dot-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.2; }
}

/* Mobile menu transitions */
.mobile-menu-enter-active,
.mobile-menu-leave-active {
  transition: opacity 0.25s ease;
}
.mobile-menu-enter-active > div:last-child,
.mobile-menu-leave-active > div:last-child {
  transition: transform 0.25s ease;
}
.mobile-menu-enter-from,
.mobile-menu-leave-to {
  opacity: 0;
}
.mobile-menu-enter-from > div:last-child {
  transform: translateX(100%);
}
.mobile-menu-leave-to > div:last-child {
  transform: translateX(100%);
}
</style>
