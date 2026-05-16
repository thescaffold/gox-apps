package app

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	goqueues "github.com/awesome-goose/goose/modules/queues"
	"github.com/thescaffold/gox-apps/libs/notification/app/log"
	"github.com/thescaffold/gox-apps/libs/notification/app/message"
	"github.com/thescaffold/gox-apps/libs/notification/app/rule"
	"github.com/thescaffold/gox-apps/libs/notification/pkg/provider"
)

// AppService implements the notification scope/priority/subscription queries
// and the message.new persistence side. Mirrors
// ntx-apps/libs/notification/src/app.service.ts + app.controller.ts.
type AppService struct {
	logEntity     *log.LogEntity            `inject:""`
	messageEntity *message.MessageEntity    `inject:""`
	ruleEntity    *rule.RuleEntity          `inject:""`
	provider      *provider.ProviderService `inject:""`
}

// GetHello mirrors TS getHello() — "Hello World!".
func (s *AppService) GetHello() string { return "Hello World!" }

// OnMessageJob mirrors TS jobs[0] for `queue/apps/notification`:
// look up the persisted Log by reference and dispatch it through the
// ProviderService. Returns a logged payload + nil error so the worker
// treats the job as drained even when the Log is missing (matches TS
// idempotency).
func (s *AppService) OnMessageJob(job *goqueues.QueueJob) (any, error) {
	if job == nil || s.logEntity == nil {
		return nil, nil
	}
	var data map[string]any
	if len(job.Data) > 0 {
		_ = json.Unmarshal(job.Data, &data)
	}
	ref, _ := data["reference"].(string)
	if ref == "" {
		if l, ok := data["log"].(map[string]any); ok {
			ref, _ = l["reference"].(string)
		}
	}
	if ref == "" {
		return map[string]any{"jobId": job.Id, "status": "skipped"}, nil
	}
	row, err := s.logEntity.First(`"reference" = ?`, ref)
	if err != nil || row == nil {
		return map[string]any{"jobId": job.Id, "status": "missing", "reference": ref}, nil
	}
	if s.provider != nil {
		s.provider.Send(row)
	}
	return map[string]any{"jobId": job.Id, "status": "dispatched", "reference": ref}, nil
}

// ScopeQuery is the input set for FindByScope.
type ScopeQuery struct {
	Channel     string
	Scope       string
	UserID      string
	ClientID    string
	WorkspaceID string
	Page        int
	PerPage     int
}

// PriorityQuery is the input set for FindByPriority.
type PriorityQuery struct {
	Channel     string
	Priority    string
	UserID      string
	ClientID    string
	WorkspaceID string
	Page        int
	PerPage     int
}

// PaginatedMessages is the typed slice returned by FindByScope / FindByPriority.
type PaginatedMessages struct {
	Items   []message.Message
	Total   int64
	Page    int
	PerPage int
}

// FindByScope returns unread messages matching (channel, scope) for the user/client/workspace.
// Three OR'd filters mirror TS: client-only, user+client (no workspace), user+client+workspace.
// Mirrors TS app.controller.ts findByScope().
func (s *AppService) FindByScope(q ScopeQuery) (*PaginatedMessages, error) {
	return s.findMessages(q.Page, q.PerPage,
		map[string]any{"channel": q.Channel, "scope": q.Scope},
		q.UserID, q.ClientID, q.WorkspaceID,
	)
}

// FindByPriority returns unread messages matching (channel, priority) for the user/client/workspace.
// Mirrors TS app.controller.ts findByPriority().
func (s *AppService) FindByPriority(q PriorityQuery) (*PaginatedMessages, error) {
	return s.findMessages(q.Page, q.PerPage,
		map[string]any{"channel": q.Channel, "priority": q.Priority},
		q.UserID, q.ClientID, q.WorkspaceID,
	)
}

func (s *AppService) findMessages(page, perPage int, extra map[string]any, userID, clientID, workspaceID string) (*PaginatedMessages, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 12
	}

	// Build OR'd WHERE matching TS's three filters.
	var orParts []string
	var args []any
	add := func(userIsNull, workspaceIsNull bool) {
		var parts []string
		if userIsNull {
			parts = append(parts, `"user_id" IS NULL`)
		} else {
			parts = append(parts, `"user_id" = ?`)
			args = append(args, userID)
		}
		parts = append(parts, `"client_id" = ?`)
		args = append(args, clientID)
		if workspaceIsNull {
			parts = append(parts, `"workspace_id" IS NULL`)
		} else {
			parts = append(parts, `"workspace_id" = ?`)
			args = append(args, workspaceID)
		}
		for k, v := range extra {
			parts = append(parts, `"`+k+`" = ?`)
			args = append(args, v)
		}
		parts = append(parts, `"read_at" IS NULL`)
		orParts = append(orParts, "("+strings.Join(parts, " AND ")+")")
	}
	add(true, true)
	add(false, true)
	add(false, false)

	where := strings.Join(orParts, " OR ")
	total, err := s.messageEntity.Count(where, args...)
	if err != nil {
		return nil, err
	}
	offset := (page - 1) * perPage
	items, err := s.messageEntity.Find(offset, perPage, where, args...)
	if err != nil {
		return nil, err
	}
	return &PaginatedMessages{Items: items, Total: total, Page: page, PerPage: perPage}, nil
}

// GetSubscription returns the subscription rule for the given ref.
// Mirrors TS app.controller.ts getSubscription().
func (s *AppService) GetSubscription(ref string) (*rule.Rule, error) {
	if ref == "" {
		return nil, errors.New("ref required")
	}
	return s.ruleEntity.First(`"group" = ? AND "key" = ?`, "subscription", ref)
}

// UpdateSubscription upserts the subscription rule with the supplied rules JSON.
// Mirrors TS app.controller.ts updateSubscription() + the
// apps.identity.attribute.type.update subscription handler in app.controller.ts.
func (s *AppService) UpdateSubscription(ref string, rules json.RawMessage) (*rule.Rule, error) {
	if ref == "" {
		return nil, errors.New("ref required")
	}
	existing, _ := s.ruleEntity.First(`"group" = ? AND "key" = ?`, "subscription", ref)
	if existing != nil {
		existing.Rules = rules
		if _, err := s.ruleEntity.Update(existing, "id = ?", existing.Id); err != nil {
			return nil, err
		}
		return existing, nil
	}
	r := &rule.Rule{Group: "subscription", Key: ref, Rules: rules}
	if err := s.ruleEntity.Insert(r); err != nil {
		return nil, err
	}
	return r, nil
}

// OnMessageNew mirrors the TS subscription 'apps.notification.message.new'
// body: persist a Log row with publishAt/expireAt defaults, then dispatch
// the row through ProviderService — which either enqueues onto
// queue/apps/notification/message (when AppController.OnRegister wired the
// queuePump) or runs synchronous Send as a fallback.
func (s *AppService) OnMessageNew(p map[string]any) {
	if s.logEntity == nil {
		return
	}
	ref, _ := p["reference"].(string)
	key, _ := p["key"].(string)
	if ref == "" || key == "" {
		return
	}
	l := &log.Log{
		Reference: ref,
		Key:       key,
	}
	if v, ok := p["userId"].(string); ok && v != "" {
		l.UserId = &v
	}
	if v, ok := p["clientId"].(string); ok && v != "" {
		l.ClientId = &v
	}
	if v, ok := p["workspaceId"].(string); ok && v != "" {
		l.WorkspaceId = &v
	}
	if v, ok := p["subject"].(string); ok {
		l.Subject = v
	}
	if v, ok := p["channels"]; ok && v != nil {
		if b, err := json.Marshal(v); err == nil {
			l.Channels = b
		}
	}
	if v, ok := p["data"]; ok && v != nil {
		if b, err := json.Marshal(v); err == nil {
			l.Data = b
		}
	}
	for _, k := range []string{"type", "priority", "theme", "scope", "position"} {
		if v, ok := p[k].(string); ok && v != "" {
			val := v
			switch k {
			case "type":
				l.Type = &val
			case "priority":
				l.Priority = &val
			case "theme":
				l.Theme = &val
			case "scope":
				l.Scope = &val
			case "position":
				l.Position = &val
			}
		}
	}
	now := time.Now().UTC()
	if v, ok := p["publishAt"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			l.PublishAt = &t
		}
	}
	if l.PublishAt == nil {
		l.PublishAt = &now
	}
	if v, ok := p["expireAt"].(string); ok && v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			l.ExpireAt = &t
		}
	}
	if l.ExpireAt == nil {
		exp := l.PublishAt.Add(3 * 24 * time.Hour)
		l.ExpireAt = &exp
	}
	status := string(MessageStatusNew)
	l.Status = &status
	if err := s.logEntity.Insert(l); err != nil {
		return
	}
	// Dispatch the freshly-persisted Log row through the ProviderService.
	// When a QueuePusher is wired (set via ProviderService.SetQueuePusher),
	// the dispatch enqueues a `queue/apps/notification/message` job for
	// durable retry; otherwise it falls through to synchronous Send.
	if s.provider != nil {
		s.provider.Dispatch(l)
	}
}
