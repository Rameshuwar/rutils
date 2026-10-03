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
var formatMeta = map[string]FormatInfo{
	"jpg":  {Type: "jpg", MIME: "image/jpeg", Ext: ".jpg", Category: CategoryImage},
	"png":  {Type: "png", MIME: "image/png", Ext: ".png", Category: CategoryImage},
	"webp": {Type: "webp", MIME: "image/webp", Ext: ".webp", Category: CategoryImage},
	"tiff": {Type: "tiff", MIME: "image/tiff", Ext: ".tiff", Category: CategoryImage},
	"bmp":  {Type: "bmp", MIME: "image/bmp", Ext: ".bmp", Category: CategoryImage},
	"pdf":  {Type: "pdf", MIME: "application/pdf", Ext: ".pdf", Category: CategoryDocument},
	"docx": {Type: "docx", MIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Ext: ".docx", Category: CategoryDocument},
	"txt":  {Type: "txt", MIME: "text/plain; charset=utf-8", Ext: ".txt", Category: CategoryDocument},
	"csv":  {Type: "csv", MIME: "text/csv", Ext: ".csv", Category: CategoryData},
	"json": {Type: "json", MIME: "application/json", Ext: ".json", Category: CategoryData},
}

// Sources returns every format known to the registry, keyed by type.
// Metadata comes from formatMeta — never from individual formatters.
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
			// Unknown format — emit a stub so it still appears,
			// but the missing entry is a bug worth seeing.
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
func Detail(formatType string) (FormatDetail, bool) {
	formatType = strings.ToLower(strings.TrimSpace(formatType))

	regMu.RLock()
	defer regMu.RUnlock()

	detail := FormatDetail{Type: formatType}
	found := false
	seenFrom := map[string]bool{}
	seenTo := map[string]bool{}

	for _, f := range registry {
		if f.From == formatType {
			found = true
			if !seenTo[f.To] {
				detail.CanConvertTo = append(detail.CanConvertTo, f.To)
				seenTo[f.To] = true
			}
		}
		if f.To == formatType {
			found = true
			if !seenFrom[f.From] {
				detail.CanConvertFrom = append(detail.CanConvertFrom, f.From)
				seenFrom[f.From] = true
			}
		}
	}

	if !found {
		return FormatDetail{}, false
	}

	// Pull MIME/ext/category from the authoritative table.
	if meta, ok := formatMeta[formatType]; ok {
		detail.MIME = meta.MIME
		detail.Ext = meta.Ext
		detail.Category = meta.Category
	}

	sort.Strings(detail.CanConvertTo)
	sort.Strings(detail.CanConvertFrom)
	return detail, true
}

// knownOutputMeta is a small fallback table for formats that only ever
// appear as a source (or that are registered with sparse metadata).
// Extend this when adding new formats.
var knownOutputMeta = map[string]FormatInfo{
	"jpg":  {Type: "jpg", MIME: "image/jpeg", Ext: ".jpg", Category: CategoryImage},
	"png":  {Type: "png", MIME: "image/png", Ext: ".png", Category: CategoryImage},
	"webp": {Type: "webp", MIME: "image/webp", Ext: ".webp", Category: CategoryImage},
	"tiff": {Type: "tiff", MIME: "image/tiff", Ext: ".tiff", Category: CategoryImage},
	"bmp":  {Type: "bmp", MIME: "image/bmp", Ext: ".bmp", Category: CategoryImage},
	"pdf":  {Type: "pdf", MIME: "application/pdf", Ext: ".pdf", Category: CategoryDocument},
	"txt":  {Type: "txt", MIME: "text/plain; charset=utf-8", Ext: ".txt", Category: CategoryDocument},
	"csv":  {Type: "csv", MIME: "text/csv", Ext: ".csv", Category: CategoryData},
	"json": {Type: "json", MIME: "application/json", Ext: ".json", Category: CategoryData},
	"docx": {Type: "docx", MIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Ext: ".docx", Category: CategoryDocument},
}
