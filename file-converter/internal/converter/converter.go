package converter

import (
	"archive/zip"
	"bufio"
	"bytes"
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

// ConvertJPGtoPNG takes a JPEG from an io.Reader and writes a PNG to an io.Writer.
func ConvertJPGtoPNG(in io.Reader, out io.Writer) error {
	img, err := jpeg.Decode(in)
	if err != nil {
		return fmt.Errorf("failed to decode jpeg: %w", err)
	}
	err = png.Encode(out, img)
	if err != nil {
		return fmt.Errorf("failed to encode png: %w", err)
	}
	return nil
}

// ConvertCSVtoPDF reads CSV data from an io.Reader and writes a PDF to an io.Writer.
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

// ConvertPDFtoTXT extracts text from a PDF file
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

// ConvertTXTtoPDF converts text to PDF
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

// ConvertCSVtoTXT writes a CSV directly out as plain text
func ConvertCSVtoTXT(in io.Reader, out io.Writer) error {
	_, err := io.Copy(out, in)
	return err
}

// ConvertTXTtoJSON converts text lines into a JSON array of strings
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

// ConvertCSVtoDOCX converts CSV data into a DOCX document
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

// ConvertTXTtoCSV converts text lines into a single-column CSV
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

// ConvertPDFtoDOCX extracts text from PDF and writes it to a DOCX
func ConvertPDFtoDOCX(in io.ReaderAt, size int64, out io.Writer) error {
	fpdf, err := pdf.NewReader(in, size)
	if err != nil {
		return err
	}
	b, err := fpdf.GetPlainText()
	if err != nil {
		return err
	}

	doc := docx.NewFile()
	
	scanner := bufio.NewScanner(b)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)
	for scanner.Scan() {
		p := doc.AddParagraph()
		p.AddText(scanner.Text())
	}

	tmpFile, err := os.CreateTemp("", "out-*.docx")
	if err != nil {
		return err
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFilePath)

	if err := doc.Save(tmpFilePath); err != nil {
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

// ConvertPDFtoCSV extracts text from PDF and writes it to a CSV
func ConvertPDFtoCSV(in io.ReaderAt, size int64, out io.Writer) error {
	fpdf, err := pdf.NewReader(in, size)
	if err != nil {
		return err
	}
	b, err := fpdf.GetPlainText()
	if err != nil {
		return err
	}

	return ConvertTXTtoCSV(b, out)
}

// ConvertPDFtoJSON extracts text from PDF and writes it to JSON
func ConvertPDFtoJSON(in io.ReaderAt, size int64, out io.Writer) error {
	fpdf, err := pdf.NewReader(in, size)
	if err != nil {
		return err
	}
	b, err := fpdf.GetPlainText()
	if err != nil {
		return err
	}

	return ConvertTXTtoJSON(b, out)
}

// ConvertJSONtoDOCX converts JSON to a DOCX document
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

// ConvertJSONtoPDF converts JSON to a PDF document
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

// ConvertDOCXtoJSON extracts text from DOCX and formats it as JSON
func ConvertDOCXtoJSON(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertDOCXtoCSV(in, &buf); err != nil {
		return err
	}
	return ConvertTXTtoJSON(&buf, out)
}

// ConvertDOCXtoPDF extracts text from DOCX and writes it to PDF
func ConvertDOCXtoPDF(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertDOCXtoCSV(in, &buf); err != nil {
		return err
	}
	return ConvertTXTtoPDF(&buf, out)
}

// ConvertPDFtoJPG renders the first page of a PDF to JPG
func ConvertPDFtoJPG(in io.Reader, out io.Writer) error {
	return convertPDFToImage(in, out, "-jpeg")
}

// ConvertPDFtoPNG renders the first page of a PDF to PNG
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

// ConvertTXTtoJPG converts text to a JPG image
func ConvertTXTtoJPG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertTXTtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&buf, out)
}

// ConvertTXTtoPNG converts text to a PNG image
func ConvertTXTtoPNG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertTXTtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&buf, out)
}

// ConvertDOCXtoJPG converts DOCX to a JPG image
func ConvertDOCXtoJPG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertDOCXtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&buf, out)
}

// ConvertDOCXtoPNG converts DOCX to a PNG image
func ConvertDOCXtoPNG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertDOCXtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&buf, out)
}

// ConvertCSVtoJPG converts CSV to a JPG image
func ConvertCSVtoJPG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertCSVtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&buf, out)
}

// ConvertCSVtoPNG converts CSV to a PNG image
func ConvertCSVtoPNG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertCSVtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&buf, out)
}

// ConvertJSONtoJPG converts JSON to a JPG image
func ConvertJSONtoJPG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertJSONtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoJPG(&buf, out)
}

// ConvertJSONtoPNG converts JSON to a PNG image
func ConvertJSONtoPNG(in io.Reader, out io.Writer) error {
	var buf bytes.Buffer
	if err := ConvertJSONtoPDF(in, &buf); err != nil {
		return err
	}
	return ConvertPDFtoPNG(&buf, out)
}
