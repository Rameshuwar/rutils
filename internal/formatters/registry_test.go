package formatters

import (
	"context"
	"io"
	"sort"
	"strings"
	"testing"
)

// ============================================================
// Helpers
// ============================================================

// noopConvert is a stub Convert func used by tests that only care about
// registry metadata, not about actually transforming bytes.
func noopConvert(_ context.Context, _ io.Reader, _ io.Writer) error {
	return nil
}

// newStub builds a Formatter with sane defaults for the registry tests.
// Callers override whichever fields they care about.
func newStub(from, to string) Formatter {
	return Formatter{
		From:       from,
		To:         to,
		OutputMIME: "application/octet-stream",
		OutputExt:  "." + to,
		Category:   CategoryData,
		Convert:    noopConvert,
	}
}

// ============================================================
// Register / duplicate detection
// ============================================================

func TestRegister_RejectsDuplicatePair(t *testing.T) {
	// Register a unique pair so we don't collide with any real formatter.
	f := newStub("zzz_test_src", "zzz_test_dst")
	Register(f)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration, got none")
		}
	}()

	// Registering the exact same (from, to) again must panic.
	Register(f)
}

func TestRegister_RejectsMissingFields(t *testing.T) {
	cases := []struct {
		name string
		f    Formatter
	}{
		{
			name: "missing From",
			f: Formatter{
				To:         "zzz_dst",
				OutputMIME: "text/plain",
				OutputExt:  ".txt",
				Category:   CategoryData,
				Convert:    noopConvert,
			},
		},
		{
			name: "missing Convert",
			f: Formatter{
				From:       "zzz_src",
				To:         "zzz_dst",
				OutputMIME: "text/plain",
				OutputExt:  ".txt",
				Category:   CategoryData,
			},
		},
		{
			name: "missing OutputMIME",
			f: Formatter{
				From:      "zzz_src",
				To:        "zzz_dst",
				OutputExt: ".txt",
				Category:  CategoryData,
				Convert:   noopConvert,
			},
		},
		{
			name: "missing OutputExt",
			f: Formatter{
				From:       "zzz_src",
				To:         "zzz_dst",
				OutputMIME: "text/plain",
				Category:   CategoryData,
				Convert:    noopConvert,
			},
		},
		{
			name: "missing Category",
			f: Formatter{
				From:       "zzz_src",
				To:         "zzz_dst",
				OutputMIME: "text/plain",
				OutputExt:  ".txt",
				Convert:    noopConvert,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic for %s, got none", tc.name)
				}
			}()
			Register(tc.f)
		})
	}
}

func TestRegister_AppliesDefaultMaxInput(t *testing.T) {
	// Register a unique pair with MaxInput left at zero.
	from, to := "zzz_default_max_src", "zzz_default_max_dst"
	Register(newStub(from, to))

	got, ok := Lookup(from, to)
	if !ok {
		t.Fatalf("expected %s→%s to be registered", from, to)
	}
	if got.MaxInput != DefaultMaxInput {
		t.Fatalf("expected default MaxInput %d, got %d", DefaultMaxInput, got.MaxInput)
	}
}

// ============================================================
// Lookup
// ============================================================

func TestLookup_CaseInsensitive(t *testing.T) {
	from, to := "zzz_case_src", "zzz_case_dst"
	Register(newStub(from, to))

	// Mixed case + surrounding whitespace must still resolve.
	got, ok := Lookup("  ZZZ_CASE_SRC  ", "ZZZ_CASE_DST")
	if !ok {
		t.Fatal("expected mixed-case lookup to succeed")
	}
	if got.From != from || got.To != to {
		t.Fatalf("expected %s→%s, got %s→%s", from, to, got.From, got.To)
	}
}

func TestLookup_UnknownPairReturnsFalse(t *testing.T) {
	if _, ok := Lookup("zzz_never_src", "zzz_never_dst"); ok {
		t.Fatal("expected unknown pair to return false")
	}
}

// ============================================================
// All
// ============================================================

func TestAll_IsSortedAndNonEmpty(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("expected non-empty registry — check register_all.go init()")
	}

	// Verify sorted by (From, To).
	for i := 1; i < len(all); i++ {
		prev, cur := all[i-1], all[i]
		if prev.From > cur.From {
			t.Fatalf("not sorted by From: %s > %s at index %d", prev.From, cur.From, i)
		}
		if prev.From == cur.From && prev.To > cur.To {
			t.Fatalf("not sorted by To within From=%s: %s > %s at index %d",
				prev.From, prev.To, cur.To, i)
		}
	}
}

// ============================================================
// Sources
// ============================================================

func TestSources_IncludesEveryKnownFormat(t *testing.T) {
	srcs := Sources()
	if len(srcs) == 0 {
		t.Fatal("expected non-empty sources map")
	}

	// Every real formatter's From and To must appear as a source entry.
	for _, f := range All() {
		if _, ok := srcs[f.From]; !ok {
			t.Errorf("expected source %q to be present", f.From)
		}
		if _, ok := srcs[f.To]; !ok {
			t.Errorf("expected target %q to be present as a source entry", f.To)
		}
	}
}

func TestSources_TargetsHaveMIMEAndExt(t *testing.T) {
	// Every *real* format (one present in formatMeta) that appears as a
	// target must have MIME + Ext populated. Test stubs like "zzz_test_dst"
	// are intentionally excluded — they're registered by other tests and
	// are not expected to have metadata.
	srcs := Sources()
	for _, f := range All() {
		if _, isReal := formatMeta[f.To]; !isReal {
			continue
		}
		info, ok := srcs[f.To]
		if !ok {
			t.Fatalf("target %q missing from sources map", f.To)
		}
		if info.MIME == "" {
			t.Errorf("target %q has empty MIME", f.To)
		}
		if info.Ext == "" {
			t.Errorf("target %q has empty Ext", f.To)
		}
	}
}
// ============================================================
// Conversions
// ============================================================

func TestConversions_MatchesAll(t *testing.T) {
	all := All()
	convs := Conversions()

	if len(convs) != len(all) {
		t.Fatalf("expected %d conversions, got %d", len(all), len(convs))
	}

	// Build a set of "from→to" from All() and verify every conversion
	// edge appears in it.
	want := map[string]bool{}
	for _, f := range all {
		want[f.From+"→"+f.To] = true
	}
	for _, c := range convs {
		if !want[c.From+"→"+c.To] {
			t.Errorf("unexpected conversion %s→%s", c.From, c.To)
		}
	}
}

// ============================================================
// Detail
// ============================================================

func TestDetail_KnownFormat(t *testing.T) {
	// "jpg" is guaranteed to be registered by image_jpg.go.
	detail, ok := Detail("jpg")
	if !ok {
		t.Fatal("expected jpg to be registered")
	}
	if detail.Type != "jpg" {
		t.Fatalf("expected type=jpg, got %q", detail.Type)
	}
	if detail.MIME == "" || detail.Ext == "" {
		t.Fatalf("expected non-empty MIME/Ext, got %q / %q", detail.MIME, detail.Ext)
	}
	if len(detail.CanConvertTo) == 0 {
		t.Fatal("expected jpg to have at least one reachable target")
	}

	// CanConvertTo must be sorted.
	if !sort.StringsAreSorted(detail.CanConvertTo) {
		t.Fatalf("CanConvertTo not sorted: %v", detail.CanConvertTo)
	}
	if !sort.StringsAreSorted(detail.CanConvertFrom) {
		t.Fatalf("CanConvertFrom not sorted: %v", detail.CanConvertFrom)
	}
}

func TestDetail_CaseInsensitive(t *testing.T) {
	lower, ok1 := Detail("jpg")
	upper, ok2 := Detail("  JPG  ")

	if !ok1 || !ok2 {
		t.Fatal("expected both lookups to succeed")
	}
	if lower.Type != upper.Type || lower.MIME != upper.MIME {
		t.Fatalf("case-sensitive lookup mismatch: %+v vs %+v", lower, upper)
	}
}

func TestDetail_UnknownFormat(t *testing.T) {
	if _, ok := Detail("zzz_madeup_format"); ok {
		t.Fatal("expected unknown format to return false")
	}
}

func TestDetail_ReachableTargetsMatchRegistry(t *testing.T) {
	// For a known source, every entry in CanConvertTo must correspond
	// to a registered formatter whose From matches.
	detail, ok := Detail("jpg")
	if !ok {
		t.Fatal("expected jpg to be registered")
	}

	for _, to := range detail.CanConvertTo {
		if _, exists := Lookup("jpg", to); !exists {
			t.Errorf("Detail claims jpg→%s is reachable but Lookup says it isn't", to)
		}
	}
}

// ============================================================
// knownOutputMeta coverage
// ============================================================

func TestFormatMeta_HasAllCommonFormats(t *testing.T) {
	required := []string{
		"jpg", "png", "webp", "tiff", "bmp",
		"pdf", "txt", "csv", "json", "docx",
	}
	for _, k := range required {
		if _, ok := formatMeta[k]; !ok {
			t.Errorf("formatMeta is missing entry for %q", k)
		}
	}
}

// ============================================================
// init-time sanity (implicit coverage via other tests)
// ============================================================

func TestRegistry_NonNullConvertForEveryEntry(t *testing.T) {
	for _, f := range All() {
		if f.Convert == nil {
			t.Errorf("formatter %s→%s has nil Convert", f.From, f.To)
		}
	}
}

func TestRegistry_NoEmptyStrings(t *testing.T) {
	for _, f := range All() {
		if strings.TrimSpace(f.From) == "" ||
			strings.TrimSpace(f.To) == "" ||
			strings.TrimSpace(f.OutputMIME) == "" ||
			strings.TrimSpace(f.OutputExt) == "" {
			t.Errorf("formatter %s→%s has empty metadata: %+v", f.From, f.To, f)
		}
	}
}
