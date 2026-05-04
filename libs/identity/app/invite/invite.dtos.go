package invite

import "time"

type CreateInviteDto struct {
	UserId      string     `json:"userId"      binding:"required"`
	ClientId    string     `json:"clientId"    binding:"required"`
	WorkspaceId string     `json:"workspaceId" binding:"required"`
	Email       string     `json:"email"       binding:"required"`
	RoleId      *string    `json:"roleId,omitempty"`
	Token       string     `json:"token"       binding:"required"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Status      *string    `json:"status,omitempty"`
}

type UpdateInviteDto struct {
	RoleId    *string    `json:"roleId,omitempty"`
	Token     *string    `json:"token,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Status    *string    `json:"status,omitempty"`
}
