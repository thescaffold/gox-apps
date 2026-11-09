package app

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	"github.com/thescaffold/gox-apps/libs/polylog/pkg/signature"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/events"
)

// EventIngested is published on the process bus after a source or channel
// ingest is stored, for in-process consumers of provider webhooks (inbound
// mail, chain deposits, payments). Payload: {event, sourceId, type,
// category, verified, body (the raw body as text), headers (without
// Authorization and Cookie)}. The name is fixed: a caller's ?type= never
// becomes an internal event name.
const EventIngested = "apps.polylog.ingest.received"

// RequireSignatureVar, set to true, refuses source and channel ingests whose
// source has no signature rule (signature.FromMeta).
const RequireSignatureVar = "POLYLOG_INGEST_REQUIRE_SIGNATURE"

// ErrSignature is returned for a webhook whose signature does not verify.
var ErrSignature = errors.New("polylog: the webhook signature did not verify")

// verifySource checks the request against the source's signature rule. A
// source with no rule (or no such source) passes unverified, as before,
// unless RequireSignatureVar is set.
func (s *AppService) verifySource(sourceID string, body []byte, headers map[string]string) (verified bool, err error) {
	var meta json.RawMessage
	if s.sourceEntity != nil && sourceID != "" {
		if src, _ := s.sourceEntity.First(`"id" = ? OR "key" = ?`, sourceID, sourceID); src != nil {
			meta = src.Meta
		}
	}
	rule, ok := signature.FromMeta(meta)
	if !ok {
		if strings.EqualFold(os.Getenv(RequireSignatureVar), "true") {
			return false, ErrSignature
		}
		return false, nil
	}
	if err := signature.Verify(rule, body, headers, time.Now()); err != nil {
		return false, errors.Join(ErrSignature, err)
	}
	return true, nil
}

// asPayload is the body as the event's JSON payload: as sent when it is
// JSON, else {"raw": "<text>"}.
func asPayload(body []byte) json.RawMessage {
	if len(body) == 0 {
		return nil
	}
	if json.Valid(body) {
		return json.RawMessage(body)
	}
	raw := string(body)
	if !utf8.ValidString(raw) {
		raw = strings.ToValidUTF8(raw, "�")
	}
	b, _ := json.Marshal(map[string]string{"raw": raw})
	return b
}

func forwardedHeaders(h map[string]string) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		switch strings.ToLower(k) {
		case "authorization", "cookie", "proxy-authorization":
			continue
		}
		out[k] = v
	}
	return out
}

// ingestWebhook is the shared path of /ingest/:id, /ingest/source/:id and
// /ingest/channel/:id: verify, store, dispatch, announce.
func (s *AppService) ingestWebhook(item pipeline.Item, body []byte, headers map[string]string, properties map[string]string, ctx ntxctx.NTXContext) (any, error) {
	verified, err := s.verifySource(item.Source, body, headers)
	if err != nil {
		return nil, err
	}
	item.Payload = asPayload(body)
	if t, ok := properties["type"]; ok {
		item.Type = t
	}
	if c, ok := properties["category"]; ok {
		item.Category = c
	}
	ev, err := s.pipelineService.Ingest(item, ctx.UserID, ctx.ClientID, ctx.WorkspaceID)
	if err != nil || ev == nil {
		return ev, err
	}
	events.Default().Publish(EventIngested, map[string]any{
		"event":    ev,
		"sourceId": item.Source,
		"type":     item.Type,
		"category": item.Category,
		"verified": verified,
		"body":     string(body),
		"headers":  forwardedHeaders(headers),
	})
	return ev, nil
}
