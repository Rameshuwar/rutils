package repair

import "errors"

// Sentinel errors returned by the repair engine.
//
// The API layer maps these to HTTP status codes so the handler never
// has to inspect error strings — mirroring the pattern already used in
// internal/formatters/errors.go.

var (
	// ErrUnsupportedFormat is returned when DetectFormat cannot identify
	// the input, or when no Repairer is registered for the detected format.
	ErrUnsupportedFormat = errors.New("input format is not supported by the repair engine")

	// ErrEmptyInput is returned when the caller hands the engine zero bytes.
	ErrEmptyInput = errors.New("input is empty")

	// ErrUnrepairable is returned when every stage ran and none of them
	// produced a document that a strict parser would accept. The caller
	// should surface the Report alongside the error so the user can see
	// which stages were attempted.
	ErrUnrepairable = errors.New("input could not be repaired within the requested level")

	// ErrInvalidLevel is returned when the caller passes a level other
	// than strict / normal / lenient (or the empty string, which we do
	// NOT silently default — the caller must be explicit).
	ErrInvalidLevel = errors.New("repair level must be 'strict', 'normal', or 'lenient'")
)