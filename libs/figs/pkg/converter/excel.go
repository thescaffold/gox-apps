package converter

import "bytes"

type ExcelProvider struct{}

func (p *ExcelProvider) Convert(args ...any) (*Response, error) {
	// stub: returns empty xlsx bytes
	return &Response{
		Buffer:    bytes.NewBufferString("").Bytes(),
		Mime:      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Encoding:  "binary",
		Extension: "xlsx",
	}, nil
}
