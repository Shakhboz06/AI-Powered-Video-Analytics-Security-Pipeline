/** Global open/close state for the command palette (Ctrl/Cmd-K). */
export function useCommandPalette() {
  const open = useState<boolean>('command-palette-open', () => false)

  return {
    open,
    openPalette: () => { open.value = true },
    closePalette: () => { open.value = false },
    toggle: () => { open.value = !open.value },
  }
}
