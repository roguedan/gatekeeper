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
    <div className="space-y-6 sm:space-y-8">
      <div>
        <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white leading-tight">Dashboard</h1>
        <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 mt-1 sm:mt-2 leading-relaxed">
          Overview of your Gatekeeper account and activity
        </p>
      </div>

      {/* Account Info */}
      <Card title="Account Information">
        <div className="space-y-3 sm:space-y-4">
          <div className="flex items-start gap-3">
            <User className="h-5 w-5 text-gray-500 flex-shrink-0 mt-0.5" />
            <div className="flex-1 min-w-0">
              <p className="text-xs sm:text-sm font-medium text-gray-700 dark:text-gray-300">Wallet Address</p>
              <p className="text-xs sm:text-sm text-gray-900 dark:text-white font-mono break-all">{address || 'Not connected'}</p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <Shield className="h-5 w-5 text-gray-500 flex-shrink-0 mt-0.5" />
            <div className="flex-1 min-w-0">
              <p className="text-xs sm:text-sm font-medium text-gray-700 dark:text-gray-300">Authentication Status</p>
              <p className="text-xs sm:text-sm">
                {isAuthenticated ? (
                  <span className="text-green-600 dark:text-green-400 font-medium">Authenticated</span>
                ) : (
                  <span className="text-red-600 dark:text-red-400 font-medium">Not Authenticated</span>
                )}
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <Clock className="h-5 w-5 text-gray-500 flex-shrink-0 mt-0.5" />
            <div className="flex-1 min-w-0">
              <p className="text-xs sm:text-sm font-medium text-gray-700 dark:text-gray-300">Session Started</p>
              <p className="text-xs sm:text-sm text-gray-900 dark:text-white break-words">{format(new Date(), 'PPpp')}</p>
            </div>
          </div>
        </div>
      </Card>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5 lg:gap-6">
        <Card>
          <div className="flex items-center justify-between gap-3">
            <div className="flex-1 min-w-0">
              <p className="text-xs sm:text-sm font-medium text-gray-600 dark:text-gray-400">Total API Keys</p>
              <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white mt-1">{totalKeys}</p>
            </div>
            <div className="h-11 w-11 sm:h-12 sm:w-12 bg-primary-100 dark:bg-primary-900 rounded-lg flex items-center justify-center flex-shrink-0">
              <Key className="h-5 w-5 sm:h-6 sm:w-6 text-primary-600 dark:text-primary-400" />
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between gap-3">
            <div className="flex-1 min-w-0">
              <p className="text-xs sm:text-sm font-medium text-gray-600 dark:text-gray-400">Active Keys</p>
              <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white mt-1">{activeKeys}</p>
            </div>
            <div className="h-11 w-11 sm:h-12 sm:w-12 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center flex-shrink-0">
              <Shield className="h-5 w-5 sm:h-6 sm:w-6 text-green-600 dark:text-green-400" />
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between gap-3">
            <div className="flex-1 min-w-0">
              <p className="text-xs sm:text-sm font-medium text-gray-600 dark:text-gray-400">Expired Keys</p>
              <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white mt-1">{totalKeys - activeKeys}</p>
            </div>
            <div className="h-11 w-11 sm:h-12 sm:w-12 bg-red-100 dark:bg-red-900 rounded-lg flex items-center justify-center flex-shrink-0">
              <Clock className="h-5 w-5 sm:h-6 sm:w-6 text-red-600 dark:text-red-400" />
            </div>
          </div>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card title="Quick Actions" description="Common tasks and shortcuts">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
          <a
            href="/api-keys"
            className="p-3 sm:p-4 border-2 border-gray-200 dark:border-gray-700 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors min-h-[80px] flex items-center"
          >
            <div className="flex items-center gap-3">
              <Key className="h-5 w-5 sm:h-6 sm:w-6 text-primary-600 dark:text-primary-400 flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <p className="text-sm sm:text-base font-medium text-gray-900 dark:text-white">Manage API Keys</p>
                <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 leading-relaxed">Create, view, and revoke API keys</p>
              </div>
            </div>
          </a>

          <a
            href="/token-gating"
            className="p-3 sm:p-4 border-2 border-gray-200 dark:border-gray-700 rounded-lg hover:border-primary-500 dark:hover:border-primary-500 transition-colors min-h-[80px] flex items-center"
          >
            <div className="flex items-center gap-3">
              <Shield className="h-5 w-5 sm:h-6 sm:w-6 text-primary-600 dark:text-primary-400 flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <p className="text-sm sm:text-base font-medium text-gray-900 dark:text-white">Token Gating Demo</p>
                <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 leading-relaxed">Test blockchain-based access control</p>
              </div>
            </div>
          </a>
        </div>
      </Card>

      {/* Info Alert */}
      <Alert variant="info">
        <div>
          <p className="font-medium mb-1 text-sm sm:text-base">Welcome to Gatekeeper!</p>
          <p className="text-xs sm:text-sm leading-relaxed">
            This dashboard provides an overview of your account. Use the navigation menu to access API keys, token-gating features, and more.
          </p>
        </div>
      </Alert>
    </div>
  )
}
