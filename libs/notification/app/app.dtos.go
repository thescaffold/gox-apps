package app

import (
	"encoding/json"

	"github.com/awesome-goose/goose/openapi"
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

// CaptureDto carries GET /capture: read the messages the capture provider
// stored (NOTIFICATION_DEFAULT_PROVIDER=capture), so end-to-end tests can read
// OTPs and links. The caller proves itself with X-Capture-Token.
type CaptureDto struct {
	Token string `header:"X-Capture-Token"`
	To    string `query:"to"`
	Limit string `query:"limit"`
}

// InboxDto carries the query params for GET /inbox.
type InboxDto struct {
	Cursor string            `query:"cursor"`
	Limit  int               `query:"limit"`
	Unread bool              `query:"unread"`
	Ctx    ntxctx.NTXContext `context:"ntx"`
}

// InboxCountDto carries the caller for GET /inbox/unread-count and POST /inbox/read.
type InboxCountDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// InboxMarkAllDto carries the request context for POST /inbox/read.
type InboxMarkAllDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// InboxCount is GET /inbox/unread-count's body.
type InboxCount struct {
	Count int64 `json:"count" binding:"required"`
}

// InboxUpdated is POST /inbox/read's body: how many messages it marked read.
type InboxUpdated struct {
	Updated int64 `json:"updated" binding:"required"`
}

// OpenAPI documents the inbox (PLAN U25): a page is the shared {items, next} shape.
func (InboxDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "The signed-in person's in-app notifications, newest first", Response: openapi.Page[InboxItem]{}}
}

func (InboxCountDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "How many of the signed-in person's notifications are unread", Response: InboxCount{}}
}

func (InboxMarkAllDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "Mark every notification read", Response: InboxUpdated{}}
}

func (InboxReadDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "Mark one notification read", Response: InboxItem{}}
}

// InboxReadDto carries the message id for POST /inbox/:id/read.
type InboxReadDto struct {
	ID  string            `param:"id"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}
