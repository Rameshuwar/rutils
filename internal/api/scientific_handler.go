package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// ScientificRequest is the incoming JSON body for /calculate-scientific.
type ScientificRequest struct {
	Operation string   `json:"operation"`
	Value1    float64  `json:"value1"`
	Value2    *float64 `json:"value2,omitempty"`
	AngleUnit string   `json:"angleUnit,omitempty"`
}

// HandleScientificCalculate handles HTTP requests for all scientific
// calculator operations.
//
// @Summary      Scientific Calculator
// @Description  Performs any scientific / mathematical operation through a single `operation`-discriminated endpoint.
// @Description
// @Description  **Operation groups (`operation` field):**
// @Description  - Constants:      `pi`, `e`
// @Description  - Trigonometric:  `sin`, `cos`, `tan`, `csc`, `sec`, `cot`
// @Description  - Inverse trig:   `asin`, `acos`, `atan`, `atan2` (2-arg)
// @Description  - Hyperbolic:     `sinh`, `cosh`, `tanh`, `asinh`, `acosh`, `atanh`
// @Description  - Log / Exp:      `log`, `ln`, `log_base` (2-arg), `exp`
// @Description  - Powers / Roots: `pow` (2-arg), `sqrt`, `cbrt`, `nth_root` (2-arg)
// @Description  - Rounding:       `abs`, `floor`, `ceil`, `round`, `trunc`, `sign`
// @Description  - Combinatorics:  `factorial`, `ncr` (2-arg), `npr` (2-arg)
// @Description  - Misc:           `gcd` (2-arg), `lcm` (2-arg), `mod` (2-arg), `hypot` (2-arg)
// @Description
// @Description  **Angle unit:** Trig and inverse-trig operations accept an optional
// @Description  `angleUnit` field (`"degrees"` or `"radians"`, default `"radians"`).
// @Description  The same unit governs both the input of `sin/cos/tan/...` and the
// @Description  output of `asin/acos/atan/atan2`. Hyperbolic operations are unaffected.
// @Description
// @Description  **Validation:** Domain violations (e.g. `sqrt(-1)`, `ln(0)`, `asin(2)`,
// @Description  `tan(90°)`) return a 400 error with a descriptive message.
// @Tags         Scientific Calculator
// @Accept       json
// @Produce      json
// @Param        request body api.ScientificRequest true "Scientific Calculation Request"
// @Success      200 {object} converter.ScientificResponse "Scientific Result"
// @Failure      400 {string} string "Bad Request - invalid operation, parameters, or domain error"
// @Router       /calculate-scientific [post]
func HandleScientificCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScientificRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.CalculateScientific(converter.ScientificRequest{
		Operation: req.Operation,
		Value1:    req.Value1,
		Value2:    req.Value2,
		AngleUnit: req.AngleUnit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
