package formatters

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"
)

func init() {
	// png → jpg
	Register(Formatter{
		From:       "png",
		To:         "jpg",
		OutputMIME: "image/jpeg",
		OutputExt:  ".jpg",
		Category:   CategoryImage,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			img, format, err := decodeAny(in)
			if err != nil {
				return err
			}
			if format != "png" {
				return fmt.Errorf("%w: expected PNG, got %s", ErrInvalidInput, format)
			}
			return jpeg.Encode(out, img, &jpeg.Options{Quality: 90})
		},
	})

	// png → pdf
	Register(Formatter{
		From:       "png",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryImage,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			img, format, err := decodeAny(in)
			if err != nil {
				return err
			}
			if format != "png" {
				return fmt.Errorf("%w: expected PNG, got %s", ErrInvalidInput, format)
			}
			return imageToSinglePagePDF(img, out)
		},
	})

	// png → txt (OCR)
	Register(Formatter{
		From:       "png",
		To:         "txt",
		OutputMIME: "text/plain; charset=utf-8",
		OutputExt:  ".txt",
		Category:   CategoryImage,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, in); err != nil {
				return err
			}
			if len(buf.Bytes()) == 0 {
				return ErrEmptyInput
			}
			if sniffImageFormat(buf.Bytes()) != "png" {
				return fmt.Errorf("%w: expected PNG", ErrInvalidInput)
			}
			return ocrImageToText(ctx, bytes.NewReader(buf.Bytes()), "upload.png", "eng", out)
		},
	})
}
