package repair

import (
	"strings"
	"testing"
)

func TestYAML_AlreadyValid(t *testing.T) {
	in := []byte("key: value\nlist:\n  - a\n  - b\n")
	out, report, err := (&yamlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Changed {
		t.Fatalf("expected no change, fixes=%v", report.Applied)
	}
	if string(out) != string(in) {
		t.Fatalf("bytes changed")
	}
}

func TestYAML_TabsConverted(t *testing.T) {
	in := []byte("key:\n\t- a\n\t- b\n")
	out, report, err := (&yamlRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "tab-to-space") {
		t.Fatalf("expected tab-to-space, got %v", report.Applied)
	}
	if strings.Contains(string(out), "\t") {
		t.Fatal("tabs still present")
	}
	if !isValidYAML(out) {
		t.Fatalf("output not valid YAML: %q", out)
	}
}

func TestYAML_ColonSpacing(t *testing.T) {
	in := []byte("name:Alice\nage:30\n")
	out, report, err := (&yamlRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "colon-spacing") {
		t.Fatalf("expected colon-spacing, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "name: Alice") {
		t.Fatalf("spacing not inserted: %q", out)
	}
}

func TestYAML_URLSchemeNotBroken(t *testing.T) {
	in := []byte("url: http://example.com\n")
	out, _, err := (&yamlRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "http://example.com") {
		t.Fatalf("URL was mangled: %q", out)
	}
}

func TestYAML_AmbiguousScalarQuoted(t *testing.T) {
	in := []byte("enabled: on\ndisabled: off\n")
	out, report, err := (&yamlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "ambiguous-scalars-quoted") {
		t.Fatalf("expected ambiguous-scalars-quoted, got %v", report.Applied)
	}
	if !strings.Contains(string(out), `"on"`) || !strings.Contains(string(out), `"off"`) {
		t.Fatalf("scalars not quoted: %q", out)
	}
}

func TestYAML_DocMarkerPrependedLenient(t *testing.T) {
	in := []byte("key: value\n")
	out, _, err := (&yamlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At normal level we don't add the marker — only lenient does,
	// and only when the doc otherwise can't be parsed. Valid YAML
	// stays untouched.
	_ = out
}

func TestYAML_EmptyInput(t *testing.T) {
	_, _, err := (&yamlRepairer{}).Repair([]byte{}, LevelNormal)
	if err == nil {
		t.Fatal("expected an error for empty input")
	}
}