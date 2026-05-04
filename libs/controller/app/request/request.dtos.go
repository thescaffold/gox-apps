package request

import "encoding/json"

type CreateRequestDto struct {
	Group    *string         `json:"group,omitempty"`
	Service  *string         `json:"service,omitempty"`
	Type     string          `json:"type"`
	Method   string          `json:"method"  binding:"required"`
	Url      string          `json:"url"     binding:"required"`
	Queries  json.RawMessage `json:"queries,omitempty"`
	Body     json.RawMessage `json:"body,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

type UpdateRequestDto struct {
	Group    *string         `json:"group,omitempty"`
	Service  *string         `json:"service,omitempty"`
	Type     *string         `json:"type,omitempty"`
	Method   *string         `json:"method,omitempty"`
	Url      *string         `json:"url,omitempty"`
	Queries  json.RawMessage `json:"queries,omitempty"`
	Body     json.RawMessage `json:"body,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
