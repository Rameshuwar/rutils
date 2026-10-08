// ============================================================
// GPA-UI.TS — GPA / CGPA calculator view controller
//
// Reads from and writes to the #gpa-calculator-view block in
// index.html. Idempotent init: initGPA() can be called many
// times without re-wiring listeners.
//
// Responsibilities:
//   1. Show/hide field sections based on the selected mode
//   2. Manage dynamic course / semester row lists
//   3. Read form values, POST to /calculate-gpa, render result
//   4. Toggle the steps accordion and copy-to-clipboard
//
// The HTML skeleton (inputs, buttons, result box) lives in
// index.html. This file only wires behavior.
// ============================================================

import {
  calculateGPA,
  GPARequest,
  GPAResponse,
  GPACourse,
  GPASemester,
} from './gpa-service';

// ─── Module-level state ─────────────────────────────────────

let initialized = false;

/** The most recent successful response, cached for the copy button. */
let lastResult: GPAResponse | null = null;

/** Whether the steps accordion is currently expanded. */
let stepsOpen = false;

// ─── Public API ─────────────────────────────────────────────

/**
 * Idempotent initializer. Safe to call multiple times.
 * Resolves DOM references, wires listeners, and paints the
 * default mode's field visibility.
 */
export function initGPA(): void {
  if (initialized) return;

  const root = document.getElementById('gpa-calculator-view');
  if (!root) {
    // View not present in this build — silently skip.
    return;
  }

  initialized = true;

  wireModeSelector();
  wireDynamicLists();
  wireStepsAccordion();
  wireCopyButton();
  wireSubmitButton();

  // Paint the initial mode's sections.
  applyModeConfig(currentMode());
}

// ─── DOM lookup helpers ─────────────────────────────────────

function el<T extends HTMLElement>(id: string): T {
  const node = document.getElementById(id);
  if (!node) {
    throw new Error(`[gpa-ui] missing DOM node #${id}`);
  }
  return node as T;
}

function elOpt<T extends HTMLElement>(id: string): T | null {
  return document.getElementById(id) as T | null;
}

function currentMode(): string {
  const sel = elOpt<HTMLSelectElement>('gpa-mode');
  return sel?.value ?? 'semester_gpa';
}

// ─── Mode selector ──────────────────────────────────────────

function wireModeSelector(): void {
  const sel = elOpt<HTMLSelectElement>('gpa-mode');
  if (!sel) return;

  sel.addEventListener('change', () => {
    applyModeConfig(sel.value);
    // Clear any previous result / error when switching modes.
    hideResult();
    hideStatus();
    lastResult = null;
  });
}

/**
 * Show/hide the field sections according to the selected mode.
 * Every section except the one that matches `mode` gets the
 * `hidden` class.
 */
function applyModeConfig(mode: string): void {
  const sections: Record<string, string> = {
    semester_gpa:        'gpa-semester-gpa-section',
    cumulative_cgpa:     'gpa-cumulative-cgpa-section',
    cgpa_to_percentage:  'gpa-cgpa-to-percentage-section',
    percentage_to_cgpa:  'gpa-percentage-to-cgpa-section',
    grade_to_point:      'gpa-grade-to-point-section',
    target_gpa:          'gpa-target-gpa-section',
  };

  for (const [key, sectionId] of Object.entries(sections)) {
    const section = elOpt<HTMLDivElement>(sectionId);
    if (!section) continue;
    if (key === mode) {
      section.classList.remove('hidden');
    } else {
      section.classList.add('hidden');
    }
  }
}

// ─── Dynamic row lists ──────────────────────────────────────

function wireDynamicLists(): void {
  // "Add course" button
  const addCourseBtn = elOpt<HTMLButtonElement>('gpa-add-course');
  addCourseBtn?.addEventListener('click', () => {
    appendCourseRow();
  });

  // "Add semester" button
  const addSemesterBtn = elOpt<HTMLButtonElement>('gpa-add-semester');
  addSemesterBtn?.addEventListener('click', () => {
    appendSemesterRow();
  });

  // Seed one empty row each so the user sees something on first paint.
  // Guard against double-seeding if initGPA is called twice — the
  // initialized flag already prevents that, but defensive here too.
  const courseList = elOpt<HTMLDivElement>('gpa-course-list');
  if (courseList && courseList.children.length === 0) {
    appendCourseRow();
  }
  const semesterList = elOpt<HTMLDivElement>('gpa-semester-list');
  if (semesterList && semesterList.children.length === 0) {
    appendSemesterRow();
  }
}

/** Create a single course row and append it to #gpa-course-list. */
function appendCourseRow(name = '', credits = '', gradePoint = ''): void {
  const list = elOpt<HTMLDivElement>('gpa-course-list');
  if (!list) return;

  const index = list.children.length + 1;

  const row = document.createElement('div');
  row.className = 'gpa-course-row grid grid-cols-12 gap-2 mb-2 items-center';
  row.innerHTML = `
    <input type="text" class="gpa-course-name col-span-4 bg-white border border-gray-300 rounded-lg py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" placeholder="Course name (optional)" value="${escapeAttr(name)}" />
    <input type="number" step="any" min="0" class="gpa-course-credits col-span-3 bg-white border border-gray-300 rounded-lg py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" placeholder="Credits" value="${escapeAttr(credits)}" />
    <input type="number" step="any" min="0" max="10" class="gpa-course-gp col-span-3 bg-white border border-gray-300 rounded-lg py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" placeholder="Grade point" value="${escapeAttr(gradePoint)}" />
    <button type="button" class="gpa-row-remove col-span-2 px-3 py-2 text-xs font-semibold text-red-600 bg-red-50 hover:bg-red-100 rounded-lg transition-colors" aria-label="Remove course">
      Remove
    </button>
  `;

  row.querySelector('.gpa-row-remove')?.addEventListener('click', () => {
    row.remove();
  });

  // Update all data-index attributes for clarity.
  list.appendChild(row);
  reindexRows(list, 'gpa-course-row');
}

/** Create a single semester row and append it to #gpa-semester-list. */
function appendSemesterRow(name = '', credits = '', gpa = ''): void {
  const list = elOpt<HTMLDivElement>('gpa-semester-list');
  if (!list) return;

  const row = document.createElement('div');
  row.className = 'gpa-semester-row grid grid-cols-12 gap-2 mb-2 items-center';
  row.innerHTML = `
    <input type="text" class="gpa-semester-name col-span-4 bg-white border border-gray-300 rounded-lg py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" placeholder="Semester name (optional)" value="${escapeAttr(name)}" />
    <input type="number" step="any" min="0" class="gpa-semester-credits col-span-3 bg-white border border-gray-300 rounded-lg py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" placeholder="Credits" value="${escapeAttr(credits)}" />
    <input type="number" step="any" min="0" max="10" class="gpa-semester-gpa col-span-3 bg-white border border-gray-300 rounded-lg py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" placeholder="GPA" value="${escapeAttr(gpa)}" />
    <button type="button" class="gpa-row-remove col-span-2 px-3 py-2 text-xs font-semibold text-red-600 bg-red-50 hover:bg-red-100 rounded-lg transition-colors" aria-label="Remove semester">
      Remove
    </button>
  `;

  row.querySelector('.gpa-row-remove')?.addEventListener('click', () => {
    row.remove();
  });

  list.appendChild(row);
  reindexRows(list, 'gpa-semester-row');
}

/** Keeps the numbering tidy when rows are added/removed. Cosmetic. */
function reindexRows(list: HTMLElement, rowClass: string): void {
  const rows = list.querySelectorAll<HTMLDivElement>(`.${rowClass}`);
  rows.forEach((r, i) => {
    r.dataset.index = String(i);
  });
}

// ─── Steps accordion ────────────────────────────────────────

function wireStepsAccordion(): void {
  const toggle = elOpt<HTMLButtonElement>('gpa-steps-toggle');
  const label = elOpt<HTMLSpanElement>('gpa-steps-toggle-label');
  const box = elOpt<HTMLDivElement>('gpa-steps-box');
  if (!toggle || !label || !box) return;

  toggle.addEventListener('click', () => {
    stepsOpen = !stepsOpen;
    if (stepsOpen) {
      box.classList.remove('hidden');
      label.textContent = '▾ Calculation Steps';
    } else {
      box.classList.add('hidden');
      label.textContent = '▸ Calculation Steps';
    }
  });
}

// ─── Copy button ────────────────────────────────────────────

function wireCopyButton(): void {
  const btn = elOpt<HTMLButtonElement>('gpa-copy-btn');
  const label = elOpt<HTMLSpanElement>('gpa-copy-btn-label');
  if (!btn || !label) return;

  btn.addEventListener('click', async () => {
    if (!lastResult) return;

    const text = buildCopyText(lastResult);
    const ok = await copyToClipboard(text);

    const original = label.textContent || 'Copy Result';
    label.textContent = ok ? '✓ Copied' : '⚠ Copy failed';
    btn.disabled = true;
    setTimeout(() => {
      label.textContent = original;
      btn.disabled = false;
    }, 1500);
  });
}

/** Format a full plain-text report of the last result for clipboard. */
function buildCopyText(res: GPAResponse): string {
  const line = '─'.repeat(48);
  const out: string[] = [];
  out.push(line);
  out.push('  GPA / CGPA CALCULATION');
  out.push(line);
  out.push(`  Mode:      ${res.mode}`);
  if (res.gpa !== undefined)         out.push(`  GPA:       ${res.gpa}`);
  if (res.cgpa !== undefined)        out.push(`  CGPA:      ${res.cgpa}`);
  if (res.percentage !== undefined)  out.push(`  Percent:   ${res.percentage}%`);
  if (res.gradePoint !== undefined)  out.push(`  Grade pt:  ${res.gradePoint}`);
  if (res.requiredGpa !== undefined) out.push(`  Required:  ${res.requiredGpa}`);
  if (res.totalCredits !== undefined) out.push(`  Credits:   ${res.totalCredits}`);

  if (res.steps && res.steps.length > 0) {
    out.push('');
    out.push('  Steps:');
    res.steps.forEach((s, i) => out.push(`    ${i + 1}. ${s}`));
  }
  out.push(line);
  return out.join('\n');
}

// ─── Submit handler ─────────────────────────────────────────

function wireSubmitButton(): void {
  const form = elOpt<HTMLFormElement>('gpa-form');
  if (!form) return;

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    hideResult();
    hideStatus();

    let payload: GPARequest;
    try {
      payload = buildPayload(currentMode());
    } catch (err) {
      showStatus((err as Error).message, true);
      return;
    }

    showStatus('Calculating…', false);

    try {
      const res = await calculateGPA(payload);
      lastResult = res;
      renderResult(res);
      hideStatus();
    } catch (err) {
      showStatus((err as Error).message, true);
    }
  });
}

/**
 * Read the visible inputs, validate them lightly, and build the
 * GPARequest payload for the current mode.
 *
 * All validation here is "fail fast with a friendly message" —
 * the backend does the authoritative validation.
 */
function buildPayload(mode: string): GPARequest {
  switch (mode) {
    case 'semester_gpa': {
      const courses = readCourseRows();
      if (courses.length === 0) {
        throw new Error('Add at least one course.');
      }
      return { mode: 'semester_gpa', courses };
    }

    case 'cumulative_cgpa': {
      const semesters = readSemesterRows();
      if (semesters.length === 0) {
        throw new Error('Add at least one semester.');
      }
      return { mode: 'cumulative_cgpa', semesters };
    }

    case 'cgpa_to_percentage': {
      const cgpa = readNumber('gpa-cgpa-input');
      if (cgpa === null) throw new Error('Enter a valid CGPA.');
      const mult = readNumber('gpa-multiplier-input');
      const req: GPARequest = { mode: 'cgpa_to_percentage', cgpa };
      if (mult !== null && mult !== 0) req.multiplier = mult;
      return req;
    }

    case 'percentage_to_cgpa': {
      const pct = readNumber('gpa-percentage-input');
      if (pct === null) throw new Error('Enter a valid percentage.');
      const mult = readNumber('gpa-percentage-multiplier-input');
      const req: GPARequest = { mode: 'percentage_to_cgpa', percentage: pct };
      if (mult !== null && mult !== 0) req.multiplier = mult;
      return req;
    }

    case 'grade_to_point': {
      const gradeEl = elOpt<HTMLInputElement>('gpa-grade-input');
      const grade = gradeEl?.value.trim() ?? '';
      if (!grade) throw new Error('Enter a grade letter.');
      const scaleEl = elOpt<HTMLSelectElement>('gpa-grade-scale');
      const scale = scaleEl?.value ?? '10';
      return { mode: 'grade_to_point', grade, gradingScale: scale };
    }

    case 'target_gpa': {
      const currentCgpa = readNumber('gpa-current-cgpa-input');
      const completedCredits = readNumber('gpa-completed-credits-input');
      const targetCgpa = readNumber('gpa-target-cgpa-input');
      const remainingCredits = readNumber('gpa-remaining-credits-input');
      const scale = readNumber('gpa-scale-input');

      if (currentCgpa === null)      throw new Error('Enter a valid current CGPA.');
      if (completedCredits === null) throw new Error('Enter completed credits.');
      if (targetCgpa === null)       throw new Error('Enter a target CGPA.');
      if (remainingCredits === null) throw new Error('Enter remaining credits.');

      const req: GPARequest = {
        mode: 'target_gpa',
        currentCgpa,
        completedCredits,
        targetCgpa,
        remainingCredits,
      };
      if (scale !== null && scale > 0) req.scale = scale;
      return req;
    }

    default:
      throw new Error(`Unsupported mode: ${mode}`);
  }
}

function readCourseRows(): GPACourse[] {
  const list = elOpt<HTMLDivElement>('gpa-course-list');
  if (!list) return [];

  const rows = list.querySelectorAll<HTMLDivElement>('.gpa-course-row');
  const out: GPACourse[] = [];

  rows.forEach((row) => {
    const name = (row.querySelector<HTMLInputElement>('.gpa-course-name')?.value ?? '').trim();
    const creditsRaw = row.querySelector<HTMLInputElement>('.gpa-course-credits')?.value ?? '';
    const gpRaw = row.querySelector<HTMLInputElement>('.gpa-course-gp')?.value ?? '';

    const credits = Number(creditsRaw);
    const gradePoint = Number(gpRaw);

    // Skip entirely-empty rows silently — a user may add a slot and
    // never fill it in.
    if (creditsRaw === '' && gpRaw === '' && name === '') return;

    if (!Number.isFinite(credits) || credits <= 0) {
      throw new Error('Every course needs a positive credits value.');
    }
    if (!Number.isFinite(gradePoint) || gradePoint < 0 || gradePoint > 10) {
      throw new Error('Every course needs a grade point between 0 and 10.');
    }

    const c: GPACourse = { credits, gradePoint };
    if (name) c.name = name;
    out.push(c);
  });

  return out;
}

function readSemesterRows(): GPASemester[] {
  const list = elOpt<HTMLDivElement>('gpa-semester-list');
  if (!list) return [];

  const rows = list.querySelectorAll<HTMLDivElement>('.gpa-semester-row');
  const out: GPASemester[] = [];

  rows.forEach((row) => {
    const name = (row.querySelector<HTMLInputElement>('.gpa-semester-name')?.value ?? '').trim();
    const creditsRaw = row.querySelector<HTMLInputElement>('.gpa-semester-credits')?.value ?? '';
    const gpaRaw = row.querySelector<HTMLInputElement>('.gpa-semester-gpa')?.value ?? '';

    const credits = Number(creditsRaw);
    const gpa = Number(gpaRaw);

    if (creditsRaw === '' && gpaRaw === '' && name === '') return;

    if (!Number.isFinite(credits) || credits <= 0) {
      throw new Error('Every semester needs a positive credits value.');
    }
    if (!Number.isFinite(gpa) || gpa < 0 || gpa > 10) {
      throw new Error('Every semester needs a GPA between 0 and 10.');
    }

    const s: GPASemester = { credits, gpa };
    if (name) s.name = name;
    out.push(s);
  });

  return out;
}

function readNumber(id: string): number | null {
  const node = elOpt<HTMLInputElement>(id);
  if (!node) return null;
  const raw = node.value.trim();
  if (raw === '') return null;
  const n = Number(raw);
  return Number.isFinite(n) ? n : null;
}

// ─── Result rendering ───────────────────────────────────────

function renderResult(res: GPAResponse): void {
  const box = elOpt<HTMLDivElement>('gpa-result-box');
  const headline = elOpt<HTMLSpanElement>('gpa-result-formatted');
  const chips = elOpt<HTMLDivElement>('gpa-extra-chips');
  const stepsList = elOpt<HTMLOListElement>('gpa-steps-list');
  const stepsBox = elOpt<HTMLDivElement>('gpa-steps-box');
  const stepsLabel = elOpt<HTMLSpanElement>('gpa-steps-toggle-label');

  if (!box || !headline || !chips || !stepsList) return;

  headline.textContent = res.formatted;

  // Build the chips row from the mode-relevant supplementary fields.
  chips.innerHTML = '';
  const chipData: [string, string | number][] = [];

  const pushChip = (label: string, value: string | number | undefined) => {
    if (value === undefined || value === null || value === '') return;
    chipData.push([label, value]);
  };

  if (res.totalCredits !== undefined)  pushChip('Total Credits', res.totalCredits);
  if (res.multiplier !== undefined)    pushChip('Multiplier', res.multiplier);
  if (res.currentCgpa !== undefined)   pushChip('Current CGPA', res.currentCgpa);
  if (res.targetCgpa !== undefined)    pushChip('Target CGPA', res.targetCgpa);

  if (res.extra) {
    if (typeof res.extra.achievable === 'boolean') {
      pushChip('Achievable', res.extra.achievable ? 'Yes' : 'No');
    }
    if (typeof res.extra.scaleNote === 'string') {
      pushChip('Convention', res.extra.scaleNote);
    }
    if (typeof res.extra.scale === 'number' || typeof res.extra.scale === 'string') {
      pushChip('Scale', String(res.extra.scale));
    }
  }

  for (const [label, value] of chipData) {
    const chip = document.createElement('span');
    chip.className =
      'inline-flex items-center gap-1 bg-white border border-violet-200 text-violet-800 text-xs font-medium px-3 py-1 rounded-full';
    chip.textContent = `${label}: ${value}`;
    chips.appendChild(chip);
  }

  // Steps
  stepsList.innerHTML = '';
  if (Array.isArray(res.steps) && res.steps.length > 0) {
    for (const s of res.steps) {
      const li = document.createElement('li');
      li.textContent = s;
      stepsList.appendChild(li);
    }
    // Collapsed by default on each new result.
    stepsOpen = false;
    stepsBox?.classList.add('hidden');
    if (stepsLabel) stepsLabel.textContent = '▸ Calculation Steps';
  }

  // Optional warning from extra.warning (e.g. impossible target_gpa)
  const warnBox = elOpt<HTMLParagraphElement>('gpa-warning-message');
  if (warnBox) {
    const w = res.extra?.warning;
    if (typeof w === 'string' && w.length > 0) {
      warnBox.textContent = w;
      warnBox.classList.remove('hidden');
    } else {
      warnBox.classList.add('hidden');
      warnBox.textContent = '';
    }
  }

  box.classList.remove('hidden');
}

function hideResult(): void {
  elOpt<HTMLDivElement>('gpa-result-box')?.classList.add('hidden');
  elOpt<HTMLParagraphElement>('gpa-warning-message')?.classList.add('hidden');
  lastResult = null;
}

// ─── Status message ─────────────────────────────────────────

function showStatus(message: string, isError: boolean): void {
  const node = elOpt<HTMLParagraphElement>('gpa-status-message');
  if (!node) return;
  node.textContent = message;
  node.classList.remove('hidden', 'text-red-600', 'text-green-600', 'text-gray-500');
  node.classList.add(isError ? 'text-red-600' : 'text-gray-500');
}

function hideStatus(): void {
  elOpt<HTMLParagraphElement>('gpa-status-message')?.classList.add('hidden');
}

// ─── Small utilities ────────────────────────────────────────

function escapeAttr(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;');
}

async function copyToClipboard(text: string): Promise<boolean> {
  if (navigator.clipboard && window.isSecureContext) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      /* fall through */
    }
  }
  try {
    const ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.left = '-9999px';
    document.body.appendChild(ta);
    ta.select();
    const ok = document.execCommand('copy');
    document.body.removeChild(ta);
    return ok;
  } catch {
    return false;
  }
}