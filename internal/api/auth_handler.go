package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"file-converter/internal/auth"
)

// AuthHandler handles HTTP requests for user authentication and account management
type AuthHandler struct {
	service *auth.AuthService
	cfg     *auth.Config
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(service *auth.AuthService, cfg *auth.Config) *AuthHandler {
	return &AuthHandler{
		service: service,
		cfg:     cfg,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Register creates a new user account
// @Summary Register a new user
// @Description Registers a new user with name, email, password, and confirm_password. Requires strong password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "User Registration Information"
// @Success 201 {object} auth.UserResponse "User registered successfully"
// @Failure 400 {object} map[string]string "Validation error or weak password"
// @Failure 409 {object} map[string]string "Email already registered"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		if errors.Is(err, auth.ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// Login authenticates a user and returns a JWT token
// @Summary Login with email and password
// @Description Authenticates user credentials and returns a signed JWT token. If logged in with a temporary password, must_reset_password will be true.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "User Login Credentials"
// @Success 200 {object} auth.LoginResponse "Login successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	res, err := h.service.Login(req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// Logout invalidates the client-side session
// @Summary Logout user
// @Description Clears the user's authentication session
// @Tags Authentication
// @Produce json
// @Success 200 {object} auth.MessageResponse "Logout successful"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, auth.MessageResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// ForgotPassword sends a temporary password via SMTP
// @Summary Send temporary password via Google SMTP
// @Description Sends a temporary password to the user's email. When used to log in, forces password reset.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body auth.ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} auth.MessageResponse "Temporary password dispatched"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Account not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req auth.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	if err := h.service.ForgotPassword(req); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "No account found with this email address")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, auth.MessageResponse{
		Success: true,
		Message: "A temporary password has been dispatched to your email address.",
	})
}

// ResetPassword sets a permanent password after logging in with a temporary password
// @Summary Reset password
// @Description Sets a permanent password for a user logged in with a temporary password. Requires Bearer JWT.
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body auth.ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} auth.MessageResponse "Password updated successfully"
// @Failure 400 {object} map[string]string "Validation error or weak password"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := GetUserClaims(r.Context())
	if !ok || claims.UserID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req auth.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	if err := h.service.ResetPassword(claims.UserID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, auth.MessageResponse{
		Success: true,
		Message: "Permanent password updated successfully. You can now use your new password.",
	})
}

// ChangePassword changes the password for an authenticated user
// @Summary Change password for logged-in user
// @Description Changes the password for the currently authenticated user and sends an email confirmation. Requires Bearer JWT.
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body auth.ChangePasswordRequest true "Change Password Request"
// @Success 200 {object} auth.MessageResponse "Password changed successfully"
// @Failure 400 {object} map[string]string "Validation error or weak password"
// @Failure 401 {object} map[string]string "Current password incorrect or unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := GetUserClaims(r.Context())
	if !ok || claims.UserID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req auth.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	if err := h.service.ChangePassword(claims.UserID, req); err != nil {
		if errors.Is(err, auth.ErrCurrentPasswordInvalid) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, auth.MessageResponse{
		Success: true,
		Message: "Password changed successfully. A confirmation email has been dispatched.",
	})
}

// Me returns the profile of the currently logged-in user
// @Summary Get logged-in user profile
// @Description Returns the profile details for the authenticated user. Requires Bearer JWT.
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} auth.UserResponse "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	claims, ok := GetUserClaims(r.Context())
	if !ok || claims.UserID == "" {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.service.GetUserProfile(claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}
