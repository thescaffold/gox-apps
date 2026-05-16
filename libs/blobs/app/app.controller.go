package app

import (
	"encoding/base64"

	"github.com/awesome-goose/goose/io/output"
	"github.com/awesome-goose/goose/types"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// AppController mirrors ntx-apps/libs/blobs/src/app.controller.ts. Endpoints:
//
//	GET  /                  hello
//	POST /upload/init       init (find-or-create file)
//	POST /upload/batch      batch (upload page chunks)
//	POST /upload/verify     verify (file + page count)
//	POST /upload            upload (init + batch in one shot)
//	GET  /download/:id      download (concatenated raw bytes)
type AppController struct {
	appService *AppService   `inject:""`
	lang       *i18n.Service `inject:""`
	log        types.Log     `inject:""`
}

// GetHello mirrors TS @Get() getHello(): success(getHello()) with no envelope.
func (c *AppController) GetHello(dto *HelloDto) types.Output {
	return response.Success(c.appService.GetHello(), "", "", nil)
}

// Init handles POST /upload/init — find-or-create file by (user,client,workspace,type,name).
func (c *AppController) Init(dto *InitDto) types.Output {
	pref := dto.Ctx.Preference
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	if userID == "" || clientID == "" || workspaceID == "" {
		return response.BadRequest(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			"user, client, workspace required",
		)
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
		return response.InternalServerError(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(out,
		c.lang.Translate("apps.blobs.app.title", nil, pref),
		c.lang.Translate("apps.blobs.app.post.remote.success", nil, pref),
		nil)
}

// Batch handles POST /upload/batch — uploads base64-encoded page chunks.
func (c *AppController) Batch(dto *BatchDto) types.Output {
	pref := dto.Ctx.Preference
	if len(dto.Pages) == 0 {
		return response.BadRequest(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			"pages required",
		)
	}
	inputs := make([]BatchInput, 0, len(dto.Pages))
	for _, p := range dto.Pages {
		if p.FileId == nil || p.Index == nil {
			return response.BadRequest(
				c.lang.Translate("apps.blobs.app.title", nil, pref),
				"page fileId and index required",
			)
		}
		raw, err := base64.StdEncoding.DecodeString(p.Raw)
		if err != nil {
			return response.BadRequest(
				c.lang.Translate("apps.blobs.app.title", nil, pref),
				"invalid base64 chunk",
			)
		}
		inputs = append(inputs, BatchInput{FileId: *p.FileId, Index: *p.Index, Raw: raw})
	}
	pages, err := c.appService.Batch(inputs)
	if err != nil {
		return response.InternalServerError(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(pages,
		c.lang.Translate("apps.blobs.app.title", nil, pref),
		c.lang.Translate("apps.blobs.app.post.remote.success", nil, pref),
		nil)
}

// Verify handles POST /upload/verify — confirms upload completion. Mirrors TS
// which uses `apps.blobs.app.post.verify.success` (note distinct from remote).
func (c *AppController) Verify(dto *VerifyDto) types.Output {
	pref := dto.Ctx.Preference
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	if userID == "" || clientID == "" || workspaceID == "" {
		return response.BadRequest(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			"user, client, workspace required",
		)
	}
	out, err := c.appService.Verify(VerifyInput{
		UserId: userID, ClientId: clientID, WorkspaceId: workspaceID,
		Type: dto.Type, Name: dto.Name, ParentId: dto.ParentId,
	})
	if err != nil {
		return response.NotFound(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(out,
		c.lang.Translate("apps.blobs.app.title", nil, pref),
		c.lang.Translate("apps.blobs.app.post.verify.success", nil, pref),
		nil)
}

// Upload handles POST /upload — single-shot file+pages upload.
func (c *AppController) Upload(dto *FileUploadDto) types.Output {
	pref := dto.Ctx.Preference
	userID, clientID, workspaceID := identityFromContext(dto.Ctx)
	if userID == "" || clientID == "" || workspaceID == "" {
		return response.BadRequest(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			"user, client, workspace required",
		)
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
		return response.InternalServerError(
			c.lang.Translate("apps.blobs.app.title", nil, pref),
			err.Error(),
		)
	}
	return response.Success(out,
		c.lang.Translate("apps.blobs.app.title", nil, pref),
		c.lang.Translate("apps.blobs.app.post.remote.success", nil, pref),
		nil)
}

// Download handles GET /download/:id — returns the assembled file bytes.
// Mirrors TS app.controller.ts download(): octet-stream + attachment header.
func (c *AppController) Download(dto *DownloadDto) types.Output {
	data, resolvedID, err := c.appService.Download(dto.Id)
	if err != nil {
		return response.NotFound("blobs", err.Error())
	}
	return output.DownloadFromContent(data, resolvedID+".bin",
		output.WithFileContentType("application/octet-stream"))
}

// identityFromContext extracts user.id, client.id, workspace.id from the
// parsed NTXContext header map, falling back to the flat ID fields when the
// nested objects are absent.
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
