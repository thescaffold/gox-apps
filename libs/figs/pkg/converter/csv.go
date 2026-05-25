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
		if rows, ok := normalizeRows(arg); ok {
			if err := w.WriteAll(rows); err != nil {
				return nil, err
			}
			continue
		}
		if err := w.Write([]string{fmt.Sprintf("%v", arg)}); err != nil {
			return nil, err
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
