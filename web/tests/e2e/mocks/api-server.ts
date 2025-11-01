/**
 * Mock API Server for E2E Tests
 * Intercepts and responds to backend API calls without requiring a running server
 *
 * @file Tests E2E test suite with API mocking
 * @see playwright.config.ts for integration
 */

import { Page, Route } from '@playwright/test';

interface MockAPIConfig {
  baseURL?: string;
  delay?: number;
  enableLogging?: boolean;
}

/**
 * Sample API keys for testing
 */
const mockAPIKeys = [
  {
    id: 'key_1',
    name: 'Development Key',
    key: 'gk_dev_1234567890abcdef',
    scopes: ['read', 'write'],
    createdAt: '2024-10-01T10:00:00Z',
    lastUsedAt: '2024-10-31T12:00:00Z',
  },
  {
    id: 'key_2',
    name: 'Production Key',
    key: 'gk_prod_1234567890abcdef',
    scopes: ['read'],
    createdAt: '2024-09-15T15:30:00Z',
    lastUsedAt: '2024-10-30T08:00:00Z',
  },
];

/**
 * Generate a random nonce for SIWE authentication
 */
function generateNonce(): string {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}

/**
 * Mock SIWE nonce endpoint
 * GET /auth/siwe/nonce
 */
function handleSIWENonce(route: Route): void {
  route.fulfill({
    status: 200,
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      nonce: generateNonce(),
    }),
  });
}

/**
 * Mock SIWE verify endpoint
 * POST /api/v1/auth/verify (also /auth/siwe/verify)
 */
function handleSIWEVerify(route: Route): void {
  const request = route.request();
  let postData: any = {};

  try {
    postData = request.postDataJSON() || {};
  } catch {
    // If postData is not JSON, try text
  }

  // Mock signature verification - in real scenario would validate signature
  const isValid = postData.signature && postData.message && postData.address;

  if (isValid) {
    route.fulfill({
      status: 200,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIweDc0MmQzNUNjNjYzNEM1MDNlOTI1YTNiODQ0QmM5RTc1OTVmMGJFYiIsImlhdCI6MTcyOTI4Njc5MCwiZXhwIjoxNzI5MzcwNzkwfQ.mock_signature',
        expiresIn: 86400,
        userAddress: postData.address,
      }),
    });
  } else {
    route.fulfill({
      status: 400,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'Invalid SIWE signature',
      }),
    });
  }
}

/**
 * Mock API Keys listing endpoint
 * GET /api/keys
 */
function handleAPIKeysGet(route: Route): void {
  const request = route.request();
  const authHeader = request.headerValue('Authorization');

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    route.fulfill({
      status: 401,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'Unauthorized',
      }),
    });
    return;
  }

  route.fulfill({
    status: 200,
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      data: mockAPIKeys.map(key => ({
        ...key,
        key: `${key.key.substring(0, 7)}...${key.key.substring(key.key.length - 4)}`, // Mask key
      })),
      total: mockAPIKeys.length,
    }),
  });
}

/**
 * Mock API Keys creation endpoint
 * POST /api/keys
 */
function handleAPIKeysCreate(route: Route): void {
  const request = route.request();
  const authHeader = request.headerValue('Authorization');
  let postData: any = {};

  try {
    postData = request.postDataJSON() || {};
  } catch {
    // If postData is not JSON
  }

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    route.fulfill({
      status: 401,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'Unauthorized',
      }),
    });
    return;
  }

  if (!postData.name) {
    route.fulfill({
      status: 400,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'API key name is required',
      }),
    });
    return;
  }

  const newKey = {
    id: `key_${Date.now()}`,
    name: postData.name,
    key: `gk_${Math.random().toString(36).substring(2, 24)}`,
    scopes: postData.scopes || ['read'],
    createdAt: new Date().toISOString(),
    lastUsedAt: null,
  };

  route.fulfill({
    status: 201,
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      data: newKey,
    }),
  });
}

/**
 * Mock API Keys revocation endpoint
 * DELETE /api/keys/:id
 */
function handleAPIKeysDelete(route: Route, keyId: string): void {
  const request = route.request();
  const authHeader = request.headerValue('Authorization');

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    route.fulfill({
      status: 401,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'Unauthorized',
      }),
    });
    return;
  }

  const keyExists = mockAPIKeys.some(k => k.id === keyId);

  if (!keyExists) {
    route.fulfill({
      status: 404,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'API key not found',
      }),
    });
    return;
  }

  route.fulfill({
    status: 204,
    headers: {
      'Content-Type': 'application/json',
    },
  });
}

/**
 * Mock health check endpoint
 * GET /health
 */
function handleHealth(route: Route): void {
  route.fulfill({
    status: 200,
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      status: 'ok',
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      checks: {
        database: {
          status: 'ok',
          responseTime: 15,
        },
        ethereum: {
          status: 'ok',
          responseTime: 42,
        },
      },
    }),
  });
}

/**
 * Setup API mocking for E2E tests
 *
 * @param page Playwright page object
 * @param config Optional configuration
 *
 * @example
 * ```typescript
 * test('API key management', async ({ page }) => {
 *   await setupAPIMocks(page);
 *   // ... rest of test
 * });
 * ```
 */
export async function setupAPIMocks(
  page: Page,
  config: MockAPIConfig = {},
): Promise<void> {
  const {
    baseURL = 'http://localhost:8080',
    delay = 100,
    enableLogging = false,
  } = config;

  // Intercept all requests to /auth/siwe/nonce
  await page.route(`${baseURL}/auth/siwe/nonce`, (route) => {
    if (enableLogging) console.log('Mocking: /auth/siwe/nonce');
    setTimeout(() => handleSIWENonce(route), delay);
  });

  // Intercept all requests to /api/v1/auth/verify
  await page.route(`${baseURL}/api/v1/auth/verify`, (route) => {
    if (enableLogging) console.log('Mocking: /api/v1/auth/verify');
    setTimeout(() => handleSIWEVerify(route), delay);
  });

  // Intercept all requests to /auth/siwe/verify
  await page.route(`${baseURL}/auth/siwe/verify`, (route) => {
    if (enableLogging) console.log('Mocking: /auth/siwe/verify');
    setTimeout(() => handleSIWEVerify(route), delay);
  });

  // Intercept GET /api/keys
  await page.route(`${baseURL}/api/keys`, (route) => {
    if (route.request().method() === 'GET') {
      if (enableLogging) console.log('Mocking: GET /api/keys');
      setTimeout(() => handleAPIKeysGet(route), delay);
    } else if (route.request().method() === 'POST') {
      if (enableLogging) console.log('Mocking: POST /api/keys');
      setTimeout(() => handleAPIKeysCreate(route), delay);
    }
  });

  // Intercept DELETE /api/keys/:id
  await page.route(`${baseURL}/api/keys/*`, (route) => {
    if (route.request().method() === 'DELETE') {
      const url = new URL(route.request().url());
      const keyId = url.pathname.split('/').pop();
      if (enableLogging) console.log(`Mocking: DELETE /api/keys/${keyId}`);
      setTimeout(() => handleAPIKeysDelete(route, keyId!), delay);
    }
  });

  // Intercept GET /health
  await page.route(`${baseURL}/health`, (route) => {
    if (enableLogging) console.log('Mocking: /health');
    setTimeout(() => handleHealth(route), delay);
  });
}

/**
 * Remove all API mocks
 * Useful for cleanup or per-test configuration
 *
 * @param page Playwright page object
 */
export async function removeAPIMocks(page: Page): Promise<void> {
  await page.unroute('**/*');
}

/**
 * Create a new API key in the mock store
 * Used to set up test data
 *
 * @param key API key details
 *
 * @example
 * ```typescript
 * addMockAPIKey({
 *   id: 'test_key',
 *   name: 'Test Key',
 *   key: 'gk_test123',
 *   scopes: ['read', 'write'],
 *   createdAt: new Date().toISOString(),
 * });
 * ```
 */
export function addMockAPIKey(key: typeof mockAPIKeys[0]): void {
  mockAPIKeys.push(key);
}

/**
 * Clear all mock API keys
 * Useful for test cleanup
 */
export function clearMockAPIKeys(): void {
  mockAPIKeys.length = 0;
}

/**
 * Get all mock API keys
 * Useful for assertions in tests
 */
export function getMockAPIKeys(): typeof mockAPIKeys {
  return [...mockAPIKeys];
}
