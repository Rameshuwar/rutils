package main

import (
	"log"
	"github.com/gingfrederik/docx"
	"github.com/go-pdf/fpdf"
)

func main() {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Hello World")
	err := pdf.OutputFileAndClose("persona-tests/tests/test-files/dummy.pdf")
	if err != nil {
		log.Fatal(err)
	}

	f := docx.NewFile()
	p := f.AddParagraph()
	p.AddText("Hello World")
	err = f.Save("persona-tests/tests/test-files/dummy.docx")
	if err != nil {
		log.Fatal(err)
	}
}
