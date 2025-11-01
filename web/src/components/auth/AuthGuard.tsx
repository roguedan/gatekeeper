import { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuthContext } from '@/contexts'
import { LoadingSpinner } from '@/components/common'

interface AuthGuardProps {
  children: ReactNode
  redirectTo?: string
}

export const AuthGuard = ({ children, redirectTo = '/' }: AuthGuardProps) => {
  const { isAuthenticated, isLoading } = useAuthContext()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" text="Checking authentication..." />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to={redirectTo} replace />
  }

  return <>{children}</>
}
