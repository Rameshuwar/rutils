import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

// Persona tests for negative paths and edge cases in the format suite.
//
// The discovery and consistency specs cover the "happy path". This file
// covers everything that should fail — and verifies it fails in the
// right way:
//
//   - 400 with a specific message, not 500
//   - no accidental success on malformed input
//   - no crashes on malicious input
//   - correct handling of empty/missing fields

const API_URL = 'http://localhost:8080';

function fixturePath(ext: string): string {
  return path.resolve(__dirname, `../test-files/dummy.${ext}`);
}

test.describe('Backend API Format Registry — Negative Persona', () => {

  // ================================================================
  // /formats discovery endpoint — negative cases
  // ================================================================

  test('Persona: GET /formats/unknown returns 404, not 200', async ({ request }) => {
    await test.step('Why we use this test: Clients should not receive an empty 200 for an unknown format, because that masks typos. A 404 is the correct signal.', async () => {});

    const response = await request.get(`${API_URL}/formats/nonexistent-format-xyz`);

    expect(response.status()).toBe(404);
    expect(await response.text()).toContain('unknown format');
  });

  test('Persona: GET /formats/ with no type returns 400, not 404 or 500', async ({ request }) => {
    await test.step('Why we use this test: The path is malformed — the required segment is missing. 400 signals "your request is bad" and 404 signals "the resource doesn\'t exist". The former is correct here.', async () => {});

    const response = await request.get(`${API_URL}/formats/`);

    expect(response.status()).toBe(400);
    expect(await response.text()).toContain('format type is required');
  });

  test('Persona: DELETE /formats returns 405', async ({ request }) => {
    await test.step('Why we use this test: The discovery endpoint is read-only. Any method other than GET must be rejected with the correct status.', async () => {});

    const response = await request.delete(`${API_URL}/formats`);
    expect(response.status()).toBe(405);
  });

  test('Persona: PUT /formats returns 405', async ({ request }) => {
    await test.step('Why we use this test: Same as DELETE — the endpoint has no writable surface.', async () => {});

    const response = await request.put(`${API_URL}/formats`, { data: {} });
    expect(response.status()).toBe(405);
  });

  // ================================================================
  // /convert — malformed requests
  // ================================================================

  test('Persona: /convert rejects missing fromType', async ({ request }) => {
    await test.step('Why we use this test: A request without a source format cannot be served. It must fail cleanly with 400.', async () => {});

    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        toType: 'pdf',
        file: {
          name: 'dummy.txt',
          mimeType: 'text/plain',
          buffer: fs.readFileSync(fixturePath('txt')),
        },
      },
    });

    expect(response.status()).toBe(400);
    expect(await response.text()).toContain('Missing fromType or toType');
  });

  test('Persona: /convert rejects missing toType', async ({ request }) => {
    await test.step('Why we use this test: Same as above but the other direction.', async () => {});

    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        fromType: 'txt',
        file: {
          name: 'dummy.txt',
          mimeType: 'text/plain',
          buffer: fs.readFileSync(fixturePath('txt')),
        },
      },
    });

    expect(response.status()).toBe(400);
    expect(await response.text()).toContain('Missing fromType or toType');
  });

  test('Persona: /convert rejects missing file field', async ({ request }) => {
    await test.step('Why we use this test: A request with valid types but no file cannot be converted. The error must name the missing field so clients can fix their request.', async () => {});

    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        fromType: 'txt',
        toType: 'pdf',
        // no `file` key
      },
    });

    expect(response.status()).toBe(400);
    const text = await response.text();
    expect(text).toMatch(/file/i);
  });

  test('Persona: /convert rejects a GET request', async ({ request }) => {
    await test.step('Why we use this test: The endpoint requires a multipart body, so GET is invalid. Returns 405 with the exact message the old handler used, to preserve client compatibility.', async () => {});

    const response = await request.get(`${API_URL}/convert`);
    expect(response.status()).toBe(405);
    expect(await response.text()).toContain('Only POST method is allowed');
  });

  // ================================================================
  // /convert — malformed content
  // ================================================================

  test('Persona: /convert with text named .jpg fails cleanly', async ({ request }) => {
    await test.step('Why we use this test: Attackers and confused clients may send a text file with an image extension. The decoder must reject it with a 400 — never a 500 or a partial write.', async () => {});

    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        fromType: 'jpg',
        toType: 'png',
        file: {
          name: 'fake.jpg',
          mimeType: 'image/jpeg',
          buffer: Buffer.from('this is definitely not a real jpeg file'),
        },
      },
    });

    expect(response.status()).toBe(400);
    const text = await response.text();
    expect(text).toMatch(/invalid|failed|not a JPEG|decode/i);
  });

  test('Persona: /convert with an empty file fails cleanly', async ({ request }) => {
    await test.step('Why we use this test: An empty upload should be treated as invalid input, not silently converted to an empty output.', async () => {});

    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        fromType: 'jpg',
        toType: 'png',
        file: {
          name: 'empty.jpg',
          mimeType: 'image/jpeg',
          buffer: Buffer.alloc(0),
        },
      },
    });

    expect(response.status()).toBe(400);
  });

  test('Persona: /convert with mismatched magic bytes fails cleanly', async ({ request }) => {
    await test.step('Why we use this test: Sending a PNG file with fromType=jpg must fail — the decoder for jpg cannot read PNG bytes. This guards against the server trusting the client\'s declared type over the actual bytes.', async () => {});

    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        fromType: 'jpg',
        toType: 'png',
        file: {
          name: 'sneaky.png',
          mimeType: 'image/jpeg',
          buffer: fs.readFileSync(fixturePath('png')),
        },
      },
    });

    expect(response.status()).toBe(400);
  });

  // ================================================================
  // /formats — content sanity
  // ================================================================

  test('Persona: /formats sources map is not empty', async ({ request }) => {
    await test.step('Why we use this test: A registry misconfiguration (all formatters failed to init) would produce an empty sources map. Catch that early.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    expect(Object.keys(body.sources).length).toBeGreaterThan(0);
    expect(body.conversions.length).toBeGreaterThan(0);
  });

  test('Persona: /formats/{type} response for each source is well-formed', async ({ request }) => {
    await test.step('Why we use this test: Every source the matrix advertises must have a working detail endpoint. If one 404s, the frontend can\'t display it.', async () => {});

    const listResponse = await request.get(`${API_URL}/formats`);
    const { sources } = await listResponse.json();

    for (const fmt of Object.keys(sources)) {
      const detailResponse = await request.get(`${API_URL}/formats/${fmt}`);

      expect(
        detailResponse.status(),
        `/formats/${fmt} returned ${detailResponse.status()}`,
      ).toBe(200);

      const detail = await detailResponse.json();
      expect(detail.type, `/formats/${fmt} has wrong type field`).toBe(fmt);
      expect(detail.mime).not.toBe('');
      expect(detail.ext).not.toBe('');
      expect(detail.category).not.toBe('');
    }
  });

  test('Persona: /formats does not leak filesystem paths', async ({ request }) => {
    await test.step('Why we use this test: The response should describe formats, not reveal server internals. If a path like /home/ram/... shows up, that is an information leak.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const text = await response.text();

    expect(text).not.toContain('/home/');
    expect(text).not.toContain('/usr/');
    expect(text).not.toContain('internal/formatters');
  });
});