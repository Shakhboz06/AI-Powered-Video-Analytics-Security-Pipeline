<template>
  <div class="app-shell-bg app-grid relative min-h-screen overflow-x-hidden font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh />

    <header class="sticky top-0 z-30 nav-glass">
      <nav class="mx-auto flex h-16 max-w-7xl items-center justify-between px-5 md:px-8">
        <AppBrand to="/" :show-tagline="false" />
        <div class="flex items-center gap-2.5">
          <NuxtLink to="/auth/login" class="btn-ghost hidden sm:inline-flex">Sign in</NuxtLink>
          <NuxtLink to="/" class="btn-ghost">Home</NuxtLink>
        </div>
      </nav>
    </header>

    <section class="relative z-10 mx-auto max-w-3xl px-5 pb-20 pt-12 md:px-8 md:pt-16">
      <div class="text-center animate-fade-up">
        <span class="app-chip border-teal-500/30 bg-teal-950/30 text-teal-300">
          <span class="live-dot h-1.5 w-1.5 rounded-full bg-teal-400" />
          Free anomaly scan · no account needed
        </span>
        <h1 class="mt-5 text-3xl font-bold tracking-tight text-gray-50 md:text-4xl">
          Analyze a video for <span class="app-gradient-text">security anomalies</span>
        </h1>
        <p class="mx-auto mt-3 max-w-xl text-gray-400">
          Upload a clip and our detection pipeline will scan it for falls, fights, weapons,
          running and abandoned objects. You'll get a shareable link with the results.
        </p>
      </div>

      <div class="mt-10 app-card p-6 md:p-8">
        <!-- Dropzone -->
        <div
          v-if="!uploading"
          class="group relative grid cursor-pointer place-items-center rounded-2xl border-2 border-dashed px-6 py-14 text-center transition-colors"
          :class="dragging ? 'border-teal-400/70 bg-teal-950/20' : 'border-white/15 hover:border-teal-500/40 hover:bg-white/[0.02]'"
          @click="fileInput?.click()"
          @dragover.prevent="dragging = true"
          @dragleave.prevent="dragging = false"
          @drop.prevent="onDrop"
        >
          <input
            ref="fileInput"
            type="file"
            class="hidden"
            :accept="ACCEPT"
            @change="onPick"
          >
          <div class="pointer-events-none">
            <span class="mx-auto grid h-14 w-14 place-items-center rounded-2xl border border-teal-500/30 bg-teal-950/30 text-teal-300 transition-transform duration-300 group-hover:scale-110">
              <svg class="h-7 w-7" fill="none" stroke="currentColor" stroke-width="1.6" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
              </svg>
            </span>
            <p class="mt-4 text-base font-semibold text-gray-100">
              {{ selectedFile ? selectedFile.name : 'Drop your video here' }}
            </p>
            <p class="mt-1 text-sm text-gray-500">
              <template v-if="selectedFile">
                {{ prettySize(selectedFile.size) }} — click to choose a different file
              </template>
              <template v-else>
                or click to browse · MP4, MOV, AVI, MKV, WebM · up to {{ MAX_MB }} MB
              </template>
            </p>
          </div>
        </div>

        <!-- Upload progress -->
        <div v-else class="px-2 py-10 text-center">
          <p class="text-base font-semibold text-gray-100">Uploading {{ selectedFile?.name }}…</p>
          <div class="mx-auto mt-5 h-2 max-w-md overflow-hidden rounded-full bg-white/10">
            <div
              class="h-full rounded-full bg-gradient-to-r from-teal-400 to-emerald-400 transition-all duration-200"
              :style="{ width: `${uploadPercent}%` }"
            />
          </div>
          <p class="mt-2 text-sm tabular-nums text-gray-400">{{ uploadPercent }}%</p>
        </div>

        <p v-if="errorMsg" class="mt-4 rounded-xl border border-red-800/60 bg-red-950/40 px-4 py-3 text-sm text-red-200">
          {{ errorMsg }}
        </p>

        <div class="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
          <p class="text-xs text-gray-500">
            Only the first {{ MAX_ANALYZED_SECONDS }}s are analyzed. Results stay available via a private link.
          </p>
          <button
            class="btn-primary w-full px-6 py-3 text-base sm:w-auto"
            :disabled="!selectedFile || uploading"
            :class="(!selectedFile || uploading) ? 'cursor-not-allowed opacity-50' : ''"
            @click="startUpload"
          >
            <span v-if="uploading" class="live-dot h-2 w-2 rounded-full bg-white" />
            {{ uploading ? 'Uploading…' : 'Analyze video' }}
          </button>
        </div>
      </div>

      <div class="stagger-in mt-10 grid gap-4 sm:grid-cols-3">
        <div v-for="step in steps" :key="step.title" class="app-card p-5 text-center">
          <span class="app-gradient-text text-xl font-bold">{{ step.num }}</span>
          <h3 class="mt-2 text-sm font-semibold text-gray-100">{{ step.title }}</h3>
          <p class="mt-1 text-sm text-gray-500">{{ step.desc }}</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import AppBrand from '~/components/ui/AppBrand.vue'
import AmbientMesh from '~/components/ui/AmbientMesh.vue'

definePageMeta({ layout: false })

useHead({ title: 'Analyze a video — Security Ops' })

const ACCEPT = '.mp4,.mov,.avi,.mkv,.webm,video/*'
const ALLOWED_EXTS = ['mp4', 'mov', 'avi', 'mkv', 'webm']
const MAX_MB = 200
const MAX_ANALYZED_SECONDS = 180

const api = useApi()

const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const dragging = ref(false)
const uploading = ref(false)
const uploadPercent = ref(0)
const errorMsg = ref('')

const steps = [
  { num: '01', title: 'Upload', desc: 'Your clip is queued for the detection pipeline.' },
  { num: '02', title: 'Detect', desc: 'YOLO, pose and video models scan every frame.' },
  { num: '03', title: 'Share', desc: 'Anomalies appear on a link anyone can open.' },
]

function validate(file: File): string {
  const ext = file.name.split('.').pop()?.toLowerCase() ?? ''
  if (!ALLOWED_EXTS.includes(ext))
    return `Unsupported file type ".${ext}" — use MP4, MOV, AVI, MKV or WebM.`
  if (file.size > MAX_MB * 1024 * 1024)
    return `File is ${prettySize(file.size)}, the limit is ${MAX_MB} MB.`
  return ''
}

function setFile(file: File | undefined | null) {
  errorMsg.value = ''
  if (!file) return
  const problem = validate(file)
  if (problem) {
    errorMsg.value = problem
    selectedFile.value = null
    return
  }
  selectedFile.value = file
}

function onPick(e: Event) {
  setFile((e.target as HTMLInputElement).files?.[0])
}

function onDrop(e: DragEvent) {
  dragging.value = false
  setFile(e.dataTransfer?.files?.[0])
}

async function startUpload() {
  if (!selectedFile.value || uploading.value) return
  errorMsg.value = ''
  uploading.value = true
  uploadPercent.value = 0

  try {
    const job = await api.uploadVideo(selectedFile.value, (p) => { uploadPercent.value = p })
    await navigateTo(`/results/${job.job_id}`)
  }
  catch (err) {
    errorMsg.value = err instanceof Error ? err.message : 'Upload failed, please try again.'
    uploading.value = false
  }
}

function prettySize(bytes: number): string {
  if (bytes >= 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${Math.ceil(bytes / 1024)} KB`
}
</script>
