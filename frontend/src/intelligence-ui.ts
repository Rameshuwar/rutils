import { getAuthState } from './auth';

// ─── State ────────────────────────────────────────────────
let _currentData: any = null;
let _currentTab: 'quarterly' | 'annual' = 'quarterly';

// ─── Entrypoints ──────────────────────────────────────────
export async function openCompanyIntelligence(symbol: string) {
  const container = document.getElementById('company-intelligence-view');
  if (!container) return;

  document.querySelectorAll('.tool-view').forEach(v => v.classList.add('hidden'));
  container.classList.remove('hidden');

  const categoryTitle = document.getElementById('category-title');
  if (categoryTitle) categoryTitle.textContent = 'Company Intelligence';
  const breadcrumbTool = document.getElementById('breadcrumb-tool');
  if (breadcrumbTool) breadcrumbTool.textContent = symbol;

  _currentTab = 'quarterly';
  await renderCompanyIntelligence(symbol);
}

export function closeCompanyIntelligence() {
  const container = document.getElementById('company-intelligence-view');
  if (!container) return;
  container.classList.add('hidden');
  const tab = document.querySelector('button[data-tool="nifty50"]') as HTMLButtonElement;
  if (tab) tab.click();
}

// Tab switcher — called inline from HTML
(window as any).ciSwitchTab = function (tab: 'quarterly' | 'annual') {
  _currentTab = tab;
  const qBtn = document.getElementById('ci-tab-quarterly');
  const aBtn = document.getElementById('ci-tab-annual');
  if (!qBtn || !aBtn) return;
  qBtn.className = `ci-tab-btn ${tab === 'quarterly' ? 'active' : 'inactive'}`;
  aBtn.className = `ci-tab-btn ${tab === 'annual' ? 'active' : 'inactive'}`;
  if (_currentData) renderFinancials(_currentData);
};

(window as any).openCompanyIntelligence = openCompanyIntelligence;
(window as any).closeCompanyIntelligence = closeCompanyIntelligence;

// ─── API Fetch ────────────────────────────────────────────
async function fetchIntelligenceData(symbol: string) {
  const { token } = getAuthState();
  const isLocal = location.hostname === 'localhost' || location.hostname === '127.0.0.1';
  const base = isLocal ? (location.port === '8080' ? '' : 'http://localhost:8080') : 'https://utils.api.srilakshmiretail.in';
  const headers: HeadersInit = {};
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const r = await fetch(`${base}/nifty50/companies/${symbol}/intelligence`, { headers });
  if (!r.ok) throw new Error(`HTTP ${r.status}`);
  return r.json();
}

// ─── Main Render ──────────────────────────────────────────
async function renderCompanyIntelligence(symbol: string) {
  const contentEl  = document.getElementById('ci-content');
  const loadingEl  = document.getElementById('ci-loading');
  const errorEl    = document.getElementById('ci-error');
  const errorMsg   = document.getElementById('ci-error-msg');
  if (!contentEl || !loadingEl || !errorEl) return;

  loadingEl.classList.remove('hidden');
  contentEl.classList.add('hidden');
  errorEl.classList.add('hidden');

  try {
    const data = await fetchIntelligenceData(symbol);
    _currentData = data;

    // ── Header ────────────────────────────────────────────
    setText('ci-company-name', data.company_name ?? '—');
    setText('ci-symbol', data.symbol ?? '');
    setText('ci-industry', data.industry ?? '');
    setText('ci-updated-at', data.updated_at
      ? new Date(data.updated_at).toLocaleString('en-IN', { dateStyle: 'medium', timeStyle: 'short' })
      : '—');

    // ── KPI Cards ─────────────────────────────────────────
    renderKPICards(data);

    // ── Company Profile ───────────────────────────────────
    const profileEl = document.getElementById('ci-profile-desc')!;
    if (data.profile?.description) {
      profileEl.innerHTML = `<p>${escHtml(data.profile.description)}</p>`;
    } else {
      profileEl.innerHTML = unavailable();
    }

    // ── Financials ────────────────────────────────────────
    renderFinancials(data);

    // ── Results Calendar ──────────────────────────────────
    renderCalendar(data.results_calendar);

    // ── News Feed ─────────────────────────────────────────
    renderNews(data.news);

    // ── Disclaimer ────────────────────────────────────────
    setText('ci-disclaimer',
      "RUtils provides company information collected from publicly available sources for informational and educational purposes only. " +
      "Data may be delayed, incomplete or incorrectly extracted. Please independently verify important information on the company's " +
      "official website and stock exchange filings before making any decision. This is not investment advice.");

    loadingEl.classList.add('hidden');
    contentEl.classList.remove('hidden');
  } catch (err) {
    console.error(err);
    loadingEl.classList.add('hidden');
    errorEl.classList.remove('hidden');
    if (errorMsg) errorMsg.textContent = `Failed to load intelligence data: ${(err as Error).message}`;
  }
}

// ─── KPI Cards ────────────────────────────────────────────
function renderKPICards(data: any) {
  const wrap = document.getElementById('ci-kpi-cards');
  if (!wrap) return;

  const results: any[] = data.quarterly_results ?? [];
  const latest  = results[0];
  const prev    = results[1];

  const kpis = [
    {
      label: 'Revenue (Cr)',
      value: latest?.revenue ?? '—',
      prev:  prev?.revenue,
      color: '#6366f1',
      bg:    'rgba(99,102,241,0.12)',
      icon:  '₹',
    },
    {
      label: 'Net Profit (Cr)',
      value: latest?.net_profit ?? '—',
      prev:  prev?.net_profit,
      color: '#10b981',
      bg:    'rgba(16,185,129,0.12)',
      icon:  '↗',
    },
    {
      label: 'Op. Profit (Cr)',
      value: latest?.operating_profit ?? '—',
      prev:  prev?.operating_profit,
      color: '#f59e0b',
      bg:    'rgba(245,158,11,0.12)',
      icon:  '⚡',
    },
    {
      label: 'EPS (₹)',
      value: latest?.eps ?? '—',
      prev:  prev?.eps,
      color: '#8b5cf6',
      bg:    'rgba(139,92,246,0.12)',
      icon:  '◈',
    },
  ];

  wrap.innerHTML = kpis.map(k => {
    const { trend, arrow } = calcTrend(k.value, k.prev);
    return `
      <div class="ci-kpi-card" style="background:${k.bg};border:1px solid ${k.color}30">
        <div style="font-size:22px;margin-bottom:6px;opacity:.8">${k.icon}</div>
        <div style="font-size:18px;font-weight:800;color:#fff;letter-spacing:-0.03em;line-height:1.1">
          ${k.value !== '—' ? k.value : '<span style="opacity:.4">—</span>'}
        </div>
        <div style="font-size:11px;color:rgba(255,255,255,.55);margin-top:4px;font-weight:600;text-transform:uppercase;letter-spacing:.04em">${k.label}</div>
        ${trend ? `<div style="font-size:11px;margin-top:6px;font-weight:700;color:${trend === 'up' ? '#34d399' : '#f87171'}">${arrow} vs prev Q</div>` : ''}
      </div>`;
  }).join('');
}

function calcTrend(cur: string, prev?: string): { trend: string; arrow: string } {
  if (!prev || cur === '—') return { trend: '', arrow: '' };
  const c = parseNum(cur), p = parseNum(prev);
  if (isNaN(c) || isNaN(p) || p === 0) return { trend: '', arrow: '' };
  const pct = ((c - p) / Math.abs(p) * 100).toFixed(1);
  return c >= p
    ? { trend: 'up',   arrow: `▲ ${pct}%` }
    : { trend: 'down', arrow: `▼ ${Math.abs(+pct)}%` };
}

function parseNum(s: string): number {
  return parseFloat(s.replace(/,/g, ''));
}

// ─── Financials: Bar Chart + Table ────────────────────────
function renderFinancials(data: any) {
  const results: any[] = (_currentTab === 'quarterly'
    ? data.quarterly_results
    : data.annual_results) ?? [];

  renderBarChart(results);
  renderFinTable(results);
}

function renderBarChart(results: any[]) {
  const wrap = document.getElementById('ci-bar-chart');
  if (!wrap) return;
  if (!results.length) { wrap.innerHTML = ''; return; }

  // We show up to 6 bars for net profit
  const data = [...results].reverse().slice(-6);
  const maxVal = Math.max(...data.map(r => parseNum(r.net_profit || '0')).filter(v => !isNaN(v)), 1);

  const colors = ['#6366f1','#8b5cf6','#a78bfa','#7c3aed','#4f46e5','#818cf8'];

  wrap.innerHTML = `
    <div style="display:flex;flex-direction:column;gap:4px">
      <div style="font-size:11px;color:#94a3b8;font-weight:600;text-transform:uppercase;letter-spacing:.05em;margin-bottom:4px">Net Profit Trend (₹ Cr)</div>
      <div style="display:flex;align-items:flex-end;gap:6px;height:80px;padding-bottom:0">
        ${data.map((r, i) => {
          const v = parseNum(r.net_profit || '0');
          const h = isNaN(v) ? 4 : Math.max(4, Math.round((v / maxVal) * 76));
          const color = colors[i % colors.length];
          return `
            <div style="flex:1;display:flex;flex-direction:column;align-items:center;gap:3px;cursor:default"
                 title="${r.period}: ₹${r.net_profit || '—'} Cr">
              <div style="font-size:9px;color:#94a3b8;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;width:100%;text-align:center">
                ${r.net_profit || '—'}
              </div>
              <div class="ci-chart-bar" style="height:${h}px;width:100%;background:${color};border-radius:4px 4px 0 0;opacity:.85;transition:opacity .15s" onmouseover="this.style.opacity=1" onmouseout="this.style.opacity=.85"></div>
              <div style="font-size:9px;color:#94a3b8;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;width:100%;text-align:center">${r.period ?? ''}</div>
            </div>`;
        }).join('')}
      </div>
    </div>`;
}

function renderFinTable(results: any[]) {
  const wrap = document.getElementById('ci-fin-table-wrap');
  if (!wrap) return;
  if (!results.length) {
    wrap.innerHTML = unavailable();
    return;
  }

  const rows = [
    { key: 'revenue',          label: 'Revenue',        color: '#6366f1' },
    { key: 'operating_profit', label: 'Op. Profit',     color: '#f59e0b' },
    { key: 'net_profit',       label: 'Net Profit',     color: '#10b981' },
    { key: 'eps',              label: 'EPS (₹)',         color: '#8b5cf6' },
  ];

  wrap.innerHTML = `
    <table class="ci-tbl">
      <thead>
        <tr>
          <th style="text-align:left">Metric</th>
          ${results.map(r => `<th>${escHtml(r.period ?? '')}</th>`).join('')}
        </tr>
      </thead>
      <tbody>
        ${rows.map(row => {
          const vals = results.map(r => r[row.key] || '—');
          // Calculate max for in-row mini-bar
          const nums = vals.map(v => parseNum(v)).filter(v => !isNaN(v));
          const max = Math.max(...nums, 1);
          return `
            <tr>
              <td>
                <span style="display:inline-block;width:8px;height:8px;border-radius:50%;background:${row.color};margin-right:7px;vertical-align:middle;"></span>
                ${row.label}
              </td>
              ${vals.map((v, i) => {
                const n = parseNum(v);
                const w = isNaN(n) ? 0 : Math.round((n / max) * 56);
                const prev = i > 0 ? parseNum(vals[i - 1]) : NaN;
                const trendColor = !isNaN(n) && !isNaN(prev) ? (n >= prev ? '#10b981' : '#ef4444') : 'transparent';
                return `
                  <td>
                    <div style="display:flex;flex-direction:column;align-items:flex-end;gap:3px">
                      <span style="color:${trendColor !== 'transparent' && i > 0 ? trendColor : 'inherit'};font-weight:${i === 0 ? '700' : '400'}">${escHtml(v)}</span>
                      <div style="height:3px;width:${w}px;background:${row.color};border-radius:2px;opacity:.6;transition:width .5s"></div>
                    </div>
                  </td>`;
              }).join('')}
            </tr>`;
        }).join('')}
      </tbody>
    </table>`;
}

// ─── Results Calendar ─────────────────────────────────────
function renderCalendar(cal: any) {
  const wrap = document.getElementById('ci-calendar');
  if (!wrap) return;
  if (!cal) {
    wrap.innerHTML = `<p style="font-size:13px;color:#94a3b8;font-style:italic">Not announced by the company based on sources checked.</p>`;
    return;
  }

  const items = [
    { label: 'Previous Result', value: cal.previous_publication_date, icon: '✓', color: '#10b981', bg: '#f0fdf4', border: '#bbf7d0' },
    { label: 'Next Board Meeting', value: cal.next_meeting_date || 'Not yet announced', icon: '◷', color: '#6366f1', bg: '#eef2ff', border: '#c7d2fe' },
    { label: 'Purpose', value: cal.purpose || 'Quarterly Results', icon: '📋', color: '#f59e0b', bg: '#fffbeb', border: '#fde68a' },
  ];

  wrap.innerHTML = items.map(item => `
    <div style="display:flex;align-items:flex-start;gap:10px;padding:10px;border-radius:10px;background:${item.bg};border:1px solid ${item.border}">
      <span style="font-size:14px;margin-top:1px">${item.icon}</span>
      <div>
        <div style="font-size:11px;font-weight:700;text-transform:uppercase;letter-spacing:.05em;color:${item.color}">${item.label}</div>
        <div style="font-size:13px;color:#374151;font-weight:500;margin-top:2px">${escHtml(item.value || '—')}</div>
      </div>
    </div>`).join('');
}

// ─── News Feed ────────────────────────────────────────────
function renderNews(news: any[]) {
  const wrap = document.getElementById('ci-news');
  if (!wrap) return;
  if (!news?.length) {
    wrap.innerHTML = unavailable();
    return;
  }

  const dotColors = ['#6366f1','#10b981','#f59e0b','#ef4444','#8b5cf6'];

  wrap.innerHTML = news.map((n: any, i: number) => {
    const dateStr = n.date ? formatNewsDate(n.date) : '';
    const dotColor = dotColors[i % dotColors.length];
    const headline = escHtml(n.headline ?? '');
    const source = n.source ? `<span style="font-size:10px;font-weight:700;color:#94a3b8;margin-left:6px;text-transform:uppercase">${escHtml(n.source)}</span>` : '';
    const link = n.original_url || n.link
      ? `href="${escHtml(n.original_url || n.link)}" target="_blank" rel="noopener"`
      : '';

    return `
      <div class="ci-news-item">
        <span style="width:8px;height:8px;border-radius:50%;background:${dotColor};flex-shrink:0;margin-top:5px;display:inline-block"></span>
        <div style="flex:1;min-width:0">
          <a ${link} style="font-size:13px;font-weight:600;color:#1e293b;text-decoration:none;line-height:1.4;display:block" class="hover:text-indigo-600">
            ${headline}
          </a>
          <div style="margin-top:4px;display:flex;align-items:center;flex-wrap:wrap;gap:4px">
            ${dateStr ? `<span style="font-size:11px;color:#94a3b8">${dateStr}</span>` : ''}
            ${source}
          </div>
        </div>
      </div>`;
  }).join('');
}

// ─── Helpers ──────────────────────────────────────────────
function setText(id: string, val: string) {
  const el = document.getElementById(id);
  if (el) el.textContent = val;
}

function escHtml(s: string): string {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function unavailable(): string {
  return `<p style="font-size:13px;color:#94a3b8;font-style:italic">Currently unavailable from verified sources.</p>`;
}

function formatNewsDate(raw: string): string {
  try {
    const d = new Date(raw);
    if (isNaN(d.getTime())) return raw;
    return d.toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch { return raw; }
}
