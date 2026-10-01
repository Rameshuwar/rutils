package converter

import (
	"math"
	"testing"
)

// ============================================================
// SCIENTIFIC CALCULATOR — unit tests
//
// Grouping mirrors the dispatcher switch in scientific.go:
//   1. Constants
//   2. Trigonometric (sin/cos/tan/csc/sec/cot) — degrees & radians
//   3. Inverse trigonometric (asin/acos/atan/atan2)
//   4. Hyperbolic
//   5. Log / Exp
//   6. Powers / Roots
//   7. Rounding & Sign
//   8. Combinatorics (factorial / ncr / npr)
//   9. Misc (gcd / lcm / mod / hypot)
//  10. Validation errors (angle unit, domain, NaN/Inf, missing value2, bad op)
// ============================================================

// almostEqualScientific gives scientific tests a slightly looser
// tolerance than the 1e-6 used by percentage_test.go, because trig
// results are computed via radians conversion and 10-place rounding.
func almostEqualScientific(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// ptr is a small helper for building *float64 literals in test cases.
func ptr(v float64) *float64 { return &v }

// ============================================================
// 1. CONSTANTS
// ============================================================

func TestScientific_Pi(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{Operation: "pi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, math.Pi) {
		t.Fatalf("expected π=%v, got %v", math.Pi, res.Result)
	}
	if res.Formatted == "" {
		t.Fatal("expected non-empty formatted value")
	}
}

func TestScientific_E(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{Operation: "e"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, math.E) {
		t.Fatalf("expected e=%v, got %v", math.E, res.Result)
	}
}

// ============================================================
// 2. TRIGONOMETRIC
// ============================================================

func TestScientific_Sin_Degrees(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "sin",
		Value1:    30,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 0.5) {
		t.Fatalf("expected sin(30°)=0.5, got %v", res.Result)
	}
	if res.AngleUnit != "degrees" {
		t.Fatalf("expected angleUnit=degrees, got %q", res.AngleUnit)
	}
}

func TestScientific_Sin_Radians_Default(t *testing.T) {
	// No angleUnit specified → defaults to radians
	res, err := CalculateScientific(ScientificRequest{
		Operation: "sin",
		Value1:    math.Pi / 6, // 30°
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 0.5) {
		t.Fatalf("expected sin(π/6)=0.5, got %v", res.Result)
	}
	if res.AngleUnit != "radians" {
		t.Fatalf("expected default angleUnit=radians, got %q", res.AngleUnit)
	}
}

func TestScientific_Cos_Degrees(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "cos",
		Value1:    60,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 0.5) {
		t.Fatalf("expected cos(60°)=0.5, got %v", res.Result)
	}
}

func TestScientific_Tan_Degrees(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "tan",
		Value1:    45,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 1.0) {
		t.Fatalf("expected tan(45°)=1, got %v", res.Result)
	}
}

func TestScientific_Tan_Undefined(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "tan",
		Value1:    90,
		AngleUnit: "degrees",
	})
	if err == nil {
		t.Fatal("expected error for tan(90°)")
	}
}

func TestScientific_Csc(t *testing.T) {
	// csc(30°) = 1/sin(30°) = 1/0.5 = 2
	res, err := CalculateScientific(ScientificRequest{
		Operation: "csc",
		Value1:    30,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 2.0) {
		t.Fatalf("expected csc(30°)=2, got %v", res.Result)
	}
}

func TestScientific_Csc_Undefined(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "csc",
		Value1:    0, // sin(0) = 0 → csc undefined
		AngleUnit: "degrees",
	})
	if err == nil {
		t.Fatal("expected error for csc(0)")
	}
}

func TestScientific_Sec(t *testing.T) {
	// sec(60°) = 1/cos(60°) = 1/0.5 = 2
	res, err := CalculateScientific(ScientificRequest{
		Operation: "sec",
		Value1:    60,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 2.0) {
		t.Fatalf("expected sec(60°)=2, got %v", res.Result)
	}
}

func TestScientific_Cot(t *testing.T) {
	// cot(45°) = cos/sin = 1
	res, err := CalculateScientific(ScientificRequest{
		Operation: "cot",
		Value1:    45,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 1.0) {
		t.Fatalf("expected cot(45°)=1, got %v", res.Result)
	}
}

// ============================================================
// 3. INVERSE TRIGONOMETRIC
// ============================================================

func TestScientific_Asin_Degrees(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "asin",
		Value1:    0.5,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 30.0) {
		t.Fatalf("expected asin(0.5)=30°, got %v", res.Result)
	}
	if res.AngleUnit != "degrees" {
		t.Fatalf("expected angleUnit=degrees, got %q", res.AngleUnit)
	}
}

func TestScientific_Asin_OutOfDomain(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "asin",
		Value1:    2,
	})
	if err == nil {
		t.Fatal("expected error for asin(2)")
	}
}

func TestScientific_Acos_Degrees(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "acos",
		Value1:    0.5,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 60.0) {
		t.Fatalf("expected acos(0.5)=60°, got %v", res.Result)
	}
}

func TestScientific_Atan_Degrees(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "atan",
		Value1:    1,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 45.0) {
		t.Fatalf("expected atan(1)=45°, got %v", res.Result)
	}
}

func TestScientific_Atan2_Quadrant1(t *testing.T) {
	// atan2(y=1, x=1) = 45°
	res, err := CalculateScientific(ScientificRequest{
		Operation: "atan2",
		Value1:    1,
		Value2:    ptr(1),
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 45.0) {
		t.Fatalf("expected atan2(1,1)=45°, got %v", res.Result)
	}
}

func TestScientific_Atan2_Quadrant2(t *testing.T) {
	// atan2(y=1, x=-1) = 135°
	res, err := CalculateScientific(ScientificRequest{
		Operation: "atan2",
		Value1:    1,
		Value2:    ptr(-1),
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 135.0) {
		t.Fatalf("expected atan2(1,-1)=135°, got %v", res.Result)
	}
}

func TestScientific_Atan2_MissingValue2(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "atan2",
		Value1:    1,
	})
	if err == nil {
		t.Fatal("expected error for atan2 without value2")
	}
}

func TestScientific_Atan2_Origin(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "atan2",
		Value1:    0,
		Value2:    ptr(0),
	})
	if err == nil {
		t.Fatal("expected error for atan2(0,0)")
	}
}

// ============================================================
// 4. HYPERBOLIC
// ============================================================

func TestScientific_Sinh(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "sinh",
		Value1:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, math.Sinh(1)) {
		t.Fatalf("expected sinh(1)=%v, got %v", math.Sinh(1), res.Result)
	}
}

func TestScientific_Cosh(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "cosh",
		Value1:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 1.0) {
		t.Fatalf("expected cosh(0)=1, got %v", res.Result)
	}
}

func TestScientific_Tanh(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "tanh",
		Value1:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 0.0) {
		t.Fatalf("expected tanh(0)=0, got %v", res.Result)
	}
}

func TestScientific_Asinh(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "asinh",
		Value1:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 0.0) {
		t.Fatalf("expected asinh(0)=0, got %v", res.Result)
	}
}

func TestScientific_Acosh(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "acosh",
		Value1:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 0.0) {
		t.Fatalf("expected acosh(1)=0, got %v", res.Result)
	}
}

func TestScientific_Acosh_OutOfDomain(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "acosh",
		Value1:    0.5,
	})
	if err == nil {
		t.Fatal("expected error for acosh(0.5)")
	}
}

func TestScientific_Atanh(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "atanh",
		Value1:    0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 0.0) {
		t.Fatalf("expected atanh(0)=0, got %v", res.Result)
	}
}

func TestScientific_Atanh_OutOfDomain(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "atanh",
		Value1:    1, // atanh(1) → +Inf
	})
	if err == nil {
		t.Fatal("expected error for atanh(1)")
	}
}

// ============================================================
// 5. LOG / EXP
// ============================================================

func TestScientific_Log10(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "log",
		Value1:    1000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 3.0) {
		t.Fatalf("expected log10(1000)=3, got %v", res.Result)
	}
}

func TestScientific_Log_Domain(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{Operation: "log", Value1: 0})
	if err == nil {
		t.Fatal("expected error for log(0)")
	}
	_, err = CalculateScientific(ScientificRequest{Operation: "log", Value1: -1})
	if err == nil {
		t.Fatal("expected error for log(-1)")
	}
}

func TestScientific_Ln(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "ln",
		Value1:    math.E,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 1.0) {
		t.Fatalf("expected ln(e)=1, got %v", res.Result)
	}
}

func TestScientific_Ln_Domain(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{Operation: "ln", Value1: 0})
	if err == nil {
		t.Fatal("expected error for ln(0)")
	}
}

func TestScientific_LogBase(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "log_base",
		Value1:    1000,
		Value2:    ptr(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 3.0) {
		t.Fatalf("expected log_base(1000,10)=3, got %v", res.Result)
	}
}

func TestScientific_LogBase_MissingValue2(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "log_base",
		Value1:    100,
	})
	if err == nil {
		t.Fatal("expected error for log_base without base")
	}
}

func TestScientific_LogBase_InvalidBase(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "log_base",
		Value1:    100,
		Value2:    ptr(1), // base 1 → division by log(1)=0
	})
	if err == nil {
		t.Fatal("expected error for log_base with base 1")
	}
}

func TestScientific_Exp(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "exp",
		Value1:    1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, math.E) {
		t.Fatalf("expected exp(1)=e, got %v", res.Result)
	}
}

func TestScientific_Exp_Overflow(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "exp",
		Value1:    800, // e^800 overflows float64
	})
	if err == nil {
		t.Fatal("expected error for exp(800)")
	}
}

// ============================================================
// 6. POWERS / ROOTS
// ============================================================

func TestScientific_Pow(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "pow",
		Value1:    2,
		Value2:    ptr(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 1024) {
		t.Fatalf("expected 2^10=1024, got %v", res.Result)
	}
}

func TestScientific_Pow_MissingValue2(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "pow",
		Value1:    2,
	})
	if err == nil {
		t.Fatal("expected error for pow without exponent")
	}
}

func TestScientific_Pow_NegBaseFractionalExp(t *testing.T) {
	// (-2)^0.5 → NaN in Go
	_, err := CalculateScientific(ScientificRequest{
		Operation: "pow",
		Value1:    -2,
		Value2:    ptr(0.5),
	})
	if err == nil {
		t.Fatal("expected error for (-2)^0.5")
	}
}

func TestScientific_Sqrt(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "sqrt",
		Value1:    144,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 12) {
		t.Fatalf("expected sqrt(144)=12, got %v", res.Result)
	}
}

func TestScientific_Sqrt_Negative(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "sqrt",
		Value1:    -1,
	})
	if err == nil {
		t.Fatal("expected error for sqrt(-1)")
	}
}

func TestScientific_Cbrt(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "cbrt",
		Value1:    27,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 3.0) {
		t.Fatalf("expected cbrt(27)=3, got %v", res.Result)
	}
}

func TestScientific_Cbrt_Negative(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "cbrt",
		Value1:    -27,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, -3.0) {
		t.Fatalf("expected cbrt(-27)=-3, got %v", res.Result)
	}
}

func TestScientific_NthRoot(t *testing.T) {
	// 4th root of 16 = 2
	res, err := CalculateScientific(ScientificRequest{
		Operation: "nth_root",
		Value1:    16,
		Value2:    ptr(4),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 2.0) {
		t.Fatalf("expected 4th-root(16)=2, got %v", res.Result)
	}
}

func TestScientific_NthRoot_EvenRootOfNegative(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "nth_root",
		Value1:    -16,
		Value2:    ptr(4),
	})
	if err == nil {
		t.Fatal("expected error for even root of negative")
	}
}

func TestScientific_NthRoot_ZeroN(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "nth_root",
		Value1:    16,
		Value2:    ptr(0),
	})
	if err == nil {
		t.Fatal("expected error for nth_root with n=0")
	}
}

// ============================================================
// 7. ROUNDING & SIGN
// ============================================================

func TestScientific_Abs(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "abs",
		Value1:    -7.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 7.5) {
		t.Fatalf("expected abs(-7.5)=7.5, got %v", res.Result)
	}
}

func TestScientific_Floor(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "floor",
		Value1:    3.7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 3.0) {
		t.Fatalf("expected floor(3.7)=3, got %v", res.Result)
	}
}

func TestScientific_Ceil(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "ceil",
		Value1:    3.2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 4.0) {
		t.Fatalf("expected ceil(3.2)=4, got %v", res.Result)
	}
}

func TestScientific_Round(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "round",
		Value1:    3.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 4.0) {
		t.Fatalf("expected round(3.5)=4, got %v", res.Result)
	}
}

func TestScientific_Trunc(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "trunc",
		Value1:    -3.9,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, -3.0) {
		t.Fatalf("expected trunc(-3.9)=-3, got %v", res.Result)
	}
}

func TestScientific_Sign(t *testing.T) {
	cases := map[float64]float64{
		-5: -1,
		0:  0,
		5:  1,
	}
	for in, want := range cases {
		res, err := CalculateScientific(ScientificRequest{
			Operation: "sign",
			Value1:    in,
		})
		if err != nil {
			t.Fatalf("unexpected error for sign(%v): %v", in, err)
		}
		if res.Result != want {
			t.Fatalf("expected sign(%v)=%v, got %v", in, want, res.Result)
		}
	}
}

// ============================================================
// 8. COMBINATORICS
// ============================================================

func TestScientific_Factorial(t *testing.T) {
	cases := map[float64]float64{
		0:  1,
		1:  1,
		5:  120,
		10: 3628800,
	}
	for in, want := range cases {
		res, err := CalculateScientific(ScientificRequest{
			Operation: "factorial",
			Value1:    in,
		})
		if err != nil {
			t.Fatalf("unexpected error for %v!: %v", in, err)
		}
		if !almostEqualScientific(res.Result, want) {
			t.Fatalf("expected %v! = %v, got %v", in, want, res.Result)
		}
	}
}

func TestScientific_Factorial_Negative(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "factorial",
		Value1:    -3,
	})
	if err == nil {
		t.Fatal("expected error for (-3)!")
	}
}

func TestScientific_Factorial_NonInteger(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "factorial",
		Value1:    3.5,
	})
	if err == nil {
		t.Fatal("expected error for 3.5!")
	}
}

func TestScientific_Factorial_Overflow(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "factorial",
		Value1:    200, // > 170 → float64 overflow
	})
	if err == nil {
		t.Fatal("expected error for 200!")
	}
}

func TestScientific_NCR(t *testing.T) {
	// C(5, 2) = 10
	res, err := CalculateScientific(ScientificRequest{
		Operation: "ncr",
		Value1:    5,
		Value2:    ptr(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 10) {
		t.Fatalf("expected C(5,2)=10, got %v", res.Result)
	}
}

func TestScientific_NCR_RGreaterThanN(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "ncr",
		Value1:    3,
		Value2:    ptr(5),
	})
	if err == nil {
		t.Fatal("expected error for C(3,5)")
	}
}

func TestScientific_NPR(t *testing.T) {
	// P(5, 2) = 20
	res, err := CalculateScientific(ScientificRequest{
		Operation: "npr",
		Value1:    5,
		Value2:    ptr(2),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 20) {
		t.Fatalf("expected P(5,2)=20, got %v", res.Result)
	}
}

func TestScientific_NPR_NonInteger(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "npr",
		Value1:    5.5,
		Value2:    ptr(2),
	})
	if err == nil {
		t.Fatal("expected error for P(5.5, 2)")
	}
}

// ============================================================
// 9. MISC (gcd / lcm / mod / hypot)
// ============================================================

func TestScientific_GCD(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "gcd",
		Value1:    48,
		Value2:    ptr(18),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 6) {
		t.Fatalf("expected gcd(48,18)=6, got %v", res.Result)
	}
}

func TestScientific_GCD_Zero(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "gcd",
		Value1:    0,
		Value2:    ptr(5),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 5) {
		t.Fatalf("expected gcd(0,5)=5, got %v", res.Result)
	}
}

func TestScientific_LCM(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "lcm",
		Value1:    4,
		Value2:    ptr(6),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 12) {
		t.Fatalf("expected lcm(4,6)=12, got %v", res.Result)
	}
}

func TestScientific_LCM_Zero(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "lcm",
		Value1:    0,
		Value2:    ptr(5),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Result != 0 {
		t.Fatalf("expected lcm(0,5)=0, got %v", res.Result)
	}
}

func TestScientific_Mod(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "mod",
		Value1:    10,
		Value2:    ptr(3),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 1) {
		t.Fatalf("expected 10 mod 3 = 1, got %v", res.Result)
	}
}

func TestScientific_Mod_DivideByZero(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "mod",
		Value1:    10,
		Value2:    ptr(0),
	})
	if err == nil {
		t.Fatal("expected error for mod by zero")
	}
}

func TestScientific_Hypot(t *testing.T) {
	// hypot(3, 4) = 5
	res, err := CalculateScientific(ScientificRequest{
		Operation: "hypot",
		Value1:    3,
		Value2:    ptr(4),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualScientific(res.Result, 5.0) {
		t.Fatalf("expected hypot(3,4)=5, got %v", res.Result)
	}
}

func TestScientific_Hypot_MissingValue2(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "hypot",
		Value1:    3,
	})
	if err == nil {
		t.Fatal("expected error for hypot without value2")
	}
}

// ============================================================
// 10. VALIDATION / DISPATCHER ERRORS
// ============================================================

func TestScientific_EmptyOperation(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{})
	if err == nil {
		t.Fatal("expected error for empty operation")
	}
}

func TestScientific_UnsupportedOperation(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "magic_trick",
		Value1:    1,
	})
	if err == nil {
		t.Fatal("expected error for unsupported operation")
	}
}

func TestScientific_InvalidAngleUnit(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "sin",
		Value1:    30,
		AngleUnit: "gradians",
	})
	if err == nil {
		t.Fatal("expected error for invalid angleUnit")
	}
}

func TestScientific_AngleUnitAliases(t *testing.T) {
	// "deg" / "rad" are accepted as aliases.
	aliases := []string{"degrees", "degree", "deg"}
	for _, a := range aliases {
		res, err := CalculateScientific(ScientificRequest{
			Operation: "sin",
			Value1:    30,
			AngleUnit: a,
		})
		if err != nil {
			t.Fatalf("alias %q rejected: %v", a, err)
		}
		if !almostEqualScientific(res.Result, 0.5) {
			t.Fatalf("alias %q: expected 0.5, got %v", a, res.Result)
		}
		if res.AngleUnit != "degrees" {
			t.Fatalf("alias %q: expected normalized angleUnit=degrees, got %q", a, res.AngleUnit)
		}
	}
}

func TestScientific_NaNInputRejected(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "abs",
		Value1:    math.NaN(),
	})
	if err == nil {
		t.Fatal("expected error for NaN input")
	}
}

func TestScientific_InfInputRejected(t *testing.T) {
	_, err := CalculateScientific(ScientificRequest{
		Operation: "abs",
		Value1:    math.Inf(1),
	})
	if err == nil {
		t.Fatal("expected error for +Inf input")
	}
	_, err = CalculateScientific(ScientificRequest{
		Operation: "sqrt",
		Value1:    math.Inf(-1),
	})
	if err == nil {
		t.Fatal("expected error for -Inf input")
	}
}

func TestScientific_ResponseShape(t *testing.T) {
	res, err := CalculateScientific(ScientificRequest{
		Operation: "sqrt",
		Value1:    4,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Operation != "sqrt" {
		t.Fatalf("expected operation=sqrt, got %q", res.Operation)
	}
	if res.Formatted == "" {
		t.Fatal("expected non-empty formatted field")
	}
	if len(res.Steps) == 0 {
		t.Fatal("expected non-empty steps array")
	}
}

func TestScientific_RoundingPrecision(t *testing.T) {
	// sin(30°) is exactly 0.5 mathematically, but floating-point	// conversion introduces tiny error. The 10-place rounding must
	// clean it up so the caller sees exactly 0.5.
	res, err := CalculateScientific(ScientificRequest{
		Operation: "sin",
		Value1:    30,
		AngleUnit: "degrees",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Result != 0.5 {
		t.Fatalf("expected exact 0.5 after rounding, got %.20f", res.Result)
	}
}
