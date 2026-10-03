package formatters

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
)

func init() {
	// bmp → jpg
	Register(Formatter{
		From:       "bmp",
		To:         "jpg",
		OutputMIME: "image/jpeg",
		OutputExt:  ".jpg",
		Category:   CategoryImage,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			img, format, err := decodeAny(in)
			if err != nil {
				return err
			}
			if format != "bmp" {
				return fmt.Errorf("%w: expected BMP, got %s", ErrInvalidInput, format)
			}
			return jpeg.Encode(out, img, &jpeg.Options{Quality: 90})
		},
	})

	// bmp → png
	Register(Formatter{
		From:       "bmp",
		To:         "png",
		OutputMIME: "image/png",
		OutputExt:  ".png",
		Category:   CategoryImage,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			img, format, err := decodeAny(in)
			if err != nil {
				return err
			}
			if format != "bmp" {
				return fmt.Errorf("%w: expected BMP, got %s", ErrInvalidInput, format)
			}
			return png.Encode(out, img)
		},
	})

	// bmp → txt (OCR)
	Register(Formatter{
		From:       "bmp",
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
			if sniffImageFormat(buf.Bytes()) != "bmp" {
				return fmt.Errorf("%w: expected BMP", ErrInvalidInput)
			}
			return ocrImageToText(ctx, bytes.NewReader(buf.Bytes()), "upload.bmp", "eng", out)
		},
	})
}
