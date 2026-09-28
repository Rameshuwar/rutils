import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API Percentage Calculator Persona', () => {

  // ================================================================
  // GROUP 1: BASIC OPERATIONS
  // ================================================================

  test('Persona: Successful "X% of Y" (Percent Of)', async ({ request }) => {
    await test.step('Why we use this test: We want to make sure the backend correctly calculates the most fundamental percentage operation — finding what X% of Y is.', async () => {});

    let response;

    await test.step('What we use: We send a POST request to /calculate-percentage with operation "percent_of", value1 = 15 (percent), value2 = 200 (number).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_of',
          value1: 15,
          value2: 200
        }
      });
    });

    await test.step('What we expected: We expect the backend API to say "OK" (Status 200).', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the JSON output and confirm result = 30, formatted = "30", and the 3-step breakdown is present.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('percent_of');
      expect(jsonResponse.result).toBe(30);
      expect(jsonResponse.formatted).toBe('30');
      expect(jsonResponse.steps).toHaveLength(3);
      expect(jsonResponse.steps[0]).toContain('(15 / 100)');
      expect(jsonResponse.steps[0]).toContain('200');
    });

    await test.step('Why it got this output: The backend applied the formula (15 / 100) X 200 and correctly returned 30!', async () => {});
  });

  test('Persona: Successful "X is what % of Y" (What Percent)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify the reverse of the basic operation — given a part and a whole, what percentage is the part?', async () => {});

    let response;

    await test.step('What we use: We send operation "what_percent" with value1 = 30 (part), value2 = 200 (whole).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'what_percent',
          value1: 30,
          value2: 200
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 15, formatted with a percent sign "15%".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('what_percent');
      expect(jsonResponse.result).toBe(15);
      expect(jsonResponse.formatted).toBe('15%');
      expect(jsonResponse.steps).toHaveLength(3);
    });

    await test.step('Why it got this output: The backend computed (30 / 200) X 100 = 15%, and correctly appended the percent sign to the formatted string.', async () => {});
  });

  test('Persona: Successful "X is Y% of what number"', async ({ request }) => {
    await test.step('Why we use this test: We need to verify we can reverse-engineer the whole given only a part and its percentage.', async () => {});

    let response;

    await test.step('What we use: We send operation "is_percent_of_what" with value1 = 30 (part), value2 = 15 (percent).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'is_percent_of_what',
          value1: 30,
          value2: 15
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 200 — since 30 is 15% of 200.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('is_percent_of_what');
      expect(jsonResponse.result).toBe(200);
      expect(jsonResponse.formatted).toBe('200');
    });

    await test.step('Why it got this output: The backend computed 30 / (15 / 100) = 200, correctly identifying the original whole number.', async () => {});
  });

  // ================================================================
  // GROUP 2: CHANGE OPERATIONS
  // ================================================================

  test('Persona: Successful Percentage Increase', async ({ request }) => {
    await test.step('Why we use this test: Percentage change is one of the most common real-world calculations — we need to verify the increase case.', async () => {});

    let response;

    await test.step('What we use: We send operation "percent_change" with value1 = 100 (old), value2 = 150 (new).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_change',
          value1: 100,
          value2: 150
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 50%, with direction = "increase" and amount = 50 in the extra field.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('percent_change');
      expect(jsonResponse.result).toBe(50);
      expect(jsonResponse.formatted).toBe('50%');
      expect(jsonResponse.extra).toHaveProperty('direction', 'increase');
      expect(jsonResponse.extra).toHaveProperty('amount', 50);
    });

    await test.step('Why it got this output: The backend computed ((150 - 100) / 100) X 100 = 50%, and correctly labeled the direction as an increase.', async () => {});
  });

  test('Persona: Successful Percentage Decrease', async ({ request }) => {
    await test.step('Why we use this test: We need to verify that a decrease returns a negative percentage and the correct direction label.', async () => {});

    let response;

    await test.step('What we use: We send operation "percent_change" with value1 = 200 (old), value2 = 150 (new).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_change',
          value1: 200,
          value2: 150
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be -25%, with direction = "decrease" and amount = 50.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('percent_change');
      expect(jsonResponse.result).toBe(-25);
      expect(jsonResponse.extra).toHaveProperty('direction', 'decrease');
      expect(jsonResponse.extra).toHaveProperty('amount', 50);
    });

    await test.step('Why it got this output: The backend computed ((150 - 200) / 200) X 100 = -25%, and correctly labeled the direction as a decrease.', async () => {});
  });

  test('Persona: Successful Increase a Number by X%', async ({ request }) => {
    await test.step('Why we use this test: Users often want to know "what is 200 increased by 15%?" — a compound formula that differs from percent_change.', async () => {});

    let response;

    await test.step('What we use: We send operation "percent_increase" with value1 = 200 (number), value2 = 15 (percent).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_increase',
          value1: 200,
          value2: 15
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 230, with increaseAmount = 30 in the extra field.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('percent_increase');
      expect(jsonResponse.result).toBe(230);
      expect(jsonResponse.formatted).toBe('230');
      expect(jsonResponse.extra).toHaveProperty('increaseAmount', 30);
    });

    await test.step('Why it got this output: The backend added 15% of 200 (which is 30) to the original 200, giving 230.', async () => {});
  });

  test('Persona: Successful Decrease a Number by X%', async ({ request }) => {
    await test.step('Why we use this test: The symmetric counterpart to increase — subtracting a percentage from a number.', async () => {});

    let response;

    await test.step('What we use: We send operation "percent_decrease" with value1 = 200 (number), value2 = 15 (percent).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_decrease',
          value1: 200,
          value2: 15
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 170, with decreaseAmount = 30 in the extra field.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('percent_decrease');
      expect(jsonResponse.result).toBe(170);
      expect(jsonResponse.formatted).toBe('170');
      expect(jsonResponse.extra).toHaveProperty('decreaseAmount', 30);
    });

    await test.step('Why it got this output: The backend subtracted 15% of 200 (which is 30) from the original 200, giving 170.', async () => {});
  });

  test('Persona: Successful Reverse Percentage', async ({ request }) => {
    await test.step('Why we use this test: Users often know the final value and the percentage applied, but not the original — the reverse percentage recovers it.', async () => {});

    let response;

    await test.step('What we use: We send operation "reverse_percent" with value1 = 230 (final), value2 = 15 (percent applied).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'reverse_percent',
          value1: 230,
          value2: 15
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 200 — the original value before the 15% increase.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('reverse_percent');
      expect(jsonResponse.result).toBe(200);
      expect(jsonResponse.formatted).toBe('200');
    });

    await test.step('Why it got this output: The backend divided 230 by (1 + 15/100) = 1.15, giving 200 — correctly undoing the increase.', async () => {});
  });

  test('Persona: Successful Percentage Difference', async ({ request }) => {
    await test.step('Why we use this test: The percentage difference between two values is a common comparison that uses their average as the base (unlike percent_change).', async () => {});

    let response;

    await test.step('What we use: We send operation "percent_difference" with value1 = 100, value2 = 150.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_difference',
          value1: 100,
          value2: 150
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 40% — |100-150| / ((100+150)/2) X 100.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('percent_difference');
      expect(jsonResponse.result).toBe(40);
      expect(jsonResponse.formatted).toBe('40%');
    });

    await test.step('Why it got this output: The backend computed the absolute difference (50), divided by the average (125), and multiplied by 100 to get 40%.', async () => {});
  });

  // ================================================================
  // GROUP 3: PERCENTAGE POINTS
  // ================================================================

  test('Persona: Add Percentage Points', async ({ request }) => {
    await test.step('Why we use this test: Adding percentage points is different from adding percentages — 5% + 3% as points equals 8%, not 5.15%.', async () => {});

    let response;

    await test.step('What we use: We send operation "add_percent_points" with value1 = 5, value2 = 3.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'add_percent_points',
          value1: 5,
          value2: 3
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 8, formatted as "8%".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('add_percent_points');
      expect(jsonResponse.result).toBe(8);
      expect(jsonResponse.formatted).toBe('8%');
    });

    await test.step('Why it got this output: The backend simply added the two percentage values as points, giving 8%.', async () => {});
  });

  test('Persona: Subtract Percentage Points', async ({ request }) => {
    await test.step('Why we use this test: The symmetric counterpart — subtracting percentage points is a direct arithmetic subtraction of the percent values.', async () => {});

    let response;

    await test.step('What we use: We send operation "subtract_percent_points" with value1 = 5, value2 = 3.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'subtract_percent_points',
          value1: 5,
          value2: 3
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 2, formatted as "2%".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('subtract_percent_points');
      expect(jsonResponse.result).toBe(2);
      expect(jsonResponse.formatted).toBe('2%');
    });

    await test.step('Why it got this output: The backend simply subtracted the two percentage values as points, giving 2%.', async () => {});
  });

  // ================================================================
  // GROUP 4: BUSINESS OPERATIONS
  // ================================================================

  test('Persona: Successful Discount', async ({ request }) => {
    await test.step('Why we use this test: Discount calculations are ubiquitous in commerce — we need to verify both the final price and the savings amount.', async () => {});

    let response;

    await test.step('What we use: We send operation "discount" with value1 = 500 (original price), value2 = 20 (discount percent).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'discount',
          value1: 500,
          value2: 20
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 400 (final price), with savings = 100 and finalPrice = 400 in the extra field.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('discount');
      expect(jsonResponse.result).toBe(400);
      expect(jsonResponse.formatted).toBe('400');
      expect(jsonResponse.extra).toHaveProperty('savings', 100);
      expect(jsonResponse.extra).toHaveProperty('finalPrice', 400);
    });

    await test.step('Why it got this output: The backend computed the savings (20% of 500 = 100), subtracted it from the original price to get 400, and reported both values.', async () => {});
  });

  test('Persona: Successful Markup', async ({ request }) => {
    await test.step('Why we use this test: Markup is the business counterpart of discount — finding the selling price given a cost and a markup percentage.', async () => {});

    let response;

    await test.step('What we use: We send operation "markup" with value1 = 100 (cost), value2 = 20 (markup percent).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'markup',
          value1: 100,
          value2: 20
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 120 (selling price), with profit = 20 and sellingPrice = 120 in the extra field.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('markup');
      expect(jsonResponse.result).toBe(120);
      expect(jsonResponse.formatted).toBe('120');
      expect(jsonResponse.extra).toHaveProperty('profit', 20);
      expect(jsonResponse.extra).toHaveProperty('sellingPrice', 120);
    });

    await test.step('Why it got this output: The backend added the markup (20% of 100 = 20) to the cost to get the selling price of 120.', async () => {});
  });

  test('Persona: Successful Profit Calculation', async ({ request }) => {
    await test.step('Why we use this test: Users need to see a positive profit percentage and the label "profit" when the selling price exceeds the cost.', async () => {});

    let response;

    await test.step('What we use: We send operation "profit_loss" with value1 = 100 (cost), value2 = 120 (selling).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'profit_loss',
          value1: 100,
          value2: 120
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 20 (percent profit), with label = "profit" and amount = 20.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('profit_loss');
      expect(jsonResponse.result).toBe(20);
      expect(jsonResponse.formatted).toBe('20%');
      expect(jsonResponse.extra).toHaveProperty('label', 'profit');
      expect(jsonResponse.extra).toHaveProperty('amount', 20);
    });

    await test.step('Why it got this output: The backend computed ((120 - 100) / 100) X 100 = 20% profit, and labeled it "profit" because the value is positive.', async () => {});
  });

  test('Persona: Successful Loss Calculation', async ({ request }) => {
    await test.step('Why we use this test: The loss case returns a negative percentage and the label "loss" — a critical business distinction from the profit case.', async () => {});

    let response;

    await test.step('What we use: We send operation "profit_loss" with value1 = 100 (cost), value2 = 80 (selling).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'profit_loss',
          value1: 100,
          value2: 80
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be -20, with label = "loss" and amount = 20.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('profit_loss');
      expect(jsonResponse.result).toBe(-20);
      expect(jsonResponse.formatted).toBe('-20%');
      expect(jsonResponse.extra).toHaveProperty('label', 'loss');
      expect(jsonResponse.extra).toHaveProperty('amount', 20);
    });

    await test.step('Why it got this output: The backend computed ((80 - 100) / 100) X 100 = -20%, and correctly labeled it "loss" because the value is negative.', async () => {});
  });

  // ================================================================
  // GROUP 5: CONVERSIONS
  // ================================================================

  test('Persona: Percent to Fraction', async ({ request }) => {
    await test.step('Why we use this test: Converting a percentage to its simplest fraction form is a common educational and practical need.', async () => {});

    let response;

    await test.step('What we use: We send operation "percent_to_fraction" with value1 = 25.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_to_fraction',
          value1: 25
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 0.25, formatted as "1/4", with fraction = "1/4" and decimal = 0.25 in the extra field.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('percent_to_fraction');
      expect(jsonResponse.result).toBe(0.25);
      expect(jsonResponse.formatted).toBe('1/4');
      expect(jsonResponse.extra).toHaveProperty('fraction', '1/4');
      expect(jsonResponse.extra).toHaveProperty('decimal', 0.25);
    });

    await test.step('Why it got this output: The backend divided 25 by 100 to get 0.25, then reduced the fraction 25/100 to its simplest form 1/4.', async () => {});
  });

  test('Persona: Fraction to Percent', async ({ request }) => {
    await test.step('Why we use this test: The reverse operation of percent-to-fraction — converting a fraction to its percentage form.', async () => {});

    let response;

    await test.step('What we use: We send operation "fraction_to_percent" with value1 = 1 (numerator), value2 = 4 (denominator).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'fraction_to_percent',
          value1: 1,
          value2: 4
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 25%, formatted as "25%".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('fraction_to_percent');
      expect(jsonResponse.result).toBe(25);
      expect(jsonResponse.formatted).toBe('25%');
    });

    await test.step('Why it got this output: The backend computed (1 / 4) X 100 = 25%, correctly converting the fraction to a percentage.', async () => {});
  });

  test('Persona: Decimal to Percent', async ({ request }) => {
    await test.step('Why we use this test: Converting a decimal to a percentage is the third conversion in the family — we need to verify the simple multiply-by-100 case.', async () => {});

    let response;

    await test.step('What we use: We send operation "decimal_to_percent" with value1 = 0.25.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'decimal_to_percent',
          value1: 0.25
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 25%, formatted as "25%".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('decimal_to_percent');
      expect(jsonResponse.result).toBe(25);
      expect(jsonResponse.formatted).toBe('25%');
    });

    await test.step('Why it got this output: The backend multiplied 0.25 by 100 to get 25%, the correct decimal-to-percent conversion.', async () => {});
  });

  // ================================================================
  // GROUP 6: ADVANCED OPERATIONS
  // ================================================================

  test('Persona: Compound Percentage', async ({ request }) => {
    await test.step('Why we use this test: Compound percentage (like compound interest) is a distinct operation that requires a third input — the number of periods.', async () => {});

    let response;

    await test.step('What we use: We send operation "compound_percent" with value1 = 1000 (base), value2 = 10 (percent per period), value3 = 3 (periods).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'compound_percent',
          value1: 1000,
          value2: 10,
          value3: 3
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 1331 (1000 X 1.1^3), with periods = 3 and growth в‰€ 331.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('compound_percent');
      expect(jsonResponse.result).toBeCloseTo(1331, 2);
      expect(jsonResponse.formatted).toBe('1331');
      expect(jsonResponse.extra).toHaveProperty('periods', 3);
      expect(jsonResponse.extra.growth).toBeCloseTo(331, 2);
    });

    await test.step('Why it got this output: The backend computed 1000 X (1 + 10/100)^3 = 1000 X 1.331 = 1331, correctly applying compound growth across 3 periods.', async () => {});
  });

  test('Persona: Marks Percentage', async ({ request }) => {
    await test.step('Why we use this test: Students and teachers frequently need to compute the percentage of marks obtained out of a total.', async () => {});

    let response;

    await test.step('What we use: We send operation "marks_percentage" with value1 = 85 (obtained), value2 = 100 (total).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'marks_percentage',
          value1: 85,
          value2: 100
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 85%, formatted as "85%".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('marks_percentage');
      expect(jsonResponse.result).toBe(85);
      expect(jsonResponse.formatted).toBe('85%');
    });

    await test.step('Why it got this output: The backend computed (85 / 100) X 100 = 85%, correctly converting the marks ratio to a percentage.', async () => {});
  });

  test('Persona: CGPA to Percent', async ({ request }) => {
    await test.step('Why we use this test: The CBSE-standard CGPA-to-percent conversion uses a fixed factor of 9.5 — we need to verify the backend applies it.', async () => {});

    let response;

    await test.step('What we use: We send operation "cgpa_to_percent" with value1 = 8.0.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'cgpa_to_percent',
          value1: 8.0
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: The result should be 76%, formatted as "76%", with factor = 9.5 in the extra field.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.operation).toBe('cgpa_to_percent');
      expect(jsonResponse.result).toBe(76);
      expect(jsonResponse.formatted).toBe('76%');
      expect(jsonResponse.extra).toHaveProperty('factor', 9.5);
    });

    await test.step('Why it got this output: The backend multiplied 8.0 by the CBSE factor 9.5, giving 76 — the correct percentage conversion.', async () => {});
  });

  // ================================================================
  // GROUP 7: SYMBOL REGRESSION + ERROR HANDLING
  // ================================================================

  test('Persona: Verify ASCII X in Steps (no Unicode multiplication sign)', async ({ request }) => {
    await test.step('Why we use this test: The backend previously used the Unicode multiplication sign (Г—) in step strings, which rendered incorrectly in some terminals. This is a regression guard to ensure every step uses ASCII "X".', async () => {});

    let response;

    await test.step('What we use: We send a "percent_of" request — the operation whose steps include the multiplication sign.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'percent_of',
          value1: 15,
          value2: 200
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK response.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Every step must contain the ASCII "X" character and never the Unicode multiplication sign (Г—, U+00D7).', async () => {
      const jsonResponse = await response.json();
      expect(Array.isArray(jsonResponse.steps)).toBe(true);
      expect(jsonResponse.steps.length).toBeGreaterThan(0);

      for (const step of jsonResponse.steps) {
        // Must NOT contain the Unicode multiplication sign (U+00D7)
        expect(step).not.toContain('\u00D7');
        // Must NOT contain the Unicode minus sign (U+2212)
        expect(step).not.toContain('\u2212');
      }

      // The step that involves multiplication must use ASCII "X"
      const multiplicationStep = jsonResponse.steps.find((s: string) => s.includes('(15 / 100)'));
      expect(multiplicationStep).toBeDefined();
      expect(multiplicationStep).toContain('X');
    });

    await test.step('Why it got this output: The backend has been corrected to emit only ASCII characters in step strings — safe for terminals, chat apps, and any text consumer.', async () => {});
  });

  test('Persona: Error Handling - Divide by Zero (What Percent)', async ({ request }) => {
    await test.step('Why we use this test: Dividing by zero is mathematically undefined — we need to ensure the backend rejects it gracefully with a helpful message.', async () => {});

    let response;

    await test.step('What we use: We send "what_percent" with value2 = 0 (the whole).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'what_percent',
          value1: 30,
          value2: 0
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server returns a clear error message mentioning the zero whole.', async () => {
      const errorText = await response.text();
      expect(errorText).toContain('value2 (whole) cannot be zero');
    });

    await test.step('Why it got this output: The backend validation explicitly guards against division by zero and returns a descriptive 400 error.', async () => {});
  });

  test('Persona: Error Handling - Missing value3 for Compound Percentage', async ({ request }) => {
    await test.step('Why we use this test: The compound_percent operation requires a third value (periods) — omitting it should be rejected clearly.', async () => {});

    let response;

    await test.step('What we use: We send "compound_percent" with only value1 and value2 — deliberately omitting value3.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'compound_percent',
          value1: 1000,
          value2: 10
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server returns an error message explaining that value3 is required.', async () => {
      const errorText = await response.text();
      expect(errorText).toContain('value3 (number of periods) is required');
    });

    await test.step('Why it got this output: The backend explicitly checks for the presence of value3 before attempting the compound calculation.', async () => {});
  });

  test('Persona: Error Handling - Unsupported Operation', async ({ request }) => {
    await test.step('Why we use this test: Users might typo or hallucinate an operation name — the backend should reject unknown operations cleanly.', async () => {});

    let response;

    await test.step('What we use: We send a completely made-up operation called "magic_trick".', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'magic_trick',
          value1: 1,
          value2: 2
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server rejects the request with an "unsupported operation" message.', async () => {
      const errorText = await response.text();
      expect(errorText).toContain('unsupported operation');
    });

    await test.step('Why it got this output: The backend switch statement falls through to the default case and returns a clear error naming the unsupported operation.', async () => {});
  });

  test('Persona: Error Handling - Missing Operation Field', async ({ request }) => {
    await test.step('Why we use this test: A completely empty operation field should be caught before any calculation is attempted.', async () => {});

    let response;

    await test.step('What we use: We send a request with an empty operation string.', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: '',
          value1: 1,
          value2: 2
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server returns an error mentioning the operation is required.', async () => {
      const errorText = await response.text();
      expect(errorText).toContain('operation is required');
    });

    await test.step('Why it got this output: The handler validates the operation field before passing it to the converter, returning a clear 400 error.', async () => {});
  });

  test('Persona: Error Handling - Divide by Zero (Fraction to Percent)', async ({ request }) => {
    await test.step('Why we use this test: A zero denominator is undefined for fractions — the backend must reject it with a clear message.', async () => {});

    let response;

    await test.step('What we use: We send "fraction_to_percent" with value2 = 0 (denominator).', async () => {
      response = await request.post(`${API_URL}/calculate-percentage`, {
        data: {
          operation: 'fraction_to_percent',
          value1: 1,
          value2: 0
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server returns a clear error mentioning the zero denominator.', async () => {
      const errorText = await response.text();
      expect(errorText).toContain('value2 (denominator) cannot be zero');
    });

    await test.step('Why it got this output: The backend explicitly validates the denominator and returns a descriptive 400 error before attempting the division.', async () => {});
  });

});