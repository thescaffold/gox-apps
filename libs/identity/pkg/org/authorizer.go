package org

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// The org-level actions identity checks. A host that has its own roles (custom roles that may hold
// some of these) answers them through SetAuthorizer; without one, each is "an owner or admin of the
// workspace", exactly as before the hook.
const (
	// ManageMembers: seats, invitations, teams and who is in them.
	ManageMembers = "org.members.manage"
	// ManageSecurity: the single sign-on connection and the two-factor policy.
	ManageSecurity = "org.security.manage"
	// ManageSettings: the workspace's name and handle.
	ManageSettings = "org.settings.manage"
	// ManageTokens: other people's API keys (list and revoke; a key is only ever rotated by its holder).
	ManageTokens = "org.tokens.manage"
	// ManageRoles: identity's custom roles and who holds them.
	ManageRoles = "org.roles.manage"
	// ReadAudit: the organisation's activity.
	ReadAudit = "org.audit.read"
)

// Authorizer decides whether userID may take action in workspace ws. It replaces identity's own
// owner-or-admin check for that action; the seat rules that follow it (nobody grants or touches a
// seat above their own, only an owner grants or removes owner, one owner stays) still apply.
type Authorizer func(ctx context.Context, ws, userID, action string) (bool, error)

var authorizer struct {
	mu sync.RWMutex
	fn Authorizer
}

// SetAuthorizer installs the host's authorizer; nil restores identity's own checks.
func SetAuthorizer(fn Authorizer) {
	authorizer.mu.Lock()
	authorizer.fn = fn
	authorizer.mu.Unlock()
}

// Authorize answers action for userID in ws: the host's authorizer when one is set, otherwise
// builtin, the check the call site always made.
func Authorize(ctx context.Context, ws, userID, action string, builtin func() (bool, error)) (bool, error) {
	if ws == "" || userID == "" {
		return false, nil
	}
	authorizer.mu.RLock()
	fn := authorizer.fn
	authorizer.mu.RUnlock()
	if fn != nil {
		return fn(ctx, ws, userID, action)
	}
	return builtin()
}

// Allowed answers action for userID in ws, with "an owner or admin" (IsAdmin) as the default.
func (s *Service) Allowed(ctx context.Context, ws, userID, action string) (bool, error) {
	return Authorize(ctx, ws, userID, action, func() (bool, error) { return s.IsAdmin(ctx, ws, userID) })
}

// SeatRemoved is told, inside the transaction that removes userID's seat in ws, so a host can end
// what the seat gave them (project memberships, overrides) atomically with it. actorID is who
// removed them. An error rolls the removal back.
type SeatRemoved func(ctx context.Context, tx *gorm.DB, ws, userID, actorID string) error

var seatRemoved struct {
	mu sync.RWMutex
	fn SeatRemoved
}

// OnSeatRemoved installs the host's clean-up for a removed seat; nil removes it.
func OnSeatRemoved(fn SeatRemoved) {
	seatRemoved.mu.Lock()
	seatRemoved.fn = fn
	seatRemoved.mu.Unlock()
}

// SeatRemovedHook is the installed clean-up (nil when none), for hosts that remove seats outside
// RemoveSeat (a SCIM deprovisioning, say) and must run the same clean-up.
func SeatRemovedHook() SeatRemoved {
	seatRemoved.mu.RLock()
	defer seatRemoved.mu.RUnlock()
	return seatRemoved.fn
}
