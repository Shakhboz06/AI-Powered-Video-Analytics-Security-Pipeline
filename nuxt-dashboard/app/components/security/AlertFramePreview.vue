<template>
  <div class="relative aspect-[4/3] w-full max-w-[640px]">
    <div
      v-if="loading"
      class="absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-black/40 backdrop-blur-[1px]"
    >
      <svg class="h-4 w-4 animate-spin text-teal-400" fill="none" viewBox="0 0 24 24" aria-hidden="true">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
      </svg>
    </div>

    <!-- Real alert frame only — never overlay bbox on a placeholder photo -->
    <BoundingBoxPreview
      v-if="frameUrl"
      :bound-box="boundBox"
      :color="color"
      :label="label"
      :src="frameUrl"
    />

    <!-- No alert_id: show box on 640×480 grid (no misleading camera photo) -->
    <BoundingBoxPreview
      v-else-if="!alertId && !loading && hasBox"
      :bound-box="boundBox"
      :color="color"
      :label="label"
    />

    <div
      v-else-if="alertId && unavailable && !loading"
      class="flex h-full min-h-[120px] items-center justify-center rounded-lg border border-white/[0.06] bg-gray-950/60 text-xs text-gray-500"
    >
      Frame unavailable
    </div>
  </div>
</template>

<script setup lang="ts">
import BoundingBoxPreview from '~/components/security/BoundingBoxPreview.vue'
import { useAlertFrameImage } from '~/composables/useAlertFrameImage'
import { hasValidBoundBox } from '~/composables/useBoundBox'

const props = defineProps<{
  alertId?: string | null
  camera: string
  boundBox: [number, number, number, number] | number[] | null | undefined
  color?: string
  label?: string
}>()

const { frameUrl, loading, unavailable } = useAlertFrameImage(() => props.alertId)
const hasBox = computed(() => hasValidBoundBox(props.boundBox))
</script>
