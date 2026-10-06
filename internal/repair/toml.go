package repair

import (
	"bytes"
	"strings"
)

func init() {
	register(&tomlRepairer{})
}

// tomlRepairer implements Repairer for TOML documents.
type tomlRepairer struct{}

func (t *tomlRepairer) Format() string { return "toml" }

// Repair normalises a TOML document.
//
// Rules:
//
//	strict  — strip BOM, normalise line endings, strip trailing
//	          whitespace per line, convert `;` comments to `#`,
//	          quote bare string values that contain spaces.
//	normal  — quote unquoted string values.
//	lenient — drop lines that are neither key=value, [section],
//	          comments, nor blank.
func (t *tomlRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "toml", Level: level}

	report.AddStage(0)
	work, hadBOM := stripBOM(raw)
	if hadBOM {
		report.AddFix("bom-stripped")
	}
	work = normaliseLineEndings(work)

	lines := strings.Split(string(work), "\n")

	// ---- Stage 1: per-line cleanups ----
	report.AddStage(1)
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		indent := line[:len(line)-len(trimmed)]

		// Convert `;` comments to `#` (INI-style habit).
		if idx := strings.Index(trimmed, ";"); idx >= 0 && !isInsideString(trimmed, idx) {
			trimmed = trimmed[:idx] + "#" + trimmed[idx+1:]
			report.AddFix("semicolon-comment")
		}

		// Strip trailing whitespace.
		if t := rstrip(trimmed); t != trimmed {
			report.AddFix("trailing-whitespace")
			trimmed = t
		}

		lines[i] = indent + trimmed
	}

	// ---- Stage 2: quote unquoted string values with spaces ----
	if level != LevelStrict {
		report.AddStage(2)
		for i, line := range lines {
			if out, changed := quoteTomlValueIfNeeded(line); changed {
				lines[i] = out
				report.AddFix("value-quoted")
			}
		}
	}

	// ---- Stage 3: drop junk lines ----
	if level == LevelLenient {
		report.AddStage(3)
		out := make([]string, 0, len(lines))
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") ||
				strings.HasPrefix(trimmed, "[") {
				out = append(out, line)
				continue
			}
			if strings.Contains(trimmed, "=") {
				out = append(out, line)
				continue
			}
			report.AddFix("junk-line-dropped")
			report.AddWarning("dropped malformed line: " + trimmed)
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

// isInsideString reports whether position i in line falls between
// matching quotes.
func isInsideString(line string, i int) bool {
	inQuote := byte(0)
	for j := 0; j < i && j < len(line); j++ {
		c := line[j]
		if inQuote == 0 {
			if c == '"' || c == '\'' {
				inQuote = c
			}
		} else if c == inQuote {
			inQuote = 0
		}
	}
	return inQuote != 0
}

// quoteTomlValueIfNeeded wraps a bare value in double quotes when it
// contains a space. Skips values that are already quoted, arrays,
// numbers, booleans, and dates.
func quoteTomlValueIfNeeded(line string) (string, bool) {
	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return line, false
	}
	key := line[:eq]
	val := strings.TrimSpace(line[eq+1:])
	if val == "" {
		return line, false
	}
	// Already quoted?
	if val[0] == '"' || val[0] == '\'' || val[0] == '[' || val[0] == '{' {
		return line, false
	}
	// Numbers, booleans, dates?
	if isTomlScalar(val) {
		return line, false
	}
	// Contains a space? Quote it.
	if strings.ContainsAny(val, " \t") {
		return key + " = \"" + val + "\"", true
	}
	return line, false
}

// isTomlScalar returns true for values that don't need quoting.
func isTomlScalar(v string) bool {
	if v == "true" || v == "false" {
		return true
	}
	for i, c := range v {
		switch {
		case c >= '0' && c <= '9':
		case c == '.' || c == '-' || c == '+' || c == '_' || c == ':':
		case (c == 'T' || c == 'Z') && i > 0:
		case c == 'e' || c == 'E':
		default:
			return false
		}
	}
	return true
}