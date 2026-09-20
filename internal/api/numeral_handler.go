package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// NumeralRequest represents the incoming JSON request for numeral conversions
type NumeralRequest struct {
	FromBase string `json:"fromBase"`
	ToBase   string `json:"toBase"`
	Value    string `json:"value"`
}

// NumeralResponse represents the successful JSON response for numeral conversions
type NumeralResponse struct {
	Result string `json:"result"`
}

// HandleNumeralConvert handles HTTP requests for numeral system conversions
// @Summary Numeral System Converter
// @Description Converts between numeral systems like binary, decimal, octal, hexadecimal.
// @Accept json
// @Produce json
// @Param request body api.NumeralRequest true "Conversion Request"
// @Success 200 {object} api.NumeralResponse "Conversion Result"
// @Failure 400 {string} string "Bad Request"
// @Router /convert-numeral [post]
func HandleNumeralConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req NumeralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.FromBase == "" || req.ToBase == "" || req.Value == "" {
		http.Error(w, "fromBase, toBase, and value are required", http.StatusBadRequest)
		return
	}

	result, err := converter.ConvertNumeral(req.FromBase, req.ToBase, req.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := NumeralResponse{Result: result}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
