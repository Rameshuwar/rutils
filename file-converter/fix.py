import re
with open('internal/converter/converter.go', 'r') as f:
    content = f.read()

old_func = """func ConvertPDFtoTXT(in io.ReaderAt, size int64, out io.Writer) error {
	f, err := pdf.NewReader(in, size)
	if err != nil {
		return fmt.Errorf("failed to read pdf: %w", err)
	}
	b, err := f.GetPlainText()
	if err != nil {
		return fmt.Errorf("failed to extract text from pdf: %w", err)
	}
	if len(bytes.TrimSpace(b)) == 0 {
		return fmt.Errorf("no text found. Scanned PDFs or image-based PDFs are not supported")
	}
	_, err = io.Copy(out, b)
	return err
}"""

new_func = """func ConvertPDFtoTXT(in io.ReaderAt, size int64, out io.Writer) error {
	f, err := pdf.NewReader(in, size)
	if err != nil {
		return fmt.Errorf("failed to read pdf: %w", err)
	}
	b, err := f.GetPlainText()
	if err != nil {
		return fmt.Errorf("failed to extract text from pdf: %w", err)
	}
	textBytes, err := io.ReadAll(b)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(textBytes)) == 0 {
		return fmt.Errorf("no text found. Scanned PDFs or image-based PDFs are not supported")
	}
	_, err = io.Copy(out, bytes.NewReader(textBytes))
	return err
}"""

content = content.replace(old_func, new_func)
with open('internal/converter/converter.go', 'w') as f:
    f.write(content)
