import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8080';

test.describe('Backend API Numeral System Conversion Persona', () => {

  test('Persona: Successful Conversion (Binary to Decimal)', async ({ request }) => {
    await test.step('Why we use this test: We need to ensure that binary strings are correctly parsed to base-10 integers.', async () => {});
    
    let response;
    
    await test.step('What we use: We convert the binary value "1011" to decimal.', async () => {
      response = await request.post(`${API_URL}/convert-numeral`, {
        data: {
          fromBase: 'binary',
          toBase: 'decimal',
          value: '1011'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the result which should be "11".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.result).toBe('11');
    });

    await test.step('Why it got this output: The backend parsed base-2 "1011" into integer 11, and formatted it as base-10 string "11".', async () => {});
  });

  test('Persona: Successful Conversion (Decimal to Hexadecimal)', async ({ request }) => {
    await test.step('Why we use this test: We need to verify that hexadecimal output uses uppercase letters for A-F.', async () => {});
    
    let response;
    
    await test.step('What we use: We convert the decimal value "255" to hexadecimal.', async () => {
      response = await request.post(`${API_URL}/convert-numeral`, {
        data: {
          fromBase: 'decimal',
          toBase: 'hexadecimal',
          value: '255'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the result which should be "FF".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.result).toBe('FF');
    });

    await test.step('Why it got this output: The backend parsed base-10 "255", formatted it as base-16 string "ff", and uppercased it to "FF".', async () => {});
  });

  test('Persona: Successful Conversion (Octal to Binary)', async ({ request }) => {
    await test.step('Why we use this test: We need to ensure conversions between non-decimal bases work seamlessly.', async () => {});
    
    let response;
    
    await test.step('What we use: We convert the octal value "7" to binary.', async () => {
      response = await request.post(`${API_URL}/convert-numeral`, {
        data: {
          fromBase: 'octal',
          toBase: 'binary',
          value: '7'
        }
      });
    });

    await test.step('What we expected: We expect a 200 OK status.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the result which should be "111".', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse.result).toBe('111');
    });

    await test.step('Why it got this output: Octal 7 is correctly mapped to integer 7, and formatted back as binary "111".', async () => {});
  });

  test('Persona: Invalid Input for Base Error', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure that parsing fails safely when the string does not match the base.', async () => {});
    
    let response;
    
    await test.step('What we use: We attempt to parse "9" as a binary number.', async () => {
      response = await request.post(`${API_URL}/convert-numeral`, {
        data: {
          fromBase: 'binary',
          toBase: 'decimal',
          value: '9'
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We read the error text returned by the server.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('invalid value');
    });

    await test.step('Why it got this output: The Go standard library strconv.ParseInt returns an error when encountering a digit out of bounds for the given base.', async () => {});
  });

  test('Persona: Unsupported Base Error', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure unknown bases are rejected.', async () => {});
    
    let response;
    
    await test.step('What we use: We specify a fake base "base32".', async () => {
      response = await request.post(`${API_URL}/convert-numeral`, {
        data: {
          fromBase: 'decimal',
          toBase: 'base32',
          value: '10'
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server specifies the base is unsupported.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('unsupported numeral base');
    });

    await test.step('Why it got this output: The base name is checked against a hardcoded switch statement and falls through to the error condition.', async () => {});
  });

});
