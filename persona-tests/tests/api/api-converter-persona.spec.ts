import { test, expect } from '@playwright/test';

// Configuration for API testing. The Swagger API is hosted at localhost:8080 by default.
const API_URL = 'http://localhost:8080';

test.describe('Backend API Converter Persona', () => {

  test('Persona: Successful Conversion (CSV to JSON)', async ({ request }) => {
    await test.step('Why we use this test: We want to make sure the backend API can correctly read a simple CSV file and transform it into JSON format without losing data.', async () => {});
    
    let response;
    
    await test.step('What we use: We are sending a POST request to the /convert endpoint with a fake CSV file containing "id,name" and "1,test".', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'csv',
          toType: 'json',
          file: {
            name: 'test.csv',
            mimeType: 'text/csv',
            buffer: Buffer.from('id,name\n1,test\n2,hello')
          }
        }
      });
    });

    await test.step('What we expected: We expect the backend API to say "OK" (Status 200) because a valid CSV was provided.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: We read the JSON output returned by the backend.', async () => {
      const jsonResponse = await response.json();
      expect(jsonResponse).toHaveLength(2);
      expect(jsonResponse[0].id).toBe('1');
      expect(jsonResponse[0].name).toBe('test');
    });

    await test.step('Why it got this output: The backend successfully parsed the CSV columns and mapped them to JSON properties!', async () => {});
  });

  test('Persona: Un-convertible Complicated File (Malformed JSON to CSV)', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure that if someone sends a broken or complicated file that cannot be read, the server safely rejects it instead of crashing.', async () => {});
    
    let response;
    
    await test.step('What we use: We send a POST request with a broken JSON file "{ this is not valid json ]" and ask the server to convert it to CSV.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'json',
          toType: 'csv',
          file: {
            name: 'broken.json',
            mimeType: 'application/json',
            buffer: Buffer.from('{ this is not valid json ]')
          }
        }
      });
    });

    await test.step('What we expected: We expect the backend to return an Error Status (400 Bad Request) because the file is broken.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: We read the error message returned by the server.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('Conversion failed');
      expect(textResponse).toContain('failed to decode json');
    });

    await test.step('Why it got this output: The backend converter tried to decode the JSON, realized it was malformed, safely caught the error, and sent a polite error message back to us.', async () => {});
  });

  test('Persona: Unsupported Conversion Type from Swagger (CSV to Unknown)', async ({ request }) => {
    await test.step('Why we use this test: We need to check what happens if someone asks for a conversion format that does not exist in our swagger documentation.', async () => {});
    
    let response;
    
    await test.step('What we use: We ask the server to convert a CSV to a fake format called "magic".', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'csv',
          toType: 'magic',
          file: {
            name: 'test.csv',
            mimeType: 'text/csv',
            buffer: Buffer.from('id,name\n1,test')
          }
        }
      });
    });

    await test.step('What we expected: We expect a 400 Bad Request error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server tells us this combination is not supported.', async () => {
      const textResponse = await response.text();
      expect(textResponse).toContain('not yet supported');
    });

    await test.step('Why it got this output: The API handler checked its list of supported formats from the Swagger spec and rejected the unknown "magic" format.', async () => {});
  });

});
