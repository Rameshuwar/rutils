import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

// Persona tests for registry ↔ /convert consistency.
//
// This is the highest-value spec in the format suite. It answers the
// question the other specs can't:
//
//     "If /formats says the server supports X→Y, does /convert
//      actually honor it?"
//
// That's the whole point of a registry: one source of truth. If these
// two endpoints disagree, the frontend advertises an option that
// silently fails, which is worse than not showing the option at all.
//
// The test walks the /formats matrix and, for each edge it can find a
// real fixture for, sends the file to /convert and asserts success.
// For edges it can't test end-to-end (no fixture yet), it asserts the
// server at least *claims* the edge.

const API_URL = 'http://localhost:8080';

// Fixtures we actually have on disk. If you add a format, generate
// the fixture and add it to this map — the test will automatically
// start exercising it end-to-end.
const FIXTURE_EXT: Record<string, string> = {
  csv:  'csv',
  json: 'json',
  txt:  'txt',
  docx: 'docx',
  pdf:  'pdf',
  jpg:  'jpg',
  png:  'png',
  webp: 'webp',
  tiff: 'tiff',
  bmp:  'bmp',
};

// MIME types for multipart uploads.
const MIME_TYPES: Record<string, string> = {
  csv:  'text/csv',
  json: 'application/json',
  txt:  'text/plain',
  docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  pdf:  'application/pdf',
  jpg:  'image/jpeg',
  png:  'image/png',
  webp: 'image/webp',
  tiff: 'image/tiff',
  bmp:  'image/bmp',
};

// Edges where the OCR pipeline is required. On a dev machine without
// tesseract installed these will legitimately fail with a dependency
// error, so we treat their failure as "expected environment gap"
// rather than a bug.
const OCR_DEPENDENT_EDGES = new Set([
  'jpg→txt', 'png→txt', 'webp→txt', 'tiff→txt', 'bmp→txt',
]);

function fixturePath(ext: string): string {
  return path.resolve(__dirname, `../test-files/dummy.${FIXTURE_EXT[ext]}`);
}

function hasFixture(ext: string): boolean {
  return ext in FIXTURE_EXT && fs.existsSync(fixturePath(ext));
}

test.describe('Backend API Format Registry — Convert Consistency Persona', () => {

  test('Persona: Every advertised edge is at least acknowledged by /convert', async ({ request }) => {
    await test.step('Why we use this test: The registry and the /convert dispatcher must agree on what is supported. This walks the whole matrix and verifies no advertised edge is silently rejected.', async () => {});

    const formatsResponse = await request.get(`${API_URL}/formats`);
    expect(formatsResponse.status()).toBe(200);
    const { conversions } = await formatsResponse.json();
    expect(Array.isArray(conversions)).toBe(true);
    expect(conversions.length).toBeGreaterThan(0);

    for (const edge of conversions) {
      const { from, to } = edge;

      if (!hasFixture(from)) {
        // No fixture for this source yet — skip, but note it.
        test.info().annotations.push({
          type: 'skip',
          description: `no fixture for ${from}, cannot exercise ${from}→${to}`,
        });
        continue;
      }

      const fileBuffer = fs.readFileSync(fixturePath(from));

      const response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: from,
          toType: to,
          file: {
            name: `dummy.${from}`,
            mimeType: MIME_TYPES[from],
            buffer: fileBuffer,
          },
        },
      });

      const status = response.status();

      if (OCR_DEPENDENT_EDGES.has(`${from}→${to}`)) {
        // On a machine without tesseract, image→txt fails cleanly with
        // a dependency error. Accept either 200 (tesseract installed)
        // or 400 (missing dependency) — never 500.
        expect([200, 400]).toContain(status);
        if (status === 400) {
          const text = await response.text();
          expect(text).toMatch(/tesseract|OCR|dependency/i);
        }
        continue;
      }

      // For all other edges, /convert MUST return 200. A 400 here
      // means the registry is lying — that's the exact bug this test
      // exists to catch.
      expect(
        status,
        `/convert returned ${status} for ${from}→${to} (registry claims this is supported)`,
      ).toBe(200);

      const buffer = await response.body();
      expect(buffer.length, `${from}→${to} returned an empty body`).toBeGreaterThan(0);
    }
  });

  test('Persona: /convert rejects unregistered pairs with the correct message', async ({ request }) => {
    await test.step('Why we use this test: The registry is the sole source of truth. If /convert accepts a pair that the registry does not advertise, that is a bug — the frontend would never show it and users would have no way to discover it.', async () => {});

    // mp3 is not a supported source format at all.
    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        fromType: 'mp3',
        toType: 'pdf',
        file: {
          name: 'dummy.txt',
          mimeType: 'text/plain',
          buffer: Buffer.from('not really an mp3'),
        },
      },
    });

    expect(response.status()).toBe(400);
    const text = await response.text();
    expect(text).toContain('not yet supported');
    expect(text).toContain('mp3');
    expect(text).toContain('pdf');
  });

  test('Persona: /convert rejects a registered source with an unregistered target', async ({ request }) => {
    await test.step('Why we use this test: Even if the source is valid, an unregistered *target* must be rejected. Guards against a copy-paste bug where a formatter accidentally declares support for a format that does not exist.', async () => {});

    const response = await request.post(`${API_URL}/convert`, {
      multipart: {
        fromType: 'jpg',
        toType: 'mp3',
        file: {
          name: 'dummy.jpg',
          mimeType: 'image/jpeg',
          buffer: fs.readFileSync(fixturePath('jpg')),
        },
      },
    });

    expect(response.status()).toBe(400);
    const text = await response.text();
    expect(text).toContain('not yet supported');
  });

  test('Persona: Content-Type of /convert matches what /formats advertises', async ({ request }) => {
    await test.step('Why we use this test: The frontend builds a download link from the response. If /formats says a jpg→png returns image/png but /convert returns application/octet-stream, browsers may still save the file but tooling will mis-classify it.', async () => {});

    const formatsResponse = await request.get(`${API_URL}/formats`);
    const { sources } = await formatsResponse.json();

    // Test a small representative set rather than every edge, to keep
    // the suite fast. The all-edges test above already checks status;
    // this one checks header correctness on a curated sample.
    const sample: Array<[string, string]> = [
      ['jpg',  'png'],
      ['png',  'jpg'],
      ['jpg',  'pdf'],
      ['png',  'pdf'],
      ['webp', 'jpg'],
      ['tiff', 'png'],
      ['bmp',  'jpg'],
      ['pdf',  'png'],
      ['pdf',  'jpg'],
      ['csv',  'json'],
      ['json', 'csv'],
      ['csv',  'pdf'],
      ['txt',  'pdf'],
      ['docx', 'txt'],
    ];

    for (const [from, to] of sample) {
      const buffer = fs.readFileSync(fixturePath(from));

      const response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: from,
          toType: to,
          file: {
            name: `dummy.${from}`,
            mimeType: MIME_TYPES[from],
            buffer,
          },
        },
      });

      expect(response.status(), `${from}→${to} failed`).toBe(200);

      const expectedMIME = sources[to].mime;
      const actualMIME = response.headers()['content-type'] ?? '';

      // The response's Content-Type must start with the registered MIME.
      // We compare with `startsWith` because some MIMEs include a
      // charset suffix (e.g. "text/plain; charset=utf-8").
      expect(
        actualMIME.startsWith(expectedMIME.split(';')[0].trim()),
        `${from}→${to}: expected Content-Type starting with "${expectedMIME}", got "${actualMIME}"`,
      ).toBe(true);
    }
  });

  test('Persona: Content-Disposition filename uses the registered extension', async ({ request }) => {
    await test.step('Why we use this test: When the browser saves the file, it uses the filename from Content-Disposition. If it says "converted.txt" but the payload is a PDF, users get a confusing mislabeled download.', async () => {});

    const formatsResponse = await request.get(`${API_URL}/formats`);
    const { sources } = await formatsResponse.json();

    const cases: Array<[string, string]> = [
      ['jpg',  'png'],  // → converted.png
      ['jpg',  'pdf'],  // → converted.pdf
      ['csv',  'json'], // → converted.json
      ['txt',  'docx'], // → converted.docx
    ];

    for (const [from, to] of cases) {
      const buffer = fs.readFileSync(fixturePath(from));

      const response = await request.post(`${API_URL}/convert`, {
        multipart: {
          fromType: from,
          toType: to,
          file: {
            name: `dummy.${from}`,
            mimeType: MIME_TYPES[from],
            buffer,
          },
        },
      });

      expect(response.status()).toBe(200);

      const disposition = response.headers()['content-disposition'] ?? '';
      const expectedExt = sources[to].ext;

      expect(
        disposition,
        `${from}→${to}: Content-Disposition missing expected extension ${expectedExt}`,
      ).toContain(expectedExt);
    }
  });

  test('Persona: Tier-1 has no self-edges or duplicate edges', async ({ request }) => {
    await test.step('Why we use this test: A self-edge (jpg→jpg) is nonsensical; a duplicate edge means two formatters registered the same pair, which should be impossible given the registry panics on duplicates — but this is a belt-and-braces check.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const { conversions } = await response.json();

    const seen = new Set<string>();
    for (const edge of conversions) {
      const key = `${edge.from}→${edge.to}`;

      expect(edge.from).not.toBe(edge.to);

      expect(
        seen.has(key),
        `duplicate edge: ${key}`,
      ).toBe(false);

      seen.add(key);
    }
  });
});