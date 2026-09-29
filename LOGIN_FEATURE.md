# Authentication & Login Feature Documentation (`LOGIN_FEATURE.md`)

This document provides a comprehensive developer guide for the Authentication System implemented in the Rutils utility platform.

---

## 1. Architectural Overview

The authentication system is completely isolated and decoupled from existing utility services (file conversion, text extraction, calculations). It provides:
- **User Persistence**: File-based JSON storage in `data/users.json` with thread safety (`sync.RWMutex`) and atomic file replacement.
- **Credential Security**: Passwords hashed using `bcrypt` at cost factor `12`.
- **Session Management**: Stateless JSON Web Tokens (JWT) signed with HMAC-SHA256 (`HS256`).
- **Google SMTP Integration**: Automated dispatching of temporary passwords and security notices with zero hardcoded credentials and an automatic console-logging development fallback when credentials are not configured.
- **Swagger Documentation**: OpenAPI 2.0 annotations integrated into Swagger UI with Bearer authentication support.

---

## 2. Configuration (`config.json`)

Configuration is stored in `config.json` at the project root. Never commit real credentials to version control.

### Default Structure:
```json
{
  "smtp_email": "",
  "smtp_app_password": "",
  "smtp_host": "smtp.gmail.com",
  "smtp_port": "587",
  "jwt_secret": "rutils-utility-jwt-secret-change-in-production-2026",
  "jwt_expiration_hours": 24,
  "data_dir": "data"
}
```

### Configuration Parameters:
| Key | Type | Description | Default Fallback |
| :--- | :--- | :--- | :--- |
| `smtp_email` | string | Google Account email address | `""` (Dev Fallback active) |
| `smtp_app_password` | string | 16-character Google App Password | `""` (Dev Fallback active) |
| `smtp_host` | string | SMTP server hostname | `"smtp.gmail.com"` |
| `smtp_port` | string | SMTP server port with STARTTLS | `"587"` |
| `jwt_secret` | string | Secret key used to sign and verify JWT tokens | `"rutils-utility-jwt-secret-change-in-production-2026"` |
| `jwt_expiration_hours`| int | Lifetime of generated JWT tokens | `24` |
| `data_dir` | string | Directory containing JSON storage files | `"data"` |

### How to Configure Google SMTP:
1. Log in to your Google Account and navigate to **Security** ([myaccount.google.com/security](https://myaccount.google.com/security)).
2. Ensure **2-Step Verification** is turned **ON**.
3. Under "How you sign in to Google", select **App passwords** (or search "App passwords" in the search bar).
4. Create a new App password named `Rutils`.
5. Google will generate a 16-character password (e.g. `abcd efgh ijkl mnop`).
6. Copy the password and set it in `config.json`:
   ```json
   "smtp_email": "your-email@gmail.com",
   "smtp_app_password": "abcdefghijklmnop"
   ```

### Development / Local Mock Mode:
When `smtp_email` or `smtp_app_password` are left empty, the server automatically enters **Development Fallback Mode**:
- Emails are not dispatched over the network.
- The system logs the generated temporary password directly to the console:
  ```text
  [AUTH-DEV] SMTP not configured in config.json. Temporary password for <user@example.com> is: Xy9#mK!28zQ1
  ```
- All API requests succeed smoothly, enabling offline development and testing.

---

## 3. Storage (`data/users.json`)

User accounts are stored in `data/users.json`. All read and write operations are synchronized with a read-write mutex (`sync.RWMutex`). Writes are performed **atomically** by writing to a temporary file (`users-*.tmp`) and renaming it to `data/users.json`, preventing data corruption or race conditions.

### User Schema:
```json
[
  {
    "id": "usr_94e5088c4b18cb6a",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password_hash": "$2a$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
    "temp_password_hash": "",
    "temp_password_expiry": null,
    "must_reset_password": false,
    "created_at": "2026-09-29T12:00:00Z",
    "updated_at": "2026-09-29T12:00:00Z"
  }
]
```

---

## 4. Password Security & Validation Rules

All user passwords must satisfy the following strict complexity requirements:
- **Minimum Length**: 8 characters
- **Uppercase Letter**: At least 1 (`A-Z`)
- **Lowercase Letter**: At least 1 (`a-z`)
- **Number**: At least 1 (`0-9`)
- **Special Character**: At least 1 (`!@#$%^&*()-_=+[]{}|;:'",.<>/?~`\``)

### Password Hashing:
- Passwords are encrypted using **bcrypt** with cost factor `12`.
- Plaintext passwords are never stored or returned in any API responses.

---

## 5. JWT Authentication Flow

1. Upon successful login (`POST /auth/login`), the server issues a signed JWT token:
   ```json
   {
     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "user": {
       "id": "usr_94e5088c4b18cb6a",
       "name": "Jane Doe",
       "email": "jane@example.com",
       "must_reset_password": false,
       "created_at": "2026-09-29T12:00:00Z"
     },
     "must_reset_password": false,
     "message": "Login successful"
   }
   ```
2. For protected endpoints (`/auth/reset-password`, `/auth/change-password`, `/auth/me`), include the token in the HTTP `Authorization` header:
   ```http
   Authorization: Bearer <YOUR_JWT_TOKEN>
   ```
3. **Forced Reset Workflow**: If a user logs in using a temporary password issued via `/auth/forgot-password`:
   - `must_reset_password` is set to `true` in the JWT claims and login response.
   - The user must call `POST /auth/reset-password` with their new password before continuing normal operations.

---

## 6. Endpoints Reference

### 1. Register User
- **Route**: `POST /auth/register`
- **Access**: Public
- **Request Body**:
  ```json
  {
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "StrongPassword!2026",
    "confirm_password": "StrongPassword!2026"
  }
  ```
- **Responses**:
  - `201 Created`:
    ```json
    {
      "id": "usr_94e5088c4b18cb6a",
      "name": "Jane Doe",
      "email": "jane@example.com",
      "must_reset_password": false,
      "created_at": "2026-09-29T12:00:00Z"
    }
    ```
  - `400 Bad Request`: Validation failure (e.g. password mismatch or weak password).
  - `409 Conflict`: Email is already registered.

### 2. Login User
- **Route**: `POST /auth/login`
- **Access**: Public
- **Request Body**:
  ```json
  {
    "email": "jane@example.com",
    "password": "StrongPassword!2026"
  }
  ```
- **Responses**:
  - `200 OK`: Returns JWT token and user profile.
  - `401 Unauthorized`: Invalid email or password.

### 3. Logout User
- **Route**: `POST /auth/logout`
- **Access**: Public
- **Responses**:
  - `200 OK`:
    ```json
    {
      "success": true,
      "message": "Logged out successfully"
    }
    ```

### 4. Forgot Password
- **Route**: `POST /auth/forgot-password`
- **Access**: Public
- **Request Body**:
  ```json
  {
    "email": "jane@example.com"
  }
  ```
- **Responses**:
  - `200 OK`:
    ```json
    {
      "success": true,
      "message": "A temporary password has been dispatched to your email address."
    }
    ```
  - `404 Not Found`: No account registered with that email.

### 5. Reset Password (Forced)
- **Route**: `POST /auth/reset-password`
- **Access**: Authenticated (`Authorization: Bearer <token>`)
- **Request Body**:
  ```json
  {
    "new_password": "BrandNewPassword#999",
    "confirm_password": "BrandNewPassword#999"
  }
  ```
- **Responses**:
  - `200 OK`:
    ```json
    {
      "success": true,
      "message": "Permanent password updated successfully. You can now use your new password."
    }
    ```
  - `400 Bad Request`: Password mismatch or does not meet complexity rules.
  - `401 Unauthorized`: Missing or invalid Bearer token.

### 6. Change Password (Logged-in User)
- **Route**: `POST /auth/change-password`
- **Access**: Authenticated (`Authorization: Bearer <token>`)
- **Request Body**:
  ```json
  {
    "current_password": "CurrentPassword!2026",
    "new_password": "BrandNewPassword#999",
    "confirm_password": "BrandNewPassword#999"
  }
  ```
- **Responses**:
  - `200 OK`:
    ```json
    {
      "success": true,
      "message": "Password changed successfully. A confirmation email has been dispatched."
    }
    ```
  - `401 Unauthorized`: Current password incorrect or missing token.
  - `400 Bad Request`: Password mismatch or weak new password.

### 7. Get Current User Profile
- **Route**: `GET /auth/me`
- **Access**: Authenticated (`Authorization: Bearer <token>`)
- **Responses**:
  - `200 OK`: Returns user profile without password hashes.
  - `401 Unauthorized`: Invalid or expired token.

---

## 7. Swagger Documentation

All authentication endpoints are fully documented and integrated with Swagger UI:
- **Swagger JSON**: `http://localhost:8080/swagger/doc.json`
- **Swagger UI**: `http://localhost:8080/swagger/`

To authorize in Swagger UI:
1. Click the **Authorize** button (green lock icon).
2. Enter your Bearer token in the format: `Bearer <YOUR_TOKEN>`.
3. Click **Authorize** to test protected endpoints (`/auth/me`, `/auth/reset-password`, `/auth/change-password`).

To regenerate documentation after making changes:
```bash
~/go/bin/swag init -g cmd/server/main.go -o docs
```

---

## 8. Automated Testing

Run all unit and integration tests across the project:
```bash
# Run internal auth unit tests (password validation, hashing, JWT, user store)
go test -v ./internal/auth

# Run end-to-end API integration tests
go test -v ./cmd/server -run TestAuth

# Run entire test suite
go test -v ./...
```
