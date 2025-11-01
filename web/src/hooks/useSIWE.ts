import { useState, useCallback } from 'react'
import { useSignMessage, useAccount } from 'wagmi'
import { authService } from '@/services'
import { useAuth } from './useAuth'
import { useToastContext } from '@/contexts'

export const useSIWE = () => {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { signMessageAsync } = useSignMessage()
  const { address, chainId } = useAccount()
  const { setAuthenticated } = useAuth()
  const { showPending, showError, updateToast } = useToastContext()

  const signIn = useCallback(async () => {
    if (!address || !chainId) {
      const errorMsg = 'Wallet not connected'
      setError(errorMsg)
      showError('Connection Required', errorMsg)
      return
    }

    setIsLoading(true)
    setError(null)

    // Show initial pending toast
    const toastId = showPending('Signing in...', 'Please sign the message in your wallet')

    try {
      // 1. Get nonce from backend
      const nonce = await authService.getNonce()

      // 2. Create SIWE message
      const message = authService.createSiweMessage(address, nonce, chainId)

      // Update toast for signing step
      updateToast(toastId, {
        title: 'Waiting for signature...',
        description: 'Please check your wallet to sign the message'
      })

      // 3. Sign message with wallet
      const signature = await signMessageAsync({ message })

      // Update toast for verification step
      updateToast(toastId, {
        title: 'Verifying signature...',
        description: 'Authenticating with backend server'
      })

      // 4. Verify signature and get JWT token
      const { token, address: verifiedAddress } = await authService.verifySiwe(message, signature)

      // 5. Update auth state
      setAuthenticated(verifiedAddress, token)

      // Update toast to success
      updateToast(toastId, {
        variant: 'success',
        title: 'Signed in successfully!',
        description: `Welcome back, ${verifiedAddress.slice(0, 6)}...${verifiedAddress.slice(-4)}`
      })

      return { success: true, token, address: verifiedAddress }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to sign in'
      setError(errorMessage)

      // Update toast to error
      updateToast(toastId, {
        variant: 'error',
        title: 'Sign in failed',
        description: errorMessage
      })

      return { success: false, error: errorMessage }
    } finally {
      setIsLoading(false)
    }
  }, [address, chainId, signMessageAsync, setAuthenticated, showPending, showError, updateToast])

  return {
    signIn,
    isLoading,
    error,
  }
}
