package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"file-converter/internal/converter"
)

// HandleExtractText extracts text from PDF and image uploads.
//
// @Summary      Extract text from PDF or image
// @Description  Upload a PDF or image and receive the extracted text as JSON.
// @Description
// @Description  **Supported inputs:**
// @Description  - PDF (digital and scanned -- scanned pages automatically fall back to OCR)
// @Description  - Images: JPG, JPEG, PNG, WEBP, TIFF, BMP
// @Description
// @Description  **Optional parameters:**
// @Description  - lang:   OCR language code (default eng). Examples: eng, hin, eng+hin.
// @Description  - page:   For PDFs only -- extract a single 1-indexed page. Omit for all pages.
// @Description  - format: plain (default) or structured (include per-page array).
// @Description
// @Description  **Limits:**
// @Description  - Upload size must not exceed 50 MB.
// @Description  - PDFs may have at most 500 pages.
// @Tags         Text Extraction
// @Accept       multipart/form-data
// @Produce      json
// @Param        file   formData file   true  "PDF or image file (max 50 MB)"
// @Param        lang   formData string false "OCR language code (default eng)"
// @Param        page   formData int    false "PDF page number to extract (1-indexed)"
// @Param        format formData string false "plain or structured (default plain)"
// @Success      200 {object} converter.ExtractResponse "Extraction result"
// @Failure      400 {string} string "Bad Request - invalid file, parameters, or unsupported type"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /extract-text [post]
func HandleExtractText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Enforce the 50 MB cap before parsing the multipart body.
	r.Body = http.MaxBytesReader(w, r.Body, converter.MaxExtractUploadSize)

	if err := r.ParseMultipartForm(converter.MaxExtractUploadSize); err != nil {
		http.Error(
			w,
			fmt.Sprintf("failed to parse multipart form (max %d MB): %v",
				converter.MaxExtractUploadSize/(1024*1024), err),
			http.StatusBadRequest,
		)
		return
	}

	// --- File ---
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required (form-data key must be 'file')", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read uploaded file: %v", err), http.StatusBadRequest)
		return
	}

	// --- Optional params ---
	req := converter.ExtractRequest{
		Lang:   strings.TrimSpace(r.FormValue("lang")),
		Format: strings.ToLower(strings.TrimSpace(r.FormValue("format"))),
	}

	if pageStr := strings.TrimSpace(r.FormValue("page")); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			http.Error(w, "page must be a positive integer", http.StatusBadRequest)
			return
		}
		req.Page = p
	}

	// --- Run extraction ---
	res, err := converter.ExtractText(fileHeader.Filename, data, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
