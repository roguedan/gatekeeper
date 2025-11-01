import { Page, BrowserContext } from '@playwright/test';
import { createSIWEMessage } from './api';

/**
 * Wallet Helper Functions for E2E Tests
 *
 * Purpose: Provide wallet interaction and mocking helpers
 * Framework: Agentic QE with Playwright
 */

/**
 * Mock Ethereum provider interface
 */
export interface MockEthereumProvider {
  request: (args: { method: string; params?: any[] }) => Promise<any>;
  on: (event: string, handler: (...args: any[]) => void) => void;
  removeListener: (event: string, handler: (...args: any[]) => void) => void;
  isMetaMask?: boolean;
}

/**
 * Mock wallet account for testing
 */
export const mockWalletAccount = {
  address: '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb',
  privateKey: '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80',
  chainId: 1,
};

/**
 * Additional test accounts
 */
export const testAccounts = [
  {
    address: '0x70997970C51812dc3A010C7d01b50e0d17dc79C8',
    privateKey: '0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d',
  },
  {
    address: '0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC',
    privateKey: '0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a',
  },
];

/**
 * Install mock Ethereum provider in browser context
 *
 * @param page - Playwright page
 * @param address - Mock wallet address
 * @param chainId - Chain ID
 */
export async function installMockWallet(
  page: Page,
  address: string = mockWalletAccount.address,
  chainId: number = 1
): Promise<void> {
  await page.addInitScript(
    ({ address, chainId }) => {
      // Mock Ethereum provider
      const mockProvider: any = {
        isMetaMask: true,
        chainId: `0x${chainId.toString(16)}`,
        selectedAddress: address,
        networkVersion: chainId.toString(),

        request: async ({ method, params }: { method: string; params?: any[] }) => {
          console.log('Mock wallet request:', method, params);

          switch (method) {
            case 'eth_requestAccounts':
              return [address];

            case 'eth_accounts':
              return [address];

            case 'eth_chainId':
              return `0x${chainId.toString(16)}`;

            case 'net_version':
              return chainId.toString();

            case 'personal_sign':
            case 'eth_sign':
              // Return mock signature
              return '0x' + '0'.repeat(130);

            case 'eth_signTypedData_v4':
              // Return mock signature for typed data
              return '0x' + '0'.repeat(130);

            case 'wallet_switchEthereumChain':
              return null;

            case 'wallet_addEthereumChain':
              return null;

            default:
              throw new Error(`Unsupported method: ${method}`);
          }
        },

        on: (event: string, handler: (...args: any[]) => void) => {
          console.log('Mock wallet event listener:', event);
        },

        removeListener: (event: string, handler: (...args: any[]) => void) => {
          console.log('Mock wallet remove listener:', event);
        },
      };

      // Install mock provider
      (window as any).ethereum = mockProvider;
    },
    { address, chainId }
  );
}

/**
 * Sign a SIWE message with mock wallet
 *
 * @param page - Playwright page
 * @param message - SIWE message to sign
 * @returns Promise<string> - Mock signature
 */
export async function signSIWEMessage(page: Page, message: string): Promise<string> {
  return await page.evaluate(
    async ({ message }) => {
      if (!window.ethereum) {
        throw new Error('No Ethereum provider found');
      }

      try {
        const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
        const address = accounts[0];

        const signature = await window.ethereum.request({
          method: 'personal_sign',
          params: [message, address],
        });

        return signature;
      } catch (error) {
        console.error('Error signing message:', error);
        throw error;
      }
    },
    { message }
  );
}

/**
 * Connect mock wallet
 *
 * @param page - Playwright page
 * @param address - Wallet address to connect
 * @returns Promise<string> - Connected address
 */
export async function connectMockWallet(
  page: Page,
  address: string = mockWalletAccount.address
): Promise<string> {
  return await page.evaluate(async () => {
    if (!window.ethereum) {
      throw new Error('No Ethereum provider found');
    }

    const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
    return accounts[0];
  });
}

/**
 * Disconnect wallet
 *
 * @param page - Playwright page
 */
export async function disconnectWallet(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem('wagmi.connected');
    localStorage.removeItem('wagmi.wallet');
    localStorage.removeItem('wagmi.account');
  });
}

/**
 * Get current wallet address from provider
 *
 * @param page - Playwright page
 * @returns Promise<string | null> - Current address or null
 */
export async function getCurrentAddress(page: Page): Promise<string | null> {
  return await page.evaluate(async () => {
    if (!window.ethereum) {
      return null;
    }

    const accounts = await window.ethereum.request({ method: 'eth_accounts' });
    return accounts.length > 0 ? accounts[0] : null;
  });
}

/**
 * Get current chain ID from provider
 *
 * @param page - Playwright page
 * @returns Promise<number | null> - Current chain ID or null
 */
export async function getCurrentChainId(page: Page): Promise<number | null> {
  return await page.evaluate(async () => {
    if (!window.ethereum) {
      return null;
    }

    const chainId = await window.ethereum.request({ method: 'eth_chainId' });
    return parseInt(chainId, 16);
  });
}

/**
 * Switch to different chain
 *
 * @param page - Playwright page
 * @param chainId - Target chain ID
 */
export async function switchChain(page: Page, chainId: number): Promise<void> {
  await page.evaluate(
    async ({ chainId }) => {
      if (!window.ethereum) {
        throw new Error('No Ethereum provider found');
      }

      await window.ethereum.request({
        method: 'wallet_switchEthereumChain',
        params: [{ chainId: `0x${chainId.toString(16)}` }],
      });
    },
    { chainId }
  );
}

/**
 * Simulate wallet connection rejection
 *
 * @param page - Playwright page
 */
export async function rejectWalletConnection(page: Page): Promise<void> {
  await page.addInitScript(() => {
    const mockProvider: any = {
      isMetaMask: true,

      request: async ({ method }: { method: string }) => {
        if (method === 'eth_requestAccounts') {
          throw new Error('User rejected the request');
        }
        throw new Error('Wallet not connected');
      },

      on: () => {},
      removeListener: () => {},
    };

    (window as any).ethereum = mockProvider;
  });
}

/**
 * Simulate signature rejection
 *
 * @param page - Playwright page
 */
export async function rejectSignature(page: Page): Promise<void> {
  await page.evaluate(() => {
    const originalRequest = window.ethereum?.request;

    if (window.ethereum && originalRequest) {
      window.ethereum.request = async ({ method, params }: { method: string; params?: any[] }) => {
        if (method === 'personal_sign' || method === 'eth_sign' || method === 'eth_signTypedData_v4') {
          throw new Error('User rejected signature request');
        }

        return originalRequest({ method, params });
      };
    }
  });
}

/**
 * Extract address from signature
 *
 * Note: This is a simplified version. In production, use proper signature verification.
 *
 * @param message - Original message that was signed
 * @param signature - Signature to verify
 * @returns string - Recovered address (mock implementation)
 */
export function extractAddressFromSignature(message: string, signature: string): string {
  // In a real implementation, you would use ethers.js or viem to recover the address
  // For testing purposes, we return the mock address
  return mockWalletAccount.address;
}

/**
 * Verify that a signature is valid
 *
 * @param message - Original message
 * @param signature - Signature to verify
 * @param expectedAddress - Expected signer address
 * @returns boolean - Whether signature is valid
 */
export function verifySignature(
  message: string,
  signature: string,
  expectedAddress: string
): boolean {
  // Simplified mock verification
  // In production, use proper cryptographic verification
  const recoveredAddress = extractAddressFromSignature(message, signature);
  return recoveredAddress.toLowerCase() === expectedAddress.toLowerCase();
}

/**
 * Generate mock signature for testing
 *
 * @param message - Message to sign
 * @param address - Signer address
 * @returns string - Mock signature
 */
export function generateMockSignature(message: string, address: string): string {
  // Generate a deterministic but fake signature
  const messageHash = Array.from(message)
    .reduce((hash, char) => hash + char.charCodeAt(0), 0)
    .toString(16);

  const addressHash = address.slice(2, 10);

  return `0x${messageHash}${addressHash}${'0'.repeat(122 - messageHash.length)}`;
}

/**
 * Setup wallet event listeners for testing
 *
 * @param page - Playwright page
 * @param events - Object with event handlers
 */
export async function setupWalletEventListeners(
  page: Page,
  events: {
    accountsChanged?: (accounts: string[]) => void;
    chainChanged?: (chainId: string) => void;
    disconnect?: () => void;
  }
): Promise<void> {
  await page.evaluate(
    ({ events }) => {
      if (!window.ethereum) {
        return;
      }

      if (events.accountsChanged) {
        window.ethereum.on('accountsChanged', (accounts: string[]) => {
          console.log('Accounts changed:', accounts);
        });
      }

      if (events.chainChanged) {
        window.ethereum.on('chainChanged', (chainId: string) => {
          console.log('Chain changed:', chainId);
        });
      }

      if (events.disconnect) {
        window.ethereum.on('disconnect', () => {
          console.log('Wallet disconnected');
        });
      }
    },
    { events }
  );
}

/**
 * Trigger account change event
 *
 * @param page - Playwright page
 * @param newAddress - New wallet address
 */
export async function triggerAccountChange(page: Page, newAddress: string): Promise<void> {
  await page.evaluate(
    ({ newAddress }) => {
      if (window.ethereum) {
        // Update selected address
        (window.ethereum as any).selectedAddress = newAddress;

        // Trigger event if handler exists
        const event = new CustomEvent('accountsChanged', { detail: [newAddress] });
        window.dispatchEvent(event);
      }
    },
    { newAddress }
  );
}

/**
 * Trigger chain change event
 *
 * @param page - Playwright page
 * @param newChainId - New chain ID
 */
export async function triggerChainChange(page: Page, newChainId: number): Promise<void> {
  await page.evaluate(
    ({ newChainId }) => {
      if (window.ethereum) {
        // Update chain ID
        (window.ethereum as any).chainId = `0x${newChainId.toString(16)}`;

        // Trigger event if handler exists
        const event = new CustomEvent('chainChanged', {
          detail: `0x${newChainId.toString(16)}`,
        });
        window.dispatchEvent(event);
      }
    },
    { newChainId }
  );
}

/**
 * Check if wallet is installed
 *
 * @param page - Playwright page
 * @returns Promise<boolean> - Whether wallet is available
 */
export async function isWalletInstalled(page: Page): Promise<boolean> {
  return await page.evaluate(() => {
    return typeof window.ethereum !== 'undefined';
  });
}

/**
 * Wait for wallet to be ready
 *
 * @param page - Playwright page
 * @param timeout - Maximum wait time in milliseconds
 * @returns Promise<boolean> - Whether wallet is ready
 */
export async function waitForWallet(page: Page, timeout: number = 5000): Promise<boolean> {
  try {
    await page.waitForFunction(() => typeof window.ethereum !== 'undefined', { timeout });
    return true;
  } catch {
    return false;
  }
}

// Type augmentation for window.ethereum
declare global {
  interface Window {
    ethereum?: MockEthereumProvider & {
      request: (args: { method: string; params?: any[] }) => Promise<any>;
      on: (event: string, handler: (...args: any[]) => void) => void;
      removeListener: (event: string, handler: (...args: any[]) => void) => void;
    };
  }
}
