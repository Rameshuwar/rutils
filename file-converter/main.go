package main

import (
	"fmt"
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
	// Unified conversion endpoint
	http.HandleFunc("/convert", handleConvert)

	// Register Swagger UI
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Serve the compiled frontend
	fs := http.FileServer(http.Dir("./frontend/dist"))
	http.Handle("/", fs)

	port := ":8080"
	fmt.Println("==================================================")
	fmt.Printf(" Utility Microservice starting on http://localhost%s\n", port)
	fmt.Println("==================================================")
	fmt.Println("Available endpoints:")
	fmt.Println(" -> POST http://localhost:8080/convert          (Dynamic Converter)")
	fmt.Println(" -> GET  http://localhost:8080/swagger/         (Swagger UI)")
	fmt.Println("==================================================")

	// Start the HTTP server
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
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
		log.Printf("Conversion error (%s): %v", conversionPath, err)
		return
	}
}
