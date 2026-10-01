package converter

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ============================================================
// Request / Response types
// ============================================================

// EMIRequest carries the validated parameters for an EMI calculation.
type EMIRequest struct {
	Principal          float64 // Loan amount (must be > 0)
	AnnualInterestRate float64 // Annual rate in percent (>= 0)
	Tenure             float64 // Duration of the loan (> 0)
	TenureUnit         string  // "months" | "years"
}

// EMIBreakdown is the principal-vs-interest split as percentages.
type EMIBreakdown struct {
	PrincipalPercent float64 `json:"principalPercent"`
	InterestPercent  float64 `json:"interestPercent"`
}

// AmortizationRow is one row of the month-by-month repayment schedule.
type AmortizationRow struct {
	Month          int     `json:"month"`
	OpeningBalance float64 `json:"openingBalance"`
	PrincipalPaid  float64 `json:"principalPaid"`
	InterestPaid   float64 `json:"interestPaid"`
	TotalPaid      float64 `json:"totalPaid"`
	ClosingBalance float64 `json:"closingBalance"`
}

// EMIResponse is the unified result returned by /calculate-emi.
type EMIResponse struct {
	EMI                float64           `json:"emi"`
	Principal          float64           `json:"principal"`
	TotalInterest      float64           `json:"totalInterest"`
	TotalPayment       float64           `json:"totalPayment"`
	TenureMonths       int               `json:"tenureMonths"`
	MonthlyRatePercent float64           `json:"monthlyRatePercent"`
	Breakdown          EMIBreakdown      `json:"breakdown"`
	Amortization       []AmortizationRow `json:"amortization"`
}

// ============================================================
// Constants
// ============================================================

const (
	// MaxTenureMonths guards against overflow / absurd inputs.
	MaxTenureMonths = 600 // 50 years

	// MaxPrincipal caps the loan amount to something reasonable.
	MaxPrincipal = 1e12 // one trillion

	// emiRoundingPlaces is how many decimals EMI and schedule values
	// are rounded to. 2 mirrors currency conventions.
	emiRoundingPlaces = 2
)

// ============================================================
// Public entry point
// ============================================================

// CalculateEMI computes the Equated Monthly Installment along with the
// full repayment summary and a month-by-month amortization schedule.
func CalculateEMI(req EMIRequest) (*EMIResponse, error) {
	// --- Validate ---
	if err := validateEMIRequest(req); err != nil {
		return nil, err
	}

	// --- Normalize tenure to months ---
	tenureMonths, err := normalizeTenureToMonths(req.Tenure, req.TenureUnit)
	if err != nil {
		return nil, err
	}

	// --- Monthly interest rate (decimal, not percent) ---
	monthlyRate := req.AnnualInterestRate / 12.0 / 100.0
	monthlyRatePercent := req.AnnualInterestRate / 12.0

	// --- Compute EMI ---
	emi := calculateEMIAmount(req.Principal, monthlyRate, tenureMonths)
	emi = roundTo(emi, emiRoundingPlaces)

	// --- Build amortization schedule (also reconciles totals) ---
	schedule, totalInterest, totalPayment := buildAmortizationSchedule(
		req.Principal,
		monthlyRate,
		tenureMonths,
		emi,
	)

	// --- Breakdown percentages ---
	principalPct := 0.0
	interestPct := 0.0
	if totalPayment > 0 {
		principalPct = roundTo((req.Principal/totalPayment)*100.0, emiRoundingPlaces)
		interestPct = roundTo((totalInterest/totalPayment)*100.0, emiRoundingPlaces)
		// Absorb rounding drift so the two always add to 100.
		interestPct = roundTo(100.0-principalPct, emiRoundingPlaces)
	}

	return &EMIResponse{
		EMI:                emi,
		Principal:          roundTo(req.Principal, emiRoundingPlaces),
		TotalInterest:      roundTo(totalInterest, emiRoundingPlaces),
		TotalPayment:       roundTo(totalPayment, emiRoundingPlaces),
		TenureMonths:       tenureMonths,
		MonthlyRatePercent: roundTo(monthlyRatePercent, 6),
		Breakdown: EMIBreakdown{
			PrincipalPercent: principalPct,
			InterestPercent:  interestPct,
		},
		Amortization: schedule,
	}, nil
}

// ============================================================
// Validation
// ============================================================

func validateEMIRequest(req EMIRequest) error {
	if req.Principal <= 0 {
		return errors.New("principal must be greater than zero")
	}
	if req.Principal > MaxPrincipal {
		return fmt.Errorf("principal exceeds the maximum allowed (%.0f)", MaxPrincipal)
	}
	if req.AnnualInterestRate < 0 {
		return errors.New("annualInterestRate cannot be negative")
	}
	if req.AnnualInterestRate > 100 {
		return errors.New("annualInterestRate cannot exceed 100%")
	}
	if req.Tenure <= 0 {
		return errors.New("tenure must be greater than zero")
	}

	unit := strings.ToLower(strings.TrimSpace(req.TenureUnit))
	if unit != "months" && unit != "years" {
		return errors.New("tenureUnit must be 'months' or 'years'")
	}

	return nil
}

func normalizeTenureToMonths(tenure float64, unit string) (int, error) {
	unit = strings.ToLower(strings.TrimSpace(unit))

	var months float64
	switch unit {
	case "months":
		months = tenure
	case "years":
		months = tenure * 12.0
	default:
		return 0, errors.New("tenureUnit must be 'months' or 'years'")
	}

	if months < 1 {
		return 0, errors.New("tenure must be at least 1 month")
	}

	rounded := int(math.Round(months))
	if rounded > MaxTenureMonths {
		return 0, fmt.Errorf(
			"tenure exceeds the maximum of %d months (%d years)",
			MaxTenureMonths, MaxTenureMonths/12,
		)
	}
	return rounded, nil
}

// ============================================================
// Core math
// ============================================================

// calculateEMIAmount applies the standard EMI formula:
//
//	EMI = [P * r * (1+r)^N] / [(1+r)^N − 1]
//
// Special case: r == 0 (interest-free loan) → EMI = P / N.
func calculateEMIAmount(principal, monthlyRate float64, months int) float64 {
	if monthlyRate == 0 {
		return principal / float64(months)
	}

	factor := math.Pow(1+monthlyRate, float64(months))
	return (principal * monthlyRate * factor) / (factor - 1)
}

// buildAmortizationSchedule walks the loan month by month, applying the
// EMI and splitting each payment into interest and principal.
//
// The LAST row is adjusted so that the closing balance lands exactly on
// zero — absorbing any rounding drift from the intermediate months.
//
// Returns (schedule, totalInterest, totalPayment).
func buildAmortizationSchedule(
	principal float64,
	monthlyRate float64,
	months int,
	emi float64,
) ([]AmortizationRow, float64, float64) {

	schedule := make([]AmortizationRow, 0, months)
	balance := principal
	var totalInterest, totalPaid float64

	for m := 1; m <= months; m++ {
		opening := balance

		interest := roundTo(opening*monthlyRate, emiRoundingPlaces)
		principalComponent := roundTo(emi-interest, emiRoundingPlaces)

		// On the final month, sweep any remaining balance so we close at 0.
		if m == months {
			principalComponent = roundTo(opening, emiRoundingPlaces)
		}

		// Guard against a principal component that overshoots the balance.
		if principalComponent > opening {
			principalComponent = roundTo(opening, emiRoundingPlaces)
		}

		actualPayment := roundTo(principalComponent+interest, emiRoundingPlaces)
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
			// Early payoff (shouldn't normally happen) — stop.
			if m < months {
				months = m
			}
			break
		}
	}

	return schedule, roundTo(totalInterest, emiRoundingPlaces), roundTo(totalPaid, emiRoundingPlaces)
}

// ============================================================
// Helper
// ============================================================

// roundTo is defined in percentage.go — reused here so we don't duplicate.
