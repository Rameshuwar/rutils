package converter

import (
	"math"
	"strings"
	"testing"
)

func almostEqualLT(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ============================================================
// HAPPY PATHS
// ============================================================

func TestLoanTenure_StandardCase(t *testing.T) {
	// ₹5,00,000 @ 10.5% annual, paying ₹15,000/month
	// Expected: ~41 months, total interest ~₹1,15,000
	res, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          500000,
		AnnualInterestRate: 10.5,
		MonthlyPayment:     15000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TenureMonths < 40 || res.TenureMonths > 42 {
		t.Fatalf("expected tenure 40-42 months, got %d", res.TenureMonths)
	}
	if res.TotalPayment <= res.Principal {
		t.Fatalf("total payment (%.2f) must exceed principal (%.2f)", res.TotalPayment, res.Principal)
	}
	if res.TotalInterest <= 0 {
		t.Fatalf("total interest must be positive, got %.2f", res.TotalInterest)
	}
	if res.EMI != 15000 {
		t.Fatalf("EMI should echo the input 15000, got %.2f", res.EMI)
	}
	if len(res.Amortization) != res.TenureMonths {
		t.Fatalf("amortization length (%d) must match tenureMonths (%d)",
			len(res.Amortization), res.TenureMonths)
	}

	// Last row must close at exactly zero.
	last := res.Amortization[len(res.Amortization)-1]
	if last.ClosingBalance != 0 {
		t.Fatalf("final closing balance must be 0, got %.2f", last.ClosingBalance)
	}
}

func TestLoanTenure_ZeroInterest(t *testing.T) {
	// ₹1,20,000 @ 0%, paying ₹10,000/month → exactly 12 months
	res, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          120000,
		AnnualInterestRate: 0,
		MonthlyPayment:     10000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TenureMonths != 12 {
		t.Fatalf("expected 12 months, got %d", res.TenureMonths)
	}
	if res.TotalInterest != 0 {
		t.Fatalf("expected zero interest, got %.2f", res.TotalInterest)
	}
	if res.TotalPayment != 120000 {
		t.Fatalf("expected total payment 120000, got %.2f", res.TotalPayment)
	}
	if res.Breakdown.PrincipalPercent != 100 {
		t.Fatalf("expected 100%% principal, got %.2f", res.Breakdown.PrincipalPercent)
	}
}

func TestLoanTenure_SingleMonthPayoff(t *testing.T) {
	// Principal + one month's interest paid in a single month
	res, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          10000,
		AnnualInterestRate: 12,
		MonthlyPayment:     10100, // 10000 principal + 100 interest
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TenureMonths != 1 {
		t.Fatalf("expected 1 month, got %d", res.TenureMonths)
	}
}

func TestLoanTenure_LongTermHomeLoan(t *testing.T) {
	// ₹50,00,000 @ 8.5% paying ₹40,000/month
	res, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          5000000,
		AnnualInterestRate: 8.5,
		MonthlyPayment:     40000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The exact tenure for ₹50,00,000 @ 8.5% with a ₹40,000/month
	// payment is 307 months (~25.6 years). This matches independent
	// loan calculators, and we assert it to within a 3-month window
	// to allow for rounding conventions.
	if res.TenureMonths < 305 || res.TenureMonths > 309 {
		t.Fatalf("expected tenure 305-309 months, got %d", res.TenureMonths)
	}
	if res.TenureMonths > MaxTenureMonths {
		t.Fatalf("tenure must not exceed %d months", MaxTenureMonths)
	}
}

// ============================================================
// RECONCILIATION
// ============================================================

func TestLoanTenure_TotalsReconcile(t *testing.T) {
	res, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          750000,
		AnnualInterestRate: 9.25,
		MonthlyPayment:     20000,
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

	if !almostEqualLT(sumPrincipal, res.Principal, 0.05) {
		t.Fatalf("principal sum mismatch: %.2f vs %.2f", sumPrincipal, res.Principal)
	}
	if !almostEqualLT(sumInterest, res.TotalInterest, 0.05) {
		t.Fatalf("interest sum mismatch: %.2f vs %.2f", sumInterest, res.TotalInterest)
	}
	if !almostEqualLT(sumPaid, res.TotalPayment, 0.05) {
		t.Fatalf("payment sum mismatch: %.2f vs %.2f", sumPaid, res.TotalPayment)
	}
}

func TestLoanTenure_BreakdownSumsTo100(t *testing.T) {
	res, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          300000,
		AnnualInterestRate: 14,
		MonthlyPayment:     15000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sum := res.Breakdown.PrincipalPercent + res.Breakdown.InterestPercent
	if !almostEqualLT(sum, 100, 0.01) {
		t.Fatalf("breakdown percentages must sum to 100, got %.4f", sum)
	}
}

func TestLoanTenure_OpeningBalanceChains(t *testing.T) {
	res, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          400000,
		AnnualInterestRate: 11,
		MonthlyPayment:     18000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Row 1 opening == principal
	if !almostEqualLT(res.Amortization[0].OpeningBalance, res.Principal, 0.01) {
		t.Fatalf("first opening balance must equal principal")
	}

	// Every subsequent row opens where the previous row closed.
	for i := 1; i < len(res.Amortization); i++ {
		prev := res.Amortization[i-1].ClosingBalance
		cur := res.Amortization[i].OpeningBalance
		if !almostEqualLT(prev, cur, 0.02) {
			t.Fatalf("row %d opening (%.2f) != row %d closing (%.2f)",
				i+1, cur, i, prev)
		}
	}
}

// ============================================================
// VALIDATION ERRORS
// ============================================================

func TestLoanTenure_PaymentTooLow(t *testing.T) {
	// 12% annual on 1,00,000 → monthly interest is 1000
	// Paying 500/month never touches the principal.
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          100000,
		AnnualInterestRate: 12,
		MonthlyPayment:     500,
	})
	if err == nil {
		t.Fatal("expected error when payment can't cover monthly interest")
	}
	if !errContains(err.Error(), "too low") {
		t.Fatalf("expected 'too low' error, got %q", err.Error())
	}
}

func TestLoanTenure_PaymentExactlyEqualsInterest(t *testing.T) {
	// Boundary case: EMI == P × r means loan never amortizes.
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          100000,
		AnnualInterestRate: 12, // monthly = 1000
		MonthlyPayment:     1000,
	})
	if err == nil {
		t.Fatal("expected error when payment equals monthly interest")
	}
}

func TestLoanTenure_NegativePrincipal(t *testing.T) {
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          -1,
		AnnualInterestRate: 10,
		MonthlyPayment:     5000,
	})
	if err == nil {
		t.Fatal("expected error for negative principal")
	}
}

func TestLoanTenure_ZeroPrincipal(t *testing.T) {
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          0,
		AnnualInterestRate: 10,
		MonthlyPayment:     5000,
	})
	if err == nil {
		t.Fatal("expected error for zero principal")
	}
}

func TestLoanTenure_NegativeRate(t *testing.T) {
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          100000,
		AnnualInterestRate: -1,
		MonthlyPayment:     5000,
	})
	if err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestLoanTenure_RateTooHigh(t *testing.T) {
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          100000,
		AnnualInterestRate: 150,
		MonthlyPayment:     50000,
	})
	if err == nil {
		t.Fatal("expected error for rate > 100%")
	}
}

func TestLoanTenure_ZeroPayment(t *testing.T) {
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          100000,
		AnnualInterestRate: 10,
		MonthlyPayment:     0,
	})
	if err == nil {
		t.Fatal("expected error for zero payment")
	}
}

func TestLoanTenure_NegativePayment(t *testing.T) {
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          100000,
		AnnualInterestRate: 10,
		MonthlyPayment:     -5000,
	})
	if err == nil {
		t.Fatal("expected error for negative payment")
	}
}

func TestLoanTenure_TenureExceedsMax(t *testing.T) {
	// Very small payment relative to a large loan → tenure > 600 months.
	_, err := CalculateLoanTenure(LoanTenureRequest{
		Principal:          100000000, // 10 crore
		AnnualInterestRate: 15,
		MonthlyPayment:     1000, // absurdly small
	})
	if err == nil {
		t.Fatal("expected error when tenure exceeds 600 months")
	}
}

// ============================================================
// HELPER: contains is defined in percentage_test.go of this package
// ============================================================

// errContains reports whether the given error message contains the
// given substring. Named distinctly from the production slice-based
// `contains` helper in loan_tenure.go to avoid a package-level name
// collision.
func errContains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}