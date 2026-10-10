package converter

import (
	"math"
	"testing"
)

// almostEqualGPA is the tolerance used throughout these tests. The
// calculator rounds to 2 decimals, so 0.01 covers the worst case.
func almostEqualGPA(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ============================================================
// Dispatcher / mode validation
// ============================================================

func TestGPA_EmptyMode(t *testing.T) {
	_, err := CalculateGPA(GPARequest{})
	if err == nil {
		t.Fatal("expected error for empty mode")
	}
}

func TestGPA_UnsupportedMode(t *testing.T) {
	_, err := CalculateGPA(GPARequest{Mode: "magic"})
	if err == nil {
		t.Fatal("expected error for unsupported mode")
	}
}

func TestGPA_ModeIsCaseInsensitive(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode: "  SEMESTER_GPA  ",
		Courses: []GPACourse{
			{Credits: 3, GradePoint: 9},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Mode != "semester_gpa" {
		t.Fatalf("expected mode=semester_gpa, got %q", res.Mode)
	}
}

// ============================================================
// Mode 1 — semester_gpa
// ============================================================

func TestSemesterGPA_StandardThreeCourses(t *testing.T) {
	// (4×9 + 3×8 + 3×10) / (4+3+3) = (36 + 24 + 30) / 10 = 90 / 10 = 9.0
	res, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Name: "Math", Credits: 4, GradePoint: 9},
			{Name: "Physics", Credits: 3, GradePoint: 8},
			{Name: "Chemistry", Credits: 3, GradePoint: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Mode != "semester_gpa" {
		t.Fatalf("expected mode=semester_gpa, got %q", res.Mode)
	}
	if !almostEqualGPA(res.GPA, 9.0, 0.01) {
		t.Fatalf("expected GPA 9.0, got %v", res.GPA)
	}
	if !almostEqualGPA(res.TotalCredits, 10, 0.01) {
		t.Fatalf("expected totalCredits 10, got %v", res.TotalCredits)
	}
	if !almostEqualGPA(res.TotalWeighted, 90, 0.01) {
		t.Fatalf("expected totalWeighted 90, got %v", res.TotalWeighted)
	}
	if len(res.Breakdown) != 3 {
		t.Fatalf("expected 3 breakdown rows, got %d", len(res.Breakdown))
	}
	if res.Breakdown[0].Weighted != 36 {
		t.Fatalf("expected first row weighted=36, got %v", res.Breakdown[0].Weighted)
	}
	if res.Formatted != "9" {
		t.Fatalf("expected formatted=9, got %q", res.Formatted)
	}
	if len(res.Steps) == 0 {
		t.Fatal("expected non-empty steps")
	}
}

func TestSemesterGPA_SingleCourse(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 5, GradePoint: 8},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GPA != 8 {
		t.Fatalf("expected GPA 8, got %v", res.GPA)
	}
}

func TestSemesterGPA_FractionalCredits(t *testing.T) {
	// (1.5×9 + 2.5×7) / 4 = (13.5 + 17.5) / 4 = 31/4 = 7.75
	res, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 1.5, GradePoint: 9},
			{Credits: 2.5, GradePoint: 7},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.GPA, 7.75, 0.01) {
		t.Fatalf("expected GPA 7.75, got %v", res.GPA)
	}
}

func TestSemesterGPA_FractionalGradePoints(t *testing.T) {
	// (3×8.5 + 3×9.5) / 6 = (25.5 + 28.5) / 6 = 54/6 = 9.0
	res, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 3, GradePoint: 8.5},
			{Credits: 3, GradePoint: 9.5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.GPA, 9.0, 0.01) {
		t.Fatalf("expected GPA 9.0, got %v", res.GPA)
	}
}

func TestSemesterGPA_PerfectScore(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 4, GradePoint: 10},
			{Credits: 4, GradePoint: 10},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GPA != 10 {
		t.Fatalf("expected GPA 10, got %v", res.GPA)
	}
}

func TestSemesterGPA_ZeroGradePoints(t *testing.T) {
	// All F grades → GPA 0.
	res, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 3, GradePoint: 0},
			{Credits: 4, GradePoint: 0},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GPA != 0 {
		t.Fatalf("expected GPA 0, got %v", res.GPA)
	}
}

func TestSemesterGPA_EmptyCourseList(t *testing.T) {
	_, err := CalculateGPA(GPARequest{Mode: "semester_gpa"})
	if err == nil {
		t.Fatal("expected error for empty course list")
	}
}

func TestSemesterGPA_ZeroCredits(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 0, GradePoint: 9},
		},
	})
	if err == nil {
		t.Fatal("expected error for zero credits")
	}
}

func TestSemesterGPA_NegativeCredits(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: -3, GradePoint: 9},
		},
	})
	if err == nil {
		t.Fatal("expected error for negative credits")
	}
}

func TestSemesterGPA_NegativeGradePoint(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 3, GradePoint: -1},
		},
	})
	if err == nil {
		t.Fatal("expected error for negative gradePoint")
	}
}

func TestSemesterGPA_GradePointAbove10(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 3, GradePoint: 11},
		},
	})
	if err == nil {
		t.Fatal("expected error for gradePoint > 10")
	}
}

func TestSemesterGPA_NaNInputRejected(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: math.NaN(), GradePoint: 9},
		},
	})
	if err == nil {
		t.Fatal("expected error for NaN credits")
	}
}

func TestSemesterGPA_InfInputRejected(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "semester_gpa",
		Courses: []GPACourse{
			{Credits: 3, GradePoint: math.Inf(1)},
		},
	})
	if err == nil {
		t.Fatal("expected error for infinite gradePoint")
	}
}

// ============================================================
// Mode 2 — cumulative_cgpa
// ============================================================

func TestCumulativeCGPA_ThreeSemesters(t *testing.T) {
	// (20×8 + 20×9 + 20×9.5) / 60 = (160 + 180 + 190) / 60 = 530 / 60 = 8.8333...
	res, err := CalculateGPA(GPARequest{
		Mode: "cumulative_cgpa",
		Semesters: []GPASemester{
			{Name: "Sem 1", Credits: 20, GPA: 8},
			{Name: "Sem 2", Credits: 20, GPA: 9},
			{Name: "Sem 3", Credits: 20, GPA: 9.5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Mode != "cumulative_cgpa" {
		t.Fatalf("expected mode=cumulative_cgpa, got %q", res.Mode)
	}
	if !almostEqualGPA(res.CGPA, 8.83, 0.01) {
		t.Fatalf("expected CGPA 8.83, got %v", res.CGPA)
	}
	if !almostEqualGPA(res.TotalCredits, 60, 0.01) {
		t.Fatalf("expected totalCredits 60, got %v", res.TotalCredits)
	}
	if len(res.Semesters) != 3 {
		t.Fatalf("expected 3 semester rows, got %d", len(res.Semesters))
	}
}

func TestCumulativeCGPA_UnequalCredits(t *testing.T) {
	// (24×9 + 18×7) / 42 = (216 + 126) / 42 = 342/42 = 8.1428...
	res, err := CalculateGPA(GPARequest{
		Mode: "cumulative_cgpa",
		Semesters: []GPASemester{
			{Credits: 24, GPA: 9},
			{Credits: 18, GPA: 7},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.CGPA, 8.14, 0.01) {
		t.Fatalf("expected CGPA 8.14, got %v", res.CGPA)
	}
}

func TestCumulativeCGPA_SingleSemester(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode: "cumulative_cgpa",
		Semesters: []GPASemester{
			{Credits: 20, GPA: 8.5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CGPA != 8.5 {
		t.Fatalf("expected CGPA 8.5, got %v", res.CGPA)
	}
}

func TestCumulativeCGPA_EmptySemesterList(t *testing.T) {
	_, err := CalculateGPA(GPARequest{Mode: "cumulative_cgpa"})
	if err == nil {
		t.Fatal("expected error for empty semester list")
	}
}

func TestCumulativeCGPA_ZeroCredits(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "cumulative_cgpa",
		Semesters: []GPASemester{
			{Credits: 0, GPA: 8},
		},
	})
	if err == nil {
		t.Fatal("expected error for zero credits")
	}
}

func TestCumulativeCGPA_NegativeGPA(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "cumulative_cgpa",
		Semesters: []GPASemester{
			{Credits: 20, GPA: -1},
		},
	})
	if err == nil {
		t.Fatal("expected error for negative GPA")
	}
}

func TestCumulativeCGPA_GPAAbove10(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "cumulative_cgpa",
		Semesters: []GPASemester{
			{Credits: 20, GPA: 11},
		},
	})
	if err == nil {
		t.Fatal("expected error for GPA > 10")
	}
}

// ============================================================
// Mode 3 — cgpa_to_percentage
// ============================================================

func TestCGPAToPercentage_DefaultMultiplier(t *testing.T) {
	// 8.0 × 9.5 = 76
	res, err := CalculateGPA(GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: 8.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Mode != "cgpa_to_percentage" {
		t.Fatalf("expected mode=cgpa_to_percentage, got %q", res.Mode)
	}
	if !almostEqualGPA(res.Percentage, 76, 0.01) {
		t.Fatalf("expected percentage 76, got %v", res.Percentage)
	}
	if res.Multiplier != 9.5 {
		t.Fatalf("expected multiplier 9.5, got %v", res.Multiplier)
	}
	if res.Formatted != "76%" {
		t.Fatalf("expected formatted=76%%, got %q", res.Formatted)
	}
}

func TestCGPAToPercentage_VTUMultiplier(t *testing.T) {
	// 8.4 × 10 = 84
	res, err := CalculateGPA(GPARequest{
		Mode:       "cgpa_to_percentage",
		CGPA:       8.4,
		Multiplier: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.Percentage, 84, 0.01) {
		t.Fatalf("expected percentage 84, got %v", res.Percentage)
	}
	if res.Multiplier != 10 {
		t.Fatalf("expected multiplier 10, got %v", res.Multiplier)
	}
}

func TestCGPAToPercentage_MumbaiMultiplier(t *testing.T) {
	// 8.0 × 7.25 = 58
	res, err := CalculateGPA(GPARequest{
		Mode:       "cgpa_to_percentage",
		CGPA:       8.0,
		Multiplier: 7.25,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.Percentage, 58, 0.01) {
		t.Fatalf("expected percentage 58, got %v", res.Percentage)
	}
}

func TestCGPAToPercentage_CappedAt100(t *testing.T) {
	// 10 × 10 = 100, exactly at the cap.
	res, err := CalculateGPA(GPARequest{
		Mode:       "cgpa_to_percentage",
		CGPA:       10,
		Multiplier: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Percentage != 100 {
		t.Fatalf("expected percentage 100, got %v", res.Percentage)
	}
}

func TestCGPAToPercentage_ZeroCGPA(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Percentage != 0 {
		t.Fatalf("expected percentage 0, got %v", res.Percentage)
	}
}

func TestCGPAToPercentage_NegativeCGPA(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: -1,
	})
	if err == nil {
		t.Fatal("expected error for negative CGPA")
	}
}

func TestCGPAToPercentage_CGPAAbove10(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: 11,
	})
	if err == nil {
		t.Fatal("expected error for CGPA > 10")
	}
}

func TestCGPAToPercentage_NegativeMultiplier(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:       "cgpa_to_percentage",
		CGPA:       8,
		Multiplier: -1,
	})
	if err == nil {
		t.Fatal("expected error for negative multiplier")
	}
}

// ============================================================
// Mode 4 — percentage_to_cgpa
// ============================================================

func TestPercentageToCGPA_DefaultMultiplier(t *testing.T) {
	// 76 / 9.5 = 8.0
	res, err := CalculateGPA(GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: 76,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.CGPA, 8.0, 0.01) {
		t.Fatalf("expected CGPA 8.0, got %v", res.CGPA)
	}
	if res.Multiplier != 9.5 {
		t.Fatalf("expected multiplier 9.5, got %v", res.Multiplier)
	}
}

func TestPercentageToCGPA_CustomMultiplier(t *testing.T) {
	// 84 / 10 = 8.4
	res, err := CalculateGPA(GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: 84,
		Multiplier: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.CGPA, 8.4, 0.01) {
		t.Fatalf("expected CGPA 8.4, got %v", res.CGPA)
	}
}

func TestPercentageToCGPA_Zero(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CGPA != 0 {
		t.Fatalf("expected CGPA 0, got %v", res.CGPA)
	}
}

func TestPercentageToCGPA_NegativePercentage(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: -5,
	})
	if err == nil {
		t.Fatal("expected error for negative percentage")
	}
}

func TestPercentageToCGPA_Above100(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: 101,
	})
	if err == nil {
		t.Fatal("expected error for percentage > 100")
	}
}

// Round-trip: CGPA → Percentage → CGPA should be within tolerance.
func TestCGPAPercentageRoundTrip(t *testing.T) {
	original := 8.4

	toPct, err := CalculateGPA(GPARequest{
		Mode:       "cgpa_to_percentage",
		CGPA:       original,
		Multiplier: 9.5,
	})
	if err != nil {
		t.Fatalf("unexpected error in cgpa→pct: %v", err)
	}

	back, err := CalculateGPA(GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: toPct.Percentage,
		Multiplier: 9.5,
	})
	if err != nil {
		t.Fatalf("unexpected error in pct→cgpa: %v", err)
	}

	if !almostEqualGPA(back.CGPA, original, 0.01) {
		t.Fatalf("round-trip failed: started at %v, ended at %v", original, back.CGPA)
	}
}

// ============================================================
// Mode 5 — grade_to_point
// ============================================================

func TestGradeToPoint_TenPointScale(t *testing.T) {
	cases := map[string]float64{
		"O":  10,
		"A+": 10,
		"A":  9,
		"B+": 8,
		"B":  7,
		"C":  5,
		"F":  0,
	}
	for grade, want := range cases {
		res, err := CalculateGPA(GPARequest{
			Mode:         "grade_to_point",
			Grade:        grade,
			GradingScale: "10",
		})
		if err != nil {
			t.Fatalf("grade %q: unexpected error: %v", grade, err)
		}
		if res.GradePoint != want {
			t.Fatalf("grade %q: expected %v, got %v", grade, want, res.GradePoint)
		}
	}
}

func TestGradeToPoint_FourPointScale(t *testing.T) {
	cases := map[string]float64{
		"A":  4.0,
		"A-": 3.7,
		"B+": 3.3,
		"B":  3.0,
		"C":  2.0,
		"F":  0.0,
	}
	for grade, want := range cases {
		res, err := CalculateGPA(GPARequest{
			Mode:         "grade_to_point",
			Grade:        grade,
			GradingScale: "4",
		})
		if err != nil {
			t.Fatalf("grade %q: unexpected error: %v", grade, err)
		}
		if !almostEqualGPA(res.GradePoint, want, 0.01) {
			t.Fatalf("grade %q: expected %v, got %v", grade, want, res.GradePoint)
		}
	}
}

func TestGradeToPoint_FivePointScale(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode:         "grade_to_point",
		Grade:        "B",
		GradingScale: "5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GradePoint != 4 {
		t.Fatalf("expected 4, got %v", res.GradePoint)
	}
}

func TestGradeToPoint_DefaultScaleIs10(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode:  "grade_to_point",
		Grade: "A",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GradePoint != 9 {
		t.Fatalf("expected 9, got %v", res.GradePoint)
	}
	if res.Extra["scale"] != "10" {
		t.Fatalf("expected scale=10 in extra, got %v", res.Extra["scale"])
	}
}

func TestGradeToPoint_CaseInsensitive(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode:  "grade_to_point",
		Grade: "a+",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GradePoint != 10 {
		t.Fatalf("expected 10, got %v", res.GradePoint)
	}
}

func TestGradeToPoint_EmptyGrade(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode: "grade_to_point",
	})
	if err == nil {
		t.Fatal("expected error for empty grade")
	}
}

func TestGradeToPoint_UnknownGrade(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:  "grade_to_point",
		Grade: "Z",
	})
	if err == nil {
		t.Fatal("expected error for unknown grade")
	}
}

func TestGradeToPoint_InvalidScale(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:         "grade_to_point",
		Grade:        "A",
		GradingScale: "7",
	})
	if err == nil {
		t.Fatal("expected error for invalid gradingScale")
	}
}

// ============================================================
// Mode 6 — target_gpa
// ============================================================

func TestTargetGPA_Achievable(t *testing.T) {
	// Current CGPA 7.5 over 60 credits. Target 8.0 over 100 total.
	// Required in remaining 40 credits = (8.0×100 − 7.5×60) / 40
	//                                = (800 − 450) / 40 = 350/40 = 8.75
	res, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      7.5,
		CompletedCredits: 60,
		TargetCGPA:       8.0,
		RemainingCredits: 40,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !almostEqualGPA(res.RequiredGPA, 8.75, 0.01) {
		t.Fatalf("expected required GPA 8.75, got %v", res.RequiredGPA)
	}
	if res.TotalCredits != 100 {
		t.Fatalf("expected totalCredits 100, got %v", res.TotalCredits)
	}
	if res.Extra["achievable"] != true {
		t.Fatalf("expected achievable=true, got %v", res.Extra["achievable"])
	}
	if _, warned := res.Extra["warning"]; warned {
		t.Fatal("did not expect a warning for an achievable target")
	}
}

// A student with 9.0 CGPA over 60 credits who targets 8.0 over 100 total
// credits still needs 6.5 GPA in their remaining 40 credits — the low
// remaining-credit performance is what pulls the average down toward the
// target.
func TestTargetGPA_LowerTargetThanCurrent(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      9.0,
		CompletedCredits: 60,
		TargetCGPA:       8.0,
		RemainingCredits: 40,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.RequiredGPA, 6.5, 0.01) {
		t.Fatalf("expected requiredGPA 6.5, got %v", res.RequiredGPA)
	}
	if res.Extra["achievable"] != true {
		t.Fatalf("expected achievable=true, got %v", res.Extra["achievable"])
	}
	if _, warned := res.Extra["warning"]; warned {
		t.Fatal("did not expect a warning")
	}
}

// When the current weighted score already exceeds the target's total,
// the required GPA clamps to zero — the student has nothing left to do.
//
// currentCGPA × completedCredits = 9.0 × 100 = 900
// targetCGPA  × totalCredits     = 8.0 × 110 = 880
// 900 > 880, so the target is already met.
func TestTargetGPA_AlreadyExceeded_ClampsToZero(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      9.0,
		CompletedCredits: 100,
		TargetCGPA:       8.0,
		RemainingCredits: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequiredGPA != 0 {
		t.Fatalf("expected requiredGPA 0, got %v", res.RequiredGPA)
	}
	if res.Extra["achievable"] != true {
		t.Fatalf("expected achievable=true, got %v", res.Extra["achievable"])
	}
}

func TestTargetGPA_Impossible(t *testing.T) {
	// Current CGPA 6.0 over 100 credits. Target 9.0 over 110 total.
	// Required in remaining 10 = (9×110 − 6×100)/10 = (990 − 600)/10 = 39
	// Way above the 10-point scale → not achievable.
	res, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      6.0,
		CompletedCredits: 100,
		TargetCGPA:       9.0,
		RemainingCredits: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.RequiredGPA <= 10 {
		t.Fatalf("expected requiredGPA > 10, got %v", res.RequiredGPA)
	}
	if res.Extra["achievable"] != false {
		t.Fatalf("expected achievable=false, got %v", res.Extra["achievable"])
	}
	if _, warned := res.Extra["warning"]; !warned {
		t.Fatal("expected a warning for an impossible target")
	}
}

func TestTargetGPA_CustomScale(t *testing.T) {
	// 4-point scale. Current 3.5 over 60 credits. Target 3.8 over 100 total.
	// Required = (3.8×100 − 3.5×60) / 40 = (380 − 210)/40 = 4.25 > 4 → impossible.
	res, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      3.5,
		CompletedCredits: 60,
		TargetCGPA:       3.8,
		RemainingCredits: 40,
		Scale:            4.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !almostEqualGPA(res.RequiredGPA, 4.25, 0.01) {
		t.Fatalf("expected requiredGPA 4.25, got %v", res.RequiredGPA)
	}
	if res.Extra["achievable"] != false {
		t.Fatalf("expected achievable=false on 4-point scale, got %v", res.Extra["achievable"])
	}
}

func TestTargetGPA_ZeroRemainingCredits(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      7.0,
		CompletedCredits: 60,
		TargetCGPA:       8.0,
		RemainingCredits: 0,
	})
	if err == nil {
		t.Fatal("expected error for zero remaining credits")
	}
}

func TestTargetGPA_NegativeRemainingCredits(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      7.0,
		CompletedCredits: 60,
		TargetCGPA:       8.0,
		RemainingCredits: -10,
	})
	if err == nil {
		t.Fatal("expected error for negative remaining credits")
	}
}

func TestTargetGPA_CurrentCGPAAboveScale(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      11,
		CompletedCredits: 60,
		TargetCGPA:       9,
		RemainingCredits: 40,
	})
	if err == nil {
		t.Fatal("expected error for currentCGPA > scale")
	}
}

func TestTargetGPA_TargetCGPAAboveScale(t *testing.T) {
	_, err := CalculateGPA(GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      7,
		CompletedCredits: 60,
		TargetCGPA:       11,
		RemainingCredits: 40,
	})
	if err == nil {
		t.Fatal("expected error for targetCGPA > scale")
	}
}

// ============================================================
// Response shape consistency
// ============================================================

func TestGPA_ResponseAlwaysHasFormattedAndSteps(t *testing.T) {
	// Every mode must populate `Formatted` and non-empty `Steps`.
	requests := []GPARequest{
		{
			Mode:    "semester_gpa",
			Courses: []GPACourse{{Credits: 3, GradePoint: 9}},
		},
		{
			Mode:      "cumulative_cgpa",
			Semesters: []GPASemester{{Credits: 20, GPA: 8}},
		},
		{
			Mode: "cgpa_to_percentage",
			CGPA: 8,
		},
		{
			Mode:       "percentage_to_cgpa",
			Percentage: 76,
		},
		{
			Mode:  "grade_to_point",
			Grade: "A",
		},
		{
			Mode:             "target_gpa",
			CurrentCGPA:      7,
			CompletedCredits: 60,
			TargetCGPA:       8,
			RemainingCredits: 40,
		},
	}

	for _, req := range requests {
		res, err := CalculateGPA(req)
		if err != nil {
			t.Fatalf("mode %q: unexpected error: %v", req.Mode, err)
		}
		if res.Formatted == "" {
			t.Errorf("mode %q: empty Formatted", req.Mode)
		}
		if len(res.Steps) == 0 {
			t.Errorf("mode %q: empty Steps", req.Mode)
		}
	}
}

func TestGPA_RejectsNaNInEveryNumericField(t *testing.T) {
	nan := math.NaN()

	cases := []struct {
		name string
		req  GPARequest
	}{
		{"semester_gpa credits", GPARequest{
			Mode:    "semester_gpa",
			Courses: []GPACourse{{Credits: nan, GradePoint: 9}},
		}},
		{"semester_gpa gradePoint", GPARequest{
			Mode:    "semester_gpa",
			Courses: []GPACourse{{Credits: 3, GradePoint: nan}},
		}},
		{"cumulative_cgpa credits", GPARequest{
			Mode:      "cumulative_cgpa",
			Semesters: []GPASemester{{Credits: nan, GPA: 8}},
		}},
		{"cumulative_cgpa gpa", GPARequest{
			Mode:      "cumulative_cgpa",
			Semesters: []GPASemester{{Credits: 20, GPA: nan}},
		}},
		{"cgpa_to_percentage cgpa", GPARequest{
			Mode: "cgpa_to_percentage",
			CGPA: nan,
		}},
		{"percentage_to_cgpa percentage", GPARequest{
			Mode:       "percentage_to_cgpa",
			Percentage: nan,
		}},
		{"target_gpa currentCgpa", GPARequest{
			Mode:             "target_gpa",
			CurrentCGPA:      nan,
			CompletedCredits: 60,
			TargetCGPA:       8,
			RemainingCredits: 40,
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := CalculateGPA(c.req); err == nil {
				t.Fatalf("expected error for %s", c.name)
			}
		})
	}
}

// ============================================================
// Multiplier resolution
// ============================================================

func TestCGPAToPercentage_MultiplierDefaultsTo9_5WhenZero(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode:       "cgpa_to_percentage",
		CGPA:       8,
		Multiplier: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Multiplier != 9.5 {
		t.Fatalf("expected default multiplier 9.5, got %v", res.Multiplier)
	}
}

func TestCGPAToPercentage_ScaleNoteForCBSE(t *testing.T) {
	res, err := CalculateGPA(GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: 8,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	note, ok := res.Extra["scaleNote"].(string)
	if !ok || note == "" {
		t.Fatalf("expected non-empty scaleNote, got %v", res.Extra["scaleNote"])
	}
}