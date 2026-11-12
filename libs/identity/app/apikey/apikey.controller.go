package apikey

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

// expiryFail answers a key refused for its expiry: 422, with the workspace's maximum.
func (c *ApiKeyController) expiryFail(ws string, err error) (types.Output, bool) {
	code := ""
	switch {
	case errors.Is(err, apikey.ErrExpiryRequired):
		code = "expiry_required"
	case errors.Is(err, apikey.ErrExpiryTooLong):
		code = "expiry_too_long"
	default:
		return nil, false
	}
	max, _ := apikey.MaxDays(context.Background(), c.svc.Keys().DB, ws)
	msg := fmt.Sprintf("Every token expires: choose an expiry of at most %d days.", max)
	return response.ErrorDetails(title, msg, http.StatusUnprocessableEntity, code, map[string]any{"maxDays": max}), true
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
	if out, ok := c.expiryFail(dto.Ctx.WorkspaceID, err); ok {
		return out
	}
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

// Policy is GET /keys/policy: the workspace's maximum key life, whether keys must expire, and
// whether the caller may change it (org.security.manage).
func (c *ApiKeyController) Policy(dto *ListReq) types.Output {
	if dto.Ctx.WorkspaceID == "" {
		return fail(apikey.ErrScopeRequired)
	}
	ctx := context.Background()
	max, err := apikey.MaxDays(ctx, c.svc.Keys().DB, dto.Ctx.WorkspaceID)
	if err != nil {
		return fail(err)
	}
	can, err := orgsvc.New(c.svc.Keys().DB).Allowed(ctx, dto.Ctx.WorkspaceID, dto.Ctx.UserID, orgsvc.ManageSecurity)
	if err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{
		"maxDays": max, "defaultMaxDays": apikey.DefaultMaxDays, "required": c.svc.Keys().MaxTTL != nil, "canManage": can,
	}, title, "key policy", nil)
}

// SetPolicy is PUT /keys/policy {maxDays}: owners and admins (org.security.manage) set the
// longest a new key may live, 1 to 365 days. Keys already issued keep their expiry.
func (c *ApiKeyController) SetPolicy(dto *PolicyReq) types.Output {
	if dto.Ctx.WorkspaceID == "" {
		return fail(apikey.ErrScopeRequired)
	}
	ctx := context.Background()
	db := c.svc.Keys().DB
	can, err := orgsvc.New(db).Allowed(ctx, dto.Ctx.WorkspaceID, dto.Ctx.UserID, orgsvc.ManageSecurity)
	if err != nil {
		return fail(err)
	}
	if !can {
		return response.Forbidden(title, "only organization owners and admins set how long tokens may live")
	}
	if err := apikey.SetMaxDays(ctx, db, dto.Ctx.WorkspaceID, dto.Ctx.UserID, dto.MaxDays); err != nil {
		if errors.Is(err, apikey.ErrInvalid) {
			return response.ErrorDetails(title, fmt.Sprintf("Choose 1 to %d days.", apikey.DefaultMaxDays), http.StatusUnprocessableEntity, "invalid_max_days", map[string]any{"maxDays": apikey.DefaultMaxDays})
		}
		return fail(err)
	}
	return c.Policy(&ListReq{Ctx: dto.Ctx})
}
