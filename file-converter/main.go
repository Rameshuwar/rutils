package main

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


	"log"
	"net/http"

	_ "file-converter/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title File Converter API
// @version 1.0
// @description Utility Microservice for converting files.
// @host localhost:8080
// @BasePath /
func main() {
	// API Server
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/convert", handleConvert)
	apiMux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// UI Server
	uiMux := http.NewServeMux()
	uiMux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	fmt.Println("==================================================")
	fmt.Println(" Utility Microservice Starting...")
	fmt.Println("==================================================")
	fmt.Println("Backend API:")
	fmt.Println(" -> POST http://localhost:8080/convert          (Dynamic Converter)")
	fmt.Println(" -> GET  http://localhost:8080/swagger/         (Swagger UI)")
	fmt.Println("Frontend UI:")
	fmt.Println(" -> GET  http://localhost:3000/")
	fmt.Println("==================================================")

	// Start API server in background
	go func() {
		if err := http.ListenAndServe(":8080", enableCORS(apiMux)); err != nil {
			log.Fatalf("API Server failed to start: %v", err)
		}
	}()

	// Start UI server in foreground
	if err := http.ListenAndServe(":3000", uiMux); err != nil {
		log.Fatalf("UI Server failed to start: %v", err)
	}
}

// enableCORS adds CORS headers to allow cross-origin requests
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 10 MB for safety
	r.ParseMultipartForm(10 << 20)

	fromType := r.FormValue("fromType")
	toType := r.FormValue("toType")

	if fromType == "" || toType == "" {
		http.Error(w, "Missing fromType or toType in form data", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request. Ensure form-data key is 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Default disposition
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"converted.%s\"", toType))

	conversionPath := fmt.Sprintf("%s-to-%s", fromType, toType)

	switch conversionPath {
	case "jpg-to-png":
		w.Header().Set("Content-Type", "image/png")
		err = ConvertJPGtoPNG(file, w)
	case "csv-to-pdf":
		w.Header().Set("Content-Type", "application/pdf")
		err = ConvertCSVtoPDF(file, w)
	case "csv-to-json":
		w.Header().Set("Content-Type", "application/json")
		err = ConvertCSVtoJSON(file, w)
	case "json-to-csv":
		w.Header().Set("Content-Type", "text/csv")
		err = ConvertJSONtoCSV(file, w)
	case "json-to-txt":
		w.Header().Set("Content-Type", "text/plain")
		err = ConvertJSONtoTXT(file, w)
	case "txt-to-docx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		err = ConvertTXTtoDOCX(file, w)
	case "pdf-to-txt":
		w.Header().Set("Content-Type", "text/plain")
		err = ConvertPDFtoTXT(file, fileHeader.Size, w)
	case "txt-to-pdf":
		w.Header().Set("Content-Type", "application/pdf")
		err = ConvertTXTtoPDF(file, w)
	case "png-to-jpg":
		w.Header().Set("Content-Type", "image/jpeg")
		err = ConvertPNGtoJPG(file, w)
	case "csv-to-txt":
		w.Header().Set("Content-Type", "text/plain")
		err = ConvertCSVtoTXT(file, w)
	case "docx-to-csv":
		w.Header().Set("Content-Type", "text/csv")
		err = ConvertDOCXtoCSV(file, w)
	case "docx-to-txt":
		// Same as docx to csv since it extracts text
		w.Header().Set("Content-Type", "text/plain")
		err = ConvertDOCXtoCSV(file, w)
	default:
		http.Error(w, fmt.Sprintf("Conversion from %s to %s is not yet supported. Please choose a different combination.", fromType, toType), http.StatusBadRequest)
		return
	}

	if err != nil {
		w.Header().Del("Content-Disposition")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusBadRequest)
		log.Printf("Conversion error (%s): %v", conversionPath, err)
		return
	}
}

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

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)
	pdf.SetY(10.0)

	for _, row := range records {
		for _, col := range row {
			pdf.CellFormat(40, 10, col, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}

	err = pdf.Output(out)
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
		// If it's an array, flatten each element as a separate row
		for _, item := range v {
			flat := make(map[string]interface{})
			flattenJSON("", item, flat)
			data = append(data, flat)
		}
	default:
		// If it's a single object or primitive, flatten it as a single row
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

	// Pretty print JSON as plain text
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

	// Save to a temporary file because gingfrederik/docx only supports writing to a file path
	tmpFile, err := os.CreateTemp("", "out-*.docx")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFilePath := tmpFile.Name()
	tmpFile.Close() // Close so the library can write to it
	defer os.Remove(tmpFilePath)

	if err := f.Save(tmpFilePath); err != nil {
		return fmt.Errorf("failed to save docx: %w", err)
	}

	// Read from temp file and stream to response
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
	// Read full DOCX into memory (it's a zip archive)
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
	writer.Write([]string{"Extracted Text"}) // Write header

	decoder := xml.NewDecoder(docXML)
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" { // <w:t> contains the actual text in DOCX
				var text string
				if err := decoder.DecodeElement(&text, &se); err == nil {
					// Write each text element as a new row in CSV
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
