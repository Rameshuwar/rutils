import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API Compound Interest Calculator Persona', () => {

  // ================================================================
  // GROUP 1: HAPPY PATHS вҖ” one test per compounding frequency
  // ================================================================

  test('Persona: Successful Compound Interest (yearly)', async ({ request }) => {
    await test.step('Why we use this test: Verify the default yearly compounding produces the expected total for a 5-year investment.', async () => {});

    let response;

    await test.step('What we use: POST /calculate-compound-interest with P=100000, R=8.5, T=5, yearly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: A = 100000 Г— 1.085^5 вүҲ 150365.57, CI вүҲ 50365.57, n=1.', async () => {
      const json = await response.json();
      expect(json.compoundsPerYear).toBe(1);
      expect(json.totalAmount).toBeCloseTo(150365.57, 0.5);
      expect(json.interest).toBeCloseTo(50365.57, 0.5);
      expect(json.compoundingFrequency).toBe('yearly');
    });

    await test.step('Why it got this output: The backend applied A = P Г— (1 + r/n)^(nГ—t) with n=1 for yearly compounding.', async () => {});
  });

  test('Persona: Successful Compound Interest (half_yearly)', async ({ request }) => {
    await test.step('Why we use this test: Verify half-yearly compounding uses n=2 in the formula.', async () => {});

    let response;

    await test.step('What we use: POST with P=100000, R=8, T=1, half_yearly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 1,
          timeUnit: 'years',
          compoundingFrequency: 'half_yearly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: A = 100000 Г— 1.04^2 = 108160, n=2.', async () => {
      const json = await response.json();
      expect(json.compoundsPerYear).toBe(2);
      expect(json.totalAmount).toBeCloseTo(108160, 0.5);
    });

    await test.step('Why it got this output: The backend routed the request to n=2 for half-yearly compounding.', async () => {});
  });

  test('Persona: Successful Compound Interest (quarterly)', async ({ request }) => {
    await test.step('Why we use this test: Verify quarterly compounding uses n=4 and produces a slightly higher total than yearly.', async () => {});

    let response;

    await test.step('What we use: POST with P=100000, R=8, T=1, quarterly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 1,
          timeUnit: 'years',
          compoundingFrequency: 'quarterly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: A = 100000 Г— (1.02)^4 вүҲ 108243.22, n=4.', async () => {
      const json = await response.json();
      expect(json.compoundsPerYear).toBe(4);
      expect(json.totalAmount).toBeCloseTo(108243.22, 0.5);
    });

    await test.step('Why it got this output: The backend routed the request to n=4 for quarterly compounding.', async () => {});
  });

  test('Persona: Successful Compound Interest (monthly)', async ({ request }) => {
    await test.step('Why we use this test: Verify monthly compounding uses n=12 and matches the standard formula.', async () => {});

    let response;

    await test.step('What we use: POST with P=100000, R=12, T=1, monthly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 12,
          time: 1,
          timeUnit: 'years',
          compoundingFrequency: 'monthly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: A = 100000 Г— (1.01)^12 вүҲ 112682.50, n=12.', async () => {
      const json = await response.json();
      expect(json.compoundsPerYear).toBe(12);
      expect(json.totalAmount).toBeCloseTo(112682.50, 0.5);
    });

    await test.step('Why it got this output: The backend routed the request to n=12 for monthly compounding.', async () => {});
  });

  test('Persona: Successful Compound Interest (daily)', async ({ request }) => {
    await test.step('Why we use this test: Verify the daily frequency uses n=365 вҖ” the densest compounding option offered.', async () => {});

    let response;

    await test.step('What we use: POST with P=100000, R=8, T=1, daily.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 1,
          timeUnit: 'years',
          compoundingFrequency: 'daily'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: A вүҲ 108327.76, n=365.', async () => {
      const json = await response.json();
      expect(json.compoundsPerYear).toBe(365);
      // 100000 * (1 + 0.08/365)^365 вүҲ 108327.757
      expect(json.totalAmount).toBeCloseTo(108327.76, 1);
    });

    await test.step('Why it got this output: The backend routed the request to n=365 for daily compounding.', async () => {});
  });

  test('Persona: Successful Compound Interest (month-based tenure)', async ({ request }) => {
    await test.step('Why we use this test: Verify month-based tenure is normalized to years before applying the formula.', async () => {});

    let response;

    await test.step('What we use: POST with P=50000, R=10, T=24 months, yearly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 50000,
          annualInterestRate: 10,
          time: 24,
          timeUnit: 'months',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: 24 months = 2 years, so A = 50000 Г— 1.1^2 = 60500.', async () => {
      const json = await response.json();
      expect(json.timeYears).toBeCloseTo(2, 6);
      expect(json.totalAmount).toBeCloseTo(60500, 0.5);
    });

    await test.step('Why it got this output: The backend divided months by 12 to get the year tenure before applying the formula.', async () => {});
  });

  test('Persona: Successful Compound Interest (zero rate)', async ({ request }) => {
    await test.step('Why we use this test: A zero-rate compound investment must return the principal unchanged.', async () => {});

    let response;

    await test.step('What we use: POST with annualInterestRate=0.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 0,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: interest must be 0 and totalAmount must equal principal.', async () => {
      const json = await response.json();
      expect(json.interest).toBe(0);
      expect(json.totalAmount).toBe(100000);
      expect(json.effectiveAnnualRate).toBe(0);
    });

    await test.step('Why it got this output: (1 + 0/n)^anything = 1, so A = P, and CI = 0.', async () => {});
  });

  test('Persona: Successful Compound Interest (default frequency = yearly)', async ({ request }) => {
    await test.step('Why we use this test: When compoundingFrequency is omitted, the backend must default to "yearly".', async () => {});

    let response;

    await test.step('What we use: POST without compoundingFrequency.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 1,
          timeUnit: 'years'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: compoundingFrequency = "yearly", n = 1.', async () => {
      const json = await response.json();
      expect(json.compoundingFrequency).toBe('yearly');
      expect(json.compoundsPerYear).toBe(1);
    });

    await test.step('Why it got this output: The backend substitutes "yearly" whenever the field is empty.', async () => {});
  });

  // ================================================================
  // GROUP 2: RESPONSE SHAPE CORRECTNESS
  // ================================================================

  test('Persona: Response shape contains all required fields', async ({ request }) => {
    await test.step('Why we use this test: Confirm the API contract includes every field the frontend depends on.', async () => {});

    let response;

    await test.step('What we use: We send a standard quarterly request.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'quarterly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Every required field is present.', async () => {
      const json = await response.json();
      expect(json).toHaveProperty('principal');
      expect(json).toHaveProperty('interest');
      expect(json).toHaveProperty('totalAmount');
      expect(json).toHaveProperty('annualInterestRate');
      expect(json).toHaveProperty('timeYears');
      expect(json).toHaveProperty('compoundingFrequency');
      expect(json).toHaveProperty('compoundsPerYear');
      expect(json).toHaveProperty('effectiveAnnualRate');
      expect(json).toHaveProperty('breakdown');
      expect(json.breakdown).toHaveProperty('principalPercent');
      expect(json.breakdown).toHaveProperty('interestPercent');
      expect(Array.isArray(json.steps)).toBe(true);
      expect(json.steps.length).toBeGreaterThan(0);
    });

    await test.step('Why it got this output: The handler marshals the full CompoundInterestResponse struct.', async () => {});
  });

  test('Persona: Effective Annual Rate matches (1 + r/n)^n вҲ’ 1', async ({ request }) => {
    await test.step('Why we use this test: The EAR is a derived quantity вҖ” verify it uses the correct formula, not the nominal rate.', async () => {});

    let response;

    await test.step('What we use: POST with R=12, monthly (n=12).', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 12,
          time: 1,
          timeUnit: 'years',
          compoundingFrequency: 'monthly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: EAR = (1 + 0.12/12)^12 вҲ’ 1 = 12.6825%.', async () => {
      const json = await response.json();
      expect(json.effectiveAnnualRate).toBeCloseTo(12.6825, 0.01);
    });

    await test.step('Why it got this output: The backend computed EAR = ((1 + r/n)^n вҲ’ 1) Г— 100.', async () => {});
  });

  test('Persona: Breakdown percentages sum to exactly 100', async ({ request }) => {
    await test.step('Why we use this test: Internal-consistency check вҖ” the principal/interest split must always sum to 100.', async () => {});

    let response;

    await test.step('What we use: We send a standard yearly request.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The two percentages sum to 100 (within rounding).', async () => {
      const json = await response.json();
      const sum = json.breakdown.principalPercent + json.breakdown.interestPercent;
      expect(sum).toBeCloseTo(100, 1);
    });

    await test.step('Why it got this output: The backend computes the interest percent as 100 вҲ’ principal percent to absorb rounding drift.', async () => {});
  });

  test('Persona: Total amount reconciles with principal + interest', async ({ request }) => {
    await test.step('Why we use this test: Total must always equal principal + interest вҖ” the core arithmetic invariant.', async () => {});

    let response;

    await test.step('What we use: We send a quarterly request with a fractional rate.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 250000,
          annualInterestRate: 7.25,
          time: 3,
          timeUnit: 'years',
          compoundingFrequency: 'quarterly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: totalAmount === principal + interest.', async () => {
      const json = await response.json();
      expect(json.totalAmount).toBeCloseTo(json.principal + json.interest, 2);
    });

    await test.step('Why it got this output: The backend computes total via the compound formula, then derives interest as total вҲ’ principal.', async () => {});
  });

  // ================================================================
  // GROUP 3: YEARLY BREAKDOWN SCHEDULE
  // ================================================================

  test('Persona: Yearly breakdown has one row per full year (tenure вүҘ 1)', async ({ request }) => {
    await test.step('Why we use this test: The yearly schedule must have exactly floor(timeYears) rows for tenures of 1+ years.', async () => {});

    let response;

    await test.step('What we use: POST with T=5 years, yearly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: yearlyBreakdown has 5 rows.', async () => {
      const json = await response.json();
      expect(Array.isArray(json.yearlyBreakdown)).toBe(true);
      expect(json.yearlyBreakdown).toHaveLength(5);
    });

    await test.step('Why it got this output: The backend walks the growth one year at a time up to floor(timeYears).', async () => {});
  });

  test('Persona: Yearly breakdown absent for tenure < 1 year', async ({ request }) => {
    await test.step('Why we use this test: A sub-year tenure has no full-year rows, so the schedule must be omitted.', async () => {});

    let response;

    await test.step('What we use: POST with T=6 months.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 6,
          timeUnit: 'months',
          compoundingFrequency: 'monthly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: yearlyBreakdown is absent or empty.', async () => {
      const json = await response.json();
      const rows = json.yearlyBreakdown;
      expect(rows === undefined || (Array.isArray(rows) && rows.length === 0)).toBe(true);
    });

    await test.step('Why it got this output: The builder returns nil when timeYears < 1.', async () => {});
  });

  test('Persona: Yearly breakdown rows chain opening вҶ’ closing balances', async ({ request }) => {
    await test.step('Why we use this test: Every row\'s opening balance must equal the previous row\'s closing balance вҖ” a strictly chained series.', async () => {});

    let response;

    await test.step('What we use: POST with T=4 years, quarterly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 4,
          timeUnit: 'years',
          compoundingFrequency: 'quarterly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Row N+1 opening == Row N closing for all N, and row 1 opening == principal.', async () => {
      const json = await response.json();
      const rows = json.yearlyBreakdown;
      expect(Array.isArray(rows)).toBe(true);
      expect(rows.length).toBeGreaterThan(1);

      // First row opens at the principal.
      expect(rows[0].openingBalance).toBeCloseTo(json.principal, 2);

      // Every subsequent row opens where the previous row closed.
      for (let i = 1; i < rows.length; i++) {
        expect(rows[i].openingBalance).toBeCloseTo(rows[i - 1].closingBalance, 2);
      }
    });

    await test.step('Why it got this output: The schedule builder carries the running balance forward exactly between rows.', async () => {});
  });

  test('Persona: Final yearly row closing balance reconciles with totalAmount', async ({ request }) => {
    await test.step('Why we use this test: The last row\'s closing balance must equal the top-level totalAmount вҖ” otherwise the schedule is inconsistent with the summary.', async () => {});

    let response;

    await test.step('What we use: POST with T=5 years, quarterly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 500000,
          annualInterestRate: 7.5,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'quarterly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: last.closingBalance вүҲ totalAmount.', async () => {
      const json = await response.json();
      const rows = json.yearlyBreakdown;
      const last = rows[rows.length - 1];
      expect(last.closingBalance).toBeCloseTo(json.totalAmount, 0.5);
    });

    await test.step('Why it got this output: The row builder evaluates the exact compound formula for each year boundary, so the final row\'s closing matches the top-level total.', async () => {});
  });

  test('Persona: Each yearly row\'s interest equals closing вҲ’ opening', async ({ request }) => {
    await test.step('Why we use this test: Internal per-row arithmetic invariant вҖ” interestEarned must always equal closingBalance вҲ’ openingBalance.', async () => {});

    let response;

    await test.step('What we use: POST with T=3 years, monthly.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 200000,
          annualInterestRate: 9,
          time: 3,
          timeUnit: 'years',
          compoundingFrequency: 'monthly'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: For every row, interestEarned == closingBalance вҲ’ openingBalance.', async () => {
      const json = await response.json();
      for (const row of json.yearlyBreakdown) {
        const diff = row.closingBalance - row.openingBalance;
        expect(row.interestEarned).toBeCloseTo(diff, 2);
      }
    });

    await test.step('Why it got this output: The builder computes interest as closing вҲ’ opening for each row.', async () => {});
  });

  // ================================================================
  // GROUP 4: VALIDATION ERRORS
  // ================================================================

  test('Persona: Error - invalid compoundingFrequency (hourly)', async ({ request }) => {
    await test.step('Why we use this test: Only the five documented frequencies are supported.', async () => {});

    let response;

    await test.step('What we use: POST with compoundingFrequency="hourly".', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'hourly'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error lists the allowed frequencies.', async () => {
      const txt = await response.text();
      expect(txt).toContain('compoundingFrequency must be one of');
    });

    await test.step('Why it got this output: The validator checks the frequency against a whitelist.', async () => {});
  });

  test('Persona: Error - day-based tenure rejected', async ({ request }) => {
    await test.step('Why we use this test: Day-based tenure is not supported for compound interest (years or months only).', async () => {});

    let response;

    await test.step('What we use: POST with timeUnit="days".', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 365,
          timeUnit: 'days',
          compoundingFrequency: 'daily'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions "years" or "months".', async () => {
      const txt = await response.text();
      expect(txt).toContain('years');
      expect(txt).toContain('months');
    });

    await test.step('Why it got this output: The handler explicitly rejects day-based tenure before invoking the compound formula.', async () => {});
  });

  test('Persona: Error - negative principal', async ({ request }) => {
    await test.step('Why we use this test: A negative principal is nonsensical.', async () => {});

    let response;

    await test.step('What we use: POST with principal=-1.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: -1,
          annualInterestRate: 8,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions principal must be greater than zero.', async () => {
      const txt = await response.text();
      expect(txt).toContain('principal must be greater than zero');
    });

    await test.step('Why it got this output: The shared validator rejects non-positive principal.', async () => {});
  });

  test('Persona: Error - rate above 100%', async ({ request }) => {
    await test.step('Why we use this test: A rate above 100% is outside the supported range.', async () => {});

    let response;

    await test.step('What we use: POST with annualInterestRate=150.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 150,
          time: 5,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions annualInterestRate cannot exceed 100%.', async () => {
      const txt = await response.text();
      expect(txt).toContain('annualInterestRate cannot exceed 100');
    });

    await test.step('Why it got this output: The shared validator enforces the 0вҖ“100 range.', async () => {});
  });

  test('Persona: Error - zero time', async ({ request }) => {
    await test.step('Why we use this test: A zero tenure makes no sense for compound interest.', async () => {});

    let response;

    await test.step('What we use: POST with time=0.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 0,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions time must be greater than zero.', async () => {
      const txt = await response.text();
      expect(txt).toContain('time must be greater than zero');
    });

    await test.step('Why it got this output: The time normalizer rejects non-positive time.', async () => {});
  });

  test('Persona: Error - invalid timeUnit (weeks)', async ({ request }) => {
    await test.step('Why we use this test: Only "years" and "months" are valid for compound interest.', async () => {});

    let response;

    await test.step('What we use: POST with timeUnit="weeks".', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 5,
          timeUnit: 'weeks',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the allowed time units.', async () => {
      const txt = await response.text();
      expect(txt).toContain('timeUnit must be');
    });

    await test.step('Why it got this output: The validator rejects unknown time units.', async () => {});
  });

  test('Persona: Error - tenure exceeds maximum (200 years)', async ({ request }) => {
    await test.step('Why we use this test: Extremely long tenures must be rejected to prevent overflow.', async () => {});

    let response;

    await test.step('What we use: POST with time=200 years.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8,
          time: 200,
          timeUnit: 'years',
          compoundingFrequency: 'yearly'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the maximum tenure.', async () => {
      const txt = await response.text();
      expect(txt).toContain('time exceeds the maximum');
    });

    await test.step('Why it got this output: The normalizer caps tenure at 100 years.', async () => {});
  });

  test('Persona: Error - malformed JSON body', async ({ request }) => {
    await test.step('Why we use this test: A malformed JSON payload must be rejected before any processing.', async () => {});

    let response;

    await test.step('What we use: POST with a raw non-JSON string.', async () => {
      response = await request.post(`${API_URL}/calculate-compound-interest`, {
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

  test('Persona: Error - wrong HTTP method (GET)', async ({ request }) => {
    await test.step('Why we use this test: The endpoint only accepts POST вҖ” GET must be rejected.', async () => {});

    let response;

    await test.step('What we use: GET /calculate-compound-interest.', async () => {
      response = await request.get(`${API_URL}/calculate-compound-interest`);
    });

    await test.step('What we expected: 405 Method Not Allowed.', async () => {
      expect(response.status()).toBe(405);
    });

    await test.step('What we get: The error mentions only POST is allowed.', async () => {
      const txt = await response.text();
      expect(txt).toContain('Only POST method is allowed');
    });

    await test.step('Why it got this output: The handler guards the method at the top.', async () => {});
  });

});