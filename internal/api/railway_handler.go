package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// HandleRailwayConvert handles HTTP requests for 12hr <-> 24hr railway time conversions
// @Summary Railway Time Converter
// @Description Converts between 12-hour AM/PM and 24-hour Indian Railway time format.
// @Accept json
// @Produce json
// @Param request body converter.RailwayConversionRequest true "Railway Conversion Request"
// @Success 200 {object} converter.RailwayConversionResponse "Conversion Result"
// @Failure 400 {string} string "Bad Request"
// @Router /convert-railway [post]
func HandleRailwayConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req converter.RailwayConversionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.ConvertRailwayTime(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
