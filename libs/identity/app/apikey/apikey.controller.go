package apikey

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/apikey"
	orgsvc "github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
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

// others reports whether the caller may list and revoke other people's keys in their workspace
// (org.tokens.manage: owners and admins, unless the host's authorizer decides).
func (c *ApiKeyController) others(n ntxctx.NTXContext) (bool, error) {
	return orgsvc.New(c.svc.Keys().DB).Allowed(context.Background(), n.WorkspaceID, n.UserID, orgsvc.ManageTokens)
}

// List returns the caller's own keys; owners and admins see everyone's in the workspace. Values
// are never listed.
func (c *ApiKeyController) List(dto *ListReq) types.Output {
	all, err := c.others(dto.Ctx)
	if err != nil {
		return fail(err)
	}
	keys, err := c.svc.Keys().ListIn(context.Background(), dto.Ctx.ClientID, dto.Ctx.WorkspaceID, dto.Ctx.UserID, all)
	if err != nil {
		return fail(err)
	}
	return response.Success(keys, title, "keys", nil)
}

// Revoke disables a key: the caller's own, or, for owners and admins, anyone's in the workspace.
// Someone else's key is "not found" to everyone else.
func (c *ApiKeyController) Revoke(dto *IDReq) types.Output {
	ctx := context.Background()
	holder, err := c.svc.Keys().HolderIn(ctx, dto.Ctx.ClientID, dto.Ctx.WorkspaceID, dto.Id)
	if err != nil {
		return fail(err)
	}
	if holder != dto.Ctx.UserID {
		ok, err := c.others(dto.Ctx)
		if err != nil {
			return fail(err)
		}
		if !ok {
			return fail(apikey.ErrNotFound)
		}
	}
	if err := c.svc.Keys().RevokeIn(ctx, dto.Ctx.ClientID, dto.Ctx.WorkspaceID, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"id": dto.Id, "revoked": true}, title, "key revoked", nil)
}

// Rotate replaces the caller's own key. Nobody rotates someone else's: the new secret would be
// shown to them.
func (c *ApiKeyController) Rotate(dto *IDReq) types.Output {
	ctx := context.Background()
	holder, err := c.svc.Keys().HolderIn(ctx, dto.Ctx.ClientID, dto.Ctx.WorkspaceID, dto.Id)
	if err != nil {
		return fail(err)
	}
	if holder != dto.Ctx.UserID {
		if ok, err := c.others(dto.Ctx); err != nil {
			return fail(err)
		} else if ok {
			return response.Forbidden(title, "only its holder rotates a key; revoke it instead")
		}
		return fail(apikey.ErrNotFound)
	}
	secs, _ := strconv.Atoi(dto.Grace)
	k, err := c.svc.Keys().RotateIn(ctx, dto.Ctx.UserID, dto.Ctx.ClientID, dto.Ctx.WorkspaceID, dto.Id, time.Duration(secs)*time.Second)
	if err != nil {
		return fail(err)
	}
	return response.Created(k, title, "key rotated; copy the new key now")
}
