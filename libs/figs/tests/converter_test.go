package tests

import (
	"strings"
	"testing"

	test "github.com/awesome-goose/goose/testing"
	"github.com/thescaffold/gox-apps/libs/figs/pkg/converter"
	"github.com/thescaffold/gox-apps/libs/figs/pkg/mapper"
)

func TestConverter(t *testing.T) {
	test.NewSuiteRunner(t, &ConverterSuite{}).Run()
}

type ConverterSuite struct{ test.Suite }

// The data mapper returns [headers, data]; the CSV converter must write the
// header row followed by the data rows, handling JSON-decoded []any shapes.
// Mirrors TS DataService.map → CSVService.convert(headers, data).
func (s *ConverterSuite) TestDataMapperToCSV_HeaderThenRows() {
	payload := &mapper.Payload{
		Input: map[string]any{
			"templateObject": map[string]any{"headers": []any{"name", "age"}},
			"data":           []any{[]any{"alice", 30}, []any{"bob", 25}},
		},
	}
	mapResult, err := (&mapper.Service{}).Use(mapper.Data).Map(payload)
	s.T.Expect(err == nil).ToEqual(true)
	// [headers, data]
	s.T.Expect(len(mapResult)).ToEqual(2)

	resp, err := (&converter.Service{}).Use(converter.CSV).Convert(mapResult...)
	s.T.Expect(err == nil).ToEqual(true)
	got := strings.TrimSpace(string(resp.Buffer))
	// header row first, then data rows — no "[alice 30]" garbage cells.
	s.T.Expect(got).ToEqual("name,age\nalice,30\nbob,25")
}
