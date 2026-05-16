package token

import (
	"github.com/awesome-goose/goose/modules/sql"
	"time"
)

// Token mirrors ntx-apps/libs/identity/src/api/token/entities/token.entity.ts.
// Phase Q adds the TS-aligned columns (name/desc/value/check/expired_at)
// alongside the legacy (type/token/expires_at) shape; the API-key path on
// /token + /token/user writes to the new columns.
type Token struct {
	sql.BaseEntity
	UserId      string     `gorm:"column:user_id;type:varchar(36);not null"      json:"userId"`
	ClientId    string     `gorm:"column:client_id;type:varchar(255);not null"   json:"clientId"`
	WorkspaceId *string    `gorm:"column:workspace_id;type:varchar(36)"          json:"workspaceId,omitempty"`
	Type        string     `gorm:"column:type;type:varchar(255);not null"        json:"type"`
	Token       string     `gorm:"column:token;type:text;not null"               json:"token"`
	Name        *string    `gorm:"column:name;type:varchar(255)"                 json:"name,omitempty"`
	Desc        *string    `gorm:"column:desc;type:text"                         json:"desc,omitempty"`
	Value       *string    `gorm:"column:value;type:varchar(255)"                json:"value,omitempty"`
	Check       *string    `gorm:"column:check;type:varchar(255)"                json:"-"`
	MaskedValue *string    `gorm:"-"                                              json:"maskedValue,omitempty"`
	DeviceId    *string    `gorm:"column:device_id;type:varchar(36)"             json:"deviceId,omitempty"`
	ExpiresAt   *time.Time `gorm:"column:expires_at;type:timestamp"              json:"expiresAt,omitempty"`
	ExpiredAt   *time.Time `gorm:"column:expired_at;type:timestamp"              json:"expiredAt,omitempty"`
	Status      *string    `gorm:"column:status;type:varchar(255)"               json:"status,omitempty"`
}

func (Token) TableName() string { return "IdentityTokens" }

type TokenEntity struct {
	*sql.Entity[Token] `inject:""`
}

func (e *TokenEntity) OnRegister() {
	e.Hydrate("IdentityTokens", []string{"user_id", "client_id", "type"}, nil, nil, nil, nil, nil, "created_at desc")
}
