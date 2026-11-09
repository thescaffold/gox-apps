package app

import (
	"testing"

	"github.com/awesome-goose/goose/io/input"
	goosetest "github.com/awesome-goose/goose/testing"
)

// The :key, :code and :value path segments reach the handler (goose binds
// path segments with the param tag; path:"..." was never read, so every lookup
// ran with an empty key).
func TestPathParamsBind(t *testing.T) {
	ctx := goosetest.NewMockContext()
	ctx.MockRequest().WithParams(map[string]string{"key": "k1", "code": "c1", "value": "v1"})
	var a FilterByKeyDto
	var b FindByCodeDto
	var c FindByValueDto
	for _, dto := range []any{&a, &b, &c} {
		if err := input.NewInput(ctx).Populate(dto); err != nil {
			t.Fatal(err)
		}
	}
	if a.Key != "k1" || b.Code != "c1" || c.Value != "v1" {
		t.Fatalf("bound %q %q %q", a.Key, b.Code, c.Value)
	}
}
