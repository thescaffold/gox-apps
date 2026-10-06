package pipeline

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"context"

	polyloevent "github.com/thescaffold/gox-apps/libs/polylog/app/event"
	polylogsink "github.com/thescaffold/gox-apps/libs/polylog/app/sink"
	"github.com/thescaffold/gox-packages/libs/webhook"
)

type recorder struct {
	url string
	res webhook.Result
}

func (r *recorder) Record(_ SinkPayload, url string, res webhook.Result) { r.url, r.res = url, res }

func sinkFor(url string, meta map[string]any) SinkPayload {
	m, _ := json.Marshal(meta)
	typ := "wallet.credited"
	ev := &polyloevent.Event{Type: &typ, Payload: json.RawMessage(`{"amount":500000}`)}
	ev.Id = "evt_1"
	return SinkPayload{Event: ev, Sink: &polylogsink.Sink{Meta: m}}
}

func TestSignedWebhookIsVerifiableRetriedAndRecorded(t *testing.T) {
	var calls int32
	var verified error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		verified = webhook.Verify("whsec_test", r.Header.Get("Legger-Signature"), body, time.Now(), 0)
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(502)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	w := &WebhookSink{}
	w.SetSender(&webhook.Sender{Backoff: func(int) time.Duration { return time.Millisecond }, Sleep: func(context.Context, time.Duration) {}})
	rec := &recorder{}
	w.SetRecorder(rec)

	err := w.Sink(sinkFor(srv.URL, map[string]any{"method": "POST", "url": srv.URL, "secret": "whsec_test", "signatureHeader": "Legger-Signature"}))
	if err != nil {
		t.Fatal(err)
	}
	if verified != nil {
		t.Fatalf("receiver could not verify the signature: %v", verified)
	}
	if calls != 2 || !rec.res.Delivered || len(rec.res.Attempts) != 2 || rec.url != srv.URL {
		t.Fatalf("calls=%d %+v", calls, rec.res)
	}
}

func TestSignedWebhookReportsFailureAfterTheConfiguredAttempts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	w := &WebhookSink{}
	w.SetSender(&webhook.Sender{Backoff: func(int) time.Duration { return 0 }, Sleep: func(context.Context, time.Duration) {}})
	rec := &recorder{}
	w.SetRecorder(rec)
	err := w.Sink(sinkFor(srv.URL, map[string]any{"method": "POST", "url": srv.URL, "secret": "s", "maxAttempts": 3}))
	if err == nil || len(rec.res.Attempts) != 3 || rec.res.Delivered {
		t.Fatalf("%v %+v", err, rec.res)
	}
}
