package repair

import (
	"strings"
	"testing"
)

func TestXML_AlreadyValid(t *testing.T) {
	in := []byte(`<root><a>1</a></root>`)
	out, report, err := (&xmlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Changed {
		t.Fatalf("expected no change; fixes=%v", report.Applied)
	}
	if string(out) != string(in) {
		t.Fatalf("bytes changed: %q", out)
	}
}

func TestXML_CaseMismatch(t *testing.T) {
	in := []byte(`<root><A>1</a></root>`)
	out, report, err := (&xmlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isValidXML(out) {
		t.Fatalf("output invalid: %s", out)
	}
	if !contains(report.Applied, "tag-case") {
		t.Fatalf("expected tag-case, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "</A>") {
		t.Fatalf("close tag not recased: %s", out)
	}
}

func TestXML_BareAmpersand(t *testing.T) {
	in := []byte(`<note>Tom & Jerry</note>`)
	out, report, err := (&xmlRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "bare-ampersand") {
		t.Fatalf("expected bare-ampersand, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "Tom &amp; Jerry") {
		t.Fatalf("ampersand not escaped: %s", out)
	}
}

func TestXML_ExistingEntityPreserved(t *testing.T) {
	in := []byte(`<a>Tom &amp; Jerry &lt;3</a>`)
	out, _, err := (&xmlRepairer{}).Repair(in, LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Count(string(out), "&amp;") != 1 {
		t.Fatalf("existing entity damaged: %s", out)
	}
	if strings.Count(string(out), "&lt;") != 1 {
		t.Fatalf("existing entity damaged: %s", out)
	}
}

func TestXML_UnclosedTag(t *testing.T) {
	in := []byte(`<root><a>1`)
	out, report, err := (&xmlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "unclosed-tags") {
		t.Fatalf("expected unclosed-tags, got %v", report.Applied)
	}
	if !isValidXML(out) {
		t.Fatalf("output still invalid: %s", out)
	}
}

func TestXML_MultipleRoots(t *testing.T) {
	in := []byte(`<a>1</a><b>2</b>`)
	out, report, err := (&xmlRepairer{}).Repair(in, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(report.Applied, "multiple-roots") {
		t.Fatalf("expected multiple-roots, got %v", report.Applied)
	}
	if !strings.Contains(string(out), "<root>") {
		t.Fatalf("root wrapper missing: %s", out)
	}
	if !isValidXML(out) {
		t.Fatalf("output still invalid: %s", out)
	}
}