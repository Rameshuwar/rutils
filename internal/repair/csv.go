package repair

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"
)

func init() {
	register(&csvRepairer{})
}

// csvRepairer implements Repairer for CSV documents. See the header
// comment on Repair for the exact rules and their levels.
type csvRepairer struct{}

func (c *csvRepairer) Format() string { return "csv" }

// Repair normalises a CSV document.
//
// Rules, by level:
//
//	strict  — strip BOM, normalise line endings, strip trailing
//	          whitespace per line, drop trailing blank lines.
//	normal  — pad short rows with empty fields, truncate long rows
//	          (with a warning). Refuse files whose header and data
//	          use different delimiters — that is genuinely ambiguous.
//	lenient — rename empty header cells to column_N.
//
// DECISION: ragged rows are padded when short and truncated when
// long. This matches Excel / Google Sheets / most importers.
func (c *csvRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "csv", Level: level}

	// ---- Stage 0: normalise ----
	report.AddStage(0)
	work, hadBOM := stripBOM(raw)
	if hadBOM {
		report.AddFix("bom-stripped")
	}
	work = normaliseLineEndings(work)

	lines := splitLines(work)
	if len(lines) == 0 {
		return raw, report, ErrEmptyInput
	}

	// If splitLines dropped trailing blank lines, record that as a
	// fix. This is cheap to compute: if the input ended with more
	// than one newline, splitLines dropped at least one blank line.
	if len(work) > 0 &&
		(bytes.HasSuffix(work, []byte("\n\n")) ||
			bytes.HasSuffix(work, []byte("\n \n"))) {
		report.AddFix("trailing-blank-lines")
	}

	// Detect the dominant delimiter. If we can't, the file is
	// probably not CSV — fall back to comma but warn.
	delim, ok := detectDominantDelimiter(lines, 20)
	if !ok {
		delim = ','
		report.AddWarning("could not detect a dominant delimiter; assuming comma")
	}

	// Detect a specific kind of ambiguity: the header uses one
	// delimiter, but the data rows use another. This is a real
	// pattern (locale-mismatched Excel exports), but there is no
	// unambiguous way to decide which schema the user meant. We
	// refuse rather than guess and silently destroy data.
	if len(lines) >= 2 {
		headerDelim := sniffHeaderDelimiter(lines[0])
		dataDelim := sniffHeaderDelimiter(strings.Join(lines[1:], "\n"))
		if headerDelim != 0 && dataDelim != 0 && headerDelim != dataDelim {
			return raw, report, ErrUnrepairable
		}
	}

	// ---- Stage 1: strip trailing whitespace per line ----
	report.AddStage(1)
	trimmed := make([]string, len(lines))
	for i, line := range lines {
		t := rstrip(line)
		if t != line {
			report.AddFix("trailing-whitespace")
		}
		trimmed[i] = t
	}
	lines = trimmed

	// ---- Stage 2: parse, repair ragged rows, re-emit ----
	report.AddStage(2)

	reader := csv.NewReader(strings.NewReader(strings.Join(lines, "\n")))
	reader.Comma = rune(delim)
	reader.FieldsPerRecord = -1 // don't error on ragged rows
	reader.LazyQuotes = true    // tolerate slightly broken quoting

	var records [][]string
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return raw, report, ErrUnrepairable
		}
		records = append(records, rec)
	}

	if len(records) == 0 {
		return raw, report, ErrEmptyInput
	}

	// Determine the header width from the first row. This is the
	// canonical column count for the whole document.
	width := len(records[0])
	if width == 0 {
		return raw, report, ErrUnrepairable
	}

	// Trim trailing empty columns: some exports add a trailing
	// comma, producing a spurious empty last column.
	for width > 1 {
		allEmpty := true
		for _, rec := range records {
			if len(rec) >= width && strings.TrimSpace(rec[width-1]) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			width--
			report.AddFix("trailing-empty-column")
		} else {
			break
		}
	}

	// Normalise every row to exactly `width` columns.
	ragged := 0
	for i, rec := range records {
		switch {
		case len(rec) < width:
			padded := make([]string, width)
			copy(padded, rec)
			records[i] = padded
			ragged++
			report.AddFix("row-padded")
		case len(rec) > width:
			records[i] = rec[:width]
			ragged++
			report.AddFix("row-truncated")
		}
	}
	if ragged > 0 {
		report.AddWarning("one or more rows had the wrong number of columns and were normalised")
	}

	// ---- Stage 3: lenient-only header cleanup ----
	if level == LevelLenient {
		report.AddStage(3)
		header := records[0]
		for i, h := range header {
			if strings.TrimSpace(h) == "" {
				header[i] = "column_" + itoa(i+1)
				report.AddFix("empty-header-renamed")
			}
		}
	}

	// ---- Re-emit ----
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = rune(delim)
	for _, rec := range records {
		if err := writer.Write(rec); err != nil {
			return raw, report, ErrUnrepairable
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return raw, report, ErrUnrepairable
	}

	// If every stage was a no-op, return the original bytes to
	// preserve byte-for-byte fidelity (including original line
	// endings and quoting style).
	if !report.Changed {
		return raw, report, nil
	}

	return buf.Bytes(), report, nil
}

// sniffHeaderDelimiter returns the delimiter that appears most often
// in s, or 0 if none appears. Used to detect mixed-delimiter CSV
// files where the header and data disagree.
func sniffHeaderDelimiter(s string) byte {
	candidates := []byte{',', ';', '\t', '|'}
	best := byte(0)
	bestCount := 0
	for _, d := range candidates {
		n := countDelimiter(s, d)
		if n > bestCount {
			best = d
			bestCount = n
		}
	}
	return best
}
