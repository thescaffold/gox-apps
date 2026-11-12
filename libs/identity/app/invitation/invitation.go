// Package invitation serves invitations into a workspace (pkg/invitation)
// under apps/identity:
//
//	GET    /org/invitations             open invitations and the current link (owners, admins)
//	POST   /org/invitations             invite {emails, role}; each address gets an email
//	DELETE /org/invitations/:id         revoke
//	POST   /org/invitations/link        a new shareable link {role}
//	GET    /invitations/:token          what an invitation offers (no sign-in: the token is the secret)
//	POST   /invitations/:token/accept   join, signed in
package invitation

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/awesome-goose/goose/modules/router"
	"github.com/awesome-goose/goose/modules/sql"
	"github.com/awesome-goose/goose/types"
	"github.com/thescaffold/gox-apps/libs/identity/app/auth"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/invitation"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/mfa"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/events"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

var ROUTES = router.ForRoutes(
	router.Get("/org/invitations", []any{InvitationController{}, "Open"}, auth.RequireUser()...),
	router.Post("/org/invitations", []any{InvitationController{}, "Invite"}, auth.RequireUser()...),
	router.Delete("/org/invitations/:id", []any{InvitationController{}, "Revoke"}, auth.RequireUser()...),
	router.Post("/org/invitations/link", []any{InvitationController{}, "Link"}, auth.RequireUser()...),
	router.Get("/invitations/:token", []any{InvitationController{}, "Look"}),
	router.Post("/invitations/:token/accept", []any{InvitationController{}, "Accept"}, auth.RequireUser()...),
)

// InvitationModule mounts the routes.
type InvitationModule struct{}

func (m *InvitationModule) Imports() []types.Module {
	return []types.Module{ROUTES, sql.Child(&sql.Config{})}
}
func (m *InvitationModule) Exports() []any { return nil }
func (m *InvitationModule) Declarations() []any {
	return []any{&InvitationController{}}
}

// Mailer sends an invitation's email (tests replace it).
type Mailer func(inv invitation.Invitation, token, workspaceName, inviterName string)

// InvitationController answers the routes.
type InvitationController struct {
	db      *sql.Db                `inject:""`
	tracker *events.TrackerService `inject:""`

	once sync.Once
	svc  *invitation.Service
	mail Mailer
}

// NewControllerForTest builds a controller over svc that mails through m.
func NewControllerForTest(svc *invitation.Service, m Mailer) *InvitationController {
	c := &InvitationController{svc: svc, mail: m}
	c.once.Do(func() {})
	return c
}

func (c *InvitationController) service() *invitation.Service {
	c.once.Do(func() {
		c.svc = &invitation.Service{DB: c.db.DB}
		c.mail = c.trackerMail
	})
	return c.svc
}

const title = "invitations"

// trackerMail sends through the notification app (template key
// "organization-invitation"; the host provides the template).
func (c *InvitationController) trackerMail(inv invitation.Invitation, token, workspaceName, inviterName string) {
	if c.tracker == nil {
		return
	}
	brand := os.Getenv("BUSINESS_NAME")
	if brand == "" {
		brand = os.Getenv("APP_NAME")
	}
	role := inv.Seat
	if role != "" {
		role = strings.ToUpper(role[:1]) + role[1:]
	}
	c.tracker.Message("apps.notification.message.new", map[string]any{
		"reference": utils.Reference("INV", 32),
		"userId":    inv.ID,
		"clientId":  inv.ClientID,
		"key":       "organization-invitation",
		"channels":  []string{"email"},
		"data": map[string]any{
			"owner": inv.Email, "workspace": workspaceName, "inviter": inviterName, "role": role,
			"url": invitation.URL(token), "expires": inv.ExpiresAt.Format("Jan 2"), "brand": brand,
		},
		"subject":  "Join " + workspaceName + " on " + brand,
		"type":     "security",
		"priority": "medium",
		"scope":    "user",
	})
}

func fail(err error) types.Output {
	var hs interface {
		HTTPStatus() int
		ErrorCode() string
	}
	switch {
	case errors.As(err, &hs):
		return response.ErrorCode(title, err.Error(), hs.HTTPStatus(), hs.ErrorCode())
	case errors.Is(err, invitation.ErrNotFound):
		return response.ErrorCode(title, err.Error(), 404, "invitation_invalid")
	case errors.Is(err, invitation.ErrExpired):
		return response.ErrorCode(title, err.Error(), 410, "invitation_expired")
	case errors.Is(err, invitation.ErrUsed):
		return response.ErrorCode(title, err.Error(), 410, "invitation_used")
	case errors.Is(err, invitation.ErrWrongPerson):
		return response.ErrorCode(title, err.Error(), 403, "invitation_other_person")
	case errors.Is(err, invitation.ErrInvalid), errors.Is(err, invitation.ErrTooMany), errors.Is(err, org.ErrInvalid):
		return response.ErrorCode(title, strings.TrimPrefix(err.Error(), "invalid: "), 400, "invalid")
	case errors.Is(err, org.ErrNotAllowed):
		return response.Forbidden(title, strings.TrimPrefix(err.Error(), "not allowed: "))
	}
	return response.Internal(title, err)
}

// OpenReq, InviteReq, IDReq, LinkReq and TokenReq are the requests.
type OpenReq struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
}

type InviteReq struct {
	Ctx    ntxctx.NTXContext `context:"ntx"`
	Emails []string          `json:"emails" binding:"required"`
	Role   string            `json:"role"   binding:"required"`
}

type IDReq struct {
	Ctx ntxctx.NTXContext `context:"ntx"`
	Id  string            `param:"id"`
}

type LinkReq struct {
	Ctx  ntxctx.NTXContext `context:"ntx"`
	Role string            `json:"role"`
}

type TokenReq struct {
	Ctx   ntxctx.NTXContext `context:"ntx"`
	Token string            `param:"token"`
}

func signedIn(ctx ntxctx.NTXContext) types.Output {
	if ctx.UserID == "" || ctx.WorkspaceID == "" {
		return response.Unauthorized(title, "sign in first")
	}
	return nil
}

func linkView(l *invitation.Sent) map[string]any {
	if l == nil {
		return nil
	}
	return map[string]any{"id": l.Invitation.ID, "role": l.Invitation.Seat, "url": invitation.URL(l.Token), "expiresAt": l.Invitation.ExpiresAt}
}

func (c *InvitationController) Open(dto *OpenReq) types.Output {
	if out := signedIn(dto.Ctx); out != nil {
		return out
	}
	emails, link, err := c.service().Open(context.Background(), dto.Ctx.WorkspaceID, dto.Ctx.UserID)
	if err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"invitations": emails, "link": linkView(link)}, title, "invitations", nil)
}

func (c *InvitationController) names(ws, user string) (string, string) {
	var wsName, inviter string
	db := c.service().DB
	_ = db.Raw(`SELECT "name" FROM "IdentityWorkspaces" WHERE "id" = ?`, ws).Scan(&wsName).Error
	_ = db.Raw(`SELECT "name" FROM "IdentityUsers" WHERE "id" = ?`, user).Scan(&inviter).Error
	return wsName, inviter
}

func (c *InvitationController) Invite(dto *InviteReq) types.Output {
	if out := signedIn(dto.Ctx); out != nil {
		return out
	}
	sent, members, err := c.service().Invite(context.Background(), dto.Ctx.WorkspaceID, dto.Ctx.ClientID, dto.Ctx.UserID, dto.Emails, dto.Role)
	if err != nil {
		return fail(err)
	}
	wsName, inviter := c.names(dto.Ctx.WorkspaceID, dto.Ctx.UserID)
	out := make([]invitation.Invitation, 0, len(sent))
	for _, s := range sent {
		if c.mail != nil {
			c.mail(s.Invitation, s.Token, wsName, inviter)
		}
		out = append(out, s.Invitation)
	}
	if members == nil {
		members = []string{}
	}
	return response.Created(map[string]any{"invited": out, "alreadyMembers": members}, title, "invitations sent")
}

func (c *InvitationController) Revoke(dto *IDReq) types.Output {
	if out := signedIn(dto.Ctx); out != nil {
		return out
	}
	if err := c.service().Revoke(context.Background(), dto.Ctx.WorkspaceID, dto.Ctx.UserID, dto.Id); err != nil {
		return fail(err)
	}
	return response.Success(map[string]any{"revoked": true}, title, "revoked", nil)
}

func (c *InvitationController) Link(dto *LinkReq) types.Output {
	if out := signedIn(dto.Ctx); out != nil {
		return out
	}
	l, err := c.service().Link(context.Background(), dto.Ctx.WorkspaceID, dto.Ctx.ClientID, dto.Ctx.UserID, dto.Role)
	if err != nil {
		return fail(err)
	}
	return response.Created(linkView(&l), title, "link ready")
}

func (c *InvitationController) Look(dto *TokenReq) types.Output {
	p, err := c.service().Look(context.Background(), dto.Token)
	if err != nil {
		return fail(err)
	}
	out := map[string]any{"invitation": p}
	if m := mfa.Default(); m != nil {
		req, _ := m.Required(context.Background(), p.WorkspaceID)
		out["mfaRequired"] = req
	}
	return response.Success(out, title, "invitation", nil)
}

func (c *InvitationController) Accept(dto *TokenReq) types.Output {
	if dto.Ctx.UserID == "" {
		return response.Unauthorized(title, "sign in first")
	}
	var email string
	_ = c.service().DB.Raw(`SELECT "email" FROM "IdentityUsers" WHERE "id" = ?`, dto.Ctx.UserID).Scan(&email).Error
	// The membership is the accepting client's: the organization appears where they signed in.
	ws, err := c.service().AcceptAs(context.Background(), dto.Token, dto.Ctx.UserID, email, dto.Ctx.ClientID)
	if err != nil {
		return fail(err)
	}
	wsName, _ := c.names(ws, "")
	return response.Success(map[string]any{"workspaceId": ws, "workspaceName": wsName}, title, "joined", nil)
}
