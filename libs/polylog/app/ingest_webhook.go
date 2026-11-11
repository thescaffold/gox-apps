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

// ErrUnknownSource is returned by the webhook routes for a source or channel that is not
// registered (PLAN U26): they answered 200 and stored, and announced on the bus, anything an
// anonymous caller posted under any id. Senders with no registered source use POST /ingest.
var ErrUnknownSource = errors.New("polylog: no such source or channel")

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

// checkIntake guards POST /ingest and /ingest/batch, which carry no signature (PLAN U26): an
// item that names a source with a signature rule (by id or key, as item.Source or
// item.EntityId) is refused, so the open intake browser SDKs post to cannot stand in for a
// provider webhook the source only accepts signed. With RequireSignatureVar set, every item
// is refused there: such a deployment takes signed webhooks only.
func (s *AppService) checkIntake(items ...pipeline.Item) error {
	if strings.EqualFold(os.Getenv(RequireSignatureVar), "true") {
		return ErrSignature
	}
	if s.sourceEntity == nil {
		return nil
	}
	for _, it := range items {
		for _, id := range []string{it.Source, it.EntityId} {
			if id == "" {
				continue
			}
			if src, _ := s.sourceEntity.First(`"id" = ? OR "key" = ?`, id, id); src != nil {
				if _, signed := signature.FromMeta(src.Meta); signed {
					return ErrSignature
				}
			}
		}
	}
	return nil
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
	sourceID, err := s.webhookSource(item)
	if err != nil {
		return nil, err
	}
	verified, err := s.verifySource(sourceID, body, headers)
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

// webhookSource is the registered source a webhook delivery belongs to: the source named by id or
// key, or, for POST /ingest/channel/:id, the channel's own source (so a channel's deliveries are
// held to its source's signature rule, which they never were). Unknown is ErrUnknownSource. A
// service built without entities (some tests) checks nothing, as before.
func (s *AppService) webhookSource(item pipeline.Item) (string, error) {
	if item.EntityName == "channel" {
		if s.channelEntity == nil {
			return item.Source, nil
		}
		ch, _ := s.channelEntity.First(`"id" = ?`, item.Source)
		if ch == nil {
			return "", ErrUnknownSource
		}
		return ch.SourceId, nil
	}
	if s.sourceEntity == nil {
		return item.Source, nil
	}
	if item.Source == "" {
		return "", ErrUnknownSource
	}
	src, _ := s.sourceEntity.First(`"id" = ? OR "key" = ?`, item.Source, item.Source)
	if src == nil {
		return "", ErrUnknownSource
	}
	return src.Id, nil
}
