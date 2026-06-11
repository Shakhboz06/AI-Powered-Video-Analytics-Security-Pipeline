<template>
  <div class="app-card app-card-hover kpi-card group" :class="cardClass">
    <div class="flex items-center justify-between gap-2">
      <div class="flex items-center gap-2">
        <div v-if="iconPath" class="kpi-card-icon transition-transform duration-300 group-hover:scale-110" :class="iconWrapClass">
          <svg class="h-4 w-4" :class="accentClass" fill="none" stroke="currentColor" stroke-width="1.6" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" :d="iconPath" />
          </svg>
        </div>
        <p class="app-label">{{ label }}</p>
      </div>
      <div v-if="trend" class="flex items-center gap-1 text-xs font-medium" :class="trendClass">
        <svg v-if="trend === 'up'" class="h-3 w-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18L9 11.25l4.306 4.307a11.95 11.95 0 015.814-5.519l2.74-1.22" /><path stroke-linecap="round" stroke-linejoin="round" d="M18.75 7.5h3v3" /></svg>
        <svg v-else class="h-3 w-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 6L9 12.75l4.286-4.286a11.948 11.948 0 014.306 6.43l.776 2.898" /><path stroke-linecap="round" stroke-linejoin="round" d="M18.75 16.5h3v-3" /></svg>
        {{ trendLabel }}
      </div>
    </div>
    <p class="mt-2 text-2xl font-semibold tabular-nums text-gray-100">
      <AnimatedNumber v-if="typeof value === 'number'" :value="value" :decimals="decimals" />
      <span v-else>{{ value }}</span>
      <span v-if="suffix" class="text-base font-normal text-gray-500">{{ suffix }}</span>
    </p>
    <div v-if="spark?.length > 1" class="mt-2">
      <Sparkline :data="spark" :color="sparkColor" />
    </div>
    <p v-if="hint" class="mt-1 text-xs text-gray-500">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import AnimatedNumber from '~/components/ui/AnimatedNumber.vue'
import Sparkline from '~/components/ui/Sparkline.vue'

const props = withDefaults(defineProps<{
  label: string
  value: number | string
  suffix?: string
  hint?: string
  decimals?: number
  iconPath?: string
  iconWrapClass?: string
  accentClass?: string
  cardClass?: string
  spark?: number[]
  sparkColor?: string
  trend?: 'up' | 'down' | null
  trendLabel?: string
}>(), {
  decimals: 0,
  iconWrapClass: 'border-teal-500/20 bg-teal-950/30',
  accentClass: 'text-teal-400',
  sparkColor: '#2dd4bf',
  spark: () => [],
  trend: null,
  trendLabel: '',
})

const trendClass = computed(() => {
  if (props.trend === 'up') return 'text-emerald-400'
  if (props.trend === 'down') return 'text-red-400'
  return 'text-gray-400'
})
</script>
