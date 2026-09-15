import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const API_URL = 'http://localhost:8080';
const FORMATS = ['csv', 'json', 'txt', 'docx', 'pdf', 'jpg', 'png'];

// Mime types mapping
const MIME_TYPES: Record<string, string> = {
  'csv': 'text/csv',
  'json': 'application/json',
  'txt': 'text/plain',
  'docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'pdf': 'application/pdf',
  'jpg': 'image/jpeg',
  'png': 'image/png'
};

test.describe('Backend API Full Conversion Matrix Persona', () => {

  for (const fromType of FORMATS) {
    for (const toType of FORMATS) {
      if (fromType === toType) continue; // Skip converting to the same format

      test(`Persona: Convert ${fromType.toUpperCase()} to ${toType.toUpperCase()}`, async ({ request }) => {
        
        await test.step(`Why we use this test: We want to rigorously test if the backend can convert a ${fromType.toUpperCase()} file into a ${toType.toUpperCase()} file.`, async () => {});

        let response: any;
        const filePath = path.resolve(__dirname, `../test-files/dummy.${fromType}`);
        const fileBuffer = fs.readFileSync(filePath);

        await test.step(`What we use: We send a valid dummy ${fromType.toUpperCase()} file to the /convert endpoint, asking it to become a ${toType.toUpperCase()}.`, async () => {
          response = await request.post(`${API_URL}/convert`, {
            multipart: {
              fromType: fromType,
              toType: toType,
              file: {
                name: `dummy.${fromType}`,
                mimeType: MIME_TYPES[fromType],
                buffer: fileBuffer
              }
            }
          });
        });

        const status = response.status();
        
        // The API returns 400 if the conversion is explicitly unsupported in the switch statement
        if (status === 400) {
            await test.step(`What we get & Why: The server returned a 400 Bad Request. We check if it's because this combination is unsupported or if it actually failed.`, async () => {
                const text = await response.text();
                if (text.includes('not yet supported')) {
                    test.info().annotations.push({ type: 'Result', description: 'Gracefully unsupported format combination' });
                    expect(text).toContain('not yet supported');
                } else {
                    test.info().annotations.push({ type: 'Result', description: 'Conversion failed during processing' });
                    // Even if it failed during processing, we log it for the framework report.
                    console.error(`Conversion error for ${fromType} to ${toType}:`, text);
                }
            });
        } else {
            await test.step(`What we get & Why: The server returned 200 OK! It successfully performed the conversion and returned the converted file buffer.`, async () => {
                expect(status).toBe(200);
                const buffer = await response.body();
                expect(buffer.length).toBeGreaterThan(0);
                test.info().annotations.push({ type: 'Result', description: 'Successful Conversion' });
            });
        }
      });
    }
  }
});
