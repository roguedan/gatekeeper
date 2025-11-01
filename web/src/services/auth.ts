import { SiweMessage } from 'siwe'
import { apiClient } from './api'
import { storage } from './storage'
import { env } from '@/config'
import {
  NonceResponse,
  VerifySIWERequest,
  TokenResponse,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
  ListAPIKeysResponse,
  ProtectedDataResponse
} from '@/types'

export const authService = {
  /**
   * Get a nonce for SIWE authentication
   */
  getNonce: async (): Promise<string> => {
    const response = await apiClient.get<NonceResponse>('/auth/siwe/nonce')
    return response.data.nonce
  },

  /**
   * Create a SIWE message for signing
   */
  createSiweMessage: (address: string, nonce: string, chainId: number): string => {
    const message = new SiweMessage({
      domain: env.appDomain,
      address,
      statement: `Sign in with Ethereum to ${env.appName}`,
      uri: `https://${env.appDomain}`,
      version: '1',
      chainId,
      nonce,
    })
    return message.prepareMessage()
  },

  /**
   * Verify SIWE signature and get JWT token
   */
  verifySiwe: async (message: string, signature: string): Promise<TokenResponse> => {
    const payload: VerifySIWERequest = { message, signature }
    const response = await apiClient.post<TokenResponse>('/auth/siwe/verify', payload)

    // Store token and address
    const { token, address } = response.data
    storage.setToken(token)
    storage.setAddress(address)

    return response.data
  },

  /**
   * Logout - clear stored credentials
   */
  logout: (): void => {
    storage.clear()
    window.dispatchEvent(new Event('auth:logout'))
  },

  /**
   * Check if user is authenticated
   */
  isAuthenticated: (): boolean => {
    return !!storage.getToken()
  },

  /**
   * Get current user address
   */
  getCurrentAddress: (): string | null => {
    return storage.getAddress()
  },
}

export const apiKeyService = {
  /**
   * Create a new API key
   */
  create: async (data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> => {
    const response = await apiClient.post<CreateAPIKeyResponse>('/api/keys', data)
    return response.data
  },

  /**
   * List all API keys for the current user
   */
  list: async (): Promise<ListAPIKeysResponse> => {
    const response = await apiClient.get<ListAPIKeysResponse>('/api/keys')
    return response.data
  },

  /**
   * Revoke (delete) an API key
   */
  revoke: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/keys/${id}`)
  },
}

export const protectedService = {
  /**
   * Get protected data (demo endpoint)
   */
  getData: async (): Promise<ProtectedDataResponse> => {
    const response = await apiClient.get<ProtectedDataResponse>('/api/data')
    return response.data
  },
}
