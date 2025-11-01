import { ReactNode } from 'react'
import clsx from 'clsx'
import { X, AlertCircle, CheckCircle, Info, AlertTriangle } from 'lucide-react'

interface AlertProps {
  children: ReactNode
  variant?: 'success' | 'error' | 'warning' | 'info'
  onClose?: () => void
  className?: string
}

export const Alert = ({ children, variant = 'info', onClose, className }: AlertProps) => {
  const variantClasses = {
    success: 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-800 dark:text-green-300',
    error: 'bg-red-50 border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-800 dark:text-red-300',
    warning: 'bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-900/20 dark:border-yellow-800 dark:text-yellow-300',
    info: 'bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900/20 dark:border-blue-800 dark:text-blue-300',
  }

  const icons = {
    success: CheckCircle,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info,
  }

  const Icon = icons[variant]

  return (
    <div className={clsx('border rounded-lg p-4', variantClasses[variant], className)}>
      <div className="flex items-start">
        <Icon className="h-5 w-5 mr-3 flex-shrink-0 mt-0.5" />
        <div className="flex-1">{children}</div>
        {onClose && (
          <button onClick={onClose} className="ml-3 flex-shrink-0 hover:opacity-70">
            <X className="h-5 w-5" />
          </button>
        )}
      </div>
    </div>
  )
}
