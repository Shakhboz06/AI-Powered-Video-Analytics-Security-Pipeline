<template>
  <ThemeProvider>
    <div class="flex h-screen bg-gray-100 dark:bg-gray-900">
      <AppSidebar />
      <div class="flex flex-col w-full">
        <main class="p-6">
          <slot />
        </main>
      </div>
    </div>
  </ThemeProvider>
</template>

<script setup>
import ThemeProvider from '~/components/layout/ThemeProvider.vue'
import AppHeader from '../components/layout/AppHeader.vue';
import AppSidebar from '../components/layout/AppSidebar.vue';
</script>
