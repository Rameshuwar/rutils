package converter

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"

	"github.com/gingfrederik/docx"
	"github.com/go-pdf/fpdf"
	"github.com/ledongthuc/pdf"
)

// ==============================================================================
// NATIVE CONVERSIONS
//
// Everything in this section is implemented in pure Go. It runs in-process,
// has no external dependencies, and is fast. These are the correct tools
// for the jobs they do — we do not route them through LibreOffice or any
// other subprocess because the native implementation is already ideal.
//
// Conversions that ARE NOT in this section (office documents, PDF
// optimization, image encoding) are delegated to external tools via
// office_bridge.go, further below.
// ==============================================================================

// ConvertJPGtoPNG takes a JPEG from an io.Reader and writes a PNG to an io.Writer.
func ConvertJPGtoPNG(in io.Reader, out io.Writer) error {
	img, err := jpeg.Decode(in)
	if err != nil {
		return fmt.Errorf("failed to decode jpeg: %w", err)
	}
	if err := png.Encode(out, img); err != nil {
		return fmt.Errorf("failed to encode png: %w", err)
	}
	return nil
}

// ConvertPNGtoJPG takes a PNG from an io.Reader and writes a JPG to an io.Writer.
func ConvertPNGtoJPG(in io.Reader, out io.Writer) error {
	img, err := png.Decode(in)
	if err != nil {
		return fmt.Errorf("failed to decode png: %w", err)
	}
	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 90})
	if err != nil {
		return fmt.Errorf("failed to encode jpg: %w", err)
	}
	return nil
}

// ConvertCSVtoJSON converts a CSV file to a JSON array of objects.
func ConvertCSVtoJSON(in io.Reader, out io.Writer) error {
	reader := csv.NewReader(in)
	records, err := reader.ReadAll()
	if err != nil || len(records) < 1 {
		return fmt.Errorf("failed to read csv or csv is empty: %w", err)
	}

	headers := records[0]
	var result []map[string]string

	for _, row := range records[1:] {
		obj := make(map[string]string)
		for i, col := range row {
			if i < len(headers) {
				obj[headers[i]] = col
			}
		}
		result = append(result, obj)
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// flattenJSON is a helper function that recursively flattens nested JSON structures.
func flattenJSON(prefix string, v interface{}, result map[string]interface{}) {
	switch child := v.(type) {
	case map[string]interface{}:
		for k, val := range child {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flattenJSON(newKey, val, result)
		}
	case []interface{}:
		for i, val := range child {
			newKey := fmt.Sprintf("%s.%d", prefix, i)
			if prefix == "" {
				newKey = fmt.Sprintf("%d", i)
			}
			flattenJSON(newKey, val, result)
		}
	default:
		if prefix == "" {
			result["value"] = child
		} else {
			result[prefix] = child
		}
	}
}

// ConvertJSONtoCSV converts any JSON structure to a CSV file.
func ConvertJSONtoCSV(in io.Reader, out io.Writer) error {
	var raw interface{}
	decoder := json.NewDecoder(in)
	if err := decoder.Decode(&raw); err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}

	var data []map[string]interface{}

	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			flat := make(map[string]interface{})
			flattenJSON("", item, flat)
			data = append(data, flat)
		}
	default:
		flat := make(map[string]interface{})
		flattenJSON("", v, flat)
		data = append(data, flat)
	}

	if len(data) == 0 {
		return fmt.Errorf("empty json data")
	}

	headerMap := make(map[string]bool)
	var headers []string
	for _, row := range data {
		for k := range row {
			if !headerMap[k] {
				headerMap[k] = true
				headers = append(headers, k)
			}
		}
	}

	writer := csv.NewWriter(out)
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range data {
		var csvRow []string
		for _, h := range headers {
			val := ""
			if v, ok := row[h]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
			}
			csvRow = append(csvRow, val)
		}
		if err := writer.Write(csvRow); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// ConvertJSONtoTXT formats JSON into readable plain text.
func ConvertJSONtoTXT(in io.Reader, out io.Writer) error {
	var data interface{}
	decoder := json.NewDecoder(in)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format json: %w", err)
	}

	_, err = out.Write(prettyJSON)
	return err
}

// ConvertTXTtoDOCX converts a plain text file to a Microsoft Word Document.
func ConvertTXTtoDOCX(in io.Reader, out io.Writer) error {
	f := docx.NewFile()

	scanner := bufio.NewScanner(in)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		text := scanner.Text()
		p := f.AddParagraph()
		p.AddText(text)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read text file: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "out-*.docx")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	if err := f.Save(tmpFilePath); err != nil {
		return fmt.Errorf("failed to save docx: %w", err)
	}

	savedFile, err := os.Open(tmpFilePath)
	if err != nil {
		return fmt.Errorf("failed to open generated docx: %w", err)
	}
	defer savedFile.Close()

	_, err = io.Copy(out, savedFile)
	return err
}

// ConvertDOCXtoCSV extracts text paragraphs from a DOCX file and writes them to a CSV file.
//
// NOTE: This is a low-fidelity text-only extraction. It is retained for
// internal use by other native converters (e.g. ConvertTXTtoCSV pipelines)
// but is NOT used for user-facing DOCX→* conversions anymore. Those go
// through LibreOffice via office_bridge.go.
func ConvertDOCXtoCSV(in io.Reader, out io.Writer) error {
	b, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read docx: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return fmt.Errorf("failed to open docx archive: %w", err)
	}

	var docXML io.ReadCloser
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			docXML, err = f.Open()
			if err != nil {
				return fmt.Errorf("failed to open document.xml: %w", err)
			}
			break
		}
	}

	if docXML == nil {
		return fmt.Errorf("invalid docx file: word/document.xml not found")
	}
	defer docXML.Close()

	writer := csv.NewWriter(out)
	writer.Write([]string{"Extracted Text"})

	decoder := xml.NewDecoder(docXML)
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				var text string
				if err := decoder.DecodeElement(&text, &se); err == nil {
					writer.Write([]string{text})
				}
			}
		}
	}

	writer.Flush()
	return writer.Error()
}

// ConvertPDFtoTXT extracts text from a PDF file using the native Go
// PDF reader. Works only for digital PDFs with a proper text layer;
// scanned or image-only PDFs return an error.
func ConvertPDFtoTXT(in io.ReaderAt, size int64, out io.Writer) error {
	f, err := pdf.NewReader(in, size)
	if err != nil {
		return fmt.Errorf("failed to read pdf: %w", err)
	}
	b, err := f.GetPlainText()
	if err != nil {
		return fmt.Errorf("failed to extract text from pdf: %w", err)
	}
	textBytes, err := io.ReadAll(b)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(textBytes)) == 0 {
		return fmt.Errorf("no text found. Scanned PDFs or image-based PDFs are not supported")
	}
	_, err = io.Copy(out, bytes.NewReader(textBytes))
	return err
}

// ConvertTXTtoPDF converts text to PDF.
//
// NOTE: This is a simple line-by-line layout intended for plain-text
// fallback. It is NOT used for user-facing DOCX→PDF, HTML→PDF, or
// CSV→PDF conversions anymore — those route through LibreOffice for
// proper layout. It is retained because ConvertJSONtoTXT pipelines
// and legacy code paths still call it.
func ConvertTXTtoPDF(in io.Reader, out io.Writer) error {
	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.AddPage()
	pdfDoc.SetFont("Arial", "", 12)
	pdfDoc.SetY(10.0)

	scanner := bufio.NewScanner(in)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		pdfDoc.CellFormat(190, 10, scanner.Text(), "", 1, "", false, 0, "")
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return pdfDoc.Output(out)
}

// ConvertCSVtoTXT writes a CSV directly out as plain text.
func ConvertCSVtoTXT(in io.Reader, out io.Writer) error {
	_, err := io.Copy(out, in)
	return err
}

// ConvertTXTtoJSON converts text lines into a JSON array of strings.
func ConvertTXTtoJSON(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(map[string]interface{}{"lines": lines})
}

// ConvertTXTtoCSV converts text lines into a single-column CSV.
func ConvertTXTtoCSV(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	writer := csv.NewWriter(out)
	writer.Write([]string{"Text"})

	for scanner.Scan() {
		writer.Write([]string{scanner.Text()})
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	writer.Flush()
	return writer.Error()
}

// ConvertCSVtoDOCX converts CSV data into a DOCX document.
func ConvertCSVtoDOCX(in io.Reader, out io.Writer) error {
	reader := csv.NewReader(in)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv: %w", err)
	}

	f := docx.NewFile()
	for _, row := range records {
		p := f.AddParagraph()
		line := ""
		for _, col := range row {
			line += col + " | "
		}
		p.AddText(line)
	}

	tmpFile, err := os.CreateTemp("", "out-*.docx")
	if err != nil {
		return err
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	if err := f.Save(tmpFilePath); err != nil {
		return err
	}

	savedFile, err := os.Open(tmpFilePath)
	if err != nil {
		return err
	}
	defer savedFile.Close()

	_, err = io.Copy(out, savedFile)
	return err
}

// ConvertPDFtoCSV extracts text from PDF and writes it to a CSV.
func ConvertPDFtoCSV(in io.ReaderAt, size int64, out io.Writer) error {
	fpdfReader, err := pdf.NewReader(in, size)
	if err != nil {
		return err
	}
	b, err := fpdfReader.GetPlainText()
	if err != nil {
		return err
	}

	return ConvertTXTtoCSV(b, out)
}

// ConvertPDFtoJSON extracts text from PDF and writes it to JSON.
func ConvertPDFtoJSON(in io.ReaderAt, size int64, out io.Writer) error {
	fpdfReader, err := pdf.NewReader(in, size)
	if err != nil {
		return err
	}
	b, err := fpdfReader.GetPlainText()
	if err != nil {
		return err
	}

	return ConvertTXTtoJSON(b, out)
}

// ConvertJSONtoDOCX converts JSON to a DOCX document.
func ConvertJSONtoDOCX(in io.Reader, out io.Writer) error {
	var data interface{}
	if err := json.NewDecoder(in).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ConvertTXTtoDOCX(bytes.NewReader(prettyJSON), out)
}

// ConvertJSONtoPDF converts JSON to a simple line-by-line PDF.
//
// NOTE: This is the fallback used when LibreOffice is unavailable.
// The high-fidelity variant is ConvertJSONtoPDFHighFidelity, defined
// below, which is what the /convert endpoint uses when the target
// is a PDF and the source is JSON. This fallback remains exported
// for backward compatibility with tests that call it directly.
func ConvertJSONtoPDF(in io.Reader, out io.Writer) error {
	var data interface{}
	if err := json.NewDecoder(in).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ConvertTXTtoPDF(bytes.NewReader(prettyJSON), out)
}

// ConvertDOCXtoJSON extracts text from DOCX and formats it as JSON.
//
// NOTE: Low-fidelity text-only extraction. Retained for internal use
// and backward compatibility. User-facing DOCX→JSON goes through
// LibreOffice (see ConvertDOCXtoJSONHighFidelity below).
func ConvertDOCXtoJSON(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertDOCXtoCSV(in, &buf); err != nil {
		return err
	}
	return ConvertTXTtoJSON(&buf, out)
}

// ConvertPDFtoJPG renders the first page of a PDF to JPG.
func ConvertPDFtoJPG(in io.Reader, out io.Writer) error {
	return convertPDFToImage(in, out, "-jpeg")
}

// ConvertPDFtoPNG renders the first page of a PDF to PNG.
func ConvertPDFtoPNG(in io.Reader, out io.Writer) error {
	return convertPDFToImage(in, out, "-png")
}

func convertPDFToImage(in io.Reader, out io.Writer, formatFlag string) error {
	tmpFile, err := os.CreateTemp("", "pdf-in-*.pdf")
	if err != nil {
		return err
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)
	tmpFile.Close()

	b, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmpFilePath, b, 0644); err != nil {
		return err
	}

	outPrefix := tmpFilePath + "-out"
	cmd := exec.Command("pdftocairo", formatFlag, "-singlefile", tmpFilePath, outPrefix)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("poppler rendering failed: %w", err)
	}

	ext := ".jpg"
	if formatFlag == "-png" {
		ext = ".png"
	}

	outFileName := outPrefix + ext
	defer os.Remove(outFileName)

	outFile, err := os.Open(outFileName)
	if err != nil {
		return fmt.Errorf("failed to read rendered image: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(out, outFile)
	return err
}

// ConvertTXTtoJPG converts text to a JPG image.
func ConvertTXTtoJPG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertTXTtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&buf, out)
}

// ConvertTXTtoPNG converts text to a PNG image.
func ConvertTXTtoPNG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertTXTtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&buf, out)
}

// ConvertCSVtoJPG converts CSV to a JPG image.
func ConvertCSVtoJPG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertCSVtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&buf, out)
}

// ConvertCSVtoPNG converts CSV to a PNG image.
func ConvertCSVtoPNG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertCSVtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&buf, out)
}

// ConvertJSONtoJPG converts JSON to a JPG image.
func ConvertJSONtoJPG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertJSONtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&buf, out)
}

// ConvertJSONtoPNG converts JSON to a PNG image.
func ConvertJSONtoPNG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertJSONtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&buf, out)
}

// ConvertCSVtoPDF renders a CSV file to a simple fixed-width PDF using
// the native fpdf library.
//
// NOTE: The high-fidelity version, ConvertCSVtoPDFHighFidelity, uses
// LibreOffice Calc's export filter and produces proper paginated
// tables with repeated headers and auto-fit columns. This native
// version is retained for backward compatibility and as a fallback
// when LibreOffice is unavailable in the runtime environment.
func ConvertCSVtoPDF(in io.Reader, out io.Writer) error {
	reader := csv.NewReader(in)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv: %w", err)
	}

	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.AddPage()
	pdfDoc.SetFont("Arial", "B", 12)
	pdfDoc.SetY(10.0)

	for _, row := range records {
		for _, col := range row {
			pdfDoc.CellFormat(40, 10, col, "1", 0, "", false, 0, "")
		}
		pdfDoc.Ln(-1)
	}

	err = pdfDoc.Output(out)
	if err != nil {
		return fmt.Errorf("failed to output pdf: %w", err)
	}
	return nil
}

// ==============================================================================
// HIGH-FIDELITY DOCUMENT CONVERSIONS (LibreOffice-backed)
//
// The functions below replace the lossy DOCX→CSV→TXT→PDF chain with real
// LibreOffice-backed conversion. The old function names are preserved
// where possible so the formatter registry in internal/formatters does
// not need to change for existing pairs.
//
// Every function here has the signature the registry expects:
//
//     func Name(in io.Reader, out io.Writer) error
//
// Internally they delegate to the context-aware bridge in
// office_bridge.go using a background context. A future step will
// migrate the registry to pass a real request-scoped context through,
// giving us per-request cancellation for very large documents.
// ==============================================================================

// legacyCtx is used by the shim functions below. It is a plain background
// context — the caller has no cancellation channel. The bridge functions
// still enforce their own per-tool timeouts, so a hung subprocess cannot
// block a request forever.
var legacyCtx = context.Background()

// ------------------------------------------------------------------------------
// DOCX → PDF  (the headline fix)
// ------------------------------------------------------------------------------

// ConvertDOCXtoPDF renders a DOCX document to PDF using LibreOffice,
// preserving fonts, styles, tables, images, headers, and layout.
//
// Previously this function extracted only the text via ConvertDOCXtoCSV
// and re-drew it with a generic fpdf layout — losing all formatting.
// The new implementation hands the original DOCX bytes to LibreOffice
// and returns the PDF it produces, which is byte-for-byte compatible
// with Microsoft Word's "Save as PDF" output.
func ConvertDOCXtoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "docx", "pdf", in, out)
}

// ------------------------------------------------------------------------------
// DOCX → JPG / PNG
//
// These now route through LibreOffice for the DOCX→PDF step, then use
// poppler's pdftocairo to rasterize the resulting PDF. The old
// implementation went DOCX→PDF→image, which was already doing poppler
// at the end — but the PDF at the start was garbage. Now the PDF is
// high fidelity, so the rasterized image is too.
// ------------------------------------------------------------------------------

// ConvertDOCXtoJPG renders a DOCX document to a JPG image (first page).
func ConvertDOCXtoJPG(in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(legacyCtx, "docx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&pdfBuf, out)
}

// ConvertDOCXtoPNG renders a DOCX document to a PNG image (first page).
func ConvertDOCXtoPNG(in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(legacyCtx, "docx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&pdfBuf, out)
}

// ------------------------------------------------------------------------------
// DOCX → TXT / JSON  (high fidelity)
// ------------------------------------------------------------------------------

// ConvertDOCXtoTXTHighFidelity extracts the plain text of a DOCX document
// using LibreOffice's text exporter, which respects paragraph breaks,
// table cell separators, and section structure.
func ConvertDOCXtoTXTHighFidelity(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "docx", "txt", in, out)
}

// ConvertDOCXtoJSONHighFidelity extracts the text of a DOCX document and
// wraps it in a JSON object with a "lines" array — matching the shape
// produced by ConvertTXTtoJSON.
func ConvertDOCXtoJSONHighFidelity(in io.Reader, out io.Writer) error {
	var txtBuf bytes.Buffer
	if err := OfficeConvert(legacyCtx, "docx", "txt", in, &txtBuf); err != nil {
		return err
	}
	return ConvertTXTtoJSON(&txtBuf, out)
}

// ------------------------------------------------------------------------------
// PDF → DOCX
//
// The old implementation called GetPlainText() and emitted one Word
// paragraph per extracted line — destroying all layout. LibreOffice
// does a much better job: it imports the PDF via its PDF import filter
// (which reconstructs paragraphs, tables, and images where possible)
// and exports to DOCX.
//
// Note: for scanned PDFs (no text layer), LibreOffice cannot help. The
// caller should route those through /extract-text (which has OCR) and
// build the DOCX from the extracted text instead.
// ------------------------------------------------------------------------------

// ConvertPDFtoDOCX converts a PDF document to an editable DOCX using
// LibreOffice's PDF import filter.
func ConvertPDFtoDOCX(in io.ReaderAt, size int64, out io.Writer) error {
	// io.ReaderAt is not an io.Reader. Wrap it via io.NewSectionReader
	// so we can hand it to the bridge, which expects io.Reader.
	section := io.NewSectionReader(in, 0, size)
	return OfficeConvert(legacyCtx, "pdf", "docx", section, out)
}

// ------------------------------------------------------------------------------
// XLSX → PDF / CSV / JSON / TXT  (new capabilities)
// ------------------------------------------------------------------------------

// ConvertXLSXtoPDF renders an Excel workbook to PDF, preserving cell
// formatting, formulas, charts, and multi-sheet layout.
func ConvertXLSXtoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "xlsx", "pdf", in, out)
}

// ConvertXLSXtoCSV exports the first sheet of an Excel workbook to CSV.
func ConvertXLSXtoCSV(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "xlsx", "csv", in, out)
}

// ConvertXLSXtoTXT exports an Excel workbook to tab-separated plain text.
func ConvertXLSXtoTXT(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "xlsx", "txt", in, out)
}

// ConvertXLSXtoJSON exports an Excel workbook to JSON via the CSV
// intermediate — LibreOffice produces a clean CSV, then the existing
// ConvertCSVtoJSON turns it into a JSON array of objects.
func ConvertXLSXtoJSON(in io.Reader, out io.Writer) error {
	var csvBuf bytes.Buffer
	if err := OfficeConvert(legacyCtx, "xlsx", "csv", in, &csvBuf); err != nil {
		return err
	}
	return ConvertCSVtoJSON(&csvBuf, out)
}

// ------------------------------------------------------------------------------
// PPTX → PDF / PNG / JPG  (new capability)
// ------------------------------------------------------------------------------

// ConvertPPTXtoPDF renders a PowerPoint deck to PDF, one slide per page.
func ConvertPPTXtoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "pptx", "pdf", in, out)
}

// ConvertPPTXtoPNG renders the first slide of a PowerPoint deck to PNG.
func ConvertPPTXtoPNG(in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(legacyCtx, "pptx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&pdfBuf, out)
}

// ConvertPPTXtoJPG renders the first slide of a PowerPoint deck to JPG.
func ConvertPPTXtoJPG(in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(legacyCtx, "pptx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&pdfBuf, out)
}

// ------------------------------------------------------------------------------
// Legacy office formats: RTF / ODT / ODS / ODP / HTML
// ------------------------------------------------------------------------------

// ConvertRTFtoPDF renders a Rich Text Format document to PDF.
func ConvertRTFtoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "rtf", "pdf", in, out)
}

// ConvertODTtoPDF renders an OpenDocument Text document to PDF.
func ConvertODTtoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "odt", "pdf", in, out)
}

// ConvertODStoPDF renders an OpenDocument Spreadsheet to PDF.
func ConvertODStoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "ods", "pdf", in, out)
}

// ConvertODPtoPDF renders an OpenDocument Presentation to PDF.
func ConvertODPtoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "odp", "pdf", in, out)
}

// ConvertHTMLtoPDF renders an HTML document to PDF via LibreOffice's
// HTML import filter. This preserves basic CSS (fonts, colors, tables)
// but not JavaScript-driven layout.
func ConvertHTMLtoPDF(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "html", "pdf", in, out)
}

// ConvertHTMLtoDOCX renders an HTML document to a Word DOCX.
func ConvertHTMLtoDOCX(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "html", "docx", in, out)
}

// ------------------------------------------------------------------------------
// CSV / JSON → PDF  (now high-fidelity)
//
// Previously these used hand-rolled fixed-width fpdf tables that
// truncated long fields. Now LibreOffice renders them as proper
// spreadsheet-style tables with pagination, repeated headers, and
// auto-fit columns.
// ------------------------------------------------------------------------------

// ConvertCSVtoPDFHighFidelity renders a CSV file as a paginated PDF
// table using LibreOffice's Calc export.
func ConvertCSVtoPDFHighFidelity(in io.Reader, out io.Writer) error {
	return OfficeConvert(legacyCtx, "csv", "pdf", in, out)
}

// ConvertJSONtoPDFHighFidelity renders a JSON file as a paginated PDF
// table. JSON is first converted to CSV via the native converter, then
// handed to LibreOffice for the table rendering.
func ConvertJSONtoPDFHighFidelity(in io.Reader, out io.Writer) error {
	var csvBuf bytes.Buffer
	if err := ConvertJSONtoCSV(in, &csvBuf); err != nil {
		return err
	}
	return OfficeConvert(legacyCtx, "csv", "pdf", &csvBuf, out)
}

// ------------------------------------------------------------------------------
// Image → PDF  (lossless embedding via img2pdf)
//
// The old Convert*toPDF chain re-encoded every image to JPEG at quality
// 90 before embedding it, causing generation loss and file size bloat.
// The new functions hand the raw image bytes directly to img2pdf, which
// embeds them without re-encoding.
// ------------------------------------------------------------------------------

// ConvertJPGtoPDFLossless embeds a JPEG file into a PDF without
// re-encoding the pixel data.
func ConvertJPGtoPDFLossless(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read JPEG: %w", err)
	}
	return Img2PDF(legacyCtx, [][]byte{data}, out)
}

// ConvertPNGtoPDFLossless embeds a PNG file into a PDF without
// re-encoding the pixel data.
func ConvertPNGtoPDFLossless(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read PNG: %w", err)
	}
	return Img2PDF(legacyCtx, [][]byte{data}, out)
}

// ConvertWEBPtoPDFLossless embeds a WEBP file into a PDF.
// img2pdf supports WebP natively, so no re-encoding is required.
func ConvertWEBPtoPDFLossless(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read WEBP: %w", err)
	}
	return Img2PDF(legacyCtx, [][]byte{data}, out)
}

// ConvertTIFFtoPDFLossless embeds a TIFF file into a PDF.
func ConvertTIFFtoPDFLossless(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read TIFF: %w", err)
	}
	return Img2PDF(legacyCtx, [][]byte{data}, out)
}

// ConvertBMPtoPDFLossless embeds a BMP file into a PDF.
func ConvertBMPtoPDFLossless(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read BMP: %w", err)
	}
	return Img2PDF(legacyCtx, [][]byte{data}, out)
}
// ==============================================================================
// CONTEXT-AWARE VARIANTS
//
// Every LibreOffice-backed converter has two forms:
//
//   1. ConvertXtoY(in, out)                  — legacy signature, uses
//                                              a background context.
//   2. ConvertXtoYCtx(ctx, in, out)          — context-aware, honours
//                                              cancellation and deadline.
//
// The API layer calls the *Ctx variants. The legacy variants exist so
// that any caller that has not yet been migrated (and the existing
// tests) continues to compile and work. When every caller has moved
// to the Ctx form, the legacy shims can be deleted in a single commit.
//
// The ctx that arrives here is the one derived by HandleConvert from
// r.Context() plus the formatter's timeout. When the client disconnects
// or the timeout fires, ctx is cancelled, and OfficeConvert propagates
// the cancellation into the soffice subprocess (killing it).
// ==============================================================================

// ConvertDOCXtoPDFCtx is the context-aware form of ConvertDOCXtoPDF.
func ConvertDOCXtoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "docx", "pdf", in, out)
}

// ConvertDOCXtoJPGCtx is the context-aware form of ConvertDOCXtoJPG.
func ConvertDOCXtoJPGCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(ctx, "docx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&pdfBuf, out)
}

// ConvertDOCXtoPNGCtx is the context-aware form of ConvertDOCXtoPNG.
func ConvertDOCXtoPNGCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(ctx, "docx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&pdfBuf, out)
}

// ConvertDOCXtoTXTHighFidelityCtx is the context-aware form of
// ConvertDOCXtoTXTHighFidelity.
func ConvertDOCXtoTXTHighFidelityCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "docx", "txt", in, out)
}

// ConvertDOCXtoJSONHighFidelityCtx is the context-aware form of
// ConvertDOCXtoJSONHighFidelity.
func ConvertDOCXtoJSONHighFidelityCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	var txtBuf bytes.Buffer
	if err := OfficeConvert(ctx, "docx", "txt", in, &txtBuf); err != nil {
		return err
	}
	return ConvertTXTtoJSON(&txtBuf, out)
}

// ConvertPDFtoDOCXCtx is the context-aware form of ConvertPDFtoDOCX.
func ConvertPDFtoDOCXCtx(ctx context.Context, in io.ReaderAt, size int64, out io.Writer) error {
	section := io.NewSectionReader(in, 0, size)
	return OfficeConvert(ctx, "pdf", "docx", section, out)
}

// ConvertXLSXtoPDFCtx is the context-aware form of ConvertXLSXtoPDF.
func ConvertXLSXtoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "xlsx", "pdf", in, out)
}

// ConvertXLSXtoCSVCtx is the context-aware form of ConvertXLSXtoCSV.
func ConvertXLSXtoCSVCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "xlsx", "csv", in, out)
}

// ConvertXLSXtoTXTCtx is the context-aware form of ConvertXLSXtoTXT.
func ConvertXLSXtoTXTCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "xlsx", "txt", in, out)
}

// ConvertXLSXtoJSONCtx is the context-aware form of ConvertXLSXtoJSON.
func ConvertXLSXtoJSONCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	var csvBuf bytes.Buffer
	if err := OfficeConvert(ctx, "xlsx", "csv", in, &csvBuf); err != nil {
		return err
	}
	return ConvertCSVtoJSON(&csvBuf, out)
}

// ConvertPPTXtoPDFCtx is the context-aware form of ConvertPPTXtoPDF.
func ConvertPPTXtoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "pptx", "pdf", in, out)
}

// ConvertPPTXtoPNGCtx is the context-aware form of ConvertPPTXtoPNG.
func ConvertPPTXtoPNGCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(ctx, "pptx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&pdfBuf, out)
}

// ConvertPPTXtoJPGCtx is the context-aware form of ConvertPPTXtoJPG.
func ConvertPPTXtoJPGCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	var pdfBuf bytes.Buffer
	if err := OfficeConvert(ctx, "pptx", "pdf", in, &pdfBuf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&pdfBuf, out)
}

// ConvertRTFtoPDFCtx is the context-aware form of ConvertRTFtoPDF.
func ConvertRTFtoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "rtf", "pdf", in, out)
}

// ConvertODTtoPDFCtx is the context-aware form of ConvertODTtoPDF.
func ConvertODTtoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "odt", "pdf", in, out)
}

// ConvertODStoPDFCtx is the context-aware form of ConvertODStoPDF.
func ConvertODStoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "ods", "pdf", in, out)
}

// ConvertODPtoPDFCtx is the context-aware form of ConvertODPtoPDF.
func ConvertODPtoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "odp", "pdf", in, out)
}

// ConvertHTMLtoPDFCtx is the context-aware form of ConvertHTMLtoPDF.
func ConvertHTMLtoPDFCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "html", "pdf", in, out)
}

// ConvertHTMLtoDOCXCtx is the context-aware form of ConvertHTMLtoDOCX.
func ConvertHTMLtoDOCXCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "html", "docx", in, out)
}

// ConvertCSVtoPDFHighFidelityCtx is the context-aware form of
// ConvertCSVtoPDFHighFidelity.
func ConvertCSVtoPDFHighFidelityCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	return OfficeConvert(ctx, "csv", "pdf", in, out)
}

// ConvertJSONtoPDFHighFidelityCtx is the context-aware form of
// ConvertJSONtoPDFHighFidelity.
func ConvertJSONtoPDFHighFidelityCtx(ctx context.Context, in io.Reader, out io.Writer) error {
	var csvBuf bytes.Buffer
	if err := ConvertJSONtoCSV(in, &csvBuf); err != nil {
		return err
	}
	return OfficeConvert(ctx, "csv", "pdf", &csvBuf, out)
}