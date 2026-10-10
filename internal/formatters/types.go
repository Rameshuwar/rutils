package formatters

import (
	"context"
	"io"
	"time"
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
//
// Timeout, when non-zero, caps the wall-clock time the API layer allows
// for this conversion. When zero, the API layer falls back to
// DefaultTimeoutForCategory(f.Category).
type Formatter struct {
	From       string        // e.g. "jpg"
	To         string        // e.g. "png"
	OutputMIME string        // e.g. "image/png"
	OutputExt  string        // e.g. ".png"
	Category   Category      // used for grouping in /formats
	MaxInput   int64         // 0 = use DefaultMaxInput
	Timeout    time.Duration // 0 = use DefaultTimeoutForCategory(Category)

	Convert func(ctx context.Context, in io.Reader, out io.Writer) error
}

// DefaultMaxInput is applied when a Formatter leaves MaxInput at zero.
// Matches the existing /convert handler's 10 MB ParseMultipartForm limit.
const DefaultMaxInput int64 = 10 << 20 // 10 MB

// Default timeouts, keyed by category. These are the budgets the API
// layer applies when a Formatter leaves its Timeout field at zero.
//
// The values are chosen to be generous enough that a legitimate slow
// conversion (a 500-page DOCX through LibreOffice) still succeeds, while
// tight enough that a runaway subprocess is killed before it can consume
// a request slot for an unreasonable amount of time.
const (
	// DefaultTimeoutNative covers pure-Go conversions: CSV↔JSON,
	// TXT↔JSON, image↔image, numeral, measurement, timezone. These
	// run in-process and should complete in well under a second.
	DefaultTimeoutNative = 15 * time.Second

	// DefaultTimeoutImage covers image processing that shells out to
	// pdftocairo, cjpeg, pngquant, or oxipng. These are fast but
	// non-trivial: a 4000×3000 PNG through pngquant at speed 1 can
	// take a few seconds.
	DefaultTimeoutImage = 60 * time.Second

	// DefaultTimeoutDocument covers office-document conversions that
	// go through LibreOffice. Warm-up is ~2s, and a large document
	// can take 30+ seconds on the first request after a cold start.
	DefaultTimeoutDocument = 90 * time.Second

	// DefaultTimeoutRasterize covers PDF→image conversions and any
	// other operation that rasterizes a full document.
	DefaultTimeoutRasterize = 120 * time.Second
)

// DefaultTimeoutForCategory returns the timeout the API layer should
// apply when a Formatter does not declare its own.
//
// The mapping is category-based, not format-based. This keeps the
// policy in one place: if we later decide "documents should get 180s
// instead of 90s", we change one constant here and every document
// formatter inherits the new budget.
func DefaultTimeoutForCategory(c Category) time.Duration {
	switch c {
	case CategoryImage:
		return DefaultTimeoutImage
	case CategoryDocument:
		return DefaultTimeoutDocument
	case CategoryData, CategorySpreadsheet, CategoryMarkup:
		return DefaultTimeoutNative
	default:
		// Unknown category — fall back to the most generous budget
		// rather than the most restrictive, on the theory that an
		// unrecognized category is a bug and should not cause a
		// false timeout in production.
		return DefaultTimeoutDocument
	}
}

// EffectiveTimeout returns the timeout the API layer should apply to
// this Formatter: the formatter's own Timeout if set, otherwise the
// category default.
func (f Formatter) EffectiveTimeout() time.Duration {
	if f.Timeout > 0 {
		return f.Timeout
	}
	return DefaultTimeoutForCategory(f.Category)
}

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