package auth

import "time"

// User represents a stored user account in data/users.json
type User struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Email              string     `json:"email"`
	PasswordHash       string     `json:"password_hash"`
	TempPasswordHash   string     `json:"temp_password_hash,omitempty"`
	TempPasswordExpiry *time.Time `json:"temp_password_expiry,omitempty"`
	MustResetPassword  bool       `json:"must_reset_password"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// UserResponse is the safe public representation of a user without sensitive hashes
type UserResponse struct {
	ID                string    `json:"id" example:"u_984f1a23"`
	Name              string    `json:"name" example:"Jane Doe"`
	Email             string    `json:"email" example:"jane@example.com"`
	MustResetPassword bool      `json:"must_reset_password" example:"false"`
	CreatedAt         time.Time `json:"created_at"`
}

// ToResponse converts a User entity to a safe UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                u.ID,
		Name:              u.Name,
		Email:             u.Email,
		MustResetPassword: u.MustResetPassword,
		CreatedAt:         u.CreatedAt,
	}
}

// RegisterRequest represents the payload for user registration
type RegisterRequest struct {
	Name            string `json:"name" example:"Jane Doe"`
	Email           string `json:"email" example:"jane@example.com"`
	Password        string `json:"password" example:"StrongP@ssw0rd!"`
	ConfirmPassword string `json:"confirm_password" example:"StrongP@ssw0rd!"`
}

// LoginRequest represents the payload for user login
type LoginRequest struct {
	Email    string `json:"email" example:"jane@example.com"`
	Password string `json:"password" example:"StrongP@ssw0rd!"`
}

// LoginResponse contains the JWT token and user profile
type LoginResponse struct {
	Token             string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User              UserResponse `json:"user"`
	MustResetPassword bool         `json:"must_reset_password" example:"false"`
	Message           string       `json:"message" example:"Login successful"`
}

// ForgotPasswordRequest represents the payload to request a temporary password via SMTP
type ForgotPasswordRequest struct {
	Email string `json:"email" example:"jane@example.com"`
}

// ResetPasswordRequest is used to set a new permanent password after logging in with a temporary password
type ResetPasswordRequest struct {
	NewPassword     string `json:"new_password" example:"NewStrongP@ssw0rd123!"`
	ConfirmPassword string `json:"confirm_password" example:"NewStrongP@ssw0rd123!"`
}

// ChangePasswordRequest represents the payload for logged-in users to update their password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" example:"CurrentStrongP@ss1!"`
	NewPassword     string `json:"new_password" example:"BrandNewStrongP@ss2!"`
	ConfirmPassword string `json:"confirm_password" example:"BrandNewStrongP@ss2!"`
}

// MessageResponse represents a standard generic API response
type MessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}
