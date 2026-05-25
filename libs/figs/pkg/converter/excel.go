package converter

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type ExcelProvider struct{}

// Convert accepts rows of []string or [][]string and returns an xlsx buffer.
// Each []string argument is treated as one row; [][]string is expanded into multiple rows.
func (p *ExcelProvider) Convert(args ...any) (*Response, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	row := 1
	for _, arg := range args {
		rows, ok := normalizeRows(arg)
		if !ok {
			colName, _ := excelize.ColumnNumberToName(1)
			_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, row), fmt.Sprintf("%v", arg))
			row++
			continue
		}
		for _, r := range rows {
			for col, cell := range r {
				colName, _ := excelize.ColumnNumberToName(col + 1)
				_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, row), cell)
			}
			row++
		}
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("excel: write: %w", err)
	}

	return &Response{
		Buffer:    buf.Bytes(),
		Mime:      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		// TS excel.service.ts sets encoding 'utf-8' (persisted to file.output.encoding).
		Encoding:  "utf-8",
		Extension: "xlsx",
	}, nil
}
