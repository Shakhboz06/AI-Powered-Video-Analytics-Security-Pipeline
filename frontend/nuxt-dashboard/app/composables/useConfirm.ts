/**
 * Promise-based confirmation dialogs.
 *
 * Usage:
 *   const { confirm } = useConfirm()
 *   if (await confirm({ title: 'Delete camera?', tone: 'danger' })) { ... }
 */
export interface ConfirmOptions {
  title: string
  message?: string
  confirmLabel?: string
  cancelLabel?: string
  tone?: 'danger' | 'default'
}

interface ConfirmState extends ConfirmOptions {
  open: boolean
}

let resolver: ((value: boolean) => void) | null = null

export function useConfirm() {
  const state = useState<ConfirmState>('app-confirm', () => ({
    open: false,
    title: '',
  }))

  function confirm(opts: ConfirmOptions): Promise<boolean> {
    state.value = { ...opts, open: true }
    return new Promise<boolean>((resolve) => {
      resolver = resolve
    })
  }

  function resolve(value: boolean) {
    state.value = { ...state.value, open: false }
    if (resolver) {
      resolver(value)
      resolver = null
    }
  }

  return { state, confirm, accept: () => resolve(true), cancel: () => resolve(false) }
}
