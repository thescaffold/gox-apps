package apikey

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/apikey"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

type ApiKeyController struct {
	svc *ApiKeyService `inject:""`
}

const title = "api keys"

func fail(err error) types.Output {
	switch {
	case errors.Is(err, apikey.ErrScopeRequired):
		return response.Unauthorized(title, err.Error())
	case errors.Is(err, apikey.ErrNotFound):
		return response.NotFound(title, "api key not found")
	case errors.Is(err, apikey.ErrLiveLocked):
		return response.Forbidden(title, err.Error())
	case errors.Is(err, apikey.ErrInvalid):
		return response.BadRequest(title, err.Error())
	}
	return response.InternalServerError(title, "the key could not be changed")
}

func (c *ApiKeyController) Issue(dto *IssueReq) types.Output {
	kind := dto.Kind
	if kind == "" {
		kind = apikey.Secret
	}
	env := dto.Environment
	if env == "" {
		env = apikey.Test
	}
	var ttl time.Duration
	if dto.ExpiresIn > 0 {
		ttl = time.Duration(dto.ExpiresIn) * time.Second
	}
	k, err := c.svc.Keys().Issue(context.Background(), dto.Ctx.UserID, dto.Ctx.ClientID, dto.Ctx.WorkspaceID, dto.Name, kind, env, dto.Scopes, ttl)
	if err != nil {
		return fail(err)
	}
	return response.Created(k, title, "key issued; copy it now, it is not shown again")
}

func (c *ApiKeyController) List(dto *ListReq) types.Output {
	keys, err := c.svc.Keys().List(context.Background(), dto.Ctx.ClientID, dto.Ctx.WorkspaceID)
	if err != nil {
		return fail(err)
	}
	return response.Success(keys, title, "keys", nil)
}

func (c *ApiKeyController) Revoke(dto *IDReq) types.Output {
	if err := c.svc.Keys().Revoke(context.Background(), dto.Ctx.ClientID, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"id": dto.Id, "revoked": true}, title, "key revoked", nil)
}

func (c *ApiKeyController) Rotate(dto *IDReq) types.Output {
	secs, _ := strconv.Atoi(dto.Grace)
	k, err := c.svc.Keys().Rotate(context.Background(), dto.Ctx.UserID, dto.Ctx.ClientID, dto.Id, time.Duration(secs)*time.Second)
	if err != nil {
		return fail(err)
	}
	return response.Created(k, title, "key rotated; copy the new key now")
}
