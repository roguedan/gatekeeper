import { useAccount } from 'wagmi'
import { ConnectButton } from '@rainbow-me/rainbowkit'
import { useSIWE } from '@/hooks'
import { useAuthContext } from '@/contexts'
import { Button, Card, Alert, LoadingSpinner } from '@/components/common'
import { LogIn, CheckCircle, Shield, Wallet } from 'lucide-react'

export const SignInFlow = () => {
  const { isConnected } = useAccount()
  const { signIn, isLoading, error } = useSIWE()
  const { isAuthenticated } = useAuthContext()

  if (isAuthenticated) {
    return (
      <Card className="max-w-md mx-auto">
        <div className="text-center py-2">
          <div className="mx-auto flex items-center justify-center h-12 w-12 sm:h-14 sm:w-14 rounded-full bg-green-100 dark:bg-green-900 mb-3 sm:mb-4">
            <CheckCircle className="h-6 w-6 sm:h-7 sm:w-7 text-green-600 dark:text-green-400" />
          </div>
          <h3 className="text-base sm:text-lg font-medium text-gray-900 dark:text-white mb-2 leading-tight">Already Signed In</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">You are already authenticated with Gatekeeper.</p>
        </div>
      </Card>
    )
  }

  return (
    <Card title="Sign In with Ethereum" description="Connect your wallet and sign a message to authenticate" className="max-w-md mx-auto">
      <div className="space-y-4">
        {error && <Alert variant="error">{error}</Alert>}

        {!isConnected ? (
          <div className="text-center space-y-4 py-2">
            <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-primary-100 dark:bg-primary-900/30 mb-2">
              <Wallet className="h-8 w-8 text-primary-600 dark:text-primary-400" />
            </div>
            <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed px-2">First, connect your wallet to continue</p>
            <div className="flex justify-center">
              <ConnectButton />
            </div>
          </div>
        ) : isLoading ? (
          <div className="py-8">
            <LoadingSpinner size="lg" text="Waiting for signature..." />
            <div className="mt-6 space-y-2">
              <div className="flex items-center justify-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                <Shield className="h-4 w-4" />
                <span>Secure authentication in progress</span>
              </div>
              <p className="text-xs text-gray-500 dark:text-gray-500 text-center">
                Check your wallet for the signature request
              </p>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <Alert variant="info">
              <p className="text-sm leading-relaxed">
                Click below to sign a message with your wallet. This proves you own the address without revealing your private key.
              </p>
            </Alert>

            <Button onClick={signIn} isLoading={isLoading} fullWidth>
              <LogIn className="h-5 w-5 mr-2" />
              <span className="text-sm sm:text-base">Sign In with Ethereum</span>
            </Button>

            <p className="text-xs sm:text-sm text-center text-gray-500 dark:text-gray-400 leading-relaxed px-2">
              By signing in, you agree to our Terms of Service and Privacy Policy
            </p>
          </div>
        )}
      </div>
    </Card>
  )
}
