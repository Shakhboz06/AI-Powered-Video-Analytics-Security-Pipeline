// import { ref } from 'vue'

// export function useSidebar() {
//   const isMobileOpen = ref(false)
//   const isDesktopOpen = ref(true)

//   const toggleSidebar = () => {
//     isDesktopOpen.value = !isDesktopOpen.value
//   }

//   const toggleMobileSidebar = () => {
//     isMobileOpen.value = !isMobileOpen.value
//   }

//   return {
//     isMobileOpen,
//     isDesktopOpen,
//     toggleSidebar,
//     toggleMobileSidebar,
//   }
// }

// app/composables/useSidebar.ts
import { ref, computed, onMounted, onUnmounted, provide, inject } from 'vue'
import type { Ref } from 'vue'

export const SidebarSymbol = Symbol('Sidebar')

export interface SidebarContextType {
  isExpanded: Ref<boolean>
  isMobileOpen: Ref<boolean>
  isHovered: Ref<boolean>
  activeItem: Ref<string | null>
  openSubmenu: Ref<string | null>
  toggleSidebar: () => void
  toggleMobileSidebar: () => void
  setIsHovered: (isHovered: boolean) => void
  setActiveItem: (item: string | null) => void
  toggleSubmenu: (item: string) => void
}

export function createSidebarContext(): SidebarContextType {
  const isExpanded = ref(true)
  const isMobileOpen = ref(false)
  const isMobile = ref(false)
  const isHovered = ref(false)
  const activeItem = ref<string | null>(null)
  const openSubmenu = ref<string | null>(null)

  const handleResize = () => {
    const mobile = window.innerWidth < 768
    isMobile.value = mobile
    if (!mobile) isMobileOpen.value = false
  }

  if (process.client) {
    handleResize()
    window.addEventListener('resize', handleResize)
  }

  const toggleSidebar = () => {
    if (isMobile.value) isMobileOpen.value = !isMobileOpen.value
    else isExpanded.value = !isExpanded.value
  }
  const toggleMobileSidebar = () => { isMobileOpen.value = !isMobileOpen.value }
  const setIsHovered = (v: boolean) => { isHovered.value = v }
  const setActiveItem = (item: string | null) => { activeItem.value = item }
  const toggleSubmenu = (item: string) => {
    openSubmenu.value = openSubmenu.value === item ? null : item
  }

  return {
    isExpanded: computed(() => (isMobile.value ? false : isExpanded.value)),
    isMobileOpen, isHovered, activeItem, openSubmenu,
    toggleSidebar, toggleMobileSidebar, setIsHovered, setActiveItem, toggleSubmenu
  }
}

export function provideSidebar(ctx = createSidebarContext()) {
  provide(SidebarSymbol, ctx)
  return ctx
}

export function useSidebar(): SidebarContextType {
  const ctx = inject<SidebarContextType>(SidebarSymbol)
  if (!ctx) throw new Error('useSidebar must be used within a component that has SidebarProvider as an ancestor')
  return ctx
}
