package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"file-converter/internal/repair"
)

// HandleRepair implements POST /repair.
//
// Accepts multipart/form-data:
//   file         (required) — the malformed bytes
//   format       (optional) — force a source format: json, csv, xml, yaml, html, toml, ini, md
//   repairLevel  (optional) — strict | normal (default) | lenient
//   describe     (optional) — "true" to wrap the response in a JSON report
//
// When describe=true, the response is:
//
//	{
//	  "format":   "json",
//	  "level":    "normal",
//	  "changed":  true,
//	  "applied":  ["trailing-comma"],
//	  "warnings": [],
//	  "output":   "<the repaired document as a string>"
//	}
//
// When describe is unset, the repaired bytes are returned directly.
func HandleRepair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	// ---- Parse multipart form (10 MB cap) ----
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required: "+err.Error())
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read uploaded file: "+err.Error())
		return
	}
	if len(raw) == 0 {
		writeError(w, http.StatusBadRequest, "input is empty")
		return
	}

	// ---- Resolve the repair level ----
	level := repair.LevelNormal
	if lvl := strings.TrimSpace(r.FormValue("repairLevel")); lvl != "" {
		parsed, err := repair.ParseLevel(lvl)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid repairLevel: "+err.Error())
			return
		}
		level = parsed
	}

	// ---- Determine the filename passed to the engine (drives DetectFormat) ----
	filename := ""
	if header != nil {
		filename = header.Filename
	}
	// If the caller forces a format via the "format" field, we spoof a
	// filename with the matching extension so DetectFormat's extension
	// fallback picks it up. Content sniffing (magic bytes) still wins
	// when the bytes are unambiguous.
	if forced := strings.ToLower(strings.TrimSpace(r.FormValue("format"))); forced != "" {
		filename = "upload." + forced
	}

	// ---- Run the repair pipeline ----
	engine := repair.NewEngine()

	repaired, report, err := engine.Repair(filename, raw, level)
	if err != nil {
		switch err {
		case repair.ErrUnsupportedFormat:
			writeError(w, http.StatusUnsupportedMediaType, err.Error())
		case repair.ErrEmptyInput:
			writeError(w, http.StatusBadRequest, err.Error())
		case repair.ErrInvalidLevel:
			writeError(w, http.StatusBadRequest, err.Error())
		case repair.ErrUnrepairable:
			writeError(w, http.StatusBadRequest,
				"input could not be repaired at the requested level")
		default:
			log.Printf("[repair] unexpected error: %v", err)
			writeError(w, http.StatusInternalServerError, "Repair failed: "+err.Error())
		}
		return
	}

	// ---- Emit the compact X-Repair-* headers used by the frontend ----
	w.Header().Set("X-Repair-Applied", report.HeaderValue())
	w.Header().Set("X-Repair-Format", report.Format)
	w.Header().Set("X-Repair-Level", string(report.Level))
	if report.Changed {
		w.Header().Set("X-Repair-Changed", "true")
	} else {
		w.Header().Set("X-Repair-Changed", "false")
	}
	// Make them readable from JS (CORS defaults to exposing nothing).
	w.Header().Set("Access-Control-Expose-Headers",
		"X-Repair-Applied, X-Repair-Format, X-Repair-Level, X-Repair-Changed")

	// ---- describe=true → JSON envelope ----
	if strings.EqualFold(strings.TrimSpace(r.FormValue("describe")), "true") {
		w.Header().Set("Content-Type", "application/json")

		applied := report.Applied
		if applied == nil {
			applied = []string{}
		}
		warnings := report.Warnings
		if warnings == nil {
			warnings = []string{}
		}

		envelope := map[string]any{
			"format":   report.Format,
			"level":    string(report.Level),
			"changed":  report.Changed,
			"applied":  applied,
			"warnings": warnings,
			"output":   string(repaired),
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(envelope); err != nil {
			log.Printf("[repair] failed to encode JSON envelope: %v", err)
		}
		return
	}

	// ---- describe=false → raw repaired bytes ----
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition",
		`attachment; filename="repaired.`+report.Format+`"`)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(repaired); err != nil {
		log.Printf("[repair] failed to write repaired bytes: %v", err)
	}
}