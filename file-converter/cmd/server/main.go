package main

import (
	"fmt"
	"log"
	"net/http"

	"file-converter/internal/api"
	_ "file-converter/docs" // Import swagger docs
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
	apiMux.HandleFunc("/convert", api.HandleConvert)
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
		if err := http.ListenAndServe(":8080", api.EnableCORS(apiMux)); err != nil {
			log.Fatalf("API Server failed to start: %v", err)
		}
	}()

	// Start UI server in foreground
	if err := http.ListenAndServe(":3000", uiMux); err != nil {
		log.Fatalf("UI Server failed to start: %v", err)
	}
}
