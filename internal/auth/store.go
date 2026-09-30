package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ErrUserNotFound is returned when a user cannot be located
var ErrUserNotFound = errors.New("user not found")

// ErrEmailAlreadyExists is returned when attempting to register an existing email
var ErrEmailAlreadyExists = errors.New("email is already registered")

// Store manages persistence of User accounts in a JSON file
type Store struct {
	filePath string
	mu       sync.RWMutex
}

// NewStore initializes a JSON file store at the specified file path
func NewStore(filePath string) (*Store, error) {
	if filePath == "" {
		filePath = "data/users.json"
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for users store: %w", err)
	}

	s := &Store{
		filePath: filePath,
	}

	// If file does not exist, initialize with empty array
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := s.saveUsersUnlocked([]User{}); err != nil {
			return nil, fmt.Errorf("failed to initialize users store: %w", err)
		}
	}

	return s, nil
}

// loadUsersUnlocked reads users from the JSON file without locking (caller must hold lock)
func (s *Store) loadUsersUnlocked() ([]User, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []User{}, nil
		}
		return nil, err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return []User{}, nil
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("failed to parse users JSON: %w", err)
	}

	return users, nil
}

// saveUsersUnlocked writes users atomically to disk without locking (caller must hold write lock)
func (s *Store) saveUsersUnlocked(users []User) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users JSON: %w", err)
	}

	dir := filepath.Dir(s.filePath)
	tmpFile, err := os.CreateTemp(dir, "users-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpName, s.filePath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to replace users file: %w", err)
	}

	return nil
}

// generateID creates a random hex string ID
func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GetByEmail searches for a user by email (case-insensitive)
func (s *Store) GetByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.loadUsersUnlocked()
	if err != nil {
		return nil, err
	}

	normalized := strings.ToLower(strings.TrimSpace(email))
	for _, u := range users {
		if strings.ToLower(u.Email) == normalized {
			userCopy := u
			return &userCopy, nil
		}
	}

	return nil, ErrUserNotFound
}

// GetByID searches for a user by user ID
func (s *Store) GetByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users, err := s.loadUsersUnlocked()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.ID == id {
			userCopy := u
			return &userCopy, nil
		}
	}

	return nil, ErrUserNotFound
}

// Create stores a new user record in the JSON file
func (s *Store) Create(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.loadUsersUnlocked()
	if err != nil {
		return err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))
	for _, u := range users {
		if strings.ToLower(u.Email) == normalizedEmail {
			return ErrEmailAlreadyExists
		}
	}

	if user.ID == "" {
		user.ID = "usr_" + generateID()
	}
	user.Email = normalizedEmail
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	users = append(users, *user)
	return s.saveUsersUnlocked(users)
}

// Update updates an existing user record in the JSON file
func (s *Store) Update(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	users, err := s.loadUsersUnlocked()
	if err != nil {
		return err
	}

	foundIndex := -1
	for i, u := range users {
		if u.ID == user.ID {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return ErrUserNotFound
	}

	user.UpdatedAt = time.Now().UTC()
	users[foundIndex] = *user

	return s.saveUsersUnlocked(users)
}

// GetAll returns all users (for administrative/testing purposes)
func (s *Store) GetAll() ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.loadUsersUnlocked()
}
