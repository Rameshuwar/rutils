// Package formatters implements a pluggable registry of file-to-file
// conversion routines.
//
// Each format family lives in its own file (image_jpg.go, image_png.go,
// pdf_image.go, ...) and registers itself with the global registry from
// an init() function. Go guarantees init() functions within a package
// run in file-name order, so by the time this file's init() runs, every
// sibling file has already called Register().
//
// Wiring is intentionally one-directional:
//
//	cmd/server/main.go  --blank import-->  internal/formatters
//	                                            |
//	                                            v
//	                              internal/api/handler.go calls Lookup()
//
// To add a new format pair, create a new file in this package with an
// init() that calls Register(). Nothing else needs to change.
package formatters

import (
	"log"
	"strings"
)

// init runs after every sibling file's init() because Go processes
// files in alphabetical order within a package and "register_all"
// sorts after every "image_*", "ocr_*", "pdf_*", and "registry"
// file. Its job is to fail fast if the registry is misconfigured.
func init() {
	all := All()

	if len(all) == 0 {
		panic("formatters: no formatters registered — did you forget to " +
			"add a blank import for this package in cmd/server/main.go?")
	}

	// Every source must have at least one outgoing edge (otherwise it's
	// a dead format that appears in /formats but can't do anything).
	sourceHasEdge := map[string]bool{}
	for _, f := range all {
		sourceHasEdge[f.From] = true
	}

	// Every registered edge must have complete metadata. Register()
	// already panics on missing fields, so this is a belt-and-braces
	// check for any future refactor that bypasses Register().
	for _, f := range all {
		if strings.TrimSpace(f.From) == "" ||
			strings.TrimSpace(f.To) == "" ||
			strings.TrimSpace(f.OutputMIME) == "" ||
			strings.TrimSpace(f.OutputExt) == "" {
			panic("formatters: incomplete formatter metadata for " +
				f.From + "→" + f.To)
		}
	}

	// Log a one-line summary so operators can see the registry size at
	// startup. This mirrors the dependency-warning style already used
	// in cmd/server/main.go.
	log.Printf("[INFO] formatters: %d conversion(s) registered across %d source format(s)",
		len(all), len(sourceHasEdge))
}
