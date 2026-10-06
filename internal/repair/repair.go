package repair

import (
	"bytes"
	"path/filepath"
	"strings"
	"sync"
)

// Level controls how aggressive the repairer is allowed to be.
type Level string

const (
	// LevelStrict applies only stage 0 (normalize) and stage 1 (lexical).
	// These transformations are provably lossless: no byte that carries
	// meaning is ever removed. Use this when the caller cannot tolerate
	// any semantic ambiguity.
	LevelStrict Level = "strict"

	// LevelNormal additionally applies stage 2 (structural). Structural
	// fixes resolve genuine syntax errors (trailing commas, missing
	// commas, single quotes) but never change a *value*. This is the
	// recommended default.
	LevelNormal Level = "normal"

	// LevelLenient additionally applies stage 3 (value coercion) and
	// stage 4 (bracket balancing). These CAN change the meaning of a
	// document — e.g. "NaN" → null. Every such change is recorded in
	// Report.Warnings so the caller can see exactly what happened.
	LevelLenient Level = "lenient"
)

// ParseLevel normalises a caller-supplied string into a Level.
// The empty string is NOT accepted — callers must be explicit. This
// mirrors the strictness of the auth package's config loader.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "strict":
		return LevelStrict, nil
	case "normal":
		return LevelNormal, nil
	case "lenient":
		return LevelLenient, nil
	default:
		return "", ErrInvalidLevel
	}
}

// Repairer is implemented by every supported format. One instance per
// format is registered in init() — the same pattern used by
// internal/formatters.
type Repairer interface {
	// Format returns the canonical lower-case name: "json", "csv", ...
	Format() string

	// Repair takes the raw uploaded bytes and returns a repaired copy.
	// The input slice must not be mutated. If the input was already
	// valid, Repair returns the original bytes unchanged with
	// Report.Changed == false.
	Repair(raw []byte, level Level) ([]byte, Report, error)
}

// registry holds every registered Repairer, keyed by Format().
// Populated by init() functions in json.go, csv.go, xml.go, yaml.go.
var (
	regMu    sync.RWMutex
	registry = map[string]Repairer{}
)

// register is called from each format file's init(). Duplicate
// registrations panic — a duplicate is always a bug and we want it
// loud at startup, exactly as internal/formatters does.
func register(r Repairer) {
	regMu.Lock()
	defer regMu.Unlock()

	name := strings.ToLower(r.Format())
	if name == "" {
		panic("repair: Register called with empty Format()")
	}
	if _, exists := registry[name]; exists {
		panic("repair: duplicate registration for format " + name)
	}
	registry[name] = r
}

// Engine is the top-level dispatcher the API layer calls. It is safe
// for concurrent use.
type Engine struct{}

// NewEngine returns an Engine backed by the global registry. There is
// exactly one logical engine per process; the struct is empty so it can
// be copied freely.
func NewEngine() *Engine { return &Engine{} }

func (e *Engine) DetectFormat(filename string, raw []byte) (string, error) {
	if len(raw) == 0 {
		return "", ErrEmptyInput
	}

	// 1. Extension — authoritative when it matches a registered
	//    repairer. This lets a file named `pasted.ini` or `broken.csv`
	//    bypass magic-byte sniffing entirely, because the user (or the
	//    frontend) has already declared the format explicitly.
	//
	//    This is important because several valid formats share prefixes
	//    with JSON (e.g. INI files start with `[section]`, which the
	//    JSON sniffer would otherwise claim as a `[` array).
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
	if ext != "" {
		if _, ok := lookup(ext); ok {
			return ext, nil
		}
	}

	// 2. Magic-byte sniffing — fallback for extensionless or
	//    unrecognised-extension uploads (e.g. an image dropped in
	//    with no name).
	if format := sniffFormat(raw); format != "" {
		if _, ok := lookup(format); ok {
			return format, nil
		}
	}

	return "", ErrUnsupportedFormat
}

// Repair is the single entry point for the API layer. It detects the
// format, dispatches to the registered Repairer, and returns the
// repaired bytes plus a Report.
func (e *Engine) Repair(filename string, raw []byte, level Level) ([]byte, Report, error) {
	if len(raw) == 0 {
		return nil, Report{}, ErrEmptyInput
	}

	// Validate the level up front — an unknown level is a caller bug,
	// not a repair failure.
	if level != LevelStrict && level != LevelNormal && level != LevelLenient {
		return nil, Report{}, ErrInvalidLevel
	}

	format, err := e.DetectFormat(filename, raw)
	if err != nil {
		return nil, Report{}, err
	}

	r, ok := lookup(format)
	if !ok {
		return nil, Report{}, ErrUnsupportedFormat
	}

	// Repairers must not mutate their input. We defensively copy so a
	// buggy future repairer cannot corrupt the caller's slice.
	safeCopy := bytes.Clone(raw)

	repaired, report, err := r.Repair(safeCopy, level)
	if err != nil {
		return nil, report, err
	}

	// Defensive: if a repairer forgot to set Format/Level, fill them in.
	if report.Format == "" {
		report.Format = format
	}
	if report.Level == "" {
		report.Level = level
	}

	return repaired, report, nil
}

// lookup returns the registered Repairer for a canonical format name.
func lookup(format string) (Repairer, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	r, ok := registry[format]
	return r, ok
}

// sniffFormat examines the leading bytes of raw and returns the
// canonical format name if the magic bytes are decisive. Returns "" if
// the bytes are ambiguous and the caller should fall back to extension.
func sniffFormat(raw []byte) string {
	// Skip a UTF-8 BOM if present — it is common in Windows-saved files
	// and would otherwise mask the real first byte.
	raw = trimBOM(raw)
	if len(raw) == 0 {
		return ""
	}

	// JSON: first non-space byte is { or [ (object or array).
	trimmed := trimLeadingSpace(raw)
	if len(trimmed) > 0 {
		switch trimmed[0] {
		case '{':
			return "json"
		case '[':
			// Distinguish JSON array from INI section header.
			// JSON arrays start with `[` followed by a value or `]`:
			//   [1,2,3]  ["a","b"]  []  [{"k":1}]  [ true ]
			// INI sections start with `[` followed by a name and `]`:
			//   [section]  [server]  [my.section]
			//
			// Heuristic: an INI section header contains a `]` on the
			// SAME line, with no comma or quote between `[` and `]`.
			if isINISectionHeader(trimmed) {
				return "ini"
			}
			return "json"
		}
	}

	// XML: `<?xml` or `<` at the very start, and not `<!DOCTYPE html`.
	if len(raw) >= 5 && bytes.EqualFold(raw[:5], []byte("<?xml")) {
		return "xml"
	}

	// YAML: extremely hard to sniff reliably. The `---` document start
	// marker is decisive when present.
	if len(raw) >= 3 && bytes.Equal(raw[:3], []byte("---")) {
		return "yaml"
	}

	return ""
}

// isINISectionHeader reports whether the first line of `b` looks like
// an INI section header — i.e. `[name]` — as opposed to a JSON array.
//
// A JSON array's first line never contains an unquoted `]` before a
// comma or another value: `[1]` is technically valid JSON, but in
// practice a file whose first line is `[something]` with no commas,
// no quotes, and no digits-in-brackets is far more likely to be INI.
func isINISectionHeader(b []byte) bool {
	// Find the end of the first line.
	lineEnd := bytes.IndexByte(b, '\n')
	if lineEnd < 0 {
		lineEnd = len(b)
	}
	firstLine := bytes.TrimSpace(b[:lineEnd])

	// Must start with `[` and end with `]`.
	if len(firstLine) < 3 {
		return false
	}
	if firstLine[0] != '[' || firstLine[len(firstLine)-1] != ']' {
		// Not a single-line `[...]` — could be a multi-line JSON array.
		return false
	}

	// The inside must not contain a comma (that would make it a JSON
	// array) or a quote (that would make it a JSON string array).
	inner := firstLine[1 : len(firstLine)-1]
	if bytes.ContainsAny(inner, `,"'`) {
		return false
	}

	// The inside must look like a section name: letters, digits, dots,
	// underscores, hyphens, spaces. No JSON-significant punctuation.
	for _, c := range inner {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '.' || c == '_' || c == '-' || c == ' ' || c == '\t':
		default:
			return false
		}
	}

	return true
}

// trimBOM removes a leading UTF-8 BOM (EF BB BF) if present.
func trimBOM(b []byte) []byte {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return b[3:]
	}
	return b
}

// trimLeadingSpace returns b with leading ASCII whitespace removed.
// Kept local so this file has no import beyond bytes/path/sync/strings.
func trimLeadingSpace(b []byte) []byte {
	i := 0
	for i < len(b) {
		switch b[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return b[i:]
		}
	}
	return b[i:]
}