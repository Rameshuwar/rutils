package chart

// Metadata represents company-level metadata from metadata.json
type Metadata struct {
	CompanyName      string `json:"companyName"`
	SymbolToken      string `json:"symbolToken"`
	Exchange         string `json:"exchange"`
	Interval         string `json:"interval"`
	LastFetchedAt    string `json:"lastFetchedAt"`
	LastCalculatedAt string `json:"lastCalculatedAt"`
	LatestCandleDate string `json:"latestCandleDate"`
	CandleCount      int    `json:"candleCount"`
	AdviceCount      int    `json:"adviceCount"`
}

// CompanySummary is returned in the company directory list
type CompanySummary struct {
	Symbol           string `json:"symbol"`
	DisplayName      string `json:"displayName"`
	Industry         string `json:"industry,omitempty"`
	SymbolToken      string `json:"symbolToken"`
	Exchange         string `json:"exchange"`
	Interval         string `json:"interval"`
	CandleCount      int    `json:"candleCount"`
	AdviceCount      int    `json:"adviceCount"`
	LatestCandleDate string `json:"latestCandleDate"`
}

// Candle represents a single OHLCV trading candle from candles.json
type Candle struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// Indicator represents pre-calculated technical indicators from indicators.json
type Indicator struct {
	Date   string  `json:"date"`
	Close  float64 `json:"close"`
	EMA5   float64 `json:"ema5"`
	EMA13  float64 `json:"ema13"`
	EMA26  float64 `json:"ema26"`
	SMA200 float64 `json:"sma200"`
	RSI    float64 `json:"rsi"`
}

// Advice represents trading signals from advice.json
type Advice struct {
	NiftyIdentifier string  `json:"niftyIdentifier"`
	Index           int     `json:"index"`
	Advice          string  `json:"advice"`
	TargetPrice     float64 `json:"targetPrice"`
	SMA200Support   float64 `json:"sma200Support"`
	Date            string  `json:"date"`
	Open            float64 `json:"open"`
	High            float64 `json:"high"`
	Low             float64 `json:"low"`
	Close           float64 `json:"close"`
	Volume          float64 `json:"volume"`
}

// CompaniesResponse is the JSON payload for GET /api/market/chart/companies
type CompaniesResponse struct {
	Success   bool             `json:"success"`
	Count     int              `json:"count"`
	Companies []CompanySummary `json:"companies"`
}

// ChartDataResponse is the complete payload for GET /api/market/chart/data
type ChartDataResponse struct {
	Success    bool        `json:"success"`
	Symbol     string      `json:"symbol"`
	Metadata   Metadata    `json:"metadata"`
	Candles    []Candle    `json:"candles"`
	Indicators []Indicator `json:"indicators"`
	Advice     []Advice    `json:"advice"`
}
