<template>
  <div class="ambient-mesh pointer-events-none fixed inset-0 z-0 overflow-hidden" aria-hidden="true">
    <div class="ai-orb -left-[20%] -top-[15%] h-[36rem] w-[36rem] bg-teal-500/14" />
    <div class="ai-orb -right-[15%] top-[10%] h-[32rem] w-[32rem] bg-indigo-500/12" style="animation-delay: -7s" />
    <div class="ai-orb bottom-[-10%] left-[30%] h-[28rem] w-[28rem] bg-cyan-500/10" style="animation-delay: -14s" />
    <div class="ai-orb bottom-[20%] right-[25%] h-64 w-64 bg-violet-500/8" style="animation-delay: -20s" />
    <div class="ambient-noise" />
    <div v-if="particles" class="ambient-particles" />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  particles?: boolean
}>(), {
  particles: true,
})
</script>
