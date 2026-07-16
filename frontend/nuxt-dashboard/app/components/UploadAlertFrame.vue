<template>
  <div class="relative">
    <img v-if="frameUrl" :src="frameUrl" alt="Alert capture frame" class="h-full w-full object-cover" loading="lazy">
    <div v-else class="grid h-full w-full place-items-center text-gray-600">
      <svg v-if="loading" class="h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v3a5 5 0 00-5 5H4z" />
      </svg>
      <svg v-else class="h-6 w-6" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round"
          d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3.75 21h16.5A1.5 1.5 0 0021.75 19.5V4.5A1.5 1.5 0 0020.25 3H3.75A1.5 1.5 0 002.25 4.5v15A1.5 1.5 0 003.75 21z" />
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  jobId: string
  alertId: string | null
}>()

const api = useApi()

const frameUrl = ref<string | null>(null)
const loading = ref(false)

const max_attempts = 5
const retry_delay_ms = 3000
let attempts = 0
let retryTimer: ReturnType<typeof setTimeout> | null = null

async function fetchFrame(id: string) {
  loading.value = true
  try {
    const { signed_url } = await api.getUploadAlertImage(props.jobId, id)
    frameUrl.value = signed_url
    loading.value = false          
  }
  catch {
    attempts++
    if (attempts < max_attempts) {
      retryTimer = setTimeout(() => fetchFrame(id), retry_delay_ms)
    }
    else {
      loading.value = false       
    }
  }
}

watch(
  () => props.alertId,
  (id) => {
    if (retryTimer) { clearTimeout(retryTimer); retryTimer = null }
    attempts = 0
    frameUrl.value = null
    if (!import.meta.client || !id?.trim()) return
    fetchFrame(id.trim())
  },
  { immediate: true },
)

onUnmounted(() => {
  if (retryTimer) clearTimeout(retryTimer)
})

</script>
