import { createContext, useContext, ReactNode } from 'react'
import { useAuth } from '@/hooks'
import { AuthState } from '@/types'

interface AuthContextValue extends AuthState {
  setAuthenticated: (address: string, token: string) => void
  setError: (error: string) => void
  setLoading: (isLoading: boolean) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const auth = useAuth()

  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>
}

export const useAuthContext = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuthContext must be used within AuthProvider')
  }
  return context
}
