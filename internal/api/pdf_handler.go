package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"file-converter/internal/converter"
)

// Upload limit raised to 600 MB per requirement.
const maxPDFUploadSize = 600 * 1024 * 1024 // 600 MB

// ConvertPDFSize handles PDF compression and expansion.
//
// @Summary      Compress or expand PDF
// @Description  Upload a PDF and convert it toward a requested target file size.
// @Description
// @Description  **Ranges:**
// @Description  - Compression: target must be between **10 KB** and **400 MB**.
// @Description  - Expansion:   target must be between **10 KB** and **6000 KB**.
// @Description  - Uploaded file must not exceed **600 MB**.
// @Description
// @Description  **Response headers:**
// @Description  - `X-Conversion-Target-Met`: `true` if the target was reached, `false` if the smallest achievable size was returned.
// @Description  - `X-Conversion-Target-Size`: the requested target in bytes.
// @Description  - `X-Conversion-Actual-Size`: the actual output size in bytes.
// @Tags         PDF Converter
// @Accept       multipart/form-data
// @Produce      application/pdf
// @Param        file formData file true "PDF file (max 600 MB)"
// @Param        conversionType formData string true "compression or expand"
// @Param        dataType formData string true "KB or MB"
// @Param        targetSize formData number true "Target file size"
// @Success      200 {file} binary "Converted PDF"
// @Failure      400 {string} string "Bad Request - invalid parameters"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /convert-pdf-size [post]
func ConvertPDFSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPDFUploadSize)

	if err := r.ParseMultipartForm(maxPDFUploadSize); err != nil {
		http.Error(
			w,
			fmt.Sprintf("failed to parse multipart form (max 600 MB): %v", err),
			http.StatusBadRequest,
		)
		return
	}

	// ----- Get uploaded file -----
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	input, err := io.ReadAll(file)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("failed to read uploaded PDF: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	if len(input) == 0 {
		http.Error(w, "uploaded PDF is empty", http.StatusBadRequest)
		return
	}

	// ----- Validate PDF -----
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".pdf") {
		http.Error(w, "only PDF files are supported", http.StatusBadRequest)
		return
	}

	if len(input) < 5 || string(input[:5]) != "%PDF-" {
		http.Error(w, "uploaded file is not a valid PDF", http.StatusBadRequest)
		return
	}

	// ----- Get conversion parameters -----
	conversionType := strings.TrimSpace(r.FormValue("conversionType"))
	dataType := strings.TrimSpace(r.FormValue("dataType"))
	targetSizeString := strings.TrimSpace(r.FormValue("targetSize"))

	if conversionType == "" {
		http.Error(w, "conversionType is required", http.StatusBadRequest)
		return
	}
	if dataType == "" {
		http.Error(w, "dataType is required", http.StatusBadRequest)
		return
	}
	if targetSizeString == "" {
		http.Error(w, "targetSize is required", http.StatusBadRequest)
		return
	}

	targetSize, err := strconv.ParseFloat(targetSizeString, 64)
	if err != nil {
		http.Error(w, "targetSize must be a valid number", http.StatusBadRequest)
		return
	}
	if targetSize <= 0 {
		http.Error(w, "targetSize must be greater than zero", http.StatusBadRequest)
		return
	}

	conversionType = strings.ToLower(conversionType)
	if conversionType != "compression" && conversionType != "expand" {
		http.Error(
			w,
			"conversionType must be compression or expand",
			http.StatusBadRequest,
		)
		return
	}

	dataType = strings.ToUpper(dataType)
	if dataType != "KB" && dataType != "MB" {
		http.Error(w, "dataType must be KB or MB", http.StatusBadRequest)
		return
	}

	// ----- Convert -----
	result, err := converter.ConvertPDFToTargetSize(
		input,
		conversionType,
		dataType,
		targetSize,
	)
	if err != nil {
		log.Printf("PDF size conversion failed: %v", err)
		http.Error(
			w,
			fmt.Sprintf("PDF conversion failed: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	// ----- Output filename -----
	outputFilename := converter.EnsurePDFOutputPath(
		filepath.Base(fileHeader.Filename),
		conversionType,
	)

	// ----- Response headers -----
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, outputFilename),
	)
	w.Header().Set("Content-Length", strconv.Itoa(len(result.Output)))

	// Conversion metadata
	w.Header().Set(
		"X-Conversion-Target-Met",
		strconv.FormatBool(result.TargetMet),
	)
	w.Header().Set(
		"X-Conversion-Target-Size",
		strconv.FormatInt(result.TargetSize, 10),
	)
	w.Header().Set(
		"X-Conversion-Actual-Size",
		strconv.FormatInt(result.ActualSize, 10),
	)

	// Expose custom headers to browser JS via CORS.
	w.Header().Set(
		"Access-Control-Expose-Headers",
		"X-Conversion-Target-Met, X-Conversion-Target-Size, X-Conversion-Actual-Size",
	)

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(result.Output); err != nil {
		log.Printf("failed to send converted PDF: %v", err)
	}
}