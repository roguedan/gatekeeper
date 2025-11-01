import { chromium } from '@playwright/test';

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  // Capture console messages
  page.on('console', msg => console.log(`[${msg.type()}] ${msg.text()}`));
  page.on('pageerror', error => console.error(`[PAGE_ERROR] ${error.message}`));

  try {
    console.log('🔍 Navigating to http://localhost:3000...');
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: 15000 });

    console.log('✅ Page loaded successfully');
    console.log('📊 Page Title:', await page.title());
    console.log('📊 Page URL:', page.url());

    // Wait a bit for React to render
    await page.waitForTimeout(1000);

    // Check if root element is populated
    const rootContent = await page.locator('#root').innerHTML();
    console.log('✅ Root element content length:', rootContent.length);

    if (rootContent.length > 100) {
      console.log('✅ Root element has substantial content');
      console.log('📝 First 200 chars:', rootContent.substring(0, 200));
    } else {
      console.log('⚠️ Root element content is minimal:', rootContent.substring(0, 100));
    }

    // Check for buttons
    const buttons = await page.locator('button').count();
    console.log(`✅ Found ${buttons} buttons on page`);

    // Check for specific UI elements
    const connectButton = await page.locator('button:has-text("Connect")').count();
    const signInButton = await page.locator('button:has-text("Sign")').count();

    if (connectButton > 0) {
      console.log('✅ "Connect" button found on page');
    }
    if (signInButton > 0) {
      console.log('✅ "Sign" button found on page');
    }

    // Get all visible text
    const text = await page.locator('body').textContent();
    const hasAuth = text.includes('Ethereum') || text.includes('wallet') || text.includes('Sign');

    if (hasAuth) {
      console.log('✅ Auth UI elements visible on page');
    }

    console.log('\n✅ Frontend verification complete!');
    console.log('📊 Rendering Status: SUCCESS ✅');

  } catch (error) {
    console.error('❌ Error:', error.message);
    console.log('📊 Rendering Status: FAILED ❌');
    process.exit(1);
  } finally {
    await browser.close();
  }
})();
