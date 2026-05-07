package converter

import (
	"bytes"
	"fmt"

	"github.com/go-pdf/fpdf"
)

type PDFProvider struct{}

// Convert accepts rows of []string or [][]string and renders each row as a line in the PDF.
func (p *PDFProvider) Convert(args ...any) (*Response, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 11)

	lineH := 7.0
	for _, arg := range args {
		switch v := arg.(type) {
		case []string:
			pdf.CellFormat(0, lineH, joinCells(v), "", 1, "", false, 0, "")
		case [][]string:
			for _, row := range v {
				pdf.CellFormat(0, lineH, joinCells(row), "", 1, "", false, 0, "")
			}
		default:
			pdf.CellFormat(0, lineH, fmt.Sprintf("%v", v), "", 1, "", false, 0, "")
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf: output: %w", err)
	}

	return &Response{
		Buffer:    buf.Bytes(),
		Mime:      "application/pdf",
		Encoding:  "binary",
		Extension: "pdf",
	}, nil
}

func joinCells(cells []string) string {
	out := ""
	for i, c := range cells {
		if i > 0 {
			out += "  "
		}
		out += c
	}
	return out
}
