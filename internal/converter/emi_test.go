package converter

import (
	"math"
	"testing"
)

func almostEqualEMI(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ------------------------------------------------------------
// Core formula
// ------------------------------------------------------------

func TestCalculateEMI_StandardLoan(t *testing.T) {
	// ₹5,00,000 @ 10.5% annual for 60 months.
	// Expected EMI ≈ ₹10,747.28 (matches standard online calculators).
	res, err := CalculateEMI(EMIRequest{
		Principal:          500000,
		AnnualInterestRate: 10.5,
		Tenure:             60,
		TenureUnit:         "months",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !almostEqualEMI(res.EMI, 10747.28, 0.5) {
		t.Fatalf("expected EMI ≈ 10747.28, got %.2f", res.EMI)
	}
	if res.TenureMonths != 60 {
		t.Fatalf("expected 60 months, got %d", res.TenureMonths)
	}
	if len(res.Amortization) != 60 {
		t.Fatalf("expected 60 amortization rows, got %d", len(res.Amortization))
	}

	// Last row must close at exactly zero.
	last := res.Amortization[len(res.Amortization)-1]
	if last.ClosingBalance != 0 {
		t.Fatalf("expected final closing balance 0, got %.2f", last.ClosingBalance)
	}
}

func TestCalculateEMI_YearsTenure(t *testing.T) {
	// 5 years should behave identically to 60 months.
	monthsRes, _ := CalculateEMI(EMIRequest{
		Principal:          500000,
		AnnualInterestRate: 10.5,
		Tenure:             60,
		TenureUnit:         "months",
	})
	yearsRes, err := CalculateEMI(EMIRequest{
		Principal:          500000,
		AnnualInterestRate: 10.5,
		Tenure:             5,
		TenureUnit:         "years",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if monthsRes.EMI != yearsRes.EMI {
		t.Fatalf("months and years gave different EMI: %.2f vs %.2f", monthsRes.EMI, yearsRes.EMI)
	}
	if yearsRes.TenureMonths != 60 {
		t.Fatalf("expected 60 months, got %d", yearsRes.TenureMonths)
	}
}

func TestCalculateEMI_ZeroInterest(t *testing.T) {
	// Interest-free loan: EMI = principal / months.
	res, err := CalculateEMI(EMIRequest{
		Principal:          120000,
		AnnualInterestRate: 0,
		Tenure:             12,
		TenureUnit:         "months",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.EMI != 10000 {
		t.Fatalf("expected EMI 10000, got %.2f", res.EMI)
	}
	if res.TotalInterest != 0 {
		t.Fatalf("expected zero total interest, got %.2f", res.TotalInterest)
	}
	if res.TotalPayment != 120000 {
		t.Fatalf("expected total payment 120000, got %.2f", res.TotalPayment)
	}
}

func TestCalculateEMI_SingleMonth(t *testing.T) {
	// One-month loan: EMI = principal + one month's interest.
	res, err := CalculateEMI(EMIRequest{
		Principal:          10000,
		AnnualInterestRate: 12,
		Tenure:             1,
		TenureUnit:         "months",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1% monthly interest → 10100
	if !almostEqualEMI(res.EMI, 10100, 0.01) {
		t.Fatalf("expected EMI ≈ 10100, got %.2f", res.EMI)
	}
}

func TestCalculateEMI_LongTenure(t *testing.T) {
	// 30-year home loan.
	res, err := CalculateEMI(EMIRequest{
		Principal:          5000000,
		AnnualInterestRate: 8.5,
		Tenure:             30,
		TenureUnit:         "years",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TenureMonths != 360 {
		t.Fatalf("expected 360 months, got %d", res.TenureMonths)
	}
	if len(res.Amortization) != 360 {
		t.Fatalf("expected 360 amortization rows, got %d", len(res.Amortization))
	}
}

// ------------------------------------------------------------
// Reconciliation / rounding
// ------------------------------------------------------------

func TestCalculateEMI_TotalsReconcile(t *testing.T) {
	res, err := CalculateEMI(EMIRequest{
		Principal:          750000,
		AnnualInterestRate: 9.25,
		Tenure:             84,
		TenureUnit:         "months",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sumPrincipal, sumInterest, sumPaid float64
	for _, row := range res.Amortization {
		sumPrincipal += row.PrincipalPaid
		sumInterest += row.InterestPaid
		sumPaid += row.TotalPaid
	}

	if !almostEqualEMI(sumPrincipal, res.Principal, 0.05) {
		t.Fatalf("principal sum mismatch: %.2f vs %.2f", sumPrincipal, res.Principal)
	}
	if !almostEqualEMI(sumInterest, res.TotalInterest, 0.05) {
		t.Fatalf("interest sum mismatch: %.2f vs %.2f", sumInterest, res.TotalInterest)
	}
	if !almostEqualEMI(sumPaid, res.TotalPayment, 0.05) {
		t.Fatalf("payment sum mismatch: %.2f vs %.2f", sumPaid, res.TotalPayment)
	}
}

func TestCalculateEMI_BreakdownSumsTo100(t *testing.T) {
	res, err := CalculateEMI(EMIRequest{
		Principal:          300000,
		AnnualInterestRate: 14,
		Tenure:             24,
		TenureUnit:         "months",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum := res.Breakdown.PrincipalPercent + res.Breakdown.InterestPercent
	if !almostEqualEMI(sum, 100, 0.01) {
		t.Fatalf("breakdown percentages should sum to 100, got %.4f", sum)
	}
}

// ------------------------------------------------------------
// Validation errors
// ------------------------------------------------------------

func TestCalculateEMI_NegativePrincipal(t *testing.T) {
	_, err := CalculateEMI(EMIRequest{
		Principal:          -1,
		AnnualInterestRate: 10,
		Tenure:             12,
		TenureUnit:         "months",
	})
	if err == nil {
		t.Fatal("expected error for negative principal")
	}
}

func TestCalculateEMI_ZeroPrincipal(t *testing.T) {
	_, err := CalculateEMI(EMIRequest{
		Principal:          0,
		AnnualInterestRate: 10,
		Tenure:             12,
		TenureUnit:         "months",
	})
	if err == nil {
		t.Fatal("expected error for zero principal")
	}
}

func TestCalculateEMI_NegativeRate(t *testing.T) {
	_, err := CalculateEMI(EMIRequest{
		Principal:          100000,
		AnnualInterestRate: -1,
		Tenure:             12,
		TenureUnit:         "months",
	})
	if err == nil {
		t.Fatal("expected error for negative interest rate")
	}
}

func TestCalculateEMI_RateTooHigh(t *testing.T) {
	_, err := CalculateEMI(EMIRequest{
		Principal:          100000,
		AnnualInterestRate: 150,
		Tenure:             12,
		TenureUnit:         "months",
	})
	if err == nil {
		t.Fatal("expected error for rate > 100%")
	}
}

func TestCalculateEMI_ZeroTenure(t *testing.T) {
	_, err := CalculateEMI(EMIRequest{
		Principal:          100000,
		AnnualInterestRate: 10,
		Tenure:             0,
		TenureUnit:         "months",
	})
	if err == nil {
		t.Fatal("expected error for zero tenure")
	}
}

func TestCalculateEMI_InvalidUnit(t *testing.T) {
	_, err := CalculateEMI(EMIRequest{
		Principal:          100000,
		AnnualInterestRate: 10,
		Tenure:             12,
		TenureUnit:         "weeks",
	})
	if err == nil {
		t.Fatal("expected error for invalid tenureUnit")
	}
}

func TestCalculateEMI_TenureTooLong(t *testing.T) {
	_, err := CalculateEMI(EMIRequest{
		Principal:          100000,
		AnnualInterestRate: 10,
		Tenure:             60,
		TenureUnit:         "years", // 720 months > 600
	})
	if err == nil {
		t.Fatal("expected error for tenure exceeding maximum")
	}
}
