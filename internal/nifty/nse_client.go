package nifty

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultNSEURL is the primary official downloadable CSV URL for NIFTY 50
	DefaultNSEURL = "https://nsearchives.nseindia.com/content/indices/ind_nifty50list.csv"

	// FallbackNSEURL is the secondary official downloadable CSV URL
	FallbackNSEURL = "https://archives.nseindia.com/content/indices/ind_nifty50list.csv"

	// ExpectedNifty50Count is the expected number of constituent companies in NIFTY 50
	ExpectedNifty50Count = 50
)

var isinRegex = regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)

// NSEClient interface for fetching and parsing NIFTY 50 data
type NSEClient interface {
	FetchAndParse(ctx context.Context) ([]Company, error)
	FetchCSV(ctx context.Context, url string) ([]byte, error)
}

// HTTPNSEClient is the default HTTP client implementation for NSE
type HTTPNSEClient struct {
	httpClient *http.Client
	urls       []string
	userAgent  string
}

// NewNSEClient creates a new HTTPNSEClient with production-ready defaults
func NewNSEClient(urls ...string) *HTTPNSEClient {
	if len(urls) == 0 {
		urls = []string{DefaultNSEURL, FallbackNSEURL}
	}
	return &HTTPNSEClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		urls:      urls,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
}

// SetHTTPClient allows overriding the internal http.Client (useful for testing)
func (c *HTTPNSEClient) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// FetchCSV downloads raw CSV bytes from a specific URL with browser-like headers
func (c *HTTPNSEClient) FetchCSV(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/csv,text/plain,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	if len(bodyBytes) == 0 {
		return nil, fmt.Errorf("empty response received from %s", url)
	}

	return bodyBytes, nil
}

// FetchAndParse attempts to fetch the CSV from configured URLs and parse it
func (c *HTTPNSEClient) FetchAndParse(ctx context.Context) ([]Company, error) {
	var lastErr error

	for _, u := range c.urls {
		body, err := c.FetchCSV(ctx, u)
		if err != nil {
			lastErr = err
			continue
		}

		companies, err := ParseCSV(strings.NewReader(string(body)))
		if err != nil {
			lastErr = fmt.Errorf("parse error from %s: %w", u, err)
			continue
		}

		if err := ValidateCompanies(companies); err != nil {
			lastErr = fmt.Errorf("validation error from %s: %w", u, err)
			continue
		}

		return companies, nil
	}

	return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, lastErr)
}

// ParseCSV parses an io.Reader containing NSE NIFTY 50 CSV data
func ParseCSV(r io.Reader) ([]Company, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable fields in case of trailing blanks

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv parsing error: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("%w: csv has insufficient rows (found %d)", ErrInvalidDataset, len(records))
	}

	headerRow := records[0]
	colMap := make(map[string]int)
	for idx, col := range headerRow {
		cleaned := strings.ToLower(strings.TrimSpace(col))
		// Remove BOM or special characters if present
		cleaned = strings.TrimPrefix(cleaned, "\ufeff")
		colMap[cleaned] = idx
	}

	// Required column keys
	nameIdx, hasName := getHeaderIndex(colMap, "company name", "company", "name")
	indIdx, hasInd := getHeaderIndex(colMap, "industry", "sector")
	symIdx, hasSym := getHeaderIndex(colMap, "symbol", "stock symbol")
	serIdx, hasSer := getHeaderIndex(colMap, "series")
	isinIdx, hasISIN := getHeaderIndex(colMap, "isin code", "isin")

	if !hasName || !hasInd || !hasSym || !hasSer || !hasISIN {
		return nil, fmt.Errorf("%w: missing required columns in CSV header (header: %v)", ErrInvalidDataset, headerRow)
	}

	var companies []Company
	seenSymbols := make(map[string]bool)

	for rowIdx, row := range records[1:] {
		if len(row) == 0 {
			continue
		}

		// Check if empty line
		allEmpty := true
		for _, field := range row {
			if strings.TrimSpace(field) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		companyName := getField(row, nameIdx)
		industry := getField(row, indIdx)
		symbol := getField(row, symIdx)
		series := getField(row, serIdx)
		isin := getField(row, isinIdx)

		if companyName == "" || symbol == "" || isin == "" {
			return nil, fmt.Errorf("%w: row %d has missing mandatory fields (symbol=%q, name=%q, isin=%q)",
				ErrInvalidDataset, rowIdx+2, symbol, companyName, isin)
		}

		if seenSymbols[symbol] {
			return nil, fmt.Errorf("%w: duplicate symbol detected: %q at row %d", ErrInvalidDataset, symbol, rowIdx+2)
		}
		seenSymbols[symbol] = true

		companies = append(companies, Company{
			CompanyName: companyName,
			Industry:    industry,
			Symbol:      symbol,
			Series:      series,
			ISIN:        isin,
		})
	}

	return companies, nil
}

// ValidateCompanies verifies that the parsed companies slice conforms to NIFTY 50 specifications
func ValidateCompanies(companies []Company) error {
	if len(companies) != ExpectedNifty50Count {
		return fmt.Errorf("%w: expected exactly %d companies, got %d", ErrInvalidDataset, ExpectedNifty50Count, len(companies))
	}

	for i, c := range companies {
		if strings.TrimSpace(c.CompanyName) == "" {
			return fmt.Errorf("%w: company %d has empty name", ErrInvalidDataset, i+1)
		}
		if strings.TrimSpace(c.Symbol) == "" {
			return fmt.Errorf("%w: company %d has empty symbol", ErrInvalidDataset, i+1)
		}
		if strings.TrimSpace(c.Industry) == "" {
			return fmt.Errorf("%w: company %q has empty industry", ErrInvalidDataset, c.Symbol)
		}
		if strings.TrimSpace(c.Series) == "" {
			return fmt.Errorf("%w: company %q has empty series", ErrInvalidDataset, c.Symbol)
		}
		isin := strings.TrimSpace(c.ISIN)
		if isin == "" {
			return fmt.Errorf("%w: company %q has empty ISIN", ErrInvalidDataset, c.Symbol)
		}
		if !isinRegex.MatchString(isin) {
			return fmt.Errorf("%w: company %q has invalid ISIN format: %q", ErrInvalidDataset, c.Symbol, isin)
		}
	}

	return nil
}

func getHeaderIndex(colMap map[string]int, possibleNames ...string) (int, bool) {
	for _, name := range possibleNames {
		if idx, ok := colMap[name]; ok {
			return idx, true
		}
	}
	return -1, false
}

func getField(row []string, idx int) string {
	if idx >= 0 && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}
