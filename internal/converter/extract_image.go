package converter

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// extractFromImage OCRs a single image.
//
// For v1 this is a straight passthrough to Tesseract. The function is
// structured so that pre-processing (grayscale, binarisation, deskew,
// upscaling) can be slotted in later without touching the dispatcher.
func extractFromImage(data []byte, req ExtractRequest) (*ExtractResponse, error) {
	if len(data) == 0 {
		return nil, errors.New("image file is empty")
	}

	if err := EnsureTesseract(); err != nil {
		return nil, err
	}

	// Pre-processing hook (currently a no-op).
	processed, err := preprocessImageForOCR(data)
	if err != nil {
		return nil, fmt.Errorf("image pre-processing failed: %w", err)
	}

	text, err := runTesseract(processed, req.Lang)
	if err != nil {
		return nil, err
	}

	result := &ExtractResponse{
		PageCount: 1,
		Pages: []PageExtract{
			{Page: 1, Text: text},
		},
	}

	if len(trimWhitespace(text)) == 0 {
		result.Warnings = append(result.Warnings,
			"OCR returned no text — the image may be blank, low-contrast, or in an unsupported script")
	}

	return result, nil
}

// preprocessImageForOCR is a placeholder for future image cleanup.
// For now it returns the input unchanged.
func preprocessImageForOCR(data []byte) ([]byte, error) {
	return data, nil
}

// trimWhitespace removes leading/trailing whitespace from a string.
func trimWhitespace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// io.Discard reference to keep the import if pre-processing is expanded later.
var _ = io.Discard

// ensure os is used even if the file becomes shorter in future edits.
var _ = os.TempDir