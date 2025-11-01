import clsx from 'clsx'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  className?: string
  text?: string
}

export const LoadingSpinner = ({ size = 'md', className, text }: LoadingSpinnerProps) => {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-7 w-7 sm:h-8 sm:w-8',
    lg: 'h-10 w-10 sm:h-12 sm:w-12',
  }

  return (
    <div className="flex flex-col items-center justify-center gap-2 sm:gap-3">
      <svg
        className={clsx('animate-spin text-primary-600', sizeClasses[size], className)}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
      {text && (
        <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 animate-pulse text-center leading-relaxed px-2">
          {text}
        </p>
      )}
    </div>
  )
}
