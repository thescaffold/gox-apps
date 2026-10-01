package auth

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

// InitiateDto carries POST /initiate — register step 1. Mirrors TS
// InitiateUserDto (ntx-apps app.dto.ts:270-287), minus `sessionId` (the
// invite flow it threads through is out of scope here — see auth.controller.go's
// package doc).
type InitiateDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	Ref      string            `json:"ref"      binding:"required"`
	Type     string            `json:"type"     binding:"required"`
	ClientId string            `json:"clientId" binding:"required"`
}

// VerifyDto carries POST /verify and POST /reset-secret/verify — both just
// forward the OTP token minted by Initiate/ResetSecretInitiate.
type VerifyDto struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Token string            `json:"token" binding:"required"`
}

// SecretDto carries POST /secret — register step 3 (set password, create the
// account). Mirrors TS SecretUserDto (app.dto.ts:295-342), minus the invite
// fields (workspaceId/roleTypeId) — Origine has no invite flow wired yet.
type SecretDto struct {
	Ctx           ntxctx.NTXContext `context:"ntx"`
	Token         string            `json:"token"          binding:"required"`
	Secret        string            `json:"secret"         binding:"required"`
	ConfirmSecret string            `json:"confirmSecret"  binding:"required"`
	FirstName     string            `json:"firstName"`
	LastName      string            `json:"lastName"`
	ReferralCode  string            `json:"referralCode"`
	Theme         string            `json:"theme"`
	Language      string            `json:"language"`
	Timezone      string            `json:"timezone"`
	Country       string            `json:"country"`
	Currency      string            `json:"currency"`
	Cookie        string            `header:"Cookie"`
}

// LoginDto carries POST /login. Mirrors TS LoginUserDto (app.dto.ts:344-361)
// minus the `x_ntx_device_id` cookie handling — response.Success has no
// header/cookie support in gox (see auth.controller.go's package doc), so a
// device row is bootstrapped fresh on every login instead of being reused
// across sessions via a cookie.
type LoginDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Ref       string            `json:"ref"       binding:"required"`
	Secret    string            `json:"secret"    binding:"required"`
	ClientId  string            `json:"clientId"  binding:"required"`
	UserAgent string            `json:"userAgent"`
	// Cookie is the request's Cookie header; the device the browser already
	// holds (x_ntx_device_id) is reused rather than a new one made per login.
	Cookie string `header:"Cookie"`
}

// ResetSecretUpdateDto carries POST /reset-secret/update. Mirrors TS
// UpdateResetSecretDto (app.dto.ts:440-454).
type ResetSecretUpdateDto struct {
	Ctx              ntxctx.NTXContext `context:"ntx"`
	Token            string            `json:"token"            binding:"required"`
	NewSecret        string            `json:"newSecret"        binding:"required"`
	ConfirmNewSecret string            `json:"confirmNewSecret" binding:"required"`
}

// SwitchWorkspaceDto carries POST /switch (U-S9, PLAN M1-02): re-issue
// tokens scoped to a different workspace the already-authenticated user
// belongs to. Ctx.UserID/ClientID come from the verified Bearer token
// (requireUser()), never the request body — only WorkspaceId is
// client-supplied, and the handler verifies membership before trusting it.
type SwitchWorkspaceDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	WorkspaceId string            `json:"workspaceId" binding:"required"`
}

// AttributesQueryDto carries GET /attributes?type=. Always scoped to the
// current authenticated user (dto.Ctx.UserID) — mirrors TS :1563-1616.
type AttributesQueryDto struct {
	Ctx  ntxctx.NTXContext `context:"ntx"`
	Type string            `query:"type"`
}

// UpdateAttributesDto carries PATCH /attributes/:type. Body is a flat
// {<attrKey>: value} map (mirrors TS :1618-1711) — `,merge` binds the whole
// JSON body into Body rather than a single named field (see goose's
// io/input.Populate, the `json:"...,merge"` case).
type UpdateAttributesDto struct {
	Ctx  ntxctx.NTXContext `context:"ntx"`
	Type string            `param:"type"`
	Body map[string]any    `json:",merge"`
}

// SessionsDto carries GET /sessions. It is answered from the device cookie, not
// a bearer token: the sign-in screens ask which accounts this browser already
// holds before anyone has signed in.
type SessionsDto struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Cookie string            `header:"Cookie"`
}

// LogoutDto carries POST /logout. `device` is "yes" (this browser's device, from
// its cookie), "no", or a device id; userId and clientId narrow what is signed
// out, as TS's LogoutUserDto does.
type LogoutDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	Cookie   string            `header:"Cookie"`
	Device   string            `json:"device"`
	UserId   string            `json:"userId"`
	ClientId string            `json:"clientId"`
}

// UserClientsDto carries GET /user-clients: the signed-in user's apps.
type UserClientsDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// ClientLogRegisterDto carries POST /client-log/register: "this user opened this
// app". The user is always the signed-in one; a userId in the body is ignored.
type ClientLogRegisterDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	ClientId    string            `json:"clientId"    binding:"required"`
	WorkspaceId string            `json:"workspaceId"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
}
