package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// HandleTimeConvert handles HTTP requests for universal time conversions
// @Summary Universal Time Converter
// @Description Converts date and time from any location/time zone to any other location/time zone accurately.
// @Accept json
// @Produce json
// @Param request body converter.TimeConversionRequest true "Time Conversion Request"
// @Success 200 {object} converter.TimeConversionResponse "Conversion Result"
// @Failure 400 {string} string "Bad Request"
// @Router /convert-time [post]
func HandleTimeConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req converter.TimeConversionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.ConvertTime(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
