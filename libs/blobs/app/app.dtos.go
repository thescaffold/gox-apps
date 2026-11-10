package app

import (
	"encoding/json"

	"github.com/awesome-goose/goose/openapi"
	"github.com/thescaffold/gox-apps/libs/blobs/app/page"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// HelloDto carries the request context for GET / (getHello).
type HelloDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// InitDto mirrors TS InitDto (== CreateFileDto): all fields used to find-or-create a file.
type InitDto struct {
	Type     string            `json:"type"  binding:"required"`
	ParentId *string           `json:"parentId,omitempty"`
	Name     string            `json:"name"  binding:"required"`
	Tags     []string          `json:"tags,omitempty"`
	Size     int64             `json:"size"  binding:"required"`
	Mime     string            `json:"mime"  binding:"required"`
	Meta     json.RawMessage   `json:"meta,omitempty"`
	Status   *string           `json:"status,omitempty"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// PageInput is the wire shape for one chunk in BatchDto/FileUploadDto pages.
type PageInput struct {
	FileId *string `json:"fileId,omitempty"`
	Index  *int    `json:"index,omitempty"`
	// Raw is a base64-encoded chunk (matches TS Buffer.from(raw, 'base64')).
	Raw    string  `json:"raw"             binding:"required"`
	Status *string `json:"status,omitempty"`
}

// BatchDto mirrors TS BatchDto.
type BatchDto struct {
	Pages []PageInput       `json:"pages"`
	Ctx   ntxctx.NTXContext `context:"ntx"`
}

// VerifyDto mirrors TS VerifyDto (== CreateFileDto subset used for verify).
type VerifyDto struct {
	Type     string            `json:"type"  binding:"required"`
	ParentId *string           `json:"parentId,omitempty"`
	Name     string            `json:"name"  binding:"required"`
	Tags     []string          `json:"tags,omitempty"`
	Size     int64             `json:"size"  binding:"required"`
	Mime     string            `json:"mime"  binding:"required"`
	Meta     json.RawMessage   `json:"meta,omitempty"`
	Status   *string           `json:"status,omitempty"`
	Ctx      ntxctx.NTXContext `context:"ntx"`
}

// FileUploadDto mirrors TS FileUploadDto — combines file + pages in one POST.
// Note: the nested `file` body uses InitDto without its Ctx field — the
// outer dto's Ctx covers context for the whole request.
type FileUploadDto struct {
	File  InitDto           `json:"file"`
	Pages []PageInput       `json:"pages,omitempty"`
	Ctx   ntxctx.NTXContext `context:"ntx"`
}

// DownloadDto carries the path :id for GET /download/:id.
type DownloadDto struct {
	Id  string            `param:"id" binding:"required"`
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// OpenAPI documents what each upload route answers (PLAN U25).
func (InitDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "Find or create a file in the caller's workspace", Response: InitOutput{}}
}

func (BatchDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "Upload base64 pages of the caller's files", Response: []page.Page{}}
}

func (VerifyDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "A file of the caller's workspace and how many pages it has", Response: VerifyOutput{}}
}

func (FileUploadDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "Create a file and its pages in one call", Response: UploadOutput{}}
}

func (DownloadDto) OpenAPI() openapi.Operation {
	return openapi.Operation{Summary: "The bytes of a file of the caller's workspace (application/octet-stream)"}
}
