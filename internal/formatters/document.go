package formatters

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"file-converter/internal/converter"
)

// document.go ports the PDF / DOCX / TXT conversions from the old
// handler.go switch into the registry. Each Convert closure reads the
// full input into memory (the old switch did the same via the multipart
// parser) and delegates to the existing internal/converter helper.
func init() {

	// ---- PDF → DOCX ----
	Register(Formatter{
		From:       "pdf",
		To:         "docx",
		OutputMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		OutputExt:  ".docx",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			data, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				return ErrEmptyInput
			}
			if !bytes.HasPrefix(data, []byte("%PDF-")) {
				return fmt.Errorf("%w: not a PDF", ErrInvalidInput)
			}
			return converter.ConvertPDFtoDOCX(bytes.NewReader(data), int64(len(data)), out)
		},
	})

	// ---- PDF → TXT ----
	Register(Formatter{
		From:       "pdf",
		To:         "txt",
		OutputMIME: "text/plain; charset=utf-8",
		OutputExt:  ".txt",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			data, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				return ErrEmptyInput
			}
			if !bytes.HasPrefix(data, []byte("%PDF-")) {
				return fmt.Errorf("%w: not a PDF", ErrInvalidInput)
			}
			return converter.ConvertPDFtoTXT(bytes.NewReader(data), int64(len(data)), out)
		},
	})

	// ---- PDF → CSV ----
	Register(Formatter{
		From:       "pdf",
		To:         "csv",
		OutputMIME: "text/csv",
		OutputExt:  ".csv",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			data, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				return ErrEmptyInput
			}
			if !bytes.HasPrefix(data, []byte("%PDF-")) {
				return fmt.Errorf("%w: not a PDF", ErrInvalidInput)
			}
			return converter.ConvertPDFtoCSV(bytes.NewReader(data), int64(len(data)), out)
		},
	})

	// ---- PDF → JSON ----
	Register(Formatter{
		From:       "pdf",
		To:         "json",
		OutputMIME: "application/json",
		OutputExt:  ".json",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			data, err := io.ReadAll(in)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				return ErrEmptyInput
			}
			if !bytes.HasPrefix(data, []byte("%PDF-")) {
				return fmt.Errorf("%w: not a PDF", ErrInvalidInput)
			}
			return converter.ConvertPDFtoJSON(bytes.NewReader(data), int64(len(data)), out)
		},
	})

	// ---- TXT → DOCX ----
	Register(Formatter{
		From:       "txt",
		To:         "docx",
		OutputMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		OutputExt:  ".docx",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertTXTtoDOCX(in, out)
		},
	})

	// ---- TXT → PDF ----
	Register(Formatter{
		From:       "txt",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertTXTtoPDF(in, out)
		},
	})

	// ---- TXT → JSON ----
	Register(Formatter{
		From:       "txt",
		To:         "json",
		OutputMIME: "application/json",
		OutputExt:  ".json",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertTXTtoJSON(in, out)
		},
	})

	// ---- TXT → CSV ----
	Register(Formatter{
		From:       "txt",
		To:         "csv",
		OutputMIME: "text/csv",
		OutputExt:  ".csv",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertTXTtoCSV(in, out)
		},
	})

	// ---- DOCX → CSV ----
	Register(Formatter{
		From:       "docx",
		To:         "csv",
		OutputMIME: "text/csv",
		OutputExt:  ".csv",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertDOCXtoCSV(in, out)
		},
	})

	// ---- DOCX → TXT ----
	Register(Formatter{
		From:       "docx",
		To:         "txt",
		OutputMIME: "text/plain; charset=utf-8",
		OutputExt:  ".txt",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertDOCXtoCSV(in, out)
		},
	})

	// ---- DOCX → JSON ----
	Register(Formatter{
		From:       "docx",
		To:         "json",
		OutputMIME: "application/json",
		OutputExt:  ".json",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertDOCXtoJSON(in, out)
		},
	})

	// ---- DOCX → PDF ----
	Register(Formatter{
		From:       "docx",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertDOCXtoPDF(in, out)
		},
	})
}
