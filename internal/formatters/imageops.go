package formatters

import (
	"bytes"
	"fmt"
	"image"
	"io"

	// Stdlib decoders.
	_ "image/jpeg"
	_ "image/png"

	// Extra decoders — require golang.org/x/image.
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// decodeAny reads the full input, sniffs its magic bytes, and returns a
// decoded image. It rejects inputs smaller than 5 bytes and inputs whose
// magic bytes do not match any known image format.
func decodeAny(in io.Reader) (image.Image, string, error) {
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image input: %w", err)
	}
	if len(data) < 5 {
		return nil, "", ErrInvalidInput
	}

	format := sniffImageFormat(data)
	if format == "" {
		return nil, "", ErrInvalidInput
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode %s image: %w", format, err)
	}
	return img, format, nil
}

// sniffImageFormat returns a short format name ("jpeg", "png", "webp",
// "tiff", "bmp") or "" if the magic bytes are unrecognised.
func sniffImageFormat(data []byte) string {
	switch {
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "jpeg"
	case len(data) >= 8 && bytes.Equal(data[:8], []byte("\x89PNG\r\n\x1a\n")):
		return "png"
	case len(data) >= 12 && bytes.Equal(data[0:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return "webp"
	case len(data) >= 4 && (bytes.Equal(data[:4], []byte("II*\x00")) || bytes.Equal(data[:4], []byte("MM\x00*"))):
		return "tiff"
	case len(data) >= 2 && bytes.Equal(data[:2], []byte("BM")):
		return "bmp"
	}
	return ""
}
