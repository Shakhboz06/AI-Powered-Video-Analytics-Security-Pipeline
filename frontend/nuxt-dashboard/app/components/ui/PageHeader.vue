<template>
  <header class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
    <div class="min-w-0">
      <p v-if="eyebrow" class="page-eyebrow">{{ eyebrow }}</p>
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
