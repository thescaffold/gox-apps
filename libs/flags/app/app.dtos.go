package app

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries the request context for GET / (getHello).
type HelloDto struct {
	NTX ntxctx.NTXContext `context:"ntx"`
}

// FlagDef mirrors TS FlagDto used inside RegisterDto.
type FlagDef struct {
	Name     string          `json:"name"     binding:"required"`
	Limit    int             `json:"limit"    binding:"required"`
	Priority int             `json:"priority" binding:"required"`
	Level    string          `json:"level"    binding:"required"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

// EnvironmentRef mirrors TS environment / environmentType ref. TS gives them
// distinct types (EnvironmentDto / EnvironmentTypeDto) but both shapes are
// `{name, ...}` at the wire level — we use a single Go type for both.
type EnvironmentRef struct {
	ID   *string `json:"id,omitempty"`
	Name string  `json:"name"`
}

// RegisterDto mirrors TS RegisterDto for POST /register.
type RegisterDto struct {
	NTX             ntxctx.NTXContext `context:"ntx"`
	EnvironmentType EnvironmentRef    `json:"environmentType" binding:"required"`
	Environment     EnvironmentRef    `json:"environment"     binding:"required"`
	Flags           []FlagDef         `json:"flags"           binding:"required"`
}

// StatusDto is the shared body for /log /status /limit endpoints.
type StatusDto struct {
	NTX         ntxctx.NTXContext `context:"ntx"`
	Environment EnvironmentRef    `json:"environment" binding:"required"`
	Name        string            `json:"name,omitempty"`
	Names       [][]string        `json:"names,omitempty"`
	Limit       *int              `json:"limit,omitempty"`
	Level       *string           `json:"level,omitempty"`
	UserId      *string           `json:"userId,omitempty"`
	ClientId    *string           `json:"clientId,omitempty"`
	WorkspaceId *string           `json:"workspaceId,omitempty"`
}

// FlagLevelUser is the default level used by /log /status /limit when omitted
// (TS FlagLevelType.User).
const FlagLevelUser = "user"
