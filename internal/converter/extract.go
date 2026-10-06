package converter

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ============================================================
// Public types
// ============================================================

// ExtractRequest carries the validated parameters for a text-extraction job.
type ExtractRequest struct {
	Lang   string // ISO 639-1 language code for OCR (default "eng")
	Page   int    // 0 = all pages (PDF only); 1-indexed otherwise
	Format string // "plain" (default) or "structured"
}

// ExtractResponse is the unified JSON response for /extract-text.
type ExtractResponse struct {
	SourceType  string        `json:"sourceType"` // "pdf" | "image"
	PageCount   int           `json:"pageCount"`
	ExtractedAt string        `json:"extractedAt"` // RFC3339 UTC
	Language    string        `json:"language"`
	Text        string        `json:"text"`               // concatenated text
	Pages       []PageExtract `json:"pages,omitempty"`    // per-page breakdown
	Warnings    []string      `json:"warnings,omitempty"` // non-fatal notices
}

// PageExtract is one page's worth of extracted text.
type PageExtract struct {
	Page int    `json:"page"`
	Text string `json:"text"`
}

// ============================================================
// Supported source types
// ============================================================

const (
	SourcePDF   = "pdf"
	SourceImage = "image"

	// MaxExtractUploadSize caps the bytes accepted by /extract-text.
	// Kept smaller than the PDF-size endpoint because OCR is CPU-heavy.
	MaxExtractUploadSize int64 = 50 * 1024 * 1024 // 50 MB

	// MaxExtractPages caps PDF page count to defend against zip-bomb PDFs.
	MaxExtractPages = 500
)

// SupportedImageExts lists file extensions treated as images.
var SupportedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".tiff": true,
	".tif":  true,
	".bmp":  true,
}

// SupportedExtractExts is the union of PDF + image extensions.
var SupportedExtractExts = func() map[string]bool {
	m := map[string]bool{".pdf": true}
	for k := range SupportedImageExts {
		m[k] = true
	}
	return m
}()

// ============================================================
// Source-type detection
// ============================================================

// DetectSourceType inspects the filename and magic bytes to decide
// whether the upload is a PDF or an image. Returns an error if neither.
func DetectSourceType(filename string, data []byte) (string, error) {
	if len(data) < 5 {
		return "", errors.New("file is too small to identify")
	}

	// PDF magic bytes take priority — a .jpg-named PDF is still a PDF.
	if strings.HasPrefix(string(data[:5]), "%PDF-") {
		return SourcePDF, nil
	}

	// Image magic bytes.
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return SourceImage, nil // JPEG
	case len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n":
		return SourceImage, nil // PNG
	case len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return SourceImage, nil // WEBP
	case len(data) >= 4 && (string(data[:4]) == "II*\x00" || string(data[:4]) == "MM\x00*"):
		return SourceImage, nil // TIFF (little/big endian)
	case len(data) >= 2 && string(data[:2]) == "BM":
		return SourceImage, nil // BMP
	}

	// Fall back to extension if magic bytes are inconclusive.
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".pdf") {
		return SourcePDF, nil
	}
	for ext := range SupportedImageExts {
		if strings.HasSuffix(lower, ext) {
			return SourceImage, nil
		}
	}

	return "", fmt.Errorf("unsupported file type: only PDF and image files are supported")
}

// ============================================================
// Top-level dispatcher
// ============================================================

// ExtractText is the single entry point used by the HTTP handler.
// It validates the request, detects the source type, and routes to
// the PDF or image pipeline.
func ExtractText(filename string, data []byte, req ExtractRequest) (*ExtractResponse, error) {
	if len(data) == 0 {
		return nil, errors.New("uploaded file is empty")
	}

	if int64(len(data)) > MaxExtractUploadSize {
		return nil, fmt.Errorf(
			"file exceeds the %d MB limit for extraction",
			MaxExtractUploadSize/(1024*1024),
		)
	}

	// Normalise request.
	if req.Lang == "" {
		req.Lang = "eng"
	}
	if req.Format == "" {
		req.Format = "plain"
	}
	if req.Format != "plain" && req.Format != "structured" {
		return nil, errors.New("format must be 'plain' or 'structured'")
	}

	sourceType, err := DetectSourceType(filename, data)
	if err != nil {
		return nil, err
	}

	var result *ExtractResponse

	switch sourceType {
	case SourcePDF:
		result, err = extractFromPDF(data, req)
	case SourceImage:
		result, err = extractFromImage(data, req)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}
	if err != nil {
		return nil, err
	}

	result.SourceType = sourceType
	result.Language = req.Lang
	result.ExtractedAt = time.Now().UTC().Format(time.RFC3339)

	// Concatenate per-page text into the top-level `text` field.
	if len(result.Pages) > 0 {
		var b strings.Builder
		for i, p := range result.Pages {
			if i > 0 {
				b.WriteString("\n\n")
			}
			b.WriteString(p.Text)
		}
		result.Text = b.String()
	}

	// In "plain" mode, drop the per-page array to keep the payload small.
	if req.Format == "plain" {
		result.Pages = nil
	}

	if result.PageCount == 0 {
		result.PageCount = len(result.Pages)
		if result.PageCount == 0 {
			result.PageCount = 1
		}
	}

	return result, nil
}
