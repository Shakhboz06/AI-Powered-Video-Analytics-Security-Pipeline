<template>
  <div class="app-card relative flex flex-col gap-4 overflow-hidden p-4 md:p-5 lg:flex-row lg:flex-wrap lg:items-end">
    <div class="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-teal-500/30 to-transparent" />
    <div class="pointer-events-none absolute inset-y-0 left-0 w-px bg-gradient-to-b from-transparent via-teal-500/20 to-transparent" />
    <div class="flex flex-1 flex-wrap items-end gap-4">
      <slot />
    </div>
    <div v-if="$slots.actions" class="flex flex-wrap items-center gap-3 lg:ml-auto">
      <slot name="actions" />
    </div>
  </div>
</template>
