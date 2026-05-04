package log

import "encoding/json"

type CreateLogDto struct {
	Group      string          `json:"group"      binding:"required"`
	Service    string          `json:"service"    binding:"required"`
	EntityId   string          `json:"entityId"   binding:"required"`
	EntityName string          `json:"entityName" binding:"required"`
	Action     string          `json:"action"     binding:"required"`
	Desc       *string         `json:"desc,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Status     *string         `json:"status,omitempty"`
}

type UpdateLogDto struct {
	Group      *string         `json:"group,omitempty"`
	Service    *string         `json:"service,omitempty"`
	EntityId   *string         `json:"entityId,omitempty"`
	EntityName *string         `json:"entityName,omitempty"`
	Action     *string         `json:"action,omitempty"`
	Desc       *string         `json:"desc,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Status     *string         `json:"status,omitempty"`
}
