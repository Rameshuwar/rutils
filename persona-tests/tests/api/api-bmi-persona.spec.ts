import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API BMI Calculator Persona', () => {

  test('Persona: Successful BMI Calculation (Normal Weight - Metric Units)', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure the backend API correctly calculates a normal BMI using standard metric units.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a POST request to the /calculate-bmi endpoint with a JSON body specifying weight in kg and height in meters.', async () => {
      response = await request.post(`${API_URL}/calculate-bmi`, {
        data: {
          weight: 70,
          weightUnit: 'kilograms',
          height: 1.75,
          heightUnit: 'meters'
        }
      });
    });

    await test.step('What we expected: We expect the backend API to say "OK" (Status 200).', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the JSON output returned by the backend.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse).toHaveProperty('bmi');
      expect(jsonResponse).toHaveProperty('category');
      
      // BMI = 70 / (1.75^2) = 22.857...
      expect(jsonResponse.bmi).toBeCloseTo(22.857, 2);
      expect(jsonResponse.category).toBe('Normal weight');
    });

    await test.step('Why it got this output: The backend successfully received valid metrics, applied the BMI formula, and correctly identified the normal weight category!', async () => {});
  });

  test('Persona: Successful BMI Calculation (Obesity - Imperial Units)', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure the backend API automatically converts imperial units to metric units before calculating BMI.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a POST request with weight in pounds and height in inches (which triggers Obesity category).', async () => {
      response = await request.post(`${API_URL}/calculate-bmi`, {
        data: {
          weight: 250, // lbs
          weightUnit: 'pounds',
          height: 65,  // inches
          heightUnit: 'inches'
        }
      });
    });

    await test.step('What we expected: We expect the backend API to say "OK" (Status 200).', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the JSON output and verify the calculation.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse).toHaveProperty('bmi');
      expect(jsonResponse).toHaveProperty('category');
      
      // BMI = (250 * 0.453592) / (65 * 0.0254)^2 = 113.398 / (1.651)^2 = 113.398 / 2.7258 = 41.6
      expect(jsonResponse.bmi).toBeGreaterThan(30);
      expect(jsonResponse.category).toBe('Obesity');
    });

    await test.step('Why it got this output: The backend successfully converted pounds to kg, inches to meters, and calculated the obese category BMI!', async () => {});
  });

  test('Persona: Error Handling (Missing Units)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify the API returns a 400 Bad Request when mandatory unit fields are missing.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a POST request without specifying heightUnit.', async () => {
      response = await request.post(`${API_URL}/calculate-bmi`, {
        data: {
          weight: 70,
          weightUnit: 'kilograms',
          height: 1.75
        }
      });
    });

    await test.step('What we expected: We expect the API to reject it with Status 400.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We receive an error message text.', async () => {
      const errorText = await response.text();
      expect(errorText).toContain('weightUnit and heightUnit are required');
    });

    await test.step('Why it got this output: The backend correctly validated the payload and detected the missing height unit.', async () => {});
  });

  test('Persona: Error Handling (Zero Height)', async ({ request }) => {
    await test.step('Why we use this test: We want to make sure the backend prevents division by zero when height is zero.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a POST request with zero height.', async () => {
      response = await request.post(`${API_URL}/calculate-bmi`, {
        data: {
          weight: 70,
          weightUnit: 'kilograms',
          height: 0,
          heightUnit: 'meters'
        }
      });
    });

    await test.step('What we expected: We expect the API to return a 400 error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We receive an error message indicating invalid height.', async () => {
      const errorText = await response.text();
      expect(errorText).toContain('height must be greater than zero');
    });

    await test.step('Why it got this output: The backend calculation logic explicitly prevents a division-by-zero crash by validating the converted height.', async () => {});
  });
});
