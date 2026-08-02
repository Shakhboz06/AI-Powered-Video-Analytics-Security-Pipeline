<template>
  <Teleport to="body">
    <Transition name="overlay">
      <div
        v-if="open"
        class="app-modal-overlay"
        @click.self="$emit('close')"
      >
        <Transition name="pop" appear>
          <div v-if="open" class="app-modal" :class="wide ? 'max-w-lg' : ''">
            <div v-if="title" class="mb-4">
              <h3 class="text-lg font-semibold text-gray-100">{{ title }}</h3>
              <p v-if="subtitle" class="mt-1 text-sm text-gray-500">{{ subtitle }}</p>
            </div>
            <slot />
            <div v-if="$slots.footer" class="mt-6 flex justify-end gap-2">
              <slot name="footer" />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{
  open: boolean
  title?: string
  subtitle?: string
  wide?: boolean
}>()

defineEmits<{
  close: []
}>()
</script>
