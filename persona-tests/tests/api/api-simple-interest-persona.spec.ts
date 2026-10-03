import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API Simple Interest Calculator Persona', () => {

  // ================================================================
  // GROUP 1: HAPPY PATHS
  // ================================================================

  test('Persona: Successful Simple Interest (standard 5-year loan)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify the backend correctly computes simple interest for a typical 5-year loan at 8.5% on вӮ№1,00,000.', async () => {});

    let response;

    await test.step('What we use: We send a POST to /calculate-simple-interest with principal=100000, annualInterestRate=8.5, time=5, timeUnit="years".', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the JSON and confirm SI = (100000 Г— 8.5 Г— 5) / 100 = 42500 and Total = 142500.', async () => {
      const json = await response.json();

      expect(json).toHaveProperty('principal', 100000);
      expect(json).toHaveProperty('interest');
      expect(json).toHaveProperty('totalAmount');
      expect(json).toHaveProperty('annualInterestRate', 8.5);
      expect(json).toHaveProperty('timeYears');
      expect(json).toHaveProperty('breakdown');
      expect(json).toHaveProperty('steps');

      expect(json.interest).toBeCloseTo(42500, 2);
      expect(json.totalAmount).toBeCloseTo(142500, 2);
      expect(json.timeYears).toBeCloseTo(5, 6);
    });

    await test.step('Why it got this output: The backend applied SI = (P X R X T) / 100 with T in years and added the interest to the principal.', async () => {});
  });

  test('Persona: Successful Simple Interest (month-based tenure)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify the backend normalizes months to years before applying the SI formula.', async () => {});

    let response;

    await test.step('What we use: We send principal=12000, annualInterestRate=10, time=6, timeUnit="months".', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 12000,
          annualInterestRate: 10,
          time: 6,
          timeUnit: 'months'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: 6 months = 0.5 years, so SI = (12000 Г— 10 Г— 0.5) / 100 = 600.', async () => {
      const json = await response.json();
      expect(json.timeYears).toBeCloseTo(0.5, 6);
      expect(json.interest).toBeCloseTo(600, 2);
      expect(json.totalAmount).toBeCloseTo(12600, 2);
    });

    await test.step('Why it got this output: The backend divided months by 12 to get years before computing the simple interest.', async () => {});
  });

  test('Persona: Successful Simple Interest (day-based tenure)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify the backend converts days to years using a 365-day year.', async () => {});

    let response;

    await test.step('What we use: We send principal=100000, annualInterestRate=12, time=365, timeUnit="days".', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 12,
          time: 365,
          timeUnit: 'days'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: 365 days = 1 year, so SI = (100000 Г— 12 Г— 1) / 100 = 12000.', async () => {
      const json = await response.json();
      expect(json.timeYears).toBeCloseTo(1, 6);
      expect(json.interest).toBeCloseTo(12000, 2);
      expect(json.totalAmount).toBeCloseTo(112000, 2);
    });

    await test.step('Why it got this output: The backend divided days by 365 to get the year fraction before applying the SI formula.', async () => {});
  });

  test('Persona: Successful Simple Interest (zero interest rate)', async ({ request }) => {
    await test.step('Why we use this test: An interest-free loan is a valid edge case вҖ” interest must be exactly 0 and total must equal the principal.', async () => {});

    let response;

    await test.step('What we use: We send a request with annualInterestRate=0.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 0,
          time: 5,
          timeUnit: 'years'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: interest must be 0 and totalAmount must equal principal.', async () => {
      const json = await response.json();
      expect(json.interest).toBe(0);
      expect(json.totalAmount).toBe(100000);
      expect(json.breakdown.principalPercent).toBe(100);
      expect(json.breakdown.interestPercent).toBe(0);
    });

    await test.step('Why it got this output: The backend multiplied by a zero rate and returned interest = 0, with the total unchanged from the principal.', async () => {});
  });

  test('Persona: Successful Simple Interest (large principal near cap)', async ({ request }) => {
    await test.step('Why we use this test: We want to sanity-check that the backend handles a large principal without overflow or precision loss.', async () => {});

    let response;

    await test.step('What we use: We send principal=1,000,000,000 (1 billion), annualInterestRate=10, time=10 years.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 1000000000,
          annualInterestRate: 10,
          time: 10,
          timeUnit: 'years'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: SI = (1e9 Г— 10 Г— 10) / 100 = 1e9, Total = 2e9.', async () => {
      const json = await response.json();
      expect(json.interest).toBeCloseTo(1000000000, 2);
      expect(json.totalAmount).toBeCloseTo(2000000000, 2);
    });

    await test.step('Why it got this output: The float64 math handled the large values without loss of precision.', async () => {});
  });

  // ================================================================
  // GROUP 2: RESPONSE SHAPE CORRECTNESS
  // ================================================================

  test('Persona: Response shape contains all required fields', async ({ request }) => {
    await test.step('Why we use this test: We want to confirm the API contract includes every field the frontend depends on.', async () => {});

    let response;

    await test.step('What we use: We send a standard request.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Every required field is present and non-null.', async () => {
      const json = await response.json();
      expect(json).toHaveProperty('principal');
      expect(json).toHaveProperty('interest');
      expect(json).toHaveProperty('totalAmount');
      expect(json).toHaveProperty('annualInterestRate');
      expect(json).toHaveProperty('timeYears');
      expect(json).toHaveProperty('breakdown');
      expect(json.breakdown).toHaveProperty('principalPercent');
      expect(json.breakdown).toHaveProperty('interestPercent');
      expect(Array.isArray(json.steps)).toBe(true);
      expect(json.steps.length).toBeGreaterThan(0);
    });

    await test.step('Why it got this output: The handler marshals the full SimpleInterestResponse struct, guaranteeing all fields are emitted.', async () => {});
  });

  test('Persona: Breakdown percentages sum to exactly 100', async ({ request }) => {
    await test.step('Why we use this test: A critical internal-consistency check вҖ” the principal/interest split must always sum to 100.', async () => {});

    let response;

    await test.step('What we use: We send a standard request with non-trivial interest.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years'
        }
      });
    });

    await test.step('What we expected: A 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The two percentages sum to 100 (within rounding tolerance).', async () => {
      const json = await response.json();
      const sum = json.breakdown.principalPercent + json.breakdown.interestPercent;
      expect(sum).toBeCloseTo(100, 1);
    });

    await test.step('Why it got this output: The backend computes principal percent first, then derives interest percent as 100 вҲ’ principal percent to absorb rounding.', async () => {});
  });

  test('Persona: Total amount reconciles with principal + interest', async ({ request }) => {
    await test.step('Why we use this test: The total must always equal principal + interest вҖ” this is the core arithmetic invariant of simple interest.', async () => {});

    let response;

    await test.step('What we use: We send a request with a fractional interest rate.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 250000,
          annualInterestRate: 7.25,
          time: 3,
          timeUnit: 'years'
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

    await test.step('Why it got this output: The backend computes total as principal + interest using the same rounded interest value.', async () => {});
  });

  // ================================================================
  // GROUP 3: VALIDATION ERRORS
  // ================================================================

  test('Persona: Error - missing body (empty JSON)', async ({ request }) => {
    await test.step('Why we use this test: An empty request body must be rejected with a clear 400 error.', async () => {});

    let response;

    await test.step('What we use: We POST an empty JSON body.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, { data: {} });
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

  test('Persona: Error - negative principal', async ({ request }) => {
    await test.step('Why we use this test: A negative principal is nonsensical and must be rejected.', async () => {});

    let response;

    await test.step('What we use: We send principal = -100000.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: -100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years'
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

    await test.step('Why it got this output: The validator explicitly rejects non-positive principal.', async () => {});
  });

  test('Persona: Error - zero principal', async ({ request }) => {
    await test.step('Why we use this test: A zero principal yields zero interest but is still an invalid input.', async () => {});

    let response;

    await test.step('What we use: We send principal = 0.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 0,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'years'
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

    await test.step('Why it got this output: The validator rejects zero as well as negative principals.', async () => {});
  });

  test('Persona: Error - negative interest rate', async ({ request }) => {
    await test.step('Why we use this test: A negative annual rate is not a valid loan/investment.', async () => {});

    let response;

    await test.step('What we use: We send annualInterestRate = -1.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: -1,
          time: 5,
          timeUnit: 'years'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions annualInterestRate cannot be negative.', async () => {
      const txt = await response.text();
      expect(txt).toContain('annualInterestRate cannot be negative');
    });

    await test.step('Why it got this output: The validator explicitly rejects negative rates.', async () => {});
  });

  test('Persona: Error - interest rate above 100%', async ({ request }) => {
    await test.step('Why we use this test: A rate above 100% is outside the supported range and must be rejected.', async () => {});

    let response;

    await test.step('What we use: We send annualInterestRate = 150.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 150,
          time: 5,
          timeUnit: 'years'
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

    await test.step('Why it got this output: The validator enforces the 0вҖ“100 range.', async () => {});
  });

  test('Persona: Error - zero time', async ({ request }) => {
    await test.step('Why we use this test: A zero tenure makes no sense for an interest calculation.', async () => {});

    let response;

    await test.step('What we use: We send time = 0.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 0,
          timeUnit: 'years'
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

    await test.step('Why it got this output: The normalizer rejects non-positive time.', async () => {});
  });

  test('Persona: Error - invalid timeUnit', async ({ request }) => {
    await test.step('Why we use this test: Only "years", "months", and "days" are valid time units.', async () => {});

    let response;

    await test.step('What we use: We send timeUnit = "weeks".', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 5,
          timeUnit: 'weeks'
        }
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the allowed time units.', async () => {
      const txt = await response.text();
      expect(txt).toContain("timeUnit must be 'years', 'months', or 'days'");
    });

    await test.step('Why it got this output: The validator checks the unit against a whitelist.', async () => {});
  });

  test('Persona: Error - tenure exceeds maximum (200 years)', async ({ request }) => {
    await test.step('Why we use this test: Extremely long tenures must be rejected to prevent overflow and absurd results.', async () => {});

    let response;

    await test.step('What we use: We send time = 200 years.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
        data: {
          principal: 100000,
          annualInterestRate: 8.5,
          time: 200,
          timeUnit: 'years'
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

    await test.step('Why it got this output: The normalizer caps tenure at MaxInterestTimeYears (100).', async () => {});
  });

  test('Persona: Error - malformed JSON body', async ({ request }) => {
    await test.step('Why we use this test: A malformed JSON payload must be rejected before any processing happens.', async () => {});

    let response;

    await test.step('What we use: We send a raw non-JSON string as the body.', async () => {
      response = await request.post(`${API_URL}/calculate-simple-interest`, {
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

    await test.step('What we use: GET /calculate-simple-interest.', async () => {
      response = await request.get(`${API_URL}/calculate-simple-interest`);
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