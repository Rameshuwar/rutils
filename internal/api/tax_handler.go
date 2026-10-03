package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// TaxRequest is the incoming JSON body for /calculate-tax.
type TaxRequest struct {
	Mode string `json:"mode"`

	// GST / VAT style
	Amount    float64 `json:"amount,omitempty"`
	TaxRate   float64 `json:"taxRate,omitempty"`
	TaxType   string  `json:"taxType,omitempty"`
	NetAmount float64 `json:"netAmount,omitempty"`
	GrossAmt  float64 `json:"grossAmount,omitempty"`
	TaxPaid   float64 `json:"taxPaid,omitempty"`

	// Income-tax style
	Income float64             `json:"income,omitempty"`
	Regime string              `json:"regime,omitempty"`
	Slabs  []converter.TaxSlab `json:"slabs,omitempty"`
}

// HandleTaxCalculate handles HTTP requests for all tax / VAT / GST calculations.
// @Summary      Tax / VAT / GST Calculator
// @Description  Performs any tax calculation through a single `mode`-discriminated endpoint.
// @Description
// @Description  **Supported modes (`mode` field):**
// @Description  - `add_tax`      (amount, taxRate)                       — exclusive → inclusive
// @Description  - `remove_tax`   (amount, taxRate)                       — inclusive → exclusive
// @Description  - `find_rate`    (netAmount, grossAmount)                — effective tax rate
// @Description  - `split_gst`    (amount, taxRate, taxType)              — CGST+SGST or IGST breakdown
// @Description  - `reverse_gst`  (taxPaid, taxRate)                      — recover net from tax paid
// @Description  - `income_tax`   (income, slabs[], regime?)              — progressive slab tax
// @Description
// @Description  **Notes:**
// @Description  - The endpoint is country-neutral: the client supplies the rate and, for income tax, the slab schedule.
// @Description  - For `split_gst`, `taxType` must be `cgst_sgst` (default) or `igst`.
// @Description  - For `income_tax`, slabs must be non-overlapping; at most one slab may have `to = 0` (open-ended).
// @Tags         Calculators
// @Accept       json
// @Produce      json
// @Param        request body api.TaxRequest true "Tax Calculation Request"
// @Success      200 {object} converter.TaxResponse "Tax Result"
// @Failure      400 {string} string "Bad Request - invalid mode or parameters"
// @Router       /calculate-tax [post]
func HandleTaxCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TaxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.CalculateTax(converter.TaxRequest{
		Mode:      req.Mode,
		Amount:    req.Amount,
		TaxRate:   req.TaxRate,
		TaxType:   req.TaxType,
		NetAmount: req.NetAmount,
		GrossAmt:  req.GrossAmt,
		TaxPaid:   req.TaxPaid,
		Income:    req.Income,
		Regime:    req.Regime,
		Slabs:     req.Slabs,
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
