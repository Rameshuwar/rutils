package formatters

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// registry stores every registered Formatter keyed by "from→to".
// It is populated by init() functions in the individual format files
// (image_jpg.go, image_png.go, ...) which are pulled in by register_all.go.
var (
	regMu    sync.RWMutex
	registry = map[string]Formatter{}
)

// pairKey returns the canonical lookup key for a directed conversion.
// Both sides are lowercased and trimmed so "JPG" → "jpg".
func pairKey(from, to string) string {
	return strings.ToLower(strings.TrimSpace(from)) + "→" +
		strings.ToLower(strings.TrimSpace(to))
}

// Register adds a formatter to the global registry. Calling Register with
// the same (from, to) pair twice panics — a duplicate is always a bug,
// and we want it loud at startup rather than silently shadowed.
func Register(f Formatter) {
	if f.From == "" || f.To == "" {
		panic("formatters: Register called with empty From or To")
	}
	if f.Convert == nil {
		panic(fmt.Sprintf("formatters: Register(%s→%s) has nil Convert", f.From, f.To))
	}
	if f.OutputMIME == "" || f.OutputExt == "" {
		panic(fmt.Sprintf("formatters: Register(%s→%s) missing OutputMIME/OutputExt", f.From, f.To))
	}
	if f.Category == "" {
		panic(fmt.Sprintf("formatters: Register(%s→%s) missing Category", f.From, f.To))
	}
	if f.MaxInput == 0 {
		f.MaxInput = DefaultMaxInput
	}

	key := pairKey(f.From, f.To)

	regMu.Lock()
	defer regMu.Unlock()

	if _, exists := registry[key]; exists {
		panic(fmt.Sprintf("formatters: duplicate registration for %s", key))
	}
	registry[key] = f
}

// Lookup returns the Formatter for a directed pair, or (zero, false).
func Lookup(from, to string) (Formatter, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	f, ok := registry[pairKey(from, to)]
	return f, ok
}

// All returns every registered formatter as a slice, sorted by
// (from, to) for deterministic output. Used by tests and /formats.
func All() []Formatter {
	regMu.RLock()
	defer regMu.RUnlock()

	out := make([]Formatter, 0, len(registry))
	for _, f := range registry {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// formatMeta is the authoritative description of every format the
// registry knows about. This is what Sources() and Detail() consult —
// never the individual Formatter.Category field (which describes the
// edge, not the format).
//
// When a new format is added to any formatter's From or To field, an
// entry must be added here too. Otherwise /formats emits a stub with
// empty MIME/Ext fields, and the frontend dropdown renders a blank
// label. The registry_test.go file enforces this contract.
var formatMeta = map[string]FormatInfo{
	// ---- Images ----
	"jpg":  {Type: "jpg", MIME: "image/jpeg", Ext: ".jpg", Category: CategoryImage},
	"jpeg": {Type: "jpeg", MIME: "image/jpeg", Ext: ".jpeg", Category: CategoryImage},
	"png":  {Type: "png", MIME: "image/png", Ext: ".png", Category: CategoryImage},
	"webp": {Type: "webp", MIME: "image/webp", Ext: ".webp", Category: CategoryImage},
	"tiff": {Type: "tiff", MIME: "image/tiff", Ext: ".tiff", Category: CategoryImage},
	"bmp":  {Type: "bmp", MIME: "image/bmp", Ext: ".bmp", Category: CategoryImage},

	// ---- Documents (layout-preserving office formats) ----
	"pdf":  {Type: "pdf", MIME: "application/pdf", Ext: ".pdf", Category: CategoryDocument},
	"docx": {Type: "docx", MIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Ext: ".docx", Category: CategoryDocument},
	"pptx": {Type: "pptx", MIME: "application/vnd.openxmlformats-officedocument.presentationml.presentation", Ext: ".pptx", Category: CategoryDocument},
	"xlsx": {Type: "xlsx", MIME: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Ext: ".xlsx", Category: CategoryDocument},

	// ---- Legacy / OpenDocument office formats ----
	"rtf": {Type: "rtf", MIME: "application/rtf", Ext: ".rtf", Category: CategoryDocument},
	"odt": {Type: "odt", MIME: "application/vnd.oasis.opendocument.text", Ext: ".odt", Category: CategoryDocument},
	"ods": {Type: "ods", MIME: "application/vnd.oasis.opendocument.spreadsheet", Ext: ".ods", Category: CategoryDocument},
	"odp": {Type: "odp", MIME: "application/vnd.oasis.opendocument.presentation", Ext: ".odp", Category: CategoryDocument},

	// ---- Markup ----
	"html": {Type: "html", MIME: "text/html; charset=utf-8", Ext: ".html", Category: CategoryDocument},
	"htm":  {Type: "htm", MIME: "text/html; charset=utf-8", Ext: ".htm", Category: CategoryDocument},

	// ---- Plain text and data ----
	"txt":  {Type: "txt", MIME: "text/plain; charset=utf-8", Ext: ".txt", Category: CategoryDocument},
	"csv":  {Type: "csv", MIME: "text/csv; charset=utf-8", Ext: ".csv", Category: CategoryData},
	"json": {Type: "json", MIME: "application/json", Ext: ".json", Category: CategoryData},
}

// Sources returns every format known to the registry, keyed by type.
// Metadata comes from formatMeta — never from individual formatters.
//
// If a format appears as a From/To in the registry but is missing from
// formatMeta, the returned FormatInfo carries only the Type field and
// leaves MIME/Ext/Category empty. Tests in registry_test.go enforce
// that no such format exists.
func Sources() map[string]FormatInfo {
	regMu.RLock()
	defer regMu.RUnlock()

	seen := map[string]bool{}
	for _, f := range registry {
		seen[f.From] = true
		seen[f.To] = true
	}

	out := make(map[string]FormatInfo, len(seen))
	for t := range seen {
		if meta, ok := formatMeta[t]; ok {
			out[t] = meta
		} else {
			// Unknown format — emit a stub so it still appears, but
			// the missing entry is a bug worth seeing in the response.
			out[t] = FormatInfo{Type: t}
		}
	}
	return out
}

// Conversions returns every registered edge, sorted.
func Conversions() []ConversionEntry {
	all := All()
	out := make([]ConversionEntry, 0, len(all))
	for _, f := range all {
		out = append(out, ConversionEntry{
			From:     f.From,
			To:       f.To,
			Category: f.Category,
		})
	}
	return out
}

// Detail returns the per-format view used by GET /formats/{type}.
//
// The MIME, Ext, and Category fields come from formatMeta — the single
// authoritative source of format metadata.
//
// IMPORTANT: The metadata lookup happens BEFORE the "not found" check,
// so a format that is present in formatMeta but has no registered edges
// (a "reserved" format) still gets its metadata populated. This is what
// allows /formats/{type} to return accurate information for formats
// that we have declared but not yet wired up — for example, "xlsx"
// before the Excel formatters are registered.
//
// If the format is present in NEITHER formatMeta NOR the registry, the
// function returns (zero, false) and the API layer responds 404.
func Detail(formatType string) (FormatDetail, bool) {
	formatType = strings.ToLower(strings.TrimSpace(formatType))
	if formatType == "" {
		return FormatDetail{}, false
	}

	regMu.RLock()
	defer regMu.RUnlock()

	detail := FormatDetail{Type: formatType}

	// ---- Step 1: Pull metadata from the authoritative table ----
	//
	// We do this FIRST, before the registry walk, so that the presence
	// of metadata alone is enough to make the format "known". This
	// decouples "the format exists" from "the format has edges".
	meta, inMeta := formatMeta[formatType]
	if inMeta {
		detail.MIME = meta.MIME
		detail.Ext = meta.Ext
		detail.Category = meta.Category
	}

	// ---- Step 2: Walk the registry to collect reachable targets/sources ----
	foundInRegistry := false
	seenFrom := map[string]bool{}
	seenTo := map[string]bool{}

	for _, f := range registry {
		if f.From == formatType {
			foundInRegistry = true
			if !seenTo[f.To] {
				detail.CanConvertTo = append(detail.CanConvertTo, f.To)
				seenTo[f.To] = true
			}
		}
		if f.To == formatType {
			foundInRegistry = true
			if !seenFrom[f.From] {
				detail.CanConvertFrom = append(detail.CanConvertFrom, f.From)
				seenFrom[f.From] = true
			}
		}
	}

	// ---- Step 3: Decide whether the format is "known" ----
	//
	// A format is known if it appears in formatMeta OR in the registry.
	// Only when it appears in neither do we treat the request as a 404.
	if !inMeta && !foundInRegistry {
		return FormatDetail{}, false
	}

	sort.Strings(detail.CanConvertTo)
	sort.Strings(detail.CanConvertFrom)
	return detail, true
}