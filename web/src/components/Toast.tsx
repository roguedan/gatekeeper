import { useEffect } from 'react'
import clsx from 'clsx'
import { X, CheckCircle, AlertCircle, Loader2, Info } from 'lucide-react'

export type ToastVariant = 'info' | 'success' | 'error' | 'pending'

export interface ToastProps {
  id: string
  variant: ToastVariant
  title: string
  description?: string
  duration?: number
  onClose: (id: string) => void
}

export const Toast = ({ id, variant, title, description, duration, onClose }: ToastProps) => {
  useEffect(() => {
    // Auto-dismiss on success after duration (default 5s)
    // Errors and pending states persist until manually dismissed
    if (variant === 'success' || variant === 'info') {
      const timer = setTimeout(() => {
        onClose(id)
      }, duration || 5000)

      return () => clearTimeout(timer)
    }
  }, [id, variant, duration, onClose])

  const variantClasses = {
    info: 'bg-blue-50 border-blue-200 text-blue-900 dark:bg-blue-900/20 dark:border-blue-800 dark:text-blue-100',
    success: 'bg-green-50 border-green-200 text-green-900 dark:bg-green-900/20 dark:border-green-800 dark:text-green-100',
    error: 'bg-red-50 border-red-200 text-red-900 dark:bg-red-900/20 dark:border-red-800 dark:text-red-100',
    pending: 'bg-blue-50 border-blue-200 text-blue-900 dark:bg-blue-900/20 dark:border-blue-800 dark:text-blue-100',
  }

  const icons = {
    info: Info,
    success: CheckCircle,
    error: AlertCircle,
    pending: Loader2,
  }

  const Icon = icons[variant]

  return (
    <div
      className={clsx(
        'flex items-start gap-2 sm:gap-3 p-3 sm:p-4 rounded-lg border shadow-lg transition-all duration-300 animate-slide-in',
        'min-w-[280px] max-w-[calc(100vw-2rem)] sm:min-w-[300px] sm:max-w-md',
        variantClasses[variant]
      )}
      role="alert"
      aria-live="polite"
    >
      <Icon
        className={clsx(
          'h-4 w-4 sm:h-5 sm:w-5 flex-shrink-0 mt-0.5',
          variant === 'pending' && 'animate-spin'
        )}
      />

      <div className="flex-1 min-w-0">
        <p className="text-xs sm:text-sm font-semibold leading-tight break-words">{title}</p>
        {description && (
          <p className="text-xs sm:text-sm mt-1 opacity-90 break-words leading-relaxed">{description}</p>
        )}
      </div>

      <button
        onClick={() => onClose(id)}
        className="flex-shrink-0 hover:opacity-70 transition-opacity p-1 min-h-[32px] min-w-[32px] flex items-center justify-center"
        aria-label="Close notification"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  )
}
