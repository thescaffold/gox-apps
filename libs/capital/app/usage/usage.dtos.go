package usage

import (
	"encoding/json"
	"time"
)

type CreateUsageDto struct {
	UserId      string          `json:"userId"      binding:"required"`
	ClientId    string          `json:"clientId"    binding:"required"`
	WorkspaceId string          `json:"workspaceId" binding:"required"`
	RateId      string          `json:"rateId"      binding:"required"`
	EntityId    string          `json:"entityId"    binding:"required"`
	EntityName  string          `json:"entityName"  binding:"required"`
	Quantity    int             `json:"quantity"    binding:"required"`
	StartAt     *time.Time      `json:"startAt,omitempty"`
	StopAt      *time.Time      `json:"stopAt,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateUsageDto struct {
	Quantity *int            `json:"quantity,omitempty"`
	StartAt  *time.Time      `json:"startAt,omitempty"`
	StopAt   *time.Time      `json:"stopAt,omitempty"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}
