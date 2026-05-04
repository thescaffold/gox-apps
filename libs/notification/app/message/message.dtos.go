package message

import (
	"encoding/json"
	"time"
)

type CreateMessageDto struct {
	Reference   *string         `json:"reference,omitempty"`
	UserId      *string         `json:"userId,omitempty"`
	ClientId    *string         `json:"clientId,omitempty"`
	WorkspaceId *string         `json:"workspaceId,omitempty"`
	Key         *string         `json:"key,omitempty"`
	Subject     *string         `json:"subject,omitempty"`
	Channel     *string         `json:"channel,omitempty"`
	Message     json.RawMessage `json:"message,omitempty"`
	ReadAt      *time.Time      `json:"readAt,omitempty"`
	PublishAt   *time.Time      `json:"publishAt,omitempty"`
	ExpireAt    *time.Time      `json:"expireAt,omitempty"`
	Type        *string         `json:"type,omitempty"`
	Priority    *string         `json:"priority,omitempty"`
	Theme       *string         `json:"theme,omitempty"`
	Scope       *string         `json:"scope,omitempty"`
	Position    *string         `json:"position,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateMessageDto struct {
	ReadAt    *time.Time `json:"readAt,omitempty"`
	PublishAt *time.Time `json:"publishAt,omitempty"`
	ExpireAt  *time.Time `json:"expireAt,omitempty"`
	Status    *string    `json:"status,omitempty"`
}
