package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	ntxhttp "github.com/thescaffold/gox-packages/libs/core/http"
	"github.com/thescaffold/gox-packages/libs/webhook"
)

// DeliveryRecorder receives the outcome of every signed delivery, so a host can
// keep a delivery log (and offer replay). Called once per Sink, after retries.
type DeliveryRecorder interface {
	Record(payload SinkPayload, url string, result webhook.Result)
}

type WebhookSink struct {
	client   *ntxhttp.Client `inject:""`
	recorder DeliveryRecorder
	sender   *webhook.Sender
}

// SetRecorder wires (or clears) the delivery recorder. Not goroutine-safe.
func (w *WebhookSink) SetRecorder(r DeliveryRecorder) { w.recorder = r }

// SetSender replaces the signed-delivery sender (tests, custom clients).
func (w *WebhookSink) SetSender(s *webhook.Sender) { w.sender = s }

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

	// A sink with a secret is delivered signed, with retries and a delivery
	// record: meta {"secret": "...", "signatureHeader": "Legger-Signature",
	// "maxAttempts": 5}. Without one the legacy single unsigned POST is used.
	if secret, _ := meta["secret"].(string); secret != "" {
		return w.sendSigned(payload, method, url, secret, meta)
	}

	ok, status, _, _, _ := w.client.External(method, url, event, nil, nil, 0)
	if !ok {
		return fmt.Errorf("webhook sink failed for %s with status %d", url, status)
	}
	return nil
}

func (w *WebhookSink) sendSigned(payload SinkPayload, method, url, secret string, meta map[string]any) error {
	event := payload.Event
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	sender := w.sender
	if sender == nil {
		sender = &webhook.Sender{}
	}
	cp := *sender
	if h, _ := meta["signatureHeader"].(string); h != "" {
		cp.Header = h
	}
	if n, ok := meta["maxAttempts"].(float64); ok && n >= 1 {
		cp.MaxAttempts = int(n)
	}
	name := ""
	if event.Type != nil {
		name = *event.Type
	}
	res := cp.Deliver(context.Background(), webhook.Request{
		URL: url, Method: method, Secret: secret, EventID: event.Id, EventName: name, Body: body,
	})
	if w.recorder != nil {
		w.recorder.Record(payload, url, res)
	}
	if !res.Delivered {
		return fmt.Errorf("webhook sink failed for %s after %d attempt(s), last status %d", url, len(res.Attempts), res.Last().Status)
	}
	return nil
}
