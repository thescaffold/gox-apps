package mapper

import "github.com/cbroglie/mustache"

type HTMLProvider struct{}

// Map mirrors TS HTMLService.map: it renders the templateBody as a mustache
// template against the input `data` (TS spreads {...data, ...renderUtils}; the
// renderUtils i18n helpers are not ported here). An empty template yields nil.
// (TS also supports loading templateUrl via the media service; the Go mapper has
// no media access, so only templateBody is handled.)
func (p *HTMLProvider) Map(payload *Payload) ([]any, error) {
	template, _ := payload.Input["templateBody"].(string)
	if template == "" {
		return nil, nil
	}
	rendered, err := mustache.Render(template, payload.Input["data"])
	if err != nil {
		rendered = template
	}
	return []any{rendered}, nil
}
