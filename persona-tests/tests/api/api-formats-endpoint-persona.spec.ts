import { test, expect } from '@playwright/test';

// Persona tests for the file-format discovery API.
//
// These tests exercise the two read-only endpoints that back the
// registry:
//
//   GET /formats           → full conversion matrix
//   GET /formats/{type}    → per-format detail
//
// They never touch /convert — that's the consistency spec's job.

const API_URL = 'http://localhost:8080';

// The canonical set of formats the backend is expected to advertise.
// If this list changes, update it in exactly one place: here.
const EXPECTED_SOURCES = [
  'bmp', 'csv', 'docx', 'jpg', 'json', 'pdf', 'png', 'tiff', 'txt', 'webp',
];

// Authoritative category per format — matches internal/formatters/registry.go's
// `formatMeta` table. Any drift here means the backend's table changed.
const EXPECTED_CATEGORY: Record<string, string> = {
  bmp:  'image',
  csv:  'data',
  docx: 'document',
  jpg:  'image',
  json: 'data',
  pdf:  'document',
  png:  'image',
  tiff: 'image',
  txt:  'document',
  webp: 'image',
};

test.describe('Backend API Format Registry — Discovery Persona', () => {

  test('Persona: GET /formats returns the full registry', async ({ request }) => {
    await test.step('Why we use this test: The frontend calls this endpoint to build its dropdowns at runtime. It must return every source format and every conversion edge in one response.', async () => {});

    let response;

    await test.step('What we use: A plain GET request to /formats.', async () => {
      response = await request.get(`${API_URL}/formats`);
    });

    await test.step('What we expected: 200 OK with a JSON body containing a sources map and a conversions array.', async () => {
      expect(response.status()).toBe(200);
      expect(response.headers()['content-type']).toContain('application/json');
    });

    await test.step('What we get: The response has the right shape and every expected source format is present.', async () => {
      const body = await response.json();
      expect(body).toHaveProperty('sources');
      expect(body).toHaveProperty('conversions');
      expect(typeof body.sources).toBe('object');
      expect(Array.isArray(body.conversions)).toBe(true);

      for (const fmt of EXPECTED_SOURCES) {
        expect(body.sources).toHaveProperty(fmt);
      }
    });

    await test.step('Why it got this output: The handler walks the registry and emits one entry per registered format.', async () => {});
  });

  test('Persona: Every source has complete metadata', async ({ request }) => {
    await test.step('Why we use this test: The frontend renders "JPEG (.jpg)" in dropdowns. If any field is empty, the UI shows "undefined (.undefined)". This test makes that impossible.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    expect(response.status()).toBe(200);

    const body = await response.json();

    for (const fmt of EXPECTED_SOURCES) {
      const info = body.sources[fmt];
      expect(info, `missing metadata for ${fmt}`).toBeTruthy();
      expect(info.type,     `${fmt}.type is empty`).toBe(fmt);
      expect(info.mime,     `${fmt}.mime is empty`).not.toBe('');
      expect(info.ext,      `${fmt}.ext is empty`).not.toBe('');
      expect(info.category, `${fmt}.category is empty`).not.toBe('');
    }
  });

  test('Persona: Categories are correct per format', async ({ request }) => {
    await test.step('Why we use this test: Category drives UI grouping. A TXT file showing up under "image" or a JPG under "data" would confuse every user. This test pins the correct value for every format.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    for (const [fmt, expectedCategory] of Object.entries(EXPECTED_CATEGORY)) {
      expect(
        body.sources[fmt]?.category,
        `expected ${fmt} to have category "${expectedCategory}"`,
      ).toBe(expectedCategory);
    }
  });

  test('Persona: Every conversion edge has non-empty from/to/category', async ({ request }) => {
    await test.step('Why we use this test: A conversion entry with an empty "from" or "to" would silently fail downstream. Guard against it here rather than at /convert time.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    expect(body.conversions.length).toBeGreaterThan(0);

    for (const edge of body.conversions) {
      expect(edge.from,     `edge with empty from: ${JSON.stringify(edge)}`).not.toBe('');
      expect(edge.to,       `edge with empty to: ${JSON.stringify(edge)}`).not.toBe('');
      expect(edge.category, `edge with empty category: ${JSON.stringify(edge)}`).not.toBe('');

      // No self-edges.
      expect(edge.from).not.toBe(edge.to);
    }
  });

  test('Persona: Total registered edges is exactly 45', async ({ request }) => {
    await test.step('Why we use this test: This is a regression guard. The Tier-1 release registers exactly 45 edges. If someone accidentally deletes a formatter, this count drops and we know immediately.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    expect(body.conversions.length).toBe(45);
  });

  // ---------------------------------------------------------------
  // /formats/{type} — per-format detail
  // ---------------------------------------------------------------

  test('Persona: GET /formats/jpg returns the image detail', async ({ request }) => {
    await test.step('Why we use this test: The most common image source. Verify the response advertises the exact reachable targets.', async () => {});

    const response = await request.get(`${API_URL}/formats/jpg`);
    expect(response.status()).toBe(200);

    const detail = await response.json();
    expect(detail.type).toBe('jpg');
    expect(detail.mime).toBe('image/jpeg');
    expect(detail.ext).toBe('.jpg');
    expect(detail.category).toBe('image');

    // jpg can be converted to pdf, png, txt (registered via image_jpg.go).
    expect(detail.canConvertTo.sort()).toEqual(['pdf', 'png', 'txt'].sort());
    // Only png, tiff, webp, bmp can be converted to jpg (all images).
    expect(detail.canConvertFrom).toContain('png');
    expect(detail.canConvertFrom).toContain('tiff');
    expect(detail.canConvertFrom).toContain('webp');
    expect(detail.canConvertFrom).toContain('bmp');
  });

  test('Persona: GET /formats/txt returns the document detail', async ({ request }) => {
    await test.step('Why we use this test: TXT is a document target for every image format (OCR) but a source for fewer paths. Verify the asymmetry is reported honestly.', async () => {});

    const response = await request.get(`${API_URL}/formats/txt`);
    expect(response.status()).toBe(200);

    const detail = await response.json();
    expect(detail.type).toBe('txt');
    expect(detail.category).toBe('document');
    expect(detail.mime).toContain('text/plain');

    // docx → txt, pdf → txt, and OCR (jpg/png/webp/tiff/bmp) → txt all target it.
    expect(detail.canConvertFrom).toContain('docx');
    expect(detail.canConvertFrom).toContain('pdf');
    expect(detail.canConvertFrom).toContain('jpg');
  });

  test('Persona: GET /formats/csv returns the data detail', async ({ request }) => {
    await test.step('Why we use this test: CSV is a "data" category format. Verify its metadata is right and that it advertises the expected targets.', async () => {});

    const response = await request.get(`${API_URL}/formats/csv`);
    expect(response.status()).toBe(200);

    const detail = await response.json();
    expect(detail.category).toBe('data');
    expect(detail.mime).toBe('text/csv');

    // csv → docx, json, pdf, txt are registered.
    expect(detail.canConvertTo).toEqual(
      expect.arrayContaining(['docx', 'json', 'pdf', 'txt']),
    );
  });

  test('Persona: GET /formats/WEBP is case-insensitive', async ({ request }) => {
    await test.step('Why we use this test: Clients may send uppercase or mixed-case type names. The handler must normalize them before lookup.', async () => {});

    const lower = await request.get(`${API_URL}/formats/webp`);
    const upper = await request.get(`${API_URL}/formats/WEBP`);

    expect(lower.status()).toBe(200);
    expect(upper.status()).toBe(200);

    const lowerBody = await lower.json();
    const upperBody = await upper.json();

    // The body's `type` field should be the normalized lower-case form,
    // and everything else should be identical.
    expect(lowerBody).toEqual(upperBody);
    expect(upperBody.type).toBe('webp');
  });

  test('Persona: GET /formats/unknownformat returns 404', async ({ request }) => {
    await test.step('Why we use this test: An unknown format name must return 404, not 500 or an empty 200. This lets the frontend distinguish "typo" from "server broken".', async () => {});

    const response = await request.get(`${API_URL}/formats/mp3`);

    expect(response.status()).toBe(404);
    const text = await response.text();
    expect(text).toContain('unknown format');
  });

  test('Persona: GET /formats/ (trailing slash, no type) returns 400', async ({ request }) => {
    await test.step('Why we use this test: A request missing the required path segment should fail cleanly, not panic or return 200.', async () => {});

    const response = await request.get(`${API_URL}/formats/`);

    expect(response.status()).toBe(400);
    const text = await response.text();
    expect(text).toContain('format type is required');
  });

  test('Persona: POST /formats returns 405 Method Not Allowed', async ({ request }) => {
    await test.step('Why we use this test: The endpoint is read-only. Write methods must be rejected with the correct status.', async () => {});

    const response = await request.post(`${API_URL}/formats`, { data: {} });

    expect(response.status()).toBe(405);
    const text = await response.text();
    expect(text).toContain('Method not allowed');
  });

  test('Persona: Path traversal attempt is safely rejected', async ({ request }) => {
    await test.step('Why we use this test: A malicious client might try to escape the format namespace. Verify the handler does not crash and returns a clean error.', async () => {});

    const response = await request.get(`${API_URL}/formats/..%2F..%2Fetc%2Fpasswd`);

    // Should be a 404 (unknown format) or 400 (bad request), never 500.
    expect([400, 404]).toContain(response.status());
  });

  test('Persona: Every advertised target is also a valid source', async ({ request }) => {
    await test.step('Why we use this test: Internal consistency check — every format that appears as a target must also appear in the sources map. Otherwise the frontend cannot let users pick that format as an input.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    const sources = new Set(Object.keys(body.sources));

    for (const edge of body.conversions) {
      expect(
        sources.has(edge.to),
        `target "${edge.to}" (from ${edge.from}→${edge.to}) is not present in sources map`,
      ).toBe(true);
    }
  });
});