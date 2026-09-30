package converter

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ============================================================
// Scientific Calculator — mode-discriminated single-endpoint
//
// Mirrors the design of percentage.go and tax.go:
//   one request struct → one dispatcher switch → one response struct.
//
// All angle-aware trig ops accept an optional `angleUnit` field
// ("degrees" | "radians", default "radians"). The same unit governs
// both the INPUT of sin/cos/tan/... and the OUTPUT of asin/acos/atan.
// ============================================================

// ScientificRequest is the unified input for all scientific operations.
type ScientificRequest struct {
	Operation string   `json:"operation"`
	Value1    float64  `json:"value1"`
	Value2    *float64 `json:"value2,omitempty"` // second argument for 2-arg ops
	AngleUnit string   `json:"angleUnit,omitempty"`
}

// ScientificResponse is the unified output for all scientific operations.
type ScientificResponse struct {
	Operation string                 `json:"operation"`
	Result    float64                `json:"result"`
	Formatted string                 `json:"formatted"`
	AngleUnit string                 `json:"angleUnit,omitempty"`
	Steps     []string               `json:"steps,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// scientificRoundingPlaces is the decimal precision used for all
// scientific results. Higher than the currency-oriented 2 places so
// that trig/log results retain meaningful accuracy.
const scientificRoundingPlaces = 10

// maxFactorialInput caps n! at 170 — 171! overflows float64.
const maxFactorialInput = 170

// ============================================================
// Public entry point
// ============================================================

// CalculateScientific dispatches to the correct scientific operation.
//
// Value semantics per operation:
//   - 1-arg ops (sin, log, sqrt, abs, factorial, ...)  → Value1 only
//   - 2-arg ops (pow, log_base, atan2, ncr, npr, ...)  → Value1 + Value2
//   - 0-arg ops (pi, e)                                → no values needed
func CalculateScientific(req ScientificRequest) (*ScientificResponse, error) {
	op := strings.ToLower(strings.TrimSpace(req.Operation))
	if op == "" {
		return nil, errors.New("operation is required")
	}

	// Resolve angle unit (default radians) — validated once, shared.
	angleUnit, err := normalizeAngleUnit(req.AngleUnit)
	if err != nil {
		return nil, err
	}

	// value2 is optional; unpack once.
	var v2 float64
	hasV2 := req.Value2 != nil
	if hasV2 {
		v2 = *req.Value2
	}

	switch op {

	// --------------------------------------------------------
	// CONSTANTS (0-arg)
	// --------------------------------------------------------
	case "pi":
		return simpleResponse(op, math.Pi, []string{"π = 3.1415926535..."}), nil
	case "e":
		return simpleResponse(op, math.E, []string{"e = 2.7182818284..."}), nil

	// --------------------------------------------------------
	// TRIGONOMETRIC — input in angleUnit, output is a ratio
	// --------------------------------------------------------
	case "sin":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		rad := toRadians(req.Value1, angleUnit)
		res := roundTo(math.Sin(rad), scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps: []string{
				fmt.Sprintf("sin(%s%s)", formatNumber(req.Value1), angleUnitSuffix(angleUnit)),
				formatNumber(res),
			},
			Extra: map[string]interface{}{
				"inputRadians": roundTo(rad, scientificRoundingPlaces),
			},
		}, nil

	case "cos":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		rad := toRadians(req.Value1, angleUnit)
		res := roundTo(math.Cos(rad), scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps: []string{
				fmt.Sprintf("cos(%s%s)", formatNumber(req.Value1), angleUnitSuffix(angleUnit)),
				formatNumber(res),
			},
			Extra: map[string]interface{}{
				"inputRadians": roundTo(rad, scientificRoundingPlaces),
			},
		}, nil

	case "tan":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		rad := toRadians(req.Value1, angleUnit)
		// tan is undefined where cos = 0 (i.e. 90°, 270°, ...)
		if math.Abs(math.Cos(rad)) < 1e-15 {
			return nil, errors.New("tan is undefined at this angle (cosine is zero)")
		}
		res := roundTo(math.Tan(rad), scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps: []string{
				fmt.Sprintf("tan(%s%s)", formatNumber(req.Value1), angleUnitSuffix(angleUnit)),
				formatNumber(res),
			},
			Extra: map[string]interface{}{
				"inputRadians": roundTo(rad, scientificRoundingPlaces),
			},
		}, nil

	case "csc":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		rad := toRadians(req.Value1, angleUnit)
		s := math.Sin(rad)
		if math.Abs(s) < 1e-15 {
			return nil, errors.New("csc is undefined where sine is zero")
		}
		res := roundTo(1.0/s, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps:     []string{fmt.Sprintf("csc(%s%s) = 1 / sin", formatNumber(req.Value1), angleUnitSuffix(angleUnit)), formatNumber(res)},
		}, nil

	case "sec":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		rad := toRadians(req.Value1, angleUnit)
		c := math.Cos(rad)
		if math.Abs(c) < 1e-15 {
			return nil, errors.New("sec is undefined where cosine is zero")
		}
		res := roundTo(1.0/c, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps:     []string{fmt.Sprintf("sec(%s%s) = 1 / cos", formatNumber(req.Value1), angleUnitSuffix(angleUnit)), formatNumber(res)},
		}, nil

	case "cot":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		rad := toRadians(req.Value1, angleUnit)
		s := math.Sin(rad)
		if math.Abs(s) < 1e-15 {
			return nil, errors.New("cot is undefined where sine is zero")
		}
		res := roundTo(math.Cos(rad)/s, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps:     []string{fmt.Sprintf("cot(%s%s) = cos / sin", formatNumber(req.Value1), angleUnitSuffix(angleUnit)), formatNumber(res)},
		}, nil

	// --------------------------------------------------------
	// INVERSE TRIGONOMETRIC — input is a ratio, output in angleUnit
	// --------------------------------------------------------
	case "asin":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if req.Value1 < -1 || req.Value1 > 1 {
			return nil, errors.New("asin is defined only for inputs in [-1, 1]")
		}
		rad := math.Asin(req.Value1)
		out := fromRadians(rad, angleUnit)
		res := roundTo(out, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps: []string{
				fmt.Sprintf("asin(%s)", formatNumber(req.Value1)),
				formatNumber(res) + angleUnitSuffix(angleUnit),
			},
			Extra: map[string]interface{}{
				"resultRadians": roundTo(rad, scientificRoundingPlaces),
			},
		}, nil

	case "acos":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if req.Value1 < -1 || req.Value1 > 1 {
			return nil, errors.New("acos is defined only for inputs in [-1, 1]")
		}
		rad := math.Acos(req.Value1)
		out := fromRadians(rad, angleUnit)
		res := roundTo(out, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps: []string{
				fmt.Sprintf("acos(%s)", formatNumber(req.Value1)),
				formatNumber(res) + angleUnitSuffix(angleUnit),
			},
			Extra: map[string]interface{}{
				"resultRadians": roundTo(rad, scientificRoundingPlaces),
			},
		}, nil

	case "atan":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		rad := math.Atan(req.Value1)
		out := fromRadians(rad, angleUnit)
		res := roundTo(out, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps: []string{
				fmt.Sprintf("atan(%s)", formatNumber(req.Value1)),
				formatNumber(res) + angleUnitSuffix(angleUnit),
			},
			Extra: map[string]interface{}{
				"resultRadians": roundTo(rad, scientificRoundingPlaces),
			},
		}, nil

	case "atan2":
		if !hasV2 {
			return nil, errors.New("value2 is required for atan2 (y, x)")
		}
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if err := guardFinite(v2, "value2"); err != nil {
			return nil, err
		}
		if req.Value1 == 0 && v2 == 0 {
			return nil, errors.New("atan2(0, 0) is undefined")
		}
		rad := math.Atan2(req.Value1, v2)
		out := fromRadians(rad, angleUnit)
		res := roundTo(out, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			AngleUnit: angleUnit,
			Steps: []string{
				fmt.Sprintf("atan2(y=%s, x=%s)", formatNumber(req.Value1), formatNumber(v2)),
				formatNumber(res) + angleUnitSuffix(angleUnit),
			},
			Extra: map[string]interface{}{
				"resultRadians": roundTo(rad, scientificRoundingPlaces),
			},
		}, nil

	// --------------------------------------------------------
	// HYPERBOLIC (always radians — no angle unit)
	// --------------------------------------------------------
	case "sinh":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := roundTo(math.Sinh(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("sinh(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "cosh":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := roundTo(math.Cosh(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("cosh(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "tanh":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := roundTo(math.Tanh(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("tanh(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "asinh":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := roundTo(math.Asinh(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("asinh(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "acosh":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if req.Value1 < 1 {
			return nil, errors.New("acosh is defined only for inputs ≥ 1")
		}
		res := roundTo(math.Acosh(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("acosh(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "atanh":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if req.Value1 <= -1 || req.Value1 >= 1 {
			return nil, errors.New("atanh is defined only for inputs in (-1, 1)")
		}
		res := roundTo(math.Atanh(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("atanh(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	// --------------------------------------------------------
	// LOGARITHMIC & EXPONENTIAL
	// --------------------------------------------------------
	case "log":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if req.Value1 <= 0 {
			return nil, errors.New("log is defined only for positive inputs")
		}
		res := roundTo(math.Log10(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("log10(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "ln":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if req.Value1 <= 0 {
			return nil, errors.New("ln is defined only for positive inputs")
		}
		res := roundTo(math.Log(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("ln(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "log_base":
		if !hasV2 {
			return nil, errors.New("value2 (base) is required for log_base")
		}
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if err := guardFinite(v2, "value2"); err != nil {
			return nil, err
		}
		if req.Value1 <= 0 {
			return nil, errors.New("value1 (argument) must be positive for log_base")
		}
		if v2 <= 0 || v2 == 1 {
			return nil, errors.New("value2 (base) must be positive and not equal to 1")
		}
		res := roundTo(math.Log(req.Value1)/math.Log(v2), scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("log_base(%s, base=%s)", formatNumber(req.Value1), formatNumber(v2)),
				fmt.Sprintf("= ln(%s) / ln(%s)", formatNumber(req.Value1), formatNumber(v2)),
				formatNumber(res),
			},
		}, nil

	case "exp":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		// Guard against overflow (e^709 is near float64 max).
		if req.Value1 > 709 {
			return nil, errors.New("exp overflows float64 for inputs > 709")
		}
		res := roundTo(math.Exp(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("e^%s", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "pow":
		if !hasV2 {
			return nil, errors.New("value2 (exponent) is required for pow")
		}
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if err := guardFinite(v2, "value2"); err != nil {
			return nil, err
		}
		res := math.Pow(req.Value1, v2)
		if math.IsNaN(res) {
			return nil, errors.New("pow is undefined for this input combination (negative base with fractional exponent)")
		}
		if math.IsInf(res, 0) {
			return nil, errors.New("pow overflows float64")
		}
		rounded := roundTo(res, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    rounded,
			Formatted: formatNumber(rounded),
			Steps: []string{
				fmt.Sprintf("%s ^ %s", formatNumber(req.Value1), formatNumber(v2)),
				formatNumber(rounded),
			},
		}, nil

	case "sqrt":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if req.Value1 < 0 {
			return nil, errors.New("sqrt is defined only for non-negative inputs")
		}
		res := roundTo(math.Sqrt(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("sqrt(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "cbrt":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := roundTo(math.Cbrt(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("cbrt(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "nth_root":
		if !hasV2 {
			return nil, errors.New("value2 (n) is required for nth_root")
		}
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if err := guardFinite(v2, "value2"); err != nil {
			return nil, err
		}
		if v2 == 0 {
			return nil, errors.New("value2 (n) cannot be zero for nth_root")
		}
		// Even roots of negatives are not real.
		if req.Value1 < 0 && math.Mod(math.Abs(v2), 2) == 0 {
			return nil, errors.New("even root of a negative number is not real")
		}
		var res float64
		if req.Value1 < 0 {
			// Odd root of a negative number: -((−x)^(1/n))
			res = -math.Pow(-req.Value1, 1.0/v2)
		} else {
			res = math.Pow(req.Value1, 1.0/v2)
		}
		rounded := roundTo(res, scientificRoundingPlaces)
		return &ScientificResponse{
			Operation: op,
			Result:    rounded,
			Formatted: formatNumber(rounded),
			Steps: []string{
				fmt.Sprintf("%s ^ (1/%s)", formatNumber(req.Value1), formatNumber(v2)),
				formatNumber(rounded),
			},
		}, nil

	// --------------------------------------------------------
	// ROUNDING & SIGN
	// --------------------------------------------------------
	case "abs":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := roundTo(math.Abs(req.Value1), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("|%s|", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "floor":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := math.Floor(req.Value1)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("floor(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "ceil":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := math.Ceil(req.Value1)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("ceil(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "round":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := math.Round(req.Value1)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("round(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "trunc":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		res := math.Trunc(req.Value1)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("trunc(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	case "sign":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		var res float64
		switch {
		case req.Value1 > 0:
			res = 1
		case req.Value1 < 0:
			res = -1
		default:
			res = 0
		}
		return simpleResponse(op, res, []string{
			fmt.Sprintf("sign(%s)", formatNumber(req.Value1)),
			formatNumber(res),
		}), nil

	// --------------------------------------------------------
	// FACTORIAL & COMBINATORICS
	// --------------------------------------------------------
	case "factorial":
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		n := int(req.Value1)
		if req.Value1 != float64(n) {
			return nil, errors.New("factorial requires a non-negative integer")
		}
		if n < 0 {
			return nil, errors.New("factorial requires a non-negative integer")
		}
		if n > maxFactorialInput {
			return nil, fmt.Errorf("factorial input exceeds %d (float64 overflow)", maxFactorialInput)
		}
		res := factorialFloat(n)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("%d!", n),
			formatNumber(res),
		}), nil

	case "ncr":
		if !hasV2 {
			return nil, errors.New("value2 (r) is required for ncr")
		}
		n, r, err := validateTwoNonNegInts(req.Value1, v2, "ncr")
		if err != nil {
			return nil, err
		}
		if r > n {
			return nil, errors.New("ncr requires r ≤ n")
		}
		res := combinationsFloat(n, r)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("C(%d, %d)", n, r),
				formatNumber(res),
			},
		}, nil

	case "npr":
		if !hasV2 {
			return nil, errors.New("value2 (r) is required for npr")
		}
		n, r, err := validateTwoNonNegInts(req.Value1, v2, "npr")
		if err != nil {
			return nil, err
		}
		if r > n {
			return nil, errors.New("npr requires r ≤ n")
		}
		res := permutationsFloat(n, r)
		return &ScientificResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("P(%d, %d)", n, r),
				formatNumber(res),
			},
		}, nil

	// --------------------------------------------------------
	// MISC
	// --------------------------------------------------------
	case "gcd":
		if !hasV2 {
			return nil, errors.New("value2 is required for gcd")
		}
		a, b, err := validateTwoNonNegInts(req.Value1, v2, "gcd")
		if err != nil {
			return nil, err
		}
		res := float64(gcdInt(int64(a), int64(b)))
		return simpleResponse(op, res, []string{
			fmt.Sprintf("gcd(%d, %d)", a, b),
			formatNumber(res),
		}), nil

	case "lcm":
		if !hasV2 {
			return nil, errors.New("value2 is required for lcm")
		}
		a, b, err := validateTwoNonNegInts(req.Value1, v2, "lcm")
		if err != nil {
			return nil, err
		}
		if a == 0 || b == 0 {
			return simpleResponse(op, 0, []string{
				fmt.Sprintf("lcm(%d, %d)", a, b),
				"0",
			}), nil
		}
		g := gcdInt(int64(a), int64(b))
		res := float64(int64(a) * int64(b) / g)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("lcm(%d, %d)", a, b),
			formatNumber(res),
		}), nil

	case "mod":
		if !hasV2 {
			return nil, errors.New("value2 (divisor) is required for mod")
		}
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if err := guardFinite(v2, "value2"); err != nil {
			return nil, err
		}
		if v2 == 0 {
			return nil, errors.New("value2 (divisor) cannot be zero for mod")
		}
		res := roundTo(math.Mod(req.Value1, v2), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("%s mod %s", formatNumber(req.Value1), formatNumber(v2)),
			formatNumber(res),
		}), nil

	case "hypot":
		if !hasV2 {
			return nil, errors.New("value2 is required for hypot")
		}
		if err := guardFinite(req.Value1, "value1"); err != nil {
			return nil, err
		}
		if err := guardFinite(v2, "value2"); err != nil {
			return nil, err
		}
		res := roundTo(math.Hypot(req.Value1, v2), scientificRoundingPlaces)
		return simpleResponse(op, res, []string{
			fmt.Sprintf("hypot(%s, %s)", formatNumber(req.Value1), formatNumber(v2)),
			fmt.Sprintf("= sqrt(%s^2 + %s^2)", formatNumber(req.Value1), formatNumber(v2)),
			formatNumber(res),
		}), nil

	default:
		return nil, fmt.Errorf("unsupported operation: %s", req.Operation)
	}
}

// ============================================================
// Helpers
// ============================================================

// normalizeAngleUnit validates and defaults the angleUnit field.
func normalizeAngleUnit(unit string) (string, error) {
	u := strings.ToLower(strings.TrimSpace(unit))
	switch u {
	case "", "radians", "radian", "rad":
		return "radians", nil
	case "degrees", "degree", "deg":
		return "degrees", nil
	default:
		return "", errors.New("angleUnit must be 'degrees' or 'radians'")
	}
}

// angleUnitSuffix returns a short suffix for step strings.
func angleUnitSuffix(unit string) string {
	if unit == "degrees" {
		return "°"
	}
	return " rad"
}

// toRadians converts an angle from the given unit to radians.
func toRadians(v float64, unit string) float64 {
	if unit == "degrees" {
		return v * math.Pi / 180.0
	}
	return v
}

// fromRadians converts an angle from radians to the given unit.
func fromRadians(v float64, unit string) float64 {
	if unit == "degrees" {
		return v * 180.0 / math.Pi
	}
	return v
}

// guardFinite rejects NaN and ±Inf inputs with a clear message.
func guardFinite(v float64, field string) error {
	if math.IsNaN(v) {
		return fmt.Errorf("%s cannot be NaN", field)
	}
	if math.IsInf(v, 0) {
		return fmt.Errorf("%s cannot be infinite", field)
	}
	return nil
}

// simpleResponse builds a response with no angleUnit and no Extra.
func simpleResponse(op string, result float64, steps []string) *ScientificResponse {
	rounded := roundTo(result, scientificRoundingPlaces)
	return &ScientificResponse{
		Operation: op,
		Result:    rounded,
		Formatted: formatNumber(rounded),
		Steps:     steps,
	}
}

// validateTwoNonNegInts checks that both values are non-negative integers
// and returns them as ints. Used by ncr, npr, gcd, lcm.
func validateTwoNonNegInts(a, b float64, opName string) (int, int, error) {
	ai := int(a)
	bi := int(b)
	if a != float64(ai) || ai < 0 {
		return 0, 0, fmt.Errorf("%s requires value1 to be a non-negative integer", opName)
	}
	if b != float64(bi) || bi < 0 {
		return 0, 0, fmt.Errorf("%s requires value2 to be a non-negative integer", opName)
	}
	return ai, bi, nil
}

// factorialFloat returns n! as a float64. Caller must guarantee 0 ≤ n ≤ 170.
func factorialFloat(n int) float64 {
	res := 1.0
	for i := 2; i <= n; i++ {
		res *= float64(i)
	}
	return res
}

// combinationsFloat returns C(n, r) = n! / (r! (n−r)!) using an
// overflow-resistant iterative product.
func combinationsFloat(n, r int) float64 {
	if r > n-r {
		r = n - r
	}
	res := 1.0
	for i := 0; i < r; i++ {
		res = res * float64(n-i) / float64(i+1)
	}
	return roundTo(res, 0)
}

// permutationsFloat returns P(n, r) = n! / (n−r)! using an
// overflow-resistant iterative product.
func permutationsFloat(n, r int) float64 {
	res := 1.0
	for i := 0; i < r; i++ {
		res *= float64(n - i)
	}
	return roundTo(res, 0)
}

// gcdInt is Euclidean GCD for non-negative int64.
func gcdInt(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}