<template>
  <div class="app-shell-bg app-grid relative min-h-screen font-[family-name:var(--font-outfit)] text-gray-200">
    <AmbientMesh />
    <SecuritySidebar />
    <div class="relative z-10 min-h-screen lg:pl-72">
      <SecurityTopbar />
      <main class="mx-auto max-w-[1600px] p-4 md:p-6 lg:p-8">
        <div :key="route.path" class="animate-fade-up">
          <slot />
        </div>
      </main>
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
