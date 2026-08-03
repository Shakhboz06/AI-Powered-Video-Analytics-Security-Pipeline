<template>
  <div class="skeleton" :class="rounded" :style="style" />
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  width?: string
  height?: string
  rounded?: string
}>(), {
  width: '100%',
  height: '1rem',
  rounded: 'rounded-md',
})

const style = computed(() => ({ width: props.width, height: props.height }))
</script>
