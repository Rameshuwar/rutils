package converter

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

// ============================================================
// Types
// ============================================================

// TaxSlab is one progressive bracket for income-tax style calculations.
// A `To` of 0 means "and above" (open-ended top slab).
type TaxSlab struct {
	From float64 `json:"from"`
	To   float64 `json:"to"` // 0 == no upper bound
	Rate float64 `json:"rate"`
}

// SlabBreakdown is the per-slab contribution in the response.
type SlabBreakdown struct {
	From      float64 `json:"from"`
	To        float64 `json:"to"`
	Rate      float64 `json:"rate"`
	TaxableIn float64 `json:"taxableIn"`
	Tax       float64 `json:"tax"`
}

// TaxRequest is the unified input for all tax operations.
//
// `Mode` selects the operation; only the fields relevant to that mode
// need to be supplied.
type TaxRequest struct {
	Mode string `json:"mode"`

	// GST / VAT style fields
	Amount    float64 `json:"amount,omitempty"`
	TaxRate   float64 `json:"taxRate,omitempty"`
	TaxType   string  `json:"taxType,omitempty"`   // "cgst_sgst" | "igst"
	NetAmount float64 `json:"netAmount,omitempty"` // used by find_rate
	GrossAmt  float64 `json:"grossAmount,omitempty"`
	TaxPaid   float64 `json:"taxPaid,omitempty"` // used by reverse_gst

	// Income-tax style fields
	Income float64   `json:"income,omitempty"`
	Regime string    `json:"regime,omitempty"`
	Slabs  []TaxSlab `json:"slabs,omitempty"`
}

// TaxResponse is the unified output for all tax operations.
type TaxResponse struct {
	Mode        string                 `json:"mode"`
	NetAmount   float64                `json:"netAmount"`
	TaxAmount   float64                `json:"taxAmount"`
	GrossAmount float64                `json:"grossAmount"`
	TaxRate     float64                `json:"taxRate"`
	Formatted   string                 `json:"formatted"`
	Steps       []string               `json:"steps,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// ============================================================
// Public entry point
// ============================================================

// CalculateTax dispatches to the correct tax operation.
func CalculateTax(req TaxRequest) (*TaxResponse, error) {
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		return nil, errors.New("mode is required")
	}

	switch mode {

	// ---------- 1. Add tax (exclusive → inclusive) ----------
	// GST / VAT: given a net amount and a rate, produce the gross.
	case "add_tax":
		if err := validateAmount(req.Amount, "amount"); err != nil {
			return nil, err
		}
		if err := validateRate(req.TaxRate, "taxRate"); err != nil {
			return nil, err
		}
		taxAmount := roundTo((req.TaxRate/100.0)*req.Amount, 2)
		gross := roundTo(req.Amount+taxAmount, 2)
		return &TaxResponse{
			Mode:        mode,
			NetAmount:   roundTo(req.Amount, 2),
			TaxAmount:   taxAmount,
			GrossAmount: gross,
			TaxRate:     req.TaxRate,
			Formatted:   formatNumber(gross),
			Steps: []string{
				fmt.Sprintf("Tax = %s%% of %s = %s",
					formatNumber(req.TaxRate), formatNumber(req.Amount), formatNumber(taxAmount)),
				fmt.Sprintf("Gross = %s + %s",
					formatNumber(req.Amount), formatNumber(taxAmount)),
				formatNumber(gross),
			},
			Extra: map[string]interface{}{
				"taxLabel": taxLabelForRate(req.TaxRate),
			},
		}, nil

	// ---------- 2. Remove tax (inclusive → exclusive) ----------
	// Given a gross (tax-inclusive) amount, extract the net and tax.
	case "remove_tax":
		if err := validateAmount(req.Amount, "amount"); err != nil {
			return nil, err
		}
		if req.TaxRate <= 0 {
			return nil, errors.New("taxRate must be greater than zero for remove_tax")
		}
		if req.TaxRate >= 100 {
			return nil, errors.New("taxRate must be less than 100 for remove_tax")
		}
		divisor := 1.0 + (req.TaxRate / 100.0)
		net := roundTo(req.Amount/divisor, 2)
		tax := roundTo(req.Amount-net, 2)
		return &TaxResponse{
			Mode:        mode,
			NetAmount:   net,
			TaxAmount:   tax,
			GrossAmount: roundTo(req.Amount, 2),
			TaxRate:     req.TaxRate,
			Formatted:   formatNumber(net),
			Steps: []string{
				fmt.Sprintf("Net = %s / (1 + %s/100)",
					formatNumber(req.Amount), formatNumber(req.TaxRate)),
				fmt.Sprintf("Net = %s / %s", formatNumber(req.Amount), formatNumber(divisor)),
				fmt.Sprintf("Net = %s, Tax = %s", formatNumber(net), formatNumber(tax)),
			},
			Extra: map[string]interface{}{
				"taxLabel": taxLabelForRate(req.TaxRate),
			},
		}, nil

	// ---------- 3. Find rate (net → gross) ----------
	// Given net and gross, compute the effective tax rate.
	case "find_rate":
		if err := validateAmount(req.NetAmount, "netAmount"); err != nil {
			return nil, err
		}
		if err := validateAmount(req.GrossAmt, "grossAmount"); err != nil {
			return nil, err
		}
		if req.NetAmount == 0 {
			return nil, errors.New("netAmount cannot be zero")
		}
		if req.GrossAmt < req.NetAmount {
			return nil, errors.New("grossAmount must be greater than or equal to netAmount")
		}
		tax := roundTo(req.GrossAmt-req.NetAmount, 2)
		rate := roundTo((tax/req.NetAmount)*100.0, 4)
		return &TaxResponse{
			Mode:        mode,
			NetAmount:   roundTo(req.NetAmount, 2),
			TaxAmount:   tax,
			GrossAmount: roundTo(req.GrossAmt, 2),
			TaxRate:     rate,
			Formatted:   formatNumber(rate) + "%",
			Steps: []string{
				fmt.Sprintf("Tax = Gross − Net = %s − %s = %s",
					formatNumber(req.GrossAmt), formatNumber(req.NetAmount), formatNumber(tax)),
				fmt.Sprintf("Rate = (%s / %s) X 100",
					formatNumber(tax), formatNumber(req.NetAmount)),
				formatNumber(rate) + "%",
			},
		}, nil

	// ---------- 4. Split GST (CGST + SGST / IGST) ----------
	// Given a GST-inclusive amount, break it into net + tax, and split
	// the tax into CGST+SGST or IGST.
	case "split_gst":
		if err := validateAmount(req.Amount, "amount"); err != nil {
			return nil, err
		}
		if req.TaxRate <= 0 {
			return nil, errors.New("taxRate must be greater than zero for split_gst")
		}
		if req.TaxRate >= 100 {
			return nil, errors.New("taxRate must be less than 100 for split_gst")
		}
		taxType := strings.ToLower(strings.TrimSpace(req.TaxType))
		if taxType == "" {
			taxType = "cgst_sgst"
		}
		if taxType != "cgst_sgst" && taxType != "igst" {
			return nil, errors.New("taxType must be 'cgst_sgst' or 'igst'")
		}

		divisor := 1.0 + (req.TaxRate / 100.0)
		net := roundTo(req.Amount/divisor, 2)
		totalTax := roundTo(req.Amount-net, 2)

		var cgst, sgst, igst float64
		var steps []string
		if taxType == "igst" {
			igst = totalTax
			steps = []string{
				fmt.Sprintf("Net = %s / (1 + %s/100) = %s",
					formatNumber(req.Amount), formatNumber(req.TaxRate), formatNumber(net)),
				fmt.Sprintf("IGST = %s − %s = %s",
					formatNumber(req.Amount), formatNumber(net), formatNumber(igst)),
			}
		} else {
			cgst = roundTo(totalTax/2.0, 2)
			sgst = roundTo(totalTax-cgst, 2) // absorb rounding drift on SGST
			steps = []string{
				fmt.Sprintf("Net = %s / (1 + %s/100) = %s",
					formatNumber(req.Amount), formatNumber(req.TaxRate), formatNumber(net)),
				fmt.Sprintf("Total GST = %s − %s = %s",
					formatNumber(req.Amount), formatNumber(net), formatNumber(totalTax)),
				fmt.Sprintf("CGST = %s / 2 = %s", formatNumber(totalTax), formatNumber(cgst)),
				fmt.Sprintf("SGST = %s / 2 = %s", formatNumber(totalTax), formatNumber(sgst)),
			}
		}

		return &TaxResponse{
			Mode:        mode,
			NetAmount:   net,
			TaxAmount:   totalTax,
			GrossAmount: roundTo(req.Amount, 2),
			TaxRate:     req.TaxRate,
			Formatted:   formatNumber(net),
			Steps:       steps,
			Extra: map[string]interface{}{
				"taxType": taxType,
				"cgst":    cgst,
				"sgst":    sgst,
				"igst":    igst,
			},
		}, nil

	// ---------- 5. Reverse GST (from tax paid) ----------
	// Given the tax amount and the rate, find the original taxable value.
	case "reverse_gst":
		if err := validateAmount(req.TaxPaid, "taxPaid"); err != nil {
			return nil, err
		}
		if req.TaxRate <= 0 {
			return nil, errors.New("taxRate must be greater than zero for reverse_gst")
		}
		net := roundTo(req.TaxPaid/(req.TaxRate/100.0), 2)
		gross := roundTo(net+req.TaxPaid, 2)
		return &TaxResponse{
			Mode:        mode,
			NetAmount:   net,
			TaxAmount:   roundTo(req.TaxPaid, 2),
			GrossAmount: gross,
			TaxRate:     req.TaxRate,
			Formatted:   formatNumber(net),
			Steps: []string{
				fmt.Sprintf("Net = Tax / Rate = %s / (%s/100)",
					formatNumber(req.TaxPaid), formatNumber(req.TaxRate)),
				fmt.Sprintf("Net = %s / %s",
					formatNumber(req.TaxPaid), formatNumber(req.TaxRate/100.0)),
				fmt.Sprintf("Net = %s, Gross = %s",
					formatNumber(net), formatNumber(gross)),
			},
		}, nil

	// ---------- 6. Income tax (slab-based) ----------
	// Progressive tax using client-supplied slabs.
	case "income_tax":
		if err := validateAmount(req.Income, "income"); err != nil {
			return nil, err
		}
		if len(req.Slabs) == 0 {
			return nil, errors.New("slabs are required for income_tax")
		}
		if err := validateSlabs(req.Slabs); err != nil {
			return nil, err
		}

		breakdown, totalTax := applySlabs(req.Income, req.Slabs)
		effectiveRate := 0.0
		if req.Income > 0 {
			effectiveRate = roundTo((totalTax/req.Income)*100.0, 4)
		}

		steps := []string{
			fmt.Sprintf("Income = %s", formatNumber(req.Income)),
		}
		for _, b := range breakdown {
			if b.TaxableIn <= 0 {
				continue
			}
			steps = append(steps, fmt.Sprintf(
				"%s–%s @ %s%% on %s = %s",
				formatNumber(b.From),
				slabUpperLabel(b.To),
				formatNumber(b.Rate),
				formatNumber(b.TaxableIn),
				formatNumber(b.Tax),
			))
		}
		steps = append(steps,
			fmt.Sprintf("Total Tax = %s", formatNumber(totalTax)),
			fmt.Sprintf("Effective Rate = %s%%", formatNumber(effectiveRate)),
		)

		return &TaxResponse{
			Mode:        mode,
			NetAmount:   roundTo(req.Income, 2),
			TaxAmount:   totalTax,
			GrossAmount: roundTo(req.Income, 2),
			TaxRate:     effectiveRate,
			Formatted:   formatNumber(totalTax),
			Steps:       steps,
			Extra: map[string]interface{}{
				"regime":        req.Regime,
				"effectiveRate": effectiveRate,
				"slabBreakdown": breakdown,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported mode: %s", req.Mode)
	}
}

// ============================================================
// Slab helpers
// ============================================================

// applySlabs walks the slabs, taxes the portion of `income` that falls
// inside each bracket, and returns the per-slab breakdown + total.
func applySlabs(income float64, slabs []TaxSlab) ([]SlabBreakdown, float64) {
	// Sort a copy so we don't mutate the caller's slice.
	sorted := make([]TaxSlab, len(slabs))
	copy(sorted, slabs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].From < sorted[j].From })

	var breakdown []SlabBreakdown
	var total float64

	for _, s := range sorted {
		if income <= s.From {
			break
		}
		upper := s.To
		if upper == 0 || upper > income {
			upper = income
		}
		taxable := upper - s.From
		if taxable < 0 {
			taxable = 0
		}
		tax := roundTo(taxable*(s.Rate/100.0), 2)

		breakdown = append(breakdown, SlabBreakdown{
			From:      s.From,
			To:        s.To,
			Rate:      s.Rate,
			TaxableIn: roundTo(taxable, 2),
			Tax:       tax,
		})
		total += tax
	}

	return breakdown, roundTo(total, 2)
}

// validateSlabs ensures slabs are sorted, non-overlapping, and have
// valid rates. At most one slab may have To == 0 (the top open slab).
func validateSlabs(slabs []TaxSlab) error {
	sorted := make([]TaxSlab, len(slabs))
	copy(sorted, slabs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].From < sorted[j].From })

	// Count open-ended slabs FIRST — this is the more specific error.
	openEndedSeen := 0
	for _, s := range sorted {
		if s.To == 0 {
			openEndedSeen++
		}
	}
	if openEndedSeen > 1 {
		return errors.New("only one slab may be open-ended")
	}

	// Now walk and validate each slab.
	for i, s := range sorted {
		if s.From < 0 {
			return errors.New("slab 'from' cannot be negative")
		}
		if s.To != 0 && s.To <= s.From {
			return errors.New("slab 'to' must be greater than 'from' (or 0 for open-ended)")
		}
		if s.Rate < 0 || s.Rate > 100 {
			return errors.New("slab rate must be between 0 and 100")
		}
		if s.To == 0 && i != len(sorted)-1 {
			return errors.New("open-ended slab must be the last slab")
		}
		if i < len(sorted)-1 {
			next := sorted[i+1]
			if s.To != 0 && s.To > next.From {
				return errors.New("slabs must not overlap")
			}
		}
	}
	return nil
}

// ============================================================
// Small helpers
// ============================================================

func validateAmount(v float64, field string) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%s must be a finite number", field)
	}
	if v < 0 {
		return fmt.Errorf("%s must be greater than or equal to zero", field)
	}
	return nil
}

func validateRate(v float64, field string) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("%s must be a finite number", field)
	}
	if v < 0 || v > 100 {
		return fmt.Errorf("%s must be between 0 and 100", field)
	}
	return nil
}

// taxLabelForRate returns a human-friendly label for common GST/VAT rates.
// Falls back to "Tax" for unknown rates — the backend is country-neutral.
func taxLabelForRate(rate float64) string {
	switch rate {
	case 0, 5, 12, 18, 28:
		return "GST"
	case 20:
		return "VAT"
	default:
		return "Tax"
	}
}

// slabUpperLabel renders the upper bound of a slab for the steps output.
func slabUpperLabel(to float64) string {
	if to == 0 {
		return "∞"
	}
	return formatNumber(to)
}
