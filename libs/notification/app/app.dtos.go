package app

import "encoding/json"

// HealthDto is the body for GET /.
type HealthDto struct{}

// FindByScopeDto carries the query params for GET /scope.
type FindByScopeDto struct {
	Channel string `query:"channel"`
	Scope   string `query:"scope"`
	Page    int    `query:"page"`
	PerPage int    `query:"perPage"`
}

// FindByPriorityDto carries the query params for GET /priority.
type FindByPriorityDto struct {
	Channel  string `query:"channel"`
	Priority string `query:"priority"`
	Page     int    `query:"page"`
	PerPage  int    `query:"perPage"`
}

// GetSubscriptionDto carries the query param for GET /subscription.
type GetSubscriptionDto struct {
	Ref string `query:"ref"`
}

// UpdateSubscriptionDto carries the body for PATCH /subscription.
type UpdateSubscriptionDto struct {
	Ref   string          `json:"ref"   binding:"required"`
	Rules json.RawMessage `json:"rules"`
}

// MessageChannelType / MessageScopeType / MessagePriorityType / MessageStatusType
// mirror TS app.dto.ts enums.
type MessageChannelType string
type MessageScopeType string
type MessagePriorityType string
type MessageStatusType string

const (
	MessageStatusNew  MessageStatusType = "new"
	MessageStatusSent MessageStatusType = "sent"
	MessageStatusRead MessageStatusType = "read"
)
