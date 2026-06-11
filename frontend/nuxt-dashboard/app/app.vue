<template>
  <div>
    <NuxtLoadingIndicator :height="3" color="repeating-linear-gradient(to right, #2dd4bf 0%, #22d3ee 50%, #818cf8 100%)" />
    <NuxtLayout>
      <NuxtPage />
      <NuxtRouteAnnouncer />
    </NuxtLayout>
    <SplashScreen />
  </div>
</template>

<script setup lang="ts">
import SplashScreen from '~/components/ui/SplashScreen.vue'
</script>
