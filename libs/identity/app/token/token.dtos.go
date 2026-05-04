package token

import "time"

type CreateTokenDto struct {
	UserId      string     `json:"userId"      binding:"required"`
	ClientId    string     `json:"clientId"    binding:"required"`
	WorkspaceId *string    `json:"workspaceId,omitempty"`
	Type        string     `json:"type"        binding:"required"`
	Token       string     `json:"token"       binding:"required"`
	DeviceId    *string    `json:"deviceId,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Status      *string    `json:"status,omitempty"`
}

type UpdateTokenDto struct {
	Type      *string    `json:"type,omitempty"`
	Token     *string    `json:"token,omitempty"`
	DeviceId  *string    `json:"deviceId,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Status    *string    `json:"status,omitempty"`
}
