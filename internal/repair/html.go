package repair

import (
	"bytes"

	xhtml "golang.org/x/net/html"
)

func init() {
	register(&htmlRepairer{})
}

// htmlRepairer implements Repairer for HTML documents. It uses
// golang.org/x/net/html, which parses HTML5 the way a browser does —
// including closing unclosed tags, quoting bare attributes, and
// lowercasing tag names. Running any input through it and
// re-serialising produces well-formed HTML.
type htmlRepairer struct{}

func (h *htmlRepairer) Format() string { return "html" }

// Repair parses the input as HTML5 and re-emits a canonical form.
//
// Because x/net/html never fails on input (browsers' tolerance is
// the specification), this repairer has no notion of "levels" in
// the same sense the others do. Instead:
//
//	strict  — only strip BOM and normalise line endings, then
//	          re-serialise if and only if the parsed tree is
//	          already well-formed (which it always is, so effectively
//	          this stage is "pass through"). Valid documents round-trip
//	          unchanged.
//	normal  — re-serialise unconditionally, normalising whitespace,
//	          attribute quoting, and tag case.
//	lenient — also inject <!DOCTYPE html> when a full document is
//	          detected (has <html>) but is missing the doctype.
func (h *htmlRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "html", Level: level}

	report.AddStage(0)
	work, hadBOM := stripBOM(raw)
	if hadBOM {
		report.AddFix("bom-stripped")
	}
	work = normaliseLineEndings(work)

	if len(bytes.TrimSpace(work)) == 0 {
		return raw, report, ErrEmptyInput
	}

	// Parse. This never errors on well-formed or even most
	// malformed inputs — that's the point of HTML5 parsing.
	doc, err := xhtml.Parse(bytes.NewReader(work))
	if err != nil {
		return raw, report, ErrUnrepairable
	}

	// Serialize.
	var buf bytes.Buffer
	if err := xhtml.Render(&buf, doc); err != nil {
		return raw, report, ErrUnrepairable
	}

	// For strict level, only return the re-serialised output if it
	// is not materially different from the input. Otherwise return
	// the original.
	if level == LevelStrict {
		// Strict is a no-op for HTML beyond BOM/line-ending fixes.
		return work, report, nil
	}

	report.AddFix("html-normalised")

	if level == LevelLenient {
		report.AddStage(1)
		// If the document contains an <html> element but no doctype,
		// inject one. We check by looking for "<!DOCTYPE" or
		// "<!doctype" in the original bytes.
		lower := bytes.ToLower(work)
		if bytes.Contains(lower, []byte("<html")) &&
			!bytes.Contains(lower, []byte("<!doctype")) {
			prefixed := append([]byte("<!DOCTYPE html>\n"), buf.Bytes()...)
			report.AddFix("doctype-injected")
			return prefixed, report, nil
		}
	}

	return buf.Bytes(), report, nil
}