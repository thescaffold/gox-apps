package provider

import (
	"os"
	"strings"

	"github.com/awesome-goose/goose/types"
)

type ProviderService struct {
	entity *ProviderEntity `inject:""`
}

// builtInProviderSeed describes one of the four OAuth providers Origine
// ships with (TRD §6.23 / PLAN M0-27a). siteURL/scope are public, fixed
// facts about that provider's own OAuth API (matching the literals each
// provider's own Authorize method in pkg/provider/*.go already hardcodes,
// even though that method itself is never called here — see Boot's doc
// comment) — never deployment-specific, unlike oauth_client_id/redirect_url.
type builtInProviderSeed struct {
	key, name, siteURL, scope, clientIDEnv string
}

var builtInProviderSeeds = []builtInProviderSeed{
	{"github", "GitHub", "https://github.com/login/oauth/authorize", "read:user user:email", "GITHUB_CLIENT_ID"},
	{"gitlab", "GitLab", gitlabAuthorizeURL(), "read_user email", "GITLAB_CLIENT_ID"},
	{"bitbucket", "Bitbucket", "https://bitbucket.org/site/oauth2/authorize", "account email", "BITBUCKET_CLIENT_ID"},
	{"google", "Google", "https://accounts.google.com/o/oauth2/v2/auth", "openid email profile", "GOOGLE_CLIENT_ID"},
}

func gitlabAuthorizeURL() string {
	host := os.Getenv("GITLAB_HOST")
	if host == "" {
		host = "https://gitlab.com"
	}
	return strings.TrimRight(host, "/") + "/oauth/authorize"
}

// Boot seeds (or refreshes) an IdentityProviders row for each built-in
// provider that has a client id configured — that DB row, not the env var
// itself, is what actually makes `GET apps/identity/provider` (and so
// guard's sign-in buttons) offer it; see .env.example's OAuth block. A
// provider with no client id is left alone entirely (no row created, an
// existing row untouched) rather than surfaced with a button that can never
// complete.
//
// redirect_url is derived from BASE_URL + the fixed guard route
// `guard/provider/:provider/login`. This is intentionally the same path
// regardless of login vs. register: the OAuth `state` value already encodes
// which flow initiated it (`type|nonce`, parsed back out by guard's
// `initiators.provider` on return), and the return leg never reads the
// URL's own `:type` segment — only the initiating leg does, from whatever
// page the user actually started on. A blank BASE_URL (no domain registered
// yet, TRD Q9) produces a relative redirect_url; that's a real provider
// misconfiguration until Q9 is answered, not a bug here — left blank rather
// than invented, per this project's established convention.
func (s *ProviderService) Boot(_ types.Kernel) error {
	baseURL := strings.TrimRight(os.Getenv("BASE_URL"), "/")
	for _, seed := range builtInProviderSeeds {
		clientID := os.Getenv(seed.clientIDEnv)
		if clientID == "" {
			continue
		}
		redirectURL := baseURL + "/guard/provider/" + seed.key + "/login"

		existing, _ := s.entity.First(`"key" = ?`, seed.key)
		if existing == nil {
			key, name, siteURL, scope := seed.key, seed.name, seed.siteURL, seed.scope
			if err := s.entity.Insert(&Provider{
				Key: &key, Name: &name, SiteUrl: &siteURL, Scope: &scope,
				RedirectUrl: &redirectURL, OauthClientId: &clientID,
			}); err != nil {
				return err
			}
			continue
		}

		existing.OauthClientId = &clientID
		existing.RedirectUrl = &redirectURL
		if existing.SiteUrl == nil {
			existing.SiteUrl = &seed.siteURL
		}
		if existing.Scope == nil {
			existing.Scope = &seed.scope
		}
		if _, err := s.entity.Update(existing, `"id" = ?`, existing.Id); err != nil {
			return err
		}
	}
	return nil
}
