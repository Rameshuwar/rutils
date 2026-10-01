package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// ============================================================
// SIMPLE INTEREST
// ============================================================

// SimpleInterestRequest is the incoming JSON body for
// /calculate-simple-interest.
type SimpleInterestRequest struct {
	Principal          float64 `json:"principal"`
	AnnualInterestRate float64 `json:"annualInterestRate"`
	Time               float64 `json:"time"`
	TimeUnit           string  `json:"timeUnit"` // "years" | "months" | "days"
}

// HandleSimpleInterestCalculate handles HTTP requests for the Simple
// Interest Calculator.
//
// @Summary      Simple Interest Calculator
// @Description  Calculates the simple interest and total amount for a loan or investment.
// @Description
// @Description  **Formula:** `SI = (P Г— R Г— T) / 100`
// @Description  where `P` is the principal, `R` is the annual rate in percent,
// @Description  and `T` is the tenure expressed in years.
// @Description
// @Description  **Limits:**
// @Description  - `principal` must be positive and вүӨ 1 trillion.
// @Description  - `annualInterestRate` must be between 0 and 100.
// @Description  - `time` must be positive; tenure must be вүӨ 100 years.
// @Description  - `timeUnit` must be `"years"`, `"months"`, or `"days"`.
// @Tags         Calculators
// @Accept       json
// @Produce      json
// @Param        request body api.SimpleInterestRequest true "Simple Interest Calculation Request"
// @Success      200 {object} converter.SimpleInterestResponse "Simple Interest Result"
// @Failure      400 {string} string "Bad Request - invalid input"
// @Router       /calculate-simple-interest [post]
func HandleSimpleInterestCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SimpleInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.CalculateSimpleInterest(converter.SimpleInterestRequest{
		Principal:          req.Principal,
		AnnualInterestRate: req.AnnualInterestRate,
		Time:               req.Time,
		TimeUnit:           req.TimeUnit,
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

// ============================================================
// COMPOUND INTEREST
// ============================================================

// CompoundInterestRequest is the incoming JSON body for
// /calculate-compound-interest.
type CompoundInterestRequest struct {
	Principal            float64 `json:"principal"`
	AnnualInterestRate   float64 `json:"annualInterestRate"`
	Time                 float64 `json:"time"`
	TimeUnit             string  `json:"timeUnit"`             // "years" | "months"
	CompoundingFrequency string  `json:"compoundingFrequency"` // "yearly" | "half_yearly" | "quarterly" | "monthly" | "daily"
}

// HandleCompoundInterestCalculate handles HTTP requests for the
// Compound Interest Calculator.
//
// @Summary      Compound Interest Calculator
// @Description  Calculates the compound interest, total amount, effective annual rate,
// @Description  and (for tenures вүҘ 1 year) a year-by-year growth schedule.
// @Description
// @Description  **Formula:** `A = P Г— (1 + r/n)^(nГ—t)`   and   `CI = A вҲ’ P`
// @Description  where `r` is the annual rate in decimal form, `n` is the number of
// @Description  compounding periods per year, and `t` is the tenure in years.
// @Description
// @Description  **Limits:**
// @Description  - `principal` must be positive and вүӨ 1 trillion.
// @Description  - `annualInterestRate` must be between 0 and 100.
// @Description  - `time` must be positive; tenure must be вүӨ 100 years.
// @Description  - `timeUnit` must be `"years"` or `"months"` (day-based tenure is rejected).
// @Description  - `compoundingFrequency` must be one of `yearly`, `half_yearly`,
// @Description    `quarterly`, `monthly`, `daily` (defaults to `yearly`).
// @Tags         Calculators
// @Accept       json
// @Produce      json
// @Param        request body api.CompoundInterestRequest true "Compound Interest Calculation Request"
// @Success      200 {object} converter.CompoundInterestResponse "Compound Interest Result"
// @Failure      400 {string} string "Bad Request - invalid input"
// @Router       /calculate-compound-interest [post]
func HandleCompoundInterestCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompoundInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.CalculateCompoundInterest(converter.CompoundInterestRequest{
		Principal:            req.Principal,
		AnnualInterestRate:   req.AnnualInterestRate,
		Time:                 req.Time,
		TimeUnit:             req.TimeUnit,
		CompoundingFrequency: req.CompoundingFrequency,
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
