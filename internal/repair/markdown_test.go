package repair

import (
	"strings"
	"testing"
)

func TestMarkdown_HeadingSpace(t *testing.T) {
	in := []byte("#Hello\n\nworld\n")
	out, report, err := (&markdownRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "heading-space") {
		t.Fatalf("expected heading-space, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "# Hello") {
		t.Fatalf("heading not fixed: %q", out)
	}
}

func TestMarkdown_TrailingWhitespace(t *testing.T) {
	in := []byte("hello   \nworld\t\n")
	out, report, err := (&markdownRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "trailing-whitespace") {
		t.Fatalf("expected trailing-whitespace, got %v", report.Applied)
	}
	if strings.Contains(string(out), "hello   ") {
		t.Fatalf("trailing whitespace not stripped: %q", out)
	}
}

func TestMarkdown_BlankLinesCollapsed(t *testing.T) {
	in := []byte("a\n\n\n\nb\n")
	out, report, err := (&markdownRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "blank-lines-collapsed") {
		t.Fatalf("expected blank-lines-collapsed, got %v", report.Applied)
	}
	if strings.Contains(string(out), "\n\n\n") {
		t.Fatalf("blank lines not collapsed: %q", out)
	}
}

func TestMarkdown_AlreadyClean(t *testing.T) {
	in := []byte("# Title\n\nsome text\n")
	out, report, err := (&markdownRepairer{}).Repair(in, LevelNormal)
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