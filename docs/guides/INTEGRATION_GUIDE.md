# Frontend-Backend Integration Guide

Complete guide for integrating the Gatekeeper React frontend with the Go backend API.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [API Authentication Flow](#api-authentication-flow)
- [SIWE Verification Process](#siwe-verification-process)
- [Protected Endpoint Flow](#protected-endpoint-flow)
- [API Key Integration](#api-key-integration)
- [Error Handling](#error-handling)
- [Real-World Examples](#real-world-examples)
- [Debugging Guide](#debugging-guide)
- [Network Troubleshooting](#network-troubleshooting)

---

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         Browser / Wallet                         │
│  ┌────────────┐      ┌────────────┐      ┌────────────────┐    │
│  │  MetaMask  │      │ Coinbase   │      │ WalletConnect  │    │
│  └────────────┘      └────────────┘      └────────────────┘    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               │ Sign-In with Ethereum (EIP-4361)
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                      React Frontend (Port 3000)                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  wagmi + viem (Ethereum Integration)                      │  │
│  │  ├─ Wallet connection                                     │  │
│  │  ├─ Sign SIWE message                                     │  │
│  │  └─ Read blockchain data                                  │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  API Client (Axios)                                       │  │
│  │  ├─ JWT token management                                  │  │
│  │  ├─ Request/response interceptors                         │  │
│  │  └─ Error handling                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               │ HTTP/JSON API
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                      Go Backend (Port 8080)                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Authentication Layer                                     │  │
│  │  ├─ SIWE verification                                     │  │
│  │  ├─ JWT generation/validation                             │  │
│  │  └─ API key validation                                    │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  Middleware Chain                                         │  │
│  │  ├─ CORS                                                  │  │
│  │  ├─ Rate limiting                                         │  │
│  │  ├─ Policy enforcement                                    │  │
│  │  └─ Audit logging                                         │  │
│  ├──────────────────────────────────────────────────────────┤  │
│  │  Business Logic                                           │  │
│  │  ├─ User management                                       │  │
│  │  ├─ API key management                                    │  │
│  │  └─ Policy evaluation                                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
                    ┌────────────────────┐
                    │   PostgreSQL DB    │
                    └────────────────────┘
```

### Technology Stack

**Frontend:**
- React 18 with TypeScript
- wagmi/viem for Ethereum
- RainbowKit for wallet UI
- Axios for HTTP requests
- TanStack Query for data fetching

**Backend:**
- Go 1.21+
- Gorilla Mux router
- JWT authentication
- PostgreSQL database
- Ethereum RPC integration

---

## API Authentication Flow

### 1. SIWE Sign-In Flow

The complete authentication flow using Sign-In with Ethereum (SIWE):

```typescript
// Frontend: Complete SIWE authentication flow

import { useAccount, useSignMessage } from 'wagmi';
import { SiweMessage } from 'siwe';
import axios from 'axios';

async function signInWithEthereum() {
  const { address, chainId } = useAccount();

  // Step 1: Get nonce from backend
  const { data } = await axios.get('http://localhost:8080/auth/siwe/nonce');
  const nonce = data.nonce;

  // Step 2: Create SIWE message
  const message = new SiweMessage({
    domain: window.location.host,
    address: address,
    statement: 'Sign in to Gatekeeper',
    uri: window.location.origin,
    version: '1',
    chainId: chainId,
    nonce: nonce,
    issuedAt: new Date().toISOString(),
  });

  // Step 3: Sign message with wallet
  const { signMessageAsync } = useSignMessage();
  const signature = await signMessageAsync({
    message: message.prepareMessage(),
  });

  // Step 4: Verify signature and get JWT
  const response = await axios.post('http://localhost:8080/auth/siwe/verify', {
    message: message.prepareMessage(),
    signature: signature,
  });

  const { token, expiresIn, address: verifiedAddress } = response.data;

  // Step 5: Store JWT token
  localStorage.setItem('auth_token', token);
  localStorage.setItem('token_expiry', Date.now() + expiresIn * 1000);
  localStorage.setItem('user_address', verifiedAddress);

  return { token, address: verifiedAddress };
}
```

**Backend handling:**

```go
// Backend: SIWE nonce generation
// GET /auth/siwe/nonce
func (s *Server) handleGetNonce(w http.ResponseWriter, r *http.Request) {
    nonce, err := s.siweService.GenerateNonce(r.Context())
    if err != nil {
        http.Error(w, "Failed to generate nonce", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "nonce":     nonce,
        "expiresIn": int(s.config.NonceTTL.Seconds()),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Backend: SIWE verification
// POST /auth/siwe/verify
func (s *Server) handleVerifySIWE(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Message   string `json:"message"`
        Signature string `json:"signature"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Verify SIWE message and signature
    address, err := s.siweService.VerifySignature(r.Context(), req.Message, req.Signature)
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // Generate JWT token
    token, err := s.jwtService.GenerateToken(r.Context(), address, []string{})
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "token":     token,
        "expiresIn": int(s.config.JWTExpiry.Seconds()),
        "address":   address,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### 2. JWT Token Management

**Frontend JWT handling:**

```typescript
// api/client.ts - Axios instance with JWT

import axios from 'axios';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: Add JWT to all requests
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor: Handle 401 and refresh token
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // If 401 and not already retrying
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      // Clear invalid token
      localStorage.removeItem('auth_token');
      localStorage.removeItem('token_expiry');

      // Redirect to sign in
      window.location.href = '/signin';
    }

    return Promise.reject(error);
  }
);

export default apiClient;
```

**Backend JWT validation:**

```go
// Middleware: JWT authentication
func JWTMiddleware(jwtService *auth.JWTService) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            // Remove "Bearer " prefix
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
                return
            }

            // Validate token
            claims, err := jwtService.ValidateToken(r.Context(), tokenString)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Add claims to context
            ctx := context.WithValue(r.Context(), "claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

---

## SIWE Verification Process

### Message Format

SIWE messages follow EIP-4361 standard:

```
example.com wants you to sign in with your Ethereum account:
0x1234567890123456789012345678901234567890

Sign in to Gatekeeper

URI: https://example.com
Version: 1
Chain ID: 1
Nonce: aBcDeFgHiJkLmNoPqRsTuVwXyZ
Issued At: 2025-11-01T12:00:00.000Z
Expiration Time: 2025-11-01T12:15:00.000Z
```

### Frontend SIWE Implementation

```typescript
// hooks/useSIWE.ts

import { useAccount, useSignMessage } from 'wagmi';
import { SiweMessage } from 'siwe';
import { useState } from 'react';
import apiClient from '../api/client';

export function useSIWE() {
  const { address, chainId } = useAccount();
  const { signMessageAsync } = useSignMessage();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const signIn = async () => {
    if (!address || !chainId) {
      setError('Wallet not connected');
      return null;
    }

    setIsLoading(true);
    setError(null);

    try {
      // 1. Get nonce
      const { data: nonceData } = await apiClient.get('/auth/siwe/nonce');

      // 2. Create SIWE message
      const siweMessage = new SiweMessage({
        domain: window.location.host,
        address,
        statement: 'Sign in to Gatekeeper',
        uri: window.location.origin,
        version: '1',
        chainId,
        nonce: nonceData.nonce,
        issuedAt: new Date().toISOString(),
        expirationTime: new Date(Date.now() + 15 * 60 * 1000).toISOString(), // 15 min
      });

      const message = siweMessage.prepareMessage();

      // 3. Sign message
      const signature = await signMessageAsync({ message });

      // 4. Verify and get JWT
      const { data: authData } = await apiClient.post('/auth/siwe/verify', {
        message,
        signature,
      });

      // 5. Store authentication
      localStorage.setItem('auth_token', authData.token);
      localStorage.setItem('token_expiry', String(Date.now() + authData.expiresIn * 1000));
      localStorage.setItem('user_address', authData.address);

      return authData;
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || err.message || 'Sign-in failed';
      setError(errorMessage);
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  const signOut = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('token_expiry');
    localStorage.removeItem('user_address');
  };

  return { signIn, signOut, isLoading, error };
}
```

### Backend SIWE Verification

```go
// internal/auth/siwe.go

import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common/hexutil"
)

func (s *SIWEService) VerifySignature(ctx context.Context, message, signature string) (string, error) {
    // Parse SIWE message
    siweMsg, err := parseSIWEMessage(message)
    if err != nil {
        return "", fmt.Errorf("invalid SIWE message: %w", err)
    }

    // Verify nonce is valid and not expired
    if !s.ValidateNonce(ctx, siweMsg.Nonce) {
        return "", errors.New("invalid or expired nonce")
    }

    // Verify message hasn't expired
    if time.Now().After(siweMsg.ExpirationTime) {
        return "", errors.New("message expired")
    }

    // Decode signature
    sigBytes, err := hexutil.Decode(signature)
    if err != nil {
        return "", fmt.Errorf("invalid signature format: %w", err)
    }

    // Ethereum signatures have v at the end, adjust it
    if sigBytes[64] >= 27 {
        sigBytes[64] -= 27
    }

    // Hash the message
    hash := crypto.Keccak256Hash([]byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(message)) + message))

    // Recover public key from signature
    pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
    if err != nil {
        return "", fmt.Errorf("failed to recover public key: %w", err)
    }

    // Get address from public key
    recoveredAddress := crypto.PubkeyToAddress(*pubKey).Hex()

    // Verify it matches the address in the message
    if !strings.EqualFold(recoveredAddress, siweMsg.Address) {
        return "", errors.New("signature does not match address")
    }

    // Invalidate nonce
    s.InvalidateNonce(ctx, siweMsg.Nonce)

    return recoveredAddress, nil
}
```

---

## Protected Endpoint Flow

### Frontend: Accessing Protected Endpoints

```typescript
// hooks/useProtectedData.ts

import { useQuery } from '@tanstack/react-query';
import apiClient from '../api/client';

interface ProtectedData {
  message: string;
  address: string;
  data?: any;
}

export function useProtectedData() {
  return useQuery<ProtectedData>({
    queryKey: ['protectedData'],
    queryFn: async () => {
      const response = await apiClient.get('/api/data');
      return response.data;
    },
    // Only fetch if authenticated
    enabled: !!localStorage.getItem('auth_token'),
    // Retry on 401
    retry: (failureCount, error: any) => {
      if (error.response?.status === 401) return false;
      return failureCount < 3;
    },
  });
}

// Component usage
function Dashboard() {
  const { data, isLoading, error } = useProtectedData();

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <h1>Protected Data</h1>
      <p>Message: {data?.message}</p>
      <p>Your Address: {data?.address}</p>
    </div>
  );
}
```

### Backend: Protected Route Handler

```go
// Protected endpoint with JWT middleware
func (s *Server) setupProtectedRoutes(router *mux.Router) {
    // Create protected subrouter
    apiRouter := router.PathPrefix("/api").Subrouter()

    // Apply middleware chain
    apiRouter.Use(mux.MiddlewareFunc(s.jwtMiddleware))
    apiRouter.Use(mux.MiddlewareFunc(s.rateLimitMiddleware))

    // Protected endpoints
    apiRouter.HandleFunc("/data", s.handleProtectedData).Methods("GET")
    apiRouter.HandleFunc("/keys", s.handleListAPIKeys).Methods("GET")
    apiRouter.HandleFunc("/keys", s.handleCreateAPIKey).Methods("POST")
}

func (s *Server) handleProtectedData(w http.ResponseWriter, r *http.Request) {
    // Extract claims from context (added by JWT middleware)
    claims := r.Context().Value("claims").(*auth.Claims)

    response := map[string]interface{}{
        "message": "Access granted",
        "address": claims.Address,
        "data":    "This is protected data",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

---

## API Key Integration

### Frontend: API Key Management

```typescript
// services/apiKeys.ts

import apiClient from './client';

export interface APIKey {
  id: string;
  name: string;
  key?: string; // Only present on creation
  createdAt: string;
  lastUsed?: string;
  rateLimit: number;
  policies: string[];
}

export async function createAPIKey(name: string, policies: string[] = []): Promise<APIKey> {
  const response = await apiClient.post('/api/keys', {
    name,
    policies,
  });
  return response.data;
}

export async function listAPIKeys(): Promise<APIKey[]> {
  const response = await apiClient.get('/api/keys');
  return response.data.keys || [];
}

export async function revokeAPIKey(id: string): Promise<void> {
  await apiClient.delete(`/api/keys/${id}`);
}

// Component: API Key List
function APIKeyManager() {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [newKeyName, setNewKeyName] = useState('');
  const [createdKey, setCreatedKey] = useState<APIKey | null>(null);

  useEffect(() => {
    loadKeys();
  }, []);

  const loadKeys = async () => {
    const data = await listAPIKeys();
    setKeys(data);
  };

  const handleCreate = async () => {
    const newKey = await createAPIKey(newKeyName);
    setCreatedKey(newKey);
    setNewKeyName('');
    await loadKeys();
  };

  const handleRevoke = async (id: string) => {
    await revokeAPIKey(id);
    await loadKeys();
  };

  return (
    <div>
      <h2>API Keys</h2>

      {/* Show new key once (can't be retrieved again) */}
      {createdKey && (
        <div className="alert alert-warning">
          <strong>Save this API key - it won't be shown again!</strong>
          <code>{createdKey.key}</code>
        </div>
      )}

      {/* Create new key */}
      <div>
        <input
          value={newKeyName}
          onChange={(e) => setNewKeyName(e.target.value)}
          placeholder="Key name"
        />
        <button onClick={handleCreate}>Create API Key</button>
      </div>

      {/* List existing keys */}
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Created</th>
            <th>Last Used</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {keys.map((key) => (
            <tr key={key.id}>
              <td>{key.name}</td>
              <td>{new Date(key.createdAt).toLocaleDateString()}</td>
              <td>{key.lastUsed ? new Date(key.lastUsed).toLocaleDateString() : 'Never'}</td>
              <td>
                <button onClick={() => handleRevoke(key.id)}>Revoke</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Using API Keys (Alternative to JWT)

```typescript
// Alternative authentication with API key
const apiKeyClient = axios.create({
  baseURL: 'http://localhost:8080',
  headers: {
    'X-API-Key': 'your-api-key-here',
  },
});

// Use API key for server-to-server calls
const response = await apiKeyClient.get('/api/data');
```

---

## Error Handling

### Frontend Error Handling

```typescript
// utils/errorHandler.ts

export interface APIError {
  message: string;
  code?: string;
  status?: number;
  details?: any;
}

export function handleAPIError(error: any): APIError {
  if (error.response) {
    // Server responded with error status
    return {
      message: error.response.data?.error || error.response.data?.message || 'Server error',
      code: error.response.data?.code,
      status: error.response.status,
      details: error.response.data,
    };
  } else if (error.request) {
    // Request made but no response
    return {
      message: 'No response from server. Please check your connection.',
      code: 'NETWORK_ERROR',
    };
  } else {
    // Error setting up request
    return {
      message: error.message || 'Unknown error occurred',
      code: 'CLIENT_ERROR',
    };
  }
}

// Hook for error display
export function useErrorHandler() {
  const [error, setError] = useState<APIError | null>(null);

  const handleError = (err: any) => {
    const apiError = handleAPIError(err);
    setError(apiError);

    // Auto-clear after 5 seconds
    setTimeout(() => setError(null), 5000);
  };

  const clearError = () => setError(null);

  return { error, handleError, clearError };
}

// Component usage
function MyComponent() {
  const { error, handleError, clearError } = useErrorHandler();

  const fetchData = async () => {
    try {
      const response = await apiClient.get('/api/data');
      // Handle success
    } catch (err) {
      handleError(err);
    }
  };

  return (
    <div>
      {error && (
        <div className="alert alert-danger">
          <button onClick={clearError}>×</button>
          <strong>Error:</strong> {error.message}
          {error.status && <span> (Status: {error.status})</span>}
        </div>
      )}
      {/* Rest of component */}
    </div>
  );
}
```

### Backend Error Responses

```go
// Standardized error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Details any    `json:"details,omitempty"`
}

func sendError(w http.ResponseWriter, statusCode int, message, code string, details any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)

    response := ErrorResponse{
        Error:   message,
        Code:    code,
        Details: details,
    }

    json.NewEncoder(w).Encode(response)
}

// Usage examples
func (s *Server) handleExample(w http.ResponseWriter, r *http.Request) {
    // Validation error
    if invalidInput {
        sendError(w, http.StatusBadRequest, "Invalid input", "INVALID_INPUT", map[string]string{
            "field": "address",
            "issue": "must be valid Ethereum address",
        })
        return
    }

    // Authentication error
    if !authenticated {
        sendError(w, http.StatusUnauthorized, "Authentication required", "AUTH_REQUIRED", nil)
        return
    }

    // Not found
    if !found {
        sendError(w, http.StatusNotFound, "Resource not found", "NOT_FOUND", nil)
        return
    }

    // Rate limit
    if rateLimited {
        sendError(w, http.StatusTooManyRequests, "Rate limit exceeded", "RATE_LIMIT", map[string]int{
            "limit": 100,
            "retryAfter": 60,
        })
        return
    }

    // Server error
    if internalError != nil {
        s.logger.Error("Internal error", zap.Error(internalError))
        sendError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR", nil)
        return
    }
}
```

---

## Real-World Examples

### Example 1: Token-Gated Content

```typescript
// Frontend: Token-gated page component

import { useAccount } from 'wagmi';
import { useQuery } from '@tanstack/react-query';
import apiClient from '../api/client';

function TokenGatedContent() {
  const { address } = useAccount();

  // Fetch protected content
  const { data, isLoading, error } = useQuery({
    queryKey: ['tokenGatedData', address],
    queryFn: async () => {
      const response = await apiClient.get('/api/data', {
        headers: {
          'X-Policy': 'token-gate', // Optional: specify policy
        },
      });
      return response.data;
    },
    enabled: !!address && !!localStorage.getItem('auth_token'),
    retry: false,
  });

  if (!address) {
    return <div>Please connect your wallet</div>;
  }

  if (!localStorage.getItem('auth_token')) {
    return <div>Please sign in</div>;
  }

  if (isLoading) {
    return <div>Checking access...</div>;
  }

  if (error) {
    const err = error as any;
    if (err.response?.status === 403) {
      return (
        <div>
          <h2>Access Denied</h2>
          <p>You don't own the required tokens to access this content.</p>
          <p>Required: {err.response.data?.details?.required}</p>
        </div>
      );
    }
    return <div>Error: {err.message}</div>;
  }

  return (
    <div>
      <h1>Exclusive Content</h1>
      <p>Welcome, token holder!</p>
      <pre>{JSON.stringify(data, null, 2)}</pre>
    </div>
  );
}
```

**Backend policy enforcement:**

```go
// Policy middleware checks token ownership
func (m *PolicyMiddleware) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("claims").(*auth.Claims)

            // Get policy from header or default
            policyName := r.Header.Get("X-Policy")
            if policyName == "" {
                policyName = "default"
            }

            // Evaluate policy
            result, err := m.policyManager.Evaluate(r.Context(), policyName, claims.Address)
            if err != nil {
                sendError(w, http.StatusInternalServerError, "Policy evaluation failed", "POLICY_ERROR", nil)
                return
            }

            if !result.Allowed {
                sendError(w, http.StatusForbidden, "Access denied", "ACCESS_DENIED", map[string]interface{}{
                    "required": result.Requirements,
                    "reason":   result.Reason,
                })
                return
            }

            // Access granted
            next.ServeHTTP(w, r)
        })
    }
}
```

### Example 2: API Key Creation Flow

```typescript
// Complete API key creation flow with error handling

import { useState } from 'react';
import apiClient from '../api/client';

function CreateAPIKey() {
  const [name, setName] = useState('');
  const [policies, setPolicies] = useState<string[]>([]);
  const [rateLimit, setRateLimit] = useState(100);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleCreate = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await apiClient.post('/api/keys', {
        name,
        policies,
        rateLimit,
      });

      setCreatedKey(response.data.key);
      setName('');
      setPolicies([]);
    } catch (err: any) {
      if (err.response?.status === 429) {
        setError('Rate limit exceeded. Please try again later.');
      } else if (err.response?.status === 400) {
        setError(err.response.data.error || 'Invalid input');
      } else {
        setError('Failed to create API key');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = () => {
    if (createdKey) {
      navigator.clipboard.writeText(createdKey);
      alert('API key copied to clipboard!');
    }
  };

  return (
    <div>
      {createdKey ? (
        <div className="success-message">
          <h3>API Key Created!</h3>
          <p className="warning">
            ⚠️ Save this key securely - it won't be shown again!
          </p>
          <div className="key-display">
            <code>{createdKey}</code>
            <button onClick={copyToClipboard}>Copy</button>
          </div>
          <button onClick={() => setCreatedKey(null)}>Create Another</button>
        </div>
      ) : (
        <form onSubmit={(e) => { e.preventDefault(); handleCreate(); }}>
          <div>
            <label>Key Name:</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              placeholder="My API Key"
            />
          </div>

          <div>
            <label>Rate Limit (requests/minute):</label>
            <input
              type="number"
              value={rateLimit}
              onChange={(e) => setRateLimit(Number(e.target.value))}
              min={1}
              max={1000}
            />
          </div>

          <div>
            <label>Policies:</label>
            <select
              multiple
              value={policies}
              onChange={(e) => setPolicies(Array.from(e.target.selectedOptions, o => o.value))}
            >
              <option value="token-gate">Token Gate</option>
              <option value="nft-holder">NFT Holder</option>
              <option value="allowlist">Allowlist</option>
            </select>
          </div>

          {error && <div className="error">{error}</div>}

          <button type="submit" disabled={isLoading || !name}>
            {isLoading ? 'Creating...' : 'Create API Key'}
          </button>
        </form>
      )}
    </div>
  );
}
```

---

## Debugging Guide

### Frontend Debugging

**1. Check Wallet Connection:**
```typescript
import { useAccount } from 'wagmi';

function Debug() {
  const { address, isConnected, chain } = useAccount();

  console.log('Wallet Debug:', {
    address,
    isConnected,
    chainId: chain?.id,
    chainName: chain?.name,
  });

  return <pre>{JSON.stringify({ address, isConnected, chain }, null, 2)}</pre>;
}
```

**2. Check Authentication State:**
```typescript
function AuthDebug() {
  const token = localStorage.getItem('auth_token');
  const expiry = localStorage.getItem('token_expiry');
  const address = localStorage.getItem('user_address');

  const isExpired = expiry ? Date.now() > Number(expiry) : true;

  console.log('Auth Debug:', {
    hasToken: !!token,
    isExpired,
    expiresIn: expiry ? Number(expiry) - Date.now() : 0,
    address,
  });

  return (
    <div>
      <p>Token: {token ? '✅ Present' : '❌ Missing'}</p>
      <p>Expired: {isExpired ? '❌ Yes' : '✅ No'}</p>
      <p>Address: {address || 'None'}</p>
    </div>
  );
}
```

**3. Network Request Debugging:**
```typescript
// Add detailed logging to Axios
apiClient.interceptors.request.use(
  (config) => {
    console.log('API Request:', {
      method: config.method?.toUpperCase(),
      url: config.url,
      headers: config.headers,
      data: config.data,
    });
    return config;
  },
  (error) => {
    console.error('Request Error:', error);
    return Promise.reject(error);
  }
);

apiClient.interceptors.response.use(
  (response) => {
    console.log('API Response:', {
      status: response.status,
      url: response.config.url,
      data: response.data,
    });
    return response;
  },
  (error) => {
    console.error('Response Error:', {
      status: error.response?.status,
      url: error.config?.url,
      data: error.response?.data,
      message: error.message,
    });
    return Promise.reject(error);
  }
);
```

### Backend Debugging

**1. Enable Debug Logging:**
```go
// Set log level to debug
logger, err := log.New("debug")

// Log all requests
func loggingMiddleware(logger *log.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger.Debug("Request",
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.String("remote", r.RemoteAddr),
                zap.String("user-agent", r.UserAgent()),
            )
            next.ServeHTTP(w, r)
        })
    }
}
```

**2. Test Endpoints with curl:**
```bash
# Get nonce
curl -v http://localhost:8080/auth/siwe/nonce

# Verify SIWE (with message and signature)
curl -v -X POST http://localhost:8080/auth/siwe/verify \
  -H "Content-Type: application/json" \
  -d '{"message":"...","signature":"..."}'

# Access protected endpoint
curl -v http://localhost:8080/api/data \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Create API key
curl -v -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Key"}'
```

**3. Database Inspection:**
```bash
# Connect to database
psql $DATABASE_URL

# Check users
SELECT * FROM users WHERE address = '0x...';

# Check API keys
SELECT id, name, created_at, last_used FROM api_keys WHERE user_id = '...';

# Check nonces
SELECT * FROM siwe_nonces WHERE created_at > NOW() - INTERVAL '15 minutes';
```

---

## Network Troubleshooting

### CORS Issues

**Symptom:** Browser blocks requests with CORS error

**Solution - Backend CORS configuration:**
```go
import "github.com/rs/cors"

func setupCORS(router *mux.Router) http.Handler {
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000", "https://yourdomain.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Policy"},
        ExposedHeaders:   []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           300,
    })

    return c.Handler(router)
}
```

### Connection Refused

**Symptom:** `ERR_CONNECTION_REFUSED` or `ECONNREFUSED`

**Checklist:**
```bash
# 1. Check backend is running
ps aux | grep server

# 2. Check port is listening
lsof -i :8080

# 3. Check firewall
sudo ufw status

# 4. Check environment variable
echo $VITE_API_URL  # Should be http://localhost:8080

# 5. Test with curl
curl http://localhost:8080/health
```

### JWT Expiry Issues

**Symptom:** Repeated 401 errors, forced re-login

**Debug:**
```typescript
// Check token expiry
const checkTokenExpiry = () => {
  const expiry = localStorage.getItem('token_expiry');
  if (!expiry) {
    console.log('No token expiry set');
    return;
  }

  const expiryTime = Number(expiry);
  const now = Date.now();
  const timeLeft = expiryTime - now;

  console.log({
    expiresAt: new Date(expiryTime).toISOString(),
    timeLeftMs: timeLeft,
    timeLeftMin: Math.floor(timeLeft / 1000 / 60),
    isExpired: timeLeft < 0,
  });
};

// Call before API requests
checkTokenExpiry();
```

### Rate Limiting

**Symptom:** 429 Too Many Requests

**Debug:**
```bash
# Backend logs will show rate limit info
# Check current limits
curl http://localhost:8080/api/data -I | grep X-RateLimit

# Expected headers:
# X-RateLimit-Limit: 100
# X-RateLimit-Remaining: 95
# X-RateLimit-Reset: 1234567890
```

---

## Additional Resources

### Tools
- [MetaMask](https://metamask.io/) - Browser wallet
- [WalletConnect](https://walletconnect.com/) - Mobile wallet bridge
- [Tenderly](https://tenderly.co/) - Transaction debugging
- [Postman](https://www.postman.com/) - API testing

### Documentation
- [EIP-4361 (SIWE)](https://eips.ethereum.org/EIPS/eip-4361)
- [wagmi Documentation](https://wagmi.sh/)
- [Axios Documentation](https://axios-http.com/)
- [JWT.io](https://jwt.io/) - JWT debugging

### Support
- GitHub Issues: https://github.com/roguedan/gatekeeper/issues
- Discord: (add your Discord link)
- Documentation: https://docs.gatekeeper.example.com

---

**Last Updated:** November 1, 2025
**Version:** 1.0.0
