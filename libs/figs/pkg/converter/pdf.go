package converter

import "bytes"

type PDFProvider struct{}

func (p *PDFProvider) Convert(args ...any) (*Response, error) {
	// stub: returns empty pdf bytes
	return &Response{
		Buffer:    bytes.NewBufferString("").Bytes(),
		Mime:      "application/pdf",
		Encoding:  "binary",
		Extension: "pdf",
	}, nil
}
