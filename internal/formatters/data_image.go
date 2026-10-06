package formatters

import (
	"context"
	"io"

	"file-converter/internal/converter"
)

// data_image.go ports the "documents rendered to image" conversions
// from the old handler.go switch. Every one of these routes through an
// intermediate PDF + poppler rasterization.
func init() {

	registerDocToImage := func(from, to, mime, ext string, convert func(io.Reader, io.Writer) error) {
		Register(Formatter{
			From:       from,
			To:         to,
			OutputMIME: mime,
			OutputExt:  ext,
			Category:   CategoryData,
			Convert: func(_ context.Context, in io.Reader, out io.Writer) error {
				return convert(in, out)
			},
		})
	}

	registerDocToImage("csv", "jpg", "image/jpeg", ".jpg", converter.ConvertCSVtoJPG)
	registerDocToImage("csv", "png", "image/png", ".png", converter.ConvertCSVtoPNG)
	registerDocToImage("json", "jpg", "image/jpeg", ".jpg", converter.ConvertJSONtoJPG)
	registerDocToImage("json", "png", "image/png", ".png", converter.ConvertJSONtoPNG)
	registerDocToImage("txt", "jpg", "image/jpeg", ".jpg", converter.ConvertTXTtoJPG)
	registerDocToImage("txt", "png", "image/png", ".png", converter.ConvertTXTtoPNG)
	registerDocToImage("docx", "jpg", "image/jpeg", ".jpg", converter.ConvertDOCXtoJPG)
	registerDocToImage("docx", "png", "image/png", ".png", converter.ConvertDOCXtoPNG)
}
