import { ReactNode } from 'react'
import clsx from 'clsx'

interface CardProps {
  children: ReactNode
  className?: string
  title?: string
  description?: string
}

export const Card = ({ children, className, title, description }: CardProps) => {
  return (
    <div className={clsx('bg-white dark:bg-gray-800 rounded-lg shadow-md p-6', className)}>
      {title && (
        <div className="mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h3>
          {description && <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{description}</p>}
        </div>
      )}
      {children}
    </div>
  )
}
