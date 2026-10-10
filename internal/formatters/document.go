package formatters

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"file-converter/internal/converter"
)

// document.go registers every non-image, non-data document conversion
// that the /convert endpoint can serve.
//
// Two categories of converter live here:
//
//   1. Native Go converters (fast, in-process). Used for conversions
//      where the input is plain text, CSV, or JSON, or where we are
//      extracting text from a PDF with a working ToUnicode table.
//      These closures ignore the context because the operation always
//      completes in well under a second.
//
//   2. LibreOffice-backed converters (via converter.OfficeConvert).
//      Used for every office-document pair — DOCX, XLSX, PPTX, RTF,
//      ODT, ODS, ODP, HTML — because those formats require a real
//      layout engine to preserve formatting, fonts, images, and
//      pagination. These closures FORWARD the request context into
//      the Ctx variant of the converter, so that:
//        - a client disconnect kills the soffice subprocess, and
//        - the formatter's per-category timeout budget is honoured.
func init() {

	// ======================================================================
	// PDF → DOCX / TXT / CSV / JSON
	// ======================================================================
	//
	// PDF → DOCX routes through LibreOffice's PDF import filter, which
	// reconstructs paragraphs, tables, and images where possible.
	// The three text-extraction pairs (TXT/CSV/JSON) remain on the
	// native Go PDF reader because text extraction does not benefit
	// from layout reconstruction — we only want the characters.

	Register(Formatter{
		From:       "pdf",
		To:         "docx",
		OutputMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		OutputExt:  ".docx",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
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
			return converter.ConvertPDFtoDOCXCtx(ctx, bytes.NewReader(data), int64(len(data)), out)
		},
	})

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

	// ======================================================================
	// TXT → DOCX / PDF / JSON / CSV
	// ======================================================================
	//
	// TXT sources stay on the native converters. The input is plain
	// text, so LibreOffice adds no fidelity — only latency.

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

	// ======================================================================
	// DOCX → PDF / TXT / JSON / CSV
	// ======================================================================
	//
	// These are the pairs that used to go through the lossy
	// DOCX → CSV → TXT → PDF chain. Every one now routes through
	// LibreOffice (via the high-fidelity converter functions), which
	// preserves fonts, styles, tables, images, headers, footers, and
	// pagination exactly as Word would.
	//
	// The one exception is docx → csv, which is a pure text extraction
	// with no layout component — the native XML walker is both faster
	// and sufficient for that specific case.

	Register(Formatter{
		From:       "docx",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertDOCXtoPDFCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "docx",
		To:         "txt",
		OutputMIME: "text/plain; charset=utf-8",
		OutputExt:  ".txt",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertDOCXtoTXTHighFidelityCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "docx",
		To:         "json",
		OutputMIME: "application/json",
		OutputExt:  ".json",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertDOCXtoJSONHighFidelityCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "docx",
		To:         "csv",
		OutputMIME: "text/csv",
		OutputExt:  ".csv",
		Category:   CategoryDocument,
		Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
			// CSV is a flat, text-only format. There is no layout to
			// preserve, so the fast native extractor is the right tool.
			return converter.ConvertDOCXtoCSV(in, out)
		},
	})

	// ======================================================================
	// PPTX → PDF / PNG / JPG
	// ======================================================================
	//
	// New capability. Previously the registry had no PPTX pairs at all.
	// LibreOffice renders each slide as a page with full fidelity:
	// slide masters, vector shapes, SmartArt, and embedded media.

	Register(Formatter{
		From:       "pptx",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertPPTXtoPDFCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "pptx",
		To:         "png",
		OutputMIME: "image/png",
		OutputExt:  ".png",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertPPTXtoPNGCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "pptx",
		To:         "jpg",
		OutputMIME: "image/jpeg",
		OutputExt:  ".jpg",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertPPTXtoJPGCtx(ctx, in, out)
		},
	})

	// ======================================================================
	// RTF / ODT / ODS / ODP → PDF
	// ======================================================================
	//
	// New capabilities. All four are OpenDocument or legacy office
	// formats that LibreOffice reads natively and renders with full
	// fidelity. Each is a one-line wrapper around OfficeConvert.

	Register(Formatter{
		From:       "rtf",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertRTFtoPDFCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "odt",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertODTtoPDFCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "ods",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertODStoPDFCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "odp",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertODPtoPDFCtx(ctx, in, out)
		},
	})

	// ======================================================================
	// HTML → PDF / DOCX
	// ======================================================================
	//
	// New capability. LibreOffice's HTML import filter handles inline
	// CSS, tables, images (from relative or absolute paths), and basic
	// fonts. It does not execute JavaScript, so SPA-generated pages
	// with client-rendered content are not supported — the caller
	// must fetch the fully rendered HTML before sending it here.

	Register(Formatter{
		From:       "html",
		To:         "pdf",
		OutputMIME: "application/pdf",
		OutputExt:  ".pdf",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertHTMLtoPDFCtx(ctx, in, out)
		},
	})

	Register(Formatter{
		From:       "html",
		To:         "docx",
		OutputMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		OutputExt:  ".docx",
		Category:   CategoryDocument,
		Convert: func(ctx context.Context, in io.Reader, out io.Writer) error {
			return converter.ConvertHTMLtoDOCXCtx(ctx, in, out)
		},
	})
}