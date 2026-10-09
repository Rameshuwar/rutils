package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// ============================================================
// Helpers
// ============================================================

// makeNoisyImage generates a "photo-like" image that JPEG can actually
// compress. Pure random noise is incompressible — even at quality 10 a
// 1200x900 noise image stays above 100 KB, which makes the compression
// tests flaky.
//
// This fixture combines a smooth diagonal gradient (highly compressible)
// with periodic sharp edges (needs some quality budget). The result
// compresses predictably across quality levels.
func makeNoisyImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Smooth gradient base
			r := uint8((x * 255) / w)
			g := uint8((y * 255) / h)
			b := uint8(((x + y) * 255) / (w + h))

			// Periodic edges — a "grid" pattern that forces the
			// encoder to spend some quality budget, so lowering the
			// quality knob actually shrinks the file.
			if (x/32+y/32)%2 == 0 {
				r = 255 - r
				g = 255 - g
				b = 255 - b
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// buildMultipartForm constructs a request with the given fields and
// the given file bytes under the "file" key.
func buildMultipartForm(
	t *testing.T,
	fields map[string]string,
	fileBytes []byte,
	fileName string,
) (*http.Request, *bytes.Buffer) {
	t.Helper()

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("failed to write field %q: %v", k, err)
		}
	}

	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fw.Write(fileBytes); err != nil {
		t.Fatalf("failed to write file bytes: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/compress-image", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, body
}

// noisyJPEGBytes builds a JPEG-encoded test fixture. It's a local
// re-declaration of the same helper that lives in
// internal/converter/image_compress_test.go, because Go test helpers
// are not visible across packages.
//
// The pattern is a smooth gradient overlaid with periodic sharp
// edges — compressible enough that the encoder has real work to do,
// but not so noisy that JPEG quality becomes meaningless.
func noisyJPEGBytes(t *testing.T, w, h, quality int) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := uint8((x * 255) / w)
			g := uint8((y * 255) / h)
			b := uint8(((x + y) * 255) / (w + h))

			if (x/32+y/32)%2 == 0 {
				r = 255 - r
				g = 255 - g
				b = 255 - b
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("failed to encode jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

// ============================================================
// 1. Method guard
// ============================================================

func TestCompressImageHandler_RejectsGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/compress-image", nil)
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Only POST method is allowed") {
		t.Fatalf("expected method-not-allowed message, got %q", rec.Body.String())
	}
}

func TestCompressImageHandler_RejectsPUT(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/compress-image", strings.NewReader(""))
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

// ============================================================
// 2. Missing fields → 400
// ============================================================

func TestCompressImageHandler_MissingFileField(t *testing.T) {
	// Build a multipart body with only the text fields — no file.
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	_ = w.WriteField("conversionType", "compress")
	_ = w.WriteField("dataType", "percentage")
	_ = w.WriteField("targetValue", "50")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/compress-image", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "file is required") {
		t.Fatalf("expected 'file is required', got %q", rec.Body.String())
	}
}

func TestCompressImageHandler_MissingConversionType(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"dataType":    "percentage",
		"targetValue": "50",
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "conversionType is required") {
		t.Fatalf("expected conversionType-required error, got %q", rec.Body.String())
	}
}

func TestCompressImageHandler_MissingDataType(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"targetValue":    "50",
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "dataType is required") {
		t.Fatalf("expected dataType-required error, got %q", rec.Body.String())
	}
}

func TestCompressImageHandler_MissingTargetValue(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "targetValue is required") {
		t.Fatalf("expected targetValue-required error, got %q", rec.Body.String())
	}
}

func TestCompressImageHandler_NonNumericTargetValue(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
		"targetValue":    "not-a-number",
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "targetValue must be a valid number") {
		t.Fatalf("expected non-numeric error, got %q", rec.Body.String())
	}
}

// ============================================================
// 3. Out-of-range targets → 400
// ============================================================

func TestCompressImageHandler_CompressPercentOutOfRange(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
		"targetValue":    "100", // out of range for compress
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "compress percentage must be between") {
		t.Fatalf("expected range error, got %q", body)
	}
}

func TestCompressImageHandler_ExpandWithSizeDataType(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "expand",
		"dataType":       "size",
		"targetValue":    "500",
		"sizeUnit":       "KB",
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "expand supports dataType='percentage' only") {
		t.Fatalf("expected expand+size error, got %q", rec.Body.String())
	}
}

func TestCompressImageHandler_InvalidTargetFormat(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
		"targetValue":    "50",
		"targetFormat":   "gif",
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "targetFormat must be") {
		t.Fatalf("expected targetFormat error, got %q", rec.Body.String())
	}
}

func TestCompressImageHandler_InvalidSizeUnit(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "size",
		"targetValue":    "200",
		"sizeUnit":       "GB",
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "sizeUnit must be") {
		t.Fatalf("expected sizeUnit error, got %q", rec.Body.String())
	}
}

// ============================================================
// 4. Happy path
// ============================================================

func TestCompressImageHandler_HappyPath_JPEG(t *testing.T) {
	src := noisyJPEGBytes(t, 800, 600, 95)
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
		"targetValue":    "50",
		"targetFormat":   "auto",
	}, src, "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}

	// Header assertions.
	h := rec.Header()
	if h.Get("Content-Type") != "image/jpeg" {
		t.Errorf("expected Content-Type image/jpeg, got %q", h.Get("Content-Type"))
	}
	if h.Get("X-Compress-Target-Met") != "true" {
		t.Errorf("expected Target-Met=true, got %q", h.Get("X-Compress-Target-Met"))
	}
	if h.Get("X-Compress-Format") != "jpg" {
		t.Errorf("expected Format=jpg, got %q", h.Get("X-Compress-Format"))
	}
	if h.Get("X-Compress-Dimensions") != "800x600" {
		t.Errorf("expected 800x600, got %q", h.Get("X-Compress-Dimensions"))
	}

	targetBytes, err := strconv.ParseInt(h.Get("X-Compress-Target-Size"), 10, 64)
	if err != nil {
		t.Fatalf("failed to parse X-Compress-Target-Size: %v", err)
	}
	actualBytes, err := strconv.ParseInt(h.Get("X-Compress-Actual-Size"), 10, 64)
	if err != nil {
		t.Fatalf("failed to parse X-Compress-Actual-Size: %v", err)
	}
	if actualBytes > targetBytes {
		t.Fatalf("actual (%d) exceeds target (%d)", actualBytes, targetBytes)
	}

	// Body assertions.
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty response body")
	}
	if actualBytes != int64(rec.Body.Len()) {
		t.Fatalf("X-Compress-Actual-Size (%d) does not match body length (%d)",
			actualBytes, rec.Body.Len())
	}
}

func TestCompressImageHandler_HappyPath_ForcePNG(t *testing.T) {
	src := noisyJPEGBytes(t, 400, 400, 90)
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
		"targetValue":    "50",
		"targetFormat":   "png",
	}, src, "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("expected image/png, got %q", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("X-Compress-Format") != "png" {
		t.Fatalf("expected Format=png, got %q", rec.Header().Get("X-Compress-Format"))
	}
	// Verify PNG magic bytes.
	if !bytes.HasPrefix(rec.Body.Bytes(), []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatal("response body does not start with PNG magic bytes")
	}
}

// ============================================================
// 5. Error envelope shape
// ============================================================

func TestCompressImageHandler_ErrorIsJSONEnvelope(t *testing.T) {
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
		"targetValue":    "100", // out of range
	}, noisyJPEGBytes(t, 200, 200, 90), "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content-type, got %q", rec.Header().Get("Content-Type"))
	}

	var envelope map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("failed to decode JSON error envelope: %v", err)
	}
	if _, ok := envelope["error"]; !ok {
		t.Fatalf("expected 'error' key, got %v", envelope)
	}
}

// ============================================================
// 6. CORS exposure
// ============================================================

func TestCompressImageHandler_ExposesCustomHeaders(t *testing.T) {
	src := noisyJPEGBytes(t, 300, 300, 90)
	req, _ := buildMultipartForm(t, map[string]string{
		"conversionType": "compress",
		"dataType":       "percentage",
		"targetValue":    "50",
	}, src, "in.jpg")
	rec := httptest.NewRecorder()

	CompressImage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	exposed := rec.Header().Get("Access-Control-Expose-Headers")
	for _, want := range []string{
		"X-Compress-Target-Met",
		"X-Compress-Target-Size",
		"X-Compress-Actual-Size",
		"X-Compress-Format",
		"X-Compress-Quality",
		"X-Compress-Dimensions",
	} {
		if !strings.Contains(exposed, want) {
			t.Errorf("Access-Control-Expose-Headers missing %q. Got: %q", want, exposed)
		}
	}
}
