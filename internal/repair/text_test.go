package repair

import "testing"

func TestStripBOM(t *testing.T) {
	withBOM := []byte{0xEF, 0xBB, 0xBF, 'h', 'i'}
	out, had := stripBOM(withBOM)
	if !had {
		t.Fatal("expected hadBOM=true")
	}
	if string(out) != "hi" {
		t.Fatalf("expected 'hi', got %q", out)
	}

	without := []byte("hi")
	out, had = stripBOM(without)
	if had {
		t.Fatal("expected hadBOM=false for input without BOM")
	}
	if string(out) != "hi" {
		t.Fatalf("expected 'hi', got %q", out)
	}
}

func TestNormaliseLineEndings(t *testing.T) {
	in := []byte("a\r\nb\rc\nd")
	out := normaliseLineEndings(in)
	if string(out) != "a\nb\nc\nd" {
		t.Fatalf("got %q", out)
	}
}

func TestSplitLinesDropsTrailingBlanks(t *testing.T) {
	in := []byte("a\nb\n\n\n")
	lines := splitLines(in)
	if len(lines) != 2 || lines[0] != "a" || lines[1] != "b" {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}

func TestCountDelimiterIgnoresQuoted(t *testing.T) {
	line := `a,"b,c,d",e`
	if n := countDelimiter(line, ','); n != 2 {
		t.Fatalf("expected 2 unquoted commas, got %d", n)
	}
}

func TestDetectDominantDelimiter(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  byte
		ok    bool
	}{
		{"comma", []string{"a,b,c", "1,2,3", "4,5,6"}, ',', true},
		{"semicolon", []string{"a;b;c", "1;2;3", "4;5;6"}, ';', true},
		{"tab", []string{"a\tb\tc", "1\t2\t3", "4\t5\t6"}, '\t', true},
		{"nothing", []string{"hello", "world", "no delims here"}, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := detectDominantDelimiter(tc.lines, 10)
			if ok != tc.ok {
				t.Fatalf("ok mismatch: got %v want %v", ok, tc.ok)
			}
			if ok && got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestLeadingSpaces(t *testing.T) {
	if leadingSpaces("   hello") != 3 {
		t.Fatal("expected 3")
	}
	if leadingSpaces("\thello") != 0 {
		t.Fatal("tabs must not be counted as spaces")
	}
}