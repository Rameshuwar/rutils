package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is set to 12 for strong security
const BcryptCost = 12

// SpecialCharacters is the allowed set of special characters for passwords
const SpecialCharacters = "!@#$%^&*()-_=+[]{}|;:'\",.<>/?~`"

// ValidatePassword checks all password complexity rules:
// - Minimum 8 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 digit
// - At least 1 special character
func ValidatePassword(pw string) error {
	if len(pw) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(SpecialCharacters, r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	var missing []string
	if !hasUpper {
		missing = append(missing, "at least one uppercase letter (A-Z)")
	}
	if !hasLower {
		missing = append(missing, "at least one lowercase letter (a-z)")
	}
	if !hasDigit {
		missing = append(missing, "at least one number (0-9)")
	}
	if !hasSpecial {
		missing = append(missing, "at least one special character (!@#$%^&* etc.)")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain %s", strings.Join(missing, ", "))
	}

	return nil
}

// HashPassword hashes a plain text password using bcrypt with cost 12
func HashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword compares a hashed password with a plain text candidate
func ComparePassword(hash, pw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
	return err == nil
}

// GenerateTemporaryPassword generates a cryptographically secure random password
// that satisfies all complexity requirements (length 14, uppercase, lowercase, digit, special).
func GenerateTemporaryPassword() string {
	const (
		upperChars   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		lowerChars   = "abcdefghijkmnpqrstuvwxyz"
		digitChars   = "23456789"
		specialChars = "!@#$%^&*_+="
		allChars     = upperChars + lowerChars + digitChars + specialChars
	)

	// Guarantee at least one character from each set
	pick := func(set string) byte {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			return set[0]
		}
		return set[n.Int64()]
	}

	passwordBytes := []byte{
		pick(upperChars),
		pick(lowerChars),
		pick(digitChars),
		pick(specialChars),
	}

	// Fill the remaining characters up to length 14
	for i := 4; i < 14; i++ {
		passwordBytes = append(passwordBytes, pick(allChars))
	}

	// Shuffle the bytes cryptographically
	for i := len(passwordBytes) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		j := int(jBig.Int64())
		passwordBytes[i], passwordBytes[j] = passwordBytes[j], passwordBytes[i]
	}

	return string(passwordBytes)
}
