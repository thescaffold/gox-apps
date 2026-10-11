package auth

import (
	"context"
	"strings"
	"time"

	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/revocation"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// CloseAccountDto carries POST /account/close.
type CloseAccountDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
	// Confirm is the account's email address, typed.
	Confirm string `json:"confirm" binding:"required"`
}

// CloseAccount closes the caller's Envelope account: it can no longer sign
// in, its devices and refresh tokens are dropped, and every access token
// already issued stops working. It does not delete mail or companies: the
// Mail Gateway's own deletion does the former, and a company the person owns
// must be handed over first (the console says so).
func (c *AuthController) CloseAccount(dto *CloseAccountDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title, "sign in first")
	}
	if !strings.EqualFold(strings.TrimSpace(dto.Confirm), u.Email) {
		return response.BadRequest(title, "type your email address to confirm")
	}
	c.upsertAttribute(u.Id, "account", "closed_at", time.Now().UTC().Format(time.RFC3339))
	if c.deviceSessionEntity != nil {
		_, _ = c.deviceSessionEntity.Delete(`"user_id" = ?`, u.Id)
	}
	if c.db != nil {
		_ = c.db.DB.Exec(`DELETE FROM "IdentityTokens" WHERE "user_id" = ?`, u.Id).Error
	}
	if _, err := revocation.RevokeUser(context.Background(), u.Id, "closed"); err != nil {
		return response.InternalServerError(title, "could not end your sessions")
	}
	return response.Success(map[string]any{"closed": true}, title, "account closed", nil)
}
