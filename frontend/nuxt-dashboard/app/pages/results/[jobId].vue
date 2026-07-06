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
              {{ job.original_filename }}
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

        <!-- Progress / share bar -->
        <div class="app-card mt-6 p-5 md:p-6">
          <div v-if="isRunning">
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-300">{{ statusMeta.hint }}</span>
              <span class="tabular-nums text-gray-400">{{ job.progress }}%</span>
            </div>
            <div class="mt-3 h-2 overflow-hidden rounded-full bg-white/10">
              <div
                class="h-full rounded-full bg-gradient-to-r from-teal-400 to-emerald-400 transition-all duration-500"
                :style="{ width: `${Math.max(job.progress, 3)}%` }"
              />
            </div>
            <p class="mt-2 text-xs text-gray-500">This page updates automatically — you can already share the link below.</p>
          </div>

          <div v-else-if="job.status === 'failed'" class="rounded-xl border border-red-800/60 bg-red-950/40 px-4 py-3 text-sm text-red-200">
            Analysis failed{{ job.error ? `: ${job.error}` : '.' }} Try uploading the video again.
          </div>

          <div v-else class="flex items-center gap-2 text-sm text-gray-300">
            <svg class="h-5 w-5 text-emerald-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            Analysis complete — {{ alerts.length }} {{ alerts.length === 1 ? 'anomaly' : 'anomalies' }} detected.
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

        <!-- Summary by type -->
        <div v-if="typeSummary.length" class="mt-6 flex flex-wrap gap-2">
          <span
            v-for="entry in typeSummary"
            :key="entry.meta.key"
            class="inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium"
            :class="entry.meta.badgeClass"
          >
            <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
              <path v-for="(d, i) in entry.meta.iconPaths" :key="i" stroke-linecap="round" stroke-linejoin="round" :d="d" />
            </svg>
            {{ entry.meta.label }} × {{ entry.count }}
          </span>
        </div>

        <!-- Alerts -->
        <div class="mt-6 space-y-3">
          <div v-if="!alerts.length && isRunning" class="app-card p-8 text-center text-sm text-gray-500">
            Watching for anomalies… detections appear here in real time.
          </div>

          <div v-else-if="!alerts.length && job.status === 'done'" class="app-card p-10 text-center">
            <span class="mx-auto grid h-14 w-14 place-items-center rounded-2xl border border-emerald-500/30 bg-emerald-950/30 text-emerald-300">
              <svg class="h-7 w-7" fill="none" stroke="currentColor" stroke-width="1.6" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" /></svg>
            </span>
            <h2 class="mt-4 text-lg font-semibold text-gray-100">No anomalies detected</h2>
            <p class="mt-1 text-sm text-gray-500">The pipeline didn't flag any falls, fights, weapons, running or abandoned objects.</p>
          </div>

          <article
            v-for="alert in alerts"
            :key="alert.id"
            class="app-card relative overflow-hidden p-0"
          >
            <span class="absolute inset-y-0 left-0 w-1" :class="metaFor(alert).stripeClass" />
            <div class="flex flex-col gap-4 p-4 pl-5 sm:flex-row sm:items-center">
              <!-- Frame thumbnail -->
              <UploadAlertFrame
                :job-id="jobId"
                :alert-id="alert.alert_id ?? null"
                class="h-28 w-full shrink-0 overflow-hidden rounded-xl border border-white/10 bg-black/40 sm:w-44"
              />

              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-xs font-semibold" :class="metaFor(alert).badgeClass">
                    <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24">
                      <path v-for="(d, i) in metaFor(alert).iconPaths" :key="i" stroke-linecap="round" stroke-linejoin="round" :d="d" />
                    </svg>
                    {{ metaFor(alert).label }}
                  </span>
                  <span
                    v-if="alert.severity"
                    class="rounded-full border px-2.5 py-0.5 text-xs font-medium capitalize"
                    :class="severityChip(alert.severity)"
                  >
                    {{ alert.severity }}
                  </span>
                </div>
                <p class="mt-2 text-sm text-gray-300">
                  <span class="capitalize">{{ alert.label }}</span>
                  <template v-if="alert.zone_name"> in zone “{{ alert.zone_name }}”</template>
                  <template v-if="alert.tracker_id > 0"> · subject #{{ alert.tracker_id }}</template>
                </p>
                <p class="mt-1 font-mono text-xs text-gray-500">
                  at {{ offsetInVideo(alert.recorded_at) }} in the video
                </p>
              </div>
            </div>
          </article>
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
import UploadAlertFrame from '~/components/UploadAlertFrame.vue'
import { getAlertMeta } from '~/composables/useAlertMeta'
import type { AnalysisJob, SecurityAlert } from '~/types/security'

definePageMeta({ layout: false })

const route = useRoute()
const api = useApi()

const jobId = computed(() => String(route.params.jobId ?? ''))

const job = ref<AnalysisJob | null>(null)
const alerts = ref<SecurityAlert[]>([])
const notFound = ref(false)
const copied = ref(false)

useHead(() => ({ title: job.value ? `Results — ${job.value.original_filename}` : 'Video analysis results' }))

const isRunning = computed(() => !!job.value && ['queued', 'processing', 'finalizing'].includes(job.value.status))

const statusMeta = computed(() => {
  switch (job.value?.status) {
    case 'queued':
      return { label: 'Queued', hint: 'Waiting for an analysis slot…', chip: 'border-gray-600 bg-gray-900/60 text-gray-300', dot: 'bg-gray-400' }
    case 'processing':
      return { label: 'Analyzing', hint: 'Scanning frames for anomalies…', chip: 'border-teal-700/60 bg-teal-950/40 text-teal-200', dot: 'bg-teal-400' }
    case 'finalizing':
      return { label: 'Finalizing', hint: 'Wrapping up the last detections…', chip: 'border-sky-700/60 bg-sky-950/40 text-sky-200', dot: 'bg-sky-400' }
    case 'failed':
      return { label: 'Failed', hint: '', chip: 'border-red-800/60 bg-red-950/40 text-red-200', dot: 'bg-red-400' }
    default:
      return { label: 'Complete', hint: '', chip: 'border-emerald-700/60 bg-emerald-950/40 text-emerald-200', dot: 'bg-emerald-400' }
  }
})

const shareUrl = computed(() =>
  import.meta.client ? `${window.location.origin}/results/${jobId.value}` : `/results/${jobId.value}`,
)

const typeSummary = computed(() => {
  const counts = new Map<string, { meta: ReturnType<typeof getAlertMeta>, count: number }>()
  for (const a of alerts.value) {
    const meta = getAlertMeta(a.alert_type)
    const entry = counts.get(meta.key)
    if (entry) entry.count++
    else counts.set(meta.key, { meta, count: 1 })
  }
  return [...counts.values()].sort((x, y) => y.count - x.count)
})

function metaFor(alert: SecurityAlert) {
  return getAlertMeta(alert.alert_type)
}

function severityChip(severity: string): string {
  switch (severity.toLowerCase()) {
    case 'high': return 'border-red-800/60 bg-red-950/40 text-red-200'
    case 'medium': return 'border-amber-700/60 bg-amber-950/40 text-amber-200'
    default: return 'border-sky-800/60 bg-sky-950/40 text-sky-200'
  }
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}

/** Frame timestamps are anchored at the moment analysis started, so the delta is the position inside the video. */
function offsetInVideo(recordedAt: string): string {
  const anchor = job.value?.started_at ?? job.value?.created_at
  if (!anchor) return formatDateTime(recordedAt)
  const seconds = Math.max(0, Math.round((new Date(recordedAt).getTime() - new Date(anchor).getTime()) / 1000))
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${String(s).padStart(2, '0')}`
}

async function copyLink() {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
  catch { /* clipboard unavailable */ }
}

let pollTimer: ReturnType<typeof setTimeout> | null = null

async function refresh() {
  try {
    const res = await api.getUploadJob(jobId.value)
    job.value = res.job
    alerts.value = res.alerts ?? []
    notFound.value = false
  }
  catch (err: unknown) {
    const status = (err as { statusCode?: number, status?: number })?.statusCode ?? (err as { status?: number })?.status
    if (status === 404 || status === 400) {
      notFound.value = true
      return
    }
    // transient error — keep polling
  }

  if (!notFound.value && (!job.value || ['queued', 'processing', 'finalizing'].includes(job.value.status)))
    pollTimer = setTimeout(refresh, 2500)
}

onMounted(refresh)
onUnmounted(() => { if (pollTimer) clearTimeout(pollTimer) })
</script>
