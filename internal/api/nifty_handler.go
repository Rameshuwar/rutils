package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"file-converter/internal/nifty"
)

// NiftyHandler handles HTTP requests for NIFTY 50 market data
type NiftyHandler struct {
	service *nifty.Service
}

// NewNiftyHandler creates a new NiftyHandler
func NewNiftyHandler(service *nifty.Service) *NiftyHandler {
	return &NiftyHandler{
		service: service,
	}
}

// GetCompanies handles GET /nifty50/companies
// @Summary Get NIFTY 50 Constituent Companies
// @Description Returns the latest official 50 NIFTY 50 constituent companies. Requires JWT Authentication.
// @Tags markets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} nifty.CompaniesResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 503 {object} map[string]interface{} "NIFTY 50 data is currently unavailable"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /nifty50/companies [get]
func (h *NiftyHandler) GetCompanies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Only GET method is allowed",
		})
		return
	}

	data, err := h.service.GetCompanies()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, nifty.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "NIFTY 50 data is currently unavailable",
			})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to load NIFTY 50 constituent data",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(nifty.CompaniesResponse{
		Success:   true,
		UpdatedAt: data.UpdatedAt,
		Count:     data.Count,
		Companies: data.Companies,
	})
}
