import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8080';

test.describe('Backend API Tax / VAT / GST Calculator Persona', () => {

  // ================================================================
  // add_tax
  // ================================================================

  test('Persona: Add GST to a base price (exclusive → inclusive)', async ({ request }) => {
    await test.step('Why we use this test: The most common tax operation — add a rate to a net amount to get the gross.', async () => {});

    let response;
    await test.step('What we use: POST /calculate-tax with mode=add_tax, amount=1000, taxRate=18.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'add_tax', amount: 1000, taxRate: 18 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: net=1000, tax=180, gross=1180.', async () => {
      const json = await response.json();
      expect(json.mode).toBe('add_tax');
      expect(json.netAmount).toBe(1000);
      expect(json.taxAmount).toBe(180);
      expect(json.grossAmount).toBe(1180);
      expect(json.formatted).toBe('1180');
      expect(Array.isArray(json.steps)).toBe(true);
    });

    await test.step('Why it got this output: tax = 18% of 1000 = 180; gross = 1000 + 180.', async () => {});
  });

  test('Persona: Add VAT with 0% rate (no-op)', async ({ request }) => {
    await test.step('Why we use this test: A zero rate is a valid edge case — gross must equal net.', async () => {});

    let response;
    await test.step('What we use: POST with mode=add_tax, amount=500, taxRate=0.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'add_tax', amount: 500, taxRate: 0 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: net=500, tax=0, gross=500.', async () => {
      const json = await response.json();
      expect(json.netAmount).toBe(500);
      expect(json.taxAmount).toBe(0);
      expect(json.grossAmount).toBe(500);
    });

    await test.step('Why it got this output: 0% of anything is 0, so gross = net.', async () => {});
  });

  // ================================================================
  // remove_tax
  // ================================================================

  test('Persona: Remove GST from a tax-inclusive price', async ({ request }) => {
    await test.step('Why we use this test: Users often know the final (inclusive) price and need the pre-tax value.', async () => {});

    let response;
    await test.step('What we use: POST with mode=remove_tax, amount=1180, taxRate=18.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'remove_tax', amount: 1180, taxRate: 18 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: net ≈ 1000, tax ≈ 180.', async () => {
      const json = await response.json();
      expect(json.netAmount).toBeCloseTo(1000, 2);
      expect(json.taxAmount).toBeCloseTo(180, 2);
    });

    await test.step('Why it got this output: net = 1180 / 1.18 = 1000.', async () => {});
  });

  test('Persona: Error - remove_tax with zero rate', async ({ request }) => {
    await test.step('Why we use this test: Removing a 0% tax is a divide-by-1, which means "remove" is meaningless — the backend must reject it.', async () => {});

    let response;
    await test.step('What we use: POST mode=remove_tax with taxRate=0.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'remove_tax', amount: 1000, taxRate: 0 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions taxRate must be greater than zero.', async () => {
      const txt = await response.text();
      expect(txt).toContain('taxRate must be greater than zero');
    });

    await test.step('Why it got this output: The converter guards against nonsensical zero-rate removals.', async () => {});
  });

  // ================================================================
  // find_rate
  // ================================================================

  test('Persona: Find the effective tax rate from net and gross', async ({ request }) => {
    await test.step('Why we use this test: Recovering the rate when only the two amounts are known.', async () => {});

    let response;
    await test.step('What we use: POST with mode=find_rate, netAmount=1000, grossAmount=1180.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'find_rate', netAmount: 1000, grossAmount: 1180 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: taxRate = 18%, taxAmount = 180.', async () => {
      const json = await response.json();
      expect(json.taxRate).toBeCloseTo(18, 2);
      expect(json.taxAmount).toBe(180);
      expect(json.netAmount).toBe(1000);
      expect(json.grossAmount).toBe(1180);
    });

    await test.step('Why it got this output: rate = ((1180 − 1000) / 1000) X 100 = 18%.', async () => {});
  });

  test('Persona: Error - find_rate with gross less than net', async ({ request }) => {
    await test.step('Why we use this test: A gross smaller than net implies a negative tax rate, which is nonsensical.', async () => {});

    let response;
    await test.step('What we use: POST mode=find_rate with net=1000, gross=900.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'find_rate', netAmount: 1000, grossAmount: 900 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions gross must be ≥ net.', async () => {
      const txt = await response.text();
      expect(txt).toContain('grossAmount must be greater than or equal to netAmount');
    });

    await test.step('Why it got this output: The validator checks the ordering before computing.', async () => {});
  });

  // ================================================================
  // split_gst
  // ================================================================

  test('Persona: Split GST into CGST + SGST (intra-state)', async ({ request }) => {
    await test.step('Why we use this test: Intra-state GST is always split 50/50 between CGST and SGST — a core Indian tax requirement.', async () => {});

    let response;
    await test.step('What we use: POST mode=split_gst, amount=1180, taxRate=18, taxType=cgst_sgst.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'split_gst', amount: 1180, taxRate: 18, taxType: 'cgst_sgst' },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: net=1000, total tax=180, CGST=90, SGST=90.', async () => {
      const json = await response.json();
      expect(json.netAmount).toBeCloseTo(1000, 2);
      expect(json.taxAmount).toBeCloseTo(180, 2);
      expect(json.extra.cgst).toBeCloseTo(90, 2);
      expect(json.extra.sgst).toBeCloseTo(90, 2);
      expect(json.extra.taxType).toBe('cgst_sgst');
    });

    await test.step('Why it got this output: The backend split the 180 GST evenly between CGST and SGST.', async () => {});
  });

  test('Persona: Split GST into IGST (inter-state)', async ({ request }) => {
    await test.step('Why we use this test: Inter-state GST is a single IGST bucket — no split.', async () => {});

    let response;
    await test.step('What we use: POST mode=split_gst, amount=1180, taxRate=18, taxType=igst.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'split_gst', amount: 1180, taxRate: 18, taxType: 'igst' },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: net=1000, IGST=180, CGST=0, SGST=0.', async () => {
      const json = await response.json();
      expect(json.netAmount).toBeCloseTo(1000, 2);
      expect(json.extra.igst).toBeCloseTo(180, 2);
      expect(json.extra.cgst).toBe(0);
      expect(json.extra.sgst).toBe(0);
      expect(json.extra.taxType).toBe('igst');
    });

    await test.step('Why it got this output: The backend routed the entire GST into the IGST bucket for inter-state supplies.', async () => {});
  });

  test('Persona: Error - split_gst with invalid taxType', async ({ request }) => {
    await test.step('Why we use this test: Only "cgst_sgst" and "igst" are valid taxTypes.', async () => {});

    let response;
    await test.step('What we use: POST mode=split_gst with taxType="vat".', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'split_gst', amount: 1180, taxRate: 18, taxType: 'vat' },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the allowed taxTypes.', async () => {
      const txt = await response.text();
      expect(txt).toContain("taxType must be 'cgst_sgst' or 'igst'");
    });

    await test.step('Why it got this output: The validator rejects unknown taxType values.', async () => {});
  });

  // ================================================================
  // reverse_gst
  // ================================================================

  test('Persona: Reverse GST from tax paid', async ({ request }) => {
    await test.step('Why we use this test: Users may know only the GST amount and the rate, and need the original taxable value.', async () => {});

    let response;
    await test.step('What we use: POST mode=reverse_gst, taxPaid=180, taxRate=18.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'reverse_gst', taxPaid: 180, taxRate: 18 },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: net=1000, gross=1180.', async () => {
      const json = await response.json();
      expect(json.netAmount).toBeCloseTo(1000, 2);
      expect(json.grossAmount).toBeCloseTo(1180, 2);
      expect(json.taxAmount).toBe(180);
    });

    await test.step('Why it got this output: net = 180 / 0.18 = 1000, then gross = net + tax.', async () => {});
  });

  // ================================================================
  // income_tax (slab-based)
  // ================================================================

  test('Persona: Income tax with progressive slabs', async ({ request }) => {
    await test.step('Why we use this test: Progressive slab tax is the hallmark of income tax — each bracket applies only to income within it.', async () => {});

    let response;
    await test.step('What we use: POST mode=income_tax with income=900000 and three slabs: 0–3L@0%, 3–6L@5%, 6L+@10%.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: {
          mode: 'income_tax',
          income: 900000,
          slabs: [
            { from: 0,      to: 300000, rate: 0 },
            { from: 300000, to: 600000, rate: 5 },
            { from: 600000, to: 0,      rate: 10 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: total tax = 45000, effective rate = 5%.', async () => {
      const json = await response.json();
      // 0 + 5%*300000 + 10%*300000 = 0 + 15000 + 30000 = 45000
      expect(json.taxAmount).toBe(45000);
      expect(json.taxRate).toBeCloseTo(5, 2);
      expect(Array.isArray(json.extra.slabBreakdown)).toBe(true);
      expect(json.extra.slabBreakdown).toHaveLength(3);
    });

    await test.step('Why it got this output: The backend walked the slabs and taxed only the portion of income that fell inside each bracket.', async () => {});
  });

  test('Persona: Income tax below first slab', async ({ request }) => {
    await test.step('Why we use this test: Income below the lowest taxable bracket should result in zero tax.', async () => {});

    let response;
    await test.step('What we use: POST mode=income_tax with income=200000 and a 3L exemption slab.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: {
          mode: 'income_tax',
          income: 200000,
          slabs: [
            { from: 0,      to: 300000, rate: 0 },
            { from: 300000, to: 0,      rate: 10 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: tax = 0, effective rate = 0%.', async () => {
      const json = await response.json();
      expect(json.taxAmount).toBe(0);
      expect(json.taxRate).toBe(0);
    });

    await test.step('Why it got this output: The income never crossed into the taxable slab.', async () => {});
  });

  test('Persona: Income tax handles unsorted slabs', async ({ request }) => {
    await test.step('Why we use this test: Clients may send slabs out of order — the backend must sort them before applying.', async () => {});

    let response;
    await test.step('What we use: POST mode=income_tax with intentionally unsorted slabs.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: {
          mode: 'income_tax',
          income: 900000,
          slabs: [
            { from: 600000, to: 0,      rate: 10 },
            { from: 0,      to: 300000, rate: 0 },
            { from: 300000, to: 600000, rate: 5 },
          ],
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The same result as if the slabs had been sorted — total tax = 45000.', async () => {
      const json = await response.json();
      expect(json.taxAmount).toBe(45000);
    });

    await test.step('Why it got this output: The backend sorts a copy of the slab array before applying it.', async () => {});
  });

  test('Persona: Error - income_tax with no slabs', async ({ request }) => {
    await test.step('Why we use this test: The income_tax mode is meaningless without a slab schedule.', async () => {});

    let response;
    await test.step('What we use: POST mode=income_tax with income=500000 and no slabs.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'income_tax', income: 500000 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions slabs are required.', async () => {
      const txt = await response.text();
      expect(txt).toContain('slabs are required for income_tax');
    });

    await test.step('Why it got this output: The validator rejects an empty slab array.', async () => {});
  });

  test('Persona: Error - income_tax with overlapping slabs', async ({ request }) => {
    await test.step('Why we use this test: Overlapping slabs would double-tax the overlapping income — the backend must reject them.', async () => {});

    let response;
    await test.step('What we use: POST mode=income_tax with slabs that overlap (0–5L and 3L–∞).', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: {
          mode: 'income_tax',
          income: 400000,
          slabs: [
            { from: 0,      to: 500000, rate: 5 },
            { from: 300000, to: 0,      rate: 10 },
          ],
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions slabs must not overlap.', async () => {
      const txt = await response.text();
      expect(txt).toContain('slabs must not overlap');
    });

    await test.step('Why it got this output: The validator detects the overlap and aborts before any tax is computed.', async () => {});
  });

  test('Persona: Error - income_tax with multiple open-ended slabs', async ({ request }) => {
    await test.step('Why we use this test: Only the final slab may have "to = 0" (open-ended) — otherwise income has no well-defined top rate.', async () => {});

    let response;
    await test.step('What we use: POST mode=income_tax with two open-ended slabs.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: {
          mode: 'income_tax',
          income: 200000,
          slabs: [
            { from: 0,      to: 0, rate: 5 },
            { from: 100000, to: 0, rate: 10 },
          ],
        },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions only one open-ended slab is allowed.', async () => {
      const txt = await response.text();
      expect(txt).toContain('only one slab may be open-ended');
    });

    await test.step('Why it got this output: The validator counts open-ended slabs and rejects more than one.', async () => {});
  });

  // ================================================================
  // Cross-cutting: mode + malformed input
  // ================================================================

  test('Persona: Error - unsupported mode', async ({ request }) => {
    await test.step('Why we use this test: An unknown mode must be rejected cleanly.', async () => {});

    let response;
    await test.step('What we use: POST with mode="magic".', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'magic', amount: 100, taxRate: 18 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions "unsupported mode".', async () => {
      const txt = await response.text();
      expect(txt).toContain('unsupported mode');
    });

    await test.step('Why it got this output: The dispatcher switch falls through to the default case.', async () => {});
  });

  test('Persona: Error - empty mode', async ({ request }) => {
    await test.step('Why we use this test: An empty body without a mode is a client bug — the backend must say so.', async () => {});

    let response;
    await test.step('What we use: POST with an empty JSON body.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, { data: {} });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions "mode is required".', async () => {
      const txt = await response.text();
      expect(txt).toContain('mode is required');
    });

    await test.step('Why it got this output: The dispatcher validates the mode before doing anything else.', async () => {});
  });

  test('Persona: Error - negative amount', async ({ request }) => {
    await test.step('Why we use this test: A negative amount is nonsensical for a tax calculation.', async () => {});

    let response;
    await test.step('What we use: POST mode=add_tax with amount=-1.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'add_tax', amount: -1, taxRate: 18 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions amount must be ≥ 0.', async () => {
      const txt = await response.text();
      expect(txt).toContain('amount must be greater than or equal to zero');
    });

    await test.step('Why it got this output: The validator rejects negative amounts.', async () => {});
  });

  test('Persona: Error - rate above 100%', async ({ request }) => {
    await test.step('Why we use this test: A tax rate above 100% is outside the supported range and must be rejected.', async () => {});

    let response;
    await test.step('What we use: POST mode=add_tax with taxRate=150.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        data: { mode: 'add_tax', amount: 100, taxRate: 150 },
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions taxRate must be between 0 and 100.', async () => {
      const txt = await response.text();
      expect(txt).toContain('taxRate must be between 0 and 100');
    });

    await test.step('Why it got this output: The validator enforces the 0–100 range.', async () => {});
  });

  test('Persona: Error - malformed JSON body', async ({ request }) => {
    await test.step('Why we use this test: A syntactically invalid body must be caught before any processing.', async () => {});

    let response;
    await test.step('What we use: POST with a raw non-JSON string.', async () => {
      response = await request.post(`${API_URL}/calculate-tax`, {
        headers: { 'Content-Type': 'application/json' },
        data: 'this is not valid json',
      });
    });

    await test.step('What we expected: 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions "Invalid JSON".', async () => {
      const txt = await response.text();
      expect(txt).toContain('Invalid JSON');
    });

    await test.step('Why it got this output: The JSON decoder fails and the handler returns a 400 immediately.', async () => {});
  });

  test('Persona: Error - wrong HTTP method (GET)', async ({ request }) => {
    await test.step('Why we use this test: The endpoint only accepts POST — GET must be rejected.', async () => {});

    let response;
    await test.step('What we use: GET /calculate-tax.', async () => {
      response = await request.get(`${API_URL}/calculate-tax`);
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