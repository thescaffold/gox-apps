package token

import (
	"time"

	ntxctx "github.com/thescaffold/gox-packages/libs/core/context"
)

type CreateTokenDto struct {
	UserId         string     `json:"userId"      binding:"required"`
	ClientId       string     `json:"clientId"    binding:"required"`
	WorkspaceId    *string    `json:"workspaceId,omitempty"`
	Type           string     `json:"type"        binding:"required"`
	Token          string     `json:"token"       binding:"required"`
	Name           *string    `json:"name,omitempty"`
	Desc           *string    `json:"desc,omitempty"`
	Value          *string    `json:"value,omitempty"`
	Check          *string    `json:"check,omitempty"`
	DurationInSec  *int       `json:"durationInSec,omitempty"`
	RoleTypeIds    []string   `json:"roleTypeIds,omitempty"`
	Status         *string    `json:"status,omitempty"`
	DeviceId       *string    `json:"deviceId,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	ExpiredAt      *time.Time `json:"expiredAt,omitempty"`
}

type UpdateTokenDto struct {
	Type      *string    `json:"type,omitempty"`
	Token     *string    `json:"token,omitempty"`
	Name      *string    `json:"name,omitempty"`
	Desc      *string    `json:"desc,omitempty"`
	DeviceId  *string    `json:"deviceId,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	ExpiredAt *time.Time `json:"expiredAt,omitempty"`
	Status    *string    `json:"status,omitempty"`
}

// UserTokenDto carries the body for POST /token/user — name + optional desc
// + optional duration. The handler mints a fresh value+check pair on each
// call when ?refresh is set, or reuses an existing row when one matches the
// (userId, clientId, workspaceId, name) tuple.
type UserTokenDto struct {
	Ctx           ntxctx.NTXContext `context:"ntx"`
	Name          string            `json:"name"           binding:"required"`
	Desc          *string           `json:"desc,omitempty"`
	DurationInSec *int              `json:"durationInSec,omitempty"`
	Refresh       string            `query:"refresh"`
}
