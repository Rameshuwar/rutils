package nifty

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper: generates a valid 50-company CSV
func generateSampleCSV(count int) string {
	var sb strings.Builder
	sb.WriteString("Company Name,Industry,Symbol,Series,ISIN Code\n")
	for i := 1; i <= count; i++ {
		isin := fmt.Sprintf("INE%09d", i)
		sb.WriteString(fmt.Sprintf("Company %d,Sector %d,SYM%d,EQ,%s\n", i, (i%5)+1, i, isin))
	}
	return sb.String()
}

func TestParseCSV_Valid(t *testing.T) {
	csvData := generateSampleCSV(50)
	companies, err := ParseCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error parsing valid CSV: %v", err)
	}

	if len(companies) != 50 {
		t.Fatalf("expected 50 companies, got %d", len(companies))
	}

	first := companies[0]
	if first.CompanyName != "Company 1" || first.Industry != "Sector 2" || first.Symbol != "SYM1" || first.Series != "EQ" || first.ISIN != "INE000000001" {
		t.Errorf("first company fields mismatch: %+v", first)
	}
}

func TestParseCSV_CaseInsensitiveHeaders(t *testing.T) {
	csvData := "company name,INDUSTRY,Symbol,SERIES,isin code\nReliance Industries Ltd.,Oil Gas & Consumable Fuels,RELIANCE,EQ,INE002A01018\n"
	companies, err := ParseCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("unexpected error parsing CSV with mixed case headers: %v", err)
	}

	if len(companies) != 1 {
		t.Fatalf("expected 1 company, got %d", len(companies))
	}
	if companies[0].Symbol != "RELIANCE" || companies[0].ISIN != "INE002A01018" {
		t.Errorf("parsed company mismatch: %+v", companies[0])
	}
}

func TestParseCSV_MissingHeaders(t *testing.T) {
	// Missing ISIN Code
	csvData := "Company Name,Industry,Symbol,Series\nReliance Industries Ltd.,Oil Gas,RELIANCE,EQ\n"
	_, err := ParseCSV(strings.NewReader(csvData))
	if err == nil {
		t.Fatalf("expected error for missing headers, got nil")
	}
}

func TestParseCSV_EmptyOrInsufficientRows(t *testing.T) {
	_, err := ParseCSV(strings.NewReader(""))
	if err == nil {
		t.Fatalf("expected error for empty CSV, got nil")
	}

	_, err = ParseCSV(strings.NewReader("Company Name,Industry,Symbol,Series,ISIN Code\n"))
	if err != nil {
		// Parsing header only returns empty slice without error, but validation will catch it
		companies, _ := ParseCSV(strings.NewReader("Company Name,Industry,Symbol,Series,ISIN Code\n"))
		if len(companies) != 0 {
			t.Fatalf("expected 0 companies from header-only CSV, got %d", len(companies))
		}
	}
}

func TestParseCSV_MissingFieldsInRow(t *testing.T) {
	csvData := "Company Name,Industry,Symbol,Series,ISIN Code\n,Healthcare,APOLLO,EQ,INE437A01024\n"
	_, err := ParseCSV(strings.NewReader(csvData))
	if err == nil {
		t.Fatalf("expected error for missing Company Name in row, got nil")
	}
}

func TestParseCSV_DuplicateSymbols(t *testing.T) {
	csvData := "Company Name,Industry,Symbol,Series,ISIN Code\nCompany 1,Tech,TCS,EQ,INE467B01029\nCompany 2,Tech,TCS,EQ,INE467B01030\n"
	_, err := ParseCSV(strings.NewReader(csvData))
	if err == nil {
		t.Fatalf("expected error for duplicate symbol TCS, got nil")
	}
}

func TestValidateCompanies_Exactly50(t *testing.T) {
	// 49 companies should fail
	csv49 := generateSampleCSV(49)
	companies49, err := ParseCSV(strings.NewReader(csv49))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if err := ValidateCompanies(companies49); err == nil {
		t.Fatalf("expected validation error for 49 companies, got nil")
	}

	// 51 companies should fail
	csv51 := generateSampleCSV(51)
	companies51, err := ParseCSV(strings.NewReader(csv51))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if err := ValidateCompanies(companies51); err == nil {
		t.Fatalf("expected validation error for 51 companies, got nil")
	}

	// 50 companies should succeed
	csv50 := generateSampleCSV(50)
	companies50, err := ParseCSV(strings.NewReader(csv50))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if err := ValidateCompanies(companies50); err != nil {
		t.Fatalf("expected validation to pass for 50 companies, got: %v", err)
	}
}

func TestValidateCompanies_InvalidISIN(t *testing.T) {
	csvData := generateSampleCSV(50)
	companies, _ := ParseCSV(strings.NewReader(csvData))
	companies[5].ISIN = "INVALID_ISIN"

	if err := ValidateCompanies(companies); err == nil {
		t.Fatalf("expected validation error for invalid ISIN, got nil")
	}
}

func TestFileStorage_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nifty50.json")

	storage := NewFileStorage(filePath)

	// Load non-existent file
	_, err := storage.Load()
	if err == nil {
		t.Fatalf("expected ErrNotFound for non-existent file, got nil")
	}

	// Save valid data
	csvData := generateSampleCSV(50)
	companies, _ := ParseCSV(strings.NewReader(csvData))
	data := &NiftyData{
		UpdatedAt: "2026-09-30T10:00:00+05:30",
		Count:     50,
		Companies: companies,
	}

	if err := storage.Save(data); err != nil {
		t.Fatalf("failed to save data: %v", err)
	}

	// Load and verify
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("failed to load data: %v", err)
	}
	if loaded.Count != 50 || loaded.UpdatedAt != "2026-09-30T10:00:00+05:30" {
		t.Errorf("loaded data mismatch: count=%d, updated_at=%s", loaded.Count, loaded.UpdatedAt)
	}
	if len(loaded.Companies) != 50 {
		t.Fatalf("loaded companies count mismatch: %d", len(loaded.Companies))
	}
}

func TestFileStorage_AtomicReplacementSafety(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nifty50.json")
	storage := NewFileStorage(filePath)

	// Initial save
	csvData1 := generateSampleCSV(50)
	companies1, _ := ParseCSV(strings.NewReader(csvData1))
	data1 := &NiftyData{
		UpdatedAt: "2026-09-29T10:00:00+05:30",
		Count:     50,
		Companies: companies1,
	}
	if err := storage.Save(data1); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Attempt to save invalid data (e.g. 40 companies)
	invalidCompanies := companies1[:40]
	invalidData := &NiftyData{
		UpdatedAt: "2026-09-30T10:00:00+05:30",
		Count:     40,
		Companies: invalidCompanies,
	}
	err := storage.Save(invalidData)
	if err == nil {
		t.Fatalf("expected Save to fail with invalid data, got nil")
	}

	// Verify original file is still intact and unaffected
	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("failed to load data after failed save: %v", err)
	}
	if loaded.UpdatedAt != "2026-09-29T10:00:00+05:30" || loaded.Count != 50 {
		t.Fatalf("original data was mutated or corrupted: %+v", loaded)
	}
}

func TestSyncer_SuccessAndFailureKeepsPrevious(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nifty50.json")
	storage := NewFileStorage(filePath)

	validCSV := generateSampleCSV(50)

	// Mock server that returns valid CSV on first request, 500 on second request
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(validCSV))
			return
		}
		http.Error(w, "NSE Service Unavailable", http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	client := NewNSEClient(ts.URL)
	client.SetHTTPClient(ts.Client())

	syncer := NewSyncer(client, storage, nil)

	// First sync should succeed
	ctx := context.Background()
	data1, err := syncer.Sync(ctx)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}
	if data1.Count != 50 {
		t.Fatalf("expected count 50, got %d", data1.Count)
	}

	// Verify file exists
	loaded1, err := storage.Load()
	if err != nil {
		t.Fatalf("failed to load storage: %v", err)
	}
	if loaded1.Count != 50 {
		t.Fatalf("expected loaded count 50, got %d", loaded1.Count)
	}

	// Second sync should fail due to 503 from server
	_, err = syncer.Sync(ctx)
	if err == nil {
		t.Fatalf("expected second sync to fail, got nil")
	}

	// Verify previous data is preserved intact
	loaded2, err := storage.Load()
	if err != nil {
		t.Fatalf("failed to load storage after failed sync: %v", err)
	}
	if loaded2.UpdatedAt != loaded1.UpdatedAt || loaded2.Count != 50 {
		t.Fatalf("previous data was not preserved: %+v", loaded2)
	}
}

func TestSyncer_InvalidDatasetRejection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nifty50.json")
	storage := NewFileStorage(filePath)

	// Server returning invalid CSV (only 10 rows)
	invalidCSV := generateSampleCSV(10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(invalidCSV))
	}))
	defer ts.Close()

	client := NewNSEClient(ts.URL)
	client.SetHTTPClient(ts.Client())

	syncer := NewSyncer(client, storage, nil)
	ctx := context.Background()
	_, err := syncer.Sync(ctx)
	if err == nil {
		t.Fatalf("expected sync to reject invalid dataset, got nil")
	}

	// Verify no file was created
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("expected file to not exist after failed sync, got err=%v", err)
	}
}
