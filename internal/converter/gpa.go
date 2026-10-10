package converter

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// ============================================================
// Types
// ============================================================

// GPACourse is one course/paper contributing to a semester GPA.
type GPACourse struct {
	Name       string  `json:"name,omitempty"`
	Credits    float64 `json:"credits"`
	GradePoint float64 `json:"gradePoint"`
	Grade      string  `json:"grade,omitempty"` // optional; if set, GradePoint is derived
}

// GPASemester is one already-computed semester contributing to a CGPA.
type GPASemester struct {
	Name    string  `json:"name,omitempty"`
	Credits float64 `json:"credits"`
	GPA     float64 `json:"gpa"`
}

// GPARequest is the unified input for all GPA operations.
//
// `Mode` selects the operation; only the fields relevant to that mode
// need to be supplied.
type GPARequest struct {
	Mode string `json:"mode"`

	// mode = "semester_gpa"
	Courses []GPACourse `json:"courses,omitempty"`

	// mode = "cumulative_cgpa"
	Semesters []GPASemester `json:"semesters,omitempty"`

	// mode = "cgpa_to_percentage" | "percentage_to_cgpa"
	CGPA       float64 `json:"cgpa,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
	Multiplier float64 `json:"multiplier,omitempty"` // default 9.5 (CBSE)

	// mode = "grade_to_point"
	Grade        string `json:"grade,omitempty"`
	GradingScale string `json:"gradingScale,omitempty"` // "10" | "4" | "5"

	// mode = "target_gpa"
	CurrentCGPA      float64 `json:"currentCgpa,omitempty"`
	CompletedCredits float64 `json:"completedCredits,omitempty"`
	TargetCGPA       float64 `json:"targetCgpa,omitempty"`
	RemainingCredits float64 `json:"remainingCredits,omitempty"`
	Scale            float64 `json:"scale,omitempty"` // max grade point (default 10)
}

// GPACourseResult is one row in the response breakdown.
type GPACourseResult struct {
	Name       string  `json:"name,omitempty"`
	Credits    float64 `json:"credits"`
	GradePoint float64 `json:"gradePoint"`
	Weighted   float64 `json:"weighted"` // credits × gradePoint
}

// GPASemesterResult is one row in the CGPA breakdown.
type GPASemesterResult struct {
	Name     string  `json:"name,omitempty"`
	Credits  float64 `json:"credits"`
	GPA      float64 `json:"gpa"`
	Weighted float64 `json:"weighted"` // credits × gpa
}

// GPAResponse is the unified output for all GPA operations.
// GPAResponse is the unified output for all GPA operations.
// GPAResponse is the unified output for all GPA operations.
type GPAResponse struct {
	Mode          string                 `json:"mode"`
	GPA           float64                `json:"gpa"`
	CGPA          float64                `json:"cgpa"`
	Percentage    float64                `json:"percentage"`
	GradePoint    float64                `json:"gradePoint"`
	RequiredGPA   float64                `json:"requiredGpa"`
	CurrentCGPA   float64                `json:"currentCgpa"`
	TargetCGPA    float64                `json:"targetCgpa"`
	TotalCredits  float64                `json:"totalCredits"`
	TotalWeighted float64                `json:"totalWeighted"`
	Multiplier    float64                `json:"multiplier"`
	Formatted     string                 `json:"formatted"`
	Breakdown     []GPACourseResult      `json:"breakdown,omitempty"`
	Semesters     []GPASemesterResult    `json:"semesters,omitempty"`
	Steps         []string               `json:"steps,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// ============================================================
// Constants
// ============================================================

const (
	// defaultCGPAFactor is the CBSE convention: Percentage = CGPA × 9.5.
	defaultCGPAFactor = 9.5

	// defaultGPAScale is the standard 10-point scale.
	defaultGPAScale = 10.0

	// gpaRoundingPlaces mirrors the 2-decimal convention used by
	// every other calculator in this package.
	gpaRoundingPlaces = 2
)

// ============================================================
// Public entry point
// ============================================================

// CalculateGPA dispatches to the correct GPA/CGPA operation.
func CalculateGPA(req GPARequest) (*GPAResponse, error) {
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		return nil, errors.New("mode is required")
	}

	switch mode {

	// ---------- 1. Semester GPA from a course list ----------
	case "semester_gpa":
		return calculateSemesterGPA(req)

	// ---------- 2. Cumulative CGPA from semester summaries ----------
	case "cumulative_cgpa":
		return calculateCumulativeCGPA(req)

	// ---------- 3. CGPA → Percentage ----------
	case "cgpa_to_percentage":
		return calculateCGPAToPercentage(req)

	// ---------- 4. Percentage → CGPA ----------
	case "percentage_to_cgpa":
		return calculatePercentageToCGPA(req)

	// ---------- 5. Letter grade → grade point ----------
	case "grade_to_point":
		return calculateGradeToPoint(req)

	// ---------- 6. Target GPA planner ----------
	case "target_gpa":
		return calculateTargetGPA(req)

	default:
		return nil, fmt.Errorf("unsupported mode: %s", req.Mode)
	}
}

// ============================================================
// Mode 1 — semester_gpa
// ============================================================

// calculateSemesterGPA computes the credit-weighted GPA for a single
// semester given a list of courses.
//
//	GPA = Σ(credits_i × gradePoint_i) / Σ(credits_i)
func calculateSemesterGPA(req GPARequest) (*GPAResponse, error) {
	if len(req.Courses) == 0 {
		return nil, errors.New("courses are required for semester_gpa")
	}

	var (
		totalCredits  float64
		totalWeighted float64
		breakdown     []GPACourseResult
	)

	for i, c := range req.Courses {
		if err := guardFiniteGPA(c.Credits, fmt.Sprintf("courses[%d].credits", i)); err != nil {
			return nil, err
		}
		if err := guardFiniteGPA(c.GradePoint, fmt.Sprintf("courses[%d].gradePoint", i)); err != nil {
			return nil, err
		}
		if c.Credits <= 0 {
			return nil, fmt.Errorf("courses[%d].credits must be greater than zero", i)
		}
		if c.GradePoint < 0 {
			return nil, fmt.Errorf("courses[%d].gradePoint cannot be negative", i)
		}
		if c.GradePoint > 10 {
			return nil, fmt.Errorf("courses[%d].gradePoint cannot exceed 10", i)
		}

		weighted := roundTo(c.Credits*c.GradePoint, gpaRoundingPlaces)
		totalCredits += c.Credits
		totalWeighted += weighted

		breakdown = append(breakdown, GPACourseResult{
			Name:       c.Name,
			Credits:    c.Credits,
			GradePoint: c.GradePoint,
			Weighted:   weighted,
		})
	}

	if totalCredits <= 0 {
		return nil, errors.New("total credits must be greater than zero")
	}

	gpa := roundTo(totalWeighted/totalCredits, gpaRoundingPlaces)

	steps := []string{
		"GPA = Σ(credits × gradePoint) / Σ(credits)",
	}
	for i, b := range breakdown {
		label := b.Name
		if label == "" {
			label = fmt.Sprintf("Course %d", i+1)
		}
		steps = append(steps, fmt.Sprintf("%s: %s × %s = %s",
			label,
			formatNumber(b.Credits),
			formatNumber(b.GradePoint),
			formatNumber(b.Weighted),
		))
	}
	steps = append(steps,
		fmt.Sprintf("Total Weighted = %s", formatNumber(roundTo(totalWeighted, gpaRoundingPlaces))),
		fmt.Sprintf("Total Credits  = %s", formatNumber(roundTo(totalCredits, gpaRoundingPlaces))),
		fmt.Sprintf("GPA = %s / %s = %s",
			formatNumber(roundTo(totalWeighted, gpaRoundingPlaces)),
			formatNumber(roundTo(totalCredits, gpaRoundingPlaces)),
			formatNumber(gpa),
		),
	)

	return &GPAResponse{
		Mode:          "semester_gpa",
		GPA:           gpa,
		TotalCredits:  roundTo(totalCredits, gpaRoundingPlaces),
		TotalWeighted: roundTo(totalWeighted, gpaRoundingPlaces),
		Formatted:     formatNumber(gpa),
		Breakdown:     breakdown,
		Steps:         steps,
	}, nil
}

// ============================================================
// Mode 2 — cumulative_cgpa
// ============================================================

// calculateCumulativeCGPA computes the credit-weighted CGPA across
// several already-scored semesters.
//
//	CGPA = Σ(credits_i × gpa_i) / Σ(credits_i)
func calculateCumulativeCGPA(req GPARequest) (*GPAResponse, error) {
	if len(req.Semesters) == 0 {
		return nil, errors.New("semesters are required for cumulative_cgpa")
	}

	var (
		totalCredits  float64
		totalWeighted float64
		breakdown     []GPASemesterResult
	)

	for i, s := range req.Semesters {
		if err := guardFiniteGPA(s.Credits, fmt.Sprintf("semesters[%d].credits", i)); err != nil {
			return nil, err
		}
		if err := guardFiniteGPA(s.GPA, fmt.Sprintf("semesters[%d].gpa", i)); err != nil {
			return nil, err
		}
		if s.Credits <= 0 {
			return nil, fmt.Errorf("semesters[%d].credits must be greater than zero", i)
		}
		if s.GPA < 0 {
			return nil, fmt.Errorf("semesters[%d].gpa cannot be negative", i)
		}
		if s.GPA > 10 {
			return nil, fmt.Errorf("semesters[%d].gpa cannot exceed 10", i)
		}

		weighted := roundTo(s.Credits*s.GPA, gpaRoundingPlaces)
		totalCredits += s.Credits
		totalWeighted += weighted

		breakdown = append(breakdown, GPASemesterResult{
			Name:     s.Name,
			Credits:  s.Credits,
			GPA:      s.GPA,
			Weighted: weighted,
		})
	}

	if totalCredits <= 0 {
		return nil, errors.New("total credits must be greater than zero")
	}

	cgpa := roundTo(totalWeighted/totalCredits, gpaRoundingPlaces)

	steps := []string{
		"CGPA = Σ(semester credits × semester GPA) / Σ(semester credits)",
	}
	for i, b := range breakdown {
		label := b.Name
		if label == "" {
			label = fmt.Sprintf("Semester %d", i+1)
		}
		steps = append(steps, fmt.Sprintf("%s: %s × %s = %s",
			label,
			formatNumber(b.Credits),
			formatNumber(b.GPA),
			formatNumber(b.Weighted),
		))
	}
	steps = append(steps,
		fmt.Sprintf("Total Weighted = %s", formatNumber(roundTo(totalWeighted, gpaRoundingPlaces))),
		fmt.Sprintf("Total Credits  = %s", formatNumber(roundTo(totalCredits, gpaRoundingPlaces))),
		fmt.Sprintf("CGPA = %s / %s = %s",
			formatNumber(roundTo(totalWeighted, gpaRoundingPlaces)),
			formatNumber(roundTo(totalCredits, gpaRoundingPlaces)),
			formatNumber(cgpa),
		),
	)

	return &GPAResponse{
		Mode:          "cumulative_cgpa",
		CGPA:          cgpa,
		TotalCredits:  roundTo(totalCredits, gpaRoundingPlaces),
		TotalWeighted: roundTo(totalWeighted, gpaRoundingPlaces),
		Formatted:     formatNumber(cgpa),
		Semesters:     breakdown,
		Steps:         steps,
	}, nil
}

// ============================================================
// Mode 3 — cgpa_to_percentage
// ============================================================

// calculateCGPAToPercentage converts a CGPA to a percentage.
//
//	Percentage = CGPA × multiplier
//
// Multiplier defaults to 9.5 (CBSE). Common alternatives:
//   - VTU       → 10.0
//   - Mumbai    → 7.25
//   - KTU       → 10.0 (approximate)
func calculateCGPAToPercentage(req GPARequest) (*GPAResponse, error) {
	if err := guardFiniteGPA(req.CGPA, "cgpa"); err != nil {
		return nil, err
	}
	if req.CGPA < 0 {
		return nil, errors.New("cgpa cannot be negative")
	}
	if req.CGPA > 10 {
		return nil, errors.New("cgpa cannot exceed 10")
	}

	multiplier := req.Multiplier
	if multiplier == 0 {
		multiplier = defaultCGPAFactor
	}
	if err := guardFiniteGPA(multiplier, "multiplier"); err != nil {
		return nil, err
	}
	if multiplier <= 0 {
		return nil, errors.New("multiplier must be greater than zero")
	}

	percentage := roundTo(req.CGPA*multiplier, gpaRoundingPlaces)
	if percentage > 100 {
		percentage = 100
	}

	steps := []string{
		"Percentage = CGPA × multiplier",
		fmt.Sprintf("Percentage = %s × %s",
			formatNumber(req.CGPA),
			formatNumber(multiplier),
		),
		formatNumber(percentage) + "%",
	}

	return &GPAResponse{
		Mode:       "cgpa_to_percentage",
		CGPA:       req.CGPA,
		Percentage: percentage,
		Multiplier: multiplier,
		Formatted:  formatNumber(percentage) + "%",
		Steps:      steps,
		Extra: map[string]interface{}{
			"multiplier": multiplier,
			"scaleNote":  scaleNoteForMultiplier(multiplier),
		},
	}, nil
}

// ============================================================
// Mode 4 — percentage_to_cgpa
// ============================================================

// calculatePercentageToCGPA converts a percentage to a CGPA.
//
//	CGPA = Percentage / multiplier
func calculatePercentageToCGPA(req GPARequest) (*GPAResponse, error) {
	if err := guardFiniteGPA(req.Percentage, "percentage"); err != nil {
		return nil, err
	}
	if req.Percentage < 0 {
		return nil, errors.New("percentage cannot be negative")
	}
	if req.Percentage > 100 {
		return nil, errors.New("percentage cannot exceed 100")
	}

	multiplier := req.Multiplier
	if multiplier == 0 {
		multiplier = defaultCGPAFactor
	}
	if err := guardFiniteGPA(multiplier, "multiplier"); err != nil {
		return nil, err
	}
	if multiplier <= 0 {
		return nil, errors.New("multiplier must be greater than zero")
	}

	cgpa := roundTo(req.Percentage/multiplier, gpaRoundingPlaces)

	steps := []string{
		"CGPA = Percentage / multiplier",
		fmt.Sprintf("CGPA = %s / %s",
			formatNumber(req.Percentage),
			formatNumber(multiplier),
		),
		formatNumber(cgpa),
	}

	return &GPAResponse{
		Mode:       "percentage_to_cgpa",
		CGPA:       cgpa,
		Percentage: req.Percentage,
		Multiplier: multiplier,
		Formatted:  formatNumber(cgpa),
		Steps:      steps,
		Extra: map[string]interface{}{
			"multiplier": multiplier,
			"scaleNote":  scaleNoteForMultiplier(multiplier),
		},
	}, nil
}

// ============================================================
// Mode 5 — grade_to_point
// ============================================================

// gradeScaleTables maps a scale name to its grade→point table.
//
// The 10-point table follows the CBSE / AICTE convention; the 4-point
// table follows the US university convention; the 5-point table is a
// generic European secondary-school convention.
var gradeScaleTables = map[string]map[string]float64{
	"10": {
		"O":  10, "A+": 10,
		"A":  9, "A-": 8,
		"B+": 8, "B":  7, "B-": 6,
		"C+": 6, "C":  5, "C-": 5,
		"D":  4, "E":  3,
		"F":  0, "FAIL": 0,
	},
	"4": {
		"A+": 4.0, "A": 4.0, "A-": 3.7,
		"B+": 3.3, "B": 3.0, "B-": 2.7,
		"C+": 2.3, "C": 2.0, "C-": 1.7,
		"D+": 1.3, "D": 1.0, "D-": 0.7,
		"F":  0.0, "FAIL": 0.0,
	},
	"5": {
		"A":  5, "B":  4, "C":  3,
		"D":  2, "E":  1, "F":  0,
		"FAIL": 0,
	},
}

// calculateGradeToPoint looks up the numeric grade point for a letter
// grade on a given scale.
func calculateGradeToPoint(req GPARequest) (*GPAResponse, error) {
	grade := strings.ToUpper(strings.TrimSpace(req.Grade))
	if grade == "" {
		return nil, errors.New("grade is required for grade_to_point")
	}

	scale := strings.TrimSpace(req.GradingScale)
	if scale == "" {
		scale = "10"
	}
	table, ok := gradeScaleTables[scale]
	if !ok {
		return nil, errors.New("gradingScale must be one of: 10, 4, 5")
	}

	point, ok := table[grade]
	if !ok {
		return nil, fmt.Errorf("grade %q is not recognised on the %s-point scale", grade, scale)
	}

	maxPoint := 10.0
	switch scale {
	case "4":
		maxPoint = 4.0
	case "5":
		maxPoint = 5.0
	}

	steps := []string{
		fmt.Sprintf("Grading scale: %s-point", scale),
		fmt.Sprintf("Grade %s → %s grade points", grade, formatNumber(point)),
	}

	return &GPAResponse{
		Mode:       "grade_to_point",
		GradePoint: point,
		Formatted:  formatNumber(point),
		Steps:      steps,
		Extra: map[string]interface{}{
			"grade": grade,
			"scale": scale,
			"max":   maxPoint,
		},
	}, nil
}

// ============================================================
// Mode 6 — target_gpa
// ============================================================

// calculateTargetGPA solves for the GPA a student must achieve in their
// remaining credits to reach a desired overall CGPA.
//
//	requiredGPA = (targetCGPA × totalCredits − currentCGPA × completedCredits) / remainingCredits
//
// Where totalCredits = completedCredits + remainingCredits.
func calculateTargetGPA(req GPARequest) (*GPAResponse, error) {
	if err := guardFiniteGPA(req.CurrentCGPA, "currentCgpa"); err != nil {
		return nil, err
	}
	if err := guardFiniteGPA(req.CompletedCredits, "completedCredits"); err != nil {
		return nil, err
	}
	if err := guardFiniteGPA(req.TargetCGPA, "targetCgpa"); err != nil {
		return nil, err
	}
	if err := guardFiniteGPA(req.RemainingCredits, "remainingCredits"); err != nil {
		return nil, err
	}

	scale := req.Scale
	if scale == 0 {
		scale = defaultGPAScale
	}

	if req.CurrentCGPA < 0 || req.CurrentCGPA > scale {
		return nil, fmt.Errorf("currentCgpa must be between 0 and %s", formatNumber(scale))
	}
	if req.TargetCGPA < 0 || req.TargetCGPA > scale {
		return nil, fmt.Errorf("targetCgpa must be between 0 and %s", formatNumber(scale))
	}
	if req.CompletedCredits < 0 {
		return nil, errors.New("completedCredits cannot be negative")
	}
	if req.RemainingCredits <= 0 {
		return nil, errors.New("remainingCredits must be greater than zero")
	}

	totalCredits := req.CompletedCredits + req.RemainingCredits
	numerator := req.TargetCGPA*totalCredits - req.CurrentCGPA*req.CompletedCredits
	requiredGPA := roundTo(numerator/req.RemainingCredits, gpaRoundingPlaces)

	if requiredGPA < 0 {
		requiredGPA = 0
	}

	achievable := requiredGPA <= scale

	steps := []string{
		"requiredGPA = (targetCGPA × totalCredits − currentCGPA × completedCredits) / remainingCredits",
		fmt.Sprintf("totalCredits = %s + %s = %s",
			formatNumber(req.CompletedCredits),
			formatNumber(req.RemainingCredits),
			formatNumber(totalCredits),
		),
		fmt.Sprintf("numerator = %s × %s − %s × %s = %s",
			formatNumber(req.TargetCGPA),
			formatNumber(totalCredits),
			formatNumber(req.CurrentCGPA),
			formatNumber(req.CompletedCredits),
			formatNumber(roundTo(numerator, gpaRoundingPlaces)),
		),
		fmt.Sprintf("requiredGPA = %s / %s = %s",
			formatNumber(roundTo(numerator, gpaRoundingPlaces)),
			formatNumber(req.RemainingCredits),
			formatNumber(requiredGPA),
		),
	}

	extra := map[string]interface{}{
		"achievable":   achievable,
		"scale":        scale,
		"currentCGPA":  req.CurrentCGPA,
		"targetCGPA":   req.TargetCGPA,
		"totalCredits": roundTo(totalCredits, gpaRoundingPlaces),
	}

	if !achievable {
		extra["warning"] = fmt.Sprintf(
			"Reaching CGPA %s requires a %.2f GPA in your remaining credits, "+
				"which exceeds the maximum of %s on this scale.",
			formatNumber(req.TargetCGPA),
			requiredGPA,
			formatNumber(scale),
		)
	}

	return &GPAResponse{
		Mode:          "target_gpa",
		RequiredGPA:   requiredGPA,
		TotalCredits:  roundTo(totalCredits, gpaRoundingPlaces),
		CurrentCGPA:   req.CurrentCGPA,
		TargetCGPA:    req.TargetCGPA,
		Formatted:     formatNumber(requiredGPA),
		Steps:         steps,
		Extra:         extra,
	}, nil
}

// ============================================================
// Helpers
// ============================================================

// guardFiniteGPA rejects NaN / ±Inf with a descriptive field name.
func guardFiniteGPA(v float64, field string) error {
	if math.IsNaN(v) {
		return fmt.Errorf("%s cannot be NaN", field)
	}
	if math.IsInf(v, 0) {
		return fmt.Errorf("%s cannot be infinite", field)
	}
	return nil
}

// scaleNoteForMultiplier returns a short human-readable hint about the
// origin of a given CGPA→percentage multiplier.
func scaleNoteForMultiplier(m float64) string {
	switch {
	case m == 9.5:
		return "CBSE / AICTE standard"
	case m == 10.0:
		return "VTU / KTU / 10-point scale"
	case m == 7.25:
		return "Mumbai University standard"
	case m >= 9.0 && m <= 10.0:
		return "typical 10-point scale institution"
	case m >= 7.0 && m < 9.0:
		return "regional university convention"
	default:
		return "custom multiplier"
	}
}