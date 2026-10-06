package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// EMIRequest is the incoming JSON body for /calculate-emi.
type EMIRequest struct {
	Principal          float64 `json:"principal"`
	AnnualInterestRate float64 `json:"annualInterestRate"`
	Tenure             float64 `json:"tenure"`
	TenureUnit         string  `json:"tenureUnit"` // "months" or "years"
}

// HandleEMICalculate handles HTTP requests for Loan EMI calculations.
// @Summary      Loan EMI Calculator
// @Description  Calculates the Equated Monthly Installment (EMI) for a loan along with
// @Description  the total interest, total payment, principal/interest split, and a
// @Description  month-by-month amortization schedule.
// @Description
// @Description  **Formula:** `EMI = [P × r × (1+r)^N] / [(1+r)^N − 1]`
// @Description  where `r` is the monthly interest rate (annual / 12 / 100) and `N` is
// @Description  the tenure in months.
// @Description
// @Description  **Special case:** an annual interest rate of `0` returns a simple
// @Description  `principal / tenure` division (interest-free loan).
// @Description
// @Description  **Limits:**
// @Description  - `principal` must be positive and ≤ 1 trillion.
// @Description  - `annualInterestRate` must be between 0 and 100.
// @Description  - `tenure` must be positive; tenure in months must be ≤ 600.
// @Description  - `tenureUnit` must be exactly `"months"` or `"years"`.
// @Tags         Calculators
// @Accept       json
// @Produce      json
// @Param        request body api.EMIRequest true "EMI Calculation Request"
// @Success      200 {object} converter.EMIResponse "EMI Result"
// @Failure      400 {string} string "Bad Request - invalid input"
// @Router       /calculate-emi [post]
func HandleEMICalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EMIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.CalculateEMI(converter.EMIRequest{
		Principal:          req.Principal,
		AnnualInterestRate: req.AnnualInterestRate,
		Tenure:             req.Tenure,
		TenureUnit:         req.TenureUnit,
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
