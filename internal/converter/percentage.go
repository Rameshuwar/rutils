package converter

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// PercentageResponse is the unified response for all percentage operations.
type PercentageResponse struct {
	Operation string                 `json:"operation"`
	Result    float64                `json:"result"`
	Formatted string                 `json:"formatted"`
	Steps     []string               `json:"steps,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// CalculatePercentage dispatches to the correct operation.
//
// value1 / value2 / value3 mean different things per operation вЂ”
// see the per-operation comments below.
func CalculatePercentage(
	operation string,
	value1, value2, value3 float64,
	hasValue3 bool,
) (*PercentageResponse, error) {

	op := strings.ToLower(strings.TrimSpace(operation))
	if op == "" {
		return nil, errors.New("operation is required")
	}

	switch op {

	// ---------- 1. X% of Y ----------
	// value1 = percent, value2 = number
	case "percent_of":
		res := (value1 / 100.0) * value2
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("(%s / 100) Г— %s", formatNumber(value1), formatNumber(value2)),
				fmt.Sprintf("%s Г— %s", formatNumber(value1/100.0), formatNumber(value2)),
				formatNumber(res),
			},
		}, nil

	// ---------- 2. X is what % of Y ----------
	// value1 = part, value2 = whole
	case "what_percent":
		if value2 == 0 {
			return nil, errors.New("value2 (whole) cannot be zero")
		}
		res := (value1 / value2) * 100.0
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("(%s / %s) Г— 100", formatNumber(value1), formatNumber(value2)),
				fmt.Sprintf("%s Г— 100", formatNumber(value1/value2)),
				formatNumber(res) + "%",
			},
		}, nil

	// ---------- 3. Percentage change old -> new ----------
	// value1 = old, value2 = new
	case "percent_change":
		if value1 == 0 {
			return nil, errors.New("value1 (old value) cannot be zero")
		}
		diff := value2 - value1
		res := (diff / value1) * 100.0
		direction := "increase"
		if res < 0 {
			direction = "decrease"
		} else if res == 0 {
			direction = "no change"
		}
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("((%s в€’ %s) / %s) Г— 100", formatNumber(value2), formatNumber(value1), formatNumber(value1)),
				fmt.Sprintf("(%s / %s) Г— 100", formatNumber(diff), formatNumber(value1)),
				formatNumber(res) + "%",
			},
			Extra: map[string]interface{}{
				"direction": direction,
				"amount":    math.Abs(diff),
			},
		}, nil

	// ---------- 4. Increase Y by X% ----------
	// value1 = number, value2 = percent
	case "percent_increase":
		increase := (value2 / 100.0) * value1
		res := value1 + increase
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("%s + (%s%% of %s)", formatNumber(value1), formatNumber(value2), formatNumber(value1)),
				fmt.Sprintf("%s + %s", formatNumber(value1), formatNumber(increase)),
				formatNumber(res),
			},
			Extra: map[string]interface{}{
				"increaseAmount": increase,
			},
		}, nil

	// ---------- 5. Decrease Y by X% ----------
	// value1 = number, value2 = percent
	case "percent_decrease":
		decrease := (value2 / 100.0) * value1
		res := value1 - decrease
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("%s в€’ (%s%% of %s)", formatNumber(value1), formatNumber(value2), formatNumber(value1)),
				fmt.Sprintf("%s в€’ %s", formatNumber(value1), formatNumber(decrease)),
				formatNumber(res),
			},
			Extra: map[string]interface{}{
				"decreaseAmount": decrease,
			},
		}, nil

	// ---------- 6. Reverse percentage ----------
	// value1 = final (after percent applied), value2 = percent (signed)
	case "reverse_percent":
		divisor := 1.0 + (value2 / 100.0)
		if divisor == 0 {
			return nil, errors.New("cannot reverse a -100% change")
		}
		res := value1 / divisor
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("%s / (1 + %s/100)", formatNumber(value1), formatNumber(value2)),
				fmt.Sprintf("%s / %s", formatNumber(value1), formatNumber(divisor)),
				formatNumber(res),
			},
		}, nil

	// ---------- 7. X is P% of what number ----------
	// value1 = part, value2 = percent
	case "is_percent_of_what":
		if value2 == 0 {
			return nil, errors.New("value2 (percent) cannot be zero")
		}
		res := value1 / (value2 / 100.0)
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("%s / (%s / 100)", formatNumber(value1), formatNumber(value2)),
				fmt.Sprintf("%s / %s", formatNumber(value1), formatNumber(value2/100.0)),
				formatNumber(res),
			},
		}, nil

	// ---------- 8. Percentage difference between two numbers ----------
	// value1 = a, value2 = b
	case "percent_difference":
		avg := (value1 + value2) / 2.0
		if avg == 0 {
			return nil, errors.New("average of values cannot be zero")
		}
		res := (math.Abs(value1-value2) / avg) * 100.0
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("|%s в€’ %s| / ((%s + %s) / 2) Г— 100",
					formatNumber(value1), formatNumber(value2),
					formatNumber(value1), formatNumber(value2)),
				fmt.Sprintf("%s / %s Г— 100",
					formatNumber(math.Abs(value1-value2)), formatNumber(avg)),
				formatNumber(res) + "%",
			},
		}, nil

	// ---------- 9. Add percentage points ----------
	// value1 = p1, value2 = p2 (both are percents)
	case "add_percent_points":
		res := value1 + value2
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("%s%% + %s%%", formatNumber(value1), formatNumber(value2)),
				formatNumber(res) + "%",
			},
		}, nil

	// ---------- 10. Subtract percentage points ----------
	case "subtract_percent_points":
		res := value1 - value2
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("%s%% в€’ %s%%", formatNumber(value1), formatNumber(value2)),
				formatNumber(res) + "%",
			},
		}, nil

	// ---------- 11. Discount ----------
	// value1 = original price, value2 = discount %
	case "discount":
		savings := (value2 / 100.0) * value1
		finalPrice := value1 - savings
		return &PercentageResponse{
			Operation: op,
			Result:    finalPrice,
			Formatted: formatNumber(finalPrice),
			Steps: []string{
				fmt.Sprintf("Savings = %s%% of %s = %s", formatNumber(value2), formatNumber(value1), formatNumber(savings)),
				fmt.Sprintf("Final Price = %s в€’ %s", formatNumber(value1), formatNumber(savings)),
				formatNumber(finalPrice),
			},
			Extra: map[string]interface{}{
				"savings":    savings,
				"finalPrice": finalPrice,
			},
		}, nil

	// ---------- 12. Markup ----------
	// value1 = cost, value2 = markup %
	case "markup":
		profit := (value2 / 100.0) * value1
		sellingPrice := value1 + profit
		return &PercentageResponse{
			Operation: op,
			Result:    sellingPrice,
			Formatted: formatNumber(sellingPrice),
			Steps: []string{
				fmt.Sprintf("Profit = %s%% of %s = %s", formatNumber(value2), formatNumber(value1), formatNumber(profit)),
				fmt.Sprintf("Selling Price = %s + %s", formatNumber(value1), formatNumber(profit)),
				formatNumber(sellingPrice),
			},
			Extra: map[string]interface{}{
				"profit":       profit,
				"sellingPrice": sellingPrice,
			},
		}, nil

	// ---------- 13. Profit / Loss % ----------
	// value1 = cost price, value2 = selling price
	case "profit_loss":
		if value1 == 0 {
			return nil, errors.New("value1 (cost price) cannot be zero")
		}
		diff := value2 - value1
		res := (diff / value1) * 100.0
		label := "profit"
		if res < 0 {
			label = "loss"
		} else if res == 0 {
			label = "no profit no loss"
		}
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("((%s в€’ %s) / %s) Г— 100", formatNumber(value2), formatNumber(value1), formatNumber(value1)),
				formatNumber(res) + "%",
			},
			Extra: map[string]interface{}{
				"label":  label,
				"amount": math.Abs(diff),
			},
		}, nil

	// ---------- 14. Percent в†’ Fraction ----------
	// value1 = percent
	case "percent_to_fraction":
		decimal := value1 / 100.0
		fraction := percentToFraction(value1)
		return &PercentageResponse{
			Operation: op,
			Result:    decimal,
			Formatted: fraction,
			Steps: []string{
				fmt.Sprintf("%s%% = %s / 100", formatNumber(value1), formatNumber(value1)),
				fmt.Sprintf("= %s (decimal)", formatNumber(decimal)),
				fmt.Sprintf("= %s (fraction)", fraction),
			},
			Extra: map[string]interface{}{
				"fraction": fraction,
				"decimal":  decimal,
			},
		}, nil

	// ---------- 15. Fraction в†’ Percent ----------
	// value1 = numerator, value2 = denominator
	case "fraction_to_percent":
		if value2 == 0 {
			return nil, errors.New("value2 (denominator) cannot be zero")
		}
		res := (value1 / value2) * 100.0
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("(%s / %s) Г— 100", formatNumber(value1), formatNumber(value2)),
				formatNumber(res) + "%",
			},
		}, nil

	// ---------- 16. Decimal в†’ Percent ----------
	// value1 = decimal
	case "decimal_to_percent":
		res := value1 * 100.0
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("%s Г— 100", formatNumber(value1)),
				formatNumber(res) + "%",
			},
		}, nil

	// ---------- 17. Compound percentage ----------
	// value1 = base, value2 = percent per period, value3 = number of periods
	case "compound_percent":
		if !hasValue3 {
			return nil, errors.New("value3 (number of periods) is required for compound_percent")
		}
		if value3 < 0 {
			return nil, errors.New("value3 (number of periods) cannot be negative")
		}
		factor := 1.0 + (value2 / 100.0)
		res := value1 * math.Pow(factor, value3)
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res),
			Steps: []string{
				fmt.Sprintf("%s Г— (1 + %s/100)^%s",
					formatNumber(value1), formatNumber(value2), formatNumber(value3)),
				fmt.Sprintf("%s Г— %s^%s",
					formatNumber(value1), formatNumber(factor), formatNumber(value3)),
				formatNumber(res),
			},
			Extra: map[string]interface{}{
				"periods": value3,
				"growth":  res - value1,
			},
		}, nil

	// ---------- 18. Marks percentage ----------
	// value1 = obtained marks, value2 = total marks
	case "marks_percentage":
		if value2 == 0 {
			return nil, errors.New("value2 (total marks) cannot be zero")
		}
		res := (value1 / value2) * 100.0
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("(%s / %s) Г— 100", formatNumber(value1), formatNumber(value2)),
				formatNumber(res) + "%",
			},
		}, nil

	// ---------- 19. CGPA в†’ Percent (CBSE factor 9.5) ----------
	// value1 = cgpa
	case "cgpa_to_percent":
		const factor = 9.5
		res := value1 * factor
		return &PercentageResponse{
			Operation: op,
			Result:    res,
			Formatted: formatNumber(res) + "%",
			Steps: []string{
				fmt.Sprintf("%s Г— 9.5", formatNumber(value1)),
				formatNumber(res) + "%",
			},
			Extra: map[string]interface{}{
				"factor": factor,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// formatNumber formats a float to a compact string:
// - integers are shown without decimals
// - otherwise up to 6 decimals, trailing zeros trimmed
func formatNumber(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "0"
	}
	if v == math.Trunc(v) && math.Abs(v) < 1e15 {
		return fmt.Sprintf("%.0f", v)
	}
	s := fmt.Sprintf("%.6f", v)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// percentToFraction converts a percentage value to its simplest fraction string.
// e.g. 25 в†’ "1/4", 50 в†’ "1/2", 12.5 в†’ "1/8", 30 в†’ "3/10"
func percentToFraction(percent float64) string {
	// percent = n / d  в†’  we want n/d simplified from (percent / 100)
	if percent == 0 {
		return "0"
	}

	// Handle non-integer percents by scaling up (e.g. 12.5 в†’ 125/10 в†’ 25/2 в†’ 1/8)
	num := percent
	den := 100.0

	// Scale until numerator is (nearly) an integer
	for i := 0; i < 6; i++ {
		if math.Abs(num-math.Round(num)) < 1e-9 {
			break
		}
		num *= 10
		den *= 10
	}

	n := int64(math.Round(num))
	d := int64(math.Round(den))

	if d == 0 {
		return "undefined"
	}

	g := gcd(abs64(n), abs64(d))
	if g == 0 {
		g = 1
	}
	n /= g
	d /= g

	if d == 1 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d/%d", n, d)
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}