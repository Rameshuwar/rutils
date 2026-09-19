const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('http://localhost:5173');
  await page.click('#btn-measure-converter');
  
  await page.selectOption('#measure-category', 'speed');
  
  const fromUnits = await page.$$eval('#measure-from option', opts => opts.map(o => o.value));
  console.log("From units:", fromUnits);
  
  await page.selectOption('#measure-from', 'kilometers_per_hour');
  await page.selectOption('#measure-to', 'meters_per_second');
  await page.fill('#measure-value', '100');
  
  await page.click('#measure-form button[type="submit"]');
  
  await page.waitForTimeout(1000);
  
  const result = await page.textContent('#measure-result-text');
  const error = await page.textContent('#measure-status-message');
  
  console.log("Result:", result);
  console.log("Error:", error);
  
  await browser.close();
})();
