// ============================================================
// CHART-UI.TS — Chartink-style Technical Chart Controller
// ============================================================

import {
  createChart,
  CandlestickSeries,
  LineSeries,
  HistogramSeries,
  ColorType,
  IChartApi,
  ISeriesApi,
  CandlestickData,
  LineData,
  HistogramData,
  Time,
  CrosshairMode,
} from 'lightweight-charts';

import {
  fetchChartCompanies,
  fetchChartData,
  ChartCompanySummary,
  ChartDataResponse,
} from './chart-service';

import { getAuthState, isLoggedIn } from './auth';

export type ChartTheme = 'classic' | 'dark';

export class ChartController {
  private containerMain: HTMLElement | null = null;
  private containerRsi: HTMLElement | null = null;
  private resizeObserver: ResizeObserver | null = null;

  private mainChart: IChartApi | null = null;
  private rsiChart: IChartApi | null = null;

  // Series
  private candleSeries: ISeriesApi<'Candlestick'> | null = null;
  private lineSeries: ISeriesApi<'Line'> | null = null;
  private volumeSeries: ISeriesApi<'Histogram'> | null = null;
  private ema5Series: ISeriesApi<'Line'> | null = null;
  private ema13Series: ISeriesApi<'Line'> | null = null;
  private ema26Series: ISeriesApi<'Line'> | null = null;
  private sma200Series: ISeriesApi<'Line'> | null = null;
  private rsiSeries: ISeriesApi<'Line'> | null = null;

  // State
  private companies: ChartCompanySummary[] = [];
  private currentSymbol = 'HINDUNILVR';
  private currentData: ChartDataResponse | null = null;
  private currentTheme: ChartTheme = 'classic';
  private currentChartType: 'candlestick' | 'line' = 'candlestick';
  private currentPeriod: '6m' | '1y' | 'all' = '6m';

  // Visibility Flags
  private showEMA5 = true;
  private showEMA13 = true;
  private showEMA26 = true;
  private showSMA200 = true;
  private showVolume = true;
  private showRSI = true;
  private isSyncingTimeScale = false;

  // Autocomplete state
  private activeSuggestionIndex = -1;
  private isInitialized = false;

  constructor() {}

  public async init(): Promise<void> {
    if (this.isInitialized) return;

    this.containerMain = document.getElementById('chart-main-canvas');
    this.containerRsi = document.getElementById('chart-rsi-canvas');

    if (!this.containerMain || !this.containerRsi) return;

    this.bindDOMEvents();
    this.setupResizeObserver();
    this.createChartInstances();

    await this.loadCompanies();

    if (this.companies.length > 0) {
      const defaultSym =
        this.companies.find((c) => c.symbol === 'HINDUNILVR')?.symbol ||
        this.companies[0].symbol;
      await this.selectCompany(defaultSym);
    }

    this.isInitialized = true;
  }

  public onViewShown(): void {
    if (!this.isInitialized) {
      this.init();
      return;
    }

    // Force layout recalculation and fit
    setTimeout(() => {
      this.handleResize();
      if (this.currentData && this.currentData.candles.length > 0) {
        this.applyPeriodFilter();
      }
    }, 50);
  }

  public async selectCompany(symbol: string): Promise<void> {
    const sym = symbol.toUpperCase().trim();
    if (!sym) return;

    this.currentSymbol = sym;

    const symbolInput = document.getElementById('chart-symbol-input') as HTMLInputElement;
    if (symbolInput) symbolInput.value = this.currentSymbol;

    const selectEl = document.getElementById('chart-company-select') as HTMLSelectElement;
    if (selectEl) selectEl.value = this.currentSymbol;

    this.hideSuggestions();
    await this.fetchAndRender(this.currentSymbol);
  }

  private setupResizeObserver(): void {
    if (typeof ResizeObserver === 'undefined' || !this.containerMain) return;

    this.resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const width = entry.contentRect.width;
        if (width > 50) {
          this.mainChart?.applyOptions({ width });
          this.rsiChart?.applyOptions({ width });
        }
      }
    });

    this.resizeObserver.observe(this.containerMain);
  }

  private handleResize(): void {
    if (!this.containerMain || !this.mainChart) return;
    const width = this.containerMain.clientWidth;
    if (width > 50) {
      this.mainChart.applyOptions({ width });
      if (this.containerRsi && this.rsiChart) {
        this.rsiChart.applyOptions({ width });
      }
    }
  }

  private bindDOMEvents(): void {
    // Search input & button
    const searchBtn = document.getElementById('btn-chart-search');
    const symbolInput = document.getElementById('chart-symbol-input') as HTMLInputElement;
    const companySelect = document.getElementById('chart-company-select') as HTMLSelectElement;
    const searchDropdown = document.getElementById('chart-search-results');

    searchBtn?.addEventListener('click', () => {
      if (symbolInput?.value) this.selectCompany(symbolInput.value);
    });

    symbolInput?.addEventListener('input', () => {
      const query = symbolInput.value.trim().toUpperCase();
      this.renderSuggestions(query);
    });

    symbolInput?.addEventListener('focus', () => {
      const query = symbolInput.value.trim().toUpperCase();
      this.renderSuggestions(query);
    });

    symbolInput?.addEventListener('keydown', (e) => {
      const items = searchDropdown?.querySelectorAll('.chart-search-item');
      if (!items || items.length === 0 || searchDropdown?.classList.contains('hidden')) {
        if (e.key === 'Enter') {
          e.preventDefault();
          if (symbolInput.value) this.selectCompany(symbolInput.value);
        }
        return;
      }

      if (e.key === 'ArrowDown') {
        e.preventDefault();
        this.activeSuggestionIndex = (this.activeSuggestionIndex + 1) % items.length;
        this.highlightSuggestion(items);
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        this.activeSuggestionIndex =
          (this.activeSuggestionIndex - 1 + items.length) % items.length;
        this.highlightSuggestion(items);
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (this.activeSuggestionIndex >= 0 && this.activeSuggestionIndex < items.length) {
          const sym = (items[this.activeSuggestionIndex] as HTMLElement).dataset.symbol;
          if (sym) this.selectCompany(sym);
        } else if (symbolInput.value) {
          this.selectCompany(symbolInput.value);
        }
      } else if (e.key === 'Escape') {
        this.hideSuggestions();
      }
    });

    // Close suggestions on outside click
    document.addEventListener('click', (e) => {
      const target = e.target as HTMLElement;
      if (!target.closest('#chart-symbol-input') && !target.closest('#chart-search-results')) {
        this.hideSuggestions();
      }
    });

    companySelect?.addEventListener('change', () => {
      if (companySelect.value) this.selectCompany(companySelect.value);
    });

    // Update Chart button
    const updateBtn = document.getElementById('btn-chart-update');
    updateBtn?.addEventListener('click', () => {
      this.fetchAndRender(this.currentSymbol);
    });

    // Period selector
    const periodSelect = document.getElementById('chart-period-select') as HTMLSelectElement;
    periodSelect?.addEventListener('change', () => {
      this.currentPeriod = (periodSelect.value as '6m' | '1y' | 'all') || '6m';
      this.applyPeriodFilter();
    });

    // Chart Type selector
    const typeSelect = document.getElementById('chart-type-select') as HTMLSelectElement;
    typeSelect?.addEventListener('change', () => {
      const val = typeSelect.value.toLowerCase();
      this.currentChartType = val.includes('line') ? 'line' : 'candlestick';
      this.toggleChartType();
    });

    // Theme selector
    const themeSelect = document.getElementById('chart-theme-select') as HTMLSelectElement;
    themeSelect?.addEventListener('change', () => {
      this.currentTheme = (themeSelect.value as ChartTheme) || 'classic';
      this.applyTheme(this.currentTheme);
    });

    // Indicator Checkboxes
    const chkEMA5 = document.getElementById('chk-ema-5') as HTMLInputElement;
    const chkEMA13 = document.getElementById('chk-ema-13') as HTMLInputElement;
    const chkEMA26 = document.getElementById('chk-ema-26') as HTMLInputElement;
    const chkSMA200 = document.getElementById('chk-sma-200') as HTMLInputElement;
    const chkVolume = document.getElementById('chk-volume') as HTMLInputElement;
    const chkRSI = document.getElementById('chk-rsi') as HTMLInputElement;

    chkEMA5?.addEventListener('change', () => {
      this.showEMA5 = chkEMA5.checked;
      this.ema5Series?.applyOptions({ visible: this.showEMA5 });
    });

    chkEMA13?.addEventListener('change', () => {
      this.showEMA13 = chkEMA13.checked;
      this.ema13Series?.applyOptions({ visible: this.showEMA13 });
    });

    chkEMA26?.addEventListener('change', () => {
      this.showEMA26 = chkEMA26.checked;
      this.ema26Series?.applyOptions({ visible: this.showEMA26 });
    });

    chkSMA200?.addEventListener('change', () => {
      this.showSMA200 = chkSMA200.checked;
      this.sma200Series?.applyOptions({ visible: this.showSMA200 });
    });

    chkVolume?.addEventListener('change', () => {
      this.showVolume = chkVolume.checked;
      this.volumeSeries?.applyOptions({ visible: this.showVolume });
    });

    chkRSI?.addEventListener('change', () => {
      this.showRSI = chkRSI.checked;
      const rsiCard = document.getElementById('chart-rsi-wrapper');
      if (rsiCard) {
        if (this.showRSI) {
          rsiCard.classList.remove('hidden');
          setTimeout(() => this.handleResize(), 50);
        } else {
          rsiCard.classList.add('hidden');
        }
      }
    });

    // Save Image button
    const saveImgBtn = document.getElementById('btn-chart-save-image');
    saveImgBtn?.addEventListener('click', () => this.exportImage());

    // Window resize fallback
    window.addEventListener('resize', () => this.handleResize());
  }

  private renderSuggestions(query: string): void {
    const searchDropdown = document.getElementById('chart-search-results');
    if (!searchDropdown) return;

    this.activeSuggestionIndex = -1;

    let matches = this.companies;
    if (query) {
      matches = this.companies.filter((c) => {
        const sym = c.symbol.toUpperCase();
        const name = (c.displayName || '').toUpperCase();
        const ind = (c.industry || '').toUpperCase();
        return sym.includes(query) || name.includes(query) || ind.includes(query);
      });
    }

    if (matches.length === 0) {
      searchDropdown.innerHTML = `<div class="p-3 text-xs text-slate-400 text-center">No companies found for "${query}"</div>`;
      searchDropdown.classList.remove('hidden');
      return;
    }

    const isClassic = this.currentTheme === 'classic';
    const topMatches = matches.slice(0, 15);

    const html = topMatches
      .map((c) => {
        const badgeClass = isClassic
          ? 'bg-teal-100 text-teal-800 border-teal-300'
          : 'bg-teal-900 text-teal-200 border-teal-700';

        return `
        <div class="chart-search-item" data-symbol="${c.symbol}">
          <div class="flex items-center gap-2 overflow-hidden">
            <span class="px-1.5 py-0.5 rounded font-mono font-bold text-xs border ${badgeClass}">
              ${c.symbol}
            </span>
            <span class="font-medium truncate text-xs ${isClassic ? 'text-slate-900' : 'text-slate-100'}">
              ${c.displayName || c.symbol}
            </span>
          </div>
          ${c.industry ? `<span class="text-[10px] text-slate-400 whitespace-nowrap">${c.industry}</span>` : ''}
        </div>
      `;
      })
      .join('');

    searchDropdown.innerHTML = html;
    searchDropdown.classList.remove('hidden');

    // Attach click listeners to items
    searchDropdown.querySelectorAll('.chart-search-item').forEach((item) => {
      item.addEventListener('click', () => {
        const sym = (item as HTMLElement).dataset.symbol;
        if (sym) this.selectCompany(sym);
      });
    });
  }

  private highlightSuggestion(items: NodeListOf<Element>): void {
    items.forEach((item, idx) => {
      if (idx === this.activeSuggestionIndex) {
        item.classList.add('active');
        item.scrollIntoView({ block: 'nearest' });
      } else {
        item.classList.remove('active');
      }
    });
  }

  private hideSuggestions(): void {
    const searchDropdown = document.getElementById('chart-search-results');
    searchDropdown?.classList.add('hidden');
    this.activeSuggestionIndex = -1;
  }

  private async loadCompanies(): Promise<void> {
    const { token } = getAuthState();

    try {
      const res = await fetchChartCompanies(token || undefined);
      this.companies = res.companies || [];

      const selectEl = document.getElementById('chart-company-select') as HTMLSelectElement;

      if (selectEl) {
        selectEl.innerHTML = '';
        this.companies.forEach((c) => {
          const opt = document.createElement('option');
          opt.value = c.symbol;
          opt.textContent = `${c.symbol} - ${c.displayName || c.symbol}`;
          selectEl.appendChild(opt);
        });
        if (this.currentSymbol) {
          selectEl.value = this.currentSymbol;
        }
      }
    } catch (err) {
      console.error('Failed to load chart companies:', err);
    }
  }

  private async fetchAndRender(symbol: string): Promise<void> {
    const loadingCard = document.getElementById('chart-loading-indicator');
    const errorCard = document.getElementById('chart-error-banner');
    const contentCard = document.getElementById('chart-content-area');

    loadingCard?.classList.remove('hidden');
    errorCard?.classList.add('hidden');

    const { token } = getAuthState();

    try {
      const data = await fetchChartData(symbol, token || undefined);
      this.currentData = data;

      this.renderCharts(data);

      loadingCard?.classList.add('hidden');
      contentCard?.classList.remove('hidden');
    } catch (err) {
      console.error('Failed to fetch chart data:', err);
      loadingCard?.classList.add('hidden');
      if (errorCard) {
        errorCard.textContent = `Error loading chart for ${symbol}: ${(err as Error).message}`;
        errorCard.classList.remove('hidden');
      }
    }
  }

  private createChartInstances(): void {
    if (!this.containerMain || !this.containerRsi) return;

    // Clean up previous instances if any
    if (this.mainChart) {
      this.mainChart.remove();
      this.mainChart = null;
    }
    if (this.rsiChart) {
      this.rsiChart.remove();
      this.rsiChart = null;
    }

    this.containerMain.innerHTML = '';
    this.containerRsi.innerHTML = '';

    const isClassic = this.currentTheme === 'classic';

    const bgColors = {
      bg: isClassic ? '#ffffff' : '#0f172a',
      text: isClassic ? '#1a202c' : '#cbd5e1',
      grid: isClassic ? '#edf2f7' : '#1e293b',
      border: isClassic ? '#cbd5e0' : '#334155',
    };

    const initialWidth = this.containerMain.clientWidth || 900;

    // 1. Create Main Chart
    this.mainChart = createChart(this.containerMain, {
      width: initialWidth,
      height: 480,
      layout: {
        background: { type: ColorType.Solid, color: bgColors.bg },
        textColor: bgColors.text,
      },
      grid: {
        vertLines: { color: bgColors.grid },
        horzLines: { color: bgColors.grid },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
      },
      rightPriceScale: {
        borderColor: bgColors.border,
        scaleMargins: {
          top: 0.08,
          bottom: 0.22, // space for volume overlay
        },
      },
      timeScale: {
        borderColor: bgColors.border,
        timeVisible: true,
        secondsVisible: false,
      },
    });

    // 2. Add Candlestick Series
    this.candleSeries = this.mainChart.addSeries(CandlestickSeries, {
      upColor: isClassic ? '#ffffff' : '#26a69a',
      downColor: '#ef5350',
      borderVisible: true,
      borderUpColor: isClassic ? '#2b6cb0' : '#26a69a',
      borderDownColor: '#ef5350',
      wickUpColor: isClassic ? '#2b6cb0' : '#26a69a',
      wickDownColor: '#ef5350',
    });

    // 3. Add Line Series (alternative to Candlesticks)
    this.lineSeries = this.mainChart.addSeries(LineSeries, {
      color: '#3182ce',
      lineWidth: 2,
      visible: false,
    });

    // 4. Add Volume Histogram (bottom overlay with separate price scale)
    this.volumeSeries = this.mainChart.addSeries(HistogramSeries, {
      priceFormat: {
        type: 'volume',
      },
      priceScaleId: 'volume',
    });

    this.mainChart.priceScale('volume').applyOptions({
      scaleMargins: {
        top: 0.78,
        bottom: 0,
      },
    });

    // 5. Add Moving Averages on Main Chart
    // EMA (5) - Purple (#805ad5)
    this.ema5Series = this.mainChart.addSeries(LineSeries, {
      color: '#9f7aea',
      lineWidth: 2,
      title: 'EMA 5',
    });

    // EMA (13) - Green (#38a169)
    this.ema13Series = this.mainChart.addSeries(LineSeries, {
      color: '#38a169',
      lineWidth: 2,
      title: 'EMA 13',
    });

    // EMA (26) - Black / Silver
    this.ema26Series = this.mainChart.addSeries(LineSeries, {
      color: isClassic ? '#1a202c' : '#cbd5e1',
      lineWidth: 2,
      title: 'EMA 26',
    });

    // SMA (200) - Red / Maroon (#e53e3e)
    this.sma200Series = this.mainChart.addSeries(LineSeries, {
      color: '#e53e3e',
      lineWidth: 2,
      title: 'SMA 200',
    });

    // 6. Create RSI Sub-Chart
    this.rsiChart = createChart(this.containerRsi, {
      width: initialWidth,
      height: 140,
      layout: {
        background: { type: ColorType.Solid, color: bgColors.bg },
        textColor: bgColors.text,
      },
      grid: {
        vertLines: { color: bgColors.grid },
        horzLines: { color: bgColors.grid },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
      },
      rightPriceScale: {
        borderColor: bgColors.border,
        scaleMargins: {
          top: 0.1,
          bottom: 0.1,
        },
      },
      timeScale: {
        borderColor: bgColors.border,
        timeVisible: true,
      },
    });

    // 7. Add RSI Line Series
    this.rsiSeries = this.rsiChart.addSeries(LineSeries, {
      color: '#3182ce',
      lineWidth: 2,
      title: 'RSI 14',
    });

    // Add 70 / 30 reference lines on RSI chart
    this.rsiSeries.createPriceLine({
      price: 70,
      color: '#e53e3e',
      lineWidth: 1,
      lineStyle: 2, // Dashed
      axisLabelVisible: true,
      title: '70 (Overbought)',
    });

    this.rsiSeries.createPriceLine({
      price: 30,
      color: '#38a169',
      lineWidth: 1,
      lineStyle: 2, // Dashed
      axisLabelVisible: true,
      title: '30 (Oversold)',
    });

    // 8. Synchronize Timescales between Main and RSI charts
    this.mainChart.timeScale().subscribeVisibleLogicalRangeChange((range) => {
      if (this.isSyncingTimeScale || !range || !this.rsiChart) return;
      this.isSyncingTimeScale = true;
      this.rsiChart.timeScale().setVisibleLogicalRange(range);
      this.isSyncingTimeScale = false;
    });

    this.rsiChart.timeScale().subscribeVisibleLogicalRangeChange((range) => {
      if (this.isSyncingTimeScale || !range || !this.mainChart) return;
      this.isSyncingTimeScale = true;
      this.mainChart.timeScale().setVisibleLogicalRange(range);
      this.isSyncingTimeScale = false;
    });

    // 9. Subscribe to Crosshair Moves to update top status stats
    this.mainChart.subscribeCrosshairMove((param) => {
      this.updateHeaderStats(param);
    });
  }

  private renderCharts(data: ChartDataResponse): void {
    if (!this.mainChart || !this.rsiChart) {
      this.createChartInstances();
    }

    if (!data.candles || data.candles.length === 0) return;

    // 1. Prepare Candlestick & Volume data (deduplicated & sorted)
    const candleData: CandlestickData<Time>[] = [];
    const lineData: LineData<Time>[] = [];
    const volumeData: HistogramData<Time>[] = [];
    const seenTimes = new Set<string>();

    data.candles.forEach((c) => {
      const timeStr = c.date.substring(0, 10);
      if (seenTimes.has(timeStr)) return;
      seenTimes.add(timeStr);

      const timeVal = timeStr as Time;
      const isUp = c.close >= c.open;

      candleData.push({
        time: timeVal,
        open: c.open,
        high: c.high,
        low: c.low,
        close: c.close,
      });

      lineData.push({
        time: timeVal,
        value: c.close,
      });

      volumeData.push({
        time: timeVal,
        value: c.volume,
        color: isUp ? 'rgba(56, 161, 105, 0.65)' : 'rgba(239, 83, 80, 0.65)',
      });
    });

    this.candleSeries?.setData(candleData);
    this.lineSeries?.setData(lineData);
    this.volumeSeries?.setData(volumeData);

    // 2. Prepare Indicator series
    if (data.indicators && data.indicators.length > 0) {
      const ema5Data: LineData<Time>[] = [];
      const ema13Data: LineData<Time>[] = [];
      const ema26Data: LineData<Time>[] = [];
      const sma200Data: LineData<Time>[] = [];
      const rsiData: LineData<Time>[] = [];
      const seenIndTimes = new Set<string>();

      data.indicators.forEach((ind) => {
        const timeStr = ind.date.substring(0, 10);
        if (seenIndTimes.has(timeStr)) return;
        seenIndTimes.add(timeStr);

        const timeVal = timeStr as Time;
        if (ind.ema5 > 0) ema5Data.push({ time: timeVal, value: ind.ema5 });
        if (ind.ema13 > 0) ema13Data.push({ time: timeVal, value: ind.ema13 });
        if (ind.ema26 > 0) ema26Data.push({ time: timeVal, value: ind.ema26 });
        if (ind.sma200 > 0) sma200Data.push({ time: timeVal, value: ind.sma200 });
        if (ind.rsi >= 0) rsiData.push({ time: timeVal, value: ind.rsi });
      });

      this.ema5Series?.setData(ema5Data);
      this.ema13Series?.setData(ema13Data);
      this.ema26Series?.setData(ema26Data);
      this.sma200Series?.setData(sma200Data);
      this.rsiSeries?.setData(rsiData);
    }

    // 3. Render latest advice banner
    this.renderAdviceBanner(data);

    // 4. Default header metrics to the last candle
    const lastCandle = data.candles[data.candles.length - 1];
    const prevCandle = data.candles.length > 1 ? data.candles[data.candles.length - 2] : null;
    this.setHeaderValues(lastCandle, prevCandle);

    // 5. Fit content to view & apply period
    this.applyPeriodFilter();
  }

  private applyPeriodFilter(): void {
    if (!this.mainChart || !this.currentData || this.currentData.candles.length === 0) return;

    const totalCandles = this.currentData.candles.length;
    let rangeCount = totalCandles;

    if (this.currentPeriod === '6m') {
      rangeCount = Math.min(126, totalCandles);
    } else if (this.currentPeriod === '1y') {
      rangeCount = Math.min(252, totalCandles);
    }

    this.mainChart.timeScale().setVisibleLogicalRange({
      from: totalCandles - rangeCount,
      to: totalCandles - 1,
    });
  }

  private toggleChartType(): void {
    if (this.currentChartType === 'line') {
      this.candleSeries?.applyOptions({ visible: false });
      this.lineSeries?.applyOptions({ visible: true });
    } else {
      this.candleSeries?.applyOptions({ visible: true });
      this.lineSeries?.applyOptions({ visible: false });
    }
  }

  public applyTheme(theme: ChartTheme): void {
    this.currentTheme = theme;
    const isClassic = theme === 'classic';

    const bgColors = {
      bg: isClassic ? '#ffffff' : '#0f172a',
      text: isClassic ? '#1a202c' : '#cbd5e1',
      grid: isClassic ? '#edf2f7' : '#1e293b',
      border: isClassic ? '#cbd5e0' : '#334155',
    };

    const chartOptions = {
      layout: {
        background: { type: ColorType.Solid, color: bgColors.bg },
        textColor: bgColors.text,
      },
      grid: {
        vertLines: { color: bgColors.grid },
        horzLines: { color: bgColors.grid },
      },
      rightPriceScale: { borderColor: bgColors.border },
      timeScale: { borderColor: bgColors.border },
    };

    this.mainChart?.applyOptions(chartOptions);
    this.rsiChart?.applyOptions(chartOptions);

    this.candleSeries?.applyOptions({
      upColor: isClassic ? '#ffffff' : '#26a69a',
      borderUpColor: isClassic ? '#2b6cb0' : '#26a69a',
      wickUpColor: isClassic ? '#2b6cb0' : '#26a69a',
    });

    this.ema26Series?.applyOptions({
      color: isClassic ? '#1a202c' : '#cbd5e1',
    });

    // Container theme classes
    const wrapper = document.getElementById('chartink-wrapper');
    if (wrapper) {
      if (isClassic) {
        wrapper.classList.remove('chartink-theme-dark');
        wrapper.classList.add('chartink-theme-classic');
      } else {
        wrapper.classList.remove('chartink-theme-classic');
        wrapper.classList.add('chartink-theme-dark');
      }
    }
  }

  private updateHeaderStats(param: any): void {
    if (!param.time || !this.currentData) return;

    const timeStr = typeof param.time === 'string' ? param.time : '';
    const candle = this.currentData.candles.find((c) => c.date.startsWith(timeStr));
    if (!candle) return;

    const candleIdx = this.currentData.candles.indexOf(candle);
    const prevCandle = candleIdx > 0 ? this.currentData.candles[candleIdx - 1] : null;

    this.setHeaderValues(candle, prevCandle);

    // Also look up indicators for that date
    const ind = this.currentData.indicators.find((i) => i.date.startsWith(timeStr));
    if (ind) {
      this.setIndicatorStats(ind);
    }
  }

  private setHeaderValues(candle: any, prevCandle: any | null): void {
    const elDt = document.getElementById('stat-dt');
    const elOp = document.getElementById('stat-op');
    const elHi = document.getElementById('stat-hi');
    const elLw = document.getElementById('stat-lw');
    const elCl = document.getElementById('stat-cl');
    const elCh = document.getElementById('stat-ch');
    const elVl = document.getElementById('stat-vl');
    const elBadge = document.getElementById('stat-symbol-badge');

    if (!candle) return;

    if (elDt) elDt.textContent = candle.date.substring(0, 10);
    if (elOp) elOp.textContent = candle.open.toFixed(2);
    if (elHi) elHi.textContent = candle.high.toFixed(2);
    if (elLw) elLw.textContent = candle.low.toFixed(2);
    if (elCl) elCl.textContent = candle.close.toFixed(2);

    let chgText = '0.00%';
    let isPositive = true;

    if (prevCandle && prevCandle.close > 0) {
      const diff = candle.close - prevCandle.close;
      const pct = (diff / prevCandle.close) * 100;
      isPositive = pct >= 0;
      chgText = `${isPositive ? '+' : ''}${pct.toFixed(2)}%`;
    }

    if (elCh) {
      elCh.textContent = chgText;
      elCh.className = isPositive ? 'text-emerald-500 font-bold' : 'text-rose-500 font-bold';
    }

    if (elVl) {
      const v = candle.volume;
      elVl.textContent =
        v >= 1e7
          ? `${(v / 1e7).toFixed(2)} Cr`
          : v >= 1e5
          ? `${(v / 1e5).toFixed(2)} L`
          : v.toLocaleString();
    }

    if (elBadge) {
      elBadge.textContent = `${this.currentSymbol} ₹${candle.close.toFixed(2)}`;
    }
  }

  private setIndicatorStats(ind: any): void {
    const elEma5 = document.getElementById('stat-ema5');
    const elEma13 = document.getElementById('stat-ema13');
    const elEma26 = document.getElementById('stat-ema26');
    const elSma200 = document.getElementById('stat-sma200');
    const elRsi = document.getElementById('stat-rsi');

    if (elEma5 && ind.ema5) elEma5.textContent = ind.ema5.toFixed(2);
    if (elEma13 && ind.ema13) elEma13.textContent = ind.ema13.toFixed(2);
    if (elEma26 && ind.ema26) elEma26.textContent = ind.ema26.toFixed(2);
    if (elSma200 && ind.sma200) elSma200.textContent = ind.sma200.toFixed(2);
    if (elRsi && ind.rsi !== undefined) elRsi.textContent = ind.rsi.toFixed(2);
  }

  private renderAdviceBanner(data: ChartDataResponse): void {
    const adviceCard = document.getElementById('chart-advice-banner');
    if (!adviceCard) return;

    if (!data.advice || data.advice.length === 0) {
      adviceCard.classList.add('hidden');
      return;
    }

    const latest = data.advice[data.advice.length - 1];
    adviceCard.classList.remove('hidden');

    const badge = document.getElementById('advice-badge');
    const text = document.getElementById('advice-details');

    if (badge) {
      badge.textContent = latest.advice;
      badge.className =
        latest.advice === 'BUY'
          ? 'px-3 py-1 bg-emerald-600 text-white font-bold rounded-lg uppercase tracking-wider text-xs'
          : 'px-3 py-1 bg-rose-600 text-white font-bold rounded-lg uppercase tracking-wider text-xs';
    }

    if (text) {
      text.textContent = `Target: ₹${latest.targetPrice.toFixed(2)} | 200-SMA Support: ₹${latest.sma200Support.toFixed(2)} | Date: ${latest.date.substring(0, 10)}`;
    }
  }

  private exportImage(): void {
    if (!this.containerMain) return;

    const mainCanvases = this.containerMain.querySelectorAll('canvas');
    if (mainCanvases.length === 0) return;

    // Merge main chart and RSI chart into a single export canvas
    const rsiCanvases = this.containerRsi ? this.containerRsi.querySelectorAll('canvas') : null;

    const mainCv = mainCanvases[0];
    const rsiCv = rsiCanvases && rsiCanvases.length > 0 ? rsiCanvases[0] : null;

    const exportCv = document.createElement('canvas');
    exportCv.width = mainCv.width;
    exportCv.height = mainCv.height + (rsiCv && this.showRSI ? rsiCv.height : 0);

    const ctx = exportCv.getContext('2d');
    if (!ctx) return;

    // Fill background
    ctx.fillStyle = this.currentTheme === 'classic' ? '#ffffff' : '#0f172a';
    ctx.fillRect(0, 0, exportCv.width, exportCv.height);

    // Draw main chart
    ctx.drawImage(mainCv, 0, 0);

    // Draw RSI if visible
    if (rsiCv && this.showRSI) {
      ctx.drawImage(rsiCv, 0, mainCv.height);
    }

    const dataUrl = exportCv.toDataURL('image/png');
    const a = document.createElement('a');
    a.href = dataUrl;
    a.download = `inkchart-${this.currentSymbol}-${new Date().toISOString().substring(0, 10)}.png`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }
}

// Global instance export
export const chartController = new ChartController();
