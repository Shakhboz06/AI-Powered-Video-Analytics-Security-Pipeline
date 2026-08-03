/**
 * Global toast notifications for the Security Ops dashboard.
 */
export type ToastType = 'success' | 'error' | 'info' | 'warning'

export interface ToastItem {
  id: number
  type: ToastType
  message: string
  description?: string
  duration: number
}

let counter = 0

export function useToast() {
  const toasts = useState<ToastItem[]>('app-toasts', () => [])

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function push(type: ToastType, message: string, opts?: { description?: string; duration?: number }) {
    const id = ++counter
    const duration = opts?.duration ?? 4000
    const item: ToastItem = { id, type, message, description: opts?.description, duration }
    toasts.value = [...toasts.value, item].slice(-5)
    if (duration > 0 && import.meta.client) {
      setTimeout(() => dismiss(id), duration)
    }
    return id
  }

  return {
    toasts,
    dismiss,
    push,
    success: (message: string, opts?: { description?: string; duration?: number }) => push('success', message, opts),
    error: (message: string, opts?: { description?: string; duration?: number }) => push('error', message, opts),
    info: (message: string, opts?: { description?: string; duration?: number }) => push('info', message, opts),
    warning: (message: string, opts?: { description?: string; duration?: number }) => push('warning', message, opts),
  }
}
