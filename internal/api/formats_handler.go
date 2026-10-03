package api

import (
	"net/http"
	"strings"

	formatters "file-converter/internal/formatters"
)

// HandleListFormats returns the full conversion matrix.
//
// @Summary      List supported file formats
// @Description  Returns every supported source format plus every directed conversion edge. Intended for the frontend to populate dropdowns at runtime.
// @Tags         Format Registry
// @Produce      json
// @Success      200 {object} formatters.FormatListResponse "Conversion matrix"
// @Router       /formats [get]
func HandleListFormats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, formatters.FormatListResponse{
		Sources:     formatters.Sources(),
		Conversions: formatters.Conversions(),
	})
}

// HandleFormatDetail returns metadata for a single format.
//
// @Summary      Describe a single format
// @Description  Returns the MIME type, extension, category, and reachable conversions for one format (e.g. "jpg"). Path is /formats/{type}.
// @Tags         Format Registry
// @Produce      json
// @Param        type path string true "Format identifier (e.g. jpg, pdf, csv)"
// @Success      200 {object} formatters.FormatDetail "Format detail"
// @Failure      404 {object} map[string]string "Unknown format"
// @Router       /formats/{type} [get]
func HandleFormatDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// /formats/{type} → strip the "/formats/" prefix.
	formatType := strings.TrimPrefix(r.URL.Path, "/formats/")
	formatType = strings.Trim(formatType, "/")

	if formatType == "" {
		writeError(w, http.StatusBadRequest, "format type is required")
		return
	}

	detail, ok := formatters.Detail(formatType)
	if !ok {
		writeError(w, http.StatusNotFound, "unknown format: "+formatType)
		return
	}

	writeJSON(w, http.StatusOK, detail)
}
