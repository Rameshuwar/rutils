package formatters

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/go-pdf/fpdf"
)

// imageToSinglePagePDF wraps a decoded image into a one-page PDF whose
// page dimensions match the image's pixel dimensions (converted to
// points at 72 DPI). The image is re-encoded to JPEG at quality 90
// before being embedded so PNG/WebP/TIFF/BMP inputs all work through
// the same code path.
func imageToSinglePagePDF(img image.Image, out io.Writer) error {
	bounds := img.Bounds()
	widthPt := float64(bounds.Dx())
	heightPt := float64(bounds.Dy())

	// Re-encode to JPEG because fpdf's RegisterImageOptions only
	// supports JPG, PNG, and GIF — using JPG keeps behaviour uniform
	// across every input format.
	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("failed to re-encode image for PDF: %w", err)
	}

	pdfDoc := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "pt",
		Size:           fpdf.SizeType{Wd: widthPt, Ht: heightPt},
	})
	pdfDoc.SetMargins(0, 0, 0)
	pdfDoc.AddPage()

	opt := fpdf.ImageOptions{ImageType: "JPG", ReadDpi: false}
	pdfDoc.RegisterImageOptionsReader("embedded.jpg", opt, bytes.NewReader(jpegBuf.Bytes()))
	pdfDoc.ImageOptions("embedded.jpg", 0, 0, widthPt, heightPt, false, opt, 0, "")

	if err := pdfDoc.Output(out); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}
	return nil
}
