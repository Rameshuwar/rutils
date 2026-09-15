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

  test('Persona: Empty File (0 Bytes)', async ({ request }) => {
    await test.step('Why we use this test: We want to ensure the API safely rejects files that have absolutely no data inside them.', async () => {});
    let response: any;
    await test.step('What we use: We send a completely empty 0-byte CSV file to convert to JSON.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'csv',
          toType: 'json',
          file: { name: 'empty.csv', mimeType: 'text/csv', buffer: Buffer.from('') }
        }
      });
    });
    await test.step('What we expected: A 400 Bad Request error.', async () => { expect(response.status()).toBe(400); });
    await test.step('What we get: The server identifies that the CSV has no data.', async () => {
      const text = await response.text();
      expect(text).toContain('failed to read csv or csv is empty');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });
    await test.step('Why it got this output: The Go CSV parser expects at least a header row. Finding nothing, it safely errors out instead of generating an empty JSON file.', async () => {});
  });

  test('Persona: Broken Structure CSV (Varying Columns)', async ({ request }) => {
    await test.step('Why we use this test: We want to see how the parser handles a CSV file that is severely broken (rows have different number of columns).', async () => {});
    let response: any;
    await test.step('What we use: We send a CSV where row 1 has 2 columns, but row 2 has 5 columns.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'csv',
          toType: 'json',
          file: { name: 'broken.csv', mimeType: 'text/csv', buffer: Buffer.from('id,name\n1,test,extra,data,here') }
        }
      });
    });
    await test.step('What we expected: A 400 Bad Request error due to parsing failure.', async () => { expect(response.status()).toBe(400); });
    await test.step('What we get: The server rejects the mismatched columns.', async () => {
      const text = await response.text();
      expect(text).toContain('wrong number of fields');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });
    await test.step('Why it got this output: The Go standard library CSV parser strictly enforces that all rows match the header length. It safely catches the bad row and aborts.', async () => {});
  });

  test('Persona: Meaningless Data (Empty JSON Array to CSV)', async ({ request }) => {
    await test.step('Why we use this test: A valid JSON format might still contain meaningless data for conversion, such as an empty array.', async () => {});
    let response: any;
    await test.step('What we use: We send a completely empty JSON array "[]" and ask for a CSV.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'json',
          toType: 'csv',
          file: { name: 'empty.json', mimeType: 'application/json', buffer: Buffer.from('[]') }
        }
      });
    });
    await test.step('What we expected: A 400 Bad Request error.', async () => { expect(response.status()).toBe(400); });
    await test.step('What we get: The server reports empty JSON data.', async () => {
      const text = await response.text();
      expect(text).toContain('empty json data');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });
    await test.step('Why it got this output: The JSON-to-CSV converter requires at least one object to determine the CSV headers. It realizes it cannot build headers and aborts.', async () => {});
  });

  test('Persona: Fake Image File (Text disguised as JPG)', async ({ request }) => {
    await test.step('Why we use this test: We need to ensure that the image processing library does not crash when fed non-image data with an image extension.', async () => {});
    let response: any;
    await test.step('What we use: We send a file named "fake.jpg" containing plain text, requesting conversion to PNG.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'jpg',
          toType: 'png',
          file: { name: 'fake.jpg', mimeType: 'image/jpeg', buffer: Buffer.from('I am just plain text pretending to be a JPG image!') }
        }
      });
    });
    await test.step('What we expected: A 400 Bad Request error.', async () => { expect(response.status()).toBe(400); });
    await test.step('What we get: The server successfully catches the invalid format.', async () => {
      const text = await response.text();
      expect(text).toContain('failed to decode jpeg');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });
    await test.step('Why it got this output: The JPEG decoder reads the file headers, realizes the magic bytes for JPEG are missing, and throws an error rather than panicking.', async () => {});
  });

  test('Persona: Malformed API Request (Missing toType)', async ({ request }) => {
    await test.step('Why we use this test: We want to see how the API handles a user who forgets to provide all the required form fields.', async () => {});
    let response: any;
    await test.step('What we use: We send a valid file and a "fromType", but we purposefully omit the "toType" field.', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'csv',
          file: { name: 'dummy.csv', mimeType: 'text/csv', buffer: Buffer.from('id,name\n1,test') }
        }
      });
    });
    await test.step('What we expected: A 400 Bad Request error.', async () => { expect(response.status()).toBe(400); });
    await test.step('What we get: The server explicitly asks for the missing field.', async () => {
      const text = await response.text();
      expect(text).toContain('Missing fromType or toType');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });
    await test.step('Why it got this output: The API handler validates the form fields at the very beginning of the request and immediately aborts if required parameters are missing.', async () => {});
  });

  test('Persona: Malformed API Request (Wrong File Key Name)', async ({ request }) => {
    await test.step('Why we use this test: Sometimes integrations send files under the wrong form-data key. We need to ensure the server handles this gracefully.', async () => {});
    let response: any;
    await test.step('What we use: We send the file under the form key "document" instead of the expected key "file".', async () => {
      response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: 'csv',
          toType: 'json',
          document: { name: 'dummy.csv', mimeType: 'text/csv', buffer: Buffer.from('id,name\n1,test') }
        }
      });
    });
    await test.step('What we expected: A 400 Bad Request error.', async () => { expect(response.status()).toBe(400); });
    await test.step('What we get: The server reports it cannot find the file.', async () => {
      const text = await response.text();
      expect(text).toContain('Failed to get file from request');
      test.info().annotations.push({ type: 'Result', description: `Error Caught: ${text.trim()}` });
    });
    await test.step('Why it got this output: The API looks specifically for the key "file" as defined in the Swagger doc. When it does not find it, it returns an explicit error.', async () => {});
  });

});
