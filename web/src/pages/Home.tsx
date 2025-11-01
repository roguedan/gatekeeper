import { Link } from 'react-router-dom'
import { Shield, Key, Lock, Zap } from 'lucide-react'
import { useAuthContext } from '@/contexts'
import { SignInFlow } from '@/components/auth'
import { Card, Button } from '@/components/common'

export const Home = () => {
  const { isAuthenticated } = useAuthContext()

  return (
    <div className="space-y-8 sm:space-y-10 lg:space-y-12">
      {/* Hero Section */}
      <div className="text-center space-y-4 sm:space-y-6 px-4 sm:px-0">
        <h1 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold text-gray-900 dark:text-white leading-tight">
          Welcome to <span className="text-primary-600">Gatekeeper</span>
        </h1>
        <p className="text-base sm:text-lg lg:text-xl text-gray-600 dark:text-gray-400 max-w-3xl mx-auto leading-relaxed px-2">
          Wallet-native authentication gateway using Sign-In with Ethereum (SIWE) and blockchain-based access control
        </p>
      </div>

      {/* Sign In Section */}
      {!isAuthenticated && (
        <div className="max-w-2xl mx-auto">
          <SignInFlow />
        </div>
      )}

      {/* Features Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-5 lg:gap-6">
        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-11 w-11 sm:h-12 sm:w-12 rounded-full bg-primary-100 dark:bg-primary-900 mb-3 sm:mb-4">
            <Shield className="h-5 w-5 sm:h-6 sm:w-6 text-primary-600 dark:text-primary-400" />
          </div>
          <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-2 leading-tight">SIWE Authentication</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
            Sign in with your Ethereum wallet using the EIP-4361 standard
          </p>
        </Card>

        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-11 w-11 sm:h-12 sm:w-12 rounded-full bg-secondary-100 dark:bg-secondary-900 mb-3 sm:mb-4">
            <Key className="h-5 w-5 sm:h-6 sm:w-6 text-secondary-600 dark:text-secondary-400" />
          </div>
          <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-2 leading-tight">API Key Management</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
            Create and manage API keys for programmatic access to protected resources
          </p>
        </Card>

        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-11 w-11 sm:h-12 sm:w-12 rounded-full bg-green-100 dark:bg-green-900 mb-3 sm:mb-4">
            <Lock className="h-5 w-5 sm:h-6 sm:w-6 text-green-600 dark:text-green-400" />
          </div>
          <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-2 leading-tight">Token Gating</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
            Restrict access based on blockchain token holdings and on-chain data
          </p>
        </Card>

        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-11 w-11 sm:h-12 sm:w-12 rounded-full bg-yellow-100 dark:bg-yellow-900 mb-3 sm:mb-4">
            <Zap className="h-5 w-5 sm:h-6 sm:w-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-2 leading-tight">JWT Tokens</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
            Secure, stateless authentication with industry-standard JSON Web Tokens
          </p>
        </Card>
      </div>

      {/* CTA Section */}
      {isAuthenticated && (
        <div className="bg-primary-50 dark:bg-primary-900/20 rounded-lg p-5 sm:p-6 lg:p-8 text-center">
          <h2 className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white mb-3 sm:mb-4 leading-tight">Ready to get started?</h2>
          <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 mb-5 sm:mb-6 leading-relaxed px-2">
            Explore the dashboard and create your first API key
          </p>
          <div className="flex flex-col sm:flex-row justify-center gap-3 sm:gap-4 max-w-md mx-auto">
            <Link to="/dashboard" className="flex-1">
              <Button className="w-full">Go to Dashboard</Button>
            </Link>
            <Link to="/api-keys" className="flex-1">
              <Button variant="outline" className="w-full">Manage API Keys</Button>
            </Link>
          </div>
        </div>
      )}

      {/* How It Works */}
      <div className="max-w-4xl mx-auto px-2 sm:px-0">
        <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white text-center mb-6 sm:mb-8 leading-tight">How It Works</h2>
        <div className="space-y-5 sm:space-y-6">
          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              1
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">Connect Your Wallet</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                Use MetaMask, WalletConnect, or any supported Ethereum wallet to connect
              </p>
            </div>
          </div>

          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              2
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">Sign the Message</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                Sign a secure SIWE message to prove ownership of your wallet address
              </p>
            </div>
          </div>

          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              3
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">Get JWT Token</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                Receive a JWT token that authenticates all your API requests
              </p>
            </div>
          </div>

          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              4
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">Access Protected Resources</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                Use your token to access protected endpoints and create API keys
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
