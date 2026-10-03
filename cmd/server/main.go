package main

import (
	"fmt"
	"log"
	"net/http"

	_ "file-converter/docs" // Import swagger docs
	"file-converter/internal/api"
	"file-converter/internal/auth"
	"file-converter/internal/converter" // NEW: for CheckExtractDependencies()
	_ "file-converter/internal/formatters"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title File Converter API
// @version 1.0
// @description Utility Microservice for converting files and user authentication.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// ============================================================
	// Startup environment checks (non-fatal)
	// ============================================================
	// /extract-text relies on external binaries at runtime:
	//   - tesseract  → OCR for images and scanned PDFs
	//   - pdftoppm   → PDF→image rendering (poppler-utils)
	// We do NOT fail fast — the rest of the service works fine
	// without them. We just warn the operator once at boot.
	if missing := converter.CheckExtractDependencies(); len(missing) > 0 {
		log.Printf(
			"[WARN] /extract-text missing runtime dependencies: %v — image and scanned-PDF extraction will be unavailable.",
			missing,
		)
	}

	// API Server
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/convert", api.HandleConvert)
	apiMux.HandleFunc("/convert-measurement", api.HandleMeasurementConvert)
	apiMux.HandleFunc("/convert-time", api.HandleTimeConvert)
	apiMux.HandleFunc("/convert-railway", api.HandleRailwayConvert)
	apiMux.HandleFunc("/convert-numeral", api.HandleNumeralConvert)
	apiMux.HandleFunc("/calculate-bmi", api.HandleBMICalculate)
	apiMux.HandleFunc("/calculate-age", api.HandleAgeCalculate)
	apiMux.HandleFunc("/convert-pdf-size", api.ConvertPDFSize)

	// ⬅️ NEW: Text Extraction (PDF + Image, with OCR fallback)
	apiMux.HandleFunc("/extract-text", api.HandleExtractText)

	// ⬅️ NEW: Percentage Calculator
	apiMux.HandleFunc("/calculate-percentage", api.HandlePercentageCalculate)

	// ⬅️ NEW: Loan EMI Calculator
	apiMux.HandleFunc("/calculate-emi", api.HandleEMICalculate)
	apiMux.HandleFunc("/calculate-tax", api.HandleTaxCalculate)
	apiMux.HandleFunc("/calculate-simple-interest", api.HandleSimpleInterestCalculate)
	apiMux.HandleFunc("/calculate-compound-interest", api.HandleCompoundInterestCalculate)
	apiMux.HandleFunc("/calculate-scientific", api.HandleScientificCalculate)
	apiMux.HandleFunc("/formats", api.HandleListFormats)
	apiMux.HandleFunc("/formats/", api.HandleFormatDetail)
	// ============================================================
	// Authentication Service & Endpoints
	// ============================================================
	cfg, err := auth.LoadConfig("config.json")
	if err != nil {
		log.Printf("[WARN] Failed to load config.json, using defaults: %v", err)
		cfg = auth.DefaultConfig()
	}

	userStore, err := auth.NewStore(cfg.GetUsersFilePath())
	if err != nil {
		log.Fatalf("Failed to initialize user store: %v", err)
	}

	emailSender := auth.NewEmailSender(cfg)
	authService := auth.NewAuthService(userStore, cfg, emailSender)
	authHandler := api.NewAuthHandler(authService, cfg)

	apiMux.HandleFunc("/auth/register", authHandler.Register)
	apiMux.HandleFunc("/auth/login", authHandler.Login)
	apiMux.HandleFunc("/auth/logout", authHandler.Logout)
	apiMux.HandleFunc("/auth/forgot-password", authHandler.ForgotPassword)
	apiMux.HandleFunc("/auth/reset-password", api.RequireAuth(cfg, authHandler.ResetPassword))
	apiMux.HandleFunc("/auth/change-password", api.RequireAuth(cfg, authHandler.ChangePassword))
	apiMux.HandleFunc("/auth/me", api.RequireAuth(cfg, authHandler.Me))
	// ── Repair engine ─────────────────────────────────────────
	apiMux.HandleFunc("/repair", api.HandleRepair)

	apiMux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// UI Server
	uiMux := http.NewServeMux()
	uiMux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	fmt.Println("==================================================")
	fmt.Println(" Utility Microservice Starting...")
	fmt.Println("==================================================")
	fmt.Println("Backend API:")
	fmt.Println(" -> POST http://localhost:8080/convert               (File Converter)")
	fmt.Println(" -> POST http://localhost:8080/convert-measurement   (Measurement Converter)")
	fmt.Println(" -> POST http://localhost:8080/convert-time          (Time Converter)")
	fmt.Println(" -> POST http://localhost:8080/convert-railway       (Railway Converter)")
	fmt.Println(" -> POST http://localhost:8080/convert-numeral       (Numeral Converter)")
	fmt.Println(" -> POST http://localhost:8080/calculate-bmi         (BMI Calculator)")
	fmt.Println(" -> POST http://localhost:8080/calculate-age         (Age Calculator)")
	fmt.Println(" -> POST http://localhost:8080/convert-pdf-size      (PDF Size Converter)")

	// ⬅️ NEW: Text Extraction in startup banner
	fmt.Println(" -> POST http://localhost:8080/extract-text          (Text Extraction: PDF + Image OCR)")
	fmt.Println(" -> POST http://localhost:8080/repair               (Rules-based file repair)")
	// ⬅️ NEW: Percentage Calculator in startup banner
	fmt.Println(" -> POST http://localhost:8080/calculate-percentage  (Percentage Calculator)")
	fmt.Println(" -> POST http://localhost:8080/calculate-simple-interest   (Simple Interest Calculator)")
	fmt.Println(" -> POST http://localhost:8080/calculate-compound-interest (Compound Interest Calculator)")
	fmt.Println(" -> POST http://localhost:8080/calculate-scientific      (Scientific Calculator)")
	// ⬅️ Authentication Endpoints in startup banner
	fmt.Println(" -> POST http://localhost:8080/auth/register         (User Registration)")
	fmt.Println(" -> POST http://localhost:8080/auth/login            (User Login - JWT)")
	fmt.Println(" -> POST http://localhost:8080/auth/logout           (User Logout)")
	fmt.Println(" -> POST http://localhost:8080/auth/forgot-password  (Forgot Password - Google SMTP)")
	fmt.Println(" -> POST http://localhost:8080/auth/reset-password   (Forced Password Reset)")
	fmt.Println(" -> POST http://localhost:8080/auth/change-password  (Change Password)")
	fmt.Println(" -> GET  http://localhost:8080/auth/me               (Current User Profile)")

	// ⬅️ NEW: Loan EMI Calculator in startup banner
	fmt.Println(" -> POST http://localhost:8080/calculate-emi         (Loan EMI Calculator)")
	fmt.Println(" -> POST http://localhost:8080/calculate-emi         (Loan EMI Calculator)")

	// ⬅️ NEW: Tax / VAT / GST Calculator in startup banner
	fmt.Println(" -> POST http://localhost:8080/calculate-tax         (Tax / VAT / GST Calculator)")

	fmt.Println(" -> GET  http://localhost:8080/swagger/doc.json       (Swagger JSON)")
	fmt.Println(" -> GET  http://localhost:8080/swagger/              (Swagger UI)")
	fmt.Println("Frontend UI:")
	fmt.Println(" -> GET  http://localhost:3000/")
	fmt.Println("==================================================")

	// Start API server in background
	go func() {
		if err := http.ListenAndServe(":8080", api.EnableCORS(apiMux)); err != nil {
			log.Fatalf("API Server failed to start: %v", err)
		}
	}()

	// Start UI server in a goroutine – it’s optional and should not crash the whole process
	go func() {
		if err := http.ListenAndServe(":3000", uiMux); err != nil {
			log.Printf("UI Server failed to start (non‑fatal): %v", err)
		}
	}()

	// Block forever so the program stays alive after both goroutines start
	select {}
}
