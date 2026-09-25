// persona-tests/test_ui.js
//
// Ad-hoc Playwright debugging script (NOT a spec file — not auto-run by
// `npx playwright test`). Updated for the new 2-category UI:
//   Sidebar → Conversion / Calculations
//   Tab strip → per-category tools (File, PDF Size, Measurements, Time Zones,
//               Railway | BMI, Age)
//
// Run manually with:  node test_ui.js
// (Requires both frontend @ :5173 and backend @ :8080 running.
//  Use ./dev.sh from the project root to start both.)

const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  try {
    // ---------------------------------------------------------------
    // 1. Load the app (defaults to Conversion → File)
    // ---------------------------------------------------------------
    await page.goto('http://localhost:5173');
    console.log('✅ Page loaded');

    // ---------------------------------------------------------------
    // 2. Switch to the "Measurements" tab inside the Conversion category
    //    (Old selector was `#btn-measure-converter`, which no longer exists.
    //     New selector is the tab button with data-tool="measure".)
    // ---------------------------------------------------------------
    await page.click('button[data-tool="measure"]');
    console.log('✅ Clicked Measurements tab');

    // ---------------------------------------------------------------
    // 3. Pick the "speed" category inside the Measurement Converter
    // ---------------------------------------------------------------
    await page.selectOption('#measure-category', 'speed');

    // ---------------------------------------------------------------
    // 4. Dump the available "from" units (sanity check)
    // ---------------------------------------------------------------
    const fromUnits = await page.$$eval('#measure-from option', opts =>
      opts.map(o => o.value)
    );
    console.log('From units:', fromUnits);

    // ---------------------------------------------------------------
    // 5. Perform a speed conversion: 100 km/h → m/s
    // ---------------------------------------------------------------
    await page.selectOption('#measure-from', 'kilometers_per_hour');
    await page.selectOption('#measure-to', 'meters_per_second');
    await page.fill('#measure-value', '100');

    await page.click('#measure-form button[type="submit"]');

    // ---------------------------------------------------------------
    // 6. Wait until the result box becomes visible (conversion done)
    // ---------------------------------------------------------------
    await page
      .waitForSelector('#measure-result-box:not(.hidden)', { timeout: 5000 })
      .catch(() => {});

    // ---------------------------------------------------------------
    // 7. Read the result + (only if visible) the status message
    // ---------------------------------------------------------------
    const result = await page.textContent('#measure-result-text');

    const statusVisible = await page
      .locator('#measure-status-message')
      .isVisible();
    const error = statusVisible
      ? await page.textContent('#measure-status-message')
      : '(none)';

    console.log('Result:', result);
    console.log('Error :', error);

  } catch (err) {
    console.error('❌ Script failed:', err);
  } finally {
    await browser.close();
  }
})();