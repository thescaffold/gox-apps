package auth

import (
	"github.com/awesome-goose/goose/types"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/security"
)

// The length a password must have: the sign-up screens' rule (7 to 21
// characters), held here too so this endpoint cannot be used around it.
const (
	secretMin = 7
	secretMax = 21
)

// ChangeSecretDto carries POST /secret/change.
type ChangeSecretDto struct {
	Ctx     ntxctx.NTXContext `context:"ntx"`
	Current string            `json:"current" binding:"required"`
	Secret  string            `json:"secret"  binding:"required"`
}

// ChangeSecret is POST /secret/change: a signed-in person changes their
// password by giving the current one. An account without a password (one made
// with an emailed code or a provider) sets one through the reset flow instead.
func (c *AuthController) ChangeSecret(dto *ChangeSecretDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title, "sign in first")
	}
	if u.Secret == nil || *u.Secret == "" || !security.Compare(dto.Current, *u.Secret) {
		return response.ErrorCode(title, "the current password is not right", 400, "secret_mismatch")
	}
	if n := len(dto.Secret); n < secretMin || n > secretMax {
		return response.ErrorCode(title, "use 7 to 21 characters", 400, "secret_length")
	}
	if dto.Secret == dto.Current {
		return response.ErrorCode(title, "choose a password you have not used here", 400, "secret_same")
	}
	hashed, err := security.Hash(dto.Secret)
	if err != nil {
		return response.InternalServerError(title, "the password could not be changed")
	}
	if _, err := c.userEntity.Update(&identityuser.User{Secret: &hashed}, `"id" = ?`, u.Id); err != nil {
		return response.InternalServerError(title, "the password could not be changed")
	}
	return response.Success(map[string]any{"changed": true}, title, "password changed", nil)
}
