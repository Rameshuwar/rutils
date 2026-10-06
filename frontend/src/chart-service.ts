// ============================================================
// CHART-SERVICE.TS — Technical Chart API client & types
// ============================================================

import { apiGet } from './auth';

export interface ChartCompanySummary {
  symbol: string;
  displayName: string;
  industry?: string;
  symbolToken?: string;
  exchange: string;
  interval: string;
  candleCount: number;
  adviceCount: number;
  latestCandleDate?: string;
}

export interface ChartCandle {
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface ChartIndicator {
  date: string;
  close: number;
  ema5: number;
  ema13: number;
  ema26: number;
  sma200: number;
  rsi: number;
}

export interface ChartAdvice {
  niftyIdentifier: string;
  index: number;
  advice: 'BUY' | 'SELL';
  targetPrice: number;
  sma200Support: number;
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface ChartDataResponse {
  success: boolean;
  symbol: string;
  metadata: {
    companyName: string;
    symbolToken?: string;
    exchange?: string;
    interval?: string;
    candleCount?: number;
    latestCandleDate?: string;
  };
  candles: ChartCandle[];
  indicators: ChartIndicator[];
  advice: ChartAdvice[];
}

export interface ChartCompaniesResponse {
  success: boolean;
  count: number;
  companies: ChartCompanySummary[];
}

export async function fetchChartCompanies(token?: string): Promise<ChartCompaniesResponse> {
  return apiGet<ChartCompaniesResponse>('/market/chart/companies', token);
}

export async function fetchChartData(symbol: string, token?: string): Promise<ChartDataResponse> {
  return apiGet<ChartDataResponse>(`/market/chart/data?symbol=${encodeURIComponent(symbol)}`, token);
}

