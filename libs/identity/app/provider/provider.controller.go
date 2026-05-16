package provider

import (
	"sort"
	"strings"
	"time"

	"github.com/awesome-goose/goose/types"
	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
	"github.com/thescaffold/gox-packages/libs/core/crud"
	"github.com/thescaffold/gox-packages/libs/core/i18n"
	"github.com/thescaffold/gox-packages/libs/core/response"
	"github.com/thescaffold/gox-packages/libs/core/services"
	"github.com/thescaffold/gox-packages/libs/core/utils"
)

// ProviderController mirrors ntx-apps/libs/identity/src/api/provider/provider.controller.ts.
// Phase H wired the TS-aligned columns (name/key/desc); morphs.afterList
// stable-sorts the page by github→bitbucket→gitlab→google preference.
//
// Phase Q adds the OAuth listing endpoints — `/providers` (list with masked
// secrets + state-cached fullSiteUrl) and `/:provider/provider` (single
// lookup). The actual OAuth flow (authorize/callback/login) is driven by
// identity/pkg/oauth.controller.go through the per-provider HTTP integrations
// under identity/pkg/provider/.
type ProviderController struct {
	crud.CrudResource[Provider, CreateProviderDto, UpdateProviderDto]

	entity *ProviderEntity       `inject:""`
	cache  services.CacheBackend `inject:""`
	lang   *i18n.Service         `inject:""`
}

// providerOrder is the explicit display order for the listing endpoints.
// Mirrors TS hard-coded list — newer providers slot at the end via the
// default-Infinity comparator.
var providerOrder = map[string]int{"github": 0, "bitbucket": 1, "gitlab": 2, "google": 3}

func (c *ProviderController) OnRegister() {
	c.Hydrate(c.entity, crud.Config[Provider, CreateProviderDto, UpdateProviderDto]{
		Name:       "provider",
		Searchable: []string{"name", "desc"},
		Unique: func(d *CreateProviderDto) []utils.KeyValue {
			out := []utils.KeyValue{}
			if d.Name != nil {
				out = append(out, utils.KeyValue{"name": *d.Name})
			}
			if d.Key != nil {
				out = append(out, utils.KeyValue{"key": *d.Key})
			}
			out = append(out, utils.KeyValue{"type": d.Type, "reference": d.Reference})
			return out
		},
		Morphs: map[string]crud.MorphFn{
			crud.AfterList: func(payload any, ctx ntxctx.NTXContext) (any, error) {
				rows, ok := payload.([]Provider)
				if !ok {
					return payload, nil
				}
				sort.SliceStable(rows, func(i, j int) bool {
					return providerOrderIndex(rows[i].Name) < providerOrderIndex(rows[j].Name)
				})
				return rows, nil
			},
		},
	})
	c.SetLang(c.lang)
}

// Providers mirrors TS @Get('providers') — lists every registered provider
// ordered by github→bitbucket→gitlab→google, masks the client secret, mints
// a fresh state token for each, caches the (state → clientId) binding for
// 10 minutes, and surfaces `fullSiteUrl` (the OAuth authorize URL the caller
// can redirect to).
func (c *ProviderController) Providers(dto *ProvidersListDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.provider.title", nil, pref)
	rows, _ := c.entity.Find(0, 0, ``)
	sort.SliceStable(rows, func(i, j int) bool {
		return providerOrderIndex(rows[i].Name) < providerOrderIndex(rows[j].Name)
	})
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		p := &rows[i]
		state := utils.UUID()
		c.cacheStateForClient(state, dto.ClientId)
		out = append(out, providerEnvelope(p, dto.Type, state))
	}
	return response.Success(out, title,
		c.lang.Translate("apps.identity.provider.get.providers.success", nil, pref), nil)
}

// Provider mirrors TS @Get(':provider/provider') — single lookup with the
// same state-caching contract as Providers.
func (c *ProviderController) Provider(dto *ProviderLookupDto) types.Output {
	pref := dto.Ctx.Preference
	title := c.lang.Translate("apps.identity.provider.title", nil, pref)
	row, _ := c.entity.First(`"key" = ?`, dto.Provider)
	if row == nil {
		return response.BadRequest(title,
			c.lang.Translate("apps.identity.provider.get.provider.error.invalid-request", nil, pref))
	}
	state := utils.UUID()
	c.cacheStateForClient(state, dto.ClientId)
	return response.Success(providerEnvelope(row, dto.Type, state), title,
		c.lang.Translate("apps.identity.provider.get.provider.success", nil, pref), nil)
}

// cacheStateForClient writes the (state → clientId) binding used by the
// /authorize and /login callbacks to recover the originating client. 10-min
// TTL matches TS cacheService.set(..., 1000 * 60 * 10).
func (c *ProviderController) cacheStateForClient(state, clientId string) {
	if c.cache == nil || state == "" {
		return
	}
	c.cache.Set("provider:oauth:initiate:"+state, clientId, 10*time.Minute)
}

// providerEnvelope builds the response shape — masked client secret,
// fullSiteUrl built from siteUrl + the provider-specific query string.
func providerEnvelope(p *Provider, authType, state string) map[string]any {
	out := map[string]any{
		"id":     p.Id,
		"key":    p.Key,
		"name":   p.Name,
		"desc":   p.Desc,
		"detail": p.Detail,
		"status": p.Status,
	}
	if p.LogoUrl != nil {
		out["logoUrl"] = *p.LogoUrl
	}
	if p.SiteUrl != nil {
		out["siteUrl"] = *p.SiteUrl
	}
	if p.RedirectUrl != nil {
		out["redirectUrl"] = *p.RedirectUrl
	}
	if p.Scope != nil {
		out["scope"] = *p.Scope
	}
	if p.OauthClientId != nil {
		out["clientId"] = *p.OauthClientId
	}
	if p.OauthClientSecret != nil {
		out["clientSecret"] = utils.Mask(*p.OauthClientSecret)
	}
	q := buildOAuthQuery(p, authType, state)
	if p.SiteUrl != nil && q != "" {
		out["fullSiteUrl"] = *p.SiteUrl + "?" + q
	}
	return out
}

// buildOAuthQuery encodes the per-provider authorize query string.
// Falls back to a generic OAuth 2.0 shape (client_id + redirect_uri + state +
// response_type=code + scope) so unknown providers still produce a workable
// URL. Per-provider overrides live in identity/pkg/provider/<name>.go.
func buildOAuthQuery(p *Provider, authType, state string) string {
	parts := []string{}
	if p.OauthClientId != nil && *p.OauthClientId != "" {
		parts = append(parts, "client_id="+*p.OauthClientId)
	}
	if p.RedirectUrl != nil && *p.RedirectUrl != "" {
		parts = append(parts, "redirect_uri="+*p.RedirectUrl)
	}
	parts = append(parts, "response_type=code")
	if p.Scope != nil && *p.Scope != "" {
		parts = append(parts, "scope="+*p.Scope)
	}
	if state != "" {
		parts = append(parts, "state="+state)
	}
	if authType != "" {
		parts = append(parts, "auth_type="+authType)
	}
	return strings.Join(parts, "&")
}

func providerOrderIndex(name *string) int {
	if name == nil {
		return 1 << 30
	}
	if v, ok := providerOrder[*name]; ok {
		return v
	}
	return 1 << 30
}
