import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API Loan EMI Calculator Persona', () => {

  // ================================================================
  // GROUP 1: HAPPY PATH
  // ================================================================

  test('Persona: Successful EMI Calculation (Standard Loan)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify the backend correctly computes the EMI for a typical car/personal loan with a 10.5% annual rate over 5 years.', async () => {});

    let response;

    await test.step('What we use: We send a POST to /calculate-emi with principal 500000, annualInterestRate 10.5, tenure 60 months.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: {
          principal: 500000,
          annualInterestRate: 10.5,
          tenure: 60,
          tenureUnit: 'months'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the JSON and confirm the EMI, totals, breakdown, and amortization schedule are all present and internally consistent.', async () => {
      const json = await response.json();

      expect(json).toHaveProperty('emi');
      expect(json).toHaveProperty('principal', 500000);
      expect(json).toHaveProperty('totalInterest');
      expect(json).toHaveProperty('totalPayment');
      expect(json).toHaveProperty('tenureMonths', 60);
      expect(json).toHaveProperty('monthlyRatePercent');
      expect(json).toHaveProperty('breakdown');
      expect(Array.isArray(json.amortization)).toBe(true);
      expect(json.amortization).toHaveLength(60);

      // EMI ≈ 10747.28 (standard value from online calculators)
      expect(json.emi).toBeGreaterThan(10700);
      expect(json.emi).toBeLessThan(10800);

      // Breakdown percentages must sum to 100
      const sum = json.breakdown.principalPercent + json.breakdown.interestPercent;
      expect(sum).toBeCloseTo(100, 1);

      // Last row closes at zero
      const last = json.amortization[json.amortization.length - 1];
      expect(last.closingBalance).toBe(0);
    });

    await test.step('Why it got this output: The backend applied the EMI formula EMI = [P × r × (1+r)^N] / [(1+r)^N − 1] with r = 10.5/12/100 and N = 60.', async () => {});
  });

  test('Persona: Successful EMI with Year-Based Tenure', async ({ request }) => {
    await test.step('Why we use this test: We want to verify that specifying tenure in "years" produces the same result as the equivalent months-based calculation.', async () => {});

    let monthsResponse: any;
    let yearsResponse: any;

    await test.step('What we use: We call /calculate-emi twice — once with 60 months, once with 5 years.', async () => {
      monthsResponse = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 500000, annualInterestRate: 10.5, tenure: 60, tenureUnit: 'months' }
      });
      yearsResponse = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 500000, annualInterestRate: 10.5, tenure: 5, tenureUnit: 'years' }
      });
    });

    await test.step('What we expected: Both responses should be 200 OK and produce identical EMI values.', async () => {
      expect(monthsResponse.status()).toBe(200);
      expect(yearsResponse.status()).toBe(200);

      const monthsJson = await monthsResponse.json();
      const yearsJson = await yearsResponse.json();

      expect(yearsJson.emi).toBe(monthsJson.emi);
      expect(yearsJson.tenureMonths).toBe(60);
    });

    await test.step('Why it got this output: The backend normalizes "years" to months by multiplying by 12 before applying the formula.', async () => {});
  });

  test('Persona: Successful Zero-Interest Loan', async ({ request }) => {
    await test.step('Why we use this test: An interest-free loan is a valid edge case where the standard EMI formula would divide by zero — the backend must special-case it.', async () => {});

    let response;

    await test.step('What we use: We send a request with annualInterestRate = 0 and a 12-month tenure on a ₹1,20,000 loan.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 120000, annualInterestRate: 0, tenure: 12, tenureUnit: 'months' }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: EMI must equal principal / months (10000), totalInterest must be 0, and totalPayment must equal principal.', async () => {
      const json = await response.json();
      expect(json.emi).toBe(10000);
      expect(json.totalInterest).toBe(0);
      expect(json.totalPayment).toBe(120000);
      expect(json.breakdown.principalPercent).toBe(100);
      expect(json.breakdown.interestPercent).toBe(0);
    });

    await test.step('Why it got this output: The backend detects monthlyRate == 0 and returns principal / months directly, avoiding division by zero.', async () => {});
  });

  test('Persona: Successful Long-Term Home Loan', async ({ request }) => {
    await test.step('Why we use this test: We want to stress-test the amortization schedule builder with a 30-year (360-month) loan — the most common home-loan scenario.', async () => {});

    let response;

    await test.step('What we use: We send a request for ₹50,00,000 @ 8.5% over 30 years.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 5000000, annualInterestRate: 8.5, tenure: 30, tenureUnit: 'years' }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: tenureMonths = 360, amortization array has 360 rows, final row closes at 0.', async () => {
      const json = await response.json();
      expect(json.tenureMonths).toBe(360);
      expect(json.amortization).toHaveLength(360);
      const last = json.amortization[359];
      expect(last.closingBalance).toBe(0);
    });

    await test.step('Why it got this output: The backend iterates one row per month and reconciles the final row so the closing balance is exactly zero.', async () => {});
  });

  // ================================================================
  // GROUP 2: AMORTIZATION SCHEDULE CORRECTNESS
  // ================================================================

  test('Persona: Amortization Totals Reconcile with Summary', async ({ request }) => {
    await test.step('Why we use this test: The sum of every row in the amortization schedule must exactly equal the summary totals — a critical internal-consistency check.', async () => {});

    let response;

    await test.step('What we use: We request EMI for ₹7,50,000 @ 9.25% for 84 months.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 750000, annualInterestRate: 9.25, tenure: 84, tenureUnit: 'months' }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Sum of principalPaid == principal, sum of interestPaid == totalInterest, sum of totalPaid == totalPayment.', async () => {
      const json = await response.json();

      let sumPrincipal = 0, sumInterest = 0, sumPaid = 0;
      for (const row of json.amortization) {
        sumPrincipal += row.principalPaid;
        sumInterest += row.interestPaid;
        sumPaid += row.totalPaid;
      }

      expect(sumPrincipal).toBeCloseTo(json.principal, 1);
      expect(sumInterest).toBeCloseTo(json.totalInterest, 1);
      expect(sumPaid).toBeCloseTo(json.totalPayment, 1);
    });

    await test.step('Why it got this output: The backend accumulates totals while building the schedule, and the final row is adjusted to absorb rounding drift.', async () => {});
  });

  test('Persona: Amortization Opening Balance Chains Correctly', async ({ request }) => {
    await test.step('Why we use this test: Every row\'s opening balance must equal the previous row\'s closing balance — a strictly chained series.', async () => {});

    let response;

    await test.step('What we use: We request EMI for ₹3,00,000 @ 12% over 24 months.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 300000, annualInterestRate: 12, tenure: 24, tenureUnit: 'months' }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Row N+1 opening == Row N closing, for all N.', async () => {
      const json = await response.json();
      for (let i = 1; i < json.amortization.length; i++) {
        expect(json.amortization[i].openingBalance)
          .toBeCloseTo(json.amortization[i - 1].closingBalance, 2);
      }
    });

    await test.step('Why it got this output: The schedule builder carries the running balance forward exactly, applying rounding at each step.', async () => {});
  });

  // ================================================================
  // GROUP 3: VALIDATION ERRORS
  // ================================================================

  test('Persona: Error - Missing Required Fields', async ({ request }) => {
    await test.step('Why we use this test: An empty request body should be rejected with a clear 400 error.', async () => {});

    let response;

    await test.step('What we use: We POST an empty JSON body.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, { data: {} });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions principal.', async () => {
      const txt = await response.text();
      expect(txt).toContain('principal');
    });

    await test.step('Why it got this output: The validator runs before any math and rejects the zero principal.', async () => {});
  });

  test('Persona: Error - Negative Principal', async ({ request }) => {
    await test.step('Why we use this test: A negative loan amount is nonsensical and must be rejected.', async () => {});

    let response;

    await test.step('What we use: We send principal = -50000.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: -50000, annualInterestRate: 10, tenure: 12, tenureUnit: 'months' }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions principal.', async () => {
      const txt = await response.text();
      expect(txt).toContain('principal must be greater than zero');
    });

    await test.step('Why it got this output: The validator explicitly rejects non-positive principal.', async () => {});
  });

  test('Persona: Error - Negative Interest Rate', async ({ request }) => {
    await test.step('Why we use this test: A negative annual rate is not a valid loan; it must be rejected.', async () => {});

    let response;

    await test.step('What we use: We send annualInterestRate = -1.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 100000, annualInterestRate: -1, tenure: 12, tenureUnit: 'months' }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions annualInterestRate.', async () => {
      const txt = await response.text();
      expect(txt).toContain('annualInterestRate cannot be negative');
    });

    await test.step('Why it got this output: The validator explicitly rejects negative rates.', async () => {});
  });

  test('Persona: Error - Zero Tenure', async ({ request }) => {
    await test.step('Why we use this test: A zero-month loan makes no sense and must be rejected.', async () => {});

    let response;

    await test.step('What we use: We send tenure = 0.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 100000, annualInterestRate: 10, tenure: 0, tenureUnit: 'months' }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions tenure.', async () => {
      const txt = await response.text();
      expect(txt).toContain('tenure must be greater than zero');
    });

    await test.step('Why it got this output: The validator explicitly rejects non-positive tenure.', async () => {});
  });

  test('Persona: Error - Invalid Tenure Unit', async ({ request }) => {
    await test.step('Why we use this test: Only "months" and "years" are valid tenure units — anything else must be rejected.', async () => {});

    let response;

    await test.step('What we use: We send tenureUnit = "weeks".', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 100000, annualInterestRate: 10, tenure: 12, tenureUnit: 'weeks' }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions tenureUnit.', async () => {
      const txt = await response.text();
      expect(txt).toContain("tenureUnit must be 'months' or 'years'");
    });

    await test.step('Why it got this output: The validator checks the unit against a whitelist.', async () => {});
  });

  test('Persona: Error - Tenure Exceeds Maximum', async ({ request }) => {
    await test.step('Why we use this test: Extremely long tenures must be rejected to prevent overflow and absurd amortization schedules.', async () => {});

    let response;

    await test.step('What we use: We send tenure = 60 years (720 months > 600 cap).', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        data: { principal: 100000, annualInterestRate: 10, tenure: 60, tenureUnit: 'years' }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the maximum tenure.', async () => {
      const txt = await response.text();
      expect(txt).toContain('tenure exceeds the maximum');
    });

    await test.step('Why it got this output: The normalizer caps tenure at MaxTenureMonths (600).', async () => {});
  });

  test('Persona: Error - Malformed JSON Body', async ({ request }) => {
    await test.step('Why we use this test: A malformed JSON payload must be rejected before any processing happens.', async () => {});

    let response;

    await test.step('What we use: We send a raw non-JSON string as the body.', async () => {
      response = await request.post(`${API_URL}/calculate-emi`, {
        headers: { 'Content-Type': 'application/json' },
        data: 'this is not valid json'
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions Invalid JSON.', async () => {
      const txt = await response.text();
      expect(txt).toContain('Invalid JSON');
    });

    await test.step('Why it got this output: The json.Decoder fails and the handler returns a 400 immediately.', async () => {});
  });

});