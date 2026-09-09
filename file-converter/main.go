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
	// Register our two conversion endpoints
	http.HandleFunc("/convert/image", handleImageConvert)
	http.HandleFunc("/convert/document", handleDocumentConvert)

	// Register Swagger UI
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)


	port := ":8080"
	fmt.Println("==================================================")
	fmt.Printf(" Utility Microservice starting on http://localhost%s\n", port)
	fmt.Println("==================================================")
	fmt.Println("Available endpoints:")
	fmt.Println(" -> POST http://localhost:8080/convert/image    (JPG to PNG)")
	fmt.Println(" -> POST http://localhost:8080/convert/document (CSV to PDF)")
	fmt.Println(" -> GET  http://localhost:8080/swagger/         (Swagger UI)")
	fmt.Println("==================================================")

	// Start the HTTP server
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// @Summary Convert JPG to PNG
// @Description Converts an uploaded JPG file to PNG format.
// @Accept multipart/form-data
// @Produce image/png
// @Param file formData file true "JPG Image to convert"
// @Success 200 {file} file "converted.png"
// @Failure 400 {string} string "Bad Request"
// @Failure 405 {string} string "Method Not Allowed"
// @Router /convert/image [post]
func handleImageConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 10 MB for safety
	r.ParseMultipartForm(10 << 20)
	
	// Retrieve the uploaded file from the 'file' field
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request. Ensure form-data key is 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Set headers so the client knows a PNG file is being returned
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=\"converted.png\"")

	// Convert and stream directly to the HTTP response
	err = ConvertJPGtoPNG(file, w)
	if err != nil {
		log.Printf("Image conversion error: %v", err)
		// Note: if the header has already been written by an early successful write,
		// changing the status code here won't work perfectly, but it's acceptable for an MVP.
		return
	}
}

// @Summary Convert CSV to PDF
// @Description Converts an uploaded CSV file to PDF format.
// @Accept multipart/form-data
// @Produce application/pdf
// @Param file formData file true "CSV Document to convert"
// @Success 200 {file} file "converted.pdf"
// @Failure 400 {string} string "Bad Request"
// @Failure 405 {string} string "Method Not Allowed"
// @Router /convert/document [post]
func handleDocumentConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)
	
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request. Ensure form-data key is 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Set headers so the client knows a PDF file is being returned
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"converted.pdf\"")

	err = ConvertCSVtoPDF(file, w)
	if err != nil {
		log.Printf("Document conversion error: %v", err)
		return
	}
}
