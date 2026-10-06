package chart

import (
	"testing"
)

func TestChartService_GetCompanies(t *testing.T) {
	svc := NewService("../../data/nse/data", nil)
	companies, err := svc.GetCompanies()
	if err != nil {
		t.Fatalf("expected no error getting companies, got: %v", err)
	}

	if len(companies) == 0 {
		t.Fatalf("expected at least 1 company, got 0")
	}

	// Verify HINDUNILVR exists
	var found bool
	for _, c := range companies {
		if c.Symbol == "HINDUNILVR" {
			found = true
			if c.CandleCount == 0 {
				t.Errorf("expected positive candle count for HINDUNILVR, got %d", c.CandleCount)
			}
			break
		}
	}

	if !found {
		t.Errorf("expected to find HINDUNILVR in companies list")
	}
}

func TestChartService_GetChartData_Valid(t *testing.T) {
	svc := NewService("../../data/nse/data", nil)
	data, err := svc.GetChartData("HINDUNILVR")
	if err != nil {
		t.Fatalf("expected no error getting chart data for HINDUNILVR, got: %v", err)
	}

	if !data.Success {
		t.Errorf("expected success to be true")
	}
	if data.Symbol != "HINDUNILVR" {
		t.Errorf("expected symbol HINDUNILVR, got %s", data.Symbol)
	}
	if len(data.Candles) == 0 {
		t.Errorf("expected candles to be non-empty")
	}
	if len(data.Indicators) == 0 {
		t.Errorf("expected indicators to be non-empty")
	}

	// Verify sample candle
	firstCandle := data.Candles[0]
	if firstCandle.Open == 0 || firstCandle.Close == 0 {
		t.Errorf("invalid candle prices: %+v", firstCandle)
	}
}

func TestChartService_GetChartData_InvalidSymbol(t *testing.T) {
	svc := NewService("../../data/nse/data", nil)
	_, err := svc.GetChartData("NON_EXISTENT_COMPANY_XYZ")
	if err != ErrCompanyNotFound {
		t.Errorf("expected ErrCompanyNotFound, got: %v", err)
	}

	_, err = svc.GetChartData("../bad/symbol")
	if err != ErrInvalidSymbol {
		t.Errorf("expected ErrInvalidSymbol, got: %v", err)
	}
}
