package pipeline

import (
	"encoding/json"
	"os"
	"strings"

	polylogchannel "github.com/thescaffold/gox-apps/libs/polylog/app/channel"
	polyloevent "github.com/thescaffold/gox-apps/libs/polylog/app/event"
	polylogeventlog "github.com/thescaffold/gox-apps/libs/polylog/app/eventlog"
	polylogsink "github.com/thescaffold/gox-apps/libs/polylog/app/sink"
	polylogsinktype "github.com/thescaffold/gox-apps/libs/polylog/app/sinktype"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

const asterisk = "*"

var unsafeEventList = []string{
	"apps.common.log.request",
	"apps.common.log.response",
}

var unsafeQueueList = []string{
	"apps.common.log.request",
	"apps.common.log.response",
	"apps.common.log.error",
	"apps.common.translate.miss",
}

// SinkPayload is passed to sink processors.
type SinkPayload struct {
	Event   *polyloevent.Event
	Channel *polylogchannel.Channel
	Sink    *polylogsink.Sink
}

// Item is the ingest input shape (matches ntx-core Item type).
type Item struct {
	EntityId   string          `json:"entityId"`
	EntityName string          `json:"entityName"`
	Source     string          `json:"source"`
	Category   string          `json:"category"`
	Type       string          `json:"type"`
	Reference  string          `json:"reference"`
	Version    string          `json:"version"`
	Payload    json.RawMessage `json:"payload"`
}

type PipelineService struct {
	eventEntity    *polyloevent.EventEntity        `inject:""`
	channelEntity  *polylogchannel.ChannelEntity   `inject:""`
	eventlogEntity *polylogeventlog.EventLogEntity `inject:""`
	sinkEntity     *polylogsink.SinkEntity         `inject:""`
	sinkTypeEntity *polylogsinktype.SinkTypeEntity `inject:""`
	webhookSink    *WebhookSink                    `inject:""`

	// queuePusher, when non-nil, is preferred over in-process goroutine
	// dispatch — see queue.go.
	queuePusher QueuePusher
}

// Ingest saves the event and processes it synchronously (gox has no queue).
func (s *PipelineService) Ingest(item Item, userId, clientId, workspaceId string) (*polyloevent.Event, error) {
	if item.Type != "" && containsStr(unsafeEventList, item.Type) {
		return nil, nil
	}

	if userId == "" {
		userId = os.Getenv("ROOT_USER_ID")
	}
	if clientId == "" {
		clientId = os.Getenv("ROOT_CLIENT_ID")
	}
	if workspaceId == "" {
		workspaceId = os.Getenv("ROOT_WORKSPACE_ID")
	}

	ref := item.Reference
	if ref == "" {
		ref = utils.Reference("EVT", 36)
	}
	ver := item.Version
	if ver == "" {
		ver = "1.0.0"
	}

	shouldQueue := item.Type != "" && !containsStr(unsafeQueueList, item.Type)
	status := "new"
	if shouldQueue {
		status = "queued"
	}

	payload := item.Payload
	if payload == nil {
		payload = json.RawMessage(`{}`)
	}

	ev := &polyloevent.Event{
		UserId:      userId,
		ClientId:    clientId,
		WorkspaceId: workspaceId,
		EntityId:    item.EntityId,
		EntityName:  item.EntityName,
		Category:    item.Category,
		Reference:   ref,
		Version:     ver,
		Payload:     payload,
		Status:      &status,
	}
	if item.Type != "" {
		ev.Type = &item.Type
	}

	if err := s.eventEntity.Insert(ev); err != nil {
		return nil, err
	}

	if shouldQueue {
		s.dispatch(ev)
	}

	return ev, nil
}

// Process handles a queued event: finds channels and discharges to sinks.
func (s *PipelineService) Process(ev *polyloevent.Event) error {
	if ev.Type != nil && containsStr(unsafeEventList, *ev.Type) {
		return nil
	}

	inProgress := "in_progress"
	_, _ = s.eventEntity.Update(&polyloevent.Event{Status: &inProgress}, `"id" = ?`, ev.Id)

	channels, err := s.getConfig(
		ev.UserId, ev.ClientId, ev.WorkspaceId,
		ev.EntityId, ev.EntityName,
		ptrStr(ev.Type), ev.Category,
	)
	if err != nil {
		return err
	}

	var processErr error
	for _, ch := range channels {
		if ch.SinkId == nil {
			continue
		}
		sink, err := s.sinkEntity.First(`"id" = ?`, *ch.SinkId)
		if err != nil || sink == nil {
			continue
		}
		if err := s.discharge(SinkPayload{Event: ev, Channel: &ch, Sink: sink}); err != nil {
			processErr = err
		}
	}

	finalStatus := "completed"
	if processErr != nil {
		finalStatus = "failed"
	}
	_, _ = s.eventEntity.Update(&polyloevent.Event{Status: &finalStatus}, `"id" = ?`, ev.Id)
	return processErr
}

func (s *PipelineService) discharge(payload SinkPayload) error {
	if payload.Sink == nil {
		return nil
	}

	sinkType, err := s.sinkTypeEntity.First(`"id" = ?`, payload.Sink.TypeId)
	if err != nil || sinkType == nil {
		return nil
	}

	key := ""
	if sinkType.Key != nil {
		key = *sinkType.Key
	}

	switch key {
	case "webhook":
		return s.webhookSink.Sink(payload)
	}
	return nil
}

func (s *PipelineService) getConfig(userId, clientId, workspaceId, entityId, entityName, typ, category string) ([]polylogchannel.Channel, error) {
	if entityName == "source" {
		variations := dottedStringVariations(typ)
		var whereParts []string
		var args []any

		// base conditions
		whereParts = append(whereParts, `("user_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "category" = ? AND "type" = ?)`)
		args = append(args, userId, workspaceId, entityId, category, asterisk)
		whereParts = append(whereParts, `("user_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "category" = ? AND "type" = ?)`)
		args = append(args, userId, workspaceId, asterisk, category, asterisk)

		for _, v := range variations {
			whereParts = append(whereParts, `("user_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "category" = ? AND "type" = ?)`)
			args = append(args, userId, workspaceId, entityId, category, v)
			whereParts = append(whereParts, `("user_id" = ? AND "workspace_id" = ? AND "source_id" = ? AND "category" = ? AND "type" = ?)`)
			args = append(args, userId, workspaceId, asterisk, category, v)
		}

		where := strings.Join(whereParts, " OR ")
		channels, err := s.channelEntity.Some(where, args...)
		if err != nil {
			return nil, err
		}
		return dedup(channels), nil
	}

	// entityName == "channel"
	channels, err := s.channelEntity.Some(`"id" = ?`, entityId)
	if err != nil {
		return nil, err
	}
	return dedup(channels), nil
}

func dottedStringVariations(s string) []string {
	segments := strings.Split(s, ".")
	total := 1 << len(segments)
	var out []string
	for i := 0; i < total; i++ {
		parts := make([]string, len(segments))
		for j, seg := range segments {
			if i&(1<<j) != 0 {
				parts[j] = asterisk
			} else {
				parts[j] = seg
			}
		}
		out = append(out, strings.Join(parts, "."))
	}
	return out
}

func dedup(channels []polylogchannel.Channel) []polylogchannel.Channel {
	seen := map[string]bool{}
	var out []polylogchannel.Channel
	for _, ch := range channels {
		if !seen[ch.Id] {
			seen[ch.Id] = true
			out = append(out, ch)
		}
	}
	return out
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
