package pipeline

import (
	"encoding/json"
	"fmt"

	ntxhttp "github.com/thescaffold/gox-packages-core/http"
)

type WebhookSink struct {
	client *ntxhttp.Client `inject:""`
}

// Sink dispatches an event payload to the configured webhook URL.
func (w *WebhookSink) Sink(payload SinkPayload) error {
	sink := payload.Sink
	event := payload.Event

	if sink == nil {
		return nil
	}

	var meta map[string]any
	if sink.Meta != nil {
		_ = json.Unmarshal(sink.Meta, &meta)
	}

	method, _ := meta["method"].(string)
	url, _ := meta["url"].(string)

	if method == "" || url == "" {
		return nil
	}

	ok, status, _, _, _ := w.client.External(method, url, event, nil, nil, 0)
	if !ok {
		return fmt.Errorf("webhook sink failed for %s with status %d", url, status)
	}
	return nil
}
