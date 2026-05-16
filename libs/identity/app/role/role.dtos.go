package role

import ntxctx "github.com/thescaffold/gox-packages/libs/core/context"

type CreateRoleDto struct {
	Name        string  `json:"name"     binding:"required"`
	ClientId    string  `json:"clientId" binding:"required"`
	WorkspaceId *string `json:"workspaceId,omitempty"`
	Type        string  `json:"type"     binding:"required"`
	EntityName  *string `json:"entityName,omitempty"`
	EntityId    *string `json:"entityId,omitempty"`
	RoleTypeId  *string `json:"roleTypeId,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UpdateRoleDto struct {
	Name        *string `json:"name,omitempty"`
	WorkspaceId *string `json:"workspaceId,omitempty"`
	Type        *string `json:"type,omitempty"`
	EntityName  *string `json:"entityName,omitempty"`
	EntityId    *string `json:"entityId,omitempty"`
	RoleTypeId  *string `json:"roleTypeId,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// LinkRoleDto carries the body for POST /role/attach|detach|sync. Mirrors TS
// LinkRoleDto — pairs an entity (entityName, entityId) with a list of role-type
// ids. ClientId + WorkspaceId can be supplied or hydrated from context.
type LinkRoleDto struct {
	Ctx         ntxctx.NTXContext `context:"ntx"`
	EntityName  string            `json:"entityName"  binding:"required"`
	EntityId    string            `json:"entityId"    binding:"required"`
	ClientId    string            `json:"clientId"    binding:"required"`
	WorkspaceId *string           `json:"workspaceId,omitempty"`
	RoleTypeIds []string          `json:"roleTypeIds" binding:"required"`
}
