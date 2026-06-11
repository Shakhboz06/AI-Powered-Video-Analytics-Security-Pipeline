<template>
  <div class="app-shell-bg app-grid relative min-h-screen overflow-x-hidden font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh />

    <header class="sticky top-0 z-30 nav-glass">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 md:px-8">
        <AppBrand to="/" :show-tagline="false" />
        <div class="hidden items-center gap-8 text-sm text-gray-400 md:flex">
          <a href="#pipeline" class="transition-colors hover:text-gray-100">How it works</a>
          <a href="#capabilities" class="transition-colors hover:text-gray-100">Detection</a>
          <a href="#stats" class="transition-colors hover:text-gray-100">Platform</a>
        </div>
        <div class="flex items-center gap-2.5">
          <template v-if="isAuthenticated">
            <NuxtLink to="/dashboard" class="btn-primary">Open Dashboard</NuxtLink>
          </template>
          <template v-else>
            <NuxtLink to="/auth/login" class="btn-ghost hidden sm:inline-flex">Sign in</NuxtLink>
            <NuxtLink to="/auth/register" class="btn-primary">Get started</NuxtLink>
          </template>
        </div>
      </nav>
    </header>

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
            <NuxtLink :to="isAuthenticated ? '/alerts' : '/auth/login'" class="btn-ghost px-5 py-3 text-base">
              {{ isAuthenticated ? 'View alerts' : 'Sign in' }}
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
              <span class="font-mono text-[11px] text-gray-500">{{ clock }}</span>
            </div>
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
          <p class="app-gradient-text text-4xl font-bold tracking-tight">{{ s.value }}</p>
          <p class="mt-1 text-sm text-gray-400">{{ s.label }}</p>
        </div>
      </div>
    </section>

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
        <p>© {{ year }} Security Ops · Video Analytics Platform</p>
        <div class="flex items-center gap-5">
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
onMounted(() => {
  const tick = () => { clock.value = new Date().toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' }) }
  tick()
  timer = setInterval(tick, 1000)
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

const stats = [
  { value: '7+', label: 'Threat types detected' },
  { value: '<300ms', label: 'Inference latency' },
  { value: '24/7', label: 'Continuous monitoring' },
  { value: '∞', label: 'Cameras supported' },
]
</script>
