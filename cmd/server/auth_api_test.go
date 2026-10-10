package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"file-converter/internal/api"
	"file-converter/internal/auth"
)

func setupTestServer(t *testing.T) (*httptest.Server, *auth.Store, *auth.Config) {
	tempDir, err := os.MkdirTemp("", "auth-api-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	cfg := &auth.Config{
		SMTPEmail:          "",
		SMTPAppPassword:    "",
		SMTPHost:           "smtp.gmail.com",
		SMTPPort:           "587",
		JWTSecret:          "test-secret-key-1234567890",
		JWTExpirationHours: 24,
		DataDir:            tempDir,
	}

	store, err := auth.NewStore(filepath.Join(tempDir, "users.json"))
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}

	emailSender := auth.NewEmailSender(cfg)
	authService := auth.NewAuthService(store, cfg, emailSender)
	authHandler := api.NewAuthHandler(authService, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/logout", authHandler.Logout)
	mux.HandleFunc("/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("/auth/reset-password", api.RequireAuth(cfg, authHandler.ResetPassword))
	mux.HandleFunc("/auth/change-password", api.RequireAuth(cfg, authHandler.ChangePassword))
	mux.HandleFunc("/auth/me", api.RequireAuth(cfg, authHandler.Me))

	ts := httptest.NewServer(api.EnableCORS(mux))
	t.Cleanup(ts.Close)

	return ts, store, cfg
}

func postJSON(t *testing.T, url string, token string, body any) (*http.Response, map[string]any) {
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var resBody map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&resBody)
	resp.Body.Close()

	return resp, resBody
}

func getWithAuth(t *testing.T, url string, token string) (*http.Response, map[string]any) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("failed to create GET request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var resBody map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&resBody)
	resp.Body.Close()

	return resp, resBody
}

func TestAuthEndToEnd(t *testing.T) {
	ts, store, _ := setupTestServer(t)

	// 1. Register with weak password -> Expect 400 Bad Request
	{
		resp, body := postJSON(t, ts.URL+"/auth/register", "", auth.RegisterRequest{
			Name:            "Bob Ross",
			Email:           "bob@example.com",
			Password:        "weak",
			ConfirmPassword: "weak",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for weak password, got %d. body=%v", resp.StatusCode, body)
		}
	}

	// 2. Register with mismatched password -> Expect 400 Bad Request
	{
		resp, body := postJSON(t, ts.URL+"/auth/register", "", auth.RegisterRequest{
			Name:            "Bob Ross",
			Email:           "bob@example.com",
			Password:        "P@ssword123!",
			ConfirmPassword: "DifferentP@ssword123!",
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for password mismatch, got %d. body=%v", resp.StatusCode, body)
		}
	}

	// 3. Register valid user -> Expect 201 Created
	{
		resp, body := postJSON(t, ts.URL+"/auth/register", "", auth.RegisterRequest{
			Name:            "Bob Ross",
			Email:           "bob@example.com",
			Password:        "HappyLittleClouds#1",
			ConfirmPassword: "HappyLittleClouds#1",
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 for valid registration, got %d. body=%v", resp.StatusCode, body)
		}
		if body["email"] != "bob@example.com" {
			t.Fatalf("expected email bob@example.com, got %v", body["email"])
		}
	}

	// 4. Duplicate registration -> Expect 409 Conflict
	{
		resp, body := postJSON(t, ts.URL+"/auth/register", "", auth.RegisterRequest{
			Name:            "Bob Ross Twin",
			Email:           "bob@example.com",
			Password:        "HappyLittleClouds#1",
			ConfirmPassword: "HappyLittleClouds#1",
		})
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409 for duplicate email, got %d. body=%v", resp.StatusCode, body)
		}
	}

	// 5. Login with wrong password -> Expect 401 Unauthorized
	{
		resp, _ := postJSON(t, ts.URL+"/auth/login", "", auth.LoginRequest{
			Email:    "bob@example.com",
			Password: "WrongPassword#999",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 for bad password, got %d", resp.StatusCode)
		}
	}

	// 6. Login valid -> Expect 200 OK and JWT
	var token string
	{
		resp, body := postJSON(t, ts.URL+"/auth/login", "", auth.LoginRequest{
			Email:    "bob@example.com",
			Password: "HappyLittleClouds#1",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for valid login, got %d. body=%v", resp.StatusCode, body)
		}
		tok, ok := body["token"].(string)
		if !ok || tok == "" {
			t.Fatalf("expected JWT token string in login response, got %v", body["token"])
		}
		token = tok

		if body["must_reset_password"] != false {
			t.Fatalf("expected must_reset_password false, got %v", body["must_reset_password"])
		}
	}

	// 7. Verify /auth/me with Bearer token -> Expect 200 OK
	{
		resp, body := getWithAuth(t, ts.URL+"/auth/me", token)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for /auth/me, got %d. body=%v", resp.StatusCode, body)
		}
		if body["email"] != "bob@example.com" {
			t.Fatalf("expected email bob@example.com, got %v", body["email"])
		}
	}

	// 8. Verify /auth/me without token -> Expect 401 Unauthorized
	{
		resp, _ := getWithAuth(t, ts.URL+"/auth/me", "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 for unauthenticated /auth/me, got %d", resp.StatusCode)
		}
	}

	// 9. Change Password (valid) -> Expect 200 OK
	{
		resp, body := postJSON(t, ts.URL+"/auth/change-password", token, auth.ChangePasswordRequest{
			CurrentPassword: "HappyLittleClouds#1",
			NewPassword:     "BrandNewMasterpiece!2026",
			ConfirmPassword: "BrandNewMasterpiece!2026",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for change password, got %d. body=%v", resp.StatusCode, body)
		}
	}

	// 10. Login with old password should fail, new password should succeed
	{
		resp, _ := postJSON(t, ts.URL+"/auth/login", "", auth.LoginRequest{
			Email:    "bob@example.com",
			Password: "HappyLittleClouds#1",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 for old password, got %d", resp.StatusCode)
		}

		respNew, bodyNew := postJSON(t, ts.URL+"/auth/login", "", auth.LoginRequest{
			Email:    "bob@example.com",
			Password: "BrandNewMasterpiece!2026",
		})
		if respNew.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for new password, got %d. body=%v", respNew.StatusCode, bodyNew)
		}
		token = bodyNew["token"].(string)
	}

	// 11. Forgot Password -> Expect 200 OK
	{
		resp, body := postJSON(t, ts.URL+"/auth/forgot-password", "", auth.ForgotPasswordRequest{
			Email: "bob@example.com",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for forgot-password, got %d. body=%v", resp.StatusCode, body)
		}
	}

	// Fetch user from store to read the generated temporary password for automated test
	user, err := store.GetByEmail("bob@example.com")
	if err != nil {
		t.Fatalf("failed to fetch user from store: %v", err)
	}
	if user.TempPasswordHash == "" {
		t.Fatalf("expected user to have TempPasswordHash set")
	}

	// 11b. Verify that logging in with the old permanent password is BLOCKED after forgot password
	{
		resp, body := postJSON(t, ts.URL+"/auth/login", "", auth.LoginRequest{
			Email:    "bob@example.com",
			Password: "BrandNewMasterpiece!2026",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 when logging in with old password after forgot-password, got %d. body=%v", resp.StatusCode, body)
		}
	}

	// 12. Test login with temporary password
	// Since the dev email sender logs to stdout, we can verify that the user can login with their temp password
	// In order to get the exact raw temp password in our test, let's set a known valid temp password
	knownTemp := "Temp#Secure99!x"
	tempHash, _ := auth.HashPassword(knownTemp)
	user.TempPasswordHash = tempHash
	user.MustResetPassword = true
	_ = store.Update(user)

	var tempSessionToken string
	{
		resp, body := postJSON(t, ts.URL+"/auth/login", "", auth.LoginRequest{
			Email:    "bob@example.com",
			Password: knownTemp,
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for temp password login, got %d. body=%v", resp.StatusCode, body)
		}
		if body["must_reset_password"] != true {
			t.Fatalf("expected must_reset_password true, got %v", body["must_reset_password"])
		}
		tempSessionToken = body["token"].(string)
	}

	// 13. Call /auth/reset-password with temp session token -> Expect 200 OK
	{
		resp, body := postJSON(t, ts.URL+"/auth/reset-password", tempSessionToken, auth.ResetPasswordRequest{
			NewPassword:     "FinalFreshPassword@777",
			ConfirmPassword: "FinalFreshPassword@777",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for reset-password, got %d. body=%v", resp.StatusCode, body)
		}
	}

	// 14. Verify user can log in with FinalFreshPassword@777 and must_reset_password is false
	{
		resp, body := postJSON(t, ts.URL+"/auth/login", "", auth.LoginRequest{
			Email:    "bob@example.com",
			Password: "FinalFreshPassword@777",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for login after reset, got %d. body=%v", resp.StatusCode, body)
		}
		if body["must_reset_password"] != false {
			t.Fatalf("expected must_reset_password false, got %v", body["must_reset_password"])
		}
	}

	// 15. Logout -> Expect 200 OK
	{
		resp, body := postJSON(t, ts.URL+"/auth/logout", "", nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for logout, got %d. body=%v", resp.StatusCode, body)
		}
		if body["success"] != true {
			t.Fatalf("expected success true in logout response, got %v", body["success"])
		}
	}
}

func TestCORSHeadersOnAuth(t *testing.T) {
	ts, _, _ := setupTestServer(t)

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/auth/login", nil)
	if err != nil {
		t.Fatalf("failed to create OPTIONS request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:5173")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Errorf("expected Access-Control-Allow-Headers to contain 'Authorization', got %q", resp.Header.Get("Access-Control-Allow-Headers"))
	}
}
