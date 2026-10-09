package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"file-converter/internal/converter"
)

// ============================================================
// Image Compress — HTTP handler
//
// Thin shim over converter.CompressImage. Responsibilities:
//   1. Parse multipart form + validate ranges
//   2. Guard against concurrent-encode memory blowups
//   3. Set X-Compress-* response headers
//   4. Stream the encoded bytes
// ============================================================

// imageCompressSemaphore caps concurrent encodes. Tune this based
// on the number of cores available. On a 2-vCPU VPS, use 2; on an
// 8-vCPU box, 4–6 is comfortable.
var imageCompressSemaphore = make(chan struct{}, 4)

// maxImageUploadBytes mirrors converter.MaxImageUploadBytes. Kept
// as a local const so the handler never has to reach across
// packages for a limiter value.
const maxImageUploadBytes = 25 * 1024 * 1024 // 25 MB

// CompressImage handles POST /compress-image.
//
// @Summary      Compress or expand an image
// @Description  Upload an image and re-encode it toward a requested target size or percentage.
// @Description
// @Description  **Supported inputs:** JPG, PNG, WEBP, TIFF, BMP.
// @Description
// @Description  **Conversion types:**
// @Description  - `compress` — quality ladder walks down from 95 until the target is met (or the floor is reached).
// @Description  - `expand`   — quality ladder walks up from 80 toward 100. Only `dataType=percentage` is allowed.
// @Description
// @Description  **Data types:**
// @Description  - `percentage` — `targetValue` is 1–99 for compress, 101–1000 for expand.
// @Description  - `size`       — `targetValue` is a size number in `sizeUnit` (KB or MB). Compress only.
// @Description
// @Description  **Target formats:** `auto` keeps the source (TIFF/BMP fall back to JPG), or force `jpg`, `png`, or `webp`.
// @Description  WebP encoding requires the `cwebp` binary on the server's PATH.
// @Description
// @Description  **Limits:** upload ≤ 25 MB, output ≤ 25 MB, size target 1 KB – 20 MB.
// @Description
// @Description  **Response headers:**
// @Description  - `X-Compress-Target-Met` : `true` if the target was reached, `false` if the smallest achievable size was returned.
// @Description  - `X-Compress-Target-Size`: the requested target in bytes.
// @Description  - `X-Compress-Actual-Size`: the actual output size in bytes.
// @Description  - `X-Compress-Format`     : the output format: `jpg`, `png`, or `webp`.
// @Description  - `X-Compress-Quality`    : the final quality used (1–100), or `0` for PNG.
// @Description  - `X-Compress-Dimensions` : the final output pixel dimensions, e.g. `1920x1080`.
// @Tags         Image Compressor
// @Accept       multipart/form-data
// @Produce      application/octet-stream
// @Param        file           formData file    true  "Image file (max 25 MB)"
// @Param        conversionType formData string  true  "compress or expand"
// @Param        dataType       formData string  true  "percentage or size"
// @Param        targetValue    formData number  true  "Target percentage (1-99 / 101-1000) or size number"
// @Param        sizeUnit       formData string  false "KB or MB (only when dataType=size; default KB)"
// @Param        targetFormat   formData string  false "auto, jpg, png, or webp (default auto)"
// @Param        maxDimension   formData integer false "Cap on the longest edge in pixels (0 = no resize)"
// @Success      200 {file} binary "Compressed image"
// @Failure      400 {string} string "Bad Request - invalid parameters or unsupported format"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /compress-image [post]
func CompressImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// ---- Body cap before parsing ----
	r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadBytes)

	if err := r.ParseMultipartForm(maxImageUploadBytes); err != nil {
		http.Error(
			w,
			fmt.Sprintf("failed to parse multipart form (max %d MB): %v",
				maxImageUploadBytes/(1024*1024), err),
			http.StatusBadRequest,
		)
		return
	}

	// ---- Extract file ----
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required (form-data key must be 'file')", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read uploaded file: %v", err), http.StatusBadRequest)
		return
	}

	// ---- Parse form fields ----
	req, err := parseImageCompressForm(r, fileBytes, fileHeader.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ---- Acquire semaphore slot ----
	imageCompressSemaphore <- struct{}{}
	defer func() { <-imageCompressSemaphore }()

	// ---- Run the engine ----
	result, err := converter.CompressImage(*req)
	if err != nil {
		if errors.Is(err, converter.ErrWebPEncoderUnavailable) {
			writeImageError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeImageError(w, http.StatusBadRequest, err.Error())
		return
	}

	// ---- Response headers ----
	w.Header().Set("Content-Type", mimeForOutputFormat(result.Format))
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(result.Output)), 10))
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="compressed.%s"`, result.Format))

	w.Header().Set("X-Compress-Target-Met", strconv.FormatBool(result.TargetMet))
	w.Header().Set("X-Compress-Target-Size", strconv.FormatInt(result.TargetBytes, 10))
	w.Header().Set("X-Compress-Actual-Size", strconv.FormatInt(result.ActualBytes, 10))
	w.Header().Set("X-Compress-Format", result.Format)
	w.Header().Set("X-Compress-Quality", strconv.Itoa(result.Quality))
	w.Header().Set("X-Compress-Dimensions",
		fmt.Sprintf("%dx%d", result.Width, result.Height))

	// Expose the custom headers to browser JS.
	w.Header().Set(
		"Access-Control-Expose-Headers",
		"X-Compress-Target-Met, X-Compress-Target-Size, X-Compress-Actual-Size, "+
			"X-Compress-Format, X-Compress-Quality, X-Compress-Dimensions",
	)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(result.Output); err != nil {
		log.Printf("[image-compress] failed to write response: %v", err)
	}
}

// ============================================================
// Form parsing
// ============================================================

func parseImageCompressForm(
	r *http.Request,
	fileBytes []byte,
	filename string,
) (*converter.ImageCompressRequest, error) {

	conversionType := strings.TrimSpace(r.FormValue("conversionType"))
	dataType := strings.TrimSpace(r.FormValue("dataType"))
	targetValueStr := strings.TrimSpace(r.FormValue("targetValue"))
	sizeUnit := strings.TrimSpace(r.FormValue("sizeUnit"))
	targetFormat := strings.TrimSpace(r.FormValue("targetFormat"))
	maxDimensionStr := strings.TrimSpace(r.FormValue("maxDimension"))

	if conversionType == "" {
		return nil, errors.New("conversionType is required")
	}
	if dataType == "" {
		return nil, errors.New("dataType is required")
	}
	if targetValueStr == "" {
		return nil, errors.New("targetValue is required")
	}

	targetValue, err := strconv.ParseFloat(targetValueStr, 64)
	if err != nil {
		return nil, errors.New("targetValue must be a valid number")
	}

	maxDim := 0
	if maxDimensionStr != "" {
		maxDim, err = strconv.Atoi(maxDimensionStr)
		if err != nil {
			return nil, errors.New("maxDimension must be an integer")
		}
	}

	return &converter.ImageCompressRequest{
		FileBytes:      fileBytes,
		Filename:       filename,
		ConversionType: conversionType,
		DataType:       dataType,
		TargetValue:    targetValue,
		SizeUnit:       sizeUnit,
		TargetFormat:   targetFormat,
		MaxDimension:   maxDim,
	}, nil
}

// ============================================================
// Small helpers
// ============================================================

func mimeForOutputFormat(format string) string {
	switch format {
	case "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// writeImageError emits a JSON error envelope, matching the shape
// used by the auth endpoints ({"error": "..."}) so the frontend can
// reuse its existing error-unwrapping logic.
func writeImageError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}