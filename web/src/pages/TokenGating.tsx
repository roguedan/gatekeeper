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
    <div className="space-y-6 sm:space-y-8">
      <div>
        <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white leading-tight">Token Gating Demo</h1>
        <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 mt-1 sm:mt-2 leading-relaxed">
          Test blockchain-based access control and protected resource access
        </p>
      </div>

      {/* What is Token Gating */}
      <Card title="What is Token Gating?">
        <div className="space-y-3 text-sm sm:text-base text-gray-700 dark:text-gray-300">
          <p className="leading-relaxed">
            Token gating is a method of restricting access to resources based on blockchain token ownership or on-chain data. This allows you to:
          </p>
          <ul className="list-disc list-inside space-y-2 ml-2 sm:ml-4 leading-relaxed">
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
            <div className="text-center py-6 sm:py-8">
              <div className="mx-auto flex items-center justify-center h-14 w-14 sm:h-16 sm:w-16 rounded-full bg-primary-100 dark:bg-primary-900 mb-3 sm:mb-4">
                <Lock className="h-7 w-7 sm:h-8 sm:w-8 text-primary-600 dark:text-primary-400" />
              </div>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 mb-5 sm:mb-6 leading-relaxed px-2">
                Click below to attempt accessing a protected resource
              </p>
              <Button onClick={handleFetchData} isLoading={isLoading}>
                <Shield className="h-5 w-5 mr-2" />
                <span className="text-sm sm:text-base">Access Protected Data</span>
              </Button>
            </div>
          ) : (
            <>
              {isLoading && (
                <div className="text-center py-6 sm:py-8">
                  <LoadingSpinner size="lg" />
                  <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 mt-4 leading-relaxed">Checking access permissions...</p>
                </div>
              )}

              {error && !isLoading && (
                <Alert variant="error">
                  <div>
                    <p className="font-medium mb-1 text-sm sm:text-base">Access Denied</p>
                    <p className="text-xs sm:text-sm leading-relaxed">{error}</p>
                  </div>
                </Alert>
              )}

              {message && !isLoading && (
                <Alert variant="success">
                  <div className="space-y-3">
                    <div className="flex items-center gap-2">
                      <Unlock className="h-5 w-5 flex-shrink-0" />
                      <p className="font-medium text-sm sm:text-base">Access Granted!</p>
                    </div>
                    <p className="text-xs sm:text-sm leading-relaxed">{message}</p>

                    {data && (
                      <div className="mt-3 p-2 sm:p-3 bg-white/50 dark:bg-gray-900/50 rounded border border-green-300 dark:border-green-700">
                        <div className="flex items-start gap-2">
                          <Database className="h-5 w-5 flex-shrink-0 mt-0.5" />
                          <div className="flex-1 min-w-0">
                            <p className="font-medium text-xs sm:text-sm mb-2">Protected Data:</p>
                            <pre className="text-xs font-mono overflow-x-auto break-words whitespace-pre-wrap">
                              {JSON.stringify(data, null, 2)}
                            </pre>
                          </div>
                        </div>
                      </div>
                    )}

                    <Button onClick={handleFetchData} variant="outline" size="sm">
                      <span className="text-sm">Fetch Again</span>
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
        <div className="space-y-4 sm:space-y-5">
          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              1
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">User Authentication</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                User signs in with their Ethereum wallet using SIWE
              </p>
            </div>
          </div>

          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              2
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">Policy Evaluation</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                Backend checks if user meets the requirements (token holdings, NFTs, etc.)
              </p>
            </div>
          </div>

          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              3
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">Access Decision</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                If policy passes, user gets access to the protected resource
              </p>
            </div>
          </div>

          <div className="flex gap-3 sm:gap-4">
            <div className="flex-shrink-0 w-9 h-9 sm:w-10 sm:h-10 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold text-sm sm:text-base">
              4
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-white mb-1 leading-tight">Resource Delivery</h3>
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 leading-relaxed">
                Protected data or functionality is delivered to authorized users
              </p>
            </div>
          </div>
        </div>
      </Card>

      {/* Use Cases */}
      <Card title="Common Use Cases">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
          <div className="p-3 sm:p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="text-sm sm:text-base font-medium text-gray-900 dark:text-white mb-1.5 sm:mb-2 leading-tight">NFT Memberships</h4>
            <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
              Grant access to exclusive content or features based on NFT ownership
            </p>
          </div>

          <div className="p-3 sm:p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="text-sm sm:text-base font-medium text-gray-900 dark:text-white mb-1.5 sm:mb-2 leading-tight">Token Minimum Balance</h4>
            <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
              Require users to hold a minimum amount of a specific token
            </p>
          </div>

          <div className="p-3 sm:p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="text-sm sm:text-base font-medium text-gray-900 dark:text-white mb-1.5 sm:mb-2 leading-tight">DAO Governance</h4>
            <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
              Restrict voting or proposal access to governance token holders
            </p>
          </div>

          <div className="p-3 sm:p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
            <h4 className="text-sm sm:text-base font-medium text-gray-900 dark:text-white mb-1.5 sm:mb-2 leading-tight">Premium Features</h4>
            <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 leading-relaxed">
              Enable premium functionality for users with specific on-chain credentials
            </p>
          </div>
        </div>
      </Card>

      <Alert variant="info">
        <div>
          <p className="font-medium mb-1 text-sm sm:text-base">Note</p>
          <p className="text-xs sm:text-sm leading-relaxed">
            The demo endpoint uses configurable policies. In production, you can define complex rules based on smart contract state, token balances, NFT ownership, and more.
          </p>
        </div>
      </Alert>
    </div>
  )
}
