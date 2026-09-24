package converter

import (
	"bytes"
	"testing"
)

// minimalPDF is a very small valid-looking PDF used for unit tests.
var minimalPDF = []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
>>
endobj
trailer
<<
/Root 1 0 R
>>
%%EOF
`)

func TestPDFSizeInBytesKB(t *testing.T) {
	result, err := PDFSizeInBytes(2, "KB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int64(2 * 1024)
	if result != expected {
		t.Fatalf("expected %d bytes, got %d", expected, result)
	}
}

func TestPDFSizeInBytesMB(t *testing.T) {
	result, err := PDFSizeInBytes(2, "MB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := int64(2 * 1024 * 1024)
	if result != expected {
		t.Fatalf("expected %d bytes, got %d", expected, result)
	}
}

func TestPDFSizeInBytesInvalidUnit(t *testing.T) {
	_, err := PDFSizeInBytes(2, "GB")
	if err == nil {
		t.Fatal("expected error for unsupported unit")
	}
}

func TestPDFSizeInBytesInvalidValue(t *testing.T) {
	_, err := PDFSizeInBytes(0, "MB")
	if err == nil {
		t.Fatal("expected error for zero target size")
	}
}

func TestConvertPDFToTargetSizeInvalidPDF(t *testing.T) {
	input := []byte("this is not a PDF")
	_, err := ConvertPDFToTargetSize(input, "expand", "KB", 10)
	if err == nil {
		t.Fatal("expected error for invalid PDF")
	}
}

func TestConvertPDFToTargetSizeInvalidConversionType(t *testing.T) {
	_, err := ConvertPDFToTargetSize(minimalPDF, "invalid", "KB", 10)
	if err == nil {
		t.Fatal("expected error for invalid conversion type")
	}
}

func TestConvertPDFToTargetSizeInvalidDataType(t *testing.T) {
	_, err := ConvertPDFToTargetSize(minimalPDF, "expand", "GB", 10)
	if err == nil {
		t.Fatal("expected error for invalid data type")
	}
}

func TestConvertPDFToTargetSizeInvalidTarget(t *testing.T) {
	_, err := ConvertPDFToTargetSize(minimalPDF, "expand", "KB", 0)
	if err == nil {
		t.Fatal("expected error for zero target size")
	}
}

func TestCompressionTargetBelowMinimum(t *testing.T) {
	// 9 KB is below the 10 KB compression floor.
	_, err := ConvertPDFToTargetSize(minimalPDF, "compression", "KB", 9)
	if err == nil {
		t.Fatal("expected error for compression target < 10 KB")
	}
}

func TestExpansionTargetAboveMaximum(t *testing.T) {
	// 7000 KB is above the 6000 KB expansion ceiling.
	_, err := ConvertPDFToTargetSize(minimalPDF, "expand", "KB", 7000)
	if err == nil {
		t.Fatal("expected error for expansion target > 6000 KB")
	}
}

func TestExpandPDF(t *testing.T) {
	targetSize := int64(len(minimalPDF)) + 5000

	result, err := expandPDF(minimalPDF, targetSize)
	if err != nil {
		t.Fatalf("expandPDF returned unexpected error: %v", err)
	}

	if !isValidPDF(result.Output) {
		t.Fatal("expanded result is not a valid PDF")
	}

	if result.ActualSize < targetSize {
		t.Fatalf(
			"expanded PDF is too small: got %d, expected at least %d",
			result.ActualSize,
			targetSize,
		)
	}

	if !result.TargetMet {
		t.Fatal("expected TargetMet=true for successful expansion")
	}
}

func TestExpandPDFAlreadyLargeEnough(t *testing.T) {
	targetSize := int64(len(minimalPDF)) - 1

	result, err := expandPDF(minimalPDF, targetSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(result.Output, minimalPDF) {
		t.Fatal("PDF should not have changed when already large enough")
	}

	if !result.TargetMet {
		t.Fatal("expected TargetMet=true when input already satisfies target")
	}
}

func TestIsValidPDF(t *testing.T) {
	if !isValidPDF(minimalPDF) {
		t.Fatal("expected minimalPDF to be valid")
	}
}

func TestIsValidPDFRejectsInvalidHeader(t *testing.T) {
	invalid := []byte("NOTPDF\n%%EOF\n")
	if isValidPDF(invalid) {
		t.Fatal("expected invalid PDF header to be rejected")
	}
}

func TestIsValidPDFRejectsMissingEOF(t *testing.T) {
	invalid := []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\n")
	if isValidPDF(invalid) {
		t.Fatal("expected missing EOF marker to be rejected")
	}
}

func TestFormatPDFSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"bytes", 500, "500 bytes"},
		{"kilobytes", 2048, "2.00 KB"},
		{"megabytes", 2 * 1024 * 1024, "2.00 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPDFSize(tt.size)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestEnsurePDFOutputPath(t *testing.T) {
	tests := []struct {
		name             string
		originalName     string
		conversionType   string
		wantSuffix       string
	}{
		{"compression", "file.pdf", "compression", "_compressed.pdf"},
		{"expand", "file.pdf", "expand", "_expanded.pdf"},
		{"unknown", "file.pdf", "weird", "_converted.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsurePDFOutputPath(tt.originalName, tt.conversionType)
			if len(got) < len(tt.wantSuffix) ||
				got[len(got)-len(tt.wantSuffix):] != tt.wantSuffix {
				t.Fatalf("expected suffix %q, got %q", tt.wantSuffix, got)
			}
		})
	}
}