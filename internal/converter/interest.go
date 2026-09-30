package converter

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ============================================================
// Shared constants
// ============================================================

const (
	// MaxInterestPrincipal caps the loan/investment amount.
	MaxInterestPrincipal = 1e12 // one trillion

	// MaxInterestTimeYears guards against overflow.
	MaxInterestTimeYears = 100 // 100 years

	// interestRoundingPlaces mirrors emiRoundingPlaces for currency output.
	interestRoundingPlaces = 2

	// daysInYear is used when converting a day-based tenure to years.
	daysInYear = 365.0
)

// ============================================================
// SIMPLE INTEREST
// ============================================================

// SimpleInterestRequest carries the validated parameters for a simple
// interest calculation.
type SimpleInterestRequest struct {
	Principal          float64 // Must be > 0
	AnnualInterestRate float64 // Annual rate in percent (>= 0)
	Time               float64 // Duration (> 0)
	TimeUnit           string  // "years" | "months" | "days"
}

// InterestBreakdown is the principal-vs-interest split as percentages.
type InterestBreakdown struct {
	PrincipalPercent float64 `json:"principalPercent"`
	InterestPercent  float64 `json:"interestPercent"`
}

// SimpleInterestResponse is the unified result returned by
// /calculate-simple-interest.
type SimpleInterestResponse struct {
	Principal          float64           `json:"principal"`
	Interest           float64           `json:"interest"`
	TotalAmount        float64           `json:"totalAmount"`
	AnnualInterestRate float64           `json:"annualInterestRate"`
	TimeYears          float64           `json:"timeYears"`
	Breakdown          InterestBreakdown `json:"breakdown"`
	Steps              []string          `json:"steps"`
}

// CalculateSimpleInterest computes SI and the total amount.
//
//	SI = (P × R × T) / 100
//
// where T is the tenure expressed in years.
func CalculateSimpleInterest(req SimpleInterestRequest) (*SimpleInterestResponse, error) {
	// --- Validate ---
	if err := validateInterestCommon(req.Principal, req.AnnualInterestRate); err != nil {
		return nil, err
	}

	timeYears, err := normalizeTimeToYears(req.Time, req.TimeUnit)
	if err != nil {
		return nil, err
	}

	// --- Compute ---
	interest := roundTo((req.Principal*req.AnnualInterestRate*timeYears)/100.0, interestRoundingPlaces)
	total := roundTo(req.Principal+interest, interestRoundingPlaces)

	breakdown := buildInterestBreakdown(req.Principal, total)

	steps := []string{
		"SI = (P X R X T) / 100",
		fmt.Sprintf("SI = (%s X %s X %s) / 100",
			formatNumber(req.Principal),
			formatNumber(req.AnnualInterestRate),
			formatNumber(timeYears),
		),
		fmt.Sprintf("SI = %s", formatNumber(interest)),
		fmt.Sprintf("Total Amount = %s + %s = %s",
			formatNumber(req.Principal),
			formatNumber(interest),
			formatNumber(total),
		),
	}

	return &SimpleInterestResponse{
		Principal:          roundTo(req.Principal, interestRoundingPlaces),
		Interest:           interest,
		TotalAmount:        total,
		AnnualInterestRate: req.AnnualInterestRate,
		TimeYears:          roundTo(timeYears, 6),
		Breakdown:          breakdown,
		Steps:              steps,
	}, nil
}

// ============================================================
// COMPOUND INTEREST
// ============================================================

// CompoundInterestRequest carries the validated parameters for a
// compound interest calculation.
type CompoundInterestRequest struct {
	Principal             float64 // Must be > 0
	AnnualInterestRate    float64 // Annual rate in percent (>= 0)
	Time                  float64 // Duration (> 0)
	TimeUnit              string  // "years" | "months"
	CompoundingFrequency  string  // "yearly" | "half_yearly" | "quarterly" | "monthly" | "daily"
}

// YearlyRow is one row of the year-by-year growth schedule.
type YearlyRow struct {
	Year            int     `json:"year"`
	OpeningBalance  float64 `json:"openingBalance"`
	InterestEarned  float64 `json:"interestEarned"`
	ClosingBalance  float64 `json:"closingBalance"`
}

// CompoundInterestResponse is the unified result returned by
// /calculate-compound-interest.
type CompoundInterestResponse struct {
	Principal             float64           `json:"principal"`
	Interest              float64           `json:"interest"`
	TotalAmount           float64           `json:"totalAmount"`
	AnnualInterestRate    float64           `json:"annualInterestRate"`
	TimeYears             float64           `json:"timeYears"`
	CompoundingFrequency  string            `json:"compoundingFrequency"`
	CompoundsPerYear      int               `json:"compoundsPerYear"`
	EffectiveAnnualRate   float64           `json:"effectiveAnnualRate"`
	Breakdown             InterestBreakdown `json:"breakdown"`
	YearlyBreakdown       []YearlyRow       `json:"yearlyBreakdown,omitempty"`
	Steps                 []string          `json:"steps"`
}

// compoundsPerYearMap maps a frequency name to compounds per year.
var compoundsPerYearMap = map[string]int{
	"yearly":      1,
	"half_yearly": 2,
	"quarterly":   4,
	"monthly":     12,
	"daily":       365,
}

// CalculateCompoundInterest computes CI and the total amount.
//
//	A = P × (1 + r/n)^(n×t)
//	CI = A − P
//
// where r is the annual rate in decimal form, n is the number of
// compounding periods per year, and t is the tenure in years.
func CalculateCompoundInterest(req CompoundInterestRequest) (*CompoundInterestResponse, error) {
	// --- Validate ---
	if err := validateInterestCommon(req.Principal, req.AnnualInterestRate); err != nil {
		return nil, err
	}

	timeYears, err := normalizeTimeToYears(req.Time, req.TimeUnit)
	if err != nil {
		return nil, err
	}
	// Daily compounding is only meaningful for year/month tenures —
	// reject day-based tenure to keep things sane.
	if strings.ToLower(strings.TrimSpace(req.TimeUnit)) == "days" {
		return nil, errors.New("timeUnit must be 'years' or 'months' for compound interest")
	}

	freq := strings.ToLower(strings.TrimSpace(req.CompoundingFrequency))
	if freq == "" {
		freq = "yearly"
	}
	n, ok := compoundsPerYearMap[freq]
	if !ok {
		return nil, errors.New("compoundingFrequency must be one of: yearly, half_yearly, quarterly, monthly, daily")
	}

	// --- Compute ---
	r := req.AnnualInterestRate / 100.0
	total := roundTo(req.Principal*math.Pow(1+r/float64(n), float64(n)*timeYears), interestRoundingPlaces)
	interest := roundTo(total-req.Principal, interestRoundingPlaces)

	breakdown := buildInterestBreakdown(req.Principal, total)

	// Effective annual rate (EAR)
	effectiveAnnualRate := 0.0
	if req.AnnualInterestRate > 0 {
		effectiveAnnualRate = roundTo((math.Pow(1+r/float64(n), float64(n))-1)*100.0, 6)
	}

	// Year-by-year growth schedule (only when tenure вүҘ 1 year).
	schedule := buildYearlyBreakdown(req.Principal, r, n, timeYears)

	steps := []string{
		"A = P X (1 + r/n)^(n X t)",
		fmt.Sprintf("A = %s X (1 + %s/%d)^(%d X %s)",
			formatNumber(req.Principal),
			formatNumber(r),
			n, n,
			formatNumber(timeYears),
		),
		fmt.Sprintf("A = %s", formatNumber(total)),
		fmt.Sprintf("CI = A - P = %s - %s = %s",
			formatNumber(total),
			formatNumber(req.Principal),
			formatNumber(interest),
		),
	}

	return &CompoundInterestResponse{
		Principal:            roundTo(req.Principal, interestRoundingPlaces),
		Interest:             interest,
		TotalAmount:          total,
		AnnualInterestRate:   req.AnnualInterestRate,
		TimeYears:            roundTo(timeYears, 6),
		CompoundingFrequency: freq,
		CompoundsPerYear:     n,
		EffectiveAnnualRate:  effectiveAnnualRate,
		Breakdown:            breakdown,
		YearlyBreakdown:      schedule,
		Steps:                steps,
	}, nil
}

// ============================================================
// Validation
// ============================================================

func validateInterestCommon(principal, rate float64) error {
	if math.IsNaN(principal) || math.IsInf(principal, 0) {
		return errors.New("principal must be a finite number")
	}
	if principal <= 0 {
		return errors.New("principal must be greater than zero")
	}
	if principal > MaxInterestPrincipal {
		return fmt.Errorf("principal exceeds the maximum allowed (%.0f)", MaxInterestPrincipal)
	}
	if math.IsNaN(rate) || math.IsInf(rate, 0) {
		return errors.New("annualInterestRate must be a finite number")
	}
	if rate < 0 {
		return errors.New("annualInterestRate cannot be negative")
	}
	if rate > 100 {
		return errors.New("annualInterestRate cannot exceed 100%")
	}
	return nil
}

func normalizeTimeToYears(time float64, unit string) (float64, error) {
	if math.IsNaN(time) || math.IsInf(time, 0) {
		return 0, errors.New("time must be a finite number")
	}
	if time <= 0 {
		return 0, errors.New("time must be greater than zero")
	}

	unit = strings.ToLower(strings.TrimSpace(unit))

	var years float64
	switch unit {
	case "years", "":
		years = time
	case "months":
		years = time / 12.0
	case "days":
		years = time / daysInYear
	default:
		return 0, errors.New("timeUnit must be 'years', 'months', or 'days'")
	}

	if years > MaxInterestTimeYears {
		return 0, fmt.Errorf("time exceeds the maximum of %d years", MaxInterestTimeYears)
	}
	return years, nil
}

// ============================================================
// Helpers
// ============================================================

// buildInterestBreakdown computes the principal/interest split as
// percentages. The interest side absorbs the rounding drift so the
// two always sum to exactly 100.
func buildInterestBreakdown(principal, total float64) InterestBreakdown {
	if total <= 0 {
		return InterestBreakdown{PrincipalPercent: 100, InterestPercent: 0}
	}
	principalPct := roundTo((principal/total)*100.0, interestRoundingPlaces)
	interestPct := roundTo(100.0-principalPct, interestRoundingPlaces)
	return InterestBreakdown{
		PrincipalPercent: principalPct,
		InterestPercent:  interestPct,
	}
}

// buildYearlyBreakdown walks the compound growth one year at a time
// and produces a per-year schedule. Returns nil if tenure < 1 year.
//
// Each row reflects the interest earned during that year and the
// closing balance after that year — computed from the exact compound
// formula (not a simple r*n approximation), so that the final row's
// closing balance reconciles with the top-level TotalAmount.
func buildYearlyBreakdown(principal, r float64, n int, timeYears float64) []YearlyRow {
	if timeYears < 1 {
		return nil
	}
	fullYears := int(math.Floor(timeYears))
	if fullYears <= 0 {
		return nil
	}

	rows := make([]YearlyRow, 0, fullYears)
	opening := principal

	for y := 1; y <= fullYears; y++ {
		closing := roundTo(
			principal*math.Pow(1+r/float64(n), float64(n)*float64(y)),
			interestRoundingPlaces,
		)
		interest := roundTo(closing-opening, interestRoundingPlaces)

		rows = append(rows, YearlyRow{
			Year:           y,
			OpeningBalance: roundTo(opening, interestRoundingPlaces),
			InterestEarned: interest,
			ClosingBalance: closing,
		})
		opening = closing
	}
	return rows
}