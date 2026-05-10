package app

import "encoding/json"

type HealthDto struct{}

type CheckDto struct {
	Id string `param:"id" binding:"required"`
}

// PingDto mirrors TS PingDto used by POST /ping.
type PingDto struct {
	Name  string          `json:"name"  binding:"required"`
	State string          `json:"state" binding:"required"`
	Meta  json.RawMessage `json:"meta,omitempty"`
}

// RegisterServiceDto mirrors TS CreateServiceDto used by POST /register.
type RegisterServiceDto struct {
	Name   string  `json:"name"             binding:"required"`
	Desc   *string `json:"desc,omitempty"`
	Status *string `json:"status,omitempty"`
}
