package app

import (
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// GetDto is the query payload for GET /get. Mirrors TS @Query('key').
type GetDto struct {
	Key string            `query:"key"   binding:"required"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// SetDto is the body for POST /set, /setnx and /getset. Mirrors TS SetDto:
// `{ key, value?, duration? }`. `duration` is a number of MILLISECONDS — TS
// passes it to dayjs `.add(duration)` whose default unit is millisecond.
type SetDto struct {
	Key      string            `json:"key"      binding:"required"`
	Value    any               `json:"value,omitempty"`
	Duration int64             `json:"duration,omitempty"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// DelDto carries the key for the /del endpoint. Mirrors TS @Query('key').
type DelDto struct {
	Key string            `query:"key"   binding:"required"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// TTLDto carries the key for the /ttl endpoint.
type TTLDto struct {
	Key string            `query:"key"   binding:"required"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// IncrDto carries the key for the /incr endpoint. TS extracts the key from
// the JSON body via @Body('key'), so the wire shape is `{ "key": "…" }`.
type IncrDto struct {
	Key string            `json:"key"   binding:"required"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}
