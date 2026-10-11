package sso

import "context"

var store *Store

// SetStore registers the connection store used by Enforced. Identity's auth
// controller calls it at boot.
func SetStore(s *Store) { store = s }

// Enforced reports whether the email's company requires single sign-on. It is
// false when SSO is not configured at all, so callers can check it
// unconditionally.
func Enforced(ctx context.Context, email string) bool {
	if store == nil || email == "" {
		return false
	}
	return store.Enforced(ctx, email)
}
