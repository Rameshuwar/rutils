package converter

import (
	"math"
	"testing"
)

func almostEqualInterest(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ------------------------------------------------------------
// SIMPLE INTEREST — happy paths
// ------------------------------------------------------------

func TestSimpleInterest_Basic(t *testing.T) {
	// P=100000, R=8.5%, T=5 years
	// SI = (100000 × 8.5 × 5) / 100 = 42500
	// A  = 142500
	res, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal:          100000,
		AnnualInterestRate: 8.5,
		Time:               5,
		TimeUnit:           "years",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !almostEqualInterest(res.Interest, 42500, 0.01) {
		t.Fatalf("expected interest 42500, got %.2f", res.Interest)
	}
	if !almostEqualInterest(res.TotalAmount, 142500, 0.01) {
		t.Fatalf("expected total 142500, got %.2f", res.TotalAmount)
	}
	if !almostEqualInterest(res.TimeYears, 5, 1e-9) {
		t.Fatalf("expected timeYears 5, got %v", res.TimeYears)
	}
	if len(res.Steps) == 0 {
		t.Fatal("expected non-empty steps")
	}
}

func TestSimpleInterest_MonthTenure(t *testing.T) {
	// P=12000, R=10%, T=6 months = 0.5 years
	// SI = (12000 × 10 × 0.5) / 100 = 600
	res, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal:          12000,
		AnnualInterestRate: 10,
		Time:               6,
		TimeUnit:           "months",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualInterest(res.TimeYears, 0.5, 1e-9) {
		t.Fatalf("expected timeYears 0.5, got %v", res.TimeYears)
	}
	if !almostEqualInterest(res.Interest, 600, 0.01) {
		t.Fatalf("expected interest 600, got %.2f", res.Interest)
	}
}

func TestSimpleInterest_DayTenure(t *testing.T) {
	// P=100000, R=12%, T=365 days = 1 year
	// SI = (100000 × 12 × 1) / 100 = 12000
	res, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal:          100000,
		AnnualInterestRate: 12,
		Time:               365,
		TimeUnit:           "days",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualInterest(res.Interest, 12000, 0.01) {
		t.Fatalf("expected interest 12000, got %.2f", res.Interest)
	}
}

func TestSimpleInterest_ZeroRate(t *testing.T) {
	res, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal:          100000,
		AnnualInterestRate: 0,
		Time:               5,
		TimeUnit:           "years",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Interest != 0 {
		t.Fatalf("expected zero interest, got %.2f", res.Interest)
	}
	if res.TotalAmount != 100000 {
		t.Fatalf("expected total 100000, got %.2f", res.TotalAmount)
	}
}

func TestSimpleInterest_BreakdownSumsTo100(t *testing.T) {
	res, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal:          100000,
		AnnualInterestRate: 8.5,
		Time:               5,
		TimeUnit:           "years",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum := res.Breakdown.PrincipalPercent + res.Breakdown.InterestPercent
	if !almostEqualInterest(sum, 100, 0.01) {
		t.Fatalf("breakdown percentages should sum to 100, got %.4f", sum)
	}
}

// ------------------------------------------------------------
// SIMPLE INTEREST — validation
// ------------------------------------------------------------

func TestSimpleInterest_NegativePrincipal(t *testing.T) {
	_, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal: -1, AnnualInterestRate: 8, Time: 5, TimeUnit: "years",
	})
	if err == nil {
		t.Fatal("expected error for negative principal")
	}
}

func TestSimpleInterest_ZeroPrincipal(t *testing.T) {
	_, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal: 0, AnnualInterestRate: 8, Time: 5, TimeUnit: "years",
	})
	if err == nil {
		t.Fatal("expected error for zero principal")
	}
}

func TestSimpleInterest_NegativeRate(t *testing.T) {
	_, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal: 100000, AnnualInterestRate: -1, Time: 5, TimeUnit: "years",
	})
	if err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestSimpleInterest_ZeroTime(t *testing.T) {
	_, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal: 100000, AnnualInterestRate: 8, Time: 0, TimeUnit: "years",
	})
	if err == nil {
		t.Fatal("expected error for zero time")
	}
}

func TestSimpleInterest_InvalidTimeUnit(t *testing.T) {
	_, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal: 100000, AnnualInterestRate: 8, Time: 5, TimeUnit: "weeks",
	})
	if err == nil {
		t.Fatal("expected error for invalid timeUnit")
	}
}

func TestSimpleInterest_TimeTooLong(t *testing.T) {
	_, err := CalculateSimpleInterest(SimpleInterestRequest{
		Principal: 100000, AnnualInterestRate: 8, Time: 200, TimeUnit: "years",
	})
	if err == nil {
		t.Fatal("expected error for time exceeding maximum")
	}
}

// ------------------------------------------------------------
// COMPOUND INTEREST — happy paths
// ------------------------------------------------------------

func TestCompoundInterest_YearlyBasic(t *testing.T) {
	// P=100000, R=8.5%, T=5y, yearly
	// A = 100000 × 1.085^5 = 150365.57 (approx)
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   8.5,
		Time:                 5,
		TimeUnit:             "years",
		CompoundingFrequency: "yearly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !almostEqualInterest(res.TotalAmount, 150365.57, 0.5) {
		t.Fatalf("expected total вүҲ 150365.57, got %.2f", res.TotalAmount)
	}
	if !almostEqualInterest(res.Interest, 50365.57, 0.5) {
		t.Fatalf("expected interest вүҲ 50365.57, got %.2f", res.Interest)
	}
	if res.CompoundsPerYear != 1 {
		t.Fatalf("expected compoundsPerYear 1, got %d", res.CompoundsPerYear)
	}
}

func TestCompoundInterest_Quarterly(t *testing.T) {
	// P=100000, R=8%, T=1y, quarterly n=4
	// A = 100000 × (1 + 0.08/4)^4 = 108243.22
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   8,
		Time:                 1,
		TimeUnit:             "years",
		CompoundingFrequency: "quarterly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualInterest(res.TotalAmount, 108243.22, 0.05) {
		t.Fatalf("expected total вүҲ 108243.22, got %.2f", res.TotalAmount)
	}
	if res.CompoundsPerYear != 4 {
		t.Fatalf("expected compoundsPerYear 4, got %d", res.CompoundsPerYear)
	}
}

func TestCompoundInterest_Monthly(t *testing.T) {
	// P=100000, R=12%, T=1y, monthly n=12
	// A = 100000 × (1 + 0.12/12)^12 = 112682.50
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   12,
		Time:                 1,
		TimeUnit:             "years",
		CompoundingFrequency: "monthly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualInterest(res.TotalAmount, 112682.50, 0.5) {
		t.Fatalf("expected total вүҲ 112682.50, got %.2f", res.TotalAmount)
	}
	if res.CompoundsPerYear != 12 {
		t.Fatalf("expected compoundsPerYear 12, got %d", res.CompoundsPerYear)
	}
}

func TestCompoundInterest_ZeroRate(t *testing.T) {
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   0,
		Time:                 5,
		TimeUnit:             "years",
		CompoundingFrequency: "yearly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Interest != 0 {
		t.Fatalf("expected zero interest, got %.2f", res.Interest)
	}
	if res.TotalAmount != 100000 {
		t.Fatalf("expected total 100000, got %.2f", res.TotalAmount)
	}
}

func TestCompoundInterest_DefaultFrequency(t *testing.T) {
	// Empty frequency should default to "yearly".
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:          100000,
		AnnualInterestRate: 8,
		Time:               1,
		TimeUnit:           "years",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CompoundingFrequency != "yearly" {
		t.Fatalf("expected default frequency yearly, got %s", res.CompoundingFrequency)
	}
}

func TestCompoundInterest_YearlyBreakdownLength(t *testing.T) {
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   8,
		Time:                 5,
		TimeUnit:             "years",
		CompoundingFrequency: "yearly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.YearlyBreakdown) != 5 {
		t.Fatalf("expected 5 yearly rows, got %d", len(res.YearlyBreakdown))
	}
	// Final closing balance must reconcile with total.
	last := res.YearlyBreakdown[len(res.YearlyBreakdown)-1]
	if !almostEqualInterest(last.ClosingBalance, res.TotalAmount, 0.5) {
		t.Fatalf("final closing %v != total %v", last.ClosingBalance, res.TotalAmount)
	}
}

func TestCompoundInterest_BreakdownSumsTo100(t *testing.T) {
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   8,
		Time:                 5,
		TimeUnit:             "years",
		CompoundingFrequency: "yearly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum := res.Breakdown.PrincipalPercent + res.Breakdown.InterestPercent
	if !almostEqualInterest(sum, 100, 0.01) {
		t.Fatalf("breakdown percentages should sum to 100, got %.4f", sum)
	}
}

func TestCompoundInterest_EffectiveAnnualRate(t *testing.T) {
	// 12% monthly → EAR = (1 + 0.12/12)^12 - 1 = 0.126825 → 12.6825%
	res, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   12,
		Time:                 1,
		TimeUnit:             "years",
		CompoundingFrequency: "monthly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualInterest(res.EffectiveAnnualRate, 12.6825, 0.01) {
		t.Fatalf("expected EAR вүҲ 12.6825, got %.4f", res.EffectiveAnnualRate)
	}
}

// ------------------------------------------------------------
// COMPOUND INTEREST — validation
// ------------------------------------------------------------

func TestCompoundInterest_InvalidFrequency(t *testing.T) {
	_, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   8,
		Time:                 5,
		TimeUnit:             "years",
		CompoundingFrequency: "hourly",
	})
	if err == nil {
		t.Fatal("expected error for invalid compoundingFrequency")
	}
}

func TestCompoundInterest_RejectsDaysTenure(t *testing.T) {
	_, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   8,
		Time:                 365,
		TimeUnit:             "days",
		CompoundingFrequency: "daily",
	})
	if err == nil {
		t.Fatal("expected error for day-based tenure")
	}
}

func TestCompoundInterest_NegativePrincipal(t *testing.T) {
	_, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            -1,
		AnnualInterestRate:   8,
		Time:                 5,
		TimeUnit:             "years",
		CompoundingFrequency: "yearly",
	})
	if err == nil {
		t.Fatal("expected error for negative principal")
	}
}

func TestCompoundInterest_RateTooHigh(t *testing.T) {
	_, err := CalculateCompoundInterest(CompoundInterestRequest{
		Principal:            100000,
		AnnualInterestRate:   150,
		Time:                 5,
		TimeUnit:             "years",
		CompoundingFrequency: "yearly",
	})
	if err == nil {
		t.Fatal("expected error for rate > 100%")
	}
}
