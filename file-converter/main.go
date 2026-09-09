package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Register our two conversion endpoints
	http.HandleFunc("/convert/image", handleImageConvert)
	http.HandleFunc("/convert/document", handleDocumentConvert)

	port := ":8080"
	fmt.Println("==================================================")
	fmt.Printf(" Utility Microservice starting on http://localhost%s\n", port)
	fmt.Println("==================================================")
	fmt.Println("Available endpoints:")
	fmt.Println(" -> POST http://localhost:8080/convert/image    (JPG to PNG)")
	fmt.Println(" -> POST http://localhost:8080/convert/document (CSV to PDF)")
	fmt.Println("==================================================")

	// Start the HTTP server
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// handleImageConvert handles POST requests for JPG to PNG conversions.
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

// handleDocumentConvert handles POST requests for CSV to PDF conversions.
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
