package api

import (
	"encoding/json"
	"net/http"

	"file-converter/internal/converter"
)

// GPARequest is the incoming JSON body for /calculate-gpa.
//
// It mirrors converter.GPARequest exactly, but is redeclared here so
// that swaggo generates a clean `api.GPARequest` definition in the
// OpenAPI spec instead of reaching into internal/converter types.
type GPARequest struct {
	Mode string `json:"mode"`

	// mode = "semester_gpa"
	Courses []converter.GPACourse `json:"courses,omitempty"`

	// mode = "cumulative_cgpa"
	Semesters []converter.GPASemester `json:"semesters,omitempty"`

	// mode = "cgpa_to_percentage" | "percentage_to_cgpa"
	CGPA       float64 `json:"cgpa,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
	Multiplier float64 `json:"multiplier,omitempty"`

	// mode = "grade_to_point"
	Grade        string `json:"grade,omitempty"`
	GradingScale string `json:"gradingScale,omitempty"`

	// mode = "target_gpa"
	CurrentCGPA      float64 `json:"currentCgpa,omitempty"`
	CompletedCredits float64 `json:"completedCredits,omitempty"`
	TargetCGPA       float64 `json:"targetCgpa,omitempty"`
	RemainingCredits float64 `json:"remainingCredits,omitempty"`
	Scale            float64 `json:"scale,omitempty"`
}

// HandleGPACalculate handles HTTP requests for all GPA / CGPA calculations.
//
// @Summary      GPA / CGPA Calculator
// @Description  Performs any GPA or CGPA calculation through a single `mode`-discriminated endpoint.
// @Description
// @Description  **Supported modes (`mode` field):**
// @Description  - `semester_gpa`        (courses[])                            — credit-weighted GPA for one semester
// @Description  - `cumulative_cgpa`     (semesters[])                          — credit-weighted CGPA across semesters
// @Description  - `cgpa_to_percentage`  (cgpa, multiplier?)                    — CGPA → percentage (default 9.5)
// @Description  - `percentage_to_cgpa`  (percentage, multiplier?)              — percentage → CGPA
// @Description  - `grade_to_point`      (grade, gradingScale?)                 — letter grade → numeric grade point
// @Description  - `target_gpa`          (currentCgpa, completedCredits,
// @Description                          targetCgpa, remainingCredits, scale?)  — required GPA to reach a target CGPA
// @Description
// @Description  **Mode details:**
// @Description
// @Description  **semester_gpa** — send a `courses` array where each item is:
// @Description  `{"name": "Math", "credits": 4, "gradePoint": 9}`.
// @Description  GPA = Σ(credits × gradePoint) / Σ(credits). The response
// @Description  includes a per-course breakdown and the formula steps.
// @Description
// @Description  **cumulative_cgpa** — send a `semesters` array where each item is:
// @Description  `{"name": "Sem 1", "credits": 20, "gpa": 8.5}`.
// @Description  CGPA = Σ(credits × gpa) / Σ(credits). The response includes
// @Description  a per-semester breakdown.
// @Description
// @Description  **cgpa_to_percentage** — converts a CGPA to a percentage using
// @Description  the supplied `multiplier`. Defaults to 9.5 (CBSE). Common
// @Description  alternatives: 10.0 (VTU/KTU), 7.25 (Mumbai University).
// @Description  The output is capped at 100%.
// @Description
// @Description  **percentage_to_cgpa** — inverse of the above.
// @Description
// @Description  **grade_to_point** — looks up the numeric grade point for a
// @Description  letter grade. Supported scales: `"10"` (default, CBSE/AICTE),
// @Description  `"4"` (US university), `"5"` (European secondary).
// @Description
// @Description  **target_gpa** — solves for the GPA the student must average
// @Description  in their remaining credits to reach a desired overall CGPA:
// @Description  `requiredGPA = (targetCGPA × totalCredits − currentCGPA × completedCredits) / remainingCredits`.
// @Description  The response always carries an `extra.achievable` boolean; if
// @Description  the required GPA exceeds the scale, `extra.warning` explains why.
// @Description
// @Description  **Limits (v1):**
// @Description  - Grade points, GPAs, and CGPAs must be in [0, 10] for modes
// @Description    1–4. Mode 6 accepts a custom `scale` (default 10).
// @Description  - Credits must be positive.
// @Description  - All numeric inputs must be finite (NaN / Inf rejected).
// @Tags         Calculators
// @Accept       json
// @Produce      json
// @Param        request body api.GPARequest true "GPA / CGPA Calculation Request"
// @Success      200 {object} converter.GPAResponse "GPA Result"
// @Failure      400 {string} string "Bad Request - invalid mode or parameters"
// @Router       /calculate-gpa [post]
func HandleGPACalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GPARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Translate the API-layer request into the converter-layer request.
	// The two structs are identical in shape; the translation is
	// explicit so the two packages can evolve independently.
	res, err := converter.CalculateGPA(converter.GPARequest{
		Mode:             req.Mode,
		Courses:          req.Courses,
		Semesters:        req.Semesters,
		CGPA:             req.CGPA,
		Percentage:       req.Percentage,
		Multiplier:       req.Multiplier,
		Grade:            req.Grade,
		GradingScale:     req.GradingScale,
		CurrentCGPA:      req.CurrentCGPA,
		CompletedCredits: req.CompletedCredits,
		TargetCGPA:       req.TargetCGPA,
		RemainingCredits: req.RemainingCredits,
		Scale:            req.Scale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}