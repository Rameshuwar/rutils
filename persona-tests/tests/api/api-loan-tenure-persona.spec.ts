import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API Loan Tenure Calculator Persona (Borrower-Centric)', () => {

  // ================================================================
  // GROUP 1: HAPPY PATHS
  // ================================================================

  test('Persona: Successful Loan Tenure (standard car loan)', async ({ request }) => {
    await test.step('Why we use this test: A common person takes a ₹5,00,000 car loan at 10.5% and can only afford ₹15,000/month. Verify the backend tells them how long it will take and what it will cost.', async () => {});

    let response;

    await test.step('What we use: POST /calculate-loan-tenure with principal=500000, annualInterestRate=10.5, monthlyPayment=15000.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: {
          principal: 500000,
          annualInterestRate: 10.5,
          monthlyPayment: 15000,
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: tenure ≈ 41 months, totalPayment > principal, amortization array populated.', async () => {
      const json = await response.json();

      expect(json).toHaveProperty('tenureMonths');
      expect(json).toHaveProperty('tenureYears');
      expect(json).toHaveProperty('emi', 15000);
      expect(json).toHaveProperty('principal', 500000);
      expect(json).toHaveProperty('totalInterest');
      expect(json).toHaveProperty('totalPayment');
      expect(json).toHaveProperty('breakdown');
      expect(Array.isArray(json.amortization)).toBe(true);

      expect(json.tenureMonths).toBeGreaterThanOrEqual(40);
      expect(json.tenureMonths).toBeLessThanOrEqual(42);
      expect(json.totalPayment).toBeGreaterThan(json.principal);
      expect(json.totalInterest).toBeGreaterThan(0);

      // Every amortization row must be present
      expect(json.amortization).toHaveLength(json.tenureMonths);

      // Last row must close at exactly zero.
      const last = json.amortization[json.amortization.length - 1];
      expect(last.closingBalance).toBe(0);
    });

    await test.step('Why it got this output: The backend solved N = -log(1 - (P × r)/EMI) / log(1 + r) and walked the schedule month by month.', async () => {});
  });

  test('Persona: Zero-Interest Loan (equally split)', async ({ request }) => {
    await test.step('Why we use this test: A 0% interest loan is a valid edge case — tenure must equal principal / payment exactly.', async () => {});

    let response;

    await test.step('What we use: POST with principal=120000, rate=0, payment=10000.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: {
          principal: 120000,
          annualInterestRate: 0,
          monthlyPayment: 10000,
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: exactly 12 months, zero interest, principal only.', async () => {
      const json = await response.json();
      expect(json.tenureMonths).toBe(12);
      expect(json.totalInterest).toBe(0);
      expect(json.totalPayment).toBe(120000);
      expect(json.breakdown.principalPercent).toBe(100);
      expect(json.breakdown.interestPercent).toBe(0);
    });

    await test.step('Why it got this output: The backend took the zero-rate shortcut N = ceil(P / EMI).', async () => {});
  });

  test('Persona: Long-term home loan', async ({ request }) => {
    await test.step('Why we use this test: Verify the calculator handles a large multi-decade home loan scenario.', async () => {});

    let response;

    await test.step('What we use: POST with principal=5000000, rate=8.5, payment=40000.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: {
          principal: 5000000,
          annualInterestRate: 8.5,
          monthlyPayment: 40000,
        },
      });
    });

    await test.step('What we expected: 200 OK with tenure in the 305-309 month range (matches independent loan calculators).', async () => {
      expect(response.status()).toBe(200);
      const json = await response.json();
      expect(json.tenureMonths).toBeGreaterThanOrEqual(305);
      expect(json.tenureMonths).toBeLessThanOrEqual(309);
      expect(json.tenureYears).toBeGreaterThan(25);
      expect(json.tenureYears).toBeLessThan(26);
    });

    await test.step('Why it got this output: The N-solve returned a ~25.6-year tenure (307 months), correctly reported in both months and years.', async () => {});
  });

  test('Persona: Single-month payoff', async ({ request }) => {
    await test.step('Why we use this test: Verify the edge case where the borrower clears the loan in one payment.', async () => {});

    let response;

    await test.step('What we use: POST with principal=10000, rate=12, payment=10100 (covers interest + principal in one shot).', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: {
          principal: 10000,
          annualInterestRate: 12,
          monthlyPayment: 10100,
        },
      });
    });

    await test.step('What we expected: 200 OK, tenureMonths = 1.', async () => {
      expect(response.status()).toBe(200);
      const json = await response.json();
      expect(json.tenureMonths).toBe(1);
      expect(json.amortization).toHaveLength(1);
    });

    await test.step('Why it got this output: N rounded up to 1 because the payment covered everything in the first month.', async () => {});
  });

  // ================================================================
  // GROUP 2: RESPONSE SHAPE AND INVARIANTS
  // ================================================================

  test('Persona: Response shape matches the existing EMI response', async ({ request }) => {
    await test.step('Why we use this test: Confirm the new endpoint returns a superset of the EMI response, so the frontend can render both with the same components.', async () => {});

    let response;

    await test.step('What we use: POST a standard request.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 300000, annualInterestRate: 11, monthlyPayment: 12000 },
      });
    });

    await test.step('What we expected: 200 OK and all the keys the EMI endpoint emits, plus tenureYears.', async () => {
      expect(response.status()).toBe(200);
      const json = await response.json();

      const required = [
        'emi', 'principal', 'totalInterest', 'totalPayment',
        'tenureMonths', 'monthlyRatePercent', 'breakdown', 'amortization',
      ];
      for (const key of required) {
        expect(json, `missing key: ${key}`).toHaveProperty(key);
      }
      expect(json).toHaveProperty('tenureYears');
      expect(json.breakdown).toHaveProperty('principalPercent');
      expect(json.breakdown).toHaveProperty('interestPercent');
    });

    await test.step('Why it got this output: The response struct reuses AmortizationRow and EMIBreakdown from the existing EMI types.', async () => {});
  });

  test('Persona: Totals reconcile with the amortization schedule', async ({ request }) => {
    await test.step('Why we use this test: The sum of every row must equal the summary totals — the core internal consistency check.', async () => {});

    let response;

    await test.step('What we use: POST a standard request.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 750000, annualInterestRate: 9.25, monthlyPayment: 20000 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: sum(principalPaid) = principal, sum(interestPaid) = totalInterest, sum(totalPaid) = totalPayment.', async () => {
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

    await test.step('Why it got this output: The builder accumulates totals while producing rows, and the final row sweeps any rounding drift.', async () => {});
  });

  test('Persona: Opening balances chain correctly', async ({ request }) => {
    await test.step('Why we use this test: Every row\'s opening balance must equal the previous row\'s closing balance.', async () => {});

    let response;

    await test.step('What we use: POST a standard request.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 400000, annualInterestRate: 11, monthlyPayment: 18000 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Strictly chained balances.', async () => {
      const json = await response.json();
      const rows = json.amortization;

      expect(rows[0].openingBalance).toBeCloseTo(json.principal, 1);
      for (let i = 1; i < rows.length; i++) {
        expect(rows[i].openingBalance).toBeCloseTo(rows[i - 1].closingBalance, 1);
      }
    });

    await test.step('Why it got this output: Each row opens with the previous row\'s closing balance.', async () => {});
  });

  test('Persona: Breakdown percentages sum to 100', async ({ request }) => {
    await test.step('Why we use this test: Internal consistency — the principal/interest split must always sum to 100.', async () => {});

    let response;

    await test.step('What we use: POST a standard request.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 250000, annualInterestRate: 12, monthlyPayment: 15000 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: sum ≈ 100.', async () => {
      const json = await response.json();
      const sum = json.breakdown.principalPercent + json.breakdown.interestPercent;
      expect(sum).toBeCloseTo(100, 1);
    });

    await test.step('Why it got this output: The interest percent is derived as 100 − principal percent to absorb rounding.', async () => {});
  });

  // ================================================================
  // GROUP 3: VALIDATION ERRORS
  // ================================================================

  test('Persona: Error - monthly payment too low to cover interest', async ({ request }) => {
    await test.step('Why we use this test: If the payment cannot even cover the monthly interest, the loan is mathematically unrepayable. The backend must reject it with a helpful message.', async () => {});

    let response;

    await test.step('What we use: POST principal=100000, rate=12, payment=500 (monthly interest is 1000).', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: {
          principal: 100000,
          annualInterestRate: 12,
          monthlyPayment: 500,
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: A message explaining the payment is too low and suggesting a minimum.', async () => {
      const txt = await response.text();
      expect(txt).toContain('too low');
      expect(txt.toLowerCase()).toContain('interest');
    });

    await test.step('Why it got this output: The backend detected EMI ≤ P × r and refused to compute an infinite tenure.', async () => {});
  });

  test('Persona: Error - payment exactly equals monthly interest', async ({ request }) => {
    await test.step('Why we use this test: Boundary case — the loan never amortizes if EMI = P × r. Must be rejected.', async () => {});

    let response;

    await test.step('What we use: POST principal=100000, rate=12, payment=1000.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: {
          principal: 100000,
          annualInterestRate: 12,
          monthlyPayment: 1000,
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: A descriptive error.', async () => {
      const txt = await response.text();
      expect(txt).toContain('too low');
    });

    await test.step('Why it got this output: The guard condition uses strict "≤" so this boundary is caught.', async () => {});
  });

  test('Persona: Error - negative principal', async ({ request }) => {
    await test.step('Why we use this test: A negative principal is nonsensical.', async () => {});

    let response;

    await test.step('What we use: POST principal=-50000.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: -50000, annualInterestRate: 10, monthlyPayment: 5000 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions principal.', async () => {
      const txt = await response.text();
      expect(txt).toContain('principal must be greater than zero');
    });

    await test.step('Why it got this output: The validator rejects non-positive principal.', async () => {});
  });

  test('Persona: Error - zero principal', async ({ request }) => {
    await test.step('Why we use this test: A zero-loan is meaningless.', async () => {});

    let response;

    await test.step('What we use: POST principal=0.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 0, annualInterestRate: 10, monthlyPayment: 5000 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: Error mentions principal.', async () => {
      const txt = await response.text();
      expect(txt).toContain('principal must be greater than zero');
    });
  });

  test('Persona: Error - negative interest rate', async ({ request }) => {
    await test.step('Why we use this test: Negative rates are not valid loan rates.', async () => {});

    let response;

    await test.step('What we use: POST rate=-1.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 100000, annualInterestRate: -1, monthlyPayment: 10000 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: Error mentions annualInterestRate.', async () => {
      const txt = await response.text();
      expect(txt).toContain('annualInterestRate cannot be negative');
    });
  });

  test('Persona: Error - rate above 100%', async ({ request }) => {
    await test.step('Why we use this test: Rates above 100% are rejected by the shared validator.', async () => {});

    let response;

    await test.step('What we use: POST rate=150.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 100000, annualInterestRate: 150, monthlyPayment: 50000 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: Error mentions 100%.', async () => {
      const txt = await response.text();
      expect(txt).toContain('annualInterestRate cannot exceed 100');
    });
  });

  test('Persona: Error - zero monthly payment', async ({ request }) => {
    await test.step('Why we use this test: A zero payment can never repay a loan.', async () => {});

    let response;

    await test.step('What we use: POST monthlyPayment=0.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 100000, annualInterestRate: 10, monthlyPayment: 0 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: Error mentions monthlyPayment.', async () => {
      const txt = await response.text();
      expect(txt).toContain('monthlyPayment must be greater than zero');
    });
  });

  test('Persona: Error - negative monthly payment', async ({ request }) => {
    await test.step('Why we use this test: A negative payment is nonsensical.', async () => {});

    let response;

    await test.step('What we use: POST monthlyPayment=-5000.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: { principal: 100000, annualInterestRate: 10, monthlyPayment: -5000 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: Error mentions monthlyPayment.', async () => {
      const txt = await response.text();
      expect(txt).toContain('monthlyPayment must be greater than zero');
    });
  });

  test('Persona: Error - tenure exceeds maximum (600 months)', async ({ request }) => {
    await test.step('Why we use this test: An extremely low payment that just barely covers interest pushes the payoff beyond 600 months (50 years). The backend must cap this and reject.', async () => {});

    let response;

    await test.step('What we use: POST principal=10000000 (1 crore), rate=15, monthlyPayment=125050 (just ₹50 above the ₹1,25,000 monthly interest).', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        data: {
          principal: 10000000,
          annualInterestRate: 15,
          monthlyPayment: 125050,
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: An error mentioning the loan cannot be repaid within the maximum tenure.', async () => {
      const txt = await response.text();
      expect(txt).toContain('cannot be repaid within');
    });

    await test.step('Why it got this output: The solver computed N ≈ 630 months, which exceeds the 600-month cap. The backend correctly rejected rather than returning an unreasonably long tenure.', async () => {});
  });

  test('Persona: Error - malformed JSON body', async ({ request }) => {
    await test.step('Why we use this test: An invalid body must be rejected before processing.', async () => {});

    let response;

    await test.step('What we use: POST a raw non-JSON string.', async () => {
      response = await request.post(`${API_URL}/calculate-loan-tenure`, {
        headers: { 'Content-Type': 'application/json' },
        data: 'this is not valid json',
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions Invalid JSON.', async () => {
      const txt = await response.text();
      expect(txt).toContain('Invalid JSON');
    });
  });

  test('Persona: Error - wrong HTTP method (GET)', async ({ request }) => {
    await test.step('Why we use this test: The endpoint only accepts POST.', async () => {});

    let response;

    await test.step('What we use: GET /calculate-loan-tenure.', async () => {
      response = await request.get(`${API_URL}/calculate-loan-tenure`);
    });

    await test.step('What we expected: 405 Method Not Allowed.', async () => {
      expect(response.status()).toBe(405);
    });

    await test.step('What we get: The error mentions Only POST is allowed.', async () => {
      const txt = await response.text();
      expect(txt).toContain('Only POST method is allowed');
    });
  });
});