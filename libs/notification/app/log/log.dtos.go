package log

import (
	"encoding/json"
	"time"
)

// CreateLogDto mirrors ntx-apps/libs/notification/src/api/log/dto/create-log.dto.ts.
// TS marks reference/passcode optional and key/subject/channels required; gox
// preserves those constraints. publishAt/expireAt default via the controller
// morph, not via DTO defaults.
type CreateLogDto struct {
	Reference   *string         `json:"reference,omitempty"`
	UserId      *string         `json:"userId,omitempty"`
	ClientId    *string         `json:"clientId,omitempty"`
	WorkspaceId *string         `json:"workspaceId,omitempty"`
	Key         string          `json:"key"         binding:"required"`
	Subject     string          `json:"subject"     binding:"required"`
	Channels    json.RawMessage `json:"channels"    binding:"required"`
	Data        json.RawMessage `json:"data,omitempty"`
	Message     json.RawMessage `json:"message,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	ReadAt      *time.Time      `json:"readAt,omitempty"`
	PublishAt   *time.Time      `json:"publishAt,omitempty"`
	ExpireAt    *time.Time      `json:"expireAt,omitempty"`
	Type        *string         `json:"type,omitempty"`
	Priority    *string         `json:"priority,omitempty"`
	Theme       *string         `json:"theme,omitempty"`
	Scope       *string         `json:"scope,omitempty"`
	Position    *string         `json:"position,omitempty"`
	Status      *string         `json:"status,omitempty"`
	TemplateId  *string         `json:"templateId,omitempty"`
	RuleId      *string         `json:"ruleId,omitempty"`
	Passcode    *string         `json:"passcode,omitempty"`
}

type UpdateLogDto struct {
	Status *string `json:"status,omitempty"`
}
