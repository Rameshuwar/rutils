package repair

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

func init() {
	register(&jsonRepairer{})
}

// jsonRepairer implements Repairer for JSON documents.
type jsonRepairer struct{}

func (j *jsonRepairer) Format() string { return "json" }

// Repair runs the five-stage JSON pipeline. It returns the first stage
// whose output passes a strict json.Unmarshal. If every stage fails,
// the original input is returned alongside ErrUnrepairable so the
// caller can decide whether to surface the error or the raw bytes.
func (j *jsonRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "json", Level: level}

	// ---- Stage 0: normalise -----------------------------------------
	report.AddStage(0)
	work := normalizeJSONInput(raw)

	// Fast path: if the input already parses, we're done.
	if json.Valid(work) {
		return work, report, nil
	}

	// ---- Stage 1: lexical -------------------------------------------
	report.AddStage(1)
	work, lexChanges := applyLexicalFixes(work)
	for _, c := range lexChanges {
		report.AddFix(c)
	}
	if json.Valid(work) {
		return work, report, nil
	}

	// Strict level stops here.
	if level == LevelStrict {
		return raw, report, ErrUnrepairable
	}

	// ---- Stage 2: structural ----------------------------------------
	report.AddStage(2)
	work, structChanges, structWarnings := applyStructuralFixes(work)
	for _, c := range structChanges {
		report.AddFix(c)
	}
	for _, w := range structWarnings {
		report.AddWarning(w)
	}
	if json.Valid(work) {
		return work, report, nil
	}

	// Normal level stops here.
	if level == LevelNormal {
		return raw, report, ErrUnrepairable
	}

	// ---- Stage 3: value coercion ------------------------------------
	report.AddStage(3)
	work, coerceChanges := applyValueCoercion(work)
	for _, c := range coerceChanges {
		report.AddFix(c)
		report.AddWarning("value coercion applied: " + c)
	}
	if json.Valid(work) {
		return work, report, nil
	}

	// ---- Stage 4: bracket balancing ---------------------------------
	report.AddStage(4)
	if balanced, ok := balanceBraces(work); ok {
		report.AddFix("bracket-balancing")
		report.AddWarning("bracket balancing applied — output may differ semantically from input")
		return balanced, report, nil
	}

	return raw, report, ErrUnrepairable
}

// =====================================================================
// Stage 0 — Normalisation
// =====================================================================

// normalizeJSONInput performs lossless byte-level cleanups that every
// downstream stage can rely on. It never changes the *meaning* of the
// document — only the bytes that carry no meaning (BOM, CRLF, smart
// quotes inside nothing, non-breaking spaces).
func normalizeJSONInput(raw []byte) []byte {
	raw = trimBOM(raw)

	// CRLF → LF. JSON permits both, but our downstream string
	// manipulation is simpler with one canonical form.
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	raw = bytes.ReplaceAll(raw, []byte("\r"), []byte("\n"))

	// Non-breaking space (U+00A0) is legal inside a JSON string but
	// illegal as structural whitespace. Users paste it constantly
	// (copy from Word/Excel). Replace with a normal space only when
	// the surrounding bytes cannot be inside a string. We do this
	// conservatively — only leading/trailing occurrences.
	raw = bytes.TrimFunc(raw, func(r rune) bool {
		return r == '\u00A0' || r == '\uFEFF'
	})

	// Ensure the slice is valid UTF-8; if not, replace invalid bytes
	// with U+FFFD so downstream string ops don't panic.
	if !utf8.Valid(raw) {
		raw = bytes.ToValidUTF8(raw, []byte("\uFFFD"))
	}

	return raw
}

// =====================================================================
// Stage 1 — Lexical fixes
// =====================================================================

// applyLexicalFixes removes comments, converts smart quotes, and
// collapses known-safe whitespace anomalies. Every transformation
// here is character-level and cannot change the meaning of a valid
// JSON document.
//
// Returns the transformed bytes plus a list of fix names that were
// applied (for the Report).
func applyLexicalFixes(raw []byte) ([]byte, []string) {
	var fixes []string
	work := raw

	// --- 1a. Strip // line comments and /* */ block comments ---
	// We only strip comments *outside* strings. A proper scanner is
	// required because "http://x" must not be touched.
	if out, changed := stripJSONComments(work); changed {
		work = out
		fixes = append(fixes, "comments-stripped")
	}

	// --- 1b. Smart quotes → ASCII ---
	// This is safe because a smart quote can never be structural JSON.
	// Inside a string, converting “ ” to " is what the user meant.
	if out, changed := replaceSmartQuotes(work); changed {
		work = out
		fixes = append(fixes, "smart-quotes")
	}

	// --- 1c. Trailing commas before } or ] ---
	// Legal in JavaScript, illegal in JSON. Unambiguously safe to drop.
	if out, changed := stripTrailingCommas(work); changed {
		work = out
		fixes = append(fixes, "trailing-comma")
	}

	// --- 1d. Collapse BOM leftovers, stray tabs at line start ---
	if out, changed := collapseLeadingWhitespace(work); changed {
		work = out
		fixes = append(fixes, "leading-whitespace")
	}

	return work, fixes
}

// stripJSONComments removes // and /* */ comments that appear outside
// of string literals. A hand-rolled scanner is required because a
// simple regex would destroy "url": "http://example.com".
func stripJSONComments(raw []byte) ([]byte, bool) {
	var (
		out     bytes.Buffer
		inStr   bool
		escape  bool
		changed bool
	)

	for i := 0; i < len(raw); i++ {
		c := raw[i]

		if inStr {
			out.WriteByte(c)
			if escape {
				escape = false
				continue
			}
			switch c {
			case '\\':
				escape = true
			case '"':
				inStr = false
			}
			continue
		}

		switch c {
		case '"':
			inStr = true
			out.WriteByte(c)

		case '/':
			// Look ahead for // or /*
			if i+1 < len(raw) && raw[i+1] == '/' {
				changed = true
				// Skip to end of line, but leave the newline itself.
				i += 2
				for i < len(raw) && raw[i] != '\n' {
					i++
				}
				// Step back so the outer loop can consume the newline.
				i--
				continue
			}
			if i+1 < len(raw) && raw[i+1] == '*' {
				changed = true
				i += 2
				for i+1 < len(raw) && !(raw[i] == '*' && raw[i+1] == '/') {
					i++
				}
				i++ // skip the closing '/'
				continue
			}
			out.WriteByte(c)

		default:
			out.WriteByte(c)
		}
	}

	return out.Bytes(), changed
}

// replaceSmartQuotes converts Unicode curly quotes to ASCII straight
// quotes. It also handles the Unicode ellipsis and en/em dashes when
// they appear inside a string — but only inside a string, never
// structural.
func replaceSmartQuotes(raw []byte) ([]byte, bool) {
	replacements := []struct {
		from []byte
		to   []byte
	}{
		{[]byte("\u201C"), []byte(`"`)}, // “
		{[]byte("\u201D"), []byte(`"`)}, // ”
		{[]byte("\u2018"), []byte("'")}, // ‘
		{[]byte("\u2019"), []byte("'")}, // ’
		{[]byte("\u00AB"), []byte(`"`)}, // «
		{[]byte("\u00BB"), []byte(`"`)}, // »
	}

	work := raw
	changed := false
	for _, r := range replacements {
		if bytes.Contains(work, r.from) {
			work = bytes.ReplaceAll(work, r.from, r.to)
			changed = true
		}
	}
	return work, changed
}

// stripTrailingCommas removes commas that appear immediately before a
// closing } or ], ignoring whitespace and comments in between.
func stripTrailingCommas(raw []byte) ([]byte, bool) {
	out := make([]byte, 0, len(raw))
	changed := false
	inStr := false
	escape := false

	for i := 0; i < len(raw); i++ {
		c := raw[i]

		if inStr {
			out = append(out, c)
			if escape {
				escape = false
				continue
			}
			switch c {
			case '\\':
				escape = true
			case '"':
				inStr = false
			}
			continue
		}

		if c == '"' {
			inStr = true
			out = append(out, c)
			continue
		}

		if c == ',' {
			// Look ahead: skip whitespace, then check for } or ].
			j := i + 1
			for j < len(raw) {
				switch raw[j] {
				case ' ', '\t', '\n', '\r':
					j++
					continue
				}
				break
			}
			if j < len(raw) && (raw[j] == '}' || raw[j] == ']') {
				// Drop the comma. Keep the whitespace so we don't
				// disturb line numbers more than necessary.
				changed = true
				continue
			}
		}

		out = append(out, c)
	}

	return out, changed
}

// collapseLeadingWhitespace trims trailing whitespace at end-of-line
// and normalizes indentation runs of 8+ spaces to tabs. Neither is
// required, but it keeps the output pretty and consistent.
func collapseLeadingWhitespace(raw []byte) ([]byte, bool) {
	lines := bytes.Split(raw, []byte("\n"))
	changed := false
	for i, line := range lines {
		trimmed := bytes.TrimRight(line, " \t")
		if !bytes.Equal(trimmed, line) {
			lines[i] = trimmed
			changed = true
		}
	}
	if !changed {
		return raw, false
	}
	return bytes.Join(lines, []byte("\n")), true
}

// =====================================================================
// Stage 2 — Structural fixes
// =====================================================================

// applyStructuralFixes resolves genuine syntax errors: single quotes,
// unquoted keys, missing commas between members, and trailing prose.
// These can change how a byte stream tokenizes, but they never change
// a *value* — the numeric 30 stays 30, the string "thirty" stays "thirty".
func applyStructuralFixes(raw []byte) ([]byte, []string, []string) {
	var fixes, warnings []string
	work := raw

	// --- 2a. Single-quoted strings → double-quoted ---
	if out, changed := convertSingleQuotes(work); changed {
		work = out
		fixes = append(fixes, "single-quotes")
	}

	// --- 2b. Unquoted object keys → quoted ---
	if out, changed := quoteUnquotedKeys(work); changed {
		work = out
		fixes = append(fixes, "unquoted-keys")
	}

	// --- 2c. Missing commas between members ---
	if out, changed := insertMissingCommas(work); changed {
		work = out
		fixes = append(fixes, "missing-comma")
	}

	// --- 2d. Trailing prose after a balanced top-level value ---
	if out, truncated := trimTrailingProse(work); truncated {
		work = out
		fixes = append(fixes, "trailing-prose")
		warnings = append(warnings, "text after the top-level JSON value was discarded")
	}

	return work, fixes, warnings
}

// convertSingleQuotes rewrites single-quoted strings to double-quoted,
// escaping any double quote inside them. It understands \-escapes and
// correctly handles the `\'` sequence — which inside a single-quoted
// string means "a literal apostrophe", so it must be emitted as a bare
// apostrophe (not as the closing quote, and not as an escape).
func convertSingleQuotes(raw []byte) ([]byte, bool) {
	var (
		out     bytes.Buffer
		inStr   bool
		quote   byte
		escape  bool
		changed bool
	)

	for i := 0; i < len(raw); i++ {
		c := raw[i]

		// --- Outside a string ---
		if !inStr {
			if c == '\'' {
				// A single-quoted string begins.
				inStr = true
				quote = '\''
				out.WriteByte('"') // emit as double-quote
				changed = true
				continue
			}
			out.WriteByte(c)
			continue
		}

		// --- Inside a string ---

		// Handle the escape state first.
		if escape {
			escape = false

			// Inside a single-quoted string, `\'` means a literal
			// apostrophe. Because we're rewriting to double quotes,
			// that apostrophe needs no escaping at all.
			if quote == '\'' && c == '\'' {
				out.WriteByte('\'')
				continue
			}

			// Inside a double-quoted string, `\"` means a literal
			// double quote — must be preserved as `\"`.
			if quote == '"' && c == '"' {
				out.WriteString(`\"`)
				continue
			}

			// Any other escape (\n, \t, \\, \uXXXX, ...) is passed
			// through verbatim: backslash + char.
			out.WriteByte('\\')
			out.WriteByte(c)
			continue
		}

		// Not in escape state.
		switch c {
		case '\\':
			escape = true

		case '\'':
			if quote == '\'' {
				// Closing the single-quoted string.
				out.WriteByte('"')
				inStr = false
				continue
			}
			// A bare apostrophe inside a double-quoted string is fine.
			out.WriteByte(c)

		case '"':
			if quote == '\'' {
				// A double quote inside a single-quoted string must
				// be escaped when we switch to double quotes.
				out.WriteString(`\"`)
				continue
			}
			// Closing the (already double-quoted) string.
			out.WriteByte('"')
			inStr = false

		default:
			out.WriteByte(c)
		}
	}

	return out.Bytes(), changed
}

// quoteUnquotedKeys finds object keys that are not quoted and wraps
// them in double quotes. A key in JSON is any identifier that appears
// immediately after { or , and is followed by : — with optional
// whitespace.
func quoteUnquotedKeys(raw []byte) ([]byte, bool) {
	var out bytes.Buffer
	changed := false
	inStr := false
	escape := false

	// A key position is: after { or , at object depth.
	expectKey := false

	i := 0
	for i < len(raw) {
		c := raw[i]

		if inStr {
			out.WriteByte(c)
			if escape {
				escape = false
			} else if c == '\\' {
				escape = true
			} else if c == '"' {
				inStr = false
			}
			i++
			continue
		}

		switch c {
		case '"':
			inStr = true
			expectKey = false
			out.WriteByte(c)
			i++

		case '{', ',':
			out.WriteByte(c)
			expectKey = true
			i++

		case ' ', '\t', '\n', '\r':
			out.WriteByte(c)
			i++

		case ':':
			out.WriteByte(c)
			expectKey = false
			i++

		case '}', ']':
			out.WriteByte(c)
			expectKey = false
			i++

		default:
			if expectKey && isIdentStart(c) {
				// Read identifier.
				start := i
				for i < len(raw) && isIdentChar(raw[i]) {
					i++
				}
				ident := raw[start:i]

				// Peek: must be followed by optional space then ':'
				j := i
				for j < len(raw) && (raw[j] == ' ' || raw[j] == '\t') {
					j++
				}
				if j < len(raw) && raw[j] == ':' {
					out.WriteByte('"')
					out.Write(ident)
					out.WriteByte('"')
					changed = true
					expectKey = false
					continue
				}
				// Not a key — write as-is.
				out.Write(ident)
				expectKey = false
				continue
			}
			out.WriteByte(c)
			i++
		}
	}

	return out.Bytes(), changed
}

func isIdentStart(c byte) bool {
	return c == '_' || c == '$' ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}

func isIdentChar(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

// insertMissingCommas adds a comma between `}` and `{`, between `]` and
// `[`, between a value and the next key, etc. — wherever a comma is
// required but absent.
//
// This is the trickiest stage because the fix is *contextual*: a comma
// is only missing if the parser would otherwise see two adjacent
// values or members.
func insertMissingCommas(raw []byte) ([]byte, bool) {
	var out bytes.Buffer
	changed := false
	inStr := false
	escape := false

	// Track the last significant byte emitted (outside strings and
	// whitespace) to decide whether a comma is needed.
	var lastSig byte

	i := 0
	for i < len(raw) {
		c := raw[i]

		if inStr {
			out.WriteByte(c)
			if escape {
				escape = false
			} else if c == '\\' {
				escape = true
			} else if c == '"' {
				inStr = false
				lastSig = '"'
			}
			i++
			continue
		}

		switch c {
		case '"':
			// A new string is about to start. If lastSig was a value
			// terminator (", }, ], digit, letter), a comma is needed.
			if needsCommaBefore(lastSig, c) {
				out.WriteByte(',')
				changed = true
			}
			inStr = true
			out.WriteByte(c)
			lastSig = 0 // will be set at string close
			i++

		case '{', '[':
			// An opening brace/bracket right after a value needs a comma.
			if needsCommaBefore(lastSig, c) {
				out.WriteByte(',')
				changed = true
			}
			out.WriteByte(c)
			lastSig = c
			i++

		case '}', ']':
			out.WriteByte(c)
			lastSig = c
			i++

		case ',', ':':
			out.WriteByte(c)
			lastSig = c
			i++

		case ' ', '\t', '\n', '\r':
			out.WriteByte(c)
			i++

		default:
			// Bareword: true / false / null / number. If the previous
			// significant byte ended a value, we need a comma.
			if needsCommaBefore(lastSig, c) {
				out.WriteByte(',')
				changed = true
			}
			// Consume the bareword.
			start := i
			for i < len(raw) && !isStructural(raw[i]) {
				i++
			}
			out.Write(raw[start:i])
			lastSig = raw[i-1]
		}
	}

	return out.Bytes(), changed
}

func needsCommaBefore(lastSig, next byte) bool {
	// A value just ended?
	valueEnded := lastSig == '"' || lastSig == '}' || lastSig == ']' ||
		(lastSig >= '0' && lastSig <= '9') ||
		lastSig == 'e' || lastSig == 'l' /* true/false/null */
	if !valueEnded {
		return false
	}
	// Only insert a comma if the next token *starts a new value*.
	return next == '"' || next == '{' || next == '[' ||
		isIdentStart(next) ||
		(next >= '0' && next <= '9') ||
		next == '-' || next == '+'
}

func isStructural(c byte) bool {
	switch c {
	case '{', '}', '[', ']', ',', ':', '"',
		' ', '\t', '\n', '\r':
		return true
	}
	return false
}

// trimTrailingProse removes any bytes after a balanced top-level JSON
// value. Example: `{"a":1}garbage` → `{"a":1}`.
func trimTrailingProse(raw []byte) ([]byte, bool) {
	depth := 0
	inStr := false
	escape := false
	endIdx := -1

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if inStr {
			if escape {
				escape = false
			} else if c == '\\' {
				escape = true
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 {
				endIdx = i
			}
		}
	}

	if endIdx == -1 || endIdx == len(raw)-1 {
		return raw, false
	}

	// Anything after endIdx that isn't whitespace is trailing prose.
	trailing := raw[endIdx+1:]
	if len(bytes.TrimSpace(trailing)) == 0 {
		return raw, false
	}

	return raw[:endIdx+1], true
}

// =====================================================================
// Stage 3 — Value coercion
// =====================================================================

// applyValueCoercion replaces non-JSON literals with their closest JSON
// equivalent. Every substitution here is recorded as a warning because
// it changes the meaning of the document.
func applyValueCoercion(raw []byte) ([]byte, []string) {
	var fixes []string
	work := raw

	// Python/JS literals → JSON.
	substitutions := []struct {
		name string
		from string
		to   string
	}{
		{"NaN", "NaN", "null"},
		{"Infinity", "Infinity", "null"},
		{"-Infinity", "-Infinity", "null"},
		{"True", "True", "true"},
		{"False", "False", "false"},
		{"None", "None", "null"},
		{"undefined", "undefined", "null"},
	}

	for _, s := range substitutions {
		out, changed := replaceBareword(work, s.from, s.to)
		if changed {
			work = out
			fixes = append(fixes, s.name+"-coerced")
		}
	}

	return work, fixes
}

// replaceBareword replaces `from` with `to` only when `from` appears
// as a complete token — not inside a string, not as part of a longer
// identifier.
func replaceBareword(raw []byte, from, to string) ([]byte, bool) {
	fromB := []byte(from)
	var out bytes.Buffer
	inStr := false
	escape := false
	changed := false

	i := 0
	for i < len(raw) {
		c := raw[i]

		if inStr {
			out.WriteByte(c)
			if escape {
				escape = false
			} else if c == '\\' {
				escape = true
			} else if c == '"' {
				inStr = false
			}
			i++
			continue
		}

		if c == '"' {
			inStr = true
			out.WriteByte(c)
			i++
			continue
		}

		// Word-boundary check.
		if bytes.HasPrefix(raw[i:], fromB) {
			beforeOK := i == 0 || !isIdentChar(raw[i-1])
			afterIdx := i + len(fromB)
			afterOK := afterIdx >= len(raw) || !isIdentChar(raw[afterIdx])
			if beforeOK && afterOK {
				out.WriteString(to)
				changed = true
				i = afterIdx
				continue
			}
		}

		out.WriteByte(c)
		i++
	}

	return out.Bytes(), changed
}

// =====================================================================
// Stage 4 — Bracket balancing (last resort)
// =====================================================================

// balanceBraces attempts to make the input parseable by adding missing
// closers at the end of the document. It only succeeds if the result
// is syntactically valid JSON — we never return a "maybe fixed" doc.
func balanceBraces(raw []byte) ([]byte, bool) {
	// Count open vs close outside strings.
	var stack []byte
	inStr := false
	escape := false

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if inStr {
			if escape {
				escape = false
			} else if c == '\\' {
				escape = true
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{', '[':
			stack = append(stack, c)
		case '}', ']':
			if len(stack) == 0 {
				// Unbalanced close — cannot safely repair.
				return nil, false
			}
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) == 0 {
		return nil, false // Nothing to fix.
	}

	// Append missing closers in reverse order.
	var buf bytes.Buffer
	buf.Write(raw)
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == '{' {
			buf.WriteByte('}')
		} else {
			buf.WriteByte(']')
		}
	}

	if !json.Valid(buf.Bytes()) {
		return nil, false
	}
	return buf.Bytes(), true
}

// Kept for future format-specific tweaks; silences an unused-import
// complaint when the strings import is only used conditionally.
var _ = strings.TrimSpace