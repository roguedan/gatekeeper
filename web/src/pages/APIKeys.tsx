import { useState } from 'react'
import { useAPIKeys } from '@/hooks'
import { Card, Button, Alert, LoadingSpinner } from '@/components/common'
import { Plus, Key, Trash2, Copy, Clock, CheckCircle, XCircle } from 'lucide-react'
import { format } from 'date-fns'

export const APIKeys = () => {
  const { apiKeys, isLoading, createKey, revokeKey, isCreating, isRevoking, createError } = useAPIKeys()
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newKeyData, setNewKeyData] = useState({ name: '', scopes: 'read,write', expiresInDays: '' })
  const [createdKey, setCreatedKey] = useState<string | null>(null)
  const [copiedKeyId, setCopiedKeyId] = useState<number | null>(null)

  const handleCreateKey = () => {
    const scopes = newKeyData.scopes.split(',').map((s) => s.trim()).filter(Boolean)
    const expiresInSeconds = newKeyData.expiresInDays ? parseInt(newKeyData.expiresInDays) * 24 * 60 * 60 : undefined

    createKey(
      { name: newKeyData.name, scopes, expiresInSeconds },
      {
        onSuccess: (data) => {
          setCreatedKey(data.key)
          setNewKeyData({ name: '', scopes: 'read,write', expiresInDays: '' })
          setShowCreateForm(false)
        },
      }
    )
  }

  const handleCopyKey = async (keyHash: string) => {
    await navigator.clipboard.writeText(keyHash)
    setCopiedKeyId(Number(keyHash))
    setTimeout(() => setCopiedKeyId(null), 2000)
  }

  const handleCopyCreatedKey = async () => {
    if (createdKey) {
      await navigator.clipboard.writeText(createdKey)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <LoadingSpinner size="lg" text="Loading your API keys..." />
      </div>
    )
  }

  return (
    <div className="space-y-6 sm:space-y-8">
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-3 sm:gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white leading-tight">API Keys</h1>
          <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 mt-1 sm:mt-2 leading-relaxed">
            Manage your API keys for programmatic access to protected resources
          </p>
        </div>
        <Button onClick={() => setShowCreateForm(!showCreateForm)} data-testid="toggle-create-form-button" className="sm:whitespace-nowrap">
          <Plus className="h-5 w-5 mr-2" />
          <span className="text-sm sm:text-base">Create API Key</span>
        </Button>
      </div>

      {/* Created Key Alert */}
      {createdKey && (
        <Alert variant="success" onClose={() => setCreatedKey(null)}>
          <div className="space-y-2">
            <p className="font-medium text-sm sm:text-base">API Key Created Successfully!</p>
            <p className="text-xs sm:text-sm leading-relaxed">Save this key securely - you won't be able to see it again.</p>
            <div className="mt-3 p-2 sm:p-3 bg-white/50 dark:bg-gray-900/50 rounded border border-green-300 dark:border-green-700">
              <div className="flex flex-col sm:flex-row sm:items-center gap-2 sm:justify-between">
                <code className="text-xs sm:text-sm font-mono break-all flex-1">{createdKey}</code>
                <Button size="sm" variant="outline" onClick={handleCopyCreatedKey} className="self-end sm:self-auto">
                  <Copy className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        </Alert>
      )}

      {/* Create Form */}
      {showCreateForm && (
        <Card title="Create New API Key">
          {createError && <Alert variant="error" className="mb-4">{createError}</Alert>}

          <form className="space-y-4" data-testid="create-api-key-form">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
                Key Name
              </label>
              <input
                type="text"
                data-testid="api-key-name-input"
                value={newKeyData.name}
                onChange={(e) => setNewKeyData({ ...newKeyData, name: e.target.value })}
                className="w-full px-3 sm:px-4 py-2.5 sm:py-3 text-sm sm:text-base border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white min-h-[44px]"
                placeholder="Production API Key"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
                Scopes (comma-separated)
              </label>
              <input
                type="text"
                data-testid="api-key-scopes-input"
                value={newKeyData.scopes}
                onChange={(e) => setNewKeyData({ ...newKeyData, scopes: e.target.value })}
                className="w-full px-3 sm:px-4 py-2.5 sm:py-3 text-sm sm:text-base border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white min-h-[44px]"
                placeholder="read,write"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
                Expires In (days, optional)
              </label>
              <input
                type="number"
                data-testid="api-key-expiry-input"
                value={newKeyData.expiresInDays}
                onChange={(e) => setNewKeyData({ ...newKeyData, expiresInDays: e.target.value })}
                className="w-full px-3 sm:px-4 py-2.5 sm:py-3 text-sm sm:text-base border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white min-h-[44px]"
                placeholder="30"
              />
            </div>

            <div className="flex flex-col sm:flex-row gap-3">
              <Button onClick={handleCreateKey} isLoading={isCreating} disabled={!newKeyData.name || !newKeyData.scopes} data-testid="create-api-key-button" className="sm:flex-1">
                Create Key
              </Button>
              <Button variant="outline" onClick={() => setShowCreateForm(false)} data-testid="cancel-create-api-key-button" className="sm:flex-1">
                Cancel
              </Button>
            </div>
          </form>
        </Card>
      )}

      {/* API Keys List */}
      <div className="space-y-4">
        <h2 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-white">Your API Keys ({apiKeys.length})</h2>

        {apiKeys.length === 0 ? (
          <Card>
            <div className="text-center py-6 sm:py-8">
              <Key className="h-10 w-10 sm:h-12 sm:w-12 text-gray-400 mx-auto mb-3 sm:mb-4" />
              <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 px-4">No API keys yet. Create your first key to get started.</p>
            </div>
          </Card>
        ) : (
          <div className="space-y-3">
            {apiKeys.map((key) => (
              <Card key={key.id}>
                <div className="flex flex-col sm:flex-row sm:items-start gap-3 sm:gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex flex-wrap items-center gap-2 mb-2">
                      <h3 className="text-base sm:text-lg font-medium text-gray-900 dark:text-white break-words">{key.name}</h3>
                      {key.isExpired ? (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300 whitespace-nowrap">
                          <XCircle className="h-3 w-3 mr-1" />
                          Expired
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300 whitespace-nowrap">
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Active
                        </span>
                      )}
                    </div>

                    <div className="space-y-2 text-xs sm:text-sm text-gray-600 dark:text-gray-400">
                      <div className="flex items-center gap-2 flex-wrap">
                        <Key className="h-4 w-4 flex-shrink-0" />
                        <span className="font-mono break-all">{key.keyHash}...</span>
                        <button
                          onClick={() => handleCopyKey(key.keyHash)}
                          className="text-primary-600 hover:text-primary-700 dark:text-primary-400 p-1 min-h-[32px] min-w-[32px] flex items-center justify-center"
                        >
                          {copiedKeyId === Number(key.keyHash) ? <span className="text-xs whitespace-nowrap">✓ Copied</span> : <Copy className="h-4 w-4" />}
                        </button>
                      </div>

                      <div className="flex items-start gap-2">
                        <span className="font-medium whitespace-nowrap">Scopes:</span>
                        <span className="font-mono break-words">{key.scopes.join(', ')}</span>
                      </div>

                      <div className="flex items-start gap-2">
                        <Clock className="h-4 w-4 flex-shrink-0 mt-0.5" />
                        <span className="break-words">Created: {format(new Date(key.createdAt), 'PPp')}</span>
                      </div>

                      {key.expiresAt && (
                        <div className="flex items-start gap-2">
                          <Clock className="h-4 w-4 flex-shrink-0 mt-0.5" />
                          <span className="break-words">Expires: {format(new Date(key.expiresAt), 'PPp')}</span>
                        </div>
                      )}

                      {key.lastUsedAt && (
                        <div className="flex items-start gap-2">
                          <Clock className="h-4 w-4 flex-shrink-0 mt-0.5" />
                          <span className="break-words">Last used: {format(new Date(key.lastUsedAt), 'PPp')}</span>
                        </div>
                      )}
                    </div>
                  </div>

                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => revokeKey(key.id)}
                    isLoading={isRevoking}
                    className="sm:self-start"
                  >
                    <Trash2 className="h-4 w-4 mr-1" />
                    <span className="text-sm">Revoke</span>
                  </Button>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* Info */}
      <Alert variant="info">
        <div>
          <p className="font-medium mb-1 text-sm sm:text-base">About API Keys</p>
          <p className="text-xs sm:text-sm leading-relaxed">
            API keys allow programmatic access to protected endpoints. Include your key in the <code className="font-mono bg-blue-100 dark:bg-blue-900/30 px-1 py-0.5 rounded text-xs">X-API-Key</code> header or <code className="font-mono bg-blue-100 dark:bg-blue-900/30 px-1 py-0.5 rounded text-xs">Authorization: Bearer</code> header when making requests.
          </p>
        </div>
      </Alert>
    </div>
  )
}
