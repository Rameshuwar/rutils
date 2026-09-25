import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

// Base URL of the API
const API_URL = 'http://localhost:8080';

// The actual PDF size endpoint
const PDF_SIZE_ENDPOINT = '/convert-pdf-size';

test.describe('Backend API PDF Size-Work Persona', () => {

  // ---------------------------------------------------------------
  // Helper: load the dummy PDF fixture
  // ---------------------------------------------------------------
  const pdfPath = path.resolve(__dirname, '../../test-files/dummy.pdf');
  const originalBuffer = fs.readFileSync(pdfPath);

  // ---------------------------------------------------------------
  // Persona 1: Successful PDF Compression
  // ---------------------------------------------------------------
  test('Persona: Successful PDF Compression', async ({ request }) => {
    await test.step('Why we use this test: Verify the API can compress a PDF toward a target size.', async () => {});

    let response;
    await test.step('What we use: Send the PDF with conversionType=compression, dataType=KB, targetSize=25.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          conversionType: 'compression',
          dataType: 'KB',
          targetSize: '25',
        },
      });
    });

    await test.step('What we expected: API replies with HTTP 200 and PDF content type.', async () => {
      expect(response.status()).toBe(200);
      expect(response.headers()['content-type']).toContain('application/pdf');
    });

    await test.step('What we get: Verify response headers describe the conversion.', async () => {
      const headers = response.headers();
      expect(headers).toHaveProperty('x-conversion-target-met');
      expect(headers).toHaveProperty('x-conversion-target-size');
      expect(headers).toHaveProperty('x-conversion-actual-size');
    });

    await test.step('Why it got this output: The server ran pdftoppm + JPEG re-encoding and returned a valid PDF.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 2: Successful PDF Expansion
  // ---------------------------------------------------------------
  test('Persona: Successful PDF Expansion', async ({ request }) => {
    await test.step('Why we use this test: Verify the API can expand a PDF to a larger target size.', async () => {});

    let response;
    await test.step('What we use: Send the PDF with conversionType=expand, dataType=KB, targetSize=100.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          conversionType: 'expand',
          dataType: 'KB',
          targetSize: '100',
        },
      });
    });

    await test.step('What we expected: API replies with HTTP 200 and a larger PDF.', async () => {
      expect(response.status()).toBe(200);
      expect(response.headers()['content-type']).toContain('application/pdf');
    });

    await test.step('What we get: The returned PDF should be at least the target size.', async () => {
      const expandedBuffer = await response.body();
      expect(expandedBuffer.length).toBeGreaterThanOrEqual(100 * 1024);
    });

    await test.step('Why it got this output: The server padded the PDF with comment lines before %%EOF.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 3: Invalid conversionType (PDF → Magic)
  // ---------------------------------------------------------------
  test('Persona: Invalid conversionType (Magic)', async ({ request }) => {
    await test.step('Why we use this test: Ensure the API rejects unknown conversionType values.', async () => {});

    let response;
    await test.step('What we use: Send conversionType=magic with valid other fields.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          conversionType: 'magic',
          dataType: 'KB',
          targetSize: '10',
        },
      });
    });

    await test.step('What we expected: The service answers with 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions conversionType must be compression or expand.', async () => {
      const txt = await response.text();
      expect(txt).toContain('conversionType must be compression or expand');
    });

    await test.step('Why it got this output: The API validates conversionType against a whitelist.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 4: Invalid dataType
  // ---------------------------------------------------------------
  test('Persona: Invalid dataType (GB)', async ({ request }) => {
    await test.step('Why we use this test: Ensure the API rejects dataType values other than KB or MB.', async () => {});

    let response;
    await test.step('What we use: Send dataType=GB with valid other fields.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          conversionType: 'compression',
          dataType: 'GB',
          targetSize: '10',
        },
      });
    });

    await test.step('What we expected: The service answers with 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions dataType must be KB or MB.', async () => {
      const txt = await response.text();
      expect(txt).toContain('dataType must be KB or MB');
    });

    await test.step('Why it got this output: The API strictly validates the dataType field.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 5: Compression target below minimum
  // ---------------------------------------------------------------
  test('Persona: Compression Target Below 10 KB', async ({ request }) => {
    await test.step('Why we use this test: Ensure the API rejects compression targets below 10 KB.', async () => {});

    let response;
    await test.step('What we use: Send conversionType=compression with targetSize=5 KB.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          conversionType: 'compression',
          dataType: 'KB',
          targetSize: '5',
        },
      });
    });

    await test.step('What we expected: The service answers with 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the 10 KB minimum.', async () => {
      const txt = await response.text();
      expect(txt).toContain('10 KB');
    });

    await test.step('Why it got this output: The backend enforces a 10 KB compression floor.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 6: Expansion target above maximum
  // ---------------------------------------------------------------
  test('Persona: Expansion Target Above 6000 KB', async ({ request }) => {
    await test.step('Why we use this test: Ensure the API rejects expansion targets above 6000 KB.', async () => {});

    let response;
    await test.step('What we use: Send conversionType=expand with targetSize=7000 KB.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          conversionType: 'expand',
          dataType: 'KB',
          targetSize: '7000',
        },
      });
    });

    await test.step('What we expected: The service answers with 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions the 6000 KB ceiling.', async () => {
      const txt = await response.text();
      expect(txt).toContain('6000');
    });

    await test.step('Why it got this output: The backend enforces a 6000 KB expansion ceiling.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 7: Malformed Input File (Corrupted PDF)
  // ---------------------------------------------------------------
  test('Persona: Malformed Input File (Corrupted PDF)', async ({ request }) => {
    await test.step('Why we use this test: Verify graceful handling of a broken PDF.', async () => {});

    let response;
    await test.step('What we use: Send an intentionally broken PDF binary with valid parameters.', async () => {
      const brokenPdf = Buffer.from('this is not a valid pdf file');
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'broken.pdf',
            mimeType: 'application/pdf',
            buffer: brokenPdf,
          },
          conversionType: 'expand',
          dataType: 'KB',
          targetSize: '100',
        },
      });
    });

    await test.step('What we expected: The API responds with 400 and an informative error.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error text references an invalid PDF.', async () => {
      const txt = await response.text();
      expect(txt).toMatch(/not a valid PDF|failed to decode pdf|Conversion failed/i);
    });

    await test.step('Why it got this output: The PDF magic-byte check failed early and the service returned a friendly error.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 8: Missing conversionType
  // ---------------------------------------------------------------
  test('Persona: Missing Required Field (conversionType)', async ({ request }) => {
    await test.step('Why we use this test: Ensure the API rejects requests missing conversionType.', async () => {});

    let response;
    await test.step('What we use: Send only the file, dataType, and targetSize.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          dataType: 'KB',
          targetSize: '100',
        },
      });
    });

    await test.step('What we expected: The service answers with 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions conversionType is required.', async () => {
      const txt = await response.text();
      expect(txt).toContain('conversionType is required');
    });

    await test.step('Why it got this output: The backend validates required fields up front.', async () => {});
  });

  // ---------------------------------------------------------------
  // Persona 9: Missing dataType
  // ---------------------------------------------------------------
  test('Persona: Missing Required Field (dataType)', async ({ request }) => {
    await test.step('Why we use this test: Ensure the API rejects requests missing dataType.', async () => {});

    let response;
    await test.step('What we use: Send file, conversionType, and targetSize only.', async () => {
      response = await request.post(`${API_URL}${PDF_SIZE_ENDPOINT}`, {
        multipart: {
          file: {
            name: 'dummy.pdf',
            mimeType: 'application/pdf',
            buffer: originalBuffer,
          },
          conversionType: 'compression',
          targetSize: '100',
        },
      });
    });

    await test.step('What we expected: The service answers with 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The error mentions dataType is required.', async () => {
      const txt = await response.text();
      expect(txt).toContain('dataType is required');
    });

    await test.step('Why it got this output: The backend validates required fields up front.', async () => {});
  });

});