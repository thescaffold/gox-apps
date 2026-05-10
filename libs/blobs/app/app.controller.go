package app

import (
	"encoding/base64"

	"github.com/awesome-goose/goose/io/output"
	"github.com/awesome-goose/goose/types"
	ntxctx "github.com/thescaffold/gox-packages-core/context"
	"github.com/thescaffold/gox-packages-core/response"
)

type AppController struct {
	appService *AppService `inject:""`
	log        types.Log   `inject:""`
}

// Health is the GET / handler.
func (c *AppController) Health(dto *HealthDto) types.Output {
	return response.Success(c.appService.GetHello(), "blobs", "ok", nil)
}

// Init handles POST /upload/init — creates (or returns) a file record.
func (c *AppController) Init(ctx types.Context, dto *InitDto) types.Output {
	ntx := ntxctx.Get(ctx)
	userID, clientID, workspaceID := identityFromContext(ntx)
	if userID == "" || clientID == "" || workspaceID == "" {
		return response.Unauthorized("blobs", "user, client, workspace required")
	}
	out, err := c.appService.Init(InitInput{
		UserId:      userID,
		ClientId:    clientID,
		WorkspaceId: workspaceID,
		Type:        dto.Type,
		Name:        dto.Name,
		ParentId:    dto.ParentId,
		Tags:        dto.Tags,
		Size:        dto.Size,
		Mime:        dto.Mime,
		Status:      dto.Status,
		Meta:        dto.Meta,
	})
	if err != nil {
		return response.InternalServerError("blobs", err.Error())
	}
	return response.Success(out, "blobs", "ok", nil)
}

// Batch handles POST /upload/batch — uploads base64-encoded page chunks.
func (c *AppController) Batch(ctx types.Context, dto *BatchDto) types.Output {
	if len(dto.Pages) == 0 {
		return response.BadRequest("blobs", "pages required")
	}
	inputs := make([]BatchInput, 0, len(dto.Pages))
	for _, p := range dto.Pages {
		if p.FileId == nil || p.Index == nil {
			return response.BadRequest("blobs", "page fileId and index required")
		}
		raw, err := base64.StdEncoding.DecodeString(p.Raw)
		if err != nil {
			return response.BadRequest("blobs", "invalid base64 chunk")
		}
		inputs = append(inputs, BatchInput{FileId: *p.FileId, Index: *p.Index, Raw: raw})
	}
	pages, err := c.appService.Batch(inputs)
	if err != nil {
		return response.InternalServerError("blobs", err.Error())
	}
	return response.Success(pages, "blobs", "ok", nil)
}

// Verify handles POST /upload/verify — confirms upload completion.
func (c *AppController) Verify(ctx types.Context, dto *VerifyDto) types.Output {
	ntx := ntxctx.Get(ctx)
	userID, clientID, workspaceID := identityFromContext(ntx)
	if userID == "" || clientID == "" || workspaceID == "" {
		return response.Unauthorized("blobs", "user, client, workspace required")
	}
	out, err := c.appService.Verify(VerifyInput{
		UserId: userID, ClientId: clientID, WorkspaceId: workspaceID,
		Type: dto.Type, Name: dto.Name, ParentId: dto.ParentId,
	})
	if err != nil {
		return response.NotFound("blobs", err.Error())
	}
	return response.Success(out, "blobs", "ok", nil)
}

// Upload handles POST /upload — single-shot file+pages upload.
func (c *AppController) Upload(ctx types.Context, dto *FileUploadDto) types.Output {
	ntx := ntxctx.Get(ctx)
	userID, clientID, workspaceID := identityFromContext(ntx)
	if userID == "" || clientID == "" || workspaceID == "" {
		return response.Unauthorized("blobs", "user, client, workspace required")
	}
	in := InitInput{
		UserId:      userID,
		ClientId:    clientID,
		WorkspaceId: workspaceID,
		Type:        dto.File.Type,
		Name:        dto.File.Name,
		ParentId:    dto.File.ParentId,
		Tags:        dto.File.Tags,
		Size:        dto.File.Size,
		Mime:        dto.File.Mime,
		Status:      dto.File.Status,
		Meta:        dto.File.Meta,
	}
	raws := make([]string, 0, len(dto.Pages))
	for _, p := range dto.Pages {
		raws = append(raws, p.Raw)
	}
	out, err := c.appService.Upload(in, raws)
	if err != nil {
		return response.InternalServerError("blobs", err.Error())
	}
	return response.Success(out, "blobs", "ok", nil)
}

// Download handles GET /download/:id — returns the assembled file bytes.
// Mirrors TS app.controller.ts download(): octet-stream + attachment header.
func (c *AppController) Download(ctx types.Context, dto *DownloadDto) types.Output {
	data, resolvedID, err := c.appService.Download(dto.Id)
	if err != nil {
		return response.NotFound("blobs", err.Error())
	}
	return output.DownloadFromContent(data, resolvedID+".bin",
		output.WithFileContentType("application/octet-stream"))
}

// identityFromContext extracts user.id, client.id, workspace.id from the
// parsed NTXContext header map.
func identityFromContext(ntx ntxctx.NTXContext) (string, string, string) {
	var userID, clientID, workspaceID string
	if ntx.User != nil {
		userID, _ = ntx.User["id"].(string)
	}
	if userID == "" {
		userID = ntx.UserID
	}
	if ntx.Client != nil {
		clientID, _ = ntx.Client["id"].(string)
	}
	if clientID == "" {
		clientID = ntx.ClientID
	}
	if ntx.Workspace != nil {
		workspaceID, _ = ntx.Workspace["id"].(string)
	}
	if workspaceID == "" {
		workspaceID = ntx.WorkspaceID
	}
	return userID, clientID, workspaceID
}
