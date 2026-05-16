package channel

import "encoding/json"

// CreateChannelDto mirrors ntx-apps/libs/polylog/src/api/channel/dto/create-channel.dto.ts.
// userId/clientId/workspaceId are NOT in the TS DTO — they are populated by
// the controller morph from request context.
type CreateChannelDto struct {
	UserId      string          `json:"-"`
	ClientId    string          `json:"-"`
	WorkspaceId string          `json:"-"`
	Category    string          `json:"category"   binding:"required"`
	Key         *string         `json:"key,omitempty"`
	Visibility  *string         `json:"visibility,omitempty"`
	SourceId    string          `json:"sourceId"   binding:"required"`
	SinkId      *string         `json:"sinkId,omitempty"`
	Type        *string         `json:"type,omitempty"`
	Tags        json.RawMessage `json:"tags,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateChannelDto struct {
	SinkId *string `json:"sinkId,omitempty"`
	Status *string `json:"status,omitempty"`
}
