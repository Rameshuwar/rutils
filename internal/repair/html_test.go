package repair

import (
	"strings"
	"testing"
)

func TestHTML_UnclosedTagClosed(t *testing.T) {
	in := []byte("<p>hello")
	out, report, err := (&htmlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "html-normalised") {
		t.Fatalf("expected html-normalised, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "</p>") {
		t.Fatalf("tag not closed: %q", out)
	}
}

func TestHTML_UnquotedAttributeQuoted(t *testing.T) {
	in := []byte(`<a href=foo>x</a>`)
	out, _, err := (&htmlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), `href="foo"`) {
		t.Fatalf("attribute not quoted: %q", out)
	}
}

func TestHTML_UppercaseTagsLowercased(t *testing.T) {
	in := []byte(`<DIV>text</DIV>`)
	out, _, err := (&htmlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "<div>") {
		t.Fatalf("tag not lowercased: %q", out)
	}
}

func TestHTML_StrictPassesThrough(t *testing.T) {
	in := []byte(`<p>hello</p>`)
	_, report, err := (&htmlRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Changed {
		t.Fatalf("strict should not modify well-formed HTML")
	}
}

func TestHTML_DoctypeInjectedAtLenient(t *testing.T) {
	in := []byte(`<html><body>hi</body></html>`)
	out, report, err := (&htmlRepairer{}).Repair(in, LevelLenient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "doctype-injected") {
		t.Fatalf("expected doctype-injected, got %v", report.Applied)
	}
	if !strings.HasPrefix(strings.ToLower(string(out)), "<!doctype html>") {
		t.Fatalf("doctype missing: %q", out[:60])
	}
}

func TestHTML_EmptyInput(t *testing.T) {
	_, _, err := (&htmlRepairer{}).Repair([]byte("   \n"), LevelNormal)
	if err != ErrEmptyInput {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
}