package repair

import (
	"bytes"
	"regexp"
	"strings"
)

func init() {
	register(&markdownRepairer{})
}

// markdownRepairer implements Repairer for Markdown documents.
//
// Repair here is stylistic: Markdown has no well-defined "valid" /
// "invalid" boundary, so the repairer applies opinionated
// consistency fixes that make the document render predictably
// across CommonMark, GFM, and the various static-site generators.
type markdownRepairer struct{}

func (m *markdownRepairer) Format() string { return "md" }

var (
	// #Heading (no space) → ATX heading requires a space after #.
	reHeadingNoSpace = regexp.MustCompile(`^(#{1,6})([^\s#])`)
)

func (m *markdownRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "md", Level: level}

	report.AddStage(0)
	work, hadBOM := stripBOM(raw)
	if hadBOM {
		report.AddFix("bom-stripped")
	}
	work = normaliseLineEndings(work)

	lines := strings.Split(string(work), "\n")

	// ---- Stage 1: heading spacing + trailing whitespace ----
	report.AddStage(1)
	for i, line := range lines {
		// #Heading → # Heading
		if m := reHeadingNoSpace.FindStringSubmatchIndex(line); m != nil {
			// m[3] is the END of group 1 (the # run). Inserting a
			// space there gives "# Heading".
			lines[i] = line[:m[3]] + " " + line[m[3]:]
			report.AddFix("heading-space")
			line = lines[i]
		}
		// Trailing whitespace.
		if t := rstrip(line); t != line {
			lines[i] = t
			report.AddFix("trailing-whitespace")
		}
	}

	// ---- Stage 2: collapse multiple blank lines ----
	if level != LevelStrict {
		report.AddStage(2)
		collapsed := make([]string, 0, len(lines))
		blank := 0
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				blank++
				if blank > 1 {
					report.AddFix("blank-lines-collapsed")
					continue
				}
			} else {
				blank = 0
			}
			collapsed = append(collapsed, line)
		}
		lines = collapsed
	}

	// ---- Stage 3 (lenient): ensure trailing newline ----
	if level == LevelLenient {
		report.AddStage(3)
		// Nothing additional — the joinLines helper handles it.
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