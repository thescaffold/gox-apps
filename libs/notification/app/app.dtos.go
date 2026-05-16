package app

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// FindByScopeDto carries the query params for GET /scope.
type FindByScopeDto struct {
	Channel string            `query:"channel"`
	Scope   string            `query:"scope"`
	Page    int               `query:"page"`
	PerPage int               `query:"perPage"`
	Ctx     ntxctx.NTXContext `context:"ntx"`
}

// FindByPriorityDto carries the query params for GET /priority.
type FindByPriorityDto struct {
	Channel  string            `query:"channel"`
	Priority string            `query:"priority"`
	Page     int               `query:"page"`
	PerPage  int               `query:"perPage"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// GetSubscriptionDto carries the query param for GET /subscription.
type GetSubscriptionDto struct {
	Ref string            `query:"ref"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// UpdateSubscriptionDto carries the body for PATCH /subscription.
type UpdateSubscriptionDto struct {
	Ref   string            `json:"ref"`
	Rules json.RawMessage   `json:"rules"`
	Ctx   ntxctx.NTXContext `context:"ntx"`
}

// MessageChannelType / MessageScopeType / MessagePriorityType / MessageStatusType
// mirror ntx-apps/libs/notification/src/app.dto.ts enums.
type MessageChannelType string
type MessageScopeType string
type MessagePriorityType string
type MessageStatusType string

const (
	MessageStatusNew     MessageStatusType = "new"
	MessageStatusQueued  MessageStatusType = "queued"
	MessageStatusFailed  MessageStatusType = "failed"
	MessageStatusSuccess MessageStatusType = "success"
)
