package converter

import (
	"math"
	"testing"
)

func almostEqualTax(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ------------------------------------------------------------
// add_tax
// ------------------------------------------------------------

func TestTax_AddTax_Standard(t *testing.T) {
	res, err := CalculateTax(TaxRequest{
		Mode:    "add_tax",
		Amount:  1000,
		TaxRate: 18,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NetAmount != 1000 {
		t.Fatalf("expected net 1000, got %v", res.NetAmount)
	}
	if res.TaxAmount != 180 {
		t.Fatalf("expected tax 180, got %v", res.TaxAmount)
	}
	if res.GrossAmount != 1180 {
		t.Fatalf("expected gross 1180, got %v", res.GrossAmount)
	}
	if res.Formatted != "1180" {
		t.Fatalf("expected formatted 1180, got %q", res.Formatted)
	}
}

func TestTax_AddTax_ZeroRate(t *testing.T) {
	res, err := CalculateTax(TaxRequest{Mode: "add_tax", Amount: 500, TaxRate: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GrossAmount != 500 || res.TaxAmount != 0 {
		t.Fatalf("expected no tax, got gross %v tax %v", res.GrossAmount, res.TaxAmount)
	}
}

func TestTax_AddTax_NegativeAmount(t *testing.T) {
	_, err := CalculateTax(TaxRequest{Mode: "add_tax", Amount: -1, TaxRate: 18})
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestTax_AddTax_RateTooHigh(t *testing.T) {
	_, err := CalculateTax(TaxRequest{Mode: "add_tax", Amount: 100, TaxRate: 150})
	if err == nil {
		t.Fatal("expected error for rate > 100")
	}
}

// ------------------------------------------------------------
// remove_tax
// ------------------------------------------------------------

func TestTax_RemoveTax_Standard(t *testing.T) {
	res, err := CalculateTax(TaxRequest{Mode: "remove_tax", Amount: 1180, TaxRate: 18})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualTax(res.NetAmount, 1000, 0.01) {
		t.Fatalf("expected net ≈ 1000, got %v", res.NetAmount)
	}
	if !almostEqualTax(res.TaxAmount, 180, 0.01) {
		t.Fatalf("expected tax ≈ 180, got %v", res.TaxAmount)
	}
}

func TestTax_RemoveTax_ZeroRate(t *testing.T) {
	_, err := CalculateTax(TaxRequest{Mode: "remove_tax", Amount: 1000, TaxRate: 0})
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
}

// ------------------------------------------------------------
// find_rate
// ------------------------------------------------------------

func TestTax_FindRate_Standard(t *testing.T) {
	res, err := CalculateTax(TaxRequest{
		Mode:      "find_rate",
		NetAmount: 1000,
		GrossAmt:  1180,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualTax(res.TaxRate, 18, 0.01) {
		t.Fatalf("expected rate 18, got %v", res.TaxRate)
	}
	if !almostEqualTax(res.TaxAmount, 180, 0.01) {
		t.Fatalf("expected tax 180, got %v", res.TaxAmount)
	}
}

func TestTax_FindRate_GrossLessThanNet(t *testing.T) {
	_, err := CalculateTax(TaxRequest{Mode: "find_rate", NetAmount: 1000, GrossAmt: 900})
	if err == nil {
		t.Fatal("expected error when gross < net")
	}
}

// ------------------------------------------------------------
// split_gst
// ------------------------------------------------------------

func TestTax_SplitGST_CGSTSGST(t *testing.T) {
	res, err := CalculateTax(TaxRequest{
		Mode:    "split_gst",
		Amount:  1180,
		TaxRate: 18,
		TaxType: "cgst_sgst",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualTax(res.NetAmount, 1000, 0.01) {
		t.Fatalf("expected net 1000, got %v", res.NetAmount)
	}
	if !almostEqualTax(res.TaxAmount, 180, 0.01) {
		t.Fatalf("expected tax 180, got %v", res.TaxAmount)
	}
	cgst := res.Extra["cgst"].(float64)
	sgst := res.Extra["sgst"].(float64)
	if !almostEqualTax(cgst, 90, 0.01) || !almostEqualTax(sgst, 90, 0.01) {
		t.Fatalf("expected cgst=90 sgst=90, got cgst=%v sgst=%v", cgst, sgst)
	}
}

func TestTax_SplitGST_IGST(t *testing.T) {
	res, err := CalculateTax(TaxRequest{
		Mode:    "split_gst",
		Amount:  1180,
		TaxRate: 18,
		TaxType: "igst",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	igst := res.Extra["igst"].(float64)
	if !almostEqualTax(igst, 180, 0.01) {
		t.Fatalf("expected igst 180, got %v", igst)
	}
}

func TestTax_SplitGST_InvalidTaxType(t *testing.T) {
	_, err := CalculateTax(TaxRequest{
		Mode: "split_gst", Amount: 1180, TaxRate: 18, TaxType: "vat",
	})
	if err == nil {
		t.Fatal("expected error for invalid taxType")
	}
}

func TestTax_SplitGST_OddTaxAmount_SumsExactly(t *testing.T) {
	// 18% of 999 / 2 produces a fractional CGST; SGST must absorb the
	// rounding drift so CGST + SGST == total GST exactly.
	res, err := CalculateTax(TaxRequest{
		Mode: "split_gst", Amount: 999, TaxRate: 18, TaxType: "cgst_sgst",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cgst := res.Extra["cgst"].(float64)
	sgst := res.Extra["sgst"].(float64)
	if !almostEqualTax(cgst+sgst, res.TaxAmount, 0.01) {
		t.Fatalf("cgst+sgst (%v) != total tax (%v)", cgst+sgst, res.TaxAmount)
	}
}

// ------------------------------------------------------------
// reverse_gst
// ------------------------------------------------------------

func TestTax_ReverseGST_Standard(t *testing.T) {
	res, err := CalculateTax(TaxRequest{
		Mode: "reverse_gst", TaxPaid: 180, TaxRate: 18,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualTax(res.NetAmount, 1000, 0.01) {
		t.Fatalf("expected net 1000, got %v", res.NetAmount)
	}
	if !almostEqualTax(res.GrossAmount, 1180, 0.01) {
		t.Fatalf("expected gross 1180, got %v", res.GrossAmount)
	}
}

// ------------------------------------------------------------
// income_tax
// ------------------------------------------------------------

func TestTax_IncomeTax_ProgressiveSlabs(t *testing.T) {
	// Classic 3-slab example: 0-3L @0%, 3-6L @5%, 6L+ @10%
	slabs := []TaxSlab{
		{From: 0, To: 300000, Rate: 0},
		{From: 300000, To: 600000, Rate: 5},
		{From: 600000, To: 0, Rate: 10},
	}
	res, err := CalculateTax(TaxRequest{
		Mode:   "income_tax",
		Income: 900000,
		Slabs:  slabs,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Tax: 0 + 5%*300000 + 10%*300000 = 15000 + 30000 = 45000
	if !almostEqualTax(res.TaxAmount, 45000, 0.01) {
		t.Fatalf("expected tax 45000, got %v", res.TaxAmount)
	}
	if !almostEqualTax(res.TaxRate, 5, 0.01) {
		t.Fatalf("expected effective rate 5, got %v", res.TaxRate)
	}
}

func TestTax_IncomeTax_BelowFirstSlab(t *testing.T) {
	slabs := []TaxSlab{{From: 0, To: 300000, Rate: 0}, {From: 300000, To: 0, Rate: 10}}
	res, err := CalculateTax(TaxRequest{Mode: "income_tax", Income: 200000, Slabs: slabs})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TaxAmount != 0 {
		t.Fatalf("expected zero tax, got %v", res.TaxAmount)
	}
}

func TestTax_IncomeTax_UnsortedSlabs(t *testing.T) {
	// Slabs intentionally out of order — backend must sort before applying.
	slabs := []TaxSlab{
		{From: 600000, To: 0, Rate: 10},
		{From: 0, To: 300000, Rate: 0},
		{From: 300000, To: 600000, Rate: 5},
	}
	res, err := CalculateTax(TaxRequest{Mode: "income_tax", Income: 900000, Slabs: slabs})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualTax(res.TaxAmount, 45000, 0.01) {
		t.Fatalf("expected tax 45000, got %v", res.TaxAmount)
	}
}

func TestTax_IncomeTax_EmptySlabs(t *testing.T) {
	_, err := CalculateTax(TaxRequest{Mode: "income_tax", Income: 100000})
	if err == nil {
		t.Fatal("expected error for empty slabs")
	}
}

func TestTax_IncomeTax_OverlappingSlabs(t *testing.T) {
	slabs := []TaxSlab{
		{From: 0, To: 500000, Rate: 5},
		{From: 300000, To: 0, Rate: 10}, // overlaps previous
	}
	_, err := CalculateTax(TaxRequest{Mode: "income_tax", Income: 400000, Slabs: slabs})
	if err == nil {
		t.Fatal("expected error for overlapping slabs")
	}
}

func TestTax_IncomeTax_MultipleOpenEnded(t *testing.T) {
	slabs := []TaxSlab{
		{From: 0, To: 0, Rate: 5},
		{From: 100000, To: 0, Rate: 10},
	}
	_, err := CalculateTax(TaxRequest{Mode: "income_tax", Income: 200000, Slabs: slabs})
	if err == nil {
		t.Fatal("expected error for multiple open-ended slabs")
	}
}

// ------------------------------------------------------------
// Misc
// ------------------------------------------------------------

func TestTax_UnsupportedMode(t *testing.T) {
	_, err := CalculateTax(TaxRequest{Mode: "magic"})
	if err == nil {
		t.Fatal("expected error for unsupported mode")
	}
}

func TestTax_EmptyMode(t *testing.T) {
	_, err := CalculateTax(TaxRequest{})
	if err == nil {
		t.Fatal("expected error for empty mode")
	}
}