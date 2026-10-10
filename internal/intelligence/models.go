package intelligence

type Profile struct {
	Description      string `json:"description"`
	BusinessSegments string `json:"business_segments"`
}

type ManagementChange struct {
	Date            string `json:"date"`
	Headline        string `json:"headline"`
	Summary         string `json:"summary"`
	Source          string `json:"source"`
	OriginalURL     string `json:"original_url"`
}

type Announcement struct {
	Date            string `json:"date"`
	Headline        string `json:"headline"`
	Summary         string `json:"summary"`
	Source          string `json:"source"`
	OriginalURL     string `json:"original_url"`
}

type News struct {
	Date            string `json:"date"`
	Headline        string `json:"headline"`
	Summary         string `json:"summary"`
	Source          string `json:"source"`
	OriginalURL     string `json:"original_url"`
	Classification  string `json:"classification"`
}

type FinancialResult struct {
	Period          string            `json:"period"` // e.g., "Q1 2024", "FY24"
	StatementType   string            `json:"statement_type"` // e.g., "Consolidated", "Standalone"
	Revenue         string            `json:"revenue"`
	OperatingProfit string            `json:"operating_profit"`
	NetProfit       string            `json:"net_profit"`
	EPS             string            `json:"eps"`
	OriginalURL     string            `json:"original_url"`
}

type FinancialMetrics struct {
	TotalAssets       string `json:"total_assets"`
	TotalLiabilities  string `json:"total_liabilities"`
	Borrowings        string `json:"borrowings"`
	CashFlowOperating string `json:"cash_flow_operating"`
}

type ResultsCalendar struct {
	PreviousPublicationDate string `json:"previous_publication_date"`
	NextMeetingDate         string `json:"next_meeting_date"`
	Purpose                 string `json:"purpose"`
	SourceURL               string `json:"source_url"`
}

type Report struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	OriginalURL string `json:"original_url"`
}

type SourceMetadata struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	LastChecked string `json:"last_checked"`
	Status      string `json:"status"` // "success", "failed", "unavailable"
}

// CompanyIntelligence represents the on-disk JSON structure for a single company's intelligence
type CompanyIntelligence struct {
	Symbol              string               `json:"symbol"`
	CompanyName         string               `json:"company_name"`
	Industry            string               `json:"industry"`
	Profile             *Profile             `json:"profile"`
	ManagementChanges   []ManagementChange   `json:"management_changes"`
	Announcements       []Announcement       `json:"announcements"`
	News                []News               `json:"news"`
	QuarterlyResults    []FinancialResult    `json:"quarterly_results"`
	AnnualResults       []FinancialResult    `json:"annual_results"`
	FinancialMetrics    *FinancialMetrics    `json:"financial_metrics"`
	ResultsCalendar     *ResultsCalendar     `json:"results_calendar"`
	Reports             []Report             `json:"reports"`
	Sources             []SourceMetadata     `json:"sources"`
	UpdatedAt           string               `json:"updated_at"`
	LastSuccessfulFetch string               `json:"last_successful_fetch"`
	FetchStatus         string               `json:"fetch_status"`
}
