package formatters

import (
	"context"
	"io"

	"file-converter/internal/converter"
)

// data.go ports the CSV/JSON conversions from the old handler.go switch.
func init() {

	// ---- CSV → JSON ----
	Register(Formatter{
		From:       "csv",
		To:         "json",
		OutputMIME: "application/json",
		OutputExt:  ".json",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertCSVtoJSON(in, out)
		},
	})

	// ---- CSV → PDF ----
	Register(Formatter{
		From:       "csv",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertCSVtoPDF(in, out)
		},
	})

	// ---- CSV → TXT ----
	Register(Formatter{
		From:       "csv",
		To:         "txt",
		OutputMIME: "text/plain; charset=utf-8",
		OutputExt:  ".txt",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertCSVtoTXT(in, out)
		},
	})

	// ---- CSV → DOCX ----
	Register(Formatter{
		From:       "csv",
		To:         "docx",
		OutputMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		OutputExt:  ".docx",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertCSVtoDOCX(in, out)
		},
	})

	// ---- JSON → CSV ----
	Register(Formatter{
		From:       "json",
		To:         "csv",
		OutputMIME: "text/csv",
		OutputExt:  ".csv",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertJSONtoCSV(in, out)
		},
	})

	// ---- JSON → TXT ----
	Register(Formatter{
		From:       "json",
		To:         "txt",
		OutputMIME: "text/plain; charset=utf-8",
		OutputExt:  ".txt",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertJSONtoTXT(in, out)
		},
	})

	// ---- JSON → DOCX ----
	Register(Formatter{
		From:       "json",
		To:         "docx",
		OutputMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		OutputExt:  ".docx",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertJSONtoDOCX(in, out)
		},
	})

	// ---- JSON → PDF ----
	Register(Formatter{
		From:       "json",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryData,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertJSONtoPDF(in, out)
		},
	})
}
