package auth

import (
	"context"
	"errors"

	"github.com/awesome-goose/goose/types"
	"github.com/google/uuid"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	identityworkspace "github.com/thescaffold/gox-apps/libs/identity/app/workspace"
	identityotp "github.com/thescaffold/gox-apps/libs/identity/pkg/otp"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/sso"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Passwordless sign-in (PRD "Flexible onboarding"): one step asks for a code by
// email or SMS, the next verifies it and returns tokens, creating the account
// on first use. There is no password and no separate registration.

// OtpRequestDto carries POST /otp.
type OtpRequestDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	Ref      string            `json:"ref"      binding:"required"` // email or phone
	ClientId string            `json:"clientId" binding:"required"`
}

// OtpVerifyDto carries POST /otp/verify.
type OtpVerifyDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Ref       string            `json:"ref"       binding:"required"`
	Code      string            `json:"code"      binding:"required"`
	ClientId  string            `json:"clientId"  binding:"required"`
	UserAgent string            `json:"userAgent"`
	Cookie    string            `header:"Cookie"`
	Country   string            `json:"country"`
	Currency  string            `json:"currency"`
	Timezone  string            `json:"timezone"`
	Language  string            `json:"language"`
}

// OtpRequest sends a one-time code. The answer is the same whether or not the
// account exists, so the endpoint cannot be used to find out who has one.
func (c *AuthController) OtpRequest(dto *OtpRequestDto) types.Output {
	title := c.lang.Translate("apps.identity.app.title", nil, dto.Ctx.Preference)
	ref, typ, err := identityotp.Normalize(dto.Ref)
	if err != nil {
		return response.BadRequest(title, err.Error())
	}
	if sso.Enforced(context.Background(), ref) {
		return c.ssoEnforcedResponse(dto.Ctx.Preference)
	}
	code, err := c.signin.Issue(context.Background(), ref, dto.ClientId)
	switch {
	case errors.Is(err, identityotp.ErrThrottled), errors.Is(err, identityotp.ErrTooMany):
		return response.Error(title, err.Error(), 429)
	case err != nil:
		return response.InternalServerError(title, "the code could not be sent")
	}
	userID := uuid.NewString() // notifications need an owner id; the real one is unknown until verify
	if u, _ := c.userEntity.First(`"ref" = ? OR "email" = ?`, ref, ref); u != nil {
		userID = u.Id
	}
	channel := "email"
	if typ == identityotp.Phone {
		channel = "sms"
	}
	if c.tracker != nil {
		c.tracker.Message("apps.notification.message.new", map[string]any{
			"reference": "otp-" + uuid.NewString(),
			"userId":    userID,
			"clientId":  dto.ClientId,
			"key":       "otp-login",
			"channels":  []string{channel},
			"data":      map[string]any{"code": code, "ref": ref, "owner": ref},
			"subject":   "Your sign-in code",
			"type":      "security",
			"priority":  "high",
			"scope":     "user",
		})
	}
	return response.Success(map[string]any{"sent": true, "type": typ}, title, "if the address is valid, a code is on its way", nil)
}

// OtpVerify checks the code and signs the user in, creating the account when
// this is the first time the reference is seen.
func (c *AuthController) OtpVerify(dto *OtpVerifyDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.app.title", nil, pref)
	ref, typ, nerr := identityotp.Normalize(dto.Ref)
	if nerr != nil {
		return response.BadRequest(title, nerr.Error())
	}
	if sso.Enforced(context.Background(), ref) {
		return c.ssoEnforcedResponse(pref)
	}
	if err := c.signin.Check(context.Background(), ref, dto.ClientId, dto.Code); err != nil {
		if errors.Is(err, identityotp.ErrLocked) {
			return response.Error(title, err.Error(), 429)
		}
		if errors.Is(err, identityotp.ErrInvalid) || errors.Is(err, identityotp.ErrExpired) {
			return response.Unauthorized(title, "the code is invalid or has expired")
		}
		return response.InternalServerError(title, "the code could not be checked")
	}

	u, _ := c.userEntity.First(`"ref" = ? OR "email" = ?`, ref, ref)
	isNew := u == nil
	var ws *identityworkspace.Workspace
	if isNew {
		r, t := ref, typ
		u = &identityuser.User{Ref: &r, Type: &t, Name: ref, Email: ref}
		if typ == identityotp.Phone {
			p := ref
			u.Phone = &p
		}
		if err := c.userEntity.Insert(u); err != nil {
			return response.InternalServerError(title, err.Error())
		}
		ws = c.bootstrapWorkspace(u, dto.ClientId)
		c.upsertAttributes(u.Id, "user", map[string]string{typ: ref})
		c.upsertAttributes(u.Id, "preference", map[string]string{
			"country": dto.Country, "currency": dto.Currency, "timezone": dto.Timezone, "language": dto.Language,
		})
		c.seedDefaultRoles()
	} else {
		if closed, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "closed_at"); closed != nil {
			return response.Unauthorized(title, "the code is invalid or has expired")
		}
		if suspended, _ := c.attributeEntity.First(`"user_id" = ? AND "key" = ?`, u.Id, "suspended_at"); suspended != nil {
			return response.Unauthorized(title, c.lang.Translate("apps.identity.app.post.login.error.suspended", nil, pref))
		}
	}

	workspaces := c.workspacesForUser(u.Id, dto.ClientId)
	if isNew && ws != nil && len(workspaces) == 0 {
		workspaces = []identityworkspace.Workspace{*ws}
	}
	workspaceID := ""
	if len(workspaces) > 0 {
		workspaceID = workspaces[0].Id
	}
	access, refresh, err := c.auth.IssueTokens(u, dto.ClientId, workspaceID)
	if err != nil {
		return response.Unauthorized(title, err.Error())
	}
	deviceID := ""
	if len(workspaces) > 0 {
		var sessionID string
		deviceID, sessionID = c.bootstrapDevice(u.Id, dto.ClientId, dto.UserAgent, cookieValue(dto.Cookie, DeviceCookie))
		c.logDeviceLogin(deviceID, sessionID, dto.Ctx)
	}
	return withDeviceCookie(response.Success(c.sessionPayload(u, dto.ClientId, map[string]any{
		"user":         maskUser(u),
		"accessToken":  access,
		"refreshToken": refresh,
		"isNew":        isNew,
	}), title, "signed in", nil), deviceID)
}
