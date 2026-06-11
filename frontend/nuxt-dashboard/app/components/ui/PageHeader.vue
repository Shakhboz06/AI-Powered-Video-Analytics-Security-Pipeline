<template>
  <header class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
    <div class="min-w-0">
      <p v-if="eyebrow" class="page-eyebrow flex items-center gap-2">
        <span class="inline-block h-1.5 w-1.5 rounded-full bg-teal-400" style="box-shadow: 0 0 8px rgba(45, 212, 191, 0.5)" />
        {{ eyebrow }}
      </p>
      <h1 class="page-title">{{ title }}</h1>
      <div class="page-header-accent glow-pulse" />
      <p v-if="subtitle" class="page-subtitle max-w-2xl">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="flex shrink-0 flex-wrap items-center gap-2">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{
  eyebrow?: string
  title: string
  subtitle?: string
}>()
</script>
