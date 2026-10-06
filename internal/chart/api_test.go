package chart_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"file-converter/internal/api"
	"file-converter/internal/chart"
)

func TestChartHandler_HTTP_Endpoints(t *testing.T) {
	svc := chart.NewService("../../data/nse/data", nil)
	handler := api.NewChartHandler(svc)

	// 1. Test GET /market/chart/companies
	reqCompanies := httptest.NewRequest(http.MethodGet, "/market/chart/companies", nil)
	wCompanies := httptest.NewRecorder()
	handler.GetCompanies(wCompanies, reqCompanies)

	if wCompanies.Code != http.StatusOK {
		t.Fatalf("expected status 200 for GetCompanies, got %d: %s", wCompanies.Code, wCompanies.Body.String())
	}

	var compResp chart.CompaniesResponse
	if err := json.Unmarshal(wCompanies.Body.Bytes(), &compResp); err != nil {
		t.Fatalf("failed to parse companies response: %v", err)
	}

	if !compResp.Success || compResp.Count < 50 {
		t.Errorf("expected at least 50 companies, got %d", compResp.Count)
	}

	// 2. Test GET /market/chart/data?symbol=HINDUNILVR
	reqData := httptest.NewRequest(http.MethodGet, "/market/chart/data?symbol=HINDUNILVR", nil)
	wData := httptest.NewRecorder()
	handler.GetChartData(wData, reqData)

	if wData.Code != http.StatusOK {
		t.Fatalf("expected status 200 for GetChartData, got %d: %s", wData.Code, wData.Body.String())
	}

	var chartData chart.ChartDataResponse
	if err := json.Unmarshal(wData.Body.Bytes(), &chartData); err != nil {
		t.Fatalf("failed to parse chart data response: %v", err)
	}

	if !chartData.Success || chartData.Symbol != "HINDUNILVR" {
		t.Errorf("expected success for HINDUNILVR, got symbol: %s", chartData.Symbol)
	}

	if len(chartData.Candles) == 0 {
		t.Errorf("expected candles to be populated")
	}

	if len(chartData.Indicators) == 0 {
		t.Errorf("expected indicators to be populated")
	}

	// 3. Test Missing Symbol parameter -> 400 Bad Request
	reqMissing := httptest.NewRequest(http.MethodGet, "/market/chart/data", nil)
	wMissing := httptest.NewRecorder()
	handler.GetChartData(wMissing, reqMissing)
	if wMissing.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing symbol, got %d", wMissing.Code)
	}
}
