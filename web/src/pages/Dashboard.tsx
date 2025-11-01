import { useAccount } from 'wagmi'
import { useAuthContext } from '@/contexts'
import { Card, Alert } from '@/components/common'
import { User, Key, Shield, Clock } from 'lucide-react'
import { useAPIKeys } from '@/hooks'
import { format } from 'date-fns'

export const Dashboard = () => {
  const { address } = useAccount()
  const { isAuthenticated } = useAuthContext()
  const { apiKeys } = useAPIKeys()

  const activeKeys = apiKeys.filter((key) => !key.isExpired).length
  const totalKeys = apiKeys.length

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Dashboard</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          Overview of your Gatekeeper account and activity
        </p>
      </div>

      {/* Account Info */}
      <Card title="Account Information">
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <User className="h-5 w-5 text-gray-500" />
            <div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Wallet Address</p>
              <p className="text-sm text-gray-900 dark:text-white font-mono">{address || 'Not connected'}</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <Shield className="h-5 w-5 text-gray-500" />
            <div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Authentication Status</p>
              <p className="text-sm">
                {isAuthenticated ? (
                  <span className="text-green-600 dark:text-green-400 font-medium">Authenticated</span>
                ) : (
                  <span className="text-red-600 dark:text-red-400 font-medium">Not Authenticated</span>
                )}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <Clock className="h-5 w-5 text-gray-500" />
            <div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">Session Started</p>
              <p className="text-sm text-gray-900 dark:text-white">{format(new Date(), 'PPpp')}</p>
            </div>
          </div>
        </div>
      </Card>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Total API Keys</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">{totalKeys}</p>
            </div>
            <div className="h-12 w-12 bg-primary-100 dark:bg-primary-900 rounded-lg flex items-center justify-center">
              <Key className="h-6 w-6 text-primary-600 dark:text-primary-400" />
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Active Keys</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">{activeKeys}</p>
            </div>
            <div className="h-12 w-12 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
              <Shield className="h-6 w-6 text-green-600 dark:text-green-400" />
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Expired Keys</p>
              <p className="text-3xl font-bold text-gray-900 dark:text-white mt-1">{totalKeys - activeKeys}</p>
            </div>
            <div className="h-12 w-12 bg-red-100 dark:bg-red-900 rounded-lg flex items-center justify-center">
              <Clock className="h-6 w-6 text-red-600 dark:text-red-400" />
            </div>
          </div>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card title="Quick Actions" description="Common tasks and shortcuts">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <a
            href="/api-keys"
            className="p-4 border-2 border-gray-200 dark:border-gray-700 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <div className="flex items-center gap-3">
              <Key className="h-5 w-5 text-primary-600 dark:text-primary-400" />
              <div>
                <p className="font-medium text-gray-900 dark:text-white">Manage API Keys</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Create, view, and revoke API keys</p>
              </div>
            </div>
          </a>

          <a
            href="/token-gating"
            className="p-4 border-2 border-gray-200 dark:border-gray-700 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors"
          >
            <div className="flex items-center gap-3">
              <Shield className="h-5 w-5 text-primary-600 dark:text-primary-400" />
              <div>
                <p className="font-medium text-gray-900 dark:text-white">Token Gating Demo</p>
                <p className="text-sm text-gray-600 dark:text-gray-400">Test blockchain-based access control</p>
              </div>
            </div>
          </a>
        </div>
      </Card>

      {/* Info Alert */}
      <Alert variant="info">
        <div>
          <p className="font-medium mb-1">Welcome to Gatekeeper!</p>
          <p className="text-sm">
            This dashboard provides an overview of your account. Use the navigation menu to access API keys, token-gating features, and more.
          </p>
        </div>
      </Alert>
    </div>
  )
}
