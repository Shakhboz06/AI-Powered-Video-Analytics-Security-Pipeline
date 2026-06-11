<template>
  <NuxtLink v-if="to" :to="to" class="group inline-flex items-center gap-2.5">
    <span
      class="relative grid place-items-center rounded-xl text-[#04201c] shadow-lg transition-all duration-300 group-hover:scale-105 group-hover:shadow-[0_0_24px_-4px_rgba(45,212,191,0.4)]"
      :class="markClass"
      style="background-image: linear-gradient(120deg, #2dd4bf, #22d3ee)"
    >
      <svg :class="iconClass" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path d="M12 3L4 7v10l8 4 8-4V7l-8-4z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" />
        <path d="M12 12l8-4M12 12v10M12 12L4 8" stroke="currentColor" stroke-width="1.6" />
      </svg>
    </span>
    <span v-if="showText" class="flex flex-col leading-tight">
      <span class="app-gradient-text font-bold tracking-tight" :class="textClass">Security Ops</span>
      <span v-if="showTagline" class="font-medium uppercase tracking-[0.18em] text-gray-500" :class="taglineClass">Video Analytics</span>
    </span>
  </NuxtLink>
  <div v-else class="group inline-flex items-center gap-2.5">
    <span
      class="relative grid place-items-center rounded-xl text-[#04201c] shadow-lg transition-all duration-300 group-hover:scale-105 group-hover:shadow-[0_0_24px_-4px_rgba(45,212,191,0.4)]"
      :class="markClass"
      style="background-image: linear-gradient(120deg, #2dd4bf, #22d3ee)"
    >
      <svg :class="iconClass" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path d="M12 3L4 7v10l8 4 8-4V7l-8-4z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" />
        <path d="M12 12l8-4M12 12v10M12 12L4 8" stroke="currentColor" stroke-width="1.6" />
      </svg>
    </span>
    <span v-if="showText" class="flex flex-col leading-tight">
      <span class="app-gradient-text font-bold tracking-tight" :class="textClass">Security Ops</span>
      <span v-if="showTagline" class="font-medium uppercase tracking-[0.18em] text-gray-500" :class="taglineClass">Video Analytics</span>
    </span>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  to?: string
  size?: 'sm' | 'md' | 'lg'
  showText?: boolean
  showTagline?: boolean
}>(), {
  size: 'md',
  showText: true,
  showTagline: false,
})

const markClass = computed(() => {
  if (props.size === 'lg') return 'h-14 w-14 rounded-2xl'
  if (props.size === 'sm') return 'h-8 w-8'
  return 'h-9 w-9'
})
const iconClass = computed(() => {
  if (props.size === 'lg') return 'h-7 w-7'
  if (props.size === 'sm') return 'h-4 w-4'
  return 'h-5 w-5'
})
const textClass = computed(() => {
  if (props.size === 'lg') return 'text-2xl'
  if (props.size === 'sm') return 'text-sm'
  return 'text-base'
})
const taglineClass = computed(() => (props.size === 'lg' ? 'text-xs' : 'text-[10px]'))
</script>
