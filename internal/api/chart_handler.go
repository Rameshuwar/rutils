package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"file-converter/internal/chart"
)

// ChartHandler exposes HTTP endpoints for interactive technical charts
type ChartHandler struct {
	service *chart.Service
}

// NewChartHandler creates a new ChartHandler
func NewChartHandler(service *chart.Service) *ChartHandler {
	return &ChartHandler{
		service: service,
	}
}

// GetCompanies handles GET /market/chart/companies or GET /api/market/chart/companies
func (h *ChartHandler) GetCompanies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Only GET method is allowed",
		})
		return
	}

	companies, err := h.service.GetCompanies()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp := chart.CompaniesResponse{
		Success:   true,
		Count:     len(companies),
		Companies: companies,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetChartData handles GET /market/chart/data?symbol=HINDUNILVR
func (h *ChartHandler) GetChartData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Only GET method is allowed",
		})
		return
	}

	symbol := strings.TrimSpace(r.URL.Query().Get("symbol"))
	if symbol == "" {
		// Fallback: check if symbol was passed in URL path, e.g. /data/HINDUNILVR
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) > 0 {
			last := pathParts[len(pathParts)-1]
			if !strings.EqualFold(last, "data") {
				symbol = last
			}
		}
	}

	if symbol == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Query parameter 'symbol' is required (e.g. ?symbol=HINDUNILVR)",
		})
		return
	}

	data, err := h.service.GetChartData(symbol)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, chart.ErrCompanyNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Company not found: " + symbol,
			})
			return
		}
		if errors.Is(err, chart.ErrInvalidSymbol) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid company symbol: " + symbol,
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}
