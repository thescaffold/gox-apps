package mfa

import (
	"context"
	"sync"
)

// Event is a security-relevant change a host may record (an audit log).
// WorkspaceID is the organization the person was acting in, when known.
type Event struct {
	Action      string // mfa.enable, mfa.disable, mfa.factor.add, mfa.factor.remove, mfa.recovery.regenerate, mfa.recovery.use, mfa.policy.change
	UserID      string // whose two-factor setup, or who changed the policy
	WorkspaceID string
	Detail      map[string]any // e.g. kind, name, require
}

var (
	mu       sync.RWMutex
	observer func(context.Context, Event)
	def      *Service
)

// SetObserver registers the host's recorder for Events (nil turns it off).
func SetObserver(f func(context.Context, Event)) {
	mu.Lock()
	defer mu.Unlock()
	observer = f
}

// Emit hands e to the observer, if any.
func Emit(ctx context.Context, e Event) {
	mu.RLock()
	f := observer
	mu.RUnlock()
	if f != nil {
		f(ctx, e)
	}
}

// SetDefault registers the Service identity built at boot, for other packages
// (the provider sign-in) and hosts (Gate, EnrolledAmong).
func SetDefault(s *Service) {
	mu.Lock()
	defer mu.Unlock()
	def = s
}

// Default returns the Service registered with SetDefault, or nil.
func Default() *Service {
	mu.RLock()
	defer mu.RUnlock()
	return def
}
