package app

import (
	"encoding/json"
)

// HealthDto is the body for the GET / endpoint.
type HealthDto struct{}

// InitDto mirrors TS InitDto (== CreateFileDto): all fields used to find-or-create a file.
type InitDto struct {
	Type     string          `json:"type"  binding:"required"`
	ParentId *string         `json:"parentId,omitempty"`
	Name     string          `json:"name"  binding:"required"`
	Tags     []string        `json:"tags,omitempty"`
	Size     int64           `json:"size"  binding:"required"`
	Mime     string          `json:"mime"  binding:"required"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
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
	Pages []PageInput `json:"pages"`
}

// VerifyDto mirrors TS VerifyDto (== CreateFileDto subset used for verify).
type VerifyDto struct {
	Type     string          `json:"type"  binding:"required"`
	ParentId *string         `json:"parentId,omitempty"`
	Name     string          `json:"name"  binding:"required"`
	Tags     []string        `json:"tags,omitempty"`
	Size     int64           `json:"size"  binding:"required"`
	Mime     string          `json:"mime"  binding:"required"`
	Meta     json.RawMessage `json:"meta,omitempty"`
	Status   *string         `json:"status,omitempty"`
}

// FileUploadDto mirrors TS FileUploadDto — combines file + pages in one POST.
type FileUploadDto struct {
	File  InitDto     `json:"file"`
	Pages []PageInput `json:"pages,omitempty"`
}

// DownloadDto carries the path :id for GET /download/:id.
type DownloadDto struct {
	Id string `param:"id" binding:"required"`
}
