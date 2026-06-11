<template>
  <Teleport to="body">
    <Transition name="overlay">
      <div
        v-if="current"
        class="fixed inset-0 z-[90] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm"
        @click.self="answer(current.id, false)"
        @keydown.esc="answer(current.id, false)"
      >
        <Transition name="pop" appear>
          <div
            v-if="current"
            :key="current.id"
            class="w-full max-w-sm rounded-2xl border border-white/10 bg-[#0c1117] p-5 shadow-2xl"
            role="alertdialog"
            aria-modal="true"
          >
            <div class="flex items-start gap-3">
              <span
                class="grid h-10 w-10 shrink-0 place-items-center rounded-full"
                :class="current.danger ? 'bg-red-500/15 text-red-300' : 'bg-teal-500/15 text-teal-300'"
              >
                <svg class="h-5 w-5" fill="none" stroke="currentColor" stroke-width="1.8" viewBox="0 0 24 24" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" :d="current.danger ? dangerIcon : infoIcon" />
                </svg>
              </span>
              <div class="min-w-0">
                <h3 class="text-base font-semibold text-gray-100">{{ current.title }}</h3>
                <p v-if="current.message" class="mt-1 text-sm text-gray-400">{{ current.message }}</p>
              </div>
            </div>
            <div class="mt-6 flex justify-end gap-2">
              <button type="button" class="btn-ghost" @click="answer(current.id, false)">
                {{ current.cancelLabel || 'Cancel' }}
              </button>
              <button
                type="button"
                :class="current.danger ? 'btn-danger' : 'btn-primary'"
                @click="answer(current.id, true)"
              >
                {{ current.confirmLabel || 'Confirm' }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
const { current, answer } = useConfirm()

const dangerIcon = 'M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z'
const infoIcon = 'M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z'
</script>
