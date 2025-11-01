export interface NonceResponse {
  nonce: string
  expiresIn: number
}

export interface VerifySIWERequest {
  message: string
  signature: string
}

export interface TokenResponse {
  token: string
  expiresIn: number
  address: string
}

export interface AuthState {
  isAuthenticated: boolean
  address: string | null
  token: string | null
  isLoading: boolean
  error: string | null
}

export interface User {
  address: string
  chainId: number
  isConnected: boolean
}
