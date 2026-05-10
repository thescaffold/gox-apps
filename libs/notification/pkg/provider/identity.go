package provider

// UserContext is the enriched per-user context the template renderer expects.
// Mirrors the ContextDto returned by ntx-apps-identity AppService.contextByUserId():
//
//	{ user, address, preference, organization }
type UserContext struct {
	User         map[string]any `json:"user,omitempty"`
	Address      map[string]any `json:"address,omitempty"`
	Preference   map[string]any `json:"preference,omitempty"`
	Organization map[string]any `json:"organization,omitempty"`
}

// IdentityFetcher resolves a UserContext by user id (or by the "owner" address
// the notification carries). Implementations typically delegate to identity's
// HTTP client. Pass a nil fetcher to disable enrichment — the renderer will
// then use only the bare event `data`, matching pre-enrichment behaviour.
type IdentityFetcher interface {
	// FetchByUserId returns the enriched context for a known user id.
	FetchByUserId(userId string) (*UserContext, error)
	// FetchByOwner is a fallback: when the message has no userId, look up by
	// the "owner" string (typically email or phone). Returning nil + nil is
	// acceptable when no match is found.
	FetchByOwner(owner string) (*UserContext, error)
}
