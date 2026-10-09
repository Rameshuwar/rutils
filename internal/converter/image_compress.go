package converter

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"

	// Extra decoders — required so image.Decode can sniff/parse
	// TIFF, BMP, and WebP inputs.
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// ============================================================
// Image Compress — pure-Go compression engine
//
// Peer of pdf.go's ConvertPDFToTargetSize, but for raster images.
// The algorithm is a quality ladder with an optional binary
// refinement pass, plus an optional dimension downscale step.
//
// This file is deliberately self-contained:
//   - It carries its own copies of sniffImageFormat / decodeAny so
//     the converter package has zero dependency on the formatters
//     package.
//   - It has no HTTP or multipart dependencies.
//   - It does not touch the formatters registry.
// ============================================================

// ---- Public size ranges (mirrored in api/image_handler.go) ----

const (
	// MaxImageUploadBytes caps the uploaded file size. Kept well below
	// the 600 MB PDF cap because raster decode allocates 3–10x the
	// compressed size in RAM.
	MaxImageUploadBytes int64 = 25 * 1024 * 1024 // 25 MB

	// MaxImageOutputBytes caps the encoded output. If a re-encode
	// produced something larger than this (pathological case), we
	// reject rather than return a monster.
	MaxImageOutputBytes int64 = 25 * 1024 * 1024 // 25 MB

	// Percentage bounds.
	MinCompressPercent = 1.0   // below 1% is visually destroyed
	MaxCompressPercent = 99.0  // 100% is a no-op
	MinExpandPercent   = 101.0 // below 101% is not really "expand"
	MaxExpandPercent   = 1000.0

	// Size-target bounds (in bytes).
	MinImageTargetBytes int64 = 1 * 1024         // 1 KB
	MaxImageTargetBytes int64 = 20 * 1024 * 1024 // 20 MB
)

// ---- Request / Result types ----

// ImageCompressRequest carries the validated parameters for an
// image-compression job. Populated by the HTTP layer.
type ImageCompressRequest struct {
	// FileBytes is the raw uploaded image.
	FileBytes []byte

	// Filename is used only for diagnostics and format sniffing
	// (extension fallback). Optional.
	Filename string

	// ConversionType is "compress" or "expand".
	ConversionType string

	// DataType is "percentage" or "size".
	DataType string

	// TargetValue is the percent (1..99 for compress, 101..1000
	// for expand) OR the size number when DataType="size".
	TargetValue float64

	// SizeUnit is "KB" or "MB". Only honoured when DataType="size".
	// Defaults to "KB" when empty.
	SizeUnit string

	// TargetFormat is "auto", "jpg", "png", or "webp".
	// "auto" keeps the source format (with the TIFF/BMP → JPG
	// fallback described in the design doc). Empty defaults to
	// "auto".
	TargetFormat string

	// MaxDimension, when > 0, caps the longest edge of the image in
	// pixels. 0 disables resizing.
	MaxDimension int
}

// ImageCompressResult is the outcome of a compression job.
type ImageCompressResult struct {
	// Output is the encoded image bytes. Always non-empty on
	// success.
	Output []byte

	// TargetMet is true if Output's size satisfies the requested
	// direction (i.e. ≤ target for compress, ≥ target for expand).
	TargetMet bool

	// TargetBytes is the target size the user asked for, in bytes.
	TargetBytes int64

	// ActualBytes is len(Output).
	ActualBytes int64

	// Format is the canonical output format: "jpg", "png", or "webp".
	Format string

	// Quality is the final quality used (1..100). For PNG this is 0
	// because PNG is lossless.
	Quality int

	// Width and Height are the final output pixel dimensions.
	Width  int
	Height int
}

// ---- Quality ladders ----

// imageQualityStep describes one attempt in the compression ladder.
type imageQualityStep struct {
	Quality      int // 1..100; ignored for PNG
	MaxDimension int // longest edge cap; 0 = no resize
}

// compressLadder is walked from highest quality downward. The first
// step whose output is ≤ targetBytes wins. The step just before it
// (the last one that was still > targetBytes) becomes the upper
// bracket for the refinement pass.
//
// The ladder is intentionally deep so that very aggressive targets
// can still be reached without falling off the end.
var compressLadder = []imageQualityStep{
	{95, 0},
	{90, 0},
	{85, 0},
	{80, 0},
	{75, 0},
	{70, 0},
	{65, 0},
	{60, 0},
	{55, 0},
	{50, 0},
	{45, 0},
	{40, 0},
	{35, 0},
	{30, 0},
	{25, 0},
	{20, 0},
	{15, 0},
	{10, 0},
}

// expandLadder is the mirror image for the "expand" direction. Since
// JPEG quality above ~95 saturates, the practical effect of walking
// this ladder is a mild size increase via higher-quality re-encode.
var expandLadder = []imageQualityStep{
	{80, 0},
	{85, 0},
	{90, 0},
	{92, 0},
	{95, 0},
	{97, 0},
	{99, 0},
	{100, 0},
}

// imageLadderResult pairs a step with the actual encoded output it
// produced. Used internally by the search.
type imageLadderResult struct {
	Step   imageQualityStep
	Bytes  int64
	Output []byte
	Format string
	Width  int
	Height int
}

// ---- Public entry point ----

// CompressImage converts an uploaded image toward a target size or
// percentage, re-encoding it at iteratively lower qualities until
// the target is met.
//
// Returns a non-nil *ImageCompressResult on both success and
// best-effort. The only reason to return a non-nil error is:
//   - invalid input (empty file, undecodable image)
//   - invalid parameters (bad conversionType / dataType / range)
//   - missing external dependency (cwebp when targetFormat=webp)
//   - unrecoverable encode failure
func CompressImage(req ImageCompressRequest) (*ImageCompressResult, error) {
	// ---- 1. Validate parameters -------------------------------------
	if err := validateImageCompressRequest(&req); err != nil {
		return nil, err
	}

	if len(req.FileBytes) == 0 {
		return nil, errors.New("uploaded image is empty")
	}
	if int64(len(req.FileBytes)) > MaxImageUploadBytes {
		return nil, fmt.Errorf(
			"uploaded image exceeds the %d MB limit",
			MaxImageUploadBytes/(1024*1024),
		)
	}

	// ---- 2. Sniff source format -------------------------------------
	sourceFormat := sniffImageFormat(req.FileBytes)
	if sourceFormat == "" {
		// Extension fallback for inputs whose magic bytes are
		// inconclusive (rare, but user-friendly).
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(req.Filename)), ".")
		switch ext {
		case "jpg", "jpeg":
			sourceFormat = "jpeg"
		case "png", "webp", "tiff", "tif", "bmp":
			sourceFormat = ext
			if sourceFormat == "tif" {
				sourceFormat = "tiff"
			}
		}
	}
	if sourceFormat == "" {
		return nil, errors.New("unsupported image format (expected jpg, png, webp, tiff, or bmp)")
	}

	// ---- 3. Resolve target format -----------------------------------
	outFormat, err := resolveTargetFormat(sourceFormat, req.TargetFormat)
	if err != nil {
		return nil, err
	}

	// ---- 4. Compute target bytes ------------------------------------
	originalSize := int64(len(req.FileBytes))
	targetBytes, err := computeTargetBytes(
		originalSize,
		req.DataType,
		req.ConversionType,
		req.TargetValue,
		req.SizeUnit,
	)
	if err != nil {
		return nil, err
	}

	// ---- 5. Fast path: already satisfies compress -------------------
	//
	// IMPORTANT: this path is only safe when the caller did NOT ask
	// for a dimension cap or a format change. Otherwise a legitimate
	// "compress this 2000px PNG to 800px" request would short-circuit
	// here and return the untouched original.
	if req.ConversionType == "compress" &&
		originalSize <= targetBytes &&
		req.MaxDimension == 0 &&
		outFormat == canonicalFormat(sourceFormat) {

		w, h := decodeDimensions(req.FileBytes)
		return &ImageCompressResult{
			Output:      req.FileBytes,
			TargetMet:   true,
			TargetBytes: targetBytes,
			ActualBytes: originalSize,
			Format:      canonicalFormat(sourceFormat),
			Quality:     0,
			Width:       w,
			Height:      h,
		}, nil
	}

	// ---- 6. Decode once, reuse across all ladder attempts -----------
	img, decodedFormat, err := decodeAny(bytes.NewReader(req.FileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	_ = decodedFormat // already validated via sniff

	// ---- 7. Optional dimension downscale ----------------------------
	if req.MaxDimension > 0 {
		img = downscaleToMaxDimension(img, req.MaxDimension)
	}

	// ---- 8. Walk the ladder -----------------------------------------
	ladder := compressLadder
	if req.ConversionType == "expand" {
		ladder = expandLadder
	}

	var (
		lowerBracket *imageLadderResult // last step still > target (for compress) or < target (for expand)
		justRight    *imageLadderResult // first step that satisfies the target
		smallest     *imageLadderResult // absolute smallest seen (fallback)
	)

	for _, step := range ladder {
		res, err := encodeImageStep(img, outFormat, step)
		if err != nil {
			// A single step failing should not abort the whole
			// search — try the next rung.
			continue
		}

		if smallest == nil || res.Bytes < smallest.Bytes {
			smallest = res
		}

		if satisfiesTarget(req.ConversionType, res.Bytes, targetBytes) {
			justRight = res
			break
		}

		// Still not satisfied — remember as the bracket.
		lowerBracket = res
	}

	// ---- 9. Binary refinement ---------------------------------------
	// Only meaningful for JPEG/WebP compress, when both brackets
	// exist and the winning quality was above the ladder floor.
	if req.ConversionType == "compress" &&
		outFormat != "png" &&
		justRight != nil &&
		lowerBracket != nil &&
		justRight.Step.Quality > 1 {

		refined := refineImageQuality(
			img, outFormat, targetBytes,
			lowerBracket, justRight,
		)
		if refined != nil && refined.Bytes > justRight.Bytes {
			justRight = refined
		}
	}

	//
	// We resolve this in three steps:
	// ---- 10. Pick the winner ----------------------------------------
	//
	// We resolve this in three steps:
	//
	//   (a) If the ladder produced a result that satisfies the target
	//       (justRight != nil), that's our primary candidate.
	//   (b) Otherwise, take the smallest output the ladder produced
	//       as a best-effort candidate.
	//   (c) Apply the "never bigger than input" guard — but ONLY when
	//       it is safe to do so. See the guard's comment below.
	//
	// Step (c) is what guarantees the compress contract: when the
	// user did NOT force a format change, they always get a file
	// that is ≤ the input size.

	var candidate *imageLadderResult

	if justRight != nil {
		candidate = justRight
	} else if smallest != nil {
		candidate = smallest
	} else {
		return nil, errors.New("image compression failed: no valid output was produced")
	}

	if int64(len(candidate.Output)) > MaxImageOutputBytes {
		return nil, fmt.Errorf(
			"encoded output exceeds the %d MB hard limit",
			MaxImageOutputBytes/(1024*1024),
		)
	}

	// ---- 10a. Conditional "never bigger than input" guard ----------
	//
	// The guard protects the compress contract: "the output is never
	// larger than the input". But it must NOT fire when the user
	// explicitly forced a different output format, because in that
	// case the size is a *consequence* of the format change they
	// asked for — not a compression failure.
	//
	// Example that must be allowed to grow:
	//   JPEG (10 KB) → targetFormat=png → 250 KB PNG
	//
	// Example that must be guarded:
	//   PNG (160 KB) → targetFormat=auto → 216 KB PNG
	//   (same format, re-encode grew it → return original)
	//
	// So the guard only fires when the output format matches the
	// canonical source format. Any explicit format change is honoured
	// even if the file grows.
	sourceCanonical := canonicalFormat(sourceFormat)
	sameFormat := candidate.Format == sourceCanonical

	if req.ConversionType == "compress" &&
		sameFormat &&
		int64(len(candidate.Output)) >= originalSize {

		w, h := decodeDimensions(req.FileBytes)
		return &ImageCompressResult{
			Output:      req.FileBytes,
			TargetMet:   originalSize <= targetBytes,
			TargetBytes: targetBytes,
			ActualBytes: originalSize,
			Format:      sourceCanonical,
			Quality:     0,
			Width:       w,
			Height:      h,
		}, nil
	}

	// ---- 10b. Return the candidate ---------------------------------
	//
	// We reached here because either:
	//   - the candidate is legitimately smaller than the input, OR
	//   - the user forced a format change and we're honouring it
	//     even though the file grew.
	targetMet := satisfiesTarget(req.ConversionType, candidate.Bytes, targetBytes)

	return &ImageCompressResult{
		Output:      candidate.Output,
		TargetMet:   targetMet,
		TargetBytes: targetBytes,
		ActualBytes: candidate.Bytes,
		Format:      candidate.Format,
		Quality:     candidate.Step.Quality,
		Width:       candidate.Width,
		Height:      candidate.Height,
	}, nil
}

// ============================================================
// Validation
// ============================================================

func validateImageCompressRequest(req *ImageCompressRequest) error {
	req.ConversionType = strings.ToLower(strings.TrimSpace(req.ConversionType))
	req.DataType = strings.ToLower(strings.TrimSpace(req.DataType))
	req.TargetFormat = strings.ToLower(strings.TrimSpace(req.TargetFormat))
	req.SizeUnit = strings.ToUpper(strings.TrimSpace(req.SizeUnit))

	if req.ConversionType != "compress" && req.ConversionType != "expand" {
		return errors.New("conversionType must be 'compress' or 'expand'")
	}
	if req.DataType != "percentage" && req.DataType != "size" {
		return errors.New("dataType must be 'percentage' or 'size'")
	}
	if req.TargetFormat == "" {
		req.TargetFormat = "auto"
	}
	switch req.TargetFormat {
	case "auto", "jpg", "png", "webp":
	default:
		return errors.New("targetFormat must be 'auto', 'jpg', 'png', or 'webp'")
	}

	if req.SizeUnit == "" {
		req.SizeUnit = "KB"
	}
	if req.SizeUnit != "KB" && req.SizeUnit != "MB" {
		return errors.New("sizeUnit must be 'KB' or 'MB'")
	}

	if math.IsNaN(req.TargetValue) || math.IsInf(req.TargetValue, 0) {
		return errors.New("targetValue must be a finite number")
	}

	switch req.DataType {
	case "percentage":
		switch req.ConversionType {
		case "compress":
			if req.TargetValue < MinCompressPercent || req.TargetValue > MaxCompressPercent {
				return fmt.Errorf(
					"compress percentage must be between %.0f and %.0f",
					MinCompressPercent, MaxCompressPercent,
				)
			}
		case "expand":
			if req.TargetValue < MinExpandPercent || req.TargetValue > MaxExpandPercent {
				return fmt.Errorf(
					"expand percentage must be between %.0f and %.0f",
					MinExpandPercent, MaxExpandPercent,
				)
			}
		}

	case "size":
		if req.ConversionType == "expand" {
			return errors.New(
				"expand supports dataType='percentage' only — " +
					"a size target cannot be reached by quality alone",
			)
		}
		if req.TargetValue <= 0 {
			return errors.New("targetValue must be greater than zero")
		}
		targetBytes := int64(req.TargetValue * float64(sizeUnitMultiplier(req.SizeUnit)))
		if targetBytes < MinImageTargetBytes {
			return fmt.Errorf(
				"size target cannot be less than %d KB",
				MinImageTargetBytes/1024,
			)
		}
		if targetBytes > MaxImageTargetBytes {
			return fmt.Errorf(
				"size target cannot exceed %d MB",
				MaxImageTargetBytes/(1024*1024),
			)
		}
	}

	if req.MaxDimension < 0 {
		return errors.New("maxDimension cannot be negative")
	}
	if req.MaxDimension > 20000 {
		return errors.New("maxDimension cannot exceed 20000 pixels")
	}

	return nil
}

// ============================================================
// Format resolution
// ============================================================

// resolveTargetFormat picks the output encoder. See the design doc
// for the full decision matrix.
func resolveTargetFormat(source, requested string) (string, error) {
	requested = strings.ToLower(strings.TrimSpace(requested))
	if requested == "" || requested == "auto" {
		switch source {
		case "jpeg":
			return "jpg", nil
		case "png":
			return "png", nil
		case "webp":
			return "webp", nil
		case "tiff", "bmp":
			// Lossless sources → lossy output is the whole point.
			return "jpg", nil
		}
		return "", fmt.Errorf("unsupported source format: %s", source)
	}

	switch requested {
	case "jpg", "jpeg":
		return "jpg", nil
	case "png":
		return "png", nil
	case "webp":
		return "webp", nil
	}
	return "", fmt.Errorf("unsupported targetFormat: %s", requested)
}

// canonicalFormat maps the internal sniffed name to the canonical
// short name we expose in headers.
func canonicalFormat(sniffed string) string {
	switch sniffed {
	case "jpeg":
		return "jpg"
	case "tiff", "tif":
		return "tiff"
	default:
		return sniffed
	}
}

// ============================================================
// Target-size math
// ============================================================

func sizeUnitMultiplier(unit string) int64 {
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "MB":
		return 1024 * 1024
	default:
		return 1024
	}
}

func computeTargetBytes(
	originalSize int64,
	dataType, conversionType string,
	targetValue float64,
	sizeUnit string,
) (int64, error) {
	switch dataType {
	case "percentage":
		ratio := targetValue / 100.0
		return int64(float64(originalSize) * ratio), nil
	case "size":
		return int64(targetValue * float64(sizeUnitMultiplier(sizeUnit))), nil
	}
	return 0, fmt.Errorf("unsupported dataType: %s", dataType)
}

// satisfiesTarget reports whether size meets the directional target.
func satisfiesTarget(conversionType string, size, target int64) bool {
	if conversionType == "expand" {
		return size >= target
	}
	return size <= target
}

// ============================================================
// Dimension downscale
// ============================================================

// downscaleToMaxDimension returns a copy of img whose longest edge
// is ≤ maxDim. If img is already small enough, it is returned
// unchanged. Uses CatmullRom for high-quality resampling.
func downscaleToMaxDimension(img image.Image, maxDim int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxDim && h <= maxDim {
		return img
	}

	var newW, newH int
	if w >= h {
		newW = maxDim
		newH = int(math.Round(float64(h) * float64(maxDim) / float64(w)))
	} else {
		newH = maxDim
		newW = int(math.Round(float64(w) * float64(maxDim) / float64(h)))
	}
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	return dst
}

// ============================================================
// Encode
// ============================================================

// encodeImageStep encodes img in the requested format at the given
// quality step.
func encodeImageStep(img image.Image, format string, step imageQualityStep) (*imageLadderResult, error) {
	var buf bytes.Buffer

	switch format {
	case "jpg":
		q := step.Quality
		if q <= 0 {
			q = 85
		}
		if q > 100 {
			q = 100
		}
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
			return nil, fmt.Errorf("jpeg encode failed: %w", err)
		}

	case "png":
		// PNG is lossless. Quality is meaningless; the compression
		// lever is the dimension downscale that happens upstream.
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("png encode failed: %w", err)
		}

	case "webp":
		// Pure-Go WebP encoding does not exist. Shell out to cwebp.
		out, err := encodeWebPViaCwebp(img, step.Quality)
		if err != nil {
			return nil, err
		}
		buf.Write(out)

	default:
		return nil, fmt.Errorf("unsupported output format: %s", format)
	}

	b := img.Bounds()
	return &imageLadderResult{
		Step:   step,
		Bytes:  int64(buf.Len()),
		Output: buf.Bytes(),
		Format: format,
		Width:  b.Dx(),
		Height: b.Dy(),
	}, nil
}

// encodeWebPViaCwebp is the only place in this file that touches the
// filesystem. It writes a temporary PNG (intermediate), invokes the
// cwebp binary, and reads the result back.
//
// A missing cwebp binary is reported as a specific error so the
// HTTP layer can translate it into a helpful 400.
func encodeWebPViaCwebp(img image.Image, quality int) ([]byte, error) {
	cwebpPath, err := exec.LookPath("cwebp")
	if err != nil {
		return nil, ErrWebPEncoderUnavailable
	}

	// Clamp the quality into the sane 1..100 range before we hand it
	// to cwebp.
	if quality <= 0 {
		quality = 85
	}
	if quality > 100 {
		quality = 100
	}

	// 1. Encode the source image as a temporary PNG (lossless
	//    intermediate so we don't compound JPEG artifacts).
	tmpIn, err := os.CreateTemp("", "img-compress-in-*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp input: %w", err)
	}
	tmpInPath := tmpIn.Name()
	defer os.Remove(tmpInPath)

	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(tmpIn, img); err != nil {
		tmpIn.Close()
		return nil, fmt.Errorf("failed to write intermediate PNG: %w", err)
	}
	if err := tmpIn.Close(); err != nil {
		return nil, fmt.Errorf("failed to close intermediate PNG: %w", err)
	}

	// 2. Invoke cwebp:  cwebp -q <q> input.png -o output.webp
	tmpOut, err := os.CreateTemp("", "img-compress-out-*.webp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp output: %w", err)
	}
	tmpOutPath := tmpOut.Name()
	tmpOut.Close()
	defer os.Remove(tmpOutPath)

	cmd := exec.Command(
		cwebpPath,
		"-quiet",
		"-q", fmt.Sprintf("%d", quality),
		tmpInPath,
		"-o", tmpOutPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("cwebp failed: %s", msg)
	}

	return os.ReadFile(tmpOutPath)
}

// ============================================================
// Binary refinement
// ============================================================

// refineImageQuality bisects the quality range between two
// bracketing ladder steps to find an output as close to target as
// possible without exceeding it.
//
// lowerBracket — a step whose output was > target (for compress)
// justRight    — a step whose output was ≤ target
//
// JPEG output size is monotone in quality for a fixed image, so
// this is a valid bisection.
func refineImageQuality(
	img image.Image,
	format string,
	targetBytes int64,
	lowerBracket, justRight *imageLadderResult,
) *imageLadderResult {

	loQ := justRight.Step.Quality
	hiQ := lowerBracket.Step.Quality
	if loQ >= hiQ {
		return nil // nothing to bisect
	}

	best := justRight

	// 5 iterations shrink the interval by 32×, which is more than
	// enough for 1..100 quality space.
	for i := 0; i < 5; i++ {
		if hiQ-loQ <= 1 {
			break
		}
		mid := (loQ + hiQ) / 2

		step := imageQualityStep{Quality: mid, MaxDimension: justRight.Step.MaxDimension}
		candidate, err := encodeImageStep(img, format, step)
		if err != nil {
			// If a mid-step fails, tighten from the hi side.
			hiQ = mid
			continue
		}

		if candidate.Bytes > targetBytes {
			// Still too big → tighten the upper bound.
			hiQ = mid
			continue
		}

		// Fits → better candidate if larger than current best.
		if candidate.Bytes > best.Bytes {
			best = candidate
		}
		loQ = mid
	}

	return best
}

// ============================================================
// Small helpers
// ============================================================

// decodeDimensions returns the pixel dimensions of an encoded
// image without a full decode. Falls back to (0, 0) on failure.
func decodeDimensions(data []byte) (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// ErrWebPEncoderUnavailable is returned by encodeWebPViaCwebp when
// the cwebp binary is not on PATH. The HTTP layer translates this
// into a 400 with a helpful suggestion to use targetFormat=jpg.
var ErrWebPEncoderUnavailable = errors.New(
	"webp encoding requires the 'cwebp' binary, which was not found in PATH; " +
		"install 'webp' (apt-get install webp) or use targetFormat=jpg",
)

// ============================================================
// Local reimplementations of the two helpers that would otherwise
// be imported from internal/formatters/imageops.go.
//
// Inlining them here keeps the converter package free of any
// dependency on the formatters package.
// ============================================================

// sniffImageFormat returns a short format name ("jpeg", "png", "webp",
// "tiff", "bmp") or "" if the magic bytes are unrecognised.
func sniffImageFormat(data []byte) string {
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "jpeg"
	case len(data) >= 8 && bytes.Equal(data[:8], []byte("\x89PNG\r\n\x1a\n")):
		return "png"
	case len(data) >= 12 &&
		bytes.Equal(data[0:4], []byte("RIFF")) &&
		bytes.Equal(data[8:12], []byte("WEBP")):
		return "webp"
	case len(data) >= 4 &&
		(bytes.Equal(data[:4], []byte("II*\x00")) ||
			bytes.Equal(data[:4], []byte("MM\x00*"))):
		return "tiff"
	case len(data) >= 2 && bytes.Equal(data[:2], []byte("BM")):
		return "bmp"
	}
	return ""
}

// decodeAny reads the full input, sniffs its magic bytes, and returns
// a decoded image together with the sniffed format name.
//
// The blank imports of x/image/bmp, x/image/tiff, and x/image/webp at
// the top of this file are what register those decoders with
// image.Decode.
func decodeAny(in io.Reader) (image.Image, string, error) {
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image input: %w", err)
	}
	if len(data) < 5 {
		return nil, "", errors.New("input file is not valid for the declared source format")
	}

	format := sniffImageFormat(data)
	if format == "" {
		return nil, "", errors.New("input file is not valid for the declared source format")
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode %s image: %w", format, err)
	}
	return img, format, nil
}
