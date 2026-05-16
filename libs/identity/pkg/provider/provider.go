// Package provider hosts the OAuth provider integrations for identity.
// Mirrors ntx-apps/libs/identity/src/api/provider/providers/{github,gitlab,
// bitbucket,google}.service.ts.
//
// Each Provider exposes three operations:
//   - Authorize(state) — returns the URL the user should be redirected to
//   - Exchange(code)   — swaps the auth code for an access token
//   - Profile(token)   — fetches the user profile from the provider
//
// Providers self-register with ProviderRegistry on construction so the OAuth
// controller can dispatch by name.
package provider

import (
	"net/url"
	"sync"
)

// Token carries the access/refresh pair returned by Exchange. Fields are
// expressed as plain strings — providers normalize their own response shape.
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Scope        string
	TokenType    string
}

// Profile is the normalized user profile returned by Profile.
type Profile struct {
	ID         string
	Email      string
	Name       string
	AvatarURL  string
	Username   string
	Provider   string
	RawPayload map[string]any
}

// Provider is the contract every OAuth provider satisfies.
type Provider interface {
	Name() string
	Authorize(state, redirectURI string) string
	Exchange(code, redirectURI string) (*Token, error)
	Profile(token string) (*Profile, error)
}

// Registry holds the set of registered providers keyed by name.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry { return &Registry{providers: map[string]Provider{}} }

// Register installs (or replaces) a provider. Safe to call from module init.
func (r *Registry) Register(p Provider) {
	if p == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.providers == nil {
		r.providers = map[string]Provider{}
	}
	r.providers[p.Name()] = p
}

// Get returns the provider for name, or nil when missing.
func (r *Registry) Get(name string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// Names returns the slice of registered provider names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.providers))
	for k := range r.providers {
		out = append(out, k)
	}
	return out
}

// buildAuthorizeURL constructs a standard OAuth 2.0 authorization URL.
// Used by the per-provider Authorize implementations.
func buildAuthorizeURL(base, clientID, redirectURI, scope, state string, extra url.Values) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	if scope != "" {
		q.Set("scope", scope)
	}
	if state != "" {
		q.Set("state", state)
	}
	for k, vs := range extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return base + "?" + q.Encode()
}
