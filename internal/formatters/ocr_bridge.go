package formatters

import (
	"context"
	"fmt"
	"io"

	"file-converter/internal/converter"
)

// ocrImageToText runs the existing ExtractText pipeline on an uploaded
// image and writes the extracted text to `out`. It reuses the OCR
// machinery built for /extract-text so image→txt in /convert behaves
// identically to image uploads on /extract-text.
func ocrImageToText(ctx context.Context, in io.Reader, filename, lang string, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}
	if len(data) == 0 {
		return ErrEmptyInput
	}

	// ExtractText accepts a filename and raw bytes. We pass a synthetic
	// name so its magic-byte sniffer has something to fall back on.
	if filename == "" {
		filename = "upload.img"
	}

	res, err := converter.ExtractText(filename, data, converter.ExtractRequest{
		Lang:   lang,
		Format: "plain",
	})
	if err != nil {
		return err
	}

	_, err = io.WriteString(out, res.Text)
	return err
}
