package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// PercentageRequest is the incoming JSON body for /calculate-percentage
type PercentageRequest struct {
	Operation string   `json:"operation"`
	Value1    float64  `json:"value1"`
	Value2    float64  `json:"value2"`
	Value3    *float64 `json:"value3,omitempty"`
}

// HandlePercentageCalculate handles HTTP requests for all percentage calculations
// @Summary Percentage Calculator
// @Description Performs any percentage operation (X% of Y, what %, change, increase, decrease,
//
//	reverse, difference, discount, markup, profit/loss, fraction/decimal conversion,
//	compound %, marks %, CGPA-to-%) via a single `operation`-discriminated endpoint.
//
// @Description
// @Description **Supported operations (`operation` field):**
// @Description - `percent_of`              (value1 = percent, value2 = number)
// @Description - `what_percent`            (value1 = part, value2 = whole)
// @Description - `percent_change`          (value1 = old, value2 = new)
// @Description - `percent_increase`        (value1 = number, value2 = percent)
// @Description - `percent_decrease`        (value1 = number, value2 = percent)
// @Description - `reverse_percent`         (value1 = final, value2 = percent)
// @Description - `is_percent_of_what`      (value1 = part, value2 = percent)
// @Description - `percent_difference`      (value1 = a, value2 = b)
// @Description - `add_percent_points`      (value1 = p1%, value2 = p2%)
// @Description - `subtract_percent_points` (value1 = p1%, value2 = p2%)
// @Description - `discount`                (value1 = price, value2 = discount%)
// @Description - `markup`                  (value1 = cost, value2 = markup%)
// @Description - `profit_loss`             (value1 = cost, value2 = selling)
// @Description - `percent_to_fraction`     (value1 = percent)
// @Description - `fraction_to_percent`     (value1 = numerator, value2 = denominator)
// @Description - `decimal_to_percent`      (value1 = decimal)
// @Description - `compound_percent`        (value1 = base, value2 = percent, value3 = periods)
// @Description - `marks_percentage`        (value1 = obtained, value2 = total)
// @Description - `cgpa_to_percent`         (value1 = cgpa; uses CBSE factor 9.5)
// @Tags Percentage Calculator
// @Accept json
// @Produce json
// @Param request body api.PercentageRequest true "Percentage Calculation Request"
// @Success 200 {object} converter.PercentageResponse "Percentage Result"
// @Failure 400 {string} string "Bad Request - invalid operation or parameters"
// @Router /calculate-percentage [post]
func HandlePercentageCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PercentageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Operation == "" {
		http.Error(w, "operation is required", http.StatusBadRequest)
		return
	}

	hasValue3 := req.Value3 != nil
	var value3 float64
	if hasValue3 {
		value3 = *req.Value3
	}

	res, err := converter.CalculatePercentage(
		req.Operation,
		req.Value1,
		req.Value2,
		value3,
		hasValue3,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}