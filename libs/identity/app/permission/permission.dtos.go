package permission

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

type CreatePermissionDto struct {
	UserId           *string `json:"userId,omitempty"`
	ClientId         *string `json:"clientId,omitempty"`
	WorkspaceId      *string `json:"workspaceId,omitempty"`
	RoleId           *string `json:"roleId,omitempty"`
	RoleTypeId       *string `json:"roleTypeId,omitempty"`
	PermissionTypeId *string `json:"permissionTypeId,omitempty"`
	Key              string  `json:"key"      binding:"required"`
	Scope            string  `json:"scope"    binding:"required"`
	Resource         string  `json:"resource" binding:"required"`
	Action           string  `json:"action"   binding:"required"`
	Status           *string `json:"status,omitempty"`
}

// LinkPermissionDto carries the body for POST /permission/attach|detach|sync.
// Mirrors TS LinkPermissionDto — pairs a roleTypeId with a list of
// permissionTypeIds.
type LinkPermissionDto struct {
	Ctx               ntxctx.NTXContext `context:"ntx"`
	RoleTypeId        string            `json:"roleTypeId"        binding:"required"`
	PermissionTypeIds []string          `json:"permissionTypeIds" binding:"required"`
}

type UpdatePermissionDto struct {
	Key      *string `json:"key,omitempty"`
	Scope    *string `json:"scope,omitempty"`
	Resource *string `json:"resource,omitempty"`
	Action   *string `json:"action,omitempty"`
	Status   *string `json:"status,omitempty"`
}
