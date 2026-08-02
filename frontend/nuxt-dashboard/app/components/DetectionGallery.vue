<template>
  <section class="relative z-10 mx-auto max-w-7xl px-5 py-12 md:px-8">
    <div class="text-center">
      <p class="page-eyebrow">Evidence</p>
      <h2 class="mt-2 text-3xl font-bold tracking-tight text-gray-50">
        Detected in <span class="app-gradient-text">real footage</span>
      </h2>
      <p class="mx-auto mt-3 max-w-2xl text-gray-400">
        Frames captured by the pipeline the moment a threat was flagged — weapons, fights,
        falls and abandoned objects, each stamped with its severity and confidence.
      </p>
    </div>

    <div class="stagger-in mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <article
        v-for="(card, i) in cards"
        :key="`${card.image}-${i}`"
        class="app-card app-card-hover group overflow-hidden p-0"
      >
        <!-- Evidence frame -->
        <div class="relative aspect-video w-full overflow-hidden bg-[#05080c]">
          <img
            :src="card.image"
            :alt="`${card.label} detection frame`"
            class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105"
            loading="lazy"
          >
          <div class="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent" />

          <!-- Label badge (type-coloured, reuses alert meta) -->
          <span
            class="absolute left-3 top-3 inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-xs font-semibold backdrop-blur-sm"
            :class="meta(card).badgeClass"
          >
            <svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path v-for="(d, j) in meta(card).iconPaths" :key="j" :d="d" />
            </svg>
            {{ card.label }}
          </span>

          <!-- Severity chip (colour-coded) -->
          <span
            class="absolute right-3 top-3 inline-flex items-center gap-1.5 rounded-md border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide backdrop-blur-sm"
            :class="severityChip(card.severity)"
          >
            <span class="h-1.5 w-1.5 rounded-full" :class="severityDot(card.severity)" />
            {{ card.severity }}
          </span>

          <!-- Footer: timestamp + confidence -->
          <div class="absolute inset-x-0 bottom-0 flex items-center justify-between px-3 py-2.5">
            <span class="inline-flex items-center gap-1.5 font-mono text-xs text-gray-200">
              <svg class="h-3.5 w-3.5 text-gray-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              {{ card.timestamp }}
            </span>
            <span class="rounded bg-black/50 px-2 py-0.5 font-mono text-xs text-teal-300 backdrop-blur-sm">
              {{ Math.round(card.confidence * 100) }}%
            </span>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { getAlertMeta } from '~/composables/useAlertMeta'
import { detectionCards, type DetectionCard } from '~/data/detections'

// Data-driven: defaults to the editable list in ~/data/detections.ts, but a
// parent can override via the `cards` prop.
const props = withDefaults(defineProps<{ cards?: DetectionCard[] }>(), {
  cards: () => detectionCards,
})

const cards = computed(() => props.cards)

/** Reuse the platform's per-type badge styling (icon + colour). */
function meta(card: DetectionCard) {
  return getAlertMeta(card.type)
}

/** Severity colour coding — red = critical, amber = medium (per design tokens). */
function severityChip(severity: DetectionCard['severity']): string {
  switch (severity) {
    case 'critical': return 'border-red-700/60 bg-red-950/60 text-red-200'
    case 'high': return 'border-orange-700/60 bg-orange-950/60 text-orange-200'
    case 'medium': return 'border-amber-700/60 bg-amber-950/60 text-amber-200'
    default: return 'border-sky-700/60 bg-sky-950/60 text-sky-200'
  }
}

function severityDot(severity: DetectionCard['severity']): string {
  switch (severity) {
    case 'critical': return 'bg-red-400'
    case 'high': return 'bg-orange-400'
    case 'medium': return 'bg-amber-400'
    default: return 'bg-sky-400'
  }
}
</script>
