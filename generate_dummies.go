package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/gingfrederik/docx"
	"github.com/go-pdf/fpdf"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// outDir is the destination for every fixture. Kept relative so this
// program can be run from the repo root.
const outDir = "persona-tests/tests/test-files"

func main() {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatalf("failed to create %s: %v", outDir, err)
	}

	must(writePDF(filepath.Join(outDir, "dummy.pdf")))
	must(writeDOCX(filepath.Join(outDir, "dummy.docx")))
	must(writePNG(filepath.Join(outDir, "dummy.png")))
	must(writeJPG(filepath.Join(outDir, "dummy.jpg")))
	must(writeBMP(filepath.Join(outDir, "dummy.bmp")))
	must(writeTIFF(filepath.Join(outDir, "dummy.tiff")))
	must(writeWEBP(filepath.Join(outDir, "dummy.webp")))

	log.Println("all fixtures written to", outDir)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// ---------------------------------------------------------------------
// Document fixtures (unchanged from the original generate_dummies.go)
// ---------------------------------------------------------------------

func writePDF(path string) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Hello World")
	return pdf.OutputFileAndClose(path)
}

func writeDOCX(path string) error {
	f := docx.NewFile()
	p := f.AddParagraph()
	p.AddText("Hello World")
	return f.Save(path)
}

// ---------------------------------------------------------------------
// Image fixtures
//
// We build a 64x64 solid-red image once and encode it into every
// format that has an encoder in the Go standard library or
// golang.org/x/image.
// ---------------------------------------------------------------------

func makeTestImage() image.Image {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(
		img,
		img.Bounds(),
		&image.Uniform{color.RGBA{R: 220, G: 40, B: 40, A: 255}},
		image.Point{},
		draw.Src,
	)
	return img
}

func writePNG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, makeTestImage())
}

func writeJPG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, makeTestImage(), &jpeg.Options{Quality: 85})
}

func writeBMP(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return bmp.Encode(f, makeTestImage())
}

func writeTIFF(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	// tiff.Encode accepts a *tiff.Options; nil uses the package default
	// (uncompressed, which is what our tests expect).
	return tiff.Encode(f, makeTestImage(), nil)
}

// writeWEBP writes a hand-crafted minimal WEBP file.
//
// Why not use x/image/webp to encode? Because golang.org/x/image/webp
// only supports DECODING. There is no public encoder in the Go
// ecosystem that doesn't pull in cgo. So we embed the smallest
// possible valid WEBP (a 1x1 lossless VP8L white pixel) as raw bytes.
//
// This is enough to prove the /convert pipeline handles the format.
func writeWEBP(path string) error {
	// Minimal valid WEBP: RIFF header + VP8L chunk + 1x1 white pixel.
	// Reference: https://developers.google.com/speed/webp/docs/riff_container
	data := []byte{
		// RIFF header
		0x52, 0x49, 0x46, 0x46, // "RIFF"
		0x24, 0x00, 0x00, 0x00, // file size - 8 (36 bytes)
		0x57, 0x45, 0x42, 0x50, // "WEBP"

		// VP8L chunk
		0x56, 0x50, 0x38, 0x4C, // "VP8L"
		0x18, 0x00, 0x00, 0x00, // chunk size (24 bytes)

		// VP8L bitstream (1x1 lossless image, white pixel)
		0x2F, 0x00, 0x00, 0x00, 0x10, 0x07, 0x10, 0x11,
		0x11, 0x88, 0x88, 0xFE, 0x07, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	return os.WriteFile(path, data, 0o644)
}