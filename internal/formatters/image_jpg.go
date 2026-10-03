package formatters

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"io"
)

func init() {
	// jpg → png
	Register(Formatter{
		From:       "jpg",
		To:         "png",
		OutputMIME: "image/png",
		OutputExt:  ".png",
		Category:   CategoryImage,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			img, format, err := decodeAny(in)
			if err != nil {
				return err
			}
			if format != "jpeg" {
				return fmt.Errorf("%w: expected JPEG, got %s", ErrInvalidInput, format)
			}
			return png.Encode(out, img)
		},
	})

	// jpg → pdf
	Register(Formatter{
		From:       "jpg",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryImage,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			img, format, err := decodeAny(in)
			if err != nil {
				return err
			}
			if format != "jpeg" {
				return fmt.Errorf("%w: expected JPEG, got %s", ErrInvalidInput, format)
			}
			return imageToSinglePagePDF(img, out)
		},
	})

	// jpg → txt (OCR)
	Register(Formatter{
		From:       "jpg",
		To:         "txt",
		OutputMIME: "text/plain; charset=utf-8",
		OutputExt:  ".txt",
		Category:   CategoryImage,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			// Read once to sniff the format, then replay for OCR.
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, in); err != nil {
				return err
			}
			if len(buf.Bytes()) == 0 {
				return ErrEmptyInput
			}
			if sniffImageFormat(buf.Bytes()) != "jpeg" {
				return fmt.Errorf("%w: expected JPEG", ErrInvalidInput)
			}
			return ocrImageToText(ctx, bytes.NewReader(buf.Bytes()), "upload.jpg", "eng", out)
		},
	})
}
