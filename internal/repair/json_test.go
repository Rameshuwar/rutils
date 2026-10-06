package repair

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestJSON_AlreadyValid verifies the fast path: valid JSON comes back
// unchanged, with Changed=false.
func TestJSON_AlreadyValid(t *testing.T) {
	input := []byte(`{"name":"Alice","age":30}`)
	out, report, err := (&jsonRepairer{}).Repair(input, LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Changed {
		t.Fatalf("expected Changed=false for valid input, got true (fixes: %v)", report.Applied)
	}
	if string(out) != string(input) {
		t.Fatalf("expected identical bytes, got %s", out)
	}
}

// TestJSON_TrailingComma covers the single most common real-world bug.
func TestJSON_TrailingComma(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"object", `{"a":1,}`},
		{"array", `[1,2,3,]`},
		{"nested", `{"a":[1,2,],"b":2,}`},
		{"whitespace", "{\"a\":1 , \n }"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, report, err := (&jsonRepairer{}).Repair([]byte(tc.in), LevelStrict)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !json.Valid(out) {
				t.Fatalf("output is not valid JSON: %s", out)
			}
			if !contains(report.Applied, "trailing-comma") {
				t.Fatalf("expected 'trailing-comma' fix, got %v", report.Applied)
			}
		})
	}
}

// TestJSON_Comments covers // and /* */.
func TestJSON_Comments(t *testing.T) {
	in := `{
  // this is a comment
  "a": 1, /* inline */ "b": 2
}`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if !contains(report.Applied, "comments-stripped") {
		t.Fatalf("expected comments fix, got %v", report.Applied)
	}
}

// TestJSON_CommentInsideString ensures URLs are not destroyed.
func TestJSON_CommentInsideString(t *testing.T) {
	in := `{"url":"http://example.com","x":1,}`
	out, _, err := (&jsonRepairer{}).Repair([]byte(in), LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("output invalid: %v", err)
	}
	if v["url"] != "http://example.com" {
		t.Fatalf("URL was mangled: got %v", v["url"])
	}
}

// TestJSON_SingleQuotes verifies both simple and escaped cases.
func TestJSON_SingleQuotes(t *testing.T) {
	in := `{'name':'O\'Brien'}`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if !contains(report.Applied, "single-quotes") {
		t.Fatalf("expected single-quotes fix, got %v", report.Applied)
	}
}

// TestJSON_UnquotedKeys covers JS-style object literals.
func TestJSON_UnquotedKeys(t *testing.T) {
	in := `{name:"Alice",age:30,active:true}`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if !contains(report.Applied, "unquoted-keys") {
		t.Fatalf("expected unquoted-keys fix, got %v", report.Applied)
	}
}

// TestJSON_MissingComma covers `{"a":1 "b":2}`.
func TestJSON_MissingComma(t *testing.T) {
	in := `{"a":1 "b":2}`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if !contains(report.Applied, "missing-comma") {
		t.Fatalf("expected missing-comma fix, got %v", report.Applied)
	}
}

// TestJSON_TrailingProse covers `{"a":1}garbage`.
func TestJSON_TrailingProse(t *testing.T) {
	in := `{"a":1}
Here are some notes the user pasted after the JSON.`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if !contains(report.Applied, "trailing-prose") {
		t.Fatalf("expected trailing-prose fix, got %v", report.Applied)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected a warning about discarded prose")
	}
}

// TestJSON_SmartQuotes covers copy-paste from Word/Excel.
func TestJSON_SmartQuotes(t *testing.T) {
	in := `{“a”: 1}`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelStrict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if !contains(report.Applied, "smart-quotes") {
		t.Fatalf("expected smart-quotes fix, got %v", report.Applied)
	}
}

// TestJSON_LevelStrictStopsAtLexical ensures strict refuses stage 2.
func TestJSON_LevelStrictStopsAtLexical(t *testing.T) {
	in := `{a:1}` // requires stage 2 (unquoted keys)
	_, _, err := (&jsonRepairer{}).Repair([]byte(in), LevelStrict)
	if err != ErrUnrepairable {
		t.Fatalf("expected ErrUnrepairable, got %v", err)
	}
}

// TestJSON_LevelNormalStopsBeforeCoercion ensures normal refuses stage 3.
func TestJSON_LevelNormalStopsBeforeCoercion(t *testing.T) {
	in := `{"x": NaN}`
	_, _, err := (&jsonRepairer{}).Repair([]byte(in), LevelNormal)
	if err != ErrUnrepairable {
		t.Fatalf("expected ErrUnrepairable, got %v", err)
	}
}

// TestJSON_LevelLenientCoercesNaN verifies stage 3 works.
func TestJSON_LevelLenientCoercesNaN(t *testing.T) {
	in := `{"x": NaN, "y": Infinity}`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelLenient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if v["x"] != nil || v["y"] != nil {
		t.Fatalf("expected null values, got x=%v y=%v", v["x"], v["y"])
	}
	if len(report.Warnings) < 2 {
		t.Fatalf("expected warnings for each coercion, got %v", report.Warnings)
	}
}

// TestJSON_LevelLenientBalancesBraces verifies stage 4.
func TestJSON_LevelLenientBalancesBraces(t *testing.T) {
	in := `{"a": {"b": 1}`
	out, report, err := (&jsonRepairer{}).Repair([]byte(in), LevelLenient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid(out) {
		t.Fatalf("output is not valid JSON: %s", out)
	}
	if !contains(report.Applied, "bracket-balancing") {
		t.Fatalf("expected bracket-balancing, got %v", report.Applied)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected a warning about semantic risk")
	}
}

// TestJSON_Report_HeaderValue verifies the header serialization.
func TestJSON_Report_HeaderValue(t *testing.T) {
	r := Report{
		Format:  "json",
		Level:   LevelNormal,
		Changed: true,
		Applied: []string{"trailing-comma", "single-quotes"},
		Warnings: []string{"one", "two"},
	}
	h := r.HeaderValue()
	if !strings.Contains(h, "json") ||
		!strings.Contains(h, "level=normal") ||
		!strings.Contains(h, "trailing-comma") {
		t.Fatalf("bad header: %s", h)
	}
}

// TestJSON_NeverChangesValues guarantees the invariant that values are
// never silently changed at LevelNormal.
func TestJSON_NeverChangesValues(t *testing.T) {
	in := `{"age": "thirty", "score": 42}`
	out, _, err := (&jsonRepairer{}).Repair([]byte(in), LevelNormal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if v["age"] != "thirty" {
		t.Fatalf("value was changed! got %v", v["age"])
	}
	if v["score"] != float64(42) {
		t.Fatalf("score was changed! got %v", v["score"])
	}
}

// contains is a tiny slice helper.
func contains(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}