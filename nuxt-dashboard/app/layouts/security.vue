<template>
  <div class="app-shell-bg app-grid relative min-h-screen font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh />
    <SecuritySidebar />
    <div class="relative z-10 min-h-screen lg:pl-72">
      <SecurityTopbar />
      <main class="mx-auto max-w-[1600px] p-4 md:p-6 lg:p-8">
        <Transition name="page" mode="out-in">
          <div :key="route.path" class="animate-fade-up">
            <slot />
          </div>
        </Transition>
      </main>
      <footer class="border-t px-4 py-6 text-center text-xs text-gray-600 md:px-8" style="border-color: var(--app-border)">
        <p>Security Ops Video Analytics Platform &middot; All systems monitored</p>
      </footer>
    </div>
    <ToastHost />
    <ConfirmHost />
    <CommandPalette />
  </div>
</template>

<script setup lang="ts">
import AmbientMesh from '~/components/ui/AmbientMesh.vue'
import SecuritySidebar from '~/components/security/SecuritySidebar.vue'
import SecurityTopbar from '~/components/security/SecurityTopbar.vue'
import ToastHost from '~/components/ui/ToastHost.vue'
import ConfirmHost from '~/components/ui/ConfirmHost.vue'
import CommandPalette from '~/components/ui/CommandPalette.vue'

const route = useRoute()
</script>

<style scoped>
.page-enter-active {
  transition: all 0.3s cubic-bezier(0.22, 1, 0.36, 1);
}
.page-leave-active {
  transition: all 0.15s ease;
}
.page-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.page-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
