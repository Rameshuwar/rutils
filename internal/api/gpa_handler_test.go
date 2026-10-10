package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"file-converter/internal/converter"    // ← line 11 (was already there)
	                                           //    (blank line added by sed    // ← line 13 (added by sed)
)

// ---------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------

// postGPAJSON marshals body, POSTs it to /calculate-gpa, and returns the
// httptest recorder. Any marshal failure aborts the test.
func postGPAJSON(t *testing.T, body any) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/calculate-gpa", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	HandleGPACalculate(rec, req)
	return rec
}

// postGPARaw POSTs an arbitrary byte slice to /calculate-gpa. Used for
// malformed-body tests where we don't want the marshaller to "help" us.
func postGPARaw(t *testing.T, raw string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/calculate-gpa", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	HandleGPACalculate(rec, req)
	return rec
}

// =====================================================================
// Method guard
// =====================================================================

func TestGPAHandler_MethodNotAllowed_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/calculate-gpa", nil)
	rec := httptest.NewRecorder()

	HandleGPACalculate(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Only POST method is allowed") {
		t.Fatalf("expected method-not-allowed message, got %q", rec.Body.String())
	}
}

func TestGPAHandler_MethodNotAllowed_PUT(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/calculate-gpa", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	HandleGPACalculate(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for PUT, got %d", rec.Code)
	}
}

func TestGPAHandler_MethodNotAllowed_DELETE(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/calculate-gpa", nil)
	rec := httptest.NewRecorder()

	HandleGPACalculate(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for DELETE, got %d", rec.Code)
	}
}

// =====================================================================
// Body parsing
// =====================================================================

func TestGPAHandler_MalformedJSON(t *testing.T) {
	rec := postGPARaw(t, "this is not json")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Invalid JSON") {
		t.Fatalf("expected 'Invalid JSON' message, got %q", rec.Body.String())
	}
}

func TestGPAHandler_EmptyBody(t *testing.T) {
	// An empty body is valid JSON-wise? No — the decoder returns EOF.
	rec := postGPARaw(t, "")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", rec.Code)
	}
}

func TestGPAHandler_ValidJSONButEmptyObject(t *testing.T) {
	// `{}` decodes fine but has an empty mode, so the converter rejects it.
	rec := postGPAJSON(t, map[string]any{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing mode, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "mode is required") {
		t.Fatalf("expected 'mode is required' message, got %q", rec.Body.String())
	}
}

// =====================================================================
// Happy path — one test per mode
// =====================================================================

func TestGPAHandler_SemesterGPA(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode: "semester_gpa",
		Courses: []converter.GPACourse{
			{Name: "Math", Credits: 4, GradePoint: 9},
			{Name: "Physics", Credits: 3, GradePoint: 8},
			{Name: "Chemistry", Credits: 3, GradePoint: 10},
		},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	var res converter.GPAResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Mode != "semester_gpa" {
		t.Fatalf("expected mode=semester_gpa, got %q", res.Mode)
	}
	if res.GPA != 9 {
		t.Fatalf("expected GPA 9, got %v", res.GPA)
	}
	if len(res.Breakdown) != 3 {
		t.Fatalf("expected 3 breakdown rows, got %d", len(res.Breakdown))
	}
	if res.Formatted == "" {
		t.Fatal("expected non-empty Formatted")
	}
	if len(res.Steps) == 0 {
		t.Fatal("expected non-empty Steps")
	}

	// Content-Type must be JSON.
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected JSON content-type, got %q", ct)
	}
}

func TestGPAHandler_CumulativeCGPA(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode: "cumulative_cgpa",
		Semesters: []converter.GPASemester{
			{Name: "Sem 1", Credits: 20, GPA: 8},
			{Name: "Sem 2", Credits: 20, GPA: 9},
			{Name: "Sem 3", Credits: 20, GPA: 9.5},
		},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	var res converter.GPAResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Mode != "cumulative_cgpa" {
		t.Fatalf("expected mode=cumulative_cgpa, got %q", res.Mode)
	}
	if res.CGPA < 8.8 || res.CGPA > 8.9 {
		t.Fatalf("expected CGPA ≈ 8.83, got %v", res.CGPA)
	}
	if len(res.Semesters) != 3 {
		t.Fatalf("expected 3 semester rows, got %d", len(res.Semesters))
	}
}

func TestGPAHandler_CGPAToPercentage_Default9_5(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: 8.0,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	var res converter.GPAResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Percentage != 76 {
		t.Fatalf("expected percentage 76, got %v", res.Percentage)
	}
	if res.Multiplier != 9.5 {
		t.Fatalf("expected multiplier 9.5, got %v", res.Multiplier)
	}
	if res.Formatted != "76%" {
		t.Fatalf("expected formatted=76%%, got %q", res.Formatted)
	}
}

func TestGPAHandler_CGPAToPercentage_CustomMultiplier(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:       "cgpa_to_percentage",
		CGPA:       8.4,
		Multiplier: 10,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res converter.GPAResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res.Percentage != 84 {
		t.Fatalf("expected 84, got %v", res.Percentage)
	}
}

func TestGPAHandler_PercentageToCGPA(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: 76,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	var res converter.GPAResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res.CGPA != 8.0 {
		t.Fatalf("expected CGPA 8.0, got %v", res.CGPA)
	}
}

func TestGPAHandler_GradeToPoint_TenPointScale(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:         "grade_to_point",
		Grade:        "A",
		GradingScale: "10",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	var res converter.GPAResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res.GradePoint != 9 {
		t.Fatalf("expected gradePoint 9, got %v", res.GradePoint)
	}
}

func TestGPAHandler_GradeToPoint_FourPointScale(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:         "grade_to_point",
		Grade:        "A-",
		GradingScale: "4",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res converter.GPAResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res.GradePoint != 3.7 {
		t.Fatalf("expected gradePoint 3.7, got %v", res.GradePoint)
	}
}

func TestGPAHandler_TargetGPA_Achievable(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      7.5,
		CompletedCredits: 60,
		TargetCGPA:       8.0,
		RemainingCredits: 40,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	var res converter.GPAResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res.RequiredGPA != 8.75 {
		t.Fatalf("expected requiredGPA 8.75, got %v", res.RequiredGPA)
	}
	if res.Extra["achievable"] != true {
		t.Fatalf("expected achievable=true, got %v", res.Extra["achievable"])
	}
}

func TestGPAHandler_TargetGPA_ImpossibleCarriesWarning(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      6.0,
		CompletedCredits: 100,
		TargetCGPA:       9.0,
		RemainingCredits: 10,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	var res converter.GPAResponse
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res.Extra["achievable"] != false {
		t.Fatalf("expected achievable=false, got %v", res.Extra["achievable"])
	}
	if _, hasWarning := res.Extra["warning"]; !hasWarning {
		t.Fatal("expected a warning in extra for an impossible target")
	}
}

// =====================================================================
// Validation errors → 400
// =====================================================================

func TestGPAHandler_UnsupportedMode(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{Mode: "magic"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unsupported mode") {
		t.Fatalf("expected 'unsupported mode' message, got %q", rec.Body.String())
	}
}

func TestGPAHandler_SemesterGPA_EmptyCourseList(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{Mode: "semester_gpa"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "courses are required") {
		t.Fatalf("expected 'courses are required' message, got %q", rec.Body.String())
	}
}

func TestGPAHandler_SemesterGPA_ZeroCredits(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode: "semester_gpa",
		Courses: []converter.GPACourse{
			{Credits: 0, GradePoint: 9},
		},
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero credits, got %d", rec.Code)
	}
}

func TestGPAHandler_CumulativeCGPA_EmptySemesterList(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{Mode: "cumulative_cgpa"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "semesters are required") {
		t.Fatalf("expected 'semesters are required' message, got %q", rec.Body.String())
	}
}

func TestGPAHandler_CGPAToPercentage_NegativeCGPA(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: -1,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGPAHandler_CGPAToPercentage_Above10(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: 11,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for CGPA > 10, got %d", rec.Code)
	}
}

func TestGPAHandler_PercentageToCGPA_Above100(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:       "percentage_to_cgpa",
		Percentage: 101,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for percentage > 100, got %d", rec.Code)
	}
}

func TestGPAHandler_GradeToPoint_EmptyGrade(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{Mode: "grade_to_point"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "grade is required") {
		t.Fatalf("expected 'grade is required' message, got %q", rec.Body.String())
	}
}

func TestGPAHandler_GradeToPoint_UnknownGrade(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:  "grade_to_point",
		Grade: "Z",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown grade, got %d", rec.Code)
	}
}

func TestGPAHandler_GradeToPoint_InvalidScale(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:         "grade_to_point",
		Grade:        "A",
		GradingScale: "7",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid scale, got %d", rec.Code)
	}
}

func TestGPAHandler_TargetGPA_ZeroRemainingCredits(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      7.0,
		CompletedCredits: 60,
		TargetCGPA:       8.0,
		RemainingCredits: 0,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero remaining credits, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "remainingCredits") {
		t.Fatalf("expected 'remainingCredits' in error, got %q", rec.Body.String())
	}
}

// =====================================================================
// Response shape integrity
// =====================================================================

func TestGPAHandler_ResponseIsJSONContenthType(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode: "cgpa_to_percentage",
		CGPA: 8,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected JSON content type, got %q", ct)
	}
}

func TestGPAHandler_ResponseShapeContainsFormattedAndSteps(t *testing.T) {
	// Every successful response must carry `formatted` and a non-empty
	// `steps` array — the frontend depends on both.
	modes := []GPARequest{
		{
			Mode:    "semester_gpa",
			Courses: []converter.GPACourse{{Credits: 3, GradePoint: 9}},
		},
		{
			Mode:      "cumulative_cgpa",
			Semesters: []converter.GPASemester{{Credits: 20, GPA: 8}},
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

	for _, req := range modes {
		rec := postGPAJSON(t, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("mode %q: expected 200, got %d. body=%s",
				req.Mode, rec.Code, rec.Body.String())
		}

		var res converter.GPAResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("mode %q: failed to decode: %v", req.Mode, err)
		}
		if res.Formatted == "" {
			t.Errorf("mode %q: empty formatted field", req.Mode)
		}
		if len(res.Steps) == 0 {
			t.Errorf("mode %q: empty steps array", req.Mode)
		}
		if res.Mode != req.Mode {
			t.Errorf("mode %q: response.mode=%q does not match request",
				req.Mode, res.Mode)
		}
	}
}
// =====================================================================
// Wire-format regression tests
//
// The tests above decode responses into `converter.GPAResponse` structs
// and assert on the Go field values. That approach cannot catch bugs in
// the JSON struct tags themselves — e.g. an `omitempty` that swallows a
// legitimately-zero field. The tests below parse the raw response bytes
// so the wire format is actually exercised.
// =====================================================================

// requiredGpa must be present in the JSON even when its value is exactly
// zero. This is the clamp-to-zero case for the target_gpa mode.
//
// Without this test, a stray `,omitempty` on the struct tag would silently
// strip the field from the response, and no test that decodes into the
// typed struct would notice. Only a persona test (or this) would catch it.
func TestGPAHandler_TargetGPA_ZeroRequiredGPAIsEmittedInJSON(t *testing.T) {
	rec := postGPAJSON(t, GPARequest{
		Mode:             "target_gpa",
		CurrentCGPA:      9.0,
		CompletedCredits: 100,
		TargetCGPA:       8.0,
		RemainingCredits: 10,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	// Parse into a generic map so we're testing the JSON structure,
	// not the Go struct. The presence check below would fail if
	// `requiredGpa` had been tagged with `omitempty`.
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	v, present := raw["requiredGpa"]
	if !present {
		t.Fatalf("requiredGpa key missing from JSON — the `omitempty` tag was probably reintroduced on the struct field")
	}
	if v != float64(0) {
		t.Fatalf("expected requiredGpa=0 in JSON, got %v (%T)", v, v)
	}
}

// Every mode that reports a computed numeric field must emit that field
// even when its value is zero. This generalizes the check above for a
// couple of the most likely offenders.
func TestGPAHandler_ZeroValuesAreEmittedInJSON(t *testing.T) {
	cases := []struct {
		name       string
		req        GPARequest
		assertKeys []string
	}{
		{
			name: "semester_gpa with all-zero grade points",
			req: GPARequest{
				Mode:    "semester_gpa",
				Courses: []converter.GPACourse{{Credits: 3, GradePoint: 0}},
			},
			assertKeys: []string{"gpa", "totalCredits", "formatted"},
		},
		{
			name: "cgpa_to_percentage with cgpa=0",
			req: GPARequest{
				Mode: "cgpa_to_percentage",
				CGPA: 0,
			},
			assertKeys: []string{"cgpa", "percentage", "formatted"},
		},
		{
			name: "grade_to_point for grade F (zero points)",
			req: GPARequest{
				Mode:         "grade_to_point",
				Grade:        "F",
				GradingScale: "10",
			},
			assertKeys: []string{"gradePoint", "formatted"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := postGPAJSON(t, c.req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
			}

			var raw map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			for _, key := range c.assertKeys {
				if _, present := raw[key]; !present {
					t.Errorf("key %q missing from JSON (value may have been swallowed by omitempty). Full body: %s",
						key, rec.Body.String())
				}
			}
		})
	}
}
