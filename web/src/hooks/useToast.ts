import { useState, useCallback } from 'react'
import { ToastVariant } from '@/components/Toast'

export interface ToastMessage {
  id: string
  variant: ToastVariant
  title: string
  description?: string
  duration?: number
}

let toastCounter = 0

export const useToast = () => {
  const [toasts, setToasts] = useState<ToastMessage[]>([])

  const showToast = useCallback(
    (variant: ToastVariant, title: string, description?: string, duration?: number) => {
      const id = `toast-${Date.now()}-${toastCounter++}`
      const newToast: ToastMessage = {
        id,
        variant,
        title,
        description,
        duration,
      }

      setToasts((prev) => [...prev, newToast])
      return id
    },
    []
  )

  const showPending = useCallback(
    (title: string, description?: string) => {
      return showToast('pending', title, description)
    },
    [showToast]
  )

  const showSuccess = useCallback(
    (title: string, description?: string, duration?: number) => {
      return showToast('success', title, description, duration)
    },
    [showToast]
  )

  const showError = useCallback(
    (title: string, description?: string) => {
      return showToast('error', title, description)
    },
    [showToast]
  )

  const showInfo = useCallback(
    (title: string, description?: string, duration?: number) => {
      return showToast('info', title, description, duration)
    },
    [showToast]
  )

  const dismissToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id))
  }, [])

  const updateToast = useCallback(
    (id: string, updates: Partial<Omit<ToastMessage, 'id'>>) => {
      setToasts((prev) =>
        prev.map((toast) =>
          toast.id === id
            ? { ...toast, ...updates }
            : toast
        )
      )
    },
    []
  )

  const dismissAll = useCallback(() => {
    setToasts([])
  }, [])

  return {
    toasts,
    showPending,
    showSuccess,
    showError,
    showInfo,
    dismissToast,
    updateToast,
    dismissAll,
  }
}
