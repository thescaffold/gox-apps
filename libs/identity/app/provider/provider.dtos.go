package provider

import (
	"encoding/json"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

// ProvidersListDto carries GET /provider/providers query params.
type ProvidersListDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	ClientId string            `query:"clientId"`
	Type     string            `query:"type"`
}

// ProviderLookupDto carries GET /provider/:provider/provider params.
type ProviderLookupDto struct {
	Ctx      ntxctx.NTXContext `context:"ntx"`
	Provider string            `param:"provider"`
	ClientId string            `query:"clientId"`
	Type     string            `query:"type"`
}

// AuthorizeProviderDto carries POST /provider/:provider/authorize body.
// Mirrors TS AuthorizeProviderDto — `code` is the provider's auth code,
// `state` is the cache-validated nonce, `type` is register vs login.
type AuthorizeProviderDto struct {
	Ctx          ntxctx.NTXContext `context:"ntx"`
	Provider     string            `param:"provider"`
	Code         string            `json:"code"         binding:"required"`
	ClientId     string            `json:"clientId"     binding:"required"`
	State        string            `json:"state,omitempty"`
	UserAgent    string            `json:"userAgent,omitempty"`
	Type         string            `json:"type,omitempty"`
	ReferralCode string            `json:"referralCode,omitempty"`
}

type CreateProviderDto struct {
	UserId      string          `json:"userId"     binding:"required"`
	ClientId    *string         `json:"clientId,omitempty"`
	Type        string          `json:"type"       binding:"required"`
	Reference   string          `json:"reference"  binding:"required"`
	Key         *string         `json:"key,omitempty"`
	Name        *string         `json:"name,omitempty"`
	Desc        *string         `json:"desc,omitempty"`
	Detail      *string         `json:"detail,omitempty"`
	LogoUrl     *string         `json:"logoUrl,omitempty"`
	SiteUrl     *string         `json:"siteUrl,omitempty"`
	RedirectUrl *string         `json:"redirectUrl,omitempty"`
	Scope       *string         `json:"scope,omitempty"`
	Email       *string         `json:"email,omitempty"`
	Token       *string         `json:"token,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}

type UpdateProviderDto struct {
	Key         *string         `json:"key,omitempty"`
	Name        *string         `json:"name,omitempty"`
	Desc        *string         `json:"desc,omitempty"`
	Detail      *string         `json:"detail,omitempty"`
	LogoUrl     *string         `json:"logoUrl,omitempty"`
	SiteUrl     *string         `json:"siteUrl,omitempty"`
	RedirectUrl *string         `json:"redirectUrl,omitempty"`
	Scope       *string         `json:"scope,omitempty"`
	Email       *string         `json:"email,omitempty"`
	Token       *string         `json:"token,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Status      *string         `json:"status,omitempty"`
}
