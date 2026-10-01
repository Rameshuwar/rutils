import { test, expect } from '@playwright/test';

test.describe('UI NIFTY 50 Markets Persona', () => {

  test('Persona: Unauthenticated user sees authentication prompt on Markets → NIFTY 50', async ({ page }) => {
    await page.goto('/');

    // Click Markets in sidebar
    const marketsBtn = page.locator('#btn-category-markets');
    await expect(marketsBtn).toBeVisible();
    await marketsBtn.click();

    // Verify category header and breadcrumb
    await expect(page.locator('#category-title')).toHaveText('Markets');
    await expect(page.locator('#breadcrumb-tool')).toHaveText('NIFTY 50');

    // Verify unauthenticated card is visible
    const unauthCard = page.locator('#nifty-unauth-card');
    await expect(unauthCard).toBeVisible();
    await expect(unauthCard.locator('h3')).toHaveText('Authentication Required');
  });

  test('Persona: Authenticated user views 50 constituents, filters, and searches', async ({ page }) => {
    await page.goto('/');

    const testEmail = `ui_tester_${Date.now()}@example.com`;
    const testPassword = 'Password123!';

    // 1. Open Sign In modal
    const signInBtn = page.locator('[data-auth-btn]:visible');
    await signInBtn.click();

    // 2. Switch to Sign Up
    await page.locator('#goto-register').click();
    await page.locator('#reg-name').fill('UI Nifty Tester');
    await page.locator('#reg-email').fill(testEmail);
    await page.locator('#reg-password').fill(testPassword);
    await page.locator('#reg-confirm').fill(testPassword);
    await page.locator('#auth-form-register button[type="submit"]').click();

    // Wait for auth modal to close
    await expect(page.locator('#auth-overlay')).toBeHidden();

    // 3. Navigate to Markets
    await page.locator('#btn-category-markets').click();

    // 4. Verify NIFTY 50 data table is loaded
    const tableCard = page.locator('#nifty-table-card');
    await expect(tableCard).toBeVisible();

    // Verify 50 rows
    const rows = page.locator('#nifty-table-body tr');
    await expect(rows).toHaveCount(50);

    // Verify meta badges
    await expect(page.locator('#nifty-meta-count')).toHaveText('50');

    // 5. Test Search Filter
    const searchInput = page.locator('#nifty-search-input');
    await searchInput.fill('RELIANCE');
    await expect(page.locator('#nifty-table-body tr')).toHaveCount(1);
    await expect(page.locator('#nifty-table-body tr').first()).toContainText('Reliance Industries');

    // Clear search
    await searchInput.fill('');
    await expect(page.locator('#nifty-table-body tr')).toHaveCount(50);

    // 6. Test Industry Filter
    const industrySelect = page.locator('#nifty-industry-select');
    await industrySelect.selectOption({ label: 'Information Technology' });
    const itRows = page.locator('#nifty-table-body tr');
    const itCount = await itRows.count();
    expect(itCount).toBeGreaterThan(0);
    expect(itCount).toBeLessThan(50);
  });

});
