package formatters

import "errors"

// Sentinel errors returned by formatters. The API layer maps these to
// HTTP status codes so the handler never has to inspect error strings.
var (
	// ErrUnsupportedPair is returned when Lookup fails or a formatter
	// explicitly rejects the requested (from, to) combination.
	ErrUnsupportedPair = errors.New("conversion pair is not supported")

	// ErrEmptyInput is returned when the uploaded file has zero bytes.
	ErrEmptyInput = errors.New("input file is empty")

	// ErrInputTooLarge is returned when the uploaded file exceeds the
	// per-formatter size cap.
	ErrInputTooLarge = errors.New("input file exceeds the maximum allowed size")

	// ErrInvalidInput is returned when the file does not match the
	// declared source format (e.g. a JPEG that isn't a real JPEG).
	ErrInvalidInput = errors.New("input file is not valid for the declared source format")

	// ErrMissingDependency is returned when a formatter needs an external
	// binary (tesseract, pdftoppm) that is not installed on the host.
	ErrMissingDependency = errors.New("required external dependency is not installed")
)
