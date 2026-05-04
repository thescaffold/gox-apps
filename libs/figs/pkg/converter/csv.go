package converter

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

type CSVProvider struct{}

func (p *CSVProvider) Convert(args ...any) (*Response, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, arg := range args {
		switch v := arg.(type) {
		case []string:
			if err := w.Write(v); err != nil {
				return nil, err
			}
		case [][]string:
			if err := w.WriteAll(v); err != nil {
				return nil, err
			}
		default:
			if err := w.Write([]string{fmt.Sprintf("%v", v)}); err != nil {
				return nil, err
			}
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return &Response{
		Buffer:    buf.Bytes(),
		Mime:      "text/csv",
		Encoding:  "utf-8",
		Extension: "csv",
	}, nil
}
