package app

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/thescaffold/gox-apps/libs/notification/app/message"
	"github.com/thescaffold/gox-apps/libs/notification/app/rule"
)

// AppService implements the notification scope/priority/subscription queries.
// Mirrors ntx-apps/libs/notification/src/app.service.ts + app.controller.ts.
type AppService struct {
	messageEntity *message.MessageEntity `inject:""`
	ruleEntity    *rule.RuleEntity       `inject:""`
}

// GetHello mirrors TS getHello() — "Hello World!".
func (s *AppService) GetHello() string { return "Hello World!" }

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
	if q.Channel == "" || q.Scope == "" {
		return &PaginatedMessages{}, nil
	}
	return s.findMessages(q.Page, q.PerPage,
		map[string]any{"channel": q.Channel, "scope": q.Scope},
		q.UserID, q.ClientID, q.WorkspaceID,
	)
}

// FindByPriority returns unread messages matching (channel, priority) for the user/client/workspace.
// Mirrors TS app.controller.ts findByPriority().
func (s *AppService) FindByPriority(q PriorityQuery) (*PaginatedMessages, error) {
	if q.Channel == "" || q.Priority == "" {
		return &PaginatedMessages{}, nil
	}
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
		// fixed extra filters and read_at IS NULL
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
// Mirrors TS app.controller.ts updateSubscription().
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
