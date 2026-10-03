package repair

import (
	"sort"
	"strings"
)

// Report describes exactly what the engine did to a document.
//
// It is deliberately a plain data struct — no methods that do I/O, no
// pointers into internal state. The API layer receives a copy and can
// serialise it however it likes.
type Report struct {
	Format   string   `json:"format"`             // "json", "csv", "xml", "yaml"
	Level    Level    `json:"level"`              // strict / normal / lenient
	Changed  bool     `json:"changed"`            // false → input was already valid
	Applied  []string `json:"applied,omitempty"`  // e.g. ["trailing-comma","single-quotes"]
	Warnings []string `json:"warnings,omitempty"` // e.g. ["NaN coerced to null"]
	Stages   []int    `json:"stages,omitempty"`   // which stages actually ran: [0,1,2]
}

// AddFix records a repair that was applied. It de-duplicates, so a
// repairer can call AddFix("trailing-comma") on every occurrence
// without the caller seeing the same label forty times.
func (r *Report) AddFix(name string) {
	for _, a := range r.Applied {
		if a == name {
			return
		}
	}
	r.Applied = append(r.Applied, name)
	r.Changed = true
}

// AddWarning records a non-fatal notice. Warnings are surfaced to the
// user but do not cause the request to fail.
func (r *Report) AddWarning(msg string) {
	r.Warnings = append(r.Warnings, msg)
}

// AddStage records that a stage was entered. Stages are recorded even
// when they made no changes, because "stage 2 ran and found nothing to
// do" is different from "stage 2 was skipped".
func (r *Report) AddStage(n int) {
	for _, s := range r.Stages {
		if s == n {
			return
		}
	}
	r.Stages = append(r.Stages, n)
	sort.Ints(r.Stages)
}

// HeaderValue returns a compact, HTTP-header-safe summary of the report.
//
//	"json;level=normal;fixed=trailing-comma,single-quotes;warn=1;changed=true"
//
// The value is deliberately short and ASCII-only so it survives
// proxies, Cloudflare, and every browser without escaping.
func (r *Report) HeaderValue() string {
	if r == nil || !r.Changed {
		if r == nil {
			return "none"
		}
		return r.Format + ";changed=false"
	}

	var b strings.Builder
	b.WriteString(r.Format)
	b.WriteString(";level=")
	b.WriteString(string(r.Level))
	b.WriteString(";changed=true")

	if len(r.Applied) > 0 {
		b.WriteString(";fixed=")
		b.WriteString(strings.Join(r.Applied, ","))
	}
	if len(r.Warnings) > 0 {
		b.WriteString(";warn=")
		// Use a simple decimal count instead of the text itself — the
		// header must stay well under 8 KB regardless of input size.
		b.WriteString(itoa(len(r.Warnings)))
	}
	return b.String()
}

// itoa is a tiny dependency-free int→string for header building. We
// avoid strconv here purely to keep this file import-free beyond strings.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}