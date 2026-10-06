package converter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ocrTimeout caps a single Tesseract invocation.
const ocrTimeout = 60 * time.Second

// ocrTempExt is the file extension used for the temp image handed to Tesseract.
const ocrTempExt = ".png"

// ============================================================
// Dependency checks
// ============================================================

// TesseractAvailable reports whether the tesseract binary is on PATH.
func TesseractAvailable() bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

// EnsureTesseract returns a helpful error if tesseract is missing.
func EnsureTesseract() error {
	if !TesseractAvailable() {
		return errors.New(
			"OCR engine (tesseract) is not available on this server; " +
				"install tesseract-ocr and the required language pack",
		)
	}
	return nil
}

// CheckExtractDependencies returns the names of any external binaries
// required by /extract-text that are missing from PATH. It is safe to
// call at startup — it never panics and never has side effects.
func CheckExtractDependencies() []string {
	var missing []string

	if _, err := exec.LookPath("tesseract"); err != nil {
		missing = append(missing, "tesseract")
	}
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		missing = append(missing, "pdftoppm")
	}

	return missing
}

// ============================================================
// Tesseract subprocess wrapper
// ============================================================

// runTesseract writes imgBytes to a temp file, invokes tesseract, and
// returns the extracted UTF-8 text. lang is a Tesseract language code
// (e.g. "eng", "eng+hin").
func runTesseract(imgBytes []byte, lang string) (string, error) {
	if err := EnsureTesseract(); err != nil {
		return "", err
	}

	if lang == "" {
		lang = "eng"
	}

	// Write the image to a temp file.
	tmp, err := os.CreateTemp("", "extract-ocr-*"+ocrTempExt)
	if err != nil {
		return "", fmt.Errorf("failed to create temp image: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(imgBytes); err != nil {
		tmp.Close()
		return "", fmt.Errorf("failed to write temp image: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp image: %w", err)
	}

	// tesseract <input> stdout -l <lang>
	ctx, cancel := context.WithTimeout(context.Background(), ocrTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tesseract", tmpPath, "stdout", "-l", lang)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.New("OCR timed out")
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		// Tesseract returns a specific error when a language pack is missing.
		if strings.Contains(msg, "Failed loading language") ||
			strings.Contains(msg, "Error opening data file") {
			return "", fmt.Errorf("OCR language pack %q is not installed", lang)
		}
		return "", fmt.Errorf("OCR failed: %s", msg)
	}

	return stdout.String(), nil
}

// ============================================================
// Small helpers
// ============================================================

// ocrTempDirForImage returns a temp directory for intermediate
// rendered files in the PDF -> image pipeline.
func ocrTempDirForImage() (string, error) {
	return os.MkdirTemp("", "extract-render-*")
}

// cleanExt returns the extension of a path without the dot, lowercased.
func cleanExt(p string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(p)), ".")
}
