package main

import (
	"encoding/csv"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/go-pdf/fpdf"
)

// ConvertJPGtoPNG takes a JPEG from an io.Reader and writes a PNG to an io.Writer.
func ConvertJPGtoPNG(in io.Reader, out io.Writer) error {
	// Decode the JPEG image
	img, err := jpeg.Decode(in)
	if err != nil {
		return fmt.Errorf("failed to decode jpeg: %w", err)
	}

	// Encode the image as PNG and write to the output
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

	// Initialize a new A4 portrait PDF
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 12)

	// Set initial coordinates
	y := 10.0
	pdf.SetY(y)

	// Iterate through records and print them in a simple grid
	for _, row := range records {
		for _, col := range row {
			// Using a fixed width of 40mm for simplicity for our MVP
			pdf.CellFormat(40, 10, col, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1) // Move to next line
	}

	// Write the resulting PDF to the output
	err = pdf.Output(out)
	if err != nil {
		return fmt.Errorf("failed to output pdf: %w", err)
	}

	return nil
}
