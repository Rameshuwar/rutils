package converter

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestPercentOf(t *testing.T) {
	res, err := CalculatePercentage("percent_of", 15, 200, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 30) {
		t.Fatalf("expected 30, got %v", res.Result)
	}
}

func TestWhatPercent(t *testing.T) {
	res, err := CalculatePercentage("what_percent", 30, 200, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 15) {
		t.Fatalf("expected 15, got %v", res.Result)
	}
}

func TestWhatPercent_DivideByZero(t *testing.T) {
	_, err := CalculatePercentage("what_percent", 30, 0, 0, false)
	if err == nil {
		t.Fatal("expected error for zero whole")
	}
}

func TestPercentChange_Increase(t *testing.T) {
	res, err := CalculatePercentage("percent_change", 100, 150, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 50) {
		t.Fatalf("expected 50, got %v", res.Result)
	}
	if res.Extra["direction"] != "increase" {
		t.Fatalf("expected direction increase, got %v", res.Extra["direction"])
	}
}

func TestPercentChange_Decrease(t *testing.T) {
	res, err := CalculatePercentage("percent_change", 200, 150, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, -25) {
		t.Fatalf("expected -25, got %v", res.Result)
	}
	if res.Extra["direction"] != "decrease" {
		t.Fatalf("expected direction decrease, got %v", res.Extra["direction"])
	}
}

func TestPercentIncrease(t *testing.T) {
	res, err := CalculatePercentage("percent_increase", 200, 15, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 230) {
		t.Fatalf("expected 230, got %v", res.Result)
	}
}

func TestPercentDecrease(t *testing.T) {
	res, err := CalculatePercentage("percent_decrease", 200, 15, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 170) {
		t.Fatalf("expected 170, got %v", res.Result)
	}
}

func TestReversePercent(t *testing.T) {
	res, err := CalculatePercentage("reverse_percent", 230, 15, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 200) {
		t.Fatalf("expected 200, got %v", res.Result)
	}
}

func TestIsPercentOfWhat(t *testing.T) {
	res, err := CalculatePercentage("is_percent_of_what", 30, 15, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 200) {
		t.Fatalf("expected 200, got %v", res.Result)
	}
}

func TestPercentDifference(t *testing.T) {
	res, err := CalculatePercentage("percent_difference", 100, 150, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// |100-150| / ((100+150)/2) * 100 = 50/125*100 = 40
	if !almostEqual(res.Result, 40) {
		t.Fatalf("expected 40, got %v", res.Result)
	}
}

func TestAddPercentPoints(t *testing.T) {
	res, err := CalculatePercentage("add_percent_points", 5, 3, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 8) {
		t.Fatalf("expected 8, got %v", res.Result)
	}
}

func TestSubtractPercentPoints(t *testing.T) {
	res, err := CalculatePercentage("subtract_percent_points", 5, 3, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 2) {
		t.Fatalf("expected 2, got %v", res.Result)
	}
}

func TestDiscount(t *testing.T) {
	res, err := CalculatePercentage("discount", 500, 20, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 400) {
		t.Fatalf("expected 400, got %v", res.Result)
	}
	if !almostEqual(res.Extra["savings"].(float64), 100) {
		t.Fatalf("expected savings 100, got %v", res.Extra["savings"])
	}
}

func TestMarkup(t *testing.T) {
	res, err := CalculatePercentage("markup", 100, 20, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 120) {
		t.Fatalf("expected 120, got %v", res.Result)
	}
}

func TestProfitLoss_Profit(t *testing.T) {
	res, err := CalculatePercentage("profit_loss", 100, 120, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 20) {
		t.Fatalf("expected 20, got %v", res.Result)
	}
	if res.Extra["label"] != "profit" {
		t.Fatalf("expected profit, got %v", res.Extra["label"])
	}
}

func TestProfitLoss_Loss(t *testing.T) {
	res, err := CalculatePercentage("profit_loss", 100, 80, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, -20) {
		t.Fatalf("expected -20, got %v", res.Result)
	}
	if res.Extra["label"] != "loss" {
		t.Fatalf("expected loss, got %v", res.Extra["label"])
	}
}

func TestPercentToFraction(t *testing.T) {
	cases := map[float64]string{
		25:   "1/4",
		50:   "1/2",
		12.5: "1/8",
		30:   "3/10",
		100:  "1",
		0:    "0",
	}
	for pct, want := range cases {
		res, err := CalculatePercentage("percent_to_fraction", pct, 0, 0, false)
		if err != nil {
			t.Fatalf("unexpected error for %v: %v", pct, err)
		}
		if res.Formatted != want {
			t.Fatalf("percent %v: expected %q, got %q", pct, want, res.Formatted)
		}
	}
}

func TestFractionToPercent(t *testing.T) {
	res, err := CalculatePercentage("fraction_to_percent", 1, 4, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 25) {
		t.Fatalf("expected 25, got %v", res.Result)
	}
}

func TestFractionToPercent_DivideByZero(t *testing.T) {
	_, err := CalculatePercentage("fraction_to_percent", 1, 0, 0, false)
	if err == nil {
		t.Fatal("expected error for zero denominator")
	}
}

func TestDecimalToPercent(t *testing.T) {
	res, err := CalculatePercentage("decimal_to_percent", 0.25, 0, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 25) {
		t.Fatalf("expected 25, got %v", res.Result)
	}
}

func TestCompoundPercent(t *testing.T) {
	// 1000 base, 10% per year, 3 years в†’ 1000 * 1.1^3 = 1331
	res, err := CalculatePercentage("compound_percent", 1000, 10, 3, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 1331) {
		t.Fatalf("expected 1331, got %v", res.Result)
	}
}

func TestCompoundPercent_MissingPeriods(t *testing.T) {
	_, err := CalculatePercentage("compound_percent", 1000, 10, 0, false)
	if err == nil {
		t.Fatal("expected error for missing value3")
	}
}

func TestMarksPercentage(t *testing.T) {
	res, err := CalculatePercentage("marks_percentage", 85, 100, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 85) {
		t.Fatalf("expected 85, got %v", res.Result)
	}
}

func TestCgpaToPercent(t *testing.T) {
	res, err := CalculatePercentage("cgpa_to_percent", 8.0, 0, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqual(res.Result, 76) {
		t.Fatalf("expected 76, got %v", res.Result)
	}
}

func TestUnsupportedOperation(t *testing.T) {
	_, err := CalculatePercentage("magic_trick", 1, 2, 0, false)
	if err == nil {
		t.Fatal("expected error for unsupported operation")
	}
}

func TestEmptyOperation(t *testing.T) {
	_, err := CalculatePercentage("", 1, 2, 0, false)
	if err == nil {
		t.Fatal("expected error for empty operation")
	}
}