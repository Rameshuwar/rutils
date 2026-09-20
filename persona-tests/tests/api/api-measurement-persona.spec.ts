import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API Measurement Persona', () => {

  test('Persona: Successful Conversion (Length: meters to feet)', async ({ request }) => {
    await test.step('Why we use this test: We want to make sure the backend API can correctly convert length units, such as meters to feet.', async () => {});
    
    let response;
    
    await test.step('What we use: We are sending a POST request to the /convert-measurement endpoint with a JSON body specifying length, meters, feet, and a value of 1.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'length',
          fromUnit: 'meters',
          toUnit: 'feet',
          value: 1
        }
      });
    });

    await test.step('What we expected: We expect the backend API to say "OK" (Status 200).', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the JSON output returned by the backend.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse).toHaveProperty('result');
      // 1 meter is approximately 3.28084 feet
      expect(jsonResponse.result).toBeCloseTo(3.28084, 4);
    });

    await test.step('Why it got this output: The backend successfully found the conversion factors for length and correctly computed the result!', async () => {});
  });

  test('Persona: Successful Conversion (Weight: kilograms to pounds)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify that weight conversions work accurately.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a POST request to convert 10 kilograms to pounds.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'weight',
          fromUnit: 'kilograms',
          toUnit: 'pounds',
          value: 10
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We check the converted weight value.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse).toHaveProperty('result');
      // 10 kg is approx 22.0462 pounds
      expect(jsonResponse.result).toBeCloseTo(22.0462, 3);
    });

    await test.step('Why it got this output: The backend accurately applied the weight conversion formula.', async () => {});
  });

  test('Persona: Successful Conversion (Volume: liters to gallons)', async ({ request }) => {
    await test.step('Why we use this test: We want to verify volume conversions function properly.', async () => {});
    
    let response;
    
    await test.step('What we use: We ask the server to convert 5 liters to gallons.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'volume',
          fromUnit: 'liters',
          toUnit: 'gallons',
          value: 5
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We check the converted volume value.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse).toHaveProperty('result');
      // 5 liters / 3.78541 = ~1.32086 gallons
      expect(jsonResponse.result).toBeCloseTo(1.32086, 4);
    });

    await test.step('Why it got this output: The API correctly routed the volume category and computed the units.', async () => {});
  });

  test('Persona: Successful Conversion (Area: acres to square_meters)', async ({ request }) => {
    await test.step('Why we use this test: We want to check area conversions handling large multipliers.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a POST request to convert 2 acres to square meters.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'area',
          fromUnit: 'acres',
          toUnit: 'square_meters',
          value: 2
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the returned JSON.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse).toHaveProperty('result');
      // 2 acres = 2 * 4046.8564224 = 8093.7128448 sq meters
      expect(jsonResponse.result).toBeCloseTo(8093.7128, 3);
    });

    await test.step('Why it got this output: The server properly calculated the area conversion using its predefined factors.', async () => {});
  });

  test('Persona: Successful Conversion (Time: hours to seconds)', async ({ request }) => {
    await test.step('Why we use this test: We need to ensure time-based conversions are reliable.', async () => {});
    
    let response;
    
    await test.step('What we use: We request the conversion of 2 hours into seconds.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'time',
          fromUnit: 'hours',
          toUnit: 'seconds',
          value: 2
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the result in seconds.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.result).toBe(7200); // 2 hours = 7200 seconds
    });

    await test.step('Why it got this output: Time factors are strictly defined and correctly evaluated by the backend.', async () => {});
  });

  test('Persona: Successful Conversion (Speed: kilometers_per_hour to meters_per_second)', async ({ request }) => {
    await test.step('Why we use this test: We need to ensure the newly added speed category performs conversions accurately.', async () => {});
    
    let response;
    
    await test.step('What we use: We request the conversion of 100 kilometers_per_hour into meters_per_second.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'speed',
          fromUnit: 'kilometers_per_hour',
          toUnit: 'meters_per_second',
          value: 100
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the result in meters per second.', async () => {
      const jsonResponse = await response.json();
      // 100 km/h = ~27.7778 m/s
      expect(jsonResponse.result).toBeCloseTo(27.7778, 3);
    });

    await test.step('Why it got this output: The speed factors (including 1000/3600 for km/h) were correctly evaluated by the backend.', async () => {});
  });

  test('Persona: Successful Conversion (Data: megabytes to kilobytes)', async ({ request }) => {
    await test.step('Why we use this test: We need to ensure data conversions correctly handle base-2 (1024) multipliers.', async () => {});
    
    let response;
    
    await test.step('What we use: We request the conversion of 5 megabytes into kilobytes.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'data',
          fromUnit: 'megabytes',
          toUnit: 'kilobytes',
          value: 5
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the result in kilobytes.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.result).toBe(5120); // 5 MB = 5 * 1024 = 5120 KB
    });

    await test.step('Why it got this output: The backend uses 1024-based multipliers for data units, successfully computing 5 * 1024.', async () => {});
  });

  test('Persona: Special Category Conversion (Temperature: celsius to fahrenheit)', async ({ request }) => {
    await test.step('Why we use this test: Temperature conversions use complex offset logic instead of simple multiplication. We need to verify this special case works.', async () => {});
    
    let response;
    
    await test.step('What we use: We request the server to convert 100 Celsius to Fahrenheit.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'temperature',
          fromUnit: 'celsius',
          toUnit: 'fahrenheit',
          value: 100
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the resulting temperature.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.result).toBe(212); // 100 C = 212 F
    });

    await test.step('Why it got this output: The backend safely branched into the special temperature logic and applied the formula (C * 9/5) + 32.', async () => {});
  });

  test('Persona: Invalid Category Error', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure the API safely handles unknown measurement categories without crashing.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a request with a fake category called "magic_power".', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'magic_power',
          fromUnit: 'mana',
          toUnit: 'energy',
          value: 50
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We read the error text returned by the server.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('unsupported category');
    });

    await test.step('Why it got this output: The backend realized the category does not exist in its lookup map and gracefully rejected the request.', async () => {});
  });

  test('Persona: Unsupported Unit Error', async ({ request }) => {
    await test.step('Why we use this test: We need to see what happens when a valid category is provided, but an invalid unit is given.', async () => {});
    
    let response;
    
    await test.step('What we use: We request length conversion from "meters" to a non-existent unit "lightyears".', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {
          category: 'length',
          fromUnit: 'meters',
          toUnit: 'lightyears',
          value: 1
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server specifies that the unit is unsupported.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('unsupported toUnit');
    });

    await test.step('Why it got this output: The length category exists, but the "lightyears" unit is not in its conversion dictionary.', async () => {});
  });

  test('Persona: Missing Fields Error', async ({ request }) => {
    await test.step('Why we use this test: We need to verify that incomplete requests are caught early.', async () => {});
    
    let response;
    
    await test.step('What we use: We send an empty JSON body.', async () => {
      response = await request.post(`${API_URL}/convert-measurement`, {
        data: {}
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We check the error message.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('are required');
    });

    await test.step('Why it got this output: The API validator caught the missing "category", "fromUnit", and "toUnit" fields before processing.', async () => {});
  });

});
