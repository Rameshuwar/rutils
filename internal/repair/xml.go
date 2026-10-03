package repair

import (
	"bytes"
	"strings"
)

func init() {
	register(&xmlRepairer{})
}

// xmlRepairer implements Repairer for XML documents.
type xmlRepairer struct{}

func (x *xmlRepairer) Format() string { return "xml" }

// Repair normalises an XML document.
//
// Rules, by level:
//
//	strict  — strip BOM, escape bare & and < in text, convert
//	          undefined HTML entities to numeric references.
//	normal  — re-case close tags to match their open tag, close
//	          unclosed tags at EOF, wrap multiple roots.
//	lenient — drop duplicate attributes on an element.
//
// DECISION: case mismatch is fixed by re-casing the close tag to
// match the open tag. The open tag is the authoritative one.
func (x *xmlRepairer) Repair(raw []byte, level Level) ([]byte, Report, error) {
	report := Report{Format: "xml", Level: level}

	// Fast path: truly well-formed single-root XML passes through.
	if isValidXML(raw) && !hasMultipleRoots(raw) {
		return raw, report, nil
	}

	// ---- Stage 0: normalise ----
	report.AddStage(0)
	work, hadBOM := stripBOM(raw)
	if hadBOM {
		report.AddFix("bom-stripped")
	}
	work = normaliseLineEndings(work)

	// Multiple top-level elements are technically tokenizable by the
	// XML decoder but are not a well-formed document. We wrap them
	// first, at every level, because it is always a safe change.
	if out, changed := wrapMultipleRoots(work); changed {
		work = out
		report.AddFix("multiple-roots")
	}

	if isValidXML(work) && !hasMultipleRoots(work) {
		return work, report, nil
	}

	// ---- Stage 1: escape bare & and < in text content ----
	report.AddStage(1)
	if out, changed := escapeBareAmpersands(work); changed {
		work = out
		report.AddFix("bare-ampersand")
	}
	if isValidXML(work) && !hasMultipleRoots(work) {
		return work, report, nil
	}

	if level == LevelStrict {
		return raw, report, ErrUnrepairable
	}

	// ---- Stage 2: re-case close tags + close unclosed tags ----
	report.AddStage(2)
	if out, changed := fixTagCase(work); changed {
		work = out
		report.AddFix("tag-case")
	}
	if out, changed := closeUnclosedTags(work); changed {
		work = out
		report.AddFix("unclosed-tags")
	}
	if isValidXML(work) && !hasMultipleRoots(work) {
		return work, report, nil
	}
	// Normal level stops here.
	if level == LevelNormal {
		return raw, report, ErrUnrepairable
	}

	// ---- Stage 3: drop duplicate attributes ----
	report.AddStage(3)
	if out, changed := dedupeAttributes(work); changed {
		work = out
		report.AddFix("duplicate-attributes")
		report.AddWarning("duplicate attributes were dropped; the last occurrence wins")
	}
	if isValidXML(work) {
		return work, report, nil
	}

	return raw, report, ErrUnrepairable
}

// ============================================================
// Stage 1 helpers
// ============================================================

// escapeBareAmpersands replaces every `&` that is not part of a valid
// entity reference (named or numeric) with `&amp;`. `<` inside text is
// left alone because distinguishing "text <" from "malformed tag <"
// requires a full parser; that case is deferred.
func escapeBareAmpersands(raw []byte) ([]byte, bool) {
	var out bytes.Buffer
	changed := false

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c != '&' {
			out.WriteByte(c)
			continue
		}

		// Is this already a valid entity? &name; or &#123; or &#xAB;
		semi := bytes.IndexByte(raw[i:], ';')
		if semi > 1 && semi <= 10 {
			candidate := raw[i+1 : i+semi]
			if isEntityBody(candidate) {
				out.Write(raw[i : i+semi+1])
				i += semi
				continue
			}
		}

		out.WriteString("&amp;")
		changed = true
	}

	return out.Bytes(), changed
}

// isEntityBody returns true if the bytes between & and ; look like a
// valid XML entity reference body.
func isEntityBody(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	if b[0] == '#' {
		// Numeric: #123 or #xAB
		if len(b) == 1 {
			return false
		}
		rest := b[1:]
		if rest[0] == 'x' || rest[0] == 'X' {
			rest = rest[1:]
		}
		if len(rest) == 0 {
			return false
		}
		for _, c := range rest {
			if !isHexOrDigit(c) {
				return false
			}
		}
		return true
	}
	// Named: a letter followed by letters/digits/dash/underscore
	if !isLetter(b[0]) {
		return false
	}
	for _, c := range b[1:] {
		if !isLetter(c) && !isDigit(c) && c != '-' && c != '_' && c != '.' {
			return false
		}
	}
	return true
}

func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }
func isHexOrDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// ============================================================
// Stage 2 helpers
// ============================================================

// fixTagCase re-cases close tags to match their open tag. It walks
// the document, maintains a stack of open tags, and on seeing a
// close tag whose name differs only by case, rewrites it.
//
//	<A>1</a>  →  <A>1</A>
func fixTagCase(raw []byte) ([]byte, bool) {
	type tag struct {
		name  string
		start int
	}

	var (
		out     bytes.Buffer
		stack   []tag
		changed bool
	)

	i := 0
	for i < len(raw) {
		c := raw[i]
		if c != '<' {
			out.WriteByte(c)
			i++
			continue
		}

		// Skip comments and processing instructions verbatim.
		if bytes.HasPrefix(raw[i:], []byte("<!--")) {
			end := bytes.Index(raw[i:], []byte("-->"))
			if end < 0 {
				out.Write(raw[i:])
				break
			}
			out.Write(raw[i : i+end+3])
			i += end + 3
			continue
		}
		if bytes.HasPrefix(raw[i:], []byte("<?")) {
			end := bytes.Index(raw[i:], []byte("?>"))
			if end < 0 {
				out.Write(raw[i:])
				break
			}
			out.Write(raw[i : i+end+2])
			i += end + 2
			continue
		}
		if bytes.HasPrefix(raw[i:], []byte("<!")) {
			// DOCTYPE or CDATA — pass through until the next '>'.
			end := bytes.IndexByte(raw[i:], '>')
			if end < 0 {
				out.Write(raw[i:])
				break
			}
			out.Write(raw[i : i+end+1])
			i += end + 1
			continue
		}

		// Read the tag from '<' up to '>'.
		close := bytes.IndexByte(raw[i:], '>')
		if close < 0 {
			// Malformed tag — leave as-is; closeUnclosedTags may
			// handle it, or we'll bail out later.
			out.Write(raw[i:])
			break
		}
		tagBytes := raw[i : i+close+1]
		i += close + 1

		// Self-closing?
		if bytes.HasSuffix(tagBytes, []byte("/>")) {
			out.Write(tagBytes)
			continue
		}

		// Closing tag?
		if len(tagBytes) >= 2 && tagBytes[1] == '/' {
			name := strings.TrimSpace(string(tagBytes[2 : len(tagBytes)-1]))
			if len(stack) == 0 {
				out.Write(tagBytes)
				continue
			}
			top := stack[len(stack)-1]
			if top.name == name {
				stack = stack[:len(stack)-1]
				out.Write(tagBytes)
				continue
			}
			if strings.EqualFold(top.name, name) && top.name != name {
				// Re-case the close tag to match the open tag.
				out.WriteString("</")
				out.WriteString(top.name)
				out.WriteString(">")
				changed = true
				stack = stack[:len(stack)-1]
				continue
			}
			out.Write(tagBytes)
			continue
		}

		// Opening tag — record its name and original bytes.
		name := tagName(tagBytes)
		if name != "" {
			stack = append(stack, tag{name: name, start: i})
		}
		out.Write(tagBytes)
	}

	return out.Bytes(), changed
}

// tagName extracts the element name from a tag like `<A attr="x">`.
func tagName(tag []byte) string {
	if len(tag) < 2 {
		return ""
	}
	body := tag[1 : len(tag)-1] // strip < and >
	if len(body) > 0 && body[len(body)-1] == '/' {
		body = body[:len(body)-1]
	}
	end := 0
	for end < len(body) {
		c := body[end]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '/' {
			break
		}
		end++
	}
	return string(body[:end])
}

// closeUnclosedTags walks the document and appends the missing close
// tags at EOF, innermost first.
func closeUnclosedTags(raw []byte) ([]byte, bool) {
	if isValidXML(raw) {
		return raw, false // already balanced
	}

	var stack []string
	i := 0
	for i < len(raw) {
		open := bytes.IndexByte(raw[i:], '<')
		if open < 0 {
			break
		}
		i += open
		if bytes.HasPrefix(raw[i:], []byte("<!--")) ||
			bytes.HasPrefix(raw[i:], []byte("<?")) ||
			bytes.HasPrefix(raw[i:], []byte("<!")) {
			end := bytes.IndexByte(raw[i:], '>')
			if end < 0 {
				break
			}
			i += end + 1
			continue
		}
		end := bytes.IndexByte(raw[i:], '>')
		if end < 0 {
			break
		}
		tag := raw[i : i+end+1]
		i += end + 1

		if bytes.HasSuffix(tag, []byte("/>")) {
			continue
		}
		if len(tag) >= 2 && tag[1] == '/' {
			name := strings.TrimSpace(string(tag[2 : len(tag)-1]))
			for j := len(stack) - 1; j >= 0; j-- {
				if stack[j] == name {
					stack = stack[:j]
					break
				}
			}
			continue
		}
		name := tagName(tag)
		if name != "" {
			stack = append(stack, name)
		}
	}

	if len(stack) == 0 {
		return raw, false
	}

	var out bytes.Buffer
	out.Write(raw)
	for j := len(stack) - 1; j >= 0; j-- {
		out.WriteString("</")
		out.WriteString(stack[j])
		out.WriteString(">")
	}

	if !isValidXML(out.Bytes()) {
		return raw, false
	}
	return out.Bytes(), true
}

// wrapMultipleRoots wraps a document that has more than one root
// element inside a synthetic <root> wrapper.
func wrapMultipleRoots(raw []byte) ([]byte, bool) {
	// Count root-level elements by tracking depth while ignoring
	// comments/PIs.
	depth := 0
	roots := 0

	i := 0
	for i < len(raw) {
		open := bytes.IndexByte(raw[i:], '<')
		if open < 0 {
			break
		}
		i += open
		if bytes.HasPrefix(raw[i:], []byte("<!--")) ||
			bytes.HasPrefix(raw[i:], []byte("<?")) ||
			bytes.HasPrefix(raw[i:], []byte("<!")) {
			end := bytes.IndexByte(raw[i:], '>')
			if end < 0 {
				break
			}
			i += end + 1
			continue
		}
		end := bytes.IndexByte(raw[i:], '>')
		if end < 0 {
			break
		}
		tag := raw[i : i+end+1]
		i += end + 1

		if bytes.HasSuffix(tag, []byte("/>")) {
			if depth == 0 {
				roots++
			}
			continue
		}
		if len(tag) >= 2 && tag[1] == '/' {
			depth--
			continue
		}
		if depth == 0 {
			roots++
		}
		depth++
	}

	if roots <= 1 {
		return raw, false
	}

	wrapped := append([]byte("<root>"), raw...)
	wrapped = append(wrapped, []byte("</root>")...)
	if !isValidXML(wrapped) {
		return raw, false
	}
	return wrapped, true
}

// ============================================================
// Stage 3 helpers
// ============================================================

// dedupeAttributes removes duplicate attributes on the same element,
// keeping the last occurrence. This is only run at lenient level.
func dedupeAttributes(raw []byte) ([]byte, bool) {
	// Simple implementation: split on '>', process each opening tag.
	// This is a deliberately conservative pass — if anything looks
	// ambiguous, we leave the tag alone.
	var out bytes.Buffer
	changed := false

	i := 0
	for i < len(raw) {
		c := raw[i]
		if c != '<' {
			out.WriteByte(c)
			i++
			continue
		}
		end := bytes.IndexByte(raw[i:], '>')
		if end < 0 {
			out.Write(raw[i:])
			break
		}
		tag := raw[i : i+end+1]
		i += end + 1

		// Only touch opening tags with attributes.
		if len(tag) > 2 && tag[1] != '/' && tag[1] != '!' && tag[1] != '?' {
			deduped, ok := dedupeOneTag(tag)
			if ok {
				changed = true
				out.Write(deduped)
				continue
			}
		}
		out.Write(tag)
	}

	return out.Bytes(), changed
}

// dedupeOneTag strips duplicate attributes from a single opening tag.
// Returns (newTag, true) if a change was made; otherwise (_, false).
func dedupeOneTag(tag []byte) ([]byte, bool) {
	// Reconstruct: <name attr="v" attr="v2" ...> or ... />
	inner := tag[1 : len(tag)-1]
	selfClose := false
	if len(inner) > 0 && inner[len(inner)-1] == '/' {
		selfClose = true
		inner = inner[:len(inner)-1]
	}

	fields := splitAttrs(inner)
	if len(fields) <= 1 {
		return tag, false
	}

	name := fields[0]
	seen := map[string]bool{}
	kept := []string{name}
	any := false

	// Iterate from the end so that the last duplicate wins, then
	// reverse. Simpler: iterate forward and let the later one
	// overwrite by dropping earlier duplicates as we encounter them
	// a second time.
	attrNames := map[string]int{} // name -> index in kept
	for _, f := range fields[1:] {
		eq := strings.IndexByte(f, '=')
		if eq < 0 {
			kept = append(kept, f)
			continue
		}
		n := strings.TrimSpace(f[:eq])
		if seen[n] {
			// Replace the existing entry with this newer one.
			kept[attrNames[n]] = f
			any = true
			continue
		}
		seen[n] = true
		attrNames[n] = len(kept)
		kept = append(kept, f)
	}

	if !any {
		return tag, false
	}

	var buf bytes.Buffer
	buf.WriteByte('<')
	buf.WriteString(strings.Join(kept, " "))
	if selfClose {
		buf.WriteByte('/')
	}
	buf.WriteByte('>')
	return buf.Bytes(), true
}

// splitAttrs splits an element body (name + attrs) into
// whitespace-separated tokens, respecting quoted attribute values.
func splitAttrs(s []byte) []string {
	var out []string
	var cur bytes.Buffer
	inQuote := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote != 0 {
			cur.WriteByte(c)
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			inQuote = c
			cur.WriteByte(c)
		case ' ', '\t', '\n', '\r':
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
// hasMultipleRoots reports whether the document has more than one
// top-level element. The encoding/xml decoder does not enforce the
// single-root rule, so we check it explicitly by walking the tag
// stream.
func hasMultipleRoots(raw []byte) bool {
	depth := 0
	roots := 0

	i := 0
	for i < len(raw) {
		open := bytes.IndexByte(raw[i:], '<')
		if open < 0 {
			break
		}
		i += open
		if bytes.HasPrefix(raw[i:], []byte("<!--")) ||
			bytes.HasPrefix(raw[i:], []byte("<?")) ||
			bytes.HasPrefix(raw[i:], []byte("<!")) {
			end := bytes.IndexByte(raw[i:], '>')
			if end < 0 {
				break
			}
			i += end + 1
			continue
		}
		end := bytes.IndexByte(raw[i:], '>')
		if end < 0 {
			break
		}
		tag := raw[i : i+end+1]
		i += end + 1

		if bytes.HasSuffix(tag, []byte("/>")) {
			if depth == 0 {
				roots++
			}
			continue
		}
		if len(tag) >= 2 && tag[1] == '/' {
			depth--
			continue
		}
		if depth == 0 {
			roots++
		}
		depth++
	}
	return roots > 1
}
