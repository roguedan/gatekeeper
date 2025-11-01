export interface APIKeyMetadata {
  id: number
  keyHash: string
  name: string
  scopes: string[]
  expiresAt: string | null
  lastUsedAt: string | null
  createdAt: string
  isExpired: boolean
}

export interface CreateAPIKeyRequest {
  name: string
  scopes: string[]
  expiresInSeconds?: number
}

export interface CreateAPIKeyResponse {
  key: string
  keyHash: string
  name: string
  scopes: string[]
  expiresAt: string | null
  createdAt: string
  message: string
}

export interface ListAPIKeysResponse {
  keys: APIKeyMetadata[]
}

export interface ErrorResponse {
  message: string
  code?: string
  details?: Record<string, unknown>
}

export interface ProtectedDataResponse {
  message: string
  data: Record<string, unknown>
}
