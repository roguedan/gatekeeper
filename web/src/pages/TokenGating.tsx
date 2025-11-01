import { useState } from 'react'
import { useProtectedData } from '@/hooks'
import { Card, Button, Alert, LoadingSpinner } from '@/components/common'
import { Shield, Lock, Unlock, Database } from 'lucide-react'

export const TokenGating = () => {
  const { data, message, isLoading, error, fetchData } = useProtectedData()
  const [hasAttempted, setHasAttempted] = useState(false)

  const handleFetchData = () => {
    setHasAttempted(true)
    fetchData()
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Token Gating Demo</h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          Test blockchain-based access control and protected resource access
        </p>
      </div>

      {/* What is Token Gating */}
      <Card title="What is Token Gating?">
        <div className="space-y-3 text-gray-700 dark:text-gray-300">
          <p>
            Token gating is a method of restricting access to resources based on blockchain token ownership or on-chain data. This allows you to:
          </p>
          <ul className="list-disc list-inside space-y-2 ml-4">
            <li>Require users to hold specific NFTs or tokens</li>
            <li>Check minimum token balances before granting access</li>
            <li>Verify on-chain reputation or credentials</li>
            <li>Create exclusive content for token holders</li>
          </ul>
        </div>
      </Card>

      {/* Demo Section */}
      <Card title="Protected Resource Access" description="Try accessing a token-gated endpoint">
        <div className="space-y-4">
          {!hasAttempted ? (
            <div className="text-center py-8">
              <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-primary-100 dark:bg-primary-900 mb-4">
                <Lock className="h-8 w-8 text-primary-600 dark:text-primary-400" />
              </div>
              <p className="text-gray-600 dark:text-gray-400 mb-6">
                Click below to attempt accessing a protected resource
              </p>
              <Button onClick={handleFetchData} isLoading={isLoading}>
                <Shield className="h-5 w-5 mr-2" />
                Access Protected Data
              </Button>
            </div>
          ) : (
            <>
              {isLoading && (
                <div className="text-center py-8">
                  <LoadingSpinner size="lg" />
                  <p className="text-gray-600 dark:text-gray-400 mt-4">Checking access permissions...</p>
                </div>
              )}

              {error && !isLoading && (
                <Alert variant="error">
                  <div>
                    <p className="font-medium mb-1">Access Denied</p>
                    <p className="text-sm">{error}</p>
                  </div>
                </Alert>
              )}

              {message && !isLoading && (
                <Alert variant="success">
                  <div className="space-y-3">
                    <div className="flex items-center gap-2">
                      <Unlock className="h-5 w-5" />
                      <p className="font-medium">Access Granted!</p>
                    </div>
                    <p className="text-sm">{message}</p>

                    {data && (
                      <div className="mt-3 p-3 bg-white/50 dark:bg-gray-900/50 rounded border border-green-300 dark:border-green-700">
                        <div className="flex items-start gap-2">
                          <Database className="h-5 w-5 flex-shrink-0 mt-0.5" />
                          <div className="flex-1">
                            <p className="font-medium text-sm mb-2">Protected Data:</p>
                            <pre className="text-xs font-mono overflow-x-auto">
                              {JSON.stringify(data, null, 2)}
                            </pre>
                          </div>
                        </div>
                      </div>
                    )}

                    <Button onClick={handleFetchData} variant="outline" size="sm">
                      Fetch Again
                    </Button>
                  </div>
                </Alert>
              )}
            </>
          )}
        </div>
      </Card>

      {/* How It Works */}
      <Card title="How Token Gating Works">
        <div className="space-y-4">
          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              1
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">User Authentication</h3>
              <p className="text-gray-600 dark:text-gray-400">
                User signs in with their Ethereum wallet using SIWE
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              2
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">Policy Evaluation</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Backend checks if user meets the requirements (token holdings, NFTs, etc.)
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              3
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">Access Decision</h3>
              <p className="text-gray-600 dark:text-gray-400">
                If policy passes, user gets access to the protected resource
              </p>
            </div>
          </div>

          <div className="flex gap-4">
            <div className="flex-shrink-0 w-10 h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
              4
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-1">Resource Delivery</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Protected data or functionality is delivered to authorized users
              </p>
            </div>
          </div>
        </div>
      </Card>

      {/* Use Cases */}
      <Card title="Common Use Cases">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="font-medium text-gray-900 dark:text-white mb-2">NFT Memberships</h4>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Grant access to exclusive content or features based on NFT ownership
            </p>
          </div>

          <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="font-medium text-gray-900 dark:text-white mb-2">Token Minimum Balance</h4>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Require users to hold a minimum amount of a specific token
            </p>
          </div>

          <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="font-medium text-gray-900 dark:text-white mb-2">DAO Governance</h4>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Restrict voting or proposal access to governance token holders
            </p>
          </div>

          <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="font-medium text-gray-900 dark:text-white mb-2">Premium Features</h4>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Enable premium functionality for users with specific on-chain credentials
            </p>
          </div>
        </div>
      </Card>

      <Alert variant="info">
        <div>
          <p className="font-medium mb-1">Note</p>
          <p className="text-sm">
            The demo endpoint uses configurable policies. In production, you can define complex rules based on smart contract state, token balances, NFT ownership, and more.
          </p>
        </div>
      </Alert>
    </div>
  )
}
