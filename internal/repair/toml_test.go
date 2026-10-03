package repair

import (
	"strings"
	"testing"
)

func TestTOML_SemicolonComment(t *testing.T) {
	in := []byte("key = 1 ; comment\n")
	out, report, err := (&tomlRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "semicolon-comment") {
		t.Fatalf("expected semicolon-comment, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "# comment") {
		t.Fatalf("comment marker not converted: %q", out)
	}
}

func TestTOML_QuoteUnquotedValueWithSpaces(t *testing.T) {
	in := []byte(`name = Alice Smith` + "\n")
	out, report, err := (&tomlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "value-quoted") {
		t.Fatalf("expected value-quoted, got %v", report.Applied)
	}
	if !strings.Contains(string(out), `"Alice Smith"`) {
		t.Fatalf("value not quoted: %q", out)
	}
}

func TestTOML_NumberUntouched(t *testing.T) {
	in := []byte("count = 42\n")
	_, report, err := (&tomlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contains(report.Applied, "value-quoted") {
		t.Fatalf("number was quoted: %v", report.Applied)
	}
}

func TestTOML_DropJunkLineLenient(t *testing.T) {
	in := []byte("good = 1\nthis is junk\n")
	out, report, err := (&tomlRepairer{}).Repair(in, LevelLenient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "junk-line-dropped") {
		t.Fatalf("expected junk-line-dropped, got %v", report.Applied)
	}
	if strings.Contains(string(out), "this is junk") {
		t.Fatalf("junk line survived: %q", out)
	}
}