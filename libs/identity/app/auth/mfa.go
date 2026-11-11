package auth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/awesome-goose/goose/types"
	identityuser "github.com/thescaffold/gox-apps/libs/identity/app/user"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/mfa"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/response"
)

// Two-factor sign-in (U-S20, pkg/mfa). The sign-in endpoints call secondStep
// once the first factor has passed: for a person with 2FA they answer with a
// challenge, no tokens and no device session, and MfaVerify finishes the
// sign-in. The rest is the person's own setup (authenticator app, security
// keys, recovery codes) and the workspace policy.

// MfaVerifyDto carries POST /mfa/verify: the challenge from the first step
// and one proof.
type MfaVerifyDto struct {
	Ctx          ntxctx.NTXContext `context:"ntx"`
	Challenge    string            `json:"challenge" binding:"required"`
	Code         string            `json:"code"`
	RecoveryCode string            `json:"recoveryCode"`
	Assertion    json.RawMessage   `json:"assertion"`
	Cookie       string            `header:"Cookie"`
}

// MfaChallengeDto carries POST /mfa/verify/key.
type MfaChallengeDto struct {
	Ctx       ntxctx.NTXContext `context:"ntx"`
	Challenge string            `json:"challenge" binding:"required"`
}

// MfaDto is a signed-in person's request with no body.
type MfaDto struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

// MfaProofDto is a signed-in person's request that must re-verify first
// (once they have 2FA): a challenge from POST /mfa/challenge and one proof.
type MfaProofDto struct {
	Ctx          ntxctx.NTXContext `context:"ntx"`
	Id           string            `param:"id"`
	Challenge    string            `json:"challenge"`
	Code         string            `json:"code"`
	RecoveryCode string            `json:"recoveryCode"`
	Assertion    json.RawMessage   `json:"assertion"`
}

func (d *MfaProofDto) proof() mfa.Proof {
	return mfa.Proof{Code: d.Code, RecoveryCode: d.RecoveryCode, Assertion: d.Assertion}
}

// MfaTotpConfirmDto carries POST /mfa/totp/confirm.
type MfaTotpConfirmDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	FactorId string            `json:"factorId" binding:"required"`
	Code     string            `json:"code"     binding:"required"`
}

// MfaKeyConfirmDto carries POST /mfa/keys/confirm.
type MfaKeyConfirmDto struct {
	Ctx        ntxctx.NTXContext `context:"ntx"`
	Token      string            `json:"token"      binding:"required"`
	Name       string            `json:"name"`
	Credential json.RawMessage   `json:"credential" binding:"required"`
}

// MfaPolicyDto carries PUT /mfa/policy.
type MfaPolicyDto struct {
	Ctx     ntxctx.NTXContext `context:"ntx"`
	Require bool              `json:"require"`
}

func (c *AuthController) bootMfa() {
	if c.mfa != nil || c.db == nil {
		return
	}
	keys, err := mfa.KeysFromEnv()
	if err != nil {
		keys = nil // security keys stay off; the authenticator app still works
	}
	c.mfa = &mfa.Service{DB: c.db.DB, Keys: keys}
	mfa.SetDefault(c.mfa)
}

// mfaTitle is the response title for this surface.
func (c *AuthController) mfaTitle(pref map[string]any) string {
	return c.lang.Translate("apps.identity.app.title", nil, pref)
}

// mfaError answers a pkg/mfa error with its status and a stable code.
func mfaError(title string, err error) types.Output {
	switch {
	case errors.Is(err, mfa.ErrInvalidCode):
		return response.ErrorCode(title, err.Error(), 401, "mfa_invalid")
	case errors.Is(err, mfa.ErrChallengeExpired):
		return response.ErrorCode(title, err.Error(), 401, "mfa_expired")
	case errors.Is(err, mfa.ErrTooManyAttempts):
		return response.ErrorCode(title, err.Error(), 429, "mfa_locked")
	case errors.Is(err, mfa.ErrNoProof):
		return response.ErrorCode(title, err.Error(), 403, "mfa_proof_required")
	case errors.Is(err, mfa.ErrNotEnrolled), errors.Is(err, mfa.ErrEnrollFirst), errors.Is(err, mfa.ErrRequiredByWorkspace):
		return response.ErrorCode(title, err.Error(), 409, "mfa_conflict")
	case errors.Is(err, mfa.ErrKeysUnavailable):
		return response.ErrorCode(title, err.Error(), 501, "mfa_keys_unavailable")
	case errors.Is(err, mfa.ErrNotFound):
		return response.NotFound(title, "that factor was not found")
	}
	msg := err.Error()
	if strings.HasPrefix(msg, "the security key") || strings.HasPrefix(msg, "this security key") {
		return response.BadRequest(title, msg)
	}
	return response.InternalServerError(title, "two-factor authentication failed; try again")
}

// secondStep stops a sign-in whose first factor passed when the person has
// 2FA: it answers with a challenge and the methods they can use. The bool is
// false when there is no second step.
func (c *AuthController) secondStep(u *identityuser.User, clientID, userAgent string, pref map[string]any) (types.Output, bool) {
	if c.mfa == nil {
		return nil, false
	}
	title := c.mfaTitle(pref)
	on, err := c.mfa.Enrolled(context.Background(), u.Id)
	if err != nil {
		return response.InternalServerError(title, "the sign-in could not be completed"), true
	}
	if !on {
		return nil, false
	}
	p, err := c.mfa.StartSignIn(context.Background(), u.Id, clientID, userAgent)
	if err != nil {
		return response.InternalServerError(title, "the sign-in could not be completed"), true
	}
	return response.Success(map[string]any{
		"mfaRequired": true,
		"challenge":   p.Challenge,
		"methods":     p.Methods,
		"expiresAt":   p.ExpiresAt,
	}, title, "enter your second factor", nil), true
}

// MfaVerify is POST /mfa/verify: the second step of a sign-in. With a valid
// proof it signs the person in exactly as the first step would have.
func (c *AuthController) MfaVerify(dto *MfaVerifyDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.mfaTitle(pref)
	if c.mfa == nil {
		return response.InternalServerError(title, "two-factor authentication is not available")
	}
	done, err := c.mfa.CompleteSignIn(context.Background(), dto.Challenge, mfa.Proof{Code: dto.Code, RecoveryCode: dto.RecoveryCode, Assertion: dto.Assertion})
	if err != nil {
		return mfaError(title, err)
	}
	u, _ := c.userEntity.First(`"id" = ?`, done.UserID)
	if u == nil || !c.usable(u.Id) {
		return mfaError(title, mfa.ErrChallengeExpired)
	}
	workspaces := c.workspacesForUser(u.Id, done.ClientID)
	workspaceID := ""
	if len(workspaces) > 0 {
		workspaceID = workspaces[0].Id
	}
	access, refresh, err := c.auth.IssueTokens(u, done.ClientID, workspaceID)
	if err != nil {
		return response.Unauthorized(title, "the sign-in could not be completed")
	}
	if done.Method == mfa.MethodRecovery {
		mfa.Emit(context.Background(), mfa.Event{Action: "mfa.recovery.use", UserID: u.Id, WorkspaceID: workspaceID})
	}
	// A device session for everyone signed in, with or without a workspace: it is
	// what the browser re-mints its token from (GET sessions), and it hands out the
	// workspace's token once the person creates or joins one.
	deviceID, sessionID := c.bootstrapDevice(u.Id, done.ClientID, done.UserAgent, cookieValue(dto.Cookie, DeviceCookie))
	c.logDeviceLogin(deviceID, sessionID, dto.Ctx)
	return withDeviceCookie(response.Success(c.sessionPayload(u, done.ClientID, map[string]any{
		"user":         maskUser(u),
		"accessToken":  access,
		"refreshToken": refresh,
		"mfaMethod":    done.Method,
	}), title, c.lang.Translate("apps.identity.app.post.login.success", nil, pref), nil), deviceID)
}

// MfaKeyOptions is POST /mfa/verify/key: the options for the browser's
// navigator.credentials.get, for a sign-in or re-verification in progress.
func (c *AuthController) MfaKeyOptions(dto *MfaChallengeDto) types.Output {
	title := c.mfaTitle(dto.Ctx.Preference)
	if c.mfa == nil {
		return mfaError(title, mfa.ErrKeysUnavailable)
	}
	opts, err := c.mfa.KeyOptions(context.Background(), dto.Challenge)
	if err != nil {
		return mfaError(title, err)
	}
	return response.Success(map[string]any{"options": opts}, title, "use your security key", nil)
}

func (c *AuthController) mfaReady(ctx ntxctx.NTXContext) (types.Output, bool) {
	title := c.mfaTitle(ctx.Preference)
	if c.mfa == nil {
		return response.InternalServerError(title, "two-factor authentication is not available"), false
	}
	if ctx.UserID == "" {
		return response.Unauthorized(title, "sign in first"), false
	}
	return nil, true
}

// MfaStatus is GET /mfa: the caller's setup.
func (c *AuthController) MfaStatus(dto *MfaDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	st, err := c.mfa.Status(context.Background(), dto.Ctx.UserID)
	if err != nil {
		return mfaError(c.mfaTitle(dto.Ctx.Preference), err)
	}
	return response.Success(st, c.mfaTitle(dto.Ctx.Preference), "two-factor authentication", nil)
}

// MfaStepUp is POST /mfa/challenge: a re-verification for a sensitive change.
func (c *AuthController) MfaStepUp(dto *MfaDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	p, err := c.mfa.StartStepUp(context.Background(), dto.Ctx.UserID)
	if err != nil {
		return mfaError(c.mfaTitle(dto.Ctx.Preference), err)
	}
	return response.Success(p, c.mfaTitle(dto.Ctx.Preference), "confirm it's you", nil)
}

// reverify requires a re-verification once the caller has 2FA (adding or
// changing a factor, removing one, new recovery codes); before that there is
// nothing to re-verify with.
func (c *AuthController) reverify(dto *MfaProofDto) error {
	ctx := context.Background()
	on, err := c.mfa.Enrolled(ctx, dto.Ctx.UserID)
	if err != nil || !on {
		return err
	}
	return c.mfa.Prove(ctx, dto.Ctx.UserID, dto.Challenge, dto.proof())
}

func issuer() string {
	for _, k := range []string{"BUSINESS_NAME", "APP_NAME"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return "Account"
}

// MfaTotpStart is POST /mfa/totp: a new authenticator-app secret to scan.
func (c *AuthController) MfaTotpStart(dto *MfaProofDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	title := c.mfaTitle(dto.Ctx.Preference)
	if err := c.reverify(dto); err != nil {
		return mfaError(title, err)
	}
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title, "sign in first")
	}
	setup, err := c.mfa.StartTOTP(context.Background(), u.Id, issuer(), u.Email)
	if err != nil {
		return mfaError(title, err)
	}
	return response.Success(setup, title, "scan the code with your authenticator app", nil)
}

// MfaTotpConfirm is POST /mfa/totp/confirm: the first code turns the app on.
func (c *AuthController) MfaTotpConfirm(dto *MfaTotpConfirmDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	title := c.mfaTitle(dto.Ctx.Preference)
	codes, err := c.mfa.ConfirmTOTP(context.Background(), dto.Ctx.UserID, dto.FactorId, dto.Code)
	if err != nil {
		return mfaError(title, err)
	}
	c.emitAdded(dto.Ctx, codes != nil, mfa.KindTOTP, "Authenticator app")
	return c.mfaAnswer(dto.Ctx, codes, nil, "authenticator app enabled")
}

func (c *AuthController) mfaAnswer(ctx ntxctx.NTXContext, codes []string, extra map[string]any, msg string) types.Output {
	title := c.mfaTitle(ctx.Preference)
	st, err := c.mfa.Status(context.Background(), ctx.UserID)
	if err != nil {
		return mfaError(title, err)
	}
	out := map[string]any{"status": st}
	if codes != nil {
		out["recoveryCodes"] = codes
	}
	for k, v := range extra {
		out[k] = v
	}
	return response.Success(out, title, msg, nil)
}

// MfaKeyStart is POST /mfa/keys: options for navigator.credentials.create.
func (c *AuthController) MfaKeyStart(dto *MfaProofDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	title := c.mfaTitle(dto.Ctx.Preference)
	if err := c.reverify(dto); err != nil {
		return mfaError(title, err)
	}
	u, _ := c.userEntity.First(`"id" = ?`, dto.Ctx.UserID)
	if u == nil {
		return response.Unauthorized(title, "sign in first")
	}
	opts, token, err := c.mfa.BeginKey(context.Background(), u.Id, u.Email, u.Name)
	if err != nil {
		return mfaError(title, err)
	}
	return response.Success(map[string]any{"options": opts, "token": token}, title, "touch your security key", nil)
}

// MfaKeyConfirm is POST /mfa/keys/confirm: the browser's attestation stores the key.
func (c *AuthController) MfaKeyConfirm(dto *MfaKeyConfirmDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	key, codes, err := c.mfa.FinishKey(context.Background(), dto.Ctx.UserID, dto.Token, dto.Name, dto.Credential)
	if err != nil {
		return mfaError(c.mfaTitle(dto.Ctx.Preference), err)
	}
	c.emitAdded(dto.Ctx, codes != nil, mfa.KindWebAuthn, key.Name)
	return c.mfaAnswer(dto.Ctx, codes, map[string]any{"key": key}, "security key added")
}

// MfaRemove is POST /mfa/factors/:id/remove.
func (c *AuthController) MfaRemove(dto *MfaProofDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	title := c.mfaTitle(dto.Ctx.Preference)
	if err := c.mfa.Prove(context.Background(), dto.Ctx.UserID, dto.Challenge, dto.proof()); err != nil {
		return mfaError(title, err)
	}
	if err := c.mfa.RemoveFactor(context.Background(), dto.Ctx.UserID, dto.Id); err != nil {
		return mfaError(title, err)
	}
	mfa.Emit(context.Background(), mfa.Event{Action: "mfa.factor.remove", UserID: dto.Ctx.UserID, WorkspaceID: dto.Ctx.WorkspaceID, Detail: map[string]any{"factorId": dto.Id}})
	if on, _ := c.mfa.Enrolled(context.Background(), dto.Ctx.UserID); !on {
		mfa.Emit(context.Background(), mfa.Event{Action: "mfa.disable", UserID: dto.Ctx.UserID, WorkspaceID: dto.Ctx.WorkspaceID})
	}
	return c.mfaAnswer(dto.Ctx, nil, nil, "removed")
}

// MfaRecovery is POST /mfa/recovery: ten new recovery codes, shown once.
func (c *AuthController) MfaRecovery(dto *MfaProofDto) types.Output {
	if out, ok := c.mfaReady(dto.Ctx); !ok {
		return out
	}
	title := c.mfaTitle(dto.Ctx.Preference)
	if err := c.mfa.Prove(context.Background(), dto.Ctx.UserID, dto.Challenge, dto.proof()); err != nil {
		return mfaError(title, err)
	}
	codes, err := c.mfa.RegenerateRecovery(context.Background(), dto.Ctx.UserID)
	if err != nil {
		return mfaError(title, err)
	}
	mfa.Emit(context.Background(), mfa.Event{Action: "mfa.recovery.regenerate", UserID: dto.Ctx.UserID, WorkspaceID: dto.Ctx.WorkspaceID})
	return c.mfaAnswer(dto.Ctx, codes, nil, "new recovery codes")
}

// policyView is the workspace's requirement and who has not set up 2FA yet.
func (c *AuthController) policyView(ctx context.Context, ws, caller string) (map[string]any, error) {
	req, err := c.mfa.Required(ctx, ws)
	if err != nil {
		return nil, err
	}
	members, err := org.New(c.mfa.DB).Members(ctx, ws)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	on, err := c.mfa.EnrolledAmong(ctx, ids)
	if err != nil {
		return nil, err
	}
	missing := []map[string]any{}
	for _, m := range members {
		if !on[m.UserID] {
			missing = append(missing, map[string]any{"userId": m.UserID, "name": m.Name, "email": m.Email})
		}
	}
	return map[string]any{"require": req, "missing": missing, "members": len(members), "youEnrolled": on[caller]}, nil
}

func (c *AuthController) policyAdmin(ctx ntxctx.NTXContext) (types.Output, bool) {
	if out, ok := c.mfaReady(ctx); !ok {
		return out, false
	}
	title := c.mfaTitle(ctx.Preference)
	if ctx.WorkspaceID == "" {
		return response.Forbidden(title, "choose an organization first"), false
	}
	ok, err := org.New(c.mfa.DB).Allowed(context.Background(), ctx.WorkspaceID, ctx.UserID, org.ManageSecurity)
	if err != nil {
		return response.InternalServerError(title, "the policy could not be read"), false
	}
	if !ok {
		return response.Forbidden(title, "only organization owners and admins manage two-factor policy"), false
	}
	return nil, true
}

// MfaPolicy is GET /mfa/policy (owners and admins of the token's workspace).
func (c *AuthController) MfaPolicy(dto *MfaDto) types.Output {
	if out, ok := c.policyAdmin(dto.Ctx); !ok {
		return out
	}
	v, err := c.policyView(context.Background(), dto.Ctx.WorkspaceID, dto.Ctx.UserID)
	if err != nil {
		return mfaError(c.mfaTitle(dto.Ctx.Preference), err)
	}
	return response.Success(v, c.mfaTitle(dto.Ctx.Preference), "two-factor policy", nil)
}

// MfaSetPolicy is PUT /mfa/policy. Requiring 2FA needs the caller's own.
func (c *AuthController) MfaSetPolicy(dto *MfaPolicyDto) types.Output {
	if out, ok := c.policyAdmin(dto.Ctx); !ok {
		return out
	}
	title := c.mfaTitle(dto.Ctx.Preference)
	ctx := context.Background()
	if err := c.mfa.SetRequired(ctx, dto.Ctx.WorkspaceID, dto.Ctx.UserID, dto.Require); err != nil {
		return mfaError(title, err)
	}
	mfa.Emit(ctx, mfa.Event{Action: "mfa.policy.change", UserID: dto.Ctx.UserID, WorkspaceID: dto.Ctx.WorkspaceID, Detail: map[string]any{"require": dto.Require}})
	v, err := c.policyView(ctx, dto.Ctx.WorkspaceID, dto.Ctx.UserID)
	if err != nil {
		return mfaError(title, err)
	}
	return response.Success(v, title, "two-factor policy saved", nil)
}

// emitAdded reports a new factor, and 2FA turning on when it is the first.
func (c *AuthController) emitAdded(ctx ntxctx.NTXContext, first bool, kind, name string) {
	bg := context.Background()
	mfa.Emit(bg, mfa.Event{Action: "mfa.factor.add", UserID: ctx.UserID, WorkspaceID: ctx.WorkspaceID, Detail: map[string]any{"kind": kind, "name": name}})
	if first {
		mfa.Emit(bg, mfa.Event{Action: "mfa.enable", UserID: ctx.UserID, WorkspaceID: ctx.WorkspaceID})
	}
}
