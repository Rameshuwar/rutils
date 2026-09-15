import { test, expect } from '@playwright/test';

const API_URL = 'http://localhost:8080';

test.describe('Backend API Hard/Un-convertible Files Persona', () => {

  test('Persona: Huge File Size Limit (11MB File)', async ({ request }) => {
    await test.step('Why we use this test: We want to see how the system handles a file that is too large (over 10MB limit) to protect the server from memory exhaustion.', async () => {});

    let response: any;

    await test.step('What we use: We dynamically generate an 11MB text file full of zeros and send it for conversion to JSON.', async () => {
      // Create an 11 MB buffer
      const hugeBuffer = Buffer.alloc(11 * 1024 * 1024, '0');
      
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'txt',
          toType: 'json',
          file: {
            name: 'huge.txt',
            mimeType: 'text/plain',
            buffer: hugeBuffer
          }
        }
      });
    });

    await test.step('What we expected: The server should reject the request (return an error or 400 Bad Request) because it exceeds the 10MB limit.', async () => {
      expect(response.status()).not.toBe(200);
    });

    await test.step('What we get: We read the error message.', async () => {
      const text = await response.text();
      // The Go server might just truncate or throw a multipart parse error depending on its strictness.
      // We expect some form of failure.
      test.info().annotations.push({ type: 'Result', description: `Server response: ${text.substring(0, 100)}` });
    });

    await test.step('Why it got this output: The Go server uses r.ParseMultipartForm(10 << 20), which limits parsed memory. Handling large files gracefully is critical for stability.', async () => {});
  });

  test('Persona: Corrupted / Fake DOCX File', async ({ request }) => {
    await test.step('Why we use this test: We want to see what happens when a user uploads a file that is completely fake or corrupted (e.g. a plain text file renamed to .docx).', async () => {});

    let response: any;

    await test.step('What we use: We send a file that claims to be a DOCX, but its contents are just "I am not a real zip file!". We try to convert it to CSV.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'docx',
          toType: 'csv',
          file: {
            name: 'fake.docx',
            mimeType: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
            buffer: Buffer.from('I am not a real zip file!')
          }
        }
      });
    });

    await test.step('What we expected: We expect the conversion to fail and return a 400 Bad Request because a DOCX must be a valid ZIP archive containing XMLs.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server safely intercepts the error.', async () => {
      const text = await response.text();
      expect(text).toContain('failed');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });

    await test.step('Why it got this output: The backend zip reader (archive/zip) noticed the file lacks a valid zip signature, stopped processing immediately, and safely rejected the request without crashing!', async () => {});
  });

  test('Persona: Corrupted / Fake PDF File', async ({ request }) => {
    await test.step('Why we use this test: We want to test how the PDF parser handles a totally invalid PDF.', async () => {});

    let response: any;

    await test.step('What we use: We send a file that claims to be a PDF, but its contents are just junk data. We try to convert it to TXT.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'pdf',
          toType: 'txt',
          file: {
            name: 'fake.pdf',
            mimeType: 'application/pdf',
            buffer: Buffer.from('this is completely junk data and definitely not a PDF')
          }
        }
      });
    });

    await test.step('What we expected: We expect the conversion to fail and return a 400 Bad Request.', async () => {
      expect(response.status()).toBe(400);
    });

    await test.step('What we get: The server safely reports the parsing failure.', async () => {
      const text = await response.text();
      expect(text).toContain('failed');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });

    await test.step('Why it got this output: The pdf library attempted to read the PDF trailer and format, failed, and safely returned the error to the API handler.', async () => {});
  });

});
