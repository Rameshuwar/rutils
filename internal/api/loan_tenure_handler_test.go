package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postLoanTenureJSON(t *testing.T, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/calculate-loan-tenure", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	HandleLoanTenureCalculate(rec, req)
	return rec
}

func TestLoanTenureHandler_Success(t *testing.T) {
	rec := postLoanTenureJSON(t, LoanTenureRequest{
		Principal:          500000,
		AnnualInterestRate: 10.5,
		MonthlyPayment:     15000,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res["emi"] != float64(15000) {
		t.Errorf("expected emi=15000, got %v", res["emi"])
	}
	if _, ok := res["tenureMonths"]; !ok {
		t.Error("response must include tenureMonths")
	}
	if _, ok := res["totalInterest"]; !ok {
		t.Error("response must include totalInterest")
	}
	if _, ok := res["amortization"]; !ok {
		t.Error("response must include amortization")
	}
}

func TestLoanTenureHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/calculate-loan-tenure", nil)
	rec := httptest.NewRecorder()
	HandleLoanTenureCalculate(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Only POST method is allowed") {
		t.Fatalf("expected method-not-allowed message, got %q", rec.Body.String())
	}
}

func TestLoanTenureHandler_MalformedJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/calculate-loan-tenure",
		strings.NewReader("this is not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	HandleLoanTenureCalculate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Invalid JSON") {
		t.Fatalf("expected 'Invalid JSON' message, got %q", rec.Body.String())
	}
}

func TestLoanTenureHandler_EmptyBody(t *testing.T) {
	// Zero principal → 400
	rec := postLoanTenureJSON(t, LoanTenureRequest{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", rec.Code)
	}
}

func TestLoanTenureHandler_PaymentTooLow(t *testing.T) {
	rec := postLoanTenureJSON(t, LoanTenureRequest{
		Principal:          100000,
		AnnualInterestRate: 12,
		MonthlyPayment:     500, // less than the 1000 monthly interest
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "too low") {
		t.Fatalf("expected 'too low' message, got %q", rec.Body.String())
	}
}

func TestLoanTenureHandler_ZeroInterest(t *testing.T) {
	rec := postLoanTenureJSON(t, LoanTenureRequest{
		Principal:          120000,
		AnnualInterestRate: 0,
		MonthlyPayment:     10000,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&res)

	if res["tenureMonths"] != float64(12) {
		t.Errorf("expected tenureMonths=12, got %v", res["tenureMonths"])
	}
	if res["totalInterest"] != float64(0) {
		t.Errorf("expected zero interest, got %v", res["totalInterest"])
	}
}

func TestLoanTenureHandler_ResponseShapeMatchesEMI(t *testing.T) {
	// The response shape must be a superset of the existing EMI response
	// so the frontend can reuse the same rendering code.
	rec := postLoanTenureJSON(t, LoanTenureRequest{
		Principal:          200000,
		AnnualInterestRate: 8,
		MonthlyPayment:     10000,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&res)

	required := []string{
		"emi", "principal", "totalInterest", "totalPayment",
		"tenureMonths", "monthlyRatePercent", "breakdown", "amortization",
	}
	for _, k := range required {
		if _, ok := res[k]; !ok {
			t.Errorf("response is missing required key %q", k)
		}
	}

	// breakdown sub-keys
	breakdown, ok := res["breakdown"].(map[string]any)
	if !ok {
		t.Fatal("breakdown must be an object")
	}
	for _, k := range []string{"principalPercent", "interestPercent"} {
		if _, ok := breakdown[k]; !ok {
			t.Errorf("breakdown missing key %q", k)
		}
	}
}