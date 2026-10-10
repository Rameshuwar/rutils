package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// LoanTenureRequest is the incoming JSON body for /calculate-loan-tenure.
type LoanTenureRequest struct {
	Principal          float64 `json:"principal"`
	AnnualInterestRate float64 `json:"annualInterestRate"`
	MonthlyPayment     float64 `json:"monthlyPayment"`
}

// HandleLoanTenureCalculate handles HTTP requests for the
// borrower-centric loan tenure calculator.
//
// @Summary      Loan Tenure Calculator (Borrower-Centric)
// @Description  Computes how long it will take to repay a loan when the borrower fixes a monthly payment.
// @Description
// @Description  **Use case:** a user has a loan amount and knows only how much they can pay per month — this endpoint tells them how many months/years it will take to repay, and how much interest they will pay in total.
// @Description
// @Description  **Formula (loan amortization solved for N):**
// @Description  `N = -log(1 - (P × r) / EMI) / log(1 + r)`
// @Description  where `r` is the monthly rate (annual / 12 / 100) and `EMI` is the borrower's fixed monthly payment.
// @Description
// @Description  **Rejection rule:** if the monthly payment does not at least cover the monthly interest on the principal (i.e. `EMI ≤ P × r`), the loan would never be repaid, so the endpoint returns a 400 with a helpful suggestion.
// @Description
// @Description  **Limits:**
// @Description  - `principal` must be positive and ≤ 1 trillion.
// @Description  - `annualInterestRate` must be between 0 and 100.
// @Description  - `monthlyPayment` must be positive.
// @Description  - Computed tenure must not exceed 600 months (50 years).
// @Tags         Calculators
// @Accept       json
// @Produce      json
// @Param        request body api.LoanTenureRequest true "Loan Tenure Calculation Request"
// @Success      200 {object} converter.LoanTenureResponse "Loan Tenure Result"
// @Failure      400 {string} string "Bad Request - invalid input or unrepayable loan"
// @Router       /calculate-loan-tenure [post]
func HandleLoanTenureCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoanTenureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	res, err := converter.CalculateLoanTenure(converter.LoanTenureRequest{
		Principal:          req.Principal,
		AnnualInterestRate: req.AnnualInterestRate,
		MonthlyPayment:     req.MonthlyPayment,
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