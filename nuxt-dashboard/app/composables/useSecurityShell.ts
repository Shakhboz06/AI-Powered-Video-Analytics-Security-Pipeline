/**
 * Shared UI state for the Security Ops app shell (layout + sidebar + topbar).
 */
export function useSecurityShell() {
  const sidebarOpen = useState<boolean>('security-sidebar-open', () => false)

  function openSidebar() {
    sidebarOpen.value = true
  }
  function closeSidebar() {
    sidebarOpen.value = false
  }
  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
  }

  return { sidebarOpen, openSidebar, closeSidebar, toggleSidebar }
}
