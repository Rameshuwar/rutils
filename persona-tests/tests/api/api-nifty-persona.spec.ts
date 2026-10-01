import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8080';

test.describe('Backend API NIFTY 50 Market Data Persona', () => {

  test('Persona: Unauthenticated request is rejected with 401 Unauthorized', async ({ request }) => {
    await test.step('Why we use this test: Verify that NIFTY 50 constituents data is strictly protected by JWT authentication.', async () => {});

    let response;
    await test.step('What we use: GET /nifty50/companies without Authorization header.', async () => {
      response = await request.get(`${API_URL}/nifty50/companies`);
    });

    await test.step('What we expected: 401 Unauthorized.', async () => {
      expect(response.status()).toBe(401);
    });

    await test.step('What we get: Error message indicating authorization header is required.', async () => {
      const json = await response.json();
      expect(json.error).toBeDefined();
    });
  });

  test('Persona: Authenticated request returns 50 NIFTY 50 constituent companies', async ({ request }) => {
    await test.step('Why we use this test: An authenticated investor requests the official NIFTY 50 constituent companies.', async () => {});

    const testEmail = `nifty_tester_${Date.now()}@example.com`;
    const testPassword = 'Password123!';

    // 1. Register
    await request.post(`${API_URL}/auth/register`, {
      data: {
        name: 'NIFTY Investor',
        email: testEmail,
        password: testPassword,
        confirm_password: testPassword,
      },
    });

    // 2. Login
    const loginRes = await request.post(`${API_URL}/auth/login`, {
      data: {
        email: testEmail,
        password: testPassword,
      },
    });
    expect(loginRes.status()).toBe(200);
    const loginData = await loginRes.json();
    const token = loginData.token;
    expect(token).toBeDefined();

    // 3. Request NIFTY 50 data
    let response;
    await test.step('What we use: GET /nifty50/companies with Bearer JWT token.', async () => {
      response = await request.get(`${API_URL}/nifty50/companies`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
    });

    await test.step('What we expected: 200 OK with success=true and exactly 50 companies.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: 50 constituents with company_name, industry, symbol, series, isin.', async () => {
      const json = await response.json();
      expect(json.success).toBe(true);
      expect(json.count).toBe(50);
      expect(json.updated_at).toBeDefined();
      expect(Array.isArray(json.companies)).toBe(true);
      expect(json.companies.length).toBe(50);

      // Verify each company has required fields
      for (const comp of json.companies) {
        expect(comp.company_name).toBeTruthy();
        expect(comp.industry).toBeTruthy();
        expect(comp.symbol).toBeTruthy();
        expect(comp.series).toBeTruthy();
        expect(comp.isin).toBeTruthy();
        expect(comp.isin.length).toBe(12);
      }
    });

    await test.step('Why it got this output: The backend securely loaded official constituent data from data/nifty50.json.', async () => {});
  });

});
