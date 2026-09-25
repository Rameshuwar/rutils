import { test, expect } from '@playwright/test';

// Persona: A user who wants to convert files using our UI
test.describe('File Converter Persona', () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to the frontend page
    await page.goto('/');
  });

  test('Successful File Conversion Persona', async ({ page }) => {
    // Mock the backend API conversion endpoint to isolate UI test
    await page.route('**/convert', async route => {
      // Simulate a successful conversion response with some fake blob data
      await route.fulfill({
        status: 200,
        contentType: 'application/octet-stream',
        body: 'mocked file content'
      });
    });

    // 1. The user selects a file (e.g. dummy.csv)
    // Create a dummy file buffer to simulate file upload
    const fileBuffer = Buffer.from('id,name\n1,test');
    await page.locator('#file-input').setInputFiles({
      name: 'dummy.csv',
      mimeType: 'text/csv',
      buffer: fileBuffer
    });

    // 2. The UI should auto-detect the format as CSV
    await expect(page.locator('#detected-format-text')).toHaveText('CSV (Detected)');

    // 3. The user selects the target format they want to convert to (e.g. JSON)
    await page.locator('#to-type').selectOption('json');

    // 4. The user clicks "Convert"
    //    Note: the DOM contains multiple `button[type="submit"]` elements
    //    (one per tool form: File / PDF / Measure / Time / BMI / Age).
    //    Playwright's strict mode refuses ambiguous selectors, so we
    //    scope the locator to the File Converter's form.
    const downloadPromise = page.waitForEvent('download');
    await page.locator('#convert-form button[type="submit"]').click();

    // 5. The download should start and the success message should be shown
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toBe('converted.json');
    
    await expect(page.locator('#status-message')).toHaveText('Success! File converted to JSON.');
    await expect(page.locator('#status-message')).toHaveClass(/text-green-600/);
  });

  test('Error Persona: User tries to convert without a file', async ({ page }) => {
    // The user immediately clicks convert without selecting a file
    // (scoped to the File form to avoid strict-mode ambiguity)
    await page.locator('#convert-form button[type="submit"]').click();

    // The UI uses HTML5 'required' attribute, so the browser prevents submission.
    // Check that the file input is flagged as invalid.
    const fileInput = page.locator('#file-input');
    
    // Playwright evaluates the native validity state
    const isRequired = await fileInput.evaluate((el: HTMLInputElement) => el.validity.valueMissing);
    expect(isRequired).toBe(true);
  });

  test('Error Persona: User tries to convert to the same format', async ({ page }) => {
    // 1. The user selects a JSON file
    const fileBuffer = Buffer.from('{"id": 1, "name": "test"}');
    await page.locator('#file-input').setInputFiles({
      name: 'dummy.json',
      mimeType: 'application/json',
      buffer: fileBuffer
    });

    // 2. The UI detects JSON
    await expect(page.locator('#detected-format-text')).toHaveText('JSON (Detected)');

    // 3. The user selects the SAME target format (JSON)
    await page.locator('#to-type').selectOption('json');

    // 4. The user clicks Convert
    // (scoped to the File form to avoid strict-mode ambiguity)
    await page.locator('#convert-form button[type="submit"]').click();

    // 5. The UI should show an error message
    await expect(page.locator('#status-message')).toHaveText('Source and target formats cannot be the same.');
    await expect(page.locator('#status-message')).toHaveClass(/text-red-600/);
  });

});