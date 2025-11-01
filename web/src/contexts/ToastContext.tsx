import { createContext, useContext, ReactNode } from 'react'
import { useToast } from '@/hooks/useToast'
import { ToastContainer } from '@/components/ToastContainer'

interface ToastContextValue {
  showPending: (title: string, description?: string) => string
  showSuccess: (title: string, description?: string, duration?: number) => string
  showError: (title: string, description?: string) => string
  showInfo: (title: string, description?: string, duration?: number) => string
  dismissToast: (id: string) => void
  updateToast: (id: string, updates: { variant?: 'info' | 'success' | 'error' | 'pending'; title?: string; description?: string }) => void
  dismissAll: () => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

export const ToastProvider = ({ children }: { children: ReactNode }) => {
  const {
    toasts,
    showPending,
    showSuccess,
    showError,
    showInfo,
    dismissToast,
    updateToast,
    dismissAll,
  } = useToast()

  return (
    <ToastContext.Provider
      value={{
        showPending,
        showSuccess,
        showError,
        showInfo,
        dismissToast,
        updateToast,
        dismissAll,
      }}
    >
      {children}
      <ToastContainer toasts={toasts} onClose={dismissToast} />
    </ToastContext.Provider>
  )
}

export const useToastContext = () => {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToastContext must be used within ToastProvider')
  }
  return context
}
