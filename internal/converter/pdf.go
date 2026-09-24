package converter

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

// PDF size units.
const (
	pdfKB = int64(1024)
	pdfMB = int64(1024 * 1024)
)

// Public size-range constants used by the API layer.
const (
	// MinCompressTargetBytes is the smallest target size a user may request
	// for compression.
	MinCompressTargetBytes int64 = 10 * pdfKB // 10 KB

	// MaxCompressTargetBytes is the largest target size a user may request
	// for compression.
	MaxCompressTargetBytes int64 = 400 * pdfMB // 400 MB

	// MinExpandTargetBytes is the smallest target size a user may request
	// for expansion.
	MinExpandTargetBytes int64 = 10 * pdfKB // 10 KB

	// MaxExpandTargetBytes is the largest target size a user may request
	// for expansion.
	MaxExpandTargetBytes int64 = 6000 * pdfKB // 6000 KB
)

// PDFConversionResult carries the outcome of a PDF size conversion.
//
// It is returned even when the requested target could not be reached
// (in which case TargetMet is false and the caller may still deliver
// the smaller file to the user).
type PDFConversionResult struct {
	Output     []byte // The converted PDF bytes.
	TargetSize int64  // The target the user asked for, in bytes.
	ActualSize int64  // The actual size of Output, in bytes.
	TargetMet  bool   // True if ActualSize satisfies the target direction.
}

// rasterQualityStep describes one attempt in the rasterization ladder.
type rasterQualityStep struct {
	DPI           int // Dots-per-inch used when rendering pages.
	JPEGQuality   int // JPEG encoder quality (1..100).
}

// rasterLadder is the ordered list of quality steps tried during
// compression. Higher quality is attempted first; the ladder ends at
// the lowest acceptable quality floor.
//
// The first attempt that meets or beats the target is returned
// immediately, giving the user the best possible quality for their
// requested size.
//
// The ladder is intentionally deep so that multi-page PDFs can still
// approach aggressive targets the way iLovePDF and PI7 do. The floor
// is set at 20 DPI / 3% JPEG quality — below this, the output is
// visually blank and useless.
var rasterLadder = []rasterQualityStep{
	{150, 85},
	{150, 70},
	{120, 60},
	{100, 50},
	{80, 40},
	{72, 35},
	{60, 30},
	{55, 25},
	{50, 20},
	{45, 15},
	{40, 12},
	{35, 10},
	{30, 8},
	{25, 5},
	{20, 3}, // absolute floor — output becomes unusable below this
}

// ConvertPDFToTargetSize converts a PDF toward the requested target size.
//
// conversionType:
//   - "compression" → aggressive rasterization (matches iLovePDF/PI7)
//   - "expand"      → pad the PDF with comments
//
// dataType:
//   - "KB"
//   - "MB"
//
// On success (even best-effort), a *PDFConversionResult is returned.
// A non-nil error indicates an unrecoverable failure (invalid input,
// missing tools, invalid parameters).
func ConvertPDFToTargetSize(
	input []byte,
	conversionType string,
	dataType string,
	targetSize float64,
) (*PDFConversionResult, error) {

	if len(input) == 0 {
		return nil, errors.New("PDF input cannot be empty")
	}

	if !bytes.HasPrefix(input, []byte("%PDF-")) {
		return nil, errors.New("input file is not a valid PDF")
	}

	conversionType = strings.ToLower(strings.TrimSpace(conversionType))
	dataType = strings.ToUpper(strings.TrimSpace(dataType))

	if conversionType != "compression" && conversionType != "expand" {
		return nil, fmt.Errorf(
			"unsupported conversionType %q: use compression or expand",
			conversionType,
		)
	}

	if dataType != "KB" && dataType != "MB" {
		return nil, fmt.Errorf(
			"unsupported dataType %q: use KB or MB",
			dataType,
		)
	}

	if targetSize <= 0 {
		return nil, errors.New("targetSize must be greater than zero")
	}

	targetBytes, err := targetSizeToBytes(targetSize, dataType)
	if err != nil {
		return nil, err
	}

	if err := validateTargetRange(conversionType, targetBytes); err != nil {
		return nil, err
	}

	switch conversionType {
	case "compression":
		return compressPDF(input, targetBytes)
	case "expand":
		return expandPDF(input, targetBytes)
	default:
		return nil, errors.New("unsupported PDF conversion type")
	}
}

// validateTargetRange enforces the min/max target size rules per direction.
func validateTargetRange(conversionType string, targetBytes int64) error {
	switch conversionType {
	case "compression":
		if targetBytes < MinCompressTargetBytes {
			return errors.New("compression target size cannot be less than 10 KB")
		}
		if targetBytes > MaxCompressTargetBytes {
			return errors.New("compression target size cannot exceed 400 MB")
		}
	case "expand":
		if targetBytes < MinExpandTargetBytes {
			return errors.New("expansion target size cannot be less than 10 KB")
		}
		if targetBytes > MaxExpandTargetBytes {
			return errors.New("expansion can take place till 6000kb")
		}
	}
	return nil
}

// targetSizeToBytes converts KB/MB to bytes.
func targetSizeToBytes(targetSize float64, dataType string) (int64, error) {
	var multiplier int64

	switch dataType {
	case "KB":
		multiplier = pdfKB
	case "MB":
		multiplier = pdfMB
	default:
		return 0, fmt.Errorf("unsupported data type: %s", dataType)
	}

	bytesValue := targetSize * float64(multiplier)

	if bytesValue <= 0 {
		return 0, errors.New("target size must be greater than zero")
	}

	if bytesValue > float64(^uint64(0)>>1) {
		return 0, errors.New("target size is too large")
	}

	return int64(bytesValue), nil
}

// ============================================================
// COMPRESSION — rasterization pipeline (iLovePDF/PI7 style)
// ============================================================

// compressPDF converts the input PDF to the smallest possible file
// using page rasterization + JPEG re-encoding, walking down a quality
// ladder until either the target is met or the quality floor is reached.
//
// If the target is never met, the smallest (floor) result is returned
// with TargetMet=false so the caller can still deliver a useful file.
func compressPDF(input []byte, targetBytes int64) (*PDFConversionResult, error) {
	if int64(len(input)) <= targetBytes {
		// Already smaller than or equal to target — nothing to do.
		return &PDFConversionResult{
			Output:     input,
			TargetSize: targetBytes,
			ActualSize: int64(len(input)),
			TargetMet:  true,
		}, nil
	}

	pdftoppmPath, err := exec.LookPath("pdftoppm")
	if err != nil {
		return nil, errors.New(
			"pdftoppm (poppler-utils) is required for PDF compression but was not found in PATH",
		)
	}

	// Write input to a temp file once, reuse for every attempt.
	inputFile, err := os.CreateTemp("", "pdf-compress-in-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp input file: %w", err)
	}
	inputPath := inputFile.Name()
	defer os.Remove(inputPath)

	if _, err := inputFile.Write(input); err != nil {
		inputFile.Close()
		return nil, fmt.Errorf("failed to write temp PDF: %w", err)
	}
	if err := inputFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp PDF: %w", err)
	}

	var smallest *PDFConversionResult

	for _, step := range rasterLadder {
		result, err := rasterizeToPDF(
			pdftoppmPath,
			inputPath,
			step.DPI,
			step.JPEGQuality,
		)
		if err != nil {
			// If a step fails (e.g. pdftoppm errors), try the next.
			continue
		}

		result.TargetSize = targetBytes
		result.ActualSize = int64(len(result.Output))
		result.TargetMet = result.ActualSize <= targetBytes

		// Remember the smallest so far.
		if smallest == nil || result.ActualSize < smallest.ActualSize {
			smallest = result
		}

		// First attempt that meets the target wins (best quality).
		if result.TargetMet {
			return result, nil
		}
	}

	if smallest == nil {
		return nil, errors.New(
			"PDF compression failed: no valid compressed PDF was produced",
		)
	}

	// Best-effort: return the smallest even though target was not met.
	return smallest, nil
}

// rasterizeToPDF renders every page of inputPath as a JPEG at the given
// DPI, re-encodes each JPEG at jpegQuality, and rebuilds a PDF from
// those JPEGs.
func rasterizeToPDF(
	pdftoppmPath string,
	inputPath string,
	dpi int,
	jpegQuality int,
) (*PDFConversionResult, error) {

	tmpDir, err := os.MkdirTemp("", "pdf-raster-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	prefix := filepath.Join(tmpDir, "page")

	// pdftoppm -jpeg -r <dpi> input.pdf <prefix>
	// Produces files like: prefix-1.jpg, prefix-2.jpg, ...
	args := []string{
		"-jpeg",
		"-r", strconv.Itoa(dpi),
		inputPath,
		prefix,
	}

	cmd := exec.Command(pdftoppmPath, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"pdftoppm rendering failed at %d DPI: %w: %s",
			dpi,
			err,
			strings.TrimSpace(stderr.String()),
		)
	}

	// Collect generated page files in order.
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list rendered pages: %w", err)
	}

	var pageFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "page-") && strings.HasSuffix(name, ".jpg") {
			pageFiles = append(pageFiles, filepath.Join(tmpDir, name))
		}
	}

	if len(pageFiles) == 0 {
		return nil, errors.New("pdftoppm produced no pages")
	}

	// Sort lexicographically — pdftoppm pads numbers so this is safe.
	sortStrings(pageFiles)

	// Build a new PDF from the re-encoded JPEGs.
	pdfDoc := fpdf.New("P", "pt", "A4", "")

	for _, pagePath := range pageFiles {
		// Re-encode the JPEG at the requested quality.
		reencodedPath, err := reencodeJPEG(pagePath, jpegQuality)
		if err != nil {
			return nil, fmt.Errorf("failed to re-encode page: %w", err)
		}
		// We must keep the file alive until fpdf writes it out.
		defer os.Remove(reencodedPath)

		// Determine image dimensions to size the PDF page.
		imgWidth, imgHeight, err := jpegDimensions(reencodedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read page dimensions: %w", err)
		}

		// Use points (1/72 inch). Convert pixel dimensions at 72 DPI.
		pageWidthPt := float64(imgWidth) * 72.0 / float64(dpi)
		pageHeightPt := float64(imgHeight) * 72.0 / float64(dpi)

		pdfDoc.AddPageFormat("P", fpdf.SizeType{Wd: pageWidthPt, Ht: pageHeightPt})

		opt := fpdf.ImageOptions{ImageType: "JPG", ReadDpi: false}
		pdfDoc.RegisterImageOptions(reencodedPath, opt)
		pdfDoc.ImageOptions(reencodedPath, 0, 0, pageWidthPt, pageHeightPt, false, opt, 0, "")
	}

	var buf bytes.Buffer
	if err := pdfDoc.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to build compressed PDF: %w", err)
	}

	output := buf.Bytes()

	if !isValidPDF(output) {
		return nil, errors.New("compressed output is not a valid PDF")
	}

	return &PDFConversionResult{
		Output: output,
	}, nil
}

// reencodeJPEG reads a JPEG file, decodes it, re-encodes it at the given
// quality, and writes it to a new temp file. It returns the path to the
// re-encoded JPEG. The caller is responsible for removing the file.
func reencodeJPEG(srcPath string, quality int) (string, error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	img, err := jpeg.Decode(src)
	if err != nil {
		return "", err
	}

	dst, err := os.CreateTemp("", "pdf-page-reenc-*.jpg")
	if err != nil {
		return "", err
	}
	dstPath := dst.Name()

	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}

	if err := jpeg.Encode(dst, img, &jpeg.Options{Quality: quality}); err != nil {
		dst.Close()
		os.Remove(dstPath)
		return "", err
	}
	if err := dst.Close(); err != nil {
		os.Remove(dstPath)
		return "", err
	}

	return dstPath, nil
}

// jpegDimensions returns the pixel width and height of a JPEG file.
func jpegDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, err := jpeg.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}
	return cfg.Width, cfg.Height, nil
}

// sortStrings sorts a slice of strings in ascending order.
// Small helper to avoid importing "sort" for a one-liner.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

// ============================================================
// EXPANSION — pure-Go padding (no external tools)
// ============================================================

// expandPDF increases the PDF size by adding PDF comments before %%EOF.
func expandPDF(input []byte, targetBytes int64) (*PDFConversionResult, error) {
	if int64(len(input)) >= targetBytes {
		return &PDFConversionResult{
			Output:     input,
			TargetSize: targetBytes,
			ActualSize: int64(len(input)),
			TargetMet:  true,
		}, nil
	}

	eofIndex := bytes.LastIndex(input, []byte("%%EOF"))
	if eofIndex == -1 {
		return nil, errors.New("cannot expand PDF: %%EOF marker was not found")
	}

	needed := targetBytes - int64(len(input))
	const commentPrefix = "% universal-converter-padding\n"

	var padding bytes.Buffer
	for int64(padding.Len()) < needed {
		remaining := needed - int64(padding.Len())
		chunkSize := int64(4096)
		if remaining < chunkSize {
			chunkSize = remaining
		}
		if chunkSize <= 0 {
			break
		}
		chunk := bytes.Repeat([]byte{'%'}, int(chunkSize))
		padding.Write(chunk)
		padding.WriteString("\n")
	}
	for int64(padding.Len()) < needed {
		padding.WriteString("%")
	}

	var result bytes.Buffer
	result.Grow(len(input) + padding.Len())
	result.Write(input[:eofIndex])
	if result.Len() > 0 && result.Bytes()[result.Len()-1] != '\n' {
		result.WriteByte('\n')
	}
	result.WriteString(commentPrefix)
	result.Write(padding.Bytes())
	result.Write(input[eofIndex:])

	output := result.Bytes()

	if !isValidPDF(output) {
		return nil, errors.New("expanded output is not a valid PDF")
	}

	if int64(len(output)) < targetBytes {
		extra := targetBytes - int64(len(output))
		var final bytes.Buffer
		final.Grow(len(output) + int(extra))
		final.Write(output[:len(output)-len("%%EOF")])
		final.Write(bytes.Repeat([]byte{'%'}, int(extra)))
		final.WriteString("%%EOF")
		output = final.Bytes()
	}

	return &PDFConversionResult{
		Output:     output,
		TargetSize: targetBytes,
		ActualSize: int64(len(output)),
		TargetMet:  true,
	}, nil
}

// ============================================================
// Helpers
// ============================================================

// isValidPDF performs a basic PDF structural check.
func isValidPDF(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return false
	}
	if !bytes.Contains(data, []byte("%%EOF")) {
		return false
	}
	return true
}

// PDFSizeInBytes converts a human-readable size to bytes.
func PDFSizeInBytes(value float64, unit string) (int64, error) {
	unit = strings.ToUpper(strings.TrimSpace(unit))
	return targetSizeToBytes(value, unit)
}

// FormatPDFSize returns a readable PDF size.
func FormatPDFSize(size int64) string {
	if size >= pdfMB {
		return strconv.FormatFloat(float64(size)/float64(pdfMB), 'f', 2, 64) + " MB"
	}
	if size >= pdfKB {
		return strconv.FormatFloat(float64(size)/float64(pdfKB), 'f', 2, 64) + " KB"
	}
	return strconv.FormatInt(size, 10) + " bytes"
}

// EnsurePDFOutputPath returns a safe output filename.
func EnsurePDFOutputPath(originalName, conversionType string) string {
	base := filepath.Base(originalName)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		name = "converted"
	}
	switch strings.ToLower(conversionType) {
	case "compression":
		return name + "_compressed.pdf"
	case "expand":
		return name + "_expanded.pdf"
	default:
		return name + "_converted.pdf"
	}
}

// Silence unused import warning for io — kept for potential future use.
var _ = io.Discard