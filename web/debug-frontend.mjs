import { chromium } from '@playwright/test';

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  // Capture console messages
  page.on('console', msg => console.log(`[${msg.type()}] ${msg.text()}`));
  page.on('pageerror', error => console.error(`[ERROR] ${error.message}`));

  try {
    console.log('🔍 Navigating to http://localhost:3000...');
    await page.goto('http://localhost:3000', { waitUntil: 'load', timeout: 30000 });

    console.log('\n📊 Page Title:', await page.title());
    console.log('📊 Page URL:', page.url());

    // Wait a bit for React to render
    console.log('\n⏳ Waiting for React app to render...');
    await page.waitForTimeout(3000);

    // Check if root element is populated
    const rootContent = await page.locator('#root').innerHTML();
    console.log('\n📝 Root element content length:', rootContent.length);
    console.log('📝 First 300 chars:', rootContent.substring(0, 300));

    // Check for errors in console
    const errors = await page.evaluate(() => {
      return {
        hasReact: !!window.React,
        hasReactDOM: !!window.ReactDOM
      };
    });

    console.log('\n🔧 Page state:', errors);

    // Try to find connect wallet button
    const buttons = await page.locator('button').all();
    console.log(`\n🔘 Found ${buttons.length} buttons`);
    for (let i = 0; i < Math.min(5, buttons.length); i++) {
      const text = await buttons[i].textContent();
      console.log(`  Button ${i}: ${text?.trim()}`);
    }

    // Check for visible text
    const text = await page.locator('body').textContent();
    console.log('\n📝 Page text content (first 500 chars):');
    console.log(text?.substring(0, 500) || '(empty)');

    console.log('\n✅ Debug complete!');

  } catch (error) {
    console.error('❌ Error:', error.message);
  } finally {
    await browser.close();
  }
})();
