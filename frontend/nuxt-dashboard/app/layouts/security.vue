<template>
  <div class="min-h-screen bg-[#0f1419] font-[family-name:var(--font-outfit)]">
    <SecuritySidebar />
    <div class="pl-64 min-h-screen">
      <main class="p-4 md:p-6 lg:p-8 max-w-[1600px] mx-auto">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import SecuritySidebar from '~/components/security/SecuritySidebar.vue'
</script>
