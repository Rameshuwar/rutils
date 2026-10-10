package intelligence

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

type Storage interface {
	Save(symbol string, data *CompanyIntelligence) error
	Load(symbol string) (*CompanyIntelligence, error)
	Delete(symbol string) error
}

type FileStorage struct {
	baseDir string
	mu      sync.RWMutex
}

func NewFileStorage(baseDir string) *FileStorage {
	if baseDir == "" {
		baseDir = filepath.Join("data", "company-intelligence")
	}
	return &FileStorage{
		baseDir: baseDir,
	}
}

func (s *FileStorage) filePath(symbol string) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%s.json", symbol))
}

func (s *FileStorage) Save(symbol string, data *CompanyIntelligence) error {
	if data == nil {
		return fmt.Errorf("cannot save nil data")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", s.baseDir, err)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to marshal company intelligence: %w", err)
	}

	fp := s.filePath(symbol)
	tmpFile := filepath.Join(s.baseDir, fmt.Sprintf(".%s_%d_%d.tmp", symbol, time.Now().UnixNano(), os.Getpid()))

	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tmpFile, err)
	}

	testBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to verify temporary file: %w", err)
	}
	var testData CompanyIntelligence
	if err := json.Unmarshal(testBytes, &testData); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to parse temporary file verification: %w", err)
	}

	if err := os.Rename(tmpFile, fp); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename failed from %s to %s: %w", tmpFile, fp, err)
	}

	return nil
}

func (s *FileStorage) Load(symbol string) (*CompanyIntelligence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fp := s.filePath(symbol)
	bytes, err := os.ReadFile(fp)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("failed to read %s: %w", fp, err)
	}

	var data CompanyIntelligence
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", fp, err)
	}

	return &data, nil
}

func (s *FileStorage) Delete(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	fp := s.filePath(symbol)
	if err := os.Remove(fp); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to delete file %s: %w", fp, err)
	}
	return nil
}
