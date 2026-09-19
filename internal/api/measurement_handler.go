package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// MeasurementRequest represents the incoming JSON request for measurement conversion
type MeasurementRequest struct {
	Category string  `json:"category"`
	FromUnit string  `json:"fromUnit"`
	ToUnit   string  `json:"toUnit"`
	Value    float64 `json:"value"`
}

// MeasurementResponse represents the successful JSON response
type MeasurementResponse struct {
	Result float64 `json:"result"`
}

// HandleMeasurementConvert handles HTTP requests for measurement conversions
// @Summary Measurement Converter
// @Description Converts measurements between different units (e.g., meters to feet)
// @Accept json
// @Produce json
// @Param request body api.MeasurementRequest true "Conversion Request"
// @Success 200 {object} api.MeasurementResponse "Conversion Result"
// @Failure 400 {string} string "Bad Request"
// @Router /convert-measurement [post]
func HandleMeasurementConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MeasurementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Category == "" || req.FromUnit == "" || req.ToUnit == "" {
		http.Error(w, "category, fromUnit, and toUnit are required", http.StatusBadRequest)
		return
	}

	result, err := converter.ConvertMeasurement(req.Category, req.FromUnit, req.ToUnit, req.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := MeasurementResponse{Result: result}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
