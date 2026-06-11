export type ThemeMode = 'dark' | 'light'

const STORAGE_KEY = 'app_theme'

/** App theme (dark default). Persists to localStorage and toggles `light` class on <html>. */
export function useTheme() {
  const theme = useState<ThemeMode>('app-theme', () => 'dark')

  function apply(mode: ThemeMode) {
    if (!import.meta.client) return
    const html = document.documentElement
    html.classList.toggle('light', mode === 'light')
    html.classList.toggle('dark', mode === 'dark')
  }

  function setTheme(mode: ThemeMode) {
    theme.value = mode
    if (import.meta.client) localStorage.setItem(STORAGE_KEY, mode)
    apply(mode)
  }

  function toggle() {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  function init() {
    if (!import.meta.client) return
    const saved = localStorage.getItem(STORAGE_KEY) as ThemeMode | null
    const mode: ThemeMode = saved === 'light' || saved === 'dark' ? saved : 'dark'
    theme.value = mode
    apply(mode)
  }

  return { theme, setTheme, toggle, init }
}
