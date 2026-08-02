import { defineNuxtPlugin } from '#app'
import { SidebarSymbol, createSidebarContext } from '~/composables/useSidebar'

export default defineNuxtPlugin((nuxtApp) => {
  const ctx = createSidebarContext()
  nuxtApp.vueApp.provide(SidebarSymbol, ctx)
})
