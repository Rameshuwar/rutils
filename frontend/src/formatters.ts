// ============================================================
// FORMATTERS.TS — File Repair / Formatter UI controller
//
// Renders the formatter tiles, handles the repair form submission,
// and downloads the repaired file from the /repair backend endpoint.
//
// Exports:
//   initFormatters()          — resolve DOM refs, wire listeners (idempotent)
//   renderFormattersCategory() — called from main.ts when the user
//                                clicks "Formatters" in the sidebar
// ============================================================

// ------------------------------------------------------------
// Supported formats
// ------------------------------------------------------------
interface FormatterFormat {
  id: string;
  label: string;
  ext: string;
  icon: string;
  accept: string;
}

const FORMATTER_FORMATS: FormatterFormat[] = [
  { id: 'json', label: 'JSON', ext: '.json',  icon: '{}', accept: '.json' },
  { id: 'csv',  label: 'CSV',  ext: '.csv',   icon: '≣',  accept: '.csv' },
  { id: 'xml',  label: 'XML',  ext: '.xml',   icon: '<>', accept: '.xml' },
  { id: 'yaml', label: 'YAML', ext: '.yaml',  icon: 'Y:', accept: '.yaml,.yml' },
  { id: 'html', label: 'HTML', ext: '.html',  icon: '</>', accept: '.html,.htm' },
  { id: 'toml', label: 'TOML', ext: '.toml',  icon: 'T:', accept: '.toml' },
  { id: 'ini',  label: 'INI',  ext: '.ini',   icon: 'I:', accept: '.ini' },
  { id: 'md',   label: 'MD',   ext: '.md',    icon: 'M↓', accept: '.md,.markdown' },
];

// ------------------------------------------------------------
// Module state
// ------------------------------------------------------------
let initialized = false;
let activeFormat: string = '';

// ------------------------------------------------------------
// Public API — idempotent init
// ------------------------------------------------------------
export function initFormatters(): void {
  if (initialized) return;

  // The DOM refs only exist if index.html contains #formatters-view.
  // If it doesn't, we bail out early (so the rest of the app keeps
  // working and the browser console tells you exactly what's missing).
  const grid = document.getElementById('formatters-grid');
  if (!grid) {
    console.warn(
      '[formatters] #formatters-grid not found. ' +
      'Add the <div id="formatters-view"> block to index.html.',
    );
    return;
  }

  initialized = true;

  renderTiles(grid);
  wireRepairForm();
}

// Called from main.ts every time the user clicks "Formatters" in the
// sidebar. Safe to call repeatedly — initFormatters() is a no-op after
// the first run.
export function renderFormattersCategory(): void {
  initFormatters();

  // If the user re-enters the category, keep their last-selected format
  // highlighted and the repair panel open. If nothing was selected yet,
  // just leave the grid showing.
  if (activeFormat) {
    const panel = document.getElementById('formatters-repair-panel');
    panel?.classList.remove('hidden');
  }
}

// ------------------------------------------------------------
// Tiles
// ------------------------------------------------------------
function renderTiles(grid: HTMLElement): void {
  grid.innerHTML = '';

  FORMATTER_FORMATS.forEach(fmt => {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.dataset.format = fmt.id;
    btn.className = [
      'formatters-tile',
      'flex',
      'flex-col',
      'items-center',
      'justify-center',
      'gap-1.5',
      'py-5',
      'px-3',
      'rounded-xl',
      'border',
      'border-gray-300',
      'bg-white',
      'text-gray-700',
      'font-medium',
      'text-sm',
      'hover:bg-amber-50',
      'hover:border-amber-400',
      'hover:text-amber-700',
      'transition-colors',
      'cursor-pointer',
    ].join(' ');

    btn.innerHTML = `
      <span class="text-2xl font-mono text-amber-600 leading-none">${escapeHtml(fmt.icon)}</span>
      <span class="mt-1">${escapeHtml(fmt.label)}</span>
      <span class="text-[11px] text-gray-400">${escapeHtml(fmt.ext)}</span>
    `;

    btn.addEventListener('click', () => selectFormat(fmt.id));

    grid.appendChild(btn);
  });
}

function selectFormat(formatId: string): void {
  activeFormat = formatId;
  const fmt = FORMATTER_FORMATS.find(f => f.id === formatId);

  // Highlight the active tile.
  document.querySelectorAll<HTMLElement>('.formatters-tile').forEach(tile => {
    const isActive = tile.dataset.format === formatId;
    tile.classList.toggle('bg-amber-100', isActive);
    tile.classList.toggle('border-amber-500', isActive);
    tile.classList.toggle('text-amber-800', isActive);
    tile.classList.toggle('ring-2', isActive);
    tile.classList.toggle('ring-amber-200', isActive);
  });

  // Reveal the repair panel.
  const panel = document.getElementById('formatters-repair-panel');
  panel?.classList.remove('hidden');

  // Update the header text.
  const label = document.getElementById('formatters-active-format');
  if (label && fmt) label.textContent = fmt.label;

  // Update the file input's accept attribute so the OS picker shows
  // only relevant files (nice UX touch).
  const fileInput = document.getElementById('formatters-repair-file') as HTMLInputElement | null;
  if (fileInput && fmt) {
    fileInput.accept = fmt.accept;
    fileInput.value = '';
  }

  // Reset any prior result.
  document.getElementById('formatters-repair-report')?.classList.add('hidden');
  const status = document.getElementById('formatters-repair-status');
  if (status) {
    status.classList.add('hidden');
    status.textContent = '';
  }
}

// ------------------------------------------------------------
// Repair form
// ------------------------------------------------------------
function wireRepairForm(): void {
  const form = document.getElementById('formatters-repair-form') as HTMLFormElement | null;
  if (!form) {
    console.warn('[formatters] #formatters-repair-form not found.');
    return;
  }
  if ((form as any)._formattersWired) return; // idempotent
  (form as any)._formattersWired = true;

  form.addEventListener('submit', handleRepairSubmit);
}

async function handleRepairSubmit(e: Event): Promise<void> {
  e.preventDefault();

  const fileInput = document.getElementById('formatters-repair-file')     as HTMLInputElement;
  const levelSel  = document.getElementById('formatters-repair-level')    as HTMLSelectElement;
  const describe  = document.getElementById('formatters-repair-describe') as HTMLInputElement;
  const statusEl  = document.getElementById('formatters-repair-status')   as HTMLParagraphElement;
  const reportBox = document.getElementById('formatters-repair-report')   as HTMLDivElement;

  statusEl.classList.remove('hidden', 'text-red-600', 'text-green-600');
  statusEl.classList.add('text-gray-500');
  reportBox.classList.add('hidden');

  if (!activeFormat) {
    statusEl.textContent = 'Pick a format above first.';
    statusEl.classList.replace('text-gray-500', 'text-red-600');
    return;
  }

  if (!fileInput.files || fileInput.files.length === 0) {
    statusEl.textContent = 'Please select a file to repair.';
    statusEl.classList.replace('text-gray-500', 'text-red-600');
    return;
  }

  const fd = new FormData();
  fd.append('file', fileInput.files[0]);
  fd.append('repairLevel', levelSel.value);
  if (describe.checked) fd.append('describe', 'true');

  try {
    statusEl.textContent = 'Repairing…';

    const isLocal =
      window.location.hostname === 'localhost' ||
      window.location.hostname === '127.0.0.1';

    const apiUrl = isLocal
      ? 'http://localhost:8080/repair'
      : 'https://utils.api.srilakshmiretail.in/repair';

    const res = await fetch(apiUrl, { method: 'POST', body: fd });

    if (!res.ok) {
      const errText = await res.text().catch(() => '');
      throw new Error(errText || `Server error: ${res.status}`);
    }

    const ct = (res.headers.get('content-type') || '').toLowerCase();

    if (ct.includes('application/json')) {
      const json = await res.json();

      populateReport({
        format:  json.format,
        level:   json.level,
        changed: json.changed,
        applied: json.applied,
        warnings: json.warnings,
      });

      // Download the repaired payload.
      const ext = activeFormat || json.format || 'txt';
      downloadText(json.output ?? '', `repaired.${ext}`);

      statusEl.textContent = json.changed
        ? 'Repair complete — file downloaded.'
        : 'File was already valid — downloaded unchanged.';
      statusEl.classList.replace('text-gray-500', 'text-green-600');

    } else {
      // describe=false path — raw file stream.
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = `repaired.${activeFormat || 'txt'}`;
      document.body.appendChild(a);
      a.click();
      URL.revokeObjectURL(url);
      document.body.removeChild(a);

      // Populate the report card from the X-Repair-* response headers.
      populateReport({
        format:  res.headers.get('X-Repair-Format')  ?? undefined,
        level:   res.headers.get('X-Repair-Level')   ?? undefined,
        changed: res.headers.get('X-Repair-Changed') === 'true',
        applied: parseAppliedHeader(res.headers.get('X-Repair-Applied')),
      });

      statusEl.textContent = 'Repair complete — file downloaded.';
      statusEl.classList.replace('text-gray-500', 'text-green-600');
    }

  } catch (err) {
    console.error('[formatters] repair error:', err);
    statusEl.textContent = `Error: ${err instanceof Error ? err.message : 'Unknown error occurred'}`;
    statusEl.classList.replace('text-gray-500', 'text-red-600');
  }
}

// ------------------------------------------------------------
// Report card
// ------------------------------------------------------------
interface ReportPayload {
  format?: string;
  level?: string;
  changed?: boolean;
  applied?: string[];
  warnings?: string[];
}

function populateReport(json: ReportPayload): void {
  const reportBox = document.getElementById('formatters-repair-report');
  reportBox?.classList.remove('hidden');

  setText('formatters-result-format',  json.format ?? '—');
  setText('formatters-result-level',   json.level  ?? '—');
  setText('formatters-result-changed', json.changed ? 'Yes' : 'No');

  const fixesWrap = document.getElementById('formatters-result-fixes');
  if (fixesWrap) {
    fixesWrap.innerHTML = '';
    const list =
      json.applied && json.applied.length > 0
        ? json.applied
        : ['No fixes applied'];

    list.forEach(fix => {
      const chip = document.createElement('span');
      chip.className = 'inline-flex items-center gap-1 bg-white border border-amber-200 text-amber-800 text-xs font-medium px-3 py-1 rounded-full';
      chip.textContent = fix;
      fixesWrap.appendChild(chip);
    });
  }

  const warnWrap = document.getElementById('formatters-result-warnings-wrap');
  const warnList = document.getElementById('formatters-result-warnings');
  if (warnWrap && warnList) {
    warnList.innerHTML = '';
    if (json.warnings && json.warnings.length > 0) {
      json.warnings.forEach(w => {
        const li = document.createElement('li');
        li.textContent = w;
        warnList.appendChild(li);
      });
      warnWrap.classList.remove('hidden');
    } else {
      warnWrap.classList.add('hidden');
    }
  }
}

// ------------------------------------------------------------
// Small helpers
// ------------------------------------------------------------
function setText(id: string, text: string): void {
  const el = document.getElementById(id);
  if (el) el.textContent = text;
}

function downloadText(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.style.display = 'none';
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  URL.revokeObjectURL(url);
  document.body.removeChild(a);
}

// Parses the compact header emitted by the backend:
//   "json;level=normal;changed=true;fixed=trailing-comma,single-quotes;warn=1"
function parseAppliedHeader(header: string | null): string[] {
  if (!header || header === 'none') return [];

  // The "fixed=" segment holds a comma-separated list of fix names.
  const parts = header.split(';');
  for (const part of parts) {
    if (part.startsWith('fixed=')) {
      return part
        .slice('fixed='.length)
        .split(',')
        .map(s => s.trim())
        .filter(Boolean);
    }
  }
  return [];
}

// Escapes any HTML in user-facing strings (icons/labels) so a stray
// `<` doesn't break the layout.
function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}