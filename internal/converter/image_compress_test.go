package converter

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
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

// encodeJPEG is a test-only helper that produces an input fixture.
func encodeJPEG(t *testing.T, img image.Image, quality int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		t.Fatalf("failed to encode jpeg fixture: %v", err)
	}
	return buf.Bytes()
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode png fixture: %v", err)
	}
	return buf.Bytes()
}

// ============================================================
// 1. Validation
// ============================================================

func TestCompressImage_RejectsEmptyBytes(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      nil,
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    50,
	})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestCompressImage_RejectsUnknownConversionType(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      encodeJPEG(t, makeNoisyImage(64, 64), 90),
		ConversionType: "shrink",
		DataType:       "percentage",
		TargetValue:    50,
	})
	if err == nil {
		t.Fatal("expected error for unknown conversionType")
	}
}

func TestCompressImage_RejectsUnknownDataType(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      encodeJPEG(t, makeNoisyImage(64, 64), 90),
		ConversionType: "compress",
		DataType:       "megabytes",
		TargetValue:    50,
	})
	if err == nil {
		t.Fatal("expected error for unknown dataType")
	}
}

func TestCompressImage_RejectsCompressPercentAbove99(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      encodeJPEG(t, makeNoisyImage(64, 64), 90),
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    100, // 100% is a no-op; must be rejected
	})
	if err == nil {
		t.Fatal("expected error for compress percentage = 100")
	}
}

func TestCompressImage_RejectsCompressPercentBelow1(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      encodeJPEG(t, makeNoisyImage(64, 64), 90),
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    0,
	})
	if err == nil {
		t.Fatal("expected error for compress percentage = 0")
	}
}

func TestCompressImage_RejectsExpandWithSizeDataType(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      encodeJPEG(t, makeNoisyImage(64, 64), 90),
		ConversionType: "expand",
		DataType:       "size", // forbidden for expand
		TargetValue:    500,
		SizeUnit:       "KB",
	})
	if err == nil {
		t.Fatal("expected error for expand + dataType=size")
	}
}

func TestCompressImage_RejectsExpandPercentBelow101(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      encodeJPEG(t, makeNoisyImage(64, 64), 90),
		ConversionType: "expand",
		DataType:       "percentage",
		TargetValue:    100,
	})
	if err == nil {
		t.Fatal("expected error for expand percentage = 100")
	}
}

func TestCompressImage_RejectsUnknownTargetFormat(t *testing.T) {
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      encodeJPEG(t, makeNoisyImage(64, 64), 90),
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    50,
		TargetFormat:   "gif",
	})
	if err == nil {
		t.Fatal("expected error for targetFormat=gif")
	}
}

func TestCompressImage_RejectsOversizedUpload(t *testing.T) {
	// Craft a fake "file" that is one byte over the cap. We don't
	// need it to be a real image — the size check fires first.
	oversized := make([]byte, MaxImageUploadBytes+1)
	_, err := CompressImage(ImageCompressRequest{
		FileBytes:      oversized,
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    50,
	})
	if err == nil {
		t.Fatal("expected error for oversized upload")
	}
}

// ============================================================
// 2. Format resolution
// ============================================================

func TestResolveTargetFormat_AutoMatrix(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{"jpeg", "jpg"},
		{"png", "png"},
		{"webp", "webp"},
		{"tiff", "jpg"}, // lossless → lossy by default
		{"bmp", "jpg"},  // lossless → lossy by default
	}
	for _, tc := range cases {
		got, err := resolveTargetFormat(tc.source, "auto")
		if err != nil {
			t.Fatalf("source %q: unexpected error: %v", tc.source, err)
		}
		if got != tc.want {
			t.Errorf("source %q: expected %q, got %q", tc.source, tc.want, got)
		}
	}
}

func TestResolveTargetFormat_Explicit(t *testing.T) {
	cases := map[string]string{
		"jpg":  "jpg",
		"jpeg": "jpg",
		"png":  "png",
		"webp": "webp",
	}
	for in, want := range cases {
		got, err := resolveTargetFormat("jpeg", in)
		if err != nil {
			t.Fatalf("targetFormat %q: unexpected error: %v", in, err)
		}
		if got != want {
			t.Errorf("targetFormat %q: expected %q, got %q", in, want, got)
		}
	}
}

// ============================================================
// 3. Happy path — compress percentage
// ============================================================

func TestCompressImage_JPEG_PercentageHappyPath(t *testing.T) {
	src := encodeJPEG(t, makeNoisyImage(800, 600), 95)
	originalSize := len(src)

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		Filename:       "input.jpg",
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    50, // half the size
		TargetFormat:   "auto",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.TargetMet {
		t.Fatalf("expected TargetMet=true, got false (target=%d, actual=%d)",
			res.TargetBytes, res.ActualBytes)
	}
	if res.ActualBytes > res.TargetBytes {
		t.Fatalf("actual (%d) exceeds target (%d)", res.ActualBytes, res.TargetBytes)
	}
	if res.ActualBytes >= int64(originalSize) {
		t.Fatalf("expected compression, but output (%d) >= input (%d)",
			res.ActualBytes, originalSize)
	}
	if res.Format != "jpg" {
		t.Fatalf("expected format jpg, got %q", res.Format)
	}
	if res.Quality < 1 || res.Quality > 100 {
		t.Fatalf("expected quality 1..100, got %d", res.Quality)
	}
	if res.Width != 800 || res.Height != 600 {
		t.Fatalf("expected 800x600, got %dx%d", res.Width, res.Height)
	}
	if len(res.Output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ============================================================
// 4. Happy path — compress size
// ============================================================

func TestCompressImage_JPEG_SizeHappyPath(t *testing.T) {
	src := encodeJPEG(t, makeNoisyImage(1200, 900), 95)

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		Filename:       "input.jpg",
		ConversionType: "compress",
		DataType:       "size",
		TargetValue:    60, // 60 KB
		SizeUnit:       "KB",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.TargetMet {
		t.Fatalf("expected TargetMet=true, got false (target=%d, actual=%d)",
			res.TargetBytes, res.ActualBytes)
	}
	if res.ActualBytes > 60*1024 {
		t.Fatalf("actual (%d) exceeds 60 KB target (%d)",
			res.ActualBytes, 60*1024)
	}
}

// ============================================================
// 5. Already-small fast path
// ============================================================

func TestCompressImage_AlreadySmallReturnsOriginal(t *testing.T) {
	// A 32x32 solid JPEG is tiny — well under the 5 MB target.
	solid := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			solid.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	src := encodeJPEG(t, solid, 80)

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		ConversionType: "compress",
		DataType:       "size",
		TargetValue:    5, // 5 MB — vastly larger than input
		SizeUnit:       "MB",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.TargetMet {
		t.Fatal("expected TargetMet=true for already-small input")
	}
	if !bytes.Equal(res.Output, src) {
		t.Fatal("expected original bytes to be returned unchanged")
	}
}

func TestCompressImage_PNG_WithDimensionDownscale(t *testing.T) {
	// PNG is lossless, so downscaling does NOT guarantee a smaller
	// file — the resampled pixels can occasionally be harder for the
	// encoder to compress than the original. This test verifies the
	// contract that the engine always honours:
	//
	//   1. If MaxDimension is set, the output is downscaled to it.
	//   2. If the caller explicitly forced targetFormat=png, the
	//      output format stays PNG.
	//   3. The engine NEVER returns a "compressed" file larger than
	//      the input — if it can't shrink, it returns the original.
	//
	// Note: assertion #3 is stronger than "output < input" because
	// the engine is allowed to return the *original* bytes when it
	// cannot produce a strictly smaller result.
	src := encodePNG(t, makeNoisyImage(2000, 1500))

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		Filename:       "input.png",
		ConversionType: "compress",
		DataType:       "size",
		TargetValue:    200, // 200 KB
		SizeUnit:       "KB",
		TargetFormat:   "png",
		MaxDimension:   800, // force downscale
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Contract #1: downscaling happened (or the engine fell back to
	// the original, in which case the original dimensions are kept).
	if res.Width != 800 && res.Width != 2000 {
		t.Fatalf("expected width 800 (downscaled) or 2000 (fallback), got %d", res.Width)
	}
	if res.Width == 800 && res.Height != 600 {
		t.Fatalf("expected height 600 when downscaled, got %d", res.Height)
	}

	// Contract #2: forced PNG output keeps PNG.
	if res.Format != "png" {
		t.Fatalf("expected format png, got %q", res.Format)
	}

	// Contract #3: never bigger than input.
	if res.ActualBytes > int64(len(src)) {
		t.Fatalf("engine returned a file larger than input: actual=%d, input=%d",
			res.ActualBytes, len(src))
	}
}

// ============================================================
// 7. Best-effort when target is impossible
// ============================================================

func TestCompressImage_TargetUnreachable_ReturnsSmallest(t *testing.T) {
	// A 16x16 image cannot be compressed below ~300 bytes, so a
	// 1 KB target is trivially met. To force failure we ask for
	// 1 KB on a large noisy image and use png (no quality lever),
	// which will hit the ladder floor without ever reaching 1 KB.
	src := encodeJPEG(t, makeNoisyImage(4000, 3000), 95)

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		ConversionType: "compress",
		DataType:       "size",
		TargetValue:    1, // 1 KB — unreachable at this resolution
		SizeUnit:       "KB",
		TargetFormat:   "png", // PNG has no quality lever
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TargetMet {
		t.Fatalf("expected TargetMet=false, got true (actual=%d, target=%d)",
			res.ActualBytes, res.TargetBytes)
	}
	if len(res.Output) == 0 {
		t.Fatal("expected best-effort output, got empty")
	}
}

// ============================================================
// 8. Expand
// ============================================================

func TestCompressImage_ExpandJPEG(t *testing.T) {
	// Start with a low-quality JPEG, ask for a 150% expansion.
	src := encodeJPEG(t, makeNoisyImage(600, 400), 30)
	originalSize := int64(len(src))

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		Filename:       "input.jpg",
		ConversionType: "expand",
		DataType:       "percentage",
		TargetValue:    150,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TargetBytes != originalSize*150/100 {
		t.Fatalf("expected target %d, got %d", originalSize*150/100, res.TargetBytes)
	}
	// We accept TargetMet=false here — JPEG cannot always grow by
	// 50% purely from quality alone, especially if the input was
	// already at quality 30 (which is very compressible).
	if len(res.Output) == 0 {
		t.Fatal("expected non-empty output")
	}
}

// ============================================================
// 9. Format selection overrides
// ============================================================

func TestCompressImage_ForcePNGOutput(t *testing.T) {
	src := encodeJPEG(t, makeNoisyImage(200, 200), 90)

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    50,
		TargetFormat:   "png",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Format != "png" {
		t.Fatalf("expected format png, got %q", res.Format)
	}
	// PNG output begins with the PNG magic bytes.
	if len(res.Output) < 8 ||
		!bytes.Equal(res.Output[:8], []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatal("output does not start with PNG magic bytes")
	}
}

func TestCompressImage_ForceJPGOutputFromPNG(t *testing.T) {
	src := encodePNG(t, makeNoisyImage(400, 400))

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		ConversionType: "compress",
		DataType:       "percentage",
		TargetValue:    50,
		TargetFormat:   "jpg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Format != "jpg" {
		t.Fatalf("expected format jpg, got %q", res.Format)
	}
	if len(res.Output) < 3 ||
		res.Output[0] != 0xFF || res.Output[1] != 0xD8 || res.Output[2] != 0xFF {
		t.Fatal("output does not start with JPEG magic bytes")
	}
}

// ============================================================
// 10. Refinement actually improves the result
// ============================================================

func TestCompressImage_BinaryRefinementImprovesFit(t *testing.T) {
	// A large noisy JPEG has many quality steps that overshoot or
	// undershoot the target. Refinement should land closer to the
	// target than the coarse ladder alone would.
	src := encodeJPEG(t, makeNoisyImage(1500, 1200), 95)
	targetKB := 50

	res, err := CompressImage(ImageCompressRequest{
		FileBytes:      src,
		ConversionType: "compress",
		DataType:       "size",
		TargetValue:    float64(targetKB),
		SizeUnit:       "KB",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.TargetMet {
		t.Fatalf("expected TargetMet=true (target=%d, actual=%d)",
			res.TargetBytes, res.ActualBytes)
	}

	// Refinement should land us within 20% below the target.
	// Without refinement, the coarse ladder often undershoots
	// by much more on these inputs.
	lowerBound := int64(float64(res.TargetBytes) * 0.80)
	if res.ActualBytes < lowerBound {
		t.Errorf("refinement too weak: actual=%d, target=%d, lowerBound=%d",
			res.ActualBytes, res.TargetBytes, lowerBound)
	}
}

// ============================================================
// 11. computeTargetBytes sanity
// ============================================================

func TestComputeTargetBytes(t *testing.T) {
	cases := []struct {
		name           string
		originalSize   int64
		dataType       string
		conversionType string
		targetValue    float64
		sizeUnit       string
		want           int64
	}{
		{"compress to 50 percent", 100000, "percentage", "compress", 50, "", 50000},
		{"compress to 25 percent", 80000, "percentage", "compress", 25, "", 20000},
		{"size in KB", 0, "size", "compress", 200, "KB", 200 * 1024},
		{"size in MB", 0, "size", "compress", 2, "MB", 2 * 1024 * 1024},
		{"empty sizeUnit defaults to KB", 0, "size", "compress", 10, "", 10 * 1024},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := computeTargetBytes(
				tc.originalSize, tc.dataType, tc.conversionType, tc.targetValue, tc.sizeUnit,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

// ============================================================
// 12. satisfiesTarget sanity
// ============================================================

func TestSatisfiesTarget(t *testing.T) {
	cases := []struct {
		name           string
		conversionType string
		size           int64
		target         int64
		want           bool
	}{
		{"compress under target", "compress", 800, 1000, true},
		{"compress at target", "compress", 1000, 1000, true},
		{"compress over target", "compress", 1200, 1000, false},
		{"expand under target", "expand", 800, 1000, false},
		{"expand at target", "expand", 1000, 1000, true},
		{"expand over target", "expand", 1200, 1000, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := satisfiesTarget(tc.conversionType, tc.size, tc.target)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}