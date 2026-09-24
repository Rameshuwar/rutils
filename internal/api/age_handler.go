package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

type AgeRequest struct {
	DOB   string `json:"dob"`   // format: YYYY-MM-DD
	Today string `json:"today"` // format: YYYY-MM-DD
}

// HandleAgeCalculate handles HTTP requests for Age calculations
// @Summary Age Calculator
// @Description Calculates exact age, next birthday, and life summary based on date of birth and a reference date.
// @Accept json
// @Produce json
// @Param request body api.AgeRequest true "Age Calculation Request (Dates in YYYY-MM-DD format)"
// @Success 200 {object} converter.AgeCalculationResponse "Age Result"
// @Failure 400 {string} string "Bad Request"
// @Router /calculate-age [post]
func HandleAgeCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.DOB == "" || req.Today == "" {
		http.Error(w, "dob and today fields are required (Format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	res, err := converter.CalculateAgeData(req.DOB, req.Today)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
