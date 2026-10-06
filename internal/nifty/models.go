package nifty

import "errors"

var (
	// ErrNotFound indicates that the NIFTY 50 dataset file does not exist
	ErrNotFound = errors.New("nifty50 data file not found")

	// ErrInvalidDataset indicates that the dataset failed validation checks
	ErrInvalidDataset = errors.New("invalid or incomplete NSE dataset")

	// ErrDownloadFailed indicates that the CSV could not be fetched from NSE
	ErrDownloadFailed = errors.New("unable to download NSE data")
)

// Company represents a single constituent company in the NIFTY 50 index
type Company struct {
	CompanyName string `json:"company_name"`
	Industry    string `json:"industry"`
	Symbol      string `json:"symbol"`
	Series      string `json:"series"`
	ISIN        string `json:"isin"`
}

// NiftyData represents the on-disk JSON structure for NIFTY 50 data
type NiftyData struct {
	UpdatedAt string    `json:"updated_at"`
	Count     int       `json:"count"`
	Companies []Company `json:"companies"`
}

// CompaniesResponse represents the HTTP API response for GET /nifty50/companies
type CompaniesResponse struct {
	Success   bool      `json:"success"`
	UpdatedAt string    `json:"updated_at"`
	Count     int       `json:"count"`
	Companies []Company `json:"companies"`
	Error     string    `json:"error,omitempty"`
}
