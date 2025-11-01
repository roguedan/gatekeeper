import { useState, useCallback } from 'react'
import { useSignMessage, useAccount } from 'wagmi'
import { authService } from '@/services'
import { useAuth } from './useAuth'

export const useSIWE = () => {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { signMessageAsync } = useSignMessage()
  const { address, chainId } = useAccount()
  const { setAuthenticated } = useAuth()

  const signIn = useCallback(async () => {
    if (!address || !chainId) {
      setError('Wallet not connected')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      // 1. Get nonce from backend
      const nonce = await authService.getNonce()

      // 2. Create SIWE message
      const message = authService.createSiweMessage(address, nonce, chainId)

      // 3. Sign message with wallet
      const signature = await signMessageAsync({ message })

      // 4. Verify signature and get JWT token
      const { token, address: verifiedAddress } = await authService.verifySiwe(message, signature)

      // 5. Update auth state
      setAuthenticated(verifiedAddress, token)

      return { success: true, token, address: verifiedAddress }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to sign in'
      setError(errorMessage)
      return { success: false, error: errorMessage }
    } finally {
      setIsLoading(false)
    }
  }, [address, chainId, signMessageAsync, setAuthenticated])

  return {
    signIn,
    isLoading,
    error,
  }
}
