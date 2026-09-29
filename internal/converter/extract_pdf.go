package converter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// extractFromPDF runs the two-tier pipeline:
//  1. Try the pure-Go native text extractor (fast, lossless).
//  2. If a page yields no text (scanned PDF), render that page to an image
//     via pdftoppm and OCR it with Tesseract.
func extractFromPDF(data []byte, req ExtractRequest) (*ExtractResponse, error) {
	reader := bytes.NewReader(data)

	// --- Tier 1: native extraction ---
	pdfReader, err := pdf.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF: %w", err)
	}

	totalPages := pdfReader.NumPage()
	if totalPages == 0 {
		return nil, errors.New("PDF has no pages")
	}
	if totalPages > MaxExtractPages {
		return nil, fmt.Errorf(
			"PDF has %d pages, which exceeds the %d-page limit",
			totalPages, MaxExtractPages,
		)
	}

	// Optional single-page selection.
	pagesToProcess := make([]int, 0, totalPages)
	if req.Page > 0 {
		if req.Page > totalPages {
			return nil, fmt.Errorf(
				"page %d is out of range (document has %d pages)",
				req.Page, totalPages,
			)
		}
		pagesToProcess = append(pagesToProcess, req.Page)
	} else {
		for i := 1; i <= totalPages; i++ {
			pagesToProcess = append(pagesToProcess, i)
		}
	}

	result := &ExtractResponse{PageCount: totalPages}
	var ocrPages []int // pages that had to fall back to OCR

	for _, pageNum := range pagesToProcess {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		text = strings.TrimSpace(text)

		if err != nil || text == "" {
			// Tier 2: OCR this page.
			ocrText, ocrErr := ocrSinglePDFPage(data, pageNum, req.Lang)
			if ocrErr != nil {
				// Record a warning and continue with what we have.
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("page %d: native extraction failed and OCR also failed (%v)", pageNum, ocrErr))
				result.Pages = append(result.Pages, PageExtract{Page: pageNum, Text: ""})
				continue
			}
			result.Pages = append(result.Pages, PageExtract{Page: pageNum, Text: ocrText})
			ocrPages = append(ocrPages, pageNum)
			continue
		}

		result.Pages = append(result.Pages, PageExtract{Page: pageNum, Text: text})
	}

	if len(ocrPages) > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("pages %v appeared to be scanned and were processed with OCR (accuracy may be lower)", ocrPages))
	}

	// If every page came back empty, treat as failure.
	empty := true
	for _, p := range result.Pages {
		if strings.TrimSpace(p.Text) != "" {
			empty = false
			break
		}
	}
	if empty {
		return nil, errors.New("no text could be extracted from this PDF (native and OCR both returned empty)")
	}

	return result, nil
}

// ocrSinglePDFPage renders one page with pdftoppm and OCRs the result.
func ocrSinglePDFPage(pdfBytes []byte, pageNum int, lang string) (string, error) {
	if err := EnsureTesseract(); err != nil {
		return "", err
	}

	pdftoppmPath, err := exec.LookPath("pdftoppm")
	if err != nil {
		return "", errors.New("pdftoppm (poppler-utils) is required for OCR fallback but was not found in PATH")
	}

	// Write the PDF to a temp file.
	tmpPDF, err := os.CreateTemp("", "extract-pdf-*.pdf")
	if err != nil {
		return "", err
	}
	tmpPDFPath := tmpPDF.Name()
	defer os.Remove(tmpPDFPath)

	if _, err := tmpPDF.Write(pdfBytes); err != nil {
		tmpPDF.Close()
		return "", err
	}
	if err := tmpPDF.Close(); err != nil {
		return "", err
	}

	// Render the single page.
	tmpDir, err := ocrTempDirForImage()
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	outPrefix := filepath.Join(tmpDir, "page")

	// pdftoppm -png -r 200 -f <page> -l <page> input.pdf outPrefix
	cmd := exec.Command(
		pdftoppmPath,
		"-png",
		"-r", "200",
		"-f", fmt.Sprintf("%d", pageNum),
		"-l", fmt.Sprintf("%d", pageNum),
		tmpPDFPath,
		outPrefix,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftoppm failed: %s", strings.TrimSpace(stderr.String()))
	}

	// Locate the generated file.
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}
	var imgPath string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "page") {
			imgPath = filepath.Join(tmpDir, e.Name())
			break
		}
	}
	if imgPath == "" {
		return "", errors.New("pdftoppm produced no image")
	}

	f, err := os.Open(imgPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	imgBytes, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return runTesseract(imgBytes, lang)
}