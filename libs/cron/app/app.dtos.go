package app

import (
	"time"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// RegisterConfigDto mirrors TS app.dto.ts RegisterConfigDto — every field is
// optional. Dates are ISO strings on the wire and parsed into time.Time.
type RegisterConfigDto struct {
	Priority   *int       `json:"priority,omitempty"`
	RetryLimit *int       `json:"retryLimit,omitempty"`
	RetryDelay *int       `json:"retryDelay,omitempty"`
	StartAt    *time.Time `json:"startAt,omitempty"`
	ExpireAt   *time.Time `json:"expireAt,omitempty"`
}

// RegisterDto is the body of POST / (register). Mirrors TS RegisterDto: the
// optional knobs live in a nested `config` object, not on the top level.
type RegisterDto struct {
	Group   string             `json:"group"   binding:"required"`
	Name    string             `json:"name"    binding:"required"`
	Pattern string             `json:"pattern" binding:"required"`
	Config  *RegisterConfigDto `json:"config,omitempty"`
	Ctx     ntxctx.NTXContext  `context:"ntx"`
}

// SelectDto is the query payload for GET /. Both fields are required and TS
// returns a translated error envelope when either is missing.
type SelectDto struct {
	Group string            `query:"group"`
	Name  string            `query:"name"`
	Ctx   ntxctx.NTXContext `context:"ntx"`
}

// LogDto is the body of PATCH / (log). Mirrors TS LogDto.
type LogDto struct {
	JobId  string            `json:"jobId"  binding:"required"`
	Output any               `json:"output,omitempty"`
	Status string            `json:"status" binding:"required"`
	Ctx    ntxctx.NTXContext `context:"ntx"`
}
