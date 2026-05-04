package log

import "encoding/json"

type CreateLogDto struct {
	Reference   string          `json:"reference"   binding:"required"`
	UserId      *string         `json:"userId,omitempty"`
	ClientId    *string         `json:"clientId,omitempty"`
	WorkspaceId *string         `json:"workspaceId,omitempty"`
	Key         string          `json:"key"         binding:"required"`
	Subject     string          `json:"subject"     binding:"required"`
	Channels    json.RawMessage `json:"channels"    binding:"required"`
	Data        json.RawMessage `json:"data,omitempty"`
	Message     json.RawMessage `json:"message,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	TemplateId  *string         `json:"templateId,omitempty"`
	RuleId      *string         `json:"ruleId,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateLogDto struct {
	Status *string `json:"status,omitempty"`
}
