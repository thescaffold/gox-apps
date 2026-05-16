package app

import (
	"encoding/json"

	"github.com/thescaffold/gox-apps/libs/polylog/pkg/pipeline"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries request context for GET / (getHello).
type HelloDto struct {
	NTX ntxctx.NTXContext `context:"ntx"`
}

// IngestDto carries one ingest item.
type IngestDto struct {
	NTX        ntxctx.NTXContext `context:"ntx"`
	EntityId   string            `json:"entityId"`
	EntityName string            `json:"entityName"`
	Source     string            `json:"source"`
	Category   string            `json:"category"`
	Type       string            `json:"type"`
	Reference  string            `json:"reference"`
	Version    string            `json:"version"`
	Payload    json.RawMessage   `json:"payload"`
}

// IngestBatchDto mirrors TS Items wrapper.
type IngestBatchDto struct {
	NTX   ntxctx.NTXContext `context:"ntx"`
	Items []pipeline.Item   `json:"items"`
}

// IngestSourceParamDto carries the :id path param + arbitrary query map for
// /ingest/:id, /ingest/source/:id and /ingest/channel/:id.
type IngestSourceParamDto struct {
	NTX        ntxctx.NTXContext `context:"ntx"`
	ID         string            `param:"id"`
	Properties map[string]string `query:"-"`
	Payload    json.RawMessage   `json:"-"`
}

// SourceTypeRef mirrors TS SourceTypeDto — nested {key} reference.
type SourceTypeRef struct {
	Key string `json:"key" binding:"required"`
}

// SourceRef mirrors TS SourceDto — the source body inside SetConfigDto.
type SourceRef struct {
	Id         *string         `json:"id,omitempty"`
	Category   string          `json:"category"  binding:"required"`
	Key        *string         `json:"key,omitempty"`
	Visibility *string         `json:"visibility,omitempty"`
	Name       string          `json:"name"      binding:"required"`
	Desc       *string         `json:"desc,omitempty"`
	Detail     *string         `json:"detail,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Status     *string         `json:"status,omitempty"`
}

// SetConfigDto mirrors TS SetConfigDto — drives the 3-way source/sink/channel
// upsert in AppService.SetConfig.
type SetConfigDto struct {
	NTX         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `json:"userId"      binding:"required"`
	ClientId    string            `json:"clientId"    binding:"required"`
	WorkspaceId string            `json:"workspaceId" binding:"required"`
	SourceType  SourceTypeRef     `json:"sourceType"`
	Source      SourceRef         `json:"source"`
	SinkType    *SinkTypeRef      `json:"sinkType,omitempty"`
	Sink        *SinkRef          `json:"sink,omitempty"`
	Type        *string           `json:"type,omitempty"`
	Category    string            `json:"category"    binding:"required"`
	Key         *string           `json:"key,omitempty"`
	Visibility  *string           `json:"visibility,omitempty"`
}

// GetConfigDto carries the lookup query params for GET /config.
type GetConfigDto struct {
	NTX         ntxctx.NTXContext `context:"ntx"`
	UserId      string            `query:"userId"`
	ClientId    string            `query:"clientId"`
	WorkspaceId string            `query:"workspaceId"`
	EntityId    string            `query:"entityId"`
	EntityName  string            `query:"entityName"`
	Type        string            `query:"type"`
	Category    string            `query:"category"`
}

// UpdateConfigDto mirrors TS UpdateConfigDto — patches a Source row's status
// by id.
type UpdateConfigDto struct {
	NTX    ntxctx.NTXContext `context:"ntx"`
	ID     string            `param:"id"`
	Status string            `json:"status" binding:"required"`
}

// UpsertRootSourceDto mirrors TS UpsertRootSourceDto for POST /root-source.
type UpsertRootSourceDto struct {
	NTX        ntxctx.NTXContext `context:"ntx"`
	Id         *string           `json:"id,omitempty"`
	Category   string            `json:"category"  binding:"required"`
	Key        *string           `json:"key,omitempty"`
	Visibility *string           `json:"visibility,omitempty"`
	TypeKey    string            `json:"typeKey"   binding:"required"`
	Name       string            `json:"name"      binding:"required"`
	Desc       *string           `json:"desc,omitempty"`
	Meta       json.RawMessage   `json:"meta,omitempty"`
	Status     *string           `json:"status,omitempty"`
}

// SinkTypeRef is the nested sinkType handle for UpsertRootSink.
type SinkTypeRef struct {
	Key string `json:"key" binding:"required"`
}

// SinkRef is the nested sink body for UpsertRootSink.
type SinkRef struct {
	Id         *string         `json:"id,omitempty"`
	Category   string          `json:"category" binding:"required"`
	Key        *string         `json:"key,omitempty"`
	Visibility *string         `json:"visibility,omitempty"`
	Name       string          `json:"name"     binding:"required"`
	Desc       *string         `json:"desc,omitempty"`
	Detail     *string         `json:"detail,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
}

// UpsertRootSinkDto mirrors TS UpsertRootSinkDto for POST /root-sink.
type UpsertRootSinkDto struct {
	NTX      ntxctx.NTXContext `context:"ntx"`
	Category string            `json:"category"  binding:"required"`
	Type     string            `json:"type"      binding:"required"`
	SinkType SinkTypeRef       `json:"sinkType"`
	Sink     SinkRef           `json:"sink"`
}
