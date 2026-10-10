package converter

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ============================================================
// LOAN TENURE CALCULATOR (Borrower-Centric)
//
// Answers the question: "If I can only pay ₹X per month, how long
// will it take to clear my loan, and what will it cost me?"
//
// Formula (loan amortization solved for N):
//
//	N = -log(1 - (P × r) / EMI) / log(1 + r)
//
// where:
//   P   = principal
//   r   = monthly interest rate (annual / 12 / 100)
//   EMI = the fixed monthly payment the borrower can afford
//
// Guard: if EMI <= P × r, the payment doesn't even cover the
// monthly interest, so the loan is mathematically impossible to
// repay. We reject that case.
// ============================================================

// LoanTenureRequest carries the validated parameters for a
// borrower-centric loan tenure calculation.
type LoanTenureRequest struct {
	Principal          float64 // Loan amount (must be > 0)
	AnnualInterestRate float64 // Annual rate in percent (>= 0)
	MonthlyPayment     float64 // Fixed monthly payment the user can afford (> 0)
}

// LoanTenureResponse is the unified result returned by
// /calculate-loan-tenure. It reuses the existing EMI response
// shape (AmortizationRow, EMIBreakdown) so the frontend can render
// it with the same components it already has.
type LoanTenureResponse struct {
	EMI                float64           `json:"emi"`                // Echoes the input MonthlyPayment
	Principal          float64           `json:"principal"`
	TotalInterest      float64           `json:"totalInterest"`
	TotalPayment       float64           `json:"totalPayment"`
	TenureMonths       int               `json:"tenureMonths"`       // ← the key output
	TenureYears        float64           `json:"tenureYears"`        // ← convenience
	MonthlyRatePercent float64           `json:"monthlyRatePercent"`
	Breakdown          EMIBreakdown      `json:"breakdown"`
	Amortization       []AmortizationRow `json:"amortization"`
}

// ============================================================
// Public entry point
// ============================================================

// CalculateLoanTenure computes how long it takes to repay a loan
// given a fixed monthly payment.
func CalculateLoanTenure(req LoanTenureRequest) (*LoanTenureResponse, error) {
	// --- Validate ---
	if err := validateLoanTenureRequest(req); err != nil {
		return nil, err
	}

	// --- Monthly interest rate (decimal, not percent) ---
	monthlyRate := req.AnnualInterestRate / 12.0 / 100.0
	monthlyRatePercent := req.AnnualInterestRate / 12.0

	// --- Zero-interest shortcut: N = P / EMI ---
	if monthlyRate == 0 {
		tenureMonths := int(math.Ceil(req.Principal / req.MonthlyPayment))
		if tenureMonths < 1 {
			tenureMonths = 1
		}
		if tenureMonths > MaxTenureMonths {
			return nil, fmt.Errorf(
				"loan cannot be repaid within %d months (%.1f years) at this payment; "+
					"increase your monthly payment or reduce the principal",
				MaxTenureMonths, float64(MaxTenureMonths)/12.0,
			)
		}
		return buildLoanTenureResponse(req, monthlyRate, monthlyRatePercent, tenureMonths), nil
	}

	// --- Guard: payment must cover monthly interest ---
	monthlyInterestOnPrincipal := req.Principal * monthlyRate
	if req.MonthlyPayment <= monthlyInterestOnPrincipal {
		return nil, fmt.Errorf(
			"monthly payment (%.2f) is too low to cover the monthly interest (%.2f). "+
				"Increase your payment to at least %.2f",
			req.MonthlyPayment,
			monthlyInterestOnPrincipal,
			monthlyInterestOnPrincipal*1.01, // suggest 1% above the minimum
		)
	}

	// --- Solve for N ---
	// N = -log(1 - (P × r) / EMI) / log(1 + r)
	numerator := -math.Log(1-(req.Principal*monthlyRate)/req.MonthlyPayment)
	denominator := math.Log(1 + monthlyRate)

	rawTenure := numerator / denominator

	if math.IsNaN(rawTenure) || math.IsInf(rawTenure, 0) || rawTenure <= 0 {
		return nil, errors.New("unable to compute loan tenure from the given inputs")
	}

	// Round up to the next whole month. The final payment will be
	// smaller than the regular EMI (the "sweep" payment).
	tenureMonths := int(math.Ceil(rawTenure))

	if tenureMonths < 1 {
		tenureMonths = 1
	}
	if tenureMonths > MaxTenureMonths {
		return nil, fmt.Errorf(
			"loan cannot be repaid within %d months (%.1f years) at this payment; "+
				"increase your monthly payment or reduce the principal",
			MaxTenureMonths, float64(MaxTenureMonths)/12.0,
		)
	}

	return buildLoanTenureResponse(req, monthlyRate, monthlyRatePercent, tenureMonths), nil
}

// ============================================================
// Validation
// ============================================================

func validateLoanTenureRequest(req LoanTenureRequest) error {
	if math.IsNaN(req.Principal) || math.IsInf(req.Principal, 0) {
		return errors.New("principal must be a finite number")
	}
	if req.Principal <= 0 {
		return errors.New("principal must be greater than zero")
	}
	if req.Principal > MaxPrincipal {
		return fmt.Errorf("principal exceeds the maximum allowed (%.0f)", MaxPrincipal)
	}

	if math.IsNaN(req.AnnualInterestRate) || math.IsInf(req.AnnualInterestRate, 0) {
		return errors.New("annualInterestRate must be a finite number")
	}
	if req.AnnualInterestRate < 0 {
		return errors.New("annualInterestRate cannot be negative")
	}
	if req.AnnualInterestRate > 100 {
		return errors.New("annualInterestRate cannot exceed 100%")
	}

	if math.IsNaN(req.MonthlyPayment) || math.IsInf(req.MonthlyPayment, 0) {
		return errors.New("monthlyPayment must be a finite number")
	}
	if req.MonthlyPayment <= 0 {
		return errors.New("monthlyPayment must be greater than zero")
	}

	return nil
}

// ============================================================
// Response builder
// ============================================================

// buildLoanTenureResponse walks the loan month by month, applying
// the fixed monthly payment, and produces the full response. The
// final row is adjusted so the closing balance lands exactly on 0,
// absorbing any accumulated rounding drift.
func buildLoanTenureResponse(
	req LoanTenureRequest,
	monthlyRate float64,
	monthlyRatePercent float64,
	tenureMonths int,
) *LoanTenureResponse {

	schedule := make([]AmortizationRow, 0, tenureMonths)
	balance := req.Principal
	var totalInterest, totalPaid float64

	for m := 1; m <= tenureMonths; m++ {
		opening := balance
		interest := roundTo(opening*monthlyRate, emiRoundingPlaces)

		// On the final month, pay off whatever is left.
		var principalComponent, actualPayment float64
		if m == tenureMonths {
			principalComponent = roundTo(opening, emiRoundingPlaces)
			actualPayment = roundTo(principalComponent+interest, emiRoundingPlaces)
		} else {
			principalComponent = roundTo(req.MonthlyPayment-interest, emiRoundingPlaces)
			// Guard against a component that overshoots the balance.
			if principalComponent > opening {
				principalComponent = roundTo(opening, emiRoundingPlaces)
			}
			actualPayment = roundTo(principalComponent+interest, emiRoundingPlaces)
		}

		closing := roundTo(opening-principalComponent, emiRoundingPlaces)
		if closing < 0 {
			closing = 0
		}

		schedule = append(schedule, AmortizationRow{
			Month:          m,
			OpeningBalance: opening,
			PrincipalPaid:  principalComponent,
			InterestPaid:   interest,
			TotalPaid:      actualPayment,
			ClosingBalance: closing,
		})

		totalInterest += interest
		totalPaid += actualPayment
		balance = closing

		if balance == 0 {
			break
		}
	}

	totalInterest = roundTo(totalInterest, emiRoundingPlaces)
	totalPaid = roundTo(totalPaid, emiRoundingPlaces)

	// --- Breakdown percentages ---
	principalPct := 0.0
	interestPct := 0.0
	if totalPaid > 0 {
		principalPct = roundTo((req.Principal/totalPaid)*100.0, emiRoundingPlaces)
		interestPct = roundTo(100.0-principalPct, emiRoundingPlaces)
	}

	return &LoanTenureResponse{
		EMI:                roundTo(req.MonthlyPayment, emiRoundingPlaces),
		Principal:          roundTo(req.Principal, emiRoundingPlaces),
		TotalInterest:      totalInterest,
		TotalPayment:       totalPaid,
		TenureMonths:       len(schedule),
		TenureYears:        roundTo(float64(len(schedule))/12.0, 2),
		MonthlyRatePercent: roundTo(monthlyRatePercent, 6),
		Breakdown: EMIBreakdown{
			PrincipalPercent: principalPct,
			InterestPercent:  interestPct,
		},
		Amortization: schedule,
	}
}

// ============================================================
// Small helpers
// ============================================================

// loanTenureLabel is a friendlier formatter for the tenure output
// used in error messages and step strings.
func loanTenureLabel(months int) string {
	if months < 12 {
		return fmt.Sprintf("%d month(s)", months)
	}
	years := months / 12
	rem := months % 12
	if rem == 0 {
		return fmt.Sprintf("%d year(s)", years)
	}
	return fmt.Sprintf("%d year(s), %d month(s)", years, rem)
}

// ensure strings is imported for any future extension; keeps the
// compiler happy if label helpers are expanded later.
var _ = strings.TrimSpace
// contains is a tiny slice-membership helper used by the tests in this
// file. The `converter` package doesn't define one, so we add it locally
// to avoid pulling in a test helper from another package.
func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}