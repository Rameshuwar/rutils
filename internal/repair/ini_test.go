package repair

import (
	"strings"
	"testing"
)

func TestINI_ColonToEquals(t *testing.T) {
	in := []byte("key: value\n")
	out, report, err := (&iniRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "colon-to-equals") {
		t.Fatalf("expected colon-to-equals, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "key=value") {
		t.Fatalf("colon not converted: %q", out)
	}
}

func TestINI_DuplicateKeyDroppedLenient(t *testing.T) {
	in := []byte("[s]\nkey=1\nkey=2\n")
	out, report, err := (&iniRepairer{}).Repair(in, LevelLenient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "duplicate-key-dropped") {
		t.Fatalf("expected duplicate-key-dropped, got %v", report.Applied)
	}
	// Last wins — key=2 stays.
	if !strings.Contains(string(out), "key=2") {
		t.Fatalf("last value not kept: %q", out)
	}
	if strings.Count(string(out), "key=") > 2 {
		t.Fatalf("duplicate lines survived: %q", out)
	}
}

func TestINI_AlreadyClean(t *testing.T) {
	in := []byte("[section]\nkey=value\n")
	out, report, err := (&iniRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Changed {
		t.Fatalf("expected no changes, got %v", report.Applied)
	}
	if string(out) != string(in) {
		t.Fatalf("bytes changed")
	}
}