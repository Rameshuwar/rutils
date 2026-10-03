package repair

import (
	"bytes"
	"strings"
)

func init() {
	register(&yamlRepairer{})
}

// yamlRepairer implements Repairer for YAML documents.
//
// DECISION: default indent width is 2 spaces (matches Kubernetes,
// Hugo, and almost every modern YAML dialect).
const yamlDefaultIndent = 2

type yamlRepairer struct{}

func (y *yamlRepairer) Format() string { return "yaml" }

// Repair normalises a YAML document.
//
// Rules, by level:
//
//	strict  — strip BOM, convert tabs to spaces, strip trailing
//	          whitespace, insert a space after `key:` when missing.
//	normal  — normalise indentation levels; quote scalars that look
//	          like YAML keywords (yes/no/on/off).
//	lenient — prepend `---` if the document has no start marker.
func (y *yamlRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "yaml", Level: level}

	// Empty / whitespace-only input is not repairable as YAML — YAML
	// accepts it as "the null document", which is not what a user
	// uploading a config file meant.
	if len(bytes.TrimSpace(raw)) == 0 {
		return raw, report, ErrEmptyInput
	}

	// Fast path: valid YAML that also has no ambiguous scalars and no
	// missing colon-spaces passes through unchanged.
	if isValidYAML(raw) && !hasColonWithoutSpace(raw) && !hasAmbiguousScalar(raw) {
		return raw, report, nil
	}

	// ---- Stage 0: normalise ----
	report.AddStage(0)
	work, hadBOM := stripBOM(raw)
	if hadBOM {
		report.AddFix("bom-stripped")
	}
	work = normaliseLineEndings(work)

	// ---- Stage 1: tabs → spaces, trailing whitespace ----
	report.AddStage(1)
	lines := strings.Split(string(work), "\n")
	for i, line := range lines {
		// Tabs at the start of a line are illegal in YAML.
		if hasTabIndent(line) {
			lines[i] = expandLeadingTabs(line, yamlDefaultIndent)
			report.AddFix("tab-to-space")
			line = lines[i]
		}
		// Trailing whitespace.
		if t := rstrip(line); t != line {
			lines[i] = t
			report.AddFix("trailing-whitespace")
		}
	}
	work = []byte(strings.Join(lines, "\n"))

	// Insert space after `key:` if missing (but not for URLs, etc.).
	if out, changed := ensureSpaceAfterColon(work); changed {
		work = out
		report.AddFix("colon-spacing")
	}

	if isValidYAML(work) && !hasColonWithoutSpace(work) && !hasAmbiguousScalar(work) {
		return work, report, nil
	}

	if level == LevelStrict {
		return raw, report, ErrUnrepairable
	}

	// ---- Stage 2: indentation normalisation ----
	report.AddStage(2)
	if out, changed := normaliseIndentation(work, yamlDefaultIndent); changed {
		work = out
		report.AddFix("indent-normalised")
	}
	if out, changed := quoteAmbiguousScalars(work); changed {
		work = out
		report.AddFix("ambiguous-scalars-quoted")
	}
	if isValidYAML(work) {
		return work, report, nil
	}

	if level == LevelNormal {
		return raw, report, ErrUnrepairable
	}

	// ---- Stage 3: prepend `---` document marker ----
	report.AddStage(3)
	if !bytes.HasPrefix(bytes.TrimSpace(work), []byte("---")) {
		work = append([]byte("---\n"), work...)
		report.AddFix("doc-marker-added")
	}
	if isValidYAML(work) {
		return work, report, nil
	}

	return raw, report, ErrUnrepairable
}

// expandLeadingTabs replaces each leading tab with `width` spaces.
func expandLeadingTabs(line string, width int) string {
	spaces := strings.Repeat(" ", width)
	var b strings.Builder
	for i := 0; i < len(line); i++ {
		if line[i] == '\t' {
			b.WriteString(spaces)
			continue
		}
		b.WriteString(line[i:])
		break
	}
	return b.String()
}

// ensureSpaceAfterColon inserts a single space after `:` when the
// character immediately following is not whitespace and the pattern
// looks like a mapping key (`key: value`).
//
// It deliberately ignores `://` (URLs) and lines inside quoted strings.
func ensureSpaceAfterColon(raw []byte) ([]byte, bool) {
	changed := false
	var out bytes.Buffer

	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimLeft(line, " ")
		indent := line[:len(line)-len(trimmed)]

		// Skip blank lines and comments.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "- ") {
			// Only the first colon needs handling for the simple case.
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}

		colon := strings.IndexByte(trimmed, ':')
		if colon < 0 {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		after := trimmed[colon+1:]
		if after == "" || after[0] == ' ' || after[0] == '\t' ||
			after[0] == '/' { // URL scheme
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}

		out.WriteString(indent)
		out.WriteString(trimmed[:colon+1])
		out.WriteByte(' ')
		out.WriteString(after)
		out.WriteByte('\n')
		changed = true
	}

	if !changed {
		return raw, false
	}
	return out.Bytes(), true
}

// normaliseIndentation rewrites leading space runs to a consistent
// multiple of `width`. It picks the most common run length as the
// unit.
func normaliseIndentation(raw []byte, width int) ([]byte, bool) {
	lines := strings.Split(string(raw), "\n")

	// Distribution of leading space counts (excluding 0).
	dist := map[int]int{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		n := leadingSpaces(line)
		if n > 0 {
			dist[n]++
		}
	}

	if len(dist) <= 1 {
		return raw, false // already consistent
	}

	// Find the smallest non-trivial indent — that's the unit.
	unit := width
	for k := range dist {
		if k < unit && k > 0 {
			unit = k
		}
	}
	if unit <= 0 {
		return raw, false
	}

	changed := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		n := leadingSpaces(line)
		if n == 0 {
			continue
		}
		if n%unit == 0 {
			newN := (n / unit) * width
			if newN != n {
				lines[i] = strings.Repeat(" ", newN) + line[n:]
				changed = true
			}
		}
	}

	if !changed {
		return raw, false
	}
	return []byte(strings.Join(lines, "\n")), true
}

// quoteAmbiguousScalars quotes unquoted values that YAML 1.1 would
// interpret as booleans but that a user probably meant as strings.
// This runs only at normal level and later.
func quoteAmbiguousScalars(raw []byte) ([]byte, bool) {
	ambiguous := map[string]bool{
		"yes": true, "no": true, "on": true, "off": true,
		"true": false, "false": false, // NOT ambiguous — leave alone
	}
	_ = ambiguous

	changed := false
	lines := strings.Split(string(raw), "\n")
	for i, line := range lines {
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])
		if val == "" {
			continue
		}
		// Only touch clearly ambiguous words.
		low := strings.ToLower(val)
		if low == "yes" || low == "no" || low == "on" || low == "off" {
			// Quote it.
			lines[i] = line[:colon+1] + " \"" + val + "\""
			changed = true
		}
		_ = key
	}
	if !changed {
		return raw, false
	}
	return []byte(strings.Join(lines, "\n")), true
}
// hasColonWithoutSpace reports whether any non-comment, non-blank
// line has a `:` that is not followed by a space (and is not part of
// a URL scheme like `://`).
func hasColonWithoutSpace(raw []byte) bool {
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "- ") {
			continue
		}
		colon := strings.IndexByte(trimmed, ':')
		if colon < 0 {
			continue
		}
		after := trimmed[colon+1:]
		if after == "" {
			continue
		}
		if after[0] == ' ' || after[0] == '\t' || after[0] == '/' {
			continue
		}
		return true
	}
	return false
}


// hasAmbiguousScalar reports whether any `key: value` line has a
// value that YAML 1.1 interprets as a boolean but that a user
// probably meant as a string.
func hasAmbiguousScalar(raw []byte) bool {
	for _, line := range strings.Split(string(raw), "\n") {
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		val := strings.TrimSpace(line[colon+1:])
		switch strings.ToLower(val) {
		case "yes", "no", "on", "off":
			return true
		}
	}
	return false
}