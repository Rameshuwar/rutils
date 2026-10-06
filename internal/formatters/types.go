package formatters

import (
	"context"
	"io"
)

// Category classifies a formatter's output so clients can group
// dropdown entries without hardcoding format names.
type Category string

const (
	CategoryImage       Category = "image"
	CategoryDocument    Category = "document"
	CategoryData        Category = "data"
	CategorySpreadsheet Category = "spreadsheet"
	CategoryMarkup      Category = "markup"
)

// Formatter describes one directed conversion: From → To.
//
// Convert receives the raw uploaded bytes as an io.Reader and must write
// the fully converted output to `out`. Returning a non-nil error causes
// the API layer to respond 400 with the error message.
type Formatter struct {
	From       string   // e.g. "jpg"
	To         string   // e.g. "png"
	OutputMIME string   // e.g. "image/png"
	OutputExt  string   // e.g. ".png"
	Category   Category // used for grouping in /formats
	MaxInput   int64    // 0 = use DefaultMaxInput

	Convert func(ctx context.Context, in io.Reader, out io.Writer) error
}

// DefaultMaxInput is applied when a Formatter leaves MaxInput at zero.
// Matches the existing /convert handler's 10 MB ParseMultipartForm limit.
const DefaultMaxInput int64 = 10 << 20 // 10 MB

// FormatInfo describes a single format (independent of direction) and is
// what GET /formats returns in its `sources` map.
type FormatInfo struct {
	Type     string   `json:"type"`
	MIME     string   `json:"mime"`
	Ext      string   `json:"ext"`
	Category Category `json:"category"`
}

// ConversionEntry is one edge in the conversion graph — what GET /formats
// returns in its `conversions` array.
type ConversionEntry struct {
	From     string   `json:"from"`
	To       string   `json:"to"`
	Category Category `json:"category"`
}

// FormatDetail is the payload for GET /formats/{type}.
type FormatDetail struct {
	Type           string   `json:"type"`
	MIME           string   `json:"mime"`
	Ext            string   `json:"ext"`
	Category       Category `json:"category"`
	CanConvertTo   []string `json:"canConvertTo"`
	CanConvertFrom []string `json:"canConvertFrom"`
}

// FormatListResponse is the payload for GET /formats.
type FormatListResponse struct {
	Sources     map[string]FormatInfo `json:"sources"`
	Conversions []ConversionEntry     `json:"conversions"`
}
