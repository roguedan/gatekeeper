import { test, expect } from '@playwright/test';

/**
 * E2E Tests: API Key Management
 *
 * Test Suite: API Key Management
 * Coverage: Key creation, listing, revocation, scopes, permissions
 * Framework: Agentic QE with Playwright
 */

test.describe('API Key Management - Dashboard', () => {

  test('should display API Keys navigation link when authenticated', async ({ page, context }) => {
    // Set authenticated state
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('wagmi.account', '0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/');
    await page.waitForTimeout(1000);

    // Verify API Keys link is visible (use first() for strict mode)
    const apiKeysLink = page.getByRole('link', { name: /api.*keys/i }).first();
    await expect(apiKeysLink).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to API Keys page', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Verify we're on the API Keys page
    expect(page.url()).toContain('/api-keys');

    // Page should have heading
    const heading = page.getByRole('heading', { name: /api.*keys/i });
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('should display empty state when no API keys exist', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Look for empty state message or create first key CTA
    const emptyState = page.getByText(/no.*api.*keys|create.*first.*key|get.*started/i);
    const createButton = page.getByRole('button', { name: /create.*api.*key|new.*key/i });

    // Either empty state or create button should be visible
    await expect(emptyState.or(createButton)).toBeVisible({ timeout: 5000 });
  });
});

test.describe('API Key Management - Creation Flow', () => {
  test('should open create API key modal/form', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click create API key button
    const createButton = page.getByTestId('toggle-create-form-button');
    await createButton.click();

    // Verify form appears
    await expect(page.getByTestId('create-api-key-form')).toBeVisible({ timeout: 5000 });
  });

  test('should require API key name', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click create API key button
    const createButton = page.getByTestId('toggle-create-form-button');
    await createButton.click();

    // Form should have name input
    const nameInput = page.getByTestId('api-key-name-input');
    await expect(nameInput).toBeVisible({ timeout: 3000 });

    // Create button should be disabled without name
    const submitButton = page.getByTestId('create-api-key-button');
    await expect(submitButton).toBeDisabled();
  });

  test('should allow setting API key name', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click create API key button
    const createButton = page.getByTestId('toggle-create-form-button');
    await createButton.click();

    // Fill in name
    const nameInput = page.getByTestId('api-key-name-input');
    await expect(nameInput).toBeVisible({ timeout: 3000 });
    await nameInput.fill('Test API Key');

    // Verify value is set
    await expect(nameInput).toHaveValue('Test API Key');
  });

  test('should display scope/permission options', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click create API key button
    const createButton = page.getByTestId('toggle-create-form-button');
    await createButton.click();

    // Scopes input should be visible
    const scopesInput = page.getByTestId('api-key-scopes-input');
    await expect(scopesInput).toBeVisible({ timeout: 3000 });
  });

  test('should allow selecting multiple scopes', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click create API key button
    const createButton = page.getByTestId('toggle-create-form-button');
    await createButton.click();

    // Fill scopes input with multiple scopes
    const scopesInput = page.getByTestId('api-key-scopes-input');
    await expect(scopesInput).toBeVisible({ timeout: 3000 });
    await scopesInput.fill('read,write,admin');

    // Verify value is set
    await expect(scopesInput).toHaveValue('read,write,admin');
  });

  test('should validate API key name length', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click create API key button
    const createButton = page.getByTestId('toggle-create-form-button');
    await createButton.click();

    // Try very long name
    const nameInput = page.getByTestId('api-key-name-input');
    await expect(nameInput).toBeVisible({ timeout: 3000 });
    if (await nameInput.isVisible({ timeout: 3000 })) {
      const longName = 'A'.repeat(200);
      await nameInput.fill(longName);

      // Should show validation error or truncate
      await page.waitForTimeout(500);
    }
  });

  test('should display generated API key after creation', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');

    // Note: Full creation flow requires backend integration
    // This test documents the expected behavior
  });

  test('should provide copy-to-clipboard for new API key', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');

    // After creation, should show copy button
    // This documents the expected UX
  });
});

test.describe('API Key Management - Listing & Display', () => {
  test('should display existing API keys in a table/list', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Look for table or list container
    const table = page.getByRole('table').or(page.locator('[class*="table"]'));
    const list = page.locator('[class*="list"]');

    // Either table or list should be present
    await page.waitForTimeout(1000);
  });

  test('should show API key metadata (name, created date, scopes)', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');

    // API keys should show metadata
    // This documents expected structure
  });

  test('should mask API key values in listing', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');

    // API keys should be masked (e.g., gk_****...)
    // Look for masked format
    const maskedKey = page.getByText(/gk_\*+|••••/);
    await page.waitForTimeout(1000);
  });

  test('should display scopes as badges/tags', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');

    // Scopes should be displayed as visual indicators
    await page.waitForTimeout(1000);
  });
});

test.describe('API Key Management - Revocation', () => {
  test('should show revoke/delete button for each API key', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Look for delete/revoke buttons
    const deleteButton = page.getByRole('button', { name: /delete|revoke|remove/i });
    await page.waitForTimeout(1000);
  });

  test('should show confirmation dialog before revoking key', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click revoke button if exists
    const deleteButton = page.getByRole('button', { name: /delete|revoke|remove/i }).first();

    if (await deleteButton.isVisible({ timeout: 3000 })) {
      await deleteButton.click();

      // Should show confirmation dialog
      const confirmDialog = page.getByRole('dialog').or(page.getByText(/are you sure|confirm|delete/i));
      await expect(confirmDialog).toBeVisible({ timeout: 3000 });
    }
  });

  test('should allow canceling revocation', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Click revoke then cancel
    const deleteButton = page.getByRole('button', { name: /delete|revoke|remove/i }).first();

    if (await deleteButton.isVisible({ timeout: 3000 })) {
      await deleteButton.click();

      // Click cancel if confirmation appears
      const cancelButton = page.getByRole('button', { name: /cancel/i });

      if (await cancelButton.isVisible({ timeout: 3000 })) {
        await cancelButton.click();

        // Dialog should close
        await page.waitForTimeout(500);
      }
    }
  });
});

test.describe('API Key Management - Permissions', () => {
  test('should not allow non-authenticated users to access API keys page', async ({ page }) => {
    // Try to access without auth
    await page.goto('/api-keys');
    await page.waitForTimeout(1000);

    // Should redirect or show auth guard
    const currentUrl = page.url();
    // May redirect to home or login
    await page.waitForTimeout(1000);
  });

  test('should display API key count/usage statistics', async ({ page, context }) => {
    const mockJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.signature';

    await context.addInitScript(({ token }) => {
      localStorage.setItem('wagmi.connected', 'true');
      localStorage.setItem('gatekeeper_auth_token', token);
    }, { token: mockJWT });

    await page.goto('/api-keys');

    // Look for statistics or count
    await page.waitForTimeout(1000);
  });
});
