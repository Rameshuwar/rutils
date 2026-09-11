package api

import (
	"fmt"
	"log"
	"net/http"

	"file-converter/internal/converter"
)

// HandleConvert is the HTTP handler for file conversions
func HandleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 10 MB for safety
	r.ParseMultipartForm(10 << 20)

	fromType := r.FormValue("fromType")
	toType := r.FormValue("toType")

	if fromType == "" || toType == "" {
		http.Error(w, "Missing fromType or toType in form data", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request. Ensure form-data key is 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Default disposition
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"converted.%s\"", toType))

	conversionPath := fmt.Sprintf("%s-to-%s", fromType, toType)

	switch conversionPath {
	case "jpg-to-png":
		w.Header().Set("Content-Type", "image/png")
		err = converter.ConvertJPGtoPNG(file, w)
	case "csv-to-pdf":
		w.Header().Set("Content-Type", "application/pdf")
		err = converter.ConvertCSVtoPDF(file, w)
	case "csv-to-json":
		w.Header().Set("Content-Type", "application/json")
		err = converter.ConvertCSVtoJSON(file, w)
	case "json-to-csv":
		w.Header().Set("Content-Type", "text/csv")
		err = converter.ConvertJSONtoCSV(file, w)
	case "json-to-txt":
		w.Header().Set("Content-Type", "text/plain")
		err = converter.ConvertJSONtoTXT(file, w)
	case "txt-to-docx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		err = converter.ConvertTXTtoDOCX(file, w)
	case "pdf-to-txt":
		w.Header().Set("Content-Type", "text/plain")
		err = converter.ConvertPDFtoTXT(file, fileHeader.Size, w)
	case "txt-to-pdf":
		w.Header().Set("Content-Type", "application/pdf")
		err = converter.ConvertTXTtoPDF(file, w)
	case "png-to-jpg":
		w.Header().Set("Content-Type", "image/jpeg")
		err = converter.ConvertPNGtoJPG(file, w)
	case "csv-to-txt":
		w.Header().Set("Content-Type", "text/plain")
		err = converter.ConvertCSVtoTXT(file, w)
	case "docx-to-csv":
		w.Header().Set("Content-Type", "text/csv")
		err = converter.ConvertDOCXtoCSV(file, w)
	case "docx-to-txt":
		w.Header().Set("Content-Type", "text/plain")
		err = converter.ConvertDOCXtoCSV(file, w)
	default:
		http.Error(w, fmt.Sprintf("Conversion from %s to %s is not yet supported. Please choose a different combination.", fromType, toType), http.StatusBadRequest)
		return
	}

	if err != nil {
		w.Header().Del("Content-Disposition")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusBadRequest)
		log.Printf("Conversion error (%s): %v", conversionPath, err)
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
