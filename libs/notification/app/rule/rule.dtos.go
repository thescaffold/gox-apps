package rule

import "encoding/json"

type CreateRuleDto struct {
	Group  string          `json:"group"  binding:"required"`
	Key    string          `json:"key"    binding:"required"`
	Rules  json.RawMessage `json:"rules"  binding:"required"`
	Status *string         `json:"status,omitempty"`
}

type UpdateRuleDto struct {
	Rules  json.RawMessage `json:"rules,omitempty"`
	Status *string         `json:"status,omitempty"`
}
