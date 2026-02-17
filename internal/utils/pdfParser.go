package utils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ledongthuc/pdf"
)

// ExtractPDFText extracts text content from a PDF file
func ExtractPDFText(filepath string) (string, error) {
	f, r, err := pdf.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			// Skip pages that fail to parse
			continue
		}

		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// IsTextFile checks if file is text-based
func IsTextFile(filepath string) bool {
	textExts := map[string]bool{
		".txt": true, ".md": true, ".go": true, ".js": true,
		".py": true, ".java": true, ".c": true, ".cpp": true,
		".h": true, ".json": true, ".xml": true, ".yaml": true,
		".yml": true, ".toml": true, ".ini": true, ".conf": true,
		".sh": true, ".bat": true, ".ps1": true, ".html": true,
		".css": true, ".sql": true, ".log": true, ".csv": true,
	}

	ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])
	return textExts[ext]
}

// IsPDF checks if file is a PDF
func IsPDF(filepath string) bool {
	return strings.HasSuffix(strings.ToLower(filepath), ".pdf")
}
