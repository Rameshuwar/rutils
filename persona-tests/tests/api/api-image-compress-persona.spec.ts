import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const API_URL = 'http://localhost:8080';

// Fixture: reuse the repo's dummy image files. These are already
// used by the /convert matrix spec, so they exist in the tree.
const FIXTURE_DIR = path.resolve(__dirname, '../test-files');

test.describe('Backend API Image Compress Persona', () => {

  // ================================================================
  // GROUP 1: Happy paths
  // ================================================================

  test('Persona: Compress a JPEG by 50% (percentage target)', async ({ request }) => {
    await test.step('Why we use this test: the most common request — halve a JPEG in size via the quality ladder.', async () => {});

    const original = fs.readFileSync(path.join(FIXTURE_DIR, 'dummy.jpg'));
    let response;

    await test.step('What we use: POST /compress-image with conversionType=compress, dataType=percentage, targetValue=50.', async () => {
      response = await request.post(`${API_URL}/compress-image`, {
        multipart: {
          file: {
            name: 'dummy.jpg',
            mimeType: 'image/jpeg',
            buffer: original,
          },
          conversionType: 'compress',
          dataType: 'percentage',
          targetValue: '50',
        },
      });
    });

    await test.step('What we expected: 200 OK with an image/jpeg response.', async () => {
      expect(response.status()).toBe(200);
      expect(response.headers()['content-type']).toContain('image/jpeg');
    });

    await test.step('What we get: X-Compress-* headers describe the outcome; body is smaller than input.', async () => {
      const headers = response.headers();
      expect(headers).toHaveProperty('x-compress-target-met');
      expect(headers).toHaveProperty('x-compress-target-size');
      expect(headers).toHaveProperty('x-compress-actual-size');
      expect(headers).toHaveProperty('x-compress-format');
      expect(headers).toHaveProperty('x-compress-dimensions');

      const body = await response.body();
      expect(body.length).toBeGreaterThan(0);
      expect(body.length).toBeLessThanOrEqual(original.length);
    });

    await test.step('Why it got this output: the ladder walked down from quality 95 until the encoded bytes fell under the target.', async () => {});
  });

  test('Persona: Compress a JPEG to a specific KB target (size target)', async ({ request }) => {
    await test.step('Why we use this test: users often think in KB, not percent.', async () => {});

    const original = fs.readFileSync(path.join(FIXTURE_DIR, 'dummy.jpg'));
    let response;

    await test.step('What we use: POST with dataType=size, targetValue=20, sizeUnit=KB.', async () => {
      response = await request.post(`${API_URL}/compress-image`, {
        multipart: {
          file: {
            name: 'dummy.jpg',
            mimeType: 'image/jpeg',
            buffer: original,
          },
          conversionType: 'compress',
          dataType: 'size',
          targetValue: '20',
          sizeUnit: 'KB',
        },
      });
    });

    await test.step('What we expected: 200 OK.', async () => {
      expect(response.status()).toBe(200);
    });

    await test.step('What we get: Target-Size matches 20 KB; Target-Met is true (or false with a smaller body).', async () => {
      const headers = response.headers();
      const targetSize = parseInt(headers['x-compress-target-size'], 10);
      expect(targetSize).toBe(20 * 1024);

      const actualSize = parseInt(headers['x-compress-actual-size'], 10);
      const body = await response.body();
      expect(actualSize).toBe(body.length);
    });

    await test.step('Why it got this output: computeTargetBytes multiplied 20 by the KB multiplier and the ladder drove quality accordingly.', async () => {});
  });

  test('Persona: Force PNG output (lossless) from a JPEG input', async ({ request }) => {
    await test.step('Why we use this test: verify targetFormat override works and PNG magic bytes are emitted.', async () => {});

    const original = fs.readFileSync(path.join(FIXTURE_DIR, 'dummy.jpg'));
    let response;

    await test.step('What we use: POST with targetFormat=png.', async () => {
      response = await request.post(`${API_URL}/compress-image`, {
        multipart: {
          file: {
            name: 'dummy.jpg',
            mimeType: 'image/jpeg',
            buffer: original,
          },
          conversionType: 'compress',
          dataType: 'percentage',
          targetValue: '50',
          targetFormat: 'png',
        },
      });
    });

    await test.step('What we expected: 200 OK with image/png.', async () => {
      expect(response.status()).toBe(200);
      expect(response.headers()['content-type']).toContain('image/png');
      expect(response.headers()['x-compress-format']).toBe('png');
    });

    await test.step('What we get: The body starts with the PNG magic bytes.', async () => {
      const body = await response.body();
      expect(body.length).toBeGreaterThan(8);
      // PNG signature: 89 50 4E 47 0D 0A 1A 0A
      expect(body[0]).toBe(0x89);
      expect(body[1]).toBe(0x50);
      expect(body[2]).toBe(0x4E);
      expect(body[3]).toBe(0x47);
    });

    await test.step('Why it got this output: resolveTargetFormat honoured the explicit override and encodeImageStep used image/png.', async () => {});
  });

  // ================================================================
  // GROUP 2: Validation errors
  // ================================================================

  test('Persona: Error - compress percentage above 99', async ({ request }) => {
    await test.step('Why we use this test: 100% compression is a no-op and must be rejected.', async () => {});

    const original = fs.readFileSync(path.join(FIXTURE_DIR, 'dummy.jpg'));
    const response = await request.post(`${API_URL}/compress-image`, {
      multipart: {
        file: { name: 'dummy.jpg', mimeType: 'image/jpeg', buffer: original },
        conversionType: 'compress',
        dataType: 'percentage',
        targetValue: '100',
      },
    });

    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.error).toContain('compress percentage must be between');
  });

  test('Persona: Error - expand with dataType=size', async ({ request }) => {
    await test.step('Why we use this test: expanding toward a specific byte count is impossible via quality alone.', async () => {});

    const original = fs.readFileSync(path.join(FIXTURE_DIR, 'dummy.jpg'));
    const response = await request.post(`${API_URL}/compress-image`, {
      multipart: {
        file: { name: 'dummy.jpg', mimeType: 'image/jpeg', buffer: original },
        conversionType: 'expand',
        dataType: 'size',
        targetValue: '500',
        sizeUnit: 'KB',
      },
    });

    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.error).toContain("expand supports dataType='percentage' only");
  });

  test('Persona: Error - invalid targetFormat', async ({ request }) => {
    await test.step('Why we use this test: only auto/jpg/png/webp are valid.', async () => {});

    const original = fs.readFileSync(path.join(FIXTURE_DIR, 'dummy.jpg'));
    const response = await request.post(`${API_URL}/compress-image`, {
      multipart: {
        file: { name: 'dummy.jpg', mimeType: 'image/jpeg', buffer: original },
        conversionType: 'compress',
        dataType: 'percentage',
        targetValue: '50',
        targetFormat: 'gif',
      },
    });

    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.error).toContain('targetFormat must be');
  });

  test('Persona: Error - missing file field', async ({ request }) => {
    await test.step('Why we use this test: a request without a file must fail cleanly.', async () => {});

    const response = await request.post(`${API_URL}/compress-image`, {
      multipart: {
        conversionType: 'compress',
        dataType: 'percentage',
        targetValue: '50',
      },
    });

    expect(response.status()).toBe(400);
    const body = await response.text();
    expect(body).toContain('file is required');
  });

  test('Persona: Error - wrong HTTP method (GET)', async ({ request }) => {
    await test.step('Why we use this test: the endpoint is POST-only.', async () => {});

    const response = await request.get(`${API_URL}/compress-image`);
    expect(response.status()).toBe(405);
    expect(await response.text()).toContain('Only POST method is allowed');
  });
});