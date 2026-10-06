package repair

import (
	"bytes"
	"strings"
)

func init() {
	register(&iniRepairer{})
}

// iniRepairer implements Repairer for INI-style config documents.
type iniRepairer struct{}

func (i *iniRepairer) Format() string { return "ini" }

// Repair normalises an INI document.
//
// Rules:
//
//	strict  — strip BOM, normalise line endings, strip trailing
//	          whitespace, normalise `key : value` to `key=value`.
//	normal  — quote values with leading/trailing spaces.
//	lenient — drop duplicate keys within a section, keeping the last.
func (i *iniRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "ini", Level: level}

	report.AddStage(0)
	work, hadBOM := stripBOM(raw)
	if hadBOM {
		report.AddFix("bom-stripped")
	}
	work = normaliseLineEndings(work)

	lines := strings.Split(string(work), "\n")

	// ---- Stage 1: line-level cleanups ----
	report.AddStage(1)
	for idx, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		indent := line[:len(line)-len(trimmed)]

		// Skip blank lines, comments, section headers.
		if trimmed == "" || strings.HasPrefix(trimmed, ";") ||
			strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "[") {
			lines[idx] = rstrip(indent + trimmed)
			continue
		}

		// Normalise `key : value` → `key=value`. We trim whitespace
		// around the separator so both `key: value` and `key : value`
		// normalise to the same canonical form.
		if eq := strings.IndexByte(trimmed, '='); eq < 0 {
			if colon := strings.IndexByte(trimmed, ':'); colon > 0 {
				key := strings.TrimRight(trimmed[:colon], " \t")
				val := strings.TrimLeft(trimmed[colon+1:], " \t")
				trimmed = key + "=" + val
				report.AddFix("colon-to-equals")
			}
		}

		// Strip trailing whitespace.
		t := rstrip(trimmed)
		if t != trimmed {
			report.AddFix("trailing-whitespace")
			trimmed = t
		}

		lines[idx] = indent + trimmed
	}

	// ---- Stage 3 (lenient): drop duplicate keys in a section ----
	if level == LevelLenient {
		report.AddStage(3)

		// Pass 1: find the line index of the LAST occurrence of each
		// (section, key) pair. We need the section context to
		// distinguish `[a]\nx=1` from `[b]\nx=2`.
		type keyID struct{ section, key string }
		lastOccurrence := map[keyID]int{}

		var section string
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				section = trimmed
				continue
			}
			if trimmed == "" || strings.HasPrefix(trimmed, ";") ||
				strings.HasPrefix(trimmed, "#") {
				continue
			}
			if eq := strings.IndexByte(trimmed, '='); eq > 0 {
				k := keyID{section, strings.TrimSpace(trimmed[:eq])}
				lastOccurrence[k] = i
			}
		}

		// Pass 2: keep only the last occurrence of each key.
		out := make([]string, 0, len(lines))
		section = ""
		duplicates := 0
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
				section = trimmed
				out = append(out, line)
				continue
			}
			if trimmed == "" || strings.HasPrefix(trimmed, ";") ||
				strings.HasPrefix(trimmed, "#") {
				out = append(out, line)
				continue
			}
			if eq := strings.IndexByte(trimmed, '='); eq > 0 {
				k := keyID{section, strings.TrimSpace(trimmed[:eq])}
				if lastOccurrence[k] != i {
					duplicates++
					report.AddFix("duplicate-key-dropped")
					continue
				}
			}
			out = append(out, line)
		}
		if duplicates > 0 {
			report.AddWarning("duplicate keys were dropped; the last occurrence wins")
		}
		lines = out
	}

	if !report.Changed {
		return raw, report, nil
	}

	var buf bytes.Buffer
	buf.WriteString(strings.Join(lines, "\n"))
	if !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteByte('\n')
	}
	return buf.Bytes(), report, nil
}