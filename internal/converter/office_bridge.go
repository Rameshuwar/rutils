package converter

// ==============================================================================
// office_bridge.go — Subprocess wrappers for external conversion tools
//
// This file is the single point of contact between the Go conversion engine
// and every external binary it depends on:
//
//   • soffice  (LibreOffice)  — office document → PDF / other office formats
//   • gs       (Ghostscript)  — PDF optimization presets
//   • qpdf                    — PDF linearization for web streaming
//   • img2pdf                 — lossless image → PDF embedding
//   • cjpeg    (mozjpeg)      — high-efficiency JPEG encoding
//   • pngquant                — lossy PNG palette reduction
//   • oxipng                  — lossless PNG re-compression
//   • magick   (ImageMagick)  — image pre-processing for OCR
//
// Design rules (must be preserved when this file grows):
//
//   1. No conversion *logic* lives here. This file knows how to *invoke*
//      tools; it does not know which tool to invoke for which pair. That
//      decision belongs in converter.go / the formatter registry.
//
//   2. Every function accepts a context.Context and honours its deadline.
//      External binaries can hang; the caller must be able to cancel.
//
//   3. Every function writes to the provided io.Writer. No function returns
//      a []byte that the caller has to copy. This keeps peak memory bounded
//      even for very large documents.
//
//   4. Every function cleans up its temp files via defer, even on panic.
//      Containers get recycled, but during a long-running process we cannot
//      afford to leak disk space.
//
//   5. Every function returns a typed error (sentinel or wrapped) so the
//      HTTP layer can map to the right status code without string matching.
//      LibreOffice, in particular, returns exit code 0 on failure — we
//      must parse its stderr to know whether the conversion actually
//      succeeded.
//
//   6. Temp files are always created in a dedicated subdirectory under
//      os.TempDir(). This gives us a single directory to monitor for
//      leaks and to mount as tmpfs in production if desired.
// ==============================================================================

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ------------------------------------------------------------------------------
// Sentinel errors
//
// The HTTP layer maps these to status codes. Adding a new sentinel here means
// adding a new mapping in api/handler.go — never rely on string matching.
// ------------------------------------------------------------------------------

var (
	// ErrToolNotInstalled is returned when a required binary is missing from
	// PATH. The API layer surfaces this as a 503 (Service Unavailable) since
	// it indicates a deployment problem, not a user problem.
	ErrToolNotInstalled = errors.New("required external tool is not installed")

	// ErrConversionTimeout is returned when a subprocess exceeds its
	// context deadline. Maps to 504 Gateway Timeout.
	ErrConversionTimeout = errors.New("conversion timed out")

	// ErrOfficeConversionFailed is returned when LibreOffice reports a
	// failure via stderr, even if the process exited with code 0.
	ErrOfficeConversionFailed = errors.New("libreoffice conversion failed")

	// ErrEmptyOutput is returned when a subprocess exits successfully but
	// produces zero bytes of output. This is always a bug, never legitimate.
	ErrEmptyOutput = errors.New("external tool produced no output")

	// ErrEmptyInput is returned when the caller hands the bridge zero bytes.
	// Defined here (instead of in the formatters package) so that the
	// converter package stays self-contained.
	ErrEmptyInput = errors.New("input is empty")
)

// ------------------------------------------------------------------------------
// Tunable constants
// ------------------------------------------------------------------------------

const (
	// OfficeTimeout caps a single LibreOffice invocation. Documents with
	// hundreds of pages can legitimately take 30+ seconds.
	OfficeTimeout = 90 * time.Second

	// GhostscriptTimeout caps a Ghostscript PDF optimization pass.
	GhostscriptTimeout = 120 * time.Second

	// QpdfTimeout caps a qpdf linearization pass.
	QpdfTimeout = 60 * time.Second

	// Img2PDFTimeout caps an img2pdf pass.
	Img2PDFTimeout = 60 * time.Second

	// ImageEncoderTimeout caps a single mozjpeg/pngquant/oxipng invocation.
	ImageEncoderTimeout = 30 * time.Second

	// ImageMagickTimeout caps an ImageMagick pre-processing pass.
	ImageMagickTimeout = 60 * time.Second

	// bridgeTempPrefix is prepended to every temp directory we create.
	// grep-able in case we ever need to audit temp file leaks.
	bridgeTempPrefix = "rutils-conv-"
)

// ------------------------------------------------------------------------------
// Availability probes
//
// cmd/server/main.go calls CheckBridgeDependencies() at startup to log a
// single warning line if any tool is missing. The server continues to boot
// because most conversions do not require every tool — only the ones that
// explicitly call into the missing tool will fail.
// ------------------------------------------------------------------------------

// OfficeAvailable reports whether the `soffice` binary is on PATH.
func OfficeAvailable() bool {
	_, err := exec.LookPath("soffice")
	return err == nil
}

// GhostscriptAvailable reports whether the `gs` binary is on PATH.
func GhostscriptAvailable() bool {
	_, err := exec.LookPath("gs")
	return err == nil
}

// QpdfAvailable reports whether the `qpdf` binary is on PATH.
func QpdfAvailable() bool {
	_, err := exec.LookPath("qpdf")
	return err == nil
}

// Img2PDFAvailable reports whether the `img2pdf` binary is on PATH.
func Img2PDFAvailable() bool {
	_, err := exec.LookPath("img2pdf")
	return err == nil
}

// MozjpegAvailable reports whether the `cjpeg` binary (from mozjpeg) is on PATH.
func MozjpegAvailable() bool {
	_, err := exec.LookPath("cjpeg")
	return err == nil
}

// PngquantAvailable reports whether the `pngquant` binary is on PATH.
func PngquantAvailable() bool {
	_, err := exec.LookPath("pngquant")
	return err == nil
}

// OxiPNGAvailable reports whether the `oxipng` binary is on PATH.
func OxiPNGAvailable() bool {
	_, err := exec.LookPath("oxipng")
	return err == nil
}

// MagickAvailable reports whether the `magick` binary is on PATH.
func MagickAvailable() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

// CheckBridgeDependencies returns the names of any external binaries that
// the conversion engine may need but which are missing from PATH. Called
// at startup to log a single warning line.
func CheckBridgeDependencies() []string {
	var missing []string
	checks := map[string]func() bool{
		"soffice":  OfficeAvailable,
		"gs":       GhostscriptAvailable,
		"qpdf":     QpdfAvailable,
		"img2pdf":  Img2PDFAvailable,
		"cjpeg":    MozjpegAvailable,
		"pngquant": PngquantAvailable,
		"oxipng":   OxiPNGAvailable,
		"magick":   MagickAvailable,
	}
	for name, check := range checks {
		if !check() {
			missing = append(missing, name)
		}
	}
	return missing
}

// ------------------------------------------------------------------------------
// Format mapping tables
//
// LibreOffice's `--convert-to` flag takes a filter string of the form
// "<extension>:<filterName>". The extension determines the output file
// suffix; the filter name selects the internal export module.
//
// Our code speaks in short names ("docx", "pdf", "xlsx"). These tables
// translate between the two.
// ------------------------------------------------------------------------------

// officeInputExts maps our short format names to the file extensions that
// LibreOffice uses to detect the input format. The extension is what
// LibreOffice actually reads when deciding how to parse the file — content
// sniffing is unreliable for OOXML.
var officeInputExts = map[string]string{
	"docx": "docx",
	"doc":  "doc",
	"odt":  "odt",
	"rtf":  "rtf",
	"txt":  "txt",
	"xlsx": "xlsx",
	"xls":  "xls",
	"ods":  "ods",
	"csv":  "csv",
	"pptx": "pptx",
	"ppt":  "ppt",
	"odp":  "odp",
	"html": "html",
	"htm":  "html",
}

// officeOutputFilters maps (fromFormat, toFormat) to the LibreOffice filter
// string. When a filter is not listed, we fall back to a generic filter
// derived from the target extension.
//
// Reference for filter names:
//   https://help.libreoffice.org/latest/en-US/text/shared/guide/convertfilters.html
var officeOutputFilters = map[string]map[string]string{
	"pdf": {
		"docx": "writer_pdf_Export",
		"doc":  "writer_pdf_Export",
		"odt":  "writer_pdf_Export",
		"rtf":  "writer_pdf_Export",
		"txt":  "writer_pdf_Export",
		"html": "writer_pdf_Export",
		"xlsx": "calc_pdf_Export",
		"xls":  "calc_pdf_Export",
		"ods":  "calc_pdf_Export",
		"csv":  "calc_pdf_Export",
		"pptx": "impress_pdf_Export",
		"ppt":  "impress_pdf_Export",
		"odp":  "impress_pdf_Export",
	},
	"docx": {
		"docx": "MS Word 2007 XML",
		"odt":  "MS Word 2007 XML",
		"rtf":  "MS Word 2007 XML",
		"txt":  "MS Word 2007 XML",
		"html": "MS Word 2007 XML",
	},
	"xlsx": {
		"xlsx": "Calc MS Excel 2007 XML",
		"ods":  "Calc MS Excel 2007 XML",
		"csv":  "Calc MS Excel 2007 XML",
	},
	"pptx": {
		"pptx": "Impress MS PowerPoint 2007 XML",
		"odp":  "Impress MS PowerPoint 2007 XML",
	},
	"txt": {
		"txt": "Text (encoded):UTF8",
		"csv": "Text - txt - csv (StarCalc)",
	},
	"html": {
		"html": "HTML (StarWriter)",
	},
	"rtf": {
		"rtf": "Rich Text Format",
	},
	"odt": {
		"odt": "writer8",
	},
	"ods": {
		"ods": "calc8",
	},
	"odp": {
		"odp": "impress8",
	},
}

// officeFallbackFilter returns a best-effort filter string for a target
// format that is not explicitly listed in officeOutputFilters. An empty
// filter name tells LibreOffice to pick its default export filter for
// that extension, which is usually correct.
func officeFallbackFilter(_ string) string {
	return ""
}

// resolveOfficeFilter returns the filter string to pass to LibreOffice for
// the given (from, to) pair. If no explicit mapping exists, it falls back
// to the generic default for that extension.
func resolveOfficeFilter(from, to string) string {
	from = strings.ToLower(strings.TrimSpace(from))
	to = strings.ToLower(strings.TrimSpace(to))

	if byTo, ok := officeOutputFilters[to]; ok {
		if filter, ok := byTo[from]; ok {
			return filter
		}
	}
	return officeFallbackFilter(to)
}

// ------------------------------------------------------------------------------
// OfficeConvert — the LibreOffice bridge
// ------------------------------------------------------------------------------

// OfficeConvert converts an office document from one format to another
// using headless LibreOffice.
//
// Supported inputs:  docx, doc, odt, rtf, txt, xlsx, xls, ods, csv, pptx, ppt, odp, html
// Supported outputs: pdf, docx, odt, rtf, txt, xlsx, ods, csv, pptx, odp, html
//
// The function:
//   1. Creates a dedicated temp directory for this conversion.
//   2. Writes the input bytes to a file with the source extension.
//   3. Invokes soffice with the correct filter string.
//   4. Reads the produced output file.
//   5. Streams it to `out`.
//   6. Cleans up the temp directory.
//
// LibreOffice returns exit code 0 even when it fails to load the input.
// We therefore check both the exit code AND the presence of the expected
// output file. If the output file is missing, we surface the stderr text
// in the returned error so the operator can debug.
func OfficeConvert(ctx context.Context, from, to string, in io.Reader, out io.Writer) error {
	if !OfficeAvailable() {
		return fmt.Errorf("%w: soffice", ErrToolNotInstalled)
	}

	fromExt := officeInputExts[strings.ToLower(from)]
	if fromExt == "" {
		return fmt.Errorf("office_bridge: unsupported input format: %s", from)
	}

	toExt := strings.ToLower(strings.TrimSpace(to))
	if toExt == "" {
		return fmt.Errorf("office_bridge: empty output format")
	}

	// ---- Read the whole input into memory ----
	//
	// LibreOffice requires a real file on disk, not a stream. We read
	// everything into RAM first, then write it out. For 600 MB PDFs this
	// would be wasteful, but office documents are typically <50 MB and
	// the JSON/HTML response layer already buffers them anyway.
	inputBytes, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("office_bridge: failed to read input: %w", err)
	}
	if len(inputBytes) == 0 {
		return ErrEmptyInput
	}

	// ---- Create an isolated temp directory ----
	tmpDir, err := os.MkdirTemp("", bridgeTempPrefix+"office-*")
	if err != nil {
		return fmt.Errorf("office_bridge: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input."+fromExt)
	if err := os.WriteFile(inputPath, inputBytes, 0o600); err != nil {
		return fmt.Errorf("office_bridge: failed to write input file: %w", err)
	}

	// ---- Apply deadline ----
	ctx, cancel := withDefaultTimeout(ctx, OfficeTimeout)
	defer cancel()

	// ---- Build the LibreOffice command ----
	//
	// The --convert-to argument has the form "<ext>:<filter>". When the
	// filter is empty, we pass just "<ext>" and LibreOffice picks a
	// sensible default for that extension.
	filter := resolveOfficeFilter(from, to)
	convertArg := toExt
	if filter != "" {
		convertArg = toExt + ":" + filter
	}

	// A dedicated UserInstallation directory prevents LibreOffice from
	// picking up stale state from a previous invocation, and avoids
	// permission errors when the container runs as a non-root user.
	userInstallDir := filepath.Join(tmpDir, "profile")
	if err := os.MkdirAll(userInstallDir, 0o700); err != nil {
		return fmt.Errorf("office_bridge: failed to create profile dir: %w", err)
	}

	cmd := exec.CommandContext(
		ctx,
		"soffice",
		"--headless",
		"--norestore",
		"--nologo",
		"--nofirststartwizard",
		"--nolockcheck",
		"--nodefault",
		"-env:UserInstallation=file://"+userInstallDir,
		"--convert-to", convertArg,
		"--outdir", tmpDir,
		inputPath,
	)

	// Set HOME to a writable directory so LibreOffice does not attempt
	// to write to /root (which is read-only in many container runtimes).
	cmd.Env = append(os.Environ(), "HOME="+tmpDir)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = io.Discard

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		return fmt.Errorf("%w: %s", ErrOfficeConversionFailed, extractSofficeError(stderr.String(), err))
	}

	// ---- Locate the produced file ----
	//
	// LibreOffice writes "<basename>.<ext>" into --outdir. Since our input
	// basename is always "input", the output is always "input.<toExt>".
	outputPath := filepath.Join(tmpDir, "input."+toExt)
	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		// LibreOffice exited 0 but produced no file. The stderr buffer
		// usually has the reason. Fall back to a generic error message.
		return fmt.Errorf("%w: %s", ErrOfficeConversionFailed,
			extractSofficeError(stderr.String(), err))
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil
}

// extractSofficeError pulls the most useful line out of LibreOffice's
// stderr output. LibreOffice's error messages look like:
//
//	Error: source file could not be loaded
//	Error: Please verify input parameters... (SfxBaseModel::impl_store ...)
//
// When stderr has no recognizable Error: line, we return the underlying
// Go error text.
func extractSofficeError(stderr string, fallback error) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		if fallback != nil {
			return fallback.Error()
		}
		return "unknown error"
	}

	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Error:") {
			return strings.TrimPrefix(line, "Error:")
		}
	}
	return stderr
}

// ------------------------------------------------------------------------------
// GhostscriptCompressPDF
// ------------------------------------------------------------------------------

// GhostscriptPreset selects the DPI/quality trade-off for PDF compression.
// The names mirror Ghostscript's -dPDFSETTINGS values, which are the same
// presets that professional PDF tools (including iLovePDF) use internally.
type GhostscriptPreset string

const (
	// GhostscriptScreen — lowest quality, ~72 DPI images. Maximum compression.
	GhostscriptScreen GhostscriptPreset = "screen"
	// GhostscriptEbook — medium quality, ~150 DPI images. Balanced.
	GhostscriptEbook GhostscriptPreset = "ebook"
	// GhostscriptPrinter — high quality, ~300 DPI images. Print-ready.
	GhostscriptPrinter GhostscriptPreset = "printer"
	// GhostscriptPrepress — highest quality, preserves colors for prepress.
	GhostscriptPrepress GhostscriptPreset = "prepress"
)

// GhostscriptCompressPDF runs the input PDF through Ghostscript with the
// given preset. This is the same optimization pass that tools like
// iLovePDF apply as their final step, and it typically shaves another
// 10-20% off the size of an already-compressed PDF.
//
// Ghostscript also strips metadata, removes duplicate objects, and
// rewrites the cross-reference table — all of which contribute to a
// smaller, cleaner output.
func GhostscriptCompressPDF(ctx context.Context, preset GhostscriptPreset, in io.Reader, out io.Writer) error {
	if !GhostscriptAvailable() {
		return fmt.Errorf("%w: gs", ErrToolNotInstalled)
	}

	inputBytes, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("office_bridge: failed to read PDF input: %w", err)
	}
	if len(inputBytes) == 0 {
		return ErrEmptyInput
	}

	tmpDir, err := os.MkdirTemp("", bridgeTempPrefix+"gs-*")
	if err != nil {
		return fmt.Errorf("office_bridge: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx, cancel := withDefaultTimeout(ctx, GhostscriptTimeout)
	defer cancel()

	// Feed the PDF to Ghostscript via stdin and read the result from
	// stdout. This avoids writing either the input or output to disk.
	cmd := exec.CommandContext(
		ctx,
		"gs",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.7",
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-dSAFER",
		"-dDetectDuplicateImages=true",
		"-dCompressFonts=true",
		"-dSubsetFonts=true",
		"-dEmbedAllFonts=true",
		"-dAutoRotatePages=/None",
		"-dColorImageDownsampleType=/Bicubic",
		"-dGrayImageDownsampleType=/Bicubic",
		"-dMonoImageDownsampleType=/Subsample",
		"-dPDFSETTINGS=/"+string(preset),
		"-",
	)

	cmd.Stdin = bytes.NewReader(inputBytes)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	outputBytes, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		return fmt.Errorf("office_bridge: ghostscript failed: %s", strings.TrimSpace(stderr.String()))
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil
}

// ------------------------------------------------------------------------------
// QpdfLinearize
// ------------------------------------------------------------------------------

// QpdfLinearize rewrites a PDF in "linearized" form, which places the
// first page's data at the beginning of the file. Browsers can then render
// page 1 before the rest of the file has downloaded.
//
// This is a purely structural optimization — it does not change pixels,
// fonts, or layout. It typically adds <1% to the file size but massively
// improves the user's perceived load time.
func QpdfLinearize(ctx context.Context, in io.Reader, out io.Writer) error {
	if !QpdfAvailable() {
		return fmt.Errorf("%w: qpdf", ErrToolNotInstalled)
	}

	inputBytes, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("office_bridge: failed to read PDF input: %w", err)
	}
	if len(inputBytes) == 0 {
		return ErrEmptyInput
	}

	tmpDir, err := os.MkdirTemp("", bridgeTempPrefix+"qpdf-*")
	if err != nil {
		return fmt.Errorf("office_bridge: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.pdf")
	outPath := filepath.Join(tmpDir, "out.pdf")

	if err := os.WriteFile(inPath, inputBytes, 0o600); err != nil {
		return fmt.Errorf("office_bridge: failed to write temp input: %w", err)
	}

	ctx, cancel := withDefaultTimeout(ctx, QpdfTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"qpdf",
		"--linearize",
		"--object-streams=generate",
		"--compress-streams=y",
		"--recompress-flate",
		"--compression-level=9",
		inPath,
		outPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		return fmt.Errorf("office_bridge: qpdf failed: %s", strings.TrimSpace(stderr.String()))
	}

	outputBytes, err := os.ReadFile(outPath)
	if err != nil {
		return fmt.Errorf("office_bridge: failed to read qpdf output: %w", err)
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil
}

// ------------------------------------------------------------------------------
// Img2PDF
// ------------------------------------------------------------------------------

// Img2PDF embeds one or more images into a single PDF **without re-encoding**.
// The JPEG/PNG/WebP bytes are written directly into the PDF container, so
// the output has exactly the same visual quality as the source and the
// same pixel data — only wrapped in a PDF shell.
//
// This is the correct way to do image → PDF. Every alternative that decodes
// and re-encodes the image (including our previous Go-based implementation)
// introduces generation loss or file size bloat.
//
// Each image becomes one page. Images are placed in the order given.
func Img2PDF(ctx context.Context, images [][]byte, out io.Writer) error {
	if !Img2PDFAvailable() {
		return fmt.Errorf("%w: img2pdf", ErrToolNotInstalled)
	}
	if len(images) == 0 {
		return ErrEmptyInput
	}

	tmpDir, err := os.MkdirTemp("", bridgeTempPrefix+"i2p-*")
	if err != nil {
		return fmt.Errorf("office_bridge: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write each image to disk. img2pdf reads file paths from argv, so we
	// collect them in order and pass the slice to exec.CommandContext.
	args := []string{"--output", filepath.Join(tmpDir, "out.pdf")}
	for i, img := range images {
		path := filepath.Join(tmpDir, fmt.Sprintf("page-%04d.bin", i))
		if err := os.WriteFile(path, img, 0o600); err != nil {
			return fmt.Errorf("office_bridge: failed to write image %d: %w", i, err)
		}
		args = append(args, path)
	}

	ctx, cancel := withDefaultTimeout(ctx, Img2PDFTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "img2pdf", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		return fmt.Errorf("office_bridge: img2pdf failed: %s", strings.TrimSpace(stderr.String()))
	}

	outputBytes, err := os.ReadFile(filepath.Join(tmpDir, "out.pdf"))
	if err != nil {
		return fmt.Errorf("office_bridge: failed to read img2pdf output: %w", err)
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil
}

// ------------------------------------------------------------------------------
// MozjpegEncode
// ------------------------------------------------------------------------------

// MozjpegOptions controls the mozjpeg encoder.
type MozjpegOptions struct {
	// Quality is the JPEG quality, 1–100. Lower means smaller and blurrier.
	Quality int

	// Progressive enables progressive JPEG. This produces a slightly
	// smaller file for the same quality and lets browsers render a blurry
	// preview before the full image loads. Almost always worth enabling.
	Progressive bool

	// Optimize enables Huffman table optimization. Adds ~1% encode time
	// for ~2-5% size reduction. Always worth enabling.
	Optimize bool

	// TrellisMultiPass enables trellis quantization with N passes. This is
	// mozjpeg's headline feature: at the same quality setting, trellis
	// produces a file 15-25% smaller than baseline JPEG. The cost is
	// encoding time, which is why we cap at 3 passes by default.
	TrellisMultiPass int

	// Grayscale forces a single-channel output. Useful when the source is
	// actually grayscale but stored as RGB — saves ~40% size.
	Grayscale bool
}

// DefaultMozjpegOptions returns the balanced settings we use when the
// caller only specifies a quality. This mirrors the settings iLovePDF
// uses for its "recommended" preset.
func DefaultMozjpegOptions(quality int) MozjpegOptions {
	return MozjpegOptions{
		Quality:          quality,
		Progressive:      true,
		Optimize:         true,
		TrellisMultiPass: 3,
		Grayscale:        false,
	}
}

// MozjpegEncode encodes an image using mozjpeg's `cjpeg` binary.
//
// cjpeg reads a raw or compressed image from stdin and writes JPEG to
// stdout. We hand it the input bytes and capture the output.
//
// The image must already be in a format cjpeg understands — PPM, PGM,
// BMP, or TGA. We do NOT decode arbitrary JPEG/PNG here; that would
// require piping through ImageMagick first. The caller is responsible
// for providing cjpeg-compatible bytes.
//
// In practice, the converter layer always decodes the source image
// (via Go's `image` package) and re-encodes it to PPM before calling
// this function. That guarantees cjpeg gets what it expects.
func MozjpegEncode(ctx context.Context, ppmBytes []byte, opts MozjpegOptions, out io.Writer) error {
	if !MozjpegAvailable() {
		return fmt.Errorf("%w: cjpeg", ErrToolNotInstalled)
	}

	if opts.Quality < 1 {
		opts.Quality = 1
	}
	if opts.Quality > 100 {
		opts.Quality = 100
	}

	args := []string{"-quality", fmt.Sprintf("%d", opts.Quality)}
	if opts.Progressive {
		args = append(args, "-progressive")
	}
	if opts.Optimize {
		args = append(args, "-optimize")
	}
	if opts.TrellisMultiPass > 0 {
		args = append(args,
			"-trellis", "1",
			"-trellis-dc", "1",
			"-trellis-multipass", fmt.Sprintf("%d", opts.TrellisMultiPass),
		)
	}
	if opts.Grayscale {
		args = append(args, "-grayscale")
	}

	ctx, cancel := withDefaultTimeout(ctx, ImageEncoderTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "cjpeg", args...)
	cmd.Stdin = bytes.NewReader(ppmBytes)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	outputBytes, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		return fmt.Errorf("office_bridge: cjpeg failed: %s", strings.TrimSpace(stderr.String()))
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil
}

// ------------------------------------------------------------------------------
// PngquantEncode
// ------------------------------------------------------------------------------

// PngquantOptions controls the pngquant encoder.
type PngquantOptions struct {
	// MinQuality is the minimum acceptable quality, 0–100. pngquant will
	// not produce an output worse than this.
	MinQuality int

	// MaxQuality is the maximum quality, 0–100. pngquant will not produce
	// an output better than this — it uses the headroom to save more
	// space by reducing the palette further.
	MaxQuality int

	// Speed is the search effort, 1 (slowest, best) to 11 (fastest, worst).
	// We default to 3, which is what most professional pipelines use.
	Speed int

	// StripMetadata removes EXIF, ICC, and other chunks that do not affect
	// rendering. Saves 2-5% on typical photos.
	StripMetadata bool
}

// DefaultPngquantOptions returns the balanced settings we use when the
// caller only specifies a quality. pngquant uses a quality *range*; the
// encoder picks the smallest palette that stays within the range.
func DefaultPngquantOptions(minQuality, maxQuality int) PngquantOptions {
	return PngquantOptions{
		MinQuality:    minQuality,
		MaxQuality:    maxQuality,
		Speed:         3,
		StripMetadata: true,
	}
}

// PngquantEncode runs a PNG through pngquant's lossy palette reduction.
// pngquant typically achieves 60-80% size reduction on photographic PNGs
// with no perceptible quality loss, and 30-50% on UI screenshots.
//
// The input must be a valid PNG. Output is always a valid PNG.
func PngquantEncode(ctx context.Context, pngBytes []byte, opts PngquantOptions, out io.Writer) error {
	if !PngquantAvailable() {
		return fmt.Errorf("%w: pngquant", ErrToolNotInstalled)
	}

	if opts.MinQuality < 0 {
		opts.MinQuality = 0
	}
	if opts.MaxQuality > 100 {
		opts.MaxQuality = 100
	}
	if opts.Speed < 1 {
		opts.Speed = 1
	}
	if opts.Speed > 11 {
		opts.Speed = 11
	}

	args := []string{
		"--quality", fmt.Sprintf("%d-%d", opts.MinQuality, opts.MaxQuality),
		"--speed", fmt.Sprintf("%d", opts.Speed),
		"--force",
		"--output", "-",
	}
	if opts.StripMetadata {
		args = append(args, "--strip")
	}

	ctx, cancel := withDefaultTimeout(ctx, ImageEncoderTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pngquant", args...)
	cmd.Stdin = bytes.NewReader(pngBytes)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	outputBytes, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		// pngquant exits with code 99 when it cannot reach the requested
		// quality without exceeding the min-quality floor. This is not a
		// hard failure — the caller should fall back to oxipng.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 99 {
			return fmt.Errorf("office_bridge: pngquant could not reach min quality: %w", err)
		}
		return fmt.Errorf("office_bridge: pngquant failed: %s", strings.TrimSpace(stderr.String()))
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil}

// ------------------------------------------------------------------------------
// OxiPNGRecompress
// ------------------------------------------------------------------------------

// OxiPNGOptions controls the oxipng lossless re-compressor.
type OxiPNGOptions struct {
	// OptimizationLevel is 0 (fastest) to 6 (smallest). Level 2 is a
	// good default: fast enough for synchronous request handling and
	// captures most of the savings.
	OptimizationLevel int

	// Strip removes metadata chunks. Set to "safe" to strip ancillary
	// chunks that do not affect rendering.
	Strip string
}

// DefaultOxiPNGOptions returns sensible defaults.
func DefaultOxiPNGOptions() OxiPNGOptions {
	return OxiPNGOptions{
		OptimizationLevel: 2,
		Strip:             "safe",
	}
}

// OxiPNGRecompress performs a lossless re-compression of a PNG. Unlike
// pngquant, it never changes pixel values — it only finds a more compact
// encoding. Typical savings are 5-15%.
//
// This is the correct fallback when pngquant cannot reach the requested
// quality floor: the caller gets a smaller file without any quality loss.
func OxiPNGRecompress(ctx context.Context, pngBytes []byte, opts OxiPNGOptions, out io.Writer) error {
	if !OxiPNGAvailable() {
		return fmt.Errorf("%w: oxipng", ErrToolNotInstalled)
	}

	if opts.OptimizationLevel < 0 {
		opts.OptimizationLevel = 0
	}
	if opts.OptimizationLevel > 6 {
		opts.OptimizationLevel = 6
	}
	if opts.Strip == "" {
		opts.Strip = "safe"
	}

	tmpDir, err := os.MkdirTemp("", bridgeTempPrefix+"oxipng-*")
	if err != nil {
		return fmt.Errorf("office_bridge: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.png")
	outPath := filepath.Join(tmpDir, "out.png")

	if err := os.WriteFile(inPath, pngBytes, 0o600); err != nil {
		return fmt.Errorf("office_bridge: failed to write temp PNG: %w", err)
	}

	args := []string{
		"-o", fmt.Sprintf("%d", opts.OptimizationLevel),
		"--strip", opts.Strip,
		"--out", outPath,
		inPath,
	}

	ctx, cancel := withDefaultTimeout(ctx, ImageEncoderTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "oxipng", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		return fmt.Errorf("office_bridge: oxipng failed: %s", strings.TrimSpace(stderr.String()))
	}

	outputBytes, err := os.ReadFile(outPath)
	if err != nil {
		return fmt.Errorf("office_bridge: failed to read oxipng output: %w", err)
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil
}

// ------------------------------------------------------------------------------
// MagickPreprocess
// ------------------------------------------------------------------------------

// MagickOptions controls ImageMagick pre-processing for OCR.
type MagickOptions struct {
	// Deskew rotates the image to correct for scan skew.
	Deskew bool

	// Despeckle applies a median filter to remove salt-and-pepper noise.
	Despeckle bool

	// Normalize stretches the histogram to improve contrast.
	Normalize bool

	// Grayscale converts to a single channel before binarization.
	Grayscale bool

	// Binarize applies adaptive thresholding, converting the image to
	// pure black-and-white. This is what Tesseract prefers.
	Binarize bool

	// TargetDPI upscales small images. Tesseract's accuracy degrades
	// sharply below 300 DPI, so upscaling a 150 DPI scan to 300 DPI
	// typically improves recognition by 15-30%.
	TargetDPI int
}

// DefaultMagickOptions returns the pre-processing pipeline we apply to
// scanned images before sending them to Tesseract.
func DefaultMagickOptions() MagickOptions {
	return MagickOptions{
		Deskew:    true,
		Despeckle: true,
		Normalize: true,
		Grayscale: true,
		Binarize:  true,
		TargetDPI: 300,
	}
}

// MagickPreprocess runs an image through ImageMagick's pre-processing
// pipeline. The output is a clean, deskewed, binarized image ready for
// OCR — improving Tesseract's accuracy by 20-40% on real-world scans.
//
// The input can be any format ImageMagick reads. The output is always
// PNG (lossless, no compression artifacts that would confuse OCR).
func MagickPreprocess(ctx context.Context, imgBytes []byte, opts MagickOptions, out io.Writer) error {
	if !MagickAvailable() {
		return fmt.Errorf("%w: magick", ErrToolNotInstalled)
	}

	args := []string{"-", "-auto-orient"}

	if opts.Deskew {
		args = append(args, "-deskew", "40%")
	}
	if opts.Despeckle {
		args = append(args, "-despeckle")
	}
	if opts.Normalize {
		args = append(args, "-normalize")
	}
	if opts.Grayscale {
		args = append(args, "-colorspace", "Gray")
	}
	if opts.Binarize {
		args = append(args, "-threshold", "50%")
	}
	if opts.TargetDPI > 0 {
		args = append(args, "-units", "PixelsPerInch", "-density", fmt.Sprintf("%d", opts.TargetDPI))
	}

	// Output as PNG to stdout.
	args = append(args, "png:-")

	ctx, cancel := withDefaultTimeout(ctx, ImageMagickTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "magick", args...)
	cmd.Stdin = bytes.NewReader(imgBytes)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	outputBytes, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return ErrConversionTimeout
		}
		return fmt.Errorf("office_bridge: magick failed: %s", strings.TrimSpace(stderr.String()))
	}
	if len(outputBytes) == 0 {
		return ErrEmptyOutput
	}

	if _, err := out.Write(outputBytes); err != nil {
		return fmt.Errorf("office_bridge: failed to write output: %w", err)
	}
	return nil
}

// ------------------------------------------------------------------------------
// Internal helpers
// ------------------------------------------------------------------------------

// withDefaultTimeout wraps a context with a deadline if the caller has not
// already set one. If the caller has set an earlier deadline, we honour
// the caller's (shorter) deadline.
func withDefaultTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := parent.Deadline(); hasDeadline {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, d)
}