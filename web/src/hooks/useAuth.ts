import { useState, useEffect, useCallback } from 'react'
import { authService } from '@/services'
import { AuthState } from '@/types'

export const useAuth = () => {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: authService.isAuthenticated(),
    address: authService.getCurrentAddress(),
    token: null,
    isLoading: false,
    error: null,
  })

  useEffect(() => {
    const handleLogout = () => {
      setAuthState({
        isAuthenticated: false,
        address: null,
        token: null,
        isLoading: false,
        error: null,
      })
    }

    window.addEventListener('auth:logout', handleLogout)
    return () => window.removeEventListener('auth:logout', handleLogout)
  }, [])

  const setAuthenticated = useCallback((address: string, token: string) => {
    setAuthState({
      isAuthenticated: true,
      address,
      token,
      isLoading: false,
      error: null,
    })
  }, [])

  const setError = useCallback((error: string) => {
    setAuthState((prev) => ({
      ...prev,
      isLoading: false,
      error,
    }))
  }, [])

  const setLoading = useCallback((isLoading: boolean) => {
    setAuthState((prev) => ({
      ...prev,
      isLoading,
    }))
  }, [])

  const logout = useCallback(() => {
    authService.logout()
    setAuthState({
      isAuthenticated: false,
      address: null,
      token: null,
      isLoading: false,
      error: null,
    })
  }, [])

  return {
    ...authState,
    setAuthenticated,
    setError,
    setLoading,
    logout,
  }
}
