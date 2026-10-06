package formatters

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"file-converter/internal/converter"
)

// pdf_image.go registers the PDF → image conversions. Both delegate to
// the existing internal/converter helpers so no logic is duplicated.
func init() {

	// pdf → png
	Register(Formatter{
		From:       "pdf",
		To:         "png",
		OutputMIME: "image/png",
		OutputExt:  ".png",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			data, err := io.ReadAll(in)
			if err != nil {
				return fmt.Errorf("failed to read PDF: %w", err)
			}
			if len(data) == 0 {
				return ErrEmptyInput
			}
			if !bytes.HasPrefix(data, []byte("%PDF-")) {
				return fmt.Errorf("%w: input is not a PDF", ErrInvalidInput)
			}
			return converter.ConvertPDFtoPNG(bytes.NewReader(data), out)
		},
	})

	// pdf → jpg
	Register(Formatter{
		From:       "pdf",
		To:         "jpg",
		OutputMIME: "image/jpeg",
		OutputExt:  ".jpg",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			data, err := io.ReadAll(in)
			if err != nil {
				return fmt.Errorf("failed to read PDF: %w", err)
			}
			if len(data) == 0 {
				return ErrEmptyInput
			}
			if !bytes.HasPrefix(data, []byte("%PDF-")) {
				return fmt.Errorf("%w: input is not a PDF", ErrInvalidInput)
			}
			return converter.ConvertPDFtoJPG(bytes.NewReader(data), out)
		},
	})
}
