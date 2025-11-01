import { useEffect } from 'react'
import { useAccount } from 'wagmi'
import { ConnectButton } from '@rainbow-me/rainbowkit'
import { useSIWE } from '@/hooks'
import { useAuthContext } from '@/contexts'
import { Button, Card, Alert } from '@/components/common'
import { LogIn } from 'lucide-react'

export const SignInFlow = () => {
  const { isConnected } = useAccount()
  const { signIn, isLoading, error } = useSIWE()
  const { isAuthenticated } = useAuthContext()

  if (isAuthenticated) {
    return (
      <Card className="max-w-md mx-auto">
        <div className="text-center">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 dark:bg-green-900 mb-4">
            <LogIn className="h-6 w-6 text-green-600 dark:text-green-400" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">Already Signed In</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">You are already authenticated with Gatekeeper.</p>
        </div>
      </Card>
    )
  }

  return (
    <Card title="Sign In with Ethereum" description="Connect your wallet and sign a message to authenticate" className="max-w-md mx-auto">
      <div className="space-y-4">
        {error && <Alert variant="error">{error}</Alert>}

        {!isConnected ? (
          <div className="text-center space-y-4">
            <p className="text-sm text-gray-600 dark:text-gray-400">First, connect your wallet to continue</p>
            <ConnectButton />
          </div>
        ) : (
          <div className="space-y-4">
            <Alert variant="info">
              <p className="text-sm">
                Click below to sign a message with your wallet. This proves you own the address without revealing your private key.
              </p>
            </Alert>

            <Button onClick={signIn} isLoading={isLoading} fullWidth>
              <LogIn className="h-5 w-5 mr-2" />
              Sign In with Ethereum
            </Button>

            <p className="text-xs text-center text-gray-500 dark:text-gray-400">
              By signing in, you agree to our Terms of Service and Privacy Policy
            </p>
          </div>
        )}
      </div>
    </Card>
  )
}
