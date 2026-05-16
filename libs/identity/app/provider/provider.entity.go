package provider

import (
	"encoding/json"
	"github.com/awesome-goose/goose/modules/sql"
)

// Provider mirrors ntx-apps/libs/identity/src/api/provider/entities/provider.entity.ts.
// TS treats `Provider` as a global OAuth-provider registry row (key+name+desc+
// logo_url+site_url+redirect_url+scope+oauth_client_id+oauth_client_secret).
// gox originally used Provider as a per-user linked-account row
// (user_id+type+reference+email+token). Both shapes coexist after Phase 4
// (shadow-column strategy); the OAuth provider port in Phase 7.6 writes to
// the TS-side columns. TS calls the credential pair `client_id` /
// `client_secret`; gox stores them under `oauth_client_id` /
// `oauth_client_secret` to avoid collision with the gox per-user-link's
// `client_id` column.
type Provider struct {
	sql.BaseEntity

	// TS-side OAuth provider registry columns (Phase 4).
	Key               *string `gorm:"column:key;type:varchar(255)"                 json:"key,omitempty"`
	Name              *string `gorm:"column:name;type:varchar(255)"                json:"name,omitempty"`
	Desc              *string `gorm:"column:desc;type:varchar(255)"                json:"desc,omitempty"`
	Detail            *string `gorm:"column:detail;type:text"                      json:"detail,omitempty"`
	LogoUrl           *string `gorm:"column:logo_url;type:varchar(255)"            json:"logoUrl,omitempty"`
	SiteUrl           *string `gorm:"column:site_url;type:varchar(255)"            json:"siteUrl,omitempty"`
	RedirectUrl       *string `gorm:"column:redirect_url;type:varchar(255)"        json:"redirectUrl,omitempty"`
	Scope             *string `gorm:"column:scope;type:varchar(255)"               json:"scope,omitempty"`
	OauthClientId     *string `gorm:"column:oauth_client_id;type:varchar(255)"     json:"-"`
	OauthClientSecret *string `gorm:"column:oauth_client_secret;type:varchar(255)" json:"-"`

	// Legacy gox per-user-link columns — kept until callers migrate.
	UserId    string          `gorm:"column:user_id;type:varchar(36);not null"    json:"userId"`
	ClientId  *string         `gorm:"column:client_id;type:varchar(255)"          json:"clientId,omitempty"`
	Type      string          `gorm:"column:type;type:varchar(255);not null"      json:"type"`
	Reference string          `gorm:"column:reference;type:varchar(255);not null" json:"reference"`
	Email     *string         `gorm:"column:email;type:varchar(255)"              json:"email,omitempty"`
	Token     *string         `gorm:"column:token;type:text"                      json:"token,omitempty"`
	Meta      json.RawMessage `gorm:"column:meta;type:jsonb"                      json:"meta,omitempty"`
	Status    *string         `gorm:"column:status;type:varchar(255)"             json:"status,omitempty"`
}

func (Provider) TableName() string { return "IdentityProviders" }

type ProviderEntity struct {
	*sql.Entity[Provider] `inject:""`
}

func (e *ProviderEntity) OnRegister() {
	e.Hydrate("IdentityProviders", []string{"user_id", "type", "reference"}, nil, nil, nil, nil, nil, "created_at desc")
}
