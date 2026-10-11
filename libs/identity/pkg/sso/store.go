package sso

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Directory is what identity needs to know about a company that it cannot
// know itself. The host (which owns the company's team data) implements it
// and registers it with SetDirectory. With no directory registered no domain
// can be claimed, so SSO cannot be enabled by accident.
type Directory interface {
	// OwnsDomain reports whether the workspace has verified control of the
	// email domain. Without this check any workspace could claim someone
	// else's domain and lock their people out.
	OwnsDomain(ctx context.Context, workspaceID, domain string) (bool, error)
	// Invited reports whether the email has been invited to (or is already a
	// member of) the workspace. SSO never admits anyone who is not.
	Invited(ctx context.Context, workspaceID, email string) (bool, error)
	// Joined is called after a successful SSO sign-in, to complete the
	// membership (accept the invitation, grant the mailbox). Idempotent.
	Joined(ctx context.Context, workspaceID, userID, email string) error
}

var directory Directory

// SetDirectory registers the host's Directory.
func SetDirectory(d Directory) { directory = d }

// CurrentDirectory returns the registered Directory, or nil.
func CurrentDirectory() Directory { return directory }

// Connection is one workspace's identity-provider registration.
type Connection struct {
	ID           string          `gorm:"column:id;primaryKey" json:"id"`
	WorkspaceID  string          `gorm:"column:workspace_id" json:"workspaceId"`
	Issuer       string          `gorm:"column:issuer" json:"issuer"`
	OIDCClientID string          `gorm:"column:oidc_client_id" json:"clientId"`
	SecretSealed string          `gorm:"column:client_secret_sealed" json:"-"`
	DomainsJSON  json.RawMessage `gorm:"column:domains;type:jsonb" json:"-"`
	Enforce      bool            `gorm:"column:enforce" json:"enforce"`
	CreatedAt    time.Time       `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time       `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName is the connections table.
func (Connection) TableName() string { return "IdentitySsoConnections" }

// Domains decodes the claimed email domains.
func (c Connection) Domains() []string {
	var d []string
	_ = json.Unmarshal(c.DomainsJSON, &d)
	return d
}

// Store persists connections.
type Store struct {
	DB *gorm.DB
	// LookupTXT and Now are seams for tests; nil means the real resolver and clock.
	LookupTXT func(ctx context.Context, name string) ([]string, error)
	Now       func() time.Time
}

// ErrNotFound is returned when no connection matches.
var ErrNotFound = errors.New("sso: no connection")

// ForWorkspace returns a workspace's connection.
func (s *Store) ForWorkspace(ctx context.Context, workspaceID string) (*Connection, error) {
	var c Connection
	err := s.DB.WithContext(ctx).Where("workspace_id = ?", workspaceID).Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

// ByID returns a connection by id.
func (s *Store) ByID(ctx context.Context, id string) (*Connection, error) {
	var c Connection
	err := s.DB.WithContext(ctx).Where("id = ?", id).Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

// ForEmail finds the connection that owns an email's domain.
func (s *Store) ForEmail(ctx context.Context, email string) (*Connection, error) {
	dom := EmailDomain(email)
	if dom == "" {
		return nil, ErrNotFound
	}
	var c Connection
	err := s.DB.WithContext(ctx).Where("domains @> ?::jsonb", `["`+dom+`"]`).Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

// Enforced reports whether the email's company requires SSO (password and
// code sign-in are refused for it).
func (s *Store) Enforced(ctx context.Context, email string) bool {
	c, err := s.ForEmail(ctx, email)
	if err != nil || !c.Enforce {
		return false
	}
	// A DNS-proven recovery lifts enforcement for that address for a while.
	return !s.recoveryGranted(ctx, email)
}

// Save inserts or replaces a workspace's connection.
func (s *Store) Save(ctx context.Context, c *Connection) error {
	var existing Connection
	err := s.DB.WithContext(ctx).Where("workspace_id = ?", c.WorkspaceID).Take(&existing).Error
	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.ID, c.CreatedAt, c.UpdatedAt = uuid.NewString(), now, now
		return s.DB.WithContext(ctx).Create(c).Error
	}
	if err != nil {
		return err
	}
	c.ID, c.CreatedAt, c.UpdatedAt = existing.ID, existing.CreatedAt, now
	return s.DB.WithContext(ctx).Save(c).Error
}

// Delete removes a workspace's connection (and with it any enforcement).
func (s *Store) Delete(ctx context.Context, workspaceID string) error {
	return s.DB.WithContext(ctx).Where("workspace_id = ?", workspaceID).Delete(&Connection{}).Error
}

// EmailDomain returns the lower-cased domain of an address, or "".
func EmailDomain(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	i := strings.LastIndex(email, "@")
	if i < 1 || i == len(email)-1 {
		return ""
	}
	return email[i+1:]
}
