package repair

import (
	"encoding/csv"
	"strings"
	"testing"
)

func TestCSV_AlreadyValid(t *testing.T) {
	in := []byte("a,b,c\n1,2,3\n")
	out, report, err := (&csvRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Changed {
		t.Fatalf("expected Changed=false for valid input, fixes: %v", report.Applied)
	}
	if string(out) != string(in) {
		t.Fatalf("bytes changed: got %q", out)
	}
}

func TestCSV_BOMStripped(t *testing.T) {
	in := append([]byte{0xEF, 0xBB, 0xBF}, []byte("a,b\n1,2\n")...)
	out, report, err := (&csvRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "bom-stripped") {
		t.Fatalf("expected bom-stripped, got %v", report.Applied)
	}
	if strings.HasPrefix(string(out), "\uFEFF") || (len(out) >= 3 && out[0] == 0xEF) {
		t.Fatal("BOM still present in output")
	}
}

func TestCSV_RaggedShortRowPadded(t *testing.T) {
	in := []byte("id,name,age\n1,Alice\n")
	_, report, err := (&csvRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "row-padded") {
		t.Fatalf("expected row-padded, got %v", report.Applied)
	}
}

func TestCSV_RaggedLongRowTruncated(t *testing.T) {
	in := []byte("id,name,age\n1,Alice,30,extra1,extra2\n")
	out, report, err := (&csvRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "row-truncated") {
		t.Fatalf("expected row-truncated, got %v", report.Applied)
	}
	// Every row must be 3 fields.
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		rec, err := csv.NewReader(strings.NewReader(line)).Read()
		if err != nil {
			t.Fatalf("output row unreadable: %v", err)
		}
		if len(rec) != 3 {
			t.Fatalf("expected 3 fields, got %d (%q)", len(rec), line)
		}
	}
}

func TestCSV_QuotedDelimiterPreserved(t *testing.T) {
	in := []byte("id,name\n1,\"Doe, John\"\n")
	out, report, err := (&csvRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "Doe, John") {
		t.Fatalf("quoted delimiter lost: %q", out)
	}
	if report.Changed && !contains(report.Applied, "row-") {
		// should be unchanged; check we didn't mangle
		if !strings.Contains(string(out), `"Doe, John"`) {
			t.Fatalf("expected quoted field preserved: %q", out)
		}
	}
}

func TestCSV_SemicolonDelimiter(t *testing.T) {
	in := []byte("a;b;c\n1;2;3\n4;5;6\n")
	out, _, err := (&csvRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output should still use semicolons — the delimiter wasn't
	// changed because it was already consistent.
	if !strings.Contains(string(out), "1;2;3") {
		t.Fatalf("semicolons lost: %q", out)
	}
}

func TestCSV_TrailingBlankLinesDropped(t *testing.T) {
	in := []byte("a,b\n1,2\n\n\n")
	out, report, err := (&csvRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = report
	if strings.HasSuffix(string(out), "\n\n") {
		t.Fatalf("expected no double trailing newline: %q", out)
	}
}

func TestCSV_EmptyInput(t *testing.T) {
	_, _, err := (&csvRepairer{}).Repair(nil, LevelNormal)
	if err != ErrEmptyInput {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
}

func TestCSV_LenientEmptyHeaderRenamed(t *testing.T) {
	in := []byte("a,,c\n1,2,3\n")
	out, report, err := (&csvRepairer{}).Repair(in, LevelLenient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "empty-header-renamed") {
		t.Fatalf("expected empty-header-renamed, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "column_2") {
		t.Fatalf("header not renamed: %q", out)
	}
}