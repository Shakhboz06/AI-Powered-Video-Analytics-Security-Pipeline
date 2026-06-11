<template>
  <Teleport to="body">
    <div class="pointer-events-none fixed bottom-4 right-4 z-[100] flex w-[calc(100vw-2rem)] max-w-sm flex-col gap-2.5">
      <TransitionGroup name="toast">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="pointer-events-auto flex items-start gap-3 rounded-xl border p-3.5 shadow-2xl backdrop-blur-xl transition-transform duration-200 hover:scale-[1.02]"
          :class="styles[t.type].card"
          role="status"
        >
          <span class="mt-0.5 grid h-6 w-6 shrink-0 place-items-center rounded-full" :class="styles[t.type].iconWrap">
            <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" :d="styles[t.type].icon" />
            </svg>
          </span>
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-gray-100">{{ t.message }}</p>
            <p v-if="t.description" class="mt-0.5 text-xs text-gray-400">{{ t.description }}</p>
          </div>
          <button
            type="button"
            class="shrink-0 rounded-md p-1 text-gray-500 transition-colors hover:bg-white/5 hover:text-gray-300"
            aria-label="Dismiss"
            @click="dismiss(t.id)"
          >
            <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { ToastType } from '~/composables/useToast'

const { toasts, dismiss } = useToast()

const styles: Record<ToastType, { card: string; iconWrap: string; icon: string }> = {
  success: {
    card: 'border-emerald-500/30 bg-emerald-950/70',
    iconWrap: 'bg-emerald-500/15 text-emerald-300',
    icon: 'M4.5 12.75l6 6 9-13.5',
  },
  error: {
    card: 'border-red-500/30 bg-red-950/70',
    iconWrap: 'bg-red-500/15 text-red-300',
    icon: 'M6 18L18 6M6 6l12 12',
  },
  warning: {
    card: 'border-amber-500/30 bg-amber-950/70',
    iconWrap: 'bg-amber-500/15 text-amber-300',
    icon: 'M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z',
  },
  info: {
    card: 'border-teal-500/30 bg-teal-950/70',
    iconWrap: 'bg-teal-500/15 text-teal-300',
    icon: 'M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z',
  },
}
</script>
