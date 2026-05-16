package app

import (
	"time"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// PushConfigDto mirrors TS app.dto.ts PushConfigDto. Each field is optional;
// TS only forwards `config` to the service when at least one inner field is set.
type PushConfigDto struct {
	Priority   *int       `json:"priority,omitempty"`
	RetryLimit *int       `json:"retryLimit,omitempty"`
	RetryDelay *int       `json:"retryDelay,omitempty"`
	StartAt    *time.Time `json:"startAt,omitempty"`
	ExpireAt   *time.Time `json:"expireAt,omitempty"`
	Singleton  *bool      `json:"singleton,omitempty"`
	Frequency  *string    `json:"frequency,omitempty"`
}

// PushDto is the body of POST /. Mirrors TS PushDto: the priority/retry/etc.
// knobs live in a nested `config` object, not on the top level.
type PushDto struct {
	Queue  string            `json:"queue" binding:"required"`
	Job    string            `json:"job"   binding:"required"`
	Data   any               `json:"data,omitempty"`
	Config *PushConfigDto    `json:"config,omitempty"`
	Ctx    ntxctx.NTXContext `context:"ntx"`
}

// PopDto is the query payload for GET /. Both fields are required and TS
// returns a translated error envelope when either is missing.
type PopDto struct {
	Queue string            `query:"queue"`
	Job   string            `query:"job"`
	Ctx   ntxctx.NTXContext `context:"ntx"`
}

// LogDto is the body of PATCH /. Mirrors TS LogDto.
type LogDto struct {
	JobId  string            `json:"jobId"  binding:"required"`
	Output any               `json:"output,omitempty"`
	Status string            `json:"status" binding:"required"`
	Ctx    ntxctx.NTXContext `context:"ntx"`
}
