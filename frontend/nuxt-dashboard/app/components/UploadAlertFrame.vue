<template>
  <div class="relative">
    <img
      v-if="frameUrl"
      :src="frameUrl"
      alt="Alert capture frame"
      class="h-full w-full object-cover"
      loading="lazy"
    >
    <div v-else class="grid h-full w-full place-items-center text-gray-600">
      <svg v-if="loading" class="h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v3a5 5 0 00-5 5H4z" />
      </svg>
      <svg v-else class="h-6 w-6" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3.75 21h16.5A1.5 1.5 0 0021.75 19.5V4.5A1.5 1.5 0 0020.25 3H3.75A1.5 1.5 0 002.25 4.5v15A1.5 1.5 0 003.75 21z" />
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

// The worker uploads the capture frame to storage *after* the alert row is
// already queryable, so the first fetch can 404 on a frame that simply
// isn't there yet. Five attempts with mild backoff cover that race window;
// anything still missing afterwards is a permanent miss (evicted frame).
const MAX_ATTEMPTS = 5
const RETRY_BASE_DELAY_MS = 2000

let attempts = 0
let retryTimer: ReturnType<typeof setTimeout> | null = null

function clearRetry() {
  if (retryTimer) {
    clearTimeout(retryTimer)
    retryTimer = null
  }
}

async function fetchFrame(id: string) {
  // loading stays true across the whole retry window — the spinner reads as
  // "frame's coming". It only drops on the two terminal outcomes: success
  // or attempts exhausted (which falls through to the placeholder icon).
  loading.value = true
  try {
    const { signed_url } = await api.getUploadAlertImage(props.jobId, id)
    clearRetry()
    frameUrl.value = signed_url
    loading.value = false
  }
  catch {
    frameUrl.value = null
    attempts += 1
    if (attempts < MAX_ATTEMPTS) {
      retryTimer = setTimeout(() => {
        retryTimer = null
        void fetchFrame(id)
      }, RETRY_BASE_DELAY_MS * attempts)
    }
    else {
      loading.value = false
    }
  }
}

watch(
  () => props.alertId,
  (raw) => {
    // Reset must also kill any pending retry and zero the counter — an old
    // alert's scheduled retry must never clobber the new alert's frame.
    clearRetry()
    attempts = 0
    frameUrl.value = null
    loading.value = false
    if (!import.meta.client || !raw?.trim()) return
    void fetchFrame(raw.trim())
  },
  { immediate: true },
)

// Navigating away mid-retry: drop the pending timer so no fetch fires
// against a dead component.
onUnmounted(clearRetry)
</script>
