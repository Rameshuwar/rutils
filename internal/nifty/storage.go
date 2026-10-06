package nifty

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Storage defines the interface for reading and writing NIFTY 50 constituent data
type Storage interface {
	Save(data *NiftyData) error
	Load() (*NiftyData, error)
	FilePath() string
}

// FileStorage implements Storage using the local filesystem with atomic replacement
type FileStorage struct {
	filePath string
	mu       sync.RWMutex
}

// NewFileStorage creates a new FileStorage with the given path
func NewFileStorage(filePath string) *FileStorage {
	if filePath == "" {
		filePath = filepath.Join("data", "nifty50.json")
	}
	return &FileStorage{
		filePath: filePath,
	}
}

// FilePath returns the configured file path
func (s *FileStorage) FilePath() string {
	return s.filePath
}

// Save atomically writes the NiftyData to disk
func (s *FileStorage) Save(data *NiftyData) error {
	if data == nil {
		return fmt.Errorf("cannot save nil data")
	}

	if err := ValidateCompanies(data.Companies); err != nil {
		return fmt.Errorf("cannot save invalid data: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to marshal nifty data: %w", err)
	}

	// Create a unique temporary file in the same directory for atomic rename
	tmpFile := filepath.Join(dir, fmt.Sprintf(".nifty50_%d_%d.tmp", time.Now().UnixNano(), os.Getpid()))

	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tmpFile, err)
	}

	// Verify temporary file can be read and decoded before replacing
	testBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to verify temporary file: %w", err)
	}

	var testData NiftyData
	if err := json.Unmarshal(testBytes, &testData); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to parse temporary file verification: %w", err)
	}

	if len(testData.Companies) != len(data.Companies) {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("verification company count mismatch: got %d, expected %d", len(testData.Companies), len(data.Companies))
	}

	// Atomically replace target file
	if err := os.Rename(tmpFile, s.filePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename failed from %s to %s: %w", tmpFile, s.filePath, err)
	}

	return nil
}

// Load reads and parses the NiftyData from disk
func (s *FileStorage) Load() (*NiftyData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bytes, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to read %s: %w", s.filePath, err)
	}

	var data NiftyData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", s.filePath, err)
	}

	if err := ValidateCompanies(data.Companies); err != nil {
		return nil, fmt.Errorf("data in %s failed validation: %w", s.filePath, err)
	}

	data.Count = len(data.Companies)
	return &data, nil
}
