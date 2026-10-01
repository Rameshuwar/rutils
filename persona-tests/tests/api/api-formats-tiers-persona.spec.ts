import { test, expect } from '@playwright/test';

// Persona tests for the Tier-1 formatter registry.
//
// This file is deliberately different from the other specs: it asserts
// the EXACT set of 45 edges that the Tier-1 release ships. If you add
// a new format in the future, you update this file first — it's the
// "truth table" that the rest of the suite trusts.
//
// Why split it out? Because a regression can take two forms:
//   1. A previously-working edge breaks → caught by api-matrix or
//      api-formats-convert-consistency.
//   2. An edge silently disappears from the registry → caught HERE.
//
// Form (2) is the most dangerous: the frontend keeps rendering the
// dropdown option but every request 400s. This file makes that
// impossible.

const API_URL = 'http://localhost:8080';

// The canonical Tier-1 set. Keys are "from", values are the set of
// reachable "to" formats. Both sides are lower-case.
//
// When you add a new format or a new edge, add it here FIRST, then
// implement it. The tests will fail until the implementation catches
// up — which is exactly what you want.
const TIER_1_EDGES: Record<string, string[]> = {
  // -- image → image / pdf / txt --
  jpg:  ['pdf', 'png', 'txt'],
  png:  ['jpg', 'pdf', 'txt'],
  webp: ['jpg', 'png', 'txt'],
  tiff: ['jpg', 'png', 'txt'],
  bmp:  ['jpg', 'png', 'txt'],

  // -- pdf → image / document --
  pdf:  ['png', 'jpg', 'docx', 'txt', 'csv', 'json'],

  // -- document family --
  txt:  ['docx', 'pdf', 'json', 'csv', 'jpg', 'png'],
  docx: ['csv', 'txt', 'json', 'pdf', 'jpg', 'png'],

  // -- data family --
  csv:  ['json', 'pdf', 'txt', 'docx', 'jpg', 'png'],
  json: ['csv', 'txt', 'docx', 'pdf', 'jpg', 'png'],
};

// Flatten to a set of "from→to" strings for direct membership checks.
const TIER_1_SET = new Set<string>();
for (const [from, tos] of Object.entries(TIER_1_EDGES)) {
  for (const to of tos) {
    TIER_1_SET.add(`${from}→${to}`);
  }
}

test.describe('Backend API Format Registry — Tier-1 Truth Persona', () => {

  test('Persona: Registry has exactly the expected 45 edges', async ({ request }) => {
    await test.step('Why we use this test: The advertised edge count must match the code\'s intent. If it drops, a formatter file was deleted or a Register() call went missing. If it grows, the truth file wasn\'t updated.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    expect(response.status()).toBe(200);

    const body = await response.json();

    expect(
      body.conversions.length,
      `expected exactly ${TIER_1_SET.size} edges, got ${body.conversions.length}`,
    ).toBe(TIER_1_SET.size);
  });

  test('Persona: Every advertised edge matches the Tier-1 truth table', async ({ request }) => {
    await test.step('Why we use this test: This is the strictest assertion in the suite. Every edge the server advertises must be one that the truth table knows about — no extras, no omissions.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    const serverSet = new Set<string>(
      body.conversions.map((e: { from: string; to: string }) => `${e.from}→${e.to}`),
    );

    // Every server-advertised edge must be in the truth table.
    for (const edge of serverSet) {
      expect(
        TIER_1_SET.has(edge),
        `server advertises "${edge}" but it is not in the Tier-1 truth table`,
      ).toBe(true);
    }

    // Every truth-table edge must be advertised by the server.
    for (const edge of TIER_1_SET) {
      expect(
        serverSet.has(edge),
        `truth table expects "${edge}" but the server does not advertise it`,
      ).toBe(true);
    }
  });

  test('Persona: Image pairs are all registered', async ({ request }) => {
    await test.step('Why we use this test: The image-family edges were the primary motivation for this release. Verify they all landed.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    const serverSet = new Set<string>(
      body.conversions.map((e: { from: string; to: string }) => `${e.from}→${e.to}`),
    );

    const imageEdges = [
      'jpg→png', 'jpg→pdf', 'jpg→txt',
      'png→jpg', 'png→pdf', 'png→txt',
      'webp→jpg', 'webp→png', 'webp→txt',
      'tiff→jpg', 'tiff→png', 'tiff→txt',
      'bmp→jpg', 'bmp→png', 'bmp→txt',
      'pdf→png', 'pdf→jpg',
    ];

    for (const edge of imageEdges) {
      expect(serverSet.has(edge), `image edge ${edge} is missing`).toBe(true);
    }
  });

  test('Persona: Document pairs are all registered', async ({ request }) => {
    await test.step('Why we use this test: The document family covers the pre-existing converter behavior. Verify the refactor didn\'t drop any of them.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    const serverSet = new Set<string>(
      body.conversions.map((e: { from: string; to: string }) => `${e.from}→${e.to}`),
    );

    const documentEdges = [
      'pdf→docx', 'pdf→txt', 'pdf→csv', 'pdf→json',
      'txt→docx', 'txt→pdf', 'txt→json', 'txt→csv',
      'docx→csv', 'docx→txt', 'docx→json', 'docx→pdf',
    ];

    for (const edge of documentEdges) {
      expect(serverSet.has(edge), `document edge ${edge} is missing`).toBe(true);
    }
  });

  test('Persona: Data pairs are all registered', async ({ request }) => {
    await test.step('Why we use this test: The CSV/JSON family is the original core use case. Any missing edge here is a regression.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    const serverSet = new Set<string>(
      body.conversions.map((e: { from: string; to: string }) => `${e.from}→${e.to}`),
    );

    const dataEdges = [
      'csv→json', 'csv→pdf', 'csv→txt', 'csv→docx',
      'json→csv', 'json→txt', 'json→docx', 'json→pdf',
    ];

    for (const edge of dataEdges) {
      expect(serverSet.has(edge), `data edge ${edge} is missing`).toBe(true);
    }
  });

  test('Persona: Document-to-image pairs are all registered', async ({ request }) => {
    await test.step('Why we use this test: These render documents to images via an intermediate PDF. Any regression here would break a common download use case.', async () => {});

    const response = await request.get(`${API_URL}/formats`);
    const body = await response.json();

    const serverSet = new Set<string>(
      body.conversions.map((e: { from: string; to: string }) => `${e.from}→${e.to}`),
    );

    const docImageEdges = [
      'csv→jpg', 'csv→png',
      'json→jpg', 'json→png',
      'txt→jpg', 'txt→png',
      'docx→jpg', 'docx→png',
    ];

    for (const edge of docImageEdges) {
      expect(serverSet.has(edge), `doc→image edge ${edge} is missing`).toBe(true);
    }
  });
});