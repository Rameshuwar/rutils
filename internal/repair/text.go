package repair

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ============================================================
// text.go — shared helpers for the text-format repairers
//
// These helpers are deliberately tiny and dependency-free. Every
// repairer that deals with line-oriented text uses them so that
// BOM handling, CRLF normalisation, and delimiter detection behave
// identically across formats.
// ============================================================

// stripBOM removes a UTF-8 BOM (EF BB BF) if present and reports
// whether one was there. Mirrors trimBOM in repair.go but exposes
// the boolean so repairers can record the fix in their Report.
func stripBOM(raw []byte) ([]byte, bool) {
	if len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF {
		return raw[3:], true
	}
	return raw, false
}

// normaliseLineEndings converts CRLF and lone CR to LF. Callers get
// a slice that can be safely split on "\n" afterwards.
func normaliseLineEndings(raw []byte) []byte {
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	raw = bytes.ReplaceAll(raw, []byte("\r"), []byte("\n"))
	return raw
}

// splitLines returns the lines of a text document with any trailing
// blank lines dropped. Empty input returns an empty slice.
func splitLines(raw []byte) []string {
	raw = normaliseLineEndings(raw)
	lines := strings.Split(string(raw), "\n")

	// Drop trailing empty lines (very common from Windows exports).
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// joinLines re-joins a slice of lines into a single byte slice,
// always terminating with a single newline. Consistent trailing
// newline keeps downstream diffing and hashing stable.
func joinLines(lines []string) []byte {
	if len(lines) == 0 {
		return []byte{}
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

// rstrip trims ASCII spaces and tabs from the right end of s.
// Faster than strings.TrimRightFunc because it doesn't allocate a
// closure per call — relevant when we call it on every line of a
// 100k-line CSV.
func rstrip(s string) string {
	i := len(s)
	for i > 0 {
		c := s[i-1]
		if c != ' ' && c != '\t' {
			break
		}
		i--
	}
	return s[:i]
}

// countDelimiter returns the number of times delim appears on a
// single line, ignoring occurrences inside double-quoted regions.
// Used by CSV repairer to pick the dominant delimiter.
func countDelimiter(line string, delim byte) int {
	n := 0
	inQuote := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			// Handle "" escape inside quoted fields.
			if inQuote && i+1 < len(line) && line[i+1] == '"' {
				i++
				continue
			}
			inQuote = !inQuote
			continue
		}
		if !inQuote && c == delim {
			n++
		}
	}
	return n
}

// detectDominantDelimiter examines the first maxLines non-empty lines
// of a document and returns the delimiter that appears with the same
// count on the most lines. Ties are broken by preference order
// (comma > semicolon > tab > pipe), matching common exports.
//
// Returns (0, false) if no delimiter appears consistently — the
// caller should then reject the input as "not CSV-like".
func detectDominantDelimiter(lines []string, maxLines int) (byte, bool) {
	type stat struct {
		count        int // how many lines had this exact count
		countValue   int // what the count was
	}

	// Preferred order — used only to break ties deterministically.
	preferred := []byte{',', ';', '\t', '|'}

	// For each candidate delimiter, record the count-per-line
	// distribution.
	results := map[byte]map[int]int{}
	for _, d := range preferred {
		results[d] = map[int]int{}
	}

	considered := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		for _, d := range preferred {
			results[d][countDelimiter(line, d)]++
		}
		considered++
		if considered >= maxLines {
			break
		}
	}

	if considered == 0 {
		return 0, false
	}

	// Pick the (delimiter, count) pair that appeared on the most lines.
	// The count must be > 0 — a delimiter that appears zero times on
	// every line is not "the" delimiter.
	var (
		bestDelim byte
		bestCount int
		bestFreq  int
	)
	for _, d := range preferred {
		for cnt, freq := range results[d] {
			if cnt == 0 {
				continue
			}
			if freq > bestFreq || (freq == bestFreq && cnt > bestCount) {
				bestDelim = d
				bestCount = cnt
				bestFreq = freq
			}
		}
	}

	// Require at least 2 lines to agree, otherwise we're guessing.
	if bestFreq < 2 && considered >= 2 {
		return 0, false
	}
	if bestFreq == 0 {
		return 0, false
	}
	return bestDelim, true
}

// leadingSpaces counts the number of leading space characters in s
// (tabs are NOT counted — the YAML repairer handles those separately).
func leadingSpaces(s string) int {
	n := 0
	for n < len(s) && s[n] == ' ' {
		n++
	}
	return n
}

// hasTabIndent reports whether a line begins with one or more tabs.
func hasTabIndent(s string) bool {
	return len(s) > 0 && s[0] == '\t'
}


func isValidXML(b []byte) bool {
	dec := xml.NewDecoder(bytes.NewReader(b))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
	}
}

// isValidYAML reports whether b parses as YAML. Same rationale as
// isValidXML — gopkg.in/yaml.v3 offers no Valid() helper.
func isValidYAML(b []byte) bool {
	var v any
	return yaml.Unmarshal(b, &v) == nil
}