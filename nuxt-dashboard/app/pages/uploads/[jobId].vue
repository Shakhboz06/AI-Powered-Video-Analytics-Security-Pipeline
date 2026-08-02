<template>
  <div class="app-shell-bg app-grid relative min-h-screen overflow-x-hidden font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh />

    <header class="sticky top-0 z-30 nav-glass">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 md:px-8">
        <AppBrand to="/" :show-tagline="false" />
        <div class="flex items-center gap-2.5">
          <NuxtLink to="/upload" class="btn-primary">Analyze another video</NuxtLink>
        </div>
      </nav>
    </header>

    <section class="relative z-10 mx-auto max-w-4xl px-5 pb-20 pt-10 md:px-8">
      <!-- Not found -->
      <div v-if="notFound" class="app-card mt-10 p-10 text-center">
        <h1 class="text-2xl font-bold text-gray-50">Analysis not found</h1>
        <p class="mt-2 text-gray-400">This link doesn't match any video analysis. It may have been mistyped.</p>
        <NuxtLink to="/upload" class="btn-primary mt-6 inline-flex">Analyze a video</NuxtLink>
      </div>

      <template v-else-if="job">
        <!-- Header -->
        <div class="animate-fade-up">
          <p class="page-eyebrow">Video analysis</p>
          <div class="mt-2 flex flex-wrap items-center gap-3">
            <h1 class="max-w-full truncate text-2xl font-bold tracking-tight text-gray-50 md:text-3xl">
              {{ job.filename || 'Uploaded video' }}
            </h1>
            <span
              class="inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-semibold"
              :class="statusMeta.chip"
            >
              <span v-if="isRunning" class="live-dot h-1.5 w-1.5 rounded-full" :class="statusMeta.dot" />
              {{ statusMeta.label }}
            </span>
          </div>
          <p class="mt-1.5 text-sm text-gray-500">
            Uploaded {{ formatDateTime(job.created_at) }}
          </p>
        </div>

        <!-- Status banner + share link -->
        <div class="app-card mt-6 p-5 md:p-6">
          <div v-if="isRunning" class="flex items-center gap-3 text-sm text-gray-300">
            <svg class="h-5 w-5 animate-spin text-teal-300" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v3a5 5 0 00-5 5H4z" />
            </svg>
            <span>
              {{ job.status === 'queued' ? 'Queued — waiting for an analysis slot…' : 'Processing your video…' }}
              <span class="text-gray-500">Detections appear below as they're found.</span>
            </span>
          </div>

          <div v-else-if="job.status === 'failed'" class="rounded-xl border border-red-800/60 bg-red-950/40 px-4 py-3 text-sm text-red-200">
            Analysis failed — the video couldn't be processed. Try uploading it again.
          </div>

          <div v-else class="flex items-center gap-2 text-sm text-gray-300">
            <svg class="h-5 w-5 text-emerald-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            Done — {{ alerts.length }} {{ alerts.length === 1 ? 'anomaly' : 'anomalies' }} found.
          </div>

          <!-- Share link -->
          <div class="mt-5 flex flex-col gap-2 sm:flex-row sm:items-center">
            <div class="flex-1 truncate rounded-xl border border-white/10 bg-white/[0.03] px-4 py-2.5 font-mono text-xs text-gray-400">
              {{ shareUrl }}
            </div>
            <button class="btn-ghost shrink-0 justify-center" @click="copyLink">
              <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244" /></svg>
              {{ copied ? 'Copied!' : 'Copy link' }}
            </button>
          </div>
        </div>

        <!-- Anomalies — reuses the platform's alert cards -->
        <div class="mt-6 space-y-3">
          <div v-if="!alerts.length && isRunning" class="app-card p-8 text-center text-sm text-gray-500">
            Watching for anomalies…
          </div>

          <div v-else-if="!alerts.length && job.status === 'done'" class="app-card p-10 text-center">
            <span class="mx-auto grid h-14 w-14 place-items-center rounded-2xl border border-emerald-500/30 bg-emerald-950/30 text-emerald-300">
              <svg class="h-7 w-7" fill="none" stroke="currentColor" stroke-width="1.6" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" /></svg>
            </span>
            <h2 class="mt-4 text-lg font-semibold text-gray-100">Done — no anomalies detected</h2>
            <p class="mt-1 text-sm text-gray-500">The pipeline didn't flag any falls, fights, weapons, running or abandoned objects.</p>
          </div>

          <AlertCard
            v-for="alert in readOnlyAlerts"
            :key="alert.alert_id ?? `${alert.recorded_at}-${alert.tracker_id}`"
            :alert="alert"
            :busy-id="null"
          />
        </div>
      </template>

      <!-- Initial loading -->
      <div v-else class="app-card mt-10 p-10 text-center text-sm text-gray-500">
        Loading analysis…
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import AppBrand from '~/components/ui/AppBrand.vue'
import AmbientMesh from '~/components/ui/AmbientMesh.vue'
import AlertCard from '~/components/AlertCard.vue'
import type { SecurityAlert, UploadJob } from '~/types/security'

definePageMeta({ layout: false })

const route = useRoute()
const api = useApi()

const jobId = computed(() => String(route.params.jobId ?? ''))

const job = ref<UploadJob | null>(null)
const alerts = ref<SecurityAlert[]>([])
const notFound = ref(false)
const copied = ref(false)

useHead(() => ({ title: job.value ? `Results — ${job.value.filename}` : 'Video analysis results' }))

const isRunning = computed(() => !!job.value && ['queued', 'processing'].includes(job.value.status))

/** Negative ids switch AlertCard into its read-only mode (no ack/resolve). */
const readOnlyAlerts = computed(() =>
  alerts.value.map((a, i) => ({ ...a, id: -(i + 1) })),
)

const statusMeta = computed(() => {
  switch (job.value?.status) {
    case 'queued':
      return { label: 'Queued', chip: 'border-gray-600 bg-gray-900/60 text-gray-300', dot: 'bg-gray-400' }
    case 'processing':
      return { label: 'Processing', chip: 'border-teal-700/60 bg-teal-950/40 text-teal-200', dot: 'bg-teal-400' }
    case 'failed':
      return { label: 'Failed', chip: 'border-red-800/60 bg-red-950/40 text-red-200', dot: 'bg-red-400' }
    default:
      return { label: 'Done', chip: 'border-emerald-700/60 bg-emerald-950/40 text-emerald-200', dot: 'bg-emerald-400' }
  }
})

const shareUrl = computed(() =>
  import.meta.client ? `${window.location.origin}/uploads/${jobId.value}` : `/uploads/${jobId.value}`,
)

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
  catch { /* clipboard unavailable */ }
}

async function fetchStatus() {
  try {
    const res = await $fetch<{ job: UploadJob }>(`/api/uploads/${encodeURIComponent(jobId.value)}`)
    job.value = res.job
    notFound.value = false
  }
  catch (err: unknown) {
    const status = (err as { statusCode?: number })?.statusCode ?? (err as { status?: number })?.status
    if (status === 404 || status === 400) notFound.value = true
  }
}

async function fetchAlerts() {
  try {
    const res = await api.listAlerts({ camera: jobId.value })
    alerts.value = res.alerts ?? []
  }
  catch { /* transient — next poll retries */ }
}

let pollTimer: ReturnType<typeof setTimeout> | null = null

async function poll() {
  await Promise.all([fetchStatus(), fetchAlerts()])

  if (notFound.value) return

  if (job.value && ['done', 'failed'].includes(job.value.status)) {
    // one final alerts fetch to catch stragglers still landing in the DB
    setTimeout(fetchAlerts, 3000)
    return
  }

  pollTimer = setTimeout(poll, 3000)
}

onMounted(poll)
onUnmounted(() => { if (pollTimer) clearTimeout(pollTimer) })
</script>
