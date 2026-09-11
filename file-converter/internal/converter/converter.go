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
	_, err = io.Copy(out, b)
	return err
}

// ConvertTXTtoPDF converts text to PDF
func ConvertTXTtoPDF(in io.Reader, out io.Writer) error {
	pdfDoc := fpdf.New("P", "mm", "A4", "")
	pdfDoc.AddPage()
	pdfDoc.SetFont("Arial", "", 12)
	pdfDoc.SetY(10.0)

	scanner := bufio.NewScanner(in)
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
