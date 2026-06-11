<template>
  <div class="kpi-card" :class="cardClass">
    <div class="flex items-center gap-2">
      <div v-if="iconPath" class="kpi-card-icon" :class="iconWrapClass">
        <svg class="h-4 w-4" :class="accentClass" fill="none" stroke="currentColor" stroke-width="1.6" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" :d="iconPath" />
        </svg>
      </div>
      <p class="app-label">{{ label }}</p>
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

withDefaults(defineProps<{
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
}>(), {
  decimals: 0,
  iconWrapClass: 'border-teal-500/20 bg-teal-950/30',
  accentClass: 'text-teal-400',
  sparkColor: '#2dd4bf',
  spark: () => [],
})
</script>
