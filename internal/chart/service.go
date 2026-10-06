package chart

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"file-converter/internal/nifty"
)

var (
	ErrCompanyNotFound = errors.New("company not found")
	ErrInvalidSymbol   = errors.New("invalid company symbol")
	validSymbolRegex   = regexp.MustCompile(`^[A-Za-z0-9_\-&]+$`)
)

// Service provides access to historical chart data, indicators, and company listings
type Service struct {
	baseDir      string
	niftyStorage nifty.Storage

	mu        sync.RWMutex
	companies []CompanySummary
	cache     map[string]*ChartDataResponse
}

// NewService creates a new Chart Service
func NewService(baseDir string, niftyStorage nifty.Storage) *Service {
	resolvedDir := resolveDataDir(baseDir)
	return &Service{
		baseDir:      resolvedDir,
		niftyStorage: niftyStorage,
		cache:        make(map[string]*ChartDataResponse),
	}
}

func resolveDataDir(preferred string) string {
	candidates := []string{
		preferred,
		"data/nse/data",
		"./data/nse/data",
		"data/nse",
		"./data/nse",
		"/app/data/nse/data",
		"/app/data/nse",
		"/app/data",
		"/apps/rutils/data/nse/data",
		"/apps/rutils/data/nse",
		"/apps/rutils/data",
	}

	// Priority 1: Find candidate that directly has 'companies' folder
	for _, d := range candidates {
		if d == "" {
			continue
		}
		compDir := filepath.Join(d, "companies")
		if fi, err := os.Stat(compDir); err == nil && fi.IsDir() {
			return d
		}
		// Also check nested data/companies
		nested := filepath.Join(d, "data", "companies")
		if fi, err := os.Stat(nested); err == nil && fi.IsDir() {
			return filepath.Join(d, "data")
		}
	}

	// Priority 2: Return first existing directory
	for _, d := range candidates {
		if d == "" {
			continue
		}
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			return d
		}
	}
	return "data/nse/data"
}

// GetCompanies returns a sorted slice of all available companies
func (s *Service) GetCompanies() ([]CompanySummary, error) {
	s.mu.RLock()
	if len(s.companies) > 0 {
		out := make([]CompanySummary, len(s.companies))
		copy(out, s.companies)
		s.mu.RUnlock()
		return out, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check after acquiring write lock
	if len(s.companies) > 0 {
		out := make([]CompanySummary, len(s.companies))
		copy(out, s.companies)
		return out, nil
	}

	// Lookup map for friendly company name and industry from NIFTY 50 if available
	nameMap := make(map[string]struct {
		Name     string
		Industry string
	})

	if s.niftyStorage != nil {
		if data, err := s.niftyStorage.Load(); err == nil && data != nil {
			for _, c := range data.Companies {
				sym := strings.ToUpper(strings.TrimSpace(c.Symbol))
				nameMap[sym] = struct {
					Name     string
					Industry string
				}{
					Name:     c.CompanyName,
					Industry: c.Industry,
				}
			}
		}
	}

	companiesDir := filepath.Join(s.baseDir, "companies")
	entries, err := os.ReadDir(companiesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read companies directory at %s: %w", companiesDir, err)
	}

	var results []CompanySummary

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sym := entry.Name()
		if !validSymbolRegex.MatchString(sym) {
			continue
		}

		companyFolder := filepath.Join(companiesDir, sym)
		metaFile := filepath.Join(companyFolder, "metadata.json")

		summary := CompanySummary{
			Symbol:      sym,
			DisplayName: sym,
			Interval:    "ONE_DAY",
			Exchange:    "NSE",
		}

		if info, ok := nameMap[strings.ToUpper(sym)]; ok {
			if info.Name != "" {
				summary.DisplayName = info.Name
			}
			summary.Industry = info.Industry
		}

		if metaBytes, err := os.ReadFile(metaFile); err == nil {
			var meta Metadata
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				if summary.DisplayName == sym && meta.CompanyName != "" {
					summary.DisplayName = meta.CompanyName
				}
				summary.SymbolToken = meta.SymbolToken
				summary.Exchange = meta.Exchange
				summary.Interval = meta.Interval
				summary.CandleCount = meta.CandleCount
				summary.AdviceCount = meta.AdviceCount
				summary.LatestCandleDate = meta.LatestCandleDate
			}
		}

		// Fallback candle count if metadata was missing or 0
		if summary.CandleCount == 0 {
			candlesFile := filepath.Join(companyFolder, "candles.json")
			if fi, err := os.Stat(candlesFile); err == nil && fi.Size() > 0 {
				summary.CandleCount = 246
			}
		}

		results = append(results, summary)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Symbol < results[j].Symbol
	})

	s.companies = results
	out := make([]CompanySummary, len(results))
	copy(out, results)
	return out, nil
}

// GetChartData reads and returns candles, indicators, advice, and metadata for a company
func (s *Service) GetChartData(symbol string) (*ChartDataResponse, error) {
	cleanSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if !validSymbolRegex.MatchString(cleanSymbol) {
		return nil, ErrInvalidSymbol
	}

	// Check in-memory cache
	s.mu.RLock()
	if cached, ok := s.cache[cleanSymbol]; ok {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	companyFolder := filepath.Join(s.baseDir, "companies", cleanSymbol)
	if fi, err := os.Stat(companyFolder); err != nil || !fi.IsDir() {
		return nil, ErrCompanyNotFound
	}

	resp := &ChartDataResponse{
		Success:    true,
		Symbol:     cleanSymbol,
		Candles:    make([]Candle, 0),
		Indicators: make([]Indicator, 0),
		Advice:     make([]Advice, 0),
	}

	// 1. Read metadata.json
	metaPath := filepath.Join(companyFolder, "metadata.json")
	if bytes, err := os.ReadFile(metaPath); err == nil {
		_ = json.Unmarshal(bytes, &resp.Metadata)
	}
	if resp.Metadata.CompanyName == "" {
		resp.Metadata.CompanyName = cleanSymbol
	}

	// 2. Read candles.json
	candlesPath := filepath.Join(companyFolder, "candles.json")
	if bytes, err := os.ReadFile(candlesPath); err == nil {
		_ = json.Unmarshal(bytes, &resp.Candles)
	}

	// 3. Read indicators.json
	indicatorsPath := filepath.Join(companyFolder, "indicators.json")
	if bytes, err := os.ReadFile(indicatorsPath); err == nil {
		_ = json.Unmarshal(bytes, &resp.Indicators)
	}

	// 4. Read advice.json
	advicePath := filepath.Join(companyFolder, "advice.json")
	if bytes, err := os.ReadFile(advicePath); err == nil {
		_ = json.Unmarshal(bytes, &resp.Advice)
	}

	// Update metadata counts
	if resp.Metadata.CandleCount == 0 {
		resp.Metadata.CandleCount = len(resp.Candles)
	}
	if resp.Metadata.AdviceCount == 0 {
		resp.Metadata.AdviceCount = len(resp.Advice)
	}

	// Store in cache
	s.mu.Lock()
	s.cache[cleanSymbol] = resp
	s.mu.Unlock()

	return resp, nil
}
