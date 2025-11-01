import { Link } from 'react-router-dom'
import { Shield, Key, Lock, Zap } from 'lucide-react'
import { useAuthContext } from '@/contexts'
import { SignInFlow } from '@/components/auth'
import { Card, Button } from '@/components/common'

export const Home = () => {
  const { isAuthenticated } = useAuthContext()

  return (
    <div className="space-y-12">
      {/* Hero Section */}
      <div className="text-center space-y-6">
        <h1 className="text-4xl md:text-6xl font-bold text-gray-900 dark:text-white">
          Welcome to <span className="text-primary-600">Gatekeeper</span>
        </h1>
        <p className="text-xl text-gray-600 dark:text-gray-400 max-w-3xl mx-auto">
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
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mt-12">
        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-primary-100 dark:bg-primary-900 mb-4">
            <Shield className="h-6 w-6 text-primary-600 dark:text-primary-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">SIWE Authentication</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Sign in with your Ethereum wallet using the EIP-4361 standard
          </p>
        </Card>

        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-secondary-100 dark:bg-secondary-900 mb-4">
            <Key className="h-6 w-6 text-secondary-600 dark:text-secondary-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">API Key Management</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Create and manage API keys for programmatic access to protected resources
          </p>
        </Card>

        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 dark:bg-green-900 mb-4">
            <Lock className="h-6 w-6 text-green-600 dark:text-green-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">Token Gating</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Restrict access based on blockchain token holdings and on-chain data
          </p>
        </Card>

        <Card className="text-center">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-yellow-100 dark:bg-yellow-900 mb-4">
            <Zap className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">JWT Tokens</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Secure, stateless authentication with industry-standard JSON Web Tokens
          </p>
        </Card>
      </div>

      {/* CTA Section */}
      {isAuthenticated && (
        <div className="bg-primary-50 dark:bg-primary-900/20 rounded-lg p-8 text-center">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">Ready to get started?</h2>
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            Explore the dashboard and create your first API key
          </p>
          <div className="flex justify-center gap-4">
            <Link to="/dashboard">
              <Button>Go to Dashboard</Button>
            </Link>
            <Link to="/api-keys">
              <Button variant="outline">Manage API Keys</Button>
            </Link>
          </div>
        </div>
      )}

      {/* How It Works */}
      <div className="max-w-4xl mx-auto">
        <h2 className="text-3xl font-bold text-gray-900 dark:text-white text-center mb-8">How It Works</h2>
        <div className="space-y-6">
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              1
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">Connect Your Wallet</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Use MetaMask, WalletConnect, or any supported Ethereum wallet to connect
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              2
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">Sign the Message</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Sign a secure SIWE message to prove ownership of your wallet address
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              3
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">Get JWT Token</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Receive a JWT token that authenticates all your API requests
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              4
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">Access Protected Resources</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Use your token to access protected endpoints and create API keys
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
