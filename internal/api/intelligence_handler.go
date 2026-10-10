package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"file-converter/internal/intelligence"
)

type IntelligenceHandler struct {
	service *intelligence.Service
}

func NewIntelligenceHandler(service *intelligence.Service) *IntelligenceHandler {
	return &IntelligenceHandler{
		service: service,
	}
}

func (h *IntelligenceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Only GET method is allowed",
		})
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	
	symbol := parts[2]
	endpoint := parts[3]

	data, err := h.service.GetIntelligence(symbol)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Intelligence data not found for symbol",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	switch endpoint {
	case "intelligence":
		_ = json.NewEncoder(w).Encode(data)
	case "announcements":
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"announcements": data.Announcements,
		})
	case "financials":
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"quarterly_results": data.QuarterlyResults,
			"annual_results":    data.AnnualResults,
			"financial_metrics": data.FinancialMetrics,
		})
	case "reports":
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"reports": data.Reports,
		})
	case "news":
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"news": data.News,
		})
	default:
		http.NotFound(w, r)
	}
}
