// Package invitation brings people into a workspace (PRD PLT-03): an email
// invitation with a role, accepted once by that address, or the workspace's
// shareable link, accepted by anyone signed in. Both expire after a week and
// can be revoked. Only an owner or admin invites, never above their own role
// (org.CheckGrant). The host may refuse a join (SetJoinGuard: a plan's seat
// limit) inside the accepting transaction.
//
// An email invitation's token is kept only as a hash. The link's token is
// sealed (IDENTITY_TOKEN_KEY) so owners and admins can copy it again.
package invitation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/thescaffold/gox-apps/libs/identity/pkg/org"
	"github.com/thescaffold/gox-apps/libs/identity/pkg/tokenstore"
)

// Kinds.
const (
	KindEmail = "email"
	KindLink  = "link"
)

// TTL is how long an invitation stays open.
const TTL = 7 * 24 * time.Hour

// MaxPerRequest caps how many addresses one request invites.
const MaxPerRequest = 50

// Errors a caller can match.
var (
	ErrNotFound    = errors.New("this invitation is not valid")
	ErrExpired     = errors.New("this invitation has expired; ask for a new one")
	ErrUsed        = errors.New("this invitation has already been used")
	ErrWrongPerson = errors.New("this invitation is for another email address; sign in with that address")
	ErrInvalid     = errors.New("enter one or more email addresses")
	ErrTooMany     = errors.New("invite at most 50 people at a time")
)

// Invitation is one invitation.
type Invitation struct {
	ID          string     `gorm:"column:id;primaryKey" json:"id"`
	WorkspaceID string     `gorm:"column:workspace_id" json:"workspaceId"`
	ClientID    string     `gorm:"column:client_id" json:"-"`
	Kind        string     `gorm:"column:kind" json:"kind"`
	Email       string     `gorm:"column:email" json:"email,omitempty"`
	Seat        string     `gorm:"column:seat" json:"role"`
	TokenHash   string     `gorm:"column:token_hash" json:"-"`
	TokenSealed *string    `gorm:"column:token_sealed" json:"-"`
	InvitedBy   string     `gorm:"column:invited_by" json:"invitedBy"`
	Uses        int        `gorm:"column:uses" json:"uses"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"createdAt"`
	ExpiresAt   time.Time  `gorm:"column:expires_at" json:"expiresAt"`
	AcceptedAt  *time.Time `gorm:"column:accepted_at" json:"acceptedAt,omitempty"`
	AcceptedBy  *string    `gorm:"column:accepted_by" json:"-"`
	RevokedAt   *time.Time `gorm:"column:revoked_at" json:"-"`
}

func (Invitation) TableName() string { return "IdentityInvitations" }

// Sent is a new invitation and the token its link carries (shown once).
type Sent struct {
	Invitation Invitation
	Token      string
}

// Event is a change a host may record (an audit log, a notification).
type Event struct {
	Action      string // invitation.send, invitation.link, invitation.revoke, invitation.accept
	WorkspaceID string
	ActorID     string // who invited, revoked or accepted
	Invitation  Invitation
}

var (
	mu       sync.RWMutex
	observer func(context.Context, Event)
	guard    func(tx *gorm.DB, workspaceID, seat string) error
)

// SetObserver registers the host's recorder (nil turns it off).
func SetObserver(f func(context.Context, Event)) { mu.Lock(); observer = f; mu.Unlock() }

// SetJoinGuard registers the host's check before someone joins with seat (a
// seat limit). It runs inside the accepting transaction; its error refuses the join.
func SetJoinGuard(f func(tx *gorm.DB, workspaceID, seat string) error) {
	mu.Lock()
	guard = f
	mu.Unlock()
}

func emit(ctx context.Context, e Event) {
	mu.RLock()
	f := observer
	mu.RUnlock()
	if f != nil {
		f(ctx, e)
	}
}

func joinGuard() func(tx *gorm.DB, workspaceID, seat string) error {
	mu.RLock()
	defer mu.RUnlock()
	return guard
}

// URL is the address an invitation's token opens: PUBLIC_BASE_URL (else
// APP_URL, else BASE_URL) + /invite/ + token.
func URL(token string) string {
	base := ""
	for _, k := range []string{"PUBLIC_BASE_URL", "APP_URL", "BASE_URL"} {
		if v := strings.TrimRight(strings.TrimSpace(os.Getenv(k)), "/"); v != "" {
			base = v
			break
		}
	}
	return base + "/invite/" + token
}

// Service is the store and the rules. Now and Key are seams for tests.
type Service struct {
	DB  *gorm.DB
	Now func() time.Time
	Key func() string
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *Service) key() string {
	if s.Key != nil {
		return s.Key()
	}
	return tokenstore.Key()
}

func newToken() (string, string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	t := base64.RawURLEncoding.EncodeToString(b)
	return t, hashToken(t), nil
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte("identity-invitation:" + t))
	return hex.EncodeToString(sum[:])
}

func normalizeEmails(in []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, raw := range in {
		for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\n' }) {
			e := strings.ToLower(strings.TrimSpace(part))
			if e == "" {
				continue
			}
			a, err := mail.ParseAddress(e)
			if err != nil || a.Address != e || !strings.Contains(e[strings.LastIndex(e, "@"):], ".") {
				return nil, fmt.Errorf("%w: %q is not an email address", ErrInvalid, part)
			}
			if !seen[e] {
				seen[e] = true
				out = append(out, e)
			}
		}
	}
	if len(out) == 0 {
		return nil, ErrInvalid
	}
	if len(out) > MaxPerRequest {
		return nil, ErrTooMany
	}
	return out, nil
}

// Invite sends email invitations with seat, acting as actorID. Addresses
// already in the workspace are skipped (returned in members); an address
// with an open invitation gets a fresh one in its place.
func (s *Service) Invite(ctx context.Context, ws, clientID, actorID string, emails []string, seat string) (sent []Sent, members []string, err error) {
	seat = strings.ToLower(strings.TrimSpace(seat))
	if err := org.New(s.DB).CheckGrant(ctx, ws, actorID, seat); err != nil {
		return nil, nil, err
	}
	list, err := normalizeEmails(emails)
	if err != nil {
		return nil, nil, err
	}
	var in []string
	if err := s.DB.WithContext(ctx).Raw(`SELECT lower(u."email") FROM "IdentityUserClientWorkspaces" m
		JOIN "IdentityUsers" u ON u."id" = m."user_id"
		WHERE m."workspace_id" = ? AND m."deleted_at" IS NULL AND lower(u."email") IN ?`, ws, list).Scan(&in).Error; err != nil {
		return nil, nil, err
	}
	isMember := map[string]bool{}
	for _, e := range in {
		isMember[e] = true
	}
	now := s.now()
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, e := range list {
			if isMember[e] {
				members = append(members, e)
				continue
			}
			if err := tx.Model(&Invitation{}).Where(`"workspace_id" = ? AND "kind" = ? AND "email" = ? AND "accepted_at" IS NULL AND "revoked_at" IS NULL`, ws, KindEmail, e).
				Update("revoked_at", now).Error; err != nil {
				return err
			}
			tok, h, err := newToken()
			if err != nil {
				return err
			}
			inv := Invitation{ID: uuid.NewString(), WorkspaceID: ws, ClientID: clientID, Kind: KindEmail, Email: e, Seat: seat,
				TokenHash: h, InvitedBy: actorID, CreatedAt: now, ExpiresAt: now.Add(TTL)}
			if err := tx.Create(&inv).Error; err != nil {
				return err
			}
			sent = append(sent, Sent{Invitation: inv, Token: tok})
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	for _, x := range sent {
		emit(ctx, Event{Action: "invitation.send", WorkspaceID: ws, ActorID: actorID, Invitation: x.Invitation})
	}
	return sent, members, nil
}

// Link makes the workspace's shareable link for seat, replacing any earlier one.
func (s *Service) Link(ctx context.Context, ws, clientID, actorID, seat string) (Sent, error) {
	seat = strings.ToLower(strings.TrimSpace(seat))
	if seat == "" {
		seat = "member"
		for _, r := range org.SeatRoles() {
			if r == "developer" {
				seat = r
			}
		}
	}
	if err := org.New(s.DB).CheckGrant(ctx, ws, actorID, seat); err != nil {
		return Sent{}, err
	}
	tok, h, err := newToken()
	if err != nil {
		return Sent{}, err
	}
	sealed, err := tokenstore.SealProviderToken(tok, s.key())
	if err != nil {
		return Sent{}, err
	}
	now := s.now()
	inv := Invitation{ID: uuid.NewString(), WorkspaceID: ws, ClientID: clientID, Kind: KindLink, Seat: seat,
		TokenHash: h, TokenSealed: &sealed, InvitedBy: actorID, CreatedAt: now, ExpiresAt: now.Add(TTL)}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Invitation{}).Where(`"workspace_id" = ? AND "kind" = ? AND "revoked_at" IS NULL`, ws, KindLink).
			Update("revoked_at", now).Error; err != nil {
			return err
		}
		return tx.Create(&inv).Error
	})
	if err != nil {
		return Sent{}, err
	}
	emit(ctx, Event{Action: "invitation.link", WorkspaceID: ws, ActorID: actorID, Invitation: inv})
	return Sent{Invitation: inv, Token: tok}, nil
}

func (s *Service) admin(ctx context.Context, ws, actorID string) error {
	ok, err := org.New(s.DB).IsAdmin(ctx, ws, actorID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: only an owner or admin can see and change invitations", org.ErrNotAllowed)
	}
	return nil
}

// Open lists the workspace's open invitations (owners and admins), with the
// current link's token when there is one.
func (s *Service) Open(ctx context.Context, ws, actorID string) (emails []Invitation, link *Sent, err error) {
	if err := s.admin(ctx, ws, actorID); err != nil {
		return nil, nil, err
	}
	var all []Invitation
	if err := s.DB.WithContext(ctx).Where(`"workspace_id" = ? AND "revoked_at" IS NULL AND "accepted_at" IS NULL AND "expires_at" > ?`, ws, s.now()).
		Order(`"created_at" DESC`).Find(&all).Error; err != nil {
		return nil, nil, err
	}
	emails = []Invitation{}
	for _, inv := range all {
		if inv.Kind == KindEmail {
			emails = append(emails, inv)
			continue
		}
		if link == nil && inv.TokenSealed != nil {
			if tok, err := tokenstore.OpenProviderToken(*inv.TokenSealed, s.key()); err == nil {
				link = &Sent{Invitation: inv, Token: tok}
			}
		}
	}
	return emails, link, nil
}

// Revoke closes an invitation (owners and admins).
func (s *Service) Revoke(ctx context.Context, ws, actorID, id string) error {
	if err := s.admin(ctx, ws, actorID); err != nil {
		return err
	}
	var inv Invitation
	if err := s.DB.WithContext(ctx).Where(`"id" = ? AND "workspace_id" = ?`, id, ws).Take(&inv).Error; err != nil {
		return ErrNotFound
	}
	if err := s.DB.WithContext(ctx).Model(&Invitation{}).Where(`"id" = ? AND "revoked_at" IS NULL`, id).Update("revoked_at", s.now()).Error; err != nil {
		return err
	}
	emit(ctx, Event{Action: "invitation.revoke", WorkspaceID: ws, ActorID: actorID, Invitation: inv})
	return nil
}

// Preview is what the invitation screen shows before someone accepts.
type Preview struct {
	Kind          string    `json:"kind"`
	Email         string    `json:"email,omitempty"`
	Role          string    `json:"role"`
	WorkspaceID   string    `json:"workspaceId"`
	WorkspaceName string    `json:"workspaceName"`
	Handle        string    `json:"handle,omitempty"`
	InviterName   string    `json:"inviterName"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

func (s *Service) find(ctx context.Context, tx *gorm.DB, token string, lock bool) (Invitation, error) {
	var inv Invitation
	q := tx.WithContext(ctx).Where(`"token_hash" = ?`, hashToken(strings.TrimSpace(token)))
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Take(&inv).Error; err != nil {
		return Invitation{}, ErrNotFound
	}
	if inv.RevokedAt != nil {
		return Invitation{}, ErrNotFound
	}
	if inv.Kind == KindEmail && inv.AcceptedAt != nil {
		return Invitation{}, ErrUsed
	}
	if !s.now().Before(inv.ExpiresAt) {
		return Invitation{}, ErrExpired
	}
	return inv, nil
}

// Look returns the preview for token (no sign-in needed: the token is the
// secret).
func (s *Service) Look(ctx context.Context, token string) (Preview, error) {
	inv, err := s.find(ctx, s.DB, token, false)
	if err != nil {
		return Preview{}, err
	}
	var row struct {
		Name   string
		Handle *string
	}
	_ = s.DB.WithContext(ctx).Raw(`SELECT "name", "handle" FROM "IdentityWorkspaces" WHERE "id" = ?`, inv.WorkspaceID).Scan(&row).Error
	var inviter string
	_ = s.DB.WithContext(ctx).Raw(`SELECT "name" FROM "IdentityUsers" WHERE "id" = ?`, inv.InvitedBy).Scan(&inviter).Error
	p := Preview{Kind: inv.Kind, Email: inv.Email, Role: inv.Seat, WorkspaceID: inv.WorkspaceID, WorkspaceName: row.Name, InviterName: inviter, ExpiresAt: inv.ExpiresAt}
	if row.Handle != nil {
		p.Handle = *row.Handle
	}
	return p, nil
}

// Accept joins userID (whose address is email) to the invitation's workspace
// with its role and returns the workspace. Someone already in the workspace
// keeps their seat; the invitation is still used up.
func (s *Service) Accept(ctx context.Context, token, userID, email string) (string, error) {
	var ws string
	var accepted Invitation
	joined := false
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inv, err := s.find(ctx, tx, token, true)
		if err != nil {
			return err
		}
		if inv.Kind == KindEmail && !strings.EqualFold(strings.TrimSpace(email), inv.Email) {
			return ErrWrongPerson
		}
		ws, accepted = inv.WorkspaceID, inv
		var n int64
		if err := tx.Table(`"IdentityUserClientWorkspaces"`).Where(`"workspace_id" = ? AND "user_id" = ? AND "deleted_at" IS NULL`, ws, userID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if g := joinGuard(); g != nil {
				if err := g(tx, ws, inv.Seat); err != nil {
					return err
				}
			}
			now := s.now()
			if err := tx.Exec(`INSERT INTO "IdentityUserClientWorkspaces" ("id", "created_at", "updated_at", "user_id", "client_id", "workspace_id", "status")
				VALUES (?, ?, ?, ?, ?, ?, ?)`, uuid.NewString(), now, now, userID, inv.ClientID, ws, inv.Seat).Error; err != nil {
				return err
			}
			joined = true
		}
		upd := map[string]any{"uses": gorm.Expr(`"uses" + 1`)}
		if inv.Kind == KindEmail {
			upd["accepted_at"], upd["accepted_by"] = s.now(), userID
		}
		return tx.Model(&Invitation{}).Where(`"id" = ?`, inv.ID).Updates(upd).Error
	})
	if err != nil {
		return "", err
	}
	if joined {
		emit(ctx, Event{Action: "invitation.accept", WorkspaceID: ws, ActorID: userID, Invitation: accepted})
	}
	return ws, nil
}
