package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"file-converter/internal/formatters"
)

// HandleConvert is the universal HTTP handler for file conversions.
// It delegates to the format registry (internal/formatters) instead of
// a hardcoded switch, so any registered (fromType, toType) pair is
// automatically reachable. The full set of supported pairs is
// discoverable at GET /formats.
//
// @Summary Universal File Converter
// @Description Converts any supported file type to another supported file type dynamically. The set of supported (fromType, toType) pairs is discoverable at GET /formats.
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param fromType formData string true "Source file type (e.g. txt, csv, json, pdf, docx, jpg, png, webp, tiff, bmp)"
// @Param toType formData string true "Target file type (e.g. txt, csv, json, pdf, docx, jpg, png, webp, tiff, bmp)"
// @Param file formData file true "The file to convert"
// @Success 200 {file} file "The converted file"
// @Failure 400 {string} string "Bad Request or unsupported conversion pair"
// @Failure 405 {string} string "Method Not Allowed"
// @Router /convert [post]
func HandleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 10 MB for safety.
	r.ParseMultipartForm(10 << 20)

	fromType := strings.ToLower(strings.TrimSpace(r.FormValue("fromType")))
	toType := strings.ToLower(strings.TrimSpace(r.FormValue("toType")))

	if fromType == "" || toType == "" {
		http.Error(w, "Missing fromType or toType in form data", http.StatusBadRequest)
		return
	}

	// Look up the registered formatter for this directed pair.
	formatter, ok := formatters.Lookup(fromType, toType)
	if !ok {
		http.Error(
			w,
			fmt.Sprintf("Conversion from %s to %s is not yet supported. Please choose a different combination.", fromType, toType),
			http.StatusBadRequest,
		)
		return
	}

	// Pull the file out of the multipart form.
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request. Ensure form-data key is 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Write the response headers up front so the client always sees
	// the correct MIME and filename, even if the conversion later fails.
	w.Header().Set("Content-Type", formatter.OutputMIME)
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=\"converted%s\"", formatter.OutputExt),
	)

	if err := formatter.Convert(r.Context(), file, w); err != nil {
		// Undo the success headers and replace with an error.
		w.Header().Del("Content-Disposition")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusBadRequest)
		log.Printf("Conversion error (%s→%s): %v", fromType, toType, err)
		return
	}
}

// EnableCORS adds CORS headers to allow cross-origin requests
func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
