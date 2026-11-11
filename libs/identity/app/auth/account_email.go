package auth

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/awesome-goose/goose/types"
	"github.com/google/uuid"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityotp "github.com/thescaffold/gox-apps/libs/identity/pkg/otp"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Changes a person makes to their own account: their name, and their sign-in email. The email
// changes only once a code sent to the new address is entered, and the old address is told, so a
// borrowed session cannot move an account quietly.

// UpdateAccountDto carries PATCH /account.
type UpdateAccountDto struct {
	Ctx  ntxctx.NTXContext `context:"ntx"`
	Name string            `json:"name"`
}

// EmailChangeDto carries POST /account/email (Email) and POST /account/email/verify (Email, Code).
type EmailChangeDto struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Email string            `json:"email"`
	Code  string            `json:"code"`
}

// emailChangeClient keeps a person's email-change codes apart from sign-in codes for the same
// address: the code store is keyed by (address, client).
func emailChangeClient(userID string) string { return "email-change:" + userID }

// UpdateAccount is PATCH /account {name}: a person renames themself.
func (c *AuthController) UpdateAccount(dto *UpdateAccountDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if dto.Ctx.UserID == "" {
		return response.Unauthorized(title, "sign in first")
	}
	name := strings.Join(strings.Fields(dto.Name), " ")
	if n := utf8.RuneCountInString(name); n < 1 || n > 80 {
		return response.ErrorCode(title, "a name is 1 to 80 characters", 400, "invalid")
	}
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title, "sign in first")
	}
	if _, err := c.userEntity.Update(&identityuser.User{Name: name}, `"id" = ?`, u.Id); err != nil {
		return response.InternalServerError(title, "the name could not be saved")
	}
	u.Name = name
	return response.Success(maskUser(u), title, "saved", nil)
}

// emailTaken reports whether ref is another account's sign-in address.
func (c *AuthController) emailTaken(ref, self string) bool {
	u, _ := c.userEntity.First(`("ref" = ? OR lower("email") = ?) AND "id" <> ?`, ref, ref, self)
	return u != nil
}

// EmailChangeStart is POST /account/email {email}: a code goes to the new address. Nothing changes
// until it is entered.
func (c *AuthController) EmailChangeStart(dto *EmailChangeDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if dto.Ctx.UserID == "" || c.signin == nil {
		return response.Unauthorized(title, "sign in first")
	}
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title, "sign in first")
	}
	ref, typ, err := identityotp.Normalize(dto.Email)
	if err != nil || typ != identityotp.Email {
		return response.ErrorCode(title, "enter an email address", 400, "invalid")
	}
	if strings.EqualFold(ref, u.Email) {
		return response.ErrorCode(title, "that is already your email address", 400, "invalid")
	}
	if c.emailTaken(ref, u.Id) {
		return response.ErrorCode(title, "that address belongs to another account", 409, "email_taken")
	}
	if sso.Enforced(context.Background(), ref) {
		return response.ErrorCode(title, "that address's organization signs in with single sign-on; ask its admin", 400, "sso_enforced")
	}
	code, err := c.signin.Issue(context.Background(), ref, emailChangeClient(u.Id))
	switch {
	case errors.Is(err, identityotp.ErrThrottled), errors.Is(err, identityotp.ErrTooMany):
		return response.Error(title, err.Error(), 429)
	case err != nil:
		return response.InternalServerError(title, "the code could not be sent")
	}
	if c.tracker != nil {
		c.tracker.Message("apps.notification.message.new", map[string]any{
			"reference": "email-change-" + uuid.NewString(),
			"userId":    u.Id,
			"clientId":  dto.Ctx.ClientID,
			"key":       "email-change",
			"channels":  []string{"email"},
			"data":      map[string]any{"code": code, "ref": ref, "owner": ref},
			"subject":   "Confirm your new email address",
			"type":      "security",
			"priority":  "high",
			"scope":     "user",
		})
	}
	return response.Success(map[string]any{"sent": true, "email": ref}, title, "a code is on its way to the new address", nil)
}

// EmailChangeVerify is POST /account/email/verify {email, code}: the address changes, and the old
// one is told.
func (c *AuthController) EmailChangeVerify(dto *EmailChangeDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	if dto.Ctx.UserID == "" || c.signin == nil || c.db == nil {
		return response.Unauthorized(title, "sign in first")
	}
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title, "sign in first")
	}
	ref, typ, err := identityotp.Normalize(dto.Email)
	if err != nil || typ != identityotp.Email {
		return response.ErrorCode(title, "enter an email address", 400, "invalid")
	}
	if err := c.signin.Check(context.Background(), ref, emailChangeClient(u.Id), dto.Code); err != nil {
		if errors.Is(err, identityotp.ErrLocked) {
			return response.Error(title, err.Error(), 429)
		}
		return response.Unauthorized(title, "the code is invalid or has expired")
	}
	if c.emailTaken(ref, u.Id) {
		return response.ErrorCode(title, "that address belongs to another account", 409, "email_taken")
	}
	old := u.Email
	if err := c.db.DB.Exec(`UPDATE "IdentityUsers" SET "email" = ?, "ref" = ?, "updated_at" = now() WHERE "id" = ?`, ref, ref, u.Id).Error; err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return response.ErrorCode(title, "that address belongs to another account", 409, "email_taken")
		}
		return response.InternalServerError(title, "the email could not be changed")
	}
	c.upsertAttribute(u.Id, "user", "email", ref)
	u.Email, u.Ref = ref, &ref
	if c.tracker != nil && old != "" {
		c.tracker.Message("apps.notification.message.new", map[string]any{
			"reference": "email-changed-" + uuid.NewString(),
			"userId":    u.Id,
			"clientId":  dto.Ctx.ClientID,
			"key":       "email-changed",
			"channels":  []string{"email"},
			"data":      map[string]any{"ref": ref, "old": old, "owner": old},
			"subject":   "Your sign-in email was changed",
			"type":      "security",
			"priority":  "high",
			"scope":     "user",
		})
	}
	return response.Success(maskUser(u), title, "email changed", nil)
}
